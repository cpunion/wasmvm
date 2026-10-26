# Package and Directory Design

Status: design draft, 2026-10-26. **Only entries marked as existing have been created; all other paths in the tree are planned.** The public packages currently contain only `doc.go`, with no VM API or execution capability.

## Directory tree

```text
wasmvm/
├── go.mod                         # Existing: github.com/cpunion/wasmvm
├── README.md                      # Existing: project status and entry point
├── .github/workflows/ci.yml        # Existing: the same tests under Go and llgo
├── scripts/
│   ├── check.sh                   # Existing: local compiler selection
│   └── update-corpus.sh           # Planned: manual updates, pinned sources
├── docs/                          # Existing
│   ├── research/{standards,implementations}.md
│   ├── design/{architecture,packages,llgo-ci}.md
│   ├── roadmap.md
│   └── validation.md
├── cmd/
│   └── wasmvm/                    # Planned: run/inspect/validate CLI
├── pkg/
│   ├── README.md                  # Existing: public package overview
│   ├── wasm/                      # Existing doc.go: public types/values/FeatureSet
│   ├── vm/                        # Existing doc.go: Engine/Store/Linker/Instance
│   ├── component/                 # Existing doc.go: component compile/link/run API
│   └── wasi/
│       ├── preview1/              # Existing doc.go: wasi_snapshot_preview1
│       ├── threads/               # Existing doc.go: optional legacy wasi-threads
│       ├── preview2/              # Existing doc.go: WASI 0.2 worlds
│       └── preview3/              # Existing doc.go: WASI 0.3 worlds
├── internal/
│   ├── compat/                    # Existing: shared host compiler probes
│   ├── binary/                    # Planned: bounded reader, LEB128
│   ├── core/
│   │   ├── module/                # Planned: internal module/type/instruction IR
│   │   ├── decode/                # Planned: Core binary decoding
│   │   ├── validate/              # Planned: control/type stacks, feature validation
│   │   └── types/                 # Planned: recursive type canonicalization/subtyping
│   ├── engine/
│   │   ├── state/                 # Planned: internal Store/Instance/frame state
│   │   ├── interp/                # Planned: resumable interpreter, EH, tail calls
│   │   ├── memory/                # Planned: private/shared memory, atomics, waiters
│   │   ├── heap/                  # Planned: WasmGC, roots, externref, handles
│   │   └── schedule/              # Planned: task budgets, cancellation, thread lifecycle
│   ├── component/
│   │   ├── decode/                # Planned: component binary decoding
│   │   ├── validate/              # Planned: component types and composition rules
│   │   ├── canon/                 # Planned: lift/lower, post-return
│   │   ├── resource/              # Planned: own/borrow/drop and handles
│   │   └── async/                 # Planned: task/future/stream state machines
│   ├── wasi/
│   │   ├── host/                  # Planned: host capabilities behind public adapters
│   │   ├── preview1/              # Planned: P1 ABI implementation
│   │   ├── threads/               # Planned: instance-per-thread ABI
│   │   ├── preview2/              # Planned: P2 bindings
│   │   └── preview3/              # Planned: P3 bindings
│   ├── platform/                 # Planned: isolated OS/compiler differences
│   └── spectest/                 # Planned: WAST assertions, fixture loader
├── tests/                         # Planned: cross-package/language integration
│   ├── spec/                     # Core/proposal conformance
│   ├── wasi/                     # P1/P2/P3 and capability tests
│   ├── threads/                  # Litmus tests, TLS, wait/notify, exit
│   ├── gc/                       # Roots, cycles, calls/exceptions, quotas
│   ├── llgo/                     # Pinned llgo guest producers
│   └── differential/             # Features shared with reference engines
├── testdata/                      # Planned: small, reproducible fixtures
│   ├── manifest.json             # Source, SHA, license, generation command, profile
│   ├── core/
│   ├── components/
│   └── llgo/
├── examples/                      # Planned: embed/hostfunc/wasi/thread/gc
└── benchmarks/                    # Planned: startup, throughput, memory, threads, GC
```

Create internal packages as their milestones require them, rather than precreating every empty directory. Examples must become runnable before being presented as working programs. Start with broader internal packages where useful, and split them once type boundaries are clear to avoid import cycles and layers that only forward calls.

## Public API boundaries

| Package | Public responsibility | Keep private or exclude |
| --- | --- | --- |
| `pkg/wasm` | ValueType/Value, FuncType, FeatureSet, feature descriptions | Decoder scratch state, pointer representation, GC slot implementation |
| `pkg/vm` | Lifecycle, Compile/Instantiate/Call, configuration/limits, host imports, error categories, controlled memory/reference access | Internal opcodes, Go runtime ABI, shared `[]byte` views that callers can retain indefinitely |
| `pkg/component` | Component loading/linking/calls, component values, controlled resource/task lifecycles | Canonical ABI scratch buffers, internal task queues |
| `pkg/wasi/preview1` | P1 host function registration and args/env/stdio/preopens/poll configuration | Legacy threads and custom socket extensions |
| `pkg/wasi/threads` | Explicit opt-in to legacy spawn/child-entry support for a P1 context | Claims of P2/P3 threading support |
| `pkg/wasi/preview2` | Registration of versioned WIT worlds and P2 resource/pollable adapters | P1 memory layout details |
| `pkg/wasi/preview3` | Registration of versioned P3 worlds and host adapters for async calls/streams | Substituting ordinary channel behavior for Canonical ABI semantics |

The requested `pkg` layout defines public entry points; implementation details remain in `internal`. The CLI, tests, and SDK use these entry points. Internal packages may be refactored without creating public compatibility commitments. The initial API remains experimental until it has been exercised by at least one real embedding scenario and complete lifecycle tests.

Place host capability interfaces in the public package closest to their callers. Public Config types must not require callers to name inaccessible `internal` types. Shared filesystem/clock/random implementations initially remain in `internal/wasi/host`; consider a public `pkg/wasi` package only when callers actually need to share configuration directly across WASI versions.

## Dependency direction and cycle prevention

```text
cmd / examples / tests
  ├─ pkg/wasi/preview1 ──────────────────────────┐
  ├─ pkg/wasi/threads → pkg/wasi/preview1       │
  └─ pkg/wasi/preview2,preview3 → pkg/component │
                                     ↓        ↓
                                  pkg/vm → pkg/wasm
                                     ↓
                         internal/engine + internal/core
```

Internal dependencies flow through `interp → state → memory/heap` and `interp → core/module/types`. Decoding and validation depend only on static Core types, not instances or the interpreter. Scheduling drives Execution through narrow interfaces; `state` must not import the entire interpreter in return. `state` implements a root visitor consumed by `heap`, avoiding a `heap → state → heap` cycle.

`pkg/vm` assembles and encapsulates the runtime. Internal execution code accepts internal interfaces/state and must not import `pkg/vm`. WASI adapters may register host functions through `pkg/vm` because the VM does not import WASI. The CLI selects and assembles adapters.

Definitions needed by both public and internal APIs belong in `pkg/wasm`, or in a low-level internal package with conversion at the public boundary. Cross-layer type aliases must not expose entire internal structures. Package count is not a design metric; understandable dependencies and isolated tests take priority.

## Implementation by phase

1. M0: retain the public `doc.go` declarations and compatibility probes; evaluate Wago/wazero and pin the corpus manifest and tools.
2. M1: implement `binary`, `core/{module,decode,validate}`, `engine/{state,interp,memory}`, and a minimal P1 host; add the CLI and embedding example.
3. M2a: add `engine/schedule`, shared-memory waiters, and WASI threads. In parallel, M2b develops `core/types`, `engine/heap`, and final EH support.
4. M3: complete the Core 3.0 target and broader tests; internal optimizations should not expand the public surface unnecessarily.
5. M4/M5: create component decoding/validation/canonical ABI/resource/async packages and P2/P3 bindings as needed.
6. M6: evaluate optional performance work and experimental proposals through separate design reviews and feature gates.

The [development roadmap](../roadmap.md) is the authoritative phase plan and defines the exit criteria for each phase.
