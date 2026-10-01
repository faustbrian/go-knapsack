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
Recovered panic values and rejected JSON field/key/version details are omitted
from default errors. Generated-plan and rendering rejections also omit verifier
findings. Decoder errors retain their classification chains for deliberate
`errors.Is`/`errors.As` inspection; underlying decoder causes can contain input
details. `verify.Result.Violations()` is an explicit diagnostic interface, and
trusted callbacks may return their own errors or bounded decision messages.
Applications own access control and redaction before logging or returning those
explicit diagnostics. Review this boundary when adding a new error formatter,
callback integration, or externally exposed diagnostic endpoint.
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
