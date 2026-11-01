# GC interpreter operand and callback costs: 2026-11-02

The checked interpreter now passes ordinary GC operands directly and constructs
collection callbacks only when a caller needs a temporary root visitor. This
removes recurring host work without changing guest GC, fuel or trap semantics.
Competitive performance acceptance remains open.

## Revisions and method

Base `d07be4b7940a12f7d3239cb3cf91daffbbcdb20d` is the final PR #92 head;
its engine sources are identical to merged base
`237a03b327feab521a029dcc9ea9d7afab7b6ab3`. Measured head
`c175527323323d81dceceb5eba8fce08f19262a4` contains the implementation.
Subsequent documentation changes do not alter measured production sources.

The Apple M4 Max/macOS arm64 comparison independently builds both revisions
with Go 1.27.3 and LLGo `8e45ed5` using `-O3 -lto=full`. Reference pins remain
Wasmtime 48.0.1 Cranelift, WasmEdge 0.17.1 interpreter, wazero 1.12.0 and
Wasmer 7.3.0. Binary, fixture, compiler and live reference SDK identities are
retained and checked. Wazero/Wasmer GC rejections are unsupported outcomes,
without synthetic zero timings.

The unchanged four GC fixtures use 10,000 iterations, alternate seed/count
checks, three fresh interleaved process blocks, two warmups and seven warm
calls. The unchanged private-memory fixture is a control. Startup, preparation,
instantiation, first call, warm calls and cleanup remain separate; no setup or
empty-call timing is subtracted. All required safety/cancellation checks remain
enabled. These are descriptive shared-workstation observations.

| Interpreter, ns/source iteration | Go base | Go head | LLGo base | LLGo head | WasmEdge |
| --- | ---: | ---: | ---: | ---: | ---: |
| gc/i31-roundtrip | 84.283 | 73.404 | 552.229 | 61.979 | 44.800 |
| gc/struct-access | 102.750 | 89.379 | 591.100 | 78.021 | 54.900 |
| gc/struct-allocation | 157.571 | 145.725 | 584.913 | 168.396 | 106.600 |
| gc/struct-indirect-access | 172.042 | 146.379 | 1200.092 | 129.371 | 94.500 |
| threads/private-memory | 64.171 | 65.304 | 120.625 | 129.204 | 38.700 |

Allocation also uses a distinct predeclared diagnostic: 100,000 linked
allocations per call, 63 warm calls per process, three fresh interleaved blocks
and two warmups. Report the median of process means, including **every GC
pause**. Wasmtime's short-call medians omit substantial periodic collection
cost and cannot establish allocation throughput. Pinned C adapters retain
seven independent clock calibration probes; all 63 warm timings/results are
validated separately, with no padding or truncation.

| Interpreter, ns/object including all pauses | Base | Head |
| --- | ---: | ---: |
| wasmvm Go | 152.311 | 141.251 |
| wasmvm LLGo | 587.615 | 171.269 |
| WasmEdge | — | 104.156 |
| Wasmtime native reference | — | 60.671 |

## Mechanism and bounds

`GCAccessArity` classifies checked nonallocating struct/array access, length,
i31 and descriptor lookup. The interpreter consumes the already lowered first
two operands and reads a third array-store operand directly. `GCAccess` shares
the existing typed heap, full handle/generation checks, packed-field behavior,
v128 bits, null traps and array bounds. Constant expressions and native helper
exits continue to use the same semantics through `Store.GC`.

The slow path now calls ordinary helpers for packing, polling, object checks
and rooted allocation. It does not construct unrelated local closures for each
operation. The embedding Store's existing visitor remains authoritative;
internal callers without one get a temporary collector only for allocating
operations. Constructor reference operands remain temporarily rooted, including
on constant-expression stacks. A deferred method restores the borrowed root
extent on errors and collector panics.

Slow operands use an execution-owned, cleared scratch slice. Idle executions
retain at most 64 values (1 KiB); larger validated tuples remain transient and
are released after success, errors or abort. Demand remains bounded by compiled
tuple/node limits. Close releases the scratch. Scratch bits are not an
additional root domain; live guest references remain in typed frame/root maps.
Source PCs, exact fuel charges, polls and native helper boundaries are unchanged.

LLGo assembly from intermediate `ba72a05` shows repeated `runtime.AllocU` calls
before the generic helper's opcode branch, associated with local closure
construction. The refactor removes those closures. Earlier callback-only,
scalar-operand, scratch and helper experiments remain in the evidence. One
intermediate indexed-operand diagnostic overlapped the tail of a test run; it
is explicitly excluded from conclusions and retained for provenance.

## Validation and remaining work

Affected state, execution and embedding packages pass with both Go and pinned
LLGo, including the complete Core 2.0/3.0 and proposal execution inventories.
Additional tests alternate small/large tuples, force collection, exhaust heap
quotas, reuse an execution after failure and verify temporary-root cleanup after
a collector panic. Focused race checks pass. Final PR CI retains the complete
31-package Go/LLGo dual-platform matrix and all 30 coverage gates.

The interpreter still trails WasmEdge on these cases. Broader dispatch,
allocation/collection, arrays/casts, shared memory, host/component boundaries
and controlled amd64/arm64 portfolio acceptance remain open. The unchanged LLGo private-memory control
varies from 120.625 to 129.204 ns/iteration in this run; its existing deficit
and the uncalibrated variation remain recorded. Native short controls are near clock resolution and do not establish a native improvement.
See the [complete evidence](evidence/performance-gc-interpreter-2026-11-02.json).
