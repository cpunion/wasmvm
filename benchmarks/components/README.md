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


## WASI wall clock

`BenchmarkWallClockNow` reads the standard datetime through canonical lowering,
loads seconds in guest Core code, and returns u64. File/Engine/compiler,
provider initialization, Store/binding/instantiation, lookup and first call are
reported separately. `BenchmarkWallClockRead` isolates the default clock read
and provider record/allocation; these scopes are diagnostic and not additive.
The unchanged environment benchmark supplies the source base/head common path.

Build `c/wall_clock.c` against the pinned C SDK and run:

```sh
"$OUTPUT/wasmtime-wall-clock" pkg/wasi/preview2/testdata/wall-clock.wasm 100000
```

The actual Wasmtime 48.0.1 WASIp2 provider verifies both full datetime records
outside the warm timer, then times the same scalar-result guest export. No
per-field dynamic C result graph dominates this warm reference. Both sources
measure realtime, but own default ticks are explicitly quantized to microseconds;
reference clock precision/default Store policies differ. Fuel is disabled and
own cancellation/logical quotas remain enabled. Local timings and source-base
unsupported APIs do not establish calibrated performance acceptance.


## Poll/monotonic subscription lifetime

`BenchmarkPollClockSubscriptions` invokes `churn(128)`: every guest iteration
creates an immediate monotonic subscription, blocks on it and explicitly drops
its resource. Compilation, Engine/provider/Store binding, lookup and first call
are separately reported outside warm timing. ns/subscription divides complete
batch time by 128; it is not a pure timer or handle-table primitive. Every batch
result and final provider entry count are checked. Source base has no Poller API;
retain the missing-API compilation failures and use the unchanged wall-clock
benchmark separately for common-path base/head regression.

Build `c/poll_clock.c` against the pinned Wasmtime 48.0.1 SDK and run:

```sh
"$OUTPUT/wasmtime-poll-clock" pkg/wasi/preview2/testdata/poll-clock.wasm 10000
```

It uses the actual WASIp2 provider, verifies now/resolution and the in-guest probe
of every remaining operation, then times the same 128-lifetime batch. Scalar C
result conversion is amortized over the batch. Its zero-duration subscriptions
yield to Tokio once, while own immediate subscriptions are ready directly; keep
this policy difference explicit. Fuel is off, own cancellation/quotas remain
enabled, and reference Stores use defaults. Local samples alone establish no
performance acceptance. The original probe's premature ready assertion and its
reference failure are retained separately from the final block-before-ready
probe. See the pinned primary reference's clocks.rs for the scheduling rationale.


## IO transfer and backend costs

`BenchmarkIOWriteBatch` invokes `write-batch(128)`: each guest iteration writes
and flushes 4 KiB through the standard stdout stream. The sink checks first/last
bytes and counts every byte; source ownership and every returned batch count
are verified. File load, Engine/compiler, provider setup, Store/binding,
lookup and first call are separately reported outside warm time. ns/write
amortizes one guest call across 128 complete operations; it is not a pure memcpy.
Source base lacks the IO provider APIs; the unchanged wall-clock probe supplies
the independent common-path base/head comparison.

`BenchmarkIOOutput` separates sink-only, blocking-backend and queued-backend
operation costs on the same checked source. Diagnostics include their normal
allocation/scheduling policies and are not an additive profile of guest time.

`c/io_streams.c` uses Wasmtime 48.0.1's actual WASIp2 provider with explicit stdin
bytes and a custom checked stdout sink. It verifies the full exercise and terminal
absence before timing the same binary/batch. Error and protocol cases are
separate process modes, outside timing:

```sh
"$OUTPUT/wasmtime-io-streams" pkg/wasi/preview2/testdata/io-streams.wasm 1000
"$OUTPUT/wasmtime-io-streams" pkg/wasi/preview2/testdata/io-streams.wasm 1 error
"$OUTPUT/wasmtime-io-streams" pkg/wasi/preview2/testdata/io-streams.wasm 1 parent-drop
```

Scalar C result overhead is amortized across 128 operations. Both sinks use the
same endpoint/count checks. Reference scheduling/buffering and default Store
policies differ; own cancellation/quotas remain enabled and fuel is off. Preserve
the observed reference acceptance of unchecked stdio writes, rejected by own
WIT permit checks. Initial error-probe result alignment and harness-only negative
failures are retained separately from the corrected final probes. Local samples
alone establish no calibrated performance acceptance.

### WASI filesystem

`BenchmarkFileSystemStatBatch` and `c/filesystem.c` execute the same pinned
`filesystem.wasm`. Each call acquires one preopen, opens the same 16-byte file,
performs 128 stats with an in-guest size assertion, and drops both descriptors.
Rates include amortized open/drop costs. Loading, engine creation, compilation,
provider setup, instantiation, function lookup, first call and teardown are
reported separately. File creation/semantic fixture mutations are outside warm
timing. Run on the same local filesystem with warmed metadata caches:

```sh
go test -run '^$' -bench '^BenchmarkFileSystemStat' -count=3 ./pkg/wasi/preview2
"$OUTPUT/filesystem-wasmtime" pkg/wasi/preview2/testdata/filesystem.wasm "$SANDBOX" 1000
```

The C runner uses Wasmtime 48.0.1's actual WASIp2 provider, grants only its explicit
sandbox, and executes the full shared exercise before timing. C scalar result
conversion is amortized across the batch. Wasmtime and own provider scheduling
and safety policies differ. `BenchmarkFileSystemStatStages` measures host fstat
and stat-value construction independently; these diagnostic scopes are not an
additive cost decomposition. Source base has no FileSystem API; retain that
failure and compare the unchanged wall-clock workload on source base/head.

### WASI IP literal resolution

`BenchmarkNetworkLiteralBatch` and `c/network.c` execute the same
`network-common.wasm`. Each call obtains a network handle, then performs 128
literal lookup lifetimes: resolve, subscribe/block/check-ready, next address,
next EOF, drop pollable and drop stream. The guest checks `127.0.0.1`; the shared
semantic exercise also checks `::1`. This measures provider/resource/canonical
costs without DNS server latency or network packets. Load, engine, compile,
provider/Store setup, bind, lookup, first call and teardown are separate metrics.
Pure parse and IP-value construction diagnostics are separate, nonadditive
scopes. Fuel is off; own quotas/cancellation remain enabled. Reference defaults
and scheduling differ, with name lookup explicitly enabled in both providers.

Wasmtime 48.0.1 fails to bind the full pinned consumer because its C WASIp2 linker
lacks `network-error-code`. Retain that failure. The common consumer omits only
that unused requirement and has the same authored guest workload; both engines
run the exact common binary. No replacement implementation is installed in the
reference linker. Source base lacks NewNetwork; the unchanged wall-clock probe
supplies an independent own source-base/head comparison.

### WASI TCP persistent transfer

`BenchmarkTCPRoundtripBatch` and the Rust `tcp` binary execute the identical
`tcp-common.wasm` on actual providers. Before timing, both run the complete
IPv4/IPv6 semantic exercise. A separate session opens a loopback IPv4 connection;
each timed guest call makes 128 roundtrips, transferring 4 KiB each direction
and checking lengths/pattern endpoints. Ten batches warm the persistent session.
Socket options, connection creation, bind/accept, guest shutdown and drops stay
outside the measured transfer loop. Load, engine, compilation, provider/Store
setup, bind/instantiate, lookup, session setup, first batch, session shutdown
and owner teardown are reported separately. Fuel is off; own quotas and
cancellation remain enabled. These local serial samples are descriptive.

Build the reference with `cargo +1.98.0 build --release --manifest-path
benchmarks/components/rust/Cargo.toml --bin tcp`, then invoke `tcp COMPONENT COUNT`.
Wasmtime 48.0.1's actual Rust WASIp2 provider uses
`inherit_network().allow_tcp(true)`; own callbacks authorize loopback only.
Other default scheduling/safety policies differ. Keep TCP_NODELAY at the OS
default in both. The full consumer's missing `network-error-code` and the C API
common consumer's TCP access-denied are separate retained failures: the pinned
C API has no TCP-enable setter. `c/tcp.c` is that diagnostic probe, not a valid
TCP performance reference. The normal Rust build uses pinned stable 1.98.0;
an initial default-nightly dependency compile failure is retained without
patching the shared dependency cache. Source base lacks NewTCP; preserve that
API-baseline failure and compare the unchanged wall-clock workload separately.

### WASI UDP persistent transfer

`BenchmarkUDPRoundtripBatch` and the Rust `udp` binary run `udp-common.wasm`.
Both execute connected/unconnected IPv4/IPv6 assertions before timing a
persistent connected IPv4 session. Each guest call checks 128 roundtrips of
4 KiB per direction, including check-send/send and receive/readiness. Ten
batches warm the session. Load, engine, compile, provider/Store setup,
bind/instantiate, lookup, endpoint/stream setup, first batch, warm batches,
owned-resource drops and owner teardown remain separate. Own fuel is off;
quota/cancellation checks remain enabled, and grants authorize loopback only.
The actual Wasmtime 48.0.1 Rust WASIp2 provider uses explicit
`inherit_network().allow_udp(true)`, with different default scheduling/policy.
Build with pinned `cargo +1.98.0 ... --bin udp`; invoke `udp COMPONENT COUNT`.
No replacement reference provider is installed. Preserve source-base NewUDP
API absence separately from the unchanged wall-clock probe.
