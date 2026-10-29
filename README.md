# wasmvm

A WebAssembly virtual machine written in Go, designed to build and run with both Go and LLGo.

**M1 is in progress.** The [embedding API](docs/embedding.md) now provides Engine/Store/module lifetimes, typed host imports and reference roots. The project-owned Core 2.0 frontend, typed IR, slot interpreter and private runtime state pass all **51,139 binary assertions in 148 pinned scripts** under Go and LLGo on Linux amd64 and macOS arm64. This includes scalar/fixed-SIMD execution, calls, references, tables, memory, linking, initialization and expected traps. The 1,091 malformed-text parser assertions are classified separately.

The Store reclaims unreachable failed instances and releases state after active calls leave on close, preserving imported effects and escaped references. The [core CLI](docs/cli.md) supports validate, inspect and typed run commands. Own-engine phase benchmarks separate startup, compilation, instantiation, first call and warm execution; final M1 acceptance remains open. The development native backend now executes the same complete Core 2.0 corpus under both host compilers on the required targets through the shared runtime, with all 236 fixed-SIMD operations generated natively on amd64 and arm64. Mutable state/bulk operations use explicit runtime helpers; general calls use the shared frame driver. CPU selection preserves SSE2 portability and adds OS-validated AVX2 and short-vector AVX-512 capabilities. A capable native AMD host now executes both new AVX-512 paths; calibrated performance acceptance remains open. Exact SIMD immediates and direct shuffle operands reduce materialization/copy work. Each Store retains one bounded idle call workspace, releasing former frames, memory bindings and cancellation registrations before reuse. The shared interpreter reads operands and immutable metadata directly, avoiding per-instruction closure allocation and aggregate copies. Pure SIMD uses borrowed instructions and value lane updates; packed integer arithmetic isolates carries and byte popcounts. Native performance acceptance remains open. The opt-in final Exception Handling, WasmGC, Typed Function References, Memory64/Table64, Relaxed SIMD, Multi-Memory, Tail Call and Extended Constant Expressions features work on both backends. The complete pinned Core 3.0 binary profile is selectable with `--profile core3`; Core Threads is separately enabled with `--threads`, with same-Engine shared-memory capabilities, all atomic instructions and wait/notify on both backends. The P1 host adapter now implements all 46 imports on Linux/macOS, with explicit process capabilities and a complete 72-program official corpus. Legacy wasi-threads, components and WASI 0.2/0.3 remain subsequent work. Separate evaluation modules keep existing engines as semantic/performance references.

## Goals

- Build a project-owned Go decoder, validator, fast portable interpreter and native compiler, with complete Core 2.0 first and Core 3.0 next, including WasmGC, final exception handling, SIMD, Memory64 and Table64.
- Implement Core Threads shared linear memory, atomics, and wait/notify, with a separate legacy `wasi-threads` adapter for existing LLGo guests.
- Support WASI Preview 1, Component Model-based WASI 0.2, and WASI 0.3 with native async as separate, versioned targets.
- Run the same source and semantic tests under `go test` and `llgo test`. Building the host VM with LLGo and executing LLGo-produced Wasm guests are separate acceptance targets.

The research snapshot is dated **2026-10-26**: WASI 0.3 has shipped, GC is part of Core 3.0, and Core Threads remains distinct from Shared-Everything Threads and shared GC objects. The Go ecosystem already includes wazero and Wago, which overlaps substantially with this project's goals. See the [standards report](docs/research/standards.md) and [implementation comparison](docs/research/implementations.md) for evidence and limitations.

## Documentation

| Document | Contents |
| --- | --- |
| [Standards and proposals](docs/research/standards.md) | Proposal stages, Core/WASI versions, official sources, and conflicting or outdated material |
| [Existing implementations](docs/research/implementations.md) | Runtime capability matrix and Wago/wazero reuse assessment |
| [Wazero evaluation](docs/research/wazero-evaluation.md) | Actual feature and compiler results, known gaps, and reproducible experiments |
| [Own-core decision](docs/design/own-core.md) | Selected implementation ownership, reference-engine role and acceptance boundaries |
| [Core compiler](docs/design/core-compiler.md) | Typed IR, interpreter dispatch, optimization, native code generation and semantic preservation |
| [Core Threads](docs/design/core-threads.md) | Shared memory, synchronization, wait/notify, quotas and ownership |
| [Native ABI](docs/design/native-abi.md) | Go/LLGo bridge, guest frames, roots, continuations and executable-code lifecycle |
| [Performance contract](docs/design/performance-contract.md) | Subsystem measurements, own base/head comparisons and required performance targets |
| [Performance comparison](docs/research/performance.md) | Matched startup/small/sustained CLI workloads, raw samples, and measurement limits |
| [Architecture](docs/design/architecture.md) | Execution, types, GC, threads, WASI, isolation, and draft embedding API |
| [Package layout](docs/design/packages.md) | Public `pkg` APIs, internal layers, dependency direction, and implementation order |
| [Go/LLGo and CI](docs/design/llgo-ci.md) | Compiler constraints, pinned toolchains, and host-versus-guest testing |
| [Development plan](docs/roadmap.md) | Milestones, work packages, dependencies, deliverables, acceptance gates, and initial backlog |
| [Validation record](docs/validation.md) | Actual local and CI results, with their scope and limitations |
| [Core 2.0 input inventory](testdata/core2/README.md) | Complete pinned scripts, binary/text assertion counts, producer hashes and regeneration |

The owner selected a **project-owned core** after M0. Wazero and other engines remain independent semantic/performance references. Native compilation is a required early milestone. The interpreter and compiler logic remain usable without cgo; standard-Go native execution initially uses a small C ABI bridge. See the [decision](docs/design/own-core.md) and [historical reuse assessment](docs/design/reuse-decision.md).

## Development phases

| Phase | Outcome | Status |
| --- | --- | --- |
| Design baseline | Research, architecture, package contracts and compiler probes | Complete; own-core design now replaces reuse-first direction |
| M0 | Candidate evaluation and reproducible comparison infrastructure | Complete through PRs #1–#3 |
| M1 | Complete Core 2.0 interpreter, validation, resource limits, API and core CLI | In progress: complete binary corpus and embedding API implemented; own-engine measurement and final acceptance remain |
| M2 | Fast interpreter and complete Core 2.0 native compiler for amd64/arm64 | In progress: complete native semantic corpus; call/helper optimization and performance acceptance remain |
| M3 | Full Core 3.0 and separate Core Threads on both backends | Full Core 3.0 and Core Threads implemented; cross-platform validation and performance acceptance remain |
| M4 | P1, legacy wasi-threads and pinned Go/LLGo guest integration | P1 implemented; legacy threads and threaded LLGo guests remain |
| M5 | Component Model and WASI 0.2 by declared world/interface | Planned |
| M6 | WASI 0.3 native async and cumulative component features | Planned |
| M7 | Further optimization and separately gated experimental proposals | Deferred |

Follow the [full plan](docs/roadmap.md) in order. A small instruction subset or passing microbenchmark does not complete a stage. Benchmark jobs compare exact wasmvm base/head sources and independent candidate engines. The design's performance budgets are requirements to validate, not measured achievements.

## Local validation

```sh
python3 scripts/core2_corpus.py
go test -count=1 ./...
llgo test -O3 -lto=full -count=1 ./...
# Equivalent entry points that select the host compiler:
./scripts/check.sh go
./scripts/check.sh llgo
```

The root tests cover full module/instruction decoding and typing, source ownership, section ordering, resource limits, all pinned binary inputs, LEB128 width/sign/padding rules, UTF-8 names, host numeric bit patterns, reference reachability, goroutines, locks, atomics, condition variables, and cancellation. CI runs both compilers on Linux amd64 and macOS arm64, checks each executable core package's statement coverage against 90%, and fuzzes binary primitives and full module decoding and validation with Go. The [wazero evaluator](experiments/wazero/README.md) is a separate module with 13 pinned capability fixtures, negative-path tests, shared-memory concurrency checks, and a bounded cancellation observation. Its owned-code coverage is measured separately from engine conformance. Expected rejection of an unsupported fixture is an assessment result, not feature support.

Module path: `github.com/cpunion/wasmvm`. A project license has not yet been selected. No third-party implementation code is imported at this stage. The generated Core 2.0 tests retain their upstream Apache-2.0 notices; see [input provenance](testdata/core2/README.md#provenance-and-notices).
