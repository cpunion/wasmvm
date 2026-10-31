# Unmodified upstream standard-library commands

Fourteen Rust test programs are copied without edits from Wasmtime 48.0.1,
commit `7bac2c2775808aaec5d4aa5627a5e447b51102cf`, under
`crates/test-programs/src/bin/p2_cli_*.rs`. Their Apache-2.0 WITH LLVM-exception
license is retained in LICENSE. The manifest pins every source and output,
the complete upstream archive, Rust 1.98.0 and its wasm32-wasip2 standard-library
archive, producer flags and wasm-tools 1.258.0.

Install the pinned target with `rustup target add wasm32-wasip2 --toolchain 1.98.0`
and reproduce with `python3 scripts/build_upstream_wasi_fixtures.py`. The script
uses edition 2024, `-O -C panic=abort`, remaps source paths and removes custom
sections with `wasm-tools strip --all`, then independently validates every
component. It uses Rust's supplied component linker and target libraries;
no guest test body or host-provider substitute is injected. Rust/libc notices
are also retained in [the existing notice inventory](../guests/notices).

These guests import WASI 0.2.9 and export CLI run 0.2.0. They exercise real Rust
standard-library arguments (including empty/Unicode arguments and an absent
argv0), environment, stdout/stderr, stdin, clocks, sleep, exit, panic, file reads,
seek-preserving append and file sync. Fixtures receive the upstream expected
arguments/input/environment/file contents. Nonzero exit and panic are expected
outcomes; they are not converted into successful execution. The actual pinned
Wasmtime WASIp2 provider is the independent comparison.

This is a bounded upstream CLI subset, not the entire Wasmtime test suite,
complete WASI conformance, HTTP-world acceptance, or WASI 0.3 support.
