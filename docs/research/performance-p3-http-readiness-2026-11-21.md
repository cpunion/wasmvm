# P3 HTTP readiness and complete runtime comparison, 2026-11-21

HTTP transmission results complete directly from the publisher. Body completion and trailers drive component tasks while awaiting readiness. External reads and writes retain the I/O worker limit. Publication preserves cancellation, failed trailers, and affine resource ownership.

- P3 16 B, Go. CPU 88.336 → 82.593 µs/request and throughput 38,549.33 → 40,848.55 requests/s. Paired CPU comparison against source base 0.943 [0.931, 0.956] · Δ -5.71% · speed +6.06% [+4.55%, +7.38%] · unresolved. Against Wasmtime 1.521 [1.458, 1.577] · Δ +52.13% · speed -34.27% [-36.60%, -31.41%] · unresolved.
- P3 16 B, LLGo. CPU 210.17 → 145.37 µs/request and throughput 14,060.55 → 19,313.95 requests/s. Paired CPU comparison against source base 0.684 [0.665, 0.698] · Δ -31.63% · speed +46.25% [+43.29%, +50.28%] · unresolved. Against Wasmtime 2.569 [2.417, 2.700] · Δ +156.86% · speed -61.07% [-62.96%, -58.63%] · unresolved.
- P3 65,536 B, Go. CPU 128.73 → 123.44 µs/request and throughput 26,337.30 → 27,533.06 requests/s. Paired CPU comparison against source base 0.957 [0.944, 0.971] · Δ -4.30% · speed +4.49% [+2.99%, +5.91%] · unresolved. Against Wasmtime 1.884 [1.778, 1.970] · Δ +88.37% · speed -46.91% [-49.25%, -43.75%] · unresolved.
- P3 65,536 B, LLGo. CPU 252.02 → 191.18 µs/request and throughput 12,148.81 → 16,472.55 requests/s. Paired CPU comparison against source base 0.756 [0.734, 0.779] · Δ -24.39% · speed +32.25% [+28.37%, +36.19%] · unresolved. Against Wasmtime 2.900 [2.786, 3.002] · Δ +190.04% · speed -65.52% [-66.69%, -64.11%] · unresolved.

Against the fastest measured competitor median per workload, native Go and LLGo lead on 13 and 15 of 33 cases. Interpreter Go and LLGo lead on 24 and 30 cases. These are descriptive median counts and the paired timing controls remain unresolved.

All 33 core workloads and two complete P3 payloads are measured on the same pinned toolchains. Tables give preparation and steady execution independently. The performance acceptance gate remains open.

## Percentage differences

Cells show `absolute observation (signed speed Δ%)`. Positive speed change means faster. Native execution and P3 use **Wasmtime** as a fixed reference. Interpreter execution uses **wazero** and uses **WasmEdge** where the pinned wazero profile rejects GC. Using a fixed reference keeps the percentages comparable across revisions. The fastest observed runtime varies by workload.

For time, speed Δ = `(reference / candidate - 1) * 100%`. For throughput, speed Δ = `(candidate / reference - 1) * 100%`. Thus 100 ns to 50 ns is **−50% time**, **+100% speed**. P3 CPU percentages mean work per CPU time, separate from wall-clock throughput. Instruction/allocation/byte counters show quantity Δ rather than speed. Source comparisons retain paired ratio intervals and report quantity Δ plus speed Δ where applicable. Reciprocal interval endpoints are reversed.

Competitive cell percentages are descriptive quotients of the displayed medians. Paired block estimates below can differ. `†` on these cells means either the candidate or the named reference has unresolved A/A controls and qualified conclusions require resolved controls. Phase boundaries with different contracts have separate absolute observations. The percentages apply to the frozen event and measured revision.

## Revisions, profiles and protocol

- The engine base is `d45d609d` (merged PR #205), measured head `b5d9a828`. Reused binaries retain their producer revisions. The build records verify production-source equivalence for each reused binary.
- Go **1.27.3** and LLGo **3c6b33ec** compile both source revisions. LLGo uses **`-O3 -lto=full`** with LLVM **22.1.8** and default production GC.
- macOS arm64 runs the complete changed component and WASI Preview 3 suites on both pinned Go and LLGo. Go statement coverage is 91.9% in component and 91.3% in Preview 3. Scoped race checks pass.

- The peers are wazero **1.12.0** in compiler and interpreter modes, Wasmtime **48.0.1** with Cranelift speed, Wasmer **7.3.0** with Cranelift, and WasmEdge **0.17.1** in interpreter and O3 AOT modes.
- Local Apple M4 Max, macOS arm64, 128 GiB RAM. The record includes external desktop load. The timing windows are reserved for benchmark work. The timing scope is macOS arm64.
- Core uses six fixed fresh-process blocks, each with symmetric same-binary A/A, two warmups and seven interleaved measured calls. **5,364 successful process records** after excluding exactly three failing sustained WasmEdge AOT cells. All 33 workloads remain represented. Allocation workloads use means that include pauses. Other workloads use medians.
- P3 uses **72 processes per payload**, six roles with A/A, six fixed blocks, four workers, sixteen clients and twenty interleaved 512-request windows plus a separate flush. Verify status, headers, full body, trailers, CRC and joined guest completion. CPU/instruction counters include the driver and checks. Mach ticks use the recorded timebase.
- Ratios use paired block geometric means and 20,000 fixed-seed bootstrap resamples with marginal 95% intervals. Both relevant A/A intervals must fit `[0.975, 1.025]`. Each interval applies to its own comparison. The protocol fixes the sample size in advance and measures each phase directly.

**Legend** `†` means the role or its percentage reference has an unresolved same-binary timing control. Values are observations. A qualified comparison additionally needs both controls and its ratio interval. `US` means this pinned profile rejected the feature in semantic preflight. `NM` identifies an unmeasured benchmark adapter or profile. This status applies only to this measurement. `N/A` means the metric does not apply. `FAIL` is an execution failure in the named profile.

## Native compiled execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

The reference runtime is **Wasmtime**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 8,187.00 (-66.40% speed) † | 5,583.00 (-50.73% speed) † | 1,271.00 (+116.40% speed) † | 2,750.50 (ref) † | 5,062.50 (-45.67% speed) † | 6,187.50 (-55.55% speed) † |
| original22-v1/core/integer-loop | 2.027 (-0.26% speed) † | 2.086 (-3.05% speed) † | 2.272 (-11.01% speed) † | 2.022 (ref) † | 2.087 (-3.11% speed) † | 2.014 (+0.42% speed) † |
| original22-v1/core/function-calls | 1.028 (+17.77% speed) † | 1.044 (+15.95% speed) † | 1.850 (-34.57% speed) † | 1.210 (ref) † | 1.074 (+12.71% speed) † | 1.034 (+17.04% speed) † |
| original22-v1/core/memory-load-store | 0.599 (+7.74% speed) † | 0.606 (+6.44% speed) † | 1.220 (-47.12% speed) † | 0.645 (ref) † | 0.717 (-10.01% speed) † | 0.375 (+72.18% speed) † |
| original22-v1/memory/narrow-unaligned | 0.902 (+4.54% speed) † | 0.939 (+0.45% speed) † | 1.411 (-33.17% speed) † | 0.943 (ref) † | 1.008 (-6.46% speed) † | 0.685 (+37.62% speed) † |
| original22-v1/memory/dependent-address | 2.438 (+11.50% speed) † | 2.438 (+11.49% speed) † | 4.161 (-34.67% speed) † | 2.718 (ref) † | 2.815 (-3.45% speed) † | 1.890 (+43.83% speed) † |
| original22-v1/memory/float64-unaligned | 0.962 (-1.04% speed) † | 1.001 (-4.91% speed) † | 1.250 (-23.87% speed) † | 0.952 (ref) | 0.935 (+1.76% speed) † | 0.906 (+5.05% speed) † |
| original22-v1/runtime/simd-memory-full | 4.288 (+8.19% speed) † | 4.249 (+9.20% speed) † | 4.817 (-3.67% speed) † | 4.640 (ref) † | 4.424 (+4.87% speed) † | 4.745 (-2.23% speed) † |
| original22-v1/runtime/simd-memory-widen | 1.352 (-2.14% speed) † | 1.340 (-1.26% speed) † | 1.545 (-14.33% speed) † | 1.323 (ref) † | 1.353 (-2.21% speed) † | 1.362 (-2.85% speed) † |
| original22-v1/runtime/simd-memory-lanes | 6.409 (+16.57% speed) † | 6.360 (+17.47% speed) † | 7.970 (-6.26% speed) † | 7.471 (ref) † | 7.389 (+1.10% speed) † | 6.410 (+16.56% speed) † |
| original22-v1/gc/i31-roundtrip | 3.200 (-14.52% speed) † | 2.548 (+7.36% speed) † | US | 2.735 (ref) † | US | 2.494 (+9.69% speed) † |
| original22-v1/gc/struct-access | 3.344 (-44.48% speed) † | 2.506 (-25.93% speed) † | US | 1.856 (ref) † | US | 12.704 (-85.39% speed) † |
| original22-v1/gc/struct-allocation | 42.321 (+158.80% speed) † | 40.934 (+167.57% speed) † | US | 109.53 (ref) † | US | 56.174 (+94.98% speed) † |
| original22-v1/gc/struct-indirect-access | 4.369 (-12.21% speed) † | 3.392 (+13.08% speed) † | US | 3.835 (ref) † | US | 35.213 (-89.11% speed) † |
| original22-v1/threads/private-memory | 2.038 (-8.69% speed) † | 1.254 (+48.34% speed) † | 2.223 (-16.31% speed) † | 1.860 (ref) † | 2.190 (-15.03% speed) † | 2.288 (-18.67% speed) † |
| original22-v1/threads/shared-memory | 3.417 (-37.50% speed) † | 2.506 (-14.80% speed) † | 2.160 (-1.16% speed) † | 2.135 (ref) † | 2.252 (-5.18% speed) † | 2.258 (-5.44% speed) † |
| original22-v1/threads/atomic-rmw | 3.733 (-24.00% speed) † | 3.383 (-16.13% speed) † | 3.331 (-14.82% speed) † | 2.837 (ref) † | 5.954 (-52.35% speed) † | 3.373 (-15.87% speed) † |
| original22-v1/threads/notify-empty | 3.027 (+199.18% speed) † | 2.054 (+340.87% speed) † | 28.019 (-67.68% speed) † | 9.056 (ref) † | 6.944 (+30.42% speed) † | 12.098 (-25.14% speed) † |
| original22-v1/threads/wait-mismatch | 3.379 (+258.14% speed) † | 2.688 (+350.31% speed) † | 35.910 (-66.30% speed) † | 12.102 (ref) † | 4.158 (+191.03% speed) † | 5.115 (+136.62% speed) † |
| original22-v1/threads/wait-timeout | 3.521 (+1479.95% speed) † | 2.256 (+2365.46% speed) † | 451.28 (-87.67% speed) † | 55.627 (ref) † | 232.93 (-76.12% speed) † | 81.448 (-31.70% speed) † |
| original22-v1/applications/sha256-clang | 276.87 (-3.67% speed) † | 284.08 (-6.12% speed) † | 398.12 (-33.01% speed) † | 266.70 (ref) † | 273.33 (-2.43% speed) † | 245.99 (+8.42% speed) † |
| original22-v1/applications/sha256-rust | 289.85 (-9.71% speed) † | 312.05 (-16.13% speed) † | 455.12 (-42.50% speed) † | 261.71 (ref) † | 270.41 (-3.22% speed) † | 264.98 (-1.24% speed) † |
| gc-sustained-v1/gc/i31-roundtrip | 1.823 (+31.82% speed) † | 1.807 (+32.99% speed) † | US | 2.403 (ref) † | US | 1.668 (+44.06% speed) † |
| gc-sustained-v1/gc/struct-access | 1.393 (-3.38% speed) | 1.348 (-0.15% speed) | US | 1.346 (ref) | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 42.858 (+58.97% speed) † | 46.362 (+46.95% speed) † | US | 68.130 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 2.133 (+67.39% speed) † | 2.192 (+62.87% speed) † | US | 3.571 (ref) † | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.575 (+137.75% speed) † | 0.436 (+213.82% speed) † | 1.654 (-17.36% speed) † | 1.367 (ref) † | 1.427 (-4.19% speed) † | 1.407 (-2.87% speed) † |
| threads-sustained-v2/threads/shared-memory | 1.812 (-29.73% speed) † | 1.918 (-33.61% speed) † | 1.522 (-16.32% speed) † | 1.273 (ref) † | 1.407 (-9.51% speed) † | 1.389 (-8.32% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 2.515 (-5.06% speed) † | 2.377 (+0.46% speed) † | 2.665 (-10.40% speed) † | 2.388 (ref) † | 5.038 (-52.61% speed) † | 2.384 (+0.14% speed) † |
| threads-sustained-v2/threads/notify-empty | 1.538 (+484.83% speed) † | 1.256 (+616.19% speed) † | 28.998 (-68.98% speed) † | 8.994 (ref) † | 6.213 (+44.78% speed) † | 12.403 (-27.48% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 1.654 (+625.37% speed) † | 1.467 (+717.33% speed) † | 33.668 (-64.37% speed) † | 11.994 (ref) † | 3.270 (+266.85% speed) † | 4.255 (+181.86% speed) † |
| threads-sustained-v2/threads/wait-timeout | 1.429 (+3960.50% speed) † | 1.260 (+4502.76% speed) † | 537.72 (-89.21% speed) † | 58.014 (ref) † | 227.73 (-74.53% speed) † | 81.558 (-28.87% speed) † |
| generic-calls-v1/core/generic-calls | 1.329 (+101.42% speed) | 1.331 (+101.19% speed) † | 2.797 (-4.27% speed) † | 2.678 (ref) | 2.683 (-0.20% speed) | 1.323 (+102.37% speed) |

## Interpreter execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

The reference runtime is **wazero**. GC rows use **WasmEdge**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,104.00 (-47.71% speed) † | 4,270.50 (-49.75% speed) † | 2,146.00 (ref) † | NM | NM | 5,333.00 (-59.76% speed) † |
| original22-v1/core/integer-loop | 40.850 (+270.05% speed) † | 34.740 (+335.14% speed) † | 151.17 (ref) † | NM | NM | 71.141 (+112.49% speed) † |
| original22-v1/core/function-calls | 63.485 (+165.45% speed) † | 60.286 (+179.54% speed) † | 168.52 (ref) † | NM | NM | 83.089 (+102.82% speed) † |
| original22-v1/core/memory-load-store | 91.084 (+279.20% speed) † | 77.624 (+344.96% speed) † | 345.39 (ref) † | NM | NM | 142.56 (+142.29% speed) † |
| original22-v1/memory/narrow-unaligned | 115.95 (+149.48% speed) † | 99.988 (+189.29% speed) † | 289.26 (ref) † | NM | NM | 157.39 (+83.79% speed) † |
| original22-v1/memory/dependent-address | 54.693 (+138.68% speed) † | 46.407 (+181.29% speed) † | 130.54 (ref) † | NM | NM | 73.649 (+77.25% speed) † |
| original22-v1/memory/float64-unaligned | 90.539 (+166.31% speed) † | 79.133 (+204.69% speed) † | 241.11 (ref) † | NM | NM | 117.36 (+105.44% speed) † |
| original22-v1/runtime/simd-memory-full | 109.39 (+119.84% speed) † | 89.565 (+168.50% speed) † | 240.48 (ref) † | NM | NM | 109.03 (+120.56% speed) † |
| original22-v1/runtime/simd-memory-widen | 82.122 (+169.95% speed) † | 69.652 (+218.28% speed) † | 221.69 (ref) † | NM | NM | 82.795 (+167.75% speed) † |
| original22-v1/runtime/simd-memory-lanes | 140.49 (+236.81% speed) † | 149.81 (+215.86% speed) † | 473.18 (ref) † | NM | NM | 182.82 (+158.83% speed) † |
| original22-v1/gc/i31-roundtrip (vs WasmEdge) | 48.892 (+20.91% speed) † | 43.223 (+36.77% speed) † | US | NM | NM | 59.117 (ref) † |
| original22-v1/gc/struct-access (vs WasmEdge) | 56.906 (+25.96% speed) † | 54.140 (+32.40% speed) † | US | NM | NM | 71.679 (ref) † |
| original22-v1/gc/struct-allocation (vs WasmEdge) | 140.73 (+3.88% speed) † | 145.64 (+0.38% speed) † | US | NM | NM | 146.19 (ref) † |
| original22-v1/gc/struct-indirect-access (vs WasmEdge) | 103.42 (+30.13% speed) † | 98.129 (+37.14% speed) † | US | NM | NM | 134.58 (ref) † |
| original22-v1/threads/private-memory | 52.004 (+64.68% speed) † | 43.869 (+95.22% speed) † | 85.642 (ref) † | NM | NM | 50.856 (+68.40% speed) † |
| original22-v1/threads/shared-memory | 101.60 (-19.54% speed) † | 43.490 (+87.97% speed) † | 81.746 (ref) † | NM | NM | 49.913 (+63.78% speed) † |
| original22-v1/threads/atomic-rmw | 70.700 (-14.27% speed) † | 44.669 (+35.68% speed) † | 60.608 (ref) † | NM | NM | 46.037 (+31.65% speed) † |
| original22-v1/threads/notify-empty | 42.358 (+28.28% speed) † | 35.379 (+53.58% speed) † | 54.335 (ref) † | NM | NM | 51.140 (+6.25% speed) † |
| original22-v1/threads/wait-mismatch | 80.419 (+14.75% speed) † | 49.348 (+87.01% speed) † | 92.283 (ref) † | NM | NM | 58.071 (+58.92% speed) † |
| original22-v1/threads/wait-timeout | 76.346 (+574.67% speed) † | 46.912 (+997.97% speed) † | 515.09 (ref) † | NM | NM | 125.71 (+309.74% speed) † |
| original22-v1/applications/sha256-clang | 30,319.17 (+202.10% speed) † | 22,384.71 (+309.18% speed) † | 91,593.52 (ref) † | NM | NM | 51,875.99 (+76.56% speed) † |
| original22-v1/applications/sha256-rust | 42,460.45 (+262.00% speed) † | 33,094.40 (+364.45% speed) † | 153,707.60 (ref) † | NM | NM | 62,978.59 (+144.06% speed) † |
| gc-sustained-v1/gc/i31-roundtrip (vs WasmEdge) | 47.500 (+33.14% speed) † | 41.730 (+51.55% speed) † | US | NM | NM | 63.243 (ref) † |
| gc-sustained-v1/gc/struct-access (vs WasmEdge) | 54.576 (+39.31% speed) † | 52.122 (+45.87% speed) † | US | NM | NM | 76.030 (ref) † |
| gc-sustained-v1/gc/struct-allocation (vs WasmEdge) | 133.20 (+6.34% speed) † | 145.44 (-2.61% speed) † | US | NM | NM | 141.64 (ref) † |
| gc-sustained-v1/gc/struct-indirect-access (vs WasmEdge) | 100.38 (+29.58% speed) † | 96.286 (+35.09% speed) † | US | NM | NM | 130.08 (ref) † |
| threads-sustained-v2/threads/private-memory | 54.351 (+70.65% speed) † | 50.229 (+84.66% speed) † | 92.751 (ref) † | NM | NM | 55.661 (+66.63% speed) † |
| threads-sustained-v2/threads/shared-memory | 105.31 (-19.34% speed) † | 45.582 (+86.36% speed) † | 84.945 (ref) † | NM | NM | 51.409 (+65.23% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 75.437 (-13.02% speed) † | 43.769 (+49.90% speed) † | 65.612 (ref) † | NM | NM | 43.002 (+52.58% speed) † |
| threads-sustained-v2/threads/notify-empty | 41.404 (+38.55% speed) † | 36.353 (+57.80% speed) † | 57.365 (ref) † | NM | NM | 53.491 (+7.24% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 76.170 (+20.99% speed) † | 52.484 (+75.59% speed) † | 92.157 (ref) † | NM | NM | 60.326 (+52.76% speed) † |
| threads-sustained-v2/threads/wait-timeout | 75.530 (+673.23% speed) † | 47.042 (+1141.47% speed) † | 584.02 (ref) † | NM | NM | 132.73 (+340.00% speed) † |
| generic-calls-v1/core/generic-calls | 162.46 (+85.45% speed) † | 134.43 (+124.12% speed) † | 301.28 (ref) † | NM | NM | 172.59 (+74.56% speed) † |

### Warm speed against the fastest observed competitor

Each row selects the fastest measured competitor median for that backend and workload. The observed percentages remain descriptive when the selected reference or candidate has unresolved controls.

#### Native

| Workload | Fastest observed competitor | wasmvm Go speed Δ | wasmvm LLGo speed Δ |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-compiler | -84.48% † | -77.23% † |
| original22-v1/core/integer-loop | wasmedge-compiler | -0.68% † | -3.45% † |
| original22-v1/core/function-calls | wasmedge-compiler | +0.63% † | -0.93% † |
| original22-v1/core/memory-load-store | wasmedge-compiler | -37.42% † | -38.18% † |
| original22-v1/memory/narrow-unaligned | wasmedge-compiler | -24.03% † | -27.00% † |
| original22-v1/memory/dependent-address | wasmedge-compiler | -22.48% † | -22.49% † |
| original22-v1/memory/float64-unaligned | wasmedge-compiler | -5.80% † | -9.49% † |
| original22-v1/runtime/simd-memory-full | wasmer-compiler | +3.17% † | +4.13% † |
| original22-v1/runtime/simd-memory-widen | wasmtime-compiler | -2.14% † | -1.26% † |
| original22-v1/runtime/simd-memory-lanes | wasmedge-compiler | +0.01% † | +0.79% † |
| original22-v1/gc/i31-roundtrip | wasmedge-compiler | -22.07% † | -2.13% † |
| original22-v1/gc/struct-access | wasmtime-compiler | -44.48% † | -25.93% † |
| original22-v1/gc/struct-allocation | wasmedge-compiler | +32.73% † | +37.23% † |
| original22-v1/gc/struct-indirect-access | wasmtime-compiler | -12.21% † | +13.08% † |
| original22-v1/threads/private-memory | wasmtime-compiler | -8.69% † | +48.34% † |
| original22-v1/threads/shared-memory | wasmtime-compiler | -37.50% † | -14.80% † |
| original22-v1/threads/atomic-rmw | wasmtime-compiler | -24.00% † | -16.13% † |
| original22-v1/threads/notify-empty | wasmer-compiler | +129.39% † | +238.04% † |
| original22-v1/threads/wait-mismatch | wasmer-compiler | +23.06% † | +54.73% † |
| original22-v1/threads/wait-timeout | wasmtime-compiler | +1479.95% † | +2365.46% † |
| original22-v1/applications/sha256-clang | wasmedge-compiler | -11.15% † | -13.41% † |
| original22-v1/applications/sha256-rust | wasmtime-compiler | -9.71% † | -16.13% † |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-compiler | -8.49% † | -7.68% † |
| gc-sustained-v1/gc/struct-access | wasmtime-compiler | -3.38% | -0.15% |
| gc-sustained-v1/gc/struct-allocation | wasmtime-compiler | +58.97% † | +46.95% † |
| gc-sustained-v1/gc/struct-indirect-access | wasmtime-compiler | +67.39% † | +62.87% † |
| threads-sustained-v2/threads/private-memory | wasmtime-compiler | +137.75% † | +213.82% † |
| threads-sustained-v2/threads/shared-memory | wasmtime-compiler | -29.73% † | -33.61% † |
| threads-sustained-v2/threads/atomic-rmw | wasmedge-compiler | -5.19% † | +0.32% † |
| threads-sustained-v2/threads/notify-empty | wasmer-compiler | +303.95% † | +394.69% † |
| threads-sustained-v2/threads/wait-mismatch | wasmer-compiler | +97.73% † | +122.80% † |
| threads-sustained-v2/threads/wait-timeout | wasmtime-compiler | +3960.50% † | +4502.76% † |
| generic-calls-v1/core/generic-calls | wasmedge-compiler | -0.47% | -0.59% † |

#### Interpreter

| Workload | Fastest observed competitor | wasmvm Go speed Δ | wasmvm LLGo speed Δ |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-interpreter | -47.71% † | -49.75% † |
| original22-v1/core/integer-loop | wasmedge-interpreter | +74.15% † | +104.78% † |
| original22-v1/core/function-calls | wasmedge-interpreter | +30.88% † | +37.82% † |
| original22-v1/core/memory-load-store | wasmedge-interpreter | +56.51% † | +83.65% † |
| original22-v1/memory/narrow-unaligned | wasmedge-interpreter | +35.74% † | +57.41% † |
| original22-v1/memory/dependent-address | wasmedge-interpreter | +34.66% † | +58.70% † |
| original22-v1/memory/float64-unaligned | wasmedge-interpreter | +29.63% † | +48.31% † |
| original22-v1/runtime/simd-memory-full | wasmedge-interpreter | -0.33% † | +21.73% † |
| original22-v1/runtime/simd-memory-widen | wasmedge-interpreter | +0.82% † | +18.87% † |
| original22-v1/runtime/simd-memory-lanes | wasmedge-interpreter | +30.13% † | +22.04% † |
| original22-v1/gc/i31-roundtrip | wasmedge-interpreter | +20.91% † | +36.77% † |
| original22-v1/gc/struct-access | wasmedge-interpreter | +25.96% † | +32.40% † |
| original22-v1/gc/struct-allocation | wasmedge-interpreter | +3.88% † | +0.38% † |
| original22-v1/gc/struct-indirect-access | wasmedge-interpreter | +30.13% † | +37.14% † |
| original22-v1/threads/private-memory | wasmedge-interpreter | -2.21% † | +15.93% † |
| original22-v1/threads/shared-memory | wasmedge-interpreter | -50.87% † | +14.77% † |
| original22-v1/threads/atomic-rmw | wasmedge-interpreter | -34.88% † | +3.06% † |
| original22-v1/threads/notify-empty | wasmedge-interpreter | +20.73% † | +44.55% † |
| original22-v1/threads/wait-mismatch | wasmedge-interpreter | -27.79% † | +17.68% † |
| original22-v1/threads/wait-timeout | wasmedge-interpreter | +64.66% † | +167.97% † |
| original22-v1/applications/sha256-clang | wasmedge-interpreter | +71.10% † | +131.75% † |
| original22-v1/applications/sha256-rust | wasmedge-interpreter | +48.32% † | +90.30% † |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-interpreter | +33.14% † | +51.55% † |
| gc-sustained-v1/gc/struct-access | wasmedge-interpreter | +39.31% † | +45.87% † |
| gc-sustained-v1/gc/struct-allocation | wasmedge-interpreter | +6.34% † | -2.61% † |
| gc-sustained-v1/gc/struct-indirect-access | wasmedge-interpreter | +29.58% † | +35.09% † |
| threads-sustained-v2/threads/private-memory | wasmedge-interpreter | +2.41% † | +10.82% † |
| threads-sustained-v2/threads/shared-memory | wasmedge-interpreter | -51.18% † | +12.78% † |
| threads-sustained-v2/threads/atomic-rmw | wasmedge-interpreter | -43.00% † | -1.75% † |
| threads-sustained-v2/threads/notify-empty | wasmedge-interpreter | +29.19% † | +47.14% † |
| threads-sustained-v2/threads/wait-mismatch | wasmedge-interpreter | -20.80% † | +14.94% † |
| threads-sustained-v2/threads/wait-timeout | wasmedge-interpreter | +75.73% † | +182.15% † |
| generic-calls-v1/core/generic-calls | wasmedge-interpreter | +6.23% † | +28.39% † |

## Complete P3 HTTP requests

P3 competitor cells marked `NM` have no measured adapter for the pinned component/WASI workload. Wasmtime uses the same validated workload. Managed allocation counts exclude foreign Rust/C/JIT allocations and cover only the managed share of allocation.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU, us/request | 82.593 (-33.01% speed) † | 145.37 (-61.94% speed) † | NM | 55.326 (ref) † | NM | NM |
| 16 B instructions/request | 674,761.51 (+188.63% Δ) † | 920,567.58 (+293.77% Δ) | NM | 233,783.01 (ref) | NM | NM |
| 16 B requests/s | 40,848.55 (+2.09% speed) † | 19,313.95 (-51.73% speed) † | NM | 40,013.87 (ref) † | NM | NM |
| 65536 B CPU, us/request | 123.44 (-47.80% speed) † | 191.18 (-66.29% speed) † | NM | 64.437 (ref) † | NM | NM |
| 65536 B instructions/request | 960,459.64 (+198.15% Δ) † | 1,254,670.58 (+289.49% Δ) | NM | 322,135.22 (ref) | NM | NM |
| 65536 B requests/s | 27,533.06 (-20.80% speed) † | 16,472.55 (-52.62% speed) † | NM | 34,763.33 (ref) † | NM | NM |

### P3 source contrasts

| Metric | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| 16 B cpuUs | 0.943 [0.931, 0.956] · Δ -5.71% · speed +6.06% [+4.55%, +7.38%] · unresolved | 0.684 [0.665, 0.698] · Δ -31.63% · speed +46.25% [+43.29%, +50.28%] · unresolved |
| 16 B instructions | 0.960 [0.945, 0.979] · Δ -3.97% · unresolved | 0.724 [0.717, 0.731] · Δ -27.62% · pass |
| 16 B throughput | 1.041 [0.996, 1.082] · Δ +4.07% · speed +4.07% [-0.35%, +8.20%] · unresolved | 1.446 [1.388, 1.518] · Δ +44.62% · speed +44.62% [+38.75%, +51.78%] · unresolved |
| 16 B allocsPerRequest | 0.920 [0.920, 0.920] · Δ -7.96% · pass | 0.921 [0.921, 0.921] · Δ -7.86% · pass |
| 16 B bytesPerRequest | 0.908 [0.908, 0.908] · Δ -9.22% · pass | 0.914 [0.914, 0.914] · Δ -8.64% · pass |
| 65536 B cpuUs | 0.957 [0.944, 0.971] · Δ -4.30% · speed +4.49% [+2.99%, +5.91%] · unresolved | 0.756 [0.734, 0.779] · Δ -24.39% · speed +32.25% [+28.37%, +36.19%] · unresolved |
| 65536 B instructions | 0.965 [0.958, 0.972] · Δ -3.54% · unresolved | 0.785 [0.778, 0.791] · Δ -21.50% · pass |
| 65536 B throughput | 1.029 [1.000, 1.061] · Δ +2.94% · speed +2.94% [-0.01%, +6.14%] · unresolved | 1.367 [1.254, 1.526] · Δ +36.74% · speed +36.74% [+25.36%, +52.59%] · unresolved |
| 65536 B allocsPerRequest | 0.927 [0.927, 0.927] · Δ -7.34% · pass | 0.930 [0.930, 0.930] · Δ -7.03% · pass |
| 65536 B bytesPerRequest | 0.965 [0.965, 0.965] · Δ -3.48% · pass | 0.964 [0.964, 0.964] · Δ -3.61% · pass |

Throughput ratios above one improve. CPU, instruction and allocation ratios below one improve. The JSON retains all own-baseline observations and every per-role A/A interval.

### Competitive P3 ratios

| Payload / compiler | CPU ratio to Wasmtime [95% CI]; controls | Throughput ratio to Wasmtime [95% CI]; controls |
| --- | ---: | ---: |
| 16 B go | 1.521 [1.458, 1.577] · Δ +52.13% · speed -34.27% [-36.60%, -31.41%] · unresolved | 1.096 [1.015, 1.214] · Δ +9.58% · speed +9.58% [+1.50%, +21.39%] · unresolved |
| 16 B llgo | 2.569 [2.417, 2.700] · Δ +156.86% · speed -61.07% [-62.96%, -58.63%] · unresolved | 0.536 [0.496, 0.581] · Δ -46.42% · speed -46.42% [-50.43%, -41.86%] · unresolved |
| 65536 B go | 1.884 [1.778, 1.970] · Δ +88.37% · speed -46.91% [-49.25%, -43.75%] · unresolved | 0.783 [0.746, 0.820] · Δ -21.75% · speed -21.75% [-25.37%, -17.95%] · unresolved |
| 65536 B llgo | 2.900 [2.786, 3.002] · Δ +190.04% · speed -65.52% [-66.69%, -64.11%] · unresolved | 0.476 [0.451, 0.503] · Δ -52.39% · speed -52.39% [-54.88%, -49.74%] · unresolved |

## Core source baseline

Head/base ratios below one improve. These marginal intervals describe this event; an apparent change in an unchanged mechanism requires diagnosis before being attributed to source changes. The machine-readable report includes every competitor ratio and A/A interval.

### native

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 1.152 [1.004, 1.370] · Δ +15.22% · speed -13.21% [-27.00%, -0.37%] · unresolved | 0.987 [0.842, 1.149] · Δ -1.33% · speed +1.35% [-12.95%, +18.80%] · unresolved |
| original22-v1/core/integer-loop | 0.981 [0.963, 1.004] · Δ -1.87% · speed +1.90% [-0.43%, +3.83%] · unresolved | 1.003 [0.968, 1.042] · Δ +0.26% · speed -0.26% [-4.01%, +3.29%] · unresolved |
| original22-v1/core/function-calls | 1.022 [0.993, 1.056] · Δ +2.21% · speed -2.16% [-5.34%, +0.71%] · unresolved | 1.009 [0.974, 1.044] · Δ +0.87% · speed -0.86% [-4.25%, +2.71%] · unresolved |
| original22-v1/core/memory-load-store | 0.915 [0.801, 1.047] · Δ -8.54% · speed +9.33% [-4.46%, +24.77%] · unresolved | 0.942 [0.847, 1.045] · Δ -5.83% · speed +6.19% [-4.32%, +18.02%] · unresolved |
| original22-v1/memory/narrow-unaligned | 1.064 [1.008, 1.124] · Δ +6.36% · speed -5.98% [-11.05%, -0.77%] · unresolved | 0.987 [0.934, 1.042] · Δ -1.34% · speed +1.36% [-4.02%, +7.04%] · unresolved |
| original22-v1/memory/dependent-address | 0.983 [0.964, 0.999] · Δ -1.67% · speed +1.69% [+0.13%, +3.73%] · unresolved | 1.005 [0.980, 1.033] · Δ +0.45% · speed -0.45% [-3.19%, +2.00%] · unresolved |
| original22-v1/memory/float64-unaligned | 0.995 [0.977, 1.022] · Δ -0.49% · speed +0.50% [-2.17%, +2.38%] · unresolved | 1.021 [0.997, 1.044] · Δ +2.10% · speed -2.05% [-4.26%, +0.27%] · unresolved |
| original22-v1/runtime/simd-memory-full | 1.014 [0.999, 1.033] · Δ +1.38% · speed -1.36% [-3.24%, +0.13%] · unresolved | 0.995 [0.968, 1.019] · Δ -0.50% · speed +0.50% [-1.89%, +3.26%] · unresolved |
| original22-v1/runtime/simd-memory-widen | 0.981 [0.952, 1.005] · Δ -1.91% · speed +1.95% [-0.53%, +5.09%] · unresolved | 1.006 [0.978, 1.030] · Δ +0.55% · speed -0.55% [-2.90%, +2.21%] · unresolved |
| original22-v1/runtime/simd-memory-lanes | 0.998 [0.982, 1.015] · Δ -0.20% · speed +0.20% [-1.52%, +1.87%] · unresolved | 1.009 [0.987, 1.029] · Δ +0.85% · speed -0.85% [-2.82%, +1.37%] · unresolved |
| original22-v1/gc/i31-roundtrip | 1.058 [0.951, 1.178] · Δ +5.84% · speed -5.52% [-15.08%, +5.19%] · unresolved | 0.985 [0.959, 1.011] · Δ -1.52% · speed +1.54% [-1.04%, +4.26%] · unresolved |
| original22-v1/gc/struct-access | 0.979 [0.888, 1.076] · Δ -2.08% · speed +2.12% [-7.02%, +12.55%] · unresolved | 0.991 [0.966, 1.014] · Δ -0.86% · speed +0.86% [-1.37%, +3.56%] · unresolved |
| original22-v1/gc/struct-allocation | 0.956 [0.913, 0.997] · Δ -4.44% · speed +4.65% [+0.26%, +9.48%] · unresolved | 1.001 [0.922, 1.064] · Δ +0.09% · speed -0.09% [-5.97%, +8.51%] · unresolved |
| original22-v1/gc/struct-indirect-access | 0.965 [0.823, 1.075] · Δ -3.49% · speed +3.62% [-7.01%, +21.47%] · unresolved | 0.813 [0.623, 1.016] · Δ -18.66% · speed +22.94% [-1.56%, +60.60%] · unresolved |
| original22-v1/threads/private-memory | 1.025 [0.947, 1.109] · Δ +2.50% · speed -2.44% [-9.84%, +5.57%] · unresolved | 0.985 [0.931, 1.040] · Δ -1.49% · speed +1.51% [-3.85%, +7.46%] · unresolved |
| original22-v1/threads/shared-memory | 1.072 [0.972, 1.195] · Δ +7.16% · speed -6.68% [-16.33%, +2.83%] · unresolved | 0.983 [0.915, 1.051] · Δ -1.73% · speed +1.76% [-4.85%, +9.33%] · unresolved |
| original22-v1/threads/atomic-rmw | 0.913 [0.848, 0.979] · Δ -8.73% · speed +9.57% [+2.19%, +17.87%] · unresolved | 0.966 [0.880, 1.054] · Δ -3.38% · speed +3.50% [-5.11%, +13.69%] · unresolved |
| original22-v1/threads/notify-empty | 1.168 [1.042, 1.313] · Δ +16.81% · speed -14.39% [-23.86%, -4.08%] · unresolved | 1.055 [0.889, 1.216] · Δ +5.48% · speed -5.20% [-17.74%, +12.52%] · unresolved |
| original22-v1/threads/wait-mismatch | 0.982 [0.901, 1.060] · Δ -1.83% · speed +1.86% [-5.69%, +11.04%] · unresolved | 1.086 [0.991, 1.249] · Δ +8.57% · speed -7.89% [-19.95%, +0.89%] · unresolved |
| original22-v1/threads/wait-timeout | 1.015 [0.843, 1.148] · Δ +1.51% · speed -1.49% [-12.90%, +18.67%] · unresolved | 0.999 [0.898, 1.098] · Δ -0.14% · speed +0.14% [-8.95%, +11.35%] · unresolved |
| original22-v1/applications/sha256-clang | 0.973 [0.891, 1.061] · Δ -2.75% · speed +2.82% [-5.76%, +12.19%] · unresolved | 1.063 [1.005, 1.131] · Δ +6.29% · speed -5.92% [-11.59%, -0.50%] · unresolved |
| original22-v1/applications/sha256-rust | 0.895 [0.837, 0.960] · Δ -10.55% · speed +11.79% [+4.14%, +19.51%] · unresolved | 1.062 [1.040, 1.089] · Δ +6.25% · speed -5.88% [-8.21%, -3.82%] · unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 1.000 [0.972, 1.033] · Δ -0.02% · speed +0.02% [-3.20%, +2.92%] · unresolved | 0.995 [0.966, 1.026] · Δ -0.51% · speed +0.51% [-2.53%, +3.56%] · unresolved |
| gc-sustained-v1/gc/struct-access | 1.006 [0.991, 1.018] · Δ +0.61% · speed -0.60% [-1.75%, +0.87%] · pass | 1.001 [0.982, 1.017] · Δ +0.14% · speed -0.14% [-1.67%, +1.86%] · pass |
| gc-sustained-v1/gc/struct-allocation | 1.007 [0.992, 1.023] · Δ +0.65% · speed -0.65% [-2.21%, +0.80%] · unresolved | 1.035 [0.993, 1.081] · Δ +3.45% · speed -3.34% [-7.50%, +0.71%] · unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 0.952 [0.911, 0.996] · Δ -4.77% · speed +5.01% [+0.36%, +9.79%] · unresolved | 1.022 [0.992, 1.056] · Δ +2.15% · speed -2.11% [-5.30%, +0.79%] · unresolved |
| threads-sustained-v2/threads/private-memory | 1.144 [1.014, 1.263] · Δ +14.41% · speed -12.60% [-20.81%, -1.35%] · unresolved | 0.880 [0.813, 0.949] · Δ -12.03% · speed +13.67% [+5.33%, +22.99%] · unresolved |
| threads-sustained-v2/threads/shared-memory | 1.021 [0.929, 1.120] · Δ +2.14% · speed -2.09% [-10.68%, +7.70%] · unresolved | 1.079 [1.022, 1.140] · Δ +7.92% · speed -7.34% [-12.27%, -2.12%] · unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.972 [0.929, 1.012] · Δ -2.76% · speed +2.84% [-1.22%, +7.65%] · unresolved | 0.989 [0.967, 1.009] · Δ -1.12% · speed +1.14% [-0.85%, +3.38%] · unresolved |
| threads-sustained-v2/threads/notify-empty | 1.047 [0.898, 1.242] · Δ +4.72% · speed -4.50% [-19.45%, +11.33%] · unresolved | 0.819 [0.638, 0.988] · Δ -18.09% · speed +22.08% [+1.23%, +56.84%] · unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.011 [0.915, 1.123] · Δ +1.08% · speed -1.07% [-10.98%, +9.33%] · unresolved | 0.984 [0.961, 1.006] · Δ -1.56% · speed +1.58% [-0.64%, +4.02%] · unresolved |
| threads-sustained-v2/threads/wait-timeout | 1.035 [0.930, 1.160] · Δ +3.52% · speed -3.40% [-13.81%, +7.58%] · unresolved | 1.029 [0.846, 1.230] · Δ +2.88% · speed -2.80% [-18.67%, +18.26%] · unresolved |
| generic-calls-v1/core/generic-calls | 0.980 [0.960, 0.998] · Δ -2.01% · speed +2.05% [+0.19%, +4.12%] · unresolved | 1.003 [0.988, 1.017] · Δ +0.30% · speed -0.30% [-1.70%, +1.19%] · unresolved |

### interpreter

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 0.910 [0.834, 1.000] · Δ -9.02% · speed +9.92% [-0.03%, +19.94%] · unresolved | 0.948 [0.793, 1.110] · Δ -5.18% · speed +5.46% [-9.93%, +26.07%] · unresolved |
| original22-v1/core/integer-loop | 0.992 [0.985, 0.998] · Δ -0.80% · speed +0.80% [+0.16%, +1.47%] · unresolved | 1.012 [0.993, 1.028] · Δ +1.21% · speed -1.20% [-2.74%, +0.68%] · unresolved |
| original22-v1/core/function-calls | 1.006 [0.988, 1.026] · Δ +0.62% · speed -0.61% [-2.57%, +1.20%] · unresolved | 1.014 [0.983, 1.048] · Δ +1.36% · speed -1.35% [-4.54%, +1.72%] · unresolved |
| original22-v1/core/memory-load-store | 1.077 [0.994, 1.181] · Δ +7.69% · speed -7.14% [-15.29%, +0.61%] · unresolved | 0.945 [0.870, 1.003] · Δ -5.48% · speed +5.80% [-0.29%, +14.99%] · unresolved |
| original22-v1/memory/narrow-unaligned | 0.966 [0.913, 1.006] · Δ -3.35% · speed +3.47% [-0.56%, +9.50%] · unresolved | 0.982 [0.878, 1.056] · Δ -1.76% · speed +1.79% [-5.32%, +13.89%] · unresolved |
| original22-v1/memory/dependent-address | 0.984 [0.916, 1.026] · Δ -1.60% · speed +1.62% [-2.56%, +9.16%] · unresolved | 0.987 [0.966, 1.008] · Δ -1.33% · speed +1.35% [-0.77%, +3.56%] · unresolved |
| original22-v1/memory/float64-unaligned | 0.934 [0.790, 1.036] · Δ -6.61% · speed +7.08% [-3.50%, +26.59%] · unresolved | 0.961 [0.909, 1.011] · Δ -3.95% · speed +4.11% [-1.05%, +10.03%] · unresolved |
| original22-v1/runtime/simd-memory-full | 1.058 [0.967, 1.180] · Δ +5.78% · speed -5.46% [-15.23%, +3.44%] · unresolved | 0.987 [0.940, 1.022] · Δ -1.29% · speed +1.31% [-2.17%, +6.41%] · unresolved |
| original22-v1/runtime/simd-memory-widen | 0.978 [0.950, 1.003] · Δ -2.16% · speed +2.21% [-0.28%, +5.28%] · unresolved | 1.026 [0.988, 1.101] · Δ +2.59% · speed -2.52% [-9.16%, +1.19%] · unresolved |
| original22-v1/runtime/simd-memory-lanes | 1.009 [0.932, 1.132] · Δ +0.92% · speed -0.91% [-11.66%, +7.30%] · unresolved | 1.174 [1.003, 1.313] · Δ +17.45% · speed -14.85% [-23.81%, -0.30%] · unresolved |
| original22-v1/gc/i31-roundtrip | 0.966 [0.907, 1.008] · Δ -3.35% · speed +3.47% [-0.77%, +10.29%] · unresolved | 1.022 [0.956, 1.112] · Δ +2.19% · speed -2.14% [-10.04%, +4.65%] · unresolved |
| original22-v1/gc/struct-access | 0.995 [0.970, 1.018] · Δ -0.53% · speed +0.53% [-1.76%, +3.04%] · unresolved | 1.072 [0.989, 1.184] · Δ +7.19% · speed -6.71% [-15.52%, +1.15%] · unresolved |
| original22-v1/gc/struct-allocation | 1.056 [1.022, 1.085] · Δ +5.63% · speed -5.33% [-7.80%, -2.16%] · unresolved | 1.004 [0.972, 1.030] · Δ +0.44% · speed -0.44% [-2.95%, +2.84%] · unresolved |
| original22-v1/gc/struct-indirect-access | 0.977 [0.955, 0.996] · Δ -2.33% · speed +2.38% [+0.37%, +4.71%] · unresolved | 1.005 [0.985, 1.028] · Δ +0.52% · speed -0.52% [-2.75%, +1.49%] · unresolved |
| original22-v1/threads/private-memory | 0.999 [0.967, 1.029] · Δ -0.13% · speed +0.13% [-2.80%, +3.36%] · unresolved | 1.001 [0.989, 1.015] · Δ +0.14% · speed -0.14% [-1.46%, +1.11%] · unresolved |
| original22-v1/threads/shared-memory | 1.027 [0.986, 1.070] · Δ +2.74% · speed -2.67% [-6.56%, +1.40%] · unresolved | 0.960 [0.888, 1.018] · Δ -4.05% · speed +4.22% [-1.74%, +12.60%] · unresolved |
| original22-v1/threads/atomic-rmw | 1.023 [0.975, 1.074] · Δ +2.33% · speed -2.28% [-6.92%, +2.52%] · unresolved | 0.960 [0.911, 1.009] · Δ -3.96% · speed +4.12% [-0.91%, +9.82%] · unresolved |
| original22-v1/threads/notify-empty | 0.979 [0.913, 1.026] · Δ -2.08% · speed +2.12% [-2.56%, +9.53%] · unresolved | 0.968 [0.931, 1.004] · Δ -3.22% · speed +3.33% [-0.36%, +7.44%] · unresolved |
| original22-v1/threads/wait-mismatch | 1.008 [0.886, 1.160] · Δ +0.82% · speed -0.81% [-13.82%, +12.84%] · unresolved | 0.925 [0.884, 0.967] · Δ -7.51% · speed +8.12% [+3.37%, +13.14%] · unresolved |
| original22-v1/threads/wait-timeout | 1.001 [0.965, 1.055] · Δ +0.07% · speed -0.07% [-5.23%, +3.65%] · unresolved | 1.025 [0.954, 1.108] · Δ +2.48% · speed -2.42% [-9.74%, +4.78%] · unresolved |
| original22-v1/applications/sha256-clang | 1.027 [0.926, 1.152] · Δ +2.72% · speed -2.65% [-13.20%, +7.95%] · unresolved | 1.009 [0.855, 1.207] · Δ +0.95% · speed -0.94% [-17.17%, +17.00%] · unresolved |
| original22-v1/applications/sha256-rust | 1.165 [0.855, 1.596] · Δ +16.45% · speed -14.13% [-37.36%, +16.95%] · unresolved | 1.090 [0.955, 1.299] · Δ +9.00% · speed -8.26% [-23.01%, +4.73%] · unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 0.999 [0.929, 1.058] · Δ -0.15% · speed +0.15% [-5.53%, +7.66%] · unresolved | 1.011 [0.948, 1.085] · Δ +1.11% · speed -1.10% [-7.83%, +5.44%] · unresolved |
| gc-sustained-v1/gc/struct-access | 1.004 [0.959, 1.056] · Δ +0.37% · speed -0.37% [-5.26%, +4.27%] · unresolved | 0.979 [0.962, 0.997] · Δ -2.07% · speed +2.11% [+0.30%, +3.96%] · unresolved |
| gc-sustained-v1/gc/struct-allocation | 1.043 [0.963, 1.132] · Δ +4.29% · speed -4.11% [-11.66%, +3.87%] · unresolved | 1.023 [0.978, 1.073] · Δ +2.32% · speed -2.27% [-6.82%, +2.23%] · unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.001 [0.986, 1.013] · Δ +0.11% · speed -0.11% [-1.25%, +1.37%] · unresolved | 1.017 [0.962, 1.087] · Δ +1.69% · speed -1.66% [-8.02%, +3.97%] · unresolved |
| threads-sustained-v2/threads/private-memory | 0.981 [0.921, 1.036] · Δ -1.94% · speed +1.98% [-3.51%, +8.55%] · unresolved | 1.060 [1.017, 1.129] · Δ +5.97% · speed -5.63% [-11.46%, -1.66%] · unresolved |
| threads-sustained-v2/threads/shared-memory | 0.984 [0.933, 1.033] · Δ -1.62% · speed +1.65% [-3.18%, +7.13%] · unresolved | 1.007 [0.969, 1.049] · Δ +0.72% · speed -0.71% [-4.67%, +3.25%] · unresolved |
| threads-sustained-v2/threads/atomic-rmw | 1.061 [1.015, 1.134] · Δ +6.06% · speed -5.72% [-11.83%, -1.46%] · unresolved | 1.031 [0.995, 1.076] · Δ +3.06% · speed -2.97% [-7.09%, +0.53%] · unresolved |
| threads-sustained-v2/threads/notify-empty | 1.005 [0.974, 1.047] · Δ +0.48% · speed -0.48% [-4.49%, +2.69%] · unresolved | 1.006 [0.940, 1.087] · Δ +0.56% · speed -0.56% [-8.03%, +6.38%] · unresolved |
| threads-sustained-v2/threads/wait-mismatch | 0.960 [0.918, 0.986] · Δ -4.02% · speed +4.19% [+1.42%, +8.98%] · unresolved | 1.027 [0.947, 1.169] · Δ +2.71% · speed -2.64% [-14.43%, +5.64%] · unresolved |
| threads-sustained-v2/threads/wait-timeout | 0.988 [0.960, 1.023] · Δ -1.21% · speed +1.23% [-2.28%, +4.14%] · unresolved | 0.999 [0.974, 1.031] · Δ -0.06% · speed +0.06% [-2.96%, +2.67%] · unresolved |
| generic-calls-v1/core/generic-calls | 0.993 [0.899, 1.081] · Δ -0.70% · speed +0.71% [-7.47%, +11.21%] · unresolved | 0.921 [0.830, 0.996] · Δ -7.86% · speed +8.53% [+0.41%, +20.50%] · unresolved |

## Preparation, first call and cleanup

All entries below are **microseconds per process or invocation**, including unresolved observations. Phase intervals are not inferred from warm-call controls. Startup is independently observed process readiness. Native compile/validate boundaries differ by API; WasmEdge AOT includes validation and artifact I/O, plus separately recorded native-image loading. Every process creates a fresh native image. First-call and warm execution have their own boundaries; values are never subtracted to invent pure costs.

<details>
<summary>Native: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5,658.75 | 8,032.38 | 5,797.12 | 6,185.83 | 8,230.19 | 21,052.02 |
| original22-v1/core/integer-loop | 7,006.69 | 9,121.67 | 7,395.46 | 6,729.21 | 8,659.35 | 23,455.19 |
| original22-v1/core/function-calls | 6,218.77 | 8,726.98 | 6,735.58 | 5,815.31 | 8,225.31 | 22,166.21 |
| original22-v1/core/memory-load-store | 7,313.15 | 10,539.56 | 7,326.35 | 8,060.44 | 9,221.60 | 23,390.46 |
| original22-v1/memory/narrow-unaligned | 6,508.10 | 8,929.21 | 5,917.56 | 7,025.85 | 8,880.38 | 22,638.35 |
| original22-v1/memory/dependent-address | 6,885.94 | 9,476.42 | 7,058.21 | 6,403.46 | 9,073.38 | 22,543.88 |
| original22-v1/memory/float64-unaligned | 6,176.77 | 9,182.81 | 7,284.40 | 7,147.12 | 8,126.98 | 22,253.83 |
| original22-v1/runtime/simd-memory-full | 6,012.77 | 8,985.69 | 7,580.10 | 6,050.52 | 8,455.10 | 23,040.04 |
| original22-v1/runtime/simd-memory-widen | 8,020.96 | 8,921.67 | 6,947.48 | 7,678.90 | 9,398.25 | 21,645.79 |
| original22-v1/runtime/simd-memory-lanes | 6,926.75 | 10,893.62 | 7,451.06 | 6,244.79 | 8,822.31 | 23,277.48 |
| original22-v1/gc/i31-roundtrip | 6,562.58 | 8,612.44 | US | 6,256.83 | US | 21,039.35 |
| original22-v1/gc/struct-access | 6,468.40 | 9,245.33 | US | 6,169.06 | US | 21,609.85 |
| original22-v1/gc/struct-allocation | 6,678.42 | 9,793.04 | US | 7,145.94 | US | 21,082.46 |
| original22-v1/gc/struct-indirect-access | 6,151.73 | 8,730.40 | US | 7,141.62 | US | 21,526.29 |
| original22-v1/threads/private-memory | 6,422.48 | 9,242.77 | 6,544.85 | 6,082.65 | 9,340.65 | 24,235.96 |
| original22-v1/threads/shared-memory | 5,974.56 | 8,494.42 | 5,659.44 | 6,898.56 | 8,813.58 | 20,803.96 |
| original22-v1/threads/atomic-rmw | 6,355.00 | 9,704.27 | 6,342.56 | 6,493.46 | 9,191.56 | 23,510.04 |
| original22-v1/threads/notify-empty | 7,017.44 | 9,012.46 | 7,509.35 | 6,611.65 | 8,465.52 | 22,190.52 |
| original22-v1/threads/wait-mismatch | 7,389.17 | 9,026.79 | 8,627.17 | 6,176.81 | 9,012.27 | 21,823.52 |
| original22-v1/threads/wait-timeout | 6,415.92 | 8,860.83 | 7,167.40 | 5,964.56 | 9,155.50 | 21,760.52 |
| original22-v1/applications/sha256-clang | 7,359.12 | 9,518.44 | 7,602.73 | 6,834.88 | 9,623.92 | 22,237.90 |
| original22-v1/applications/sha256-rust | 7,559.08 | 11,178.54 | 6,551.15 | 7,344.88 | 9,306.94 | 23,493.12 |
| gc-sustained-v1/gc/i31-roundtrip | 7,826.48 | 9,007.10 | US | 7,710.79 | US | 23,271.10 |
| gc-sustained-v1/gc/struct-access | 5,886.08 | 10,116.79 | US | 7,070.02 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 6,694.56 | 9,324.19 | US | 7,088.52 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 6,881.75 | 9,624.88 | US | 6,879.48 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 6,695.63 | 10,275.71 | 6,505.62 | 6,649.42 | 9,090.04 | 23,487.27 |
| threads-sustained-v2/threads/shared-memory | 6,896.75 | 10,856.29 | 7,194.27 | 6,904.25 | 8,956.06 | 22,380.92 |
| threads-sustained-v2/threads/atomic-rmw | 6,074.52 | 8,084.44 | 5,729.48 | 6,802.50 | 7,903.44 | 20,976.33 |
| threads-sustained-v2/threads/notify-empty | 6,928.62 | 9,129.02 | 6,879.67 | 6,759.19 | 8,627.54 | 23,154.06 |
| threads-sustained-v2/threads/wait-mismatch | 6,290.08 | 8,264.27 | 6,301.48 | 6,030.75 | 9,171.23 | 20,985.15 |
| threads-sustained-v2/threads/wait-timeout | 6,933.65 | 9,699.79 | 7,395.25 | 5,983.98 | 7,876.94 | 22,055.27 |
| generic-calls-v1/core/generic-calls | 6,699.71 | 9,610.23 | 6,693.48 | 6,665.44 | 10,449.81 | 24,397.87 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 59.772 | 67.812 | 47.041 | 38.084 | 47.312 | 53.270 |
| original22-v1/core/integer-loop | 53.458 | 63.188 | 58.605 | 51.875 | 53.479 | 73.541 |
| original22-v1/core/function-calls | 53.230 | 67.709 | 64.646 | 53.500 | 84.105 | 68.541 |
| original22-v1/core/memory-load-store | 72.604 | 59.480 | 54.229 | 67.291 | 63.354 | 62.688 |
| original22-v1/memory/narrow-unaligned | 58.228 | 62.833 | 96.521 | 45.438 | 47.020 | 65.000 |
| original22-v1/memory/dependent-address | 64.334 | 69.042 | 70.854 | 76.104 | 54.541 | 64.875 |
| original22-v1/memory/float64-unaligned | 54.666 | 86.542 | 57.312 | 66.312 | 54.000 | 72.625 |
| original22-v1/runtime/simd-memory-full | 54.896 | 66.750 | 66.291 | 61.438 | 55.626 | 59.166 |
| original22-v1/runtime/simd-memory-widen | 63.834 | 56.020 | 49.938 | 45.125 | 70.562 | 55.791 |
| original22-v1/runtime/simd-memory-lanes | 92.979 | 67.938 | 62.895 | 79.730 | 61.396 | 63.520 |
| original22-v1/gc/i31-roundtrip | 51.041 | 65.230 | US | 46.229 | US | 48.479 |
| original22-v1/gc/struct-access | 55.188 | 54.958 | US | 49.562 | US | 69.625 |
| original22-v1/gc/struct-allocation | 49.875 | 71.375 | US | 40.521 | US | 57.146 |
| original22-v1/gc/struct-indirect-access | 52.166 | 55.979 | US | 48.896 | US | 65.833 |
| original22-v1/threads/private-memory | 41.938 | 48.834 | 49.479 | 40.146 | 55.000 | 64.270 |
| original22-v1/threads/shared-memory | 56.166 | 62.104 | 54.625 | 43.104 | 50.916 | 50.480 |
| original22-v1/threads/atomic-rmw | 58.229 | 70.167 | 50.688 | 40.083 | 49.229 | 59.021 |
| original22-v1/threads/notify-empty | 48.688 | 50.729 | 47.708 | 55.480 | 44.375 | 53.520 |
| original22-v1/threads/wait-mismatch | 51.334 | 60.833 | 51.354 | 66.145 | 42.146 | 79.896 |
| original22-v1/threads/wait-timeout | 52.001 | 51.896 | 47.230 | 43.105 | 42.709 | 56.375 |
| original22-v1/applications/sha256-clang | 56.521 | 61.500 | 50.666 | 54.896 | 47.584 | 82.834 |
| original22-v1/applications/sha256-rust | 66.584 | 105.81 | 48.854 | 71.084 | 54.730 | 67.020 |
| gc-sustained-v1/gc/i31-roundtrip | 57.938 | 68.438 | US | 59.958 | US | 82.562 |
| gc-sustained-v1/gc/struct-access | 52.272 | 57.480 | US | 45.916 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 46.833 | 57.104 | US | 51.354 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 71.125 | 65.042 | US | 44.459 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 63.709 | 73.459 | 57.500 | 55.750 | 43.604 | 77.688 |
| threads-sustained-v2/threads/shared-memory | 48.709 | 60.959 | 51.999 | 51.770 | 49.417 | 52.813 |
| threads-sustained-v2/threads/atomic-rmw | 48.604 | 58.687 | 48.396 | 43.834 | 42.584 | 56.438 |
| threads-sustained-v2/threads/notify-empty | 59.000 | 67.750 | 57.584 | 57.188 | 44.584 | 60.292 |
| threads-sustained-v2/threads/wait-mismatch | 45.855 | 57.938 | 45.375 | 39.500 | 43.438 | 55.417 |
| threads-sustained-v2/threads/wait-timeout | 56.084 | 62.751 | 57.875 | 37.438 | 38.938 | 54.271 |
| generic-calls-v1/core/generic-calls | 72.938 | 60.209 | 67.125 | 50.333 | 51.854 | 69.583 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 7.396 | 7.458 | 69.479 | 162.38 | 52.458 | 39.521 |
| original22-v1/core/integer-loop | 7.792 | 7.625 | 82.562 | 171.69 | 53.125 | 47.771 |
| original22-v1/core/function-calls | 8.084 | 6.958 | 70.688 | 175.52 | 61.334 | 35.959 |
| original22-v1/core/memory-load-store | 8.208 | 6.833 | 72.209 | 187.12 | 56.022 | 49.625 |
| original22-v1/memory/narrow-unaligned | 7.667 | 6.667 | 99.000 | 160.25 | 56.604 | 39.084 |
| original22-v1/memory/dependent-address | 7.792 | 7.083 | 73.896 | 173.81 | 50.917 | 35.562 |
| original22-v1/memory/float64-unaligned | 8.041 | 7.917 | 70.355 | 188.58 | 54.917 | 44.042 |
| original22-v1/runtime/simd-memory-full | 7.938 | 7.646 | 82.083 | 216.56 | 52.230 | 37.708 |
| original22-v1/runtime/simd-memory-widen | 8.374 | 6.771 | 74.896 | 171.50 | 53.479 | 51.687 |
| original22-v1/runtime/simd-memory-lanes | 9.479 | 7.167 | 107.33 | 259.58 | 54.500 | 53.126 |
| original22-v1/gc/i31-roundtrip | 7.938 | 7.437 | US | 161.77 | US | 37.584 |
| original22-v1/gc/struct-access | 8.271 | 7.000 | US | 180.54 | US | 46.250 |
| original22-v1/gc/struct-allocation | 8.001 | 7.396 | US | 157.12 | US | 50.145 |
| original22-v1/gc/struct-indirect-access | 8.646 | 6.688 | US | 186.46 | US | 40.416 |
| original22-v1/threads/private-memory | 7.062 | 6.208 | 70.896 | 172.40 | 63.646 | 45.459 |
| original22-v1/threads/shared-memory | 7.958 | 8.376 | 68.959 | 164.10 | 56.041 | 35.730 |
| original22-v1/threads/atomic-rmw | 8.896 | 7.646 | 76.291 | 157.98 | 55.896 | 41.480 |
| original22-v1/threads/notify-empty | 7.896 | 6.250 | 72.105 | 162.96 | 53.584 | 39.125 |
| original22-v1/threads/wait-mismatch | 8.229 | 8.312 | 71.125 | 216.17 | 50.833 | 49.188 |
| original22-v1/threads/wait-timeout | 8.021 | 6.292 | 70.146 | 159.83 | 51.126 | 38.334 |
| original22-v1/applications/sha256-clang | 8.000 | 7.187 | 76.583 | 161.83 | 54.625 | 58.688 |
| original22-v1/applications/sha256-rust | 8.208 | 7.021 | 73.750 | 204.52 | 51.854 | 46.938 |
| gc-sustained-v1/gc/i31-roundtrip | 8.062 | 7.333 | US | 180.33 | US | 44.666 |
| gc-sustained-v1/gc/struct-access | 7.125 | 6.792 | US | 164.08 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 7.667 | 6.750 | US | 169.27 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 8.271 | 7.521 | US | 161.62 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 9.416 | 8.958 | 80.291 | 184.52 | 62.521 | 53.666 |
| threads-sustained-v2/threads/shared-memory | 7.542 | 7.938 | 73.521 | 180.56 | 62.021 | 41.250 |
| threads-sustained-v2/threads/atomic-rmw | 8.208 | 6.729 | 77.896 | 218.69 | 53.917 | 44.479 |
| threads-sustained-v2/threads/notify-empty | 8.062 | 7.646 | 76.562 | 157.00 | 52.333 | 33.666 |
| threads-sustained-v2/threads/wait-mismatch | 8.375 | 6.604 | 71.354 | 157.58 | 52.333 | 42.709 |
| threads-sustained-v2/threads/wait-timeout | 7.771 | 7.688 | 73.271 | 164.85 | 49.729 | 34.749 |
| generic-calls-v1/core/generic-calls | 9.812 | 6.896 | 73.771 | 161.38 | 58.813 | 55.895 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 234.83 | 205.23 | 112.23 | 1,079.17 | 1,199.71 | 290,628.65 |
| original22-v1/core/integer-loop | 297.31 | 240.62 | 161.50 | 1,104.62 | 1,397.29 | 281,233.10 |
| original22-v1/core/function-calls | 374.65 | 331.40 | 163.71 | 1,549.62 | 1,541.73 | 289,748.15 |
| original22-v1/core/memory-load-store | 439.75 | 372.71 | 170.87 | 1,326.98 | 1,589.98 | 302,347.54 |
| original22-v1/memory/narrow-unaligned | 382.52 | 351.58 | 196.92 | 1,062.73 | 1,301.23 | 301,812.42 |
| original22-v1/memory/dependent-address | 358.40 | 364.75 | 180.98 | 1,228.50 | 1,670.42 | 289,896.33 |
| original22-v1/memory/float64-unaligned | 432.73 | 557.21 | 184.92 | 1,227.19 | 1,577.17 | 287,162.75 |
| original22-v1/runtime/simd-memory-full | 353.56 | 319.69 | 162.58 | 1,511.42 | 1,549.98 | 294,863.27 |
| original22-v1/runtime/simd-memory-widen | 334.92 | 298.17 | 157.83 | 2,184.65 | 1,352.56 | 293,523.46 |
| original22-v1/runtime/simd-memory-lanes | 466.92 | 310.25 | 207.00 | 1,585.85 | 1,609.10 | 311,413.98 |
| original22-v1/gc/i31-roundtrip | 258.69 | 209.08 | US | 1,173.75 | US | 306,687.69 |
| original22-v1/gc/struct-access | 302.10 | 256.71 | US | 1,466.96 | US | 312,876.60 |
| original22-v1/gc/struct-allocation | 283.71 | 247.27 | US | 4,558.29 | US | 281,429.04 |
| original22-v1/gc/struct-indirect-access | 430.23 | 305.29 | US | 2,376.02 | US | 323,205.23 |
| original22-v1/threads/private-memory | 301.48 | 236.50 | 154.63 | 1,271.33 | 1,366.27 | 289,427.81 |
| original22-v1/threads/shared-memory | 289.58 | 239.21 | 161.77 | 1,060.98 | 1,259.54 | 284,693.25 |
| original22-v1/threads/atomic-rmw | 306.62 | 237.08 | 149.04 | 1,094.65 | 1,457.94 | 285,298.37 |
| original22-v1/threads/notify-empty | 281.88 | 223.69 | 166.79 | 1,122.12 | 1,288.12 | 295,163.00 |
| original22-v1/threads/wait-mismatch | 337.33 | 239.94 | 159.46 | 1,395.38 | 1,614.08 | 287,956.17 |
| original22-v1/threads/wait-timeout | 324.46 | 217.52 | 164.40 | 1,347.79 | 1,466.62 | 284,862.56 |
| original22-v1/applications/sha256-clang | 1,910.50 | 2,357.88 | 384.48 | 2,579.92 | 2,720.83 | 290,886.90 |
| original22-v1/applications/sha256-rust | 1,856.48 | 2,509.81 | 410.17 | 10,352.04 | 2,587.92 | 308,270.00 |
| gc-sustained-v1/gc/i31-roundtrip | 264.23 | 203.21 | US | 2,725.67 | US | 295,649.29 |
| gc-sustained-v1/gc/struct-access | 304.75 | 276.23 | US | 1,279.46 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 274.56 | 227.88 | US | 2,112.54 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 369.50 | 315.88 | US | 2,279.77 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 381.90 | 356.29 | 163.75 | 1,164.96 | 1,393.29 | 309,173.92 |
| threads-sustained-v2/threads/shared-memory | 302.79 | 271.38 | 168.58 | 3,132.04 | 1,472.02 | 328,871.79 |
| threads-sustained-v2/threads/atomic-rmw | 273.23 | 212.02 | 152.83 | 1,227.73 | 1,502.54 | 327,764.46 |
| threads-sustained-v2/threads/notify-empty | 331.50 | 233.12 | 155.12 | 1,818.92 | 1,527.23 | 293,804.27 |
| threads-sustained-v2/threads/wait-mismatch | 275.88 | 236.17 | 163.29 | 1,489.62 | 1,242.29 | 291,987.73 |
| threads-sustained-v2/threads/wait-timeout | 305.67 | 240.81 | 164.48 | 1,331.00 | 1,380.79 | 298,016.58 |
| generic-calls-v1/core/generic-calls | 459.02 | 342.52 | 161.56 | 1,796.04 | 2,264.12 | 298,602.92 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 10.625 | 10.771 | 16.791 | 10.041 | 22.021 | 22.438 |
| original22-v1/core/integer-loop | 9.417 | 10.729 | 17.813 | 9.438 | 23.041 | 18.854 |
| original22-v1/core/function-calls | 11.541 | 10.916 | 17.562 | 8.834 | 21.479 | 17.417 |
| original22-v1/core/memory-load-store | 17.104 | 17.646 | 22.583 | 26.209 | 40.208 | 41.479 |
| original22-v1/memory/narrow-unaligned | 17.000 | 16.855 | 21.771 | 26.209 | 38.167 | 39.959 |
| original22-v1/memory/dependent-address | 17.270 | 17.395 | 22.980 | 27.167 | 35.874 | 39.667 |
| original22-v1/memory/float64-unaligned | 17.229 | 20.271 | 21.875 | 25.583 | 38.791 | 40.916 |
| original22-v1/runtime/simd-memory-full | 19.084 | 20.271 | 26.791 | 68.146 | 42.604 | 53.250 |
| original22-v1/runtime/simd-memory-widen | 20.062 | 20.396 | 23.313 | 71.145 | 42.500 | 49.624 |
| original22-v1/runtime/simd-memory-lanes | 20.938 | 18.291 | 26.042 | 67.270 | 43.874 | 51.855 |
| original22-v1/gc/i31-roundtrip | 10.624 | 10.542 | US | 9.354 | US | 18.459 |
| original22-v1/gc/struct-access | 12.104 | 11.166 | US | 28.917 | US | 19.646 |
| original22-v1/gc/struct-allocation | 11.917 | 12.749 | US | 63.041 | US | 23.896 |
| original22-v1/gc/struct-indirect-access | 28.208 | 27.250 | US | 82.458 | US | 33.583 |
| original22-v1/threads/private-memory | 15.875 | 17.146 | 20.688 | 27.500 | 39.583 | 44.417 |
| original22-v1/threads/shared-memory | 16.666 | 16.666 | 21.855 | 26.604 | 46.312 | 43.208 |
| original22-v1/threads/atomic-rmw | 17.709 | 20.438 | 21.417 | 26.812 | 45.834 | 42.896 |
| original22-v1/threads/notify-empty | 17.562 | 18.250 | 22.146 | 24.688 | 44.124 | 42.271 |
| original22-v1/threads/wait-mismatch | 18.542 | 18.875 | 23.041 | 27.084 | 47.875 | 44.250 |
| original22-v1/threads/wait-timeout | 17.604 | 17.812 | 22.605 | 25.812 | 42.395 | 54.750 |
| original22-v1/applications/sha256-clang | 26.167 | 24.084 | 20.396 | 68.334 | 47.729 | 52.750 |
| original22-v1/applications/sha256-rust | 24.854 | 24.209 | 31.708 | 73.625 | 44.354 | 53.396 |
| gc-sustained-v1/gc/i31-roundtrip | 10.229 | 10.209 | US | 9.854 | US | 17.958 |
| gc-sustained-v1/gc/struct-access | 12.021 | 11.624 | US | 27.834 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 11.646 | 11.354 | US | 66.125 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 26.395 | 26.646 | US | 81.166 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 18.250 | 19.896 | 24.438 | 35.750 | 39.646 | 47.334 |
| threads-sustained-v2/threads/shared-memory | 17.500 | 16.270 | 22.541 | 27.770 | 40.334 | 42.438 |
| threads-sustained-v2/threads/atomic-rmw | 17.042 | 18.605 | 23.166 | 28.520 | 44.166 | 41.771 |
| threads-sustained-v2/threads/notify-empty | 19.375 | 18.750 | 23.812 | 26.313 | 46.229 | 39.730 |
| threads-sustained-v2/threads/wait-mismatch | 17.542 | 18.333 | 22.792 | 27.188 | 42.459 | 42.062 |
| threads-sustained-v2/threads/wait-timeout | 17.230 | 20.938 | 22.895 | 27.229 | 42.250 | 40.438 |
| generic-calls-v1/core/generic-calls | 12.812 | 13.041 | 17.729 | 10.500 | 22.125 | 18.209 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.292 | 0.479 | 2.708 | 4.771 | 5.562 | 0.312 |
| original22-v1/core/integer-loop | 0.271 | 0.542 | 2.646 | 5.229 | 5.958 | 0.333 |
| original22-v1/core/function-calls | 0.312 | 0.541 | 2.896 | 4.708 | 5.562 | 0.292 |
| original22-v1/core/memory-load-store | 0.312 | 0.562 | 4.021 | 4.979 | 6.063 | 0.292 |
| original22-v1/memory/narrow-unaligned | 0.334 | 0.500 | 2.542 | 5.062 | 6.042 | 0.334 |
| original22-v1/memory/dependent-address | 0.333 | 0.542 | 3.125 | 5.104 | 5.438 | 0.333 |
| original22-v1/memory/float64-unaligned | 0.333 | 0.584 | 2.396 | 4.917 | 6.021 | 0.313 |
| original22-v1/runtime/simd-memory-full | 0.292 | 0.542 | 2.812 | 2.438 | 5.854 | 0.312 |
| original22-v1/runtime/simd-memory-widen | 0.333 | 0.583 | 2.729 | 2.333 | 6.271 | 0.230 |
| original22-v1/runtime/simd-memory-lanes | 0.334 | 0.542 | 3.542 | 2.312 | 5.792 | 0.334 |
| original22-v1/gc/i31-roundtrip | 0.292 | 0.479 | US | 4.771 | US | 0.291 |
| original22-v1/gc/struct-access | 0.312 | 0.520 | US | 4.833 | US | 0.312 |
| original22-v1/gc/struct-allocation | 0.271 | 0.500 | US | 2.188 | US | 0.334 |
| original22-v1/gc/struct-indirect-access | 0.333 | 0.562 | US | 2.625 | US | 0.292 |
| original22-v1/threads/private-memory | 0.312 | 0.541 | 2.375 | 5.250 | 6.208 | 0.312 |
| original22-v1/threads/shared-memory | 0.334 | 0.583 | 2.979 | 5.042 | 5.875 | 0.292 |
| original22-v1/threads/atomic-rmw | 0.333 | 0.583 | 3.375 | 4.979 | 6.333 | 0.312 |
| original22-v1/threads/notify-empty | 0.292 | 0.500 | 2.708 | 5.000 | 5.625 | 0.312 |
| original22-v1/threads/wait-mismatch | 0.334 | 0.542 | 2.834 | 5.312 | 6.229 | 0.375 |
| original22-v1/threads/wait-timeout | 0.292 | 0.542 | 3.083 | 5.000 | 5.854 | 0.334 |
| original22-v1/applications/sha256-clang | 0.333 | 0.521 | 2.958 | 2.500 | 6.854 | 0.334 |
| original22-v1/applications/sha256-rust | 0.334 | 0.584 | 3.562 | 2.542 | 7.375 | 0.292 |
| gc-sustained-v1/gc/i31-roundtrip | 0.292 | 0.520 | US | 4.896 | US | 0.291 |
| gc-sustained-v1/gc/struct-access | 0.312 | 0.458 | US | 5.167 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 0.312 | 0.500 | US | 2.375 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 0.354 | 0.541 | US | 2.500 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.354 | 0.708 | 2.875 | 6.250 | 5.979 | 0.291 |
| threads-sustained-v2/threads/shared-memory | 0.334 | 0.584 | 3.271 | 5.354 | 5.708 | 0.271 |
| threads-sustained-v2/threads/atomic-rmw | 0.312 | 0.562 | 3.271 | 5.542 | 5.729 | 0.291 |
| threads-sustained-v2/threads/notify-empty | 0.312 | 0.625 | 3.208 | 4.688 | 6.062 | 0.250 |
| threads-sustained-v2/threads/wait-mismatch | 0.375 | 0.583 | 2.979 | 4.958 | 5.729 | 0.250 |
| threads-sustained-v2/threads/wait-timeout | 0.354 | 0.563 | 3.312 | 4.833 | 5.833 | 0.333 |
| generic-calls-v1/core/generic-calls | 0.292 | 0.459 | 3.021 | 5.125 | 6.083 | 0.292 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 12.729 | 9.688 | 10.521 | 36.792 | 16.271 | 17.750 |
| original22-v1/core/integer-loop | 2,072.79 | 2,002.08 | 2,034.85 | 2,006.00 | 2,022.60 | 2,046.67 |
| original22-v1/core/function-calls | 1,014.83 | 1,039.29 | 1,693.27 | 1,191.12 | 1,000.31 | 1,017.60 |
| original22-v1/core/memory-load-store | 559.42 | 570.62 | 1,044.54 | 711.19 | 703.58 | 366.23 |
| original22-v1/memory/narrow-unaligned | 844.35 | 873.42 | 1,244.02 | 925.40 | 967.90 | 679.71 |
| original22-v1/memory/dependent-address | 2,385.02 | 2,448.23 | 3,797.85 | 2,844.13 | 2,768.46 | 1,954.98 |
| original22-v1/memory/float64-unaligned | 947.54 | 1,006.29 | 1,017.54 | 939.42 | 958.90 | 893.46 |
| original22-v1/runtime/simd-memory-full | 4,248.08 | 4,267.08 | 4,744.50 | 4,560.77 | 4,354.79 | 4,997.08 |
| original22-v1/runtime/simd-memory-widen | 1,304.23 | 1,321.67 | 1,317.17 | 1,324.27 | 1,353.77 | 1,320.00 |
| original22-v1/runtime/simd-memory-lanes | 6,395.25 | 6,295.10 | 7,400.00 | 7,354.06 | 7,105.85 | 6,547.10 |
| original22-v1/gc/i31-roundtrip | 31.688 | 26.355 | US | 57.833 | US | 33.812 |
| original22-v1/gc/struct-access | 39.584 | 38.083 | US | 66.896 | US | 157.31 |
| original22-v1/gc/struct-allocation | 500.48 | 506.94 | US | 183.96 | US | 636.88 |
| original22-v1/gc/struct-indirect-access | 36.770 | 31.271 | US | 38.126 | US | 422.23 |
| original22-v1/threads/private-memory | 18.521 | 16.438 | 24.895 | 55.354 | 35.145 | 37.458 |
| original22-v1/threads/shared-memory | 30.500 | 28.292 | 23.750 | 56.375 | 32.417 | 39.479 |
| original22-v1/threads/atomic-rmw | 40.041 | 36.146 | 33.480 | 64.458 | 68.230 | 46.750 |
| original22-v1/threads/notify-empty | 26.896 | 22.750 | 278.79 | 123.92 | 74.208 | 136.75 |
| original22-v1/threads/wait-mismatch | 29.312 | 25.625 | 330.46 | 157.90 | 51.979 | 63.666 |
| original22-v1/threads/wait-timeout | 28.125 | 23.188 | 3,684.85 | 597.02 | 2,345.85 | 945.27 |
| original22-v1/applications/sha256-clang | 1,382.08 | 1,285.35 | 1,770.08 | 1,216.52 | 1,292.54 | 1,341.50 |
| original22-v1/applications/sha256-rust | 1,351.62 | 1,397.48 | 1,866.71 | 1,231.67 | 1,219.85 | 1,342.04 |
| gc-sustained-v1/gc/i31-roundtrip | 1,737.42 | 1,701.87 | US | 2,403.50 | US | 1,736.42 |
| gc-sustained-v1/gc/struct-access | 1,378.62 | 1,360.81 | US | 1,378.69 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 33,433.04 | 21,818.02 | US | 4,174.67 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 2,194.21 | 2,247.33 | US | 3,311.44 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 51.417 | 49.209 | 147.85 | 185.27 | 147.35 | 142.75 |
| threads-sustained-v2/threads/shared-memory | 170.02 | 166.58 | 140.15 | 170.60 | 127.12 | 152.44 |
| threads-sustained-v2/threads/atomic-rmw | 242.67 | 238.15 | 239.48 | 272.77 | 528.71 | 249.85 |
| threads-sustained-v2/threads/notify-empty | 132.77 | 124.56 | 2,904.42 | 888.29 | 592.40 | 1,204.04 |
| threads-sustained-v2/threads/wait-mismatch | 147.56 | 141.98 | 3,215.29 | 1,250.65 | 313.48 | 412.81 |
| threads-sustained-v2/threads/wait-timeout | 134.50 | 126.10 | 37,998.96 | 5,666.62 | 22,698.33 | 8,078.85 |
| generic-calls-v1/core/generic-calls | 6,727.02 | 6,557.67 | 13,198.62 | 14,027.85 | 13,368.62 | 6,556.90 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.958 | 6.875 | 4.958 | 70.709 | 28.312 | 436.58 |
| original22-v1/core/integer-loop | 7.688 | 7.667 | 5.208 | 93.000 | 42.313 | 493.33 |
| original22-v1/core/function-calls | 8.916 | 11.021 | 12.312 | 105.42 | 39.438 | 486.50 |
| original22-v1/core/memory-load-store | 19.688 | 18.938 | 10.750 | 103.96 | 75.375 | 538.33 |
| original22-v1/memory/narrow-unaligned | 17.959 | 9.646 | 5.417 | 107.73 | 61.916 | 533.52 |
| original22-v1/memory/dependent-address | 13.834 | 9.625 | 11.042 | 100.06 | 68.584 | 549.77 |
| original22-v1/memory/float64-unaligned | 21.584 | 16.104 | 12.896 | 138.52 | 64.625 | 519.58 |
| original22-v1/runtime/simd-memory-full | 19.146 | 10.626 | 5.729 | 107.83 | 63.479 | 533.40 |
| original22-v1/runtime/simd-memory-widen | 16.855 | 13.124 | 9.416 | 110.69 | 73.770 | 518.81 |
| original22-v1/runtime/simd-memory-lanes | 22.521 | 20.333 | 12.583 | 158.50 | 78.875 | 541.71 |
| original22-v1/gc/i31-roundtrip | 6.104 | 5.500 | US | 75.438 | US | 447.02 |
| original22-v1/gc/struct-access | 14.501 | 12.479 | US | 119.54 | US | 504.65 |
| original22-v1/gc/struct-allocation | 171.12 | 122.17 | US | 133.38 | US | 2,571.44 |
| original22-v1/gc/struct-indirect-access | 18.584 | 12.667 | US | 125.67 | US | 508.67 |
| original22-v1/threads/private-memory | 17.021 | 13.021 | 9.584 | 112.12 | 69.645 | 499.94 |
| original22-v1/threads/shared-memory | 18.645 | 8.812 | 12.042 | 103.65 | 66.250 | 535.35 |
| original22-v1/threads/atomic-rmw | 22.167 | 14.041 | 9.666 | 116.08 | 78.916 | 559.96 |
| original22-v1/threads/notify-empty | 13.104 | 10.166 | 5.875 | 91.166 | 61.562 | 530.75 |
| original22-v1/threads/wait-mismatch | 11.666 | 11.896 | 10.124 | 97.895 | 63.875 | 500.12 |
| original22-v1/threads/wait-timeout | 20.562 | 11.396 | 4.375 | 103.92 | 66.980 | 494.69 |
| original22-v1/applications/sha256-clang | 25.020 | 20.125 | 11.334 | 161.40 | 90.625 | 552.29 |
| original22-v1/applications/sha256-rust | 26.020 | 22.562 | 14.541 | 187.71 | 138.21 | 563.50 |
| gc-sustained-v1/gc/i31-roundtrip | 9.437 | 9.833 | US | 98.896 | US | 467.96 |
| gc-sustained-v1/gc/struct-access | 25.312 | 19.104 | US | 125.10 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 3,919.42 | 2,534.88 | US | 916.75 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 21.896 | 17.438 | US | 124.98 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 10.521 | 17.688 | 10.001 | 115.00 | 66.645 | 498.69 |
| threads-sustained-v2/threads/shared-memory | 20.750 | 18.166 | 15.062 | 120.27 | 69.291 | 537.38 |
| threads-sustained-v2/threads/atomic-rmw | 12.874 | 7.208 | 5.646 | 89.834 | 70.146 | 516.27 |
| threads-sustained-v2/threads/notify-empty | 13.750 | 11.334 | 5.833 | 101.56 | 60.916 | 529.33 |
| threads-sustained-v2/threads/wait-mismatch | 23.334 | 13.125 | 5.021 | 109.10 | 74.605 | 493.38 |
| threads-sustained-v2/threads/wait-timeout | 23.771 | 10.563 | 3.833 | 95.625 | 66.958 | 526.83 |
| generic-calls-v1/core/generic-calls | 14.979 | 15.146 | 15.125 | 146.65 | 55.812 | 537.23 |

</details>

<details>
<summary>Interpreter: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5,765.42 | 7,578.02 | 6,120.33 | NM | NM | 21,563.54 |
| original22-v1/core/integer-loop | 6,490.15 | 9,574.35 | 5,935.83 | NM | NM | 21,504.92 |
| original22-v1/core/function-calls | 6,371.58 | 9,330.65 | 5,723.88 | NM | NM | 23,191.62 |
| original22-v1/core/memory-load-store | 7,762.04 | 8,658.04 | 7,586.42 | NM | NM | 24,551.77 |
| original22-v1/memory/narrow-unaligned | 6,623.31 | 9,078.21 | 6,902.15 | NM | NM | 21,903.71 |
| original22-v1/memory/dependent-address | 6,334.15 | 9,924.56 | 7,085.54 | NM | NM | 23,340.69 |
| original22-v1/memory/float64-unaligned | 6,174.60 | 9,996.15 | 7,507.56 | NM | NM | 22,427.85 |
| original22-v1/runtime/simd-memory-full | 6,116.60 | 9,361.65 | 6,606.77 | NM | NM | 23,172.81 |
| original22-v1/runtime/simd-memory-widen | 7,005.19 | 8,450.77 | 6,281.19 | NM | NM | 23,215.75 |
| original22-v1/runtime/simd-memory-lanes | 6,435.17 | 8,652.67 | 6,618.96 | NM | NM | 22,775.40 |
| original22-v1/gc/i31-roundtrip | 6,950.58 | 8,893.04 | US | NM | NM | 22,048.44 |
| original22-v1/gc/struct-access | 7,541.15 | 8,748.65 | US | NM | NM | 21,626.40 |
| original22-v1/gc/struct-allocation | 5,722.56 | 8,946.67 | US | NM | NM | 21,817.85 |
| original22-v1/gc/struct-indirect-access | 7,020.33 | 9,525.71 | US | NM | NM | 21,840.42 |
| original22-v1/threads/private-memory | 6,727.73 | 10,805.81 | 6,799.50 | NM | NM | 22,123.88 |
| original22-v1/threads/shared-memory | 6,769.67 | 9,484.92 | 6,760.10 | NM | NM | 21,614.13 |
| original22-v1/threads/atomic-rmw | 7,611.31 | 9,398.90 | 6,838.56 | NM | NM | 21,893.46 |
| original22-v1/threads/notify-empty | 6,905.56 | 8,372.29 | 6,660.65 | NM | NM | 21,001.15 |
| original22-v1/threads/wait-mismatch | 5,884.42 | 9,938.33 | 6,289.62 | NM | NM | 22,844.58 |
| original22-v1/threads/wait-timeout | 5,955.33 | 9,221.79 | 7,660.06 | NM | NM | 22,645.65 |
| original22-v1/applications/sha256-clang | 5,968.71 | 9,792.21 | 6,944.65 | NM | NM | 23,191.65 |
| original22-v1/applications/sha256-rust | 8,298.10 | 10,417.52 | 7,567.12 | NM | NM | 24,577.92 |
| gc-sustained-v1/gc/i31-roundtrip | 7,844.81 | 9,985.94 | US | NM | NM | 24,047.00 |
| gc-sustained-v1/gc/struct-access | 6,962.40 | 8,561.23 | US | NM | NM | 22,822.42 |
| gc-sustained-v1/gc/struct-allocation | 6,730.04 | 8,736.58 | US | NM | NM | 23,112.58 |
| gc-sustained-v1/gc/struct-indirect-access | 6,742.15 | 10,516.23 | US | NM | NM | 23,647.25 |
| threads-sustained-v2/threads/private-memory | 6,361.79 | 8,703.88 | 7,008.69 | NM | NM | 22,025.77 |
| threads-sustained-v2/threads/shared-memory | 6,880.00 | 9,072.23 | 7,163.19 | NM | NM | 23,030.94 |
| threads-sustained-v2/threads/atomic-rmw | 5,811.06 | 8,904.02 | 7,776.04 | NM | NM | 23,507.12 |
| threads-sustained-v2/threads/notify-empty | 6,275.58 | 9,101.31 | 7,071.44 | NM | NM | 22,994.52 |
| threads-sustained-v2/threads/wait-mismatch | 5,765.04 | 8,580.85 | 7,086.04 | NM | NM | 21,501.88 |
| threads-sustained-v2/threads/wait-timeout | 6,152.06 | 9,332.98 | 6,656.02 | NM | NM | 22,936.50 |
| generic-calls-v1/core/generic-calls | 6,385.94 | 9,103.15 | 6,651.85 | NM | NM | 23,113.40 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 61.062 | 50.646 | 50.709 | NM | NM | 56.188 |
| original22-v1/core/integer-loop | 94.417 | 76.834 | 81.209 | NM | NM | 60.104 |
| original22-v1/core/function-calls | 64.188 | 59.312 | 64.021 | NM | NM | 86.104 |
| original22-v1/core/memory-load-store | 73.166 | 76.875 | 68.062 | NM | NM | 76.520 |
| original22-v1/memory/narrow-unaligned | 57.688 | 82.916 | 49.604 | NM | NM | 52.459 |
| original22-v1/memory/dependent-address | 53.438 | 78.875 | 54.854 | NM | NM | 60.541 |
| original22-v1/memory/float64-unaligned | 56.666 | 60.313 | 59.666 | NM | NM | 66.605 |
| original22-v1/runtime/simd-memory-full | 52.688 | 59.209 | 61.499 | NM | NM | 90.312 |
| original22-v1/runtime/simd-memory-widen | 56.688 | 70.813 | 54.854 | NM | NM | 59.646 |
| original22-v1/runtime/simd-memory-lanes | 78.229 | 86.917 | 94.834 | NM | NM | 72.958 |
| original22-v1/gc/i31-roundtrip | 60.021 | 50.896 | US | NM | NM | 87.459 |
| original22-v1/gc/struct-access | 52.520 | 51.396 | US | NM | NM | 55.416 |
| original22-v1/gc/struct-allocation | 54.750 | 53.270 | US | NM | NM | 47.604 |
| original22-v1/gc/struct-indirect-access | 67.834 | 48.396 | US | NM | NM | 53.125 |
| original22-v1/threads/private-memory | 58.750 | 67.459 | 47.312 | NM | NM | 55.584 |
| original22-v1/threads/shared-memory | 52.478 | 55.938 | 52.354 | NM | NM | 55.667 |
| original22-v1/threads/atomic-rmw | 50.604 | 58.438 | 52.541 | NM | NM | 57.499 |
| original22-v1/threads/notify-empty | 54.062 | 48.125 | 47.688 | NM | NM | 51.479 |
| original22-v1/threads/wait-mismatch | 53.376 | 68.437 | 47.959 | NM | NM | 61.645 |
| original22-v1/threads/wait-timeout | 75.791 | 71.833 | 66.146 | NM | NM | 67.999 |
| original22-v1/applications/sha256-clang | 61.605 | 61.583 | 55.353 | NM | NM | 74.646 |
| original22-v1/applications/sha256-rust | 54.188 | 75.980 | 93.084 | NM | NM | 64.771 |
| gc-sustained-v1/gc/i31-roundtrip | 50.584 | 64.000 | US | NM | NM | 71.605 |
| gc-sustained-v1/gc/struct-access | 56.812 | 59.125 | US | NM | NM | 72.188 |
| gc-sustained-v1/gc/struct-allocation | 66.416 | 63.625 | US | NM | NM | 59.312 |
| gc-sustained-v1/gc/struct-indirect-access | 58.208 | 70.938 | US | NM | NM | 63.979 |
| threads-sustained-v2/threads/private-memory | 79.334 | 48.834 | 61.042 | NM | NM | 89.584 |
| threads-sustained-v2/threads/shared-memory | 53.020 | 65.874 | 43.834 | NM | NM | 48.584 |
| threads-sustained-v2/threads/atomic-rmw | 57.146 | 49.749 | 53.605 | NM | NM | 65.521 |
| threads-sustained-v2/threads/notify-empty | 55.312 | 66.187 | 52.730 | NM | NM | 58.771 |
| threads-sustained-v2/threads/wait-mismatch | 54.312 | 51.000 | 46.750 | NM | NM | 55.979 |
| threads-sustained-v2/threads/wait-timeout | 66.021 | 70.355 | 49.354 | NM | NM | 63.875 |
| generic-calls-v1/core/generic-calls | 55.375 | 93.896 | 63.042 | NM | NM | 57.605 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.917 | 6.562 | 2.728 | NM | NM | 37.666 |
| original22-v1/core/integer-loop | 6.542 | 7.000 | 4.042 | NM | NM | 34.938 |
| original22-v1/core/function-calls | 6.708 | 6.083 | 3.542 | NM | NM | 57.167 |
| original22-v1/core/memory-load-store | 6.688 | 8.104 | 3.292 | NM | NM | 51.937 |
| original22-v1/memory/narrow-unaligned | 6.208 | 7.250 | 3.250 | NM | NM | 38.084 |
| original22-v1/memory/dependent-address | 5.125 | 8.917 | 3.354 | NM | NM | 38.188 |
| original22-v1/memory/float64-unaligned | 5.833 | 6.188 | 3.229 | NM | NM | 42.479 |
| original22-v1/runtime/simd-memory-full | 6.250 | 6.958 | 3.604 | NM | NM | 52.688 |
| original22-v1/runtime/simd-memory-widen | 5.396 | 6.750 | 2.896 | NM | NM | 37.562 |
| original22-v1/runtime/simd-memory-lanes | 5.834 | 6.646 | 3.750 | NM | NM | 68.395 |
| original22-v1/gc/i31-roundtrip | 7.125 | 6.208 | US | NM | NM | 49.125 |
| original22-v1/gc/struct-access | 6.438 | 7.104 | US | NM | NM | 44.645 |
| original22-v1/gc/struct-allocation | 6.312 | 6.521 | US | NM | NM | 35.458 |
| original22-v1/gc/struct-indirect-access | 6.229 | 6.958 | US | NM | NM | 37.478 |
| original22-v1/threads/private-memory | 6.792 | 6.583 | 3.000 | NM | NM | 37.688 |
| original22-v1/threads/shared-memory | 6.104 | 6.417 | 3.042 | NM | NM | 36.688 |
| original22-v1/threads/atomic-rmw | 6.292 | 8.291 | 3.188 | NM | NM | 43.125 |
| original22-v1/threads/notify-empty | 6.167 | 6.187 | 3.125 | NM | NM | 36.312 |
| original22-v1/threads/wait-mismatch | 6.833 | 7.479 | 2.876 | NM | NM | 44.730 |
| original22-v1/threads/wait-timeout | 8.250 | 7.354 | 3.958 | NM | NM | 51.708 |
| original22-v1/applications/sha256-clang | 6.562 | 6.604 | 3.167 | NM | NM | 47.312 |
| original22-v1/applications/sha256-rust | 6.125 | 6.979 | 3.771 | NM | NM | 42.979 |
| gc-sustained-v1/gc/i31-roundtrip | 6.000 | 6.875 | US | NM | NM | 36.355 |
| gc-sustained-v1/gc/struct-access | 6.417 | 6.938 | US | NM | NM | 36.479 |
| gc-sustained-v1/gc/struct-allocation | 6.646 | 6.875 | US | NM | NM | 49.437 |
| gc-sustained-v1/gc/struct-indirect-access | 6.000 | 8.166 | US | NM | NM | 35.230 |
| threads-sustained-v2/threads/private-memory | 6.125 | 6.438 | 3.521 | NM | NM | 39.979 |
| threads-sustained-v2/threads/shared-memory | 6.687 | 6.583 | 3.126 | NM | NM | 37.666 |
| threads-sustained-v2/threads/atomic-rmw | 5.729 | 6.813 | 3.250 | NM | NM | 39.750 |
| threads-sustained-v2/threads/notify-empty | 7.000 | 8.688 | 3.312 | NM | NM | 36.312 |
| threads-sustained-v2/threads/wait-mismatch | 6.729 | 6.583 | 3.062 | NM | NM | 40.270 |
| threads-sustained-v2/threads/wait-timeout | 6.188 | 7.854 | 3.000 | NM | NM | 35.230 |
| generic-calls-v1/core/generic-calls | 6.042 | 7.938 | 3.479 | NM | NM | 36.209 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 67.584 | 58.874 | 43.312 | NM | NM | 41.896 |
| original22-v1/core/integer-loop | 89.479 | 91.084 | 68.041 | NM | NM | 43.770 |
| original22-v1/core/function-calls | 96.875 | 83.292 | 56.584 | NM | NM | 49.792 |
| original22-v1/core/memory-load-store | 103.12 | 104.75 | 59.666 | NM | NM | 51.250 |
| original22-v1/memory/narrow-unaligned | 108.08 | 114.52 | 60.459 | NM | NM | 52.416 |
| original22-v1/memory/dependent-address | 93.000 | 102.62 | 57.104 | NM | NM | 49.354 |
| original22-v1/memory/float64-unaligned | 101.65 | 99.271 | 62.708 | NM | NM | 53.041 |
| original22-v1/runtime/simd-memory-full | 97.291 | 87.334 | 66.376 | NM | NM | 61.666 |
| original22-v1/runtime/simd-memory-widen | 90.541 | 80.771 | 57.645 | NM | NM | 49.999 |
| original22-v1/runtime/simd-memory-lanes | 104.17 | 89.125 | 64.667 | NM | NM | 50.812 |
| original22-v1/gc/i31-roundtrip | 77.667 | 71.625 | US | NM | NM | 50.771 |
| original22-v1/gc/struct-access | 104.48 | 75.479 | US | NM | NM | 65.062 |
| original22-v1/gc/struct-allocation | 85.020 | 71.355 | US | NM | NM | 62.416 |
| original22-v1/gc/struct-indirect-access | 106.58 | 103.29 | US | NM | NM | 62.647 |
| original22-v1/threads/private-memory | 81.021 | 76.042 | 50.208 | NM | NM | 46.083 |
| original22-v1/threads/shared-memory | 76.416 | 67.146 | 52.334 | NM | NM | 46.875 |
| original22-v1/threads/atomic-rmw | 77.895 | 70.375 | 56.291 | NM | NM | 46.895 |
| original22-v1/threads/notify-empty | 79.166 | 65.959 | 54.916 | NM | NM | 48.480 |
| original22-v1/threads/wait-mismatch | 86.312 | 86.896 | 51.834 | NM | NM | 49.000 |
| original22-v1/threads/wait-timeout | 99.562 | 75.020 | 62.729 | NM | NM | 48.376 |
| original22-v1/applications/sha256-clang | 382.88 | 457.83 | 104.90 | NM | NM | 109.15 |
| original22-v1/applications/sha256-rust | 374.65 | 449.73 | 106.10 | NM | NM | 109.65 |
| gc-sustained-v1/gc/i31-roundtrip | 77.084 | 67.583 | US | NM | NM | 43.083 |
| gc-sustained-v1/gc/struct-access | 90.000 | 77.521 | US | NM | NM | 54.999 |
| gc-sustained-v1/gc/struct-allocation | 85.709 | 73.583 | US | NM | NM | 59.063 |
| gc-sustained-v1/gc/struct-indirect-access | 105.83 | 105.52 | US | NM | NM | 60.813 |
| threads-sustained-v2/threads/private-memory | 80.834 | 68.000 | 54.374 | NM | NM | 54.270 |
| threads-sustained-v2/threads/shared-memory | 84.145 | 68.312 | 51.416 | NM | NM | 46.646 |
| threads-sustained-v2/threads/atomic-rmw | 74.916 | 69.105 | 98.646 | NM | NM | 46.999 |
| threads-sustained-v2/threads/notify-empty | 81.042 | 75.896 | 53.875 | NM | NM | 48.167 |
| threads-sustained-v2/threads/wait-mismatch | 83.667 | 69.917 | 55.209 | NM | NM | 50.584 |
| threads-sustained-v2/threads/wait-timeout | 84.604 | 93.042 | 54.938 | NM | NM | 46.584 |
| generic-calls-v1/core/generic-calls | 87.666 | 96.230 | 48.667 | NM | NM | 55.104 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 9.854 | 8.771 | 18.979 | NM | NM | 31.855 |
| original22-v1/core/integer-loop | 10.812 | 10.001 | 23.042 | NM | NM | 34.834 |
| original22-v1/core/function-calls | 9.854 | 9.166 | 20.645 | NM | NM | 42.876 |
| original22-v1/core/memory-load-store | 15.209 | 17.584 | 21.937 | NM | NM | 70.770 |
| original22-v1/memory/narrow-unaligned | 14.708 | 15.750 | 21.709 | NM | NM | 62.562 |
| original22-v1/memory/dependent-address | 14.854 | 19.438 | 20.166 | NM | NM | 70.334 |
| original22-v1/memory/float64-unaligned | 14.666 | 16.729 | 20.687 | NM | NM | 62.959 |
| original22-v1/runtime/simd-memory-full | 16.916 | 17.146 | 22.583 | NM | NM | 93.875 |
| original22-v1/runtime/simd-memory-widen | 16.625 | 17.395 | 20.459 | NM | NM | 72.354 |
| original22-v1/runtime/simd-memory-lanes | 17.313 | 16.625 | 21.541 | NM | NM | 74.271 |
| original22-v1/gc/i31-roundtrip | 10.000 | 9.334 | US | NM | NM | 41.312 |
| original22-v1/gc/struct-access | 10.125 | 10.312 | US | NM | NM | 32.625 |
| original22-v1/gc/struct-allocation | 12.729 | 11.229 | US | NM | NM | 37.563 |
| original22-v1/gc/struct-indirect-access | 18.438 | 19.917 | US | NM | NM | 44.562 |
| original22-v1/threads/private-memory | 13.875 | 16.209 | 19.354 | NM | NM | 59.645 |
| original22-v1/threads/shared-memory | 19.209 | 17.042 | 21.459 | NM | NM | 60.459 |
| original22-v1/threads/atomic-rmw | 19.291 | 18.125 | 21.854 | NM | NM | 57.271 |
| original22-v1/threads/notify-empty | 19.979 | 16.500 | 20.916 | NM | NM | 61.020 |
| original22-v1/threads/wait-mismatch | 20.270 | 18.480 | 20.355 | NM | NM | 56.916 |
| original22-v1/threads/wait-timeout | 21.208 | 18.980 | 21.020 | NM | NM | 60.874 |
| original22-v1/applications/sha256-clang | 18.813 | 29.125 | 24.500 | NM | NM | 69.771 |
| original22-v1/applications/sha256-rust | 18.833 | 107.85 | 28.146 | NM | NM | 72.626 |
| gc-sustained-v1/gc/i31-roundtrip | 9.438 | 8.854 | US | NM | NM | 31.000 |
| gc-sustained-v1/gc/struct-access | 11.354 | 10.396 | US | NM | NM | 34.270 |
| gc-sustained-v1/gc/struct-allocation | 12.229 | 11.916 | US | NM | NM | 49.584 |
| gc-sustained-v1/gc/struct-indirect-access | 19.480 | 22.855 | US | NM | NM | 41.980 |
| threads-sustained-v2/threads/private-memory | 14.750 | 16.375 | 24.250 | NM | NM | 62.188 |
| threads-sustained-v2/threads/shared-memory | 20.084 | 18.271 | 21.896 | NM | NM | 58.770 |
| threads-sustained-v2/threads/atomic-rmw | 19.166 | 19.709 | 21.667 | NM | NM | 55.854 |
| threads-sustained-v2/threads/notify-empty | 19.729 | 19.416 | 20.562 | NM | NM | 59.834 |
| threads-sustained-v2/threads/wait-mismatch | 20.605 | 17.375 | 21.688 | NM | NM | 61.145 |
| threads-sustained-v2/threads/wait-timeout | 19.688 | 21.855 | 20.562 | NM | NM | 55.895 |
| generic-calls-v1/core/generic-calls | 10.667 | 13.479 | 19.500 | NM | NM | 32.812 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.250 | 0.500 | 0.292 | NM | NM | 0.208 |
| original22-v1/core/integer-loop | 0.292 | 0.584 | 0.292 | NM | NM | 0.271 |
| original22-v1/core/function-calls | 0.291 | 0.500 | 0.375 | NM | NM | 0.250 |
| original22-v1/core/memory-load-store | 0.312 | 0.521 | 0.333 | NM | NM | 0.271 |
| original22-v1/memory/narrow-unaligned | 0.250 | 0.520 | 0.312 | NM | NM | 0.230 |
| original22-v1/memory/dependent-address | 0.271 | 0.479 | 0.292 | NM | NM | 0.250 |
| original22-v1/memory/float64-unaligned | 0.333 | 0.479 | 0.250 | NM | NM | 0.292 |
| original22-v1/runtime/simd-memory-full | 0.292 | 0.520 | 0.292 | NM | NM | 0.208 |
| original22-v1/runtime/simd-memory-widen | 0.250 | 0.542 | 0.312 | NM | NM | 0.250 |
| original22-v1/runtime/simd-memory-lanes | 0.334 | 0.500 | 0.292 | NM | NM | 0.271 |
| original22-v1/gc/i31-roundtrip | 0.292 | 0.541 | US | NM | NM | 0.250 |
| original22-v1/gc/struct-access | 0.312 | 0.459 | US | NM | NM | 0.250 |
| original22-v1/gc/struct-allocation | 0.292 | 0.438 | US | NM | NM | 0.209 |
| original22-v1/gc/struct-indirect-access | 0.292 | 0.521 | US | NM | NM | 0.208 |
| original22-v1/threads/private-memory | 0.250 | 0.521 | 0.271 | NM | NM | 0.250 |
| original22-v1/threads/shared-memory | 0.312 | 0.500 | 0.250 | NM | NM | 0.250 |
| original22-v1/threads/atomic-rmw | 0.334 | 0.604 | 0.333 | NM | NM | 0.230 |
| original22-v1/threads/notify-empty | 0.292 | 0.500 | 0.291 | NM | NM | 0.250 |
| original22-v1/threads/wait-mismatch | 0.333 | 0.500 | 0.354 | NM | NM | 0.250 |
| original22-v1/threads/wait-timeout | 0.334 | 0.541 | 0.333 | NM | NM | 0.250 |
| original22-v1/applications/sha256-clang | 0.292 | 0.520 | 0.271 | NM | NM | 0.291 |
| original22-v1/applications/sha256-rust | 0.312 | 0.562 | 0.291 | NM | NM | 0.271 |
| gc-sustained-v1/gc/i31-roundtrip | 0.250 | 0.479 | US | NM | NM | 0.209 |
| gc-sustained-v1/gc/struct-access | 0.333 | 0.417 | US | NM | NM | 0.250 |
| gc-sustained-v1/gc/struct-allocation | 0.334 | 0.541 | US | NM | NM | 0.271 |
| gc-sustained-v1/gc/struct-indirect-access | 0.271 | 0.521 | US | NM | NM | 0.271 |
| threads-sustained-v2/threads/private-memory | 0.292 | 0.479 | 0.333 | NM | NM | 0.250 |
| threads-sustained-v2/threads/shared-memory | 0.354 | 0.542 | 0.312 | NM | NM | 0.250 |
| threads-sustained-v2/threads/atomic-rmw | 0.312 | 0.521 | 0.313 | NM | NM | 0.250 |
| threads-sustained-v2/threads/notify-empty | 0.292 | 0.562 | 0.313 | NM | NM | 0.208 |
| threads-sustained-v2/threads/wait-mismatch | 0.376 | 0.542 | 0.312 | NM | NM | 0.271 |
| threads-sustained-v2/threads/wait-timeout | 0.333 | 0.583 | 0.271 | NM | NM | 0.230 |
| generic-calls-v1/core/generic-calls | 0.292 | 0.542 | 0.250 | NM | NM | 0.230 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 7.312 | 8.500 | 4.458 | NM | NM | 8.979 |
| original22-v1/core/integer-loop | 40,174.19 | 35,833.88 | 205,148.90 | NM | NM | 88,106.73 |
| original22-v1/core/function-calls | 65,491.75 | 59,743.58 | 127,164.77 | NM | NM | 82,871.56 |
| original22-v1/core/memory-load-store | 98,420.81 | 84,742.38 | 258,733.92 | NM | NM | 237,491.79 |
| original22-v1/memory/narrow-unaligned | 108,557.15 | 97,434.75 | 246,359.96 | NM | NM | 156,882.96 |
| original22-v1/memory/dependent-address | 53,876.92 | 45,463.06 | 137,440.50 | NM | NM | 74,426.62 |
| original22-v1/memory/float64-unaligned | 87,976.46 | 76,839.50 | 212,465.19 | NM | NM | 111,143.40 |
| original22-v1/runtime/simd-memory-full | 100,943.17 | 89,580.08 | 263,357.17 | NM | NM | 117,867.52 |
| original22-v1/runtime/simd-memory-widen | 81,662.02 | 71,727.12 | 193,736.04 | NM | NM | 78,743.04 |
| original22-v1/runtime/simd-memory-lanes | 138,095.46 | 143,122.81 | 279,401.12 | NM | NM | 208,508.79 |
| original22-v1/gc/i31-roundtrip | 495.58 | 431.29 | US | NM | NM | 578.77 |
| original22-v1/gc/struct-access | 563.90 | 565.19 | US | NM | NM | 752.79 |
| original22-v1/gc/struct-allocation | 1,233.75 | 1,296.38 | US | NM | NM | 1,528.08 |
| original22-v1/gc/struct-indirect-access | 991.58 | 966.06 | US | NM | NM | 1,242.06 |
| original22-v1/threads/private-memory | 572.62 | 484.38 | 796.81 | NM | NM | 607.35 |
| original22-v1/threads/shared-memory | 1,024.35 | 417.73 | 814.83 | NM | NM | 507.65 |
| original22-v1/threads/atomic-rmw | 694.27 | 448.71 | 618.40 | NM | NM | 434.62 |
| original22-v1/threads/notify-empty | 406.88 | 350.88 | 520.62 | NM | NM | 517.54 |
| original22-v1/threads/wait-mismatch | 826.27 | 502.60 | 902.17 | NM | NM | 581.83 |
| original22-v1/threads/wait-timeout | 861.35 | 503.52 | 4,366.56 | NM | NM | 1,348.52 |
| original22-v1/applications/sha256-clang | 150,472.06 | 105,187.19 | 469,972.04 | NM | NM | 179,907.96 |
| original22-v1/applications/sha256-rust | 151,418.25 | 119,375.27 | 573,220.69 | NM | NM | 217,411.98 |
| gc-sustained-v1/gc/i31-roundtrip | 47,495.21 | 42,241.44 | US | NM | NM | 70,624.40 |
| gc-sustained-v1/gc/struct-access | 52,547.19 | 53,848.23 | US | NM | NM | 72,122.25 |
| gc-sustained-v1/gc/struct-allocation | 64,855.79 | 46,634.27 | US | NM | NM | 40,833.56 |
| gc-sustained-v1/gc/struct-indirect-access | 108,185.42 | 96,015.94 | US | NM | NM | 134,510.98 |
| threads-sustained-v2/threads/private-memory | 5,014.10 | 4,339.25 | 8,433.42 | NM | NM | 5,454.73 |
| threads-sustained-v2/threads/shared-memory | 9,887.92 | 4,280.96 | 8,338.73 | NM | NM | 5,110.46 |
| threads-sustained-v2/threads/atomic-rmw | 6,327.25 | 4,310.02 | 6,108.90 | NM | NM | 4,094.06 |
| threads-sustained-v2/threads/notify-empty | 4,075.73 | 3,769.29 | 5,602.44 | NM | NM | 5,608.25 |
| threads-sustained-v2/threads/wait-mismatch | 6,994.40 | 4,717.60 | 8,830.60 | NM | NM | 5,525.06 |
| threads-sustained-v2/threads/wait-timeout | 7,149.46 | 4,517.52 | 45,780.90 | NM | NM | 13,790.17 |
| generic-calls-v1/core/generic-calls | 761,586.92 | 754,111.85 | 1,827,126.33 | NM | NM | 1,102,266.79 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2.542 | 1.979 | 6.250 | NM | NM | 7.312 |
| original22-v1/core/integer-loop | 2.354 | 2.188 | 5.125 | NM | NM | 9.084 |
| original22-v1/core/function-calls | 2.458 | 2.104 | 4.729 | NM | NM | 11.249 |
| original22-v1/core/memory-load-store | 3.000 | 2.500 | 9.312 | NM | NM | 55.562 |
| original22-v1/memory/narrow-unaligned | 2.751 | 2.479 | 4.979 | NM | NM | 57.396 |
| original22-v1/memory/dependent-address | 2.875 | 2.583 | 5.083 | NM | NM | 54.584 |
| original22-v1/memory/float64-unaligned | 3.167 | 2.500 | 7.771 | NM | NM | 60.895 |
| original22-v1/runtime/simd-memory-full | 2.917 | 2.333 | 5.375 | NM | NM | 54.959 |
| original22-v1/runtime/simd-memory-widen | 2.958 | 2.479 | 8.688 | NM | NM | 56.084 |
| original22-v1/runtime/simd-memory-lanes | 5.083 | 2.438 | 10.666 | NM | NM | 59.833 |
| original22-v1/gc/i31-roundtrip | 2.333 | 2.438 | US | NM | NM | 6.708 |
| original22-v1/gc/struct-access | 2.562 | 2.063 | US | NM | NM | 7.521 |
| original22-v1/gc/struct-allocation | 117.56 | 68.041 | US | NM | NM | 2,047.38 |
| original22-v1/gc/struct-indirect-access | 2.542 | 2.208 | US | NM | NM | 8.583 |
| original22-v1/threads/private-memory | 2.624 | 2.229 | 12.021 | NM | NM | 52.001 |
| original22-v1/threads/shared-memory | 16.895 | 7.708 | 7.771 | NM | NM | 51.500 |
| original22-v1/threads/atomic-rmw | 25.500 | 9.250 | 10.042 | NM | NM | 52.708 |
| original22-v1/threads/notify-empty | 13.312 | 8.188 | 5.562 | NM | NM | 51.312 |
| original22-v1/threads/wait-mismatch | 12.187 | 8.562 | 10.562 | NM | NM | 51.771 |
| original22-v1/threads/wait-timeout | 19.041 | 8.396 | 4.167 | NM | NM | 53.209 |
| original22-v1/applications/sha256-clang | 2.771 | 2.229 | 9.979 | NM | NM | 65.958 |
| original22-v1/applications/sha256-rust | 3.875 | 2.417 | 15.249 | NM | NM | 75.459 |
| gc-sustained-v1/gc/i31-roundtrip | 2.458 | 2.354 | US | NM | NM | 9.791 |
| gc-sustained-v1/gc/struct-access | 2.667 | 2.542 | US | NM | NM | 11.688 |
| gc-sustained-v1/gc/struct-allocation | 3,346.29 | 2,067.25 | US | NM | NM | 69,643.69 |
| gc-sustained-v1/gc/struct-indirect-access | 2.604 | 2.292 | US | NM | NM | 11.209 |
| threads-sustained-v2/threads/private-memory | 3.000 | 2.375 | 7.083 | NM | NM | 56.666 |
| threads-sustained-v2/threads/shared-memory | 22.084 | 13.249 | 11.104 | NM | NM | 61.917 |
| threads-sustained-v2/threads/atomic-rmw | 12.625 | 7.167 | 4.958 | NM | NM | 53.646 |
| threads-sustained-v2/threads/notify-empty | 16.145 | 8.104 | 5.688 | NM | NM | 54.229 |
| threads-sustained-v2/threads/wait-mismatch | 15.646 | 7.792 | 9.271 | NM | NM | 56.584 |
| threads-sustained-v2/threads/wait-timeout | 14.438 | 10.001 | 4.375 | NM | NM | 53.541 |
| generic-calls-v1/core/generic-calls | 2.875 | 2.167 | 4.542 | NM | NM | 19.145 |

</details>

### WasmEdge AOT preparation detail

| Workload | Compile + validation + artifact I/O, us | Native image load, us | Post-load validation, us |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 26,255.54 | 255,914.81 | 3.458 |
| original22-v1/core/integer-loop | 28,572.54 | 251,715.25 | 6.875 |
| original22-v1/core/function-calls | 28,126.58 | 253,623.06 | 5.688 |
| original22-v1/core/memory-load-store | 34,644.83 | 260,664.98 | 7.958 |
| original22-v1/memory/narrow-unaligned | 28,806.29 | 270,438.31 | 9.500 |
| original22-v1/memory/dependent-address | 31,317.77 | 255,301.38 | 8.166 |
| original22-v1/memory/float64-unaligned | 33,979.73 | 255,005.42 | 9.562 |
| original22-v1/runtime/simd-memory-full | 27,875.94 | 255,002.92 | 8.229 |
| original22-v1/runtime/simd-memory-widen | 25,944.88 | 266,601.83 | 8.104 |
| original22-v1/runtime/simd-memory-lanes | 35,927.35 | 282,923.19 | 7.542 |
| original22-v1/gc/i31-roundtrip | 24,987.88 | 282,258.29 | 6.250 |
| original22-v1/gc/struct-access | 28,357.71 | 263,952.65 | 3.562 |
| original22-v1/gc/struct-allocation | 25,949.50 | 253,023.08 | 7.125 |
| original22-v1/gc/struct-indirect-access | 27,754.35 | 285,937.27 | 11.000 |
| original22-v1/threads/private-memory | 28,226.73 | 259,811.17 | 6.375 |
| original22-v1/threads/shared-memory | 25,633.69 | 259,419.77 | 6.375 |
| original22-v1/threads/atomic-rmw | 29,081.60 | 257,538.15 | 7.562 |
| original22-v1/threads/notify-empty | 25,342.06 | 265,180.19 | 7.375 |
| original22-v1/threads/wait-mismatch | 28,249.44 | 257,432.83 | 7.833 |
| original22-v1/threads/wait-timeout | 27,300.54 | 257,752.44 | 7.729 |
| original22-v1/applications/sha256-clang | 36,607.35 | 253,987.25 | 12.916 |
| original22-v1/applications/sha256-rust | 37,908.71 | 258,816.21 | 12.688 |
| gc-sustained-v1/gc/i31-roundtrip | 28,736.56 | 261,663.42 | 7.583 |
| gc-sustained-v1/gc/struct-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 26,642.65 | 281,334.98 | 6.937 |
| threads-sustained-v2/threads/shared-memory | 27,479.44 | 293,974.46 | 7.125 |
| threads-sustained-v2/threads/atomic-rmw | 25,513.10 | 301,617.90 | 5.833 |
| threads-sustained-v2/threads/notify-empty | 31,320.21 | 260,100.83 | 6.521 |
| threads-sustained-v2/threads/wait-mismatch | 26,120.69 | 255,622.40 | 5.646 |
| threads-sustained-v2/threads/wait-timeout | 28,673.08 | 267,006.58 | 5.729 |
| generic-calls-v1/core/generic-calls | 30,842.75 | 265,661.02 | 6.396 |

### P3 phase observations

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B / cleanup | 80.166 | 1,022.71 | NM | 328.10 | NM | NM |
| 16 B / compileValidate | 105,515.04 | 200,978.48 | NM | 66,552.96 | NM | NM |
| 16 B / engineSetup | 0.646 | 0.250 | NM | 267.69 | NM | NM |
| 16 B / executorSetup | N/A | N/A | NM | 59.562 | NM | NM |
| 16 B / exportLookup | 0.875 | 1.313 | NM | 7.833 | NM | NM |
| 16 B / firstCall | 12,680.79 | 30,540.08 | NM | 22,994.23 | NM | NM |
| 16 B / instantiate | 574.15 | 1,200.33 | NM | 242.46 | NM | NM |
| 16 B / load | 118.19 | 86.625 | NM | 95.959 | NM | NM |
| 16 B / processStartup | 9,296.94 | 17,415.10 | NM | 10,164.10 | NM | NM |
| 16 B / providerSetup | 57.374 | 434.31 | NM | 298.48 | NM | NM |
| 16 B / requestPreparation | 0.750 | 0.895 | NM | N/A | NM | NM |
| 65536 B / cleanup | 79.562 | 800.04 | NM | 301.31 | NM | NM |
| 65536 B / compileValidate | 97,477.15 | 159,077.81 | NM | 67,833.23 | NM | NM |
| 65536 B / engineSetup | 0.541 | 0.250 | NM | 219.65 | NM | NM |
| 65536 B / executorSetup | N/A | N/A | NM | 56.520 | NM | NM |
| 65536 B / exportLookup | 1.145 | 1.375 | NM | 7.938 | NM | NM |
| 65536 B / firstCall | 22,997.52 | 30,084.48 | NM | 20,290.29 | NM | NM |
| 65536 B / instantiate | 831.19 | 984.15 | NM | 232.46 | NM | NM |
| 65536 B / load | 89.145 | 94.166 | NM | 80.626 | NM | NM |
| 65536 B / processStartup | 10,007.62 | 17,224.52 | NM | 9,089.60 | NM | NM |
| 65536 B / providerSetup | 70.709 | 433.46 | NM | 284.10 | NM | NM |
| 65536 B / requestPreparation | 11.938 | 9.562 | NM | N/A | NM | NM |

## Native controls (different contracts)

Unchecked C kernels measure the scalar recurrence directly. The native Go control measures the external HTTP reply contract. These are overhead controls with different contracts. The table reports their measured timing directly. The general-call C control runs twice as many iterations of one recurrence; The optimized Wasm combines two recurrence updates into one affine operation.

| Workload | Unchecked native C, ns/iteration |
| --- | ---: |
| original22-v1/core/integer-loop | 2.009 |
| original22-v1/core/function-calls | 0.991 |
| original22-v1/core/memory-load-store | 0.364 |
| original22-v1/memory/narrow-unaligned | 0.620 |
| original22-v1/memory/dependent-address | 1.920 |
| original22-v1/memory/float64-unaligned | 0.892 |
| original22-v1/runtime/simd-memory-full | 4.434 |
| original22-v1/runtime/simd-memory-widen | 1.354 |
| original22-v1/runtime/simd-memory-lanes | 5.794 |
| original22-v1/applications/sha256-clang | 234.26 |
| original22-v1/applications/sha256-rust | 223.38 |
| generic-calls-v1/core/generic-calls | 2.656 |

| Payload | Native Go CPU, us/request | Native Go requests/s |
| --- | ---: | ---: |
| 16 B | 3.379 | 848,994.63 |
| 65536 B | 81.345 | 24,448.82 |

## WasmEdge sustained AOT failures

Full-size WasmEdge0.17.1 O3 AOT preflight freshly reproduces SIGSEGV in sustained struct access, allocation and indirect access. Small original cases pass. Only those exact three cells remain failures in the timing matrix. The original investigation and the observed cumulative stack growth are retained in the [PR #200 report](performance-pr200-complete-comparison-2026-11-17.md). All table timings come from the current event.

## Mechanism evidence

A separate Go runtime trace for 2,048 completed direct requests records 8,403 scheduler events on main and 160 with direct readiness. The queued trace records 16,882 and 4,282 events. Event counts describe these trace runs. The timing cohorts do not use the profiler. A separate Go-only 16 B direct cohort records 7.45% fewer retired instructions than the base, with both instruction controls resolved. Its CPU speed point is +26.5% against base, with unresolved CPU controls. The primary four-worker P3 data remain separate. The saturated-I/O regression succeeds on both backends and fails with ErrLimit when using the base HTTP source.

## Resolved adverse primary intervals

No source comparison has a resolved adverse interval in this event.

## Reproducible evidence

- [Machine-readable results, all controls and source contrasts](data/p3-http-readiness-2026-11-21.json).
- [All 5,508 primary process records](data/p3-http-readiness-2026-11-21-raw.jsonl.gz), SHA-256 `372befc89d86366ff4f78ea8948e56747701073467b647f40713ef40bf8d2337`.
- [Frozen plans, commands, hashes, analysis/export scripts, semantic checks and failure evidence](data/p3-http-readiness-2026-11-21-protocol.tar.gz), SHA-256 `b162ed5aa855db87af01c810eb8617c97f2df558aa3bb1011d85e8fc497c94c6`.
- Per-request vectors are represented by count, extrema and checksum. Tail latency remains unqualified. Public load records omit unrelated desktop command arguments. The separately archived screening observations are excluded from primary ratios.
