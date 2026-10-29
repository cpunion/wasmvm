# WASI Preview 1

`pkg/wasi/preview1` implements the 46 typed `wasi_snapshot_preview1` imports
on Linux and macOS. The [pinned WITX and program inventory](../../testdata/wasi-preview1/README.md)
define the scope. This host adapter is independent of Core Threads, legacy
`wasi.thread-spawn`, components and WASI 0.2/0.3.

## Process and resource ownership

A Process owns arguments/environment, descriptor rights, preopens, stdio,
clocks/random providers, cancellation and process exit. `Install(store, linker)`
atomically binds the complete import module. Multiple Stores can use one
Process without sharing guest tables, stacks, globals or GC roots. This is the
resource boundary for the subsequent legacy thread adapter.

Arguments and environment are snapshots. Defaults expose realtime/monotonic
clocks and cryptographic random bytes, empty stdin, discarded output, and no
environment, directory or network capabilities. Preopens explicitly map a host
directory to a guest name; read-only mappings remove mutation rights, including
hard-link source rights. Socket capabilities require `ImportSocket(file,
listening)`; the adapter never opens ambient network connections.

Descriptor bindings and in-flight operations have independent leases. Closing
or overwriting a binding cannot recycle its host descriptor under an active
operation. Comparable aliases to the same stdio Stream share ownership. Other
injected streams must supply independent ownership leases. All Stream methods
must honor their context and be safe for concurrent use; Close must return
promptly. No callback or worker retains guest slices or Caller capabilities.

Derive guest-call deadlines from Process.Context so process cancellation also
interrupts pure Wasm loops between host calls. Process.Close cancels callbacks,
releases bindings and closes resources outside
its mutex. WaitClosed joins final cleanup. Closing one Store cancels its own
calls while sibling Stores retain the Process. `proc_exit` records an immutable
exit code, cancels the Process and returns a typed ExitError; errno, trap,
context cancellation and process exit remain separate outcomes.

## ABI, files and polling

The ABI uses checked Memory32 pointers and little-endian layouts. Every iovec,
path and output range is validated before external effects. Scratch memory,
iovecs, paths, descriptors, directory snapshots and aggregate active poll
workers have explicit limits. Rights only decrease; renumber requires both
source and destination bindings, rather than introducing dup2 semantics.

Directory operations resolve ancestors through os.Root and hold stable parent
handles. Final no-follow operations use `*at` syscalls with the final basename;
`.`/`..` are resolved through the root before mutation. Reopened directory
streams verify object identity. No string-prefix containment or path cleaning
substitutes for filesystem resolution. Cross-directory rename/link use held
handles. Timestamp updates preserve nanoseconds and target held objects;
Darwin uses fsetattrlist instead of reopening a mutable pathname.

Owned file streams remain nonblocking at the kernel boundary. Descriptor access
uses SyscallConn, since File.Fd can restore blocking mode. Read/write, socket
accept/receive/send, poll and file-stream Close observe cancellation without
abandoned read goroutines. Poll first snapshots all ready subscriptions; only
an actual wait starts bounded workers, which are canceled and joined before
return. FileStream retains its descriptor until its own active operations drain.
File flags implement append, nonblocking and synchronous I/O; abstract streams
explicitly reject unsupported changes instead of ignoring them.

The default host supports clocks 0/1; optional CPU clocks 2/3 return EINVAL.
Nonzero `proc_raise` returns ENOTSUP, without signaling the embedding process.
Absolute symlink targets are rejected. Stream-specific unsupported flags return
ENOTSUP. These declared host limitations are distinct from missing ABI imports.
Host filesystem errors remain observable; allocation is subject to its actual
filesystem support and configured resource limits.

## Validation and native code capacity

All 72 complete upstream C/Rust/AssemblyScript P1 programs execute on the
interpreter, unmetered native and precise-fuel native tiers. Authored tests cover
socket capabilities, flags, partial I/O, quotas, pointer overflow, shared process
resources, poll cancellation, descriptor reuse, close and symlink/rename races.
Reports count completed guest programs, not unknown internal assertion totals.

A Go 1.27.3 guest combines goroutines/channels, linear-memory GC, panic/recover,
random, file I/O and timed waits on the same tiers under both host compilers.
It needs an explicit 256 MiB native code budget; the default module budget stays
64 MiB and its observed rejection is retained. Darwin's one bounded MAP_JIT arena
now reserves 1 GiB of virtual address space, with pages populated on publication.
Compiler errors identify the function, tier and consumed code budget.

Legacy wasi-threads and LLGo's threaded guest integration remain the next M4
capability. Performance comparisons and calibrated M1–M3 acceptance remain open.
