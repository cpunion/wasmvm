# Heap-cache register reuse and allocation throughput: 2026-11-08

Source `93caad4` compares with main `d543b9e`. Compiler, reference and CPU-feature
pins stay unchanged: Go 1.27.3, LLGo `3571c25 -O3 -lto=full`, wazero 1.12.0,
Wasmtime 48.0.1, Wasmer 7.3.0 and WasmEdge 0.17.1. Controlled acceptance remains
open. Two independent changes address an amd64 allocation restriction and a
misinterpretation of GC timing summaries.

## Existing registers and entry lifetime

The checked amd64 heap cache previously required two callee-saved registers.
Under scalar pressure this rejected a profitable candidate even when the
ordinary guest pool still contained free caller registers. The existing
one-pair limit, frequency priorities and ordering now apply to the entire
existing guest GPR pool. No bank, instruction set or execution framework is added.

A cache lives only inside one native entry, exactly like caller-register guest
homes. Calls that enter another body, runtime helpers, polls and host/GC
boundaries reenter through the normal header and reset its key. Inline expansion
uses the same exclusion inventory. Hits still compare the full non-null handle;
misses retain null/index/generation checks. Array bounds, field widths, trap
sources and precise fallback retain their contracts. Existing slot encoders
already preserve this pool; cache registers are excluded from slot allocation.

Only the original indirect-access x86 module changes: one cache replaces zero,
with its function growing **3943 → 4455 bytes (13.0%)**. The increase contains
hit branches and the complete checked miss/cold paths. All other 21 x86 modules
remain byte-identical. Static size/selection proves no elapsed improvement.

## Allocation medians and observed pauses

The allocation fixture retains a linked chain. Its warm calls include periodic
collector pauses. The original reporter preserved every raw sample but selected
the median call inside each process. Interpreting that typical-call latency as
sustained allocation throughput omits much of the pause cost. The
[earlier GC-cycle study](performance-native-allocation-2026-11-02.md) already
explains this distinction; later routing summaries failed to carry it forward.

The following re-reduction uses **only the existing event 42376300827**. It
collects no additional timing and does not mix source revisions or workloads.
The statistic is the median of five independent process means, with all seven
measured warm calls included in each process. First calls and two warmups retain
their own roles. Every original median, process mean, sample and both A/A
calibrations remain in the record.

| Allocation cohort | LLGo ns/object, process means | Wasmtime ns/object, process means | LLGo / Wasmtime | LLGo / Go | All mean A/A |
| --- | ---: | ---: | ---: | ---: | --- |
| Original, 10,000 | 77.93 | 156.49 | 0.498 | 1.287 | Inconclusive |
| Sustained, 300,000 | 88.34 | 113.13 | 0.781 | 1.147 | Inconclusive |

The earlier **2.858×** sustained ratio describes warm medians, not complete
allocation throughput. Neither reduction establishes a calibrated winner.
Seven calls per process also do not guarantee a complete steady-state GC cycle.
The remaining LLGo/Go difference is an observation, not an independently
reproduced compiler bug. Wazero/Wasmer still reject these GC modules; no equivalent
native C GC case is invented.

The reporter now marks allocation cases as `mean-including-pauses-v1`, reduces
each process independently and calibrates that same statistic. It also emits
the original median diagnostic and its controls. Other workloads keep process
medians. Historical records without a marker keep their original reduction;
no old samples, flags, counts or calibration results are overwritten.

## Qualification and new measurement scope

Complete local codegen/nativeexec/public-VM tests pass with both host compilers;
vet and all **89** benchmark-tool tests pass. Codegen coverage is **97.73%
(8616/8816)** and nativeexec **92.54% (459/496)**. These are statement counts,
not specification coverage. A new high-scalar-pressure variant exercises every
field width, array bounds, invalid handles and complete cold frames while
forcing the cache into caller registers. Inventory/reset tests cover both
caller pairs and exhausted budgets. Existing call, collection and forced-poll
contracts remain in the final semantic matrix.

The first broad local command lacked the pinned corpus in the new worktree and
failed explicitly. After linking the unchanged verified inputs, both compiler
runs pass. The setup failure remains retained and is not a semantic pass.

All **88 fresh ARM64 publications** validate checksums. All 22 base/head and
Go/LLGo images and entry lists match; no redundant local warm campaign is
collected for unchanged guest code. x86 inspection compiles/links all 22 modules
with feature mask 7 and stops before publication. Actual x86 guest execution
belongs to Linux CI and the
[new source-pinned event](https://github.com/cpunion/wasmvm/actions/runs/42400928847).
It retains all original 22 workloads and the separate fixed sustained-GC cohort,
including all runtimes, phases, controls, unsupported and adverse results.
The final PR body records completed observations; final four-profile semantic
CI, all 31 packages per LLGo target, coverage and review remain merge requirements.

The [machine-readable record](data/gc-cache-throughput-2026-11-08.json) pins the
historical correction, local qualification and publication/module witnesses.
Immutable local archive: `artifacts/gc-cache-throughput-20261108/local-evidence.tar.gz`,
2264116 bytes, 422 files; SHA-256
`6c679bc53afac241f5560af68eb45ca33b50b80b5d2585686ef8ba38a52b13fd`.
