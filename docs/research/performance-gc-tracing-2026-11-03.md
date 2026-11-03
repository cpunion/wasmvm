# Shared GC trace work and remaining allocation gaps: 2026-11-03

Status: primary six-case diagnostics, complete original twenty-case comparison
and affected semantics verified. Additional native duration controls are verified;
new final-head CI remains required before integration. Performance acceptance
remains open.

## Evidence and implementation

[PR112 final CI and controls](pr112-final-ci-2026-11-03.md) establish the merged
baseline, including a sustained native allocation deficit about 6.90x Wasmtime.
Five-second sampling points to shared collection, compaction and batch metadata
commit. Profiled durations do not enter any performance comparison.

Three source-changing experiments retain every original outcome. A pending-object
worklist bounds chain traversal scratch space but the first slice-stack version
regresses native LLGo allocation 8.9%. An explicit cursor reuses slots and removes
that regression, yet its approximately 1% improvement does not resolve the gap.
The final candidate additionally caches reference byte positions and value kinds
once per immutable heap layout. Collection visits only these reference positions;
arrays share their element classification and exception arguments use the same
layout. Function owners, custom descriptors, continuation captures and instance
edges keep their existing roots. No unsafe payload reads, guest format, quota,
compiler/profile, feature or safepoint change.

The cache adds per-type metadata, not a new per-object header. Type-heavy peak
memory and preparation acceptance still need measurement. Runtime accesses retain
bounds checks and full-width array offsets. The collector stack clears popped
slots and retains only pending objects; a 4096-node shared cycle uses bounded
scratch space and is reclaimed after losing its roots.

## Three separate duration campaigns

Baseline `12599967425d1085039ee00aafd6b31dd049284b` equals merged PR112 engine
source `84928483664d14fd33909da24e57432e96b6d262`. The candidates are `aa28a14`,
`b2d904a` and `354abd3`. Fixed Go 1.27.3/LLGo `983465e` with the production
full-LTO profile; Apple M4 Max/macOS arm64. Each campaign independently compares
its own fresh interleaved baseline, head and exact baseline A/A aliases. They
are not pooled or used to manufacture a cross-campaign improvement.

Each scope retains six backend/cases, 220 observations, 44 alternate checks,
50 independent affine-oracle checks and 44 summaries. Five fresh process blocks,
two warmups and seven measured calls. Allocation uses 300k objects on both
backends; all configurations within a case share counts and guest bytes. The
fixed 64 MiB quota is unchanged: packed payload is 12 bytes plus 64 accounting
bytes/object, or 22.8 MB/45.6 MB for one/old+new chains.

Native allocation, ns/object:

| Source experiment | Go base → head | LLGo base → head |
| --- | ---: | ---: |
| LIFO slice | 55.461 → 55.751 (+0.5%) | 61.841 → 67.335 (+8.9%) |
| Pending cursor | 56.070 → 55.728 (-0.6%) | 61.772 → 61.138 (-1.0%) |
| Typed trace layout | 55.537 → 46.495 (-16.3%) | 61.614 → 56.505 (-8.3%) |

The typed-layout source observes Go/LLGo native allocation improvements of
16.3%/8.3%; native A/A differences are about -0.12%/-0.16%. Wasmtime is 9.000
ns/object versus LLGo 56.505: about 6.28x remains. Interpreter allocation is
121.404→111.531 ns for Go and 142.888→139.659 for LLGo, versus WasmEdge 103.977.
LLGo interpreter A/A +2.25% limits interpretation of its smaller gain.

Retain typed-layout Go direct-GC interpretation +0.65%, Go/LLGo native
wait-mismatch +1.17%/+0.68%, and every short/adverse outcome from the earlier
complete twenty-case experiment. Longer counts do not replace original results.
All preparation, first/warm calls, startup and cleanup remain separate; no
clock, boundary or process median is subtracted. Ordinary C has no claimed
GC equivalent in this scope; nine existing C references remain independent.

## Complete original workload comparison

The final typed-layout source also completes the original twenty-case matrix:
864 VM/reference observations, 27 ordinary-C observations, 288 alternate checks,
twelve explicit unsupported GC outcomes and 297 independently regenerated
summaries. Counts, fixtures, policies and exact source/toolchain inventories stay
unchanged. This separate campaign does not pool samples with the duration study.

Short allocation observes native Go/LLGo -9.5%/-9.9%, Go interpretation -6.9%
and LLGo interpretation +3.1%. Retain LLGo short i31 interpretation/native
+5.7%/+5.7% and other adverse rows in the complete evidence. Non-GC sustained
interpreter/native controls remain recorded case by case, including Go native
memory +2.6%, LLGo float-memory interpretation +1.7% and Go function-call
interpretation +1.1%.

Retain Go empty native +20.8%, direct GC native +13.4%, notify native +12.1%,
wait-mismatch native +15.8% and wait-timeout native +37.8%. Longer direct-GC and
wait-mismatch controls are separate observations, not replacements. Two separately
declared five-million-iteration native notify/timeout controls retain all three
compiled references and exact baseline A/A. They verify 90 observations, 18
alternate checks and 18 summaries. Go notify/timeout changes +0.51%/+0.48%;
LLGo changes +1.08%/+1.66%, with LLGo A/A +0.83%/-1.85%. These 4.06–4.59 ms
own process medians preserve the original tiny-loop data as separate observations.
Single-call and tiny-loop percentages do not establish precise instruction costs.

## Competitor mechanism audit

Pinned Wasmtime 48.0.1 source
[`7bac2c2`](https://github.com/bytecodealliance/wasmtime/tree/7bac2c2775808aaec5d4aa5627a5e447b51102cf)
prefers its copying collector when compiled in. The exact measured SDK includes
defined copying-compiler/layout symbols, while the adapter retains the default
collector. Its [allocation code](https://github.com/bytecodealliance/wasmtime/blob/7bac2c2775808aaec5d4aa5627a5e447b51102cf/crates/cranelift/src/func_environ/gc/copying.rs#L100)
emits checked inline bump allocation and only calls the allocation helper when
that region is exhausted. Small object trace positions are represented by an
[inline bitmap](https://github.com/bytecodealliance/wasmtime/blob/7bac2c2775808aaec5d4aa5627a5e447b51102cf/crates/environ/src/gc/copying.rs#L43).
Its compact offsets and copying layout differ from our 64-bit generational
handles and host metadata. These are structural leads, not a cycle attribution
or proof that one line explains the measured gap. The source audit and SDK hash
remain separate from measurements; no competitor or production compiler is upgraded.

## Validation and remaining acceptance

Go and production full-LTO LLGo each pass complete state/SIMD/execute/VM suites
and all 48 expected execution reports. All five mandatory extensions execute
three tiers, including forced Stack Switching GC. New semantic tests cover a
deep shared cycle, external references, numeric pointer-shaped bits, packed
unaligned fields, arrays, retained function owners and complete reclamation.
This is affected-package validation, not the full 31-package final CI inventory.

New exact-head dual-platform CI, thirty coverage gates and review remain required.
Native GC still trails Wasmtime;
public-boundary, host-compiler, ordinary-C, startup/allocation/cancellation,
expanded WASI-world and controlled portfolio acceptance remain open. No new
LLGo defect is inferred from throughput; #2983 stays with its other contributor.

[Complete original observations, compiler/source inventories, semantic reports,
both earlier experiments, duration controls and primary-source audit identities](evidence/performance-gc-tracing-2026-11-03.json.gz).
