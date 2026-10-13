// Package objective defines exact lexicographic packing objectives. New builds
// built-in criteria; SafeCompare and SafeComponents adapt trusted custom providers.
// Scoring trusts supplied plan statistics and does not verify feasibility. No
// floating-point conversion, currency conversion or background work is owned.
package objective

import (
	"context"
	"errors"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/faustbrian/go-knapsack/v2"
)

var (
	// ErrInvalidObjective identifies an empty, contradictory, or malformed
	// built-in or custom objective.
	ErrInvalidObjective = errors.New("objective: invalid definition or score")
	// ErrCallbackPanic identifies a callback panic without disclosing its value.
	ErrCallbackPanic = errors.New("objective: callback panic")
)

const (
	maxComponents     = 32
	maxComponentBytes = 1024
	maxCriteria       = 7
)

// PlanObjective compares whole candidate plans, which may contain only the
// current partial packing or unused containers. Implementations receive copied
// collection accessors and must be deterministic, bounded and context-cooperative.
// Providers shared between concurrent solver calls must own their synchronization.
type PlanObjective interface {
	// Valid reports usable provider configuration and must not block or panic;
	// validation calls occur outside SafeCompare/SafeComponents panic recovery.
	Valid() bool
	// ComparePlans returns -1 when left is preferred, 0 for a tie and 1 when
	// right is preferred; it must accept partial candidate plans.
	ComparePlans(context.Context, knapsack.NormalizedRequest, knapsack.Plan, knapsack.Plan) (int, error)
	// Components returns 1 through 32 stable score records, with nonempty name,
	// unit and value and min/max direction. SafeComponents checks byte bounds
	// only after return; the provider owns allocation and numerical semantics.
	Components(context.Context, knapsack.NormalizedRequest, knapsack.Plan) ([]knapsack.ScoreComponent, error)
}

// Metric identifies an exact plan property used in lexicographic comparison.
type Metric string

const (
	// ContainerCount minimizes or maximizes the number of container instances.
	ContainerCount Metric = "container_count"
	// TotalPackagingCost is reserved for exact monetary objective adapters.
	TotalPackagingCost Metric = "total_packaging_cost"
	// UnusedVolume measures unused lattice volume across selected containers.
	UnusedVolume Metric = "unused_volume"
	// UnusedWeight measures remaining content-weight capacity.
	UnusedWeight Metric = "unused_weight"
	// WeightImbalance measures the heaviest-minus-lightest content weight.
	WeightImbalance Metric = "weight_imbalance"
	// MaximumUsedHeight measures the highest exclusive placement coordinate.
	MaximumUsedHeight Metric = "maximum_used_height"
	// PackedPriority sums exact caller-provided item priorities.
	PackedPriority Metric = "packed_priority"
)

// Direction controls whether lower or higher metric values are preferred.
type Direction int8

const (
	// Min prefers the lower exact value.
	Min Direction = -1
	// Max prefers the higher exact value.
	Max Direction = 1
)

// Criterion pairs a metric with its lexicographic comparison direction.
type Criterion struct {
	// Metric is the exact plan property to evaluate.
	Metric Metric
	// Direction selects minimization or maximization.
	Direction Direction
}

// Minimize constructs a minimizing criterion.
func Minimize(metric Metric) Criterion { return Criterion{Metric: metric, Direction: Min} }

// Maximize constructs a maximizing criterion.
func Maximize(metric Metric) Criterion { return Criterion{Metric: metric, Direction: Max} }

// Score contains exact criterion values followed by a canonical tie-break.
type Score struct {
	// Values correspond positionally to Objective.Criteria.
	Values []int64
	// TieBreak is compared lexically only after every criterion is equal.
	TieBreak string
}

// Objective is an immutable ordered lexicographic objective.
type Objective struct{ criteria []Criterion }

// New validates and copies 1 through 7 unique criteria. TotalPackagingCost is
// reserved for adapters and cannot be scored by this built-in Objective.
func New(criteria ...Criterion) (Objective, error) {
	if len(criteria) == 0 || len(criteria) > maxCriteria {
		return Objective{}, ErrInvalidObjective
	}
	seen := make(map[Metric]struct{}, len(criteria))
	for _, criterion := range criteria {
		if !validMetric(criterion.Metric) || criterion.Direction != Min && criterion.Direction != Max {
			return Objective{}, ErrInvalidObjective
		}
		if _, duplicate := seen[criterion.Metric]; duplicate {
			return Objective{}, ErrInvalidObjective
		}
		seen[criterion.Metric] = struct{}{}
	}
	return Objective{criteria: slices.Clone(criteria)}, nil
}

// Criteria returns a defensive copy in precedence order.
func (o Objective) Criteria() []Criterion { return slices.Clone(o.criteria) }

// Valid reports whether the objective has at least one validated criterion.
func (o Objective) Valid() bool { return len(o.criteria) > 0 }

// ComparePlans scores two plans and returns the exact lexicographic ordering.
func (o Objective) ComparePlans(ctx context.Context, request knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	if ctx == nil {
		return 0, knapsack.ErrInvalidOptions
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	leftScore, _, err := o.ScorePlan(request, left)
	if err != nil {
		return 0, err
	}
	rightScore, _, err := o.ScorePlan(request, right)
	if err != nil {
		return 0, err
	}
	return o.Compare(leftScore, rightScore)
}

// Components returns bounded serializable score evidence for a plan.
func (o Objective) Components(ctx context.Context, request knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	if ctx == nil {
		return nil, knapsack.ErrInvalidOptions
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	_, components, err := o.scorePlan(request, plan, false)
	return components, err
}

// SafeCompare invokes a plan objective with validation, panic conversion, and
// a canonical final tie-break.
func SafeCompare(ctx context.Context, objective PlanObjective, request knapsack.NormalizedRequest, left, right knapsack.Plan) (comparison int, err error) {
	if err := validateCallback(ctx, objective); err != nil {
		return 0, err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			comparison = 0
			err = ErrCallbackPanic
		}
	}()
	comparison, err = objective.ComparePlans(ctx, request, left, right)
	if comparison < -1 || comparison > 1 {
		return 0, ErrInvalidObjective
	}
	if comparison == 0 && err == nil {
		comparison = strings.Compare(left.CanonicalString(), right.CanonicalString())
	}
	return comparison, err
}

// SafeComponents invokes a plan objective and bounds its serialized result.
func SafeComponents(ctx context.Context, objective PlanObjective, request knapsack.NormalizedRequest, plan knapsack.Plan) (components []knapsack.ScoreComponent, err error) {
	if err := validateCallback(ctx, objective); err != nil {
		return nil, err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			components = nil
			err = ErrCallbackPanic
		}
	}()
	components, err = objective.Components(ctx, request, plan)
	if err != nil {
		return nil, err
	}
	if len(components) == 0 || len(components) > maxComponents {
		return nil, ErrInvalidObjective
	}
	for _, component := range components {
		if component.Name == "" || component.Direction != "min" && component.Direction != "max" ||
			component.Unit == "" || component.Value == "" || len(component.Name)+len(component.Unit)+len(component.Value) > maxComponentBytes {
			return nil, ErrInvalidObjective
		}
	}
	return slices.Clone(components), nil
}

func validateCallback(ctx context.Context, objective PlanObjective) error {
	if ctx == nil {
		return knapsack.ErrInvalidOptions
	}
	if objective == nil || typedNil(objective) || !objective.Valid() {
		return ErrInvalidObjective
	}
	return ctx.Err()
}

func typedNil(value any) bool {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

// ScorePlan computes checked integer metrics without binary floating point.
// Several values come directly from plan.Statistics; use a verified plan when
// the score must represent actual feasibility/accounting. It has no context
// argument.
func (o Objective) ScorePlan(request knapsack.NormalizedRequest, plan knapsack.Plan) (Score, []knapsack.ScoreComponent, error) {
	return o.scorePlan(request, plan, true)
}

// Component evidence does not consume the canonical tie. Keep its checked
// metric path identical without serializing metadata that will be discarded.
func (o Objective) scorePlan(request knapsack.NormalizedRequest, plan knapsack.Plan, canonicalTie bool) (Score, []knapsack.ScoreComponent, error) {
	if !o.Valid() {
		return Score{}, nil, ErrInvalidObjective
	}
	stats := plan.Statistics()
	placements := plan.Placements()
	items := request.Items()
	priorities := make(map[string]int64, len(items))
	for _, item := range items {
		priorities[item.ID] = item.Priority
	}
	weights := make(map[string]int64)
	var maximumHeight, packedPriority int64
	for _, placement := range placements {
		if !checkedAddMap(weights, placement.ContainerID, placement.Weight) {
			return Score{}, nil, ErrInvalidObjective
		}
		if placement.Origin.Z > math.MaxInt64-placement.Dimensions.Z {
			return Score{}, nil, ErrInvalidObjective
		}
		height := placement.Origin.Z + placement.Dimensions.Z
		if height > maximumHeight {
			maximumHeight = height
		}
		priority, known := priorities[placement.ItemID]
		if !known || priority > 0 && packedPriority > math.MaxInt64-priority || priority < 0 && packedPriority < math.MinInt64-priority {
			return Score{}, nil, ErrInvalidObjective
		}
		packedPriority += priority
	}
	var minimumWeight, maximumWeight int64
	first := true
	for _, container := range plan.Containers() {
		weight := weights[container.ID]
		if first || weight < minimumWeight {
			minimumWeight = weight
		}
		if first || weight > maximumWeight {
			maximumWeight = weight
		}
		first = false
	}
	if minimumWeight < 0 && maximumWeight > math.MaxInt64+minimumWeight {
		return Score{}, nil, ErrInvalidObjective
	}
	imbalance := maximumWeight - minimumWeight
	score := Score{Values: make([]int64, len(o.criteria))}
	if canonicalTie {
		score.TieBreak = plan.CanonicalString()
	}
	components := make([]knapsack.ScoreComponent, len(o.criteria))
	for index, criterion := range o.criteria {
		var value int64
		unit := "count"
		switch criterion.Metric {
		case ContainerCount:
			value = int64(stats.ContainerCount)
		case UnusedVolume:
			value, unit = stats.RemainingVolume, "lattice^3"
		case UnusedWeight:
			value, unit = stats.RemainingWeight, "mass_lattice"
		case WeightImbalance:
			value, unit = imbalance, "mass_lattice"
		case MaximumUsedHeight:
			value, unit = maximumHeight, "length_lattice"
		case PackedPriority:
			value, unit = packedPriority, "priority"
		case TotalPackagingCost:
			return Score{}, nil, ErrInvalidObjective
		}
		direction := "min"
		if criterion.Direction == Max {
			direction = "max"
		}
		score.Values[index] = value
		components[index] = knapsack.ScoreComponent{Name: string(criterion.Metric), Direction: direction, Unit: unit, Value: strconv.FormatInt(value, 10)}
	}
	return score, components, nil
}

func checkedAddMap(values map[string]int64, key string, value int64) bool {
	current := values[key]
	if value > 0 && current > math.MaxInt64-value || value < 0 && current < math.MinInt64-value {
		return false
	}
	values[key] = current + value
	return true
}

// Compare returns negative when left is preferred, positive when right is
// preferred, and zero only for canonically identical scores.
func (o Objective) Compare(left, right Score) (int, error) {
	if len(o.criteria) == 0 || len(left.Values) != len(o.criteria) || len(right.Values) != len(o.criteria) {
		return 0, ErrInvalidObjective
	}
	for index, criterion := range o.criteria {
		if left.Values[index] == right.Values[index] {
			continue
		}
		comparison := -1
		if left.Values[index] > right.Values[index] {
			comparison = 1
		}
		if criterion.Direction == Max {
			comparison = -comparison
		}
		return comparison, nil
	}
	return strings.Compare(left.TieBreak, right.TieBreak), nil
}

func validMetric(metric Metric) bool {
	switch metric {
	case ContainerCount, TotalPackagingCost, UnusedVolume, UnusedWeight, WeightImbalance, MaximumUsedHeight, PackedPriority:
		return true
	default:
		return false
	}
}
