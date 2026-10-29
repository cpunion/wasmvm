# Typed Function References proposal corpus

Source: [WebAssembly/function-references](https://github.com/WebAssembly/function-references/tree/74d2ec81d15efd3c0f2fba46a023f376101d8e46),
revision `74d2ec81d15efd3c0f2fba46a023f376101d8e46`.

All 19 root scripts containing typed/non-null references or the new instructions
are included in full: 1,174 commands, 978 binary assertions and 29 malformed text
assertions classified separately. Generic Core 2.0 behavior also has its own
independent complete pinned corpus. No assertion or command is removed.

The manifest pins all 588 source, fixture and license hashes. Produce with WABT
`wast2json` 1.0.41 (`03a00a1334e6121fb0cce4fccbd6bb109b68acaa`):

```sh
wast2json SOURCE.wast --enable-function-references --enable-tail-call --no-check -o SOURCE/script.json
```

WABT encodes wildcard `(ref.func)` expectations as function index `0`.
`producerJSONCorrections` records source-checked removal of that value field
only, with counts and before/after SHA-256. The harness then asserts a nonnull
function reference, exactly as upstream requests. Concrete function-index
expectations use exact instance identities. Original WAST inputs and the
Apache-2.0 test license are retained. CI requires no producer downloads.

The standalone proposal uses acyclic invariant function types; this corpus does
not establish recursive types, GC, final EH, shared memory or all Core 3.0.
