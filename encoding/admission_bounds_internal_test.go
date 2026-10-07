package encoding

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	knapsack "github.com/faustbrian/go-knapsack/v2"
)

type admissionDecoder struct {
	name   string
	decode func([]byte, Limits) ([]byte, bool, error)
}

func admissionDecoders() []admissionDecoder {
	return []admissionDecoder{
		{"plan", func(input []byte, limits Limits) ([]byte, bool, error) {
			value, err := UnmarshalPlan(input, limits)
			zero := reflect.DeepEqual(value, knapsack.Plan{})
			if err != nil {
				return nil, zero, err
			}
			encoded, err := MarshalPlan(value)
			return encoded, zero, err
		}},
		{"request", func(input []byte, limits Limits) ([]byte, bool, error) {
			value, err := UnmarshalRequest(input, limits)
			zero := reflect.DeepEqual(value, knapsack.NormalizedRequest{})
			if err != nil {
				return nil, zero, err
			}
			encoded, err := MarshalRequest(value)
			return encoded, zero, err
		}},
	}
}

func TestDecodeRejectsEachNonpositiveLimitBeforeParsing(t *testing.T) {
	for _, decoder := range admissionDecoders() {
		for _, name := range []string{"bytes", "depth", "collection"} {
			for _, value := range []int{0, -1} {
				limits := DefaultLimits()
				switch name {
				case "bytes":
					limits.MaxBytes = int64(value)
				case "depth":
					limits.MaxDepth = value
				case "collection":
					limits.MaxCollection = value
				}
				_, zero, err := decoder.decode(nil, limits)
				if !zero || !errors.Is(err, ErrEncodingLimit) {
					t.Fatalf("%s %s=%d: zero=%v error=%v", decoder.name, name, value, zero, err)
				}
			}
		}
		_, zero, err := decoder.decode(nil, Limits{MaxBytes: 1, MaxDepth: 1, MaxCollection: 1})
		if !zero || !errors.Is(err, ErrInvalidEncoding) || !errors.Is(err, io.EOF) {
			t.Fatalf("%s positive limits did not reach parsing: %v", decoder.name, err)
		}
	}
}

func TestDecodePreservesInclusiveFixtureByteBudget(t *testing.T) {
	for _, decoder := range admissionDecoders() {
		input, err := os.ReadFile(filepath.Join("testdata", "v1", decoder.name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		input = bytes.TrimSpace(input)
		for _, extra := range []int64{-1, 0, 1} {
			limits := DefaultLimits()
			limits.MaxBytes = int64(len(input)) + extra
			encoded, zero, err := decoder.decode(input, limits)
			if extra < 0 {
				if !zero || !errors.Is(err, ErrEncodingLimit) {
					t.Fatalf("%s excess fixture admitted: %v", decoder.name, err)
				}
			} else if err != nil || zero || !bytes.Equal(encoded, input) {
				t.Fatalf("%s inclusive fixture budget changed artifact: %v", decoder.name, err)
			}
		}
	}
}

func TestStrictValidationPreservesInclusiveTraversalBudgets(t *testing.T) {
	for _, test := range []struct {
		name, input       string
		depth, collection int
		limited           bool
	}{
		{"scalar depth", `1`, 1, 1, false},
		{"empty object", `{}`, 1, 1, false},
		{"object exact depth", `{"a":1}`, 2, 1, false},
		{"object excess depth", `{"a":1}`, 1, 1, true},
		{"nested object exact depth", `{"a":{"b":1}}`, 3, 1, false},
		{"nested object excess depth", `{"a":{"b":1}}`, 2, 1, true},
		{"object excess count", `{"a":1,"b":2}`, 2, 1, true},
		{"object exact count", `{"a":1,"b":2}`, 2, 2, false},
		{"empty array", `[]`, 1, 1, false},
		{"array exact count", `[1]`, 2, 1, false},
		{"array excess count", `[1,2]`, 2, 1, true},
		{"array below count", `[1]`, 2, 2, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateStrict([]byte(test.input), Limits{MaxBytes: 1024, MaxDepth: test.depth, MaxCollection: test.collection})
			if test.limited {
				if !errors.Is(err, ErrEncodingLimit) {
					t.Fatalf("excess traversal admitted: %v", err)
				}
			} else if err != nil {
				t.Fatalf("inclusive traversal rejected: %v", err)
			}
		})
	}
}

func TestDecodePreservesTrailingSyntaxCauseWithoutInput(t *testing.T) {
	for _, decoder := range admissionDecoders() {
		_, zero, err := decoder.decode([]byte(`{},`), DefaultLimits())
		var syntax *json.SyntaxError
		if !zero || !errors.Is(err, ErrInvalidEncoding) || !errors.As(err, &syntax) || err.Error() != ErrInvalidEncoding.Error() {
			t.Fatalf("%s trailing syntax classification changed: %v", decoder.name, err)
		}
		_, zero, err = decoder.decode([]byte(`{} 1`), DefaultLimits())
		if !zero || !errors.Is(err, ErrInvalidEncoding) || err.Error() != ErrInvalidEncoding.Error() {
			t.Fatalf("%s trailing value admitted or exposed: %v", decoder.name, err)
		}
	}
}
