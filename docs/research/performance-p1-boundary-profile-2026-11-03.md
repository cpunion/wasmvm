# P1 host-boundary stack and mechanism audit: 2026-11-03

These are qualitative macOS sample captures from selected production source
`6a13a48`, excluding the held FFI experiment. The ready handshake precedes
sampling, but captures begin during the first long call, possibly continuing
into warmup. They are not warmed-only CPU profiles. Go/LLGo/wazero use different
counts (five million/two million/one hundred million) to remain active for a
three-second capture. Stack occurrences, idle/background threads and unknown
native frames are not converted to per-function time or cross-runtime CPU
percentages. Primary performance studies keep independent timed processes.

The LLGo main-thread stacks repeatedly pass through Process.begin,
context.WithCancel, AllocU and BDWGC allocation/collection, plus VM host-value
encoding, capability operations and native entry. Go captures the same VM/P1
framework. This supports investigating repeated host-boundary and context work;
it does not assign the entire Go/LLGo difference to GC or establish a new
compiler defect. #2958, #2959 and #2974 retain their independent reproductions.

Pinned wazero 1.12.0 has a shorter provider path:

- [Host dispatch](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/engine/wazevo/call_engine.go#L415) passes an existing numeric stack view to the Go host function.
- [The WASI wrapper](https://github.com/tetratelabs/wazero/blob/v1.12.0/imports/wasi_snapshot_preview1/wasi.go#L293) writes errno back into that stack.
- [args-sizes](https://github.com/tetratelabs/wazero/blob/v1.12.0/imports/wasi_snapshot_preview1/args.go#L83) reads immutable process metadata and writes two checked integers. [Argument byte size](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/sys/sys.go#L44) is precomputed; [memory writes](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/wasm/memory.go#L359) still check bounds.

In contrast, our generic public callback path creates rooted public values and
an independently expiring Caller, and P1 admits each invocation and merges two
cancellation bindings. Those are explicit lifecycle contracts; removing them
without an equivalent shared protocol would change behavior. The structural
comparison motivates numeric argument/result marshaling and cheaper admission,
not a benchmark-specific import shortcut or a competitor fallback.

The audited reference harness uses default RuntimeConfig without
WithCloseOnContextDone. [The separate policy comparison](performance-p1-cancellation-policy-2026-11-03.md)
retains both reference modes. Enabled wazero [inserts a cancellation trampoline
at loop headers](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/engine/wazevo/frontend/lower.go#L1355)
and [returns to Go to check module closure](https://github.com/tetratelabs/wazero/blob/v1.12.0/internal/engine/wazevo/call_engine.go#L473).
Source inspection explains a mechanism, not an attributed count of nanoseconds.
Our own bounded polls and its loop checks have different granularity; matching
cancellation-latency acceptance remains required.

[All captures, commands, exact binary identities, parser correction and source hashes](evidence/performance-p1-boundary-profile-2026-11-03.json.gz)
