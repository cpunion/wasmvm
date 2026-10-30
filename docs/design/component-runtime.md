# Synchronous component execution

`pkg/component` owns the embedding API and executes ordered plans from the
component validator. It compiles embedded Core binaries with `pkg/vm.Engine`;
all actual instructions use the existing interpreter or native compiler.
There is no competitor fallback and no second Core execution framework.

The current execution profile includes embedded/aliased Core modules and
instances, synthesized instances, child component closures and instantiation,
imports/exports, narrowed instance views, literal values and component starts.
Host bindings supply callbacks and copied values through declared schemas.
Fixed equality-bound value schemas are inferred from those declarations;
generative host resource types, host components and host Core module imports
remain unsupported. Other incomplete executable operations are rejected during
compilation or host binding, before instantiation invokes guest code.

Canonical lift/lower shares `internal/component/canonical` for scalars,
strings, lists, records/tuples, variants, flags/enums and option/results.
Memory32/64, all three string encodings, direct/spilled arguments/results,
realloc and post-return use checked memory capabilities. Own/borrow/resource
conversion, map/fixed-list ABI conversion and async execution remain required;
literal or layout support does not establish those runtime capabilities.

## Calls and lifetime

`vm.Store.WithCaller` spans each component operation. Realloc, Core execution,
callbacks and post-return share one cancellation context, fuel balance and
existing Core stack/reentry policy. Caller memory operations accept only
same-Store capabilities and copy bytes; a borrowed guest slice never survives
allocation or reentry. Component callback capabilities expire at callback
return and bound pure host reentry too. Codecs retain aggregate byte/value
budgets; each instance caches its immutable type layouts.

Arguments are validated before guest allocation effects. Host aggregates,
literals, start results and exported values use type-directed copies so host
mutation cannot alter cached constants or instance state. Equality-bound host
values are checked across nested interfaces. Ascription restricts visible
instance exports rather than exposing the underlying wider namespace.

Realloc and post-return block canonical leave/reentry. Canonical traps or host
failures lock the affected component's canonical state. Invalid host input
rejected before execution leaves the instance usable. Imported effects are not
rolled back. Core starts and component starts execute during instantiation;
partial failure closes the unpublished Store and its code/resource leases.
Pure component expansion also consumes instance, nesting and aggregate work
budgets, even when it executes no Core instructions and spends no Core fuel.

Engine close prevents new compilation/Store creation. Compiled close prevents
new instantiation while existing instances retain Core code leases. Instance
close requests cancellation; `WaitClosed` observes final reclamation after
active callbacks leave. Successful synchronous results are copied before
post-return runs. There is no claim of resource-transfer transactions or async
cleanup until those capabilities are implemented.

## Embedding and evidence

Create an Engine with `component.NewEngine(component.DefaultConfig())` and
select `config.Core.Backend` before creation. Compile a component, instantiate
it with `component.Imports`, look up a function with `instance.Function(...)`,
and call it with `[]component.Value` plus `component.CallOptions`. Call results
have a nil `Value` for a function without a result. Nested exported interfaces
use a path of names. `instance.Value(ctx, path...)` returns an owned value copy.
Keep Engine, compiled component and instance lifetimes explicit.

Tests run through both own backends and both host compilers, including real
string/host/post-return calls, spilled records, memory64, starts, value bounds,
outer closures, export views, quotas, cancellation and failure. Value-definition
fixtures remain authored grammar tests because the pinned producer cannot
encode their text syntax. See [validation evidence](../validation.md) and the
[phase probes](../../benchmarks/components/README.md). Tiny-call LLGo latency
remains behind Wasmtime and does not meet the performance target. M5/M6 and
WASI 0.2/0.3 worlds remain open.
