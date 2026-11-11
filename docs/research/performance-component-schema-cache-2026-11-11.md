# Component schema cache performance study

**Local experiment.** Pinned LLGo has an adverse throughput point estimate. Performance and release acceptance remain open.

## Mechanism

The candidate retains up to 32 successful structural schema pairs and 4096 nominal identity obligations per component instance. Every cache hit checks the current namespace resource identities. The validation budget participates in the key. Mismatches, errors and capacity limits use the existing comparator. The Store actor owns lookup/insertion, and shutdown clears entries after active calls and source cleanup finish.

The cache lookup and closure-based miss calculation have separate functions. A minimal cold-closure probe finds 100,000 extra allocation events for 100,000 hit-path calls under LLGo 3d92a58. Go 1.27.3 and LLGo e2cca2c record zero. The parameter-capture missed-path control records 100,000 events under Go/current LLGo and 200,000 under the pinned LLGo. The probe identifies a toolchain-version difference. Attribution to an individual merged PR remains pending.

## HTTP measurements

M4 Max/macOS arm64, plain 16-byte Rust P3 HTTP, 4 workers and 16 clients. Source base `4ce75302c8def7dfd85680a33d2e7a55593c8dfa` has the exact production implementation of `0aaab1adfdb5a14d88a108bea06017fe0698a675`. Timed candidate `dd77037afc23da873ec658a9c38c89bbab5f782f`. Every process uses 256 requests/client, 2 warmups and 7 measured batches.

| Toolchain profile | Base req/s | Candidate req/s | Base API μs | Candidate API μs | Paired warm-time ratio |
|---|---:|---:|---:|---:|---:|
| Go 1.27.3 | 26,626 | 29,767 | 148.973 | 133.462 | 0.8927 |
| Pinned LLGo 3d92a58 | 6,996 | 6,561 | 571.558 | 599.708 | 1.0408 |
| Experimental LLGo e2cca2c / sharded runtime f2e4cb426 | 12,753 | 13,376 | 308.261 | 294.799 | 0.9782 |

Three fixed primary/A/A blocks, seed `2026111181`. All 60 throughput processes and 1,720,320 measured warm responses pass the common byte/result checks. 10 qualification processes and 10 independent resource processes also complete. The paired ratio is the geometric mean of the three block ratios after equally averaging the primary/A/A log times within each block. Lower warm-time ratios indicate shorter elapsed time.

| Independent control | req/s | Mean API μs |
|---|---:|---:|
| Wasmtime 48.0.1 | 74,994 | 52.951 |
| Ordinary Go HTTP | 1,912,975 | 0.643 |
| Ordinary pinned LLGo HTTP | 124,212 | 1.782 |
| Ordinary experimental LLGo HTTP | 130,514 | 1.672 |

These observations cover complete component HTTP execution and ordinary host implementations. Public API timers, request timers and service-wall throughput have their respective recorded boundaries. Each metric uses its directly recorded interval. The competing profile uses the same guest bytes and request/result checks. Runtime policies and host implementations have their recorded differences.

## Controls and limitations

| Participant | Three A/A warm-time ratios |
|---|---|
| base-go | 1.0168, 0.9496, 1.0067 |
| base-llgo | 1.0002, 1.0449, 0.9515 |
| base-sharded | 0.9920, 0.9594, 0.9918 |
| go-host | 1.0165, 1.0726, 0.9866 |
| head-go | 0.9800, 1.0401, 0.9893 |
| head-llgo | 0.8069, 1.0341, 1.0274 |
| head-sharded | 1.0068, 0.9716, 1.0059 |
| llgo-host | 0.9648, 1.1111, 0.9885 |
| sharded-host | 1.0114, 0.9050, 1.0009 |
| wasmtime | 1.0488, 1.0088, 0.9771 |

Several controls exceed the predeclared 2.5% band. The pinned LLGo candidate has an A/A ratio 0.8069 in one block. Its source-pair block ratios are 1.0871, 1.0494 and 0.9884. The source improvement under pinned LLGo remains unestablished. The experimental runtime profile has source-pair ratios 0.9605, 1.0125 and 0.9627. Its small point improvement also requires stronger evidence. All adverse and reference observations are retained.

Preflight observed an unrelated CPU-bound test process and its parent. Owned compilation, tests and allocation diagnostics completed before the throughput event. Workstation timing and the three-block count limit the conclusions. The current compiler/runtime profile has separate identities and source pairs. The existing open LLGo #3045 interpreter regression remains relevant to production toolchain qualification.

## Allocation events

A separate single-worker observer warms 10 requests, brackets 20,000 checked requests, and records an empty-observer control. Every empty control reports zero events. Go uses its public Mallocs counter. LLGo uses the existing exact size-class MemProfile counters because public native Mallocs remains unavailable before #3043. These are managed allocation events, including short-lived objects. Timing from this observer is excluded.

| Toolchain profile | Base events/request | Candidate events/request |
|---|---:|---:|
| Go 1.27.3 | 1207.0075 | 1159.0106 |
| Pinned LLGo 3d92a58 | 3528.5081 | 3386.5928 |
| Experimental LLGo e2cca2c / sharded runtime | 2453.6174 | 2405.6350 |

The initial single-function cache candidate `f6830a4` records 3530.77 LLGo events/request and 1159.00 Go events/request. Separating the cache miss captures yields 3386.59 and 1159.01 respectively. The initial source, builds and count observations are retained independently.

The full toolchain-profile change also reduces the base request count from 3528.51 to 2453.62 events/request. This includes compiler and runtime revisions. The isolated cold-closure example verifies one concrete allocation improvement.

## Preparation phases

Median primary-process phases in milliseconds. First call is one full 4096-request batch.

| Participant | Compile/validate | Instantiate | First batch | Cleanup |
|---|---:|---:|---:|
| base-go | 161.888 | 0.467 | 156.075 | 0.058 |
| head-go | 162.054 | 0.457 | 142.145 | 0.063 |
| base-llgo | 239.156 | 0.714 | 576.528 | 0.594 |
| head-llgo | 242.589 | 0.740 | 608.521 | 0.538 |
| base-sharded | 227.897 | 0.770 | 312.118 | 0.563 |
| head-sharded | 231.340 | 0.769 | 310.947 | 0.561 |
| wasmtime | 51.470 | 0.215 | 59.620 | 0.236 |

Process readiness, fixture load, provider/engine setup, lookup, warm batches and native executable builds are recorded separately in the data. Every LLGo build uses `-O3 -lto=full`. Go uses 1.27.3 and native builds use LLVM 22.1.8.

## Collector marker discovery

A separate single-process sweep uses the identical base experimental executable, fixed arguments and supported GC_MARKERS settings. Seed `2026111183` selects the order. This discovery sweep has one throughput and one resource process per setting.

| GC_MARKERS | req/s | Idle OS threads |
|---|---:|---:|
| 2 | 12,260 | 43 |
| 8 | 13,421 | 49 |
| 1 | 12,705 | 42 |
| default | 13,153 | 57 |
| 4 | 12,919 | 45 |

The sweep observes approximately 12,260–13,421 req/s. This single sweep leaves collector-default qualification open. Installed BDWGC 8.2.12 headers and the library disassembly expose the existing thread-local small-object allocation path.

## Validation and retention

- macOS arm64 passes the full affected component/P3 suites under Go and pinned LLGo, covering both forced backends.
- Go component race tests pass. The measured component coverage is 91.66% and P3 coverage 91.30%.
- Local Docker Linux amd64 passes focused schema/source lifecycle, real Rust plain/composed HTTP, concurrent Store buffers and cancellation tests under both compilers/backends.
- Namespace rebinding, identical type pointers with distinct current resource identities, budget changes, negative matches, cache bounds and shutdown ordering have regression coverage.
- The final hosted full package/platform matrix remains pending while external LLGo PR #3057 is active. The local candidate remains on its contribution branch.

The compact JSON retains every throughput process, warm-call array, phase interval, resource summary, source/toolchain identity and raw-file checksum. Full per-request arrays, qualification records, scripts and observer outputs are retained in the local `artifacts/schema-matches-20261111` directory. The initial implementation and its diagnostics remain under the `initial-` artifact names. [Measurement data](data/performance-component-schema-cache-2026-11-11.json)
