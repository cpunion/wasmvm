# WasmGC execution and ownership

`wasm.GC` (CLI `--gc`) enables the complete standalone GC v1 proposal, including
its typed-function-reference prerequisite. Tail Call remains separately enabled.
This GC increment does not select a complete Core 3.0 profile. Final
[exception handling](exception-handling.md) is independently enabled; Threads,
shared GC, components and WASI remain subsequent work.

The frontend supports recursive groups, final/open definitions, explicit
single-parent subtyping, function variance, struct width/field variance, arrays,
packed fields and the complete abstract heap lattice of the GC proposal. It
rejects forward references outside a recursive group, cyclic/final supertypes,
mutable-field covariance and unsupported heap families. Iso-recursive keys
retain group position, finality and declared parents; group-relative back edges
are distinct from already canonical external identities. Equivalent groups
share Store-local IDs. A binary ancestor index bounds nominal subtype queries
by 32 steps. Static validation retains identities/ancestry without retaining a
second copy of every composite shape.

All struct/array construction, accesses, packed extensions, fills/copies,
data/element operations, casts/tests, cast branches, i31 operations and
internal/external conversions pass through the existing compiler and explicit
frame driver. i31 and `ref.eq` emit native instructions on amd64 and arm64.
Other individual heap operations use checked shared helpers; native function
bodies are never interpreted as a fallback. Sources retain effects, traps,
fuel and byte offsets. Cast branch edges carry their validated refinements.

Guest frame values remain 16 bytes. Heap objects use compact little-endian
field layouts: packed i8/i16 take one/two bytes, scalar fields their natural
width, references eight bytes and v128 sixteen. Layouts are cached by canonical
type; reference tracing follows their typed fields. Object handles contain a Store-private slot and
31-bit generation. Released slots increment generations; an exhausted generation
retires its slot. i31 values are immediate tagged bits. Internalized external
identities and externalized guest references preserve roundtrip identity without
putting Go object addresses in numeric frames. Host extern identities use the
remaining 61-bit range and never wrap.

Each Store has a nonmoving tracing heap with independent object/byte quotas.
Collection traverses the joint object/instance graph iteratively, including
function references inside guest objects. Roots include host roots, active and
suspended frames, globals, tables, elements, construction arguments, initializer
stacks and unpublished instances still being initialized. Allocation safepoints
are shared by both backends; large bulk operations poll cancellation. Failed
allocation rolls back new objects and type groups, while escaped failed-instance
references retain their owners. Close clears backing values and host references.

`StoreLimits.MaxHeapObjects` and `MaxHeapBytes` set quotas. Defaults are 1,048,576
objects and 64 MiB. `StoreConfig.CollectEveryAllocation` is a stress mode that
traces before every allocation. `Usage.Core` reports live allocated object count
and accounted heap bytes; `Store.Collect` reclaims unreachable cycles.

Values returned to the host use the existing Root/Release contract. Function
bindings may supply an optional canonical function-reference type as the third
argument to `Store.NewFunction`; this preserves the nominal identity of open or
recursive function types. The explicit signature must match that type. Obtain
canonical types from exports/signatures of the same Store. For an externalized
guest object, `Store.ExternRef` returns an opaque reference Value sharing the
original root; `Root` duplicates ownership, and `NewExternRef` can preserve its
identity. Ordinary host payloads keep their existing behavior.

The pinned complete proposal corpus has 23 scripts, 964 commands and 802 binary
assertions per tier. One malformed text assertion is classified separately.
Additional tests force collection through cycles, constructors, host reentry,
vector fields, subtype calls, external roundtrips, failed-instance escapes,
resource exhaustion, generation retirement, cancellation and deep type chains.
[Fixture provenance](../../testdata/gc/README.md) pins producer/source/license
hashes. Passing these suites does not establish calibrated performance acceptance
or unimplemented Core 3.0 feature combinations.
