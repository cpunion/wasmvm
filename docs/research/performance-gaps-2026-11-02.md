# Competitor, host-compiler and native gaps: 2026-11-02

The remaining gaps are tracked separately. Fast native integer/atomic loops do
not establish a fast interpreter. A throughput difference alone does not establish
an LLGo defect. Controlled Linux amd64/macOS arm64 portfolio acceptance is open.

| Gap | Evidence and action | Remaining deficit |
| --- | --- | --- |
| LLGo shared/atomic interpreter helpers | [Cached cancellation and numeric C bindings](performance-interpreter-atomics-2026-11-02.md): about 21%/22% improvement; checked helper allocation controls become zero | RMW remains about 2.12x WasmEdge and 1.40x wazero interpreter in that campaign |
| Native scalar memory | The initial [word-loop expansion](performance-native-word-loops-2026-11-02.md) regressed in the [complete combination](performance-atomic-combined-2026-11-02.md) and was held; the bounded [longer follow-up](performance-native-word-long-2026-11-02.md) advanced through merged PR #97 | C and Linux competitor deficits still need resolution; the held prototype is not treated as integrated |
| Native atomic throughput | [100-million-iteration C comparison](performance-native-atomic-c-2026-11-02.md): LLGo 1.776, Go 1.772, C 1.723, Wasmtime 1.709 ns/iteration | About 3.1% versus C and 3.9% versus Wasmtime on one ARM64 kernel |
| Native i31 payloads | [Direct payload selection](performance-i31-payload-2026-11-03.md): LLGo 2.543→1.350, Go 1.363, Wasmtime 1.811 and ordinary C 1.286 ns/iteration in one final campaign | Merged PR #106 with dual-platform semantic validation: 46.9% LLGo improvement, 25.5% below this Wasmtime's time, 4.9% above C; other GC, startup and controlled all-platform performance gates remain open |
| Interpreter operands and dispatcher counter | [Two complete twenty-case experiments](performance-interpreter-rejected-2026-11-02.md), including GC, SIMD and wait/notify | No consistent meaningful LLGo benefit; neither candidate is adopted |
| GC and SIMD interpretation | All results retained in those twenty-case experiments | WasmEdge still leads most cases; removing operand checks does not solve this gap |

Each campaign has its own source revisions and host conditions. Preparation,
first-call and warm execution remain separate. Do not pool these tables into a
synthetic fastest-per-case engine or calculate improvements across campaigns.

The [two native-cleanup scope experiments](performance-native-allocation-scope-held-2026-11-03.md)
remove a known LLGo defer allocation but do not improve the complete measured
scope. Both production changes remain held; their adverse results and new
pre-entry cancellation/panic reservation tests are retained. The i31 follow-up
keeps its own three source-changing campaigns separate from those experiments.

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
| [#2973](https://github.com/xgo-dev/llgo/issues/2973) | Independent startup markers narrow the gap to first-goroutine creator symbolization and Darwin's default metadata path; same-source constructor-to-main is 67.0 ms versus 2.64/1.93 ms under separate diagnostic metadata profiles | [Retain all marker/handshake/profile campaigns](performance-llgo-startup-markers-2026-11-03.md); update the existing issue; preserve normal traceback/debugger contracts and the production compiler profile |
| [#2975](https://github.com/xgo-dev/llgo/issues/2975) | Nonescaping function-to-interface conversion retains AllocU under full LTO; 100,000 conversions allocate about 3.2 MB versus zero in Go | Preboxed assertion controls allocate zero; this reproduction does not attribute warm VM callback costs |

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

The later [independent heap null-branch experiment](performance-heap-null-branches-held-2026-11-03.md)
also remains held: smaller ARM64 bodies do not close the direct/indirect heap
gap, and the full campaign retains adverse Go/LLGo controls. The selected
implementation keeps its original null-safe cache protocol.

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
## Coalesced-read prefix candidate

The [complete twenty-case comparison](performance-prefix-chains-2026-11-02.md)
reduces dispatch for already-coalesced local reads, preserving all original PCs,
fuel and polling. LLGo interpreter integer/calls/memory improve 16.4%/12.6%/19.0%;
integer/memory lead the named WasmEdge configuration in that local campaign.
SIMD widening, lanes and RMW still take 1.47x/1.75x/1.97x its time.

All 86 captured native bodies match, but a native GC regression and short wait
regressions remain retained. A fresh native A/A follow-up narrows but does not
close that uncertainty; the candidate is not integrated or performance-accepted.
Go/LLGo allocation gaps and ordinary-C native memory deficit remain open.

## Complete dispatcher/allocation integration candidate

The [twenty-case combined comparison](performance-core-dispatch-allocation-2026-11-02.md)
keeps exact native and interpreter baseline A/A controls. Both host compilers
improve every nonempty interpreter case. LLGo native allocation improves
62.604→54.746 ns/iteration, while short native controls retain material
regressions. Guest-body identity does not waive them; controlled acceptance
and integration remain open. Loop-aware reservation sizing is a candidate
investigation for one-constructor functions, not an attributed result.

The subsequent [general I32-prefix experiment](performance-all-i32-prefixes-held-2026-11-02.md)
is held despite a memory/RMW win: LLGo float and SIMD-lane cases regress
20.9%/25.2% against stable corresponding A/A controls. Its semantic checks
pass, but its production code is excluded from this selected branch. The
previous prefix implementation and all negative evidence remain preserved.

## Loop-sensitive native reservations and upstream status

The [seven-case native batch-sizing comparison](performance-native-batch-hints-2026-11-03.md)
uses the exact combined source as its baseline, five interleaved blocks and
Go/LLGo baseline A/A controls. LLGo short struct-access medians improve 11.2%,
Go 5.2%; the loop retains the 256-handle batch and identical guest code. Native
allocation regresses Go 4.4% and LLGo 2.1%, with respective A/A +0.8%/+2.0%.
Every adverse block and first-call/setup sample remains retained. This is a
local candidate, not final-head CI or controlled acceptance.

The three deficits remain separate in that campaign: LLGo allocation takes
1.169x Go's time; native struct access/indirect access take about 1.66x/1.86x
Wasmtime; native scalar memory takes 1.566x unchecked C despite leading the
named Wasmtime case. Adapter readiness is about 93.1 ms LLGo versus 3.9 ms Go,
distinct from warm execution. No result from an earlier campaign is substituted.

At the new upstream inspection, [LLGo #2959](https://github.com/xgo-dev/llgo/issues/2959)
is closed by merged [#2963](https://github.com/xgo-dev/llgo/pull/2963). The
pinned release-profile reproduction remains historical evidence; current
upstream was not rebuilt or timed. #2932, #2957, #2958, #2965 and #2966 remain
open, and #2971 remains an open contribution. No duplicate issue is created
from an already-reported problem or an unexplained timing ratio.

The [independent startup investigation](performance-llgo-startup-2026-11-03.md)
reproduces the readiness deficit without VM work: os-only LLGo 34.112 ms versus
Go 3.508 ms. A separate C GC/FFI control is 4.247 ms versus LLGo 34.052 ms.
It is reported as [LLGo 2973](https://github.com/xgo-dev/llgo/issues/2973), explicitly retaining
an open root cause and pinned-only scope. Warm execution remains a distinct gap.

## P1 host-boundary follow-up

[PR #102 final audit](pr102-final-ci-2026-11-03.md) retains all 240 reports and
both distinct-host benchmarks, including adverse original Intel results. Its
rebalanced six-shard CI still executes all 31 packages/compiler/platform.

The [raw host-buffer experiment](performance-host-storage-held-2026-11-03.md)
is held: LLGo throughput does not consistently improve and Go's native control
regresses 9.8%. Its production code is excluded from the selected
[operation cancellation-view candidate](performance-operation-context-2026-11-03.md).
The latter improves LLGo native P1 by 4.4–5.2%, interpreter by 4.5–8.7%, while
retaining all original control regressions and A/A samples. Args-sizes still
takes 3.42x Go, 29.9x Wasmtime and 68.9x wazero native. Ordinary C's same-algorithm
control stays separately identified with its missing VM policies and coarse
short-call timing. Final contribution CI and controlled acceptance remain open.

The [independent input-parameter reproduction](performance-llgo-value-parameter-2026-11-03.md)
is reported as [LLGo #2974](https://github.com/xgo-dev/llgo/issues/2974): a
nonescaping 112-byte input's address causes an unconditional AllocU(112), about
128 reported bytes/call, versus Go zero. The value-only control allocates zero
under both compilers. This narrows the native state-copy investigation without
assigning all remaining VM allocation to it or changing the compiler pin.

The [borrowed native-state comparison](performance-native-state-borrow-2026-11-03.md)
removes one pinned-LLGo 112-byte input-parameter allocation from each native
entry using execution-owned numeric storage. Native P1 improves 2.7–6.1% and
reported host allocation drops about 1.28 MB per 10000 callbacks; all original
short-arithmetic regressions and the separate sustained diagnostic stay visible.
Args-sizes still takes 3.28x Go, 28.51x Wasmtime and 64.52x wazero native in this
campaign. Ordinary C remains a separate unchecked lower bound. The candidate
preserves the native record/lease/close protocol; integration and controlled
acceptance remain open.

[PR #103's final audit](pr103-final-ci-2026-11-03.md) now records its merge,
all required semantic checks and retained core benchmark regressions. The
[immutable P1 process view](performance-p1-process-context-2026-11-03.md)
separately improves LLGo native P1 by another 4.5–6.5% against the state-borrow
baseline and saves about 3.2 MB per 10000 callbacks. The two studies keep their
own A/A and adverse control results; their percentages are not added or pooled.
Args-sizes still takes 3.10x Go, 27.75x Wasmtime and 61.16x wazero native in the
new campaign. This known #2959 workaround creates no duplicate upstream issue;
all-target controlled acceptance and host-boundary work remain open.

## Host-boundary architecture and reference policy audit

The [explicit entry-FFI experiment](performance-native-entry-ffi-held-2026-11-03.md)
is held despite removing about eighty reported bytes per native entry: P1 wins
are only 2.4–3.0% and original native arithmetic controls regress materially.
It is excluded from PR #104. The [qualitative stack/source audit](performance-p1-boundary-profile-2026-11-03.md)
identifies repeated context/admission/allocation and public marshaling paths,
while wazero uses its existing numeric host stack and precomputed metadata.
Captures are first-long-call diagnostics, without CPU-time attribution.

The [separate cancellation-policy comparison](performance-p1-cancellation-policy-2026-11-03.md)
keeps default and cancellation-enabled wazero configurations explicitly separate.
Args-sizes LLGo still takes 37.06x the enabled reference and 3.10x Go in that
campaign. A separate fixed one-million-iteration arithmetic diagnostic finds
LLGo 1.563542, Go 1.582500 and C 1.560000 ns/iteration; default/enabled wazero is
1.692125/13.894500. The original short controls and all A/A samples remain.
Different poll granularity, incomplete calibration and the rest of the portfolio
keep acceptance open. Next investigate shared numeric host marshaling and
cheaper admission/context construction while preserving lifecycle contracts.

## Shared scalar host ABI and P1 operation effects

The [general scalar ABI comparison](performance-scalar-host-abi-2026-11-03.md)
adds independent numeric callback stacks to the existing driver, preserving
Caller/reentry/GC/error/suspension contracts and public reference callbacks.
Go native P1 improves 6.3–6.8%, LLGo native 2.4–4.9%; its separate existing-public-
host control retains 1–2% regressions and every original arithmetic regression.
No later operation-effect percentage is combined with this source comparison.

The [operation-effect comparison](performance-p1-call-effects-2026-11-03.md)
then avoids unpublished private child contexts for the complete four-operation
configuration/guest-memory group. I/O retains independent callback cancellation.
Go native improves 37.3–38.7%, LLGo native 37.6–42.6%, interpreter 42.6–48.1%.
Reported LLGo native args allocation falls 33.44→18.09 MB per 10000 callbacks;
Go falls 4.48→2.08 MB. Original arithmetic regressions and all A/A records remain.

Args-sizes still takes 3.09x Go, 21.14x cancellation-enabled wazero and 15.43x
default Wasmtime in this latest study; policies and poll granularity differ.
Ordinary C remains an explicitly unchecked lower bound. The selected source
has no new controlled portfolio or sustained native acceptance claim. Next
investigate capability/memory helpers and native-only LLGo allocation while
retaining existing independently reproduced upstream issue boundaries.

The [generic memory-batch comparison](performance-host-memory-batch-2026-11-03.md)
reduces repeated Caller entry for complete prechecked multi-field outputs.
Paired-size LLGo native improves another 11.1–11.7% against the exact operation-
effect source, with its own A/A and adverse controls retained. Reported native
args allocation falls 18.09→15.85 MB per 10000 callbacks; Go stays about 2.08 MB.
Args still takes 2.97x Go and 19.35x cancellation-enabled wazero in that study.
No earlier percentage or favorable control replaces any new slower result.

The [independent memory-profile reproduction](performance-llgo-memprofile-2026-11-03.md)
confirms that the pin exposes size-class records without allocation stacks.
Go identifies both allocation sites; LLGo reports none. Evidence was added to
existing closed proposal #2351, whose PR #2027 was closed without a merge,
rather than creating another issue. This is an attribution limitation, not a
new VM timing claim or an assertion about unbuilt current upstream main.

The [type-only dispatch experiment](performance-host-type-dispatch-held-2026-11-03.md)
is held. Its exact-source public-host control does not establish a useful gain,
and the full original P1 scope retains env/copy regressions. The experimental
production change is excluded; the selected checked-memory implementation stays
unchanged. Existing public-host regressions and their cause remain unresolved.

The [final combined scalar-host comparison](performance-scalar-host-combined-2026-11-03.md)
measures actual PR #104 production against the complete selected contribution:
LLGo native P1 improves 39.7–50.4%, interpreter 45.5–57.6%, Go native
44.6–54.9%. No per-source percentages are composed. Original integer regressions
and all A/A rows remain. Args-sizes still takes 2.96x Go and 19.15x enabled
wazero; startup, native C and controlled portfolio gaps remain open.
The [PR #104 audit](pr104-final-ci-2026-11-03.md) records its completed merge,
nineteen required jobs and 240 verified execution reports, including the
macOS capacity history without rerunning successful work.

The [PR #105 audit](pr105-final-ci-2026-11-03.md) records the selected scalar-host
implementation's completed dual-platform validation and adverse Linux controls.
The subsequent [host-binding experiment](performance-host-binding-held-2026-11-03.md)
and [heap-to-local experiment](performance-heap-local-held-2026-11-03.md) remain
held: neither establishes a useful gain for its target. Their production changes
are excluded from the selected i31 follow-up; additional heap continuation tests
pass against the existing implementation with both compilers.
The [independent boxing report](performance-llgo-function-boxing-2026-11-03.md)
retains allocation controls and disassembly identifying conversion, rather than
assertion extraction, as the allocating operation.

The [PR #106 audit](pr106-final-ci-2026-11-03.md) closes its semantic integration,
not the GC/portfolio performance gates. Two subsequent
[native heap-cache experiments](performance-native-heap-caches-held-2026-11-03.md)
remain held: extra cache homes and a two-way cache do not establish a gain.
Their production changes are excluded; independent multi-handle/cold-entry/
trap-state tests also pass against selected production with Go and LLGo.
Exact disassembly identifies redundant constant materialization before ARM
immediate operations; investigate sharing the existing amd64 constant fusion
before adding more cache state. No new LLGo defect is inferred from these timings.

The [shared immediate selector and overwritten-literal follow-up](performance-immediate-literals-2026-11-03.md)
uses the existing constant fusion on ARM as well as amd64, retaining all cold
entries. The final original-production comparison improves LLGo indirect access
4.587916→4.496333 ns/source iteration (2.0%, A/A +0.09%), still 1.76x Wasmtime
and 1.60x ordinary C. LLGo calls +1.08%, Go allocation +5.26% and unstable
empty-call controls remain adverse evidence. Final contribution CI is pending;
this does not close portfolio acceptance. A separate clock reproduction finds
Darwin C CLOCK_MONOTONIC's 1000-ns granularity. The final study uses the same
finer clock for both C references and native C, retains Go timers, and keeps
the earlier source/clock study separate. Single-call C results near resolution
are not precise pure-call costs. Component clocks and batched latency remain open.

The [PR #107 final audit](pr107-final-ci-2026-11-03.md) retains the corrected
Linux integer result and the earlier adverse head. The subsequent
[atomic native lease admission](performance-native-lease-admission-2026-11-03.md)
removes hot code-pin locking while preserving concurrent close and executing
entries. [PR #108 merged with full CI](pr108-final-ci-2026-11-03.md): Linux LLGo
empty native latency improves 5461→2885 ns, still 2.06x Go; sustained
integer/calls/memory are 1.020x/0.516x/1.115x Wasmtime in this core-only campaign.
Separate macOS batches improve full LLGo empty native API 5.4% and public-host
callback 10.2%, with about 224 fewer reported allocation bytes per short call.
The bounded argument-cache experiment is held; all adverse controls remain.
Independent startup evidence narrows [LLGo #2973](https://github.com/xgo-dev/llgo/issues/2973#issuecomment-6593936533),
and the actual boundary-allocation evidence updates [LLGo #2932](https://github.com/xgo-dev/llgo/issues/2932#issuecomment-6596297878).
Public-boundary/compiler/native-C gaps remain separate and open; no standalone
ABI component or tiny C identity result replaces complete-VM acceptance.

[Operation-owned interruption](performance-native-operation-cancel-2026-11-03.md)
reduces background native public allocation from about 3168 to 2096 reported
LLGo bytes/call, with about 31% local API and 43% session improvements. It also
fixes an exact-main session teardown deadlock involving concurrent capability
use. [PR #109 final CI](pr109-final-ci-2026-11-03.md) verifies all profiles after
balancing the existing shards: Linux VM 7m32s versus its original canceled
12m34s job; all final jobs stay below ten minutes. Final Linux native short call
is 1888 ns, still 2.81x Go. LLGo integer/calls/memory remain behind Wasmer in
this core-only run; Go memory +17.5% and earlier live-parent/dispatch regressions
stay open. Evaluate the already merged LLGo named-result fix separately from
the pinned compiler before retaining or removing related workarounds. Complete
native-C/controlled portfolio and expanded-world acceptance still requires work.
