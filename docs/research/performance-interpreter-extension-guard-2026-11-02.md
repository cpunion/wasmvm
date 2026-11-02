# Grouped interpreter extension checks: 2026-11-02

This compact candidate improves every non-control LLGo interpreter case in the
completed local campaign, typically about 2–7%. It retains the original scalar
switch layout, places existing extension classifiers behind one guard and loads
the third operand only in select/extended paths. The earlier repeatable 17%
private-memory regression is eliminated. No instruction implementation, fuel
boundary, cancellation poll, trap, operand validation or continuation is removed.

**Integration and controlled acceptance remain pending.** Some short native
controls regress. All unfavorable observations remain in the tables and evidence.

## Source and method

Base `38692eae91174042aa56553c9da286afa4443c79` is integrated main after PR #97.
Candidate `e76f172a371482b721812367005a5d31309c8d9a` differs in the common
dispatcher. Both include the integrated native-memory transformation; its gain
is not attributed to this interpreter change. The [two earlier complete
experiments](performance-interpreter-dispatch-experiments-2026-11-02.md) remain held.

Apple M4 Max/macOS arm64; Go 1.27.3; LLGo `8e45ed5` `-O3 -lto=full`;
Wasmtime 48.0.1/Wasmer 7.3.0 Cranelift, wazero 1.12.0 compiler/interpreter,
WasmEdge 0.17.1 interpreter. Engine, executable, compiler, SDK and fixture hashes
are checked. Twenty unchanged cases each run three fresh serially interleaved
blocks, two warmups and seven warm calls. Twelve loops use one million iterations;
four GC and three notify/wait kernels retain ten thousand; call-control performs
one call. The known private regression runs first, declared before measurement.
Exact baseline Go/LLGo interpreter binaries also run as independent A/A aliases.
Six applicable C scalar controls use the same inputs/counts. All original and
alternate independent checks pass; twelve unsupported GC/reference outcomes
remain explicit rejections, not executions.

Readiness, load, setup, preparation, instantiation, lookup, first call, warm calls
and cleanup remain separate. There is no startup subtraction. Host load changes
16.0→7.5 on 16 CPUs; five/fifteen-minute load 12.9/14.3→9.1/12.0. No affinity
or isolation is used. Empty controls are noisy; A/A GC allocation and i31 controls
vary several percent, so not every small improvement exceeds observed variation.
No calibrated parity or controlled performance acceptance is inferred.

## Complete interpreter comparison

Units: ns/source iteration, except call-control (ns/call).

| Workload | Go base | Head | LLGo base | Head | WasmEdge interp | wazero interp |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 3291.000 | 2583.000 | 3583.000 | 3209.000 | 3000.000 | 833.000 |
| core/function-calls | 95.588 | 85.677 | 93.058 | 91.205 | 62.140 | 79.668 |
| core/integer-loop | 72.678 | 64.381 | 62.388 | 59.633 | 48.801 | 81.774 |
| core/memory-load-store | 141.872 | 124.611 | 119.888 | 113.400 | 90.876 | 143.266 |
| gc/i31-roundtrip | 71.633 | 67.362 | 62.150 | 59.008 | 45.400 | — |
| gc/struct-access | 88.958 | 77.838 | 73.650 | 68.604 | 55.500 | — |
| gc/struct-allocation | 147.954 | 135.408 | 170.229 | 163.108 | 107.800 | — |
| gc/struct-indirect-access | 144.354 | 127.754 | 124.492 | 118.771 | 97.100 | — |
| memory/dependent-address | 84.456 | 73.614 | 72.641 | 67.428 | 53.212 | 95.319 |
| memory/float64-unaligned | 136.269 | 118.473 | 115.976 | 108.178 | 84.366 | 135.561 |
| memory/narrow-unaligned | 176.374 | 157.149 | 149.995 | 139.783 | 113.281 | 184.376 |
| runtime/simd-memory-full | 143.217 | 132.441 | 122.066 | 119.446 | 77.136 | 148.893 |
| runtime/simd-memory-lanes | 173.430 | 164.170 | 156.244 | 152.519 | 89.848 | 185.342 |
| runtime/simd-memory-widen | 135.024 | 122.874 | 125.038 | 121.416 | 60.622 | 142.484 |
| threads/atomic-rmw | 74.886 | 69.486 | 66.924 | 65.319 | 31.783 | 47.841 |
| threads/notify-empty | 57.462 | 51.067 | 57.054 | 55.417 | 40.900 | 42.742 |
| threads/private-memory | 67.143 | 59.323 | 57.133 | 55.171 | 38.338 | 61.272 |
| threads/shared-memory | 99.904 | 92.695 | 56.893 | 54.403 | 38.148 | 61.844 |
| threads/wait-mismatch | 87.775 | 82.487 | 77.688 | 74.317 | 44.100 | 68.737 |
| threads/wait-timeout | 88.875 | 82.508 | 78.500 | 74.263 | 98.200 | 290.571 |

LLGo private memory improves 57.133→55.171, integer 62.388→59.633,
scalar memory 119.888→113.400 and RMW 66.924→65.319. Gains are smaller than
some held layout experiments, but avoid their private-loop regression. Go
usually improves more. RMW remains 2.05x WasmEdge and 1.37x wazero interpreter;
calls still trail both. Shared/private memory and several scalar cases beat
wazero; most still trail WasmEdge. Zero-timeout wait retains its advantage.
These outcomes do not establish an accepted portfolio score.

## Complete native/C comparison

The guest compiler is unchanged between these sources, but API/helper work
still uses the common dispatcher. Units are as above. C has fewer safety/setup
contracts; missing VM phases are absent, not invented zeroes.

| Workload | Go base | Head | LLGo base | Head | Wasmtime | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 5458.000 | 5125.000 | 7500.000 | 6542.000 | 1000.000 | 3000.000 | — |
| core/function-calls | 0.804 | 0.807 | 0.792 | 0.788 | 0.906 | 0.781 | 0.776 |
| core/integer-loop | 1.585 | 1.587 | 1.560 | 1.562 | 1.557 | 1.554 | 1.548 |
| core/memory-load-store | 0.489 | 0.492 | 0.478 | 0.475 | 0.476 | 0.536 | 0.269 |
| gc/i31-roundtrip | 3.413 | 3.438 | 3.400 | 3.388 | 2.000 | — | — |
| gc/struct-access | 2.596 | 3.013 | 2.667 | 2.508 | 1.500 | — | — |
| gc/struct-allocation | 47.812 | 47.521 | 64.000 | 61.221 | 106.400 | — | — |
| gc/struct-indirect-access | 6.167 | 6.296 | 5.654 | 5.763 | 3.000 | — | — |
| memory/dependent-address | 1.873 | 1.854 | 1.842 | 1.842 | 2.071 | 2.096 | 1.404 |
| memory/float64-unaligned | 0.742 | 0.737 | 0.723 | 0.718 | 0.716 | 0.716 | 0.661 |
| memory/narrow-unaligned | 0.712 | 0.714 | 0.696 | 0.698 | 0.693 | 0.729 | 0.459 |
| runtime/simd-memory-full | 3.373 | 3.372 | 3.334 | 3.310 | 3.614 | 3.455 | — |
| runtime/simd-memory-lanes | 5.030 | 4.994 | 5.003 | 4.882 | 5.762 | 5.838 | — |
| runtime/simd-memory-widen | 1.065 | 1.059 | 1.056 | 1.050 | 1.037 | 1.039 | — |
| threads/atomic-rmw | 1.831 | 1.837 | 1.816 | 1.806 | 1.797 | 3.771 | — |
| threads/notify-empty | 2.217 | 1.925 | 1.758 | 1.962 | 7.100 | 5.200 | — |
| threads/private-memory | 0.343 | 0.340 | 0.332 | 0.330 | 0.982 | 1.038 | — |
| threads/shared-memory | 0.849 | 0.855 | 0.844 | 0.895 | 1.016 | 0.988 | — |
| threads/wait-mismatch | 1.954 | 2.671 | 1.938 | 1.950 | 9.200 | 3.200 | — |
| threads/wait-timeout | 2.708 | 2.400 | 2.112 | 1.996 | 43.000 | 169.900 | — |

LLGo native notify regresses 1.758→1.962 and shared memory 0.844→0.895.
These do not disappear behind interpreter gains. Short native intervals have no
native A/A calibration here; unchanged guest compute code does not prove all
host-boundary work is unchanged. The [long native/C atomic probe](performance-native-atomic-c-2026-11-02.md)
is a separate experiment. Native scalar memory is 0.475 versus C 0.269 (about
1.77x) and Wasmtime 0.476 on this ARM64 run. The larger Linux memory gap from
PR #97 remains open and cannot be replaced by this local table.

## Validation

At the measured candidate, all 31 Go packages, all 30 cross-package coverage
gates, pinned LLGo execute/VM suites, scoped race, no-cgo and vet pass. Existing
complete Core, GC, wide-arithmetic and stack-switching suites exercise the
classification and exact continuation/fuel behavior on both native tiers and
interpreter. Final-head Linux amd64/macOS arm64 CI and PR benchmark remain required.

[CPU profiles and the independent channel-polling reproduction](performance-interpreter-profile-2026-11-02.md)
retain the separate [LLGo #2965](https://github.com/xgo-dev/llgo/issues/2965) gap.
No unsafe cancellation shortcut is introduced to bypass it.

[Lossless evidence](evidence/performance-interpreter-extension-guard-2026-11-02.json.gz)
contains all 294 summaries, every raw call/phase, A/A aliases, twelve unsupported
outcomes, source/build/toolchain/SDK identities, measurement code and validation.


## Follow-up: exact I32 result width

The initial final-head Linux benchmark at `081d2f6` completed successfully:
LLGo integer 88.544→86.489, guest calls 139.498→125.299 and memory
169.752→156.242 ns/iteration. Go improves 9–18% in that campaign. WasmEdge
remains faster at 53.050 / 68.529 / 104.188 respectively. All 228 phase records,
76 summaries and 19 configurations validate; [complete raw data](evidence/pr98-initial-linux-2026-11-02.json.gz)
remain evidence for that exact earlier source. Its semantic matrix was still
queued when the public result-width regression below was discovered; it does
not stand in for validation of the subsequent head.

The cached constant dispatcher previously combined both words for I32 as well
as 64-bit constants. Decoded `i32.const -1` therefore exposed
`0xffffffffffffffff` through interpreter Value.Bits, while native exposed the
correct `0xffffffff`. Minimum I32 similarly exposed unwanted upper bits.
The new public-API regression fails on the earlier interpreter in both fuel
policies and passes on the corrected interpreter/native backends under Go and
pinned LLGo. I64 full-word and F32/F64 NaN payload controls retain exact bits.

The correction reconstructs I32 only from its 32-bit Extra word; I64 and FP
constant layouts remain unchanged. Original instructions, fuel, PCs and polls
remain intact. This correctness fix is included in PR #98; the subsequent
integer-pair optimization remains on a separate contribution branch. The fixed
source `94c2c16` passes all 31 local Go packages, all 30 coverage gates, affected
pinned LLGo suites, race, no-cgo and vet. New final-head CI and benchmarking are
required; earlier results are not relabeled as that head's measurements.
[Regression and validation evidence](evidence/interpreter-i32-result-width-2026-11-02.json.gz).
