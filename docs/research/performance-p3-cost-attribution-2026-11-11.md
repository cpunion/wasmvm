# P3 request cost attribution

**Diagnostic study.** The largest gaps include native execution boundaries, component runtime work and host-runtime scheduling. Performance and release acceptance remain open. The selected next optimization is compiler lowering of canonical task context access.

The [follow-up implementation](performance-task-context-2026-11-11.md) now
provides checked context accessors in both backends. The measurements below
retain their original base-source scope.

## Exact process counters

M4 Max/macOS arm64. Main source `4ce75302c8def7dfd85680a33d2e7a55593c8dfa`, Go 1.27.3, experimental LLGo compiler e2cca2c with sharded runtime f2e4cb426, and the independent Wasmtime 48.0.1 reference. Every LLGo binary uses `-O3 -lto=full`. The retained executables have the same implementation as the earlier main-source HTTP event.

The same 16-byte plain Rust P3 HTTP component runs with four workers, 16 clients and 256 requests/client. Three fixed rounds, seed 2026111192, record 30 measured batches per process. The existing interleaved protocol exposes batch readiness. Darwin `proc_pid_rusage(RUSAGE_INFO_V4)` records 29 intervals between readiness markers after preparation, first execution and two warmups. Each interval contains 4096 checked requests plus existing batch-statistics/readiness bookkeeping. Cleanup follows the counter window.

The 15 processes include three engine profiles and two ordinary HTTP controls. All 1,843,200 measured responses pass the common result checks. The hardware-counter windows cover 1,781,760 requests.

| Configuration | CPU μs/request | System CPU μs/request | Retired instructions/request | Instructions / Wasmtime |
|---|---:|---:|---:|---:|
| Wasmtime 48.0.1 | 22.76 | 5.82 | 206,158 | 1.00× |
| wasmvm / Go | 136.64 | 0.72 | 1,894,737 | 9.19× |
| wasmvm / experimental LLGo | 357.92 | 90.87 | 2,537,099 | 12.31× |

These are medians of three per-process observations. Process CPU sums all threads. The phase and request-wall timers have their own boundaries. Retired instruction counts show substantial excess work under Go as well as LLGo.

Darwin's `ri_user_time` and `ri_system_time` use Mach absolute ticks on this host. The timebase is 125/3 ns per tick. A local CPU-burn calibration agrees with `resource.getrusage` within 0.03 ms. Initial unconverted derived CPU labels remain in the raw diagnostic file. The corrected JSON supplies the canonical CPU values. Hardware instruction/cycle counts retain their native units.

## Finite collector counterfactual

A separate short-batch cohort uses identical executables, four workers, 16 clients and 16 requests/client. Twelve measured batches follow first execution and two warmups. Three fixed rounds, seed 2026111193, compare default collection with Go `GOGC=off` and BDWGC `GC_DONT_GC=1`. A native library probe verifies the BDWGC switch, and every collector-disabled measured batch records zero collections. Object allocations and full request checks execute in all configurations.

| Configuration | Default CPU μs/request | Collector-disabled CPU μs/request | Default / disabled peak physical footprint MiB |
|---|---:|---:|---:|
| wasmvm / Go | 150.84 | 145.38 | 112 / 656 |
| wasmvm / experimental LLGo | 364.91 | 302.90 | 179 / 909 |
| Wasmtime control | 25.46 | — | 14 |

All 15 finite processes and 46,080 measured responses pass. Their windows contain 42,240 checked requests. The short-batch cohort has its own parameters and controls. Collector-disabled execution grows memory until process exit, and its cache/page behavior differs from recycling collection. It provides diagnostic evidence about collection costs. Stable tuning acceptance remains open.

Even with collection disabled, the instruction counts remain approximately 1.77 million for Go and 2.27 million for LLGo, against 215,334 for the same-cohort Wasmtime control. The remaining generic work and allocation paths require further changes.

## CPU call stacks

Three fresh Instruments 27 Time Profiler runs use the matched main-source profiles, 30 measured batches and the same full response checks. Each target exits successfully and verifies 122,880 measured responses. Waiting-thread recording is disabled. The analysis window spans the first to last observed request-kernel call, including first/two warmups and 30 measured batches. The sample weights below are approximate CPU attribution normalized by 135,168 checked requests per trace.

| Exclusive sample category | Go sample μs/request | LLGo sample μs/request | Wasmtime sample μs/request |
|---|---:|---:|---:|
| Core execution, VM API and native entry | 45.5 | 59.2 | 1.5 |
| Component calls/tasks, type comparison and canonical values | 40.0 | 67.8 | 7.5 |
| Collector, allocation/free and accounting paths | 27.1 | 155.8 | 2.4 |
| Anonymous/JIT/unresolved leaf | 21.9 | 26.4 | 3.1 |
| HTTP/provider, service harness and other runtime/system paths | 13.0 | 48.9 | 7.6 |

Memory-management frames take priority. Other samples use the nearest owned semantic frame or native Rust component frame. Runtime architecture and inlining affect classification. The broad frame groups support attribution within each implementation. API contracts and compiler inlining differ across implementations. Sample weights and process-counter measurements retain independent observations.

Most owned anonymous leaves return through `wasmvm_enter`, accounting for approximately 21.5 μs/request under Go and 26.0 μs under LLGo. Individual guest-function identities remain unresolved in these traces. Wasmtime's anonymous/deduplicated leaves also remain explicit. Native frame adoption, entry state, global/memory binding and Core dispatch have separately resolved frames.

Type-comparison frames are reached predominantly from `asyncEnd.matches` and `Instance.end`. They account for approximately 14 μs/request under Go and 13 μs under LLGo in these traces. The existing bounded schema-cache experiment addresses this category. Its recorded throughput qualification remains open.

## Actual dispatch counts

A temporary counter-only copy executes 1,000 checked requests under both Go and experimental LLGo. The two compiler profiles produce identical per-target counts. Instrumented timing is excluded.

| Event | Count/request |
|---|---:|
| Native entries | 186 |
| Native Call exits | 99 |
| Native Return exits | 87 |
| Core Run boundaries | 93 |
| Host callbacks | 81 |
| Cross-Core-instance guest calls | 18 |
| Component tasks created | 9 |

The observed native exits contain Call and Return. Poll and Move counters are zero in this workload. This leaves cancellation-budget and large parallel-move changes outside the selected first optimization.

A further Go counter probe identifies the 81 host callbacks.

| Callback group | Count/request |
|---|---:|
| Canonical context.get | 13 |
| Canonical context.set | 24 |
| task.return | 1 |
| Waitable builtins | 19 |
| Async value builtins | 17 |
| Generic canonical lower | 7 |

Plain `resourceBuiltin` counters are zero. A tentative resource-new/rep/drop attribution was disproved by the opcode counters and retained as a diagnostic result. The 37 context accesses currently pass through ordinary `vm.HostFunction`, Caller lifetime handling and scalar value-array conversion.

## Wasmtime mechanism

The pinned Wasmtime translator resolves canonical context.get/set to its context intrinsics. Its Cranelift compiler emits a load/store of the corresponding `VMStoreContext` context slot. The installed Cargo source bytes match GitHub's v48.0.1 file blobs. [Translator](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/environ/src/component/translate/inline.rs#L1074), [native lowering](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/cranelift/src/compiler/component.rs#L2368).

The reference also uses generated typed Service bindings and cached export lookup. Its input is a reused `Bytes` value exposed through `http_body_util::Full`. The owned HTTP adapter accepts Go `http.Request`/`io.Reader` and uses bounded off-Store workers for blocking reads and related completion work. These host adapter policies affect end-to-end costs. They remain explicit in interpreting the same-guest comparison.

The ordinary HTTP controls expose further host-runtime cost. In the long counter cohort, Go records 1.78 CPU μs/request with 0.27 system CPU μs, while experimental LLGo records 50.72 CPU μs with 44.29 system CPU μs. These controls include the shared bounded queue, HTTP work and result checking. They support a separate channel/synchronization investigation. The attribution of P3 system CPU requires its own traced stacks.

## Selected implementation order

1. Lower canonical context.get/set through the owned compiler/runtime contract. The interpreter and native backend must share the same checked task-storage semantics. Native code should access Store-owned scalar storage directly. Support both slots and I32/I64, preserve active-task/failure checks, save/restore across task switches, stackful reentry, suspension, cancellation, panic and shutdown. Maintain the existing fuel, frame and lifetime contracts. The first structural targets are 37 fewer generic host callbacks and 37 fewer native Call exits for this fixture. Throughput improvement requires a measured source comparison.
2. Address the 18 cross-instance guest calls with explicit per-frame instance/runtime-data ownership and code lifetime. The same-instance native linking from PR156 provides a base for this work.
3. Specialize canonical conversions and type handling from checked plans. Preserve dynamic nominal identities, validation budgets and public host-function ownership contracts. Measure conversion/copy work as well as allocation counts.
4. Isolate LLGo channel/synchronization costs before filing an additional runtime issue. Existing allocation-counter and capture-placement contributions retain their separate scope.

The follow-up context implementation passes local dual-compiler/platform
validation; its full hosted matrix remains pending. Performance parity and
cross-target release acceptance remain open.

## Evidence handling

The first xctrace-managed stdout capture produced a ready line without a complete result record, despite a successful target exit. That initial trace remains unqualified. Final matched traces use an exec wrapper with direct file redirection and verify each target exit/result. Machine frequency and neighbor load remain uncontrolled. Throughput acceptance uses independent unprofiled processes. All raw observations, source/toolchain identities and checksums are retained.

[Compact observations and provenance](data/performance-p3-cost-attribution-2026-11-11.json)
