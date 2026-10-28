# Active program-counter scope

Baseline engine: `9769078806e174ace62788c74e0674896f7f56c8` (scope-only head
`567746a9df24b2fe5f0604c62b6f3089ae389c01`). Keep the currently executing PC
in a local loop variable. Publish the caller's next PC before pushing a guest
frame or exposing a host request; return/resume reloads that saved continuation.
Keep all original instruction indexes, fuel boundaries, trap offsets, polls and
backend continuations. No execution is observable between guest instructions
outside the serialized driver, and current root scanning uses typed slots.

Validate suspended host calls, nested/reentrant calls, frame/arena relocation,
branches, traps, fuel exhaustion, cancellation and workspace reuse. Execute the
complete Go/LLGo Core 2.0 interpreter and both native-tier corpus, retaining
per-package coverage and the final amd64/arm64 semantic matrix.

Reuse exact prior base builds and the existing interleaved phase protocol.
Compare the final source against the pre-optimization `6e061ac` baseline with
the same LLGo `8e45ed5e` compiler, plus Go and pinned reference engines. Include
the ten integer/call/memory/SIMD/FP families and two pinned C/Rust applications.
Retain the intermediate direct-dispatch and consumer-instantiation regressions;
neither zero helper calls nor selected point estimates close a performance gate.
Keep startup, preparation, first call, API/dispatch and all native rows separate.
