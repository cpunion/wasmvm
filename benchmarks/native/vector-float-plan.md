# Native SIMD floating-point scope

Baseline: `6a5dc6b` (production code equal to `aa89f0d`). Implement all 52 Core
2.0 SIMD floating-point comparisons, arithmetic/rounding operations and numeric
conversions directly on amd64 and arm64. Keep the existing vector register,
spill, cold-continuation and exact-fuel contracts. Baseline amd64 remains SSE2;
no optional CPU instruction is silently enabled. No whole-function interpreter
or competitor fallback is allowed.

Before implementation, declare f32x4/f64x2 arithmetic, min/max with comparisons
and pseudo-min/max, rounding/square-root, signed/unsigned conversion workloads,
and a promote/demote workload. Every lane contributes to an independent Python
checksum. Check alternate seeds and iteration counts against pinned competitors.
Use the public native API, separate startup/compile/instantiate/first-call/warm
phases, and retain own Go/LLGo baseline/head plus wazero/Wasmtime/Wasmer. Include
integer, calls, scalar memory and prior SIMD regression families.

Short 10,000-iteration runs with three independent processes, seven warm calls
and two warmups are diagnostic only. A sustained acceptance run must predeclare
its iteration count, retain 20/60-block planned looks, randomized warm-call
interleaving, ±0.5% parity, ±0.25% A/A calibration and simultaneous intervals
across all selected comparisons. Preserve failures and raw provenance; a faster
ARM64 diagnostic does not accept amd64 or close M2. Require the full Go/LLGo
semantic matrix and at least 90% statement coverage per owned executable package.

Check every opcode with mixed lanes, signed zero, subnormals, infinities,
canonical/arithmetic/signaling NaNs, halfway rounding, saturation boundaries,
input/result aliases, register pressure, cold entries and fuel/cancellation.
NaN checks follow the allowed canonical/arithmetic classes; all other values,
pseudo-min/max selections and sign-only operations require exact bits.

## Implementation and diagnostic record

Engine `c1c63d5` implements all 52 opcodes on both targets. Its four semantic CI
jobs pass, and downloaded artifacts confirm the full 51,139-assertion corpus for
each interpreter/native tier. `7c58d3b` consumes pooled literals without a
preceding local copy; `8d697a1` targets ARM64 BSL/BIT/BIF results directly;
`0084b38` prioritizes literals used on backward-edge spans. No relaxed floating
arithmetic, reciprocal approximation or optional CPU instruction is used.

All 275 baseline/reference oracle cases and 110 initial candidate cases pass.
The first 10,000-iteration diagnostic measured the pre-native baseline
sequentially and candidate/references with interleaved warm calls. It demonstrates
the large helper-boundary reduction but cannot resolve small baseline changes.
Its apparent old-family regressions are retained, not discarded: 44
function/target/fuel-tier code hashes across the eight earlier families are
identical before and after the 52-opcode change. A second diagnostic applies
identical call interleaving to baseline and candidate, and increases floating
workloads to 100,000 iterations. Short old-family results remain noisy.

The 100,000-iteration diagnostic interleaves both Go/LLGo candidates, both initial native
floating-point binaries (`c1c63d5`), wazero 1.12.0, Wasmtime 48.0.1 and Wasmer
7.3.0. Each family has three process blocks, seven timed calls and two warmups;
startup, compilation, instantiation and first call are separate record fields.
These are local macOS arm64 medians in ns/source iteration, **not calibrated
acceptance**:

| Workload | Initial LLGo | Final Go | Final LLGo | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| arithmetic-f32x4 | 5.083 | 5.039 | 5.079 | 5.092 | 4.940 | 4.950 |
| arithmetic-f64x2 | 5.818 | 5.790 | 5.785 | 5.770 | 5.620 | 5.630 |
| minmax-f32x4 | 7.930 | 6.814 | 6.811 | 6.861 | 6.700 | 6.710 |
| minmax-f64x2 | 7.554 | 6.827 | 6.785 | 6.887 | 6.690 | 6.730 |
| rounding-f32x4 | 10.442 | 10.512 | 10.440 | 10.482 | 10.170 | 10.210 |
| rounding-f64x2 | 12.225 | 12.237 | 12.228 | 12.205 | 11.930 | 11.980 |
| convert-signed-f32x4 | 5.036 | 5.043 | 5.033 | 5.616 | 5.610 | 5.650 |
| convert-signed-f64x2 | 6.803 | 6.855 | 6.795 | 7.385 | 7.370 | 7.400 |
| convert-unsigned-f32x4 | 5.030 | 5.033 | 5.023 | 5.610 | 5.610 | 5.650 |
| convert-unsigned-f64x2 | 6.793 | 6.858 | 6.822 | 7.365 | 7.380 | 7.440 |
| casts-f32x4 | 5.400 | 5.395 | 5.402 | 5.289 | 5.300 | 5.330 |

Direct result selection reduced the two min/max LLGo families by approximately
10–14% versus the initial native lowering in matched blocks. Four conversion
families lead the fastest reference by approximately 7–10%; arithmetic,
min/max, rounding and casts still trail their fastest references by approximately
1–3%. Host timing changed materially between diagnostics, including competitor
rows; compare paired runs rather than absolute medians from different runs.
The loop-priority change has no separately established speedup. No sustained
20/60-block floating-point acceptance run has been justified yet, and amd64
performance remains unmeasured. **The feature remains blocked for merging.**

Retained raw records, including all failures (SHA-256):

- `fixture-checks.json`: `ad28c1b5d4b7ebe959a9a05418fffcdf76ff2c7cb5348eb5862a88be9153cb8a`.
- `initial-results.json`: `80e13c49bdd27ce281afeb181ec91f726daf40bc36195ef04071c64cbb4f1c74`.
- `matched-results.json`: `3440bd82781075e5dbc09150ed08bb8f474e4d1391dfc24e5266265a022f29d8`.
- `pooled-results.json`: `1f00465c254528b4b361eafb14c48869cdfec4229f0f55f4527344ceadb2f9ff`.
- `select-results.json`: `7ba74950b3eba943e9d9ab9b0a2f4242f7d7111e06c6a6829e209f0c35c58ba6`.
- `loop-results.json`: `ef205dda8b01396775ecddd61016b6c828787b728820fbd8b98ac616637d5527`.

## Boundary and policy isolation

A follow-up diagnostic uses one million source iterations per call for the seven
non-conversion families, with the same three-block/seven-call protocol. The
production build is unchanged. A separate, temporary checkout disables generated
interrupt polls solely to estimate policy effects; its patch and source hashes
are retained, and it is **not a production candidate or an acceptable merge
configuration**. Removing polls also changes code layout, so the difference is
not a precise isolated instruction-cost estimate.

| Workload | LLGo API, normal policy | LLGo API, diagnostic no polls | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| arithmetic-f32x4 | 4.966 | 4.927 | 5.288 | 4.908 | 4.916 |
| arithmetic-f64x2 | 5.662 | 5.640 | 5.971 | 5.619 | 5.627 |
| minmax-f32x4 | 6.678 | 6.674 | 6.881 | 6.660 | 6.678 |
| minmax-f64x2 | 6.673 | 6.679 | 6.865 | 6.655 | 6.657 |
| rounding-f32x4 | 10.329 | 10.273 | 10.664 | 10.182 | 10.180 |
| rounding-f64x2 | 12.074 | 12.008 | 12.388 | 11.937 | 11.931 |
| casts-f32x4 | 5.264 | 5.273 | 5.473 | 5.272 | 5.275 |

Units remain ns/source iteration. Normal-policy min/max and casts now fall within
the nominal parity band in these medians; arithmetic/rounding still trail the
fastest reference by approximately 0.8–1.5%. This short diagnostic has neither
simultaneous acceptance intervals nor A/A calibration and cannot close a gate.

One-source-iteration arithmetic/rounding probes separately measure LLGo public
calls at 6.25/8.67 microseconds, versus 1.79/2.50 microseconds with the existing
internal dispatch mode, which reuses its execution outside the public API. Both
retain cancellation. These probes include guest setup and one iteration; they
are not empty-entry latency. Public calls currently allocate an execution and
native arena, register cancellation and release them on every call. Startup,
compilation and instantiation remain outside the warmed timer. For example,
the 100,000-iteration arithmetic-f32x4 diagnostic reports LLGo compile/validate
at 0.389 ms, separately from 0.0075 ms instantiation and 0.507 ms first call.
The corresponding compile/validate medians are 0.164/0.915/1.441 ms for
wazero/Wasmtime/Wasmer; these are illustrative diagnostic phases, not accepted
compile-latency gates.

Disassembly of the pinned Wasmer library shows the same FMUL/FADD/FDIV arithmetic
chain. Its loop has six instructions; the owned loop additionally preserves
continuation state and checks cancellation. Independent source hashes confirm
that the no-poll experiment is the sole diagnostic patch. It was never pushed;
the temporary checkout was removed after recording its patch and outputs.

- `policy-results.json`: `0d355cab165047ffb617714a39ae2fc75432fb59e849e296b40074b76b3575ee`.
- `boundary-results.json`: `f4206210a9b220a090976bf0286e52230f05a656b6e1a2fcc78a78dc667ac4eb`.
- `nopoll-diagnostic.patch`: `2c1e82fa3a8b3b46a9a730322755d62e784f6dbb7365ef8180970b1d1742af05`.
