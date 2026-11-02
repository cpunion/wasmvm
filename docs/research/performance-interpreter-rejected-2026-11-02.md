# Interpreter experiments retained without adoption: 2026-11-02

Neither experiment below is selected for production. Both pass semantic tests,
but neither demonstrates enough consistent LLGo benefit to justify its added
complexity. Keeping these negative results avoids treating an architectural
hypothesis as an established explanation of a competitor's lead.

Both use base `72d74d170068e59adbdcf8c6328c0a483c6d0f24`, Apple M4 Max/macOS
arm64, Go 1.27.3, pinned LLGo `8e45ed5` with `-O3 -lto=full`, Wasmtime 48.0.1,
Wasmer 7.3.0, wazero 1.12.0 and WasmEdge 0.17.1. They are separate descriptive
campaigns, not directly comparable absolute timings. Each completes 20 unchanged
cases, three fresh serially interleaved process blocks, two warmups and seven
warm calls. Twelve loops use 1,000,000 source iterations; four GC and three
notify/wait kernels retain 10,000; the empty control performs one call. Independent
and alternate-input oracles pass. Three reference configurations explicitly
reject each GC case (12 recorded unsupported results); rejection is not execution.
Six applicable scalar cases also interleave a separate C reference. All startup,
preparation, first-call, warm and cleanup observations remain separate.

## Checked frame admission and direct operand access

Candidate `4aa4ce76540fd10261921b4337a67a296b2c2ebf` validates every operand
against its frame budget during slot construction and checks admission before
frame allocation. It then replaces four operand slice bounds checks with typed
pointer access within that admitted frame. Guest linear-memory bounds, fuel,
limits, traps and cancellation remain enforced. Tests include corrupt operands,
sentinels, zero frames, subslice aliasing, race and `checkptr=2`.

Go improves about 2–4% on several loops; LLGo integer/memory are almost flat,
while calls, narrow memory and several GC cases regress. This does not establish
an LLGo bounds-check elimination defect. The added unsafe access is not adopted.

## Local dispatcher counter

Candidate `29da08cdf92e3965de602b5a6558fb88d8a3d001` increments a local counter
and publishes it before every existing polling/helper boundary and normal Run
return, without a defer or closure. Tests retain exact 63/64/65 boundaries,
uint64 wrap, fuel failure before effects, host suspension/resumption, failures
and Rebind. It does not remove a cancellation poll.

LLGo integer/memory regress about 1%; most other changes overlap same-binary
A/A variation. Go memory regresses about 7%. The independent A/A aliases, all
regressions and short-call noise are retained. This candidate is not adopted.

## Complete LLGo interpreter results

Values are ns/source iteration, except call-control (ns/call). Each pair belongs
to its own campaign; no subtraction or cross-campaign percentage is used.

| Workload | Frame base | Frame head | Counter base | Counter head |
| --- | ---: | ---: | ---: | ---: |
| core/call-control | 3334.000 | 3375.000 | 4916.000 | 5291.000 |
| core/function-calls | 93.023 | 93.948 | 107.797 | 107.423 |
| core/integer-loop | 62.388 | 62.372 | 67.128 | 67.818 |
| core/memory-load-store | 119.534 | 119.086 | 140.972 | 142.197 |
| gc/i31-roundtrip | 61.083 | 57.525 | 89.496 | 82.362 |
| gc/struct-access | 74.050 | 75.358 | 95.708 | 95.279 |
| gc/struct-allocation | 166.363 | 170.412 | 170.754 | 169.613 |
| gc/struct-indirect-access | 124.446 | 125.967 | 124.158 | 124.017 |
| memory/dependent-address | 72.710 | 72.847 | 73.899 | 73.700 |
| memory/float64-unaligned | 115.728 | 115.494 | 132.254 | 131.245 |
| memory/narrow-unaligned | 150.059 | 152.264 | 151.562 | 149.841 |
| runtime/simd-memory-full | 122.997 | 122.109 | 213.930 | 209.752 |
| runtime/simd-memory-lanes | 155.324 | 152.080 | 244.500 | 240.496 |
| runtime/simd-memory-widen | 124.358 | 121.599 | 211.665 | 212.241 |
| threads/atomic-rmw | 68.105 | 66.114 | 67.685 | 66.420 |
| threads/notify-empty | 57.062 | 55.921 | 57.504 | 55.612 |
| threads/private-memory | 56.739 | 56.684 | 57.187 | 56.943 |
| threads/shared-memory | 57.502 | 57.072 | 57.680 | 57.220 |
| threads/wait-mismatch | 78.213 | 77.850 | 76.525 | 77.992 |
| threads/wait-timeout | 77.862 | 77.629 | 77.263 | 79.013 |

At each candidate, all 31 Go semantic packages and all 30 cross-package coverage
gates pass, as do affected pinned LLGo suites, scoped race, no-cgo and vet.
The frame candidate additionally passes checkptr. An initial expected-step
assertion counted cold continuation instructions as executed instructions; the
corrected test uses the independent source NOP count. Initial fixture-selection
and missing-artifact setup failures are retained. The shared thread cancellation
fixture now compiles before starting its 100 ms execution deadline; its original
failure, repeated checks and correction are also retained.

[Lossless evidence](evidence/performance-interpreter-rejected-2026-11-02.json.gz)
contains both complete matrices, all Go/native/reference rows and phases,
A/A aliases, hashes, measurement sources, patches and validation logs. Shared-host
load and short controls prevent a controlled acceptance claim.
