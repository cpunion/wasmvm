# Source cleanup allocation and CPU study

**Local candidate.** Allocation reduction is verified. Stable source throughput improvement and release acceptance remain open.

## Implementation

Source base `4ce75302c8def7dfd85680a33d2e7a55593c8dfa`, timed candidate `93a60f11e5ffd7b886c17f2feaa15c293fe9e23f`. The base has the exact production implementation of `0aaab1adfdb5a14d88a108bea06017fe0698a675`.

`Instance.flushSourceCleanup` checks the queue under the existing mutex. It calls `drainSourceCleanup` when work is queued. Store-serialized execution owns the reentrancy flag. A source queued after the empty check is handled by the next Store turn or quiescent shutdown. The drain owns panic guards, affine releases and quota accounting. The source change adds one helper and 10 net lines.

## Allocation diagnosis

A temporary SDK instruments the first return PC of the native allocation recorder. A fixed 16,384-entry table counts managed allocation events and requested bytes without allocating in the hook. Profiling stops before symbolization. An empty control records zero events. A positive control records exactly 10,000 objects and 160,000 requested bytes at one call site, with zero overflow.

The unchanged base workload executes 1,000 checked 16-byte plain Rust P3 HTTP requests. The site table and existing size-class counters both record **2,452,778** events, with 294 sites and zero overflow. First-PC attribution retains runtime allocator-wrapper names where LLVM keeps those wrappers. These counts describe allocations. CPU cost is measured separately below.

| Attributed function | Events/request | Share of all events |
|---|---:|---:|
| `Instance.flushSourceCleanup` | 202.760 | 8.27% |
| `operation.host` | 192.000 | 7.83% |
| `hostCall.finish` | 117.000 | 4.77% |
| `componentTask.resume` | 100.000 | 4.08% |

Disassembly shows captured ownership/result cells allocated before the drain's queue test. Separating the empty check removes approximately 197.7 events/request.

A separate observer warms 10 requests and counts 20,000 checked requests per source/toolchain participant. Go uses `MemStats.Mallocs`. LLGo uses exact native size-class `MemProfile` counters pending upstream #3043. All six empty controls report zero. Observer timing is excluded from throughput results.

| Toolchain | Base events/request | Candidate events/request | Reduction |
|---|---:|---:|---:|
| Go 1.27.3 | 1207.0072 | 1207.0067 | Approximately zero |
| Pinned LLGo 3d92a58 | 3528.6586 | 3331.0141 | 5.60% |
| Experimental LLGo e2cca2c / sharded runtime f2e4cb426 | 2453.7599 | 2256.0271 | 8.06% |

Independent boundary diagnostics find zero new execution workspaces and 12 reused workspaces per request. VM scalar encoding/decoding accounts for approximately 136–138 events/request. Codec constructors account for approximately 60 events/request. Those observations left workspace pooling and constructor-only changes outside this increment.

The standalone current-upstream mutable-defer capture reproducer records one additional managed object/call under LLGo and zero under Go. Immutable captures, direct deferred methods and an escaping positive control identify its boundary. The conservative recover-visible placement enhancement is tracked in [LLGo #3058](https://github.com/xgo-dev/llgo/issues/3058).

## Uninstrumented HTTP comparison

M4 Max/macOS arm64, the same 16-byte plain Rust P3 HTTP guest, four workers, 16 clients and 256 requests/client. Every process has two warmups and seven measured batches. Three fixed primary/A/A blocks use seed `2026111186`. All 60 throughput processes and **1,720,320 measured warm responses** pass the common checks. Ten qualification and ten independent resource processes complete. Nine actual loader probes verify dependencies and retained non-system library hashes.

| Toolchain | Base req/s | Candidate req/s | Base mean API μs | Candidate mean API μs | Paired warm-time ratio |
|---|---:|---:|---:|---:|---:|
| Go 1.27.3 | 25,755 | 25,304 | 154.116 | 158.612 | 0.9867 |
| Pinned LLGo | 6,414 | 6,307 | 632.391 | 634.108 | 0.9914 |
| Experimental LLGo / sharded runtime | 12,501 | 12,046 | 330.234 | 324.215 | 0.9974 |

The paired ratio equally averages primary/A/A log times within each block and takes the geometric mean of the three source-pair ratios. Throughput uses median primary service-wall intervals. API values use their directly recorded boundaries. Each metric retains its own aggregation.

| Independent control | req/s | Mean API μs |
|---|---:|---:|
| Wasmtime 48.0.1 | 73,818 | 53.764 |
| Ordinary Go HTTP | 1,597,660 | 0.722 |
| Ordinary pinned LLGo HTTP | 122,737 | 2.049 |
| Ordinary experimental LLGo HTTP | 128,257 | 2.134 |

The base/candidate point estimates and controls are retained in full. Nine of 30 A/A controls fall within the predeclared 2.5% band. The ordinary Go control includes a 0.2786 ratio. This event leaves stable source improvement, controlled competitor ranking and release acceptance open. Preparation, first full batch, warm execution and cleanup have separate recorded phases.

The initial preflight assertion expected a newline before an `ok` log line and rejected the valid single-line race log. The outer shell continued and launched the original fixed measurement. Corrected validation confirms the completed semantic logs. The process snapshot was captured after measurement started. It cannot establish pre-start host quiescence. Unrelated tests and system work may affect the timing. The original event is retained without restarting.

## CPU profile

A separate Instruments 27 Time Profiler run launches the identical experimental candidate with 30 measured batches. The target exits successfully and verifies 122,880 measured warm responses. Waiting-thread recording is disabled. The analysis selects the interval from the first to last observed call of the actual buffered owned-kernel closure, including first/warmup execution. Its 46,852 running samples carry 46.852 seconds of accumulated process CPU weight. Traced timing is excluded from the table above.

| Exclusive sampled CPU category | Share |
|---|---:|
| BDWGC marker threads | 21.06% |
| Mutator allocator call paths | 18.66% |
| Other process CPU | 60.28% |

Marker-thread samples take priority in the exclusive classification. Allocator paths include collector helpers and synchronization reached from allocation. The two GC/allocation categories account for **39.72%** of sampled process CPU in this trace. `GC_mark_from` alone has 15.38% leaf weight. Native entry, type comparison, TLS access and profile counters remain individually visible in the data. The source allocation reduction has a small throughput point effect in this event, while collector/allocator CPU remains substantial.

## Collector budget discovery

The installed BDWGC 8.2.12 documentation describes its free-space divisor's collection/space tradeoff. A native library probe verifies the installed default is 3 and that process environment settings 2/1 take effect. The identical candidate binary executes two fixed primary/A/A blocks, seed `2026111188`, plus separate resource processes. Production uses the existing collector configuration.

| Free-space divisor | req/s | Collections per measured 4096-request resource batch | Process peak RSS MiB |
|---|---:|---:|---:|
| Default 3 | 12,550 | 15 | 177.0 |
| 2 | 12,480 | 10 | 211.3 |
| 1 | 13,549 | 5 | 241.7 |

All six A/A ratios exceed the 2.5% band. The divisor-1 throughput point estimate increases about 8% and its independent process peak increases about 37%. This small discovery event leaves controlled tuning qualification open. The collector setting remains an environment-only diagnostic.

## Validation and remaining work

- macOS arm64 Go executes all 32 root-module packages with all six pinned corpora prepared. All 31 owned executable-package coverage gates exceed 90%. Minimum coverage is 90.93%, and component coverage is 92.41%.
- Go component race tests, whole-module vet and the `CGO_ENABLED=0` whole-module build pass.
- Pinned LLGo `-O3 -lto=full` executes the full component, VM and P3 packages on macOS arm64.
- Local Linux amd64 Docker executes focused source/cleanup lifecycle, real Rust plain/composed HTTP, concurrent buffers and cancellation under both compilers/backends. All four package groups pass.
- The first Go test attempt used incomplete historical artifact layouts. Its failures are retained. Final validation uses an isolated artifact directory with all six verified pinned inputs. Initial loader-validator format and executable-image metadata assertions are also retained with the corrected successful nine-probe run.
- The full hosted platform/package matrix is pending if this candidate is proposed. The one-active-contribution rule keeps it on a branch while external LLGo #3057 is active.
- Remaining performance work includes recover-visible capture placement, collector/allocator costs, canonical/task overhead, the independent LLGo #3045 upgrade regression and controlled cross-target acceptance.

[Compact measurements and provenance](data/performance-source-cleanup-2026-11-11.json)
