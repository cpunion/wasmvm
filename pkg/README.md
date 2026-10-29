# Public Packages

`wasm` implements Core 2.0 type/error contracts. `vm` implements the owned interpreter/native embedding API, including numeric/reference Values, Engine, Store, CompiledModule, Linker, Instance, Execution, typed host functions, object access and lifecycle management. See the [embedding guide](../docs/embedding.md) and [executable examples](vm/example_test.go). Component and WASI packages remain reserved.

| Path | Responsibility |
| --- | --- |
| `wasm` | Core 2.0 value-type identifiers, errors and traps |
| `vm` | Engine/Store lifetimes, modules, instances, executions, typed Values/references, host imports and calls |
| `component` | Component Model embedding APIs |
| `wasi/preview1` | WASI Preview 1 |
| `wasi/threads` | Separate, optional legacy wasi-threads support |
| `wasi/preview2` | WASI 0.2 |
| `wasi/preview3` | WASI 0.3 |

See [packages.md](../docs/design/packages.md) for the complete layout and dependency direction, and [README](../README.md) for the current implementation status.
