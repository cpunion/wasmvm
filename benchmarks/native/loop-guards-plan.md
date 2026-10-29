# Bounded single-block loop span guards

Baseline `77fe022a8f6e696cfeef07eb4237893a9e64e808`, engine `102623b`.
Recognize bounded, single-block memory loops with a local i32 induction variable,
a single +1 update after the accesses and a literal exit bound. Prove that the
address root advances with that induction variable while its base stays stable.
Reject internal CFG entries, edge moves, calls, growth, possible arithmetic traps,
multiple guard groups, unsupported recurrences and prefixes longer than 16 ops.
Keep analysis within the existing 256-instruction region budget.

At the original header, execute the canonical prefix and check the complete
remaining iteration span without trapping or changing mutable operands. Failure
continues at the exact original checked memory PC. Success enters a shared hot
body whose private backedge repeats the prefix without rechecking the span.
Original PCs and cold branches never target that private header. Preserve polls,
all physical exit slots, fuel, memory effects, exact trap order and memory32 wrap.
Use shared region selection and target arithmetic, not workload-specific code.

Test arbitrary cold/resumed PCs, guard failures, canonical prefixes, both edge
polarities, aliases, wrap/endpoints, memory rebinding, cancellation and resource
limits. Execute full Go/LLGo release semantics, >=90% measured-package coverage and
the final Linux amd64/macOS arm64 matrix. Compare exact base/head and pinned native
competitors across the existing twelve families, separating preparation/first/warm
phases. Retain regressions; local diagnostics cannot accept parity. Keep one active
PR and all existing merge gates.
