# Public Packages

All packages belong to `github.com/cpunion/wasmvm`. They are available from
source with experimental APIs; no tagged module version or release has been
published. Pin a commit for dependencies. See the root [README](../README.md)
for package publication, CI and release status.

`wasm` implements Core type/error contracts. `vm` implements the owned interpreter/native embedding API, including numeric/reference Values, Engine, Store, CompiledModule, Linker, Instance, Execution, typed host functions, object access and lifecycle management. See the [embedding guide](../docs/embedding.md) and [executable examples](vm/example_test.go). `component` executes synchronous and native async composition/canonical calls through the same Core backends, including owned continuations, affine endpoints and scheduled host work. P1 and legacy wasi-threads adapters execute their pinned operation/guest corpora; P2 supplies the explicitly listed interfaces in [its design](../docs/design/wasi-preview2.md); P3 supplies stable native-async CLI/service/middleware worlds. Expanded world conformance and performance acceptance remain open.

| Path | Responsibility |
| --- | --- |
| `wasm` | Core value-type identifiers, features, errors and traps |
| `vm` | Engine/Store lifetimes, modules, instances, executions, typed Values/references, host imports and calls |
| `component` | Sync/async composition, canonical calls, resources, streams/futures and host bindings |
| `wasi/preview1` | WASI Preview 1 |
| `wasi/threads` | Separate, optional legacy wasi-threads support |
| `wasi/preview2` | Versioned CLI/HTTP providers, native sockets and explicit capability owners |
| `wasi/preview3` | Native async providers and unified CLI/service/middleware worlds |

See [packages.md](../docs/design/packages.md) for the complete layout and dependency direction, and [README](../README.md) for the current implementation status.
