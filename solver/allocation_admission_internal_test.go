package solver

import (
	"strconv"
	"testing"

	"github.com/faustbrian/go-knapsack/v2/geometry"
)

// Large cases exercise scalar admission only: they must never reach a slice
// allocation, even when testing an intentionally broken admission predicate.
func TestExactLatticeScalarAdmission(t *testing.T) {
	for _, test := range []struct {
		name       string
		dimensions geometry.Dimensions
		budget     uint64
		accepted   bool
	}{
		{"inclusive", geometry.Dimensions{X: 2, Y: 2, Z: 2}, 240, true},
		{"point shortage", geometry.Dimensions{X: 2, Y: 2, Z: 2}, 239, false},
		{"coordinate shortage", geometry.Dimensions{X: 2, Y: 2, Z: 2}, 47, false},
		{"point bytes overflow", geometry.Dimensions{X: 1 << 60, Y: 1, Z: 1}, ^uint64(0), false},
		{"combined bytes overflow", geometry.Dimensions{X: 1 << 59, Y: 1, Z: 1}, ^uint64(0), false},
		{"32-bit integer ceiling", geometry.Dimensions{X: 1<<31 - 1, Y: 1, Z: 1}, ^uint64(0), true},
		{"integer width", geometry.Dimensions{X: 1 << 31, Y: 1, Z: 1}, ^uint64(0), strconv.IntSize == 64},
	} {
		t.Run(test.name, func(t *testing.T) {
			lengths, ok := exactLatticeLengths(test.dimensions, test.budget)
			if ok != test.accepted {
				t.Fatalf("admission=%v, want %v", ok, test.accepted)
			}
			want := [3]int64{}
			if test.accepted {
				want = [3]int64{test.dimensions.X, test.dimensions.Y, test.dimensions.Z}
			}
			if lengths != want {
				t.Fatalf("admitted lengths=%v, want %v", lengths, want)
			}
		})
	}
}

func TestExactCartesianScalarAdmission(t *testing.T) {
	for _, test := range []struct {
		name     string
		x, y, z  uint64
		budget   uint64
		capacity int
		accepted bool
	}{
		{"inclusive", 2, 3, 4, 576, 24, true},
		{"one byte shortage", 2, 3, 4, 575, 0, false},
		{"product overflow", 1 << 32, 1 << 32, 1, ^uint64(0), 0, false},
		{"integer overflow", uint64(^uint(0) >> 1), 2, 1, ^uint64(0), 0, false},
		{"point bytes shortage", 1 << 60, 1, 1, ^uint64(0), 0, false},
		{"32-bit integer ceiling", 1<<31 - 1, 1, 1, ^uint64(0), 1<<31 - 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			capacity, ok := exactPointCapacity(test.x, test.y, test.z, test.budget)
			if ok != test.accepted || capacity != test.capacity {
				t.Fatalf("capacity=%d admitted=%v, want %d/%v", capacity, ok, test.capacity, test.accepted)
			}
		})
	}
}
