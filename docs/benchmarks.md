# Benchmark methodology

[Documentation index](README.md) · [Repository entry](../README.md)

The dated results and `verification/evidence.json` are historical v1 records.
Their original source identities, dependencies, hashes and raw samples are
retained unchanged. They are not v2 performance or mutation proof. Current
packing behavior, security boundaries and public APIs are checked independently;
a new performance claim requires a new applicable benchmark run.

Each result records fixture hash, item and container counts, constraints,
objective order, feasibility and verification time, score, container count,
utilization, known bound or optimum, solver work, allocations, bytes, machine,
Go version, execution revision, input fingerprint, seed, and parallelism. Raw
results are evidence, not API. Freshness follows the complete input
fingerprint; the execution revision is retained only for traceability.
Dependency selection and integrity are fingerprinted through `go.mod` and
`go.sum`, including owned dependencies used by each benchmark profile. The
BoxPacker profile excludes only the nested reference module's checksum of the
parent Knapsack zip because that zip contains this evidence and would create a
self-referential digest cycle; its `go.mod` checksum and every other checksum
remain inputs.

Reference comparisons use identical normalized lattices, weights, rotations,
stock, constraints, and objectives. Invalid or lower-quality output is not a
runtime win. Claims such as faster, better, or optimal require reproducible raw
evidence and a valid common semantic subset.

The pinned BoxPacker 4.2.0 adapter uses millimetres, grams, unrestricted
orthogonal rotation, unlimited copies of one box type, and pack-all semantics.
The root module's typed test operation runs the PHP reference integration and
independently verifies every emitted placement with the Go verifier before
comparing exact statistics.

The typed benchmark operation builds the Go adapter once, performs one
warm-up process per implementation, then alternates ten fresh PHP and Go
processes. Wall time includes process and runtime startup, Composer autoload,
fixture construction, solving, and JSON serialization. It excludes compilation,
Composer installation, and independent verification. Every result is decoded
and independently verified before its timing, internal solve time, quality, or
peak RSS is recorded. Both delayed adapters are also killed through the same
50 ms external deadline. The raw evidence reports that PHP and Go do not expose
comparable cross-process allocation counts. No speed ratio or superiority claim
is made from this single tiny common-subset fixture.

The checked D-Wave `sample_data_1` fixture omits weight because the source does
not define it, preserves all permitted orthogonal rotations, and records a
reproducible heuristic upper bound of one container. That bound is a regression
threshold, not an independently proven optimum. The pinned revision, source
SHA-256, Apache-2.0 license, and conversion notice are recorded in
`verification/corpora.tsv` and `NOTICE`.

Checked-in raw evidence:

- [2026-09-13 Apple M4 Max](benchmarks/2026-09-13-darwin-arm64.md)
- [2026-09-13 BoxPacker fresh-process comparison](benchmarks/raw/2026-09-13-boxpacker-runtime.json)
- [2026-08-15 Apple M4 Max](benchmarks/2026-08-15-darwin-arm64.md)
- [BoxPacker fresh-process comparison](benchmarks/raw/2026-08-15-boxpacker-runtime.json)
- [2026-08-12 Apple M4 Max](benchmarks/2026-08-12-darwin-arm64.md)
- [2026-08-08 Apple M4 Max](benchmarks/2026-08-08-darwin-arm64.md)
- [2026-08-02 Apple M4 Max](benchmarks/2026-08-02-darwin-arm64.md)

`benchmark-compare` collects ten fresh samples for tiny exact, ordinary,
orientation-heavy, weight-limited, stability-heavy, finite-stock, impossible,
fragmented adversarial, large bounded, verifier, and cancellation workloads.
The stability-heavy fixture combines exact support, transitive load, stack
count, and exact three-axis content center-of-gravity bounds.
It enforces reviewed p50, p95, allocation, byte, solution-quality, and
search-work budgets from
`verification/benchmark-thresholds.tsv`. It does not claim that a change beats
an external solver or prior release.

The same operation compiles the solver benchmark binary once and executes
every benchmark in five fresh processes under `/usr/bin/time`. The gate records
the maximum resident set size for each workload and fails unless the benchmark
inventory exactly matches `verification/benchmark-rss-thresholds.tsv`. The RSS
measurement includes the Go runtime and benchmark process but excludes
compilation. Darwin reports bytes directly; GNU time reports KiB, which the gate
converts to bytes before comparison.

## Complete-search calibration: 2026-10-13

Only `BenchmarkExactTinyFixed` is recalibrated for the corrected complete
integer-lattice search. Its unchanged fixture packs three labelled unit cubes
with XYZ orientation into one 3-by-1-by-1 container, using default options.
Complete search attempts 171 placements and visits 82 nodes; the earlier compact
search attempted 123 placements and visited 16 nodes. Restoring that incomplete
search would not be a valid performance improvement.

A matched Go 1.27.2 Darwin/arm64 Apple M4 Max run used `GOMAXPROCS=2`, seed zero,
solver parallelism one and ten 100 ms samples of this exact fixture. Setup and
the benchmark helper's final verification were excluded from timing; solver-owned
verification remained included. Before and after component-only scoring stopped
serializing a discarded canonical tie, the complete-search measurements were:

| Metric | Before scoring optimization | After scoring optimization | Current budget |
| --- | ---: | ---: | ---: |
| p50 ns | 558,784 | 392,290 | 1,000,000 |
| p95 ns | 582,746 | 442,253 | 1,200,000 |
| Maximum B/op | 1,096,392 | 880,467 | 1,048,576 |
| Maximum allocs/op | 5,647 | 5,347 | 6,000 |
| Candidates/op | 171 | 171 | exactly 171 |

The comparison used source `4166acb5415fe3760fd50c34552bef8ab4038f72` and its
scoring correction, whose `objective/objective.go` SHA-256 is
`b01acb00a80759c4f1f893f9246daad57467451f017fd0c0c446c64490197177`.
The unchanged benchmark source SHA-256 is
`c4a929f1f05bff4c0894d2da7394d76b1ee1aa23c1c50915706c8c6d59ed85fc`;
module and dependency selections were identical. These are focused local
measurements, not a complete benchmark qualification or a public latency SLA.
The first Linux qualification of this scoring correction,
[run 38469017853](https://github.com/faustbrian/go-knapsack/actions/runs/38469017853)
at `da1230be89d868ca996ee2ac133e6982c9cb69c0`, used Go 1.27.2 on
Linux/amd64, AMD EPYC 7763, benchmark parallelism suffix 4 and ten 100 ms
samples. Tiny Exact measured p50 888,188 ns, p95 898,474 ns, maximum 880,713
B/op and 5,347 allocs/op; it retained 171 placements, 82 nodes and 100%
utilization. Its earlier 600,000/800,000 ns latency budgets failed, while
allocation, byte, search-work and every RSS budget passed. This failed run
is not passing qualification. The earlier complete pre-optimization Linux
[run 38451595094](https://github.com/faustbrian/go-knapsack/actions/runs/38451595094)
used a different AMD EPYC 9V45 and measured 641,643/651,042 ns; it is not a
matched host comparison and does not prove a Linux speedup or regression.

The current 1,000,000/1,200,000 ns budgets provide approximately 12%/34%
headroom over the measured supported-host median/tail. The byte and allocation
budgets retain about 19%/12% headroom. This envelope is selected from actual
local and hosted observations, not a public latency SLA or automatic host-based
waiver. Native qualification must still pass it on its actual execution host.
Only this case's two latency budgets change; the one-container, exact-placement,
100% utilization, all other case and every RSS requirement remain unchanged.
Historical v1 raw samples, hashes and `verification/evidence.json`
retain their original identities; the historical integrity assertion checks the
original threshold hash rather than attributing those results to today's budgets.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
