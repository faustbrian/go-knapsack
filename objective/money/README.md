# Exact-money knapsack objective

`moneyobjective` is the canonical optional adapter for minimizing exact
container costs with the Knapsack solvers. Its target-oriented import path
keeps the Money dependency outside the Knapsack core.

The independently published v1.0.0 module is stable and requires Go 1.26.6.
Current unreleased source requires Go 1.27.0.

## Install

```sh
go get github.com/faustbrian/go-knapsack/objective/money@v1
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
`money` identifier used for `github.com/faustbrian/go-money` without an import
alias.

## Guarantees and limitations

The [complete guide](docs/reference.md) defines ownership, failure semantics,
bounds, concurrency, security, and unsupported behavior. Do not infer
additional guarantees beyond the documented module boundary.

## Documentation

- [Documentation index](docs/README.md)
- [Complete technical guide](docs/reference.md)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-knapsack/objective/money)
- [Parent package documentation](https://github.com/faustbrian/go-knapsack/blob/objective/money/v1.0.0/docs/README.md)
- [Versioned Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/README.md)
- [Domain utilities family guidance](https://github.com/faustbrian/go-library-tools/blob/v1.6.1/docs/ecosystem/design-language.md#package-families-and-selection)

## Compatibility and support

This stable v1 module follows Semantic Versioning. Use the
[parent support policy](https://github.com/faustbrian/go-knapsack/blob/objective/money/v1.0.0/SUPPORT.md) for adoption and defect reports, and
report vulnerabilities through the [parent security policy](https://github.com/faustbrian/go-knapsack/blob/objective/money/v1.0.0/SECURITY.md).
The released `objective/gomoney` module remains available as a compatibility
path; new adoption should use this module.

## License

MIT. See [LICENSE](LICENSE).
