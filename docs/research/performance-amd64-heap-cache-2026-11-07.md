# Checked amd64 heap cache: 2026-11-07

Engine `d9f62f5` adds a bounded amd64 version of the existing checked per-entry
heap payload/descriptor cache. The measurement baseline is main `9b7cff0`.
Compiler, reference, ISA and release-profile pins stay unchanged.

## Shared implementation and lifetime

At most one key/address pair competes with ordinary scalar homes in the existing
costed allocator. Its fixed per-register priority is twice eligible loop access
frequency. Only the existing callee-saved GPR inventory is used. Bounded root
hints reject mixed/unknown references, loop calls, handlers and prompts; hints
affect profitability, never handle validity. Same-slot `local.get` is a read,
including after slot coalescing. No new execution framework or register bank is
introduced.

Every hit compares the full non-null handle. Misses retain index, tag and
generation checks. Array bounds and field-width behavior are unchanged. Keys
reset at every native entry, including cold continuation, poll/helper reentry
and linked-call return; addresses cannot survive host/GC relocation. Bounded
bump allocations preserve addresses within entry. Precise mode retains the
established checked path. ARM64 cache selection and emitted images are unchanged.

The original 22 x86 modules compile/link with feature mask 7. Only struct access
changes: **2282 → 2562 bytes**, adding the cache's entry reset, hit branches and
checked miss paths. The 12.3% code increase retains full miss/trap handling; it
must be assessed against complete API timing rather than instruction count.
Indirect struct access keeps its existing path because its loaded references
are not stable roots under these conservative hints. Allocation and other
21 module images remain byte-identical to the baseline.

## Qualification boundaries

Complete local codegen/nativeexec/public-VM tests pass with Go 1.27.3 and release
LLGo `3571c25 -O3 -lto=full`; vet and all 88 benchmark-tool tests pass.
Exact package coverage is **codegen 97.73% (8616/8816)** and
**nativeexec 92.54% (459/496)**. These are statement counts, not specification
coverage. Integration additionally requires final-head four-profile semantic
CI, all package gates and resolved review findings.

Tests cover scalar, packed, reference and vector struct/array accesses, signed
loads, full Lo/Hi frames, all cold PCs, changed references, warmed invalid handles,
moved payloads across reentry, forced ordinary backedge polling, original trap
sources, fuel/budget and precise fallback. Both targets compile; the ARM64
counterpart executes locally. Actual x86 execution belongs to Linux CI and the
source-pinned hosted event, not the local compile-only diagnostic.

All 22 ARM64 published module images and entry lists match baseline/head and
Go/LLGo. Forty-four fresh candidate checksum witnesses pass; the 44 exact-main
witnesses are reused after source, compiler/profile, executable and fixture hash
validation. No extra local warm campaign is collected for unchanged guest code.

One development test initially assumed unmetered polling consumed Budget and
timed out; it now deterministically forces the verified Control poll. The
failed observation is retained and is not classified as a production bug.

## Independent timing cohorts

[Event 42376300827](https://github.com/cpunion/wasmvm/actions/runs/42376300827)
compares `d9f62f5` with `9b7cff0`, retaining the original 22-case portfolio and
all preparation/first/warm/shutdown/process intervals. A separate `gc-sustained-v1`
cohort uses fixed **300,000 allocation / 1,000,000 other GC iterations**, independent
checksums and its own four A/A controls. It runs in the existing GC job with the
same already-built executables. Counts, samples and source events are never pooled.
The PR body links the complete event and records its final observations; hosted
results do not close controlled performance acceptance.

The completed event passes all six jobs. Original22 retains 362 alternate
checks and 12 explicit reference-GC rejections; the separate four-case long
cohort retains 56 checks and its own 12 rejections. All four controls resolve
8/22 original and 2/4 sustained cases. Build takes 371 seconds; the longest
measurement group takes 251 seconds. Ratios below use full warm-call medians.

| Workload/cohort | Go head/base | LLGo head/base | LLGo / Go native | LLGo / Wasmtime | All four A/A |
| --- | ---: | ---: | ---: | ---: | --- |
| Struct, original count | 0.771 | 0.736 | 1.003 | 1.618 | Inconclusive |
| Struct, 1,000,000 | 0.885 | 0.950 | 0.995 | 1.256 | Inconclusive |
| Allocation, 300,000 | 1.002 | 0.998 | 1.156 | 2.858 | Inconclusive |
| Indirect struct, 1,000,000 | 0.996 | 0.997 | 0.998 | 2.744 | Resolved, advisory |

Allocation rows above use warm medians. The [throughput correction](performance-gc-cache-throughput-2026-11-08.md)
re-reduces the same samples with every observed pause included per process;
2.858 is a latency diagnostic rather than a sustained allocation deficit.

The changed workload observes lower elapsed time in both cohorts, but unresolved
controls do not establish a reliable speedup. The long candidate native control
resolves; its baseline native control does not. Indirect access retains a material competitor deficit. wazero/Wasmer reject these GC fixtures;
there is no equivalent native C GC row. Other 21 x86 images are unchanged; all
original runtime/interpreter/competitor/lifecycle observations remain in the
complete event. No cross-event extrapolation is used.

## Corrected CI execution cost

The first PR head `51c9a47` passes 20 required jobs, but its macOS legacy job
exceeds the 12-minute job limit. It serially links/runs P1 and legacy-thread
packages, and recompiles the large LLGo guest for both native policies. The
unchanged guest assertions executed so far pass; cancellation occurs during
the second compilation, not a reported semantic failure. The incomplete trace
and reports remain retained as a failed CI attempt, never counted as acceptance.

P1 and threads now have separate LLGo shards. All 31 test packages remain once
per target across eight shards; the required matrix expands to 23 jobs. The
large threaded guest compiles once per backend; its native artifact already
contains both policy tiers. Every mode still receives independent process,
Store, memory and thread state. Interpreter, unmetered native and fuel executions
remain separate, including their three original reports. Compiler/profile and
job limits are unchanged. The final PR checks/audit supply actual completion
and duration evidence; the incomplete first attempt cannot authorize merging.

The preceding source `cd0dd65` treated same-slot reads as mutations and enabled
no cache in any benchmark module. Its complete
[event 42374722485](https://github.com/cpunion/wasmvm/actions/runs/42374722485)
is retained separately: 22 original cases/362 alternate checks/12 expected GC
rejections, plus four long GC cases/56 checks/12 expected rejections. Only 4/22
original cases and 0/4 long cases resolve all controls. Its apparent long struct
ratio of 0.797 occurs with identical guest code and unresolved controls; it is
not evidence of a cache benefit. A direct regression fixes the analysis defect
before collecting the new source event. No favorable retry or count/profile
sweep is used.

The [machine-readable record](data/amd64-heap-cache-2026-11-07.json) retains local
qualification, all module-change summaries, publication comparisons and the
independent inactive-hint results. Immutable local archive:
`artifacts/amd64-heap-cache-20261107/local-evidence.tar.gz`, 6506203 bytes,
661 files; SHA-256 `ff374513a7432868aa9f5c868f272d17a89d53ac4f69ff4750f0769d4ed7d123`.
