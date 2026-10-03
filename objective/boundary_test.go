package objective_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/objective"
)

func TestSafeComponentsInclusiveLimits(t *testing.T) {
	for _, count := range []int{31, 32, 33} {
		t.Run(fmt.Sprintf("count-%d", count), func(t *testing.T) {
			entries := make([]knapsack.ScoreComponent, count)
			for i := range entries {
				entries[i] = knapsack.ScoreComponent{Name: "n", Direction: "min", Unit: "u", Value: "1"}
			}
			assertComponentBoundary(t, entries, count <= 32)
		})
	}
	for _, size := range []int{1023, 1024, 1025} {
		t.Run(fmt.Sprintf("bytes-%d", size), func(t *testing.T) {
			entries := []knapsack.ScoreComponent{{Name: "n", Direction: "max", Unit: "u", Value: strings.Repeat("1", size-2)}}
			assertComponentBoundary(t, entries, size <= 1024)
		})
	}
}

func assertComponentBoundary(t *testing.T, entries []knapsack.ScoreComponent, accepted bool) {
	t.Helper()
	callback := &objectiveCallback{components: entries}
	got, err := objective.SafeComponents(context.Background(), callback, knapsack.NormalizedRequest{}, knapsack.Plan{})
	if !accepted {
		if !errors.Is(err, objective.ErrInvalidObjective) || got != nil {
			t.Fatalf("over-limit payload: components=%v error=%v", got, err)
		}
		return
	}
	if err != nil || !reflect.DeepEqual(got, entries) {
		t.Fatalf("inclusive payload: components=%v error=%v", got, err)
	}
	original := entries[0].Value
	got[0].Value = "changed"
	if callback.components[0].Value != original {
		t.Fatal("result aliases callback components")
	}
}

func TestObjectivePrioritySignedBounds(t *testing.T) {
	tests := []struct {
		name       string
		priorities []int64
		want       int64
		invalid    bool
	}{
		{"positive equality", []int64{math.MaxInt64 - 1, 1}, math.MaxInt64, false},
		{"positive overflow", []int64{math.MaxInt64, 1}, 0, true},
		{"negative equality", []int64{math.MinInt64 + 1, -1}, math.MinInt64, false},
		{"negative overflow", []int64{math.MinInt64, -1}, 0, true},
		{"negative finite", []int64{-7, -3, 0}, -10, false},
		{"intermediate overflow", []int64{math.MaxInt64, 1, -1}, 0, true},
	}
	goal, err := objective.New(objective.Maximize(objective.PackedPriority))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := boundaryRequest(t, tc.priorities)
			placements := make([]knapsack.Placement, len(tc.priorities))
			for i := range placements {
				placements[i] = knapsack.Placement{ItemID: fmt.Sprintf("i%d", i), ContainerID: "box", Weight: 1}
			}
			plan := boundaryPlan(t, []knapsack.ContainerInstance{{ID: "box"}}, placements, knapsack.Statistics{})
			score, components, err := goal.ScorePlan(request, plan)
			assertScoreBoundary(t, score, components, err, tc.invalid, tc.want, objective.PackedPriority, "max", "priority", plan)
		})
	}
}

func TestObjectiveHeightWeightAndImbalanceBounds(t *testing.T) {
	request := boundaryRequest(t, []int64{0, 0})
	tests := []struct {
		name       string
		metric     objective.Metric
		containers []knapsack.ContainerInstance
		placements []knapsack.Placement
		want       int64
		invalid    bool
	}{
		{"height equality", objective.MaximumUsedHeight, nil, []knapsack.Placement{{ItemID: "i0", Origin: geometry.Point{Z: math.MaxInt64 - 1}, Dimensions: geometry.Dimensions{Z: 1}}}, math.MaxInt64, false},
		{"height overflow", objective.MaximumUsedHeight, nil, []knapsack.Placement{{ItemID: "i0", Origin: geometry.Point{Z: math.MaxInt64}, Dimensions: geometry.Dimensions{Z: 1}}}, 0, true},
		{"weight positive equality", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "box"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "box", Weight: math.MaxInt64 - 1}, {ItemID: "i1", ContainerID: "box", Weight: 1}}, 0, false},
		{"weight positive overflow", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "box"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "box", Weight: math.MaxInt64}, {ItemID: "i1", ContainerID: "box", Weight: 1}}, 0, true},
		{"weight negative equality", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "box"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "box", Weight: math.MinInt64 + 1}, {ItemID: "i1", ContainerID: "box", Weight: -1}}, 0, false},
		{"weight negative overflow", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "box"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "box", Weight: math.MinInt64}, {ItemID: "i1", ContainerID: "box", Weight: -1}}, 0, true},
		{"imbalance equality", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "a"}, {ID: "b"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "a", Weight: -1}, {ItemID: "i1", ContainerID: "b", Weight: math.MaxInt64 - 1}}, math.MaxInt64, false},
		{"imbalance overflow", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "a"}, {ID: "b"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "a", Weight: -1}, {ItemID: "i1", ContainerID: "b", Weight: math.MaxInt64}}, 0, true},
		{"zero minimum", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "a"}, {ID: "b"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "b", Weight: math.MaxInt64}}, math.MaxInt64, false},
		{"three selected weights", objective.WeightImbalance, []knapsack.ContainerInstance{{ID: "a"}, {ID: "b"}, {ID: "c"}}, []knapsack.Placement{{ItemID: "i0", ContainerID: "a", Weight: 3}, {ItemID: "i1", ContainerID: "b", Weight: 8}, {ItemID: "i0", ContainerID: "c", Weight: 5}}, 5, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			goal, err := objective.New(objective.Minimize(tc.metric))
			if err != nil {
				t.Fatal(err)
			}
			plan := boundaryPlan(t, tc.containers, tc.placements, knapsack.Statistics{})
			score, components, err := goal.ScorePlan(request, plan)
			unit := "mass_lattice"
			if tc.metric == objective.MaximumUsedHeight {
				unit = "length_lattice"
			}
			assertScoreBoundary(t, score, components, err, tc.invalid, tc.want, tc.metric, "min", unit, plan)
		})
	}
}

func assertScoreBoundary(t *testing.T, score objective.Score, components []knapsack.ScoreComponent, err error, invalid bool, want int64, metric objective.Metric, direction, unit string, plan knapsack.Plan) {
	t.Helper()
	if invalid {
		if !errors.Is(err, objective.ErrInvalidObjective) || score.Values != nil || score.TieBreak != "" || components != nil {
			t.Fatalf("invalid score returned partial result: score=%+v components=%v error=%v", score, components, err)
		}
		return
	}
	expected := []knapsack.ScoreComponent{{Name: string(metric), Direction: direction, Unit: unit, Value: strconv.FormatInt(want, 10)}}
	if err != nil || !reflect.DeepEqual(score.Values, []int64{want}) || score.TieBreak != plan.CanonicalString() || !reflect.DeepEqual(components, expected) {
		t.Fatalf("exact score: score=%+v components=%v error=%v; want %d/%v", score, components, err, want, expected)
	}
}

func TestObjectiveCompareShapesAndEqualPrefix(t *testing.T) {
	for _, direction := range []objective.Direction{objective.Min, objective.Max} {
		goal, err := objective.New(objective.Minimize(objective.ContainerCount), objective.Criterion{Metric: objective.PackedPriority, Direction: direction})
		if err != nil {
			t.Fatal(err)
		}
		for _, values := range []struct{ left, right []int64 }{{nil, []int64{1, 2}}, {[]int64{1, 2}, nil}, {[]int64{1}, []int64{1, 2}}, {[]int64{1, 2}, []int64{1}}} {
			got, err := goal.Compare(objective.Score{Values: values.left}, objective.Score{Values: values.right})
			if got != 0 || !errors.Is(err, objective.ErrInvalidObjective) {
				t.Fatalf("malformed shape: comparison=%d error=%v", got, err)
			}
		}
		left := objective.Score{Values: []int64{1, math.MinInt64}, TieBreak: "z"}
		right := objective.Score{Values: []int64{1, math.MaxInt64}, TieBreak: "a"}
		want := -1
		if direction == objective.Max {
			want = 1
		}
		got, err := goal.Compare(left, right)
		if err != nil || got != want {
			t.Fatalf("secondary criterion: got %d/%v want %d", got, err, want)
		}
		got, err = goal.Compare(right, left)
		if err != nil || got != -want {
			t.Fatalf("reverse secondary criterion: got %d/%v want %d", got, err, -want)
		}
		left.Values = []int64{1, 2}
		right.Values = []int64{1, 2}
		left.TieBreak = "a"
		right.TieBreak = "b"
		got, err = goal.Compare(left, right)
		if err != nil || got != -1 {
			t.Fatalf("lexical tie: %d/%v", got, err)
		}
		got, err = goal.Compare(right, left)
		if err != nil || got != 1 {
			t.Fatalf("reverse lexical tie: %d/%v", got, err)
		}
		got, err = goal.Compare(left, left)
		if err != nil || got != 0 {
			t.Fatalf("identity tie: %d/%v", got, err)
		}
	}
	got, err := (objective.Objective{}).Compare(objective.Score{}, objective.Score{})
	if got != 0 || !errors.Is(err, objective.ErrInvalidObjective) {
		t.Fatalf("zero objective: %d/%v", got, err)
	}
}

func TestObjectiveAllMetricEvidence(t *testing.T) {
	request := boundaryRequest(t, []int64{7, -2, 4})
	goal, err := objective.New(objective.Minimize(objective.ContainerCount), objective.Minimize(objective.UnusedVolume), objective.Minimize(objective.UnusedWeight), objective.Minimize(objective.WeightImbalance), objective.Minimize(objective.MaximumUsedHeight), objective.Maximize(objective.PackedPriority))
	if err != nil {
		t.Fatal(err)
	}
	plan := boundaryPlan(t, []knapsack.ContainerInstance{{ID: "a"}, {ID: "b"}, {ID: "c"}}, []knapsack.Placement{
		{ItemID: "i0", ContainerID: "a", Weight: 3, Origin: geometry.Point{Z: 1}, Dimensions: geometry.Dimensions{Z: 2}},
		{ItemID: "i1", ContainerID: "b", Weight: 8, Dimensions: geometry.Dimensions{Z: 1}},
		{ItemID: "i2", ContainerID: "c", Weight: 5, Origin: geometry.Point{Z: 4}, Dimensions: geometry.Dimensions{Z: 3}},
	}, knapsack.Statistics{ContainerCount: 3, RemainingVolume: 11, RemainingWeight: 13})
	score, components, err := goal.ScorePlan(request, plan)
	expected := []knapsack.ScoreComponent{
		{Name: "container_count", Direction: "min", Unit: "count", Value: "3"},
		{Name: "unused_volume", Direction: "min", Unit: "lattice^3", Value: "11"},
		{Name: "unused_weight", Direction: "min", Unit: "mass_lattice", Value: "13"},
		{Name: "weight_imbalance", Direction: "min", Unit: "mass_lattice", Value: "5"},
		{Name: "maximum_used_height", Direction: "min", Unit: "length_lattice", Value: "7"},
		{Name: "packed_priority", Direction: "max", Unit: "priority", Value: "9"},
	}
	if err != nil || !reflect.DeepEqual(score.Values, []int64{3, 11, 13, 5, 7, 9}) || score.TieBreak != plan.CanonicalString() || !reflect.DeepEqual(components, expected) {
		t.Fatalf("metric evidence: score=%+v components=%v error=%v", score, components, err)
	}
}

func boundaryRequest(t *testing.T, priorities []int64) knapsack.NormalizedRequest {
	t.Helper()
	items := make([]knapsack.NormalizedItem, len(priorities))
	for i, priority := range priorities {
		items[i] = knapsack.NormalizedItem{ID: fmt.Sprintf("i%d", i), Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}, Priority: priority}
	}
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: items, Containers: []knapsack.NormalizedContainer{{ID: "box", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, MaxContentWeight: math.MaxInt64, Stock: knapsack.UnlimitedStock()}}, Resolution: testResolution(), Limits: knapsack.DefaultLimits()})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func boundaryPlan(t *testing.T, containers []knapsack.ContainerInstance, placements []knapsack.Placement, statistics knapsack.Statistics) knapsack.Plan {
	t.Helper()
	// NewPlan's public resource boundary is deliberately distinct from physical verification.
	plan, err := knapsack.NewPlan(knapsack.PlanSpec{Containers: containers, Placements: placements, Statistics: statistics, Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
