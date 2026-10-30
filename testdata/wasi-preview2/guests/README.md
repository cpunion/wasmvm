# Independent WASI 0.2.12 command guests

The C command uses Clang 22.1.8 `-O2` with wasi-libc wasi-sdk-33. The Rust command
uses Rust 1.98.0, `opt-level=2`, `panic=abort` and the retained Cargo.lock. Both
use wit-bindgen 0.54.0 and the unchanged WASI 0.2.12 archive. Dependencies are
extracted into the ignored producer work directory, rather than duplicated here.
No Preview 1 adapter, host-provider replacement or hand-authored guest body is
substituted for either compiled language program.

The C component exports an arithmetic interface and CLI run. The Rust CLI calls
that C interface after composition, then checks configured arguments, randomness,
a pollable, a preopened directory and output. Each command also runs against the
actual Wasmtime 48.0.1 WASIp2 provider. The complete import-surface fixture is
separate: ordinary compiler DCE narrows language guests' unused imports. This
fixture binds all stable CLI interfaces and returns success, while dedicated
provider tests exercise the individual operations. Neither a no-op full import
check nor two small programs alone establish complete upstream conformance.

Install the fixed binding generator and reproduce with:

```sh
cargo +1.98.0 install wit-bindgen-cli --version 0.54.0 --locked \
  --no-default-features --features c,rust --root artifacts/components/wit-bindgen-0.54.0
python3 scripts/build_wasi_world_fixtures.py --sysroot /path/to/wasi-sysroot
```

The script validates every component and records source/binary hashes. It also
produces valid standard-entry controls for success, result failure, trap,
nonreturning exit and incorrect signatures. The latter must fail before their
observable host effect. Source archive SHA-256 for wasi-libc wasi-sdk-33 is
`d6ad9eb089fc00b4fe980556125c133406668d23562e32fa273b5c4955eeefc3`.
Generated binding/runtime code uses the upstream MIT license option. Notices
are retained in `notices`; Rust's full standard-library copyright inventory is
compressed without loss as `Rust-COPYRIGHT-library.html.gz`.
