# Embedding

`pkg/vm` provides Engine, CompiledModule, Store, Linker, Instance, Extern and
Execution. See the executable [VM examples](../pkg/vm/example_test.go).

An Engine validates and compiles immutable modules. Each Store serializes its
own guest execution and mutable objects; distinct Stores can execute in
parallel. Select `EngineConfig.Backend` (`interpreter` or `native`) explicitly.
`Profile: "core3"` enables the complete final core profile. Core Threads remains
an independent feature bit. Native execution has no interpreter fallback.

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
execution remains subsequent scope.


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
