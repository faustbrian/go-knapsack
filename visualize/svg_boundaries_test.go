package visualize_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/verify"
	"github.com/faustbrian/go-knapsack/v2/visualize"
	"github.com/faustbrian/go-math/decimal"
	"github.com/faustbrian/go-measurement/v2"
)

func TestSVGCoordinateAndSceneWidthBoundaries(t *testing.T) {
	for _, test := range []struct {
		name    string
		sizes   []geometry.Dimensions
		viewBox string
		refused bool
	}{
		{"inclusive width", []geometry.Dimensions{{X: 999_980, Y: 1, Z: 1}}, `viewBox="0 0 1000000 21"`, false},
		{"width one beyond", []geometry.Dimensions{{X: 999_981, Y: 1, Z: 1}}, "", true},
		{"inclusive height", []geometry.Dimensions{{X: 1, Y: 1_000_000, Z: 1}}, `viewBox="0 0 21 1000020"`, false},
		{"height one beyond", []geometry.Dimensions{{X: 1, Y: 1_000_001, Z: 1}}, "", true},
		{"inclusive aggregate", []geometry.Dimensions{{X: 499_980, Y: 1, Z: 1}, {X: 499_990, Y: 1, Z: 1}}, `viewBox="0 0 1000000 21"`, false},
		{"aggregate one beyond", []geometry.Dimensions{{X: 499_980, Y: 1, Z: 1}, {X: 499_991, Y: 1, Z: 1}}, "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, plan := emptySVGScene(t, test.sizes)
			output, err := visualize.SVG(request, plan, verify.AllowUnpacked())
			if test.refused {
				if !errors.Is(err, visualize.ErrRenderLimit) || output != "" {
					t.Fatalf("out-of-budget scene was not refused: error=%v", err)
				}
				return
			}
			if err != nil || !strings.Contains(output, test.viewBox) {
				t.Fatalf("inclusive scene geometry changed: error=%v output=%q", err, output)
			}
		})
	}
}

func TestSVGPreservesPlacementGroupsAndOffsets(t *testing.T) {
	base, empty := emptySVGScene(t, []geometry.Dimensions{{X: 1, Y: 2, Z: 1}, {X: 3, Y: 1, Z: 1}})
	first, second := base.Items()[0], base.Items()[0]
	first.ID, second.ID = "first", "second"
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: []knapsack.NormalizedItem{first, second}, Containers: base.Containers(), Resolution: base.Resolution(), Limits: base.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	instances := empty.Containers()
	// Deliberately reverse placement order: each group must skip unrelated
	// placements and still emit its own item exactly once.
	plan, err := knapsack.NewPlan(knapsack.PlanSpec{
		Containers: instances,
		Placements: []knapsack.Placement{
			{ItemID: "second", ContainerID: instances[1].ID, Orientation: geometry.OrientationXYZ, Dimensions: second.Dimensions, Weight: 1},
			{ItemID: "first", ContainerID: instances[0].ID, Orientation: geometry.OrientationXYZ, Dimensions: first.Dimensions, Weight: 1},
		},
		Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted,
		Statistics: knapsack.Statistics{PackedItems: 2, ContainerCount: 2, ItemWeight: 2, ItemVolume: 2, ContainerVolume: 5, RemainingVolume: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
		t.Fatalf("invalid grouped SVG fixture: %v", result.Violations())
	}
	output, err := visualize.SVG(request, plan, verify.RequireAll())
	if err != nil {
		t.Fatal(err)
	}
	groups := strings.Split(output, "</g>")
	if len(groups) != 3 || !strings.Contains(output, `viewBox="0 0 34 22"`) {
		t.Fatalf("unexpected scene groups or dimensions: %q", output)
	}
	for index, expected := range []struct{ transform, title, other string }{
		{`transform="translate(10 10)"`, "<title>first</title>", "<title>second</title>"},
		{`transform="translate(21 10)"`, "<title>second</title>", "<title>first</title>"},
	} {
		if !strings.Contains(groups[index], expected.transform) || strings.Count(groups[index], expected.title) != 1 || strings.Contains(groups[index], expected.other) || strings.Count(groups[index], "<rect ") != 2 {
			t.Fatalf("group %d lost its offset or item ownership: %q", index, groups[index])
		}
	}
}

// Each fixture is independently verified before it reaches the renderer.
func emptySVGScene(t *testing.T, sizes []geometry.Dimensions) (knapsack.NormalizedRequest, knapsack.Plan) {
	t.Helper()
	containers := make([]knapsack.NormalizedContainer, len(sizes))
	instances := make([]knapsack.ContainerInstance, len(sizes))
	var volume int64
	for index, size := range sizes {
		id := fmt.Sprintf("box-%d", index)
		containers[index] = knapsack.NormalizedContainer{ID: id, Dimensions: size, MaxContentWeight: 1, Stock: knapsack.FiniteStock(1)}
		instances[index] = knapsack.ContainerInstance{ID: id + "#1", TypeID: id}
		volume += size.X * size.Y * size.Z
	}
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items:      []knapsack.NormalizedItem{{ID: "item", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}}},
		Containers: containers,
		Resolution: knapsack.Resolution{Length: measurement.MustNew(decimal.New(1), measurement.Metre), Mass: measurement.MustNew(decimal.New(1), measurement.Kilogram)},
		Limits:     knapsack.DefaultLimits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := knapsack.NewPlan(knapsack.PlanSpec{
		Containers: instances, UnpackedItemIDs: []string{"item"},
		Status: knapsack.StatusBestKnown, Termination: knapsack.TerminationNoPlacement,
		Statistics: knapsack.Statistics{ContainerCount: uint32(len(instances)), ContainerVolume: volume, RemainingVolume: volume, RemainingWeight: int64(len(instances))},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
		t.Fatalf("invalid SVG fixture: %v", result.Violations())
	}
	return request, plan
}
