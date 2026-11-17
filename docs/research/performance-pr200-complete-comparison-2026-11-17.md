# PR #200: complete runtime comparison, 2026-11-17

This event measures PR #200 against its own source baseline and four independent runtimes. It adds WasmEdge O3 AOT alongside its interpreter. All **33 declared core workloads** and **two complete P3 HTTP payloads** are reported below, with separate wasmvm Go and LLGo columns. Competitive parity remains open. General native calls, SHA, GC field access, shared memory and P3 still show substantial observed gaps; unresolved controls limit precise rankings.

Source CPU improvements with resolving controls in this event: **Go 64 KiB P3 −7.61%**, **LLGo 16 B −4.62%**, and **LLGo 64 KiB −2.90%**. These are new independent measurements, not pooled with the original PR #200 event. The Go 16 B CPU control is unresolved. Against Wasmtime, both P3 payloads have unresolved controls; the observed CPU ratios are approximately **1.76–2.05× for Go** and **4.89–4.99× for LLGo**, not accepted exact competitive ratios.

## Revisions, profiles and protocol

- Engine source: base `b3925191`, head/merged PR #200 `c8936b9`. Harness/policy head `34c3a5e0` has identical engine inputs. Reused binaries retain their actual producer revisions; `build.json` records verified production-source equivalence to the qualified source. No full source copies or new worktrees were created.
- Go **1.27.3**; LLGo **114c98b4**, **`-O3 -lto=full`**, LLVM **22.1.8**, default production GC. Source base/head share the compiler profile. LLGo **2e3e02f4** on the same source baseline measures the SDK upgrade independently.
- wazero **1.12.0**, native compiler/interpreter; Wasmtime **48.0.1**, Cranelift speed; Wasmer **7.3.0**, Cranelift; WasmEdge **0.17.1**, interpreter and **O3 AOT**. These are explicit pinned profiles, not a claim that every available backend or configuration was measured.
- Local Apple M4 Max, macOS arm64, 128 GiB RAM. External desktop load is uncontrolled and recorded; no owned builds, tests, profiles, analysis or edits overlap timing. This is not a Linux amd64 performance result.
- Core: six fixed fresh-process blocks, each with symmetric same-binary A/A, two warmups and seven interleaved measured calls. **6,156 successful process records** after excluding exactly three failing sustained WasmEdge AOT cells. All 33 workloads remain represented. Allocation workloads use means including pauses; others use medians.
- P3: **84 processes per payload**, seven roles with A/A, six fixed blocks, four workers, sixteen clients and twenty interleaved 512-request windows plus a separate flush. Verify status, headers, full body, trailers, CRC and joined guest completion. CPU/instruction counters include the driver and checks. Mach ticks use the recorded timebase.
- Ratios use paired block geometric means and 20,000 fixed-seed bootstrap resamples with marginal 95% intervals. Both relevant A/A intervals must fit `[0.975, 1.025]`. No simultaneous portfolio-acceptance claim, adaptive extension, phase subtraction or cross-event pooling.

**Legend:** `†` means that role’s same-binary timing control is unresolved. Values are observations; a qualified comparison additionally needs both controls and its ratio interval. `US` means this pinned profile rejected the feature in semantic preflight. `NM` means this benchmark adapter/profile was not measured; it does not claim the engine lacks the capability. `N/A` means the metric does not apply. `FAIL` is an execution failure, never an unsupported-feature classification or performance win.

## Native compiled execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 3,687.00 † | 2,666.50 † | 479.00 † | 1,395.50 † | 2,791.50 † | 3,958.50 † |
| original22-v1/core/integer-loop | 1.751 | 1.763 | 2.259 † | 1.764 † | 1.731 † | 1.781 † |
| original22-v1/core/function-calls | 1.014 † | 0.982 † | 1.882 † | 1.146 † | 1.001 † | 0.998 † |
| original22-v1/core/memory-load-store | 0.488 † | 0.486 † | 1.115 † | 0.535 † | 0.605 † | 0.328 † |
| original22-v1/memory/narrow-unaligned | 0.711 † | 0.704 | 1.189 † | 0.770 † | 0.784 † | 0.569 † |
| original22-v1/memory/dependent-address | 2.374 † | 2.360 † | 4.158 † | 2.724 † | 2.710 | 1.820 † |
| original22-v1/memory/float64-unaligned | 1.004 † | 0.889 † | 1.226 † | 0.868 † | 0.878 † | 0.802 † |
| original22-v1/runtime/simd-memory-full | 4.047 † | 4.007 † | 4.785 † | 4.472 † | 4.230 † | 4.539 † |
| original22-v1/runtime/simd-memory-widen | 1.260 † | 1.392 † | 1.428 † | 1.313 † | 1.239 † | 1.266 † |
| original22-v1/runtime/simd-memory-lanes | 5.807 † | 6.106 † | 6.892 † | 6.834 † | 6.800 † | 5.729 † |
| original22-v1/gc/i31-roundtrip | 2.773 † | 2.290 † | US | 2.527 † | US | 2.448 † |
| original22-v1/gc/struct-access | 3.165 † | 2.273 † | US | 1.552 † | US | 10.835 |
| original22-v1/gc/struct-allocation | 37.331 † | 34.543 † | US | 84.063 † | US | 52.188 † |
| original22-v1/gc/struct-indirect-access | 3.615 † | 2.602 † | US | 3.317 † | US | 30.656 † |
| original22-v1/threads/private-memory | 2.106 † | 1.104 † | 2.581 † | 1.831 † | 2.163 † | 2.200 † |
| original22-v1/threads/shared-memory | 2.917 † | 2.173 † | 1.915 † | 1.673 † | 2.025 † | 2.027 † |
| original22-v1/threads/atomic-rmw | 3.558 † | 2.735 † | 2.877 † | 2.333 † | 5.177 † | 2.800 † |
| original22-v1/threads/notify-empty | 2.852 † | 1.796 † | 26.610 † | 8.796 † | 6.796 † | 11.535 † |
| original22-v1/threads/wait-mismatch | 3.333 † | 2.033 † | 33.487 † | 11.450 † | 3.638 † | 4.658 † |
| original22-v1/threads/wait-timeout | 3.627 † | 2.021 † | 395.06 † | 58.313 † | 230.17 † | 82.685 † |
| original22-v1/applications/sha256-clang | 295.77 † | 284.00 † | 378.46 † | 250.89 † | 236.62 † | 225.75 † |
| original22-v1/applications/sha256-rust | 292.79 † | 273.56 † | 420.73 † | 229.91 † | 237.45 † | 228.89 † |
| gc-sustained-v1/gc/i31-roundtrip | 1.568 † | 1.512 † | US | 2.064 | US | 1.476 † |
| gc-sustained-v1/gc/struct-access | 1.560 † | 1.538 † | US | 1.228 † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 42.464 † | 38.518 † | US | 68.555 † | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 2.266 † | 2.176 † | US | 3.379 † | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.552 † | 0.410 † | 1.657 † | 1.309 † | 1.428 † | 1.388 † |
| threads-sustained-v2/threads/shared-memory | 1.809 † | 1.629 † | 1.697 † | 1.353 † | 1.381 † | 1.385 † |
| threads-sustained-v2/threads/atomic-rmw | 2.438 † | 2.311 | 2.607 † | 2.141 † | 4.706 † | 2.228 † |
| threads-sustained-v2/threads/notify-empty | 1.238 † | 1.072 † | 23.190 † | 7.490 | 5.194 † | 9.857 † |
| threads-sustained-v2/threads/wait-mismatch | 1.608 † | 1.445 † | 32.511 † | 11.601 † | 3.184 † | 4.436 † |
| threads-sustained-v2/threads/wait-timeout | 1.359 † | 1.138 † | 345.18 † | 56.500 † | 212.34 † | 74.644 † |
| generic-calls-v1/core/generic-calls | 4.635 | 4.795 † | 2.622 † | 2.550 † | 2.518 † | 1.237 † |

## Interpreter execution

Warm **ns/source iteration**, except SHA uses **ns/message block**. Cohorts distinguish the original 22 cases from four sustained GC, six sustained Threads, and the general-call workload.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2,479.00 † | 2,396.00 † | 1,083.50 † | NM | NM | 3,604.50 † |
| original22-v1/core/integer-loop | 35.381 † | 28.705 † | 92.504 † | NM | NM | 56.321 † |
| original22-v1/core/function-calls | 62.380 † | 54.190 † | 111.67 † | NM | NM | 80.868 † |
| original22-v1/core/memory-load-store | 73.036 † | 64.407 † | 170.89 † | NM | NM | 101.59 † |
| original22-v1/memory/narrow-unaligned | 94.359 † | 81.710 † | 218.79 † | NM | NM | 124.16 † |
| original22-v1/memory/dependent-address | 54.281 | 45.352 † | 128.91 † | NM | NM | 69.977 † |
| original22-v1/memory/float64-unaligned | 86.787 † | 75.366 † | 169.36 † | NM | NM | 107.19 † |
| original22-v1/runtime/simd-memory-full | 91.473 † | 79.543 | 193.18 † | NM | NM | 95.891 † |
| original22-v1/runtime/simd-memory-widen | 73.780 † | 89.954 † | 180.56 † | NM | NM | 76.909 † |
| original22-v1/runtime/simd-memory-lanes | 105.27 † | 102.69 † | 237.38 † | NM | NM | 110.35 † |
| original22-v1/gc/i31-roundtrip | 45.848 † | 38.108 † | US | NM | NM | 54.865 † |
| original22-v1/gc/struct-access | 45.996 † | 42.621 † | US | NM | NM | 61.377 † |
| original22-v1/gc/struct-allocation | 119.82 † | 124.65 † | US | NM | NM | 136.43 † |
| original22-v1/gc/struct-indirect-access | 87.502 † | 78.310 † | US | NM | NM | 108.84 |
| original22-v1/threads/private-memory | 52.685 † | 46.379 † | 80.063 † | NM | NM | 51.742 † |
| original22-v1/threads/shared-memory | 95.600 | 42.104 † | 78.856 † | NM | NM | 48.810 † |
| original22-v1/threads/atomic-rmw | 58.708 † | 38.681 † | 52.067 † | NM | NM | 37.496 † |
| original22-v1/threads/notify-empty | 38.829 † | 34.819 † | 51.508 † | NM | NM | 49.669 † |
| original22-v1/threads/wait-mismatch | 74.212 † | 46.077 † | 87.931 † | NM | NM | 57.013 † |
| original22-v1/threads/wait-timeout | 75.171 † | 48.529 | 462.04 † | NM | NM | 132.53 † |
| original22-v1/applications/sha256-clang | 23,950.33 † | 20,251.95 † | 51,130.76 † | NM | NM | 33,442.31 † |
| original22-v1/applications/sha256-rust | 22,331.42 † | 20,704.81 † | 51,210.94 † | NM | NM | 31,640.35 † |
| gc-sustained-v1/gc/i31-roundtrip | 40.811 † | 35.679 † | US | NM | NM | 51.456 † |
| gc-sustained-v1/gc/struct-access | 49.729 † | 47.603 | US | NM | NM | 68.386 † |
| gc-sustained-v1/gc/struct-allocation | 130.78 † | 131.24 † | US | NM | NM | 141.17 † |
| gc-sustained-v1/gc/struct-indirect-access | 103.85 † | 89.409 † | US | NM | NM | 131.42 † |
| threads-sustained-v2/threads/private-memory | 53.308 † | 44.188 † | 79.598 † | NM | NM | 50.796 † |
| threads-sustained-v2/threads/shared-memory | 97.708 † | 43.889 † | 81.595 † | NM | NM | 49.973 † |
| threads-sustained-v2/threads/atomic-rmw | 59.450 † | 38.401 † | 52.746 † | NM | NM | 38.214 |
| threads-sustained-v2/threads/notify-empty | 34.695 † | 30.601 † | 46.388 † | NM | NM | 44.929 † |
| threads-sustained-v2/threads/wait-mismatch | 77.362 † | 48.860 † | 91.511 † | NM | NM | 56.512 † |
| threads-sustained-v2/threads/wait-timeout | 65.655 † | 41.910 † | 418.78 † | NM | NM | 124.93 † |
| generic-calls-v1/core/generic-calls | 133.65 † | 113.13 † | 225.43 † | NM | NM | 151.74 † |

## Complete P3 HTTP requests

P3 competitor cells marked `NM` have no measured adapter for the pinned component/WASI workload. Wasmtime uses the same validated workload. Managed allocation counts exclude foreign Rust/C/JIT allocations and cannot compare total allocation across engines.

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU, us/request | 75.250 † | 206.22 | NM | 42.356 † | NM | NM |
| 16 B instructions/request | 711,818.87 | 1,302,517.96 | NM | 229,405.72 † | NM | NM |
| 16 B requests/s | 49,868.76 † | 21,968.38 | NM | 51,141.53 † | NM | NM |
| 65536 B CPU, us/request | 95.991 | 236.39 | NM | 46.094 † | NM | NM |
| 65536 B instructions/request | 991,040.79 | 1,634,625.45 | NM | 314,754.39 | NM | NM |
| 65536 B requests/s | 39,526.23 † | 19,625.73 † | NM | 54,888.34 † | NM | NM |

### P3 source and SDK contrasts

| Metric | Go source head/base | LLGo source head/base | LLGo SDK114/SDK2e on fixed base |
| --- | ---: | ---: | ---: |
| 16 B cpuUs | 0.920 [0.910, 0.930]; unresolved | 0.954 [0.947, 0.958]; pass | 1.002 [0.997, 1.009]; pass |
| 16 B instructions | 0.913 [0.911, 0.915]; pass | 0.941 [0.938, 0.944]; pass | 1.001 [0.998, 1.004]; pass |
| 16 B throughput | 1.089 [1.064, 1.111]; unresolved | 1.056 [1.039, 1.074]; unresolved | 1.007 [0.989, 1.025]; unresolved |
| 16 B allocsPerRequest | 0.948 [0.948, 0.948]; pass | 0.949 [0.949, 0.949]; pass | 1.000 [1.000, 1.000]; pass |
| 16 B bytesPerRequest | 0.958 [0.958, 0.958]; pass | 0.956 [0.956, 0.956]; pass | 1.000 [1.000, 1.000]; pass |
| 65536 B cpuUs | 0.924 [0.909, 0.940]; pass | 0.971 [0.960, 0.986]; pass | 0.998 [0.988, 1.008]; unresolved |
| 65536 B instructions | 0.933 [0.929, 0.937]; pass | 0.957 [0.951, 0.965]; pass | 1.000 [0.995, 1.005]; pass |
| 65536 B throughput | 1.089 [1.050, 1.120]; unresolved | 1.030 [1.000, 1.056]; unresolved | 0.996 [0.978, 1.013]; unresolved |
| 65536 B allocsPerRequest | 0.952 [0.952, 0.952]; pass | 0.954 [0.954, 0.954]; pass | 1.000 [1.000, 1.000]; pass |
| 65536 B bytesPerRequest | 0.983 [0.983, 0.983]; pass | 0.981 [0.981, 0.981]; pass | 1.000 [1.000, 1.000]; pass |

Throughput ratios above one improve; CPU/instructions/allocation ratios below one improve. The JSON also retains the combined PR source+SDK contrast, all own-baseline observations and every per-role A/A interval.

### Competitive P3 ratios

| Payload / compiler | CPU ratio to Wasmtime [95% CI]; controls | Throughput ratio to Wasmtime [95% CI]; controls |
| --- | ---: | ---: |
| 16 B go | 1.762 [1.694, 1.827]; unresolved | 0.971 [0.949, 0.993]; unresolved |
| 16 B llgo | 4.890 [4.721, 5.041]; unresolved | 0.428 [0.423, 0.433]; unresolved |
| 65536 B go | 2.050 [1.976, 2.143]; unresolved | 0.729 [0.696, 0.767]; unresolved |
| 65536 B llgo | 4.987 [4.638, 5.296]; unresolved | 0.369 [0.351, 0.388]; unresolved |

## Core source baseline and SDK contrasts

Head/base ratios below one improve. These marginal intervals describe this event; an apparent change in an unchanged mechanism requires diagnosis before being attributed to source changes. The machine-readable report includes every competitor ratio and A/A interval.

### native

| Workload | Go source head/base | LLGo source head/base | LLGo SDK114/SDK2e on fixed base |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.899 [0.692, 1.062]; unresolved | 0.964 [0.894, 1.047]; unresolved | 0.995 [0.888, 1.105]; unresolved |
| original22-v1/core/integer-loop | 0.992 [0.981, 1.005]; unresolved | 1.003 [0.995, 1.011]; pass | 0.965 [0.945, 0.983]; unresolved |
| original22-v1/core/function-calls | 1.007 [0.970, 1.037]; unresolved | 0.981 [0.919, 1.030]; unresolved | 0.998 [0.930, 1.080]; unresolved |
| original22-v1/core/memory-load-store | 0.995 [0.970, 1.018]; unresolved | 0.975 [0.942, 1.011]; unresolved | 1.025 [0.983, 1.076]; unresolved |
| original22-v1/memory/narrow-unaligned | 1.006 [0.969, 1.047]; unresolved | 0.990 [0.954, 1.023]; unresolved | 1.007 [0.995, 1.020]; unresolved |
| original22-v1/memory/dependent-address | 1.005 [0.981, 1.028]; unresolved | 0.981 [0.949, 1.004]; unresolved | 1.007 [0.978, 1.048]; unresolved |
| original22-v1/memory/float64-unaligned | 1.000 [0.966, 1.049]; unresolved | 0.992 [0.966, 1.017]; unresolved | 0.996 [0.971, 1.021]; unresolved |
| original22-v1/runtime/simd-memory-full | 0.985 [0.954, 1.013]; unresolved | 0.995 [0.981, 1.011]; unresolved | 1.016 [0.999, 1.040]; unresolved |
| original22-v1/runtime/simd-memory-widen | 0.974 [0.946, 0.996]; unresolved | 1.052 [1.006, 1.102]; unresolved | 0.990 [0.961, 1.012]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 1.010 [0.984, 1.035]; unresolved | 1.018 [0.986, 1.067]; unresolved | 0.987 [0.963, 1.008]; unresolved |
| original22-v1/gc/i31-roundtrip | 0.990 [0.914, 1.071]; unresolved | 0.967 [0.921, 1.019]; unresolved | 1.025 [0.974, 1.085]; unresolved |
| original22-v1/gc/struct-access | 1.029 [0.955, 1.116]; unresolved | 0.966 [0.937, 0.996]; unresolved | 1.013 [0.976, 1.055]; unresolved |
| original22-v1/gc/struct-allocation | 0.997 [0.953, 1.036]; unresolved | 1.004 [0.964, 1.054]; unresolved | 0.973 [0.923, 1.016]; unresolved |
| original22-v1/gc/struct-indirect-access | 1.004 [0.933, 1.075]; unresolved | 1.002 [0.960, 1.048]; unresolved | 0.961 [0.892, 1.014]; unresolved |
| original22-v1/threads/private-memory | 1.092 [0.969, 1.218]; unresolved | 1.006 [0.964, 1.051]; unresolved | 0.983 [0.930, 1.051]; unresolved |
| original22-v1/threads/shared-memory | 0.934 [0.866, 1.018]; unresolved | 0.945 [0.851, 1.047]; unresolved | 0.967 [0.825, 1.085]; unresolved |
| original22-v1/threads/atomic-rmw | 1.032 [0.970, 1.095]; unresolved | 1.016 [0.975, 1.052]; unresolved | 0.998 [0.900, 1.089]; unresolved |
| original22-v1/threads/notify-empty | 0.948 [0.895, 1.004]; unresolved | 0.968 [0.896, 1.027]; unresolved | 0.987 [0.957, 1.010]; unresolved |
| original22-v1/threads/wait-mismatch | 1.028 [0.962, 1.096]; unresolved | 1.022 [0.982, 1.070]; unresolved | 0.932 [0.837, 1.016]; unresolved |
| original22-v1/threads/wait-timeout | 1.051 [0.924, 1.210]; unresolved | 1.029 [0.992, 1.074]; unresolved | 0.952 [0.854, 1.038]; unresolved |
| original22-v1/applications/sha256-clang | 1.013 [0.959, 1.077]; unresolved | 0.986 [0.930, 1.041]; unresolved | 1.034 [0.990, 1.086]; unresolved |
| original22-v1/applications/sha256-rust | 1.029 [0.982, 1.079]; unresolved | 0.982 [0.942, 1.019]; unresolved | 0.989 [0.960, 1.021]; unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 0.977 [0.951, 0.998]; unresolved | 0.991 [0.968, 1.013]; unresolved | 1.015 [1.007, 1.023]; unresolved |
| gc-sustained-v1/gc/struct-access | 0.977 [0.952, 0.996]; unresolved | 0.987 [0.969, 1.003]; unresolved | 1.010 [0.979, 1.038]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 1.013 [0.993, 1.036]; unresolved | 1.000 [0.974, 1.029]; unresolved | 0.980 [0.897, 1.033]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 0.961 [0.910, 1.001]; unresolved | 0.946 [0.884, 0.998]; unresolved | 0.999 [0.959, 1.034]; unresolved |
| threads-sustained-v2/threads/private-memory | 0.994 [0.923, 1.065]; unresolved | 1.037 [0.973, 1.139]; unresolved | 0.981 [0.960, 0.998]; unresolved |
| threads-sustained-v2/threads/shared-memory | 0.997 [0.969, 1.023]; unresolved | 0.914 [0.786, 0.996]; unresolved | 1.035 [0.996, 1.079]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.999 [0.982, 1.015]; unresolved | 1.005 [0.976, 1.035]; unresolved | 1.003 [0.979, 1.027]; unresolved |
| threads-sustained-v2/threads/notify-empty | 1.019 [0.984, 1.060]; unresolved | 1.031 [0.994, 1.067]; unresolved | 0.949 [0.865, 1.023]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 0.973 [0.914, 1.023]; unresolved | 0.933 [0.850, 1.002]; unresolved | 1.062 [0.958, 1.207]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 0.985 [0.908, 1.067]; unresolved | 1.007 [0.968, 1.044]; unresolved | 0.974 [0.943, 0.998]; unresolved |
| generic-calls-v1/core/generic-calls | 0.971 [0.944, 0.990]; unresolved | 1.007 [0.981, 1.043]; unresolved | 1.015 [0.993, 1.034]; unresolved |

### interpreter

| Workload | Go source head/base | LLGo source head/base | LLGo SDK114/SDK2e on fixed base |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 1.019 [0.859, 1.199]; unresolved | 0.960 [0.876, 1.075]; unresolved | 1.056 [0.977, 1.147]; unresolved |
| original22-v1/core/integer-loop | 1.001 [0.981, 1.023]; unresolved | 0.985 [0.973, 0.997]; unresolved | 1.009 [0.981, 1.033]; unresolved |
| original22-v1/core/function-calls | 1.020 [0.986, 1.068]; unresolved | 1.015 [0.984, 1.061]; unresolved | 0.995 [0.978, 1.013]; unresolved |
| original22-v1/core/memory-load-store | 0.971 [0.943, 1.000]; unresolved | 0.979 [0.959, 1.000]; unresolved | 0.988 [0.954, 1.021]; unresolved |
| original22-v1/memory/narrow-unaligned | 0.980 [0.950, 1.005]; unresolved | 1.002 [0.981, 1.025]; unresolved | 0.983 [0.953, 1.004]; unresolved |
| original22-v1/memory/dependent-address | 1.007 [0.993, 1.022]; unresolved | 1.016 [0.989, 1.052]; unresolved | 0.978 [0.950, 1.005]; unresolved |
| original22-v1/memory/float64-unaligned | 1.034 [1.008, 1.075]; unresolved | 0.978 [0.929, 1.015]; unresolved | 1.016 [0.992, 1.057]; unresolved |
| original22-v1/runtime/simd-memory-full | 0.986 [0.956, 1.015]; unresolved | 0.973 [0.936, 1.003]; unresolved | 1.010 [0.983, 1.053]; unresolved |
| original22-v1/runtime/simd-memory-widen | 0.967 [0.924, 1.009]; unresolved | 0.918 [0.879, 0.963]; unresolved | 1.036 [1.020, 1.051]; unresolved |
| original22-v1/runtime/simd-memory-lanes | 0.927 [0.834, 0.989]; unresolved | 0.987 [0.953, 1.014]; unresolved | 1.022 [0.980, 1.063]; unresolved |
| original22-v1/gc/i31-roundtrip | 0.991 [0.968, 1.012]; unresolved | 0.977 [0.958, 0.997]; unresolved | 1.021 [1.006, 1.038]; unresolved |
| original22-v1/gc/struct-access | 1.013 [0.992, 1.041]; unresolved | 1.005 [0.982, 1.024]; unresolved | 0.974 [0.918, 1.014]; unresolved |
| original22-v1/gc/struct-allocation | 1.002 [0.990, 1.015]; unresolved | 1.000 [0.954, 1.044]; unresolved | 1.031 [1.016, 1.045]; unresolved |
| original22-v1/gc/struct-indirect-access | 1.007 [0.999, 1.015]; unresolved | 0.992 [0.968, 1.012]; unresolved | 1.021 [1.002, 1.042]; unresolved |
| original22-v1/threads/private-memory | 1.003 [0.972, 1.042]; unresolved | 1.018 [0.993, 1.049]; unresolved | 1.001 [0.975, 1.038]; unresolved |
| original22-v1/threads/shared-memory | 1.038 [0.997, 1.099]; unresolved | 0.998 [0.970, 1.026]; unresolved | 1.004 [0.954, 1.059]; unresolved |
| original22-v1/threads/atomic-rmw | 0.973 [0.936, 1.001]; unresolved | 1.015 [0.978, 1.073]; unresolved | 0.980 [0.958, 1.002]; unresolved |
| original22-v1/threads/notify-empty | 0.962 [0.922, 1.002]; unresolved | 1.013 [0.981, 1.043]; unresolved | 0.985 [0.974, 1.001]; unresolved |
| original22-v1/threads/wait-mismatch | 0.986 [0.953, 1.038]; unresolved | 0.980 [0.939, 1.016]; unresolved | 1.006 [0.953, 1.072]; unresolved |
| original22-v1/threads/wait-timeout | 1.011 [0.972, 1.051]; unresolved | 0.981 [0.960, 1.002]; unresolved | 0.991 [0.962, 1.018]; unresolved |
| original22-v1/applications/sha256-clang | 1.004 [0.973, 1.039]; unresolved | 0.992 [0.942, 1.043]; unresolved | 0.993 [0.935, 1.030]; unresolved |
| original22-v1/applications/sha256-rust | 0.965 [0.900, 1.034]; unresolved | 1.012 [0.979, 1.063]; unresolved | 0.986 [0.946, 1.023]; unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 1.012 [0.984, 1.042]; unresolved | 0.965 [0.953, 0.976]; unresolved | 0.990 [0.946, 1.029]; unresolved |
| gc-sustained-v1/gc/struct-access | 0.977 [0.943, 1.003]; unresolved | 0.998 [0.972, 1.029]; unresolved | 1.007 [0.990, 1.025]; unresolved |
| gc-sustained-v1/gc/struct-allocation | 0.984 [0.886, 1.084]; unresolved | 0.986 [0.965, 1.005]; unresolved | 1.017 [0.991, 1.042]; unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.022 [0.977, 1.068]; unresolved | 0.984 [0.954, 1.011]; unresolved | 1.049 [1.017, 1.088]; unresolved |
| threads-sustained-v2/threads/private-memory | 1.009 [0.995, 1.025]; unresolved | 0.994 [0.988, 1.000]; unresolved | 0.996 [0.983, 1.009]; unresolved |
| threads-sustained-v2/threads/shared-memory | 1.002 [0.990, 1.018]; unresolved | 1.004 [0.995, 1.016]; unresolved | 0.984 [0.964, 1.003]; unresolved |
| threads-sustained-v2/threads/atomic-rmw | 0.990 [0.973, 1.002]; unresolved | 0.989 [0.972, 1.006]; unresolved | 1.004 [0.991, 1.018]; pass |
| threads-sustained-v2/threads/notify-empty | 0.988 [0.962, 1.009]; unresolved | 0.997 [0.985, 1.013]; unresolved | 1.005 [1.000, 1.010]; unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.033 [1.006, 1.074]; unresolved | 1.014 [1.003, 1.025]; unresolved | 1.008 [0.988, 1.028]; unresolved |
| threads-sustained-v2/threads/wait-timeout | 0.998 [0.983, 1.016]; unresolved | 0.995 [0.976, 1.017]; unresolved | 1.006 [0.993, 1.019]; unresolved |
| generic-calls-v1/core/generic-calls | 1.024 [0.987, 1.062]; unresolved | 1.016 [0.977, 1.062]; unresolved | 1.005 [0.986, 1.026]; unresolved |

## Preparation, first call and cleanup

All entries below are **microseconds per process or invocation**, including unresolved observations. Phase intervals are not inferred from warm-call controls. Startup is independently observed process readiness. Native compile/validate boundaries differ by API; WasmEdge AOT includes validation and artifact I/O, plus separately recorded native-image loading. Every process creates a fresh native image. First-call and warm execution have their own boundaries; values are never subtracted to invent pure costs.

<details>
<summary>Native: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,949.60 | 7,453.31 | 4,748.10 | 4,703.60 | 6,677.15 | 20,688.29 |
| original22-v1/core/integer-loop | 4,739.19 | 6,926.13 | 4,813.98 | 5,773.02 | 6,696.15 | 17,970.69 |
| original22-v1/core/function-calls | 5,363.79 | 8,970.65 | 5,282.79 | 5,160.35 | 7,823.83 | 19,585.02 |
| original22-v1/core/memory-load-store | 5,312.69 | 7,038.15 | 5,149.52 | 4,635.83 | 7,079.40 | 19,818.02 |
| original22-v1/memory/narrow-unaligned | 5,044.83 | 6,604.17 | 4,335.79 | 5,171.10 | 7,702.52 | 17,452.79 |
| original22-v1/memory/dependent-address | 5,722.65 | 7,921.85 | 5,134.98 | 5,575.90 | 7,445.67 | 20,498.81 |
| original22-v1/memory/float64-unaligned | 5,873.71 | 7,387.63 | 5,363.23 | 5,404.48 | 7,264.83 | 20,234.46 |
| original22-v1/runtime/simd-memory-full | 5,196.42 | 7,642.40 | 5,896.62 | 5,224.90 | 8,176.33 | 21,895.06 |
| original22-v1/runtime/simd-memory-widen | 5,921.65 | 7,130.83 | 6,015.50 | 6,691.85 | 7,346.69 | 21,068.67 |
| original22-v1/runtime/simd-memory-lanes | 4,881.71 | 7,549.58 | 5,961.29 | 5,373.73 | 8,455.44 | 19,450.62 |
| original22-v1/gc/i31-roundtrip | 4,781.15 | 6,787.48 | US | 4,916.48 | US | 17,552.35 |
| original22-v1/gc/struct-access | 4,503.67 | 7,088.02 | US | 4,882.04 | US | 18,186.02 |
| original22-v1/gc/struct-allocation | 5,518.46 | 8,011.60 | US | 5,219.42 | US | 21,832.38 |
| original22-v1/gc/struct-indirect-access | 5,470.44 | 7,080.27 | US | 4,675.21 | US | 18,852.23 |
| original22-v1/threads/private-memory | 5,328.52 | 7,739.67 | 6,129.40 | 5,296.48 | 7,861.52 | 20,668.23 |
| original22-v1/threads/shared-memory | 4,867.04 | 7,468.67 | 4,950.08 | 5,445.75 | 7,407.73 | 19,398.65 |
| original22-v1/threads/atomic-rmw | 4,387.50 | 6,430.96 | 4,693.96 | 4,373.81 | 6,390.81 | 18,200.46 |
| original22-v1/threads/notify-empty | 4,850.25 | 7,619.27 | 4,793.38 | 5,151.63 | 6,869.65 | 19,341.00 |
| original22-v1/threads/wait-mismatch | 4,703.44 | 6,997.65 | 5,350.06 | 4,820.94 | 7,216.46 | 19,084.12 |
| original22-v1/threads/wait-timeout | 5,394.19 | 8,652.44 | 5,836.71 | 5,325.98 | 8,056.98 | 21,150.19 |
| original22-v1/applications/sha256-clang | 4,877.98 | 6,867.85 | 5,333.27 | 5,464.54 | 7,942.21 | 20,405.87 |
| original22-v1/applications/sha256-rust | 5,175.69 | 7,289.65 | 5,426.98 | 5,361.44 | 6,915.40 | 20,099.19 |
| gc-sustained-v1/gc/i31-roundtrip | 4,850.33 | 6,829.60 | US | 5,191.67 | US | 18,434.54 |
| gc-sustained-v1/gc/struct-access | 5,484.21 | 7,343.88 | US | 5,162.21 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 5,174.35 | 7,001.69 | US | 4,935.83 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 5,882.52 | 7,724.50 | US | 5,425.56 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 5,321.31 | 9,096.94 | 6,863.25 | 6,269.12 | 8,345.23 | 21,576.81 |
| threads-sustained-v2/threads/shared-memory | 5,320.96 | 8,932.69 | 5,282.69 | 5,277.35 | 7,530.77 | 22,089.10 |
| threads-sustained-v2/threads/atomic-rmw | 4,981.46 | 7,746.42 | 5,008.60 | 5,039.54 | 7,851.98 | 19,777.12 |
| threads-sustained-v2/threads/notify-empty | 4,606.60 | 7,123.06 | 4,945.94 | 5,043.73 | 7,350.08 | 18,774.29 |
| threads-sustained-v2/threads/wait-mismatch | 5,262.79 | 7,859.79 | 6,007.71 | 5,234.48 | 7,490.98 | 23,141.92 |
| threads-sustained-v2/threads/wait-timeout | 5,335.73 | 7,596.46 | 5,667.60 | 5,205.92 | 7,411.96 | 20,643.96 |
| generic-calls-v1/core/generic-calls | 5,556.88 | 7,875.62 | 5,245.02 | 5,042.06 | 7,273.46 | 21,721.58 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 64.938 | 57.001 | 43.062 | 43.292 | 45.688 | 58.792 |
| original22-v1/core/integer-loop | 51.666 | 67.271 | 58.396 | 56.000 | 52.896 | 69.980 |
| original22-v1/core/function-calls | 61.917 | 68.167 | 59.062 | 52.625 | 66.312 | 68.480 |
| original22-v1/core/memory-load-store | 63.083 | 59.562 | 53.021 | 48.458 | 57.104 | 72.188 |
| original22-v1/memory/narrow-unaligned | 66.730 | 65.729 | 45.000 | 42.375 | 47.437 | 53.999 |
| original22-v1/memory/dependent-address | 62.771 | 63.396 | 63.458 | 49.230 | 45.791 | 75.188 |
| original22-v1/memory/float64-unaligned | 82.146 | 75.917 | 55.459 | 49.708 | 72.395 | 59.980 |
| original22-v1/runtime/simd-memory-full | 59.791 | 63.666 | 63.729 | 58.416 | 57.167 | 66.209 |
| original22-v1/runtime/simd-memory-widen | 62.938 | 58.395 | 61.917 | 60.688 | 63.834 | 56.354 |
| original22-v1/runtime/simd-memory-lanes | 63.396 | 65.459 | 62.771 | 71.521 | 75.791 | 59.166 |
| original22-v1/gc/i31-roundtrip | 50.791 | 59.125 | US | 45.729 | US | 61.834 |
| original22-v1/gc/struct-access | 47.937 | 51.062 | US | 44.188 | US | 54.250 |
| original22-v1/gc/struct-allocation | 56.355 | 62.438 | US | 55.604 | US | 68.979 |
| original22-v1/gc/struct-indirect-access | 50.750 | 46.647 | US | 48.458 | US | 53.251 |
| original22-v1/threads/private-memory | 50.688 | 54.396 | 60.541 | 71.750 | 50.145 | 59.875 |
| original22-v1/threads/shared-memory | 48.375 | 58.812 | 53.625 | 45.458 | 58.480 | 55.042 |
| original22-v1/threads/atomic-rmw | 48.603 | 53.375 | 50.959 | 41.479 | 42.271 | 63.833 |
| original22-v1/threads/notify-empty | 45.938 | 63.978 | 52.958 | 43.270 | 57.500 | 60.083 |
| original22-v1/threads/wait-mismatch | 58.271 | 52.646 | 46.499 | 39.958 | 54.022 | 58.042 |
| original22-v1/threads/wait-timeout | 50.562 | 67.771 | 50.646 | 48.604 | 65.041 | 61.853 |
| original22-v1/applications/sha256-clang | 58.021 | 54.709 | 54.333 | 51.792 | 73.105 | 69.708 |
| original22-v1/applications/sha256-rust | 64.604 | 61.105 | 61.229 | 58.520 | 52.646 | 65.729 |
| gc-sustained-v1/gc/i31-roundtrip | 56.166 | 47.916 | US | 46.229 | US | 63.084 |
| gc-sustained-v1/gc/struct-access | 65.459 | 64.833 | US | 45.791 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 60.772 | 53.666 | US | 56.500 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 56.020 | 56.541 | US | 57.291 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 56.500 | 68.688 | 55.855 | 45.729 | 54.459 | 64.062 |
| threads-sustained-v2/threads/shared-memory | 65.938 | 59.875 | 59.646 | 46.354 | 60.270 | 65.709 |
| threads-sustained-v2/threads/atomic-rmw | 74.479 | 75.354 | 59.334 | 41.312 | 56.959 | 62.812 |
| threads-sustained-v2/threads/notify-empty | 50.625 | 58.688 | 57.834 | 46.916 | 56.750 | 65.000 |
| threads-sustained-v2/threads/wait-mismatch | 63.209 | 74.438 | 76.958 | 49.749 | 59.438 | 69.230 |
| threads-sustained-v2/threads/wait-timeout | 50.959 | 61.646 | 60.562 | 48.834 | 47.062 | 64.875 |
| generic-calls-v1/core/generic-calls | 63.895 | 59.041 | 67.584 | 46.605 | 55.688 | 70.187 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 7.042 | 6.187 | 67.084 | 141.88 | 47.666 | 31.355 |
| original22-v1/core/integer-loop | 6.812 | 5.604 | 63.584 | 151.23 | 47.562 | 33.917 |
| original22-v1/core/function-calls | 7.188 | 6.021 | 81.479 | 146.33 | 48.812 | 33.250 |
| original22-v1/core/memory-load-store | 7.167 | 4.792 | 63.792 | 134.25 | 48.709 | 32.750 |
| original22-v1/memory/narrow-unaligned | 7.208 | 6.729 | 54.834 | 137.67 | 50.896 | 30.312 |
| original22-v1/memory/dependent-address | 7.542 | 5.771 | 71.708 | 145.94 | 48.166 | 31.250 |
| original22-v1/memory/float64-unaligned | 8.292 | 6.063 | 70.750 | 139.77 | 54.688 | 30.313 |
| original22-v1/runtime/simd-memory-full | 7.062 | 5.396 | 70.209 | 154.52 | 59.730 | 33.354 |
| original22-v1/runtime/simd-memory-widen | 6.708 | 6.271 | 67.896 | 172.77 | 53.854 | 30.209 |
| original22-v1/runtime/simd-memory-lanes | 7.375 | 6.083 | 94.562 | 159.54 | 58.271 | 30.625 |
| original22-v1/gc/i31-roundtrip | 7.312 | 6.083 | US | 144.85 | US | 32.896 |
| original22-v1/gc/struct-access | 7.000 | 5.479 | US | 140.08 | US | 29.416 |
| original22-v1/gc/struct-allocation | 7.396 | 6.021 | US | 141.04 | US | 32.563 |
| original22-v1/gc/struct-indirect-access | 7.146 | 5.500 | US | 145.23 | US | 31.375 |
| original22-v1/threads/private-memory | 7.542 | 5.854 | 77.709 | 148.25 | 51.292 | 33.209 |
| original22-v1/threads/shared-memory | 7.333 | 6.146 | 65.230 | 139.19 | 49.812 | 30.334 |
| original22-v1/threads/atomic-rmw | 7.333 | 5.312 | 61.896 | 130.52 | 44.292 | 33.854 |
| original22-v1/threads/notify-empty | 6.646 | 5.688 | 66.479 | 150.54 | 47.792 | 32.480 |
| original22-v1/threads/wait-mismatch | 7.083 | 5.583 | 61.209 | 140.63 | 49.230 | 28.292 |
| original22-v1/threads/wait-timeout | 7.125 | 6.417 | 65.479 | 151.10 | 63.999 | 31.334 |
| original22-v1/applications/sha256-clang | 6.792 | 5.812 | 66.959 | 151.00 | 53.791 | 27.417 |
| original22-v1/applications/sha256-rust | 7.125 | 5.958 | 70.979 | 151.12 | 50.041 | 31.521 |
| gc-sustained-v1/gc/i31-roundtrip | 6.938 | 5.312 | US | 145.10 | US | 30.354 |
| gc-sustained-v1/gc/struct-access | 7.500 | 5.729 | US | 146.00 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 7.354 | 5.312 | US | 146.00 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 7.750 | 5.375 | US | 163.60 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 7.459 | 6.333 | 121.00 | 152.56 | 53.167 | 32.916 |
| threads-sustained-v2/threads/shared-memory | 8.000 | 6.771 | 77.749 | 158.10 | 52.416 | 40.146 |
| threads-sustained-v2/threads/atomic-rmw | 6.958 | 5.834 | 69.125 | 141.38 | 50.770 | 27.938 |
| threads-sustained-v2/threads/notify-empty | 7.083 | 5.542 | 59.562 | 144.08 | 51.041 | 32.042 |
| threads-sustained-v2/threads/wait-mismatch | 7.083 | 5.812 | 70.520 | 163.17 | 49.646 | 33.041 |
| threads-sustained-v2/threads/wait-timeout | 6.688 | 6.208 | 65.334 | 160.23 | 49.062 | 34.730 |
| generic-calls-v1/core/generic-calls | 8.334 | 6.708 | 62.917 | 155.62 | 47.479 | 30.708 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 186.46 | 177.56 | 102.85 | 840.79 | 1,278.94 | 285,412.15 |
| original22-v1/core/integer-loop | 231.65 | 216.85 | 118.90 | 789.33 | 1,340.40 | 285,018.60 |
| original22-v1/core/function-calls | 351.60 | 333.44 | 153.94 | 964.27 | 1,366.42 | 285,889.56 |
| original22-v1/core/memory-load-store | 396.75 | 327.29 | 173.52 | 781.85 | 1,412.25 | 293,004.73 |
| original22-v1/memory/narrow-unaligned | 333.27 | 322.79 | 147.15 | 832.15 | 1,503.98 | 284,741.52 |
| original22-v1/memory/dependent-address | 355.23 | 330.81 | 171.79 | 975.02 | 1,364.81 | 287,020.83 |
| original22-v1/memory/float64-unaligned | 357.15 | 351.73 | 181.17 | 860.96 | 1,338.46 | 297,603.25 |
| original22-v1/runtime/simd-memory-full | 334.06 | 306.90 | 144.08 | 1,216.77 | 1,537.83 | 285,729.17 |
| original22-v1/runtime/simd-memory-widen | 284.06 | 284.00 | 167.52 | 1,415.48 | 1,357.23 | 290,478.56 |
| original22-v1/runtime/simd-memory-lanes | 335.23 | 296.23 | 147.67 | 1,311.02 | 1,679.69 | 284,162.77 |
| original22-v1/gc/i31-roundtrip | 199.98 | 181.27 | US | 789.56 | US | 286,190.10 |
| original22-v1/gc/struct-access | 242.42 | 208.27 | US | 853.71 | US | 287,494.75 |
| original22-v1/gc/struct-allocation | 296.42 | 230.23 | US | 1,197.73 | US | 286,055.33 |
| original22-v1/gc/struct-indirect-access | 288.44 | 277.88 | US | 1,687.83 | US | 289,570.21 |
| original22-v1/threads/private-memory | 302.79 | 233.25 | 163.65 | 897.37 | 1,281.71 | 291,733.98 |
| original22-v1/threads/shared-memory | 251.04 | 231.56 | 149.83 | 762.10 | 1,272.23 | 288,317.48 |
| original22-v1/threads/atomic-rmw | 238.60 | 182.25 | 141.69 | 799.40 | 1,234.23 | 289,637.73 |
| original22-v1/threads/notify-empty | 220.10 | 236.92 | 149.75 | 1,001.98 | 1,250.71 | 284,283.73 |
| original22-v1/threads/wait-mismatch | 255.08 | 211.31 | 154.85 | 928.58 | 1,291.42 | 283,049.23 |
| original22-v1/threads/wait-timeout | 295.33 | 237.56 | 156.77 | 1,076.54 | 1,428.60 | 287,082.69 |
| original22-v1/applications/sha256-clang | 1,623.92 | 1,947.75 | 372.83 | 1,554.35 | 2,400.33 | 286,058.81 |
| original22-v1/applications/sha256-rust | 1,713.90 | 2,118.62 | 396.33 | 1,626.38 | 2,123.10 | 303,439.60 |
| gc-sustained-v1/gc/i31-roundtrip | 202.06 | 173.02 | US | 841.73 | US | 288,943.23 |
| gc-sustained-v1/gc/struct-access | 295.79 | 228.17 | US | 951.98 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 262.54 | 220.75 | US | 1,279.67 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 319.29 | 285.94 | US | 1,849.60 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 316.29 | 248.62 | 182.29 | 967.10 | 1,468.54 | 289,244.69 |
| threads-sustained-v2/threads/shared-memory | 255.58 | 255.12 | 154.52 | 992.69 | 1,383.65 | 286,462.81 |
| threads-sustained-v2/threads/atomic-rmw | 248.69 | 200.58 | 146.25 | 864.65 | 1,280.38 | 284,624.42 |
| threads-sustained-v2/threads/notify-empty | 241.08 | 192.79 | 144.35 | 934.17 | 1,275.31 | 286,617.88 |
| threads-sustained-v2/threads/wait-mismatch | 279.04 | 250.42 | 169.88 | 1,111.21 | 1,313.02 | 287,723.23 |
| threads-sustained-v2/threads/wait-timeout | 254.81 | 209.50 | 142.75 | 962.10 | 1,294.04 | 286,483.04 |
| generic-calls-v1/core/generic-calls | 354.69 | 310.38 | 151.33 | 1,020.40 | 1,344.04 | 288,810.92 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 10.084 | 9.416 | 15.229 | 8.291 | 19.187 | 18.312 |
| original22-v1/core/integer-loop | 8.354 | 8.166 | 14.896 | 8.104 | 19.812 | 16.166 |
| original22-v1/core/function-calls | 10.479 | 9.688 | 16.666 | 7.896 | 19.980 | 18.041 |
| original22-v1/core/memory-load-store | 15.625 | 13.729 | 19.563 | 21.583 | 33.354 | 38.001 |
| original22-v1/memory/narrow-unaligned | 15.312 | 13.792 | 18.270 | 22.459 | 34.125 | 36.209 |
| original22-v1/memory/dependent-address | 15.459 | 14.646 | 20.812 | 23.792 | 35.105 | 39.791 |
| original22-v1/memory/float64-unaligned | 16.229 | 15.291 | 21.562 | 23.333 | 34.688 | 39.041 |
| original22-v1/runtime/simd-memory-full | 16.812 | 17.354 | 20.709 | 66.584 | 44.209 | 50.146 |
| original22-v1/runtime/simd-memory-widen | 18.083 | 16.291 | 24.750 | 69.709 | 39.021 | 49.459 |
| original22-v1/runtime/simd-memory-lanes | 16.938 | 15.583 | 20.542 | 70.230 | 42.834 | 48.209 |
| original22-v1/gc/i31-roundtrip | 9.542 | 10.000 | US | 8.062 | US | 15.854 |
| original22-v1/gc/struct-access | 10.501 | 9.979 | US | 23.604 | US | 15.896 |
| original22-v1/gc/struct-allocation | 11.812 | 10.938 | US | 56.812 | US | 24.021 |
| original22-v1/gc/struct-indirect-access | 22.645 | 24.000 | US | 64.583 | US | 24.396 |
| original22-v1/threads/private-memory | 16.854 | 15.916 | 22.084 | 23.791 | 34.521 | 40.375 |
| original22-v1/threads/shared-memory | 15.271 | 14.646 | 20.625 | 21.604 | 38.604 | 39.062 |
| original22-v1/threads/atomic-rmw | 14.021 | 12.875 | 20.145 | 21.792 | 39.084 | 37.771 |
| original22-v1/threads/notify-empty | 14.917 | 15.459 | 18.291 | 23.812 | 38.396 | 37.250 |
| original22-v1/threads/wait-mismatch | 15.479 | 14.876 | 18.375 | 23.271 | 39.645 | 43.874 |
| original22-v1/threads/wait-timeout | 16.958 | 15.521 | 22.562 | 25.291 | 42.292 | 39.666 |
| original22-v1/applications/sha256-clang | 20.730 | 19.166 | 20.521 | 65.563 | 41.916 | 48.500 |
| original22-v1/applications/sha256-rust | 20.479 | 19.062 | 22.625 | 64.875 | 40.792 | 51.562 |
| gc-sustained-v1/gc/i31-roundtrip | 9.334 | 8.251 | US | 8.542 | US | 17.167 |
| gc-sustained-v1/gc/struct-access | 12.312 | 10.104 | US | 24.062 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 11.250 | 10.479 | US | 58.250 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 25.729 | 24.605 | US | 74.749 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 16.875 | 15.625 | 22.625 | 25.521 | 37.584 | 44.312 |
| threads-sustained-v2/threads/shared-memory | 16.771 | 15.916 | 23.271 | 24.729 | 42.812 | 43.812 |
| threads-sustained-v2/threads/atomic-rmw | 14.562 | 15.375 | 19.125 | 25.854 | 40.958 | 40.000 |
| threads-sustained-v2/threads/notify-empty | 15.021 | 14.666 | 19.645 | 23.021 | 36.980 | 40.770 |
| threads-sustained-v2/threads/wait-mismatch | 16.188 | 15.771 | 23.021 | 24.562 | 41.416 | 41.042 |
| threads-sustained-v2/threads/wait-timeout | 15.438 | 15.750 | 19.563 | 24.145 | 38.208 | 39.417 |
| generic-calls-v1/core/generic-calls | 10.646 | 10.562 | 15.417 | 8.834 | 17.625 | 16.270 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.312 | 0.437 | 2.604 | 4.125 | 5.271 | 0.292 |
| original22-v1/core/integer-loop | 0.250 | 0.396 | 2.271 | 3.938 | 5.229 | 0.292 |
| original22-v1/core/function-calls | 0.312 | 0.458 | 2.292 | 4.271 | 5.438 | 0.292 |
| original22-v1/core/memory-load-store | 0.291 | 0.438 | 2.542 | 3.833 | 4.979 | 0.230 |
| original22-v1/memory/narrow-unaligned | 0.291 | 0.438 | 1.750 | 4.188 | 5.292 | 0.250 |
| original22-v1/memory/dependent-address | 0.271 | 0.416 | 2.146 | 4.479 | 5.438 | 0.230 |
| original22-v1/memory/float64-unaligned | 0.312 | 0.417 | 2.667 | 4.292 | 5.396 | 0.230 |
| original22-v1/runtime/simd-memory-full | 0.292 | 0.417 | 2.396 | 2.229 | 5.812 | 0.250 |
| original22-v1/runtime/simd-memory-widen | 0.291 | 0.500 | 2.583 | 2.333 | 5.438 | 0.250 |
| original22-v1/runtime/simd-memory-lanes | 0.334 | 0.458 | 2.501 | 2.375 | 5.854 | 0.209 |
| original22-v1/gc/i31-roundtrip | 0.292 | 0.396 | US | 3.854 | US | 0.333 |
| original22-v1/gc/struct-access | 0.291 | 0.416 | US | 4.229 | US | 0.271 |
| original22-v1/gc/struct-allocation | 0.291 | 0.458 | US | 1.979 | US | 0.291 |
| original22-v1/gc/struct-indirect-access | 0.271 | 0.479 | US | 2.396 | US | 0.250 |
| original22-v1/threads/private-memory | 0.271 | 0.479 | 2.521 | 4.188 | 5.292 | 0.271 |
| original22-v1/threads/shared-memory | 0.291 | 0.458 | 3.146 | 4.083 | 5.271 | 0.230 |
| original22-v1/threads/atomic-rmw | 0.271 | 0.459 | 2.833 | 4.042 | 5.229 | 0.250 |
| original22-v1/threads/notify-empty | 0.333 | 0.500 | 2.208 | 4.396 | 5.521 | 0.292 |
| original22-v1/threads/wait-mismatch | 0.292 | 0.520 | 2.583 | 4.188 | 5.458 | 0.271 |
| original22-v1/threads/wait-timeout | 0.292 | 0.438 | 3.396 | 4.521 | 5.854 | 0.208 |
| original22-v1/applications/sha256-clang | 0.292 | 0.438 | 2.708 | 2.146 | 6.417 | 0.250 |
| original22-v1/applications/sha256-rust | 0.291 | 0.459 | 2.771 | 2.521 | 5.667 | 0.271 |
| gc-sustained-v1/gc/i31-roundtrip | 0.291 | 0.416 | US | 4.396 | US | 0.249 |
| gc-sustained-v1/gc/struct-access | 0.250 | 0.374 | US | 4.250 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 0.291 | 0.416 | US | 2.021 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 0.396 | 0.458 | US | 2.250 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 0.292 | 0.541 | 2.688 | 4.771 | 5.771 | 0.333 |
| threads-sustained-v2/threads/shared-memory | 0.313 | 0.458 | 3.333 | 4.771 | 5.688 | 0.271 |
| threads-sustained-v2/threads/atomic-rmw | 0.291 | 0.396 | 2.792 | 4.604 | 5.479 | 0.250 |
| threads-sustained-v2/threads/notify-empty | 0.250 | 0.520 | 2.604 | 4.062 | 4.729 | 0.271 |
| threads-sustained-v2/threads/wait-mismatch | 0.292 | 0.479 | 3.042 | 4.438 | 5.521 | 0.271 |
| threads-sustained-v2/threads/wait-timeout | 0.271 | 0.458 | 2.979 | 4.604 | 5.125 | 0.292 |
| generic-calls-v1/core/generic-calls | 0.292 | 0.417 | 3.063 | 4.250 | 5.187 | 0.271 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 12.313 | 8.791 | 8.792 | 30.917 | 17.104 | 18.688 |
| original22-v1/core/integer-loop | 1,694.69 | 1,729.50 | 1,674.02 | 1,828.56 | 1,684.83 | 1,707.96 |
| original22-v1/core/function-calls | 973.71 | 1,002.06 | 1,569.08 | 1,137.54 | 971.52 | 1,070.25 |
| original22-v1/core/memory-load-store | 551.02 | 506.19 | 878.52 | 562.65 | 664.33 | 345.21 |
| original22-v1/memory/narrow-unaligned | 828.73 | 700.33 | 993.79 | 794.75 | 843.17 | 587.06 |
| original22-v1/memory/dependent-address | 2,198.10 | 2,286.15 | 3,698.67 | 2,632.31 | 2,558.77 | 1,831.94 |
| original22-v1/memory/float64-unaligned | 987.06 | 897.40 | 988.46 | 926.44 | 857.94 | 854.73 |
| original22-v1/runtime/simd-memory-full | 4,232.06 | 4,245.58 | 4,746.06 | 4,700.56 | 4,589.58 | 4,520.75 |
| original22-v1/runtime/simd-memory-widen | 1,320.65 | 1,202.42 | 1,331.17 | 1,395.00 | 1,386.27 | 1,364.60 |
| original22-v1/runtime/simd-memory-lanes | 6,059.83 | 6,160.60 | 7,418.77 | 7,468.04 | 7,515.58 | 6,106.88 |
| original22-v1/gc/i31-roundtrip | 25.812 | 24.250 | US | 51.666 | US | 33.062 |
| original22-v1/gc/struct-access | 38.230 | 34.021 | US | 57.270 | US | 136.88 |
| original22-v1/gc/struct-allocation | 493.98 | 471.35 | US | 161.58 | US | 576.77 |
| original22-v1/gc/struct-indirect-access | 32.959 | 26.896 | US | 32.980 | US | 351.62 |
| original22-v1/threads/private-memory | 18.688 | 13.688 | 22.480 | 52.209 | 33.229 | 36.416 |
| original22-v1/threads/shared-memory | 28.229 | 25.604 | 20.313 | 44.791 | 28.562 | 34.895 |
| original22-v1/threads/atomic-rmw | 33.271 | 29.355 | 28.125 | 51.604 | 60.916 | 41.770 |
| original22-v1/threads/notify-empty | 26.021 | 21.021 | 267.65 | 125.19 | 76.125 | 131.48 |
| original22-v1/threads/wait-mismatch | 27.208 | 23.083 | 300.73 | 148.08 | 50.480 | 59.604 |
| original22-v1/threads/wait-timeout | 27.188 | 22.125 | 3,785.42 | 633.58 | 2,548.71 | 796.50 |
| original22-v1/applications/sha256-clang | 1,422.42 | 1,254.04 | 1,686.31 | 1,112.29 | 1,188.81 | 1,047.25 |
| original22-v1/applications/sha256-rust | 1,462.31 | 1,450.50 | 1,815.75 | 1,216.08 | 1,105.40 | 1,145.27 |
| gc-sustained-v1/gc/i31-roundtrip | 1,573.31 | 1,528.02 | US | 2,081.56 | US | 1,462.21 |
| gc-sustained-v1/gc/struct-access | 1,699.94 | 1,619.35 | US | 1,264.44 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 31,704.96 | 16,833.81 | US | 3,624.67 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 2,289.12 | 1,888.58 | US | 3,184.83 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 47.770 | 44.480 | 154.04 | 168.08 | 147.85 | 149.08 |
| threads-sustained-v2/threads/shared-memory | 172.60 | 180.54 | 141.00 | 167.50 | 145.06 | 157.73 |
| threads-sustained-v2/threads/atomic-rmw | 224.96 | 220.10 | 221.81 | 250.54 | 487.10 | 260.27 |
| threads-sustained-v2/threads/notify-empty | 119.52 | 115.02 | 2,209.17 | 762.90 | 536.96 | 967.54 |
| threads-sustained-v2/threads/wait-mismatch | 139.85 | 151.21 | 3,291.90 | 1,197.56 | 328.12 | 397.79 |
| threads-sustained-v2/threads/wait-timeout | 133.81 | 106.15 | 31,703.54 | 5,472.96 | 20,115.79 | 7,312.65 |
| generic-calls-v1/core/generic-calls | 24,960.27 | 24,234.27 | 12,265.12 | 11,936.13 | 11,819.27 | 6,283.75 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.604 | 5.896 | 12.291 | 79.875 | 37.959 | 421.25 |
| original22-v1/core/integer-loop | 11.541 | 8.291 | 9.271 | 83.021 | 39.188 | 480.54 |
| original22-v1/core/function-calls | 12.667 | 5.771 | 10.000 | 106.12 | 42.709 | 510.52 |
| original22-v1/core/memory-load-store | 16.604 | 17.438 | 13.042 | 124.37 | 65.854 | 520.67 |
| original22-v1/memory/narrow-unaligned | 17.313 | 22.313 | 12.875 | 131.06 | 66.292 | 530.08 |
| original22-v1/memory/dependent-address | 18.459 | 13.584 | 5.917 | 117.02 | 64.188 | 541.21 |
| original22-v1/memory/float64-unaligned | 12.708 | 13.249 | 6.041 | 101.60 | 64.292 | 591.96 |
| original22-v1/runtime/simd-memory-full | 20.709 | 16.541 | 12.771 | 139.17 | 67.480 | 542.54 |
| original22-v1/runtime/simd-memory-widen | 15.542 | 23.416 | 12.166 | 143.42 | 73.354 | 540.12 |
| original22-v1/runtime/simd-memory-lanes | 21.542 | 23.521 | 13.396 | 155.75 | 75.646 | 579.31 |
| original22-v1/gc/i31-roundtrip | 7.104 | 11.021 | US | 97.895 | US | 449.17 |
| original22-v1/gc/struct-access | 14.146 | 12.271 | US | 105.79 | US | 407.29 |
| original22-v1/gc/struct-allocation | 178.75 | 118.60 | US | 135.08 | US | 2,335.75 |
| original22-v1/gc/struct-indirect-access | 17.688 | 15.563 | US | 116.58 | US | 431.38 |
| original22-v1/threads/private-memory | 17.834 | 21.334 | 13.250 | 113.60 | 68.709 | 546.00 |
| original22-v1/threads/shared-memory | 18.313 | 14.354 | 5.583 | 87.041 | 58.625 | 450.75 |
| original22-v1/threads/atomic-rmw | 14.666 | 15.416 | 9.771 | 99.562 | 65.520 | 489.33 |
| original22-v1/threads/notify-empty | 12.521 | 10.646 | 4.042 | 89.667 | 57.375 | 500.23 |
| original22-v1/threads/wait-mismatch | 21.208 | 18.834 | 5.271 | 101.35 | 65.730 | 532.48 |
| original22-v1/threads/wait-timeout | 19.688 | 13.667 | 4.396 | 105.38 | 68.645 | 537.23 |
| original22-v1/applications/sha256-clang | 19.604 | 15.479 | 9.125 | 157.79 | 77.374 | 539.46 |
| original22-v1/applications/sha256-rust | 20.646 | 22.000 | 13.541 | 178.48 | 79.354 | 562.85 |
| gc-sustained-v1/gc/i31-roundtrip | 10.375 | 6.125 | US | 88.146 | US | 492.44 |
| gc-sustained-v1/gc/struct-access | 14.834 | 16.084 | US | 97.624 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | 3,884.62 | 2,486.48 | US | 867.65 | US | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | 21.479 | 19.041 | US | 185.65 | US | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 15.729 | 13.938 | 6.500 | 98.020 | 63.501 | 511.29 |
| threads-sustained-v2/threads/shared-memory | 21.105 | 13.021 | 9.959 | 106.19 | 66.896 | 514.48 |
| threads-sustained-v2/threads/atomic-rmw | 10.813 | 10.084 | 4.271 | 87.291 | 58.416 | 533.48 |
| threads-sustained-v2/threads/notify-empty | 11.104 | 10.229 | 4.333 | 99.250 | 62.938 | 507.67 |
| threads-sustained-v2/threads/wait-mismatch | 18.958 | 13.084 | 6.563 | 117.54 | 69.520 | 662.12 |
| threads-sustained-v2/threads/wait-timeout | 18.291 | 16.667 | 3.833 | 123.13 | 68.041 | 516.79 |
| generic-calls-v1/core/generic-calls | 13.333 | 14.062 | 12.396 | 142.10 | 54.751 | 669.98 |

</details>

<details>
<summary>Interpreter: all 33 workloads and phase observations</summary>

### processStartup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4,589.06 | 6,758.06 | 5,024.44 | NM | NM | 20,223.73 |
| original22-v1/core/integer-loop | 4,775.63 | 6,551.25 | 4,897.52 | NM | NM | 17,898.58 |
| original22-v1/core/function-calls | 5,113.96 | 7,168.98 | 6,079.96 | NM | NM | 19,614.56 |
| original22-v1/core/memory-load-store | 5,112.58 | 7,156.81 | 5,348.81 | NM | NM | 22,666.52 |
| original22-v1/memory/narrow-unaligned | 4,870.38 | 7,245.27 | 5,271.79 | NM | NM | 20,207.48 |
| original22-v1/memory/dependent-address | 5,199.79 | 7,617.81 | 6,104.25 | NM | NM | 21,637.23 |
| original22-v1/memory/float64-unaligned | 5,370.42 | 8,114.29 | 5,242.19 | NM | NM | 19,431.02 |
| original22-v1/runtime/simd-memory-full | 5,241.27 | 7,606.40 | 5,346.48 | NM | NM | 22,685.58 |
| original22-v1/runtime/simd-memory-widen | 5,577.44 | 7,949.71 | 5,085.81 | NM | NM | 19,273.10 |
| original22-v1/runtime/simd-memory-lanes | 5,648.44 | 7,670.23 | 5,051.15 | NM | NM | 20,115.71 |
| original22-v1/gc/i31-roundtrip | 4,557.00 | 6,764.79 | US | NM | NM | 19,563.50 |
| original22-v1/gc/struct-access | 4,399.58 | 6,482.04 | US | NM | NM | 18,444.65 |
| original22-v1/gc/struct-allocation | 5,033.50 | 7,854.42 | US | NM | NM | 20,347.38 |
| original22-v1/gc/struct-indirect-access | 4,490.94 | 6,860.44 | US | NM | NM | 19,855.81 |
| original22-v1/threads/private-memory | 4,982.40 | 7,920.25 | 5,977.25 | NM | NM | 21,180.23 |
| original22-v1/threads/shared-memory | 4,944.27 | 7,320.42 | 5,774.06 | NM | NM | 19,452.50 |
| original22-v1/threads/atomic-rmw | 4,300.02 | 7,144.10 | 4,610.96 | NM | NM | 17,466.94 |
| original22-v1/threads/notify-empty | 5,140.88 | 6,761.17 | 4,945.08 | NM | NM | 18,913.85 |
| original22-v1/threads/wait-mismatch | 4,614.04 | 6,997.56 | 4,721.52 | NM | NM | 20,357.19 |
| original22-v1/threads/wait-timeout | 5,040.19 | 8,166.65 | 5,853.52 | NM | NM | 22,006.27 |
| original22-v1/applications/sha256-clang | 4,828.35 | 6,678.52 | 5,317.52 | NM | NM | 19,988.00 |
| original22-v1/applications/sha256-rust | 4,829.62 | 8,222.21 | 4,933.21 | NM | NM | 20,137.98 |
| gc-sustained-v1/gc/i31-roundtrip | 4,756.38 | 6,943.67 | US | NM | NM | 20,064.77 |
| gc-sustained-v1/gc/struct-access | 5,142.50 | 7,318.69 | US | NM | NM | 20,272.79 |
| gc-sustained-v1/gc/struct-allocation | 5,173.62 | 7,484.44 | US | NM | NM | 19,330.83 |
| gc-sustained-v1/gc/struct-indirect-access | 5,724.27 | 8,123.10 | US | NM | NM | 21,920.71 |
| threads-sustained-v2/threads/private-memory | 5,058.75 | 8,925.90 | 6,315.08 | NM | NM | 22,200.27 |
| threads-sustained-v2/threads/shared-memory | 5,641.15 | 8,649.48 | 5,778.17 | NM | NM | 20,274.94 |
| threads-sustained-v2/threads/atomic-rmw | 4,529.27 | 7,151.50 | 5,211.50 | NM | NM | 21,416.69 |
| threads-sustained-v2/threads/notify-empty | 4,818.00 | 6,917.79 | 5,308.21 | NM | NM | 20,147.27 |
| threads-sustained-v2/threads/wait-mismatch | 5,281.58 | 9,831.29 | 5,594.08 | NM | NM | 21,982.58 |
| threads-sustained-v2/threads/wait-timeout | 5,499.19 | 7,433.23 | 5,236.15 | NM | NM | 20,834.75 |
| generic-calls-v1/core/generic-calls | 4,818.58 | 7,836.00 | 5,262.52 | NM | NM | 20,477.17 |

### load

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 45.708 | 50.688 | 55.791 | NM | NM | 56.333 |
| original22-v1/core/integer-loop | 53.000 | 51.000 | 54.146 | NM | NM | 83.855 |
| original22-v1/core/function-calls | 56.750 | 60.501 | 93.167 | NM | NM | 65.500 |
| original22-v1/core/memory-load-store | 71.292 | 67.084 | 86.812 | NM | NM | 76.625 |
| original22-v1/memory/narrow-unaligned | 53.916 | 56.375 | 52.709 | NM | NM | 73.375 |
| original22-v1/memory/dependent-address | 67.374 | 68.583 | 69.834 | NM | NM | 66.188 |
| original22-v1/memory/float64-unaligned | 65.125 | 72.312 | 57.084 | NM | NM | 70.333 |
| original22-v1/runtime/simd-memory-full | 52.584 | 53.875 | 52.334 | NM | NM | 75.375 |
| original22-v1/runtime/simd-memory-widen | 55.020 | 58.562 | 54.188 | NM | NM | 52.771 |
| original22-v1/runtime/simd-memory-lanes | 64.687 | 60.750 | 58.355 | NM | NM | 54.208 |
| original22-v1/gc/i31-roundtrip | 43.271 | 55.376 | US | NM | NM | 57.000 |
| original22-v1/gc/struct-access | 44.624 | 54.041 | US | NM | NM | 57.855 |
| original22-v1/gc/struct-allocation | 57.417 | 54.459 | US | NM | NM | 68.292 |
| original22-v1/gc/struct-indirect-access | 49.646 | 52.625 | US | NM | NM | 55.750 |
| original22-v1/threads/private-memory | 62.833 | 60.605 | 70.437 | NM | NM | 73.147 |
| original22-v1/threads/shared-memory | 47.084 | 63.812 | 56.062 | NM | NM | 64.812 |
| original22-v1/threads/atomic-rmw | 46.959 | 56.521 | 48.709 | NM | NM | 60.500 |
| original22-v1/threads/notify-empty | 46.500 | 57.605 | 49.813 | NM | NM | 52.874 |
| original22-v1/threads/wait-mismatch | 44.416 | 58.271 | 49.875 | NM | NM | 60.375 |
| original22-v1/threads/wait-timeout | 59.604 | 63.354 | 55.147 | NM | NM | 67.334 |
| original22-v1/applications/sha256-clang | 50.937 | 57.480 | 62.520 | NM | NM | 71.500 |
| original22-v1/applications/sha256-rust | 61.125 | 72.750 | 56.062 | NM | NM | 74.749 |
| gc-sustained-v1/gc/i31-roundtrip | 57.062 | 62.771 | US | NM | NM | 54.521 |
| gc-sustained-v1/gc/struct-access | 67.187 | 69.251 | US | NM | NM | 67.062 |
| gc-sustained-v1/gc/struct-allocation | 63.688 | 71.438 | US | NM | NM | 82.709 |
| gc-sustained-v1/gc/struct-indirect-access | 62.812 | 57.500 | US | NM | NM | 66.584 |
| threads-sustained-v2/threads/private-memory | 54.521 | 74.375 | 56.229 | NM | NM | 66.417 |
| threads-sustained-v2/threads/shared-memory | 70.458 | 79.812 | 74.541 | NM | NM | 53.249 |
| threads-sustained-v2/threads/atomic-rmw | 62.750 | 56.416 | 58.438 | NM | NM | 64.438 |
| threads-sustained-v2/threads/notify-empty | 47.730 | 57.188 | 50.750 | NM | NM | 61.021 |
| threads-sustained-v2/threads/wait-mismatch | 54.187 | 61.729 | 59.958 | NM | NM | 72.938 |
| threads-sustained-v2/threads/wait-timeout | 55.105 | 57.021 | 59.271 | NM | NM | 64.958 |
| generic-calls-v1/core/generic-calls | 52.229 | 68.250 | 58.188 | NM | NM | 71.520 |

### engineSetup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 4.896 | 5.667 | 2.958 | NM | NM | 29.980 |
| original22-v1/core/integer-loop | 5.479 | 5.542 | 2.812 | NM | NM | 32.334 |
| original22-v1/core/function-calls | 5.854 | 5.583 | 3.333 | NM | NM | 32.542 |
| original22-v1/core/memory-load-store | 5.500 | 5.375 | 2.771 | NM | NM | 30.292 |
| original22-v1/memory/narrow-unaligned | 5.416 | 5.750 | 2.979 | NM | NM | 31.312 |
| original22-v1/memory/dependent-address | 6.083 | 5.542 | 2.979 | NM | NM | 33.104 |
| original22-v1/memory/float64-unaligned | 5.271 | 5.771 | 2.583 | NM | NM | 31.396 |
| original22-v1/runtime/simd-memory-full | 5.250 | 5.375 | 2.521 | NM | NM | 33.209 |
| original22-v1/runtime/simd-memory-widen | 5.625 | 5.812 | 2.772 | NM | NM | 33.979 |
| original22-v1/runtime/simd-memory-lanes | 5.646 | 5.354 | 3.292 | NM | NM | 30.354 |
| original22-v1/gc/i31-roundtrip | 5.083 | 5.187 | US | NM | NM | 29.541 |
| original22-v1/gc/struct-access | 4.917 | 5.208 | US | NM | NM | 29.479 |
| original22-v1/gc/struct-allocation | 5.208 | 5.708 | US | NM | NM | 33.562 |
| original22-v1/gc/struct-indirect-access | 4.688 | 5.271 | US | NM | NM | 31.354 |
| original22-v1/threads/private-memory | 5.542 | 5.708 | 3.312 | NM | NM | 38.500 |
| original22-v1/threads/shared-memory | 4.958 | 6.021 | 2.938 | NM | NM | 33.729 |
| original22-v1/threads/atomic-rmw | 4.917 | 5.167 | 2.667 | NM | NM | 31.604 |
| original22-v1/threads/notify-empty | 5.417 | 5.583 | 2.501 | NM | NM | 35.896 |
| original22-v1/threads/wait-mismatch | 4.875 | 5.916 | 2.604 | NM | NM | 29.604 |
| original22-v1/threads/wait-timeout | 5.667 | 5.958 | 2.917 | NM | NM | 35.646 |
| original22-v1/applications/sha256-clang | 4.729 | 5.354 | 3.062 | NM | NM | 32.187 |
| original22-v1/applications/sha256-rust | 6.104 | 5.688 | 2.479 | NM | NM | 32.667 |
| gc-sustained-v1/gc/i31-roundtrip | 4.688 | 5.333 | US | NM | NM | 31.625 |
| gc-sustained-v1/gc/struct-access | 5.125 | 5.688 | US | NM | NM | 29.813 |
| gc-sustained-v1/gc/struct-allocation | 5.562 | 6.146 | US | NM | NM | 32.896 |
| gc-sustained-v1/gc/struct-indirect-access | 6.354 | 5.354 | US | NM | NM | 33.438 |
| threads-sustained-v2/threads/private-memory | 5.938 | 5.813 | 3.271 | NM | NM | 33.770 |
| threads-sustained-v2/threads/shared-memory | 6.416 | 5.875 | 3.354 | NM | NM | 34.458 |
| threads-sustained-v2/threads/atomic-rmw | 5.625 | 5.854 | 2.667 | NM | NM | 30.208 |
| threads-sustained-v2/threads/notify-empty | 5.208 | 5.938 | 2.771 | NM | NM | 32.750 |
| threads-sustained-v2/threads/wait-mismatch | 6.021 | 7.896 | 2.854 | NM | NM | 31.895 |
| threads-sustained-v2/threads/wait-timeout | 5.896 | 5.479 | 2.646 | NM | NM | 31.562 |
| generic-calls-v1/core/generic-calls | 4.938 | 5.229 | 2.729 | NM | NM | 33.188 |

### compileValidate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 63.541 | 53.312 | 42.396 | NM | NM | 39.728 |
| original22-v1/core/integer-loop | 64.541 | 61.667 | 46.979 | NM | NM | 36.688 |
| original22-v1/core/function-calls | 81.438 | 76.666 | 72.187 | NM | NM | 41.770 |
| original22-v1/core/memory-load-store | 90.854 | 93.625 | 56.166 | NM | NM | 42.563 |
| original22-v1/memory/narrow-unaligned | 88.583 | 92.438 | 52.708 | NM | NM | 45.542 |
| original22-v1/memory/dependent-address | 90.542 | 90.291 | 55.959 | NM | NM | 44.522 |
| original22-v1/memory/float64-unaligned | 89.959 | 100.94 | 59.291 | NM | NM | 48.771 |
| original22-v1/runtime/simd-memory-full | 84.791 | 84.312 | 59.188 | NM | NM | 48.583 |
| original22-v1/runtime/simd-memory-widen | 82.666 | 74.105 | 56.312 | NM | NM | 44.959 |
| original22-v1/runtime/simd-memory-lanes | 95.624 | 80.583 | 54.416 | NM | NM | 45.917 |
| original22-v1/gc/i31-roundtrip | 65.354 | 55.625 | US | NM | NM | 36.353 |
| original22-v1/gc/struct-access | 69.271 | 60.291 | US | NM | NM | 45.104 |
| original22-v1/gc/struct-allocation | 76.980 | 63.605 | US | NM | NM | 51.438 |
| original22-v1/gc/struct-indirect-access | 83.020 | 84.083 | US | NM | NM | 53.271 |
| original22-v1/threads/private-memory | 69.855 | 63.874 | 51.958 | NM | NM | 44.146 |
| original22-v1/threads/shared-memory | 66.020 | 61.479 | 46.376 | NM | NM | 42.209 |
| original22-v1/threads/atomic-rmw | 61.729 | 57.874 | 47.959 | NM | NM | 43.313 |
| original22-v1/threads/notify-empty | 69.874 | 56.645 | 46.521 | NM | NM | 42.020 |
| original22-v1/threads/wait-mismatch | 69.709 | 63.958 | 47.771 | NM | NM | 45.062 |
| original22-v1/threads/wait-timeout | 74.270 | 71.812 | 52.062 | NM | NM | 40.458 |
| original22-v1/applications/sha256-clang | 298.88 | 362.98 | 91.562 | NM | NM | 90.959 |
| original22-v1/applications/sha256-rust | 307.77 | 460.33 | 89.729 | NM | NM | 92.041 |
| gc-sustained-v1/gc/i31-roundtrip | 67.000 | 56.249 | US | NM | NM | 38.188 |
| gc-sustained-v1/gc/struct-access | 73.125 | 69.770 | US | NM | NM | 46.688 |
| gc-sustained-v1/gc/struct-allocation | 81.583 | 67.167 | US | NM | NM | 49.188 |
| gc-sustained-v1/gc/struct-indirect-access | 95.041 | 92.042 | US | NM | NM | 54.166 |
| threads-sustained-v2/threads/private-memory | 73.812 | 64.167 | 51.020 | NM | NM | 46.313 |
| threads-sustained-v2/threads/shared-memory | 73.250 | 70.312 | 53.980 | NM | NM | 44.666 |
| threads-sustained-v2/threads/atomic-rmw | 64.271 | 57.541 | 48.895 | NM | NM | 42.895 |
| threads-sustained-v2/threads/notify-empty | 67.520 | 60.958 | 46.729 | NM | NM | 43.291 |
| threads-sustained-v2/threads/wait-mismatch | 81.709 | 73.250 | 58.625 | NM | NM | 42.646 |
| threads-sustained-v2/threads/wait-timeout | 72.334 | 64.167 | 53.688 | NM | NM | 42.938 |
| generic-calls-v1/core/generic-calls | 79.084 | 80.020 | 44.146 | NM | NM | 46.834 |

### instantiate

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 7.333 | 8.104 | 18.459 | NM | NM | 29.375 |
| original22-v1/core/integer-loop | 7.125 | 8.313 | 16.875 | NM | NM | 31.291 |
| original22-v1/core/function-calls | 7.979 | 8.479 | 17.855 | NM | NM | 29.062 |
| original22-v1/core/memory-load-store | 12.521 | 13.001 | 19.334 | NM | NM | 51.104 |
| original22-v1/memory/narrow-unaligned | 12.688 | 13.541 | 17.416 | NM | NM | 53.646 |
| original22-v1/memory/dependent-address | 12.834 | 15.000 | 19.645 | NM | NM | 55.521 |
| original22-v1/memory/float64-unaligned | 12.917 | 15.312 | 20.104 | NM | NM | 60.292 |
| original22-v1/runtime/simd-memory-full | 13.666 | 13.792 | 20.229 | NM | NM | 61.438 |
| original22-v1/runtime/simd-memory-widen | 14.104 | 13.312 | 18.750 | NM | NM | 65.041 |
| original22-v1/runtime/simd-memory-lanes | 13.479 | 13.479 | 21.958 | NM | NM | 61.812 |
| original22-v1/gc/i31-roundtrip | 7.479 | 7.896 | US | NM | NM | 30.459 |
| original22-v1/gc/struct-access | 7.750 | 8.417 | US | NM | NM | 25.395 |
| original22-v1/gc/struct-allocation | 8.374 | 9.999 | US | NM | NM | 37.541 |
| original22-v1/gc/struct-indirect-access | 13.812 | 17.208 | US | NM | NM | 35.541 |
| original22-v1/threads/private-memory | 11.729 | 13.791 | 19.770 | NM | NM | 56.792 |
| original22-v1/threads/shared-memory | 15.624 | 14.229 | 17.250 | NM | NM | 57.480 |
| original22-v1/threads/atomic-rmw | 16.000 | 13.688 | 18.480 | NM | NM | 54.709 |
| original22-v1/threads/notify-empty | 17.500 | 14.249 | 17.979 | NM | NM | 53.626 |
| original22-v1/threads/wait-mismatch | 16.208 | 14.063 | 17.688 | NM | NM | 52.916 |
| original22-v1/threads/wait-timeout | 16.895 | 15.416 | 18.896 | NM | NM | 53.834 |
| original22-v1/applications/sha256-clang | 16.021 | 19.625 | 24.334 | NM | NM | 66.397 |
| original22-v1/applications/sha256-rust | 17.896 | 83.937 | 26.250 | NM | NM | 63.854 |
| gc-sustained-v1/gc/i31-roundtrip | 8.124 | 8.209 | US | NM | NM | 30.916 |
| gc-sustained-v1/gc/struct-access | 8.041 | 9.250 | US | NM | NM | 28.104 |
| gc-sustained-v1/gc/struct-allocation | 8.709 | 9.979 | US | NM | NM | 38.771 |
| gc-sustained-v1/gc/struct-indirect-access | 15.188 | 17.563 | US | NM | NM | 37.999 |
| threads-sustained-v2/threads/private-memory | 12.396 | 13.541 | 22.062 | NM | NM | 56.458 |
| threads-sustained-v2/threads/shared-memory | 19.125 | 15.791 | 20.250 | NM | NM | 55.791 |
| threads-sustained-v2/threads/atomic-rmw | 17.812 | 13.854 | 18.291 | NM | NM | 50.459 |
| threads-sustained-v2/threads/notify-empty | 17.000 | 13.501 | 17.729 | NM | NM | 52.499 |
| threads-sustained-v2/threads/wait-mismatch | 17.083 | 15.229 | 19.312 | NM | NM | 58.229 |
| threads-sustained-v2/threads/wait-timeout | 18.438 | 14.458 | 19.438 | NM | NM | 61.249 |
| generic-calls-v1/core/generic-calls | 8.354 | 10.292 | 18.459 | NM | NM | 29.834 |

### exportLookup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 0.250 | 0.375 | 0.208 | NM | NM | 0.230 |
| original22-v1/core/integer-loop | 0.250 | 0.417 | 0.230 | NM | NM | 0.230 |
| original22-v1/core/function-calls | 0.291 | 0.396 | 0.312 | NM | NM | 0.208 |
| original22-v1/core/memory-load-store | 0.250 | 0.416 | 0.167 | NM | NM | 0.229 |
| original22-v1/memory/narrow-unaligned | 0.271 | 0.396 | 0.229 | NM | NM | 0.208 |
| original22-v1/memory/dependent-address | 0.292 | 0.438 | 0.271 | NM | NM | 0.209 |
| original22-v1/memory/float64-unaligned | 0.271 | 0.438 | 0.271 | NM | NM | 0.250 |
| original22-v1/runtime/simd-memory-full | 0.250 | 0.500 | 0.208 | NM | NM | 0.230 |
| original22-v1/runtime/simd-memory-widen | 0.292 | 0.416 | 0.230 | NM | NM | 0.250 |
| original22-v1/runtime/simd-memory-lanes | 0.333 | 0.375 | 0.250 | NM | NM | 0.230 |
| original22-v1/gc/i31-roundtrip | 0.292 | 0.396 | US | NM | NM | 0.208 |
| original22-v1/gc/struct-access | 0.271 | 0.375 | US | NM | NM | 0.208 |
| original22-v1/gc/struct-allocation | 0.250 | 0.437 | US | NM | NM | 0.229 |
| original22-v1/gc/struct-indirect-access | 0.250 | 0.375 | US | NM | NM | 0.208 |
| original22-v1/threads/private-memory | 0.271 | 0.458 | 0.250 | NM | NM | 0.250 |
| original22-v1/threads/shared-memory | 0.292 | 0.417 | 0.250 | NM | NM | 0.230 |
| original22-v1/threads/atomic-rmw | 0.291 | 0.396 | 0.230 | NM | NM | 0.209 |
| original22-v1/threads/notify-empty | 0.271 | 0.438 | 0.208 | NM | NM | 0.208 |
| original22-v1/threads/wait-mismatch | 0.291 | 0.438 | 0.230 | NM | NM | 0.229 |
| original22-v1/threads/wait-timeout | 0.250 | 0.417 | 0.291 | NM | NM | 0.209 |
| original22-v1/applications/sha256-clang | 0.292 | 0.416 | 0.250 | NM | NM | 0.229 |
| original22-v1/applications/sha256-rust | 0.291 | 0.479 | 0.208 | NM | NM | 0.250 |
| gc-sustained-v1/gc/i31-roundtrip | 0.250 | 0.416 | US | NM | NM | 0.230 |
| gc-sustained-v1/gc/struct-access | 0.292 | 0.354 | US | NM | NM | 0.167 |
| gc-sustained-v1/gc/struct-allocation | 0.250 | 0.417 | US | NM | NM | 0.230 |
| gc-sustained-v1/gc/struct-indirect-access | 0.291 | 0.417 | US | NM | NM | 0.208 |
| threads-sustained-v2/threads/private-memory | 0.333 | 0.417 | 0.271 | NM | NM | 0.250 |
| threads-sustained-v2/threads/shared-memory | 0.292 | 0.459 | 0.271 | NM | NM | 0.250 |
| threads-sustained-v2/threads/atomic-rmw | 0.291 | 0.375 | 0.229 | NM | NM | 0.208 |
| threads-sustained-v2/threads/notify-empty | 0.250 | 0.416 | 0.250 | NM | NM | 0.250 |
| threads-sustained-v2/threads/wait-mismatch | 0.292 | 0.479 | 0.250 | NM | NM | 0.250 |
| threads-sustained-v2/threads/wait-timeout | 0.291 | 0.459 | 0.271 | NM | NM | 0.209 |
| generic-calls-v1/core/generic-calls | 0.271 | 0.416 | 0.271 | NM | NM | 0.250 |

### firstCall

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 6.792 | 7.312 | 4.167 | NM | NM | 7.979 |
| original22-v1/core/integer-loop | 33,800.06 | 27,373.50 | 93,572.08 | NM | NM | 53,048.56 |
| original22-v1/core/function-calls | 63,713.04 | 53,351.98 | 130,168.29 | NM | NM | 76,047.85 |
| original22-v1/core/memory-load-store | 81,027.15 | 65,639.98 | 169,001.06 | NM | NM | 119,372.06 |
| original22-v1/memory/narrow-unaligned | 97,118.27 | 97,888.50 | 200,538.90 | NM | NM | 126,712.08 |
| original22-v1/memory/dependent-address | 52,507.67 | 42,699.56 | 128,780.33 | NM | NM | 71,575.58 |
| original22-v1/memory/float64-unaligned | 87,525.40 | 79,877.10 | 166,842.31 | NM | NM | 101,335.67 |
| original22-v1/runtime/simd-memory-full | 91,516.58 | 86,834.40 | 196,093.44 | NM | NM | 99,508.19 |
| original22-v1/runtime/simd-memory-widen | 75,742.27 | 93,927.38 | 191,359.33 | NM | NM | 72,665.13 |
| original22-v1/runtime/simd-memory-lanes | 122,119.58 | 116,506.54 | 217,683.19 | NM | NM | 113,768.42 |
| original22-v1/gc/i31-roundtrip | 428.92 | 380.08 | US | NM | NM | 497.71 |
| original22-v1/gc/struct-access | 471.04 | 419.42 | US | NM | NM | 630.46 |
| original22-v1/gc/struct-allocation | 1,126.79 | 1,182.48 | US | NM | NM | 1,630.31 |
| original22-v1/gc/struct-indirect-access | 865.06 | 811.88 | US | NM | NM | 1,135.27 |
| original22-v1/threads/private-memory | 538.35 | 466.92 | 799.58 | NM | NM | 552.56 |
| original22-v1/threads/shared-memory | 877.31 | 412.73 | 806.40 | NM | NM | 475.12 |
| original22-v1/threads/atomic-rmw | 573.48 | 377.52 | 511.06 | NM | NM | 389.23 |
| original22-v1/threads/notify-empty | 406.08 | 339.52 | 502.38 | NM | NM | 487.29 |
| original22-v1/threads/wait-mismatch | 667.67 | 461.50 | 860.92 | NM | NM | 551.96 |
| original22-v1/threads/wait-timeout | 741.29 | 488.58 | 4,209.23 | NM | NM | 1,451.06 |
| original22-v1/applications/sha256-clang | 110,594.10 | 97,350.29 | 249,758.48 | NM | NM | 159,759.29 |
| original22-v1/applications/sha256-rust | 118,629.06 | 103,533.08 | 243,059.88 | NM | NM | 149,039.96 |
| gc-sustained-v1/gc/i31-roundtrip | 40,439.67 | 35,301.10 | US | NM | NM | 56,218.92 |
| gc-sustained-v1/gc/struct-access | 48,087.60 | 44,982.38 | US | NM | NM | 64,394.94 |
| gc-sustained-v1/gc/struct-allocation | 59,985.04 | 39,359.27 | US | NM | NM | 38,168.67 |
| gc-sustained-v1/gc/struct-indirect-access | 99,095.62 | 82,537.75 | US | NM | NM | 128,684.94 |
| threads-sustained-v2/threads/private-memory | 5,112.25 | 4,690.85 | 8,439.54 | NM | NM | 5,282.10 |
| threads-sustained-v2/threads/shared-memory | 9,397.75 | 4,737.56 | 8,229.08 | NM | NM | 5,096.25 |
| threads-sustained-v2/threads/atomic-rmw | 6,124.33 | 3,832.67 | 5,370.33 | NM | NM | 3,899.96 |
| threads-sustained-v2/threads/notify-empty | 3,548.60 | 2,964.00 | 4,580.96 | NM | NM | 4,431.83 |
| threads-sustained-v2/threads/wait-mismatch | 6,971.48 | 5,051.96 | 9,397.81 | NM | NM | 5,711.96 |
| threads-sustained-v2/threads/wait-timeout | 6,801.29 | 4,090.12 | 40,895.67 | NM | NM | 12,391.35 |
| generic-calls-v1/core/generic-calls | 642,297.90 | 546,715.67 | 1,076,634.33 | NM | NM | 724,005.46 |

### cleanup

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 5.875 | 1.375 | 7.833 | NM | NM | 8.084 |
| original22-v1/core/integer-loop | 7.500 | 2.062 | 6.083 | NM | NM | 8.125 |
| original22-v1/core/function-calls | 7.354 | 2.042 | 4.604 | NM | NM | 11.271 |
| original22-v1/core/memory-load-store | 2.562 | 2.167 | 8.896 | NM | NM | 50.854 |
| original22-v1/memory/narrow-unaligned | 2.604 | 2.126 | 10.834 | NM | NM | 60.416 |
| original22-v1/memory/dependent-address | 2.479 | 2.354 | 6.146 | NM | NM | 55.187 |
| original22-v1/memory/float64-unaligned | 2.562 | 2.188 | 5.625 | NM | NM | 60.166 |
| original22-v1/runtime/simd-memory-full | 2.542 | 2.187 | 9.291 | NM | NM | 59.708 |
| original22-v1/runtime/simd-memory-widen | 2.396 | 2.292 | 10.500 | NM | NM | 52.333 |
| original22-v1/runtime/simd-memory-lanes | 3.583 | 2.292 | 10.396 | NM | NM | 58.562 |
| original22-v1/gc/i31-roundtrip | 5.479 | 1.792 | US | NM | NM | 7.667 |
| original22-v1/gc/struct-access | 6.541 | 1.938 | US | NM | NM | 8.709 |
| original22-v1/gc/struct-allocation | 119.81 | 66.438 | US | NM | NM | 2,035.50 |
| original22-v1/gc/struct-indirect-access | 5.563 | 1.896 | US | NM | NM | 7.917 |
| original22-v1/threads/private-memory | 2.250 | 2.126 | 8.541 | NM | NM | 52.874 |
| original22-v1/threads/shared-memory | 16.562 | 6.583 | 8.229 | NM | NM | 47.250 |
| original22-v1/threads/atomic-rmw | 15.625 | 7.854 | 7.438 | NM | NM | 46.417 |
| original22-v1/threads/notify-empty | 8.938 | 4.542 | 4.208 | NM | NM | 45.666 |
| original22-v1/threads/wait-mismatch | 12.479 | 7.875 | 4.917 | NM | NM | 50.604 |
| original22-v1/threads/wait-timeout | 20.708 | 9.209 | 7.125 | NM | NM | 49.208 |
| original22-v1/applications/sha256-clang | 2.542 | 2.354 | 8.666 | NM | NM | 64.041 |
| original22-v1/applications/sha256-rust | 2.542 | 2.438 | 11.375 | NM | NM | 67.166 |
| gc-sustained-v1/gc/i31-roundtrip | 5.917 | 2.042 | US | NM | NM | 6.938 |
| gc-sustained-v1/gc/struct-access | 7.938 | 2.250 | US | NM | NM | 8.334 |
| gc-sustained-v1/gc/struct-allocation | 3,341.77 | 2,191.50 | US | NM | NM | 73,982.90 |
| gc-sustained-v1/gc/struct-indirect-access | 7.688 | 2.251 | US | NM | NM | 10.375 |
| threads-sustained-v2/threads/private-memory | 2.562 | 2.292 | 10.416 | NM | NM | 55.459 |
| threads-sustained-v2/threads/shared-memory | 13.188 | 11.250 | 7.896 | NM | NM | 50.416 |
| threads-sustained-v2/threads/atomic-rmw | 9.750 | 5.229 | 4.792 | NM | NM | 44.562 |
| threads-sustained-v2/threads/notify-empty | 14.000 | 5.979 | 4.625 | NM | NM | 52.584 |
| threads-sustained-v2/threads/wait-mismatch | 16.875 | 11.438 | 7.625 | NM | NM | 49.896 |
| threads-sustained-v2/threads/wait-timeout | 23.125 | 15.562 | 3.875 | NM | NM | 51.229 |
| generic-calls-v1/core/generic-calls | 9.042 | 2.062 | 4.646 | NM | NM | 10.459 |

</details>

### WasmEdge AOT preparation detail

| Workload | Compile + validation + artifact I/O, us | Native image load, us | Post-load validation, us |
| --- | ---: | ---: | ---: |
| original22-v1/core/call-control | 26,413.25 | 263,138.67 | 3.499 |
| original22-v1/core/integer-loop | 22,035.98 | 263,112.35 | 5.688 |
| original22-v1/core/function-calls | 23,542.69 | 262,270.06 | 5.417 |
| original22-v1/core/memory-load-store | 25,153.92 | 261,249.81 | 7.083 |
| original22-v1/memory/narrow-unaligned | 20,617.73 | 262,101.21 | 9.062 |
| original22-v1/memory/dependent-address | 24,744.42 | 261,741.46 | 8.791 |
| original22-v1/memory/float64-unaligned | 26,730.48 | 262,073.88 | 7.833 |
| original22-v1/runtime/simd-memory-full | 25,184.04 | 262,223.08 | 7.104 |
| original22-v1/runtime/simd-memory-widen | 24,051.00 | 261,176.60 | 7.854 |
| original22-v1/runtime/simd-memory-lanes | 22,654.71 | 261,332.04 | 7.583 |
| original22-v1/gc/i31-roundtrip | 22,577.31 | 262,484.87 | 5.875 |
| original22-v1/gc/struct-access | 22,343.52 | 264,332.40 | 3.583 |
| original22-v1/gc/struct-allocation | 24,682.67 | 259,388.75 | 6.229 |
| original22-v1/gc/struct-indirect-access | 23,567.56 | 264,564.88 | 7.083 |
| original22-v1/threads/private-memory | 25,067.73 | 262,492.83 | 6.812 |
| original22-v1/threads/shared-memory | 22,555.69 | 266,562.58 | 5.500 |
| original22-v1/threads/atomic-rmw | 22,158.98 | 267,120.87 | 5.500 |
| original22-v1/threads/notify-empty | 22,795.38 | 263,484.79 | 5.271 |
| original22-v1/threads/wait-mismatch | 21,316.52 | 260,291.73 | 6.667 |
| original22-v1/threads/wait-timeout | 24,408.46 | 265,765.00 | 6.000 |
| original22-v1/applications/sha256-clang | 28,248.38 | 259,075.40 | 11.334 |
| original22-v1/applications/sha256-rust | 32,735.67 | 267,080.29 | 11.479 |
| gc-sustained-v1/gc/i31-roundtrip | 20,871.21 | 262,176.96 | 7.708 |
| gc-sustained-v1/gc/struct-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-allocation | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| gc-sustained-v1/gc/struct-indirect-access | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** | **FAIL: SIGSEGV** |
| threads-sustained-v2/threads/private-memory | 24,609.98 | 266,407.54 | 6.458 |
| threads-sustained-v2/threads/shared-memory | 24,798.10 | 263,776.12 | 5.625 |
| threads-sustained-v2/threads/atomic-rmw | 23,017.98 | 260,653.15 | 5.709 |
| threads-sustained-v2/threads/notify-empty | 22,815.15 | 261,781.04 | 5.417 |
| threads-sustained-v2/threads/wait-mismatch | 24,438.60 | 259,377.85 | 5.542 |
| threads-sustained-v2/threads/wait-timeout | 23,685.02 | 266,413.90 | 6.479 |
| generic-calls-v1/core/generic-calls | 24,576.60 | 263,611.81 | 5.938 |

### P3 phase observations

| Workload / metric | wasmvm Go | wasmvm LLGo | wazero | Wasmtime | Wasmer | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B / cleanup | 69.479 | 751.35 | NM | 321.29 | NM | NM |
| 16 B / compileValidate | 67,101.12 | 108,949.90 | NM | 59,152.29 | NM | NM |
| 16 B / engineSetup | 0.520 | 0.229 | NM | 154.85 | NM | NM |
| 16 B / executorSetup | N/A | N/A | NM | 47.605 | NM | NM |
| 16 B / exportLookup | 0.916 | 1.042 | NM | 9.854 | NM | NM |
| 16 B / firstCall | 11,687.90 | 25,363.23 | NM | 10,393.46 | NM | NM |
| 16 B / instantiate | 552.46 | 869.19 | NM | 207.35 | NM | NM |
| 16 B / load | 84.666 | 87.730 | NM | 92.605 | NM | NM |
| 16 B / processStartup | 7,162.25 | 13,564.08 | NM | 7,509.69 | NM | NM |
| 16 B / providerSetup | 59.792 | 272.10 | NM | 243.10 | NM | NM |
| 16 B / requestPreparation | 0.729 | 0.750 | NM | N/A | NM | NM |
| 65536 B / cleanup | 70.604 | 693.40 | NM | 291.75 | NM | NM |
| 65536 B / compileValidate | 67,690.88 | 110,451.96 | NM | 58,866.02 | NM | NM |
| 65536 B / engineSetup | 0.459 | 0.312 | NM | 168.73 | NM | NM |
| 65536 B / executorSetup | N/A | N/A | NM | 50.312 | NM | NM |
| 65536 B / exportLookup | 0.771 | 1.167 | NM | 9.792 | NM | NM |
| 65536 B / firstCall | 15,634.87 | 29,554.98 | NM | 12,265.50 | NM | NM |
| 65536 B / instantiate | 539.77 | 932.02 | NM | 208.94 | NM | NM |
| 65536 B / load | 64.896 | 83.688 | NM | 85.271 | NM | NM |
| 65536 B / processStartup | 7,081.90 | 13,846.25 | NM | 7,360.42 | NM | NM |
| 65536 B / providerSetup | 57.084 | 279.69 | NM | 247.04 | NM | NM |
| 65536 B / requestPreparation | 7.938 | 11.687 | NM | N/A | NM | NM |

## Native controls (different contracts)

Unchecked C kernels omit runtime safety/lifetime contracts; native Go implements HTTP replies without Component/WASI. These are overhead oracles, not equivalent competitors, and are never subtracted from runtime timings.

| Workload | Unchecked native C, ns/iteration |
| --- | ---: |
| original22-v1/core/integer-loop | 1.771 |
| original22-v1/core/function-calls | 0.978 |
| original22-v1/core/memory-load-store | 0.315 |
| original22-v1/memory/narrow-unaligned | 0.511 |
| original22-v1/memory/dependent-address | 1.916 |
| original22-v1/memory/float64-unaligned | 0.787 |
| original22-v1/runtime/simd-memory-full | 4.395 |
| original22-v1/runtime/simd-memory-widen | 1.293 |
| original22-v1/runtime/simd-memory-lanes | 5.190 |
| original22-v1/applications/sha256-clang | 215.81 |
| original22-v1/applications/sha256-rust | 207.46 |
| generic-calls-v1/core/generic-calls | 2.496 |

| Payload | Native Go CPU, us/request | Native Go requests/s |
| --- | ---: | ---: |
| 16 B | 3.299 | 1,165,910.02 |
| 65536 B | 81.873 | 30,688.26 |

## WasmEdge sustained AOT failures

The initial core campaign stopped on sustained indirect struct access. All incomplete timing rows are retained separately and excluded. Full-size semantic checks then found SIGSEGV in three AOT cells: sustained struct access, struct allocation and indirect struct access. Their small original cases pass. The indirect case reproduces three times in the C harness; WasmEdge’s own CLI also fails at 1,000,000 iterations (SIGILL in that invocation), while its interpreter returns the expected checksum. LLDB catches `EXC_BAD_ACCESS` in `WasmEdge::Executor::Executor::arrayGet`, on `strb w12, [x8]`. This identifies a failure site, not a proven root cause. The complete core cohort was restarted once with those exact cells labeled failed; independently completed P3 cohorts were retained unchanged.

## Next implementation work

1. General native calls: inspect actual emitted frame/parameter/root bookkeeping versus competitor output. Previous record-cursor variants regressed and remain rejected. WasmEdge AOT is now a measured strong reference on this case; its inlining/code shape must be checked before attributing the entire ratio to an ABI.
2. SHA and field access: improve compiler range/region coverage and invariant heap-reference guards while preserving trap order, cold resumes, full reference metadata and GC roots.
3. Shared memory: investigate the measured code and execution counters. The current short loop polls control state every iteration; the parked four-copy grouping experiment did not demonstrate a qualified gain. No mutex-only explanation is established.
4. P3: retain PR #200’s reduced ready-source work and continue canonical/host-transition work. These measurements show a large instruction gap; Go’s scheduler alone does not guarantee a faster complete Component/WASI request. Separate LLGo generated host code, allocation, runtime synchronization and guest-native execution before filing compiler issues.

## Reproducible evidence

- [Machine-readable results, all controls and source/SDK contrasts](data/pr200-complete-comparison-2026-11-17.json).
- [All 6,324 verified process records](data/pr200-complete-comparison-2026-11-17-raw.jsonl.gz), SHA-256 `de2f4c5ccb477cdec62c38180eaf03894c2acaf11982b0f281699bf59b80a388`.
- [Frozen plans, commands, hashes, analysis/export scripts, semantic checks and failure evidence](data/pr200-complete-comparison-2026-11-17-protocol.tar.gz), SHA-256 `0860f9f03f084dea6a5eea87b2ecce3d51e5517372becae4846ca8af760273c3`.
- Per-request vectors unused by these metrics are projected to count/extrema/hash; no latency-tail claim. Public load records omit unrelated desktop command arguments. The incomplete first core run is retained locally with its source hash in the JSON.
