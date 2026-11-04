# Native entry effects experiment: 2026-11-04

Status: **held, not integrated**. The complete candidate passes affected Go/LLGo semantics, but longer LLGo native linked GC regresses **4.56%** (A/A -0.46%) and the independent original-count LLGo interpreter SIMD lane workload regresses **17.67%** (A/A +1.01%). Short-call improvements do not resolve either adverse outcome. No PR or target CI is started for this candidate.

Measured source `cd2607eda4d87036a4fcb3c124c860b790fc0e71`, baseline `df3ac3314f83b77bd1f1e88d7bacb342052df30a` (engine-file identical to merged PR #119 / `3ee9890`). Apple M4 Max/macOS arm64; Go 1.27.3, LLGo `983465e779527ff496857a9a1915c1b57131091f`, LLVM 22.1.8, `-O3 -lto=full`. All baseline and competitor binaries, SDKs, fixture hashes, counts and corrected uptime clocks are reused. No compiler upgrade is mixed into this source experiment.

## Implemented mechanism and validation

The compiler proves complete i32/i64-only entries from all original PCs and every committed inline body. Other storage and unknown operations retain canonical FP handling. Immutable lease-owned entry metadata selects the C path; guest state cannot supply effects. Record v2/168 bytes, approved entry checks, memory/lease/retirement ownership, cancellation and macOS JIT execution protection remain intact. Metadata adds one boolean with padding: approved entries occupy eight bytes instead of four. This is a representation cost, not an RSS result.

The candidate includes classification tests for both ISAs/tiers, cold/unknown operations, immutable publication, version/cancellation guards and actual compiler-produced integer/f32/f64 entries under downward rounding, flush controls and sticky exceptions. Nearest rounding, gradual subnormals, complete host-state restoration and high-word canonicalization pass. Compiler test contexts exercise the published effects. Eight affected packages and all 48 pinned reports pass under each compiler, including every mandatory extension at three tiers and forced Stack Switching collection. Race, no-cgo and cgocheck2 pass. Local affected-package coverage is not final cross-package acceptance; the full 31-package dual-platform CI is deliberately not started for a held candidate.

Disassembly confirms the new integer C entry retains JIT execution protection and omits the three FP-environment calls. The default C entry retains them. Go/LLGo native adapters and GC helpers still have other costs; equal static allocation-site counts do not prove zero hot allocations or explain the GC/SIMD regressions. No new LLGo defect is inferred from these timing changes.

## Batch-inclusive boundaries

Two unchanged fixtures, 10000 complete calls per timed batch, five fresh interleaved process blocks, two warmups/seven samples: **520 observations, 104 alternate checks, 104 summaries, 54 configurations**. Exact baseline A/A aliases cover both compilers/backends and background/live parents. Setup/first batch/warm/cleanup remain separate. Units are ns/complete public call.

| Guest / backend / parent | Go base → head | Change / A/A | LLGo base → head | Change / A/A |
| --- | ---: | ---: | ---: | ---: |
| call-control / native / background | 293.479 → 267.358 | -8.90% / -0.63% | 722.408 → 703.237 | -2.65% / -0.00% |
| call-control / native / live | 521.125 → 492.579 | -5.48% / +0.72% | 1445.912 → 1433.862 | -0.83% / -1.50% |
| call-control / interpreter / background | 226.450 → 226.646 | +0.09% / +0.10% | 634.229 → 622.729 | -1.81% / -0.75% |
| call-control / interpreter / live | 264.058 → 263.371 | -0.26% / +0.68% | 762.208 → 757.800 | -0.58% / +0.91% |
| integer-loop / native / background | 372.012 → 346.054 | -6.98% / -0.12% | 829.417 → 809.562 | -2.39% / -0.64% |
| integer-loop / native / live | 602.392 → 571.154 | -5.19% / +0.54% | 1543.079 → 1554.763 | +0.76% / +1.37% |
| integer-loop / interpreter / background | 2697.842 → 2706.846 | +0.33% / +0.36% | 2593.817 → 2610.804 | +0.65% / +0.20% |
| integer-loop / interpreter / live | 2745.412 → 2728.637 | -0.61% / +0.33% | 2736.942 → 2736.083 | -0.03% / -0.67% |

LLGo native empty public calls improve 2.65%, session calls 5.07%, and dispatch 11.62%; Go public calls improve 8.90%. The Go/LLGo public-call gap therefore remains material. Retain LLGo live-parent integer +0.76% and Go empty interpreted dispatch +2.09%.

| Native component / guest | Go base → head | LLGo base → head |
| --- | ---: | ---: |
| session / call-control | 199.846 → 174.396 | 398.712 → 378.508 |
| dispatch / call-control | 116.888 → 90.838 | 145.996 → 129.033 |
| session / integer-loop | 282.925 → 258.792 | 489.492 → 465.250 |
| dispatch / integer-loop | 202.933 → 173.125 | 228.304 → 212.608 |

Default raw Context.RunState stays conservative: Go 82.346→75.888 ns, LLGo 110.129→111.692. Its lifecycle policy excludes an invocation cancellation binding and its cleanup is unmeasured. It does not replace public calls.

| Reference policy | Empty | Integer 64 |
| --- | ---: | ---: |
| wazero-compiler-default-background | 13.175 | 78.454 |
| wazero-compiler-enabled-background | 244.954 | 1260.400 |
| wazero-compiler-enabled-live | 360.258 | 1256.592 |
| wazero-interpreter-default-background | 67.600 | 5319.575 |
| wasmtime-cranelift | 59.671 | 143.058 |
| wasmer-cranelift | 137.137 | 209.808 |

Reference lifecycle/cancellation and embedding policies remain named separately. Wasmtime/Wasmer use C APIs. No timing subtraction or whole-policy equivalence is claimed.

## Longer affected kernels

The eight unchanged longer cases retain **300 observations, 60 alternate checks, 75 independent oracle checks and 60 summaries**. Five process blocks, two warmups/seven calls; counts are 300k allocation, 100M direct GC/i31, 30M indirect/wait, 1M interpreted access and 50M native memory. Units are ns/source iteration.

| Kernel / backend | Go base → head | Change / A/A | LLGo base → head | Change / A/A |
| --- | ---: | ---: | ---: | ---: |
| gc/struct-allocation / native | 34.795 → 34.563 | -0.67% / -0.48% | 30.605 → 32.000 | +4.56% / -0.46% |
| gc/struct-allocation / interpreter | 101.974 → 101.282 | -0.68% / +0.36% | 122.318 → 126.479 | +3.40% / +2.08% |
| gc/struct-access / native | 1.281 → 1.285 | +0.29% / +0.37% | 1.284 → 1.284 | +0.05% / +0.22% |
| gc/struct-indirect-access / native | 4.513 → 4.509 | -0.08% / -0.36% | 4.509 → 4.500 | -0.18% / -0.02% |
| gc/struct-access / interpreter | 50.178 → 50.717 | +1.07% / -0.43% | 41.475 → 41.593 | +0.28% / +1.10% |
| threads/wait-mismatch / native | 1.025 → 0.977 | -4.64% / -3.89% | 1.016 → 1.014 | -0.21% / +0.20% |
| gc/i31-roundtrip / native | 1.336 → 1.336 | +0.03% / -0.13% | 1.336 → 1.335 | -0.09% / -0.09% |
| core/memory-load-store / native | 0.419 → 0.420 | +0.22% / -0.27% | 0.418 → 0.419 | +0.20% / -0.03% |

The native GC regression remains unresolved. An unchanged successful measurement is not retried to obtain a different result.

## Independent original-count portfolio

The original20-case comparison retains **864 VM/reference + 27 ordinary C observations, 288 alternate checks, 12 explicit unsupported outcomes and 297 summaries**. Three process blocks, two warmups/seven calls; one public call, 1M ordinary iterations, 10k GC/notify/wait iterations. All phases and shorter adverse observations remain distinct from the longer/batched cohorts. Head/base changes:

| Kernel | Go native / interpreter | LLGo native / interpreter |
| --- | ---: | ---: |
| threads/private-memory | +0.07% / +0.09% | -0.39% / +0.21% |
| core/call-control | +3.90% / +0.00% | +25.82% / -12.89% |
| core/integer-loop | -0.24% / +0.46% | +0.16% / +0.50% |
| core/function-calls | -0.42% / -4.22% | -0.93% / -1.45% |
| core/memory-load-store | +3.20% / +0.90% | +0.17% / +1.96% |
| memory/narrow-unaligned | +2.97% / -0.63% | -0.40% / -0.87% |
| memory/dependent-address | -0.57% / +1.24% | -0.74% / +1.09% |
| memory/float64-unaligned | +0.29% / -0.44% | +0.81% / -0.43% |
| runtime/simd-memory-full | -0.04% / +0.93% | -0.43% / -2.88% |
| runtime/simd-memory-widen | +1.76% / +1.89% | +1.22% / -1.60% |
| runtime/simd-memory-lanes | -0.03% / -0.45% | -0.58% / +17.67% |
| gc/i31-roundtrip | +3.34% / +3.54% | +5.43% / -4.90% |
| gc/struct-access | +0.42% / +1.13% | -3.25% / +1.02% |
| gc/struct-allocation | -0.77% / -0.92% | -7.95% / +1.42% |
| gc/struct-indirect-access | +1.41% / +0.59% | -3.57% / +0.50% |
| threads/shared-memory | -1.32% / -1.99% | -0.93% / +1.42% |
| threads/atomic-rmw | +0.15% / +0.82% | +1.35% / -1.36% |
| threads/notify-empty | +11.71% / -1.66% | -9.19% / +1.81% |
| threads/wait-mismatch | +18.55% / +0.70% | -8.69% / -2.48% |
| threads/wait-timeout | +8.39% / +1.34% | +3.65% / +0.54% |

The 17.67% SIMD lane regression is not dismissed by unchanged interpreter source, different code addresses or gains elsewhere. The unbatched LLGo native public control also increases 25.82%; its different protocol does not replace the batch-inclusive result. Ordinary Clang controls omit VM guards, cancellation, fuel, GC and embedding lifecycle; unsupported GC references are explicit, never zero-valued.

## Next compiler baseline event

A read-only upstream check confirms [LLGo #2958](https://github.com/xgo-dev/llgo/issues/2958) closed through [PR #2971](https://github.com/xgo-dev/llgo/pull/2971), merged at `dd109ecb9bd4a5d9794c83a9bfa8e9f0e3decd16` on 2026-11-03. Its immediate Cgo closure fix is newer than this study's fixed compiler. The next experiment will evaluate that compiler independently over the unchanged merged engine, preserving source/compiler attribution. It is not evidence that all Cgo allocations are fixed. #2983 remains with its existing contributor; no upstream write is made.

[Lossless experiment sources/patches, hashes, harnesses, raw records/orders, all reports, assembly, issue-status reads and verification](evidence/performance-entry-effects-2026-11-04.json.xz).

Evidence archive: **252 files**, **619744 bytes**, SHA-256 `5956c6bcfde34bfcb37910e09854ee89bd4336c4ad5fa8d4b971fd600974663a`. Every decoded file was round-trip checked. No target CI was started for this held candidate.
