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


Resource fixtures cover new/rep/drop, i32/i64 representations, guest destructor
effects/traps, nominal identity across repeated child instantiation, imported
type round trips, aggregate handle quotas, post-return leave guards and failed initialization
(including host panic/cancellation after resource creation). Raw u32
handle exports are intentional test plumbing, not host ownership capabilities.
Wasmtime 48.0.1 executes the 32-bit probes. With
`-W component-model-memory64=y`, the 64-bit probe currently triggers a Cranelift
I64/I32 assertion; retain that failure separately from successful own execution.


Ownership fixtures additionally cover guest/host declarations, nominal equality,
opaque moves and aliases, borrowed callback expiry/escape, cross-component
forwarding, list/map/option/variant/fixed-list/record transfers with spilled ABI,
failed input/result cleanup, host panic/cancellation, suspended component entry
and pure host depth limits. `ownership64` preserves cm64 own representations;
`ownership64-borrow` is a validated fixture intentionally rejected by the current
runtime profile. The Wasmtime C API reference executes `ownership`, including
all four calls and guest/explicit destructors. Its stale-capability negative
check runs after successful assertions because the reference locks down that
instance on this canonical trap. wasmvm validates bad host capabilities before
guest effects and keeps the instance usable; the embedding error policy differs.


`function-check-own` is an independently valid component that declares an own
parameter where the provider requires borrow. Binding rejects it through the
host contract before execution, rather than consuming the argument at a later
call. Contract failure before an actual Core start is tested separately.


`bytes`/`bytes64` echo arbitrary list<u8> data through cm32/cm64 canonical
memory, realloc and post-return reset. `bytes-host`/`bytes-host64` additionally
round-trip through a lowered host callback. They are project-authored, generated
and validated with wasm-tools 1.258.0 (`--features all`). Retained results must
survive allocator reuse. Packed raw-byte storage is independent of the selected
string encoding; it performs no UTF-8 validation. The pure `bytes` fixture is
also executed unchanged by Wasmtime 48.0.1 C API and typed Rust API probes.


`byte-value-imports` is the existing authored value-import grammar with a u8
list element. `byte-value-constant-bound` adds a four-byte literal and equality
bound. They are manually authored from the same pinned value grammar; the WAT
files describe their intended structure. wasm-tools 1.258.0 does not support
these value imports/value section, so its diagnostic failures are retained.
Own Go/LLGo tests verify equivalent packed/element-wise nested bindings and
literal bounds, owned exports and rejection of unequal bytes.
