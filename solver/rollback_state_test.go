package solver

import (
	"reflect"
	"slices"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestGroupRollbackRebuildsRetainedStateAndContinues(t *testing.T) {
	t.Parallel()
	for _, discardEmpty := range []bool{false, true} {
		name := "fixed"
		if discardEmpty {
			name = "variable"
		}
		t.Run(name, func(t *testing.T) {
			removed, first, second := baseInternalItem(), baseInternalItem(), baseInternalItem()
			removed.ID, removed.Group, removed.Weight = "gone-mixed", "linked", 2
			first.ID, first.Weight = "keep-first", 5
			first.Dimensions = geometry.Dimensions{X: 2, Y: 3, Z: 4}
			second.ID, second.Weight = "keep-second", 7
			second.Dimensions = geometry.Dimensions{X: 1, Y: 1, Z: 2}
			firstPlacement := knapsack.Placement{ItemID: first.ID, ContainerID: "mixed", Origin: geometry.Point{X: 1},
				Orientation: geometry.OrientationXYZ, Dimensions: first.Dimensions, Weight: first.Weight}
			secondPlacement := knapsack.Placement{ItemID: second.ID, ContainerID: "mixed", Origin: geometry.Point{X: 3},
				Orientation: geometry.OrientationXYZ, Dimensions: second.Dimensions, Weight: second.Weight}
			mixed := baseInternalBin()
			mixed.instance.ID = "mixed"
			mixed.info.Dimensions = geometry.Dimensions{X: 4, Y: 3, Z: 4}
			mixed.info.MaxContentWeight, mixed.info.Stock = 20, knapsack.FiniteStock(2)
			mixed.items = []knapsack.NormalizedItem{removed, first, second}
			mixed.placements = []knapsack.Placement{
				{ItemID: removed.ID, ContainerID: "mixed", Orientation: geometry.OrientationXYZ, Dimensions: removed.Dimensions, Weight: removed.Weight},
				firstPlacement, secondPlacement,
			}
			mixed.weight = 14
			mixed.points = []geometry.Point{{}, {X: 1}, {X: 3}, {X: 4}, {Y: 1}, {X: 3, Y: 1}, {X: 1, Y: 3}, {Z: 1}, {X: 3, Z: 2}, {X: 1, Z: 4}}

			empty := baseInternalBin()
			empty.instance.ID = "empty"
			empty.info.Stock = knapsack.FiniteStock(2)
			emptyItem := baseInternalItem()
			emptyItem.ID, emptyItem.Group, emptyItem.Weight = "gone-first", "linked", 3
			empty.items = []knapsack.NormalizedItem{emptyItem}
			empty.placements = []knapsack.Placement{{ItemID: emptyItem.ID, ContainerID: "empty", Orientation: geometry.OrientationXYZ,
				Dimensions: emptyItem.Dimensions, Weight: emptyItem.Weight}}
			empty.weight = 3
			empty.points = []geometry.Point{{}, {X: 1}, {Y: 1}, {Z: 1}}
			var stock map[string]uint32
			if discardEmpty {
				stock = map[string]uint32{"box": 2}
			}

			remaining, removedIDs := rollbackGroup([]*bin{empty, mixed}, "linked", discardEmpty, stock)
			wantBins := []*bin{empty, mixed}
			if discardEmpty {
				wantBins = []*bin{mixed}
			}
			if !slices.Equal(remaining, wantBins) || !slices.Equal(removedIDs, []string{"gone-first", "gone-mixed"}) {
				t.Fatalf("rollback bin/item conservation: bins=%d removed=%v", len(remaining), removedIDs)
			}
			wantPoints := []geometry.Point{{}, {X: 3}, {X: 4}, {X: 3, Y: 1}, {X: 1, Y: 3}, {X: 3, Z: 2}, {X: 1, Z: 4}}
			wantItems := []knapsack.NormalizedItem{first, second}
			wantPlacements := []knapsack.Placement{firstPlacement, secondPlacement}
			check := func() {
				t.Helper()
				if mixed.weight != 12 || !slices.Equal(mixed.points, wantPoints) ||
					!reflect.DeepEqual(mixed.items, wantItems) || !reflect.DeepEqual(mixed.placements, wantPlacements) {
					t.Fatalf("retained state: weight=%d points=%+v items=%+v placements=%+v", mixed.weight, mixed.points, mixed.items, mixed.placements)
				}
				if empty.weight != 0 || len(empty.items) != 0 || len(empty.placements) != 0 || !slices.Equal(empty.points, []geometry.Point{{}}) {
					t.Fatalf("emptied state: weight=%d points=%+v", empty.weight, empty.points)
				}
				if discardEmpty && stock["box"] != 1 {
					t.Fatalf("retained instance lost stock: %+v", stock)
				}
			}
			check()
			repeated, repeatedRemoved := rollbackGroup(remaining, "linked", discardEmpty, stock)
			if !slices.Equal(repeated, wantBins) || len(repeatedRemoved) != 0 {
				t.Fatalf("repeat rollback changed bins/items: bins=%d removed=%v", len(repeated), repeatedRemoved)
			}
			check()
		})
	}
}
