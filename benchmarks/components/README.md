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


`BenchmarkOwnedResourceLifetime` measures four public calls per lifetime:
make, get(borrow), pass(own), consume(own). Export lookup, startup and the first
lifetime stay outside timing. Each result and the final destructor checksum are
checked. `benchmarks/components/c/resources.c` uses the pinned Wasmtime C API
and the same binary/four-call sequence, with opaque resource wrappers and dynamic
move checks. Compile it with the same SDK arguments as the scalar C probe:

```sh
go test -run '^$' -bench '^BenchmarkOwnedResourceLifetime$' -benchtime=200ms -count=3 ./pkg/component
llgo test -O3 -lto=full -run '^$' -bench '^BenchmarkOwnedResourceLifetime$' -benchtime=200ms -count=3 ./pkg/component
"$OUTPUT/wasmtime-resources" pkg/component/testdata/ownership.wasm 100000
```

The reference also checks borrowing, moved-handle rejection and explicit drop
outside its warm timer. The negative stale-handle case runs last; its canonical
trap locks down the reference instance. Keep these four-boundary lifetime
measurements separate from the 128-iteration guest-loop probe and pure-table
probe. No calibrated performance or identical embedding error policy is implied.


`BenchmarkEnvironmentArgumentCount` in `pkg/wasi/preview2` executes an actual
WASI environment call, copies three argument strings into guest memory and
returns their count. Compilation, binding, lookup and the first call are outside
timing; guest post-return resets the fixture allocator on every iteration.
`c/environment.c` uses `wasmtime_component_linker_add_wasip2` and an explicit
`wasi_config_t`, rather than a handwritten reference callback. It additionally
checks full arguments, environment tuples and default current directory outside
the warm timer. Use the same pinned SDK/compile flags as the other C probes.

```sh
go test -run '^$' -bench '^BenchmarkEnvironmentArgumentCount$' -benchtime=200ms -count=3 ./pkg/wasi/preview2
llgo test -O3 -lto=full -run '^$' -bench '^BenchmarkEnvironmentArgumentCount$' -benchtime=200ms -count=3 ./pkg/wasi/preview2
"$OUTPUT/wasmtime-environment" pkg/wasi/preview2/testdata/environment.wasm 100000
```

All phases retain the usual unmetered/default-reference policy differences.
The source base has no P2 provider, so it has no successful interface timing;
the existing component PublicWarm probe is a separate regression diagnostic.

## WASI exit lifetime

`BenchmarkExitCall` prepares batches of fresh component instances before timing
`exit-with-code(42)`. Every call must return a typed exit and reclaim its Store.
Compilation, import binding/instantiation and export lookup are excluded. This
operation cannot reuse an exited instance and is not a warm repeated call.

`c/exit.c` uses Wasmtime 48.0.1's actual WASIp2 provider on the same fixture. It
checks both operations and codes 0/1/42/255, then reports file/engine/compiler/
linker, per-instance setup, instantiate, lookup, exit call and explicit Store
verification/destruction as distinct phases. Its call timer excludes Store
destruction; own calls include cancellation and deferred Store reclamation.
Keep both reference phases visible when comparing these different policies.

Build with the pinned SDK as for `c/environment.c`, then run:

```sh
"$OUTPUT/wasmtime-exit" pkg/wasi/preview2/testdata/exit.wasm 10000
```

Source base has no exit provider: retain that unsupported result, and use the
existing component public-call benchmark separately for common-path regression.
