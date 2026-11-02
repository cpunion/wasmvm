# LLGo nonescaping value-parameter storage: 2026-11-03

[LLGo 2974](https://github.com/xgo-dev/llgo/issues/2974) reports a separately reproduced allocation
in an address-taken 112-byte input parameter. A defined noinline helper reads and
updates the borrowed object without retaining the pointer. The direct value-only
control performs the same arithmetic. Every loop checks an independent checksum.

Apple M4 Max/macOS arm64, Go 1.27.3, pinned LLGo `8e45ed5`, `-O3 -lto=full`,
GOMAXPROCS=1. Five shuffled fresh processes/compiler, seven warm loops/case,
100000 calls/loop. Medians of process medians:

| Case | Go reported bytes/call | LLGo reported bytes/call | Go ns/call | LLGo ns/call |
| --- | ---: | ---: | ---: | ---: |
| Direct value | 0 | 0 | 9.244 | 6.095 |
| Borrow address of input | 0 | 128 | 9.357 | 53.534 |

TotalAlloc is approximate; Mallocs is not an allocation-event counter. The final
LLGo borrowed body unconditionally calls AllocU(112) before main.modify; the
reported 128 bytes includes collector rounding/overhead. Its value-only body
has no allocator call. The helper's final body only checks nil, reads/writes the
object and returns. Go's final paths have no newobject/mallocgc call. Neither
reflection nor Cgo is needed in this standalone reproduction.

Related #2959/merged #2963 concerns named-result storage; this is the input
parameter. Broader escape-analysis proposals #2244/#2173 remain separate.
Current upstream was not rebuilt or timed. Native Context.Run has an analogous
112-byte state parameter whose address is borrowed by the common adapter, but
this program alone does not assign the complete VM's extra allocation or CPU
time to this one shape. A local in-place state API must preserve value-copy,
lease/close, snapshot, fuel and reentry contracts before being adopted.

[Complete source, every sample, compiler/binary hashes, both final assemblies and submitted issue](evidence/performance-llgo-value-parameter-2026-11-03.json.gz)
