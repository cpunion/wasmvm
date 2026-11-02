# Interpreter atomics and LLGo C-call allocation: 2026-11-02

The final local candidate reduces LLGo interpreted shared memory from
**73.170 to 57.821 ns/source iteration** and atomic RMW from **86.293 to 67.626**.
Go RMW remains 75.683→75.816 after retaining its standard cancellation path.
LLGo now beats Go on both affected loops in this sample, but still trails
WasmEdge and trails wazero on RMW. Controlled cross-platform and portfolio
performance acceptance remain open. This candidate is not yet integrated.

## Mechanism and independently reproduced LLGo defects

`Memory.Atomic` and its bounded native retry helper both check cancellation.
The pinned LLGo `context.Err` implementation allocates on the uncancelled path
([#2959](https://github.com/xgo-dev/llgo/issues/2959#issuecomment-6566786471)).
An immutable execution-bound context view caches Done and delegates Err only
after cancellation. Deadline, Value, Cause and retained contexts across Rebind
remain correct. Err-only contexts retain their original behavior. The view is
LLGo-specific: Go already has an allocation-free, cheaper standard Err path.
No polling point, error precedence, retry bound or cancellation check is removed.

A separate minimal Cgo reproduction has only integer parameters and either a
scalar or a two-word struct result. Over 1,000,000 calls, Go allocates 0 B for
both; pinned LLGo allocates approximately 16 MB for each. Equivalent pure-Go
results have no comparable per-call pattern (4,096 B total). Optimized assembly
in both C-call wrappers calls AllocU(8), zeroes the allocation, then overwrites
the pointer before the indirect C call. This narrows an already recorded
residual allocation; it is not evidence that struct returns alone cause it.
The complete reproduction was added to
[LLGo #2958](https://github.com/xgo-dev/llgo/issues/2958#issuecomment-6568684544),
without opening a duplicate. Current upstream main was not rebuilt.

For retained C-owned guest memory, the LLGo adapter now uses its supported
direct C declarations for shared load/store, atomic access, fill and fence.
Only numeric arguments cross these bindings; the C ABI and atomic algorithms
are unchanged. Go keeps its regular Cgo adapter. Host-buffer transfers keep
their normal pointer checking/liveness paths. These are small compiler-specific
FFI bindings over the same owned runtime, not a second execution framework.

Allocation-only diagnostics, 100,000 native-backed helper operations with
checked results (cold context/view creation excluded):

| Operation | Raw cancellable context, old Cgo | Cached context, old Cgo | Cached context, direct C |
| --- | ---: | ---: | ---: |
| LLGo RMW | 8,000,304 B | 1,600,752 B | 0 B |
| LLGo notify, no waiters | 3,197,488 B | 0 B | 0 B |
| LLGo mismatching wait | 7,999,840 B | 1,598,336 B | 0 B |

Go reports 0 B in these controls. These totals are approximate host-collector
accounting, not a statement that whole VM calls or heap-backed synchronized
memory never allocate. One immutable view can allocate per LLGo binding;
setup, GC metadata and other execution paths remain separate.

## Final paired throughput and the three gaps

Base `ffd6300e596735d2c51181897ae92499fcde347b`; final measured source
`72d74d170068e59adbdcf8c6328c0a483c6d0f24`. Apple M4 Max/macOS arm64,
Go 1.27.3, LLGo `8e45ed5` with `-O3 -lto=full`. Wasmtime 48.0.1 and
Wasmer 7.3.0 use Cranelift; wazero 1.12.0 and WasmEdge 0.17.1 retain their
named compiler/interpreter configurations. Fixture, implementation, executable
and SDK-library hashes are checked. Independent and alternate-input oracles
validate every configuration.

Six unchanged cases run three fresh serially interleaved process blocks, two
warmups and seven calls. Loops use 1,000,000 source iterations; the empty call
control uses exactly one call. Independent A/A aliases execute the exact
baseline Go/LLGo binaries in the same blocks. No median subtraction is used.
Load, setup, preparation, instantiation, first call, warm calls and cleanup
remain separate. Units below are ns/source iteration, except the empty control
(ns/call).

| Workload | Go base | Head | LLGo base | Head | WasmEdge interp | wazero interp |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core/call-control | 3042.000 | 2750.000 | 2750.000 | 2583.000 | 3000.000 | 1042.000 |
| core/integer-loop | 73.426 | 73.903 | 62.151 | 62.240 | 48.648 | 81.707 |
| core/memory-load-store | 141.254 | 141.099 | 118.621 | 119.868 | 90.707 | 142.544 |
| threads/private-memory | 66.483 | 66.376 | 56.336 | 56.795 | 38.431 | 61.619 |
| threads/shared-memory | 100.634 | 100.690 | 73.170 | 57.821 | 38.412 | 62.012 |
| threads/atomic-rmw | 75.683 | 75.816 | 86.293 | 67.626 | 31.913 | 48.357 |

The affected shared/RMW improvements are approximately 21%/22%. RMW changes
LLGo from slower than Go to about 11% faster; it remains approximately 2.12x
WasmEdge and 1.40x wazero. Shared memory beats wazero by about 7% but remains
1.51x WasmEdge. LLGo scalar memory regresses 118.621→119.868 (about 1%);
private memory regresses 56.336→56.795. Neither is hidden.

Native execution and ordinary C, ns/source iteration:

| Workload | Go native | LLGo native | C/Clang | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| core/integer-loop | 1.574 | 1.559 | 1.545 | 1.550 | 1.552 |
| core/memory-load-store | 0.535 | 0.517 | 0.269 | 0.476 | 0.538 |
| threads/private-memory | 0.339 | 0.331 | — | 1.002 | 1.030 |
| threads/shared-memory | 0.859 | 0.850 | — | 0.698 | 0.998 |
| threads/atomic-rmw | 1.838 | 1.817 | — | 1.799 | 3.864 |

The separate C implementation uses Clang -O3 without fast-math, with checked
checksums and alias-safe unaligned memory access. It has fewer execution
contracts; absent VM setup is not assigned invented zero timings. LLGo native
integer execution is approximately 1.009x C here, while scalar memory is
1.92x C and 1.09x Wasmtime. Native shared memory remains behind Wasmtime.
The bindings primarily change interpreter/helper overhead, not generated guest
machine code; near-C integer execution does not complete native acceptance.

## Regressions, staging and uncertainty

The first context-view candidate `79a1e8a` ran a complete twelve-kernel paired
campaign, including all Go/LLGo interpreter/native and five reference modes
plus six applicable C controls. It improved LLGo RMW 124.018→109.944, but
regressed Go RMW 108.782→119.009. Go SIMD full memory was approximately flat;
LLGo SIMD full memory regressed 177.601→186.041. That complete campaign, all
regressions and all phases remain in the evidence. The Go constructor branch
was then restored to its original standard context path; final typed-FFI
measurement covers the six declared affected/control cases, not a new whole
portfolio acceptance claim. Intermediate `2a37ba3` was built but not timed.

The first campaign's host load dropped from about 126 to 9 on 16 cores. The
final targeted campaign ran near load 8–9; neither host is controlled. Final
compute-loop baseline aliases differ by at most roughly 1.6% in these samples,
but these controls do not define a calibrated acceptance threshold. Empty-call
controls vary much more (LLGo baseline alias 3,208 vs baseline 2,750 ns).
Those short controls are retained and are not used to remove a boundary cost
from compute loops. All comparisons remain descriptive.

## Validation and retained evidence

At the final source: all 31 Go semantic packages, all 30 cross-package coverage
gates (minimum 90.73%), pinned LLGo platform/memory/execute/VM suites, scoped
Go race, no-cgo and vet pass. Context tests cover metadata, cancellation and
immutable retained bindings. Existing atomics tests cover exact return values,
all widths, bit patterns, compare-exchange, overlapping accesses and concurrent
RMW. Final-head Linux amd64/macOS arm64 CI and wider portfolio evaluation remain
required before claiming integration or performance acceptance.

[Lossless compressed evidence](evidence/performance-interpreter-atomics-2026-11-02.json.gz)
retains both complete campaigns, same-binary controls, all phase samples,
source/build/toolchain/SDK hashes, measurement sources, both allocation
diagnostics, the minimal upstream reproduction/assembly and validation logs.
An initial package-only coverage command correctly failed the gates; subsequent
runs use the CI cross-package coverage scope and pass. Both outcomes remain
recorded. The pinned source root was recovered from an exact git archive after
the original worktree disappeared externally; compiler and binary identities
remain unchanged.

A separate [long native C atomic probe](performance-native-atomic-c-2026-11-02.md)
now supplies the missing ordinary-native RMW comparison. The [gap ledger and
competitor mechanism audit](performance-gaps-2026-11-02.md) keep compiler defects,
engine deficits and native comparisons distinct.
