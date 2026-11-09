# Reusable scalar public calls

Prepared after merged PR #141, baseline
`9f04e1a3785303532a8ab77e0f3ba383dd2e0b71`. Preserve one active contribution
PR. Final source qualification, affected API comparisons and the full semantic
matrix remain required before integration.

The typed Value API constructs raw argument and public result slices for every
invocation. `Extern.Scalar` captures a numeric signature once and keeps private
raw argument storage. `ScalarFunc.Call` receives a caller-owned uint64 stack,
copies all inputs before execution and replaces only the result prefix on
success. I32/F32 use their low32 bits; I64/F64 retain all bits. The stack length
is exactly max(parameter count, result count). References and V128 keep the
typed Value path. Signature checking does not select a workload or backend.

The implementation uses the existing Store operation, execution driver,
workspace, fuel, native polling, traps, callbacks and cleanup. It introduces no
operation/context pool, native ABI, register plan or unchecked execution mode.
Bindings serialize their scratch storage; Store admission remains authoritative
across bindings. Host callbacks use Caller.Call for controlled reentry. A host
panic releases both admission locks. Closed/canceled Store admission precedes
argument validation, matching the ordinary entry point. Neither input nor
result storage is exposed to the guest or retained as an opaque public root.

```go
call, err := function.Scalar()
if err != nil {
    return err
}
stack := []uint64{17, 64}
fuel, err := call.Call(ctx, stack, vm.CallOptions{})
// On success stack[0] contains the result; fuel is separate policy metadata.
```

Tests cover all four numeric types, NaN/sign bits,32-bit canonicalization,
overlapping input/results, repeated use, zero values, unsupported signatures,
exact fuel, traps and unmodified failure inputs, parent cancellation, closed
Stores, same/different-binding contention, host panic cleanup, controlled host
reentry and retained context expiry on both backends and host compilers.

## Measurement boundary

No timing result is claimed yet. A separate allocation-only diagnostic at
source49d02bd retains all three rounds and zero-call controls: Go ordinary
calls264B/four allocations versus stack calls200B/two; pinned LLGo ordinary
about976 versus stack about912 collector bytes. LLGo Mallocs is unavailable,
and collector bytes are not equivalent Go object accounting. These original
source records are not relabelled after the documentation-only rebase.
A minimal literal/field-assignment control finds no sustained per-call LLGo
allocation, so it does not justify a compiler issue or rewriting the ordinary
struct literal. Remaining allocation sites still require investigation.

Keep the ordinary typed API as a distinct
row; a faster numeric interface cannot erase an ordinary API regression. Compare
equivalent reusable numeric-stack interfaces in pinned references, explicitly
retaining their termination/metering policies. The existing10000-call batch
diagnostic separates readiness/preparation/first batch/warm/cleanup and preserves
Go/LLGo, parent-context and A/A series. Reuse it rather than inventing a second
execution framework. A baseline without the new API must be explicitly absent
for that API row, while its ordinary API is still built and measured; never
copy head engine code into a baseline to make it compile.

Finish every local build/test/probe/archive before one frozen-source event.
Keep caller-buffer resets and checks at matching boundaries, all samples and
adverse/unsupported outcomes, the pinned release profiles and original counts.
Do not subtract API/dispatch/entry medians or pool different interfaces. Inspect
allocation sites independently: service TotalAlloc also includes queue/timer
work and cannot be attributed solely to VM calls. Full performance and release
acceptance remain open.
