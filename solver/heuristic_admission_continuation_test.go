package solver_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

type requiredOrigins map[string]geometry.Point

func (origins requiredOrigins) Check(_ context.Context, view constraint.PlacementView) constraint.Decision {
	placement := view.Candidate()
	if placement.Origin == origins[placement.ItemID] {
		return constraint.Accept()
	}
	return constraint.Reject("origin", "placement must use its eligible position")
}

type acceptAnyPlacement struct{}

func (acceptAnyPlacement) Check(context.Context, constraint.PlacementView) constraint.Decision {
	return constraint.Accept()
}

func TestHeuristicContinuesAfterOverlappingOrientation(t *testing.T) {
	baseline := exactRequest(t, 5, 1)
	limits := baseline.Limits()
	limits.MaxImprovementRounds = 0
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: []knapsack.NormalizedItem{
			{ID: "anchor-a", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 2}, Weight: 3, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
			{ID: "anchor-b", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 2}, Weight: 2, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
			{ID: "moving", Dimensions: geometry.Dimensions{X: 2, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ, geometry.OrientationYXZ}},
		},
		Containers: []knapsack.NormalizedContainer{{
			ID: "box", Dimensions: geometry.Dimensions{X: 5, Y: 2, Z: 2}, MaxContentWeight: 6,
			Stock:           knapsack.FiniteStock(1),
			CenterOfGravity: &knapsack.CenterOfGravityBounds{MaxXPPM: 1_000_000, MaxYPPM: 1_000_000, MaxZPPM: 1_000_000},
		}},
		Resolution: baseline.Resolution(), Limits: limits,
	})
	if err != nil {
		t.Fatal(err)
	}
	origins := requiredOrigins{"anchor-a": {}, "anchor-b": {X: 2}, "moving": {X: 1}}
	checkHeuristicAdmissionPlan(t, request, solver.Options{Constraints: []constraint.Placement{origins}}, map[string]knapsack.Placement{
		"anchor-a": {Origin: origins["anchor-a"], Orientation: geometry.OrientationXYZ, Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 2}},
		"anchor-b": {Origin: origins["anchor-b"], Orientation: geometry.OrientationXYZ, Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 2}},
		"moving":   {Origin: origins["moving"], Orientation: geometry.OrientationYXZ, Dimensions: geometry.Dimensions{X: 1, Y: 2, Z: 1}},
	})
}

func TestHeuristicContinuesAfterUnsupportedOrientation(t *testing.T) {
	baseline := exactRequest(t, 2, 1)
	limits := baseline.Limits()
	limits.MaxImprovementRounds = 0
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: []knapsack.NormalizedItem{
			{ID: "support", Dimensions: geometry.Dimensions{X: 1, Y: 2, Z: 1}, Weight: 2, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
			{ID: "moving", Dimensions: geometry.Dimensions{X: 2, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ, geometry.OrientationYXZ}, MinimumSupportPPM: 1_000_000},
		},
		Containers: []knapsack.NormalizedContainer{{ID: "box", Dimensions: geometry.Dimensions{X: 2, Y: 2, Z: 2}, MaxContentWeight: 3, Stock: knapsack.FiniteStock(1)}},
		Resolution: baseline.Resolution(), Limits: limits,
	})
	if err != nil {
		t.Fatal(err)
	}
	origins := requiredOrigins{"support": {}, "moving": {Z: 1}}
	checkHeuristicAdmissionPlan(t, request, solver.Options{Constraints: []constraint.Placement{origins}}, map[string]knapsack.Placement{
		"support": {Origin: origins["support"], Orientation: geometry.OrientationXYZ, Dimensions: geometry.Dimensions{X: 1, Y: 2, Z: 1}},
		"moving":  {Origin: origins["moving"], Orientation: geometry.OrientationYXZ, Dimensions: geometry.Dimensions{X: 1, Y: 2, Z: 1}, SupporterIDs: []string{"support"}},
	})
}

func TestHeuristicWithoutCallbacksDoesNotApplyCallbackViewLimit(t *testing.T) {
	baseline := exactRequest(t, 1, 1)
	limits := baseline.Limits()
	limits.MaxImprovementRounds = 0
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: []knapsack.NormalizedItem{{
			ID: "item", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1,
			Orientations: []geometry.Orientation{geometry.OrientationXYZ},
			Attributes:   map[string]string{"metadata": strings.Repeat("m", 16<<20)},
		}},
		Containers: []knapsack.NormalizedContainer{{ID: "box", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, MaxContentWeight: 1, Stock: knapsack.FiniteStock(1)}},
		Resolution: baseline.Resolution(), Limits: limits,
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, callbacks := range map[string][]constraint.Placement{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			checkHeuristicAdmissionPlan(t, request, solver.Options{Constraints: callbacks}, map[string]knapsack.Placement{
				"item": {Orientation: geometry.OrientationXYZ, Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}},
			})
		})
	}
	for name, pack := range heuristicAdmissionPackers(request, solver.Options{Constraints: []constraint.Placement{acceptAnyPlacement{}}}) {
		t.Run("callback/"+name, func(t *testing.T) {
			if _, err := pack(t.Context()); !errors.Is(err, constraint.ErrViewLimit) {
				t.Fatalf("explicit callback view limit: error=%v", err)
			}
		})
	}
}

func heuristicAdmissionPackers(request knapsack.NormalizedRequest, options solver.Options) map[string]func(context.Context) (knapsack.Plan, error) {
	return map[string]func(context.Context) (knapsack.Plan, error){
		"fixed": func(ctx context.Context) (knapsack.Plan, error) {
			return (solver.Heuristic{}).PackFixed(ctx, request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, options)
		},
		"variable": func(ctx context.Context) (knapsack.Plan, error) {
			return (solver.Heuristic{}).PackAll(ctx, request, options)
		},
	}
}

func checkHeuristicAdmissionPlan(t *testing.T, request knapsack.NormalizedRequest, options solver.Options, want map[string]knapsack.Placement) {
	t.Helper()
	for name, pack := range heuristicAdmissionPackers(request, options) {
		t.Run(name, func(t *testing.T) {
			plan, err := pack(t.Context())
			placements := plan.Placements()
			if err != nil || len(placements) != len(want) || len(plan.UnpackedItemIDs()) != 0 {
				t.Fatalf("complete admission: placements=%d unpacked=%d error=%v", len(placements), len(plan.UnpackedItemIDs()), err)
			}
			for _, placement := range placements {
				expected, ok := want[placement.ItemID]
				if !ok || placement.Origin != expected.Origin || placement.Orientation != expected.Orientation || placement.Dimensions != expected.Dimensions ||
					placement.ContainerID == "" || name == "fixed" && placement.ContainerID != "box#1" ||
					len(placement.SupporterIDs) != len(expected.SupporterIDs) {
					t.Fatalf("incorrect admitted placement: %+v", placement)
				}
				for index, supporter := range expected.SupporterIDs {
					if placement.SupporterIDs[index] != supporter {
						t.Fatalf("incorrect admitted supporters: %+v", placement.SupporterIDs)
					}
				}
			}
			if result := verify.Plan(request, plan, verify.RequireAll().WithConstraints(options.Constraints...)); !result.Valid() {
				t.Fatalf("admission violates packing constraints: %+v", result.Violations())
			}
		})
	}
}
