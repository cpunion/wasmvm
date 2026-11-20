# Guarded ARM64 loop registers and runtime comparison, 2026-11-21

Guarded ARM64 memory loops can save and reuse the three callee-saved call-state registers after every entry guard succeeds. Normal exits transition to the enclosing register plan and restore the call state. The private backedge still polls and materializes canonical slots before publishing a continuation. Guards, independent entry, metering, and guest-visible bounds retain their ordinary semantics.

- SHA-256 Clang, GO: 253.94 → 219.78 ns/message block. Paired source contrast: 0.891 [0.883, 0.900]; Δ -10.91%; speed +12.25% [+11.12%, +13.30%]; unresolved. Against Wasmtime: 1.066 [1.031, 1.098]; Δ +6.56%; speed -6.16% [-8.95%, -3.00%]; unresolved. Against WasmEdge: 1.198 [1.167, 1.223]; Δ +19.84%; speed -16.56% [-18.25%, -14.32%]; unresolved.
- SHA-256 Clang, LLGO: 253.45 → 214.39 ns/message block. Paired source contrast: 0.870 [0.858, 0.879]; Δ -13.05%; speed +15.00% [+13.80%, +16.58%]; unresolved. Against Wasmtime: 1.038 [1.006, 1.072]; Δ +3.83%; speed -3.69% [-6.69%, -0.58%]; unresolved. Against WasmEdge: 1.168 [1.138, 1.197]; Δ +16.77%; speed -14.36% [-16.46%, -12.14%]; unresolved.
- SHA-256 Rust, GO: 267.65 → 235.08 ns/message block. Paired source contrast: 0.872 [0.841, 0.894]; Δ -12.83%; speed +14.72% [+11.83%, +18.96%]; unresolved. Against Wasmtime: 1.118 [1.101, 1.132]; Δ +11.81%; speed -10.56% [-11.69%, -9.14%]; unresolved. Against WasmEdge: 1.091 [1.005, 1.163]; Δ +9.08%; speed -8.32% [-13.98%, -0.47%]; unresolved.
- SHA-256 Rust, LLGO: 269.02 → 233.72 ns/message block. Paired source contrast: 0.849 [0.781, 0.907]; Δ -15.07%; speed +17.74% [+10.21%, +27.99%]; unresolved. Against Wasmtime: 1.107 [1.089, 1.124]; Δ +10.71%; speed -9.67% [-11.00%, -8.21%]; unresolved. Against WasmEdge: 1.080 [0.992, 1.153]; Δ +8.00%; speed -7.41% [-13.30%, +0.79%]; unresolved.

All 33 core workloads and two complete P3 payloads are measured on the same pinned toolchains. Tables give preparation and steady execution independently. The performance acceptance gate remains open.

## Percentage differences

Cells show `absolute observation (signed speed Δ%)`. Positive speed change means faster. Native execution and P3 use **Wasmtime** as a fixed reference; interpreter execution uses **wazero**, with **WasmEdge** explicitly named for GC rows where wazero rejects the feature. Using a fixed reference keeps the percentages comparable across revisions. The fastest observed runtime varies by workload.

For time, speed Δ = `(reference / candidate - 1) * 100%`; for throughput, speed Δ = `(candidate / reference - 1) * 100%`. Thus 100 ns to 50 ns is **−50% time**, **+100% speed**. P3 CPU percentages mean work per CPU time, separate from wall-clock throughput. Instruction/allocation/byte counters show quantity Δ rather than speed. Source comparisons retain paired ratio intervals and report quantity Δ plus speed Δ where applicable; reciprocal interval endpoints are reversed.

Competitive cell percentages are descriptive quotients of the displayed medians. Paired block estimates below can differ. `†` on these cells means either the candidate or the named reference has unresolved A/A controls; qualified conclusions require resolved controls. Phase boundaries with different contracts have separate absolute observations. The percentages apply to the frozen event and measured revision.

## Revisions, profiles and protocol

- Engine source: base `08ad6e30` (merged PR #203), measured head `6a531fac`. Reused binaries retain their producer revisions. The build records verify production-source equivalence for each reused binary.
- Go **1.27.3**; LLGo **3c6b33ec**, **`-O3 -lto=full`**, LLVM **22.1.8**, default production GC. Source base/head share the compiler profile.
- wazero **1.12.0**, native compiler/interpreter; Wasmtime **48.0.1**, Cranelift speed; Wasmer **7.3.0**, Cranelift; WasmEdge **0.17.1**, interpreter and **O3 AOT**. Comparison scope is the named profile of each runtime.
- Local Apple M4 Max, macOS arm64, 128 GiB RAM. External desktop load is uncontrolled and recorded; no owned builds, tests, profiles, analysis or edits overlap timing. The timing scope is macOS arm64.
- Core uses six fixed fresh-process blocks, each with symmetric same-binary A/A, two warmups and seven interleaved measured calls. **5,364 successful process records** after excluding exactly three failing sustained WasmEdge AOT cells. All 33 workloads remain represented. Allocation workloads use means including pauses; others use medians.
- P3 uses **72 processes per payload**, six roles with A/A, six fixed blocks, four workers, sixteen clients and twenty interleaved 512-request windows plus a separate flush. Verify status, headers, full body, trailers, CRC and joined guest completion. CPU/instruction counters include the driver and checks. Mach ticks use the recorded timebase.
- Ratios use paired block geometric means and 20,000 fixed-seed bootstrap resamples with marginal 95% intervals. Both relevant A/A intervals must fit `[0.975, 1.025]`. Each interval applies to its own comparison. The protocol fixes the sample size in advance and measures each phase directly.
- The fifth Core block hit the outer collection deadline. A replay verified the original process and sample order. The final two blocks completed under a longer outer deadline with identical binaries, commands and sample counts. The discarded 74 process records are archived separately.

**Legend:** `†` means the role or its percentage reference has an unresolved same-binary timing control. Values are observations; a qualified comparison additionally needs both controls and its ratio interval. `US` means this pinned profile rejected the feature in semantic preflight. `NM` means this benchmark adapter/profile was not measured; this status applies only to this measurement. `N/A` means the metric does not apply. `FAIL` is an execution failure in the named profile.

## Native compiled execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

The reference runtime is **Wasmtime**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,750.50 (-69.30% speed) † | 3,729.00 (-60.89% speed) † | 854.00 (+70.78% speed) † | 1,458.50 (ref) † | 3,688.00 (-60.45% speed) † | 4,041.00 (-63.91% speed) † |
| original22-v1/core/integer-loop | 1.805 (-0.92% speed) † | 1.793 (-0.26% speed) | 1.929 (-7.26% speed) † | 1.789 (ref) | 1.790 (-0.05% speed) † | 1.791 (-0.10% speed) |
| original22-v1/core/function-calls | 0.941 (+14.41% speed) † | 0.938 (+14.81% speed) † | 1.721 (-37.41% speed) † | 1.077 (ref) † | 0.951 (+13.27% speed) † | 0.928 (+16.08% speed) † |
| original22-v1/core/memory-load-store | 0.490 (+7.53% speed) | 0.475 (+11.07% speed) † | 0.949 (-44.45% speed) † | 0.527 (ref) | 0.594 (-11.20% speed) † | 0.312 (+69.04% speed) † |
| original22-v1/memory/narrow-unaligned | 0.765 (+7.28% speed) † | 0.751 (+9.21% speed) † | 1.220 (-32.75% speed) † | 0.821 (ref) † | 0.846 (-2.99% speed) † | 0.593 (+38.44% speed) † |
| original22-v1/memory/dependent-address | 2.190 (+11.47% speed) † | 2.179 (+12.04% speed) † | 3.709 (-34.17% speed) † | 2.442 (ref) † | 2.469 (-1.10% speed) † | 1.689 (+44.53% speed) † |
| original22-v1/memory/float64-unaligned | 0.834 (-4.58% speed) † | 0.826 (-3.69% speed) † | 0.952 (-16.38% speed) † | 0.796 (ref) † | 0.803 (-0.87% speed) † | 0.754 (+5.58% speed) † |
| original22-v1/runtime/simd-memory-full | 3.956 (+7.63% speed) † | 3.875 (+9.90% speed) † | 4.400 (-3.23% speed) † | 4.258 (ref) † | 3.991 (+6.70% speed) † | 4.223 (+0.82% speed) † |
| original22-v1/runtime/simd-memory-widen | 1.167 (-0.19% speed) † | 1.154 (+0.87% speed) † | 1.285 (-9.41% speed) † | 1.164 (ref) † | 1.168 (-0.29% speed) † | 1.162 (+0.19% speed) † |
| original22-v1/runtime/simd-memory-lanes | 5.817 (+17.97% speed) | 5.782 (+18.68% speed) | 7.124 (-3.68% speed) † | 6.862 (ref) | 6.880 (-0.27% speed) | 5.857 (+17.15% speed) † |
| original22-v1/gc/i31-roundtrip | 2.358 (+0.17% speed) † | 2.198 (+7.49% speed) † | US | 2.362 (ref) † | US | 2.144 (+10.20% speed) † |
| original22-v1/gc/struct-access | 2.750 (-38.79% speed) † | 2.165 (-22.23% speed) † | US | 1.683 (ref) † | US | 11.525 (-85.39% speed) † |
| original22-v1/gc/struct-allocation | 33.018 (+126.97% speed) † | 31.835 (+135.41% speed) † | US | 74.941 (ref) | US | 44.756 (+67.44% speed) † |
| original22-v1/gc/struct-indirect-access | 3.127 (+6.60% speed) † | 2.513 (+32.67% speed) † | US | 3.333 (ref) † | US | 29.446 (-88.68% speed) † |
| original22-v1/threads/private-memory | 1.377 (+12.71% speed) † | 0.956 (+62.31% speed) † | 1.931 (-19.63% speed) † | 1.552 (ref) † | 1.898 (-18.22% speed) † | 1.790 (-13.27% speed) † |
| original22-v1/threads/shared-memory | 3.083 (-44.66% speed) † | 2.254 (-24.31% speed) † | 2.029 (-15.91% speed) † | 1.706 (ref) † | 1.994 (-14.42% speed) † | 2.000 (-14.68% speed) † |
| original22-v1/threads/atomic-rmw | 3.242 (-21.14% speed) † | 2.881 (-11.28% speed) † | 2.867 (-10.83% speed) † | 2.556 (ref) † | 5.196 (-50.80% speed) † | 2.835 (-9.85% speed) † |
| original22-v1/threads/notify-empty | 2.279 (+240.68% speed) † | 1.583 (+390.39% speed) † | 23.483 (-66.94% speed) † | 7.765 (ref) † | 5.933 (+30.86% speed) † | 10.438 (-25.61% speed) † |
| original22-v1/threads/wait-mismatch | 2.796 (+305.66% speed) † | 2.121 (+434.76% speed) † | 30.596 (-62.93% speed) † | 11.342 (ref) † | 3.752 (+202.27% speed) † | 4.829 (+134.86% speed) † |
| original22-v1/threads/wait-timeout | 2.435 (+2020.38% speed) † | 1.994 (+2490.08% speed) † | 352.30 (-85.34% speed) † | 51.640 (ref) † | 195.49 (-73.59% speed) † | 72.046 (-28.32% speed) † |
| original22-v1/applications/sha256-clang | 219.78 (-1.81% speed) † | 214.39 (+0.65% speed) † | 307.33 (-29.79% speed) † | 215.79 (ref) † | 214.43 (+0.63% speed) † | 186.17 (+15.91% speed) † |
| original22-v1/applications/sha256-rust | 235.08 (-9.75% speed) † | 233.72 (-9.23% speed) † | 368.27 (-42.39% speed) † | 212.15 (ref) † | 220.30 (-3.70% speed) † | 212.63 (-0.22% speed) † |
| gc-sustained-v1/gc/i31-roundtrip | 1.502 (+33.92% speed) † | 1.503 (+33.78% speed) † | US | 2.011 (ref) | US | 1.447 (+38.98% speed) |
| gc-sustained-v1/gc/struct-access | 1.220 (-0.85% speed) | 1.204 (+0.46% speed) † | US | 1.210 (ref) | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 38.014 (+51.05% speed) † | 39.963 (+43.69% speed) † | US | 57.423 (ref) | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1.904 (+56.13% speed) † | 1.931 (+53.93% speed) † | US | 2.972 (ref) † | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.459 (+152.47% speed) † | 0.375 (+209.17% speed) † | 1.303 (-11.08% speed) † | 1.159 (ref) † | 1.177 (-1.56% speed) † | 1.158 (+0.09% speed) † |
| threads-sustained-v2/threads/shared-memory | 1.579 (-29.01% speed) † | 1.501 (-25.33% speed) † | 1.373 (-18.36% speed) † | 1.121 (ref) † | 1.262 (-11.19% speed) † | 1.204 (-6.91% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 2.341 (-7.17% speed) | 2.233 (-2.66% speed) | 2.395 (-9.23% speed) † | 2.174 (ref) | 4.642 (-53.18% speed) | 2.217 (-1.95% speed) |
| threads-sustained-v2/threads/notify-empty | 1.174 (+561.08% speed) † | 1.057 (+634.18% speed) | 23.583 (-67.09% speed) † | 7.761 (ref) | 5.237 (+48.18% speed) † | 10.061 (-22.87% speed) |
| threads-sustained-v2/threads/wait-mismatch | 1.380 (+647.34% speed) † | 1.285 (+702.73% speed) † | 28.361 (-63.64% speed) | 10.313 (ref) | 2.833 (+264.03% speed) † | 3.686 (+179.79% speed) † |
| threads-sustained-v2/threads/wait-timeout | 1.224 (+4013.73% speed) † | 1.116 (+4409.85% speed) † | 355.47 (-85.84% speed) † | 50.350 (ref) | 199.12 (-74.71% speed) † | 70.643 (-28.73% speed) † |
| generic-calls-v1/core/generic-calls | 1.190 (+98.18% speed) | 1.162 (+103.05% speed) † | 2.466 (-4.34% speed) | 2.359 (ref) | 2.363 (-0.16% speed) | 1.162 (+103.09% speed) |

## Interpreter execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

The reference runtime is **wazero**. GC rows use **WasmEdge**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2,354.50 (-36.29% speed) † | 2,354.50 (-36.29% speed) † | 1,500.00 (ref) † | NM | NM | 3,979.50 (-62.31% speed) † |
| original22-v1/core/integer-loop | 35.816 (+165.18% speed) † | 28.884 (+228.83% speed) † | 94.977 (ref) † | NM | NM | 56.917 (+66.87% speed) † |
| original22-v1/core/function-calls | 58.617 (+91.03% speed) † | 54.420 (+105.76% speed) † | 111.97 (ref) † | NM | NM | 74.801 (+49.70% speed) † |
| original22-v1/core/memory-load-store | 74.741 (+112.43% speed) † | 64.345 (+146.75% speed) † | 158.77 (ref) † | NM | NM | 100.31 (+58.29% speed) † |
| original22-v1/memory/narrow-unaligned | 101.21 (+118.00% speed) † | 88.533 (+149.21% speed) † | 220.63 (ref) † | NM | NM | 134.87 (+63.59% speed) † |
| original22-v1/memory/dependent-address | 49.030 (+137.69% speed) † | 42.621 (+173.44% speed) † | 116.54 (ref) † | NM | NM | 63.333 (+84.01% speed) † |
| original22-v1/memory/float64-unaligned | 77.411 (+98.73% speed) † | 67.979 (+126.31% speed) † | 153.84 (ref) † | NM | NM | 95.526 (+61.05% speed) † |
| original22-v1/runtime/simd-memory-full | 88.752 (+99.14% speed) † | 77.454 (+128.19% speed) | 176.74 (ref) | NM | NM | 90.490 (+95.32% speed) † |
| original22-v1/runtime/simd-memory-widen | 70.188 (+131.17% speed) | 62.648 (+158.99% speed) | 162.25 (ref) | NM | NM | 68.750 (+136.00% speed) † |
| original22-v1/runtime/simd-memory-lanes | 110.10 (+99.70% speed) † | 101.95 (+115.66% speed) † | 219.88 (ref) † | NM | NM | 107.44 (+104.65% speed) † |
| original22-v1/gc/i31-roundtrip (vs WasmEdge) | 41.440 (+23.07% speed) † | 36.825 (+38.49% speed) † | US | NM | NM | 50.998 (ref) † |
| original22-v1/gc/struct-access (vs WasmEdge) | 48.777 (+32.46% speed) † | 46.129 (+40.07% speed) † | US | NM | NM | 64.612 (ref) † |
| original22-v1/gc/struct-allocation (vs WasmEdge) | 100.77 (+17.68% speed) † | 132.26 (-10.35% speed) † | US | NM | NM | 118.58 (ref) † |
| original22-v1/gc/struct-indirect-access (vs WasmEdge) | 85.537 (+25.89% speed) † | 79.679 (+35.15% speed) † | US | NM | NM | 107.68 (ref) † |
| original22-v1/threads/private-memory | 42.087 (+62.16% speed) † | 36.440 (+87.29% speed) † | 68.248 (ref) † | NM | NM | 41.204 (+65.63% speed) † |
| original22-v1/threads/shared-memory | 87.933 (-16.12% speed) † | 41.194 (+79.06% speed) † | 73.763 (ref) † | NM | NM | 46.242 (+59.52% speed) † |
| original22-v1/threads/atomic-rmw | 60.858 (-16.78% speed) † | 37.479 (+35.13% speed) † | 50.646 (ref) † | NM | NM | 37.300 (+35.78% speed) † |
| original22-v1/threads/notify-empty | 34.619 (+34.94% speed) † | 31.081 (+50.30% speed) † | 46.715 (ref) † | NM | NM | 44.185 (+5.72% speed) † |
| original22-v1/threads/wait-mismatch | 70.300 (+17.34% speed) † | 45.340 (+81.93% speed) † | 82.487 (ref) † | NM | NM | 54.690 (+50.83% speed) † |
| original22-v1/threads/wait-timeout | 65.631 (+512.25% speed) † | 41.179 (+875.80% speed) † | 401.82 (ref) † | NM | NM | 115.72 (+247.23% speed) † |
| original22-v1/applications/sha256-clang | 21,037.06 (+111.91% speed) † | 17,755.56 (+151.07% speed) † | 44,578.78 (ref) † | NM | NM | 29,270.97 (+52.30% speed) † |
| original22-v1/applications/sha256-rust | 20,345.60 (+130.56% speed) † | 17,701.88 (+164.99% speed) † | 46,908.69 (ref) † | NM | NM | 28,097.14 (+66.95% speed) † |
| gc-sustained-v1/gc/i31-roundtrip (vs WasmEdge) | 40.064 (+24.68% speed) † | 34.591 (+44.41% speed) † | US | NM | NM | 49.952 (ref) |
| gc-sustained-v1/gc/struct-access (vs WasmEdge) | 46.330 (+36.40% speed) † | 44.531 (+41.91% speed) † | US | NM | NM | 63.193 (ref) † |
| gc-sustained-v1/gc/struct-allocation (vs WasmEdge) | 111.84 (+8.89% speed) † | 139.30 (-12.57% speed) † | US | NM | NM | 121.79 (ref) † |
| gc-sustained-v1/gc/struct-indirect-access (vs WasmEdge) | 85.349 (+29.09% speed) † | 80.287 (+37.23% speed) † | US | NM | NM | 110.18 (ref) † |
| threads-sustained-v2/threads/private-memory | 43.314 (+60.00% speed) † | 37.515 (+84.74% speed) † | 69.304 (ref) † | NM | NM | 42.923 (+61.46% speed) † |
| threads-sustained-v2/threads/shared-memory | 85.275 (-15.37% speed) † | 39.759 (+81.52% speed) † | 72.171 (ref) † | NM | NM | 44.706 (+61.43% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 61.100 (-7.19% speed) † | 38.468 (+47.41% speed) † | 56.708 (ref) † | NM | NM | 37.765 (+50.16% speed) † |
| threads-sustained-v2/threads/notify-empty | 34.301 (+33.08% speed) † | 31.082 (+46.86% speed) † | 45.647 (ref) | NM | NM | 44.894 (+1.68% speed) |
| threads-sustained-v2/threads/wait-mismatch | 67.096 (+14.69% speed) † | 43.150 (+78.34% speed) † | 76.952 (ref) † | NM | NM | 50.570 (+52.17% speed) † |
| threads-sustained-v2/threads/wait-timeout | 64.424 (+511.65% speed) † | 39.979 (+885.66% speed) † | 394.05 (ref) † | NM | NM | 111.85 (+252.29% speed) † |
| generic-calls-v1/core/generic-calls | 121.96 (+72.25% speed) † | 108.44 (+93.72% speed) † | 210.08 (ref) † | NM | NM | 136.85 (+53.52% speed) † |

## Complete P3 HTTP requests

P3 competitor cells marked `NM` have no measured adapter for the pinned component/WASI workload. Wasmtime uses the same validated workload. Managed allocation counts exclude foreign Rust/C/JIT allocations and cover only the managed share of allocation.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU, us/request | 74.274 (-45.23% speed) † | 206.78 (-80.33% speed) † | NM | 40.682 (ref) † | NM | NM |
| 16 B instructions/request | 710,794.91 (+218.70% Δ) † | 1,301,017.32 (+483.33% Δ) † | NM | 223,031.93 (ref) † | NM | NM |
| 16 B requests/s | 54,979.48 (-4.06% speed) † | 21,581.19 (-62.34% speed) † | NM | 57,308.28 (ref) † | NM | NM |
| 65536 B CPU, us/request | 108.17 (-44.35% speed) † | 250.68 (-75.99% speed) † | NM | 60.195 (ref) † | NM | NM |
| 65536 B instructions/request | 995,463.49 (+208.54% Δ) | 1,635,576.56 (+406.94% Δ) | NM | 322,638.02 (ref) | NM | NM |
| 65536 B requests/s | 35,448.46 (-17.32% speed) † | 16,143.48 (-62.35% speed) † | NM | 42,873.98 (ref) † | NM | NM |

### P3 source contrasts

| Metric | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| 16 B cpuUs | 1.000 [0.990, 1.012]; Δ +0.04%; speed -0.04% [-1.15%, +0.98%]; unresolved | 1.001 [0.992, 1.009]; Δ +0.11%; speed -0.11% [-0.90%, +0.78%]; pass |
| 16 B instructions | 1.001 [1.000, 1.002]; Δ +0.09%; pass | 1.000 [0.997, 1.006]; Δ +0.04%; pass |
| 16 B throughput | 0.988 [0.981, 0.994]; Δ -1.23%; speed -1.23% [-1.95%, -0.59%]; unresolved | 1.000 [0.958, 1.041]; Δ -0.05%; speed -0.05% [-4.24%, +4.12%]; unresolved |
| 16 B allocsPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 16 B bytesPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 65536 B cpuUs | 0.995 [0.977, 1.018]; Δ -0.50%; speed +0.50% [-1.74%, +2.38%]; unresolved | 1.002 [0.991, 1.012]; Δ +0.15%; speed -0.15% [-1.16%, +0.90%]; unresolved |
| 65536 B instructions | 0.999 [0.996, 1.001]; Δ -0.14%; pass | 1.002 [0.994, 1.009]; Δ +0.21%; pass |
| 65536 B throughput | 1.010 [0.992, 1.029]; Δ +1.03%; speed +1.03% [-0.82%, +2.91%]; unresolved | 0.977 [0.933, 1.021]; Δ -2.26%; speed -2.26% [-6.71%, +2.13%]; unresolved |
| 65536 B allocsPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 65536 B bytesPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |

Throughput ratios above one improve. CPU, instruction and allocation ratios below one improve. The JSON retains all own-baseline observations and every per-role A/A interval.

### Competitive P3 ratios

| Payload / compiler | CPU ratio to Wasmtime [95% CI]; controls | Throughput ratio to Wasmtime [95% CI]; controls |
| --- | ---: | ---: |
| 16 B go | 1.837 [1.768, 1.916]; Δ +83.74%; speed -45.58% [-47.80%, -43.45%]; unresolved | 0.929 [0.919, 0.939]; Δ -7.05%; speed -7.05% [-8.13%, -6.08%]; unresolved |
| 16 B llgo | 5.072 [4.723, 5.469]; Δ +407.18%; speed -80.28% [-81.71%, -78.83%]; unresolved | 0.364 [0.333, 0.391]; Δ -63.58%; speed -63.58% [-66.69%, -60.90%]; unresolved |
| 65536 B go | 1.830 [1.711, 1.965]; Δ +82.99%; speed -45.35% [-49.12%, -41.54%]; unresolved | 0.830 [0.765, 0.929]; Δ -16.96%; speed -16.96% [-23.53%, -7.06%]; unresolved |
| 65536 B llgo | 4.195 [3.853, 4.571]; Δ +319.54%; speed -76.16% [-78.12%, -74.05%]; unresolved | 0.361 [0.353, 0.374]; Δ -63.88%; speed -63.88% [-64.74%, -62.63%]; unresolved |

## Core source baseline

Head/base ratios below one improve. These marginal intervals describe this event; an apparent change in an unchanged mechanism requires diagnosis before being attributed to source changes. The machine-readable report includes every competitor ratio and A/A interval.

### native

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 1.112 [1.047, 1.191]; Δ +11.22%; speed -10.09% [-16.02%, -4.48%]; unresolved | 0.983 [0.765, 1.258]; Δ -1.74%; speed +1.77% [-20.49%, +30.69%]; unresolved |
| original22-v1/core/integer-loop | 1.002 [0.982, 1.024]; Δ +0.19%; speed -0.19% [-2.34%, +1.83%]; unresolved | 1.006 [0.990, 1.023]; Δ +0.58%; speed -0.57% [-2.22%, +1.04%]; unresolved |
| original22-v1/core/function-calls | 1.024 [1.001, 1.067]; Δ +2.44%; speed -2.38% [-6.28%, -0.14%]; unresolved | 0.987 [0.974, 1.000]; Δ -1.27%; speed +1.29% [-0.01%, +2.70%]; unresolved |
| original22-v1/core/memory-load-store | 0.984 [0.956, 1.010]; Δ -1.62%; speed +1.64% [-1.01%, +4.64%]; unresolved | 1.005 [0.995, 1.018]; Δ +0.55%; speed -0.54% [-1.75%, +0.50%]; unresolved |
| original22-v1/memory/narrow-unaligned | 1.000 [0.973, 1.030]; Δ +0.02%; speed -0.02% [-2.89%, +2.80%]; unresolved | 0.997 [0.980, 1.015]; Δ -0.29%; speed +0.29% [-1.46%, +2.03%]; unresolved |
| original22-v1/memory/dependent-address | 0.995 [0.981, 1.009]; Δ -0.52%; speed +0.52% [-0.92%, +1.92%]; unresolved | 1.010 [1.008, 1.012]; Δ +1.00%; speed -0.99% [-1.22%, -0.77%]; unresolved |
| original22-v1/memory/float64-unaligned | 0.992 [0.981, 0.999]; Δ -0.85%; speed +0.85% [+0.08%, +1.91%]; unresolved | 1.003 [0.991, 1.016]; Δ +0.31%; speed -0.30% [-1.54%, +0.92%]; unresolved |
| original22-v1/runtime/simd-memory-full | 1.007 [1.001, 1.014]; Δ +0.72%; speed -0.71% [-1.40%, -0.07%]; unresolved | 0.998 [0.989, 1.006]; Δ -0.23%; speed +0.23% [-0.59%, +1.09%]; pass |
| original22-v1/runtime/simd-memory-widen | 0.981 [0.969, 0.990]; Δ -1.90%; speed +1.94% [+1.03%, +3.24%]; unresolved | 0.997 [0.987, 1.007]; Δ -0.34%; speed +0.34% [-0.65%, +1.28%]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 0.990 [0.985, 0.997]; Δ -0.98%; speed +0.99% [+0.30%, +1.55%]; unresolved | 1.001 [0.987, 1.017]; Δ +0.12%; speed -0.12% [-1.63%, +1.32%]; pass |
| original22-v1/gc/i31-roundtrip | 0.976 [0.907, 1.081]; Δ -2.45%; speed +2.51% [-7.53%, +10.20%]; unresolved | 1.021 [0.978, 1.077]; Δ +2.12%; speed -2.07% [-7.11%, +2.29%]; unresolved |
| original22-v1/gc/struct-access | 0.982 [0.844, 1.124]; Δ -1.83%; speed +1.87% [-11.06%, +18.45%]; unresolved | 1.048 [0.947, 1.185]; Δ +4.78%; speed -4.56% [-15.61%, +5.60%]; unresolved |
| original22-v1/gc/struct-allocation | 1.086 [0.994, 1.258]; Δ +8.63%; speed -7.94% [-20.48%, +0.63%]; unresolved | 0.935 [0.878, 0.993]; Δ -6.48%; speed +6.93% [+0.73%, +13.95%]; unresolved |
| original22-v1/gc/struct-indirect-access | 0.942 [0.866, 1.022]; Δ -5.84%; speed +6.20% [-2.20%, +15.42%]; unresolved | 1.038 [0.998, 1.100]; Δ +3.76%; speed -3.63% [-9.12%, +0.19%]; unresolved |
| original22-v1/threads/private-memory | 1.017 [0.882, 1.170]; Δ +1.73%; speed -1.70% [-14.53%, +13.39%]; unresolved | 0.967 [0.923, 1.021]; Δ -3.28%; speed +3.39% [-2.07%, +8.31%]; unresolved |
| original22-v1/threads/shared-memory | 1.024 [0.921, 1.132]; Δ +2.43%; speed -2.37% [-11.62%, +8.53%]; unresolved | 0.976 [0.939, 1.011]; Δ -2.44%; speed +2.50% [-1.07%, +6.50%]; unresolved |
| original22-v1/threads/atomic-rmw | 0.991 [0.886, 1.108]; Δ -0.89%; speed +0.90% [-9.75%, +12.81%]; unresolved | 1.019 [1.002, 1.039]; Δ +1.88%; speed -1.84% [-3.73%, -0.17%]; unresolved |
| original22-v1/threads/notify-empty | 1.054 [0.935, 1.219]; Δ +5.40%; speed -5.12% [-17.98%, +6.98%]; unresolved | 1.031 [0.929, 1.186]; Δ +3.12%; speed -3.02% [-15.68%, +7.60%]; unresolved |
| original22-v1/threads/wait-mismatch | 0.968 [0.822, 1.130]; Δ -3.25%; speed +3.35% [-11.51%, +21.68%]; unresolved | 0.958 [0.906, 1.010]; Δ -4.18%; speed +4.36% [-0.98%, +10.40%]; unresolved |
| original22-v1/threads/wait-timeout | 0.960 [0.826, 1.132]; Δ -4.01%; speed +4.17% [-11.66%, +21.08%]; unresolved | 0.924 [0.811, 1.003]; Δ -7.63%; speed +8.25% [-0.33%, +23.34%]; unresolved |
| original22-v1/applications/sha256-clang | 0.891 [0.883, 0.900]; Δ -10.91%; speed +12.25% [+11.12%, +13.30%]; unresolved | 0.870 [0.858, 0.879]; Δ -13.05%; speed +15.00% [+13.80%, +16.58%]; unresolved |
| original22-v1/applications/sha256-rust | 0.872 [0.841, 0.894]; Δ -12.83%; speed +14.72% [+11.83%, +18.96%]; unresolved | 0.849 [0.781, 0.907]; Δ -15.07%; speed +17.74% [+10.21%, +27.99%]; unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 1.001 [0.986, 1.024]; Δ +0.06%; speed -0.06% [-2.37%, +1.43%]; unresolved | 0.983 [0.946, 1.006]; Δ -1.70%; speed +1.73% [-0.62%, +5.68%]; unresolved |
| gc-sustained-v1/gc/struct-access | 1.001 [0.990, 1.011]; Δ +0.07%; speed -0.07% [-1.07%, +1.06%]; unresolved | 0.997 [0.981, 1.012]; Δ -0.34%; speed +0.34% [-1.19%, +1.96%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 1.026 [0.988, 1.080]; Δ +2.65%; speed -2.58% [-7.44%, +1.25%]; unresolved | 0.989 [0.956, 1.016]; Δ -1.07%; speed +1.08% [-1.55%, +4.61%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 0.988 [0.950, 1.026]; Δ -1.15%; speed +1.16% [-2.54%, +5.21%]; unresolved | 1.028 [0.994, 1.057]; Δ +2.80%; speed -2.73% [-5.42%, +0.58%]; unresolved |
| threads-sustained-v2/threads/private-memory | 1.004 [0.932, 1.077]; Δ +0.44%; speed -0.44% [-7.18%, +7.33%]; unresolved | 0.965 [0.931, 0.992]; Δ -3.48%; speed +3.61% [+0.82%, +7.38%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 1.008 [0.978, 1.042]; Δ +0.78%; speed -0.77% [-4.01%, +2.29%]; unresolved | 1.006 [0.983, 1.037]; Δ +0.59%; speed -0.59% [-3.53%, +1.69%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.984 [0.972, 0.995]; Δ -1.63%; speed +1.66% [+0.54%, +2.89%]; unresolved | 0.999 [0.994, 1.005]; Δ -0.07%; speed +0.07% [-0.47%, +0.64%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 0.989 [0.961, 1.015]; Δ -1.14%; speed +1.16% [-1.52%, +4.00%]; unresolved | 0.976 [0.964, 0.988]; Δ -2.38%; speed +2.44% [+1.17%, +3.72%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 0.941 [0.895, 0.982]; Δ -5.95%; speed +6.32% [+1.85%, +11.74%]; unresolved | 0.988 [0.975, 1.003]; Δ -1.15%; speed +1.17% [-0.30%, +2.52%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 0.996 [0.971, 1.027]; Δ -0.41%; speed +0.42% [-2.63%, +3.02%]; unresolved | 1.022 [0.996, 1.054]; Δ +2.22%; speed -2.18% [-5.09%, +0.41%]; unresolved |
| generic-calls-v1/core/generic-calls | 1.002 [0.993, 1.012]; Δ +0.16%; speed -0.16% [-1.17%, +0.68%]; unresolved | 0.999 [0.985, 1.019]; Δ -0.07%; speed +0.07% [-1.82%, +1.54%]; unresolved |

### interpreter

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 1.125 [0.921, 1.338]; Δ +12.47%; speed -11.08% [-25.24%, +8.61%]; unresolved | 0.983 [0.895, 1.099]; Δ -1.68%; speed +1.71% [-9.00%, +11.74%]; unresolved |
| original22-v1/core/integer-loop | 0.855 [0.810, 0.881]; Δ -14.49%; speed +16.94% [+13.53%, +23.42%]; unresolved | 0.990 [0.967, 1.011]; Δ -1.03%; speed +1.04% [-1.10%, +3.44%]; unresolved |
| original22-v1/core/function-calls | 0.916 [0.904, 0.930]; Δ -8.38%; speed +9.15% [+7.57%, +10.61%]; unresolved | 0.996 [0.976, 1.012]; Δ -0.36%; speed +0.36% [-1.21%, +2.49%]; unresolved |
| original22-v1/core/memory-load-store | 0.928 [0.913, 0.944]; Δ -7.25%; speed +7.81% [+5.96%, +9.54%]; unresolved | 1.002 [0.991, 1.014]; Δ +0.22%; speed -0.21% [-1.43%, +0.95%]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.924 [0.916, 0.933]; Δ -7.56%; speed +8.18% [+7.18%, +9.12%]; unresolved | 1.014 [0.992, 1.042]; Δ +1.39%; speed -1.37% [-4.00%, +0.83%]; unresolved |
| original22-v1/memory/dependent-address | 0.890 [0.865, 0.911]; Δ -11.03%; speed +12.40% [+9.73%, +15.55%]; unresolved | 1.005 [0.995, 1.014]; Δ +0.48%; speed -0.48% [-1.39%, +0.52%]; unresolved |
| original22-v1/memory/float64-unaligned | 0.907 [0.887, 0.923]; Δ -9.26%; speed +10.20% [+8.36%, +12.73%]; unresolved | 1.006 [0.988, 1.027]; Δ +0.63%; speed -0.62% [-2.59%, +1.23%]; unresolved |
| original22-v1/runtime/simd-memory-full | 0.944 [0.927, 0.961]; Δ -5.58%; speed +5.91% [+4.04%, +7.86%]; unresolved | 1.006 [0.999, 1.014]; Δ +0.63%; speed -0.63% [-1.36%, +0.10%]; pass |
| original22-v1/runtime/simd-memory-widen | 0.941 [0.930, 0.950]; Δ -5.88%; speed +6.24% [+5.21%, +7.51%]; pass | 1.014 [1.006, 1.023]; Δ +1.38%; speed -1.36% [-2.26%, -0.61%]; pass |
| original22-v1/runtime/simd-memory-lanes | 0.956 [0.946, 0.967]; Δ -4.37%; speed +4.57% [+3.39%, +5.67%]; unresolved | 1.007 [0.995, 1.026]; Δ +0.75%; speed -0.74% [-2.57%, +0.52%]; unresolved |
| original22-v1/gc/i31-roundtrip | 0.890 [0.865, 0.912]; Δ -11.04%; speed +12.41% [+9.66%, +15.56%]; unresolved | 0.986 [0.953, 1.017]; Δ -1.36%; speed +1.38% [-1.68%, +4.91%]; unresolved |
| original22-v1/gc/struct-access | 0.899 [0.885, 0.918]; Δ -10.09%; speed +11.22% [+8.91%, +12.95%]; unresolved | 1.003 [0.975, 1.022]; Δ +0.34%; speed -0.34% [-2.12%, +2.57%]; unresolved |
| original22-v1/gc/struct-allocation | 0.958 [0.931, 0.978]; Δ -4.20%; speed +4.38% [+2.26%, +7.38%]; unresolved | 0.895 [0.827, 0.961]; Δ -10.51%; speed +11.74% [+4.01%, +20.88%]; unresolved |
| original22-v1/gc/struct-indirect-access | 0.904 [0.878, 0.929]; Δ -9.61%; speed +10.63% [+7.68%, +13.94%]; unresolved | 1.027 [0.987, 1.096]; Δ +2.65%; speed -2.58% [-8.77%, +1.31%]; unresolved |
| original22-v1/threads/private-memory | 0.960 [0.925, 1.007]; Δ -3.99%; speed +4.15% [-0.74%, +8.07%]; unresolved | 1.005 [0.979, 1.031]; Δ +0.54%; speed -0.53% [-3.04%, +2.15%]; unresolved |
| original22-v1/threads/shared-memory | 0.956 [0.946, 0.963]; Δ -4.44%; speed +4.65% [+3.89%, +5.75%]; unresolved | 0.993 [0.967, 1.020]; Δ -0.68%; speed +0.69% [-1.99%, +3.46%]; unresolved |
| original22-v1/threads/atomic-rmw | 0.987 [0.967, 1.018]; Δ -1.31%; speed +1.32% [-1.81%, +3.36%]; unresolved | 1.006 [0.981, 1.041]; Δ +0.62%; speed -0.61% [-3.95%, +1.90%]; unresolved |
| original22-v1/threads/notify-empty | 0.933 [0.911, 0.954]; Δ -6.72%; speed +7.20% [+4.86%, +9.72%]; unresolved | 0.992 [0.973, 1.009]; Δ -0.83%; speed +0.84% [-0.92%, +2.80%]; unresolved |
| original22-v1/threads/wait-mismatch | 0.914 [0.809, 0.991]; Δ -8.62%; speed +9.44% [+0.91%, +23.61%]; unresolved | 1.022 [0.976, 1.096]; Δ +2.21%; speed -2.16% [-8.77%, +2.47%]; unresolved |
| original22-v1/threads/wait-timeout | 0.952 [0.920, 1.005]; Δ -4.76%; speed +4.99% [-0.45%, +8.68%]; unresolved | 0.993 [0.977, 1.009]; Δ -0.71%; speed +0.71% [-0.88%, +2.40%]; unresolved |
| original22-v1/applications/sha256-clang | 0.954 [0.942, 0.966]; Δ -4.56%; speed +4.78% [+3.53%, +6.13%]; unresolved | 1.004 [0.980, 1.029]; Δ +0.37%; speed -0.37% [-2.84%, +2.00%]; unresolved |
| original22-v1/applications/sha256-rust | 0.925 [0.894, 0.946]; Δ -7.52%; speed +8.13% [+5.74%, +11.91%]; unresolved | 0.999 [0.987, 1.011]; Δ -0.12%; speed +0.12% [-1.13%, +1.35%]; unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 0.903 [0.887, 0.916]; Δ -9.71%; speed +10.75% [+9.16%, +12.76%]; unresolved | 0.992 [0.979, 1.004]; Δ -0.81%; speed +0.82% [-0.39%, +2.12%]; unresolved |
| gc-sustained-v1/gc/struct-access | 0.896 [0.876, 0.908]; Δ -10.38%; speed +11.58% [+10.19%, +14.09%]; unresolved | 0.989 [0.964, 1.009]; Δ -1.11%; speed +1.12% [-0.92%, +3.73%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 0.958 [0.923, 0.989]; Δ -4.19%; speed +4.37% [+1.13%, +8.39%]; unresolved | 1.035 [0.921, 1.157]; Δ +3.55%; speed -3.43% [-13.54%, +8.58%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 0.878 [0.804, 0.922]; Δ -12.20%; speed +13.90% [+8.50%, +24.35%]; unresolved | 0.987 [0.972, 1.001]; Δ -1.35%; speed +1.36% [-0.06%, +2.93%]; unresolved |
| threads-sustained-v2/threads/private-memory | 0.933 [0.926, 0.939]; Δ -6.72%; speed +7.20% [+6.47%, +8.05%]; unresolved | 1.009 [0.993, 1.038]; Δ +0.88%; speed -0.87% [-3.67%, +0.74%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 0.949 [0.932, 0.965]; Δ -5.14%; speed +5.42% [+3.64%, +7.32%]; unresolved | 0.983 [0.955, 1.006]; Δ -1.72%; speed +1.75% [-0.60%, +4.67%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.965 [0.956, 0.974]; Δ -3.48%; speed +3.61% [+2.69%, +4.60%]; unresolved | 0.991 [0.962, 1.012]; Δ -0.93%; speed +0.94% [-1.23%, +3.99%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 0.940 [0.934, 0.946]; Δ -6.05%; speed +6.44% [+5.65%, +7.05%]; unresolved | 1.001 [0.986, 1.012]; Δ +0.13%; speed -0.13% [-1.17%, +1.42%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 0.948 [0.925, 0.973]; Δ -5.16%; speed +5.44% [+2.78%, +8.11%]; unresolved | 1.002 [0.991, 1.017]; Δ +0.18%; speed -0.18% [-1.63%, +0.93%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 0.955 [0.951, 0.960]; Δ -4.47%; speed +4.68% [+4.19%, +5.20%]; unresolved | 0.999 [0.980, 1.020]; Δ -0.06%; speed +0.06% [-1.96%, +2.03%]; unresolved |
| generic-calls-v1/core/generic-calls | 0.900 [0.886, 0.914]; Δ -10.00%; speed +11.11% [+9.37%, +12.86%]; unresolved | 1.023 [1.011, 1.037]; Δ +2.30%; speed -2.24% [-3.59%, -1.09%]; unresolved |

## Preparation, first call and cleanup

All entries below are **microseconds per process or invocation**, including unresolved observations. Phase intervals are not inferred from warm-call controls. Startup is independently observed process readiness. Native compile/validate boundaries differ by API; WasmEdge AOT includes validation and artifact I/O, plus separately recorded native-image loading. Every process creates a fresh native image. First-call and warm execution have their own boundaries; values are never subtracted to invent pure costs.

<details>
<summary>Native: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,969.90 | 8,184.92 | 6,301.85 | 5,750.31 | 7,099.81 | 19,293.58 |
| original22-v1/core/integer-loop | 5,257.81 | 7,094.44 | 6,097.94 | 4,686.42 | 7,032.85 | 20,711.48 |
| original22-v1/core/function-calls | 5,325.67 | 7,915.42 | 5,023.77 | 5,021.50 | 7,016.35 | 19,136.10 |
| original22-v1/core/memory-load-store | 4,472.33 | 6,880.52 | 4,550.04 | 4,647.08 | 6,466.60 | 18,965.48 |
| original22-v1/memory/narrow-unaligned | 5,133.83 | 7,823.79 | 5,134.29 | 5,261.69 | 7,118.77 | 18,489.25 |
| original22-v1/memory/dependent-address | 5,048.00 | 7,003.08 | 5,541.54 | 5,745.42 | 7,798.33 | 18,914.79 |
| original22-v1/memory/float64-unaligned | 4,795.17 | 6,982.19 | 4,685.63 | 4,661.33 | 6,384.77 | 18,566.73 |
| original22-v1/runtime/simd-memory-full | 5,524.17 | 7,447.17 | 4,810.29 | 4,966.08 | 7,087.02 | 19,207.81 |
| original22-v1/runtime/simd-memory-widen | 5,037.52 | 7,704.15 | 4,782.65 | 4,707.27 | 6,644.98 | 18,835.29 |
| original22-v1/runtime/simd-memory-lanes | 5,237.08 | 7,003.31 | 5,013.81 | 4,882.52 | 7,141.46 | 19,412.62 |
| original22-v1/gc/i31-roundtrip | 5,508.21 | 6,666.83 | US | 5,678.23 | US | 19,389.96 |
| original22-v1/gc/struct-access | 4,923.83 | 6,987.12 | US | 5,326.81 | US | 19,145.08 |
| original22-v1/gc/struct-allocation | 5,080.27 | 7,217.96 | US | 4,841.25 | US | 18,497.98 |
| original22-v1/gc/struct-indirect-access | 4,543.94 | 6,828.19 | US | 5,171.81 | US | 19,017.94 |
| original22-v1/threads/private-memory | 4,632.81 | 6,542.29 | 4,568.27 | 4,742.00 | 6,857.96 | 17,291.44 |
| original22-v1/threads/shared-memory | 5,296.40 | 7,170.77 | 5,840.56 | 5,295.21 | 6,720.21 | 18,959.75 |
| original22-v1/threads/atomic-rmw | 4,575.56 | 6,620.06 | 4,768.23 | 4,709.31 | 7,231.33 | 18,901.92 |
| original22-v1/threads/notify-empty | 5,422.98 | 7,272.08 | 4,750.92 | 4,716.21 | 6,592.50 | 18,280.23 |
| original22-v1/threads/wait-mismatch | 6,239.33 | 8,377.98 | 6,244.27 | 6,875.71 | 7,636.77 | 21,061.27 |
| original22-v1/threads/wait-timeout | 5,694.00 | 7,780.15 | 5,793.56 | 6,811.96 | 6,595.00 | 17,854.19 |
| original22-v1/applications/sha256-clang | 4,795.60 | 6,979.69 | 4,662.00 | 4,876.44 | 6,735.50 | 17,986.54 |
| original22-v1/applications/sha256-rust | 4,881.46 | 7,204.77 | 5,242.69 | 5,114.69 | 6,934.58 | 18,069.35 |
| gc-sustained-v1/gc/i31-roundtrip | 4,585.50 | 6,925.12 | US | 5,101.50 | US | 18,473.25 |
| gc-sustained-v1/gc/struct-access | 6,426.08 | 7,238.83 | US | 4,875.69 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 5,158.94 | 7,247.19 | US | 5,102.44 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 5,726.56 | 8,151.56 | US | 5,665.67 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 5,118.98 | 7,046.60 | 4,923.21 | 4,800.17 | 7,222.15 | 17,967.06 |
| threads-sustained-v2/threads/shared-memory | 4,934.23 | 7,807.48 | 5,147.25 | 4,850.60 | 7,293.29 | 20,131.31 |
| threads-sustained-v2/threads/atomic-rmw | 6,381.38 | 9,400.23 | 5,565.12 | 6,160.52 | 7,102.71 | 20,793.08 |
| threads-sustained-v2/threads/notify-empty | 4,623.85 | 6,780.69 | 4,954.44 | 4,993.12 | 6,545.56 | 18,198.33 |
| threads-sustained-v2/threads/wait-mismatch | 4,731.85 | 7,829.77 | 4,875.77 | 5,404.40 | 6,778.29 | 21,311.81 |
| threads-sustained-v2/threads/wait-timeout | 5,543.71 | 7,278.10 | 6,013.35 | 4,922.10 | 7,486.52 | 19,283.23 |
| generic-calls-v1/core/generic-calls | 5,380.83 | 8,258.17 | 5,078.90 | 5,131.31 | 7,100.88 | 18,184.33 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 49.042 | 42.500 | 40.666 | 45.312 | 36.291 | 47.562 |
| original22-v1/core/integer-loop | 47.729 | 46.062 | 47.230 | 41.188 | 43.229 | 49.980 |
| original22-v1/core/function-calls | 46.312 | 42.272 | 46.938 | 37.875 | 40.145 | 51.354 |
| original22-v1/core/memory-load-store | 46.646 | 44.979 | 51.334 | 39.709 | 39.062 | 49.687 |
| original22-v1/memory/narrow-unaligned | 54.542 | 47.521 | 46.416 | 37.520 | 46.188 | 51.104 |
| original22-v1/memory/dependent-address | 44.084 | 48.084 | 46.062 | 33.312 | 52.624 | 54.916 |
| original22-v1/memory/float64-unaligned | 44.771 | 55.395 | 48.666 | 33.876 | 41.542 | 48.333 |
| original22-v1/runtime/simd-memory-full | 44.083 | 49.916 | 51.312 | 42.772 | 50.584 | 48.500 |
| original22-v1/runtime/simd-memory-widen | 44.375 | 45.479 | 50.896 | 36.854 | 49.730 | 47.417 |
| original22-v1/runtime/simd-memory-lanes | 50.376 | 49.355 | 45.709 | 37.666 | 46.979 | 59.750 |
| original22-v1/gc/i31-roundtrip | 40.876 | 42.791 | US | 29.875 | US | 41.959 |
| original22-v1/gc/struct-access | 41.416 | 42.750 | US | 39.875 | US | 43.125 |
| original22-v1/gc/struct-allocation | 45.459 | 50.333 | US | 35.250 | US | 44.188 |
| original22-v1/gc/struct-indirect-access | 43.375 | 44.687 | US | 43.729 | US | 48.792 |
| original22-v1/threads/private-memory | 37.604 | 40.709 | 35.770 | 35.834 | 34.812 | 45.041 |
| original22-v1/threads/shared-memory | 39.959 | 42.416 | 40.291 | 36.166 | 41.625 | 42.812 |
| original22-v1/threads/atomic-rmw | 32.251 | 41.250 | 39.479 | 37.979 | 37.084 | 45.667 |
| original22-v1/threads/notify-empty | 41.084 | 44.001 | 43.020 | 37.104 | 36.125 | 45.584 |
| original22-v1/threads/wait-mismatch | 55.833 | 46.376 | 57.770 | 33.230 | 44.041 | 52.272 |
| original22-v1/threads/wait-timeout | 42.834 | 55.062 | 37.959 | 40.500 | 38.292 | 44.063 |
| original22-v1/applications/sha256-clang | 49.791 | 45.688 | 44.250 | 35.520 | 43.270 | 55.604 |
| original22-v1/applications/sha256-rust | 47.229 | 55.062 | 45.895 | 35.375 | 47.395 | 49.937 |
| gc-sustained-v1/gc/i31-roundtrip | 43.875 | 45.584 | US | 44.730 | US | 49.855 |
| gc-sustained-v1/gc/struct-access | 47.750 | 45.042 | US | 35.896 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 52.104 | 51.625 | US | 44.021 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 43.895 | 52.146 | US | 40.688 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 40.438 | 46.229 | 41.708 | 34.416 | 37.270 | 45.688 |
| threads-sustained-v2/threads/shared-memory | 42.146 | 39.708 | 42.521 | 34.625 | 51.730 | 49.938 |
| threads-sustained-v2/threads/atomic-rmw | 52.958 | 39.750 | 38.834 | 39.041 | 36.062 | 49.230 |
| threads-sustained-v2/threads/notify-empty | 40.438 | 40.375 | 38.812 | 34.105 | 35.729 | 41.228 |
| threads-sustained-v2/threads/wait-mismatch | 37.605 | 46.876 | 43.250 | 36.583 | 39.062 | 63.334 |
| threads-sustained-v2/threads/wait-timeout | 40.041 | 44.000 | 42.541 | 32.876 | 36.230 | 51.749 |
| generic-calls-v1/core/generic-calls | 39.333 | 50.541 | 41.250 | 38.875 | 44.521 | 49.520 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.500 | 6.042 | 65.459 | 136.50 | 45.291 | 30.729 |
| original22-v1/core/integer-loop | 7.562 | 5.771 | 64.937 | 130.27 | 45.312 | 29.354 |
| original22-v1/core/function-calls | 6.333 | 5.875 | 64.646 | 139.62 | 51.562 | 30.688 |
| original22-v1/core/memory-load-store | 6.312 | 5.812 | 60.750 | 138.56 | 43.938 | 30.729 |
| original22-v1/memory/narrow-unaligned | 8.042 | 6.042 | 68.500 | 141.29 | 47.938 | 32.896 |
| original22-v1/memory/dependent-address | 6.688 | 5.792 | 65.541 | 140.85 | 53.395 | 30.958 |
| original22-v1/memory/float64-unaligned | 7.083 | 5.917 | 61.354 | 128.33 | 46.063 | 29.917 |
| original22-v1/runtime/simd-memory-full | 6.333 | 5.833 | 66.750 | 135.96 | 49.084 | 31.916 |
| original22-v1/runtime/simd-memory-widen | 6.104 | 5.542 | 65.854 | 134.60 | 47.584 | 31.041 |
| original22-v1/runtime/simd-memory-lanes | 7.125 | 5.917 | 65.188 | 138.67 | 49.834 | 33.812 |
| original22-v1/gc/i31-roundtrip | 7.104 | 6.042 | US | 136.27 | US | 28.875 |
| original22-v1/gc/struct-access | 6.562 | 5.833 | US | 144.50 | US | 33.291 |
| original22-v1/gc/struct-allocation | 7.167 | 6.250 | US | 139.48 | US | 27.604 |
| original22-v1/gc/struct-indirect-access | 5.854 | 5.812 | US | 145.81 | US | 33.833 |
| original22-v1/threads/private-memory | 6.646 | 5.771 | 61.479 | 133.75 | 44.270 | 31.688 |
| original22-v1/threads/shared-memory | 6.583 | 5.771 | 64.521 | 140.81 | 49.250 | 32.812 |
| original22-v1/threads/atomic-rmw | 6.292 | 5.708 | 65.646 | 140.62 | 46.916 | 34.604 |
| original22-v1/threads/notify-empty | 6.750 | 5.583 | 63.750 | 143.62 | 45.166 | 28.917 |
| original22-v1/threads/wait-mismatch | 7.312 | 5.354 | 67.625 | 139.50 | 49.166 | 32.854 |
| original22-v1/threads/wait-timeout | 6.688 | 6.417 | 60.709 | 142.54 | 46.208 | 31.583 |
| original22-v1/applications/sha256-clang | 6.646 | 5.583 | 63.541 | 137.54 | 48.209 | 31.584 |
| original22-v1/applications/sha256-rust | 6.521 | 6.062 | 63.541 | 139.58 | 47.646 | 29.521 |
| gc-sustained-v1/gc/i31-roundtrip | 6.438 | 5.375 | US | 134.42 | US | 28.605 |
| gc-sustained-v1/gc/struct-access | 7.000 | 6.021 | US | 144.52 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 7.021 | 6.042 | US | 136.04 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 6.812 | 5.729 | US | 141.73 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 6.396 | 5.625 | 62.896 | 132.88 | 47.542 | 30.833 |
| threads-sustained-v2/threads/shared-memory | 6.292 | 5.938 | 62.459 | 135.23 | 45.646 | 29.500 |
| threads-sustained-v2/threads/atomic-rmw | 6.750 | 6.042 | 62.917 | 140.67 | 46.834 | 30.354 |
| threads-sustained-v2/threads/notify-empty | 6.125 | 5.729 | 59.895 | 139.08 | 46.042 | 28.020 |
| threads-sustained-v2/threads/wait-mismatch | 6.292 | 6.271 | 64.479 | 139.44 | 48.895 | 35.333 |
| threads-sustained-v2/threads/wait-timeout | 5.938 | 5.833 | 65.084 | 137.65 | 46.854 | 29.271 |
| generic-calls-v1/core/generic-calls | 6.167 | 5.938 | 64.334 | 143.31 | 48.084 | 31.209 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 213.69 | 164.25 | 108.75 | 811.10 | 1,298.54 | 280,236.81 |
| original22-v1/core/integer-loop | 267.98 | 214.21 | 134.75 | 1,104.75 | 1,183.12 | 274,324.58 |
| original22-v1/core/function-calls | 317.88 | 286.40 | 142.81 | 1,134.48 | 1,291.65 | 271,686.08 |
| original22-v1/core/memory-load-store | 341.60 | 318.79 | 146.29 | 926.71 | 1,264.19 | 273,438.62 |
| original22-v1/memory/narrow-unaligned | 350.31 | 318.29 | 162.04 | 924.37 | 1,200.56 | 273,065.67 |
| original22-v1/memory/dependent-address | 332.23 | 293.79 | 167.60 | 864.90 | 1,320.54 | 271,994.23 |
| original22-v1/memory/float64-unaligned | 349.90 | 332.19 | 153.83 | 853.19 | 1,245.06 | 273,815.73 |
| original22-v1/runtime/simd-memory-full | 314.19 | 276.77 | 134.17 | 1,205.25 | 1,277.10 | 275,969.21 |
| original22-v1/runtime/simd-memory-widen | 287.46 | 266.38 | 144.58 | 1,349.92 | 1,268.08 | 271,772.42 |
| original22-v1/runtime/simd-memory-lanes | 316.73 | 286.44 | 141.06 | 1,132.48 | 1,228.56 | 272,672.29 |
| original22-v1/gc/i31-roundtrip | 230.54 | 177.40 | US | 1,050.02 | US | 272,444.31 |
| original22-v1/gc/struct-access | 264.42 | 223.21 | US | 1,127.81 | US | 276,469.37 |
| original22-v1/gc/struct-allocation | 248.88 | 212.10 | US | 1,272.50 | US | 273,821.56 |
| original22-v1/gc/struct-indirect-access | 310.85 | 262.44 | US | 1,795.98 | US | 277,429.92 |
| original22-v1/threads/private-memory | 247.56 | 204.50 | 128.56 | 861.50 | 1,147.25 | 276,582.88 |
| original22-v1/threads/shared-memory | 266.54 | 214.83 | 148.85 | 1,065.29 | 1,173.98 | 275,024.27 |
| original22-v1/threads/atomic-rmw | 234.23 | 189.54 | 135.81 | 918.42 | 1,179.40 | 274,924.19 |
| original22-v1/threads/notify-empty | 244.00 | 197.60 | 136.85 | 1,046.48 | 1,174.85 | 272,971.75 |
| original22-v1/threads/wait-mismatch | 247.42 | 211.52 | 163.83 | 1,121.33 | 1,296.50 | 279,877.71 |
| original22-v1/threads/wait-timeout | 239.21 | 217.10 | 142.71 | 1,104.69 | 1,184.60 | 272,410.62 |
| original22-v1/applications/sha256-clang | 1,503.17 | 1,903.56 | 339.02 | 1,408.27 | 1,867.75 | 279,606.67 |
| original22-v1/applications/sha256-rust | 1,552.92 | 1,990.69 | 364.54 | 1,650.50 | 2,048.81 | 284,468.44 |
| gc-sustained-v1/gc/i31-roundtrip | 219.02 | 171.31 | US | 810.21 | US | 271,251.21 |
| gc-sustained-v1/gc/struct-access | 265.19 | 238.21 | US | 1,076.21 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 247.02 | 211.52 | US | 2,074.48 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 307.21 | 290.02 | US | 1,851.33 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 240.40 | 202.19 | 136.17 | 807.17 | 1,155.52 | 270,048.29 |
| threads-sustained-v2/threads/shared-memory | 263.06 | 213.65 | 135.90 | 968.44 | 1,168.21 | 272,289.65 |
| threads-sustained-v2/threads/atomic-rmw | 240.96 | 197.73 | 137.50 | 1,056.65 | 1,181.73 | 269,930.85 |
| threads-sustained-v2/threads/notify-empty | 237.23 | 196.52 | 131.85 | 1,020.65 | 1,229.69 | 268,025.96 |
| threads-sustained-v2/threads/wait-mismatch | 242.73 | 208.25 | 140.56 | 1,045.98 | 1,183.56 | 275,607.83 |
| threads-sustained-v2/threads/wait-timeout | 238.50 | 202.46 | 143.77 | 969.56 | 1,256.42 | 280,070.31 |
| generic-calls-v1/core/generic-calls | 327.31 | 289.50 | 135.23 | 1,021.06 | 1,292.90 | 273,384.17 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 9.875 | 9.188 | 15.729 | 8.834 | 19.770 | 16.438 |
| original22-v1/core/integer-loop | 8.521 | 9.000 | 15.791 | 8.292 | 20.520 | 16.646 |
| original22-v1/core/function-calls | 9.750 | 9.083 | 15.750 | 8.729 | 18.812 | 16.855 |
| original22-v1/core/memory-load-store | 13.104 | 14.521 | 19.834 | 22.938 | 33.958 | 39.249 |
| original22-v1/memory/narrow-unaligned | 15.021 | 14.500 | 20.729 | 24.084 | 34.520 | 37.104 |
| original22-v1/memory/dependent-address | 15.729 | 14.812 | 20.812 | 23.416 | 35.083 | 37.688 |
| original22-v1/memory/float64-unaligned | 14.875 | 15.084 | 19.291 | 21.979 | 35.875 | 36.208 |
| original22-v1/runtime/simd-memory-full | 17.083 | 16.896 | 18.438 | 63.166 | 39.584 | 46.334 |
| original22-v1/runtime/simd-memory-widen | 17.125 | 15.916 | 19.333 | 62.042 | 38.166 | 45.229 |
| original22-v1/runtime/simd-memory-lanes | 18.229 | 16.312 | 20.041 | 63.416 | 40.500 | 45.687 |
| original22-v1/gc/i31-roundtrip | 9.166 | 10.312 | US | 8.916 | US | 15.354 |
| original22-v1/gc/struct-access | 10.479 | 10.375 | US | 25.688 | US | 17.312 |
| original22-v1/gc/struct-allocation | 11.459 | 11.604 | US | 58.355 | US | 21.792 |
| original22-v1/gc/struct-indirect-access | 22.625 | 23.187 | US | 75.979 | US | 23.812 |
| original22-v1/threads/private-memory | 14.542 | 14.999 | 18.771 | 22.250 | 33.855 | 38.772 |
| original22-v1/threads/shared-memory | 15.354 | 15.229 | 20.730 | 23.605 | 39.291 | 40.230 |
| original22-v1/threads/atomic-rmw | 14.291 | 15.375 | 19.708 | 24.562 | 39.938 | 37.354 |
| original22-v1/threads/notify-empty | 14.188 | 14.646 | 17.812 | 22.625 | 40.499 | 36.334 |
| original22-v1/threads/wait-mismatch | 16.334 | 16.459 | 22.520 | 22.834 | 38.646 | 38.105 |
| original22-v1/threads/wait-timeout | 15.312 | 16.291 | 19.541 | 24.584 | 40.125 | 38.188 |
| original22-v1/applications/sha256-clang | 20.709 | 19.646 | 17.521 | 61.438 | 40.959 | 45.291 |
| original22-v1/applications/sha256-rust | 19.875 | 19.541 | 22.520 | 63.542 | 42.104 | 46.125 |
| gc-sustained-v1/gc/i31-roundtrip | 9.209 | 10.021 | US | 8.354 | US | 16.000 |
| gc-sustained-v1/gc/struct-access | 10.792 | 10.583 | US | 25.479 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 11.249 | 11.396 | US | 57.396 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 23.084 | 24.959 | US | 70.833 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 14.354 | 14.251 | 17.917 | 22.458 | 33.229 | 37.292 |
| threads-sustained-v2/threads/shared-memory | 14.917 | 14.334 | 19.562 | 23.730 | 39.771 | 37.438 |
| threads-sustained-v2/threads/atomic-rmw | 15.083 | 17.062 | 18.270 | 25.500 | 40.709 | 39.666 |
| threads-sustained-v2/threads/notify-empty | 14.521 | 14.687 | 17.250 | 22.604 | 39.584 | 38.354 |
| threads-sustained-v2/threads/wait-mismatch | 16.084 | 16.500 | 21.000 | 23.334 | 39.750 | 39.209 |
| threads-sustained-v2/threads/wait-timeout | 15.291 | 15.604 | 18.125 | 24.312 | 42.396 | 38.584 |
| generic-calls-v1/core/generic-calls | 10.396 | 10.812 | 15.166 | 8.688 | 19.417 | 14.646 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.250 | 0.396 | 2.146 | 4.042 | 5.083 | 0.292 |
| original22-v1/core/integer-loop | 0.250 | 0.375 | 2.396 | 4.229 | 5.229 | 0.292 |
| original22-v1/core/function-calls | 0.292 | 0.375 | 2.042 | 4.292 | 4.938 | 0.230 |
| original22-v1/core/memory-load-store | 0.250 | 0.458 | 2.333 | 4.417 | 4.938 | 0.271 |
| original22-v1/memory/narrow-unaligned | 0.292 | 0.459 | 2.062 | 4.750 | 4.854 | 0.250 |
| original22-v1/memory/dependent-address | 0.292 | 0.438 | 2.292 | 4.312 | 4.938 | 0.312 |
| original22-v1/memory/float64-unaligned | 0.271 | 0.417 | 2.417 | 4.104 | 5.166 | 0.251 |
| original22-v1/runtime/simd-memory-full | 0.292 | 0.479 | 2.042 | 2.188 | 5.083 | 0.230 |
| original22-v1/runtime/simd-memory-widen | 0.271 | 0.479 | 2.521 | 2.062 | 4.833 | 0.208 |
| original22-v1/runtime/simd-memory-lanes | 0.291 | 0.417 | 2.208 | 2.229 | 5.062 | 0.250 |
| original22-v1/gc/i31-roundtrip | 0.250 | 0.417 | US | 4.292 | US | 0.292 |
| original22-v1/gc/struct-access | 0.250 | 0.375 | US | 4.542 | US | 0.271 |
| original22-v1/gc/struct-allocation | 0.250 | 0.396 | US | 2.000 | US | 0.230 |
| original22-v1/gc/struct-indirect-access | 0.271 | 0.417 | US | 2.124 | US | 0.208 |
| original22-v1/threads/private-memory | 0.250 | 0.459 | 2.167 | 4.167 | 5.167 | 0.229 |
| original22-v1/threads/shared-memory | 0.292 | 0.458 | 3.271 | 4.833 | 5.083 | 0.291 |
| original22-v1/threads/atomic-rmw | 0.250 | 0.417 | 2.438 | 4.542 | 5.083 | 0.292 |
| original22-v1/threads/notify-empty | 0.292 | 0.438 | 2.000 | 4.229 | 5.208 | 0.250 |
| original22-v1/threads/wait-mismatch | 0.292 | 0.438 | 3.854 | 4.417 | 4.917 | 0.250 |
| original22-v1/threads/wait-timeout | 0.271 | 0.438 | 2.813 | 4.250 | 5.021 | 0.208 |
| original22-v1/applications/sha256-clang | 0.271 | 0.417 | 2.646 | 2.188 | 5.646 | 0.271 |
| original22-v1/applications/sha256-rust | 0.292 | 0.416 | 2.708 | 2.146 | 5.437 | 0.271 |
| gc-sustained-v1/gc/i31-roundtrip | 0.291 | 0.375 | US | 4.188 | US | 0.271 |
| gc-sustained-v1/gc/struct-access | 0.250 | 0.458 | US | 4.250 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 0.250 | 0.354 | US | 2.104 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 0.292 | 0.417 | US | 2.249 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.291 | 0.417 | 2.208 | 4.146 | 5.000 | 0.250 |
| threads-sustained-v2/threads/shared-memory | 0.292 | 0.416 | 2.979 | 4.479 | 5.292 | 0.291 |
| threads-sustained-v2/threads/atomic-rmw | 0.292 | 0.479 | 2.229 | 4.708 | 5.271 | 0.230 |
| threads-sustained-v2/threads/notify-empty | 0.291 | 0.438 | 2.271 | 4.354 | 5.062 | 0.291 |
| threads-sustained-v2/threads/wait-mismatch | 0.271 | 0.458 | 2.854 | 4.479 | 5.208 | 0.250 |
| threads-sustained-v2/threads/wait-timeout | 0.291 | 0.458 | 2.812 | 4.354 | 5.208 | 0.271 |
| generic-calls-v1/core/generic-calls | 0.250 | 0.375 | 2.479 | 4.417 | 4.917 | 0.271 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 11.562 | 8.917 | 9.771 | 30.688 | 15.251 | 17.250 |
| original22-v1/core/integer-loop | 1,837.44 | 1,831.06 | 1,829.35 | 1,826.25 | 1,838.98 | 1,835.75 |
| original22-v1/core/function-calls | 884.75 | 904.92 | 1,534.88 | 1,045.88 | 926.19 | 891.29 |
| original22-v1/core/memory-load-store | 497.27 | 489.21 | 889.02 | 565.15 | 594.00 | 312.08 |
| original22-v1/memory/narrow-unaligned | 725.85 | 755.85 | 1,075.33 | 830.48 | 815.98 | 604.96 |
| original22-v1/memory/dependent-address | 2,090.85 | 2,170.75 | 3,360.60 | 2,352.60 | 2,415.40 | 1,651.15 |
| original22-v1/memory/float64-unaligned | 818.00 | 813.27 | 863.02 | 828.85 | 798.10 | 726.25 |
| original22-v1/runtime/simd-memory-full | 3,775.54 | 3,777.79 | 4,024.40 | 4,082.58 | 4,056.56 | 4,026.38 |
| original22-v1/runtime/simd-memory-widen | 1,200.46 | 1,193.17 | 1,197.42 | 1,158.12 | 1,169.04 | 1,147.38 |
| original22-v1/runtime/simd-memory-lanes | 5,807.02 | 5,640.58 | 6,710.85 | 6,663.75 | 6,854.25 | 5,560.94 |
| original22-v1/gc/i31-roundtrip | 27.145 | 23.959 | US | 53.791 | US | 30.646 |
| original22-v1/gc/struct-access | 34.083 | 33.563 | US | 61.541 | US | 143.44 |
| original22-v1/gc/struct-allocation | 407.79 | 430.00 | US | 159.52 | US | 554.67 |
| original22-v1/gc/struct-indirect-access | 30.750 | 27.230 | US | 34.084 | US | 341.04 |
| original22-v1/threads/private-memory | 15.167 | 13.250 | 20.812 | 46.292 | 28.625 | 33.812 |
| original22-v1/threads/shared-memory | 29.584 | 24.541 | 20.771 | 49.917 | 29.271 | 35.354 |
| original22-v1/threads/atomic-rmw | 35.500 | 30.166 | 30.375 | 59.625 | 59.500 | 41.271 |
| original22-v1/threads/notify-empty | 24.459 | 20.541 | 250.73 | 111.79 | 68.000 | 119.69 |
| original22-v1/threads/wait-mismatch | 27.438 | 23.396 | 302.21 | 138.17 | 42.062 | 57.792 |
| original22-v1/threads/wait-timeout | 24.709 | 20.395 | 2,892.31 | 502.67 | 2,014.15 | 720.94 |
| original22-v1/applications/sha256-clang | 1,178.65 | 1,099.00 | 1,437.46 | 1,048.50 | 1,069.33 | 978.58 |
| original22-v1/applications/sha256-rust | 1,220.15 | 1,142.42 | 1,651.73 | 1,049.79 | 1,058.67 | 1,001.79 |
| gc-sustained-v1/gc/i31-roundtrip | 1,485.58 | 1,517.38 | US | 1,981.10 | US | 1,410.94 |
| gc-sustained-v1/gc/struct-access | 1,229.85 | 1,188.50 | US | 1,210.60 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 30,881.31 | 18,635.73 | US | 3,499.48 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1,813.23 | 1,837.92 | US | 3,035.17 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 42.688 | 38.438 | 124.50 | 141.94 | 122.00 | 137.33 |
| threads-sustained-v2/threads/shared-memory | 136.46 | 151.00 | 131.08 | 151.58 | 135.73 | 136.88 |
| threads-sustained-v2/threads/atomic-rmw | 222.04 | 222.15 | 220.77 | 247.52 | 460.92 | 232.12 |
| threads-sustained-v2/threads/notify-empty | 112.50 | 106.15 | 2,359.98 | 788.00 | 528.33 | 1,005.83 |
| threads-sustained-v2/threads/wait-mismatch | 136.56 | 130.52 | 2,805.15 | 1,071.21 | 289.73 | 388.56 |
| threads-sustained-v2/threads/wait-timeout | 116.02 | 109.81 | 31,992.02 | 4,911.56 | 20,334.48 | 6,921.25 |
| generic-calls-v1/core/generic-calls | 5,973.23 | 5,667.42 | 11,493.60 | 11,472.50 | 11,410.23 | 5,708.33 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.896 | 4.125 | 4.125 | 56.812 | 30.000 | 359.44 |
| original22-v1/core/integer-loop | 6.229 | 6.375 | 10.750 | 70.001 | 36.021 | 413.94 |
| original22-v1/core/function-calls | 14.001 | 12.479 | 12.188 | 102.25 | 45.291 | 433.10 |
| original22-v1/core/memory-load-store | 13.438 | 11.126 | 4.416 | 83.709 | 60.083 | 441.15 |
| original22-v1/memory/narrow-unaligned | 18.604 | 14.166 | 4.833 | 91.770 | 62.416 | 465.90 |
| original22-v1/memory/dependent-address | 10.688 | 15.521 | 5.292 | 97.374 | 57.666 | 492.56 |
| original22-v1/memory/float64-unaligned | 10.041 | 12.958 | 9.542 | 96.416 | 64.520 | 886.17 |
| original22-v1/runtime/simd-memory-full | 9.792 | 13.354 | 9.084 | 94.041 | 61.084 | 495.94 |
| original22-v1/runtime/simd-memory-widen | 13.146 | 16.833 | 7.979 | 105.02 | 60.458 | 466.85 |
| original22-v1/runtime/simd-memory-lanes | 12.084 | 14.813 | 5.708 | 110.79 | 63.105 | 479.15 |
| original22-v1/gc/i31-roundtrip | 5.354 | 4.458 | US | 61.416 | US | 403.25 |
| original22-v1/gc/struct-access | 15.333 | 11.479 | US | 86.938 | US | 414.00 |
| original22-v1/gc/struct-allocation | 137.65 | 108.29 | US | 128.56 | US | 2,115.46 |
| original22-v1/gc/struct-indirect-access | 12.688 | 14.124 | US | 104.73 | US | 390.54 |
| original22-v1/threads/private-memory | 6.417 | 12.750 | 6.417 | 77.167 | 59.312 | 416.42 |
| original22-v1/threads/shared-memory | 10.646 | 10.291 | 5.958 | 82.145 | 61.021 | 454.94 |
| original22-v1/threads/atomic-rmw | 10.438 | 9.500 | 6.292 | 96.666 | 59.979 | 447.02 |
| original22-v1/threads/notify-empty | 11.333 | 9.521 | 3.458 | 76.875 | 57.875 | 438.67 |
| original22-v1/threads/wait-mismatch | 16.188 | 9.959 | 6.229 | 84.062 | 60.166 | 461.73 |
| original22-v1/threads/wait-timeout | 10.416 | 10.021 | 3.646 | 86.312 | 58.750 | 437.92 |
| original22-v1/applications/sha256-clang | 9.250 | 11.709 | 8.563 | 125.10 | 62.938 | 473.12 |
| original22-v1/applications/sha256-rust | 11.271 | 12.959 | 9.251 | 149.31 | 70.479 | 488.42 |
| gc-sustained-v1/gc/i31-roundtrip | 11.229 | 6.896 | US | 80.021 | US | 424.90 |
| gc-sustained-v1/gc/struct-access | 16.062 | 15.250 | US | 89.438 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 3,362.81 | 2,190.04 | US | 843.46 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 13.041 | 13.646 | US | 96.876 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 6.604 | 14.688 | 4.292 | 79.459 | 55.604 | 492.48 |
| threads-sustained-v2/threads/shared-memory | 17.750 | 10.167 | 4.542 | 78.250 | 59.854 | 497.29 |
| threads-sustained-v2/threads/atomic-rmw | 15.500 | 10.291 | 5.750 | 88.104 | 56.104 | 454.67 |
| threads-sustained-v2/threads/notify-empty | 10.396 | 9.813 | 4.688 | 79.083 | 57.874 | 456.48 |
| threads-sustained-v2/threads/wait-mismatch | 12.167 | 9.646 | 4.083 | 78.499 | 56.812 | 465.75 |
| threads-sustained-v2/threads/wait-timeout | 11.750 | 9.438 | 3.688 | 96.000 | 63.729 | 474.73 |
| generic-calls-v1/core/generic-calls | 15.938 | 7.688 | 8.500 | 90.520 | 38.479 | 478.77 |

</details>

<details>
<summary>Interpreter: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6,223.73 | 7,757.77 | 5,032.04 | NM | NM | 19,136.42 |
| original22-v1/core/integer-loop | 4,999.96 | 6,771.58 | 5,145.85 | NM | NM | 19,496.58 |
| original22-v1/core/function-calls | 5,139.65 | 7,821.48 | 5,303.85 | NM | NM | 19,804.81 |
| original22-v1/core/memory-load-store | 4,742.54 | 6,608.62 | 5,013.90 | NM | NM | 17,602.15 |
| original22-v1/memory/narrow-unaligned | 4,961.71 | 6,958.98 | 5,158.96 | NM | NM | 19,583.38 |
| original22-v1/memory/dependent-address | 4,958.46 | 8,456.06 | 5,869.94 | NM | NM | 19,313.00 |
| original22-v1/memory/float64-unaligned | 4,660.33 | 7,017.33 | 4,878.94 | NM | NM | 17,757.65 |
| original22-v1/runtime/simd-memory-full | 4,891.06 | 7,292.33 | 4,933.46 | NM | NM | 18,716.23 |
| original22-v1/runtime/simd-memory-widen | 4,698.90 | 7,576.94 | 5,064.06 | NM | NM | 18,334.29 |
| original22-v1/runtime/simd-memory-lanes | 5,974.46 | 7,301.35 | 5,515.21 | NM | NM | 19,249.50 |
| original22-v1/gc/i31-roundtrip | 4,650.98 | 6,884.69 | US | NM | NM | 18,738.50 |
| original22-v1/gc/struct-access | 5,220.04 | 7,139.27 | US | NM | NM | 19,268.98 |
| original22-v1/gc/struct-allocation | 4,708.10 | 6,606.96 | US | NM | NM | 18,629.13 |
| original22-v1/gc/struct-indirect-access | 4,646.79 | 6,504.25 | US | NM | NM | 18,801.81 |
| original22-v1/threads/private-memory | 4,262.75 | 6,484.46 | 5,023.42 | NM | NM | 18,124.50 |
| original22-v1/threads/shared-memory | 5,519.69 | 7,587.77 | 5,434.19 | NM | NM | 19,831.19 |
| original22-v1/threads/atomic-rmw | 5,184.69 | 7,269.50 | 4,974.15 | NM | NM | 18,109.40 |
| original22-v1/threads/notify-empty | 4,579.98 | 7,040.56 | 4,706.96 | NM | NM | 18,876.92 |
| original22-v1/threads/wait-mismatch | 6,351.75 | 7,605.42 | 6,175.06 | NM | NM | 20,491.00 |
| original22-v1/threads/wait-timeout | 5,944.31 | 7,737.35 | 5,299.50 | NM | NM | 18,463.04 |
| original22-v1/applications/sha256-clang | 4,775.94 | 6,936.73 | 4,785.67 | NM | NM | 17,819.81 |
| original22-v1/applications/sha256-rust | 4,499.00 | 6,687.62 | 4,850.60 | NM | NM | 17,880.48 |
| gc-sustained-v1/gc/i31-roundtrip | 5,476.92 | 6,968.04 | US | NM | NM | 17,514.83 |
| gc-sustained-v1/gc/struct-access | 5,196.81 | 7,650.58 | US | NM | NM | 19,786.65 |
| gc-sustained-v1/gc/struct-allocation | 4,865.67 | 7,053.29 | US | NM | NM | 19,988.58 |
| gc-sustained-v1/gc/struct-indirect-access | 4,988.38 | 6,886.29 | US | NM | NM | 18,631.83 |
| threads-sustained-v2/threads/private-memory | 4,438.81 | 6,347.79 | 4,665.60 | NM | NM | 18,045.92 |
| threads-sustained-v2/threads/shared-memory | 5,935.73 | 6,996.27 | 4,887.06 | NM | NM | 18,932.10 |
| threads-sustained-v2/threads/atomic-rmw | 6,738.71 | 8,707.08 | 5,897.54 | NM | NM | 20,099.12 |
| threads-sustained-v2/threads/notify-empty | 4,594.56 | 6,748.54 | 4,820.31 | NM | NM | 17,848.69 |
| threads-sustained-v2/threads/wait-mismatch | 4,696.88 | 6,893.33 | 5,004.33 | NM | NM | 19,263.42 |
| threads-sustained-v2/threads/wait-timeout | 4,848.60 | 7,459.62 | 5,347.29 | NM | NM | 18,941.60 |
| generic-calls-v1/core/generic-calls | 5,539.75 | 7,685.81 | 5,190.04 | NM | NM | 20,366.35 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 47.125 | 43.188 | 39.916 | NM | NM | 50.666 |
| original22-v1/core/integer-loop | 50.021 | 52.855 | 46.625 | NM | NM | 46.772 |
| original22-v1/core/function-calls | 49.084 | 46.979 | 47.584 | NM | NM | 48.188 |
| original22-v1/core/memory-load-store | 49.584 | 53.125 | 40.541 | NM | NM | 45.750 |
| original22-v1/memory/narrow-unaligned | 44.709 | 45.896 | 43.853 | NM | NM | 48.480 |
| original22-v1/memory/dependent-address | 47.292 | 47.251 | 47.792 | NM | NM | 46.230 |
| original22-v1/memory/float64-unaligned | 43.208 | 44.521 | 49.480 | NM | NM | 46.709 |
| original22-v1/runtime/simd-memory-full | 45.875 | 45.771 | 53.083 | NM | NM | 48.959 |
| original22-v1/runtime/simd-memory-widen | 44.709 | 42.855 | 43.459 | NM | NM | 46.125 |
| original22-v1/runtime/simd-memory-lanes | 49.854 | 48.395 | 44.917 | NM | NM | 49.521 |
| original22-v1/gc/i31-roundtrip | 41.541 | 38.958 | US | NM | NM | 39.937 |
| original22-v1/gc/struct-access | 46.001 | 43.041 | US | NM | NM | 55.833 |
| original22-v1/gc/struct-allocation | 45.667 | 41.250 | US | NM | NM | 51.355 |
| original22-v1/gc/struct-indirect-access | 47.562 | 44.834 | US | NM | NM | 44.500 |
| original22-v1/threads/private-memory | 41.958 | 41.292 | 38.938 | NM | NM | 43.062 |
| original22-v1/threads/shared-memory | 47.375 | 51.875 | 42.791 | NM | NM | 44.375 |
| original22-v1/threads/atomic-rmw | 44.395 | 40.791 | 42.459 | NM | NM | 50.312 |
| original22-v1/threads/notify-empty | 38.229 | 40.916 | 40.167 | NM | NM | 45.479 |
| original22-v1/threads/wait-mismatch | 48.417 | 48.062 | 48.125 | NM | NM | 52.978 |
| original22-v1/threads/wait-timeout | 52.791 | 46.834 | 51.937 | NM | NM | 42.188 |
| original22-v1/applications/sha256-clang | 48.562 | 45.083 | 38.438 | NM | NM | 52.625 |
| original22-v1/applications/sha256-rust | 41.376 | 45.395 | 49.333 | NM | NM | 49.376 |
| gc-sustained-v1/gc/i31-roundtrip | 43.188 | 48.124 | US | NM | NM | 42.541 |
| gc-sustained-v1/gc/struct-access | 47.188 | 51.396 | US | NM | NM | 51.354 |
| gc-sustained-v1/gc/struct-allocation | 43.875 | 44.355 | US | NM | NM | 48.834 |
| gc-sustained-v1/gc/struct-indirect-access | 41.979 | 46.208 | US | NM | NM | 51.291 |
| threads-sustained-v2/threads/private-memory | 40.312 | 40.188 | 44.084 | NM | NM | 46.521 |
| threads-sustained-v2/threads/shared-memory | 44.833 | 44.646 | 45.334 | NM | NM | 44.417 |
| threads-sustained-v2/threads/atomic-rmw | 44.730 | 49.395 | 44.292 | NM | NM | 50.645 |
| threads-sustained-v2/threads/notify-empty | 41.000 | 43.187 | 36.938 | NM | NM | 41.709 |
| threads-sustained-v2/threads/wait-mismatch | 41.125 | 39.979 | 41.666 | NM | NM | 36.229 |
| threads-sustained-v2/threads/wait-timeout | 42.688 | 47.980 | 37.499 | NM | NM | 43.729 |
| generic-calls-v1/core/generic-calls | 41.521 | 47.500 | 49.062 | NM | NM | 48.355 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4.688 | 6.146 | 2.708 | NM | NM | 28.813 |
| original22-v1/core/integer-loop | 5.021 | 5.625 | 2.688 | NM | NM | 34.416 |
| original22-v1/core/function-calls | 5.229 | 5.625 | 2.833 | NM | NM | 33.375 |
| original22-v1/core/memory-load-store | 4.667 | 5.208 | 2.479 | NM | NM | 28.770 |
| original22-v1/memory/narrow-unaligned | 5.208 | 5.708 | 2.833 | NM | NM | 30.979 |
| original22-v1/memory/dependent-address | 5.000 | 6.083 | 2.688 | NM | NM | 31.480 |
| original22-v1/memory/float64-unaligned | 4.750 | 5.458 | 2.854 | NM | NM | 28.541 |
| original22-v1/runtime/simd-memory-full | 5.083 | 5.625 | 2.667 | NM | NM | 30.020 |
| original22-v1/runtime/simd-memory-widen | 4.812 | 5.167 | 2.542 | NM | NM | 29.541 |
| original22-v1/runtime/simd-memory-lanes | 5.750 | 5.417 | 2.521 | NM | NM | 30.334 |
| original22-v1/gc/i31-roundtrip | 4.938 | 5.604 | US | NM | NM | 30.125 |
| original22-v1/gc/struct-access | 4.208 | 5.958 | US | NM | NM | 31.750 |
| original22-v1/gc/struct-allocation | 5.312 | 6.125 | US | NM | NM | 31.396 |
| original22-v1/gc/struct-indirect-access | 5.396 | 5.583 | US | NM | NM | 30.895 |
| original22-v1/threads/private-memory | 5.646 | 5.583 | 2.896 | NM | NM | 28.416 |
| original22-v1/threads/shared-memory | 5.646 | 5.708 | 2.791 | NM | NM | 27.562 |
| original22-v1/threads/atomic-rmw | 6.208 | 5.521 | 2.604 | NM | NM | 30.187 |
| original22-v1/threads/notify-empty | 4.688 | 5.687 | 2.521 | NM | NM | 29.375 |
| original22-v1/threads/wait-mismatch | 5.396 | 6.021 | 2.708 | NM | NM | 32.438 |
| original22-v1/threads/wait-timeout | 5.896 | 5.687 | 2.958 | NM | NM | 29.834 |
| original22-v1/applications/sha256-clang | 4.729 | 5.667 | 2.646 | NM | NM | 33.104 |
| original22-v1/applications/sha256-rust | 4.896 | 5.375 | 2.896 | NM | NM | 30.646 |
| gc-sustained-v1/gc/i31-roundtrip | 4.625 | 5.479 | US | NM | NM | 30.542 |
| gc-sustained-v1/gc/struct-access | 5.458 | 5.938 | US | NM | NM | 29.791 |
| gc-sustained-v1/gc/struct-allocation | 4.500 | 5.583 | US | NM | NM | 33.396 |
| gc-sustained-v1/gc/struct-indirect-access | 4.583 | 5.542 | US | NM | NM | 29.666 |
| threads-sustained-v2/threads/private-memory | 4.938 | 5.312 | 2.792 | NM | NM | 28.271 |
| threads-sustained-v2/threads/shared-memory | 4.979 | 6.042 | 2.854 | NM | NM | 29.979 |
| threads-sustained-v2/threads/atomic-rmw | 5.333 | 5.604 | 3.042 | NM | NM | 30.354 |
| threads-sustained-v2/threads/notify-empty | 4.229 | 5.229 | 2.438 | NM | NM | 28.604 |
| threads-sustained-v2/threads/wait-mismatch | 4.833 | 5.438 | 2.833 | NM | NM | 28.146 |
| threads-sustained-v2/threads/wait-timeout | 4.854 | 5.625 | 2.750 | NM | NM | 30.479 |
| generic-calls-v1/core/generic-calls | 5.042 | 5.688 | 2.666 | NM | NM | 30.188 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 64.291 | 56.562 | 39.855 | NM | NM | 38.145 |
| original22-v1/core/integer-loop | 74.459 | 65.249 | 46.667 | NM | NM | 41.583 |
| original22-v1/core/function-calls | 84.270 | 75.459 | 46.062 | NM | NM | 42.772 |
| original22-v1/core/memory-load-store | 86.875 | 89.708 | 50.458 | NM | NM | 42.834 |
| original22-v1/memory/narrow-unaligned | 92.208 | 91.001 | 55.105 | NM | NM | 45.917 |
| original22-v1/memory/dependent-address | 90.500 | 88.645 | 50.959 | NM | NM | 46.374 |
| original22-v1/memory/float64-unaligned | 83.063 | 87.709 | 56.875 | NM | NM | 43.021 |
| original22-v1/runtime/simd-memory-full | 82.812 | 77.791 | 54.375 | NM | NM | 44.229 |
| original22-v1/runtime/simd-memory-widen | 82.395 | 73.625 | 52.021 | NM | NM | 44.855 |
| original22-v1/runtime/simd-memory-lanes | 86.750 | 78.041 | 55.938 | NM | NM | 47.062 |
| original22-v1/gc/i31-roundtrip | 68.625 | 60.271 | US | NM | NM | 39.646 |
| original22-v1/gc/struct-access | 73.020 | 65.291 | US | NM | NM | 53.145 |
| original22-v1/gc/struct-allocation | 77.624 | 66.021 | US | NM | NM | 51.812 |
| original22-v1/gc/struct-indirect-access | 87.938 | 81.626 | US | NM | NM | 53.688 |
| original22-v1/threads/private-memory | 69.271 | 58.105 | 45.625 | NM | NM | 39.292 |
| original22-v1/threads/shared-memory | 72.666 | 62.626 | 46.938 | NM | NM | 39.958 |
| original22-v1/threads/atomic-rmw | 69.375 | 59.771 | 47.208 | NM | NM | 40.917 |
| original22-v1/threads/notify-empty | 66.167 | 59.625 | 45.688 | NM | NM | 42.397 |
| original22-v1/threads/wait-mismatch | 73.020 | 63.791 | 50.666 | NM | NM | 43.458 |
| original22-v1/threads/wait-timeout | 71.792 | 62.271 | 47.103 | NM | NM | 41.480 |
| original22-v1/applications/sha256-clang | 292.31 | 392.52 | 87.396 | NM | NM | 89.938 |
| original22-v1/applications/sha256-rust | 286.79 | 403.88 | 89.604 | NM | NM | 94.001 |
| gc-sustained-v1/gc/i31-roundtrip | 66.583 | 56.104 | US | NM | NM | 38.209 |
| gc-sustained-v1/gc/struct-access | 75.458 | 69.687 | US | NM | NM | 50.750 |
| gc-sustained-v1/gc/struct-allocation | 72.438 | 65.084 | US | NM | NM | 52.834 |
| gc-sustained-v1/gc/struct-indirect-access | 89.999 | 85.292 | US | NM | NM | 53.583 |
| threads-sustained-v2/threads/private-memory | 68.209 | 57.083 | 43.167 | NM | NM | 41.270 |
| threads-sustained-v2/threads/shared-memory | 68.499 | 59.666 | 46.062 | NM | NM | 43.084 |
| threads-sustained-v2/threads/atomic-rmw | 69.854 | 60.021 | 50.979 | NM | NM | 42.584 |
| threads-sustained-v2/threads/notify-empty | 67.520 | 59.354 | 45.416 | NM | NM | 40.688 |
| threads-sustained-v2/threads/wait-mismatch | 70.751 | 62.666 | 51.355 | NM | NM | 43.041 |
| threads-sustained-v2/threads/wait-timeout | 70.666 | 60.188 | 49.312 | NM | NM | 44.251 |
| generic-calls-v1/core/generic-calls | 76.312 | 74.251 | 43.292 | NM | NM | 44.667 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 8.354 | 8.646 | 17.334 | NM | NM | 27.375 |
| original22-v1/core/integer-loop | 8.167 | 8.791 | 17.500 | NM | NM | 33.709 |
| original22-v1/core/function-calls | 9.041 | 8.938 | 15.625 | NM | NM | 37.416 |
| original22-v1/core/memory-load-store | 12.646 | 13.500 | 17.459 | NM | NM | 49.938 |
| original22-v1/memory/narrow-unaligned | 12.541 | 13.396 | 19.084 | NM | NM | 54.126 |
| original22-v1/memory/dependent-address | 13.291 | 14.604 | 18.520 | NM | NM | 52.605 |
| original22-v1/memory/float64-unaligned | 12.416 | 13.854 | 18.125 | NM | NM | 48.604 |
| original22-v1/runtime/simd-memory-full | 14.938 | 14.104 | 19.771 | NM | NM | 60.729 |
| original22-v1/runtime/simd-memory-widen | 14.437 | 13.188 | 18.292 | NM | NM | 61.562 |
| original22-v1/runtime/simd-memory-lanes | 14.979 | 13.688 | 19.709 | NM | NM | 59.188 |
| original22-v1/gc/i31-roundtrip | 8.187 | 8.750 | US | NM | NM | 27.271 |
| original22-v1/gc/struct-access | 9.166 | 9.646 | US | NM | NM | 32.938 |
| original22-v1/gc/struct-allocation | 10.334 | 10.459 | US | NM | NM | 32.854 |
| original22-v1/gc/struct-indirect-access | 16.105 | 17.875 | US | NM | NM | 39.876 |
| original22-v1/threads/private-memory | 11.521 | 13.751 | 17.833 | NM | NM | 49.146 |
| original22-v1/threads/shared-memory | 17.291 | 15.959 | 17.791 | NM | NM | 54.000 |
| original22-v1/threads/atomic-rmw | 18.104 | 14.229 | 17.520 | NM | NM | 50.270 |
| original22-v1/threads/notify-empty | 16.916 | 14.249 | 18.229 | NM | NM | 52.833 |
| original22-v1/threads/wait-mismatch | 18.000 | 14.979 | 19.687 | NM | NM | 51.271 |
| original22-v1/threads/wait-timeout | 17.937 | 14.751 | 18.479 | NM | NM | 50.417 |
| original22-v1/applications/sha256-clang | 17.021 | 20.875 | 20.875 | NM | NM | 61.499 |
| original22-v1/applications/sha256-rust | 19.146 | 83.542 | 32.000 | NM | NM | 63.105 |
| gc-sustained-v1/gc/i31-roundtrip | 8.125 | 8.626 | US | NM | NM | 28.229 |
| gc-sustained-v1/gc/struct-access | 9.229 | 10.146 | US | NM | NM | 26.937 |
| gc-sustained-v1/gc/struct-allocation | 10.188 | 10.791 | US | NM | NM | 37.771 |
| gc-sustained-v1/gc/struct-indirect-access | 16.250 | 17.125 | US | NM | NM | 37.250 |
| threads-sustained-v2/threads/private-memory | 11.604 | 13.271 | 17.605 | NM | NM | 48.145 |
| threads-sustained-v2/threads/shared-memory | 16.312 | 14.375 | 18.104 | NM | NM | 52.646 |
| threads-sustained-v2/threads/atomic-rmw | 16.313 | 14.584 | 19.812 | NM | NM | 53.271 |
| threads-sustained-v2/threads/notify-empty | 14.938 | 14.417 | 18.521 | NM | NM | 48.646 |
| threads-sustained-v2/threads/wait-mismatch | 16.709 | 14.938 | 17.854 | NM | NM | 49.230 |
| threads-sustained-v2/threads/wait-timeout | 15.833 | 14.646 | 18.958 | NM | NM | 51.188 |
| generic-calls-v1/core/generic-calls | 9.334 | 10.062 | 17.188 | NM | NM | 30.479 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.250 | 0.416 | 0.250 | NM | NM | 0.250 |
| original22-v1/core/integer-loop | 0.250 | 0.375 | 0.208 | NM | NM | 0.209 |
| original22-v1/core/function-calls | 0.250 | 0.375 | 0.250 | NM | NM | 0.271 |
| original22-v1/core/memory-load-store | 0.250 | 0.396 | 0.250 | NM | NM | 0.208 |
| original22-v1/memory/narrow-unaligned | 0.250 | 0.375 | 0.271 | NM | NM | 0.209 |
| original22-v1/memory/dependent-address | 0.250 | 0.375 | 0.250 | NM | NM | 0.230 |
| original22-v1/memory/float64-unaligned | 0.250 | 0.396 | 0.208 | NM | NM | 0.208 |
| original22-v1/runtime/simd-memory-full | 0.271 | 0.396 | 0.250 | NM | NM | 0.230 |
| original22-v1/runtime/simd-memory-widen | 0.250 | 0.375 | 0.250 | NM | NM | 0.188 |
| original22-v1/runtime/simd-memory-lanes | 0.291 | 0.396 | 0.250 | NM | NM | 0.209 |
| original22-v1/gc/i31-roundtrip | 0.250 | 0.459 | US | NM | NM | 0.250 |
| original22-v1/gc/struct-access | 0.250 | 0.375 | US | NM | NM | 0.230 |
| original22-v1/gc/struct-allocation | 0.250 | 0.416 | US | NM | NM | 0.230 |
| original22-v1/gc/struct-indirect-access | 0.250 | 0.354 | US | NM | NM | 0.312 |
| original22-v1/threads/private-memory | 0.250 | 0.417 | 0.209 | NM | NM | 0.208 |
| original22-v1/threads/shared-memory | 0.271 | 0.438 | 0.250 | NM | NM | 0.250 |
| original22-v1/threads/atomic-rmw | 0.271 | 0.375 | 0.208 | NM | NM | 0.230 |
| original22-v1/threads/notify-empty | 0.271 | 0.416 | 0.230 | NM | NM | 0.208 |
| original22-v1/threads/wait-mismatch | 0.250 | 0.438 | 0.230 | NM | NM | 0.292 |
| original22-v1/threads/wait-timeout | 0.291 | 0.417 | 0.229 | NM | NM | 0.250 |
| original22-v1/applications/sha256-clang | 0.271 | 0.396 | 0.230 | NM | NM | 0.230 |
| original22-v1/applications/sha256-rust | 0.250 | 0.375 | 0.250 | NM | NM | 0.229 |
| gc-sustained-v1/gc/i31-roundtrip | 0.250 | 0.375 | US | NM | NM | 0.208 |
| gc-sustained-v1/gc/struct-access | 0.250 | 0.375 | US | NM | NM | 0.312 |
| gc-sustained-v1/gc/struct-allocation | 0.250 | 0.458 | US | NM | NM | 0.292 |
| gc-sustained-v1/gc/struct-indirect-access | 0.291 | 0.334 | US | NM | NM | 0.208 |
| threads-sustained-v2/threads/private-memory | 0.333 | 0.375 | 0.209 | NM | NM | 0.230 |
| threads-sustained-v2/threads/shared-memory | 0.271 | 0.458 | 0.292 | NM | NM | 0.230 |
| threads-sustained-v2/threads/atomic-rmw | 0.271 | 0.438 | 0.292 | NM | NM | 0.250 |
| threads-sustained-v2/threads/notify-empty | 0.291 | 0.417 | 0.188 | NM | NM | 0.250 |
| threads-sustained-v2/threads/wait-mismatch | 0.250 | 0.417 | 0.208 | NM | NM | 0.230 |
| threads-sustained-v2/threads/wait-timeout | 0.271 | 0.396 | 0.250 | NM | NM | 0.208 |
| generic-calls-v1/core/generic-calls | 0.250 | 0.396 | 0.250 | NM | NM | 0.208 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.979 | 7.875 | 3.792 | NM | NM | 8.312 |
| original22-v1/core/integer-loop | 36,096.81 | 29,434.35 | 93,365.94 | NM | NM | 58,238.81 |
| original22-v1/core/function-calls | 58,267.71 | 53,521.71 | 96,352.04 | NM | NM | 73,996.27 |
| original22-v1/core/memory-load-store | 72,626.29 | 65,121.67 | 162,072.65 | NM | NM | 99,911.69 |
| original22-v1/memory/narrow-unaligned | 98,359.48 | 87,793.65 | 217,700.88 | NM | NM | 138,484.98 |
| original22-v1/memory/dependent-address | 47,102.69 | 41,499.52 | 117,018.52 | NM | NM | 62,803.60 |
| original22-v1/memory/float64-unaligned | 77,315.44 | 68,076.23 | 154,285.94 | NM | NM | 94,652.33 |
| original22-v1/runtime/simd-memory-full | 87,262.75 | 76,319.42 | 168,448.81 | NM | NM | 88,650.79 |
| original22-v1/runtime/simd-memory-widen | 70,696.98 | 64,398.98 | 164,922.44 | NM | NM | 70,289.33 |
| original22-v1/runtime/simd-memory-lanes | 108,987.88 | 101,637.04 | 217,139.19 | NM | NM | 106,613.96 |
| original22-v1/gc/i31-roundtrip | 428.88 | 384.21 | US | NM | NM | 506.62 |
| original22-v1/gc/struct-access | 508.98 | 456.44 | US | NM | NM | 622.71 |
| original22-v1/gc/struct-allocation | 1,034.12 | 1,336.73 | US | NM | NM | 1,333.90 |
| original22-v1/gc/struct-indirect-access | 840.94 | 814.81 | US | NM | NM | 1,076.65 |
| original22-v1/threads/private-memory | 436.62 | 375.56 | 683.33 | NM | NM | 457.90 |
| original22-v1/threads/shared-memory | 857.88 | 415.23 | 743.21 | NM | NM | 456.65 |
| original22-v1/threads/atomic-rmw | 582.02 | 365.79 | 495.65 | NM | NM | 391.42 |
| original22-v1/threads/notify-empty | 350.10 | 314.21 | 467.67 | NM | NM | 461.29 |
| original22-v1/threads/wait-mismatch | 661.33 | 421.85 | 787.19 | NM | NM | 506.85 |
| original22-v1/threads/wait-timeout | 627.67 | 388.54 | 3,464.25 | NM | NM | 1,150.58 |
| original22-v1/applications/sha256-clang | 103,366.94 | 88,812.08 | 220,540.98 | NM | NM | 151,340.19 |
| original22-v1/applications/sha256-rust | 97,767.65 | 88,864.81 | 230,555.08 | NM | NM | 134,590.54 |
| gc-sustained-v1/gc/i31-roundtrip | 40,189.94 | 32,475.85 | US | NM | NM | 50,577.42 |
| gc-sustained-v1/gc/struct-access | 46,205.38 | 44,873.73 | US | NM | NM | 64,745.56 |
| gc-sustained-v1/gc/struct-allocation | 57,982.67 | 43,575.48 | US | NM | NM | 35,752.23 |
| gc-sustained-v1/gc/struct-indirect-access | 85,060.31 | 79,262.96 | US | NM | NM | 108,531.60 |
| threads-sustained-v2/threads/private-memory | 4,150.40 | 3,620.06 | 6,958.15 | NM | NM | 4,215.40 |
| threads-sustained-v2/threads/shared-memory | 8,525.40 | 3,929.44 | 7,291.83 | NM | NM | 4,738.12 |
| threads-sustained-v2/threads/atomic-rmw | 5,879.31 | 3,711.23 | 5,604.10 | NM | NM | 3,766.62 |
| threads-sustained-v2/threads/notify-empty | 3,353.98 | 3,081.88 | 4,533.92 | NM | NM | 4,502.65 |
| threads-sustained-v2/threads/wait-mismatch | 6,478.73 | 4,332.60 | 7,382.54 | NM | NM | 5,061.92 |
| threads-sustained-v2/threads/wait-timeout | 6,080.79 | 4,002.67 | 37,878.31 | NM | NM | 11,383.50 |
| generic-calls-v1/core/generic-calls | 586,696.98 | 537,992.44 | 1,056,683.25 | NM | NM | 713,965.23 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.375 | 1.938 | 4.938 | NM | NM | 6.500 |
| original22-v1/core/integer-loop | 6.375 | 2.125 | 7.021 | NM | NM | 9.438 |
| original22-v1/core/function-calls | 7.604 | 2.125 | 4.542 | NM | NM | 13.771 |
| original22-v1/core/memory-load-store | 2.562 | 7.500 | 5.229 | NM | NM | 53.895 |
| original22-v1/memory/narrow-unaligned | 2.625 | 7.604 | 5.000 | NM | NM | 53.812 |
| original22-v1/memory/dependent-address | 2.583 | 6.688 | 8.146 | NM | NM | 52.958 |
| original22-v1/memory/float64-unaligned | 2.729 | 6.938 | 9.562 | NM | NM | 54.062 |
| original22-v1/runtime/simd-memory-full | 2.500 | 6.792 | 4.396 | NM | NM | 52.729 |
| original22-v1/runtime/simd-memory-widen | 2.292 | 6.833 | 4.875 | NM | NM | 50.666 |
| original22-v1/runtime/simd-memory-lanes | 2.667 | 7.167 | 7.208 | NM | NM | 53.188 |
| original22-v1/gc/i31-roundtrip | 5.625 | 1.938 | US | NM | NM | 6.542 |
| original22-v1/gc/struct-access | 5.896 | 2.209 | US | NM | NM | 7.500 |
| original22-v1/gc/struct-allocation | 104.02 | 58.709 | US | NM | NM | 1,704.29 |
| original22-v1/gc/struct-indirect-access | 2.021 | 1.958 | US | NM | NM | 7.271 |
| original22-v1/threads/private-memory | 2.313 | 5.354 | 4.083 | NM | NM | 44.124 |
| original22-v1/threads/shared-memory | 9.646 | 8.312 | 4.688 | NM | NM | 45.666 |
| original22-v1/threads/atomic-rmw | 9.917 | 9.875 | 4.396 | NM | NM | 46.188 |
| original22-v1/threads/notify-empty | 8.999 | 9.187 | 4.542 | NM | NM | 43.313 |
| original22-v1/threads/wait-mismatch | 10.041 | 9.251 | 5.458 | NM | NM | 47.334 |
| original22-v1/threads/wait-timeout | 11.354 | 8.791 | 3.667 | NM | NM | 48.145 |
| original22-v1/applications/sha256-clang | 2.750 | 6.542 | 5.292 | NM | NM | 57.375 |
| original22-v1/applications/sha256-rust | 2.354 | 2.292 | 6.750 | NM | NM | 57.479 |
| gc-sustained-v1/gc/i31-roundtrip | 6.438 | 2.062 | US | NM | NM | 7.312 |
| gc-sustained-v1/gc/struct-access | 7.312 | 2.458 | US | NM | NM | 8.521 |
| gc-sustained-v1/gc/struct-allocation | 3,055.00 | 1,859.06 | US | NM | NM | 61,447.58 |
| gc-sustained-v1/gc/struct-indirect-access | 2.292 | 2.333 | US | NM | NM | 7.521 |
| threads-sustained-v2/threads/private-memory | 2.375 | 6.812 | 5.042 | NM | NM | 48.979 |
| threads-sustained-v2/threads/shared-memory | 11.188 | 8.916 | 4.938 | NM | NM | 50.728 |
| threads-sustained-v2/threads/atomic-rmw | 10.688 | 10.604 | 4.521 | NM | NM | 49.876 |
| threads-sustained-v2/threads/notify-empty | 10.291 | 10.709 | 4.875 | NM | NM | 48.646 |
| threads-sustained-v2/threads/wait-mismatch | 10.021 | 9.834 | 4.688 | NM | NM | 49.229 |
| threads-sustained-v2/threads/wait-timeout | 10.271 | 9.521 | 3.833 | NM | NM | 51.229 |
| generic-calls-v1/core/generic-calls | 6.500 | 2.021 | 4.479 | NM | NM | 8.334 |

</details>

### WasmEdge AOT preparation detail

| Workload | Compile + validation + artifact I/O, us | Native image load, us | Post-load validation, us |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 23,694.96 | 247,603.23 | 3.167 |
| original22-v1/core/integer-loop | 25,578.92 | 249,542.90 | 5.604 |
| original22-v1/core/function-calls | 21,819.60 | 249,230.17 | 5.688 |
| original22-v1/core/memory-load-store | 23,090.88 | 248,738.21 | 6.229 |
| original22-v1/memory/narrow-unaligned | 23,542.52 | 249,107.23 | 7.688 |
| original22-v1/memory/dependent-address | 24,752.33 | 249,309.17 | 6.958 |
| original22-v1/memory/float64-unaligned | 25,431.00 | 249,154.17 | 7.854 |
| original22-v1/runtime/simd-memory-full | 21,879.25 | 253,190.08 | 5.292 |
| original22-v1/runtime/simd-memory-widen | 25,649.15 | 248,669.73 | 5.771 |
| original22-v1/runtime/simd-memory-lanes | 23,882.48 | 247,723.77 | 6.229 |
| original22-v1/gc/i31-roundtrip | 21,908.65 | 251,004.38 | 5.104 |
| original22-v1/gc/struct-access | 23,520.00 | 252,845.60 | 4.208 |
| original22-v1/gc/struct-allocation | 23,529.35 | 249,695.40 | 6.750 |
| original22-v1/gc/struct-indirect-access | 22,758.62 | 251,801.94 | 8.541 |
| original22-v1/threads/private-memory | 21,224.90 | 252,330.52 | 5.479 |
| original22-v1/threads/shared-memory | 23,991.21 | 250,746.19 | 5.250 |
| original22-v1/threads/atomic-rmw | 24,059.06 | 249,921.58 | 5.750 |
| original22-v1/threads/notify-empty | 21,015.31 | 250,205.90 | 6.083 |
| original22-v1/threads/wait-mismatch | 24,953.38 | 251,865.27 | 5.729 |
| original22-v1/threads/wait-timeout | 23,031.52 | 251,026.21 | 5.625 |
| original22-v1/applications/sha256-clang | 28,146.90 | 251,830.94 | 10.854 |
| original22-v1/applications/sha256-rust | 29,788.40 | 251,414.35 | 10.438 |
| gc-sustained-v1/gc/i31-roundtrip | 20,604.19 | 248,741.44 | 5.229 |
| gc-sustained-v1/gc/struct-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 20,532.79 | 248,146.54 | 4.896 |
| threads-sustained-v2/threads/shared-memory | 23,997.88 | 249,296.50 | 5.666 |
| threads-sustained-v2/threads/atomic-rmw | 22,577.56 | 249,395.21 | 5.542 |
| threads-sustained-v2/threads/notify-empty | 20,070.50 | 246,039.98 | 5.792 |
| threads-sustained-v2/threads/wait-mismatch | 25,775.00 | 252,560.04 | 5.396 |
| threads-sustained-v2/threads/wait-timeout | 23,835.23 | 254,328.77 | 6.062 |
| generic-calls-v1/core/generic-calls | 24,658.77 | 250,106.94 | 5.167 |

### P3 phase observations

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B / cleanup | 74.812 | 618.77 | NM | 250.10 | NM | NM |
| 16 B / compileValidate | 65,456.69 | 109,472.38 | NM | 58,443.29 | NM | NM |
| 16 B / engineSetup | 0.396 | 0.334 | NM | 159.81 | NM | NM |
| 16 B / executorSetup | N/A | N/A | NM | 46.770 | NM | NM |
| 16 B / exportLookup | 0.708 | 1.355 | NM | 9.874 | NM | NM |
| 16 B / firstCall | 11,138.29 | 29,327.96 | NM | 9,897.81 | NM | NM |
| 16 B / instantiate | 507.06 | 793.56 | NM | 212.52 | NM | NM |
| 16 B / load | 64.104 | 72.645 | NM | 63.083 | NM | NM |
| 16 B / processStartup | 6,400.98 | 13,485.42 | NM | 7,189.50 | NM | NM |
| 16 B / providerSetup | 50.313 | 216.71 | NM | 253.96 | NM | NM |
| 16 B / requestPreparation | 0.708 | 0.895 | NM | N/A | NM | NM |
| 65536 B / cleanup | 58.812 | 567.52 | NM | 268.42 | NM | NM |
| 65536 B / compileValidate | 68,988.37 | 109,289.94 | NM | 56,868.73 | NM | NM |
| 65536 B / engineSetup | 0.459 | 0.208 | NM | 165.96 | NM | NM |
| 65536 B / executorSetup | N/A | N/A | NM | 46.708 | NM | NM |
| 65536 B / exportLookup | 0.771 | 1.500 | NM | 9.562 | NM | NM |
| 65536 B / firstCall | 15,164.46 | 31,296.12 | NM | 14,407.52 | NM | NM |
| 65536 B / instantiate | 524.42 | 777.15 | NM | 212.12 | NM | NM |
| 65536 B / load | 67.709 | 69.438 | NM | 64.917 | NM | NM |
| 65536 B / processStartup | 6,834.12 | 13,210.60 | NM | 7,376.08 | NM | NM |
| 65536 B / providerSetup | 54.562 | 190.62 | NM | 258.75 | NM | NM |
| 65536 B / requestPreparation | 8.396 | 7.729 | NM | N/A | NM | NM |

## Native controls (different contracts)

Unchecked C kernels omit runtime safety/lifetime contracts; native Go implements HTTP replies without Component/WASI. These are overhead controls with different contracts. The table reports their measured timing directly. The general-call C control runs twice as many iterations of one recurrence; The optimized Wasm combines two recurrence updates into one affine operation.

| Workload | Unchecked native C, ns/iteration |
| --- | ---: |
| original22-v1/core/integer-loop | 1.786 |
| original22-v1/core/function-calls | 0.919 |
| original22-v1/core/memory-load-store | 0.295 |
| original22-v1/memory/narrow-unaligned | 0.538 |
| original22-v1/memory/dependent-address | 1.666 |
| original22-v1/memory/float64-unaligned | 0.744 |
| original22-v1/runtime/simd-memory-full | 4.025 |
| original22-v1/runtime/simd-memory-widen | 1.159 |
| original22-v1/runtime/simd-memory-lanes | 5.236 |
| original22-v1/applications/sha256-clang | 181.38 |
| original22-v1/applications/sha256-rust | 194.60 |
| generic-calls-v1/core/generic-calls | 2.323 |

| Payload | Native Go CPU, us/request | Native Go requests/s |
| --- | ---: | ---: |
| 16 B | 3.287 | 1,100,268.40 |
| 65536 B | 99.828 | 24,494.85 |

## WasmEdge sustained AOT failures

Full-size WasmEdge0.17.1 O3 AOT preflight freshly reproduces SIGSEGV in sustained struct access, allocation and indirect access. Small original cases pass. Only those exact three cells remain failures in the timing matrix. The original investigation and the observed cumulative stack growth are retained in the [PR #200 report](performance-pr200-complete-comparison-2026-11-17.md). All table timings come from the current event.

## Generated-code mechanism

The ARM64 diagnostic compression loop from the Clang guest drops from 54 to 45 native instructions per iteration and from seven to one canonical frame accesses. The Rust guest drops from 47 to 39 instructions and from three to zero frame accesses. The extra stack reservation lies outside both loops. One candidate restricted the wider register bank to split value plans and left both compression loops unchanged. The measured implementation also uses the bank for the static register plan. The screen and assembly extracts are included in the protocol archive.

## Resolved adverse primary intervals

| Workload or metric | Source comparison | Head/base ratio and delta |
| --- | ---: | ---: |
| original22-v1/runtime/simd-memory-widen | head-llgo-interpreter/base-llgo-interpreter | 1.014 [1.006, 1.023]; Δ +1.38%; speed -1.36% [-2.26%, -0.61%]; pass |
| P3 65536 B allocsPerRequest | head-go/base-go | 1.000 [1.000, 1.000]; Δ +0.00%; pass |

## Independent counter replay

Six fresh blocks on the exact primary binaries captured retired instructions,
CPU time and cycles around five steady calls per process, with symmetric A/A
controls. These 168 records are separate from the 5,508 primary observations.

| Warm SHA cohort | Go retired instructions Δ vs source base | LLGo retired instructions Δ vs source base | Go CPU speed Δ vs source base | LLGo CPU speed Δ vs source base |
| --- | ---: | ---: | ---: | ---: |
| Clang | −10.41% | −10.49% | +14.54% † | +15.95% † |
| Rust | −10.18% | −10.19% | +10.35% | +10.56% |

The instruction controls resolved for all four comparisons. In the Rust cohort,
relative to Wasmtime, the Go build retires 6.18% more instructions and uses
10.52% more cycles per instruction. LLGo retires 5.68% more instructions and
uses 7.19% more cycles per instruction. The Rust CPU controls resolve against
Wasmtime. Clang CPU controls and the WasmEdge CPU comparison remain unresolved.
Wasmtime's audited compression bodies issue both memory loads ahead of the
arithmetic, while this change leaves wasmvm's loads closer to their consumers.
The independent counters cover the full guest call. The effect of moving the
loads remains to be measured in isolation.

### Adverse source observations

- LLGo interpreter SIMD widening took 1.38% longer in the primary event. The
  independent replay retired the same instruction count to within 0.002% and
  measured 0.81% less CPU time and 0.86% fewer cycles per instruction. The
  instruction count and virtual address of each of `runSlotsLoop` and `runSIMD`
  match across binaries. Changes in their disassembly are relocated data and
  call references. The timing regression did not reproduce. The exact source
  of the cycle variation remains open.
- The Go 64 KiB P3 allocation median moved from 510.504150 to 510.505420
  allocations per request, about 0.65 event per 512-request batch. The baseline
  and current windows overlap in the 261,374–261,390 range. The counter samples
  process-wide managed allocation and the aggregate observation leaves allocation
  sites unresolved. CPU and throughput show no resolved adverse source interval
  for this cohort.

## Reproducible evidence

- [Machine-readable results, all controls and source contrasts](data/guarded-arm64-loop-registers-2026-11-21.json).
- [All 5,508 primary process records](data/guarded-arm64-loop-registers-2026-11-21-raw.jsonl.gz), SHA-256 `9da11c02150dd7fb2beeded3e0a4b6748ae567eb39e8457cc86b9795e21a5245`.
- [Frozen plans, commands, hashes, analysis/export scripts, semantic checks and failure evidence](data/guarded-arm64-loop-registers-2026-11-21-protocol.tar.gz), SHA-256 `9ec9bc661c3988b2c0e22efc71f6e62d8bc53ee9863be7b4464b4bf54ea9dbf0`.
- Per-request vectors unused by these metrics are projected to count/extrema/hash; no latency-tail claim. Public load records omit unrelated desktop command arguments. The separately archived screening observations are excluded from primary ratios.
