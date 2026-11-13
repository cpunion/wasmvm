# Bounded concurrent embedding

The existing phase adapter now has an opt-in `service` mode. It compiles one
pure Core module once, instantiates one independent Store per worker, and keeps
workers and clients alive across the first, warmup and measured batches. The
queue capacity equals the worker count. Clients have at most one outstanding
request each. Each batch waits for an actual ready acknowledgement from every
client, then broadcasts one shared start gate. First-request arrival time is
the common release epoch, so scheduler wakeup is included. Worker readiness is
joined during setup. Every reply must pass the existing signature/checksum oracle.
This is an embedding application, not a new VM scheduler or production API.

Use the existing ready/run and optional sampleReady/sample protocol. Existing
api/dispatch modes and the routing portfolio keep their original schema and
semantics. Service-specific counts require explicit `--mode service`; aggregate
recording is limited to1048576 requests. Failed setup/calls/cancellation do not
produce a successful result. Shutdown joins all client/worker goroutines before
closing Stores, the shared compiled module and engine.

```sh
go run ./benchmarks/embedded/wasmvm \
  --mode service --backend native \
  --wasm benchmarks/embedded/fixtures/call-control.wasm \
  --iterations 1 --seed 17 --expected 17 \
  --workers 4 --clients 64 --requests 16 --samples 7 --warmups 2
# After the ready record, send: run
```

The first interval is the first entire batch, not one exported call. Retain
three different observations without subtraction:

- Batch wall time includes admission/backpressure, all public calls and reply
  delivery. Throughput is completed requests divided by this actual interval.
- Request latency runs from attempted queue admission through observed reply;
  retain all raw samples, including queue waits, scheduler work and pauses.
- Each public API interval uses the original timed kernel boundary. It does
  not include the surrounding application queue. Signature/checksum checking
  remains outside that API interval and inside the complete service request.

Setup/load/prepare/instantiate/lookup, worker setup and completed cleanup remain
separate. Per-batch readiness has its own directly measured interval.
Sample allocation, the readiness barrier, MemStats queries and reduction stay outside
the batch timer. Report first/warmup separately; derive per-process warm means,
throughput and nearest-rank p50/p99 from the retained raw request samples before
comparing independent processes. Requests within one process are not independent
machine runs. Never infer a guest-only interval by subtracting these measures.

Runtime allocation bytes and heap snapshots have different meanings under the
two host compilers: Go reports managed allocations, LLGo reports BDWGC collector
bytes. Starting with the pinned LLGo56fa9ae baseline, Mallocs reports managed
allocation events under both compilers. Earlier LLGo records omitted it; do not
interpret their missing field as zero allocations. Neither counter covers all
native/JIT/linear memory; snapshots are not peak RSS. Optional
`--service-resources` uses a **separate process cohort** to obtain actual idle
OS-thread/RSS snapshots and getrusage process-lifetime peak RSS. Its ps calls
can perturb later work, so those timings cannot stand in for the uninstrumented
throughput cohort. Thread snapshots are not a measured peak.

## Declared comparison boundary

The next application experiment uses both committed SHA-256 guest producers,
the existing independent checksum oracle, seed17,64 message blocks/request,
64 clients,16 requests/client/batch and1/2/4 workers. Use five fresh process
blocks, two warmups/seven measured batches, both owned backends and the pinned
Go/LLGo release profiles. Prepare/build/check everything before timing; keep
one event and all adverse/failed outcomes. Resource cohorts retain the same
workloads/capacities and use first+one measured batch without warmup. These are
separate diagnostics, never retimed replacements for a failed throughput run.

The shared driver now lives in `internal/bench/service`, without a dependency
on the owned VM. The [independent reference module](../service/README.md) supplies
wazero interpreter/compiler with explicit termination policies, and the ordinary
C SHA-256 algorithm under both Go and LLGo admission/scheduling. The owned engine
keeps one Store per worker and the original public typed call boundary. All use
the same broadcast barrier and raw batch/request/API intervals. The common
harness overlay includes this driver when comparing historical engine revisions.

The initial strict-C11 Cgo build fails under LLGo while the identical Go build
succeeds: LLGo emits unqualified GNU `typeof` in its wrapper. The standalone
four-way C11/GNU11 build/run reproduction is reported in
[llgo #3034](https://github.com/xgo-dev/llgo/issues/3034). Both native reference
builds use the same GNU11 policy; no failed build is relabeled as execution.

Wasmtime/Wasmer/WasmEdge adapters now use the same driver. The complete fixed
[72-profile event](../../docs/research/performance-service-all-runtimes-2026-11-10.md)
retains every throughput/resource process and independent control, separately
from the earlier 54-profile event. P3 I/O applications and controlled acceptance
remain open. Existing serial results do not establish concurrent service rates;
passing adapter tests alone supplies no throughput result. This addition changes
no engine source or accepted proposal inventory and closes no release gate.

Go's M:N scheduling and LLGo's pthread/BDWGC choices remain explicit runtime
policies. Application workers bound CPU admission for both. GOMAXPROCS is
recorded, not treated as a native guest CPU quota. Do not classify accepted
scheduler/collector differences as compiler bugs or infer a speedup from static
runtime mechanisms. Any claimed benefit needs the actual throughput, latency,
thread and memory evidence above. Known accepted LLGo cache behavior is outside
this experiment; source timestamps are not normalized.
