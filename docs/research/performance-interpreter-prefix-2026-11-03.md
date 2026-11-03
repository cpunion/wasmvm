# Unified interpreter prefix dispatch: 2026-11-03

Status: candidate, pending exact-head dual-platform CI. Performance acceptance remains open.

Route SIMD, GC and Wide Arithmetic by their encoded prefix before unrelated
scalar/extension checks. Stack Switching retains its existing instruction range.
SIMD memory operations still use checked memory helpers and the common trap/slot
writeback; GC retains PC publication, roots and errors. Fuel, polling frequency,
operands and continuation contracts do not change. Replace the pure-SIMD guard's
per-call mask array with scalar masks. Linked Go/LLGo assembly is retained.
The separate held direct-vector-constant prototype is not included.

## Identities and separate campaigns

Base `b34e74bcce8e0ae607ca31a89977fc1d217ea9e7` is PR111's integrated implementation;
candidate `469b3e5c8f0be41212ccaba1b9f1e40780abc370`. Go 1.27.3, LLGo `983465e`
with `-O3 -lto=full`, Apple M4 Max/macOS arm64. Only execute/execution.go and
simd/simd.go differ in the measured engine source inventories. Common frozen
instrumentation enables the declared Core 3/Threads feature profile. Reference
pins remain wazero 1.12.0, WasmEdge 0.17.1, Wasmtime 48.0.1 and Wasmer 7.3.0.

The first SIMD-first candidate `e4562fc` passes semantics but retains LLGo i31
interpretation +4.1% and native/Go regressions. Its complete twenty-case campaign
is retained separately: 864 VM and 27 C observations, 288 alternate checks, twelve
explicit unsupported GC outcomes and 297 summaries. Unified dispatch replaces
that candidate; no original outcome is dropped or pooled.

The unified twenty-case campaign uses the same original counts and three fresh
serially interleaved process blocks, two warmups/seven samples. It independently
verifies the same 864/27/288/12/297 complete inventory. All preparation, readiness,
first calls, warm calls and cleanup remain separate. Ordinary C uses nine checked
scalar/SIMD algorithms and has fewer execution contracts than a VM.

## Complete interpreter outcomes

Warm ns/source iteration, except empty in ns/call; lower is better. GC reference
omissions are explicit unsupported outcomes, not zero times.

| Kernel | Go base → head | LLGo base → head | WasmEdge | wazero |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 49.786 → 50.154 | 43.654 → 40.888 | 40.562 | 61.708 |
| core/call-control | 2542.000 → 2667.000 | 2625.000 → 3250.000 | 3000.000 | 1042.000 |
| core/integer-loop | 41.294 → 41.729 | 35.537 → 32.480 | 48.992 | 81.609 |
| core/function-calls | 64.744 → 64.559 | 59.622 → 57.682 | 64.303 | 81.722 |
| core/memory-load-store | 84.805 → 85.416 | 74.626 → 70.363 | 93.208 | 145.504 |
| memory/narrow-unaligned | 111.768 → 110.382 | 98.203 → 94.031 | 113.341 | 185.226 |
| memory/dependent-address | 56.852 → 55.678 | 48.792 → 45.987 | 53.737 | 95.972 |
| memory/float64-unaligned | 87.111 → 89.206 | 78.011 → 73.829 | 84.395 | 135.695 |
| runtime/simd-memory-full | 106.624 → 97.154 | 95.795 → 73.118 | 77.493 | 149.384 |
| runtime/simd-memory-widen | 89.006 → 79.193 | 84.461 → 67.762 | 61.439 | 143.245 |
| runtime/simd-memory-lanes | 139.158 → 128.444 | 134.913 → 104.969 | 98.058 | 202.318 |
| gc/i31-roundtrip | 54.421 → 50.763 | 45.783 → 43.200 | 46.400 | — |
| gc/struct-access | 76.508 → 85.392 | 66.354 → 60.388 | 75.400 | — |
| gc/struct-allocation | 121.829 → 120.733 | 147.875 → 144.938 | 112.800 | — |
| gc/struct-indirect-access | 105.537 → 110.808 | 97.025 → 94.917 | 103.400 | — |
| threads/shared-memory | 94.326 → 94.944 | 48.887 → 45.649 | 43.486 | 69.494 |
| threads/atomic-rmw | 74.429 → 75.418 | 51.853 → 48.045 | 37.675 | 54.739 |
| threads/notify-empty | 55.083 → 58.579 | 49.646 → 46.104 | 55.200 | 52.142 |
| threads/wait-mismatch | 95.546 → 92.254 | 64.325 → 56.029 | 58.300 | 85.650 |
| threads/wait-timeout | 76.388 → 78.792 | 54.850 → 49.987 | 109.400 | 325.504 |

LLGo improves all nineteen sustained interpreter kernels, 2.0–23.7%; single-call
empty increases 23.8%. Full-vector SIMD takes 73.118 versus WasmEdge 77.493;
widening/lane remain about 1.103x/1.070x WasmEdge. Go SIMD improves 7.7–11.0%,
but direct/indirect GC increase 11.6%/5.0% and notify 6.3%. Native short controls
retain LLGo lanes +4.5%, i31 +6.9%, direct GC +7.8% and notify +12.3%, plus Go
i31 +11.0%. The complete native/reference/C table remains in lossless evidence.

## Separately declared long-loop controls

Exact same binaries; no repeated short campaign. Five fresh interleaved process
blocks, two warmups/seven samples and exact native/interpreter baseline A/A.
Four native SIMD/scalar cases keep the [earlier long-C counts](performance-native-simd-long-2026-11-03.md).
Native i31/direct/indirect GC use 100m/100m/30m iterations; interpreter GC uses
1m each. Counts match every runtime within each backend/case. Independent
geometric-series LCG/Python and previously sanitized C oracles check results.
The complete scope retains 390 VM and twenty C observations, 42 additional
alternate checks and 82 independently regenerated summaries. VM warm intervals
are 36–181 ms. Native GC references Wasmtime; interpreter GC references WasmEdge.
No ordinary-C GC equivalence is claimed. Original adverse short controls remain.

| Kernel/backend | Go base → head | LLGo base → head | Reference / ordinary C |
| --- | ---: | ---: | --- |
| runtime/simd-memory-widen / native | 1.02343 → 1.02336 | 1.02451 → 1.02770 | wasmtime-cranelift: 1.02559; C: 1.02523 |
| runtime/simd-memory-lanes / native | 4.97015 → 4.97856 | 4.96527 → 4.96841 | wasmtime-cranelift: 5.79140; C: 4.49760 |
| runtime/simd-memory-full / native | 3.31767 → 3.31422 | 3.30827 → 3.31149 | wasmtime-cranelift: 3.90687; C: 3.58327 |
| core/memory-load-store / native | 0.41790 → 0.41732 | 0.41723 → 0.41660 | wasmtime-cranelift: 0.47087; C: 0.26780 |
| gc/i31-roundtrip / native | 1.32963 → 1.32885 | 1.32868 → 1.32724 | wasmtime-cranelift: 1.79373 |
| gc/i31-roundtrip / interpreter | 50.23087 → 49.74079 | 42.26233 → 37.70471 | wasmedge-interpreter: 44.04600 |
| gc/struct-access / native | 1.28402 → 1.28105 | 1.28485 → 1.28344 | wasmtime-cranelift: 1.02439 |
| gc/struct-access / interpreter | 53.94087 → 55.14350 | 47.97504 → 44.31850 | wasmedge-interpreter: 54.24200 |
| gc/struct-indirect-access / native | 4.51885 → 4.49249 | 4.52872 → 4.52635 | wasmtime-cranelift: 2.54777 |
| gc/struct-indirect-access / interpreter | 96.92675 → 98.47125 | 89.07625 → 83.02204 | wasmedge-interpreter: 94.78400 |

All native base/head medians differ by about −0.6% to +0.3%; this longer scope
does not repeat the larger short-control regressions, but does not erase them.
LLGo GC interpretation improves 6.8–10.8%. Retain Go direct/indirect interpretation
+2.2%/+1.6%; no controlled no-regression or whole-portfolio parity claim follows.

## Public call controls

A separate common-harness study batches 10,000 public calls, five process blocks,
two warmups/seven samples and exact baseline A/A under both backends/compilers.
It retains 180 observations, 36 alternate checks and 36 summaries. Guest cases
are empty and 64 integer iterations. No session/internal component substitutes
for the public API. Wazero default/enabled termination, Wasmtime and Wasmer
remain separate references with their own contracts. Warm ns/complete call:

| Configuration | Empty base → head | Integer 64 base → head |
| --- | ---: | ---: |
| go interpreter public | 269.408 → 275.850 | 2993.121 → 3037.092 |
| go native public | 352.400 → 341.904 | 385.571 → 381.929 |
| llgo interpreter public | 755.058 → 745.121 | 2896.133 → 2853.800 |
| llgo native public | 963.192 → 979.938 | 962.946 → 939.529 |

LLGo interpreter empty decreases 1.3%; its earlier single-call +23.8% is not
reproduced in this scope. LLGo native empty increases 1.7%, with baseline A/A
also +1.7%. Go interpreter empty increases 2.4%, with A/A +2.5%. These controls
illustrate unresolved calibration; they do not prove every adverse outcome is
harmless. Public-boundary, startup and allocation deficits remain open.

## Validation and retained limits

Each candidate executes the full affected SIMD/execute/VM suites under Go and
fixed full-LTO LLGo, generating 48 passing reports per compiler. Both cover
all five mandatory extensions on three tiers and forced Stack Switching GC.
The 81 benchmark harness tests pass, including 234 independent C oracle cases.
The first candidate initially lacks prepared Core inputs; its original failure
and partial reports are retained, then corrected before successful validation.
The long-C study separately retains its three measurement setup failures.

Final contribution-head full 31-package Go/LLGo × Linux amd64/macOS arm64 CI,
thirty coverage gates and review remain required before integration. Controlled
portfolio/native-C acceptance, startup/allocation/cancellation and expanded WASI
world conformance stay open. No new LLGo compiler defect is inferred from these
throughput outcomes.

[All source identities, both complete campaigns, original failures, semantic
reports, assembly, batch/long controls and independent verification](evidence/performance-interpreter-prefix-2026-11-03.json.gz).
