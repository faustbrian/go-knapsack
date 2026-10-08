package solver

import (
	"errors"
	"math"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
)

func TestPlanTotalAdmissionIsExactAndNonMutatingOnRefusal(t *testing.T) {
	for _, test := range []struct {
		name         string
		initial, add int64
		want         int64
		overflow     bool
	}{
		{"zero", 0, 0, 0, false},
		{"inclusive", math.MaxInt64 - 1, 1, math.MaxInt64, false},
		{"overflow", math.MaxInt64, 1, math.MaxInt64, true},
		{"negative total", -1, 1, -1, true},
		{"negative contribution", 1, -1, 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			total := test.initial
			err := addPlanTotal(&total, test.add)
			if errors.Is(err, knapsack.ErrOverflow) != test.overflow || total != test.want {
				t.Fatalf("total=%d error=%v, want %d overflow=%v", total, err, test.want, test.overflow)
			}
		})
	}
}
