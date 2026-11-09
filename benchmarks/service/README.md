# Shared concurrent service references

This independent benchmark module uses the same bounded admission, broadcast
readiness barrier and raw measurement driver as the owned engine's `service`
mode. It imports only `internal/bench/service` from the root module: no owned
decoder, compiler, VM or WASI provider is used by the reference engines.

- Wazero **1.12.0**, built with pinned Go, compiles a module once and creates
  one independent instance and argument/result stack per worker. Both interpreter
  and compiler modes support explicit default or enabled context-termination
  policies. Imports and start functions are rejected after engine validation.
- Optional Go C bridges select Wasmtime **48.0.1** Cranelift speed, Wasmer
  **7.3.0** Cranelift, or WasmEdge **0.17.1** interpreter. They prepare one
  immutable module/AST and instantiate independent mutable state per worker.
  Their default unmetered policy has no context termination or disk cache.
  The callback interval includes the ordinary Go C bridge, argument packing
  and checked SDK call. Guest checksums are checked after that interval.
- Native C compiles the unchanged `benchmarks/applications/sha256.c` kernel
  with Clang **22.1.8**, `-O3 -std=gnu11`, and calls it through the ordinary host
  C bridge. The source's standalone CLI is renamed and never invoked. Go and
  LLGo builds use their own schedulers with the same admission driver. This is
  an ordinary C algorithm under a host-language service, not a bare C scheduler;
  it has no Wasm safety, fuel, cancellation, JIT or Store lifecycle contract.
- The LLGo reference binary supports native C only. Other references require
  Go; selecting an unavailable adapter returns an explicit error.

```sh
go -C benchmarks/service build -o /tmp/service-reference .
/tmp/service-reference --engine wazero --backend compiler \
  --wasm benchmarks/applications/fixtures/sha256-clang.wasm \
  --seed 17 --iterations 64 --expected 3381501185 \
  --workers 4 --clients 64 --requests 16 --samples 7 --warmups 2
# After {"ready":true}, send run followed by a newline.

/tmp/service-reference --engine native-c --backend native \
  --seed 17 --iterations 64 --expected 3381501185 \
  --workers 4 --clients 64 --requests 16 --samples 7 --warmups 2
```

For recorded native measurements, set `CC` to the pinned Clang binary before
building. LLGo uses `llgo build -O3 -lto=full` and the repository's compiler pin.
The normal Go build above is a usage example, not a provenance-qualified result.
Do not compare measurements built with unidentified compilers.

Build exactly one SDK adapter per executable. Wasmtime and Wasmer export common
`wasm_*` symbols and must not be linked into the same process. For an existing
pinned SDK, set `CGO_CFLAGS=-I/path/to/include` and
`CGO_LDFLAGS='-L/path/to/lib -Wl,-rpath,/path/to/lib -lwasmtime'`, then use
`go -C benchmarks/service build -tags=wasmtime -o /tmp/wasmtime-service .`.
Replace the tag/library for Wasmer or WasmEdge. Wasmtime/Wasmer require
`--backend compiler`; WasmEdge requires `--backend interpreter`. The other
arguments and input/output protocol above are unchanged. SDK version mismatch
fails rather than silently changing the reference.

On Linux, reuse `benchmarks/install_tools.py --sdk` and run:

```sh
CC=/usr/lib/llvm-22/bin/clang python3 benchmarks/service/check_sdks.py \
  --tools-dir /path/to/verified-tools --output-dir /tmp/service-sdk-checks
```

This compiles and executes race-enabled semantic tests for all three SDKs,
enforces 90% Go adapter statement coverage and builds ordinary reference
executables. Compile/link and execution durations, commands, compiler and SDK
hashes are separate. This is correctness evidence, not a performance event or
C/SDK coverage. Tests cover independent persistent memory/globals, both C/Rust
SHA producers, alternate seeds, invalid modules, imports/starts, signatures,
instantiation failure, traps, bad checksums and cleanup. Opt-in benchmark build
jobs reuse their already installed SDKs for these checks; normal semantic CI
does not download three SDKs on every PR.

Pinned Wasmer's module wraps immutable compiled code and its instance constructor
uses the supplied Store. Its generic header declares `wasm_module_share/obtain`,
but the pinned library does not export them. The adapter uses the actual
[7.3.0 module implementation](https://github.com/wasmerio/wasmer/blob/v7.3.0/lib/c-api/src/wasm_c_api/module.rs)
and [instance implementation](https://github.com/wasmerio/wasmer/blob/v7.3.0/lib/c-api/src/wasm_c_api/instance.rs)
and tests independent state; it does not serialize/deserialise code per worker.

The protocol is `ready` / `run`; `WASMVM_INTERLEAVE=1` additionally requires
`sampleReady` / `sample` for every measured batch. JSON retains first, warmup,
measured batches, preparation phases, completed cleanup and every request/API
interval. Native C omits inapplicable Wasm preparation phases. Resource
instrumentation (`--service-resources`) belongs to a separate cohort: it runs
`ps` outside the batch timer and records actual RSS/threads plus process-lifetime
peak RSS. Go and LLGo allocation counters have distinct scopes.

Use both committed C/Rust Wasm producers, independent checksum checks, fixed
capacities and order, fresh processes and identical-binary controls. Complete
builds/tests/probes before timing. Preserve failed and adverse outcomes; do not
subtract phases, mix scheduler policies, pool separate cohorts or reuse serial
numbers as concurrent service results. SDK concurrent measurements and WASI P3
I/O application comparisons remain separate deliverables until their complete
data is retained; passing adapter tests supplies no throughput result.

The [fixed 2026-11-09 event](../../docs/research/performance-service-references-2026-11-09.md)
retains all 54 profiles and all unresolved A/A controls. Owned callbacks use
typed `Function.Call`, whereas wazero reuses raw stack storage; this policy
difference is explicit. The standalone channel-waiter allocation finding is
reported in [llgo #3035](https://github.com/xgo-dev/llgo/issues/3035), separately
from any throughput or scheduler claim. The subsequent
[complete event](../../docs/research/performance-service-all-runtimes-2026-11-10.md)
includes all three SDKs, 72 profiles and every lifecycle/resource/control outcome.
