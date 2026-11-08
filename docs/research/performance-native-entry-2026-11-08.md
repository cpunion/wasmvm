# Native entry investigation: 2026-11-08

This investigation follows the [complete integrated-main inventory](performance-main-inventory-2026-11-08.md).
Functional coverage is delivered for the declared scope; controlled performance
and production release remain open. Source, host and compiler events below are
independent. No timing subtraction, favorable retry or profile/ISA/count sweep
is used.

## Rejected ARM64 cache hypothesis

Experimental `b896381`, based on main `eac84b0`, reused the existing amd64
stable-origin priority analysis for equal-frequency ARM64 heap-cache ties.
Frequency still dominated; identity/index guards and the two-register budget
were unchanged. Targeted Go/LLGo tests passed and all 22 amd64 compiled/linked
images were unchanged. ARM64 Go/LLGo images agreed, with only indirect access
changing from 10,016 to 10,192 bytes.

The separate original/sustained GC cohorts observed indirect-access head/base
elapsed ratios of **1.307/1.351 under LLGo** and **1.230/1.361 under Go**. All
four-case controls were unresolved in each cohort. These adverse observations
were retained, the experiment was never pushed or merged, and its branch and
worktree were removed after archival. Stable reference identity alone does not
establish the most profitable cache target.

The complete local archive retains the patch, verified Git bundle, tests,
image checks and observations: `artifacts/arm-heap-reuse-20261108/evidence.tar.gz`,
3,251,455 bytes, 182 files, SHA-256
`f024330b5d03a4fa8e413e1464480ed3ed7aaf289e8d429d144cbb35d8571b9e`.

## Public call profiles

Two ordinary exact-main binaries use Go 1.27.3 and pinned LLGo `3571c25`
with `-O3 -lto=full` and LLVM 22.1.8. The artifact-only driver repeatedly calls
the public `Extern.Call` with a background context and checks every result.
External macOS `sample` runs for five seconds at a one-millisecond interval,
separately for each binary, after priming. Builds finish before sampling.

The profiles expose LLGo allocation/BDWGC activity through Store admission and
the public call stack. Disassembly identifies the inlined `context.WithCancel`
allocation at `Store.begin+860`; allocating a cancellation context is expected,
and this sample does not itself prove a new compiler defect. Existing
[defer](https://github.com/xgo-dev/llgo/issues/2932) and
[escape](https://github.com/xgo-dev/llgo/issues/2974) reports remain separate.
No lifecycle guards or required cancellation contexts are removed.

Both profiles also expose libc floating-environment operations at native entry.
Profiles include driver checks, scheduler and collector activity; stack samples
are not elapsed API benchmarks, allocation counts, or percentages of guest work.
Idle/collector threads must not be added to infer a measured guest time.

The pinned wazero source explains part of the boundary difference: wazevo
reuses its call engine and enters generated code through a private assembly
bridge. Ordinary `Call` allocates one parameter/result word slice;
`CallWithStack` can accept caller-owned words, but the retained reference uses
ordinary `Call`. Its default `ensureTermination` is false, and our reference
adapter does not enable `WithCloseOnContextDone`. The owned runtime retains
Store admission, a cancellation context, typed marshalling/root ownership and
the C ABI. These are concrete structural differences, not a measured allocation
or latency attribution. See the pinned
[call engine](https://github.com/wazero/wazero/blob/v1.12.0/internal/engine/wazevo/call_engine.go)
and [configuration](https://github.com/wazero/wazero/blob/v1.12.0/config.go).
Reference default policies and unchecked native C remain explicit; elapsed
ratios do not establish equal responsiveness, isolation or resource policies.

## ARM64 floating environment

Measured candidate `1d431f5`, based on `eac84b0`, replaces three libc environment
calls per ARM64 native entry with direct FPCR/FPSR snapshots and restoration.
The guest starts with zeroed registers: nearest-even, gradual underflow, masked
exceptions, cleared sticky/saturation flags. Both complete host registers are
restored after the guest returns. Instruction synchronization follows writes.
The amd64 MXCSR/x87 path, JIT write protection, generated guest code, resource
quotas, cancellation, roots, cold continuations and compiler profile stay intact.

ABI probes now exercise all 13 raw status values, invalid version/arena/depth
and pre-cancelled admission under a hostile host environment. ARM64 probes
snapshot both guest registers and contaminate guest state; host rounding,
flush/default-NaN modes, sticky exceptions and saturation must survive. The
full Go/LLGo/platform semantic matrix remains required before merging.

The independent local original22 event uses fresh exact base/head Go/LLGo
runners and every named reference, with executable, SDK and source hashes
checked. It retains five process blocks, two warmups, seven interleaved calls,
all alternate checks, complete phases and native C where applicable. Darwin
affinity and physical frequency/background isolation are unavailable; this is
advisory. Allocation statistics include pauses. A metadata-only preflight
failed because PATH omitted `/usr/sbin`; no observation existed, and the failure
is retained alongside the corrected PATH run. Timers, counts and reducers did
not change.


## Independent ARM64 observations

All 22 cases complete: 362 alternate configuration checks, 12 explicit GC
rejections, 1,810 ordinary/A/A process records and 55 native-C records. All
results/checksums pass. Only **2/22** complete control sets resolve. The raw
public-call medians observe LLGo native **3,333 → 2,417 ns** and Go native
**3,708 → 3,417 ns**. The head LLGo alias is 3,208 ns: this unresolved spread
prevents attributing the apparent 27% LLGo difference to this source change.
No measured speedup or broad parity is established, and no timing retry follows.
Static disassembly independently verifies that Go/LLGo entries contain direct
register snapshots/restores and no libc environment calls, retaining JIT
protection. Guest compiler sources are identical. Targeted Go/LLGo tests and
vet pass; standalone native adapter coverage is 420/443 (94.81%). Standalone
nativeexec unit coverage is 421/496 (84.88%); full cross-package CI, including
its public engine callers, must still clear the unchanged 90% gate. The mistaken
local coverage-scope assertion is retained and does not lower that gate.

Ratios below use LLGo and complete API intervals on this single Apple M4 event;
**lower is better**. Unsupported GC references are not timed. The primary
allocation statistic remains the process mean including pauses. These rows
must not be pooled with the separate Linux inventory or the rejected cache
experiment.

| Workload | Native head/base | Native / Go | / wazero | / Wasmtime | / Wasmer | Interpreter / Go | / wazero interpreter | / WasmEdge | Native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 0.725× | 0.707× | 5.796× | 2.523× | 1.208× | 1.015× | 3.286× | 1.113× | unavailable | inconclusive |
| core/integer-loop | 0.987× | 0.964× | 0.901× | 0.996× | 0.993× | 0.816× | 0.343× | 0.571× | 1.019× | inconclusive |
| core/function-calls | 1.002× | 0.991× | 0.551× | 0.884× | 1.037× | 0.859× | 0.534× | 0.771× | 1.043× | inconclusive |
| core/memory-load-store | 1.007× | 0.954× | 0.493× | 0.899× | 0.827× | 0.819× | 0.429× | 0.679× | 1.541× | inconclusive |
| memory/narrow-unaligned | 1.021× | 1.007× | 0.615× | 0.948× | 0.919× | 0.856× | 0.444× | 0.719× | 1.441× | inconclusive |
| memory/dependent-address | 1.011× | 0.991× | 0.582× | 0.908× | 0.907× | 0.863× | 0.400× | 0.738× | 1.311× | inconclusive |
| memory/float64-unaligned | 0.991× | 0.978× | 0.826× | 1.001× | 1.029× | 0.851× | 0.478× | 0.771× | 1.100× | inconclusive |
| runtime/simd-memory-full | 0.992× | 0.997× | 0.869× | 0.916× | 0.956× | 0.855× | 0.464× | 0.892× | 0.964× | inconclusive |
| runtime/simd-memory-widen | 1.019× | 0.990× | 0.914× | 1.002× | 1.007× | 0.911× | 0.418× | 0.988× | 1.022× | inconclusive |
| runtime/simd-memory-lanes | 0.985× | 0.985× | 0.821× | 0.857× | 0.846× | 0.882× | 0.472× | 0.968× | 1.095× | resolved, advisory |
| gc/i31-roundtrip | 1.020× | 0.946× | unavailable | 0.903× | unavailable | 0.828× | unavailable | 0.770× | unavailable | inconclusive |
| gc/struct-access | 1.041× | 0.900× | unavailable | 1.378× | unavailable | 0.879× | unavailable | 0.724× | unavailable | inconclusive |
| gc/struct-allocation | 0.974× | 0.885× | unavailable | 0.353× | unavailable | 1.365× | unavailable | 1.234× | unavailable | inconclusive |
| gc/struct-indirect-access | 0.981× | 0.899× | unavailable | 1.784× | unavailable | 0.870× | unavailable | 0.746× | unavailable | inconclusive |
| threads/private-memory | 0.906× | 0.900× | 1.221× | 1.239× | 1.110× | 0.844× | 0.602× | 0.962× | unavailable | inconclusive |
| threads/shared-memory | 0.953× | 0.706× | 1.052× | 1.184× | 1.095× | 0.477× | 0.628× | 0.999× | unavailable | inconclusive |
| threads/atomic-rmw | 0.965× | 0.930× | 1.087× | 1.110× | 0.530× | 0.634× | 0.778× | 1.084× | unavailable | inconclusive |
| threads/notify-empty | 0.994× | 0.937× | 0.068× | 0.196× | 0.273× | 0.863× | 0.748× | 0.773× | unavailable | inconclusive |
| threads/wait-mismatch | 1.023× | 0.872× | 0.059× | 0.162× | 0.518× | 0.652× | 0.597× | 0.944× | unavailable | inconclusive |
| threads/wait-timeout | 0.910× | 0.596× | 0.006× | 0.035× | 0.009× | 0.636× | 0.122× | 0.409× | unavailable | inconclusive |
| applications/sha256-clang | 1.004× | 0.984× | 0.681× | 1.055× | 1.014× | 0.867× | 0.440× | 0.689× | 1.218× | inconclusive |
| applications/sha256-rust | 0.993× | 0.984× | 0.666× | 1.142× | 1.127× | 0.866× | 0.430× | 0.726× | 1.270× | resolved, advisory |

Public-call latency, GC direct/indirect access, private/shared Threads memory,
atomic RMW, application throughput and native-C differences remain open. Some
ratios exceed one. Two calibrated cases cannot establish complete acceptance.
The [machine record](data/performance-native-entry-2026-11-08.json) retains all
observations, host/reference identities, phases, orders, checks, controls,
profiles, rejected-experiment data and failed preflight/coverage-scope evidence.
