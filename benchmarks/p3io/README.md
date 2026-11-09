# Complete embedded P3 HTTP measurement

This module uses the existing `internal/bench/service` admission/measurement
driver. It measures actual P3 HTTP input/output streams, response metadata,
trailer futures and joined cleanup. It does not measure a network socket or
infer a warm-call cost from CLI startup. Production VM/provider code is unchanged.

- Owned Go/LLGo builds use `pkg/component` and `pkg/wasi/preview3`, one World and
  Store per worker, sharing one compiled component. Both backends execute the
  pinned Rust service and its composed middleware/service fixture. Fuel is off;
  logical limits and cancellation remain. Per-call handler lookup/signature
  checks inside the public `CallIncoming` API remain in its timer.
- The independent `p3reference` Go build excludes the owned VM/providers. A
  Rust1.98.0 static C bridge uses Wasmtime48.0.1 Cranelift speed and its actual
  `wasmtime-wasi/p3` and `wasmtime-wasi-http/p3` providers. Each worker owns a
  serialized Store and current-thread Tokio executor. `Store::run_concurrent`
  drives the handler, body and trailer collection. Export bindings are cached;
  the Go C bridge and executor entry remain inside its timer. Outgoing network
  requests are denied. It retains the reference's default unmetered policy.
- `native-go` performs ordinary in-process echo/body collection with the same
  response data and admission driver under Go/LLGo. It has no component, P3,
  guest, JIT or Wasm isolation contract and is a host/application control.

Each response must have status201, `x-guest: rust-p3`, exact payload bytes and
`x-guest-trailer: done`. The returned CRC denotes that completely checked body;
it is not a checksum computed by guest code. Independent Python zlib oracles
are1757737011 for16 bytes and2629362236 for64KiB of repeated
`0123456789abcdef`. Reply-byte leases in the reference are checked before their
private worker can call again; driver shutdown joins every worker before close.
Failed reference workers cannot execute another request on a failed Store.

Preparation, first/warmup/measured batches and cleanup are distinct. The complete
API timer includes invocation, bounded body collection and close. Byte/status/
header/trailer validation follows that timer and remains inside request wall
time. Keep batch throughput, admission-to-reply latency and public API intervals
separate without subtracting them. Returning response headers is not completion.
Allocation/RSS/thread scopes and separate resource cohorts follow the shared
[driver contract](../applications/service-plan.md).

```sh
go -C benchmarks/p3io build -o /tmp/p3-owned .
/tmp/p3-owned --engine wasmvm --backend native \
  --wasm pkg/wasi/preview3/testdata/world-rust-http.wasm \
  --payload-bytes 65536 --expected 2629362236 \
  --workers 2 --clients 16 --requests 4 --samples 7 --warmups 2
# After {"ready":true}, send run followed by a newline.
```

For LLGo use the pinned `-O3 -lto=full` release profile. Ordinary Go examples
without recorded source/compiler/library identities are usage, not qualified
performance results. `WASMVM_INTERLEAVE=1` enables `sampleReady`/`sample` per
measured batch. Zero-warmup output uses empty JSON arrays.

Build and actually execute the independent reference with the pinned Go,
Rust1.98.0/LLVM22.1.8 and Clang22.1.8 toolchains:

```sh
CC=/path/to/clang22 python3 benchmarks/p3io/check_reference.py \
  --output-dir /tmp/p3-reference-checks --rust-target-dir /tmp/p3-rust-target
```

It records separate build/test durations, source/toolchain/archive/executable
hashes, exact platform native linker flags, actual race-test logs and >=90% Go
adapter coverage. It checks that the independent reference imports only the
shared driver from the owned root module. Go adapter coverage is not coverage
of C, Rust, Wasmtime or the specification. The reference executable accepts
`--engine wasmtime --backend compiler` with the same workload arguments.
Use `--engine native-go --backend host` without `--wasm` for the native control.
The `p3host` build tag excludes both the owned VM and Rust bridge from a
standalone Go/LLGo host control; its only owned dependency is the shared driver.

The static bridge avoids a macOS27 LINKEDIT validation failure observed while
linking the initial Rust dylib into Go. The initial failure is retained as
build evidence; no timing started from it. No compiler/module cache was patched.
An optional `suite=p3io` benchmark-workflow dispatch checks the actual Linux
reference with a ten-minute budget and cached locked dependencies. Normal
semantic CI executes the owned measurement tests under both compilers and
platforms; it does not rebuild a Rust runtime on every PR.

This addition supplies execution/measurement mechanisms. Complete fixed
comparisons, controls, both-target evidence and performance/release acceptance
remain separate deliverables; passing tests supplies no performance ranking.

The [fixed 2026-11-10 event](../../docs/research/performance-p3io-2026-11-10.md)
now retains all 36 profiles,360 throughput and36 separate resource processes,
including 161,280 completely checked warm responses. Only 6/36 throughput
controls resolve. The observed P3 deficit remains open; no controlled win or
release gate is claimed.
