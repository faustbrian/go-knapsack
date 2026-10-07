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
)

func TestComponentsHonorInclusiveCountAndAggregateByteBounds(t *testing.T) {
	valid := knapsack.ScoreComponent{Name: "metric", Direction: "min", Unit: "count", Value: "1"}
	for _, count := range []int{32, 33} {
		components := make([]knapsack.ScoreComponent, count)
		for index := range components {
			components[index] = valid
		}
		got, err := objective.SafeComponents(context.Background(), &objectiveCallback{components: components}, knapsack.NormalizedRequest{}, knapsack.Plan{})
		if count == 32 {
			if err != nil || !reflect.DeepEqual(got, components) {
				t.Fatal("inclusive component count rejected or altered")
			}
		} else if got != nil || !errors.Is(err, objective.ErrInvalidObjective) {
			t.Fatal("excess component count accepted")
		}
	}
	for _, nameBytes := range []int{511, 512} {
		component := knapsack.ScoreComponent{Name: strings.Repeat("n", nameBytes), Direction: "min", Unit: "u", Value: strings.Repeat("1", 512)}
		got, err := objective.SafeComponents(context.Background(), &objectiveCallback{components: []knapsack.ScoreComponent{component}}, knapsack.NormalizedRequest{}, knapsack.Plan{})
		if nameBytes == 511 {
			if err != nil || !reflect.DeepEqual(got, []knapsack.ScoreComponent{component}) {
				t.Fatal("inclusive aggregate component bytes rejected or altered")
			}
		} else if got != nil || !errors.Is(err, objective.ErrInvalidObjective) {
			t.Fatal("excess aggregate component bytes accepted")
		}
	}
}

func objectiveBoundRequest(t *testing.T, priorities []int64) knapsack.NormalizedRequest {
	t.Helper()
	ids := []string{"a", "b"}
	items := make([]knapsack.NormalizedItem, len(priorities))
	for index, priority := range priorities {
		items[index] = knapsack.NormalizedItem{ID: ids[index], Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}, Priority: priority}
	}
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: items, Containers: []knapsack.NormalizedContainer{{ID: "box", Dimensions: geometry.Dimensions{X: 2, Y: 1, Z: 1}, MaxContentWeight: math.MaxInt64, Stock: knapsack.UnlimitedStock()}}, Resolution: testResolution(), Limits: knapsack.DefaultLimits()})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestScoresPreserveInclusiveSignedBounds(t *testing.T) {
	for _, test := range []struct {
		name       string
		metric     objective.Metric
		priorities []int64
		weights    []int64
		z          int64
		want       int64
		invalid    bool
	}{
		{"height exact", objective.MaximumUsedHeight, []int64{0}, []int64{0}, math.MaxInt64 - 1, math.MaxInt64, false},
		{"height excess", objective.MaximumUsedHeight, []int64{0}, []int64{0}, math.MaxInt64, 0, true},
		{"priority max", objective.PackedPriority, []int64{math.MaxInt64 - 1, 1}, []int64{0, 0}, 0, math.MaxInt64, false},
		{"priority min", objective.PackedPriority, []int64{math.MinInt64 + 1, -1}, []int64{0, 0}, 0, math.MinInt64, false},
		{"priority negative", objective.PackedPriority, []int64{-1}, []int64{0}, 0, -1, false},
		{"priority overflow", objective.PackedPriority, []int64{math.MaxInt64, 1}, []int64{0, 0}, 0, 0, true},
		{"priority underflow", objective.PackedPriority, []int64{math.MinInt64, -1}, []int64{0, 0}, 0, 0, true},
		{"weight negative", objective.WeightImbalance, []int64{0}, []int64{-1}, 0, 0, false},
		{"weight max", objective.WeightImbalance, []int64{0}, []int64{math.MaxInt64}, 0, 0, false},
		{"weight min", objective.WeightImbalance, []int64{0}, []int64{math.MinInt64}, 0, 0, false},
		{"weight overflow", objective.WeightImbalance, []int64{0, 0}, []int64{math.MaxInt64, 1}, 0, 0, true},
		{"weight underflow", objective.WeightImbalance, []int64{0, 0}, []int64{math.MinInt64, -1}, 0, 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := objectiveBoundRequest(t, test.priorities)
			placements := make([]knapsack.Placement, len(test.weights))
			ids := []string{"a", "b"}
			for index, weight := range test.weights {
				placements[index] = knapsack.Placement{ItemID: ids[index], ContainerID: "left", Weight: weight, Origin: geometry.Point{Z: test.z}, Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}}
			}
			plan, err := knapsack.NewPlan(knapsack.PlanSpec{Containers: []knapsack.ContainerInstance{{ID: "left"}}, Placements: placements, Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted})
			if err != nil {
				t.Fatal(err)
			}
			goal, err := objective.New(objective.Minimize(test.metric))
			if err != nil {
				t.Fatal(err)
			}
			score, components, err := goal.ScorePlan(request, plan)
			if test.invalid {
				if !errors.Is(err, objective.ErrInvalidObjective) || len(score.Values) != 0 || components != nil {
					t.Fatal("unrepresentable score was accepted")
				}
			} else if err != nil || len(score.Values) != 1 || score.Values[0] != test.want || len(components) != 1 || components[0].Value != strconv.FormatInt(test.want, 10) {
				t.Fatal("inclusive signed score was rejected or altered")
			}
		})
	}
}

func TestWeightImbalanceIsExactAndContainerOrderIndependent(t *testing.T) {
	request := objectiveBoundRequest(t, []int64{0, 0})
	goal, err := objective.New(objective.Minimize(objective.WeightImbalance))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		left, right, want int64
		invalid           bool
	}{{1, 2, 1, false}, {-1, 0, 1, false}, {-1, math.MaxInt64 - 1, math.MaxInt64, false}, {-1, math.MaxInt64, 0, true}} {
		for _, ids := range [][]string{{"left", "right"}, {"right", "left"}} {
			plan, err := knapsack.NewPlan(knapsack.PlanSpec{Containers: []knapsack.ContainerInstance{{ID: ids[0]}, {ID: ids[1]}}, Placements: []knapsack.Placement{{ItemID: "a", ContainerID: "left", Weight: test.left}, {ItemID: "b", ContainerID: "right", Weight: test.right}}, Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted})
			if err != nil {
				t.Fatal(err)
			}
			score, components, err := goal.ScorePlan(request, plan)
			if test.invalid {
				if !errors.Is(err, objective.ErrInvalidObjective) || len(score.Values) != 0 || components != nil {
					t.Fatal("imbalance overflow accepted")
				}
			} else if err != nil || len(score.Values) != 1 || score.Values[0] != test.want || len(components) != 1 || components[0].Value != strconv.FormatInt(test.want, 10) {
				t.Fatal("container order changed exact imbalance")
			}
		}
	}
}

func TestCompareValidatesEachOperandAndPreservesLaterPrecedence(t *testing.T) {
	for _, second := range []objective.Criterion{objective.Minimize(objective.MaximumUsedHeight), objective.Maximize(objective.MaximumUsedHeight)} {
		goal, err := objective.New(objective.Minimize(objective.ContainerCount), second)
		if err != nil {
			t.Fatal(err)
		}
		left, right := objective.Score{Values: []int64{1, 1}, TieBreak: "z"}, objective.Score{Values: []int64{1, 2}, TieBreak: "a"}
		want := -1
		if second.Direction == objective.Max {
			want = 1
		}
		if got, err := goal.Compare(left, right); err != nil || got != want {
			t.Fatal("tie-break overrode later criterion")
		}
		if got, err := goal.Compare(right, left); err != nil || got != -want {
			t.Fatal("reversed comparison lost criterion direction")
		}
		right.Values = []int64{1, 1}
		if got, err := goal.Compare(left, right); err != nil || got <= 0 {
			t.Fatal("equal criteria lost canonical tie-break")
		}
		malformed := objective.Score{Values: []int64{1}}
		if _, err := goal.Compare(malformed, right); !errors.Is(err, objective.ErrInvalidObjective) {
			t.Fatal("malformed left accepted")
		}
		if _, err := goal.Compare(left, malformed); !errors.Is(err, objective.ErrInvalidObjective) {
			t.Fatal("malformed right accepted")
		}
	}
	if _, err := (objective.Objective{}).Compare(objective.Score{}, objective.Score{}); !errors.Is(err, objective.ErrInvalidObjective) {
		t.Fatal("zero objective accepted")
	}
}
