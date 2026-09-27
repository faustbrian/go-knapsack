# Compatibility Policy

The current root module uses `github.com/faustbrian/go-knapsack/v2` and plain
major-version tags such as `v2.0.0` from main. The previously published
`github.com/faustbrian/go-knapsack` v1 module remains an immutable
legacy consumer dependency; it is not the current root source. The v2 root
uses `github.com/faustbrian/go-measurement/v2` quantities, so callers update
both imports together. Canonical request and plan schema v1 is retained.

The canonical optional
`objective/money/v2` module and deprecated `objective/gomoney/v2` compatibility
facade are released independently using `objective/money/v<version>` and
`objective/gomoney/v<version>` tags, respectively.

Published facade v1 remains supported throughout its v1 line and delegates
exact-money behavior to published `objective/money` v1. Within that legacy type cohort,
migrate by changing the module and import path to `objective/money` and using
its `moneyobjective` package identifier. Current root v2 types instead require
the published `objective/money/v2`. Current facade source prepares
`objective/gomoney/v2` against that canonical dependency; adopt the facade
once its own public tag and artifacts exist. Its defined types remain
facade-owned and its sentinel errors directly share canonical identity. The
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
