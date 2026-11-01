# Pinned upstream component runtime corpus

All 75 original Component Model scripts at revision
`d095e2f60ddae1f9a1e9b1e2fb547ec1114da677` and their Apache notices are retained.
The source archive, generator, converted JSON and binary inputs are hashed.
Ordinary CI uses the offline archive:

```sh
python3 scripts/component_corpus.py
go test ./pkg/component -run '^TestUpstreamComponentRuntime$'
```

Regenerate with `python3 scripts/component_runtime_corpus.py --regenerate` and
the pinned upstream archive plus wasm-tools 1.258.0. The separate
[grammar/type corpus](../component-binary/README.md) covers those scripts.

The declared runtime profile selects 759 commands across 48 script views:
381 successful result assertions, 76 trapping invocations, one action, 258
successful instantiations, two initialization traps, eight invalid rejections
and 33 component definitions. Each owned backend executes every selected command
under both Go and LLGo. The runner distinguishes semantic traps from timeouts,
resource limits and unsupported features; it never substitutes another runtime.
No expected-unsupported assertion counts as successful execution.

The manifest separately retains 167 commands requiring the optional
component-threading profile. Classification uses independently printed canonical
thread opcodes and follows module/instance dependencies. Two mixed scripts also
contain selected commands: cancel-targeted-resume and reentrance. Those views
are not claimed as complete-script passes. Component threading is M7, distinct
from implemented Core Threads and P3 async tasks.

Four scripts remain unconverted by the pinned producer: async/forward,
async/big-interleaving-test and values/post-return require stream.forward text
syntax; validation/version-suffix requires its text attribute. Original sources
and exact diagnostics are retained. Existing authored forwarding/ABI fixtures
remain separate evidence. These limitations do not constitute conformance passes
or establish exhaustive component/WASI acceptance.
