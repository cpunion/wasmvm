# Shared loop versions and invariant literals: 2026-11-05

**Implemented on the contribution branch; local Go/LLGo semantics pass.
Elapsed-time acceptance remains open: none of the 22 workload calibrations resolve.**
Engine source: `68b2269cc662302bd77eb35cc93cfe415fc99173`.
Baseline: `3df460408a12236531ff60ac2ffd17c64177d3b8` (engine `00bb578`).
The [preceding nonresident-value event](performance-loop-nonresident-values-2026-11-05.md)
and [hosted scratch-register event](performance-loop-scratch-registers-hosted-2026-11-05.md)
remain separate source/host cohorts; their ratios must not be multiplied or pooled.

## Mechanism and execution

One bounded, nontrapping-loop version graph now serves register coalescing and
invariant literal materialization. A slot qualifies only when its old entry word
is unread and every write assigns the same typed literal bits. Shared physical
homes, mixed literals, copy/arithmetic writes, header polls, possible traps and
unproved memory accesses retain ordinary emission. Six literal classes cover
I32/I64/F32/F64/V128/null references. Definitions move after entry guards and
coalesced live-in setup, before the private header. Only the private body elides
them; all original cold PCs and failed guards keep their original operations.
Complete frame words, roots, existing polls and precise fuel retain their contracts.

The complete affected codegen package executes successfully with Go 1.27.3 and
LLGo `3571c25` (`-O3 -lto=full`) on macOS arm64. Its Go statement coverage is
**96.44% (7583/7863)**; this is one-package coverage. Models compare every cold
PC, complete frame state, valid/OOB accesses and forced private-loop polls on
the executing architecture, while compiling both targets. New literal cases
retain negative-zero, signaling-NaN and nontrivial vector bits.

The unchanged 92 P2 and 60 P3 upstream consumers execute on interpreter/native
with both compilers: **608 guest/backend/compiler invocations**, with no test
skips. Each compiler reports 388 passing nodes including parents, interactions
and negative controls; those nodes are not 388 different guests. Broader WASI-world
conformance remains open.

The pure x86 compile diagnostic uses feature mask 7 and the exact original
private/shared memory fixtures. Both hot loops change **16 → 13 instructions**,
retain **0 frame references and 1 control load**, and eliminate the same four
proved bounds checks. Complete function bytes change 2333 → 2328 and
2477 → 2472. All cold spans and metadata remain retained. This arm64-hosted
diagnostic neither publishes nor executes x86 code and makes no timing claim.

## Local competitive event

Apple M4 Max, macOS arm64, GOMAXPROCS=1; five fresh blocks, two warmups and seven
serially interleaved calls per runtime/block. Exact source, toolchain, fixture,
reference executable and SDK library hashes are retained. Fresh own base/head
Go/LLGo binaries use identical flags. References remain wazero 1.12.0, Wasmtime
48.0.1, Wasmer 7.3.0 and WasmEdge 0.17.1 (interpreter). Native C uses LLVM 22.1.8
`-O3` and the existing same-count checksum contract.

The sole measurement adaptation removes Linux affinity calls and records macOS
CPU/scheduler metadata. It does not change samples, order, oracles, policies or
phase boundaries. CPU frequency and neighbor contention are uncontrolled; the
recorded host loads are retained. There are no timing retries, alignment/ISA/flag
sweeps, timer subtractions or cross-host aggregations.

Independent recomputation validates **22 cases, 362 alternative argument/runtime
configurations and 12 explicit unsupported GC configurations**. Every returned
checksum, phase shape and recorded calibration is checked. The three unsupported
references reject GC explicitly; no missing reference is treated as a win.
Of 88 individual ±2.5% marginal A/A intervals, 86 remain unresolved.
Every workload fails at least one of its four intervals.
**Resolved workloads: 0/22.** All ratios below describe observed complete API
intervals and remain inconclusive, including apparent improvements.

### Source changes

Elapsed head/base ratio; lower is better. Every row is inconclusive.

| Workload | Go native | LLGo native | Go interpreter | LLGo interpreter |
| --- | ---: | ---: | ---: | ---: |
| core/call-control | 1.181× | 1.000× | 0.956× | 0.836× |
| core/integer-loop | 1.039× | 1.011× | 0.992× | 1.080× |
| core/function-calls | 1.001× | 0.983× | 0.994× | 1.000× |
| core/memory-load-store | 0.831× | 0.714× | 1.033× | 1.020× |
| memory/narrow-unaligned | 0.929× | 0.606× | 1.013× | 1.020× |
| memory/dependent-address | 0.940× | 1.097× | 1.032× | 0.992× |
| memory/float64-unaligned | 0.960× | 0.922× | 0.997× | 1.009× |
| runtime/simd-memory-full | 0.871× | 0.983× | 1.002× | 1.161× |
| runtime/simd-memory-widen | 1.038× | 1.022× | 1.032× | 0.999× |
| runtime/simd-memory-lanes | 0.955× | 1.017× | 1.037× | 1.038× |
| gc/i31-roundtrip | 0.988× | 0.981× | 1.086× | 1.237× |
| gc/struct-access | 0.891× | 0.995× | 1.015× | 0.593× |
| gc/struct-allocation | 0.960× | 1.080× | 1.154× | 0.969× |
| gc/struct-indirect-access | 1.053× | 1.075× | 0.886× | 1.235× |
| threads/private-memory | 0.948× | 0.951× | 0.983× | 1.024× |
| threads/shared-memory | 0.938× | 0.917× | 1.019× | 0.976× |
| threads/atomic-rmw | 1.021× | 1.025× | 1.109× | 0.970× |
| threads/notify-empty | 1.021× | 0.691× | 0.984× | 0.843× |
| threads/wait-mismatch | 0.894× | 0.824× | 0.921× | 0.737× |
| threads/wait-timeout | 0.808× | 0.905× | 1.098× | 0.905× |
| applications/sha256-clang | 1.241× | 0.936× | 1.022× | 0.951× |
| applications/sha256-rust | 1.210× | 1.010× | 1.048× | 1.044× |

Observed adverse outcomes include Go native SHA +21–24%, LLGo native dependent
addresses +9.7%, LLGo interpreter SIMD full +16.1%, i31 +23.7% and indirect
struct access +23.5%. The LLGo native memory/narrow observations improve by
28.6%/39.4%; they also remain unresolved and cannot offset adverse rows.

### Competitor and compiler gaps

Elapsed ratios; lower is better. All rows retain the failed A/A qualification.
Native rows use complete public calls; interpreter ratios use complete
interpreter calls. The data file also retains every source-iteration median.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 0.649× | 0.898× | 4.472× | 2.321× | 1.302× | 1.990× | 0.890× | unavailable | inconclusive |
| core/integer-loop | 1.031× | 0.818× | 0.923× | 1.105× | 1.055× | 0.354× | 0.590× | 1.017× | inconclusive |
| core/function-calls | 1.048× | 0.864× | 0.601× | 0.919× | 1.070× | 0.489× | 0.748× | 1.070× | inconclusive |
| core/memory-load-store | 0.963× | 0.840× | 0.452× | 0.919× | 0.776× | 0.429× | 0.690× | 1.151× | inconclusive |
| memory/narrow-unaligned | 0.865× | 0.858× | 0.602× | 0.970× | 0.961× | 0.442× | 0.703× | 1.513× | inconclusive |
| memory/dependent-address | 1.122× | 0.872× | 0.567× | 0.907× | 0.974× | 0.427× | 0.744× | 1.427× | inconclusive |
| memory/float64-unaligned | 0.947× | 0.849× | 0.678× | 1.057× | 0.957× | 0.477× | 0.732× | 1.081× | inconclusive |
| runtime/simd-memory-full | 1.017× | 0.911× | 0.923× | 0.867× | 0.992× | 0.460× | 0.891× | 0.913× | inconclusive |
| runtime/simd-memory-widen | 0.954× | 0.879× | 0.797× | 0.994× | 0.969× | 0.416× | 0.942× | 0.922× | inconclusive |
| runtime/simd-memory-lanes | 1.021× | 0.856× | 0.778× | 0.888× | 0.845× | 0.468× | 0.945× | 1.066× | inconclusive |
| gc/i31-roundtrip | 0.790× | 0.843× | unsupported | 0.972× | unsupported | unsupported | 0.901× | unavailable | inconclusive |
| gc/struct-access | 0.856× | 0.798× | unsupported | 1.741× | unsupported | unsupported | 0.521× | unavailable | inconclusive |
| gc/struct-allocation | 1.166× | 1.238× | unsupported | 0.395× | unsupported | unsupported | 1.434× | unavailable | inconclusive |
| gc/struct-indirect-access | 0.966× | 1.131× | unsupported | 0.898× | unsupported | unsupported | 0.985× | unavailable | inconclusive |
| threads/private-memory | 0.903× | 0.933× | 1.532× | 1.360× | 1.099× | 0.703× | 0.746× | unavailable | inconclusive |
| threads/shared-memory | 0.854× | 0.444× | 1.438× | 1.340× | 1.028× | 0.688× | 1.005× | unavailable | inconclusive |
| threads/atomic-rmw | 0.897× | 0.550× | 1.469× | 1.355× | 0.531× | 0.888× | 0.884× | unavailable | inconclusive |
| threads/notify-empty | 0.619× | 0.724× | 0.054× | 0.152× | 0.175× | 0.565× | 0.528× | unavailable | inconclusive |
| threads/wait-mismatch | 0.714× | 0.495× | 0.043× | 0.169× | 0.412× | 0.523× | 0.919× | unavailable | inconclusive |
| threads/wait-timeout | 0.675× | 0.522× | 0.005× | 0.039× | 0.009× | 0.096× | 0.368× | unavailable | inconclusive |
| applications/sha256-clang | 0.741× | 0.834× | 0.621× | 0.885× | 0.999× | 0.431× | 0.662× | 1.204× | inconclusive |
| applications/sha256-rust | 0.811× | 0.863× | 0.602× | 1.123× | 1.012× | 0.434× | 0.723× | 1.190× | inconclusive |

### Preparation remains separate

Median validate/compile phase, **microseconds**. Warm-call ratios above contain
no phase subtraction. Loading, engine setup, instantiation, export lookup, first
call, cleanup, process readiness and process wall time remain separate fields
for every runtime/case in the data and raw evidence.

| SHA producer | Go native | LLGo native | wazero JIT | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| clang | 2743.17 | 7337.75 | 689.25 | 2970.79 | 3477.54 |
| rust | 2685.62 | 6497.46 | 445.62 | 2068.71 | 2826.83 |

LLGo preparation is observed at 2.67×/2.42× Go for Clang/Rust SHA. These remain
uncontrolled phase observations, not an established LLGo compiler defect.
The five sequential measurement groups take 159/400/61/88/153 seconds; these
wall times describe collection, not guest execution or hosted CI duration.

## Actual published code audit

Output-only diagnostics archive each exact source and change only the native
publication hook, with before/after source hashes and the patch retained.
All 22 head Go/LLGo executable images and entry-offset lists match **byte for
byte**. The baseline Go comparison produces **17 unchanged and 5 changed**
images. Only the three SIMD and two private/shared-memory fixtures change.
Both SHA images are unchanged at 101308/101476 bytes. The diagnostic validates
66 actual native publications and checksums; its recorded clocks are not
performance samples.

Thus neither apparent SHA gains nor adverse Go SHA observations establish a
generated-code effect. Equal guest code also does not isolate host glue,
publication placement, process layout or scheduler effects. No causal attribution
or new LLGo issue follows from this noisy event. LLGo #2983 remains with its
existing contributor.

## Integration and remaining work

The earlier source `00bb578` has a successful full CI run
[41992896773](https://github.com/cpunion/wasmvm/actions/runs/41992896773): 21 required
jobs pass, with one intentional optional skip; the longest job takes **504 s**.
Its report/coverage artifact audit is recorded with that source's evidence.
That CI does not validate the newer `68b2269` source. Its independent hosted
[comparison 41992901763](https://github.com/cpunion/wasmvm/actions/runs/41992901763)
is still pending at this record and has not been resubmitted.

Exact final-head four-profile CI and review remain required before integration.
The branch stays unpublished as a PR under the one-active-contribution policy
while LLGo #2991 is open. Remaining work includes general value allocation,
splitting/rematerialization, GC/helper/shared-memory/boundary costs, preparation
costs, robust competitive qualification, broader WASI-world acceptance and release
qualification. No overall milestone is marked complete by this increment.

## Retained evidence

The [data index](data/loop-invariant-literals-2026-11-05.json) retains all per-case
ratios/phases, qualification commands, compiler identities, manifest checks and
publication comparisons. The raw bundle includes all five result groups,
orders, calibration intervals, unsupported probes, source/probe/portability
patches, build hashes and complete static spans. A baseline checkout guard
rejected the first setup before compilation or measurement; its reason is
retained separately from the successful build. No timed collection was retried.

Local bundle: `artifacts/loop-invariant-literals-20261105/local-evidence.tar.gz`, **3318185 bytes, 230 files**;
SHA-256 `64480d606c0659dfe46677c7ccf826c593d1a71f363a47a6e886635116b12a2d`. Earlier frozen archives are preserved.
