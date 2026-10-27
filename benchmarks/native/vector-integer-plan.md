# Complete native fixed-SIMD integer scope

Baseline: `3d64ec5` (production code equal to `0084b38`). Lower the remaining
39 Core 2.0 fixed-SIMD integer/shuffle instructions on amd64 and arm64: shuffle,
swizzle, byte population count, narrowing, signed/unsigned widening and widening
multiplication, pairwise widening addition, rounded average, Q15 multiplication
and signed dot products. Preserve SSE2-only portability, full-width vector
ownership, all cold continuations, exact fuel and bounded cancellation.

Use NEON packed operations and SSE2 packed legalizations where available.
General byte permutations on SSE2 require explicit legalization; do not assume
SSSE3 or silently execute a guest interpreter/helper. Report this baseline's
performance limits explicitly before introducing optional CPU profiles.

Before implementation, declare shuffle/swizzle, population count, narrowing,
widening, widening multiplication, pairwise sums, averages, Q15 and dot-product
workloads. Every lane contributes to a Python scalar oracle. Verify alternate
seeds and iteration counts against pinned wazero/Wasmtime/Wasmer. Compare own
Go/LLGo baseline/head through the public native API, with separately recorded
startup/compile/instantiate/first-call/warm phases. Retain prior integer, memory,
call and SIMD regression families.

Initial three-process/seven-call/two-warmup measurements are diagnostics. Any
sustained acceptance run must predeclare iteration counts, use the established
20/60-block planned looks, randomized warm-call interleaving, simultaneous
intervals, ±0.5% parity and ±0.25% A/A calibration. Missing/slower/inconclusive
required compiler/target evidence blocks merging. No milestone closes from a
narrow workload or semantic CI alone. Require all Go/LLGo semantic jobs and at
least 90% statement coverage per owned executable package.
