package moneyobjective

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-international/currency"
	"github.com/faustbrian/go-knapsack"
	"github.com/faustbrian/go-knapsack/objective"
	"github.com/faustbrian/go-money"
)

var _ objective.PlanObjective = Costs{}

func TestPublicIdentityAndStableErrors(t *testing.T) {
	t.Parallel()

	wantErrors := map[error]string{
		ErrInvalidCosts:     "gomoney objective: invalid costs",
		ErrMissingCost:      "gomoney objective: missing container cost",
		ErrDuplicateTypeID:  "gomoney objective: duplicate container type ID",
		ErrUnsupportedScale: "gomoney objective: unsupported money scale",
		ErrNegativeCost:     "gomoney objective: negative container cost",
	}
	for err, want := range wantErrors {
		if err.Error() != want {
			t.Errorf("error text = %q, want %q", err, want)
		}
	}

	const packagePath = "github.com/faustbrian/go-knapsack/objective/money"
	for _, value := range []any{Limits{}, Policy{}, Entry{}, Costs{}} {
		typeOf := reflect.TypeOf(value)
		if typeOf.PkgPath() != packagePath {
			t.Errorf("%s package path = %q, want %q", typeOf.Name(), typeOf.PkgPath(), packagePath)
		}
	}

	assertFields(t, reflect.TypeFor[Limits](), []string{"MaxTypes", "MaxIDBytes"})
	assertFields(t, reflect.TypeFor[Policy](), []string{"Limits", "AllowNegativeCosts"})
	assertFields(t, reflect.TypeFor[Entry](), []string{"TypeID", "Cost"})
}

func TestConstructionCopiesInputsAndRetainsContextPrecedence(t *testing.T) {
	t.Parallel()

	euro, err := currency.Parse("EUR")
	if err != nil {
		t.Fatal(err)
	}
	moneyContext, err := money.DefaultContext(euro)
	if err != nil {
		t.Fatal(err)
	}
	one, err := money.Parse("1.00", euro, moneyContext)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]money.Money{"box": one}
	costs, err := New(values)
	if err != nil {
		t.Fatal(err)
	}
	clear(values)
	if !costs.Valid() {
		t.Fatal("constructed costs are invalid after caller map mutation")
	}

	limits := DefaultLimits()
	if limits != (Limits{MaxTypes: 1_000, MaxIDBytes: 1_024}) {
		t.Fatalf("DefaultLimits() = %#v", limits)
	}
	if DefaultPolicy() != (Policy{Limits: limits}) {
		t.Fatalf("DefaultPolicy() = %#v", DefaultPolicy())
	}

	var nilContext context.Context
	if _, err := (Costs{}).ComparePlans(nilContext, knapsack.NormalizedRequest{}, knapsack.Plan{}, knapsack.Plan{}); !errors.Is(err, knapsack.ErrInvalidOptions) {
		t.Fatalf("ComparePlans(nil) error = %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Costs{}).Components(canceled, knapsack.NormalizedRequest{}, knapsack.Plan{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Components(canceled) error = %v", err)
	}
}

func assertFields(t *testing.T, typeOf reflect.Type, want []string) {
	t.Helper()
	if typeOf.NumField() != len(want) {
		t.Fatalf("%s has %d fields, want %d", typeOf.Name(), typeOf.NumField(), len(want))
	}
	for index, name := range want {
		if typeOf.Field(index).Name != name {
			t.Errorf("%s field %d = %q, want %q", typeOf.Name(), index, typeOf.Field(index).Name, name)
		}
	}
}
