# Direct numeric globals

Baseline `ccbcbd5868de49651402737ccef0e4e9cc8b13e5`, production engine
`4304f73fb52e2e946e6f2b8877d90e4b87d7ee4c`.

Lower validated i32/i64/f32/f64/v128 global.get/set directly on amd64 and arm64,
in precise and unmetered entries. Keep reference globals on the existing helper
and root protocol. Extend the execution-owned native runtime view with global
cell pointers; deduplicate imported aliases by canonical Global identity. Cache
bindings, snapshot only the current function's referenced globals on entry, and
publish their values after every native exit before servicing helpers, calls,
traps or cancellation. Host reentry and cross-instance calls must observe those
writes. Clear cached Go bindings on suspension/close. Never pass Go addresses
through the native ABI or retain stale bindings for functions without globals.
Cache immutable cell pointers in spare arm64 registers, sharing register pressure
with the existing table metadata cache. Preserve scalar canonicalization, exact
v128 bits, source PCs, fuel, code limits and cold continuations.

Test all numeric types, mutable aliases, precise high words, missing runtime views,
reference fallbacks, host/global interactions, trap order and lifecycle cleanup.
Run Go and pinned LLGo release semantics, >=90% measured-package coverage and the
full Linux amd64/macOS arm64 CI matrix. Compare exact base/head and pinned native
competitors for affected global workloads and the existing portfolio, separating
preparation, first and warm execution. Retain failures and regressions; local
diagnostics do not satisfy cross-target performance gates. Keep the single active
PR and do not merge incomplete or slower profiles.
