package constraint

import (
	"math/big"
	"testing"
)

func TestViewEstimateMatchesBoundedExactArithmetic(t *testing.T) {
	const bound = uint64(16 << 20)
	cases := []struct {
		name              string
		used, count, size uint64
	}{
		{"zero-cost", 0, ^uint64(0), 0},
		{"zero-cost-at-limit", bound, ^uint64(0), 0},
		{"empty-at-limit", bound, 0, 256},
		{"one-byte-exact", bound - 1, 1, 1},
		{"one-byte-over", bound, 1, 1},
		{"slice-exact", 0, bound / 16, 16},
		{"slice-over", 0, bound/16 + 1, 16},
		{"map-exact", 0, bound / 64, 64},
		{"record-exact", 0, bound / 256, 256},
		{"count-overflow", 0, ^uint64(0), 256},
		{"size-overflow", 0, 2, ^uint64(0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			total := new(big.Int).Mul(new(big.Int).SetUint64(tc.count), new(big.Int).SetUint64(tc.size))
			total.Add(total, new(big.Int).SetUint64(tc.used))
			want := total.Cmp(new(big.Int).SetUint64(bound)) <= 0
			estimate := viewEstimate{used: tc.used}
			accepted := estimate.add(tc.count, tc.size)
			if accepted != want {
				t.Fatalf("accepted = %t, want %t for exact total %s", accepted, want, total.String())
			}
			if !accepted {
				if estimate.used != tc.used {
					t.Fatal("rejected charge changed consumed budget")
				}
				return
			}
			if estimate.used != total.Uint64() {
				t.Fatalf("used = %d, want %s", estimate.used, total.String())
			}
		})
	}
}
