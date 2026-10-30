# Public Packages

`wasm` implements Core type/error contracts. `vm` implements the owned interpreter/native embedding API, including numeric/reference Values, Engine, Store, CompiledModule, Linker, Instance, Execution, typed host functions, object access and lifecycle management. See the [embedding guide](../docs/embedding.md) and [executable examples](vm/example_test.go). `component` executes synchronous composition and canonical calls through the same Core backends. P1 and legacy wasi-threads adapters execute their pinned operation/guest corpora; P2 supplies the explicitly listed interfaces in [its design](../docs/design/wasi-preview2.md). Complete P2 worlds and P3 native async remain required.

| Path | Responsibility |
| --- | --- |
| `wasm` | Core value-type identifiers, features, errors and traps |
| `vm` | Engine/Store lifetimes, modules, instances, executions, typed Values/references, host imports and calls |
| `component` | Synchronous composition/calls, own/borrow resources, checked host Core modules |
| `wasi/preview1` | WASI Preview 1 |
| `wasi/threads` | Separate, optional legacy wasi-threads support |
| `wasi/preview2` | Versioned environment, exit, random, clocks, IO, stdio, terminals, filesystem and IP name lookup providers |
| `wasi/preview3` | Reserved for WASI 0.3 native async |

See [packages.md](../docs/design/packages.md) for the complete layout and dependency direction, and [README](../README.md) for the current implementation status.
