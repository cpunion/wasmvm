# Immutable Canonical byte transfers: 2026-11-10

[Machine record](data/performance-immutable-memory-2026-11-10.json).

M4 Max/macOS27.0.1; base `88c7272de0c83f96b0fee16466feeeae9a9aee4e`, head `10176c0b2270235a8f05a033988cd44259dacd52`. Go1.27.3 and LLGo3d92a58 `-O3 -lto=full`; independent actual Wasmtime48.0.1 P3 providers through the retained Go C/Rust bridge. The change removes private-memory staging for packed bytes, preserving immutable ownership, full range checks, logical transfer charges, shared synchronization and post-copy cancellation.

**Outcome:** about64KiB fewer runtime-accounted bytes per64KiB response in both host compilers. Throughput remains mixed and the large P3 gap remains open. Only13/40 identical-binary throughput controls resolve. This event closes no controlled performance or release gate. Go managed-byte and LLGo BDWGC counters have different scopes; the Go reference counter excludes Rust/JIT allocation.

## Fixed event and evidence

Twenty independent process blocks,16 clients,4 requests/client,1/4 workers,2 warmup and7 measured batches, both Rust service and composed middleware/service fixtures. Primary and identical-binary aliases are randomized together using seed2026111004. All1600 throughput processes,40 separate resource processes,80 qualification processes and7 actual loader probes completed;716800 warm responses validate every64KiB byte, status201, header and trailer. Process means include all pauses. Pointwise paired log-ratio intervals use8000 bootstrap draws and seed2026111004; no family-wise portfolio acceptance is claimed. The comparison includes owned base/head under both backends/compilers, actual Wasmtime and bare Go/LLGo application controls.

Preparation, first/warm batches, full public API, request wall and joined cleanup remain separate. Request construction is outside owned API timing but inside request wall; the reference constructs its request inside the C/Rust bridge. Collection and close are inside API timing. Full response validation follows the API timer and remains in request wall time. No phase subtraction, cross-host pooling or workload-specific production path. Native-Go is an ordinary echo/body-collection application control, not an optimized native lower bound.

The reference/control binaries reuse their exact frozen earlier source/build identities; all four owned binaries are fresh base/head builds. Source archives, compiler/binary hashes and actual loaded-library probes are checked. All measurement-driver and P3 harness sources match across base/head.

## Every owned comparison

Warm ratio is head/base process-mean batch time, with a pointwise95% interval. Allocation is median runtime bytes/request. Controls require each A/A interval inside[0.975,1.025].

| Fixture / compiler / backend / workers | Base req/s | Head req/s | Warm ratio [95%] | Both controls resolve | Base B/req | Head B/req |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| world-rust-composed-http/go/interpreter/1 | 1487 | 1495 | 1.0004 [0.9908,1.0105] | yes | 821685 | 756128 |
| world-rust-composed-http/go/interpreter/4 | 5882 | 5980 | 0.9799 [0.9631,0.9950] | yes | 821647 | 756096 |
| world-rust-composed-http/go/native/1 | 1919 | 1927 | 0.9969 [0.9817,1.0090] | no | 818069 | 752512 |
| world-rust-composed-http/go/native/4 | 6905 | 7122 | 0.9521 [0.9181,0.9785] | no | 818034 | 752476 |
| world-rust-composed-http/llgo/interpreter/1 | 847 | 843 | 1.0348 [1.0066,1.0820] | no | 1359418 | 1293715 |
| world-rust-composed-http/llgo/interpreter/4 | 1473 | 1475 | 0.9959 [0.9715,1.0168] | no | 1359476 | 1293997 |
| world-rust-composed-http/llgo/native/1 | 928 | 935 | 0.9880 [0.9635,1.0078] | yes | 1355847 | 1290503 |
| world-rust-composed-http/llgo/native/4 | 1445 | 1419 | 1.0046 [0.9829,1.0260] | no | 1356065 | 1290442 |
| world-rust-http/go/interpreter/1 | 1582 | 1597 | 0.9869 [0.9650,1.0066] | no | 805799 | 740250 |
| world-rust-http/go/interpreter/4 | 6074 | 6206 | 0.9789 [0.9645,0.9914] | yes | 805777 | 740229 |
| world-rust-http/go/native/1 | 2058 | 2056 | 1.0009 [0.9927,1.0087] | yes | 806454 | 740915 |
| world-rust-http/go/native/4 | 7309 | 7294 | 0.9909 [0.9655,1.0163] | no | 806429 | 740878 |
| world-rust-http/llgo/interpreter/1 | 874 | 881 | 1.0064 [0.9868,1.0312] | no | 1337751 | 1271945 |
| world-rust-http/llgo/interpreter/4 | 1457 | 1456 | 1.0031 [0.9757,1.0334] | no | 1337805 | 1272327 |
| world-rust-http/llgo/native/1 | 964 | 964 | 1.0065 [0.9896,1.0228] | no | 1338789 | 1273196 |
| world-rust-http/llgo/native/4 | 1454 | 1454 | 0.9920 [0.9714,1.0113] | no | 1339095 | 1273491 |

The LLGo composed interpreter/1-worker case has an adverse1.0348 warm-time ratio, interval[1.0066,1.0820], with unresolved controls. Preserve this observation; do not call all timings improved or claim no regression. The allocation reduction is distinct from throughput acceptance.

## All profiles and controls

| Profile | req/s | Complete API us | Request p99 us | Runtime B/req | Idle threads | Peak RSS MiB | A/A95% | Resolved |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| base/world-rust-composed-http/go/interpreter/1 | 1487 | 670.018 | 12008.542 | 821685 | 7 | 104.922 | [0.9837,1.0069] | yes |
| base/world-rust-composed-http/go/interpreter/4 | 5882 | 659.629 | 4801.979 | 821647 | 8 | 147.688 | [0.9804,1.0129] | yes |
| base/world-rust-composed-http/go/native/1 | 1919 | 518.397 | 9141.562 | 818069 | 7 | 146.281 | [0.9960,1.0474] | no |
| base/world-rust-composed-http/go/native/4 | 6905 | 562.131 | 4378.271 | 818034 | 8 | 162.547 | [0.9468,1.0048] | no |
| base/world-rust-composed-http/llgo/interpreter/1 | 847 | 1171.196 | 22552.375 | 1359418 | 40 | 163.922 | [0.9813,1.0208] | yes |
| base/world-rust-composed-http/llgo/interpreter/4 | 1473 | 2669.001 | 14038.584 | 1359476 | 49 | 180.938 | [0.9935,1.0935] | no |
| base/world-rust-composed-http/llgo/native/1 | 928 | 1067.289 | 20392.209 | 1355847 | 40 | 180.312 | [0.9862,1.0088] | yes |
| base/world-rust-composed-http/llgo/native/4 | 1445 | 2729.011 | 14166.896 | 1356065 | 49 | 194.281 | [0.9465,0.9979] | no |
| base/world-rust-http/go/interpreter/1 | 1582 | 629.632 | 10869.959 | 805799 | 7 | 81.500 | [0.9712,1.0086] | no |
| base/world-rust-http/go/interpreter/4 | 6074 | 639.029 | 4229.208 | 805777 | 7 | 105.484 | [0.9843,1.0062] | yes |
| base/world-rust-http/go/native/1 | 2058 | 483.353 | 8400.166 | 806454 | 7 | 118.344 | [0.9909,1.0174] | yes |
| base/world-rust-http/go/native/4 | 7309 | 531.122 | 3636.001 | 806429 | 8 | 126.453 | [0.9914,1.0351] | no |
| base/world-rust-http/llgo/interpreter/1 | 874 | 1134.477 | 21465.729 | 1337751 | 40 | 140.016 | [0.9878,1.0509] | no |
| base/world-rust-http/llgo/interpreter/4 | 1457 | 2700.197 | 13722.854 | 1337805 | 49 | 140.438 | [0.9950,1.0471] | no |
| base/world-rust-http/llgo/native/1 | 964 | 1028.072 | 19195.605 | 1338789 | 40 | 151.859 | [0.9951,1.0333] | no |
| base/world-rust-http/llgo/native/4 | 1454 | 2709.303 | 13708.022 | 1339095 | 49 | 158.500 | [0.9741,1.0563] | no |
| head/world-rust-composed-http/go/interpreter/1 | 1495 | 666.424 | 11927.166 | 756128 | 7 | 105.000 | [0.9819,1.0173] | yes |
| head/world-rust-composed-http/go/interpreter/4 | 5980 | 646.494 | 4682.875 | 756096 | 7 | 145.922 | [0.9973,1.0199] | yes |
| head/world-rust-composed-http/go/native/1 | 1927 | 516.271 | 9297.729 | 752512 | 7 | 143.938 | [0.9927,1.0300] | no |
| head/world-rust-composed-http/go/native/4 | 7122 | 543.247 | 4093.896 | 752476 | 7 | 159.031 | [0.9837,1.0109] | yes |
| head/world-rust-composed-http/llgo/interpreter/1 | 843 | 1177.868 | 23054.688 | 1293715 | 40 | 161.547 | [0.9253,1.0073] | no |
| head/world-rust-composed-http/llgo/interpreter/4 | 1475 | 2681.967 | 14493.271 | 1293997 | 49 | 181.062 | [0.9827,1.0641] | no |
| head/world-rust-composed-http/llgo/native/1 | 935 | 1062.499 | 20228.562 | 1290503 | 40 | 180.547 | [0.9755,1.0197] | yes |
| head/world-rust-composed-http/llgo/native/4 | 1419 | 2781.677 | 14450.188 | 1290442 | 49 | 209.516 | [0.9372,1.0010] | no |
| head/world-rust-http/go/interpreter/1 | 1597 | 623.730 | 10678.812 | 740250 | 8 | 80.906 | [0.9844,1.0048] | yes |
| head/world-rust-http/go/interpreter/4 | 6206 | 626.098 | 4103.145 | 740229 | 7 | 104.828 | [0.9881,1.0120] | yes |
| head/world-rust-http/go/native/1 | 2056 | 483.818 | 8682.438 | 740915 | 7 | 118.594 | [0.9781,1.0098] | yes |
| head/world-rust-http/go/native/4 | 7294 | 532.499 | 3739.062 | 740878 | 7 | 121.922 | [0.9797,1.0375] | no |
| head/world-rust-http/llgo/interpreter/1 | 881 | 1127.190 | 21200.645 | 1271945 | 40 | 140.078 | [0.9703,1.0099] | no |
| head/world-rust-http/llgo/interpreter/4 | 1456 | 2711.373 | 13751.291 | 1272327 | 49 | 140.672 | [0.9691,1.0159] | no |
| head/world-rust-http/llgo/native/1 | 964 | 1030.108 | 19476.438 | 1273196 | 40 | 151.953 | [0.9707,1.0069] | no |
| head/world-rust-http/llgo/native/4 | 1454 | 2712.669 | 13744.438 | 1273491 | 49 | 161.688 | [0.9760,1.0276] | no |
| host/native-go/go/host/1 | 40590 | 20.929 | 534.313 | 277914 | 7 | 16.984 | [0.9050,1.0025] | no |
| host/native-go/go/host/4 | 46295 | 74.054 | 530.375 | 277918 | 6 | 18.125 | [0.9687,1.0680] | no |
| host/native-go/llgo/host/1 | 8143 | 116.591 | 3530.479 | 732588 | 37 | 68.094 | [0.9530,1.0262] | no |
| host/native-go/llgo/host/4 | 8995 | 399.497 | 3750.750 | 732584 | 40 | 68.016 | [0.9607,1.0740] | no |
| reference/world-rust-composed-http/go/compiler/1 | 14972 | 64.278 | 1174.583 | 32 | 7 | 32.266 | [0.9828,1.0404] | no |
| reference/world-rust-composed-http/go/compiler/4 | 51642 | 72.881 | 372.625 | 32 | 9 | 33.844 | [0.9909,1.0331] | no |
| reference/world-rust-http/go/compiler/1 | 17305 | 55.369 | 1048.646 | 32 | 7 | 29.922 | [0.9929,1.0305] | no |
| reference/world-rust-http/go/compiler/4 | 59162 | 62.939 | 336.417 | 32 | 7 | 30.500 | [1.0028,1.0271] | no |

## Complete lifecycle intervals

Median milliseconds; warm is the median of20 process-mean batches. Parallel/setup phases are not added or subtracted to infer an unmeasured interval. The machine record retains all per-process reductions and32 paired head/reference intervals.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1 | 5.4934 | 0.0619 | 0.0003 | 26.2020 | 0.0369 | 0.2195 | 0.0003 | N/A | 0.0122 | 47.0792 | 43.0469 | 0.0295 | 469.0501 |
| base/world-rust-composed-http/go/interpreter/4 | 5.5274 | 0.0609 | 0.0003 | 26.1167 | 0.0558 | 0.6998 | 0.0007 | N/A | 0.0158 | 13.1039 | 10.8801 | 0.0451 | 147.9657 |
| base/world-rust-composed-http/go/native/1 | 5.4966 | 0.0607 | 0.0003 | 215.5058 | 0.0365 | 0.2135 | 0.0004 | N/A | 0.0123 | 33.9566 | 33.3441 | 0.0363 | 557.9271 |
| base/world-rust-composed-http/go/native/4 | 5.4969 | 0.0611 | 0.0003 | 217.7395 | 0.0532 | 0.6522 | 0.0006 | N/A | 0.0133 | 9.9356 | 9.2680 | 0.0630 | 320.8303 |
| base/world-rust-composed-http/llgo/interpreter/1 | 8.8240 | 0.0730 | 0.0002 | 44.5364 | 0.1164 | 0.4367 | 0.0005 | N/A | 0.2781 | 77.6930 | 75.5947 | 0.1933 | 816.0928 |
| base/world-rust-composed-http/llgo/interpreter/4 | 8.8711 | 0.0711 | 0.0003 | 44.2259 | 0.2347 | 1.5511 | 0.0010 | N/A | 0.3064 | 49.5731 | 43.4493 | 0.4438 | 496.0237 |
| base/world-rust-composed-http/llgo/native/1 | 8.8470 | 0.0701 | 0.0002 | 327.3079 | 0.1059 | 0.2855 | 0.0005 | N/A | 0.2765 | 71.0760 | 68.9675 | 0.1970 | 1031.7373 |
| base/world-rust-composed-http/llgo/native/4 | 8.9533 | 0.0690 | 0.0002 | 326.9793 | 0.2526 | 1.0433 | 0.0010 | N/A | 0.3205 | 45.0197 | 44.2775 | 0.4623 | 782.7107 |
| base/world-rust-http/go/interpreter/1 | 5.4586 | 0.0543 | 0.0003 | 18.1711 | 0.0393 | 0.1676 | 0.0004 | N/A | 0.0132 | 44.2084 | 40.4596 | 0.0302 | 435.8208 |
| base/world-rust-http/go/interpreter/4 | 5.4914 | 0.0549 | 0.0003 | 18.2496 | 0.0594 | 0.5253 | 0.0007 | N/A | 0.0158 | 12.4564 | 10.5374 | 0.0414 | 134.2124 |
| base/world-rust-http/go/native/1 | 5.5506 | 0.0522 | 0.0004 | 149.6745 | 0.0333 | 0.1661 | 0.0004 | N/A | 0.0129 | 32.0301 | 31.0974 | 0.0358 | 470.8113 |
| base/world-rust-http/go/native/4 | 5.5042 | 0.0558 | 0.0004 | 148.6698 | 0.0505 | 0.4812 | 0.0006 | N/A | 0.0135 | 10.1350 | 8.7565 | 0.0538 | 246.9262 |
| base/world-rust-http/llgo/interpreter/1 | 9.0089 | 0.0600 | 0.0002 | 29.8008 | 0.0998 | 3.0819 | 0.0005 | N/A | 0.2896 | 74.5701 | 73.2587 | 0.1948 | 777.3758 |
| base/world-rust-http/llgo/interpreter/4 | 8.8298 | 0.0620 | 0.0002 | 29.5902 | 0.2551 | 3.4299 | 0.0011 | N/A | 0.3054 | 45.4384 | 43.9294 | 0.4316 | 485.2341 |
| base/world-rust-http/llgo/native/1 | 8.9231 | 0.0635 | 0.0002 | 228.7146 | 0.1034 | 0.2195 | 0.0005 | N/A | 0.2760 | 67.6465 | 66.3872 | 0.2106 | 908.7462 |
| base/world-rust-http/llgo/native/4 | 8.8906 | 0.0617 | 0.0002 | 226.8995 | 0.2415 | 0.7464 | 0.0013 | N/A | 0.3100 | 44.5320 | 44.0163 | 0.4440 | 685.1734 |
| head/world-rust-composed-http/go/interpreter/1 | 5.4562 | 0.0634 | 0.0003 | 26.1906 | 0.0370 | 0.2320 | 0.0004 | N/A | 0.0124 | 47.3727 | 42.8131 | 0.0302 | 470.2763 |
| head/world-rust-composed-http/go/interpreter/4 | 5.5181 | 0.0614 | 0.0003 | 26.1308 | 0.0554 | 0.6942 | 0.0007 | N/A | 0.0164 | 13.0499 | 10.7032 | 0.0488 | 146.2086 |
| head/world-rust-composed-http/go/native/1 | 5.4291 | 0.0614 | 0.0003 | 217.5823 | 0.0340 | 0.2114 | 0.0005 | N/A | 0.0120 | 33.8612 | 33.2118 | 0.0376 | 557.9966 |
| head/world-rust-composed-http/go/native/4 | 5.5009 | 0.0606 | 0.0003 | 216.9049 | 0.0510 | 0.6478 | 0.0007 | N/A | 0.0133 | 10.0521 | 8.9867 | 0.0635 | 318.0467 |
| head/world-rust-composed-http/llgo/interpreter/1 | 8.9140 | 0.0718 | 0.0002 | 44.8906 | 0.1198 | 0.4360 | 0.0005 | N/A | 0.2746 | 77.7694 | 75.9405 | 0.2094 | 820.2269 |
| head/world-rust-composed-http/llgo/interpreter/4 | 8.9249 | 0.0696 | 0.0002 | 43.8297 | 0.2484 | 1.5401 | 0.0010 | N/A | 0.3057 | 50.2948 | 43.3930 | 0.4467 | 499.4880 |
| head/world-rust-composed-http/llgo/native/1 | 8.8550 | 0.0694 | 0.0002 | 326.1762 | 0.1088 | 0.2889 | 0.0005 | N/A | 0.2843 | 69.3574 | 68.4615 | 0.2109 | 1021.5328 |
| head/world-rust-composed-http/llgo/native/4 | 8.8850 | 0.0728 | 0.0002 | 326.7745 | 0.2463 | 1.0148 | 0.0010 | N/A | 0.3151 | 44.5303 | 45.1106 | 0.4609 | 788.9202 |
| head/world-rust-http/go/interpreter/1 | 5.4727 | 0.0596 | 0.0003 | 18.1751 | 0.0372 | 0.1619 | 0.0004 | N/A | 0.0137 | 44.3068 | 40.0804 | 0.0279 | 432.3042 |
| head/world-rust-http/go/interpreter/4 | 5.4916 | 0.0544 | 0.0003 | 18.0852 | 0.0668 | 0.9835 | 0.0009 | N/A | 0.0193 | 11.9847 | 10.3126 | 0.0389 | 132.5639 |
| head/world-rust-http/go/native/1 | 5.4760 | 0.0550 | 0.0003 | 151.0255 | 0.0340 | 0.1643 | 0.0005 | N/A | 0.0127 | 31.8993 | 31.1291 | 0.0364 | 472.6917 |
| head/world-rust-http/go/native/4 | 5.4601 | 0.0569 | 0.0003 | 150.3793 | 0.0501 | 0.4675 | 0.0008 | N/A | 0.0150 | 10.2871 | 8.7747 | 0.0512 | 247.3916 |
| head/world-rust-http/llgo/interpreter/1 | 8.8509 | 0.0615 | 0.0002 | 30.5599 | 0.1104 | 2.6670 | 0.0005 | N/A | 0.3054 | 75.2539 | 72.6055 | 0.2070 | 773.4633 |
| head/world-rust-http/llgo/interpreter/4 | 8.9685 | 0.0640 | 0.0002 | 30.5951 | 0.2598 | 3.5871 | 0.0010 | N/A | 0.3133 | 46.8998 | 43.9527 | 0.4357 | 488.7101 |
| head/world-rust-http/llgo/native/1 | 8.8356 | 0.0649 | 0.0002 | 227.0549 | 0.1060 | 0.2165 | 0.0005 | N/A | 0.2842 | 67.8520 | 66.4103 | 0.1924 | 907.4103 |
| head/world-rust-http/llgo/native/4 | 8.8591 | 0.0657 | 0.0002 | 225.7046 | 0.2275 | 0.7269 | 0.0010 | N/A | 0.3108 | 44.1754 | 44.0128 | 0.4455 | 677.9162 |
| host/native-go/go/host/1 | 4.9643 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0132 | 1.7548 | 1.5767 | 0.0074 | 22.2754 |
| host/native-go/go/host/4 | 5.0641 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0156 | 1.2899 | 1.3825 | 0.0060 | 20.2367 |
| host/native-go/llgo/host/1 | 8.6372 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2488 | 9.6797 | 7.8595 | 0.0805 | 89.8424 |
| host/native-go/llgo/host/4 | 8.7039 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3133 | 9.9860 | 7.1154 | 0.0879 | 86.2674 |
| reference/world-rust-composed-http/go/compiler/1 | 5.8826 | 0.0599 | 0.1356 | 72.3342 | 0.1153 | 0.1389 | 0.0065 | 0.0311 | 0.0353 | 4.7164 | 4.2746 | 0.2355 | 124.0192 |
| reference/world-rust-composed-http/go/compiler/4 | 5.9319 | 0.0624 | 0.1343 | 72.2453 | 0.2194 | 0.2457 | 0.0080 | 0.0421 | 0.0489 | 1.5373 | 1.2393 | 0.3326 | 93.5374 |
| reference/world-rust-http/go/compiler/1 | 5.8039 | 0.0545 | 0.1435 | 50.3840 | 0.1185 | 0.1239 | 0.0065 | 0.0324 | 0.0337 | 4.0556 | 3.6983 | 0.1649 | 95.4792 |
| reference/world-rust-http/go/compiler/4 | 5.8505 | 0.0564 | 0.1353 | 50.3032 | 0.2237 | 0.1936 | 0.0077 | 0.0414 | 0.0459 | 1.3815 | 1.0818 | 0.2273 | 69.5957 |

## Validation and retained limitations

Complete normal Go owned-package tests pass, with all31 measured packages >=90% coverage (minimum90.72%). Affected memory/Canonical/VM/component race tests, P3 race tests and vet pass. Pinned LLGo release tests actually execute the four affected packages and both P3 backends. An unnecessarily broad local full-corpus race+coverpkg diagnostic timed out at5 minutes/package; its raw stacks remain explicit and it is not passing semantic evidence. Normal full-corpus coverage and focused race checks are separate. Final-head full Go/LLGo CI remains required for the contribution.

An initial campaign was stopped after detecting another local Go compiler job; all165 incomplete rows remain in `artifacts/memory-transfers-20261110` and are never pooled here. This complete cohort reuses identical fixed sources, binaries, counts and profiles. The supplemental process guard initially matched its own shell arguments rather than executable paths; its assertion failed and the shell proceeded. Replaying the retained actual pre-start process table by executable confirms no compiler. That supplemental metadata audit finished after timing began; the original source/binary/qualification preflight had passed beforehand. No binary, sampling rule or timed outcome changed, and this complete cohort was not restarted. Background GUI applications and uncontrolled affinity/frequency remain; do not promote this event to controlled acceptance.

Raw results and independent audit are retained in `artifacts/memory-transfers-isolated-20261110`; result SHA256 `d90ebe4b91448d6278b711dd162f48ecbe0f5164c69b0648677033e6af13b17b`. The machine record retains all40 profiles,1600 process reductions, resource snapshots, controls and adverse ratios. Broader component scheduling/copy costs, native compiler gaps, both-target controlled acceptance and release publication remain open.
