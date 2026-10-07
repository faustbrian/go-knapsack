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

func TestPackFixedAcceptsProportionalSharedSupportLoad(t *testing.T) {
	t.Parallel()
	limit := quantity(2, measurement.Kilogram)
	items := make([]knapsack.Item, 0, 3)
	for _, id := range []string{"a", "b", "top"} {
		spec := knapsack.ItemSpec{
			ID: id,
			Dimensions: knapsack.PhysicalDimensions{
				X: quantity(1, measurement.Metre), Y: quantity(1, measurement.Metre), Z: quantity(2, measurement.Metre),
			},
			Weight:             quantity(5, measurement.Kilogram),
			Orientations:       []geometry.Orientation{geometry.OrientationXYZ},
			MaxSupportedWeight: &limit,
		}
		if id == "top" {
			// Prevent an alternative solution with both supports above the top.
			spec.FragileTop = true
			spec.Dimensions.X, spec.Dimensions.Z = quantity(2, measurement.Metre), quantity(1, measurement.Metre)
			spec.Weight, spec.MaxSupportedWeight, spec.MinimumSupportPPM = quantity(4, measurement.Kilogram), nil, 1_000_000
		}
		item, err := knapsack.NewItem(spec)
		if err != nil {
			t.Fatal(err)
		}
		items = append(items, item)
	}
	box, err := knapsack.NewContainerType(knapsack.ContainerTypeSpec{
		ID: "box",
		InternalDimensions: knapsack.PhysicalDimensions{
			X: quantity(2, measurement.Metre), Y: quantity(1, measurement.Metre), Z: quantity(3, measurement.Metre),
		},
		MaxContentWeight: quantity(14, measurement.Kilogram), Stock: knapsack.FiniteStock(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := knapsack.NewRequest(items, []knapsack.ContainerType{box}, knapsack.Resolution{
		Length: quantity(1, measurement.Metre), Mass: quantity(1, measurement.Kilogram),
	}, knapsack.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for name, pack := range map[string]func(context.Context, knapsack.NormalizedRequest, []knapsack.ContainerInstance, solver.Options) (knapsack.Plan, error){
		"exact":     (solver.Exact{}).PackFixed,
		"heuristic": (solver.Heuristic{}).PackFixed,
	} {
		t.Run(name, func(t *testing.T) {
			plan, err := pack(context.Background(), request.Normalized(), []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{})
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Placements()) != 3 || len(plan.UnpackedItemIDs()) != 0 {
				t.Fatalf("shared load prevented complete packing: %s", plan.CanonicalString())
			}
			if result := verify.Plan(request.Normalized(), plan, verify.RequireAll()); !result.Valid() {
				t.Fatalf("invalid shared-load plan: %+v", result.Violations())
			}
			for _, placement := range plan.Placements() {
				if placement.ItemID == "top" && (placement.Origin.Z != 2 || !slices.Equal(placement.SupporterIDs, []string{"a", "b"})) {
					t.Fatalf("incorrect load-sharing geometry: %+v", placement)
				}
			}
		})
	}
}
