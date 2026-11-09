# Shared concurrent service references: 2026-11-09

The owned engine, independent wazero and ordinary native C now use one bounded
admission/measurement driver. The fixed M4 Max event completes **540 throughput
processes, 54 separate resource processes and 3,870,720 checked warm requests**.
All **54 identical-binary A/A controls remain unresolved**: results are
observations, not a controlled ranking, optimization acceptance or release gate.

Source `ee28d454652542d159f6c3c9064a875968361424` has exactly the same production
engine files as PR #145 head `e02fe34`. The common service driver moves from the
phase adapter to `internal/bench/service`, without importing the owned VM.
Independent dependency inspection confirms the reference module imports only
this driver from the root module. No owned decoding/compilation or engine
fallback supplies the independent result.

## Scope and protocol

Both committed C/Rust SHA-256 Wasm producers use seed 17, 64 generated message
blocks/request, 64 clients, 16 requests/client/batch and 1/2/4 workers. Five
fresh process blocks retain primary and source-identical A/A aliases, two
warmups and seven measured batches. Every client acknowledges readiness before
one common release epoch; all replies and completed cleanup are checked.
The queue capacity equals the worker count. GOMAXPROCS=4 is recorded, not
equated to a native CPU quota. Frequency, affinity and workstation background
load remain uncontrolled.

The owned callbacks retain typed `Function.Call`; wazero uses private reused
`CallWithStack` buffers. Wazero 1.12 interpreter/compiler default and enabled
context-termination policies are distinct configurations. Native C compiles
the unchanged scalar SHA-256 source under each host's release build, with
Clang/LLVM 22.1.8 and `-O3 -std=gnu11`, and uses the ordinary C bridge. This is
one C algorithm control under Go/LLGo scheduling; no Rust-native producer,
identical native instruction stream or Wasm safety/fuel/lifecycle contract is
implied. Go 1.27.3 and LLGo `3d92a58 -O3 -lto=full` remain pinned.

Batch wall time includes queue admission, calls and replies. Request latency
starts at attempted admission, including scheduler wakeup after broadcast for
each client's first request. API intervals retain their actual boundary and
guest work. First/warmup, load, engine, compilation/validation, instantiation,
lookup, worker setup, readiness, cleanup and process intervals stay separate.
No interval is manufactured by subtracting independently reduced measurements.

## Observed native results

Requests/second; median of five per-process warm means, retaining every pause.

| Wasm producer / workers | Own Go | Own LLGo | Wazero compiler default | Native C / Go | Native C / LLGo |
| --- | ---: | ---: | ---: | ---: | ---: |
| C / 1 | 60,779 | 56,149 | 43,507 | 65,854 | 68,877 |
| C / 2 | 98,540 | 99,731 | 64,668 | 138,240 | 121,473 |
| C / 4 | 208,248 | 131,208 | 195,836 | 295,742 | 132,451 |
| Rust / 1 | 57,653 | 43,457 | 35,604 | 65,854 | 68,877 |
| Rust / 2 | 111,116 | 97,497 | 87,709 | 138,240 | 121,473 |
| Rust / 4 | 228,160 | 133,946 | 178,809 | 295,742 | 132,451 |

The full two-producer, both-backend, termination-policy, latency, lifecycle,
resource and A/A tables are in the
[PR result](https://github.com/cpunion/wasmvm/pull/145#issuecomment-6742127206).
The [machine record](data/performance-service-references-2026-11-09.json)
retains all 54 profiles, all 540 per-process reductions, control intervals and
source/compiler/binary/library provenance. The retained raw event preserves
every individual request/API interval and checksum, including adverse outcomes.

At four workers, the C-producer owned LLGo result is close to its native-C
service control, while directly measured mean API time is **17.407 vs
12.656 microseconds**. End-to-end throughput parity does not demonstrate equal
execution efficiency. The native-C control itself shows a Go/LLGo throughput
gap, despite API times of **12.769/12.656 microseconds**. Common admission,
communication, host scheduling and native code generation remain separate
investigation targets; these noisy rates do not isolate a cause.

Native-C request p99 is **2,780/986 microseconds** under Go/LLGo at four workers.
That observed latency/throughput tradeoff is retained alongside every A/A
failure; no universal host-runtime or scheduler win is inferred. Request
percentiles are computed within each process from all 7,168 warm requests
before taking the median across processes.

## Resource and LLGo findings

Resource instrumentation uses separate first-plus-one-measured-batch processes,
without warmup. Actual idle OS-thread snapshots are 6–7 under Go versus 84–87
under LLGo. RSS snapshots and getrusage peak RSS include runtime/JIT/native
memory and preparation; peaks extend only through the final pre-cleanup
snapshot. These are not peak owned-memory or completed-process measurements.

The native-C service records roughly **438–453 LLGo collector bytes/request**
versus near-zero Go managed allocation. Standalone allocation-only controls
then execute three rounds of zero/20,000 exchanges after 2,000 warmups:

| Channel case | Go B/exchange | LLGo collector B/exchange, three rounds |
| --- | ---: | --- |
| Same-goroutine buffered send/receive | 0 | 0 / 0 / 0 |
| Buffered two-goroutine ping-pong | 0 | 416.198 / 415.979 / 415.938 |
| Unbuffered two-goroutine ping-pong | 0 | 416.031 / 416.198 / 416.208 |

Every zero-work control is zero; Go Mallocs is zero. LLGo Mallocs is unavailable.
The normal release binary retains a 192-byte collector request in
`runtime.newChanWaiter`, followed by signal initialization. Requested size and
reported collector bytes have different granularity. This supports a bounded
waiter-reuse investigation, reported in
[llgo #3035](https://github.com/xgo-dev/llgo/issues/3035), preserving channel
ownership and synchronization within the current runtime design. It does not
attribute the entire service throughput gap to that allocation.

An independent strict-C11 Cgo reproduction succeeds under Go and fails under
LLGo's generated unqualified `typeof` declaration; GNU11 builds and actually
runs under both. The compatibility report is
[llgo #3034](https://github.com/xgo-dev/llgo/issues/3034). Both native benchmark
builds use the same GNU11 option.

## Validation and remaining work

Affected Go race coverage is **95.30% phase, 95.04% shared driver and 96.7%
independent reference module**. Pinned LLGo actually executes driver/phase and
native-C/reference protocol tests. The root semantic inventory increases to
32 packages across the same eight LLGo shards, with 31 owned coverage gates;
the independent reference module runs separately. Final-head full CI remains
required before merge.

The initial source `7003744` preflight fails because zero warmups encode as
JSON null. No performance sampling starts. The failure and original binaries
remain in `artifacts/service-references-20261109`. Source `ee28d45` fixes empty
arrays and adds a regression, then rebuilds all four binaries and passes 162
two-seed/resource probes before this fixed event. A metadata-only unresolved
versus resolved dylib-path comparison is also corrected before timing; both
declared and actually loaded names/hashes remain recorded.

Raw data, source bundles, checks and independent audits are retained under
`artifacts/service-references-arrays-20261109`. This event does not complete
Wasmtime/Wasmer/WasmEdge concurrent adapters, P3 I/O applications, remaining
serial/native/compiler deficits or controlled release acceptance. The original
32-case serial/sustained compiler event and this service event are not pooled.
