package constraint_test

import (
	"errors"
	"strings"
	"testing"

	c "github.com/faustbrian/go-knapsack/v2/constraint"
)

func TestConstraintDecisionByteAndFieldLimits(t *testing.T) {
	cases := []struct {
		name, code, message string
		invalid             bool
	}{
		{"empty-code", "", "message", true},
		{"empty-message", "code", "", true},
		{"below-limit", strings.Repeat("c", 63), strings.Repeat("m", 1023), false},
		{"exact-limit", strings.Repeat("c", 64), strings.Repeat("m", 1024), false},
		{"multibyte-exact", strings.Repeat("é", 32), strings.Repeat("é", 512), false},
		{"code-over-limit", strings.Repeat("c", 65), strings.Repeat("m", 1024), true},
		{"message-over-limit", strings.Repeat("c", 64), strings.Repeat("m", 1025), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := c.Reject(tc.code, tc.message)
			got, err := c.ValidateDecision(input)
			if tc.invalid {
				if !errors.Is(err, c.ErrInvalidDecision) || got != (c.Decision{}) {
					t.Fatalf("invalid rejection accepted: got=%#v err=%v", got, err)
				}
				return
			}
			if err != nil || got != input {
				t.Fatalf("valid bounded decision changed: got=%#v err=%v", got, err)
			}
		})
	}
}
