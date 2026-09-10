# Compatibility Policy

The root module is released as `github.com/faustbrian/go-knapsack` and uses
plain semantic-version tags such as `v1.0.0`. The canonical optional
`objective/money` module and deprecated `objective/gomoney` compatibility
facade are released independently using `objective/money/v<version>` and
`objective/gomoney/v<version>` tags, respectively.

The facade remains supported throughout its v1 line and delegates exact-money
behavior to `objective/money`. Migrate by changing the module and import path
to `objective/money` and using its `moneyobjective` package identifier. The
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
