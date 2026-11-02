# Interpreter dispatch and three-way gap audit: 2026-11-02

LLGo guest calls improve from **132.794 to 93.345 ns/source iteration** and
shared-memory interpretation from **94.846 to 71.317**. Go guest calls remain
97.924→97.964. The owned interpreter still trails WasmEdge; native memory loops
still trail ordinary C. **Competitive, cross-platform and full-portfolio
performance acceptance remain open.** No average conceals an individual deficit.

## Sources and measurement

Measured base `b8adaa6598177a13aca6e35efccee63e30f9c883` already contains the value-returning memory
optimization. Final `ffd6300e596735d2c51181897ae92499fcde347b` retains lazy source metadata, immediates
in the existing 32-byte slot instruction, direct writes into reusable frame
capacity, and cached cancellation channels. Original Wasm sources remain
available for exact fuel, trap offsets and native lowering. Bounds checks,
frame limits and the existing cancellation polling cadence remain intact.
Contexts with no Done channel retain direct Err polling. Workspace rebind and
quiescence replace/release both the context and its cached channel.

Apple M4 Max/macOS arm64, Go 1.27.3, pinned LLGo `8e45ed5` with `-O3 -lto=full`.
References remain Wasmtime 48.0.1 Cranelift, Wasmer 7.3.0 Cranelift,
wazero 1.12.0 compiler/interpreter, and WasmEdge 0.17.1 interpreter.
Live binary, archived implementation and SDK-library hashes were verified.
Every campaign uses three fresh serially interleaved process blocks, two
warmups and seven warm calls. Every result is checked against independent
oracles, including alternate seed/count checks. Startup, load, preparation,
instantiation, lookup, first call, warm execution and cleanup stay separate;
no setup or clock control is subtracted. These shared-workstation observations
are descriptive, without calibrated A/A or a statistical parity verdict.

## Interpreter and host-compiler gaps

Unchanged fixture counts; **ns/source iteration**:

| Workload | Go base | Head | LLGo base | Head | WasmEdge | wazero interp |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 68.233 | 68.100 | 60.733 | 58.017 | 39.400 | 62.100 |
| threads/shared-memory | 103.554 | 104.046 | 94.846 | 71.317 | 40.100 | 61.596 |
| threads/atomic-rmw | 77.562 | 76.371 | 102.071 | 94.025 | 32.500 | 43.858 |
| core/integer-loop | 76.581 | 73.983 | 73.395 | 63.152 | 49.104 | 81.622 |
| core/function-calls | 97.924 | 97.964 | 132.794 | 93.345 | 62.473 | 83.509 |
| core/memory-load-store | 150.727 | 145.305 | 134.373 | 120.712 | 92.820 | 145.336 |
| memory/narrow-unaligned | 183.631 | 176.724 | 164.610 | 149.387 | 113.157 | 185.008 |
| memory/dependent-address | 85.593 | 86.522 | 82.570 | 72.896 | 53.653 | 99.331 |
| memory/float64-unaligned | 139.633 | 136.745 | 126.596 | 116.074 | 84.251 | 136.187 |
| runtime/simd-memory-full | 150.546 | 148.458 | 134.242 | 129.713 | 79.300 | 152.617 |
| runtime/simd-memory-widen | 138.621 | 140.738 | 130.967 | 130.613 | 62.600 | 144.150 |
| runtime/simd-memory-lanes | 182.129 | 179.804 | 157.767 | 162.475 | 92.300 | 189.938 |

LLGo now approaches Go on direct guest calls and exceeds Go in several scalar
memory cases. It still loses to Go on atomic RMW (94.025 vs 76.371), and to
WasmEdge throughout these twelve cases. SIMD lanes regress 157.767→162.475;
Go dependent addresses regress 85.593→86.522, and widening 138.621→140.738.
These results remain visible; no selective subset is an acceptance result.

A separate three-second diagnostic sample of the preceding immediate-only
head places 114 samples under guest-frame push; 106 descend into runtime.AllocU.
Another cancellation-poll subtree includes 33 allocation samples. These are
sampled stack attributions, not exact allocation counts or elapsed-time shares;
profiling is excluded from all timing campaigns. Final frame assembly moves the
append temporary behind the actual-capacity-growth branch. The existing-capacity
path writes the same frame value directly. Growth and error paths still allocate.

Two independent compiler reproductions explain avoidable LLGo host GC work:

- [LLGo #2957](https://github.com/xgo-dev/llgo/issues/2957): scalar append creates
  temporary argument storage despite sufficient capacity. Frame reuse now
  avoids it without changing guest call/return semantics.
- [LLGo #2959, standard-library reproduction](https://github.com/xgo-dev/llgo/issues/2959#issuecomment-6566786471):
  1,000,000 uncancelled context.Err calls allocate 0 B with Go and 32,001,856 B
  with pinned LLGo; an empty atomic.Value.Load similarly allocates 32,002,080 B.
  Cached-channel polling allocates 0 B in that minimal reproduction, and still
  detects cancellation. Assembly allocates the 16-byte named return before the
  atomic nil branch. This supplements the existing return-escape issue rather
  than creating a duplicate. Current upstream main was not rebuilt.

The runtime exchanges one Done lookup per context binding for repeated hot Err
calls. It does not claim that all host allocations disappear. The independent
public empty-call control regresses Go interpretation 2,792→4,000 ns and LLGo
3,500→3,833 ns; native controls are 6,334→6,292 and 5,750→5,833 ns respectively.
All three blocks remain recorded. These short controls do not isolate the cost
of one change, and their times are never subtracted from compute kernels.

## Ordinary native C and named compiled competitors

Six unchanged scalar Wasm binaries use 1,000,000 iterations and recomputed
checksums. The separate [C implementation](../../benchmarks/native/c/scalar_baseline.c)
uses Clang -O3, no fast-math, and memcpy for unaligned/alias-safe values.
The C reference at `cc6ba8c` corrects an earlier translated +4096-byte base in
the scalar-memory kernel to the current fixture's zero base. The earlier
checksums agreed, but that entire campaign remains separately retained;
the table below comes from the corrected source and a fresh six-case campaign.
A volatile indirect entry prevents pure repeated calls from moving outside
clocks; assembly confirms the entry between clocks. All 36 release/sanitized
checks cover both alternate inputs and the full count with ASan/UBSan.
The timer includes required reset work. Native C has no Wasm validation,
bounds, fuel or cancellation contract; its absent VM setup phases are not
invented zeroes. First-call and warm times remain separate.

**Native execution, ns/source iteration**:

| Workload | Go | LLGo | C/Clang | Wasmtime | Wasmer | wazero |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core/integer-loop | 1.759 | 1.747 | 1.740 | 1.742 | 1.747 | 1.932 |
| core/function-calls | 0.867 | 0.840 | 0.834 | 0.972 | 0.837 | 1.462 |
| core/memory-load-store | 0.613 | 0.577 | 0.310 | 0.540 | 0.616 | 1.052 |
| memory/narrow-unaligned | 0.750 | 0.739 | 0.480 | 0.741 | 0.763 | 1.150 |
| memory/dependent-address | 1.939 | 1.928 | 1.473 | 2.151 | 2.175 | 3.241 |
| memory/float64-unaligned | 0.729 | 0.722 | 0.667 | 0.713 | 0.715 | 0.871 |

LLGo native/C ratios are 1.004x integer, 1.007x calls, 1.860x scalar memory,
1.540x narrow/unaligned, 1.309x dependent addresses and 1.082x float64.
Native integer and call kernels approach C in this sample; the three integer
memory kernels require further compiler work. Native source calls can be
inlined in both languages; these are source-loop costs, not machine-call latency.

An offline diagnostic compiles those three memory functions with the production
single-private-memory assumptions for both targets. It records zero runtime
helpers, with 2/3/1 proven hot-loop accesses respectively. Reset/setup checks
remain; per-access bounds checks in the main loops and interpreter fallback
do not explain their remaining work. arm64 assembly
still has a scalar iteration, address/mask moves, and bounded interrupt polls.
Clang unrolls scalar memory/dependent-address loops by four and uses separate
accumulators; narrow accesses use a strength-reduced address induction.
These are concrete optimization leads, not a measured attribution of every
nanosecond. Preserve continuation, trap, exact-fuel and cancellation contracts
when reducing redundant materialization or adding general loop transforms.

## Sustained native controls

Short ordinary samples showed apparent native regressions. The same exact
binaries were subsequently compared on five predeclared existing kernels at
1,000,000 iterations, independently recomputing their outputs. No native code
selection or acceptance margin changes between campaigns.

| Workload | Go base | Head | LLGo base | Head | Wasmtime | Wasmer | wazero |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 0.323 | 0.321 | 0.321 | 0.324 | 1.030 | 1.034 | 1.033 |
| threads/shared-memory | 0.854 | 0.852 | 0.836 | 0.841 | 1.027 | 0.994 | 1.035 |
| threads/atomic-rmw | 1.841 | 1.833 | 1.817 | 1.823 | 1.808 | 3.889 | 1.954 |
| runtime/simd-memory-full | 3.391 | 3.391 | 3.373 | 3.376 | 3.638 | 3.509 | 3.647 |
| runtime/simd-memory-widen | 1.050 | 1.054 | 1.048 | 1.038 | 1.043 | 1.036 | 1.040 |

These longer calls retain the original short observations rather than replacing
them. Near-equal values still do not establish controlled superiority. Native
shared/private memory and SIMD controls remain distinct from interpreter rows.

## Rejected work and validation

A separate experiment added slot-bound proofs and unchecked frame indexing
(`a993e3b4be5bc1af2f27e792b048255f0643a19e`). It produced little consistent LLGo
improvement and substantial variability, including a Go atomic regression.
It was reverted (`cec60eb`); its unchecked access and proof machinery are absent
from the retained implementation. The complete twelve-case measurements of
that experiment, metadata-only `be7494d`, immediate-only `29de9fd`, final head,
native-C comparison and sustained controls are retained separately in the
[evidence](evidence/performance-interpreter-dispatch-2026-11-02.json), together
with scripts, source manifests, binary/SDK identities, sampled stacks, assembly,
compiler reproduction, and initial missing-ack/import input failures.

Final Go semantics cover all 31 matrix packages; all 30 executable-package
coverage gates pass (minimum 90.66%), and vet passes. Go and pinned LLGo execute
the affected interpreter/public VM packages, including complete Core 2.0/3.0,
pinned proposals, precise fuel, cancellation, workspace rebinding, stack
switching and cleanup. Affected Go race and no-cgo suites pass. The required final-head
Linux amd64/macOS arm64 matrix remains pending; a local measurement
or successful build does not replace those final PR checks.

Next priorities are atomic/shared interpreter adapters, general dispatch and
SIMD throughput, then native loop instruction selection/unrolling and memory
addressing. GC/component/WASI and application-weighted portfolio gaps from the
[broader audit](performance-2026-11-01.md) remain open; this scalar campaign does
not claim to have requalified every family or target.
