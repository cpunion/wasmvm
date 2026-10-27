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

## Optional x86 instruction profiles

After the portable 39-opcode implementation, add explicit SSSE3 and SSE4.1
compiler inputs. A zero mask remains baseline SSE2. Only the native platform
adapter detects host CPUID; code generation stays pure Go and rejects unknown
features or an x86 mask on another target. The runtime passes the detected mask
at compilation and does not assume support from the host compiler or build CPU.
Record the mask in phase reports. Test forced baseline/SSSE3/SSE4.1 combinations,
including all mixed-lane/alias/spill/fuel/cold-entry contracts, and execute an
optional instruction only on a host advertising it. No AVX/OSXSAVE contract is
introduced. Use detected profiles for byte permutations, packed population
count, Q15 multiplication, widening, unsigned narrowing, i32 multiplication and
floating-point rounding; preserve portable legalizations for masked features.

## Packed shifts and reductions

The complete SIMD inventory still includes scalar-per-lane legalization for
variable shifts and reductions. Replace those with packed NEON/SSE2 sequences,
covering every lane width, masked counts, signed i64 right shift, byte-lane
isolation, all-true, any-true and sign masks. Add four dynamic-count shift and
four reduction families before implementation; retain the earlier families and
CPU-mask combinations. Pure throughput and public-call costs remain separate.

## Retained implementation and measurement record

`2d859c9` completes all 39 opcodes; `4c30fad` removes unnecessary NEON scratch
copies; `49a5719` introduces explicitly detected x86 profiles. `64e5e69`
implements packed shifts/reductions. `24a19bb` and `0ccf8f4` share legalization
masks with the literal pool, shorten ARM64 reductions and use SSE4.1 floating
selection. The compiler's entire fixed-SIMD inventory is now 236 operations;
state/bulk helpers and the ordinary guest-call frame driver remain separate.

Four Go/LLGo × target semantic jobs pass at `49a5719` and `64e5e69`, including
51,139 assertions under the interpreter and both native tiers. The local final
package matrix passes, with codegen at 97.25% (4101/4217 statements), every
measured executable package above 90%, and all 67 measurement tests passing.

The initial 14 integer families pass 490 independent baseline/head/reference
oracle cases. The first native ARM64 lowering is far faster than the historical
helper path but trails the fastest references in several families. Selecting
allocated NEON registers reduces widening/widening-multiply/pairwise costs by
roughly 20–30% in matched short diagnostics. These measurements retain startup,
compile, instantiate, first-call and warm phases rather than subtracting them.

The first eight packed-shift/reduction families pass another 280 independent
oracle cases. LLGo shift medians improve about 4.5–13.5 times over the former
lane-by-lane sequence. Initial i32/i64 reductions regress; those failures remain
in `packed-results.json`. The subsequent hoisted-mask diagnostic fixes the
observed reduction regression but still has slower competitor comparisons.
Absolute medians differ materially between runs, including reference engines;
use matched process blocks and do not infer calibrated parity from a short run.

Hosted scope `native-simd` reuses `native_ci.py`, existing pinned builders and
warm-call barriers. It runs both targets and both compilers, records feature
masks, and avoids semantic/CLI repetition. Run
`gh workflow run benchmarks.yml --ref <branch> -f suite=native-simd -f baseline=<full-sha>`.
Use a full 40-character SHA for commit baselines: checkout treats a short SHA
as a branch pattern. Scope completion is diagnostic; A/A calibration and the
established simultaneous sustained intervals are still required before merging.

[Run 40509770203](https://github.com/cpunion/wasmvm/actions/runs/40509770203)
completes 33 families and 924 alternate-input oracle checks on each target.
The preceding amd64 report confirms actual mask `0x3` and faster SSE4.1 rounding,
but shuffle/popcount and several other families still trail the fastest
reference. The first ARM64 packed reductions also fail parity. These results
are retained as failures; no threshold, profile or workload is silently weakened.

Local raw records (SHA-256):

- `initial-results.json`: `40b580641a8924b2d89745b5a47eeebc0b2f05dbdf2437d3fe762216d51c222e`.
- `allocated-results.json`: `3f273e4ee214443f052223586e7c3dcf1351c1ffad421cac7bce0251710aefc1`.
- `packed-results.json`: `bfe101f1f5117567b895158dea404e87a2d32032c3af7fa4d5fd1969b98d14cc`.
- `hoisted-results.json`: `841796082da22012797a8462bc37a1556d9fc58e4a31ae06352fddf4239eee25`.
