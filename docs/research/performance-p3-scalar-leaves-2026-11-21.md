# P3 scalar scheduler leaves and synchronous host transactions, 2026-11-21

Core scalar leaves update serialized host metadata inside the active frame. Snapshot-based synchronous component calls and ready-source reads reuse private transaction and result storage. Scalar load descriptors encode the access width, sign extension and result mask. Memory accesses validate the exclusive endpoint and both address carries before slicing. Scalar callable classification runs within the host path. The changes preserve public result ownership, resource accounting, cancellation and the suspendable call path.

- P3 16 B, GO. CPU 64.602 → 63.286 µs/request and throughput 61,354.95 → 62,613.51 requests/s. Paired CPU comparison against source base 0.983 [0.970, 0.996]; Δ -1.69%; speed +1.72% [+0.35%, +3.14%]; pass. Against Wasmtime 2.112 [1.968, 2.242]; Δ +111.19%; speed -52.65% [-55.40%, -49.19%]; unresolved.
- P3 16 B, LLGO. CPU 188.50 → 186.19 µs/request and throughput 25,601.18 → 26,608.55 requests/s. Paired CPU comparison against source base 0.994 [0.983, 1.005]; Δ -0.63%; speed +0.64% [-0.48%, +1.72%]; pass. Against Wasmtime 6.217 [5.798, 6.585]; Δ +521.66%; speed -83.91% [-84.81%, -82.75%]; unresolved.
- P3 65,536 B, GO. CPU 89.259 → 90.210 µs/request and throughput 43,976.17 → 43,881.66 requests/s. Paired CPU comparison against source base 1.001 [0.992, 1.011]; Δ +0.14%; speed -0.14% [-1.07%, +0.78%]; pass. Against Wasmtime 2.127 [1.986, 2.283]; Δ +112.73%; speed -52.99% [-56.19%, -49.65%]; unresolved.
- P3 65,536 B, LLGO. CPU 225.63 → 223.15 µs/request and throughput 21,158.72 → 21,609.08 requests/s. Paired CPU comparison against source base 0.995 [0.985, 1.004]; Δ -0.53%; speed +0.54% [-0.37%, +1.55%]; unresolved. Against Wasmtime 5.367 [4.877, 5.932]; Δ +436.66%; speed -81.37% [-83.14%, -79.49%]; unresolved.

All 33 core workloads and two complete P3 payloads are measured on the same pinned toolchains. Tables give preparation and steady execution independently. The performance acceptance gate remains open.

## Percentage differences

Cells show `absolute observation (signed speed Δ%)`. Positive speed change means faster. Native execution and P3 use **Wasmtime** as a fixed reference; interpreter execution uses **wazero**, with **WasmEdge** explicitly named for GC rows where wazero rejects the feature. Using a fixed reference keeps the percentages comparable across revisions. The fastest observed runtime varies by workload.

For time, speed Δ = `(reference / candidate - 1) * 100%`; for throughput, speed Δ = `(candidate / reference - 1) * 100%`. Thus 100 ns to 50 ns is **−50% time**, **+100% speed**. P3 CPU percentages mean work per CPU time, separate from wall-clock throughput. Instruction/allocation/byte counters show quantity Δ rather than speed. Source comparisons retain paired ratio intervals and report quantity Δ plus speed Δ where applicable; reciprocal interval endpoints are reversed.

Competitive cell percentages are descriptive quotients of the displayed medians. Paired block estimates below can differ. `†` on these cells means either the candidate or the named reference has unresolved A/A controls; qualified conclusions require resolved controls. Phase boundaries with different contracts have separate absolute observations. The percentages apply to the frozen event and measured revision.

## Revisions, profiles and protocol

- Engine source: base `57e67995` (merged PR #204), measured head `731b1120`. Reused binaries retain their producer revisions. The build records verify production-source equivalence for each reused binary.
- Go **1.27.3**; LLGo **3c6b33ec**, **`-O3 -lto=full`**, LLVM **22.1.8**, default production GC. Source base/head share the compiler profile.
- Separate SDK validation executes affected suites with LLGo **19a6779f**. The separate harness build and checksum smoke apply to the previous source candidate. The primary matrix retains its common pinned SDK.
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
| original22-v1/core/call-control | 3,563.00 (-59.64% speed) † | 2,875.00 (-49.98% speed) † | 666.50 (+115.75% speed) † | 1,438.00 (ref) † | 4,083.00 (-64.78% speed) † | 4,771.00 (-69.86% speed) † |
| original22-v1/core/integer-loop | 1.598 (+0.21% speed) † | 1.622 (-1.27% speed) † | 1.730 (-7.47% speed) † | 1.601 (ref) † | 1.627 (-1.59% speed) † | 1.571 (+1.89% speed) † |
| original22-v1/core/function-calls | 0.802 (+13.31% speed) † | 0.788 (+15.37% speed) † | 1.430 (-36.41% speed) † | 0.909 (ref) | 0.786 (+15.69% speed) † | 0.786 (+15.64% speed) † |
| original22-v1/core/memory-load-store | 0.447 (+7.83% speed) † | 0.438 (+10.05% speed) † | 0.874 (-44.82% speed) † | 0.482 (ref) † | 0.536 (-9.95% speed) † | 0.286 (+68.57% speed) † |
| original22-v1/memory/narrow-unaligned | 0.731 (+4.17% speed) † | 0.742 (+2.64% speed) † | 1.132 (-32.73% speed) † | 0.762 (ref) † | 0.809 (-5.81% speed) † | 0.571 (+33.33% speed) † |
| original22-v1/memory/dependent-address | 1.894 (+9.89% speed) † | 1.861 (+11.84% speed) † | 3.173 (-34.41% speed) † | 2.081 (ref) † | 2.124 (-2.01% speed) † | 1.446 (+43.94% speed) † |
| original22-v1/memory/float64-unaligned | 0.815 (-5.89% speed) † | 0.791 (-2.94% speed) † | 0.897 (-14.44% speed) † | 0.767 (ref) † | 0.772 (-0.56% speed) † | 0.710 (+8.07% speed) † |
| original22-v1/runtime/simd-memory-full | 3.384 (+7.04% speed) † | 3.364 (+7.67% speed) | 3.884 (-6.74% speed) † | 3.622 (ref) | 3.457 (+4.77% speed) | 3.605 (+0.49% speed) |
| original22-v1/runtime/simd-memory-widen | 1.055 (-0.75% speed) | 1.046 (+0.15% speed) † | 1.143 (-8.33% speed) † | 1.047 (ref) | 1.045 (+0.19% speed) | 1.044 (+0.28% speed) |
| original22-v1/runtime/simd-memory-lanes | 5.100 (+15.40% speed) † | 5.094 (+15.54% speed) † | 6.129 (-3.96% speed) † | 5.886 (ref) † | 5.888 (-0.04% speed) † | 5.063 (+16.26% speed) † |
| original22-v1/gc/i31-roundtrip | 2.050 (+4.07% speed) † | 1.790 (+19.21% speed) † | US | 2.133 (ref) † | US | 1.829 (+16.62% speed) † |
| original22-v1/gc/struct-access | 2.413 (-31.95% speed) † | 2.033 (-19.27% speed) † | US | 1.642 (ref) † | US | 10.342 (-84.13% speed) † |
| original22-v1/gc/struct-allocation | 28.926 (+144.10% speed) † | 28.142 (+150.91% speed) † | US | 70.610 (ref) † | US | 39.276 (+79.78% speed) † |
| original22-v1/gc/struct-indirect-access | 3.727 (-13.47% speed) † | 2.629 (+22.66% speed) † | US | 3.225 (ref) † | US | 28.788 (-88.80% speed) † |
| original22-v1/threads/private-memory | 1.350 (+6.64% speed) † | 0.850 (+69.36% speed) † | 1.563 (-7.87% speed) † | 1.440 (ref) † | 1.740 (-17.25% speed) † | 1.679 (-14.27% speed) † |
| original22-v1/threads/shared-memory | 2.554 (-41.93% speed) † | 1.910 (-22.35% speed) † | 1.710 (-13.28% speed) † | 1.483 (ref) † | 1.706 (-13.06% speed) † | 1.704 (-12.96% speed) † |
| original22-v1/threads/atomic-rmw | 3.379 (-30.21% speed) † | 2.735 (-13.79% speed) † | 2.581 (-8.64% speed) † | 2.358 (ref) † | 4.650 (-49.28% speed) † | 2.567 (-8.12% speed) † |
| original22-v1/threads/notify-empty | 1.673 (+328.02% speed) † | 1.387 (+416.08% speed) † | 21.888 (-67.29% speed) † | 7.160 (ref) | 5.183 (+38.14% speed) † | 9.471 (-24.39% speed) † |
| original22-v1/threads/wait-mismatch | 2.902 (+236.11% speed) † | 2.190 (+345.48% speed) † | 27.233 (-64.18% speed) † | 9.754 (ref) † | 3.352 (+190.99% speed) † | 4.040 (+141.47% speed) † |
| original22-v1/threads/wait-timeout | 2.808 (+1464.13% speed) † | 1.869 (+2250.61% speed) † | 307.84 (-85.73% speed) † | 43.927 (ref) † | 174.42 (-74.82% speed) † | 62.219 (-29.40% speed) † |
| original22-v1/applications/sha256-clang | 205.36 (+1.06% speed) † | 204.83 (+1.32% speed) † | 296.57 (-30.02% speed) † | 207.54 (ref) † | 213.62 (-2.85% speed) † | 189.53 (+9.50% speed) † |
| original22-v1/applications/sha256-rust | 209.95 (-11.26% speed) † | 209.47 (-11.06% speed) † | 322.70 (-42.27% speed) † | 186.30 (ref) † | 188.70 (-1.27% speed) † | 182.80 (+1.92% speed) † |
| gc-sustained-v1/gc/i31-roundtrip | 1.376 (+32.87% speed) † | 1.367 (+33.77% speed) † | US | 1.829 (ref) † | US | 1.322 (+38.36% speed) † |
| gc-sustained-v1/gc/struct-access | 1.077 (-0.38% speed) † | 1.072 (+0.02% speed) † | US | 1.073 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 32.791 (+54.04% speed) † | 33.589 (+50.37% speed) † | US | 50.510 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1.813 (+56.23% speed) † | 1.745 (+62.31% speed) † | US | 2.832 (ref) † | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.438 (+146.67% speed) † | 0.349 (+209.66% speed) † | 1.204 (-10.22% speed) † | 1.081 (ref) † | 1.114 (-2.95% speed) † | 1.113 (-2.81% speed) † |
| threads-sustained-v2/threads/shared-memory | 1.409 (-23.42% speed) † | 1.308 (-17.46% speed) † | 1.223 (-11.75% speed) † | 1.079 (ref) † | 1.114 (-3.12% speed) † | 1.115 (-3.18% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 2.079 (-8.45% speed) † | 1.919 (-0.81% speed) † | 2.061 (-7.68% speed) † | 1.903 (ref) † | 3.960 (-51.94% speed) † | 1.906 (-0.15% speed) † |
| threads-sustained-v2/threads/notify-empty | 1.086 (+530.74% speed) † | 0.973 (+603.96% speed) † | 21.307 (-67.86% speed) | 6.848 (ref) | 4.681 (+46.29% speed) † | 8.989 (-23.82% speed) |
| threads-sustained-v2/threads/wait-mismatch | 1.230 (+622.83% speed) † | 1.121 (+693.21% speed) † | 24.476 (-63.67% speed) † | 8.892 (ref) † | 2.418 (+267.77% speed) † | 3.207 (+177.27% speed) † |
| threads-sustained-v2/threads/wait-timeout | 1.224 (+3799.47% speed) † | 1.054 (+4426.76% speed) † | 331.42 (-85.60% speed) † | 47.720 (ref) † | 181.24 (-73.67% speed) † | 66.283 (-28.01% speed) † |
| generic-calls-v1/core/generic-calls | 1.052 (+103.49% speed) † | 1.044 (+105.05% speed) † | 2.151 (-0.45% speed) † | 2.141 (ref) † | 2.189 (-2.21% speed) † | 1.046 (+104.77% speed) † |

## Interpreter execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

The reference runtime is **wazero**. GC rows use **WasmEdge**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2,645.50 (-52.75% speed) † | 2,458.50 (-49.16% speed) † | 1,250.00 (ref) † | NM | NM | 3,458.50 (-63.86% speed) † |
| original22-v1/core/integer-loop | 33.098 (+161.09% speed) † | 27.753 (+211.38% speed) † | 86.416 (ref) † | NM | NM | 51.724 (+67.07% speed) † |
| original22-v1/core/function-calls | 48.847 (+86.58% speed) † | 46.102 (+97.69% speed) † | 91.138 (ref) † | NM | NM | 63.044 (+44.56% speed) † |
| original22-v1/core/memory-load-store | 67.573 (+119.47% speed) † | 57.937 (+155.97% speed) † | 148.30 (ref) † | NM | NM | 93.606 (+58.43% speed) † |
| original22-v1/memory/narrow-unaligned | 96.276 (+118.80% speed) † | 81.173 (+159.52% speed) † | 210.66 (ref) † | NM | NM | 131.59 (+60.09% speed) † |
| original22-v1/memory/dependent-address | 41.265 (+143.98% speed) † | 35.281 (+185.36% speed) † | 100.68 (ref) † | NM | NM | 55.156 (+82.53% speed) † |
| original22-v1/memory/float64-unaligned | 73.271 (+99.82% speed) † | 64.021 (+128.69% speed) † | 146.41 (ref) † | NM | NM | 89.849 (+62.95% speed) † |
| original22-v1/runtime/simd-memory-full | 76.980 (+98.05% speed) † | 67.616 (+125.47% speed) † | 152.45 (ref) † | NM | NM | 79.750 (+91.17% speed) † |
| original22-v1/runtime/simd-memory-widen | 63.069 (+130.37% speed) † | 55.253 (+162.96% speed) † | 145.29 (ref) † | NM | NM | 61.920 (+134.65% speed) † |
| original22-v1/runtime/simd-memory-lanes | 97.133 (+95.79% speed) † | 87.663 (+116.94% speed) † | 190.18 (ref) † | NM | NM | 91.701 (+107.39% speed) † |
| original22-v1/gc/i31-roundtrip (vs WasmEdge) | 36.763 (+22.71% speed) † | 33.085 (+36.35% speed) † | US | NM | NM | 45.110 (ref) † |
| original22-v1/gc/struct-access (vs WasmEdge) | 46.262 (+27.21% speed) † | 42.475 (+38.56% speed) † | US | NM | NM | 58.852 (ref) † |
| original22-v1/gc/struct-allocation (vs WasmEdge) | 93.017 (+17.68% speed) † | 101.89 (+7.43% speed) † | US | NM | NM | 109.46 (ref) † |
| original22-v1/gc/struct-indirect-access (vs WasmEdge) | 82.817 (+24.89% speed) † | 82.127 (+25.94% speed) † | US | NM | NM | 103.43 (ref) † |
| original22-v1/threads/private-memory | 39.985 (+57.64% speed) † | 34.183 (+84.39% speed) † | 63.031 (ref) † | NM | NM | 39.396 (+59.99% speed) † |
| original22-v1/threads/shared-memory | 78.004 (-17.68% speed) † | 33.577 (+91.23% speed) † | 64.210 (ref) † | NM | NM | 39.660 (+61.90% speed) † |
| original22-v1/threads/atomic-rmw | 55.502 (-10.34% speed) † | 33.896 (+46.81% speed) † | 49.763 (ref) † | NM | NM | 33.104 (+50.32% speed) † |
| original22-v1/threads/notify-empty | 32.010 (+35.30% speed) † | 28.185 (+53.66% speed) † | 43.310 (ref) † | NM | NM | 40.552 (+6.80% speed) † |
| original22-v1/threads/wait-mismatch | 59.677 (+20.86% speed) † | 40.002 (+80.30% speed) † | 72.125 (ref) † | NM | NM | 45.977 (+56.87% speed) † |
| original22-v1/threads/wait-timeout | 58.948 (+488.54% speed) † | 37.779 (+818.31% speed) † | 346.93 (ref) † | NM | NM | 98.998 (+250.44% speed) † |
| original22-v1/applications/sha256-clang | 19,394.49 (+119.45% speed) † | 16,293.77 (+161.21% speed) † | 42,560.62 (ref) † | NM | NM | 27,535.03 (+54.57% speed) † |
| original22-v1/applications/sha256-rust | 17,867.20 (+132.83% speed) † | 15,763.63 (+163.91% speed) † | 41,601.03 (ref) † | NM | NM | 24,459.38 (+70.08% speed) † |
| gc-sustained-v1/gc/i31-roundtrip (vs WasmEdge) | 35.977 (+24.96% speed) † | 32.261 (+39.35% speed) † | US | NM | NM | 44.956 (ref) † |
| gc-sustained-v1/gc/struct-access (vs WasmEdge) | 42.228 (+33.46% speed) † | 40.999 (+37.46% speed) † | US | NM | NM | 56.357 (ref) |
| gc-sustained-v1/gc/struct-allocation (vs WasmEdge) | 94.278 (+12.39% speed) † | 107.01 (-0.98% speed) † | US | NM | NM | 105.96 (ref) † |
| gc-sustained-v1/gc/struct-indirect-access (vs WasmEdge) | 82.759 (+30.88% speed) † | 78.408 (+38.14% speed) † | US | NM | NM | 108.31 (ref) † |
| threads-sustained-v2/threads/private-memory | 38.491 (+65.05% speed) † | 33.316 (+90.68% speed) † | 63.529 (ref) † | NM | NM | 38.716 (+64.09% speed) † |
| threads-sustained-v2/threads/shared-memory | 75.978 (-16.05% speed) † | 33.451 (+90.69% speed) † | 63.786 (ref) † | NM | NM | 38.752 (+64.60% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 52.052 (-6.50% speed) † | 32.914 (+47.87% speed) † | 48.671 (ref) † | NM | NM | 32.062 (+51.80% speed) † |
| threads-sustained-v2/threads/notify-empty | 30.615 (+31.35% speed) † | 27.819 (+44.55% speed) † | 40.213 (ref) | NM | NM | 39.834 (+0.95% speed) |
| threads-sustained-v2/threads/wait-mismatch | 56.287 (+24.52% speed) † | 36.788 (+90.52% speed) † | 70.087 (ref) † | NM | NM | 43.148 (+62.43% speed) † |
| threads-sustained-v2/threads/wait-timeout | 60.318 (+544.19% speed) † | 37.969 (+923.35% speed) † | 388.56 (ref) † | NM | NM | 106.46 (+264.98% speed) † |
| generic-calls-v1/core/generic-calls | 108.29 (+78.61% speed) † | 97.034 (+99.32% speed) † | 193.41 (ref) † | NM | NM | 125.23 (+54.44% speed) † |

### Warm speed against the fastest observed competitor

Each row selects the fastest measured competitor median for that backend and workload. The observed percentages remain descriptive when the selected reference or candidate has unresolved controls.

#### Native

| Workload | Fastest observed competitor | wasmvm Go speed Δ | wasmvm LLGo speed Δ |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-compiler | -81.29% † | -76.82% † |
| original22-v1/core/integer-loop | wasmedge-compiler | -1.65% † | -3.10% † |
| original22-v1/core/function-calls | wasmer-compiler | -2.05% † | -0.28% † |
| original22-v1/core/memory-load-store | wasmedge-compiler | -36.03% † | -34.71% † |
| original22-v1/memory/narrow-unaligned | wasmedge-compiler | -21.87% † | -23.02% † |
| original22-v1/memory/dependent-address | wasmedge-compiler | -23.65% † | -22.30% † |
| original22-v1/memory/float64-unaligned | wasmedge-compiler | -12.92% † | -10.19% † |
| original22-v1/runtime/simd-memory-full | wasmer-compiler | +2.17% † | +2.77% |
| original22-v1/runtime/simd-memory-widen | wasmedge-compiler | -1.03% | -0.14% † |
| original22-v1/runtime/simd-memory-lanes | wasmedge-compiler | -0.74% | -0.62% † |
| original22-v1/gc/i31-roundtrip | wasmedge-compiler | -10.77% † | +2.22% † |
| original22-v1/gc/struct-access | wasmtime-compiler | -31.95% † | -19.27% † |
| original22-v1/gc/struct-allocation | wasmedge-compiler | +35.78% † | +39.56% † |
| original22-v1/gc/struct-indirect-access | wasmtime-compiler | -13.47% † | +22.66% † |
| original22-v1/threads/private-memory | wasmtime-compiler | +6.64% † | +69.36% † |
| original22-v1/threads/shared-memory | wasmtime-compiler | -41.93% † | -22.35% † |
| original22-v1/threads/atomic-rmw | wasmtime-compiler | -30.21% † | -13.79% † |
| original22-v1/threads/notify-empty | wasmer-compiler | +209.84% † | +273.59% † |
| original22-v1/threads/wait-mismatch | wasmer-compiler | +15.51% † | +53.09% † |
| original22-v1/threads/wait-timeout | wasmtime-compiler | +1464.13% † | +2250.61% † |
| original22-v1/applications/sha256-clang | wasmedge-compiler | -7.71% † | -7.47% † |
| original22-v1/applications/sha256-rust | wasmedge-compiler | -12.93% † | -12.73% † |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-compiler | -3.97% † | -3.32% † |
| gc-sustained-v1/gc/struct-access | wasmtime-compiler | -0.38% † | +0.02% † |
| gc-sustained-v1/gc/struct-allocation | wasmtime-compiler | +54.04% † | +50.37% † |
| gc-sustained-v1/gc/struct-indirect-access | wasmtime-compiler | +56.23% † | +62.31% † |
| threads-sustained-v2/threads/private-memory | wasmtime-compiler | +146.67% † | +209.66% † |
| threads-sustained-v2/threads/shared-memory | wasmtime-compiler | -23.42% † | -17.46% † |
| threads-sustained-v2/threads/atomic-rmw | wasmtime-compiler | -8.45% † | -0.81% † |
| threads-sustained-v2/threads/notify-empty | wasmer-compiler | +331.16% † | +381.22% † |
| threads-sustained-v2/threads/wait-mismatch | wasmer-compiler | +96.54% † | +115.68% † |
| threads-sustained-v2/threads/wait-timeout | wasmtime-compiler | +3799.47% † | +4426.76% † |
| generic-calls-v1/core/generic-calls | wasmedge-compiler | -0.63% † | +0.14% † |

#### Interpreter

| Workload | Fastest observed competitor | wasmvm Go speed Δ | wasmvm LLGo speed Δ |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-interpreter | -52.75% † | -49.16% † |
| original22-v1/core/integer-loop | wasmedge-interpreter | +56.28% † | +86.38% † |
| original22-v1/core/function-calls | wasmedge-interpreter | +29.06% | +36.75% |
| original22-v1/core/memory-load-store | wasmedge-interpreter | +38.53% † | +61.56% † |
| original22-v1/memory/narrow-unaligned | wasmedge-interpreter | +36.68% † | +62.11% † |
| original22-v1/memory/dependent-address | wasmedge-interpreter | +33.66% † | +56.33% † |
| original22-v1/memory/float64-unaligned | wasmedge-interpreter | +22.63% † | +40.34% † |
| original22-v1/runtime/simd-memory-full | wasmedge-interpreter | +3.60% † | +17.95% † |
| original22-v1/runtime/simd-memory-widen | wasmedge-interpreter | -1.82% † | +12.06% † |
| original22-v1/runtime/simd-memory-lanes | wasmedge-interpreter | -5.59% † | +4.61% |
| original22-v1/gc/i31-roundtrip | wasmedge-interpreter | +22.71% † | +36.35% † |
| original22-v1/gc/struct-access | wasmedge-interpreter | +27.21% † | +38.56% † |
| original22-v1/gc/struct-allocation | wasmedge-interpreter | +17.68% † | +7.43% † |
| original22-v1/gc/struct-indirect-access | wasmedge-interpreter | +24.89% † | +25.94% † |
| original22-v1/threads/private-memory | wasmedge-interpreter | -1.47% † | +15.25% † |
| original22-v1/threads/shared-memory | wasmedge-interpreter | -49.16% † | +18.12% † |
| original22-v1/threads/atomic-rmw | wasmedge-interpreter | -40.35% † | -2.34% † |
| original22-v1/threads/notify-empty | wasmedge-interpreter | +26.68% † | +43.88% † |
| original22-v1/threads/wait-mismatch | wasmedge-interpreter | -22.96% † | +14.94% † |
| original22-v1/threads/wait-timeout | wasmedge-interpreter | +67.94% † | +162.04% † |
| original22-v1/applications/sha256-clang | wasmedge-interpreter | +41.97% † | +68.99% † |
| original22-v1/applications/sha256-rust | wasmedge-interpreter | +36.90% † | +55.16% † |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-interpreter | +24.96% † | +39.35% † |
| gc-sustained-v1/gc/struct-access | wasmedge-interpreter | +33.46% † | +37.46% † |
| gc-sustained-v1/gc/struct-allocation | wasmedge-interpreter | +12.39% † | -0.98% † |
| gc-sustained-v1/gc/struct-indirect-access | wasmedge-interpreter | +30.88% † | +38.14% † |
| threads-sustained-v2/threads/private-memory | wasmedge-interpreter | +0.58% † | +16.21% † |
| threads-sustained-v2/threads/shared-memory | wasmedge-interpreter | -49.00% † | +15.85% † |
| threads-sustained-v2/threads/atomic-rmw | wasmedge-interpreter | -38.40% † | -2.59% † |
| threads-sustained-v2/threads/notify-empty | wasmedge-interpreter | +30.12% † | +43.19% † |
| threads-sustained-v2/threads/wait-mismatch | wasmedge-interpreter | -23.34% † | +17.29% † |
| threads-sustained-v2/threads/wait-timeout | wasmedge-interpreter | +76.50% † | +180.39% † |
| generic-calls-v1/core/generic-calls | wasmedge-interpreter | +15.65% † | +29.06% † |

## Complete P3 HTTP requests

P3 competitor cells marked `NM` have no measured adapter for the pinned component/WASI workload. Wasmtime uses the same validated workload. Managed allocation counts exclude foreign Rust/C/JIT allocations and cover only the managed share of allocation.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU, us/request | 63.286 (-54.36% speed) † | 186.19 (-84.49% speed) † | NM | 28.884 (ref) † | NM | NM |
| 16 B instructions/request | 684,784.79 (+221.82% Δ) | 1,270,961.23 (+497.30% Δ) | NM | 212,785.75 (ref) | NM | NM |
| 16 B requests/s | 62,613.51 (-10.39% speed) † | 26,608.55 (-61.92% speed) † | NM | 69,876.94 (ref) † | NM | NM |
| 65536 B CPU, us/request | 90.210 (-51.35% speed) † | 223.15 (-80.33% speed) † | NM | 43.890 (ref) † | NM | NM |
| 65536 B instructions/request | 969,773.66 (+212.33% Δ) † | 1,620,064.97 (+421.76% Δ) † | NM | 310,501.05 (ref) † | NM | NM |
| 65536 B requests/s | 43,881.66 (-24.31% speed) † | 21,609.08 (-62.73% speed) † | NM | 57,977.00 (ref) † | NM | NM |

### P3 source contrasts

| Metric | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| 16 B cpuUs | 0.983 [0.970, 0.996]; Δ -1.69%; speed +1.72% [+0.35%, +3.14%]; pass | 0.994 [0.983, 1.005]; Δ -0.63%; speed +0.64% [-0.48%, +1.72%]; pass |
| 16 B instructions | 0.969 [0.968, 0.970]; Δ -3.12%; pass | 0.977 [0.971, 0.982]; Δ -2.30%; pass |
| 16 B throughput | 1.020 [1.004, 1.035]; Δ +1.97%; speed +1.97% [+0.40%, +3.51%]; pass | 1.027 [1.007, 1.044]; Δ +2.69%; speed +2.69% [+0.66%, +4.40%]; unresolved |
| 16 B allocsPerRequest | 0.979 [0.979, 0.979]; Δ -2.12%; pass | 0.982 [0.982, 0.982]; Δ -1.85%; pass |
| 16 B bytesPerRequest | 0.964 [0.964, 0.964]; Δ -3.56%; pass | 0.967 [0.967, 0.967]; Δ -3.26%; pass |
| 65536 B cpuUs | 1.001 [0.992, 1.011]; Δ +0.14%; speed -0.14% [-1.07%, +0.78%]; pass | 0.995 [0.985, 1.004]; Δ -0.53%; speed +0.54% [-0.37%, +1.55%]; unresolved |
| 65536 B instructions | 0.981 [0.979, 0.982]; Δ -1.92%; pass | 0.984 [0.973, 0.993]; Δ -1.64%; pass |
| 65536 B throughput | 1.005 [0.996, 1.015]; Δ +0.50%; speed +0.50% [-0.42%, +1.46%]; pass | 1.016 [1.004, 1.026]; Δ +1.57%; speed +1.57% [+0.40%, +2.56%]; unresolved |
| 65536 B allocsPerRequest | 0.980 [0.980, 0.980]; Δ -1.96%; pass | 0.983 [0.983, 0.983]; Δ -1.65%; pass |
| 65536 B bytesPerRequest | 0.986 [0.986, 0.986]; Δ -1.37%; pass | 0.986 [0.986, 0.986]; Δ -1.39%; pass |

Throughput ratios above one improve. CPU, instruction and allocation ratios below one improve. The JSON retains all own-baseline observations and every per-role A/A interval.

### Competitive P3 ratios

| Payload / compiler | CPU ratio to Wasmtime [95% CI]; controls | Throughput ratio to Wasmtime [95% CI]; controls |
| --- | ---: | ---: |
| 16 B go | 2.112 [1.968, 2.242]; Δ +111.19%; speed -52.65% [-55.40%, -49.19%]; unresolved | 0.910 [0.883, 0.940]; Δ -9.05%; speed -9.05% [-11.70%, -5.95%]; unresolved |
| 16 B llgo | 6.217 [5.798, 6.585]; Δ +521.66%; speed -83.91% [-84.81%, -82.75%]; unresolved | 0.384 [0.378, 0.393]; Δ -61.59%; speed -61.59% [-62.19%, -60.72%]; unresolved |
| 65536 B go | 2.127 [1.986, 2.283]; Δ +112.73%; speed -52.99% [-56.19%, -49.65%]; unresolved | 0.762 [0.746, 0.780]; Δ -23.78%; speed -23.78% [-25.43%, -22.02%]; unresolved |
| 65536 B llgo | 5.367 [4.877, 5.932]; Δ +436.66%; speed -81.37% [-83.14%, -79.49%]; unresolved | 0.371 [0.365, 0.377]; Δ -62.92%; speed -62.92% [-63.54%, -62.29%]; unresolved |

## Core source baseline

Head/base ratios below one improve. These marginal intervals describe this event; an apparent change in an unchanged mechanism requires diagnosis before being attributed to source changes. The machine-readable report includes every competitor ratio and A/A interval.

### native

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 0.901 [0.834, 0.981]; Δ -9.94%; speed +11.04% [+1.96%, +19.85%]; unresolved | 0.912 [0.809, 1.007]; Δ -8.82%; speed +9.67% [-0.69%, +23.60%]; unresolved |
| original22-v1/core/integer-loop | 1.036 [0.991, 1.110]; Δ +3.64%; speed -3.52% [-9.88%, +0.89%]; unresolved | 0.993 [0.948, 1.034]; Δ -0.71%; speed +0.71% [-3.28%, +5.49%]; unresolved |
| original22-v1/core/function-calls | 1.011 [0.996, 1.037]; Δ +1.14%; speed -1.12% [-3.61%, +0.37%]; unresolved | 0.990 [0.964, 1.006]; Δ -0.98%; speed +0.99% [-0.55%, +3.76%]; unresolved |
| original22-v1/core/memory-load-store | 1.018 [0.957, 1.112]; Δ +1.78%; speed -1.75% [-10.07%, +4.55%]; unresolved | 1.155 [0.996, 1.358]; Δ +15.46%; speed -13.39% [-26.34%, +0.44%]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.981 [0.923, 1.024]; Δ -1.89%; speed +1.93% [-2.33%, +8.37%]; unresolved | 1.010 [0.820, 1.227]; Δ +0.96%; speed -0.95% [-18.53%, +21.93%]; unresolved |
| original22-v1/memory/dependent-address | 0.978 [0.921, 1.018]; Δ -2.17%; speed +2.22% [-1.75%, +8.61%]; unresolved | 1.008 [0.992, 1.033]; Δ +0.80%; speed -0.80% [-3.22%, +0.82%]; unresolved |
| original22-v1/memory/float64-unaligned | 1.010 [0.987, 1.040]; Δ +0.99%; speed -0.98% [-3.87%, +1.34%]; pass | 0.986 [0.967, 1.004]; Δ -1.36%; speed +1.38% [-0.40%, +3.41%]; unresolved |
| original22-v1/runtime/simd-memory-full | 0.997 [0.992, 1.002]; Δ -0.29%; speed +0.29% [-0.20%, +0.82%]; unresolved | 0.995 [0.988, 1.002]; Δ -0.50%; speed +0.51% [-0.17%, +1.25%]; pass |
| original22-v1/runtime/simd-memory-widen | 1.015 [0.993, 1.042]; Δ +1.46%; speed -1.44% [-4.02%, +0.67%]; unresolved | 0.995 [0.983, 1.004]; Δ -0.54%; speed +0.54% [-0.42%, +1.74%]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 0.996 [0.989, 1.003]; Δ -0.41%; speed +0.41% [-0.31%, +1.06%]; unresolved | 0.997 [0.985, 1.007]; Δ -0.34%; speed +0.34% [-0.68%, +1.55%]; unresolved |
| original22-v1/gc/i31-roundtrip | 0.965 [0.939, 0.992]; Δ -3.51%; speed +3.64% [+0.82%, +6.44%]; unresolved | 1.015 [0.997, 1.036]; Δ +1.46%; speed -1.44% [-3.49%, +0.31%]; unresolved |
| original22-v1/gc/struct-access | 1.013 [0.939, 1.104]; Δ +1.34%; speed -1.32% [-9.46%, +6.53%]; unresolved | 1.032 [0.961, 1.121]; Δ +3.25%; speed -3.15% [-10.76%, +4.01%]; unresolved |
| original22-v1/gc/struct-allocation | 0.995 [0.942, 1.043]; Δ -0.48%; speed +0.48% [-4.17%, +6.11%]; unresolved | 0.959 [0.906, 1.006]; Δ -4.11%; speed +4.28% [-0.63%, +10.38%]; unresolved |
| original22-v1/gc/struct-indirect-access | 1.149 [0.798, 1.715]; Δ +14.93%; speed -12.99% [-41.67%, +25.35%]; unresolved | 0.836 [0.684, 1.020]; Δ -16.41%; speed +19.63% [-2.00%, +46.20%]; unresolved |
| original22-v1/threads/private-memory | 1.013 [0.874, 1.156]; Δ +1.25%; speed -1.24% [-13.49%, +14.45%]; unresolved | 0.987 [0.933, 1.058]; Δ -1.32%; speed +1.34% [-5.50%, +7.23%]; unresolved |
| original22-v1/threads/shared-memory | 0.959 [0.821, 1.115]; Δ -4.07%; speed +4.25% [-10.31%, +21.78%]; unresolved | 1.013 [0.951, 1.085]; Δ +1.33%; speed -1.32% [-7.87%, +5.11%]; unresolved |
| original22-v1/threads/atomic-rmw | 1.053 [0.986, 1.142]; Δ +5.30%; speed -5.03% [-12.43%, +1.46%]; unresolved | 1.004 [0.979, 1.027]; Δ +0.41%; speed -0.41% [-2.65%, +2.18%]; unresolved |
| original22-v1/threads/notify-empty | 0.945 [0.890, 1.017]; Δ -5.52%; speed +5.84% [-1.70%, +12.32%]; unresolved | 0.966 [0.872, 1.036]; Δ -3.44%; speed +3.56% [-3.46%, +14.72%]; unresolved |
| original22-v1/threads/wait-mismatch | 1.059 [0.968, 1.142]; Δ +5.94%; speed -5.61% [-12.41%, +3.31%]; unresolved | 0.940 [0.832, 1.035]; Δ -5.96%; speed +6.33% [-3.38%, +20.13%]; unresolved |
| original22-v1/threads/wait-timeout | 0.999 [0.794, 1.173]; Δ -0.10%; speed +0.10% [-14.75%, +25.93%]; unresolved | 1.034 [0.906, 1.170]; Δ +3.40%; speed -3.29% [-14.53%, +10.34%]; unresolved |
| original22-v1/applications/sha256-clang | 0.993 [0.982, 1.004]; Δ -0.66%; speed +0.67% [-0.39%, +1.79%]; unresolved | 0.999 [0.988, 1.011]; Δ -0.06%; speed +0.06% [-1.04%, +1.25%]; unresolved |
| original22-v1/applications/sha256-rust | 0.989 [0.969, 1.005]; Δ -1.10%; speed +1.11% [-0.52%, +3.22%]; unresolved | 0.993 [0.984, 1.000]; Δ -0.74%; speed +0.75% [-0.04%, +1.60%]; unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 1.022 [0.991, 1.071]; Δ +2.17%; speed -2.12% [-6.66%, +0.91%]; unresolved | 0.962 [0.926, 0.999]; Δ -3.79%; speed +3.94% [+0.12%, +8.01%]; unresolved |
| gc-sustained-v1/gc/struct-access | 0.993 [0.978, 1.008]; Δ -0.72%; speed +0.73% [-0.78%, +2.25%]; unresolved | 0.983 [0.924, 1.024]; Δ -1.66%; speed +1.69% [-2.31%, +8.24%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 1.006 [0.991, 1.022]; Δ +0.58%; speed -0.58% [-2.12%, +0.86%]; unresolved | 0.919 [0.895, 0.948]; Δ -8.09%; speed +8.80% [+5.46%, +11.77%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.041 [0.982, 1.109]; Δ +4.08%; speed -3.92% [-9.79%, +1.83%]; unresolved | 1.014 [0.984, 1.040]; Δ +1.38%; speed -1.36% [-3.84%, +1.60%]; unresolved |
| threads-sustained-v2/threads/private-memory | 1.003 [0.953, 1.052]; Δ +0.30%; speed -0.29% [-4.94%, +4.92%]; unresolved | 1.007 [0.991, 1.022]; Δ +0.73%; speed -0.73% [-2.16%, +0.88%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 1.026 [0.978, 1.102]; Δ +2.56%; speed -2.49% [-9.29%, +2.26%]; unresolved | 1.037 [1.003, 1.080]; Δ +3.72%; speed -3.59% [-7.44%, -0.27%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 1.006 [0.985, 1.026]; Δ +0.57%; speed -0.57% [-2.57%, +1.57%]; unresolved | 0.997 [0.990, 1.004]; Δ -0.30%; speed +0.30% [-0.43%, +1.01%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 1.013 [0.991, 1.033]; Δ +1.27%; speed -1.26% [-3.23%, +0.88%]; unresolved | 1.006 [0.998, 1.014]; Δ +0.57%; speed -0.56% [-1.38%, +0.22%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.089 [0.961, 1.322]; Δ +8.95%; speed -8.21% [-24.37%, +4.06%]; unresolved | 1.036 [0.994, 1.086]; Δ +3.60%; speed -3.48% [-7.92%, +0.58%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 1.079 [0.985, 1.245]; Δ +7.85%; speed -7.28% [-19.70%, +1.49%]; unresolved | 0.999 [0.966, 1.031]; Δ -0.06%; speed +0.06% [-3.01%, +3.53%]; unresolved |
| generic-calls-v1/core/generic-calls | 1.001 [0.979, 1.024]; Δ +0.14%; speed -0.14% [-2.32%, +2.12%]; unresolved | 1.006 [0.996, 1.016]; Δ +0.59%; speed -0.59% [-1.58%, +0.40%]; unresolved |

### interpreter

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 1.040 [0.922, 1.135]; Δ +3.96%; speed -3.81% [-11.93%, +8.42%]; unresolved | 1.065 [1.012, 1.144]; Δ +6.46%; speed -6.07% [-12.56%, -1.16%]; unresolved |
| original22-v1/core/integer-loop | 0.990 [0.959, 1.018]; Δ -1.01%; speed +1.02% [-1.74%, +4.28%]; unresolved | 1.011 [0.979, 1.040]; Δ +1.13%; speed -1.11% [-3.87%, +2.12%]; unresolved |
| original22-v1/core/function-calls | 0.990 [0.975, 1.001]; Δ -1.01%; speed +1.02% [-0.11%, +2.57%]; pass | 0.997 [0.981, 1.010]; Δ -0.26%; speed +0.26% [-0.94%, +1.93%]; pass |
| original22-v1/core/memory-load-store | 1.030 [0.996, 1.068]; Δ +2.97%; speed -2.88% [-6.38%, +0.36%]; unresolved | 0.979 [0.972, 0.986]; Δ -2.09%; speed +2.14% [+1.45%, +2.92%]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.991 [0.963, 1.031]; Δ -0.89%; speed +0.89% [-3.05%, +3.89%]; unresolved | 0.954 [0.933, 0.969]; Δ -4.65%; speed +4.88% [+3.22%, +7.15%]; unresolved |
| original22-v1/memory/dependent-address | 0.986 [0.974, 1.001]; Δ -1.43%; speed +1.45% [-0.05%, +2.72%]; unresolved | 0.967 [0.950, 0.987]; Δ -3.29%; speed +3.40% [+1.31%, +5.28%]; unresolved |
| original22-v1/memory/float64-unaligned | 0.997 [0.982, 1.020]; Δ -0.28%; speed +0.28% [-2.00%, +1.86%]; unresolved | 0.967 [0.957, 0.977]; Δ -3.27%; speed +3.38% [+2.38%, +4.48%]; unresolved |
| original22-v1/runtime/simd-memory-full | 1.009 [0.996, 1.023]; Δ +0.91%; speed -0.91% [-2.29%, +0.43%]; unresolved | 1.002 [0.964, 1.037]; Δ +0.24%; speed -0.24% [-3.55%, +3.70%]; unresolved |
| original22-v1/runtime/simd-memory-widen | 0.987 [0.972, 0.998]; Δ -1.34%; speed +1.36% [+0.19%, +2.89%]; unresolved | 0.954 [0.923, 0.980]; Δ -4.64%; speed +4.87% [+2.01%, +8.33%]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 1.020 [1.008, 1.034]; Δ +2.03%; speed -1.99% [-3.32%, -0.84%]; unresolved | 1.005 [0.993, 1.019]; Δ +0.46%; speed -0.46% [-1.89%, +0.70%]; unresolved |
| original22-v1/gc/i31-roundtrip | 1.008 [0.981, 1.036]; Δ +0.81%; speed -0.81% [-3.47%, +1.97%]; unresolved | 0.985 [0.950, 1.016]; Δ -1.46%; speed +1.48% [-1.55%, +5.28%]; unresolved |
| original22-v1/gc/struct-access | 1.033 [1.009, 1.059]; Δ +3.32%; speed -3.21% [-5.55%, -0.86%]; unresolved | 1.034 [1.007, 1.061]; Δ +3.43%; speed -3.32% [-5.78%, -0.69%]; unresolved |
| original22-v1/gc/struct-allocation | 1.013 [1.008, 1.018]; Δ +1.28%; speed -1.26% [-1.75%, -0.81%]; unresolved | 0.842 [0.825, 0.865]; Δ -15.85%; speed +18.84% [+15.54%, +21.28%]; unresolved |
| original22-v1/gc/struct-indirect-access | 0.983 [0.928, 1.015]; Δ -1.67%; speed +1.70% [-1.50%, +7.81%]; unresolved | 0.964 [0.879, 1.033]; Δ -3.61%; speed +3.74% [-3.22%, +13.79%]; unresolved |
| original22-v1/threads/private-memory | 0.988 [0.974, 1.003]; Δ -1.18%; speed +1.19% [-0.28%, +2.62%]; unresolved | 0.925 [0.841, 0.982]; Δ -7.49%; speed +8.10% [+1.79%, +18.98%]; unresolved |
| original22-v1/threads/shared-memory | 0.980 [0.891, 1.044]; Δ -1.98%; speed +2.01% [-4.22%, +12.27%]; unresolved | 0.941 [0.878, 0.987]; Δ -5.92%; speed +6.29% [+1.34%, +13.92%]; unresolved |
| original22-v1/threads/atomic-rmw | 0.999 [0.982, 1.014]; Δ -0.15%; speed +0.15% [-1.34%, +1.86%]; unresolved | 0.979 [0.963, 0.993]; Δ -2.15%; speed +2.20% [+0.71%, +3.87%]; unresolved |
| original22-v1/threads/notify-empty | 1.014 [1.004, 1.025]; Δ +1.41%; speed -1.39% [-2.41%, -0.42%]; unresolved | 0.994 [0.982, 1.007]; Δ -0.56%; speed +0.57% [-0.72%, +1.82%]; unresolved |
| original22-v1/threads/wait-mismatch | 0.993 [0.979, 1.009]; Δ -0.73%; speed +0.74% [-0.86%, +2.11%]; unresolved | 0.986 [0.930, 1.029]; Δ -1.37%; speed +1.39% [-2.79%, +7.50%]; unresolved |
| original22-v1/threads/wait-timeout | 0.998 [0.881, 1.130]; Δ -0.17%; speed +0.17% [-11.48%, +13.46%]; unresolved | 0.998 [0.934, 1.046]; Δ -0.20%; speed +0.20% [-4.42%, +7.10%]; unresolved |
| original22-v1/applications/sha256-clang | 0.988 [0.978, 0.998]; Δ -1.18%; speed +1.19% [+0.24%, +2.20%]; unresolved | 0.991 [0.968, 1.019]; Δ -0.90%; speed +0.91% [-1.82%, +3.28%]; unresolved |
| original22-v1/applications/sha256-rust | 0.978 [0.967, 0.988]; Δ -2.16%; speed +2.21% [+1.20%, +3.46%]; unresolved | 0.981 [0.960, 0.996]; Δ -1.93%; speed +1.97% [+0.39%, +4.20%]; pass |
| gc-sustained-v1/gc/i31-roundtrip | 0.999 [0.989, 1.012]; Δ -0.06%; speed +0.06% [-1.14%, +1.15%]; unresolved | 1.016 [0.994, 1.040]; Δ +1.63%; speed -1.61% [-3.87%, +0.62%]; unresolved |
| gc-sustained-v1/gc/struct-access | 1.016 [0.988, 1.058]; Δ +1.55%; speed -1.53% [-5.46%, +1.19%]; unresolved | 1.024 [1.006, 1.039]; Δ +2.35%; speed -2.30% [-3.71%, -0.62%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 0.988 [0.977, 0.999]; Δ -1.23%; speed +1.24% [+0.07%, +2.39%]; unresolved | 0.828 [0.794, 0.856]; Δ -17.22%; speed +20.80% [+16.86%, +25.99%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.002 [0.981, 1.023]; Δ +0.22%; speed -0.22% [-2.26%, +1.96%]; unresolved | 1.032 [1.013, 1.051]; Δ +3.19%; speed -3.09% [-4.85%, -1.31%]; unresolved |
| threads-sustained-v2/threads/private-memory | 0.987 [0.979, 0.996]; Δ -1.27%; speed +1.29% [+0.41%, +2.17%]; unresolved | 0.974 [0.965, 0.982]; Δ -2.62%; speed +2.69% [+1.87%, +3.67%]; pass |
| threads-sustained-v2/threads/shared-memory | 1.080 [1.020, 1.198]; Δ +8.00%; speed -7.41% [-16.53%, -1.95%]; unresolved | 0.960 [0.950, 0.971]; Δ -3.99%; speed +4.15% [+3.01%, +5.31%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.993 [0.986, 0.999]; Δ -0.74%; speed +0.74% [+0.11%, +1.43%]; unresolved | 0.989 [0.975, 1.002]; Δ -1.06%; speed +1.07% [-0.20%, +2.56%]; pass |
| threads-sustained-v2/threads/notify-empty | 1.014 [0.996, 1.040]; Δ +1.43%; speed -1.41% [-3.83%, +0.40%]; unresolved | 1.001 [0.984, 1.020]; Δ +0.10%; speed -0.10% [-1.92%, +1.68%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.027 [1.000, 1.076]; Δ +2.75%; speed -2.67% [-7.04%, +0.04%]; unresolved | 1.040 [0.987, 1.116]; Δ +3.97%; speed -3.82% [-10.40%, +1.29%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 0.999 [0.960, 1.025]; Δ -0.06%; speed +0.06% [-2.47%, +4.13%]; unresolved | 0.993 [0.967, 1.019]; Δ -0.74%; speed +0.74% [-1.88%, +3.43%]; unresolved |
| generic-calls-v1/core/generic-calls | 0.949 [0.887, 0.988]; Δ -5.13%; speed +5.41% [+1.22%, +12.72%]; unresolved | 0.969 [0.910, 1.006]; Δ -3.12%; speed +3.22% [-0.61%, +9.88%]; unresolved |

## Preparation, first call and cleanup

All entries below are **microseconds per process or invocation**, including unresolved observations. Phase intervals are not inferred from warm-call controls. Startup is independently observed process readiness. Native compile/validate boundaries differ by API; WasmEdge AOT includes validation and artifact I/O, plus separately recorded native-image loading. Every process creates a fresh native image. First-call and warm execution have their own boundaries; values are never subtracted to invent pure costs.

<details>
<summary>Native: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,480.46 | 6,376.69 | 4,808.02 | 4,795.50 | 6,239.27 | 19,024.79 |
| original22-v1/core/integer-loop | 4,372.98 | 6,486.10 | 4,559.77 | 4,533.04 | 6,019.46 | 16,503.38 |
| original22-v1/core/function-calls | 4,286.44 | 6,086.83 | 4,197.92 | 4,191.73 | 5,930.15 | 16,565.42 |
| original22-v1/core/memory-load-store | 5,207.56 | 6,150.92 | 4,403.02 | 4,375.96 | 6,176.69 | 16,681.08 |
| original22-v1/memory/narrow-unaligned | 4,627.83 | 5,928.21 | 4,518.00 | 4,742.35 | 6,430.19 | 16,229.92 |
| original22-v1/memory/dependent-address | 4,627.10 | 7,561.94 | 4,754.92 | 4,329.12 | 6,120.79 | 16,872.58 |
| original22-v1/memory/float64-unaligned | 4,980.50 | 6,410.60 | 4,461.85 | 4,421.02 | 7,174.96 | 16,783.92 |
| original22-v1/runtime/simd-memory-full | 4,237.60 | 5,648.42 | 4,456.17 | 4,233.58 | 6,008.81 | 15,945.88 |
| original22-v1/runtime/simd-memory-widen | 4,018.33 | 5,862.54 | 4,204.52 | 4,225.48 | 5,871.73 | 16,289.85 |
| original22-v1/runtime/simd-memory-lanes | 4,187.42 | 6,110.90 | 4,281.85 | 4,359.94 | 6,112.58 | 16,421.48 |
| original22-v1/gc/i31-roundtrip | 3,812.21 | 5,735.17 | US | 4,127.33 | US | 15,963.12 |
| original22-v1/gc/struct-access | 4,443.92 | 6,588.44 | US | 4,382.54 | US | 16,891.10 |
| original22-v1/gc/struct-allocation | 4,107.40 | 5,956.33 | US | 4,280.73 | US | 16,227.19 |
| original22-v1/gc/struct-indirect-access | 4,153.69 | 5,844.00 | US | 4,163.71 | US | 16,768.71 |
| original22-v1/threads/private-memory | 3,753.54 | 5,998.38 | 4,393.40 | 4,114.50 | 5,953.15 | 16,201.12 |
| original22-v1/threads/shared-memory | 4,315.79 | 5,942.77 | 3,983.15 | 4,052.29 | 5,856.60 | 16,043.52 |
| original22-v1/threads/atomic-rmw | 3,940.25 | 5,905.71 | 4,370.50 | 4,214.38 | 5,912.00 | 16,666.00 |
| original22-v1/threads/notify-empty | 3,845.46 | 5,853.98 | 3,937.08 | 3,995.06 | 5,848.69 | 15,659.46 |
| original22-v1/threads/wait-mismatch | 4,204.27 | 6,617.42 | 4,594.00 | 5,466.50 | 6,712.42 | 17,067.40 |
| original22-v1/threads/wait-timeout | 4,211.10 | 6,088.00 | 4,522.08 | 4,696.17 | 6,163.08 | 17,536.29 |
| original22-v1/applications/sha256-clang | 4,303.67 | 5,947.87 | 4,104.44 | 4,479.12 | 5,956.19 | 16,447.56 |
| original22-v1/applications/sha256-rust | 4,324.19 | 6,149.04 | 4,316.48 | 4,403.12 | 6,198.96 | 16,190.54 |
| gc-sustained-v1/gc/i31-roundtrip | 4,244.35 | 6,313.96 | US | 4,358.46 | US | 16,409.23 |
| gc-sustained-v1/gc/struct-access | 4,352.10 | 6,134.98 | US | 4,285.83 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 4,589.62 | 6,108.44 | US | 4,236.08 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 4,270.17 | 6,027.17 | US | 4,584.06 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 4,587.77 | 5,987.15 | 4,806.42 | 4,217.85 | 6,070.69 | 16,749.83 |
| threads-sustained-v2/threads/shared-memory | 4,210.81 | 5,987.65 | 4,170.02 | 4,317.81 | 6,095.25 | 16,333.44 |
| threads-sustained-v2/threads/atomic-rmw | 4,217.65 | 5,716.87 | 4,182.98 | 4,115.04 | 6,205.83 | 16,288.13 |
| threads-sustained-v2/threads/notify-empty | 4,210.88 | 6,159.60 | 4,181.02 | 4,340.37 | 5,905.31 | 16,221.12 |
| threads-sustained-v2/threads/wait-mismatch | 4,234.40 | 7,034.54 | 4,461.13 | 4,438.46 | 6,451.46 | 16,349.08 |
| threads-sustained-v2/threads/wait-timeout | 4,644.02 | 6,926.81 | 4,529.42 | 4,357.35 | 6,371.31 | 16,831.19 |
| generic-calls-v1/core/generic-calls | 4,177.60 | 5,904.04 | 4,466.10 | 4,759.06 | 5,942.83 | 16,184.58 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 41.500 | 46.209 | 38.062 | 32.084 | 43.020 | 48.167 |
| original22-v1/core/integer-loop | 37.416 | 39.001 | 41.083 | 35.603 | 31.458 | 41.355 |
| original22-v1/core/function-calls | 36.688 | 46.624 | 38.334 | 32.395 | 39.541 | 43.958 |
| original22-v1/core/memory-load-store | 42.688 | 39.833 | 40.834 | 33.729 | 36.562 | 43.791 |
| original22-v1/memory/narrow-unaligned | 44.812 | 46.541 | 40.958 | 39.416 | 38.749 | 41.354 |
| original22-v1/memory/dependent-address | 44.021 | 44.374 | 44.625 | 40.499 | 45.458 | 42.000 |
| original22-v1/memory/float64-unaligned | 37.355 | 45.041 | 41.021 | 39.834 | 65.166 | 40.209 |
| original22-v1/runtime/simd-memory-full | 38.208 | 38.334 | 36.041 | 32.103 | 36.229 | 39.834 |
| original22-v1/runtime/simd-memory-widen | 38.270 | 39.209 | 37.812 | 33.208 | 32.500 | 46.833 |
| original22-v1/runtime/simd-memory-lanes | 38.291 | 41.854 | 36.104 | 30.855 | 34.312 | 43.667 |
| original22-v1/gc/i31-roundtrip | 31.750 | 34.770 | US | 25.062 | US | 37.312 |
| original22-v1/gc/struct-access | 39.021 | 42.541 | US | 35.999 | US | 40.188 |
| original22-v1/gc/struct-allocation | 31.541 | 34.084 | US | 29.458 | US | 33.562 |
| original22-v1/gc/struct-indirect-access | 33.541 | 41.895 | US | 28.458 | US | 40.749 |
| original22-v1/threads/private-memory | 32.791 | 32.708 | 36.291 | 33.166 | 29.333 | 42.208 |
| original22-v1/threads/shared-memory | 36.104 | 33.896 | 34.730 | 30.270 | 30.938 | 40.292 |
| original22-v1/threads/atomic-rmw | 40.416 | 37.146 | 35.938 | 27.209 | 33.229 | 38.188 |
| original22-v1/threads/notify-empty | 29.042 | 34.354 | 30.375 | 29.084 | 25.500 | 40.292 |
| original22-v1/threads/wait-mismatch | 40.938 | 45.520 | 36.625 | 67.813 | 41.959 | 46.437 |
| original22-v1/threads/wait-timeout | 44.854 | 44.458 | 45.980 | 37.584 | 31.500 | 38.896 |
| original22-v1/applications/sha256-clang | 40.145 | 34.688 | 42.104 | 46.104 | 42.250 | 39.605 |
| original22-v1/applications/sha256-rust | 43.146 | 35.395 | 41.354 | 31.354 | 35.230 | 40.188 |
| gc-sustained-v1/gc/i31-roundtrip | 38.666 | 39.104 | US | 30.687 | US | 37.146 |
| gc-sustained-v1/gc/struct-access | 35.937 | 39.145 | US | 33.251 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 36.521 | 40.500 | US | 37.146 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 39.166 | 46.854 | US | 35.355 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 37.916 | 44.001 | 39.562 | 30.271 | 29.312 | 41.250 |
| threads-sustained-v2/threads/shared-memory | 43.480 | 38.000 | 38.833 | 28.542 | 32.312 | 39.292 |
| threads-sustained-v2/threads/atomic-rmw | 34.355 | 32.979 | 36.605 | 31.562 | 31.208 | 35.125 |
| threads-sustained-v2/threads/notify-empty | 34.292 | 38.145 | 32.666 | 26.312 | 33.208 | 40.792 |
| threads-sustained-v2/threads/wait-mismatch | 42.375 | 41.709 | 41.750 | 30.834 | 42.146 | 46.729 |
| threads-sustained-v2/threads/wait-timeout | 38.395 | 49.479 | 51.230 | 34.291 | 40.688 | 57.083 |
| generic-calls-v1/core/generic-calls | 42.562 | 41.812 | 41.041 | 36.375 | 35.188 | 46.105 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.938 | 5.584 | 60.626 | 134.77 | 46.250 | 29.834 |
| original22-v1/core/integer-loop | 5.625 | 5.542 | 58.792 | 126.81 | 44.229 | 26.167 |
| original22-v1/core/function-calls | 6.125 | 5.500 | 59.459 | 132.90 | 41.417 | 25.312 |
| original22-v1/core/memory-load-store | 6.542 | 5.062 | 59.480 | 121.42 | 43.084 | 28.209 |
| original22-v1/memory/narrow-unaligned | 6.438 | 5.250 | 61.916 | 144.65 | 45.167 | 29.291 |
| original22-v1/memory/dependent-address | 6.312 | 5.458 | 57.646 | 128.92 | 44.542 | 29.125 |
| original22-v1/memory/float64-unaligned | 6.125 | 5.396 | 95.854 | 143.25 | 74.666 | 32.375 |
| original22-v1/runtime/simd-memory-full | 5.792 | 4.896 | 55.834 | 121.40 | 39.750 | 25.541 |
| original22-v1/runtime/simd-memory-widen | 5.958 | 4.812 | 56.083 | 120.35 | 40.958 | 26.791 |
| original22-v1/runtime/simd-memory-lanes | 6.521 | 4.896 | 55.416 | 125.88 | 40.896 | 27.584 |
| original22-v1/gc/i31-roundtrip | 5.688 | 4.604 | US | 122.79 | US | 25.750 |
| original22-v1/gc/struct-access | 6.000 | 5.146 | US | 130.27 | US | 26.709 |
| original22-v1/gc/struct-allocation | 5.583 | 4.750 | US | 127.73 | US | 25.541 |
| original22-v1/gc/struct-indirect-access | 6.146 | 5.333 | US | 133.06 | US | 28.105 |
| original22-v1/threads/private-memory | 5.750 | 4.687 | 58.999 | 121.58 | 40.688 | 27.334 |
| original22-v1/threads/shared-memory | 6.438 | 5.000 | 56.917 | 125.44 | 42.083 | 29.312 |
| original22-v1/threads/atomic-rmw | 6.208 | 5.083 | 57.812 | 128.52 | 43.791 | 29.062 |
| original22-v1/threads/notify-empty | 5.750 | 4.854 | 57.916 | 124.62 | 38.209 | 28.667 |
| original22-v1/threads/wait-mismatch | 6.188 | 5.479 | 65.145 | 136.06 | 46.834 | 32.375 |
| original22-v1/threads/wait-timeout | 6.729 | 5.375 | 64.438 | 137.15 | 41.396 | 30.000 |
| original22-v1/applications/sha256-clang | 6.312 | 4.979 | 59.584 | 134.21 | 42.854 | 26.895 |
| original22-v1/applications/sha256-rust | 6.688 | 4.875 | 59.230 | 129.81 | 40.354 | 27.125 |
| gc-sustained-v1/gc/i31-roundtrip | 5.729 | 5.312 | US | 125.33 | US | 27.730 |
| gc-sustained-v1/gc/struct-access | 5.646 | 5.375 | US | 124.52 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 5.938 | 4.521 | US | 120.92 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 6.250 | 5.479 | US | 129.31 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 6.313 | 5.312 | 56.292 | 129.19 | 41.541 | 30.521 |
| threads-sustained-v2/threads/shared-memory | 6.354 | 4.708 | 61.062 | 124.96 | 41.521 | 25.646 |
| threads-sustained-v2/threads/atomic-rmw | 6.021 | 4.729 | 61.166 | 122.13 | 41.646 | 25.584 |
| threads-sustained-v2/threads/notify-empty | 5.521 | 4.792 | 59.354 | 122.17 | 42.916 | 25.896 |
| threads-sustained-v2/threads/wait-mismatch | 6.500 | 5.979 | 60.105 | 127.06 | 42.979 | 27.542 |
| threads-sustained-v2/threads/wait-timeout | 6.958 | 5.000 | 61.334 | 122.23 | 46.499 | 31.209 |
| generic-calls-v1/core/generic-calls | 6.104 | 5.021 | 62.271 | 128.44 | 41.583 | 26.709 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 206.37 | 160.44 | 100.29 | 855.12 | 1,096.52 | 286,750.27 |
| original22-v1/core/integer-loop | 241.15 | 209.27 | 126.29 | 756.48 | 1,156.90 | 275,238.85 |
| original22-v1/core/function-calls | 307.52 | 257.83 | 132.21 | 767.79 | 1,249.56 | 281,569.85 |
| original22-v1/core/memory-load-store | 336.67 | 291.15 | 139.04 | 722.40 | 1,224.08 | 284,949.50 |
| original22-v1/memory/narrow-unaligned | 311.42 | 287.60 | 154.75 | 889.54 | 1,308.71 | 278,682.56 |
| original22-v1/memory/dependent-address | 300.31 | 273.15 | 162.44 | 797.60 | 1,214.90 | 282,388.21 |
| original22-v1/memory/float64-unaligned | 332.19 | 297.15 | 201.67 | 964.94 | 2,010.92 | 275,229.00 |
| original22-v1/runtime/simd-memory-full | 277.25 | 253.48 | 130.69 | 1,019.54 | 1,174.46 | 272,927.23 |
| original22-v1/runtime/simd-memory-widen | 264.50 | 234.96 | 126.67 | 1,020.44 | 1,142.08 | 278,813.52 |
| original22-v1/runtime/simd-memory-lanes | 293.69 | 255.12 | 127.06 | 989.42 | 1,209.69 | 273,522.83 |
| original22-v1/gc/i31-roundtrip | 203.50 | 158.17 | US | 721.58 | US | 273,836.52 |
| original22-v1/gc/struct-access | 245.85 | 213.48 | US | 958.56 | US | 289,449.27 |
| original22-v1/gc/struct-allocation | 226.92 | 187.54 | US | 1,136.23 | US | 278,053.62 |
| original22-v1/gc/struct-indirect-access | 287.10 | 249.50 | US | 1,626.31 | US | 283,526.83 |
| original22-v1/threads/private-memory | 229.17 | 187.58 | 125.67 | 753.50 | 1,118.62 | 279,303.94 |
| original22-v1/threads/shared-memory | 247.69 | 197.23 | 135.52 | 754.83 | 1,137.50 | 278,300.73 |
| original22-v1/threads/atomic-rmw | 219.50 | 184.29 | 129.67 | 763.96 | 1,253.52 | 277,400.15 |
| original22-v1/threads/notify-empty | 218.92 | 180.02 | 129.12 | 867.48 | 1,144.27 | 281,334.33 |
| original22-v1/threads/wait-mismatch | 232.15 | 199.12 | 137.79 | 1,383.52 | 1,164.83 | 277,539.71 |
| original22-v1/threads/wait-timeout | 240.38 | 192.23 | 155.73 | 1,025.19 | 1,201.60 | 278,378.00 |
| original22-v1/applications/sha256-clang | 1,395.17 | 1,749.83 | 302.00 | 1,429.94 | 1,791.92 | 284,814.79 |
| original22-v1/applications/sha256-rust | 1,402.88 | 1,736.75 | 340.98 | 1,457.62 | 2,023.83 | 282,253.94 |
| gc-sustained-v1/gc/i31-roundtrip | 202.92 | 168.65 | US | 761.73 | US | 278,793.71 |
| gc-sustained-v1/gc/struct-access | 241.65 | 205.58 | US | 915.15 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 227.77 | 186.81 | US | 1,121.69 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 302.44 | 255.10 | US | 1,625.58 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 237.19 | 192.81 | 127.06 | 772.27 | 1,176.58 | 276,878.08 |
| threads-sustained-v2/threads/shared-memory | 244.35 | 190.60 | 135.90 | 795.94 | 1,161.88 | 275,713.08 |
| threads-sustained-v2/threads/atomic-rmw | 211.04 | 174.23 | 125.04 | 769.35 | 1,193.60 | 278,786.50 |
| threads-sustained-v2/threads/notify-empty | 214.98 | 176.81 | 131.48 | 900.10 | 1,252.15 | 277,894.56 |
| threads-sustained-v2/threads/wait-mismatch | 229.10 | 188.48 | 138.65 | 942.65 | 1,351.35 | 273,997.25 |
| threads-sustained-v2/threads/wait-timeout | 251.96 | 206.73 | 135.15 | 991.04 | 1,239.65 | 285,416.83 |
| generic-calls-v1/core/generic-calls | 315.33 | 269.96 | 130.44 | 977.23 | 1,731.71 | 279,490.81 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 9.271 | 8.459 | 14.812 | 8.791 | 18.604 | 18.375 |
| original22-v1/core/integer-loop | 8.062 | 8.250 | 14.396 | 7.854 | 17.041 | 15.333 |
| original22-v1/core/function-calls | 8.896 | 8.042 | 13.813 | 6.979 | 17.125 | 20.416 |
| original22-v1/core/memory-load-store | 13.479 | 13.771 | 16.875 | 21.312 | 31.958 | 34.500 |
| original22-v1/memory/narrow-unaligned | 14.459 | 13.626 | 20.521 | 22.250 | 32.499 | 34.958 |
| original22-v1/memory/dependent-address | 13.854 | 13.959 | 19.250 | 21.980 | 31.521 | 33.834 |
| original22-v1/memory/float64-unaligned | 14.271 | 14.625 | 20.041 | 23.209 | 47.291 | 37.479 |
| original22-v1/runtime/simd-memory-full | 14.562 | 14.521 | 19.146 | 56.041 | 34.625 | 44.334 |
| original22-v1/runtime/simd-memory-widen | 14.729 | 15.021 | 17.271 | 57.688 | 36.334 | 43.854 |
| original22-v1/runtime/simd-memory-lanes | 14.979 | 14.542 | 18.605 | 53.376 | 35.459 | 42.541 |
| original22-v1/gc/i31-roundtrip | 8.292 | 8.250 | US | 7.396 | US | 14.312 |
| original22-v1/gc/struct-access | 9.709 | 9.249 | US | 24.188 | US | 18.209 |
| original22-v1/gc/struct-allocation | 10.062 | 9.292 | US | 53.916 | US | 21.125 |
| original22-v1/gc/struct-indirect-access | 20.562 | 21.458 | US | 67.396 | US | 28.438 |
| original22-v1/threads/private-memory | 12.771 | 12.792 | 18.146 | 20.875 | 31.687 | 35.812 |
| original22-v1/threads/shared-memory | 14.084 | 13.479 | 17.604 | 22.187 | 34.728 | 35.666 |
| original22-v1/threads/atomic-rmw | 14.521 | 14.958 | 16.105 | 22.312 | 34.999 | 36.209 |
| original22-v1/threads/notify-empty | 12.979 | 14.166 | 18.896 | 20.917 | 36.666 | 35.541 |
| original22-v1/threads/wait-mismatch | 14.438 | 15.084 | 19.145 | 23.959 | 37.980 | 39.938 |
| original22-v1/threads/wait-timeout | 15.291 | 15.041 | 19.709 | 22.230 | 38.729 | 38.000 |
| original22-v1/applications/sha256-clang | 18.188 | 17.687 | 16.333 | 57.312 | 36.000 | 43.895 |
| original22-v1/applications/sha256-rust | 17.750 | 18.395 | 19.334 | 52.959 | 36.500 | 42.250 |
| gc-sustained-v1/gc/i31-roundtrip | 8.604 | 9.291 | US | 8.271 | US | 16.480 |
| gc-sustained-v1/gc/struct-access | 9.583 | 8.750 | US | 23.125 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 10.104 | 9.500 | US | 57.042 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 22.354 | 21.188 | US | 67.292 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 14.979 | 14.625 | 16.062 | 22.459 | 31.146 | 38.312 |
| threads-sustained-v2/threads/shared-memory | 13.062 | 13.854 | 17.896 | 22.229 | 36.291 | 34.312 |
| threads-sustained-v2/threads/atomic-rmw | 12.916 | 14.688 | 18.291 | 21.375 | 35.792 | 38.855 |
| threads-sustained-v2/threads/notify-empty | 13.146 | 13.521 | 18.000 | 21.521 | 37.688 | 36.562 |
| threads-sustained-v2/threads/wait-mismatch | 14.626 | 15.917 | 18.667 | 22.541 | 37.978 | 36.271 |
| threads-sustained-v2/threads/wait-timeout | 15.042 | 15.271 | 19.375 | 23.063 | 35.729 | 37.771 |
| generic-calls-v1/core/generic-calls | 10.542 | 9.959 | 15.791 | 8.416 | 19.125 | 16.334 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.250 | 0.438 | 2.146 | 4.250 | 4.771 | 0.292 |
| original22-v1/core/integer-loop | 0.250 | 0.459 | 2.312 | 3.938 | 4.604 | 0.208 |
| original22-v1/core/function-calls | 0.250 | 0.438 | 1.854 | 3.958 | 4.792 | 0.334 |
| original22-v1/core/memory-load-store | 0.271 | 0.458 | 2.083 | 4.062 | 4.688 | 0.230 |
| original22-v1/memory/narrow-unaligned | 0.333 | 0.459 | 2.354 | 4.354 | 5.042 | 0.271 |
| original22-v1/memory/dependent-address | 0.250 | 0.479 | 2.417 | 4.396 | 4.833 | 0.230 |
| original22-v1/memory/float64-unaligned | 0.250 | 0.479 | 2.499 | 4.271 | 6.667 | 0.271 |
| original22-v1/runtime/simd-memory-full | 0.292 | 0.500 | 2.750 | 2.042 | 4.604 | 0.229 |
| original22-v1/runtime/simd-memory-widen | 0.292 | 0.459 | 2.125 | 2.146 | 4.646 | 0.250 |
| original22-v1/runtime/simd-memory-lanes | 0.250 | 0.458 | 2.229 | 1.917 | 4.583 | 0.208 |
| original22-v1/gc/i31-roundtrip | 0.250 | 0.459 | US | 3.979 | US | 0.271 |
| original22-v1/gc/struct-access | 0.250 | 0.416 | US | 4.021 | US | 0.250 |
| original22-v1/gc/struct-allocation | 0.250 | 0.458 | US | 1.792 | US | 0.291 |
| original22-v1/gc/struct-indirect-access | 0.292 | 0.500 | US | 2.042 | US | 0.291 |
| original22-v1/threads/private-memory | 0.250 | 0.521 | 2.333 | 4.083 | 4.708 | 0.271 |
| original22-v1/threads/shared-memory | 0.291 | 0.520 | 2.604 | 4.333 | 4.688 | 0.291 |
| original22-v1/threads/atomic-rmw | 0.271 | 0.479 | 1.834 | 4.354 | 4.938 | 0.292 |
| original22-v1/threads/notify-empty | 0.250 | 0.500 | 1.896 | 4.042 | 4.875 | 0.271 |
| original22-v1/threads/wait-mismatch | 0.292 | 0.438 | 2.896 | 4.708 | 4.979 | 0.250 |
| original22-v1/threads/wait-timeout | 0.292 | 0.500 | 2.729 | 4.333 | 4.854 | 0.271 |
| original22-v1/applications/sha256-clang | 0.250 | 0.438 | 2.417 | 2.104 | 5.104 | 0.291 |
| original22-v1/applications/sha256-rust | 0.250 | 0.459 | 2.438 | 1.958 | 5.167 | 0.209 |
| gc-sustained-v1/gc/i31-roundtrip | 0.291 | 0.459 | US | 4.167 | US | 0.292 |
| gc-sustained-v1/gc/struct-access | 0.208 | 0.375 | US | 4.104 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 0.271 | 0.438 | US | 1.750 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 0.312 | 0.500 | US | 2.021 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.292 | 0.458 | 2.333 | 4.125 | 4.792 | 0.250 |
| threads-sustained-v2/threads/shared-memory | 0.271 | 0.459 | 2.521 | 4.354 | 4.812 | 0.271 |
| threads-sustained-v2/threads/atomic-rmw | 0.250 | 0.500 | 2.417 | 4.229 | 4.833 | 0.271 |
| threads-sustained-v2/threads/notify-empty | 0.271 | 0.458 | 2.042 | 4.208 | 4.646 | 0.209 |
| threads-sustained-v2/threads/wait-mismatch | 0.250 | 0.500 | 2.896 | 4.188 | 4.979 | 0.312 |
| threads-sustained-v2/threads/wait-timeout | 0.313 | 0.500 | 2.624 | 4.375 | 4.937 | 0.208 |
| generic-calls-v1/core/generic-calls | 0.250 | 0.396 | 2.708 | 3.979 | 4.896 | 0.291 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 11.000 | 8.584 | 9.375 | 32.105 | 14.854 | 12.687 |
| original22-v1/core/integer-loop | 1,583.54 | 1,622.94 | 1,573.71 | 1,621.77 | 1,566.50 | 1,576.63 |
| original22-v1/core/function-calls | 794.33 | 787.79 | 1,296.83 | 926.54 | 791.25 | 805.42 |
| original22-v1/core/memory-load-store | 448.06 | 436.27 | 803.92 | 505.50 | 539.54 | 293.23 |
| original22-v1/memory/narrow-unaligned | 675.15 | 667.27 | 1,009.77 | 777.71 | 805.44 | 540.02 |
| original22-v1/memory/dependent-address | 1,892.69 | 1,882.12 | 3,086.73 | 2,187.94 | 2,132.79 | 1,478.73 |
| original22-v1/memory/float64-unaligned | 765.08 | 765.98 | 846.90 | 765.04 | 869.27 | 704.12 |
| original22-v1/runtime/simd-memory-full | 3,399.94 | 3,386.10 | 3,681.73 | 3,602.15 | 3,618.31 | 3,624.96 |
| original22-v1/runtime/simd-memory-widen | 1,045.77 | 1,042.65 | 1,052.81 | 1,041.98 | 1,074.42 | 1,054.81 |
| original22-v1/runtime/simd-memory-lanes | 4,918.98 | 5,007.77 | 5,873.04 | 5,802.77 | 5,726.75 | 5,000.15 |
| original22-v1/gc/i31-roundtrip | 23.729 | 21.375 | US | 46.104 | US | 27.187 |
| original22-v1/gc/struct-access | 30.625 | 30.104 | US | 51.625 | US | 137.88 |
| original22-v1/gc/struct-allocation | 366.19 | 405.42 | US | 143.42 | US | 532.81 |
| original22-v1/gc/struct-indirect-access | 27.959 | 24.979 | US | 30.729 | US | 318.98 |
| original22-v1/threads/private-memory | 15.416 | 11.729 | 20.041 | 45.771 | 27.916 | 31.270 |
| original22-v1/threads/shared-memory | 27.188 | 22.750 | 21.625 | 46.562 | 29.666 | 31.437 |
| original22-v1/threads/atomic-rmw | 32.792 | 28.250 | 27.375 | 52.021 | 54.166 | 37.209 |
| original22-v1/threads/notify-empty | 22.645 | 18.375 | 230.02 | 105.60 | 71.770 | 110.71 |
| original22-v1/threads/wait-mismatch | 25.729 | 20.458 | 283.40 | 141.17 | 42.396 | 55.562 |
| original22-v1/threads/wait-timeout | 24.834 | 19.584 | 2,867.29 | 507.00 | 1,903.71 | 665.02 |
| original22-v1/applications/sha256-clang | 1,161.35 | 1,015.81 | 1,348.27 | 1,045.08 | 962.88 | 1,008.38 |
| original22-v1/applications/sha256-rust | 1,089.71 | 1,087.06 | 1,504.54 | 968.60 | 948.19 | 913.31 |
| gc-sustained-v1/gc/i31-roundtrip | 1,359.21 | 1,378.06 | US | 1,869.79 | US | 1,321.71 |
| gc-sustained-v1/gc/struct-access | 1,115.10 | 1,097.33 | US | 1,099.31 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 25,881.54 | 15,473.60 | US | 3,079.77 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1,799.08 | 1,670.48 | US | 2,673.56 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 39.959 | 36.437 | 112.67 | 136.06 | 119.79 | 126.33 |
| threads-sustained-v2/threads/shared-memory | 140.15 | 130.83 | 115.75 | 137.50 | 119.42 | 123.83 |
| threads-sustained-v2/threads/atomic-rmw | 199.25 | 191.54 | 191.83 | 213.88 | 405.88 | 201.98 |
| threads-sustained-v2/threads/notify-empty | 107.17 | 95.459 | 2,114.02 | 733.77 | 471.00 | 929.38 |
| threads-sustained-v2/threads/wait-mismatch | 119.71 | 114.75 | 2,532.02 | 953.85 | 268.98 | 341.04 |
| threads-sustained-v2/threads/wait-timeout | 165.19 | 106.60 | 29,164.85 | 4,496.00 | 17,867.54 | 6,307.33 |
| generic-calls-v1/core/generic-calls | 5,184.77 | 5,259.54 | 10,787.69 | 10,351.15 | 10,386.67 | 5,318.69 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.854 | 5.021 | 9.001 | 81.375 | 28.562 | 361.67 |
| original22-v1/core/integer-loop | 7.688 | 7.729 | 6.271 | 83.333 | 31.250 | 374.06 |
| original22-v1/core/function-calls | 10.541 | 6.104 | 8.854 | 82.667 | 34.416 | 379.29 |
| original22-v1/core/memory-load-store | 10.250 | 7.458 | 5.208 | 80.625 | 55.812 | 429.06 |
| original22-v1/memory/narrow-unaligned | 10.646 | 14.688 | 6.062 | 89.291 | 60.458 | 430.79 |
| original22-v1/memory/dependent-address | 18.645 | 17.416 | 8.750 | 112.81 | 70.188 | 425.25 |
| original22-v1/memory/float64-unaligned | 10.833 | 12.374 | 5.271 | 86.521 | 63.979 | 440.21 |
| original22-v1/runtime/simd-memory-full | 13.063 | 12.584 | 7.958 | 141.77 | 65.500 | 449.83 |
| original22-v1/runtime/simd-memory-widen | 17.334 | 9.667 | 11.396 | 117.65 | 58.999 | 445.21 |
| original22-v1/runtime/simd-memory-lanes | 9.416 | 10.646 | 9.959 | 106.17 | 60.333 | 444.58 |
| original22-v1/gc/i31-roundtrip | 4.917 | 3.812 | US | 57.209 | US | 351.44 |
| original22-v1/gc/struct-access | 17.041 | 13.000 | US | 93.521 | US | 379.04 |
| original22-v1/gc/struct-allocation | 138.38 | 91.812 | US | 111.25 | US | 1,994.19 |
| original22-v1/gc/struct-indirect-access | 14.312 | 13.500 | US | 109.65 | US | 423.77 |
| original22-v1/threads/private-memory | 6.875 | 7.396 | 7.438 | 82.875 | 53.729 | 406.92 |
| original22-v1/threads/shared-memory | 9.479 | 7.625 | 4.896 | 84.562 | 57.104 | 377.58 |
| original22-v1/threads/atomic-rmw | 10.313 | 6.125 | 3.917 | 78.104 | 53.480 | 400.79 |
| original22-v1/threads/notify-empty | 12.666 | 9.062 | 4.625 | 88.499 | 51.000 | 426.52 |
| original22-v1/threads/wait-mismatch | 16.209 | 9.771 | 6.292 | 93.167 | 57.396 | 423.40 |
| original22-v1/threads/wait-timeout | 14.938 | 8.604 | 4.667 | 95.292 | 61.333 | 425.48 |
| original22-v1/applications/sha256-clang | 15.104 | 15.521 | 13.062 | 168.25 | 65.125 | 447.75 |
| original22-v1/applications/sha256-rust | 10.583 | 11.126 | 5.417 | 138.58 | 67.896 | 438.52 |
| gc-sustained-v1/gc/i31-roundtrip | 9.541 | 8.874 | US | 95.187 | US | 377.77 |
| gc-sustained-v1/gc/struct-access | 14.562 | 14.771 | US | 90.000 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 3,019.69 | 2,034.88 | US | 803.65 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 23.000 | 16.062 | US | 153.73 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 13.375 | 13.041 | 6.167 | 102.58 | 54.771 | 418.88 |
| threads-sustained-v2/threads/shared-memory | 13.875 | 8.562 | 7.291 | 85.042 | 57.666 | 412.48 |
| threads-sustained-v2/threads/atomic-rmw | 15.271 | 9.229 | 7.479 | 94.584 | 57.125 | 424.92 |
| threads-sustained-v2/threads/notify-empty | 10.999 | 7.771 | 4.208 | 81.855 | 54.938 | 387.31 |
| threads-sustained-v2/threads/wait-mismatch | 13.312 | 10.312 | 4.750 | 87.291 | 59.105 | 440.27 |
| threads-sustained-v2/threads/wait-timeout | 17.583 | 13.229 | 3.917 | 86.980 | 62.250 | 441.83 |
| generic-calls-v1/core/generic-calls | 11.459 | 5.333 | 8.042 | 73.104 | 37.541 | 404.12 |

</details>

<details>
<summary>Interpreter: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,466.75 | 7,639.23 | 5,008.42 | NM | NM | 18,189.90 |
| original22-v1/core/integer-loop | 4,275.48 | 6,277.15 | 4,361.65 | NM | NM | 17,700.83 |
| original22-v1/core/function-calls | 4,124.62 | 5,846.15 | 4,432.58 | NM | NM | 16,357.42 |
| original22-v1/core/memory-load-store | 4,256.12 | 6,329.69 | 4,563.08 | NM | NM | 16,800.96 |
| original22-v1/memory/narrow-unaligned | 4,216.81 | 6,373.31 | 4,760.10 | NM | NM | 17,769.88 |
| original22-v1/memory/dependent-address | 4,493.00 | 7,726.67 | 4,419.29 | NM | NM | 16,372.58 |
| original22-v1/memory/float64-unaligned | 4,605.77 | 5,935.02 | 4,413.10 | NM | NM | 16,918.94 |
| original22-v1/runtime/simd-memory-full | 4,240.38 | 6,009.85 | 4,342.31 | NM | NM | 15,769.83 |
| original22-v1/runtime/simd-memory-widen | 4,180.40 | 6,007.77 | 4,133.77 | NM | NM | 16,428.65 |
| original22-v1/runtime/simd-memory-lanes | 4,283.06 | 6,212.65 | 4,528.69 | NM | NM | 16,033.21 |
| original22-v1/gc/i31-roundtrip | 3,888.35 | 5,774.46 | US | NM | NM | 16,062.48 |
| original22-v1/gc/struct-access | 4,209.48 | 6,761.58 | US | NM | NM | 17,270.77 |
| original22-v1/gc/struct-allocation | 4,203.15 | 5,924.02 | US | NM | NM | 16,167.31 |
| original22-v1/gc/struct-indirect-access | 4,114.42 | 6,062.44 | US | NM | NM | 17,028.52 |
| original22-v1/threads/private-memory | 4,002.06 | 6,076.56 | 4,133.12 | NM | NM | 16,762.96 |
| original22-v1/threads/shared-memory | 4,334.15 | 5,679.08 | 3,911.00 | NM | NM | 15,753.54 |
| original22-v1/threads/atomic-rmw | 4,029.65 | 5,759.88 | 4,344.21 | NM | NM | 16,555.44 |
| original22-v1/threads/notify-empty | 4,024.90 | 5,824.35 | 3,865.58 | NM | NM | 16,138.00 |
| original22-v1/threads/wait-mismatch | 4,794.08 | 6,638.98 | 4,718.79 | NM | NM | 17,090.04 |
| original22-v1/threads/wait-timeout | 4,599.46 | 6,333.35 | 4,502.54 | NM | NM | 16,737.06 |
| original22-v1/applications/sha256-clang | 4,126.56 | 6,320.50 | 4,451.35 | NM | NM | 16,767.06 |
| original22-v1/applications/sha256-rust | 4,119.50 | 6,184.00 | 4,475.90 | NM | NM | 16,142.46 |
| gc-sustained-v1/gc/i31-roundtrip | 4,184.23 | 6,172.33 | US | NM | NM | 16,370.10 |
| gc-sustained-v1/gc/struct-access | 4,244.46 | 6,242.44 | US | NM | NM | 16,714.06 |
| gc-sustained-v1/gc/struct-allocation | 4,346.75 | 6,078.04 | US | NM | NM | 16,367.25 |
| gc-sustained-v1/gc/struct-indirect-access | 4,429.98 | 6,256.83 | US | NM | NM | 17,043.85 |
| threads-sustained-v2/threads/private-memory | 4,089.46 | 6,103.92 | 4,493.27 | NM | NM | 16,218.87 |
| threads-sustained-v2/threads/shared-memory | 4,345.69 | 6,039.31 | 4,270.71 | NM | NM | 16,737.96 |
| threads-sustained-v2/threads/atomic-rmw | 4,230.69 | 5,813.33 | 4,349.44 | NM | NM | 16,767.96 |
| threads-sustained-v2/threads/notify-empty | 4,123.12 | 5,823.96 | 4,385.02 | NM | NM | 16,241.90 |
| threads-sustained-v2/threads/wait-mismatch | 4,480.79 | 5,936.23 | 4,332.29 | NM | NM | 15,802.58 |
| threads-sustained-v2/threads/wait-timeout | 5,151.88 | 6,794.85 | 5,095.81 | NM | NM | 17,301.69 |
| generic-calls-v1/core/generic-calls | 4,473.12 | 5,935.37 | 4,551.77 | NM | NM | 15,864.71 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 37.166 | 45.687 | 41.916 | NM | NM | 46.895 |
| original22-v1/core/integer-loop | 39.416 | 43.646 | 42.895 | NM | NM | 57.584 |
| original22-v1/core/function-calls | 38.666 | 39.417 | 40.542 | NM | NM | 42.541 |
| original22-v1/core/memory-load-store | 40.062 | 43.521 | 36.354 | NM | NM | 42.209 |
| original22-v1/memory/narrow-unaligned | 45.001 | 46.041 | 43.230 | NM | NM | 55.812 |
| original22-v1/memory/dependent-address | 45.645 | 51.020 | 40.396 | NM | NM | 52.124 |
| original22-v1/memory/float64-unaligned | 45.208 | 44.271 | 39.520 | NM | NM | 48.355 |
| original22-v1/runtime/simd-memory-full | 37.229 | 40.833 | 35.667 | NM | NM | 41.895 |
| original22-v1/runtime/simd-memory-widen | 34.791 | 40.354 | 35.459 | NM | NM | 44.291 |
| original22-v1/runtime/simd-memory-lanes | 44.250 | 42.938 | 36.812 | NM | NM | 37.666 |
| original22-v1/gc/i31-roundtrip | 30.417 | 34.126 | US | NM | NM | 39.834 |
| original22-v1/gc/struct-access | 36.396 | 40.750 | US | NM | NM | 48.021 |
| original22-v1/gc/struct-allocation | 32.792 | 37.250 | US | NM | NM | 39.105 |
| original22-v1/gc/struct-indirect-access | 32.417 | 44.854 | US | NM | NM | 45.500 |
| original22-v1/threads/private-memory | 32.229 | 36.416 | 35.479 | NM | NM | 38.146 |
| original22-v1/threads/shared-memory | 34.250 | 35.875 | 29.584 | NM | NM | 35.375 |
| original22-v1/threads/atomic-rmw | 32.938 | 36.084 | 39.271 | NM | NM | 38.645 |
| original22-v1/threads/notify-empty | 35.084 | 35.104 | 29.396 | NM | NM | 41.542 |
| original22-v1/threads/wait-mismatch | 38.416 | 42.812 | 38.416 | NM | NM | 41.874 |
| original22-v1/threads/wait-timeout | 41.021 | 50.145 | 45.458 | NM | NM | 49.604 |
| original22-v1/applications/sha256-clang | 33.542 | 41.062 | 43.000 | NM | NM | 43.459 |
| original22-v1/applications/sha256-rust | 39.459 | 40.625 | 42.521 | NM | NM | 40.604 |
| gc-sustained-v1/gc/i31-roundtrip | 34.791 | 35.458 | US | NM | NM | 35.834 |
| gc-sustained-v1/gc/struct-access | 34.396 | 38.959 | US | NM | NM | 45.895 |
| gc-sustained-v1/gc/struct-allocation | 38.062 | 35.479 | US | NM | NM | 38.188 |
| gc-sustained-v1/gc/struct-indirect-access | 35.980 | 39.230 | US | NM | NM | 40.230 |
| threads-sustained-v2/threads/private-memory | 37.105 | 36.562 | 39.938 | NM | NM | 38.771 |
| threads-sustained-v2/threads/shared-memory | 46.145 | 40.958 | 32.416 | NM | NM | 48.688 |
| threads-sustained-v2/threads/atomic-rmw | 30.312 | 35.562 | 37.125 | NM | NM | 45.812 |
| threads-sustained-v2/threads/notify-empty | 34.146 | 36.959 | 30.000 | NM | NM | 40.438 |
| threads-sustained-v2/threads/wait-mismatch | 32.666 | 37.000 | 33.062 | NM | NM | 43.375 |
| threads-sustained-v2/threads/wait-timeout | 45.375 | 49.729 | 45.145 | NM | NM | 47.396 |
| generic-calls-v1/core/generic-calls | 41.062 | 48.521 | 33.021 | NM | NM | 39.041 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4.667 | 5.521 | 2.562 | NM | NM | 29.750 |
| original22-v1/core/integer-loop | 4.375 | 5.062 | 2.708 | NM | NM | 30.062 |
| original22-v1/core/function-calls | 4.375 | 4.708 | 2.188 | NM | NM | 27.146 |
| original22-v1/core/memory-load-store | 4.375 | 4.792 | 2.604 | NM | NM | 26.250 |
| original22-v1/memory/narrow-unaligned | 4.625 | 5.271 | 2.667 | NM | NM | 29.062 |
| original22-v1/memory/dependent-address | 5.333 | 4.896 | 2.979 | NM | NM | 28.209 |
| original22-v1/memory/float64-unaligned | 5.021 | 4.958 | 2.375 | NM | NM | 29.604 |
| original22-v1/runtime/simd-memory-full | 4.208 | 4.854 | 2.126 | NM | NM | 25.709 |
| original22-v1/runtime/simd-memory-widen | 4.375 | 4.792 | 2.354 | NM | NM | 26.562 |
| original22-v1/runtime/simd-memory-lanes | 4.312 | 4.500 | 2.479 | NM | NM | 26.063 |
| original22-v1/gc/i31-roundtrip | 4.146 | 4.687 | US | NM | NM | 28.584 |
| original22-v1/gc/struct-access | 4.667 | 4.500 | US | NM | NM | 29.354 |
| original22-v1/gc/struct-allocation | 4.417 | 4.750 | US | NM | NM | 26.875 |
| original22-v1/gc/struct-indirect-access | 5.000 | 4.875 | US | NM | NM | 27.104 |
| original22-v1/threads/private-memory | 4.396 | 4.500 | 2.333 | NM | NM | 26.166 |
| original22-v1/threads/shared-memory | 4.562 | 4.521 | 2.604 | NM | NM | 25.521 |
| original22-v1/threads/atomic-rmw | 4.812 | 4.750 | 3.125 | NM | NM | 25.959 |
| original22-v1/threads/notify-empty | 4.167 | 4.375 | 2.271 | NM | NM | 27.271 |
| original22-v1/threads/wait-mismatch | 4.688 | 5.604 | 2.708 | NM | NM | 28.230 |
| original22-v1/threads/wait-timeout | 4.812 | 5.229 | 2.688 | NM | NM | 30.270 |
| original22-v1/applications/sha256-clang | 4.542 | 5.063 | 2.438 | NM | NM | 26.438 |
| original22-v1/applications/sha256-rust | 4.646 | 4.771 | 2.417 | NM | NM | 28.437 |
| gc-sustained-v1/gc/i31-roundtrip | 4.542 | 4.479 | US | NM | NM | 27.021 |
| gc-sustained-v1/gc/struct-access | 4.604 | 5.083 | US | NM | NM | 26.083 |
| gc-sustained-v1/gc/struct-allocation | 4.729 | 4.583 | US | NM | NM | 24.709 |
| gc-sustained-v1/gc/struct-indirect-access | 4.604 | 4.750 | US | NM | NM | 33.312 |
| threads-sustained-v2/threads/private-memory | 5.188 | 5.062 | 2.688 | NM | NM | 29.480 |
| threads-sustained-v2/threads/shared-memory | 5.000 | 4.646 | 2.333 | NM | NM | 30.230 |
| threads-sustained-v2/threads/atomic-rmw | 4.208 | 4.792 | 2.333 | NM | NM | 26.104 |
| threads-sustained-v2/threads/notify-empty | 4.938 | 4.625 | 2.021 | NM | NM | 24.812 |
| threads-sustained-v2/threads/wait-mismatch | 3.979 | 4.812 | 2.312 | NM | NM | 25.792 |
| threads-sustained-v2/threads/wait-timeout | 4.812 | 5.250 | 2.958 | NM | NM | 27.750 |
| generic-calls-v1/core/generic-calls | 5.688 | 4.896 | 2.521 | NM | NM | 23.875 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 62.501 | 55.313 | 41.375 | NM | NM | 35.979 |
| original22-v1/core/integer-loop | 66.625 | 61.291 | 44.479 | NM | NM | 37.812 |
| original22-v1/core/function-calls | 80.001 | 67.041 | 43.771 | NM | NM | 36.812 |
| original22-v1/core/memory-load-store | 77.708 | 80.688 | 45.604 | NM | NM | 40.229 |
| original22-v1/memory/narrow-unaligned | 93.209 | 88.854 | 52.229 | NM | NM | 43.875 |
| original22-v1/memory/dependent-address | 83.958 | 79.750 | 49.312 | NM | NM | 44.416 |
| original22-v1/memory/float64-unaligned | 83.001 | 83.938 | 49.750 | NM | NM | 43.271 |
| original22-v1/runtime/simd-memory-full | 75.416 | 67.146 | 49.063 | NM | NM | 41.374 |
| original22-v1/runtime/simd-memory-widen | 72.542 | 68.979 | 46.604 | NM | NM | 38.791 |
| original22-v1/runtime/simd-memory-lanes | 74.459 | 69.895 | 49.291 | NM | NM | 41.520 |
| original22-v1/gc/i31-roundtrip | 67.979 | 57.042 | US | NM | NM | 35.520 |
| original22-v1/gc/struct-access | 69.896 | 60.416 | US | NM | NM | 45.459 |
| original22-v1/gc/struct-allocation | 70.646 | 58.292 | US | NM | NM | 47.291 |
| original22-v1/gc/struct-indirect-access | 83.271 | 80.126 | US | NM | NM | 52.250 |
| original22-v1/threads/private-memory | 61.000 | 56.417 | 40.270 | NM | NM | 38.480 |
| original22-v1/threads/shared-memory | 66.438 | 54.459 | 40.959 | NM | NM | 38.229 |
| original22-v1/threads/atomic-rmw | 61.625 | 56.291 | 49.396 | NM | NM | 40.022 |
| original22-v1/threads/notify-empty | 63.604 | 53.646 | 43.438 | NM | NM | 38.229 |
| original22-v1/threads/wait-mismatch | 68.250 | 59.396 | 45.750 | NM | NM | 40.729 |
| original22-v1/threads/wait-timeout | 67.417 | 60.041 | 46.209 | NM | NM | 40.916 |
| original22-v1/applications/sha256-clang | 262.48 | 362.63 | 80.937 | NM | NM | 81.958 |
| original22-v1/applications/sha256-rust | 281.33 | 346.94 | 88.625 | NM | NM | 77.730 |
| gc-sustained-v1/gc/i31-roundtrip | 66.105 | 54.667 | US | NM | NM | 36.959 |
| gc-sustained-v1/gc/struct-access | 70.853 | 62.375 | US | NM | NM | 44.313 |
| gc-sustained-v1/gc/struct-allocation | 67.916 | 60.605 | US | NM | NM | 46.730 |
| gc-sustained-v1/gc/struct-indirect-access | 84.897 | 75.688 | US | NM | NM | 53.187 |
| threads-sustained-v2/threads/private-memory | 62.584 | 56.291 | 43.188 | NM | NM | 40.500 |
| threads-sustained-v2/threads/shared-memory | 66.604 | 54.791 | 43.625 | NM | NM | 39.167 |
| threads-sustained-v2/threads/atomic-rmw | 61.562 | 54.771 | 42.562 | NM | NM | 39.395 |
| threads-sustained-v2/threads/notify-empty | 62.104 | 53.959 | 40.875 | NM | NM | 40.167 |
| threads-sustained-v2/threads/wait-mismatch | 65.416 | 55.584 | 42.020 | NM | NM | 38.875 |
| threads-sustained-v2/threads/wait-timeout | 66.604 | 59.062 | 47.291 | NM | NM | 39.916 |
| generic-calls-v1/core/generic-calls | 76.209 | 66.438 | 41.270 | NM | NM | 41.605 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 7.771 | 7.792 | 17.708 | NM | NM | 29.520 |
| original22-v1/core/integer-loop | 7.979 | 7.771 | 16.166 | NM | NM | 29.584 |
| original22-v1/core/function-calls | 7.958 | 7.271 | 14.396 | NM | NM | 25.438 |
| original22-v1/core/memory-load-store | 11.624 | 13.750 | 16.584 | NM | NM | 45.291 |
| original22-v1/memory/narrow-unaligned | 12.166 | 13.396 | 18.125 | NM | NM | 52.188 |
| original22-v1/memory/dependent-address | 13.041 | 14.916 | 16.937 | NM | NM | 54.041 |
| original22-v1/memory/float64-unaligned | 13.459 | 13.709 | 16.791 | NM | NM | 52.041 |
| original22-v1/runtime/simd-memory-full | 13.229 | 13.521 | 16.729 | NM | NM | 51.541 |
| original22-v1/runtime/simd-memory-widen | 13.792 | 12.312 | 16.333 | NM | NM | 55.688 |
| original22-v1/runtime/simd-memory-lanes | 13.666 | 12.688 | 17.333 | NM | NM | 53.625 |
| original22-v1/gc/i31-roundtrip | 7.708 | 7.229 | US | NM | NM | 26.250 |
| original22-v1/gc/struct-access | 8.521 | 8.146 | US | NM | NM | 28.396 |
| original22-v1/gc/struct-allocation | 9.417 | 9.063 | US | NM | NM | 29.959 |
| original22-v1/gc/struct-indirect-access | 14.938 | 15.834 | US | NM | NM | 31.459 |
| original22-v1/threads/private-memory | 10.583 | 12.666 | 16.437 | NM | NM | 46.250 |
| original22-v1/threads/shared-memory | 15.125 | 12.688 | 15.646 | NM | NM | 47.083 |
| original22-v1/threads/atomic-rmw | 15.104 | 15.209 | 16.771 | NM | NM | 46.521 |
| original22-v1/threads/notify-empty | 14.584 | 12.250 | 16.584 | NM | NM | 47.291 |
| original22-v1/threads/wait-mismatch | 16.979 | 13.479 | 17.854 | NM | NM | 50.979 |
| original22-v1/threads/wait-timeout | 16.791 | 15.938 | 17.062 | NM | NM | 50.626 |
| original22-v1/applications/sha256-clang | 15.000 | 22.083 | 18.541 | NM | NM | 57.396 |
| original22-v1/applications/sha256-rust | 27.771 | 76.645 | 32.959 | NM | NM | 58.438 |
| gc-sustained-v1/gc/i31-roundtrip | 8.251 | 7.354 | US | NM | NM | 28.895 |
| gc-sustained-v1/gc/struct-access | 8.438 | 8.312 | US | NM | NM | 24.562 |
| gc-sustained-v1/gc/struct-allocation | 9.521 | 9.646 | US | NM | NM | 28.521 |
| gc-sustained-v1/gc/struct-indirect-access | 15.834 | 15.396 | US | NM | NM | 34.812 |
| threads-sustained-v2/threads/private-memory | 10.667 | 13.562 | 16.646 | NM | NM | 50.980 |
| threads-sustained-v2/threads/shared-memory | 15.458 | 14.021 | 16.584 | NM | NM | 47.770 |
| threads-sustained-v2/threads/atomic-rmw | 14.312 | 13.749 | 17.041 | NM | NM | 46.479 |
| threads-sustained-v2/threads/notify-empty | 15.271 | 13.729 | 16.042 | NM | NM | 46.395 |
| threads-sustained-v2/threads/wait-mismatch | 14.667 | 13.688 | 16.270 | NM | NM | 47.355 |
| threads-sustained-v2/threads/wait-timeout | 15.646 | 16.146 | 22.833 | NM | NM | 47.146 |
| generic-calls-v1/core/generic-calls | 8.312 | 9.292 | 16.958 | NM | NM | 22.667 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.230 | 0.479 | 0.271 | NM | NM | 0.230 |
| original22-v1/core/integer-loop | 0.230 | 0.417 | 0.208 | NM | NM | 0.250 |
| original22-v1/core/function-calls | 0.230 | 0.438 | 0.230 | NM | NM | 0.208 |
| original22-v1/core/memory-load-store | 0.230 | 0.416 | 0.250 | NM | NM | 0.188 |
| original22-v1/memory/narrow-unaligned | 0.230 | 0.416 | 0.208 | NM | NM | 0.230 |
| original22-v1/memory/dependent-address | 0.271 | 0.479 | 0.250 | NM | NM | 0.230 |
| original22-v1/memory/float64-unaligned | 0.292 | 0.417 | 0.208 | NM | NM | 0.250 |
| original22-v1/runtime/simd-memory-full | 0.250 | 0.458 | 0.229 | NM | NM | 0.208 |
| original22-v1/runtime/simd-memory-widen | 0.209 | 0.375 | 0.208 | NM | NM | 0.230 |
| original22-v1/runtime/simd-memory-lanes | 0.250 | 0.396 | 0.188 | NM | NM | 0.208 |
| original22-v1/gc/i31-roundtrip | 0.209 | 0.416 | US | NM | NM | 0.271 |
| original22-v1/gc/struct-access | 0.209 | 0.396 | US | NM | NM | 0.209 |
| original22-v1/gc/struct-allocation | 0.271 | 0.417 | US | NM | NM | 0.209 |
| original22-v1/gc/struct-indirect-access | 0.251 | 0.500 | US | NM | NM | 0.250 |
| original22-v1/threads/private-memory | 0.250 | 0.416 | 0.230 | NM | NM | 0.208 |
| original22-v1/threads/shared-memory | 0.250 | 0.479 | 0.187 | NM | NM | 0.208 |
| original22-v1/threads/atomic-rmw | 0.250 | 0.458 | 0.230 | NM | NM | 0.209 |
| original22-v1/threads/notify-empty | 0.250 | 0.500 | 0.167 | NM | NM | 0.208 |
| original22-v1/threads/wait-mismatch | 0.291 | 0.500 | 0.250 | NM | NM | 0.188 |
| original22-v1/threads/wait-timeout | 0.250 | 0.521 | 0.230 | NM | NM | 0.208 |
| original22-v1/applications/sha256-clang | 0.188 | 0.417 | 0.208 | NM | NM | 0.208 |
| original22-v1/applications/sha256-rust | 0.250 | 0.438 | 0.312 | NM | NM | 0.208 |
| gc-sustained-v1/gc/i31-roundtrip | 0.208 | 0.459 | US | NM | NM | 0.208 |
| gc-sustained-v1/gc/struct-access | 0.208 | 0.416 | US | NM | NM | 0.208 |
| gc-sustained-v1/gc/struct-allocation | 0.230 | 0.417 | US | NM | NM | 0.250 |
| gc-sustained-v1/gc/struct-indirect-access | 0.209 | 0.438 | US | NM | NM | 0.208 |
| threads-sustained-v2/threads/private-memory | 0.230 | 0.438 | 0.229 | NM | NM | 0.250 |
| threads-sustained-v2/threads/shared-memory | 0.312 | 0.438 | 0.208 | NM | NM | 0.229 |
| threads-sustained-v2/threads/atomic-rmw | 0.292 | 0.459 | 0.208 | NM | NM | 0.208 |
| threads-sustained-v2/threads/notify-empty | 0.250 | 0.459 | 0.187 | NM | NM | 0.271 |
| threads-sustained-v2/threads/wait-mismatch | 0.230 | 0.417 | 0.208 | NM | NM | 0.209 |
| threads-sustained-v2/threads/wait-timeout | 0.271 | 0.459 | 0.250 | NM | NM | 0.208 |
| generic-calls-v1/core/generic-calls | 0.250 | 0.417 | 0.271 | NM | NM | 0.230 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.771 | 7.667 | 4.000 | NM | NM | 8.062 |
| original22-v1/core/integer-loop | 32,084.21 | 26,780.58 | 88,321.60 | NM | NM | 51,080.56 |
| original22-v1/core/function-calls | 48,721.88 | 46,788.17 | 89,244.15 | NM | NM | 64,155.15 |
| original22-v1/core/memory-load-store | 66,544.69 | 59,073.54 | 148,488.62 | NM | NM | 94,605.29 |
| original22-v1/memory/narrow-unaligned | 89,189.23 | 76,404.15 | 205,054.56 | NM | NM | 124,329.48 |
| original22-v1/memory/dependent-address | 42,236.46 | 35,647.67 | 99,711.40 | NM | NM | 55,616.44 |
| original22-v1/memory/float64-unaligned | 69,419.15 | 61,789.71 | 142,335.50 | NM | NM | 90,929.79 |
| original22-v1/runtime/simd-memory-full | 76,998.46 | 67,486.15 | 152,905.90 | NM | NM | 78,615.25 |
| original22-v1/runtime/simd-memory-widen | 64,321.19 | 55,474.08 | 144,557.62 | NM | NM | 61,186.85 |
| original22-v1/runtime/simd-memory-lanes | 95,155.48 | 86,996.42 | 186,899.27 | NM | NM | 91,698.73 |
| original22-v1/gc/i31-roundtrip | 367.96 | 354.48 | US | NM | NM | 467.60 |
| original22-v1/gc/struct-access | 442.71 | 437.00 | US | NM | NM | 565.81 |
| original22-v1/gc/struct-allocation | 973.98 | 1,036.33 | US | NM | NM | 1,211.52 |
| original22-v1/gc/struct-indirect-access | 787.19 | 760.23 | US | NM | NM | 1,013.25 |
| original22-v1/threads/private-memory | 398.58 | 356.27 | 645.37 | NM | NM | 405.65 |
| original22-v1/threads/shared-memory | 788.31 | 346.90 | 635.98 | NM | NM | 405.54 |
| original22-v1/threads/atomic-rmw | 538.83 | 349.79 | 490.65 | NM | NM | 342.33 |
| original22-v1/threads/notify-empty | 328.12 | 286.40 | 428.02 | NM | NM | 412.52 |
| original22-v1/threads/wait-mismatch | 599.60 | 387.62 | 719.10 | NM | NM | 469.54 |
| original22-v1/threads/wait-timeout | 587.10 | 385.90 | 3,300.77 | NM | NM | 1,128.27 |
| original22-v1/applications/sha256-clang | 92,766.90 | 80,917.94 | 209,414.06 | NM | NM | 142,602.31 |
| original22-v1/applications/sha256-rust | 90,795.96 | 81,094.42 | 211,405.10 | NM | NM | 126,135.37 |
| gc-sustained-v1/gc/i31-roundtrip | 35,272.98 | 31,899.21 | US | NM | NM | 48,045.21 |
| gc-sustained-v1/gc/struct-access | 43,403.52 | 42,041.08 | US | NM | NM | 57,106.35 |
| gc-sustained-v1/gc/struct-allocation | 51,393.38 | 34,487.60 | US | NM | NM | 32,200.96 |
| gc-sustained-v1/gc/struct-indirect-access | 77,379.85 | 73,297.94 | US | NM | NM | 101,843.33 |
| threads-sustained-v2/threads/private-memory | 3,966.31 | 3,462.92 | 6,514.17 | NM | NM | 3,905.33 |
| threads-sustained-v2/threads/shared-memory | 7,485.73 | 3,338.02 | 6,365.65 | NM | NM | 3,919.21 |
| threads-sustained-v2/threads/atomic-rmw | 5,227.67 | 3,315.35 | 4,821.35 | NM | NM | 3,255.40 |
| threads-sustained-v2/threads/notify-empty | 3,137.44 | 2,765.83 | 4,088.29 | NM | NM | 4,147.92 |
| threads-sustained-v2/threads/wait-mismatch | 5,642.83 | 3,650.65 | 6,950.21 | NM | NM | 4,315.67 |
| threads-sustained-v2/threads/wait-timeout | 5,630.00 | 3,715.83 | 37,003.92 | NM | NM | 10,180.46 |
| generic-calls-v1/core/generic-calls | 570,971.77 | 477,417.65 | 956,255.48 | NM | NM | 596,493.38 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2.354 | 1.854 | 6.979 | NM | NM | 7.750 |
| original22-v1/core/integer-loop | 2.251 | 2.167 | 8.021 | NM | NM | 7.479 |
| original22-v1/core/function-calls | 2.083 | 1.855 | 4.708 | NM | NM | 11.375 |
| original22-v1/core/memory-load-store | 2.562 | 2.333 | 4.250 | NM | NM | 53.124 |
| original22-v1/memory/narrow-unaligned | 2.458 | 2.271 | 4.875 | NM | NM | 52.666 |
| original22-v1/memory/dependent-address | 3.792 | 2.208 | 9.792 | NM | NM | 52.688 |
| original22-v1/memory/float64-unaligned | 2.438 | 2.188 | 5.083 | NM | NM | 47.749 |
| original22-v1/runtime/simd-memory-full | 2.396 | 2.479 | 6.688 | NM | NM | 53.708 |
| original22-v1/runtime/simd-memory-widen | 2.750 | 2.375 | 9.250 | NM | NM | 52.751 |
| original22-v1/runtime/simd-memory-lanes | 2.542 | 2.292 | 7.000 | NM | NM | 58.980 |
| original22-v1/gc/i31-roundtrip | 1.604 | 1.125 | US | NM | NM | 5.250 |
| original22-v1/gc/struct-access | 2.146 | 2.021 | US | NM | NM | 10.646 |
| original22-v1/gc/struct-allocation | 95.125 | 56.979 | US | NM | NM | 1,586.35 |
| original22-v1/gc/struct-indirect-access | 1.875 | 2.062 | US | NM | NM | 8.541 |
| original22-v1/threads/private-memory | 2.124 | 1.958 | 4.688 | NM | NM | 41.000 |
| original22-v1/threads/shared-memory | 10.291 | 6.292 | 5.542 | NM | NM | 42.438 |
| original22-v1/threads/atomic-rmw | 10.167 | 5.104 | 4.000 | NM | NM | 43.688 |
| original22-v1/threads/notify-empty | 8.438 | 5.917 | 4.646 | NM | NM | 42.062 |
| original22-v1/threads/wait-mismatch | 18.479 | 5.979 | 8.438 | NM | NM | 46.062 |
| original22-v1/threads/wait-timeout | 14.667 | 7.000 | 5.688 | NM | NM | 43.959 |
| original22-v1/applications/sha256-clang | 2.646 | 2.292 | 9.208 | NM | NM | 57.166 |
| original22-v1/applications/sha256-rust | 2.292 | 2.249 | 4.625 | NM | NM | 54.833 |
| gc-sustained-v1/gc/i31-roundtrip | 2.271 | 2.250 | US | NM | NM | 7.146 |
| gc-sustained-v1/gc/struct-access | 2.312 | 2.229 | US | NM | NM | 8.708 |
| gc-sustained-v1/gc/struct-allocation | 2,757.46 | 1,647.48 | US | NM | NM | 58,077.75 |
| gc-sustained-v1/gc/struct-indirect-access | 2.292 | 2.229 | US | NM | NM | 9.834 |
| threads-sustained-v2/threads/private-memory | 2.396 | 2.271 | 6.354 | NM | NM | 43.833 |
| threads-sustained-v2/threads/shared-memory | 12.249 | 8.834 | 4.646 | NM | NM | 47.104 |
| threads-sustained-v2/threads/atomic-rmw | 17.854 | 7.042 | 7.271 | NM | NM | 47.375 |
| threads-sustained-v2/threads/notify-empty | 12.166 | 6.104 | 4.333 | NM | NM | 47.896 |
| threads-sustained-v2/threads/wait-mismatch | 13.583 | 8.584 | 5.187 | NM | NM | 47.062 |
| threads-sustained-v2/threads/wait-timeout | 14.917 | 8.438 | 4.000 | NM | NM | 48.229 |
| generic-calls-v1/core/generic-calls | 2.167 | 2.062 | 4.875 | NM | NM | 6.542 |

</details>

### WasmEdge AOT preparation detail

| Workload | Compile + validation + artifact I/O, us | Native image load, us | Post-load validation, us |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 20,328.35 | 256,653.79 | 3.125 |
| original22-v1/core/integer-loop | 19,220.31 | 254,017.06 | 7.292 |
| original22-v1/core/function-calls | 19,032.69 | 255,892.46 | 5.500 |
| original22-v1/core/memory-load-store | 23,106.04 | 259,058.21 | 6.250 |
| original22-v1/memory/narrow-unaligned | 20,159.81 | 255,192.81 | 7.979 |
| original22-v1/memory/dependent-address | 21,820.12 | 254,771.73 | 7.167 |
| original22-v1/memory/float64-unaligned | 23,471.81 | 251,277.06 | 7.625 |
| original22-v1/runtime/simd-memory-full | 18,348.25 | 254,431.52 | 7.271 |
| original22-v1/runtime/simd-memory-widen | 19,262.15 | 258,162.69 | 6.979 |
| original22-v1/runtime/simd-memory-lanes | 19,188.40 | 253,903.92 | 6.229 |
| original22-v1/gc/i31-roundtrip | 18,133.94 | 252,682.52 | 4.583 |
| original22-v1/gc/struct-access | 19,739.12 | 261,408.85 | 3.271 |
| original22-v1/gc/struct-allocation | 18,247.92 | 256,852.04 | 8.313 |
| original22-v1/gc/struct-indirect-access | 18,858.79 | 255,547.60 | 7.208 |
| original22-v1/threads/private-memory | 18,045.50 | 253,830.48 | 5.375 |
| original22-v1/threads/shared-memory | 17,974.44 | 256,090.31 | 5.938 |
| original22-v1/threads/atomic-rmw | 18,131.02 | 259,040.29 | 5.125 |
| original22-v1/threads/notify-empty | 17,534.23 | 261,647.83 | 5.604 |
| original22-v1/threads/wait-mismatch | 20,098.67 | 254,710.31 | 6.750 |
| original22-v1/threads/wait-timeout | 20,256.52 | 253,528.27 | 5.958 |
| original22-v1/applications/sha256-clang | 23,958.92 | 256,966.21 | 12.438 |
| original22-v1/applications/sha256-rust | 23,678.27 | 256,960.85 | 9.438 |
| gc-sustained-v1/gc/i31-roundtrip | 17,714.25 | 257,809.33 | 5.167 |
| gc-sustained-v1/gc/struct-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 19,513.50 | 256,863.40 | 7.354 |
| threads-sustained-v2/threads/shared-memory | 18,023.40 | 252,028.62 | 6.229 |
| threads-sustained-v2/threads/atomic-rmw | 18,500.04 | 259,493.67 | 5.375 |
| threads-sustained-v2/threads/notify-empty | 17,976.79 | 258,439.75 | 5.562 |
| threads-sustained-v2/threads/wait-mismatch | 19,763.88 | 252,008.29 | 5.896 |
| threads-sustained-v2/threads/wait-timeout | 22,746.56 | 259,502.54 | 5.479 |
| generic-calls-v1/core/generic-calls | 18,427.52 | 258,739.54 | 4.500 |

### P3 phase observations

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B / cleanup | 53.667 | 495.75 | NM | 237.83 | NM | NM |
| 16 B / compileValidate | 58,450.83 | 91,224.21 | NM | 51,310.10 | NM | NM |
| 16 B / engineSetup | 0.354 | 0.167 | NM | 142.54 | NM | NM |
| 16 B / executorSetup | N/A | N/A | NM | 44.209 | NM | NM |
| 16 B / exportLookup | 0.645 | 0.979 | NM | 8.918 | NM | NM |
| 16 B / firstCall | 8,684.60 | 19,821.38 | NM | 7,904.87 | NM | NM |
| 16 B / instantiate | 509.44 | 731.23 | NM | 197.90 | NM | NM |
| 16 B / load | 62.666 | 64.000 | NM | 56.188 | NM | NM |
| 16 B / processStartup | 6,160.38 | 12,253.67 | NM | 6,369.52 | NM | NM |
| 16 B / providerSetup | 47.479 | 189.62 | NM | 228.29 | NM | NM |
| 16 B / requestPreparation | 0.750 | 0.750 | NM | N/A | NM | NM |
| 65536 B / cleanup | 51.688 | 542.52 | NM | 248.42 | NM | NM |
| 65536 B / compileValidate | 59,682.58 | 94,623.58 | NM | 51,696.06 | NM | NM |
| 65536 B / engineSetup | 0.375 | 0.167 | NM | 136.77 | NM | NM |
| 65536 B / executorSetup | N/A | N/A | NM | 42.625 | NM | NM |
| 65536 B / exportLookup | 0.730 | 1.145 | NM | 9.104 | NM | NM |
| 65536 B / firstCall | 12,755.58 | 23,005.00 | NM | 9,076.56 | NM | NM |
| 65536 B / instantiate | 503.98 | 742.90 | NM | 199.88 | NM | NM |
| 65536 B / load | 59.999 | 59.312 | NM | 51.812 | NM | NM |
| 65536 B / processStartup | 6,043.69 | 12,448.65 | NM | 6,417.17 | NM | NM |
| 65536 B / providerSetup | 47.438 | 173.27 | NM | 232.29 | NM | NM |
| 65536 B / requestPreparation | 8.291 | 6.770 | NM | N/A | NM | NM |

## Native controls (different contracts)

Unchecked C kernels omit runtime safety/lifetime contracts; native Go implements HTTP replies without Component/WASI. These are overhead controls with different contracts. The table reports their measured timing directly. The general-call C control runs twice as many iterations of one recurrence; The optimized Wasm combines two recurrence updates into one affine operation.

| Workload | Unchecked native C, ns/iteration |
| --- | ---: |
| original22-v1/core/integer-loop | 1.614 |
| original22-v1/core/function-calls | 0.775 |
| original22-v1/core/memory-load-store | 0.272 |
| original22-v1/memory/narrow-unaligned | 0.485 |
| original22-v1/memory/dependent-address | 1.426 |
| original22-v1/memory/float64-unaligned | 0.722 |
| original22-v1/runtime/simd-memory-full | 3.669 |
| original22-v1/runtime/simd-memory-widen | 1.038 |
| original22-v1/runtime/simd-memory-lanes | 4.494 |
| original22-v1/applications/sha256-clang | 170.37 |
| original22-v1/applications/sha256-rust | 160.83 |
| generic-calls-v1/core/generic-calls | 2.133 |

| Payload | Native Go CPU, us/request | Native Go requests/s |
| --- | ---: | ---: |
| 16 B | 2.757 | 1,355,705.60 |
| 65536 B | 77.718 | 33,468.87 |

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


### Narrow memory investigation

The initial PR revision exposed a +3.09% CPU time increase in the LLGo
interpreter for narrow unaligned accesses. Retired instructions differed
by +0.002%, and cycles per instruction increased 3.30%. Matching the
interpreter loop address left a +3.03% increase in CPU time. The initial
report links both counter replays and the disassembly.

A scalar load descriptor records the width, extension shift and result mask.
The range helper validates both unsigned address additions and checks the
exclusive endpoint against the current logical size. Existing exhaustive
tests cover widths, signed results, carry, logical limits, and rejected stores.

The independent 288-process screen uses identical guest inputs and checksums
across both compilers, the initial PR build, and main. It reports **+1.37% CPU
speed** for Go and **+2.46%** for LLGo against main in the previously regressing
narrow case, with resolved A/A controls. The four fixtures each retire fewer
instructions with both compilers. The separate shape-only arm leaves the
LLGo narrow CPU interval crossing parity. The screen, its exact source
overlay, and initial diagnosis are archived independently from this report's
fresh full comparison.

- [Initial PR revision and adverse evidence](https://github.com/cpunion/wasmvm/blob/d4c62f8f790eb402119e309d20c21d7d78edd173/docs/research/performance-p3-scalar-leaves-2026-11-21.md)

### Earlier full-candidate follow-up

The independently rebuilt exact candidate passed a 408-process counter replay.
Go and LLGo respectively reduced CPU time in the narrow unaligned interpreter
by 0.71% and 2.58% versus main. Both controls resolved. The primary LLGo
empty-notify interval indicated a slowdown. The replay measured 1.47% lower
CPU time, alongside 0.44% fewer retired instructions.

The same replay resolves excess work in two call profiles. Go nested
interpretation uses 1.06% more CPU and retires 0.81% more instructions. LLGo
generic-call interpretation uses 0.96% more CPU and retires 0.31% more
instructions. Go interpretation of sustained GC indirect access uses 1.24%
more CPU and retires 0.003% more instructions. Its extra cycles remain a
separate observation. The source now classifies scalar descriptors only on the host path. An
independent 216-process screen removes 1.04% of Go nested-call instructions
and 0.83% of LLGo generic-call instructions relative to the previous source.
Both equivalent guest call profiles return to near the main branch's CPU
time. Narrow accesses retain their gains. The screen keeps CPU, wall time,
retired instructions and cycles separate.

### Final candidate verification

A fresh 288-process check investigates all ten unresolved adverse Core
intervals from the final primary event along with both narrow-memory
interpreter arms. Narrow-memory CPU speed improves 1.58% with Go and 2.81%
with LLGo against main. Both controls resolve. The adverse primary cases
remain unresolved in the replay. Go shared-memory interpreter CPU has a
+2.27% time point with unresolved controls and a resolved −0.26% retired
instruction change. The measured runtime cost there remains open.

A separate 96-process linker experiment examines GC indirect access. The
sustained Go interpreter's default-link CPU interval is +1.66%
[+1.04%, +2.29%]. With both source revisions linked at 64-byte
function alignment, the CPU difference is −0.24% [−0.60%, +0.31%].
Retired instructions match within 0.01% in both comparisons. The linker
intervention moves several code and data addresses. It identifies layout
sensitivity and leaves the exact microarchitectural cause open. The primary
matrix uses the default linker.

## Reproducible evidence

- [Machine-readable results, all controls and source contrasts](data/p3-scalar-leaves-2026-11-21.json).
- [All 5,508 primary process records](data/p3-scalar-leaves-2026-11-21-raw.jsonl.gz), SHA-256 `977695cc33128834b4dc8258876b250b725d3c04b76f9bdb932619db2b720140`.
- [Frozen plans, commands, hashes, analysis/export scripts, semantic checks and failure evidence](data/p3-scalar-leaves-2026-11-21-protocol.tar.gz), SHA-256 `f8190950f4d2b4c9a18498280eff7d0c32988be599a77d7788283745135c37a0`.
- Per-request vectors unused by these metrics are projected to count/extrema/hash; no latency-tail claim. Public load records omit unrelated desktop command arguments. The separately archived screening observations are excluded from primary ratios.
