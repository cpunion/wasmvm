# Competitor, host-compiler and native gaps: 2026-11-02

The remaining gaps are tracked separately. Fast native integer/atomic loops do
not establish a fast interpreter. A throughput difference alone does not establish
an LLGo defect. Controlled Linux amd64/macOS arm64 portfolio acceptance is open.

| Gap | Evidence and action | Remaining deficit |
| --- | --- | --- |
| LLGo shared/atomic interpreter helpers | [Cached cancellation and numeric C bindings](performance-interpreter-atomics-2026-11-02.md): about 21%/22% improvement; checked helper allocation controls become zero | RMW remains about 2.12x WasmEdge and 1.40x wazero interpreter in that campaign |
| Native scalar memory | The initial [word-loop expansion](performance-native-word-loops-2026-11-02.md) regressed in the [complete combination](performance-atomic-combined-2026-11-02.md) and was held; the bounded [longer follow-up](performance-native-word-long-2026-11-02.md) advanced through merged PR #97 | C and Linux competitor deficits still need resolution; the held prototype is not treated as integrated |
| Native atomic throughput | [100-million-iteration C comparison](performance-native-atomic-c-2026-11-02.md): LLGo 1.776, Go 1.772, C 1.723, Wasmtime 1.709 ns/iteration | About 3.1% versus C and 3.9% versus Wasmtime on one ARM64 kernel |
| Interpreter operands and dispatcher counter | [Two complete twenty-case experiments](performance-interpreter-rejected-2026-11-02.md), including GC, SIMD and wait/notify | No consistent meaningful LLGo benefit; neither candidate is adopted |
| GC and SIMD interpretation | All results retained in those twenty-case experiments | WasmEdge still leads most cases; removing operand checks does not solve this gap |

Each campaign has its own source revisions and host conditions. Preparation,
first-call and warm execution remain separate. Do not pool these tables into a
synthetic fastest-per-case engine or calculate improvements across campaigns.

The [final Linux PR #97 benchmark](https://github.com/cpunion/wasmvm/actions/runs/41403501261)
improves LLGo native memory 1.823→1.649 ns/iteration, versus Wasmtime 0.944
and Wasmer 0.880. Its completed integration does not close that remaining x86
gap or permit the favorable local ARM64 result to replace it.

## Independently reproduced LLGo problems

All reports concern pinned `8e45ed5` with `-O3 -lto=full`. Current upstream main
was not rebuilt; these reports do not assert its status.

| Report | Independent observation | Local treatment |
| --- | --- | --- |
| [#2932](https://github.com/xgo-dev/llgo/issues/2932#issuecomment-6579659757) | A return before reaching defer still allocates about 64 reported bytes/call under the pinned full-LTO profile; static/captured defers report about 112/176 | Retain complete cleanup semantics; compiler setup sinking/open-coding remains upstream work |
| [#2957](https://github.com/xgo-dev/llgo/issues/2957) | Append within existing capacity still materializes a temporary | Reuse admitted guest-frame capacity directly |
| [#2959](https://github.com/xgo-dev/llgo/issues/2959#issuecomment-6566786471) | Uncancelled context.Err/atomic.Value.Load path allocates; unused return storage is materialized | Cache the execution-bound cancellation view for LLGo; retain Go's cheaper standard path |
| [#2958](https://github.com/xgo-dev/llgo/issues/2958#issuecomment-6568684544) | Numeric-only Cgo scalar and two-word-result calls each allocate about 16 MB over one million calls; Go allocates zero | Direct LLGo C declarations for retained C-owned numeric memory; host-buffer checks remain |
| [#2966](https://github.com/xgo-dev/llgo/issues/2966) | Native BDWGC AllocZ emits bzero after GC_malloc has already cleared storage; a minimal program and final disassembly reproduce it | Keep the compiler pin; track an upstream allocator correction as a separate future baseline event |

The numeric Cgo reproduction narrows the existing residual-allocation report;
it does not show that struct returns alone cause it. Optimized wrappers allocate
temporary pointer storage then overwrite it before the C call. Approximate
allocation totals and assembly are retained. Throughput-only GC/Go differences
remain investigations. No bounds-check issue was filed from an inconclusive test.

## Competitor mechanism audit

WasmEdge 0.17.1, commit `0502c560d787d969c66f5676b9e11ac64bba3656`, matches the
pinned reference version. Its [dispatch](https://github.com/WasmEdge/WasmEdge/blob/0502c560d787d969c66f5676b9e11ac64bba3656/lib/executor/engine/engine.cpp#L83)
uses a C++ switch and type-specialized helpers. Computed-goto dispatch therefore
does not explain its lead.

Its [atomic helpers](https://github.com/WasmEdge/WasmEdge/blob/0502c560d787d969c66f5676b9e11ac64bba3656/include/executor/engine/atomic.ipp#L126)
use typed atomic addresses after address/alignment checks; its
[operand stack](https://github.com/WasmEdge/WasmEdge/blob/0502c560d787d969c66f5676b9e11ac64bba3656/include/runtime/stackmgr.h#L64)
has unchecked internal getters under validation invariants. These compact paths
motivate removing redundant helper/FFI work. Our held frame experiment shows that
four removed slice checks alone do not explain the LLGo deficit. This is an
inference, not a causal attribution of nanoseconds to individual C++ statements.

WasmEdge checks a stop token at [branches](https://github.com/WasmEdge/WasmEdge/blob/0502c560d787d969c66f5676b9e11ac64bba3656/lib/executor/helper.cpp#L250)
and [returns](https://github.com/WasmEdge/WasmEdge/blob/0502c560d787d969c66f5676b9e11ac64bba3656/lib/executor/engine/controlInstr.cpp#L150).
Its interpreter does not simply omit cancellation. The
[interruptible option](https://github.com/WasmEdge/WasmEdge/blob/0502c560d787d969c66f5676b9e11ac64bba3656/include/api/wasmedge/wasmedge_configure.h#L304)
applies to AOT compilation, not interpreter configuration.

For matching zero-timeout waits, its [wait implementation](https://github.com/WasmEdge/WasmEdge/blob/0502c560d787d969c66f5676b9e11ac64bba3656/include/executor/engine/atomic.ipp#L350)
still registers a waiter and enters a timed condition wait. Our helper already
returns timeout after bounds, cancellation and value checks. This is consistent
with our observed lead on that case, but no isolated causal timing is claimed.
Native RMW being close to C also suggests that the larger interpreter deficit
lies in execution/helper work rather than the atomic instruction alone.

## Remaining acceptance

Profile the remaining interpreter helpers/dispatch and native memory/GC costs,
preserving safety and exact continuations. Reject speculative complexity without
demonstrated benefit. Minimize independent reproductions for further LLGo issues.
Validate selected combined changes and both platforms, then run the declared
controlled portfolio, cancellation and allocation gates. Local wins do not close
those gates or the expanded WASI world-conformance boundary.

[Primary-source audit identities](evidence/performance-gaps-2026-11-02.json)
retain pinned paths, URLs and content hashes without copying competitor sources.

The [complete twenty-case atomic comparison](performance-atomic-combined-2026-11-02.md)
confirms shared/RMW/wait interpreter gains and records why native expansion is held.

[CPU profiles and a new independent channel-polling reproduction](performance-interpreter-profile-2026-11-02.md)
identify an additional runtime cost, reported as [LLGo #2965](https://github.com/xgo-dev/llgo/issues/2965).
Its current-main source was inspected, while timings remain for the pinned compiler.

The [grouped extension-check candidate](performance-interpreter-extension-guard-2026-11-02.md)
improves the complete local interpreter matrix without the earlier private-loop
regression. Native short-control regressions and final-head acceptance remain open.

## Isolated integer-pair follow-up

The [complete integer-pair comparison](performance-interpreter-integer-pairs-2026-11-02.md)
uses the corrected I32-width source as its baseline. The LLGo interpreter improves
integer 60.280→49.789, memory 130.860→100.099 and calls 91.448→82.646
ns/iteration. Baseline A/A controls are retained. This is a descriptive candidate,
not integrated final-head or controlled portfolio acceptance.

The three remaining comparisons in that same campaign are:

| Comparison | Observed head deficit | Next investigation |
| --- | --- | --- |
| Named competitors | LLGo interpreter integer is 2.3% behind WasmEdge, memory 8.6%, calls 32.9%; SIMD widening and RMW take about twice its time | Shared dispatch/SIMD/helper paths, preserving exact polling and fuel |
| LLGo versus Go | GC allocation 166.904 versus 136.300 ns/iteration; notify 57.167 versus 48.500 | Allocation and synchronization profiles; independent reproductions before additional LLGo reports |
| Ordinary native C | LLGo native memory 0.474 versus 0.267 ns/iteration, despite Wasmtime 0.478; integer/calls are much closer | Address lowering and accumulator dependencies; retain bounds, overflow and cold-entry contracts |

Notify and wait-mismatch interpreter results regress 4.4% and 2.7% against the
corrected baseline. Their short samples, all native regressions, unsupported
reference outcomes and the earlier compound experiment remain visible. A fast
case is not substituted for any slower case. The negative-I32 public bit-width
bug found during this investigation belongs to wasmvm and is fixed in PR #98;
it is not an LLGo issue. Existing independently reproduced LLGo reports are
#2957, #2958, #2959, #2965 and #2966 above; no duplicate issue is opened for a timing gap.

The [native-allocation investigation](performance-llgo-native-allocation-2026-11-02.md)
retains the independent #2966 reproduction, matching-check C mechanism control
and original corrected diagnostic. It does not attribute the complete Go/LLGo
allocation gap or claim an unmeasured compiler-patch speedup.

## Scoped packed-vector follow-up

The [packed-vector comparisons](performance-packed-vector-loads-2026-11-02.md)
retain both candidates. The first inlined helper improves widening but regresses
LLGo scalar/full-vector/lane-memory controls 12–20% and is held. Keeping the
widening operations in an outlined helper removes those observed large control
regressions: widening improves 120.236→100.346 ns/iteration (16.5%) for LLGo and
120.094→100.495 (16.3%) for Go. LLGo native controls are still 0.1–0.8% slower.
WasmEdge widening remains 60.690; this 1.65x interpreter gap stays open.
These are separate scoped shared-host campaigns, with their own baseline A/A
controls and every phase retained. Final-head cross-platform validation is
pending; the result is a candidate, not an integrated or controlled parity claim.

## Native comparison-liveness follow-up

The [bounded unroll cleanup](performance-native-unroll-predicates-2026-11-02.md)
keeps live comparison results but removes repeated writes before a proven
overwrite. A capture of 86 body pairs changes only unmetered scalar/float64
memory on amd64/arm64, each 24 bytes smaller. All metered and other captured
functions remain byte-identical.

The native-only scoped campaign uses 50 million core-memory iterations and
five million for five other scalar workloads, matching counts for every engine and C.
LLGo core memory improves 0.450242→0.442667 ns/iteration; every paired block
improves 1.2–1.7%. Go improves as well. A/A calibration remains outside the
controlled target; float64 is inconclusive and control regressions remain
retained. The ordinary C memory control is still 0.265180 (a 1.67x gap).
The diagnostic C CLI admits longer inputs with a separate safe float cap;
its timed kernel assembly is unchanged. This small native cleanup does not
replace the remaining induction, reduction, GC or SIMD work or the x86 gap.

PR #98's [current-head Linux follow-up](performance-interpreter-extension-guard-2026-11-02.md#current-pr-head-linux-follow-up)
also passes its runtime benchmark job and retains its native integer regression;
semantic/report checks remain pending. The candidate branches do not replace
that PR head's integration evidence.

The [isolated reduction split](performance-native-reduction-held-2026-11-02.md)
is held: its only changed native body is the target ARM64 memory loop, but every
paired process block regresses for both Go and LLGo (median tables +2.61%/+2.14%).
Passing semantics and resemblance to C's independent accumulators do not justify
integrating it. The contribution branch retains the simpler implementation;
repeated group control and address induction remain separate hypotheses.

The [combined Linux preflight](performance-linux-combined-2026-11-02.md) also
passes its scoped workflow and source/artifact audits, while retaining Go
interpreter call/memory regressions and a 2.1% LLGo call regression. It does not
replace PR #99's isolated result. Its native memory remains 1.33x Wasmtime.
A compile-only allocation study identifies a concrete x86 issue: fixed-register
weights favor literal operations that encode their operands directly, while a
running local incurs a real frame store and two reloads per source iteration.
Model those emitted costs before further speculative reduction complexity.


## X86 immediate-cost investigation

The [native cost-model candidate](performance-native-immediate-weights-2026-11-02.md)
corrects an owned allocator mismatch: encoded literal work had displaced a
running value, adding a store/two reloads per source iteration. Seven unmetered
x86 bodies change; all ARM64/metered captures remain identical. Local Go/LLGo
and emulated x86 semantics pass. The short real-Linux preflight improves LLGo
memory 11.2% but regresses Go memory 1.5%; it is not a stable-win claim. Longer
native/C comparisons below cover the nine kernels; final contribution validation
and controlled acceptance remain pending. The
compiler pin stays fixed; this is not reported upstream as an LLGo issue.


The [ordinary-C mechanism comparison](performance-native-c-memory-mechanism-2026-11-02.md)
finds no emitted-kernel change when disabling C vectorization. Disabling loop
unrolling raises the observed C memory time 0.267→0.366 ns/iteration; wasmvm
Go/LLGo remain about 0.448. C's induction updates and partial sums motivate
compact compiler work; another speculative reduction rewrite is not accepted
from source resemblance. A/A calibration and controlled performance remain open.


The [longer x86 follow-up](performance-native-immediate-weights-2026-11-02.md#longer-linux-observations-eight-complete-cases)
validates eight complete workloads from an explicitly failed nine-case workflow.
Memory improves both Go/LLGo 11–12% with consistent paired blocks, yet LLGo
remains 1.42x Wasmtime and 1.84x ordinary C. Narrow memory and dependent-address
C gaps remain 2.26x/1.55x. The final struct-allocation reference-classification
failure is retained; that missing case completes separately at its original
manifest count below. Narrow/i31/Go
struct-access regressions and A/A discrepancies remain visible. No new LLGo
issue is inferred from these native throughput differences.


The [bounded allocation follow-up](performance-native-immediate-weights-2026-11-02.md#remaining-allocation-case)
passes separately at the manifest's original 10,000 retained allocations. The
one-million request correctly exceeds unchanged heap limits and is retained as
a resource rejection. Native LLGo allocation is 184.627 versus Go 102.263
ns/iteration (1.81x); its apparent base/head gain is smaller than A/A drift.
This sharpens the host-compiler allocation/collection investigation without
inventing a new LLGo defect report. The nine-kernel observations span two
separate hosts; no aggregate controlled performance claim is made.


## Integrated PR #99 Linux observations

The [final-head audit](performance-interpreter-integer-pairs-2026-11-02.md#integrated-final-head)
passes all required checks and retains exact source/artifact identities. LLGo
interpreter integer/calls/memory improve 23.5%/17.9%/16.6% but still take
1.24x/1.75x/1.34x WasmEdge's time. LLGo calls remain 11.4% behind Go.
Native LLGo memory is 1.312 versus Wasmtime 0.670 ns/iteration. Native
regressions remain retained; the diagnostic PR scope has no ordinary C kernel.
Earlier ARM64 ordinary-C comparisons and later candidates stay separate.
This integration does not close controlled performance or WASI world acceptance.


## Integrated packed loads

PR #100's [final audit and scoped Linux evidence](performance-packed-vector-loads-2026-11-02.md#integrated-final-head-and-linux-follow-up)
pass semantics and retain all diagnostic outcomes. Linux widening improves
LLGo 33.5% but remains 1.65x WasmEdge and 1.23x Go; the separate complete lane
follow-up remains 1.89x WasmEdge and 1.28x Go. Short PR scalar interpreter
regressions of 3–7% remain unresolved; the longer scoped memory control improves
and cannot substitute for integer/call cases. No controlled acceptance is closed.
The native cost-model candidate remains separate from this merged implementation.

## Three-gap follow-up: native group size and LLGo defer

The [eight-copy native memory increment](performance-native-eight-copy-groups-2026-11-02.md)
compares exact `872e257`/`25298a6` sources under both compilers, three native
competitors and matching ordinary C. LLGo memory improves 0.451034→0.419618
ns/iteration, versus Wasmtime 0.474800 and C 0.285740: a local competitor win,
with a 1.469x C deficit still open. Narrow/dependent memory remain 1.481x/1.296x
C and float remains 1.8% behind Wasmtime. Small control/float regressions and
all baseline A/A observations are retained. Linux performance and final-head
cross-platform CI are pending; the earlier x86 cost-model study is distinct.

[The new defer reproduction](performance-llgo-defer-2026-11-02.md) confirms the
existing #2932 issue on the pinned release profile, including allocation on a
return before the defer. It joins #2957, #2958, #2959, #2965 and #2966 as an
independent compiler/runtime finding. Nearby NewSlice2 calls and the complete
GC/Go throughput difference remain investigations. No compiler pin is changed
and no generic timing gap is filed as a new defect.

## Native allocation follow-up candidate

The [bounded batch comparison](performance-native-allocation-batches-2026-11-02.md)
keeps the helper-only experiment and its adverse observations. A separate fixed
twenty-block comparison against the current PR engine finds combined LLGo
allocation-loop medians 63.831→56.883 ns/iteration, Go 47.500→46.394 and
Wasmtime 107.000. The head still takes 1.226x Go's time. Independent counters
reduce native publication batches 154→39 and third-call LLGo host allocation
162032→38032 bytes for the unchanged globally retained 10000-node chain.
The helper-only change removes an independently proved 80-byte idle-call
allocation but does not alone establish a VM throughput gain. No duplicate
LLGo issue or compiler upgrade is introduced. The combination remains on a
follow-up branch while PR #101 checks run; cross-platform integration, startup,
zero engine allocation and controlled portfolio acceptance remain open.
