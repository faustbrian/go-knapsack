package solver

import (
	"context"
	"slices"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestPlacementOwnersSumSupportAndShareLoadAtBoundaries(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		supports  int
		loadLimit int64
		fragile   bool
		want      bool
	}{
		{"full support exact shared load", 2, 2, false, true},
		{"shared load exceeds limit", 2, 1, false, false},
		{"support area deficit", 1, 4, false, false},
		{"fragile supporter", 2, 2, true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, strategy := range []string{"exact", "heuristic"} {
				t.Run(strategy, func(t *testing.T) {
					target := baseInternalBin()
					target.info.Dimensions = geometry.Dimensions{X: 2, Y: 1, Z: 2}
					for index := range test.supports {
						base := baseInternalItem()
						base.ID = []string{"left", "right"}[index]
						base.MaxSupportedWeight = &test.loadLimit
						base.FragileTop = test.fragile
						target.items = append(target.items, base)
						target.placements = append(target.placements, knapsack.Placement{
							ItemID: base.ID, ContainerID: target.instance.ID,
							Origin: geometry.Point{X: int64(index)}, Dimensions: base.Dimensions,
							Orientation: geometry.OrientationXYZ,
						})
						target.weight += base.Weight
					}
					item := baseInternalItem()
					item.ID, item.Weight, item.MinimumSupportPPM = "top", 4, 1_000_000
					item.Dimensions.X = 2
					point := geometry.Point{Z: 1}
					var placement knapsack.Placement
					var accepted bool
					if strategy == "exact" {
						placement, accepted = exactPlacement(item, target, point, geometry.OrientationXYZ)
					} else {
						target.points = []geometry.Point{point}
						var candidates uint64
						var err error
						accepted, err = tryPlace(context.Background(), item, target, &candidates, 10, nil)
						if err != nil {
							t.Fatal(err)
						}
						if accepted {
							placement = target.placements[len(target.placements)-1]
						}
					}
					if accepted != test.want {
						t.Fatalf("accepted=%v want=%v", accepted, test.want)
					}
					if accepted && (placement.Origin != point || !slices.Equal(placement.SupporterIDs, []string{"left", "right"})) {
						t.Fatalf("incorrect supported placement: %+v", placement)
					}
				})
			}
		})
	}
}

func TestStackBoundaryIncludesTransitiveLoadAndDepth(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		load  int64
		depth uint32
		want  bool
	}{
		{"exact transitive load and depth", 2, 2, true},
		{"transitive load excess", 1, 2, false},
		{"transitive depth excess", 2, 1, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := baseInternalBin()
			bottom, middle, top := baseInternalItem(), baseInternalItem(), baseInternalItem()
			bottom.ID, middle.ID, top.ID = "bottom", "middle", "top"
			bottom.MaxSupportedWeight, bottom.MaxStackCount = &test.load, test.depth
			middleLimit := int64(1)
			middle.MaxSupportedWeight = &middleLimit
			target.items = []knapsack.NormalizedItem{bottom, middle}
			target.placements = []knapsack.Placement{
				{ItemID: bottom.ID, Dimensions: bottom.Dimensions},
				{ItemID: middle.ID, Origin: geometry.Point{Z: 1}, Dimensions: middle.Dimensions},
			}
			box, err := geometry.NewCuboid(geometry.Point{Z: 2}, top.Dimensions)
			if err != nil {
				t.Fatal(err)
			}
			if got := physicalPlacementAllowed(top, target, box, []string{"middle"}); got != test.want {
				t.Fatalf("physical placement accepted=%v want=%v", got, test.want)
			}
		})
	}
}
