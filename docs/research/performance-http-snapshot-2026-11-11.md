# Explicit immutable HTTP inputs and remaining Go overhead

An explicit immutable request body improves the Go P3 HTTP candidate, while
ordinary Readers retain their scheduled I/O contract. Competitive parity is
still open. This event follows the [retained request-storage comparison](performance-go-frame-plans-2026-11-11.md).

## Implementation and input contracts

Candidate `457921f6857abb91b483c684c69ecf5906342012` adds
`preview3.NewSnapshotRequest(ctx, method, url, body string)`. A caller supplies
already available immutable bytes, including arbitrary binary data. Bounded
string slices can reach stream completion without a mutable transfer-buffer
snapshot or an external I/O worker. Both body and reader locks use nonblocking
admission; contention falls back to the existing worker path. Ordinary
`http.Request` Readers still use that path.

Both paths share length accounting, EOF/trailer publication and reader closure.
Cancellation, resource limits, guest memory copies, response collection and
joined cleanup remain. No guest/compiler shortcut or competitor fallback is
introduced. The constructor also supplies an independent replayable `GetBody`.

The reference uses its existing ready `Full<Bytes>` input. Snapshot mode makes
the owned input's readiness explicit; it does not establish the same speedup
for arbitrary Readers. Conversion from mutable input to a string is charged to
the separate `requestPreparation` phase, once per worker. Per-request HTTP
construction remains inside batch wall/CPU measurements and outside the public
API interval. Mutable data that changes each request must pay its own snapshot
cost; this event does not measure that workload.

## Complete production comparison

M4 Max/macOS arm64, Go 1.27.3, Clang 22.1.8 and Wasmtime 48.0.1. The same
pinned Rust P3 HTTP component runs with four stores, 16 clients and 256
requests/client. Three randomized rounds, seed `2026111241`, compare retained
source, new ordinary Reader, explicit snapshot and Wasmtime, each with its own
identical-binary A/A alias, at 16 B and 64 KiB. First execution and two warmups
precede 30 measured 4096-request batches. Every response checks status, header,
complete body, trailer and CRC32.

All **48 processes and 5,898,240 measured responses** pass. Process counters
cover 29 inter-batch intervals each, totaling 5,701,632 checked requests. CPU
includes all process threads, with Darwin ticks converted at 125/3 ns. Rates
are medians of three process observations; each observation uses its median
warm batch. No tests, builds or profilers run concurrently with measurement.

| Payload | Configuration | req/s | CPU µs/request | Instructions/request |
|---|---|---:|---:|---:|
| 16 B | Retained Go | 34,889 | 113.39 | 1,331,515 |
| 16 B | Candidate Go, ordinary Reader | 35,401 | 112.00 | 1,332,273 |
| 16 B | Candidate Go, snapshot | 39,894 | 99.12 | 1,175,107 |
| 16 B | Wasmtime | 75,858 | 23.84 | 206,449 |
| 64 KiB | Retained Go | 22,681 | 172.55 | 2,058,303 |
| 64 KiB | Candidate Go, ordinary Reader | 23,014 | 170.34 | 2,059,480 |
| 64 KiB | Candidate Go, snapshot | 27,532 | 143.17 | 1,762,757 |
| 64 KiB | Wasmtime | 66,118 | 32.54 | 297,668 |

Snapshot versus the candidate's ordinary Reader improves throughput points
**12.69% / 19.63%**, reduces CPU **11.50% / 15.95%**, and reduces instructions
**11.80% / 14.41%** for 16 B/64 KiB. Approximately 55 allocations/request are
removed: **929 -> 874 / 1277 -> 1222**. Go allocated bytes/request fall
**71,624 -> 68,336 / 230,024 -> 161,216**. These are complete request counters,
not an allocation-only probe. Reference Go counters exclude Rust/native
allocations and must not be compared as total engine allocation costs.

All **24/24 throughput and 24/24 instruction A/A controls** fall within ±2.5%.
The ordinary path shows no material regression in this cohort. Three local
rounds with default scheduling and background system activity remain advisory,
not controlled milestone acceptance. Every process and adverse control is
published. Even snapshot input retains **1.90× / 2.40× throughput gaps**,
**4.16× / 4.40× CPU gaps** and **5.69× / 5.92× instruction gaps** to Wasmtime.
This P3 event supplies no new native-host or other-engine Core ranking.

Preparation stays separate. Snapshot Go validate/compile takes 165.17/163.75 ms,
instantiation 0.483/0.494 ms, input preparation 0.000584/0.007500 ms across all
four workers, first 4096-request batch 103.93/152.68 ms, and cleanup
0.060/0.055 ms. Wasmtime compile takes 49.72/50.93 ms and its first batch
58.63/69.11 ms. Observed lifetime-peak footprint at batch markers is
113.0/117.8 MiB for snapshot Go and 21.5/23.0 MiB for Wasmtime; this includes
preparation and is not a final post-exit peak. Snapshot allocation reduction
does not establish a footprint reduction.

## Causal screening and JIT mapping

An earlier diagnostic-only 48-process screen keeps one binary and independently
toggles handler prevalidation and inline execution of the benchmark's known
nonblocking Reader. All 589,824 measured responses pass. Prevalidation removes
only about 0.5% of instructions and establishes no CPU improvement. Inline
execution reduces CPU approximately 8.6%/7.2%. Arbitrary Readers cannot assume
that readiness; these diagnostic shortcuts are absent from production. A
handler cache was therefore not implemented.

Two separate Running-only traces map compiled function ranges and the original
Wasm name section. Each checks 135,168 responses including first/warmup work.
Mapped JIT accounts for **20.49/24.54 µs per request**, with only
0.007/0.030 µs unresolved, from total sampled 112.43/166.53 µs. The approximately
54 µs payload-dependent increase contains only approximately 4 µs more mapped
JIT execution. Host work deserves priority; this is single-trace localization,
not a paired causal subtraction against another engine.

The leading Rust async `FutureState::callback` costs 2.54/2.66 sampled µs;
handler closure, exported handler and stream callbacks also contribute.
No single guest copy function dominates. Existing ARM64 bulk-copy generation
already uses vector operations. Allocation/collection, checked Canonical ABI
conversion, response collection, and VM/host state preparation remain the next
targets. Metadata caching or treating the full deficit as unavoidable GC cost
is unsupported by these measurements.

## Validation and integration

The full 32 Go test-package groups pass; all 31 executable packages retain at
least **90.87%** statement coverage. P3 provider coverage is **91.46%**, and the
measurement module reaches **96.18%**. Focused HTTP/snapshot race tests,
whole-module vet and no-cgo build pass. Actual guest tests cover both backends,
binary/empty/large bodies, bounded read-ahead, length mismatch, cancellation,
limits, replay, lock contention, trailer failure and retained response bytes.

Pinned LLGo `3d92a58 -O3 -lto=full` passes the complete provider and
measurement module on macOS arm64. Docker Linux amd64 passes HTTP/snapshot/Rust
guest execution and the complete measurement module under both compilers.
The full macOS LLGo invocations take 105.3 s for the provider and 60.6 s for the
measurement module, including compilation/linking as well as test execution.
The compiler pin is unchanged. All 32 root packages remain assigned exactly once across eight CI
shards; no duplicate full-core LLGo campaign is needed for this provider change.

[Data and provenance](data/performance-http-snapshot-2026-11-11.json) and
[all 96 production/screen process records](evidence/performance-http-snapshot-2026-11-11.jsonl.gz)
retain counters, controls, phase arrays, JIT attribution and raw-file checksums.
Only per-request latency vectors are omitted from compressed publication;
their original local files remain checksummed. Contribution integration waits
for the single active PR slot, currently occupied by
[LLGo #3057](https://github.com/xgo-dev/llgo/pull/3057). Final-head hosted checks,
review, full performance acceptance and production release remain open.
