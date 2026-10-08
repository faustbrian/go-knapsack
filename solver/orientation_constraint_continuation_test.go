package solver_test

import (
	"context"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

type requireWideOrientation struct{}

func (requireWideOrientation) Check(_ context.Context, view constraint.PlacementView) constraint.Decision {
	if view.Candidate().Dimensions.X == 2 {
		return constraint.Accept()
	}
	return constraint.Reject("orientation", "placement must use the wider X dimension")
}

func TestFixedPackersContinueAfterRejectedOrientation(t *testing.T) {
	baseline := exactRequest(t, 2, 1)
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: []knapsack.NormalizedItem{{
			ID: "item", Dimensions: geometry.Dimensions{X: 1, Y: 2, Z: 1}, Weight: 1,
			Orientations: []geometry.Orientation{geometry.OrientationXYZ, geometry.OrientationYXZ},
		}},
		Containers: []knapsack.NormalizedContainer{{
			ID: "box", Dimensions: geometry.Dimensions{X: 2, Y: 2, Z: 1},
			MaxContentWeight: 1, Stock: knapsack.FiniteStock(1),
		}},
		Resolution: baseline.Resolution(), Limits: baseline.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	instances := []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}
	callback := requireWideOrientation{}
	for name, pack := range map[string]func(context.Context, knapsack.NormalizedRequest, []knapsack.ContainerInstance, solver.Options) (knapsack.Plan, error){
		"heuristic": (solver.Heuristic{}).PackFixed,
		"exact":     (solver.Exact{}).PackFixed,
	} {
		t.Run(name, func(t *testing.T) {
			plan, err := pack(t.Context(), request, instances, solver.Options{Constraints: []constraint.Placement{callback}})
			placements := plan.Placements()
			if err != nil || len(placements) != 1 || len(plan.UnpackedItemIDs()) != 0 {
				t.Fatalf("later accepted orientation: placements=%d unpacked=%d error=%v", len(placements), len(plan.UnpackedItemIDs()), err)
			}
			placement := placements[0]
			if placement.ItemID != "item" || placement.ContainerID != "box#1" ||
				placement.Origin != (geometry.Point{}) || placement.Orientation != geometry.OrientationYXZ ||
				placement.Dimensions != (geometry.Dimensions{X: 2, Y: 1, Z: 1}) {
				t.Fatalf("incorrect accepted alternative: %+v", placement)
			}
			if result := verify.Plan(request, plan, verify.RequireAll().WithConstraints(callback)); !result.Valid() {
				t.Fatalf("accepted alternative violates packing constraints: %+v", result.Violations())
			}
		})
	}
}
