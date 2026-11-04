# LLGo Cgo compiler study: 2026-11-04

This is an independent compiler baseline event over unchanged merged PR #119 engine `df3ac3314f83b77bd1f1e88d7bacb342052df30a`, byte-identical to main `3ee9890`. Its 1483 engine files have digest `56cef1eeaf84398aaeef19a30a095491f5a4c138cf40d3e3ea9b34c148338459`. No production VM implementation, cancellation policy, quota, feature gate or ABI changes. Apple M4 Max/macOS arm64; Go 1.27.3, LLVM 22.1.8; both LLGo candidates and baseline retain `-O3 -lto=full`. Controlled whole-portfolio acceptance remains open.

Existing compiler: `983465e779527ff496857a9a1915c1b57131091f`. Complete upstream candidate: `dd109ecb9bd4a5d9794c83a9bfa8e9f0e3decd16`, 44 reachable commits ahead, containing several compiler/runtime fixes. Cgo-only candidate: `3571c2523e010e6842c4dd8a5213554fdc72d2a3`, exactly five commits based directly on the existing pin, already included in [merged LLGo #2971](https://github.com/xgo-dev/llgo/pull/2971), which resolves [#2958](https://github.com/xgo-dev/llgo/issues/2958). Complete-range results cannot all be attributed to that one fix. The Cgo-only comparison isolates its existing branch from unrelated runtime/append changes; no LLGo source or upstream state is modified. [#2983](https://github.com/xgo-dev/llgo/issues/2983) remains with its existing contributor.

## Method and execution evidence

Each candidate is a separate experiment, with its own contemporaneous baseline and A/A controls. Exact successful PR119 Go, old-LLGo and reference binaries are reused; only candidate-compiler runners are built. All four role/source manifests, compiler executable/build identities, common instrumentation and runner hashes are checked. Unchanged Go base/head roles use the same binary as a control; their differences expose run variation. Candidate semantic execution passes eight affected packages and all 48 expected reports, with exact pinned corpus counts/features, all five mandatory extensions on three tiers and forced Stack Switching collection. This local macOS scope does not establish final Linux/full-package CI acceptance.

Batch calls use five fresh serially interleaved process blocks, two warmups, seven measured batches and 10000 complete calls per batch. Empty and 64-iteration integer guests retain 520 observations, 104 alternate-seed checks/summaries and eight exact A/A aliases across 54 configurations. Public API, retained session, internal dispatch and raw native entry are distinct boundaries; background/live parent and reference termination settings remain separate. Preparation, first batch, warm batches, cleanup and process readiness are retained; a first batch is not an unbatched first call. No clock, startup or control time is subtracted.

The eight longer-count cases retain 300 observations, 60 alternate checks, 75 independent oracle checks and 60 summaries per candidate. The separate original20 portfolio retains 864 VM/reference observations, 27 ordinary-C observations, 288 alternate checks, 12 explicitly unsupported reference outcomes and 297 summaries per candidate. It uses three process blocks, one invocation per sample, 1M ordinary iterations and 10k GC/notify/wait iterations. Counts/cohorts are never pooled; shorter cases cannot erase adverse longer ones. References stay wazero 1.12.0, Wasmtime 48.0.1, Wasmer 7.3.0 and WasmEdge 0.17.1. C references retain the previously corrected CLOCK_UPTIME_RAW; Go/LLGo clocks are unchanged.

## Independent allocation diagnostic

Identical source executes one million operations/case, with checksum 500001000000 and checked C/Go-buffer pointer writes. Reported TotalAlloc is approximate under LLGo/BDWGC. Unsupported Mallocs counters cannot establish zero allocation events. This is allocation evidence, separate from VM throughput. Go and old-LLGo records are reused for the second candidate.

| Case | Go bytes | Existing LLGo bytes | Complete candidate bytes | Cgo-only bytes |
| --- | ---: | ---: | ---: | ---: |
| c-scalar | 0 | 16002304 | 0 | 0 |
| c-pair | 0 | 15997840 | 0 | 0 |
| c-owned-pointer | 16 | 80004256 | 0 | 0 |
| go-buffer-pointer | 0 | 80002192 | 0 | 0 |
| go-pair | 0 | 0 | 0 | 0 |

## Complete upstream candidate — held

Batch-inclusive warm ns/call; changes compare each experiment's own baseline:

| Guest / backend / boundary | LLGo base → candidate | Go unchanged control |
| --- | ---: | ---: |
| call-control / native / api | 731.046 → 773.217 (+5.77%) | 296.113 → 298.846 (+0.92%) |
| call-control / native / api-live | 1438.783 → 1632.513 (+13.46%) | 537.462 → 534.471 (-0.56%) |
| call-control / native / session | 393.629 → 360.750 (-8.35%) | 203.367 → 205.158 (+0.88%) |
| call-control / native / dispatch | 145.654 → 94.296 (-35.26%) | 116.662 → 120.704 (+3.46%) |
| call-control / interpreter / api | 629.696 → 722.663 (+14.76%) | 228.933 → 228.188 (-0.33%) |
| call-control / interpreter / api-live | 760.279 → 871.600 (+14.64%) | 267.579 → 267.337 (-0.09%) |
| call-control / interpreter / session | 305.371 → 311.846 (+2.12%) | 133.392 → 138.454 (+3.80%) |
| call-control / interpreter / dispatch | 59.850 → 58.763 (-1.82%) | 63.346 → 66.438 (+4.88%) |
| integer-loop / native / api | 822.996 → 902.692 (+9.68%) | 373.504 → 371.800 (-0.46%) |
| integer-loop / native / api-live | 1523.692 → 1769.062 (+16.10%) | 605.967 → 608.304 (+0.39%) |
| integer-loop / native / session | 489.521 → 460.146 (-6.00%) | 285.225 → 285.113 (-0.04%) |
| integer-loop / native / dispatch | 233.221 → 183.967 (-21.12%) | 203.000 → 203.229 (+0.11%) |
| integer-loop / interpreter / api | 2628.129 → 2656.333 (+1.07%) | 2707.463 → 2709.658 (+0.08%) |
| integer-loop / interpreter / api-live | 2740.171 → 2814.808 (+2.72%) | 2749.562 → 2743.183 (-0.23%) |
| integer-loop / interpreter / session | 2238.525 → 2226.175 (-0.55%) | 2631.125 → 2620.733 (-0.39%) |
| integer-loop / interpreter / dispatch | 2020.167 → 1944.883 (-3.73%) | 2541.996 → 2527.821 (-0.56%) |

| Empty native component / reference | ns/call |
| --- | ---: |
| base-go-native-entry | 75.854 |
| head-go-native-entry | 83.092 |
| base-llgo-native-entry | 113.062 |
| head-llgo-native-entry | 64.338 |
| wasmtime-cranelift | 59.825 |
| wasmer-cranelift | 139.033 |
| wazero-compiler-default-background | 13.787 |
| wazero-compiler-enabled-background | 267.817 |
| wazero-compiler-enabled-live | 400.646 |

Raw Context.RunState lacks complete invocation cancellation binding, public ownership and lifecycle setup. Its preparation is grouped and cleanup is unmeasured; the raw zero is retained. Reference cancellation policies differ. None of these component rows substitutes for the full public API.

Longer-count warm ns/source iteration:

| Workload / backend / count | LLGo base → candidate | Go base → candidate |
| --- | ---: | ---: |
| gc/struct-allocation / native / 300000 | 30.554 → 30.892 (+1.11%) | 34.914 → 35.194 (+0.80%) |
| gc/struct-allocation / interpreter / 300000 | 121.778 → 126.725 (+4.06%) | 101.060 → 101.815 (+0.75%) |
| gc/struct-access / native / 100000000 | 1.296 → 1.293 (-0.23%) | 1.293 → 1.295 (+0.20%) |
| gc/struct-indirect-access / native / 30000000 | 4.743 → 4.792 (+1.02%) | 4.865 → 4.875 (+0.21%) |
| gc/struct-access / interpreter / 1000000 | 49.805 → 49.336 (-0.94%) | 57.687 → 58.401 (+1.24%) |
| threads/wait-mismatch / native / 30000000 | 1.081 → 1.084 (+0.30%) | 1.062 → 1.079 (+1.61%) |
| gc/i31-roundtrip / native / 100000000 | 1.334 → 1.334 (+0.02%) | 1.331 → 1.332 (+0.13%) |
| core/memory-load-store / native / 50000000 | 0.418 → 0.417 (-0.30%) | 0.417 → 0.420 (+0.81%) |

Original20 head/base changes; native/interpreter remain separate. Single empty calls are timer-sensitive and do not replace batch evidence.

| Workload | Go native / interpreter | LLGo native / interpreter |
| --- | ---: | ---: |
| threads/private-memory | +5.80% / +1.08% | -0.70% / +2.37% |
| core/call-control | +0.00% / +8.33% | -31.62% / -1.99% |
| core/integer-loop | -0.72% / +0.05% | +0.51% / +3.47% |
| core/function-calls | +0.87% / +0.30% | -1.93% / -1.00% |
| core/memory-load-store | -0.49% / -0.05% | -0.09% / +0.41% |
| memory/narrow-unaligned | +0.82% / +2.49% | -1.07% / +0.05% |
| memory/dependent-address | -0.54% / +1.01% | +0.25% / +2.38% |
| memory/float64-unaligned | +1.48% / +0.92% | +0.75% / -0.03% |
| runtime/simd-memory-full | -1.37% / +0.18% | +0.08% / -1.36% |
| runtime/simd-memory-widen | -1.37% / +0.99% | +0.35% / +0.22% |
| runtime/simd-memory-lanes | -0.39% / -0.33% | -1.07% / +0.47% |
| gc/i31-roundtrip | +8.38% / +0.25% | +5.47% / -3.62% |
| gc/struct-access | -0.39% / -2.57% | +6.28% / +2.39% |
| gc/struct-allocation | +7.71% / +1.12% | -0.81% / +2.16% |
| gc/struct-indirect-access | -0.44% / +2.87% | -1.94% / -0.46% |
| threads/shared-memory | -0.85% / -1.02% | +0.04% / +1.41% |
| threads/atomic-rmw | -0.55% / +0.61% | -0.17% / +1.07% |
| threads/notify-empty | +12.09% / +0.63% | +31.87% / +3.90% |
| threads/wait-mismatch | -14.50% / +4.13% | +7.92% / -3.16% |
| threads/wait-timeout | +13.42% / +1.47% | +7.50% / +1.19% |

The complete candidate improves empty raw native entry and dispatch substantially, but background native public calls increase 5.77%, live-parent native empty/integer calls increase 13.46%/16.10%, and interpreter public empty increases 14.76%. Old-LLGo A/A remains close in these rows. Long native/interpreted GC allocation increases 1.11%/4.06%. Original20 native notify increases 31.87%; unbatched empty native public improves 31.62%. These are independent retained outcomes, not permission to replace the adverse batch campaign. This complete-range candidate is not selected for integration. The origin of the non-Cgo regressions remains unproven.

## Cgo-only candidate

Batch-inclusive warm ns/call; changes compare each experiment's own baseline:

| Guest / backend / boundary | LLGo base → candidate | Go unchanged control |
| --- | ---: | ---: |
| call-control / native / api | 802.025 → 748.233 (-6.71%) | 328.529 → 319.079 (-2.88%) |
| call-control / native / api-live | 1572.075 → 1511.408 (-3.86%) | 595.829 → 592.013 (-0.64%) |
| call-control / native / session | 441.488 → 392.800 (-11.03%) | 228.750 → 232.050 (+1.44%) |
| call-control / native / dispatch | 162.279 → 104.971 (-35.31%) | 133.863 → 127.754 (-4.56%) |
| call-control / interpreter / api | 690.092 → 691.042 (+0.14%) | 258.196 → 252.429 (-2.23%) |
| call-control / interpreter / api-live | 829.017 → 839.833 (+1.30%) | 298.038 → 294.613 (-1.15%) |
| call-control / interpreter / session | 331.492 → 334.258 (+0.83%) | 147.346 → 145.150 (-1.49%) |
| call-control / interpreter / dispatch | 66.071 → 67.229 (+1.75%) | 69.825 → 68.342 (-2.12%) |
| integer-loop / native / api | 934.254 → 894.700 (-4.23%) | 419.717 → 420.087 (+0.09%) |
| integer-loop / native / api-live | 1772.079 → 1688.683 (-4.71%) | 682.737 → 695.971 (+1.94%) |
| integer-loop / native / session | 559.292 → 497.508 (-11.05%) | 318.579 → 332.421 (+4.34%) |
| integer-loop / native / dispatch | 253.508 → 208.183 (-17.88%) | 230.733 → 227.050 (-1.60%) |
| integer-loop / interpreter / api | 3002.504 → 2874.996 (-4.25%) | 3097.537 → 3090.988 (-0.21%) |
| integer-loop / interpreter / api-live | 3125.762 → 3037.829 (-2.81%) | 3277.383 → 3251.188 (-0.80%) |
| integer-loop / interpreter / session | 2582.150 → 2586.725 (+0.18%) | 2984.829 → 3023.613 (+1.30%) |
| integer-loop / interpreter / dispatch | 2352.621 → 2287.204 (-2.78%) | 2875.008 → 2884.954 (+0.35%) |

| Empty native component / reference | ns/call |
| --- | ---: |
| base-go-native-entry | 92.975 |
| head-go-native-entry | 87.854 |
| base-llgo-native-entry | 121.408 |
| head-llgo-native-entry | 71.725 |
| wasmtime-cranelift | 66.987 |
| wasmer-cranelift | 153.892 |
| wazero-compiler-default-background | 14.675 |
| wazero-compiler-enabled-background | 339.525 |
| wazero-compiler-enabled-live | 414.837 |

Raw Context.RunState lacks complete invocation cancellation binding, public ownership and lifecycle setup. Its preparation is grouped and cleanup is unmeasured; the raw zero is retained. Reference cancellation policies differ. None of these component rows substitutes for the full public API.

Longer-count warm ns/source iteration:

| Workload / backend / count | LLGo base → candidate | Go base → candidate |
| --- | ---: | ---: |
| gc/struct-allocation / native / 300000 | 31.299 → 32.328 (+3.29%) | 35.326 → 35.643 (+0.90%) |
| gc/struct-allocation / interpreter / 300000 | 142.060 → 148.938 (+4.84%) | 118.604 → 118.568 (-0.03%) |
| gc/struct-access / native / 100000000 | 1.290 → 1.287 (-0.25%) | 1.289 → 1.286 (-0.18%) |
| gc/struct-indirect-access / native / 30000000 | 4.523 → 4.507 (-0.36%) | 4.504 → 4.503 (-0.02%) |
| gc/struct-access / interpreter / 1000000 | 41.379 → 39.940 (-3.48%) | 49.489 → 49.795 (+0.62%) |
| threads/wait-mismatch / native / 30000000 | 1.022 → 1.017 (-0.43%) | 1.022 → 1.019 (-0.22%) |
| gc/i31-roundtrip / native / 100000000 | 1.311 → 1.312 (+0.07%) | 1.308 → 1.316 (+0.62%) |
| core/memory-load-store / native / 50000000 | 0.419 → 0.418 (-0.22%) | 0.420 → 0.420 (+0.07%) |

Original20 head/base changes; native/interpreter remain separate. Single empty calls are timer-sensitive and do not replace batch evidence.

| Workload | Go native / interpreter | LLGo native / interpreter |
| --- | ---: | ---: |
| threads/private-memory | +2.06% / +0.78% | -1.19% / +0.96% |
| core/call-control | +3.23% / -4.96% | -24.71% / -8.70% |
| core/integer-loop | -1.04% / +0.42% | +0.50% / -3.43% |
| core/function-calls | +2.23% / -0.09% | +1.60% / -2.38% |
| core/memory-load-store | +0.15% / -0.62% | -0.31% / -5.60% |
| memory/narrow-unaligned | -0.39% / +1.72% | +0.86% / -5.52% |
| memory/dependent-address | -2.33% / -0.88% | +0.39% / -1.94% |
| memory/float64-unaligned | -1.33% / +0.68% | +0.33% / -5.46% |
| runtime/simd-memory-full | +1.13% / +0.16% | +1.77% / -1.14% |
| runtime/simd-memory-widen | -0.68% / +0.36% | -1.44% / +0.94% |
| runtime/simd-memory-lanes | -2.84% / +0.89% | +0.24% / +0.76% |
| gc/i31-roundtrip | +7.02% / +1.82% | -1.82% / -4.73% |
| gc/struct-access | -0.21% / -0.62% | +1.76% / -2.28% |
| gc/struct-allocation | +0.86% / +0.79% | +4.39% / +0.42% |
| gc/struct-indirect-access | +0.38% / -1.68% | -1.40% / -3.46% |
| threads/shared-memory | +1.97% / -1.62% | +0.06% / -0.77% |
| threads/atomic-rmw | +0.56% / -0.13% | -0.17% / -1.56% |
| threads/notify-empty | -7.40% / +2.59% | -0.27% / +3.53% |
| threads/wait-mismatch | -2.91% / -2.87% | +3.15% / -1.31% |
| threads/wait-timeout | +16.55% / +0.90% | +1.62% / +0.56% |

## Cgo-only residual gaps

Each comparison below comes from its own contemporaneous cohort; no cross-campaign normalization or phase subtraction. Smaller elapsed-time ratios are better.

| Longer native kernel | Candidate LLGo ns/iteration | Go | Wasmtime | LLGo / Wasmtime |
| --- | ---: | ---: | ---: | ---: |
| gc/struct-allocation | 32.328 | 35.643 | 9.102 | 3.55x |
| gc/struct-access | 1.287 | 1.286 | 1.028 | 1.25x |
| gc/struct-indirect-access | 4.507 | 4.503 | 2.568 | 1.75x |
| threads/wait-mismatch | 1.017 | 1.019 | 8.702 | 0.12x |
| gc/i31-roundtrip | 1.312 | 1.316 | 1.738 | 0.75x |
| core/memory-load-store | 0.418 | 0.420 | 0.470 | 0.89x |

| Original-count native kernel | LLGo ns/iteration | Ordinary C | LLGo / C |
| --- | ---: | ---: | ---: |
| core/integer-loop | 1.557 | 1.545 | 1.01x |
| core/function-calls | 0.784 | 0.780 | 1.01x |
| core/memory-load-store | 0.433 | 0.271 | 1.60x |
| memory/narrow-unaligned | 0.688 | 0.454 | 1.51x |
| memory/dependent-address | 1.813 | 1.411 | 1.29x |
| memory/float64-unaligned | 0.726 | 0.657 | 1.11x |
| runtime/simd-memory-full | 3.203 | 3.226 | 0.99x |
| runtime/simd-memory-widen | 1.011 | 1.011 | 1.00x |
| runtime/simd-memory-lanes | 4.891 | 4.437 | 1.10x |

The Cgo-only empty native public boundary changes 802.025 → 748.233 ns (-6.71%), remaining 2.34x Go and 11.17x Wasmtime in this batch cohort. Its raw entry is 71.725 ns versus Go 87.854; that component improvement does not establish full-call parity. Long native/interpreted GC allocation retains +3.29%/+4.84%. The complete-range migration stays held, and the selected pin still requires final-head CI/review.

## Remaining gaps and integration gate

The evidence decision JSON includes an unweighted per-case original20 geometric mean as a diagnostic only. It includes call controls and is not the family-weighted controlled acceptance metric in the performance contract. The unchanged Go binary retains an original-count native wait-timeout increase of 16.55%; this outcome stays visible alongside the candidate results.

Both compiler studies preserve all adverse phases, Go controls, A/A rows, unsupported outcomes and unchecked ordinary-C controls in the archive. Ordinary C omits Wasm guards, cancellation, fuel, guest GC and embedding lifecycle; its retained ns/iteration ratios are diagnostic, not full equivalence. The original complete-range candidate and the earlier [held entry-effects implementation](performance-entry-effects-2026-11-04.md) are not integrated by a compiler-pin contribution.

A pin contribution requires final-head Go/LLGo × Linux amd64/macOS arm64 execution of all 31 packages, all 30 coverage gates, pinned release flags, current review and the standard PR benchmark. That PR benchmark uses the selected compiler on both source base/head, separately from the compiler comparisons here. Full public-call/guest-GC/ordinary-native parity, allocation-event acceptance, cancellation calibration, startup and expanded WASI-world conformance remain open. No milestone or release is completed by this study.

The first complete-candidate bootstrap used the correct source and explicit version but inherited the enclosing VM repository's Go VCS metadata. It was preserved and never used in a probe or timing. An isolated clean compiler checkout corrected VCS identity before all execution; the Cgo-only compiler starts with that isolation. Build/provenance records preserve this correction.

[Lossless plans, compiler/source/binary/SDK identities, exact harnesses, raw attempts and orders, semantic reports, separate phases, controls and verification](evidence/performance-llgo-cgo-upgrade-2026-11-04.json.xz).

Evidence archive: **379 files**, **1445480 bytes**, SHA-256 `5f00dfde9cf5e96b82771b7f76288ba8415bef533bf1fdcc51398734eb846ffb`. Every decoded file was round-trip checked; compiler/runner executables, SDK copies and source checkouts are excluded.
