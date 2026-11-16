# Native BDWGC pacing diagnostic, 2026-11-16

This records the investigation for [LLGo #3106](https://github.com/xgo-dev/llgo/issues/3106)
and [PR #3120](https://github.com/xgo-dev/llgo/pull/3120). The implementation makes
native `GOGC`/`debug.SetGCPercent` control BDWGC's scan-work and occupied-heap
allocation budgets. It deliberately permits a larger heap to reduce collection
work. These application measurements are focused diagnostics, not full-portfolio
acceptance or a universal speedup claim.

## Exact scope and method

The application source is frozen at wasmvm
`393bf75c433668084611d6f6e69fb9f84b20def5`. It runs the existing complete P3 HTTP
Rust guest `pkg/wasi/preview3/testdata/world-rust-http.wasm`, native owned backend,
4 workers, 16 clients, buffered responses and snapshot requests. Every guest
result is checked: CRC 1757737011 for 16 bytes and 2629362236 for 65536 bytes.
Each response also goes through the existing status/header/body/trailer checks.
Production VM/provider code is unchanged.

Host: Apple M4 Max, macOS arm64, Go 1.27.3, LLVM 22.1.8, BDWGC 8.2.12.
All builds use `-O3 -lto=full`; all processes use normal GC with `GOGC=100` and
`GOMAXPROCS=4`. No artificial allocation warmup, GC disablement or source padding
is introduced. The final cohort compares main
`5860fb6f9c13371f6cdb456a987a8e95844066d4` with PR head
`e5dc41d1da1cfa27343f2688612dbe3980413730`. Both application binaries use the same
compiler executable; only the SDK/runtime root differs. Compiler build identity,
SDK revisions and executable hashes are recorded with the data.

For each cohort/payload there are exactly 20 fixed blocks and 80 fresh processes:
base, base-aa, head and head-aa. The two labels in each A/A pair execute the same
binary. Each process supplies 20 interleaved batches of 512 requests. Between
batches the coordinator stops/resumes children and samples macOS process CPU and
instruction counters. CPU includes the measured service/coordinator handshake
window; it excludes off-CPU waiting. Startup, first request, warmups and final
flush stay separate and remain in the records. No owned compilation, tests or
source edits ran during the final timed cohort. Host-load snapshots are included.

CPU/instruction/throughput ratios are paired by block and summarized as the
geometric mean of ratios; intervals are a 10,000-resample block bootstrap with
seed 20261116. Throughput uses each process's median batch wall time. It is not
the inverse of total process CPU time. Cohorts with different bases or heads are
never pooled. Lower CPU ratios and higher throughput ratios are better.

## All candidate cohorts

| Cohort | Payload bytes | CPU head/base, 95% interval | Throughput head/base, 95% interval |
|---|---:|---:|---:|
| heap-floor-rejected | 16 | 0.9989 [0.9713, 1.0280] | 1.0091 [0.9828, 1.0360] |
| heap-floor-rejected | 65536 | 1.0477 [0.9936, 1.1014] | 0.9648 [0.9369, 0.9953] |
| scan-prototype | 16 | 0.8750 [0.8577, 0.8913] | 1.0714 [1.0566, 1.0869] |
| scan-prototype | 65536 | 0.8984 [0.8579, 0.9395] | 1.0619 [1.0335, 1.0918] |
| pre-review-360f6026 | 16 | 0.8576 [0.8361, 0.8805] | 1.0688 [1.0531, 1.0838] |
| pre-review-360f6026 | 65536 | 0.9085 [0.8619, 0.9546] | 1.0628 [1.0236, 1.1063] |
| reviewed-e5dc41d1 | 16 | 0.9054 [0.8918, 0.9191] | 1.0582 [1.0169, 1.0916] |
| reviewed-e5dc41d1 | 65536 | 0.8915 [0.8516, 0.9307] | 1.0663 [1.0344, 1.0984] |

The heap-floor-only prototype is rejected: it did not reliably reduce CPU work
and increased collections in the larger-payload case. It used main 113daaad5;
its plan and binary hashes survive, but an immutable source revision for that
uncommitted prototype was not captured. It cannot serve as source-reproducible
acceptance evidence. The scan-work prototype also used main 113daaad5. The
pre-review cohort used main 5860fb6f9 and PR head 360f6026; it is retained
separately from the final reviewed head. The exact final source diff is supplied
as `llgo-e5dc41d1.patch`.

The reviewed head preserves the last active budget while automatic GC is off,
avoids forced collection on an unchanged restore or a small below-budget
reduction, preserves the host's preexisting automatic-disable flag, and yields
when contending for the policy lock. Native policy tests require BDWGC >= 8.2;
no older-library compatibility layer is provided.

## Final-head collection and footprint cost

Medians below describe individual fresh processes. GC counts cover the measured
20-batch window. Managed heap is the largest post-batch occupied-block estimate,
not exact live Go object bytes. Physical footprint is the largest macOS
`ri_phys_footprint` sampled in that window, not process lifetime peak RSS. It
includes foreign/JIT/linear memory as well as the managed heap.

| Payload bytes | Label | CPU us/request | Collections/window | Managed heap MiB | Physical footprint MiB |
|---|---|---:|---:|---:|---:|
| 16 | base | 253.3 | 28 | 43.7 | 172.7 |
| 16 | base-aa | 264.7 | 36 | 38.2 | 111.9 |
| 16 | head | 229.5 | 12 | 78.3 | 215.3 |
| 16 | head-aa | 232.2 | 12 | 78.2 | 215.3 |
| 65536 | base | 333.9 | 39.5 | 57.7 | 136.9 |
| 65536 | base-aa | 331.8 | 59 | 44.9 | 107.3 |
| 65536 | head | 274.2 | 24 | 78.0 | 215.2 |
| 65536 | head-aa | 278.5 | 24 | 80.0 | 215.8 |

The fixed default changes BDWGC's free-space divisor from 3 to 1. That increases
both the permitted scan-work budget and the heap expansion increment from about
heap/3 to about heap, subject to collector limits. The occupied-heap floor adds
another minimum. Consequently a lower collection count has a material footprint
cost; this is not a memory-neutral optimization. Percentages above 100 leave
the divisor at 1 while still scaling the occupied-heap floor.

Application allocation counts/bytes remain approximately 1014 events/73120 bytes
per 16-byte request and 1194 events/148241 bytes per 64-KiB request in both builds.
The policy changes collection scheduling, not application allocation semantics.
Small per-process differences include harness/runtime bookkeeping.

## Controls and limits

The final cohort's baseline A/A CPU intervals exceed the prespecified
[0.975, 1.025] calibration band on both payloads. The head A/A CPU intervals are
inside that band. Baseline collection counts and footprints also vary substantially
between identical-binary processes; all observations are retained above and in
the data. Several throughput A/A intervals also exceed the band. Thus the final
paired CPU ratios are observations with unresolved baseline controls, not a
calibrated throughput acceptance result. Prior cohorts have their own unresolved
controls, preserved in each `analysis.json`; repeating cohorts does not erase them.

The final head consistently uses fewer collections and a larger managed/physical
heap in this workload. Pacing reduces sensitivity here but does not establish
that BDWGC's conservative-root/address sensitivity is eliminated on all programs.
No Linux/Windows performance claim follows from this macOS cohort. API/GC tests
and application performance measurements are distinct evidence.

## Data and reproduction

[Data directory](data/llgo-gc-pacing-20261116/) includes plans, executable/compiler
hashes, all process-level counters, service batches, CPU/GC/heap observations,
analysis, host-load snapshots and each executed measurement script. The scripts
retain their original absolute paths as a record; replace those paths to rerun
with separately built frozen-source binaries. Build the P3 module from
`benchmarks/p3io` using the same compiler executable and each SDK root, with
`LLGO_ROOT=<sdk> llgo build -O3 -lto=full -o <binary> .`.

Only the large per-request `RequestNs`/`PublicAPINs` vectors are projected to
count/min/max/canonical-JSON digest in `rows.projected.jsonl.gz`; no latency-tail
claim is made from those projections. All other fields, including adverse
processes, first/warmup/final-flush records and A/A controls, are preserved. The
manifest also records the SHA-256 of each original full raw gzip, which remains
in the local audit archive. `analyze.py` computes the published summaries from
the original rows; to use a projection, change its filename to
`rows.projected.jsonl.gz` (it does not read the omitted vectors).
