# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected package's docs.
2. Run `GOWORK=off go test ./...` or the affected package's narrower tests.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected stable decision entries, and complete the Specification Decisions
section of the pull request template. An unresolved interpretation or stale
source pin is release-blocking; peer behavior cannot silently select policy.

Mutation is optional local or scheduled assurance, not a required PR status.
Confirmed behavioral or security defects still require a focused regression
and repair; retain reviewed narrow equivalent/invalid-mutant dispositions.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Standalone `GOWORK=off go test ./...` runs the root module's deterministic tests,
fuzz seeds and saved corpus regressions without sibling repositories or
ecosystem ledgers. The checked-in workspace contains only owned nested modules;
it is optional for an isolated root change. To check an affected nested module,
run `GOWORK=off go test ./...` from its directory. Every declared module has no
required services, build tags or test tags.

With the checksum-pinned public verifier selected by `.golib.yaml`, use:

```bash
make inventory
make check
make ci
```

`check` and `ci` select bounded local verification for all declared modules.
One existing solver race assertion checks that both solvers return cancellation
and do not retain goroutines, with its original 32 rounds per solver, once with
a two-minute timeout. The full package operation retains its original five
repetitions, corpus, BoxPacker and dependency-license checks. The separate hosted
32-bit scalar allocation-admission job remains Required.
The full BoxPacker operation installs its pinned Composer dependencies and
requires Composer/PHP; ordinary Go tests leave that opt-in integration disabled
unless `BOXPACKER_INTEGRATION=1` is selected.

Select one target for short active fuzz exploration. Optional local mutation
selects one package; neither command is an all-change prerequisite:

```bash
GOWORK=off GOMAXPROCS=2 go test -run '^$' -fuzz '^FuzzItemContainerValidation$' -fuzztime=2s -parallel=1 -timeout=2m .
golib mutation --module . --package ./solver
```

Packing, cancellation, allocation, canonical encoding, objectives and public API
changes retain affected oracle, race, architecture, compatibility and direct
consumer checks. Changes reaching the money adapters or reference harness must
check those owned modules; their public dependency contracts stay independent.
Preserve reviewed fixtures and exact fuzz/scenario budgets.

`make check-full` preserves all full operations, including the complete native
fuzz mappings for the root and both objective adapters and comparative benchmark
contracts. Relevant release rehearsal remains `golib release dry-run`; see
[release policy](docs/release.md). Select full/release checks for a named risk or
release boundary, not every commit. Report unavailable, skipped and failed
boundaries honestly; a local pass is not hosted-CI, performance, optimality or
public-release proof.

## Selected Monthly Assurance

[`assurance.yml`](.github/workflows/assurance.yml) runs on main at 07:11 UTC on
the 21st day of each month. It selects one of the 15 declared root-module fuzz
targets using the shared UTC-week selector; it does not run every target or
promise a complete monthly rotation. Explicit main-only dispatch accepts one
exact declared selector, for example `^FuzzCancellationAndBudgets$`.

The selected risks span request/plan decoding, geometry, objective comparison,
packing validity and bounded solver cancellation. Existing targets include
canonical replay, independent plan verification and tiny exact-solver agreement.
This is not general optimality, performance, leak or BoxPacker interoperability
proof; retain the solver race, 32-bit admission check, reference harness and
separate objective-adapter module contracts.

The immutable shared runner bounds active fuzzing to 30 seconds, two workers
and two CPUs, with a five-minute combined native session, 2 GiB memory and
1 MiB output limits inside a ten-minute job. It installs the declared public
CLI without a source build, uses private disposable caches and keeps failures,
timeouts and exhausted budgets visible. Ordinary deterministic tests still run
seeds/corpus; ordinary CI and its Required contract are unchanged. Full/manual
operations and their original budgets remain in `.golib.yaml`.

Treat findings as owner-attributed diagnostics: promptly repair viable defects
with focused regressions; retain reviewed narrow equivalent/invalid dispositions
rather than weakening tests. Review cadence against useful findings and actual
cost; do not interpret a selected pass as full or release qualification.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
