# Go performance and the LLGo main baseline

Go-compiled wasmvm reaching independent competitors remains the priority.
Whole-profile performance and release acceptance are open.

## Retained Go compiler improvement

PR [#162](https://github.com/cpunion/wasmvm/pull/162) removes two serial register
moves from a proven private-word loop. Every original store, cold continuation,
bounds/trap contract and mandatory cancellation poll remains. Its final 25
hosted checks passed before merging; codegen statement coverage is 95.2%.

A separate macOS arm64 event lengthens each call from 50 million to 500 million
iterations, with 20 blocks, 240 fresh processes and seven interleaved warm calls
per process after first-call separation and two warmups. Go1.27.3 and LLGo3d92a58
`-O3 -lto=full` use the same owned backend. These are workload iterations, not
individual Wasm instructions.

| Configuration | Median ns/iteration |
| --- | ---: |
| Go before the source change | 0.814 |
| Go after the source change | 0.271 |
| LLGo after the source change | 0.272 |
| wazero1.12.0 compiler | 0.948 |
| Wasmtime48.0.1 Cranelift | 0.478 |
| Wasmer7.3.0 Cranelift | 0.459 |

The paired Go head/base ratio is 0.3347, 95% diagnostic interval
[0.3311, 0.3386]. Both Go A/A controls resolve within 2.5%. LLGo and all reference
A/A controls remain unresolved. Reference points therefore do not establish
formal competitive acceptance. This event is not pooled with the earlier
50-million-iteration cohort. [Identities, individual intervals and raw records](data/performance-private-word-long-calls-2026-11-13.json)
retain every control and result.

The similar Go/LLGo hot-loop points do not measure pure C-call overhead. The
same owned guest compiler normally emits the same code for the same target and
policy; a sustained loop amortizes the host boundary. Direct LLGo C calls can
still matter for short entries and frequent host transitions. Whole API,
component requests and isolated boundary costs remain distinct measurements.

## Remaining P3 cost

A current-source, post-warmup cycle profile attributes about 34% of cycles to
Go runtime code, 22% to guest code, 17% to component/canonical code and 10% to the
engine. Protocol bookkeeping and final cleanup remain in that diagnostic;
these shares are not isolated subsystem timings or additive marginal savings.
Entry counters find only 38 dead-before-read/exit integer reloads per request,
against roughly 526 register reloads and 892,000 total process instructions.
Entry-mask complexity was not promoted. Two conditional FP-environment
prototypes also passed ABI tests but failed to establish a robust P3 benefit.

A separate fixed 24-process diagnostic runs the identical Go binary with
`GOGC=100` and `GOGC=off`. It uses bounded 512-request batches, full response
checks and matched controls. Turning GC off changes paired CPU by about -5%
and instructions by -7%; allocations remain about705/request. Peak physical
footprint rises from123MB to968MB. The no-GC instruction point remains about
3.7 times Wasmtime's. Reference timing controls are unresolved. This is a
cost investigation, not an exact exclusive GC fraction or a production tuning
recommendation. [Protocol, controls, source identity and raw observations](data/performance-go-gc-cost-2026-11-13.json)
remain separate from the longer throughput cohorts.

Collector removal alone does not close the gap. Further work targets canonical
intermediate values, ownership operations and runtime transitions, alongside
guest code generation. Normal GC and all production checks remain enabled.

## Updated LLGo main

The new compiler baseline is [56fa9aedc957](https://github.com/xgo-dev/llgo/commit/56fa9aedc957cf3c9656c1a70f82c5424ab3b1c5),
with Go1.27.3, LLVM/Clang22.1.8 and the unchanged `-O3 -lto=full` profile.
Follow-up changes include invocation-owned defer captures, natural slice
allocation, static itabs and LLVM bindings0.10.4. macOS executes all seven
codegen/canonical/nativeexec/VM/component/P2/P3 suites successfully. Both
platforms' final semantic CI remains required before integration; Docker Linux
results establish semantics only.

This baseline also supplies real `runtime.MemStats.Mallocs` managed-allocation
events. The service driver now reports them under both compilers, with retained
storage and resource-cohort regression checks. Earlier omitted LLGo counters
remain unavailable, rather than zero allocations. Go managed bytes and LLGo
BDWGC collector bytes keep distinct meanings; neither includes every native,
JIT or linear-memory allocation. Driver statement coverage is95.0% with Go race
checks, not specification coverage.

The compiler update is a separate baseline event. The private-word and GC-off
numbers above retain their original compiler identities and are not presented
as effects of the new LLGo revision. Dedicated LLGo optimization remains
secondary to closing the Go performance gap.
