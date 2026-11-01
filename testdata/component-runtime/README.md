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
the pinned upstream archive plus wasm-tools 1.258.0 and the independently
pinned 1.262.0 producer. The older producer still generates its original inputs
byte-for-byte except for the declared zero-length semantic adaptation below;
the additional producer converts only the four former limitations.
Its release archive and executable checksums are retained in the manifest. The separate
[grammar/type corpus](../component-binary/README.md) covers those scripts.

The declared runtime profile selects 962 commands across 52 script views:
458 successful result assertions, 96 trapping invocations, one action, 317
successful instantiations, two initialization traps, 45 invalid rejections
and 43 component definitions. Each owned backend executes every selected command
under both Go and LLGo. The runner distinguishes semantic traps from timeouts,
resource limits and unsupported features; it never substitutes another runtime.
No expected-unsupported assertion counts as successful execution.

The manifest separately retains 228 commands requiring the optional
component-threading profile. Classification uses independently printed canonical
thread opcodes and follows module/instance dependencies. Three mixed scripts also
contain selected commands: cancel-targeted-resume, reentrance and post-return. Those views
are not claimed as complete-script passes. Component threading is M7, distinct
from implemented Core Threads and P3 async tasks.

The former stream.forward and version-suffix producer limitations are resolved.
The original sources and prior diagnostics remain alongside the new conversion
identity. Actual official forwarding, mixed asynchronous interleavings,
non-threaded post-return cases and all version-suffix assertions execute on the
owned engine. This is complete conversion of the pinned script inventory,
with component threading still outside the declared runtime profile; it does
not establish exhaustive WASI acceptance or close experimental cm64 semantics.

The cumulative October 30 Canonical ABI correction checks typed buffer pointers
even at length zero. `semanticCorrections` retains the original zero-length
script and an adapted copy replacing its two `0xdeadbeef` addresses with the valid
one-page memory endpoint. Every command and expected result is preserved; the
runner executes the adapted binary. Offline preparation verifies the exact
two-address transformation. Separate owned negative tests require invalid empty
pointers to trap. See [the correction pins](../component-cm64/README.md).
