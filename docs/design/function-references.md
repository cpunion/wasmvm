# Typed Function References

`wasm.FunctionReferences` enables the standalone acyclic, invariant proposal.
The strict default Core 2.0 profile and the separate Tail Call flag remain.

The decoder retains module-local indexed reference types, nullability, explicit
table initializers and all five new instructions. Validation normalizes earlier
equivalent function types, checks subtype-compatible operands and immutable
imports, and tracks nondefaultable local initialization in each structured scope.
Recursive/forward type references and GC heap types remain rejected.

Full reference metadata lives in `wasm.ValueType` words in static schemas.
Shared physical slots project reference refinements onto their function/external
storage domains. No guest value or native frame grows: values remain 16 bytes.
Null branches become two-edge SSA terminators; only their nonnull edge refines
the condition type. Both native targets emit full-width null tests. Null
assertions retain trap effects during copy/loop optimization.

At instantiation, a Store canonicalizes acyclic function types in order before
linking imports. Keys retain all reference metadata. Equivalent types from
different modules receive the same Store-local ID. Failed allocation rolls back
both interning tables; function identities and references keep existing lifetime
rules. Mutable imports are invariant; immutable globals permit covariance.

`call_ref` and `return_call_ref` use the shared explicit frame driver and native
checked exits, preserving fuel, cancellation, pending roots, host reentry and
proper tail retirement. Native code does not interpret function bodies. General
reference-call boundary cost remains a local optimization task; the complete
feature need not wait for that optimization under the owner's advancement rule.

The pinned proposal corpus contains all 19 root scripts using typed/non-null
references or the new instructions, including local initialization and linking.
Each tier runs 1,174 commands / 978 binary assertions. Twenty-nine malformed text
assertions remain separate from binary acceptance. The source, license,
producer flags, SHA-256 inventory and wildcard expectation correction are in
[testdata/function-references](../../testdata/function-references/README.md).
Public API tests additionally cover typed host tables/globals, import type-index
reordering, wrong/foreign/stale references, full-width external-reference IDs,
host collection/reentry, null traps, tail recursion, fuel and cancellation.

Recursive groups, explicit GC subtyping, structs, arrays, casts, exceptions,
shared memory, Component Model and WASI remain subsequent capabilities. This
increment does not complete M1/M2 performance acceptance or M3/Core 3.0.
