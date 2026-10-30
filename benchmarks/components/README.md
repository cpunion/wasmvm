# Component phase probes

`pkg/component/benchmark_test.go` separates compile, Store creation with
instantiation/close, first call and warm calls. The first-call metric is one
fresh-instance observation per benchmark sample, outside the warm timer.
`SharedCoreSessionWarm` is an internal diagnostic that amortizes one existing
Core session across calls; it is not the public Function.Call API latency.

```sh
go test -run '^$' -bench '^BenchmarkComponentRuntime$' -benchtime=300ms -count=5 ./pkg/component
llgo test -O3 -lto=full -run '^$' -bench '^BenchmarkComponentRuntime$' -benchtime=300ms -count=5 ./pkg/component
```

The independent C runner uses the component API from Wasmtime 48.0.1. Its SDK
URL, SHA256 and license are pinned in `benchmarks/tools-*.json`; use
`benchmarks/install_tools.py --sdk` to verify/install the SDK. A filtered copy
of that manifest can retain only the Wasmtime SDK to avoid unrelated downloads.
Compile against that SDK's `include` and `lib` directories:

```sh
clang -O3 -I "$SDK/include" -L "$SDK/lib" -Wl,-rpath,"$SDK/lib" benchmarks/components/c/wasmtime.c -lwasmtime -o "$OUTPUT/wasmtime-component"
"$OUTPUT/wasmtime-component" internal/component/decode/testdata/composition.wasm 1000000
```

The runner verifies every result and reports load, engine initialization,
compilation, Store/linker setup, instantiation, export lookup, first call and
aggregate warm time separately. Process startup and runner compilation stay
outside those timers. Both warm probes disable fuel; wasmvm still checks
cancellation and its logical limits, while the reference uses Wasmtime's
default Store policy. These descriptive samples are not calibrated performance
acceptance or identical safety-policy measurements. LLGo allocation counters
remain unavailable; zero reported allocations is not treated as zero allocation.

The CLI is used only as a separate semantic cross-check, following the
[official WAVE invocation syntax](https://docs.wasmtime.dev/cli-options.html#running-custom-component-exports):
`wasmtime run --invoke 'add(4, 5)' internal/component/decode/testdata/composition.wasm`
returns `9`. CLI process timing is not a warm-call comparison.


`BenchmarkResourceInstructions` calls the resource fixture's `churn(128)` export.
The timer contains guest looping, resource.new/rep/drop and the Core destructor;
compile, Store setup, lookup and first call are excluded. `ns/resource` amortizes
one public component call across 128 complete resource lifetimes. It is not a
pure handle-table operation measurement. Use the same fixture and batch size
with the existing C runner's optional fourth argument:

```sh
go test -run '^$' -bench '^BenchmarkResourceInstructions$' -benchtime=200ms -count=3 ./pkg/component
llgo test -O3 -lto=full -run '^$' -bench '^BenchmarkResourceInstructions$' -benchtime=200ms -count=3 ./pkg/component
"$OUTPUT/wasmtime-component" pkg/component/testdata/resources.wasm 5000 128
```

Divide the reference's `warm_ns_per_call` by 128 to compare the same complete
resource lifetime. Both probes check every batch result. The source base rejects
resource canonical instructions and therefore has no successful throughput row.
The existing policy differences and uncalibrated local-sample limits apply.


`BenchmarkResourceTable` in `internal/component/canonical` isolates New + Rep +
Drop with no guest code or destructor. Run it separately from compiler/test load.
For a source comparison, extract the exact base and add only this unchanged
benchmark file. Keep these pure-table results separate from complete guest
resource lifetimes; the reference C API does not expose a comparable primitive.
