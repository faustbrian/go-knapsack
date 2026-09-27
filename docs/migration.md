# Migration

## Root module v2

Use `github.com/faustbrian/go-knapsack/v2` and
`github.com/faustbrian/go-measurement/v2` together. The physical input fields
and quantity-returning methods accept measurement v2 types, not v1 types.
No numeric conversion or canonical schema migration is required: exact lattice
normalization and request/plan schema v1 are retained. Reject errors from
measurement quantity construction before constructing packing inputs.

Independently versioned objective adapters and the reference harness retain
their published v1 dependency sets until their separate v2 migrations.

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
