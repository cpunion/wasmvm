# Native comparison-branch dependency scope

Baseline: `20d0a41368e25c71003e5d74f3965fb08791b8e2`. ARM64 already lets a
conditional branch consume its predecessor's integer comparison flags. On
amd64, the intervening cancellation check destroys flags, forcing a test of
the materialized boolean. Replay the pure comparison of unchanged allocated
input registers after that existing poll, then branch directly on its flags.
Keep the canonical boolean for slots, copied edges and cold entries. Predicates
that overwrite either input, or require spilled operands, retain the old path.
Metered code and the ARM64 implementation remain unchanged.

Keep the original poll positions and continuation PCs. Moving a poll before
the comparison could expose slots that earlier copy fusion legitimately omitted
until the original poll. Replaying flags preserves those proofs and snapshots
without a second poll scheduler or a weaker slot-observability contract. A cold
branch entry still checks its retained low-i32 condition, independently of entry
flags and the original operands. Periodic polls at PC 256 remain intact.

Execute all integer predicates, widths, branch layouts, copied edges, source/
result aliases, spills, independent entries, canceled continuations and periodic
PC boundaries on both compilers/targets. Reuse the same 42-family paired
public-call/competitor diagnostic with the exact baseline. Preserve failed and
inconclusive measurements; this scope cannot close PR #26 or M2 acceptance.
