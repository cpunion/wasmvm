# Direct checked heap field addressing

Baseline: `07fd46fc662d245f7f0bb67e934bf5b1d27a75ed` (PR #133).
This experiment changes native instruction selection for checked GC fields.
It does not close controlled performance or release acceptance.

## Mechanism

Struct field loads/stores use the checked payload register directly, with the
field displacement encoded in the access. AMD64 supports disp8/disp32, high
registers, SIB bases and the mandatory zero displacement of RBP/R13. ARM64 uses
scaled imm12 for aligned offsets or unscaled imm9 for small packed fields;
large unencodable offsets retain the scratch address calculation. Scalar,
reference, floating-point and v128 fields share the address contract. Array
bounds/indexing and allocation stores keep their existing behavior.

Full handle equality on cache hits and null/index/generation checks on misses
remain mandatory. Cache count, register budgets, relocation/reentry resets,
source offsets, cold continuations, cancellation, fuel and complete-frame
publication remain unchanged. No unchecked execution profile is introduced.

## Validation and measurement scope

Existing heap tests exercise every field width, packed sign extension,
aliases, spills, invalid/null/stale handles, moved payloads, metering and cold
entries. Added real execution cases cover displacement boundaries, packed
unaligned fields, large ARM64 fallback offsets, both cache register inventories
and sentinel bytes around each write. Both target encoders are compiled locally;
actual amd64 execution requires Linux CI, while local execution uses arm64.

Use the existing pinned Go/LLGo release profiles and production-policy benchmark
protocol. Compare all four original GC cases and the separate sustained GC
cohort against this baseline and every applicable named competitor, retaining
unsupported combinations and all A/A controls. Native C is absent from the GC
portfolio; it must not be invented as a comparable checked-GC reference. Keep
startup, preparation, first call, warm execution and cleanup separate. Do not
pool platforms, cohorts or historical campaigns. Final exact-head results and
raw artifacts belong in the contribution PR; unresolved controls remain open.
