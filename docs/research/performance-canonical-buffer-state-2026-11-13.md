# Canonical buffer state

Canonical codecs own their transfer state by value, allowing synchronous
resource-free buffers to use fresh local state. The Go source comparison reduces complete P3 request
CPU by 0.6%/1.5% for 16 B/64 KiB. Competitive parity remains open.

## Runtime change

The former codec pointed into its own allocation. Go escape diagnostics show
that this self-reference forced otherwise short-lived buffer state onto the
heap. The internal codec now contains its transfer state by value. Synchronous
copies construct fresh local codecs; suspended tasks and resource-bearing
transfers retain their separate ownership state. No codec or Caller is pooled.

Known scalar buffer elements also avoid building an unused canonical realloc
closure. Strings retain realloc; composite/resource elements retain the normal
codec. Caller lifetime, cross-memory options, aggregate byte/value accounting,
range validation, destination ordering, cancellation and cleanup stay shared.
The ten-file implementation changes add eight net lines, including adaptations
to internal constructors and tests. There is no public API or native frame ABI
change.

An exact baseline allocation diagnostic excludes setup and 80 warm requests,
then profiles 4,096 checked requests at sampling rate 1 for each payload. It
counts 56/92 codec constructions per request. These allocation stacks identify
sites; their profiled timings do not establish CPU cost. Full baseline profiles
and the before/after escape diagnostics are archived with the protocol.

## Complete request comparison

M4 Max/macOS 27.0.1 arm64; source base `0d966f6`, candidate `d87ea62`.
Go 1.27.3, Clang 22.1.8 and pinned
LLGo `56fa9ae` with `-O3 -lto=full`. The unchanged independent Rust bridge runs
actual Wasmtime 48.0.1 P3 providers. The same frozen Rust guest, snapshot input,
buffered output, four Stores, sixteen clients and 512 requests/batch remain.
Every response checks body, status, header, trailers and CRC32.

The independent final event uses 20 blocks, four roles and identical-binary
aliases: 320 fresh processes across separate 16 B and 64 KiB cohorts. First call,
two warmups, preparation and cleanup stay separate. Twenty measured batches
interleave individually; other child processes remain stopped. Whole-process
CPU/instruction windows include the final background tail. Batch throughput
and API intervals are distinct. GC is enabled, and no owned build, test,
profiler, source edit or pruning overlaps timing. Earlier preflights are not
pooled. Individual 95% intervals use 20,000 seeded, paired block resamples.

| Go head/base | 16 B | 64 KiB |
| --- | ---: | ---: |
| CPU ratio, individual 95% interval | 0.99393 [0.98854, 0.99964] | 0.98540 [0.98117, 0.98972] |
| Retired-instruction ratio | 0.99763 [0.99675, 0.99848] | 0.98910 [0.98838, 0.98987] |
| Throughput ratio | 1.00530 [0.99254, 1.01803] | 1.00920 [1.00214, 1.01572] |
| Managed allocations/request | 705 → 698 | 789 → 758 |
| Allocated bytes/request | 55,481 → 54,696 | 127,817 → 124,345 |

All Go CPU/instruction A/A intervals resolve ±2.5%. Both 64 KiB Go throughput
controls do too. Both 16 B Go throughput controls remain unresolved; there is
no small-request throughput claim. These are scoped diagnostics, not
simultaneous portfolio acceptance.

| CPU microseconds/request, main/alias | 16 B | 64 KiB |
| --- | ---: | ---: |
| Go base | 92.26 / 92.05 | 125.23 / 126.03 |
| Go head | 93.37 / 91.65 | 123.31 / 124.15 |
| LLGo head | 256.94 / 251.24 | 288.67 / 293.80 |
| Wasmtime | 42.87 / 42.13 | 55.92 / 56.85 |

The paired 64 KiB Go/Wasmtime CPU ratio remains 2.24194 [2.17699,2.31365],
with both CPU controls resolved. Go still executes about 3.75 times its
instructions. Wasmtime 16 B CPU and all LLGo CPU/instruction/throughput
controls remain unresolved; retain their points without a formal rank or an
LLGo source improvement claim. There is no matched LLGo source baseline.
Wasmtime's managed counters cover only the Go shim, not Rust/native heaps.

Two local prototypes remain outside production. Direct provider counts show
zero stdio calls and one pending HTTP completion future per request; canonical
import labels name shared endpoint types, not necessarily executed providers.
The separate stdio shortcut therefore missed this workload. Replacing the
pending future worker with subscriptions added four allocations without a
clear Go CPU benefit and was also rejected. Neither expands this change.

## Validation and evidence

All 32 actual-source Go test packages pass and all 31 executable packages retain
at least 90% integration coverage: component 92.47%, canonical and all other
packages satisfy the same gate. Relevant component/P3 race tests pass. Pinned
production LLGo executes canonical, component, P2 and P3 suites on macOS.
Integration requires the final Linux/macOS dual-compiler matrix, CLI, quality
and benchmark checks.

[Identities, controls and data](data/performance-canonical-buffer-state-2026-11-13.json)
retain every final process observation. Compressed projections preserve every
reported metric, counter window, phase and check. Only unused per-request
latency arrays become distributions and exact hashes; full originals remain
in local artifacts. Protocol scripts, allocation profiles and escape diagnostics
are included. This change does not close the remaining host, component,
interpreter, compiler or production performance gates.
