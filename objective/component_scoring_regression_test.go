package objective_test

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/objective"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func componentScoringFixture(t *testing.T, partial bool) (knapsack.NormalizedRequest, knapsack.Plan) {
	t.Helper()
	request := objectiveBoundRequest(t, []int64{7, 11})
	placements := []knapsack.Placement{
		{ItemID: "a", ContainerID: "box-1", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Orientation: geometry.OrientationXYZ, Weight: 1},
		{ItemID: "b", ContainerID: "box-1", Origin: geometry.Point{X: 1}, Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Orientation: geometry.OrientationXYZ, Weight: 1},
	}
	spec := knapsack.PlanSpec{
		Containers: []knapsack.ContainerInstance{{ID: "box-1", TypeID: "box"}}, Placements: placements,
		Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted,
		Statistics: knapsack.Statistics{PackedItems: 2, ContainerCount: 1, ItemWeight: 2, ItemVolume: 2, ContainerVolume: 2, RemainingWeight: math.MaxInt64 - 2},
	}
	if partial {
		spec.Placements = placements[:1]
		spec.UnpackedItemIDs = []string{"b"}
		spec.Status = knapsack.StatusBestKnown
		spec.Termination = knapsack.TerminationNoPlacement
		spec.Statistics.PackedItems = 1
		spec.Statistics.ItemWeight = 1
		spec.Statistics.ItemVolume = 1
		spec.Statistics.RemainingWeight = math.MaxInt64 - 1
		spec.Statistics.RemainingVolume = 1
	}
	plan, err := knapsack.NewPlan(spec)
	if err != nil {
		t.Fatal(err)
	}
	if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
		t.Fatalf("invalid public scoring fixture: %v", result.Violations())
	}
	return request, plan
}

func TestComponentScoringPreservesExactMetricsAndCanonicalComparisons(t *testing.T) {
	goal, err := objective.New(
		objective.Minimize(objective.ContainerCount), objective.Minimize(objective.UnusedVolume),
		objective.Minimize(objective.UnusedWeight), objective.Minimize(objective.WeightImbalance),
		objective.Minimize(objective.MaximumUsedHeight), objective.Maximize(objective.PackedPriority),
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, partial := range []bool{false, true} {
		request, plan := componentScoringFixture(t, partial)
		wantValues := []int64{1, 0, math.MaxInt64 - 2, 0, 1, 18}
		if partial {
			wantValues = []int64{1, 1, math.MaxInt64 - 1, 0, 1, 7}
		}
		names := []string{"container_count", "unused_volume", "unused_weight", "weight_imbalance", "maximum_used_height", "packed_priority"}
		units := []string{"count", "lattice^3", "mass_lattice", "mass_lattice", "length_lattice", "priority"}
		want := make([]knapsack.ScoreComponent, len(names))
		for i := range names {
			direction := "min"
			if i == len(names)-1 {
				direction = "max"
			}
			want[i] = knapsack.ScoreComponent{Name: names[i], Direction: direction, Unit: units[i], Value: strconv.FormatInt(wantValues[i], 10)}
		}
		components, err := goal.Components(t.Context(), request, plan)
		if err != nil || !reflect.DeepEqual(components, want) {
			t.Fatalf("component metrics = %v, %v; want %v", components, err, want)
		}
		score, scoredComponents, err := goal.ScorePlan(request, plan)
		if err != nil || !reflect.DeepEqual(score.Values, wantValues) || score.TieBreak != plan.CanonicalString() || !reflect.DeepEqual(scoredComponents, want) {
			t.Fatal("public score or canonical tie changed")
		}
		otherSpec := plan.Spec()
		otherSpec.Work.Seed = 1
		other, err := knapsack.NewPlan(otherSpec)
		if err != nil {
			t.Fatal(err)
		}
		otherScore, _, err := goal.ScorePlan(request, other)
		if err != nil {
			t.Fatal(err)
		}
		wantComparison := strings.Compare(plan.CanonicalString(), other.CanonicalString())
		if wantComparison == 0 {
			t.Fatal("fixture does not distinguish canonical metadata")
		}
		for _, compare := range []func() (int, error){
			func() (int, error) { return goal.Compare(score, otherScore) },
			func() (int, error) { return goal.ComparePlans(t.Context(), request, plan, other) },
			func() (int, error) { return objective.SafeCompare(t.Context(), goal, request, plan, other) },
		} {
			if got, err := compare(); err != nil || got != wantComparison {
				t.Fatalf("canonical comparison = %d, %v; want %d", got, err, wantComparison)
			}
		}
	}
}

func TestComponentScoringPreservesContextAndAdmissionPrecedence(t *testing.T) {
	request, plan := componentScoringFixture(t, false)
	goal, _ := objective.New(objective.Minimize(objective.ContainerCount))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, test := range []struct {
		ctx  context.Context
		goal objective.Objective
		want error
	}{
		{nil, objective.Objective{}, knapsack.ErrInvalidOptions},
		{ctx, objective.Objective{}, context.Canceled},
		{t.Context(), objective.Objective{}, objective.ErrInvalidObjective},
		{ctx, goal, context.Canceled},
	} {
		if got, err := test.goal.Components(test.ctx, request, plan); got != nil || !errors.Is(err, test.want) {
			t.Fatalf("components admission = %v, %v; want %v", got, err, test.want)
		}
	}
	spec := plan.Spec()
	spec.Placements[0].ItemID = "unknown"
	unknown, err := knapsack.NewPlan(spec)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := goal.Components(t.Context(), request, unknown); got != nil || !errors.Is(err, objective.ErrInvalidObjective) {
		t.Fatal("components lost unknown-item admission")
	}
	overflowRequest := objectiveBoundRequest(t, []int64{math.MaxInt64, 1})
	if got, err := goal.Components(t.Context(), overflowRequest, plan); got != nil || !errors.Is(err, objective.ErrInvalidObjective) {
		t.Fatal("components bypassed checked priority overflow outside selected metric")
	}
}

func TestComponentScoringDoesNotAllocateDiscardedCanonicalTie(t *testing.T) {
	request, plan := componentScoringFixture(t, false)
	goal, _ := objective.New(objective.Minimize(objective.ContainerCount))
	var score objective.Score
	var components []knapsack.ScoreComponent
	full := testing.AllocsPerRun(20, func() {
		var err error
		score, components, err = goal.ScorePlan(request, plan)
		if err != nil {
			t.Fatal(err)
		}
	})
	componentOnly := testing.AllocsPerRun(20, func() {
		var err error
		components, err = goal.Components(t.Context(), request, plan)
		if err != nil {
			t.Fatal(err)
		}
	})
	if score.TieBreak != plan.CanonicalString() || len(components) != 1 {
		t.Fatal("allocation workload lost observable outputs")
	}
	if componentOnly >= full {
		t.Fatalf("component-only scoring allocates discarded canonical tie: %.0f allocations, full score %.0f", componentOnly, full)
	}
}
