# Bounded pure calls: complete runtime comparison, 2026-11-20

Constant-decidable pure integer calls now expand within the shared 256-visit budget. Wrapping affine expressions compose without mutating shared nodes, and final expression literals enter the existing register cache. Unknown/effectful/recursive or over-budget paths retain ordinary calls; precise fuel, original logical stack quotas, caller cold continuations and cancellation remain supported.

The general-call workload falls from about **3.84/3.86 to 1.044/1.044 ns per iteration** for Go/LLGo. Paired source ratios are **0.26995 [0.26260,0.27766]** and **0.26423 [0.25564,0.26948]**, both controls pass: **−73.01%/−73.58% time**, **+270.44%/+278.46% speed**. Both host compilers beat Wasmtime and Wasmer on this workload with resolving controls; WasmEdge O3 AOT observes **1.037 ns**, but its control leaves precise parity unresolved. This workload now includes different engines’ successful loop expansion/inlining; it does not isolate a surviving call ABI.

This fresh event covers all **33 core workloads** and **two complete P3 payloads**, with four independent runtime columns. SHA, field access, shared memory and complete P3 still require work; there is no overall parity or milestone-completion claim. Latest LLGo also has resolved interpreter regressions, including **+15.72% time on SIMD lanes**, investigated separately below. All adverse results remain recorded.

## Percentage differences

Cells show `absolute observation (signed speed Δ%)`. Positive speed change means faster. Native execution and P3 use **Wasmtime** as a fixed reference; interpreter execution uses **wazero**, with **WasmEdge** explicitly named for GC rows where wazero rejects the feature. These choices make each table comparable across revisions; they do not assert that the reference is the fastest competitor.

For time, speed Δ = `(reference / candidate - 1) * 100%`; for throughput, speed Δ = `(candidate / reference - 1) * 100%`. Thus 100 ns to 50 ns is **−50% time**, **+100% speed**. P3 CPU percentages mean work per CPU time, separate from wall-clock throughput. Instruction/allocation/byte counters show quantity Δ rather than speed. Source/SDK comparisons retain paired ratio intervals and report quantity Δ plus speed Δ where applicable; reciprocal interval endpoints are reversed.

Competitive cell percentages are descriptive quotients of the displayed medians; paired block estimates below can differ. `†` on these cells means either the candidate or the named reference has unresolved A/A controls; it never establishes a qualified win. Incomparable preparation/phase boundaries keep separate absolute observations. This is a presentation update of the recorded event, not a measurement of later compiler changes.

## Revisions, profiles and protocol

- Engine source: base `18dbc5cb` (merged PR #201), measured head `329d2c0f` for PR #202. Reused binaries retain their actual producer revisions; `build.json` records verified production-source equivalence to the qualified source. No full source copies or new worktrees were created.
- Go **1.27.3**; LLGo **3c6b33ec**, **`-O3 -lto=full`**, LLVM **22.1.8**, default production GC. Source base/head share the compiler profile. LLGo **114c98b4** on the same source baseline measures the SDK upgrade independently.
- wazero **1.12.0**, native compiler/interpreter; Wasmtime **48.0.1**, Cranelift speed; Wasmer **7.3.0**, Cranelift; WasmEdge **0.17.1**, interpreter and **O3 AOT**. These are explicit pinned profiles, not a claim that every available backend or configuration was measured.
- Local Apple M4 Max, macOS arm64, 128 GiB RAM. External desktop load is uncontrolled and recorded; no owned builds, tests, profiles, analysis or edits overlap timing. This is not a Linux amd64 performance result.
- Core: six fixed fresh-process blocks, each with symmetric same-binary A/A, two warmups and seven interleaved measured calls. **6,156 successful process records** after excluding exactly three failing sustained WasmEdge AOT cells. All 33 workloads remain represented. Allocation workloads use means including pauses; others use medians.
- P3: **84 processes per payload**, seven roles with A/A, six fixed blocks, four workers, sixteen clients and twenty interleaved 512-request windows plus a separate flush. Verify status, headers, full body, trailers, CRC and joined guest completion. CPU/instruction counters include the driver and checks. Mach ticks use the recorded timebase.
- Ratios use paired block geometric means and 20,000 fixed-seed bootstrap resamples with marginal 95% intervals. Both relevant A/A intervals must fit `[0.975, 1.025]`. No simultaneous portfolio-acceptance claim, adaptive extension, phase subtraction or cross-event pooling.

**Legend:** `†` means the role or its percentage reference has an unresolved same-binary timing control. Values are observations; a qualified comparison additionally needs both controls and its ratio interval. `US` means this pinned profile rejected the feature in semantic preflight. `NM` means this benchmark adapter/profile was not measured; it does not claim the engine lacks the capability. `N/A` means the metric does not apply. `FAIL` is an execution failure, never an unsupported-feature classification or performance win.

## Native compiled execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

Reference: **Wasmtime**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2,687.50 (-69.77% speed) † | 2,562.50 (-68.29% speed) † | 375.00 (+116.67% speed) † | 812.50 (ref) † | 2,208.50 (-63.21% speed) † | 3,167.00 (-74.34% speed) † |
| original22-v1/core/integer-loop | 1.566 (-0.78% speed) † | 1.559 (-0.31% speed) † | 1.697 (-8.43% speed) † | 1.554 (ref) † | 1.561 (-0.41% speed) † | 1.559 (-0.32% speed) † |
| original22-v1/core/function-calls | 0.800 (+13.24% speed) | 0.783 (+15.65% speed) | 1.409 (-35.75% speed) † | 0.906 (ref) | 0.784 (+15.54% speed) † | 0.781 (+16.01% speed) |
| original22-v1/core/memory-load-store | 0.448 (+7.82% speed) † | 0.433 (+11.59% speed) † | 0.873 (-44.70% speed) † | 0.483 (ref) † | 0.537 (-10.08% speed) † | 0.285 (+69.73% speed) † |
| original22-v1/memory/narrow-unaligned | 0.672 (+7.38% speed) † | 0.664 (+8.54% speed) † | 1.074 (-32.82% speed) † | 0.721 (ref) † | 0.753 (-4.20% speed) † | 0.534 (+35.17% speed) † |
| original22-v1/memory/dependent-address | 1.878 (+11.15% speed) † | 1.852 (+12.71% speed) † | 3.160 (-33.94% speed) † | 2.087 (ref) † | 2.114 (-1.26% speed) † | 1.440 (+44.95% speed) † |
| original22-v1/memory/float64-unaligned | 0.765 (-5.74% speed) † | 0.750 (-3.83% speed) † | 0.877 (-17.73% speed) † | 0.721 (ref) † | 0.728 (-1.00% speed) † | 0.674 (+7.04% speed) † |
| original22-v1/runtime/simd-memory-full | 3.391 (+7.29% speed) † | 3.361 (+8.26% speed) † | 3.706 (-1.83% speed) † | 3.639 (ref) † | 3.504 (+3.83% speed) † | 3.585 (+1.51% speed) † |
| original22-v1/runtime/simd-memory-widen | 1.050 (-1.60% speed) † | 1.037 (-0.36% speed) | 1.121 (-7.80% speed) † | 1.033 (ref) | 1.041 (-0.70% speed) | 1.037 (-0.40% speed) |
| original22-v1/runtime/simd-memory-lanes | 5.029 (+15.99% speed) † | 4.981 (+17.13% speed) † | 5.970 (-2.29% speed) † | 5.834 (ref) † | 5.811 (+0.39% speed) † | 5.027 (+16.05% speed) † |
| original22-v1/gc/i31-roundtrip | 2.035 (+3.79% speed) † | 1.792 (+17.91% speed) † | US | 2.113 (ref) † | US | 1.860 (+13.55% speed) † |
| original22-v1/gc/struct-access | 2.331 (-39.14% speed) † | 2.006 (-29.28% speed) † | US | 1.419 (ref) † | US | 10.023 (-85.84% speed) † |
| original22-v1/gc/struct-allocation | 30.301 (+129.71% speed) † | 29.488 (+136.05% speed) † | US | 69.605 (ref) † | US | 38.949 (+78.71% speed) † |
| original22-v1/gc/struct-indirect-access | 2.860 (+9.61% speed) † | 2.425 (+29.29% speed) † | US | 3.135 (ref) † | US | 26.669 (-88.24% speed) † |
| original22-v1/threads/private-memory | 1.263 (+9.08% speed) † | 0.710 (+93.85% speed) † | 1.750 (-21.31% speed) † | 1.377 (ref) † | 1.662 (-17.16% speed) † | 1.690 (-18.50% speed) † |
| original22-v1/threads/shared-memory | 2.204 (-31.95% speed) † | 1.875 (-20.00% speed) † | 1.746 (-14.08% speed) † | 1.500 (ref) † | 1.756 (-14.59% speed) † | 1.700 (-11.76% speed) † |
| original22-v1/threads/atomic-rmw | 2.812 (-19.56% speed) † | 2.442 (-7.34% speed) † | 2.350 (-3.72% speed) † | 2.263 (ref) † | 4.675 (-51.60% speed) † | 2.415 (-6.30% speed) † |
| original22-v1/threads/notify-empty | 1.746 (+308.12% speed) † | 1.385 (+414.27% speed) † | 21.677 (-67.13% speed) † | 7.125 (ref) † | 5.321 (+33.91% speed) † | 9.321 (-23.56% speed) † |
| original22-v1/threads/wait-mismatch | 2.006 (+353.89% speed) † | 1.594 (+471.37% speed) † | 25.067 (-63.67% speed) † | 9.106 (ref) † | 2.969 (+206.73% speed) † | 3.787 (+140.43% speed) † |
| original22-v1/threads/wait-timeout | 2.329 (+1747.78% speed) † | 1.654 (+2501.79% speed) † | 282.05 (-84.74% speed) † | 43.037 (ref) | 169.55 (-74.62% speed) † | 60.690 (-29.09% speed) |
| original22-v1/applications/sha256-clang | 235.31 (-17.12% speed) † | 232.68 (-16.18% speed) † | 313.39 (-37.77% speed) † | 195.02 (ref) † | 211.30 (-7.71% speed) † | 176.29 (+10.62% speed) † |
| original22-v1/applications/sha256-rust | 233.70 (-20.99% speed) | 231.18 (-20.12% speed) | 319.95 (-42.29% speed) | 184.65 (ref) | 186.80 (-1.15% speed) | 180.92 (+2.06% speed) † |
| gc-sustained-v1/gc/i31-roundtrip | 1.363 (+33.59% speed) † | 1.350 (+34.94% speed) † | US | 1.821 (ref) † | US | 1.300 (+40.10% speed) † |
| gc-sustained-v1/gc/struct-access | 1.328 (-20.88% speed) † | 1.325 (-20.69% speed) † | US | 1.051 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 32.015 (+57.38% speed) † | 36.054 (+39.75% speed) † | US | 50.387 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1.696 (+51.92% speed) † | 1.669 (+54.37% speed) † | US | 2.577 (ref) † | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.435 (+150.43% speed) † | 0.342 (+218.53% speed) † | 1.211 (-10.10% speed) † | 1.088 (ref) † | 1.126 (-3.35% speed) † | 1.121 (-2.95% speed) † |
| threads-sustained-v2/threads/shared-memory | 1.445 (-25.15% speed) † | 1.318 (-17.91% speed) † | 1.261 (-14.24% speed) † | 1.082 (ref) | 1.113 (-2.84% speed) † | 1.117 (-3.13% speed) |
| threads-sustained-v2/threads/atomic-rmw | 1.996 (-6.91% speed) † | 1.900 (-2.17% speed) † | 2.041 (-8.93% speed) † | 1.858 (ref) † | 3.950 (-52.95% speed) † | 1.895 (-1.96% speed) † |
| threads-sustained-v2/threads/notify-empty | 1.105 (+520.75% speed) † | 0.980 (+600.06% speed) † | 21.321 (-67.82% speed) † | 6.861 (ref) † | 4.714 (+45.54% speed) † | 8.974 (-23.55% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 1.184 (+644.06% speed) † | 1.121 (+686.25% speed) † | 24.353 (-63.81% speed) † | 8.812 (ref) | 2.400 (+267.25% speed) † | 3.193 (+175.96% speed) † |
| threads-sustained-v2/threads/wait-timeout | 1.125 (+3729.12% speed) † | 0.991 (+4248.42% speed) † | 282.73 (-84.76% speed) | 43.086 (ref) | 169.69 (-74.61% speed) | 60.074 (-28.28% speed) † |
| generic-calls-v1/core/generic-calls | 1.044 (+98.71% speed) | 1.044 (+98.62% speed) | 2.137 (-2.96% speed) † | 2.074 (ref) | 2.078 (-0.19% speed) | 1.037 (+99.93% speed) † |

## Interpreter execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

Reference: **wazero**; GC rows explicitly use **WasmEdge**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2,083.50 (-57.00% speed) † | 2,020.50 (-55.65% speed) † | 896.00 (ref) † | NM | NM | 2,542.00 (-64.75% speed) † |
| original22-v1/core/integer-loop | 31.664 (+160.38% speed) | 25.445 (+224.02% speed) | 82.446 (ref) | NM | NM | 49.152 (+67.74% speed) † |
| original22-v1/core/function-calls | 48.689 (+87.19% speed) † | 45.629 (+99.74% speed) † | 91.141 (ref) † | NM | NM | 62.498 (+45.83% speed) † |
| original22-v1/core/memory-load-store | 66.075 (+123.29% speed) † | 59.239 (+149.06% speed) † | 147.54 (ref) † | NM | NM | 91.297 (+61.61% speed) † |
| original22-v1/memory/narrow-unaligned | 89.369 (+112.98% speed) | 78.792 (+141.57% speed) | 190.34 (ref) | NM | NM | 117.43 (+62.09% speed) |
| original22-v1/memory/dependent-address | 42.022 (+135.52% speed) † | 36.502 (+171.13% speed) † | 98.971 (ref) † | NM | NM | 54.637 (+81.14% speed) † |
| original22-v1/memory/float64-unaligned | 70.193 (+99.81% speed) | 62.246 (+125.33% speed) | 140.26 (ref) | NM | NM | 85.855 (+63.36% speed) † |
| original22-v1/runtime/simd-memory-full | 76.651 (+96.90% speed) | 66.523 (+126.88% speed) | 150.92 (ref) | NM | NM | 77.660 (+94.34% speed) |
| original22-v1/runtime/simd-memory-widen | 62.726 (+129.64% speed) | 55.710 (+158.56% speed) | 144.04 (ref) | NM | NM | 61.092 (+135.78% speed) |
| original22-v1/runtime/simd-memory-lanes | 95.290 (+95.15% speed) | 102.17 (+82.00% speed) | 185.96 (ref) | NM | NM | 90.638 (+105.17% speed) |
| original22-v1/gc/i31-roundtrip (vs WasmEdge) | 36.292 (+25.04% speed) † | 33.075 (+37.20% speed) † | US | NM | NM | 45.379 (ref) |
| original22-v1/gc/struct-access (vs WasmEdge) | 42.142 (+31.41% speed) † | 40.310 (+37.38% speed) | US | NM | NM | 55.379 (ref) |
| original22-v1/gc/struct-allocation (vs WasmEdge) | 92.216 (+16.70% speed) † | 101.79 (+5.72% speed) † | US | NM | NM | 107.62 (ref) † |
| original22-v1/gc/struct-indirect-access (vs WasmEdge) | 76.431 (+25.65% speed) † | 70.156 (+36.89% speed) † | US | NM | NM | 96.033 (ref) |
| original22-v1/threads/private-memory | 40.098 (+61.04% speed) † | 34.348 (+88.00% speed) † | 64.573 (ref) † | NM | NM | 39.546 (+63.29% speed) † |
| original22-v1/threads/shared-memory | 75.185 (-16.07% speed) † | 35.006 (+80.25% speed) † | 63.100 (ref) † | NM | NM | 39.713 (+58.89% speed) † |
| original22-v1/threads/atomic-rmw | 52.808 (-11.74% speed) † | 33.254 (+40.16% speed) † | 46.608 (ref) † | NM | NM | 33.062 (+40.97% speed) † |
| original22-v1/threads/notify-empty | 31.696 (+36.35% speed) † | 27.908 (+54.86% speed) † | 43.219 (ref) † | NM | NM | 39.881 (+8.37% speed) † |
| original22-v1/threads/wait-mismatch | 57.133 (+20.84% speed) † | 36.635 (+88.46% speed) † | 69.042 (ref) † | NM | NM | 43.492 (+58.75% speed) † |
| original22-v1/threads/wait-timeout | 55.696 (+487.19% speed) | 35.248 (+827.82% speed) † | 327.04 (ref) | NM | NM | 96.769 (+237.96% speed) † |
| original22-v1/applications/sha256-clang | 19,295.14 (+122.00% speed) | 16,670.81 (+156.95% speed) | 42,835.62 (ref) | NM | NM | 27,977.89 (+53.11% speed) |
| original22-v1/applications/sha256-rust | 17,995.18 (+128.16% speed) | 15,789.55 (+160.04% speed) | 41,058.38 (ref) | NM | NM | 24,387.69 (+68.36% speed) |
| gc-sustained-v1/gc/i31-roundtrip (vs WasmEdge) | 35.446 (+25.45% speed) † | 31.960 (+39.13% speed) † | US | NM | NM | 44.467 (ref) |
| gc-sustained-v1/gc/struct-access (vs WasmEdge) | 40.732 (+33.96% speed) | 39.014 (+39.85% speed) | US | NM | NM | 54.563 (ref) |
| gc-sustained-v1/gc/struct-allocation (vs WasmEdge) | 98.679 (+6.14% speed) † | 107.06 (-2.17% speed) | US | NM | NM | 104.74 (ref) |
| gc-sustained-v1/gc/struct-indirect-access (vs WasmEdge) | 73.918 (+29.41% speed) | 69.705 (+37.23% speed) † | US | NM | NM | 95.660 (ref) |
| threads-sustained-v2/threads/private-memory | 39.110 (+63.10% speed) | 34.160 (+86.74% speed) | 63.790 (ref) | NM | NM | 38.991 (+63.60% speed) |
| threads-sustained-v2/threads/shared-memory | 74.221 (-15.82% speed) † | 34.794 (+79.56% speed) † | 62.478 (ref) † | NM | NM | 39.021 (+60.11% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 51.120 (-10.57% speed) † | 32.671 (+39.94% speed) † | 45.719 (ref) † | NM | NM | 32.104 (+42.41% speed) † |
| threads-sustained-v2/threads/notify-empty | 30.939 (+35.91% speed) † | 27.786 (+51.33% speed) † | 42.048 (ref) † | NM | NM | 40.305 (+4.32% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 55.937 (+24.13% speed) † | 36.565 (+89.89% speed) | 69.433 (ref) | NM | NM | 43.000 (+61.47% speed) |
| threads-sustained-v2/threads/wait-timeout | 55.880 (+499.68% speed) | 35.745 (+837.48% speed) † | 335.10 (ref) | NM | NM | 96.390 (+247.65% speed) |
| generic-calls-v1/core/generic-calls | 106.85 (+67.20% speed) † | 96.880 (+84.41% speed) † | 178.66 (ref) † | NM | NM | 118.92 (+50.23% speed) † |

## Complete P3 HTTP requests

P3 competitor cells marked `NM` have no measured adapter for the pinned component/WASI workload. Wasmtime uses the same validated workload. Managed allocation counts exclude foreign Rust/C/JIT allocations and cannot compare total allocation across engines.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU, us/request | 67.028 (-56.49% speed) † | 185.78 (-84.30% speed) † | NM | 29.166 (ref) † | NM | NM |
| 16 B instructions/request | 707,358.50 (+228.66% Δ) † | 1,300,910.36 (+504.44% Δ) † | NM | 215,226.18 (ref) † | NM | NM |
| 16 B requests/s | 59,630.24 (-15.25% speed) † | 26,053.73 (-62.97% speed) | NM | 70,361.57 (ref) | NM | NM |
| 65536 B CPU, us/request | 91.355 (-55.83% speed) † | 218.63 (-81.54% speed) † | NM | 40.354 (ref) † | NM | NM |
| 65536 B instructions/request | 985,959.74 (+223.60% Δ) | 1,631,605.06 (+435.51% Δ) | NM | 304,682.99 (ref) | NM | NM |
| 65536 B requests/s | 42,296.97 (-29.54% speed) † | 21,671.79 (-63.90% speed) † | NM | 60,026.33 (ref) † | NM | NM |

### P3 source and SDK contrasts

| Metric | Go source head/base | LLGo source head/base | LLGo SDK3c/SDK114 on fixed base |
| --- | ---: | ---: | ---: |
| 16 B cpuUs | 0.995 [0.985, 1.002]; Δ -0.52%; speed +0.52% [-0.20%, +1.49%]; pass | 1.001 [0.989, 1.012]; Δ +0.05%; speed -0.05% [-1.22%, +1.13%]; unresolved | 1.003 [0.995, 1.010]; Δ +0.35%; speed -0.34% [-1.02%, +0.49%]; pass |
| 16 B instructions | 0.998 [0.993, 1.001]; Δ -0.24%; pass | 1.002 [0.998, 1.006]; Δ +0.23%; pass | 1.009 [1.006, 1.013]; Δ +0.94%; pass |
| 16 B throughput | 1.003 [0.995, 1.012]; Δ +0.32%; speed +0.32% [-0.45%, +1.16%]; unresolved | 1.015 [1.004, 1.028]; Δ +1.51%; speed +1.51% [+0.38%, +2.82%]; unresolved | 0.989 [0.975, 1.001]; Δ -1.07%; speed -1.07% [-2.50%, +0.09%]; unresolved |
| 16 B allocsPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 16 B bytesPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 65536 B cpuUs | 0.996 [0.991, 1.002]; Δ -0.38%; speed +0.39% [-0.21%, +0.90%]; pass | 0.996 [0.980, 1.011]; Δ -0.40%; speed +0.40% [-1.07%, +2.03%]; unresolved | 1.007 [1.000, 1.015]; Δ +0.74%; speed -0.73% [-1.49%, -0.04%]; unresolved |
| 65536 B instructions | 0.998 [0.996, 1.001]; Δ -0.18%; pass | 0.998 [0.992, 1.005]; Δ -0.16%; pass | 1.012 [1.009, 1.015]; Δ +1.17%; pass |
| 65536 B throughput | 1.005 [0.994, 1.015]; Δ +0.45%; speed +0.45% [-0.59%, +1.52%]; pass | 1.005 [0.993, 1.016]; Δ +0.45%; speed +0.45% [-0.73%, +1.55%]; unresolved | 0.993 [0.978, 1.007]; Δ -0.72%; speed -0.72% [-2.17%, +0.73%]; unresolved |
| 65536 B allocsPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 65536 B bytesPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |

Throughput ratios above one improve; CPU/instructions/allocation ratios below one improve. The JSON also retains the combined PR source+SDK contrast, all own-baseline observations and every per-role A/A interval.

### Competitive P3 ratios

| Payload / compiler | CPU ratio to Wasmtime [95% CI]; controls | Throughput ratio to Wasmtime [95% CI]; controls |
| --- | ---: | ---: |
| 16 B go | 2.335 [2.282, 2.392]; Δ +133.54%; speed -57.18% [-58.19%, -56.19%]; unresolved | 0.844 [0.837, 0.850]; Δ -15.63%; speed -15.63% [-16.32%, -15.01%]; unresolved |
| 16 B llgo | 6.487 [6.263, 6.728]; Δ +548.70%; speed -84.58% [-85.14%, -84.03%]; unresolved | 0.371 [0.369, 0.374]; Δ -62.86%; speed -62.86% [-63.08%, -62.62%]; pass |
| 65536 B go | 2.349 [2.295, 2.437]; Δ +134.90%; speed -57.43% [-58.96%, -56.44%]; unresolved | 0.702 [0.694, 0.712]; Δ -29.78%; speed -29.78% [-30.57%, -28.81%]; unresolved |
| 65536 B llgo | 5.656 [5.359, 6.080]; Δ +465.59%; speed -82.32% [-83.55%, -81.34%]; unresolved | 0.356 [0.352, 0.360]; Δ -64.40%; speed -64.40% [-64.78%, -64.04%]; unresolved |

## Core source baseline and SDK contrasts

Head/base ratios below one improve. These marginal intervals describe this event; an apparent change in an unchanged mechanism requires diagnosis before being attributed to source changes. The machine-readable report includes every competitor ratio and A/A interval.

### native

| Workload | Go source head/base | LLGo source head/base | LLGo SDK3c/SDK114 on fixed base |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.890 [0.828, 0.960]; Δ -11.05%; speed +12.42% [+4.14%, +20.82%]; unresolved | 1.066 [1.011, 1.130]; Δ +6.60%; speed -6.19% [-11.50%, -1.13%]; unresolved | 1.095 [0.941, 1.336]; Δ +9.50%; speed -8.68% [-25.14%, +6.22%]; unresolved |
| original22-v1/core/integer-loop | 0.998 [0.988, 1.012]; Δ -0.18%; speed +0.18% [-1.14%, +1.21%]; unresolved | 0.999 [0.991, 1.007]; Δ -0.06%; speed +0.06% [-0.65%, +0.88%]; pass | 1.002 [0.997, 1.009]; Δ +0.23%; speed -0.23% [-0.87%, +0.34%]; unresolved |
| original22-v1/core/function-calls | 0.998 [0.995, 1.002]; Δ -0.16%; speed +0.16% [-0.16%, +0.52%]; pass | 0.994 [0.990, 0.998]; Δ -0.60%; speed +0.61% [+0.22%, +1.03%]; unresolved | 1.005 [1.001, 1.010]; Δ +0.54%; speed -0.53% [-0.98%, -0.06%]; unresolved |
| original22-v1/core/memory-load-store | 1.018 [0.980, 1.077]; Δ +1.81%; speed -1.77% [-7.19%, +2.06%]; unresolved | 0.973 [0.942, 0.993]; Δ -2.67%; speed +2.74% [+0.66%, +6.19%]; unresolved | 0.937 [0.812, 1.015]; Δ -6.31%; speed +6.73% [-1.44%, +23.21%]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.999 [0.991, 1.005]; Δ -0.14%; speed +0.14% [-0.53%, +0.86%]; unresolved | 0.993 [0.982, 1.000]; Δ -0.72%; speed +0.73% [-0.04%, +1.80%]; unresolved | 1.005 [0.995, 1.016]; Δ +0.52%; speed -0.51% [-1.60%, +0.55%]; pass |
| original22-v1/memory/dependent-address | 1.011 [0.998, 1.026]; Δ +1.06%; speed -1.05% [-2.49%, +0.25%]; unresolved | 0.992 [0.973, 1.012]; Δ -0.77%; speed +0.77% [-1.15%, +2.78%]; unresolved | 1.012 [0.995, 1.034]; Δ +1.16%; speed -1.15% [-3.25%, +0.50%]; unresolved |
| original22-v1/memory/float64-unaligned | 1.007 [0.997, 1.018]; Δ +0.71%; speed -0.70% [-1.74%, +0.30%]; unresolved | 1.006 [0.999, 1.012]; Δ +0.59%; speed -0.59% [-1.19%, +0.14%]; pass | 0.999 [0.996, 1.003]; Δ -0.07%; speed +0.07% [-0.29%, +0.42%]; pass |
| original22-v1/runtime/simd-memory-full | 1.003 [0.990, 1.019]; Δ +0.27%; speed -0.26% [-1.85%, +1.00%]; unresolved | 0.997 [0.994, 1.000]; Δ -0.31%; speed +0.31% [+0.01%, +0.59%]; pass | 1.001 [0.993, 1.009]; Δ +0.12%; speed -0.12% [-0.90%, +0.67%]; pass |
| original22-v1/runtime/simd-memory-widen | 0.988 [0.975, 0.998]; Δ -1.17%; speed +1.18% [+0.22%, +2.60%]; unresolved | 1.002 [0.994, 1.010]; Δ +0.17%; speed -0.17% [-0.95%, +0.59%]; pass | 1.005 [0.994, 1.016]; Δ +0.50%; speed -0.50% [-1.53%, +0.58%]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 1.006 [0.997, 1.017]; Δ +0.64%; speed -0.64% [-1.72%, +0.26%]; unresolved | 1.010 [0.992, 1.030]; Δ +0.98%; speed -0.97% [-2.96%, +0.77%]; unresolved | 0.997 [0.986, 1.009]; Δ -0.28%; speed +0.28% [-0.85%, +1.42%]; pass |
| original22-v1/gc/i31-roundtrip | 0.945 [0.884, 0.998]; Δ -5.52%; speed +5.85% [+0.25%, +13.14%]; unresolved | 0.980 [0.960, 1.001]; Δ -2.02%; speed +2.06% [-0.06%, +4.16%]; unresolved | 1.017 [0.998, 1.035]; Δ +1.71%; speed -1.68% [-3.39%, +0.23%]; unresolved |
| original22-v1/gc/struct-access | 0.985 [0.922, 1.053]; Δ -1.54%; speed +1.57% [-5.00%, +8.44%]; unresolved | 1.036 [1.013, 1.056]; Δ +3.62%; speed -3.49% [-5.34%, -1.29%]; unresolved | 0.937 [0.913, 0.969]; Δ -6.32%; speed +6.75% [+3.21%, +9.51%]; unresolved |
| original22-v1/gc/struct-allocation | 1.017 [0.990, 1.064]; Δ +1.65%; speed -1.62% [-6.00%, +1.00%]; unresolved | 0.989 [0.969, 1.016]; Δ -1.10%; speed +1.11% [-1.55%, +3.15%]; unresolved | 1.099 [1.074, 1.132]; Δ +9.86%; speed -8.97% [-11.65%, -6.87%]; unresolved |
| original22-v1/gc/struct-indirect-access | 0.954 [0.878, 1.024]; Δ -4.58%; speed +4.80% [-2.30%, +13.94%]; unresolved | 1.030 [1.011, 1.047]; Δ +2.95%; speed -2.87% [-4.45%, -1.08%]; unresolved | 1.019 [0.953, 1.066]; Δ +1.92%; speed -1.88% [-6.19%, +4.93%]; unresolved |
| original22-v1/threads/private-memory | 0.992 [0.919, 1.075]; Δ -0.83%; speed +0.84% [-7.01%, +8.77%]; unresolved | 0.975 [0.917, 1.042]; Δ -2.54%; speed +2.61% [-4.00%, +9.05%]; unresolved | 0.983 [0.940, 1.030]; Δ -1.68%; speed +1.71% [-2.92%, +6.42%]; unresolved |
| original22-v1/threads/shared-memory | 1.041 [0.986, 1.099]; Δ +4.08%; speed -3.92% [-9.02%, +1.38%]; unresolved | 1.011 [0.965, 1.075]; Δ +1.15%; speed -1.13% [-6.97%, +3.61%]; unresolved | 0.992 [0.927, 1.059]; Δ -0.79%; speed +0.80% [-5.56%, +7.82%]; unresolved |
| original22-v1/threads/atomic-rmw | 1.025 [0.983, 1.059]; Δ +2.46%; speed -2.40% [-5.60%, +1.76%]; unresolved | 0.980 [0.947, 1.013]; Δ -2.04%; speed +2.08% [-1.26%, +5.61%]; unresolved | 1.012 [0.995, 1.028]; Δ +1.16%; speed -1.14% [-2.77%, +0.50%]; unresolved |
| original22-v1/threads/notify-empty | 1.037 [0.988, 1.087]; Δ +3.70%; speed -3.57% [-8.05%, +1.20%]; unresolved | 0.984 [0.946, 1.018]; Δ -1.60%; speed +1.62% [-1.80%, +5.70%]; unresolved | 1.000 [0.958, 1.050]; Δ -0.02%; speed +0.02% [-4.73%, +4.34%]; unresolved |
| original22-v1/threads/wait-mismatch | 1.003 [0.938, 1.082]; Δ +0.34%; speed -0.34% [-7.54%, +6.65%]; unresolved | 1.060 [0.991, 1.125]; Δ +6.00%; speed -5.66% [-11.08%, +0.90%]; unresolved | 0.960 [0.909, 1.026]; Δ -3.97%; speed +4.13% [-2.51%, +9.98%]; unresolved |
| original22-v1/threads/wait-timeout | 0.908 [0.841, 0.977]; Δ -9.17%; speed +10.09% [+2.34%, +18.95%]; unresolved | 0.995 [0.966, 1.027]; Δ -0.48%; speed +0.48% [-2.66%, +3.55%]; unresolved | 1.039 [0.997, 1.084]; Δ +3.90%; speed -3.76% [-7.71%, +0.27%]; unresolved |
| original22-v1/applications/sha256-clang | 1.012 [0.994, 1.039]; Δ +1.23%; speed -1.22% [-3.77%, +0.65%]; unresolved | 0.997 [0.989, 1.004]; Δ -0.33%; speed +0.34% [-0.41%, +1.08%]; unresolved | 1.002 [0.989, 1.014]; Δ +0.23%; speed -0.23% [-1.35%, +1.15%]; unresolved |
| original22-v1/applications/sha256-rust | 1.004 [0.997, 1.014]; Δ +0.44%; speed -0.44% [-1.42%, +0.31%]; unresolved | 0.997 [0.990, 1.004]; Δ -0.30%; speed +0.31% [-0.37%, +0.99%]; pass | 1.007 [1.004, 1.011]; Δ +0.72%; speed -0.72% [-1.06%, -0.42%]; pass |
| gc-sustained-v1/gc/i31-roundtrip | 0.996 [0.989, 1.002]; Δ -0.39%; speed +0.39% [-0.18%, +1.16%]; unresolved | 0.997 [0.993, 1.000]; Δ -0.31%; speed +0.31% [-0.03%, +0.66%]; unresolved | 0.999 [0.987, 1.012]; Δ -0.06%; speed +0.06% [-1.18%, +1.37%]; pass |
| gc-sustained-v1/gc/struct-access | 0.997 [0.989, 1.005]; Δ -0.32%; speed +0.32% [-0.53%, +1.16%]; pass | 1.004 [0.994, 1.014]; Δ +0.40%; speed -0.40% [-1.39%, +0.63%]; pass | 0.997 [0.987, 1.008]; Δ -0.26%; speed +0.26% [-0.81%, +1.37%]; pass |
| gc-sustained-v1/gc/struct-allocation | 0.993 [0.989, 0.996]; Δ -0.72%; speed +0.72% [+0.35%, +1.12%]; unresolved | 1.004 [0.977, 1.034]; Δ +0.44%; speed -0.44% [-3.31%, +2.32%]; unresolved | 1.215 [1.194, 1.236]; Δ +21.54%; speed -17.72% [-19.12%, -16.23%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.003 [0.977, 1.031]; Δ +0.32%; speed -0.32% [-3.00%, +2.39%]; unresolved | 0.978 [0.952, 1.004]; Δ -2.22%; speed +2.27% [-0.41%, +5.08%]; unresolved | 1.016 [0.991, 1.037]; Δ +1.59%; speed -1.56% [-3.59%, +0.94%]; unresolved |
| threads-sustained-v2/threads/private-memory | 0.967 [0.926, 1.004]; Δ -3.34%; speed +3.45% [-0.35%, +7.97%]; unresolved | 0.988 [0.976, 0.999]; Δ -1.20%; speed +1.22% [+0.11%, +2.50%]; unresolved | 1.010 [1.000, 1.019]; Δ +1.00%; speed -0.99% [-1.85%, +0.00%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 1.053 [0.997, 1.156]; Δ +5.28%; speed -5.01% [-13.47%, +0.32%]; unresolved | 0.997 [0.981, 1.008]; Δ -0.34%; speed +0.34% [-0.84%, +1.98%]; unresolved | 1.001 [0.990, 1.013]; Δ +0.07%; speed -0.07% [-1.29%, +0.98%]; pass |
| threads-sustained-v2/threads/atomic-rmw | 0.988 [0.969, 1.007]; Δ -1.16%; speed +1.17% [-0.69%, +3.15%]; pass | 0.999 [0.988, 1.015]; Δ -0.13%; speed +0.13% [-1.50%, +1.23%]; unresolved | 0.995 [0.971, 1.014]; Δ -0.54%; speed +0.54% [-1.39%, +3.04%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 0.970 [0.943, 0.999]; Δ -2.97%; speed +3.06% [+0.06%, +6.02%]; unresolved | 0.994 [0.981, 1.009]; Δ -0.62%; speed +0.62% [-0.86%, +1.90%]; unresolved | 1.001 [0.984, 1.015]; Δ +0.09%; speed -0.09% [-1.47%, +1.58%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 0.992 [0.922, 1.045]; Δ -0.80%; speed +0.80% [-4.34%, +8.48%]; unresolved | 1.009 [0.996, 1.024]; Δ +0.93%; speed -0.92% [-2.34%, +0.42%]; unresolved | 0.981 [0.937, 1.016]; Δ -1.85%; speed +1.89% [-1.54%, +6.73%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 1.027 [0.996, 1.062]; Δ +2.65%; speed -2.59% [-5.84%, +0.43%]; unresolved | 1.005 [0.988, 1.021]; Δ +0.53%; speed -0.52% [-2.08%, +1.21%]; unresolved | 0.996 [0.974, 1.018]; Δ -0.39%; speed +0.39% [-1.77%, +2.67%]; unresolved |
| generic-calls-v1/core/generic-calls | 0.270 [0.263, 0.278]; Δ -73.01%; speed +270.44% [+260.16%, +280.81%]; pass | 0.264 [0.256, 0.269]; Δ -73.58%; speed +278.46% [+271.08%, +291.17%]; pass | 1.006 [1.001, 1.009]; Δ +0.58%; speed -0.58% [-0.94%, -0.14%]; unresolved |

### interpreter

| Workload | Go source head/base | LLGo source head/base | LLGo SDK3c/SDK114 on fixed base |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.954 [0.840, 1.127]; Δ -4.58%; speed +4.80% [-11.28%, +18.99%]; unresolved | 1.021 [0.975, 1.064]; Δ +2.12%; speed -2.07% [-6.06%, +2.54%]; unresolved | 1.146 [1.023, 1.254]; Δ +14.65%; speed -12.78% [-20.28%, -2.22%]; unresolved |
| original22-v1/core/integer-loop | 1.002 [0.999, 1.006]; Δ +0.23%; speed -0.23% [-0.62%, +0.14%]; pass | 0.998 [0.987, 1.006]; Δ -0.24%; speed +0.24% [-0.57%, +1.29%]; pass | 1.004 [0.995, 1.015]; Δ +0.45%; speed -0.45% [-1.45%, +0.50%]; pass |
| original22-v1/core/function-calls | 0.986 [0.979, 0.992]; Δ -1.39%; speed +1.41% [+0.80%, +2.16%]; pass | 1.004 [0.999, 1.009]; Δ +0.37%; speed -0.37% [-0.90%, +0.14%]; pass | 1.038 [1.029, 1.047]; Δ +3.80%; speed -3.66% [-4.51%, -2.83%]; pass |
| original22-v1/core/memory-load-store | 0.992 [0.980, 1.002]; Δ -0.79%; speed +0.80% [-0.20%, +2.09%]; pass | 1.003 [0.998, 1.006]; Δ +0.26%; speed -0.25% [-0.57%, +0.18%]; unresolved | 0.994 [0.980, 1.009]; Δ -0.62%; speed +0.62% [-0.91%, +2.05%]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.990 [0.979, 1.004]; Δ -0.96%; speed +0.97% [-0.37%, +2.10%]; pass | 0.988 [0.976, 1.000]; Δ -1.20%; speed +1.21% [-0.05%, +2.44%]; pass | 1.006 [0.999, 1.013]; Δ +0.57%; speed -0.57% [-1.25%, +0.11%]; pass |
| original22-v1/memory/dependent-address | 1.002 [0.981, 1.017]; Δ +0.25%; speed -0.25% [-1.68%, +1.95%]; unresolved | 1.000 [0.980, 1.025]; Δ +0.05%; speed -0.05% [-2.45%, +2.03%]; unresolved | 1.028 [1.014, 1.043]; Δ +2.83%; speed -2.75% [-4.10%, -1.39%]; unresolved |
| original22-v1/memory/float64-unaligned | 0.997 [0.991, 1.001]; Δ -0.33%; speed +0.33% [-0.15%, +0.89%]; pass | 0.997 [0.986, 1.010]; Δ -0.29%; speed +0.29% [-0.94%, +1.37%]; pass | 0.996 [0.981, 1.011]; Δ -0.41%; speed +0.41% [-1.11%, +1.90%]; pass |
| original22-v1/runtime/simd-memory-full | 1.012 [0.994, 1.039]; Δ +1.16%; speed -1.15% [-3.71%, +0.63%]; unresolved | 0.993 [0.979, 1.005]; Δ -0.66%; speed +0.66% [-0.54%, +2.14%]; unresolved | 0.990 [0.971, 1.003]; Δ -1.04%; speed +1.05% [-0.27%, +2.96%]; unresolved |
| original22-v1/runtime/simd-memory-widen | 1.008 [1.000, 1.016]; Δ +0.78%; speed -0.77% [-1.57%, +0.03%]; pass | 0.994 [0.987, 1.001]; Δ -0.58%; speed +0.58% [-0.13%, +1.33%]; pass | 1.017 [1.013, 1.021]; Δ +1.68%; speed -1.65% [-2.04%, -1.28%]; pass |
| original22-v1/runtime/simd-memory-lanes | 1.003 [0.996, 1.009]; Δ +0.27%; speed -0.27% [-0.92%, +0.44%]; pass | 1.002 [0.999, 1.005]; Δ +0.20%; speed -0.20% [-0.54%, +0.06%]; pass | 1.157 [1.155, 1.159]; Δ +15.72%; speed -13.59% [-13.72%, -13.45%]; pass |
| original22-v1/gc/i31-roundtrip | 0.982 [0.969, 0.995]; Δ -1.85%; speed +1.88% [+0.51%, +3.15%]; unresolved | 1.010 [0.987, 1.032]; Δ +0.99%; speed -0.98% [-3.10%, +1.27%]; unresolved | 1.022 [1.005, 1.040]; Δ +2.17%; speed -2.12% [-3.81%, -0.54%]; unresolved |
| original22-v1/gc/struct-access | 1.001 [0.988, 1.019]; Δ +0.15%; speed -0.15% [-1.84%, +1.26%]; unresolved | 0.992 [0.983, 1.001]; Δ -0.82%; speed +0.83% [-0.12%, +1.74%]; unresolved | 1.042 [1.029, 1.062]; Δ +4.24%; speed -4.07% [-5.83%, -2.81%]; unresolved |
| original22-v1/gc/struct-allocation | 0.982 [0.977, 0.987]; Δ -1.80%; speed +1.83% [+1.32%, +2.35%]; pass | 1.003 [1.000, 1.007]; Δ +0.31%; speed -0.31% [-0.65%, +0.01%]; pass | 0.897 [0.891, 0.903]; Δ -10.30%; speed +11.48% [+10.80%, +12.19%]; pass |
| original22-v1/gc/struct-indirect-access | 1.013 [1.005, 1.019]; Δ +1.26%; speed -1.24% [-1.85%, -0.54%]; unresolved | 0.974 [0.965, 0.980]; Δ -2.60%; speed +2.67% [+2.06%, +3.59%]; unresolved | 1.052 [1.039, 1.064]; Δ +5.18%; speed -4.92% [-6.03%, -3.76%]; unresolved |
| original22-v1/threads/private-memory | 0.993 [0.973, 1.014]; Δ -0.69%; speed +0.70% [-1.43%, +2.80%]; pass | 1.012 [0.997, 1.030]; Δ +1.19%; speed -1.18% [-2.90%, +0.33%]; unresolved | 0.982 [0.968, 0.995]; Δ -1.82%; speed +1.85% [+0.46%, +3.32%]; unresolved |
| original22-v1/threads/shared-memory | 0.984 [0.971, 0.997]; Δ -1.64%; speed +1.67% [+0.26%, +2.97%]; unresolved | 0.991 [0.975, 1.006]; Δ -0.89%; speed +0.89% [-0.57%, +2.58%]; unresolved | 1.037 [1.025, 1.048]; Δ +3.67%; speed -3.54% [-4.54%, -2.45%]; unresolved |
| original22-v1/threads/atomic-rmw | 0.999 [0.989, 1.008]; Δ -0.14%; speed +0.14% [-0.81%, +1.16%]; unresolved | 1.014 [0.995, 1.041]; Δ +1.41%; speed -1.39% [-3.93%, +0.50%]; unresolved | 0.984 [0.965, 1.002]; Δ -1.60%; speed +1.62% [-0.22%, +3.61%]; pass |
| original22-v1/threads/notify-empty | 0.994 [0.983, 1.005]; Δ -0.59%; speed +0.59% [-0.50%, +1.74%]; unresolved | 1.016 [0.988, 1.050]; Δ +1.56%; speed -1.53% [-4.75%, +1.21%]; unresolved | 0.982 [0.930, 1.023]; Δ -1.84%; speed +1.88% [-2.22%, +7.52%]; unresolved |
| original22-v1/threads/wait-mismatch | 0.998 [0.983, 1.014]; Δ -0.25%; speed +0.25% [-1.39%, +1.75%]; unresolved | 1.000 [0.985, 1.019]; Δ +0.02%; speed -0.02% [-1.90%, +1.53%]; unresolved | 0.991 [0.976, 1.001]; Δ -0.88%; speed +0.88% [-0.13%, +2.44%]; unresolved |
| original22-v1/threads/wait-timeout | 0.991 [0.982, 1.000]; Δ -0.93%; speed +0.94% [+0.04%, +1.83%]; pass | 1.008 [0.989, 1.029]; Δ +0.78%; speed -0.78% [-2.79%, +1.13%]; unresolved | 0.990 [0.971, 1.008]; Δ -1.04%; speed +1.06% [-0.79%, +3.00%]; unresolved |
| original22-v1/applications/sha256-clang | 1.000 [0.995, 1.006]; Δ -0.02%; speed +0.02% [-0.55%, +0.55%]; pass | 1.002 [0.993, 1.014]; Δ +0.21%; speed -0.21% [-1.38%, +0.70%]; pass | 0.997 [0.988, 1.006]; Δ -0.33%; speed +0.33% [-0.58%, +1.20%]; unresolved |
| original22-v1/applications/sha256-rust | 1.003 [0.999, 1.007]; Δ +0.31%; speed -0.30% [-0.70%, +0.09%]; pass | 0.998 [0.993, 1.004]; Δ -0.15%; speed +0.15% [-0.43%, +0.66%]; pass | 0.981 [0.976, 0.987]; Δ -1.91%; speed +1.95% [+1.33%, +2.48%]; pass |
| gc-sustained-v1/gc/i31-roundtrip | 1.001 [0.993, 1.008]; Δ +0.10%; speed -0.09% [-0.79%, +0.68%]; unresolved | 0.993 [0.972, 1.015]; Δ -0.72%; speed +0.73% [-1.43%, +2.89%]; unresolved | 1.032 [1.019, 1.047]; Δ +3.19%; speed -3.09% [-4.46%, -1.89%]; unresolved |
| gc-sustained-v1/gc/struct-access | 0.999 [0.995, 1.004]; Δ -0.06%; speed +0.06% [-0.42%, +0.51%]; pass | 0.996 [0.987, 1.004]; Δ -0.45%; speed +0.45% [-0.44%, +1.29%]; pass | 1.033 [1.020, 1.045]; Δ +3.31%; speed -3.20% [-4.32%, -1.98%]; pass |
| gc-sustained-v1/gc/struct-allocation | 1.006 [0.979, 1.031]; Δ +0.61%; speed -0.60% [-3.01%, +2.17%]; unresolved | 0.997 [0.985, 1.010]; Δ -0.28%; speed +0.28% [-0.97%, +1.48%]; pass | 0.919 [0.910, 0.928]; Δ -8.08%; speed +8.79% [+7.76%, +9.84%]; pass |
| gc-sustained-v1/gc/struct-indirect-access | 1.010 [1.005, 1.015]; Δ +0.99%; speed -0.98% [-1.51%, -0.46%]; pass | 0.995 [0.987, 1.004]; Δ -0.46%; speed +0.46% [-0.41%, +1.34%]; unresolved | 1.048 [1.029, 1.068]; Δ +4.78%; speed -4.57% [-6.35%, -2.85%]; unresolved |
| threads-sustained-v2/threads/private-memory | 0.985 [0.971, 1.001]; Δ -1.49%; speed +1.51% [-0.12%, +2.97%]; pass | 1.000 [0.993, 1.008]; Δ +0.03%; speed -0.03% [-0.77%, +0.71%]; pass | 0.994 [0.987, 1.001]; Δ -0.61%; speed +0.62% [-0.09%, +1.34%]; pass |
| threads-sustained-v2/threads/shared-memory | 0.993 [0.986, 1.000]; Δ -0.71%; speed +0.72% [+0.04%, +1.43%]; unresolved | 0.995 [0.979, 1.008]; Δ -0.49%; speed +0.49% [-0.80%, +2.18%]; unresolved | 1.047 [1.026, 1.075]; Δ +4.66%; speed -4.45% [-6.95%, -2.51%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.997 [0.990, 1.006]; Δ -0.27%; speed +0.28% [-0.64%, +1.01%]; pass | 1.003 [0.995, 1.009]; Δ +0.31%; speed -0.31% [-0.93%, +0.46%]; unresolved | 1.003 [0.992, 1.014]; Δ +0.27%; speed -0.27% [-1.42%, +0.83%]; pass |
| threads-sustained-v2/threads/notify-empty | 1.002 [0.992, 1.010]; Δ +0.19%; speed -0.19% [-1.03%, +0.76%]; unresolved | 1.005 [0.982, 1.035]; Δ +0.50%; speed -0.50% [-3.41%, +1.82%]; unresolved | 1.024 [1.011, 1.039]; Δ +2.38%; speed -2.33% [-3.77%, -1.06%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 0.988 [0.977, 1.000]; Δ -1.16%; speed +1.17% [+0.00%, +2.32%]; unresolved | 1.003 [0.990, 1.017]; Δ +0.32%; speed -0.32% [-1.67%, +1.01%]; pass | 0.991 [0.982, 1.000]; Δ -0.90%; speed +0.91% [-0.03%, +1.85%]; pass |
| threads-sustained-v2/threads/wait-timeout | 1.008 [0.999, 1.017]; Δ +0.80%; speed -0.79% [-1.66%, +0.08%]; pass | 1.010 [0.997, 1.021]; Δ +0.97%; speed -0.96% [-2.07%, +0.28%]; unresolved | 1.004 [0.992, 1.016]; Δ +0.38%; speed -0.38% [-1.59%, +0.85%]; unresolved |
| generic-calls-v1/core/generic-calls | 0.996 [0.989, 1.006]; Δ -0.35%; speed +0.35% [-0.60%, +1.13%]; unresolved | 1.005 [1.001, 1.011]; Δ +0.53%; speed -0.53% [-1.10%, -0.14%]; pass | 1.038 [1.031, 1.045]; Δ +3.79%; speed -3.65% [-4.27%, -2.98%]; pass |

## Preparation, first call and cleanup

All entries below are **microseconds per process or invocation**, including unresolved observations. Phase intervals are not inferred from warm-call controls. Startup is independently observed process readiness. Native compile/validate boundaries differ by API; WasmEdge AOT includes validation and artifact I/O, plus separately recorded native-image loading. Every process creates a fresh native image. First-call and warm execution have their own boundaries; values are never subtracted to invent pure costs.

<details>
<summary>Native: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 3,878.00 | 5,530.06 | 3,939.73 | 3,935.04 | 5,445.46 | 15,885.10 |
| original22-v1/core/integer-loop | 4,184.00 | 5,796.04 | 4,339.19 | 4,286.79 | 5,758.92 | 16,226.52 |
| original22-v1/core/function-calls | 3,983.50 | 5,701.35 | 4,265.48 | 4,309.81 | 5,782.31 | 16,054.90 |
| original22-v1/core/memory-load-store | 3,985.71 | 5,943.75 | 4,117.88 | 4,318.48 | 5,828.92 | 16,110.29 |
| original22-v1/memory/narrow-unaligned | 4,218.71 | 5,936.83 | 4,287.96 | 4,258.77 | 5,745.42 | 16,249.81 |
| original22-v1/memory/dependent-address | 4,363.44 | 5,919.73 | 4,320.56 | 4,127.19 | 5,899.52 | 15,931.35 |
| original22-v1/memory/float64-unaligned | 4,222.38 | 5,792.54 | 4,305.15 | 4,200.04 | 6,162.02 | 16,104.33 |
| original22-v1/runtime/simd-memory-full | 4,439.42 | 5,914.25 | 4,250.29 | 4,576.10 | 5,789.31 | 16,205.19 |
| original22-v1/runtime/simd-memory-widen | 3,935.06 | 5,661.71 | 3,987.94 | 3,953.46 | 5,815.50 | 15,769.54 |
| original22-v1/runtime/simd-memory-lanes | 3,990.77 | 5,740.02 | 4,228.35 | 4,305.42 | 5,737.40 | 15,808.31 |
| original22-v1/gc/i31-roundtrip | 4,126.56 | 5,797.44 | US | 3,978.13 | US | 15,843.96 |
| original22-v1/gc/struct-access | 3,920.92 | 5,567.88 | US | 3,956.12 | US | 15,959.29 |
| original22-v1/gc/struct-allocation | 3,911.48 | 5,783.50 | US | 4,141.62 | US | 15,700.15 |
| original22-v1/gc/struct-indirect-access | 3,940.46 | 5,856.71 | US | 4,325.83 | US | 15,951.12 |
| original22-v1/threads/private-memory | 3,753.48 | 5,607.33 | 3,821.71 | 3,818.56 | 5,711.69 | 15,565.19 |
| original22-v1/threads/shared-memory | 3,921.56 | 5,882.40 | 3,998.21 | 3,776.29 | 5,488.98 | 15,574.67 |
| original22-v1/threads/atomic-rmw | 3,855.02 | 5,675.21 | 3,881.62 | 4,230.19 | 5,611.92 | 16,016.27 |
| original22-v1/threads/notify-empty | 3,881.00 | 6,271.75 | 4,046.88 | 4,060.54 | 5,445.92 | 15,614.35 |
| original22-v1/threads/wait-mismatch | 3,680.92 | 5,479.58 | 3,984.02 | 3,968.90 | 5,582.98 | 15,972.48 |
| original22-v1/threads/wait-timeout | 4,169.40 | 5,678.71 | 4,063.98 | 4,151.75 | 5,685.83 | 15,738.81 |
| original22-v1/applications/sha256-clang | 4,409.21 | 6,205.60 | 4,530.58 | 4,446.60 | 6,143.56 | 17,042.46 |
| original22-v1/applications/sha256-rust | 4,098.27 | 5,832.10 | 4,144.15 | 4,163.08 | 5,815.08 | 15,755.85 |
| gc-sustained-v1/gc/i31-roundtrip | 3,996.75 | 5,617.71 | US | 4,366.56 | US | 15,756.54 |
| gc-sustained-v1/gc/struct-access | 4,173.77 | 5,958.33 | US | 4,125.58 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 4,283.40 | 5,863.58 | US | 4,332.83 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 4,089.98 | 5,835.44 | US | 4,259.98 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 4,196.67 | 5,649.65 | 3,966.12 | 4,216.81 | 5,861.19 | 16,045.27 |
| threads-sustained-v2/threads/shared-memory | 4,000.71 | 5,823.69 | 3,989.04 | 4,061.21 | 5,668.75 | 16,160.54 |
| threads-sustained-v2/threads/atomic-rmw | 4,034.62 | 5,977.15 | 3,944.92 | 4,112.29 | 5,652.77 | 15,906.00 |
| threads-sustained-v2/threads/notify-empty | 3,949.31 | 5,783.54 | 4,315.04 | 4,230.92 | 5,966.46 | 15,863.58 |
| threads-sustained-v2/threads/wait-mismatch | 4,109.12 | 5,926.54 | 4,066.98 | 4,169.96 | 5,746.83 | 15,540.29 |
| threads-sustained-v2/threads/wait-timeout | 4,029.35 | 5,848.90 | 4,411.29 | 4,266.50 | 6,172.67 | 16,434.46 |
| generic-calls-v1/core/generic-calls | 4,303.56 | 6,030.35 | 4,441.92 | 4,059.40 | 5,812.54 | 16,814.92 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 30.188 | 34.750 | 32.938 | 25.542 | 36.750 | 36.770 |
| original22-v1/core/integer-loop | 35.875 | 35.980 | 32.021 | 28.584 | 35.604 | 39.062 |
| original22-v1/core/function-calls | 33.396 | 32.500 | 32.980 | 30.646 | 28.584 | 41.334 |
| original22-v1/core/memory-load-store | 34.833 | 41.416 | 38.229 | 34.604 | 36.021 | 39.646 |
| original22-v1/memory/narrow-unaligned | 46.355 | 45.855 | 45.375 | 34.688 | 40.416 | 44.125 |
| original22-v1/memory/dependent-address | 42.812 | 45.937 | 38.750 | 33.604 | 37.104 | 37.312 |
| original22-v1/memory/float64-unaligned | 39.000 | 39.188 | 43.541 | 35.417 | 37.938 | 38.167 |
| original22-v1/runtime/simd-memory-full | 43.042 | 37.812 | 39.291 | 31.541 | 34.396 | 38.729 |
| original22-v1/runtime/simd-memory-widen | 35.520 | 36.375 | 36.333 | 29.041 | 30.458 | 37.416 |
| original22-v1/runtime/simd-memory-lanes | 38.353 | 37.688 | 36.041 | 34.729 | 39.625 | 37.480 |
| original22-v1/gc/i31-roundtrip | 35.583 | 35.438 | US | 27.937 | US | 40.854 |
| original22-v1/gc/struct-access | 33.896 | 39.792 | US | 27.562 | US | 36.167 |
| original22-v1/gc/struct-allocation | 32.209 | 35.459 | US | 27.750 | US | 34.271 |
| original22-v1/gc/struct-indirect-access | 32.708 | 36.375 | US | 29.520 | US | 35.313 |
| original22-v1/threads/private-memory | 30.666 | 34.730 | 32.375 | 25.792 | 28.104 | 37.042 |
| original22-v1/threads/shared-memory | 32.999 | 38.646 | 32.188 | 25.917 | 28.188 | 38.395 |
| original22-v1/threads/atomic-rmw | 28.500 | 38.062 | 31.354 | 28.438 | 33.479 | 38.854 |
| original22-v1/threads/notify-empty | 36.520 | 38.771 | 33.562 | 25.438 | 27.750 | 36.041 |
| original22-v1/threads/wait-mismatch | 28.375 | 31.834 | 32.938 | 28.417 | 34.188 | 37.520 |
| original22-v1/threads/wait-timeout | 37.958 | 35.125 | 28.459 | 28.000 | 34.375 | 37.875 |
| original22-v1/applications/sha256-clang | 45.812 | 39.938 | 45.376 | 39.437 | 45.916 | 42.020 |
| original22-v1/applications/sha256-rust | 34.938 | 37.334 | 34.604 | 35.416 | 33.312 | 36.250 |
| gc-sustained-v1/gc/i31-roundtrip | 39.709 | 33.646 | US | 33.770 | US | 43.166 |
| gc-sustained-v1/gc/struct-access | 37.751 | 40.812 | US | 34.209 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 36.209 | 39.479 | US | 32.229 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 40.188 | 43.105 | US | 31.041 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 38.021 | 38.834 | 33.188 | 29.166 | 35.083 | 37.021 |
| threads-sustained-v2/threads/shared-memory | 32.875 | 39.645 | 34.188 | 28.062 | 34.062 | 37.812 |
| threads-sustained-v2/threads/atomic-rmw | 34.146 | 36.270 | 38.854 | 30.458 | 33.292 | 40.812 |
| threads-sustained-v2/threads/notify-empty | 40.438 | 38.812 | 38.646 | 30.521 | 35.874 | 36.688 |
| threads-sustained-v2/threads/wait-mismatch | 36.812 | 37.896 | 34.666 | 25.959 | 30.188 | 36.916 |
| threads-sustained-v2/threads/wait-timeout | 37.604 | 40.042 | 36.938 | 30.230 | 34.479 | 39.042 |
| generic-calls-v1/core/generic-calls | 46.042 | 41.917 | 51.291 | 32.249 | 40.750 | 46.708 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.917 | 4.646 | 55.459 | 119.65 | 40.562 | 28.833 |
| original22-v1/core/integer-loop | 5.771 | 4.875 | 54.666 | 125.81 | 40.688 | 26.604 |
| original22-v1/core/function-calls | 5.792 | 4.688 | 53.979 | 126.12 | 40.501 | 27.000 |
| original22-v1/core/memory-load-store | 5.875 | 5.312 | 56.854 | 131.81 | 39.208 | 27.541 |
| original22-v1/memory/narrow-unaligned | 6.354 | 4.771 | 56.729 | 127.75 | 43.124 | 29.416 |
| original22-v1/memory/dependent-address | 6.417 | 5.354 | 56.959 | 124.88 | 43.230 | 26.270 |
| original22-v1/memory/float64-unaligned | 6.292 | 5.062 | 56.730 | 128.15 | 42.729 | 28.812 |
| original22-v1/runtime/simd-memory-full | 6.146 | 4.708 | 53.666 | 131.04 | 39.604 | 27.063 |
| original22-v1/runtime/simd-memory-widen | 5.896 | 4.604 | 54.896 | 116.71 | 36.416 | 25.645 |
| original22-v1/runtime/simd-memory-lanes | 6.062 | 4.604 | 55.167 | 127.67 | 40.541 | 26.396 |
| original22-v1/gc/i31-roundtrip | 5.417 | 6.416 | US | 124.71 | US | 31.229 |
| original22-v1/gc/struct-access | 5.833 | 5.104 | US | 123.87 | US | 29.479 |
| original22-v1/gc/struct-allocation | 5.729 | 5.292 | US | 124.73 | US | 28.312 |
| original22-v1/gc/struct-indirect-access | 5.938 | 4.750 | US | 131.77 | US | 24.541 |
| original22-v1/threads/private-memory | 5.625 | 4.562 | 54.041 | 122.81 | 43.104 | 28.855 |
| original22-v1/threads/shared-memory | 5.917 | 4.812 | 57.312 | 120.60 | 38.812 | 26.291 |
| original22-v1/threads/atomic-rmw | 5.729 | 5.021 | 57.708 | 131.00 | 37.334 | 27.812 |
| original22-v1/threads/notify-empty | 5.667 | 4.646 | 55.854 | 120.25 | 39.626 | 27.770 |
| original22-v1/threads/wait-mismatch | 5.333 | 4.604 | 53.125 | 128.62 | 41.750 | 27.916 |
| original22-v1/threads/wait-timeout | 6.792 | 4.521 | 53.021 | 127.10 | 43.105 | 26.167 |
| original22-v1/applications/sha256-clang | 6.542 | 5.354 | 61.438 | 136.25 | 45.792 | 28.479 |
| original22-v1/applications/sha256-rust | 5.146 | 4.604 | 55.541 | 125.79 | 38.791 | 25.667 |
| gc-sustained-v1/gc/i31-roundtrip | 6.125 | 4.729 | US | 130.79 | US | 26.750 |
| gc-sustained-v1/gc/struct-access | 5.771 | 5.271 | US | 135.31 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 5.688 | 5.146 | US | 123.81 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 6.229 | 5.458 | US | 121.21 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 6.479 | 5.083 | 57.458 | 125.69 | 41.312 | 28.854 |
| threads-sustained-v2/threads/shared-memory | 5.938 | 4.958 | 57.312 | 125.81 | 38.084 | 27.750 |
| threads-sustained-v2/threads/atomic-rmw | 5.729 | 5.333 | 53.895 | 124.12 | 39.416 | 27.396 |
| threads-sustained-v2/threads/notify-empty | 6.146 | 5.042 | 57.562 | 127.31 | 40.854 | 28.438 |
| threads-sustained-v2/threads/wait-mismatch | 6.083 | 5.083 | 53.480 | 121.69 | 39.542 | 25.791 |
| threads-sustained-v2/threads/wait-timeout | 5.938 | 5.729 | 57.751 | 125.71 | 43.312 | 30.562 |
| generic-calls-v1/core/generic-calls | 5.958 | 5.146 | 60.042 | 123.62 | 42.895 | 27.875 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 160.58 | 142.27 | 94.479 | 723.42 | 1,126.98 | 272,019.02 |
| original22-v1/core/integer-loop | 208.40 | 182.79 | 112.42 | 767.21 | 1,143.33 | 269,644.15 |
| original22-v1/core/function-calls | 282.27 | 248.29 | 126.38 | 783.17 | 1,224.40 | 284,345.35 |
| original22-v1/core/memory-load-store | 326.71 | 297.54 | 141.21 | 759.96 | 1,152.17 | 274,472.81 |
| original22-v1/memory/narrow-unaligned | 285.81 | 277.71 | 148.48 | 814.21 | 1,195.71 | 310,551.96 |
| original22-v1/memory/dependent-address | 285.29 | 267.00 | 147.54 | 780.46 | 1,238.48 | 312,733.81 |
| original22-v1/memory/float64-unaligned | 302.46 | 287.75 | 162.42 | 808.19 | 1,256.25 | 299,166.12 |
| original22-v1/runtime/simd-memory-full | 288.60 | 242.94 | 126.90 | 1,025.50 | 1,230.33 | 278,833.04 |
| original22-v1/runtime/simd-memory-widen | 234.83 | 227.35 | 124.19 | 1,000.02 | 1,114.06 | 276,288.12 |
| original22-v1/runtime/simd-memory-lanes | 280.46 | 244.77 | 127.79 | 983.02 | 1,171.67 | 268,618.23 |
| original22-v1/gc/i31-roundtrip | 177.40 | 158.65 | US | 727.42 | US | 277,449.29 |
| original22-v1/gc/struct-access | 224.46 | 185.33 | US | 921.35 | US | 280,473.19 |
| original22-v1/gc/struct-allocation | 224.04 | 179.81 | US | 1,114.96 | US | 270,384.63 |
| original22-v1/gc/struct-indirect-access | 256.08 | 245.38 | US | 1,506.31 | US | 271,381.92 |
| original22-v1/threads/private-memory | 226.79 | 177.35 | 118.27 | 694.96 | 1,097.29 | 267,821.06 |
| original22-v1/threads/shared-memory | 221.29 | 186.29 | 120.90 | 718.81 | 1,136.50 | 270,471.17 |
| original22-v1/threads/atomic-rmw | 206.46 | 175.83 | 131.04 | 787.75 | 1,136.25 | 273,599.08 |
| original22-v1/threads/notify-empty | 214.54 | 186.40 | 134.65 | 811.81 | 1,070.12 | 277,053.98 |
| original22-v1/threads/wait-mismatch | 208.10 | 174.98 | 127.96 | 839.08 | 1,166.10 | 274,658.94 |
| original22-v1/threads/wait-timeout | 230.31 | 184.27 | 122.44 | 821.96 | 1,177.81 | 269,180.90 |
| original22-v1/applications/sha256-clang | 1,461.52 | 1,752.06 | 323.88 | 1,265.52 | 2,043.29 | 277,111.12 |
| original22-v1/applications/sha256-rust | 1,325.71 | 1,677.23 | 313.00 | 1,299.88 | 1,746.88 | 272,115.42 |
| gc-sustained-v1/gc/i31-roundtrip | 176.56 | 153.15 | US | 794.17 | US | 268,357.98 |
| gc-sustained-v1/gc/struct-access | 234.33 | 187.10 | US | 860.62 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 223.54 | 178.90 | US | 1,042.90 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 261.48 | 249.08 | US | 1,489.83 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 252.56 | 184.35 | 128.88 | 722.77 | 1,124.38 | 272,764.56 |
| threads-sustained-v2/threads/shared-memory | 213.21 | 186.69 | 126.56 | 726.25 | 1,108.08 | 275,084.17 |
| threads-sustained-v2/threads/atomic-rmw | 214.81 | 171.81 | 118.81 | 712.08 | 1,134.69 | 265,451.40 |
| threads-sustained-v2/threads/notify-empty | 203.54 | 172.98 | 132.44 | 900.10 | 1,172.62 | 269,513.88 |
| threads-sustained-v2/threads/wait-mismatch | 223.67 | 178.19 | 123.83 | 833.31 | 1,105.27 | 273,785.21 |
| threads-sustained-v2/threads/wait-timeout | 222.19 | 181.10 | 128.29 | 864.40 | 1,196.10 | 270,741.35 |
| generic-calls-v1/core/generic-calls | 295.90 | 268.23 | 132.06 | 906.25 | 1,215.42 | 275,716.21 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.438 | 8.354 | 14.708 | 7.771 | 17.125 | 15.875 |
| original22-v1/core/integer-loop | 6.146 | 7.667 | 13.854 | 6.688 | 17.854 | 14.604 |
| original22-v1/core/function-calls | 6.917 | 7.688 | 15.188 | 6.771 | 18.188 | 16.000 |
| original22-v1/core/memory-load-store | 11.041 | 12.374 | 18.584 | 21.020 | 31.291 | 36.249 |
| original22-v1/memory/narrow-unaligned | 11.625 | 12.875 | 17.541 | 20.646 | 30.584 | 34.750 |
| original22-v1/memory/dependent-address | 10.688 | 12.041 | 19.438 | 21.250 | 32.312 | 39.042 |
| original22-v1/memory/float64-unaligned | 12.062 | 13.208 | 18.666 | 21.083 | 31.291 | 40.105 |
| original22-v1/runtime/simd-memory-full | 12.333 | 13.896 | 17.188 | 60.875 | 36.666 | 44.916 |
| original22-v1/runtime/simd-memory-widen | 12.396 | 12.813 | 18.938 | 62.854 | 34.500 | 46.125 |
| original22-v1/runtime/simd-memory-lanes | 11.209 | 13.041 | 18.688 | 56.813 | 36.521 | 45.812 |
| original22-v1/gc/i31-roundtrip | 6.250 | 8.500 | US | 7.708 | US | 17.604 |
| original22-v1/gc/struct-access | 7.250 | 8.709 | US | 23.271 | US | 16.355 |
| original22-v1/gc/struct-allocation | 7.729 | 9.312 | US | 54.541 | US | 21.562 |
| original22-v1/gc/struct-indirect-access | 19.209 | 20.959 | US | 64.083 | US | 25.188 |
| original22-v1/threads/private-memory | 10.625 | 11.167 | 17.834 | 20.145 | 28.625 | 36.000 |
| original22-v1/threads/shared-memory | 10.750 | 11.500 | 18.292 | 20.688 | 37.000 | 44.354 |
| original22-v1/threads/atomic-rmw | 11.188 | 12.541 | 18.355 | 21.291 | 34.791 | 37.146 |
| original22-v1/threads/notify-empty | 10.688 | 13.709 | 18.791 | 20.875 | 34.312 | 56.562 |
| original22-v1/threads/wait-mismatch | 10.166 | 11.292 | 17.895 | 19.938 | 35.542 | 41.312 |
| original22-v1/threads/wait-timeout | 12.500 | 12.396 | 16.146 | 21.292 | 36.980 | 36.396 |
| original22-v1/applications/sha256-clang | 14.000 | 17.312 | 18.583 | 55.895 | 40.250 | 49.000 |
| original22-v1/applications/sha256-rust | 12.771 | 15.376 | 25.105 | 54.145 | 35.375 | 42.688 |
| gc-sustained-v1/gc/i31-roundtrip | 6.062 | 8.000 | US | 8.041 | US | 16.020 |
| gc-sustained-v1/gc/struct-access | 7.646 | 8.709 | US | 20.729 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 7.271 | 9.209 | US | 50.876 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 19.459 | 21.416 | US | 63.979 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 11.416 | 12.000 | 18.209 | 20.875 | 32.084 | 36.438 |
| threads-sustained-v2/threads/shared-memory | 10.833 | 11.688 | 19.229 | 21.270 | 33.854 | 36.271 |
| threads-sustained-v2/threads/atomic-rmw | 11.312 | 12.146 | 17.459 | 20.166 | 34.771 | 39.292 |
| threads-sustained-v2/threads/notify-empty | 11.125 | 12.521 | 16.645 | 22.230 | 36.667 | 37.166 |
| threads-sustained-v2/threads/wait-mismatch | 11.084 | 12.188 | 18.562 | 20.396 | 35.104 | 37.020 |
| threads-sustained-v2/threads/wait-timeout | 11.396 | 13.083 | 16.459 | 21.645 | 38.541 | 38.937 |
| generic-calls-v1/core/generic-calls | 7.562 | 10.354 | 15.124 | 8.000 | 17.959 | 18.875 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.271 | 0.334 | 2.021 | 3.812 | 4.771 | 0.291 |
| original22-v1/core/integer-loop | 0.250 | 0.333 | 1.812 | 3.854 | 4.646 | 0.271 |
| original22-v1/core/function-calls | 0.250 | 0.375 | 1.896 | 3.813 | 4.896 | 0.230 |
| original22-v1/core/memory-load-store | 0.250 | 0.396 | 2.583 | 4.146 | 4.812 | 0.292 |
| original22-v1/memory/narrow-unaligned | 0.250 | 0.458 | 1.792 | 4.250 | 4.563 | 0.250 |
| original22-v1/memory/dependent-address | 0.271 | 0.375 | 2.188 | 4.083 | 5.000 | 0.291 |
| original22-v1/memory/float64-unaligned | 0.250 | 0.396 | 2.000 | 4.083 | 4.958 | 0.208 |
| original22-v1/runtime/simd-memory-full | 0.250 | 0.416 | 1.854 | 1.938 | 4.854 | 0.271 |
| original22-v1/runtime/simd-memory-widen | 0.292 | 0.375 | 2.062 | 1.854 | 4.500 | 0.312 |
| original22-v1/runtime/simd-memory-lanes | 0.250 | 0.375 | 2.208 | 1.895 | 4.500 | 0.271 |
| original22-v1/gc/i31-roundtrip | 0.230 | 0.416 | US | 3.833 | US | 0.291 |
| original22-v1/gc/struct-access | 0.209 | 0.354 | US | 3.812 | US | 0.250 |
| original22-v1/gc/struct-allocation | 0.250 | 0.375 | US | 1.750 | US | 0.291 |
| original22-v1/gc/struct-indirect-access | 0.312 | 0.396 | US | 1.938 | US | 0.271 |
| original22-v1/threads/private-memory | 0.250 | 0.354 | 2.042 | 4.021 | 4.250 | 0.208 |
| original22-v1/threads/shared-memory | 0.271 | 0.438 | 2.542 | 3.708 | 4.562 | 0.271 |
| original22-v1/threads/atomic-rmw | 0.291 | 0.396 | 2.667 | 4.458 | 4.625 | 0.208 |
| original22-v1/threads/notify-empty | 0.291 | 0.459 | 1.958 | 3.979 | 4.438 | 0.333 |
| original22-v1/threads/wait-mismatch | 0.250 | 0.375 | 2.688 | 3.875 | 4.937 | 0.271 |
| original22-v1/threads/wait-timeout | 0.291 | 0.396 | 2.292 | 3.896 | 4.792 | 0.250 |
| original22-v1/applications/sha256-clang | 0.250 | 0.437 | 2.896 | 2.125 | 5.646 | 0.291 |
| original22-v1/applications/sha256-rust | 0.229 | 0.375 | 2.625 | 1.896 | 4.813 | 0.208 |
| gc-sustained-v1/gc/i31-roundtrip | 0.250 | 0.334 | US | 4.271 | US | 0.250 |
| gc-sustained-v1/gc/struct-access | 0.208 | 0.375 | US | 3.667 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 0.250 | 0.333 | US | 1.667 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 0.271 | 0.396 | US | 2.000 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.250 | 0.375 | 1.896 | 4.167 | 5.188 | 0.292 |
| threads-sustained-v2/threads/shared-memory | 0.250 | 0.375 | 2.521 | 4.062 | 4.333 | 0.188 |
| threads-sustained-v2/threads/atomic-rmw | 0.250 | 0.334 | 2.271 | 3.896 | 4.688 | 0.291 |
| threads-sustained-v2/threads/notify-empty | 0.291 | 0.375 | 1.855 | 4.250 | 4.688 | 0.271 |
| threads-sustained-v2/threads/wait-mismatch | 0.250 | 0.396 | 2.292 | 3.896 | 4.688 | 0.208 |
| threads-sustained-v2/threads/wait-timeout | 0.250 | 0.375 | 2.271 | 4.104 | 4.833 | 0.250 |
| generic-calls-v1/core/generic-calls | 0.291 | 0.375 | 2.500 | 4.083 | 4.875 | 0.333 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 9.667 | 7.958 | 8.709 | 27.062 | 16.209 | 14.563 |
| original22-v1/core/integer-loop | 1,619.31 | 1,576.38 | 1,566.35 | 1,650.46 | 1,679.69 | 1,657.00 |
| original22-v1/core/function-calls | 788.50 | 800.83 | 1,305.29 | 938.04 | 776.77 | 795.44 |
| original22-v1/core/memory-load-store | 448.35 | 456.65 | 809.44 | 542.31 | 584.04 | 312.21 |
| original22-v1/memory/narrow-unaligned | 660.12 | 658.00 | 1,004.10 | 773.17 | 750.10 | 533.81 |
| original22-v1/memory/dependent-address | 1,860.50 | 1,856.25 | 3,062.65 | 2,131.81 | 2,171.83 | 1,549.67 |
| original22-v1/memory/float64-unaligned | 749.25 | 749.81 | 807.58 | 776.60 | 738.35 | 736.73 |
| original22-v1/runtime/simd-memory-full | 3,321.88 | 3,384.00 | 3,561.21 | 3,604.33 | 3,530.48 | 3,641.27 |
| original22-v1/runtime/simd-memory-widen | 1,051.71 | 1,039.92 | 1,054.23 | 1,034.10 | 1,043.35 | 1,096.92 |
| original22-v1/runtime/simd-memory-lanes | 4,910.06 | 4,979.73 | 5,828.42 | 5,778.08 | 5,693.98 | 5,078.50 |
| original22-v1/gc/i31-roundtrip | 23.730 | 22.104 | US | 49.875 | US | 29.437 |
| original22-v1/gc/struct-access | 35.916 | 32.459 | US | 51.730 | US | 131.19 |
| original22-v1/gc/struct-allocation | 361.65 | 376.58 | US | 137.29 | US | 537.77 |
| original22-v1/gc/struct-indirect-access | 29.458 | 24.709 | US | 30.125 | US | 331.19 |
| original22-v1/threads/private-memory | 14.521 | 11.646 | 18.812 | 43.146 | 24.417 | 31.166 |
| original22-v1/threads/shared-memory | 26.042 | 24.105 | 17.980 | 46.791 | 27.771 | 33.500 |
| original22-v1/threads/atomic-rmw | 32.563 | 28.021 | 28.479 | 54.500 | 56.459 | 38.917 |
| original22-v1/threads/notify-empty | 22.625 | 18.980 | 222.31 | 100.77 | 63.917 | 129.77 |
| original22-v1/threads/wait-mismatch | 23.229 | 19.896 | 260.35 | 120.19 | 41.270 | 55.625 |
| original22-v1/threads/wait-timeout | 24.145 | 18.687 | 2,638.54 | 466.75 | 1,848.46 | 657.56 |
| original22-v1/applications/sha256-clang | 1,179.75 | 1,146.23 | 1,390.40 | 1,006.94 | 1,044.71 | 988.04 |
| original22-v1/applications/sha256-rust | 1,162.88 | 1,151.02 | 1,478.60 | 924.19 | 939.10 | 943.65 |
| gc-sustained-v1/gc/i31-roundtrip | 1,357.79 | 1,346.85 | US | 1,934.37 | US | 1,433.02 |
| gc-sustained-v1/gc/struct-access | 1,336.94 | 1,329.60 | US | 1,076.75 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 25,455.02 | 15,789.06 | US | 3,060.33 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1,687.23 | 1,681.56 | US | 2,592.04 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 38.312 | 34.896 | 119.02 | 133.77 | 119.38 | 124.75 |
| threads-sustained-v2/threads/shared-memory | 139.10 | 132.71 | 112.77 | 136.92 | 116.54 | 124.85 |
| threads-sustained-v2/threads/atomic-rmw | 197.40 | 190.31 | 187.90 | 211.56 | 407.35 | 212.56 |
| threads-sustained-v2/threads/notify-empty | 105.06 | 95.041 | 2,131.69 | 789.75 | 480.35 | 961.77 |
| threads-sustained-v2/threads/wait-mismatch | 119.19 | 113.15 | 2,440.42 | 911.85 | 252.31 | 337.19 |
| threads-sustained-v2/threads/wait-timeout | 105.15 | 95.812 | 27,407.44 | 4,336.83 | 17,059.75 | 6,546.25 |
| generic-calls-v1/core/generic-calls | 5,262.10 | 5,218.02 | 10,728.92 | 10,407.79 | 10,352.15 | 5,382.08 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 8.646 | 5.979 | 6.292 | 53.375 | 25.291 | 345.94 |
| original22-v1/core/integer-loop | 12.229 | 4.292 | 6.083 | 71.124 | 28.062 | 370.81 |
| original22-v1/core/function-calls | 15.271 | 8.396 | 7.167 | 96.500 | 41.271 | 380.56 |
| original22-v1/core/memory-load-store | 12.438 | 18.312 | 8.854 | 101.52 | 63.041 | 461.90 |
| original22-v1/memory/narrow-unaligned | 12.604 | 14.729 | 9.438 | 104.06 | 62.249 | 448.69 |
| original22-v1/memory/dependent-address | 15.876 | 17.562 | 8.416 | 89.980 | 57.291 | 421.29 |
| original22-v1/memory/float64-unaligned | 18.980 | 18.166 | 10.958 | 113.02 | 61.312 | 457.46 |
| original22-v1/runtime/simd-memory-full | 20.604 | 17.083 | 11.959 | 126.94 | 65.604 | 433.02 |
| original22-v1/runtime/simd-memory-widen | 11.583 | 11.896 | 3.979 | 99.438 | 51.312 | 416.15 |
| original22-v1/runtime/simd-memory-lanes | 13.437 | 11.542 | 9.834 | 84.396 | 59.271 | 397.04 |
| original22-v1/gc/i31-roundtrip | 9.125 | 4.604 | US | 63.938 | US | 373.79 |
| original22-v1/gc/struct-access | 15.271 | 11.834 | US | 81.874 | US | 374.46 |
| original22-v1/gc/struct-allocation | 132.38 | 96.855 | US | 118.75 | US | 1,968.96 |
| original22-v1/gc/struct-indirect-access | 11.979 | 13.896 | US | 96.521 | US | 363.21 |
| original22-v1/threads/private-memory | 13.500 | 16.833 | 6.604 | 85.604 | 52.521 | 420.94 |
| original22-v1/threads/shared-memory | 15.938 | 11.562 | 9.396 | 83.937 | 57.875 | 388.21 |
| original22-v1/threads/atomic-rmw | 11.208 | 8.125 | 3.771 | 71.021 | 53.334 | 406.08 |
| original22-v1/threads/notify-empty | 12.063 | 10.104 | 4.458 | 80.812 | 50.417 | 393.46 |
| original22-v1/threads/wait-mismatch | 11.938 | 11.208 | 4.208 | 85.645 | 54.938 | 396.48 |
| original22-v1/threads/wait-timeout | 11.396 | 14.438 | 2.917 | 94.042 | 54.063 | 389.67 |
| original22-v1/applications/sha256-clang | 18.021 | 18.271 | 11.021 | 179.17 | 77.584 | 451.35 |
| original22-v1/applications/sha256-rust | 17.396 | 14.229 | 5.292 | 141.75 | 68.751 | 441.40 |
| gc-sustained-v1/gc/i31-roundtrip | 9.500 | 4.500 | US | 64.562 | US | 401.62 |
| gc-sustained-v1/gc/struct-access | 18.104 | 15.021 | US | 100.65 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 2,977.31 | 2,025.27 | US | 725.96 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 16.062 | 16.916 | US | 130.17 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 16.020 | 13.500 | 5.104 | 83.312 | 55.417 | 427.27 |
| threads-sustained-v2/threads/shared-memory | 12.187 | 10.875 | 4.708 | 79.542 | 55.084 | 413.10 |
| threads-sustained-v2/threads/atomic-rmw | 13.812 | 11.375 | 4.979 | 95.895 | 56.334 | 404.65 |
| threads-sustained-v2/threads/notify-empty | 20.458 | 14.167 | 8.125 | 98.730 | 61.188 | 445.12 |
| threads-sustained-v2/threads/wait-mismatch | 15.979 | 10.458 | 5.979 | 78.687 | 53.354 | 404.25 |
| threads-sustained-v2/threads/wait-timeout | 21.875 | 17.791 | 3.562 | 104.35 | 63.562 | 428.71 |
| generic-calls-v1/core/generic-calls | 16.583 | 9.812 | 11.584 | 122.62 | 47.771 | 402.27 |

</details>

<details>
<summary>Interpreter: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 3,790.94 | 5,459.21 | 4,014.33 | NM | NM | 15,454.87 |
| original22-v1/core/integer-loop | 4,189.77 | 6,183.83 | 4,074.69 | NM | NM | 16,260.52 |
| original22-v1/core/function-calls | 4,027.56 | 5,855.58 | 3,997.50 | NM | NM | 16,372.46 |
| original22-v1/core/memory-load-store | 3,816.40 | 6,048.71 | 4,014.62 | NM | NM | 16,322.12 |
| original22-v1/memory/narrow-unaligned | 4,208.21 | 5,893.54 | 4,149.06 | NM | NM | 16,176.35 |
| original22-v1/memory/dependent-address | 4,298.62 | 6,143.71 | 4,125.33 | NM | NM | 16,130.79 |
| original22-v1/memory/float64-unaligned | 4,080.31 | 5,889.67 | 4,327.54 | NM | NM | 16,043.21 |
| original22-v1/runtime/simd-memory-full | 4,433.48 | 6,188.50 | 4,286.12 | NM | NM | 16,816.27 |
| original22-v1/runtime/simd-memory-widen | 3,880.08 | 5,757.29 | 3,931.79 | NM | NM | 15,642.90 |
| original22-v1/runtime/simd-memory-lanes | 4,067.10 | 6,090.73 | 4,237.17 | NM | NM | 15,822.81 |
| original22-v1/gc/i31-roundtrip | 4,076.02 | 5,672.56 | US | NM | NM | 15,950.56 |
| original22-v1/gc/struct-access | 3,793.08 | 5,594.23 | US | NM | NM | 15,906.98 |
| original22-v1/gc/struct-allocation | 3,958.40 | 5,701.27 | US | NM | NM | 15,756.27 |
| original22-v1/gc/struct-indirect-access | 4,012.69 | 5,925.21 | US | NM | NM | 15,998.21 |
| original22-v1/threads/private-memory | 3,903.42 | 5,559.58 | 3,965.42 | NM | NM | 16,054.73 |
| original22-v1/threads/shared-memory | 4,044.00 | 5,701.65 | 3,869.10 | NM | NM | 15,899.69 |
| original22-v1/threads/atomic-rmw | 3,971.58 | 5,617.15 | 4,096.35 | NM | NM | 15,503.10 |
| original22-v1/threads/notify-empty | 3,968.60 | 5,705.23 | 3,805.23 | NM | NM | 15,943.06 |
| original22-v1/threads/wait-mismatch | 3,951.17 | 5,254.52 | 3,992.21 | NM | NM | 15,699.77 |
| original22-v1/threads/wait-timeout | 3,832.19 | 5,493.27 | 3,923.38 | NM | NM | 15,726.23 |
| original22-v1/applications/sha256-clang | 4,181.75 | 6,044.94 | 4,223.46 | NM | NM | 16,077.56 |
| original22-v1/applications/sha256-rust | 4,029.02 | 5,688.83 | 4,309.85 | NM | NM | 15,792.67 |
| gc-sustained-v1/gc/i31-roundtrip | 3,999.83 | 6,086.35 | US | NM | NM | 16,253.10 |
| gc-sustained-v1/gc/struct-access | 4,082.62 | 5,794.46 | US | NM | NM | 15,802.96 |
| gc-sustained-v1/gc/struct-allocation | 4,123.71 | 5,818.40 | US | NM | NM | 15,878.83 |
| gc-sustained-v1/gc/struct-indirect-access | 4,073.69 | 5,616.71 | US | NM | NM | 16,034.85 |
| threads-sustained-v2/threads/private-memory | 3,925.83 | 5,916.44 | 4,170.90 | NM | NM | 15,528.65 |
| threads-sustained-v2/threads/shared-memory | 4,029.63 | 5,802.38 | 3,935.33 | NM | NM | 15,931.35 |
| threads-sustained-v2/threads/atomic-rmw | 3,855.21 | 5,581.40 | 4,171.60 | NM | NM | 16,252.65 |
| threads-sustained-v2/threads/notify-empty | 4,107.38 | 5,965.02 | 4,165.42 | NM | NM | 15,965.04 |
| threads-sustained-v2/threads/wait-mismatch | 3,845.38 | 5,817.75 | 4,176.62 | NM | NM | 15,822.13 |
| threads-sustained-v2/threads/wait-timeout | 3,938.12 | 6,137.10 | 4,229.27 | NM | NM | 16,245.25 |
| generic-calls-v1/core/generic-calls | 4,327.44 | 6,189.83 | 4,699.12 | NM | NM | 15,824.08 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 30.916 | 32.980 | 29.750 | NM | NM | 38.500 |
| original22-v1/core/integer-loop | 36.833 | 41.999 | 36.520 | NM | NM | 36.875 |
| original22-v1/core/function-calls | 34.729 | 40.666 | 29.479 | NM | NM | 39.166 |
| original22-v1/core/memory-load-store | 32.021 | 40.188 | 33.521 | NM | NM | 39.188 |
| original22-v1/memory/narrow-unaligned | 42.854 | 36.583 | 35.959 | NM | NM | 39.125 |
| original22-v1/memory/dependent-address | 39.875 | 37.124 | 35.105 | NM | NM | 41.980 |
| original22-v1/memory/float64-unaligned | 32.563 | 41.333 | 47.667 | NM | NM | 43.459 |
| original22-v1/runtime/simd-memory-full | 45.438 | 39.395 | 37.541 | NM | NM | 41.584 |
| original22-v1/runtime/simd-memory-widen | 37.938 | 32.500 | 31.583 | NM | NM | 41.812 |
| original22-v1/runtime/simd-memory-lanes | 38.521 | 41.959 | 38.854 | NM | NM | 44.041 |
| original22-v1/gc/i31-roundtrip | 33.833 | 34.041 | US | NM | NM | 38.979 |
| original22-v1/gc/struct-access | 32.542 | 33.208 | US | NM | NM | 42.271 |
| original22-v1/gc/struct-allocation | 33.771 | 37.416 | US | NM | NM | 37.646 |
| original22-v1/gc/struct-indirect-access | 36.980 | 38.062 | US | NM | NM | 39.500 |
| original22-v1/threads/private-memory | 29.729 | 36.666 | 36.522 | NM | NM | 40.417 |
| original22-v1/threads/shared-memory | 30.730 | 34.917 | 29.646 | NM | NM | 42.771 |
| original22-v1/threads/atomic-rmw | 34.084 | 37.376 | 30.729 | NM | NM | 34.833 |
| original22-v1/threads/notify-empty | 33.396 | 34.438 | 32.792 | NM | NM | 34.750 |
| original22-v1/threads/wait-mismatch | 27.229 | 33.041 | 32.812 | NM | NM | 36.083 |
| original22-v1/threads/wait-timeout | 33.916 | 34.605 | 29.125 | NM | NM | 35.770 |
| original22-v1/applications/sha256-clang | 43.249 | 44.625 | 40.000 | NM | NM | 47.063 |
| original22-v1/applications/sha256-rust | 36.480 | 37.188 | 37.938 | NM | NM | 36.583 |
| gc-sustained-v1/gc/i31-roundtrip | 30.229 | 42.896 | US | NM | NM | 37.479 |
| gc-sustained-v1/gc/struct-access | 36.209 | 37.646 | US | NM | NM | 40.083 |
| gc-sustained-v1/gc/struct-allocation | 30.625 | 36.916 | US | NM | NM | 39.271 |
| gc-sustained-v1/gc/struct-indirect-access | 38.667 | 37.229 | US | NM | NM | 40.625 |
| threads-sustained-v2/threads/private-memory | 37.229 | 38.709 | 38.396 | NM | NM | 38.750 |
| threads-sustained-v2/threads/shared-memory | 35.604 | 38.771 | 34.271 | NM | NM | 36.395 |
| threads-sustained-v2/threads/atomic-rmw | 35.855 | 40.416 | 32.521 | NM | NM | 40.562 |
| threads-sustained-v2/threads/notify-empty | 38.584 | 43.812 | 33.209 | NM | NM | 36.812 |
| threads-sustained-v2/threads/wait-mismatch | 33.812 | 35.812 | 34.396 | NM | NM | 40.979 |
| threads-sustained-v2/threads/wait-timeout | 31.646 | 40.541 | 30.541 | NM | NM | 39.270 |
| generic-calls-v1/core/generic-calls | 47.916 | 41.750 | 53.354 | NM | NM | 44.375 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4.667 | 5.209 | 2.124 | NM | NM | 27.104 |
| original22-v1/core/integer-loop | 5.062 | 4.667 | 2.583 | NM | NM | 24.980 |
| original22-v1/core/function-calls | 4.167 | 4.750 | 2.167 | NM | NM | 26.166 |
| original22-v1/core/memory-load-store | 4.417 | 5.083 | 2.458 | NM | NM | 27.104 |
| original22-v1/memory/narrow-unaligned | 5.104 | 5.042 | 2.062 | NM | NM | 27.979 |
| original22-v1/memory/dependent-address | 4.604 | 5.771 | 2.167 | NM | NM | 27.438 |
| original22-v1/memory/float64-unaligned | 4.646 | 4.979 | 2.750 | NM | NM | 28.125 |
| original22-v1/runtime/simd-memory-full | 5.250 | 4.875 | 2.062 | NM | NM | 27.480 |
| original22-v1/runtime/simd-memory-widen | 4.250 | 4.250 | 2.083 | NM | NM | 25.416 |
| original22-v1/runtime/simd-memory-lanes | 4.250 | 4.479 | 2.354 | NM | NM | 26.646 |
| original22-v1/gc/i31-roundtrip | 4.771 | 4.771 | US | NM | NM | 27.313 |
| original22-v1/gc/struct-access | 4.167 | 4.812 | US | NM | NM | 27.000 |
| original22-v1/gc/struct-allocation | 4.125 | 4.333 | US | NM | NM | 26.438 |
| original22-v1/gc/struct-indirect-access | 5.125 | 4.646 | US | NM | NM | 26.500 |
| original22-v1/threads/private-memory | 4.438 | 4.500 | 2.438 | NM | NM | 28.167 |
| original22-v1/threads/shared-memory | 4.854 | 4.625 | 2.042 | NM | NM | 26.625 |
| original22-v1/threads/atomic-rmw | 4.604 | 5.521 | 2.375 | NM | NM | 25.229 |
| original22-v1/threads/notify-empty | 4.750 | 4.771 | 1.958 | NM | NM | 25.666 |
| original22-v1/threads/wait-mismatch | 4.604 | 4.104 | 2.353 | NM | NM | 26.521 |
| original22-v1/threads/wait-timeout | 4.229 | 4.812 | 2.188 | NM | NM | 25.896 |
| original22-v1/applications/sha256-clang | 4.833 | 4.896 | 2.458 | NM | NM | 27.813 |
| original22-v1/applications/sha256-rust | 4.271 | 4.646 | 2.750 | NM | NM | 26.084 |
| gc-sustained-v1/gc/i31-roundtrip | 4.167 | 5.125 | US | NM | NM | 24.896 |
| gc-sustained-v1/gc/struct-access | 4.250 | 4.750 | US | NM | NM | 25.083 |
| gc-sustained-v1/gc/struct-allocation | 4.250 | 4.625 | US | NM | NM | 25.584 |
| gc-sustained-v1/gc/struct-indirect-access | 4.417 | 4.771 | US | NM | NM | 25.812 |
| threads-sustained-v2/threads/private-memory | 4.354 | 4.729 | 2.333 | NM | NM | 26.459 |
| threads-sustained-v2/threads/shared-memory | 4.312 | 4.917 | 2.208 | NM | NM | 25.104 |
| threads-sustained-v2/threads/atomic-rmw | 4.542 | 4.833 | 2.188 | NM | NM | 26.479 |
| threads-sustained-v2/threads/notify-empty | 4.646 | 4.354 | 2.375 | NM | NM | 27.645 |
| threads-sustained-v2/threads/wait-mismatch | 4.104 | 4.938 | 2.188 | NM | NM | 26.084 |
| threads-sustained-v2/threads/wait-timeout | 4.271 | 4.958 | 2.125 | NM | NM | 25.979 |
| generic-calls-v1/core/generic-calls | 4.646 | 5.792 | 2.771 | NM | NM | 26.520 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 56.438 | 46.730 | 35.271 | NM | NM | 38.334 |
| original22-v1/core/integer-loop | 70.770 | 61.062 | 43.750 | NM | NM | 39.459 |
| original22-v1/core/function-calls | 70.376 | 63.584 | 43.062 | NM | NM | 38.625 |
| original22-v1/core/memory-load-store | 79.895 | 74.896 | 43.917 | NM | NM | 43.833 |
| original22-v1/memory/narrow-unaligned | 86.584 | 80.999 | 45.938 | NM | NM | 43.708 |
| original22-v1/memory/dependent-address | 83.146 | 72.459 | 45.438 | NM | NM | 43.355 |
| original22-v1/memory/float64-unaligned | 77.063 | 77.708 | 53.104 | NM | NM | 42.771 |
| original22-v1/runtime/simd-memory-full | 75.729 | 64.666 | 47.688 | NM | NM | 42.291 |
| original22-v1/runtime/simd-memory-widen | 70.541 | 62.166 | 45.562 | NM | NM | 41.583 |
| original22-v1/runtime/simd-memory-lanes | 75.562 | 67.126 | 47.291 | NM | NM | 44.834 |
| original22-v1/gc/i31-roundtrip | 61.709 | 50.188 | US | NM | NM | 40.209 |
| original22-v1/gc/struct-access | 66.999 | 55.646 | US | NM | NM | 46.646 |
| original22-v1/gc/struct-allocation | 68.333 | 56.375 | US | NM | NM | 47.166 |
| original22-v1/gc/struct-indirect-access | 81.438 | 72.563 | US | NM | NM | 49.062 |
| original22-v1/threads/private-memory | 61.063 | 52.625 | 41.999 | NM | NM | 40.624 |
| original22-v1/threads/shared-memory | 63.938 | 52.584 | 39.959 | NM | NM | 39.728 |
| original22-v1/threads/atomic-rmw | 63.167 | 53.355 | 44.521 | NM | NM | 39.625 |
| original22-v1/threads/notify-empty | 66.376 | 53.104 | 41.333 | NM | NM | 38.000 |
| original22-v1/threads/wait-mismatch | 62.395 | 51.334 | 42.666 | NM | NM | 40.916 |
| original22-v1/threads/wait-timeout | 61.833 | 53.646 | 41.875 | NM | NM | 39.813 |
| original22-v1/applications/sha256-clang | 274.08 | 353.81 | 82.270 | NM | NM | 80.895 |
| original22-v1/applications/sha256-rust | 265.94 | 336.77 | 82.771 | NM | NM | 80.145 |
| gc-sustained-v1/gc/i31-roundtrip | 63.104 | 54.667 | US | NM | NM | 37.291 |
| gc-sustained-v1/gc/struct-access | 69.709 | 56.626 | US | NM | NM | 42.105 |
| gc-sustained-v1/gc/struct-allocation | 65.166 | 54.584 | US | NM | NM | 47.041 |
| gc-sustained-v1/gc/struct-indirect-access | 79.270 | 70.251 | US | NM | NM | 49.416 |
| threads-sustained-v2/threads/private-memory | 62.188 | 57.791 | 42.584 | NM | NM | 38.750 |
| threads-sustained-v2/threads/shared-memory | 62.688 | 51.104 | 41.166 | NM | NM | 39.438 |
| threads-sustained-v2/threads/atomic-rmw | 61.333 | 50.938 | 42.500 | NM | NM | 38.146 |
| threads-sustained-v2/threads/notify-empty | 65.416 | 52.041 | 43.105 | NM | NM | 41.479 |
| threads-sustained-v2/threads/wait-mismatch | 61.291 | 51.313 | 40.125 | NM | NM | 40.730 |
| threads-sustained-v2/threads/wait-timeout | 62.667 | 53.895 | 40.001 | NM | NM | 41.042 |
| generic-calls-v1/core/generic-calls | 76.854 | 69.417 | 42.166 | NM | NM | 44.730 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.667 | 6.917 | 15.709 | NM | NM | 24.230 |
| original22-v1/core/integer-loop | 7.062 | 7.125 | 16.417 | NM | NM | 23.396 |
| original22-v1/core/function-calls | 7.041 | 7.146 | 14.459 | NM | NM | 24.021 |
| original22-v1/core/memory-load-store | 9.979 | 11.251 | 16.500 | NM | NM | 46.709 |
| original22-v1/memory/narrow-unaligned | 11.084 | 11.458 | 16.459 | NM | NM | 49.291 |
| original22-v1/memory/dependent-address | 10.229 | 12.604 | 16.541 | NM | NM | 48.666 |
| original22-v1/memory/float64-unaligned | 9.459 | 12.166 | 17.334 | NM | NM | 47.938 |
| original22-v1/runtime/simd-memory-full | 11.646 | 11.396 | 17.396 | NM | NM | 54.124 |
| original22-v1/runtime/simd-memory-widen | 10.500 | 10.771 | 16.355 | NM | NM | 49.876 |
| original22-v1/runtime/simd-memory-lanes | 10.875 | 11.646 | 16.979 | NM | NM | 53.354 |
| original22-v1/gc/i31-roundtrip | 6.833 | 7.021 | US | NM | NM | 27.604 |
| original22-v1/gc/struct-access | 6.854 | 8.063 | US | NM | NM | 28.500 |
| original22-v1/gc/struct-allocation | 7.542 | 8.459 | US | NM | NM | 29.084 |
| original22-v1/gc/struct-indirect-access | 14.354 | 15.021 | US | NM | NM | 31.771 |
| original22-v1/threads/private-memory | 9.062 | 10.646 | 16.458 | NM | NM | 46.750 |
| original22-v1/threads/shared-memory | 12.479 | 12.583 | 16.417 | NM | NM | 45.479 |
| original22-v1/threads/atomic-rmw | 12.583 | 12.792 | 17.562 | NM | NM | 46.292 |
| original22-v1/threads/notify-empty | 13.291 | 11.812 | 15.938 | NM | NM | 46.709 |
| original22-v1/threads/wait-mismatch | 11.917 | 12.562 | 17.854 | NM | NM | 44.270 |
| original22-v1/threads/wait-timeout | 12.771 | 11.771 | 16.021 | NM | NM | 43.667 |
| original22-v1/applications/sha256-clang | 12.500 | 17.625 | 18.958 | NM | NM | 57.291 |
| original22-v1/applications/sha256-rust | 12.312 | 70.771 | 17.667 | NM | NM | 54.959 |
| gc-sustained-v1/gc/i31-roundtrip | 7.062 | 7.375 | US | NM | NM | 22.541 |
| gc-sustained-v1/gc/struct-access | 7.354 | 8.084 | US | NM | NM | 22.188 |
| gc-sustained-v1/gc/struct-allocation | 7.083 | 8.416 | US | NM | NM | 30.875 |
| gc-sustained-v1/gc/struct-indirect-access | 13.458 | 14.250 | US | NM | NM | 31.333 |
| threads-sustained-v2/threads/private-memory | 9.750 | 11.521 | 16.604 | NM | NM | 45.459 |
| threads-sustained-v2/threads/shared-memory | 12.479 | 12.041 | 17.041 | NM | NM | 45.708 |
| threads-sustained-v2/threads/atomic-rmw | 13.271 | 11.729 | 16.605 | NM | NM | 45.542 |
| threads-sustained-v2/threads/notify-empty | 13.167 | 11.959 | 16.605 | NM | NM | 45.104 |
| threads-sustained-v2/threads/wait-mismatch | 12.334 | 11.375 | 16.000 | NM | NM | 46.687 |
| threads-sustained-v2/threads/wait-timeout | 13.083 | 12.417 | 17.291 | NM | NM | 45.980 |
| generic-calls-v1/core/generic-calls | 7.791 | 9.041 | 16.479 | NM | NM | 26.125 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.250 | 0.333 | 0.208 | NM | NM | 0.209 |
| original22-v1/core/integer-loop | 0.229 | 0.375 | 0.208 | NM | NM | 0.188 |
| original22-v1/core/function-calls | 0.250 | 0.375 | 0.188 | NM | NM | 0.208 |
| original22-v1/core/memory-load-store | 0.250 | 0.375 | 0.208 | NM | NM | 0.208 |
| original22-v1/memory/narrow-unaligned | 0.291 | 0.334 | 0.208 | NM | NM | 0.230 |
| original22-v1/memory/dependent-address | 0.250 | 0.396 | 0.188 | NM | NM | 0.208 |
| original22-v1/memory/float64-unaligned | 0.208 | 0.354 | 0.188 | NM | NM | 0.208 |
| original22-v1/runtime/simd-memory-full | 0.271 | 0.312 | 0.166 | NM | NM | 0.209 |
| original22-v1/runtime/simd-memory-widen | 0.250 | 0.333 | 0.188 | NM | NM | 0.208 |
| original22-v1/runtime/simd-memory-lanes | 0.271 | 0.375 | 0.167 | NM | NM | 0.208 |
| original22-v1/gc/i31-roundtrip | 0.209 | 0.375 | US | NM | NM | 0.208 |
| original22-v1/gc/struct-access | 0.209 | 0.312 | US | NM | NM | 0.208 |
| original22-v1/gc/struct-allocation | 0.208 | 0.312 | US | NM | NM | 0.208 |
| original22-v1/gc/struct-indirect-access | 0.230 | 0.333 | US | NM | NM | 0.208 |
| original22-v1/threads/private-memory | 0.209 | 0.375 | 0.167 | NM | NM | 0.209 |
| original22-v1/threads/shared-memory | 0.291 | 0.437 | 0.167 | NM | NM | 0.230 |
| original22-v1/threads/atomic-rmw | 0.208 | 0.396 | 0.188 | NM | NM | 0.208 |
| original22-v1/threads/notify-empty | 0.250 | 0.438 | 0.188 | NM | NM | 0.271 |
| original22-v1/threads/wait-mismatch | 0.250 | 0.354 | 0.208 | NM | NM | 0.230 |
| original22-v1/threads/wait-timeout | 0.250 | 0.375 | 0.188 | NM | NM | 0.208 |
| original22-v1/applications/sha256-clang | 0.250 | 0.333 | 0.208 | NM | NM | 0.250 |
| original22-v1/applications/sha256-rust | 0.230 | 0.375 | 0.208 | NM | NM | 0.208 |
| gc-sustained-v1/gc/i31-roundtrip | 0.250 | 0.417 | US | NM | NM | 0.230 |
| gc-sustained-v1/gc/struct-access | 0.229 | 0.333 | US | NM | NM | 0.208 |
| gc-sustained-v1/gc/struct-allocation | 0.208 | 0.333 | US | NM | NM | 0.208 |
| gc-sustained-v1/gc/struct-indirect-access | 0.208 | 0.416 | US | NM | NM | 0.208 |
| threads-sustained-v2/threads/private-memory | 0.271 | 0.375 | 0.167 | NM | NM | 0.209 |
| threads-sustained-v2/threads/shared-memory | 0.291 | 0.375 | 0.167 | NM | NM | 0.230 |
| threads-sustained-v2/threads/atomic-rmw | 0.230 | 0.375 | 0.208 | NM | NM | 0.208 |
| threads-sustained-v2/threads/notify-empty | 0.250 | 0.375 | 0.167 | NM | NM | 0.208 |
| threads-sustained-v2/threads/wait-mismatch | 0.250 | 0.396 | 0.208 | NM | NM | 0.271 |
| threads-sustained-v2/threads/wait-timeout | 0.250 | 0.375 | 0.167 | NM | NM | 0.208 |
| generic-calls-v1/core/generic-calls | 0.250 | 0.354 | 0.291 | NM | NM | 0.188 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.833 | 7.417 | 3.562 | NM | NM | 8.271 |
| original22-v1/core/integer-loop | 31,466.04 | 26,203.94 | 83,525.00 | NM | NM | 51,552.35 |
| original22-v1/core/function-calls | 49,930.79 | 45,914.65 | 87,999.69 | NM | NM | 62,730.58 |
| original22-v1/core/memory-load-store | 67,012.27 | 59,715.71 | 146,996.04 | NM | NM | 91,793.38 |
| original22-v1/memory/narrow-unaligned | 88,664.73 | 78,119.44 | 189,608.02 | NM | NM | 115,418.00 |
| original22-v1/memory/dependent-address | 41,962.96 | 35,826.96 | 97,526.17 | NM | NM | 54,056.19 |
| original22-v1/memory/float64-unaligned | 69,492.50 | 62,140.06 | 140,718.54 | NM | NM | 85,322.00 |
| original22-v1/runtime/simd-memory-full | 76,663.19 | 66,624.94 | 151,108.62 | NM | NM | 78,135.54 |
| original22-v1/runtime/simd-memory-widen | 62,544.21 | 56,081.58 | 143,407.92 | NM | NM | 60,987.04 |
| original22-v1/runtime/simd-memory-lanes | 94,569.92 | 103,364.94 | 187,583.00 | NM | NM | 90,938.00 |
| original22-v1/gc/i31-roundtrip | 378.17 | 330.27 | US | NM | NM | 460.23 |
| original22-v1/gc/struct-access | 437.42 | 417.60 | US | NM | NM | 561.29 |
| original22-v1/gc/struct-allocation | 923.10 | 1,006.17 | US | NM | NM | 1,197.06 |
| original22-v1/gc/struct-indirect-access | 795.40 | 732.73 | US | NM | NM | 961.48 |
| original22-v1/threads/private-memory | 397.46 | 345.77 | 639.25 | NM | NM | 404.27 |
| original22-v1/threads/shared-memory | 745.67 | 363.42 | 616.33 | NM | NM | 415.10 |
| original22-v1/threads/atomic-rmw | 522.02 | 362.69 | 471.31 | NM | NM | 332.92 |
| original22-v1/threads/notify-empty | 310.23 | 289.06 | 427.60 | NM | NM | 402.90 |
| original22-v1/threads/wait-mismatch | 566.46 | 359.77 | 703.42 | NM | NM | 440.94 |
| original22-v1/threads/wait-timeout | 542.92 | 363.46 | 3,189.02 | NM | NM | 1,048.52 |
| original22-v1/applications/sha256-clang | 97,786.69 | 83,512.73 | 214,444.83 | NM | NM | 136,718.15 |
| original22-v1/applications/sha256-rust | 90,181.52 | 78,965.35 | 205,270.56 | NM | NM | 122,035.81 |
| gc-sustained-v1/gc/i31-roundtrip | 36,107.12 | 31,505.00 | US | NM | NM | 44,992.17 |
| gc-sustained-v1/gc/struct-access | 40,850.08 | 39,439.08 | US | NM | NM | 54,639.50 |
| gc-sustained-v1/gc/struct-allocation | 49,499.25 | 34,634.08 | US | NM | NM | 31,620.04 |
| gc-sustained-v1/gc/struct-indirect-access | 74,237.29 | 70,293.29 | US | NM | NM | 95,823.58 |
| threads-sustained-v2/threads/private-memory | 3,999.71 | 3,466.44 | 6,396.96 | NM | NM | 3,943.13 |
| threads-sustained-v2/threads/shared-memory | 7,398.92 | 3,549.33 | 6,301.10 | NM | NM | 3,981.38 |
| threads-sustained-v2/threads/atomic-rmw | 5,172.10 | 3,310.02 | 4,679.77 | NM | NM | 3,234.69 |
| threads-sustained-v2/threads/notify-empty | 3,189.35 | 2,768.92 | 4,153.96 | NM | NM | 4,009.69 |
| threads-sustained-v2/threads/wait-mismatch | 5,598.25 | 3,654.02 | 6,792.96 | NM | NM | 4,316.04 |
| threads-sustained-v2/threads/wait-timeout | 5,616.83 | 3,555.79 | 32,237.56 | NM | NM | 10,661.42 |
| generic-calls-v1/core/generic-calls | 552,521.48 | 508,994.52 | 924,691.12 | NM | NM | 598,319.15 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.000 | 1.438 | 9.834 | NM | NM | 5.458 |
| original22-v1/core/integer-loop | 5.958 | 2.042 | 5.125 | NM | NM | 6.312 |
| original22-v1/core/function-calls | 7.333 | 2.104 | 4.042 | NM | NM | 9.542 |
| original22-v1/core/memory-load-store | 7.688 | 7.042 | 8.166 | NM | NM | 50.709 |
| original22-v1/memory/narrow-unaligned | 6.458 | 5.979 | 6.917 | NM | NM | 48.542 |
| original22-v1/memory/dependent-address | 6.833 | 6.188 | 5.979 | NM | NM | 47.688 |
| original22-v1/memory/float64-unaligned | 6.375 | 6.125 | 9.938 | NM | NM | 52.541 |
| original22-v1/runtime/simd-memory-full | 6.917 | 6.896 | 8.688 | NM | NM | 51.688 |
| original22-v1/runtime/simd-memory-widen | 6.792 | 7.000 | 5.479 | NM | NM | 49.709 |
| original22-v1/runtime/simd-memory-lanes | 6.729 | 6.208 | 5.750 | NM | NM | 48.230 |
| original22-v1/gc/i31-roundtrip | 4.312 | 1.605 | US | NM | NM | 6.396 |
| original22-v1/gc/struct-access | 2.146 | 1.458 | US | NM | NM | 5.812 |
| original22-v1/gc/struct-allocation | 92.853 | 52.563 | US | NM | NM | 1,614.65 |
| original22-v1/gc/struct-indirect-access | 1.833 | 1.417 | US | NM | NM | 7.208 |
| original22-v1/threads/private-memory | 4.917 | 5.021 | 8.000 | NM | NM | 44.605 |
| original22-v1/threads/shared-memory | 16.895 | 10.521 | 7.646 | NM | NM | 41.229 |
| original22-v1/threads/atomic-rmw | 13.167 | 8.104 | 3.708 | NM | NM | 44.375 |
| original22-v1/threads/notify-empty | 11.750 | 7.208 | 3.979 | NM | NM | 40.604 |
| original22-v1/threads/wait-mismatch | 11.000 | 12.500 | 8.229 | NM | NM | 43.459 |
| original22-v1/threads/wait-timeout | 17.146 | 10.125 | 3.208 | NM | NM | 41.000 |
| original22-v1/applications/sha256-clang | 7.125 | 6.688 | 7.062 | NM | NM | 60.333 |
| original22-v1/applications/sha256-rust | 7.688 | 2.125 | 8.834 | NM | NM | 56.875 |
| gc-sustained-v1/gc/i31-roundtrip | 5.729 | 2.104 | US | NM | NM | 7.958 |
| gc-sustained-v1/gc/struct-access | 2.458 | 2.208 | US | NM | NM | 8.312 |
| gc-sustained-v1/gc/struct-allocation | 2,688.31 | 1,626.15 | US | NM | NM | 57,725.75 |
| gc-sustained-v1/gc/struct-indirect-access | 2.188 | 2.104 | US | NM | NM | 8.459 |
| threads-sustained-v2/threads/private-memory | 6.604 | 6.812 | 6.646 | NM | NM | 46.395 |
| threads-sustained-v2/threads/shared-memory | 11.979 | 10.104 | 4.417 | NM | NM | 44.646 |
| threads-sustained-v2/threads/atomic-rmw | 14.209 | 9.250 | 6.188 | NM | NM | 45.230 |
| threads-sustained-v2/threads/notify-empty | 19.812 | 12.271 | 7.188 | NM | NM | 47.854 |
| threads-sustained-v2/threads/wait-mismatch | 14.729 | 10.459 | 4.063 | NM | NM | 50.084 |
| threads-sustained-v2/threads/wait-timeout | 21.292 | 15.708 | 3.604 | NM | NM | 48.188 |
| generic-calls-v1/core/generic-calls | 6.646 | 2.083 | 5.917 | NM | NM | 10.626 |

</details>

### WasmEdge AOT preparation detail

| Workload | Compile + validation + artifact I/O, us | Native image load, us | Post-load validation, us |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 16,924.33 | 254,257.54 | 3.625 |
| original22-v1/core/integer-loop | 17,256.52 | 252,929.06 | 5.146 |
| original22-v1/core/function-calls | 17,921.29 | 266,575.15 | 4.708 |
| original22-v1/core/memory-load-store | 20,157.79 | 252,659.29 | 6.375 |
| original22-v1/memory/narrow-unaligned | 19,928.58 | 286,161.88 | 7.979 |
| original22-v1/memory/dependent-address | 19,178.19 | 289,243.67 | 9.854 |
| original22-v1/memory/float64-unaligned | 21,203.15 | 276,783.48 | 7.792 |
| original22-v1/runtime/simd-memory-full | 18,193.17 | 256,110.35 | 7.958 |
| original22-v1/runtime/simd-memory-widen | 17,468.58 | 259,094.15 | 6.667 |
| original22-v1/runtime/simd-memory-lanes | 17,557.29 | 251,440.67 | 7.167 |
| original22-v1/gc/i31-roundtrip | 18,820.35 | 253,187.29 | 6.792 |
| original22-v1/gc/struct-access | 18,247.12 | 252,452.67 | 4.500 |
| original22-v1/gc/struct-allocation | 17,742.71 | 252,544.10 | 7.083 |
| original22-v1/gc/struct-indirect-access | 18,561.02 | 252,971.23 | 7.646 |
| original22-v1/threads/private-memory | 16,632.92 | 250,646.00 | 6.063 |
| original22-v1/threads/shared-memory | 16,826.48 | 253,407.71 | 6.625 |
| original22-v1/threads/atomic-rmw | 17,511.73 | 256,456.48 | 5.917 |
| original22-v1/threads/notify-empty | 17,323.73 | 255,732.31 | 7.042 |
| original22-v1/threads/wait-mismatch | 18,039.15 | 254,856.08 | 7.062 |
| original22-v1/threads/wait-timeout | 17,245.67 | 250,856.21 | 6.292 |
| original22-v1/applications/sha256-clang | 25,186.29 | 250,771.56 | 13.000 |
| original22-v1/applications/sha256-rust | 23,082.44 | 249,831.31 | 10.188 |
| gc-sustained-v1/gc/i31-roundtrip | 17,504.71 | 251,788.90 | 5.042 |
| gc-sustained-v1/gc/struct-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 17,520.81 | 253,341.75 | 5.896 |
| threads-sustained-v2/threads/shared-memory | 17,367.65 | 257,154.94 | 5.021 |
| threads-sustained-v2/threads/atomic-rmw | 16,964.27 | 248,765.52 | 6.312 |
| threads-sustained-v2/threads/notify-empty | 17,412.96 | 250,671.19 | 6.271 |
| threads-sustained-v2/threads/wait-mismatch | 16,855.44 | 256,981.85 | 4.854 |
| threads-sustained-v2/threads/wait-timeout | 17,438.98 | 253,859.90 | 6.292 |
| generic-calls-v1/core/generic-calls | 18,661.21 | 256,815.81 | 6.417 |

### P3 phase observations

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B / cleanup | 61.542 | 581.69 | NM | 271.54 | NM | NM |
| 16 B / compileValidate | 60,931.73 | 96,325.54 | NM | 53,075.58 | NM | NM |
| 16 B / engineSetup | 0.396 | 0.229 | NM | 158.75 | NM | NM |
| 16 B / executorSetup | N/A | N/A | NM | 48.272 | NM | NM |
| 16 B / exportLookup | 0.750 | 1.147 | NM | 9.062 | NM | NM |
| 16 B / firstCall | 9,237.46 | 21,223.67 | NM | 8,040.54 | NM | NM |
| 16 B / instantiate | 496.35 | 775.92 | NM | 213.29 | NM | NM |
| 16 B / load | 63.562 | 73.291 | NM | 58.833 | NM | NM |
| 16 B / processStartup | 6,691.90 | 13,262.46 | NM | 6,945.83 | NM | NM |
| 16 B / providerSetup | 45.374 | 174.94 | NM | 241.58 | NM | NM |
| 16 B / requestPreparation | 0.646 | 0.708 | NM | N/A | NM | NM |
| 65536 B / cleanup | 67.020 | 564.00 | NM | 268.00 | NM | NM |
| 65536 B / compileValidate | 60,639.71 | 98,189.08 | NM | 52,498.10 | NM | NM |
| 65536 B / engineSetup | 0.375 | 0.208 | NM | 150.58 | NM | NM |
| 65536 B / executorSetup | N/A | N/A | NM | 44.916 | NM | NM |
| 65536 B / exportLookup | 0.791 | 1.084 | NM | 9.396 | NM | NM |
| 65536 B / firstCall | 13,194.81 | 24,649.92 | NM | 9,246.81 | NM | NM |
| 65536 B / instantiate | 498.21 | 780.40 | NM | 202.42 | NM | NM |
| 65536 B / load | 57.667 | 68.438 | NM | 54.041 | NM | NM |
| 65536 B / processStartup | 6,257.75 | 12,642.94 | NM | 6,613.25 | NM | NM |
| 65536 B / providerSetup | 47.480 | 181.88 | NM | 231.79 | NM | NM |
| 65536 B / requestPreparation | 8.563 | 12.104 | NM | N/A | NM | NM |

## Native controls (different contracts)

Unchecked C kernels omit runtime safety/lifetime contracts; native Go implements HTTP replies without Component/WASI. These are overhead oracles, not equivalent competitors, and are never subtracted from runtime timings.

| Workload | Unchecked native C, ns/iteration |
| --- | ---: |
| original22-v1/core/integer-loop | 1.547 |
| original22-v1/core/function-calls | 0.775 |
| original22-v1/core/memory-load-store | 0.282 |
| original22-v1/memory/narrow-unaligned | 0.476 |
| original22-v1/memory/dependent-address | 1.425 |
| original22-v1/memory/float64-unaligned | 0.662 |
| original22-v1/runtime/simd-memory-full | 3.486 |
| original22-v1/runtime/simd-memory-widen | 1.029 |
| original22-v1/runtime/simd-memory-lanes | 4.457 |
| original22-v1/applications/sha256-clang | 174.26 |
| original22-v1/applications/sha256-rust | 161.06 |
| generic-calls-v1/core/generic-calls | 2.076 |

| Payload | Native Go CPU, us/request | Native Go requests/s |
| --- | ---: | ---: |
| 16 B | 2.793 | 1,364,166.29 |
| 65536 B | 74.596 | 35,582.59 |

## WasmEdge sustained AOT failures

Full-size WasmEdge0.17.1 O3 AOT preflight freshly reproduces SIGSEGV in sustained struct access, allocation and indirect access. Small original cases pass. Only those exact three cells remain failures in the timing matrix. The original investigation and the observed cumulative stack growth are retained in the [PR #200 report](performance-pr200-complete-comparison-2026-11-17.md). No incomplete or historical timing rows are pooled into this fresh event.

## Generated-code mechanism

An untimed codegen audit of the actual general-call Wasm confirms one expanded call. Affine composition produces `x * 389569705 + 1196435762` modulo 2^32. Before the cache fix, four MOV/MOVK instructions rematerialize the two literals every iteration. The final loop uses entry-prepared registers with `madd w20,w20,w1,w2`; call and backedge polling remain. Full cold entries and ordinary/metered tiers are retained, so this is not unchecked bare code. The audit entry is 1,232 bytes including cold/control code; this is not total module code size. Assembly and original schema are in the protocol archive.

## Regression investigation

The fixed primary event finds source changes of Go sustained indirect GC interpreter **+0.99% time [0.46%,1.53%]** and LLGo general-call interpreter **+0.53% [0.14%,1.12%]**, with resolving controls. A separate predeclared four-block, seven-case, 120-process diagnostic checks CPU, instructions, cycles and CPI with same-binary A/A. It uses the same binaries and production GC; it is not a statistical extension or acceptance rerun.

Those two small source regressions do not reproduce in the independent diagnostic medians: Go indirect GC CPU 74.003→72.811 ns; LLGo general-call interpreter 96.699→96.158 ns. Their instruction counts are essentially unchanged. LLGo base/head runSlotsLoop both have 2,053 static instructions; the retained diff shows jump-table/data relocation offsets rather than added interpreter operations. This supports a layout/CPI sensitivity investigation, not a claim that the primary regression is fixed.

Fixed-base SDK114→SDK3c regressions with resolving primary controls include normal interpreter calls **+3.80%**, SIMD widening **+1.68%**, SIMD lanes **+15.72%**, sustained GC field access **+3.31%**, generic interpreter calls **+3.79%**, and native Rust SHA **+0.72%**. New-SDK source+SDK combined ratios are separately retained; they are not attributed entirely to source changes.

Independent diagnostic observations (CPU ns/iteration, dynamic instructions and CPI) follow. They describe mechanism, without replacing the primary block intervals:

| Case | Role | CPU ns | Instructions/iteration | CPI |
| --- | ---: | ---: | ---: | ---: |
| applications/sha256-rust | base-llgo-native | 236.95 | 4,945.73 | 0.182313 |
| applications/sha256-rust | old-sdk-base-llgo-native | 237.55 | 4,945.21 | 0.182487 |
| core/function-calls | base-llgo-interpreter | 45.034 | 1,243.11 | 0.141108 |
| core/function-calls | old-sdk-base-llgo-interpreter | 43.593 | 1,165.39 | 0.145801 |
| core/generic-calls | base-llgo-interpreter | 96.699 | 2,287.64 | 0.164423 |
| core/generic-calls | head-llgo-interpreter | 96.158 | 2,287.58 | 0.163945 |
| core/generic-calls | old-sdk-base-llgo-interpreter | 92.606 | 2,183.17 | 0.165322 |
| gc/struct-access | base-llgo-interpreter | 38.915 | 988.48 | 0.152280 |
| gc/struct-access | old-sdk-base-llgo-interpreter | 37.369 | 948.75 | 0.153615 |
| gc/struct-indirect-access | base-go-interpreter | 74.003 | 1,824.67 | 0.157996 |
| gc/struct-indirect-access | head-go-interpreter | 72.811 | 1,824.52 | 0.155597 |
| runtime/simd-memory-lanes | base-llgo-interpreter | 101.78 | 2,141.40 | 0.185334 |
| runtime/simd-memory-lanes | old-sdk-base-llgo-interpreter | 87.958 | 2,127.84 | 0.160938 |
| runtime/simd-memory-widen | base-llgo-interpreter | 56.301 | 1,478.24 | 0.148307 |
| runtime/simd-memory-widen | old-sdk-base-llgo-interpreter | 55.383 | 1,467.56 | 0.147257 |

Normal interpreter calls show about **+6.7% dynamic instructions** at the new SDK. SIMD lanes show about **+0.6% instructions**, but CPI rises **0.16094→0.18533 (+15.2%)**, consistent with their reproduced CPU increase. The full SIMD runner assembly changes its private-memory call from `Memory.load` (wrapper branches inlined into runSIMD) to an out-of-line `Memory.Load`. New nil guards, spills, register allocation and layout also change. These are concrete codegen differences, not proof that one accounts for the entire loss. Required nil checks must remain; upstream optimization should preserve semantics.

Native Rust SHA’s small SDK timing regression does not reproduce in the diagnostic CPU medians (237.552→236.948 ns); dynamic guest instructions remain about 4,945/iteration. Its primary adverse interval remains recorded. P3 source CPU intervals span one; SDK instruction counts increase about 0.94–1.17% with resolving controls, without a confirmed SDK CPU regression.

## Next implementation work

1. Investigate native calls whose control flow is dynamic and cannot use bounded expansion; add a separately named workload that survives each competitor’s optimizers before attributing ratios to call ABI overhead.
2. Improve SHA range/region coverage and invariant-reference field access while preserving traps, GC roots and cold resumes.
3. Improve shared-memory loop efficiency; prior poll grouping did not prove a gain and remains rejected.
4. Continue complete P3 canonical/host-transition optimization and the upstream LLGo interpreter regression investigation. Performance acceptance remains open.

## Reproducible evidence

- [Machine-readable results, all controls and source/SDK contrasts](data/bounded-pure-call-2026-11-20.json).
- [All 6,324 primary and 120 independent diagnostic process records](data/bounded-pure-call-2026-11-20-raw.jsonl.gz), SHA-256 `116a26da77bbf1838c8338e9506a136ff50da7fb9897ef235bea408fff132024`.
- [Frozen plans, commands, hashes, analysis/export scripts, semantic checks and failure evidence](data/bounded-pure-call-2026-11-20-protocol.tar.gz), SHA-256 `68010e778d3b4fc4f7a19a4208fa80a5efc497a41e1e998746054cf50a4e5897`.
- Per-request vectors unused by these metrics are projected to count/extrema/hash; no latency-tail claim. Public load records omit unrelated desktop command arguments. The separate diagnostic contains 120 process records and is never pooled into the 6,324 primary records.
