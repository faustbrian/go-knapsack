package gomoney_test

import (
	"reflect"
	"testing"

	"github.com/faustbrian/go-knapsack/objective/gomoney"
	moneyobjective "github.com/faustbrian/go-knapsack/objective/money"
	"github.com/faustbrian/go-money"
)

func TestCompatibilityFacadeRetainsIdentityAndDelegatesBehavior(t *testing.T) {
	t.Parallel()

	for _, pair := range [][2]error{
		{gomoney.ErrInvalidCosts, moneyobjective.ErrInvalidCosts},
		{gomoney.ErrMissingCost, moneyobjective.ErrMissingCost},
		{gomoney.ErrDuplicateTypeID, moneyobjective.ErrDuplicateTypeID},
		{gomoney.ErrUnsupportedScale, moneyobjective.ErrUnsupportedScale},
		{gomoney.ErrNegativeCost, moneyobjective.ErrNegativeCost},
	} {
		if pair[0] != pair[1] { //nolint:errorlint // Direct sentinel identity is the compatibility contract.
			t.Fatalf("legacy sentinel %q does not equal successor sentinel", pair[0])
		}
	}

	const legacyPath = "github.com/faustbrian/go-knapsack/objective/gomoney"
	for _, value := range []any{
		gomoney.Limits{}, gomoney.Policy{}, gomoney.Entry{}, gomoney.Costs{},
	} {
		typeOf := reflect.TypeOf(value)
		if typeOf.PkgPath() != legacyPath {
			t.Fatalf("%s package path = %q, want %q", typeOf.Name(), typeOf.PkgPath(), legacyPath)
		}
	}

	value := mustEuro(t, "12.34")
	legacy, err := gomoney.New(map[string]money.Money{"box": value})
	if err != nil {
		t.Fatal(err)
	}
	successor, err := moneyobjective.New(map[string]money.Money{"box": value})
	if err != nil {
		t.Fatal(err)
	}
	plan := mustPlan(t, "box", "box")
	legacyTotal, legacyErr := legacy.Total(plan)
	successorTotal, successorErr := successor.Total(plan)
	if legacyErr != nil {
		t.Fatalf("legacy total: %v", legacyErr)
	}
	if successorErr != nil {
		t.Fatalf("successor total: %v", successorErr)
	}
	comparison, compareErr := legacyTotal.Compare(successorTotal)
	if compareErr != nil || comparison != 0 {
		t.Fatalf("total parity: legacy=(%v, %v) successor=(%v, %v)", legacyTotal, legacyErr, successorTotal, successorErr)
	}
}
