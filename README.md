# wasmvm

A WebAssembly virtual machine written in Go, designed to build and run with both Go and LLGo.

**M1 implementation has started; no production VM API is available yet.** The owned frontend decodes the full Core 2.0 binary grammar, including all fixed SIMD encodings, with bounded input ownership, allocations and nesting. Shared Go/LLGo tests check all 4,581 corpus binaries, including 719 malformed cases. The validator checks all Core 2.0 static semantics and rejects all 2,146 semantic-invalid corpus modules. Typed SSA lowering and its dominance/type verifier cover all 6,236 defined functions in the statically valid corpus. Typed slot allocation and branch-move lowering also cover those functions. All 136 scalar numeric operations pass 11,621 direct official numeric assertions under Go and LLGo. The complete interpreter/runtime remains unfinished. `pkg/wasm` now exposes value-type identifiers and frontend error categories; the VM API is still reserved. Separate evaluation modules execute real Wasm through existing engines and report their limitations explicitly.

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
| M1 | Complete Core 2.0 interpreter, validation, resource limits, API and core CLI | In progress: full binary decoding, static validation, typed IR/slots and scalar numeric primitives; full execution remains open |
| M2 | Fast interpreter and complete Core 2.0 native compiler for amd64/arm64 | Planned; required performance gate |
| M3 | Full Core 3.0 and separate Core Threads on both backends | Planned |
| M4 | P1, legacy wasi-threads and pinned Go/LLGo guest integration | Planned |
| M5 | Component Model and WASI 0.2 by declared world/interface | Planned |
| M6 | WASI 0.3 native async and cumulative component features | Planned |
| M7 | Further optimization and separately gated experimental proposals | Deferred |

Follow the [full plan](docs/roadmap.md) in order. A small instruction subset or passing microbenchmark does not complete a stage. Existing benchmark jobs compare candidate engines; actual wasmvm base/head execution will be added in M1. The design's performance budgets are requirements to validate, not measured achievements.

## Local validation

```sh
python3 scripts/core2_corpus.py
go test -count=1 ./...
llgo test -count=1 ./...
# Equivalent entry points that select the host compiler:
./scripts/check.sh go
./scripts/check.sh llgo
```

The root tests cover full module/instruction decoding and typing, source ownership, section ordering, resource limits, all pinned binary inputs, LEB128 width/sign/padding rules, UTF-8 names, host numeric bit patterns, reference reachability, goroutines, locks, atomics, condition variables, and cancellation. CI runs both compilers on Linux amd64 and macOS arm64, checks each executable core package's statement coverage against 90%, and fuzzes binary primitives and full module decoding and validation with Go. The [wazero evaluator](experiments/wazero/README.md) is a separate module with 13 pinned capability fixtures, negative-path tests, shared-memory concurrency checks, and a bounded cancellation observation. Its owned-code coverage is measured separately from engine conformance. Expected rejection of an unsupported fixture is an assessment result, not feature support.

Module path: `github.com/cpunion/wasmvm`. A project license has not yet been selected. No third-party implementation code is imported at this stage. The generated Core 2.0 tests retain their upstream Apache-2.0 notices; see [input provenance](testdata/core2/README.md#provenance-and-notices).
