# Avoid clearing direct-call parameters before their full copy

Direct private calls previously cleared every child slot, then immediately
copied complete 128-bit parameters over the cleared prefix. The compiler now
clears only the non-parameter suffix. Full parameter copies finish before the
child becomes live. Whole-frame quota/capacity checks, zeroed locals/temporaries,
roots, bounded cancellation and metered behavior remain. Dynamic indirect calls
retain their full clear. No metadata layout, executor or guest shortcut is added.
See the [ABI contract](../design/native-abi.md#checked-internal-call-entries).

## Sources and protocol

Integrated base **fd625a2d**, final implementation **55f3e93a**. Reused baseline
binaries identify **9de037fb**; Git verifies all Go/module files match integrated
base. Two-file symmetric source snapshots, compiler profiles, hashes and original
identities remain. Go 1.27.3, LLGo 6a482f7f, LLVM 22.1.8, `-O3 -lto=full` and
default production GC are unchanged. References: wazero 1.12.0, Wasmtime 48.0.1,
Wasmer 7.3.0 and an unchecked LLVM C affine oracle (2N steps normalized per guest
iteration, without VM safety contracts).

Host: M4 Max/macOS arm64, with other user compiler/link/test and desktop activity.
No owned builds, tests, analysis, source edits or pruning overlap timed cohorts.
All load snapshots and A/A aliases remain. Fixed cohorts are neither extended
nor pooled. Intervals are individual block-bootstrap 95% intervals. Controls
resolve when their whole interval lies within 0.975–1.025. Absolute costs across
cohorts are not directly comparable; no portfolio acceptance follows.

## General native calls

The final counter audit has ten blocks/40 processes and twenty individually
interleaved five-million-iteration calls plus a separate flush. An independent
complete comparison has twelve blocks/192 processes: Go/LLGo base/head, three
references and C, all with aliases. Readiness, compilation, instantiation, first
call, two warmups and seven measured calls stay separate; all outputs are checked.

| Go final/base counter audit | Ratio [95% interval] |
| --- | ---: |
| CPU | 0.82145 [0.81985, 0.82336] |
| Instructions | 0.99383 [0.99376, 0.99392] |
| Cycles | 0.82138 [0.82052, 0.82246] |
| Cycles/instruction | 0.82648 [0.82564, 0.82751] |

Both own controls resolve for all metrics. CPU falls **17.86%** while instructions
fall only **0.62%** (about 169 to 168). Avoiding repeated writes improves cycles
far more than instruction count. Repeated-store/forwarding dependencies are a
plausible contributor, but their exact contribution versus changed code/data
layout remains unisolated; this is not a hardware-counter proof of a particular
stall. The independent warm comparison gives Go **0.82030 [0.81714, 0.82317]**,
LLGo **0.82465 [0.81978, 0.82860]** and LLGo/Go **1.00546 [1.00132, 1.00970]**.
All warm controls resolve. The small LLGo/Go difference is retained; it does not
identify a new compiler defect. A separate initial 40-process screen gives CPU
0.82172 [0.81956, 0.82355], with resolving CPU controls; its base wall control
is unresolved and it is not substituted for final source measurements.

| Final Go/reference warm time | Ratio [95% interval] |
| --- | ---: |
| wazero | 2.23426 [2.21724, 2.25532] |
| Wasmtime | 2.37824 [2.36474, 2.39069] |
| Wasmer | 2.38168 [2.37107, 2.39369] |
| Unchecked C | 2.38342 [2.37033, 2.39584] |

All reference warm controls resolve. Median Go/LLGo times are about 5.18/5.22 ns,
versus 2.33 ns wazero and 2.18 ns for the others. Reference safety contracts
remain distinct. Competitive parity is not achieved; remaining frame/metadata,
entry preparation and register allocation work stays open.

## Full requests, adverse points and independent audit

Each P3 payload has eight blocks/80 processes, Go/LLGo base/head and Wasmtime,
all with aliases. Four Stores/sixteen clients execute twenty interleaved
512-request batches plus a separate flush, using snapshot input and buffered
output. Bodies, statuses, headers, trailers and CRCs are all checked.

| Go final/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU | 1.00944 [0.99967, 1.01894] | 0.99659 [0.98998, 1.00354] |
| Instructions | 1.00004 [0.99895, 1.00099] | 0.99894 [0.99780, 1.00009] |
| Throughput | 0.97927 [0.96970, 0.98999] | 1.00505 [0.98202, 1.02648] |

Both Go CPU/instruction controls resolve for both payloads. Neither establishes
a CPU gain. The adverse 16 B throughput point is retained with an unresolved
head throughput control; both 64 KiB throughput controls are unresolved.
Go/Wasmtime CPU points are **2.064x/2.115x** with unresolved reference CPU
controls; throughput points **0.809x/0.702x**, with the 64 KiB reference throughput
control unresolved. Managed allocations remain about 696/756 per request;
foreign Rust/C allocations are not counted by these host counters.

A separate fixed twelve-block/48-process Go-only 16 B regression audit uses the
same checked request protocol and aliases. CPU is **1.00137 [0.99550, 1.00748]**,
throughput **0.99545 [0.98787, 1.00238]**, instructions **0.99773
[0.99528, 0.99988]**; both own controls resolve for all three. This does not
reproduce a resolved CPU/throughput regression. Median Go collections remain 35
per 10,240 requests in both original and audit events, allocations are unchanged,
and the instruction change is small. Exact scheduling/layout contributions to
the original adverse point remain unisolated. The audit neither extends the
original cohort nor relabels it a pass; no full-request speedup is claimed.

LLGo default-GC CPU points are **0.98345 [0.95805, 1.00967]** and **0.97631
[0.93145, 1.01621]**, with unresolved own controls. Median collection counts are
39/39 for 16 B and 60/60 for 64 KiB; the latter ranges 34–111/34–72. Allocations
remain about 1165/1430. Separate eight-block/32-process bounded no-GC diagnostics
per payload require zero collections, using five measured 512-request batches
plus a flush. Their CPU ratios **1.00326 [0.99490, 1.01104]** and **0.99734
[0.98622, 1.00646]** cross parity; 16 B base CPU control is unresolved, both
64 KiB CPU controls resolve. Instruction ratios also cross parity, with resolving
controls. The adverse 16 B no-GC point is retained, not called an improvement.
Production GC is unchanged; collection variability remains separately tracked
in [LLGo #3106](https://github.com/xgo-dev/llgo/issues/3106).

General-call Go cold compile-wall points increase **6.17%** in the counter audit
and **5.64%** in the complete comparison, with intervals crossing parity and
unresolved cold controls. P3 Go compile ratios are 0.98318/0.98948, LLGo
1.00159/1.00318, also with unresolved controls. No exact cold regression or
cold speedup is established; preparation is not folded into warm costs.

## Validation and evidence

All 32 Go root test packages pass and all 31 measured executable packages reach
at least 90% statement coverage. LLGo executes affected codegen/nativeexec tests
under the pinned production profile. Added tests execute both target encodings
on their native hosts: complete parameter words and outer guards, zero/all-
parameter frames, local suffixes across 4 KiB boundaries, large parameter
prefixes and chunk cancellation. Existing mixed results, roots, quotas, reentry,
proposal and full WASI tests remain green. Complete final-head dual-platform
Go/LLGo CI is required before merging.

The initial new test pre-set cancellation and correctly hit the entry wrapper's
Poll before its generated body, contradicting the test's expectation. The final
fixture sets Control after validation to exercise the zeroing loop. Initial
source/failure and corrected Go/LLGo logs are retained; no production engine
failure is hidden.

The [machine-readable report](data/performance-native-parameter-initialization-2026-11-15.json)
links **544 process records in eight separate cohorts**, plans, compiler/source
identities, partial source overlays, load snapshots, raw phases, output checks,
GC/allocation counters and rusage. Only unused per-request timing vectors become
count/extrema/canonical-JSON hashes; full originals remain local. General-call,
component allocation/dispatch, interpreter/other kernels, cross-platform and
release acceptance remain open.
