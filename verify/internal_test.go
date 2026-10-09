package verify

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
)

type cancelAfterInternalChecks struct{ remaining int }

func (*cancelAfterInternalChecks) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*cancelAfterInternalChecks) Done() <-chan struct{}       { return nil }
func (*cancelAfterInternalChecks) Value(any) any               { return nil }
func (c *cancelAfterInternalChecks) Err() error {
	c.remaining--
	if c.remaining <= 0 {
		return context.Canceled
	}
	return nil
}

func TestCheckedAccountingAndCycleDefenses(t *testing.T) {
	t.Parallel()
	positive := int64(math.MaxInt64)
	if addChecked(&positive, 1) {
		t.Fatal("positive overflow accepted")
	}
	negative := int64(math.MinInt64)
	if addChecked(&negative, -1) {
		t.Fatal("negative overflow accepted")
	}
	ordinary := int64(1)
	if !addChecked(&ordinary, 2) || ordinary != 3 {
		t.Fatalf("ordinary total = %d", ordinary)
	}
	if depth, err := stackDepth(context.Background(), "a", map[string][]string{"a": {"b"}, "b": {"a"}}, make(map[string]bool)); err != nil || depth != math.MaxUint32 {
		t.Fatalf("cycle depth = %d", depth)
	}
	if compareInt64(1, 2) >= 0 || compareInt64(2, 1) <= 0 || compareInt64(1, 1) != 0 {
		t.Fatal("integer comparison changed")
	}
}

func TestCheckedAccountingInclusiveLimitsAndRejectedState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		initial, value, want int64
		accepted             bool
	}{
		{"positive equality", math.MaxInt64 - 1, 1, math.MaxInt64, true},
		{"negative equality", math.MinInt64 + 1, -1, math.MinInt64, true},
		{"zero at positive limit", math.MaxInt64, 0, math.MaxInt64, true},
		{"zero at negative limit", math.MinInt64, 0, math.MinInt64, true},
		{"positive refusal preserves total", math.MaxInt64, 1, math.MaxInt64, false},
		{"negative refusal preserves total", math.MinInt64, -1, math.MinInt64, false},
		{"ordinary subtraction", 3, -2, 1, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			total := test.initial
			if accepted := addChecked(&total, test.value); accepted != test.accepted || total != test.want {
				t.Fatalf("addChecked(%d, %d) = (%d, %v), want (%d, %v)", test.initial, test.value, total, accepted, test.want, test.accepted)
			}
		})
	}
}

func TestStackDepthHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := stackDepth(ctx, "a", nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}

func TestSupportCancellationChecksNestedWork(t *testing.T) {
	t.Parallel()

	box, _ := geometry.NewCuboid(geometry.Point{Z: 1}, geometry.Dimensions{X: 1, Y: 1, Z: 1})
	all := []placed{{
		placement: knapsack.Placement{ItemID: "item"}, box: box,
		item: knapsack.NormalizedItem{ID: "item", MaxStackCount: 1},
	}}
	if err := verifySupport(&cancelAfterInternalChecks{remaining: 2}, &Result{maximum: 1}, all, "box"); !errors.Is(err, context.Canceled) {
		t.Fatalf("nested support error = %v", err)
	}
	floor, _ := geometry.NewCuboid(geometry.Point{}, geometry.Dimensions{X: 1, Y: 1, Z: 1})
	all[0].box = floor
	if err := verifySupport(&cancelAfterInternalChecks{remaining: 4}, &Result{maximum: 1}, all, "box"); !errors.Is(err, context.Canceled) {
		t.Fatalf("stack support error = %v", err)
	}
	if _, err := stackDepth(&cancelAfterInternalChecks{remaining: 2}, "a", map[string][]string{"a": {"b"}}, make(map[string]bool)); !errors.Is(err, context.Canceled) {
		t.Fatalf("nested stack error = %v", err)
	}
}

func TestSupportAreaAccumulationRejectsOverflow(t *testing.T) {
	t.Parallel()

	dimensions := geometry.Dimensions{X: math.MaxInt64, Y: 1, Z: 1}
	supporter, err := geometry.NewCuboid(geometry.Point{}, dimensions)
	if err != nil {
		t.Fatal(err)
	}
	supported, err := geometry.NewCuboid(geometry.Point{Z: 1}, dimensions)
	if err != nil {
		t.Fatal(err)
	}
	all := []placed{
		{box: supporter, item: knapsack.NormalizedItem{ID: "left"}},
		{box: supporter, item: knapsack.NormalizedItem{ID: "right"}},
		{box: supported, item: knapsack.NormalizedItem{ID: "top", MinimumSupportPPM: 1}},
	}
	result := Result{maximum: 10}
	if err := verifySupport(context.Background(), &result, all, "box"); err != nil {
		t.Fatal(err)
	}
	if !result.Has(CodeOverflow) {
		t.Fatalf("violations = %+v", result.Violations())
	}
}

func TestSupportUnionAreaMergesAndSeparatesIntervals(t *testing.T) {
	t.Parallel()

	rectangles := []supportRectangle{
		{minX: 0, maxX: 2, minY: 0, maxY: 1},
		{minX: 0, maxX: 2, minY: 2, maxY: 3},
		{minX: 1, maxX: 3, minY: 0, maxY: 2},
		{minX: 1, maxX: 3, minY: 1, maxY: 2},
	}
	if got := supportUnionArea(rectangles); got.Int64() != 7 {
		t.Fatalf("union area = %s, want 7", got)
	}
	if got := supportUnionArea(nil); got.Sign() != 0 {
		t.Fatalf("empty union area = %s", got)
	}
}

func TestSupportUnionAreaNestedAndOffsetIntervals(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		rectangles []supportRectangle
		want       int64
	}{
		{"nested", []supportRectangle{{0, 1, 0, 3}, {0, 1, 1, 2}}, 3},
		{"offset disjoint", []supportRectangle{{0, 1, 2, 3}, {0, 1, 5, 6}}, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := supportUnionArea(test.rectangles)
			if !got.IsInt64() || got.Int64() != test.want {
				t.Fatalf("union area = %s, want %d", got, test.want)
			}
		})
	}
}
