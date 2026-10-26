# WebAssembly and WASI Standards and Proposal Survey

Research date: **2026-10-26 (Asia/Shanghai)**. This document records primary sources inspected on that date. It does not treat an old roadmap as evidence that a release has shipped, or proposal maturity as evidence of an engine's default support. Priorities and implementation approaches below are this project's design judgments.

The [initial specification and conformance input pins](spec-inputs.md) record
immutable Core/proposal/test-suite revisions, directory-specific license
notices, and independently verified WASI 0.2.12/0.3.2 WIT archive digests.
These are selected inputs for later validation, not claims that the upstream
suites have already passed.

## Findings

Use **Core Wasm 3.0 as the long-term semantic target and WASI 0.1 with legacy wasi-threads as the first practical compatibility target**. Add WasmGC, Component Model / WASI 0.2, and then the already released WASI 0.3 through separate acceptance milestones. Account for GC and multithreading in the initial data model, but validate them independently. Initial support for GC modules and shared-linear-memory threads does not promise shared WasmGC objects across threads.

At the research date, GC, typed function references, tail calls, final exception handling, memory64, multi-memory, and relaxed SIMD are included in Core 3.0. Threads remains at Phase 4, Shared-Everything Threads at Phase 1, and Stack Switching at Phase 3. GC should no longer be described as an early experimental proposal, and a shared GC heap should not be described as an already standardized threading capability. Sources: [finished proposals](https://github.com/WebAssembly/proposals/blob/main/finished-proposals.md), [active proposals](https://github.com/WebAssembly/proposals/blob/main/README.md).

WASI 0.3.0 was released on **2026-06-11**. The latest stable patch found was **0.3.2 (2026-10-13)**; the latest 0.2 patch found was **0.2.12 (2026-06-02)**. WASI 0.3 must no longer be listed as unreleased. Sources: [0.3.0 release](https://github.com/WebAssembly/WASI/releases/tag/v0.3.0), [0.3.2 release](https://github.com/WebAssembly/WASI/releases/tag/v0.3.2), [0.2.12 release](https://github.com/WebAssembly/WASI/releases/tag/v0.2.12).

## 1. Three distinct meanings of standardization

| Level | Observation | How this project should record it |
| --- | --- | --- |
| Core proposal phase | Phase 4 hands a proposal to the WG for standardization; Phase 5 records WG consensus on completion, before integration into the main specification | Use individual FeatureSet flags rather than a single latest-version switch |
| Core semantic version | The official announcement declared Wasm 3.0 complete on 2025-09-17; the live specification identifies itself as Release 3.0, dated 2026-10-23 | Pin specification and test-suite commits, and declare the exact feature set that passes |
| W3C publication status | The current TR is a **Candidate Recommendation Draft, 2026-10-23**; version 1.0 from 2019-12-05 is a Recommendation | Do not rewrite Core 3.0 completion as a claim that 3.0 is a W3C Recommendation |

Sources: [phase process](https://github.com/WebAssembly/meetings/blob/main/process/phases.md), [3.0 completion announcement](https://webassembly.org/news/2025-09-17-wasm-3.0/), [current Core specification](https://webassembly.github.io/spec/core/), [current W3C TR](https://www.w3.org/TR/wasm-core/), [W3C publication series](https://www.w3.org/TR/wasm-core/all/). The TR status section also explains its living-standard maintenance model. Candidate Recommendation status and an engine's production readiness are different questions.

**Another common source of confusion:** the Core 3.0 module binary header still uses version `1`, not `3`. Features extend types, instructions, sections, and validation rules. Components have separate binary framing and version rules and should use a separate decoder. Sources: [Core binary module format](https://webassembly.github.io/spec/core/binary/modules.html#binary-module), [Component binary format](https://github.com/WebAssembly/component-model/blob/main/design/mvp/Binary.md).

## 2. Core and embedding proposal matrix

This section covers the finished, active, and inactive entries listed by `WebAssembly/proposals` in this snapshot. Open Phase 0 issues are not a complete set of accepted proposals; monitor them without treating them as committed features.

Proposal phases are pinned to [proposals commit `df6a2aca`](https://github.com/WebAssembly/proposals/tree/df6a2aca1cd9d779eb4f2f3e19dae8277f336015). The active and finished tables were last modified on 2026-10-06. They were cross-checked against the [official feature data](https://github.com/WebAssembly/website/blob/main/features.json), whose observed modification date was 2026-10-21.

### 2.1 Completed features integrated into Core or embedding specifications

All entries below are completed proposals. Versions and affected specifications are recorded in the [official finished table](https://github.com/WebAssembly/proposals/blob/df6a2aca1cd9d779eb4f2f3e19dae8277f336015/finished-proposals.md).

| Feature | Version / layer | Project approach |
| --- | --- | --- |
| MVP | Core 1.0 | Foundation for initial execution and validation |
| Import/Export of Mutable Globals | Core 1.0 | Initial implementation |
| Non-trapping float-to-int | Core 2.0 | Initial implementation; test NaN and saturation boundaries |
| Sign-extension operators | Core 2.0 | Initial implementation |
| Multi-value | Core 2.0 | Initial implementation; support multiple results in frames from the outset |
| Reference Types | Core 2.0 | Initial implementation; make `externref` and `funcref` lifetimes explicit |
| Bulk memory operations | Core 2.0 | Initial implementation, including passive segments and data count |
| Fixed-width SIMD | Core 2.0 | Required before claiming Core 2.0 coverage; scalar interpretation is acceptable initially |
| JS BigInt ↔ Wasm i64 | JS API 2.0 | Not a required host API for a standalone Go VM |
| Tail call | Core 3.0 | Preserve tail calls in the initial IR; ordinary recursion is not a substitute |
| Extended Constant Expressions | Core 3.0 | Implement with initializer validation |
| Typed Function References | Core / JS API 3.0 | Implement with recursive types and the GC type system |
| Garbage collection | Core / JS API 3.0 | Required; a separate WasmGC milestone |
| Multiple memories | Core / JS API 3.0 | Preserve memidx from the outset to avoid a single-memory architecture |
| Relaxed SIMD | Core 3.0 | Follow SIMD, with an explicit determinism profile |
| Custom Annotation Syntax | Text format 3.0 | Follow in WAT tooling; does not block binary execution |
| Branch Hinting | Core custom section 3.0 | May be recognized and ignored without changing execution semantics |
| Exception handling | Core / JS API 3.0 | Implement the final `try_table` / `exnref` form |
| JS String Builtins | Core / JS embedding 3.0 | Track as an embedding extension; implement separately if needed |
| Memory64 | Core 3.0 | Use `uint64` for internal addresses from the outset; gate execution support separately |

### 2.2 Active proposals

Each phase comes from the [pinned active table](https://github.com/WebAssembly/proposals/blob/df6a2aca1cd9d779eb4f2f3e19dae8277f336015/README.md). Proposal links provide the original definitions. Priorities are this project's decisions.

| Proposal | Phase | Project approach |
| --- | ---: | --- |
| [JS Promise Integration](https://github.com/WebAssembly/js-promise-integration) | 5 | JS embedding; a standalone Go VM need not emulate Promises |
| [Web Content Security Policy](https://github.com/WebAssembly/content-security-policy) | 5 | Web embedding |
| [Threads](https://github.com/WebAssembly/threads) | 4 | Required: shared memory, atomics, and wait/notify |
| [Compact Import Section](https://github.com/WebAssembly/compact-import-section) | 4 | Later decoder compatibility, behind a separate feature gate |
| [Wide Arithmetic](https://github.com/WebAssembly/wide-arithmetic) | 4 | Later numeric extension with a relatively isolated implementation boundary |
| [ESM Integration](https://github.com/WebAssembly/esm-integration) | 3 | JS embedding |
| [Stack Switching](https://github.com/WebAssembly/stack-switching) | 3 | Leave room for continuations in frame design; disabled by default |
| [Custom Page Sizes](https://github.com/WebAssembly/custom-page-sizes) | 3 | Avoid scattered hard-coded page shifts; experimental flag |
| [Custom Descriptors and JS Interop](https://github.com/WebAssembly/custom-descriptors) | 3 | Follow GC; outside the first GC support claim |
| [Relaxed dead code validation](https://github.com/WebAssembly/relaxed-dead-code-validation) | 2 | Keep current validation rules; experiment against a pinned snapshot |
| [Numeric Values in WAT Data Segments](https://github.com/WebAssembly/wat-numeric-values) | 2 | Follow in text tooling |
| [Extended Name Section](https://github.com/WebAssembly/extended-name-section) | 2 | Optional debug-information parsing |
| [Rounding Variants](https://github.com/WebAssembly/rounding-mode-control) | 2 | Experimental numeric extension |
| [Compilation Hints](https://github.com/WebAssembly/compilation-hints) | 2 | Ignore hints in the initial interpreter; evaluate for a future JIT |
| [JS Primitive Builtins](https://github.com/WebAssembly/js-primitive-builtins) | 2 | JS embedding |
| [Acquire-Release Atomics](https://github.com/WebAssembly/acquire-release-atomics) | 2 | Separate from existing SC atomics; later experiment |
| [Multibyte Array Access](https://github.com/WebAssembly/multibyte-array-access) | 2 | Evaluate after WasmGC |
| [FP16](https://github.com/WebAssembly/half-precision) | 2 | Experimental SIMD / numeric extension |
| [Type Imports](https://github.com/WebAssembly/proposal-type-imports) | 1 | Avoid tying the type registry to one module; no implementation commitment |
| [Component Model](https://github.com/WebAssembly/component-model) | 1 | Implement the subsets required by released WASI versions; overall Phase 1 status is not a reason to omit them |
| [C and C++ API](https://github.com/WebAssembly/wasm-c-api) | 1 | Start with a Go API; consider a C ABI adapter later |
| [Flexible Vectors](https://github.com/WebAssembly/flexible-vectors) | 1 | Outside the initial `v128` scope |
| [Memory control](https://github.com/WebAssembly/memory-control) | 1 | Current resource limits are embedding features, not an implementation of this proposal |
| [Reference-Typed Strings](https://github.com/WebAssembly/stringref) | 1 | Distinct from WasmGC; track separately |
| [Profiles](https://github.com/WebAssembly/profiles) | 1 | A broader profiles proposal, separate from existing Core 3.0 determinism constraints |
| [Shared-Everything Threads](https://github.com/WebAssembly/shared-everything-threads) | 1 | Reserve shared markers and heap ownership boundaries; outside the stable feature set |
| [Frozen Values](https://github.com/WebAssembly/frozen-values) | 1 | Research after GC |
| [More Array Constructors](https://github.com/WebAssembly/more-array-constructors) | 1 | Research after GC |
| [JIT Interface](https://github.com/WebAssembly/jit-interface) | 1 | Not a prerequisite for implementing this VM's internal JIT |
| [Type Reflection for JS API](https://github.com/WebAssembly/js-types) | 1 | JS embedding |
| [JS Text Encoding Builtins](https://github.com/WebAssembly/js-text-encoding-builtins) | 1 | JS embedding |

Phase 5 entries that remain in the active table have completed the WG process but have not yet been integrated into the main specification. Preserve this distinction from entries in the finished table; not every active-table entry is necessarily unstable.

### 2.3 Inactive or superseded proposals

Source: [pinned inactive table](https://github.com/WebAssembly/proposals/blob/df6a2aca1cd9d779eb4f2f3e19dae8277f336015/inactive-proposals.md).

| Proposal | Status / successor | Project decision |
| --- | --- | --- |
| Unmanaged closures | Superseded by Typed Function References | Do not implement the old design |
| Conditional sections | Redirected to Feature Detection, which was later withdrawn | Do not implement the old design |
| Feature Detection | Withdrawn for lack of consensus | Use a host capability API |
| Module Linking | Redirected to Component Model | Follow the Component Model |
| Interface Types | Redirected to Component Model | Follow WIT / Canonical ABI |
| Funclets | Redirected toward multiloop work | Do not adopt the old encoding |
| Call Tags | Inactive | Monitor |
| Constant Time | Inactive | Do not claim constant-time execution |
| JS Customization for GC Objects | Superseded by Custom Descriptors | Do not implement the old design |
| Instrument and Tracing Technology | Inactive | Use project debugging APIs without claiming proposal compatibility |

Record **legacy EH** explicitly. Official feature data marks the old exception format as inactive, and the final EH explainer links the old documentation separately. The old `try/catch/rethrow/delegate` form is different from Core 3.0 `try_table`, `throw_ref`, and `exnref`. Sources: [final EH documentation](https://github.com/WebAssembly/exception-handling/blob/main/proposals/exception-handling/Exceptions.md), [legacy EH documentation](https://github.com/WebAssembly/exception-handling/blob/main/proposals/exception-handling/legacy/Exceptions.md). If an actual llgo artifact uses the old encoding, provide an explicit compatibility feature and separate tests rather than an undifferentiated EH support claim.

## 3. WASI and the Component Model

### 3.1 Released versions and API scope

| Version | Status in this snapshot | Binary / ABI | Engineering target |
| --- | --- | --- | --- |
| WASI 0.1 / Preview 1 | Legacy, widely deployed | Core modules, `wasi_snapshot_preview1`, WITX | First executable compatibility target |
| WASI 0.2 / Preview 2 | Stable, succeeded by 0.3; latest patch found: 0.2.12 | Components, WIT, resource handles, Canonical ABI, `wasi:io` polling | First component milestone |
| WASI 0.3 / Preview 3 | Current stable series; latest patch found: 0.3.2 | Native `async func`, `future<T>`, and `stream<T>`; removes the `wasi:io` package | Separate async-component milestone |
| WASI 1.0 | No released 1.0 found | Do not use future semantics as current acceptance criteria | Monitor |

Sources: [official 0.1 overview](https://wasi.dev/releases/wasi-p1), [official 0.2 overview](https://wasi.dev/releases/wasi-p2), [official 0.3 overview](https://wasi.dev/releases/wasi-p3), [WASI releases](https://github.com/WebAssembly/WASI/releases). Version 0.1 does not provide general connect/listen or HTTP APIs. Full networking should use the corresponding 0.2/0.3 interfaces, rather than labeling a custom socket extension as standard WASI.

A WASI support claim must specify **version + world/interface + operation coverage + test results**. Validate CLI, clocks/random, filesystem, poll/I/O, sockets, and then HTTP. Running hello world does not demonstrate support for an entire version. Treat 0.2 `wasi:cli/command` and `wasi:http/proxy` separately from 0.3 `wasi:cli/command` and `wasi:http/service` / `middleware`. Sources: [0.2 APIs](https://wasi.dev/releases/wasi-p2), [0.3 release notes](https://github.com/WebAssembly/WASI/releases/tag/v0.3.0).

### 3.2 Component Model Phase 1 and stable subsets are compatible statements

The overall Component Model proposal remains at Phase 1, while the subsets required by released WASI versions carry toolchain compatibility commitments. Pin encoding and semantics by WASI version rather than treating the entire Component Model main branch as a stable ABI. Source: [Component README](https://github.com/WebAssembly/component-model/blob/main/README.md).

| First WASI version requiring the feature | Required Component Model features |
| --- | --- |
| 0.2.0 | The default, ungated Component Model features at that release |
| 0.3.0 | Adds async lift/lower, futures, and streams |
| 0.3.1 | Adds `map<K,V>` and `implements` / `external-id` annotations |
| 0.3.2 | Accumulates the above; no newly adopted features were observed in its release record |

Sources: [formal adoption record](https://github.com/WebAssembly/WASI/blob/main/docs/ComponentModelFeatures.md), [0.3.1 release](https://github.com/WebAssembly/WASI/releases/tag/v0.3.1). These requirements apply to a compatibility claim for the corresponding version even when a particular WASI interface does not yet use every feature.

The following remain gated and must not silently enter the stable default implementation: threading built-ins, built-ins based on Shared-Everything Threads, stackful async lift, fixed-length lists, error-context, canonical interface names, Component memory64, and others. Source: [Component gated features](https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#gated-features). In particular, **completed Core memory64 does not mean memory64 in the Component Canonical ABI is stable**.

### 3.3 WASI API proposal overview

Phases come from [WASI proposals](https://github.com/WebAssembly/WASI/blob/main/docs/Proposals.md), whose observed modification date was 2026-09-08. WASI and Core phase processes do not make identical release commitments; comparing their phase numbers alone is misleading.

| Phase | APIs / extensions in the official list | Project scope |
| --- | --- | --- |
| 3 | Clocks, Random, Filesystem, Sockets, CLI, HTTP | Implement published 0.2 / 0.3 interfaces |
| 2 | Timezone, HTTP Informational Outbound Response, I2C, Key-value Store, wasi-nn, Runtime Config, WebGPU, Messaging | Candidates for optional host plugins, outside initial standard support claims |
| 1 | Blob Store, Crypto, GPIO, Distributed Lock, Logging, Observe, Parallel, Pattern Match, SPI, SQL, SQL Embed, Threads, TLS, URL, USB, OTel | Monitor; distinguish maturity and versions from stable WASI packages |
| 4 / 5 | No entries in the observed table | Do not describe WASI 0.3 as 1.0 or as a final WG standard |
| 0 | No formal entries in the observed repository table | Do not treat scattered website or old-roadmap links as stable APIs |

### 3.4 Documentation discrepancies found during this survey

- The [wasi.dev roadmap](https://wasi.dev/roadmap) still marks 0.3.2, dated 2026-10-13, as Planned, but the [GitHub 0.3.2 release](https://github.com/WebAssembly/WASI/releases/tag/v0.3.2) contains a non-prerelease release and a WIT artifact. This report uses the observed release as evidence.
- The Threads link in the [WASI API list](https://github.com/WebAssembly/WASI/blob/main/docs/Proposals.md) uses `wasi-native-threads`; on the research date it redirected to [legacy wasi-threads](https://github.com/WebAssembly/wasi-threads). The link name is not evidence of a separate, mature native-threading API.
- The [wasi-testsuite README](https://github.com/WebAssembly/wasi-testsuite/blob/main/README.md) still refers to a forthcoming Preview 3. This does not invalidate the existing 0.3 release. Integrate tests from actual directories and pinned revisions rather than relying only on the README.
- Parts of the W3C series overview still mention 2.0, while the current specification body identifies itself as 3.0. This report records semantic version and publication status separately and does not infer a semantic version from `core-2` in a URL.

## 4. Boundaries between GC, threads, and async

### 4.1 Three different meanings of GC

1. **VM host GC:** Go or llgo memory management for the implementation's own Go objects.
2. **Guest linear-memory GC:** a language runtime manages its own heap inside Wasm linear memory; the VM sees bytes.
3. **WasmGC:** Core instruction and type semantics for structs, arrays, i31, recursive types, subtyping, and reference lifetimes.

Implementing the first does not implement the third. Running a guest with its own language GC does not prove that the VM can run WasmGC modules. WasmGC and linear memory are separate storage domains. Sources: [Core 3.0 GC overview](https://webassembly.org/news/2025-09-17-wasm-3.0/), [GC proposal](https://github.com/WebAssembly/gc).

Engineering recommendation: start with a non-moving guest heap using precise mark/sweep. Represent references through checked handle / slot tables with explicit store ownership and generations. Do not encode Go pointers in `uint64` and rely on the host to keep them alive. Operand stacks, frame locals, globals, tables, exception payloads, and host handles are roots. Tail calls, exception unwinding, host callbacks, cancellation, and instance destruction must preserve root lifetimes. Evaluate host objects as backing storage through measurements, but do not substitute the Go collector's scanning rules for Wasm type and lifetime semantics.

### 4.2 Three distinct threading capabilities

| Capability | Specification boundary | Project implementation path |
| --- | --- | --- |
| Core Threads | Shared linear memory, SC atomics, wait/notify; creation and join are embedding concerns | Implement the memory model and wait queues; keep thread lifecycle in a separate layer |
| Legacy wasi-threads | Preview 1 extension; one instance per thread, entering `wasi_thread_start` | Explicitly optional ABI for existing llgo / C / Rust artifacts |
| Shared-Everything Threads | Shared functions, tables, globals, GC references, TLS, GC atomics, and more; still Phase 1 | Reserve shared markers in the design; exclude from the stable feature set |

Sources: [Core Threads overview](https://github.com/WebAssembly/threads/blob/main/proposals/threads/Overview.md), [legacy wasi-threads](https://github.com/WebAssembly/wasi-threads/blob/main/README.md), [Shared-Everything overview](https://github.com/WebAssembly/shared-everything-threads/blob/main/proposals/shared-everything-threads/Overview.md).

Legacy `wasi-threads` has implementation experience in wasi-libc pthreads and Wasmtime / WAMR / Wasmer / toywasm. It remains a legacy Phase 1 extension, not standard WASI 0.2/0.3 pthreads. Test spawn parameters, return values, and child entry against a pinned ABI revision. Sources: [ABI definition](https://github.com/WebAssembly/wasi-threads/blob/main/wasi-threads.abi.md), [implementation list](https://github.com/WebAssembly/wasi-threads#what-has-been-implemented).

Core design requirements:

- Give threads separate execution contexts and stacks while reusing immutable compiled modules. Shared memory and a process-wide WASI context do not imply shared globals, tables, or GC references.
- Route ordinary shared-memory loads/stores and atomic operations through the same memory-access layer. Mixing direct Go `[]byte` access with atomics can introduce host data races. Establish a correct locking or consistent-width strategy before optimizing.
- Specify overlapping access widths, unaligned ordinary accesses, atomic alignment traps, bounds checks, grow, wait timeouts, and notify/cancellation races. Using goroutines is not evidence of compliance with the Threads proposal.
- Checking the expected value and registering a waiter must form a correct protocol that avoids lost wakeups. Exit must wake all waiting threads. Under legacy process semantics, `proc_exit` or a trap in any thread must terminate the entire thread group; bounded execution checkpoints and cancellable host calls must make that effective.
- Child instantiation must not arbitrarily repeat the parent's initialization. Follow the ABI and guest conventions, especially active data segments, start sections, and stack/TLS globals. Add dedicated re-instantiation fixtures.
- Reject cross-thread transfer of non-shared WasmGC references until the relevant experimental Shared-Everything Threads implementation passes its tests. Isolated GC instances may still execute in parallel.

These are project design requirements, not performance assessments of existing engines.

### 4.3 WASI 0.3 async is distinct from OS threads and Stack Switching

WASI 0.3 supplies async tasks, futures, and streams for component calls. Core Stack Switching supplies continuation instructions. Shared-Everything Threads addresses shared objects and threading semantics. They are related but not equivalent. WASI 0.3's stable adoption record does not require every unstable threading or stackful feature. Sources: [adoption record](https://github.com/WebAssembly/WASI/blob/main/docs/ComponentModelFeatures.md), [Stack Switching explainer](https://github.com/WebAssembly/stack-switching/blob/main/proposals/stack-switching/Explainer.md).

Preserve explicit frames, suspension/resumption state, and host-call boundaries in `Execution`. A component async scheduler may initially use Go synchronization primitives, but its external contract must follow the Canonical ABI rather than equating a guest future with an arbitrary goroutine.

## 5. Existing implementations as validation references

This section identifies sources for cross-checking. The detailed engine survey is in [implementations.md](implementations.md).

| Area | Documented implementations / tools | Intended use |
| --- | --- | --- |
| Core 3.0 | Official specification reference interpreter and test suite; Wasmtime; browser engines | Pin fixtures and compare results, traps, and type acceptance/rejection |
| Core threads | Browsers, Wasmtime, WAMR, and others | Regress synchronization and memory semantics; a single run of a racy program is not a sufficient oracle |
| Legacy wasi-threads | wasi-libc with Wasmtime / WAMR / Wasmer / toywasm | Execute pthread fixtures built with a pinned SDK |
| WASI 0.2 | Wasmtime, jco | Cross-check interfaces and component ABI |
| WASI 0.3 | Official pages list Wasmtime 46+ and jco | Pin producers and consumers to the exact WIT version |

Sources: [official engine feature matrix](https://webassembly.org/features/), [spec repository](https://github.com/WebAssembly/spec), [wasi-threads implementation notes](https://github.com/WebAssembly/wasi-threads#what-has-been-implemented), [WASI 0.2](https://wasi.dev/releases/wasi-p2), [WASI 0.3](https://wasi.dev/releases/wasi-p3). Matrix cells may qualify support with flags, backend restrictions, or parsing-only implementations. Confirm those conditions in engine source and configuration before adopting a reference.

## 6. Standards-oriented acceptance milestones

The [development roadmap](../roadmap.md) is authoritative. Following M0 evaluation, the owner selected a project-owned core with native compilation required early. This table maps unchanged standards targets to the revised sequence; it is not a new standards survey. Both host compilers execute the same corpus, with explicit backend/platform evidence.

| Roadmap stage | Complete scope | Acceptance evidence |
| --- | --- | --- |
| M1 | Full Core 2.0 binary frontend/interpreter, including fixed SIMD, references, multi-value and bulk operations | Complete pinned `wg-2.0` corpus, resource/lifecycle/negative tests, own-engine benchmark baseline |
| M2 | Fast interpreter and complete native compiler for that same Core 2.0 profile | Go/LLGo × interpreter/native × Linux amd64/macOS arm64; native ABI and controlled performance gates |
| M3 | Full Core 3.0, including WasmGC, typed refs, tail calls, final EH, Memory64/Table64, multi-memory, extended const and relaxed SIMD; separate Core Threads | Full profile/feature suites on both backends; forced GC, roots, EH, shared-memory and combined-workload tests |
| M4 | P1, legacy wasi-threads and pinned Go/LLGo guest ABI | Operation manifest, capability tests, pthread/TLS/exit, guest GC/panic/I/O integrations |
| M5 | Component Model and WASI 0.2.12 | Cross-language composition, declared worlds/interfaces, ownership and Canonical ABI tests |
| M6 | WASI 0.3.2 and cumulative component async | Declared CLI/HTTP worlds, futures/streams, cancellation/backpressure and lifecycle tests |
| M7 | Separately gated optional proposals/platforms/optimizations | Independent pinned scope, correctness and performance evidence |

The revised sequence completes core mechanisms and performance before WASI integration. A stage cannot close for a small opcode subset, an expected unsupported result, or a passing microbenchmark. See the [compiler design](../design/core-compiler.md) and [performance contract](../design/performance-contract.md).

## 7. Test sources and update policy

- Use [WebAssembly/spec](https://github.com/WebAssembly/spec) and [WebAssembly/testsuite](https://github.com/WebAssembly/testsuite) for Core semantics. For features not integrated into the main specification, such as Threads or Stack Switching, use the corresponding proposal test directories. Record source commits, generation-tool versions, and licenses.
- Use [WebAssembly/wasi-testsuite](https://github.com/WebAssembly/wasi-testsuite) for WASI, with [dedicated legacy-threads tests](https://github.com/WebAssembly/wasi-threads/tree/main/test). Use the [Component test suite](https://github.com/WebAssembly/component-model/tree/main/test) and additional cross-language golden fixtures for components.
- Add **explicit rejection tests** for unimplemented features. Unsupported input must not be accepted and then executed incorrectly. Distinguish unsupported features, malformed input, validation errors, link errors, traps, and WASI exit.
- Report conformance as `compiler × OS/arch × feature profile × spec/WIT revision`. Include unsupported cases, known failures, skip reasons, and review deadlines. A pass rate obtained by excluding large groups of cases is not a compatibility claim.
- Before upgrading a profile, reread the official proposal tables, release tags, WIT, and corresponding test commits. CI must not silently follow main. Preserve observation dates in research records and record the reasons for later changes.
