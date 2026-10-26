# Wazero Reuse Evaluation

Evaluation date: **2026-10-26**. This report separates upstream source evidence, executed candidate probes, and requirements that remain for M1–M5. The M0 artifacts include candidate comparisons, an LLGo guest inventory, a compiler compatibility contribution, and a reproducible performance baseline. They do not constitute full engine conformance. This supplements the [implementation survey](implementations.md) and informs the [reuse decision](../design/reuse-decision.md).

## Conclusion

If an existing engine satisfies the required features, Go/LLGo execution, resource/lifecycle contracts, and maintenance constraints, there is no reason to implement another VM core. `wasmvm` could instead provide integration, compatibility, additional WASI adapters, and reproducible acceptance tests.

**Upstream wazero is not currently a complete drop-in for this project's declared scope.** It is nevertheless a strong reuse candidate. It already provides Core 1.0/2.0, an interpreter, P1 hosting, and experimental threads/final EH/typed references. The remaining requirements include substantial missing Core and component functionality, an LLGo build boundary, and a reproduced blocked-wait cancellation limitation. These gaps do not by themselves justify rewriting the decoder, validator, interpreter, and P1 implementation.

The selected next path is reuse of the wazero interpreter, fixing compiler compatibility in LLGo and testing missing features separately. No independent VM core is justified by the measured gaps. The choice remains subject to the later milestone contracts; existing capabilities may be reused only after their required acceptance tests pass.

## Exact upstream inputs

| Input | Observed revision or status | Evidence |
| --- | --- | --- |
| Latest stable release | `v1.12.0`, published 2026-05-29; commit `2ab480b55fa408d6b35df97fe32a60d08bd6e201` | [Release](https://github.com/wazero/wazero/releases/tag/v1.12.0), [commit](https://github.com/wazero/wazero/commit/2ab480b55fa408d6b35df97fe32a60d08bd6e201) |
| Main branch snapshot | `ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7`, 2026-10-19 | [Commit](https://github.com/wazero/wazero/commit/ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7) |
| GC implementation proposal | PR #2494 is open and unmerged; head `d50c79c09032f4d1f456ec8f70eba652c115a17e` | [PR](https://github.com/wazero/wazero/pull/2494) |
| Core 3.0 tracking | Issue #2426 remains open | [Tracker](https://github.com/wazero/wazero/issues/2426) |
| Go module path | `github.com/tetratelabs/wazero`, despite the repository moving to `wazero/wazero` | [Stable go.mod](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/go.mod) |
| Minimum declared Go version | `go 1.25.0`; dependency `golang.org/x/sys v0.44.0` in stable and the inspected main snapshot | [Stable go.mod](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/go.mod), [main go.mod](https://github.com/wazero/wazero/blob/ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7/go.mod) |
| License | Apache-2.0 | [Pinned LICENSE](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/LICENSE) |

The module's Go requirement must be accounted for when adding it to this repository or an evaluation submodule; the original `go 1.24.0` scaffold declaration cannot silently serve as a validated minimum for this dependency.

## Capability findings

Stable and the inspected main snapshot expose the same experimental switches: Threads, TailCall, ExtendedConst, ExceptionHandling, and TypedFunctionReferences. They do not expose a CoreFeaturesV3 preset or a GC feature flag. Enable experimental features explicitly and report them separately from default Core 2.0 behavior. Sources: [stable features](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/api/features.go), [stable experimental features](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/experimental/features.go), [main experimental features](https://github.com/wazero/wazero/blob/ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7/experimental/features.go).

| Requirement | Upstream evidence | Reuse implication |
| --- | --- | --- |
| Core 1.0/2.0 and SIMD | Both interpreter and compiler are documented as passing their supported Core 1.0/2.0 suites | Reuse the implementation and independently validate the selected profile |
| Tail calls and extended constants | Experimental flags present in stable and main | Enable explicitly; run the relevant fixtures and depth/resource cases |
| Final EH and typed function references | Included in v1.12.0 behind experimental flags; interpreter contains `try_table` and `throw_ref` paths | Reuse candidate; test GC reachability, host boundaries, and backend combinations |
| Shared linear memory and Core Threads | Experimental support, including atomics and wait/notify | Useful foundation; it does not provide the complete process/thread lifecycle required here |
| Legacy `wasi.thread-spawn` | No built-in adapter found; official example creates child instances directly from host goroutines | Implement or reuse an explicit ABI adapter, including thread exit and rollback semantics |
| WasmGC | Absent in stable/main; open PR targets interpreter only | A substantial remaining feature; an unmerged PR is not a support commitment |
| Memory64, multi-memory, relaxed SIMD | Still uncompleted in the Core 3.0 tracker; absent from the inspected feature set | Requires upstream work or a separately validated reusable implementation |
| WASI P1 | Built-in `wasi_snapshot_preview1` host implementation | Reuse, with operation/capability and lifecycle testing for the chosen platform scope |
| Component Model and WASI 0.2 | No built-in component runtime/P2 host layer found in the inspected tree | A component layer can potentially be added above a reusable core |
| WASI 0.3 host interfaces | No built-in implementation found | Additional versioned async/component and host-interface work remains |
| LLGo host compiler | Released LLGo v1.0.8 fails at linking; the LLGo runtime contribution runs the unchanged wazero evaluator | Fix the runtime contracts in LLGo; keep compiler/native-backend support separate from interpreter acceptance |

References: [README and backend policy](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/README.md), [release notes](https://github.com/wazero/wazero/releases/tag/v1.12.0), [Core 3.0 checklist](https://github.com/wazero/wazero/issues/2426#issuecomment-3538706451), [P1 host exports](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/imports/wasi_snapshot_preview1/wasi.go), [repository tree](https://github.com/wazero/wazero/tree/ef2dd2f85376c1a6ff79baae0cc1a5aa42c91db7).

Missing implementation evidence is scoped to the inspected upstream tree. It is not a claim that no third-party layer or fork exists. Source support is also not a claim that this project reran all official conformance suites.

## Threading and cancellation details

The [official threads example](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/experimental/features_example_test.go) uses host goroutines, shared imported memory, per-child module instantiation, explicit TLS initialization, and stack allocation. It does not register the `wasi.thread-spawn` import or exercise a guest-driven spawn protocol. Its libc-layout assumptions and finalizer-based cleanup are examples to inspect, not a production thread-group contract to copy unchanged.

The [experimental feature notes](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/experimental/features.go) also distinguish guest atomics from host Memory API access. Mixed host/guest access, ordinary shared loads/stores, memory growth, and termination need their own tests. A successful atomic increment does not prove every shared-memory combination.

### Reproduced bounded-wait cancellation gap

An independent local probe used the stable interpreter with `WithCloseOnContextDone(true)`, a 20 ms call-context deadline, and the following guest:

```wat
(module
  (memory 1 1 shared)
  (func (export "run") (result i32)
    i32.const 0
    i32.const 0
    i64.const 1000000000
    memory.atomic.wait32))
```

The wait timeout is one second. The measured call returned after **1.000794333 seconds**, with result `[2]` and `module closed with context deadline exceeded`. Cancellation was reported after the wait's own timeout, rather than promptly at the context deadline. This is one bounded diagnostic observation, not a statistical latency guarantee.

The [wait implementation](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/internal/wasm/memory.go#L377) receives no context and selects only between notification and the guest timeout. A negative timeout waits solely on its ready channel. The inspected main snapshot retains this shape. This supports a risk of an indefinitely blocked wait surviving cancellation, but an infinite-wait run was deliberately not used as evidence. The required acceptance contract must cover deadline, close, notify, and process-exit races with deterministic cleanup.

The probe was generated with `wasm-tools parse wait.wat -o wait.wasm` and run with `go run -mod=mod .` from a temporary directory, using the project's local Go environment. Temporary probe sources are supporting evidence; the repository's committed evaluator should remain the reproducible entry point for ongoing validation.

## Host compiler findings

With released LLGo v1.0.8 on darwin/arm64, unchanged wazero v1.12.0 fails at native linking on `runtime.memmove` and `runtime.getAuxv`. Selecting `NewRuntimeConfigInterpreter()` does not remove all native-backend dependencies from the package graph. A temporary backend-isolation patch localized the problem, but it is not the selected production solution.

The symbol inventory and minimal reproducer were reported in [LLGo issue #2918](https://github.com/xgo-dev/llgo/issues/2918). The active stable dependency graph references `runtime.memmove`, `runtime.nanotime`, and `runtime.getAuxv`, plus seven syscall hooks and two wazero-owned assembly entry points. No other missing runtime symbol was demonstrated in that stable graph. Wazero main additionally references `runtime.memclrNoHeapPointers`, which is also missing from the inspected LLGo main. This is a scoped inventory, not an audit of every Go runtime symbol.

`getAuxv` is already fixed for non-Linux platforms in upstream LLGo commit `c0825fed6`; it is a released-version limitation, not a new unresolved main-branch defect. The remaining memory-hook contribution is [LLGo PR #2919](https://github.com/xgo-dev/llgo/pull/2919), at `bc7c991d8fe91459c8afc1eb7d0e374124d95374`. It preserves overlapping-copy behavior, nil/zero handling, and the required indivisibility of aligned pointer-word writes. The contribution is an unreleased compiler candidate; opening its PR does not mean upstream has merged or released it.

The project's CI pins that compiler candidate and requires execution of the **unchanged** wazero evaluator under both Go and LLGo. It does not count an expected linker failure as a successful compatibility run. The original failure assertion and backend-isolation scripts remain available only to reproduce the v1.0.8 investigation.

The interpreter itself still uses `unsafe` reference conversions. Final EH stores exception references through integer slots, and function references also cross `unsafe.Pointer`/`uintptr` representations. See the [interpreter](https://github.com/wazero/wazero/blob/2ab480b55fa408d6b35df97fe32a60d08bd6e201/internal/engine/interpreter/interpreter.go). This is not proof of a bug, but it requires explicit host-GC lifetime testing; the phrase “pure Go” does not establish compiler-neutral runtime semantics.

The [open GC PR](https://github.com/wazero/wazero/pull/2494) proposes host-GC-backed struct/array objects and tagged references. Its representation, collector visibility, host interoperation, and optimized-backend exclusion must be evaluated separately. Do not copy the proposal or mark GC complete solely because its instruction tests pass.

## Executed evaluation and comparison artifacts

The [standalone evaluator](../../experiments/wazero/README.md) pins v1.12.0 and 13 project-authored WAT/Wasm fixtures with source/binary hashes. Seven fixtures execute: arithmetic/traps, P1 output, host callbacks, final EH, typed references, shared-memory operations, and an imported-memory worker. Six record exact capability rejections: GC, Memory64, multi-memory, relaxed SIMD, component binaries, and the absent legacy thread-spawn import. A rejected fixture is not feature support.

Additional tests execute four workers with 512 atomic increments, shared memory growth, bounded wait/notify, manifest corruption, invalid calls, import conflicts, and unexpected errors. Go race checks pass. The owned evaluator has **100% Go statement coverage** locally; CI enforces at least 90%. This percentage excludes the upstream engine, compiler, benchmark runner, and specification coverage. [Local evidence](../../experiments/wazero/results/darwin-arm64.json) preserves the observations, toolchain identities, and cancellation timings.

The [Wago comparison](../../experiments/wago/README.md) pins `v0.1.0-beta.13`. Its Go arithmetic/trap/host-callback probes pass, while LLGo v1.0.8 fails before execution on an unsupported ARM64 `ADR` assembly instruction. No Wago GC/thread/WASI conformance is claimed. This bounded experiment identifies a larger native-backend adaptation boundary than the missing wazero runtime hooks; it is sufficient for the initial reuse choice, not a verdict that Wago cannot be adapted.

The [LLGo guest experiment](../../experiments/llgo-guest/README.md) preserves source, a generated module, hashes, actual imports/instructions, and a successful Wasmer reference run. It exercises a reference cycle, guest GC, goroutine synchronization, panic/recover, and output. It requires shared `env.memory`, legacy `wasi.thread-spawn`, P1 imports, standard EH, and atomic waits. The named `-target wasi` packaging failure is recorded separately from the working `GOOS=wasip1 GOARCH=wasm` producer path. This guest has not yet been made runnable through a complete wazero threads adapter.

The [performance comparison](performance.md) supplies matched fixtures and raw startup-inclusive measurements for wazero's interpreter/native backends, Wasmtime, Wasmer, and WasmEdge. Those measurements inform the reuse decision without replacing feature or lifecycle acceptance.

## Optional follow-up: a wazero-derived implementation

The audit also identified [samyfodil/wazy](https://github.com/samyfodil/wazy), an Apache-2.0 runtime derived from wazero. At main commit `97e115a42152668d663a6b633e2a85e16ca894fe` (2026-10-25), its [feature definitions](https://github.com/samyfodil/wazy/blob/97e115a42152668d663a6b633e2a85e16ca894fe/api/features.go) include CoreFeaturesV3 with GC, multi-memory, Memory64, and relaxed SIMD. Its [README](https://github.com/samyfodil/wazy/blob/97e115a42152668d663a6b633e2a85e16ca894fe/README.md) describes component execution, P2 hosting, and the component async ABI.

That README also explicitly distinguishes the async ABI from missing P3 host-interface worlds. Some timeline language is stale relative to the project's standards snapshot, so capability decisions must follow pinned source and actual tests. Component cooperative threads do not establish legacy wasi-threads support. No Wazy build, LLGo run, conformance suite, or benchmark was executed in this audit; it is a follow-up reuse candidate, not a validated replacement.

## Maintenance, licensing, and contribution rules

- Prefer public API integration. Most decoder/validator/interpreter implementation is under `internal`; extracting it or changing its engine requires an upstream contribution, controlled fork, or explicit code import rather than an ordinary package import.
- Preserve Apache-2.0 licensing and applicable notices for any copied or modified assets. Check test-corpus and generated-fixture licenses independently; the engine license does not automatically cover every external input.
- Record every compatibility patch against a pinned revision, its rationale, tests, and upstream disposition. A patch that silently disables the default backend must be presented as an explicit interpreter-only build configuration.
- For repositories outside the allowed owner boundary, contribute only through a permitted fork and contribution PR. Do not push directly to wazero/wazero or merge its PRs. Any wasmvm change follows the user's PR/checks/merge workflow in `cpunion/wasmvm`.
- Track test coverage of this project's evaluator and integration code separately from upstream engine feature/conformance coverage. High harness statement coverage cannot certify a VM feature that its probes intentionally reject.

## Remaining implementation acceptance

M0 records the reuse choice and measured risks; M1–M5 still require their entire declared suites and APIs. Full Core/WASI conformance, LLGo native JIT compatibility, GC/host-root lifetime stress, legacy thread lifecycle, prompt blocked-wait cancellation, and the combined LLGo guest remain outstanding. None is waived because the evaluator passes or its own statement coverage is high. The LLGo contribution's upstream review/release status must also remain visible when selecting a supported compiler version.
