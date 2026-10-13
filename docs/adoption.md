# Adoption guide

[Documentation index](README.md) · [Repository entry](../README.md)

1. Inventory every length, mass, rotation rule, stock source, and packing
   objective. Do not infer units from bare numbers.
2. Choose a lattice that exactly represents accepted business quantities.
   Define any application rounding before calling this module.
3. Give every expanded item a stable instance ID and every box a stable type
   ID. Preserve those IDs through order and warehouse systems.
4. Start with `Heuristic.PackAll`, conservative limits, and independent
   verification. Record unpacked diagnostics rather than dropping items.
5. Add physical rules with fixtures for fragile, upright, support, load,
   linked, and incompatible cases.
6. Use the exact solver as a tiny-case oracle and compare solution quality on
   representative normalized orders.
7. Pin the module version, canonical fixtures, options, and seed before
   accepting output changes during upgrades.

Adoption is complete only when monitoring distinguishes invalid input, no
heuristic placement, resource exhaustion, and proven infeasibility. Do not
replace a production packer from runtime-only benchmarks; compare feasibility
and objective quality on identical semantics.

## Corrected Exact search

The corrected Exact solver searches every admitted integer-lattice origin rather
than only compact face coordinates. This restores feasible-placement and proof
contracts, but can require more work and allocation on the same input. The tiny
fixed fixture now visits 82 nodes and attempts 171 placements instead of 16 and
123. Component-only scoring avoids unused canonical serialization; public full
scores and canonical comparison ties are unchanged.

Re-evaluate representative orders and choose the existing node, branch,
candidate and memory limits for the application's resource envelope. Resource
exhaustion is not proven infeasibility; retain verified partial output and its
termination classification where the API provides it. No caller repin, new
option or API migration is required by the scoring optimization itself.
[Benchmark calibration](benchmarks.md#complete-search-calibration-2026-10-13)
is fixture- and environment-specific, not a universal performance guarantee.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
