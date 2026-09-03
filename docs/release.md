# Release and compatibility policy

The initial minimum language and toolchain is Go 1.26, controlled by the
repository-wide version files. Public types, typed errors, canonical encoding,
objective semantics, coordinates, and proof statuses are contracts.

Every change must pass `make check`, which delegates to `golib check --all`.
A release candidate must also pass `golib release dry-run`; it proves the
release archive and clean-consumer path before running the same complete module
contract. That contract includes mutation execution, benchmark budgets, SBOM
generation, secret scanning, corpus provenance, dependency and vulnerability
review, documentation, examples, race detection, fuzzing, and meaningful 100%
production statement coverage. NilAway remains advisory and visible within the
shared module contract.

The package-specific test operation repeatedly cancels both solvers under the
race detector, proves production packages contain no unmanaged goroutine
launches, verifies the reference corpus, exercises the BoxPacker common subset,
and validates the pinned dependency-license manifest. Fuzzing uses the reviewed
exact execution counts from `verification/fuzz-budgets.tsv` without local or
CI multipliers. The consumer workflow and shared reusable workflow are pinned
to immutable commits; repository checks reject mutable or undocumented pins.

Source archives pin the repository commit being released. The shared release
rehearsal rejects tag collisions, builds the declared modules in a task-owned
proxy, and resolves each as a clean external consumer. Repository policy rejects
permanent replacements and placeholder release metadata.

The typed package test compares every compiled non-standard module with
`verification/dependency-licenses.tsv`, pins each license hash and SPDX
classification, and fails for missing, changed, local-only, or placeholder
dependencies.

Performance or quality claims require raw evidence with the machine, Go
version, execution revision, complete input fingerprint, seed, fixture hash,
constraints, objective,
verification, work, and allocations. Unsupported reference behavior must be
removed from both sides or reported separately.

Canonical encoding changes require a new version and migration documentation.
A released decoder must retain its current and immediately previous schema.
Initial v1 has no predecessor; its hashed request and plan fixtures establish
the N-1 contract that the first schema transition must continue to decode.
A heuristic improvement may change placements while preserving semantics;
consumers requiring identical bytes must pin module version, normalized
request, options, and seed.
