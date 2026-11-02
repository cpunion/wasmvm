# LLGo defer setup before an early return: 2026-11-02

The independent reproduction confirms a remaining LLGo allocation problem and
adds a cold-path case to existing [issue #2932](https://github.com/xgo-dev/llgo/issues/2932#issuecomment-6579659757). No duplicate issue is
opened. Unlike the original issue's older/default-profile result, this probe
uses Go 1.27.3 and pinned LLGo `8e45ed5` with `-O3 -lto=full` on macOS arm64 /
Apple M4 Max. Current upstream main has not been rebuilt and the pin is unchanged.

The same standalone Go source calls four noinline functions. Each directly or
deferredly adds two to a borrowed uint64. The skipped case returns before its
defer; the capture case can modify a captured local on other inputs. All five
fresh processes per compiler run 100,000 calls per case after a warmup, with
`GOMAXPROCS=1`, balanced randomized compiler order and checksum 200,000.

| Case | Go bytes/call | LLGo bytes/call, median (range) | Go ns/call | LLGo ns/call |
| --- | ---: | ---: | ---: | ---: |
| Direct | 0 | 0 | 1.304 | 1.525 |
| Static defer | 0 | 111.996 (111.978–112.029) | 2.339 | 46.556 |
| Mutable capture | 0 | 175.985 (175.948–176.019) | 2.493 | 78.902 |
| Return before defer | 0 | 64.001 (63.970–64.006) | 1.358 | 31.178 |

Byte totals use approximate `runtime.MemStats.TotalAlloc`; LLGo's Mallocs field
is not an allocation-event oracle. The final-binary disassembly independently
shows `AllocU(48)` before the skipped function tests the condition. The early
return never registers or invokes its deferred call. The direct control has no
allocation call. The reported size includes collector rounding/overhead; it is
not a claim that the generated request is 64 bytes. The ordinary static defer
requests 48- and 32-byte records in this reproduction. Shared-host nanoseconds
are descriptive diagnostics, not controlled throughput acceptance.

This exact conditional setup pattern is present in wasmvm's
`FinishNativeAllocations`, including its no-pending-allocation return. Other
AllocU sites and slice construction calls are retained in its assembly; a
NewSlice2 call alone is not evidence of allocation. We do not attribute the full
VM GC slowdown to this one function or propose dropping cleanup/unwind behavior.
Any local rewrite must preserve all early errors, reservation rollback, panic
unwinding and guest root/handle lifetime contracts.

The requested compiler improvement is to sink setup past proved early returns
and eventually keep non-loop records on the stack or open-code the normal path,
while preserving receiver/argument capture, LIFO, panic/recover and unwinding.
The exact source, build commands, source/binary/toolchain hashes, five-process
results, final assembly and posted issue response are retained below.

[Compressed independent reproduction and upstream response](evidence/performance-llgo-defer-2026-11-02.json.gz)
