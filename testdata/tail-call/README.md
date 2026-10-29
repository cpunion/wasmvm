# Tail Call proposal corpus

The complete `return_call.wast` and `return_call_indirect.wast` scripts come from
[WebAssembly/tail-call](https://github.com/WebAssembly/tail-call/tree/a6003d06aefef41e20a3e36fe2e500062555c895),
revision `a6003d06aefef41e20a3e36fe2e500062555c895`. They cover the two original
tail-call instructions, typing, recursion, tables and expected traps. The later
function-references instruction `return_call_ref` has separate prerequisites.

WABT `wast2json` 1.0.41, revision
`03a00a1334e6121fb0cce4fccbd6bb109b68acaa`, generated the checked-in fixtures:

```sh
cd testdata/tail-call
wast2json return_call.wast --enable-tail-call --no-check -o return_call/script.json
wast2json return_call_indirect.wast --enable-tail-call --no-check -o return_call_indirect/script.json
```

The manifest preserves all 46 source/license/fixture hashes and command counts.
Each interpreter/native tier executes 119 commands with 102 binary assertions;
11 text-parser assertions are classified separately. CI uses the same checked-in
inputs under Go and LLGo without downloading a producer. `LICENSE.tests` retains
the upstream Apache-2.0 notice. This is proposal coverage, not full Core 3.0.
