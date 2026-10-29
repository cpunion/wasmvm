# Legacy WASI threads

`pkg/wasi/threads` implements the separately enabled legacy `wasi.thread-spawn`
ABI on Linux/macOS, following [the pinned proposal](https://github.com/WebAssembly/wasi-threads/tree/f8ea03cc0ffbde143415e945497de112f6f9fd53).
It does not implement component threading or a shared WasmGC heap.

## Ownership and execution

`New(ctx, engine, module, process, config)` creates shared imported memories
within the configured Store limits. It accepts P1 function imports, the typed
`wasi.thread-spawn(i32) -> i32` import, and shared Memory32 imports. Other imported
object domains are rejected. The command must export `_start()`, and a spawning
module must export `wasi_thread_start(i32, i32)`.

Successful construction transfers Process ownership to the Group. Engine and
CompiledModule remain caller-owned and must stay open while new threads can be
created; a compiled module can serve successive independent Groups. Failed
construction preserves the caller's Process. All native code is compiled once
and leased by Stores. No thread recompiles or falls back to another backend.

Each thread has an independent Store, instance, tables, globals, execution stack
and GC domain. Shared memories and P1 process capabilities are the common state.
The opaque start argument is passed unchanged. The guest libc initializes and
reclaims pthread stacks, TLS and thread-specific data; the host never guesses
their ABI layout. Returning from a worker only releases its own Store. A trap,
`proc_exit`, main return or process cancellation cancels the whole group.

Normal Core initialization, including the module start function, executes on
every instance. There is no blanket suppression of active segments or start.
The pinned SDK25 producer uses passive data plus an atomic once protocol; the
integration fixture verifies child initialization preserves modified shared data
and still initializes private globals. An arbitrary module's active segments
retain their Core semantics, including repeated writes to imported memory.

TIDs increase monotonically in `[1, 2^29)`, with no reuse. A reservation counts
toward `MaxThreads` before allocation or initialization begins. The limit
includes the main thread; default 256. Failed creation releases its reservation
and returns negative EAGAIN. A guest trap during child initialization terminates
the group. Memory, descriptor and execution limits remain explicit. Fuel is a
separate per-thread budget including that instance's start; a context deadline
bounds the whole process. It is not an aggregate process fuel budget.

`Close` requests cancellation and closes all published Stores. Pending creation
rechecks closure before publication. `WaitClosed` joins worker execution, Store
reclamation and P1 callbacks/resources; it cannot complete while a child or an
initialization reservation is live. No asynchronous work retains a Caller or
guest-memory slice. Preserve typed P1 exit, trap and context errors.

## CLI and validation

`run --wasi-threads` enables P1 and Core Threads; use `--profile core3` for LLGo's
final EH. `--max-threads` sets the total live/reserved thread limit. Existing
explicit environment, stdio, directory and native-code quota flags apply.

The [offline corpus](../../testdata/wasi-threads/README.md) contains all 15 upstream
programs and a pinned LLGo guest. They run in interpreter, native and precise-fuel
tiers under Go/LLGo. Upstream cases cover main/worker exit while another thread
is in a busy loop, atomic wait, P1 poll or blocked fd read. Authored cases add
TLS/private globals, once initialization, nested and repeated spawn, cancellation,
TID exhaustion, quota/type/link failures and concurrent close. The guest exercises
C pthread join/TLS/destructors, Goexit defers, cross-thread GC, EH, actual SIMD,
concurrent file I/O and worker process exit/panic. Coverage and race checks are
additional implementation evidence, not proofs of the native memory model.

Competitive host-call and shared-memory performance, controlled M4 performance
acceptance and component/WASI 0.2/0.3 support remain separate work.
