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

func unitSearchRequest(t *testing.T, grouped bool) knapsack.NormalizedRequest {
	t.Helper()
	dims := knapsack.PhysicalDimensions{X: quantity(1, measurement.Metre), Y: quantity(1, measurement.Metre), Z: quantity(1, measurement.Metre)}
	ids := []string{"item"}
	if grouped {
		ids = []string{"a", "b", "c", "d"}
	}
	items := make([]knapsack.Item, len(ids))
	for index, id := range ids {
		group := ""
		if grouped && id != "d" {
			group = "linked"
		}
		var err error
		items[index], err = knapsack.NewItem(knapsack.ItemSpec{ID: id, Group: group, Dimensions: dims,
			Weight: quantity(1, measurement.Kilogram), Orientations: []geometry.Orientation{geometry.OrientationXYZ}})
		if err != nil {
			t.Fatal(err)
		}
	}
	box, err := knapsack.NewContainerType(knapsack.ContainerTypeSpec{ID: "box", InternalDimensions: dims,
		MaxContentWeight: quantity(1, measurement.Kilogram), Stock: knapsack.FiniteStock(1)})
	if err != nil {
		t.Fatal(err)
	}
	request, err := knapsack.NewRequest(items, []knapsack.ContainerType{box}, knapsack.Resolution{
		Length: quantity(1, measurement.Metre), Mass: quantity(1, measurement.Kilogram)}, knapsack.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return request.Normalized()
}

func TestHeuristicPreservesCompletePlanAtCandidateAndRepackLimits(t *testing.T) {
	t.Parallel()
	request := unitSearchRequest(t, false)
	for _, test := range []struct {
		name              string
		fixed, improve    bool
		limit, candidates uint64
		rounds            uint32
		exhausted         bool
	}{
		{"fixed equality", true, false, 1, 1, 0, true},
		{"fixed sufficient", true, false, 2, 1, 0, false},
		{"repack equality", false, true, 2, 2, 1, true},
		{"repack sufficient", false, true, 3, 2, 1, false},
		{"no repack", false, false, 2, 1, 0, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			limits := request.Limits()
			limits.MaxCandidatePlacements = test.limit
			if !test.improve {
				limits.MaxImprovementRounds = 0
			}
			scoped := request.WithLimits(limits)
			var plan knapsack.Plan
			var err error
			if test.fixed {
				plan, err = (solver.Heuristic{}).PackFixed(context.Background(), scoped,
					[]knapsack.ContainerInstance{{ID: "fixed", TypeID: "box"}}, solver.Options{})
			} else {
				plan, err = (solver.Heuristic{}).PackAll(context.Background(), scoped, solver.Options{})
			}
			status, termination := knapsack.StatusFeasible, knapsack.TerminationCompleted
			if test.exhausted {
				status, termination = knapsack.StatusBudgetExhausted, knapsack.TerminationCandidateLimit
				if !errors.Is(err, knapsack.ErrBudgetExhausted) {
					t.Fatalf("expected candidate exhaustion, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if plan.Status() != status || plan.Termination() != termination ||
				plan.Work().CandidatePlacements != test.candidates || plan.Work().ImprovementRounds != test.rounds ||
				len(plan.Placements()) != 1 || len(plan.UnpackedItemIDs()) != 0 {
				t.Fatalf("incorrect retained plan or work: %s %+v", plan.CanonicalString(), plan.Work())
			}
			if result := verify.Plan(scoped, plan, verify.RequireAll()); !result.Valid() {
				t.Fatalf("invalid retained plan: %+v", result.Violations())
			}
		})
	}
}

func TestHeuristicGroupRollbackRestoresFiniteStockForLaterItem(t *testing.T) {
	t.Parallel()
	request := unitSearchRequest(t, true)
	limits := request.Limits()
	limits.MaxImprovementRounds = 0
	request = request.WithLimits(limits)
	plan, err := (solver.Heuristic{}).PackAll(context.Background(), request, solver.Options{AllowUnpacked: true})
	if err != nil {
		t.Fatal(err)
	}
	placements, containers := plan.Placements(), plan.Containers()
	if plan.Status() != knapsack.StatusBestKnown || plan.Termination() != knapsack.TerminationNoPlacement ||
		len(placements) != 1 || placements[0].ItemID != "d" || len(containers) != 1 ||
		containers[0].ID != "box#000001" || placements[0].ContainerID != containers[0].ID ||
		!slices.Equal(plan.UnpackedItemIDs(), []string{"a", "b", "c"}) || plan.Work().ImprovementRounds != 0 {
		t.Fatalf("rollback did not restore stock/accounting: %s", plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
		t.Fatalf("invalid partial plan: %+v", result.Violations())
	}
}

func TestExactPackAllAggregatesWorkAcrossFailedAndFeasibleConfigurations(t *testing.T) {
	t.Parallel()
	dims := knapsack.PhysicalDimensions{X: quantity(2, measurement.Metre), Y: quantity(1, measurement.Metre), Z: quantity(1, measurement.Metre)}
	item, err := knapsack.NewItem(knapsack.ItemSpec{ID: "item", Dimensions: dims,
		Weight: quantity(1, measurement.Kilogram), Orientations: []geometry.Orientation{geometry.OrientationXYZ}})
	if err != nil {
		t.Fatal(err)
	}
	containers := make([]knapsack.ContainerType, 2)
	for index, id := range []string{"small", "large"} {
		boxDims := dims
		boxDims.X = quantity(int64(index+1), measurement.Metre)
		containers[index], err = knapsack.NewContainerType(knapsack.ContainerTypeSpec{
			ID: id, InternalDimensions: boxDims, MaxContentWeight: quantity(1, measurement.Kilogram), Stock: knapsack.FiniteStock(1)})
		if err != nil {
			t.Fatal(err)
		}
	}
	raw, err := knapsack.NewRequest([]knapsack.Item{item}, containers, knapsack.Resolution{
		Length: quantity(1, measurement.Metre), Mass: quantity(1, measurement.Kilogram)}, knapsack.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	request := raw.Normalized()
	plan, err := (solver.Exact{}).PackAll(context.Background(), request, solver.Options{})
	if err != nil || plan.Status() != knapsack.StatusOptimal || plan.Termination() != knapsack.TerminationCompleted ||
		len(plan.Containers()) != 1 || plan.Containers()[0].TypeID != "large" ||
		plan.Work().Nodes != 3 || plan.Work().Branches != 4 || plan.Work().CandidatePlacements != 2 {
		t.Fatalf("incorrect aggregate search: error=%v plan=%s work=%+v", err, plan.CanonicalString(), plan.Work())
	}
	if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
		t.Fatalf("invalid optimal plan: %+v", result.Violations())
	}
	for _, test := range []struct {
		name        string
		limit       func(*knapsack.Limits)
		termination knapsack.TerminationReason
	}{
		{"nodes", func(l *knapsack.Limits) { l.MaxSearchNodes = 2 }, knapsack.TerminationNodeLimit},
		{"branches", func(l *knapsack.Limits) { l.MaxBranches = 3 }, knapsack.TerminationBranchLimit},
		{"candidates", func(l *knapsack.Limits) { l.MaxCandidatePlacements = 1 }, knapsack.TerminationCandidateLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			limits := request.Limits()
			test.limit(&limits)
			limited, err := (solver.Exact{}).PackAll(context.Background(), request.WithLimits(limits), solver.Options{})
			if !errors.Is(err, knapsack.ErrBudgetExhausted) || errors.Is(err, knapsack.ErrProvenInfeasible) ||
				limited.Status() != knapsack.StatusBudgetExhausted || limited.Termination() != test.termination {
				t.Fatalf("incorrect exhausted classification: error=%v status=%s termination=%s", err, limited.Status(), limited.Termination())
			}
		})
	}
	if impossible, err := (solver.Exact{}).PackFixed(context.Background(), request,
		[]knapsack.ContainerInstance{{ID: "small-only", TypeID: "small"}}, solver.Options{}); !errors.Is(err, knapsack.ErrProvenInfeasible) || impossible.Status() != knapsack.StatusInfeasible {
		t.Fatalf("impossible control did not prove infeasibility: error=%v status=%s", err, impossible.Status())
	}
}
