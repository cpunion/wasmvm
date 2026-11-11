# Ready byte sources: remove asynchronous invocation overhead

Candidate `56d383c70a7c4f5cdbf77b8bf499333f40ff60f6` adds an optional immediate
packed-byte producer to component sources. Complete P3 HTTP execution improves,
especially for large bodies. Competitive parity and release acceptance remain
open. The source base is `6eea415`, following the
[immutable-input event](performance-http-snapshot-2026-11-11.md).

## Implementation

`AsyncSource.ReadBytes` may return an already available `stream<u8>` batch
without creating a Function, host invocation, task or external-work completion.
It receives a byte limit, must not block or enter the VM, and has no Caller.
An unavailable probe delegates to the existing required `Read`/`ReadOwned`
callback. Other element types cannot opt into this path.

Task/reentry admission checks, cancellation, full typed byte validation, count
and length checks, value/byte reservations, guest memory copying and exactly-once
source closure remain. Reservations stay charged while copying and notifying
the reader. The existing checked copy/destination and resource settlement code
is shared; no second scheduler or guest execution framework is introduced.

HTTP uses this path for immutable snapshots and cached read-ahead bytes.
Unready external Readers retain bounded I/O workers and the ordinary completion
protocol. Both buffered and streaming APIs retain their previous ownership
contracts. Snapshot preparation remains distinct from changing mutable input;
this change adds no benchmark selector or prepared-handler cache.

## Matched production comparison

M4 Max/macOS arm64, Go 1.27.3, Clang 22.1.8 and Wasmtime 48.0.1; identical
Rust P3 HTTP guest and payloads. Four stores, 16 clients, 256 requests/client.
Three randomized rounds, seed `2026111257`, run base/head ordinary Reader,
base/head immutable snapshot, and Wasmtime, with an independent identical-binary
A/A alias for all five configurations at both payload sizes. A first batch and
two warmups precede 30 measured batches of 4096 requests.

All **60 processes and 7,372,800 measured responses** pass complete body,
status, header, trailer and CRC32 checks. CPU/instruction counters cover
7,127,040 checked requests in 29 inter-batch windows per process. Darwin CPU
includes all threads and converts ticks at 125/3 ns. Tables show medians of
three process observations, each using its median warm batch duration. No owned
tests, builds, profiling or cache pruning overlap this production campaign.

| Payload | Configuration | req/s | CPU µs/request | Instructions/request |
|---|---|---:|---:|---:|
| 16 B | Base Go, ordinary Reader | 36,772 | 108.34 | 1,334,955 |
| 16 B | Head Go, ordinary Reader | 37,248 | 106.44 | 1,322,249 |
| 16 B | Base Go, snapshot | 41,770 | 94.66 | 1,178,088 |
| 16 B | Head Go, snapshot | 43,341 | 91.50 | 1,144,055 |
| 16 B | Wasmtime | 77,164 | 23.40 | 205,970 |
| 64 KiB | Base Go, ordinary Reader | 24,260 | 162.34 | 2,062,798 |
| 64 KiB | Head Go, ordinary Reader | 26,201 | 150.10 | 1,906,347 |
| 64 KiB | Base Go, snapshot | 28,924 | 136.56 | 1,766,764 |
| 64 KiB | Head Go, snapshot | 31,848 | 124.27 | 1,588,914 |
| 64 KiB | Wasmtime | 67,259 | 31.12 | 295,613 |

For 16 B/64 KiB, snapshot throughput points improve **3.76%/10.11%**, CPU
falls **3.33%/9.00%**, and instructions fall **2.89%/10.07%**. Allocation
events/request fall **874 -> 815 / 1222 -> 923**; Go allocated bytes fall
**68,336 -> 63,608 / 161,216 -> 137,480**. Ordinary Reader throughput points
improve **1.29%/8.00%**, with **19/259** fewer allocations. Small-request
changes remain near timing resolution; they are not a proven speedup.

**24/30 throughput, 25/30 CPU and 30/30 instruction A/A controls** fall within
±2.5%. All six unresolved timing controls, including head snapshot controls,
are retained. These local default-scheduled observations do not establish
controlled milestone acceptance. Do not compound gains from different events.
Snapshot throughput gaps remain **1.78×/2.11×**, CPU gaps **3.91×/3.99×**, and
instruction gaps **5.55×/5.37×** to Wasmtime. Reference Go allocation counters
exclude Rust/native heaps. Other engines' Core and native-host comparisons
remain separate dated records; this event supplies no new ranking for them.

Preparation is separate. Head snapshot validate/compile takes 162.14/164.39 ms,
instantiation 0.488/0.491 ms, input preparation 0.000541/0.008667 ms across all
four workers, first 4096-request batch 95.54/129.61 ms, and cleanup
0.056/0.057 ms. Wasmtime compile takes 49.62/49.38 ms and its first batch
57.87/64.97 ms. Observed lifetime-peak footprint at batch markers is
113.2/117.3 MiB for head snapshot and 21.1/22.5 MiB for Wasmtime; it includes
preparation and is not a final post-exit peak or an established memory win.

## Diagnosis and rejected compaction

Separate allocation-stack captures use 1000 direct complete kernel calls per
payload, one store and `MemProfileRate=1`, excluding admission-driver overhead.
The old source-read subtree accounts for approximately **88/364** profiled
allocations/request at 16 B/64 KiB, including **36/108** directly in invocation
setup. Its increase explains 276 of the 348 additional profiled allocations.
Large output collection allocates one 64-KiB body; repeated geometric output
growth is not the observed cause in this fixture.

Two separate Running-only traces localize source/host setup and canonical
conversion. Their sampled CPU is 96.81/173.54 µs/request; the large observation
differs materially from unprofiled execution. Inclusive source frames contain
required nested work and cannot be treated as removable overhead or as a
paired subtraction against Wasmtime.

An earlier two-factor prototype coallocated source invocation objects and Caller
expiration state. Its 48-process screen checks 589,824 measured responses and
removes up to 161 allocations/request, but establishes no throughput gain.
It is rejected and absent from the candidate. Early cache-pruning I/O overlaps
that diagnostic screen; its timings are not qualifying evidence. A controller
name collision also fails the initial preflight before the first counter
snapshot; that attempt is retained separately. Neither favorable reruns nor
allocation counts alone justify the retained optimization. The production
candidate removes the unnecessary invocation protocol itself.

Post-change captures with the same direct-call harness reduce the source-read
subtree to **28/64** profiled allocations/request, approximately 60/300 fewer.
Generated invocation setup drops to three reads per request at both sizes.
Codec creation still contributes 92 allocations for the large request, and VM
argument encoding, Caller creation and result decoding contribute 141; these
counts are not CPU fractions. New Running-only traces normalize to
90.75/122.47 µs/request, near unprofiled 91.50/124.27. They retain codec/layout,
ownership and native-driver preparation hotspots; no single sampled helper
accounts for the remaining engine gap. The initial head trace used a 16-B
command with a 64-KiB post-validator; the actual valid 16-B observation and the
controller mismatch are retained separately from the corrected 64-KiB trace.

## Validation and remaining work

All 32 Go test-package groups pass. All 31 executable packages retain at least
**90.86%** statement coverage: component **92.50%**, P3 **91.29%**, and the
measurement module **96.18%**. Focused component/source/HTTP race tests,
whole-module vet and no-cgo build pass. Tests exercise ready/zero/binary reads,
deferred fallback, count/type/bound failures, task/reentry/byte/value limits,
cancellation, destination failure, producer/close panic and cleanup, in addition
to the complete existing component and HTTP semantics.

Pinned LLGo `3d92a58 -O3 -lto=full` passes full component/P3 packages and the
measurement module on macOS arm64. Docker Linux amd64 passes all six affected
component, P3 and measurement-module compiler/package groups. All 914 frozen
runtime/test build inputs match the validated source. All 32 root test packages remain assigned once across
eight LLGo shards. No compiler upgrade or additional CI matrix is introduced.

The remaining gap still includes checked Canonical ABI conversion, owned stream
forwarding and Core/host transition preparation. These need localization in the
new candidate before another optimization; reducing object counts is not enough.
Final-head hosted checks/review, performance acceptance and release remain open.
The contribution slot remains occupied by
[LLGo #3057](https://github.com/xgo-dev/llgo/pull/3057).

[Data/provenance](data/performance-ready-bytes-2026-11-12.json) and
[all 108 production/rejected process records](evidence/performance-ready-bytes-2026-11-12.jsonl.gz)
retain counters, controls and phase arrays. Only per-request latency vectors
are omitted from compressed publication; complete local originals, allocation
profiles, failed preflight and diagnostic traces remain checksummed.
