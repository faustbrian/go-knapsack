# Package map

[Documentation index](README.md) · [Repository entry](../README.md)

This map follows the current source modules, not a synchronized release.
Use the repository README for published installation versions; an unpublished
source major is not installable merely because its package appears here.
Internal packages and non-releasable fixtures are intentionally excluded.

## github.com/faustbrian/go-knapsack/v2

`knapsack` is an exact-unit, deterministic library for offline orthogonal packing of rectangular items into rectangular containers. The sometimes-used term “4D bin packing” means three geometric dimensions plus scalar weight capacity; weight is not a fourth geometric axis. This differs from classical knapsack selection, which maximizes value while intentionally omitting items.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-knapsack/v2`](../doc.go) | Use for immutable packing requests, exact lattice normalization, bounded resource policy, and canonical plan values. |
| [`github.com/faustbrian/go-knapsack/v2/constraint`](../constraint/constraint.go) | Use to define trusted synchronous placement constraints over immutable bounded views. |
| [`github.com/faustbrian/go-knapsack/v2/encoding`](../encoding/encoding.go) | Use for strict bounded canonical JSON request and plan interchange. |
| [`github.com/faustbrian/go-knapsack/v2/geometry`](../geometry/geometry.go) | Use for checked integer cuboids, physical-axis orientations, and half-open spatial predicates. |
| [`github.com/faustbrian/go-knapsack/v2/knapsacktest`](../knapsacktest/assertions.go) | Use in tests for canonical plan comparison and independent verification assertions. |
| [`github.com/faustbrian/go-knapsack/v2/objective`](../objective/objective.go) | Use to define exact deterministic plan objectives and lexicographic criteria. |
| [`github.com/faustbrian/go-knapsack/v2/solver`](../solver/exact.go) | Use for bounded deterministic heuristic packing or exhaustive small-instance proofs. |
| [`github.com/faustbrian/go-knapsack/v2/verify`](../verify/verify.go) | Use to validate supplied plans independently from solver placement predicates. |
| [`github.com/faustbrian/go-knapsack/v2/visualize`](../visualize/svg.go) | Use to render deterministic human-readable views of canonical plans. |

## github.com/faustbrian/go-knapsack/objective/gomoney/v2

`gomoney` is the deprecated compatibility facade for the released exact-money objective path. It preserves the public API while delegating behavior to the canonical `objective/money` module.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-knapsack/objective/gomoney/v2`](../objective/gomoney/gomoney.go) | Deprecated facade retaining canonical objective/money/v2 and Money v1; new Money v2 code uses published objective/money/v3 v3.0.0. |

## github.com/faustbrian/go-knapsack/objective/money/v3

`moneyobjective` is the canonical optional adapter that lets Knapsack solvers minimize exact container costs without adding monetary policy to the root module.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-knapsack/objective/money/v3`](../objective/money/moneyobjective.go) | Use published v3.0.0 when Knapsack container costs already exist as exact go-money/v2 values in one fixed currency context. |

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
