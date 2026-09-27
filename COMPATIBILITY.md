# Compatibility Policy

The current root module uses `github.com/faustbrian/go-knapsack/v2` and plain
major-version tags such as `v2.0.0` from main. The previously published
`github.com/faustbrian/go-knapsack` v1 module remains an immutable
legacy consumer dependency; it is not the current root source. The v2 root
uses `github.com/faustbrian/go-measurement/v2` quantities, so callers update
both imports together. Canonical request and plan schema v1 is retained.

The canonical optional
`objective/money/v2` module and deprecated `objective/gomoney` v1 compatibility
facade are released independently using `objective/money/v<version>` and
`objective/gomoney/v<version>` tags, respectively.

The facade remains supported throughout its v1 line and delegates exact-money
behavior to published `objective/money` v1. Within that legacy type cohort,
migrate by changing the module and import path to `objective/money` and using
its `moneyobjective` package identifier. Current root v2 types instead require
`objective/money/v2`, whose adoption awaits its own public artifacts. The
facade must not depend on that new major before canonical publication. The
facade will remain available for the longer of 180 days and two published
stable minor releases after `objective/money` became public. Removal also
requires clean external-consumer evidence and an authorized
`objective/gomoney` v2.0.0 release.

Before `v1`, minor releases MAY contain reviewed breaking changes, but every
break MUST be documented with migration guidance. Patch releases MUST remain
backward compatible. At and after `v1`, incompatible exported API or documented
behavior changes require a new major version.

Compatibility includes exported Go APIs, error classification, serialization,
protocol behavior, persistence schemas, environment variables, command output,
resource ownership, ordering, retry/idempotency semantics, and documented
defaults. A compile-compatible change can still be behaviorally breaking.

Specification-backed modules MUST NOT diverge from their declared standards.
Ambiguities require documented decisions and stable tests. Deprecated APIs
follow [`DEPRECATION.md`](DEPRECATION.md).
