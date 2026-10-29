# Guarded memory access groups

Baseline `a35f46ca69221d5f16a85636e2ce38bf0fd67461`, engine `342676e`.
Extend the existing unmetered, positive-minimum-memory regions to group at
least three unproven scalar/SIMD accesses derived from an unchanged i32 local
and nonnegative constant additions. Preserve generic checked emission for all
other cases and the precise tier. This changes optimization, not Core support.

Track local versions and physical-slot writes with sparse, bounded region maps.
A runtime guard reads the first access's original address operand, avoiding
synthetic uses or hidden materialization requirements. For first arithmetic
bias b, minimum arithmetic bias m and maximum effective end E, require address
>= max(0,b-m), address <= 2^32-(E-b), and address+(E-b) <= bound memory length.
These conditions exclude wrap within the group while allowing common earlier
wrap. Every original memory operation still executes in order. An unsuccessful
guard continues through checked cold instructions from that same first access;
it neither reports a speculative trap nor replays the prefix.

Guard positions bound instruction fusion. Keep every original checked cold PC
and a separate first-region-PC fallback. Reset facts at CFG/helper/growth/poll
boundaries. Preserve all frame values, trap source/order, fuel, cancellation,
binding validation and cleanup. Share memory widths and existing emitters;
add no runtime executor or workload-specific path.

Cover all Core 2.0 scalar/SIMD memory widths, offset/alignment/lane cases, span
ends, i32 wrap, local/slot invalidation, pending stores before later traps,
first/interior cold entries, growth, zero memory, quotas and cancellation.
Run full Go/LLGo release semantics, per-package coverage >=90%, and final
Go/LLGo x Linux amd64/macOS arm64 CI. Compare native API/dispatch base/head
under identical pins/flags with the existing twelve families and native
references. Reuse unchanged interpreter timing evidence; its full semantic
matrix remains required. Keep preparation/first/warm separate, all failures
and regressions visible, and the current parity/calibration gates unchanged.

Retain guard-only candidate `5073a99` and its measured results. Refine guarded
sites to address stable locals directly: strengthen the lower guard to b so
the original additions cannot have wrapped, and fold their biases into memory
offsets. Wrapping cases continue through the checked path. Remove only dead
i32 constant/add temporary writes, using bounded backward liveness with actual
remapped reads; reset canonical-state requirements at guards and possible traps.
Keep every memory operation and every checked cold PC. This reuses the existing
region emitter rather than introducing a separate optimizer or executor.

Allocate fixed registers using the planned emitted operations: ignore deleted
temporaries, count remapped local reads and retain the first guard's original
address read. Build regions before register allocation; preserve the same
register/cold-continuation contract and generic allocation for other functions.
