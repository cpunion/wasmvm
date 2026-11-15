# Latest LLGo main and complete P3 request costs

The largest observed application gap remains complete P3 HTTP, especially under
LLGo. This event updates the production compiler pin and removes completed-work
bookkeeping. **Competitive parity and release acceptance remain open.**

## Implementation and qualification

Base engine **35094fd4**, implementation **9c271a3e**, qualified pin commit
**5ab3bcce**. The latter changes compiler pins only; its Go/module sources equal
the implementation. Final source comparisons use Go **1.27.3**, LLGo main
**5860fb6f**, LLVM **22.1.8**, **`-O3 -lto=full`**, and default production GC.
The initial compiler-only and runtime diagnostics separately use **113daaad**
against the retained **0539ea6e** baseline. Neither diagnostic compiler event is
pooled with the final source comparison.

Three shared changes ship:

- A synchronously completed endpoint read skips cancellation registrations.
  Its resolver still checks operation, callback and Store cancellation before
  accepting data. Pending reads retain both independent cancellation watchers.
- Completion subscription returns readiness under the same mutex. Already
  completed work needs no redundant wake send or second readiness lookup;
  later producers still notify the scheduler without a lost-wake window.
- Complete buffered HTTP replies retain their body through guest task cleanup,
  then release it directly. They no longer create streaming notification
  channels, a streaming body wrapper or its extra cancellation registration.
  Streaming replies retain their existing lifetime and join protocol.

New tests cover cancellation after data readiness but before handoff, clean
resource accounting, and concurrent subscription versus resolve/cancel. Existing
body, trailer, quota, timeout, streaming, panic and shared-world tests remain.

## Fixed comparisons

Apple M4 Max, 16 cores, macOS arm64. Each final payload/compiler cohort has ten
fixed blocks, twenty individually interleaved 512-request batches plus a flush,
four Stores and sixteen clients. Go has 60 processes per payload; LLGo has 80,
including a Go anchor. Every participant has a same-binary A/A alias. Each full
reply checks status, headers, trailers and every body byte, returning the
independently verified payload CRC token. Snapshot input and ordinary independently
owned buffered output are explicit; guest join stays inside completion.

Startup, validation/compilation, instantiation, first request, two warmups,
steady batches and cleanup remain separate. CPU/instructions include the full
driver and checks, not only guest execution. Host allocation counters omit
foreign Rust/C/JIT allocations and cannot rank total allocation against Wasmtime.
Ratios use paired block geometric means and marginal bootstrap 95% intervals;
the entire A/A interval must lie in [0.975, 1.025]. No fixed cohort is extended
or pooled. Other user build/GC/desktop activity is recorded in load snapshots;
no owned build, test, profile, edit, analysis or pruning overlaps timed cohorts.
These are scoped advisory events, not controlled portfolio acceptance.

| Final source change | Go 16 B | Go 64 KiB | LLGo 16 B | LLGo 64 KiB |
| --- | ---: | ---: | ---: | ---: |
| Allocations/request, base → head | 652 → **645** | 703 → **696** | 970 → **962** | 1141 → **1133** |
| Accounted bytes/request saved | 480 | 480 | 560 | 560 |
| CPU head/base | **0.99098** | **0.99092** | 0.98902 | 0.98544 |
| CPU 95% interval | [0.98622, 0.99582] | [0.98426, 0.99745] | [0.93936, 1.03718] | [0.90860, 1.06637] |
| Instructions head/base | **0.99365** | **0.99510** | 0.99197 | 0.98984 |

Allocation controls resolve. Both Go CPU and instruction source controls resolve:
CPU falls about **0.9%**, instructions **0.63% / 0.49%**. LLGo CPU/instruction
controls fail; no LLGo speedup is established. Go 16 B throughput improves
0.71% with resolving controls; 64 KiB throughput and both LLGo throughput gains
remain unresolved.

| Complete-request CPU comparison | 16 B | 64 KiB |
| --- | ---: | ---: |
| Go / Wasmtime 48.0.1 | 2.53x | 2.49x |
| LLGo / Go, same LLGo cohort | 3.35x | 3.03x |
| LLGo / Wasmtime 48.0.1 | 8.76x | 8.15x |

These CPU ratios are **broad-gap observations**: reference CPU controls fail,
as do LLGo CPU controls. Resolving instruction comparisons put Go at
**3.84x / 3.61x** Wasmtime. The remaining generic work is substantial even under
Go; GC is not the sole explanation. Wazero and Wasmer provide no equivalent
complete P3 world in this harness. Earlier Core comparisons remain separately
dated; this event makes no new Core or Linux performance claim.

## Largest cost evidence and compiler regressions

Separate warm-only CPU Profiler captures check 131,072 requests each at fixed
engine source under old 0539ea6e and initial latest 113daaad. Collector/allocator
leaves account for **63.6% / 68.6% of cycle sample weight**, mainly
`GC_mark_from`; the new channel park/spin leaf accounts for **1.9%**. These
statistical cycle weights are not exact exclusive CPU fractions. Cumulative
stacks overlap and must not be subtracted from process CPU.

A bounded, identical-binary 113daaad collector experiment separately compares
defaults, one marker, free-space divisor one and disabled GC. CPU ratios to
defaults are **0.7974 / 0.8964 / 0.8556**. Default CPU controls fail; precise
gain claims and a production tuning recommendation remain unqualified.
The no-GC windows assert zero collections, grow memory until exit, and are
diagnostic only. They do not establish that 14% or 69% is an exact GC share.
Additional instrumented 5860fb6f logs show **15 helper marker threads** and
roughly **40 application thread stacks** at collection, with about 10 MiB of
pointer-bearing live heap. Logged collection durations are perturbed diagnostics.
[BDWGC's marker controls](https://github.com/bdwgc/bdwgc/blob/master/include/gc/gc.h)
and [parallel collector design](https://www.hboehm.info/gc/scale.html) explain the
independent marker population. LLGo records `GOMAXPROCS` without enforcing a
native-thread parallelism limit. Existing
[LLGo #3106](https://github.com/xgo-dev/llgo/issues/3106) already tracks the
collection/heap sensitivity; no duplicate issue or new runtime patch is proposed.

The initial latest-main event removes one 64 KiB allocation/copy, consistent
with the merged `bytes.Equal` fix. Nevertheless 16 B CPU is adverse at **1.1117
[1.0513, 1.1782]** versus 0539ea6e; own CPU controls fail. It remains a regression
signal, not an accepted exact 11% regression. A separate source ablation retains
new unfair locks while changing 8192 channel spins to zero. Instructions fall
about **14% / 16%**, but CPU controls fail and no stable throughput benefit is
established. This isolates excess spin instructions, not the complete CPU cause.
The diagnostic SDK overlay is absent from production.

Ordinary native-Go HTTP application controls use the same bounded service driver
and full echo checks, with no VM or P3 contract. Latest/old LLGo CPU points are
**0.607 / 0.892**, while latest LLGo/Go remains **13.0x / 6.71x**. Controls fail:
the merged synchronization/copy changes help the observed control but do not
establish precise gains or explain the entire P3 gap. Native control costs must
not be subtracted from P3 costs; their allocation and copy contracts differ.

## Validation and evidence

All 32 root Go packages and the per-package 90% gate pass before the subscription
refinement. Fresh final affected-package coverage is **92.46% component** and
**91.26% P3**; final component race and complete P3 measurement-module tests pass.
Actual LLGo full-LTO complete guests pass throughout the fixed cohorts. Final-head
Go/LLGo semantics on Linux amd64 and macOS arm64 remain required before merge.

The [report](data/performance-latest-main-p3-2026-11-16.json),
[616 primary process records](data/performance-latest-main-p3-2026-11-16.evidence.json.gz)
and [exact protocols and diagnostics](data/performance-latest-main-p3-2026-11-16-protocol.tar.gz)
retain every adverse point/control, compiler and source identity, allocation/GC
counter, lifecycle phase, rejected spin overlay, profile summary/XML hash and
validation log. Only unused per-request latency arrays are compacted to
count/extrema/sum/hash; full originals and raw traces stay local under
`artifacts/latest-main-gap-20261116/`. No full repository copies are added.

Next work remains the generic canonical/host boundary instruction gap, native
call frames, SHA scheduling, and the independently tracked LLGo collector and
native scheduling costs. The present changes reduce completed-work overhead;
they do not complete the broader performance goal.
