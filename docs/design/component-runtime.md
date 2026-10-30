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
nominal host resource types bind through `Imports.Resources`. Host Core modules
bind through `Imports.CoreModules`; host components remain unsupported. Other
incomplete executable operations are rejected during
compilation or host binding, before instantiation invokes guest code.

Canonical lift/lower shares `internal/component/canonical` for scalars,
strings, lists/fixed lists, maps, records/tuples, variants, flags/enums and option/results.
Memory32/64, all three string encodings, direct/spilled arguments/results,
realloc and post-return use checked memory capabilities. Resource new/rep/drop
executes through the same Core session. Own/borrow call values now use opaque capabilities and scoped transfers. Async
execution, standalone resource values/starts and local cm64 borrows remain required;
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
post-return runs. Resource transfers use per-call ownership cleanup and borrowed lifetimes; async
cleanup remains unimplemented.

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


## Collections

Fixed lists use their declared element count and inline layout; flat conversion
repeats the element shape only within the existing flat-slot bound. Larger
values use the same function-level spill contracts. Maps use the specified
list-of-tuples representation and preserve input order and duplicate keys.
A map layout caches one synthetic entry tuple, preventing metadata growth on
repeated transfers. Existing list bounds, allocator/memory capabilities and
value/byte budgets apply to map entries; no separate execution path is added.


## Runtime resource declarations

Each concrete declaration creates a nominal runtime identity, separate from its
compile-time resource ID. Type imports, aliases, synthesized interfaces and
child instance exports bind checked IDs to those identities. Instantiating a
child twice creates distinct declarations and independent handle domains;
importing a resource and exporting it again retains its identity and authority.
Only the defining component may create or inspect its representation, as
already enforced by static validation. Representations preserve i32/i64 width;
handles remain i32 indexes in the pinned 28-bit domain and are never reused.

`canon resource.new`, `resource.rep` and `resource.drop` execute on those
instance tables. Drop consumes the handle before invoking the defining Core
destructor with the active Caller, preserving the same fuel/context/reentry
limits. Destructor failure cannot restore ownership and locks down the affected
component. New/drop respect canonical leave guards; rep remains available
during post-return. `Config.MaxResources` bounds live handles across every
component domain in the Store, rather than multiplying the quota per child.

Failure (including host panic during initialization) and instance close seal
the tables and release their quota. Close is a
cancellation request and does not execute implicit guest destructors; explicit
resource.drop runs them before teardown. Own/borrow calls now publish opaque capabilities and connect host resource
imports to flat/memory codecs. Standalone resource values/starts and the local
cm64 borrow ABI remain rejected; these call contracts do not complete M5/M6.


## Ownership across calls

`Value.Resource` carries an opaque `Resource`; numeric Bits and zero tokens do
not grant ownership. Copies of a token share one move state. Passing own moves
that capability, invalidating every old alias; borrow pins its lender and expires
at call return. Aggregate values use the same codec, including maps, fixed lists,
variants and spilled arguments/results. Duplicate own arguments are rejected
before guest allocation. Foreign borrowed handles must be explicitly dropped by
the guest before normal return; escape invalidates them and traps the component.

A host resource declaration is created by `NewHostResourceType(destructor)` and
bound by name through `Imports.Resources`, including nested interfaces. Copies
share the same immutable nominal declaration. Equality-bound imports must bind
the same declaration, checked before guest effects. Each Store retains its own
runtime binding; capabilities cannot cross Stores. `Instance.NewResource` and
`Caller.NewResource` mint host ownership for a bound type. `ResourceRep` checks
the host declaration and capability before exposing its u32 representation.
Guest-defined resource representations remain opaque to the host. CM64 own
values preserve their full representation through transfer and destruction.

Drop host ownership with `Instance.DropResource` or callback-scoped
`Caller.DropResource`. Destructors consume ownership before user/guest code and
share the active Core session, fuel and reentry limits. Failed provisional
transfers and callback-created resources run their host destructors even during
cancellation. Guest destructors run only while their implementation can safely
be entered; failed/canceled guest state is not reentered. Cleanup invalidates
unpublished ownership and releases quota even when a destructor fails or panics,
preserving the primary failure. Ordinary published ownership must be explicitly
dropped or transferred. Store close seals tables without synthesizing guest drops.

`vm.Store.WithCallerCleanup` handles final publication failure before the Caller
expires, retaining one operation lease and fuel/context. Synchronous component
entry is exclusive: a suspended component cannot be reentered through a host
callback. Cross-component calls and local Core destructors retain their intended
paths. Pure host callbacks/destructors additionally share the host depth bound.

The pinned MVP reference specifies u32 resources. Local cm64 borrow conversion
is rejected before instantiation rather than truncating its representation.
Standalone resource value imports/exports and resource-bearing component starts
remain a separate integration task; no support is inferred from call values.


## Checked host function imports

`Imports.FunctionChecks` optionally validates provider signatures before any
Core/component start runs. Raw `Imports.Functions` still use guest-declared
schemas; a standard interface provider supplies checks for its own contract.
All imported resource declarations bind before checks run, including nested
interfaces. A verifier can distinguish own/borrow and compare nominal host
identities with `ValueType.IsResource`, rather than trusting equal numeric
representations. Missing imports or failed checks publish no instance.

`FunctionInfo` and `ValueType` are read-only views of the existing checked type
graph. Parameter/field/case/label access is bounded; no mutable model/registry
slices are exposed. Retained views keep an immutable nominal binding map after
initialization, avoiding one expanded type copy per function. Checks consume no
guest fuel and honor cancellation between verifiers. A host verifier, like a
host callback, must cooperate if it performs external work.


## Packed byte sequences

`list<u8>` and `fixed-list<u8,N>` use immutable raw bytes in `Value.Text` with
nil `Elements`. `component.ByteList([]byte)` owns a copy; `Value.ByteString()`
reads either packed or legacy element-wise inputs. Raw bytes may contain NUL,
invalid UTF-8 and every byte value. Reusing the existing type-directed storage
keeps each generic Value at 64 bytes on the accepted 64-bit platforms. All
other aggregate/resource layouts remain unchanged.

Guest lifts and all owned clones produce packed byte values. Legacy element-wise
inputs remain supported and fully checked, including rejection of values above 255; mixed
nonempty Text and nonnil Elements are rejected. This changes the public result
representation of byte sequences: consumers should use ByteString rather than
assuming one Value per byte. Nested lists, fixed-lists, maps and variants use the
same codec and lifetime rules. Reference-oracle bytes and expected values stay
unchanged; test comparison expands packed results only at byte-sequence nodes.

Bulk memory copies retain pointer-width/range checks, MaxValues logical element
counts and nesting limits. MaxBytes charges actual scratch/output bytes rather
than a generic Value slot per packed byte. Immutable clones charge retained
bytes, and results never borrow guest memory. Cancellation is checked around
bulk copies; realloc failure, invalid ranges and transfer budgets remain errors.
The representation serves common WASI IO/random payloads, not a workload-specific
Core execution path or a separate engine.

Owned clone normalization also applies to imported values and literal bounds:
equal bytes compare equal across packed and element-wise host inputs, including
nested aggregates and aliases. Normalization occurs before value identity checks
and guest initialization. Legacy conversion charges its scratch and immutable
output bytes; packed clones charge retained bytes. Byte value fixtures exercise
this without weakening comparison of unequal data.


## Host Core module imports

`Engine.CompileCore` creates a `CoreModule` in the component Engine's existing
Core domain, using its selected interpreter/native backend, feature profile and
bounded compiler. Compilation performs no instantiation or start effects.
Bind the module through `Imports.CoreModules`, including nested `Instances`.
Cross-Engine, closed, missing or structurally incompatible bindings fail before
any Core/component start. Compiling a module with the standalone `vm.Engine`
does not confer access to the private component Engine's domain.

Matching reuses the existing Core module relation: exports are covariant,
imports contravariant. Function, table, memory, global and tag descriptors retain
canonical recursive/GC type identity, limits, mutability and pointer widths.
An immutable type-only module view retains no code bodies, data, initializers
or native publications. Each instantiation forks the compiled type registry
once, interns each actual schema once and memoizes successful type pairs.
Work is bounded by the component validation node limit; a module with too much
metadata fails binding even when only a small export is requested. Shared
compiled plans and canonical definitions remain immutable across Stores.

A `CoreModule` owns its compilation. Close prevents new bindings immediately;
already acquired bindings retain one deduplicated lease per host module until
their Instance closes or initialization fails. This includes closing the owner
from an earlier Core start, before a later module instantiation. Instantiated
Core objects retain their normal Store/native code leases. Closing a component
Engine or compiled component does not retire existing Store execution.
`CoreModule.Info` returns owned metadata and rejects a closed owner.

This capability does not bind host components or abstract host type imports;
standalone resource values/starts, local cm64 borrows, async plans and complete
WASI worlds remain separate required work.
