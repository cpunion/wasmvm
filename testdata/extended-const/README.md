# Extended Constant Expressions proposal corpus

These are the complete `global.wast`, `data.wast` and `elem.wast` scripts from
[WebAssembly/extended-const](https://github.com/WebAssembly/extended-const/tree/8d4f6aa2b00a8e7c0174410028625c6a176db8a1),
revision `8d4f6aa2b00a8e7c0174410028625c6a176db8a1`. They exercise all six
arithmetic additions to constant expressions, imported immutable globals,
segment offsets, reference initializers and invalid/linking/trapping cases.
This is the feature proposal corpus, not the complete Core 3.0 corpus.

WABT `wast2json` 1.0.41, revision
`03a00a1334e6121fb0cce4fccbd6bb109b68acaa`, generated each directory:

```sh
cd testdata/extended-const
wast2json global.wast --enable-extended-const --no-check -o global/script.json
wast2json data.wast --enable-extended-const --no-check -o data/script.json
wast2json elem.wast --enable-extended-const --no-check -o elem/script.json
```

`manifest.json` pins the source, producer, arguments and all 198 source/license/
fixture hashes. Tests verify every hash and execute 290 commands containing
215 binary assertions in each of the interpreter and two native tiers. Three
text-parser assertions are classified separately. `LICENSE.tests` preserves
the upstream Apache-2.0 notice. No producer or download is needed in CI.
