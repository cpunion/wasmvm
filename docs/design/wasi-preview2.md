# Versioned WASI 0.2 host interfaces

The implemented interfaces are `wasi:cli/environment@0.2.12` and
`wasi:cli/exit@0.2.12`, plus all three `wasi:random` interfaces and `wasi:clocks/wall-clock` at 0.2.12, using the
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

Next interface groups are io error/poll/streams and stdio/terminals, clocks and
filesystem/preopens, sockets/DNS, then the complete CLI world and
HTTP. Reuse capability-relative filesystem and stream contracts; do not marshal
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
