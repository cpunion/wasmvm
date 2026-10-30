# WASI 0.2 execution fixtures

These project-authored components are generated with wasm-tools 1.258.0 and
validated with `wasm-tools validate --features all`. `environment` executes
all three pinned WASI 0.2.12 environment operations through Canonical ABI
memory/list/string/option transfers, including guest consumers and post-return
allocation reset. The mismatch fixtures are valid components whose imports
must be rejected by the host provider before guest execution.

The pinned WIT definition, hashes and upstream license are in
`testdata/wasi-preview2`. Wasmtime 48.0.1 C API executes the same fixture using
its actual WASIp2 provider with explicit arguments/environment and no inherited
host configuration. `exit` runs both nonreturning operations through canonical lower/Core/lift,
with two independently exported child instances. The same Wasmtime provider
verifies success, failure and explicit byte codes. The exit mismatch fixtures
are valid components rejected during host signature checking. These cover two
complete interfaces, not a complete world.


`random` imports all five operations across the three pinned random interfaces.
It exposes byte results, guest-side byte counts, scalars and the 128-bit seed;
allocation resets through post-return. The fixture covers empty/16 KiB/64 KiB
requests. Four independently valid mismatch fixtures must fail binding before
execution. wasm-tools 1.258.0 reproduces/validates the binaries; the same random
fixture executes with Wasmtime 48.0.1's actual WASIp2 C API provider. Own secure
counter-stream tests also retain independent OpenSSL AES-256-CTR vectors.


`wall-clock` imports both standard operations and exercises datetime memory
lower/lift plus a seconds-only guest read. Six independently valid negative
components vary result kind, field names/order/width/count and parameter count.
All seven binaries reproduce and validate with wasm-tools 1.258.0. The same
positive binary executes with the actual Wasmtime 48.0.1 WASIp2 provider;
provider precision differs explicitly from the own quantized default.


`poll-clock` executes all poll/monotonic operations, borrowed-list duplicate
indices, full u64 future deadlines and zero-duration creation/block/drop loops.
`probe` performs shared in-guest semantic checks; `churn(128)` is a complete
subscription-lifetime workload. `poll-source` additionally imports an embedder
factory for controllable readiness/cancellation tests. Four poll and four
monotonic mismatch fixtures independently validate; a wrong owned list retains
the method's required borrow receiver, so producer grammar checks are separate
from provider signature checks. All ten binaries reproduce with wasm-tools
1.258.0. The actual Wasmtime 48.0.1 WASIp2 provider executes `poll-clock` unchanged.


`io-streams.wit` selects the pinned error/poll/streams, stdio and terminal
interfaces. `scripts/build_wasi_io_fixtures.py` checks the WASI archive hash and
wasm-tools 1.258.0, emits a legacy-name dummy Core module for the exact WIT ABI,
replaces the authored guest bodies, and encodes/validates the final component.
The Core WAT, printed component, binary and checksum manifest are retained.
`exercise` checks all stream operations, exact bytes/counts, EOF and drops;
`error-probe` checks a real failure resource, debug string and later closed state.
Other exports check terminals, write permits, blocking size and parent/child
ownership. `write-batch(128)` writes/flushes 4 KiB per iteration to a checked sink.

Eight contract components alter one valid WIT signature and retain unreachable
bodies: binding must fail. `io-errors` separately checks an embedder error factory,
owned debug snapshots, UTF-8 limits, typed causes and expired/stale capabilities.
All ten components independently validate. Reproduction does not imply an
upstream conformance corpus or complete CLI/HTTP world acceptance.

`filesystem.wit` selects the pinned IO/filesystem interfaces. Reproduce with
`scripts/build_wasi_filesystem_fixtures.py`; the script verifies the original
archive and producer version, retains authored Core/printed Component WAT and
hashes, and independently validates every binary. `exercise` executes descriptor
and directory methods with guest assertions; `stat-batch(128)` checks stat size
and closes every descriptor. `error-probe` requires host-induced filesystem
revocation and validates the downcast error resource. The actual Wasmtime
WASIp2 provider executes the common exercise and stat workload unchanged.

`filesystem-api` reexports the same imported nominal resource domain for direct
embedding tests. Nine valid WIT mutations test width, tuple/record fields,
variant/enum labels, stream resource identity and preopen tuple order. These
reject at binding before unreachable guest code runs. This authored corpus is
separate from upstream/world/toolchain conformance acceptance.

`network.wit` imports the complete pinned network, instance-network and IP name
lookup interfaces. Its authored guest checks IPv4/IPv6, readiness, EOF and child
resource drops; `lookup-batch(128)` repeats complete literal lookup lifetimes.
The API view reexports the same imported nominal resource types. Six valid
signature mutations must fail binding before guest execution. Resolver/error
provenance, IDNA, cancellation and quotas additionally have embedding tests.

`scripts/build_wasi_network_fixtures.py` reproduces and independently validates
all nine binaries. The normal consumer retains `network-error-code`, missing
from Wasmtime 48.0.1's C WASIp2 linker. `network-common` is a separately generated
narrow consumer which omits only that unused function requirement. Both engines
execute the identical common binary. The original complete-reference failure
remains recorded; this subset does not establish full reference compatibility.

`tcp.wit` imports the complete TCP/create interfaces and their pinned dependencies.
`scripts/build_wasi_tcp_fixtures.py` reproduces and independently validates eleven
components: full/API/common consumers and eight valid type mutations (address
field/width, stream order, keepalive width, shutdown label, scalar/own results).
The API view reexports the same nominal imports. The complete guest checks all
29 TCP operations and exact 4 KiB transfer lengths/pattern endpoints on IPv4/IPv6.
`setup` opens one persistent connection; `transfer(128)` performs 128 exchanges
of 4 KiB in each direction; `teardown` shuts down and drops every owned handle.
Setup/teardown stay outside warmed transfer timing. The common consumer omits
only the unused `network-error-code` import missing in Wasmtime 48.0.1. The actual
Rust WASIp2 provider executes it with TCP permission explicitly enabled; the C
API cannot enable TCP and its access-denied diagnostic is retained separately.

`udp.wit` imports all 19 UDP/create and datagram-stream operations. Reproduce
with `scripts/build_wasi_udp_fixtures.py`: full/common/API consumers and eight
valid type mutations independently validate with pinned wasm-tools 1.258.0.
The full guest checks connected/unconnected IPv4/IPv6 sessions, options,
readiness, zero-length packets, explicit peer addresses and resource drops.
`setup` opens persistent endpoints; `transfer(128)` exchanges 4 KiB in each
direction per roundtrip; `teardown` drops all owned handles. Wasmtime 48.0.1's
actual Rust WASIp2 provider runs the identical common binary with UDP enabled;
only its unused missing network-error-code requirement is omitted.
