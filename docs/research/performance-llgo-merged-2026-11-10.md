# Fully integrated LLGo proposal 3012: measured impact, 2026-11-10

[Complete machine record](data/performance-llgo-merged-2026-11-10.json).

## Result and production decision

The full integrated candidate reduces channel/map temporary allocation and improves observed concurrent P3 component throughput, while most compute-heavy core workloads change little. It also regresses the short-batch WasmGC allocation interpreter workload. **Keep the production compiler pin at 3d92a58** while that finding is investigated as [LLGo #3045](https://github.com/xgo-dev/llgo/issues/3045). Candidate Linux acceptance and controlled performance/release gates remain open. No runtime or compiler source is patched in this evaluation.

HTTP fixture, 64 KiB, four workers; all numbers below belong to this one same-source compiler event:

| Path | Old LLGo req/s | Candidate req/s | Old B/request | Candidate B/request |
| --- | ---: | ---: | ---: | ---: |
| P3 streaming / native | 2696 | 3920 | 846277 | 754768 |
| P3 buffered / native | 3226 | 5506 | 561018 | 474806 |
| Ordinary native-Go application control | 8854 | 7985 | 732584 | 732057 |

**Only 5/50 P3 A/A controls resolve, and no old/new pair resolves both.** These timing observations establish neither a stable speed win nor a controlled ranking. The buffered native paired warm-time ratio is 0.5878 [0.5752,0.6008]; the ordinary four-worker host control instead observes 1.0760 [1.0225,1.1276], also unresolved. All adverse and unchanged cases remain below.

Peak RSS rises 159.4 -> 175.8 MiB for LLGo buffered native HTTP and 160.0 -> 172.5 MiB for streaming, despite lower cumulative allocation. Observed between-batch thread counts remain 57 and 61 respectively. Same-event Go buffered native observes 12798 req/s, Wasmtime 48.0.1 observes 56515, and the ordinary Go application 42726. The remaining competitor and host-runtime gaps are substantial; ratios across these contracts do not partition their causes.

## What merged

[Proposal #3012](https://github.com/xgo-dev/llgo/issues/3012) reports all ten implementation sections merged, finalized by #3039 at `717fbe4429cbe50222eabbf0c26d04c691c439a6`. This evaluation uses that complete integration. The prior production pin `3d92a58e09bd8971899c6eaf63437fbecd18d432` already includes #2997, #2985 and #2989: their earlier gains are not counted again or multiplied into this event. The [earlier compiler baseline](performance-llgo-escape-upgrade-2026-11-09.md) remains a separate source/host event.

| New merge beyond the old pin | Area / verification scope |
| --- | --- |
| [#3007](https://github.com/xgo-dev/llgo/pull/3007) | Pointer-free typed objects/slices use non-scanning allocation; included in aggregate runtime results |
| [#3021](https://github.com/xgo-dev/llgo/pull/3021) | Bounded call-site effects, immutable captures and transient-copy proofs; included in aggregate results and exact upstream probes |
| [#3022](https://github.com/xgo-dev/llgo/pull/3022) | Typed stack-field roots on explicit-root targets; merged scope recorded, no separate J32/J64/W32 performance claim here |
| [#3023](https://github.com/xgo-dev/llgo/pull/3023) | Complete SCC facts and opt-in frontend persistence; separate warm-build experiment below |
| [#3024](https://github.com/xgo-dev/llgo/pull/3024) | HostLayout reflection alignment; included compatibility change, no isolated speed claim |
| [#3036](https://github.com/xgo-dev/llgo/pull/3036) | Strict-C11 generated declarations; unchanged reproduction fails to build on old compiler and builds/executes on candidate and Go |
| [#3037](https://github.com/xgo-dev/llgo/pull/3037) | Ordinary channel waiter reuse; independent blocking-exchange byte counters become zero |
| [#3039](https://github.com/xgo-dev/llgo/pull/3039) | Caller-owned map keys and map/string iterators; independent temporary-allocation counters become zero |

The full-version experiment cannot assign an application speed change to one PR. Completing the planned placement architecture preserves conservative lifetime/size/target fallbacks and does not promise zero allocation everywhere. In particular [#3040](https://github.com/xgo-dev/llgo/issues/3040) remains independently reproducible: append of a fresh zero slice still accounts for two buffers under both LLGo versions.

## Fixed protocol and qualification

All engine/harness inputs are held at wasmvm `2e5ee34d8070595e86cd7a7bec1ffc1da33a696e`. M4 Max/macOS 27.0.1; Go 1.27.3, LLVM 22.1.8 and identical LLGo `-O3 -lto=full`; no development tags or compiler patches. Each LLGo executable uses its matching clean runtime checkout. Compiler, program, source-archive, fixture and SDK identities are checked; seven actual core-loader and seven actual P3-loader probes are retained. The shared service driver matches the frozen independent reference byte for byte.

The original22 portfolio uses five fresh interleaved blocks with two warmups/seven samples, seed 2026111011: 1810 runtime processes, 55 native-C processes, 362 alternate-parameter checks and 12 expected unsupported GC probes. Go aliases share one physical binary; both old/new LLGo native and interpreter have same-binary A/A controls. The allocation workload uses process means including every pause, with median diagnostics kept in raw data. Other warm kernels use process medians. Timers, flags, fixtures and counts retain the established routing protocol.

P3 uses 20 randomized primary/A/A blocks, seed 2026111012: 2000 throughput processes, 50 separate resource processes, 100 qualification processes and 896000 fully checked warm responses. Both Rust plain/composed HTTP fixtures, old/new interpreter/native, streaming/buffered, one/four workers, 16 clients with four requests, 64 KiB and CRC32 2629362236. Go native and Wasmtime/ordinary host controls remain separately identified. Every byte/status/header/trailer and guest join is checked. Go interpreter is not resampled as a P3 control in this event.

All owned build/test/probe sessions exited before runtime sampling, and the actual pre-start compiler/test process table was clear. The workstation has no physical affinity/frequency/background isolation. No favorable repeat, source/profile/count sweep, cross-event pooling or phase subtraction. Pointwise paired-block log-ratio bootstrap uses 8000 draws; there is no family-wise or controlled release claim. Startup, load, engine/compile, instantiate, lookup, first/warm calls, complete API, request latency and joined cleanup remain separate in the machine record. Interleaved process wall also contains protocol parking; it is not a standalone CLI latency.

The candidate passes all 32 actual owned-package tests and both actual nested measurement modules on macOS, with both execution backends. These are candidate host results, not a new-compiler Linux acceptance claim. Original baseline final-head semantic evidence remains [PR #150](https://github.com/cpunion/wasmvm/pull/150). A first local broad discovery command accidentally scanned historical ignored artifacts and failed before tests; the explicit cmd/internal/pkg inventory then passes. An initial macOS loader probe through protected /usr/bin/env stripped DYLD logging; direct-binary probes pass. These diagnostic failures remain retained. Core raw powerPolicy retains a legacy Hosted-VM label; actual platform/CPU/classification/affinity are local Darwin/M4/none. The metadata erratum changes no timing or source.

## All original22 comparisons

Ratios are candidate/old warm time, smaller is faster. A/A means both relevant controls resolve. All44 rows remain in the machine record, including adverse observations.

| Workload | Native ratio [95%] | Native A/A | Interpreter ratio [95%] | Interpreter A/A |
| --- | --- | --- | --- | --- |
| core/call-control | 0.9157 [0.7203,1.1034] | no | 1.1555 [1.0323,1.3275] | no |
| core/integer-loop | 0.9982 [0.9904,1.0057] | yes | 1.0017 [0.9839,1.0153] | yes |
| core/function-calls | 1.0027 [1.0005,1.0049] | no | 1.0066 [0.9953,1.0180] | yes |
| core/memory-load-store | 1.0020 [0.9893,1.0134] | no | 1.0217 [1.0033,1.0379] | no |
| memory/narrow-unaligned | 0.9923 [0.9796,1.0042] | yes | 0.9984 [0.9839,1.0133] | no |
| memory/dependent-address | 0.9921 [0.9809,1.0029] | yes | 1.0131 [0.9936,1.0465] | yes |
| memory/float64-unaligned | 0.9981 [0.9846,1.0111] | yes | 1.0032 [0.9851,1.0217] | yes |
| runtime/simd-memory-full | 1.0062 [1.0006,1.0119] | yes | 1.0154 [1.0029,1.0281] | yes |
| runtime/simd-memory-widen | 1.0030 [0.9966,1.0097] | no | 0.9902 [0.9863,0.9940] | yes |
| runtime/simd-memory-lanes | 0.9997 [0.9851,1.0137] | yes | 0.9968 [0.9777,1.0163] | no |
| gc/i31-roundtrip | 0.9758 [0.9143,1.0377] | no | 0.9985 [0.9673,1.0371] | no |
| gc/struct-access | 0.9705 [0.9477,0.9901] | no | 1.0164 [0.9736,1.0588] | no |
| gc/struct-allocation | 1.1481 [1.0082,1.2639] | no | 1.0643 [1.0434,1.0856] | yes |
| gc/struct-indirect-access | 1.0142 [0.9805,1.0537] | no | 0.9858 [0.9616,1.0166] | no |
| threads/private-memory | 0.9848 [0.8921,1.0873] | no | 1.0023 [0.9796,1.0224] | yes |
| threads/shared-memory | 1.1128 [0.9836,1.2407] | no | 1.0034 [0.9894,1.0138] | no |
| threads/atomic-rmw | 1.0004 [0.9623,1.0405] | no | 0.9855 [0.9641,1.0090] | no |
| threads/notify-empty | 0.9537 [0.9105,0.9933] | no | 1.0034 [0.9906,1.0177] | no |
| threads/wait-mismatch | 1.0072 [0.9238,1.0980] | no | 1.0037 [0.9926,1.0149] | no |
| threads/wait-timeout | 1.0084 [0.9567,1.0712] | no | 0.9950 [0.9746,1.0154] | no |
| applications/sha256-clang | 0.9933 [0.9807,1.0068] | yes | 0.9945 [0.9909,0.9980] | yes |
| applications/sha256-rust | 0.9848 [0.9670,1.0055] | no | 1.0020 [0.9938,1.0087] | yes |

## GC adverse finding and independent sustained cohort

The original 10000-iteration struct-allocation interpreter observes 97.3655 -> 103.0667 ns/guest iteration, paired ratio 1.0643 [1.0434,1.0856]; both A/A intervals fit ±2.5%. This is the clearest calibrated adverse finding, reported as [#3045](https://github.com/xgo-dev/llgo/issues/3045), with exact reproduction, fixture hash and source/compiler pins. Native short-batch ratio is 1.1481 with unresolved controls.

A separate diagnostic uses the already-declared sustained-GC regime: 300000 allocation iterations, 1000000 for the other three cases, five blocks/seven samples, seed 2026111016. It completes 280 runtime processes, 56 parameter checks and 12 expected unsupported probes, reusing exactly the frozen binaries. It investigates scale and does not replace or pool with the original finding.

| Sustained workload | Native ratio [95%] | Native A/A | Interpreter ratio [95%] | Interpreter A/A |
| --- | --- | --- | --- | --- |
| gc/i31-roundtrip | 1.0020 [0.9938,1.0097] | yes | 0.9989 [0.9861,1.0118] | yes |
| gc/struct-access | 0.9982 [0.9889,1.0081] | no | 1.0094 [1.0015,1.0175] | yes |
| gc/struct-allocation | 0.9977 [0.9827,1.0128] | no | 1.0558 [1.0472,1.0660] | no |
| gc/struct-indirect-access | 1.0253 [0.9830,1.0708] | no | 1.0178 [0.9937,1.0465] | no |

The sustained allocation interpreter ratio is about 1.0558, with one control unresolved. Native sustained allocation is approximately flat. This supports further investigation of the specific interpreter path; no individual PR cause, universal allocation regression or physical-isolation claim is inferred.

## Every P3 old/new comparison

| Candidate profile | Old req/s | New req/s | Warm ratio [95%] | Both A/A | Old B/request | New B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| host/none/native-go/new/host/1 | 8263 | 8893 | 0.9540 [0.9184,0.9942] | no | 732582 | 732060 |
| host/none/native-go/new/host/4 | 8854 | 7985 | 1.0760 [1.0225,1.1276] | no | 732584 | 732057 |
| new/buffered/world-rust-composed-http/new/interpreter/1 | 2096 | 2256 | 0.9422 [0.9211,0.9668] | no | 582386 | 493709 |
| new/buffered/world-rust-composed-http/new/interpreter/4 | 3199 | 5035 | 0.6386 [0.6093,0.6639] | no | 582302 | 493729 |
| new/buffered/world-rust-composed-http/new/native/1 | 2595 | 2833 | 0.9241 [0.9070,0.9417] | no | 578132 | 489245 |
| new/buffered/world-rust-composed-http/new/native/4 | 3105 | 5182 | 0.6029 [0.5873,0.6178] | no | 577970 | 489226 |
| new/buffered/world-rust-http/new/interpreter/1 | 2394 | 2463 | 0.9671 [0.9571,0.9779] | no | 560690 | 474582 |
| new/buffered/world-rust-http/new/interpreter/4 | 3332 | 5233 | 0.6273 [0.6132,0.6415] | no | 560463 | 474584 |
| new/buffered/world-rust-http/new/native/1 | 2866 | 3102 | 0.9210 [0.8969,0.9466] | no | 561165 | 474829 |
| new/buffered/world-rust-http/new/native/4 | 3226 | 5506 | 0.5878 [0.5752,0.6008] | no | 561018 | 474806 |
| new/streaming/world-rust-composed-http/new/interpreter/1 | 1615 | 1681 | 0.9629 [0.9432,0.9845] | no | 867533 | 773632 |
| new/streaming/world-rust-composed-http/new/interpreter/4 | 2668 | 3563 | 0.7322 [0.6886,0.7729] | no | 867569 | 773729 |
| new/streaming/world-rust-composed-http/new/native/1 | 1838 | 1921 | 0.9708 [0.9196,1.0467] | no | 863228 | 769138 |
| new/streaming/world-rust-composed-http/new/native/4 | 2572 | 3700 | 0.7007 [0.6796,0.7266] | no | 863127 | 769184 |
| new/streaming/world-rust-http/new/interpreter/1 | 1729 | 1810 | 0.9328 [0.9008,0.9563] | no | 845779 | 754512 |
| new/streaming/world-rust-http/new/interpreter/4 | 2752 | 3711 | 0.7429 [0.7236,0.7647] | no | 845874 | 754550 |
| new/streaming/world-rust-http/new/native/1 | 1984 | 2110 | 0.9247 [0.8896,0.9519] | no | 846213 | 754748 |
| new/streaming/world-rust-http/new/native/4 | 2696 | 3920 | 0.6782 [0.6572,0.6977] | no | 846277 | 754768 |

## All P3 profiles and resources

| Profile | req/s | Complete API us | Runtime B/request | Between-batch threads | Process peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| go/buffered/world-rust-composed-http/go/native/1 | 3582 | 275.5 | 371428 | 7 | 144.0 | yes |
| go/buffered/world-rust-composed-http/go/native/4 | 11648 | 330.0 | 371431 | 7 | 163.8 | no |
| go/buffered/world-rust-http/go/native/1 | 4024 | 244.8 | 359824 | 7 | 111.1 | no |
| go/buffered/world-rust-http/go/native/4 | 12798 | 299.9 | 359824 | 7 | 127.1 | no |
| go/streaming/world-rust-composed-http/go/native/1 | 3252 | 304.9 | 451404 | 7 | 148.3 | no |
| go/streaming/world-rust-composed-http/go/native/4 | 10803 | 355.8 | 451376 | 7 | 154.5 | no |
| go/streaming/world-rust-http/go/native/1 | 3684 | 268.9 | 439792 | 8 | 111.2 | no |
| go/streaming/world-rust-http/go/native/4 | 11569 | 331.9 | 439775 | 7 | 125.5 | no |
| host/none/native-go/go/host/1 | 39284 | 21.5 | 277914 | 7 | 16.6 | no |
| host/none/native-go/go/host/4 | 42726 | 80.5 | 277918 | 6 | 15.9 | no |
| host/none/native-go/new/host/1 | 8893 | 106.9 | 732060 | 37 | 67.5 | no |
| host/none/native-go/new/host/4 | 7985 | 457.0 | 732057 | 40 | 68.3 | no |
| host/none/native-go/old/host/1 | 8263 | 115.2 | 732582 | 37 | 68.0 | no |
| host/none/native-go/old/host/4 | 8854 | 405.3 | 732584 | 40 | 68.4 | no |
| new/buffered/world-rust-composed-http/new/interpreter/1 | 2256 | 436.7 | 493709 | 42 | 160.8 | no |
| new/buffered/world-rust-composed-http/new/interpreter/4 | 5035 | 759.9 | 493729 | 57 | 161.4 | no |
| new/buffered/world-rust-composed-http/new/native/1 | 2833 | 346.7 | 489245 | 42 | 189.7 | no |
| new/buffered/world-rust-composed-http/new/native/4 | 5182 | 754.2 | 489226 | 57 | 205.2 | no |
| new/buffered/world-rust-http/new/interpreter/1 | 2463 | 394.5 | 474582 | 42 | 140.5 | yes |
| new/buffered/world-rust-http/new/interpreter/4 | 5233 | 745.1 | 474584 | 57 | 140.8 | no |
| new/buffered/world-rust-http/new/native/1 | 3102 | 316.0 | 474829 | 42 | 164.0 | no |
| new/buffered/world-rust-http/new/native/4 | 5506 | 709.5 | 474806 | 57 | 175.8 | no |
| new/streaming/world-rust-composed-http/new/interpreter/1 | 1681 | 586.5 | 773632 | 43 | 160.5 | yes |
| new/streaming/world-rust-composed-http/new/interpreter/4 | 3563 | 1090.6 | 773729 | 61 | 161.0 | no |
| new/streaming/world-rust-composed-http/new/native/1 | 1921 | 511.7 | 769138 | 43 | 192.2 | no |
| new/streaming/world-rust-composed-http/new/native/4 | 3700 | 1044.1 | 769184 | 61 | 216.0 | no |
| new/streaming/world-rust-http/new/interpreter/1 | 1810 | 545.4 | 754512 | 43 | 140.4 | yes |
| new/streaming/world-rust-http/new/interpreter/4 | 3711 | 1048.8 | 754550 | 61 | 140.8 | no |
| new/streaming/world-rust-http/new/native/1 | 2110 | 461.9 | 754748 | 43 | 167.0 | no |
| new/streaming/world-rust-http/new/native/4 | 3920 | 991.4 | 754768 | 61 | 172.5 | no |
| old/buffered/world-rust-composed-http/old/interpreter/1 | 2096 | 469.9 | 582386 | 42 | 150.0 | no |
| old/buffered/world-rust-composed-http/old/interpreter/4 | 3199 | 1227.3 | 582302 | 57 | 167.9 | no |
| old/buffered/world-rust-composed-http/old/native/1 | 2595 | 378.9 | 578132 | 42 | 213.9 | no |
| old/buffered/world-rust-composed-http/old/native/4 | 3105 | 1265.0 | 577970 | 57 | 194.4 | no |
| old/buffered/world-rust-http/old/interpreter/1 | 2394 | 411.3 | 560690 | 42 | 130.8 | no |
| old/buffered/world-rust-http/old/interpreter/4 | 3332 | 1182.3 | 560463 | 57 | 140.8 | no |
| old/buffered/world-rust-http/old/native/1 | 2866 | 342.6 | 561165 | 42 | 168.2 | no |
| old/buffered/world-rust-http/old/native/4 | 3226 | 1217.8 | 561018 | 57 | 159.4 | no |
| old/streaming/world-rust-composed-http/old/interpreter/1 | 1615 | 612.1 | 867533 | 43 | 163.6 | no |
| old/streaming/world-rust-composed-http/old/interpreter/4 | 2668 | 1463.1 | 867569 | 61 | 181.6 | no |
| old/streaming/world-rust-composed-http/old/native/1 | 1838 | 534.1 | 863228 | 43 | 196.8 | no |
| old/streaming/world-rust-composed-http/old/native/4 | 2572 | 1516.0 | 863127 | 61 | 209.5 | no |
| old/streaming/world-rust-http/old/interpreter/1 | 1729 | 569.7 | 845779 | 43 | 140.5 | no |
| old/streaming/world-rust-http/old/interpreter/4 | 2752 | 1424.9 | 845874 | 61 | 141.0 | no |
| old/streaming/world-rust-http/old/native/1 | 1984 | 495.4 | 846213 | 43 | 167.2 | yes |
| old/streaming/world-rust-http/old/native/4 | 2696 | 1442.0 | 846277 | 61 | 160.0 | no |
| reference/none/world-rust-composed-http/go/compiler/1 | 14634 | 65.8 | 32 | 7 | 32.7 | no |
| reference/none/world-rust-composed-http/go/compiler/4 | 49995 | 75.1 | 32 | 7 | 33.9 | no |
| reference/none/world-rust-http/go/compiler/1 | 16935 | 56.6 | 32 | 7 | 29.6 | no |
| reference/none/world-rust-http/go/compiler/4 | 56515 | 65.9 | 32 | 7 | 30.6 | no |

Process reductions, phase timings, request latency, resource snapshots and all40 owned/Wasmtime paired intervals are in the machine record. Go allocation counters exclude Rust/JIT memory. LLGo bytes are BDWGC accounting and not malloc counts; cross-compiler byte figures have different scopes. The ordinary host control has no Wasm isolation or component contract.

## Independent allocation and isolated-kernel probes

Five fixed fresh-process rounds produce 130 executions. Exact upstream coverage/borrow/boxing/defer/map-key sources come from the integrated snapshot; channel/zero-append use the unchanged earlier reproductions. A common standalone VM probe checks typed/scalar/session calls. All logical checksums match across compilers. Defer retains the upstream 500ms adaptive testing.Benchmark calibration; other embedded call counts and controls stay fixed. Kernel timings below are diagnostic, without A/A-controlled speed claims.

Each cell is accounted bytes/op; ns/op when available. Go bytes are managed allocation; LLGo bytes are collector accounting. Every raw counter observation and zero-work control remains in JSON.

| Probe / case | Go B / ns | Old LLGo B / ns | Candidate B / ns |
| --- | ---: | ---: | ---: |
| append-zero/append-make/4096 | 4096.03 / 317.58 | 8224.00 / 3185.83 | 8224.00 / 2821.03 |
| append-zero/append-make/65536 | 65536.44 / 2667.34 | 131104.00 / 45199.63 | 131104.00 / 51249.17 |
| append-zero/direct/4096 | 4096.03 / 326.22 | 4112.00 / 1554.77 | 4112.00 / 1409.31 |
| append-zero/direct/65536 | 65536.50 / 2637.64 | 65552.00 / 24135.06 | 65552.00 / 26591.08 |
| borrow/borrow | 0.00 / 9.38 | 0.00 / 9.13 | 0.00 / 9.07 |
| borrow/value | 0.00 / 9.26 | 0.00 / 6.10 | 0.00 / 6.08 |
| boxing/assert-preboxed-function | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| boxing/assert-preboxed-record | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| boxing/assert-record | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| boxing/box-function | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| boxing/direct-function | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| boxing/direct-record | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| channel/buffered-ping-pong | 0.00 / N/A | 416.11 / N/A | 0.00 / N/A |
| channel/same-goroutine-buffered | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| channel/unbuffered-ping-pong | 0.00 / N/A | 416.03 / N/A | 0.00 / N/A |
| coverage/bytes-to-string | 0.00 / 3.56 | 0.00 / 4.12 | 0.00 / 3.27 |
| coverage/dynamic-slice-large | 128.04 / 25.66 | 143.96 / 60.34 | 143.95 / 58.77 |
| coverage/dynamic-slice-small | 0.00 / 2.14 | 0.00 / 3.35 | 0.00 / 6.80 |
| coverage/fixed-capacity | 0.00 / 2.07 | 0.00 / 1.55 | 0.00 / 1.51 |
| coverage/fixed-slice | 0.00 / 1.73 | 0.00 / 1.29 | 0.00 / 1.29 |
| coverage/known-interface | 0.00 / 0.86 | 0.00 / 1.27 | 0.00 / 1.28 |
| coverage/readonly-string-bytes | 0.00 / 2.00 | 0.00 / 2.06 | 0.00 / 2.00 |
| coverage/runes-to-string | 0.00 / 21.34 | 0.00 / 8.56 | 0.00 / 7.52 |
| coverage/small-map | 0.00 / 9.75 | 0.00 / 9.97 | 0.00 / 10.25 |
| coverage/string-concat | 0.00 / 8.36 | 0.00 / 5.10 | 0.00 / 4.95 |
| coverage/string-to-bytes | 0.00 / 3.85 | 0.00 / 3.33 | 0.00 / 3.25 |
| coverage/string-to-runes | 0.00 / 11.49 | 0.00 / 4.92 | 0.00 / 4.51 |
| coverage/temporary-map-lookup | 0.00 / 8.12 | 0.00 / 7.33 | 0.00 / 7.04 |
| coverage/temporary-string-compare | 0.00 / 4.00 | 0.00 / 2.89 | 0.00 / 2.78 |
| defer/deferred | 0.00 / 4.32 | 0.00 / 11.69 | 0.00 / 11.95 |
| defer/explicit | 0.00 / 2.76 | 0.00 / 7.86 | 0.00 / 7.89 |
| mapkeys/Indirect key lookup | 0.00 / 17.65 | 272.01 / 149.19 | 0.00 / 47.23 |
| mapkeys/Map range | 0.00 / 26.69 | 112.01 / 69.34 | 0.00 / 23.59 |
| mapkeys/Scalar key control | 0.00 / 2.83 | 0.00 / 2.56 | 0.00 / 2.91 |
| mapkeys/String range | 0.00 / 3.89 | 31.97 / 19.25 | 0.00 / 3.44 |
| mapkeys/Struct key delete (missing) | 0.00 / 6.49 | 32.02 / 22.20 | 0.00 / 7.53 |
| mapkeys/Struct key lookup | 0.00 / 7.91 | 31.98 / 33.29 | 0.00 / 18.35 |
| mapkeys/Struct key update | 0.00 / 7.47 | 31.99 / 35.69 | 0.00 / 18.91 |
| vm-call/interpreter/public | 64.00 / N/A | 95.87 / N/A | 96.00 / N/A |
| vm-call/interpreter/session | 64.03 / N/A | 96.04 / N/A | 96.04 / N/A |
| vm-call/interpreter/stack | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |
| vm-call/native/public | 64.00 / N/A | 96.00 / N/A | 96.00 / N/A |
| vm-call/native/session | 64.03 / N/A | 96.00 / N/A | 96.02 / N/A |
| vm-call/native/stack | 0.00 / N/A | 0.00 / N/A | 0.00 / N/A |

Channel blocking exchanges become allocation-free, matching the integrated waiter-reuse contract. Struct/indirect map keys and map/string iterators also become allocation-free. Existing scalar-key and reusable scalar-call controls remain zero. Typed VM calls stay near 96 LLGo B/op versus 64 Go B/op; no improvement is claimed for them. Fresh zero-slice append still doubles within-LLGo bytes at 4096 and 65536 bytes, independently retaining [#3040](https://github.com/xgo-dev/llgo/issues/3040). Strict C11 is a separate compatibility fix: old compilation fails as previously reported, while candidate and Go execute the same checked C call successfully.

## Separate warm compiler-build cost

Five fixed blocks, seed 2026111015, source-identical coverage probe; `-O3 -lto=full -p=8` and tracing for every build. Three seed builds warm their archive namespaces before the 15 measured builds. Produced executables run and match all14 source checksums after the build interval. Source timestamps are unchanged; there is no cold-cache or general CI speedup claim.

| Compiler / persistence | Median build seconds | SSA construction events per build |
| --- | ---: | ---: |
| new-default | 6.8315 | 92 |
| new-opt-in | 6.9365 | 24 |
| old-default | 6.7725 | 92 |

Candidate persistence removes SSA events (92 -> 24) but the total build median is 6.9365 s versus 6.8315 s with persistence disabled. Old-default median is 6.7725 s. No same-binary build-time A/A or controlled equivalence is claimed from five blocks. Default persistence remains disabled; fewer SSA events alone is not an end-to-end build benefit. Individual trace stages can overlap and cannot be added/subtracted as wall-time partitions.

## Remaining work

Investigate the specific GC interpreter adverse result before upgrading the production pin. Attribute the ordinary-host four-worker slowdown with a separately scoped diagnostic if needed; its current controls are unresolved. Continue reducing component dynamic type/codec allocation and host synchronization, then validate candidate behavior on Linux amd64 and macOS arm64 before adoption. Full proposal implementation, allocation improvements, application performance, supported-platform acceptance and production release remain distinct gates.

All raw runs, logs, compiler bootstrap metadata, actual loader libraries, source/binary/fixture hashes, controllers and diagnostics are retained in artifacts/llgo-merged-impact-20261110. The machine record binds each cohort by SHA256; no earlier event is relabeled or pooled.
