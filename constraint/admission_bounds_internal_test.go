package constraint

import (
	"context"
	"errors"
	"strings"
	"testing"

	knapsack "github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
)

type boundaryDecisionCallback struct{ decision Decision }

func (callback boundaryDecisionCallback) Check(context.Context, PlacementView) Decision {
	return callback.decision
}

func TestDecisionAdmissionPreservesIndependentInclusiveBounds(t *testing.T) {
	tests := []struct {
		name     string
		decision Decision
		valid    bool
	}{
		{"missing code", Reject("", "message"), false},
		{"missing message", Reject("code", ""), false},
		{"code below", Reject(strings.Repeat("c", 63), "message"), true},
		{"code exact", Reject(strings.Repeat("c", 64), "message"), true},
		{"code excess", Reject(strings.Repeat("c", 65), "message"), false},
		{"message below", Reject("code", strings.Repeat("m", 1023)), true},
		{"message exact", Reject("code", strings.Repeat("m", 1024)), true},
		{"message excess", Reject("code", strings.Repeat("m", 1025)), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, evaluate := range []func() (Decision, error){
				func() (Decision, error) { return ValidateDecision(test.decision) },
				func() (Decision, error) {
					return Evaluate(context.Background(), boundaryDecisionCallback{test.decision}, PlacementView{})
				},
			} {
				got, err := evaluate()
				if test.valid {
					if err != nil || got != test.decision {
						t.Fatalf("bounded decision changed: error=%v", err)
					}
				} else if !errors.Is(err, ErrInvalidDecision) || got != (Decision{}) {
					t.Fatalf("malformed decision retained: error=%v", err)
				}
			}
		})
	}
}

func TestPlacementViewPreservesInclusivePlacementCount(t *testing.T) {
	placements := make([]knapsack.Placement, 10_001)
	for _, count := range []int{9999, 10_000, 10_001} {
		view, err := NewPlacementView(knapsack.NormalizedItem{}, knapsack.NormalizedContainer{}, knapsack.Placement{}, placements[:count])
		if count <= 10_000 {
			if err != nil || len(view.Placements()) != count {
				t.Fatalf("count %d rejected or changed: %v", count, err)
			}
		} else if !errors.Is(err, ErrViewLimit) || len(view.Placements()) != 0 {
			t.Fatalf("excess placements retained: %v", err)
		}
	}
}

func TestViewEstimateChargesExactAndExcessCounts(t *testing.T) {
	exact := viewEstimate{used: maxViewBytes - 16}
	if !exact.add(1, 16) || exact.used != maxViewBytes {
		t.Fatal("exact count budget rejected or not charged")
	}
	if !exact.add(0, 16) || exact.used != maxViewBytes {
		t.Fatal("empty collection changed the full budget")
	}
	if exact.add(1, 16) || exact.used != maxViewBytes {
		t.Fatal("excess count retained or changed accounting")
	}
	excess := viewEstimate{used: maxViewBytes - 16}
	if excess.add(2, 16) || excess.used != maxViewBytes-16 {
		t.Fatal("excess count charged before admission")
	}
}

func TestPlacementViewChargesEveryOwnedCollection(t *testing.T) {
	padding := strings.Repeat("x", 16<<20)
	for _, test := range []struct {
		name       string
		charge     int
		item       knapsack.NormalizedItem
		container  knapsack.NormalizedContainer
		candidate  knapsack.Placement
		placements []knapsack.Placement
	}{
		{name: "prior placements", charge: 256, placements: []knapsack.Placement{{}}},
		{name: "orientations", charge: 16, item: knapsack.NormalizedItem{Orientations: []geometry.Orientation{geometry.OrientationXYZ}}},
		{name: "incompatible groups", charge: 16, item: knapsack.NormalizedItem{IncompatibleGroups: []string{""}}},
		{name: "attributes", charge: 64, item: knapsack.NormalizedItem{Attributes: map[string]string{"": ""}}},
		{name: "allowed classes", charge: 16, container: knapsack.NormalizedContainer{AllowedClasses: []string{""}}},
		{name: "reserved cuboids", charge: 64, container: knapsack.NormalizedContainer{Reserved: []geometry.Cuboid{{}}}},
		{name: "candidate diagnostics", charge: 64, candidate: knapsack.Placement{Diagnostics: []knapsack.Diagnostic{{}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, excess := range []int{0, 1} {
				candidate := test.candidate
				candidate.ItemID = padding[:(16<<20)-768-test.charge+excess]
				view, err := NewPlacementView(test.item, test.container, candidate, test.placements)
				if excess == 0 {
					if err != nil || view.Candidate().ItemID != candidate.ItemID {
						t.Fatalf("exact collection charge rejected: %v", err)
					}
				} else if !errors.Is(err, ErrViewLimit) || view.Candidate().ItemID != "" {
					t.Fatalf("excess collection charge omitted: %v", err)
				}
			}
		})
	}
	for _, diagnostic := range []knapsack.Diagnostic{{ItemID: padding}, {ContainerID: padding}} {
		candidate := knapsack.Placement{Diagnostics: []knapsack.Diagnostic{diagnostic}}
		if _, err := NewPlacementView(knapsack.NormalizedItem{}, knapsack.NormalizedContainer{}, candidate, nil); !errors.Is(err, ErrViewLimit) {
			t.Fatalf("diagnostic identity omitted from admission: %v", err)
		}
	}
}
