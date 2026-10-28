# Packed integer SIMD arithmetic scope

Baseline: `1e296a8d0f7095b32d294590e754cbba0e89b542`. Pure SIMD no longer
allocates instruction/result objects for integer operations, but generic lane
extraction/insertion and per-lane kind dispatch still limit throughput. Replace
modular add/subtract for all four integer lane widths with two-word packed
arithmetic that isolates lane carries/borrows. Compute i8x16.popcnt using
independent two-/four-/eight-bit counts within each byte. Preserve wraparound,
all source/operand aliases, immutable inputs and the existing single evaluator.
Do not replace saturation, floating-point or other operations with new subsets.

Check cross-lane carries/borrows and byte popcounts against independent scalar
oracles, including exhaustive byte pairs, adjacent boundary patterns and wider
random values. Retain full pure-SIMD/Core 2.0 semantics under both compilers,
all supported targets/backends, and >=90% per-package coverage. Compare exact
base/head with pinned wazero interpreter on shuffle/popcount plus the existing
scalar/call/memory/control regressions, using the same compiler flags on both
revisions and separating phases and API/dispatch. Retain default and -O3/full
LTO rows and component allocation/time diagnostics. Native emission is
unchanged; this optimization does not accept native, PR #26 or milestone gates.
