# Direct checked tables and function references

Baseline engine `a26c57cebe8bc67734304e86c781d9b7fa020f65`. Complete direct
Core 2.0 table.get/table.set and ref.func lowering on amd64/arm64 in both native
tiers. Keep bulk table operations and growth on the existing explicit driver.
Reuse the execution-owned runtime view's immutable function IDs and per-table
length/native-base pairs. Store-owned native table cells already contain stable
reference handles, not Go pointers; imported aliases share the live allocation.
Do not copy table contents or create a parallel table representation.

Check unsigned i32 indices before forming any element address; empty tables must
trap without dereferencing a null base. Preserve exact source offsets, ordering
of preceding global/memory/table effects, fuel and arbitrary cold continuations.
Reject missing runtime views before guest execution. Every helper/call handoff
refreshes length/base bindings, including host growth and same/cross-instance
reentry. Reuse the shared spare-register metadata cache without increasing spills.
Function references load the current instance's immutable handle by validated
index, never a machine address. Reference values remain typed frame/table roots;
all exits publish the existing conservative reference map before guest collection.
No collection or host callback may read an unspilled native register.

Test funcref/externref/null values, declared imported/defined functions, table
aliases, zero/exact/large/out-of-bounds indices, handle high words, distinct table
types, root ownership and collection, host replacement/growth/reentry, traps,
fuel, cancellation, code/resource limits, cache pressure and every original PC.
Reject Go-backed tables at the native boundary. Cover direct access combined
with the numeric-global view and guarded indirect expansion, including multiple
imported/defined table and function index spaces.

Declare the existing globals/tables loop (three remaining table/reference helper
exits), isolated funcref table get/set loops with dynamic indices, and isolated
ref.func handle production. Keep reference outputs live through table mutation
and independent checksum validation. Use wasm-tools 1.258.0, owned fixtures,
pinned source/binary hashes and scalar oracles. Compare exact base/head public
API and dispatch against the pinned wazero/Wasmtime/Wasmer configurations; retain
all eighteen previous families and startup/compile/instantiate/first/warm phases.
Execute full Go/LLGo release semantics and the final Linux amd64/macOS arm64
matrix, with >=90% coverage per measured executable package. Missing, slower or
inconclusive controlled performance evidence still blocks merging and M2
acceptance. Keep the single active PR and preserve all earlier failures.

The owned fixtures are `table-roundtrip`, `table-copy-handles` and `ref-func`,
with 100,000 primary iterations. Weighted indirect calls after each loop verify
reference identity, not only nullness; this verification remains inside guest
execution timing. Alternate inputs include partial tables and both null/function
cases. Keep all baseline/candidate/reference results, including fixed-tail and
API overhead, and do not subtract verification or startup costs from samples.

The first direct lowering retains a table-copy gap (LLGo 1.725 versus Wasmtime
1.230 ns/iteration) despite large improvements over the old helper path. Select
allocated reference registers directly. On arm64, compare the full length with
a zero-extended i32 home and use UXTW scaled addressing, preserving raw upper
index words without scratch copies. Keep spilled operands and precise entries
on the same checked semantics. Add pressure that forces index/value/result homes
to spill, and preserve the first diagnostic rather than replacing it.
