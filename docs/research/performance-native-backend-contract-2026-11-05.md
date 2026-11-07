# Required native backend contract: 2026-11-05

**Retained architecture foundation; overall acceptance remains open.** LLGo native public-call batch durations fall **20.00%** with a background parent and **13.31%** with a live cancellable parent against O. Complete source/competitor results still contain regressions and deficits; this is not merge-ready or a controlled parity claim.

Measured engine `fb1c47c20142e86307344a619181c662ca734750`; retained O `d472a76ff9a8aebfdacd6ba74eeccc962d404e59`; production `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-equivalent to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). Five engine/test files change against O. The source removes the [held capability-cache experiment](performance-backend-capabilities-2026-11-05.md) from the final engine tree; its evidence and regressions remain unchanged. Go 1.27.3, LLGo 3571 (`-O3 -lto=full`), guest/reference binaries, flags, policies and clocks remain pinned.

## One complete backend contract

The actual native backend already supports explicit frame packets and bounded interruption. Require `Resize`, `RunFrames`, `Interrupt` and `Close` together. Remove optional nonempty interface assertions, both capability-cache fields, the old single-frame bridge fallback and its redundant interface argument. The shared frame packet validator, runtime effects, root/canonical-word publication, quotas, fuel, cancellation registration/retirement and owned storage remain. A nil backend still selects portable dispatch. The relocating test model now implements the same contract. The low-level native single-frame convenience method remains available for its separate direct callers; it is no longer a competing execution-driver protocol.

The `Execution` source declaration is byte-for-byte identical to O. This is source evidence, not a binary-layout or cycle attribution. No padding/alignment/ISA/flag sweep is performed. The native-only bridge remains explicitly noninlined. All original checked PCs, frame-transfer validation and shared runtime semantics remain in force. **Global live-range allocation remains unfinished.**

[LLGo #2993](https://github.com/xgo-dev/llgo/issues/2993) independently reports runtime method-table allocation for nonempty interface assertions: five fresh processes/compiler, 100000 checked assertions, ~9.6 MB/33.6 MB reported TotalAlloc for two/eight concrete methods under the fixed LLGo pin, versus 0–80 B under Go. Already converted interface invocation and concrete-pointer assertions report zero. Current upstream source still contains the copy implementation; current upstream was not rebuilt for those execution results. This differs from #2983, which remains with its contributor. The candidate changes no pinned compiler/runtime. A ten-million-call LLGo/native diagnostic sample contains no repeated `interfaceImplementation` path; stack samples do not quantify throughput or CPU-time shares.

Ten complete affected packages pass both pinned host compilers on macOS arm64: **108 actual conformance reports**, **90.71%–100% scoped Go statement coverage**, execute **96.62% (916/948)**. Actual native/portable corpora and behavioral tests cover changing invocation policies, nested/relocated frames, canceled entry, interruption, lifetime/cleanup and reuse. Race, cgocheck2 and no-cgo checks pass. Eight application configurations complete **64 checksum calls / 48 expected traps**. No preceding O CI result is substituted for final-head target qualification.

## Separate warmed-call batches

Use the established 10000-call protocol: five fresh process blocks, two warmups/seven individually interleaved samples, **270 observations, 54 alternate checks, 54 summaries**. Public API, live-parent API, session, internal dispatch and raw native-entry diagnostics retain their separate contracts. All results are checked; no mode substitutes for another and no overhead is subtracted. First-call phase in this protocol means the first **batch**, not one exported invocation. The original isolated-call portfolio remains separate below.

There are 46 unique commands; eight same-binary LLGo A/A identities are aliases. One-sample qualification initially failed the validator’s fixed 7/2 sample-count assumptions; a subsequent check exposed seven fixed C-reference calibration pairs. Preserve both successful guest probe records and the stopped preflight; reuse 44 validated observations and qualify the two C references with their formal 7/2 protocol. All preparation issues precede timing; no timed cohort was retried.

Warm ns/exported call, including the declared host-loop/check work:

| Backend/compiler | Background O → U | Change | Live parent O → U | Change |
| --- | ---: | ---: | ---: | ---: |
| native / go | 353.021 → 359.992 | +1.97% | 650.362 → 647.888 | -0.38% |
| native / llgo | 1023.837 → 819.071 | -20.00% | 1887.400 → 1636.171 | -13.31% |
| interpreter / go | 279.429 → 267.288 | -4.35% | 313.279 → 314.429 | +0.37% |
| interpreter / llgo | 722.558 → 705.612 | -2.35% | 847.221 → 860.129 | +1.52% |

Native background LLGo A/A is O +6.14% / U −1.43%; live-parent O +1.22% / U −2.27%. Interpreter controls are larger and remain explicit. These are shared-host descriptive results.

| Independent API/policy | Background | Live parent |
| --- | ---: | ---: |
| wazero compiler, termination default | 15.425 | 17.700 |
| wazero compiler, termination enabled | 344.321 | 376.887 |
| wazero interpreter, termination default | 79.588 | 80.396 |
| wazero interpreter, termination enabled | 414.137 | 519.808 |
| Wasmtime default | 69.804 | — |
| Wasmer default | 165.746 | — |

Policy/lifecycle obligations differ and stay recorded. LLGo native public API remains **2.28× Go**, **2.38× termination-enabled wazero** with a background parent, **4.34× enabled wazero** with a live parent. Default wazero and default Wasmtime are faster still. These gaps are not excused by faster component diagnostics. Native C has no equivalent rooted Store API in this protocol; it is unavailable, never zero.

| Own diagnostic mode | Go O → U | LLGo O → U |
| --- | ---: | ---: |
| native session | 257.533 → 260.979 | 670.675 → 419.929 |
| native dispatch | 152.246 → 158.233 | 401.192 → 120.438 |
| native entry | 93.542 → 96.862 | 75.729 → 75.167 |

LLGo native dispatch falls from 401.192 to 120.438 ns while raw native entry changes from 75.729 to 75.167 ns. The raw arithmetic entry omits the public VM/function lifecycle and guest dispatcher. Public interpreter remains 705.612 ns under LLGo versus 267.288 ns Go, despite internal dispatch at 67.542 versus 72.992 ns. Admission, marshalling, context and cleanup need further diagnosis; do not attribute all remaining cost to Cgo or guest instructions.

## Original 20 and compiled applications

A separate three-source cohort retains three fresh blocks, two warmups/seven samples and production/head same-binary LLGo A/A. Original 20 independently verifies **1224 VM/reference + 27 C observations, 408 alternate checks, 12 explicitly reused unsupported GC outcomes and 417 summaries**. Applications independently verify **126 VM/reference + 6 C observations, 42 alternate checks and 44 summaries**. All observations, controls and adverse rows remain. Do not pool with batches or prior experiments.

Warm ns/message block at 5000 blocks/call includes generation, chaining and padding. The two C/Rust Core modules implement one algorithm and do not establish a complete application/WASI-world portfolio.

| Guest | LLGo production → O → U | Go U | Wasmtime | Wasmer | wazero compiler | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | 798.025 → 238.883 → 239.042 | 244.892 | 222.292 | 227.383 | 347.075 | 188.200 |
| sha256-rust | 845.683 → 263.292 → 260.642 | 265.400 | 241.492 | 234.633 | 412.075 | 209.233 |

| Guest | Go native vs O | LLGo native vs O | Go interpreter vs O | LLGo interpreter vs O | LLGo native/interpreter head A/A |
| --- | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | +0.66% | +0.07% | +1.03% | -2.31% | -2.46% / -0.93% |
| sha256-rust | -0.32% | -1.01% | -0.85% | -3.23% | -0.17% / +1.03% |

The held cache-placement cohort’s Go SHA regressions of 9.28%/7.05% remain recorded. U’s current source-relative Go results are +1.03%/−0.85%; the separate cohorts are not a quantitative attribution of the old regression to one field offset. Against production, C SHA interpreter still regresses Go +4.88% / LLGo +4.74%. All preparation and startup phases remain separate from warmed execution.

## Every original 20 source outcome

Positive changes mean slower; full absolute compiler/reference values and every phase remain machine-readable.

| Workload | Go native vs O | LLGo native vs O | Go interpreter vs O | LLGo interpreter vs O | LLGo native/interpreter head A/A |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | +2.41% | +0.29% | -3.78% | +0.65% | +3.70% / -3.32% |
| core/call-control | -5.57% | -30.09% | +21.90% | -12.95% | -0.70% / -21.62% |
| core/integer-loop | -0.86% | -0.12% | -4.04% | -5.32% | +0.03% / +2.72% |
| core/function-calls | -4.73% | -1.65% | -2.36% | -4.87% | +0.47% / +0.12% |
| core/memory-load-store | -2.24% | -2.12% | +4.25% | +0.32% | +2.74% / +0.06% |
| memory/narrow-unaligned | -0.90% | +0.04% | +0.03% | -3.74% | -4.13% / +3.54% |
| memory/dependent-address | -0.41% | +0.36% | +0.58% | -2.79% | +0.38% / +0.20% |
| memory/float64-unaligned | -0.89% | -1.70% | -0.44% | -0.78% | +1.30% / +0.29% |
| runtime/simd-memory-full | +0.15% | -0.11% | +0.13% | -0.99% | +0.25% / -0.26% |
| runtime/simd-memory-widen | -0.16% | -1.02% | -1.55% | -0.44% | -0.01% / +0.31% |
| runtime/simd-memory-lanes | +5.15% | +2.01% | +0.31% | -6.60% | +1.32% / +4.12% |
| gc/i31-roundtrip | -0.81% | -10.00% | +4.95% | +1.19% | -1.37% / -3.46% |
| gc/struct-access | -1.75% | -5.99% | -1.40% | +0.50% | +5.72% / -2.20% |
| gc/struct-allocation | +4.99% | +0.42% | -1.68% | +1.80% | -3.29% / -2.03% |
| gc/struct-indirect-access | -13.65% | +0.55% | -1.56% | -2.45% | -3.68% / -0.91% |
| threads/shared-memory | -0.08% | -7.28% | -1.38% | +4.61% | -0.18% / -0.87% |
| threads/atomic-rmw | -1.25% | -2.40% | +0.19% | -2.89% | +2.31% / +1.44% |
| threads/notify-empty | +1.14% | -11.66% | -8.19% | -5.08% | +12.00% / +1.70% |
| threads/wait-mismatch | +0.42% | -16.11% | +4.98% | -2.17% | +10.73% / -1.80% |
| threads/wait-timeout | -12.50% | -6.72% | -2.93% | +1.57% | -3.89% / -2.02% |

Go native SIMD lanes +5.15%, struct allocation +4.99%, interpreter memory-load/store +4.25%, i31 +4.95%, wait mismatch +4.98%, LLGo shared-memory interpreter +4.61% and every other adverse row remain open. Tiny isolated call/control and wait rows have large A/A variation. No unchanged source region, raw-entry win or API batching result waives a regression.

## Integration and next work

Keep the compact required protocol as a foundation; overall gates remain open and no contribution PR or merge is performed. The sole external PR #2991 remains open, and external upstream merges are prohibited by owner policy. Main stays at `4ce79899a1f3ed45163afa622cc408cc6a79a599`. Full final-head Go/LLGo × Linux amd64/macOS arm64 semantics, 31 packages/profile, 30 coverage gates and review remain required before integration; a held source does not need a redundant hosted matrix.

Prioritize separating native transfer from the portable opcode hot loop through one shared runtime boundary handler. Preserve complete fuel/source/root/frame/cancellation semantics and retain the earlier held dispatch-rearrangement results. Public-call ownership/context costs, GC/indirect access, startup, global live-range allocation, controlled competitor/native acceptance and broader application/WASI-world coverage also remain open. No milestone or full goal is marked complete.

[Lossless evidence](evidence/performance-native-backend-contract-2026-11-05.json.xz) includes final source/tool identities, local semantic/coverage/ownership qualification, both batch-preparation counter failures, all complete timed observations/controls, independent verifiers, structural source checks and the ten-million-call diagnostic. Prior cache failure and LLGo #2993 reproducer remain in the linked preceding evidence.
