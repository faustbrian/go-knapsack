package solver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
	"github.com/faustbrian/go-measurement/v2"
)

func TestExactWorkCountsConfigurationsAndRetainedPlans(t *testing.T) {
	t.Parallel()
	dimensions := knapsack.PhysicalDimensions{
		X: exactQuantity(1, measurement.Metre),
		Y: exactQuantity(1, measurement.Metre),
		Z: exactQuantity(1, measurement.Metre),
	}
	item, err := knapsack.NewItem(knapsack.ItemSpec{
		ID: "item", Dimensions: dimensions,
		Weight:       exactQuantity(1, measurement.Kilogram),
		Orientations: []geometry.Orientation{geometry.OrientationXYZ},
	})
	if err != nil {
		t.Fatal(err)
	}
	var containers []knapsack.ContainerType
	for _, id := range []string{"a", "b"} {
		container, err := knapsack.NewContainerType(knapsack.ContainerTypeSpec{
			ID: id, InternalDimensions: dimensions,
			MaxContentWeight: exactQuantity(1, measurement.Kilogram),
			Stock:            knapsack.UnlimitedStock(),
		})
		if err != nil {
			t.Fatal(err)
		}
		containers = append(containers, container)
	}
	request, err := knapsack.NewRequest([]knapsack.Item{item}, containers,
		knapsack.Resolution{Length: exactQuantity(1, measurement.Metre),
			Mass: exactQuantity(1, measurement.Kilogram)}, knapsack.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}

	// A fixed unit cube has one candidate/edge and two visited states.
	// PackAll visits each of two one-bin configurations and counts each
	// configuration as an additional branch. A rejected next configuration
	// is also an attempted branch, even when no fixed search starts.
	for _, scenario := range []struct {
		name               string
		fixed              bool
		limit              func(*knapsack.Limits)
		nodes              uint64
		branches           uint64
		candidates         uint64
		termination        knapsack.TerminationReason
		pairedOrientations bool
	}{
		{"fixed", true, nil, 2, 1, 1, knapsack.TerminationCompleted, false},
		{"all", false, nil, 4, 4, 2, knapsack.TerminationCompleted, false},
		{"node equality admits both", false, func(l *knapsack.Limits) { l.MaxSearchNodes = 4 }, 4, 4, 2, knapsack.TerminationCompleted, false},
		{"node exhaustion before next search", false, func(l *knapsack.Limits) { l.MaxSearchNodes = 2 }, 2, 3, 1, knapsack.TerminationNodeLimit, false},
		{"node exhaustion inside next search", false, func(l *knapsack.Limits) { l.MaxSearchNodes = 3 }, 4, 4, 2, knapsack.TerminationNodeLimit, false},
		{"branch equality admits both", false, func(l *knapsack.Limits) { l.MaxBranches = 4 }, 4, 4, 2, knapsack.TerminationCompleted, false},
		{"branch exhaustion before next search", false, func(l *knapsack.Limits) { l.MaxBranches = 2 }, 2, 3, 1, knapsack.TerminationBranchLimit, false},
		{"candidate equality admits both", false, func(l *knapsack.Limits) { l.MaxCandidatePlacements = 2 }, 4, 4, 2, knapsack.TerminationCompleted, false},
		{"candidate exhaustion before next search", false, func(l *knapsack.Limits) { l.MaxCandidatePlacements = 1 }, 2, 3, 1, knapsack.TerminationCandidateLimit, false},
		{"branch exhaustion inside next search", false, func(l *knapsack.Limits) { l.MaxBranches = 5 }, 5, 6, 4, knapsack.TerminationBranchLimit, true},
		{"candidate exhaustion inside next search", false, func(l *knapsack.Limits) { l.MaxCandidatePlacements = 3 }, 5, 6, 4, knapsack.TerminationCandidateLimit, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			limits := request.Normalized().Limits()
			if scenario.limit != nil {
				scenario.limit(&limits)
			}
			normalized := request.Normalized().WithLimits(limits)
			if scenario.pairedOrientations {
				items := normalized.Items()
				items[0].Orientations = []geometry.Orientation{geometry.OrientationXYZ, geometry.OrientationXZY}
				var err error
				normalized, err = knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
					Items: items, Containers: normalized.Containers(),
					Resolution: normalized.Resolution(), Limits: limits,
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			options := solver.Options{Seed: 37}
			var plan knapsack.Plan
			var err error
			strategy := "exhaustive_variable_container"
			if scenario.fixed {
				strategy = "exhaustive_compact_fixed"
				plan, err = (solver.Exact{}).PackFixed(context.Background(), normalized,
					[]knapsack.ContainerInstance{{ID: "a#000001", TypeID: "a"}}, options)
			} else {
				plan, err = (solver.Exact{}).PackAll(context.Background(), normalized, options)
			}
			status := knapsack.StatusOptimal
			if scenario.termination != knapsack.TerminationCompleted {
				status = knapsack.StatusBudgetExhausted
				if !errors.Is(err, knapsack.ErrBudgetExhausted) {
					t.Fatalf("error = %v, want budget exhaustion", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			want := knapsack.Work{Solver: "exact", Strategy: strategy, Seed: 37,
				Nodes: scenario.nodes, Branches: scenario.branches,
				CandidatePlacements: scenario.candidates}
			if plan.Work() != want || plan.Status() != status || plan.Termination() != scenario.termination {
				t.Fatalf("work = %+v, want %+v; status/termination = %s/%s", plan.Work(), want, plan.Status(), plan.Termination())
			}
			placements := plan.Placements()
			containerID := "a#000001"
			if scenario.pairedOrientations {
				// The interrupted second search already found a complete
				// unit-cube packing before its next attempted orientation.
				containerID = "b#000001"
			}
			if len(placements) != 1 || placements[0].ItemID != "item" || placements[0].ContainerID != containerID ||
				placements[0].Origin != (geometry.Point{}) || placements[0].Dimensions != (geometry.Dimensions{X: 1, Y: 1, Z: 1}) ||
				len(plan.UnpackedItemIDs()) != 0 || len(plan.Containers()) != 1 {
				t.Fatalf("retained unit-cube plan = %s", plan.CanonicalString())
			}
			if result := verify.Plan(normalized, plan, verify.RequireAll()); !result.Valid() {
				t.Fatalf("invalid retained plan: %+v", result.Violations())
			}
		})
	}
}

// Cancellation in a later configuration must not replace a complete plan
// with the empty result of the interrupted fixed search. The second
// predicate observes cancellation before any subsequent candidate work.
func TestExactRetainsPlanWhenLaterConfigurationCallbackCancels(t *testing.T) {
	t.Parallel()
	base := exactRequest(t, 4, 1)
	items := base.Items()[:1]
	items[0].Dimensions = geometry.Dimensions{X: 1, Y: 1, Z: 1}
	items[0].Orientations = []geometry.Orientation{geometry.OrientationXYZ}
	containers := base.Containers()
	containers[0].ID = "a"
	containers[0].Dimensions = geometry.Dimensions{X: 1, Y: 1, Z: 1}
	second := containers[0]
	second.ID = "b"
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{
		Items: items, Containers: append(containers, second),
		Resolution: base.Resolution(), Limits: base.Limits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	callback := cancelLaterContainer{cancel: cancel}
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{Seed: 37,
		Constraints: []constraint.Placement{callback, callback}})
	if !errors.Is(err, context.Canceled) || plan.Status() != knapsack.StatusBudgetExhausted ||
		plan.Termination() != knapsack.TerminationCancelled {
		t.Fatalf("interrupted plan = %s, error = %v", plan.CanonicalString(), err)
	}
	placements := plan.Placements()
	if len(placements) != 1 || placements[0].ItemID != items[0].ID || placements[0].ContainerID != "a#000001" || len(plan.UnpackedItemIDs()) != 0 {
		t.Fatalf("earlier complete plan was lost: %s", plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.RequireAll().WithConstraints(callback)); !result.Valid() {
		t.Fatalf("invalid retained plan: %+v", result.Violations())
	}
}

type cancelLaterContainer struct{ cancel context.CancelFunc }

func (c cancelLaterContainer) Check(_ context.Context, view constraint.PlacementView) constraint.Decision {
	if view.Container().ID == "b" {
		c.cancel()
	}
	return constraint.Accept()
}
