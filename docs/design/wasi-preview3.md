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
| sockets/types, sockets/ip-name-lookup | all 41 operations: TCP, UDP, socket options and async name lookup |

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
empty cancellation acknowledges without inventing progress. Rejected output
operations drop their input immediately and resolve their error future without
waiting for a producer. Result futures report
typed filesystem errors independently of stream EOF. Provider closure cancels
FIFO readiness and restores operation/object reservations.

Current owned fixtures independently parse/validate with wasm-tools 1.258.0 and
execute on both backends under Go/LLGo. They are ABI/provider integration tests,
not complete upstream world conformance. HTTP, complete world
assembly and independent guest-toolchain fixtures remain required. Performance
campaigns are deferred until capability coverage is complete.

Actual Wasmtime 48.0.1 matches the async Core guest's successful open/stat size
and missing-file error variant. The API-only fixture reexports imported functions;
Wasmtime rejects that adapter feature, recorded separately from the two executed
reference cases. These checks do not establish full world/toolchain acceptance.


## Socket providers

NewSockets(ioOwner, SocketsConfig) binds P3 types and ip-name-lookup directly;
P3 has no P2 network/pollable/IO resource interfaces. TCP create/bind/listen,
async connect, one-shot send/receive and all options implement the versioned
contracts. Listen produces a native stream of independently owned sockets;
send/receive use the shared byte-stream and result-future kernel. Streams retain
the OS socket after its original handle is dropped. Normal send completion
half-closes the write direction; dropping receive half-closes its read direction.
Rejected duplicate streams do not shut down the original operation. Invalid
input operations publish their error future without waiting for stream reads.

UDP implements explicit/implicit bind, repeatable association, disconnect,
individual async datagrams, addresses/family and options. Empty datagrams stay
messages, and receives retain sender addresses. Connected sends accept absent or
matching destinations. Maximal datagrams, address families, mapped IPv6,
multicast/unspecified remote addresses and port zero are checked. The shared
host syscall/address/option layer preserves IPv6 flow and numeric scope for both
P2 and P3. Darwin disconnect's ambiguous errno is accepted only after a peer
query confirms that the association is gone.

Authority is explicit. Nil AllowBind/AllowConnect/AllowSend/AllowReceive deny
those operations. TCP connections inherit the grant used for connect or the
listener; UDP send and receive check peer policy per datagram. Callbacks run
without provider locks, observe cancellation and may run concurrently. Policy
panic denies access. No host network or DNS authority is inherited by default.
Literal lookup parses locally; enabled external lookup normalizes IDNA, removes
invalid/zoned results, unmapped IPv4 duplicates and empty results. Resolver
callbacks are bounded by name/address/work limits and a timeout.

MaxSockets bounds published capabilities and physical objects separately,
including unpublished accepted sockets and deferred closes. A bounded cleanup
queue retains FD leases through half-close and closes outside VM callbacks.
MaxWorkers bounds async datagram/connect/DNS/readiness work; IO limits also bound
stream operations and batches. Native batch completion cleanup releases accepted
FDs after publication, cancellation or abandonment, including late returns.
Close cancels pending operations; WaitClosed joins socket work and physical
cleanup. Close/wait for Instances, then Sockets, then its IO owner.

Four socket fixtures independently validate with wasm-tools 1.258.0. A real Core
async guest sends a datagram; actual Wasmtime 48.0.1 sends the same bytes with
explicit network authority and traps at the guest's denial assertion without it.
API, ownership, limits, cancellation, readiness, FIN, DNS and both IP families
execute on both owned backends under Go/LLGo. Full world/toolchain and HTTP
acceptance remain open.


## HTTP providers

`NewHTTP(sockets, HTTPConfig)` binds the pinned types (35 operations), async
client.send and async handler.handle. Fields and nominal request/response/options
resources implement the P3 variants and ownership rules, including immutable
header views, independent clones, request-option children and consuming body
accessors. Pure header storage is shared with P2; wire bindings remain versioned.

A message owns optional native byte contents and an original trailers future.
An affine source anchor retains these endpoints until send/consume/drop. Body,
trailers, handling-result and transmission futures have independent leases.
Closing a transmission future does not cancel the message. A scheduled trailer
observer consumes its future without retaining a callback stack; an early error
interrupts pending body I/O. Trailer results resolve only after body closure.
Dropping or failing either consumer propagates typed failure and reclaims the
remaining endpoints. The canonical stream status and detailed HTTP error future
remain separate contracts.

Client requests require separate URL and resolved-connection grants. External
DNS uses the explicit socket resolver policy. No proxy, credentials, filesystem,
network or environment authority is inherited. Callback panic denies/fails the
operation. TLS trust configuration is cloned. Connect/first-byte/between-bytes
timeouts, declared content length, headers, worker/request/message/resource
limits and nonreused representation IDs bound work across Stores.

The outbound wire profile is HTTP/1.1, including HTTPS. P3 trailer names may only
become available at body EOF. Announcing them before sending would require
buffering the complete body; Go's HTTP/2 server discards unannounced trailers.
The provider therefore uses HTTP/1.1 chunked transfer and validates declared
length independently. Broader outbound wire-protocol support remains explicit
future work, not a silent loss of trailers. The server-facing HTTPHandler uses
the embedding server's standard HTTP protocol support.

Guest byte batches enter a zero-buffer pipe through the existing scheduler;
network goroutines never enter guest code. Physical message readers are closed
by a bounded reclaimer. Operation completion retains a separate late-result
lease through publication or abandonment. Provider Close revokes endpoints;
WaitClosed joins admitted work and physical cleanup. Close/wait instances and
HTTP, then Sockets, then IO before releasing external storage.

`CallIncoming` checks the async guest handler and transfers its response as soon
as the result is available. The invocation keeps driving guest producers while
the external consumer reads. Response Body.Close cancels and joins the invocation.
`HTTPHandler` gives each request its own Store/owners through an explicit factory.
A failed invocation requires discarding that request-scoped instance.

Six fixtures are generated from the pinned WIT graph and independently validated
by wasm-tools 1.258.0. They include malformed contracts, a real outgoing Core
guest and an early-result server guest with a suspended byte producer. Wasmtime
48.0.1 executes the outgoing request (204) and serves the streaming guest (200,
`hello`). API/provider tests cover both owned backends, Go/LLGo, TLS, DNS, late
trailers, length/progress errors, early replies, quotas, cancellation and shutdown.
These checks do not close complete-world, independent language-toolchain, cm64
or full conformance acceptance. Optimization remains deferred.
