# Complete atomic-path comparison and held native expansion: 2026-11-02

The complete twenty-case experiment confirms LLGo interpreter shared-memory
**73.589→58.858**, RMW **86.141→67.719**, mismatching wait **107.233→78.329**
and zero-timeout wait **108.925→80.037 ns/source iteration**. Their gains are
approximately 20%, 21%, 27% and 27%. Go remains approximately flat on these
paths. Native loop expansion does not repeat its earlier advantage in this
campaign and is held; it is removed from the proposed production change.

## Scope and provenance

Base `ffd6300e596735d2c51181897ae92499fcde347b`, measured combination
`1c9440cec3f9b7f1c763b1d62a521a73d848b7c8`: cached LLGo cancellation,
direct numeric C ABI and candidate native word-loop expansion. The latter
changes native code generation only; all interpreter observations measure the
selected atomic path without the held unsafe-frame/local-counter experiments.
The subsequent production candidate removes native expansion and retains the
atomic changes. This record does not relabel the combination's native timings
as measurements of that later source.

Apple M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` `-O3 -lto=full`,
Wasmtime 48.0.1/Wasmer 7.3.0 Cranelift, wazero 1.12.0 compiler/interpreter and
WasmEdge 0.17.1 interpreter. Each case has three fresh serially interleaved
blocks, two warmups, seven warm calls and baseline Go/LLGo interpreter A/A
aliases. Loops use 1,000,000 iterations; four GC and three notify/wait cases use
10,000, and call-control performs one call. Every configuration passes original
and alternate independent oracles. The three unsupported GC reference modes
remain twelve explicit rejections. Six applicable C scalar controls run the
same counts and inputs, separately classified as unchecked native algorithms.

All source, executable, SDK-library and fixture hashes are checked. Readiness,
load, setup, preparation, instantiation, lookup, first call, warm and cleanup
remain separate; no startup subtraction is used. Host load changes about
12.0→7.4 on 16 CPUs, with five/fifteen-minute load 13.3/15.0→16.3/17.2. The
machine is not controlled. A partial initial campaign overlapped the extended
local race run and was interrupted; its samples are retained and excluded from
this complete campaign. The race run itself passed before its attempted stop.
The first build also rejected a missing verified LLGo-revision environment value;
the correctly pinned build then passed. No completed campaign is discarded.

## Complete interpreter comparison

Units are ns/source iteration, except call-control (ns/call).

| Workload | Go base | Head | LLGo base | Head | WasmEdge interp | wazero interp |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 5500.000 | 6667.000 | 6291.000 | 7958.000 | 6000.000 | 2417.000 |
| core/function-calls | 139.963 | 133.884 | 129.512 | 134.358 | 86.000 | 133.703 |
| core/integer-loop | 105.067 | 104.596 | 87.250 | 85.007 | 68.291 | 112.453 |
| core/memory-load-store | 199.685 | 201.430 | 166.774 | 168.941 | 128.513 | 199.106 |
| gc/i31-roundtrip | 73.438 | 74.808 | 63.304 | 61.629 | 46.700 | unsupported |
| gc/struct-access | 91.758 | 91.996 | 76.300 | 76.842 | 55.600 | unsupported |
| gc/struct-allocation | 149.363 | 147.383 | 171.304 | 169.783 | 108.300 | unsupported |
| gc/struct-indirect-access | 149.842 | 148.967 | 126.629 | 127.858 | 102.000 | unsupported |
| memory/dependent-address | 88.723 | 90.043 | 76.082 | 76.117 | 55.877 | 97.238 |
| memory/float64-unaligned | 137.064 | 137.300 | 115.778 | 117.105 | 84.540 | 135.614 |
| memory/narrow-unaligned | 182.236 | 182.241 | 154.040 | 154.979 | 115.782 | 187.078 |
| runtime/simd-memory-full | 145.581 | 143.028 | 129.020 | 123.674 | 77.731 | 151.371 |
| runtime/simd-memory-lanes | 179.318 | 178.304 | 162.726 | 160.051 | 93.254 | 194.112 |
| runtime/simd-memory-widen | 139.693 | 139.942 | 133.105 | 128.780 | 63.708 | 147.975 |
| threads/atomic-rmw | 75.837 | 75.881 | 86.141 | 67.719 | 31.784 | 48.463 |
| threads/notify-empty | 55.508 | 54.917 | 56.925 | 57.304 | 40.200 | 41.700 |
| threads/private-memory | 69.395 | 68.445 | 58.287 | 58.634 | 39.673 | 63.725 |
| threads/shared-memory | 101.516 | 101.712 | 73.589 | 58.858 | 39.341 | 62.458 |
| threads/wait-mismatch | 90.371 | 89.508 | 107.233 | 78.329 | 43.800 | 68.325 |
| threads/wait-timeout | 92.450 | 90.650 | 108.925 | 80.037 | 96.900 | 292.308 |

LLGo RMW becomes about 11% faster than Go but remains 2.13x WasmEdge.
Zero-timeout wait beats WasmEdge in this sample; mismatching wait still trails
it. Unchanged private-memory/float/scalar cases include small regressions.
Short-call controls vary substantially, and the integer baseline A/A alias
changes 87.250→84.473; these controls do not establish calibrated parity.
GC remains uneven, including an LLGo allocation deficit versus Go. All unfavorable
rows remain visible. No whole-interpreter win is claimed.

## Complete native comparison of the measured combination

This table retains the candidate expansion, which is subsequently held. Values
are ns/source iteration, except call-control. C lacks VM safety/setup contracts.

| Workload | Go base | Head | LLGo base | Head | Wasmtime | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 7875.000 | 9667.000 | 8917.000 | 12458.000 | 3000.000 | 8000.000 | — |
| core/function-calls | 1.083 | 1.091 | 1.080 | 1.140 | 1.241 | 1.107 | 1.061 |
| core/integer-loop | 2.115 | 2.119 | 2.135 | 2.098 | 2.157 | 2.131 | 2.104 |
| core/memory-load-store | 0.727 | 0.649 | 0.705 | 0.812 | 0.650 | 0.926 | 0.364 |
| gc/i31-roundtrip | 3.642 | 3.708 | 3.350 | 3.333 | 2.100 | — | — |
| gc/struct-access | 2.696 | 2.888 | 2.721 | 2.571 | 1.500 | — | — |
| gc/struct-allocation | 47.208 | 48.362 | 66.400 | 64.904 | 109.200 | — | — |
| gc/struct-indirect-access | 6.200 | 6.829 | 6.129 | 5.946 | 2.900 | — | — |
| memory/dependent-address | 1.917 | 1.922 | 1.933 | 1.921 | 2.141 | 2.169 | 1.465 |
| memory/float64-unaligned | 0.736 | 0.742 | 0.728 | 0.726 | 0.718 | 0.721 | 0.664 |
| memory/narrow-unaligned | 0.713 | 0.717 | 0.721 | 0.716 | 0.703 | 0.747 | 0.461 |
| runtime/simd-memory-full | 3.378 | 3.403 | 3.393 | 3.363 | 3.628 | 3.492 | — |
| runtime/simd-memory-lanes | 5.069 | 5.141 | 5.144 | 5.088 | 5.949 | 5.950 | — |
| runtime/simd-memory-widen | 1.137 | 1.121 | 1.085 | 1.079 | 1.089 | 1.109 | — |
| threads/atomic-rmw | 1.835 | 1.842 | 1.807 | 1.814 | 1.808 | 3.870 | — |
| threads/notify-empty | 2.033 | 2.283 | 1.867 | 2.071 | 7.000 | 5.300 | — |
| threads/private-memory | 0.351 | 0.365 | 0.332 | 0.337 | 0.891 | 0.891 | — |
| threads/shared-memory | 0.896 | 0.878 | 0.847 | 0.867 | 0.871 | 0.723 | — |
| threads/wait-mismatch | 2.492 | 2.429 | 2.146 | 2.050 | 9.100 | 3.100 | — |
| threads/wait-timeout | 2.837 | 3.054 | 2.233 | 2.275 | 43.200 | 172.100 | — |

LLGo native scalar memory regresses 0.705→0.812 here, versus Wasmtime 0.650 and
C 0.364. This reverses the [earlier word-loop result](performance-native-word-loops-2026-11-02.md).
Short sub-millisecond native intervals, changing shared-host conditions and A/A
controls for the interpreter rather than native engine prevent a stable native
win claim. The conservative choice is to retain the existing production native
loops. A future expansion proposal needs a longer matched native/A/A campaign.
The [long atomic/C probe](performance-native-atomic-c-2026-11-02.md) already
uses a different 100-million-iteration RMW count and must remain a separate table.

At this measured combination, all 31 Go packages, all 30 cross-package coverage
gates, affected pinned LLGo suites, full affected-package race, CI-scoped race,
no-cgo and vet pass. These local semantic passes do not establish final-head
cross-platform performance acceptance. The production PR still requires its
own final semantic matrix and baseline/competitor benchmark checks.

[Lossless evidence](evidence/performance-atomic-combined-2026-11-02.json.gz)
retains every sample and phase, all 294 summary rows, twelve explicit unsupported
outcomes, alternate checks, A/A aliases, source/build/SDK identities, measurement
sources, validation logs and the interrupted diagnostic campaign.
