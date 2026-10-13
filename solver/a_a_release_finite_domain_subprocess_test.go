package solver

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

func TestExactFiniteDomainTerminatesAfterLastConfiguration(t *testing.T) {
	const childMarker = "GOLIB_KNAPSACK_FINITE_DOMAIN_CHILD"
	if os.Getenv(childMarker) != "1" {
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, "-test.run=^TestExactFiniteDomainTerminatesAfterLastConfiguration$", "-test.timeout=1500ms")
		for _, entry := range os.Environ() {
			// The parent owns native Gremlins' profile lock. This outcome
			// subprocess must not acquire that same lock before running.
			if !strings.HasPrefix(entry, "GOLIB_GREMLINS_COVERAGE_PROFILE=") && !strings.HasPrefix(entry, childMarker+"=") {
				command.Env = append(command.Env, entry)
			}
		}
		command.Env = append(command.Env, childMarker+"=1")
		output, err := command.CombinedOutput()
		if ctx.Err() != nil {
			t.Fatalf("finite-domain child exceeded owning outer bound: %v", ctx.Err())
		}
		if err != nil {
			t.Fatalf("finite-domain outcome failed: %v\n%s", err, output)
		}
		return
	}
	if os.Getenv("GOLIB_GREMLINS_COVERAGE_PROFILE") != "" {
		t.Fatal("owning child inherited the parent's mutation-profile lock identity")
	}
	base := internalRequest(t)
	items, containers := base.Items(), base.Containers()
	items[0].Dimensions = geometry.Dimensions{X: 1, Y: 1, Z: 1}
	items[0].Orientations = []geometry.Orientation{geometry.OrientationXYZ}
	containers[0].Dimensions = items[0].Dimensions
	containers[0].Stock = knapsack.FiniteStock(1)
	for _, domain := range []string{"finite_one", "unlimited_native_minimum", "finite_two_types"} {
		selectedItems, selectedContainers := append([]knapsack.NormalizedItem(nil), items...), append([]knapsack.NormalizedContainer(nil), containers...)
		limits := base.Limits()
		limits.MaxBranches, limits.MaxCandidatePlacements, limits.MaxSearchNodes = 16, 16, 16
		if domain == "unlimited_native_minimum" {
			selectedContainers[0].Stock = knapsack.UnlimitedStock()
		}
		if domain == "finite_two_types" {
			selectedItems[0].ID = "first"
			second := selectedItems[0]
			second.ID = "second"
			selectedItems = append(selectedItems, second)
			selectedContainers[0].ID = "a"
			secondType := selectedContainers[0]
			secondType.ID = "b"
			selectedContainers = append(selectedContainers, secondType)
		}
		request, admissionErr := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: selectedItems, Containers: selectedContainers, Resolution: base.Resolution(), Limits: limits})
		if admissionErr != nil {
			t.Fatal(admissionErr)
		}
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		type result struct {
			plan knapsack.Plan
			err  error
		}
		completed := make(chan result, 1)
		go func() {
			plan, err := (Exact{}).PackAll(ctx, request, Options{})
			completed <- result{plan: plan, err: err}
		}()
		select {
		case got := <-completed:
			cancel()
			// Unit cubes require one unit bin per item. The two-type stock
			// fixture requires both types; native minimum-count ordering gives
			// the same finite optimum even with unlimited stock.
			if got.err != nil || got.plan.Status() != knapsack.StatusOptimal || len(got.plan.Placements()) != len(selectedItems) || len(got.plan.Containers()) != len(selectedItems) {
				t.Fatalf("%s finite optimum lost: %s, %v", domain, got.plan.CanonicalString(), got.err)
			}
			if checked := verify.Plan(request, got.plan, verify.RequireAll()); !checked.Valid() {
				t.Fatalf("%s finite optimum invalid: %v", domain, checked.Violations())
			}
		case <-time.After(time.Second):
			cancel()
			// Fail before the native/engine watchdog. This owned child exits
			// and is joined; no forced borrowed callback preemption is claimed.
			t.Fatalf("%s finite search neither completed nor returned after actual context deadline", domain)
		}
	}
}

func TestExactCancellationBetweenConfigurationAndFixedAdmissionRetainsOwnedState(t *testing.T) {
	request := internalRequest(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	// Existing owning fault-window helper cancels the real context at fixed
	// admission, after variable configuration admission. Err and Done always
	// delegate to that real monotonic context; no execution flags are fabricated.
	window := &releaseCancellationWindow{Context: ctx, cancel: cancel, liveChecks: 3}
	plan, err := (Exact{}).PackAll(window, request, Options{})
	if !errors.Is(err, context.Canceled) || plan.Status() != knapsack.StatusBudgetExhausted || plan.Termination() != knapsack.TerminationCancelled || len(plan.Placements()) != 0 || len(plan.UnpackedItemIDs()) != request.ItemCount() {
		t.Fatalf("fixed-admission cancellation lost owned item state: %s, %v", plan.CanonicalString(), err)
	}
	select {
	case <-window.Done():
	default:
		t.Fatal("actual cancellation did not close Done")
	}
	if !errors.Is(window.Err(), context.Canceled) || !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatal("actual cancellation error did not remain monotonic")
	}
	if checked := verify.Plan(request, plan, verify.AllowUnpacked()); !checked.Valid() {
		t.Fatalf("fixed-admission cancellation state invalid: %v", checked.Violations())
	}
}
