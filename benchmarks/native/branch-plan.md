# Native comparison-branch dependency scope

Baseline: `20d0a41368e25c71003e5d74f3965fb08791b8e2`. ARM64 already lets a
conditional branch consume its predecessor's integer comparison flags. On
amd64, the intervening cancellation check destroys flags, forcing a test of
the materialized boolean. Move that normal-path check immediately before the
comparison and forward its flags to the adjacent conditional branch when all
comparison slots use allocated registers. Preserve the canonical boolean for
all slots, edge copies and cold entries; do not remove its observable value.

The pre-comparison poll publishes the comparison PC and unchanged input slots.
The separate cold branch entry still polls at its original PC and reads the
retained low-i32 condition, independently of arbitrary entry flags/operands.
Moving a poll one instruction earlier preserves the maximum interval, including
an If at PC 256. Keep metered behavior and the ARM64 implementation unchanged;
spilled/ineligible amd64 comparisons retain the existing path.

Execute all integer predicates, widths, branch layouts, copied edges, source/
result aliases, spills, independent entries, canceled continuations and periodic
PC boundaries on both compilers/targets. Reuse the same 42-family paired
public-call/competitor diagnostic with the exact baseline. Preserve failed and
inconclusive measurements; this scope cannot close PR #26 or M2 acceptance.
