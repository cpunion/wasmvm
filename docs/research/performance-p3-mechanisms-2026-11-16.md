# P3 mechanisms: host boundaries before collector tuning

The common performance problem is reproducible **without HTTP, threads, or
asynchronous I/O**. Repeated scalar imports expose a large host-boundary gap
under both Go and LLGo. A guest-only arithmetic control does not reproduce that
gap. The evidence supports redesigning internal host/canonical transitions;
it does not establish an unavoidable Go-language ceiling or justify replacing
the owned compiler. This is an investigation, not a performance acceptance event.

## Sources and measurement boundaries

Engine source **a096c7b0d53100fbdbe1181c6e0a0ef4682cb94d**, Go **1.27.3**,
LLGo **5860fb6f** with **`-O3 -lto=full`**, LLVM **22.1.8**, Rust **1.98.0**,
Wasmtime **48.0.1**, wazero **1.12.0**; Apple M4 Max, macOS arm64. LLGo main
was checked again and still matched this pin. Existing checkouts and dependency
caches were reused. Production sources and runtime/collector defaults were not
changed. Neither diagnostic overlays nor the allocation-instrumented Rust
runner participate in the timing cohorts.

Four small fixtures calculate `x = 17*x + 5` with wrapping u32 arithmetic,
using runtime iteration counts and seeds. Loop controls execute the recurrence
in Wasm; import cases execute it in a scalar host function. Component fixtures
add a canonical lower/lift around the same Core module. Wasmtime additionally
binds the same import through its dynamic API. Every invocation checks the
result and exact callback count. Load, preparation, first call, steady calls
and cleanup are recorded separately. A steady iteration includes guest loop
work, the applicable boundary, and amortized top-level API work; it is **not
exclusive callback latency**. No control time is subtracted.

The first cohort has six fixed blocks per fixture, two same-binary A/A
processes per participant, two warmups and seven individually interleaved
steady calls per process: **156 processes**. The separate wazero/Core cohort
has six blocks and **96 processes**. Its Go/LLGo/Wasmtime anchors are new
processes, not pooled with the first cohort. Loop runs use 5,000,000 iterations;
import runs use 100,000. Six seeds include high-bit and all-one inputs.
Inference uses paired block geometric ratios and 20,000 bootstrap block
resamples. A/A confidence intervals must fit [0.975, 1.025]. External load was
recorded; no owned build, test, profile, edit or analysis overlapped timing.

These are scoped architectural diagnostics. Policies differ: wazero and
Wasmtime retain their default disabled fuel/epoch/termination instrumentation;
wasmvm retains its normal execution/capability checks. There is no common P3
implementation in wazero or Wasmer in this harness. Neither this small
arithmetic control nor a synchronous Component import represents all Core or
complete P3 workloads. Preparation is an exposed aggregate, not a fabricated
decode/compile/instantiate split.

## Where the gap appears

Times below are medians of paired process medians, normalized per iteration.
The Core rows use the independent second cohort; Component rows use the first.
**Except where stated, time controls fail: magnitudes are observations, not
accepted precise speed ratios.** Raw intervals and each A/A result are retained.

| Fixture | wasmvm / Go | wasmvm / LLGo | Wasmtime typed | wazero / Go |
| --- | ---: | ---: | ---: | ---: |
| Core arithmetic loop | 0.86 ns | 0.85 ns | 1.13 ns | 1.21 ns |
| Core scalar import loop | 146 ns | 217 ns | 3.97 ns | 20.5 ns |
| Component arithmetic loop | 1.07 ns | 1.08 ns | 1.43 ns | Not implemented |
| Component scalar import loop | 652 ns | 1,239 ns | 89.2 ns | Not implemented |

Wasmtime's dynamic Component import is **152 ns/iteration**, versus 89.2 ns
typed. Its paired point ratio is 1.63; controls fail, so this is not an accepted
exact dynamic-binding penalty. Go/Wasmtime arithmetic in the second cohort has
resolving controls (ratio **0.763**, 95% interval [0.755, 0.771]); this result is
limited to this dependency-chain kernel. LLGo/Go Core imports also resolve:
**1.479** [1.465, 1.495]. Other precise competitive timing claims remain open.

The first cohort's Core import points are 229/348/5.59 ns for Go/LLGo/Wasmtime,
respectively. Their controls fail and they remain recorded. Absolute times
changed between cohorts; repeating until favorable controls or pooling these
cohorts would conceal that instability. Both cohorts nevertheless locate the
large observed gap at repeated imports rather than the arithmetic control.

## Allocation work is concrete

Separate count runs at 1,001 and 10,000 imports, repeated twice, agree with
the steady timing-run counters. The following are asymptotic events/import;
fixed top-level call setup accounts for the small residuals in raw records.

| Path | Core import | Component import | Counter scope |
| --- | ---: | ---: | --- |
| wasmvm / Go | 1.25 | 9.25 | Go managed allocation events |
| wasmvm / LLGo | 3.25 | 15.25 | LLGo managed allocation events |
| Wasmtime typed | 0 | 1 | Rust System alloc/zeroed/realloc requests |
| Wasmtime dynamic | Not measured | 2 | Same Rust counter |
| wazero / Go | 0 | Not implemented | Go managed allocation events |

These counters do not cover identical allocator domains. Rust counts exclude
foreign C/JIT/mmap allocations, and LLGo accounted bytes are not Go object-size
accounting. They establish the repeated managed work in these particular
paths, not a ranking of total application heap use. Instrumented Rust timings
are excluded. Typed Component requests grow by 104 requested bytes/import;
dynamic requests grow by 296. We have not traced every Rust request to a stack.

Go allocation profiles at rate one identify the Component import's repeated
sites: two codecs, a hostCall and owned argument storage, lifted argument
storage, lowered result storage, an accepted result Value, a fresh Component
Caller, scalar operand storage and batched Core Caller capabilities. Source
and object code confirm these sites. The rate-one Go profile counts tiny
allocation blocks differently from logical `MemStats.Mallocs`; scalar operand
arrays occupy half as many profile objects through tiny-allocation packing.
Profile object percentages are **not CPU percentages** and are not added to
the managed-event table.

In the current public API, scalar operand slices may be retained and modified
after a callback without changing returned Wasm values. Caller copies must
expire together, and a retained copy must never become valid for another call.
`pkg/vm/scalar_host.go` therefore creates independent storage;
`pkg/vm/call.go` creates fresh guards, currently batching four. Those contracts
are stronger than a borrowed stack. They require lifetime separation, **not a
mathematical minimum of one heap allocation per callback**: larger fresh arenas
could reduce event counts without permitting reuse of escaped storage.

## Complete HTTP requests exercise the same mechanisms

Counter-only overlays on the unchanged Rust echo/P3 fixture count requests
after ten warmups. Go and LLGo independently agree at 1, 16 and 256 requests,
with two processes per size/payload/compiler: **24 completed primary count
processes**. Stable counts below are exact in all those runs; scheduler polls
and ready observations can vary and retain ranges in the data.

| Events per completed request | 16 B | 64 KiB |
| --- | ---: | ---: |
| Native entries | 46 | 58 |
| Host dispatches / native Call exits | 38 | 50 |
| Native Return exits | 8 | 8 |
| Core Caller creations | 39 | 51 |
| Component Caller creations | 23 | 23 |
| Core Caller capability entries | 73 | 109 |
| Codec creations | 31 | 31 |
| Clone roots / recursively cloned nodes | 30 / 47 | 42 / 59 |
| Canonical value visits | 85 | 97 |
| Layout Build calls | 116 | 152 |
| Memory Read / ReadString | 31 / 4 | 43 / 4 |
| Memory Write / WriteString | 9 / 2 | 9 / 14 |
| Native frame preparation / publication calls | 46 / 43 | 58 / 55 |
| Component tasks / task resumes | 6 / 16 | 6 / 16 |
| Host completions created/bound/resolved/taken | 5 each | 5 each |

Layout Build counts include cached lookups; they are not layout allocations.
Frame publication updates the changed suffix, not the whole historical stack.
Memory API invocations include bounds/capability checks and are not a count of
bulk copies. Transfer budget charges are logical budget accounting, not bytes
copied. Completion notify counts are not necessarily channel sends.

The 16 B request's 38 host dispatches reconcile exactly: **7 named WIT imports,
17 stream/future builtins, 13 waitable builtins and 1 task.return**. The WIT
imports are two fields constructors, two fields appends, one status setter,
one request.consume-body and one response.new. The 64 KiB request adds exactly
12 `stream.read` builtin dispatches, 12 native entries, 12 clone nodes and 12
memory Read/WriteString calls; task and WIT import counts stay unchanged.
This identifies where the payload-dependent extra transitions occur without
inventing an exclusive microsecond budget for them.

The original mutex/map counter runner had one LLGo 256-request process time
out at 120 seconds before producing its final record. A separately declared
watchdog series completed all twelve LLGo size/payload/repeat combinations,
plus one initial 256-request diagnostic, without a slow event to capture.
The original failure is retained and **unexplained**, not erased by successful
runs. Instrumentation may affect scheduling/allocation; no new production
LLGo deadlock claim follows from this non-reproduced diagnostic.

## Emitted instructions explain the boundary difference

Separate publication-only dump overlays verify that Go and LLGo emit
**byte-identical guest native images for all four fixtures**, including both
policy tiers. The arithmetic and host-import image hashes are respectively
`59ea36ba…73e127` and `b2d3e337…183db1`. Thus the compiler-to-compiler gap in
these kernels is outside their emitted guest machine code: host transitions,
generated host code, allocation/runtime work and scheduling remain candidates.
This result is scoped to these fixtures, not all Wasm instructions.

The Core import fixture makes **4,096 host exits and 4,097 native entries** for
4,096 imports under the counter overlay. A Component import adds two codecs,
two scalar Clone operations and one Component Caller per import. The common
native execution path initializes native.State, authenticates the returned
function/slot/PC state, handles the common HostRequest, then resumes generated
code. These are real host execution steps, not guest arithmetic instructions.

The dumped arm64 Core import's ordinary entry saves 224 bytes of register state
in its wrapper. Immediately before its Call exit at 0x1ac, it spills six live
scalar slot words, records continuation/source state, returns exit code 10,
and reaches the wrapper epilogue at 0x418. A later entry restores registers,
checks state and dispatches to the saved continuation. Cold paths and policy
tiers remain in the dump; disassembling their bytes does not mean they ran.
The production Go host object contains the scalar `runtime.makeslice` call.
LLGo's host object contains `GC_malloc`, `GC_malloc_atomic`, the captured
panic flag and dynamic function handling as well as fast atomic guard checks.
Static call-site totals include cold/error paths and are not dynamic counts.

Wasmtime's serialized Component code instead emits a native guest import
`blr` followed by a dedicated canonical lowering trampoline. In this fixture
the normal lowering path checks entry state, places a scalar in stack storage,
calls the Rust host entry and returns to the guest's native continuation.
The host side still performs task, scope, bounds and resource work. Its
[typed implementation](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasmtime/src/runtime/component/func/host.rs)
checks tuple types when linking and lifts/lowers their specialized storage;
the dynamic path builds `Vec<Val>` and validates dynamic values. The actual
[P3 HTTP bindings](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/p3/bindings.rs)
use generated typed bindings, with selected concurrent imports. Those imports
may allocate boxed futures. Wasmtime is not generally allocation-free.
Its [concurrent host-task path](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasmtime/src/runtime/component/concurrent.rs#L1821)
also creates/deletes task-table entries for synchronous host transitions when
concurrency support is enabled, as it is in this experiment.

wazero's Go
[call engine](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/engine/wazevo/call_engine.go#L378)
dispatches an indexed GoFunction with a view of the native call stack, then
resumes the saved native return address. It does not reconstruct wasmvm's
common interpreter frame machine on that path. Its
[API contract](https://github.com/tetratelabs/wazero/blob/v1.12.0/api/wasm.go#L424)
forbids using operand storage after return. This explains a useful design
difference and provides a Go implementation reference; it does not supply
P3/capability equivalence for free. The five cited registry/module-cache files
were hash-verified byte-for-byte against their tagged upstream sources.

## Design decision and implementation order

Keep the owned compiler and public lifetime contracts. Prioritize the common
Go path before compiler-specific GC tuning. Current evidence warrants the
following shared architectural work, with both backends remaining semantic
references for one set of plans:

1. **Bind internal canonical plans once.** Resolve immutable type/layout,
   scalar lane, ownership and memory-option plans at instantiation. Give
   synchronous internal adapters invocation-scoped codec/result storage, with
   distinct frames for reentry. Keep public dynamic Value/Caller adapters at
   the application boundary. Pending operations must promote exactly their
   owned data before the borrowed invocation ends. No global reusable Caller
   or unlocked memory slice may escape.
2. **Shorten native host transitions.** Add a private synchronous native
   transition/resume path so common execution frames are materialized for
   suspension, reentry, traps or inspection rather than every simple import.
   Go needs a correct stack/GC/unwind bridge; LLGo can use its native ABI where
   compatible. Prefer an explicit indexed adapter contract to arbitrary
   callbacks from generated machine code. Preserve metering, backedge polls,
   stack/slot bounds, reference roots and frame authentication at every
   boundary that still publishes state. This is a substantial ABI change,
   not an instruction-count-only patch.
3. **Apply typed bindings and scoped memory access to built-in WASI.** Share
   WIT-derived descriptors and ownership validation with dynamic bindings.
   Acquire one checked memory capability for a bounded conversion span;
   revalidate after growth/reentry and never retain its backing view across
   suspension. Do not fuse HTTP-specific request sequences or skip budgets.
4. **Measure ready and pending async paths after these changes.** Keep actual
   socket/stream waits, completion races, cancellation, resource retirement
   and joins distinct. Task/event counts identify work but do not establish
   which async primitive owns the remaining exclusive CPU time.

Each implementation needs retained-Caller/operand tests, malformed scalar and
canonical input, limits, reentry, panic/trap, memory growth, cancellation and
cleanup tests under Go/LLGo. Re-run the same mechanism kernels and complete P3
requests before any speed claim; finish with the full competitor portfolio.
Do not turn reduced checks or an internal-only microbenchmark into acceptance.

No hard, unbreakable design ceiling has been proved. The current path has
explicit materialization/lifetime costs that collector tuning cannot remove.
Batching may reduce their allocation event count; typed borrowed internals and
a shorter native transition can remove entire classes of repeated work.
The additional LLGo allocation slope is real, but attributing every event or
cycle to escape analysis remains unfinished. A simplified type-assert/defer
reproducer allocates zero under both compilers; it is retained as a rejected
minimization. Existing [LLGo escape-analysis work](https://github.com/xgo-dev/llgo/issues/2244)
and [GC sensitivity issue](https://github.com/xgo-dev/llgo/issues/3106) remain
the relevant upstream work. No duplicate issue is opened without a specific
reproduced defect.

Machine-readable [summaries](data/performance-p3-mechanisms-2026-11-16.json),
[raw evidence](data/performance-p3-mechanisms-2026-11-16.evidence.json.gz) and
[diagnostic sources/instructions](data/performance-p3-mechanisms-2026-11-16-protocol.tar.gz)
retain both cohorts, counts, allocation runs, compiler identities, source
verification, failures and rejected reproducer. The earlier
[complete-P3 comparison](performance-latest-main-p3-2026-11-16.md) remains the
application-level baseline; no tiny-loop ratio replaces its open gates.
