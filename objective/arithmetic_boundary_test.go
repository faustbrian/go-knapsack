package objective

import (
	"math"
	"math/big"
	"testing"
)

func TestCheckedAddMapMatchesExactIntegerBounds(t *testing.T) {
	pairs := []struct {
		name           string
		initial, delta int64
	}{
		{"zero delta", -7, 0}, {"negative finite", -7, -3},
		{"positive equality", math.MaxInt64 - 1, 1}, {"positive overflow", math.MaxInt64, 1},
		{"negative equality", math.MinInt64 + 1, -1}, {"negative overflow", math.MinInt64, -1},
		{"mixed signs", math.MinInt64, math.MaxInt64}, {"zero result", 7, -7},
	}
	for _, pair := range pairs {
		t.Run(pair.name, func(t *testing.T) {
			values := map[string]int64{"target": pair.initial, "independent": 11}
			exact := new(big.Int).Add(big.NewInt(pair.initial), big.NewInt(pair.delta))
			accepted := checkedAddMap(values, "target", pair.delta)
			if accepted != exact.IsInt64() {
				t.Fatalf("accepted=%t for exact sum %s", accepted, exact)
			}
			want := pair.initial
			if accepted {
				want = exact.Int64()
			}
			if values["target"] != want || values["independent"] != 11 {
				t.Fatalf("map=%v; want target %d and independent 11", values, want)
			}
		})
	}
	values := map[string]int64{"independent": 11}
	if !checkedAddMap(values, "new", -3) || values["new"] != -3 || values["independent"] != 11 {
		t.Fatalf("new key: %v", values)
	}
}
