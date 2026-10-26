# Owned subsystem baselines

`subsystems-v1` complements the pure Core 2.0/competitor portfolio. It measures
only the exact wasmvm base/head implementations under Go and LLGo. It has no
competitor rows and is not a replacement for the common throughput matrix.
All fixtures and their generator are authored here and pinned by SHA-256 in
the [manifest](manifest.json), with wasm-tools 1.258.0 as binary producer.

The host series calls a typed `host.mix(i32, i32) -> i32` import once or 1,000
times per exported invocation. Callback arithmetic matches the independent
Python rotate/add/xor oracle. Each first, warmup and measured call verifies
both the checksum and the exact number of host callbacks. Host binding has its
own timer. The normalized `ns/callback iteration` includes guest loop/control,
VM transitions, callback arithmetic and marshalling; no call-control or clock
median is subtracted to invent an isolated callback cost. The pure series still
rejects all imports and start sections.

Preparation cases vary independently: one empty function, 512 small functions,
two functions with 4,096 additions each, control nesting 256, 16,384 locals,
128 imported functions, eight 64-element tables with 128 globals, and a 64 KiB
active data segment. Negative cases cover a malformed section, a typed stack
mismatch and 1,048,577 declared locals (one above the default decoder quota).
The generated excessive-local WAT source is omitted to avoid storing four MB of
repeated text; its source hash and tiny compressed Wasm binary are retained.

Every preparation sample creates a fresh Engine and Store, compiles with
default limits, binds explicit inert host functions, instantiates and waits for
final cleanup. It never invokes an exported function and rejects start sections.
Read, Engine/Store creation, compile/rejection, host binding, instantiate and
cleanup are separate actual intervals. Invalid modules have no binding or
instantiation interval; these are labeled **not applicable**, never zero-cost
execution. Metadata/error checks and usage snapshots are outside timers. Memory
usage is the Store's logical live-resource accounting, not a peak RSS estimate.

There are three independent processes per implementation/case, with seven
fresh preparation samples per process, or first call plus two warmups and seven
measured calls for the host series. Summaries take medians within processes,
then the median of process medians. The full four-build matrix is 156 processes.
Raw JSON retains toolchain/source identities, fixture hashes, readiness, wall
time, every sample, resource usage and failure context. These hosted timings
are advisory and do not establish M2's controlled acceptance gates.

LLGo's native BDWGC `MemStats.Mallocs` is currently unpopulated
([llgo#2923](https://github.com/xgo-dev/llgo/issues/2923)). Therefore
`testing.AllocsPerRun` can report a false zero despite actual allocations.
The shared interpreter reuse test still executes its repetitions under both
compilers, but its zero-allocation assertion applies to Go's working counter.
No LLGo zero-allocation result is claimed. Reliable event instrumentation and
controlled allocation/performance gates remain required in M2.
