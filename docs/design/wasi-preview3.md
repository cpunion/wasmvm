# WASI 0.3.2 providers

The scope is pinned to the WASI 0.3.2 WIT archive with SHA-256
`2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311`.
The stable CLI command/imports and HTTP service/middleware worlds remain the
required complete outcome. Optional timezone and component threading are separate
profiles. Current providers implement the following subset; this does not close
M6 or complete-world conformance.

| Interfaces | Implemented operations |
| --- | --- |
| cli/environment | get-environment, get-arguments, get-initial-cwd |
| cli/exit | exit, exit-with-code |
| cli/stdin | read-via-stream |
| cli/stdout, cli/stderr | write-via-stream |
| cli/terminal-input, terminal-output | nominal resource types |
| cli/terminal-stdin, terminal-stdout, terminal-stderr | each get-terminal operation |
| cli/types | error-code type |
| clocks/types | duration type |
| clocks/monotonic-clock | now, get-resolution, async wait-for, async wait-until |
| clocks/system-clock | now, get-resolution |
| random/random, random/insecure, random/insecure-seed | all five operations |
| filesystem/types, filesystem/preopens | all 26 operations: 21 async metadata/path methods, four native-stream methods, get-directories |

All names bind under `@0.3.2`. Pure environment/exit/entropy backends share the
existing P2 implementation through explicit operation selection and signature
checks. P3 renames initial-cwd to get-initial-cwd and insecure-seed to
get-insecure-seed. Random byte requests clamp to the configured maximum: zero
returns empty, positive requests return at least one byte, and large requests
may return a short read. P2 exact-request limits remain unchanged.

System-clock instant uses signed s64 seconds and normalized u32 nanoseconds;
resolution is a nonzero u64 duration, not the P2 datetime record. Custom Now and
Resolution callbacks are supplied together, observe cancellation and may execute
concurrently across Stores. Default system/monotonic clocks expose microsecond
ticks. Monotonic timers use native async host completions, bounded concurrent
waits and full-width u64 deadlines. Pending-wait reservations are released before
publishing completion; worker join accounting remains live through publication. Relative overflow never becomes an immediate
wait. Close cancels waits; WaitClosed joins timer workers.

IO sources use the same explicit, prompt readiness contract as P2, while the P3
surface uses native streams/futures. Defaults are EOF/discard and absent terminals;
no ambient stdio is inherited. Construction transfers source ownership. Configured
sources support concurrent calls/Close, cancellation and coherent readiness.
MaxStreams bounds active operations across Stores, read/write limits bound each
batch and at most twice MaxStreams external workers may be active. Store task,
resource, fuel and value budgets also apply.

Stdin returns independent readable streams and persistent completion futures.
EOF resolves success, input failures resolve the declared error-code, and dropping
the stream resolves success. Short reads and zero-length readiness are supported.
Output moves its input stream to scheduled work immediately, independent of the
returned future being read. One bounded batch is copied to external I/O per
completion, with the next batch scheduled through the existing task scheduler.
Partial writes, readiness, final flush and failures preserve ordering. The result
future completes after EOF/flush or an error; failure drops the input endpoint.
Dropping the output result cancels scheduled consumption. Resource transfer and
publication failures reclaim unpublished endpoints and operation reservations.

External workers never retain a Caller or enter the VM. They publish one-shot
resolvers into the serialized Store session. Caller.FunctionInfo supplies the
actual invocation's immutable nominal signature when creating source endpoints;
a provider-global type capture would be invalid across generative Stores.
Close revokes operations and starts source cleanup without blocking VM callbacks;
WaitClosed joins workers and reports source-close failures. Bound Instances are
closed/waited independently before external storage is released.

RunCommand requires `wasi:cli/run@0.3.2#run` to be async with no parameters and a
payload-free result. Normal failure maps to status 1; explicit exit preserves its
status. Traps and cancellation remain errors. This invocation helper does not
synthesize a complete command world.

Filesystem binds the complete pinned types/preopens interfaces on Linux and
Darwin. NewFileSystem takes the IO operation owner and explicit preopens; no
ambient directories are inherited. Descriptor resources have generative Store
identity and provider-wide, nonreused representations. Capability-relative host
operations enforce granted roots and read-only rights. Streams retain independent
leases after the originating descriptor is dropped. Directory streams open
independent cursors and emit bounded batches of typed entries.

P3 filesystem error-code and descriptor-type are variants, including
other(option<string>); they are not P2 enums. There is no public would-block error.
Detailed errors travel in completion futures; the pinned canonical ABI reports
stream progress/status without an error-context closure operand, despite older
closure wording in the WIT comments. File I/O retries readiness internally, including nonseekable FIFO partial reads
and writes. Signed timestamps preserve negative seconds. Offsets and lengths
check signed host limits before system calls. Append serializes flag changes with
other writes on the same open object and restores its original flags.

MaxDescriptors bounds published handles and separately bounds physical backing
objects, including preopens, cursors, pending opens and queued closes. Worker,
path, entry and batch limits are explicit. A bounded reclaimer closes descriptors
outside VM callbacks; admission waits for pending reclamation with cancellation.
Async completions retain a separate result lease until publication or abandonment,
and reclaim late results. MaxWorkers reservations are released before signaling
completion while join accounting remains live. Close revokes streams and starts
shutdown; WaitClosed joins workers and physical descriptor cleanup. Close and
wait for bound Instances, then the filesystem provider, then its IO owner.

Read cancellation preserves bytes already consumed by an external operation;
empty cancellation acknowledges without inventing progress. Result futures report
typed filesystem errors independently of stream EOF. Provider closure cancels
FIFO readiness and restores operation/object reservations.

Current owned fixtures independently parse/validate with wasm-tools 1.258.0 and
execute on both backends under Go/LLGo. They are ABI/provider integration tests,
not complete upstream world conformance. Sockets, HTTP, complete world
assembly and independent guest-toolchain fixtures remain required. Performance
campaigns are deferred until capability coverage is complete.

Actual Wasmtime 48.0.1 matches the async Core guest's successful open/stat size
and missing-file error variant. The API-only fixture reexports imported functions;
Wasmtime rejects that adapter feature, recorded separately from the two executed
reference cases. These checks do not establish full world/toolchain acceptance.
