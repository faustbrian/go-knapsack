# Migration

[Documentation index](README.md) · [Repository entry](../README.md)

## Root module v2

Use `github.com/faustbrian/go-knapsack/v2` and
`github.com/faustbrian/go-measurement/v2` together. The physical input fields
and quantity-returning methods accept measurement v2 types, not v1 types.
No numeric conversion or canonical schema migration is required: exact lattice
normalization and request/plan schema v1 are retained. Reject errors from
measurement quantity construction before constructing packing inputs.

Root v2.0.0 and measurement v2.0.2 are published dependencies of the current
`integration/references` harness. Its module remains an internal harness, not
an independently released public adapter. The comparison schema and shared
packing subset are unchanged.

The published canonical adapter
`github.com/faustbrian/go-knapsack/objective/money/v3` v3.0.1 accepts Money
`github.com/faustbrian/go-money/v2` v2.0.0, retaining root and Measurement v2
packing types. Change constructor and
`Entry.Cost` inputs and `Costs.Total` outputs to Money v2 together; their Go
identities differ from Money v1 without changing exact totals or formats.

The canonical adapter's public `objective/money/v2.0.0` module remains available
with Money v1. The published deprecated facade `objective/gomoney/v2`
v2.0.2 retains that selection, not canonical v3. Update root and Measurement
imports to `/v2` together. Retain supported v1 adapter releases
and the facade's existing removal interval; root publication alone neither
migrates nor removes them. No storage or canonical schema migration is implied.

## Exact proof corrections in root v2.0.2

After its public tag and module artifacts are available, explicitly select
`github.com/faustbrian/go-knapsack/v2@v2.0.2` alongside the existing money
adapter. Canonical money v3.0.1 and facade v2.0.2 still declare a root v2.0.0
minimum; installing either adapter alone does not select the corrected solver.
Their Money v2 and Money v1 identities and prices remain unchanged.

Provide finite stock for conclusive optimization of opaque whole-plan
objectives. Unlimited domains return resource interruption, retaining feasible
state when available; grid admission does not establish infeasibility. An unrepresentable
configuration can leave a verified `feasible` alternative plus `ErrOverflow`,
not an optimum. Inspect the plan and error together. No canonical storage or
quantity conversion is required, but corrected Exact placements/proof statuses
can differ from earlier results.

## Other packing libraries

From BoxPacker, `gopackx`, or `bp3d`, first normalize every dimension and mass
through `measurement`; do not copy numeric fields as if units already match.
Expand quantities into stable item instance IDs, map rotation policy to
physical-axis orientations, declare finite stock explicitly, and choose an
ordered objective instead of relying on implicit box order.

Treat old solver output as untrusted: translate it into a `knapsack.Plan` and
run `verify.Plan`. A heuristic failure is `best_known` or a work-limit result,
never proven infeasibility. Comparative tests must remove unsupported semantics
from both sides and compare solution validity and quality before runtime.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
