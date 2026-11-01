# wasmvm

A WebAssembly virtual machine written in Go, designed to build and run with both Go and LLGo.

The declared functional implementation covers Core 2.0/3.0, Core Threads,
components and WASI 1/2/3 through the project-owned interpreter and amd64/arm64
native compiler. Go and pinned LLGo run the same semantic tests on Linux amd64
and macOS arm64. No backend silently substitutes a competitor.

- **Core:** all 51,139 binary assertions in 148 pinned Core 2.0 scripts and
  61,395 in 259 Core 3.0 scripts; text-parser cases are classified separately.
  Both backends implement SIMD, WasmGC, final EH, typed references, tail calls,
  memory64/table64, multi-memory and separate shared-memory atomics/wait/notify.
  Explicit extension switches add Compact Import Section, Wide Arithmetic and
  Custom Page Sizes and experimental non-JS Custom Descriptors;
  see their [snapshot and test scope](docs/design/core-extensions.md).
- **WASI 1:** all 46 declared imports, the complete 72-program upstream inventory,
  legacy wasi-threads' 15-program inventory, and actual Go/LLGo guest integration.
- **Components/WASI 2:** composition, values, resources, Canonical ABI, host module/
  component bindings, local i64 borrows and declared WASI 0.2.12 CLI/HTTP providers.
- **WASI 3:** native async tasks, futures/streams, cancellation, backpressure,
  reentry, host async I/O and all stable WASI 0.3.2 CLI/service/middleware interfaces.
  Independent C/Rust commands and composed Rust HTTP guests execute on both backends.
  A pinned corpus adds 60 unchanged upstream P3 consumers with mixed P2/P3
  imports, plus composed services and streaming/trailer/compression scenarios;
  see its [exact scope](testdata/wasi-preview3/upstream/README.md).
- **Embedding/CLI:** bounded Engine/Store lifetimes, explicit capabilities,
  reference roots, cancellation and joined cleanup; Core validate/inspect/typed
  calls and WASI 1/2/3 commands. See [embedding](docs/embedding.md) and [CLI](docs/cli.md).

All 75 pinned Component Model sources convert. The declared runtime profile
executes 962 selected commands, including 555 guest invocations; 228 optional
component-threading commands are retained separately. See the
[exact inventory](testdata/component-runtime/README.md).

**Acceptance remains open:** expanded WASI world conformance and controlled
performance gates. Functional implementation, selected-corpus execution and
milestone acceptance are distinct. Stack Switching is a mandatory pending
extension; other M7
proposals remain optional.
Current evidence and limitations are in the [plan](docs/roadmap.md) and
[verification record](docs/validation.md). Separate evaluation modules retain
competitor engines as independent references.

## Goals

- Build a project-owned Go decoder, validator, fast portable interpreter and native compiler, with complete Core 2.0 first and Core 3.0 next, including WasmGC, final exception handling, SIMD, Memory64 and Table64.
- Implement Core Threads shared linear memory, atomics, and wait/notify, with a separate legacy `wasi-threads` adapter for existing LLGo guests.
- Support WASI Preview 1, Component Model-based WASI 0.2, and WASI 0.3 with native async as separate, versioned targets.
- Run the same source and semantic tests under `go test` and `llgo test`. Building the host VM with LLGo and executing LLGo-produced Wasm guests are separate acceptance targets.

The latest [standards audit](docs/research/status-2026-11-01.md) and [performance audit](docs/research/performance-2026-11-01.md) are dated **2026-11-01**. WASI 0.3 has shipped, GC is part of Core 3.0, and Core Threads remains distinct from Shared-Everything Threads and shared GC objects. The original October 26 survey and reuse assessment remain historical context. Functional coverage does not imply performance parity: GC, shared-memory and host/component paths still have substantial deficits.

## Documentation

| Document | Contents |
| --- | --- |
| [Current standards and support](docs/research/status-2026-11-01.md) | Every registered proposal, accepted-feature support, competitors and exact test evidence |
| [Current performance gaps](docs/research/performance-2026-11-01.md) | Complete Core portfolio, GC/Threads, P1/P2/async, separate phases and raw evidence |
| [Standards and proposals](docs/research/standards.md) | Proposal stages, Core/WASI versions, official sources, and conflicting or outdated material |
| [Existing implementations](docs/research/implementations.md) | Runtime capability matrix and Wago/wazero reuse assessment |
| [Wazero evaluation](docs/research/wazero-evaluation.md) | Actual feature and compiler results, known gaps, and reproducible experiments |
| [Own-core decision](docs/design/own-core.md) | Selected implementation ownership, reference-engine role and acceptance boundaries |
| [Core compiler](docs/design/core-compiler.md) | Typed IR, interpreter dispatch, optimization, native code generation and semantic preservation |
| [Core Threads](docs/design/core-threads.md) | Shared memory, synchronization, wait/notify, quotas and ownership |
| [Explicit core extensions](docs/design/core-extensions.md) | Compact imports, wide arithmetic, snapshot pins and remaining mandatory extensions |
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
| M3 | Full Core 3.0 and separate Core Threads on both backends | Full Core 3.0 and Core Threads implemented and tested on both target platforms; performance acceptance remains |
| M4 | P1, legacy wasi-threads and pinned Go/LLGo guest integration | P1 and legacy threads/Go/LLGo guests implemented; performance acceptance open |
| M5 | Component Model and WASI 0.2 by declared world/interface | Composition, Canonical ABI including local i64 borrows, host bindings and declared CLI/HTTP providers implemented; expanded world conformance and performance acceptance remain |
| M6 | WASI 0.3 native async and cumulative component features | Native async, host I/O, all stable 0.3.2 providers, declared worlds and independent C/Rust guests implemented; expanded conformance and performance acceptance remain |
| M7 | Further optimization and separately gated experimental proposals | Deferred |

The current owner direction is to complete remaining capabilities before performance optimization. Follow the [full plan](docs/roadmap.md), keeping semantic validation and performance acceptance distinct. A small instruction subset or passing microbenchmark does not complete a stage. Benchmark jobs remain available by workflow dispatch or the `run-benchmarks` PR label; they compare exact wasmvm base/head sources and independent candidate engines. Broad measurements are deferred during capability completion. The design's performance budgets are requirements to validate, not measured achievements.

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
