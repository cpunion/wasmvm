# Complete applicable Core Threads corpus

Source: [WebAssembly/threads](https://github.com/WebAssembly/threads/tree/cc535ada1aa21cfaa3cabf3ac73b89acef78a0a0/test/core),
revision `cc535ada1aa21cfaa3cabf3ac73b89acef78a0a0`.
The source archive is 4,419,412 bytes, SHA-256
`8cc7e0b408c6ae84865497aec07fc3afd85f5021143111681e9abc3ead963290`.
Upstream test assets retain their [Apache-2.0 notice](LICENSE).

`manifest.json` pins all 495 source, generated fixture and license hashes.
The producer is wasm-tools 1.258.0, revision
`5c6d31c78f8bd503f558441ef9c732950a141d1a`. Regenerate each retained source with:

```sh
wasm-tools json-from-wast source/threads/atomic.wast \
  --wasm-dir threads__atomic -o threads__atomic/script.json
```

The inventory contains all 13 `threads/*.wast` scripts and the four root scripts
with shared forms: binary, elem, exports and memory. Scripts are retained and
executed in full, without corrections or excluded commands. Unchanged root Core
2.0 coverage remains in its independent complete corpus. The totals, including
nested thread commands, are **17 scripts, 946 commands, 579 binary assertions**
and **six separately classified text-parser assertions**.

`TestCoreThreadsConformance` validates every hash, executes all three available
tiers, joins nested threads before cleanup, and reports actual counts. Set
`WASMVM_CONFORMANCE_REPORT` to retain per-tier `proposals/threads-*.json` reports
alongside the other conformance results. CI executes the same inputs with Go
and LLGo on Linux amd64 and macOS arm64; the concurrent corpus also runs under
Go's race detector.
