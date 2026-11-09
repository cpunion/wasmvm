# Bounded P3 I/O worker reuse: 2026-11-10

[Machine record](data/performance-p3-worker-reuse-2026-11-10.json).

M4 Max/macOS27.0.1, base `132ac10df42e192ee197a78d09236743652ab03a`, head `dff5543412df551ab2d2e08110320c8da56f13ed`; Go1.27.3 and LLGo3d92a58 `-O3 -lto=full`. The base already includes [immutable Canonical copies](performance-immutable-memory-2026-11-10.md). This separate event changes only P3 external-I/O worker reuse and its tests/documentation. Drivers, guest fixtures, backends and immutable-copy paths are identical between roles.

**Observed result:** LLGo native/4-worker HTTP throughput1470 ->2084req/s, complete API2685.8 ->1886.3us. Paired warm-time ratio0.7019, pointwise95% interval[0.6796,0.7230]. Both controls for that comparison remain unresolved; **17/40 A/A controls resolve overall and controlled performance/release acceptance remains open**. The same event measures Wasmtime59833req/s and bare LLGo host9041req/s; substantial gaps remain.

**Resource tradeoff:** that LLGo native profile samples49 ->61 between-batch OS threads and159.7 ->177.0MiB observed peak RSS. Runtime-accounted bytes/request fall1273474 ->1202025. Thread snapshots count total process threads between application batches, not OS scheduler idle states. Idle retention defaults to four workers per IO owner, capped by the existing active-operation limit; it is an explicit bounded configuration. Go managed/LLGo BDWGC counters have different scopes; the Go reference counter excludes Rust/JIT allocation. RSS snapshots are through final pre-cleanup sampling, not an isolated allocation measurement.

## Mechanism and contracts

The pinned [LLGo native backend](https://github.com/xgo-dev/llgo/blob/3d92a58e09bd8971899c6eaf63437fbecd18d432/runtime/internal/runtime/proc_pthread.go) creates a native thread for a new goroutine. The former provider starts a goroutine per external I/O operation. Reusing provider workers therefore avoids repeated goroutine/thread creation; this is a source-based mechanism, not a claim that it explains the entire application gap. The independent Rust reference drives P3 work on one current-thread Tokio executor per Store, while retaining its Go C/Rust bridge in API timing. No upstream compiler patch or competitor fallback is involved.

Accepted operations retain the original twice-MaxStreams active quota. An idle worker receives one job; an already busy worker never queues another accepted blocking operation. Excess idle workers retire. Completed closures are cleared before parking. Admission and Close share the owner mutex; Close rejects new work, wakes idle workers and cancels operations, and WaitClosed joins both idle and active workers. The existing component task scheduler and per-Store serialization remain unchanged.

## Fixed complete event

Twenty randomized process blocks, primary/A/A aliases,16 clients,4 requests/client,1/4 workers,2 warmup and7 measured batches,64KiB payload, both Rust HTTP and composed fixtures. Seed2026111005; pointwise paired log-ratio bootstrap8000 draws with that seed. All1600 throughput processes,40 separate resource processes,80 qualification processes and7 actual loader probes complete, with716800 checked warm responses. All bytes, status201, header and trailer are checked. Counts/profiles/source are fixed before sampling. No outcome-based retiming, phase subtraction, portfolio family-wise claim or pooling with earlier events.

Preparation, first/warm batches, complete public API, request wall, readiness and joined cleanup remain distinct. Collection/close are inside the API timer; response validation follows it and remains inside request wall. Owned request construction is outside API time, while the reference constructs its request inside the bridge. Native-Go controls are ordinary echo/body-collection applications, not optimized native lower bounds. Reference/control binaries retain their original frozen source/build/library identities; owned base/head binaries are fresh builds.

## Every owned comparison

| Fixture / compiler / backend / workers | Base req/s | Head req/s | Warm head/base [95%] | Both A/A resolve | Base B/req | Head B/req |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| world-rust-composed-http/go/interpreter/1 | 1511 | 1573 | 0.9690 [0.9587,0.9811] | no | 756121 | 755652 |
| world-rust-composed-http/go/interpreter/4 | 5978 | 6199 | 0.9466 [0.8949,0.9779] | no | 756095 | 755632 |
| world-rust-composed-http/go/native/1 | 1968 | 2078 | 0.9479 [0.9393,0.9571] | yes | 752512 | 752057 |
| world-rust-composed-http/go/native/4 | 7125 | 7286 | 0.9807 [0.9640,0.9967] | no | 752476 | 752020 |
| world-rust-composed-http/llgo/interpreter/1 | 837 | 1018 | 0.8149 [0.8008,0.8263] | yes | 1293519 | 1222215 |
| world-rust-composed-http/llgo/interpreter/4 | 1482 | 2066 | 0.7193 [0.6989,0.7419] | no | 1294080 | 1222349 |
| world-rust-composed-http/llgo/native/1 | 935 | 1161 | 0.8059 [0.7930,0.8182] | no | 1290466 | 1218751 |
| world-rust-composed-http/llgo/native/4 | 1443 | 2015 | 0.7186 [0.7008,0.7384] | no | 1290735 | 1218972 |
| world-rust-http/go/interpreter/1 | 1597 | 1677 | 0.9540 [0.9439,0.9648] | no | 740249 | 739787 |
| world-rust-http/go/interpreter/4 | 6251 | 6412 | 0.9747 [0.9685,0.9820] | no | 740216 | 739762 |
| world-rust-http/go/native/1 | 2069 | 2213 | 0.9339 [0.9255,0.9417] | yes | 740913 | 740446 |
| world-rust-http/go/native/4 | 7448 | 7699 | 0.9603 [0.9477,0.9743] | no | 740874 | 740412 |
| world-rust-http/llgo/interpreter/1 | 873 | 1066 | 0.8101 [0.7899,0.8237] | yes | 1272242 | 1200471 |
| world-rust-http/llgo/interpreter/4 | 1482 | 2085 | 0.7181 [0.7012,0.7354] | no | 1272378 | 1200648 |
| world-rust-http/llgo/native/1 | 976 | 1214 | 0.8223 [0.7993,0.8572] | no | 1273013 | 1201796 |
| world-rust-http/llgo/native/4 | 1470 | 2084 | 0.7019 [0.6796,0.7230] | no | 1273474 | 1202025 |

## All profiles and controls

| Profile | req/s | Complete API us | Request p99 us | Runtime B/req | Idle threads | Peak RSS MiB | A/A95% | Resolved |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| base/world-rust-composed-http/go/interpreter/1 | 1511 | 659.241 | 11920.396 | 756121 | 7 | 107.922 | [1.0017,1.0241] | yes |
| base/world-rust-composed-http/go/interpreter/4 | 5978 | 648.902 | 4665.062 | 756095 | 8 | 144.172 | [0.9216,1.0035] | no |
| base/world-rust-composed-http/go/native/1 | 1968 | 505.644 | 8979.041 | 752512 | 7 | 146.109 | [0.9985,1.0173] | yes |
| base/world-rust-composed-http/go/native/4 | 7125 | 538.308 | 3923.041 | 752476 | 8 | 161.859 | [1.0012,1.1947] | no |
| base/world-rust-composed-http/llgo/interpreter/1 | 837 | 1183.179 | 22358.854 | 1293519 | 40 | 166.625 | [0.9757,0.9978] | yes |
| base/world-rust-composed-http/llgo/interpreter/4 | 1482 | 2668.039 | 13969.604 | 1294080 | 49 | 180.469 | [0.9796,1.0252] | no |
| base/world-rust-composed-http/llgo/native/1 | 935 | 1062.154 | 20168.500 | 1290466 | 40 | 180.594 | [1.0031,1.0353] | no |
| base/world-rust-composed-http/llgo/native/4 | 1443 | 2732.940 | 13997.459 | 1290735 | 49 | 193.078 | [0.9874,1.0670] | no |
| base/world-rust-http/go/interpreter/1 | 1597 | 623.551 | 10931.417 | 740249 | 7 | 83.484 | [0.9893,1.0543] | no |
| base/world-rust-http/go/interpreter/4 | 6251 | 621.694 | 4098.626 | 740216 | 8 | 102.219 | [0.9866,1.0069] | yes |
| base/world-rust-http/go/native/1 | 2069 | 480.599 | 8400.167 | 740913 | 7 | 112.000 | [0.9910,1.0053] | yes |
| base/world-rust-http/go/native/4 | 7448 | 518.716 | 3569.916 | 740874 | 7 | 124.859 | [1.0011,1.0363] | no |
| base/world-rust-http/llgo/interpreter/1 | 873 | 1137.587 | 21504.312 | 1272242 | 40 | 140.047 | [0.9884,1.0072] | yes |
| base/world-rust-http/llgo/interpreter/4 | 1482 | 2667.029 | 13682.666 | 1272378 | 49 | 140.672 | [0.9791,1.0157] | yes |
| base/world-rust-http/llgo/native/1 | 976 | 1014.103 | 19014.959 | 1273013 | 40 | 151.859 | [0.9986,1.0357] | no |
| base/world-rust-http/llgo/native/4 | 1470 | 2685.781 | 13357.542 | 1273474 | 49 | 159.703 | [0.9567,1.0089] | no |
| head/world-rust-composed-http/go/interpreter/1 | 1573 | 632.985 | 11514.875 | 755652 | 7 | 104.859 | [0.9910,1.0316] | no |
| head/world-rust-composed-http/go/interpreter/4 | 6199 | 623.736 | 4479.667 | 755632 | 7 | 141.906 | [0.9892,1.0117] | yes |
| head/world-rust-composed-http/go/native/1 | 2078 | 478.550 | 8700.458 | 752057 | 7 | 146.750 | [0.9950,1.0109] | yes |
| head/world-rust-composed-http/go/native/4 | 7286 | 529.022 | 3992.208 | 752020 | 8 | 162.094 | [0.9792,1.0192] | yes |
| head/world-rust-composed-http/llgo/interpreter/1 | 1018 | 971.041 | 19205.625 | 1222215 | 43 | 163.828 | [0.9966,1.0131] | yes |
| head/world-rust-composed-http/llgo/interpreter/4 | 2066 | 1907.334 | 11074.646 | 1222349 | 61 | 183.828 | [1.0017,1.1228] | no |
| head/world-rust-composed-http/llgo/native/1 | 1161 | 854.667 | 16899.666 | 1218751 | 43 | 196.734 | [0.9876,1.0238] | yes |
| head/world-rust-composed-http/llgo/native/4 | 2015 | 1955.675 | 11064.688 | 1218972 | 61 | 209.344 | [0.9809,1.0679] | no |
| head/world-rust-http/go/interpreter/1 | 1677 | 593.602 | 10324.917 | 739787 | 7 | 79.688 | [0.9883,1.0067] | yes |
| head/world-rust-http/go/interpreter/4 | 6412 | 604.367 | 4022.688 | 739762 | 7 | 102.859 | [0.9912,1.0267] | no |
| head/world-rust-http/go/native/1 | 2213 | 449.266 | 7967.146 | 740446 | 7 | 113.359 | [0.9932,1.0081] | yes |
| head/world-rust-http/go/native/4 | 7699 | 502.569 | 3493.312 | 740412 | 7 | 123.719 | [0.9946,1.0205] | yes |
| head/world-rust-http/llgo/interpreter/1 | 1066 | 929.559 | 17890.125 | 1200471 | 43 | 140.234 | [0.9899,1.0068] | yes |
| head/world-rust-http/llgo/interpreter/4 | 2085 | 1894.010 | 10390.417 | 1200648 | 61 | 140.688 | [0.9880,1.0339] | no |
| head/world-rust-http/llgo/native/1 | 1214 | 812.790 | 15828.938 | 1201796 | 43 | 152.047 | [0.9489,1.0060] | no |
| head/world-rust-http/llgo/native/4 | 2084 | 1886.301 | 10327.792 | 1202025 | 61 | 177.000 | [0.9721,1.0225] | no |
| host/native-go/go/host/1 | 41419 | 20.512 | 520.146 | 277914 | 7 | 16.984 | [0.9881,1.0263] | no |
| host/native-go/go/host/4 | 44784 | 76.425 | 563.708 | 277917 | 6 | 16.281 | [0.8801,0.9966] | no |
| host/native-go/llgo/host/1 | 8119 | 117.524 | 3703.479 | 732582 | 37 | 68.094 | [0.9514,1.0132] | no |
| host/native-go/llgo/host/4 | 9041 | 389.169 | 3473.438 | 732586 | 40 | 69.172 | [0.9544,1.0504] | no |
| reference/world-rust-composed-http/go/compiler/1 | 15020 | 64.170 | 1176.333 | 32 | 7 | 32.125 | [0.9814,1.0035] | yes |
| reference/world-rust-composed-http/go/compiler/4 | 51620 | 72.716 | 376.604 | 32 | 7 | 33.172 | [0.9700,1.0335] | no |
| reference/world-rust-http/go/compiler/1 | 17446 | 54.908 | 1025.771 | 32 | 7 | 29.734 | [0.9746,1.0140] | no |
| reference/world-rust-http/go/compiler/4 | 59833 | 62.322 | 335.584 | 32 | 7 | 30.641 | [1.0011,1.0325] | no |

## Complete lifecycle intervals

Median milliseconds over20 primary process reductions. Warm is a process-mean batch; phases are never added/subtracted to infer an unmeasured cost. The machine record also retains all32 paired head/reference intervals.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1 | 5.4792 | 0.0601 | 0.0003 | 25.9281 | 0.0356 | 0.2276 | 0.0005 | N/A | 0.0121 | 45.9834 | 42.3583 | 0.0299 | 463.0664 |
| base/world-rust-composed-http/go/interpreter/4 | 5.5814 | 0.0603 | 0.0003 | 25.9838 | 0.0538 | 0.6870 | 0.0007 | N/A | 0.0151 | 12.9536 | 10.7068 | 0.0523 | 146.0210 |
| base/world-rust-composed-http/go/native/1 | 5.4889 | 0.0617 | 0.0003 | 217.4287 | 0.0348 | 0.2105 | 0.0004 | N/A | 0.0123 | 33.4258 | 32.5239 | 0.0362 | 553.5195 |
| base/world-rust-composed-http/go/native/4 | 5.4477 | 0.0620 | 0.0003 | 218.5209 | 0.0527 | 0.6593 | 0.0008 | N/A | 0.0138 | 10.1291 | 8.9829 | 0.0624 | 318.8697 |
| base/world-rust-composed-http/llgo/interpreter/1 | 8.9169 | 0.0711 | 0.0002 | 44.2145 | 0.1119 | 0.4297 | 0.0004 | N/A | 0.2781 | 79.0156 | 76.4592 | 0.2155 | 824.9885 |
| base/world-rust-composed-http/llgo/interpreter/4 | 8.9431 | 0.0707 | 0.0002 | 43.6304 | 0.2450 | 1.4899 | 0.0009 | N/A | 0.3021 | 51.1993 | 43.1791 | 0.4508 | 498.0485 |
| base/world-rust-composed-http/llgo/native/1 | 8.8522 | 0.0721 | 0.0003 | 324.7782 | 0.1021 | 0.2924 | 0.0005 | N/A | 0.2794 | 69.1163 | 68.4362 | 0.1963 | 1024.0165 |
| base/world-rust-composed-http/llgo/native/4 | 8.9459 | 0.0715 | 0.0002 | 324.5605 | 0.2414 | 0.9787 | 0.0009 | N/A | 0.3019 | 44.2651 | 44.3470 | 0.4557 | 780.5981 |
| base/world-rust-http/go/interpreter/1 | 5.4379 | 0.0547 | 0.0003 | 18.1344 | 0.0389 | 0.1669 | 0.0004 | N/A | 0.0151 | 42.9883 | 40.0700 | 0.0290 | 430.2803 |
| base/world-rust-http/go/interpreter/4 | 5.5377 | 0.0549 | 0.0003 | 18.1491 | 0.0587 | 0.5200 | 0.0008 | N/A | 0.0151 | 12.4108 | 10.2378 | 0.0399 | 132.4358 |
| base/world-rust-http/go/native/1 | 5.4320 | 0.0548 | 0.0003 | 149.1323 | 0.0337 | 0.1599 | 0.0004 | N/A | 0.0123 | 31.3776 | 30.9261 | 0.0345 | 466.5525 |
| base/world-rust-http/go/native/4 | 5.4668 | 0.0537 | 0.0003 | 149.4643 | 0.0537 | 0.4778 | 0.0008 | N/A | 0.0138 | 9.0965 | 8.5927 | 0.0531 | 246.2986 |
| base/world-rust-http/llgo/interpreter/1 | 8.9883 | 0.0625 | 0.0002 | 28.9468 | 0.1089 | 2.6518 | 0.0005 | N/A | 0.3034 | 75.4534 | 73.3430 | 0.2045 | 777.0944 |
| base/world-rust-http/llgo/interpreter/4 | 8.8901 | 0.0651 | 0.0002 | 29.4545 | 0.2481 | 3.5811 | 0.0011 | N/A | 0.3072 | 45.5806 | 43.1934 | 0.4385 | 482.1161 |
| base/world-rust-http/llgo/native/1 | 8.8609 | 0.0605 | 0.0002 | 225.0590 | 0.1109 | 0.2084 | 0.0005 | N/A | 0.2800 | 67.5600 | 65.5767 | 0.2045 | 896.9557 |
| base/world-rust-http/llgo/native/4 | 8.9768 | 0.0629 | 0.0002 | 224.4913 | 0.2372 | 0.7260 | 0.0010 | N/A | 0.3037 | 44.7097 | 43.5315 | 0.4416 | 678.3806 |
| head/world-rust-composed-http/go/interpreter/1 | 5.4288 | 0.0610 | 0.0003 | 26.4079 | 0.0377 | 0.2338 | 0.0004 | N/A | 0.0125 | 44.6148 | 40.6827 | 0.0323 | 448.0160 |
| head/world-rust-composed-http/go/interpreter/4 | 5.5403 | 0.0622 | 0.0003 | 26.2733 | 0.0549 | 0.6906 | 0.0007 | N/A | 0.0165 | 12.7393 | 10.3235 | 0.0523 | 142.2925 |
| head/world-rust-composed-http/go/native/1 | 5.4046 | 0.0639 | 0.0003 | 217.4592 | 0.0343 | 0.2148 | 0.0004 | N/A | 0.0120 | 32.8084 | 30.7966 | 0.0360 | 537.0745 |
| head/world-rust-composed-http/go/native/4 | 5.5682 | 0.0617 | 0.0003 | 216.5928 | 0.0500 | 0.6417 | 0.0007 | N/A | 0.0133 | 9.8981 | 8.7844 | 0.0640 | 315.5570 |
| head/world-rust-composed-http/llgo/interpreter/1 | 8.9270 | 0.0700 | 0.0002 | 44.1604 | 0.1119 | 0.4488 | 0.0005 | N/A | 0.2855 | 65.3331 | 62.8560 | 0.2415 | 689.3465 |
| head/world-rust-composed-http/llgo/interpreter/4 | 8.8986 | 0.0690 | 0.0002 | 43.6384 | 0.2475 | 1.5163 | 0.0009 | N/A | 0.3065 | 36.7193 | 30.9704 | 0.6169 | 370.4071 |
| head/world-rust-composed-http/llgo/native/1 | 8.9171 | 0.0673 | 0.0002 | 324.7303 | 0.1077 | 0.2850 | 0.0005 | N/A | 0.2770 | 56.1070 | 55.1238 | 0.2256 | 888.1487 |
| head/world-rust-composed-http/llgo/native/4 | 8.9345 | 0.0691 | 0.0002 | 325.3479 | 0.2430 | 0.9833 | 0.0010 | N/A | 0.3010 | 33.5170 | 31.7648 | 0.6269 | 659.2224 |
| head/world-rust-http/go/interpreter/1 | 5.5055 | 0.0513 | 0.0003 | 17.8870 | 0.0379 | 0.1764 | 0.0004 | N/A | 0.0130 | 41.6971 | 38.1599 | 0.0310 | 412.3719 |
| head/world-rust-http/go/interpreter/4 | 5.5330 | 0.0536 | 0.0003 | 18.0477 | 0.0610 | 0.5222 | 0.0008 | N/A | 0.0146 | 11.7349 | 9.9808 | 0.0413 | 129.5489 |
| head/world-rust-http/go/native/1 | 5.4494 | 0.0543 | 0.0003 | 150.9775 | 0.0314 | 0.1641 | 0.0005 | N/A | 0.0130 | 30.0929 | 28.9200 | 0.0359 | 449.6665 |
| head/world-rust-http/go/native/4 | 5.4356 | 0.0515 | 0.0003 | 150.1923 | 0.0493 | 0.4697 | 0.0008 | N/A | 0.0138 | 8.9075 | 8.3123 | 0.0527 | 243.4625 |
| head/world-rust-http/llgo/interpreter/1 | 8.9282 | 0.0645 | 0.0002 | 29.7808 | 0.1080 | 3.1248 | 0.0005 | N/A | 0.3040 | 62.3893 | 60.0154 | 0.2320 | 650.5317 |
| head/world-rust-http/llgo/interpreter/4 | 8.9590 | 0.0597 | 0.0002 | 29.2184 | 0.2577 | 3.8314 | 0.0010 | N/A | 0.3051 | 32.9823 | 30.7029 | 0.6019 | 359.7670 |
| head/world-rust-http/llgo/native/1 | 8.9222 | 0.0647 | 0.0002 | 225.7434 | 0.1027 | 0.2125 | 0.0005 | N/A | 0.2775 | 55.6421 | 52.7231 | 0.2352 | 769.6226 |
| head/world-rust-http/llgo/native/4 | 8.9054 | 0.0647 | 0.0002 | 224.4993 | 0.2484 | 0.7726 | 0.0010 | N/A | 0.3207 | 32.5244 | 30.7062 | 0.6315 | 548.8988 |
| host/native-go/go/host/1 | 4.8728 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0130 | 1.8011 | 1.5452 | 0.0062 | 22.2570 |
| host/native-go/go/host/4 | 5.1431 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0165 | 1.3140 | 1.4291 | 0.0064 | 20.4301 |
| host/native-go/llgo/host/1 | 8.6714 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2759 | 9.8821 | 7.8832 | 0.0767 | 90.7011 |
| host/native-go/llgo/host/4 | 8.5830 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3092 | 10.1083 | 7.0804 | 0.0865 | 86.3026 |
| reference/world-rust-composed-http/go/compiler/1 | 5.8623 | 0.0632 | 0.1323 | 72.7903 | 0.1101 | 0.1362 | 0.0065 | 0.0281 | 0.0345 | 4.7640 | 4.2610 | 0.2308 | 123.7462 |
| reference/world-rust-composed-http/go/compiler/4 | 5.8371 | 0.0599 | 0.1293 | 72.4816 | 0.2166 | 0.2472 | 0.0081 | 0.0410 | 0.0470 | 1.5422 | 1.2398 | 0.3721 | 93.4862 |
| reference/world-rust-http/go/compiler/1 | 5.9075 | 0.0530 | 0.1293 | 49.9497 | 0.1113 | 0.1159 | 0.0065 | 0.0270 | 0.0338 | 4.1811 | 3.6686 | 0.1577 | 95.3268 |
| reference/world-rust-http/go/compiler/4 | 5.8240 | 0.0555 | 0.1312 | 49.8498 | 0.2191 | 0.1930 | 0.0079 | 0.0420 | 0.0455 | 1.3562 | 1.0696 | 0.2347 | 68.9794 |

## Validation and remaining work

Full normal local Go owned-package tests pass with all31 measured packages >=90% (minimum90.75%). P3 race tests, P3 measurement race tests, vet, pinned LLGo complete P3 execution and owned measurement execution on both backends pass. Worker tests exercise reuse, quota rejection, simultaneous blocked jobs, idle retirement, close waiting and concurrent submit/close. This supplements the existing ownership/cancellation/world inventories; final-head full Go/LLGo CI on both targets is still required before merge.

An earlier preflight correctly blocked sampling while an unrelated local Go compiler/test was active. The actual pre-start process table is clear before this one complete event; all owned build/test/probe/archive processes exited first. GUI/background applications, affinity, frequency and power remain uncontrolled. Do not promote this workstation event to a release acceptance run. The earlier immutable-copy event, its adverse timings, incomplete prefix and diagnostics remain independently recorded.

All40 profiles,1600 process reductions, resource snapshots, controls,16 owned and32 reference comparisons are retained. Raw result SHA256 `2151ed9ead75a01ebc8b900e5783ccb91f390b1a1e0b92e6b216717ddbc84230` in `artifacts/p3-worker-reuse-20261110`. The larger P3 component/IO and host-runtime gaps, native compiler deficits, both-target controlled acceptance and production publication remain open. LLGo3034/3035 are already being solved upstream; this contribution neither duplicates those issues nor changes the pinned compiler.
