# Shared frame operand access scope

Baseline: `7f25ed948525c73bd68f56c9546b2d705fb143be`. The shared execution loop
creates a local getter closure for every dispatched instruction. Disassembly of
the pinned LLGo build shows receiver/base capture allocations, a closure-record
allocation and three indirect getter calls on the hot path. Replace this closure
with a directly called slot-reading method, preserving the NoSlot zero value,
current frame base, canonical bits, traps, fuel, host requests and cancellation.
This is one common frame driver for interpreter and native runtime exits.

Run the existing complete semantic/lifecycle matrix and per-package coverage.
Use exact-revision repeated-call diagnostics to separate reused interpreter/
native dispatch from public API and cancellation overhead. Inspect LLGo code
for removed capture allocation/calls; hosted zero Mallocs is not allocation
proof. Compare the affected integer, function-call and memory interpreter
workloads against pinned wazero interpreter, with startup/compile/first-call
phases retained. Keep native throughput regressions in the existing portfolio.
No narrow diagnostic closes interpreter, native or PR #26 acceptance.
