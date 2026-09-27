# Release and compatibility policy

The minimum language and toolchain is Go 1.27.0, controlled by the
repository-wide version files. Public types, typed errors, canonical encoding,
objective semantics, coordinates, and proof statuses are contracts.

The root v2.0.0 release uses a `v2.0.0` tag from main and the `/v2` module
path, without a version-specific source directory. Eligibility and a dated
changelog are not publication proof. Publish and verify the root artifacts and
a clean public consumer before migrating the independently versioned adapters
as described in [migration](migration.md). The existing v1 adapters keep their
own manifests and release histories until that separate migration.

Verification follows the repository's proportional assurance policy. The v2
root migration requires current packing and hostile-boundary tests, API review,
public dependency integrity, exact-source CI and a real clean public consumer
at publication. Run race, fuzz, mutation and performance gates when they
exercise an affected material risk or the applicable release claim, not solely
because a dated historical receipt is present. Required selected checks fail
closed; NilAway remains advisory.

The v1 mutation and benchmark receipts retain their original source and
dependency identities. They are historical records, not v2 execution evidence;
no new performance or mutation result is claimed by the import migration.

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
