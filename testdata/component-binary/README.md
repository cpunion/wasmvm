# Component binary and semantic inputs

This offline corpus contains the complete 132-command `test/binary/binary.wast`
at Component Model revision `d095e2f60ddae1f9a1e9b1e2fb547ec1114da677`, converted
with wasm-tools 1.258.0. Original source and Apache-2.0 notices are retained.
The manifest records source/output hashes and conversion settings. Only the
JSON source filename is normalized; component bytes/assertions are unchanged.

```sh
python3 scripts/component_corpus.py
go test ./internal/component/decode
```

The grammar test rejects all 80 malformed encodings and decodes 52 other
definitions, including the 16 inputs whose semantic validation must later fail.
Those 16 decoded inputs are not validation passes. This is neither complete
Component Model conformance nor actual guest execution. Semantic validation is measured separately below; Canonical ABI execution
and full component conformance remain required work.

The same archive pins 13 complete generated semantic scripts: `abi`,
`core-modules`, `defined-types`, `external-visibility`, `instantiation`,
`outer-alias`, `resources`, `annotated-names`, `attributes`, `extern-names`,
`kebab`, `max-value-size` and `indicies`. Their 461 binary commands cover 100
accepted definitions and 361 invalid rejections. Five text-parser malformed
assertions are reported separately; they are producer grammar inputs rather
than binary validation passes. `internal/component/validate` checks every
binary command, including embedded Core code, composition, resource binding,
interface subtyping and canonical options. This remains a subset of the full
component validation inventory and does not imply component execution.

The pinned producer cannot convert `test/validation/version-suffix.wast`:
wasm-tools 1.258.0 rejects `(version-suffix ".2.3")` on its first import.
The original source is retained, and the manifest records this producer
limitation. It is an untested official script, not a validation pass or a
silently skipped engine feature. Authored version-attribute tests are separate.
