package geometry_test

import (
	"math"
	"testing"

	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestCuboidAdjacencyRequiresPositiveFaceArea(t *testing.T) {
	t.Parallel()
	dimensions := geometry.Dimensions{X: 2, Y: 2, Z: 2}
	first := boundaryCuboid(t, geometry.Point{}, dimensions)
	for _, test := range []struct {
		name       string
		origin     geometry.Point
		adjacent   bool
		intersects bool
		support    int64
	}{
		{"same region", geometry.Point{}, false, true, 0},
		{"xy face", geometry.Point{Z: 2}, true, false, 4},
		{"xz face", geometry.Point{Y: 2}, true, false, 0},
		{"yz face", geometry.Point{X: 2}, true, false, 0},
		{"xz edge", geometry.Point{X: 2, Z: 2}, false, false, 0},
		{"yz edge", geometry.Point{Y: 2, Z: 2}, false, false, 0},
		{"xy edge", geometry.Point{X: 2, Y: 2}, false, false, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			second := boundaryCuboid(t, test.origin, dimensions)
			if got := first.Adjacent(second); got != test.adjacent {
				t.Fatalf("forward adjacency = %v, want %v", got, test.adjacent)
			}
			if got := second.Adjacent(first); got != test.adjacent {
				t.Fatalf("reverse adjacency = %v, want %v", got, test.adjacent)
			}
			if got := first.Intersects(second); got != test.intersects {
				t.Fatalf("intersection = %v, want %v", got, test.intersects)
			}
			if area, ok := first.SupportArea(second); area != test.support || ok != (test.support > 0) {
				t.Fatalf("support = %d, %v; want %d, %v", area, ok, test.support, test.support > 0)
			}
			if area, ok := second.SupportArea(first); area != 0 || ok {
				t.Fatalf("reverse support = %d, %v; want 0, false", area, ok)
			}
		})
	}
}

func TestCuboidSupportAreaAtExtremeEndpoints(t *testing.T) {
	t.Parallel()
	first := boundaryCuboid(t, geometry.Point{X: math.MaxInt64 - 4, Y: math.MaxInt64 - 4}, geometry.Dimensions{X: 4, Y: 4, Z: 2})
	second := boundaryCuboid(t, geometry.Point{X: math.MaxInt64 - 2, Y: math.MaxInt64 - 2, Z: 2}, geometry.Dimensions{X: 2, Y: 2, Z: 1})
	if area, ok := first.SupportArea(second); area != 4 || !ok {
		t.Fatalf("extreme partial support = %d, %v; want 4, true", area, ok)
	}
	if !first.Adjacent(second) || !second.Adjacent(first) || first.Intersects(second) {
		t.Fatal("extreme face contact changed adjacency or volume intersection")
	}
	nearZero := boundaryCuboid(t, geometry.Point{}, geometry.Dimensions{X: 2, Y: 2, Z: 2})
	if area, ok := nearZero.SupportArea(second); area != 0 || ok || nearZero.Adjacent(second) {
		t.Fatalf("extreme gap support = %d, %v; want 0, false without adjacency", area, ok)
	}
}

func boundaryCuboid(t *testing.T, origin geometry.Point, dimensions geometry.Dimensions) geometry.Cuboid {
	t.Helper()
	cuboid, err := geometry.NewCuboid(origin, dimensions)
	if err != nil {
		t.Fatalf("NewCuboid(%+v, %+v) = %v", origin, dimensions, err)
	}
	return cuboid
}
