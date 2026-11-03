# Direct invocation-context polling: 2026-11-03

Status: contribution candidate; final-head dual-platform CI and performance are
pending. Controlled whole-portfolio and ordinary-native acceptance remain open.

The new LLGo host pin fixes the nil-return allocation that motivated the cached
cancellation view. Use the original invocation context directly in dispatch,
runtime helpers, VM drivers and P1 providers. Remove the wrapper, cached channel
and compiler-specific branches, with 61 fewer production/test lines overall.
Polling sites and frequency, native interrupt ownership, immutable retained
contexts, metadata, causes, nil/custom contexts and lifetime boundaries stay
intact. Avoid eagerly requesting Done for every root invocation; cancellation
still expires each host capability and retained context at its original boundary.

## Identities and method

Same compiler for source base `5c03f7d4deaf1b8fd2471ea9a1e71d3100092be3` and measured head
`d07c9e6ade60df4531076db247919d5d4998a87b`: Go 1.27.3 or LLGo `983465e`,
`-O3 -lto=full`, M4 Max/macOS arm64. This source optimization is separate from
the [compiler baseline migration](performance-llgo-result-upgrade-2026-11-03.md).
Rebased validation `da14ccf5a21c81771d3b775f6da706f635ca224e` differs from measured head only
in PR110's P3 quota-test synchronization; production source is identical.

The short-call study uses five fresh serially interleaved process blocks, two
warmups/seven samples and 10,000-call batches. It retains 485 observations,
97 summaries/alternate checks and exact baseline native public A/A aliases.
The separate twenty-kernel study uses three fresh interleaved process blocks,
two warmups/seven samples, one million loop iterations except the unchanged
10,000-iteration GC/notify/wait cases and one-call control. It retains 864 VM
observations, 18 ordinary C observations, 288 alternate checks and twelve
explicit unsupported GC-reference outcomes. Exact interpreter baseline A/A
aliases are included; the sustained native rows have no native A/A aliases.
All phases and outcomes remain separate. No subtraction or cross-study pooling
is used. Shared-host calibration does not establish controlled acceptance.

## Short-call boundaries

Warm ns/call; integer means one complete invocation of 64 guest iterations,
including its boundary, rather than sustained ns/iteration:

| Configuration | Empty, base → head | Integer 64, base → head |
| --- | ---: | ---: |
| go native-api | 326.871 → 299.933 | 409.229 → 383.754 |
| go native-api-live | 537.817 → 530.288 | 622.992 → 620.775 |
| go native-session | 214.312 → 210.379 | 300.292 → 294.688 |
| go interpreter-api | 270.946 → 236.025 | 3141.279 → 2979.904 |
| go interpreter-api-live | 308.438 → 274.246 | 3169.713 → 3008.692 |
| go interpreter-session | 154.271 → 143.233 | 3014.367 → 2924.933 |
| go interpreter-dispatch | 77.492 → 69.258 | 2929.433 → 2819.708 |
| llgo native-api | 993.179 → 852.221 | 1099.554 → 950.975 |
| llgo native-api-live | 1577.408 → 1580.071 | 1690.575 → 1668.558 |
| llgo native-session | 536.237 → 514.829 | 633.108 → 604.279 |
| llgo interpreter-api | 823.913 → 645.375 | 3581.671 → 2812.983 |
| llgo interpreter-api-live | 951.142 → 786.250 | 3713.346 → 2961.671 |
| llgo interpreter-session | 359.079 → 312.100 | 3067.496 → 2463.096 |
| llgo interpreter-dispatch | 97.333 → 67.650 | 2833.242 → 2223.371 |

LLGo native public empty improves 14.2%; interpreter public empty/short integer
improve 21.7%/21.5%. Retain native live-parent empty +0.2%. All raw native-entry,
dispatch, A/A, wazero default/enabled termination and live-parent reference
rows remain in the evidence. Public API results are never replaced by cheaper
internal components; reference cancellation policies are not asserted equivalent.

## Complete sustained interpreter comparison

Warm ns/source iteration, except the empty control in ns/call. Lower is better.
References are WasmEdge 0.17.1 interpreter, wazero 1.12.0 interpreter,
Wasmtime 48.0.1 Cranelift and Wasmer 7.3.0 Cranelift.

| Workload | Go base → head | LLGo base → head | WasmEdge | wazero |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 51.335 → 49.733 | 50.762 → 43.304 | 38.569 | 62.221 |
| core/call-control | 2875.000 → 2625.000 | 2834.000 → 2542.000 | 3000.000 | 791.000 |
| core/integer-loop | 44.273 → 41.334 | 44.361 → 35.486 | 49.188 | 81.748 |
| core/function-calls | 68.071 → 63.484 | 75.571 → 59.943 | 62.033 | 79.524 |
| core/memory-load-store | 86.347 → 84.275 | 85.740 → 75.047 | 90.784 | 143.025 |
| memory/narrow-unaligned | 110.567 → 108.970 | 109.254 → 96.948 | 113.229 | 182.204 |
| memory/dependent-address | 56.959 → 55.499 | 57.685 → 48.494 | 53.757 | 98.542 |
| memory/float64-unaligned | 89.049 → 86.496 | 87.673 → 78.207 | 84.367 | 140.451 |
| runtime/simd-memory-full | 110.685 → 110.186 | 105.551 → 97.414 | 79.756 | 152.745 |
| runtime/simd-memory-widen | 89.460 → 88.377 | 92.842 → 85.597 | 62.129 | 147.032 |
| runtime/simd-memory-lanes | 132.126 → 132.588 | 132.927 → 128.264 | 92.627 | 191.310 |
| gc/i31-roundtrip | 56.383 → 55.033 | 53.608 → 45.362 | 46.000 | — |
| gc/struct-access | 58.663 → 56.587 | 58.833 → 50.479 | 56.900 | — |
| gc/struct-allocation | 122.317 → 119.979 | 157.171 → 149.996 | 109.600 | — |
| gc/struct-indirect-access | 100.071 → 99.183 | 98.079 → 91.508 | 96.800 | — |
| threads/shared-memory | 87.343 → 84.099 | 51.485 → 43.813 | 39.258 | 63.289 |
| threads/atomic-rmw | 66.174 → 64.351 | 66.165 → 44.642 | 32.463 | 44.564 |
| threads/notify-empty | 46.229 → 44.783 | 55.587 → 40.337 | 40.800 | 42.837 |
| threads/wait-mismatch | 77.200 → 74.233 | 76.550 → 54.058 | 44.200 | 72.779 |
| threads/wait-timeout | 71.596 → 72.208 | 72.517 → 50.371 | 99.200 | 297.238 |

LLGo time decreases in all twenty cases, 3.5–32.5%. Integer/calls/memory are
35.486/59.943/75.047, versus this WasmEdge's 49.188/62.033/90.784. These
workstation observations do not establish global superiority. SIMD full/widen/
lane remain about 1.22x/1.38x/1.38x WasmEdge; allocation remains about 1.37x,
shared memory about 1.12x and RMW about 1.38x. GC allocation remains about
1.25x Go. The two reference runtimes' missing GC rows are explicit rejections.
Go regressions, including notify/wait controls, remain in the complete table.

## Complete native and ordinary C comparison

Same twenty-kernel campaign and units. C rows are separately compiled unchecked
ordinary kernels, not VM C API adapters. They omit VM bounds/cancellation/root/
quota contracts and allow ordinary compiler optimization; no C GC equivalent
is invented. GC and other missing C rows have no comparison claim.

| Workload | Go base → head | LLGo base → head | wazero | Wasmtime | Wasmer | Ordinary C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.337 → 0.334 | 0.318 → 0.321 | 1.140 | 1.013 | 1.018 | — |
| core/call-control | 3750.000 → 3959.000 | 3709.000 → 3250.000 | 416.000 | 1000.000 | 2000.000 | — |
| core/integer-loop | 1.584 → 1.589 | 1.558 → 1.555 | 1.709 | 1.566 | 1.568 | 1.555 |
| core/function-calls | 0.810 → 0.802 | 0.790 → 0.789 | 1.451 | 0.904 | 0.786 | 0.774 |
| core/memory-load-store | 0.460 → 0.455 | 0.438 → 0.441 | 0.889 | 0.479 | 0.534 | 0.270 |
| memory/narrow-unaligned | 0.714 → 0.706 | 0.689 → 0.701 | 1.059 | 0.691 | 0.722 | 0.455 |
| memory/dependent-address | 1.874 → 1.879 | 1.845 → 1.832 | 3.154 | 2.069 | 2.107 | 1.427 |
| memory/float64-unaligned | 0.743 → 0.745 | 0.722 → 0.731 | 0.881 | 0.713 | 0.715 | 0.669 |
| runtime/simd-memory-full | 3.513 → 3.475 | 3.410 → 3.439 | 3.949 | 3.694 | 3.563 | — |
| runtime/simd-memory-widen | 1.077 → 1.067 | 1.066 → 1.064 | 1.181 | 1.047 | 1.057 | — |
| runtime/simd-memory-lanes | 5.119 → 5.110 | 5.105 → 5.095 | 6.152 | 5.911 | 5.936 | — |
| gc/i31-roundtrip | 2.117 → 2.237 | 1.996 → 1.971 | — | 2.200 | — | — |
| gc/struct-access | 2.279 → 2.354 | 2.250 → 2.433 | — | 1.500 | — | — |
| gc/struct-allocation | 46.362 → 45.013 | 54.433 → 54.192 | — | 110.600 | — | — |
| gc/struct-indirect-access | 5.754 → 5.796 | 5.408 → 5.417 | — | 3.000 | — | — |
| threads/shared-memory | 0.827 → 0.814 | 0.790 → 0.807 | 1.164 | 1.031 | 0.802 | — |
| threads/atomic-rmw | 1.868 → 1.848 | 1.840 → 1.850 | 2.056 | 1.844 | 3.976 | — |
| threads/notify-empty | 1.950 → 2.196 | 1.796 → 1.792 | 21.350 | 7.200 | 5.400 | — |
| threads/wait-mismatch | 2.121 → 2.312 | 1.975 → 1.900 | 25.887 | 9.300 | 3.200 | — |
| threads/wait-timeout | 3.075 → 2.500 | 1.962 → 1.938 | 258.142 | 44.700 | 177.000 | — |

Scalar integer/calls remain close to C in this run, while ordinary C memory is
0.270 versus LLGo 0.441 ns/iteration (1.63x). Narrow/dependent/float64 memory
remain about 1.54x/1.28x/1.09x C. Native GC direct/indirect access remain about
1.62x/1.81x Wasmtime. Retain LLGo direct-struct +8.1%, narrow memory +1.7%,
float64 +1.2% and scalar memory +0.8%, plus Go i31 +5.7% and direct-struct
+3.3%. Unchanged code-generation sources do not waive any adverse observation.
No native control is rerun to replace its outcome.

## Allocation, validation and remaining gaps

Separate three-round samples (72 records, 10,000 calls/round) report native
public bytes/call: Go 408.000 → 296.000, LLGo 1968.142 → 1663.234. Go allocation
count falls from seven to six per call; LLGo Mallocs is unavailable. Retained
session remains Go about 64 and LLGo about 816 reported bytes/call. LLGo
TotalAlloc is approximate; these numbers do not attribute individual heap sites.
Initial reference-path setup failure, corrected before timing, is retained.

Affected execute/VM/P1 packages pass complete Go and full-LTO LLGo execution,
plus directed Go race checks for cancellation, context metadata, close, panic,
caller/session ownership and workspace rebinding. Full rebased Go and LLGo execution passes all 31 packages and 60 expected
conformance reports each on macOS arm64 (120 total). The five required proposal
snapshots execute on all three tiers, with forced collection for Stack Switching.
Final contribution-head cross-platform CI remains required. No unsupported
feature or package is skipped.

Keep the remaining interpreter SIMD/GC/allocation/atomic gaps, native-C memory,
public boundary, startup and expanded WASI-world acceptance open. In particular,
removing a channel-poll workaround does not fix the independently reported LLGo
channel/defer/boxing/parameter or startup problems. PR110 must finish before
this candidate becomes the next contribution PR.

[Lossless evidence](evidence/performance-direct-context-poll-2026-11-03.json.gz)
retains all sources/identities, original outcomes, independent verification,
allocation rounds, preparation/first/warm/cleanup/readiness phases and oracles.
