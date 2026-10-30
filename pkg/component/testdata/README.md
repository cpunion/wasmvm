# Synchronous component execution fixtures

These project-authored fixtures exercise actual execution through both own Core
backends. WAT files other than `value-*` are converted with wasm-tools 1.258.0
and independently checked with `wasm-tools validate --features all`. The
`invalid-index` input is intentionally invalid and is not a reference pass.

`value-imports` and `value-constant-bound` are authored directly from Component
Model Binary.md at revision `d095e2f60ddae1f9a1e9b1e2fb547ec1114da677`.
The pinned producer cannot encode their value syntax. The WAT documents their
intent; these binaries are authored tests, not independent conformance oracles.

The fixtures cover three string encodings, memory32/64, host calls,
post-return, spilled parameters/results, type ascription, outer closures,
value equality/moves, void/float calls, map/fixed-list calls, bounded instance expansion and traps.
