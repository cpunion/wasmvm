# Last-use integer operand slots

Baseline: `e68fe169e4bccba0042478b203db8b40bd183492`, engine `4f531dd`.
For nontrapping Core 2.0 scalar integer operations, release consumed SSA
operands before assigning the result, allowing its same-typed storage to be
reused. Keep the existing symbolic value/literal checks, typed free lists,
local borrowing and literal reference counts. A local, cross-block value or
still-live literal alias cannot be recycled. Calls and all other operations
retain their current allocation order, including suspended input/output slots.

Preserve instruction/source/fuel/PC boundaries and frame-count limits. No
native register or continuation contract changes. Check both operand aliases,
duplicate arguments, live old values, literal aliases, type changes, quotas,
cold native entries, all integer operations and full Core 2.0 Go/LLGo semantics.
Run the final four-target matrix with per-package coverage above 90%.

Compare exact base/head with identical pinned Go and LLGo release compilers,
the same twelve integer/call/memory/SIMD/FP/C/Rust families and references.
Keep preparation/first/warm phases separate and preserve all regressions.
Report frame/code sizes and warmed results; existing competitor and calibration
gates remain required before merging. Reuse the current harness and one active
PR rather than adding an executor or benchmark framework.
