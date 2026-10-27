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

## Final optimized comparison

Code `0ccf8f4`, checked at documentation head `eaee3f7`, passes all four
[semantic jobs](https://github.com/cpunion/wasmvm/actions/runs/40512028373).
Downloaded reports again verify all 12 compiler/target/backend-tier combinations
with 51,139 assertions and 148 scripts each. Local statement coverage is
97.25% (4101/4217 statements) for codegen; every measured executable package
remains above 90%.

[Final two-target diagnostics](https://github.com/cpunion/wasmvm/actions/runs/40512069149)
compare the same 33 families against baseline `64e5e69` and all three references.
Both targets complete 924 alternate-input checks and retain all phase fields,
CPU masks, binary/source hashes and randomized sample orders. Selected paired
LLGo medians, ns/source iteration:

| Target | Family | Baseline | Final | Fastest reference | Final/reference |
| --- | --- | ---: | ---: | ---: | ---: |
| amd64 | shuffle | 3.307 | 2.535 | 1.535 | 1.652 |
| amd64 | popcount | 3.087 | 2.446 | 2.262 | 1.081 |
| amd64 | minmax-f32x4 | 12.067 | 8.929 | 7.964 | 1.121 |
| amd64 | minmax-f64x2 | 12.141 | 8.311 | 7.671 | 1.083 |
| arm64 | reduce-i16 | 11.565 | 9.825 | 9.020 | 1.089 |
| arm64 | reduce-i32 | 12.648 | 9.919 | 9.010 | 1.101 |
| arm64 | reduce-i64 | 13.674 | 9.413 | 8.898 | 1.058 |

The paired runs show the intended improvements without establishing parity.
ARM64's four conversion families lead the fastest reference by approximately
6–9%; several other families remain slower. The 100,000-iteration public API
rows still include each call's execution setup/teardown, so they are not pure
instruction latency. Earlier one-iteration/dispatch diagnostics separate that
boundary cost; further compiler and VM work must address the remaining gaps.
There is no sustained acceptance or A/A-calibration claim. PR #26 remains
blocked independently on its existing calibration gate; subsequent work stays
on the implementation branch.

Final raw artifacts (SHA-256):

- `native-amd64-comparison/results.json`: `81a194bb9c55260be86d37540d79e6a343530565ab644592f98a26c923afafdd`.
- `native-arm64-comparison/results.json`: `c8a2e213cf8ca6527443a896538b783ed17643f72c9c5388b21a87edc6bd7544`.

## AVX2 profile scope

Baseline `c980109` completes fixed SIMD but retains extra x86 copies and legacy
variable-shift dependencies. Add an explicit 128-bit AVX2 profile: CPUID must
advertise AVX/OSXSAVE and AVX2, and XGETBV must confirm XMM/YMM state before any
VEX instruction is selected. Do not infer OS support from the CPUID AVX2 bit.
Keep SSE2/SSSE3/SSE4.1 masked profiles executable; no AVX-512 or wider guest
vector contract is introduced. Three-operand encodings should preserve aliases
without scratch copies and use per-lane variable shifts where appropriate.

Affected scope is the existing 33-family native SIMD suite, with unchanged
100,000-iteration diagnostics, alternate-input oracles, phase boundaries and
three references. Include the existing integer/call/memory regressions. Test
all eight explicit profile masks for emission; execute only host-supported
masks. Preserve every cold PC, all slots at exits, exact fuel, cancellation,
NaN/zero classes and the full Go/LLGo semantic matrix. Keep slower measurements;
missing or inconclusive sustained/calibrated evidence continues to block merge.
