package solver

import (
	"slices"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestPlacementOrderingComparesPositiveEndpoints(t *testing.T) {
	t.Parallel()
	base := knapsack.Placement{
		Origin:      geometry.Point{X: 2, Y: 3, Z: 4},
		Dimensions:  geometry.Dimensions{X: 2, Y: 2, Z: 2},
		Orientation: geometry.OrientationXYZ,
	}
	for name, compare := range map[string]func(knapsack.Placement, knapsack.Placement) int{
		"depth first": comparePlacement,
		"width first": comparePlacementWidthFirst,
	} {
		t.Run(name, func(t *testing.T) {
			if got := compare(base, base); got != 0 {
				t.Fatalf("identical placement comparison = %d", got)
			}
			for _, axis := range []string{"x", "y", "z"} {
				t.Run(axis, func(t *testing.T) {
					next := base
					switch axis {
					case "x":
						next.Dimensions.X++
					case "y":
						next.Dimensions.Y++
					case "z":
						next.Dimensions.Z++
					}
					if compare(base, next) >= 0 || compare(next, base) <= 0 {
						t.Fatalf("endpoint ordering lost: base=%+v next=%+v", base, next)
					}
				})
			}
		})
	}
}

func TestExactPointsHonorsByteBoundaryAndEnumeratesGrid(t *testing.T) {
	t.Parallel()
	target := baseInternalBin()
	target.info.Dimensions = geometry.Dimensions{X: 2, Y: 2, Z: 2}
	target.info.CenterOfGravity = &knapsack.CenterOfGravityBounds{}
	want := []geometry.Point{
		{}, {X: 1}, {Y: 1}, {X: 1, Y: 1},
		{Z: 1}, {X: 1, Z: 1}, {Y: 1, Z: 1}, {X: 1, Y: 1, Z: 1},
	}
	// Six int64 coordinates plus eight three-int64 points require 240 bytes.
	points, ok := exactPoints(target, 240)
	if !ok || !slices.Equal(points, want) {
		t.Fatalf("exact-budget grid = %v, ok=%v", points, ok)
	}
	for _, limit := range []uint64{239, 48, 47} {
		if points, ok := exactPoints(target, limit); ok || points != nil {
			t.Fatalf("grid accepted insufficient byte budget %d: %v", limit, points)
		}
	}
	// Without a gravity lattice, only the retained Cartesian points are charged.
	target.info.CenterOfGravity = nil
	target.placements = []knapsack.Placement{{Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}}}
	points, ok = exactPoints(target, 192)
	if !ok || !slices.Equal(points, want) {
		t.Fatalf("non-lattice exact-budget grid = %v, ok=%v", points, ok)
	}
	if points, ok := exactPoints(target, 191); ok || points != nil {
		t.Fatalf("non-lattice grid accepted one-byte shortage: %v", points)
	}
}

func TestExactCenterGridsChargeDistinctTypesOnce(t *testing.T) {
	t.Parallel()
	first, repeated, second, ordinary := baseInternalBin(), baseInternalBin(), baseInternalBin(), baseInternalBin()
	for _, target := range []*bin{first, repeated, second} {
		target.info.Dimensions = geometry.Dimensions{X: 2, Y: 2, Z: 2}
		target.info.CenterOfGravity = &knapsack.CenterOfGravityBounds{}
	}
	second.info.ID = "second"
	ordinary.info.ID = "ordinary"
	// Each retained grid is 192 bytes; one transient coordinate slice is 48.
	grids, remaining, ok := exactCenterGrids([]*bin{first, repeated, second, ordinary}, 432)
	if !ok || remaining != 48 || len(grids) != 2 || len(grids[first.info.ID]) != 8 || len(grids[second.info.ID]) != 8 {
		t.Fatalf("grids=%v remaining=%d ok=%v", grids, remaining, ok)
	}
	if !slices.Equal(grids[first.info.ID], grids[second.info.ID]) {
		t.Fatal("identical dimensions produced different grids")
	}
	if grids, remaining, ok := exactCenterGrids([]*bin{first, repeated, second}, 431); ok || grids != nil || remaining != 0 {
		t.Fatalf("distinct grids accepted one-byte shortage: %v remaining=%d", grids, remaining)
	}
}
