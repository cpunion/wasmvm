# Core 3.0 profile

`EngineConfig.Profile = "core3"` selects the complete pinned Core 3.0 binary
profile on the interpreter and both native targets. The CLI accepts
`--profile core3` for validate, inspect and run. Validation, inspection and
compiled metadata report `core3` and the effective feature mask. Default
configuration remains strict `core2`; archived proposal flags retain their
pinned behavior.

`wasm.Core3Features` combines Extended Const, Tail Call, Multi-Memory, Relaxed
SIMD, Memory64/Table64, Typed Function References, WasmGC, final Exception
Handling and the final `Core3Rules`. Profile selection adds the complete mask;
callers need not enumerate its prerequisites. Core Threads is independent.
Components and WASI 1/2/3 are subsequent layers, not implied by the profile.

Final Core 3.0 differs from some archived standalone proposals:

- Both i32 and i64 memory/table limit fields use u64 binary encodings. Address
  width and semantic size bounds remain distinct. A syntactically representable
  oversized i32 object is invalid, rather than a truncated or malformed u32.
- Memory argument flags retain six alignment bits. Encoded alignment values
  32 through 63 decode before semantic natural-alignment checks reject them.
  Reserved flag bits still fail binary decoding.
- Constant initializers can read preceding immutable defined globals. Each
  global is checked in definition order; self/forward/unknown and mutable
  references remain invalid. Tables are checked before global definitions and
  see imports only. Data/element segments follow all global definitions.

The existing initializer evaluator already resolves preceding cells from each
instance. Compiled constants are reused only when an expression contains no
global reads, preserving imported values and instance-dependent offsets. The
same bounded frontend, verified SSA, slots, explicit frames, heap and native
compiler execute both profiles. No competitor or interpreter fallback implements
the native profile.

The [complete corpus](../../testdata/core3/README.md) contains all 259 pinned
scripts, including final cross-proposal cases, rather than only a union of
archived proposal suites. Every available execution tier runs all 61,395 binary
assertions. Reports include source/manifest identity, compiler/runtime/target,
feature mask, tier, command counts and separately classified parser assertions.
CI preserves the same Go/LLGo × Linux amd64/macOS arm64 matrix. Native tiers
remain unmetered checked execution and precise instruction fuel.

Additional authored tests combine defined-global initialization, imported seeds,
GC construction, exception payloads, active data/elements, Table64, multiple
memories, relaxed SIMD and typed tail calls. Forced collection, per-instance
values, initialization failure cleanup, cancellation and quotas use the same
Store contracts. Decoder tests distinguish syntax from semantic overflow;
validator tests reject visibility and mutability errors.

Complete core semantics do not close calibrated performance acceptance or M3's
separate Threads requirements. Guest heap allocation, native call/helper
handoffs, exceptional unwind and remaining interpreter/compiler deficits retain
their existing local follow-ups. Performance measurements keep startup,
compilation, instantiation, first call, warm execution and cleanup separate.
