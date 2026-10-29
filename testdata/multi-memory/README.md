# Multi-Memory proposal corpus

The complete 37 scripts in `test/core/multi-memory`, plus
`test/core/memory-multi.wast` and `test/core/simd/simd_memory-multi.wast`, come from
[WebAssembly/multi-memory](https://github.com/WebAssembly/multi-memory/tree/cf8b5aa27257311b8eac80ae83f4ba22ee308064),
revision `cf8b5aa27257311b8eac80ae83f4ba22ee308064`.

The manifest pins all 197 source/license/fixture hashes and each script's command
counts. WABT `wast2json` 1.0.41, revision
`03a00a1334e6121fb0cce4fccbd6bb109b68acaa`, produced each fixture directory from
its same-named `.wast` source with:

```sh
wast2json SOURCE.wast --enable-multi-memory --no-check -o SOURCE/script.json
```

Each interpreter/native tier executes 830 commands and 697 binary assertions
across all 39 scripts. There are no text-only assertions in this corpus.
Coverage includes imports/exports/linking, scalar/SIMD accesses, growth, data,
bulk operations, alignment, encodings and traps. Additional owned tests cover
large aliased copies, all-range prechecks, empty memory zero, global aliases,
host reentry/growth and quota rollback. CI requires no producer download.
`LICENSE.tests` retains the upstream Apache-2.0 notice. This suite does not
establish Memory64, shared-memory Threads or complete Core 3.0 conformance.
