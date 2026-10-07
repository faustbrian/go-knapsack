package solver_test

import (
	"context"
	"slices"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
	"github.com/faustbrian/go-measurement/v2"
)

func groupTransitionDimensions(x int64) knapsack.PhysicalDimensions {
	return knapsack.PhysicalDimensions{X: quantity(x, measurement.Metre), Y: quantity(1, measurement.Metre), Z: quantity(1, measurement.Metre)}
}

func groupTransitionItem(t *testing.T, id, group, class string, width int64) knapsack.Item {
	t.Helper()
	item, err := knapsack.NewItem(knapsack.ItemSpec{ID: id, Group: group, Attributes: map[string]string{"class": class},
		Dimensions: groupTransitionDimensions(width), Weight: quantity(1, measurement.Kilogram),
		Orientations: []geometry.Orientation{geometry.OrientationXYZ}})
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func groupTransitionContainer(t *testing.T, id string, width, priority int64, classes []string) knapsack.ContainerType {
	t.Helper()
	container, err := knapsack.NewContainerType(knapsack.ContainerTypeSpec{ID: id,
		InternalDimensions: groupTransitionDimensions(width), MaxContentWeight: quantity(10, measurement.Kilogram),
		Stock: knapsack.FiniteStock(1), Priority: priority, AllowedClasses: classes})
	if err != nil {
		t.Fatal(err)
	}
	return container
}

func groupTransitionRequest(t *testing.T, items []knapsack.Item, containers []knapsack.ContainerType, diagnostics uint32) knapsack.NormalizedRequest {
	t.Helper()
	limits := knapsack.DefaultLimits()
	limits.MaxImprovementRounds = 0
	limits.MaxDiagnostics = diagnostics
	request, err := knapsack.NewRequest(items, containers, knapsack.Resolution{
		Length: quantity(1, measurement.Metre), Mass: quantity(1, measurement.Kilogram)}, limits)
	if err != nil {
		t.Fatal(err)
	}
	return request.Normalized()
}

func TestHeuristicSuccessfulGroupContinuesPastIneligibleContainer(t *testing.T) {
	t.Parallel()
	request := groupTransitionRequest(t, []knapsack.Item{
		groupTransitionItem(t, "a", "linked", "wanted", 1), groupTransitionItem(t, "b", "linked", "wanted", 1),
	}, []knapsack.ContainerType{
		groupTransitionContainer(t, "ineligible", 2, 0, []string{"other"}),
		groupTransitionContainer(t, "eligible", 2, 1, []string{"wanted"}),
	}, 10)
	for _, fixed := range []bool{false, true} {
		t.Run(map[bool]string{false: "all", true: "fixed"}[fixed], func(t *testing.T) {
			var plan knapsack.Plan
			var err error
			if fixed {
				plan, err = (solver.Heuristic{}).PackFixed(context.Background(), request,
					[]knapsack.ContainerInstance{{ID: "first", TypeID: "ineligible"}, {ID: "second", TypeID: "eligible"}}, solver.Options{})
			} else {
				plan, err = (solver.Heuristic{}).PackAll(context.Background(), request, solver.Options{})
			}
			if err != nil {
				t.Fatal(err)
			}
			placements := plan.Placements()
			if plan.Status() != knapsack.StatusFeasible || plan.Termination() != knapsack.TerminationCompleted ||
				len(placements) != 2 || len(plan.UnpackedItemIDs()) != 0 || plan.Work().CandidatePlacements != 6 {
				t.Fatalf("group placement/accounting: %s work=%+v", plan.CanonicalString(), plan.Work())
			}
			wantContainer := "eligible#000001"
			if fixed {
				wantContainer = "second"
			}
			for index, placement := range placements {
				if placement.ItemID != []string{"a", "b"}[index] || placement.ContainerID != wantContainer || placement.Origin != (geometry.Point{X: int64(index)}) {
					t.Fatalf("group split, omitted or misplaced: %+v", placements)
				}
			}
			if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
				t.Fatalf("invalid group plan: %+v", result.Violations())
			}
		})
	}
}

func TestHeuristicContinuesPastExhaustedType(t *testing.T) {
	t.Parallel()
	request := groupTransitionRequest(t, []knapsack.Item{
		groupTransitionItem(t, "a", "", "", 2), groupTransitionItem(t, "b", "", "", 1),
	}, []knapsack.ContainerType{
		groupTransitionContainer(t, "wide", 2, 0, nil), groupTransitionContainer(t, "unit", 1, 1, nil),
	}, 10)
	plan, err := (solver.Heuristic{}).PackAll(context.Background(), request, solver.Options{})
	if err != nil {
		t.Fatal(err)
	}
	p := plan.Placements()
	if plan.Status() != knapsack.StatusFeasible || plan.Termination() != knapsack.TerminationCompleted || len(p) != 2 ||
		len(plan.UnpackedItemIDs()) != 0 || plan.Work().CandidatePlacements != 7 || p[0].ItemID != "a" || p[0].ContainerID != "wide#000001" ||
		p[1].ItemID != "b" || p[1].ContainerID != "unit#000001" {
		t.Fatalf("exhausted stock continuation: %s work=%+v", plan.CanonicalString(), plan.Work())
	}
	if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
		t.Fatalf("invalid stock plan: %+v", result.Violations())
	}
}

func TestHeuristicFailedGroupPreservesDiagnosticCapAndLaterRecovery(t *testing.T) {
	t.Parallel()
	items := []knapsack.Item{groupTransitionItem(t, "e", "", "", 2)}
	for _, id := range []string{"a", "b", "c"} {
		items = append(items, groupTransitionItem(t, id, "linked", "", 1))
	}
	items = append(items, groupTransitionItem(t, "d", "", "", 1))
	request := groupTransitionRequest(t, items, []knapsack.ContainerType{groupTransitionContainer(t, "unit", 1, 0, nil)}, 1)
	for _, fixed := range []bool{false, true} {
		t.Run(map[bool]string{false: "all", true: "fixed"}[fixed], func(t *testing.T) {
			var plan knapsack.Plan
			var err error
			if fixed {
				plan, err = (solver.Heuristic{}).PackFixed(context.Background(), request, []knapsack.ContainerInstance{{ID: "fixed", TypeID: "unit"}}, solver.Options{AllowUnpacked: true})
			} else {
				plan, err = (solver.Heuristic{}).PackAll(context.Background(), request, solver.Options{AllowUnpacked: true})
			}
			if err != nil {
				t.Fatal(err)
			}
			p := plan.Placements()
			if plan.Status() != knapsack.StatusBestKnown || plan.Termination() != knapsack.TerminationNoPlacement || len(p) != 1 || p[0].ItemID != "d" ||
				!slices.Equal(plan.UnpackedItemIDs(), []string{"a", "b", "c", "e"}) || plan.Work().CandidatePlacements != 7 {
				t.Fatalf("failed group recovery/accounting: %s work=%+v", plan.CanonicalString(), plan.Work())
			}
			if !fixed {
				diagnostics := plan.Diagnostics()
				if len(diagnostics) != 1 || diagnostics[0].Code != "no_feasible_placement" || diagnostics[0].ItemID != "e" {
					t.Fatalf("diagnostic cap/identity: %+v", diagnostics)
				}
			}
			if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
				t.Fatalf("invalid partial plan: %+v", result.Violations())
			}
		})
	}
}
