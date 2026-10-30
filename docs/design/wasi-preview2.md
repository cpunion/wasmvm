# Versioned WASI 0.2 host interfaces

The implemented interfaces are `wasi:cli/environment@0.2.12` and
`wasi:cli/exit@0.2.12`, plus all three `wasi:random` interfaces, stable clocks,
IO, stdio, terminals, filesystem/preopens, sockets/DNS and stable HTTP at 0.2.12, using the
pinned release WIT and license in `testdata/wasi-preview2`. All operations execute
through `pkg/component` and the own Core interpreter/native compiler. The
environment operations are get-environment, get-arguments and initial-cwd.
Complete CLI/HTTP worlds, M5, component async and WASI 0.3 remain open.

`NewEnvironment(EnvironmentConfig)` owns explicit argument/environment/current
directory snapshots. It does not read host process configuration implicitly.
UTF-8/NUL validation, entry and aggregate configuration-byte limits run before
binding. Environment entries split at the first equals sign; order, duplicate
names and empty values are retained. A nil initial directory produces none;
a supplied directory is a guest path value, not filesystem authority.

`Environment.Imports()` supplies callbacks and immutable function signature
checks. Incompatible parameter/result shapes are rejected before guest starts;
the provider does not accept a component's declaration of a standard name as
proof of its type. Returned values are detached, stable across calls, and use
the shared Canonical ABI codec for strings/lists/tuples/options. Configuration
limits and runtime transfer/value/nesting limits remain independent. Callback
copies observe cancellation; guest post-return owns allocator cleanup.

Tests execute every operation through both own backends with Go and LLGo,
including UTF-8, configuration/result mutation isolation, empty defaults,
invalid schemas and cancellation. The same project-authored component also
runs with Wasmtime 48.0.1's actual WASIp2 C API provider, with explicit argv/env
and no inherited host state. This is a targeted interface check, not a complete
upstream conformance suite or two independent guest-language toolchains.

Remaining gates include complete declared worlds and their independent
guest-toolchain acceptance. Reuse capability-relative filesystem and stream contracts; do not marshal
component calls through P1 linear-memory ABIs. The shared Canonical ABI codec now supports packed immutable byte sequences;
IO/random providers must use that representation and preserve declared bounds.
Sustained IO performance acceptance remains open.

## Nonreturning exit

`ExitImports()` implements the entire pinned exit interface: `exit(result)` maps
success/failure to status 0/1, and `exit-with-code(u8)` preserves all byte values.
Both import signatures are checked before any guest starts. The host receives a
wrapped `*ExitError` even for successful exit; use `errors.As` and inspect `Code`.
An exit is an explicit guest outcome, and the callback's typed error survives
the cancellation caused by shutdown.

`Caller.CloseInstance()` requests nonblocking closure of the common component
Store. It uses the callback's existing expiry/serialization/cancellation guard;
copied or retained callers cannot acquire shutdown authority after return.
Closure stops all linked namespaces and invalidates held exports/resources.
Store reclamation waits for the active callback/Core session to unwind. A host
must wait for `Instance.WaitClosed` outside the callback, never inside it.

The execution fixture calls both operations through canonical lower/Core/lift
and makes the following guest instruction unreachable. Tests check 0/1/42/255,
all linked instance exports, invalid signatures/input, cancellation and final
reclamation under both own backends and host compilers. Wasmtime 48.0.1's actual
WASIp2 C API provider independently checks all four statuses on the same binary.
Exit consumes an instance, so the benchmark prepares fresh instances outside
the call timer. Own calls include Store cancellation/reclamation; the reference
reports call and explicit Store destruction separately. No recurring warm-call
or equal lifetime policy claim is made for a nonreturning operation.


## Random interfaces

`NewRandom(RandomConfig)` implements all five operations in random, insecure and
insecure-seed at the pinned 0.2.12 version. Secure/insecure bytes and u64 values
use separate AES-256 counter streams. The constructor obtains independent keys
and a 128-bit hash-map seed from crypto/rand before any guest can bind; callbacks
perform no entropy IO. The secure interface accepts no public deterministic
seed or source. An embedder needing deterministic execution must omit the secure
interface. It uses Go's existing [AES/CTR implementation](https://pkg.go.dev/crypto/cipher#NewCTR).

`MaxBytes` limits each request, defaulting to 64 KiB, with a nonzero range from
eight bytes through the Canonical ABI's maximum byte-list length. Scalar u64
requests always fit a valid limit. Component logical element, transfer, guest
memory and nesting budgets remain independent; a configured request limit does
not override them. Oversized requests fail before allocation or stream advance.
Packed bytes use the shared codec; a u64 consumes the same source stream with
little-endian representation. Seed results are owned, stable per provider and
independent of the secure key and insecure stream.

Stream leases serialize shared-provider calls across Stores and observe context
cancellation while waiting. Pure generation checks cancellation between 4 KiB
chunks and returns every lease on failure. Canceled generated bytes may be
consumed/discarded; they are never reused or published. Repeated calls continue
the counter rather than reseeding in guest callbacks. Zero/uninitialized providers
export no imports and fail normal binding instead of blocking on an empty lease.

Tests execute all operations and cancellation through both own backends with Go
and LLGo, including shared providers, owned results, exact independent OpenSSL
counter-stream vectors, default maximum requests, entropy initialization failure,
invalid signatures, empty requests and quotas. The same guest executes using
Wasmtime's actual WASIp2 provider. Performance compares a real 16 KiB generation
and canonical lowering/count/post-return workload, excluding initialization from
the warmed call. The declaration of a complete interface is not a full CLI/HTTP
world, upstream conformance suite or performance acceptance claim.


## Wall clock

`NewWallClock(WallClockConfig)` supplies both operations in
`wasi:clocks/wall-clock@0.2.12`: now and resolution. Both imports require exactly
a record with fields seconds:u64 and nanoseconds:u32, in that order, and no
parameters. The shared signature checker rejects wrong names/order/widths and
counts before guest starts. Numeric type exports use the existing structural
host binding; the provider creates no new nominal resource identity.

Datetime retains unsigned seconds independently of signed nanosecond duration
limits. Host callbacks must return normalized nanoseconds below one billion;
resolution must be nonzero. Zero epoch is valid for now. Invalid host values or
source errors trap the call. Returned records are owned snapshots.

A zero configuration selects the host realtime clock with explicit microsecond
granularity: reads are quantized to microseconds and resolution returns 1,000 ns.
This is the provider's exposed clock, not a claim about the OS clock's native
precision. It uses separate Unix seconds/nanoseconds and avoids UnixNano's
signed range. No timezone or monotonic guarantee is implied. An embedder may
supply both cancellable callbacks for another resolution, virtual time or
restricted clock capability. Supplying only one callback is rejected. Callbacks
can run concurrently in separate Stores and must honor context cancellation;
results are checked again after return. Store close cancels active reads.

All operations execute through both own backends/compilers and the same binary
using Wasmtime 48.0.1's actual WASIp2 provider. The complete clocks package still
requires monotonic-clock with pollable subscriptions. IO/poll, streams, remaining
world dependencies and async/WASI 0.3 remain separate required work.


## Pollables and monotonic clock

`NewPoller(PollConfig)` implements the entire `wasi:io/poll@0.2.12` interface:
pollable.ready, pollable.block and poll. `NewMonotonicClock(poller)` implements
all four stable `wasi:clocks/monotonic-clock@0.2.12` operations. Its Imports
includes both interfaces and binds the identical nominal pollable declaration.
The checker verifies borrow/own, list element kinds, parameter counts, scalar
widths and that shared declaration before any guest effects. Numeric resource
IDs, foreign Pollers and foreign Stores do not confer authority.

A Poller is an explicit host capability owner, like the P1 Process. Close it
after closing its bound Instances; Store closure cancels active calls and
invalidates guest handles, while the Poller retains its independently owned
subscription entries until explicit drop or Poller.Close. Closing the provider
wakes active waits and reclaims abandoned subscriptions, joins already dispatched
timer callbacks, and stops future publication. Concurrent/repeated Close waits
for the same reclamation. Timer callbacks run no guest/user code. This separate
ownership is intentional; a Store close does not synthesize guest destructors.

The domain defaults to 1,024 live pollables and 1,024 items per poll. MaxPollItems
may be configured through 65,534, leaving two cancellation slots in the standard
reflect.Select limit. Provider entries and Core/Canonical resource quotas remain
independent. Representation IDs never recycle within a provider. Failed
publication releases provisional timers; quota exhaustion is reported explicitly.

`PollSource.Readiness` supplies a coherent nonblocking readiness/notification
snapshot for future IO adapters or embedder capabilities. Sources may change
readiness repeatedly: after a notification, a false snapshot must supply the
next notification channel. IO failure counts as ready; the IO operation reports
the error. Poller borrows externally supplied sources and does not close their
external IO ownership. Multi-source poll returns each ready input index,
including duplicate occurrences, and rejects empty/oversized lists. Snapshot
and notification selection avoid lost wakeups and allocate no goroutine per
source. Single-source block needs neither a result list nor reflective select.
All waits observe call context/Store and provider cancellation.

The monotonic clock has a private origin and explicit microsecond ticks. Now is
nondecreasing and resolution is 1,000 ns. Absolute past and zero-duration
subscriptions become ready; future subscriptions remain pending. Relative
deadlines use unquantized elapsed time so a rounded read cannot shorten the
requested duration. Durations beyond signed time.Duration are waited in bounded
chunks. A relative deadline exceeding the u64 clock range remains pending rather
than wrapping into readiness. Explicit resource drop and provider Close cancel
and join timer work. These synchronous pollables are separate from Component
Model native async, futures/streams and WASI 0.3.

The same fixture executes all seven operations through Wasmtime 48.0.1's actual
WASIp2 provider. Its zero-deadline resource deliberately yields to its scheduler;
a first nonblocking ready may therefore be false. The shared semantic probe
blocks before asserting completed readiness, then verifies duplicate indices,
huge deadlines and explicit drops. This scheduling policy differs from own
immediate readiness, and remains visible in performance interpretation. IO error,
streams/stdio, filesystem, sockets/HTTP and complete worlds are still required.


## IO, stdio and terminals

`NewIO(IOConfig)` implements error, streams, stdin/stdout/stderr and all five
terminal interfaces at 0.2.12. It includes the shared poll interface. All 22
callable operations and resource declarations have signature/nominal checks;
wrong byte widths, result payloads, error cases and borrowed/owned identities
fail binding before guest effects. `IO.Poller()` is borrowed for a compatible
monotonic clock; IO closure also closes that pollable domain.

Zero configuration supplies EOF input and discarded output. Configured sources
transfer to a successfully constructed owner. Standard getters mint independent
Store capabilities while the provider pins the configured source. NewInput and
NewOutput accept additional owned sources in active callbacks. Rejected closed
owners retain no source ownership; accepted creation closes the source if
publication fails. Numeric representations, foreign types/Stores and expired
callers cannot mint or inspect capabilities. Stream IDs never recycle.

A source implements prompt thread-safe input Read or output Write/Flush and a
coherent readiness snapshot. IO errors count as ready. Output Write must not
retain supplied bytes. Pending source Flush must signal completion before
reporting ready again. Close is idempotent and unblocks operations. BytesInput
owns a byte snapshot; OutputBuffer is a bounded host-drained sink with change
notifications and detached Drain results. Neither inherits process stdio.

Reads return available bytes, an empty success while unavailable, or a typed
stream-error; blocking operations wait with Store/call/provider cancellation.
Requested u64 lengths are capped at MaxReadBytes (default 64 KiB), as WIT permits
short reads/skips. Partial data precedes a deferred error. Closed input rejects
even zero-length requests; open zero-length input does not wait. Splice reports
the stream that actually failed and preserves the order of check/read/write.

Output check-write grants a one-use permit. Unchecked or over-permit write and
write-zeroes trap before queue allocation. A pending flush advertises zero
capacity. Nonblocking writes copy into a bounded queue (default 64 KiB per
stream), with one active worker per output record. Blocking write-and-flush
uses the same prompt source directly when no older output is queued, preserving
ordering and cancellation without a mandatory goroutine handoff. Blocking
payloads above 4,096 bytes trap. Flush completion, backpressure and partial writes
observe coherent notifications. Worker exceptions become stream failures.

The first permanent operation failure returns last-operation-failed with a
nominal error capability; subsequent calls return closed. To-debug-string
returns a bounded valid UTF-8 snapshot (default 4 KiB), replacing invalid or
truncated encoding with ASCII markers. Debug text is human information, never
parsed as an error ABI. The host cause remains available through IOErrors
Describe for typed adapters. Explicit drops release messages and ownership.
IOErrors has its own bounded domain (default 1,024 errors).

The provider bounds stream records at 1,024 by default, including retiring
worker records. Dropping a stream with live child pollables traps; drop children
first. Drop cancels/joins output work and may discard unflushed data, as WIT
permits. IO.Close requests domain cancellation and stops pending polls; WaitClosed
joins active creation/calls, workers and source cleanup, and reports cleanup
errors. It must run outside callbacks. Store close invalidates guest handles and
cancels calls; the explicit provider remains the owner of abandoned resources
until IO.Close. Source cooperation remains part of the host callback contract.

Terminal flags are explicit configuration for each stdio endpoint. Defaults
return none, matching virtual non-terminal sources; enabled endpoints mint opaque
marker capabilities with the shared terminal-input/output declarations. These
markers grant no filesystem or extra console authority.

The pinned actual Wasmtime WASIp2 provider executes the same generated component,
including all stream operations, stdio, terminal-none and IO-error resource
creation/debug/drop. Own tests additionally exercise terminal-some and strict
write permits. The reference stdio backend accepts an initial unchecked write;
own code enforces the WIT precondition. The pinned generic reference host
implementation delegates check-write/write to its concrete stream backend.
Record this observed policy difference separately from successful common
operations. Complete worlds, filesystem, sockets/HTTP and native async remain
required; these interfaces alone close no M5/M6 acceptance gate.

## Filesystem and preopens

`FileSystem` implements all 30 callable operations in
`wasi:filesystem/{types,preopens}@0.2.12` on Linux and Darwin. It reuses
`internal/wasi/host` directory capabilities and the same IO owner's nominal
stream/error resources; no Preview 1 memory layouts or implicit cwd are involved.
The exact WIT contract is checked before guest effects, including record fields,
variant/enum/flags labels, borrowed receivers and owned result resources.

Construction opens only explicitly configured preopens. Each preopen has an
independent immutable guest name and read-only/mutable policy. Derived directory
handles retain capability-relative roots, symlink confinement and stable object
checks. Mutating operations and writable derived handles require directory
mutation authority; file data operations separately enforce read/write flags.
Absolute readlink targets are rejected; relative symlink targets convey no
additional authority. Hard links and renames authenticate both descriptors.

Descriptor and directory-entry resources have bounded provider tables and
never-reused identities. Defaults are 1,024 descriptors, 128 directory streams,
65,536 entries per snapshot, 4,096 path bytes and 64 KiB per positional read or
write. Large positional requests may make bounded partial progress. Directory
snapshots omit `.`/`..`, reject invalid UTF-8 and do not hold a directory handle
after capture. Unknown host failures map to `io`; configured resource exhaustion
maps to `insufficient-memory`. Offsets/sizes outside the host signed file domain
and unrepresentable timestamps return `overflow`, without truncation or wrapping.
The full unsigned datetime shape is validated before host timestamp conversion.
Stat timestamps use the shared host layer's unsigned nanosecond representation;
pre-epoch or larger host dates outside that representation report overflow.

Streams have independent cursors and backing leases: dropping a descriptor does
not invalidate an already returned stream. Regular files use positional IO;
other objects use nonblocking IO with coherent readiness notifications and at
most one cancellable poll waiter per stream. Atomic append temporarily enables
kernel `O_APPEND` under the shared backing IO lock; positional writes share that
lock. Failure to restore flags prevents subsequent writes through that backing.
Flush completes issued writes without forcing durable sync; `sync` and
`sync-data` explicitly synchronize writable descriptors. Advisory hints may be
ignored; integrity flags request host `O_SYNC` for both file/data integrity.

Filesystem Close revokes new operations, cancels waiters, waits for active host
operations, then closes all remaining backing objects exactly once. Concurrent
stream close joins its waiter/cleanup and releases its lease. The IO owner has a
separate lifetime: close/wait for both providers, including abandoned resources;
Store close alone does not close either owner. Filesystem-origin IO errors retain
structured causes for `filesystem-error-code`; unrelated IO errors return none.

`filesystem.wit` and its checksum manifest pin the source. The authored guest
checks file/stream IO, flags/types/stat/timestamps/hashes, links, directory
iteration and drops; the same binary executes with actual Wasmtime WASIp2.
A second valid component reexports the same nominal imported interfaces for
embedding/resource tests. Nine independently valid contract mutations must fail
binding. The revocation probe is host induced and is not a common-provider
reference workload. Filesystem interface support does not complete a CLI world,
sockets/HTTP, canonical async or WASI 0.3.

## Network capabilities and IP name lookup

`Network` supplies all five callable operations in
`wasi:sockets/{network,instance-network,ip-name-lookup}@0.2.12`. It shares the
explicit IO owner's nominal error and pollable domains. A network handle grants
only configured name lookup; TCP/UDP socket authority is not implied. Name
lookup is denied by default. An explicit grant uses a private Go resolver with
`PreferGo` or an embedder-supplied, cancellable resolver callback. No resolver is
called for literal IPv4/IPv6 addresses. Bracketed IPv6 is accepted, scoped IPv6 is
rejected by this unscoped IP value interface, and named queries use pinned
`x/net/idna` lookup normalization. DNS results normalize mapped IPv4 and remove
duplicates; explicitly supplied IPv6 literals preserve their family.

Resolution returns a nominal address-stream resource immediately. Pending
`resolve-next-address` returns `would-block`; subscribe returns a child pollable
which becomes ready on success or failure. Completion returns each address,
then persistent none; resolver failures retain their structured error code.
Default bounds are 128 address streams, 256 addresses per result, 1,024 input
name bytes and a 30-second timeout. Excess results fail without scanning an
unbounded result slice. Resolver panics become permanent resolver failure.

Each unresolved stream owns a cancellable worker. Dropping it cancels and joins
that worker; dropping before child pollables traps and preserves cleanup until
the children are released. Failed resource/pollable publication rolls back its
reservation. Provider Close rejects new operations, cancels every query, and
WaitClosed joins active creations and provider workers. Resolver callbacks must
observe cancellation and return immutable address data. Closing the IO owner
separately reclaims pollable/error resources. A Store has its own lifetime and
does not implicitly close either provider. `network-error-code` downcasts only
original typed NetworkError causes; unrelated IO errors return none.

The complete WIT consumer, a view reexporting the exact same imported nominal
domain and six independently valid type mutations exercise the full own
interface. Wasmtime 48.0.1's C WASIp2 linker lacks `network-error-code`, so the full
consumer fails there with a missing-function link error. A separately retained
consumer narrows the unused import requirement by omitting only that function;
both engines run identical bytes for IPv4/IPv6 semantics and warmed literal
resolution lifetimes. No stub/reference implementation is substituted. See the
[reference source](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/p2/ip_name_lookup.rs)
and [host lookup policy](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/sockets/ip_name_lookup.rs).
This is P2 polling over asynchronous host work, not canonical async or WASI 0.3.
HTTP and complete worlds remain required.

## TCP sockets

On Linux/Darwin, `TCP` implements all 28 methods in `wasi:sockets/tcp@0.2.12`
and `create-tcp-socket` in `tcp-create-socket`. Construct it with an explicit
`Network` backed by the IO owner. `Imports` includes these dependent interfaces
and a stable monotonic-clock provider sharing the same pollable domain. Neither
socket creation nor a network handle grants bind/connect permission: nil policy
callbacks deny both. Callbacks receive the complete address (including IPv6
flow-info/scope-id), run outside provider/socket locks, may run concurrently,
and must observe cancellation. Accepted connections inherit listener authority.

The provider supports IPv4 and IPv6-only sockets, nonblocking start/finish
bind/connect/listen, accept, local/remote addresses, all keepalive/hop/buffer
options, backlog changes, polling, independent IO streams and directional,
idempotent shutdown. Zero option values reject where WIT requires; oversized
values clamp to the documented OS-supported range. Keepalive durations round
up to seconds and clamp to 32,767 seconds; count clamps to 127. Buffer getters
return the actual OS value. Successful listener settings for keepalive idle,
hop limit and buffers are applied explicitly on accepted sockets. Flush means
kernel acceptance, without adding a delivery acknowledgement. TCP_NODELAY
retains its OS default in both own and reference providers.

The default live-socket limit is 128, including active creation reservations.
Input/output and pollable quotas come from the shared IO owner. Streams and
pollables lease their parent socket: dropping a socket before its children
traps and preserves cleanup until the last child drops. Failed publication
rolls back all partially created resources. Syscall/poll leases prevent FD
reuse until cancellation joins active work; stream IO handles partial transfers,
backpressure and EOF without a mandatory worker handoff. Typed network causes
survive IO error downcasts. Readiness uses at most one waiter per direction.

Close revokes permission, cancels waiters and closes every socket; WaitClosed
joins active calls, physical closes and already detached cleanup, retaining the
first close error. Closing the parent Network also closes TCP. Close/wait for
IO separately to reclaim its streams/errors/pollables. Store close alone does
not synthesize owner close. Embedders must not close provider-owned OS FDs.

The complete consumer executes IPv4/IPv6 guest assertions and all 29 operations
on both own backends. Eight independently valid WIT contract mutations must
fail binding; an exact-domain API view covers lifecycle/error paths. The common
consumer omits only the unused `network-error-code` import missing in the pinned
Wasmtime provider. Actual TCP reference execution uses Wasmtime 48.0.1's Rust
WASIp2 provider with explicit `inherit_network().allow_tcp(true)`. Its C API
exposes address inheritance but no `allow_tcp` setter; the retained C probe
returns access-denied even for the common consumer. No replacement network
provider is used. HTTP, complete worlds and canonical async remain required.


## UDP datagrams

On Linux/Darwin, `UDP` implements all 19 socket/create and incoming/outgoing
stream operations. It uses the same Network/IO pollable domains, with separate
bind/send/receive grants and bounded datagram batching, peer filtering, send
permits and joined resource cleanup. See [the UDP contract](wasi-udp.md) for
lifecycle, limits, errors and validation scopes. The complete IPv4/IPv6 guest
executes on both backends and the actual pinned Wasmtime Rust WASIp2 provider.
HTTP, complete worlds, remaining component mechanisms and native async remain
required.

## HTTP client and incoming handler

`HTTP` supplies all 51 stable types operations and the outgoing handler on
Linux/Darwin, with `CallIncoming`/`HTTPHandler` executing the standard guest
incoming handler. It shares the Network/IO resource domains, separate URL and
resolved-connect grants, bounded streaming and joined cancellation/cleanup.
The authored proxy consumer, HTTP/HTTPS clients and concurrent incoming handler
execute on both own backends with Go/LLGo. Actual pinned Wasmtime HTTPp2 also
executes the common HTTP client and incoming guest without replacement hooks.
See [HTTP design and embedding](wasi-http.md) for limits, ownership, reference
differences and performance phases. Stable interfaces and authored proxy flows
do not complete M5: independent guest languages, full-world conformance,
remaining component mechanisms and performance targets remain explicit gates.
WASI 0.3 requires canonical async.

## Unified world assembly and standard entries

`NewWorld(WorldConfig)` composes the existing stable providers with shared
nominal domains and explicit configuration. `RunCommand` checks the standard
CLI unit-result export before invocation and preserves result/exit/error
outcomes. `Function.Info` also supports checking HTTP incoming own-resource
signatures before effects. Independent pinned C and Rust commands compose and
execute on both backends/host compilers and actual Wasmtime WASIp2. The complete
CLI import fixture separately checks binding after compiler DCE. See
[world assembly](wasi-worlds.md) for ownership, validation boundaries and
performance phases. `assembledWorlds` records integration scope separately from
`completeWorlds`; upstream conformance, remaining component semantics and M5/M6
performance/async gates remain open.
