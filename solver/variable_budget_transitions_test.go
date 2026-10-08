package solver_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
	"github.com/faustbrian/go-measurement/v2"
)

func orientationBudgetRequest(t *testing.T, boxes int) knapsack.NormalizedRequest {
	t.Helper()
	item, err := knapsack.NewItem(knapsack.ItemSpec{ID: "item", Dimensions: knapsack.PhysicalDimensions{
		X: quantity(2, measurement.Metre), Y: quantity(1, measurement.Metre), Z: quantity(1, measurement.Metre)},
		Weight: quantity(1, measurement.Kilogram), Orientations: []geometry.Orientation{geometry.OrientationXYZ, geometry.OrientationYXZ}})
	if err != nil {
		t.Fatal(err)
	}
	containers := make([]knapsack.ContainerType, boxes)
	for index := range containers {
		containers[index], err = knapsack.NewContainerType(knapsack.ContainerTypeSpec{ID: string(rune('a' + index)),
			InternalDimensions: knapsack.PhysicalDimensions{X: quantity(2, measurement.Metre), Y: quantity(2, measurement.Metre), Z: quantity(1, measurement.Metre)},
			MaxContentWeight:   quantity(1, measurement.Kilogram), Stock: knapsack.FiniteStock(1)})
		if err != nil {
			t.Fatal(err)
		}
	}
	raw, err := knapsack.NewRequest([]knapsack.Item{item}, containers, knapsack.Resolution{
		Length: quantity(1, measurement.Metre), Mass: quantity(1, measurement.Kilogram)}, knapsack.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return raw.Normalized()
}

type cancelOnCompleteObjective struct{ cancel context.CancelFunc }

func (cancelOnCompleteObjective) Valid() bool { return true }
func (cancelOnCompleteObjective) ComparePlans(ctx context.Context, _ knapsack.NormalizedRequest, _, _ knapsack.Plan) (int, error) {
	return 0, ctx.Err()
}
func (o cancelOnCompleteObjective) Components(ctx context.Context, _ knapsack.NormalizedRequest, _ knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	o.cancel()
	return nil, ctx.Err()
}

func TestExactVariableSearchConservesItemsAfterObjectiveCancellation(t *testing.T) {
	t.Parallel()
	request := unitSearchRequest(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: cancelOnCompleteObjective{cancel: cancel}})
	if !errors.Is(err, context.Canceled) || plan.Status() != knapsack.StatusBudgetExhausted ||
		plan.Termination() != knapsack.TerminationCancelled || len(plan.Placements()) != 0 ||
		!slices.Equal(plan.UnpackedItemIDs(), []string{"item"}) {
		t.Fatalf("cancelled objective lost item conservation: error=%v plan=%s", err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
		t.Fatalf("invalid cancelled plan: %+v", result.Violations())
	}
}

func TestExactVariableLowerBoundUsesGrossCapacityBeforeSearch(t *testing.T) {
	t.Parallel()
	dims := knapsack.PhysicalDimensions{X: quantity(1, measurement.Metre), Y: quantity(1, measurement.Metre), Z: quantity(1, measurement.Metre)}
	item, err := knapsack.NewItem(knapsack.ItemSpec{ID: "item", Dimensions: dims,
		Weight: quantity(3, measurement.Kilogram), Orientations: []geometry.Orientation{geometry.OrientationXYZ}})
	if err != nil {
		t.Fatal(err)
	}
	tare, gross := quantity(1, measurement.Kilogram), quantity(3, measurement.Kilogram)
	box, err := knapsack.NewContainerType(knapsack.ContainerTypeSpec{ID: "box", InternalDimensions: dims,
		MaxContentWeight: quantity(3, measurement.Kilogram), TareWeight: &tare, MaxGrossWeight: &gross, Stock: knapsack.UnlimitedStock()})
	if err != nil {
		t.Fatal(err)
	}
	limits := knapsack.DefaultLimits()
	limits.MaxBranches = 1
	raw, err := knapsack.NewRequest([]knapsack.Item{item}, []knapsack.ContainerType{box}, knapsack.Resolution{
		Length: quantity(1, measurement.Metre), Mass: quantity(1, measurement.Kilogram)}, limits)
	if err != nil {
		t.Fatal(err)
	}
	request := raw.Normalized()
	plan, err := (solver.Exact{}).PackAll(context.Background(), request, solver.Options{})
	if !errors.Is(err, knapsack.ErrProvenInfeasible) || errors.Is(err, knapsack.ErrBudgetExhausted) ||
		plan.Status() != knapsack.StatusInfeasible || plan.Termination() != knapsack.TerminationCompleted ||
		len(plan.Placements()) != 0 || !slices.Equal(plan.UnpackedItemIDs(), []string{"item"}) {
		t.Fatalf("gross bound lost immediate infeasibility proof: error=%v plan=%s", err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
		t.Fatalf("invalid infeasible plan: %+v", result.Violations())
	}
}

func TestExactVariableSearchRetainsCompletePlanWithinSharedBudgets(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name            string
		boxes           int
		limit           func(*knapsack.Limits)
		termination     knapsack.TerminationReason
		nodes, branches uint64
		candidates      uint64
	}{
		{"first configuration branch limit", 1, func(l *knapsack.Limits) { l.MaxBranches = 2 }, knapsack.TerminationBranchLimit, 2, 3, 2},
		{"later configuration branch limit", 2, func(l *knapsack.Limits) { l.MaxBranches = 5 }, knapsack.TerminationBranchLimit, 5, 6, 4},
		{"later configuration candidate limit", 2, func(l *knapsack.Limits) { l.MaxCandidatePlacements = 3 }, knapsack.TerminationCandidateLimit, 5, 6, 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := orientationBudgetRequest(t, test.boxes)
			control, err := (solver.Exact{}).PackAll(context.Background(), request, solver.Options{})
			if err != nil || control.Status() != knapsack.StatusOptimal {
				t.Fatalf("unrestricted control: error=%v status=%s", err, control.Status())
			}
			limits := request.Limits()
			test.limit(&limits)
			request = request.WithLimits(limits)
			plan, err := (solver.Exact{}).PackAll(context.Background(), request, solver.Options{})
			if !errors.Is(err, knapsack.ErrBudgetExhausted) || errors.Is(err, knapsack.ErrProvenInfeasible) ||
				plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != test.termination {
				t.Fatalf("incorrect exhaustion: error=%v plan=%s", err, plan.CanonicalString())
			}
			work := plan.Work()
			if work.Nodes != test.nodes || work.Branches != test.branches || work.CandidatePlacements != test.candidates ||
				len(plan.Placements()) != 1 || plan.Placements()[0].ItemID != "item" || len(plan.UnpackedItemIDs()) != 0 {
				t.Fatalf("lost retained plan or aggregate work: plan=%s work=%+v", plan.CanonicalString(), work)
			}
			if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
				t.Fatalf("invalid retained plan: %+v", result.Violations())
			}
		})
	}
}
