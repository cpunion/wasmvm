# Competitor, host-compiler and native gaps: 2026-11-02

The remaining gaps are tracked separately. Fast native integer/atomic loops do
not establish a fast interpreter. A throughput difference alone does not establish
an LLGo defect. Controlled Linux amd64/macOS arm64 portfolio acceptance is open.

| Gap | Evidence and action | Remaining deficit |
| --- | --- | --- |
| LLGo shared/atomic interpreter helpers | [Cached cancellation and numeric C bindings](performance-interpreter-atomics-2026-11-02.md): about 21%/22% improvement; checked helper allocation controls become zero | RMW remains about 2.12x WasmEdge and 1.40x wazero interpreter in that campaign |
| Native scalar memory | [Word-loop expansion](performance-native-word-loops-2026-11-02.md) initially improved LLGo 0.526→0.473, but regressed in the [complete combination](performance-atomic-combined-2026-11-02.md); held | Existing native loops remain; C and competitor deficits still need resolution |
| Native atomic throughput | [100-million-iteration C comparison](performance-native-atomic-c-2026-11-02.md): LLGo 1.776, Go 1.772, C 1.723, Wasmtime 1.709 ns/iteration | About 3.1% versus C and 3.9% versus Wasmtime on one ARM64 kernel |
| Interpreter operands and dispatcher counter | [Two complete twenty-case experiments](performance-interpreter-rejected-2026-11-02.md), including GC, SIMD and wait/notify | No consistent meaningful LLGo benefit; neither candidate is adopted |
| GC and SIMD interpretation | All results retained in those twenty-case experiments | WasmEdge still leads most cases; removing operand checks does not solve this gap |

Each campaign has its own source revisions and host conditions. Preparation,
first-call and warm execution remain separate. Do not pool these tables into a
synthetic fastest-per-case engine or calculate improvements across campaigns.

## Independently reproduced LLGo problems

All reports concern pinned `8e45ed5` with `-O3 -lto=full`. Current upstream main
was not rebuilt; these reports do not assert its status.

| Report | Independent observation | Local treatment |
| --- | --- | --- |
| [#2957](https://github.com/xgo-dev/llgo/issues/2957) | Append within existing capacity still materializes a temporary | Reuse admitted guest-frame capacity directly |
| [#2959](https://github.com/xgo-dev/llgo/issues/2959#issuecomment-6566786471) | Uncancelled context.Err/atomic.Value.Load path allocates; unused return storage is materialized | Cache the execution-bound cancellation view for LLGo; retain Go's cheaper standard path |
| [#2958](https://github.com/xgo-dev/llgo/issues/2958#issuecomment-6568684544) | Numeric-only Cgo scalar and two-word-result calls each allocate about 16 MB over one million calls; Go allocates zero | Direct LLGo C declarations for retained C-owned numeric memory; host-buffer checks remain |

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
#2957, #2958, #2959 and #2965 above; no duplicate issue is opened for a timing gap.
