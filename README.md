# wasmvm

A WebAssembly virtual machine written in Go, designed to build and run with both Go and LLGo.

**Research and design stage: the VM cannot execute Wasm yet.** This repository contains standards and implementation research, architecture and package designs, a development plan, and host compatibility probes shared by both compilers. The `doc.go` files reserve package responsibilities; they do not indicate implemented features.

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
| [Architecture](docs/design/architecture.md) | Execution, types, GC, threads, WASI, isolation, and draft embedding API |
| [Package layout](docs/design/packages.md) | Public `pkg` APIs, internal layers, dependency direction, and implementation order |
| [Go/LLGo and CI](docs/design/llgo-ci.md) | Compiler constraints, pinned toolchains, and host-versus-guest testing |
| [Development plan](docs/roadmap.md) | Milestones, work packages, dependencies, deliverables, acceptance gates, and initial backlog |
| [Validation record](docs/validation.md) | Actual local and CI results, with their scope and limitations |

The current design baseline is an independent pure Go interpreter. **M0 evaluates Wago and wazero before deciding which implementation to build on.** A production JIT, LLVM API binding, or external native VM is not part of the baseline. Using LLGo as a host compiler does not require the VM to call LLVM APIs.

## Development phases

| Phase | Outcome | Status |
| --- | --- | --- |
| Design baseline | Research, architecture, package contracts, shared compiler probes, and CI | Complete |
| M0 | Reuse decision, pinned fixtures, and reproducible candidate experiments | Planned; next |
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

Current tests cover host numeric bit patterns, reference reachability, goroutines, locks, atomics, condition variables, and cancellation. Passing these probes does not establish Wasm, WASI, or thread conformance. The development plan defines the later semantic acceptance suites.

Module path: `github.com/cpunion/wasmvm`. A project license has not yet been selected. No third-party implementation code is imported at this stage; future reuse must preserve the applicable licenses and notices.
