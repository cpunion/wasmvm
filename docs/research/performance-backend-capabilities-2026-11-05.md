# Backend capability discovery: initial placement, 2026-11-05

**Held; performance acceptance fails.** Discovering an immutable backend method set once removes repeated LLGo metadata-copy work from sampled Run paths, but Go SHA interpreter duration regresses **9.28%/7.05%** against retained candidate O. LLGo atomic-RMW interpreter duration also increases **6.71%** against O. Keep all adverse outcomes and do not merge this source as a demonstrated throughput win.

Measured engine `e2fd3406cf51498c9d59cd560d4352a03f52fcd0`; O `d472a76ff9a8aebfdacd6ba74eeccc962d404e59`; production `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-equivalent to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). Three engine/test files change against O. This source excludes held native-pointer/rematerialization experiments. Go 1.27.3 and LLGo 3571 (`-O3 -lto=full`), guest/reference binaries, flags, clocks, policies and alignment stay fixed.

## Diagnosis, mechanism and qualification

A retained-O main-thread stack sample repeatedly reaches `interfaceImplementation`/`UncommonType.Methods` through `Execution.Run`’s nonempty backend assertion. The runtime copies all concrete method records on each check. [LLGo #2993](https://github.com/xgo-dev/llgo/issues/2993) contains a complete tested reproducer and five fresh processes/compiler: 100000 successful assertions report ~9.6 MB for two concrete methods and ~33.6 MB for eight under LLGo, versus 0–80 B in Go; failed one-method assertions report ~4.8 MB versus 144 B. Already converted interface calls, concrete-pointer assertions and base calls report zero for both. These are approximate TotalAlloc deltas, not allocation-event counts or throughput. Pin-specific execution and current upstream source inspection remain distinct; #2983 is unrelated and stays with its contributor.

The execution workspace now discovers FrameBackend and Interrupt method sets at creation, since Rebind keeps backend identity. Run uses the converted capability and Interrupt keeps the same receiver. Frame publication, roots, fuel, limits, cancellation registration/retirement, context lifetime, close ownership and the common runtime are unchanged. The first implementation places two interface fields immediately after config, displacing later hot fields. This layout change is a possible explanation for regressions, not an established attribution.

Ten complete affected packages pass both pinned host compilers on macOS arm64, with **108 actual conformance reports** and **90.71%–100% scoped Go statement coverage**; execute is **96.65% (924/956)**. Tests exercise changing invocation policies across Quiesce/Rebind, nested relocated frames, cancellation before entry, interrupt delivery and nil/plain/optional-capability backends. Full actual native/portable corpora remain green. Race, cgocheck2 and no-cgo checks pass. Eight application configurations complete **64 checksum calls / 48 expected traps**. Final-head four-profile CI/review remain integration requirements; preceding O CI does not qualify this head.

Matched before/after diagnostic protocols each check ten million public calls per compiler/backend. The repeated assertion allocation path appears in O LLGo native stacks and is absent after caching. Stack samples include waiting threads and unequal observation counts, so they cannot assign a quantitative CPU-time share or replace throughput measurements. No pinned compiler/runtime patch is used.

## Fixed complete cohort

Three fresh process blocks, two warmups/seven individually interleaved samples, production/head same-binary LLGo A/A controls. Original20: **1224 VM/reference +27 C observations,408 alternate checks,12 explicitly reused unsupported GC outcomes,417 summaries**. Applications: **126 VM/reference +6 C observations,42 alternate checks,44 summaries**. Independent verifiers check identities, checksums, medians and all controls. No timing retry or cross-cohort pooling.

The isolated exported-call control is LLGo native **3583→3167 ns (−11.61%)** against O, with head A/A **+14.46%**; interpreter **2292→2625 ns (+14.53%)**, head A/A **+7.96%**. Tiny isolated-call variability prevents a controlled win claim. Existing batch-call protocols are a separate measurement dimension, never a replacement for these outcomes.

Warm ns/message block at 5000 blocks/call includes generation, chaining and padding. C/Rust are two Core compute modules for one algorithm, not a full application/WASI-world portfolio.

| Guest | LLGo O → head | Go head | Wasmtime | Wasmer | wazero compiler | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | 235.392 → 238.658 | 237.967 | 216.158 | 224.250 | 346.317 | 188.142 |
| sha256-rust | 229.142 → 227.050 | 236.158 | 206.983 | 199.417 | 348.525 | 182.683 |

| Guest | Go native vs O | LLGo native vs O | Go interpreter vs O | LLGo interpreter vs O | LLGo native/interpreter head A/A |
| --- | ---: | ---: | ---: | ---: | ---: |
| sha256-clang | +2.22% | +1.39% | +9.28% | +4.38% | -1.01% / -5.29% |
| sha256-rust | -0.54% | -0.91% | +7.05% | -0.65% | -1.55% / +0.69% |

## Every original20 source outcome

Positive changes mean slower. Exact absolute runtime/reference rates and all Go/LLGo phase/control records remain lossless in evidence.

| Workload | Go native vs O | LLGo native vs O | Go interpreter vs O | LLGo interpreter vs O | LLGo native A/A base / head |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | -0.75% | -0.86% | -3.80% | +1.00% | +0.22% / +0.66% |
| core/call-control | -23.37% | -11.61% | +3.42% | +14.53% | +5.61% / +14.46% |
| core/integer-loop | +0.09% | -0.18% | -0.19% | +0.49% | +1.02% / -0.12% |
| core/function-calls | +0.37% | -0.60% | +0.49% | +0.93% | +0.20% / +0.16% |
| core/memory-load-store | +0.88% | +0.01% | -1.11% | +0.54% | +0.87% / -1.01% |
| memory/narrow-unaligned | +0.15% | +0.51% | +0.96% | -0.53% | -0.38% / -0.45% |
| memory/dependent-address | -0.02% | +0.02% | +0.41% | +1.10% | +0.68% / -0.57% |
| memory/float64-unaligned | +0.97% | +0.12% | -3.90% | -1.98% | +1.35% / +0.60% |
| runtime/simd-memory-full | +0.23% | -0.94% | +2.52% | -0.79% | +0.02% / +0.07% |
| runtime/simd-memory-widen | +7.53% | +5.34% | -0.07% | -3.72% | -2.56% / -5.31% |
| runtime/simd-memory-lanes | -0.23% | -0.36% | -0.09% | +1.57% | +3.45% / -3.37% |
| gc/i31-roundtrip | +1.76% | -12.52% | -2.31% | -5.01% | +10.62% / +2.68% |
| gc/struct-access | -2.56% | -3.09% | -0.76% | -3.46% | +5.13% / -2.32% |
| gc/struct-allocation | -2.31% | -2.64% | -4.65% | -0.41% | -3.93% / +5.33% |
| gc/struct-indirect-access | -1.61% | -5.34% | -0.47% | +2.80% | -5.15% / -0.88% |
| threads/shared-memory | -0.99% | +0.11% | +0.88% | +0.95% | +0.02% / -0.07% |
| threads/atomic-rmw | -0.39% | +0.60% | +3.00% | +6.71% | +0.28% / -0.75% |
| threads/notify-empty | -6.25% | -8.90% | -4.46% | -0.41% | -1.69% / -8.28% |
| threads/wait-mismatch | -13.35% | -10.43% | -0.93% | -0.45% | +5.39% / +0.40% |
| threads/wait-timeout | -29.19% | -12.84% | +0.36% | -1.97% | +2.06% / +7.68% |

Retain native SIMD widening Go +7.53% / LLGo +5.34%, LLGo interpreter atomic +6.71%, tiny-call interpreter +14.53%, both Go SHA regressions and every other adverse row. Source code unchanged in a subsystem does not waive measured outcomes. No startup/load/compile/instantiate/lookup/first/warmup/warm/cleanup medians are summed or subtracted.

## Disposition

Keep the upstream runtime issue and verified capability-discovery mechanism separate from acceptance of this placement. The next candidate should separate cold optional capability metadata from the hot Execution record and retain its existing field prefix. One bounded structural change will be qualified and measured as a new source event; no padding/alignment sweep or favorable timing retry. The present cohort remains unchanged.

Global live-range allocation, GC/indirect access, short-call/init gaps, controlled competitor/native acceptance and broader application/WASI-world coverage remain open. External PR2991 remains the sole contribution PR; no parallel PR or external merge is performed. Production main is unchanged.

[Lossless evidence](evidence/performance-backend-capabilities-2026-11-05.json.xz) includes sources/compilers, full local qualification/coverage/ownership checks, all timings and verifiers, before/after stack samples, the tested LLGo2993 reproducer/results/source recheck, and preceding O CI audit.
