# Native compiler layout and bounded emission

Go source `d7e513d` reduces native compilation of the real P3 HTTP component
from **179.0 to 73.9 ms**, with CPU **210.3 to 136.5 ms**. Wasmtime48.0.1
compiles it in **57.7 ms** in this cohort. **Competitive parity remains open.**
Compilation and warm HTTP throughput are separate measurements.

## Changes

Optional leaf expansion no longer recompiles whole functions that have no
eligible direct or indirect callee, only to discard their unchanged result.
Both required policy tiers retain priority over optional expansion.

Native functions compile through at most four persistent workers, respecting
GOMAXPROCS. A bounded ring holds at most four outstanding bodies. Publication,
aggregate quotas and errors retain function order. Every exit cancels and joins
workers; unexpected panics return to the publishing goroutine. No executable
allocation survives a failed module compilation.

arm64 branch compaction replaces temporary hash maps, sorting and per-offset
binary searches with deletion bitmaps and block prefixes. Metadata uses16bytes
per256reserved code bytes, plus one block. Relocations, tables, ADRs and cold
continuations keep their original positions after compaction. Dense label IDs
use indexed positions with an explicit unbound sentinel; offset zero remains a
valid target. An independent previous layout algorithm checks emitted bytes
and all labels across64mixed forward/backward and reordered relocation cases.
The existing range-boundary, invalid-label and exact-budget tests remain.

Ordered unsigned memory loops now use the existing checked range proof for
`<` backedges and their commuted/inverted forms. The positive-stride progress
guard remains mandatory; misaligned, nondivisible, initially out-of-range and
unsafe ranges retain the original checked loop. ABI guest resource IDs use a
dense table; nonreusing host IDs keep their map. Stable entry identity, moves,
borrow scopes, generations, quotas and reentrant destructor behavior remain.

## Exact-source compilation event

M4Max, macOS27.0.1 arm64, Go1.27.3, Clang22.1.8, GOMAXPROCS4. Baseline
`f500707` has the same compiler as merged PR158; head is `d7e513d`. Twenty
randomized fresh-process blocks include independent A/A aliases for baseline,
head and Wasmtime. All120processes execute two checked HTTP responses after
compilation. No owned local build, test or profiler overlaps timing.

| Configuration | Compile wall ms | A/A wall ms | CPU ms | A/A CPU ms | Instructions billions |
|---|---:|---:|---:|---:|---:|
| Go baseline |179.01|178.14|210.31|209.67|2.850|
| Go head |73.95|73.20|136.48|136.02|1.728|
| Wasmtime48 Cranelift |57.72|57.21|unavailable|unavailable|unavailable|

The public Compile wall timer includes decoding, validation, lowering, native
emission and publication. Paused protocol markers bracket its CPU/instruction
window with bounded publication overhead. Loading, process startup, engine
setup, instantiation, first execution and cleanup stay outside this window.
Wasmtime retains its Rust internal compile timer; isolated reference CPU was
not measured and is not inferred by subtraction.

Paired descriptive geometric head/base ratios are0.412wall,0.648CPU and
0.607instructions. The head/reference compile-wall ratio is1.268. A/A CPU
95%bootstrap intervals fit within2.5%; head wall does not (0.965–1.009), nor
does reference wall (0.947–1.014). Individual blocks have larger excursions.
These unadjusted diagnostic intervals do not establish full-portfolio or
all-platform performance acceptance. Raw controls and adverse blocks remain.

A separate48-process memory-loop screen observes Go0.434/0.430ns per iteration
versus wazero1.12 approximately0.817/0.816 and Wasmtime48 approximately0.484/
0.489 for small/sustained cases. Some A/A controls differ by6.4%; this is a
local diagnostic, not a closed execution gate. The corrected backend selection
and the preceding invalid setup attempt remain distinct local evidence.

The72-process full HTTP cohort has unresolved timing controls; it establishes
no throughput gain from indexed handles. The preceding quiet public HTTP
comparison (51,170/38,338versus75,733/66,238req/s for16B/64KiB) remains the
[recorded execution deficit](performance-native-runtime-image-2026-11-12.md).
Eight-worker expansion and alternative dense rank arrays are not retained.

## Validation

Whole affected Go semantics pass on macOS; exact scoped coverage for codegen,
canonical, nativeexec, VM, component, P2 and P3 is97.82/94.27/92.96/93.29/
92.44/90.87/91.28%. Final codegen/nativeexec/VM race checks pass. The preceding
compiler checkpoint also passes whole component/P3 race tests (P3:196seconds).
Pinned LLGo3d92a58 `-O3 -lto=full` executes complete macOS codegen/nativeexec/VM/
component/P3 suites; canonical/P2 pass before the final label-only change.

Linux Docker Go passes the seven affected packages before the final label-only
change. Linux LLGo codegen/canonical pass at that checkpoint. Remaining emulated
LTO links were stopped to avoid duplicating the required native GitHub matrix;
this is incomplete local validation, not a semantic pass. Final-head CI must
qualify the complete dual-compiler/platform matrix, coverage and race scope.
No safety settings, feature semantics or coverage thresholds were relaxed.

[Source identities, counters and controls](data/performance-compiler-layout-2026-11-12.json)
and [process/batch records](evidence/performance-compiler-layout-2026-11-12.jsonl.gz)
retain the final and diagnostic cohorts. Large per-request latency arrays use
count/hash summaries; original arrays remain local. Latency acceptance is open.
