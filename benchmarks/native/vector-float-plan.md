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
