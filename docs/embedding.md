# Embedding

`pkg/vm` provides Engine, CompiledModule, Store, Linker, Instance, Extern and
Execution. See the executable [VM examples](../pkg/vm/example_test.go).

An Engine validates and compiles immutable modules. Each Store serializes its
own guest execution and mutable objects; distinct Stores can execute in
parallel. Select `EngineConfig.Backend` (`interpreter` or `native`) explicitly.
`Profile: "core3"` enables the complete final core profile. Core Threads remains
an independent feature bit. Native execution has no interpreter fallback.

Set `EngineConfig.Features |= wasm.CompactImports | wasm.WideArithmetic` to
enable the separately pinned [core extensions](design/core-extensions.md).
These are also available to embedded core modules in components.
`wasm.CustomPageSizes` additionally allows `MemoryType.PageSize: 1`; zero uses
the default 65536-byte pages. Guest page counts follow that memory's type, while
Store `MaxMemoryPages` and usage counters retain 64-KiB host quota units,
rounding up per memory. Shared maximum reservations use those same quota units.
`wasm.CustomDescriptors` with GC (or `Profile: "core3"`) enables the pinned
experimental Core descriptor and exact-reference extensions. JavaScript
prototype integration is excluded. `ValueType.Exact` and `WithExact` retain
exactness alongside nullability and canonical type identity.

`wasm.StackSwitching` with GC, typed function references and exception handling
(or `Profile: "core3"`) enables experimental typed guest continuations and
control tags. Captured guest frames and references consume Store heap quotas;
restoration also checks active frame/slot limits. Handles are one-shot, and
host callbacks do not expose or capture Go/C stacks. An unhandled control tag
returns `*vm.Suspension`, matching `wasm.ErrSuspension`. `Payload` returns the
tag and owned argument Values; release those Values and the signal explicitly.
This terminal host-boundary signal is distinct from an exception, trap or
component async continuation. See the [extension contract](design/core-extensions.md).

Typical order:

1. Create an Engine and compile bytes.
2. Create a Store and Linker; define typed host functions or other imports.
3. Instantiate, obtain an exported Extern, and call it with typed Values.
4. Release owned reference Values and close the Store, module and Engine.

Host callbacks use Caller for scoped memory copies, controlled reentry and
collection. A Caller expires at callback return. Numeric Values preserve exact
bits; guest references are opaque Store-owned roots. Copying a reference Value
does not create a new root: use Root and Release explicitly.

Store.Close requests cancellation; WaitClosed waits for calls and native leases
to drain. Engine.Close prevents new compilation/Stores while existing Stores
retain their contracts. CallOptions select deterministic fuel and cancellation.
StoreConfig bounds frames, slots, memories, tables, heap objects and roots.
Failed instantiation reclaims unescaped state while preserving externally
visible effects on imported objects.

For repeated numeric calls, `function.Scalar()` binds an I32/I64/F32/F64-only
signature once and reuses private argument storage. Supply a caller-owned stack
with length exactly max(parameter count, result count):

```go
scalar, err := function.Scalar()
// Check err. This example assumes (i32, i32) -> i32.
stack := []uint64{17, 64}
fuel, err := scalar.Call(ctx, stack, vm.CallOptions{})
// Check err before consuming stack[0]; fuel is separate policy metadata.
```

Parameters occupy the prefix on entry; only the result prefix is replaced on
success. I32/F32 use low32 bits; I64/F64 preserve all bits. Calls share the
ordinary Store execution/fuel/cancellation rules. A binding or Store cannot run
two calls concurrently; use Caller.Call for controlled host reentry. References
and V128 use typed Values. `Store.NewScalarFunction` defines an independent raw
numeric host callback; its callback stack is separate from this caller buffer.

## Shared linear memory and modern core

Enable `wasm.Threads`, create shared memory with an explicit maximum, and use
`Extern.ShareMemory()` / `otherStore.ImportMemory(cap)` within one Engine.
Closing the capability stops future imports; existing Store bindings remain
valid. Each Store reserves the backing's effective maximum against its page
quota. MaxWaiters bounds the Engine's waiting resources. See
[Core Threads](design/core-threads.md) for synchronization and lifetime rules.

WasmGC references follow Store roots and quota-limited tracing; they are not
shared through a memory capability. `CollectEveryAllocation` enables stress
collection. [Final EH](design/exception-handling.md) uses tags and separate
exception references; escaping host exceptions retain their original identity.
These guest features remain distinct from Go/LLGo's own host GC.

## WASI Preview 1

`pkg/wasi/preview1.Process` owns host resources independently of a Store.
Configure explicit args/env, preopens and streams, then install it into a
Linker before instantiation:

```go
config := preview1.DefaultConfig()
config.Args = []string{"program", "argument"}
config.Env = []string{"MODE=test"}
config.Preopens = []preview1.Preopen{
    {GuestPath: "/work", HostPath: "/tmp/work", ReadOnly: false},
}
process, err := preview1.NewProcess(config)
// Check err, then defer process.Close and wait for cleanup before releasing
// host resources. Create store/linker with the existing Engine.
err = process.Install(store, linker)
// Check err, instantiate, and invoke the guest's _start export.
```

Defaults provide clocks/random, empty stdin and discarded output. There is no
ambient environment, filesystem or networking. BytesInput and bounded Buffer
provide deterministic streams. FileStream takes ownership of an OS file and
supports cancellable I/O; custom streams must observe context cancellation.
ImportSocket takes an explicitly authorized socket and listener identity.

Derive guest-call contexts from `process.Context()` to include pure Wasm
execution in process cancellation. Use an independent context for waiting on
cleanup. One Process can serve several isolated Stores. Store closure cancels only that
Store's calls. Process closure or proc_exit cancels the shared resource domain;
ExitError carries the guest status separately from errno/trap/cancellation.
Detailed [P1 ownership, limits and host behavior](design/wasi-preview1.md) apply.

For legacy threaded commands, create `threads.New(ctx, engine, compiled,
process, threads.DefaultConfig())`, then call `group.Run()`. Successful
construction transfers Process ownership; Engine and CompiledModule remain
borrowed. The same compiled module can serve successive Groups. Keep both open
until workers have joined. `group.Close()` requests cancellation;
`group.WaitClosed(cleanupContext)` joins Stores and Process resources. The
per-thread Store/fuel and process thread quotas are explicit in Config.
See [thread group ownership](design/wasi-threads.md). Component/WASI 0.2/0.3
embedding follows below.


## Component host modules

Use `pkg/component.Engine` to compile both the component and its imported Core
modules; the component Engine owns a private Core domain. After checking each
error, the binding sequence is:

```go
host, err := engine.CompileCore(ctx, coreBytes)
// Check err; defer host.Close().
compiled, err := engine.Compile(ctx, componentBytes)
// Check err; defer compiled.Close().
instance, err := compiled.Instantiate(ctx, component.Imports{
    CoreModules: map[string]*component.CoreModule{"math": host},
    Functions: map[string]component.HostFunction{
        "mark": func(*component.Caller, []component.Value) (*component.Value, error) {
            return nil, nil
        },
    },
}, component.CallOptions{})
// Check err; defer instance.Close().
```

Import names and signatures come from the component, including nested
`Imports.Instances`. All bindings are checked before guest start effects.
Closing host prevents new bindings; existing bound Instances keep their module
until close. `host.Info()` supplies an owned Core metadata snapshot. See
[component lifetime and matching](design/component-runtime.md#host-core-module-imports).


## WASI pollable ownership

Create an explicit provider domain and bind the complete clock/poll interfaces:

```go
poller, err := preview2.NewPoller(preview2.PollConfig{})
// Check err; defer poller.Close().
clock, err := preview2.NewMonotonicClock(poller)
// Check err, then use clock.Imports() when instantiating the component.
// Close/wait for the Instance before the deferred provider Close.
```

The returned Imports includes poll and monotonic-clock with one nominal pollable
type. Merge its versioned Instances with other explicit providers as needed.
An embedder can create an additional pollable in a host callback through
`poller.NewPollable(caller, source)`. PollSource returns a nonblocking coherent
readiness/change-channel snapshot and remains externally owned. Provider Close
cancels pending polling, reclaims abandoned timers and joins timer callbacks;
Store close cancels only its calls/handles. See [the full provider contract](design/wasi-preview2.md#pollables-and-monotonic-clock).


## WASI IO

Configure explicit sources and bind `process.Imports()` through `pkg/component`:

```go
output := preview2.NewOutputBuffer(1 << 20)
process, err := preview2.NewIO(preview2.IOConfig{
    Stdin: preview2.NewBytesInput([]byte("input")),
    Stdout: output,
})
// Check err. Successful construction transfers source ownership.
// Instantiate with process.Imports(); close/wait for the Instance on teardown.
// Then process.Close() and process.WaitClosed(cleanupContext).
```

Drain output concurrently when using a bounded sink that can fill. An embedder
can implement InputSource/OutputSource for authorized capabilities; prompt
methods, readiness and cancellation/close behavior are required contracts.
`NewInput`/`NewOutput` mint owned stream capabilities from active callbacks for an
additional source. Provider creation/cleanup is independent of Store close.
The IO owner also owns its borrowed Poller; use that same Poller when constructing
a monotonic clock. Terminal presence is explicit, with none as the default.
See [IO ownership and limits](design/wasi-preview2.md#io-stdio-and-terminals).

## WASI filesystem

Filesystem grants are explicit and share the IO provider's nominal resources:

```go
process, err := preview2.NewIO(preview2.IOConfig{})
// Check err.
filesystem, err := preview2.NewFileSystem(process, preview2.FileSystemConfig{
    Preopens: []preview2.FileSystemPreopen{{
        GuestPath: "/data", HostPath: configuredDirectory, ReadOnly: true,
    }},
})
// Check err, then instantiate with filesystem.Imports().
// Merge a configured WallClock provider if the selected world imports it.
// On teardown: close/wait for the Instance, close both providers, and call
// filesystem.WaitClosed(cleanupContext) and process.WaitClosed(cleanupContext).
```

Preopens are available only on Linux/Darwin. Stream cursors are independent and
retain the open object after descriptor drop. Filesystem Close revokes streams
and cancels their readiness waiters; IO Close separately releases stream/error
resources. See [permissions, limits and lifecycle](design/wasi-preview2.md#filesystem-and-preopens).

## WASI IP name lookup

```go
network, err := preview2.NewNetwork(process, preview2.NetworkConfig{
    AllowNameLookup: true,
})
// Check err, then use network.Imports(). Name lookup is denied by default.
// Close/wait for the Instance; close both network and process, then wait for
// network.WaitClosed(cleanupContext) and process.WaitClosed(cleanupContext).
```

The default resolver uses Go DNS/host configuration only after this explicit
grant. Supply `Resolve` to control name service policy and results; callbacks
must honor context cancellation. Resolution streams and their child pollables
have checked nominal identity and bounded ownership. TCP/UDP socket access is
still separate. See [network ownership and limits](design/wasi-preview2.md#network-capabilities-and-ip-name-lookup).

### Suspended host calls

Inside one `Store.WithCaller` session, register a `SuspendableHostFunction`,
start a guest with `Caller.Start`, and drive its `Continuation` with `Caller.Poll`.
An incomplete poll preserves explicit guest frames; it does not preserve the
callback's Caller or host stack. `HostSuspension.Ready` must be nonblocking and
must not enter the VM. Its Resume receives a new Caller. All continuations share
session fuel/frame/slot bounds, and `StoreConfig.MaxContinuations` bounds their
count. Abort releases a parked continuation without executing guest code;
session exit aborts abandoned continuations. Ordinary blocking Call rejects a
host suspension. The component scheduler uses this API for native async.

For component host imports that complete later, use `Imports.AsyncCompletions`.
An `AsyncHostFunction` returns a `component.NewHostCompletion()` and may pass its
Resolve method to external work. Resolve accepts a resolver function which will
run in the Store session with a fresh component Caller. Create resources and
perform guest reentry in that resolver. The starter Caller expires immediately;
its Context is safe to retain and receives cooperative cancellation. Call
Cancelled to acknowledge cancellation without returning a value. Each completion
belongs to one invocation and may publish once. Immediate async implementations
may continue to use `Imports.AsyncFunctions`; a name cannot bind both APIs.


### Host futures, streams and background work

Use `caller.NewSource(endpointType, component.AsyncSource{Read: read, Close: close})`
to return a typed future/stream from an import. Read receives a fresh Caller and
a maximum count; return `NewReadCompletion()` and resolve it from external work.
Its resolver returns `ReadResult` with typed Values, packed Bytes or a unit Count.
Set Closed for EOF. A future supplies exactly one element. Retain only the
starter's Context, immutable data and completion, never its Caller.

`caller.ReadAsync(resource, limit)` consumes a bounded batch and drives other
component tasks. Outside a callback use `instance.ReadAsync(ctx, resource,
limit, options)`. Drop the readable capability with DropAsync when finished.
To consume a guest result before its producer exits, use
`function.CallWithResult(ctx, args, options, consume)` and read inside consume.
Its Caller expires on return; the call subsequently drains remaining tasks.

Use `caller.Schedule(component.AsyncWork{Owned: refs, Start: start, Done: done,
Cancel: signal})` when host consumption must progress independently of a result
future being read. Owned capabilities move immediately. Start returns a void
HostCompletion; Done receives a fresh Caller and the cancellation outcome.
NewCancellation creates an optional one-job signal safe to cancel from another
goroutine. WithCaller supplies a scoped embedding session for scheduling work.

Close callbacks must be nonblocking and must not enter or wait for the VM. They
run once on EOF/drop/failure/shutdown; cancellation contexts terminate workers.
Close the Instance and wait for WaitClosed before releasing external backing
storage. Unpublished payloads and scheduled ownership are reclaimed on failure.


### WASI 0.3 providers

`pkg/wasi/preview3` supplies versioned Environment, Random, SystemClock,
MonotonicClock and IO providers. Bind each provider's Imports explicitly. P3 IO
reuses the prompt InputSource/OutputSource capability contracts, exposes native
streams/futures and owns accepted sources; it never inherits ambient stdio.
Close/wait for bound Instances, then Close and WaitClosed the IO/monotonic owners.
Use preview3.RunCommand for the checked async command export.

On Linux/Darwin, NewFileSystem(ioOwner, FileSystemConfig{Preopens: grants}) adds
all versioned types/preopens operations. FileSystemPreopen explicitly supplies
GuestPath, HostPath and optional ReadOnly. Streams keep descriptor leases through
EOF/drop; directory streams use independent cursors. MaxDescriptors bounds both
handles and physical objects, while worker/path/entry/batch limits bound work.
Close/wait for bound Instances, then Close/WaitClosed FileSystem before IO. No
ambient filesystem is granted.

NewSockets(ioOwner, SocketsConfig) adds the complete P3 socket/name-lookup surface.
Grant bind/connect and UDP peer send/receive explicitly with policy callbacks;
nil callbacks deny. Enable external DNS separately or supply a cancellable
resolver. TCP listen yields independently owned socket resources; send and receive
use native streams/futures and support half-close. Socket/worker/name/address
limits and the IO owner's stream/batch limits apply across bound Stores. Close
and wait for Instances, then Sockets, then IO. HTTP uses the same Sockets/IO owners.
See [provider scope and lifecycle](design/wasi-preview3.md).


Use component.NewHostCompletionWithCleanup when external work may create an
owned result before a resolver can publish it. Cleanup runs once after acceptance,
including abandonment before resolution. It must be nonblocking, must not enter
the VM, and must synchronize with the external producer. Keep an independent
result lease until fresh-Caller publication succeeds; release late results when
the invocation has already closed. Startup failure before acceptance remains the
creator's responsibility.


Use component.NewReadCompletionWithCleanup when a source read can obtain external
owned objects before its resolver publishes a batch. Cleanup follows the same
nonblocking, exactly-once contract as HostCompletion, including late cancellation
and abandonment; keep independent object leases until publication completes.


Use `preview3.NewHTTP(sockets, HTTPConfig)` for the pinned 0.3.2 types, client
and handler imports. Grant URL requests and resolved connections independently;
nil grants deny. `Handler` explicitly supplies an upstream Go handler without
inheriting network authority. Headers/options are immutable after message
construction. Body streams and trailer/transmission futures retain bounded
ownership separately from message handles. Closing a transmission future alone
does not cancel a live message. Outbound HTTP/1.1 supports late trailer names
without buffering the complete body, including over TLS.

`HTTP.CallIncoming` invokes the async guest handler. Read and close the returned
response Body; Close also joins the invocation. `preview3.HTTPHandler` accepts a
factory supplying a fresh instance and request-scoped providers, and streams the
response while the guest continues producing it. Factory cleanup closes/joins
these owners and the instance. Close HTTP before Sockets and IO.

`component.AsyncSource.Owned` moves readable/resource capabilities into a source;
`ReadOwned` receives their current references. EOF/drop/failure/shutdown reclaims
remaining ownership inside the serialized Store. `NewOwnedSource` also returns
the moved references for a private ownership anchor. Moving one onward invalidates
the anchor's copy. `AsyncSource.Cancel` is a one-source cancellation signal and
closes its writer through the same scheduler. `AsyncWork.Cleanup` runs once after
accepted work, including abandonment before Start, without entering the VM.

Use `Caller.AwaitRead` in an asynchronous starter when sibling tasks must progress
independently. Return its HostCompletion; the consumer receives a fresh Caller
and owns its payload. Keep the endpoint alive with scheduled ownership until
completion. An operation context can cancel this read without cancelling its
parent session. `ReadAsyncContext` provides the synchronous embedding form;
it drives the scheduler while waiting and must not substitute for independent
continuations in mutually dependent operations.


For the complete pinned P3 CLI/service/middleware worlds on Linux/Darwin, use
`preview3.NewWorld(preview3.WorldConfig{...})` and pass `world.Imports()` to
component instantiation. The config embeds the existing explicit provider
configs, with one shared IO domain. `world.HTTP().CallIncoming` invokes a guest
service; `preview3.RunCommand` invokes a command. Close/join bound instances,
then `world.Close()` and `world.WaitClosed(ctx)`. Later constructor failure
closes/joins accepted IO sources; pre-IO validation failure leaves them untouched.
World creation grants no ambient args, environment, filesystem or network.

Guest-defined i64 resource representations also support local borrow calls.
Pass the opaque Value returned by an own-resource result as the borrow argument;
the runtime supplies the full i64 representation only to its defining component.
Foreign components receive i32 handles, including with memory64. Borrowed resources
remain live through suspension and cannot be moved/dropped until completion or
cancellation cleanup. No additional host configuration is required.

`component.Instance.StartFuelRemaining()` returns the immutable instantiation
fuel receipt. When initialization and the next invocation share a metered
budget, use that receipt as the next CallOptions.Fuel. Later calls leave the
receipt unchanged. Every call still has an explicit caller-supplied budget.
