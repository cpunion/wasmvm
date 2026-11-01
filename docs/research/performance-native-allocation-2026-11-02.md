# Bounded native allocation: 2026-11-02

Native entries now allocate fixed-layout structs in bounded batches without
exiting for every object. A longer experiment including **every observed GC
stall** measures LLGo at **60.198 ns/object**, versus Wasmtime's **60.174**;
the validated baseline is 582.293. This is descriptive workstation proximity,
not controlled superiority or complete GC/interpreter performance acceptance.

## Sources and measurements

Base `915ca0ce2af8aaf0e801c7563293d5b98375b81b` is merged PR #91. Measured head
`b37ff6832f77f4e907307b3c88cb4f45d2877130` contains the implementation; subsequent
documentation does not change engine sources. Both are built independently
using Go 1.27.3 and LLGo `8e45ed5` with `-O3 -lto=full` on Apple M4 Max/macOS
arm64. Reference pins remain Wasmtime 48.0.1 Cranelift, WasmEdge 0.17.1
interpreter, wazero 1.12.0 and Wasmer 7.3.0. Executable and live SDK hashes were
verified against the original inventory. Wazero/Wasmer GC rejections remain
actual unsupported outcomes, without synthetic timings.

The unchanged allocation fixture retains a linked chain and publishes its last
node to a global. Its existing independent oracle checks every call, including
alternate seeds/counts. The ordinary five-case comparison retains 10,000
iterations, three fresh interleaved process blocks, two warmups and seven warm
calls. A separately declared sustained experiment uses 100,000 allocations
and million-iteration i31/struct/indirect controls, with the same protocol.

Those short series reveal a GC sampling problem: Wasmtime's allocation calls
alternate between fast nursery allocation and substantial collection pauses.
With 10,000 iterations, the median can be about 100 ns/object; with 100,000 it
can be about 8 ns/object while omitting the cost of the slow calls. Neither
median alone describes sustained allocation throughput.

The final GC-cycle diagnostic therefore retains **63 warm calls per fresh
process**, 100,000 iterations per call, three interleaved blocks and the same
two warmups. It reports the median of the three **process means**, including
all pauses, rather than pooling calls or discarding slow observations:

| Native compiler/runtime | Base ns/object | Head ns/object |
| --- | ---: | ---: |
| wasmvm Go | 211.759 | 43.540 |
| wasmvm LLGo | 582.293 | 60.198 |
| Wasmtime Cranelift | — | 60.174 |
| WasmEdge interpreter | — | 102.796 |

LLGo process means are 60.198, 59.992 and 60.416; Wasmtime's are 60.095, 60.518
and 60.174. These shared-host observations do not resolve a calibrated win.
Preparation/startup and first calls remain distinct; no timing is subtracted.
Pinned C adapters retain seven independent clock-pair calibration probes;
their 63 warm result/timing samples are validated separately without modifying
records. This diagnostic changes neither the existing benchmark suite nor its
controlled acceptance criteria.

Long-loop LLGo struct access remains 1.269 ns/iteration against base 1.271 and
Wasmtime 1.014; indirect access is 4.467 against base 4.491 and Wasmtime 2.533.
The interpreter still has substantial deficits. All ordinary, sustained and
GC-cycle samples, orders, checksum results, phase records and source/toolchain/
binary/SDK hashes are in the [evidence](evidence/performance-native-allocation-2026-11-02.json).
Earlier batch/layout/append experiments remain included rather than replaced.

## Runtime and compiler mechanism

Before an eligible native entry, the Store reserves at most 64 generation-checked
handles and a stable payload region. Its private native header contains only
numeric handles, pointers to owned native storage, offsets and quotas. All
guards run before writes. The emitter initializes packed/numeric/reference/
v128 fields before assigning an output slot, which may alias a consumed input.
The allocation cursor commits last.

Both amd64 and arm64 emit this path. Precise semantic-fuel entries, forced
collection, descriptor constructors, continuations, large structs and other
shapes retain complete checked helpers. Capacity/quota guards exit at the
original helper PC with operands intact. Bump allocation never relocates
existing objects or recycles live handles, so existing checked address caches
remain valid until entry exit.

Every exit, including traps/cancellation/errors, publishes committed objects
into the same typed heap before collectors, host calls, globals or reentry can
observe them. Local type IDs map through the current instance; usage/serials
and relocation records update together, and unused handles return to the free
list. No Go pointer enters the native protocol. The stable entry record ABI is
unchanged; only its private runtime-data view gains an allocation-header pointer.

Host metadata uses lazy, stable 64-object pages indexed by handle. Collection
checks complete generations before following old relocation tombstones. Released
cells clear every pointer-bearing field. Pages retain the handle high-water
mark, bounded by the handle quota, until Store close. The collector similarly
reuses a cleared marking worklist bounded by the object quota. These retained
host workspaces are explicit memory tradeoffs, not guest-live-byte accounting
or a claim about process RSS.

The unchanged 26-instruction/eight-slot allocation fixture grows its ordinary
entry from 1,784 to 1,975 bytes on amd64 and 1,672 to 1,848 on arm64 (about 11%).
The increase is the guarded constructor and its existing slow exit. Precise
entries remain 4,519/3,368 bytes respectively; aggregate code limits still apply.

## LLGo allocation finding

Sampling attributed significant residual work to host allocation inside scalar
`append` operations. A minimal reproduction with a preallocated 10,000-element
slice reports zero extra bytes under Go and approximately 314 KB under pinned
LLGo; direct indexed writes report zero for both. The real compaction loop emits
`AllocU(24)` followed by `GrowSliceAligned` for each append despite sufficient
capacity. LLGo's `MemStats.Mallocs` reports zero in this probe and is not used as
an allocation-count claim. The issue and exact reproduction are
[LLGo #2957](https://github.com/xgo-dev/llgo/issues/2957).

Heap publication/free lists/tracing now write directly into available capacity
and delegate actual growth to the ordinary builtin. Native compaction writes
kept records in place. This shared Go implementation preserves both compilers'
semantics without changing the toolchain baseline or adding a runtime dependency.

## Validation and remaining work

Go/LLGo tests cover field widths/raw FP/v128 bits, aliased results, large frames,
slow guards, empty objects, quotas, generation reuse, typed chains, rollback,
forced GC, trap/cancel publication and host reentry. Full pinned core/proposal
inventories continue to execute on both backends. Race, strict cgo-pointer and
no-cgo checks pass. All 30 executable-package coverage thresholds pass locally;
final target CI is required before integration.

Interpreter dispatch/GC helpers, shared-memory synchronization, WASI/component
boundaries, broader allocation/liveness behavior and controlled two-platform
performance acceptance remain open. One linked-chain workload cannot accept
the whole collector or the original performance objective.
