# Serialized scalar storage accessors

Canonical `context.get` and `context.set` bind the active task's two integer
cells through the shared VM scalar-storage primitive. It supports both I32
and I64 getters/setters. An I32 setter zero-extends into the 64-bit cell.
Initialization, realloc, stackful reentry and asynchronous resumption use the
existing namespace/task transitions. The binding reads the current task and
failed-namespace flag on every execution entry; no second task scheduler or
HTTP-specific path is introduced.

`vm.NewScalarStorage` fixes the cell count and provides a serialized binding
callback. `Store.NewScalarStorageFunction` and its Caller counterpart create
ordinary typed Store functions, with the existing function quotas and reference
type checks. The binding cannot reenter the VM. Its returned slice remains
stable until the execution entry returns; invalid or incorrectly sized bindings
trap. Aliased accessors share the same storage object. Users must serialize any
external access, as with ordinary host callback state.

The interpreter resolves an accessor after ordinary call-target, type and
cancellation checks, then reads/writes the cell without creating a host request
or Caller. Root and tail calls share the same accessor semantics through the
existing host driver. Imported accessors remain valid function-reference targets.

The compiler emits a guarded load/store at direct integer import call sites.
The runtime view appends a scalar-view pointer after the heap/allocation views,
preserving existing function, table, global, memory and heap offsets. The scalar
view contains a value address and validity address for each function index.
Ordinary imports have zero addresses and take the original checked exit.
Native cells and metadata belong to the execution workspace; no Go address
crosses the machine ABI. All accessors for one storage object share its cells.
The driver snapshots cells before entry and publishes writes on every exit,
including traps, fuel exhaustion and cancellation, before executing host code.
Idle workspaces release binding/task roots; native allocations close with their
workspace. Indirect and tail calls retain the shared checked call path.

Call instructions retain their normal fuel charge and interrupt check. Accessors
introduce no guest frame and do not change frame/slot quotas. An invalid native
binding takes the checked call path to preserve the original trap behavior.
Performance acceptance requires complete response-checked component workloads;
a generated-code test proving no host exit is a structural check only.
