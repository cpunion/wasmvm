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
Component Model conformance nor actual guest execution. Lexical type/index
validation, composition and Canonical ABI calls remain separate required work.

The same archive also pins all 81 commands in `test/validation/abi.wast`,
`core-modules.wast` and `defined-types.wast`: five accepted definitions and 76
invalid rejections. `internal/component/validate` checks their complete
scripts, including embedded Core code, linking and canonical options. These
are three semantic scripts, not the complete component validation inventory
or a claim of guest/component execution.
