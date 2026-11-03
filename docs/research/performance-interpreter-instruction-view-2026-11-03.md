# Immutable interpreter instruction views: 2026-11-03

Status: source-changing local experiment; complete affected semantics and diagnostic matrix pass. New exact-head Linux/macOS CI is required. Performance acceptance remains open.

The [cold-operand CI](pr112-cold-operands-ci-2026-11-03.md) still records Linux
LLGo scalar interpretation +6.4–12.1% and Go +6.2–10.4%. Restoring `a.Lo` register
storage alone did not resolve the regression. No previous outcome is replaced
or retimed unchanged.

## Change and actual amd64 code

Borrow the immutable instruction record from a frame-local code slice instead
of copying its eight fields at every step. The selected handler reads its actual
fields. Frame changes reacquire the slice; stack instructions keep their existing
value argument. No instruction format, unsafe pointer, feature flag or API changes.
Fuel, cancellation polls, traps, checked memory, GC roots/PC publication and native
helper execution keep their original boundaries.

Exact-source linked amd64 assembly shows the previous eager eight-field loads
removed. The slice base/length is captured when entering the guest frame. The
pinned Go 1.27.3/LLGo `983465e`/LLVM 22.1.8 builds use the production full-LTO
profile. The amd64 container is emulated on an ARM64 host and supplies only build
and disassembly evidence; no emulated throughput is measured.

## Complete local diagnostics

Base `b34e74bcce8e0ae607ca31a89977fc1d217ea9e7` is PR111's implementation;
new source `12599967425d1085039ee00aafd6b31dd049284b`. Same production host
compilers/profile, Apple M4 Max/macOS arm64, original twenty workload counts,
three fresh interleaved process blocks, two warmups and seven warm calls. The
independently verified complete grid contains 864 VM/reference and 27 ordinary-C
observations, 288 alternate-input checks, twelve explicit unsupported GC reference
outcomes and 297 summaries. Every measured phase and adverse observation is
retained without subtraction or pooling with earlier campaigns.

**Shared-host diagnostics, not controlled acceptance.** Host load before/after
is 153.10/19.23 (one-minute averages). Identical interpreter A/A aliases vary
substantially in several short cases: LLGo direct GC -30.5%, allocation +11.2%,
indirect GC +12.6%, and notify -17.6%; Go notify +21.6%. This limits inference;
it does not waive any adverse result or prove its cause. Earlier source identities'
long controls do not validate this new source.

Warm ns/source iteration, except empty ns/call:

| Kernel | Go base → head | LLGo base → head | WasmEdge | wazero |
| --- | ---: | ---: | ---: | ---: |
| threads/private-memory | 65.651 → 64.573 | 54.856 → 52.964 | 52.691 | 84.409 |
| core/call-control | 3875.000 → 4709.000 | 3958.000 → 5250.000 | 5000.000 | 2125.000 |
| core/integer-loop | 50.495 → 52.515 | 44.852 → 42.475 | 61.917 | 105.595 |
| core/function-calls | 64.537 → 61.967 | 59.757 → 57.411 | 62.496 | 83.748 |
| core/memory-load-store | 85.880 → 82.887 | 74.486 → 69.974 | 91.075 | 145.691 |
| memory/narrow-unaligned | 110.296 → 107.237 | 98.402 → 92.348 | 112.881 | 182.168 |
| memory/dependent-address | 54.878 → 53.651 | 48.167 → 44.524 | 53.396 | 98.212 |
| memory/float64-unaligned | 87.416 → 88.453 | 78.380 → 73.274 | 85.158 | 135.621 |
| runtime/simd-memory-full | 126.969 → 109.828 | 113.043 → 89.560 | 91.761 | 180.357 |
| runtime/simd-memory-widen | 86.651 → 79.993 | 86.666 → 67.655 | 62.688 | 150.811 |
| runtime/simd-memory-lanes | 130.699 → 115.104 | 124.903 → 93.804 | 90.398 | 185.293 |
| gc/i31-roundtrip | 68.683 → 62.779 | 53.717 → 47.471 | 57.700 | — |
| gc/struct-access | 80.596 → 88.862 | 94.067 → 77.896 | 76.400 | — |
| gc/struct-allocation | 167.429 → 157.662 | 194.354 → 184.846 | 150.100 | — |
| gc/struct-indirect-access | 144.442 → 121.542 | 118.500 → 114.013 | 124.000 | — |
| threads/shared-memory | 115.310 → 110.738 | 58.184 → 55.583 | 52.824 | 82.543 |
| threads/atomic-rmw | 88.499 → 82.680 | 60.459 → 53.149 | 42.308 | 57.487 |
| threads/notify-empty | 63.837 → 80.154 | 69.442 → 68.963 | 87.500 | 64.992 |
| threads/wait-mismatch | 88.446 → 101.888 | 65.496 → 66.342 | 53.700 | 89.675 |
| threads/wait-timeout | 101.292 → 91.071 | 72.600 → 62.467 | 134.800 | 464.671 |

Observed LLGo SIMD improves 20.8–24.9%; full-vector memory leads this WasmEdge,
while widening and lanes remain about 1.079x/1.038x its time. Eighteen of nineteen
sustained LLGo interpreter cases improve; wait-mismatch increases 1.3%. Retain Go
integer +4.0%, float memory +1.2%, direct GC +10.3%, notify +25.6% and
wait-mismatch +15.2%. Empty calls increase Go/LLGo 21.5%/32.6%.

## Native controls and remaining gaps

Retain Go native private memory +7.8%, float memory +2.5%, direct GC +88.6%,
allocation +25.4%, indirect GC +24.7% and empty +13.3%. LLGo native allocation
+5.3%, wait-mismatch +18.8%, RMW +2.0%, lanes +0.95% and empty +8.3% also remain.
The complete native/reference/C table is retained in the evidence. These outcomes
are not replaced by earlier campaigns, a different count or a favorable reference.

LLGo native scalar memory is 0.438 versus ordinary C 0.272 ns/iteration (1.608x);
narrow memory 0.687 versus 0.454 (1.513x); dependent memory 1.834 versus 1.418
(1.293x); lanes 5.059 versus 4.479 (1.129x). C is explicitly unchecked and does
not implement VM contracts. Native direct/indirect GC are about 2.00x/2.22x this
Wasmtime. Public-call, startup, allocation, cancellation, full WASI-world and
controlled portfolio acceptance remain open.

Both production Go and full-LTO LLGo pass the complete affected SIMD/execute/VM
suites and all 48 expected reports each. All five mandatory extensions execute
three tiers, including forced Stack Switching GC. Unchanged C/harness tests and
previous full CI remain historical evidence, not new full-31-package CI for this
source. New final-head 31-package/dual-platform semantics, thirty coverage gates
and Linux performance evidence remain required before merging. No duplicate
LLGo #2983 implementation is included; another contributor owns that issue.

[Complete original observations, source/toolchain identities, affected semantic
reports, comparison and actual amd64 assembly](evidence/performance-interpreter-instruction-view-2026-11-03.json.gz).
