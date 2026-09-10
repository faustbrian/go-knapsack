# Exact-money knapsack objective

`gomoney` is the deprecated compatibility facade for the released exact-money
objective path. It preserves the existing public API while delegating behavior
to the canonical [`objective/money`](../money/) module.

The independently published v1.1.0 module remains stable for compatibility and
requires Go 1.26.6. New code should use `objective/money`.

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
- [Parent package documentation](https://github.com/faustbrian/go-knapsack/blob/objective/gomoney/v1.1.0/docs/README.md)
- [Versioned Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.5.5/docs/ecosystem/README.md)
- [Domain utilities family guidance](https://github.com/faustbrian/go-library-tools/blob/v1.5.5/docs/ecosystem/design-language.md#package-families-and-selection)

## Compatibility and support

This deprecated stable-v1 facade follows Semantic Versioning and remains
supported throughout v1. Migrate by changing the module and import path to
`objective/money` and using its `moneyobjective` package identifier. The
constructors, methods, sentinel identities, errors, score components, and
solver behavior remain compatible. The facade will remain available for the
longer of 180 days and two published stable minor releases after
`objective/money` became public. Removal also requires clean external-consumer
evidence and an authorized v2.0.0 release.

Use the
[parent support policy](https://github.com/faustbrian/go-knapsack/blob/objective/gomoney/v1.1.0/SUPPORT.md) for adoption and defect reports, and
report vulnerabilities through the [parent security policy](https://github.com/faustbrian/go-knapsack/blob/objective/gomoney/v1.1.0/SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
