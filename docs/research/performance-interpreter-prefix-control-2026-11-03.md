# Keep internal control outside Wasm prefix dispatch: 2026-11-03

Status: held after [full semantic CI and repeated Linux regressions](pr112-control-ci-2026-11-03.md). The [operand-lifetime correction](performance-interpreter-cold-operands-2026-11-03.md) is separate; all original outcomes remain retained.

The [initial PR112 head](pr112-initial-ci-2026-11-03.md) passes all semantic CI
but regresses Linux LLGo scalar interpretation 7.5–17.1%. Preserve its entire
benchmark and all adverse outcomes. This correction adds an upper bound to
extension classification: all validated Wasm opcodes are below slots.ControlBase,
while internal control instructions already have their own dispatch cases.
They now bypass prefix classification. No source instruction, fuel, polling,
GC root/PC, memory or trap contract changes. The scalar SIMD masks remain.

## Source and method

Base `b34e74bcce8e0ae607ca31a89977fc1d217ea9e7` is the implementation integrated
by PR111; corrected source `d806b0c8fed545b872ff02dc9ed7f053b93edb7e`. Fixed Go
1.27.3, LLGo `983465e` with `-O3 -lto=full`, Apple M4 Max/macOS arm64. Exactly
two engine files differ from baseline. No experimental SIMD compiler profile
or held vector-constant prototype is adopted. Reference pins remain unchanged.

The complete twenty-kernel comparison retains the original counts, three fresh
serially interleaved process blocks, two warmups/seven samples, independent
checksums, alternate inputs, and exact baseline Go/LLGo interpreter A/A aliases.
All 864 VM/reference observations, 27 independently checked ordinary C records,
288 alternate checks, twelve unsupported GC outcomes and 297 regenerated
summaries are retained. Preparation, first/warm calls, cleanup and readiness
are separate; there is no subtraction or pooling with the earlier studies.

## Complete interpreter outcomes

Warm ns/source iteration except empty ns/call; lower is better.

| Kernel | Go base → head | LLGo base → head | WasmEdge | wazero |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 50.883 → 50.648 | 43.908 → 40.966 | 39.578 | 64.086 |
| core/call-control | 2542.000 → 2917.000 | 2417.000 → 2833.000 | 4000.000 | 875.000 |
| core/integer-loop | 41.986 → 42.236 | 36.197 → 32.801 | 49.627 | 83.164 |
| core/function-calls | 64.888 → 66.239 | 60.519 → 56.453 | 62.799 | 81.229 |
| core/memory-load-store | 84.841 → 85.191 | 75.811 → 70.211 | 91.970 | 144.734 |
| memory/narrow-unaligned | 110.744 → 110.149 | 99.981 → 93.440 | 114.632 | 185.563 |
| memory/dependent-address | 56.268 → 55.981 | 49.053 → 46.285 | 54.705 | 99.783 |
| memory/float64-unaligned | 88.844 → 87.266 | 79.479 → 74.268 | 86.767 | 138.081 |
| runtime/simd-memory-full | 109.685 → 97.855 | 98.177 → 75.909 | 79.015 | 153.631 |
| runtime/simd-memory-widen | 87.600 → 78.103 | 85.378 → 68.563 | 62.094 | 146.291 |
| runtime/simd-memory-lanes | 129.959 → 118.288 | 124.478 → 93.809 | 90.901 | 187.220 |
| gc/i31-roundtrip | 53.183 → 52.629 | 45.150 → 39.133 | 45.200 | — |
| gc/struct-access | 55.292 → 56.133 | 48.821 → 46.508 | 55.600 | — |
| gc/struct-allocation | 119.371 → 118.033 | 147.738 → 141.304 | 106.800 | — |
| gc/struct-indirect-access | 98.967 → 101.138 | 91.188 → 87.217 | 96.000 | — |
| threads/shared-memory | 82.895 → 83.719 | 42.939 → 40.080 | 38.204 | 61.532 |
| threads/atomic-rmw | 63.831 → 61.713 | 43.668 → 40.410 | 31.903 | 43.452 |
| threads/notify-empty | 44.483 → 40.721 | 39.546 → 35.608 | 40.800 | 43.188 |
| threads/wait-mismatch | 72.408 → 70.029 | 52.688 → 49.150 | 43.800 | 68.862 |
| threads/wait-timeout | 69.454 → 69.146 | 48.792 → 44.212 | 96.700 | 293.725 |

LLGo improves all nineteen sustained cases, 4.4–24.6%; single-call empty
increases 17.2%. SIMD full leads this WasmEdge; widening/lanes remain about
1.104x/1.032x its time. Go SIMD improves 9.0–10.8%, but guest calls increase
2.1%, direct/indirect GC 1.5%/2.2% and shared memory 1.0%. These are shared-host
diagnostics, not controlled significance or full-portfolio superiority.

## Complete native and C controls

Same campaign and units; the ordinary C algorithms omit VM contracts. No
ordinary-C GC equivalent is invented. All adverse controls remain visible.

| Kernel | Go base → head | LLGo base → head | wazero | Wasmtime | Wasmer | Ordinary C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.348 → 0.347 | 0.335 → 0.322 | 1.183 | 0.887 | 1.028 | — |
| core/call-control | 4083.000 → 3625.000 | 3875.000 → 4250.000 | 458.000 | 1000.000 | 5000.000 | — |
| core/integer-loop | 1.620 → 1.637 | 1.606 → 1.594 | 1.749 | 1.594 | 1.571 | 1.564 |
| core/function-calls | 0.813 → 0.818 | 0.792 → 0.802 | 1.470 | 0.934 | 0.789 | 0.785 |
| core/memory-load-store | 0.453 → 0.457 | 0.437 → 0.438 | 0.878 | 0.479 | 0.542 | 0.272 |
| memory/narrow-unaligned | 0.721 → 0.716 | 0.694 → 0.699 | 1.070 | 0.691 | 0.724 | 0.461 |
| memory/dependent-address | 1.875 → 1.892 | 1.846 → 1.852 | 3.182 | 2.109 | 2.147 | 1.433 |
| memory/float64-unaligned | 0.754 → 0.761 | 0.723 → 0.745 | 0.873 | 0.729 | 0.739 | 0.682 |
| runtime/simd-memory-full | 3.455 → 3.436 | 3.422 → 3.483 | 4.064 | 3.660 | 3.529 | 3.521 |
| runtime/simd-memory-widen | 1.072 → 1.098 | 1.073 → 1.065 | 1.181 | 1.064 | 1.067 | 1.082 |
| runtime/simd-memory-lanes | 5.073 → 5.103 | 5.030 → 5.043 | 6.124 | 5.961 | 5.861 | 4.485 |
| gc/i31-roundtrip | 2.204 → 2.104 | 2.025 → 2.013 | — | 2.200 | — | — |
| gc/struct-access | 2.396 → 2.467 | 2.258 → 2.225 | — | 1.500 | — | — |
| gc/struct-allocation | 44.842 → 46.100 | 55.237 → 53.362 | — | 108.000 | — | — |
| gc/struct-indirect-access | 5.679 → 5.825 | 5.421 → 5.408 | — | 3.000 | — | — |
| threads/shared-memory | 0.803 → 0.810 | 0.797 → 0.799 | 1.138 | 1.030 | 0.655 | — |
| threads/atomic-rmw | 1.840 → 1.837 | 1.831 → 1.831 | 1.977 | 1.807 | 3.876 | — |
| threads/notify-empty | 1.587 → 1.587 | 1.567 → 1.663 | 21.871 | 7.200 | 5.500 | — |
| threads/wait-mismatch | 1.808 → 1.754 | 1.650 → 1.700 | 24.921 | 9.000 | 3.100 | — |
| threads/wait-timeout | 2.675 → 2.546 | 1.812 → 1.904 | 246.071 | 43.200 | 170.300 | — |

Retain LLGo native empty +9.7%, float64 memory +2.9%, full-vector SIMD +1.8%,
notify +6.1% and wait-timeout +5.1%, plus Go widening +2.4%, direct/indirect GC
+3.0%/+2.6% and allocation +2.8%. Some native intervals remain short. The
earlier candidate's separate batch/long controls have different source identity
and do not waive these new outcomes. Public boundary, ordinary-C memory,
interpreter SIMD widening/lanes and controlled Linux/ARM64 acceptance stay open.

## Validation

Full affected SIMD/execute/VM suites pass under Go and pinned full-LTO LLGo,
with 48 passing conformance reports each. Both include every mandatory proposal
on all three tiers and forced collection for Stack Switching. The C/harness
implementation is unchanged from its 81 passing tests and initial-head shared
quality CI. New final-head 31-package compiler/platform CI and thirty coverage
gates remain mandatory before merge. The initial Linux regression is not fixed
merely by a favorable ARM64 observation.

[Complete source/profile identities, original outcomes, semantic reports and
independent verification](evidence/performance-interpreter-prefix-control-2026-11-03.json.gz).
