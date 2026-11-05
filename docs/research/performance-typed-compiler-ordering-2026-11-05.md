# Typed native preparation: 2026-11-05

**Candidate implementation: Go/LLGo semantics and image identity pass; elapsed
acceptance remains open. All 88 preparation A/A intervals are unresolved.**
Source `55976a42dbb4df4ef848446eb449b813f1c46fa7`; baseline
`42ca45572c2c9fd0d3996b4347dd07482c4eefa2` (engine `68b2269`).
The [preceding loop event](performance-loop-invariant-literals-2026-11-05.md)
retains its separate sustained warm measurements. These events are not pooled.

## Implementation

All 12 reflective `sort.Slice` sites in native codegen and module preparation
now use typed ordering or bounded selection. Allocation and loop residency share
one total priority comparator: descending weight, then ascending kind/key.
Interval, loop, integer/vector literal, global and runtime-offset ordering retain
their previous complete tie-breaks. Numeric comparisons avoid subtraction and
therefore retain unsigned boundary behavior.

Heap-cache planning retains only its two best eligible classes during the scan,
with the same frequency/key order. Candidate selection takes O(n) time and O(1)
candidate storage; the existing frequency map remains. Structurally canonical
type labels index the indirect-leaf map directly as `uint32`, eliminating the
decimal formatting and temporary string-key array. The labels are borrowed from
immutable validated metadata. Backend instructions, ABI, frame/root/poll/fuel
contracts and feature profiles retain their existing implementation.

The complete codegen/nativeexec/public-vm tests pass on macOS arm64 with Go
1.27.3 and LLGo `3571c25`, consistently using `-O3 -lto=full` for LLGo.
Targeted tests cover zero/one/many eligible caches, equal priorities, class zero,
maximum uint64, every caller-register budget, both architectures and both tiers.
Go vet passes. Three test packages cross-cover **codegen 97.66% (7674/7858)**
and **nativeexec 92.54% (459/496)**. The initial isolated nativeexec test scope
covered 84.9%; the public VM integration tests supply additional real paths.
These scopes differ from the earlier one-package coverage and full CI gates.

Actual native publication checks validate **all 22 baseline/head image byte
sequences and entry-offset lists**, including Go/LLGo candidate equality.
The baseline reuses 22 retained publications only after exact engine/harness
hashes, instrumentation and toolchain match; 44 candidate publications are fresh.
Diagnostic patches change only the publication hook and expose no timing claim.
Native guest bytes are unchanged. Host runner binaries grow 6495266→6560594
bytes under Go and 4329376→4345840 under LLGo; host code growth remains recorded.

## Go allocation-event diagnostic

`testing.AllocsPerRun(5)` performs one warmup and five fresh
Engine/Compile/Close operations per module. Input reading and JSON serialization
remain outside counting. Every module is actually compiled; no Store or guest
call participates. These are **Go heap event counts**, not LLGo counters,
guest allocations, executable memory, peak memory or elapsed performance.

All 22 cases record fewer Go heap events, by 18–136 per operation:

| Workload | Baseline events | Candidate events | Difference |
| --- | ---: | ---: | ---: |
| core/call-control | 479 | 457 | -22 |
| core/integer-loop | 648 | 619 | -29 |
| core/function-calls | 1141 | 1072 | -69 |
| core/memory-load-store | 798 | 762 | -36 |
| memory/narrow-unaligned | 906 | 871 | -35 |
| memory/dependent-address | 861 | 826 | -35 |
| memory/float64-unaligned | 839 | 804 | -35 |
| runtime/simd-memory-full | 781 | 758 | -23 |
| runtime/simd-memory-widen | 751 | 730 | -21 |
| runtime/simd-memory-lanes | 778 | 755 | -23 |
| gc/i31-roundtrip | 494 | 476 | -18 |
| gc/struct-access | 598 | 567 | -31 |
| gc/struct-allocation | 588 | 557 | -31 |
| gc/struct-indirect-access | 829 | 775 | -54 |
| threads/private-memory | 622 | 598 | -24 |
| threads/shared-memory | 627 | 601 | -26 |
| threads/atomic-rmw | 559 | 533 | -26 |
| threads/notify-empty | 551 | 526 | -25 |
| threads/wait-mismatch | 568 | 543 | -25 |
| threads/wait-timeout | 567 | 542 | -25 |
| applications/sha256-clang | 4531 | 4398 | -133 |
| applications/sha256-rust | 4567 | 4431 | -136 |

## Preparation comparison and adverse evidence

This single local macOS arm64 event keeps the existing 22 binary fixtures,
compiler/reference versions, five fresh blocks, two warmups, seven interleaved
calls and fixed order seed. Runtime counts are deliberately **17** (call-control
1), reducing execution work after preparation. CompileValidate still measures
the complete public compilation API on a fresh engine/process. These low-count
warm clocks are retained only as non-gating diagnostics; they do not establish
sustained throughput or supersede the earlier original-count results.

Own native preparation produces its ordinary policy tiers and applicable linked
expansion within one budget. Reference strategies remain separately pinned;
equal compiler internals or identical tier production are not claimed.
References are wazero 1.12.0, Wasmtime 48.0.1, Wasmer 7.3.0 and WasmEdge 0.17.1
(interpreter). Executables/libraries are reused only after digest/source checks;
own base/head Go/LLGo binaries are freshly built with identical flags.

Preparation and warmups run serially. Process readiness, loading, engine setup,
compile/validate, instantiation, lookup, first call, warm calls, cleanup and
process wall time remain separately recorded. There is no empty-call, startup
or clock subtraction. Native C remains a precompiled execution reference;
an absent Wasm preparation phase is not a zero-time compilation result.
The full raw phases, orders and clocks remain in the evidence bundle.

Independent audit checks **22 cases, 362 alternative runtime/argument
configurations and 12 explicit unsupported GC probes**. Preparation calibration
uses the same four duplicate-binary controls, now evaluating paired log
compileValidate ratios across the five fresh processes. Fixed-seed 8000-draw
bootstrap marginal 95% intervals must fit ±2.5%. **0/88 individual controls
resolve; 0/22 workloads resolve all four.** CPU affinity/frequency and neighbor
isolation are uncontrolled. No collection was retried for favorable outcomes.

The following are observed complete preparation-time ratios, lower is better.
Every row is inconclusive, including apparent improvements. The data index
also records absolute compileValidate medians for every runtime.

| Workload | Go native h/b | LLGo native h/b | Go interpreter h/b | LLGo interpreter h/b | LLGo / Go native | LLGo / wazero JIT | LLGo / Wasmtime | LLGo / Wasmer | Preparation A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 0.783× | 1.525× | 0.977× | 1.131× | 2.526× | 4.254× | 0.165× | 0.271× | inconclusive |
| core/integer-loop | 1.012× | 1.999× | 1.028× | 2.354× | 2.239× | 5.660× | 0.377× | 0.383× | inconclusive |
| core/function-calls | 1.151× | 1.104× | 0.898× | 3.186× | 1.692× | 4.968× | 0.261× | 0.313× | inconclusive |
| core/memory-load-store | 0.922× | 2.040× | 1.043× | 3.651× | 2.585× | 4.431× | 0.533× | 0.540× | inconclusive |
| memory/narrow-unaligned | 0.920× | 1.379× | 0.988× | 1.050× | 2.350× | 5.269× | 0.332× | 0.404× | inconclusive |
| memory/dependent-address | 0.858× | 1.974× | 1.235× | 2.892× | 2.510× | 4.754× | 0.519× | 0.434× | inconclusive |
| memory/float64-unaligned | 0.919× | 0.806× | 0.963× | 2.838× | 1.996× | 3.630× | 0.341× | 0.281× | inconclusive |
| runtime/simd-memory-full | 1.141× | 1.852× | 0.949× | 3.063× | 2.027× | 6.782× | 0.394× | 0.664× | inconclusive |
| runtime/simd-memory-widen | 0.911× | 1.495× | 0.739× | 3.204× | 1.774× | 2.764× | 0.294× | 0.346× | inconclusive |
| runtime/simd-memory-lanes | 0.844× | 1.009× | 1.279× | 3.080× | 1.603× | 3.659× | 0.333× | 0.235× | inconclusive |
| gc/i31-roundtrip | 0.839× | 1.672× | 0.676× | 0.963× | 2.825× | unsupported | 0.171× | unsupported | inconclusive |
| gc/struct-access | 0.794× | 0.568× | 0.827× | 2.167× | 1.783× | unsupported | 0.334× | unsupported | inconclusive |
| gc/struct-allocation | 1.023× | 1.796× | 1.317× | 1.756× | 2.620× | unsupported | 0.322× | unsupported | inconclusive |
| gc/struct-indirect-access | 1.168× | 2.048× | 0.931× | 2.741× | 2.059× | unsupported | 0.143× | unsupported | inconclusive |
| threads/private-memory | 0.947× | 2.642× | 0.941× | 1.239× | 4.092× | 6.510× | 0.633× | 0.582× | inconclusive |
| threads/shared-memory | 1.258× | 1.843× | 0.987× | 1.110× | 1.622× | 2.894× | 0.450× | 0.489× | inconclusive |
| threads/atomic-rmw | 1.378× | 1.375× | 0.928× | 0.738× | 1.535× | 2.436× | 0.267× | 0.330× | inconclusive |
| threads/notify-empty | 1.327× | 2.437× | 0.767× | 1.025× | 1.852× | 3.986× | 0.474× | 0.450× | inconclusive |
| threads/wait-mismatch | 1.415× | 1.863× | 0.908× | 0.927× | 2.198× | 4.647× | 0.483× | 0.483× | inconclusive |
| threads/wait-timeout | 1.035× | 1.677× | 0.986× | 1.089× | 2.265× | 3.994× | 0.340× | 0.272× | inconclusive |
| applications/sha256-clang | 1.112× | 1.299× | 0.965× | 1.280× | 1.836× | 16.342× | 2.510× | 2.320× | inconclusive |
| applications/sha256-rust | 1.042× | 1.045× | 1.613× | 1.617× | 2.366× | 11.801× | 1.519× | 1.821× | inconclusive |

Adverse observations include LLGo native integer/memory/indirect-GC preparation
at approximately 2× baseline and private-memory preparation at 2.64×. Several
LLGo interpreter observations also exceed 3× despite not executing native
codegen. Identical-binary preparation samples themselves vary substantially;
the controls do not distinguish a source effect from other variation. These
observations are retained, and no compiler defect or elapsed gain is inferred.

Clang/Rust SHA candidate native preparation is observed at 3791.58/3391.42 µs
under Go and 6960.25/8023.46 µs under LLGo. LLGo is observed at 1.84×/2.37× Go
and 16.34×/11.80× wazero JIT. Allocation reductions do not waive these deficits
or close elapsed acceptance. The five collection groups take
79/94/57/84/36 seconds; these are collection wall times, not guest rates or CI
job durations.

## Outstanding gates and evidence

This candidate stays on the contribution branch. Full exact final-head
Go/LLGo CI on both targets and review remain required. The already dispatched
[42011604985](https://github.com/cpunion/wasmvm/actions/runs/42011604985) validates
the preceding `42ca455` source; it does not validate this candidate. Existing
[41992901763](https://github.com/cpunion/wasmvm/actions/runs/41992901763) measures
the still earlier `00bb578` source. Neither is restarted or relabeled.
The one-active-contribution policy still holds while LLGo #2991 is open.
LLGo #2983 stays with its existing contributor.

General SSA value allocation, splitting/rematerialization, helper/GC/shared-
memory and boundary costs, robust comparative performance, broader WASI-world
qualification and release acceptance remain open. This implementation does not
complete those goals or claim superiority to a competitor or ordinary native.

The [data index](data/typed-compiler-ordering-2026-11-05.json) retains commands,
coverage scopes, per-case preparation medians/calibration, allocation counts,
image identities and raw input hashes. The bundle retains all five groups,
all phases and adverse outcomes, source/probe/portability patches and build pins.
An initial allocation-diagnostic source guard rejected two excluded harness
files before any build or counter sample; the reason is retained. It required
the existing harness exclusion and did not change engine code or timing data.

Local bundle: `artifacts/typed-compiler-ordering-20261105/evidence.tar.gz`, 3447715 bytes, 206 files;
SHA-256 `1b127ac948d35b9c83a1a167b294822ba6bcc5ce91abe90bcb2c15300b907436`. Earlier bundles are immutable.
