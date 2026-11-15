# Ordered call, component and native selection optimizations

Base **63a6a80e**, corrected implementation **d71876b1**. The owner requested
finishing the ordered changes before competitive timing. Native call state,
canonical preparation/ownership, and native lane/byte selection were implemented
and validated together. A combined event exposed regressions; causal ablations
preceded the corrected event. Initial, ablation and corrected cohorts remain
separate. Portfolio parity and controlled release acceptance remain open.

## Resulting implementation

Private calls retain integer transfers of complete values, including opaque high
words. Their conservative root map remains zero throughout a native run, their
entry PC is initialized explicitly, and unmetered private returns use static
resume labels instead of redundant continuation publication. Metered behavior,
root publication, frame quotas and every backedge poll remain intact.

Canonical validation builds each immutable root layout once per validation
entry, while every recursive value still charges visits/depth/cancellation.
Accepted ownership retains a fixed-length prefix of the append-only created
list; later results stay unpublished. Preparation-only Callers avoid an unused
canonical resource codec. Fresh Caller lifetimes, affine moves and failure
cleanup remain checked.

amd64 uses gated direct SSE2/SSE4.1 lane extraction, insertion and exact-width
lane memory operations. A bounded byte-expression proof selects BSWAP from
whole-byte masks/shifts/rotations/compatible merges, preserving every observable
intermediate and original cold PC. Precise fuel keeps the original operations.
The real C and Rust SHA fixtures each select15 byte reversals. The ARM byte
selector and expanded logical immediates do not ship.

## Method and scope

macOS arm64/Apple M4 Max: Go **1.27.3**, LLGo **0539ea6e**, LLVM **22.1.8**,
LLGo **`-O3 -lto=full`**, default production GC. Each corrected Core case uses
10 fixed fresh-process blocks, seven interleaved warm calls, two warmups and
symmetric same-binary A/A controls for both owned compilers, wazero **1.12.0**,
Wasmtime **48.0.1**, Wasmer **7.3.0**, and applicable unchecked native C oracles.
The generic call fixture deliberately exceeds leaf expansion eligibility; the
ordinary portfolio function-call case can inline and has a different contract.

P3 payloads16 B/64 KiB each use10 blocks:60 processes under Go and80 under
LLGo, including an independent Go anchor. Four Stores/sixteen clients run20
interleaved512-request batches plus a separate flush. Every full response checks
status, headers, body, trailers and CRC; guest join remains inside completion.
The ordinary owned buffered response and immutable snapshot input are explicit.
CPU/instructions include the complete driver and response checks, not pure guest
execution. Host allocation counters omit foreign Rust/C/JIT allocations and
cannot rank total engine allocation against Wasmtime's Go wrapper.

Displayed times are process medians; paired ratios use block-level geometric
averaging and need not equal quotients of displayed medians.

Startup, validation/backend compilation, instantiation, first call, warmups,
warmed API calls and cleanup are retained separately. No timing subtraction,
cohort extension or pooling is used. No owned measurement-host builds, tests,
profiles, analysis, edits or pruning overlap timed cohorts. Other user/desktop
activity remains uncontrolled and is recorded in load snapshots. Ratios use
block-paired geometric means and marginal bootstrap95% intervals. A/A resolves
only with its entire interval inside0.975–1.025.

The Linux hosted event retains the original22 cases and separately named four
sustained GC/six sustained Threads cases, both backends/compilers, all independent
references, unavailable capabilities and every lifecycle phase. Corrected builds
also calibrate Go interpreter/native rows; initial builds calibrated LLGo only.
Hosted affinity does not provide physical isolation. Allocation throughput uses
per-process means including pauses, never the median diagnostic.

## Corrected arm64 results

| Measurement | Go head/base [95% interval] | LLGo head/base [95% interval] |
| --- | ---: | ---: |
| Generic native calls | **0.95690 [0.94962,0.96489]** | **0.93664 [0.92208,0.94784]** |
| SIMD lane loop |0.99776 [0.99312,1.00281]|1.00547 [0.99890,1.01169]|
| Clang SHA|1.00211 [0.99827,1.00592], controls fail|**1.00855 [1.00282,1.01481] adverse**|
| Rust SHA|0.99291 [0.98674,0.99890]|0.99576 [0.98665,1.00621]|

Generic call controls resolve for both compilers. Go uses4.828ns/source
iteration, LLGo4.704ns; Wasmtime2.075ns, wazero2.129ns, Wasmer2.078ns,
unchecked native C2.083ns. The paired Go/Wasmtime gap is **2.28038
[2.26210,2.30427]** with resolving controls. The ARM lane path is unchanged;
Go/Wasmtime is **0.86587 [0.85977,0.87153]**, also resolving. Private/shared
memory source-change and competitive controls fail; their points remain
observations, not established regressions or wins.

The small LLGo Clang SHA regression remains explicit. No new ARM byte/mask
selection ships; call-stub instruction placement and host/binary layout effects
are candidates, not isolated causes. The earlier ablations exclude the removed
vector transfer as the remaining cause. A causal gap remains; no LLGo compiler
bug is established by this event.

| P3 per request | Go16 B | Go64 KiB | LLGo16 B | LLGo64 KiB |
| --- | ---: | ---: | ---: | ---: |
| Allocations, base→head|655→**652**|706→**703**|973→**970**|1145→**1142**|
| CPU head/base|0.99736|0.99813|1.01798|1.03249|
| CPU95% interval|[0.98931,1.00612]|[0.99051,1.00561]|[0.98483,1.05768]|[0.96850,1.11699]|

Allocation controls resolve. Go instruction ratios are **0.99197
[0.98917,0.99426]** and **0.99595 [0.99383,0.99764]**, with resolving own
controls. CPU improvement is not established under either compiler. LLGo CPU
and instruction controls fail; adverse points are retained. Go/Wasmtime CPU
points are2.50x/2.62x with failed reference controls; LLGo/Go3.09x/2.88x and
LLGo/Wasmtime7.77x/7.60x are likewise unresolved. Known LLGo extra-copy and
runtime synchronization work remain separate from guest code attribution.

## Rejected changes and adverse audit

The initial **909ee5bd** 128-bit call-copy path regressed arm64 generic calls:
Go **1.25631 [1.25094,1.26117]**, LLGo **1.24654 [1.24093,1.25177]**;
both own controls resolve. A separate360-process ablation finds that restoring
integer transfers changes Go to **0.95236 [0.94559,0.95907]** against base.
The vector candidate retires fewer instructions (**0.95848x**) but consumes
**1.25591x cycles**; scalar restoration consumes **0.95529x cycles**, all with
resolving controls. This identifies the value-transfer path, not the host Go
compiler or instruction count, as the major regression. Exact load/store
forwarding, dependency and execution-port shares were not isolated. The vector
helper and its tests were removed from both targets rather than retaining an
unqualified second transfer mechanism.

Initial ARM Clang SHA regressed Go1.26% and LLGo1.58% with resolving controls.
Disabling expanded masks also reduces ARM byte-fusion selections, so it is a
joint selection/layout ablation, not an isolated logical-immediate latency
measurement. Fewer retired instructions did not prove a CPU improvement. Both
ARM expansions were rejected. The corrected small LLGo adverse result above
remains open; the initial favorable P3 CPU points do not replace corrected data.

An attempted shared ordinary load/add/store selector selected zero real fixture
groups across current safety boundaries. Its positive selection assertion
failed; it was removed without timing. Shared loads/stores and per-backedge
polls remain real operations. Existing private-word caching is unchanged.

The first call-state edit accidentally removed an implicit zero-PC initialization
along with a root-map store. Existing metered linked-call tests crashed; explicit
PC initialization fixes it. The initial crash was visible in tool output but
was not fully captured to disk. Retained logs also include the rejected ARM
mask/projection selection and an invalid test mask before correction.

Initial Linux Go interpreter points were4–8% adverse without Go A/A calibration.
The corrected event adds those controls and does not generally reproduce those
large changes. The unchanged macOS Go interpreter main loop has2528 instructions
in both initial binaries; normalized instructions are identical after resolving
ADRP/address pairs, but entry placement changes from page offset3552 to64.
This supports investigating layout, not claiming its exact performance share.
Small adverse corrected rows and inconclusive controls remain in the full data;
large LLGo interpreter improvements are not attributed to native guest selection.


## Corrected Linux portfolio

[Workflow and phase artifacts](https://github.com/cpunion/wasmvm/actions/runs/43726122216). These are separate hosted groups, not a pooled score. Values below are native warm ns/source iteration (call control is ns/call); allocation uses process means including pauses. Reference rows have no symmetric reference A/A in this hosted survey, so rankings remain observations. Missing GC capability is explicit. Full interpreter/WasmEdge rows and controls remain in the evidence.

| Case | Go | LLGo | wazero | Wasmtime | Wasmer | Native C |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
|applications/sha256-clang|450.500|445.115|548.961|335.902|336.468|257.271|
|applications/sha256-rust|422.490|422.907|545.051|351.942|352.166|257.351|
|core/call-control|4266.000|4431.000|1015.000|2018.000|3441.000|unavailable|
|core/integer-loop|1.871|1.872|1.739|1.801|1.806|1.737|
|core/function-calls|1.002|0.995|1.752|1.242|1.461|1.086|
|core/memory-load-store|1.050|1.046|1.686|0.924|0.932|0.746|
|gc/i31-roundtrip|1.726|1.864|unavailable|2.286|unavailable|unavailable|
|gc/struct-access|1.974|1.966|unavailable|1.433|unavailable|unavailable|
|gc/struct-allocation|48.358|73.764|unavailable|118.288|unavailable|unavailable|
|gc/struct-indirect-access|4.105|4.523|unavailable|6.838|unavailable|unavailable|
|memory/narrow-unaligned|1.709|1.695|2.603|1.712|1.707|1.082|
|memory/dependent-address|1.908|1.905|2.924|2.210|2.214|1.888|
|memory/float64-unaligned|1.059|1.057|1.605|0.957|0.968|0.952|
|runtime/simd-memory-full|7.747|7.718|9.004|7.254|7.520|8.709|
|runtime/simd-memory-widen|0.784|0.787|1.894|0.832|0.834|3.241|
|runtime/simd-memory-lanes|8.032|7.999|9.939|8.225|8.192|7.083|
|threads/private-memory|1.134|1.131|1.641|0.488|0.613|unavailable|
|threads/shared-memory|0.998|0.974|1.952|0.488|0.621|unavailable|
|threads/atomic-rmw|2.859|2.852|2.460|2.542|2.691|unavailable|
|threads/notify-empty|2.441|2.408|41.350|9.624|13.756|unavailable|
|threads/wait-mismatch|2.756|2.746|52.717|13.191|4.938|unavailable|
|threads/wait-timeout|2.966|2.858|655.941|75.702|154.567|unavailable|

Separately declared sustained cases:

| Case | Iterations | Go | LLGo | Wasmtime |
| --- | ---: | ---: | ---: | ---: |
|gc/i31-roundtrip|1000000|1.496|1.494|2.028|
|gc/struct-access|1000000|1.161|1.166|1.139|
|gc/struct-allocation|300000|64.957|78.744|85.095|
|gc/struct-indirect-access|1000000|3.826|3.796|6.329|
|threads/private-memory|100000|0.345|0.355|0.384|
|threads/shared-memory|100000|0.423|0.434|0.381|
|threads/atomic-rmw|100000|6.645|6.633|4.807|
|threads/notify-empty|100000|1.515|1.541|12.783|
|threads/wait-mismatch|100000|1.426|1.465|16.233|
|threads/wait-timeout|100000|1.565|1.612|77.204|

Final hosted SIMD lane paired ratios are **0.85862 [0.85717,0.86021] Go** and
**0.85864 [0.85705,0.85995] LLGo**, with resolving source controls. Rust SHA is
**0.97552 [0.96837,0.98317] Go**, **0.98034 [0.97412,0.98461] LLGo**; Clang
SHA is0.99331 [0.98558,0.99975] Go and0.99777 [0.99312,1.00244] LLGo. These
are hosted source-change observations, not controlled competitive acceptance.
The revised Go interpreter has small adverse points on Clang SHA0.32% and
SIMD lanes0.62% despite unchanged interpreter source; they are retained with
resolving source controls. Their exact placement/runtime cause remains open.
Tiny call-control and short GC/memory rows often fail calibration. No claim
about a compiler defect or broad interpreter speedup follows from these points.
The longest final benchmark job is memory/SIMD at9m10s; shared build6m41s,
sustained Threads6m06s. All seven selected jobs succeed without a rerun.

## CI cold-cache correction

The first final-head semantic run canceled the macOS CLI at10 minutes and the
async shard at12 minutes. The CLI log shows no package-object cache hit and an
unfinished native linker. Observed matching caches belong to PR-scoped refs,
not a reusable default-branch warmup. [Cache scope rules](https://github.com/actions/cache#cache-scopes)
explain why the preceding PR's objects do not supply the next PR's baseline.
These are build-budget failures, not successful guest execution.

A single no-link library-union build now runs per platform before the unchanged
semantic/CLI matrix. Consumers restore its profile-qualified cache without
uploading duplicate shard copies. Main pushes run only these two library warmups,
so future PRs can restore default-branch objects without repeating semantic
execution on merge. All23 existing semantic/shared checks remain; two cache
producer checks are added. All32 test packages, pinned compiler identity and
`-O3 -lto=full` execution remain mandatory. Local union compilation succeeds and
its trace contains no native link; final-head hosted timing is still required.

## Validation, evidence and remaining work

All32 root Go test packages pass, all31 executable packages remain above90%
statement coverage (codegen97.87%), affected race tests pass, and actual LLGo
`-O3 -lto=full` codegen/native/component/P3 plus service/P3 measurement tests pass.
Cold-PC complete frames, large mixed tuples, fuel, cancellation, expired Callers,
prepared failures and affine cleanup remain covered. Ownership tests force both
same-storage append and reallocation before publication/cleanup. The additional
Go1.27.0 amd64 Docker execution is supplementary semantics, not pinned final
qualification or performance evidence. Final-head dual-platform CI remains the
merge requirement.

The [machine-readable report](data/performance-ordered-optimizations-2026-11-15.json)
and [compressed primary observations](data/performance-ordered-optimizations-2026-11-15.evidence.json.gz) retain both source events,
all raw cohorts/controls/phases, Linux capabilities, rejected patches, build and
compiler/library hashes, host observations and validation outcomes. Unused per-request latency vectors and duplicated protocol JSON are compacted
to count/hash references; full originals remain in the hashed raw files. Every
primary timing/counter/phase/control used here is exact. Full local
records are under `artifacts/ordered-optimizations-20261115/`.

Remaining priorities are general-call frame/argument metadata, shared-loop
bookkeeping that preserves real accesses and every poll, SHA dependency/register
scheduling, and complete P3 host/canonical work. The small residual SHA regression
and remaining layout/control uncertainty stay open. Controlled cross-platform
portfolio and release performance gates are not complete.
