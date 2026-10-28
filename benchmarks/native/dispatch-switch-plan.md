# Opcode switch dispatch scope

Baseline: `1e0b9175c8c5a886a49a50099a81efac515c83e7`. The current boolean
switch tests many opcode predicates before scalar numeric evaluation. Use a
value switch over validated scalar/control opcodes so host compilers can form
compact jump/range dispatch. Keep prefixed SIMD and bulk operations in the same
shared driver, with their existing checks and effects. No new opcode, bytecode
layout, handler closure, execution backend or instruction fusion is introduced.

Retain every original instruction/fuel/source boundary, cancellation poll,
frame transition and trap. Run full Core 2.0 semantics and moving-arena tests
under Go/LLGo; keep executable-package coverage >=90%. Compare final exact source
base/head against pinned wazero interpreter with integer/call/memory/control,
SIMD and floating families, separating phases, API/dispatch and compiler flags.
Retain the first frame-cache candidate's mixed results rather than presenting
this refinement as the same measured revision. Missing/slower/inconclusive
results block merge acceptance.
