package solver

import (
	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"testing"
)

func TestPlacementComparatorsOrderUnequalExtentsAndTiedFaces(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name    string
		axis    byte
		compare func(knapsack.Placement, knapsack.Placement) int
	}{
		{"primary x", 'x', comparePlacement},
		{"width-first z", 'z', comparePlacementWidthFirst},
		{"width-first x", 'x', comparePlacementWidthFirst},
		{"width-first y", 'y', comparePlacementWidthFirst},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			// All preceding maximum faces tie at one. The selected axis's
			// left face ends at eight and right face at two, so left is worse.
			left := knapsack.Placement{Origin: geometry.Point{}, Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Orientation: geometry.OrientationXYZ}
			right := left
			switch scenario.axis {
			case 'x':
				left.Origin.X = 4
				left.Dimensions.X = 4
				right.Origin.X = 1
			case 'y':
				left.Origin.Y = 4
				left.Dimensions.Y = 4
				right.Origin.Y = 1
			case 'z':
				left.Origin.Z = 4
				left.Dimensions.Z = 4
				right.Origin.Z = 1
			}
			if got := scenario.compare(left, right); got <= 0 {
				t.Fatalf("larger maximum face compared as %d", got)
			}
			if got := scenario.compare(right, left); got >= 0 {
				t.Fatalf("smaller maximum face compared as %d", got)
			}
			if got := scenario.compare(left, left); got != 0 {
				t.Fatalf("identical placement compared as %d", got)
			}
		})
	}
	// These boxes share all maximum faces at four. The smaller origin,
	// not their different dimensions, is the canonical tie-break.
	left := knapsack.Placement{Origin: geometry.Point{X: 1, Y: 1, Z: 1}, Dimensions: geometry.Dimensions{X: 3, Y: 3, Z: 3}, Orientation: geometry.OrientationXYZ}
	right := knapsack.Placement{Origin: geometry.Point{X: 2, Y: 2, Z: 2}, Dimensions: geometry.Dimensions{X: 2, Y: 2, Z: 2}, Orientation: geometry.OrientationXYZ}
	for _, compare := range []func(knapsack.Placement, knapsack.Placement) int{comparePlacement, comparePlacementWidthFirst} {
		if compare(left, right) >= 0 || compare(right, left) <= 0 {
			t.Fatal("tied maximum faces lost origin ordering")
		}
		orientation := left
		orientation.Orientation = geometry.OrientationZYX
		if compare(left, orientation) >= 0 || compare(orientation, left) <= 0 {
			t.Fatal("tied geometry lost orientation ordering")
		}
	}
}

func TestPointComparatorPreservesCanonicalOrderAndEquality(t *testing.T) {
	t.Parallel()
	// Literal order is Z, then Y, then X. Each primary-axis change
	// dominates deliberately opposing lower-axis coordinates.
	ordered := []geometry.Point{
		{X: 9, Y: 9, Z: 0},
		{X: 9, Y: 0, Z: 1},
		{X: 0, Y: 1, Z: 1},
		{X: 1, Y: 1, Z: 1},
		{X: 0, Y: 0, Z: 2},
	}
	for i, left := range ordered {
		if got := comparePoints(left, left); got != 0 {
			t.Fatalf("point %d compared with itself as %d", i, got)
		}
		for j := i + 1; j < len(ordered); j++ {
			right := ordered[j]
			if got := comparePoints(left, right); got >= 0 {
				t.Fatalf("ordered points %d/%d compared as %d", i, j, got)
			}
			if got := comparePoints(right, left); got <= 0 {
				t.Fatalf("reverse points %d/%d compared as %d", j, i, got)
			}
		}
	}
}
