package solver

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-knapsack"
	"github.com/faustbrian/go-knapsack/constraint"
)

// Keep the constraint guard ahead of exhaustive oracle cases so fail-fast
// verification rejects a broken callback boundary before expensive searches.
func TestExactSearchHonorsConstraintRejection(t *testing.T) {
	request := internalRequest(t)
	instances := []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}

	plan, err := (Exact{}).PackFixed(
		context.Background(),
		request,
		instances,
		Options{Constraints: []constraint.Placement{internalReject{}}},
	)
	if !errors.Is(err, knapsack.ErrProvenInfeasible) || plan.Status() != knapsack.StatusInfeasible {
		t.Fatalf("rejected plan=%s error=%v", plan.CanonicalString(), err)
	}
}
