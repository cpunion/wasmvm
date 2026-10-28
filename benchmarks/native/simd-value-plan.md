# Allocation-free pure SIMD evaluation scope

Baseline: `43bb8a7839b1252509b71c5a8bfee2e0db66b413`. Pinned LLGo disassembly,
including `-O3 -lto=full`, allocates a 112-byte copied instruction and a 16-byte
addressed result at pure SIMD Eval entry. Borrow the already immutable
instruction and return updated lane values from put instead of taking the
result's address. Preserve all fixed-SIMD opcodes, immediate bits, lane order,
NaNs, signed zeros, saturating semantics and caller ownership. Apply the same
helper contract to floating-point lanes, without duplicating evaluator paths.

Retain the existing full pure-SIMD oracle, Core 2.0 backend/compiler/target
matrix, limits/lifecycle checks and >=90% package coverage. Compare existing
shuffle and popcount workloads with exact base/head and pinned wazero
interpreter; retain integer, call, memory and call-control regressions. Keep
public API, reused dispatch, startup/preparation/first call and both LLGo
compiler-option sets separate. Use the same options on each revision, check
alternate inputs and inspect removed allocation calls. Hosted Mallocs zero is
not proof of allocation freedom. No short diagnostic closes performance,
PR #26 or milestone acceptance.
