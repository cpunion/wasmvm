# WASI 0.2.12 interface inputs

The WIT files are copied unchanged from the verified WASI release archive at
revision `281ba75fafcd50961ef55f9e52747afcc9b71ede`. `manifest.json` pins the
archive and individual file SHA256 values and enumerates the exact implemented
interfaces/operations. The original WASI license is preserved in
`LICENSE.WASI.md`.

Implemented interfaces include CLI environment/exit/stdio/terminals, random,
wall/monotonic clocks, IO error/poll/streams, filesystem/preopens, network/IP
lookup and TCP/create. TCP and filesystem providers currently target Linux and
Darwin. No complete CLI/HTTP world or WASI 0.3 support is claimed. UDP/HTTP and
remaining world/component mechanisms stay open in the roadmap.

Runtime fixtures are project-authored Core/Component WAT assembled from the
pinned WIT ABI and independently validated with wasm-tools 1.258.0. Their
producers and hash manifests are retained with the package testdata. These
fixtures complement, and do not replace, upstream world/toolchain conformance.
