# Component scalar carriers: measured costs and remaining gap

Candidate `d469ce2afe29ad72f04c0404993194255018f4af` connects component
canonical callbacks to the VM's existing scalar host ABI. It removes public
VM Value boxing while keeping canonical validation and ownership. Instructions
and allocation counts fall; this noisy local cohort does **not** prove a
throughput win or competitive parity. The source base is `43994e5`, with the
same runtime as the preceding [ready-byte event](performance-ready-bytes-2026-11-12.md).

## Implementation and causal screen

Synchronous canonical lowerings, resource/error-context/backpressure operations,
nonblocking stream/future operations, task.return and synchronous waitable
operations now use `Caller.NewScalarFunction`. Genuine suspension retains the
existing continuation protocol. Scalar and suspendable stream/future callbacks
share one implementation. Flat arguments use bounded local canonical slots;
parameter/result memory, reallocation, resource transfers, borrow cleanup,
limits, cancellation and task state checks retain their existing paths. Signed
zero and canonical NaN normalization remain unchanged. No public API is added.

A separate diagnostic counted 100 complete 64-KiB requests in one Store:
**7 synchronous canonical, 13 waitable, 1 task.return and 29 nonblocking async
builtin callbacks per request**. These counts locate actual boundary traffic;
they are not CPU fractions. A proposed scalar suspension/reentry API had no
observed calls in this fixture. Its source/tests and two screening cohorts are
parked, and no VM API expansion is retained.

The initial two-factor screen separates canonical lowerings from other
nonblocking builtins: 48 processes and 589,824 measured responses. Two further
24-process screens evaluate the broader VM prototype. All response checks
pass, but they establish no incremental throughput benefit from the new APIs.
The final candidate instead migrates the remaining observed task/waitable
traffic through the existing ABI. The original copied scope strings in the
later screen plans are stale; separate qualification records identify their
actual single-prototype scope. No screen replaces the final comparison below.

## Matched complete-request comparison

M4 Max/macOS arm64, Go 1.27.3, Clang 22.1.8, Wasmtime 48.0.1. Identical
Rust P3 HTTP component, four Stores, 16 clients, 256 requests/client, fuel off
with owned cancellation and logical limits retained. Three randomized rounds,
seed `2026111277`, include source base/head ordinary Reader and immutable
snapshot, plus Wasmtime; each configuration has an independent identical-binary
A/A alias. A first batch and two warmups precede 30 measured 4096-request batches.
No owned builds, tests, profiling or cache pruning overlap these processes.

All **60 processes and 7,372,800 measured responses** pass complete body,
status, header, trailer and CRC checks. Inter-batch CPU/instruction windows
cover 7,127,040 checked requests. Darwin process CPU includes every thread and
converts ticks at 125/3 ns. Values below are medians of three process
observations. The A/A columns are reported separately, not pooled selectively.

| Payload | Configuration | req/s | A/A req/s | CPU µs/request | Instructions/request | Allocations/request |
|---|---|---:|---:|---:|---:|---:|
| 16 B | Base Reader | 35,263 | 35,411 | 112.65 | 1,320,518 | 910 |
| 16 B | Head Reader | 35,040 | 35,404 | 113.24 | 1,284,747 | 882 |
| 16 B | Base snapshot | 39,848 | 40,769 | 99.14 | 1,140,333 | 815 |
| 16 B | Head snapshot | 40,691 | 39,121 | 97.40 | 1,108,929 | 789 |
| 16 B | Wasmtime | 70,554 | 74,021 | 28.03 | 210,620 | 1* |
| 64 KiB | Base Reader | 24,213 | 24,186 | 161.32 | 1,906,135 | 1,018 |
| 64 KiB | Head Reader | 24,423 | 20,826 | 164.51 | 1,857,208 | 978 |
| 64 KiB | Base snapshot | 29,509 | 30,346 | 133.06 | 1,587,308 | 923 |
| 64 KiB | Head snapshot | 31,579 | 30,415 | 125.09 | 1,543,724 | 885 |
| 64 KiB | Wasmtime | 65,420 | 62,506 | 33.55 | 298,659 | 1* |

\* Reference Go counters exclude Rust/native allocations. Snapshot allocated
bytes fall **63,608 -> 59,760 / 137,480 -> 132,288**. Reader removes **28/40**
allocations and snapshot **26/38** at 16 B/64 KiB. Primary instruction counts
fall **2.57–2.75%** across the four owned configurations; independent A/A
instruction medians also retain the reduction.

Snapshot primary throughput points change **+2.12%/+7.02%**, CPU
**−1.76%/−5.99%**. Reader throughput changes **−0.63%/+0.87%**, CPU
**+0.52%/+1.98%**. Only **12/30 throughput, 13/30 CPU and 27/30 instruction
A/A controls** stay within ±2.5%; worst deviations are 55.71%, 57.48% and
6.27%. This cohort has substantial scheduling/system noise, including adverse
head controls. The small-case snapshot direction reverses in its A/A medians.
Do not claim a general throughput improvement, select favorable aliases, or
compound these points with earlier events.

Primary head snapshot gaps to Wasmtime remain **1.73×/2.07× throughput**,
**3.47×/3.73× CPU** and **5.27×/5.17× instructions**. Ordinary Reader throughput
gaps are **2.01×/2.68×**. These are advisory cohort points, not acceptance
bounds. Other engines' Core/native-host results remain separate dated events;
this comparison adds no ranking for engines without this P3 service adapter.

## Separate preparation and lifecycle costs

| Phase | Head snapshot 16 B | Head snapshot 64 KiB | Wasmtime 16 B | Wasmtime 64 KiB |
|---|---:|---:|---:|---:|
| Validate/compile, ms | 166.25 | 162.97 | 52.94 | 51.78 |
| Instantiate, ms | 0.476 | 0.503 | 0.197 | 0.200 |
| Input preparation, ms | 0.000625 | 0.007875 | Included in reference request construction | Included in reference request construction |
| First 4096-request batch, ms | 103.74 | 133.97 | 61.78 | 65.27 |
| Cleanup, ms | 0.0617 | 0.0540 | 0.239 | 0.253 |
| Observed lifetime peak, MiB | 112.1 | 118.9 | 21.7 | 23.0 |

Input preparation covers all four owned workers and requires explicitly
immutable input; mutable Reader input retains its copy/worker contract. Peak
footprint includes preparation and is observed at batch markers, not after
process exit. Full phase arrays, load/provider setup and first/warm batches
are retained in the data. Preparation is not charged to warm throughput.

## Validation and next work

All 32 Go package groups pass, with all 31 executable packages at or above
**90.85%** statement coverage: component **92.39%**, P3 **91.31%**, and the
measurement module **96.18%**. Affected canonical/source/HTTP race tests, vet
and the no-cgo build pass. All 1,255 tracked build/test inputs checked against
the frozen source match. Existing LLGo shard coverage remains unchanged.
Pinned LLGo `3d92a58 -O3 -lto=full` passes full component/P3 and measurement
module execution on macOS arm64. Docker Linux amd64 passes all six affected
Go/LLGo component, P3 and measurement-module groups, including actual guest
execution. No compiler/profile or CI matrix change is introduced.
Hosted final-head integration and controlled performance acceptance remain open. The contribution slot is still occupied by external
[LLGo #3057](https://github.com/xgo-dev/llgo/pull/3057); this does not authorize
merging that upstream PR.

The remaining gap is much larger than this boundary reduction. Prior complete
Running traces still localize work across component task preparation, canonical
conversion/ownership and native driver transitions. Removing boxes at observed
edges does not eliminate those protocols. Next, isolate the actual CPU cost of
those paths before another structural change; allocation counts or overlapping
inclusive frames cannot establish the removable share. Keep Go performance
first and pinned LLGo semantic compatibility.

[Data and source/toolchain identity](data/performance-canonical-scalars-2026-11-12.json)
and [all 156 production/screen records](evidence/performance-canonical-scalars-2026-11-12.jsonl.gz)
retain all controls and phase/counter arrays. Only per-request latency vectors
are omitted from compressed publication. Local originals, diagnostic counts,
parked source/tests, corrected test failures and the failed no-measurement
preflight remain recorded separately.
