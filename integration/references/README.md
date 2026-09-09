# Reference integration harness

This module is an internal, non-releasable engineering harness. It is not a
public package or supported application dependency. It compares Knapsack with
`gopackx` and `bp3d` only across their shared integral-cuboid, unrestricted
rotation, weight-capacity, and pack-all behavior.

The reference libraries provide comparison evidence; Knapsack's independent
`verify` package remains authoritative for placement feasibility. Differences
outside the documented common subset are not compatibility failures.

Run the comparison tests independently with Go 1.26.6:

```sh
GOWORK=off go test ./...
```

The `cmd/knapsack-compare` executable emits the owned comparison-adapter schema
used by repository evidence tooling. See the parent
[capability assessment](https://github.com/faustbrian/go-knapsack/blob/v1.0.0/docs/capabilities.md) and
[benchmark methodology](https://github.com/faustbrian/go-knapsack/blob/v1.0.0/docs/benchmarks.md) for scope and interpretation.
