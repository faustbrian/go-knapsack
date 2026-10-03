# Exact-money knapsack objective

`moneyobjective` is the canonical optional adapter for minimizing exact
container costs with the Knapsack solvers. Its target-oriented import path
keeps the Money dependency outside the Knapsack core.

The current source prepares the v3 adapter for published Money v2.0.0 values,
retaining Knapsack and measurement v2 types, and requires Go 1.27.0. This v3
adapter is not yet published. Installation below requires its public
`objective/money/v3.0.0` tag and module artifacts. Until then, consumers retain
their published canonical v2 or legacy v1 adapter dependencies.

## Install

```sh
go get github.com/faustbrian/go-knapsack/objective/money/v3@v3.0.0
```

## Quick start

```go
euro, _ := currency.Parse("EUR")
moneyContext, _ := money.DefaultContext(euro)
small, _ := money.Parse("0.60", euro, moneyContext)
large, _ := money.Parse("1.50", euro, moneyContext)

costs, err := moneyobjective.New(map[string]money.Money{
    "small": small,
    "large": large,
})
if err != nil {
    // Configuration is invalid.
}

plan, err := (solver.Exact{}).PackAll(ctx, request, solver.Options{
    PlanObjective: costs,
})
```

The compiling examples in this module contain complete imports and setup.

The package declaration is `moneyobjective`, so it remains distinct from the
`money` identifier used for `github.com/faustbrian/go-money/v2` without an import
alias.

## Guarantees and limitations

The [complete guide](docs/reference.md) defines ownership, failure semantics,
bounds, concurrency, security, and unsupported behavior. Do not infer
additional guarantees beyond the documented module boundary.

## Documentation

- [Documentation index](docs/README.md)
- [Complete technical guide](docs/reference.md)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-knapsack/objective/money/v3)
- [Parent package documentation](https://github.com/faustbrian/go-knapsack/blob/v2.0.0/docs/README.md)
- [Versioned Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/README.md)
- [Domain utilities family guidance](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/design-language.md#package-families-and-selection)

## Compatibility and support

This v3 line follows Semantic Versioning. Use the
[parent support policy](https://github.com/faustbrian/go-knapsack/blob/v2.0.0/SUPPORT.md) for adoption and defect reports, and
report vulnerabilities through the [parent security policy](https://github.com/faustbrian/go-knapsack/blob/v2.0.0/SECURITY.md).
The released `objective/gomoney` v1 facade remains supported with v1 packing
types and its published canonical v1 dependency. It is not interchangeable with
this adapter. The deprecated `objective/gomoney/v2` facade retains canonical
`objective/money/v2` v2.0.0 and Money v1 values; it does not adopt or alias v3.
New Money v2 adoption should use this module once published.

## License

MIT. See [LICENSE](LICENSE).
