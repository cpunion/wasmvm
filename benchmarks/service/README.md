# Shared concurrent service references

This independent benchmark module uses the same bounded admission, broadcast
readiness barrier and raw measurement driver as the owned engine's `service`
mode. It imports only `internal/bench/service` from the root module: no owned
decoder, compiler, VM or WASI provider is used by the reference engines.

- Wazero **1.12.0**, built with pinned Go, compiles a module once and creates
  one independent instance and argument/result stack per worker. Both interpreter
  and compiler modes support explicit default or enabled context-termination
  policies. Imports and start functions are rejected after engine validation.
- Native C compiles the unchanged `benchmarks/applications/sha256.c` kernel
  with Clang **22.1.8**, `-O3 -std=gnu11`, and calls it through the ordinary host
  C bridge. The source's standalone CLI is renamed and never invoked. Go and
  LLGo builds use their own schedulers with the same admission driver. This is
  an ordinary C algorithm under a host-language service, not a bare C scheduler;
  it has no Wasm safety, fuel, cancellation, JIT or Store lifecycle contract.
- The LLGo reference binary supports native C only. Wazero is an independent
  Go-built reference; selecting it in the LLGo binary returns an explicit error.

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
numbers as concurrent service results. Wasmtime/Wasmer service adapters and
WASI P3 I/O application comparisons remain separate deliverables.
