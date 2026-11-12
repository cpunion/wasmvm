# Explicit caller-owned complete HTTP responses

`166e38c` adds `HTTP.CallIncomingInto`: the caller supplies exclusive bounded
output storage and receives an aliasing view after complete guest join.
`CallIncomingBuffered` continues returning independent owned bytes. Both use
the same collector, fuel, copy budgets, content-length checks, cancellation,
request quotas and cleanup. Reuse changes earlier views; a failed call may
change a written prefix. The VM retains no destination after return.

## Exact-head observations

M4 Max/macOS arm64, Go1.27.3, Clang22.1.8; independent pinned Wasmtime48.0.1.
Base57ffbf5 is runtime-equivalent to merged0dfbb1b. Four Stores,16clients,
64requests/client, first batch, two warmups and twelve measured batches.
Three seeded randomized rounds include independent identical-binary A/A aliases.
All48processes and589,824measured responses pass full status/header/body/trailer/
CRC checks. CPU windows cover540,672requests. No owned builds, tests, profiling
or cache pruning overlap timing; unrelated machine work was observed.

**Timing controls are unresolved. These numbers establish no throughput win or
competitive acceptance.** Allocation bytes are stable across aliases:64KiB
caller buffers remove one response allocation and reduce127,817 to62,280bytes/
request. Small responses remove only16bytes. The output ownership contract
changes explicitly and cannot replace the ordinary fresh-response ranking.

| Payload | API/runtime | req/s | A/A req/s | CPU us/request | Bytes/request |
|---|---|---:|---:|---:|---:|
| 16 | Base fresh | 36,298 | 36,772 | 101.69 | 55,480 |
| 16 | Head fresh | 36,980 | 35,309 | 99.35 | 55,480 |
| 16 | Caller buffer | 36,292 | 36,120 | 100.26 | 55,464 |
| 16 | Wasmtime48 | 53,489 | 50,250 | 40.99 | 32 |
| 65536 | Base fresh | 23,330 | 27,058 | 144.44 | 127,816 |
| 65536 | Head fresh | 24,192 | 27,073 | 149.00 | 127,817 |
| 65536 | Caller buffer | 30,099 | 26,996 | 122.24 | 62,281 |
| 65536 | Wasmtime48 | 45,692 | 44,486 | 51.70 | 32 |

Only 3/24 throughput and 3/24 CPU A/A pairs resolve within2.5%;
21/24 instruction pairs do. Largest timing deviations remain in the data.
Both owned APIs still trail the reference in this cohort. The previous quiet
prototype's approximate9%large-response gain is not promoted to an exact-head
claim. Go performance acceptance remains open.

Preparation, engine compilation, instantiation, first batch, input/output buffer
preparation and cleanup stay separate from warm execution. One output buffer
per serialized Store worker is prepared outside the warm interval and recorded
in requestPreparation. Full ordinary input Reader/snapshot and streaming APIs
remain available; this event uses explicitly immutable input snapshots.

## Validation and evidence

Whole Go component/P3 and measurement suites pass; scoped statement coverage
is92.44%/91.28%/96.3%. macOSarm64 pinned LLGo3d92a58 `-O3 -lto=full` executes
whole P3 and measurement tests. Linuxamd64 Go P3 passes. The complete captured
LinuxLLGo P3 binary passes on a bounded retry, as do new tests and the upstream
consumer group independently. The first emulated full run stalled for19minutes
and was terminated; its cause is unconfirmed and it is retained locally.
Final-head native CI must qualify the complete dual-compiler/platform matrix.

Alias/canary/overflow/reuse/empty-body/nil-instance/precancellation, metered guest
join and blocked external-reader cancellation cases pass both backends. Focused
race and whole-module vet pass. No checks or compiler profiles are relaxed.

[Data and controls](data/performance-caller-output-2026-11-12.json) retain source/binary identities;
[all process and batch records](evidence/performance-caller-output-2026-11-12.jsonl.gz) retain phases and
counter windows. Per-request latency arrays are summarized with hashes; full
originals remain in local artifacts. No latency acceptance is claimed.
