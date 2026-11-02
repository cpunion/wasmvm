# Native shared memory and atomics: 2026-11-02

Shared reads/writes and the complete Core Threads atomic instruction family now
use owned native code on amd64 and arm64. Wait registration, positive/infinite
blocking and nonempty notification retain checked queue helpers. The measured
policy keeps bounds, alignment, cancellation and exact optional fuel semantics.
Controlled full-portfolio performance acceptance remains open.

## Revisions and measurement

Base `56dddbd32d139d93e902b9272a63eb740dd71a77` is merged PR #93. Measured head
`b59618631ea6d8e12cf23522aaf122b8f52263b2` includes the implementation; later
documentation does not change production code. The initial native implementation
`258f3d7eeafa8053f27d8656d2425cb5accddb53` remains a separate diagnostic.

Both revisions are archived and independently built using Go 1.27.3 and pinned
LLGo `8e45ed5`, with the same `-O3 -lto=full` release profile. The host is Apple
M4 Max/macOS arm64. Runtime CPU probing selects LSE where available; baseline
arm64 LL/SC and amd64 encodings remain independently exercised. References stay
Wasmtime 48.0.1 Cranelift, WasmEdge 0.17.1 interpreter, wazero 1.12.0 (compiler
and interpreter) and Wasmer 7.3.0 Cranelift. Binary and live SDK hashes are checked.

The unchanged six private/shared/atomic/wait/notify fixtures use three fresh
interleaved process blocks, two warmups and seven warm calls at 10,000 iterations.
Every call checks the independent oracle, with alternate seed/count checks.
Startup, preparation, instantiation, first call, warm calls and cleanup remain
separate. No clock/control/setup timing is subtracted. These workstation results
are descriptive and do not establish calibrated superiority.

| 10k iterations, ns/iteration | LLGo interpreter base | Head | LLGo native base | Head | Wasmtime native | WasmEdge interpreter |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 141.758 | 97.062 | 1.887 | 1.712 | 1.600 | 42.100 |
| threads/shared-memory | 174.646 | 108.862 | 871.325 | 2.538 | 1.600 | 40.800 |
| threads/atomic-rmw | 106.933 | 106.642 | 438.575 | 3.471 | 2.400 | 35.500 |
| threads/notify-empty | 128.604 | 101.571 | 493.679 | 2.871 | 8.200 | 45.800 |
| threads/wait-mismatch | 132.892 | 120.787 | 491.296 | 2.683 | 10.400 | 48.800 |
| threads/wait-timeout | 130.108 | 124.496 | 459.825 | 2.612 | 48.800 | 107.700 |

The separately declared sustained pass uses the same binaries and all six
fixtures at 1,000,000 iterations, with a recomputed independent oracle and the
same three-process/seven-call interleave. These longer loops resolve native
throughput separately from the short ordinary samples.

| 1M iterations, ns/iteration | LLGo native base | Head | Wasmtime | Wasmer | wazero compiler | WasmEdge interpreter |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.379 | 0.382 | 1.172 | 1.183 | 1.416 | 45.748 |
| threads/shared-memory | 987.014 | 1.011 | 1.219 | 1.192 | 1.322 | 46.455 |
| threads/atomic-rmw | 499.739 | 2.160 | 2.214 | 4.590 | 2.460 | 38.570 |
| threads/notify-empty | 480.532 | 1.076 | 7.823 | 5.142 | 23.637 | 46.185 |
| threads/wait-mismatch | 438.942 | 1.121 | 9.418 | 2.503 | 25.612 | 46.311 |
| threads/wait-timeout | 432.871 | 0.982 | 44.784 | 176.495 | 253.802 | 99.145 |

The LLGo shared interpreter improves from 174.646 to 108.862 ns/iteration;
atomic RMW is approximately unchanged at 106.933 versus 106.642. Wait/notify
also improve. Go shared access and atomic RMW regress from 74.208/64.129 to
106.379/84.746: the defined Wasm overlap contract now crosses an opaque ISA
adapter instead of ordinary Go byte accesses. This remains a measured cost,
not a waived regression. Both interpreters retain gaps against WasmEdge.

The intermediate `e07f036da76047a76d2edb87ecfc715ed6fdcd31` results remain
in the artifact. Their LLGo interpreter regressions motivated isolating lock/
defer paths and avoiding generated Cgo pointer-check closures for addresses of
retained native allocations. Host-buffer transfers retain pointer parameters
and Go liveness/pinning checks; no Go heap address uses the numeric boundary.
Every Go row and all five pinned reference configurations remain in the artifact.

## Execution and ownership

Native-backed shared memories reserve zeroed storage up to their existing host
quota/declared maximum. Growth does not move the base. It atomically publishes a
new logical byte length; the physical reservation is never the accessible guest
range. The tiny native control block contains fixed-width size and pending/
registered-waiter cells. Last-owner retirement waits for all bound/active native
contexts before freeing storage and control cells. Memory64 and one-byte pages
use the same checked logical-size contract.

The private runtime-data view now reserves a length/base pair for every memory,
including a single memory. A shared length word is the address of its stable
size cell. Code reloads that cell on each bounds observation. ARM metadata
registers may cache the immutable cell/base addresses, not the mutable size or
waiter count. Growth after a synchronized signal is therefore visible within
an already running entry. The external entry-record ABI is unchanged.

All scalar load/store/RMW/CAS widths and fence operations emit directly.
amd64 uses TSO loads, exchange stores, locked XADD/CMPXCHG and MFENCE; arm64
uses acquire/release instructions and polled exclusive loops, with LSE selected
only after a successful CPU capability probe. Full widened overflow/range checks
precede alignment checks and all writes. Subword operands/results truncate and
zero-extend correctly, including aliased output slots and large frames. Precise
fuel still charges one original instruction, independent of CAS retries.

Host/interpreter shared accesses use opaque target instructions too. Wasm allows
mixed-size overlaps and defined racy behavior; LLVM atomic objects require
perfectly overlapping/disjoint atomic ranges. Ordinary C/Go data accesses cannot
silently inherit that stronger assumption. Only separate, consistently u64
control cells use ordinary host atomic facilities. Sources:
[LLVM atomic constraints](https://www.llvm.org/docs/LangRef.html#atomic-memory-ordering-constraints),
[Wasm relaxed memory model](https://webassembly.github.io/threads/core/bikeshed/index.html#relaxed-memory-model).
This is an ISA access adapter, not another Wasm execution engine. Atomic helper
retries are bounded into noncommitting batches with Go cancellation/scheduling
checks; native retry loops poll the existing control word.

Empty notification, wait mismatch and zero timeout use checked native paths.
Real blocking/notification uses the existing domain queue and quotas. The native
waiter cell includes pending registrations **before** their SC comparison, as
well as queued waiters. A notifier observing a nonzero reservation takes the
ordering-lock helper; this closes the comparison-to-enqueue lost-wakeup window.
Timer/cancel/notification removal decrements the cell once. Queue/timer captures
are isolated from the ordinary atomic helper, avoiding pinned LLGo allocations.

## Validation and remaining work

Go and pinned LLGo execute the complete core/proposal inventories. New tests
cover both target encodings, every atomic width/group, LSE and LL/SC results,
subword high bits, aliased inputs, large frames, bounds/alignment trap order,
code budgets, host transfer canaries/overlap and concurrent unique RMW results.
Integration tests keep multiple native entries active on one backing, grow
inside a running entry, close creator roots, race notification against waiter
registration, verify quotas/cleanup and cancel an atomic loop. Portable no-cgo
execution remains complete through its synchronized Go backing.

The first macOS CI execution reported an unknown WASI network error in the
unchanged `p2_tcp_busy_poll` consumer. Thirty ordinary repeats and thirty
signal-stress repeats did not reproduce that particular failure. Deterministic
fault injection separately reproduced two matching error paths: an interrupted
readiness snapshot and an interrupted nonblocking connect after it started.
Readiness now retries `EINTR`; a started connect remains pending instead of
becoming a permanent failure. Regression tests fail against the previous logic
and preserve real readiness, descriptor errors and actual connection completion
on both backends. This WASI correction is separate from the measured core source
above. The original CI failure remains recorded; final CI is still required.

A supplementary Linux amd64 emulation run uses exact source `f67b1ad` and the
Go 1.27.3 image pinned by digest. All six affected packages execute successfully,
including complete core/proposal inventories and native atomic concurrency.
This is target execution under emulation, not native-host timing evidence.
CI shards now read the same checked-in manifest independently, so a queued Linux
planning job does not block creating macOS semantic jobs. The shared package and
matrix completeness checks remain mandatory.

Final CI must retain every Go/LLGo semantic package on Linux amd64 and macOS
arm64, shared race/quality checks and all 30 coverage thresholds. The full
controlled portfolio, contended scaling, broader host/component boundaries and
remaining interpreter deficits stay open. Complete observations and prior
diagnostics are retained in the [evidence](evidence/performance-native-threads-2026-11-02.json).
