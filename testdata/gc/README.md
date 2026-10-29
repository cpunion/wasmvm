# WasmGC proposal corpus

Source: [WebAssembly/gc](https://github.com/WebAssembly/gc/tree/756060f5816c7e2159f4817fbdee76cf52f9c923),
revision `756060f5816c7e2159f4817fbdee76cf52f9c923`.

All 18 `test/core/gc` scripts and all five root scripts containing GC type/
instruction forms are retained completely. Each execution tier runs all 964
commands and 802 binary assertions. One malformed text assertion remains
separate from the binary scope. No command or assertion is removed.

`manifest.json` pins all 272 source/fixture/license hashes. The producer is
`wasm-tools json-from-wast` 1.258.0, revision
`5c6d31c78f8bd503f558441ef9c732950a141d1a`; the local producer binary digest is
also retained. WABT 1.0.41 cannot produce these full GC forms. Reproduce with:

```sh
wasm-tools json-from-wast SOURCE.wast --wasm-dir SCRIPT -o SCRIPT/script.json
```

No producer correction is applied. The harness accepts signed numeric bit
patterns at their declared width, checks reference-kind wildcards, and traces
exact internalized external identities. Original WAST sources and the upstream
Apache-2.0 test license are retained. CI requires no producer download.

This is the standalone GC v1 proposal inventory, not the later Core 3.0 suite's
EH/GC and other cross-proposal combinations. The full Core 3.0 profile remains
an additional acceptance target.
