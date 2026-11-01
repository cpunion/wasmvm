# Direct native bulk memory: 2026-11-02

Direct lowering removes native runtime-helper transitions for private
single-memory `memory.fill` and `memory.copy`. On the sustained bulk workload,
LLGo improves from **1,185.175 to 142.595 ns/source iteration**; Wasmtime measures
146.940 and the separate native C/libc reference 146.190. These workstation
observations approach native C throughput for this case. They do not establish
controlled superiority, cross-platform performance or whole-portfolio acceptance.

## Method and results

Base `88bb02ce3de712dff1c686c906864b365a8e3990` is the validated head of merged
PR #90. Measured head `8d25aed1192674fa72fbdbf8b1e1ebfe4cf6c847` contains the
implementation. The subsequent baseline merge and documentation do not alter
engine sources. Host and pins match the preceding bulk comparison: Apple M4 Max,
macOS arm64, Go 1.27.3, LLGo `8e45ed5` with `-O3 -lto=full`, Wasmtime 48.0.1,
wazero 1.12.0, Wasmer 7.3.0 and WasmEdge 0.17.1. Live SDK library and executable
hashes were checked against the original inventory.

The unchanged bulk-overlap binary fills 8 KiB, performs an overlapping 8-KiB
copy and updates a checksum per source iteration. Ordinary measurements retain
its existing 1,000 iterations and the unaffected private-memory control. A
separately declared sustained experiment uses **100,000 iterations**, the same
binary and the existing independent checksum oracle. Every pair has three
interleaved fresh process blocks, two warmups and seven warm calls per block.
All calls validate output; alternate-seed/count checks remain recorded.

Sustained warm time, **ns/source iteration; lower is better**:

| Runtime | Base | Head |
| --- | ---: | ---: |
| wasmvm Go interpreter | 318.897 | 315.776 |
| wasmvm Go native | 465.402 | 142.943 |
| wasmvm LLGo interpreter | 538.335 | 521.005 |
| wasmvm LLGo native | 1,185.175 | 142.595 |

| Reference | Warm ns/source iteration |
| --- | ---: |
| Native C/libc loop | 146.190 |
| Wasmtime Cranelift | 146.940 |
| Wasmer Cranelift | 147.780 |
| wazero compiler | 159.156 |
| WasmEdge interpreter | 218.310 |
| wazero interpreter | 248.806 |

The C reference executes the same fill/copy/checksum loop with libc calls,
compiled using the recorded Clang version, `-O3` and
`-fno-builtin-memset -fno-builtin-memmove`. Those flags retain the full memory
work. It omits guest range and cancellation checks, and is classified separately
from Wasm runtimes. Its calls share the sustained interleaving barriers; first
call and warm timers are separate. The
[source](../../benchmarks/native/c/bulk_baseline.c) includes independent-result
checking on every call. Reproduce the ordinary C probe with:

```sh
clang -O3 -fno-builtin-memset -fno-builtin-memmove \
  benchmarks/native/c/bulk_baseline.c -o /tmp/wasmvm-bulk-native
/tmp/wasmvm-bulk-native 305419896 1000 305674007
```

The ordinary LLGo native workload improves from 1,271.041 to 158.042 ns/iteration
versus Wasmtime's 165.000. Short-call/unchanged interpreter observations fluctuate;
they are not attributed to the native optimization. Startup, load, preparation,
instantiation, lookup, first call, warm calls and cleanup remain individually
timed, with no subtraction or synthetic empty preparation phases for native C.
All samples and both earlier loop variants are in the
[evidence](evidence/performance-native-bulk-2026-11-02.json). An attempted
128-byte unroll did not demonstrate improvement and was removed; the compact
64-byte loop remains. Shared-host timings do not constitute calibrated parity.

## Mechanism, size and semantics

Both emitters validate complete destination/source ranges, including u64
overflow, before any write. arm64 emits paired vector loads/stores with bounded
16-byte and byte tails. amd64 emits `REP STOSB`/`MOVSB`, restoring the context,
frame and clear direction flag on every normal or canceled exit. Overlapping
copies select a safe direction. Only scratch registers change; numeric, floating
and all 128 vector bits remain live across the operation.

The existing cancellation word is checked before every at-most-4,096-byte
chunk. A canceled operation returns through the regular spill/root protocol at
the original instruction PC. Cancellation is terminal and may leave partial
progress, as with existing runtime chunks. Shared memories, multiple memories,
memory initialization and precise fuel entries retain their checked runtime
helpers. Both i32 and i64 address/count forms use their validated widths.

The unchanged 29-instruction/8-slot fixture trades larger generated code for
removing its two bulk helper exits:

| Ordinary entry | Base bytes | Head bytes | Bulk helper exits |
| --- | ---: | ---: | ---: |
| amd64 | 1,837 | 2,272 | 2 → 0 |
| arm64 | 1,824 | 2,512 | 2 → 0 |

The 24%/38% increase is explicit checked loops and direction/poll exits;
precise entry sizes remain 5,014/3,728 bytes respectively. Code limits continue
to bound emission, including this additional code. This local tradeoff does not
accept whole-profile code size or peak-memory budgets.

Go/LLGo tests cover complete byte arrays, all overlap directions, zero lengths,
tails and chunk boundaries, i32/i64 ranges, u64 overflow, rejected writes,
preserved scalar/FP/vector values, helper exclusions and cancellation exits.
Affected native execution and full Core 2/3, Threads and Custom Page Sizes
inventories pass locally under both compilers. Race, strict cgo-pointer,
no-cgo and vet checks pass; local codegen statement coverage is 96.2%.
Final target CI is required before merging. Interpreter, shared-memory,
allocation and host/component performance gaps remain open.
