# Per-frame dispatch metadata scope

Baseline: `6d19583ae04f1ad890105b9bacf73a6cc1f73ecf`. The shared frame driver
reloads function/code/instance/frame-base metadata after every instruction.
Keep one inner dispatch loop per active frame and refresh its immutable metadata
only when entering/leaving a guest frame or resuming a Run invocation. Preserve
the same explicit frame machine for interpretation and native runtime exits.

Retain every original PC, source/fuel boundary, cancellation poll, typed host
request, root and failure cleanup. Frame slices and native value arenas may move
on push: never retain a caller-frame pointer or stale value view across growth.
Keep the existing module/source ownership contract and fresh runtime state.

Test nested frame/slot growth and arena relocation explicitly, then run existing
complete Core 2.0 semantics and lifecycle checks under both host compilers and
backends. Keep package coverage >=90%. Compare exact base/head integer, calls,
memory, SIMD and control workloads against pinned wazero interpreter, separating
API/dispatch, startup/preparation/first call and compiler-option rows. Slower or
inconclusive results stay visible and block acceptance; no small diagnostic
closes M1/M2 or PR #26.
