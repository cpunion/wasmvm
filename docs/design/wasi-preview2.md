# Versioned WASI 0.2 host interfaces

The implemented interfaces are `wasi:cli/environment@0.2.12` and
`wasi:cli/exit@0.2.12`, using the
pinned release WIT and license in `testdata/wasi-preview2`. All three operations
execute through `pkg/component` and the own Core interpreter/native compiler:
get-environment, get-arguments and initial-cwd. This does not complete the CLI
or HTTP world, M5, component async, or WASI 0.3.

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
random, filesystem/preopens, sockets/DNS, then the complete CLI world and
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
