# Public-call allocation diagnosis: 2026-11-05

**The public call path still allocates substantially more under pinned LLGo.**
These are allocation-byte diagnostics, not throughput or CPU-time attribution.
No runtime source, compiler pin, API contract or acceptance gate changes.

The VM probe uses engine `59a76f933901e56266e618bff0a502364eca0350`, built
from documentation head `e064497f06d6d72000446382ee9dbebbdbe52c8c`. Go 1.27.3
and LLGo `3571c2523e010e6842c4dd8a5213554fdc72d2a3` (`-O3 -lto=full`)
compile the same source on macOS arm64 with LLVM 22.1.8. Both report a 64-byte
`vm.Extern` and 32-byte `vm.Value`.

Each configuration runs five fresh processes with `GOMAXPROCS=1`, 1,000 warmup
calls, an explicit GC and 100,000 checked operations. The probe reads
`MemStats.TotalAlloc` around the operation loop, before cleanup. Guest bytes are
the existing `call-control.wasm`, SHA-256
`bcfd11fc8e3f474ca10508f3cb5afa7245ca2e782b2bf07bdd00686b93513ecc`;
every guest result is checked against `0x12345678`.

Median requested bytes per operation, rounded to the nearest byte:

| Path | Go interpreter | Go native | LLGo interpreter | LLGo native |
| --- | ---: | ---: | ---: | ---: |
| Complete `Extern.Call` | 296 | 296 | 1296 | 1296 |
| `Caller.Call` within one `WithCaller` session | 64 | 64 | 448 | 448 |
| Empty `WithCaller` operation | 288 | 288 | 1296 | 1296 |

The separate standard-library controls request 96 B/operation under Go and
368 B under LLGo for `WithCancel`/`Err`/cancel. Forcing `Done()` before cancel
changes these to 208 B and 640 B. These controls have different contracts from
VM operations; subtracting their costs would not isolate VM overhead. Session
setup and its callback closure are counted once; complete calls admit and finish
an operation each time. TotalAlloc is neither retained memory nor an allocation
event count, and LLGo reports approximate byte totals.

A separate standalone experiment models Extern's exact field layout and compares
a large value receiver, a small handle receiver, pointer arguments, a large
return and an empty control. All five processes for every path/compiler report
zero additional TotalAlloc bytes over 100,000 checked calls. This rejects the
simple claim that passing this 64-byte aggregate inherently allocates; it does
not establish the CPU cost of actual VM argument copies.

Exact linked LLGo assembly retains direct allocator sites in `Caller.Call`,
operation setup/cleanup, argument/result conversion and standard-library
context methods. `Caller.Call` has two direct sites around defer setup, consistent
with the already reported [LLGo #2932](https://github.com/xgo-dev/llgo/issues/2932).
Static sites include error/branch-dependent paths and allocating callees; site
counts do not count executed allocations or assign CPU shares. No duplicate
compiler issue is filed. [LLGo #2983](https://github.com/xgo-dev/llgo/issues/2983)
remains with its contributor; no SIMD-intrinsic changes are made here.

The [evidence record](data/public-allocation-2026-11-05.json) retains all 130
process results, both probe sources, build commands/hashes, fixture/source
identity, the unrounded summaries and static allocation-site context. Raw local
logs and assembly remain under `artifacts/public-allocation-20261105` and
`artifacts/extern-abi-allocation-20261105`.

This narrows the remaining investigation to call admission, standard-library
context/defer lowering and typed argument/result ownership. Preserve cancellation,
panic cleanup, escaping result ownership and expired capability behavior in any
optimization. The independent [routing comparison](performance-direct-runtime-routing-2026-11-05.md)
and controlled competitor/native acceptance remain open; no speedup or milestone
completion follows from these byte measurements.
