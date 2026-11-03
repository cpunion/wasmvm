# Direct vector-constant experiment: 2026-11-03

Status: **held**. The five-line candidate routes v128.const through the ordinary
interpreter constant cases instead of the generic SIMD evaluator. It changes
neither polling/fuel nor native code-generation sources. The first four-case
study improves LLGo full/lane SIMD about 7.0%/5.6%, but the complete later
study retains a **25.1% native LLGo widening regression**. Passing semantics
and favorable interpreter rows do not waive that control. No implementation
from this experiment is selected for main or a contribution PR.

Base `b34e74bcce8e0ae607ca31a89977fc1d217ea9e7`, local candidate
`ac51429fa3ac43282d48c6bd6fc494ee549e113a`, fixed Go 1.27.3 and LLGo `983465e`
`-O3 -lto=full`, M4 Max/macOS arm64. Both campaigns use the same exact source
binaries and all five pinned reference modes, with exact interpreter baseline
A/A aliases. The first campaign is retained separately from the subsequent
twenty-case check; neither replaces the other.

Each uses three fresh serially interleaved process blocks, two warmups/seven
warm samples, independent/alternate checks and separate preparation/first/warm/
cleanup/readiness. The first four cases use one million iterations: 180 VM/
three C observations, sixty alternate checks and 61 summaries. The broader
scope keeps the prior twenty-kernel counts (one million normal loops, 10,000
GC/notify/wait and one-call control): 864 VM/eighteen C observations, 288
alternate checks, twelve explicit GC-reference rejections and 294 summaries.
Independent source/profile/fixture/result and summary regeneration checks pass.
These shared-host diagnostics do not establish controlled acceptance.

## Four-case triage

Warm ns/source iteration:

| Workload | Go interpreter base → head | LLGo interpreter base → head | WasmEdge | wazero |
| --- | ---: | ---: | ---: | ---: |
| core/memory-load-store | 116.028 → 116.384 | 102.149 → 98.204 | 124.909 | 196.390 |
| runtime/simd-memory-full | 146.265 → 137.308 | 130.497 → 121.411 | 106.153 | 206.957 |
| runtime/simd-memory-widen | 115.557 → 114.067 | 114.062 → 111.954 | 83.214 | 195.673 |
| runtime/simd-memory-lanes | 174.240 → 168.366 | 169.892 → 160.324 | 123.354 | 252.461 |

## Complete broader outcomes

Same units, except empty in ns/call. Missing GC reference rows are explicit
unsupported outcomes. C rows remain ordinary unchecked scalar kernels.

| Workload | Go interp base → head | LLGo interp base → head | LLGo native base → head | WasmEdge interp | Ordinary C |
| --- | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 66.887 → 68.949 | 59.443 → 55.184 | 0.432 → 0.433 | 52.434 | — |
| core/call-control | 3417.000 → 3333.000 | 3583.000 → 3417.000 | 4125.000 → 4000.000 | 3000.000 | — |
| core/integer-loop | 57.115 → 58.415 | 48.691 → 45.635 | 2.114 → 2.117 | 66.650 | 2.103 |
| core/function-calls | 86.151 → 87.070 | 81.531 → 81.239 | 1.064 → 1.067 | 85.110 | 1.050 |
| core/memory-load-store | 115.949 → 117.469 | 102.124 → 98.275 | 0.590 → 0.596 | 125.359 | 0.364 |
| memory/narrow-unaligned | 152.260 → 150.315 | 135.397 → 129.348 | 0.933 → 0.935 | 155.267 | 0.619 |
| memory/dependent-address | 75.561 → 75.688 | 66.329 → 63.057 | 2.496 → 2.498 | 73.133 | 1.920 |
| memory/float64-unaligned | 119.351 → 120.035 | 106.281 → 101.014 | 0.986 → 0.983 | 115.234 | 0.901 |
| runtime/simd-memory-full | 145.311 → 136.162 | 130.253 → 120.980 | 4.568 → 4.570 | 105.946 | — |
| runtime/simd-memory-widen | 131.622 → 128.077 | 135.017 → 117.934 | 1.545 → 1.934 | 100.733 | — |
| runtime/simd-memory-lanes | 179.268 → 169.291 | 169.857 → 160.333 | 6.907 → 6.880 | 123.255 | — |
| gc/i31-roundtrip | 70.571 → 68.417 | 58.125 → 59.121 | 2.383 → 2.454 | 61.300 | — |
| gc/struct-access | 75.604 → 77.287 | 66.267 → 65.237 | 2.733 → 2.654 | 74.900 | — |
| gc/struct-allocation | 159.708 → 162.363 | 195.137 → 191.979 | 70.617 → 71.729 | 142.700 | — |
| gc/struct-indirect-access | 134.387 → 139.787 | 124.388 → 120.883 | 7.162 → 7.171 | 130.600 | — |
| threads/shared-memory | 112.140 → 115.250 | 58.892 → 55.086 | 1.063 → 1.066 | 52.517 | — |
| threads/atomic-rmw | 86.653 → 84.793 | 59.466 → 57.881 | 2.468 → 2.474 | 43.454 | — |
| threads/notify-empty | 60.646 → 58.633 | 53.900 → 51.204 | 2.150 → 2.188 | 55.400 | — |
| threads/wait-mismatch | 100.550 → 99.217 | 72.571 → 66.967 | 2.267 → 2.154 | 59.600 | — |
| threads/wait-timeout | 132.625 → 128.833 | 90.633 → 89.625 | 3.100 → 2.821 | 182.400 | — |

Retain LLGo interpreter i31 +1.7%, native widening +25.1% and native i31 +3.0%.
Go interpreter private memory +3.1%, integer +2.3%, indirect GC +4.0% and
native wait-mismatch +36.2% also remain. Every other mode is in the evidence;
no fastest-per-case engine or adjusted baseline is constructed.

Native widening block medians illustrate the unresolved host variability:
LLGo base/head are 1.5455/1.9336, 2.2654/2.2587 and 1.4093/1.4169.
Wasmtime moves 1.726/2.230/1.411 across those blocks. Correlated variation
is an observation, not proof that the adverse own row is harmless. The shortest
native warm intervals are only around 1–2 ms. A separately declared long-native
comparison with matching ordinary-C SIMD oracles is needed; repeating this
unchanged short campaign to seek a more favorable result is not justified.

## Semantic and coverage evidence

Complete execute/VM suites pass under Go and fixed full-LTO LLGo, with 48
passing Core/proposal reports each. All five mandatory extensions execute in
three tiers, including forced Stack Switching collection. Full Go semantics
then pass all 31 packages, sixty reports and all thirty 90% coverage gates.
The first coverage command omitted CI's cross-package `-coverpkg` setting;
its insufficient profile and original results are preserved. The corrected
command uses the exact CI instrumentation scope, without changing any tests.
No capability or performance milestone is closed by these results.

The final source on main remains [merged PR111](pr111-final-ci-2026-11-03.md).
Its new SIMD CPU profiles motivate further dispatcher work; the separate
[experimental intrinsic-buffer finding](performance-llgo-simd-array-escape-2026-11-03.md)
is reported upstream and does not change the production compiler profile.

[Both complete campaigns, all original controls/phases, source patch, identities and validation](evidence/performance-vector-constants-held-2026-11-03.json.gz).
