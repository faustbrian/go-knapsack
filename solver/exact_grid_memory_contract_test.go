package solver_test

import (
	"errors"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func TestExactFixedPointGridMemoryAdmission(t *testing.T) {
	for _, centered := range []bool{true, false} {
		name, pointBudget := "reserved endpoint grid", uint64(192)
		if centered {
			name, pointBudget = "center lattice", 240
		}
		t.Run(name, func(t *testing.T) {
			base := oracleRequest(t,
				[]knapsack.NormalizedItem{oracleItem(t, "item", geometry.Dimensions{X: 1, Y: 1, Z: 1})},
				geometry.Dimensions{X: 2, Y: 2, Z: 2})
			spec := knapsack.NormalizedSpec{
				Items: base.Items(), Containers: base.Containers(),
				Resolution: base.Resolution(), Limits: base.Limits(),
			}
			if centered {
				spec.Containers[0].CenterOfGravity = &knapsack.CenterOfGravityBounds{
					MaxXPPM: 1_000_000, MaxYPPM: 1_000_000, MaxZPPM: 1_000_000,
				}
			} else {
				reserved, err := geometry.NewCuboid(geometry.Point{}, geometry.Dimensions{X: 1, Y: 1, Z: 1})
				if err != nil {
					t.Fatal(err)
				}
				spec.Containers[0].Reserved = []geometry.Cuboid{reserved}
			}
			request, err := knapsack.NewNormalizedRequest(spec)
			if err != nil {
				t.Fatal(err)
			}
			// One item and one instance reserve two request copies, 2048
			// ordinary bytes and 768 exact-search bytes before point storage.
			reservation := 2*request.MemoryBytes() + 2816
			for _, shortage := range []uint64{0, 1} {
				limits := request.Limits()
				limits.MaxMemoryBytes = reservation + pointBudget - shortage
				scoped := request.WithLimits(limits)
				plan, err := (solver.Exact{}).PackFixed(t.Context(), scoped,
					[]knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{})
				if shortage != 0 {
					if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) {
						t.Fatalf("one-byte point shortage: error=%v", err)
					}
					if centered {
						if plan.Status() != "" || len(plan.Placements()) != 0 || len(plan.Containers()) != 0 {
							t.Fatalf("center presearch refusal returned a plan: %s", plan.CanonicalString())
						}
					} else if plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationMemoryLimit ||
						len(plan.Placements()) != 0 || len(plan.UnpackedItemIDs()) != 1 || plan.UnpackedItemIDs()[0] != "item" {
						t.Fatalf("endpoint-grid refusal lost its partial result: %s", plan.CanonicalString())
					}
					continue
				}
				if err != nil || plan.Status() != knapsack.StatusOptimal || len(plan.Placements()) != 1 ||
					plan.Placements()[0].ItemID != "item" || len(plan.UnpackedItemIDs()) != 0 {
					t.Fatalf("exact point allowance: plan=%s error=%v", plan.CanonicalString(), err)
				}
				if result := verify.Plan(scoped, plan, verify.RequireAll()); !result.Valid() {
					t.Fatalf("exact point allowance returned invalid placement: %+v", result.Violations())
				}
			}
		})
	}
}
