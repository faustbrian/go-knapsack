# Changelog

All notable changes to this module are documented here.

## Unreleased

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
