# SIMD immediate materialization scope

Baseline: `0bc1ba379b0b2f033c150e40f0c0f0fcc5c95079`. The shared vector
constant emitter currently builds both halves through scalar registers even
when a single SIMD instruction represents the exact 128-bit value. Select
baseline x86 zero/all-ones idioms and AArch64 modified MOVI/MVNI/FMOV immediate
forms. Keep arbitrary bit patterns on the existing
materialization path and retain the immutable register-pool priority mechanism.
No literal islands, new relocation framework, scratch register or CPU profile
is needed. This changes source-independent code selection, not fixture handling.

Recognize repeated byte/halfword/word values, shifted and complemented byte
fields, MSL fill patterns, byte-mask doublewords and exact FP32/FP64 immediate
expansions in bounded work. Preserve raw NaN/signed-zero/subnormal and unequal
half bits by rejecting unrepresentable candidates. A constant cannot overwrite
a live input or become a writable pool operand. Cold entries and all slots
remain independently observable; neither fuel nor polling moves.

Cross-check fixed encodings with LLVM. Test all 8-bit immediate values against
independently expanded lane bits, unsupported patterns, register destinations,
spills, literal-register exhaustion, every cold entry and exact code budgets.
Run the full Go/LLGo/target semantic matrix and retain per-package coverage over
90%. Reuse the 42 affected SIMD/scalar/call/memory families, exact own baseline
and pinned references. Preserve every phase and failed result. This optimization
and short diagnostics cannot close calibrated acceptance or PR #26.

Cached masks must also be consumed directly by compiler-generated floating-point
constants and NEON shuffle lookup operands, retaining each fallback's existing
scratch contract. Let AVX2 shuffles write their final allocated destination;
NEON TBL may name already-consecutive input registers directly. Preserve both
inputs through two-source lookup and keep immutable masks read-only. The same
shuffle/alias/spill/cold-entry suite and 42-family comparison cover these copy
reductions.
