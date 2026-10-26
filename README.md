# wasmvm

A WebAssembly virtual machine written in Go, designed to build and run with both Go and LLGo.

**M0 evaluation stage: no production VM API is implemented yet.** This repository contains research, architecture, a development plan, host compatibility probes, and executable candidate evaluations. The `doc.go` files reserve package responsibilities; they do not indicate implemented features. The separate evaluation modules execute real Wasm through existing engines and report their limitations explicitly.

## Goals

- Start with a portable, pure Go interpreter and work toward WebAssembly Core 3.0, including WasmGC, standard exception handling, SIMD, and Memory64.
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
| [Reuse decision](docs/design/reuse-decision.md) | Reuse-first direction, remaining work, and milestone acceptance boundaries |
| [Performance comparison](docs/research/performance.md) | Matched startup/small/sustained CLI workloads, raw samples, and measurement limits |
| [Architecture](docs/design/architecture.md) | Execution, types, GC, threads, WASI, isolation, and draft embedding API |
| [Package layout](docs/design/packages.md) | Public `pkg` APIs, internal layers, dependency direction, and implementation order |
| [Go/LLGo and CI](docs/design/llgo-ci.md) | Compiler constraints, pinned toolchains, and host-versus-guest testing |
| [Development plan](docs/roadmap.md) | Milestones, work packages, dependencies, deliverables, acceptance gates, and initial backlog |
| [Validation record](docs/validation.md) | Actual local and CI results, with their scope and limitations |

The independent pure Go interpreter design is a fallback, not an instruction to rewrite an existing engine. **M0 evaluates reuse first.** Current evidence favors resolving LLGo compatibility and retaining wazero's existing implementation while assessing the remaining feature gaps. The diagnostic interpreter-only patch is a localization experiment, not the proposed permanent fix for missing LLGo runtime symbols.

## Development phases

| Phase | Outcome | Status |
| --- | --- | --- |
| Design baseline | Research, architecture, package contracts, shared compiler probes, and CI | Complete |
| M0 | Reuse decision, pinned fixtures, and reproducible candidate experiments | Implemented; acceptance evidence in the contribution PR |
| M1 | Embeddable interpreter, validation, resource limits, and a usable P1 CLI subset | Planned |
| M2a | Core Threads, legacy wasi-threads, and the declared P1 interface scope | Planned |
| M2b | WasmGC, recursive types, typed references, tail calls, and final EH | Planned |
| M3 | Complete the targeted Core 3.0 profile and validate LLGo guests | Planned |
| M4 | Component Model and WASI 0.2, delivered by world and interface | Planned |
| M5 | WASI 0.3 native async, futures, streams, and cumulative component features | Planned |
| M6 | Optional performance work and separately gated experimental proposals | Deferred |

M2a and M2b can proceed in parallel once M1 stabilizes the value, memory, validation, and execution interfaces. LLGo guest compatibility is accepted in M3 because it also requires EH and SIMD. See the [full plan](docs/roadmap.md) for the work breakdown and exit criteria.

## Local validation

```sh
go test -count=1 ./...
llgo test -count=1 ./...
# Equivalent entry points that select the host compiler:
./scripts/check.sh go
./scripts/check.sh llgo
```

The root tests cover host numeric bit patterns, reference reachability, goroutines, locks, atomics, condition variables, and cancellation. The [wazero evaluator](experiments/wazero/README.md) is a separate module with 13 pinned capability fixtures, negative-path tests, shared-memory concurrency checks, and a bounded cancellation observation. Its owned-code coverage is measured separately from engine conformance. Expected rejection of an unsupported fixture is an assessment result, not feature support.

Module path: `github.com/cpunion/wasmvm`. A project license has not yet been selected. No third-party implementation code is imported at this stage; future reuse must preserve the applicable licenses and notices.
