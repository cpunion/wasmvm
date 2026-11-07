# Shared runtime effects outside the portable loop: 2026-11-05

**Held; performance acceptance remains open.** The implementation removes the full portable opcode frame from native effects and sharply reduces the portable loop frame. Actual measurements still contain large GC/SIMD/atomic and public-call regressions. Severe same-binary variation makes several competitor/native medians inconclusive. Smaller static frames alone do not establish faster execution.

W engine `6b5006758ee6c43f70f583bef31b0117d22820d3`; intermediate [V](performance-slot-dispatch-driver-2026-11-05.md) `fadcf7384a12efe3d98797322457efef82c8452b`; production `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-equivalent to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). Nine files change against V: six production files and three tests. Go 1.27.3, LLGo 3571 (`-O3 -lto=full`), producer/reference binaries, policies, fixtures and corrected clocks stay pinned.

## Runtime contracts and qualification

Native execution validates its returned boundary before shared runtime handling; it cannot call the portable opcode loop or fall back on ordinary scalar instructions. The portable loop no longer has a single-boundary mode. Calls/returns, exceptions, stack transitions, GC and SIMD retain shared implementations. Portable scalar/local/CFG operations remain direct; memory/table/global routing uses the existing canonical state and memory contracts. Source PCs publish before effects and advance or transfer before either driver resumes. Generated native fuel, portable source charges, step/backedge polling, canonical slots, roots, limits, host suspension, interruption and owned storage retain their contracts. The Execution declaration and native packet bridge/validator are unchanged; no ISA/profile/layout sweep occurs.

The models now execute their bounded ordinary instruction subset before returning legitimate effects. An initial model callback charged the outer packet while a copied child advanced, producing a fuel mismatch; passing the current packet fixes it. A subsequent check detected only one hidden batch where the existing tuple/root test requires at least two; continuing the model call chain preserves that original assertion. Both failed development records remain. The new test rejects a scalar native exit with `ErrInvalid`, empty results and cleaned frames. No assertion is weakened and no timed cohort is retried.

Ten complete affected packages pass pinned Go and LLGo on macOS arm64: **108 actual conformance reports**, scoped Go coverage **90.71%–100%**, execute **95.04% (1034/1088)**. Race, cgocheck2 and no-cgo checks pass. Eight compiled-application configurations pass **64 checksum calls / 48 expected traps**. Actual native/portable Core and proposal execution is distinct from test-model success. Local qualification does not substitute for the final four-platform/compiler matrix.

## Static code and measured boundaries

Release disassembly confirms LLGo portable `runSlots` shrinks from 2722 to 1413 instruction lines and its frame from **1632 to 160 bytes**. The new native runtime handler has a **496-byte** frame; GC and return helpers use **448 / 112 bytes**. It restores its frame and tail-branches to GC/SIMD/call/throw/return helpers. Tail transfers are not nested stack-frame sums, but repeated prologue/epilogue work still costs instructions. Go portable frames shrink **976 to 224 bytes**; its runtime/GC/return frames are **560 / 256 / 160 bytes**. Static instruction lines and frame reservations are diagnostics, not cycle attribution or whole-program/guest-kernel byte counts.

The separate established batch protocol uses 10000 calls/batch, five fresh process blocks, two warmups/seven interleaved samples, **270 observations / 54 alternate checks / 54 summaries**. All 46 unique command configurations qualify before timing. Public, live-parent, session, dispatch and raw native-entry modes remain separate; no overhead is subtracted. First call means the first batch, not one exported call. Warm ns/call:

| Backend/compiler | Background V → W | Change | Live parent V → W | Change |
| --- | ---: | ---: | ---: | ---: |
| native / go | 595.542 → 571.138 | -4.10% | 1186.667 → 1089.200 | -8.21% |
| native / llgo | 1106.987 → 1103.025 | -0.36% | 2279.133 → 2184.167 | -4.17% |
| interpreter / go | 460.371 → 481.913 | +4.68% | 611.904 → 517.433 | -15.44% |
| interpreter / llgo | 987.263 → 1079.000 | +9.29% | 1200.321 → 1186.804 | -1.13% |

| Diagnostic | Go V → W | LLGo V → W |
| --- | ---: | ---: |
| Native session | 365.021 → 380.158 | 573.158 → 551.825 |
| Native dispatch | 201.396 → 194.596 | 159.379 → 129.221 |
| Raw native entry | 118.883 → 134.979 | 89.442 → 86.950 |
| Interpreter session | 208.637 → 228.900 | 461.212 → 533.829 |
| Interpreter dispatch | 102.967 → 114.046 | 95.442 → 78.175 |

| Independent API/policy | Background | Live parent |
| --- | ---: | ---: |
| wazero compiler, termination default | 22.183 | 21.425 |
| wazero compiler, termination enabled | 584.408 | 606.442 |
| wazero interpreter, termination default | 137.417 | 100.233 |
| wazero interpreter, termination enabled | 617.004 | 645.163 |
| wasmtime-cranelift default | 134.975 | unavailable |
| wasmer-cranelift default | 229.504 | unavailable |

LLGo native public API remains **1.93× Go**, **1.89× termination-enabled wazero** in background calls and **3.60×** with live parents. Default policies differ and remain explicit. Native C has no equivalent rooted Store API here. Lower internal dispatch does not waive complete API/session regressions.

LLGo native public A/A changes V +2.20% / W −1.17% in background, V −1.68% / W +5.50% with live parents. Interpreter background controls are V +10.89% / W −6.11%, live V +5.67% / W +8.21%. The public native change of −0.36% is small; do not claim a meaningful API win from the component result.

## Original20 source outcomes

A distinct original20 cohort retains three fresh blocks, two warmups/seven individually interleaved samples, production/head LLGo A/A and fresh production/V/W observations. Independent checks verify **1224 VM/reference + 27 C observations, 408 alternate checks, 12 reused unsupported GC outcomes and 417 summaries**. The application cohort separately verifies **126 VM/reference + 6 C observations, 42 alternate checks and 44 summaries**. Measurement/verification takes 542.77/6.63 seconds and 117.33/0.57 seconds; batch measurement/verification takes 55.97/1.40 seconds. All phases, outcomes and controls remain. No independent cohort ratios are multiplied or pooled.

Preflight load is **73.69 / 62.38 / 43.89**. V has contemporaneous source observations but no extra V A/A alias in this inherited portfolio. Results are descriptive shared-host data. Positive source changes are slower; A/A compares the identical head LLGo binary:

| Workload | Go native vs V | LLGo native vs V | Go interpreter vs V | LLGo interpreter vs V | LLGo head A/A native / interpreter |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | -37.54% | +6.39% | -6.39% | +4.05% | -18.23% / -1.64% |
| core/call-control | +1.53% | -18.97% | +45.45% | -0.47% | -0.52% / -9.66% |
| core/integer-loop | +1.78% | +0.00% | -10.22% | -2.00% | -3.29% / +0.75% |
| core/function-calls | +2.53% | +12.31% | +11.62% | +2.91% | -0.75% / -1.24% |
| core/memory-load-store | +9.11% | -26.89% | -0.40% | -8.66% | -15.46% / +3.83% |
| memory/narrow-unaligned | +15.66% | -11.93% | +1.02% | -2.55% | -24.26% / +2.17% |
| memory/dependent-address | +0.22% | -3.39% | -6.90% | -1.90% | +2.48% / +7.85% |
| memory/float64-unaligned | -4.45% | -1.11% | +1.01% | -3.67% | -1.14% / +0.37% |
| runtime/simd-memory-full | +6.44% | -1.83% | -3.06% | +4.41% | +2.18% / -0.23% |
| runtime/simd-memory-widen | -1.14% | -2.32% | +6.24% | +10.58% | +7.59% / +0.19% |
| runtime/simd-memory-lanes | -1.22% | -5.04% | +8.80% | -1.31% | +4.08% / -2.32% |
| gc/i31-roundtrip | -0.88% | -5.11% | -21.35% | +4.55% | +0.00% / +35.02% |
| gc/struct-access | -1.12% | -2.47% | +3.14% | +21.68% | -9.69% / -23.71% |
| gc/struct-allocation | +57.04% | +21.33% | +25.43% | -8.00% | -4.92% / +10.94% |
| gc/struct-indirect-access | -14.01% | -3.01% | -0.85% | +2.41% | -1.86% / -14.59% |
| threads/shared-memory | +2.24% | -27.12% | -1.50% | -2.22% | -8.84% / -1.78% |
| threads/atomic-rmw | -0.96% | +3.36% | +8.65% | +11.28% | +5.26% / -0.32% |
| threads/notify-empty | +4.04% | +0.82% | -8.03% | +50.54% | +0.50% / +0.30% |
| threads/wait-mismatch | -1.88% | +15.05% | +6.31% | -0.48% | -0.93% / -1.51% |
| threads/wait-timeout | -5.69% | -5.82% | +16.41% | -7.74% | +2.38% / -17.67% |

Native allocation regresses **57.04% Go / 21.33% LLGo**. LLGo interpreter SIMD widen regresses **10.58%** (head A/A +0.19%), atomics **11.28%** (−0.32%), notify-empty **50.54%** (+0.30%). Other positive rows remain. Some GC controls reach +35.02%/−23.71%; large controls weaken causality and do not excuse adverse results. The additional helper dispatch/prologue and operand-view reconstruction require architectural review; static frame shrinkage is insufficient evidence of lower cost.

## Every original20 reference ratio

Ratios divide W LLGo duration by the named reference or W Go: below 1 is faster. Native C is an unchecked ordinary kernel, not an equivalent Store/API. Keep finite microfixture, policy and unit differences; unsupported reference cells are not zero. These medians cannot establish controlled parity.

| Workload | Native / Go | / Wasmtime | / Wasmer | / wazero compiler | / Native C |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.808× | 0.428× | 0.651× | 0.380× | unavailable |
| core/call-control | 0.565× | 1.725× | 0.764× | 4.178× | unavailable |
| core/integer-loop | 0.982× | 0.984× | 0.999× | 0.817× | 0.988× |
| core/function-calls | 0.993× | 0.809× | 1.000× | 0.495× | 1.040× |
| core/memory-load-store | 1.069× | 1.086× | 0.589× | 0.500× | 0.952× |
| memory/narrow-unaligned | 0.901× | 1.058× | 0.970× | 0.667× | 2.022× |
| memory/dependent-address | 0.976× | 0.892× | 0.883× | 0.571× | 1.308× |
| memory/float64-unaligned | 1.015× | 1.097× | 1.126× | 0.762× | 1.158× |
| runtime/simd-memory-full | 0.937× | 0.929× | 0.966× | 0.858× | 0.936× |
| runtime/simd-memory-widen | 0.964× | 0.939× | 0.966× | 0.920× | 0.994× |
| runtime/simd-memory-lanes | 0.993× | 0.835× | 0.828× | 0.784× | 1.049× |
| gc/i31-roundtrip | 0.865× | 1.098× | unsupported | unsupported | unavailable |
| gc/struct-access | 0.930× | 1.912× | unsupported | unsupported | unavailable |
| gc/struct-allocation | 0.946× | 0.453× | unsupported | unsupported | unavailable |
| gc/struct-indirect-access | 0.991× | 2.041× | unsupported | unsupported | unavailable |
| threads/shared-memory | 1.274× | 1.259× | 0.996× | 0.726× | unavailable |
| threads/atomic-rmw | 1.055× | 1.010× | 0.410× | 0.939× | unavailable |
| threads/notify-empty | 0.875× | 0.232× | 0.262× | 0.083× | unavailable |
| threads/wait-mismatch | 0.862× | 0.262× | 0.760× | 0.080× | unavailable |
| threads/wait-timeout | 0.634× | 0.044× | 0.010× | 0.004× | unavailable |

| Workload | Interpreter / Go | / wazero interpreter | / WasmEdge interpreter |
| --- | ---: | ---: | ---: |
| threads/private-memory | 0.888× | 0.608× | 0.922× |
| core/call-control | 0.809× | 2.275× | 0.739× |
| core/integer-loop | 0.750× | 0.338× | 0.561× |
| core/function-calls | 0.849× | 0.478× | 0.769× |
| core/memory-load-store | 0.805× | 0.407× | 0.651× |
| memory/narrow-unaligned | 0.779× | 0.426× | 0.689× |
| memory/dependent-address | 0.875× | 0.411× | 0.693× |
| memory/float64-unaligned | 0.776× | 0.469× | 0.726× |
| runtime/simd-memory-full | 0.850× | 0.492× | 0.927× |
| runtime/simd-memory-widen | 0.847× | 0.468× | 1.042× |
| runtime/simd-memory-lanes | 0.904× | 0.523× | 1.109× |
| gc/i31-roundtrip | 0.979× | unsupported | 1.033× |
| gc/struct-access | 1.002× | unsupported | 0.724× |
| gc/struct-allocation | 1.036× | unsupported | 1.086× |
| gc/struct-indirect-access | 0.998× | unsupported | 0.948× |
| threads/shared-memory | 0.464× | 0.576× | 0.941× |
| threads/atomic-rmw | 0.656× | 0.863× | 1.276× |
| threads/notify-empty | 1.050× | 0.870× | 0.734× |
| threads/wait-mismatch | 0.647× | 0.577× | 0.732× |
| threads/wait-timeout | 0.796× | 0.147× | 0.561× |

The 12 unsupported GC outcomes reuse the same unchanged reference binaries and exact fixture hashes. Wasmtime and WasmEdge provide supported GC references; allocations and waits in these finite fixtures do not establish general collector or application performance.

## Compiled SHA modules

Warm ns/message block at 5000 blocks/call includes generation, chaining and padding. C/Rust modules cover one algorithm without imports/start; broader application/WASI-world acceptance is still required.

| Guest | Go native | LLGo native | Wasmtime | Wasmer | wazero compiler | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | 358.900 | 309.475 | 371.308 | 316.892 | 532.450 | 228.958 |
| sha256-rust | 424.575 | 422.917 | 340.608 | 421.825 | 469.250 | 323.133 |

| Guest | Go native vs V | LLGo native vs V | Go interpreter vs V | LLGo interpreter vs V | LLGo head A/A native / interpreter |
| --- | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | +0.40% | +3.22% | -8.07% | -7.68% | +54.06% / +3.68% |
| sha256-rust | +24.15% | +53.51% | -4.18% | -1.62% | -21.39% / -2.22% |

C SHA head LLGo native A/A is **+54.06%**; Rust native A/A is **−21.39%**, while its W/V native median regresses **53.51%**. Do not claim the C median’s apparent Wasmtime advantage as a win or attribute the entire Rust regression to this source change. All these outcomes remain. Interpreter C changes are Go −8.07% / LLGo −7.68%; Rust Go −4.18% / LLGo −1.62%, with explicit controls above. Both compiler and native-reference gaps need adequately controlled measurements.

## Next work and integration

Keep W as a held experiment. Route native exits directly to shared leaf handlers where possible, and keep frequent portable primitive routing in its hot loop while sharing canonical semantic primitives. This targets the extra boundary/prologue work without restoring the full native-to-portable dispatch path or creating another VM. Public admission/marshalling/context costs, global live-range splitting/linear scan, startup, controlled competitor/native acceptance and a representative application/WASI-world portfolio remain open.

No new PR, hosted matrix or merge is performed. External PR #2991 remains open; owner policy prohibits upstream merge. LLGo #2983 stays with its contributor and [#2993](https://github.com/xgo-dev/llgo/issues/2993) remains the separate interface-method allocation issue. Main remains `4ce79899a1f3ed45163afa622cc408cc6a79a599`. Before eventual integration, require the exact final-head Go/LLGo × Linux amd64/macOS arm64 matrix, 31 packages/profile, 30 coverage gates, review and performance acceptance. Earlier O CI qualifies O only. No milestone or goal is complete.

[Lossless evidence](evidence/performance-runtime-effects-2026-11-05.json.xz) includes source/pins, extraction and model-failure/correction records, complete local qualification, all fixed plans/observations/controls, independent verifiers, exact release disassembly and source-contract checks.
