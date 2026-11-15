# Component call setup

This change removes a heap-allocated call-exit closure and reduces the canonical
memory adapter from a complete `Extern` copy to a pointer to its immutable
namespace definition. Callers, resource ownership, budgets and memory checks
remain independent for each invocation. No runtime synchronization change ships.

## Method

Base **da68bf24**, implementation **26b6d6a6**; reused base binaries identify
**1385093b**, whose complete Go/module sources match the merged base. Host:
Apple M4 Max/macOS arm64. Go **1.27.3**, LLGo main **0539ea6e**, LLVM **22.1.8**,
LLGo **`-O3 -lto=full`**, default production GC. The independent P3 HTTP provider
is frozen Wasmtime **48.0.1**. Wazero and Wasmer lack equivalent P3 worlds in
this comparison; previous general Core measurements remain separate.

Each Go payload has a fixed ten-block, sixty-process cohort: base, head and
reference with A/A aliases. Each LLGo payload has ten blocks/eighty processes,
adding Go head as a separate anchor. Four Stores and sixteen clients execute
twenty individually interleaved 512-request batches and a separate flush.
Every response checks body, status, headers, trailers and CRC. Startup,
compilation, instantiation, first call, two warmups and warm requests are
separate. No owned measurement-host builds, tests, profiles, analysis, edits or
pruning overlap timed cohorts. Other user/desktop activity is recorded.

Ratios use block-paired geometric means and individual bootstrap 95% intervals.
A/A resolves only when its whole interval lies inside 0.975–1.025. Cohorts are
neither extended nor pooled. These targeted requests do not close portfolio,
cross-platform or release performance acceptance.

## Go results

| Per-request measurement | 16 B base → head | 64 KiB base → head |
| --- | ---: | ---: |
| Host allocation events | 663 → **655** | 714 → **706** |
| Host allocation bytes | 54,233 → **53,209** | 123,881 → **122,089** |
| Warm CPU head/base [95% interval] | **0.98992 [0.97994, 0.99754]** | 0.99659 [0.99057, 1.00394] |
| Instructions head/base [95% interval] | **0.99449 [0.99282, 0.99612]** | **0.99465 [0.99224, 0.99660]** |

Own CPU, instruction and allocation controls resolve at both sizes. The 16 B
event supports a small CPU reduction; the 64 KiB CPU interval includes no
change. Throughput gains are unresolved. Host counters omit foreign Rust/C,
JIT and linear-memory allocations.

Go head/Wasmtime CPU is 1.89644 [1.77634, 2.07098] at 16 B with unresolved
reference control; 64 KiB is **2.16483 [2.04013, 2.32588]**, with both controls
resolving. The remaining complete-request gap is substantial. These events do
not compare interpreters or establish competitive parity.

## LLGo results

| Per-request measurement | 16 B base → head | 64 KiB base → head |
| --- | ---: | ---: |
| Host allocation events | 981 → **973** | 1153 → **1145** |
| Host allocation bytes | 72,960 → **72,160** | 213,489 → **212,113** |
| Warm CPU head/base [95% interval] | 0.99836 [0.98179, 1.01793] | 0.98792 [0.95649, 1.02069] |
| Instructions head/base [95% interval] | 1.00518 [0.99106, 1.02220] | 0.99682 [0.96040, 1.03374] |

Allocation controls resolve. CPU, instruction and throughput controls do not;
no LLGo speedup or regression is established. The adverse 16 B instruction point
is retained: base A/A alone is 1.03118 [1.00318, 1.05246], head A/A 0.98267
[0.95116, 1.01230]. The per-process collection and user/system CPU audit retains
the variability; it does not establish its exact GC, scheduling or layout cause.

LLGo/Go CPU points are **2.44x/2.25x**, LLGo/Wasmtime **4.67x/4.85x**. LLGo
and reference CPU controls fail, so these are broad-gap observations. The
64 KiB checker also includes the extra copy already reported in
[LLGo #3113](https://github.com/xgo-dev/llgo/issues/3113); it must not be attributed
to guest execution. Runtime handoff/synchronization is being handled separately.

## Rejected ownership-walk experiments

Two earlier candidates cached resource-presence flags in canonical layouts to
skip ownership walks. Both retain their complete fixed 60-process, 16 B cohorts:

- Alpha **eef5ad61** adds metadata lookups on aggregate values. Instructions
  increase **0.694% [0.403%, 0.947%]** with resolving own controls. CPU ratio is
  1.00884 [0.99814, 1.01957], also with resolving own controls. Rejected.
- Beta **4175286d** restricts lookups to nonempty lists/maps. CPU is
  1.00234 [0.99266, 1.01227], instructions 1.00299 [0.99898, 1.00787]; own CPU
  controls fail. No benefit is established. Rejected without a larger-payload
  extension. The copied plan incorrectly names alpha; an explicit erratum
  identifies beta, whose binary hash and build identity are correct.

The likely tradeoff is extra lookup/classification work on short shapes; exact
lookup versus code-layout shares are not isolated. Neither experiment ships.
An untimed allocation profile of the base locates the call-exit closure and
full-`Extern` adapter boxes. It includes sparse setup samples and is not an
exclusive CPU decomposition or an acceptance timing measurement.

## Validation and evidence

The regression test forces namespace storage growth and GC while retaining a
memory binding, checks byte/string reads and writes, rejects bounds violations,
and verifies that the binding expires with its Caller. Reentry exhaustion must
unwind call depth to zero. Existing ownership, cleanup, cancellation and async
tests continue to exercise both paths.

All 32 Go test packages pass and all 31 executable packages meet 90% statement
coverage. Affected component/P3 race tests, both extra Go benchmark modules,
and LLGo component/P3 execution pass locally. Exact final-head dual-platform CI
remains required before merge.

The [machine-readable report](data/performance-component-call-setup-2026-11-15.json)
retains source/binary identities, every cohort/control, adverse outcomes,
lifecycle phases, raw request/GC/allocation counters, load snapshots and test
logs. Full local records remain under `artifacts/protocol-cost-20261112/`.
