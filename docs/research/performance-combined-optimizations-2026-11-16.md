# Combined host-storage, native-entry and operand measurements

The Go 16-byte complete P3 request improves **3.5% in CPU and 4.6% in
throughput** against the source baseline, with resolving controls. Its managed
allocation count falls from about **616 to 521 per request**. Competitive parity
remains open: sustained general native calls take approximately **2.2 times**
the reference time, and P3 CPU observations still favor Wasmtime. Several
interpreter regressions and the adverse default-GC SDK observation remain
recorded below. This event does not close a milestone or release gate.

## Sources and fixed protocol

The source baseline is **e9d9e1b4**, before PRs
[#188](https://github.com/cpunion/wasmvm/pull/188),
[#189](https://github.com/cpunion/wasmvm/pull/189),
[#190](https://github.com/cpunion/wasmvm/pull/190) and
[#191](https://github.com/cpunion/wasmvm/pull/191). Measured source is
**477af3be**; all Go/C/header/assembly/module inputs are identical to CI head
**aa935e3e** and merged main **0b66382e**. The final source difference is only
the codegen shard's corpus prerequisite. One existing checkout and changed-file
overlays provide the baseline; no source worktree/archive copies were added.

Both source revisions use Go **1.27.3**, or LLGo **1b44c645** with
**`-O3 -lto=full`**, LLVM **22.1.8**, and default production GC. Fixed-engine
old-SDK controls use **5860fb6f**; their engine Go/C sources equal measured head.
The SDK change is a separate baseline event, not a source optimization.
References are wazero **1.12.0**, Wasmtime **48.0.1**, Wasmer **7.3.0** and
WasmEdge **0.17.1**. Binaries, fixtures and libraries have recorded hashes.
The host is an Apple M4 Max, macOS arm64, 16 cores and 128 GiB RAM. This shared
developer machine has no affinity/frequency pinning; external load is recorded.
No owned build, test, profile, edit, pruning or analysis overlaps timed windows.

Core retains **33 cases**, ten fixed fresh-process blocks, symmetric same-binary
A/A participants, two warmups and seven individually interleaved measured calls:
**9,660 processes**. The original 22 cases, four sustained GC cases, six
sustained Threads cases and general-call case remain separately named. Allocation
uses process means including pauses; other cases use process medians. Twelve
known unsupported original GC combinations are explicit and also apply to the
corresponding sustained fixtures. Other failures are never silently skipped.

P3 retains **140 processes per payload**, seven roles with symmetric A/A,
ten fixed blocks, four workers/sixteen clients, twenty interleaved 512-request
windows and a separate flush. Every response verifies status, headers, complete
body, trailers and CRC, with guest completion joined. Process CPU/instruction
counters include the driver and checks. Mach ticks are converted using the
calibrated 125/3 ns timebase. Native Go is only an HTTP host-cost oracle without
Component/WASI. Managed allocation counters omit foreign Rust/C/JIT allocation
and cannot rank total engine allocation.

Startup, preparation, first call, warmups, measured execution and cleanup stay
separate. Raw phases remain available; no phase subtraction is used. Ratios use
paired block geometric means, 20,000 fixed-seed block bootstrap resamples and
individual 95% intervals. Both relevant A/A intervals must fit [0.975,1.025].
These marginal intervals are diagnostic, not simultaneous portfolio acceptance.
No experiment is extended or pooled after inspecting results.

## Complete P3 requests

| Source head/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| Go CPU | **0.965 [0.959,0.971]**, controls pass | 0.961 [0.957,0.966], controls fail |
| Go instructions | **0.973 [0.971,0.975]**, controls pass | **0.979 [0.978,0.981]**, controls pass |
| Go throughput | **1.046 [1.039,1.052]**, controls pass | 1.047 [1.036,1.058], controls fail |
| Go allocations/request | 616.00 → **521.25** | 667.00 → **560.25** |
| LLGo CPU | 0.970 [0.955,0.990], controls fail | 0.997 [0.960,1.036], controls fail |
| LLGo allocations/request | 928.00 → **908.25** | 1099.00 → **1091.25** |

All allocation-count controls resolve. Lower allocations alone do not establish
CPU acceptance. The observed head medians below include unresolved reference
controls: **these are advisory observations, not accepted precise rankings**.

| Complete request | wasmvm Go | wasmvm LLGo | Wasmtime |
| --- | ---: | ---: | ---: |
| 16 B CPU, us/request | 95.78 | 272.84 | 59.66 |
| 16 B throughput, requests/s | 40,661 | 14,040 | 41,667 |
| 64 KiB CPU, us/request | 124.47 | 328.60 | 65.55 |
| 64 KiB throughput, requests/s | 31,298 | 13,573 | 42,086 |

Paired Go/Wasmtime CPU point ratios are 1.635 and 1.992; their controls fail.
Native Go CPU medians are 4.18/112.58 us with a different host-only ABI and
body-collection policy. It is not a Component/WASI competitor and its time is
not subtracted from the engines.

## Core compiler and interpreter

Warm ns/source iteration; SHA units are ns/message block. Rows are observations
unless a comparison is explicitly qualified below. Complete controls and all
33 cases remain in the machine-readable report.

| Native case | Go | LLGo | wazero | Wasmtime | Wasmer | Unchecked C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Integer loop | 1.657 | 1.630 | 1.998 | 1.631 | 1.624 | 1.626 |
| Memory load/store | 0.484 | 0.465 | 1.118 | 0.510 | 0.568 | 0.296 |
| SIMD lanes | 5.233 | 5.177 | 6.434 | 6.095 | 6.101 | 4.704 |
| Clang SHA | 241.36 | 236.51 | 373.21 | 199.98 | 200.56 | 172.54 |
| Rust SHA | 244.61 | 241.25 | 342.32 | 194.30 | 196.77 | 172.09 |
| Sustained GC field access | 1.388 | 1.345 | Unsupported | 1.089 | Unsupported | Unavailable |
| Sustained GC allocation, including pauses | 33.05 | 30.24 | Unsupported | 51.69 | Unsupported | Unavailable |
| Sustained shared memory | 1.598 | 1.406 | 1.497 | 1.172 | 1.232 | Unavailable |
| General calls, two affine steps | 4.837 | 4.803 | 2.244 | 2.146 | 2.135 | 2.136 |

General-call Go ratios resolve: **2.177x wazero, 2.257x Wasmtime, 2.278x
Wasmer, 2.259x C**. This closed native kernel has no host allocation or GC
that could explain the gap. The earlier [instruction audit](performance-native-frame-registers-2026-11-15.md)
identifies repeated full frame/value/metadata work versus Cranelift's register
arguments and small call frame. Native call ABI/state representation remains
the largest calibrated common compiler gap in this set.

Other resolving Go comparisons include memory load/store at **0.944x Wasmtime**,
SIMD lanes at **0.863x Wasmtime**, and sustained GC field access at **1.270x
Wasmtime**. Integer execution is close but still **1.014x Wasmtime**, with a
resolving interval [1.005,1.023]. SHA points remain about 20–27% slower than
Wasmtime, but those controls fail. amd64 operand changes need native amd64
timing; local Docker is semantic evidence only.

Go interpreter resolving ratios include integer **0.381x wazero /0.637x
WasmEdge**, load/store **0.452x/0.716x**, and Rust SHA **0.443x/0.745x**.
There are also deficits: short shared memory is **1.220x wazero**, sustained
atomic RMW **1.591x WasmEdge**, and sustained wait mismatch **1.299x WasmEdge**.
No overall winner is inferred by averaging unlike contracts or duplicated
sustained cases.

## Every adverse observation remains actionable

The source campaign has marginal adverse findings with resolving controls:
Go interpreter function calls +1.29%, float64 +1.65%, short GC indirect access
+2.00%, Rust SHA +0.74%; LLGo interpreter integer +2.07%, load/store +2.72%,
dependent address +2.26%, sustained GC access +1.24%, atomic RMW +2.92%, wait
timeout +2.55% and general calls +1.61%. LLGo general native calls are +0.33%,
with the interval only just excluding one. All other adverse/unresolved rows
are also retained. Multiple marginal comparisons are not a family-wise causal
claim and none of these rows is relabeled a pass.

An independent **144-process**, three-kernel counter diagnostic narrows the
interpreter mechanism. Go/LLGo main loops retain 2,528/2,082 instructions.
The Go first 1,852 instructions are byte-identical; inspected differences
elsewhere are relocations/data addresses. LLGo retains the same loop operation
sequence with moved addresses. Dynamic instruction ratios stay within about
0.1% of one in the diagnostic, with resolving instruction controls. However,
LLGo integer CPI rises **2.28% [1.10%,3.43%]**, Go float64 **1.71%
[1.28%,2.16%]**, and LLGo float64 **4.76% [3.43%,6.34%]**, with resolving CPI
controls. These cases show increased cycles per instruction rather than an
extra dispatch/allocation workload. Code/data placement and cache/branch effects
remain candidate causes; the exact microarchitectural cause and the other
adverse rows are not yet isolated. No padding search or favorable-run replacement
was used.

The new-SDK default-GC 64 KiB point is **1.055 CPU /1.046 instructions** versus
the old SDK, with failed controls. Median measured LLGo collections are
**48.5 versus 33**, with wide overlapping distributions. A separate fixed
**40-process** bounded no-GC diagnostic requires zero measured collections and
does not modify production settings. New/old SDK CPU becomes **0.996
[0.989,1.005]** and instructions **0.99936 [0.99856,1.00018]**, with resolving
controls. This supports collection sensitivity, not a proven new SIMD/codegen
bug or an accepted default-GC performance result. Conservative root retention,
heap sizing and remaining generated-code/layout effects stay open; the existing
[LLGo collector investigation](https://github.com/xgo-dev/llgo/issues/3106)
is not duplicated.

A separate Go CPU profile of current 16 B P3 records about 19.6% cumulative
under `runtime.cgocall`, 10.2% under `mallocgc`, and substantial runtime
locking/yielding. Cgo samples include guest native execution and are not an
exclusive bridge percentage; cumulative percentages overlap and must not be
added. Allocation stacks still include fresh callback guards, resources,
source/task state, codecs and cancellation contexts. Profiled timings do not
participate in the comparison.

## Integration and next implementation order

Final [CI](https://github.com/cpunion/wasmvm/actions/runs/43836231418) passes
all 27 checks, Go/LLGo on Linux amd64/macOS arm64, and 31 coverage gates above
90%. Eighteen LLGo archives plus core coverage are hash-verified. macOS HTTP
first timed out after every test passed; its isolated rerun passes at 711 seconds.
Moving codegen to the component shard preserves every package and required
corpus, but the maximum-job **5–10 minute target remains unmet**. Link time,
not the roughly 42-second HTTP test execution, is the main latency contributor.

The [host-storage implementation](performance-host-storage-2026-11-16.md)
completes the first ordered increment; it does not complete the larger ABI work.
Continue in this order:

1. Reduce repeated native call metadata/value materialization with a shared
   private ABI and compiler liveness plans. Keep the common interpreter frame
   format at authenticated observable exits, full 128-bit values, quotas,
   roots, cancellation, every cold PC and both architectures. A handful of
   removed locks does not close the measured 2.2x general-call gap.
2. Add indexed, typed built-in canonical/WASI adapters that share the existing
   type/layout/ownership plans. Keep dynamic public callbacks and actual
   pending/suspending ownership. Avoid HTTP-specific sequence fusion.
3. Address interpreter CPI/layout sensitivity and the remaining shared-memory,
   atomic, GC access and SHA code differences using instruction/counter evidence.
   Keep LLGo collection and synchronization investigations separate from Go's
   common architectural work. Isolate a compiler bug before opening another issue.
4. Reassess complete ready/pending requests and both native target portfolios
   after that implementation batch. Preserve all earlier cohorts and failures;
   controlled milestone, exhaustive conformance and release acceptance remain open.

The rejected direct Go assembly bridge remains rejected: concurrent native entry
and collection deadlocked. Retain the supported C transition until scheduler/GC
cooperation or a complete execution bound is implemented and validated.

Evidence: [all results, controls, provenance and phase summaries](data/combined-optimizations-2026-11-16.json),
[compact per-process records](data/combined-optimizations-2026-11-16-raw.jsonl.gz),
and [frozen protocols, diagnostics and disassembly](data/combined-optimizations-2026-11-16-protocol.tar.gz).
Only unused per-request timing vectors are projected to counts/extrema/hashes;
no latency-tail claim relies on that projection. Full originals remain local
with recorded hashes.
