# Final exception handling

`wasm.ExceptionHandling` (CLI `--exception-handling`) enables final tags,
`try_table`, `throw`, `throw_ref`, `exnref` and `nullexnref`. Legacy `try`,
`catch`, `delegate` and `rethrow` encodings are rejected. Default Core 2.0
remains strict. GC casts on exception references additionally require `wasm.GC`;
Tail Call remains independent. This feature does not complete the combined
Core 3.0 or Threads acceptance gate.

Tags have generative Store-local identity and an invariant canonical function
payload type with no results. Imports compare the entire nominal type identity,
including recursive-group position. Re-exported tags preserve identity.
`StoreLimits.MaxTags` bounds live tags; the default is 1,048,576. Binary tag
instances share canonical signatures rather than copying each payload schema.

The compiler lowers protected blocks into shared immutable handler chains.
Four catch kinds retain source order and explicit payload destination slots.
Exceptional edges participate in SSA reachability and dominance verification;
catch parameters receive runtime values, rather than fabricated SSA operands.
Node and verification budgets also bound nested exceptional control graphs.
A binary search finds protection for a throw PC. Unwinding inspects guest
frames iteratively, without host recursion or host-language panic/recover.
Normal execution performs no dynamic handler push/pop. Suspended callers save
the call PC separately from their return continuation. Tail calls retire their
frame and its handlers before dispatch. Traps, resource failures, cancellation,
ordinary host errors and host panics do not become Wasm exceptions.

Both native targets generate their function bodies and exit through checked
individual throw helpers. The shared frame machine selects the catch target
and resumes native code at the existing cold entry. There is no whole-function
interpreter fallback. Every executed throw and try instruction preserves its
source/fuel boundary; catch metadata itself has no additional fuel charge.

Exception objects use the existing generation-checked, quota-limited guest
heap. Payloads use compact typed layouts, including v128 and all reference
families. `throw_ref` preserves exception identity. Joint tracing follows
nested exception/object/function payloads and their tag owners, including
unpublished instances whose start function threw. Allocation operands,
active/suspended frames, host roots, globals and tables participate in the
same collector. Object/byte quotas, stale handles and close retain their
existing behavior. Allocation currently materializes an exception even when
no catch requests its reference; lazy materialization and direct native unwind
are local optimization follow-ups.

The embedding API provides `Store.NewTag`, `Extern.TagType`, `SameTag`,
`Store.NewException`, `Caller.NewException`, and `Store.Exception(Value)`.
An uncaught guest exception returns `*vm.Exception` with
`errors.Is(err, wasm.ErrException)`. `Exception.Reference` borrows its root;
`Payload` creates independently owned payload Values and returns the tag.
Release reference payload Values and the Exception when finished. Returning an
Exception from a host callback throws into the guest; an existing caller-owned
exception keeps its root until explicitly released. `Caller.NewException` uses
a callback-scoped root, handed off before temporary roots drain. Cross-Store,
released and null exception references cannot be thrown through the host API.

The complete pinned final-spec feature inventory contains nine scripts: all
five `exceptions/` scripts plus `imports`, `exports`, `instance`, and `ref_null`,
which contain EH forms. Each tier executes 482 commands and 305 binary
assertions. Eighteen text-parser assertions remain separately classified.
No command or assertion is removed. Source, producer, binary and license
hashes are in [the corpus manifest](../../testdata/exception-handling/README.md).
Tests additionally cover forced GC, failed-start payloads, host reentry and
rethrow, private/imported tag identity, deep unwinding, tail transfers, quotas,
released references, cancellation and panic cleanup.
