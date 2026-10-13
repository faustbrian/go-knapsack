package solver_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/constraint"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/objective"
	"github.com/faustbrian/go-knapsack/v2/solver"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

type requireOrigin struct{ x int64 }

func (predicate requireOrigin) Check(_ context.Context, view constraint.PlacementView) constraint.Decision {
	if view.Candidate().Origin.X == predicate.x {
		return constraint.Accept()
	}
	return constraint.Reject("origin", "origin is not accepted")
}

type requireInstance struct{ id string }

func (predicate requireInstance) Check(_ context.Context, view constraint.PlacementView) constraint.Decision {
	if view.Candidate().ContainerID == predicate.id {
		return constraint.Accept()
	}
	return constraint.Reject("instance", "instance is not accepted")
}

type preferOrigin struct{ x int64 }

func (preferOrigin) Valid() bool { return true }
func (goal preferOrigin) score(plan knapsack.Plan) int64 {
	placements := plan.Placements()
	if len(placements) == 0 {
		return 1 << 30
	}
	distance := placements[0].Origin.X - goal.x
	if distance < 0 {
		distance = -distance
	}
	return distance
}
func (goal preferOrigin) ComparePlans(_ context.Context, _ knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	a, b := goal.score(left), goal.score(right)
	if a < b {
		return -1, nil
	}
	if a > b {
		return 1, nil
	}
	return 0, nil
}
func (goal preferOrigin) Components(_ context.Context, _ knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	return []knapsack.ScoreComponent{{Name: "origin_distance", Direction: "min", Unit: "lattice", Value: strconv.FormatInt(goal.score(plan), 10)}}, nil
}

type preferCount struct{ count int }

func (preferCount) Valid() bool { return true }
func (goal preferCount) score(plan knapsack.Plan) int {
	distance := len(plan.Containers()) - goal.count
	if distance < 0 {
		distance = -distance
	}
	return distance
}
func (goal preferCount) ComparePlans(_ context.Context, _ knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	a, b := goal.score(left), goal.score(right)
	if a < b {
		return -1, nil
	}
	if a > b {
		return 1, nil
	}
	return 0, nil
}
func (goal preferCount) Components(_ context.Context, _ knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	return []knapsack.ScoreComponent{{Name: "count_distance", Direction: "min", Unit: "count", Value: strconv.Itoa(goal.score(plan))}}, nil
}

type preferContainerType struct{}

func (preferContainerType) Valid() bool { return true }
func (preferContainerType) score(plan knapsack.Plan) int {
	for _, container := range plan.Containers() {
		if container.TypeID == "a-good" {
			return 1
		}
	}
	return 2
}
func (goal preferContainerType) ComparePlans(_ context.Context, _ knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	a, b := goal.score(left), goal.score(right)
	if a < b {
		return -1, nil
	}
	if a > b {
		return 1, nil
	}
	return 0, nil
}
func (goal preferContainerType) Components(_ context.Context, _ knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	return []knapsack.ScoreComponent{{Name: "container_cost", Direction: "min", Unit: "integer", Value: strconv.Itoa(goal.score(plan))}}, nil
}

func completenessRequest(t *testing.T, width int64, stock knapsack.Stock) knapsack.NormalizedRequest {
	t.Helper()
	base := oracleRequest(t, []knapsack.NormalizedItem{oracleItem(t, "item", geometry.Dimensions{X: 1, Y: 1, Z: 1})}, geometry.Dimensions{X: width, Y: 1, Z: 1})
	spec := knapsack.NormalizedSpec{Items: base.Items(), Containers: base.Containers(), Resolution: base.Resolution(), Limits: base.Limits()}
	spec.Containers[0].Stock = stock
	request, err := knapsack.NewNormalizedRequest(spec)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

// Independent oracle: walk the finite lattice directly, without the solver's
// coordinate helpers or feasibility functions. These one-cube floor cases have
// neither overlaps nor support-load ambiguity.
func TestExactOriginPredicatesAgreeWithIndependentLatticeOracle(t *testing.T) {
	for accepted := range int64(4) {
		request := completenessRequest(t, 3, knapsack.FiniteStock(1))
		wantFeasible := false
		for x := int64(0); x+1 <= 3; x++ {
			wantFeasible = wantFeasible || x == accepted
		}
		predicate := requireOrigin{x: accepted}
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		plan, err := (solver.Exact{}).PackFixed(ctx, request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{Constraints: []constraint.Placement{predicate}})
		cancel()
		if (err == nil) != wantFeasible {
			t.Fatalf("accepted x=%d feasible=%t want=%t error=%v status=%s", accepted, err == nil, wantFeasible, err, plan.Status())
		}
		if wantFeasible {
			if plan.Status() != knapsack.StatusOptimal || plan.Placements()[0].Origin.X != accepted {
				t.Fatal("feasible lattice placement lost")
			}
			if result := verify.Plan(request, plan, verify.RequireAll().WithConstraints(predicate)); !result.Valid() {
				t.Fatalf("independent verifier: %v", result.Violations())
			}
		} else if !errors.Is(err, knapsack.ErrProvenInfeasible) || plan.Status() != knapsack.StatusInfeasible {
			t.Fatal("completed finite lattice should prove infeasibility")
		}
	}
}

func TestExactOriginObjectiveAgreesWithIndependentLatticeMinimum(t *testing.T) {
	request := completenessRequest(t, 3, knapsack.FiniteStock(1))
	for desired := range int64(3) {
		wantOrigin, wantScore := int64(-1), int64(1<<30)
		for x := int64(0); x+1 <= 3; x++ {
			distance := x - desired
			if distance < 0 {
				distance = -distance
			}
			if distance < wantScore {
				wantOrigin, wantScore = x, distance
			}
		}
		goal := preferOrigin{x: desired}
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		plan, err := (solver.Exact{}).PackFixed(ctx, request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{PlanObjective: goal})
		cancel()
		if err != nil || plan.Status() != knapsack.StatusOptimal || plan.Placements()[0].Origin.X != wantOrigin {
			t.Fatalf("desired=%d want origin=%d error=%v plan=%s", desired, wantOrigin, err, plan.CanonicalString())
		}
		if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !result.Valid() {
			t.Fatalf("independent verifier: %v", result.Violations())
		}
	}
}

func TestExactPackAllEnumeratesRequiredEmptyInstance(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.FiniteStock(2))
	predicate := requireInstance{id: "box#000002"}
	// Exhausting the two legal instance ordinals independently admits only #2.
	wantBins := 0
	for bins := 1; bins <= 2 && wantBins == 0; bins++ {
		for instance := 1; instance <= bins; instance++ {
			if instance == 2 {
				wantBins = bins
			}
		}
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{Constraints: []constraint.Placement{predicate}})
	if err != nil || plan.Status() != knapsack.StatusOptimal || len(plan.Containers()) != wantBins {
		t.Fatalf("want bins=%d error=%v plan=%s", wantBins, err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.RequireAll().WithConstraints(predicate)); !result.Valid() {
		t.Fatalf("independent verifier: %v", result.Violations())
	}
}

func TestExactPackAllObjectivesIncludeUnusedFiniteStock(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.FiniteStock(4))
	for desired := 2; desired <= 4; desired++ {
		// Exhaustive count oracle, independent of multiset/search implementation.
		wantCount, wantScore := 0, 1<<30
		for bins := 1; bins <= 4; bins++ {
			distance := bins - desired
			if distance < 0 {
				distance = -distance
			}
			if distance < wantScore {
				wantCount, wantScore = bins, distance
			}
		}
		goal := preferCount{count: desired}
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: goal})
		cancel()
		if err != nil || plan.Status() != knapsack.StatusOptimal || len(plan.Containers()) != wantCount {
			t.Fatalf("desired=%d want count=%d error=%v plan=%s", desired, wantCount, err, plan.CanonicalString())
		}
		if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !result.Valid() {
			t.Fatalf("independent verifier: %v", result.Violations())
		}
	}
}

func TestExactUnlimitedObjectiveDoesNotPublishFalseOptimality(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.UnlimitedStock())
	limits := request.Limits()
	limits.MaxBranches = 32
	limits.MaxSearchNodes = 32
	limits.MaxCandidatePlacements = 32
	request = request.WithLimits(limits)
	builtin, err := objective.New(objective.Maximize(objective.ContainerCount))
	if err != nil {
		t.Fatal(err)
	}
	for _, goal := range []objective.PlanObjective{preferMoreContainers{}, builtin} {
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{PlanObjective: goal})
		cancel()
		if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || len(plan.Placements()) != 1 {
			t.Fatalf("unbounded improving objective cannot prove optimum: error=%v plan=%s", err, plan.CanonicalString())
		}
		if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !result.Valid() {
			t.Fatalf("retained feasible result invalid: %v", result.Violations())
		}
	}
}

func TestExactOriginSensitiveMemoryAdmissionPrecedesAllocation(t *testing.T) {
	base := oracleRequest(t, []knapsack.NormalizedItem{oracleItem(t, "item", geometry.Dimensions{X: 1, Y: 1, Z: 1})}, geometry.Dimensions{X: 2, Y: 2, Z: 2})
	reservation := 2*base.MemoryBytes() + 2816
	for _, shortage := range []uint64{0, 1} {
		limits := base.Limits()
		limits.MaxMemoryBytes = reservation + 240 - shortage
		request := base.WithLimits(limits)
		plan, err := (solver.Exact{}).PackFixed(t.Context(), request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{Constraints: []constraint.Placement{requireOrigin{x: 1}}})
		if shortage == 0 {
			if err != nil || plan.Status() != knapsack.StatusOptimal {
				t.Fatalf("inclusive full-grid memory: error=%v plan=%s", err, plan.CanonicalString())
			}
		} else if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || plan.Status() == knapsack.StatusOptimal || plan.Status() == knapsack.StatusInfeasible {
			t.Fatalf("grid memory shortage falsely certified result: error=%v plan=%s", err, plan.CanonicalString())
		}
	}
}

func TestExactPhysicalOriginsAgreeWithIndependentLatticeOracle(t *testing.T) {
	items := []knapsack.NormalizedItem{
		oracleItem(t, "a", geometry.Dimensions{X: 1, Y: 1, Z: 1}),
		oracleItem(t, "b", geometry.Dimensions{X: 1, Y: 1, Z: 1}),
	}
	for index := range items {
		items[index].FragileTop = true
	}
	request := oracleRequest(t, items, geometry.Dimensions{X: 1, Y: 1, Z: 3})
	// Two fragile cubes cannot support each other, but a gap is legal when
	// neither requires support. Directly enumerate those independent conditions.
	wantFeasible := false
	for a := int64(0); a+1 <= 3; a++ {
		for b := int64(0); b+1 <= 3; b++ {
			difference := a - b
			if difference < 0 {
				difference = -difference
			}
			wantFeasible = wantFeasible || difference > 1
		}
	}
	if !wantFeasible {
		t.Fatal("oracle fixture is not feasible")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	plan, err := (solver.Exact{}).PackFixed(ctx, request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{})
	if err != nil || plan.Status() != knapsack.StatusOptimal {
		t.Fatalf("legal non-contact fragile placement lost: error=%v plan=%s", err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.RequireAll()); !result.Valid() {
		t.Fatalf("independent verifier: %v", result.Violations())
	}
}

func TestExactBuiltInHeightObjectiveUsesTheCompleteLattice(t *testing.T) {
	request := oracleRequest(t, []knapsack.NormalizedItem{oracleItem(t, "item", geometry.Dimensions{X: 1, Y: 1, Z: 1})}, geometry.Dimensions{X: 1, Y: 1, Z: 3})
	goal, err := objective.New(objective.Maximize(objective.MaximumUsedHeight))
	if err != nil {
		t.Fatal(err)
	}
	wantHeight := int64(0)
	for z := int64(0); z+1 <= 3; z++ {
		if z+1 > wantHeight {
			wantHeight = z + 1
		}
	}
	plan, err := (solver.Exact{}).PackFixed(t.Context(), request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{Objective: goal})
	if err != nil || plan.Status() != knapsack.StatusOptimal || plan.Placements()[0].Origin.Z+1 != wantHeight {
		t.Fatalf("height objective falsely completed: error=%v plan=%s", err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !result.Valid() {
		t.Fatalf("independent verifier: %v", result.Violations())
	}
}

func TestExactInterruptionKeepsTheBestEarlierConfiguration(t *testing.T) {
	base := completenessRequest(t, 2, knapsack.FiniteStock(1))
	containers := base.Containers()
	containers[0].ID = "a-good"
	second := containers[0]
	second.ID = "b-worse"
	limits := base.Limits()
	limits.MaxSearchNodes = 5
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: base.Items(), Containers: append(containers, second), Resolution: base.Resolution(), Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	goal := preferContainerType{}
	plan, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{PlanObjective: goal})
	if !errors.Is(err, knapsack.ErrBudgetExhausted) || plan.Status() != knapsack.StatusBudgetExhausted || len(plan.Containers()) != 1 || plan.Containers()[0].TypeID != "a-good" {
		t.Fatalf("later interrupted worse candidate replaced the best plan: error=%v plan=%s", err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !result.Valid() {
		t.Fatalf("retained plan failed independent verifier: %v", result.Violations())
	}
}

func TestExactAggregateRefusalCannotProveOptimality(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(2))
	spec := knapsack.NormalizedSpec{Items: base.Items(), Containers: base.Containers(), Resolution: base.Resolution(), Limits: base.Limits()}
	spec.Containers[0].MaxContentWeight = math.MaxInt64
	request, err := knapsack.NewNormalizedRequest(spec)
	if err != nil {
		t.Fatal(err)
	}
	goal := preferCount{count: 2}
	plan, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{PlanObjective: goal})
	if !errors.Is(err, knapsack.ErrOverflow) || plan.Status() == knapsack.StatusOptimal || len(plan.Placements()) != 1 {
		t.Fatalf("unrepresentable preferred count falsely proved optimal: error=%v plan=%s", err, plan.CanonicalString())
	}
	if plan.Work().Nodes < 4 || plan.Work().Branches < 4 {
		t.Fatalf("refused configuration work was discarded: %+v", plan.Work())
	}
	if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !result.Valid() {
		t.Fatalf("invalid retained state: %v", result.Violations())
	}
}

func TestExactNativeCanonicalTieSearchesInteriorOrigins(t *testing.T) {
	base := completenessRequest(t, 12, knapsack.FiniteStock(1))
	reserved, err := geometry.NewCuboid(geometry.Point{}, geometry.Dimensions{X: 2, Y: 1, Z: 1})
	if err != nil {
		t.Fatal(err)
	}
	spec := knapsack.NormalizedSpec{Items: base.Items(), Containers: base.Containers(), Resolution: base.Resolution(), Limits: base.Limits()}
	spec.Containers[0].Reserved = []geometry.Cuboid{reserved}
	request, err := knapsack.NewNormalizedRequest(spec)
	if err != nil {
		t.Fatal(err)
	}
	goal, err := objective.New(objective.Minimize(objective.ContainerCount))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := (solver.Exact{}).PackFixed(t.Context(), request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, solver.Options{Objective: goal})
	if err != nil {
		t.Fatal(err)
	}
	alternativeSpec := plan.Spec()
	alternativeSpec.Placements[0].Origin.X = 10
	alternative, err := knapsack.NewPlan(alternativeSpec)
	if err != nil {
		t.Fatal(err)
	}
	if result := verify.Plan(request, alternative, verify.RequireAll().WithObjective(goal)); !result.Valid() {
		t.Fatalf("oracle alternative invalid: %v", result.Violations())
	}
	comparison, err := goal.ComparePlans(t.Context(), request, alternative, plan)
	if err != nil || comparison < 0 {
		t.Fatalf("verified interior origin beats purported optimum: comparison=%d error=%v plan=%s", comparison, err, plan.CanonicalString())
	}
}

func TestExactDoesNotSwallowObjectiveMemoryErrors(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.FiniteStock(2))
	calls := 0
	failure := fmt.Errorf("objective admission-shaped failure: %w", knapsack.ErrMemoryBudgetExhausted)
	plan, err := (solver.Exact{}).PackAll(t.Context(), request, solver.Options{PlanObjective: aggregateErrorObjective{calls: &calls, failure: failure}})
	if !errors.Is(err, failure) || calls != 1 || plan.Status() != "" {
		t.Fatalf("collaborator memory error swallowed: calls=%d error=%v plan=%s", calls, err, plan.CanonicalString())
	}
}

// Count every observable objective invocation so failure injection includes
// search, verification, best admission, final scoring and outer comparison.
type phaseObjective struct {
	components, comparisons           int
	failComponent, failComparison     int
	cancelComponent, cancelComparison int
	failure                           error
	cancel                            context.CancelFunc
}

func (*phaseObjective) Valid() bool { return true }
func (goal *phaseObjective) Components(_ context.Context, _ knapsack.NormalizedRequest, plan knapsack.Plan) ([]knapsack.ScoreComponent, error) {
	goal.components++
	if goal.components == goal.failComponent {
		return nil, goal.failure
	}
	if goal.components == goal.cancelComponent {
		goal.cancel()
	}
	return []knapsack.ScoreComponent{{Name: "phase_count", Direction: "min", Unit: "count", Value: strconv.Itoa(len(plan.Containers()))}}, nil
}
func (goal *phaseObjective) ComparePlans(_ context.Context, _ knapsack.NormalizedRequest, left, right knapsack.Plan) (int, error) {
	goal.comparisons++
	if goal.comparisons == goal.failComparison {
		return 0, goal.failure
	}
	if goal.comparisons == goal.cancelComparison {
		goal.cancel()
	}
	if len(left.Containers()) < len(right.Containers()) {
		return -1, nil
	}
	if len(left.Containers()) > len(right.Containers()) {
		return 1, nil
	}
	return 0, nil
}
func phasePack(fixed bool, ctx context.Context, request knapsack.NormalizedRequest, goal *phaseObjective) (knapsack.Plan, error) {
	options := solver.Options{PlanObjective: goal}
	if fixed {
		return (solver.Exact{}).PackFixed(ctx, request, []knapsack.ContainerInstance{{ID: "box#1", TypeID: "box"}}, options)
	}
	return (solver.Exact{}).PackAll(ctx, request, options)
}
func TestExactObjectiveFailuresRemainFatalAtEveryPhase(t *testing.T) {
	request := completenessRequest(t, 2, knapsack.FiniteStock(2))
	failures := []struct {
		name string
		err  error
	}{
		{"ordinary", errors.New("synthetic objective failure")},
		{"budget", knapsack.ErrBudgetExhausted}, {"memory", knapsack.ErrMemoryBudgetExhausted},
		{"canceled-live-context", context.Canceled}, {"deadline-live-context", context.DeadlineExceeded},
		{"proof", knapsack.ErrProvenInfeasible}, {"overflow", knapsack.ErrOverflow},
	}
	for _, fixed := range []bool{true, false} {
		name := "variable"
		if fixed {
			name = "fixed"
		}
		control := &phaseObjective{}
		plan, err := phasePack(fixed, t.Context(), request, control)
		if err != nil || plan.Status() != knapsack.StatusOptimal {
			t.Fatalf("%s control: plan=%s error=%v", name, plan.CanonicalString(), err)
		}
		for _, failure := range failures {
			for _, phase := range []struct {
				name  string
				calls int
			}{{"components", control.components}, {"compare", control.comparisons}} {
				for failAt := 1; failAt <= phase.calls; failAt++ {
					t.Run(fmt.Sprintf("%s/%s/%s/%d", name, failure.name, phase.name, failAt), func(t *testing.T) {
						injected := fmt.Errorf("phase-specific objective error: %w", failure.err)
						goal := &phaseObjective{failure: injected}
						if phase.name == "components" {
							goal.failComponent = failAt
						} else {
							goal.failComparison = failAt
						}
						plan, err := phasePack(fixed, t.Context(), request, goal)
						if !errors.Is(err, injected) || plan.Status() != "" || len(plan.Placements()) != 0 || len(plan.Containers()) != 0 {
							t.Fatalf("callback error became solver evidence: error=%v plan=%s", err, plan.CanonicalString())
						}
					})
				}
			}
		}
	}
}

// Cancel the real caller context from objective verification, after valid
// components were scored. Cancellation must retain a valid partial plan rather
// than become an invariant failure returning a zero plan.
func TestExactVerificationCancellationRetainsPartialItemConservation(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.FiniteStock(1))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	goal := &phaseObjective{cancel: cancel, cancelComponent: 2}
	plan, err := phasePack(true, ctx, request, goal)
	if goal.components != 2 || !errors.Is(err, context.Canceled) {
		t.Fatalf("verification cancellation witness: components=%d error=%v", goal.components, err)
	}
	if plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationCancelled {
		t.Errorf("verification cancellation lost partial result: status=%q termination=%q", plan.Status(), plan.Termination())
	}
	seen := make(map[string]int)
	for _, placement := range plan.Placements() {
		seen[placement.ItemID]++
	}
	for _, id := range plan.UnpackedItemIDs() {
		seen[id]++
	}
	if len(seen) != request.ItemCount() {
		t.Errorf("partial plan lost item conservation: got %d IDs, want %d", len(seen), request.ItemCount())
	}
	for _, item := range request.Items() {
		if seen[item.ID] != 1 {
			t.Errorf("item %q appears %d times across placements and unpacked IDs", item.ID, seen[item.ID])
		}
	}
	if result := verify.Plan(request, plan, verify.AllowUnpacked()); !result.Valid() {
		t.Errorf("retained cancellation partial plan is invalid: %+v", result)
	}
}

func TestExactObjectiveCancellationStopsReplayAtEveryPhase(t *testing.T) {
	request := completenessRequest(t, 2, knapsack.FiniteStock(2))
	for _, fixed := range []bool{true, false} {
		name := "variable"
		if fixed {
			name = "fixed"
		}
		control := &phaseObjective{}
		if _, err := phasePack(fixed, t.Context(), request, control); err != nil {
			t.Fatal(err)
		}
		for _, phase := range []struct {
			name  string
			calls int
		}{{"components", control.components}, {"compare", control.comparisons}} {
			for cancelAt := 1; cancelAt <= phase.calls; cancelAt++ {
				t.Run(fmt.Sprintf("%s/%s/%d", name, phase.name, cancelAt), func(t *testing.T) {
					ctx, cancel := context.WithCancel(t.Context())
					defer cancel()
					goal := &phaseObjective{cancel: cancel}
					if phase.name == "components" {
						goal.cancelComponent = cancelAt
					} else {
						goal.cancelComparison = cancelAt
					}
					plan, err := phasePack(fixed, ctx, request, goal)
					calls := goal.components
					if phase.name == "compare" {
						calls = goal.comparisons
					}
					if !errors.Is(err, context.Canceled) || plan.Status() == knapsack.StatusOptimal || calls != cancelAt {
						t.Fatalf("canceled operation replayed objective or certified proof: calls=%d want=%d error=%v plan=%s", calls, cancelAt, err, plan.CanonicalString())
					}
				})
			}
		}
	}
}

func TestExactLateUnitObjectiveError(t *testing.T) {
	request := completenessRequest(t, 1, knapsack.FiniteStock(1))
	for _, failAt := range []int{4, 5} {
		t.Run(strconv.Itoa(failAt), func(t *testing.T) {
			failure := fmt.Errorf("late unit objective: %w", knapsack.ErrMemoryBudgetExhausted)
			goal := &phaseObjective{failComponent: failAt, failure: failure}
			plan, err := phasePack(false, t.Context(), request, goal)
			if !errors.Is(err, failure) || plan.Status() != "" {
				t.Fatalf("late callback error became solver interruption: components=%d error=%v plan=%s", goal.components, err, plan.CanonicalString())
			}
		})
	}
}

func TestExactInterruptedSelectionPreservesOwnedErrorAndRetainedState(t *testing.T) {
	request := completenessRequest(t, 2, knapsack.FiniteStock(2))
	limits := request.Limits()
	limits.MaxSearchNodes = 5
	request = request.WithLimits(limits)
	control := &phaseObjective{}
	retained, err := phasePack(false, t.Context(), request, control)
	if !errors.Is(err, knapsack.ErrBudgetExhausted) || retained.Status() != knapsack.StatusBudgetExhausted || len(retained.Placements()) != 1 {
		t.Fatalf("interrupted control lost feasible state: error=%v plan=%s", err, retained.CanonicalString())
	}
	for _, failure := range []error{knapsack.ErrBudgetExhausted, knapsack.ErrMemoryBudgetExhausted, knapsack.ErrProvenInfeasible, context.Canceled, context.DeadlineExceeded} {
		goal := &phaseObjective{failure: failure, failComparison: control.comparisons}
		plan, err := phasePack(false, t.Context(), request, goal)
		if !errors.Is(err, failure) || plan.Status() != "" {
			t.Fatalf("selection callback error became solver evidence: error=%v plan=%s", err, plan.CanonicalString())
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	goal := &phaseObjective{cancel: cancel, cancelComparison: control.comparisons}
	plan, err := phasePack(false, ctx, request, goal)
	if !errors.Is(err, context.Canceled) || plan.Status() != knapsack.StatusBudgetExhausted || len(plan.Placements()) != 1 || goal.comparisons != control.comparisons {
		t.Fatalf("canceled selection discarded retained plan or replayed provider: comparisons=%d error=%v plan=%s", goal.comparisons, err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(&phaseObjective{})); !result.Valid() {
		t.Fatalf("canceled retained plan invalid: %v", result.Violations())
	}
}

func TestExactAggregateFinalVerificationPreservesOwnedCancellation(t *testing.T) {
	base := completenessRequest(t, 1, knapsack.FiniteStock(2))
	containers := base.Containers()
	containers[0].MaxContentWeight = math.MaxInt64
	request, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: base.Items(), Containers: containers, Resolution: base.Resolution(), Limits: base.Limits()})
	if err != nil {
		t.Fatal(err)
	}
	control := &phaseObjective{}
	plan, err := phasePack(false, t.Context(), request, control)
	if !errors.Is(err, knapsack.ErrOverflow) || plan.Status() != knapsack.StatusFeasible {
		t.Fatalf("aggregate refusal control: error=%v plan=%s", err, plan.CanonicalString())
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	goal := &phaseObjective{cancel: cancel, cancelComponent: control.components}
	plan, err = phasePack(false, ctx, request, goal)
	if !errors.Is(err, context.Canceled) || plan.Status() != knapsack.StatusBudgetExhausted || len(plan.Placements()) != 1 || goal.components != control.components {
		t.Fatalf("aggregate final verification discarded retained plan or replayed provider: calls=%d error=%v plan=%s", goal.components, err, plan.CanonicalString())
	}
	if result := verify.Plan(request, plan, verify.RequireAll().WithObjective(&phaseObjective{})); !result.Valid() {
		t.Fatalf("aggregate canceled retained plan invalid: %v", result.Violations())
	}
}
