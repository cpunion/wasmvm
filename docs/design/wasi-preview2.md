# Versioned WASI 0.2 host interfaces

The first implemented interface is `wasi:cli/environment@0.2.12`, using the
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
random/exit, filesystem/preopens, sockets/DNS, then the complete CLI world and
HTTP. Reuse capability-relative filesystem and stream contracts; do not marshal
component calls through P1 linear-memory ABIs. Byte-list transport needs a
shared packed representation before sustained IO performance acceptance.
