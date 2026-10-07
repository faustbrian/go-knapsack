package solver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func TestPackFixedClassifiesIndependentInstanceFailures(t *testing.T) {
	t.Parallel()
	request := exactRequest(t, 4, 2)
	strategies := map[string]func(context.Context, knapsack.NormalizedRequest, []knapsack.ContainerInstance, solver.Options) (knapsack.Plan, error){
		"heuristic": (solver.Heuristic{}).PackFixed,
		"exact":     (solver.Exact{}).PackFixed,
	}
	cases := []struct {
		name      string
		instances []knapsack.ContainerInstance
		want      error
	}{
		{"unknown type", []knapsack.ContainerInstance{{ID: "box#1", TypeID: "unknown"}}, knapsack.ErrInvalidContainer},
		{"empty identity", []knapsack.ContainerInstance{{ID: "", TypeID: "box"}}, knapsack.ErrInvalidContainer},
		{"duplicate identity", []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}, {ID: "box#1", TypeID: "box"}}, knapsack.ErrDuplicateID},
		{"valid", []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, nil},
	}
	for name, pack := range strategies {
		t.Run(name, func(t *testing.T) {
			for _, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					plan, err := pack(context.Background(), request, test.instances, solver.Options{})
					if test.want != nil {
						if !errors.Is(err, test.want) {
							t.Fatalf("error = %v, want %v", err, test.want)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if len(plan.Placements()) != request.ItemCount() || len(plan.UnpackedItemIDs()) != 0 {
						t.Fatalf("valid instance did not produce a complete plan: %s", plan.CanonicalString())
					}
					if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
						t.Fatalf("invalid control plan: %+v", result.Violations())
					}
				})
			}
		})
	}
}
