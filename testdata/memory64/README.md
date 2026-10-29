# Memory64 and Table64 proposal corpus

Source: [WebAssembly/memory64](https://github.com/WebAssembly/memory64/tree/9003cd5e24e53b84cd9027ea3dd7ae57159a6db1),
revision `9003cd5e24e53b84cd9027ea3dd7ae57159a6db1`.

The inventory includes all nine dedicated memory64 scripts, all three memory
bulk scripts, and all eight table-operation scripts, including mixed widths, plus `call_indirect.wast`.
Each execution tier runs all 13,599 commands / 13,069 binary assertions in these
21 complete scripts. The 70 malformed-text assertions are classified separately.
The archive contains 23 scripts with Memory64/Table64 declarations. The
remaining two (`imports.wast`, `table.wast`) also require unimplemented GC/typed
references; those combinations remain later scope. Generic data/binary scripts
stay covered by the independent Core 2.0 corpus.
No assertion in a selected script is removed or converted to an expected skip.

The manifest pins all 1,041 source/license/fixture hashes. WABT `wast2json`
1.0.41, revision `03a00a1334e6121fb0cce4fccbd6bb109b68acaa`, produces fixtures with:

```sh
wast2json SOURCE.wast --enable-memory64 --enable-multi-memory --no-check -o SOURCE/script.json
```

WABT cannot parse upstream's `module definition` at `memory64.wast:8`.
`memory64.producer.wast` preserves every source line while replacing that marker
with a temporary module identifier. Its JSON command is then classified as
`module_definition`: decode, validate, lower and native-compile as applicable,
without allocating an instance, exactly as upstream requests. The original
source and normalized producer input are both retained and hashed.

Four WABT invalid-input fixtures omit a required data-count section. Explicit
`producerBinaryCorrections` record insertion offsets, bytes and before/after
hashes, preserving the original semantic assertions; the same encoder issue
is already pinned for the Core 2.0 corpus. `LICENSE.tests` preserves Apache-2.0.
CI requires no producer download. This suite does not establish shared memory,
GC/typed references or the complete Core 3.0 profile.
