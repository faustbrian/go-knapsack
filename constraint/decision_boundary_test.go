package constraint_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-knapsack/v2/constraint"
)

func TestRejectionDecisionDiagnosticBoundaries(t *testing.T) {
	for _, test := range []struct {
		name    string
		code    string
		message string
		refuse  bool
	}{
		{"absent code", "", "ordinary message", true},
		{"absent message", "ordinary-code", "", true},
		{"exact code length", strings.Repeat("c", 64), "ordinary message", false},
		{"one over code length", strings.Repeat("c", 65), "ordinary message", true},
		{"exact message length", "ordinary-code", strings.Repeat("m", 1024), false},
		{"one over message length", "ordinary-code", strings.Repeat("m", 1025), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := constraint.Reject(test.code, test.message)
			got, err := constraint.ValidateDecision(input)
			if test.refuse {
				if !errors.Is(err, constraint.ErrInvalidDecision) || got != (constraint.Decision{}) {
					t.Fatalf("refused decision = %+v, %v", got, err)
				}
			} else if err != nil || got != input {
				t.Fatalf("accepted rejection = %+v, %v, want unchanged decision", got, err)
			}
		})
	}
}
