# Canonical invocation storage and bound value plans

The first implementation of the [P3 mechanism direction](performance-p3-mechanisms-2026-11-16.md)
removes recurring private storage allocations and root layout lookups. Complete
P3 requests allocate **29 fewer Go allocations / 34 fewer LLGo allocations**. Go executes
about **1.2–1.6% fewer instructions** in the final source comparison. Neither
complete-request CPU improvement nor competitive performance acceptance is
established. Native host transitions, typed WASI bindings and ready/pending
async work remain separate implementation steps.

## Implementation and lifetime contract

`canonical.ValuePlan` binds immutable root layouts to a particular layout
builder. Binding stays lazy: unused types consume no layout budget, and
cancellation/value quotas retain their original precedence over validation.
The original recursive conversion engine still validates and copies nested
values, charges transfers, canonicalizes NaNs and checks resource ownership.
These are root plans, not generated typed WIT converters.

Synchronous canonical lowers lease an instance-local `canonicalFrame` under
the existing Store lease. Temporary argument carriers, the scalar result
buffer, codec/resource transaction state and memory/realloc adapters belong
to that invocation. Reentry gets a distinct frame until the outer call ends.
Cleanup finishes before release, including error and panic paths. Release
clears all live values, resource/codec state, Caller, context and memory
capabilities. Parameter spill arrays larger than the direct-ABI limit are
discarded; shutdown drops the cache. Cached realloc methods retain only their
cleared frame and rebind the active owners at the next lease.

Host calls embed their own codec and, when needed, resource state in one
allocation. Public callback arguments/results remain independently owned;
public Callers, copied Caller headers, References and BorrowScopes retain
their existing lifetime. Suspending calls retain their task-owned storage.
No public capability is revived by frame reuse.

## Sources and experiments

Source base **8c8d4a759ecdf6c826e4df6078a98ee376dab9e0**, first implementation
**b4bcc4523fc4ca6799c0a790f6b806036ccd438a**, final implementation
**121dff14e19f3d43cc739da49ec2d5367a7372f8**; Apple M4 Max, macOS arm64;
Go **1.27.3**, LLGo **5860fb6f9c13371f6cdb456a987a8e95844066d4** with
**`-O3 -lto=full`**. Reused baseline binaries have Go/module sources verified
equal to the base. Compiler profiles and normal collector settings are fixed.
Independent references retain Wasmtime **48.0.1** and their pinned Rust provider.

The first implementation has a fixed ten-block Component scalar-import cohort
(120 processes) and four separate ten-block complete-P3 cohorts (280 processes).
Each role has symmetric same-binary A/A controls. Micro runs use 100,000 imports,
ten declared seeds, two warmups and seven interleaved steady calls. P3 uses
four Stores, sixteen clients, twenty serialized/interleaved batches of 512
requests and a flush batch. Full response status, headers, trailers, body and
CRC are checked; guest joining remains in the measured driver.

The final implementation has **176 new processes**: separate ten-block Go
cohorts at both payload sizes (40 processes each), a six-block LLGo 16-byte
cohort (24), and a six-block LLGo 64-KiB source/GC diagnostic (72). LLGo
diagnostics use ten measured batches of 128 requests and a flush batch to bound
the disabled-GC heap. The last diagnostic compares base, first implementation
and final implementation, each with normal GC and `GC_DONT_GC=1`. Every
disabled-GC measured window reports zero collections. It is a causal probe,
not a deployable performance configuration or competitor acceptance.

No experiment is extended or pooled. Startup, load, preparation, first call,
warmup and cleanup are separate. Complete-driver CPU/instruction counters are
not exclusive guest or component costs. External load snapshots are retained;
no owned build, test, profile, edit or analysis overlaps timed cohorts.
Ratios use paired-block geometric means and 20,000 bootstrap block resamples.
The full 95% A/A interval must fit [0.975, 1.025].

## Mechanism and allocation results

Counter-only overlays on the first implementation, twice under each compiler
and payload, retain the same guest entries, validation visits and clone nodes:

| Per request | 16 bytes | 64 KiB |
| --- | ---: | ---: |
| Native entries / host exits | 46 / 38 | 58 / 50 |
| Canonical value visits | 85 | 97 |
| Recursive clone nodes | 47 | 59 |
| Layout Build calls, base → first implementation | 116 → 66 | 152 → 102 |
| Private frame acquire / release / warm creation | 7 / 7 / 0 | 7 / 7 / 0 |

The final clearing change does not change this control flow; these overlay
counts are explicitly attributed to the first implementation, not relabeled
as a final-head counter run. Logical codec sessions still initialize budgets
and ownership. Removing a codec constructor allocation does not remove a
session or a semantic check.

Final uninstrumented complete-P3 runs report:

| Compiler | Payload | Allocation events/request, base → final | Accounted bytes/request saved |
| --- | --- | ---: | ---: |
| Go | 16 bytes | 645 → 616 | approximately 2,816 |
| Go | 64 KiB | 696 → 667 | approximately 2,816 |
| LLGo | 16 bytes | 962 → 928 | approximately 2,976 |
| LLGo | 64 KiB | 1,133 → 1,099 | approximately 2,976 |

Go and LLGo managed counters cover different allocation domains and size
accounting. They do not rank total Go, Rust, C or JIT heap use. The first
implementation's scalar-import slopes fall from 9.25 to 5.25 Go events/import
and 15.25 to 11.25 LLGo events/import. Public argument snapshots and capabilities
remain allocated; the remaining boundary costs are substantial.

## Timing, instructions and adverse evidence

Final complete-request ratios below are **final/base**. Lower is better for
CPU and instructions; higher is better for throughput.

| Compiler / payload | CPU ratio, 95% interval | Instructions ratio, 95% interval | Matching controls |
| --- | --- | --- | --- |
| Go / 16 bytes | 0.999 [0.992, 1.007] | 0.9845 [0.9827, 0.9865] | Both pass |
| Go / 64 KiB | 1.002 [0.995, 1.011] | 0.9879 [0.9864, 0.9894] | CPU fails; instructions pass |
| LLGo / 16 bytes | 1.024 [1.000, 1.051] | 0.996 [0.981, 1.016] | Both fail |
| LLGo / 64 KiB | 0.979 [0.913, 1.058] | 0.961 [0.903, 1.029] | Both fail |

The final throughput point ratios are 0.987/0.984 for Go 16-byte/64-KiB and
0.989/1.001 for LLGo. **All matching throughput controls fail.** These adverse
points remain evidence; neither exact regressions nor exact improvements are
accepted from them. Go's cycles/instruction ratios rise about 1.4% with passing
controls while instruction counts fall. This derived quotient alone does not
identify a cache or instruction-scheduling defect, and CPU improvement remains
unresolved.

The first implementation's LLGo 64-KiB CPU point ratio was **1.031** with
failed controls. Inspection found two concrete redundancies: release cleared
the entire argument capacity, and acquisition zeroed resource state already
cleared by release. The final code clears the initialized argument prefix,
including a successful prefix before later lift failure, and only rebinds
active resource owners. ARM64 disassembly confirms that the memset length
loads the slice length at offset 0x108 rather than capacity at 0x110 and that
the acquisition path no longer repeats the resource-state zero stores.

The separate disabled-GC comparison still has **1.0050 instructions**
[1.0043, 1.0059] final/base, with passing instruction controls. CPU is **1.031**
[1.021, 1.040], with failed CPU controls. Final/first-implementation disabled-GC
instructions are **1.0005** [0.9986, 1.0020]. Therefore the clearing change
cannot be credited with resolving the earlier time observation. The new
frame/plan bookkeeping has a small residual instruction cost in this LLGo
probe despite fewer allocation requests; its exact function-level attribution
remains open. Default-GC instruction controls fail, whereas disabled-GC
instruction controls pass. This implicates collector work in variability but
does not explain every remaining cycle. Peak physical footprint grows from
roughly 169–187 MB with normal GC to 574–580 MB without it. Existing LLGo
collector issue #3106 remains relevant; this evidence does not establish a
new compiler bug or justify a duplicate issue.

For competitive context only, the first implementation's micro medians are
**487 ns/import Go, 910 ns LLGo, 70 ns Wasmtime typed and 115 ns dynamic**.
Their time controls fail. The first complete-P3 cohorts observe about **2.0–2.1x
Wasmtime CPU for Go and 5.8–6.0x for LLGo**, with unresolved reference/source
controls as recorded in the raw analyses. These are dated diagnostic points,
not final-source competitive ratios, portfolio results or achieved parity.
Broad competitive acceptance remains deferred until the ordered changes land.

## Why host compilers can differ in native mode

The owned Wasm compiler emits guest machine code. With identical input,
architecture, feature policy and compiler options, Go/LLGo builds can produce
identical guest bytes; the preceding study verifies four fixtures. This does
not establish byte equality for every module or equivalence with Wasmtime's
different optimizer and ABI. Go and LLGo separately compile the VM, Canonical
ABI, WASI provider, allocator/GC and scheduler. The current Go native entry uses
`runtime.cgocall`; LLGo enters through its native C ABI. Guest execution that
frequently leaves for host work consequently runs different host machine code
and runtime mechanisms even when its guest image is identical. The common
host materialization/transition gap remains an architecture task.

## Validation and remaining work

Go and LLGo execute canonical/component, P2/P3 and nested P3 harness tests.
Go race checks cover retained capabilities, frame reuse, reentry and host
failures; vet passes. Combined affected-package statement coverage is
**94.35% canonical (1,552/1,645)** and **91.89% component (3,445/3,749)**.
Tests cover conversion/budget parity, builder rebinding, Unicode/NaN/invalid
representations, retained arguments/results/Callers, stale references/borrows,
panic/cancellation, partial argument failure, memory growth/reallocator reentry,
large spilled arguments and shutdown. Root-module attempts to address the
nested P3 module failed during setup and are retained; separate correct-module
Go/LLGo execution passes. Final target CI/review remains the integration gate.

Next: shorten native host exit/resumption without weakening code/memory/root
leases or public Caller lifetimes; bind private typed WASI/scoped memory edges;
separate immediately-ready async work from pending ownership; finish the
ordered memory/SIMD/SHA work and then the unified competitor portfolio. None
of these remaining deliverables is closed by the allocation reductions here.

Evidence: [summary and all controls](data/canonical-invocation-frames-2026-11-16.json),
[compact process/round records](data/canonical-invocation-frames-2026-11-16-raw.jsonl.gz),
[frozen plans, scripts, counter diff, disassembly and test logs](data/canonical-invocation-frames-2026-11-16-protocol.tar.gz).
Per-request latency arrays are omitted with counts and checksums; no latency-tail
claim uses this compact export. Original compressed rows remain locally with
their hashes in the summary. No additional source checkout was created.
