# Exact-money knapsack objective

`gomoney` is the deprecated compatibility facade for the released exact-money
objective path. It preserves the existing public API while delegating behavior
to the canonical
[`objective/money/v2`](https://pkg.go.dev/github.com/faustbrian/go-knapsack/objective/money/v2)
module.

The deprecated `/v2` facade is published as v2.0.2 for Knapsack v2 and
Measurement v2 with Money v1 and requires Go 1.27.0. Published v1 releases
remain supported with their legacy packing types. New code should use
`objective/money/v2` for Money v1 or `objective/money/v3` for Money v2.

## Install

```sh
go get github.com/faustbrian/go-knapsack/objective/gomoney/v2@v2.0.2
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

The corrected Exact contracts require root Knapsack v2.0.2, not merely this
adapter installation. The adapter's existing root dependency minimum remains
v2.0.0. After the root v2.0.2 tag and module artifacts are published, select it
explicitly alongside this adapter:

```sh
go get github.com/faustbrian/go-knapsack/v2@v2.0.2
```

For conclusive `Exact.PackAll` optimization, give the request finite container
stock. Costs price every selected instance, including unused ones; negative
costs can reward unlimited extra bins, and opaque custom objectives do not
provide a solver-visible monotonicity proof. Unlimited-stock searches may return
a resource error with a retained feasible plan, which must be inspected together.

The compiling examples in this module contain complete imports and setup.

## Guarantees and limitations

The [complete guide](docs/reference.md) defines ownership, failure semantics,
bounds, concurrency, security, and unsupported behavior. Do not infer
additional guarantees beyond the documented module boundary.

## Documentation

- [Documentation index](docs/README.md)
- [Complete technical guide](docs/reference.md)
- [Go API reference](https://pkg.go.dev/github.com/faustbrian/go-knapsack/objective/gomoney/v2)
- [Parent package documentation](https://github.com/faustbrian/go-knapsack/blob/v2.0.0/docs/README.md)
- [Versioned Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.5.5/docs/ecosystem/README.md)
- [Domain utilities family guidance](https://github.com/faustbrian/go-library-tools/blob/v1.5.5/docs/ecosystem/design-language.md#package-families-and-selection)

## Compatibility and support

The facade retains its own defined types and directly shares canonical error
sentinels. The v2 facade uses v2 packing types; published v1 remains supported
throughout v1. Within the v2 packing and Money v1 cohort, migrate by
changing the module and import path to
`objective/money/v2` and using its `moneyobjective` package identifier. The
constructors, methods, sentinel identities, errors, score components, and
solver behavior remain compatible. The facade will remain available for the
longer of 180 days and two published stable minor releases after
`objective/money` became public. Removal also requires clean external-consumer
evidence and an authorized v2.0.0 release.

Use the
[parent support policy](https://github.com/faustbrian/go-knapsack/blob/v2.0.0/SUPPORT.md) for adoption and defect reports, and
report vulnerabilities through the [parent security policy](https://github.com/faustbrian/go-knapsack/blob/v2.0.0/SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).

Choose the next guide in the [documentation index](../../docs/README.md)
or return to the [repository entry](../../README.md) for installation and ownership.
