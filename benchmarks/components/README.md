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


## Packed byte transfers

`BenchmarkByteListTransfer` echoes 16 KiB through the same project-authored
component. `elements` measures the legacy per-byte Value input; `packed` uses
immutable raw bytes. Payload setup, file/engine/compiler, Store/import binding/
instantiation, export lookup and first call stay outside the warm timer and are
reported as separate metrics. Full first-result contents and per-call endpoints
are checked. cm32/cm64 and host-callback variants are additional semantic tests.

`c/bytes.c` measures the same binary through Wasmtime 48.0.1's dynamic C API,
including owned-result deletion. That API expands a dynamic value per byte.
Also run the typed Rust reference, which accepts a byte slice and returns an
owned Vec<u8>; the C API result alone is insufficient for performance acceptance.
The new Rust profile pins Rust 1.98.0, wasmtime 48.0.1, features and Cargo.lock;
it does not change the own compiler baselines or the prebuilt C SDK profile.

```sh
cargo +1.98.0 build --release --locked \
  --manifest-path benchmarks/components/rust/Cargo.toml \
  --target-dir "$OUTPUT/wasmtime-typed"
"$OUTPUT/wasmtime-typed/release/wasmvm-component-reference" \
  pkg/component/testdata/bytes.wasm 100000
"$OUTPUT/wasmtime-bytes" pkg/component/testdata/bytes.wasm 1000
```

Both references separate load, engine, compile, Store/linker, instantiate, lookup,
first and warm call phases. Typed Rust drops owned results inside the timer;
own calls include normal allocation/GC policy. Fuel is off, own cancellation/
logical quotas remain enabled, and reference Stores use defaults. Preserve the
separate profiles and unmet targets in any report.


## WASI random provider

`BenchmarkRandomBytes` executes secure generation of 16 KiB, canonical byte-list
lowering, a guest length read and post-return allocation reset. It reports file,
engine, compile, entropy/provider initialization, Store/binding/instantiate,
lookup and first-call metrics separately from warm execution. Source base has no
random provider; the existing environment benchmark supplies the independent
base/head common-path regression check.

`c/random.c` runs all five interface operations through Wasmtime 48.0.1's actual
WASIp2 provider before timing the same secure-count export. Reference context/
linker/provider setup and first call are reported separately. The warmed result
is scalar, so dynamic C API per-byte result conversion does not dominate this
reference. Own entropy-seeded AES-256 counter streams and the reference provider
use different random algorithms/policies; preserve them when interpreting times.

Build with the pinned C SDK as above and run:

```sh
"$OUTPUT/wasmtime-random" pkg/wasi/preview2/testdata/random.wasm 10000
```


## Host Core module imports

`BenchmarkHostCoreModule` compiles an external Core module, binds it by name,
and calls the component's add export with changing arguments. File loading,
Engine creation, Core/component compilation, Store/binding/instantiation,
lookup and first call are reported separately from warmed calls. The source
base lacks CompileCore/CoreModules; preserve its unsupported compilation result
and use the unchanged PublicWarm probe for a common-path base/head comparison.

Build `c/host_module.c` against the pinned Wasmtime 48.0.1 SDK. It accepts
`COMPONENT CORE EXPORT COUNT`, checks the first result and every warmed result,
and reports all phases separately. `add` binds math and checks an earlier start;
`apply` binds the rich memory64 module. The typed Rust counterpart uses the
same binaries and pinned Cargo.lock/toolchain:

```sh
cargo +1.98.0 build --release --locked --bin host_module \
  --manifest-path benchmarks/components/rust/Cargo.toml \
  --target-dir "$OUTPUT/wasmtime-typed"
"$OUTPUT/wasmtime-typed/release/host_module" \
  pkg/component/testdata/host-module.wasm \
  pkg/component/testdata/host-core-math.wasm 1000000
"$OUTPUT/wasmtime-host-module" pkg/component/testdata/host-module.wasm \
  pkg/component/testdata/host-core-math.wasm add 1000000
```

Both references drop the original Core module handle before instantiation; the
linker retains its cloned module. Own binding retains a code lease and rejects
new bindings after owner Close. These lifecycle APIs differ. Fuel is disabled;
own cancellation/logical limits remain enabled and reference Stores use default
policy. Wasmtime's GC component-module declaration abort is a capability failure,
not a throughput sample. Typed Rust is the stronger call-boundary reference;
local samples alone do not establish performance acceptance.
