# Constrained loop registers: complete hosted comparison, 2026-11-05

**Hosted diagnostic; controlled performance acceptance remains open.**
[Run 41987225706](https://github.com/cpunion/wasmvm/actions/runs/41987225706) measures engine
`2f8e44f493a127b6e9b2f82400b220626e9accbe`, published at
`11a3a18bb5d28072b8dc6d726ba0aa2df885be59`, against the preceding
shared-memory source at `9f19606e7e765287c5c11c0558317bdd2a1c7ca2`.
The [implementation record](performance-loop-scratch-registers-2026-11-05.md)
describes the common native-loop/unroll register inventory and integer
complete-write selection. No workload-specific opcode, host compiler/ISA/alignment
configuration change or timing retry is used. All prior adverse cohorts remain retained.

All six jobs succeed. The audit verifies 22 workloads, 362 alternate correctness
configurations and 12 explicit reference GC rejections. Eight workloads resolve
all four diagnostic A/A controls; none establishes controlled acceptance.
The shared build takes 399s; serial measurement jobs take 160s (Core), 500s
(memory/SIMD), 35s (GC), 62s (Threads) and 209s (applications). Each is under
10 minutes. These routing groups are serial, not concurrent measurements.

## Warm execution and regressions

Every ratio is **elapsed time**, lower is better. LLGo native integer, source-call
and memory medians are 0.823x, 0.800x and 0.783x their own base. Clang/Rust SHA
medians are 0.792x/0.816x base. SHA still costs 1.554x/1.607x Wasmtime and
1.866x/1.900x unchecked C; static loop reductions have not closed that gap.
The Clang SHA controls are inconclusive; Rust SHA resolves the diagnostic controls.

| Workload | Go native head/base | LLGo native head/base | Go interpreter head/base | LLGo interpreter head/base |
| --- | ---: | ---: | ---: | ---: |
| core/call-control | 0.998x | 1.051x | 1.000x | 1.127x |
| core/integer-loop | 0.818x | 0.823x | 1.093x | 1.041x |
| core/function-calls | 0.799x | 0.800x | 1.001x | 0.993x |
| core/memory-load-store | 0.788x | 0.783x | 1.089x | 1.007x |
| applications/sha256-clang | 0.783x | 0.792x | 1.064x | 1.014x |
| applications/sha256-rust | 0.813x | 0.816x | 1.063x | 0.977x |
| gc/i31-roundtrip | 0.989x | 0.991x | 1.099x | 1.019x |
| gc/struct-access | 0.906x | 0.932x | 1.111x | 0.989x |
| gc/struct-allocation | 1.094x | 0.987x | 1.028x | 1.002x |
| gc/struct-indirect-access | 0.983x | 0.994x | 1.101x | 1.013x |
| memory/narrow-unaligned | 0.940x | 0.946x | 1.065x | 1.003x |
| memory/dependent-address | 0.978x | 0.986x | 1.091x | 1.010x |
| memory/float64-unaligned | 0.993x | 1.008x | 1.077x | 0.998x |
| runtime/simd-memory-full | 1.001x | 1.001x | 1.039x | 1.012x |
| runtime/simd-memory-widen | 0.997x | 0.999x | 1.263x | 1.001x |
| runtime/simd-memory-lanes | 1.001x | 0.999x | 1.010x | 1.014x |
| threads/private-memory | 1.026x | 1.109x | 1.039x | 1.025x |
| threads/shared-memory | 1.047x | 0.991x | 1.021x | 1.011x |
| threads/atomic-rmw | 1.015x | 1.005x | 1.058x | 1.049x |
| threads/notify-empty | 0.992x | 1.073x | 1.058x | 1.021x |
| threads/wait-mismatch | 1.003x | 1.032x | 1.058x | 1.021x |
| threads/wait-timeout | 1.058x | 0.896x | 1.051x | 1.033x |

The Go interpreter has adverse rows, including SIMD widening (1.263x base),
struct access (1.111x) and integer/memory loops (1.093x/1.089x). Its engine
sources did not change: the exact build source-hash comparison finds changes
only under codegen. The [exact-binary static audit](data/go-runtime-static-2026-11-05.json) checks
266 engine/VM/wasm functions. All instruction shapes match after documented
relocation normalization; eight LEAQ sites per binary are independently decoded
and verified against `runtime.zerobase`, explaining the seven initial text
differences. Of those functions, 249 change their entry offset modulo 64,
including all 35 execute functions. This supports investigating layout, but
does not prove it causes the slowdown or establish host-only noise.
Keep these observations open; no favorable timing rerun replaces them.

## Current competitor and host-compiler gaps

The table retains every case, both owned host compilers, both reference
interpreters and all compiled references. Unsupported GC references remain
explicit; no omitted feature is counted as a successful zero-cost result.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 0.998× | 1.772× | 5.291× | 2.735× | 1.641× | 3.852× | 2.512× | unavailable | inconclusive |
| core/integer-loop | 0.998× | 0.760× | 1.160× | 1.121× | 1.116× | 0.702× | 0.883× | 1.162× | resolved, advisory |
| core/function-calls | 0.996× | 0.827× | 0.580× | 0.764× | 0.690× | 0.753× | 1.129× | 0.935× | resolved, advisory |
| core/memory-load-store | 0.994× | 0.774× | 0.660× | 1.214× | 1.205× | 0.786× | 0.959× | 1.496× | resolved, advisory |
| applications/sha256-clang | 1.004× | 0.862× | 0.820× | 1.554× | 1.513× | 0.508× | 0.805× | 1.866× | inconclusive |
| applications/sha256-rust | 0.991× | 0.840× | 0.809× | 1.607× | 1.512× | 0.492× | 0.749× | 1.900× | resolved, advisory |
| gc/i31-roundtrip | 1.068× | 0.827× | unsupported | 1.008× | unsupported | unsupported | 0.405× | unavailable | inconclusive |
| gc/struct-access | 0.854× | 0.772× | unsupported | 1.777× | unsupported | unsupported | 0.371× | unavailable | inconclusive |
| gc/struct-allocation | 1.150× | 1.347× | unsupported | 0.303× | unsupported | unsupported | 1.407× | unavailable | inconclusive |
| gc/struct-indirect-access | 1.010× | 0.810× | unsupported | 2.834× | unsupported | unsupported | 0.735× | unavailable | inconclusive |
| memory/narrow-unaligned | 1.005× | 0.930× | 0.842× | 1.304× | 1.304× | 0.608× | 0.813× | 2.063× | resolved, advisory |
| memory/dependent-address | 0.990× | 0.836× | 1.014× | 1.354× | 1.353× | 0.453× | 0.364× | 1.580× | inconclusive |
| memory/float64-unaligned | 1.017× | 0.890× | 0.711× | 1.181× | 1.176× | 0.556× | 0.416× | 1.198× | resolved, advisory |
| runtime/simd-memory-full | 0.995× | 0.898× | 0.891× | 1.105× | 1.066× | 0.454× | 0.933× | 0.924× | resolved, advisory |
| runtime/simd-memory-widen | 1.003× | 0.790× | 0.762× | 1.718× | 1.732× | 0.355× | 0.710× | 0.443× | inconclusive |
| runtime/simd-memory-lanes | 0.994× | 1.116× | 0.953× | 1.147× | 1.158× | 0.433× | 1.227× | 1.338× | resolved, advisory |
| threads/private-memory | 1.244× | 1.057× | 1.239× | 3.711× | 3.127× | 0.840× | 0.628× | unavailable | inconclusive |
| threads/shared-memory | 1.188× | 0.648× | 1.273× | 3.714× | 2.928× | 0.865× | 0.637× | unavailable | inconclusive |
| threads/atomic-rmw | 1.142× | 0.673× | 1.451× | 1.432× | 1.291× | 0.850× | 0.620× | unavailable | inconclusive |
| threads/notify-empty | 1.104× | 0.860× | 0.079× | 0.338× | 0.252× | 0.784× | 0.509× | unavailable | inconclusive |
| threads/wait-mismatch | 1.081× | 0.659× | 0.058× | 0.249× | 0.605× | 0.696× | 0.508× | unavailable | inconclusive |
| threads/wait-timeout | 0.860× | 0.662× | 0.006× | 0.051× | 0.026× | 0.105× | 0.001× | unavailable | inconclusive |

Native source-call work beats the listed compiled references here. Remaining
deficits include tiny API entry, integer and scalar memory, SHA, GC access,
shared memory and SIMD widening. GC allocation is faster than Wasmtime in this
event but LLGo remains 1.150x Go natively and 1.347x Go interpreted. Private
and shared Threads native LLGo/Go ratios are 1.244x/1.188x; every Threads
control is inconclusive. These rows alone do not establish a new compiler bug.
LLGo #2983 stays with its existing contributor.

## Preparation and first execution

Each interval below is a separate median, in microseconds, for the Clang SHA
guest. First-call time includes guest work. Raw records also retain file load,
lookup, cleanup, readiness, all warmups/samples and clock controls. No timer or
startup phase is subtracted, and sums of medians do not reconstruct a total.

| Runtime | Engine setup (us) | Compile/validate (us) | Instantiate (us) | First call (us) |
| --- | ---: | ---: | ---: | ---: |
| wasmvm / Go native | 16.31 | 4290.30 | 89.38 | 2716.53 |
| wasmvm / LLGo native | 10.65 | 7508.84 | 38.93 | 2656.93 |
| wazero compiler | 148.15 | 796.14 | 58.47 | 3451.15 |
| Wasmtime Cranelift | 188.11 | 2849.65 | 82.86 | 1961.05 |
| Wasmer Cranelift | 183.63 | 2513.19 | 95.74 | 2063.60 |

LLGo compile/validate remains about 1.75x Go for this guest. The separate Core
memory case observes Go compile/validate changing from 504.71us to 685.30us
and LLGo from 1194.98us to 1231.56us; these preparation regressions remain
recorded rather than being canceled by warmed-execution improvements.
Actual x86 Go/LLGo native-code identity is not inferred from host timing.

Pinned references remain wazero 1.12.0, Wasmtime 48.0.1, Wasmer 7.3.0 and
WasmEdge 0.17.1. Sources, fixture identities, compiler flags and sampling are
fixed inside this event. Results from different hosts/source events are never
pooled or multiplied. General value allocation/splitting/rematerialization,
broader application/WASI-world qualification, final PR integration and controlled
competitor/native acceptance remain open.

The [evidence index](data/loop-scratch-registers-hosted-2026-11-05.json)
retains all phases, correctness records, source/toolchain identities and raw
artifact hashes. Local archive: `artifacts/loop-scratch-registers-20261105/final/hosted-evidence.tar.gz`,
2093033 bytes, SHA-256 `d3df0f7aaf94e0a9544d58b5ccfd2cc78b0af28e1a97f12991e4cf41a171baea`.
