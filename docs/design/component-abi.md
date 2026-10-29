# Component Canonical ABI foundation

This work uses Component Model revision
`d095e2f60ddae1f9a1e9b1e2fb547ec1114da677`. The synchronous value ABI is
implemented in `internal/component/canonical`, independently of the Core
interpreter/native backend. Component binary decoding is implemented separately
in `internal/component/decode`. The [type subsystem](component-validation.md)
checks expressions and lexical declarations; complete binary linking/validation,
runtime calls/post-return and WASI 0.2 worlds are not implemented by this
foundation. It does not establish M5 completion.

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

## Resource state

`Resources` holds nominal resource declarations and owned/borrowed handles.
Handles do not reuse IDs, preventing stale handles from gaining authority.
Live entries and the 28-bit index domain are bounded. `Take` consumes ownership
without a destructor; `Drop` removes the handle before invoking a destructor
outside the lock. A failed destructor cannot resurrect ownership. `Drain`
attempts all remaining destructors and returns the first error, after call
scopes have ended.

A `BorrowScope` pins lenders until synchronous call completion and records
lowered borrowed handles. Moving or dropping a lent resource fails. Ordinary
borrowed-handle drop has no resource destructor. Ending a scope with outstanding
borrowed handles reports escape and invalidates those handles. Reentry and
concurrent independent scopes are covered by race tests.

These resource primitives are not yet wired into own/borrow value conversion.
The codec returns unsupported for those conversions until the runtime supplies
an ownership-transfer transaction and scope lifetime. Call reentry guards,
post-return obligations, failed lift/lower cleanup and instance teardown belong
to that runtime integration; raw resource representation integers must not
become public forgeable ownership capabilities.

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

Next integration is component semantic validation and lexical index/type
spaces, Core module/instance composition, canonical function lift/lower and
post-return, then resource transactions and declared WASI 0.2.12 worlds. Async
function/task/future/stream support and WASI 0.3.2 follow separately. No component
or WASI support is inferred from passing only these value primitives.
