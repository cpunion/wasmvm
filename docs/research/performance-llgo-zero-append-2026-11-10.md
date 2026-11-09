# LLGo fresh zero-slice append: 2026-11-10

[Machine record and exact reproduction](data/performance-llgo-zero-append-2026-11-10.json). Reported upstream as [LLGo #3040](https://github.com/xgo-dev/llgo/issues/3040).

The pinned LLGo release allocates both the source and result of
`append([]byte(nil), make([]byte, n)...)`. The source is fresh zero data and
need not survive as a separate allocation. Go 1.27.3 uses this expression in
[io.ReadAll](https://github.com/golang/go/blob/go1.27.3/src/io/io.go#L709), for
intermediate chunks and its final buffer. This supplies a concrete compiler-side
allocation gap alongside the separately identified P3 handoff costs.

M4 Max/macOS arm64, Go 1.27.3, LLVM 22.1.8, LLGo
`3d92a58e09bd8971899c6eaf63437fbecd18d432` with `-O3 -lto=full`.
The returning functions are noinline, the size is an argument, fresh slices
escape to a global, and 4096 iterations check length and boundary zero bytes.
`runtime.MemStats.TotalAlloc` gives bytes/operation:

| Size | Go make | Go append(make) | LLGo make | LLGo append(make) |
| --- | ---: | ---: | ---: | ---: |
| 4096 | 4097.3 | 4096.0 | 4112 | 8224 |
| 65536 | 65537.7 | 65536.3 | 65552 | 131104 |

Go and LLGo counters have different scopes. The decisive within-LLGo ratio is
exactly 2x at both sizes. Generated code independently shows a GC_malloc call in
appended, followed by SliceAppend, which allocates again and copies via memcpy.
The direct make allocates once. Go appended uses growslice and clears the new
result without allocating a source slice. This is byte accounting and code
inspection, not a Mallocs count, controlled latency result or complete
attribution of an end-to-end runtime gap. Diagnostic timings are retained in
the record without being used as acceptance evidence.

A fold must retain evaluation/panic behavior and the nil result of an empty
append. Latest upstream main was not rebuilt for this pinned-release report.
The two already assigned issues #3034/#3035 are left to upstream; no compiler
source or shared module cache was patched. Once upstream integrates a fix,
verify it in a separate compiler-upgrade event with both supported targets.
