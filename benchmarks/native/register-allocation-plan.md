# Loop-weighted shared scalar and runtime-register allocation

Baseline engine `c921b5de933aea8d5a6724ec3eff106c2525f0bd`. The table-copy
loop improves from 55 to 49 static hot-span instructions, but still performs
twelve runtime-view/metadata loads and two frame reads per iteration. Cold
post-loop verification occupies fixed homes, leaving no spare metadata cache.
The same pinned Wasmtime loop forwards a stored table value into its subsequent
get, omitting one element read; its default policy also lacks our interrupt poll.
Keep the measured gap and these policy distinctions explicit.

Integrate scalar home, immutable runtime-word and literal priorities within the
existing target register inventory, weighting uses by CFG loop structure.
Avoid assigning every register to a cold home before considering hot metadata.
Use bounded deterministic analysis and no fixture-specific priorities. Preserve
all physical exit slots, arbitrary cold entries, exact fuel/source PCs, reference
spills and root maps. Metadata caches reload on every entry. Reserve constrained
operands and platform registers; preserve x86 feature and ABI rules.

Test nested/multiple/irreducible backedges, cold tails, aliases, exhausted pools,
all index/reference spill combinations, helper/host rebinding and code limits.
Run the complete Go/LLGo release semantic matrix and coverage threshold. Retain
all twenty-one workload families, separate every phase, and compare exact source
base/head against the same pinned competitors before any merge. Keep regressions,
all-target calibration and broader M2 acceptance open. Store-to-load forwarding
and global live-range splitting are distinct optimizations requiring their own
proofs; a register-priority change does not implement them implicitly.
