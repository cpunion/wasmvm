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
