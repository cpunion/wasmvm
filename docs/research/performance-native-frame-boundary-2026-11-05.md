# Native frame boundary isolation: 2026-11-05

**Partial interpreter recovery; native gains retained.** Isolating native packet construction/validation/cleanup from portable dispatch reduces C/Rust SHA interpreter durations against the preceding native-call source by **6.91% / 6.93% under Go** and **2.19% / 2.12% under LLGo**. Against production, Go remains **+2.01% / +1.49%**, LLGo **+4.01% / +4.36%**. Complete performance acceptance remains open.

Measured source `22f1a1dd635c102b0c63b27c57d1148a85da4743`, preceding native-call engine `2e0c605b2570696b8eed9fe796fe53ca3529ac12`, production engine `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-file identical to main `4ce79899a1f3ed45163afa622cc408cc6a79a599`). Two source files change 32 inserted/25 removed lines against the preceding source. The explicit noninlined native bridge preserves packet validation, spill/root publication, consumed fuel, error cleanup and relocated storage reuse. The shared loop refreshes active frame metadata after the bridge; emitted guest code and policies remain unchanged. No padding, alignment, guest, toolchain/profile or reference-policy change is included.

## Qualification and scope

All nine complete affected packages pass Go 1.27.3 and LLGo 3571 (`-O3 -lto=full`) on macOS arm64, with **96 actual conformance reports**. Exact scoped coverage remains **91.20%–100%**. Existing frame models and actual generated-call tests exercise mixed tuples, host roots, forced guest GC, recursive growth/limits, large edge moves, cancellation, traps, cleanup and reuse. The public compiled-guest checker passes **64 checksum calls / 48 expected traps** across eight configurations. Race, cgocheck2 and no-cgo checks pass; unavailable-native outcomes are not native execution evidence. Final four-profile 31-package CI/ 30 coverage gates and review remain required.

Both linked host binaries retain a separate `runBackend` body and one call from `Execution.Run`. Static dispatch instructions change Go 3166→3041 and LLGo 2929→2855; the bridge has 136/157 instructions. Direct LLGo allocation sites in Run change 34→31, with 4 in the bridge. These are static sites, not allocation-event counts or an isolated throughput cause. The annotation keeps this architectural boundary stable under full LTO.

## New source comparison

One new three-source cohort times production, native-call K and isolated-boundary L together with unchanged pinned Go/LLGo profiles, guests and references. The original20 portfolio independently verifies **1224 VM/reference +27 C observations, 408 alternate checks, 12 explicitly reused unsupported GC outcomes and 417 summaries**. The application cohort verifies **126 VM/reference +6 C observations, 42 alternate checks and 44 summaries**. Three fresh process blocks, two warmups/seven interleaved samples, same-binary LLGo production/head A/A controls. No timing retry, discarded adverse row or multiplication across independent cohorts.

Warm ns/message block for the C/Rust full SHA-256 modules. Each call resets state and processes 5000 blocks, including message generation, chaining and padding; the independent result is first 32 digest bits. These are two compiled Core compute modules, not a full application/WASI-world portfolio.

| Guest | Go interpreter vs production | vs K | LLGo interpreter vs production | vs K | LLGo interpreter A/A production/head |
| --- | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | +2.01% | -6.91% | +4.01% | -2.19% | +0.19% / +0.41% |
| applications/sha256-rust | +1.49% | -6.93% | +4.36% | -2.12% | -0.14% / -0.21% |

| Guest | LLGo native production → K → L | Go native L | Wasmtime | Wasmer | wazero compiler | Native C | LLGo native A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| applications/sha256-clang | 713.192 → 250.100 → 249.892 | 252.675 | 185.908 | 212.933 | 290.958 | 160.525 | -1.33% / +0.28% |
| applications/sha256-rust | 665.383 → 250.642 → 249.992 | 254.625 | 185.175 | 187.617 | 317.683 | 160.192 | +2.78% / +0.46% |

LLGo native still beats wazero compiler on these inputs but remains about 1.35x Wasmtime and 1.56x unchecked C. Source-relative native changes against K are −0.08%/−0.26%, within A/A controls. Reference cancellation/fuel/lifecycle policies differ and stay explicit; C omits VM obligations. Every load/setup/validate-compile/instantiate/lookup/first/warmup/warm/cleanup and process-readiness record remains separate in evidence. LLGo initialization and overall controlled targets remain unresolved.

## Complete original20 outcomes

Positive duration changes are slower. All production/head and prior/head comparisons come directly from this cohort. The complete machine-readable evidence also includes Go/LLGo values and every reference.

| Workload | Go interpreter vs production | vs K | LLGo interpreter vs production | vs K | LLGo native vs production | vs K |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | +6.26% | -0.94% | +0.44% | -4.94% | -0.54% | -2.17% |
| core/call-control | -10.21% | -11.68% | -37.20% | -23.44% | +30.11% | -1.05% |
| core/integer-loop | +1.96% | -6.32% | +2.61% | +0.10% | +0.99% | +0.83% |
| core/function-calls | -4.68% | -10.18% | -2.07% | -2.58% | +0.97% | +0.42% |
| core/memory-load-store | +1.90% | -3.99% | +4.23% | +0.15% | +1.05% | -0.11% |
| memory/narrow-unaligned | +0.76% | -4.88% | +3.10% | +0.24% | -0.01% | +0.32% |
| memory/dependent-address | +6.01% | -0.36% | +3.80% | +2.46% | +0.37% | -0.30% |
| memory/float64-unaligned | +3.56% | -5.23% | +2.38% | -1.96% | +1.49% | -0.05% |
| runtime/simd-memory-full | +1.00% | -4.11% | +1.91% | +0.41% | +0.10% | -0.15% |
| runtime/simd-memory-widen | +2.38% | -7.29% | -0.95% | -2.68% | -0.24% | -0.02% |
| runtime/simd-memory-lanes | +2.12% | -8.85% | +2.26% | -2.19% | -0.05% | +0.54% |
| gc/i31-roundtrip | +3.04% | -6.36% | +1.14% | -0.05% | +3.00% | -3.42% |
| gc/struct-access | +1.79% | -9.26% | +0.78% | -2.04% | +6.75% | -3.06% |
| gc/struct-allocation | +2.97% | +0.01% | +0.26% | +1.07% | -2.01% | -0.61% |
| gc/struct-indirect-access | +2.88% | -6.88% | +0.59% | -1.95% | +1.37% | -1.69% |
| threads/shared-memory | +1.04% | -3.28% | +2.98% | -5.50% | +0.11% | -0.31% |
| threads/atomic-rmw | +2.36% | -2.76% | -1.60% | -0.95% | +0.71% | +0.54% |
| threads/notify-empty | +7.15% | +0.75% | +1.04% | +0.62% | +18.38% | -1.27% |
| threads/wait-mismatch | +0.90% | -0.27% | +1.81% | +0.36% | +2.08% | -11.96% |
| threads/wait-timeout | +5.09% | -2.15% | +0.63% | -1.03% | +13.97% | -1.46% |

Retain production-relative Go private-memory interpreter +6.26%, dependent-address +6.01%, LLGo memory-load-store +4.23% and all other adverse interpreter rows. Native GC struct-access +6.75%, short notify +18.38%, wait-timeout +13.97% and tiny call controls remain unresolved; large tiny-call A/A differences are not parity evidence. A partial application recovery does not erase any of these rows.

## Next step

The portable inner loop still reassigns its function/code/instance/value-view locals after a native frame transfer. This may inhibit compiler caching even when the interpreter backend is selected; it is an inference from the source structure, not an established compiler bug. Investigate keeping frame selection in the outer loop and servicing an already charged native boundary once after selecting its new frame. Preserve the existing shared runtime and exact semantic contracts; no duplicated interpreter, padding sweep or favorable timing retry.

The previous K results remain unchanged. The sole external contribution PR2991 is still open at the latest read; the prior parallel-PR question is unanswered and no new PR was opened. LLGo #2983 remains owned by another contributor. Production main is unchanged. [Lossless evidence](evidence/performance-native-frame-boundary-2026-11-05.json.xz) retains the exact patch/source/compilers, semantic/coverage/ownership checks, both complete cohorts, static linked code, independent verifiers and compiled application inputs. Final target CI/review, controlled performance and expanded application/WASI-world acceptance remain open.
