# Cross-instance call veneers and remaining P3 costs

**Implemented local candidate.** Compiled frame plans remove native execution
of transparent cross-instance forwarding bodies. The marginal LLGo throughput
result is approximately flat despite fewer instructions. Performance and release
acceptance remain open; the full hosted final-head matrix is pending.

The subsequent [compiled-schema comparison](performance-value-comparison-plans-2026-11-11.md)
adds resource-aware validation programs and retains a fresh complete matched
cohort. The numbers below remain this earlier stage's independent observations.

## Implementation and structural result

Previous candidate `f3531b0074b9123754fc808071bdef89d3909981` includes task
context lowering. Final implementation `8a25766bbb2699acbb400636c72c0c7baa4acc77`
adds [call-veneer plans](../design/call-veneers.md). An inspected counter copy
finds that all 18 cross-instance calls in this P3 fixture enter eleven short
parameter/constant-table forwarding functions. No import names or HTTP shapes
are special-cased.

The compiler rejects computed work, dynamic indices and protected control flow.
The native driver retains real frames, slot quotas, reference roots and Caller
instance, while the common dispatcher reads the live target and checks its type.
Insufficient prefix fuel executes the original machine artifact for the precise
failure boundary. Missing or closed native artifacts retain backend errors.
Same-instance native linking and ordinary machine bodies remain available.

Independent Go and production LLGo counter copies each verify 1,000 complete
responses and agree on all counts. Instrumented timing is excluded.

| Event/request | Main | Context candidate | Veneer candidate |
|---|---:|---:|---:|
| Native entries | 186 | 142 | 106 |
| Native Call exits | 99 | 62 | 44 |
| Native Return exits | 87 | 80 | 62 |
| Core Run boundaries | 93 | 56 | 56 |
| Generic host callbacks | 81 | 44 | 44 |
| Logical cross-instance calls | 18 | 18 | 18 |
| Component tasks | 9 | 9 | 9 |

This increment removes 36 native entries/request (25.4%); the combined candidate
removes 80 (43.0%) versus main. Logical calls and their limits are preserved.
General cross-instance functions with computation still use checked boundaries.

## Matched source and reference comparison

M4 Max/macOS arm64, Go 1.27.3, production LLGo `3d92a58`, experimental compiler
`e2cca2c`/sharded runtime `f2e4cb426`, and independent Wasmtime 48.0.1. LLGo uses
`-O3 -lto=full`, identical within each source series. The experimental toolchain
remains separate from production and its outstanding upgrade qualification.

The same 16-byte plain Rust HTTP guest runs with four workers, 16 clients and
256 requests/client. Status, headers, complete body, trailers and CRC32 remain
checked. Three fixed rounds, seed `2026111203`, run first execution, two warmups
and 30 measured batches. Each compiler has main, previous candidate, head and
an independent main A/A process; Wasmtime is independent. All **39 processes**
and **4,792,320 measured responses** pass. Main is `4ce75302`, represented by
retained binaries with exactly equivalent production implementation `0aaab1ad`.
Compiler, source and executable identities remain in the data.

Darwin process counters cover 29 inter-batch intervals/process and **4,632,576
checked requests**. CPU sums process threads and uses 125/3 ns per Mach tick.
Startup/cleanup are outside that window; batch/readiness bookkeeping remains
included. Each process's req/s uses its median warm batch duration. The table
reports medians of three per-process observations.

| Configuration | req/s | CPU µs/request | Retired instructions/request |
|---|---:|---:|---:|
| Main / Go | 26,667 | 148.41 | 1,893,571 |
| Context / Go | 27,988 | 141.29 | 1,786,492 |
| Veneer / Go | 30,802 | 128.39 | 1,692,651 |
| Main / production LLGo | 7,224 | 652.36 | 2,652,079 |
| Context / production LLGo | 8,022 | 584.78 | 2,463,930 |
| Veneer / production LLGo | 7,945 | 589.38 | 2,385,064 |
| Main / experimental LLGo | 13,050 | 373.49 | 2,565,909 |
| Context / experimental LLGo | 14,096 | 347.47 | 2,379,073 |
| Veneer / experimental LLGo | 14,185 | 343.69 | 2,292,844 |
| Wasmtime | 75,320 | 24.16 | 207,434 |

Relative to the context candidate, instruction counts fall **5.25% / 3.20% /
3.62%** for Go/production LLGo/experimental LLGo. Throughput points change
**+10.05% / −0.96% / +0.63%**. Production LLGo's CPU point also worsens by
0.79%. These adverse observations are retained. Fewer entries do not establish
an LLGo throughput improvement for this increment.

The complete candidate's throughput points versus main improve **15.51% /
9.98% / 8.70%**, and instructions fall **10.61% / 10.07% / 10.64%**. These are
direct same-cohort comparisons; earlier source-event percentages are not added.
Only **5/9 timing A/A controls** resolve within 2.5%; all **9/9 instruction
controls** resolve. Frequency and neighbor load remain uncontrolled. No
sustained acceptance or controlled ranking is claimed.

The final experimental LLGo throughput gap to Wasmtime is **5.31×**, with an
**11.05×** instruction gap. Production LLGo's throughput gap is **9.48×**.
The reference's ready `Full<Bytes>` body and generated typed binding differ
from the owned public `http.Request`/`io.Reader` adapter and bounded I/O workers.
Those contracts remain unchanged and explicit.

## Preparation and footprint

Separate phase medians below use milliseconds. `firstCall` is the first
4096-request batch. Instantiation sums workers. Observed lifetime-peak footprint
at batch markers includes preparation; it is not the final post-exit peak.

| Head configuration | Validate/compile | Instantiate | First batch | Cleanup | Observed peak MiB |
|---|---:|---:|---:|---:|---:|
| Go | 164.95 | 0.56 | 136.73 | 0.09 | 112.4 |
| Production LLGo | 251.50 | 0.90 | 513.79 | 0.57 | 164.4 |
| Experimental LLGo | 238.72 | 0.79 | 288.02 | 0.57 | 180.7 |
| Wasmtime | 66.03 | 0.82 | 59.33 | 0.36 | 21.6 |

## Residual CPU attribution

A separate Time Profiler diagnostic launches the exact final experimental LLGo
binary, verifies 122,880 measured responses and a successful target exit. All
47,313 exported samples have state `Running`; waiting-thread recording is off.
The interval from first to last actual kernel sample includes first execution,
two warmups and 30 measured batches: 135,168 checked requests. Normalized sample
weight is approximately 346.4 CPU µs/request. Profiled timing is excluded from
the source comparison, and this is a single post-change profile, not a paired
category comparison.

| Exclusive sampled category | Share |
|---|---:|
| Allocation/free/accounting | 21.4% |
| Collector paths | 17.2% |
| Component calls/tasks/lifetimes | 13.5% |
| P3 HTTP/provider | 12.1% |
| Core execution/native backend | 8.5% |
| Anonymous/JIT/unresolved leaf | 7.7% |
| Canonical values/resources | 5.3% |
| Component type comparison | 4.2% |
| Service harness/host scheduling | 4.0% |
| VM API/marshalling/lifetimes | 2.9% |
| Native entry | 2.4% |
| Other host/runtime | 0.9% |

Allocation/collector paths together account for **38.6%**. Component/task,
canonical-value and type-comparison work adds **23.0%** as separate categories.
`asyncEnd.matches` and `Instance.end` still dominate type-comparison callers.

A cross-cutting view of `psynch` leaf samples finds about **103 µs/request**
of system synchronization work: **48.4 µs** in collector/allocator paths,
**43.6 µs** in channel/select paths, **2.3 µs** in timer paths and **8.8 µs**
elsewhere. These values overlap the exclusive categories and must not be added
to them. Runtime thread-entry wrappers are excluded when classifying collector
synchronization. Channel stacks reach the persistent I/O workers, task awaits
and admission driver through `chanSignal.park`, `ChanRecv`, `ChanSend` and
`Select`. They identify a concrete direction for isolated scheduling tests;
they do not prove a new runtime bug or attribute all cost to one primitive.
The earlier channel-allocation issue [LLGo #3035](https://github.com/xgo-dev/llgo/issues/3035)
is closed and is not duplicated by this observation.

The remaining work should prioritize checked Canonical ABI plans and temporary
ownership, then isolate native channel/scheduling cost. Further boundary
lowering remains useful but is insufficient by itself to close the measured
gap. The earlier finite GC-off experiment remains independent evidence;
disabling collection is not a production optimization or a parity result.

## Validation and status

All 32 Go internal/pkg test groups pass; 31 executable-package coverage gates
pass, minimum **90.93%**. Whole-module vet/no-cgo build and component/VM/native
adapter race tests pass. Production LLGo executes seven complete affected test
packages on macOS arm64. Local Docker Linux amd64 executes ten successful
Go/LLGo package groups, including the real Rust HTTP guest on both backends.

Tests cover exact fuel at every prefix, live table replacement, mismatched/null
targets, mixed vector/reference tuples, GC while parked, correct Caller memory,
expired capabilities, quotas, cancellation and artifact retirement. An early
unrestricted selection bypassed same-instance native linking and failed its
structural tests; final selection is cross-instance only and backend-admitted.
The first diagnostic helper had an import cycle; corrected counter copies keep
instrumentation separate. Those failed attempts are retained and unqualified.

The candidate remains on a branch under the one-active-contribution rule while
[LLGo #3057](https://github.com/xgo-dev/llgo/pull/3057) is active. Full hosted
final-head checks, reviews and controlled cross-target performance acceptance
remain required before merge/release claims.

[All process arrays, controls, profile categories and provenance](data/performance-call-veneers-2026-11-11.json)
and [compressed counter evidence](evidence/performance-call-veneers-2026-11-11.jsonl.gz)
retain every process. Per-request latency arrays remain in the full local raw
event with its recorded checksum; published counters omit those arrays.
