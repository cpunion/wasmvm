# Component structural comparison: 32 inline pairs, 2026-11-10

[Complete machine record](data/performance-component-type-comparison32-2026-11-10.json).

## Change and observed result

Each structural comparison now keeps its first 32 distinct node pairs on the stack and spills to a map only when needed. This removes the eager per-comparison map from the common path. There is no cross-call or Store-global cache: resource identity predicates, cycle detection, exact node charging and mismatch ordering keep their original semantics. Large graphs retain a map fallback with its initial capacity based on the actual 33 entries.

The pinned compiler remains LLGo `3d92a58`, with `-O3 -lto=full`. This measures project source changes, separately from the [full LLGo #3012 candidate study](performance-llgo-merged-2026-11-10.md); its GC finding and compiler adoption gate remain open.

HTTP fixture, 64 KiB, four workers, bounded buffered response API:

| Path | Base req/s | Head req/s | Base B/request | Head B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 10651 | 11668 | 359680 | 318207 |
| wasmvm go / native | 13133 | 15150 | 359825 | 318351 |
| wasmvm llgo / interpreter | 3427 | 4663 | 560493 | 478356 |
| wasmvm llgo / native | 3394 | 4717 | 561015 | 478805 |
| Wasmtime 48.0.1 | N/A | 58196 | N/A | 32 |
| Ordinary native-Go application / Go | N/A | 43819 | N/A | 277917 |
| Ordinary native-Go application / LLGo | N/A | 8914 | N/A | 732591 |

**8/40 timing controls resolve** within ±2.5%; 0/16 source pairs resolve both controls. These are advisory workstation observations. They do not close controlled performance or release acceptance, establish a calibrated competitor ranking, or partition the host-runtime gap. All unchanged/adverse observations and reference intervals remain in the machine record.

The ordinary native-Go application is an application control, not an optimized native lower bound. Go managed bytes and LLGo collector accounting differ; LLGo Mallocs are unavailable. Rust/JIT allocation is excluded from Go counters. Resource instrumentation runs in a separate cohort.

## Shared graph microbenchmark

Five fixed randomized primary/A/A blocks, 50,000 comparisons per process, GOMAXPROCS=1. These are repeated-comparison kernel timings, separately from process startup. The test setup resolves schemas before timing. Every case and its pointwise 95% paired log-ratio bootstrap interval is retained.

| Compiler | Nodes | Base ns/call | Head ns/call | Base B/call | Head B/call | Head/base time [95%] | Both A/A |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| go | 2 | 40.41 | 22.97 | 0 | 0 | 0.5750 [0.5652,0.5931] | no |
| go | 16 | 908.9 | 152.2 | 1384 | 0 | 0.1672 [0.1627,0.1722] | no |
| go | 17 | 932 | 160 | 1384 | 0 | 0.1706 [0.1680,0.1740] | no |
| go | 64 | 4389 | 3371 | 6696 | 5368 | 0.7761 [0.7628,0.7915] | yes |
| llgo | 2 | 71.16 | 24.45 | 127 | 0 | 0.3450 [0.3391,0.3517] | no |
| llgo | 16 | 1141 | 160.5 | 2172 | 0 | 0.1389 [0.1342,0.1425] | no |
| llgo | 17 | 1178 | 169.4 | 2239 | 0 | 0.1422 [0.1390,0.1447] | no |
| llgo | 64 | 5261 | 4242 | 9465 | 7252 | 0.8069 [0.8038,0.8102] | yes |

Reproduce a kernel sample on either committed source (shared tests begin at `6c871d4`):

```sh
go test ./internal/component/validate -run '^$' -bench '^BenchmarkEqualityResolvedGraphs$' -benchtime=50000x -count=1 -benchmem
llgo test -O3 -lto=full ./internal/component/validate -run '^$' -bench '^BenchmarkEqualityResolvedGraphs$' -benchtime=50000x -count=1 -benchmem
```

An initial 32-entry spill reservation reduced LLGo bytes but increased Go bytes at 17 nodes (1384 -> 1880 B/call). That complete exploratory micro run is retained separately. The final 32-pair implementation spills only on 33 unique pairs. The actual benchmark/helper functions are identical between source base/head; only extra 32/33 semantic boundary tests differ. Final binaries were freshly built from the committed source. Only the final implementation enters the macro event; no timed macro prefix, favorable repeat or pooling is used.

## Source changes in all service profiles

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A | Base B/request | Head B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| head/world-rust-composed-http/go/interpreter/1 | 2663 | 2887 | 0.9354 [0.9115,0.9710] | no | 375558 | 332716 |
| head/world-rust-composed-http/go/interpreter/4 | 9766 | 10724 | 0.9105 [0.8972,0.9239] | no | 375558 | 332717 |
| head/world-rust-composed-http/go/native/1 | 3682 | 4103 | 0.8934 [0.8819,0.9045] | no | 371429 | 328587 |
| head/world-rust-composed-http/go/native/4 | 12183 | 13515 | 0.8690 [0.7983,0.9178] | no | 371431 | 328591 |
| head/world-rust-composed-http/llgo/interpreter/1 | 2187 | 2478 | 0.8835 [0.8559,0.9132] | no | 582391 | 497652 |
| head/world-rust-composed-http/llgo/interpreter/4 | 3413 | 4463 | 0.7392 [0.7077,0.7668] | no | 582301 | 497543 |
| head/world-rust-composed-http/llgo/native/1 | 2665 | 3130 | 0.8507 [0.8376,0.8624] | no | 578122 | 493401 |
| head/world-rust-composed-http/llgo/native/4 | 3245 | 4476 | 0.7250 [0.7050,0.7450] | no | 577992 | 493257 |
| head/world-rust-http/go/interpreter/1 | 2947 | 3153 | 0.9066 [0.8565,0.9438] | no | 359681 | 318208 |
| head/world-rust-http/go/interpreter/4 | 10651 | 11668 | 0.9104 [0.8938,0.9340] | no | 359680 | 318207 |
| head/world-rust-http/go/native/1 | 4147 | 4595 | 0.9087 [0.8956,0.9228] | no | 359824 | 318351 |
| head/world-rust-http/go/native/4 | 13133 | 15150 | 0.8916 [0.8573,0.9363] | no | 359825 | 318351 |
| head/world-rust-http/llgo/interpreter/1 | 2419 | 2706 | 0.8864 [0.8576,0.9100] | no | 560694 | 478462 |
| head/world-rust-http/llgo/interpreter/4 | 3427 | 4663 | 0.7471 [0.7233,0.7772] | no | 560493 | 478356 |
| head/world-rust-http/llgo/native/1 | 2921 | 3453 | 0.8512 [0.8151,0.8944] | no | 561150 | 478920 |
| head/world-rust-http/llgo/native/4 | 3394 | 4717 | 0.7328 [0.7020,0.7691] | no | 561015 | 478805 |

## Throughput, latency and resources

| Profile | req/s | Complete API us | p99 request us | Runtime B/request | Between-batch threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/1 | 2663 | 372.191 | 7039.000 | 375558 | 7 | 98.703 | yes |
| base/world-rust-composed-http/go/interpreter/4 | 9766 | 395.441 | 3509.166 | 375558 | 7 | 123.672 | no |
| base/world-rust-composed-http/go/native/1 | 3682 | 267.962 | 5211.895 | 371429 | 7 | 132.531 | no |
| base/world-rust-composed-http/go/native/4 | 12183 | 314.585 | 3166.833 | 371431 | 7 | 161.125 | no |
| base/world-rust-composed-http/llgo/interpreter/1 | 2187 | 450.954 | 11047.812 | 582391 | 42 | 149.203 | no |
| base/world-rust-composed-http/llgo/interpreter/4 | 3413 | 1152.292 | 8133.396 | 582301 | 57 | 162.203 | no |
| base/world-rust-composed-http/llgo/native/1 | 2665 | 367.434 | 9441.834 | 578122 | 42 | 184.797 | no |
| base/world-rust-composed-http/llgo/native/4 | 3245 | 1215.145 | 8365.229 | 577992 | 57 | 213.609 | no |
| base/world-rust-http/go/interpreter/1 | 2947 | 336.129 | 6272.083 | 359681 | 7 | 80.078 | no |
| base/world-rust-http/go/interpreter/4 | 10651 | 362.196 | 2938.667 | 359680 | 8 | 96.125 | yes |
| base/world-rust-http/go/native/1 | 4147 | 237.711 | 4542.396 | 359824 | 7 | 112.375 | no |
| base/world-rust-http/go/native/4 | 13133 | 291.815 | 2573.749 | 359825 | 7 | 122.594 | no |
| base/world-rust-http/llgo/interpreter/1 | 2419 | 405.007 | 9731.854 | 560694 | 42 | 130.750 | no |
| base/world-rust-http/llgo/interpreter/4 | 3427 | 1148.055 | 7612.917 | 560493 | 57 | 140.891 | yes |
| base/world-rust-http/llgo/native/1 | 2921 | 334.184 | 8496.375 | 561150 | 42 | 152.062 | no |
| base/world-rust-http/llgo/native/4 | 3394 | 1160.680 | 7657.812 | 561015 | 57 | 175.828 | no |
| head/world-rust-composed-http/go/interpreter/1 | 2887 | 343.215 | 6494.355 | 332716 | 7 | 97.125 | no |
| head/world-rust-composed-http/go/interpreter/4 | 10724 | 358.329 | 2964.062 | 332717 | 7 | 118.406 | no |
| head/world-rust-composed-http/go/native/1 | 4103 | 240.171 | 4799.208 | 328587 | 7 | 149.250 | yes |
| head/world-rust-composed-http/go/native/4 | 13515 | 283.440 | 2944.209 | 328591 | 7 | 159.656 | no |
| head/world-rust-composed-http/llgo/interpreter/1 | 2478 | 396.994 | 9643.229 | 497652 | 42 | 152.391 | no |
| head/world-rust-composed-http/llgo/interpreter/4 | 4463 | 877.635 | 6974.708 | 497543 | 57 | 167.938 | no |
| head/world-rust-composed-http/llgo/native/1 | 3130 | 313.190 | 8225.229 | 493401 | 42 | 184.750 | yes |
| head/world-rust-composed-http/llgo/native/4 | 4476 | 873.268 | 6755.583 | 493257 | 57 | 197.438 | no |
| head/world-rust-http/go/interpreter/1 | 3153 | 313.745 | 6016.812 | 318208 | 7 | 76.062 | no |
| head/world-rust-http/go/interpreter/4 | 11668 | 330.571 | 2813.251 | 318207 | 6 | 91.844 | no |
| head/world-rust-http/go/native/1 | 4595 | 214.011 | 4174.249 | 318351 | 7 | 118.922 | yes |
| head/world-rust-http/go/native/4 | 15150 | 253.997 | 2531.688 | 318351 | 7 | 128.562 | no |
| head/world-rust-http/llgo/interpreter/1 | 2706 | 362.372 | 8690.771 | 478462 | 42 | 140.453 | yes |
| head/world-rust-http/llgo/interpreter/4 | 4663 | 828.930 | 6277.874 | 478356 | 57 | 140.641 | no |
| head/world-rust-http/llgo/native/1 | 3453 | 283.072 | 7711.000 | 478920 | 42 | 151.078 | no |
| head/world-rust-http/llgo/native/4 | 4717 | 828.127 | 6117.604 | 478805 | 57 | 176.156 | no |
| host/native-go/go/host/1 | 40357 | 21.018 | 535.958 | 277915 | 6 | 16.766 | no |
| host/native-go/go/host/4 | 43819 | 78.784 | 608.583 | 277917 | 6 | 16.047 | no |
| host/native-go/llgo/host/1 | 8344 | 113.241 | 3655.875 | 732576 | 37 | 68.078 | no |
| host/native-go/llgo/host/4 | 8914 | 406.110 | 3528.666 | 732591 | 40 | 67.938 | no |
| reference/world-rust-composed-http/go/compiler/1 | 14741 | 65.404 | 1228.396 | 32 | 7 | 32.141 | no |
| reference/world-rust-composed-http/go/compiler/4 | 50596 | 74.169 | 382.291 | 32 | 10 | 33.656 | no |
| reference/world-rust-http/go/compiler/1 | 17272 | 55.302 | 1085.812 | 32 | 7 | 29.516 | yes |
| reference/world-rust-http/go/compiler/4 | 58196 | 64.113 | 347.209 | 32 | 7 | 30.625 | no |

## Separate lifecycle phases

Median milliseconds across 20 primary process reductions. Process wall includes protocol parking; it is not standalone CLI latency. Overlapping phases and different cohorts are not subtracted to infer costs. Compilation here includes preparing executable guest code for the chosen backend; source-host builds occur before the event.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1 | 5.4187 | 0.0651 | 0.0003 | 26.6696 | 0.0367 | 0.2295 | 0.0004 | N/A | 0.0120 | 27.0163 | 24.0318 | 0.0327 | 277.7599 |
| base/world-rust-composed-http/go/interpreter/4 | 5.5109 | 0.0649 | 0.0003 | 26.4684 | 0.0551 | 0.6779 | 0.0006 | N/A | 0.0159 | 9.0991 | 6.5536 | 0.0506 | 105.1600 |
| base/world-rust-composed-http/go/native/1 | 5.5070 | 0.0669 | 0.0003 | 221.9945 | 0.0356 | 0.2139 | 0.0004 | N/A | 0.0127 | 19.0355 | 17.3809 | 0.0402 | 406.6606 |
| base/world-rust-composed-http/go/native/4 | 5.4581 | 0.0624 | 0.0003 | 220.5184 | 0.0528 | 0.6573 | 0.0007 | N/A | 0.0137 | 5.2972 | 5.2533 | 0.0760 | 284.3823 |
| base/world-rust-composed-http/llgo/interpreter/1 | 9.0450 | 0.0768 | 0.0003 | 43.1510 | 0.1084 | 0.4520 | 0.0005 | N/A | 0.2788 | 33.7061 | 29.2592 | 0.2609 | 354.0538 |
| base/world-rust-composed-http/llgo/interpreter/4 | 8.9860 | 0.0719 | 0.0003 | 42.0750 | 0.2385 | 1.5892 | 0.0011 | N/A | 0.3211 | 21.5147 | 18.7542 | 0.5949 | 251.6546 |
| base/world-rust-composed-http/llgo/native/1 | 8.8521 | 0.0750 | 0.0003 | 328.3401 | 0.1061 | 0.2880 | 0.0005 | N/A | 0.2835 | 26.1016 | 24.0176 | 0.2533 | 582.2542 |
| base/world-rust-composed-http/llgo/native/4 | 9.0073 | 0.0662 | 0.0003 | 326.6301 | 0.2366 | 0.9792 | 0.0010 | N/A | 0.3256 | 21.1386 | 19.7199 | 0.6051 | 537.3451 |
| base/world-rust-http/go/interpreter/1 | 5.5888 | 0.0601 | 0.0003 | 18.4616 | 0.0404 | 0.1759 | 0.0004 | N/A | 0.0154 | 24.7785 | 21.7171 | 0.0323 | 246.6238 |
| base/world-rust-http/go/interpreter/4 | 5.5654 | 0.0559 | 0.0003 | 18.2404 | 0.0606 | 0.5418 | 0.0008 | N/A | 0.0173 | 7.7750 | 6.0088 | 0.0467 | 89.3953 |
| base/world-rust-http/go/native/1 | 5.4492 | 0.0611 | 0.0003 | 152.1515 | 0.0321 | 0.1693 | 0.0005 | N/A | 0.0129 | 16.8730 | 15.4344 | 0.0367 | 316.6576 |
| base/world-rust-http/go/native/4 | 5.5806 | 0.0566 | 0.0003 | 151.9900 | 0.0500 | 0.4899 | 0.0007 | N/A | 0.0144 | 5.7614 | 4.8730 | 0.0585 | 209.7633 |
| base/world-rust-http/llgo/interpreter/1 | 9.1210 | 0.0690 | 0.0003 | 30.0425 | 0.1111 | 2.8344 | 0.0005 | N/A | 0.3004 | 29.9645 | 26.4551 | 0.2435 | 313.5398 |
| base/world-rust-http/llgo/interpreter/4 | 8.9249 | 0.0670 | 0.0003 | 29.5714 | 0.2623 | 3.7207 | 0.0011 | N/A | 0.3132 | 19.9775 | 18.6785 | 0.5817 | 232.0730 |
| base/world-rust-http/llgo/native/1 | 9.0177 | 0.0649 | 0.0003 | 226.6297 | 0.1004 | 0.2162 | 0.0005 | N/A | 0.2690 | 23.6356 | 21.9090 | 0.2402 | 461.3692 |
| base/world-rust-http/llgo/native/4 | 9.0210 | 0.0730 | 0.0003 | 226.7537 | 0.2336 | 0.7301 | 0.0013 | N/A | 0.3079 | 19.4484 | 18.8590 | 0.5861 | 433.0308 |
| head/world-rust-composed-http/go/interpreter/1 | 5.7118 | 0.0658 | 0.0004 | 26.3254 | 0.0378 | 0.2282 | 0.0004 | N/A | 0.0127 | 25.5677 | 22.1655 | 0.0331 | 260.6815 |
| head/world-rust-composed-http/go/interpreter/4 | 5.3385 | 0.0612 | 0.0004 | 26.3512 | 0.0545 | 0.6870 | 0.0006 | N/A | 0.0162 | 8.6381 | 5.9677 | 0.0537 | 99.2136 |
| head/world-rust-composed-http/go/native/1 | 5.5364 | 0.0630 | 0.0004 | 220.0791 | 0.0371 | 0.2135 | 0.0003 | N/A | 0.0128 | 16.7089 | 15.5982 | 0.0393 | 385.3727 |
| head/world-rust-composed-http/go/native/4 | 5.4402 | 0.0637 | 0.0004 | 219.8063 | 0.0526 | 0.6467 | 0.0005 | N/A | 0.0134 | 5.0224 | 4.7355 | 0.0669 | 277.5165 |
| head/world-rust-composed-http/llgo/interpreter/1 | 9.0134 | 0.0709 | 0.0003 | 42.2909 | 0.1045 | 0.4386 | 0.0005 | N/A | 0.2820 | 30.0816 | 25.8286 | 0.2618 | 318.3405 |
| head/world-rust-composed-http/llgo/interpreter/4 | 8.8183 | 0.0747 | 0.0003 | 42.4125 | 0.2404 | 1.5552 | 0.0010 | N/A | 0.3194 | 17.1735 | 14.3399 | 0.5732 | 202.6977 |
| head/world-rust-composed-http/llgo/native/1 | 9.0981 | 0.0808 | 0.0003 | 326.9474 | 0.0958 | 0.2874 | 0.0005 | N/A | 0.2772 | 21.2960 | 20.4488 | 0.2494 | 546.4545 |
| head/world-rust-composed-http/llgo/native/4 | 8.9157 | 0.0772 | 0.0003 | 327.5529 | 0.2392 | 0.9640 | 0.0010 | N/A | 0.3175 | 15.8971 | 14.3087 | 0.5985 | 484.2433 |
| head/world-rust-http/go/interpreter/1 | 5.5144 | 0.0594 | 0.0004 | 18.3526 | 0.0360 | 0.1617 | 0.0004 | N/A | 0.0147 | 22.8147 | 20.2992 | 0.0304 | 231.7061 |
| head/world-rust-http/go/interpreter/4 | 5.5291 | 0.0602 | 0.0004 | 18.1812 | 0.0641 | 0.7843 | 0.0008 | N/A | 0.0162 | 7.4096 | 5.4855 | 0.0474 | 84.5687 |
| head/world-rust-http/go/native/1 | 5.6297 | 0.0595 | 0.0004 | 151.3060 | 0.0321 | 0.1668 | 0.0004 | N/A | 0.0123 | 15.4584 | 13.9273 | 0.0366 | 300.4775 |
| head/world-rust-http/go/native/4 | 5.5186 | 0.0587 | 0.0004 | 151.5965 | 0.0528 | 0.4852 | 0.0006 | N/A | 0.0141 | 4.8701 | 4.2246 | 0.0603 | 204.6065 |
| head/world-rust-http/llgo/interpreter/1 | 9.1305 | 0.0690 | 0.0003 | 29.8495 | 0.1063 | 2.8049 | 0.0005 | N/A | 0.2856 | 24.5395 | 23.6476 | 0.2482 | 284.7121 |
| head/world-rust-http/llgo/interpreter/4 | 8.9674 | 0.0682 | 0.0003 | 29.0728 | 0.2531 | 3.3259 | 0.0012 | N/A | 0.3078 | 13.4999 | 13.7277 | 0.5822 | 184.3545 |
| head/world-rust-http/llgo/native/1 | 8.9869 | 0.0671 | 0.0003 | 227.3453 | 0.1033 | 0.2153 | 0.0005 | N/A | 0.2841 | 20.7427 | 18.5365 | 0.2442 | 430.4094 |
| head/world-rust-http/llgo/native/4 | 8.9312 | 0.0667 | 0.0003 | 227.0208 | 0.2333 | 0.7236 | 0.0010 | N/A | 0.3091 | 14.6294 | 13.5687 | 0.5976 | 379.2123 |
| host/native-go/go/host/1 | 5.0053 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0146 | 1.8114 | 1.5859 | 0.0056 | 22.6213 |
| host/native-go/go/host/4 | 4.9569 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0154 | 1.3345 | 1.4606 | 0.0069 | 20.8763 |
| host/native-go/llgo/host/1 | 8.8225 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2989 | 9.6674 | 7.6701 | 0.0827 | 89.8512 |
| host/native-go/llgo/host/4 | 8.7485 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3378 | 10.5424 | 7.1800 | 0.0863 | 89.2444 |
| reference/world-rust-composed-http/go/compiler/1 | 5.9374 | 0.0659 | 0.1374 | 73.9381 | 0.1139 | 0.1341 | 0.0066 | 0.0281 | 0.0353 | 4.8093 | 4.3416 | 0.2368 | 125.5501 |
| reference/world-rust-composed-http/go/compiler/4 | 5.8164 | 0.0630 | 0.1325 | 73.2235 | 0.2228 | 0.2522 | 0.0083 | 0.0422 | 0.0501 | 1.5471 | 1.2649 | 0.3665 | 94.6281 |
| reference/world-rust-http/go/compiler/1 | 5.9834 | 0.0588 | 0.1331 | 50.8165 | 0.1160 | 0.1197 | 0.0067 | 0.0283 | 0.0347 | 4.2413 | 3.7054 | 0.1595 | 96.5487 |
| reference/world-rust-http/go/compiler/4 | 5.9283 | 0.0639 | 0.1364 | 51.2945 | 0.2235 | 0.1881 | 0.0081 | 0.0406 | 0.0436 | 1.4174 | 1.0997 | 0.2429 | 70.6561 |

## Protocol, qualification and remaining work

Source base `6c871d4b698de2b7871399eba3b3d691acabaf6a` adds the common graph tests to production-equivalent `5e4286e`; source head `4cb4ddf79921d27285c17e79131b706b9b91ceb3` changes only structural-comparison storage. Go 1.27.3, LLGo 3d, LLVM 22.1.8, M4 Max/macOS 27.0.1. Both Rust plain/composed HTTP fixtures, interpreter/native, one/four workers, 16 clients and four requests per client, 64 KiB, two warmups/seven samples. Exactly 20 randomized primary/A/A blocks, seed 2026111022, 8000 bootstrap draws: 1600 throughput processes, 40 separate resource processes, 80 qualification processes, seven actual loader probes and 716800 checked warm responses. No profile/ISA/count sweep or background isolation. Intervals are pointwise, not family-wise.

Source base/head harnesses are identical. Frozen Wasmtime and ordinary-host controls retain their original compiler/program/SDK hashes. Their shared warm driver, response checker, HTTP/native kernels and reference-specific code match the new harness; the new harness also has owned-buffered-mode flags/report metadata and cold options validation. Those additions do not run inside the reference warm kernels. All timing and phase contracts remain explicit.

Before sampling, all 32 Go root and 3 affected LLGo packages, the LLGo P3 measurement module, three affected Go race packages and vet passed. All 31 executable owned packages meet 90% coverage (minimum 90.87%). All owned build/test/probe sessions had actually exited, and the actual pre-start owned compiler/test process table was clear. Foreign tests/builds, when present, are recorded in preflight and are never interrupted; default workstation timing remains advisory. Full final-head Go/LLGo Linux-amd64/macOS-arm64 CI remains the merge requirement.

The earlier 16-pair source event is retained separately in [its complete report](performance-component-type-comparison-2026-11-10.md), including its setup diagnostics. A subsequent non-timing atomic diagnostic counted 29 comparisons of 2 nodes, 16 of 17 nodes and 15 of 19 nodes per actual HTTP request. That evidence motivates 32 inline pairs (512 bytes of pair slots, 256 more than 16). This is one distinct source event against the same production-equivalent baseline, with new frozen order/seed 2026111022, no reused timed prefix or pooling. No foreign process is interrupted.

Further work remains on ownership-safe codec/call setup, host scheduling/allocation and native compiler/GC/memory/SIMD gaps. Codec construction already reuses layout builders; it does not clone resource maps. Future reuse must preserve per-caller transfer/resource ownership and cancellation. LLGo candidate adoption remains separate, gated by the GC regression and supported-platform compatibility.
