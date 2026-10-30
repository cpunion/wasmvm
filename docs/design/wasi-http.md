# WASI 0.2.12 HTTP implementation contract

Status: all 51 stable types operations, outgoing handling and the guest incoming
handler execute on Linux/Darwin. Complete-world acceptance, independent guest
toolchains and performance gates remain open. The source is the unchanged
release `http.wit` and its checksum.

## Complete scope

The stable profile requires all 51 operations in `wasi:http/types`, the outgoing
handler and the exported incoming handler: 53 callable operations in total.
The WIT feature `informational-outbound-responses` is unstable and excluded from
the stable fixture profile; do not silently count its send-informational method
as stable support. The declared proxy world also requires the already implemented
clocks, random, IO and stdio interfaces. P2 pollable futures do not implement
canonical async or WASI 0.3.

Implement every resource together: fields, incoming/outgoing requests, request
options, response outparams, incoming/outgoing responses and bodies, future
trailers and future incoming responses. Preserve exact variants/records and
nested option/result states, with checked nominal IO/pollable identities.

## Ownership and shutdown

Headers returned from requests/responses are immutable children. Clone creates
an independent mutable snapshot. All explicit WIT children must drop before
parent drop or transfer: header views, body IO ports, future pollables and
returned trailers. A future's successful get is single-use; subsequent ready
get calls return the outer unit error. Keep pending, transport failure and
already-consumed states distinct.

Consume incoming body/obtain outgoing body and obtain their IO streams at most
once. Finish consumes the body and rejects live IO children. Incoming finish
must drain unread data without retaining it so trailers can complete even when
the guest did not read the entire body. Outgoing finish closes production and
validates any declared content length/trailers. Dropping a parent early traps
while preserving eventual cleanup through the remaining child leases.

Owned arguments move into the host. Consume their canonical handles exactly
once after protocol checks and transfer the underlying state to its new owner;
do not retain an expired Caller or guest borrow. Instance.NewResource supports
incoming request/outparam injection without a synthetic guest factory.

Close must revoke new operations, cancel requests/pumps, wake pollables, close
transport readers and join all host work. WaitClosed includes detached cleanup
and retains failures. Store close alone does not close the provider. IO has a
separate owner lifetime; its port closure must not accidentally cancel a body
that finish is about to transfer to a trailers future.

## Streaming architecture

Use bounded circular body buffers rather than materializing entire messages.
Transport/request workers own the blocking side; IO ports remain nonblocking.
Readiness snapshots and notification channels change under the same lock.
Buffers allocate lazily and grow geometrically to the configured ceiling.
Partial writes apply backpressure. HTTP source flush accepts bytes into this
bounded queue; it does not wait for a remote reader or a handle call that may
occur later. EOF/error completion follows already accepted data. No mandatory
worker belongs to each individual IO port. Body/transport ownership determines
which close cancels or drains a pump.

Preserve field-name case and byte-valued fields, case-insensitive lookup,
immutable views and independent clones. Validate HTTP token/value syntax,
forbidden names and bounds before committing a mutation. Bound both field
count and total bytes; a failed mutation must preserve the previous value.
The fields API preserves insertion order and repeated values. The Go transport
normalizes wire header names/order while preserving each name's value order;
it does not promise byte-identical header serialization.

## Transport and embedding

Use private HTTP transport/resolver state, explicit URL and resolved-connect
authority, and no ambient proxy/credentials. Reuse Network name normalization,
lookup bounds and cancellation. Default outgoing access is denied. Do not
follow redirects implicitly. TLS validation, connect/first-byte/between-byte
timeouts, body framing, cancellation, trailers and typed IO error downcasts
are required. Map unsupported u64 timeout values to the declared setter error
instead of wrapping time.Duration.

A proxy server must create an independent Store/instance per incoming request
and stream a response concurrently with guest execution. Waiting for a full
guest return before consuming its output can deadlock a bounded body writer.
Never dispatch a self-request into the same busy Store. Limit active requests,
resources, header storage and buffered bytes; close request-scoped owners even
when a guest abandons handles. Keep caller-controlled fuel/context budgets
explicit and preparation separate from warmed guest execution.

## Acceptance

Run full guest client and proxy flows on interpreter/native and Go/LLGo release:
HTTP and HTTPS, streaming in both directions, empty/partial bodies, headers,
trailers, statuses/methods/options, timeout/error variants, redirects, ownership,
limits, rollback, cancellation and shutdown. Preserve full-world and unstable
feature failures rather than weakening a consumer silently. Validate producers
and bytes independently, and keep executable-package coverage above 90%.

`http-fields` imports the complete 51-operation stable types interface. Its
field/body workloads and API view preserve nominal domains; eight independently
valid mutations reject binding. `http` includes the complete stable proxy world
and exercises outgoing requests and the exported incoming handler. Wasmtime
48.0.1's actual Rust HTTPp2 provider executes both guests, including 4096-byte
bodies and trailers, with its default sender retained. These authored consumers
do not establish independent guest-toolchain or full upstream world conformance.
All 39 error-code payloads have Canonical ABI roundtrip tests; this is separate
from causing every possible physical transport failure.

Measure fields, persistent HTTP requests and streaming separately. Report load,
compile, provider/Store setup, bind, first request, warm requests, connection
startup and teardown. Run source base/head and pinned competitors serially on
stable source, preserve failures and expose incomplete performance targets.
Cross-language/world conformance and canonical async remain separate gates.

Pinned primary reference sources are
[HTTPp2](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/p2/mod.rs),
[resource types](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/p2/types.rs)
and [field storage](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/field_map.rs).

## Public embedding API and limits

`NewHTTP(network, HTTPConfig)` reuses the explicit Network/IO owners and their
pollable/error domains. `Imports` includes HTTP types/outgoing handling, shared
IO and a stable monotonic clock; supply a wall clock and random interfaces to
instantiate the authored proxy consumer. `CallIncoming` invokes the standard
guest export on an idle instance and returns a streaming Go response.
`HTTPHandler` takes a factory supplying an independent instance and owners for
each request; its cleanup closes and joins the owners. Closing the returned
body cancels its request lifetime. Never reuse a busy Store for a self-request.

Defaults are 4096 live HTTP resources, 128 active outgoing requests, 1024 header
fields/64 KiB header bytes per fields value, 64 KiB buffered bytes per body and
a 30-second overall request limit. At most MaxResources bodies remain tracked,
including transferred bodies. Request options override connect, first-byte
and between-byte durations without removing the overall limit. Zero means an
immediate timeout; values beyond signed host duration return the setter error.
Eight private timeout-specific transports form a bounded cache; eviction closes
idle connections. Nil URL/connect grants deny outgoing access. Header
restrictions and TLS roots are immutable snapshots. No ambient proxy, redirect
following or automatic decompression is enabled. The standard transport's
supported HTTP versions are used; WebSocket upgrades and canonical async are
outside this stable HTTP profile.

## Reference differences and benchmark reproduction

Run the common external server with
`go run benchmarks/components/http_server.go AUTHORITY_FILE` and set
`WASMVM_HTTP_BENCH_AUTHORITY` to the file's contents. Run the package benchmarks
`BenchmarkHTTPFields`, `BenchmarkHTTPBodyTransfer`, `BenchmarkHTTPRequests` with
Go or the pinned LLGo release profile, and the Rust `http_fields`/`http`
executables against the exact same component files. `http_fields ... COUNT body`
selects the pure 4096-byte body workload. Each call contains 128 lifetimes;
preparation and first-batch metrics are outside warm timings. The server's
`/stats` reports accepted connections, validated requests and invalid messages.
Serialize benchmark processes and retain raw logs/checksums.

The reference's unmodified default sender opens a TCP connection per request;
own transports reuse connections. Full-request timings compare complete
provider policies and cannot establish pure engine throughput parity.
The common client explicitly declares its request trailer. An earlier valid
consumer without that declaration was rejected by the strict echo server:
the reference sent the 4096-byte body without its trailer. Preserve that fixture
and failure as diagnostic evidence, rather than weakening assertions. The
incoming-handler reference validates status, body and response trailers
separately from client throughput. Reference HTTPS with private roots has no
common default-hook configuration here; own TLS verification/rejection tests
do not establish a reference TLS comparison.

The pinned LLGo runtime can delay HTTP cancellation by approximately one second.
A standard-library-only reproduction is recorded in
[llgo #2953](https://github.com/xgo-dev/llgo/issues/2953). Tests allow three seconds
for eventual completion without changing the request's protocol timeout values
or weakening typed failure assertions. This compatibility/latency deficit remains
open; successful eventual cleanup does not establish prompt cancellation parity.
