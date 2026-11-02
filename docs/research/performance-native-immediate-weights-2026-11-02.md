# Native immediate register costs: 2026-11-02

Status: candidate, not integrated or performance-accepted. The compiler changes
its own cost model; the finding is not an LLGo defect.

The x86 allocator previously treated literal writes/right-operand reads as hot
register work even when existing lowering encodes the right operand directly.
In the scalar memory loop it caches that constant's slot while the running value
incurs a frame store and two reloads per source iteration. The new bounded scan
discounts only the excess hot credit for encoded operands and proven-dead
literal writes, retaining one cold credit. It reuses existing immediate and
read-before-overwrite proofs. Source PCs, ordinary emission, traps, polling,
cold-entry state, code budgets and ARM64/metered allocation remain unchanged.
No fixture names or constants influence allocation.

Base `0063067b1c128fa2c165734049e2a0b166ce4c5a`; engine candidate
`e9fff30093a1da5413e11b8a89f48f800b897e29`.

## Allocation and semantics

The representative x86 memory body shrinks 3629→3554 bytes. Running slot 5
receives a register; the store/two-reload chain disappears. The canonical
constant write remains in its frame home. This is a code-generation observation,
not an attributed performance saving.

Eighty-six function pairs are captured over both architectures and policy tiers,
including optional expansion. Exactly seven unmetered x86 bodies change:
integer, scalar memory, narrow memory, dependent address, i31, struct access and
struct allocation. Dependent address grows 2904→2937 bytes; other changed bodies
retain their size. Every ARM64 and metered body is byte-identical. Foreign code
is compile-only and is never published or timed.

All 31 Go packages and 30 coverage gates pass; codegen coverage is 97.91%.
Pinned LLGo codegen/execute/VM, scoped race, no-cgo and vet pass. New tests cover
cold credits, live/dead values, trap/guard/region/poll boundaries, operand aliases,
unsupported I64 immediates, shifts, architecture/tier controls and a typed loop
with independent coefficients. They compare original PCs, scalar slots, visible
memory and exit state against ordinary code.

Go 1.27.3 codegen tests also pass in Linux amd64 Docker emulation. The first
invocation failed two corpus tests because its read-only fixture mount was
missing; the corrected complete invocation passes in 7.530s. Both logs remain.
Emulation proves only execution semantics, never Linux throughput. Linux LLGo
and final contribution-head dual-platform validation remain required.

## Short Linux preflight

Workflow [41436534296](https://github.com/cpunion/wasmvm/actions/runs/41436534296)
passes. Artifact 13136776994 SHA-256
`6cde1c0cbf31e769a6c2d7ac9eb9886b9f8bc24e658b1e8942d96486277e1e69`
validates; all 228 observations / 76 summaries / 19 configurations and exact
base/head git-archive source inventories match. Go 1.27.3 and pinned LLGo
`8e45ed5` use identical base/head flags (`-O3 -lto=full` for LLGo).

Units are ns/source iteration; public native API, three fresh process blocks,
two warmups and seven calls. All preparation, first/warm, readiness and cleanup
phases remain in the evidence. This PR scope has no ordinary C algorithm.

| Workload | Go native base → head | LLGo native base → head | wazero native | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Integer | 1.9288 → 1.8648 | 1.8906 → 1.9070 | 1.9708 | 1.8846 | 1.8542 |
| Guest calls | 1.0841 → 1.0421 | 0.9737 → 0.9661 | 2.7855 | 1.8583 | 1.9092 |
| Memory | 1.5879 → 1.6113 | 1.6056 → 1.4257 | 2.0695 | 1.8315 | 0.9830 |

LLGo memory improves 11.2% but Go regresses 1.5%; LLGo integer regresses 0.9%.
The new memory observation remains 1.45x Wasmer. Short shared-host samples do
not resolve a stable compiler-dependent win, despite identical generated body
selection. Interpreter controls and all regressions remain retained. Do not pool
this runner with earlier Xeon or local ARM64 campaigns.

A separate longer Linux diagnostic covers all seven changed functions plus
call/float controls, baseline A/A aliases, ordinary LLVM-22 C for applicable
scalar cases and the three native references. Its evidence-only branch changes
no engine source. The initial run failed when standalone C output was passed
to the VM readiness parser; both the failure and corrected-protocol source are
retained. Nine kernels now have scoped observations across the retained eight-case run
and separate bounded allocation follow-up below; controlled acceptance is open. Neither this
diagnostic workflow nor its one-off runner is intended for integration.

[Lossless evidence](evidence/performance-native-immediate-weights-2026-11-02.json.gz)
retains all observed phases, original identities, affected machine-code captures,
validation logs, missing-mount failure and first diagnostic protocol failure.


The separate [ordinary-C mechanism study](performance-native-c-memory-mechanism-2026-11-02.md)
uses ARM64 engine `93190e9`, not this x86 candidate. C vectorization-disabled
assembly is unchanged; induction/unroll scheduling remains an investigation.
Do not combine its C rate with the Linux preflight above.


## Longer Linux observations (eight complete cases)

[Run 41439468275](https://github.com/cpunion/wasmvm/actions/runs/41439468275)
is **failed**, not a successful nine-case run. Six scalar workloads, i31 and
struct access complete all three blocks. Struct allocation then encounters an
explicit Wasmer GC-disabled validation message outside the earlier classifier's
accepted strings. That rejection occurs before its timed block. Its raw failure
and configured-reference limitation remain visible. An initially misselected i31
follow-up is cancelled; only struct allocation is completed in the separate follow-up below.

AMD EPYC 7763/Linux amd64; every process uses the same affinity CPU. Engine
sources remain exact `0063067`→`e9fff30`. Go 1.27.3, pinned LLGo O3/full LTO,
LLVM-22 ordinary C and pinned references are independently identified. Native
memory uses 50 million iterations, other scalar kernels five million, GC one
million. Three fresh process blocks, seven serially interleaved calls, two
warmups, exact baseline Go/LLGo A/A aliases and alternate oracles are retained.
No interpreter or CLI campaign is repeated. C has its declared safe float cap;
no C GC implementation is invented.

All eight completed cases independently validate: 222 observations, 74 summaries
and 74 alternate-input checks. Base/head source inventories match exact archives;
the artifact digest is verified. These data are usable scoped observations from
an incomplete workflow, not a complete-run or controlled acceptance verdict.
Units are ns/source iteration:

| Workload | Go base → head | LLGo base → head | Ordinary C | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| core/integer-loop | 1.8871 → 1.8867 | 1.8787 → 1.8801 | 1.8743 | 1.8737 | 1.8750 |
| core/function-calls | 0.9559 → 0.9560 | 0.9513 → 0.9476 | 1.1732 | 1.8765 | 1.9201 |
| core/memory-load-store | 1.5373 → 1.3658 | 1.5357 → 1.3564 | 0.7381 | 0.9576 | 0.9576 |
| memory/narrow-unaligned | 2.5007 → 2.5290 | 2.4725 → 2.4755 | 1.0944 | 1.6930 | 1.6924 |
| memory/dependent-address | 2.9884 → 2.9146 | 2.9827 → 2.9080 | 1.8719 | 2.1831 | 2.1901 |
| memory/float64-unaligned | 1.2668 → 1.2430 | 1.2325 → 1.2346 | 0.9400 | 0.9479 | 0.9437 |
| gc/i31-roundtrip | 2.2245 → 2.2304 | 2.2207 → 2.2277 | — | 2.1902 | — |
| gc/struct-access | 1.6064 → 1.6176 | 1.7670 → 1.6349 | — | 1.2624 | — |

Memory improves Go 11.2% and LLGo 11.7%. Every paired block improves: Go
head/base 0.8830–0.8908, LLGo 0.8832–0.8876. Corresponding memory A/A ratios
range 1.0014–1.0035 and 0.9996–1.0050, respectively. They still do not establish
the declared controlled calibration. LLGo memory remains 1.42x Wasmtime and
1.84x C; narrow memory remains 2.26x C and dependent address 1.55x C.

Retained regressions include Go narrow memory 1.1%, LLGo narrow 0.1%, i31
about 0.3% for both, and Go struct access 0.7%. LLGo struct access appears
7.5% faster, but its baseline A/A is also much faster than baseline; that is
not a reliable attributed win. Float controls and all phases stay in the raw
evidence. Go/LLGo are close on most native kernels here; these results do not
independently establish another LLGo defect.

The same-campaign C rows address the ordinary-native gap; the short PR preflight
above has only C API adapters. Never mix either Linux campaign with ARM64's C
rates. Struct allocation is recorded separately below; contribution-head semantics and
controlled whole-portfolio acceptance remain pending.


## Remaining allocation case

[Run 41442475955](https://github.com/cpunion/wasmvm/actions/runs/41442475955)
passes the remaining struct-allocation workload. This separate Linux host is
AMD EPYC 7763 64-Core Processor. Seven configurations each complete three blocks (21 observations),
seven alternate checks and two explicitly unsupported configured references.
Artifact SHA-256 and both exact source inventories validate. Compiler/profile
and engine revisions remain unchanged from the eight-case study.

The initial one-million iteration request correctly exhausts the unchanged
64 MiB guest heap: every newly allocated object retains its predecessor. Local
Go and LLGo independently reproduce that bound and execute the original 10,000
count. The final Linux comparison preserves those original manifest iterations
for every runtime, rather than weakening quotas. The failed oversized run,
classifier failure and cancelled misselected follow-up remain retained.

| Configuration | ns/allocation-loop iteration |
| --- | ---: |
| aa-go-native | 100.8302 |
| aa-llgo-native | 183.1186 |
| base-go-native | 101.8671 |
| base-llgo-native | 189.6228 |
| head-go-native | 102.2630 |
| head-llgo-native | 184.6274 |
| wasmtime-cranelift | 250.8677 |

Go head regresses 0.4%; LLGo appears 2.6% faster, while its baseline A/A is also
3.4% faster than baseline. This does not attribute an allocation gain to the
cost model. LLGo is about 1.81x Go but faster than this configured Wasmtime
reference; the host-compiler allocation/collection gap remains a priority.
Existing independent LLGo reports do not explain that whole difference, and no
new issue is filed from throughput alone. There is no ordinary C GC control.

All nine planned kernels now have complete scoped measurements across the
explicitly failed eight-case run and this successful one-case run. They are
separate campaigns/hosts, not one successful nine-case workflow or a pooled
controlled acceptance result. The consistent scalar-memory improvement supports
continuing the compact candidate; final contribution-head validation, narrower
regressions, GC compiler gap and controlled acceptance remain open.
