package solver_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/objective"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func TestExactVariableGridAdmissionIsResourceRefusalNotProof(t *testing.T) {
	base := completenessRequest(t, 2, knapsack.FiniteStock(1))
	// Two x coordinates, one y and one z: 32 transient coordinate bytes plus
	// two three-int64 points require 80 bytes after the normal working reservation.
	reservation := 2*base.MemoryBytes() + 2816
	for _, allowance := range []uint64{79, 80} {
		limits := base.Limits()
		limits.MaxMemoryBytes = reservation + allowance
		request := base.WithLimits(limits)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{})
		cancel()
		if allowance == 79 {
			if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationMemoryLimit || len(plan.Placements()) != 0 || len(plan.UnpackedItemIDs()) != 1 {
				t.Fatalf("grid refusal became proof or lost item: %s,%v", plan.CanonicalString(), err)
			}
		} else {
			if err != nil || plan.Status() != knapsack.StatusOptimal || len(plan.Placements()) != 1 {
				t.Fatalf("admitted grid did not complete: %s,%v", plan.CanonicalString(), err)
			}
			if checked := verify.Plan(request, plan, verify.RequireAll()); !checked.Valid() {
				t.Fatalf("invalid admitted grid: %v", checked.Violations())
			}
		}
	}
}

func TestExactCountGrowthMemoryStopKeepsEarlierFeasibleState(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(3))
	limits := base.Limits()
	limits.MaxMemoryBytes = 2*base.MemoryBytes() + 2816 + 48
	request := base.WithLimits(limits)
	goal := preferCount{count: 3}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: goal})
	if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationMemoryLimit || len(plan.Containers()) != 1 || len(plan.Placements()) != 1 {
		t.Fatalf("count storage refusal lost earlier state or claimed optimum: %s,%v", plan.CanonicalString(), err)
	}
	if checked := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !checked.Valid() {
		t.Fatalf("invalid retained state: %v", checked.Violations())
	}
}

func TestExactAggregatePartialRefusalPreservesOwnedNodeInterruption(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(2))
	items := []knapsack.NormalizedItem{
		{ID: "first", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
		{ID: "second", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
	}
	containers := base.Containers()
	containers[0].MaxContentWeight = math.MaxInt64
	limits := base.Limits()
	limits.MaxSearchNodes = 1
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: items, Containers: containers, Resolution: base.Resolution(), Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{})
	if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationNodeLimit || len(plan.Placements()) != 0 || len(plan.UnpackedItemIDs()) != 2 {
		t.Fatalf("unrepresentable partial erased actual interruption: %s,%v", plan.CanonicalString(), err)
	}
	if checked := verify.Plan(request, plan, verify.AllowUnpacked()); !checked.Valid() {
		t.Fatalf("invalid admitted empty interruption state: %v", checked.Violations())
	}
}

// A complete candidate retained after its own budget stop may improve on an
// earlier configuration. Failure is injected by that observable candidate state,
// not an invocation count or private call sequence.
type interruptedCandidateGoal struct{ failure error }

func (interruptedCandidateGoal) Valid() bool { return true }
func (interruptedCandidateGoal) score(plan knapsack.Plan) int {
	for _, container := range plan.Containers() {
		if container.TypeID == "b-better" {
			return 0
		}
	}
	return 1
}
func (goal interruptedCandidateGoal) Components(_ context.Context, _ knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	return []knapsack.ScoreComponent{{Name: "candidate_cost", Direction: "min", Unit: "integer", Value: strconv.Itoa(goal.score(plan))}}, nil
}
func (goal interruptedCandidateGoal) ComparePlans(_ context.Context, _ knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	if goal.failure != nil && left.Status() == knapsack.StatusBudgetExhausted {
		return 0, goal.failure
	}
	return goal.score(left) - goal.score(right), nil
}

func TestExactInterruptedCandidateComparisonPreservesBetterStateAndFailureOrigin(t *testing.T) {
	base := completenessRequest(t, 2, knapsack.FiniteStock(1))
	containers := base.Containers()
	containers[0].ID = "a-earlier"
	later := containers[0]
	later.ID = "b-better"
	limits := base.Limits()
	limits.MaxSearchNodes = 5
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: base.Items(), Containers: append(containers, later), Resolution: base.Resolution(), Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	goal := interruptedCandidateGoal{}
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: goal})
	if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || len(plan.Placements()) != 1 || len(plan.Containers()) != 1 || plan.Containers()[0].TypeID != "b-better" {
		t.Fatalf("interrupted better candidate lost: %s,%v", plan.CanonicalString(), err)
	}
	if checked := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !checked.Valid() {
		t.Fatalf("invalid better retained candidate: %v", checked.Violations())
	}
	for _, classification := range []error{knapsack.ErrBudgetExhausted, knapsack.ErrMemoryBudgetExhausted, knapsack.ErrProvenInfeasible, context.Canceled, context.DeadlineExceeded, objective.ErrInvalidObjective} {
		owned := fmt.Errorf("interrupted-candidate objective: %w", classification)
		plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: interruptedCandidateGoal{failure: owned}})
		if !errors.Is(err, owned) || plan.Status() != "" || len(plan.Placements()) != 0 || ctx.Err() != nil {
			t.Fatalf("live-context collaborator error became solver evidence: %s,%v", plan.CanonicalString(), err)
		}
	}
}

// Final aggregate publication is an observable feasible variable-search record;
// borrowed objectives must retain error provenance at that reporting boundary.
type aggregatePublicationGoal struct{ failure error }

func (aggregatePublicationGoal) Valid() bool { return true }
func (goal aggregatePublicationGoal) ComparePlans(ctx context.Context, request knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	return (preferCount{count: 2}).ComparePlans(ctx, request, left, right)
}
func (goal aggregatePublicationGoal) Components(ctx context.Context, request knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	if goal.failure != nil && plan.Status() == knapsack.StatusFeasible && plan.Work().Strategy == "exhaustive_variable_container" {
		return nil, goal.failure
	}
	return (preferCount{count: 2}).Components(ctx, request, plan)
}
func TestExactAggregatePublicationPreservesObjectiveFailureOrigin(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(2))
	containers := base.Containers()
	containers[0].MaxContentWeight = math.MaxInt64
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: base.Items(), Containers: containers, Resolution: base.Resolution(), Limits: base.Limits()})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	control, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: aggregatePublicationGoal{}})
	if !errors.Is(err, knapsack.ErrOverflow) || control.Status() != knapsack.StatusFeasible || len(control.Placements()) != 1 {
		t.Fatalf("aggregate control falsely proved optimality/lost feasible state: %s,%v", control.CanonicalString(), err)
	}
	for _, classification := range []error{knapsack.ErrBudgetExhausted, knapsack.ErrMemoryBudgetExhausted, knapsack.ErrProvenInfeasible, objective.ErrInvalidObjective} {
		owned := fmt.Errorf("aggregate-publication objective: %w", classification)
		plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: aggregatePublicationGoal{failure: owned}})
		if !errors.Is(err, owned) || plan.Status() != "" || len(plan.Placements()) != 0 || ctx.Err() != nil {
			t.Fatalf("late objective failure became aggregate/search evidence: %s,%v", plan.CanonicalString(), err)
		}
	}
}
