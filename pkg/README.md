# Public Packages

`wasm` implements Core 2.0 value-type identifiers and frontend error categories. The remaining packages reserve their intended responsibilities. No public VM execution API is implemented yet.

| Path | Planned responsibility |
| --- | --- |
| `wasm` | Public values/types and feature sets |
| `vm` | Engine, Store, modules, instances, host imports, and call APIs |
| `component` | Component Model embedding APIs |
| `wasi/preview1` | WASI Preview 1 |
| `wasi/threads` | Separate, optional legacy wasi-threads support |
| `wasi/preview2` | WASI 0.2 |
| `wasi/preview3` | WASI 0.3 |

See [packages.md](../docs/design/packages.md) for the complete layout and dependency direction, and [README](../README.md) for the current implementation status.
