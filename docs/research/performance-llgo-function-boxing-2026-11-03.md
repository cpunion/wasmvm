# LLGo nonescaping function boxing: 2026-11-03

Reported as [LLGo #2975](https://github.com/xgo-dev/llgo/issues/2975), a concrete case for the existing escape-analysis
proposal #2244. Go 1.27.3 and pinned LLGo `8e45ed5` / -O3 -lto=full, macOS
arm64/M4 Max. No Wasm, Cgo or benchmark timing dependency. Current upstream
main was not rebuilt; the report does not assert its measured status.

The standalone program constructs its closure environment/counter/stack before
measurement. Each of five fresh processes/compiler warms each case 64 times,
collects, invokes it 100000 times and verifies checksum 900000. No durations are
sampled. `TotalAlloc` is approximate collector accounting, not Mallocs events.

| Common-source path | Go TotalAlloc bytes | LLGo bytes, five processes |
| --- | ---: | --- |
| Function → local any → non-retaining extraction/call | 0 | 3195456, 3195456, 3196288, 3196672, 3195456 |
| Extract preboxed function | 0 | 0, 0, 0, 0, 0 |
| Extract preboxed record pointer | 0 | 0, 0, 0, 0, 0 |
| Box/extract record pointer | 0 | 0, 0, 0, 0, 0 |
| Direct function / direct record calls | 0 | 0 in every process |

Final assembly in `boxFunction` requests **AllocU(16)**, writes its function/
context pair and calls `assertFunction`. The extraction helper has no allocator
call. Pinned MakeInterface source and its exact blob/content hashes are retained.
There are no named result addresses. Preserve escaping boxes, closure roots and
typed-nil/interface semantics when improving this known-nonescaping case.

The initial four-case experiment implicitly converted its function argument to
any at the invocation site. Its first attribution to assertion was incorrect;
assembly identified conversion, and the corrected six-case source added explicit
preboxed controls before filing the issue. Both sources, binaries' identities,
all five-process records and final assembly remain. VM host callbacks were
already preboxed at registration, so this is not a claimed cause of their
complete warm boundary gap or a justification for the held binding record.

[Both reproductions, exact flags/binary hashes, controls, assembly and issue body](evidence/performance-llgo-function-boxing-2026-11-03.json.gz)
