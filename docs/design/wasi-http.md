# WASI 0.2.12 HTTP implementation contract

Status: implementation in progress. Header storage, bounded body buffering and
pinned type fixtures are local preparation; no HTTP interface/world is advertised
as implemented. The source is the unchanged release `http.wit` and its checksum.

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
Partial writes apply backpressure, flush readiness waits for queued bytes to
drain, and EOF/error completion follows already accepted data. No mandatory
worker belongs to each individual IO port. Body/transport ownership determines
which close cancels or drains a pump.

Preserve field-name case and byte-valued fields, case-insensitive lookup,
immutable views and independent clones. Validate HTTP token/value syntax,
forbidden names and bounds before committing a mutation. Bound both field
count and total bytes; a failed mutation must preserve the previous value.
The transport's serialization order must agree with fields.entries.

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

`http-fields` imports the complete 51-operation stable types interface, but its
authored probe exercises only field operations. Its API view preserves nominal
domains; eight valid mutations are prepared for later contract-rejection tests. This is preparation,
not full HTTP/world acceptance. Wasmtime 48.0.1's actual Rust HTTPp2 provider is
the independent reference, with its default send implementation retained.

Measure fields, persistent HTTP requests and streaming separately. Report load,
compile, provider/Store setup, bind, first request, warm requests, connection
startup and teardown. Run source base/head and pinned competitors serially on
stable source, preserve failures and expose incomplete performance targets.
Cross-language/world conformance and canonical async remain separate gates.

Pinned primary reference sources are
[HTTPp2](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/p2/mod.rs),
[resource types](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/p2/types.rs)
and [field storage](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/field_map.rs).
