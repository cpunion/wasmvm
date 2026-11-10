# Component structural comparison: stack storage, 2026-11-10

[Complete machine record](data/performance-component-type-comparison-2026-11-10.json).

## Change and observed result

Each structural comparison now keeps its first 16 distinct node pairs on the stack and spills to a map only when needed. This removes the eager per-comparison map from the common path. There is no cross-call or Store-global cache: resource identity predicates, cycle detection, exact node charging and mismatch ordering keep their original semantics. Large graphs retain a map fallback with its initial capacity based on the actual 17 entries.

The pinned compiler remains LLGo `3d92a58`, with `-O3 -lto=full`. This measures project source changes, separately from the [full LLGo #3012 candidate study](performance-llgo-merged-2026-11-10.md); its GC finding and compiler adoption gate remain open.

HTTP fixture, 64 KiB, four workers, bounded buffered response API:

| Path | Base req/s | Head req/s | Base B/request | Head B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 10803 | 11180 | 359680 | 348711 |
| wasmvm go / native | 13883 | 14686 | 359824 | 348857 |
| wasmvm llgo / interpreter | 3598 | 4090 | 560475 | 532261 |
| wasmvm llgo / native | 3614 | 3979 | 561035 | 532685 |
| Wasmtime 48.0.1 | N/A | 59132 | N/A | 32 |
| Ordinary native-Go application / Go | N/A | 45577 | N/A | 277918 |
| Ordinary native-Go application / LLGo | N/A | 9176 | N/A | 732580 |

**8/40 timing controls resolve** within ±2.5%; 2/16 source pairs resolve both controls. These are advisory workstation observations. They do not close controlled performance or release acceptance, establish a calibrated competitor ranking, or partition the host-runtime gap. All unchanged/adverse observations and reference intervals remain in the machine record.

The one-worker LLGo HTTP paired warm-time estimates are adverse: interpreter 1.0223 [0.9465,1.1705] and native 1.1018 [0.9443,1.4842]. Both have unresolved controls. Their positive throughput medians do not override those paired observations.

The ordinary native-Go application is an application control, not an optimized native lower bound. Go managed bytes and LLGo collector accounting differ; LLGo Mallocs are unavailable. Rust/JIT allocation is excluded from Go counters. Resource instrumentation runs in a separate cohort.

## Shared graph microbenchmark

Five fixed randomized primary/A/A blocks, 50,000 comparisons per process, GOMAXPROCS=1. These are repeated-comparison kernel timings, separately from process startup. The test setup resolves schemas before timing. Every case and its pointwise 95% paired log-ratio bootstrap interval is retained.

| Compiler | Nodes | Base ns/call | Head ns/call | Base B/call | Head B/call | Head/base time [95%] | Both A/A |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| go | 2 | 40.76 | 19.16 | 0 | 0 | 0.4654 [0.4503,0.4786] | no |
| go | 16 | 894.1 | 142.6 | 1384 | 0 | 0.1583 [0.1556,0.1604] | no |
| go | 17 | 945.8 | 563.5 | 1384 | 984 | 0.5952 [0.5863,0.6064] | no |
| go | 64 | 4395 | 4026 | 6696 | 6296 | 0.9199 [0.9033,0.9371] | yes |
| llgo | 2 | 74.31 | 17.54 | 127 | 0 | 0.2406 [0.2305,0.2577] | no |
| llgo | 16 | 1142 | 147.9 | 2172 | 0 | 0.1298 [0.1218,0.1373] | no |
| llgo | 17 | 1159 | 718.7 | 2239 | 1415 | 0.6252 [0.6128,0.6357] | no |
| llgo | 64 | 5203 | 4841 | 9466 | 8642 | 0.9264 [0.9175,0.9357] | yes |

Reproduce a kernel sample on either committed source (shared tests begin at `6c871d4`):

```sh
go test ./internal/component/validate -run '^$' -bench '^BenchmarkEqualityResolvedGraphs$' -benchtime=50000x -count=1 -benchmem
llgo test -O3 -lto=full ./internal/component/validate -run '^$' -bench '^BenchmarkEqualityResolvedGraphs$' -benchtime=50000x -count=1 -benchmem
```

An initial 32-entry spill reservation reduced LLGo bytes but increased Go bytes at 17 nodes (1384 -> 1880 B/call). That complete exploratory micro run is retained separately. The final implementation reserves the actual 17 entries and observes 1384 -> 984 Go B/call and 2239 -> 1415 LLGo B/call. Final binaries were freshly built from the committed source. Only the final implementation enters the macro event; no timed macro prefix, favorable repeat or pooling is used.

## Source changes in all service profiles

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A | Base B/request | Head B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| head/world-rust-composed-http/go/interpreter/1 | 2707 | 2833 | 0.9649 [0.9500,0.9822] | yes | 375558 | 364205 |
| head/world-rust-composed-http/go/interpreter/4 | 9738 | 10294 | 0.9451 [0.9162,0.9695] | no | 375558 | 364205 |
| head/world-rust-composed-http/go/native/1 | 3766 | 3976 | 0.9286 [0.8892,0.9598] | no | 371429 | 360076 |
| head/world-rust-composed-http/go/native/4 | 12780 | 13448 | 0.9443 [0.9212,0.9641] | no | 371431 | 360079 |
| head/world-rust-composed-http/llgo/interpreter/1 | 2221 | 2289 | 0.8592 [0.6715,0.9810] | no | 582405 | 553242 |
| head/world-rust-composed-http/llgo/interpreter/4 | 3631 | 3971 | 0.8974 [0.8777,0.9165] | no | 582283 | 553177 |
| head/world-rust-composed-http/llgo/native/1 | 2728 | 2857 | 0.9475 [0.8801,1.0281] | no | 578131 | 548949 |
| head/world-rust-composed-http/llgo/native/4 | 3455 | 3838 | 0.9973 [0.8833,1.2374] | no | 577983 | 548835 |
| head/world-rust-http/go/interpreter/1 | 3004 | 3159 | 0.9107 [0.8144,0.9685] | yes | 359680 | 348712 |
| head/world-rust-http/go/interpreter/4 | 10803 | 11180 | 0.9675 [0.9548,0.9820] | no | 359680 | 348711 |
| head/world-rust-http/go/native/1 | 4243 | 4462 | 0.9418 [0.9082,0.9684] | no | 359823 | 348854 |
| head/world-rust-http/go/native/4 | 13883 | 14686 | 0.9061 [0.7750,0.9931] | no | 359824 | 348857 |
| head/world-rust-http/llgo/interpreter/1 | 2463 | 2584 | 1.0223 [0.9465,1.1705] | no | 560699 | 532350 |
| head/world-rust-http/llgo/interpreter/4 | 3598 | 4090 | 0.8798 [0.8397,0.9130] | no | 560475 | 532261 |
| head/world-rust-http/llgo/native/1 | 2987 | 3127 | 1.1018 [0.9443,1.4842] | no | 561154 | 532800 |
| head/world-rust-http/llgo/native/4 | 3614 | 3979 | 0.9077 [0.8779,0.9386] | no | 561035 | 532685 |

## Throughput, latency and resources

| Profile | req/s | Complete API us | p99 request us | Runtime B/request | Between-batch threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/1 | 2707 | 366.382 | 6879.251 | 375558 | 7 | 98.312 | yes |
| base/world-rust-composed-http/go/interpreter/4 | 9738 | 392.565 | 3331.125 | 375558 | 7 | 125.219 | no |
| base/world-rust-composed-http/go/native/1 | 3766 | 262.203 | 4984.167 | 371429 | 7 | 143.281 | no |
| base/world-rust-composed-http/go/native/4 | 12780 | 299.729 | 2828.250 | 371431 | 7 | 164.734 | no |
| base/world-rust-composed-http/llgo/interpreter/1 | 2221 | 442.854 | 10334.604 | 582405 | 42 | 149.219 | no |
| base/world-rust-composed-http/llgo/interpreter/4 | 3631 | 1083.468 | 7449.333 | 582283 | 57 | 178.188 | no |
| base/world-rust-composed-http/llgo/native/1 | 2728 | 358.798 | 9121.750 | 578131 | 42 | 180.594 | no |
| base/world-rust-composed-http/llgo/native/4 | 3455 | 1134.705 | 7753.437 | 577983 | 57 | 209.234 | no |
| base/world-rust-http/go/interpreter/1 | 3004 | 329.886 | 6123.021 | 359680 | 7 | 83.969 | yes |
| base/world-rust-http/go/interpreter/4 | 10803 | 357.237 | 2940.459 | 359680 | 7 | 91.641 | no |
| base/world-rust-http/go/native/1 | 4243 | 232.317 | 4308.333 | 359823 | 8 | 111.656 | no |
| base/world-rust-http/go/native/4 | 13883 | 276.085 | 2493.125 | 359824 | 8 | 129.297 | no |
| base/world-rust-http/llgo/interpreter/1 | 2463 | 398.856 | 9504.333 | 560699 | 42 | 130.766 | yes |
| base/world-rust-http/llgo/interpreter/4 | 3598 | 1086.686 | 7185.792 | 560475 | 57 | 140.953 | no |
| base/world-rust-http/llgo/native/1 | 2987 | 328.517 | 8231.771 | 561154 | 42 | 151.016 | yes |
| base/world-rust-http/llgo/native/4 | 3614 | 1087.195 | 7357.333 | 561035 | 57 | 176.562 | no |
| head/world-rust-composed-http/go/interpreter/1 | 2833 | 350.094 | 6492.250 | 364205 | 7 | 101.828 | yes |
| head/world-rust-composed-http/go/interpreter/4 | 10294 | 371.971 | 3276.042 | 364205 | 7 | 124.812 | no |
| head/world-rust-composed-http/go/native/1 | 3976 | 248.145 | 4771.980 | 360076 | 7 | 143.703 | no |
| head/world-rust-composed-http/go/native/4 | 13448 | 282.794 | 3132.750 | 360079 | 7 | 162.391 | no |
| head/world-rust-composed-http/llgo/interpreter/1 | 2289 | 427.634 | 10192.750 | 553242 | 42 | 152.656 | no |
| head/world-rust-composed-http/llgo/interpreter/4 | 3971 | 981.975 | 7430.062 | 553177 | 57 | 175.188 | no |
| head/world-rust-composed-http/llgo/native/1 | 2857 | 343.783 | 8833.291 | 548949 | 42 | 180.719 | no |
| head/world-rust-composed-http/llgo/native/4 | 3838 | 1026.062 | 7303.771 | 548835 | 57 | 209.453 | no |
| head/world-rust-http/go/interpreter/1 | 3159 | 313.627 | 5892.417 | 348712 | 7 | 79.141 | yes |
| head/world-rust-http/go/interpreter/4 | 11180 | 344.795 | 2876.209 | 348711 | 8 | 89.188 | yes |
| head/world-rust-http/go/native/1 | 4462 | 220.500 | 4281.812 | 348854 | 7 | 118.078 | yes |
| head/world-rust-http/go/native/4 | 14686 | 261.543 | 2424.771 | 348857 | 7 | 126.062 | no |
| head/world-rust-http/llgo/interpreter/1 | 2584 | 380.168 | 9033.354 | 532350 | 42 | 127.141 | no |
| head/world-rust-http/llgo/interpreter/4 | 4090 | 955.098 | 6616.625 | 532261 | 57 | 140.891 | no |
| head/world-rust-http/llgo/native/1 | 3127 | 313.658 | 7937.355 | 532800 | 42 | 167.203 | no |
| head/world-rust-http/llgo/native/4 | 3979 | 983.670 | 6795.583 | 532685 | 57 | 176.656 | no |
| host/native-go/go/host/1 | 41284 | 20.561 | 507.667 | 277914 | 6 | 16.266 | no |
| host/native-go/go/host/4 | 45577 | 75.712 | 558.854 | 277918 | 6 | 16.312 | no |
| host/native-go/llgo/host/1 | 8299 | 115.024 | 3641.375 | 732589 | 37 | 68.109 | no |
| host/native-go/llgo/host/4 | 9176 | 392.507 | 3602.104 | 732580 | 40 | 69.188 | no |
| reference/world-rust-composed-http/go/compiler/1 | 15083 | 63.826 | 1182.167 | 32 | 7 | 32.469 | no |
| reference/world-rust-composed-http/go/compiler/4 | 51911 | 72.404 | 370.396 | 32 | 9 | 33.922 | no |
| reference/world-rust-http/go/compiler/1 | 17228 | 55.605 | 1097.750 | 32 | 7 | 29.734 | no |
| reference/world-rust-http/go/compiler/4 | 59132 | 63.300 | 333.896 | 32 | 7 | 31.016 | no |

## Separate lifecycle phases

Median milliseconds across 20 primary process reductions. Process wall includes protocol parking; it is not standalone CLI latency. Overlapping phases and different cohorts are not subtracted to infer costs. Compilation here includes preparing executable guest code for the chosen backend; source-host builds occur before the event.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1 | 5.3339 | 0.0565 | 0.0003 | 25.7510 | 0.0357 | 0.2252 | 0.0004 | N/A | 0.0123 | 26.6548 | 23.6382 | 0.0328 | 273.4668 |
| base/world-rust-composed-http/go/interpreter/4 | 5.3623 | 0.0599 | 0.0003 | 25.6634 | 0.0556 | 0.7104 | 0.0007 | N/A | 0.0170 | 8.7897 | 6.5720 | 0.0480 | 103.3980 |
| base/world-rust-composed-http/go/native/1 | 5.2614 | 0.0578 | 0.0003 | 215.2083 | 0.0350 | 0.2104 | 0.0004 | N/A | 0.0121 | 18.1214 | 16.9961 | 0.0408 | 393.7779 |
| base/world-rust-composed-http/go/native/4 | 5.3711 | 0.0579 | 0.0003 | 215.8656 | 0.0503 | 0.6367 | 0.0007 | N/A | 0.0136 | 5.7695 | 5.0078 | 0.0665 | 276.3931 |
| base/world-rust-composed-http/llgo/interpreter/1 | 8.7705 | 0.0676 | 0.0003 | 42.3249 | 0.1036 | 0.4352 | 0.0005 | N/A | 0.2906 | 33.3119 | 28.8136 | 0.2371 | 348.4509 |
| base/world-rust-composed-http/llgo/interpreter/4 | 8.8078 | 0.0659 | 0.0002 | 44.2964 | 0.2331 | 1.5398 | 0.0010 | N/A | 0.3141 | 20.1459 | 17.6257 | 0.5678 | 237.7405 |
| base/world-rust-composed-http/llgo/native/1 | 8.7096 | 0.0696 | 0.0003 | 326.2463 | 0.0993 | 0.2823 | 0.0005 | N/A | 0.2805 | 26.5707 | 23.4637 | 0.2309 | 575.2696 |
| base/world-rust-composed-http/llgo/native/4 | 8.6585 | 0.0701 | 0.0003 | 325.3931 | 0.2330 | 0.9847 | 0.0010 | N/A | 0.3131 | 20.3033 | 18.5249 | 0.5756 | 525.8820 |
| base/world-rust-http/go/interpreter/1 | 5.3323 | 0.0499 | 0.0003 | 17.7180 | 0.0354 | 0.1692 | 0.0005 | N/A | 0.0134 | 23.9657 | 21.3075 | 0.0290 | 240.9363 |
| base/world-rust-http/go/interpreter/4 | 5.3102 | 0.0514 | 0.0003 | 17.7793 | 0.0589 | 0.7788 | 0.0009 | N/A | 0.0164 | 7.6770 | 5.9240 | 0.0440 | 87.6295 |
| base/world-rust-http/go/native/1 | 5.3122 | 0.0535 | 0.0003 | 148.0452 | 0.0316 | 0.1620 | 0.0005 | N/A | 0.0123 | 16.4324 | 15.0828 | 0.0376 | 308.5750 |
| base/world-rust-http/go/native/4 | 5.4289 | 0.0549 | 0.0003 | 149.4369 | 0.0476 | 0.4774 | 0.0007 | N/A | 0.0135 | 5.5921 | 4.6100 | 0.0547 | 205.5046 |
| base/world-rust-http/llgo/interpreter/1 | 8.7248 | 0.0588 | 0.0002 | 29.1374 | 0.1001 | 2.5790 | 0.0005 | N/A | 0.2820 | 29.5400 | 25.9850 | 0.2342 | 308.7045 |
| base/world-rust-http/llgo/interpreter/4 | 8.6491 | 0.0591 | 0.0003 | 29.9818 | 0.2469 | 3.7839 | 0.0012 | N/A | 0.3056 | 18.5142 | 17.7903 | 0.5756 | 222.5418 |
| base/world-rust-http/llgo/native/1 | 8.7479 | 0.0584 | 0.0003 | 225.1548 | 0.1004 | 0.2134 | 0.0005 | N/A | 0.2765 | 23.9859 | 21.4294 | 0.2325 | 455.1728 |
| base/world-rust-http/llgo/native/4 | 8.7606 | 0.0611 | 0.0003 | 224.7647 | 0.2384 | 0.7275 | 0.0011 | N/A | 0.3131 | 18.3819 | 17.7081 | 0.5859 | 418.1304 |
| head/world-rust-composed-http/go/interpreter/1 | 5.3052 | 0.0566 | 0.0003 | 25.5954 | 0.0358 | 0.2248 | 0.0005 | N/A | 0.0128 | 25.7361 | 22.5947 | 0.0295 | 262.5473 |
| head/world-rust-composed-http/go/interpreter/4 | 5.3415 | 0.0584 | 0.0003 | 25.5660 | 0.0534 | 0.6746 | 0.0007 | N/A | 0.0161 | 8.6815 | 6.2175 | 0.0490 | 100.1911 |
| head/world-rust-composed-http/go/native/1 | 5.3019 | 0.0575 | 0.0004 | 215.5024 | 0.0339 | 0.2075 | 0.0004 | N/A | 0.0121 | 17.2606 | 16.0985 | 0.0372 | 387.1216 |
| head/world-rust-composed-http/go/native/4 | 5.3246 | 0.0588 | 0.0003 | 215.9666 | 0.0513 | 0.6438 | 0.0007 | N/A | 0.0136 | 4.9171 | 4.7592 | 0.0637 | 273.6931 |
| head/world-rust-composed-http/llgo/interpreter/1 | 8.8062 | 0.0690 | 0.0003 | 42.4990 | 0.1033 | 0.4507 | 0.0005 | N/A | 0.2845 | 31.4253 | 27.9556 | 0.2444 | 338.2282 |
| head/world-rust-composed-http/llgo/interpreter/4 | 8.6716 | 0.0657 | 0.0003 | 42.6122 | 0.2271 | 1.5361 | 0.0011 | N/A | 0.3099 | 18.7656 | 16.1184 | 0.5917 | 219.4911 |
| head/world-rust-composed-http/llgo/native/1 | 8.8237 | 0.0674 | 0.0003 | 323.4460 | 0.1000 | 0.2844 | 0.0005 | N/A | 0.2823 | 24.2095 | 22.4005 | 0.2312 | 560.6015 |
| head/world-rust-composed-http/llgo/native/4 | 8.9343 | 0.0685 | 0.0003 | 325.3690 | 0.2309 | 0.9842 | 0.0011 | N/A | 0.3081 | 18.1539 | 16.6772 | 0.5816 | 505.8959 |
| head/world-rust-http/go/interpreter/1 | 5.2778 | 0.0546 | 0.0003 | 17.6526 | 0.0384 | 0.1610 | 0.0004 | N/A | 0.0129 | 23.0126 | 20.2590 | 0.0291 | 231.8117 |
| head/world-rust-http/go/interpreter/4 | 5.2982 | 0.0515 | 0.0003 | 17.7756 | 0.0551 | 0.5008 | 0.0008 | N/A | 0.0147 | 7.2589 | 5.7247 | 0.0420 | 85.0358 |
| head/world-rust-http/go/native/1 | 5.4120 | 0.0545 | 0.0004 | 148.7318 | 0.0301 | 0.1581 | 0.0004 | N/A | 0.0127 | 15.5748 | 14.3448 | 0.0366 | 304.1313 |
| head/world-rust-http/go/native/4 | 5.3720 | 0.0535 | 0.0004 | 148.5461 | 0.0464 | 0.4778 | 0.0007 | N/A | 0.0134 | 5.5870 | 4.3580 | 0.0563 | 203.6336 |
| head/world-rust-http/llgo/interpreter/1 | 8.6293 | 0.0617 | 0.0003 | 30.3173 | 0.1030 | 2.7553 | 0.0005 | N/A | 0.2874 | 28.3145 | 24.7696 | 0.2361 | 297.3485 |
| head/world-rust-http/llgo/interpreter/4 | 8.7559 | 0.0602 | 0.0002 | 30.5143 | 0.2534 | 3.6650 | 0.0013 | N/A | 0.3105 | 15.3455 | 15.6474 | 0.5848 | 203.5631 |
| head/world-rust-http/llgo/native/1 | 8.7880 | 0.0597 | 0.0003 | 225.2040 | 0.1039 | 0.2086 | 0.0005 | N/A | 0.2749 | 22.3857 | 20.4681 | 0.2283 | 445.6016 |
| head/world-rust-http/llgo/native/4 | 8.6736 | 0.0625 | 0.0003 | 223.7788 | 0.2329 | 0.7506 | 0.0011 | N/A | 0.3116 | 17.1955 | 16.0858 | 0.5804 | 398.3511 |
| host/native-go/go/host/1 | 4.8830 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0134 | 1.8141 | 1.5502 | 0.0048 | 22.3131 |
| host/native-go/go/host/4 | 4.8854 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0156 | 1.3087 | 1.4042 | 0.0058 | 20.3320 |
| host/native-go/llgo/host/1 | 8.6299 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2758 | 9.1630 | 7.7125 | 0.0772 | 89.5939 |
| host/native-go/llgo/host/4 | 8.5469 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3131 | 10.9353 | 6.9749 | 0.1012 | 86.1038 |
| reference/world-rust-composed-http/go/compiler/1 | 5.5898 | 0.0599 | 0.1332 | 71.1069 | 0.1099 | 0.1278 | 0.0066 | 0.0283 | 0.0350 | 4.6594 | 4.2432 | 0.2199 | 121.7400 |
| reference/world-rust-composed-http/go/compiler/4 | 5.6381 | 0.0640 | 0.1321 | 71.5738 | 0.2179 | 0.2435 | 0.0081 | 0.0395 | 0.0507 | 1.5428 | 1.2329 | 0.3389 | 92.3178 |
| reference/world-rust-http/go/compiler/1 | 5.6692 | 0.0525 | 0.1285 | 49.9573 | 0.1105 | 0.1175 | 0.0065 | 0.0270 | 0.0370 | 4.0962 | 3.7150 | 0.1568 | 96.1627 |
| reference/world-rust-http/go/compiler/4 | 5.7491 | 0.0537 | 0.1292 | 50.0675 | 0.2202 | 0.1812 | 0.0078 | 0.0397 | 0.0472 | 1.3883 | 1.0823 | 0.2142 | 69.5205 |

## Protocol, qualification and remaining work

Source base `6c871d4b698de2b7871399eba3b3d691acabaf6a` adds the common graph tests to production-equivalent `5e4286e`; source head `a893ce5ea8fa2f29baabeb144cf8cbb5da5b1ac8` changes only structural-comparison storage. Go 1.27.3, LLGo 3d, LLVM 22.1.8, M4 Max/macOS 27.0.1. Both Rust plain/composed HTTP fixtures, interpreter/native, one/four workers, 16 clients and four requests per client, 64 KiB, two warmups/seven samples. Exactly 20 randomized primary/A/A blocks, seed 2026111021, 8000 bootstrap draws: 1600 throughput processes, 40 separate resource processes, 80 qualification processes, seven actual loader probes and 716800 checked warm responses. No profile/ISA/count sweep or background isolation. Intervals are pointwise, not family-wise.

Source base/head harnesses are identical. Frozen Wasmtime and ordinary-host controls retain their original compiler/program/SDK hashes. Their shared warm driver, response checker, HTTP/native kernels and reference-specific code match the new harness; the new harness also has owned-buffered-mode flags/report metadata and cold options validation. Those additions do not run inside the reference warm kernels. All timing and phase contracts remain explicit.

Before sampling, all 32 Go and 32 LLGo root packages, the LLGo P3 measurement module, three affected Go race packages and vet passed. All 31 executable owned packages meet 90% coverage (minimum 90.87%). All owned build/test/probe sessions had actually exited, and the actual pre-start compiler/test process table was clear. Full final-head Go/LLGo Linux-amd64/macOS-arm64 CI remains the merge requirement.

Setup diagnostics are retained separately: an initial package test lacked the shared corpus symlink, a mistaken nested-module cwd started a duplicate root run after the complete 32-package run had passed, and an explicit PATH omitted sysctl. The duplicate was stopped and the correct nested module passed. The PATH error occurred before random order generation or any timed runtime process; restoring /usr/sbin allowed the frozen event to start. No timing observation was discarded.

Further work remains on ownership-safe codec/call setup, host scheduling/allocation and native compiler/GC/memory/SIMD gaps. Codec construction already reuses layout builders; it does not clone resource maps. Future reuse must preserve per-caller transfer/resource ownership and cancellation. LLGo candidate adoption remains separate, gated by the GC regression and supported-platform compatibility.
