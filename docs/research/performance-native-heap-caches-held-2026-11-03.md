# Held native heap-cache experiments: 2026-11-03

Neither production candidate is selected. The retained indirect-reference tests
run against existing production `a176cb5`; they do not require either cache change.
Pinned Go 1.27.3, LLGo `8e45ed5` with `-O3 -lto=full`, Wasmtime 48.0.1,
Clang 22 O3; shared M4 Max/macOS arm64. Both campaigns use the four unchanged
GC kernels, 10,000 allocation iterations and one million for other kernels.
Five fresh process blocks, two warmups/seven interleaved samples, independent
alternate checks and original baseline A/A aliases; all preparation/first/warm/
cleanup phases remain separate. No cross-campaign percentages are composed.

## Extra cache homes alone

`b8002c5` lets type-keyed caches use idle callee-saved integer registers already
preserved by the entry ABI. Scalar/metadata allocation stays fixed; complete
null/index/generation checks remain on misses. Exact configured body witnesses
change only arm64 unmetered indirect access: one cached type becomes two.
Other GC cases, amd64 and all metered bodies remain byte-identical.

Warm ns/source iteration in this campaign:

| Kernel | Go base → head | LLGo base → head | Wasmtime | Ordinary C |
| --- | ---: | ---: | ---: | ---: |
| i31 | 2.277542 → 2.292292 | 2.235292 → 2.258208 | 3.019 | 2.144 |
| struct access | 2.250209 → 2.274917 | 2.434041 → 2.426084 | 1.939 | 1.777 |
| allocation | 72.6500 → 76.0875 | 92.2291 → 92.5792 | 175.900 | — |
| indirect access | 6.433709 → 6.544291 | 6.405750 → 6.485000 | 3.577 | 3.905 |

Target LLGo is +1.24%, versus A/A +0.31%; Go is +1.72%, versus A/A +0.01%.
The unchanged LLGo struct-access A/A changes -7.1%, Go allocation +4.4%.
Host one-minute load is 20.8→17.9. This provides no adoption evidence.
140 VM/15 C observations, 28 VM/three C alternate checks, eight explicit pinned
wazero/Wasmer GC rejections, 31 summaries remain complete.

## Two complete cached handles per varying type

`aea69c5` uses two entries when aggregate-reference reads suggest alternating
receivers. This bounded hint is only profitability, not a validity proof.
ARM64 CMP/CSEL/CCMP compares full handles and chooses an address; null traps,
checked misses, replacement and entry resets remain. Cached addresses never
survive a host/GC entry. One type now has two ways; the indirect body grows
3376→3428 bytes. All other captured bodies remain byte-identical.

This second campaign measures original production, `b8002c5`, and `aea69c5`
together, without pooling earlier observations:

| Kernel | Go base → final | LLGo base → final | Wasmtime | Ordinary C |
| --- | ---: | ---: | ---: | ---: |
| i31 | 1.959292 → 1.933458 | 1.881917 → 1.878125 | 2.562 | 1.874 |
| struct access | 2.190958 → 2.410708 | 2.472917 → 2.222500 | 1.951 | 1.939 |
| allocation | 79.1125 → 76.6625 | 91.4375 → 91.7625 | 176.800 | — |
| indirect access | 10.095541 → 11.067292 | 10.082792 → 10.960125 | 5.582 | 6.077 |

Target LLGo is +8.70%, versus A/A -0.17%; Go is +9.63%, versus A/A -0.36%.
Middle indirect Go/LLGo is 10.3150/10.1661. Unchanged struct-access controls
vary about ten percent, including LLGo A/A -10.9%. One-minute load is
116.6→75.5. Controlled attribution is not established; the candidate fails
to demonstrate benefit and remains held. All 180 VM/15 C observations,
36 VM/three C alternate checks, eight rejections and 39 summaries remain.
The C indirect control uses two ordinary stack objects and a pointer array at
the same inputs; normal optimization is allowed. C has no VM lifetime, quota,
cancellation, typed-continuation or handle-validation contract; no GC allocation
equivalent is invented. LLGo/Go timing ratios alone are not compiler defects.

## Validation and next mechanism investigation

Each source passes all 31 Go packages/thirty coverage gates, full-LTO LLGo
codegen/nativeexec/VM, scoped race/no-cgo/vet. Additional indirect-reference
tests cover both keys, replacement, every cold PC, moved payloads between
entries, warmed null/stale/forged/out-of-range handles and both cached receivers'
index traps, with complete valid/trap-frame comparison. They also pass on
selected original production with Go and LLGo. Initial missing-corpus setup,
incomplete synthetic inventory fixture and a terminal-BadState frame assumption
are retained. The latter fails identically on baseline; the ABI requires frame
publication for resumable exits/typed traps, not invalid terminal internal state.
The early allocation inventory omitted constructor configuration; exact configured
body witnesses, rather than that draft, establish allocation-body equivalence.

All three original native bodies and unchanged C disassembly are retained.
They expose redundant constant materialization before ARM immediate operations:
`i32.const 1` writes a register immediately overwritten by UBFX. The existing
constant/integer fusion is restricted to amd64. Investigate sharing that proven
selector with ARM, preserving independent cold PCs, observable live constants,
full frame state and polling. More cache state is not the next selected action.
No new LLGo issue is inferred from these JIT/backend observations.

[All campaigns, canonical coverage, commands, code witnesses and disassembly](evidence/performance-native-heap-caches-held-2026-11-03.json.gz)
