# Changelog

All notable changes to this module are documented here.

## Unreleased

### Changed

- Keep shared fuzz, mutation, dependency, and aggregate evidence under the
  repository's verification tree without changing objective behavior.

- Adopt the repository's schema-v2 cohesion contract for the independently
  releasable exact-money objective without changing its API or behavior.

- Adopt the repository's shared verification contract while retaining the
  module's API baseline, exact fuzz budgets, benchmarks, and mutation evidence.

### Documentation

- Publish selection, ownership, lifecycle, and support metadata and link to the
  immutable v1.4.0 ecosystem guidance.

- Move detailed module guidance behind a concise README and documentation index.
## 1.0.0 - 2026-08-25

### Documentation

- Link the package README to package-owned documentation.

### Added

- Add explicit negative-cost policy, duplicate-aware entry construction, exact
  empty-plan totals, and module adoption and API documentation.
- Add allocation-aware total and comparison benchmarks plus bounded fuzz and
  property coverage for exact mappings and hostile identifiers.
- Add cross-process ranking goldens, concurrent solver reuse, solver-callback
  fuzzing, direct lookup benchmarks, and an ambient-dependency policy audit.

### Fixed

- Reject unsupported money contexts and retain specific validation and
  arithmetic error identities through objective evaluation. Mixed-sign totals
  no longer depend on container iteration order near amount bounds.
- Reject oversized type identifiers before any trimming, comparison, or sorting
  work and prove ranking independence from solver candidate search order.

### Distribution

- Include the canonical MIT licence in the independently published module.

### Compatibility

- Added a pinned module export baseline so incompatible public API changes
  fail the canonical repository gate.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-knapsack/objective/gomoney` identity while preserving its documented API and behavior.
- Refresh local `v0.0.0` owned-module checksums after dependency manifests and
  release notes were normalized; runtime behavior and public APIs are
  unchanged.
- Require owned sibling modules at local `v0.0.0`; clean external consumers
  pin each module to an exact main pseudo-version.

- Pin unpublished Money and owned transitive modules to exact resolvable
  revisions so clean consumers do not depend on missing `v0.1.0` tags.
- Refresh owned-module checksums against the final consolidated archives.
- Refresh the parent Knapsack checksum used by clean consumer builds.
- Normalized standalone module metadata against the canonical owned dependency
  graph, including complete checksums for clean consumer resolution.
