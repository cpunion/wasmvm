# Interpreter memory result allocation: 2026-11-02

Memory load results now stay in values across scalar and SIMD lane operations.
This removes an unconditional LLGo host allocation without changing Wasm bits,
range checks, traps, fuel, cancellation, ownership or native instruction lowering.
Full competitive performance acceptance remains open.

## Sources and method

Measured base `b59618631ea6d8e12cf23522aaf122b8f52263b2` is the measured core
of PR #94. Measured head `b8adaa6598177a13aca6e35efccee63e30f9c883` changes one
memory primitive. Later documentation or independently recorded WASI fixes do
not supply implementation files to either archived build. Base binaries are
reused only after checking their exact revision, compiler identity and hashes.

Go 1.27.3 and pinned LLGo `8e45ed5` use the same release profile as their bases
(`-O3 -lto=full` for LLGo). The host is Apple M4 Max/macOS arm64. References
remain Wasmtime 48.0.1 Cranelift, Wasmer 7.3.0 Cranelift, WasmEdge 0.17.1
interpreter and wazero 1.12.0 in both modes. Live SDK and binary hashes match
the retained pins. These are descriptive workstation measurements.

The twelve unchanged workloads include scalar accesses, narrow/unaligned and
dependent addresses, SIMD full/widen/lane accesses, private/shared backing and
unchanged integer/call/atomic controls. Each uses its declared count, three fresh
serially interleaved process blocks, two warmups and seven warm calls. Every
result is checked, including alternate seed/count controls against independent
oracles. Startup, preparation, instantiation, first call, warm calls and cleanup
remain separate; no setup or clock control is subtracted.

| ns/source iteration | Go base | Head | LLGo base | Head | WasmEdge | wazero interpreter |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| threads/private-memory | 67.825 | 67.888 | 90.392 | 59.404 | 39.600 | 61.413 |
| threads/shared-memory | 103.479 | 101.875 | 99.604 | 93.383 | 39.500 | 63.888 |
| threads/atomic-rmw | 77.704 | 76.546 | 100.554 | 99.425 | 32.600 | 47.758 |
| core/integer-loop | 75.045 | 75.525 | 72.010 | 72.908 | 48.880 | 82.185 |
| core/function-calls | 96.286 | 97.963 | 130.094 | 131.681 | 62.273 | 83.361 |
| core/memory-load-store | 149.502 | 150.787 | 147.410 | 135.150 | 92.913 | 148.037 |
| memory/narrow-unaligned | 183.580 | 184.121 | 189.230 | 164.815 | 112.952 | 182.551 |
| memory/dependent-address | 85.246 | 85.946 | 95.440 | 83.723 | 53.398 | 94.704 |
| memory/float64-unaligned | 140.978 | 139.799 | 139.791 | 126.791 | 84.136 | 138.306 |
| runtime/simd-memory-full | 145.617 | 147.233 | 158.871 | 134.887 | 78.200 | 149.367 |
| runtime/simd-memory-widen | 137.988 | 138.650 | 163.033 | 129.300 | 62.900 | 144.017 |
| runtime/simd-memory-lanes | 176.688 | 180.871 | 197.963 | 157.483 | 91.600 | 186.800 |

LLGo private reads/writes improve from 90.392 to 59.404 ns/iteration, and
SIMD widening from 163.033 to 129.300. Both retain deficits against WasmEdge.
Go's scalar path remains approximately unchanged; its positive control
variations remain visible. The native code generator is unchanged.

Short native controls include an apparent Go shared-memory regression of
2.400→2.750 ns and SIMD widening of 2.567→2.875. A separately declared pass
uses the unchanged binaries for all nine native/reference configurations on
shared memory, SIMD full-memory and SIMD widening at 1,000,000 iterations,
recomputing the independent oracle. It retains three interleaved fresh blocks,
two warmups and seven calls, with all phases separate. These longer controls
resolve the setup/timer sensitivity of the short samples; neither campaign
establishes controlled parity. Full observations remain in the
[evidence](evidence/performance-interpreter-memory-2026-11-02.json).

| Sustained native, ns/iteration | Go base | Head | LLGo base | Head | Wasmtime | Wasmer | wazero compiler |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |

| threads/shared-memory | 0.863 | 0.843 | 0.832 | 0.841 | 0.992 | 0.979 | 1.138 |
| runtime/simd-memory-full | 3.247 | 3.258 | 3.062 | 3.238 | 3.358 | 3.227 | 3.508 |
| runtime/simd-memory-widen | 1.056 | 1.051 | 1.029 | 1.016 | 1.018 | 1.014 | 1.111 |

## Cause and implementation

A three-second macOS `sample` capture of the base private-memory loop places
75 of 108 samples under `Memory.load` inside `runtime.AllocU`; 282 other guest
samples remain in the dispatcher itself. This is a sampled attribution, not a
precise percentage of elapsed time or an allocation counter. Profiling runs
are separate diagnostics and are excluded from the timing table.

Base LLGo assembly allocates 16 bytes at `Memory.load` entry, before selecting
scalar or SIMD behavior. The result's address is passed to a lane update helper
in a SIMD branch; that conservative host escape affects scalar loads too. The
helper now accepts and returns the two scalar words by value. Fixed width tables
are read-only metadata, and scalar loads no longer reserve the vector helper
frame. LLGo scalar-load frame storage falls from 496 to 128 bytes. Its lane mask,
packed bits and preserved lanes stay identical. Final assembly has no result
allocation at this entry. This does not claim that all interpreter/shared-memory
operations are host-allocation-free. The allocation-only intermediate
`4c16294d4ed3405bdd8d8c650eb28f9e4fb19e8a` and its complete twelve-case
comparison remain separate evidence; scalar-frame outlining does not establish
an additional resolved throughput gain. The minimal compiler reproduction is
tracked in [LLGo #2959](https://github.com/xgo-dev/llgo/issues/2959).

## Validation and remaining work

Affected memory, execution and public VM packages pass with Go and LLGo release,
including complete Core 2.0/3.0 and pinned proposal inventories. Existing tests
cover bounds, signed/packed bits, unaligned vectors, lane preservation, forced
collection and backing/lifetime interactions. The initial local attempt lacked
the pinned corpus link; that input failure is retained and was resolved before
semantic validation. Final PR CI must still cover all 31 Go/LLGo packages on
Linux amd64/macOS arm64 and all 30 coverage gates.

Dispatcher, function-call, shared-adapter, general SIMD, host/component and
controlled portfolio performance gaps remain open. This targeted allocation
removal does not close any overall performance milestone.
