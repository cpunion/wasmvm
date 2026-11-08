# Private memory forwarding: target qualification

The production selector at `e7660b0` forwards a contained integer load from an
immediately preceding private-memory store on **amd64 only**. The checked store
and every original frame copy remain; interior continuations retain checked
loads. Shared/multi-memory and metered execution keep their existing paths.
The copy-only scan is bounded to eight original instructions and cannot cross
a poll, helper, guard, frame-alias or register-allocation boundary. Source homes
shared by live values are excluded. No unchecked execution mode is introduced.

## Target decision

Two independent Apple M4 Max pilots reject enabling this on arm64:

| Source | ARM64 lowering | LLGo native head/base | Go native head/base | A/A resolved |
| --- | --- | ---: | ---: | --- |
| `8dc5d27` | Scratch load, shift, extend and result copy | 1.247 | 1.234 | 6/6 workloads |
| `fe9135f` | Direct bitfield extraction between allocated homes | 1.031 | 1.034 | 1/6; affected case unresolved |

The latter reduces the affected image by four bytes but establishes no gain.
Fewer memory reads or fewer bytes alone do not prove faster execution. Moving
extraction onto a value dependency is a possible explanation; these events do
not isolate a microarchitectural cause. Both pilots and every sample remain in
[the raw archive](data/private-memory-forwarding-2026-11-08.json.xz).

The final arm64 selector retains ordinary loads. Actual published Go/LLGo images
are identical to each other and to the base for all six affected workloads.
The separate final-source M4 timing event resolves 2/6 workload controls;
it is retained as an unchanged-code comparison, not an arm64 speedup.

## AMD64 pilot and remaining gaps

The first independent Linux event measures `8dc5d27` against main `d386c58`.
It observes the affected LLGo/Go native head/base ratios at 0.938/0.937.
That case passes advisory A/A calibration; 2/6 workload controls resolve.
The final selector retains this amd64 lowering and excludes the failed arm64
variants. Exact final-source Linux qualification is recorded separately in the
contribution PR; this pilot is not relabeled as a final-head measurement.

Ratios below use LLGo native as numerator; below 1 is faster. Rows are independent
workloads on the same scoped Linux event. Native C is an unchecked reference.

| Workload | Head/base | / Go native | / wazero native | / Wasmtime | / Wasmer | / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| memory/narrow-unaligned | 0.938 | 1.003 | 0.814 | 1.186 | 1.182 | 2.092 | resolved |
| memory/dependent-address | 0.999 | 0.992 | 0.971 | 1.291 | 1.289 | 1.512 | unresolved |
| memory/float64-unaligned | 0.998 | 1.007 | 0.641 | 1.053 | 1.056 | 1.076 | unresolved |
| runtime/simd-memory-full | 0.999 | 0.995 | 0.878 | 1.049 | 1.040 | 0.805 | resolved |
| runtime/simd-memory-widen | 0.999 | 0.997 | 0.622 | 1.390 | 1.351 | 0.362 | unresolved |
| runtime/simd-memory-lanes | 0.999 | 0.995 | 1.022 | 1.261 | 1.258 | 1.440 | unresolved |

## Reproduction and validation

Go 1.27.3, LLGo `38de1e218b8146ba73f9c6f2e60e3c31f29d96b6`, LLVM 22.1.8
and LLGo `-O3 -lto=full` are fixed. References remain wazero 1.12.0, Wasmtime
48.0.1 and Wasmer 7.3.0 native backends, plus wazero and WasmEdge 0.17.1
interpreters. Each six-case event retains 102 alternative-input checks, 510
ordinary/A/A observations and 30 native-C observations. Each process has two
warmups and seven interleaved measured calls; five processes are the independent
units. Preparation, first call, warm call and cleanup keep separate API timers.
No startup or timer cost is subtracted, no sample is discarded and no host or
source event is pooled. Baseline/reference binaries are hash-verified and reused
for the final local event; both candidate host-compiler binaries are rebuilt.

Local Go/LLGo codegen, native boundary, execution and VM tests pass. Local runs
compile both targets and execute arm64; final CI must execute the amd64
optimization under both host compilers. Tests cover
all integer widths and contained offsets, signed extensions, i32/i64 addresses,
aliased destinations, spilled operands, original cold PCs, full frame words,
unchanged fuel/budget and trap locations. Barrier tests reject shared memories,
clobbered copies, overlapping accesses and poll/allocation boundaries. Local
codegen statement coverage is 96.7%; standalone package scopes differ from the
full cross-package CI gate. Final integration still requires the complete
Go/LLGo Linux amd64/macOS arm64 semantic matrix and all 30 coverage gates.

The first Linux attempt timed out downloading the pinned Wasmtime C SDK before
any measurement. Its log remains; the same failed job reran without changing
source, tools or samples. Successful build/measurement jobs take 404/392 seconds.
The workflow now selects individual routing groups without changing measurement
protocols or the semantic matrix. GitHub-hosted target runners are available;
a zero self-hosted-runner inventory is not an execution prerequisite.

All data are advisory. Five-process controls use the existing five-percentage-
point interval-width rule; they do not satisfy the narrower, at-least-20-pair
controlled milestone contract. Competitor/native-C deficits, M2 quantitative
acceptance and production-release requirements remain open. The
[machine index](data/private-memory-forwarding-2026-11-08.json) pins all four events.
