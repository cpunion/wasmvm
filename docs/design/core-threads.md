# Core Threads

Enable `wasm.Threads` in `EngineConfig.Features`, or use CLI `--threads`.
It is independent of `Profile: "core3"`; Memory64/shared combinations also
require Memory64 (or the Core 3.0 profile). Both execution backends implement
shared memory and all 67 atomic instructions. Core Threads does not create
guest threads itself: embeddings create separate Stores and explicitly share
linear memory. Legacy wasi-threads, shared GC objects and component threading
remain separate work.

## Ownership and quotas

Each Engine owns one synchronization domain. `Extern.ShareMemory()` creates
an independently retained `SharedMemory` capability; `Store.ImportMemory(cap)`
imports it into another Store of that Engine. Imports deduplicate within a
Store. Functions, tables, globals, GC references and private memories retain
their Store ownership. No mutable guest heap is implicitly shared.

Closing a capability prevents future imports. Existing bindings remain valid.
Closing the creating Store releases its binding, while sibling Stores and
capabilities retain the backing. Last release frees it. Shared memory uses
owned native allocation from creation whenever native execution is available,
including for interpreter-created memories. No migration, mirrored copy or
retained Go-slice pointer is needed when a native Store imports it.

Every owning/importing Store reserves the backing's effective host maximum
against its memory-page quota. This prevents sibling growth from silently
exceeding an importer's quota. `Usage.Core.SharedMemoryPages` reports current
shared pages and `ReservedSharedMemoryPages` reports the reservation;
`MemoryPages` includes actual private and shared pages. Declared shared memories
require an explicit maximum. The effective maximum can be lower under host
limits, and growth then returns the normal failure sentinel.

## Execution and waiting

One domain mutex orders all shared ordinary scalar/SIMD access, atomics,
size/grow, bulk operations, initialization and host byte I/O. This is a
conservative sequentially consistent implementation, including mixed widths.
Fence orders the domain even in modules with no memory. Atomic accesses check
bounds and exact natural alignment; waits trap on private memory, while private
notify returns zero. Private ordinary accesses retain their direct path.

Before native optimization, only instructions accessing shared memories are
marked for checked individual runtime helpers. Raw descriptors for shared
memories are absent. Optimizations cannot introduce unsynchronized loads or
stores; the rest of each function is compiled normally. This is not a
whole-function interpreter fallback. Per-access helper and locking costs remain
an explicit performance follow-up, especially for contention and scaling.

Wait checks the expected value and registers under the ordering lock, then
releases it before blocking. Wait32 and wait64 share an address-keyed queue.
Negative timeout waits indefinitely; mismatch returns 1, timeout 2 and notify 0.
Notify, timeout, cancellation and Store close arbitrate once under the same
lock. Intrusive queues make registration/cancellation constant time; notify is
linear in the number actually woken. `EngineConfig.MaxWaiters` bounds the
Engine-wide number of registered/in-flight waiters (default 65,536). Closing a
Store cancels its calls and drains its waiters before releasing memory.

## Evidence

The [pinned corpus](../../testdata/threads/README.md) runs all 17 applicable
complete upstream scripts, including nested thread/wait commands and concurrent
litmus tests. Each interpreter/native tier executes 946 commands and 579 binary
assertions; six text-parser assertions are classified separately. Child runners
have isolated Stores and registries, inheriting only explicit shared-memory
capabilities. Reports count actual execution, join all children before cleanup,
and retain compiler, target, source/manifest identity and feature mask.

The concurrent corpus runs under Go's race detector as well as normal Go/LLGo
tests. Authored cases add cross-Store RMW, shared growth, mixed SIMD/bulk/host
access, Memory64 overflow/alignment, wait32/64 queue identity, waiter limits,
cancellation, close, capability lifetime and quota reservations. Race checks
supplement the memory-model contract; they are not proof of native semantics.
M3 performance acceptance remains open; P1/legacy threads and components/WASI
0.2/0.3 are subsequent stages.
