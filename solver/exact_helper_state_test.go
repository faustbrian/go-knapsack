package solver

import (
	"context"
	"errors"
	"testing"
)

func TestCheckedProductPreservesExactUnsignedBoundaries(t *testing.T) {
	t.Parallel()
	maximum := ^uint64(0)
	for _, test := range []struct {
		name   string
		values []uint64
		want   uint64
		ok     bool
	}{
		{"empty identity", nil, 1, true},
		{"positive factors", []uint64{3, 5, 7}, 105, true},
		{"single maximum", []uint64{maximum}, maximum, true},
		{"maximum times identity", []uint64{maximum, 1}, maximum, true},
		{"largest even product", []uint64{maximum / 2, 2}, maximum - 1, true},
		{"zero before maximum", []uint64{0, maximum}, 0, true},
		{"zero after maximum", []uint64{maximum, 0}, 0, true},
		{"first overflowing power", []uint64{1 << 63, 2}, 0, false},
		{"maximum squared", []uint64{maximum, maximum}, 0, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ok := checkedProduct(test.values...)
			if got != test.want || ok != test.ok {
				t.Fatalf("product(%v)=(%d,%v), want (%d,%v)", test.values, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestExactSearchStopsForEachIndependentTerminalCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("terminal cause")
	for _, test := range []struct {
		name string
		set  func(*exactSearch)
	}{
		{"budget", func(s *exactSearch) { s.budgeted = true }},
		{"cancellation", func(s *exactSearch) { s.cancelled = cause }},
		{"callback", func(s *exactSearch) { s.callbackErr = cause }},
		{"objective", func(s *exactSearch) { s.objectiveErr = cause }},
		{"invariant", func(s *exactSearch) { s.invariantErr = cause }},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := &exactSearch{ctx: context.Background()}
			if s.stopped() {
				t.Fatal("ordinary search stopped without a terminal cause")
			}
			test.set(s)
			if !s.stopped() {
				t.Fatal("independent terminal cause did not stop search")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := &exactSearch{ctx: ctx}
	if s.stopped() || s.cancelled != nil {
		t.Fatal("live context stopped search")
	}
	cancel()
	if !s.stopped() || !errors.Is(s.cancelled, context.Canceled) {
		t.Fatalf("context cancellation not retained: %v", s.cancelled)
	}
	if !s.stopped() || !errors.Is(s.cancelled, context.Canceled) {
		t.Fatal("terminal cancellation was lost on re-entry")
	}
}
