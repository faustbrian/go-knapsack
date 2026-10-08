package solver_test

import (
	"context"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
)

type rejectPlacement struct{}

func (rejectPlacement) Check(context.Context, constraint.PlacementView) constraint.Decision {
	return constraint.Reject("ineligible", "candidate is ineligible")
}

type cancelRejectedPlacement struct{ cancel context.CancelFunc }

func (predicate cancelRejectedPlacement) Check(context.Context, constraint.PlacementView) constraint.Decision {
	predicate.cancel()
	return constraint.Reject("cancelled", "owner cancelled the candidate")
}

func TestHeuristicRejectedCandidateDoesNotCancelOwner(t *testing.T) {
	baseline := exactRequest(t, 2, 1)
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items:      []knapsack.NormalizedItem{{ID: "item", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}}},
		Containers: []knapsack.NormalizedContainer{{ID: "box", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, MaxContentWeight: 1, Stock: knapsack.FiniteStock(1)}},
		Resolution: baseline.Resolution(), Limits: baseline.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	plan, err := (solver.Heuristic{}).PackFixed(ctx, request,
		[]knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}},
		solver.Options{AllowUnpacked: true, Constraints: []constraint.Placement{rejectPlacement{}, cancelRejectedPlacement{cancel: cancel}}})
	if err != nil {
		t.Fatalf("rejected candidate cancelled packing: %v", err)
	}
	if ctx.Err() != nil {
		t.Fatal("ineligible candidate cancelled its owner")
	}
	if len(plan.Placements()) != 0 {
		t.Fatal("ineligible candidate was placed")
	}
	unpacked := plan.UnpackedItemIDs()
	if len(unpacked) != 1 || unpacked[0] != "item" {
		t.Fatalf("ineligible item not preserved: %v", unpacked)
	}
}
