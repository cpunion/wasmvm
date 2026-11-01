# WASI 0.3.2 independent world guests

The C and Rust command guests use the pinned CLI world, native stdio streams,
completion futures, clock waits, randomness and async filesystem metadata. C
uses synchronous bindings over async WIT plus native waitables; Rust uses async
exports/imports and suspended callbacks. Tests force three-byte batches and
repeat invocation on one instance.

The Rust service consumes an incoming body and then returns a response before
its spawned byte/trailer producer finishes. It waits for transmission completion.
The Rust middleware forwards the owned request through the standard handler and
returns the upstream response, whose body can outlive the handler invocation.
These exercise actual language-generated ownership and async protocols; no P2
adapter or alternate execution engine is used.
The composed fixture connects the Rust middleware to the Rust service through
`wasm-tools compose` and exercises body/trailer transfer across both components.

Regenerate with `python3 scripts/build_preview3_world_fixtures.py`. The builder
requires the pinned WASI archive, wasm-tools 1.258.0, wit-bindgen 0.54.0, clang
22.1.8/wasi-libc 33 and Rust 1.98.0. Cargo dependencies are locked and built offline
for wasm32-unknown-unknown; native P3 bindings provide the component ABI. The C
sysroot supplies libc memory/string support without introducing P1 imports.

The builder separately preserves full stable CLI, service and middleware import
surfaces beyond language-compiler DCE. Only the full CLI dummy export executes;
service/middleware dummy exports are binding contracts. Executed language guests
are recorded separately. The manifest pins sources, generator and binaries.
Unstable timezone/component-threading extensions are not enabled.

`notices` retains Rust standard-library and C-library notices plus deduplicated
license texts for locked Cargo dependencies. `crates.json` maps package versions
and Cargo checksums to those texts. These integration guests do not by themselves
claim exhaustive upstream conformance or close remaining cm64 contracts.
