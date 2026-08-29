# Resource and trust model

Validate before expansion or allocation. Keep counts, IDs, diagnostics,
candidates, nodes, branches, memory, and input
bytes bounded. Exact solving is for small trusted or conservatively bounded
instances. Cancellation is checked during search; callbacks run synchronously
and must honor their context without spawning unmanaged work.

Canonical JSON rejects unknown fields, duplicate keys, excess depth, excess
collections, trailing values, unsupported versions, and oversized input.
Visualization output must escape labels and may consume only verified plans.

Treat callbacks as trusted application code. Panics become typed errors, but
callbacks must still bound latency, honor cancellation, avoid shared mutation,
and never be loaded from serialized input.
Callback views reject more than 10,000 prior placements or a conservative
16 MiB owned-copy estimate before cloning any collection.
Solver and verifier options reject more than 32 callbacks before cloning or
executing the list. The monetary objective accepts at most 1,000 type costs
with 1,024-byte IDs by default; larger trusted maps require explicit limits.

For untrusted work, lower item, container, candidate, node, branch, memory,
diagnostic, and ID limits before decoding or normalization. A deadline must not
be the only bound. Log bounded diagnostics, not full hostile payloads or search
traces.

The pinned `go-library-tools` release runs secret scanning, deterministic SBOM
generation, reproducible source-archive checks, vulnerability review, and the
other supply-chain gates through `golib check --all` and
`golib release dry-run`. NilAway analyzes production files under the module
import prefix and remains visible but advisory until its signal policy is
promoted separately.
