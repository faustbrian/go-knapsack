package geometry_test

import (
	"fmt"
	"testing"

	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestAdjacencyRequiresPositiveFaceAreaOnEveryAxis(t *testing.T) {
	base, err := geometry.NewCuboid(geometry.Point{X: 2, Y: 2, Z: 2}, geometry.Dimensions{X: 2, Y: 2, Z: 2})
	if err != nil {
		t.Fatal(err)
	}
	for normal := range 3 {
		for _, plane := range []int64{0, 4, 5} {
			for _, first := range []int64{1, 4, 5} {
				for _, second := range []int64{1, 4, 5} {
					t.Run(fmt.Sprintf("axis%d/plane%d/tangents%d_%d", normal, plane, first, second), func(t *testing.T) {
						positions := [3]int64{}
						positions[normal] = plane
						positions[(normal+1)%3] = first
						positions[(normal+2)%3] = second
						other, err := geometry.NewCuboid(geometry.Point{X: positions[0], Y: positions[1], Z: positions[2]}, geometry.Dimensions{X: 2, Y: 2, Z: 2})
						if err != nil {
							t.Fatal(err)
						}
						want := plane != 5 && first == 1 && second == 1
						if base.Adjacent(other) != want || other.Adjacent(base) != want {
							t.Fatalf("face adjacency must be %v in both argument orders", want)
						}
						if normal == 2 && plane == 4 {
							area, ok := base.SupportArea(other)
							wantArea := int64(0)
							if want {
								wantArea = 1
							}
							if area != wantArea || ok != want {
								t.Fatalf("support area = %d,%v; want %d,%v", area, ok, wantArea, want)
							}
						}
					})
				}
			}
		}
	}
	for _, origin := range []geometry.Point{{X: 3, Y: 3, Z: 3}, {X: 4, Y: 4, Z: 4}} {
		other, err := geometry.NewCuboid(origin, geometry.Dimensions{X: 2, Y: 2, Z: 2})
		if err != nil {
			t.Fatal(err)
		}
		if base.Adjacent(other) || other.Adjacent(base) {
			t.Fatal("volume overlap or corner contact is not face adjacency")
		}
	}
}
