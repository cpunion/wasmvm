# Component execution

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
bind through `Imports.CoreModules`. Non-resource type imports use their checked
equality targets without runtime identities. Host components bind through
`Imports.Components` with same-Engine code leases. Other
incomplete executable operations are rejected during
compilation or host binding, before instantiation invokes guest code.

Canonical lift/lower shares `internal/component/canonical` for scalars,
strings, lists/fixed lists, maps, records/tuples, variants, flags/enums and option/results.
Memory32/64, all three string encodings, direct/spilled arguments/results,
realloc and post-return use checked memory capabilities. Resource new/rep/drop
executes through the same Core session. Own/borrow call values now use opaque capabilities and scoped transfers. Async
execution and local cm64 borrows remain required;
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
post-return runs. Resource transfers use per-call ownership cleanup and borrowed lifetimes; pending async transactions rebind cleanup to the current Store-session Caller.

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
imports to flat/memory codecs. Standalone resource values and starts now use
these same transfers. The local cm64 borrow ABI remains rejected; these call contracts do not complete M5/M6.


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
Standalone resource values and component starts now share the ownership
transaction, with separate authored value-extension fixtures described below.


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

Host component bindings use the same code domain. Local cm64 borrows, async
plans and complete upstream world conformance remain separate required work.


## Standalone resource values and starts

`Imports.ValueFactories` supplies a value with an active `Caller`, after every
static import and `FunctionChecks`/`ValueChecks` check succeeds. A factory can
mint `Caller.NewResource` for an imported nominal type. Its Caller expires at
return. Supplying both a raw value and a factory for one name is a link error;
raw resource capabilities from another Store are rejected before factory effects.
Factories also support ordinary values and retain literal/equality constraints.
The shared initialization codec charges aggregate value/byte budgets across
raw imports, factories, literals and start results.

Aliases, child imports/exports and projected instance views share a runtime
value occurrence while retaining its producer's nominal scope. Component starts
pass those values through the ordinary own/borrow call transaction and retain
their validated results. A resource-bearing `Instance.Value` moves the exported
occurrence exactly once, including empty resource containers; subsequent reads
trap. Resource-free values retain detached, repeatable reads. Raw numeric fields
never grant resource authority.

An imported borrow remains pinned through synchronous initialization, including
nested calls, then expires before publication. Its host-owned lender can be
retained by the factory and explicitly dropped afterwards. Component exports
cannot transitively contain borrows, as required by the pinned binary grammar;
exporting a borrow *type* remains legal. The value extension remains experimental
in the pinned Component Model source, independently of stable WASIp2 interfaces.

A later start error, host panic, cancellation, aggregate-copy failure or failed
final publication reclaims unpublished owned values while the Core Caller is
still active. Host destructors receive cancellation; guest destructors run only
while their component and context permit reentry. Explicit Store close keeps
its existing cancellation contract and does not synthesize guest destructors.
The implementation shares the existing resource codec and Core session rather
than introducing another execution path. These synchronous capabilities leave
local cm64 borrows and async/WASI 0.3 required.


## Equality-bound type imports

An imported non-resource type is an equality alias to a type that static
validation has already resolved. Function, instance and component type imports
now use the same metadata binding as existing value types, at root and nested
host interfaces. They carry no runtime object to allocate or independently
choose; accepting the declaration preserves its exact equality target. This is
separate from an actual component import, whose code and matching contract are
provided by `Imports.Components` as described below.

Functions and instance/component ascriptions can reference imported aliases.
They retain lexical captures and nominal resource IDs. Fresh abstract resources
still require `Imports.Resources` bindings, and equality-bound resource imports
still require the same host declaration. Provider `FunctionChecks` see the
resolved function shape and its actual nominal resource bindings before effects.
No new type registry or alternate call path is introduced.

## Host component bindings

`Imports.Components` binds independently compiled components from the same
Engine, at root or nested host interfaces. The normal component interpreter
instantiates their original plans and embedded Core code in the consumer Store.
Local nested closures retain their parent's Core map; imported roots retain
their own. Component arguments, narrowed exports, values/starts and nominal
resource transfers use the existing composition and canonical paths.

Each binding acquires a compilation lease. `Compiled.Close` prevents new
bindings, while bindings already acquired can finish initialization and keep
executing. Root instantiation uses the same lease rule. Failure, cancellation,
panic and Instance close release all compilation/Core leases; code retirement
still respects active Store publication. Compiled objects must not be copied.

The matcher copies immutable schemas into one bounded binding registry and
reuses the ordinary subtype/binder relations. Independent source resource IDs,
value equality symbols and canonical Core identities are remapped; source pools
remain unchanged under concurrent binding. Recursive Core groups, GC reference
fields/supertypes and all import/export descriptors retain their meaning.
Repeated bindings share leases and successful schema checks within an operation.
No fallback engine or duplicate execution framework is introduced.

Imported host functions have a provider namespace sharing resolved nominal
bindings. Consumer failure/leave guards belong to the consumer, so a healthy
imported component can still invoke its host provider during failure cleanup.
Guest realloc/post-return/reentry guards remain on canonical entry/leave paths.
This also preserves the identity of a re-exported host function after a
consumer's canonical trap. Global Store shutdown still invalidates all calls.

## Error contexts

The pinned Component Model's `error-context` diagnostic type now executes
synchronously on both own backends. `canon error-context.new`,
`canon error-context.debug-message` and `canon error-context.drop` share the
ordinary canonical option/signature checker, string conversion and executor.
UTF-8, UTF-16, compact latin1/UTF-16 and memory32/memory64 use the existing ABI
layouts. Flat, spilled/aggregate and cross-component values carry diagnostics
without consuming the original occurrence. This does not implement native task,
future, stream or WASI 0.3 world execution.

```go
value, err := component.NewErrorContext("request failed")
// Pass value through an error-context parameter, result or standalone value.
message, err := component.ErrorContextMessage(value)
```

Diagnostics are immutable and copyable, independently of a Store lifetime.
The existing `Value.Resource` opaque-reference slot carries either a nominal
own/borrow capability or an error context, as determined by its checked value
type. `Value` does not grow another union field. Host reference carriers are
snapshotted at clone/lift boundaries; assigning a different carrier in host code
cannot change a guest's retained message or corrupt quota accounting.
`Instance.DropResource` remains an affine-resource operation; guest error-context
drop removes one diagnostic handle and leaves other copies valid.

One tagged handle table prevents resource/context index confusion and rejects
stale or foreign-kind handles. Error contexts share the Store's live resource
quota. `Config.MaxBytes` also caps unique diagnostic-message bytes retained by
guest handles across all namespaces. Repeated handles for one diagnostic share
that byte charge; independent diagnostics consume separate charges. Failed
lowering transactions use the existing rollback queue. Drops and Store close
release retained charges; ordinary resource-only table close stays constant-time.

Default guest creation retains a well-formed message. Setting
`DiscardErrorContextMessages` creates an empty message without lifting or
checking the proposed input range, as permitted by the pinned ABI's discard
policy. Existing messages remain stable when copied/lowered. Diagnostic text
must not determine program behavior. Debug-message output checks the complete
aligned return region before entering realloc. Realloc/post-return retain the
may-leave guard, including for context drop; traps poison the affected guest
entry while host diagnostic values remain readable.

The fixed producer independently validates seven ordinary fixtures. Its cm64
error-context builtin signature remains i32, contrary to the pinned optional
memory64 rule; that rejection is explicit. A standalone-value fixture follows
the project's pinned value-extension grammar and retains the producer's older
syntax rejection. These two fixtures are not independent conformance evidence.
Actual Wasmtime 48.0.1 executes the common guest creation/drop lifecycle and
rejects the forbidden post-return leave. Its cross-component sample traps and
its cm64 sample is rejected; those boundaries remain recorded, not passed.

## Run-to-completion async canonical lifts

The next execution profile admits async component function types, ordinary
synchronous lift/lower of those types, and canonical async lifts. An async
function type describes the component call contract; the async canonical option
selects its Core ABI. The stackful ABI returns through task.return and has no
Core results. The stackless ABI returns a packed EXIT/YIELD code and repeatedly
calls the checked callback with the NONE event. Packed high bits are ignored for
EXIT/YIELD. There are no accepted waitable-set producers in this profile, so
WAIT cannot identify a valid set. Async lower, waitable sets, subtask operations,
cooperative cancellation and future/stream values remain rejected explicitly.

Function.Call runs this profile to completion within the existing Store lease,
fuel, cancellation and reentry limits. Each implicit thread has two zeroed TLS
slots (i32 or the uniform cm64 i64 width). Initialization and canonical realloc
use independent storage; realloc restores its caller's task even on failure.
Stackful nested calls save and restore the parent task without another execution
driver or an unbounded goroutine per callback. YIELD may immediately select the
current ready thread, as permitted by the pinned nondeterministic ABI. The host
may cancel the entire call through its context; that aborts execution rather
than synthesizing a guest TASK_CANCELLED event.

Task.return checks the active task, result schema, encoding and actual memory
instance identity. It uses the existing 16-slot parameter convention to lift
direct or indirect results, including cm32/cm64 strings and nominal ownership.
Foreign borrows must be dropped before result publication, and lenders are
released at task.return. Missing/duplicate returns and invalid callback codes
trap. Later traps, host errors, panic or cancellation poison the affected
namespace and reclaim unpublished results through the original live Core
session. Failed guest owners are not reentered for cleanup.

Imports.AsyncFunctions explicitly binds async host function types; Functions
continues to bind synchronous types. Both use HostFunction and its existing
callback lifetime. FunctionInfo.Async exposes the declared type independently
of the Core ABI. Stable WASI 0.2 command and incoming-HTTP exports must remain
synchronous; their provider imports cannot silently acquire async signatures.
This profile does not provide suspended native async calls or WASI 0.3 worlds.

## Eager async canonical lower

Async lower now uses the same call bridge as synchronous lower, with the pinned
four-flat-slot argument threshold and an indirect result pointer for every
nonempty result. The common layout builder selects these bounds; compound
arguments, strings, memory32/memory64 and nominal resources use the same codecs.
Async function types may use either admitted lift ABI. Every admitted callee
runs to completion; an async lower returns RETURNED (2) without allocating a
subtask handle. The existing policy may immediately resume a ready yielding
callee. This is the eager-resolution path, not pending-subtask scheduling.
Waitable producers, async host suspension, subtask operations, backpressure,
cooperative cancellation and future/stream values remain rejected explicitly.

Lowered results enter the existing ownership transaction before result-memory
checks. A bad result pointer therefore consumes and drops unpublished ownership
once, including the synchronous ABI. Bad spilled arguments fail before host
effects; result-memory failures happen after the callee's effects, as specified.
Loans are returned before resuming guest code; panic, cancellation, invalid
shapes and partial transfer use the shared cleanup queue and logical budgets.
Independent fixtures retain three precise producer rejections: pinned wasm-tools
still gives cm64 async lower an i32 result-pointer parameter instead of i64.
Those fixtures follow the pinned optional memory64 rule and are not independent
producer conformance evidence. Full native async and WASI 0.3 remain open.

### Pending callback tasks and waitables

Async lower of a callback-lifted function may now return a real `STARTING` or
`STARTED` subtask handle. Callback EXIT/YIELD/WAIT turns use the existing Core
session, explicit VM frames and canonical codecs. No Go goroutine, second
interpreter or retained Caller supplies a continuation. Every resumed turn
receives a live session capability; pending result lowering rebinds memory and
resource cleanup to that capability. `Config.MaxTasks` bounds live callback
tasks independently of the shared `MaxResources` handle quota.

Waitable sets, subtasks, resources and error contexts share one tagged,
non-reusing 28-bit handle domain. Intrusive membership and ready queues provide
constant-time join, notification and poll. Progress is coalesced until delivery;
borrow lenders remain pinned until the resolve event is delivered. Both sync
and async subtask cancellation enforce single requests, exclusive waiters,
cooperative TASK_CANCELLED delivery, `task.cancel`, and explicit subtask drop.
Cancellation before entry does not read arguments or transfer ownership.

The blocking embedding call drains its callback continuations within the same
Store lease. Session failure, panic, cancellation and close reclaim pending
transactions and newly allocated async handles. A guest with no ready producer
waits for context cancellation without spinning. Synchronous task suspension is
rejected. Existing run-to-completion stackful callees retain the shared bridge;
arbitrary stackful suspension, asynchronous host completion, explicit
backpressure, futures/streams and WASI 0.3 worlds remain unsupported. This is a
callback execution profile, not complete native async or M6 acceptance.


## Suspended Core calls and asynchronous host completion

All admitted async function types now use one component task scheduler. Callback
and stackful lifts, synchronous lifts of async types, and synchronous/async
lowers share explicit `vm.Continuation` frames. Each Core turn uses Start/Poll;
synchronous lower, wait and cancellation may park at host boundaries. No runtime
goroutine or retained host stack represents guest execution. A separate
`StoreConfig.MaxContinuations` bounds suspended Core calls, including host roots;
all peers share frame/slot quotas, roots, fuel and cancellation. Logical quotas
are recomputed before each turn and host boundary; native physical storage stays
within the Store ceiling. Session exit aborts remaining continuations.

Explicit backpressure uses the pinned 16-bit counter and lazy arguments. A
STARTING subtask reads neither spilled parameters nor result pointers before
entry. Before-entry cancellation consumes no arguments. Exclusive ownership is
held while synchronous lifts/callback bodies execute, then released at callback
YIELD/WAIT/EXIT; stackful async tasks may interleave at suspension points.
`thread.yield` returns zero and does not deliver task cancellation. Cancellation
of a started stackful task is cooperative and may resolve normally. Synchronous
non-async tasks still cannot suspend across a component boundary.

`Imports.AsyncCompletions` binds `AsyncHostFunction` separately from immediate
`AsyncFunctions`; binding both under one name is rejected. The starter returns a
one-shot `HostCompletion` and may retain its immutable argument snapshot and
Context, but its Caller expires. External work calls Resolve with a resolver or
Cancelled to acknowledge a requested cancellation. The resolver executes in the
serialized Store session with a fresh Caller for resource creation and reentry.
The starter Context receives cooperative cancellation; the resolver uses the
root invocation Context, allowing a normal result after a cancellation request.
Already-ready completions are valid and repeated publication/rebinding fails.
Wakeups are coalesced and observed without spinning or a lost pre-bind signal.

Immediate and delayed host functions share one canonical ownership transaction.
All parked host arguments/results reserve aggregate instance value/byte budgets;
cancellation, trap, panic and close release reservations and unpublished
resources. Cancellation acknowledgement consumes moved ownership and releases
borrows without producing a result. A failed consumer does not disable the
independent host provider. Futures/streams, WASI 0.3 worlds and complete M6
acceptance remain open. Performance optimization is deferred under the current
owner direction.

## Future/stream guest protocols and affine transport

Future and stream readable/writable ends use one state machine in the existing
bounded, tagged handle domain. Canonical new/read/write/cancel/drop/forward
operations enforce direction, element type, copy state and waitable membership.
Synchronous copies park the same Core continuations as synchronous lower/wait;
async copies return BLOCKED and report progress through the shared waitable
queue. Delivery returns buffer ownership and applies the specified precedence:
stream peer-drop overrides cancellation/completion, while a completed future
preserves its value even after peer-drop. Zero-length stream readiness and unit
streams keep their separate counting semantics without allocating one host Value
per event. Guest memory ranges are checked eagerly, with full-width memory64
pointers, lengths and stream results.

Parked buffers retain immutable canonical options and memory capabilities, not
Callers. Copying snapshots all source values before writing any destination,
using shared list/string/resource codecs and bounded scratch. Nested strings
use the receiving component's memory/allocator/encoding. Readable ends move
through flat/memory values and host calls as affine opaque capabilities; writes
stay local. Same-budget table moves do not transiently consume another live
handle. Runtime nominal comparisons inspect both instances' resource bindings,
even when generative instances share a static schema. `Instance.DropAsync` and
`Caller.DropAsync` consume host-owned readable endpoints and notify the peer.
Failed calls reclaim unpublished endpoints/copies; successful imported ownership
effects survive a later consumer trap.

Pinned wasm-tools 1.258.0 independently validates the ordinary guest protocols,
string composition, synchronous suspension, unit/byte streams and exports. It
lacks forward syntax/opcodes, so two forward binaries contain an explicitly
authored canonical entry from the pinned binary grammar. The same producer also
rejects memory64 copies after assigning them i32 pointers; two full-width owned
fixtures retain those recorded rejections. Neither group is counted as
independent producer validation. Wasmtime 48.0.1 matches twelve common guest
contracts and rejects the same four extension fixtures. Host stream production
and consumption APIs, WASI 0.3 worlds and complete M6 acceptance remain required;
transporting an endpoint through a host callback does not implement host I/O.


## Host async I/O and scheduled work

`Caller.NewSource`/`Instance.NewSource` create a typed future or stream from an
`AsyncSource`. Reads start lazily after the current guest turn parks; each read
returns a one-shot `ReadCompletion`. External workers publish a resolver which
runs with a fresh Caller in the serialized Store. Typed element batches, packed
bytes and counting-only unit streams share the canonical transfer transaction.
Futures produce exactly one element; empty non-final positive-capacity stream
reads are invalid. No Caller or guest execution stack escapes to a worker.

`Instance.ReadAsync` and `Caller.ReadAsync` drive one bounded host-side read in
the existing scheduler. `Function.CallWithResult` transfers an async result to
a scoped consumer as soon as task.return resolves it, permitting consumption
while the guest producer is still parked. It then drains background tasks.
Ordinary Call retains its existing completion behavior. `Instance.WithCaller`
provides the same scoped lease for embedding work without a guest entry.

`Caller.Schedule` moves explicit owned capabilities to an `AsyncWork` job, starts
it after the current turn, and invokes completion with a fresh Caller. Jobs use
the existing task, completion, cancellation and aggregate value/byte budgets;
there is no second executor. A one-job goroutine-safe `Cancellation` may be
requested before scheduling. Scheduled consumption starts independently of a
returned future being read, which is required for streaming output protocols.
Borrowed, duplicate, foreign and stale capabilities are rejected before moving.

Source Close is nonblocking, must not enter/wait for the VM, and runs once on
EOF, drop, failure or instance shutdown. Shutdown cancels workers before closure;
WaitClosed includes idle-source teardown. Cleanup releases all independent
resources even if a close callback panics, preserving the original execution
panic. Successful imported ownership effects remain committed after later
consumer failure; unpublished results and failed copies are reclaimed. The
canonical parameter-acceptance boundary publishes only accepted parameters,
not newly created result resources. Destination transactions stay live through
source closure so a failing close cannot leak owned payloads.

These APIs enable versioned WASI 0.3 providers but do not themselves implement
those interfaces or complete world/toolchain conformance. Performance acceptance
remains open under capability-first development.


## Synchronous LIFO reentry

The pinned Canonical ABI permits non-async function types to reenter in LIFO
order while ignoring explicit/implicit async backpressure. Every synchronous
canonical lift saves and restores its caller's task and entered state, installs
fresh context storage and retains the shared Store's fuel/reentry/frame quotas.
It does not borrow an outer async task's right to block. Realloc/post-return
leave guards, invalid ownership and failed-instance guards remain enforced.

Guest resource destructors are synchronous calls with fresh task state, even
for same-instance drops. They may reenter an entered defining instance, with
normal quotas and deterministic state restoration. Cleanup may invoke an eligible
defining destructor during another active call. A nested trap, panic or
cancellation restores the outer task before its own unwind/cleanup.

Three producer-validated fixtures cover host callbacks, explicit/implicit
backpressure, nested task storage, local/host-triggered guest destruction,
fully guest-composed indirect reentry, denied sync blocking and failure cleanup.
Go race and pinned LLGo execute both owned backends. Wasmtime 48.0.1 rejects the
two composed fixtures with `cannot enter component instance` and exits 134.
Its adapter generator unconditionally rejects same/ancestor/descendant adapters;
these are retained reference failures, not independent semantic passes. The
pinned ABI's Task.enter_implicit_thread and destructor rules remain the owned
implementation contract; complete cross-runtime acceptance stays open.
