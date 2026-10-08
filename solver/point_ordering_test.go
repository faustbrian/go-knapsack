package solver

import (
	"testing"

	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestPointComparisonPreservesLexicographicOrder(t *testing.T) {
	t.Parallel()
	// These points are in strict Z/Y/X order; later axes take precedence
	// over an earlier axis even when the earlier coordinate is larger.
	ordered := []geometry.Point{{}, {X: 1}, {X: 100, Y: 1}, {Y: 2}, {Z: 1}}
	for left, a := range ordered {
		for right, b := range ordered {
			want := 0
			if left < right {
				want = -1
			} else if left > right {
				want = 1
			}
			got := comparePoints(a, b)
			if (got < 0) != (want < 0) || (got > 0) != (want > 0) {
				t.Fatalf("point ordering (%d,%d): compare(%+v,%+v)=%d; want sign %d", left, right, a, b, got, want)
			}
		}
	}
}
