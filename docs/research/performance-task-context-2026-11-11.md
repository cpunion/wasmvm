# Compiled canonical task context

**Implemented local candidate.** Both execution backends share checked integer
storage accessors. Native direct calls compile to guarded loads/stores. Retired
instructions fall by 5.1–5.6% in the measured P3 workload. Throughput points
improve, with unresolved timing controls. Performance and release acceptance
remain open; the full hosted final-head matrix is pending.

## Implementation and semantics

Source base `93a60f11e5ffd7b886c17f2feaa15c293fe9e23f` already includes the
[empty source-cleanup check](performance-source-cleanup-2026-11-11.md). Source
head `d8fb0fdf373086788bf7af5cebb95132abaa9123` adds the
[shared scalar-storage contract](../design/scalar-storage.md). These source
comparisons isolate the context implementation from the earlier cleanup change.

Canonical context access binds the existing active task and failed-namespace
guard on every entry. The existing task scheduler retains initialization,
realloc isolation, stackful reentry, suspension and cancellation. Both slots
support I32/I64; I32 setters zero-extend. Native-owned cells are refreshed before
entry and published on exits, including traps and fuel exhaustion. Go addresses
never enter the native ABI. Ordinary imports retain their checked call path.
Accessors add no guest frame, and calls retain fuel and interrupt checks.

Temporary counter-only copies each execute 1,000 complete, checked requests.
Go and production LLGo produce identical counts. Counter timing is excluded.

| Event/request | Before | After |
|---|---:|---:|
| Native entries | 186 | 142 |
| Native Call exits | 99 | 62 |
| Native Return exits | 87 | 80 |
| Core Run boundaries | 93 | 56 |
| Generic host callbacks | 81 | 44 |
| Cross-instance guest calls | 18 | 18 |
| Component tasks | 9 | 9 |

All 37 context accesses leave generic host dispatch. The additional seven fewer
Return exits are an observed structural result; individual return sites were
not separately attributed. Poll and Move exits remain zero in this fixture.

## Complete P3 source comparison

M4 Max/macOS arm64, Go 1.27.3, production LLGo `3d92a58`, experimental LLGo
compiler `e2cca2c` with sharded runtime `f2e4cb426`, and Wasmtime 48.0.1. Every
LLGo binary uses `-O3 -lto=full`, unchanged within its source pair. The
experimental toolchain remains separate from the pinned production baseline
and its outstanding upgrade qualification.

The same plain Rust HTTP component processes 16-byte bodies with four workers,
16 clients and 256 requests/client. Common checks retain status, headers,
complete body, trailers and CRC32. Three fixed rounds, seed `2026111197`, run
30 measured batches after the first batch and two warmups. Each compiler has
base/head and a separate identical-base A/A process; Wasmtime is independent.
All 30 processes and **3,686,400 measured responses** pass.

Darwin `proc_pid_rusage(RUSAGE_INFO_V4)` records 29 intervals between readiness
markers, covering **3,563,520 checked requests**. CPU includes all process
threads and uses the verified Mach timebase of 125/3 ns/tick. Counter windows
include batch-statistics/readiness work; startup and cleanup are outside them.
Each process's throughput is 4096 divided by its median warm batch duration.
The table reports medians of three process observations per row.

| Configuration | req/s | CPU µs/request | Retired instructions/request |
|---|---:|---:|---:|
| Go, base | 26,485 | 149.78 | 1,891,813 |
| Go, head | 28,199 | 141.02 | 1,785,657 |
| Production LLGo, base | 7,071 | 650.04 | 2,590,783 |
| Production LLGo, head | 7,886 | 590.93 | 2,459,555 |
| Experimental LLGo, base | 13,160 | 367.73 | 2,503,503 |
| Experimental LLGo, head | 13,929 | 350.28 | 2,376,506 |
| Wasmtime | 75,100 | 24.70 | 207,959 |

The source instruction reductions are **5.61% / 5.07% / 5.07%** for Go,
production LLGo and experimental LLGo. All nine instruction A/A controls stay
within 2.5%. Throughput points improve **6.47% / 11.52% / 5.84%**, but only
**6/9 timing controls** resolve within that band. Three first-round controls
drift by −3.2%, +9.3% and −9.9%. All observations are retained. Frequency and
neighbor load are uncontrolled; these points establish no sustained ranking.
The 23.7% reduction in native entries is not a 23.7% throughput claim.

## Preparation and memory

These separately timed lifecycle phases are not added to warm request cost.
`firstCall` is the first **4096-request batch**, not one request's latency.
Instantiation sums the worker instances. All phase values below are medians
in milliseconds. The observed lifetime-peak footprint is sampled at batch
markers, includes preparation, and is not a final post-exit peak measurement.

| Head configuration | Validate/compile | Instantiate | First batch | Cleanup | Observed peak MiB |
|---|---:|---:|---:|---:|---:|
| Go | 161.78 | 0.52 | 150.40 | 0.06 | 114.0 |
| Production LLGo | 245.45 | 0.73 | 526.34 | 0.62 | 164.2 |
| Experimental LLGo | 231.38 | 0.85 | 296.00 | 0.62 | 160.5 |
| Wasmtime | 52.93 | 0.20 | 60.90 | 0.28 | 21.6 |

Wasmtime uses cached typed Service bindings and a ready `Full<Bytes>` body.
The owned adapter retains its public `http.Request`/`io.Reader` API and bounded
off-Store I/O workers. The [cost study](performance-p3-cost-attribution-2026-11-11.md)
records these policy differences and the separate ordinary HTTP controls.

## Validation and next work

- The final tracked source snapshot plus prepared pinned fixtures passes all
  32 Go internal/pkg test groups; the CLI package compiles. All 31 executable
  coverage gates pass, minimum **90.87%**, component **92.45%**, VM **93.67%**.
- macOS arm64 production LLGo executes the complete codegen, state, execute,
  nativeexec, VM, component and P3 test packages with the pinned release flags.
- Local Docker Linux amd64 executes focused accessor, async-task, stackful,
  cleanup and real Rust HTTP tests under Go/LLGo and both backends: eight
  successful package groups. Go component/VM race, whole-module vet and the
  `CGO_ENABLED=0` build pass.
- Tests cover exact widths, shared aliases, task rebinding, ordinary imports,
  direct/tail calls, invalid bindings, traps, fuel, cancellation, panic and idle
  workspace retirement. Generated-code tests verify accesses complete without
  a host exit. A suspected ordinary-import semantic failure was not reproduced
  before the explicit opcode guard; that hypothesis is not a confirmed bug.
- Early disk exhaustion interrupted builds/Docker, and benchmark-only archives
  omitted fixtures required by full semantic tests. Failed logs remain retained
  and unqualified. Final runs use the complete frozen snapshot and fixtures.
- Experimental LLGo still has a **5.39× throughput gap** to Wasmtime
  and retires **11.43×** its instructions in this fixture. Production LLGo's
  throughput gap is **9.52×**. Context lowering completes this increment; it
  does not close component-runtime performance acceptance.
- Next work addresses cross-instance calls, checked Canonical ABI conversion
  and task execution costs, followed by isolated LLGo scheduling analysis.
  Allocation/collector optimization alone cannot account for the remaining
  instruction gap. The original source study retains its independent GC-off
  evidence; percentages from those different cohorts are not combined here.
- The one-active-contribution rule keeps this implementation on a branch while
  [LLGo #3057](https://github.com/xgo-dev/llgo/pull/3057) is active. The full
  hosted final-head matrix and contribution PR remain pending.

[Process arrays, controls and provenance](data/performance-task-context-2026-11-11.json)
and [compressed counter evidence](evidence/performance-task-context-2026-11-11.jsonl.gz)
publish every process's counters, snapshots, phases, warm batches and checked
results. Per-request latency arrays remain in the complete local raw event,
whose checksum is recorded; they are omitted from the published counter file.
