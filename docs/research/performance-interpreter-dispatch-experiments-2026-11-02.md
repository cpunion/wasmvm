# Interpreter dispatch experiments: 2026-11-02

Two complete twenty-case campaigns improve most loops, but both regress LLGo
private memory by about 17%. Neither measured variant is adopted. The repeated
regression and unchanged-binary controls make it a real investigation, not a
result to remove from an aggregate score.

Base is `72d74d170068e59adbdcf8c6328c0a483c6d0f24`. Variant 1,
`9c1baf8`, moves stack-switching and wide arithmetic into the existing opcode
switch and GC into its extended-opcode default. Variant 2, `377ff87`, also
loads the third operand only in select/extended paths. Both preserve all
implementations, operands, source/fuel boundaries, polling, traps and suspension
PCs. No unsafe frame access or new execution framework is introduced.
The [CPU profiles](performance-interpreter-profile-2026-11-02.md) motivate this
classification change; they do not prove the cause of the private-loop regression.

Apple M4 Max/macOS arm64, Go 1.27.3, LLGo `8e45ed5` `-O3 -lto=full`,
Wasmtime 48.0.1/Wasmer 7.3.0 Cranelift, wazero 1.12.0 compiler/interpreter,
WasmEdge 0.17.1 interpreter. Each campaign completes the same twenty cases,
three fresh serially interleaved process blocks, two warmups and seven warm
calls. Twelve loops use one million iterations; four GC and three notify/wait
cases retain ten thousand; call-control performs one call. Variant 2 measures
the known private-memory regression first; this ordering is recorded before
measurement. Both include exact baseline Go/LLGo interpreter A/A aliases and
six applicable C controls. Independent and alternate-input checks pass;
twelve unsupported GC/reference outcomes remain explicit. Source, binary, SDK
and fixture identities are checked. Every startup/preparation/first/warm/cleanup
phase and every unfavorable sample remains retained.

## Complete LLGo interpreter results

Units are ns/source iteration, except call-control (ns/call). Each base/head pair
belongs to its own campaign; absolute timings are not pooled across campaigns.

| Workload | Variant 1 base | Head | Variant 2 base | Head |
| --- | ---: | ---: | ---: | ---: |
| core/call-control | 3375.000 | 3833.000 | 2334.000 | 3167.000 |
| core/function-calls | 93.999 | 87.442 | 93.030 | 86.026 |
| core/integer-loop | 62.051 | 55.009 | 61.949 | 55.420 |
| core/memory-load-store | 139.781 | 109.419 | 119.895 | 105.954 |
| gc/i31-roundtrip | 63.421 | 59.933 | 60.450 | 57.175 |
| gc/struct-access | 79.088 | 71.862 | 74.533 | 70.175 |
| gc/struct-allocation | 177.367 | 181.346 | 170.100 | 166.708 |
| gc/struct-indirect-access | 123.954 | 116.079 | 126.917 | 120.025 |
| memory/dependent-address | 73.682 | 64.172 | 73.374 | 63.908 |
| memory/float64-unaligned | 124.709 | 110.117 | 117.599 | 102.813 |
| memory/narrow-unaligned | 178.344 | 144.102 | 155.362 | 140.350 |
| runtime/simd-memory-full | 123.210 | 112.217 | 135.103 | 124.762 |
| runtime/simd-memory-lanes | 155.729 | 142.540 | 155.240 | 142.773 |
| runtime/simd-memory-widen | 124.899 | 111.166 | 125.490 | 113.115 |
| threads/atomic-rmw | 67.837 | 61.950 | 68.070 | 61.947 |
| threads/notify-empty | 56.962 | 52.737 | 57.775 | 51.571 |
| threads/private-memory | 56.935 | 66.811 | 56.900 | 66.696 |
| threads/shared-memory | 58.039 | 51.817 | 57.893 | 51.230 |
| threads/wait-mismatch | 78.279 | 71.646 | 78.258 | 71.283 |
| threads/wait-timeout | 77.362 | 71.062 | 79.192 | 71.975 |

The private-memory baseline A/A stays near its original baseline in both runs,
while the candidate regression repeats. Most integer, memory, shared, SIMD and
GC-access loops improve; GC allocation remains uneven. The complete evidence
retains Go results (including regressions), native results and every named
reference, not only the LLGo table above. Empty-call controls are particularly
noisy and do not calibrate a boundary subtraction or an acceptance threshold.

Both candidates pass all 31 Go semantic packages, all 30 cross-package coverage
gates, affected pinned LLGo execute/VM suites, scoped race, no-cgo and vet.
The initial scoped Go command for variant 1 accidentally ran at the baseline
checkout; its result is labeled base-go and does not serve as candidate proof.
The subsequent full candidate Go/coverage run is the authoritative validation.
No missing candidate check is replaced by that baseline run.

A [third measured structure](performance-interpreter-extension-guard-2026-11-02.md) retains the original scalar switch and
places the existing extension classifiers behind one guard. It compares
against the latest integrated main, which includes the separately validated
native-memory transformation. Earlier native rows are not relabeled as timings
of this new combination. The first two experiments remain held regardless of
what the later candidate does. Full controlled acceptance remains open.

[Lossless evidence](evidence/performance-interpreter-dispatch-experiments-2026-11-02.json.gz)
retains both complete matrices, A/A aliases, all phases, raw calls, source/build
identities, exact patches, measurement programs and validation logs.
