package solver_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/objective"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func TestExactEmptyNativeObjectiveReturnsAdmissionErrorWithoutPanic(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Errorf("public invalid objective panicked instead of returning its admission error: %v", recovered)
		}
	}()
	request := completenessRequest(t, 1, knapsack.FiniteStock(1))
	plan, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{PlanObjective: objective.Objective{}})
	if !errors.Is(err, objective.ErrInvalidObjective) || plan.Status() != "" {
		t.Fatalf("invalid objective admission became a packing/proof: %s, %v", plan.CanonicalString(), err)
	}
}

func TestExactFiniteStockCombinesIndependentContainerTypes(t *testing.T) {
	dimensions := geometry.Dimensions{X: 1, Y: 1, Z: 1}
	base := oracleRequest(t, []knapsack.NormalizedItem{
		oracleItem(t, "first", dimensions), oracleItem(t, "second", dimensions),
	}, dimensions)
	containers := base.Containers()
	containers[0].ID = "a"
	later := containers[0]
	later.ID = "b"
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: base.Items(), Containers: append(containers, later), Resolution: base.Resolution(), Limits: base.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	goal := preferCount{count: 2}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: goal})
	// Two unit cubes cannot share either unit bin. One bin of each available
	// type is a complete packing and independently establishes the optimum.
	if err != nil || plan.Status() != knapsack.StatusOptimal || len(plan.Containers()) != 2 || len(plan.Placements()) != 2 {
		t.Fatalf("finite stock was not combined: %s, %v", plan.CanonicalString(), err)
	}
	if checked := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !checked.Valid() {
		t.Fatalf("combined stock plan invalid: %v", checked.Violations())
	}
}

func TestExactNativeMinimumCountProvesUnlimitedStockOptimum(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.UnlimitedStock())
	limits := request.Limits()
	limits.MaxBranches, limits.MaxCandidatePlacements, limits.MaxSearchNodes = 16, 16, 16
	request = request.WithLimits(limits)
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{})
	// One nonempty unit bin is feasible; a complete packing cannot use zero
	// bins. Default minimum-count ordering therefore has a finite proof even
	// though the stock itself is unlimited.
	if err != nil || plan.Status() != knapsack.StatusOptimal || len(plan.Containers()) != 1 || len(plan.Placements()) != 1 {
		t.Fatalf("native lower-bound optimum lost: %s, %v", plan.CanonicalString(), err)
	}
	if checked := verify.Plan(request, plan, verify.RequireAll()); !checked.Valid() {
		t.Fatalf("native optimum invalid: %v", checked.Violations())
	}
}

func TestExactVariableSearchSharesPlacementAndBranchBudgets(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(2))
	goal := preferCount{count: 2}
	for _, test := range []struct {
		name        string
		branches    uint64
		candidates  uint64
		termination knapsack.TerminationReason
	}{
		{name: "branches", branches: 4, candidates: 100, termination: knapsack.TerminationBranchLimit},
		{name: "candidates", branches: 100, candidates: 2, termination: knapsack.TerminationCandidateLimit},
		{name: "complete_control", branches: 100, candidates: 100},
	} {
		t.Run(test.name, func(t *testing.T) {
			limits := base.Limits()
			limits.MaxBranches, limits.MaxCandidatePlacements = test.branches, test.candidates
			request := base.WithLimits(limits)
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: goal})
			if test.termination == "" {
				if err != nil || plan.Status() != knapsack.StatusOptimal || len(plan.Containers()) != 2 {
					t.Fatalf("complete control failed: %s, %v", plan.CanonicalString(), err)
				}
			} else if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != test.termination {
				t.Fatalf("shared resource refusal became proof: %s, %v", plan.CanonicalString(), err)
			}
			if len(plan.Placements()) != 1 || len(plan.UnpackedItemIDs()) != 0 {
				t.Fatalf("complete feasible candidate lost: %s", plan.CanonicalString())
			}
			if checked := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !checked.Valid() {
				t.Fatalf("retained candidate invalid: %v", checked.Violations())
			}
		})
	}
}

func TestExactCountAdmissionPrecedesLaterConfigurationBudget(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(3))
	limits := base.Limits()
	limits.MaxMemoryBytes = 2*base.MemoryBytes() + 2816 + 48
	limits.MaxBranches = 3
	request := base.WithLimits(limits)
	goal := preferCount{count: 3}
	plan, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{PlanObjective: goal})
	if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationMemoryLimit || len(plan.Containers()) != 1 || len(plan.Placements()) != 1 {
		t.Fatalf("count admission changed refusal origin or lost feasible state: %s, %v", plan.CanonicalString(), err)
	}
	if checked := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !checked.Valid() {
		t.Fatalf("retained count-admission state invalid: %v", checked.Violations())
	}
}

func TestExactAdmitsMaximumPlanCountBeforePlacementBudgetRefusal(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(10_000))
	items := base.Items()
	items[0].Dimensions = geometry.Dimensions{X: 10_000, Y: 1, Z: 1}
	items[0].Orientations = []geometry.Orientation{geometry.OrientationXYZ}
	limits := base.Limits()
	limits.MaxCandidatePlacements = 1
	request, admissionErr := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: items, Containers: base.Containers(), Resolution: base.Resolution(), Limits: limits})
	if admissionErr != nil {
		t.Fatal(admissionErr)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: preferCount{count: 10_000}})
	if !errors.Is(err, knapsack.ErrBudgetExhausted) || errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationCandidateLimit || len(plan.Placements()) != 0 || len(plan.UnpackedItemIDs()) != 1 {
		t.Fatalf("admitted maximum count became premature memory refusal/proof: %s, %v", plan.CanonicalString(), err)
	}
	if checked := verify.Plan(request, plan, verify.AllowUnpacked()); !checked.Valid() {
		t.Fatalf("maximum-count interruption invalid: %v", checked.Violations())
	}
}

type releaseCancelRejectedPlacement struct{ cancel context.CancelFunc }

func (predicate releaseCancelRejectedPlacement) Check(context.Context, constraint.PlacementView) constraint.Decision {
	predicate.cancel()
	return constraint.Reject("cancelled_selection", "the last placement was rejected")
}

func TestExactLastRejectedPlacementPreservesActualCancellation(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.FiniteStock(1))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	plan, err := (solver.Exact{}).PackFixed(ctx, request,
		[]knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}},
		solver.Options{Constraints: []constraint.Placement{releaseCancelRejectedPlacement{cancel: cancel}}})
	if !errors.Is(err, context.Canceled) || errors.Is(err, knapsack.ErrProvenInfeasible) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationCancelled {
		t.Fatalf("rejected final placement became infeasibility proof: %s, %v", plan.CanonicalString(), err)
	}
	if len(plan.Placements()) != 0 || len(plan.UnpackedItemIDs()) != 1 {
		t.Fatalf("cancelled item conservation failed: %s", plan.CanonicalString())
	}
	if checked := verify.Plan(request, plan, verify.AllowUnpacked()); !checked.Valid() {
		t.Fatalf("cancelled plan invalid: %v", checked.Violations())
	}
}

// Trigger final publication by public plan state, rather than depending on
// callback counts or the solver's internal error classification.
type cancelPublicationGoal struct {
	base                     preferCount
	cancel                   context.CancelFunc
	failure                  error
	variable                 bool
	comparison               bool
	aggregate                bool
	interruptedComparison    bool
	interruptedConfiguration bool
}

func (cancelPublicationGoal) Valid() bool { return true }

func (goal cancelPublicationGoal) Components(ctx context.Context, request knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	if !goal.comparison && (plan.Status() == knapsack.StatusOptimal && !goal.aggregate || plan.Status() == knapsack.StatusFeasible && goal.aggregate) && (plan.Work().Strategy == "exhaustive_variable_container") == goal.variable {
		goal.cancel()
		if goal.failure != nil {
			return nil, goal.failure
		}
		return nil, ctx.Err()
	}
	return goal.base.Components(ctx, request, plan)
}

func (goal cancelPublicationGoal) ComparePlans(ctx context.Context, request knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	if goal.comparison && (left.Status() == knapsack.StatusOptimal && !goal.interruptedComparison && !goal.interruptedConfiguration && (left.Work().Strategy == "exhaustive_variable_container") == goal.variable || left.Status() == knapsack.StatusBudgetExhausted && (goal.interruptedComparison && len(left.Containers()) > len(right.Containers()) || goal.interruptedConfiguration && len(left.Containers()) == len(right.Containers()))) {
		goal.cancel()
		if goal.failure != nil {
			return 0, goal.failure
		}
		return 0, ctx.Err()
	}
	return goal.base.ComparePlans(ctx, request, left, right)
}

func TestExactPublicationKeepsBorrowedFailureDistinctFromCancellation(t *testing.T) {
	marker := errors.New("publication collaborator failed independently")
	for _, phase := range []string{"fixed_score", "inner_fixed_score", "configuration_compare", "interrupted_configuration_compare", "interrupted_compare", "variable_score", "variable_compare", "aggregate_verification"} {
		for _, outcome := range []struct {
			name    string
			failure error
		}{{name: "owned_cancel"}, {name: "borrowed_failure", failure: marker}, {name: "borrowed_budget_failure", failure: errors.Join(marker, knapsack.ErrMemoryBudgetExhausted)}} {
			t.Run(phase+"/"+outcome.name, func(t *testing.T) {
				request := completenessRequest(t, 1, knapsack.FiniteStock(2))
				if phase == "interrupted_configuration_compare" {
					request = completenessRequest(t, 2, knapsack.FiniteStock(1))
					limits := request.Limits()
					limits.MaxSearchNodes = 5
					request = request.WithLimits(limits)
				}
				if phase == "configuration_compare" || phase == "interrupted_configuration_compare" {
					containers := request.Containers()
					containers[0].ID, containers[0].Stock = "a", knapsack.FiniteStock(1)
					later := containers[0]
					later.ID = "b"
					var admissionErr error
					request, admissionErr = knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: request.Items(), Containers: append(containers, later), Resolution: request.Resolution(), Limits: request.Limits()})
					if admissionErr != nil {
						t.Fatal(admissionErr)
					}
				}
				if phase == "interrupted_compare" {
					limits := request.Limits()
					limits.MaxBranches = 4
					request = request.WithLimits(limits)
				}
				if phase == "aggregate_verification" {
					containers := request.Containers()
					containers[0].MaxContentWeight = math.MaxInt64
					var admissionErr error
					request, admissionErr = knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: request.Items(), Containers: containers, Resolution: request.Resolution(), Limits: request.Limits()})
					if admissionErr != nil {
						t.Fatal(admissionErr)
					}
				}
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				goal := cancelPublicationGoal{base: preferCount{count: 2}, cancel: cancel, variable: phase != "fixed_score" && phase != "inner_fixed_score" && phase != "configuration_compare", comparison: phase == "variable_compare" || phase == "configuration_compare" || phase == "interrupted_compare" || phase == "interrupted_configuration_compare", aggregate: phase == "aggregate_verification", interruptedComparison: phase == "interrupted_compare", interruptedConfiguration: phase == "interrupted_configuration_compare", failure: outcome.failure}
				var plan knapsack.Plan
				var err error
				if phase == "fixed_score" {
					plan, err = (solver.Exact{}).PackFixed(ctx, request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{PlanObjective: goal})
				} else {
					plan, err = (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: goal})
				}
				if ctx.Err() != context.Canceled {
					t.Fatal("publication cancellation seam was not reached")
				}
				if outcome.failure != nil {
					if !errors.Is(err, marker) || errors.Is(err, context.Canceled) || plan.Status() != "" {
						t.Fatalf("borrowed failure replaced by owned cancellation/publication: %s, %v", plan.CanonicalString(), err)
					}
					return
				}
				if !errors.Is(err, context.Canceled) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationCancelled || len(plan.Placements()) != 1 || len(plan.UnpackedItemIDs()) != 0 {
					t.Fatalf("publication cancellation lost feasible state: %s, %v", plan.CanonicalString(), err)
				}
				if phase == "variable_compare" && len(plan.Containers()) != 1 {
					t.Fatalf("cancelled comparison replaced certified earlier selection with unselected candidate: %s", plan.CanonicalString())
				}
				// Do not replay the cancelled collaborator merely to validate the
				// retained physical packing; the live pure objective is a control.
				if checked := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal.base)); !checked.Valid() {
					t.Fatalf("retained publication invalid: %v", checked.Violations())
				}
			})
		}
	}
}

func TestHeuristicBudgetRefusalDoesNotPublishDetachedBorrowedObjective(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(1))
	goal := preferCount{count: 1}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	instances := []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}
	control, err := (solver.Heuristic{}).PackFixed(ctx, base, instances, solver.Options{PlanObjective: goal})
	if err != nil || len(control.Placements()) != 1 {
		t.Fatalf("live borrowed-objective control failed: %s, %v", control.CanonicalString(), err)
	}
	limits := base.Limits()
	limits.MaxCandidatePlacements = 1
	plan, err := (solver.Heuristic{}).PackFixed(ctx, base.WithLimits(limits), instances, solver.Options{PlanObjective: goal})
	if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != "" {
		t.Fatalf("resource refusal published a detached borrowed-objective result: %s, %v", plan.CanonicalString(), err)
	}
}
