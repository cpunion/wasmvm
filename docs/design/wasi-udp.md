# WASI 0.2.12 UDP implementation contract

Status: implemented on Linux/Darwin; complete-world and performance acceptance
remain open. This sockets increment does not complete M5.
The pinned source is `testdata/wasi-preview2/sockets.wit` and its manifest.

The provider implements all 19 callable operations across UDP and UDP creation: 13 socket
methods, two incoming-stream methods, three outgoing-stream methods and creation.
It shares explicit Network, IO error and pollable resource domains with TCP.
Shared helpers provide socket address validation/encoding, complete IPv6 flow/scope preservation,
FD leases, cancellation-safe readiness and network error mapping. UDP has its own protocol state and datagram resources.

An explicit UDP provider owns live sockets and datagram streams. Bind, send and
receive permissions are separate cancellable callbacks; nil denies. Connected
streams retain the exact peer; unconnected sends require an explicit peer and
perform authorization for each datagram. Receive drops unauthorized packets,
with bounded scanning so a denied-packet flood cannot monopolize a host call.
Connected receive filtering must match the OS peer, including address family.
Callbacks run outside provider/resource locks and must observe cancellation.

Sockets bind through start/finish and must bind before creating streams. Stream
reconfiguration requires dropping the old incoming/outgoing streams first;
child pollables retain stream lifetime, and stream resources retain the socket.
Dropping a parent too early traps while preserving cleanup until the last child
is released. Partial publication rolls back every resource and reservation.
Close revokes calls, cancels readiness and joins FD/cleanup work. Closing the
parent Network also closes UDP; IO and Store lifetimes remain explicit.

A nonempty send consumes a preceding check-send permit. Exceeding that permit
traps. An empty send returns zero without consuming the permit. Each datagram
is atomic: no partial byte counts. Report already-sent datagram count if a later
datagram fails or blocks; otherwise preserve its standard network error. A
receive with max-results zero returns an empty list. Zero-length UDP packets
remain packets, rather than EOF. The default MaxDatagrams is 16 (configurable up to 64), bounding per-call
receive scanning, returned packet count and check-send permits. Packets are
limited to 65,535 bytes (the OS may impose a lower payload maximum). A reusable
receive buffer holds a complete UDP payload; zero-length packets remain valid.
MaxSockets defaults to 128 and MaxStreams to 256. Address/peer validation and quotas precede
network side effects where possible. OS-backed buffer/hop getters return actual
values; zero setters reject, and oversized buffer values clamp safely.

The authored full WIT consumer must exercise all operations on IPv4 and IPv6,
including connected and unconnected sessions, zero-length packets, explicit
addresses, batching, subscribe/readiness and owned-resource drops. Valid type
mutations must reject binding. Add failure tests for grants, invalid peers,
permit misuse, packet limits, cancellation, backpressure, error provenance,
reconfiguration, orphaned resources and failed publication. Execute on both
backends and Go/LLGo release; keep every executable package at 90% coverage.

Compare the exact common guest against Wasmtime 48.0.1's actual Rust WASIp2
provider with UDP explicitly enabled. Preserve the full consumer's missing
network-error-code failure and any C API permission limitation. Time provider
setup, bind/stream setup, first batch, warmed persistent datagram exchange and
shutdown separately. Report source-base API absence and an unchanged shared
workload separately. Retain local deficits without claiming acceptance.

Pinned primary reference sources are the
[WASIp2 adapter](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/p2/host/udp.rs)
and [socket implementation](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi/src/sockets/udp.rs).
The reference's completion-based scheduling is not an implementation requirement.
