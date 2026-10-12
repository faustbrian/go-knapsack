// Package knapsacktest provides solver-independent assertions for consumers,
// fixtures, examples, differential adapters, and benchmarks. These helpers
// borrow testing.TB and can print caller identifiers or full plan bytes; they
// are not redaction or exhaustive-proof facilities.
package knapsacktest

import (
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/verify"
)

// RequireVerified fails on verifier findings. It checks Valid, not the stored
// extension error; custom-option callers must check verify.PlanContext errors too.
func RequireVerified(t testing.TB, request knapsack.NormalizedRequest, plan knapsack.Plan, options verify.Options) {
	t.Helper()
	result := verify.Plan(request, plan, options)
	if !result.Valid() {
		t.Fatalf("plan verification failed: %+v", result.Violations())
	}
}

// RequireCanonicalEqual fails the test unless both plans have identical
// deterministic canonical encodings.
func RequireCanonicalEqual(t testing.TB, left, right knapsack.Plan) {
	t.Helper()
	if left.CanonicalString() != right.CanonicalString() {
		t.Fatalf("plans differ\nleft:  %s\nright: %s", left.CanonicalString(), right.CanonicalString())
	}
}
