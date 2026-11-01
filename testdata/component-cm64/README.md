# Canonical ABI corrections for cm64 and empty buffers

`inputs.tar.gz` retains the unchanged CanonicalABI.md, Binary.md and Apache
notices from Component Model revision
[`fed10b0b823b82553dd8363f16a640816f5eac65`](https://github.com/WebAssembly/component-model/commit/fed10b0b823b82553dd8363f16a640816f5eac65).
The manifest pins every file and the deterministic archive by SHA-256.
This is a specification input, not an executed reference implementation.

This snapshot includes the October 30 corrections for direction-aware i64
borrow representations ([#754](https://github.com/WebAssembly/component-model/pull/754))
and typed zero-length stream/future buffer pointer checks
([#755](https://github.com/WebAssembly/component-model/pull/755)).
Earlier test/specification snapshots remain intact.

`scripts/build_cm64_borrow_fixtures.py` verifies this archive and generates nine
owned fixtures with wasm-tools 1.258.0 parsing. The Go/LLGo semantic matrix
executes them on the interpreter and native backend. Their manifest separately
pins the generator, specification manifest, original owned source and outputs.
Independent wasm-tools 1.262.0 validation still expects an i32 local borrow flat
parameter and rejects the corrected i64 signature; parsing is not validation.

The old upstream zero-length script uses two out-of-range addresses for typed
empty buffers. The runtime corpus retains that source, plus a separately hashed
adaptation using the one-page memory endpoint for those two buffers. All command
indices, readiness expectations and selected/deferred counts remain unchanged.
Authored tests separately require misaligned/out-of-bounds empty buffers to trap.
