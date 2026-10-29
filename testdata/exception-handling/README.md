# Final exception-handling corpus

Source: [WebAssembly/spec](https://github.com/WebAssembly/spec/tree/aa4dcf44b8de3f3f3270ca067422172518ecc341),
revision `aa4dcf44b8de3f3f3270ca067422172518ecc341`.

All five `test/core/exceptions` scripts and all four root scripts containing
EH forms (`imports`, `exports`, `instance`, `ref_null`) are retained completely.
Each execution tier runs 482 commands and 305 binary assertions. Eighteen
malformed text assertions remain separate from the binary scope. No command
or assertion is removed.

`manifest.json` pins 325 source/fixture/license hashes and the producer binary.
The producer is `wasm-tools json-from-wast` 1.258.0, revision
`5c6d31c78f8bd503f558441ef9c732950a141d1a`. Reproduce from the retained source:

```sh
wasm-tools json-from-wast testdata/exception-handling/source/SCRIPT.wast \
  --wasm-dir testdata/exception-handling/SCRIPT \
  -o testdata/exception-handling/SCRIPT/script.json
```

Original complete WAST sources and the upstream Apache-2.0 test license are
retained. No producer correction is applied. The harness implements generative
`module_definition`/`module_instance` commands and separately classifies
`assert_exception`, traps and ordinary runtime errors. CI needs no producer
download. These EH-bearing scripts do not replace the full Core 3.0 inventory.
