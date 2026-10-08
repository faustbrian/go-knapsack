package solver

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
)

func TestWorkingMemoryReservationAcceptsExactBudget(t *testing.T) {
	t.Parallel()
	request := internalRequest(t)
	for _, exact := range []bool{false, true} {
		for _, maximumBins := range []int{1, 3} {
			items := uint64(request.ItemCount())
			bins := uint64(maximumBins)
			reserved := 2*request.MemoryBytes() + 1024*items + 1024*bins
			if exact {
				reserved += 256*items*items + 512*bins*items
			}
			for _, delta := range []int64{-1, 0, 1} {
				limits := request.Limits()
				limits.MaxMemoryBytes = uint64(int64(reserved) + delta)
				available, ok := workingMemoryAvailable(request.WithLimits(limits), maximumBins, exact)
				wantOK := delta >= 0
				var wantAvailable uint64
				if wantOK {
					wantAvailable = uint64(delta)
				}
				if ok != wantOK || available != wantAvailable {
					t.Fatalf("exact=%v bins=%d delta=%d: available=%d accepted=%v; want %d/%v",
						exact, maximumBins, delta, available, ok, wantAvailable, wantOK)
				}
			}
		}
	}
}

func TestHeuristicPacksAtExactWorkingMemoryReservation(t *testing.T) {
	t.Parallel()
	request := internalRequest(t)
	if request.ItemCount() != 1 || request.ContainerTypeCount() != 1 {
		t.Fatal("fixture must have one item and one container type")
	}
	instances := []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}
	for _, shortage := range []uint64{0, 1} {
		limits := request.Limits()
		limits.MaxMemoryBytes = 2*request.MemoryBytes() + 2048 - shortage
		plan, err := (Heuristic{}).PackFixed(context.Background(), request.WithLimits(limits), instances, Options{})
		if shortage == 1 {
			if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) {
				t.Fatalf("one-byte shortage: error=%v", err)
			}
		} else if err != nil || len(plan.Placements()) != 1 || len(plan.UnpackedItemIDs()) != 0 {
			t.Fatalf("exact reservation: packed=%d unpacked=%d error=%v", len(plan.Placements()), len(plan.UnpackedItemIDs()), err)
		}
	}
}
