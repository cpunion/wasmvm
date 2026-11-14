# Checked UTF-8 transfers

UTF-8 Canonical ABI transfers now use the existing immutable-memory capability
without an intermediate byte buffer. Lowering also skips rune counting needed
only for UTF-16/Latin1; Latin1 reuses the count already computed. Slice-only
memory implementations retain their existing path. No new public API or pool
is introduced.

Logical byte charges, preallocation range checks, Unicode validation and
encoding widths remain unchanged. Immutable snapshots must have the exact
requested size and cannot alias guest storage. The optimized capability path
checks cancellation again before publishing a successful read/write. Tests
cover both pointer widths, all three encodings, later guest writes, malformed
Unicode, incomplete snapshots, cancellation, memory failures, quotas and invalid
reallocation ranges.

## Complete P3 source comparison

M4 Max/macOS arm64, Go1.27.3, LLGoac129220 (-O3/fullLTO), LLVM22.1.8 and normal
GC. Owned base86bba024 and production candidateaa217c59 use the same adapter,
Rust guest, four Stores, sixteen clients, immutable snapshot input and bounded
buffered output. Frozen Wasmtime48.0.1 retains its original independent adapter;
source differences remain listed in the data. No reference is silently rebuilt.

Separate16B/64KiB events each use20 fixed blocks,200 fresh processes and20
individually interleaved512-request warm batches. Every role has an identical
binary alias. Preparation, first call, two warmups, final flush and cleanup stay
separate. All400 processes and4096000 measured responses pass full body,
status, header, trailer and CRC checks. Instrumented counters include driver
barriers and background tail. No owned local build/test/profile/source edit or
pruning overlaps either cohort. Completed16B stdout included its summary before
the64KiB cohort ended; no plan or sample-count changes were made. Final decisions
use both complete cohorts. Intervals are individual diagnostics, not adjusted
portfolio acceptance.

| Go head/base |16B|64KiB|
|---|---:|---:|
| CPU ratio,95% interval |0.98541 [0.97689,0.99469]|1.00559 [0.99732,1.01406]|
| Instruction ratio |0.99819 [0.99680,0.99946]|0.99954 [0.99732,1.00185]|
| Managed allocations/request |698 to696|758 to756|

Both Go CPU/instruction controls resolve2.5% in each event. The small-request
CPU decrease is supported; the large-request adverse point remains unresolved,
with the interval crossing1 and no established instruction increase. Several
throughput and both reference CPU controls remain unresolved; no competitive
ranking or closed performance gate follows from these events.

LLGo CPU points improve7.6%/8.0%, with one own CPU/instruction control failing
in each payload cohort. Its compile-wall points worsen about10%; this series
is under a separate fresh-baseline/build-state audit and does not establish an
accepted LLGo source improvement. Final attribution, controls and CI results
are retained in the contribution PR body/comments. All adverse points and
phases stay in the raw records; the original cohorts are not extended.

## Validation and CI

All32 Go test packages pass; every31 executable-package coverage gate exceeds
90%, including canonical94.42%. Full canonical/component tests also execute
with pinned LLGo on macOS. The final dual-compiler/platform matrix, race/quality
checks and reviews remain merge requirements.

The CI rebalance moves legacy wasi-threads semantics beside P1 and moves the
shared service driver/native reference beside P3 measurement validation. It
retains32 packages exactly once across the same8 LLGo shards and both extra
modules. The preceding Linux threads job took667 seconds, including207-second
semantic and368-second P3 steps. No duration improvement is claimed until the
new final-head jobs complete; no tests, profile flags or timeouts are weakened.

[Data and controls](data/performance-canonical-utf8-2026-11-14.json) link to
[every projected process record](evidence/performance-canonical-utf8-2026-11-14.jsonl.gz).
Unused per-request latency arrays retain counts, extrema and exact hashes;
full originals and protocol scripts remain local. Go competitiveness and
release acceptance remain open.
