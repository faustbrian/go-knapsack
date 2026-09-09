# Exact-money knapsack objective

`gomoney` is an optional adapter that lets the Knapsack solvers minimize exact
container costs. It depends on the owned Money module so the Knapsack core can
remain independent of monetary policy.

The independently published v1.0.0 module is stable and requires Go 1.26.6.

## Install

```sh
go get github.com/faustbrian/go-knapsack/objective/gomoney@v1
```

## Quick start

```go
euro, _ := currency.Parse("EUR")
moneyContext, _ := money.DefaultContext(euro)
small, _ := money.Parse("0.60", euro, moneyContext)
large, _ := money.Parse("1.50", euro, moneyContext)

costs, err := gomoney.New(map[string]money.Money{
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

## Guarantees and limitations

The [complete guide](docs/reference.md) defines ownership, failure semantics,
bounds, concurrency, security, and unsupported behavior. Do not infer
additional guarantees beyond the documented module boundary.

## Documentation

- [Documentation index](docs/README.md)
- [Complete technical guide](docs/reference.md)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-knapsack/objective/gomoney)
- [Parent package documentation](../../docs/README.md)
- [Versioned Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.5.5/docs/ecosystem/README.md)
- [Domain utilities family guidance](https://github.com/faustbrian/go-library-tools/blob/v1.5.5/docs/ecosystem/design-language.md#package-families-and-selection)

## Compatibility and support

This stable v1 module follows Semantic Versioning. Use the
[parent support policy](../../SUPPORT.md) for adoption and defect reports, and
report vulnerabilities through the [parent security policy](../../SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
