package moneyobjective_test

import (
	"fmt"

	"github.com/faustbrian/go-international/currency"
	moneyobjective "github.com/faustbrian/go-knapsack/objective/money"
	"github.com/faustbrian/go-money"
)

func ExampleNew() {
	euro, _ := currency.Parse("EUR")
	moneyContext, _ := money.DefaultContext(euro)
	small, _ := money.Parse("0.60", euro, moneyContext)
	large, _ := money.Parse("1.50", euro, moneyContext)
	costs, _ := moneyobjective.New(map[string]money.Money{
		"small": small,
		"large": large,
	})

	total, _ := costs.Total(mustPlanForExample("small", "small"))
	fmt.Println(total)

	// Output:
	// 1.20 EUR
}

func ExampleNewWithPolicy() {
	euro, _ := currency.Parse("EUR")
	moneyContext, _ := money.DefaultContext(euro)
	credit, _ := money.Parse("-0.25", euro, moneyContext)
	policy := moneyobjective.DefaultPolicy()
	policy.AllowNegativeCosts = true
	costs, _ := moneyobjective.NewWithPolicy(
		map[string]money.Money{"reusable": credit},
		policy,
	)

	total, _ := costs.Total(mustPlanForExample("reusable"))
	fmt.Println(total)

	// Output:
	// -0.25 EUR
}
