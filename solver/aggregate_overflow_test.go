package solver_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func TestSolversRefuseUnrepresentablePlanAggregates(t *testing.T) {
	for _, field := range []string{"volume", "remaining_weight"} {
		t.Run(field, func(t *testing.T) {
			for _, strategy := range []string{"exact", "heuristic"} {
				for _, scenario := range []string{"infeasible", "candidate_limit", "placement_trial"} {
					t.Run(strategy+"/"+scenario, func(t *testing.T) {
						base := exactRequest(t, 2, 1)
						container := knapsack.NormalizedContainer{
							ID: "box", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1},
							MaxContentWeight: 1, Stock: knapsack.FiniteStock(2), AllowedClasses: []string{"other"},
						}
						if field == "volume" {
							container.Dimensions.X = math.MaxInt64
						} else {
							container.MaxContentWeight = math.MaxInt64
						}
						limits := base.Limits()
						if scenario == "candidate_limit" {
							limits.MaxCandidatePlacements = 1
						}
						if scenario == "placement_trial" {
							container.AllowedClasses = []string{"item"}
						}
						request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
							Items: []knapsack.NormalizedItem{{ID: "item", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1},
								Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}, Attributes: map[string]string{"class": "item"}}},
							Containers: []knapsack.NormalizedContainer{container}, Resolution: base.Resolution(), Limits: limits,
						})
						if err != nil {
							t.Fatal(err)
						}
						// No center lattice is requested: only two tiny origin candidates
						// are visited, regardless of the dimensions' scalar magnitudes.
						instances := []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}, {ID: "box#2", TypeID: "box"}}
						var plan knapsack.Plan
						if strategy == "exact" {
							plan, err = (solver.Exact{}).PackFixed(t.Context(), request, instances, solver.Options{})
						} else {
							plan, err = (solver.Heuristic{}).PackFixed(t.Context(), request, instances, solver.Options{})
						}
						if !errors.Is(err, knapsack.ErrOverflow) || plan.Status() != "" ||
							len(plan.Containers()) != 0 || len(plan.Placements()) != 0 {
							t.Fatalf("unrepresentable %s published status=%s statistics=%+v error=%v", field, plan.Status(), plan.Statistics(), err)
						}
					})
				}
			}
		})
	}
}

func TestExactPackAllPreservesRepresentableAlternatives(t *testing.T) {
	for _, priority := range []int64{-1, 1} {
		t.Run(map[int64]string{-1: "overflow_first", 1: "overflow_last"}[priority], func(t *testing.T) {
			base := exactRequest(t, 2, 1)
			request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
				Items: []knapsack.NormalizedItem{
					{ID: "first", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
					{ID: "second", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
				},
				Containers: []knapsack.NormalizedContainer{
					{ID: "large", Dimensions: geometry.Dimensions{X: math.MaxInt64, Y: 1, Z: 1}, MaxContentWeight: 1, Stock: knapsack.FiniteStock(2), Priority: priority},
					{ID: "small", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, MaxContentWeight: 1, Stock: knapsack.FiniteStock(2)},
				},
				Resolution: base.Resolution(), Limits: base.Limits(),
			})
			if err != nil {
				t.Fatal(err)
			}
			plan, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{})
			if err != nil || len(plan.Placements()) != 2 || plan.Statistics().ContainerVolume != 2 {
				t.Fatalf("representable alternative discarded: plan=%+v error=%v", plan.Spec(), err)
			}
			if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
				t.Fatalf("alternative failed independent verification: %+v", result)
			}
			if priority == -1 {
				for _, test := range []struct {
					name        string
					limit       func(*knapsack.Limits)
					termination knapsack.TerminationReason
				}{
					{"spent_nodes", func(l *knapsack.Limits) { l.MaxSearchNodes = 3 }, knapsack.TerminationNodeLimit},
					{"partial_nodes", func(l *knapsack.Limits) { l.MaxSearchNodes = 1 }, knapsack.TerminationNodeLimit},
					{"partial_candidates", func(l *knapsack.Limits) { l.MaxCandidatePlacements = 1 }, knapsack.TerminationCandidateLimit},
					{"partial_branches", func(l *knapsack.Limits) { l.MaxBranches = 3 }, knapsack.TerminationBranchLimit},
				} {
					t.Run(test.name, func(t *testing.T) {
						limits := request.Limits()
						test.limit(&limits)
						limited, err := (solver.Exact{}).PackAll(t.Context(), request.WithLimits(limits), solver.Options{})
						if !errors.Is(err, knapsack.ErrBudgetExhausted) || limited.Status() != knapsack.StatusBudgetExhausted || limited.Termination() != test.termination {
							t.Fatalf("refused configuration lost work or termination: plan=%+v error=%v", limited.Spec(), err)
						}
					})
				}
			} else {
				calls := 0
				failure := fmt.Errorf("objective collaborator: %w", knapsack.ErrOverflow)
				failed, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{PlanObjective: aggregateErrorObjective{calls: &calls, failure: failure}})
				if err != failure || failed.Status() != "" || calls != 1 {
					t.Fatalf("collaborator refusal swallowed: status=%s error=%v calls=%d", failed.Status(), err, calls)
				}
			}
		})
	}
}

type aggregateErrorObjective struct {
	calls   *int
	failure error
}

func TestExactPackAllRefusesWhenEveryConfigurationOverflows(t *testing.T) {
	base := exactRequest(t, 2, 1)
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: []knapsack.NormalizedItem{
			{ID: "first", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
			{ID: "second", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
		},
		Containers: []knapsack.NormalizedContainer{{ID: "large", Dimensions: geometry.Dimensions{X: math.MaxInt64, Y: 1, Z: 1}, MaxContentWeight: 1, Stock: knapsack.FiniteStock(2)}},
		Resolution: base.Resolution(), Limits: base.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{})
	if !errors.Is(err, knapsack.ErrOverflow) || plan.Status() != "" || len(plan.Containers()) != 0 {
		t.Fatalf("all-refused search falsely published plan=%+v error=%v", plan.Spec(), err)
	}
}

func TestHeuristicTerminalPublicationPropagatesPlanAdmissionFailure(t *testing.T) {
	base := exactRequest(t, 1, 1)
	limits := base.Limits()
	limits.MaxIDBytes = knapsack.DefaultPlanLimits().MaxIDBytes + 1
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items:      []knapsack.NormalizedItem{{ID: strings.Repeat("x", int(limits.MaxIDBytes)), Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}}},
		Containers: []knapsack.NormalizedContainer{{ID: "box", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, MaxContentWeight: 1, Stock: knapsack.FiniteStock(1), AllowedClasses: []string{"other"}}},
		Resolution: base.Resolution(), Limits: limits,
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := (solver.Heuristic{}).PackAll(t.Context(), request, solver.Options{})
	if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != "" || len(plan.UnpackedItemIDs()) != 0 {
		t.Fatalf("oversized unpacked ID published plan=%+v error=%v", plan.Spec(), err)
	}
}

func (aggregateErrorObjective) Valid() bool { return true }
func (o aggregateErrorObjective) ComparePlans(context.Context, knapsack.NormalizedRequest, knapsack.Plan, knapsack.Plan) (int, error) {
	return 0, o.failure
}
func (o aggregateErrorObjective) Components(context.Context, knapsack.NormalizedRequest, knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	*o.calls++
	return nil, o.failure
}

func TestSolversRetainInclusiveAggregateLimits(t *testing.T) {
	base := exactRequest(t, 2, 1)
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: []knapsack.NormalizedItem{{ID: "item", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1},
			Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}, Attributes: map[string]string{"class": "item"}}},
		Containers: []knapsack.NormalizedContainer{
			{ID: "large", Dimensions: geometry.Dimensions{X: math.MaxInt64 - 1, Y: 1, Z: 1}, MaxContentWeight: math.MaxInt64 - 1, Stock: knapsack.FiniteStock(1), AllowedClasses: []string{"other"}},
			{ID: "small", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, MaxContentWeight: 1, Stock: knapsack.FiniteStock(1), AllowedClasses: []string{"other"}},
		},
		Resolution: base.Resolution(), Limits: base.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	instances := []knapsack.ContainerInstance{{ID: "large#1", TypeID: "large"}, {ID: "small#1", TypeID: "small"}}
	for _, strategy := range []string{"exact", "heuristic"} {
		t.Run(strategy, func(t *testing.T) {
			var plan knapsack.Plan
			var err error
			if strategy == "exact" {
				plan, err = (solver.Exact{}).PackFixed(t.Context(), request, instances, solver.Options{})
				if !errors.Is(err, knapsack.ErrProvenInfeasible) {
					t.Fatalf("expected ordinary infeasibility, got %v", err)
				}
			} else {
				plan, err = (solver.Heuristic{}).PackFixed(t.Context(), request, instances, solver.Options{})
				if err != nil {
					t.Fatal(err)
				}
			}
			stats := plan.Statistics()
			if stats.ContainerVolume != math.MaxInt64 || stats.RemainingVolume != math.MaxInt64 || stats.RemainingWeight != math.MaxInt64 ||
				len(plan.Containers()) != 2 || len(plan.Placements()) != 0 || len(plan.UnpackedItemIDs()) != 1 {
				t.Fatalf("inclusive totals were not preserved: %+v", plan.Spec())
			}
		})
	}
}
