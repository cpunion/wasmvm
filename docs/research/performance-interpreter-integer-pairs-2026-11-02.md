# Integer-pair dispatch and exact constant widths: 2026-11-02

This candidate caches a following pure I32 binary-operation hint in the existing
32-byte instruction. Unmetered execution commits the constant, advances to the
consumer and enters the same operand/operation implementation once. Both original
instructions and PCs remain present. Fuel-enabled execution and a pair crossing
the next 64-step polling boundary follow the ordinary individual steps. Native
execution never takes the interpreter shortcut. No second execution framework or
new executable-bytecode copy is introduced.

The compiler scans once within its charged node budget. Eligible consumers are
nontrapping I32 comparisons, add/subtract/multiply, bitwise operations and shifts/
rotates. The consumer must read the constant slot as B and a distinct valid A;
trapping division, unary operations and other opcode classes are not hinted.
The data write occurs before operand reads, preserving two-address slot aliases.

## Correctness and representation

The hint reuses I32 Count, while Extra retains the complete 32-bit literal.
I64/F32/F64 cached representations remain unchanged. I32 reconstruction explicitly
zero-extends its bits. This also corrects an owned API bug: the prior interpreter
exposed unwanted sign-extended upper bits for negative I32 constants while native
returned the proper 32-bit bits. The public regression fails on the old baseline
in both fuel modes, and passes after correction under Go/LLGo, interpreter/native.
I64 full-word and F32/F64 NaN-payload controls preserve exact bits.

The I32-width fix is included separately in PR #98. The second comparison below
uses that already corrected source as its baseline, isolating pair dispatch from
the correctness fix. Earlier combined observations remain retained, rather than
being relabeled as measurements of the new baseline.

## Protocol and exact sources

First campaign: base `e76f172a371482b721812367005a5d31309c8d9a` versus measured
candidate `4d91cce174d3341fad34e1111cfa19a2ff0722b6`. It includes both constant
width correction and pair dispatch. LLGo integer improves 59.190→49.768 and
memory 112.160→100.183 ns/source iteration; WasmEdge is 48.511/90.465.
GC allocation regresses 163.333→169.425 and RMW 65.509→67.453. All rows and
phases remain in the evidence. An earlier goto-based prototype was built and
semantically checked but not timed; its logs are not used as performance proof.

Second campaign: corrected base `94c2c168b37a339479355ae6122057f47eb9ee7a`
versus the same measured `4d91cce` executable. Later contribution commits add
regression tests, merge the same correction and add documentation; runtime and
compiler implementation match the measured candidate.

Apple M4 Max/macOS arm64, Go 1.27.3, pinned LLGo `8e45ed5` `-O3 -lto=full`,
Wasmtime 48.0.1/Wasmer 7.3.0 Cranelift, wazero 1.12.0 compiler/interpreter and
WasmEdge 0.17.1 interpreter. Each twenty-case campaign uses three fresh serially
interleaved process blocks, two warmups and seven warm calls. Twelve loops use
one million iterations; four GC and three notify/wait cases retain ten thousand;
the empty control performs one call. Exact baseline Go/LLGo interpreter A/A
aliases and six applicable C controls run in the same blocks. All original and
alternate-input checks pass, hashes are verified, and twelve GC/reference
rejections remain explicit unsupported outcomes. Startup, setup, preparation,
instantiation, lookup, first/warm calls and cleanup stay separate. No startup
subtraction, sample removal or controlled parity verdict is used.

## Complete isolated interpreter results

Units: ns/source iteration except call-control (ns/call).

| Workload | Go base | Head | LLGo base | Head | WasmEdge interp | wazero interp |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 2833.000 | 3000.000 | 4333.000 | 3417.000 | 2000.000 | 833.000 |
| core/function-calls | 85.069 | 81.554 | 91.448 | 82.646 | 62.190 | 79.316 |
| core/integer-loop | 63.474 | 54.531 | 60.280 | 49.789 | 48.654 | 81.489 |
| core/memory-load-store | 123.089 | 112.380 | 130.860 | 100.099 | 92.192 | 143.545 |
| gc/i31-roundtrip | 64.146 | 59.879 | 58.237 | 54.208 | 45.400 | unsupported |
| gc/struct-access | 79.017 | 72.654 | 68.967 | 64.912 | 55.200 | unsupported |
| gc/struct-allocation | 139.812 | 136.300 | 176.321 | 166.904 | 104.300 | unsupported |
| gc/struct-indirect-access | 126.708 | 115.717 | 121.529 | 106.500 | 96.300 | unsupported |
| memory/dependent-address | 72.939 | 65.672 | 68.867 | 61.448 | 53.384 | 95.557 |
| memory/float64-unaligned | 117.622 | 109.543 | 126.528 | 96.983 | 84.097 | 134.975 |
| memory/narrow-unaligned | 154.328 | 138.699 | 158.383 | 124.713 | 113.149 | 183.795 |
| runtime/simd-memory-full | 132.561 | 129.236 | 137.231 | 119.085 | 77.368 | 148.580 |
| runtime/simd-memory-lanes | 162.796 | 161.630 | 182.805 | 152.695 | 90.062 | 187.043 |
| runtime/simd-memory-widen | 125.021 | 119.525 | 122.645 | 120.515 | 60.896 | 143.192 |
| threads/atomic-rmw | 69.145 | 68.910 | 66.423 | 66.546 | 31.718 | 43.311 |
| threads/notify-empty | 47.583 | 48.500 | 54.763 | 57.167 | 40.600 | 41.367 |
| threads/private-memory | 58.575 | 54.379 | 56.366 | 52.238 | 38.367 | 62.386 |
| threads/shared-memory | 92.767 | 89.652 | 56.080 | 51.781 | 38.448 | 60.918 |
| threads/wait-mismatch | 80.025 | 82.125 | 74.013 | 75.987 | 43.200 | 68.763 |
| threads/wait-timeout | 80.108 | 82.575 | 74.729 | 74.629 | 97.000 | 292.371 |

## Interpretation and remaining gaps

The isolated LLGo integer loop improves 60.280→49.789 ns/iteration (17.4%),
the memory loop 130.860→100.099 (23.5%), and calls 91.448→82.646 (9.6%).
Their baseline A/A aliases are 60.095, 130.902 and 91.186. These observations
support advancing this compact candidate; they do not close controlled acceptance.

Against the named references, LLGo integer remains 2.3% behind WasmEdge,
memory 8.6% and calls 32.9%. SIMD widening is 120.515 versus 60.896 (1.98x),
and RMW 66.546 versus 31.718 (2.10x). Notify is 57.167 versus 40.600.
The candidate improves 15 of the 19 nonempty LLGo cases by at least 1%;
RMW is essentially flat (+0.2%), notify regresses 4.4%, wait-mismatch 2.7%,
and timeout is essentially flat (-0.1%). The earlier campaign's GC-allocation
regression is not discarded: its corrected-base repeat instead improves 5.3%.
Neither observation alone proves a controlled result. Short call-control rows
are boundary diagnostics and cannot establish guest-throughput parity.

LLGo also remains slower than Go on GC allocation (166.904 versus 136.300,
22.5%) and notify (57.167 versus 48.500, 17.9%); it leads Go in several scalar
and memory cases. These are workload observations, not new LLGo defect reports.
Independently reproduced compiler/runtime problems remain tracked separately
in the gap ledger. No issue is inferred from relative throughput alone.

The native memory kernel is 0.474 versus C 0.267 (1.78x), despite being close
to Wasmtime 0.478. Native integer/calls are 1.557/0.784 versus C 1.544/0.778.
GC struct-access/indirect-access remain 2.312/5.608 versus Wasmtime 1.400/3.000.
These separate deficits prioritize memory lowering, GC runtime paths and SIMD
interpreter work; integer-loop proximity cannot stand in for their acceptance.

## Complete native and C results

The pair is an interpreter-only execution shortcut; host boundary work still
changes. Native observations are therefore retained, including regressions.
C is a separately labeled algorithm reference with fewer execution contracts;
absent VM phases are not invented zeroes. Units are as above.

| Workload | Go base | Head | LLGo base | Head | Wasmtime | Wasmer | C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 5083.000 | 4875.000 | 4834.000 | 5417.000 | 1000.000 | 2000.000 | — |
| core/function-calls | 0.802 | 0.805 | 0.800 | 0.784 | 0.907 | 0.784 | 0.778 |
| core/integer-loop | 1.572 | 1.584 | 1.562 | 1.557 | 1.548 | 1.566 | 1.544 |
| core/memory-load-store | 0.496 | 0.494 | 0.475 | 0.474 | 0.478 | 0.536 | 0.267 |
| gc/i31-roundtrip | 3.621 | 3.750 | 3.304 | 3.450 | 2.100 | — | — |
| gc/struct-access | 2.229 | 2.221 | 2.375 | 2.312 | 1.400 | — | — |
| gc/struct-allocation | 49.729 | 49.258 | 65.283 | 62.663 | 105.400 | — | — |
| gc/struct-indirect-access | 6.475 | 6.138 | 5.483 | 5.608 | 3.000 | — | — |
| memory/dependent-address | 1.860 | 1.855 | 1.839 | 1.836 | 2.064 | 2.088 | 1.413 |
| memory/float64-unaligned | 0.734 | 0.739 | 0.716 | 0.723 | 0.710 | 0.708 | 0.662 |
| memory/narrow-unaligned | 0.718 | 0.706 | 0.690 | 0.686 | 0.693 | 0.717 | 0.451 |
| runtime/simd-memory-full | 3.356 | 3.391 | 3.346 | 3.342 | 3.604 | 3.441 | — |
| runtime/simd-memory-lanes | 5.055 | 5.058 | 5.016 | 5.035 | 5.795 | 5.829 | — |
| runtime/simd-memory-widen | 1.062 | 1.062 | 1.040 | 1.035 | 1.029 | 1.048 | — |
| threads/atomic-rmw | 1.841 | 1.834 | 1.822 | 1.814 | 1.805 | 3.878 | — |
| threads/notify-empty | 2.075 | 1.979 | 1.812 | 1.933 | 6.900 | 5.000 | — |
| threads/private-memory | 0.340 | 0.340 | 0.326 | 0.325 | 0.840 | 1.024 | — |
| threads/shared-memory | 0.855 | 0.860 | 0.846 | 0.840 | 1.016 | 1.019 | — |
| threads/wait-mismatch | 2.038 | 1.962 | 1.896 | 2.025 | 9.100 | 2.900 | — |
| threads/wait-timeout | 2.500 | 3.225 | 2.117 | 2.004 | 42.500 | 169.300 | — |

## Native code identity and validation

An isolated compile-only diagnostic captures 86 body pairs across twenty fixtures,
amd64/arm64, both policy tiers and optional expansion. Every pair is byte-identical
between `e76f172` and `4d91cce`. Source hashes, feature masks and capture code are
retained. The foreign-architecture code is never published or executed. Cross-
amd64 uses its baseline feature mask; ARM64 retains the actual host mask. This
is a declared-profile code identity proof, not an emulated performance result or
a claim about all CPU masks. The subsequent corrected baseline changes only
interpreter reconstruction and public tests, not compiler inputs or native
lowering. Host-runtime boundary timing is not assumed identical from this proof.

All 31 Go semantic packages, all 30 coverage gates, pinned LLGo slots/execute/VM
suites, scoped race, no-cgo and vet pass at the measured candidate. Focused tests
cover wrapping and signed arithmetic, masked shifts/rotates, reuse, exact source
fuel, cancellation before a consumer at step 64 and unchanged visible state on
failure. Original cold PC entry remains valid because no instruction is deleted.
The public constant-width test additionally passes Go/LLGo on both backends and
both fuel policies. Final-head dual-platform CI and benchmark validation remain
required before integration. Controlled whole-portfolio acceptance remains open.

[Lossless evidence](evidence/performance-interpreter-integer-pairs-2026-11-02.json.gz)
retains both full campaigns, all 294 summaries per campaign, A/A controls, all
phases, explicit unsupported outcomes, source/build/toolchain/SDK identities,
validation and expected baseline-failure logs, native code identities and all
measurement/diagnostic programs.

Observed shared-host load: first campaign [9.0087890625, 39.5263671875, 37.611328125]→[7.98486328125, 17.75341796875, 27.63818359375]; isolated campaign [5.84814453125, 5.85595703125, 8.98681640625]→[6.69384765625, 6.5576171875, 8.328125]. No CPU affinity or host isolation is claimed.
