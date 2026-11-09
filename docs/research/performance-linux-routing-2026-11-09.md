# Linux runtime snapshot and isolated Threads completion

Measured November 9, 2026. [Machine record](data/performance-linux-routing-2026-11-09.json).

**Both source roles use LLGo 3d92a58 with identical production VM files. This is an independent new-pin competitor baseline, not an old/new compiler A/B.** Head 19fe83a / base 5762996 differ in the P2 hermetic test only. Go 1.27.3, LLVM 22.1.8, LLGo `-O3 -lto=full`, pinned SDKs; original 10 own/reference executables were freshly built, then reused byte-for-byte by the groups. Exact source files, common instrumentation, all 10 executable hashes, 3 SDK hashes and bundled host-library hashes verify. Compiler revision/profile/hash and builder ldd are recorded; compiler binaries are not supplied for a local redigest, and builder ldd is not a new measurement-host loader probe.

Hosted EPYC 7763, inherited single-CPU affinity and GOMAXPROCS=1; each group retains its actual host metadata. No cross-host reduction or physical-isolation claim. Five fresh blocks, 2 warmups / 7 randomized interleaved calls and four independent source-identical A/A pairs per case. Allocation uses process means including every pause; other workloads use process medians. No phase subtraction.

Original 22: 1,810 runtime +55 native-C process records, 362 alternative checks, 12 actual unsupported GC outcomes; all cases complete. Sustained GC: 280 runtime records, 56 checks, 12 unsupported outcomes; all 4 cases complete. All 4 A/A controls resolve in 8/22 and 2/4 cases respectively. Five sustained Threads cases also complete (425 runtime records); its final zero-timeout case retains 68 successful rows/four blocks when the 10-minute job is cancelled. The cancelled original event and its partial rows remain unchanged. A separate successful 9m59s single-case completion now adds 85 runtime records and 17 checks, all independently audited. It does not pool the earlier four partial blocks.

### Complete original22 and sustained GC / completed Threads cases

Primary LLGo-native ns/source iteration; ratios>1 are slower. Native C is present only where semantically applicable. Every adverse row is retained.

| Cohort / workload | LLGo native ns/unit | LLGo / Go | LLGo / Wasmtime | LLGo / native C | All 4 A/A |
| --- | ---: | ---: | ---: | ---: | --- |
| original22 / core/call-control | 2313.00000 | 1.736 | 3.608 | N/A | no |
| original22 / core/integer-loop | 1.42029 | 1.006 | 1.005 | 1.009 | yes |
| original22 / core/function-calls | 0.71814 | 0.973 | 0.428 | 0.812 | yes |
| original22 / core/memory-load-store | 0.71523 | 0.987 | 1.045 | 1.424 | no |
| original22 / memory/narrow-unaligned | 1.72757 | 1.000 | 1.008 | 1.592 | yes |
| original22 / memory/dependent-address | 2.86584 | 0.985 | 1.297 | 1.517 | no |
| original22 / memory/float64-unaligned | 1.01152 | 0.939 | 1.044 | 1.065 | yes |
| original22 / runtime/simd-memory-full | 7.72790 | 0.997 | 1.066 | 0.890 | yes |
| original22 / runtime/simd-memory-widen | 0.97253 | 0.998 | 1.155 | 0.298 | yes |
| original22 / runtime/simd-memory-lanes | 9.30787 | 0.994 | 1.128 | 1.311 | yes |
| original22 / gc/i31-roundtrip | 1.99570 | 0.980 | 0.857 | N/A | no |
| original22 / gc/struct-access | 2.16800 | 0.943 | 1.486 | N/A | no |
| original22 / gc/struct-allocation | 76.22974 | 1.301 | 0.486 | N/A | no |
| original22 / gc/struct-indirect-access | 4.94330 | 0.973 | 1.003 | N/A | no |
| original22 / threads/private-memory | 1.17320 | 0.860 | 1.926 | N/A | no |
| original22 / threads/shared-memory | 1.20120 | 0.824 | 1.937 | N/A | no |
| original22 / threads/atomic-rmw | 3.06080 | 0.897 | 1.072 | N/A | no |
| original22 / threads/notify-empty | 2.45660 | 0.958 | 0.261 | N/A | no |
| original22 / threads/wait-mismatch | 2.94150 | 0.806 | 0.216 | N/A | no |
| original22 / threads/wait-timeout | 3.30320 | 0.816 | 0.043 | N/A | no |
| original22 / applications/sha256-clang | 420.82100 | 0.993 | 1.239 | 1.486 | no |
| original22 / applications/sha256-rust | 439.08900 | 0.999 | 1.310 | 1.558 | yes |
| sustained-gc / gc/i31-roundtrip | 1.58405 | 0.992 | 0.720 | N/A | yes |
| sustained-gc / gc/struct-access | 1.59217 | 0.997 | 1.253 | N/A | yes |
| sustained-gc / gc/struct-allocation | 79.71678 | 1.092 | 0.700 | N/A | no |
| sustained-gc / gc/struct-indirect-access | 4.43641 | 0.990 | 0.936 | N/A | no |
| sustained-threads (5 complete) / threads/private-memory | 0.69760 | 1.030 | 2.124 | N/A | no |
| sustained-threads (5 complete) / threads/shared-memory | 0.67577 | 0.987 | 2.045 | N/A | no |
| sustained-threads (5 complete) / threads/atomic-rmw | 2.37235 | 0.970 | 0.990 | N/A | no |
| sustained-threads (5 complete) / threads/notify-empty | 1.98231 | 1.031 | 0.220 | N/A | no |
| sustained-threads (5 complete) / threads/wait-mismatch | 2.06377 | 0.961 | 0.166 | N/A | no |

| isolated wait-zero / threads/wait-timeout | 2.35839 | 1.006 | 0.032 | N/A | no |

### Every source-identical A/A interval

Paired-block log-ratio bootstrap 95% interval; each of the four intervals must fit inside [0.975, 1.025]. Resolved controls do not by themselves establish a cross-runtime or source speedup.

| Cohort / workload | Control | Lower | Upper | Resolved |
| --- | --- | ---: | ---: | --- |
| original22 / core/call-control | aa-base-llgo-interpreter | 0.86250 | 1.05959 | no |
| original22 / core/call-control | aa-base-llgo-native | 0.99371 | 1.05323 | no |
| original22 / core/call-control | aa-head-llgo-interpreter | 0.93660 | 1.13508 | no |
| original22 / core/call-control | aa-head-llgo-native | 0.89231 | 0.96249 | no |
| original22 / core/integer-loop | aa-base-llgo-interpreter | 0.98347 | 1.01834 | yes |
| original22 / core/integer-loop | aa-base-llgo-native | 1.00153 | 1.01724 | yes |
| original22 / core/integer-loop | aa-head-llgo-interpreter | 0.99602 | 1.00556 | yes |
| original22 / core/integer-loop | aa-head-llgo-native | 0.99340 | 1.00719 | yes |
| original22 / core/function-calls | aa-base-llgo-interpreter | 0.99372 | 1.01237 | yes |
| original22 / core/function-calls | aa-base-llgo-native | 0.98897 | 1.00785 | yes |
| original22 / core/function-calls | aa-head-llgo-interpreter | 0.98575 | 1.02394 | yes |
| original22 / core/function-calls | aa-head-llgo-native | 0.99476 | 1.01255 | yes |
| original22 / core/memory-load-store | aa-base-llgo-interpreter | 0.97922 | 1.03007 | no |
| original22 / core/memory-load-store | aa-base-llgo-native | 0.99078 | 1.00485 | yes |
| original22 / core/memory-load-store | aa-head-llgo-interpreter | 0.98183 | 1.00122 | yes |
| original22 / core/memory-load-store | aa-head-llgo-native | 0.99515 | 1.00334 | yes |
| original22 / memory/narrow-unaligned | aa-base-llgo-interpreter | 0.99288 | 1.00115 | yes |
| original22 / memory/narrow-unaligned | aa-base-llgo-native | 0.99674 | 1.01193 | yes |
| original22 / memory/narrow-unaligned | aa-head-llgo-interpreter | 0.99929 | 1.00258 | yes |
| original22 / memory/narrow-unaligned | aa-head-llgo-native | 0.98100 | 1.00400 | yes |
| original22 / memory/dependent-address | aa-base-llgo-interpreter | 0.99321 | 1.12166 | no |
| original22 / memory/dependent-address | aa-base-llgo-native | 0.99544 | 1.00697 | yes |
| original22 / memory/dependent-address | aa-head-llgo-interpreter | 0.99806 | 1.01814 | yes |
| original22 / memory/dependent-address | aa-head-llgo-native | 0.99596 | 1.00338 | yes |
| original22 / memory/float64-unaligned | aa-base-llgo-interpreter | 0.98670 | 1.00403 | yes |
| original22 / memory/float64-unaligned | aa-base-llgo-native | 0.98106 | 1.01528 | yes |
| original22 / memory/float64-unaligned | aa-head-llgo-interpreter | 0.98839 | 1.02017 | yes |
| original22 / memory/float64-unaligned | aa-head-llgo-native | 0.98621 | 1.00393 | yes |
| original22 / runtime/simd-memory-full | aa-base-llgo-interpreter | 1.00020 | 1.00805 | yes |
| original22 / runtime/simd-memory-full | aa-base-llgo-native | 0.99526 | 1.00303 | yes |
| original22 / runtime/simd-memory-full | aa-head-llgo-interpreter | 0.98341 | 1.00498 | yes |
| original22 / runtime/simd-memory-full | aa-head-llgo-native | 0.99762 | 0.99971 | yes |
| original22 / runtime/simd-memory-widen | aa-base-llgo-interpreter | 0.99696 | 1.00142 | yes |
| original22 / runtime/simd-memory-widen | aa-base-llgo-native | 1.00174 | 1.00551 | yes |
| original22 / runtime/simd-memory-widen | aa-head-llgo-interpreter | 0.99502 | 1.00596 | yes |
| original22 / runtime/simd-memory-widen | aa-head-llgo-native | 0.98903 | 1.00757 | yes |
| original22 / runtime/simd-memory-lanes | aa-base-llgo-interpreter | 0.99153 | 0.99841 | yes |
| original22 / runtime/simd-memory-lanes | aa-base-llgo-native | 0.99800 | 1.00311 | yes |
| original22 / runtime/simd-memory-lanes | aa-head-llgo-interpreter | 0.99812 | 1.00289 | yes |
| original22 / runtime/simd-memory-lanes | aa-head-llgo-native | 0.99964 | 1.00208 | yes |
| original22 / gc/i31-roundtrip | aa-base-llgo-interpreter | 0.98175 | 0.99271 | yes |
| original22 / gc/i31-roundtrip | aa-base-llgo-native | 0.94929 | 1.00983 | no |
| original22 / gc/i31-roundtrip | aa-head-llgo-interpreter | 0.99337 | 1.00592 | yes |
| original22 / gc/i31-roundtrip | aa-head-llgo-native | 0.97682 | 1.01067 | yes |
| original22 / gc/struct-access | aa-base-llgo-interpreter | 0.97976 | 1.01117 | yes |
| original22 / gc/struct-access | aa-base-llgo-native | 0.95457 | 1.05130 | no |
| original22 / gc/struct-access | aa-head-llgo-interpreter | 0.98681 | 1.04145 | no |
| original22 / gc/struct-access | aa-head-llgo-native | 0.98456 | 1.02919 | no |
| original22 / gc/struct-allocation | aa-base-llgo-interpreter | 0.96447 | 1.02485 | no |
| original22 / gc/struct-allocation | aa-base-llgo-native | 0.95277 | 1.03810 | no |
| original22 / gc/struct-allocation | aa-head-llgo-interpreter | 0.96881 | 1.04282 | no |
| original22 / gc/struct-allocation | aa-head-llgo-native | 0.93664 | 1.04239 | no |
| original22 / gc/struct-indirect-access | aa-base-llgo-interpreter | 0.99019 | 1.01009 | yes |
| original22 / gc/struct-indirect-access | aa-base-llgo-native | 0.95849 | 1.03397 | no |
| original22 / gc/struct-indirect-access | aa-head-llgo-interpreter | 0.99030 | 1.01292 | yes |
| original22 / gc/struct-indirect-access | aa-head-llgo-native | 0.98703 | 1.08986 | no |
| original22 / threads/private-memory | aa-base-llgo-interpreter | 0.73033 | 1.01070 | no |
| original22 / threads/private-memory | aa-base-llgo-native | 1.03559 | 1.13082 | no |
| original22 / threads/private-memory | aa-head-llgo-interpreter | 0.98457 | 1.24116 | no |
| original22 / threads/private-memory | aa-head-llgo-native | 0.90722 | 1.03456 | no |
| original22 / threads/shared-memory | aa-base-llgo-interpreter | 0.98565 | 1.00184 | yes |
| original22 / threads/shared-memory | aa-base-llgo-native | 0.95754 | 1.04330 | no |
| original22 / threads/shared-memory | aa-head-llgo-interpreter | 0.99202 | 1.00626 | yes |
| original22 / threads/shared-memory | aa-head-llgo-native | 0.95917 | 1.08842 | no |
| original22 / threads/atomic-rmw | aa-base-llgo-interpreter | 0.91836 | 1.01643 | no |
| original22 / threads/atomic-rmw | aa-base-llgo-native | 0.98660 | 1.11424 | no |
| original22 / threads/atomic-rmw | aa-head-llgo-interpreter | 0.96805 | 1.01470 | no |
| original22 / threads/atomic-rmw | aa-head-llgo-native | 0.99537 | 1.02413 | yes |
| original22 / threads/notify-empty | aa-base-llgo-interpreter | 0.97022 | 1.00114 | no |
| original22 / threads/notify-empty | aa-base-llgo-native | 0.94295 | 0.99959 | no |
| original22 / threads/notify-empty | aa-head-llgo-interpreter | 0.98328 | 1.00887 | yes |
| original22 / threads/notify-empty | aa-head-llgo-native | 0.89004 | 1.18025 | no |
| original22 / threads/wait-mismatch | aa-base-llgo-interpreter | 0.99448 | 1.52926 | no |
| original22 / threads/wait-mismatch | aa-base-llgo-native | 0.97728 | 1.28186 | no |
| original22 / threads/wait-mismatch | aa-head-llgo-interpreter | 1.00022 | 1.46069 | no |
| original22 / threads/wait-mismatch | aa-head-llgo-native | 0.94430 | 1.41634 | no |
| original22 / threads/wait-timeout | aa-base-llgo-interpreter | 0.94699 | 1.02857 | no |
| original22 / threads/wait-timeout | aa-base-llgo-native | 0.88189 | 0.98335 | no |
| original22 / threads/wait-timeout | aa-head-llgo-interpreter | 0.98045 | 1.03067 | no |
| original22 / threads/wait-timeout | aa-head-llgo-native | 0.96947 | 1.07065 | no |
| original22 / applications/sha256-clang | aa-base-llgo-interpreter | 0.99906 | 1.01831 | yes |
| original22 / applications/sha256-clang | aa-base-llgo-native | 0.99781 | 1.00501 | yes |
| original22 / applications/sha256-clang | aa-head-llgo-interpreter | 0.97363 | 0.99989 | no |
| original22 / applications/sha256-clang | aa-head-llgo-native | 0.99855 | 1.00922 | yes |
| original22 / applications/sha256-rust | aa-base-llgo-interpreter | 0.98632 | 1.01917 | yes |
| original22 / applications/sha256-rust | aa-base-llgo-native | 0.99858 | 1.01526 | yes |
| original22 / applications/sha256-rust | aa-head-llgo-interpreter | 0.99153 | 1.01073 | yes |
| original22 / applications/sha256-rust | aa-head-llgo-native | 0.99206 | 1.00296 | yes |
| sustained-gc / gc/i31-roundtrip | aa-base-llgo-interpreter | 0.99553 | 1.00606 | yes |
| sustained-gc / gc/i31-roundtrip | aa-base-llgo-native | 0.99918 | 1.00901 | yes |
| sustained-gc / gc/i31-roundtrip | aa-head-llgo-interpreter | 0.99792 | 1.00607 | yes |
| sustained-gc / gc/i31-roundtrip | aa-head-llgo-native | 0.99450 | 1.00337 | yes |
| sustained-gc / gc/struct-access | aa-base-llgo-interpreter | 0.99659 | 1.00199 | yes |
| sustained-gc / gc/struct-access | aa-base-llgo-native | 0.99542 | 1.00701 | yes |
| sustained-gc / gc/struct-access | aa-head-llgo-interpreter | 0.99829 | 1.00497 | yes |
| sustained-gc / gc/struct-access | aa-head-llgo-native | 1.00002 | 1.00738 | yes |
| sustained-gc / gc/struct-allocation | aa-base-llgo-interpreter | 0.99779 | 1.02651 | no |
| sustained-gc / gc/struct-allocation | aa-base-llgo-native | 0.97646 | 1.00381 | yes |
| sustained-gc / gc/struct-allocation | aa-head-llgo-interpreter | 0.99337 | 1.01896 | yes |
| sustained-gc / gc/struct-allocation | aa-head-llgo-native | 1.00311 | 1.00870 | yes |
| sustained-gc / gc/struct-indirect-access | aa-base-llgo-interpreter | 0.99173 | 1.00028 | yes |
| sustained-gc / gc/struct-indirect-access | aa-base-llgo-native | 0.99861 | 1.00876 | yes |
| sustained-gc / gc/struct-indirect-access | aa-head-llgo-interpreter | 1.00005 | 1.04591 | no |
| sustained-gc / gc/struct-indirect-access | aa-head-llgo-native | 0.99648 | 1.00483 | yes |
| sustained-threads / threads/private-memory | aa-base-llgo-interpreter | 0.99712 | 1.00869 | yes |
| sustained-threads / threads/private-memory | aa-base-llgo-native | 0.99801 | 1.03240 | no |
| sustained-threads / threads/private-memory | aa-head-llgo-interpreter | 0.99256 | 1.00610 | yes |
| sustained-threads / threads/private-memory | aa-head-llgo-native | 0.91502 | 0.99205 | no |
| sustained-threads / threads/shared-memory | aa-base-llgo-interpreter | 0.99587 | 1.00904 | yes |
| sustained-threads / threads/shared-memory | aa-base-llgo-native | 0.96269 | 1.02162 | no |
| sustained-threads / threads/shared-memory | aa-head-llgo-interpreter | 0.99167 | 1.00048 | yes |
| sustained-threads / threads/shared-memory | aa-head-llgo-native | 0.99784 | 1.03631 | no |
| sustained-threads / threads/atomic-rmw | aa-base-llgo-interpreter | 0.99498 | 1.05428 | no |
| sustained-threads / threads/atomic-rmw | aa-base-llgo-native | 0.96210 | 1.01862 | no |
| sustained-threads / threads/atomic-rmw | aa-head-llgo-interpreter | 1.00002 | 1.00416 | yes |
| sustained-threads / threads/atomic-rmw | aa-head-llgo-native | 0.96830 | 1.04391 | no |
| sustained-threads / threads/notify-empty | aa-base-llgo-interpreter | 0.99127 | 1.00126 | yes |
| sustained-threads / threads/notify-empty | aa-base-llgo-native | 0.98562 | 1.02046 | yes |
| sustained-threads / threads/notify-empty | aa-head-llgo-interpreter | 0.99688 | 0.99986 | yes |
| sustained-threads / threads/notify-empty | aa-head-llgo-native | 0.97283 | 1.02024 | no |
| sustained-threads / threads/wait-mismatch | aa-base-llgo-interpreter | 0.99449 | 1.00914 | yes |
| sustained-threads / threads/wait-mismatch | aa-base-llgo-native | 0.98566 | 1.03460 | no |
| sustained-threads / threads/wait-mismatch | aa-head-llgo-interpreter | 0.99274 | 1.00166 | yes |
| sustained-threads / threads/wait-mismatch | aa-head-llgo-native | 0.97076 | 1.07630 | no |

| isolated wait-zero / threads/wait-timeout | aa-base-llgo-interpreter | 0.99777 | 1.00618 | yes |
| isolated wait-zero / threads/wait-timeout | aa-base-llgo-native | 0.90098 | 1.01833 | no |
| isolated wait-zero / threads/wait-timeout | aa-head-llgo-interpreter | 0.99781 | 1.00394 | yes |
| isolated wait-zero / threads/wait-timeout | aa-head-llgo-native | 0.98326 | 1.08191 | no |

The zero-timeout fixture uses literal timeout 0. The pinned WasmEdge 0.17.1 interpreter still creates a waiter-map entry and enters a condition-variable wait-until path. The original cancelled job and all 68 partial rows remain retained. Its [isolated completion](https://github.com/cpunion/wasmvm/actions/runs/42636602225) succeeded in 9m59s with the same bytes, 100,000 iterations, five blocks, profiles and statistic. This is a separate host cohort with an explicit 15-minute diagnostic budget; ordinary semantic CI remains 10 minutes. All 85 measured rows, 17 alternative checks and four controls were revalidated. Native controls remain unresolved. No completed case was rerun and no partial observations were pooled. Archive SHA256 `a01a742b15650d800918616d588022be1f49e0c876529b27f63cea40e04f78d9` is identical to the original fresh build. Artifact 13686630402 ZIP SHA256 `8461164bcef5fd39cd4b382c8d881de24de1ed6ef36b40389afac509980c1417` verifies.

Raw completed cases retain all named competitors, full lifecycle phases, process intervals, checks, source/toolchain/binary/SDK identities and unsupported/adverse outcomes. Retained: `artifacts/llgo-escape-linux-20261109`. The M4 compiler-upgrade and concurrent-service events remain independent. Controlled/native-C/memory/SIMD/GC/Threads acceptance remains open; this snapshot establishes no broad source/compiler win.
