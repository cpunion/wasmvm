# Independent LLGo startup investigation: 2026-11-03

The Go/LLGo adapter readiness gap reproduces in an independent program that
only imports `os` and writes `ready\n`. The investigation is reported as
[LLGo 2973](https://github.com/xgo-dev/llgo/issues/2973). The root cause remains open; this is a
startup performance proposal with a minimal reproduction, not a claim that a
particular runtime mechanism is defective. Current upstream has not been rebuilt.

Apple M4 Max/macOS arm64; Go 1.27.3, pinned LLGo `8e45ed5`, `-O3 -lto=full`.
All builds precede timing. Five shuffled serial fresh-process blocks retain
launch-to-ready and complete process wall time independently. No timer/control
subtraction or pooling with the production VM measurements is performed.

| Common source | Go readiness median | LLGo readiness median |
| --- | ---: | ---: |
| os-only write | 3.508 ms | 34.112 ms |
| encoding/json ready encoder | 3.420 ms | 67.427 ms |
| os write plus blank VM import | 3.585 ms | 54.773 ms |

A separate five-block diagnostic compares the unchanged os-only binaries with
ordinary C `write` and C that calls GC initialization/thread registration,
`GC_malloc` and references libffi before writing. It links the same installed
BDWGC/libffi/libc++/resolv libraries. Ready medians: Go 3.568 ms, LLGo 34.052 ms,
C 2.376 ms, C plus GC/FFI 4.247 ms. This control is not a full Go runtime and
does not isolate every part of the remaining cost. No VM compilation,
instantiation or guest execution occurs in the os-only reproduction.

All three LLGo variants link the same library paths. Requested dynamic-loader
statistics returned no diagnostics; that negative result is retained. Package
initialization, runtime metadata and loader/runtime work need profiling before
causal attribution. Large shared-host differences justify a focused investigation,
while the warmed VM throughput deficits remain separate.

[Minimal sources, commands, hashes, linked libraries, every sample and submitted issue](evidence/performance-llgo-startup-2026-11-03.json.gz)
