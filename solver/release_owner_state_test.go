package solver

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/objective"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

// The public Exact entry point precomputes full grids. This private fallback
// still owns its admission and rollback contract; do not claim a public route.
func TestPrivateExactFallbackRefusesPointMemoryWithoutPublishingProof(t *testing.T) {
	request := internalRequest(t)
	goal, err := objective.New(objective.Minimize(objective.ContainerCount))
	if err != nil {
		t.Fatal(err)
	}
	for _, allowance := range []uint64{0, 24} {
		target := baseInternalBin()
		search := &exactSearch{ctx: context.Background(), request: request, limits: request.Limits(), bins: []*bin{target}, goal: goal, pointMemory: allowance}
		search.visit(request.Items())
		if len(target.placements) != 0 || len(target.items) != 0 || target.weight != 0 {
			t.Fatal("fallback search leaked trial placement state")
		}
		if allowance == 0 {
			if !search.budgeted || search.termination != knapsack.TerminationMemoryLimit || !errors.Is(search.budgetCause, knapsack.ErrMemoryBudgetExhausted) || search.best != nil {
				t.Fatalf("point refusal changed evidence: termination=%s error=%v best=%v", search.termination, search.budgetCause, search.best)
			}
			partial, err := search.partial(request.Items(), search.termination, search.budgetCause)
			if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || partial.Status() != knapsack.StatusBudgetExhausted || len(partial.Placements()) != 0 || len(partial.UnpackedItemIDs()) != 1 {
				t.Fatalf("point refusal published proof or lost item: %s, %v", partial.CanonicalString(), err)
			}
			if checked := verify.Plan(request, partial, verify.AllowUnpacked()); !checked.Valid() {
				t.Fatalf("invalid refusal state: %v", checked.Violations())
			}
		} else {
			if search.budgeted || search.best == nil || len(search.best.Placements()) != 1 {
				t.Fatal("admitted private point failed to retain feasible candidate")
			}
			if checked := verify.Plan(request, *search.best, verify.RequireAll().WithObjective(goal)); !checked.Valid() {
				t.Fatalf("invalid admitted candidate: %v", checked.Violations())
			}
		}
	}
}

func TestPrivateInterruptedExactPlanRejectsUnknownTermination(t *testing.T) {
	request := internalRequest(t)
	cause := errors.New("owned interruption")
	plan, err := interruptedExactVariable(request.Items(), nil, knapsack.Work{}, 0, 7, knapsack.TerminationReason("unsupported"), cause)
	if !errors.Is(err, knapsack.ErrInvalidRequest) || plan.Status() != "" {
		t.Fatalf("unknown private termination published state: %s,%v", plan.CanonicalString(), err)
	}
	plan, err = interruptedExactVariable(request.Items(), nil, knapsack.Work{}, 0, 7, knapsack.TerminationNodeLimit, cause)
	if !errors.Is(err, cause) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationNodeLimit || plan.Work().Seed != 7 || len(plan.UnpackedItemIDs()) != 1 {
		t.Fatalf("valid interruption lost owned cause/state: %s,%v", plan.CanonicalString(), err)
	}
}

// Cancellation can become visible between admission checks without a callback.
// Delegate every observed error and Done channel to a real monotonic context;
// the fault window changes, but no error can disappear after cancellation.
type releaseCancellationWindow struct {
	context.Context
	cancel     context.CancelFunc
	liveChecks int
}

func (window *releaseCancellationWindow) Err() error {
	if window.liveChecks == 0 {
		window.cancel()
	} else {
		window.liveChecks--
	}
	return window.Context.Err()
}
func TestCancellationAcrossAdmissionAndSearchPreservesOwnedState(t *testing.T) {
	request := internalRequest(t)
	goal, err := objective.New(objective.Minimize(objective.ContainerCount))
	if err != nil {
		t.Fatal(err)
	}
	for _, heuristic := range []bool{false, true} {
		for liveChecks := range 24 {
			ctx, cancel := context.WithCancel(t.Context())
			window := &releaseCancellationWindow{Context: ctx, cancel: cancel, liveChecks: liveChecks}
			var plan knapsack.Plan
			if heuristic {
				plan, err = (Heuristic{}).packAllPass(window, request, Options{}, goal, comparePlacement)
			} else {
				plan, err = (Exact{}).PackAll(window, request, Options{})
			}
			cancel()
			if err == nil {
				if checked := verify.Plan(request, plan, verify.RequireAll().WithObjective(goal)); !checked.Valid() {
					t.Fatalf("completed cancellation-window control invalid: %v", checked.Violations())
				}
				continue
			}
			if !errors.Is(err, context.Canceled) || !errors.Is(window.Context.Err(), context.Canceled) {
				t.Fatalf("cancellation lost original error: heuristic=%t window=%d error=%v", heuristic, liveChecks, err)
			}
			select {
			case <-window.Done():
			default:
				t.Fatal("cancelled error without closed Done")
			}
			if plan.Status() == "" {
				continue
			}
			if plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationCancelled {
				t.Fatalf("cancellation published proof: %s", plan.CanonicalString())
			}
			if checked := verify.Plan(request, plan, verify.AllowUnpacked()); !checked.Valid() {
				t.Fatalf("cancellation lost item conservation: %v", checked.Violations())
			}
		}
	}
}
