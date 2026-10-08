package solver

import (
	"context"
	"slices"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestGravityOriginsRetainInclusiveInteriorIntervals(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name             string
		minimum, maximum uint32
		want             []int64
	}{
		{"interior interval", 400_000, 600_000, []int64{3, 4, 5}},
		{"single interior coordinate", 500_000, 500_000, []int64{4}},
		{"infeasible upper face", 1_000_000, 1_000_000, nil},
		{"rounded empty interval", 450_000, 450_000, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := gravityOrigins(10, 2, test.minimum, test.maximum); !slices.Equal(got, test.want) {
				t.Fatalf("gravity origins = %v, want %v", got, test.want)
			}
		})
	}
}

func TestPlacementAcceptsExactGrossCapacity(t *testing.T) {
	t.Parallel()
	for _, gross := range []int64{3, 4, 5} {
		for _, exact := range []bool{false, true} {
			target := baseInternalBin()
			target.info.HasGrossWeight, target.info.TareWeight, target.info.MaxGrossWeight = true, 1, gross
			base := baseInternalItem()
			base.ID = "base"
			target.items, target.weight = []knapsack.NormalizedItem{base}, 1
			target.placements = []knapsack.Placement{{ItemID: base.ID, ContainerID: target.instance.ID,
				Dimensions: base.Dimensions, Orientation: geometry.OrientationXYZ, Weight: 1}}
			target.points = []geometry.Point{{X: 1}}
			item := baseInternalItem()
			item.ID, item.Weight = "next", 2
			want := gross >= 4 // one tare + one existing + two new units
			if exact {
				placement, ok := exactPlacement(item, target, geometry.Point{X: 1}, geometry.OrientationXYZ)
				if ok != want || ok && (placement.ItemID != item.ID || placement.Weight != 2 || placement.Origin != (geometry.Point{X: 1})) {
					t.Fatalf("exact gross=%d accepted=%v placement=%+v", gross, ok, placement)
				}
			} else {
				var candidates uint64
				ok, err := tryPlace(context.Background(), item, target, &candidates, 10, nil)
				if err != nil || ok != want || ok && (len(target.placements) != 2 || target.weight != 3) {
					t.Fatalf("heuristic gross=%d accepted=%v weight=%d error=%v", gross, ok, target.weight, err)
				}
			}
		}
	}
}

func TestSignedHeightComparisonPreservesStrictThreeWayOrder(t *testing.T) {
	t.Parallel()
	ordered := []int64{-1 << 63, -1, 0, 1, 1<<63 - 1}
	for left, a := range ordered {
		for right, b := range ordered {
			want := 0
			if left < right {
				want = -1
			} else if left > right {
				want = 1
			}
			if got := compareInt64(a, b); got != want {
				t.Fatalf("height comparison (%d,%d)=%d, want %d", a, b, got, want)
			}
		}
	}
}
