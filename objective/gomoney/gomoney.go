// Package gomoney preserves the legacy exact-money Knapsack objective API.
// New consumers should use github.com/faustbrian/go-knapsack/objective/money.
package gomoney

import (
	"context"

	"github.com/faustbrian/go-knapsack"
	moneyobjective "github.com/faustbrian/go-knapsack/objective/money"
	"github.com/faustbrian/go-money"
)

var (
	// ErrInvalidCosts identifies an empty, invalid, or incompatible cost mapping.
	//
	// Deprecated: use moneyobjective.ErrInvalidCosts.
	ErrInvalidCosts = moneyobjective.ErrInvalidCosts
	// ErrMissingCost identifies a selected container type without a configured
	// packaging cost.
	//
	// Deprecated: use moneyobjective.ErrMissingCost.
	ErrMissingCost = moneyobjective.ErrMissingCost
	// ErrDuplicateTypeID identifies repeated container type IDs supplied through
	// NewFromEntries.
	//
	// Deprecated: use moneyobjective.ErrDuplicateTypeID.
	ErrDuplicateTypeID = moneyobjective.ErrDuplicateTypeID
	// ErrUnsupportedScale identifies a money context whose scale is not fixed.
	//
	// Deprecated: use moneyobjective.ErrUnsupportedScale.
	ErrUnsupportedScale = moneyobjective.ErrUnsupportedScale
	// ErrNegativeCost identifies a negative cost rejected by the active policy.
	//
	// Deprecated: use moneyobjective.ErrNegativeCost.
	ErrNegativeCost = moneyobjective.ErrNegativeCost
)

// Limits bounds cost collections copied by the constructors.
//
// Deprecated: use moneyobjective.Limits.
type Limits struct {
	// MaxTypes bounds configured container type costs.
	MaxTypes uint32
	// MaxIDBytes bounds each container type identifier.
	MaxIDBytes uint32
}

// DefaultLimits returns conservative limits for untrusted cost maps.
//
// Deprecated: use moneyobjective.DefaultLimits.
func DefaultLimits() Limits {
	limits := moneyobjective.DefaultLimits()
	return Limits{MaxTypes: limits.MaxTypes, MaxIDBytes: limits.MaxIDBytes}
}

// Policy controls bounded construction and whether negative values explicitly
// model credits or rebates. DefaultPolicy rejects negative costs.
//
// Deprecated: use moneyobjective.Policy.
type Policy struct {
	// Limits bounds the copied mapping and its type IDs.
	Limits Limits
	// AllowNegativeCosts permits values that explicitly model credits or rebates.
	AllowNegativeCosts bool
}

// DefaultPolicy returns the default resource limits and rejects negative costs.
//
// Deprecated: use moneyobjective.DefaultPolicy.
func DefaultPolicy() Policy {
	policy := moneyobjective.DefaultPolicy()
	return Policy{
		Limits: Limits{
			MaxTypes:   policy.Limits.MaxTypes,
			MaxIDBytes: policy.Limits.MaxIDBytes,
		},
		AllowNegativeCosts: policy.AllowNegativeCosts,
	}
}

// Entry is one container type ID and exact packaging cost. Entry construction
// is useful when duplicate IDs must be detected before forming a Go map.
//
// Deprecated: use moneyobjective.Entry.
type Entry struct {
	// TypeID is the exact Knapsack container type identifier.
	TypeID string
	// Cost is the exact cost for each selected instance of TypeID.
	Cost money.Money
}

// Costs is an immutable exact packaging-cost objective keyed by container
// type ID.
//
// Deprecated: use moneyobjective.Costs.
type Costs struct {
	successor moneyobjective.Costs
}

// New validates and copies a cost map using DefaultPolicy.
//
// Deprecated: use moneyobjective.New.
func New(values map[string]money.Money) (Costs, error) {
	result, err := moneyobjective.New(values)
	return Costs{successor: result}, err
}

// NewWithLimits validates, sorts, and defensively copies a bounded nonempty
// single-currency cost map.
//
// Deprecated: use moneyobjective.NewWithLimits.
func NewWithLimits(values map[string]money.Money, limits Limits) (Costs, error) {
	result, err := moneyobjective.NewWithLimits(values, successorLimits(limits))
	return Costs{successor: result}, err
}

// NewWithPolicy validates and copies a cost map using an explicit negative-cost
// and resource policy. A Go map has already collapsed duplicate keys; callers
// that need duplicate detection must use NewFromEntries.
//
// Deprecated: use moneyobjective.NewWithPolicy.
func NewWithPolicy(values map[string]money.Money, policy Policy) (Costs, error) {
	result, err := moneyobjective.NewWithPolicy(values, successorPolicy(policy))
	return Costs{successor: result}, err
}

// NewFromEntries validates, sorts, and defensively copies a bounded nonempty
// entry sequence. It rejects duplicate type IDs and requires one currency and
// one fixed Default or Custom money context across all values.
//
// Deprecated: use moneyobjective.NewFromEntries.
func NewFromEntries(entries []Entry, policy Policy) (Costs, error) {
	if uint64(len(entries)) > uint64(policy.Limits.MaxTypes) {
		return Costs{}, ErrInvalidCosts
	}
	successorEntries := make([]moneyobjective.Entry, len(entries))
	for index, entry := range entries {
		successorEntries[index] = moneyobjective.Entry{
			TypeID: entry.TypeID,
			Cost:   entry.Cost,
		}
	}
	result, err := moneyobjective.NewFromEntries(successorEntries, successorPolicy(policy))
	return Costs{successor: result}, err
}

// Valid reports whether the objective contains aligned type IDs and costs.
//
// Deprecated: use moneyobjective.Costs.Valid.
func (c Costs) Valid() bool {
	return c.successor.Valid()
}

// ComparePlans implements objective.PlanObjective with context cancellation.
//
// Deprecated: use moneyobjective.Costs.ComparePlans.
func (c Costs) ComparePlans(
	ctx context.Context,
	request knapsack.NormalizedRequest,
	left knapsack.Plan,
	right knapsack.Plan,
) (int, error) {
	return c.successor.ComparePlans(ctx, request, left, right)
}

// Components returns the exact total packaging cost and ISO currency unit.
//
// Deprecated: use moneyobjective.Costs.Components.
func (c Costs) Components(
	ctx context.Context,
	request knapsack.NormalizedRequest,
	plan knapsack.Plan,
) ([]knapsack.ScoreComponent, error) {
	return c.successor.Components(ctx, request, plan)
}

// Total sums configured exact costs for every selected container instance.
//
// Deprecated: use moneyobjective.Costs.Total.
func (c Costs) Total(plan knapsack.Plan) (money.Money, error) {
	return c.successor.Total(plan)
}

// Compare prefers lower exact cost, then canonical plan bytes for ties.
//
// Deprecated: use moneyobjective.Costs.Compare.
func (c Costs) Compare(left, right knapsack.Plan) (int, error) {
	return c.successor.Compare(left, right)
}

func successorLimits(limits Limits) moneyobjective.Limits {
	return moneyobjective.Limits{
		MaxTypes:   limits.MaxTypes,
		MaxIDBytes: limits.MaxIDBytes,
	}
}

func successorPolicy(policy Policy) moneyobjective.Policy {
	return moneyobjective.Policy{
		Limits:             successorLimits(policy.Limits),
		AllowNegativeCosts: policy.AllowNegativeCosts,
	}
}
