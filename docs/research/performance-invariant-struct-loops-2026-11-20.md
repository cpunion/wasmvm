# Invariant struct loops and runtime comparison, 2026-11-20–21

The native compiler resolves an invariant struct handle at loop entry and reuses the checked payload on the private backedge. Guard failure runs the original field checks at their original guest positions. The private body still polls and reconstructs canonical state on exit. A forwarded i32 field reuses its already narrowed arithmetic result. Precise-fuel and independent continuations preserve their original checks.

- GO sustained field access: 1.375 → 1.113 ns/iteration. Paired head/base: 0.802 [0.796, 0.809]; Δ -19.80%; speed +24.68% [+23.65%, +25.57%]; pass.
  Against Wasmtime (1.087 ns): 1.009 [1.002, 1.016]; Δ +0.92%; speed -0.91% [-1.55%, -0.18%]; unresolved.
- LLGO sustained field access: 1.368 → 1.152 ns/iteration. Paired head/base: 0.817 [0.804, 0.832]; Δ -18.34%; speed +22.45% [+20.14%, +24.45%]; unresolved.
  Against Wasmtime (1.087 ns): 1.024 [1.005, 1.047]; Δ +2.36%; speed -2.30% [-4.48%, -0.52%]; unresolved.

All 33 core workloads and two complete P3 payloads are measured on the same pinned toolchains. Tables give preparation and steady execution independently. The performance acceptance gate remains open.

## Percentage differences

Cells show `absolute observation (signed speed Δ%)`. Positive speed change means faster. Native execution and P3 use **Wasmtime** as a fixed reference; interpreter execution uses **wazero**, with **WasmEdge** explicitly named for GC rows where wazero rejects the feature. These choices make each table comparable across revisions; they do not assert that the reference is the fastest competitor.

For time, speed Δ = `(reference / candidate - 1) * 100%`; for throughput, speed Δ = `(candidate / reference - 1) * 100%`. Thus 100 ns to 50 ns is **−50% time**, **+100% speed**. P3 CPU percentages mean work per CPU time, separate from wall-clock throughput. Instruction/allocation/byte counters show quantity Δ rather than speed. Source comparisons retain paired ratio intervals and report quantity Δ plus speed Δ where applicable; reciprocal interval endpoints are reversed.

Competitive cell percentages are descriptive quotients of the displayed medians; paired block estimates below can differ. `†` on these cells means either the candidate or the named reference has unresolved A/A controls; qualified conclusions require resolved controls. Incomparable preparation/phase boundaries keep separate absolute observations. All percentages derive from this frozen event; they do not measure any later source changes.

## Revisions, profiles and protocol

- Engine source: base `d3af015d` (merged PR #202), measured head `c374cdc2` for PR #203. Reused binaries retain their actual producer revisions; `build.json` records verified production-source equivalence to the qualified source. No full source copies or new worktrees were created.
- Go **1.27.3**; LLGo **3c6b33ec**, **`-O3 -lto=full`**, LLVM **22.1.8**, default production GC. Source base/head share the compiler profile.
- wazero **1.12.0**, native compiler/interpreter; Wasmtime **48.0.1**, Cranelift speed; Wasmer **7.3.0**, Cranelift; WasmEdge **0.17.1**, interpreter and **O3 AOT**. Comparison scope is the named profile of each runtime.
- Local Apple M4 Max, macOS arm64, 128 GiB RAM. External desktop load is uncontrolled and recorded; no owned builds, tests, profiles, analysis or edits overlap timing. The timing scope is macOS arm64.
- Core: six fixed fresh-process blocks, each with symmetric same-binary A/A, two warmups and seven interleaved measured calls. **5,364 successful process records** after excluding exactly three failing sustained WasmEdge AOT cells. All 33 workloads remain represented. Allocation workloads use means including pauses; others use medians.
- P3: **72 processes per payload**, six roles with A/A, six fixed blocks, four workers, sixteen clients and twenty interleaved 512-request windows plus a separate flush. Verify status, headers, full body, trailers, CRC and joined guest completion. CPU/instruction counters include the driver and checks. Mach ticks use the recorded timebase.
- Ratios use paired block geometric means and 20,000 fixed-seed bootstrap resamples with marginal 95% intervals. Both relevant A/A intervals must fit `[0.975, 1.025]`. Each interval applies to its own comparison. The protocol fixes the sample size in advance and measures each phase directly.

**Legend:** `†` means the role or its percentage reference has an unresolved same-binary timing control. Values are observations; a qualified comparison additionally needs both controls and its ratio interval. `US` means this pinned profile rejected the feature in semantic preflight. `NM` means this benchmark adapter/profile was not measured; this status applies only to this measurement. `N/A` means the metric does not apply. `FAIL` is an execution failure in the named profile.

## Native compiled execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

Reference: **Wasmtime**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 3,604.00 (-60.71% speed) † | 2,708.50 (-47.72% speed) † | 583.00 (+142.88% speed) † | 1,416.00 (ref) † | 2,979.00 (-52.47% speed) † | 3,354.50 (-57.79% speed) † |
| original22-v1/core/integer-loop | 1.789 (+2.12% speed) † | 1.876 (-2.66% speed) † | 1.993 (-8.37% speed) † | 1.826 (ref) † | 1.979 (-7.73% speed) † | 1.950 (-6.34% speed) † |
| original22-v1/core/function-calls | 0.840 (+10.14% speed) † | 0.810 (+14.20% speed) † | 1.474 (-37.25% speed) † | 0.925 (ref) † | 0.823 (+12.43% speed) † | 0.804 (+14.98% speed) † |
| original22-v1/core/memory-load-store | 0.476 (+9.07% speed) † | 0.467 (+11.28% speed) † | 0.914 (-43.16% speed) † | 0.519 (ref) † | 0.576 (-9.80% speed) † | 0.310 (+67.44% speed) † |
| original22-v1/memory/narrow-unaligned | 0.668 (+5.69% speed) | 0.659 (+7.14% speed) | 1.072 (-34.17% speed) † | 0.706 (ref) | 0.737 (-4.22% speed) † | 0.528 (+33.79% speed) |
| original22-v1/memory/dependent-address | 1.955 (+8.67% speed) † | 1.971 (+7.80% speed) † | 3.262 (-34.88% speed) † | 2.124 (ref) † | 2.229 (-4.68% speed) † | 1.480 (+43.50% speed) † |
| original22-v1/memory/float64-unaligned | 0.779 (-4.13% speed) † | 0.768 (-2.78% speed) † | 0.900 (-16.98% speed) † | 0.747 (ref) † | 0.753 (-0.82% speed) † | 0.709 (+5.32% speed) † |
| original22-v1/runtime/simd-memory-full | 3.450 (+7.52% speed) † | 3.463 (+7.12% speed) † | 4.009 (-7.45% speed) † | 3.710 (ref) † | 3.554 (+4.39% speed) † | 3.659 (+1.40% speed) † |
| original22-v1/runtime/simd-memory-widen | 1.074 (-1.80% speed) | 1.057 (-0.19% speed) | 1.168 (-9.74% speed) † | 1.055 (ref) | 1.056 (-0.13% speed) | 1.055 (-0.05% speed) |
| original22-v1/runtime/simd-memory-lanes | 5.199 (+17.42% speed) † | 5.191 (+17.60% speed) | 6.295 (-3.02% speed) † | 6.105 (ref) | 6.059 (+0.76% speed) | 5.290 (+15.41% speed) † |
| original22-v1/gc/i31-roundtrip | 2.173 (-0.77% speed) † | 1.927 (+11.89% speed) † | US | 2.156 (ref) † | US | 1.942 (+11.05% speed) † |
| original22-v1/gc/struct-access | 2.548 (-41.46% speed) † | 1.873 (-20.36% speed) † | US | 1.492 (ref) † | US | 10.273 (-85.48% speed) † |
| original22-v1/gc/struct-allocation | 30.798 (+133.44% speed) † | 33.486 (+114.70% speed) † | US | 71.895 (ref) † | US | 40.005 (+79.71% speed) † |
| original22-v1/gc/struct-indirect-access | 3.269 (-4.40% speed) † | 2.546 (+22.75% speed) † | US | 3.125 (ref) | US | 27.788 (-88.75% speed) † |
| original22-v1/threads/private-memory | 1.194 (+19.37% speed) † | 0.831 (+71.43% speed) † | 1.598 (-10.82% speed) † | 1.425 (ref) † | 1.671 (-14.71% speed) † | 1.667 (-14.50% speed) † |
| original22-v1/threads/shared-memory | 2.788 (-45.37% speed) † | 1.910 (-20.29% speed) † | 1.844 (-17.40% speed) † | 1.523 (ref) † | 1.729 (-11.93% speed) † | 1.783 (-14.60% speed) † |
| original22-v1/threads/atomic-rmw | 3.119 (-20.91% speed) † | 2.700 (-8.64% speed) † | 2.504 (-1.50% speed) † | 2.467 (ref) † | 4.981 (-50.48% speed) † | 2.623 (-5.96% speed) † |
| original22-v1/threads/notify-empty | 1.925 (+267.64% speed) † | 1.496 (+373.13% speed) † | 22.117 (-68.00% speed) † | 7.077 (ref) | 5.569 (+27.09% speed) † | 9.506 (-25.55% speed) † |
| original22-v1/threads/wait-mismatch | 2.173 (+324.45% speed) † | 1.650 (+458.96% speed) † | 25.748 (-64.18% speed) † | 9.223 (ref) | 3.038 (+203.63% speed) † | 3.917 (+135.48% speed) † |
| original22-v1/threads/wait-timeout | 2.558 (+1606.91% speed) † | 1.771 (+2365.97% speed) † | 280.51 (-84.43% speed) | 43.669 (ref) | 171.76 (-74.58% speed) † | 61.342 (-28.81% speed) † |
| original22-v1/applications/sha256-clang | 252.88 (-20.79% speed) † | 252.10 (-20.54% speed) † | 329.11 (-39.13% speed) † | 200.31 (ref) † | 230.59 (-13.13% speed) † | 193.80 (+3.36% speed) † |
| original22-v1/applications/sha256-rust | 237.07 (-20.74% speed) † | 235.29 (-20.14% speed) | 326.12 (-42.38% speed) † | 187.90 (ref) | 188.70 (-0.42% speed) | 184.72 (+1.72% speed) |
| gc-sustained-v1/gc/i31-roundtrip | 1.498 (+33.34% speed) † | 1.505 (+32.66% speed) † | US | 1.997 (ref) † | US | 1.425 (+40.15% speed) † |
| gc-sustained-v1/gc/struct-access | 1.113 (-2.40% speed) † | 1.152 (-5.69% speed) † | US | 1.087 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 32.675 (+59.55% speed) † | 36.848 (+41.48% speed) † | US | 52.134 (ref) † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1.734 (+60.10% speed) † | 1.740 (+59.53% speed) † | US | 2.776 (ref) † | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.438 (+150.43% speed) † | 0.349 (+214.16% speed) † | 1.179 (-7.08% speed) † | 1.096 (ref) † | 1.110 (-1.31% speed) † | 1.120 (-2.16% speed) † |
| threads-sustained-v2/threads/shared-memory | 1.450 (-25.53% speed) † | 1.294 (-16.51% speed) † | 1.264 (-14.53% speed) † | 1.080 (ref) † | 1.118 (-3.37% speed) † | 1.111 (-2.76% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 2.076 (-6.78% speed) † | 1.993 (-2.90% speed) † | 2.151 (-10.03% speed) † | 1.935 (ref) | 4.125 (-53.08% speed) | 1.973 (-1.90% speed) † |
| threads-sustained-v2/threads/notify-empty | 1.134 (+501.34% speed) † | 0.986 (+591.51% speed) † | 21.150 (-67.75% speed) † | 6.821 (ref) † | 4.770 (+43.00% speed) † | 8.990 (-24.13% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 1.283 (+607.15% speed) † | 1.139 (+696.40% speed) † | 24.672 (-63.23% speed) † | 9.071 (ref) † | 2.500 (+262.79% speed) † | 3.301 (+174.75% speed) † |
| threads-sustained-v2/threads/wait-timeout | 1.138 (+3754.93% speed) † | 0.993 (+4317.88% speed) † | 291.62 (-84.96% speed) † | 43.866 (ref) † | 172.48 (-74.57% speed) † | 61.578 (-28.76% speed) † |
| generic-calls-v1/core/generic-calls | 1.045 (+98.65% speed) † | 1.037 (+100.30% speed) † | 2.115 (-1.80% speed) † | 2.077 (ref) † | 2.071 (+0.27% speed) † | 1.037 (+100.29% speed) † |

## Interpreter execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

Reference: **wazero**; GC rows explicitly use **WasmEdge**.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2,333.00 (-57.12% speed) † | 2,333.50 (-57.12% speed) † | 1,000.50 (ref) † | NM | NM | 2,875.00 (-65.20% speed) † |
| original22-v1/core/integer-loop | 41.577 (+166.48% speed) † | 31.884 (+247.48% speed) † | 110.79 (ref) † | NM | NM | 66.357 (+66.97% speed) † |
| original22-v1/core/function-calls | 55.843 (+67.84% speed) † | 46.857 (+100.03% speed) † | 93.728 (ref) † | NM | NM | 64.060 (+46.31% speed) † |
| original22-v1/core/memory-load-store | 77.224 (+104.06% speed) † | 63.070 (+149.86% speed) † | 157.59 (ref) † | NM | NM | 100.95 (+56.11% speed) † |
| original22-v1/memory/narrow-unaligned | 95.714 (+98.51% speed) † | 77.349 (+145.64% speed) † | 190.00 (ref) † | NM | NM | 118.05 (+60.95% speed) † |
| original22-v1/memory/dependent-address | 48.265 (+110.12% speed) † | 37.439 (+170.88% speed) † | 101.41 (ref) † | NM | NM | 56.104 (+80.76% speed) † |
| original22-v1/memory/float64-unaligned | 78.423 (+80.82% speed) † | 63.348 (+123.85% speed) † | 141.81 (ref) † | NM | NM | 87.833 (+61.45% speed) † |
| original22-v1/runtime/simd-memory-full | 82.778 (+86.94% speed) † | 67.746 (+128.42% speed) † | 154.74 (ref) † | NM | NM | 79.529 (+94.57% speed) † |
| original22-v1/runtime/simd-memory-widen | 67.998 (+114.00% speed) | 57.278 (+154.05% speed) | 145.51 (ref) | NM | NM | 62.209 (+133.91% speed) |
| original22-v1/runtime/simd-memory-lanes | 104.64 (+87.34% speed) † | 91.159 (+115.04% speed) † | 196.03 (ref) † | NM | NM | 95.215 (+105.88% speed) † |
| original22-v1/gc/i31-roundtrip (vs WasmEdge) | 43.546 (+9.31% speed) † | 32.981 (+44.32% speed) † | US | NM | NM | 47.600 (ref) † |
| original22-v1/gc/struct-access (vs WasmEdge) | 47.931 (+18.03% speed) † | 39.771 (+42.25% speed) † | US | NM | NM | 56.575 (ref) † |
| original22-v1/gc/struct-allocation (vs WasmEdge) | 99.927 (+9.61% speed) † | 104.24 (+5.08% speed) † | US | NM | NM | 109.53 (ref) † |
| original22-v1/gc/struct-indirect-access (vs WasmEdge) | 85.323 (+15.98% speed) † | 74.958 (+32.02% speed) † | US | NM | NM | 98.958 (ref) † |
| original22-v1/threads/private-memory | 43.156 (+49.48% speed) † | 34.444 (+87.29% speed) † | 64.508 (ref) † | NM | NM | 39.840 (+61.92% speed) † |
| original22-v1/threads/shared-memory | 80.888 (-17.95% speed) † | 36.215 (+83.27% speed) † | 66.371 (ref) † | NM | NM | 40.540 (+63.72% speed) † |
| original22-v1/threads/atomic-rmw | 57.948 (-12.61% speed) † | 36.925 (+37.14% speed) † | 50.640 (ref) † | NM | NM | 34.633 (+46.22% speed) † |
| original22-v1/threads/notify-empty | 34.852 (+25.15% speed) † | 28.271 (+54.29% speed) † | 43.619 (ref) † | NM | NM | 40.704 (+7.16% speed) † |
| original22-v1/threads/wait-mismatch | 61.569 (+15.23% speed) † | 37.600 (+88.69% speed) † | 70.948 (ref) † | NM | NM | 44.185 (+60.57% speed) † |
| original22-v1/threads/wait-timeout | 59.921 (+451.42% speed) † | 36.148 (+814.06% speed) † | 330.41 (ref) † | NM | NM | 97.827 (+237.75% speed) † |
| original22-v1/applications/sha256-clang | 21,968.84 (+121.82% speed) † | 18,039.90 (+170.13% speed) † | 48,732.00 (ref) † | NM | NM | 30,317.46 (+60.74% speed) † |
| original22-v1/applications/sha256-rust | 19,427.32 (+112.42% speed) | 15,768.50 (+161.71% speed) | 41,267.19 (ref) | NM | NM | 24,480.33 (+68.57% speed) |
| gc-sustained-v1/gc/i31-roundtrip (vs WasmEdge) | 43.808 (+11.94% speed) † | 35.266 (+39.05% speed) † | US | NM | NM | 49.039 (ref) † |
| gc-sustained-v1/gc/struct-access (vs WasmEdge) | 48.259 (+22.85% speed) † | 40.698 (+45.67% speed) † | US | NM | NM | 59.286 (ref) † |
| gc-sustained-v1/gc/struct-allocation (vs WasmEdge) | 110.14 (-1.05% speed) † | 110.10 (-1.01% speed) † | US | NM | NM | 108.99 (ref) † |
| gc-sustained-v1/gc/struct-indirect-access (vs WasmEdge) | 86.652 (+19.23% speed) † | 74.743 (+38.23% speed) † | US | NM | NM | 103.32 (ref) † |
| threads-sustained-v2/threads/private-memory | 42.394 (+49.89% speed) † | 34.283 (+85.35% speed) † | 63.544 (ref) † | NM | NM | 39.552 (+60.66% speed) † |
| threads-sustained-v2/threads/shared-memory | 80.079 (-20.91% speed) † | 35.178 (+80.03% speed) † | 63.331 (ref) † | NM | NM | 39.486 (+60.39% speed) † |
| threads-sustained-v2/threads/atomic-rmw | 55.677 (-9.95% speed) † | 34.676 (+44.59% speed) † | 50.136 (ref) † | NM | NM | 33.748 (+48.56% speed) † |
| threads-sustained-v2/threads/notify-empty | 32.979 (+27.99% speed) † | 27.770 (+52.00% speed) † | 42.210 (ref) † | NM | NM | 40.058 (+5.37% speed) † |
| threads-sustained-v2/threads/wait-mismatch | 61.284 (+14.48% speed) † | 37.543 (+86.87% speed) † | 70.156 (ref) | NM | NM | 44.101 (+59.08% speed) † |
| threads-sustained-v2/threads/wait-timeout | 58.761 (+473.93% speed) † | 35.017 (+863.09% speed) † | 337.25 (ref) † | NM | NM | 96.927 (+247.94% speed) † |
| generic-calls-v1/core/generic-calls | 116.90 (+56.59% speed) † | 93.977 (+94.80% speed) † | 183.06 (ref) † | NM | NM | 117.38 (+55.96% speed) † |

## Complete P3 HTTP requests

P3 competitor cells marked `NM` have no measured adapter for the pinned component/WASI workload. Wasmtime uses the same validated workload. Managed allocation counts exclude foreign Rust/C/JIT allocations and cannot compare total allocation across engines.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU, us/request | 102.62 (-53.04% speed) † | 259.00 (-81.39% speed) † | NM | 48.190 (ref) † | NM | NM |
| 16 B instructions/request | 706,127.93 (+219.59% Δ) † | 1,290,294.88 (+483.99% Δ) † | NM | 220,946.43 (ref) † | NM | NM |
| 16 B requests/s | 41,796.97 (-22.18% speed) † | 19,276.71 (-64.11% speed) † | NM | 53,712.18 (ref) † | NM | NM |
| 65536 B CPU, us/request | 170.13 (-52.57% speed) † | 359.18 (-77.53% speed) † | NM | 80.698 (ref) † | NM | NM |
| 65536 B instructions/request | 985,636.03 (+212.83% Δ) | 1,627,675.03 (+416.61% Δ) | NM | 315,069.96 (ref) | NM | NM |
| 65536 B requests/s | 22,229.25 (-38.38% speed) † | 13,254.34 (-63.26% speed) † | NM | 36,074.64 (ref) † | NM | NM |

### P3 source contrasts

| Metric | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| 16 B cpuUs | 0.997 [0.986, 1.008]; Δ -0.29%; speed +0.29% [-0.80%, +1.39%]; pass | 0.996 [0.990, 1.001]; Δ -0.43%; speed +0.43% [-0.05%, +0.99%]; pass |
| 16 B instructions | 1.001 [0.998, 1.003]; Δ +0.05%; pass | 0.999 [0.994, 1.003]; Δ -0.13%; pass |
| 16 B throughput | 1.008 [0.989, 1.029]; Δ +0.79%; speed +0.79% [-1.14%, +2.92%]; unresolved | 0.991 [0.975, 1.002]; Δ -0.91%; speed -0.91% [-2.50%, +0.23%]; unresolved |
| 16 B allocsPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 16 B bytesPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 65536 B cpuUs | 1.000 [0.992, 1.007]; Δ -0.02%; speed +0.02% [-0.68%, +0.82%]; unresolved | 0.990 [0.973, 1.001]; Δ -1.04%; speed +1.05% [-0.14%, +2.76%]; unresolved |
| 65536 B instructions | 1.002 [0.999, 1.006]; Δ +0.24%; pass | 0.996 [0.987, 1.004]; Δ -0.36%; pass |
| 65536 B throughput | 0.994 [0.981, 1.005]; Δ -0.56%; speed -0.56% [-1.92%, +0.51%]; unresolved | 0.997 [0.961, 1.030]; Δ -0.28%; speed -0.28% [-3.90%, +2.96%]; unresolved |
| 65536 B allocsPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| 65536 B bytesPerRequest | 1.000 [1.000, 1.000]; Δ +0.00%; pass | 1.000 [1.000, 1.000]; Δ +0.00%; pass |

Throughput ratios above one improve; CPU/instructions/allocation ratios below one improve. The JSON retains all own-baseline observations and every per-role A/A interval.

### Competitive P3 ratios

| Payload / compiler | CPU ratio to Wasmtime [95% CI]; controls | Throughput ratio to Wasmtime [95% CI]; controls |
| --- | ---: | ---: |
| 16 B go | 2.086 [1.924, 2.261]; Δ +108.61%; speed -52.06% [-55.78%, -48.04%]; unresolved | 0.751 [0.730, 0.774]; Δ -24.93%; speed -24.93% [-27.03%, -22.63%]; unresolved |
| 16 B llgo | 5.337 [4.834, 5.897]; Δ +433.72%; speed -81.26% [-83.04%, -79.31%]; unresolved | 0.352 [0.336, 0.365]; Δ -64.84%; speed -64.84% [-66.38%, -63.51%]; unresolved |
| 65536 B go | 2.215 [2.130, 2.357]; Δ +121.47%; speed -54.85% [-57.58%, -53.05%]; unresolved | 0.639 [0.619, 0.664]; Δ -36.15%; speed -36.15% [-38.13%, -33.62%]; unresolved |
| 65536 B llgo | 4.443 [3.904, 5.116]; Δ +344.31%; speed -77.49% [-80.45%, -74.39%]; unresolved | 0.356 [0.343, 0.372]; Δ -64.37%; speed -64.37% [-65.73%, -62.83%]; unresolved |

## Core source baseline

Head/base ratios below one improve. These marginal intervals describe this event; an apparent change in an unchanged mechanism requires diagnosis before being attributed to source changes. The machine-readable report includes every competitor ratio and A/A interval.

### native

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 1.001 [0.904, 1.124]; Δ +0.10%; speed -0.10% [-11.02%, +10.62%]; unresolved | 0.912 [0.801, 1.045]; Δ -8.82%; speed +9.67% [-4.27%, +24.84%]; unresolved |
| original22-v1/core/integer-loop | 0.971 [0.919, 1.010]; Δ -2.88%; speed +2.96% [-0.98%, +8.84%]; unresolved | 1.021 [0.991, 1.053]; Δ +2.09%; speed -2.04% [-5.02%, +0.89%]; unresolved |
| original22-v1/core/function-calls | 1.032 [1.003, 1.070]; Δ +3.17%; speed -3.07% [-6.52%, -0.30%]; unresolved | 0.975 [0.934, 1.002]; Δ -2.47%; speed +2.53% [-0.16%, +7.12%]; unresolved |
| original22-v1/core/memory-load-store | 0.998 [0.986, 1.010]; Δ -0.21%; speed +0.21% [-1.01%, +1.44%]; unresolved | 0.977 [0.936, 1.011]; Δ -2.32%; speed +2.37% [-1.08%, +6.81%]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.998 [0.986, 1.007]; Δ -0.20%; speed +0.20% [-0.66%, +1.40%]; pass | 1.006 [1.001, 1.010]; Δ +0.58%; speed -0.57% [-1.00%, -0.11%]; pass |
| original22-v1/memory/dependent-address | 0.996 [0.990, 1.003]; Δ -0.38%; speed +0.38% [-0.28%, +0.96%]; unresolved | 1.013 [1.000, 1.033]; Δ +1.29%; speed -1.27% [-3.22%, +0.00%]; unresolved |
| original22-v1/memory/float64-unaligned | 0.970 [0.936, 0.995]; Δ -2.97%; speed +3.07% [+0.52%, +6.85%]; unresolved | 1.014 [0.999, 1.036]; Δ +1.40%; speed -1.38% [-3.47%, +0.11%]; unresolved |
| original22-v1/runtime/simd-memory-full | 0.999 [0.987, 1.014]; Δ -0.11%; speed +0.11% [-1.42%, +1.33%]; unresolved | 0.983 [0.949, 1.005]; Δ -1.71%; speed +1.74% [-0.52%, +5.34%]; unresolved |
| original22-v1/runtime/simd-memory-widen | 0.996 [0.986, 1.004]; Δ -0.43%; speed +0.43% [-0.35%, +1.37%]; unresolved | 1.003 [0.998, 1.009]; Δ +0.30%; speed -0.30% [-0.86%, +0.23%]; pass |
| original22-v1/runtime/simd-memory-lanes | 0.980 [0.938, 1.011]; Δ -2.01%; speed +2.05% [-1.13%, +6.61%]; unresolved | 0.992 [0.985, 0.999]; Δ -0.84%; speed +0.85% [+0.12%, +1.47%]; unresolved |
| original22-v1/gc/i31-roundtrip | 1.028 [0.978, 1.084]; Δ +2.82%; speed -2.74% [-7.77%, +2.24%]; unresolved | 0.996 [0.974, 1.019]; Δ -0.37%; speed +0.37% [-1.85%, +2.64%]; unresolved |
| original22-v1/gc/struct-access | 0.950 [0.847, 1.066]; Δ -5.04%; speed +5.30% [-6.21%, +18.04%]; unresolved | 0.870 [0.837, 0.908]; Δ -13.03%; speed +14.98% [+10.12%, +19.49%]; unresolved |
| original22-v1/gc/struct-allocation | 0.967 [0.898, 1.010]; Δ -3.31%; speed +3.42% [-1.01%, +11.31%]; unresolved | 0.993 [0.957, 1.027]; Δ -0.74%; speed +0.74% [-2.64%, +4.45%]; unresolved |
| original22-v1/gc/struct-indirect-access | 0.927 [0.839, 1.014]; Δ -7.33%; speed +7.91% [-1.41%, +19.14%]; unresolved | 0.965 [0.847, 1.058]; Δ -3.46%; speed +3.58% [-5.48%, +18.07%]; unresolved |
| original22-v1/threads/private-memory | 1.004 [0.927, 1.100]; Δ +0.42%; speed -0.41% [-9.10%, +7.83%]; unresolved | 0.927 [0.865, 0.992]; Δ -7.34%; speed +7.92% [+0.76%, +15.59%]; unresolved |
| original22-v1/threads/shared-memory | 1.107 [1.019, 1.226]; Δ +10.71%; speed -9.67% [-18.42%, -1.87%]; unresolved | 1.006 [0.963, 1.056]; Δ +0.61%; speed -0.61% [-5.31%, +3.81%]; unresolved |
| original22-v1/threads/atomic-rmw | 1.052 [0.970, 1.203]; Δ +5.22%; speed -4.96% [-16.90%, +3.11%]; unresolved | 1.037 [0.987, 1.083]; Δ +3.68%; speed -3.54% [-7.64%, +1.34%]; unresolved |
| original22-v1/threads/notify-empty | 1.011 [0.889, 1.141]; Δ +1.10%; speed -1.09% [-12.32%, +12.46%]; unresolved | 0.998 [0.975, 1.027]; Δ -0.18%; speed +0.18% [-2.59%, +2.53%]; unresolved |
| original22-v1/threads/wait-mismatch | 0.990 [0.858, 1.163]; Δ -0.96%; speed +0.97% [-14.00%, +16.56%]; unresolved | 0.965 [0.914, 1.010]; Δ -3.47%; speed +3.59% [-1.04%, +9.36%]; unresolved |
| original22-v1/threads/wait-timeout | 0.981 [0.923, 1.046]; Δ -1.87%; speed +1.90% [-4.36%, +8.31%]; unresolved | 1.018 [0.993, 1.046]; Δ +1.78%; speed -1.74% [-4.35%, +0.67%]; unresolved |
| original22-v1/applications/sha256-clang | 0.990 [0.965, 1.015]; Δ -1.03%; speed +1.05% [-1.51%, +3.60%]; unresolved | 0.959 [0.906, 1.010]; Δ -4.13%; speed +4.31% [-1.02%, +10.34%]; unresolved |
| original22-v1/applications/sha256-rust | 1.010 [0.996, 1.029]; Δ +1.00%; speed -0.99% [-2.78%, +0.38%]; unresolved | 1.002 [0.996, 1.009]; Δ +0.19%; speed -0.19% [-0.91%, +0.40%]; pass |
| gc-sustained-v1/gc/i31-roundtrip | 1.020 [0.988, 1.074]; Δ +1.97%; speed -1.93% [-6.91%, +1.23%]; unresolved | 1.021 [0.988, 1.069]; Δ +2.11%; speed -2.07% [-6.46%, +1.16%]; unresolved |
| gc-sustained-v1/gc/struct-access | 0.802 [0.796, 0.809]; Δ -19.80%; speed +24.68% [+23.65%, +25.57%]; pass | 0.817 [0.804, 0.832]; Δ -18.34%; speed +22.45% [+20.14%, +24.45%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 0.989 [0.953, 1.021]; Δ -1.09%; speed +1.10% [-2.10%, +4.93%]; unresolved | 1.012 [0.991, 1.036]; Δ +1.21%; speed -1.20% [-3.45%, +0.86%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.022 [0.970, 1.094]; Δ +2.16%; speed -2.11% [-8.58%, +3.11%]; unresolved | 1.005 [0.980, 1.022]; Δ +0.47%; speed -0.47% [-2.11%, +2.03%]; unresolved |
| threads-sustained-v2/threads/private-memory | 1.041 [1.000, 1.094]; Δ +4.07%; speed -3.91% [-8.63%, +0.03%]; unresolved | 1.004 [0.994, 1.013]; Δ +0.39%; speed -0.39% [-1.29%, +0.60%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 0.965 [0.915, 0.999]; Δ -3.51%; speed +3.64% [+0.13%, +9.33%]; unresolved | 0.992 [0.839, 1.200]; Δ -0.82%; speed +0.83% [-16.68%, +19.21%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.990 [0.981, 1.002]; Δ -1.05%; speed +1.06% [-0.20%, +1.89%]; unresolved | 0.988 [0.953, 1.014]; Δ -1.19%; speed +1.21% [-1.38%, +4.93%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 0.955 [0.815, 1.053]; Δ -4.53%; speed +4.75% [-5.08%, +22.70%]; unresolved | 1.099 [0.960, 1.385]; Δ +9.88%; speed -8.99% [-27.80%, +4.16%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.005 [0.970, 1.037]; Δ +0.54%; speed -0.54% [-3.58%, +3.14%]; unresolved | 1.026 [0.984, 1.079]; Δ +2.63%; speed -2.56% [-7.33%, +1.59%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 1.076 [1.000, 1.200]; Δ +7.60%; speed -7.06% [-16.68%, +0.01%]; unresolved | 1.042 [0.988, 1.109]; Δ +4.20%; speed -4.03% [-9.84%, +1.23%]; unresolved |
| generic-calls-v1/core/generic-calls | 1.003 [0.994, 1.016]; Δ +0.34%; speed -0.34% [-1.56%, +0.62%]; unresolved | 0.998 [0.984, 1.011]; Δ -0.16%; speed +0.16% [-1.12%, +1.64%]; pass |

### interpreter

| Workload | Go source head/base | LLGo source head/base |
| --- | ---: | ---: |
| original22-v1/core/call-control | 1.029 [0.791, 1.239]; Δ +2.87%; speed -2.79% [-19.32%, +26.38%]; unresolved | 1.018 [0.829, 1.248]; Δ +1.80%; speed -1.76% [-19.85%, +20.60%]; unresolved |
| original22-v1/core/integer-loop | 1.107 [1.042, 1.146]; Δ +10.70%; speed -9.66% [-12.78%, -4.03%]; unresolved | 1.008 [0.967, 1.048]; Δ +0.85%; speed -0.84% [-4.59%, +3.45%]; unresolved |
| original22-v1/core/function-calls | 1.095 [1.071, 1.115]; Δ +9.52%; speed -8.69% [-10.33%, -6.64%]; unresolved | 0.953 [0.923, 1.002]; Δ -4.70%; speed +4.93% [-0.17%, +8.39%]; unresolved |
| original22-v1/core/memory-load-store | 1.125 [1.101, 1.163]; Δ +12.48%; speed -11.09% [-14.04%, -9.18%]; unresolved | 0.998 [0.982, 1.011]; Δ -0.17%; speed +0.18% [-1.13%, +1.79%]; unresolved |
| original22-v1/memory/narrow-unaligned | 1.105 [1.090, 1.121]; Δ +10.54%; speed -9.53% [-10.76%, -8.27%]; pass | 1.001 [0.996, 1.005]; Δ +0.08%; speed -0.08% [-0.52%, +0.39%]; unresolved |
| original22-v1/memory/dependent-address | 1.111 [1.104, 1.119]; Δ +11.10%; speed -9.99% [-10.63%, -9.42%]; pass | 0.992 [0.988, 0.995]; Δ -0.84%; speed +0.84% [+0.49%, +1.25%]; unresolved |
| original22-v1/memory/float64-unaligned | 1.106 [1.079, 1.150]; Δ +10.63%; speed -9.61% [-13.07%, -7.34%]; unresolved | 1.008 [0.996, 1.025]; Δ +0.77%; speed -0.77% [-2.42%, +0.44%]; unresolved |
| original22-v1/runtime/simd-memory-full | 1.086 [1.049, 1.153]; Δ +8.56%; speed -7.89% [-13.28%, -4.71%]; unresolved | 1.004 [0.989, 1.021]; Δ +0.35%; speed -0.35% [-2.08%, +1.15%]; unresolved |
| original22-v1/runtime/simd-memory-widen | 1.072 [1.065, 1.081]; Δ +7.16%; speed -6.68% [-7.46%, -6.13%]; unresolved | 1.007 [1.002, 1.013]; Δ +0.73%; speed -0.72% [-1.25%, -0.22%]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 1.059 [1.053, 1.067]; Δ +5.95%; speed -5.61% [-6.31%, -5.02%]; unresolved | 0.839 [0.814, 0.855]; Δ -16.14%; speed +19.25% [+16.92%, +22.86%]; unresolved |
| original22-v1/gc/i31-roundtrip | 1.132 [1.114, 1.151]; Δ +13.22%; speed -11.67% [-13.12%, -10.27%]; unresolved | 1.000 [0.980, 1.024]; Δ +0.03%; speed -0.03% [-2.34%, +2.04%]; unresolved |
| original22-v1/gc/struct-access | 1.120 [1.111, 1.129]; Δ +12.04%; speed -10.75% [-11.45%, -10.01%]; unresolved | 0.967 [0.936, 0.987]; Δ -3.26%; speed +3.37% [+1.27%, +6.80%]; unresolved |
| original22-v1/gc/struct-allocation | 1.045 [1.016, 1.067]; Δ +4.51%; speed -4.32% [-6.24%, -1.57%]; unresolved | 1.006 [0.976, 1.035]; Δ +0.60%; speed -0.60% [-3.41%, +2.46%]; unresolved |
| original22-v1/gc/struct-indirect-access | 1.101 [1.074, 1.126]; Δ +10.11%; speed -9.18% [-11.21%, -6.86%]; unresolved | 1.001 [0.974, 1.024]; Δ +0.08%; speed -0.08% [-2.32%, +2.71%]; unresolved |
| original22-v1/threads/private-memory | 1.085 [1.070, 1.110]; Δ +8.55%; speed -7.87% [-9.95%, -6.56%]; unresolved | 0.999 [0.977, 1.028]; Δ -0.06%; speed +0.06% [-2.68%, +2.35%]; unresolved |
| original22-v1/threads/shared-memory | 1.034 [1.006, 1.058]; Δ +3.41%; speed -3.30% [-5.44%, -0.64%]; unresolved | 1.044 [0.991, 1.137]; Δ +4.35%; speed -4.17% [-12.08%, +0.91%]; unresolved |
| original22-v1/threads/atomic-rmw | 1.008 [0.983, 1.031]; Δ +0.84%; speed -0.84% [-3.01%, +1.72%]; unresolved | 0.947 [0.875, 1.013]; Δ -5.33%; speed +5.63% [-1.24%, +14.23%]; unresolved |
| original22-v1/threads/notify-empty | 1.038 [0.972, 1.083]; Δ +3.77%; speed -3.63% [-7.64%, +2.84%]; unresolved | 0.968 [0.930, 0.994]; Δ -3.21%; speed +3.32% [+0.65%, +7.53%]; unresolved |
| original22-v1/threads/wait-mismatch | 1.074 [1.057, 1.091]; Δ +7.41%; speed -6.90% [-8.34%, -5.36%]; unresolved | 1.003 [0.997, 1.013]; Δ +0.33%; speed -0.33% [-1.25%, +0.29%]; unresolved |
| original22-v1/threads/wait-timeout | 1.059 [1.052, 1.066]; Δ +5.89%; speed -5.56% [-6.15%, -4.93%]; unresolved | 1.000 [0.968, 1.031]; Δ +0.03%; speed -0.03% [-3.05%, +3.28%]; unresolved |
| original22-v1/applications/sha256-clang | 1.092 [1.072, 1.119]; Δ +9.16%; speed -8.39% [-10.63%, -6.74%]; unresolved | 1.005 [0.985, 1.029]; Δ +0.47%; speed -0.47% [-2.81%, +1.51%]; unresolved |
| original22-v1/applications/sha256-rust | 1.076 [1.070, 1.080]; Δ +7.57%; speed -7.03% [-7.37%, -6.59%]; pass | 1.001 [0.994, 1.008]; Δ +0.13%; speed -0.13% [-0.83%, +0.59%]; pass |
| gc-sustained-v1/gc/i31-roundtrip | 1.090 [1.067, 1.114]; Δ +9.04%; speed -8.29% [-10.22%, -6.29%]; unresolved | 0.975 [0.952, 1.001]; Δ -2.52%; speed +2.59% [-0.05%, +5.09%]; unresolved |
| gc-sustained-v1/gc/struct-access | 1.070 [1.001, 1.112]; Δ +7.03%; speed -6.57% [-10.07%, -0.14%]; unresolved | 1.016 [1.006, 1.028]; Δ +1.57%; speed -1.55% [-2.69%, -0.57%]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 1.030 [0.974, 1.080]; Δ +3.00%; speed -2.91% [-7.43%, +2.68%]; unresolved | 0.983 [0.963, 1.002]; Δ -1.73%; speed +1.76% [-0.21%, +3.84%]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.100 [1.088, 1.121]; Δ +10.03%; speed -9.12% [-10.75%, -8.12%]; unresolved | 0.980 [0.963, 0.993]; Δ -1.95%; speed +1.99% [+0.69%, +3.80%]; unresolved |
| threads-sustained-v2/threads/private-memory | 1.092 [1.074, 1.116]; Δ +9.16%; speed -8.39% [-10.42%, -6.87%]; unresolved | 1.012 [0.991, 1.035]; Δ +1.24%; speed -1.22% [-3.34%, +0.92%]; unresolved |
| threads-sustained-v2/threads/shared-memory | 1.060 [1.045, 1.080]; Δ +6.00%; speed -5.66% [-7.41%, -4.27%]; pass | 1.025 [0.994, 1.076]; Δ +2.47%; speed -2.41% [-7.04%, +0.59%]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 1.029 [1.007, 1.044]; Δ +2.88%; speed -2.80% [-4.18%, -0.72%]; unresolved | 1.021 [1.002, 1.045]; Δ +2.15%; speed -2.10% [-4.29%, -0.17%]; unresolved |
| threads-sustained-v2/threads/notify-empty | 1.036 [0.985, 1.069]; Δ +3.57%; speed -3.44% [-6.46%, +1.56%]; unresolved | 1.005 [0.990, 1.019]; Δ +0.54%; speed -0.53% [-1.87%, +0.96%]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.084 [1.068, 1.106]; Δ +8.43%; speed -7.78% [-9.57%, -6.40%]; unresolved | 0.984 [0.950, 1.007]; Δ -1.64%; speed +1.66% [-0.73%, +5.23%]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 1.039 [1.012, 1.057]; Δ +3.88%; speed -3.73% [-5.44%, -1.21%]; unresolved | 1.022 [0.999, 1.048]; Δ +2.21%; speed -2.16% [-4.55%, +0.07%]; unresolved |
| generic-calls-v1/core/generic-calls | 1.105 [1.099, 1.111]; Δ +10.49%; speed -9.50% [-9.97%, -9.03%]; unresolved | 0.961 [0.940, 0.976]; Δ -3.87%; speed +4.03% [+2.49%, +6.44%]; unresolved |

## Preparation, first call and cleanup

All entries below are **microseconds per process or invocation**, including unresolved observations. Phase intervals are not inferred from warm-call controls. Startup is independently observed process readiness. Native compile/validate boundaries differ by API; WasmEdge AOT includes validation and artifact I/O, plus separately recorded native-image loading. Every process creates a fresh native image. First-call and warm execution have their own boundaries; values are never subtracted to invent pure costs.

<details>
<summary>Native: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,242.38 | 6,042.25 | 4,296.54 | 4,377.06 | 5,868.79 | 16,556.77 |
| original22-v1/core/integer-loop | 5,054.12 | 6,890.50 | 6,820.79 | 6,356.58 | 7,612.73 | 20,025.21 |
| original22-v1/core/function-calls | 4,440.35 | 6,187.12 | 4,511.50 | 4,443.15 | 6,399.88 | 17,015.08 |
| original22-v1/core/memory-load-store | 4,255.96 | 6,611.54 | 4,277.88 | 4,600.85 | 6,278.62 | 17,218.06 |
| original22-v1/memory/narrow-unaligned | 4,395.12 | 6,117.27 | 4,564.50 | 4,411.83 | 6,220.40 | 16,535.54 |
| original22-v1/memory/dependent-address | 4,473.75 | 6,843.71 | 4,639.21 | 4,629.25 | 6,563.17 | 16,698.02 |
| original22-v1/memory/float64-unaligned | 4,451.52 | 6,210.46 | 4,609.48 | 4,567.08 | 6,199.06 | 17,068.31 |
| original22-v1/runtime/simd-memory-full | 4,861.29 | 6,436.90 | 4,811.29 | 4,949.17 | 7,143.83 | 19,474.17 |
| original22-v1/runtime/simd-memory-widen | 4,593.02 | 6,152.79 | 4,562.62 | 4,372.60 | 6,117.35 | 16,579.73 |
| original22-v1/runtime/simd-memory-lanes | 4,631.77 | 6,448.44 | 4,665.35 | 4,982.81 | 6,977.90 | 23,044.71 |
| original22-v1/gc/i31-roundtrip | 4,530.62 | 6,031.44 | US | 4,243.21 | US | 17,076.96 |
| original22-v1/gc/struct-access | 3,958.19 | 6,194.67 | US | 4,568.25 | US | 16,771.10 |
| original22-v1/gc/struct-allocation | 4,751.31 | 6,446.54 | US | 4,583.42 | US | 16,743.15 |
| original22-v1/gc/struct-indirect-access | 4,509.15 | 6,266.23 | US | 4,442.06 | US | 16,753.40 |
| original22-v1/threads/private-memory | 3,912.52 | 5,859.31 | 4,000.12 | 4,195.44 | 5,778.42 | 16,023.04 |
| original22-v1/threads/shared-memory | 4,222.06 | 5,981.73 | 4,373.37 | 4,425.06 | 6,038.79 | 17,067.17 |
| original22-v1/threads/atomic-rmw | 4,638.65 | 6,289.19 | 4,577.48 | 4,077.79 | 6,402.75 | 17,098.04 |
| original22-v1/threads/notify-empty | 4,459.44 | 6,330.17 | 4,436.04 | 4,378.02 | 6,043.10 | 16,418.50 |
| original22-v1/threads/wait-mismatch | 4,265.88 | 6,056.12 | 4,538.12 | 4,291.25 | 6,135.48 | 16,554.56 |
| original22-v1/threads/wait-timeout | 4,686.00 | 5,927.15 | 4,308.96 | 4,452.79 | 6,102.42 | 16,633.65 |
| original22-v1/applications/sha256-clang | 4,948.79 | 8,839.00 | 4,977.23 | 5,079.77 | 8,177.88 | 20,434.92 |
| original22-v1/applications/sha256-rust | 4,306.04 | 5,979.58 | 4,616.60 | 4,325.79 | 6,346.19 | 16,538.50 |
| gc-sustained-v1/gc/i31-roundtrip | 5,783.77 | 8,121.54 | US | 5,296.40 | US | 22,810.85 |
| gc-sustained-v1/gc/struct-access | 4,598.83 | 6,922.25 | US | 4,781.38 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 4,449.75 | 6,431.56 | US | 4,465.81 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 5,300.00 | 6,698.15 | US | 4,480.81 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 4,577.00 | 6,246.23 | 4,672.67 | 4,531.23 | 6,219.21 | 17,901.77 |
| threads-sustained-v2/threads/shared-memory | 4,235.79 | 6,062.12 | 4,246.62 | 4,331.25 | 5,908.96 | 16,448.15 |
| threads-sustained-v2/threads/atomic-rmw | 4,413.79 | 6,353.23 | 4,896.19 | 4,673.56 | 6,160.75 | 17,803.38 |
| threads-sustained-v2/threads/notify-empty | 4,385.94 | 6,150.40 | 4,641.94 | 4,395.27 | 6,184.52 | 16,976.17 |
| threads-sustained-v2/threads/wait-mismatch | 4,340.12 | 6,314.96 | 4,445.00 | 4,222.92 | 6,027.71 | 16,603.37 |
| threads-sustained-v2/threads/wait-timeout | 4,586.31 | 6,858.54 | 4,516.19 | 4,403.54 | 6,250.56 | 16,681.77 |
| generic-calls-v1/core/generic-calls | 4,125.85 | 6,266.69 | 4,464.65 | 4,427.92 | 6,374.88 | 16,576.83 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 42.812 | 47.583 | 35.479 | 33.375 | 37.125 | 51.541 |
| original22-v1/core/integer-loop | 41.791 | 52.438 | 50.020 | 35.499 | 39.771 | 46.688 |
| original22-v1/core/function-calls | 39.979 | 45.854 | 41.916 | 31.146 | 41.584 | 45.854 |
| original22-v1/core/memory-load-store | 47.146 | 53.375 | 43.146 | 35.916 | 49.563 | 45.542 |
| original22-v1/memory/narrow-unaligned | 40.645 | 41.938 | 40.729 | 39.812 | 34.624 | 48.020 |
| original22-v1/memory/dependent-address | 37.687 | 44.792 | 43.959 | 38.750 | 43.021 | 45.770 |
| original22-v1/memory/float64-unaligned | 40.938 | 58.791 | 38.895 | 33.500 | 39.458 | 46.271 |
| original22-v1/runtime/simd-memory-full | 41.625 | 44.729 | 43.333 | 33.396 | 46.917 | 52.230 |
| original22-v1/runtime/simd-memory-widen | 41.187 | 46.083 | 44.291 | 39.979 | 34.562 | 43.166 |
| original22-v1/runtime/simd-memory-lanes | 50.792 | 44.667 | 55.397 | 41.667 | 50.645 | 54.041 |
| original22-v1/gc/i31-roundtrip | 38.125 | 37.126 | US | 29.166 | US | 44.146 |
| original22-v1/gc/struct-access | 46.062 | 43.667 | US | 36.625 | US | 51.334 |
| original22-v1/gc/struct-allocation | 42.855 | 43.520 | US | 35.812 | US | 44.979 |
| original22-v1/gc/struct-indirect-access | 37.000 | 42.999 | US | 40.083 | US | 47.730 |
| original22-v1/threads/private-memory | 34.855 | 36.584 | 32.334 | 29.833 | 33.605 | 43.875 |
| original22-v1/threads/shared-memory | 34.584 | 41.105 | 37.062 | 35.166 | 34.667 | 44.959 |
| original22-v1/threads/atomic-rmw | 36.855 | 41.750 | 36.979 | 38.855 | 34.417 | 47.978 |
| original22-v1/threads/notify-empty | 42.354 | 46.646 | 41.041 | 29.562 | 37.041 | 42.479 |
| original22-v1/threads/wait-mismatch | 34.751 | 46.395 | 34.104 | 32.438 | 38.730 | 47.603 |
| original22-v1/threads/wait-timeout | 40.688 | 41.375 | 32.917 | 41.104 | 39.479 | 44.834 |
| original22-v1/applications/sha256-clang | 48.041 | 56.000 | 46.063 | 46.874 | 43.875 | 52.312 |
| original22-v1/applications/sha256-rust | 41.209 | 48.479 | 41.874 | 32.875 | 44.688 | 54.854 |
| gc-sustained-v1/gc/i31-roundtrip | 43.604 | 60.292 | US | 39.562 | US | 49.833 |
| gc-sustained-v1/gc/struct-access | 45.562 | 49.396 | US | 38.062 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 36.645 | 43.249 | US | 32.000 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 44.272 | 42.541 | US | 40.876 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 33.854 | 47.229 | 38.708 | 28.625 | 51.480 | 50.646 |
| threads-sustained-v2/threads/shared-memory | 33.062 | 40.187 | 32.396 | 29.688 | 36.771 | 36.666 |
| threads-sustained-v2/threads/atomic-rmw | 41.250 | 56.834 | 41.230 | 40.812 | 38.125 | 44.188 |
| threads-sustained-v2/threads/notify-empty | 44.938 | 49.333 | 42.646 | 31.855 | 35.646 | 49.980 |
| threads-sustained-v2/threads/wait-mismatch | 38.229 | 41.646 | 39.312 | 32.271 | 34.230 | 36.417 |
| threads-sustained-v2/threads/wait-timeout | 35.146 | 42.250 | 37.792 | 31.459 | 44.334 | 44.688 |
| generic-calls-v1/core/generic-calls | 40.750 | 44.375 | 39.667 | 34.563 | 36.417 | 53.709 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.708 | 5.750 | 56.438 | 136.31 | 47.896 | 28.521 |
| original22-v1/core/integer-loop | 7.958 | 6.396 | 67.500 | 150.35 | 45.979 | 31.688 |
| original22-v1/core/function-calls | 6.292 | 5.208 | 61.084 | 132.15 | 45.354 | 26.479 |
| original22-v1/core/memory-load-store | 6.104 | 6.187 | 59.188 | 138.44 | 49.917 | 28.896 |
| original22-v1/memory/narrow-unaligned | 6.333 | 5.333 | 58.084 | 138.44 | 44.041 | 30.542 |
| original22-v1/memory/dependent-address | 5.938 | 5.187 | 60.562 | 135.58 | 46.875 | 29.125 |
| original22-v1/memory/float64-unaligned | 6.229 | 5.417 | 64.666 | 131.65 | 46.667 | 29.396 |
| original22-v1/runtime/simd-memory-full | 5.771 | 5.438 | 66.708 | 143.69 | 46.063 | 34.334 |
| original22-v1/runtime/simd-memory-widen | 6.708 | 5.271 | 68.416 | 129.75 | 44.562 | 29.395 |
| original22-v1/runtime/simd-memory-lanes | 7.146 | 5.521 | 62.833 | 142.50 | 50.812 | 32.374 |
| original22-v1/gc/i31-roundtrip | 6.625 | 5.375 | US | 128.58 | US | 28.375 |
| original22-v1/gc/struct-access | 6.625 | 5.812 | US | 133.25 | US | 31.270 |
| original22-v1/gc/struct-allocation | 7.271 | 5.646 | US | 139.42 | US | 29.688 |
| original22-v1/gc/struct-indirect-access | 6.104 | 5.583 | US | 147.04 | US | 29.166 |
| original22-v1/threads/private-memory | 6.646 | 6.188 | 56.020 | 124.23 | 41.480 | 27.771 |
| original22-v1/threads/shared-memory | 6.125 | 5.333 | 57.562 | 130.29 | 44.291 | 27.583 |
| original22-v1/threads/atomic-rmw | 5.959 | 6.021 | 59.562 | 138.88 | 46.709 | 29.416 |
| original22-v1/threads/notify-empty | 5.833 | 5.812 | 58.855 | 132.19 | 45.562 | 27.395 |
| original22-v1/threads/wait-mismatch | 6.062 | 5.438 | 59.500 | 134.85 | 49.354 | 27.375 |
| original22-v1/threads/wait-timeout | 5.875 | 5.583 | 55.980 | 138.88 | 46.167 | 31.208 |
| original22-v1/applications/sha256-clang | 6.625 | 8.124 | 65.834 | 144.90 | 52.250 | 36.354 |
| original22-v1/applications/sha256-rust | 5.479 | 5.438 | 59.312 | 127.44 | 46.645 | 31.729 |
| gc-sustained-v1/gc/i31-roundtrip | 6.354 | 5.438 | US | 143.23 | US | 33.104 |
| gc-sustained-v1/gc/struct-access | 6.979 | 6.521 | US | 146.73 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 6.125 | 6.125 | US | 133.52 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 6.604 | 5.791 | US | 138.42 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 6.625 | 6.000 | 58.084 | 133.50 | 51.312 | 28.625 |
| threads-sustained-v2/threads/shared-memory | 5.771 | 5.042 | 60.563 | 132.02 | 44.105 | 28.458 |
| threads-sustained-v2/threads/atomic-rmw | 7.417 | 5.542 | 60.166 | 140.54 | 49.626 | 30.292 |
| threads-sustained-v2/threads/notify-empty | 6.208 | 5.959 | 60.041 | 132.60 | 43.688 | 32.791 |
| threads-sustained-v2/threads/wait-mismatch | 6.042 | 5.417 | 58.625 | 130.79 | 44.812 | 28.354 |
| threads-sustained-v2/threads/wait-timeout | 5.917 | 6.688 | 58.875 | 130.85 | 51.416 | 31.521 |
| generic-calls-v1/core/generic-calls | 5.562 | 5.646 | 59.083 | 133.10 | 46.375 | 29.604 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 201.25 | 153.79 | 99.000 | 751.75 | 1,245.73 | 299,093.21 |
| original22-v1/core/integer-loop | 248.31 | 216.23 | 135.58 | 971.40 | 1,231.63 | 299,947.90 |
| original22-v1/core/function-calls | 313.81 | 272.10 | 140.92 | 901.38 | 1,302.90 | 294,927.08 |
| original22-v1/core/memory-load-store | 325.06 | 301.65 | 143.27 | 779.00 | 1,333.58 | 291,025.46 |
| original22-v1/memory/narrow-unaligned | 313.12 | 287.02 | 149.67 | 810.81 | 1,197.98 | 281,868.75 |
| original22-v1/memory/dependent-address | 301.98 | 277.77 | 163.54 | 778.19 | 1,241.38 | 284,855.06 |
| original22-v1/memory/float64-unaligned | 351.56 | 298.48 | 156.75 | 793.04 | 1,265.98 | 300,799.81 |
| original22-v1/runtime/simd-memory-full | 298.67 | 264.63 | 141.04 | 1,059.15 | 1,284.85 | 285,508.35 |
| original22-v1/runtime/simd-memory-widen | 282.19 | 243.27 | 136.60 | 1,046.44 | 1,242.69 | 283,580.90 |
| original22-v1/runtime/simd-memory-lanes | 323.27 | 270.81 | 142.04 | 1,207.25 | 1,300.06 | 299,531.10 |
| original22-v1/gc/i31-roundtrip | 216.33 | 167.08 | US | 776.17 | US | 314,012.85 |
| original22-v1/gc/struct-access | 254.98 | 214.44 | US | 880.58 | US | 286,478.67 |
| original22-v1/gc/struct-allocation | 240.69 | 205.35 | US | 1,129.73 | US | 298,018.42 |
| original22-v1/gc/struct-indirect-access | 286.79 | 261.92 | US | 1,635.88 | US | 304,916.62 |
| original22-v1/threads/private-memory | 240.92 | 195.17 | 122.92 | 765.00 | 1,193.23 | 282,130.06 |
| original22-v1/threads/shared-memory | 242.67 | 200.67 | 148.60 | 744.42 | 1,190.35 | 293,733.42 |
| original22-v1/threads/atomic-rmw | 230.04 | 183.04 | 133.48 | 795.40 | 1,309.69 | 294,143.50 |
| original22-v1/threads/notify-empty | 226.35 | 181.85 | 133.00 | 896.60 | 1,159.17 | 290,900.94 |
| original22-v1/threads/wait-mismatch | 232.98 | 190.81 | 134.02 | 853.31 | 1,176.62 | 282,819.90 |
| original22-v1/threads/wait-timeout | 239.65 | 195.12 | 140.63 | 894.52 | 1,233.50 | 286,485.69 |
| original22-v1/applications/sha256-clang | 1,601.65 | 1,944.48 | 345.81 | 1,464.75 | 2,332.65 | 313,793.08 |
| original22-v1/applications/sha256-rust | 1,383.46 | 1,732.13 | 357.65 | 1,423.58 | 1,885.31 | 301,787.38 |
| gc-sustained-v1/gc/i31-roundtrip | 227.96 | 191.50 | US | 922.73 | US | 282,498.33 |
| gc-sustained-v1/gc/struct-access | 267.25 | 224.29 | US | 929.85 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 231.00 | 197.56 | US | 1,124.92 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 284.08 | 249.31 | US | 1,546.50 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 241.10 | 201.40 | 149.02 | 807.67 | 1,177.19 | 299,116.44 |
| threads-sustained-v2/threads/shared-memory | 233.60 | 190.21 | 132.44 | 763.17 | 1,164.46 | 280,380.08 |
| threads-sustained-v2/threads/atomic-rmw | 233.54 | 196.31 | 139.48 | 865.77 | 1,202.90 | 295,123.06 |
| threads-sustained-v2/threads/notify-empty | 233.94 | 181.94 | 134.40 | 892.35 | 1,266.90 | 305,097.65 |
| threads-sustained-v2/threads/wait-mismatch | 224.38 | 199.81 | 135.85 | 855.77 | 1,204.58 | 284,807.81 |
| threads-sustained-v2/threads/wait-timeout | 219.06 | 195.67 | 134.33 | 856.75 | 1,205.06 | 290,688.12 |
| generic-calls-v1/core/generic-calls | 304.71 | 282.94 | 126.90 | 1,002.40 | 1,307.33 | 285,435.98 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 9.000 | 9.792 | 14.562 | 7.479 | 20.104 | 17.895 |
| original22-v1/core/integer-loop | 8.604 | 8.791 | 16.834 | 8.271 | 20.375 | 17.833 |
| original22-v1/core/function-calls | 9.500 | 8.333 | 14.541 | 7.479 | 19.125 | 18.021 |
| original22-v1/core/memory-load-store | 13.501 | 14.459 | 18.270 | 22.292 | 36.312 | 38.104 |
| original22-v1/memory/narrow-unaligned | 13.542 | 13.104 | 17.250 | 22.750 | 35.396 | 36.083 |
| original22-v1/memory/dependent-address | 13.646 | 13.438 | 18.125 | 21.500 | 33.980 | 40.271 |
| original22-v1/memory/float64-unaligned | 13.749 | 14.187 | 18.458 | 22.334 | 35.166 | 36.000 |
| original22-v1/runtime/simd-memory-full | 14.438 | 17.625 | 19.604 | 60.625 | 41.771 | 51.459 |
| original22-v1/runtime/simd-memory-widen | 15.479 | 14.709 | 19.041 | 57.916 | 39.104 | 46.125 |
| original22-v1/runtime/simd-memory-lanes | 16.646 | 15.562 | 19.354 | 64.396 | 42.979 | 46.333 |
| original22-v1/gc/i31-roundtrip | 8.229 | 8.812 | US | 7.854 | US | 15.979 |
| original22-v1/gc/struct-access | 9.729 | 9.459 | US | 22.979 | US | 17.146 |
| original22-v1/gc/struct-allocation | 10.376 | 10.812 | US | 54.562 | US | 21.729 |
| original22-v1/gc/struct-indirect-access | 21.666 | 23.625 | US | 66.541 | US | 22.729 |
| original22-v1/threads/private-memory | 12.604 | 12.916 | 18.708 | 20.688 | 31.771 | 35.979 |
| original22-v1/threads/shared-memory | 13.062 | 12.958 | 18.000 | 22.688 | 38.625 | 37.542 |
| original22-v1/threads/atomic-rmw | 14.479 | 13.708 | 19.271 | 22.666 | 38.020 | 37.042 |
| original22-v1/threads/notify-empty | 14.188 | 14.209 | 19.020 | 22.083 | 39.375 | 38.854 |
| original22-v1/threads/wait-mismatch | 15.229 | 12.771 | 18.938 | 22.084 | 39.312 | 37.021 |
| original22-v1/threads/wait-timeout | 14.813 | 14.021 | 17.167 | 22.312 | 39.666 | 35.791 |
| original22-v1/applications/sha256-clang | 19.875 | 20.875 | 18.730 | 66.084 | 56.375 | 47.499 |
| original22-v1/applications/sha256-rust | 18.167 | 17.105 | 20.688 | 52.916 | 40.521 | 43.416 |
| gc-sustained-v1/gc/i31-roundtrip | 9.416 | 9.625 | US | 8.979 | US | 17.645 |
| gc-sustained-v1/gc/struct-access | 10.375 | 9.688 | US | 24.395 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 9.916 | 10.125 | US | 48.749 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 22.334 | 23.895 | US | 65.125 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 13.479 | 13.791 | 17.980 | 21.459 | 35.104 | 39.625 |
| threads-sustained-v2/threads/shared-memory | 12.958 | 12.687 | 18.333 | 22.271 | 37.626 | 34.709 |
| threads-sustained-v2/threads/atomic-rmw | 14.938 | 14.104 | 20.145 | 23.666 | 38.312 | 41.229 |
| threads-sustained-v2/threads/notify-empty | 14.063 | 14.541 | 17.209 | 22.959 | 39.417 | 37.229 |
| threads-sustained-v2/threads/wait-mismatch | 14.959 | 12.812 | 18.938 | 22.438 | 40.042 | 39.062 |
| threads-sustained-v2/threads/wait-timeout | 13.291 | 14.124 | 19.062 | 21.000 | 39.146 | 40.250 |
| generic-calls-v1/core/generic-calls | 10.000 | 10.938 | 14.438 | 8.459 | 19.354 | 14.875 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.291 | 0.396 | 1.917 | 4.333 | 5.292 | 0.291 |
| original22-v1/core/integer-loop | 0.334 | 0.375 | 2.229 | 4.521 | 5.167 | 0.333 |
| original22-v1/core/function-calls | 0.250 | 0.417 | 1.938 | 4.208 | 5.021 | 0.208 |
| original22-v1/core/memory-load-store | 0.292 | 0.416 | 3.021 | 4.542 | 5.417 | 0.230 |
| original22-v1/memory/narrow-unaligned | 0.271 | 0.416 | 2.042 | 4.438 | 5.146 | 0.271 |
| original22-v1/memory/dependent-address | 0.250 | 0.416 | 1.855 | 4.146 | 4.875 | 0.271 |
| original22-v1/memory/float64-unaligned | 0.333 | 0.500 | 1.979 | 4.187 | 5.333 | 0.208 |
| original22-v1/runtime/simd-memory-full | 0.312 | 0.438 | 2.625 | 2.104 | 5.250 | 0.271 |
| original22-v1/runtime/simd-memory-widen | 0.271 | 0.396 | 1.958 | 2.146 | 4.896 | 0.291 |
| original22-v1/runtime/simd-memory-lanes | 0.354 | 0.416 | 3.063 | 2.354 | 5.104 | 0.271 |
| original22-v1/gc/i31-roundtrip | 0.292 | 0.396 | US | 3.979 | US | 0.291 |
| original22-v1/gc/struct-access | 0.250 | 0.333 | US | 3.979 | US | 0.250 |
| original22-v1/gc/struct-allocation | 0.250 | 0.354 | US | 1.833 | US | 0.230 |
| original22-v1/gc/struct-indirect-access | 0.271 | 0.417 | US | 1.958 | US | 0.208 |
| original22-v1/threads/private-memory | 0.250 | 0.396 | 2.083 | 4.334 | 4.667 | 0.250 |
| original22-v1/threads/shared-memory | 0.271 | 0.396 | 2.624 | 4.521 | 4.917 | 0.271 |
| original22-v1/threads/atomic-rmw | 0.292 | 0.417 | 2.958 | 4.396 | 4.916 | 0.291 |
| original22-v1/threads/notify-empty | 0.271 | 0.417 | 1.875 | 4.396 | 5.000 | 0.312 |
| original22-v1/threads/wait-mismatch | 0.271 | 0.396 | 2.687 | 4.229 | 5.021 | 0.230 |
| original22-v1/threads/wait-timeout | 0.292 | 0.375 | 2.271 | 4.667 | 5.125 | 0.209 |
| original22-v1/applications/sha256-clang | 0.375 | 0.417 | 2.562 | 2.062 | 7.604 | 0.312 |
| original22-v1/applications/sha256-rust | 0.292 | 0.416 | 2.604 | 1.958 | 5.146 | 0.250 |
| gc-sustained-v1/gc/i31-roundtrip | 0.312 | 0.396 | US | 4.521 | US | 0.291 |
| gc-sustained-v1/gc/struct-access | 0.292 | 0.375 | US | 4.500 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 0.250 | 0.334 | US | 1.792 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 0.291 | 0.458 | US | 2.083 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.354 | 0.459 | 2.125 | 4.458 | 5.021 | 0.250 |
| threads-sustained-v2/threads/shared-memory | 0.250 | 0.396 | 2.646 | 4.250 | 4.854 | 0.208 |
| threads-sustained-v2/threads/atomic-rmw | 0.292 | 0.500 | 3.062 | 4.584 | 5.333 | 0.291 |
| threads-sustained-v2/threads/notify-empty | 0.271 | 0.417 | 1.854 | 4.521 | 4.813 | 0.291 |
| threads-sustained-v2/threads/wait-mismatch | 0.230 | 0.438 | 2.874 | 4.063 | 4.938 | 0.250 |
| threads-sustained-v2/threads/wait-timeout | 0.334 | 0.458 | 3.146 | 4.292 | 4.833 | 0.333 |
| generic-calls-v1/core/generic-calls | 0.250 | 0.354 | 2.750 | 4.646 | 5.271 | 0.209 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 11.209 | 8.479 | 9.958 | 29.291 | 14.834 | 15.541 |
| original22-v1/core/integer-loop | 1,713.98 | 1,683.52 | 1,818.67 | 1,815.96 | 1,681.50 | 1,686.88 |
| original22-v1/core/function-calls | 803.38 | 795.75 | 1,349.50 | 971.31 | 800.15 | 810.33 |
| original22-v1/core/memory-load-store | 463.42 | 439.98 | 803.60 | 535.88 | 582.90 | 312.12 |
| original22-v1/memory/narrow-unaligned | 670.73 | 657.40 | 973.38 | 728.35 | 738.58 | 530.60 |
| original22-v1/memory/dependent-address | 1,900.81 | 1,923.79 | 3,092.17 | 2,138.96 | 2,157.83 | 1,459.75 |
| original22-v1/memory/float64-unaligned | 781.90 | 775.54 | 799.60 | 752.10 | 735.79 | 699.44 |
| original22-v1/runtime/simd-memory-full | 3,707.12 | 3,565.56 | 4,179.08 | 3,670.83 | 4,062.88 | 4,011.06 |
| original22-v1/runtime/simd-memory-widen | 1,063.54 | 1,062.94 | 1,043.04 | 1,054.48 | 1,047.08 | 1,054.88 |
| original22-v1/runtime/simd-memory-lanes | 5,267.92 | 5,151.71 | 6,218.56 | 6,180.90 | 6,396.56 | 5,201.12 |
| original22-v1/gc/i31-roundtrip | 25.729 | 21.291 | US | 48.959 | US | 31.729 |
| original22-v1/gc/struct-access | 29.917 | 33.959 | US | 53.604 | US | 130.65 |
| original22-v1/gc/struct-allocation | 416.94 | 552.25 | US | 141.96 | US | 534.65 |
| original22-v1/gc/struct-indirect-access | 29.500 | 25.771 | US | 30.791 | US | 315.50 |
| original22-v1/threads/private-memory | 15.354 | 13.126 | 18.834 | 43.834 | 26.188 | 31.063 |
| original22-v1/threads/shared-memory | 25.875 | 22.479 | 20.709 | 42.292 | 26.375 | 35.834 |
| original22-v1/threads/atomic-rmw | 34.021 | 28.875 | 27.146 | 49.916 | 55.959 | 40.438 |
| original22-v1/threads/notify-empty | 22.645 | 18.875 | 226.67 | 104.81 | 64.604 | 116.40 |
| original22-v1/threads/wait-mismatch | 25.604 | 20.333 | 266.98 | 125.67 | 39.229 | 52.584 |
| original22-v1/threads/wait-timeout | 23.584 | 19.270 | 2,813.90 | 475.19 | 1,828.10 | 654.29 |
| original22-v1/applications/sha256-clang | 1,288.08 | 1,248.25 | 1,545.50 | 1,019.85 | 1,239.85 | 1,058.38 |
| original22-v1/applications/sha256-rust | 1,205.17 | 1,159.29 | 1,480.44 | 941.62 | 945.04 | 922.69 |
| gc-sustained-v1/gc/i31-roundtrip | 1,429.54 | 1,432.77 | US | 1,955.52 | US | 1,471.90 |
| gc-sustained-v1/gc/struct-access | 1,189.29 | 1,163.69 | US | 1,170.02 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 27,262.44 | 22,098.71 | US | 3,162.83 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 1,721.46 | 1,685.94 | US | 2,574.21 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 39.125 | 37.104 | 114.27 | 140.19 | 119.50 | 126.15 |
| threads-sustained-v2/threads/shared-memory | 139.52 | 130.10 | 112.42 | 136.25 | 119.12 | 122.79 |
| threads-sustained-v2/threads/atomic-rmw | 221.71 | 209.58 | 208.50 | 240.06 | 439.81 | 212.50 |
| threads-sustained-v2/threads/notify-empty | 104.98 | 102.81 | 2,123.85 | 739.44 | 482.40 | 931.38 |
| threads-sustained-v2/threads/wait-mismatch | 122.81 | 113.98 | 2,467.75 | 940.92 | 251.12 | 347.17 |
| threads-sustained-v2/threads/wait-timeout | 107.48 | 98.209 | 27,972.50 | 4,308.08 | 17,346.96 | 6,096.73 |
| generic-calls-v1/core/generic-calls | 5,291.21 | 5,399.27 | 10,690.46 | 10,579.02 | 10,582.19 | 5,310.83 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.729 | 7.083 | 11.084 | 67.438 | 25.250 | 382.40 |
| original22-v1/core/integer-loop | 13.791 | 14.333 | 14.167 | 108.21 | 43.916 | 526.44 |
| original22-v1/core/function-calls | 11.188 | 7.167 | 8.604 | 102.67 | 37.459 | 440.71 |
| original22-v1/core/memory-load-store | 9.751 | 16.812 | 4.771 | 86.667 | 61.562 | 486.73 |
| original22-v1/memory/narrow-unaligned | 11.479 | 19.646 | 5.312 | 106.27 | 59.208 | 444.92 |
| original22-v1/memory/dependent-address | 18.500 | 19.062 | 11.917 | 120.69 | 61.791 | 468.02 |
| original22-v1/memory/float64-unaligned | 19.750 | 21.458 | 12.292 | 127.69 | 67.062 | 604.40 |
| original22-v1/runtime/simd-memory-full | 19.459 | 20.334 | 10.312 | 129.56 | 67.958 | 479.25 |
| original22-v1/runtime/simd-memory-widen | 9.896 | 12.083 | 4.875 | 95.437 | 57.291 | 499.21 |
| original22-v1/runtime/simd-memory-lanes | 20.666 | 22.479 | 12.291 | 133.46 | 69.084 | 488.06 |
| original22-v1/gc/i31-roundtrip | 5.896 | 5.896 | US | 76.562 | US | 383.31 |
| original22-v1/gc/struct-access | 19.021 | 14.124 | US | 106.40 | US | 395.10 |
| original22-v1/gc/struct-allocation | 153.54 | 99.688 | US | 144.88 | US | 2,154.27 |
| original22-v1/gc/struct-indirect-access | 12.646 | 13.938 | US | 135.35 | US | 400.85 |
| original22-v1/threads/private-memory | 6.667 | 9.584 | 4.958 | 73.271 | 47.229 | 400.17 |
| original22-v1/threads/shared-memory | 11.896 | 9.416 | 5.188 | 94.000 | 52.625 | 421.10 |
| original22-v1/threads/atomic-rmw | 11.792 | 9.521 | 4.042 | 82.771 | 57.312 | 428.44 |
| original22-v1/threads/notify-empty | 11.833 | 12.271 | 7.125 | 84.980 | 56.854 | 860.96 |
| original22-v1/threads/wait-mismatch | 13.021 | 10.334 | 6.375 | 77.938 | 57.812 | 415.52 |
| original22-v1/threads/wait-timeout | 10.688 | 9.312 | 3.583 | 79.063 | 52.001 | 474.90 |
| original22-v1/applications/sha256-clang | 16.500 | 16.646 | 10.125 | 168.90 | 81.479 | 652.37 |
| original22-v1/applications/sha256-rust | 12.062 | 12.021 | 9.000 | 122.35 | 77.625 | 450.90 |
| gc-sustained-v1/gc/i31-roundtrip | 11.604 | 7.625 | US | 99.104 | US | 449.19 |
| gc-sustained-v1/gc/struct-access | 16.104 | 16.395 | US | 99.062 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 3,412.42 | 2,143.56 | US | 909.56 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 24.271 | 17.479 | US | 156.46 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 12.584 | 12.604 | 5.625 | 103.38 | 64.896 | 527.73 |
| threads-sustained-v2/threads/shared-memory | 16.479 | 14.125 | 7.542 | 107.19 | 61.375 | 417.96 |
| threads-sustained-v2/threads/atomic-rmw | 15.271 | 15.104 | 8.104 | 97.376 | 68.520 | 465.04 |
| threads-sustained-v2/threads/notify-empty | 18.605 | 14.646 | 10.584 | 99.062 | 60.916 | 425.25 |
| threads-sustained-v2/threads/wait-mismatch | 15.625 | 16.354 | 4.125 | 88.209 | 57.271 | 435.75 |
| threads-sustained-v2/threads/wait-timeout | 15.208 | 17.312 | 4.125 | 98.438 | 72.438 | 610.10 |
| generic-calls-v1/core/generic-calls | 6.979 | 5.417 | 9.625 | 113.65 | 37.854 | 406.62 |

</details>

<details>
<summary>Interpreter: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,210.29 | 5,768.21 | 4,237.27 | NM | NM | 16,821.44 |
| original22-v1/core/integer-loop | 5,405.88 | 9,570.56 | 4,943.92 | NM | NM | 21,017.21 |
| original22-v1/core/function-calls | 4,309.65 | 6,349.71 | 4,744.71 | NM | NM | 16,621.29 |
| original22-v1/core/memory-load-store | 4,573.35 | 6,709.10 | 4,430.29 | NM | NM | 17,372.83 |
| original22-v1/memory/narrow-unaligned | 4,279.94 | 6,345.02 | 4,626.15 | NM | NM | 17,376.81 |
| original22-v1/memory/dependent-address | 4,759.27 | 6,250.58 | 4,453.15 | NM | NM | 17,196.94 |
| original22-v1/memory/float64-unaligned | 4,936.25 | 6,301.25 | 4,371.35 | NM | NM | 17,763.62 |
| original22-v1/runtime/simd-memory-full | 4,646.88 | 7,533.88 | 4,899.69 | NM | NM | 17,734.65 |
| original22-v1/runtime/simd-memory-widen | 4,269.88 | 6,127.19 | 4,489.25 | NM | NM | 16,573.85 |
| original22-v1/runtime/simd-memory-lanes | 4,436.56 | 6,302.48 | 4,768.75 | NM | NM | 17,152.19 |
| original22-v1/gc/i31-roundtrip | 4,270.85 | 5,974.77 | US | NM | NM | 16,410.69 |
| original22-v1/gc/struct-access | 4,541.19 | 6,354.81 | US | NM | NM | 16,677.60 |
| original22-v1/gc/struct-allocation | 4,895.63 | 6,314.56 | US | NM | NM | 18,482.29 |
| original22-v1/gc/struct-indirect-access | 4,455.48 | 6,137.65 | US | NM | NM | 16,478.54 |
| original22-v1/threads/private-memory | 4,205.29 | 5,703.00 | 4,029.88 | NM | NM | 16,046.62 |
| original22-v1/threads/shared-memory | 4,154.06 | 6,011.21 | 4,247.31 | NM | NM | 16,400.98 |
| original22-v1/threads/atomic-rmw | 4,437.46 | 6,877.67 | 4,607.40 | NM | NM | 17,187.60 |
| original22-v1/threads/notify-empty | 4,317.46 | 6,309.58 | 4,314.98 | NM | NM | 16,871.69 |
| original22-v1/threads/wait-mismatch | 4,266.46 | 6,340.12 | 4,410.23 | NM | NM | 16,824.35 |
| original22-v1/threads/wait-timeout | 4,114.35 | 6,479.17 | 4,243.85 | NM | NM | 16,278.06 |
| original22-v1/applications/sha256-clang | 5,021.00 | 9,031.42 | 4,662.02 | NM | NM | 21,912.60 |
| original22-v1/applications/sha256-rust | 4,277.25 | 6,184.65 | 4,498.90 | NM | NM | 16,425.35 |
| gc-sustained-v1/gc/i31-roundtrip | 5,658.96 | 7,578.92 | US | NM | NM | 18,110.67 |
| gc-sustained-v1/gc/struct-access | 4,650.27 | 6,672.62 | US | NM | NM | 17,795.46 |
| gc-sustained-v1/gc/struct-allocation | 4,892.81 | 6,631.79 | US | NM | NM | 17,218.58 |
| gc-sustained-v1/gc/struct-indirect-access | 4,911.75 | 7,779.56 | US | NM | NM | 19,267.42 |
| threads-sustained-v2/threads/private-memory | 5,026.15 | 6,435.15 | 4,839.69 | NM | NM | 17,550.15 |
| threads-sustained-v2/threads/shared-memory | 4,131.40 | 6,222.83 | 4,606.48 | NM | NM | 17,080.71 |
| threads-sustained-v2/threads/atomic-rmw | 4,503.94 | 6,442.06 | 4,795.19 | NM | NM | 17,572.02 |
| threads-sustained-v2/threads/notify-empty | 4,322.27 | 6,366.38 | 4,401.46 | NM | NM | 16,171.02 |
| threads-sustained-v2/threads/wait-mismatch | 4,145.38 | 6,329.17 | 4,489.10 | NM | NM | 16,491.10 |
| threads-sustained-v2/threads/wait-timeout | 4,473.19 | 6,089.35 | 4,672.88 | NM | NM | 16,945.23 |
| generic-calls-v1/core/generic-calls | 4,423.40 | 6,248.08 | 4,555.88 | NM | NM | 17,272.08 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 41.438 | 40.042 | 38.750 | NM | NM | 47.209 |
| original22-v1/core/integer-loop | 49.041 | 65.167 | 45.938 | NM | NM | 62.250 |
| original22-v1/core/function-calls | 39.854 | 38.479 | 45.479 | NM | NM | 48.437 |
| original22-v1/core/memory-load-store | 50.251 | 46.250 | 46.667 | NM | NM | 48.771 |
| original22-v1/memory/narrow-unaligned | 41.291 | 43.647 | 39.500 | NM | NM | 49.958 |
| original22-v1/memory/dependent-address | 40.312 | 44.750 | 45.334 | NM | NM | 49.459 |
| original22-v1/memory/float64-unaligned | 38.791 | 40.958 | 42.145 | NM | NM | 45.874 |
| original22-v1/runtime/simd-memory-full | 50.604 | 49.375 | 48.979 | NM | NM | 47.834 |
| original22-v1/runtime/simd-memory-widen | 38.062 | 44.229 | 44.938 | NM | NM | 41.792 |
| original22-v1/runtime/simd-memory-lanes | 50.688 | 53.855 | 57.125 | NM | NM | 56.917 |
| original22-v1/gc/i31-roundtrip | 36.562 | 40.124 | US | NM | NM | 47.833 |
| original22-v1/gc/struct-access | 40.354 | 47.209 | US | NM | NM | 53.105 |
| original22-v1/gc/struct-allocation | 42.104 | 39.584 | US | NM | NM | 45.062 |
| original22-v1/gc/struct-indirect-access | 34.166 | 43.499 | US | NM | NM | 49.979 |
| original22-v1/threads/private-memory | 33.770 | 39.125 | 29.459 | NM | NM | 41.792 |
| original22-v1/threads/shared-memory | 39.271 | 44.750 | 36.645 | NM | NM | 44.312 |
| original22-v1/threads/atomic-rmw | 44.646 | 42.396 | 39.624 | NM | NM | 43.979 |
| original22-v1/threads/notify-empty | 31.583 | 51.958 | 38.229 | NM | NM | 40.855 |
| original22-v1/threads/wait-mismatch | 37.126 | 44.605 | 40.709 | NM | NM | 45.312 |
| original22-v1/threads/wait-timeout | 39.146 | 43.896 | 40.021 | NM | NM | 42.396 |
| original22-v1/applications/sha256-clang | 51.791 | 61.562 | 45.146 | NM | NM | 59.416 |
| original22-v1/applications/sha256-rust | 40.209 | 45.437 | 39.104 | NM | NM | 48.458 |
| gc-sustained-v1/gc/i31-roundtrip | 43.020 | 53.000 | US | NM | NM | 59.563 |
| gc-sustained-v1/gc/struct-access | 42.730 | 50.000 | US | NM | NM | 45.896 |
| gc-sustained-v1/gc/struct-allocation | 38.625 | 43.980 | US | NM | NM | 43.605 |
| gc-sustained-v1/gc/struct-indirect-access | 46.500 | 51.105 | US | NM | NM | 45.271 |
| threads-sustained-v2/threads/private-memory | 47.666 | 48.354 | 43.334 | NM | NM | 45.562 |
| threads-sustained-v2/threads/shared-memory | 36.542 | 43.896 | 33.500 | NM | NM | 46.834 |
| threads-sustained-v2/threads/atomic-rmw | 37.354 | 44.146 | 46.041 | NM | NM | 40.396 |
| threads-sustained-v2/threads/notify-empty | 35.938 | 48.979 | 38.666 | NM | NM | 43.937 |
| threads-sustained-v2/threads/wait-mismatch | 34.459 | 45.062 | 35.062 | NM | NM | 43.062 |
| threads-sustained-v2/threads/wait-timeout | 47.021 | 41.708 | 47.291 | NM | NM | 50.333 |
| generic-calls-v1/core/generic-calls | 44.334 | 43.229 | 34.271 | NM | NM | 56.521 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4.875 | 5.312 | 2.312 | NM | NM | 28.125 |
| original22-v1/core/integer-loop | 5.792 | 6.604 | 2.333 | NM | NM | 38.749 |
| original22-v1/core/function-calls | 4.729 | 5.188 | 2.479 | NM | NM | 27.042 |
| original22-v1/core/memory-load-store | 5.208 | 5.521 | 3.021 | NM | NM | 26.979 |
| original22-v1/memory/narrow-unaligned | 4.792 | 5.562 | 2.583 | NM | NM | 26.312 |
| original22-v1/memory/dependent-address | 4.458 | 5.125 | 3.083 | NM | NM | 28.270 |
| original22-v1/memory/float64-unaligned | 4.625 | 5.542 | 2.417 | NM | NM | 32.084 |
| original22-v1/runtime/simd-memory-full | 5.542 | 6.208 | 3.458 | NM | NM | 26.770 |
| original22-v1/runtime/simd-memory-widen | 4.146 | 5.125 | 2.625 | NM | NM | 28.709 |
| original22-v1/runtime/simd-memory-lanes | 5.021 | 6.062 | 2.833 | NM | NM | 30.875 |
| original22-v1/gc/i31-roundtrip | 5.042 | 5.208 | US | NM | NM | 28.167 |
| original22-v1/gc/struct-access | 4.562 | 5.833 | US | NM | NM | 28.291 |
| original22-v1/gc/struct-allocation | 5.333 | 5.271 | US | NM | NM | 33.624 |
| original22-v1/gc/struct-indirect-access | 5.229 | 5.542 | US | NM | NM | 26.646 |
| original22-v1/threads/private-memory | 4.604 | 4.896 | 2.417 | NM | NM | 26.834 |
| original22-v1/threads/shared-memory | 4.813 | 5.250 | 2.376 | NM | NM | 26.729 |
| original22-v1/threads/atomic-rmw | 5.208 | 5.208 | 2.479 | NM | NM | 29.584 |
| original22-v1/threads/notify-empty | 4.417 | 5.458 | 2.833 | NM | NM | 27.708 |
| original22-v1/threads/wait-mismatch | 5.354 | 5.583 | 2.333 | NM | NM | 28.166 |
| original22-v1/threads/wait-timeout | 4.792 | 5.083 | 2.792 | NM | NM | 25.812 |
| original22-v1/applications/sha256-clang | 5.708 | 6.958 | 2.667 | NM | NM | 35.876 |
| original22-v1/applications/sha256-rust | 4.479 | 5.188 | 2.521 | NM | NM | 28.312 |
| gc-sustained-v1/gc/i31-roundtrip | 5.166 | 6.271 | US | NM | NM | 33.979 |
| gc-sustained-v1/gc/struct-access | 5.208 | 5.771 | US | NM | NM | 30.562 |
| gc-sustained-v1/gc/struct-allocation | 4.542 | 5.313 | US | NM | NM | 27.646 |
| gc-sustained-v1/gc/struct-indirect-access | 4.625 | 5.729 | US | NM | NM | 31.354 |
| threads-sustained-v2/threads/private-memory | 4.688 | 5.479 | 2.771 | NM | NM | 29.334 |
| threads-sustained-v2/threads/shared-memory | 4.729 | 5.312 | 2.146 | NM | NM | 29.041 |
| threads-sustained-v2/threads/atomic-rmw | 4.583 | 5.604 | 2.833 | NM | NM | 27.562 |
| threads-sustained-v2/threads/notify-empty | 4.562 | 5.167 | 2.479 | NM | NM | 26.855 |
| threads-sustained-v2/threads/wait-mismatch | 4.667 | 5.812 | 2.313 | NM | NM | 27.813 |
| threads-sustained-v2/threads/wait-timeout | 5.813 | 5.292 | 2.583 | NM | NM | 32.396 |
| generic-calls-v1/core/generic-calls | 4.396 | 5.583 | 2.417 | NM | NM | 28.146 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 59.312 | 48.125 | 39.416 | NM | NM | 38.875 |
| original22-v1/core/integer-loop | 74.438 | 62.312 | 47.770 | NM | NM | 41.395 |
| original22-v1/core/function-calls | 73.375 | 66.917 | 44.626 | NM | NM | 38.500 |
| original22-v1/core/memory-load-store | 82.604 | 77.875 | 48.438 | NM | NM | 41.062 |
| original22-v1/memory/narrow-unaligned | 83.271 | 84.271 | 48.729 | NM | NM | 41.834 |
| original22-v1/memory/dependent-address | 81.229 | 81.417 | 47.730 | NM | NM | 43.271 |
| original22-v1/memory/float64-unaligned | 88.521 | 81.645 | 50.709 | NM | NM | 48.876 |
| original22-v1/runtime/simd-memory-full | 77.834 | 76.229 | 57.584 | NM | NM | 43.354 |
| original22-v1/runtime/simd-memory-widen | 73.605 | 65.312 | 51.979 | NM | NM | 45.562 |
| original22-v1/runtime/simd-memory-lanes | 79.708 | 75.187 | 55.500 | NM | NM | 43.812 |
| original22-v1/gc/i31-roundtrip | 64.041 | 54.375 | US | NM | NM | 37.395 |
| original22-v1/gc/struct-access | 67.041 | 60.291 | US | NM | NM | 49.334 |
| original22-v1/gc/struct-allocation | 75.584 | 60.480 | US | NM | NM | 50.145 |
| original22-v1/gc/struct-indirect-access | 81.875 | 74.480 | US | NM | NM | 48.396 |
| original22-v1/threads/private-memory | 63.937 | 55.354 | 41.334 | NM | NM | 39.750 |
| original22-v1/threads/shared-memory | 65.916 | 57.459 | 44.042 | NM | NM | 40.209 |
| original22-v1/threads/atomic-rmw | 65.229 | 53.792 | 49.041 | NM | NM | 42.645 |
| original22-v1/threads/notify-empty | 63.791 | 55.584 | 45.521 | NM | NM | 38.271 |
| original22-v1/threads/wait-mismatch | 67.376 | 58.166 | 49.105 | NM | NM | 39.812 |
| original22-v1/threads/wait-timeout | 70.937 | 55.020 | 44.875 | NM | NM | 39.562 |
| original22-v1/applications/sha256-clang | 314.00 | 456.81 | 89.312 | NM | NM | 105.75 |
| original22-v1/applications/sha256-rust | 265.42 | 369.27 | 93.499 | NM | NM | 82.416 |
| gc-sustained-v1/gc/i31-roundtrip | 70.896 | 61.666 | US | NM | NM | 42.584 |
| gc-sustained-v1/gc/struct-access | 68.355 | 61.749 | US | NM | NM | 53.209 |
| gc-sustained-v1/gc/struct-allocation | 67.709 | 62.583 | US | NM | NM | 48.167 |
| gc-sustained-v1/gc/struct-indirect-access | 79.688 | 77.562 | US | NM | NM | 53.833 |
| threads-sustained-v2/threads/private-memory | 63.459 | 55.334 | 45.271 | NM | NM | 41.438 |
| threads-sustained-v2/threads/shared-memory | 63.770 | 55.041 | 42.084 | NM | NM | 39.771 |
| threads-sustained-v2/threads/atomic-rmw | 65.542 | 57.874 | 47.000 | NM | NM | 41.084 |
| threads-sustained-v2/threads/notify-empty | 63.896 | 54.625 | 45.209 | NM | NM | 40.395 |
| threads-sustained-v2/threads/wait-mismatch | 64.083 | 58.584 | 44.291 | NM | NM | 41.562 |
| threads-sustained-v2/threads/wait-timeout | 65.208 | 57.041 | 46.834 | NM | NM | 40.772 |
| generic-calls-v1/core/generic-calls | 74.916 | 73.125 | 39.125 | NM | NM | 45.605 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 8.417 | 8.021 | 15.521 | NM | NM | 26.479 |
| original22-v1/core/integer-loop | 9.604 | 8.854 | 17.855 | NM | NM | 40.730 |
| original22-v1/core/function-calls | 8.688 | 7.917 | 15.083 | NM | NM | 26.958 |
| original22-v1/core/memory-load-store | 12.271 | 12.791 | 17.250 | NM | NM | 47.625 |
| original22-v1/memory/narrow-unaligned | 12.354 | 13.375 | 17.146 | NM | NM | 53.083 |
| original22-v1/memory/dependent-address | 11.563 | 13.354 | 16.688 | NM | NM | 51.355 |
| original22-v1/memory/float64-unaligned | 12.437 | 13.688 | 16.959 | NM | NM | 53.020 |
| original22-v1/runtime/simd-memory-full | 14.479 | 13.875 | 19.021 | NM | NM | 57.999 |
| original22-v1/runtime/simd-memory-widen | 12.604 | 12.917 | 16.834 | NM | NM | 56.646 |
| original22-v1/runtime/simd-memory-lanes | 13.062 | 14.541 | 18.270 | NM | NM | 56.541 |
| original22-v1/gc/i31-roundtrip | 8.833 | 8.167 | US | NM | NM | 27.646 |
| original22-v1/gc/struct-access | 9.376 | 9.104 | US | NM | NM | 26.020 |
| original22-v1/gc/struct-allocation | 9.729 | 9.562 | US | NM | NM | 36.459 |
| original22-v1/gc/struct-indirect-access | 15.896 | 16.542 | US | NM | NM | 33.042 |
| original22-v1/threads/private-memory | 10.896 | 12.688 | 15.979 | NM | NM | 47.438 |
| original22-v1/threads/shared-memory | 15.312 | 13.166 | 16.646 | NM | NM | 49.980 |
| original22-v1/threads/atomic-rmw | 16.229 | 13.375 | 16.917 | NM | NM | 53.146 |
| original22-v1/threads/notify-empty | 14.604 | 14.083 | 17.125 | NM | NM | 54.750 |
| original22-v1/threads/wait-mismatch | 15.812 | 13.188 | 16.709 | NM | NM | 50.583 |
| original22-v1/threads/wait-timeout | 15.771 | 12.854 | 18.083 | NM | NM | 46.853 |
| original22-v1/applications/sha256-clang | 17.959 | 24.729 | 20.209 | NM | NM | 91.521 |
| original22-v1/applications/sha256-rust | 14.729 | 75.688 | 21.855 | NM | NM | 57.520 |
| gc-sustained-v1/gc/i31-roundtrip | 9.584 | 8.416 | US | NM | NM | 30.312 |
| gc-sustained-v1/gc/struct-access | 9.750 | 8.979 | US | NM | NM | 33.896 |
| gc-sustained-v1/gc/struct-allocation | 9.375 | 10.062 | US | NM | NM | 32.105 |
| gc-sustained-v1/gc/struct-indirect-access | 15.000 | 16.416 | US | NM | NM | 37.521 |
| threads-sustained-v2/threads/private-memory | 11.416 | 12.750 | 18.334 | NM | NM | 44.626 |
| threads-sustained-v2/threads/shared-memory | 14.771 | 12.604 | 15.666 | NM | NM | 49.355 |
| threads-sustained-v2/threads/atomic-rmw | 15.208 | 14.104 | 17.667 | NM | NM | 50.667 |
| threads-sustained-v2/threads/notify-empty | 14.479 | 14.146 | 16.709 | NM | NM | 47.979 |
| threads-sustained-v2/threads/wait-mismatch | 14.750 | 13.542 | 16.688 | NM | NM | 49.709 |
| threads-sustained-v2/threads/wait-timeout | 17.438 | 13.937 | 16.833 | NM | NM | 54.500 |
| generic-calls-v1/core/generic-calls | 9.896 | 9.625 | 16.896 | NM | NM | 28.146 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.271 | 0.396 | 0.230 | NM | NM | 0.271 |
| original22-v1/core/integer-loop | 0.312 | 0.520 | 0.250 | NM | NM | 0.271 |
| original22-v1/core/function-calls | 0.250 | 0.334 | 0.250 | NM | NM | 0.271 |
| original22-v1/core/memory-load-store | 0.354 | 0.438 | 0.292 | NM | NM | 0.209 |
| original22-v1/memory/narrow-unaligned | 0.312 | 0.375 | 0.208 | NM | NM | 0.209 |
| original22-v1/memory/dependent-address | 0.250 | 0.375 | 0.167 | NM | NM | 0.250 |
| original22-v1/memory/float64-unaligned | 0.312 | 0.396 | 0.230 | NM | NM | 0.291 |
| original22-v1/runtime/simd-memory-full | 0.312 | 0.396 | 0.292 | NM | NM | 0.230 |
| original22-v1/runtime/simd-memory-widen | 0.209 | 0.375 | 0.209 | NM | NM | 0.229 |
| original22-v1/runtime/simd-memory-lanes | 0.312 | 0.416 | 0.291 | NM | NM | 0.250 |
| original22-v1/gc/i31-roundtrip | 0.230 | 0.396 | US | NM | NM | 0.250 |
| original22-v1/gc/struct-access | 0.230 | 0.334 | US | NM | NM | 0.208 |
| original22-v1/gc/struct-allocation | 0.229 | 0.396 | US | NM | NM | 0.250 |
| original22-v1/gc/struct-indirect-access | 0.291 | 0.333 | US | NM | NM | 0.208 |
| original22-v1/threads/private-memory | 0.209 | 0.396 | 0.167 | NM | NM | 0.250 |
| original22-v1/threads/shared-memory | 0.250 | 0.416 | 0.209 | NM | NM | 0.250 |
| original22-v1/threads/atomic-rmw | 0.271 | 0.417 | 0.250 | NM | NM | 0.208 |
| original22-v1/threads/notify-empty | 0.271 | 0.438 | 0.229 | NM | NM | 0.208 |
| original22-v1/threads/wait-mismatch | 0.292 | 0.417 | 0.230 | NM | NM | 0.209 |
| original22-v1/threads/wait-timeout | 0.334 | 0.375 | 0.334 | NM | NM | 0.208 |
| original22-v1/applications/sha256-clang | 0.312 | 0.479 | 0.208 | NM | NM | 0.291 |
| original22-v1/applications/sha256-rust | 0.208 | 0.396 | 0.271 | NM | NM | 0.208 |
| gc-sustained-v1/gc/i31-roundtrip | 0.291 | 0.458 | US | NM | NM | 0.250 |
| gc-sustained-v1/gc/struct-access | 0.291 | 0.354 | US | NM | NM | 0.250 |
| gc-sustained-v1/gc/struct-allocation | 0.209 | 0.396 | US | NM | NM | 0.230 |
| gc-sustained-v1/gc/struct-indirect-access | 0.250 | 0.354 | US | NM | NM | 0.250 |
| threads-sustained-v2/threads/private-memory | 0.312 | 0.438 | 0.250 | NM | NM | 0.208 |
| threads-sustained-v2/threads/shared-memory | 0.250 | 0.417 | 0.188 | NM | NM | 0.208 |
| threads-sustained-v2/threads/atomic-rmw | 0.312 | 0.396 | 0.250 | NM | NM | 0.208 |
| threads-sustained-v2/threads/notify-empty | 0.292 | 0.396 | 0.208 | NM | NM | 0.209 |
| threads-sustained-v2/threads/wait-mismatch | 0.250 | 0.459 | 0.188 | NM | NM | 0.208 |
| threads-sustained-v2/threads/wait-timeout | 0.312 | 0.438 | 0.229 | NM | NM | 0.271 |
| generic-calls-v1/core/generic-calls | 0.291 | 0.354 | 0.208 | NM | NM | 0.208 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.562 | 7.938 | 4.271 | NM | NM | 8.166 |
| original22-v1/core/integer-loop | 41,837.94 | 30,482.48 | 93,398.25 | NM | NM | 60,033.65 |
| original22-v1/core/function-calls | 54,874.90 | 46,684.98 | 88,480.54 | NM | NM | 63,511.33 |
| original22-v1/core/memory-load-store | 77,581.35 | 60,838.77 | 154,532.77 | NM | NM | 95,319.13 |
| original22-v1/memory/narrow-unaligned | 96,262.46 | 77,664.85 | 189,699.75 | NM | NM | 118,387.73 |
| original22-v1/memory/dependent-address | 48,355.79 | 36,553.81 | 100,539.67 | NM | NM | 55,968.69 |
| original22-v1/memory/float64-unaligned | 79,752.62 | 63,605.73 | 139,882.33 | NM | NM | 87,876.60 |
| original22-v1/runtime/simd-memory-full | 88,339.90 | 73,234.71 | 162,109.08 | NM | NM | 83,337.48 |
| original22-v1/runtime/simd-memory-widen | 67,799.63 | 56,484.02 | 146,470.83 | NM | NM | 62,369.96 |
| original22-v1/runtime/simd-memory-lanes | 108,229.35 | 92,138.10 | 206,405.67 | NM | NM | 98,111.27 |
| original22-v1/gc/i31-roundtrip | 424.19 | 352.52 | US | NM | NM | 459.54 |
| original22-v1/gc/struct-access | 484.35 | 442.29 | US | NM | NM | 581.54 |
| original22-v1/gc/struct-allocation | 981.06 | 1,044.50 | US | NM | NM | 1,229.48 |
| original22-v1/gc/struct-indirect-access | 878.10 | 754.21 | US | NM | NM | 986.44 |
| original22-v1/threads/private-memory | 430.15 | 359.73 | 648.77 | NM | NM | 401.08 |
| original22-v1/threads/shared-memory | 808.23 | 367.52 | 641.15 | NM | NM | 413.04 |
| original22-v1/threads/atomic-rmw | 569.90 | 364.85 | 500.21 | NM | NM | 355.15 |
| original22-v1/threads/notify-empty | 357.77 | 293.27 | 440.67 | NM | NM | 423.85 |
| original22-v1/threads/wait-mismatch | 602.73 | 401.62 | 703.92 | NM | NM | 464.38 |
| original22-v1/threads/wait-timeout | 607.04 | 375.85 | 3,213.02 | NM | NM | 1,057.10 |
| original22-v1/applications/sha256-clang | 109,111.56 | 90,030.21 | 216,278.15 | NM | NM | 175,392.62 |
| original22-v1/applications/sha256-rust | 98,007.23 | 79,946.06 | 206,826.10 | NM | NM | 122,274.44 |
| gc-sustained-v1/gc/i31-roundtrip | 47,219.52 | 35,417.08 | US | NM | NM | 50,473.79 |
| gc-sustained-v1/gc/struct-access | 51,041.31 | 41,545.73 | US | NM | NM | 58,668.56 |
| gc-sustained-v1/gc/struct-allocation | 52,584.60 | 36,140.21 | US | NM | NM | 33,662.19 |
| gc-sustained-v1/gc/struct-indirect-access | 81,823.50 | 75,274.23 | US | NM | NM | 100,664.06 |
| threads-sustained-v2/threads/private-memory | 4,278.33 | 3,459.52 | 6,352.67 | NM | NM | 4,177.62 |
| threads-sustained-v2/threads/shared-memory | 7,887.94 | 3,580.00 | 6,519.25 | NM | NM | 3,941.79 |
| threads-sustained-v2/threads/atomic-rmw | 5,704.25 | 3,687.33 | 5,281.69 | NM | NM | 3,478.62 |
| threads-sustained-v2/threads/notify-empty | 3,333.60 | 2,809.10 | 4,273.90 | NM | NM | 3,990.90 |
| threads-sustained-v2/threads/wait-mismatch | 6,004.65 | 3,780.21 | 6,825.83 | NM | NM | 4,432.12 |
| threads-sustained-v2/threads/wait-timeout | 5,992.04 | 3,620.88 | 33,109.71 | NM | NM | 9,930.06 |
| generic-calls-v1/core/generic-calls | 634,395.54 | 493,340.08 | 924,010.04 | NM | NM | 615,513.65 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2.146 | 1.854 | 4.896 | NM | NM | 5.562 |
| original22-v1/core/integer-loop | 2.271 | 2.188 | 12.584 | NM | NM | 11.624 |
| original22-v1/core/function-calls | 2.125 | 1.958 | 4.875 | NM | NM | 9.249 |
| original22-v1/core/memory-load-store | 2.374 | 7.333 | 4.958 | NM | NM | 52.271 |
| original22-v1/memory/narrow-unaligned | 2.417 | 7.646 | 6.521 | NM | NM | 51.062 |
| original22-v1/memory/dependent-address | 2.417 | 6.667 | 9.479 | NM | NM | 50.584 |
| original22-v1/memory/float64-unaligned | 3.438 | 7.417 | 11.166 | NM | NM | 59.938 |
| original22-v1/runtime/simd-memory-full | 2.667 | 7.500 | 4.917 | NM | NM | 51.916 |
| original22-v1/runtime/simd-memory-widen | 2.438 | 7.188 | 5.083 | NM | NM | 54.834 |
| original22-v1/runtime/simd-memory-lanes | 3.501 | 7.021 | 8.875 | NM | NM | 51.709 |
| original22-v1/gc/i31-roundtrip | 1.833 | 1.854 | US | NM | NM | 6.583 |
| original22-v1/gc/struct-access | 2.083 | 1.688 | US | NM | NM | 8.666 |
| original22-v1/gc/struct-allocation | 104.85 | 54.896 | US | NM | NM | 1,750.44 |
| original22-v1/gc/struct-indirect-access | 2.021 | 1.688 | US | NM | NM | 7.667 |
| original22-v1/threads/private-memory | 1.917 | 4.417 | 4.354 | NM | NM | 43.396 |
| original22-v1/threads/shared-memory | 11.416 | 9.876 | 5.708 | NM | NM | 42.792 |
| original22-v1/threads/atomic-rmw | 9.583 | 9.459 | 4.479 | NM | NM | 46.980 |
| original22-v1/threads/notify-empty | 13.125 | 8.771 | 7.625 | NM | NM | 44.771 |
| original22-v1/threads/wait-mismatch | 12.271 | 12.021 | 5.792 | NM | NM | 42.979 |
| original22-v1/threads/wait-timeout | 8.979 | 8.188 | 3.875 | NM | NM | 44.146 |
| original22-v1/applications/sha256-clang | 2.396 | 6.521 | 5.688 | NM | NM | 61.062 |
| original22-v1/applications/sha256-rust | 2.417 | 2.021 | 4.500 | NM | NM | 57.959 |
| gc-sustained-v1/gc/i31-roundtrip | 2.292 | 2.271 | US | NM | NM | 10.334 |
| gc-sustained-v1/gc/struct-access | 2.750 | 2.438 | US | NM | NM | 8.916 |
| gc-sustained-v1/gc/struct-allocation | 2,761.15 | 1,654.10 | US | NM | NM | 60,133.23 |
| gc-sustained-v1/gc/struct-indirect-access | 2.417 | 2.250 | US | NM | NM | 8.229 |
| threads-sustained-v2/threads/private-memory | 2.771 | 7.292 | 5.375 | NM | NM | 48.166 |
| threads-sustained-v2/threads/shared-memory | 19.250 | 12.126 | 6.125 | NM | NM | 47.000 |
| threads-sustained-v2/threads/atomic-rmw | 16.229 | 12.396 | 5.854 | NM | NM | 49.124 |
| threads-sustained-v2/threads/notify-empty | 15.917 | 16.292 | 6.000 | NM | NM | 50.979 |
| threads-sustained-v2/threads/wait-mismatch | 14.979 | 15.749 | 7.375 | NM | NM | 53.500 |
| threads-sustained-v2/threads/wait-timeout | 14.271 | 15.771 | 4.000 | NM | NM | 49.041 |
| generic-calls-v1/core/generic-calls | 2.146 | 2.021 | 4.458 | NM | NM | 10.125 |

</details>

### WasmEdge AOT preparation detail

| Workload | Compile + validation + artifact I/O, us | Native image load, us | Post-load validation, us |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 18,359.29 | 265,647.83 | 2.896 |
| original22-v1/core/integer-loop | 25,219.29 | 271,426.71 | 6.104 |
| original22-v1/core/function-calls | 21,152.40 | 274,707.17 | 5.125 |
| original22-v1/core/memory-load-store | 23,626.56 | 263,360.00 | 5.958 |
| original22-v1/memory/narrow-unaligned | 20,143.29 | 259,094.79 | 9.062 |
| original22-v1/memory/dependent-address | 21,630.00 | 262,180.33 | 7.938 |
| original22-v1/memory/float64-unaligned | 22,738.29 | 270,458.92 | 7.729 |
| original22-v1/runtime/simd-memory-full | 22,831.54 | 258,005.23 | 8.917 |
| original22-v1/runtime/simd-memory-widen | 19,928.77 | 259,744.94 | 7.000 |
| original22-v1/runtime/simd-memory-lanes | 19,937.46 | 273,132.67 | 8.499 |
| original22-v1/gc/i31-roundtrip | 18,404.98 | 282,330.29 | 6.938 |
| original22-v1/gc/struct-access | 19,925.06 | 263,905.40 | 4.042 |
| original22-v1/gc/struct-allocation | 21,360.71 | 265,790.96 | 7.021 |
| original22-v1/gc/struct-indirect-access | 20,962.23 | 285,452.85 | 8.354 |
| original22-v1/threads/private-memory | 17,885.27 | 262,428.17 | 5.479 |
| original22-v1/threads/shared-memory | 19,645.63 | 269,354.62 | 5.646 |
| original22-v1/threads/atomic-rmw | 20,446.79 | 273,916.50 | 5.083 |
| original22-v1/threads/notify-empty | 20,842.44 | 262,950.79 | 5.417 |
| original22-v1/threads/wait-mismatch | 18,747.92 | 264,345.27 | 5.604 |
| original22-v1/threads/wait-timeout | 19,450.90 | 265,209.94 | 5.000 |
| original22-v1/applications/sha256-clang | 33,783.94 | 262,043.29 | 14.583 |
| original22-v1/applications/sha256-rust | 24,666.25 | 277,301.42 | 10.125 |
| gc-sustained-v1/gc/i31-roundtrip | 24,936.40 | 256,076.58 | 5.271 |
| gc-sustained-v1/gc/struct-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 27,584.88 | 267,303.46 | 5.688 |
| threads-sustained-v2/threads/shared-memory | 19,269.27 | 260,711.79 | 5.333 |
| threads-sustained-v2/threads/atomic-rmw | 19,675.92 | 264,081.94 | 6.125 |
| threads-sustained-v2/threads/notify-empty | 19,110.17 | 271,135.27 | 6.062 |
| threads-sustained-v2/threads/wait-mismatch | 18,614.54 | 259,283.21 | 6.167 |
| threads-sustained-v2/threads/wait-timeout | 18,965.29 | 260,248.19 | 5.917 |
| generic-calls-v1/core/generic-calls | 19,785.17 | 267,498.44 | 4.938 |

### P3 phase observations

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B / cleanup | 76.395 | 713.77 | NM | 306.40 | NM | NM |
| 16 B / compileValidate | 79,970.50 | 126,684.79 | NM | 71,601.85 | NM | NM |
| 16 B / engineSetup | 0.500 | 0.229 | NM | 202.04 | NM | NM |
| 16 B / executorSetup | N/A | N/A | NM | 56.603 | NM | NM |
| 16 B / exportLookup | 0.916 | 1.230 | NM | 11.730 | NM | NM |
| 16 B / firstCall | 12,277.48 | 25,530.12 | NM | 9,335.85 | NM | NM |
| 16 B / instantiate | 635.46 | 1,033.33 | NM | 248.67 | NM | NM |
| 16 B / load | 70.000 | 83.646 | NM | 87.166 | NM | NM |
| 16 B / processStartup | 7,546.21 | 16,424.54 | NM | 8,595.17 | NM | NM |
| 16 B / providerSetup | 53.688 | 245.04 | NM | 308.25 | NM | NM |
| 16 B / requestPreparation | 0.791 | 0.896 | NM | N/A | NM | NM |
| 65536 B / cleanup | 99.791 | 807.52 | NM | 387.75 | NM | NM |
| 65536 B / compileValidate | 86,878.06 | 125,838.56 | NM | 72,472.50 | NM | NM |
| 65536 B / engineSetup | 0.438 | 0.271 | NM | 198.60 | NM | NM |
| 65536 B / executorSetup | N/A | N/A | NM | 57.541 | NM | NM |
| 65536 B / exportLookup | 1.021 | 1.438 | NM | 11.895 | NM | NM |
| 65536 B / firstCall | 18,784.15 | 29,351.23 | NM | 11,219.33 | NM | NM |
| 65536 B / instantiate | 714.58 | 1,000.04 | NM | 251.56 | NM | NM |
| 65536 B / load | 90.791 | 83.688 | NM | 64.104 | NM | NM |
| 65536 B / processStartup | 8,047.35 | 16,415.29 | NM | 7,993.10 | NM | NM |
| 65536 B / providerSetup | 58.166 | 250.25 | NM | 322.00 | NM | NM |
| 65536 B / requestPreparation | 11.376 | 13.479 | NM | N/A | NM | NM |

## Native controls (different contracts)

Unchecked C kernels omit runtime safety/lifetime contracts; native Go implements HTTP replies without Component/WASI. These are overhead oracles, not equivalent competitors, and are never subtracted from runtime timings. The general-call C control runs twice as many iterations of one recurrence; unlike the optimized Wasm it does not represent the same two-update affine composition as one operation. Beating that control does not establish a win over the best possible native implementation.

| Workload | Unchecked native C, ns/iteration |
| --- | ---: |
| original22-v1/core/integer-loop | 1.842 |
| original22-v1/core/function-calls | 0.812 |
| original22-v1/core/memory-load-store | 0.302 |
| original22-v1/memory/narrow-unaligned | 0.461 |
| original22-v1/memory/dependent-address | 1.469 |
| original22-v1/memory/float64-unaligned | 0.688 |
| original22-v1/runtime/simd-memory-full | 3.653 |
| original22-v1/runtime/simd-memory-widen | 1.046 |
| original22-v1/runtime/simd-memory-lanes | 4.715 |
| original22-v1/applications/sha256-clang | 183.69 |
| original22-v1/applications/sha256-rust | 167.18 |
| generic-calls-v1/core/generic-calls | 2.068 |

| Payload | Native Go CPU, us/request | Native Go requests/s |
| --- | ---: | ---: |
| 16 B | 4.187 | 980,144.08 |
| 65536 B | 136.11 | 18,941.71 |

## WasmEdge sustained AOT failures

Full-size WasmEdge0.17.1 O3 AOT preflight freshly reproduces SIGSEGV in sustained struct access, allocation and indirect access. Small original cases pass. Only those exact three cells remain failures in the timing matrix. The original investigation and the observed cumulative stack growth are retained in the [PR #200 report](performance-pr200-complete-comparison-2026-11-17.md). All table timings come from the current event.

## Generated-code mechanism

The emitted ARM64 field kernel uses one guarded loop of 1,776 bytes for the function, including cold entries, versus 1,644 bytes in the baseline audit. Its private loop retains the field store and control poll. The final body omits the three cache key checks, the boolean spill, and the redundant narrowing of an already canonical i32 value. Initial guard-only and guard-plus-predicate screening runs showed little timing change. The third screening run first showed a material improvement after removing the redundant narrowing. Screening observations and primary timing use separate fixed plans.

## Resolved adverse primary intervals

| Workload or metric | Source comparison | Head/base ratio and delta |
| --- | ---: | ---: |
| original22-v1/memory/narrow-unaligned | head-llgo-native/base-llgo-native | 1.006 [1.001, 1.010]; Δ +0.58%; speed -0.57% [-1.00%, -0.11%]; pass |
| original22-v1/memory/narrow-unaligned | head-go-interpreter/base-go-interpreter | 1.105 [1.090, 1.121]; Δ +10.54%; speed -9.53% [-10.76%, -8.27%]; pass |
| original22-v1/memory/dependent-address | head-go-interpreter/base-go-interpreter | 1.111 [1.104, 1.119]; Δ +11.10%; speed -9.99% [-10.63%, -9.42%]; pass |
| original22-v1/applications/sha256-rust | head-go-interpreter/base-go-interpreter | 1.076 [1.070, 1.080]; Δ +7.57%; speed -7.03% [-7.37%, -6.59%]; pass |
| threads-sustained-v2/threads/shared-memory | head-go-interpreter/base-go-interpreter | 1.060 [1.045, 1.080]; Δ +6.00%; speed -5.66% [-7.41%, -4.27%]; pass |
| P3 16 B allocsPerRequest | head-llgo/base-llgo | 1.000 [1.000, 1.000]; Δ +0.00%; pass |
| P3 65536 B allocsPerRequest | head-go/base-go | 1.000 [1.000, 1.000]; Δ +0.00%; pass |

## Interpreter regression investigation

The default Go build moved `runSlotsLoop` from `0x1002120f0` to `0x100212e40`. The function has 10,112 bytes in both builds. Its instruction listing changes PC-relative addresses for its jump table and linked data. A separate 96-process diagnostic reproduced higher CPU cycles per instruction while the number of executed instructions remained stable. Relinking both revisions with `-ldflags=-funcalign=64` put the function at `0x10022ba40` in both. Across 80 additional diagnostic processes the excess cycles per instruction fell to within 0.7% for all five Go interpreter workloads. This isolates sensitivity to the linked layout as a cause of the observed regressions. The intervention changes other function addresses too, so the exact hardware mechanism remains open. All primary results use the default linker flags.

| Go interpreter workload | Default CPU Δ | Default instructions Δ | Default cycles/instruction Δ | Aligned CPU Δ | Aligned cycles/instruction Δ |
| --- | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-rust | +7.77% | +0.02% | +7.35% | -1.16% | -0.64% |
| core/integer-loop | +12.34% | -0.01% | +14.12% | +2.02% | -0.28% |
| memory/dependent-address | +10.84% | -0.01% | +10.45% | +2.69% | +0.54% |
| memory/narrow-unaligned | +8.82% | -0.03% | +8.40% | -0.92% | -0.24% |
| threads/shared-memory | +9.91% | +0.03% | +7.17% | -2.18% | -0.69% |

The same narrow unaligned memory fixture emits byte-identical ARM64 native code at both revisions. The LLGo diagnostic also records a higher CPI point estimate for this workload with stable retired instructions. Host-entry placement and machine variation remain possible causes. The managed allocation differences in the two P3 adverse rows are below 0.001 allocation per request.

## Reproducible evidence

- [Machine-readable results, all controls and source contrasts](data/invariant-struct-loops-2026-11-20.json).
- [All 5,508 primary process records](data/invariant-struct-loops-2026-11-20-raw.jsonl.gz), SHA-256 `765fe89140ecd327edb6052f3931017361ef7e23456ac32392eef864c59ebe8b`.
- [Frozen plans, commands, hashes, analysis/export scripts, semantic checks and failure evidence](data/invariant-struct-loops-2026-11-20-protocol.tar.gz), SHA-256 `97b067b5e86f7e9ce6f4073e95923940a0c5890e177bf1e735ebefbd49497abe`.
- Per-request vectors unused by these metrics are projected to count/extrema/hash; no latency-tail claim. Public load records omit unrelated desktop command arguments. The separately archived screening observations are excluded from primary ratios.
