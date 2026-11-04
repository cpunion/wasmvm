# Native boundary ownership paths: 2026-11-04

Measured engine `df3ac3314f83b77bd1f1e88d7bacb342052df30a`, baseline `fe149a44aa86d175d5e35c6f15345960d2b7a6a1` (engine-file identical to merged PR #118 / `f4e402e`). Go 1.27.3, LLGo `983465e779527ff496857a9a1915c1b57131091f`, LLVM 22.1.8, `-O3 -lto=full`; Apple M4 Max/macOS arm64. No compiler, profile, quota, cancellation, feature or reference-policy upgrade. Controlled all-target performance acceptance remains open.

## Mechanism and complete contracts

The unchanged-memory path checks idle state and unlocks directly; ownership transfer retains deferred cleanup around nested memory locks. Nonallocating native entries share the ordinary driver without an allocation-reservation unwind frame. Allocating entries retain their cleanup on every error and panic, including before native entry, plus publication after each exit. The common frame machine writes complete outer result tuples into reusable storage rather than appending each value. Public results remain independently owned; multi-value order and both v128 words survive shape changes, empty returns and workspace quiescence.

Pinned LLGo assembly has one setjmp site in each former common entry (`BindMemory` and native `Run`), and none in either new common entry. The allocation-only helper still has one. This is a scoped removal of unwind setup, not evidence that whole calls or every remaining callee allocate zero bytes. The existing [LLGo defer report](https://github.com/xgo-dev/llgo/issues/2932) already covers conditional-defer setup; no duplicate issue is opened. [#2983](https://github.com/xgo-dev/llgo/issues/2983) remains with its existing contributor.

A separate checksum-checked CPU diagnostic runs ten million complete calls per compiler/backend, after 1000 warmups and a ready handshake, with three-second macOS sampling requests. Captures are qualitative stacks, not calibrated function timings or replacement throughput samples. LLGo stacks include operation allocation/host collection, ownership setup and the C entry; unknown frames and background threads are not charged to guest instructions. An initial invalid objdump option was corrected before the static audit. No failed timing run was replaced.

## Batch-inclusive boundary comparison

This is a new shared harness over exact baseline/head source archives; its extra batching requires separate builds and does not replace the original successful baseline runners. Each timed batch includes 10000 calls and result checks. Two guest fixtures use one empty invocation or 64 integer iterations per call. Five fresh processes per row, two warmups and seven interleaved calls retain 520 observations, 104 summaries and 104 alternate-seed checks across 54 configurations. Eight exact baseline A/A aliases cover both compilers/backends under each parent policy. Setup, first batch, warm batches and cleanup remain separate; a first batch is not an unbatched first call.

Warm ns/complete public call:

| Guest / backend / parent | Go base → head | Change / A/A | LLGo base → head | Change / A/A |
| --- | ---: | ---: | ---: | ---: |
| call-control / native / background | 305.946 → 292.150 | -4.51% / -2.32% | 839.383 → 717.150 | -14.56% / +0.50% |
| call-control / native / live parent | 536.875 → 522.925 | -2.60% / -0.35% | 1537.708 → 1437.471 | -6.52% / +2.04% |
| call-control / interpreter / background | 226.458 → 224.975 | -0.66% / -1.08% | 637.108 → 627.225 | -1.55% / +1.16% |
| call-control / interpreter / live parent | 263.683 → 261.796 | -0.72% / -0.41% | 776.529 → 764.788 | -1.51% / -0.66% |
| integer-loop / native / background | 395.546 → 375.092 | -5.17% / +1.15% | 968.354 → 846.263 | -12.61% / -1.11% |
| integer-loop / native / live parent | 643.850 → 625.025 | -2.92% / -1.34% | 1689.404 → 1556.271 | -7.88% / +1.39% |
| integer-loop / interpreter / background | 2788.917 → 2791.554 | +0.09% / +0.26% | 2774.279 → 2655.992 | -4.26% / +0.47% |
| integer-loop / interpreter / live parent | 2837.258 → 2847.029 | +0.34% / +1.19% | 2883.100 → 2789.025 | -3.26% / +0.52% |

The empty LLGo native public call improves 14.6%, with live-parent improvement 6.5%; the remaining full boundary is 2.45x Go in the background row. These policies are separate scopes, not interchangeable scores.

Other boundary components stay separate from the public API:

| Guest / native component | Go base → head | LLGo base → head |
| --- | ---: | ---: |
| call-control / session | 215.062 → 207.088 | 501.825 → 397.142 |
| call-control / dispatch | 130.954 → 115.117 | 219.542 → 145.208 |
| integer-loop / session | 317.971 → 297.846 | 612.800 → 491.567 |
| integer-loop / dispatch | 221.688 → 206.863 | 321.267 → 239.796 |

Empty raw Context.RunState is 78.542 ns under Go / 111.588 under LLGo. It has no invocation cancellation binding; its preparation is grouped and cleanup is unmeasured, with the raw zero retained. It does not stand in for a full public call.

| Independent reference | Empty | Integer 64 |
| --- | ---: | ---: |
| wazero-compiler-default-background | 13.658 | 82.142 |
| wazero-compiler-enabled-background | 214.217 | 1348.983 |
| wazero-compiler-enabled-live | 362.929 | 1329.250 |
| wazero-interpreter-default-background | 67.725 | 5454.125 |
| wasmtime-cranelift | 59.375 | 149.096 |
| wasmer-cranelift | 136.271 | 216.246 |

Wazero default and enabled context termination remain named separately. Wasmtime/Wasmer are C embedding APIs. Their lifecycle/cancellation policies differ from ours; no median subtraction or whole-policy equivalence is claimed. C batch adapters use the already selected CLOCK_UPTIME_RAW on Darwin, matching the owned clocks; their older raw-monotonic binaries are retained as provenance and are not timed in this cohort. Pinned SDK hashes are unchanged. Retain adverse Go interpreter empty dispatch +2.31% and any other adverse rows in the machine-readable evidence; a component win does not erase a public regression.

## Longer affected kernels

The original eight longer-count cases run once over the exact existing PR118 baseline binaries, five fresh interleaved process blocks, two warmups/seven samples. All 300 observations, 60 alternate checks and 75 independent oracle checks pass. Counts remain 300k allocation, 100M direct GC/i31, 30M indirect/wait, 1M interpreter access and 50M native memory. Warm process medians are 9.15–138.76 ms. Units are ns/source iteration.

| Kernel / backend | Go base → head | Change / A/A | LLGo base → head | Change / A/A |
| --- | ---: | ---: | ---: | ---: |
| gc/struct-allocation / native | 34.638 → 34.707 | +0.20% / -1.01% | 31.453 → 30.613 | -2.67% / -0.60% |
| gc/struct-allocation / interpreter | 104.422 → 101.074 | -3.21% / -1.12% | 125.896 → 126.907 | +0.80% / -0.47% |
| gc/struct-access / native | 1.286 → 1.285 | -0.02% / -0.12% | 1.285 → 1.286 | +0.02% / -0.10% |
| gc/struct-indirect-access / native | 4.494 → 4.478 | -0.36% / +0.42% | 4.483 → 4.510 | +0.62% / -0.09% |
| gc/struct-access / interpreter | 50.220 → 49.887 | -0.66% / -0.20% | 42.416 → 41.523 | -2.11% / +0.59% |
| threads/wait-mismatch / native | 1.019 → 1.026 | +0.69% / -2.81% | 1.024 → 1.020 | -0.38% / -0.71% |
| gc/i31-roundtrip / native | 1.352 → 1.362 | +0.73% / -0.06% | 1.353 → 1.353 | +0.01% / +0.18% |
| core/memory-load-store / native | 0.422 → 0.424 | +0.45% / +1.02% | 0.425 → 0.422 | -0.58% / -0.70% |

LLGo native linked allocation improves 2.67%; LLGo interpreted allocation increases 0.80% and native indirect access increases 0.62%. All phases, references and adverse outcomes remain in the archive.

## Independent original-count portfolio

The unchanged original20-case scope retains 864 VM/reference observations, 27 ordinary C observations, 288 alternate checks, 12 explicit unsupported outcomes and 297 summaries. Three process blocks use one public call, 1M ordinary iterations and 10k GC/notify/wait iterations. These shorter controls do not replace the longer affected cases or batch-inclusive call cohort. All phases and per-process controls remain distinct. Head/base changes:

| Kernel | Go native / interpreter | LLGo native / interpreter |
| --- | ---: | ---: |
| threads/private-memory | +0.12% / -0.05% | -1.72% / -0.97% |
| core/call-control | -4.98% / +3.42% | +5.57% / +2.80% |
| core/integer-loop | +0.17% / +0.07% | +0.96% / -1.31% |
| core/function-calls | +1.64% / +2.94% | +1.02% / -2.61% |
| core/memory-load-store | -2.71% / -0.93% | -0.48% / +0.84% |
| memory/narrow-unaligned | -0.69% / -1.61% | +0.83% / -0.28% |
| memory/dependent-address | -0.28% / -1.36% | +0.80% / -1.45% |
| memory/float64-unaligned | -0.17% / -4.14% | -1.43% / +0.02% |
| runtime/simd-memory-full | -0.79% / -2.68% | +1.01% / +0.15% |
| runtime/simd-memory-widen | +0.20% / -0.57% | +0.67% / -2.59% |
| runtime/simd-memory-lanes | -1.12% / -0.95% | +0.28% / +1.05% |
| gc/i31-roundtrip | -12.10% / +1.73% | +3.89% / -0.58% |
| gc/struct-access | +12.84% / -2.45% | +1.96% / -2.20% |
| gc/struct-allocation | -1.54% / -4.47% | -3.98% / +1.09% |
| gc/struct-indirect-access | +1.04% / -2.88% | +1.98% / -0.52% |
| threads/shared-memory | -0.18% / -0.26% | -0.79% / +0.92% |
| threads/atomic-rmw | +0.56% / -3.24% | +0.13% / +1.28% |
| threads/notify-empty | +13.40% / -5.72% | +9.52% / -2.21% |
| threads/wait-mismatch | +8.45% / -1.64% | +4.63% / -0.33% |
| threads/wait-timeout | -2.09% / -2.55% | +7.61% / -1.48% |

Retain LLGo native unbatched public-call +5.57%, notify +9.52% and timeout +7.61%, and Go native GC access +12.84%. The distinct longer/batched cohorts do not erase these original-count adverse observations. Full-portfolio acceptance remains open.

The ordinary Clang controls omit Wasm guards, cancellation, fuel, guest GC and embedding lifecycle. Their unchecked ns/iteration rates and source/flags are retained; they cannot establish full VM equivalence. No unsupported GC reference row is represented as zero.

## Validation and remaining acceptance

Both compilers execute all eight affected packages. The local report inventories match the merged baseline exactly: 48 execution reports per compiler, with unchanged pinned corpus counts/features, all five mandatory extensions at three tiers and forced Stack Switching collection. Go race, no-cgo and cgocheck2 checks pass. Direct affected-package coverage is not the final 30-package cross-package gate; final-head Go/LLGo × Linux amd64/macOS arm64 CI, package execution proofs, coverage and review remain required before integration.

Full native/public-call, interpreted dispatch, guest GC, startup, host allocation, cancellation calibration, expanded WASI world conformance and controlled whole-portfolio parity/native-C acceptance remain open. Neither a successful semantic matrix nor a local boundary improvement closes the overall goal.

[Lossless source/binary/toolchain/SDK identities, harness sources, plans, raw observations and orders, reports, static assembly, diagnostics and verification](evidence/performance-native-boundary-2026-11-04.json.xz).

Evidence archive: **266 files**, **672852 bytes**, SHA-256 `dd5f75cdd3b4fc9b3d99b1178913b25964bdea42358bfaec15f477f0262ae44c`. Every decoded file was round-trip checked. Final CI artifacts remain separate from the frozen local measurement bundle.
