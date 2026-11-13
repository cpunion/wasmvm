# Bounded nested scalar calls

Go competitiveness remains the priority. This change improves one core call
family; full-profile performance and release acceptance remain open.

## Mechanism and validation

Previously, a pure integer wrapper that called another function could not be
expanded. Repeated nested calls retained the software frame setup, argument
copies, checked entry dispatch and return bookkeeping. The existing scalar
expression planner now follows direct calls within the same immutable module,
with at most eight original frames and 256 visited slot instructions per tree.
Repeated visits to a shared definition consume the same analysis budget.
Recursion, imports, effects, references, memory, trapping arithmetic and
conditional/backward control still reject optional expansion.

The planner computes the maximum original nested frame and slot requirements,
rather than summing sequential children. The runtime selects the expanded body
only when both current invocation quotas fit, including indirect candidates;
otherwise it executes the ordinary body. Metered execution retains the original
instruction/fuel path. Existing cold continuations and the same expression,
register and instruction selectors remain in use. No additional runtime or
private unchecked entry is introduced.

Tests execute every cold caller PC with complete raw-frame comparisons, aliased
arguments/results, i32/i64 boundaries, recursive graphs and expansion budgets.
Direct/indirect calls retarget at runtime and are compared with the interpreter
at exact and insufficient stack quotas, with oversized physical workspaces and
both fuel modes. All 32 Go test packages pass; every package with executable statements
exceeds 90% CI-equivalent integration coverage; codegen is 98.01% and nativeexec 93.27%. Relevant
Go race checks and pinned LLGo macOS execution also pass. The final hosted matrix
and review state remain separate merge requirements.

## Fixed core comparison

The source event compares c0dacbe3 with 21237ca9 on macOS arm64, Go 1.27.3,
LLGo 56fa9ae `-O3 -lto=full`, LLVM/Clang 22.1.8 and normal GC. The authored fixture
has three nested integer calls per source iteration. Independent affine-map
exponentiation supplies its oracle, including alternate seeds/counts.

Twenty fixed blocks use 280 fresh processes: seven configurations and two
identical-binary aliases. Each process separates first call and two warmups
from seven individually interleaved calls of 20 million iterations. Every result
is checked. No owned build/test/profile/pruning overlaps timing.

| Configuration | Median ns/source iteration |
| --- | ---: |
| Go before | 29.489 |
| Go after | 1.119 |
| LLGo after | 1.113 |
| wazero 1.12.0 compiler | 3.421 |
| Wasmtime 48.0.1 Cranelift speed | 1.955 |
| Wasmer 7.3.0 Cranelift | 1.943 |
| Clang 22.1.8 native arithmetic control | 1.110 |

Paired Go head/reference ratios, individual diagnostic 95% intervals:

| Reference | Ratio | Interval |
| --- | ---: | --- |
| wazero | 0.3238 | [0.3193,0.3273] |
| Wasmtime | 0.5707 | [0.5647,0.5758] |
| Wasmer | 0.5751 | [0.5692,0.5810] |
| Native arithmetic | 0.9977 | [0.9845,1.0103] |

The Go head, LLGo and all four reference/control A/A intervals resolve within
the predeclared 2.5% margin. The old Go A/A interval does not
([0.9370, 1.0180]); its outliers remain in the raw event. The median source
improvement is about 26 times, with that qualification. These scoped diagnostic
intervals do not close a simultaneous full-portfolio gate or another platform.
Native arithmetic has no Wasm isolation, checks, quotas or cancellation contract;
it is a separate control. Empty volatile assembly at its function entry prevents
cross-timer common-subexpression hoisting without changing the arithmetic loop.
An initial hoisted smoke result was rejected and never enters the fixed cohort.

Separate emission diagnostics show the owned root's MADD and both outer-call
and backedge cancellation probes, with no guest call in its hot loop. The
Wasmtime speed image retains all three native BL calls, stack checks and
prologues/epilogues. This explains a call-path difference for this fixture;
it is not a claim about all Cranelift configurations or arbitrary programs.

## Complete P3 remains behind

A separate fixed 24-process screen per payload uses four Stores, 16 clients,
20 batches of 512 fully checked requests per process, snapshot input, buffered
output and normal GC. Source/compiler/driver identities remain explicit.

| Payload | Go base → head CPU µs/request | Head requests/s | Wasmtime requests/s |
| --- | ---: | ---: | ---: |
| 16 B | 96.56 → 96.32 | 37,948 | 50,026 |
| 64 KiB | 128.76 → 129.38 | 27,844 | 43,893 |

Paired head/base CPU ratios are 1.0016 [0.9976,1.0056] and 0.9994
[0.9911,1.0124]. Several P3 A/A controls remain unresolved. There is no accepted
P3 gain or ranking from this screen. Managed allocations remain 705/789 per
request. Compilation wall points are retained separately and are not a compile
speedup claim. The core improvement does not remove the complete component,
ownership, host/runtime and general guest-call costs.

The new diagnostic extends the survey from 68 to 69 cases in the same eight
families and the SIMD regression subset from 42 to 43 cases. Historical aggregates
are not pooled or reweighted into an acceptance claim. Two earlier entry/target
guard experiments passed semantics but failed to improve P3 and were not adopted.

[All identities, phases, controls and raw observations](data/performance-nested-scalar-calls-2026-11-13.json)
retain the evidence. Next work targets general guest call frames and complete
component request costs; performance acceptance stays open.
