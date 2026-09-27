package knapsacktest_test

import (
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/knapsacktest"
)

func TestRequireCanonicalEqual(t *testing.T) {
	t.Parallel()
	plan, _ := knapsack.NewPlan(knapsack.PlanSpec{Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted})
	knapsacktest.RequireCanonicalEqual(t, plan, plan)
}
