# Changelog

All notable changes to this module are documented here.

## Unreleased

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
