# Scalar floating-point development scope

Status: implemented on a development branch; not accepted and no second PR.

## Implemented scope

The shared amd64/arm64 function compiler lowers all 44 f32/f64 comparisons,
arithmetic, rounding, min/max, sign and reinterpretation operations plus all
26 promotion/demotion, checked and saturating conversions. Together with the
66 integer operations this covers all 136 Core 2.0 scalar numeric operations.
amd64 uses baseline SSE2; arm64 uses scalar FP and hardware saturation.

Integer, float and conversion instructions share one bounded copy/result
forwarding pass, register caches, continuation protocol and trap/fuel/cancellation
contracts. The longer multiply-add selector takes precedence over shorter copy
groups. Checked casts read their allocated FP source directly and publish exact
trap categories/locations; source and destination slots survive failed casts.
Every original PC retains independent semantics. Native memory, calls, SIMD and
public runtime integration remain open; this is not full native Core 2.0 support.

## Validation

Both native tiers execute all **11,970 assertions** in ten pinned numeric scripts
(11 modules). The **263** invalid/malformed frontend commands are explicitly
excluded from this native execution count. Go race and LLGo arm64 tests pass;
isolated Go amd64 instruction execution also passes. The latter uses
Docker/Rosetta with Go 1.27.0, not native amd64 timing or the final pinned host
matrix. Codegen statement coverage is **96.2%**. Tests include IEEE edges,
random inputs, exact traps/fuel, all-slot results at independent continuations,
register pressure, cancellation, branch ranges and optimization interactions.

## Declared performance portfolio and remaining gate

The ten owned fixtures were declared before implementation: dependent f32/f64
arithmetic, min/max/comparison, rounding/square-root, and signed/unsigned
conversion chains. Preserve the integer regression series. Each fixture uses
`run(i32 seed, i32 iterations) -> i32`, seed `305419896`, one million iterations,
independent Python results, and wasm-tools 1.258.0 binaries with hashes.
Startup/preparation/first call/warm execution remain separate. No unsupported
native baseline is replaced with interpreter timing.

Latest local M4 Max diagnostic, three interleaved process blocks, 150 alternate
seed/count checks, ns/source iteration (several operations per iteration):

| Case | Go | LLGo | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| arithmetic-f32 | 3.586 | 3.512 | 3.656 | 3.693 | 3.441 |
| arithmetic-f64 | 3.982 | 3.964 | 3.981 | 3.898 | 3.935 |
| minmax-f32 | 1.791 | 1.810 | 1.795 | 1.774 | 1.796 |
| minmax-f64 | 1.803 | 1.788 | 1.815 | 1.774 | 1.766 |
| rounding-f32 | 6.662 | 6.672 | 6.490 | 6.538 | 6.355 |
| rounding-f64 | 7.736 | 7.676 | 7.779 | 7.434 | 7.379 |
| convert-signed-f32 | 9.582 | 9.594 | 19.582 | 9.548 | 9.559 |
| convert-signed-f64 | 9.504 | 9.666 | 20.039 | 9.413 | 9.362 |
| convert-unsigned-f32 | 9.652 | 9.736 | 19.989 | 9.469 | 9.412 |
| convert-unsigned-f64 | 9.749 | 9.631 | 20.035 | 9.634 | 9.387 |

These are advisory point estimates, not calibrated merge evidence. Min/max has
improved from roughly 2.1 to 1.8 ns/iteration; conversion chains from roughly
10.5 to 9.5–9.7, with remaining reference deficits in several rows. Require
parity or better for **every** affected family and each Go/LLGo host on Linux
amd64 and macOS arm64, with the declared A/A calibration and simultaneous
confidence bounds. The scalar feature remains unaccepted.

Raw staged observations, compiler/binary/fixture identities and restoration
checks remain in local `artifacts/native-float/` and
`artifacts/native-conversions/`. IEEE-bit bounds checks and combined cold NaN
dispatch had no stable measured benefit and were removed. No threshold changes,
workload-specific kernels, or competitor fallbacks were introduced.
