# Coalesced local-read prefixes: 2026-11-02

Status: candidate held for further native-regression and cross-platform review,
not integrated or performance-accepted. Base `ed5a0fb837b7d54da61c2847f6aae30c777d1e36`,
engine `d8a35692817833bb77fc2a52b6e4200fb23d7b03`.

The previous local-copy experiment skipped only data movement and paid ordinary
instruction dispatch. This candidate consumes validated self-copy local reads
before entering the same ordinary dispatcher. It extends the existing constant
prefix mechanism with one bounded metadata scan and a small prefix loop. There
is no new opcode, unsafe access or duplicate arithmetic interpreter. Both source
PCs and all source instructions remain; native and precise-fuel paths stay ordinary.

A local read is hinted only when its result already occupies that local's slot.
Reads requiring a snapshot before a later local mutation remain normal copies.
Every consumed prefix increments the original step counter, commits any constant
before its consumer, and stops before the next 64-step poll, including uint64
wrap. Every value bit remains intact, including V128 and FP NaN payloads.

All 31 Go packages, 30 coverage gates, pinned LLGo slots/execute/VM, scoped race,
no-cgo and vet pass. New tests cover hint/source/word preservation, non-coalesced
copies, long prefixes, exact fuel, reuse, 63/64 boundaries, counter wrap and a
following trap after visible effects. The first test build used the nonexistent
MemoryType fixture type; it was corrected to the actual Limits schema before
committing. This was a test error, not an engine or LLGo defect.

A compile-only capture of 86 native bodies over amd64/arm64 and both tiers
finds every body byte-identical. Foreign bodies are never published or timed.
This does not establish identical host-boundary or collection performance.

## Complete local comparison

Apple M4 Max/macOS arm64, Go 1.27.3, pinned LLGo 8e45ed5 O3/full LTO, the five
pinned reference configurations and six applicable ordinary C controls. All
20 unchanged kernels complete: three fresh serially interleaved process blocks,
two warmups and seven calls, exact Go/LLGo interpreter baseline A/A aliases,
independent original/alternate oracles, twelve explicit GC reference rejections.
Twelve loops use one million iterations, four GC and three notify/wait cases
retain ten thousand, and call-control performs one call. All 294 summaries and
every preparation/first/warm/readiness/cleanup phase remain retained. Load is
8.37/8.69/8.52 before and 6.97/7.68/8.09 after, without host isolation or affinity.

Units are ns/source iteration, except the single-call control (ns/call):

| Workload | Go base → head | LLGo base → head | LLGo baseline A/A | WasmEdge interpreter | wazero interpreter |
| --- | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 2666.000 → 2083.000 | 3375.000 → 3500.000 | 3000.000 | 2000.000 | 875.000 |
| core/function-calls | 80.718 → 67.528 | 83.314 → 72.856 | 83.676 | 62.108 | 82.865 |
| core/integer-loop | 54.254 → 43.885 | 50.345 → 42.103 | 50.239 | 48.703 | 81.609 |
| core/memory-load-store | 111.362 → 85.092 | 100.217 → 81.154 | 100.150 | 90.773 | 145.853 |
| gc/i31-roundtrip | 57.975 → 52.196 | 53.671 → 50.979 | 53.462 | 44.800 | unsupported |
| gc/struct-access | 72.754 → 57.667 | 67.554 → 55.758 | 66.946 | 56.400 | unsupported |
| gc/struct-allocation | 135.575 → 116.758 | 163.621 → 157.429 | 160.450 | 106.200 | unsupported |
| gc/struct-indirect-access | 118.646 → 100.412 | 107.208 → 94.796 | 108.821 | 97.300 | unsupported |
| memory/dependent-address | 65.308 → 56.212 | 61.456 → 54.677 | 61.430 | 53.095 | 94.701 |
| memory/float64-unaligned | 108.722 → 90.749 | 96.938 → 82.465 | 96.629 | 84.409 | 138.370 |
| memory/narrow-unaligned | 137.179 → 109.066 | 124.343 → 103.799 | 124.546 | 113.196 | 184.742 |
| runtime/simd-memory-full | 127.363 → 104.892 | 118.447 → 99.359 | 117.339 | 77.386 | 149.427 |
| runtime/simd-memory-lanes | 160.368 → 128.009 | 181.426 → 157.067 | 180.298 | 89.935 | 184.639 |
| runtime/simd-memory-widen | 100.883 → 86.490 | 99.973 → 90.036 | 99.718 | 61.287 | 145.397 |
| threads/atomic-rmw | 68.808 → 65.928 | 65.674 → 62.391 | 66.701 | 31.668 | 48.079 |
| threads/notify-empty | 49.429 → 43.408 | 56.288 → 51.783 | 55.096 | 40.400 | 39.842 |
| threads/private-memory | 54.493 → 50.880 | 52.186 → 49.045 | 51.747 | 38.420 | 61.616 |
| threads/shared-memory | 87.848 → 88.692 | 51.643 → 48.171 | 51.249 | 38.376 | 62.747 |
| threads/wait-mismatch | 82.108 → 74.008 | 75.658 → 72.292 | 76.921 | 43.400 | 71.433 |
| threads/wait-timeout | 82.375 → 73.458 | 74.513 → 68.629 | 74.850 | 96.100 | 288.225 |

All nineteen nonempty LLGo interpreter cases improve in this descriptive
comparison. Integer improves 16.4%, calls 12.6%, memory 19.0%; Go improves
19.1%/16.3%/23.6%. LLGo integer/memory now lead this WasmEdge configuration,
while calls remain 17.3% slower. SIMD widening remains 1.47x WasmEdge, lanes
1.75x, atomic RMW 1.97x. Go shared memory regresses 1.0%; short LLGo call-control
regresses 3.7%, with substantial A/A control variation. No aggregate parity or
controlled acceptance is inferred.

## Native regressions and separate follow-up

The broad campaign retains native regressions: Go wait-mismatch 25.6%,
wait-timeout 17.1%; LLGo GC allocation 61.342→72.654 ns/iteration (+18.4%),
wait-timeout 8.7% and notify 6.3%. Native body identity does not erase them.
Native LLGo memory remains 0.47175 versus ordinary C 0.27100 (1.74x), despite
Wasmtime 0.48200. This prefix change does not optimize native generated loops.

A focused follow-up reuses verified exact-source binaries for only the original
10,000-allocation chain, adding exact native baseline A/A aliases for both host
compilers. Three fresh blocks, two warmups, seven interleaved calls and all
phases/checks remain. Wasmtime is the supported native GC reference; previously
recorded missing GC configurations and ordinary C are not retimed or invented.

| Native GC configuration | ns/iteration |
| --- | ---: |
| Go base | 53.2958 |
| Go head | 57.7000 |
| Go baseline A/A | 56.6042 |
| LLGo base | 73.9667 |
| LLGo head | 76.3541 |
| LLGo baseline A/A | 75.2792 |
| Wasmtime | 125.0000 |

The earlier 18.4% LLGo difference is not repeated at that magnitude, but the
follow-up still records head/base +3.2% and A/A +1.8%; paired LLGo head/base
ranges 0.9983–1.0605, A/A 0.9996–1.0667. Go regresses 8.3% while A/A also
regresses 6.2%. These shared-host observations do not resolve controlled native
regression acceptance. Continue investigating host allocation/code-layout effects
and preserve the held candidate; do not file an LLGo bug from these ratios alone.
Startup/readiness remains separate and also fails to establish the declared
controlled latency target. Native allocation, SIMD/atomic gaps and final-head
Linux/ARM64 checks remain open.

[Lossless evidence](evidence/performance-prefix-chains-2026-11-02.json.gz)
retains both complete campaigns, all unfavorable observations, A/A controls,
source/binary/toolchain/SDK identities, capture/measurement programs, and tests.
