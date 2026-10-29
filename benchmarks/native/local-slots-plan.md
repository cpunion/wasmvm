# Local slot coalescing scope

Baseline: `6fcd4e4d2e6c12431d810c42478c04740ce4c3a9`, engine `f1513a8`.
Share a local.get/local.tee result with its typed local slot when the value is
block-local and its last use precedes the next write of that local. Keep all
original instructions, source/fuel charges and PCs. Keep cross-block values
dedicated and never return local slots to the temporary free lists. An explicit
self-copy in native emission has no data effect. Apply this common slot layout
to interpretation and both native tiers rather than adding another executor.

Use bounded linear scans, including across functions with many locals/blocks;
avoid initializing a local-sized table for every block. Extend the independent
symbolic slot checker to model local value versions/copies. Cover mutations
while old reads remain live, tee inputs/results, repeated reads, edges, loops,
calls/suspension, references/vectors, quotas and zero-initialized reuse.

Run the complete Go/LLGo release Core 2.0 corpus and final amd64/arm64 matrix,
retaining per-package coverage above 90%. Compare exact source base/head with
identical host compilers/flags and pinned references, reusing the twelve scoped
integer/call/memory/SIMD/FP/C/Rust families and existing phase/interleaving
protocol. Record slot/code sizes, preparation and warm results separately;
smaller frames legitimately change slot-quota fit, not frame-count or fuel
policy. Preserve failed/regressed candidates. No merge or milestone acceptance
without its applicable competitor/calibration gates.
