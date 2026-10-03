# FAQ and troubleshooting

## Why was an exact-looking quantity rejected?

It is not an integer multiple of the selected lattice after exact conversion,
or its physical dimension differs from the resolution. Inspect
`ErrInexactResolution` and the `FieldError` field.

## Why is a touching item not overlapping?

Cuboids use half-open intervals. Face and edge contact is allowed; only
positive-volume intersection is overlap.

## Why did the heuristic return `best_known`?

At least one item remained explicitly unpacked, but infeasibility was not
proved. Check diagnostics, stock, orientations, eligibility, support, load,
groups, and candidate limits.

## Why did search return a plan and an error?

Cancellation or a work budget can stop search after a verified plan was found.
The plan is the safe best retained result; the error and termination explain
why the requested search or proof did not finish.

## How is monetary cost represented?

Use `objective/money/v3` for Money v2 with root v2 after the adapter's own
v3.0.0 publication; current v3 source is not yet published. Published canonical
v2.0.0 accepts Money v1. Neither adds Money to the root module.
`objective/gomoney` v1 remains supported for existing v1 consumers; current
`objective/gomoney/v2` source targets v2 and awaits its own public artifacts. Container count
and cost remain separate lexicographic criteria.

## How do I debug an external plan?

Decode it with `encoding.UnmarshalPlan`, then call `verify.Plan`. Use stable
violation codes rather than parsing messages. Rendering is available only
after verification and is not proof.

## Does the exact solver scale to warehouse batches?

No general guarantee exists. Three-dimensional bin packing is NP-hard. Use the
heuristic for ordinary orders and the exact solver only within measured small
instance limits.
