# Direct embedding phase benchmarks

These runners measure execution through the engines' embedding APIs. Their
purpose is to separate the guest-call interval from setup, compilation, I/O,
and teardown. They complement the earlier [CLI latency series](../README.md);
the phase fixtures have a different ABI and their measurements are not a
subtraction-based decomposition of an official CLI's total.

## Measurement boundaries

| Metric | Actual measured interval | Interpretation |
| --- | --- | --- |
| Process ready | Parent starts the process until it receives the child's ready line | OS launch, loader/native or Go initialization, and pipe notification; not bare fork cost |
| Load | Read the binary file into host memory | OS pages may be warm; no filesystem cache flushing |
| Engine/store setup | Construct backend configuration and runtime/engine/store objects | Includes the selected embedding's initialization work |
| Decode + validate/compile | Prepare the binary using the runtime's actual API | Combined for wazero, Wasmtime, and Wasmer; WasmEdge interpreter exposes parse and validate separately |
| Instantiate | Instantiate the prepared module | No guest start section, WASI, or imports are permitted |
| Export lookup | Resolve the exported `run` function | Signature checks happen outside the timing interval |
| First call | First invocation through the embedding API | Captures first-use work separately from warmed execution |
| Warm call | One API invocation after first call and two extra warmups | Includes guest work and unavoidable embedding dispatch/marshalling; excludes file I/O, setup, compilation, lookup, assertions, and printing |
| Cleanup | Explicit module/instance/store/engine close or delete operations | Host process teardown is outside this interval |
| Clock pair | Adjacent timestamp reads, recorded as a control | Never subtracted; tiny measurements near this scale are resolution-limited |
| Process wall | Parent launch through process completion and pipe draining | Includes all ten calls, protocol/JSON work, assertions, IPC, and shutdown; kept separately in raw JSON |

The child prints `{"ready":true}` and waits for `run\n` on stdin. The parent
records readiness before acknowledging it. No explicit measured engine API or
guest work begins before this barrier; package and native-library initialization
remain part of readiness. Startup notification therefore cannot overlap measured
API phases. Each API timer ends before its result/error is checked. Every
first, warmup, and measured invocation must return the expected unsigned i32
checksum. Invalid output, traps, incomplete matrices, and timeouts fail the job.

Argument parsing, pure-module prechecks, signature and checksum checks, JSON
serialization, and scheduling gaps sit outside API timers. Phase medians are
not an additive partition of whole-process time; neither sums nor residuals
are presented as a measurement of an otherwise unobserved operation.

The wazero Go runner uses `CallWithStack` with reused argument/result storage. The owned public API runner reuses its argument Values; its current API allocates result/continuation storage inside the timed call. C runners
also reuse argument/result slots. Stack setup and checksum checks sit outside
call timers. These are still **API-call wall times**, not isolated CPU cycles or
an exact count of guest instructions. The call-control module exposes the
remaining host-call boundary. Its median, the startup median, and clock samples
are never subtracted from another measurement.

## Fixtures and normalization

Four owned Core-MVP modules provide six cases: a call control, small/sustained
integer mixing, source-level function calls, and small/sustained memory access.
They export only `run(seed: i32, iterations: i32) -> i32`. Seeds and loop counts
are runtime parameters; no module imports, start section, WASI, or guest output
are present. Function calls may be inlined by native compilers.

Memory starts each invocation with a loop clearing its 32 KiB working region.
That reset cost is explicitly included in execution and amortized over the
workload iteration count. Alternate-seed and repeated-call tests verify that
mutable state does not leak into later samples. The independent Python oracle,
WAT/Wasm hashes, producer version, checksums, and small validation vectors are
recorded in [the fixture manifest](fixtures/manifest.json).

The small Go/C negative-test modules in `wazero/testdata/` are also authored
for this project, with adjacent WAT sources. Their Wasm bytes were generated
using wasm-tools 1.258.0 (`wasm-tools parse input.wat -o output.wasm`); they are
test inputs rather than measured workloads.

Reports show `ns/call` for the control and `ns/iteration` for compute cases.
An iteration is one source workload loop iteration, **not** a Wasm instruction
or a native instruction. There are three fresh process repetitions per case;
each contains one first call, two warmups, and seven measured warm calls. The
reported warm median is the median of the three process medians, avoiding the
claim that 21 calls inside three processes are 21 independent process samples.
All process/configuration/workload cases are interleaved in seeded random order.

## Backends and provenance

- **Owned wasmvm:** exact base/head commits, each built with pinned Go 1.27.3
  and LLGo runtime revision `8b1389c627c77a02e56af03f08d9cdc19edffe77`.
  `own_build.py` archives committed source into isolated trees and overlays only
  the two head instrumentation directories (`internal/bench/phase` and
  `benchmarks/embedded/wasmvm`). No engine/public-API file is copied from head
  into base. Source hashes/digests (including the production CLI), shared adapter hashes, generated provenance,
  compiler identities, dependency inventories and binary hashes are retained.
  Untracked files cannot enter an archive; tracked modifications fail identity
  verification. A baseline predating Engine/Store is explicitly unavailable;
  neither a reference engine nor a compiler failure can replace it.
- Each own build runs two separately named boundaries. **api** is the primary
  public embedding comparison. **dispatch** is a diagnostic using the same
  validated slot interpreter, reused frames and a cancellable context, without
  public root/marshalling/admission work. Bounds/trap checks and normal polling
  remain enabled in both; optional deterministic fuel is disabled. Never use
  dispatch results as a replacement for public API performance.
- Dispatch diagnostics also time decode, validation, IR lowering and slot
  allocation separately inside one outer preparation interval. The outer
  interval includes phase transitions and may exceed the parts' sum. No native
  code is generated in M1, and these parts are not added to preparation again.
- The actual `cmd/wasmvm` binary is also built from each revision/compiler. A
  separate series runs each pure fixture in three fresh CLI processes and checks
  the typed JSON result. Its complete wall time includes startup, file I/O,
  compilation, execution, JSON and exit. It is never subtracted from API phases.
  A usable base/head matrix adds 72 independently measured CLI invocations.

- Wazero v1.12.0 interpreter/native compiler: the same head runner source is
  compiled separately against the dependency selected by the base and head
  checkouts. Source hashes, module checksums, binary hashes, and Go build info
  are recorded. These extra reference base/head rows compare dependency selections; the separately named owned rows measure wasmvm implementation revisions.
- Wasmtime 48.0.1: pinned C API SDK, Cranelift with optimization level 2.
- Wasmer 7.3.0: pinned C API SDK with Cranelift selected explicitly.
- WasmEdge 0.17.1: pinned C API SDK with interpreter mode selected explicitly;
  raw parse and validate durations are recorded under `preparationPartsNs`.

The C APIs and versioned headers are the source of each measured boundary:
[Wasmtime module API](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/c-api/include/wasmtime/module.h),
[Wasmer C API](https://github.com/wasmerio/wasmer/tree/v7.3.0/lib/c-api), and
[WasmEdge C API](https://github.com/WasmEdge/WasmEdge/tree/0.17.1/include/api).
Calling a standalone validation API before a compile API that already validates
would repeat work, so those costs are not presented as disjoint phases. No disk
compilation cache or precompiled artifact is used.

The Go measurement runner is tested with Go, including both backends and race
checks. Its owned-code statement coverage is enforced at 90% or higher. The C
runners have actual positive and negative integration probes, including traps,
wrong checksums/signatures, forbidden imports/start functions, and bad protocol
inputs. Existing VM/evaluator semantic tests remain required under Go and LLGo.
The measurements use Go/LLGo hosts for the owned interpreter, Go for wazero, and C for the other references. They do not claim LLGo native backend acceptance. A separate local probe found that the fixed LLGo runtime
executes the interpreter but the wazero native backend fails on its first call,
tracked by [LLGo #2921](https://github.com/xgo-dev/llgo/issues/2921). The memory
hooks in [LLGo PR #2919](https://github.com/xgo-dev/llgo/pull/2919) fix linking;
they do not establish native-backend execution compatibility.
The complete wazero Go-runner tests retain both backends; that failure is not skipped or presented as a passing LLGo native measurement result. The owned interpreter is required to execute with both host compilers.

## Reproduction

Linux amd64, Python 3.12+, Go 1.27.3, the pinned LLGo toolchain, and a C compiler are required. Installation
and build time are excluded from measurements. Start from clean base and head
checkouts; both must contain the candidate dependency selection.

```sh
python3 benchmarks/install_tools.py --sdk --destination /tmp/wasmvm-phase-tools
WASMVM_LLGO_REVISION=8b1389c627c77a02e56af03f08d9cdc19edffe77 \
python3 benchmarks/phase_ci.py \
  --base-dir /path/to/base --base-sha FULL_BASE_COMMIT \
  --head-dir /path/to/head --head-sha FULL_HEAD_COMMIT \
  --tools-dir /tmp/wasmvm-phase-tools --output-dir artifacts/phases
```

The [PR workflow](../../.github/workflows/benchmarks.yml) builds the C runners,
executes their correctness probes, runs both series, and publishes a combined
comment with owned public API base/head and competitor comparisons first, internal diagnostics and full phase tables in details sections. With a usable baseline there are 15 configurations (four owned API, four owned dispatch, four wazero dependency-reference rows and three other references), six cases and three process repetitions: 270 fresh processes.
Raw JSON contains every call, every phase, setup metadata, and failure records.
Three process repetitions and shared hosted hardware provide an engineering
baseline, not a stable performance budget or a general engine ranking.
