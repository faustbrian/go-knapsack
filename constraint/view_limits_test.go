package constraint_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	knapsack "github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
)

func TestPlacementViewPlacementCountLimit(t *testing.T) {
	for _, count := range []int{9999, 10000, 10001} {
		placements := make([]knapsack.Placement, count)
		view, err := constraint.NewPlacementView(knapsack.NormalizedItem{}, knapsack.NormalizedContainer{}, knapsack.Placement{}, placements)
		if count > 10000 {
			if !errors.Is(err, constraint.ErrViewLimit) {
				t.Fatalf("%d placements accepted: %v", count, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%d placements rejected: %v", count, err)
		}
		if got := len(view.Placements()); got != count {
			t.Fatalf("placement copy length = %d, want %d", got, count)
		}
	}
}

func TestPlacementViewCopyBudget(t *testing.T) {
	// The existing conservative estimate reserves 256 bytes for each of the
	// item, container, and candidate before charging variable-length data.
	const identityBytes = (16 << 20) - 3*256
	identity := strings.Repeat("x", identityBytes+1)
	cases := []struct {
		name     string
		bytes    int
		sku      string
		rejected bool
	}{
		{"below", identityBytes - 1, "", false},
		{"exact", identityBytes, "", false},
		{"above", identityBytes + 1, "", true},
		{"cumulative", identityBytes, "x", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := knapsack.NormalizedItem{ID: identity[:tc.bytes], SKU: tc.sku}
			view, err := constraint.NewPlacementView(item, knapsack.NormalizedContainer{}, knapsack.Placement{}, nil)
			if tc.rejected {
				if !errors.Is(err, constraint.ErrViewLimit) {
					t.Fatalf("over-budget view accepted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("within-budget view rejected: %v", err)
			}
			if got := view.Item(); got.ID != item.ID || got.SKU != item.SKU {
				t.Fatal("view changed accepted identity")
			}
		})
	}
}

func TestPlacementViewRejectsStructuralCopyOverflow(t *testing.T) {
	const remaining = (16 << 20) - 3*256
	cases := []struct {
		name  string
		size  int
		input func(int) (knapsack.NormalizedItem, knapsack.NormalizedContainer, knapsack.Placement)
	}{
		{"orientations", 16, func(n int) (knapsack.NormalizedItem, knapsack.NormalizedContainer, knapsack.Placement) {
			return knapsack.NormalizedItem{Orientations: make([]geometry.Orientation, n)}, knapsack.NormalizedContainer{}, knapsack.Placement{}
		}},
		{"incompatible-groups", 16, func(n int) (knapsack.NormalizedItem, knapsack.NormalizedContainer, knapsack.Placement) {
			return knapsack.NormalizedItem{IncompatibleGroups: make([]string, n)}, knapsack.NormalizedContainer{}, knapsack.Placement{}
		}},
		{"attributes", 64, func(n int) (knapsack.NormalizedItem, knapsack.NormalizedContainer, knapsack.Placement) {
			m := make(map[string]string, n)
			for i := 0; i < n; i++ {
				m[strconv.Itoa(i)] = ""
			}
			return knapsack.NormalizedItem{Attributes: m}, knapsack.NormalizedContainer{}, knapsack.Placement{}
		}},
		{"allowed-classes", 16, func(n int) (knapsack.NormalizedItem, knapsack.NormalizedContainer, knapsack.Placement) {
			return knapsack.NormalizedItem{}, knapsack.NormalizedContainer{AllowedClasses: make([]string, n)}, knapsack.Placement{}
		}},
		{"reserved", 64, func(n int) (knapsack.NormalizedItem, knapsack.NormalizedContainer, knapsack.Placement) {
			return knapsack.NormalizedItem{}, knapsack.NormalizedContainer{Reserved: make([]geometry.Cuboid, n)}, knapsack.Placement{}
		}},
		{"supporters", 16, func(n int) (knapsack.NormalizedItem, knapsack.NormalizedContainer, knapsack.Placement) {
			return knapsack.NormalizedItem{}, knapsack.NormalizedContainer{}, knapsack.Placement{SupporterIDs: make([]string, n)}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item, container, candidate := tc.input(1)
			if _, err := constraint.NewPlacementView(item, container, candidate, nil); err != nil {
				t.Fatalf("small collection rejected: %v", err)
			}
			item, container, candidate = tc.input(remaining/tc.size + 1)
			if _, err := constraint.NewPlacementView(item, container, candidate, nil); !errors.Is(err, constraint.ErrViewLimit) {
				t.Fatalf("structural copy overflow accepted: %v", err)
			}
		})
	}
}

func TestPlacementViewDiagnosticIdentityBudget(t *testing.T) {
	const limit = (16 << 20) - 3*256 - 64
	payload := strings.Repeat("d", limit+1)
	for _, prior := range []bool{false, true} {
		for _, field := range []string{"code", "item-id"} {
			t.Run(strconv.FormatBool(prior)+"/"+field, func(t *testing.T) {
				bytes := limit + 1
				if prior {
					bytes -= 256
				}
				diagnostic := knapsack.Diagnostic{}
				if field == "code" {
					diagnostic.Code = payload[:bytes]
				} else {
					diagnostic.ItemID = payload[:bytes]
				}
				candidate := knapsack.Placement{Diagnostics: []knapsack.Diagnostic{diagnostic}}
				var placements []knapsack.Placement
				if prior {
					placements = []knapsack.Placement{candidate}
					candidate = knapsack.Placement{}
				}
				if _, err := constraint.NewPlacementView(knapsack.NormalizedItem{}, knapsack.NormalizedContainer{}, candidate, placements); !errors.Is(err, constraint.ErrViewLimit) {
					t.Fatalf("oversize diagnostic identity accepted: %v", err)
				}
			})
		}
	}
	small := knapsack.Diagnostic{Code: "code", ItemID: "item", ContainerID: "container", Message: "message"}
	candidate := knapsack.Placement{Diagnostics: []knapsack.Diagnostic{small}}
	view, err := constraint.NewPlacementView(knapsack.NormalizedItem{}, knapsack.NormalizedContainer{}, candidate, []knapsack.Placement{candidate})
	if err != nil {
		t.Fatalf("small diagnostics rejected: %v", err)
	}
	if view.Candidate().Diagnostics[0] != small || view.Placements()[0].Diagnostics[0] != small {
		t.Fatal("accepted diagnostic fields changed")
	}
}
