package solver_test

import (
	"errors"
	"math"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func TestHeuristicCountsOrientationsAfterEndpointOverflow(t *testing.T) {
	for _, limit := range []uint64{100, 9} {
		t.Run(map[uint64]string{100: "complete", 9: "candidate_limit"}[limit], func(t *testing.T) {
			baseline := exactRequest(t, 2, 1)
			limits := baseline.Limits()
			limits.MaxImprovementRounds = 0
			limits.MaxCandidatePlacements = limit
			// Large coordinates are scalars: the request has only two items
			// and the heuristic visits only their finite extreme points.
			request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
				Items: []knapsack.NormalizedItem{
					{ID: "anchor", Dimensions: geometry.Dimensions{X: math.MaxInt64 - 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}},
					{ID: "unpacked", Dimensions: geometry.Dimensions{X: 2, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ, geometry.OrientationYXZ}},
				},
				Containers: []knapsack.NormalizedContainer{{
					ID: "box", Dimensions: geometry.Dimensions{X: math.MaxInt64, Y: 1, Z: 1}, MaxContentWeight: 2, Stock: knapsack.FiniteStock(1),
				}},
				Resolution: baseline.Resolution(), Limits: limits,
			})
			if err != nil {
				t.Fatal(err)
			}
			plan, err := (solver.Heuristic{}).PackFixed(t.Context(), request,
				[]knapsack.ContainerInstance{{ID: "fixed", TypeID: "box"}}, solver.Options{AllowUnpacked: true})
			if limit == 9 {
				if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationCandidateLimit {
					t.Fatalf("candidate boundary: error=%v status=%s termination=%s", err, plan.Status(), plan.Termination())
				}
			} else if err != nil || plan.Status() != knapsack.StatusBestKnown || plan.Termination() != knapsack.TerminationNoPlacement {
				t.Fatalf("completed search: error=%v status=%s termination=%s", err, plan.Status(), plan.Termination())
			}
			placements := plan.Placements()
			unpacked := plan.UnpackedItemIDs()
			if len(placements) != 1 || placements[0].ItemID != "anchor" || placements[0].ContainerID != "fixed" ||
				placements[0].Origin != (geometry.Point{}) || placements[0].Dimensions != (geometry.Dimensions{X: math.MaxInt64 - 1, Y: 1, Z: 1}) ||
				len(unpacked) != 1 || unpacked[0] != "unpacked" || plan.Work().CandidatePlacements != 9 {
				t.Fatalf("orientation search lost work or item identities: placements=%+v unpacked=%v work=%+v", placements, unpacked, plan.Work())
			}
			if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
				t.Fatalf("partial plan violates conservation or geometry: %+v", result.Violations())
			}
		})
	}
}
