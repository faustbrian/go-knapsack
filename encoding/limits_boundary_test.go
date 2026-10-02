package encoding_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	packingjson "github.com/faustbrian/go-knapsack/v2/encoding"
)

// Each entry exercises a public decoder with a persisted, canonical v1 artifact.
func boundaryDecoders(t *testing.T) []struct {
	name   string
	input  []byte
	decode func([]byte, packingjson.Limits) ([]byte, error)
} {
	t.Helper()
	read := func(name string) []byte {
		t.Helper()
		input, err := os.ReadFile("testdata/v1/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		return bytes.TrimSuffix(input, []byte{'\n'})
	}
	return []struct {
		name   string
		input  []byte
		decode func([]byte, packingjson.Limits) ([]byte, error)
	}{
		{"plan", read("plan"), func(input []byte, limits packingjson.Limits) ([]byte, error) {
			plan, err := packingjson.UnmarshalPlan(input, limits)
			if err != nil {
				if !reflect.DeepEqual(plan, knapsack.Plan{}) {
					t.Fatal("rejected input returned a nonzero plan")
				}
				return nil, err
			}
			return packingjson.MarshalPlan(plan)
		}},
		{"request", read("request"), func(input []byte, limits packingjson.Limits) ([]byte, error) {
			request, err := packingjson.UnmarshalRequest(input, limits)
			if err != nil {
				if !reflect.DeepEqual(request, knapsack.NormalizedRequest{}) {
					t.Fatal("rejected input returned a nonzero request")
				}
				return nil, err
			}
			return packingjson.MarshalRequest(request)
		}},
	}
}

func TestDecodeRejectsEachInvalidLimitBeforeParsing(t *testing.T) {
	for _, decoder := range boundaryDecoders(t) {
		t.Run(decoder.name, func(t *testing.T) {
			for _, field := range []string{"bytes", "depth", "collection"} {
				for _, value := range []int{0, -1} {
					t.Run(fmt.Sprintf("%s=%d", field, value), func(t *testing.T) {
						limits := packingjson.DefaultLimits()
						switch field {
						case "bytes":
							limits.MaxBytes = int64(value)
						case "depth":
							limits.MaxDepth = value
						case "collection":
							limits.MaxCollection = value
						}
						// Empty input also proves policy rejection precedes parsing.
						for _, input := range [][]byte{nil, decoder.input} {
							if _, err := decoder.decode(input, limits); !errors.Is(err, packingjson.ErrEncodingLimit) {
								t.Fatalf("invalid policy error = %v", err)
							}
						}
					})
				}
			}
		})
	}
}

func TestDecodeInclusiveByteLimit(t *testing.T) {
	for _, decoder := range boundaryDecoders(t) {
		t.Run(decoder.name, func(t *testing.T) {
			for _, delta := range []int64{-1, 0, 1} {
				limits := packingjson.DefaultLimits()
				limits.MaxBytes = int64(len(decoder.input)) + delta
				encoded, err := decoder.decode(decoder.input, limits)
				if delta < 0 {
					if !errors.Is(err, packingjson.ErrEncodingLimit) {
						t.Fatalf("below byte size error = %v", err)
					}
					continue
				}
				if err != nil || !bytes.Equal(encoded, decoder.input) {
					t.Fatalf("byte bound delta=%d changed canonical artifact: %v", delta, err)
				}
			}
		})
	}
}

func TestPlanInclusiveDepthAndCollectionLimits(t *testing.T) {
	const minimal = `{"version":"v1","plan":{"status":"feasible","termination":"completed"}}`
	const withArray = `{"version":"v1","plan":{"status":"feasible","termination":"completed","unpacked_item_ids":["a","b","c"]}}`
	for _, test := range []struct {
		name, input       string
		depth, collection int
		wantLimit         bool
	}{
		// Current traversal counts root, plan, and scalar values at depths 1, 2, 3.
		{"depth-below", minimal, 2, 10, true},
		{"depth-exact", minimal, 3, 10, false},
		{"depth-above", minimal, 4, 10, false},
		{"object-below", minimal, 10, 1, true},
		{"object-exact", minimal, 10, 2, false},
		{"object-above", minimal, 10, 3, false},
		{"array-exact", withArray, 10, 3, false},
		{"array-above", withArray, 10, 4, false},
		{"array-over", `{"version":"v1","plan":{"status":"feasible","termination":"completed","unpacked_item_ids":["a","b","c","d"]}}`, 10, 3, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			limits := packingjson.Limits{MaxBytes: 1024, MaxDepth: test.depth, MaxCollection: test.collection}
			plan, err := packingjson.UnmarshalPlan([]byte(test.input), limits)
			if test.wantLimit {
				if !errors.Is(err, packingjson.ErrEncodingLimit) || !reflect.DeepEqual(plan, knapsack.Plan{}) {
					t.Fatalf("rejected boundary plan=%+v error=%v", plan, err)
				}
				return
			}
			if err != nil || plan.Spec().Status != knapsack.StatusFeasible || plan.Spec().Termination != knapsack.TerminationCompleted {
				t.Fatalf("accepted boundary plan=%+v error=%v", plan, err)
			}
			if test.input == withArray && !reflect.DeepEqual(plan.UnpackedItemIDs(), []string{"a", "b", "c"}) {
				t.Fatalf("array content = %v", plan.UnpackedItemIDs())
			}
		})
	}
}

func TestDecodeTrailingDataPreservesCauseAndRedaction(t *testing.T) {
	for _, decoder := range boundaryDecoders(t) {
		t.Run(decoder.name, func(t *testing.T) {
			for _, suffix := range []string{" x", " {}", " \n\t"} {
				input := append(bytes.Clone(decoder.input), []byte(suffix)...)
				encoded, err := decoder.decode(input, packingjson.DefaultLimits())
				if suffix == " \n\t" {
					if err != nil || !bytes.Equal(encoded, decoder.input) {
						t.Fatalf("whitespace changed canonical artifact: %v", err)
					}
					continue
				}
				if !errors.Is(err, packingjson.ErrInvalidEncoding) || err.Error() != packingjson.ErrInvalidEncoding.Error() {
					t.Fatalf("trailing-data classification/redaction = %v", err)
				}
				var syntax *json.SyntaxError
				if suffix == " x" && !errors.As(err, &syntax) {
					t.Fatalf("trailing syntax cause lost: %v", err)
				}
			}
		})
	}
}
