# Component Canonical ABI foundation

This work uses Component Model revision
`d095e2f60ddae1f9a1e9b1e2fb547ec1114da677`, with the separately pinned
[October 30 cm64 and empty-buffer corrections](../../testdata/component-cm64/README.md).
The synchronous value ABI is
implemented in `internal/component/canonical`, independently of the Core
interpreter/native backend. Component binary decoding is implemented separately
in `internal/component/decode`. The [type subsystem](component-validation.md)
checks ordered Core/component composition, interface subtyping, resource
substitution, synchronous canonical declarations and boundary type names.
Complete binary validation, runtime calls/post-return and WASI 0.2 worlds remain
required by this foundation. It does not establish M5 completion.

## Representation and limits

Resolved types in `internal/component/model` retain declaration order and
nominal resource IDs. Type graphs remain immutable while a layout builder uses
them. Layouts cache shared DAGs and bound aggregate metadata, nesting, size,
alignment arithmetic and flattened counts. Large shared graphs do not expand
into exponential flattened slices. Memory32 and Memory64 use their actual
canonical pointer widths. Synchronous signatures retain 16 direct parameters
and one direct result; indirect result pointers are returned by lifted guest
functions and appended to lowered host functions' parameters.

Scalars preserve signed bit patterns, canonicalize NaNs, treat nonzero guest
booleans as true and reject invalid Unicode scalar values. Records, tuples,
variants, flags/enums, option/results and nested lists support checked memory
and direct flat conversions. Variant coercion reinterprets bits and zero extends
joined 32-bit payloads; it does not numerically convert floats. Guest-owned
strings/lists are copied through memory capabilities, never retained as slices.

Transfers implement UTF-8, UTF-16 and latin1+UTF-16. Invalid encodings, unmatched
surrogates, alignment, pointer/count overflow and invalid discriminants fail.
Memory ranges are checked before scratch allocation. Per-call transfer bytes,
decoded value count and nesting have independent limits. Host-value validation
precedes memory writes and realloc. Realloc can reenter Core code; no borrowed
memory pointer crosses it. Guest allocator/memory failures can preserve earlier
allocation effects; the runtime must retain instance ownership through cleanup.

Packed byte transfers optionally use `canonical.StringMemory`: an immutable
snapshot on read and a direct immutable-byte copy on write. The VM Caller
adapter applies the same active, nonconcurrent, same-Store capability checks as
slice transfers. Private memory copies directly into the final string or guest
buffer; no guest view escapes mutation, growth, reentry or close. Domain-backed
shared memory retains its lock, and native shared memory retains its existing
byte-wise synchronized transfer implementation. Slice-only memory adapters
remain supported. Logical transfer charges, value limits, complete-range
prechecks and post-copy cancellation checks remain identical for both paths,
even when the immutable path avoids scratch allocation.

## Resource state

`Resources` holds nominal resource declarations and owned/borrowed handles.
Host capabilities do not reuse IDs. Guest tables use the specified LIFO free
list; publication generations distinguish provisional ownership after reuse.
Live entries share one bounded Store quota, and each table has a bounded 28-bit
index domain. Representations preserve all 64 bits for cm64 resource types. `Take` consumes ownership
without a destructor; `Drop` removes the handle before invoking a destructor
outside the lock. A failed destructor cannot resurrect ownership. `Drain`
attempts all remaining destructors and returns the first error, after call
scopes have ended.

A `BorrowScope` pins lenders until synchronous call completion and records
lowered borrowed handles. Moving or dropping a lent resource fails. Ordinary
borrowed-handle drop has no resource destructor. Ending a scope with outstanding
borrowed handles reports escape and invalidates those handles. Reentry and
concurrent independent scopes are covered by race tests.

The synchronous runtime now executes resource.new/rep/drop using these tables
and the active Core Caller, including nominal bindings across child instances.
The runtime now wires own/borrow call conversions through the same flat/memory
codec. Opaque host capabilities, prevalidated moves, borrowed scopes, nominal
host bindings and failure cleanup supply that ownership contract. Resource
representation integers never become forgeable public ownership capabilities.
Standalone resource values/starts share the same ownership transaction.
Local cm64 borrows preserve the full representation as described below; the
original pinned Python layout reference covers only MVP resources.

## Direction-aware cm64 borrows

The ABI correction at `fed10b0b823b82553dd8363f16a640816f5eac65`
distinguishes the component defining a resource and the direction of conversion.
Lowering a borrow into its defining component uses the resource's actual i32 or
i64 representation for its flat type, size and alignment. Every own value and
every foreign borrow remains an i32 handle. Resource representation width is
independent of the canonical memory's address width.

One shared layout builder accepts an optional local-representation resolver;
ordinary lifting retains the original i32 handle layouts. Validation and runtime
use the same rule, including records, variant joins, lists and indirect parameter
spills. Runtime layout caches belong to their namespace and retain no compilation
plan. Type declarations enforce the worst-case size below 2^28 bytes, accounting
for unknown imported representations while retaining known i32 sizes.

Lenders stay pinned through async suspension and cross-component cancellation.
Failed allocation, memory writes and guest traps release acquired lends through
the existing ownership transaction. No representation integer is a public
ownership capability. Nine authored fixtures execute on both backends; the
independent wasm-tools 1.262.0 validator still rejects the new local i64 flat
signature. Its result is a recorded reference limitation, not a passing oracle.

The cumulative correction also checks typed stream/future buffer alignment and
bounds when the element count is zero. The memory endpoint is valid; an address
beyond it or a misaligned pointer traps. Untyped unit buffers ignore their pointer.
The original official zero-length script is preserved alongside a declared
two-address adaptation; all commands and expected outcomes remain unchanged.

## Evidence and remaining integration

Go and pinned LLGo execute the same package tests. The unchanged official
Python reference generates 1,024 independent memory/flattening layout cases;
its source, seed, outputs and license are in
[`testdata/component-canonical`](../../testdata/component-canonical/README.md).
This is not a count of complete official component tests. Authored tests add
malformed memory/string/value cases, quotas, resource lifecycle and concurrency.

The binary decoder preserves ordered/repeated sections, lexical component/type
scopes, Core types/modules/instance declarations, aliases, canonical operations,
imports/exports and length-delimited values. It shares Core type/descriptor and
module grammar with the existing Core decoder; embedded module slices borrow
the enclosing decoder's owned snapshot. All nested syntax uses the same aggregate
node/name/byte quotas. This includes parsing pinned async/threading opcodes, and
does not establish their runtime support. Value-definition bytes are retained
for type-directed validation rather than treated as already valid values.

The complete pinned binary script adds 132 grammar commands: 80 malformed
rejections and 52 decoded definitions. Sixteen of those definitions require
later semantic rejection. This is not a claim that they passed validation.

Value/start definitions now have bounded static plans and a literal decoder
that produces the same canonical value representation. Next integration
executes the validated Core/component composition plans, canonical function lift/lower and post-return,
then resource transactions and declared WASI 0.2.12 worlds. The [synchronous runtime](component-runtime.md) now connects the non-resource call subset, including realloc and post-return; this does not imply full Component Model or WASI support. Async
function/task/future/stream support and WASI 0.3.2 follow separately. No component
or WASI support is inferred from passing only these value primitives.
