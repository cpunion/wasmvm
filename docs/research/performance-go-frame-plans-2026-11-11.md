# Go performance: retained request storage and rejected frame expansion

**Go performance is now the priority.** LLGo semantic compatibility remains
required; dedicated LLGo tuning is deferred. Competitive parity remains open.
Bounded host-request storage is retained. A subsequent frame-plan expansion was
measured and reverted because it did not establish a throughput improvement.

## Retained implementation

`08bb9dac778123685737d93e4804b7881b0dae8c` follows the previous `c97d6a7`
candidate. Each suspended execution now owns its internal request and copied
argument buffer. Inputs remain stable through host service/reentry, and result
handoff precedes retirement. Public callback values and Caller capabilities keep
their independent lifetimes. Success, failure, exception, abort and shutdown
clear request roots; idle retention is capped at 64 slots (1 KiB). Larger valid
tuples are released after use. No new global pool or lock is introduced.

A full Go allocation-stack diagnostic identifies 44 request objects and 38
nonempty argument slices per complete HTTP request. Separate 20,000-request
observers verify events/request **1011.004 -> 929.004 Go** and
**2889.048 -> 2807.054 production LLGo**, approximately 82 fewer events each.
All 80,000 observer requests and all four zero-event empty controls pass.
Observer timings are excluded. That prior cohort's throughput points change
+1.25% Go / +1.18% production LLGo / -4.48% experimental LLGo, with only 5/9
timing controls resolving. These observations do not establish throughput gains.
[Prior cohort data](data/performance-host-requests-2026-11-11.json) and
[counter evidence](evidence/performance-host-requests-2026-11-11.jsonl.gz) retain
all 30 processes and their qualifications.

## Go and independent component reference

M4 Max/macOS arm64, Go 1.27.3 and Wasmtime 48.0.1. Same Rust P3 HTTP guest,
four workers, 16 clients, 256 requests/client; owned responses use the bounded
buffered public API. Three rounds, seed `2026111227`, run main, retained source,
expanded candidate, retained-source A/A, Wasmtime and reference A/A for both
16-byte and 64-KiB responses. First execution and two warmups precede 30
measured 4096-request batches. Status, headers, full body, trailers and CRC32
remain checked. All **36 processes and 4,423,680 measured responses** pass.

CPU/instruction windows contain 29 inter-readiness intervals per process and
4,276,224 checked requests. Darwin CPU ticks use 125/3 ns and sum all threads.
Preparation/cleanup remain outside those windows; batch/readiness bookkeeping
is included. Each process's throughput uses its median warm batch duration;
the table reports medians of three process observations. Main `4ce75302` is
represented by its exactly equivalent `0aaab1ad` implementation binary.

| Payload | Configuration | req/s | CPU µs/request | Instructions/request |
|---|---|---:|---:|---:|
| 16 B | Main / Go | 26,179 | 150.91 | 1,891,825 |
| 16 B | Retained / Go | 34,290 | 114.93 | 1,331,075 |
| 16 B | Expanded / Go, reverted | 34,018 | 115.89 | 1,321,813 |
| 16 B | Wasmtime | 74,382 | 26.20 | 207,926 |
| 64 KiB | Main / Go | 18,275 | 214.75 | 2,672,113 |
| 64 KiB | Retained / Go | 22,072 | 177.17 | 2,058,325 |
| 64 KiB | Expanded / Go, reverted | 21,491 | 180.51 | 2,050,321 |
| 64 KiB | Wasmtime | 65,573 | 33.37 | 296,707 |

The retained branch's throughput points versus main improve **30.98% / 20.78%**
for 16 B/64 KiB; instructions fall **29.64% / 22.97%**. These are direct
same-cohort comparisons, not compounded historical percentages. Wasmtime's
throughput remains **2.17× / 2.97×** higher; instruction gaps remain about
**6.40× / 6.94×**. Only **5/12 timing A/A controls** resolve within 2.5%; all
**12/12 instruction controls** resolve. No controlled ranking or sustained
acceptance is claimed. Every process, including adverse controls, is retained.

Wasmtime's generated bindings/ready `Full<Bytes>` body differ from the owned
public `http.Request`/`io.Reader` adapter and bounded I/O workers. Those contracts
remain explicit. This P3 event does not replace the separate full Core and
synchronous-service comparisons or provide a new ordinary native-host ranking.

Preparation is separate from execution. For retained Go, validate/compile is
168.77/167.64 ms, instantiation 0.54/0.48 ms, first 4096-request batch
123.67/190.42 ms, cleanup 0.066/0.060 ms, and observed lifetime-peak footprint
112.9/118.3 MiB (16 B/64 KiB). Wasmtime observes 60.45/52.03 ms compilation,
63.53/66.86 ms first batch, and 21.5/22.4 MiB footprint. Footprint is sampled
at batch markers, includes preparation and is not a final post-exit peak.

## Rejected expansion and cost isolation

Experimental `420ed68` selected forwarding plans by their live host/cross-instance
target and handled immediate suspended returns without native entry. Ordinary
same-instance guest batches and computed bodies still executed in native code.
Counter copies verify **106 -> 99 native entries/request**, with unchanged
approximately 929 allocation events/request. Nevertheless, marginal throughput
points worsen **0.79% / 2.63%** and instructions fall only **0.70% / 0.39%**.
The extra selection path is not justified by these results. Revert `6697fcc`
restores production inputs exactly to `08bb9da`; the experiment and evidence
remain reachable in history. Passing semantic tests alone did not qualify it.

A separate Go CPU diagnostic checks 245,760 measured responses after first/two
warmups. It records 24.68 sample seconds over 7.41 wall seconds. Cgo/native
entry paths account for about 16.7% inclusive samples; span/scheduler lock paths
also appear prominently. **Those cgo samples include guest native work** and
are not a measurement of transition overhead alone. Sampling, unresolved native
frames and instrumentation prevent a causal category subtraction.

An independent system Time Profiler trace exports 15,371 `Running` samples,
15,178 inside the first-to-last HTTP kernel window (135,168 checked requests,
including first execution and warmups). Their normalized weight is 112.29
CPU µs/request, near the unprofiled 114.93. Exclusive attribution gives 14.0%
component calls/tasks/ownership, 8.1% canonical values/resources, 13.6% allocator,
8.8% collector, 12.8% Core execution/native driver, and 12.7% native C entry/JIT
protection/platform lifecycle. Another 17.1% remains anonymous/JIT/unresolved.
Collector and allocator stacks take precedence over their owned callers; these
categories do not overlap. Inlining/unresolved leaves limit precision. The large
`runtime.usleep` category in Go pprof is not reproduced as an on-CPU hotspot;
it must not be promoted to a measured GC CPU fraction. This single trace is
not a paired attribution against the candidate or competitor.

An independent approved-return probe separates raw Go/cgo, full Context entry
and batches of 64 identical C entries per cgo call. Three observations each give
median **73.37 / 79.73 / 50.53 ns per entry**, all with zero allocations.
106 full Context entries correspond to about 8.45 µs in this synthetic setup.
The single-goroutine hot probe omits request GC/contention, binding preparation
and guest work; C batching also changes loop/cache behavior. It does not bound
real request cost. It does show why the cgo sample share cannot be treated as
entirely removable overhead or justify a bridge rewrite by itself.

The next focus is checked Canonical ABI conversion, copying and ownership
transactions, plus VM/host state transitions. Their multiple passes and dynamic
intermediates need separate measurement and compiled-plan evaluation; native
driver/guest work must also improve. The Running trace assigns only about 22.4%
to allocator plus collector paths, so treating the entire gap as inevitable GC
cost is unsupported. Small frame shortcuts and allocation counts alone do not
establish progress toward parity.

## Validation and integration

The retained implementation passes all 32 Go package groups and all 31
executable-package coverage gates, minimum **90.87%**. Full execute/VM/component
race tests, whole-module vet and no-cgo build pass. Production LLGo
`3d92a58 -O3 -lto=full` executes four complete affected macOS packages; Docker
Linux amd64 passes eight affected Go/LLGo groups. Tuple snapshots, vector bits,
reference roots, aliased result handoff, reentry and bounded idle retirement
have regression tests. All 32 test packages remain assigned once across eight
LLGo CI shards.

The rejected expansion separately passed full Go coverage, focused race tests,
five complete macOS LLGo packages and ten Docker groups. Its first structural
counter test omitted native arena reservation; corrected fixture setup proves
guest batching. That initial assertion is not qualifying evidence. Returning to
identical retained source does not require another duplicate semantic campaign.

[Final comparison data](data/performance-go-frame-plans-2026-11-11.json) and
[compressed evidence](evidence/performance-go-frame-plans-2026-11-11.jsonl.gz)
publish all process counters, controls, phase arrays, probe observations and
checksums. Per-request latency arrays remain in the checksummed local raw event.
The branch awaits its contribution slot while
[LLGo #3057](https://github.com/xgo-dev/llgo/pull/3057) is active. Full hosted
final-head checks/review and performance/release acceptance remain open.
