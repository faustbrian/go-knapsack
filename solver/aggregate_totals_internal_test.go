package solver

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/objective"
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

// These malformed states cannot pass request normalization. Exercise the
// builder's own refusal boundary, not a claim that public solvers create them.
func TestBuildPlanRefusesMalformedInternalAggregates(t *testing.T) {
	for _, field := range []string{"container_dimensions", "item_dimensions", "item_volume", "item_weight"} {
		t.Run(field, func(t *testing.T) {
			target := baseInternalBin()
			target.placements = []knapsack.Placement{
				{ItemID: "first", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1},
				{ItemID: "second", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1},
			}
			switch field {
			case "container_dimensions":
				target.info.Dimensions = geometry.Dimensions{}
			case "item_dimensions":
				target.placements[0].Dimensions = geometry.Dimensions{}
			case "item_volume":
				target.placements[0].Dimensions.X = math.MaxInt64
			case "item_weight":
				target.placements[0].Weight = math.MaxInt64
			}
			plan, err := buildPlan([]*bin{target}, nil, knapsack.StatusBestKnown, knapsack.TerminationCompleted, 0, 0, nil)
			if !errors.Is(err, knapsack.ErrOverflow) || plan.Status() != "" || len(plan.Placements()) != 0 {
				t.Fatalf("malformed %s published plan=%+v error=%v", field, plan.Spec(), err)
			}
		})
	}
}

func TestHeuristicTrialPropagatesPlanAdmissionFailure(t *testing.T) {
	request := internalRequest(t)
	goal, _ := objective.New(objective.Minimize(objective.ContainerCount))
	for _, existing := range []bool{true, false} {
		t.Run(map[bool]string{true: "existing", false: "new"}[existing], func(t *testing.T) {
			target := baseInternalBin()
			var bins []*bin
			var types []knapsack.NormalizedContainer
			if existing {
				target.instance.ID = strings.Repeat("x", int(knapsack.DefaultPlanLimits().MaxIDBytes)+1)
				bins = []*bin{target}
			} else {
				target.info.ID = strings.Repeat("x", int(knapsack.DefaultPlanLimits().MaxIDBytes))
				types = []knapsack.NormalizedContainer{target.info}
			}
			var candidates uint64
			selected, _, placed, err := chooseHeuristicPlacement(t.Context(), request, request.Items()[0], bins, types, map[string]uint32{}, !existing, &candidates, Options{}, goal)
			if !errors.Is(err, knapsack.ErrBudgetExhausted) || placed || selected != nil {
				t.Fatalf("plan admission failure lost: placed=%v selected=%v error=%v", placed, selected, err)
			}
		})
	}
}

func TestExactFixedExecutionRetainsCancellationWhenPartialTotalsOverflow(t *testing.T) {
	base := internalRequest(t)
	container := base.Containers()[0]
	container.Dimensions = geometry.Dimensions{X: math.MaxInt64, Y: 1, Z: 1}
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: base.Items(), Containers: []knapsack.NormalizedContainer{container}, Resolution: base.Resolution(), Limits: base.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	execution := exactFixedExecution{}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	plan, err := packExactFixed(ctx, request,
		[]knapsack.ContainerInstance{{ID: "box#1", TypeID: container.ID}, {ID: "box#2", TypeID: container.ID}},
		Options{Constraints: []constraint.Placement{internalCancel{cancel: cancel}}}, &execution)
	if !errors.Is(err, knapsack.ErrOverflow) || plan.Status() != "" || !execution.aggregateRefusal ||
		!errors.Is(execution.interruption, context.Canceled) || execution.termination != knapsack.TerminationCancelled {
		t.Fatalf("partial refusal erased cancellation: plan=%+v execution=%+v error=%v", plan.Spec(), execution, err)
	}
}
