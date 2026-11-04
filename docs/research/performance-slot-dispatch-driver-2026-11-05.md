# Shared slot dispatcher with an outer native driver: 2026-11-05

**Held for integration; performance acceptance remains open.** Separating native transfer from the shared slot loop lowers most measured interpreter kernels and both compiled SHA workloads. Native/portable regressions, large controls and competitor/native gaps remain. This is a complete local source experiment, not controlled parity or final-target acceptance.

Measured V engine `fadcf7384a12efe3d98797322457efef82c8452b`; intermediate U engine `fb1c47c20142e86307344a619181c662ca734750`; production `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-equivalent to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). U is the [required backend-contract experiment](performance-native-backend-contract-2026-11-05.md), not a merged baseline. V changes two engine/test files against U. Go 1.27.3 and LLGo 3571 (`-O3 -lto=full`) remain pinned, as do producer/reference binaries, policies, clocks and fixtures.

## One instruction-semantic dispatcher

`Run` performs entry checks and selects portable execution or owns the native transfer loop. The existing noninlined packet bridge validates and adopts every frame/arena transfer. The native driver publishes its checked PC, then services exactly one already-charged instruction through `runSlots`; portable execution stays inside that same dispatcher. No backend interface, native packet or mutable native frame-view refresh remains in the opcode loop.

Sequential/edge continuations publish their next PC before returning to native code. Calls, returns, exceptions and stack-switching operations retain their existing complete frame transitions. Native boundaries never charge fuel twice or coalesce a second instruction. Portable coalescing, source/step polls, roots, quotas, owned storage, host lifetimes and interruption keep their previous contracts. The required backend contract and frame-transfer validator are byte-for-byte unchanged; all 86 opcode case labels are unchanged. No workload recognition, duplicate execution framework, ISA/profile change or padding/alignment sweep is introduced.

The new model test varies exact fuel through a finite loop and checks every native-entry count, coalesced local/constant boundaries, backedges, result and exhaustion outcome against portable execution. Existing hidden-frame, relocation, host/root/tuple, exception and stack-switch tests remain. The earlier held frame-scope rearrangement and capability-cache experiments remain separate evidence.

Ten affected packages pass both pinned compilers on macOS arm64: **108 actual conformance reports**, **90.71%–100% scoped Go statement coverage**, execute **96.64% (920/952)**. Race, cgocheck2 and no-cgo checks pass. Eight application configurations execute **64 checked calls / 48 expected traps**. The initial target check stopped on missing corpus paths in the new worktree; installing the existing pinned corpus links resolves that infrastructure failure. No guest mismatch or timing retry is hidden. These local checks do not replace final-head qualification on all four compiler/platform profiles.

## Frozen measurements and source changes

One original20 cohort uses three fresh process blocks, two warmups/seven individually interleaved samples, production/head same-binary LLGo A/A and contemporaneous production/U/V source observations. Independent verification checks **1224 VM/reference + 27 C observations, 408 alternate checks, 12 explicitly reused unsupported GC outcomes and 417 summaries**. Applications separately verify **126 VM/reference + 6 C observations, 42 alternate checks and 44 summaries**. Measurement/verification takes 462.22/5.23 seconds and 98.41/0.36 seconds respectively. Preparation, process startup, compilation, first call, warmed execution and cleanup remain distinct in the raw records.

Host load begins at 63.86/74.08/44.85; this shared host is not controlled hardware. Preserve every adverse row and A/A control. U has fresh normal source observations but no additional U A/A alias in this inherited protocol. No independent cohorts are pooled or ratios multiplied. The isolated call-control row has one exported call per sample and large timer/control variation; V has no warmed API-batch result, and U’s batch gains must not be attributed to V.

Changes below compare V against contemporaneous U; positive is slower. A/A columns compare the same head LLGo binary with its alias.

| Workload | Go native | LLGo native | Go interpreter | LLGo interpreter | LLGo head A/A native / interpreter |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | +5.77% | +2.94% | -5.57% | -7.41% | +0.24% / -1.15% |
| core/call-control | +15.29% | -32.40% | -15.63% | -32.67% | +23.66% / +8.08% |
| core/integer-loop | -1.60% | +1.14% | -2.98% | -7.97% | -1.25% / +2.39% |
| core/function-calls | +0.46% | -0.10% | -0.99% | -3.40% | +0.39% / +2.59% |
| core/memory-load-store | -3.90% | -0.75% | -4.62% | -2.06% | -0.15% / -0.99% |
| memory/narrow-unaligned | +0.73% | -5.20% | -6.51% | -2.71% | +8.62% / -0.63% |
| memory/dependent-address | -1.05% | +0.17% | -4.86% | -8.02% | -0.44% / +0.17% |
| memory/float64-unaligned | +4.82% | +5.31% | -3.71% | -4.00% | -1.48% / -0.03% |
| runtime/simd-memory-full | +1.55% | +3.37% | -1.88% | -6.13% | -0.64% / +3.75% |
| runtime/simd-memory-widen | -0.82% | -1.61% | -3.44% | -1.37% | +5.07% / -2.97% |
| runtime/simd-memory-lanes | +0.95% | +0.65% | -10.74% | -4.52% | +3.21% / -1.53% |
| gc/i31-roundtrip | +11.07% | -1.98% | +13.36% | -12.54% | -14.10% / +22.74% |
| gc/struct-access | +16.73% | -13.05% | -18.47% | -12.21% | +9.18% / +9.02% |
| gc/struct-allocation | +41.27% | +4.59% | -15.87% | -19.69% | -8.84% / +6.73% |
| gc/struct-indirect-access | -17.58% | -4.17% | -10.22% | -6.45% | +3.44% / -8.77% |
| threads/shared-memory | -5.24% | +2.10% | -0.47% | -9.77% | +8.90% / -0.43% |
| threads/atomic-rmw | -1.34% | +1.00% | -1.18% | +0.35% | +3.77% / -3.34% |
| threads/notify-empty | +31.46% | +1.99% | -5.15% | +31.27% | -4.39% / -30.74% |
| threads/wait-mismatch | +14.95% | +0.16% | -6.33% | -10.46% | +0.94% / +1.84% |
| threads/wait-timeout | -1.22% | -2.49% | +1.03% | -5.34% | -2.24% / +1.76% |

Go native struct allocation regresses **41.27%**, i31 **11.07%**, struct access **16.73%**, and float64 memory **4.82%**. LLGo native float64 memory regresses **5.31%**, SIMD full memory **3.37%**, allocation **4.59%** and other positive rows remain. LLGo interpreter notify-empty regresses **31.27%** with a **−30.74%** head A/A; i31 head A/A reaches **+22.74%**. Large controls weaken causal interpretation and do not waive these outcomes. Unchanged native instruction selection is not evidence that its complete source/binary event cannot regress.

## Every original20 competitive outcome

Ratios divide V LLGo duration by the named independent reference or V Go duration: below 1 is faster. Native C is the ordinary unchecked kernel and omits VM policies; it is not an equivalent rooted Store API. Unsupported/unavailable combinations stay explicit. Units, counts, absolute values, Wasmer/WasmEdge observations and all phases remain in the lossless record. These finite microfixtures do not establish whole-application or general GC performance.

| Workload | Native / Go | / Wasmtime | / Wasmer | / wazero compiler | / Native C |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.915× | 0.315× | 0.317× | 0.279× | unavailable |
| core/call-control | 0.640× | 2.113× | 0.939× | 4.446× | unavailable |
| core/integer-loop | 1.012× | 1.015× | 1.016× | 0.898× | 1.017× |
| core/function-calls | 0.986× | 0.851× | 0.979× | 0.547× | 0.999× |
| core/memory-load-store | 0.960× | 0.926× | 0.812× | 0.488× | 1.519× |
| memory/narrow-unaligned | 0.968× | 0.923× | 0.940× | 0.621× | 1.510× |
| memory/dependent-address | 0.998× | 0.893× | 0.870× | 0.579× | 1.291× |
| memory/float64-unaligned | 0.984× | 1.021× | 1.070× | 0.840× | 1.148× |
| runtime/simd-memory-full | 1.010× | 0.899× | 0.962× | 0.924× | 0.962× |
| runtime/simd-memory-widen | 0.993× | 0.987× | 0.993× | 0.826× | 0.985× |
| runtime/simd-memory-lanes | 0.986× | 0.833× | 0.843× | 0.786× | 1.096× |
| gc/i31-roundtrip | 0.665× | 1.201× | unsupported | unsupported | unavailable |
| gc/struct-access | 0.688× | 1.679× | unsupported | unsupported | unavailable |
| gc/struct-allocation | 0.820× | 0.333× | unsupported | unsupported | unavailable |
| gc/struct-indirect-access | 0.925× | 0.854× | unsupported | unsupported | unavailable |
| threads/shared-memory | 0.945× | 0.698× | 0.770× | 0.618× | unavailable |
| threads/atomic-rmw | 0.954× | 0.973× | 0.467× | 0.906× | unavailable |
| threads/notify-empty | 0.626× | 0.265× | 0.359× | 0.083× | unavailable |
| threads/wait-mismatch | 0.635× | 0.216× | 0.509× | 0.077× | unavailable |
| threads/wait-timeout | 0.518× | 0.044× | 0.011× | 0.007× | unavailable |

| Workload | Interpreter / Go | / wazero interpreter | / WasmEdge interpreter |
| --- | ---: | ---: | ---: |
| threads/private-memory | 0.811× | 0.584× | 0.912× |
| core/call-control | 1.007× | 2.000× | 0.739× |
| core/integer-loop | 0.735× | 0.330× | 0.555× |
| core/function-calls | 0.904× | 0.587× | 0.798× |
| core/memory-load-store | 0.797× | 0.421× | 0.650× |
| memory/narrow-unaligned | 0.813× | 0.430× | 0.697× |
| memory/dependent-address | 0.770× | 0.382× | 0.695× |
| memory/float64-unaligned | 0.801× | 0.460× | 0.736× |
| runtime/simd-memory-full | 0.769× | 0.442× | 0.840× |
| runtime/simd-memory-widen | 0.883× | 0.433× | 0.980× |
| runtime/simd-memory-lanes | 0.820× | 0.452× | 0.921× |
| gc/i31-roundtrip | 0.582× | unsupported | 0.763× |
| gc/struct-access | 0.895× | unsupported | 0.711× |
| gc/struct-allocation | 1.282× | unsupported | 1.115× |
| gc/struct-indirect-access | 0.823× | unsupported | 0.777× |
| threads/shared-memory | 0.458× | 0.580× | 0.925× |
| threads/atomic-rmw | 0.637× | 0.755× | 1.170× |
| threads/notify-empty | 1.165× | 1.095× | 0.746× |
| threads/wait-mismatch | 0.603× | 0.593× | 0.943× |
| threads/wait-timeout | 0.598× | 0.119× | 0.401× |

The 12 reused GC unsupported outcomes apply to the exact unchanged wazero/Wasmer reference binaries; they are not zero durations or V feature failures. Wasmtime provides the supported GC comparison. Short isolated calls, tiny finite GC loops and atomic/wait microfixtures require their own policy/count interpretation; no best-row aggregate substitutes for open workloads.

## Compiled SHA applications

Warm ns/message block at 5000 blocks/call includes message generation, chaining and padding. These C/Rust Core modules implement one algorithm, have no imports/start, and do not complete a representative application/WASI-world portfolio.

| Guest | Go native | LLGo native | Wasmtime | Wasmer | wazero compiler | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | 262.183 | 261.175 | 243.725 | 260.725 | 380.225 | 213.717 |
| sha256-rust | 268.975 | 263.858 | 232.117 | 241.083 | 410.192 | 213.558 |

| Guest | Go native vs U | LLGo native vs U | Go interpreter vs U | LLGo interpreter vs U | LLGo head A/A native / interpreter |
| --- | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | -1.25% | -1.64% | -4.88% | -3.87% | -0.49% / -0.62% |
| sha256-rust | +0.31% | -0.03% | -6.99% | -4.13% | -1.08% / +0.77% |

V LLGo native remains **7.16% / 13.67% slower than Wasmtime**, **0.17% / 9.45% slower than Wasmer**, and **22.21% / 23.55% slower than native C** for C/Rust SHA. It is faster than wazero compiler in this workload. LLGo interpreters take **22283.617 / 21409.008 ns/block**, versus Go **28072.792 / 26418.358**, wazero interpreter **52688.200 / 51539.567**, and WasmEdge **34158.917 / 30604.150**. Against production, interpreter durations fall Go **3.78% / 7.13%**, LLGo **1.13% / 2.29%**. Most of the native production-to-V gain is inherited from earlier work, not caused by this dispatcher split.

## Assembly, integration and next work

Exact release disassembly confirms separate LLGo `Run` (109 instruction lines) and `runSlots` (2722), versus U `Run` (2759). The slot body contains neither `runBackend` nor `interfaceImplementation`. Its prologue reserves 1536 bytes plus 96 bytes of saved registers, and native effect boundaries now enter that body repeatedly. Go has separate 84/2901-line bodies and a 976-byte slot stack frame. Static instruction lines and stack frames are diagnostic; they do not measure dynamic cycles or quantitatively explain any regression. This evidence directs the next architectural step: extract one shared runtime-effect handler so native exits do not enter the complete portable opcode-loop frame.

Keep V as a held experiment, with all source and results preserved. Global live-range splitting/linear scan, public admission/marshalling/context costs, GC/indirect access, startup, controlled parity and broader application/WASI-world acceptance remain open. LLGo #2983 stays with its contributor; [#2993](https://github.com/xgo-dev/llgo/issues/2993) remains the separate interface-method allocation report.

No new contribution PR, hosted matrix or merge is performed. External PR #2991 remains open; owner policy prohibits merging it upstream. Main stays at `4ce79899a1f3ed45163afa622cc408cc6a79a599`. The preceding O matrix qualifies O only. Final integration still needs the exact final-head Go/LLGo × Linux amd64/macOS arm64 matrix, all 31 packages/profile, 30 coverage gates, complete review and performance acceptance. No milestone or full goal is marked complete.

[Lossless evidence](evidence/performance-slot-dispatch-driver-2026-11-05.json.xz) contains the fixed plans, identities, source patches, semantic/coverage/ownership/application qualification, initial development failure record, all complete measured samples/controls, independent verifiers and named release disassembly.
