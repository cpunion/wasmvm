# Legacy thread corpus

`manifest.json` pins the full `WebAssembly/wasi-threads` test inventory at
`f8ea03cc0ffbde143415e945497de112f6f9fd53`: 14 WAT programs and one C program.
The archive retains original source, program metadata, generated binaries,
producer records and notices. Preparation is offline:

```sh
python3 scripts/wasi_threads_corpus.py
go test ./pkg/wasi/threads
```

WAT inputs use wasm-tools 1.258.0. The C input and original assembly shim use
Clang/LLD 22.1.8 with SDK25 headers, `crt1-command.o`, threaded libc and compiler
builtins; the manifest records the target/link flags. Original source is not
patched to remove assertions or shorten waits. The proposal archive contains
no license file; its original source and comments are retained, and redistribution
clearance remains a release requirement.

The project-authored `guest-llgo` uses LLGo revision
`8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f`, Go 1.27.3 and SDK25. Build with
the LLVM/LLD **22.1.8** directories first on PATH, `LLGO_ROOT` pointing at that
revision, and `CFLAGS=-msimd128`:

```sh
cd testdata/wasi-threads/guest-llgo
llgo build -O3 -lto=full -target wasi -o guest-original.wasm .
wasm-tools strip --all guest-original.wasm -o guest.wasm
```

The original and stripped digests, source hashes, imports, passive/active segment
counts, emitted SIMD inventory and compiler settings identify the asset. Only
custom sections are stripped; Wasm code/data are not rewritten. LLGo/Go notices
are retained. Do not infer WasmGC support from this linear-memory collector.

The normal guest prints `llgo-threads-ok 828`. Mode `exit` calls `os.Exit(23)`
from a worker; mode `panic` reports `expected-worker-panic` and exits 2. Native
tests explicitly allow 256 MiB of module code. This corpus validates a specific
SDK/producer ABI; later upgrades need new hashes and the same lifecycle checks.
