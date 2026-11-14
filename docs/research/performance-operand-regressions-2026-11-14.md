# Portable operand costs and compiler regressions

Go-compiled wasmvm reaching competitors remains the priority. Every observed
regression, including a rejected experiment, retains its source, compiler,
measurement controls, established cause and unresolved questions. Rejecting code
does not resolve its investigation. Whole-portfolio performance acceptance stays
open; the results below concern one dependent-address workload.
[Machine-readable identities, controls, ratios and phase medians](data/performance-operand-regressions-2026-11-14.json)
retain the exact event boundaries.

Production measurements use Go 1.27.3 and LLGo 56fa9ae with normal GC and
`-O3 -lto=full`. The common guest compiler, checked memory, full 128-bit value
copies, fuel, cancellation and trap provenance remain mandatory. No layout flag,
GC override, unchecked slot access or reference-engine fallback is adopted.

## Established mechanisms

The scalar-width prototype cf72501d keeps scalar values and the indexes needed
by full-value copies live together. LLGo introduces 2,300,030 extra writes, all
direct loop stack operands. One first-operand spill executes 2,273,740 times,
accounting for 98.86% of that delta. Moving reads into their opcode handlers
removes the spill. That removes extra work, but does not explain every elapsed
time difference.

The compact-dispatch prototype 26b78825 adds 31,112,300 instructions. Six key
normalization instructions contribute 13,800,180; newly outlined poll,
Memory.Load/Store and runtime.NewSlice2 contribute 13,009,916. These account
for 86.17% of the instruction delta, leaving 4,302,204. NewSlice2 performs bounds
checks and constructs a slice header here; it does not allocate a heap object.

Frozen-binary Callgrind counts include the loop's callees, with exact identical
binary instruction A/A in both events:

| Source | Instructions | Reads | Writes | Direct loop stack writes |
| --- | ---: | ---: | ---: | ---: |
| Counter baseline 5e5e6a44 | 226,569,754|51,273,076|12,868,536|624,598|
| Scalar-width cf72501d | 227,695,664|49,371,352|15,168,566|2,924,628|
| Opcode-local 92d05239 | 202,508,600|41,231,948|12,868,538|624,600|
| Selected-index 0708894e | 211,670,782|43,107,388|12,868,538|624,600|

[Initial counts](https://github.com/cpunion/wasmvm/actions/runs/43548991543),
[opcode-local counts](https://github.com/cpunion/wasmvm/actions/runs/43552741246),
[selected-index counts](https://github.com/cpunion/wasmvm/actions/runs/43556702030).
These are instrumented instructions and memory operations, not physical cycles,
cache misses, branch misses or allocated bytes.

Go has a separate problem when numeric operands move before the numeric switch.
Compared with opcode-local 92d05239, df2b0f87 adds 15,682,056 loop-self instructions
and 2,524,594 reads. Extra feature-flag reloads 1,665,556 plus net stack reads 859,038
account for the entire net read delta. runtime.x86HasPOPCNT is rematerialized
outside its handler although this guest never executes popcount. Sharing the
population-count intrinsic confines those loads to its handler in generated
Go code. This is Go code generation, not an LLGo defect.

The [Go count event](https://github.com/cpunion/wasmvm/actions/runs/43555461126)
failed its strict whole-window instruction A/A: 137 extra instructions arise in
scheduler/time sampling. The loop-self A/A is exact; only that scoped comparison
is used above. The failed whole-window gate remains failed. Disabling asynchronous
preemption for this count-only diagnostic does not disable call-boundary
preemption and is not a production setting.

## Regression ledger

Ratios below are warm time relative to the named counter/source baseline unless
stated otherwise. Different hosts/events are never pooled into a causal contrast.
An unresolved A/A control makes its event inconclusive, rather than a retry that
can be silently discarded.

| Experiment | Adverse observation | Established evidence and remaining work |
| --- | --- | --- |
| PR169 95943842 | LLGo +13.6% EPYC7763 | Relocated 7175-byte loop retains instructions/registers/control except relocations. Fixed text recovers most loss; simple alignment does not recover all hosts. Physical cycle attribution remains open in LLGo #3094. |
| Flag-only 468e6c28 | Go +0.74% Intel; LLGo +0.68% vs original fast AMD | A hot step-count store remains. The retained local-counter change removes it and recovers measured points. Source ablation is not complete microarchitectural attribution. |
| Compact 26b78825 | LLGo +25.6%/+33.1% AMD | Key normalization and lost helper inlining explain86.17% of extra instructions; residual work/cycles remain in LLGo #3098. Rejected. |
| Guarded key 4326fe36 | LLGo +19.5% Intel/+35.2% AMD; Go +0.37% AMD | Removing key math retains outlining and the 776-byte frame. Math alone is insufficient; Go's small regression and residual LLGo cycles remain open. Rejected. |
| Scalar-width cf72501d | LLGo +1.0%/+1.4% Intel and a small AMD7763 loss | Hot first-operand spill reproduced independently; elapsed-time residuals remain in LLGo #3097. Original PR head held. |
| Opcode-local 92d05239 | LLGo +2.78%9V74 despite fewer instructions | Spill eliminated; AMD7763 improves 1.04%. Dispatch/load scheduling and layout remain open. Not promoted as a universal recovery. |
| Numeric-family df2b0f87 | Large opcode-local Go gain disappears | Go feature-flag and stack reloads explain extra reads. One AMD event has 13/14 A/A and remains inconclusive. |
| Selected-index 0708894e | LLGo +20.96% Intel8573C | No first-operand spill; selected B index executes 1,249,156 CMOVs. Conditional-load ablation below confirms a substantial Intel loss from this source choice; exact cycle split remains open. |
| Shared popcount 410cee28 | LLGo +19.78% Intel/+1.38% AMD | Go feature-flag reload scope is corrected, but selected-index dependency remains. This source is not promoted. |
| Conditional numeric d1c31300 | LLGo +3.14% Intel6975P; Go 8.06% behind WasmEdge | Same-event conditional/index comparison lowers LLGo time 15.88%; source contrast resolves much of the selected-index loss, with residual regression still open. |
| Opcode-local + popcount 5f37d21a | LLGo +5.08% AMD9V74 | Go improves ~19%; Intel LLGo improves 0.69%. Host-dependent residual needs generated-code and cycle investigation. |
| Checked memory 688b40b6 | LLGo +4.02% Intel6975P; Go 0.52% behind WasmEdge | Go access helper inline cost falls 83→73 after equivalent unsigned range indexing. Baseline-relative LLGo loss is not automatically attributable to this last edit; a same-event predecessor comparison is required. |
| Immediate cache 3abcce7f | Go +3.30% and LLGo +0.86% vs 688b40b6 on Intel nonzero-offset mixed reads/writes | Go loop-self reads fall 1,040,960 but instructions rise 1,359,086, including 1,634,258 more NOPs. Stack reads/writes are unchanged. Rejected; LLGo codegen and residual cycle attribution remain open. |
| P3 caller-free scalar operation | GoCPU+1.2%/+1.7% despite 19 fewer allocations | Allocation count is not throughput. Extra ownership/path work and cycles remain open in the existing P3 investigation; not promoted. |

Timing sources: [scalar AMD](https://github.com/cpunion/wasmvm/actions/runs/43545680097),
[scalar Intel](https://github.com/cpunion/wasmvm/actions/runs/43547360024),
[compact](https://github.com/cpunion/wasmvm/actions/runs/43543803052),
[guarded key](https://github.com/cpunion/wasmvm/actions/runs/43549812034),
[opcode-local](https://github.com/cpunion/wasmvm/actions/runs/43552275491),
[numeric family](https://github.com/cpunion/wasmvm/actions/runs/43553539882),
[selected index](https://github.com/cpunion/wasmvm/actions/runs/43555187399),
[shared popcount](https://github.com/cpunion/wasmvm/actions/runs/43556904995),
[conditional numeric](https://github.com/cpunion/wasmvm/actions/runs/43558200486),
[same-event conditional/index contrast](https://github.com/cpunion/wasmvm/actions/runs/43559154198),
[opcode-local/popcount](https://github.com/cpunion/wasmvm/actions/runs/43560796867),
[checked memory](https://github.com/cpunion/wasmvm/actions/runs/43561639936).
Earlier adverse workload rows remain in their original research reports; this
campaign does not close or replace them.

## Compiler reproduction

LLGo [#3097](https://github.com/xgo-dev/llgo/issues/3097) records the hot spill,
[#3098](https://github.com/xgo-dev/llgo/issues/3098) the helper-inlining failure,
and [#3094](https://github.com/xgo-dev/llgo/issues/3094) the distinct layout issue.
LLGo main c0d07ce2 reproduces the spill and outlining on all three frozen source
variants, each with a checked guest result. This is a separate compiler event,
not current-main throughput acceptance or a production pin upgrade.

Extracting the full-LTO scalar loop and compiling it with
`llc -O3 -mcpu=x86-64 -relocation-model=static` reproduces the 7494-byte body,
696-byte stack reservation and critical spill. A bounded reduction produces a
274-line/9917-byte IR reproducer, pasted directly into
[#3097](https://github.com/xgo-dev/llgo/issues/3097#issuecomment-6871448283).
The reduced IR tests code generation only and must not be executed as guest
semantics. Omitting the CPU option changes register allocation; explicit
`generic` matches `x86-64`, and is not the same as omitting the option.

New main e6a87328 adds HostLayout/libffi tail-padding fixes. Its independent
[frozen-source verification](https://github.com/cpunion/wasmvm/actions/runs/43564954286)
builds and executes all three checked guests. The same loop sizes, frames,
critical spill and four outlined helpers persist; relocations differ. No
current-main throughput or production compiler upgrade is inferred.

## Rejected cache candidate and validation

Candidate 3abcce7f keeps opcode-local scalar reads, scoped popcount and checked
unsigned memory ranges. It caches scalar memory index/offset in the existing
32-byte immutable instruction record. Offsets at or above MaxUint32 retain
their original 64-bit source value. Native lowering still consumes canonical
metadata. No new instruction field or separately maintained execution loop is
introduced.

Complete Go and pinned LLGo execute suites pass; memory, slots, codegen and VM
checks also pass. All 31 executable core packages meet 90% statement coverage,
including slots 98.10%, execute 92.87%, memory 92.75% and VM 93.54%. Tests exercise
full-value copies, mixed-width popcount, address carry, logical capacity,
multi-memory, Memory64, wide offsets, fuel and trap-source provenance. The
checked-memory race suite passes. The final contribution head still requires
its own full dual-compiler/platform CI matrix.

[The exact candidate event](https://github.com/cpunion/wasmvm/actions/runs/43564772194)
finishes on two EPYC7763 hosts, each with 20 fixed blocks/520 observations,
25 alternate-input checks and 10/10 resolved A/A controls. Go head/base warm-time
ratios are 0.79521/0.79758; LLGo 0.98242/0.98283. Go medians are 80.14/80.31 ns per
source iteration, versus 135.63/135.41 for wazero interpreter and 244.68/244.60 for
WasmEdge interpreter. Preparation, first execution and warm execution remain
separate. These are hosted diagnostics, not controlled portfolio acceptance.

No Intel host occurred in that event. The predecessor's 0.52% Intel competitor
deficit is retained, rather than rounded to parity or assumed closed. The [nonzero-offset predecessor contrast](https://github.com/cpunion/wasmvm/actions/runs/43565591912)
finds Go +3.30% and LLGo +0.86% on Intel6975P-C, with 10/10 resolved A/A controls.
Go trails WasmEdge 2.79%. The AMD event has 9/10 resolved A/A; its adverse Go
point +4.33% remains inconclusive and retained. **The immediate cache is rejected.**

[Go count event 43566048911](https://github.com/cpunion/wasmvm/actions/runs/43566048911)
fails whole-window A/A by 1040 instructions; loop-self A/A is exact. Relative to
688b40b6, the loop reads 1,040,960 fewer operands but executes 1,359,086 additional
instructions. The mnemonic decomposition includes 1,634,258 extra NOP executions
and 595,182 extra XORs, offset by fewer moves/shift/OR/multiply operations. Two
NOP sites in complete local copies account for 1,216,392 extra instructions;
Go's x86 branch-padding logic inserts them around compare/conditional-jump
pairs. Direct stack reads/writes are unchanged. This identifies generated work,
not an exclusive elapsed-cycle explanation or an allocation/GC regression.
The whole-window failure and remaining helper-entry/LLGo cycle questions stay
open. A separate [dependent-load predecessor contrast](https://github.com/cpunion/wasmvm/actions/runs/43566440429)
confirms Go regressions 4.34% on 7763 and 7.37% on 9V45, with10/10 resolved A/A on
each host. LLGo improves 2.09%/4.24% on those hosts; its Intel regression is not
averaged away. The source is rejected even though it improves both compilers
against the older counter baseline.
Full-portfolio and P3 acceptance remain open.


## Private scalar read candidate

Source 1f659b69 starts from 688b40b6, excluding the rejected immediate cache. The
Go compiler cannot inline readAccess (cost 125 before shared-helper extraction,
105 after it; budget 80). The scalar path now keeps private fixed-width decoding
local, where Go inlines read and its endian loads. Shared memory retains one
native atomic adapter; vector paths use the existing common access wrapper.
This removes a known extra private-load function call without changing address,
overflow, logical-capacity, signed-extension or sharing contracts.

Complete memory/execute/VM suites pass with Go and pinned LLGo. New native
comparison 43566971853 is pending. The contribution's final head must pass its
own full dual-compiler/platform CI and per-package coverage gates before merge.
Every earlier adverse result remains part of the ledger; a faster new candidate
does not close an unexplained old regression.
