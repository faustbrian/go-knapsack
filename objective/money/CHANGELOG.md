# Changelog

All notable changes to this module are documented here.

## Unreleased

### Changed

- Clarify finite-stock adoption for conclusive Exact custom-cost optimization;
  unlimited objectives may return resource interruption with retained feasible
  state. Keep credits, unused-instance pricing and Money v2 semantics unchanged.

- Clarify whole-plan candidate pricing, fixed-context currency requirements and
  exact final-sum bounds without changing objective behavior.

## [3.0.1] - 2026-10-11

### Changed

- Select Math v1.1.3 and Measurement v2.0.2 for standalone consumers while
  retaining Money v2 and Knapsack v2 type identities and exact objectives.
- The earlier v3.0.0 preparation note below describes the already-published
  major migration, not a new migration required by this patch.

### Fixed

- Correct adoption guidance for the published v3.0.0 adapter. Publication now
  satisfies the condition in the earlier preparation note below; Money v2
  callers can adopt v3 while Money v1 callers retain canonical v2.

### Changed

- Prepare canonical adapter `/v3` for published Money `/v2` v2.0.0 values,
  retaining Knapsack v2 and measurement v2 identities, exact costs, errors,
  ownership, and persisted formats. Adopt only after `objective/money/v3.0.0`
  is published; the deprecated facade retains canonical v2 and Money v1.

## 2.0.1 - 2026-10-02

### Security

- Select go-math v1.1.2 so standalone users retain its checked-exponent
  and allocation-budget protections.

## 2.0.0 - 2026-09-27

### Changed

- Move the canonical adapter to `/v2` to accept published Knapsack v2 and
  measurement v2 type identities. Physical source remains in `objective/money`.
- Preserve exact totals, fixed-currency validation, immutable input ownership,
  errors, deterministic ties, score components, and persisted formats.
- Retain the immutable v1 API baseline and supported legacy facade. Adopt this
  adapter only after its public major tag and module artifacts are available.

This entry dates source preparation, not publication.


## 1.0.0 - 2026-09-09

### Added

- Publish `objective/money` as the canonical target-oriented exact-money
  objective with the released `gomoney` behavior, bounds, errors, and solver
  integration.
- Preserve immutable input ownership, deterministic ties, fixed-context
  validation, exact aggregation, and cancellation precedence.

### Compatibility

- Existing consumers may remain on `objective/gomoney`; migrate new imports to
  `objective/money` and use its natural `moneyobjective` package identifier.
