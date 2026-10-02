package knapsack

import (
	"math"
	"math/big"
	"testing"
)

func TestMemoryEstimateContract(t *testing.T) {
	cases := []struct {
		name                     string
		used, limit, count, size uint64
	}{
		{"empty zero", 0, 0, 0, 0}, {"zero size one", 0, 0, 1, 0}, {"zero size max", 0, 0, math.MaxUint64, 0},
		{"partly used zero", 40, 100, math.MaxUint64, 0}, {"exhausted zero", 100, 100, math.MaxUint64, 0},
		{"zero count", 40, 100, 0, math.MaxUint64}, {"exact", 40, 100, 30, 2}, {"one over", 40, 99, 30, 2},
		{"exhausted positive", 100, 100, 1, 1}, {"multiplication overflow", 0, math.MaxUint64, math.MaxUint64, 2},
		{"max exact", 0, math.MaxUint64, math.MaxUint64, 1}, {"addition overflow", 1, math.MaxUint64, math.MaxUint64, 1}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := memoryEstimate{used: c.used, limit: c.limit}
			sum := new(big.Int).Mul(new(big.Int).SetUint64(c.count), new(big.Int).SetUint64(c.size))
			sum.Add(sum, new(big.Int).SetUint64(c.used))
			want := sum.Cmp(new(big.Int).SetUint64(c.limit)) <= 0
			got := e.add(c.count, c.size)
			if got != want {
				t.Fatalf("accepted %v want %v", got, want)
			}
			wantUsed := c.used
			if want {
				wantUsed = sum.Uint64()
			}
			if e.used != wantUsed {
				t.Fatalf("used %d want %d", e.used, wantUsed)
			}
		})
	}
}
