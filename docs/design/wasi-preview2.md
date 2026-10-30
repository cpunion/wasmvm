# Versioned WASI 0.2 host interfaces

The implemented interfaces are `wasi:cli/environment@0.2.12` and
`wasi:cli/exit@0.2.12`, plus all three `wasi:random` interfaces at 0.2.12, using the
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
