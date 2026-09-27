# Migration

## Root module v2

Use `github.com/faustbrian/go-knapsack/v2` and
`github.com/faustbrian/go-measurement/v2` together. The physical input fields
and quantity-returning methods accept measurement v2 types, not v1 types.
No numeric conversion or canonical schema migration is required: exact lattice
normalization and request/plan schema v1 are retained. Reject errors from
measurement quantity construction before constructing packing inputs.

Root v2.0.0 and measurement v2.0.0 are published dependencies of the current
`integration/references` harness. Its module remains an internal harness, not
an independently released public adapter. The comparison schema and shared
packing subset are unchanged.

The current canonical adapter source uses
`github.com/faustbrian/go-knapsack/objective/money/v2` for root v2 packing types.
The canonical adapter's public `objective/money/v2.0.0` module is available.
Current deprecated facade source adopts it through `objective/gomoney/v2`;
adopt that facade only after its own `objective/gomoney/v2.0.0` tag and public
module artifacts exist. Update root and Measurement imports to `/v2` together. Retain supported v1 adapter releases
and the facade's existing removal interval; root publication alone neither
migrates nor removes them. No storage or canonical schema migration is implied.

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
