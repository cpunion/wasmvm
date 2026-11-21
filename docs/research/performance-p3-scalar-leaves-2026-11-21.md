# P3 scalar scheduler leaves and synchronous host transactions, 2026-11-21

Core scalar leaves update serialized host metadata inside the active frame. Snapshot-based synchronous component calls and ready-source reads reuse private transaction and result storage. The changes preserve public result ownership, resource accounting, cancellation and the suspendable call path.

- P3 16 B, GO. CPU 73.360 → 72.707 µs/request and throughput 53,186.65 → 54,254.48 requests/s. Paired CPU comparison against source base 0.976 [0.961, 0.992]; Δ -2.38%; speed +2.43% [+0.82%, +4.07%]; unresolved. Against Wasmtime 1.741 [1.598, 1.900]; Δ +74.11%; speed -42.56% [-47.37%, -37.43%]; unresolved.
- P3 16 B, LLGO. CPU 208.23 → 206.18 µs/request and throughput 22,453.01 → 22,603.74 requests/s. Paired CPU comparison against source base 0.996 [0.990, 1.002]; Δ -0.39%; speed +0.39% [-0.19%, +0.97%]; pass. Against Wasmtime 4.906 [4.503, 5.359]; Δ +390.59%; speed -79.62% [-81.34%, -77.79%]; unresolved.
- P3 65,536 B, GO. CPU 111.62 → 108.74 µs/request and throughput 34,267.51 → 35,268.67 requests/s. Paired CPU comparison against source base 0.980 [0.966, 1.001]; Δ -1.99%; speed +2.03% [-0.12%, +3.50%]; pass. Against Wasmtime 1.841 [1.699, 1.974]; Δ +84.09%; speed -45.68% [-49.33%, -41.14%]; unresolved.
- P3 65,536 B, LLGO. CPU 249.17 → 244.47 µs/request and throughput 17,178.95 → 17,167.43 requests/s. Paired CPU comparison against source base 0.971 [0.962, 0.980]; Δ -2.87%; speed +2.95% [+2.07%, +3.91%]; unresolved. Against Wasmtime 4.252 [3.844, 4.737]; Δ +325.24%; speed -76.48% [-78.89%, -73.98%]; unresolved.

All 33 core workloads and two complete P3 payloads are measured on the same pinned toolchains. Tables give preparation and steady execution independently. The performance acceptance gate remains open.

## Percentage differences

Cells show `absolute observation (signed speed Δ%)`. Positive speed change means faster. Native execution and P3 use **Wasmtime** as a fixed reference; interpreter execution uses **wazero**, with **WasmEdge** explicitly named for GC rows where wazero rejects the feature. Using a fixed reference keeps the percentages comparable across revisions. The fastest observed runtime varies by workload.

For time, speed Δ = `(reference / candidate - 1) * 100%`; for throughput, speed Δ = `(candidate / reference - 1) * 100%`. Thus 100 ns to 50 ns is **−50% time**, **+100% speed**. P3 CPU percentages mean work per CPU time, separate from wall-clock throughput. Instruction/allocation/byte counters show quantity Δ rather than speed. Source comparisons retain paired ratio intervals and report quantity Δ plus speed Δ where applicable; reciprocal interval endpoints are reversed.

Competitive cell percentages are descriptive quotients of the displayed medians. Paired block estimates below can differ. `†` on these cells means either the candidate or the named reference has unresolved A/A controls; qualified conclusions require resolved controls. Phase boundaries with different contracts have separate absolute observations. The percentages apply to the frozen event and measured revision.

## Revisions, profiles and protocol

- Engine source: base `57e67995` (merged PR #204), measured head `23854d99`. Reused binaries retain their producer revisions. The build records verify production-source equivalence for each reused binary.
- Go **1.27.3**; LLGo **3c6b33ec**, **`-O3 -lto=full`**, LLVM **22.1.8**, default production GC. Source base/head share the compiler profile.
- wazero **1.12.0**, native compiler/interpreter; Wasmtime **48.0.1**, Cranelift speed; Wasmer **7.3.0**, Cranelift; WasmEdge **0.17.1**, interpreter and **O3 AOT**. Comparison scope is the named profile of each runtime.
- Local Apple M4 Max, macOS arm64, 128 GiB RAM. External desktop load is uncontrolled and recorded; no owned builds, tests, profiles, analysis or edits overlap timing. The timing scope is macOS arm64.
- Core uses six fixed fresh-process blocks, each with symmetric same-binary A/A, two warmups and seven interleaved measured calls. **5,364 successful process records** after excluding exactly three failing sustained WasmEdge AOT cells. All 33 workloads remain represented. Allocation workloads use means including pauses; others use medians.
- P3 uses **72 processes per payload**, six roles with A/A, six fixed blocks, four workers, sixteen clients and twenty interleaved 512-request windows plus a separate flush. Verify status, headers, full body, trailers, CRC and joined guest completion. CPU/instruction counters include the driver and checks. Mach ticks use the recorded timebase.
- Ratios use paired block geometric means and 20,000 fixed-seed bootstrap resamples with marginal 95% intervals. Both relevant A/A intervals must fit `[0.975, 1.025]`. Each interval applies to its own comparison. The protocol fixes the sample size in advance and measures each phase directly.

**Legend:** `†` means the role or its percentage reference has an unresolved same-binary timing control. Values are observations; a qualified comparison additionally needs both controls and its ratio interval. `US` means this pinned profile rejected the feature in semantic preflight. `NM` means this benchmark adapter/profile was not measured; this status applies only to this measurement. `N/A` means the metric does not apply. `FAIL` is an execution failure in the named profile.

## Native compiled execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

The reference runtime is **Wasmtime**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5,750.00 (-57.61% speed) † | 5,188.00 (-53.02% speed) † | 1,271.00 (+91.78% speed) † | 2,437.50 (ref) † | 4,771.00 (-48.91% speed) † | 6,000.00 (-59.38% speed) † |
| original22-v1/core/integer-loop | 1.876 (+1.24% speed) † | 1.888 (+0.61% speed) † | 2.076 (-8.52% speed) † | 1.899 (ref) † | 1.885 (+0.76% speed) † | 1.844 (+3.00% speed) † |
| original22-v1/core/function-calls | 0.957 (+16.72% speed) † | 0.952 (+17.36% speed) † | 1.692 (-33.95% speed) † | 1.117 (ref) † | 0.937 (+19.32% speed) † | 0.958 (+16.59% speed) † |
| original22-v1/core/memory-load-store | 0.530 (+8.19% speed) † | 0.521 (+10.20% speed) † | 1.073 (-46.53% speed) † | 0.574 (ref) † | 0.658 (-12.82% speed) † | 0.341 (+68.19% speed) † |
| original22-v1/memory/narrow-unaligned | 0.777 (+7.05% speed) † | 0.778 (+6.97% speed) † | 1.292 (-35.58% speed) † | 0.832 (ref) † | 0.872 (-4.60% speed) † | 0.612 (+35.99% speed) † |
| original22-v1/memory/dependent-address | 2.145 (+13.01% speed) † | 2.217 (+9.38% speed) † | 3.792 (-36.06% speed) † | 2.425 (ref) † | 2.489 (-2.57% speed) † | 1.670 (+45.20% speed) † |
| original22-v1/memory/float64-unaligned | 0.887 (-4.68% speed) † | 0.878 (-3.70% speed) † | 1.045 (-19.07% speed) † | 0.845 (ref) | 0.840 (+0.61% speed) † | 0.799 (+5.81% speed) † |
| original22-v1/runtime/simd-memory-full | 3.982 (+7.31% speed) † | 3.977 (+7.45% speed) † | 4.449 (-3.95% speed) † | 4.273 (ref) † | 4.109 (+3.98% speed) † | 4.248 (+0.60% speed) † |
| original22-v1/runtime/simd-memory-widen | 1.219 (-1.91% speed) † | 1.202 (-0.50% speed) † | 1.312 (-8.86% speed) † | 1.196 (ref) † | 1.204 (-0.65% speed) † | 1.201 (-0.39% speed) † |
| original22-v1/runtime/simd-memory-lanes | 5.836 (+18.17% speed) † | 5.967 (+15.56% speed) † | 7.299 (-5.52% speed) † | 6.896 (ref) † | 6.908 (-0.17% speed) † | 5.917 (+16.55% speed) † |
| original22-v1/gc/i31-roundtrip | 2.783 (-9.28% speed) † | 2.227 (+13.38% speed) † | US | 2.525 (ref) † | US | 2.410 (+4.75% speed) † |
| original22-v1/gc/struct-access | 2.852 (-37.33% speed) † | 2.656 (-32.70% speed) † | US | 1.788 (ref) † | US | 12.123 (-85.26% speed) † |
| original22-v1/gc/struct-allocation | 33.418 (+134.24% speed) † | 35.164 (+122.61% speed) † | US | 78.278 (ref) † | US | 47.551 (+64.62% speed) † |
| original22-v1/gc/struct-indirect-access | 3.583 (-3.49% speed) † | 2.883 (+19.94% speed) † | US | 3.458 (ref) † | US | 30.706 (-88.74% speed) † |
| original22-v1/threads/private-memory | 1.848 (-3.83% speed) † | 1.096 (+62.17% speed) † | 2.263 (-21.45% speed) † | 1.777 (ref) † | 2.196 (-19.07% speed) † | 2.065 (-13.93% speed) † |
| original22-v1/threads/shared-memory | 3.154 (-47.69% speed) † | 2.292 (-28.00% speed) † | 1.813 (-8.97% speed) † | 1.650 (ref) † | 2.083 (-20.80% speed) † | 1.977 (-16.55% speed) † |
| original22-v1/threads/atomic-rmw | 3.652 (-31.55% speed) † | 3.015 (-17.07% speed) † | 2.823 (-11.44% speed) † | 2.500 (ref) † | 5.188 (-51.81% speed) † | 2.898 (-13.73% speed) † |
| original22-v1/threads/notify-empty | 3.390 (+179.04% speed) † | 2.175 (+334.87% speed) † | 28.248 (-66.52% speed) † | 9.458 (ref) † | 8.048 (+17.53% speed) † | 11.454 (-17.42% speed) † |
| original22-v1/threads/wait-mismatch | 3.379 (+234.09% speed) † | 2.240 (+404.10% speed) † | 29.604 (-61.86% speed) † | 11.290 (ref) † | 3.656 (+208.78% speed) † | 4.696 (+140.42% speed) † |
| original22-v1/threads/wait-timeout | 3.365 (+1464.88% speed) † | 3.329 (+1481.55% speed) † | 426.58 (-87.66% speed) † | 52.652 (ref) † | 208.10 (-74.70% speed) † | 96.435 (-45.40% speed) † |
| original22-v1/applications/sha256-clang | 240.87 (-8.31% speed) † | 250.93 (-11.99% speed) † | 350.00 (-36.90% speed) † | 220.85 (ref) † | 239.34 (-7.73% speed) † | 209.98 (+5.18% speed) † |
| original22-v1/applications/sha256-rust | 249.42 (-12.47% speed) † | 244.33 (-10.65% speed) † | 380.07 (-42.56% speed) † | 218.32 (ref) † | 222.76 (-2.00% speed) † | 218.43 (-0.05% speed) † |
| gc-sustained-v1/gc/i31-roundtrip | 1.595 (+34.60% speed) † | 1.591 (+34.93% speed) † | US | 2.147 (ref) † | US | 1.525 (+40.78% speed) † |
| gc-sustained-v1/gc/struct-access | 1.284 (-4.62% speed) † | 1.231 (-0.49% speed) † | US | 1.224 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 38.531 (+55.93% speed) † | 39.356 (+52.66% speed) † | US | 60.081 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1.965 (+52.93% speed) † | 1.910 (+57.32% speed) † | US | 3.005 (ref) † | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.545 (+132.83% speed) † | 0.408 (+211.50% speed) † | 1.387 (-8.46% speed) † | 1.269 (ref) † | 1.256 (+1.03% speed) † | 1.300 (-2.34% speed) † |
| threads-sustained-v2/threads/shared-memory | 1.705 (-19.97% speed) † | 1.886 (-27.65% speed) † | 1.513 (-9.83% speed) † | 1.364 (ref) † | 1.335 (+2.20% speed) † | 1.348 (+1.19% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 2.379 (-7.50% speed) † | 2.230 (-1.33% speed) † | 2.473 (-11.01% speed) † | 2.201 (ref) † | 4.745 (-53.62% speed) † | 2.217 (-0.71% speed) † |
| threads-sustained-v2/threads/notify-empty | 1.303 (+542.92% speed) † | 1.149 (+628.80% speed) † | 25.701 (-67.41% speed) † | 8.377 (ref) † | 5.767 (+45.26% speed) † | 10.804 (-22.46% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 1.422 (+600.66% speed) † | 1.244 (+701.09% speed) † | 27.442 (-63.68% speed) † | 9.967 (ref) | 2.814 (+254.19% speed) † | 3.613 (+175.88% speed) † |
| threads-sustained-v2/threads/wait-timeout | 1.275 (+3920.55% speed) † | 1.175 (+4260.46% speed) † | 350.39 (-85.37% speed) † | 51.254 (ref) † | 202.61 (-74.70% speed) † | 71.014 (-27.83% speed) † |
| generic-calls-v1/core/generic-calls | 1.215 (+94.04% speed) | 1.206 (+95.47% speed) † | 2.446 (-3.60% speed) † | 2.358 (ref) | 2.395 (-1.56% speed) | 1.214 (+94.13% speed) † |

## Interpreter execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

The reference runtime is **wazero**. GC rows use **WasmEdge**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 3,937.50 (-42.32% speed) † | 4,896.00 (-53.62% speed) † | 2,271.00 (ref) † | NM | NM | 6,375.00 (-64.38% speed) † |
| original22-v1/core/integer-loop | 38.000 (+162.78% speed) † | 31.743 (+214.57% speed) † | 99.855 (ref) † | NM | NM | 59.209 (+68.65% speed) † |
| original22-v1/core/function-calls | 59.738 (+91.07% speed) † | 55.672 (+105.03% speed) † | 114.14 (ref) † | NM | NM | 74.510 (+53.19% speed) † |
| original22-v1/core/memory-load-store | 78.713 (+121.16% speed) † | 69.944 (+148.89% speed) † | 174.08 (ref) † | NM | NM | 108.69 (+60.16% speed) † |
| original22-v1/memory/narrow-unaligned | 102.28 (+114.11% speed) † | 92.447 (+136.89% speed) † | 219.00 (ref) † | NM | NM | 135.66 (+61.43% speed) † |
| original22-v1/memory/dependent-address | 47.886 (+137.07% speed) † | 41.759 (+171.85% speed) † | 113.52 (ref) † | NM | NM | 63.199 (+79.63% speed) † |
| original22-v1/memory/float64-unaligned | 80.301 (+103.08% speed) | 71.525 (+128.00% speed) † | 163.08 (ref) | NM | NM | 99.897 (+63.25% speed) † |
| original22-v1/runtime/simd-memory-full | 93.436 (+96.50% speed) † | 84.840 (+116.41% speed) † | 183.60 (ref) † | NM | NM | 95.950 (+91.35% speed) † |
| original22-v1/runtime/simd-memory-widen | 72.369 (+128.86% speed) † | 63.773 (+159.71% speed) † | 165.63 (ref) † | NM | NM | 70.052 (+136.43% speed) † |
| original22-v1/runtime/simd-memory-lanes | 114.49 (+101.35% speed) † | 104.87 (+119.82% speed) † | 230.52 (ref) † | NM | NM | 109.30 (+110.91% speed) † |
| original22-v1/gc/i31-roundtrip (vs WasmEdge) | 43.748 (+22.51% speed) † | 39.552 (+35.51% speed) † | US | NM | NM | 53.596 (ref) † |
| original22-v1/gc/struct-access (vs WasmEdge) | 48.679 (+32.13% speed) † | 46.204 (+39.21% speed) † | US | NM | NM | 64.321 (ref) † |
| original22-v1/gc/struct-allocation (vs WasmEdge) | 105.65 (+20.38% speed) † | 115.42 (+10.19% speed) † | US | NM | NM | 127.18 (ref) † |
| original22-v1/gc/struct-indirect-access (vs WasmEdge) | 87.483 (+28.05% speed) † | 83.390 (+34.34% speed) † | US | NM | NM | 112.03 (ref) † |
| original22-v1/threads/private-memory | 46.775 (+59.24% speed) † | 40.110 (+85.70% speed) † | 74.485 (ref) | NM | NM | 46.715 (+59.45% speed) † |
| original22-v1/threads/shared-memory | 87.271 (-17.36% speed) † | 40.367 (+78.65% speed) † | 72.117 (ref) † | NM | NM | 44.573 (+61.79% speed) † |
| original22-v1/threads/atomic-rmw | 61.367 (-10.22% speed) † | 38.694 (+42.39% speed) † | 55.098 (ref) † | NM | NM | 38.117 (+44.55% speed) † |
| original22-v1/threads/notify-empty | 40.675 (+28.37% speed) † | 36.381 (+43.51% speed) † | 52.212 (ref) † | NM | NM | 53.858 (-3.06% speed) † |
| original22-v1/threads/wait-mismatch | 69.004 (+20.67% speed) † | 48.485 (+71.74% speed) † | 83.269 (ref) † | NM | NM | 57.419 (+45.02% speed) † |
| original22-v1/threads/wait-timeout | 68.002 (+629.85% speed) † | 50.196 (+888.76% speed) † | 496.31 (ref) † | NM | NM | 121.33 (+309.06% speed) † |
| original22-v1/applications/sha256-clang | 21,438.31 (+122.53% speed) † | 19,069.73 (+150.16% speed) † | 47,705.74 (ref) † | NM | NM | 30,067.10 (+58.66% speed) † |
| original22-v1/applications/sha256-rust | 20,876.37 (+126.38% speed) † | 18,400.94 (+156.83% speed) † | 47,259.65 (ref) † | NM | NM | 28,694.83 (+64.70% speed) † |
| gc-sustained-v1/gc/i31-roundtrip (vs WasmEdge) | 40.726 (+23.77% speed) † | 37.534 (+34.30% speed) † | US | NM | NM | 50.409 (ref) † |
| gc-sustained-v1/gc/struct-access (vs WasmEdge) | 47.412 (+36.62% speed) | 46.232 (+40.11% speed) † | US | NM | NM | 64.774 (ref) |
| gc-sustained-v1/gc/struct-allocation (vs WasmEdge) | 121.49 (+2.87% speed) † | 123.66 (+1.07% speed) † | US | NM | NM | 124.98 (ref) † |
| gc-sustained-v1/gc/struct-indirect-access (vs WasmEdge) | 85.508 (+32.19% speed) † | 81.116 (+39.34% speed) | US | NM | NM | 113.03 (ref) |
| threads-sustained-v2/threads/private-memory | 45.751 (+64.71% speed) † | 40.351 (+86.75% speed) † | 75.356 (ref) † | NM | NM | 45.851 (+64.35% speed) † |
| threads-sustained-v2/threads/shared-memory | 90.740 (-15.54% speed) † | 40.321 (+90.07% speed) † | 76.637 (ref) † | NM | NM | 47.313 (+61.98% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 60.346 (-9.97% speed) † | 38.910 (+39.63% speed) † | 54.332 (ref) † | NM | NM | 37.981 (+43.05% speed) † |
| threads-sustained-v2/threads/notify-empty | 36.276 (+40.31% speed) † | 33.372 (+52.52% speed) † | 50.900 (ref) † | NM | NM | 47.436 (+7.30% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 63.588 (+20.57% speed) † | 40.604 (+88.82% speed) | 76.668 (ref) | NM | NM | 48.375 (+58.49% speed) † |
| threads-sustained-v2/threads/wait-timeout | 65.817 (+534.75% speed) † | 41.083 (+916.91% speed) † | 417.78 (ref) † | NM | NM | 116.22 (+259.48% speed) † |
| generic-calls-v1/core/generic-calls | 127.64 (+76.17% speed) † | 109.52 (+105.31% speed) † | 224.86 (ref) † | NM | NM | 139.77 (+60.88% speed) † |

### Warm speed against the fastest observed competitor

Each row selects the fastest measured competitor median for that backend and workload. The observed percentages remain descriptive when the selected reference or candidate has unresolved controls.

#### Native

| Workload | Fastest observed competitor | wasmvm Go speed Δ | wasmvm LLGo speed Δ |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-compiler | -77.90% † | -75.50% † |
| original22-v1/core/integer-loop | wasmedge-compiler | -1.71% † | -2.32% † |
| original22-v1/core/function-calls | wasmer-compiler | -2.18% † | -1.65% † |
| original22-v1/core/memory-load-store | wasmedge-compiler | -35.68% † | -34.48% † |
| original22-v1/memory/narrow-unaligned | wasmedge-compiler | -21.28% † | -21.34% † |
| original22-v1/memory/dependent-address | wasmedge-compiler | -22.17% † | -24.67% † |
| original22-v1/memory/float64-unaligned | wasmedge-compiler | -9.91% † | -8.98% † |
| original22-v1/runtime/simd-memory-full | wasmer-compiler | +3.19% † | +3.33% † |
| original22-v1/runtime/simd-memory-widen | wasmtime-compiler | -1.91% † | -0.50% † |
| original22-v1/runtime/simd-memory-lanes | wasmedge-compiler | +1.39% † | -0.84% |
| original22-v1/gc/i31-roundtrip | wasmedge-compiler | -13.40% † | +8.23% † |
| original22-v1/gc/struct-access | wasmtime-compiler | -37.33% † | -32.70% † |
| original22-v1/gc/struct-allocation | wasmedge-compiler | +42.29% † | +35.23% † |
| original22-v1/gc/struct-indirect-access | wasmtime-compiler | -3.49% † | +19.94% † |
| original22-v1/threads/private-memory | wasmtime-compiler | -3.83% † | +62.17% † |
| original22-v1/threads/shared-memory | wasmtime-compiler | -47.69% † | -28.00% † |
| original22-v1/threads/atomic-rmw | wasmtime-compiler | -31.55% † | -17.07% † |
| original22-v1/threads/notify-empty | wasmer-compiler | +137.43% † | +270.02% † |
| original22-v1/threads/wait-mismatch | wasmer-compiler | +8.20% † | +63.26% † |
| original22-v1/threads/wait-timeout | wasmtime-compiler | +1464.88% † | +1481.55% † |
| original22-v1/applications/sha256-clang | wasmedge-compiler | -12.83% † | -16.32% † |
| original22-v1/applications/sha256-rust | wasmtime-compiler | -12.47% † | -10.65% † |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-compiler | -4.39% † | -4.15% † |
| gc-sustained-v1/gc/struct-access | wasmtime-compiler | -4.62% † | -0.49% † |
| gc-sustained-v1/gc/struct-allocation | wasmtime-compiler | +55.93% † | +52.66% † |
| gc-sustained-v1/gc/struct-indirect-access | wasmtime-compiler | +52.93% † | +57.32% † |
| threads-sustained-v2/threads/private-memory | wasmer-compiler | +130.46% † | +208.33% † |
| threads-sustained-v2/threads/shared-memory | wasmer-compiler | -21.69% † | -29.21% † |
| threads-sustained-v2/threads/atomic-rmw | wasmtime-compiler | -7.50% † | -1.33% † |
| threads-sustained-v2/threads/notify-empty | wasmer-compiler | +342.60% † | +401.72% † |
| threads-sustained-v2/threads/wait-mismatch | wasmer-compiler | +97.82% † | +126.17% † |
| threads-sustained-v2/threads/wait-timeout | wasmtime-compiler | +3920.55% † | +4260.46% † |
| generic-calls-v1/core/generic-calls | wasmedge-compiler | -0.05% † | +0.69% † |

#### Interpreter

| Workload | Fastest observed competitor | wasmvm Go speed Δ | wasmvm LLGo speed Δ |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-interpreter | -42.32% † | -53.62% † |
| original22-v1/core/integer-loop | wasmedge-interpreter | +55.81% † | +86.53% † |
| original22-v1/core/function-calls | wasmedge-interpreter | +24.73% † | +33.84% † |
| original22-v1/core/memory-load-store | wasmedge-interpreter | +38.09% † | +55.40% † |
| original22-v1/memory/narrow-unaligned | wasmedge-interpreter | +32.63% † | +46.74% † |
| original22-v1/memory/dependent-address | wasmedge-interpreter | +31.98% † | +51.34% † |
| original22-v1/memory/float64-unaligned | wasmedge-interpreter | +24.40% † | +39.67% † |
| original22-v1/runtime/simd-memory-full | wasmedge-interpreter | +2.69% † | +13.10% † |
| original22-v1/runtime/simd-memory-widen | wasmedge-interpreter | -3.20% | +9.85% |
| original22-v1/runtime/simd-memory-lanes | wasmedge-interpreter | -4.53% † | +4.23% † |
| original22-v1/gc/i31-roundtrip | wasmedge-interpreter | +22.51% † | +35.51% † |
| original22-v1/gc/struct-access | wasmedge-interpreter | +32.13% † | +39.21% † |
| original22-v1/gc/struct-allocation | wasmedge-interpreter | +20.38% † | +10.19% † |
| original22-v1/gc/struct-indirect-access | wasmedge-interpreter | +28.05% † | +34.34% † |
| original22-v1/threads/private-memory | wasmedge-interpreter | -0.13% † | +16.47% † |
| original22-v1/threads/shared-memory | wasmedge-interpreter | -48.93% † | +10.42% † |
| original22-v1/threads/atomic-rmw | wasmedge-interpreter | -37.89% † | -1.49% † |
| original22-v1/threads/notify-empty | wazero-interpreter | +28.37% † | +43.51% † |
| original22-v1/threads/wait-mismatch | wasmedge-interpreter | -16.79% † | +18.42% † |
| original22-v1/threads/wait-timeout | wasmedge-interpreter | +78.42% † | +141.71% † |
| original22-v1/applications/sha256-clang | wasmedge-interpreter | +40.25% † | +57.67% † |
| original22-v1/applications/sha256-rust | wasmedge-interpreter | +37.45% † | +55.94% † |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-interpreter | +23.77% † | +34.30% † |
| gc-sustained-v1/gc/struct-access | wasmedge-interpreter | +36.62% | +40.11% † |
| gc-sustained-v1/gc/struct-allocation | wasmedge-interpreter | +2.87% † | +1.07% † |
| gc-sustained-v1/gc/struct-indirect-access | wasmedge-interpreter | +32.19% † | +39.34% |
| threads-sustained-v2/threads/private-memory | wasmedge-interpreter | +0.22% † | +13.63% † |
| threads-sustained-v2/threads/shared-memory | wasmedge-interpreter | -47.86% † | +17.34% † |
| threads-sustained-v2/threads/atomic-rmw | wasmedge-interpreter | -37.06% † | -2.39% † |
| threads-sustained-v2/threads/notify-empty | wasmedge-interpreter | +30.76% † | +42.15% † |
| threads-sustained-v2/threads/wait-mismatch | wasmedge-interpreter | -23.92% † | +19.14% † |
| threads-sustained-v2/threads/wait-timeout | wasmedge-interpreter | +76.57% † | +182.88% † |
| generic-calls-v1/core/generic-calls | wasmedge-interpreter | +9.50% † | +27.62% † |

## Complete P3 HTTP requests

P3 competitor cells marked `NM` have no measured adapter for the pinned component/WASI workload. Wasmtime uses the same validated workload. Managed allocation counts exclude foreign Rust/C/JIT allocations and cover only the managed share of allocation.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU, us/request | 72.707 (-40.01% speed) † | 206.18 (-78.84% speed) † | NM | 43.620 (ref) † | NM | NM |
| 16 B instructions/request | 684,521.06 (+203.82% Δ) † | 1,274,348.71 (+465.61% Δ) † | NM | 225,307.13 (ref) † | NM | NM |
| 16 B requests/s | 54,254.48 (-4.21% speed) † | 22,603.74 (-60.09% speed) † | NM | 56,636.99 (ref) † | NM | NM |
| 65536 B CPU, us/request | 108.74 (-45.64% speed) † | 244.47 (-75.82% speed) † | NM | 59.112 (ref) † | NM | NM |
| 65536 B instructions/request | 970,129.35 (+202.82% Δ) † | 1,597,952.35 (+398.80% Δ) † | NM | 320,361.46 (ref) † | NM | NM |
| 65536 B requests/s | 35,268.67 (-21.74% speed) † | 17,167.43 (-61.90% speed) † | NM | 45,063.48 (ref) † | NM | NM |

### P3 source contrasts

| Metric | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| 16 B cpuUs | 0.976 [0.961, 0.992]; Δ -2.38%; speed +2.43% [+0.82%, +4.07%]; unresolved | 0.996 [0.990, 1.002]; Δ -0.39%; speed +0.39% [-0.19%, +0.97%]; pass |
| 16 B instructions | 0.970 [0.968, 0.972]; Δ -3.03%; pass | 0.976 [0.972, 0.979]; Δ -2.41%; pass |
| 16 B throughput | 1.033 [1.007, 1.058]; Δ +3.26%; speed +3.26% [+0.69%, +5.79%]; pass | 1.003 [0.980, 1.020]; Δ +0.34%; speed +0.34% [-1.96%, +2.04%]; unresolved |
| 16 B allocsPerRequest | 0.979 [0.979, 0.979]; Δ -2.12%; pass | 0.982 [0.982, 0.982]; Δ -1.85%; pass |
| 16 B bytesPerRequest | 0.964 [0.964, 0.964]; Δ -3.56%; pass | 0.967 [0.967, 0.967]; Δ -3.26%; pass |
| 65536 B cpuUs | 0.980 [0.966, 1.001]; Δ -1.99%; speed +2.03% [-0.12%, +3.50%]; pass | 0.971 [0.962, 0.980]; Δ -2.87%; speed +2.95% [+2.07%, +3.91%]; unresolved |
| 65536 B instructions | 0.982 [0.980, 0.984]; Δ -1.84%; pass | 0.973 [0.966, 0.980]; Δ -2.74%; unresolved |
| 65536 B throughput | 1.019 [0.999, 1.042]; Δ +1.89%; speed +1.89% [-0.15%, +4.20%]; unresolved | 1.019 [1.000, 1.035]; Δ +1.86%; speed +1.86% [-0.01%, +3.54%]; unresolved |
| 65536 B allocsPerRequest | 0.980 [0.980, 0.980]; Δ -1.96%; pass | 0.983 [0.983, 0.983]; Δ -1.65%; pass |
| 65536 B bytesPerRequest | 0.986 [0.986, 0.986]; Δ -1.37%; pass | 0.986 [0.986, 0.986]; Δ -1.39%; pass |

Throughput ratios above one improve. CPU, instruction and allocation ratios below one improve. The JSON retains all own-baseline observations and every per-role A/A interval.

### Competitive P3 ratios

| Payload / compiler | CPU ratio to Wasmtime [95% CI]; controls | Throughput ratio to Wasmtime [95% CI]; controls |
| --- | ---: | ---: |
| 16 B go | 1.741 [1.598, 1.900]; Δ +74.11%; speed -42.56% [-47.37%, -37.43%]; unresolved | 0.949 [0.909, 0.989]; Δ -5.10%; speed -5.10% [-9.09%, -1.12%]; unresolved |
| 16 B llgo | 4.906 [4.503, 5.359]; Δ +390.59%; speed -79.62% [-81.34%, -77.79%]; unresolved | 0.385 [0.372, 0.392]; Δ -61.53%; speed -61.53% [-62.78%, -60.75%]; unresolved |
| 65536 B go | 1.841 [1.699, 1.974]; Δ +84.09%; speed -45.68% [-49.33%, -41.14%]; unresolved | 0.802 [0.761, 0.855]; Δ -19.81%; speed -19.81% [-23.94%, -14.55%]; unresolved |
| 65536 B llgo | 4.252 [3.844, 4.737]; Δ +325.24%; speed -76.48% [-78.89%, -73.98%]; unresolved | 0.386 [0.377, 0.395]; Δ -61.40%; speed -61.40% [-62.25%, -60.47%]; unresolved |

## Core source baseline

Head/base ratios below one improve. These marginal intervals describe this event; an apparent change in an unchanged mechanism requires diagnosis before being attributed to source changes. The machine-readable report includes every competitor ratio and A/A interval.

### native

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 1.067 [0.826, 1.429]; Δ +6.69%; speed -6.27% [-30.01%, +21.05%]; unresolved | 1.060 [0.909, 1.282]; Δ +6.02%; speed -5.68% [-22.00%, +9.98%]; unresolved |
| original22-v1/core/integer-loop | 0.988 [0.972, 1.004]; Δ -1.18%; speed +1.20% [-0.39%, +2.84%]; unresolved | 0.988 [0.974, 1.001]; Δ -1.22%; speed +1.24% [-0.06%, +2.68%]; unresolved |
| original22-v1/core/function-calls | 0.997 [0.980, 1.007]; Δ -0.30%; speed +0.30% [-0.72%, +2.05%]; unresolved | 1.003 [0.972, 1.034]; Δ +0.35%; speed -0.35% [-3.28%, +2.83%]; unresolved |
| original22-v1/core/memory-load-store | 1.086 [1.008, 1.193]; Δ +8.60%; speed -7.92% [-16.14%, -0.82%]; unresolved | 0.968 [0.906, 1.013]; Δ -3.23%; speed +3.33% [-1.24%, +10.42%]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.949 [0.842, 1.032]; Δ -5.12%; speed +5.39% [-3.09%, +18.83%]; unresolved | 0.985 [0.944, 1.021]; Δ -1.48%; speed +1.50% [-2.10%, +5.89%]; unresolved |
| original22-v1/memory/dependent-address | 0.994 [0.984, 1.004]; Δ -0.57%; speed +0.57% [-0.41%, +1.64%]; unresolved | 0.985 [0.965, 1.000]; Δ -1.48%; speed +1.50% [+0.00%, +3.62%]; unresolved |
| original22-v1/memory/float64-unaligned | 0.987 [0.958, 1.010]; Δ -1.32%; speed +1.34% [-1.03%, +4.35%]; unresolved | 0.976 [0.929, 1.008]; Δ -2.39%; speed +2.45% [-0.78%, +7.67%]; unresolved |
| original22-v1/runtime/simd-memory-full | 0.998 [0.992, 1.007]; Δ -0.15%; speed +0.15% [-0.69%, +0.86%]; unresolved | 0.996 [0.988, 1.006]; Δ -0.41%; speed +0.42% [-0.58%, +1.24%]; unresolved |
| original22-v1/runtime/simd-memory-widen | 1.018 [0.992, 1.061]; Δ +1.79%; speed -1.76% [-5.76%, +0.81%]; unresolved | 1.006 [0.984, 1.043]; Δ +0.62%; speed -0.62% [-4.16%, +1.67%]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 0.996 [0.986, 1.006]; Δ -0.38%; speed +0.38% [-0.57%, +1.38%]; unresolved | 0.993 [0.989, 0.997]; Δ -0.69%; speed +0.69% [+0.34%, +1.11%]; pass |
| original22-v1/gc/i31-roundtrip | 1.031 [0.959, 1.122]; Δ +3.12%; speed -3.02% [-10.84%, +4.27%]; unresolved | 1.009 [0.973, 1.054]; Δ +0.94%; speed -0.94% [-5.14%, +2.76%]; unresolved |
| original22-v1/gc/struct-access | 0.991 [0.948, 1.036]; Δ -0.87%; speed +0.88% [-3.51%, +5.48%]; unresolved | 1.036 [0.937, 1.142]; Δ +3.57%; speed -3.44% [-12.44%, +6.75%]; unresolved |
| original22-v1/gc/struct-allocation | 0.965 [0.825, 1.094]; Δ -3.52%; speed +3.64% [-8.62%, +21.19%]; unresolved | 1.034 [0.984, 1.083]; Δ +3.36%; speed -3.25% [-7.62%, +1.62%]; unresolved |
| original22-v1/gc/struct-indirect-access | 0.961 [0.920, 1.017]; Δ -3.86%; speed +4.02% [-1.64%, +8.75%]; unresolved | 1.020 [1.004, 1.035]; Δ +1.98%; speed -1.94% [-3.36%, -0.41%]; unresolved |
| original22-v1/threads/private-memory | 1.000 [0.900, 1.114]; Δ +0.05%; speed -0.05% [-10.21%, +11.08%]; unresolved | 0.884 [0.726, 1.024]; Δ -11.58%; speed +13.10% [-2.32%, +37.71%]; unresolved |
| original22-v1/threads/shared-memory | 1.055 [0.971, 1.164]; Δ +5.49%; speed -5.21% [-14.12%, +2.97%]; unresolved | 1.107 [0.990, 1.257]; Δ +10.72%; speed -9.68% [-20.45%, +1.04%]; unresolved |
| original22-v1/threads/atomic-rmw | 1.021 [0.931, 1.127]; Δ +2.09%; speed -2.05% [-11.28%, +7.40%]; unresolved | 0.985 [0.949, 1.014]; Δ -1.47%; speed +1.49% [-1.36%, +5.34%]; unresolved |
| original22-v1/threads/notify-empty | 1.118 [0.922, 1.323]; Δ +11.76%; speed -10.52% [-24.42%, +8.49%]; unresolved | 1.106 [0.905, 1.376]; Δ +10.61%; speed -9.59% [-27.31%, +10.53%]; unresolved |
| original22-v1/threads/wait-mismatch | 0.970 [0.858, 1.093]; Δ -2.95%; speed +3.04% [-8.55%, +16.52%]; unresolved | 1.051 [0.997, 1.129]; Δ +5.14%; speed -4.89% [-11.42%, +0.31%]; unresolved |
| original22-v1/threads/wait-timeout | 0.991 [0.880, 1.115]; Δ -0.93%; speed +0.94% [-10.32%, +13.61%]; unresolved | 1.021 [0.778, 1.277]; Δ +2.12%; speed -2.08% [-21.70%, +28.62%]; unresolved |
| original22-v1/applications/sha256-clang | 0.983 [0.919, 1.039]; Δ -1.74%; speed +1.77% [-3.74%, +8.77%]; unresolved | 0.963 [0.884, 1.024]; Δ -3.66%; speed +3.80% [-2.37%, +13.07%]; unresolved |
| original22-v1/applications/sha256-rust | 1.027 [0.972, 1.106]; Δ +2.65%; speed -2.59% [-9.56%, +2.88%]; unresolved | 1.046 [0.998, 1.139]; Δ +4.56%; speed -4.36% [-12.24%, +0.23%]; unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 0.972 [0.947, 0.994]; Δ -2.77%; speed +2.85% [+0.64%, +5.64%]; unresolved | 0.992 [0.953, 1.022]; Δ -0.75%; speed +0.76% [-2.13%, +4.88%]; unresolved |
| gc-sustained-v1/gc/struct-access | 1.003 [0.985, 1.020]; Δ +0.28%; speed -0.28% [-2.01%, +1.57%]; unresolved | 1.001 [0.986, 1.023]; Δ +0.13%; speed -0.13% [-2.25%, +1.39%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 0.981 [0.953, 1.007]; Δ -1.90%; speed +1.94% [-0.72%, +4.98%]; unresolved | 0.922 [0.901, 0.939]; Δ -7.77%; speed +8.43% [+6.53%, +10.96%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.024 [0.977, 1.082]; Δ +2.43%; speed -2.37% [-7.54%, +2.37%]; unresolved | 0.994 [0.974, 1.015]; Δ -0.63%; speed +0.64% [-1.44%, +2.62%]; unresolved |
| threads-sustained-v2/threads/private-memory | 0.971 [0.813, 1.094]; Δ -2.87%; speed +2.95% [-8.60%, +23.01%]; unresolved | 0.966 [0.882, 1.055]; Δ -3.40%; speed +3.52% [-5.24%, +13.33%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 0.919 [0.851, 0.990]; Δ -8.11%; speed +8.82% [+0.99%, +17.45%]; unresolved | 1.028 [0.936, 1.165]; Δ +2.85%; speed -2.77% [-14.17%, +6.88%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.984 [0.955, 1.012]; Δ -1.64%; speed +1.67% [-1.16%, +4.72%]; unresolved | 0.994 [0.976, 1.009]; Δ -0.64%; speed +0.64% [-0.90%, +2.45%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 0.958 [0.805, 1.064]; Δ -4.23%; speed +4.42% [-6.03%, +24.21%]; unresolved | 1.015 [0.941, 1.110]; Δ +1.54%; speed -1.52% [-9.95%, +6.22%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 0.975 [0.937, 1.006]; Δ -2.51%; speed +2.58% [-0.55%, +6.67%]; unresolved | 1.032 [0.974, 1.129]; Δ +3.21%; speed -3.11% [-11.44%, +2.64%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 1.051 [0.998, 1.105]; Δ +5.09%; speed -4.84% [-9.50%, +0.22%]; unresolved | 1.060 [0.979, 1.209]; Δ +6.03%; speed -5.68% [-17.28%, +2.16%]; unresolved |
| generic-calls-v1/core/generic-calls | 1.007 [0.998, 1.017]; Δ +0.66%; speed -0.65% [-1.71%, +0.16%]; pass | 1.007 [0.994, 1.016]; Δ +0.71%; speed -0.70% [-1.59%, +0.61%]; unresolved |

### interpreter

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 0.944 [0.832, 1.048]; Δ -5.64%; speed +5.98% [-4.60%, +20.14%]; unresolved | 1.021 [0.836, 1.286]; Δ +2.13%; speed -2.09% [-22.26%, +19.63%]; unresolved |
| original22-v1/core/integer-loop | 1.002 [0.980, 1.029]; Δ +0.18%; speed -0.18% [-2.82%, +2.08%]; unresolved | 1.039 [1.017, 1.062]; Δ +3.95%; speed -3.80% [-5.87%, -1.68%]; unresolved |
| original22-v1/core/function-calls | 0.994 [0.954, 1.026]; Δ -0.63%; speed +0.63% [-2.56%, +4.78%]; unresolved | 0.923 [0.885, 0.961]; Δ -7.71%; speed +8.35% [+4.07%, +13.04%]; unresolved |
| original22-v1/core/memory-load-store | 0.997 [0.973, 1.027]; Δ -0.34%; speed +0.34% [-2.66%, +2.75%]; unresolved | 0.992 [0.979, 1.004]; Δ -0.83%; speed +0.84% [-0.43%, +2.16%]; unresolved |
| original22-v1/memory/narrow-unaligned | 1.026 [0.984, 1.096]; Δ +2.58%; speed -2.52% [-8.80%, +1.63%]; unresolved | 1.035 [1.009, 1.073]; Δ +3.53%; speed -3.41% [-6.83%, -0.90%]; unresolved |
| original22-v1/memory/dependent-address | 0.985 [0.963, 1.002]; Δ -1.46%; speed +1.48% [-0.20%, +3.87%]; unresolved | 1.013 [0.995, 1.045]; Δ +1.30%; speed -1.28% [-4.27%, +0.52%]; unresolved |
| original22-v1/memory/float64-unaligned | 0.987 [0.981, 0.992]; Δ -1.29%; speed +1.30% [+0.82%, +1.89%]; pass | 0.991 [0.978, 1.003]; Δ -0.90%; speed +0.91% [-0.29%, +2.28%]; unresolved |
| original22-v1/runtime/simd-memory-full | 0.981 [0.918, 1.025]; Δ -1.91%; speed +1.94% [-2.47%, +8.93%]; unresolved | 1.014 [0.999, 1.033]; Δ +1.45%; speed -1.43% [-3.16%, +0.15%]; unresolved |
| original22-v1/runtime/simd-memory-widen | 0.993 [0.979, 1.005]; Δ -0.73%; speed +0.74% [-0.53%, +2.19%]; unresolved | 0.984 [0.969, 0.998]; Δ -1.60%; speed +1.62% [+0.23%, +3.17%]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 1.009 [0.987, 1.028]; Δ +0.88%; speed -0.87% [-2.77%, +1.37%]; unresolved | 0.827 [0.755, 0.871]; Δ -17.26%; speed +20.86% [+14.79%, +32.49%]; unresolved |
| original22-v1/gc/i31-roundtrip | 0.977 [0.959, 0.996]; Δ -2.29%; speed +2.34% [+0.36%, +4.26%]; unresolved | 1.078 [1.020, 1.175]; Δ +7.79%; speed -7.22% [-14.86%, -1.96%]; unresolved |
| original22-v1/gc/struct-access | 0.977 [0.954, 1.002]; Δ -2.34%; speed +2.39% [-0.17%, +4.81%]; unresolved | 0.991 [0.965, 1.017]; Δ -0.88%; speed +0.89% [-1.71%, +3.62%]; unresolved |
| original22-v1/gc/struct-allocation | 0.918 [0.811, 1.002]; Δ -8.19%; speed +8.92% [-0.25%, +23.33%]; unresolved | 0.984 [0.951, 1.005]; Δ -1.62%; speed +1.65% [-0.45%, +5.20%]; unresolved |
| original22-v1/gc/struct-indirect-access | 1.008 [0.988, 1.035]; Δ +0.80%; speed -0.79% [-3.39%, +1.19%]; unresolved | 1.011 [0.991, 1.028]; Δ +1.06%; speed -1.05% [-2.71%, +0.93%]; unresolved |
| original22-v1/threads/private-memory | 1.000 [0.969, 1.034]; Δ -0.01%; speed +0.01% [-3.30%, +3.21%]; unresolved | 0.997 [0.966, 1.028]; Δ -0.32%; speed +0.32% [-2.68%, +3.54%]; unresolved |
| original22-v1/threads/shared-memory | 1.028 [1.008, 1.055]; Δ +2.78%; speed -2.71% [-5.23%, -0.76%]; unresolved | 0.997 [0.978, 1.019]; Δ -0.25%; speed +0.26% [-1.85%, +2.24%]; unresolved |
| original22-v1/threads/atomic-rmw | 1.011 [0.994, 1.029]; Δ +1.10%; speed -1.08% [-2.81%, +0.59%]; unresolved | 1.005 [0.986, 1.023]; Δ +0.54%; speed -0.54% [-2.21%, +1.40%]; unresolved |
| original22-v1/threads/notify-empty | 0.950 [0.893, 1.014]; Δ -4.98%; speed +5.24% [-1.39%, +11.94%]; unresolved | 1.000 [0.919, 1.078]; Δ +0.00%; speed +0.00% [-7.26%, +8.77%]; unresolved |
| original22-v1/threads/wait-mismatch | 0.965 [0.894, 1.036]; Δ -3.50%; speed +3.63% [-3.52%, +11.89%]; unresolved | 1.118 [0.981, 1.333]; Δ +11.77%; speed -10.53% [-24.98%, +1.92%]; unresolved |
| original22-v1/threads/wait-timeout | 1.012 [0.912, 1.156]; Δ +1.20%; speed -1.18% [-13.52%, +9.64%]; unresolved | 1.127 [0.993, 1.294]; Δ +12.72%; speed -11.28% [-22.72%, +0.67%]; unresolved |
| original22-v1/applications/sha256-clang | 0.993 [0.972, 1.008]; Δ -0.69%; speed +0.69% [-0.78%, +2.85%]; unresolved | 1.012 [0.975, 1.054]; Δ +1.25%; speed -1.23% [-5.16%, +2.57%]; unresolved |
| original22-v1/applications/sha256-rust | 1.005 [0.984, 1.035]; Δ +0.49%; speed -0.49% [-3.39%, +1.58%]; unresolved | 1.011 [1.002, 1.020]; Δ +1.09%; speed -1.08% [-1.98%, -0.16%]; unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 1.002 [0.975, 1.044]; Δ +0.20%; speed -0.20% [-4.24%, +2.52%]; unresolved | 1.032 [1.021, 1.044]; Δ +3.22%; speed -3.12% [-4.20%, -2.08%]; unresolved |
| gc-sustained-v1/gc/struct-access | 0.983 [0.968, 1.000]; Δ -1.66%; speed +1.69% [-0.05%, +3.26%]; unresolved | 0.990 [0.972, 1.007]; Δ -0.98%; speed +0.99% [-0.71%, +2.84%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 1.054 [1.000, 1.128]; Δ +5.41%; speed -5.14% [-11.35%, -0.02%]; unresolved | 0.990 [0.985, 0.997]; Δ -1.00%; speed +1.01% [+0.33%, +1.56%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.005 [0.993, 1.015]; Δ +0.46%; speed -0.46% [-1.52%, +0.74%]; unresolved | 0.999 [0.986, 1.013]; Δ -0.07%; speed +0.07% [-1.29%, +1.45%]; pass |
| threads-sustained-v2/threads/private-memory | 0.968 [0.896, 1.012]; Δ -3.25%; speed +3.35% [-1.20%, +11.65%]; unresolved | 1.002 [0.972, 1.045]; Δ +0.21%; speed -0.21% [-4.31%, +2.93%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 1.021 [0.986, 1.070]; Δ +2.15%; speed -2.10% [-6.52%, +1.41%]; unresolved | 0.983 [0.946, 1.033]; Δ -1.71%; speed +1.73% [-3.24%, +5.66%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 1.009 [0.968, 1.069]; Δ +0.86%; speed -0.85% [-6.44%, +3.31%]; unresolved | 1.042 [1.001, 1.115]; Δ +4.18%; speed -4.01% [-10.33%, -0.09%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 0.959 [0.924, 0.992]; Δ -4.07%; speed +4.25% [+0.85%, +8.23%]; unresolved | 0.987 [0.942, 1.019]; Δ -1.29%; speed +1.31% [-1.84%, +6.15%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.006 [0.994, 1.022]; Δ +0.60%; speed -0.59% [-2.11%, +0.65%]; unresolved | 0.988 [0.974, 1.001]; Δ -1.24%; speed +1.25% [-0.15%, +2.69%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 0.994 [0.973, 1.021]; Δ -0.65%; speed +0.65% [-2.02%, +2.78%]; unresolved | 0.991 [0.971, 1.012]; Δ -0.87%; speed +0.87% [-1.17%, +2.95%]; unresolved |
| generic-calls-v1/core/generic-calls | 0.966 [0.849, 1.067]; Δ -3.45%; speed +3.57% [-6.32%, +17.72%]; unresolved | 1.020 [0.970, 1.117]; Δ +2.00%; speed -1.96% [-10.51%, +3.14%]; unresolved |

## Preparation, first call and cleanup

All entries below are **microseconds per process or invocation**, including unresolved observations. Phase intervals are not inferred from warm-call controls. Startup is independently observed process readiness. Native compile/validate boundaries differ by API; WasmEdge AOT includes validation and artifact I/O, plus separately recorded native-image loading. Every process creates a fresh native image. First-call and warm execution have their own boundaries; values are never subtracted to invent pure costs.

<details>
<summary>Native: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,736.65 | 7,801.06 | 5,076.94 | 5,519.94 | 7,380.21 | 19,648.23 |
| original22-v1/core/integer-loop | 5,461.79 | 7,261.10 | 5,416.71 | 7,650.46 | 7,354.04 | 19,864.90 |
| original22-v1/core/function-calls | 7,573.29 | 7,021.04 | 5,440.48 | 5,290.54 | 7,793.46 | 20,412.73 |
| original22-v1/core/memory-load-store | 6,508.67 | 8,447.62 | 5,426.67 | 6,037.96 | 7,565.35 | 19,415.98 |
| original22-v1/memory/narrow-unaligned | 5,619.58 | 7,699.00 | 5,728.17 | 5,988.60 | 6,938.46 | 19,945.29 |
| original22-v1/memory/dependent-address | 6,110.92 | 7,829.06 | 6,118.37 | 6,890.08 | 8,001.56 | 20,873.77 |
| original22-v1/memory/float64-unaligned | 4,875.73 | 7,173.17 | 5,866.35 | 5,896.56 | 7,013.75 | 19,344.08 |
| original22-v1/runtime/simd-memory-full | 5,408.37 | 8,581.17 | 5,261.79 | 5,695.04 | 7,372.69 | 21,155.35 |
| original22-v1/runtime/simd-memory-widen | 5,318.40 | 7,528.48 | 5,023.31 | 5,092.46 | 7,108.42 | 19,290.29 |
| original22-v1/runtime/simd-memory-lanes | 5,693.27 | 9,068.08 | 5,907.77 | 6,697.12 | 8,162.60 | 19,773.02 |
| original22-v1/gc/i31-roundtrip | 4,999.48 | 6,859.58 | US | 4,772.98 | US | 19,179.52 |
| original22-v1/gc/struct-access | 4,847.83 | 6,873.63 | US | 5,101.62 | US | 18,545.27 |
| original22-v1/gc/struct-allocation | 5,042.40 | 7,667.13 | US | 4,903.96 | US | 18,721.58 |
| original22-v1/gc/struct-indirect-access | 5,550.92 | 6,903.21 | US | 6,203.52 | US | 19,666.08 |
| original22-v1/threads/private-memory | 5,498.52 | 7,795.67 | 5,412.35 | 4,851.71 | 6,748.15 | 18,860.85 |
| original22-v1/threads/shared-memory | 5,070.88 | 8,218.35 | 4,921.56 | 4,996.02 | 6,862.29 | 19,239.52 |
| original22-v1/threads/atomic-rmw | 5,021.96 | 7,009.15 | 5,468.19 | 4,902.85 | 6,843.42 | 19,023.67 |
| original22-v1/threads/notify-empty | 5,525.27 | 8,756.62 | 5,819.48 | 6,265.29 | 7,959.96 | 20,561.02 |
| original22-v1/threads/wait-mismatch | 5,732.48 | 7,270.12 | 5,663.69 | 5,027.46 | 7,130.92 | 19,262.40 |
| original22-v1/threads/wait-timeout | 5,344.90 | 8,398.56 | 5,835.56 | 6,229.85 | 7,880.67 | 19,992.31 |
| original22-v1/applications/sha256-clang | 5,148.69 | 7,317.73 | 5,313.15 | 5,052.73 | 6,852.42 | 19,518.75 |
| original22-v1/applications/sha256-rust | 5,835.33 | 7,675.40 | 5,625.48 | 5,624.17 | 9,104.58 | 19,915.31 |
| gc-sustained-v1/gc/i31-roundtrip | 5,266.06 | 7,953.77 | US | 5,523.10 | US | 19,256.67 |
| gc-sustained-v1/gc/struct-access | 5,946.29 | 8,531.56 | US | 5,192.17 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 5,589.73 | 7,696.58 | US | 6,434.00 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 5,325.50 | 7,522.85 | US | 5,835.90 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 5,176.29 | 7,193.90 | 5,738.10 | 4,771.38 | 7,242.08 | 19,053.90 |
| threads-sustained-v2/threads/shared-memory | 5,149.31 | 7,490.56 | 5,922.75 | 5,385.38 | 7,539.60 | 20,306.73 |
| threads-sustained-v2/threads/atomic-rmw | 5,165.25 | 8,964.60 | 5,291.04 | 5,117.96 | 6,963.40 | 18,490.96 |
| threads-sustained-v2/threads/notify-empty | 5,133.94 | 7,865.98 | 6,103.65 | 5,222.46 | 7,543.33 | 19,700.31 |
| threads-sustained-v2/threads/wait-mismatch | 5,249.19 | 7,105.17 | 5,602.12 | 4,893.81 | 6,539.58 | 19,017.58 |
| threads-sustained-v2/threads/wait-timeout | 5,498.96 | 7,133.02 | 5,071.48 | 5,265.10 | 7,373.48 | 18,752.75 |
| generic-calls-v1/core/generic-calls | 6,066.96 | 7,938.52 | 5,744.10 | 5,840.67 | 9,359.54 | 21,416.35 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 42.916 | 47.501 | 59.229 | 47.958 | 36.270 | 59.688 |
| original22-v1/core/integer-loop | 50.750 | 54.001 | 48.000 | 75.354 | 63.563 | 56.376 |
| original22-v1/core/function-calls | 58.062 | 53.667 | 49.562 | 36.895 | 49.062 | 49.375 |
| original22-v1/core/memory-load-store | 54.208 | 57.188 | 43.499 | 40.854 | 49.375 | 58.855 |
| original22-v1/memory/narrow-unaligned | 53.354 | 77.084 | 51.562 | 59.604 | 55.270 | 56.542 |
| original22-v1/memory/dependent-address | 55.312 | 51.666 | 50.354 | 45.417 | 43.541 | 53.104 |
| original22-v1/memory/float64-unaligned | 49.250 | 53.874 | 41.209 | 44.812 | 45.084 | 64.271 |
| original22-v1/runtime/simd-memory-full | 47.084 | 63.041 | 52.604 | 43.709 | 47.855 | 50.291 |
| original22-v1/runtime/simd-memory-widen | 53.292 | 56.375 | 47.291 | 43.896 | 41.042 | 61.416 |
| original22-v1/runtime/simd-memory-lanes | 68.874 | 56.938 | 53.000 | 39.750 | 54.126 | 56.999 |
| original22-v1/gc/i31-roundtrip | 47.354 | 52.188 | US | 39.041 | US | 65.125 |
| original22-v1/gc/struct-access | 44.041 | 51.291 | US | 51.271 | US | 44.876 |
| original22-v1/gc/struct-allocation | 46.375 | 50.521 | US | 35.667 | US | 55.166 |
| original22-v1/gc/struct-indirect-access | 44.000 | 46.562 | US | 35.791 | US | 43.500 |
| original22-v1/threads/private-memory | 42.188 | 45.771 | 40.980 | 41.208 | 36.688 | 45.187 |
| original22-v1/threads/shared-memory | 40.749 | 52.084 | 38.959 | 38.438 | 43.605 | 49.916 |
| original22-v1/threads/atomic-rmw | 43.355 | 46.166 | 41.376 | 51.166 | 36.688 | 67.124 |
| original22-v1/threads/notify-empty | 54.230 | 63.125 | 46.000 | 77.938 | 42.854 | 90.688 |
| original22-v1/threads/wait-mismatch | 49.792 | 50.437 | 51.105 | 50.354 | 42.167 | 44.709 |
| original22-v1/threads/wait-timeout | 48.334 | 58.584 | 50.396 | 50.251 | 47.125 | 48.938 |
| original22-v1/applications/sha256-clang | 63.501 | 61.062 | 50.166 | 50.041 | 53.624 | 63.834 |
| original22-v1/applications/sha256-rust | 50.416 | 55.334 | 48.812 | 41.146 | 73.188 | 64.312 |
| gc-sustained-v1/gc/i31-roundtrip | 48.042 | 61.417 | US | 36.145 | US | 50.438 |
| gc-sustained-v1/gc/struct-access | 49.541 | 49.395 | US | 49.603 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 45.542 | 52.188 | US | 42.041 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 46.001 | 56.167 | US | 88.959 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 46.625 | 53.521 | 49.625 | 36.042 | 41.938 | 55.542 |
| threads-sustained-v2/threads/shared-memory | 46.938 | 51.645 | 60.688 | 39.187 | 45.813 | 55.000 |
| threads-sustained-v2/threads/atomic-rmw | 41.605 | 55.312 | 43.688 | 42.459 | 38.875 | 50.834 |
| threads-sustained-v2/threads/notify-empty | 41.354 | 65.854 | 44.938 | 87.563 | 39.334 | 49.541 |
| threads-sustained-v2/threads/wait-mismatch | 42.105 | 48.562 | 48.605 | 34.355 | 41.896 | 48.166 |
| threads-sustained-v2/threads/wait-timeout | 41.792 | 48.375 | 45.125 | 42.625 | 42.874 | 54.209 |
| generic-calls-v1/core/generic-calls | 82.979 | 52.312 | 59.292 | 59.375 | 56.500 | 66.229 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.958 | 5.729 | 71.584 | 134.15 | 50.125 | 38.250 |
| original22-v1/core/integer-loop | 7.667 | 6.208 | 74.042 | 244.67 | 48.208 | 39.770 |
| original22-v1/core/function-calls | 7.833 | 6.479 | 65.354 | 142.67 | 49.166 | 39.584 |
| original22-v1/core/memory-load-store | 6.937 | 6.271 | 67.625 | 148.29 | 51.021 | 41.583 |
| original22-v1/memory/narrow-unaligned | 6.979 | 6.208 | 66.375 | 157.10 | 51.333 | 39.479 |
| original22-v1/memory/dependent-address | 7.438 | 6.208 | 67.041 | 149.60 | 49.166 | 42.625 |
| original22-v1/memory/float64-unaligned | 7.146 | 6.417 | 66.626 | 151.40 | 43.334 | 35.458 |
| original22-v1/runtime/simd-memory-full | 6.792 | 6.938 | 70.709 | 152.96 | 47.397 | 36.730 |
| original22-v1/runtime/simd-memory-widen | 7.354 | 6.000 | 63.041 | 153.33 | 48.854 | 41.334 |
| original22-v1/runtime/simd-memory-lanes | 7.271 | 6.021 | 67.104 | 148.58 | 53.896 | 38.688 |
| original22-v1/gc/i31-roundtrip | 7.354 | 6.125 | US | 152.50 | US | 38.876 |
| original22-v1/gc/struct-access | 6.604 | 5.792 | US | 144.98 | US | 36.145 |
| original22-v1/gc/struct-allocation | 7.000 | 6.188 | US | 147.27 | US | 51.645 |
| original22-v1/gc/struct-indirect-access | 6.812 | 4.771 | US | 143.92 | US | 39.146 |
| original22-v1/threads/private-memory | 7.667 | 5.979 | 62.812 | 152.92 | 51.000 | 38.500 |
| original22-v1/threads/shared-memory | 6.854 | 6.333 | 63.604 | 141.27 | 48.980 | 43.813 |
| original22-v1/threads/atomic-rmw | 7.458 | 5.500 | 65.709 | 144.96 | 45.500 | 40.312 |
| original22-v1/threads/notify-empty | 11.938 | 6.792 | 77.688 | 181.52 | 46.334 | 54.916 |
| original22-v1/threads/wait-mismatch | 7.146 | 6.104 | 86.145 | 147.31 | 45.000 | 35.625 |
| original22-v1/threads/wait-timeout | 7.521 | 6.250 | 69.459 | 147.73 | 52.604 | 39.021 |
| original22-v1/applications/sha256-clang | 8.312 | 6.396 | 67.083 | 150.98 | 47.417 | 42.584 |
| original22-v1/applications/sha256-rust | 7.188 | 5.750 | 66.167 | 144.58 | 51.270 | 48.229 |
| gc-sustained-v1/gc/i31-roundtrip | 7.125 | 6.500 | US | 144.44 | US | 38.126 |
| gc-sustained-v1/gc/struct-access | 6.812 | 5.708 | US | 166.69 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 6.958 | 5.875 | US | 151.77 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 7.896 | 5.708 | US | 197.98 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 6.958 | 6.625 | 67.938 | 147.10 | 48.938 | 45.375 |
| threads-sustained-v2/threads/shared-memory | 7.229 | 6.146 | 66.396 | 150.73 | 47.334 | 39.522 |
| threads-sustained-v2/threads/atomic-rmw | 6.875 | 5.667 | 66.062 | 141.38 | 47.041 | 38.541 |
| threads-sustained-v2/threads/notify-empty | 7.208 | 6.854 | 66.979 | 154.62 | 55.959 | 35.646 |
| threads-sustained-v2/threads/wait-mismatch | 6.667 | 6.500 | 64.103 | 147.38 | 50.395 | 35.458 |
| threads-sustained-v2/threads/wait-timeout | 7.062 | 5.542 | 65.979 | 149.88 | 49.062 | 37.145 |
| generic-calls-v1/core/generic-calls | 7.875 | 6.708 | 74.063 | 175.65 | 83.895 | 188.81 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 215.12 | 167.33 | 109.81 | 1,159.85 | 1,326.50 | 276,372.98 |
| original22-v1/core/integer-loop | 278.96 | 222.33 | 134.96 | 3,659.27 | 1,374.90 | 281,680.35 |
| original22-v1/core/function-calls | 344.31 | 296.27 | 142.21 | 1,187.75 | 1,314.00 | 277,816.62 |
| original22-v1/core/memory-load-store | 390.81 | 334.79 | 167.63 | 1,029.71 | 1,410.44 | 285,015.60 |
| original22-v1/memory/narrow-unaligned | 347.00 | 331.42 | 169.37 | 1,039.23 | 1,334.17 | 277,697.42 |
| original22-v1/memory/dependent-address | 341.71 | 305.96 | 166.96 | 1,268.60 | 1,333.79 | 280,658.29 |
| original22-v1/memory/float64-unaligned | 392.98 | 333.46 | 171.77 | 1,170.96 | 1,349.46 | 281,231.54 |
| original22-v1/runtime/simd-memory-full | 336.12 | 314.10 | 144.71 | 1,552.04 | 1,403.48 | 275,682.02 |
| original22-v1/runtime/simd-memory-widen | 309.83 | 270.02 | 142.79 | 1,269.35 | 1,319.27 | 272,543.75 |
| original22-v1/runtime/simd-memory-lanes | 334.75 | 336.00 | 151.58 | 4,463.21 | 1,429.63 | 280,836.73 |
| original22-v1/gc/i31-roundtrip | 244.12 | 195.35 | US | 1,048.81 | US | 275,191.23 |
| original22-v1/gc/struct-access | 273.31 | 237.54 | US | 1,112.56 | US | 278,920.65 |
| original22-v1/gc/struct-allocation | 258.08 | 214.58 | US | 3,825.81 | US | 276,019.06 |
| original22-v1/gc/struct-indirect-access | 314.29 | 282.40 | US | 8,149.10 | US | 275,504.29 |
| original22-v1/threads/private-memory | 264.44 | 213.71 | 138.94 | 1,004.37 | 1,330.54 | 275,450.56 |
| original22-v1/threads/shared-memory | 269.54 | 217.85 | 138.31 | 858.27 | 1,232.96 | 280,128.15 |
| original22-v1/threads/atomic-rmw | 260.21 | 193.94 | 147.40 | 1,097.29 | 1,243.69 | 274,628.48 |
| original22-v1/threads/notify-empty | 292.27 | 264.15 | 160.90 | 1,331.54 | 1,453.69 | 284,362.75 |
| original22-v1/threads/wait-mismatch | 252.96 | 206.69 | 161.04 | 1,067.98 | 1,250.56 | 278,935.40 |
| original22-v1/threads/wait-timeout | 262.12 | 289.88 | 150.46 | 1,572.00 | 1,883.81 | 281,141.85 |
| original22-v1/applications/sha256-clang | 1,713.77 | 2,073.92 | 354.62 | 4,142.25 | 2,714.60 | 286,205.08 |
| original22-v1/applications/sha256-rust | 1,627.48 | 2,081.10 | 377.02 | 3,045.19 | 2,910.60 | 285,332.02 |
| gc-sustained-v1/gc/i31-roundtrip | 234.69 | 187.75 | US | 865.48 | US | 279,686.02 |
| gc-sustained-v1/gc/struct-access | 268.88 | 226.58 | US | 1,154.96 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 271.04 | 245.21 | US | 6,507.31 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 342.83 | 280.77 | US | 2,501.81 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 262.27 | 212.52 | 141.46 | 1,089.10 | 1,181.54 | 280,391.90 |
| threads-sustained-v2/threads/shared-memory | 288.10 | 215.40 | 151.00 | 1,031.92 | 1,256.69 | 278,531.85 |
| threads-sustained-v2/threads/atomic-rmw | 250.46 | 200.38 | 139.21 | 1,040.56 | 1,224.94 | 280,945.29 |
| threads-sustained-v2/threads/notify-empty | 251.75 | 208.67 | 142.23 | 1,513.29 | 1,641.25 | 274,938.02 |
| threads-sustained-v2/threads/wait-mismatch | 252.67 | 201.85 | 141.83 | 1,146.08 | 1,238.33 | 274,153.00 |
| threads-sustained-v2/threads/wait-timeout | 247.17 | 206.65 | 158.73 | 1,103.40 | 1,277.48 | 276,072.56 |
| generic-calls-v1/core/generic-calls | 350.44 | 324.46 | 161.12 | 2,564.85 | 2,347.00 | 286,314.77 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 9.875 | 10.208 | 15.979 | 8.875 | 20.291 | 17.625 |
| original22-v1/core/integer-loop | 9.084 | 10.208 | 16.438 | 9.896 | 19.979 | 18.021 |
| original22-v1/core/function-calls | 10.562 | 10.938 | 15.646 | 8.376 | 20.229 | 18.605 |
| original22-v1/core/memory-load-store | 15.563 | 16.188 | 20.312 | 23.938 | 37.084 | 39.167 |
| original22-v1/memory/narrow-unaligned | 15.917 | 17.062 | 19.729 | 24.291 | 35.562 | 40.084 |
| original22-v1/memory/dependent-address | 16.687 | 17.708 | 22.375 | 23.958 | 36.979 | 36.687 |
| original22-v1/memory/float64-unaligned | 15.834 | 16.041 | 20.229 | 24.562 | 34.750 | 39.875 |
| original22-v1/runtime/simd-memory-full | 17.395 | 19.666 | 22.375 | 65.812 | 40.000 | 45.583 |
| original22-v1/runtime/simd-memory-widen | 18.208 | 18.125 | 21.730 | 64.105 | 40.896 | 47.875 |
| original22-v1/runtime/simd-memory-lanes | 18.000 | 19.021 | 21.771 | 66.876 | 42.001 | 48.062 |
| original22-v1/gc/i31-roundtrip | 9.584 | 10.729 | US | 8.708 | US | 16.605 |
| original22-v1/gc/struct-access | 10.708 | 11.562 | US | 25.583 | US | 17.438 |
| original22-v1/gc/struct-allocation | 11.209 | 11.916 | US | 65.020 | US | 22.021 |
| original22-v1/gc/struct-indirect-access | 22.459 | 25.250 | US | 78.416 | US | 24.834 |
| original22-v1/threads/private-memory | 15.229 | 15.417 | 18.125 | 24.021 | 36.541 | 41.438 |
| original22-v1/threads/shared-memory | 14.562 | 16.271 | 18.959 | 23.125 | 40.750 | 39.584 |
| original22-v1/threads/atomic-rmw | 15.583 | 15.750 | 19.791 | 24.270 | 38.750 | 38.167 |
| original22-v1/threads/notify-empty | 17.584 | 17.604 | 23.125 | 24.917 | 39.459 | 42.291 |
| original22-v1/threads/wait-mismatch | 16.605 | 15.896 | 19.166 | 25.104 | 39.791 | 40.229 |
| original22-v1/threads/wait-timeout | 16.229 | 18.812 | 21.896 | 24.812 | 41.812 | 37.645 |
| original22-v1/applications/sha256-clang | 21.041 | 21.937 | 18.938 | 64.521 | 41.500 | 48.750 |
| original22-v1/applications/sha256-rust | 22.520 | 22.541 | 22.895 | 61.229 | 44.459 | 47.625 |
| gc-sustained-v1/gc/i31-roundtrip | 9.187 | 11.021 | US | 8.938 | US | 15.562 |
| gc-sustained-v1/gc/struct-access | 11.250 | 10.812 | US | 26.187 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 11.083 | 11.875 | US | 63.375 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 25.105 | 25.791 | US | 78.188 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 15.917 | 16.166 | 22.541 | 23.521 | 35.334 | 42.459 |
| threads-sustained-v2/threads/shared-memory | 16.062 | 16.542 | 22.604 | 24.520 | 39.916 | 40.333 |
| threads-sustained-v2/threads/atomic-rmw | 14.916 | 18.896 | 18.979 | 23.584 | 40.584 | 39.583 |
| threads-sustained-v2/threads/notify-empty | 15.188 | 17.605 | 18.334 | 24.291 | 41.438 | 43.541 |
| threads-sustained-v2/threads/wait-mismatch | 16.208 | 15.437 | 20.791 | 24.438 | 40.000 | 37.959 |
| threads-sustained-v2/threads/wait-timeout | 16.854 | 17.083 | 20.500 | 25.021 | 43.062 | 42.083 |
| generic-calls-v1/core/generic-calls | 11.146 | 12.833 | 16.666 | 10.229 | 23.667 | 19.167 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.333 | 0.521 | 2.521 | 4.708 | 5.312 | 0.292 |
| original22-v1/core/integer-loop | 0.333 | 0.479 | 2.542 | 4.500 | 5.479 | 0.312 |
| original22-v1/core/function-calls | 0.291 | 0.479 | 1.855 | 4.875 | 5.083 | 0.291 |
| original22-v1/core/memory-load-store | 0.292 | 0.500 | 3.125 | 4.458 | 5.396 | 0.250 |
| original22-v1/memory/narrow-unaligned | 0.292 | 0.500 | 2.333 | 5.021 | 5.292 | 0.291 |
| original22-v1/memory/dependent-address | 0.292 | 0.521 | 3.146 | 4.688 | 5.229 | 0.292 |
| original22-v1/memory/float64-unaligned | 0.291 | 0.500 | 2.292 | 4.625 | 5.188 | 0.209 |
| original22-v1/runtime/simd-memory-full | 0.312 | 0.562 | 2.417 | 2.376 | 5.333 | 0.229 |
| original22-v1/runtime/simd-memory-widen | 0.312 | 0.479 | 2.250 | 2.250 | 5.208 | 0.250 |
| original22-v1/runtime/simd-memory-lanes | 0.334 | 0.563 | 3.000 | 2.229 | 5.604 | 0.334 |
| original22-v1/gc/i31-roundtrip | 0.334 | 0.438 | US | 4.458 | US | 0.312 |
| original22-v1/gc/struct-access | 0.271 | 0.375 | US | 4.521 | US | 0.250 |
| original22-v1/gc/struct-allocation | 0.250 | 0.459 | US | 2.021 | US | 0.208 |
| original22-v1/gc/struct-indirect-access | 0.292 | 0.521 | US | 2.188 | US | 0.250 |
| original22-v1/threads/private-memory | 0.292 | 0.520 | 2.208 | 4.396 | 5.208 | 0.354 |
| original22-v1/threads/shared-memory | 0.271 | 0.500 | 2.875 | 4.625 | 5.333 | 0.291 |
| original22-v1/threads/atomic-rmw | 0.292 | 0.521 | 2.812 | 4.708 | 5.354 | 0.271 |
| original22-v1/threads/notify-empty | 0.334 | 0.604 | 2.667 | 4.792 | 5.146 | 0.334 |
| original22-v1/threads/wait-mismatch | 0.312 | 0.479 | 2.979 | 4.750 | 5.333 | 0.312 |
| original22-v1/threads/wait-timeout | 0.334 | 0.604 | 2.833 | 4.583 | 5.125 | 0.250 |
| original22-v1/applications/sha256-clang | 0.292 | 0.500 | 2.667 | 2.333 | 5.875 | 0.291 |
| original22-v1/applications/sha256-rust | 0.292 | 0.500 | 3.000 | 2.312 | 6.083 | 0.291 |
| gc-sustained-v1/gc/i31-roundtrip | 0.271 | 0.500 | US | 4.479 | US | 0.271 |
| gc-sustained-v1/gc/struct-access | 0.250 | 0.396 | US | 4.479 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 0.291 | 0.438 | US | 2.208 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 0.334 | 0.520 | US | 2.250 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.291 | 0.500 | 2.667 | 4.646 | 5.396 | 0.292 |
| threads-sustained-v2/threads/shared-memory | 0.292 | 0.520 | 3.562 | 4.854 | 5.542 | 0.250 |
| threads-sustained-v2/threads/atomic-rmw | 0.271 | 0.520 | 3.063 | 4.709 | 5.396 | 0.271 |
| threads-sustained-v2/threads/notify-empty | 0.292 | 0.500 | 2.292 | 4.438 | 5.500 | 0.208 |
| threads-sustained-v2/threads/wait-mismatch | 0.292 | 0.541 | 3.208 | 4.979 | 5.208 | 0.292 |
| threads-sustained-v2/threads/wait-timeout | 0.333 | 0.500 | 2.708 | 4.625 | 5.500 | 0.292 |
| generic-calls-v1/core/generic-calls | 0.312 | 0.479 | 3.146 | 4.729 | 6.354 | 0.312 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 11.979 | 8.666 | 9.875 | 31.000 | 16.041 | 15.854 |
| original22-v1/core/integer-loop | 1,839.96 | 1,830.25 | 1,809.75 | 1,858.42 | 1,835.52 | 1,863.48 |
| original22-v1/core/function-calls | 963.06 | 884.40 | 1,526.63 | 1,096.88 | 931.27 | 923.58 |
| original22-v1/core/memory-load-store | 518.40 | 512.65 | 944.69 | 593.69 | 656.60 | 346.48 |
| original22-v1/memory/narrow-unaligned | 760.21 | 761.48 | 1,130.29 | 889.02 | 854.08 | 615.73 |
| original22-v1/memory/dependent-address | 2,193.23 | 2,175.08 | 3,512.38 | 2,576.81 | 2,478.54 | 1,688.06 |
| original22-v1/memory/float64-unaligned | 831.00 | 844.65 | 906.83 | 839.02 | 815.44 | 769.06 |
| original22-v1/runtime/simd-memory-full | 3,963.46 | 3,907.06 | 4,263.73 | 4,261.40 | 4,031.50 | 4,065.60 |
| original22-v1/runtime/simd-memory-widen | 1,229.02 | 1,222.27 | 1,198.19 | 1,254.21 | 1,204.58 | 1,189.04 |
| original22-v1/runtime/simd-memory-lanes | 5,770.77 | 5,925.92 | 6,793.69 | 6,877.81 | 6,824.00 | 5,907.23 |
| original22-v1/gc/i31-roundtrip | 28.980 | 24.604 | US | 54.084 | US | 30.291 |
| original22-v1/gc/struct-access | 34.063 | 35.666 | US | 59.145 | US | 143.75 |
| original22-v1/gc/struct-allocation | 413.33 | 630.81 | US | 157.98 | US | 573.50 |
| original22-v1/gc/struct-indirect-access | 32.521 | 27.875 | US | 35.104 | US | 357.40 |
| original22-v1/threads/private-memory | 18.041 | 13.979 | 22.187 | 51.771 | 30.459 | 37.520 |
| original22-v1/threads/shared-memory | 28.750 | 25.709 | 22.437 | 46.750 | 29.583 | 34.270 |
| original22-v1/threads/atomic-rmw | 36.438 | 30.812 | 31.271 | 58.041 | 63.250 | 41.791 |
| original22-v1/threads/notify-empty | 25.896 | 21.855 | 263.98 | 120.79 | 72.021 | 148.50 |
| original22-v1/threads/wait-mismatch | 27.083 | 22.355 | 303.92 | 144.19 | 43.770 | 57.542 |
| original22-v1/threads/wait-timeout | 27.396 | 23.395 | 3,139.35 | 537.27 | 2,211.83 | 742.42 |
| original22-v1/applications/sha256-clang | 1,199.75 | 1,172.54 | 1,541.60 | 1,170.33 | 1,110.31 | 996.98 |
| original22-v1/applications/sha256-rust | 1,221.96 | 1,219.37 | 1,658.38 | 1,096.62 | 1,152.94 | 1,049.65 |
| gc-sustained-v1/gc/i31-roundtrip | 1,594.54 | 1,591.65 | US | 2,084.17 | US | 1,465.94 |
| gc-sustained-v1/gc/struct-access | 1,239.81 | 1,256.96 | US | 1,261.12 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 31,618.50 | 24,698.67 | US | 3,949.73 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1,894.44 | 1,944.94 | US | 3,011.02 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 44.666 | 40.709 | 129.98 | 152.83 | 133.69 | 135.54 |
| threads-sustained-v2/threads/shared-memory | 160.90 | 155.17 | 131.15 | 158.81 | 139.88 | 139.96 |
| threads-sustained-v2/threads/atomic-rmw | 227.52 | 222.56 | 221.00 | 241.98 | 463.56 | 232.75 |
| threads-sustained-v2/threads/notify-empty | 115.69 | 114.50 | 2,484.17 | 832.79 | 580.90 | 1,063.88 |
| threads-sustained-v2/threads/wait-mismatch | 132.75 | 125.96 | 2,752.10 | 1,038.21 | 277.60 | 360.40 |
| threads-sustained-v2/threads/wait-timeout | 118.17 | 111.90 | 33,911.08 | 5,099.27 | 20,161.69 | 7,015.44 |
| generic-calls-v1/core/generic-calls | 6,091.08 | 5,975.17 | 12,471.44 | 12,179.15 | 12,173.73 | 6,108.42 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.500 | 5.562 | 10.938 | 82.604 | 33.562 | 401.00 |
| original22-v1/core/integer-loop | 12.438 | 11.604 | 11.396 | 87.791 | 42.667 | 428.73 |
| original22-v1/core/function-calls | 14.083 | 6.354 | 12.354 | 86.916 | 40.125 | 444.00 |
| original22-v1/core/memory-load-store | 21.000 | 15.479 | 11.334 | 116.71 | 72.730 | 515.25 |
| original22-v1/memory/narrow-unaligned | 15.854 | 14.208 | 8.874 | 117.60 | 72.375 | 502.35 |
| original22-v1/memory/dependent-address | 15.500 | 11.604 | 12.271 | 108.87 | 63.208 | 471.27 |
| original22-v1/memory/float64-unaligned | 14.416 | 8.874 | 6.917 | 93.291 | 66.791 | 500.02 |
| original22-v1/runtime/simd-memory-full | 18.021 | 14.709 | 10.396 | 130.85 | 71.396 | 520.31 |
| original22-v1/runtime/simd-memory-widen | 16.895 | 8.938 | 8.833 | 119.25 | 59.875 | 488.06 |
| original22-v1/runtime/simd-memory-lanes | 20.334 | 19.521 | 10.771 | 116.90 | 80.459 | 500.40 |
| original22-v1/gc/i31-roundtrip | 10.375 | 5.562 | US | 88.917 | US | 413.96 |
| original22-v1/gc/struct-access | 19.042 | 19.541 | US | 130.90 | US | 413.77 |
| original22-v1/gc/struct-allocation | 157.27 | 100.88 | US | 127.02 | US | 2,201.35 |
| original22-v1/gc/struct-indirect-access | 16.750 | 15.541 | US | 117.40 | US | 412.27 |
| original22-v1/threads/private-memory | 15.791 | 15.541 | 9.438 | 100.85 | 62.312 | 447.48 |
| original22-v1/threads/shared-memory | 19.625 | 10.166 | 9.209 | 88.083 | 60.541 | 467.96 |
| original22-v1/threads/atomic-rmw | 16.416 | 7.354 | 11.271 | 111.27 | 64.145 | 454.50 |
| original22-v1/threads/notify-empty | 11.562 | 6.959 | 5.188 | 87.249 | 63.729 | 441.73 |
| original22-v1/threads/wait-mismatch | 16.605 | 11.771 | 6.458 | 83.584 | 64.751 | 456.79 |
| original22-v1/threads/wait-timeout | 18.541 | 8.291 | 6.583 | 120.58 | 64.416 | 483.98 |
| original22-v1/applications/sha256-clang | 21.438 | 13.146 | 12.875 | 154.79 | 83.459 | 507.27 |
| original22-v1/applications/sha256-rust | 11.229 | 12.958 | 12.688 | 169.25 | 79.584 | 478.94 |
| gc-sustained-v1/gc/i31-roundtrip | 11.542 | 6.938 | US | 94.291 | US | 430.67 |
| gc-sustained-v1/gc/struct-access | 24.917 | 20.479 | US | 130.15 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 3,624.25 | 2,295.73 | US | 968.29 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 22.312 | 18.625 | US | 175.04 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 19.146 | 17.709 | 10.521 | 122.00 | 69.145 | 476.56 |
| threads-sustained-v2/threads/shared-memory | 15.729 | 9.438 | 7.292 | 107.40 | 57.834 | 491.38 |
| threads-sustained-v2/threads/atomic-rmw | 19.438 | 9.146 | 11.875 | 125.77 | 65.229 | 496.62 |
| threads-sustained-v2/threads/notify-empty | 19.000 | 14.354 | 12.625 | 95.666 | 62.500 | 474.65 |
| threads-sustained-v2/threads/wait-mismatch | 17.791 | 13.812 | 8.542 | 101.60 | 73.229 | 468.69 |
| threads-sustained-v2/threads/wait-timeout | 18.312 | 13.334 | 5.146 | 115.77 | 66.791 | 487.85 |
| generic-calls-v1/core/generic-calls | 17.271 | 9.938 | 13.000 | 149.27 | 52.229 | 465.33 |

</details>

<details>
<summary>Interpreter: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5,046.77 | 7,153.60 | 5,766.38 | NM | NM | 19,327.02 |
| original22-v1/core/integer-loop | 5,678.67 | 7,341.15 | 5,582.92 | NM | NM | 20,655.27 |
| original22-v1/core/function-calls | 5,431.44 | 7,786.69 | 5,893.35 | NM | NM | 19,837.88 |
| original22-v1/core/memory-load-store | 6,538.79 | 8,448.92 | 5,818.15 | NM | NM | 20,775.46 |
| original22-v1/memory/narrow-unaligned | 5,691.00 | 9,151.90 | 5,148.73 | NM | NM | 20,891.75 |
| original22-v1/memory/dependent-address | 5,747.56 | 7,909.92 | 5,520.92 | NM | NM | 20,471.23 |
| original22-v1/memory/float64-unaligned | 5,595.17 | 7,138.40 | 5,485.17 | NM | NM | 18,711.10 |
| original22-v1/runtime/simd-memory-full | 6,782.04 | 8,731.46 | 5,979.90 | NM | NM | 18,907.50 |
| original22-v1/runtime/simd-memory-widen | 4,842.56 | 7,569.75 | 5,192.21 | NM | NM | 18,952.77 |
| original22-v1/runtime/simd-memory-lanes | 5,255.25 | 7,420.23 | 6,223.92 | NM | NM | 19,987.38 |
| original22-v1/gc/i31-roundtrip | 4,694.71 | 6,511.79 | US | NM | NM | 18,564.42 |
| original22-v1/gc/struct-access | 4,980.33 | 6,889.56 | US | NM | NM | 18,297.04 |
| original22-v1/gc/struct-allocation | 5,328.33 | 6,988.77 | US | NM | NM | 18,487.54 |
| original22-v1/gc/struct-indirect-access | 5,504.33 | 7,861.15 | US | NM | NM | 19,001.90 |
| original22-v1/threads/private-memory | 5,405.98 | 6,944.23 | 4,952.60 | NM | NM | 18,272.21 |
| original22-v1/threads/shared-memory | 5,592.48 | 6,995.17 | 5,035.98 | NM | NM | 18,420.94 |
| original22-v1/threads/atomic-rmw | 5,231.42 | 7,576.81 | 4,997.73 | NM | NM | 19,170.81 |
| original22-v1/threads/notify-empty | 5,958.33 | 8,230.21 | 5,413.79 | NM | NM | 19,451.56 |
| original22-v1/threads/wait-mismatch | 5,235.77 | 6,817.04 | 5,422.10 | NM | NM | 19,400.48 |
| original22-v1/threads/wait-timeout | 5,700.27 | 8,422.12 | 5,656.40 | NM | NM | 22,644.06 |
| original22-v1/applications/sha256-clang | 5,682.62 | 7,239.25 | 5,888.50 | NM | NM | 20,302.10 |
| original22-v1/applications/sha256-rust | 6,785.08 | 8,400.44 | 6,239.37 | NM | NM | 19,649.00 |
| gc-sustained-v1/gc/i31-roundtrip | 4,876.48 | 8,105.52 | US | NM | NM | 19,105.44 |
| gc-sustained-v1/gc/struct-access | 5,210.88 | 7,498.96 | US | NM | NM | 20,605.92 |
| gc-sustained-v1/gc/struct-allocation | 5,656.94 | 7,634.58 | US | NM | NM | 19,441.83 |
| gc-sustained-v1/gc/struct-indirect-access | 5,263.58 | 7,879.48 | US | NM | NM | 21,105.15 |
| threads-sustained-v2/threads/private-memory | 5,197.31 | 7,824.31 | 4,910.15 | NM | NM | 19,793.29 |
| threads-sustained-v2/threads/shared-memory | 6,667.92 | 7,463.71 | 5,873.27 | NM | NM | 20,053.50 |
| threads-sustained-v2/threads/atomic-rmw | 4,920.48 | 6,995.54 | 4,870.96 | NM | NM | 18,636.94 |
| threads-sustained-v2/threads/notify-empty | 5,495.06 | 7,851.25 | 5,132.96 | NM | NM | 19,930.23 |
| threads-sustained-v2/threads/wait-mismatch | 5,393.27 | 6,945.48 | 5,202.48 | NM | NM | 17,727.77 |
| threads-sustained-v2/threads/wait-timeout | 4,957.90 | 7,029.12 | 5,452.73 | NM | NM | 18,766.77 |
| generic-calls-v1/core/generic-calls | 5,786.81 | 9,139.37 | 5,482.29 | NM | NM | 20,786.08 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 50.959 | 42.938 | 46.979 | NM | NM | 48.750 |
| original22-v1/core/integer-loop | 47.855 | 52.645 | 99.667 | NM | NM | 52.354 |
| original22-v1/core/function-calls | 53.291 | 61.125 | 86.604 | NM | NM | 92.834 |
| original22-v1/core/memory-load-store | 52.270 | 51.709 | 49.770 | NM | NM | 53.958 |
| original22-v1/memory/narrow-unaligned | 73.479 | 89.021 | 48.146 | NM | NM | 76.500 |
| original22-v1/memory/dependent-address | 43.480 | 49.647 | 52.312 | NM | NM | 52.230 |
| original22-v1/memory/float64-unaligned | 49.562 | 52.562 | 52.375 | NM | NM | 58.458 |
| original22-v1/runtime/simd-memory-full | 50.520 | 58.688 | 51.500 | NM | NM | 75.188 |
| original22-v1/runtime/simd-memory-widen | 45.604 | 50.021 | 50.125 | NM | NM | 54.270 |
| original22-v1/runtime/simd-memory-lanes | 45.250 | 58.812 | 50.105 | NM | NM | 70.271 |
| original22-v1/gc/i31-roundtrip | 58.270 | 48.541 | US | NM | NM | 48.521 |
| original22-v1/gc/struct-access | 44.895 | 51.604 | US | NM | NM | 75.937 |
| original22-v1/gc/struct-allocation | 42.541 | 42.980 | US | NM | NM | 50.791 |
| original22-v1/gc/struct-indirect-access | 39.979 | 46.834 | US | NM | NM | 46.167 |
| original22-v1/threads/private-memory | 39.188 | 48.812 | 43.063 | NM | NM | 48.647 |
| original22-v1/threads/shared-memory | 43.749 | 52.750 | 43.105 | NM | NM | 47.229 |
| original22-v1/threads/atomic-rmw | 45.709 | 46.313 | 43.688 | NM | NM | 44.730 |
| original22-v1/threads/notify-empty | 52.166 | 54.521 | 66.396 | NM | NM | 59.125 |
| original22-v1/threads/wait-mismatch | 44.166 | 48.104 | 58.979 | NM | NM | 51.480 |
| original22-v1/threads/wait-timeout | 46.979 | 74.938 | 59.104 | NM | NM | 70.854 |
| original22-v1/applications/sha256-clang | 44.209 | 61.334 | 54.666 | NM | NM | 52.896 |
| original22-v1/applications/sha256-rust | 50.980 | 50.042 | 50.604 | NM | NM | 56.354 |
| gc-sustained-v1/gc/i31-roundtrip | 41.021 | 50.604 | US | NM | NM | 50.770 |
| gc-sustained-v1/gc/struct-access | 46.334 | 54.000 | US | NM | NM | 44.375 |
| gc-sustained-v1/gc/struct-allocation | 49.145 | 48.166 | US | NM | NM | 54.188 |
| gc-sustained-v1/gc/struct-indirect-access | 44.541 | 49.750 | US | NM | NM | 57.020 |
| threads-sustained-v2/threads/private-memory | 39.875 | 43.125 | 38.542 | NM | NM | 49.396 |
| threads-sustained-v2/threads/shared-memory | 88.042 | 56.020 | 50.209 | NM | NM | 91.167 |
| threads-sustained-v2/threads/atomic-rmw | 49.104 | 57.396 | 50.375 | NM | NM | 61.770 |
| threads-sustained-v2/threads/notify-empty | 57.125 | 52.125 | 41.062 | NM | NM | 57.999 |
| threads-sustained-v2/threads/wait-mismatch | 49.854 | 45.730 | 40.188 | NM | NM | 54.917 |
| threads-sustained-v2/threads/wait-timeout | 45.730 | 43.375 | 48.458 | NM | NM | 60.416 |
| generic-calls-v1/core/generic-calls | 81.542 | 56.146 | 52.750 | NM | NM | 72.396 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.896 | 4.937 | 2.958 | NM | NM | 38.500 |
| original22-v1/core/integer-loop | 5.750 | 5.938 | 3.438 | NM | NM | 47.395 |
| original22-v1/core/function-calls | 5.875 | 5.604 | 3.104 | NM | NM | 38.812 |
| original22-v1/core/memory-load-store | 6.458 | 6.250 | 3.250 | NM | NM | 36.188 |
| original22-v1/memory/narrow-unaligned | 6.375 | 6.479 | 2.917 | NM | NM | 37.812 |
| original22-v1/memory/dependent-address | 5.125 | 5.146 | 3.249 | NM | NM | 35.708 |
| original22-v1/memory/float64-unaligned | 5.146 | 6.104 | 3.250 | NM | NM | 35.770 |
| original22-v1/runtime/simd-memory-full | 5.396 | 6.708 | 3.062 | NM | NM | 46.584 |
| original22-v1/runtime/simd-memory-widen | 5.250 | 5.792 | 2.917 | NM | NM | 34.584 |
| original22-v1/runtime/simd-memory-lanes | 5.146 | 5.771 | 3.104 | NM | NM | 40.354 |
| original22-v1/gc/i31-roundtrip | 6.292 | 5.458 | US | NM | NM | 36.729 |
| original22-v1/gc/struct-access | 5.396 | 5.271 | US | NM | NM | 36.084 |
| original22-v1/gc/struct-allocation | 5.167 | 5.458 | US | NM | NM | 38.459 |
| original22-v1/gc/struct-indirect-access | 5.667 | 5.333 | US | NM | NM | 36.458 |
| original22-v1/threads/private-memory | 4.875 | 5.771 | 2.812 | NM | NM | 34.459 |
| original22-v1/threads/shared-memory | 5.583 | 6.167 | 2.854 | NM | NM | 38.125 |
| original22-v1/threads/atomic-rmw | 5.854 | 5.562 | 2.646 | NM | NM | 39.250 |
| original22-v1/threads/notify-empty | 5.625 | 5.938 | 4.375 | NM | NM | 38.354 |
| original22-v1/threads/wait-mismatch | 5.229 | 5.750 | 3.188 | NM | NM | 37.021 |
| original22-v1/threads/wait-timeout | 6.167 | 6.646 | 3.792 | NM | NM | 51.083 |
| original22-v1/applications/sha256-clang | 6.146 | 5.167 | 2.875 | NM | NM | 35.291 |
| original22-v1/applications/sha256-rust | 5.292 | 5.708 | 2.854 | NM | NM | 39.354 |
| gc-sustained-v1/gc/i31-roundtrip | 5.750 | 5.854 | US | NM | NM | 37.438 |
| gc-sustained-v1/gc/struct-access | 5.750 | 5.500 | US | NM | NM | 35.750 |
| gc-sustained-v1/gc/struct-allocation | 5.396 | 5.750 | US | NM | NM | 40.104 |
| gc-sustained-v1/gc/struct-indirect-access | 5.521 | 7.042 | US | NM | NM | 39.666 |
| threads-sustained-v2/threads/private-memory | 5.479 | 5.729 | 2.792 | NM | NM | 46.646 |
| threads-sustained-v2/threads/shared-memory | 7.708 | 5.896 | 3.354 | NM | NM | 33.958 |
| threads-sustained-v2/threads/atomic-rmw | 6.208 | 5.750 | 3.229 | NM | NM | 39.021 |
| threads-sustained-v2/threads/notify-empty | 6.104 | 6.188 | 2.938 | NM | NM | 38.563 |
| threads-sustained-v2/threads/wait-mismatch | 5.750 | 5.229 | 3.146 | NM | NM | 36.916 |
| threads-sustained-v2/threads/wait-timeout | 5.208 | 5.271 | 3.187 | NM | NM | 38.145 |
| generic-calls-v1/core/generic-calls | 5.813 | 5.896 | 3.292 | NM | NM | 116.69 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 62.229 | 52.333 | 41.125 | NM | NM | 41.084 |
| original22-v1/core/integer-loop | 76.063 | 65.021 | 53.480 | NM | NM | 42.355 |
| original22-v1/core/function-calls | 110.42 | 75.291 | 51.583 | NM | NM | 43.479 |
| original22-v1/core/memory-load-store | 93.459 | 87.875 | 64.791 | NM | NM | 45.812 |
| original22-v1/memory/narrow-unaligned | 100.79 | 96.166 | 54.230 | NM | NM | 46.375 |
| original22-v1/memory/dependent-address | 91.271 | 91.084 | 64.166 | NM | NM | 47.584 |
| original22-v1/memory/float64-unaligned | 91.772 | 90.751 | 57.541 | NM | NM | 44.874 |
| original22-v1/runtime/simd-memory-full | 89.521 | 80.250 | 60.000 | NM | NM | 50.375 |
| original22-v1/runtime/simd-memory-widen | 85.541 | 74.584 | 58.062 | NM | NM | 45.291 |
| original22-v1/runtime/simd-memory-lanes | 86.125 | 80.709 | 57.438 | NM | NM | 50.688 |
| original22-v1/gc/i31-roundtrip | 76.583 | 57.812 | US | NM | NM | 40.812 |
| original22-v1/gc/struct-access | 73.375 | 65.646 | US | NM | NM | 50.022 |
| original22-v1/gc/struct-allocation | 77.896 | 64.354 | US | NM | NM | 55.667 |
| original22-v1/gc/struct-indirect-access | 89.916 | 84.583 | US | NM | NM | 53.917 |
| original22-v1/threads/private-memory | 73.542 | 60.980 | 47.541 | NM | NM | 42.729 |
| original22-v1/threads/shared-memory | 74.291 | 61.021 | 46.021 | NM | NM | 43.833 |
| original22-v1/threads/atomic-rmw | 71.063 | 58.062 | 48.896 | NM | NM | 44.729 |
| original22-v1/threads/notify-empty | 72.541 | 63.896 | 66.562 | NM | NM | 42.729 |
| original22-v1/threads/wait-mismatch | 73.750 | 60.833 | 48.855 | NM | NM | 42.292 |
| original22-v1/threads/wait-timeout | 74.438 | 65.645 | 53.062 | NM | NM | 51.459 |
| original22-v1/applications/sha256-clang | 309.42 | 400.60 | 91.541 | NM | NM | 90.084 |
| original22-v1/applications/sha256-rust | 307.50 | 407.54 | 97.188 | NM | NM | 95.750 |
| gc-sustained-v1/gc/i31-roundtrip | 70.105 | 60.521 | US | NM | NM | 39.605 |
| gc-sustained-v1/gc/struct-access | 76.417 | 68.375 | US | NM | NM | 50.229 |
| gc-sustained-v1/gc/struct-allocation | 78.125 | 64.938 | US | NM | NM | 55.958 |
| gc-sustained-v1/gc/struct-indirect-access | 91.500 | 87.979 | US | NM | NM | 56.146 |
| threads-sustained-v2/threads/private-memory | 71.104 | 59.188 | 45.000 | NM | NM | 42.833 |
| threads-sustained-v2/threads/shared-memory | 76.062 | 61.875 | 49.354 | NM | NM | 43.501 |
| threads-sustained-v2/threads/atomic-rmw | 72.395 | 60.834 | 59.542 | NM | NM | 42.854 |
| threads-sustained-v2/threads/notify-empty | 73.396 | 76.041 | 50.646 | NM | NM | 44.520 |
| threads-sustained-v2/threads/wait-mismatch | 72.750 | 58.917 | 48.208 | NM | NM | 47.605 |
| threads-sustained-v2/threads/wait-timeout | 73.270 | 59.416 | 48.021 | NM | NM | 43.980 |
| generic-calls-v1/core/generic-calls | 78.479 | 83.334 | 48.895 | NM | NM | 53.063 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 8.688 | 9.188 | 18.188 | NM | NM | 32.459 |
| original22-v1/core/integer-loop | 9.167 | 9.354 | 19.645 | NM | NM | 38.541 |
| original22-v1/core/function-calls | 9.896 | 9.729 | 17.229 | NM | NM | 30.813 |
| original22-v1/core/memory-load-store | 13.750 | 16.250 | 20.125 | NM | NM | 53.875 |
| original22-v1/memory/narrow-unaligned | 13.833 | 16.895 | 19.270 | NM | NM | 55.791 |
| original22-v1/memory/dependent-address | 13.959 | 17.146 | 20.146 | NM | NM | 54.334 |
| original22-v1/memory/float64-unaligned | 13.542 | 17.000 | 18.521 | NM | NM | 50.667 |
| original22-v1/runtime/simd-memory-full | 16.500 | 16.291 | 20.354 | NM | NM | 87.084 |
| original22-v1/runtime/simd-memory-widen | 15.479 | 16.021 | 19.396 | NM | NM | 62.146 |
| original22-v1/runtime/simd-memory-lanes | 15.500 | 16.084 | 19.084 | NM | NM | 67.146 |
| original22-v1/gc/i31-roundtrip | 9.396 | 9.063 | US | NM | NM | 29.104 |
| original22-v1/gc/struct-access | 9.312 | 10.896 | US | NM | NM | 30.333 |
| original22-v1/gc/struct-allocation | 10.709 | 10.792 | US | NM | NM | 38.416 |
| original22-v1/gc/struct-indirect-access | 16.167 | 18.375 | US | NM | NM | 38.250 |
| original22-v1/threads/private-memory | 12.396 | 15.688 | 18.500 | NM | NM | 54.750 |
| original22-v1/threads/shared-memory | 17.063 | 16.438 | 17.938 | NM | NM | 53.084 |
| original22-v1/threads/atomic-rmw | 16.750 | 16.688 | 19.270 | NM | NM | 55.458 |
| original22-v1/threads/notify-empty | 17.625 | 16.688 | 21.209 | NM | NM | 50.562 |
| original22-v1/threads/wait-mismatch | 18.166 | 15.709 | 18.688 | NM | NM | 52.855 |
| original22-v1/threads/wait-timeout | 17.500 | 16.938 | 19.167 | NM | NM | 58.334 |
| original22-v1/applications/sha256-clang | 17.083 | 23.167 | 21.562 | NM | NM | 63.395 |
| original22-v1/applications/sha256-rust | 18.000 | 92.687 | 35.104 | NM | NM | 66.645 |
| gc-sustained-v1/gc/i31-roundtrip | 8.979 | 9.396 | US | NM | NM | 33.083 |
| gc-sustained-v1/gc/struct-access | 9.438 | 10.646 | US | NM | NM | 31.605 |
| gc-sustained-v1/gc/struct-allocation | 10.625 | 11.458 | US | NM | NM | 39.500 |
| gc-sustained-v1/gc/struct-indirect-access | 17.166 | 18.646 | US | NM | NM | 40.355 |
| threads-sustained-v2/threads/private-memory | 12.312 | 15.104 | 17.646 | NM | NM | 63.313 |
| threads-sustained-v2/threads/shared-memory | 18.167 | 16.041 | 18.458 | NM | NM | 51.771 |
| threads-sustained-v2/threads/atomic-rmw | 20.125 | 16.917 | 19.270 | NM | NM | 54.208 |
| threads-sustained-v2/threads/notify-empty | 18.209 | 21.416 | 19.605 | NM | NM | 55.000 |
| threads-sustained-v2/threads/wait-mismatch | 16.770 | 15.000 | 18.021 | NM | NM | 56.291 |
| threads-sustained-v2/threads/wait-timeout | 17.291 | 15.854 | 18.625 | NM | NM | 55.666 |
| generic-calls-v1/core/generic-calls | 9.562 | 11.521 | 18.541 | NM | NM | 35.729 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.250 | 0.459 | 0.229 | NM | NM | 0.230 |
| original22-v1/core/integer-loop | 0.271 | 0.438 | 0.291 | NM | NM | 0.208 |
| original22-v1/core/function-calls | 0.354 | 0.479 | 0.271 | NM | NM | 0.292 |
| original22-v1/core/memory-load-store | 0.313 | 0.438 | 0.312 | NM | NM | 0.250 |
| original22-v1/memory/narrow-unaligned | 0.375 | 0.520 | 0.291 | NM | NM | 0.271 |
| original22-v1/memory/dependent-address | 0.292 | 0.458 | 0.333 | NM | NM | 0.208 |
| original22-v1/memory/float64-unaligned | 0.333 | 0.459 | 0.271 | NM | NM | 0.230 |
| original22-v1/runtime/simd-memory-full | 0.333 | 0.479 | 0.209 | NM | NM | 0.271 |
| original22-v1/runtime/simd-memory-widen | 0.312 | 0.459 | 0.250 | NM | NM | 0.250 |
| original22-v1/runtime/simd-memory-lanes | 0.333 | 0.479 | 0.250 | NM | NM | 0.250 |
| original22-v1/gc/i31-roundtrip | 0.250 | 0.458 | US | NM | NM | 0.250 |
| original22-v1/gc/struct-access | 0.230 | 0.437 | US | NM | NM | 0.230 |
| original22-v1/gc/struct-allocation | 0.250 | 0.416 | US | NM | NM | 0.250 |
| original22-v1/gc/struct-indirect-access | 0.271 | 0.458 | US | NM | NM | 0.209 |
| original22-v1/threads/private-memory | 0.271 | 0.479 | 0.208 | NM | NM | 0.209 |
| original22-v1/threads/shared-memory | 0.292 | 0.625 | 0.271 | NM | NM | 0.250 |
| original22-v1/threads/atomic-rmw | 0.291 | 0.521 | 0.250 | NM | NM | 0.250 |
| original22-v1/threads/notify-empty | 0.334 | 0.500 | 0.354 | NM | NM | 0.250 |
| original22-v1/threads/wait-mismatch | 0.354 | 0.583 | 0.271 | NM | NM | 0.250 |
| original22-v1/threads/wait-timeout | 0.291 | 0.542 | 0.292 | NM | NM | 0.312 |
| original22-v1/applications/sha256-clang | 0.291 | 0.500 | 0.271 | NM | NM | 0.229 |
| original22-v1/applications/sha256-rust | 0.271 | 0.459 | 0.333 | NM | NM | 0.250 |
| gc-sustained-v1/gc/i31-roundtrip | 0.312 | 0.500 | US | NM | NM | 0.250 |
| gc-sustained-v1/gc/struct-access | 0.250 | 0.416 | US | NM | NM | 0.250 |
| gc-sustained-v1/gc/struct-allocation | 0.250 | 0.417 | US | NM | NM | 0.209 |
| gc-sustained-v1/gc/struct-indirect-access | 0.250 | 0.479 | US | NM | NM | 0.208 |
| threads-sustained-v2/threads/private-memory | 0.271 | 0.479 | 0.230 | NM | NM | 0.167 |
| threads-sustained-v2/threads/shared-memory | 0.292 | 0.541 | 0.250 | NM | NM | 0.291 |
| threads-sustained-v2/threads/atomic-rmw | 0.334 | 0.500 | 0.271 | NM | NM | 0.250 |
| threads-sustained-v2/threads/notify-empty | 0.292 | 0.542 | 0.250 | NM | NM | 0.250 |
| threads-sustained-v2/threads/wait-mismatch | 0.271 | 0.542 | 0.209 | NM | NM | 0.250 |
| threads-sustained-v2/threads/wait-timeout | 0.271 | 0.500 | 0.271 | NM | NM | 0.291 |
| generic-calls-v1/core/generic-calls | 0.292 | 0.459 | 0.334 | NM | NM | 0.208 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.875 | 7.375 | 4.042 | NM | NM | 8.271 |
| original22-v1/core/integer-loop | 36,501.08 | 31,644.40 | 98,437.90 | NM | NM | 58,442.75 |
| original22-v1/core/function-calls | 64,473.46 | 56,135.44 | 102,796.13 | NM | NM | 78,009.65 |
| original22-v1/core/memory-load-store | 78,322.33 | 69,109.12 | 203,014.94 | NM | NM | 112,400.81 |
| original22-v1/memory/narrow-unaligned | 105,265.06 | 95,119.02 | 218,356.96 | NM | NM | 140,047.02 |
| original22-v1/memory/dependent-address | 48,207.35 | 41,865.42 | 116,002.04 | NM | NM | 63,495.56 |
| original22-v1/memory/float64-unaligned | 80,462.96 | 69,824.42 | 159,703.73 | NM | NM | 99,342.40 |
| original22-v1/runtime/simd-memory-full | 89,806.73 | 81,918.33 | 182,647.75 | NM | NM | 95,358.33 |
| original22-v1/runtime/simd-memory-widen | 73,694.58 | 65,077.48 | 166,107.52 | NM | NM | 71,991.90 |
| original22-v1/runtime/simd-memory-lanes | 113,442.79 | 105,232.25 | 228,308.42 | NM | NM | 108,460.04 |
| original22-v1/gc/i31-roundtrip | 422.15 | 377.33 | US | NM | NM | 526.12 |
| original22-v1/gc/struct-access | 465.44 | 476.79 | US | NM | NM | 633.81 |
| original22-v1/gc/struct-allocation | 1,041.23 | 1,148.73 | US | NM | NM | 1,386.00 |
| original22-v1/gc/struct-indirect-access | 843.31 | 834.25 | US | NM | NM | 1,109.50 |
| original22-v1/threads/private-memory | 465.48 | 404.35 | 737.27 | NM | NM | 466.56 |
| original22-v1/threads/shared-memory | 860.75 | 405.21 | 724.75 | NM | NM | 449.94 |
| original22-v1/threads/atomic-rmw | 605.56 | 397.00 | 540.15 | NM | NM | 415.48 |
| original22-v1/threads/notify-empty | 377.81 | 342.94 | 510.62 | NM | NM | 482.42 |
| original22-v1/threads/wait-mismatch | 696.40 | 442.17 | 803.40 | NM | NM | 520.00 |
| original22-v1/threads/wait-timeout | 638.21 | 422.67 | 3,721.35 | NM | NM | 1,251.00 |
| original22-v1/applications/sha256-clang | 110,186.92 | 93,602.58 | 243,397.00 | NM | NM | 156,107.67 |
| original22-v1/applications/sha256-rust | 103,881.88 | 92,780.67 | 246,434.50 | NM | NM | 146,005.00 |
| gc-sustained-v1/gc/i31-roundtrip | 39,262.25 | 37,850.81 | US | NM | NM | 52,357.50 |
| gc-sustained-v1/gc/struct-access | 48,308.17 | 46,670.46 | US | NM | NM | 64,559.77 |
| gc-sustained-v1/gc/struct-allocation | 60,235.38 | 41,823.27 | US | NM | NM | 37,554.15 |
| gc-sustained-v1/gc/struct-indirect-access | 86,293.12 | 85,031.52 | US | NM | NM | 112,456.56 |
| threads-sustained-v2/threads/private-memory | 4,456.00 | 3,902.08 | 7,082.10 | NM | NM | 4,549.17 |
| threads-sustained-v2/threads/shared-memory | 9,074.44 | 4,079.10 | 7,452.42 | NM | NM | 4,533.44 |
| threads-sustained-v2/threads/atomic-rmw | 6,033.58 | 3,888.40 | 5,363.33 | NM | NM | 3,664.73 |
| threads-sustained-v2/threads/notify-empty | 3,560.98 | 3,295.83 | 5,153.96 | NM | NM | 4,724.04 |
| threads-sustained-v2/threads/wait-mismatch | 6,171.65 | 3,986.92 | 7,678.65 | NM | NM | 4,958.54 |
| threads-sustained-v2/threads/wait-timeout | 6,154.10 | 3,985.04 | 38,897.69 | NM | NM | 11,542.38 |
| generic-calls-v1/core/generic-calls | 687,692.06 | 599,234.02 | 1,364,343.00 | NM | NM | 731,330.52 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.187 | 1.979 | 7.229 | NM | NM | 8.979 |
| original22-v1/core/integer-loop | 6.312 | 2.146 | 11.021 | NM | NM | 13.459 |
| original22-v1/core/function-calls | 6.750 | 2.125 | 4.521 | NM | NM | 10.584 |
| original22-v1/core/memory-load-store | 2.479 | 2.271 | 12.146 | NM | NM | 58.916 |
| original22-v1/memory/narrow-unaligned | 2.583 | 2.271 | 8.792 | NM | NM | 57.751 |
| original22-v1/memory/dependent-address | 2.688 | 2.417 | 9.666 | NM | NM | 54.751 |
| original22-v1/memory/float64-unaligned | 3.417 | 2.249 | 5.438 | NM | NM | 55.521 |
| original22-v1/runtime/simd-memory-full | 2.667 | 2.438 | 10.937 | NM | NM | 54.938 |
| original22-v1/runtime/simd-memory-widen | 2.729 | 2.312 | 9.459 | NM | NM | 54.083 |
| original22-v1/runtime/simd-memory-lanes | 4.584 | 2.292 | 8.666 | NM | NM | 56.312 |
| original22-v1/gc/i31-roundtrip | 5.292 | 1.833 | US | NM | NM | 6.146 |
| original22-v1/gc/struct-access | 5.521 | 2.271 | US | NM | NM | 9.334 |
| original22-v1/gc/struct-allocation | 107.38 | 61.375 | US | NM | NM | 1,740.19 |
| original22-v1/gc/struct-indirect-access | 2.312 | 2.146 | US | NM | NM | 7.541 |
| original22-v1/threads/private-memory | 2.333 | 2.250 | 9.792 | NM | NM | 47.125 |
| original22-v1/threads/shared-memory | 14.708 | 5.396 | 8.626 | NM | NM | 47.584 |
| original22-v1/threads/atomic-rmw | 16.500 | 8.125 | 11.021 | NM | NM | 47.355 |
| original22-v1/threads/notify-empty | 12.438 | 7.063 | 5.938 | NM | NM | 45.562 |
| original22-v1/threads/wait-mismatch | 16.916 | 5.104 | 7.146 | NM | NM | 48.459 |
| original22-v1/threads/wait-timeout | 13.312 | 9.292 | 6.312 | NM | NM | 49.605 |
| original22-v1/applications/sha256-clang | 2.542 | 2.312 | 6.541 | NM | NM | 62.959 |
| original22-v1/applications/sha256-rust | 2.479 | 2.312 | 7.417 | NM | NM | 62.562 |
| gc-sustained-v1/gc/i31-roundtrip | 6.938 | 2.063 | US | NM | NM | 11.312 |
| gc-sustained-v1/gc/struct-access | 9.354 | 2.250 | US | NM | NM | 11.084 |
| gc-sustained-v1/gc/struct-allocation | 3,125.50 | 1,881.23 | US | NM | NM | 69,602.40 |
| gc-sustained-v1/gc/struct-indirect-access | 2.292 | 2.271 | US | NM | NM | 10.084 |
| threads-sustained-v2/threads/private-memory | 2.667 | 2.479 | 12.354 | NM | NM | 54.021 |
| threads-sustained-v2/threads/shared-memory | 15.188 | 8.771 | 7.688 | NM | NM | 53.729 |
| threads-sustained-v2/threads/atomic-rmw | 19.395 | 9.542 | 9.396 | NM | NM | 52.188 |
| threads-sustained-v2/threads/notify-empty | 17.812 | 10.708 | 10.729 | NM | NM | 54.438 |
| threads-sustained-v2/threads/wait-mismatch | 18.145 | 16.229 | 7.729 | NM | NM | 51.834 |
| threads-sustained-v2/threads/wait-timeout | 21.583 | 7.875 | 4.188 | NM | NM | 53.354 |
| generic-calls-v1/core/generic-calls | 8.521 | 2.250 | 5.208 | NM | NM | 11.459 |

</details>

### WasmEdge AOT preparation detail

| Workload | Compile + validation + artifact I/O, us | Native image load, us | Post-load validation, us |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 22,572.92 | 251,231.65 | 3.146 |
| original22-v1/core/integer-loop | 26,396.10 | 254,763.35 | 6.188 |
| original22-v1/core/function-calls | 25,521.17 | 250,730.96 | 5.458 |
| original22-v1/core/memory-load-store | 28,123.77 | 252,816.42 | 7.333 |
| original22-v1/memory/narrow-unaligned | 24,662.85 | 253,858.52 | 9.291 |
| original22-v1/memory/dependent-address | 26,503.29 | 253,454.31 | 9.479 |
| original22-v1/memory/float64-unaligned | 27,346.44 | 249,946.56 | 7.583 |
| original22-v1/runtime/simd-memory-full | 25,555.96 | 250,344.35 | 7.646 |
| original22-v1/runtime/simd-memory-widen | 23,678.48 | 250,614.08 | 8.104 |
| original22-v1/runtime/simd-memory-lanes | 27,185.15 | 252,966.06 | 8.125 |
| original22-v1/gc/i31-roundtrip | 22,879.19 | 250,548.75 | 5.729 |
| original22-v1/gc/struct-access | 23,583.60 | 252,679.31 | 3.896 |
| original22-v1/gc/struct-allocation | 22,718.40 | 249,763.92 | 7.208 |
| original22-v1/gc/struct-indirect-access | 23,624.31 | 250,512.54 | 8.667 |
| original22-v1/threads/private-memory | 22,185.08 | 253,024.46 | 6.708 |
| original22-v1/threads/shared-memory | 25,653.79 | 252,459.10 | 6.708 |
| original22-v1/threads/atomic-rmw | 21,191.12 | 251,252.38 | 6.146 |
| original22-v1/threads/notify-empty | 28,688.85 | 250,096.56 | 6.458 |
| original22-v1/threads/wait-mismatch | 24,832.81 | 250,742.08 | 5.458 |
| original22-v1/threads/wait-timeout | 27,254.75 | 251,711.19 | 5.542 |
| original22-v1/applications/sha256-clang | 29,964.21 | 256,201.52 | 12.437 |
| original22-v1/applications/sha256-rust | 30,366.88 | 252,483.94 | 11.312 |
| gc-sustained-v1/gc/i31-roundtrip | 23,585.12 | 254,955.08 | 5.708 |
| gc-sustained-v1/gc/struct-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 21,975.38 | 254,335.15 | 6.271 |
| threads-sustained-v2/threads/shared-memory | 22,689.17 | 252,572.27 | 5.792 |
| threads-sustained-v2/threads/atomic-rmw | 23,943.73 | 254,500.94 | 5.458 |
| threads-sustained-v2/threads/notify-empty | 25,460.44 | 250,032.48 | 7.417 |
| threads-sustained-v2/threads/wait-mismatch | 20,546.40 | 249,847.88 | 7.312 |
| threads-sustained-v2/threads/wait-timeout | 21,893.25 | 252,246.21 | 6.375 |
| generic-calls-v1/core/generic-calls | 26,982.25 | 254,516.38 | 6.812 |

### P3 phase observations

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B / cleanup | 63.791 | 589.67 | NM | 247.50 | NM | NM |
| 16 B / compileValidate | 64,530.27 | 104,961.56 | NM | 54,463.10 | NM | NM |
| 16 B / engineSetup | 0.479 | 0.209 | NM | 156.67 | NM | NM |
| 16 B / executorSetup | N/A | N/A | NM | 47.146 | NM | NM |
| 16 B / exportLookup | 0.729 | 1.042 | NM | 9.627 | NM | NM |
| 16 B / firstCall | 9,898.42 | 24,342.48 | NM | 10,406.90 | NM | NM |
| 16 B / instantiate | 503.94 | 787.02 | NM | 203.83 | NM | NM |
| 16 B / load | 72.167 | 70.291 | NM | 64.480 | NM | NM |
| 16 B / processStartup | 6,368.17 | 13,786.46 | NM | 6,761.13 | NM | NM |
| 16 B / providerSetup | 51.728 | 231.58 | NM | 237.83 | NM | NM |
| 16 B / requestPreparation | 0.771 | 0.895 | NM | N/A | NM | NM |
| 65536 B / cleanup | 83.895 | 698.44 | NM | 317.29 | NM | NM |
| 65536 B / compileValidate | 64,937.60 | 103,768.31 | NM | 54,542.83 | NM | NM |
| 65536 B / engineSetup | 0.459 | 0.188 | NM | 155.58 | NM | NM |
| 65536 B / executorSetup | N/A | N/A | NM | 45.230 | NM | NM |
| 65536 B / exportLookup | 0.729 | 1.230 | NM | 9.416 | NM | NM |
| 65536 B / firstCall | 14,706.50 | 25,477.94 | NM | 10,348.25 | NM | NM |
| 65536 B / instantiate | 569.89 | 831.87 | NM | 218.04 | NM | NM |
| 65536 B / load | 63.104 | 71.791 | NM | 61.208 | NM | NM |
| 65536 B / processStartup | 6,657.77 | 13,541.15 | NM | 6,802.85 | NM | NM |
| 65536 B / providerSetup | 50.541 | 211.40 | NM | 247.23 | NM | NM |
| 65536 B / requestPreparation | 11.458 | 17.311 | NM | N/A | NM | NM |

## Native controls (different contracts)

Unchecked C kernels omit runtime safety/lifetime contracts; native Go implements HTTP replies without Component/WASI. These are overhead controls with different contracts. The table reports their measured timing directly. The general-call C control runs twice as many iterations of one recurrence; The optimized Wasm combines two recurrence updates into one affine operation.

| Workload | Unchecked native C, ns/iteration |
| --- | ---: |
| original22-v1/core/integer-loop | 1.847 |
| original22-v1/core/function-calls | 0.931 |
| original22-v1/core/memory-load-store | 0.325 |
| original22-v1/memory/narrow-unaligned | 0.544 |
| original22-v1/memory/dependent-address | 1.644 |
| original22-v1/memory/float64-unaligned | 0.796 |
| original22-v1/runtime/simd-memory-full | 4.207 |
| original22-v1/runtime/simd-memory-widen | 1.200 |
| original22-v1/runtime/simd-memory-lanes | 5.418 |
| original22-v1/applications/sha256-clang | 196.65 |
| original22-v1/applications/sha256-rust | 195.95 |
| generic-calls-v1/core/generic-calls | 2.373 |

| Payload | Native Go CPU, us/request | Native Go requests/s |
| --- | ---: | ---: |
| 16 B | 3.563 | 1,104,020.95 |
| 65536 B | 98.540 | 25,159.32 |

## WasmEdge sustained AOT failures

Full-size WasmEdge0.17.1 O3 AOT preflight freshly reproduces SIGSEGV in sustained struct access, allocation and indirect access. Small original cases pass. Only those exact three cells remain failures in the timing matrix. The original investigation and the observed cumulative stack growth are retained in the [PR #200 report](performance-pr200-complete-comparison-2026-11-17.md). All table timings come from the current event.

## Mechanism evidence

An independent instrumented replay counts eleven direct scalar leaves per P3 request. General host handoffs fall from 34 to 23 at 16 B and from 46 to 35 at 64 KiB. Native machine-code entries stay at 38 and 50 per request. The corrected frame variant removes seven allocations and approximately 768 bytes per request relative to the scalar-leaf-only build. The instrumented replays and source screens in the protocol archive are separate from the primary timing event.

## Resolved adverse primary intervals

No source comparison has a resolved adverse interval in this event.

## Independent diagnostics

The warm Go allocation profile uses 8,192 fully validated requests per payload.
The largest recorded 16 B allocation sites include reference allocation
(25.31 MiB), endpoint allocation (22 MiB), codecs (20 MiB), new host call
transactions (19.88 MiB), and component and Core callers (19.25 and 18.50 MiB).
This allocation profile uses a sample rate of one and forced collection at its
boundaries. It contributes no throughput or CPU observations.

The first transaction-frame version added ten allocations and around 320 bytes
per request relative to the scalar-leaf-only build. Go escape output attributes
the extra copies to the returned local and the separate storage target. Writing
the clone into its final storage and reusing frame-owned internal results removes
that regression. In the corrected screen, the 64 KiB instruction ratio versus
scalar leaves alone is 0.9961 [0.9925, 0.9994] with both A/A checks resolved.
The 16 B instruction interval crosses one. CPU and throughput controls in this
comparison remain unresolved.

A narrower native entry ABI experiment replaced two typed off-heap pointer
arguments with numeric addresses. Go assembly confirms elimination of two cgo
pointer checks per entry. It leaves the cgo transition itself in place. The
16 B instruction interval improves, while the 64 KiB instruction interval crosses
one and its throughput point is lower. The numeric ABI remains on its diagnostic
branch and is absent from the measured implementation.

Local Linux/amd64 Docker Go tests omit one stable-memory test after reproducing
the Rosetta address-space fatal error on the clean source base with the same
container. The exact test remains enabled in hosted CI. Local Go race and LLGo
preflight cover the affected packages. Hosted CI covers the full matrix.

### Native boundary and HTTP control

Wasmtime 48.0.1 emits typed trampolines for waitable operations that
call built-in runtime functions through the component context. Cranelift
passes these scalar arguments in registers, and the host libcall returns
through the guest activation. The measured wasmvm leaf path keeps the frame
and drops the public callback transaction. It still returns through the Go
native-entry driver for each such operation. The trace counts establish
the crossing frequency. A matched boundary-only benchmark is required to
attribute CPU time to this design difference.

- [Wasmtime 48.0.1 component trampoline generator](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/cranelift/src/compiler/component.rs#L250-L289)
- [Wasmtime 48.0.1 component runtime builtins](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasmtime/src/runtime/vm/component/libcalls.rs#L73-L101)

The native Go HTTP control materializes the incoming request body and outgoing
response body with two growing-buffer reads. The Wasmtime provider reuses
an immutable refcounted snapshot on request construction. The scoped
wasmvm runs use the immutable snapshot and buffered response API. The
native Go control is a separate HTTP path. Its 64 KiB CPU time cannot
be used as a lower bound on VM overhead.

- [Native Go control and benchmark response read](https://github.com/cpunion/wasmvm/blob/57e67995b4122ca8d73540eee525c804c12c6b55/benchmarks/p3io/main.go#L125-L174)
- [Wasmtime HTTP provider](https://github.com/cpunion/wasmvm/blob/57e67995b4122ca8d73540eee525c804c12c6b55/benchmarks/p3io/rust/src/lib.rs#L211-L250)


### Core source replay

The primary Core event produced ten apparent source slowdowns. All ten had
unresolved same-binary controls. A separate 240-process replay checked those
ten intervals with CPU, retired-instruction and cycle counters on the exact
primary binaries. Only LLGo interpretation of narrow unaligned memory retained
a resolved higher CPU time.

| Narrow unaligned memory, LLGo interpreter | Head/base CPU time Δ | Head/base retired instructions Δ | Head/base cycles per instruction Δ |
| --- | ---: | ---: | ---: |
| Independent replay, original binaries | +3.09% [1.81%, 4.49%] | +0.002% [−0.015%, 0.026%] | +3.30% [2.29%, 4.32%] |
| Matched interpreter entry address | +3.03% [2.42%, 3.79%] | +0.000% [−0.002%, 0.001%] | +3.24% [2.66%, 3.76%] |

The default links emit 2,052 instructions and the same opcode sequence in
`runSlotsLoop`. The changed encodings occur at linked data references and
call sites. The second replay reordered six interpreter and memory functions
for both sources. Both `runSlotsLoop` entries landed at `0x100000a40`.
The integer-loop signal flipped direction with that intervention. The
unaligned-memory signal persisted, and some callee and data locations still
differ. The exact cause of the extra cycles remains open. The contribution
remains under review pending resolution. Both diagnostic replays and the
disassembly are archived separately from the primary event.

## Reproducible evidence

- [Machine-readable results, all controls and source contrasts](data/p3-scalar-leaves-2026-11-21.json).
- [All 5,508 primary process records](data/p3-scalar-leaves-2026-11-21-raw.jsonl.gz), SHA-256 `75b2463f21a36375025deee49c7397f5093635f7f02ae61243f2115dc662a94b`.
- [Frozen plans, commands, hashes, analysis/export scripts, semantic checks and failure evidence](data/p3-scalar-leaves-2026-11-21-protocol.tar.gz), SHA-256 `f8f201e631d37f44ae88bfdfd6b27691bea1c73322dc93c6a78dc0971ca1b3e6`.
- Per-request vectors unused by these metrics are projected to count/extrema/hash; no latency-tail claim. Public load records omit unrelated desktop command arguments. The separately archived screening observations are excluded from primary ratios.
