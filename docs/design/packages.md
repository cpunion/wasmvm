# Package and Directory Design

Status: M1 implementation in progress, 2026-10-26. **Entries marked as existing or implemented have been created; other paths are planned.** The public packages currently contain only `doc.go`, with no VM API or execution capability.

## Directory tree

```text
wasmvm/
├── go.mod                         # Existing: github.com/cpunion/wasmvm
├── README.md                      # Existing: project status and entry point
├── .github/workflows/ci.yml        # Existing: the same tests under Go and llgo
├── scripts/
│   ├── check.sh                   # Existing: local compiler selection
│   ├── check_coverage.py          # Existing: per-package Go coverage gate
│   ├── test_check_coverage.py     # Existing: coverage-gate regression tests
│   └── update-corpus.sh           # Planned: manual updates, pinned sources
├── docs/                          # Existing
│   ├── research/{standards,implementations}.md
│   ├── design/{architecture,packages,llgo-ci,own-core}.md
│   ├── design/{core-compiler,native-abi,performance-contract}.md
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
│   ├── binary/                    # Implemented: bounded reader, LEB128, float bits, UTF-8 names
│   ├── core/
│   │   ├── module/                # Planned: internal module/type/instruction IR
│   │   ├── decode/                # Planned: Complete Core profile binary decoding
│   │   ├── validate/              # Planned: control/type stacks, feature validation
│   │   └── types/                 # Planned: recursive type canonicalization/subtyping
│   ├── compiler/
│   │   ├── ir/                    # Planned: typed CFG/SSA, effects and verifier
│   │   ├── lower/                 # Planned: validated Wasm to typed IR
│   │   ├── passes/                # Planned: verified optimizations
│   │   ├── machine/               # Planned: target IR and instruction selection
│   │   └── regalloc/              # Planned: liveness, registers, spills and moves
│   ├── engine/
│   │   ├── state/                 # Planned: internal Store/Instance/frame state
│   │   ├── interp/                # Planned: slot lowering, dispatch, resumable frames
│   │   ├── native/                # Planned: native artifacts, maps and driver
│   │   │   ├── amd64/             # Planned: Go instruction encoder/target lowering
│   │   │   └── arm64/             # Planned: Go instruction encoder/target lowering
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
│   │   ├── execmem/               # Planned: W^X, protection and instruction cache
│   │   └── callbridge/            # Planned: small Go/LLGo C ABI adapters
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
└── benchmarks/                    # Existing: competitor phase/CLI measurements
    └── embedded/wasmvm/           # Planned: exact own base/head protocol runner
```

Create internal packages as their milestones require them, rather than precreating every empty directory. Examples must become runnable before being presented as working programs. Start with broader internal packages where useful, and split them once type boundaries are clear to avoid import cycles and layers that only forward calls.

## Public API boundaries

| Package | Public responsibility | Keep private or exclude |
| --- | --- | --- |
| `pkg/wasm` | ValueType/Value, FuncType, FeatureSet, feature descriptions | Decoder scratch state, pointer representation, GC slot implementation |
| `pkg/vm` | Lifecycle, Compile/Instantiate/Call, configuration/limits, host imports, error categories, controlled memory/reference access | IR/opcodes, machine registers, native ABI, shared `[]byte` views that callers can retain indefinitely |
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

Internal dependencies flow through `interp/native → state → memory/heap` and `interp/native → compiler/ir → core/module/types`. The frontend lowers validated modules through `compiler/lower`; the native backend uses `compiler/{passes,machine,regalloc}` and `platform/{execmem,callbridge}`. Compiler logic must remain buildable and testable without importing a cgo-only execution adapter. `compiler/ir` does not import either backend or runtime state. Separate execution interfaces from artifact creation to avoid `native → callbridge → native` cycles. Decoding and validation depend only on static Core types, not instances or the interpreter. Scheduling drives Execution through narrow interfaces; `state` must not import the entire interpreter in return. `state` implements a root visitor consumed by `heap`, avoiding a `heap → state → heap` cycle.

`pkg/vm` assembles and encapsulates the runtime. Internal execution code accepts internal interfaces/state and must not import `pkg/vm`. WASI adapters may register host functions through `pkg/vm` because the VM does not import WASI. The CLI selects and assembles adapters.

Definitions needed by both public and internal APIs belong in `pkg/wasm`, or in a low-level internal package with conversion at the public boundary. Cross-layer type aliases must not expose entire internal structures. Package count is not a design metric; understandable dependencies and isolated tests take priority.

## Implementation by phase

1. M0: retain existing evaluation modules and comparison tooling as independent references.
2. M1: implement `binary`, `core/{module,decode,validate}`, `compiler/{ir,lower}`, `engine/{state,interp,memory}`, the core CLI and embedding examples; wire real wasmvm benchmarks. Complete Core 2.0 includes fixed SIMD.
3. M2: implement `compiler/{passes,machine,regalloc}`, `engine/native/{amd64,arm64}` and isolated execution adapters; optimize the interpreter and pass the same complete profile plus performance gates.
4. M3: extend `core/types`, `engine/heap`, scheduling, both backends and memory to full Core 3.0 plus separate Core Threads.
5. M4: implement P1 and legacy threads adapters, capabilities and Go/LLGo guest integrations.
6. M5/M6: add component decoding/validation/canonical ABI/resources/async and versioned P2/P3 bindings.
7. M7: add separately justified optimizations/platforms/experiments without leaking backend details into public APIs.

The [development roadmap](../roadmap.md) is the authoritative phase plan and defines the exit criteria for each phase.
