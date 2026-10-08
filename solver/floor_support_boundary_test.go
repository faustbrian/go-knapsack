package solver

import (
	"context"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
)

func TestPublicSolversAcceptFloorWithoutItemSupporters(t *testing.T) {
	t.Parallel()
	baseline := internalRequest(t)
	for _, support := range []uint32{0, 500_000, 1_000_000} {
		items := baseline.Items()
		items[0].MinimumSupportPPM = support
		request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
			Items: items, Containers: baseline.Containers(),
			Resolution: baseline.Resolution(), Limits: baseline.Limits(),
		})
		if err != nil {
			t.Fatal(err)
		}
		instances := []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}
		for name, solve := range map[string]func() (knapsack.Plan, error){
			"heuristic fixed": func() (knapsack.Plan, error) {
				return (Heuristic{}).PackFixed(context.Background(), request, instances, Options{})
			},
			"heuristic all": func() (knapsack.Plan, error) {
				return (Heuristic{}).PackAll(context.Background(), request, Options{})
			},
			"exact fixed": func() (knapsack.Plan, error) {
				return (Exact{}).PackFixed(context.Background(), request, instances, Options{})
			},
			"exact all": func() (knapsack.Plan, error) {
				return (Exact{}).PackAll(context.Background(), request, Options{})
			},
		} {
			plan, err := solve()
			placements := plan.Placements()
			if err != nil || len(placements) != 1 || len(plan.UnpackedItemIDs()) != 0 {
				t.Fatalf("%s support=%d: packed=%d unpacked=%d error=%v", name, support,
					len(placements), len(plan.UnpackedItemIDs()), err)
			}
			if placements[0].ItemID != items[0].ID || placements[0].Origin.Z != 0 || len(placements[0].SupporterIDs) != 0 {
				t.Fatalf("%s support=%d: incorrect floor placement %+v", name, support, placements[0])
			}
		}
	}
}
