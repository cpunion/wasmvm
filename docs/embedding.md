# Embedding the Core 2.0 runtime

`pkg/vm` implements the Core 2.0 binary profile with the project-owned interpreter
and an experimental native backend. Both use the same frame, host-call, memory,
root and lifecycle contracts. Select native explicitly:

```go
config := vm.DefaultEngineConfig()
config.Backend = "native"
engine, err := vm.NewEngineWithConfig(config)
```

Native execution requires the platform adapter on linux/amd64 or darwin/arm64;
standard Go requires `CGO_ENABLED=1`, while LLGo uses its C interop adapter.
Unsupported platforms and unknown profiles fail without backend fallback.
Compilation prepares both unmetered and precise-fuel code within the aggregate
`Native.MaxCodeBytes` limit. Calls and SIMD/state helpers still cross explicit
runtime boundaries; the complete semantic corpus does not establish competitive
performance for those paths. See the [runtime scope](../benchmarks/native/runtime-plan.md).
Core 3.0, Threads, components and WASI are still unimplemented.

The [package examples](../pkg/vm/example_test.go) are executable tests under Go
and LLGo. They demonstrate compiling and calling an exported add function, and
linking an explicitly typed host function into a guest module.

```go
engine := vm.NewEngine()
defer engine.Close()
compiled, err := engine.Compile(ctx, wasmBytes)
if err != nil { return err }
defer compiled.Close()
store, err := engine.NewStore(vm.DefaultStoreConfig())
if err != nil { return err }
defer store.Close()
instance, err := vm.NewLinker(store).Instantiate(ctx, compiled, vm.CallOptions{})
if err != nil { return err }
add, err := instance.Export("add")
if err != nil { return err }
result, err := add.Call(ctx, []vm.Value{vm.I32(17), vm.I32(25)}, vm.CallOptions{
    Metered: true, Fuel: 3,
})
if err != nil { return err }
bits, _, err := result.Values[0].Bits()
if err != nil { return err }
fmt.Println(uint32(bits)) // 42
```

## Ownership and shutdown

An Engine supports concurrent compilation. Compilation snapshots the input;
the caller must not concurrently modify it while the snapshot is taken.
`Engine.Validate` decodes and statically validates without lowering; `Inspect` adds copied import/export signatures and object limits without linking or executing.

CompiledModule owns immutable code and metadata, can instantiate in multiple
Stores of its Engine, and returns copied metadata from `Info`. Descriptors can
be shared among entries within one snapshot; copy a descriptor before editing
it independently. Snapshots never alias engine code or other snapshots. Each
referenced function signature is copied only once, keeping metadata storage
linear in the decoded module rather than multiplying signature size by the
number of functions. `Inspect` follows the same ownership rules. Closing the
Engine prevents new compilation and Store creation. Existing compiled modules,
Stores and calls remain usable. Closing a CompiledModule prevents new
instantiation through that handle; existing instances retain their code.

A Store is one isolation domain and allows one executing operation at a time.
Concurrent calls or mutable object accesses return `vm.ErrBusy`. Independent
Stores can run concurrently. Host callbacks use `Caller.Call` or `CallRef` to
reenter the current Store. A `vm.Execution` binds a function and per-run options;
`Run` can be repeated but cannot overlap. Its `Close` cancels that Execution's
active call without closing the Store. The initial public call driver allocates
continuations per call; internal warmed-dispatch allocation measurements do not
describe public API allocation costs.

`Store.Close` is an idempotent, nonblocking cancellation request, safe inside a
callback. Backing storage and host references are released after all active
calls and callbacks leave. External code can use `WaitClosed(ctx)` to await
that point. A callback must not wait for its own Store to close. User callbacks
must cooperate with `Caller.Context()` cancellation; the VM cannot preempt an
arbitrary blocked Go callback. Host panic propagation performs deferred cleanup
of frames, callback capabilities, temporary roots and the Store execution lease.

Successfully published instances and host-created memories/tables/globals live
until Store close. Failed instances are reclaimed when unreachable; references
that escaped into imported state or explicit host roots preserve their owners.
Imported-state effects are not rolled back after a later initialization/start
failure. Releasing the last root allows a subsequent `Collect` to reclaim a
failed instance, including an unreachable cycle. There is no individual
successful-instance unload operation in this version.

## Values and host references

`wasm.ValueType` identifies types. `vm.Value` contains exact scalar/vector bits
or an opaque reference. Constructors are `I32`, `I64`, `F32`, `F64`, `F32Bits`,
`F64Bits`, `V128` and `NullRef`. Integer constructors take unsigned bits so
signedness is determined by the Wasm instruction. `Bits` does not expose
reference identities. A zero Value is invalid rather than an implicit i32.

`Store.NewExternRef(object)` creates a persistent root. `Store.ExternRef(value)`
reads the Go object; the object table contains normal Go references, never
integer-encoded Go pointers. Passing a nil Go object still creates a non-null
externref; use `NullRef(wasm.ExternRef)` for a null reference. Function exports
provide `Reference()`, and `Store.CallRef` invokes a rooted function reference.

Values copied in Go share one root token. `Store.Root(value)` creates a separate
root. `Value.Release()` invalidates all copies of its token, is idempotent, and
does not invalidate independent roots or references already held by a guest.
Public call results and table/global reads return caller-owned roots. Release
them when no longer needed. Unreleased roots consume the host-root quota and
keep their targets alive. References from another Store or a released token
return `vm.ErrReference`; internal function identity reuse cannot revive them.
Externref identities increase monotonically and reject overflow rather than
wrapping. Root operations and object-table access are synchronized with GC.

Host callback arguments have temporary roots valid through the callback. Use
`caller.Store().Root(arg)` to retain one. `Caller.NewExternRef` creates a
callback-scoped root that remains valid through result handoff, then is released
automatically. This is suitable for returning a newly created externref. Roots
created through `Store.NewExternRef` remain the host's responsibility, including
when used as callback results. Do not release a result's root before the
callback hands it back to the VM.

`Release` removes a root; it does not promise immediate Go-object collection.
Use `Store.Collect` between calls or `Caller.Collect` at a host suspension point.
Collection visits running/suspended guest slots, pending host arguments,
globals, tables, passive elements, active instance owners and explicit roots.
Store close releases all remaining roots and objects. Go finalizer timing is
not part of the contract. This Core 2.0 reference management is not WasmGC.

## Host imports and memory

`Store.NewFunction(FunctionType, HostFunction)` uses explicit parameter/result
types, with no reflection inference. `Linker.Define(module, name, extern)` binds
functions, tables, memories and globals. Duplicate bindings, missing imports,
type mismatches and foreign Store objects are rejected. Callback errors are
propagated unchanged; invalid callback result values return `vm.ErrType` or
`ErrReference` without resuming the guest with unchecked data.

Extern methods provide copied memory reads/writes, checked growth, typed
table/global access and function calls. No raw memory slice escapes. During a
callback, `Caller.ReadMemory` and `WriteMemory` access the calling guest's memory
by index. A directly called host function has no guest memory. Caller operations
are serialized: concurrent capability use returns `ErrBusy`, and using a saved
Caller after callback completion returns `wasm.ErrClosed`. Callbacks must join
their own capability operations before returning; the driver drains an
in-flight operation before resuming its suspended guest.

## Limits, cancellation and errors

Start with `DefaultEngineConfig` and `DefaultStoreConfig`, then explicitly
change required quotas. Zero fields in a supplied configuration mean zero,
not an unspecified default. Compiler phases have independent byte/node/depth/
slot/work limits. Store limits include instances, function identities, types,
globals, memory/table object counts, pages/elements, passive segment storage,
external objects and host roots. Canonical signatures retain a bounded
Store-lifetime cache. Logical storage quotas do not promise a process-RSS
ceiling during host allocation/copying or deferred Go collection.

Guest frame and slot limits cover all suspended and nested continuations of
one operation. `MaxReentry` limits additional host-mediated nested calls; zero
permits the outer call and rejects reentry. Fuel v1 is shared across those
continuations, charges before Wasm instruction effects, and is reported as
`CallResult.FuelRemaining`. Every new top-level call resets its supplied fuel.
Instantiation's start call receives the supplied options. The immutable `Instance.StartFuelRemaining()` receipt can carry its remaining budget into a later call; segment application
is initialization work rather than instruction fuel.

Context cancellation is checked between bounded compiler/initialization phases,
at guest entry/calls/backedges, at most every 64 straight-line instructions and
between bounded bulk-operation chunks. Large Go allocations or arbitrary host
callbacks do not have a hard preemption guarantee. Cancellation during bulk
execution may leave partial effects. No default WASI or ambient I/O capability
is installed.

Use `errors.Is` for `wasm.ErrMalformed`, `ErrInvalid`, `ErrUnsupported`,
`ErrLimit`, `ErrTrap`, `ErrClosed`, and the `vm` link/type/reference/busy errors.
`*wasm.Trap` includes the trap code, function index and original instruction
byte offset when execution supplies them. Context cancellation and user callback
errors remain distinct from Wasm traps.

M1 remains open for own-engine benchmark integration and final
acceptance. This API increment does not claim native execution, WASI support,
controlled performance-gate acceptance or production security review.

## Individually enabled extensions

The default engine validates complete Core 2.0. To enable Extended Constant
Expressions on either backend:

```go
config := vm.DefaultEngineConfig()
config.Features = wasm.ExtendedConst
engine, err := vm.NewEngineWithConfig(config)
```

This permits i32/i64 add, sub and mul in global initializers and active segment
offsets. Validation applies the configured operand-stack limit. Imported
immutable globals resolve separately for each instance; mutable or
module-defined global reads remain invalid. `Inspect` and `CompiledModule.Info`
report the enabled feature mask. Unknown feature bits return `ErrUnsupported`.
The complete `core3` profile is not yet implemented.

Enable proper direct and indirect tail calls with `config.Features |= wasm.TailCall`.
Tail calls can cross modules and have different parameter lists with matching
results. They reuse the current guest frame, while ordinary calls retain their
normal frame requirement. Both backends support metered execution and bounded
cancellation; host callbacks retain their originating module capability after
the tail frame is retired. Combining feature bits enables only those extensions,
not the complete Core 3.0 profile.

Use `config.Features |= wasm.MultiMemory` to enable multiple private memories.
All scalar/SIMD and bulk operations address the selected module memory; exports
and host `Caller` memory indices follow the same import-then-definition order.
Aliased imports share growth and contents. Store quotas count underlying
allocations, while out-of-bounds bulk operations leave the destination unchanged.
This extension does not enable Memory64 or shared-memory Threads.

Relaxed SIMD is enabled independently with `wasm.RelaxedSIMD`. All 20
instructions work in both backends. The implementation chooses strict swizzle,
saturating conversions, bit selection, strict min/max and Q15, unfused
multiply/add, and signed byte dot products. The same permitted projection is
used on amd64/arm64 and every CPU feature mask. The four-byte dot widens before
accumulation; the two-byte dot wraps only its final i16 result. This does not
enable the complete Core 3.0 profile.

Enable both Memory64 and Table64 with `wasm.Memory64`. Set `MemoryType.Index64`
or `TableType.Limits.Index64` for host-created objects; their address widths
must match imports. `Min`/`Max`, host memory growth/page counts, and host table
indices/growth/size use `uint64`. Existing 32-bit guest instructions keep i32
operands/results; 64-bit objects use i64, including the complete `-1` growth
failure value. Store quotas remain explicit independent caps and no oversized
allocation or growth can wrap those caps. Multi-Memory still needs its own flag.
Memory64/Table64 do not enable WasmGC or the complete Core 3.0 profile.
