# WASI 0.2 execution fixtures

These project-authored components are generated with wasm-tools 1.258.0 and
validated with `wasm-tools validate --features all`. `environment` executes
all three pinned WASI 0.2.12 environment operations through Canonical ABI
memory/list/string/option transfers, including guest consumers and post-return
allocation reset. The mismatch fixtures are valid components whose imports
must be rejected by the host provider before guest execution.

The pinned WIT definition, hashes and upstream license are in
`testdata/wasi-preview2`. Wasmtime 48.0.1 C API executes the same fixture using
its actual WASIp2 provider with explicit arguments/environment and no inherited
host configuration. `exit` runs both nonreturning operations through canonical lower/Core/lift,
with two independently exported child instances. The same Wasmtime provider
verifies success, failure and explicit byte codes. The exit mismatch fixtures
are valid components rejected during host signature checking. These cover two
complete interfaces, not a complete world.
