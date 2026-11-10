# Immutable resource inspection diagnostics, 2026-11-10

[Complete machine record](data/performance-reference-inspection-2026-11-10.json).

This is the initial, independently retained source event. [Initial CI](https://github.com/cpunion/wasmvm/actions/runs/42809492665) failed its HTTP busy-Store admission test because internal cancellation could hide an already published failure. The [corrected-source event](performance-reference-inspection-corrected-2026-11-10.md) measures the follow-up fix; these observations are not pooled or replaced.

## Change and evidence

Normal affine cleanup inspects aliases invalidated by successful moves. `ReferenceInfo` previously formatted the same expired-capability error for each inspection. It now reuses one private immutable diagnostic, created during package initialization. The error message, `errors.Is(wasm.ErrTrap)`, mutexes, resource validity and active borrow-scope checks are unchanged. Live references and invalid nil references keep their existing results. No capability, caller or per-call mutable state is cached.

The preceding [runtime-only allocation profile](performance-byte-destination-2026-11-10.md) excludes compilation/setup and identifies this formatting under `publishReference` and `cleanupResource`. It is normal ownership bookkeeping, not evidence of failed HTTP requests or an LLGo defect. Core value encoding, caller and codec allocation remain independent hotspots.

## Inspection allocation diagnostic

The identical checked-in benchmark runs 50000 inspections per case on each source/compiler. Setup is outside the measured loop. These are allocation counters; retained ns/op are diagnostic, without a timing-ranking claim. LLGo Mallocs are unavailable and raw zero counters do not mean zero allocations.

| Source | Compiler | Case | B/inspection | Go mallocs/inspection |
| --- | --- | --- | ---: | ---: |
| base | go | live | 0 | 0 |
| base | go | expired | 128 | 3 |
| base | go | ended-borrow | 128 | 3 |
| base | go | nil | 0 | 0 |
| base | llgo | live | 0 | N/A |
| base | llgo | expired | 160 | N/A |
| base | llgo | ended-borrow | 160 | N/A |
| base | llgo | nil | 0 | N/A |
| head | go | live | 0 | 0 |
| head | go | expired | 0 | 0 |
| head | go | ended-borrow | 0 | 0 |
| head | go | nil | 0 | 0 |
| head | llgo | live | 0 | N/A |
| head | llgo | expired | 0 | N/A |
| head | llgo | ended-borrow | 0 | N/A |
| head | llgo | nil | 0 | N/A |

## HTTP fixture, 16 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 16666 | 16929 | 87557 | 85390 |
| wasmvm go / native | 20898 | 22455 | 87699 | 85522 |
| wasmvm llgo / interpreter | 6192 | 6394 | 158137 | 155424 |
| wasmvm llgo / native | 6221 | 6230 | 158634 | 155949 |
| Wasmtime 48.0.1 | N/A | 64413 | N/A | 44 |
| Ordinary native-Go application / Go | N/A | 1621473 | N/A | 2712 |
| Ordinary native-Go application / LLGo | N/A | 120244 | N/A | 3499 |

## HTTP fixture, 65536 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 11750 | 11558 | 252775 | 250595 |
| wasmvm go / native | 15425 | 16392 | 252921 | 250738 |
| wasmvm llgo / interpreter | 4643 | 4836 | 411894 | 409179 |
| wasmvm llgo / native | 4458 | 4792 | 412175 | 409508 |
| Wasmtime 48.0.1 | N/A | 59092 | N/A | 45 |
| Ordinary native-Go application / Go | N/A | 43943 | N/A | 277915 |
| Ordinary native-Go application / LLGo | N/A | 9221 | N/A | 732584 |

**11/80 A/A controls resolve** within ±2.5%; 1/32 source pairs resolve both. Default-workstation timings are advisory; neither a controlled ranking nor release-performance acceptance is established. Every adverse, unchanged and unresolved result is retained below.

Go managed bytes and LLGo collector accounting differ; Go counters exclude Rust/JIT allocation. Native-Go is an ordinary application control, not an optimized native lower bound. Resource instrumentation runs separately. Throughput medians and paired geometric warm-time ratios summarize different reductions.

## All source comparisons

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A | Base B/request | Head B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| head/world-rust-composed-http/go/interpreter/1/16 | 3743 | 3760 | 0.9922 [0.9765,1.0074] | no | 102065 | 99886 |
| head/world-rust-composed-http/go/interpreter/1/65536 | 2902 | 2914 | 0.9885 [0.9660,1.0124] | no | 267285 | 265103 |
| head/world-rust-composed-http/go/interpreter/4/16 | 14189 | 14571 | 0.9664 [0.9401,0.9962] | no | 102067 | 99890 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10409 | 10175 | 1.0153 [0.9727,1.0566] | no | 267286 | 265107 |
| head/world-rust-composed-http/go/native/1/16 | 5018 | 5035 | 0.9884 [0.9747,1.0089] | yes | 97936 | 95757 |
| head/world-rust-composed-http/go/native/1/65536 | 4217 | 4267 | 0.9809 [0.9486,1.0002] | no | 263154 | 260973 |
| head/world-rust-composed-http/go/native/4/16 | 18562 | 19099 | 0.9715 [0.9590,0.9844] | no | 97940 | 95762 |
| head/world-rust-composed-http/go/native/4/65536 | 13814 | 14860 | 0.9977 [0.9267,1.1009] | no | 263158 | 260979 |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 3260 | 3307 | 0.9641 [0.9214,0.9994] | no | 177386 | 174647 |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 2534 | 2505 | 0.9904 [0.9535,1.0170] | no | 431019 | 428319 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 6150 | 6204 | 0.9941 [0.9633,1.0237] | no | 177298 | 174627 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4439 | 4609 | 0.9696 [0.9361,1.0228] | no | 430958 | 428248 |
| head/world-rust-composed-http/llgo/native/1/16 | 3988 | 3969 | 0.9978 [0.9830,1.0134] | no | 173160 | 170415 |
| head/world-rust-composed-http/llgo/native/1/65536 | 3258 | 3286 | 0.9889 [0.9721,1.0058] | no | 426761 | 424019 |
| head/world-rust-composed-http/llgo/native/4/16 | 5877 | 5943 | 1.0259 [0.9574,1.1374] | no | 172983 | 170365 |
| head/world-rust-composed-http/llgo/native/4/65536 | 4464 | 4571 | 0.9897 [0.9643,1.0188] | no | 426628 | 423931 |
| head/world-rust-http/go/interpreter/1/16 | 4360 | 4308 | 1.0233 [0.9991,1.0608] | no | 87553 | 85374 |
| head/world-rust-http/go/interpreter/1/65536 | 3250 | 3257 | 0.9991 [0.9828,1.0157] | no | 252774 | 250591 |
| head/world-rust-http/go/interpreter/4/16 | 16666 | 16929 | 0.9834 [0.9620,1.0038] | no | 87557 | 85390 |
| head/world-rust-http/go/interpreter/4/65536 | 11750 | 11558 | 1.0060 [0.9739,1.0416] | no | 252775 | 250595 |
| head/world-rust-http/go/native/1/16 | 5704 | 5871 | 0.9674 [0.9363,0.9900] | no | 87698 | 85518 |
| head/world-rust-http/go/native/1/65536 | 4727 | 4864 | 0.9789 [0.9615,0.9969] | no | 252917 | 250734 |
| head/world-rust-http/go/native/4/16 | 20898 | 22455 | 0.9640 [0.9451,0.9803] | no | 87699 | 85522 |
| head/world-rust-http/go/native/4/65536 | 15425 | 16392 | 0.9478 [0.9344,0.9661] | no | 252921 | 250738 |
| head/world-rust-http/llgo/interpreter/1/16 | 3677 | 3721 | 0.9940 [0.9835,1.0062] | no | 158210 | 155504 |
| head/world-rust-http/llgo/interpreter/1/65536 | 2779 | 2774 | 1.0116 [0.9879,1.0485] | no | 411898 | 409189 |
| head/world-rust-http/llgo/interpreter/4/16 | 6192 | 6394 | 1.0053 [0.9606,1.0732] | no | 158137 | 155424 |
| head/world-rust-http/llgo/interpreter/4/65536 | 4643 | 4836 | 0.9652 [0.9510,0.9798] | no | 411894 | 409179 |
| head/world-rust-http/llgo/native/1/16 | 4492 | 4624 | 0.9585 [0.9300,0.9782] | no | 158733 | 155985 |
| head/world-rust-http/llgo/native/1/65536 | 3533 | 3519 | 1.0179 [0.9752,1.0680] | no | 412327 | 409586 |
| head/world-rust-http/llgo/native/4/16 | 6221 | 6230 | 0.9901 [0.9528,1.0264] | no | 158634 | 155949 |
| head/world-rust-http/llgo/native/4/65536 | 4458 | 4792 | 0.9125 [0.8622,0.9479] | no | 412175 | 409508 |

## Every profile and separate resource cohort

| Profile | req/s | API us | p99 request us | Runtime B/request | Go mallocs/request | Between-batch threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/1/16 | 3743 | 265.361 | 5034.291 | 102065 | 1445.0 | 6 | 87.625 | yes |
| base/world-rust-composed-http/go/interpreter/1/65536 | 2902 | 341.262 | 6334.166 | 267285 | 1853.1 | 7 | 101.688 | no |
| base/world-rust-composed-http/go/interpreter/4/16 | 14189 | 272.828 | 2244.375 | 102067 | 1445.0 | 7 | 93.125 | yes |
| base/world-rust-composed-http/go/interpreter/4/65536 | 10409 | 369.492 | 3548.708 | 267286 | 1853.0 | 7 | 111.516 | no |
| base/world-rust-composed-http/go/native/1/16 | 5018 | 197.581 | 3863.917 | 97936 | 1451.0 | 7 | 144.016 | yes |
| base/world-rust-composed-http/go/native/1/65536 | 4217 | 233.531 | 4565.709 | 263154 | 1859.0 | 7 | 148.781 | no |
| base/world-rust-composed-http/go/native/4/16 | 18562 | 208.023 | 1713.666 | 97940 | 1451.0 | 7 | 151.672 | no |
| base/world-rust-composed-http/go/native/4/65536 | 13814 | 279.643 | 2490.000 | 263158 | 1859.0 | 7 | 158.922 | no |
| base/world-rust-composed-http/llgo/interpreter/1/16 | 3260 | 303.761 | 8780.917 | 177386 | N/A | 42 | 135.188 | no |
| base/world-rust-composed-http/llgo/interpreter/1/65536 | 2534 | 388.007 | 9480.959 | 431019 | N/A | 42 | 148.062 | no |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 6150 | 632.083 | 5734.375 | 177298 | N/A | 57 | 150.891 | no |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 4439 | 884.192 | 7340.666 | 430958 | N/A | 57 | 158.109 | no |
| base/world-rust-composed-http/llgo/native/1/16 | 3988 | 247.975 | 8044.792 | 173160 | N/A | 42 | 198.938 | no |
| base/world-rust-composed-http/llgo/native/1/65536 | 3258 | 300.853 | 8146.209 | 426761 | N/A | 42 | 180.547 | yes |
| base/world-rust-composed-http/llgo/native/4/16 | 5877 | 667.256 | 5882.959 | 172983 | N/A | 57 | 195.828 | no |
| base/world-rust-composed-http/llgo/native/4/65536 | 4464 | 869.149 | 6534.334 | 426628 | N/A | 57 | 226.000 | no |
| base/world-rust-http/go/interpreter/1/16 | 4360 | 227.527 | 4468.000 | 87553 | 1318.0 | 7 | 65.688 | no |
| base/world-rust-http/go/interpreter/1/65536 | 3250 | 304.680 | 5699.959 | 252774 | 1726.1 | 7 | 73.141 | no |
| base/world-rust-http/go/interpreter/4/16 | 16666 | 230.288 | 2188.333 | 87557 | 1318.0 | 7 | 73.516 | yes |
| base/world-rust-http/go/interpreter/4/65536 | 11750 | 327.321 | 2834.125 | 252775 | 1726.0 | 7 | 84.172 | no |
| base/world-rust-http/go/native/1/16 | 5704 | 173.497 | 3673.958 | 87698 | 1321.0 | 7 | 117.984 | no |
| base/world-rust-http/go/native/1/65536 | 4727 | 208.202 | 4010.583 | 252917 | 1729.1 | 7 | 109.906 | no |
| base/world-rust-http/go/native/4/16 | 20898 | 185.275 | 1611.291 | 87699 | 1321.0 | 7 | 117.812 | no |
| base/world-rust-http/go/native/4/65536 | 15425 | 248.961 | 2489.000 | 252921 | 1729.1 | 7 | 126.688 | no |
| base/world-rust-http/llgo/interpreter/1/16 | 3677 | 268.973 | 6952.458 | 158210 | N/A | 42 | 110.938 | no |
| base/world-rust-http/llgo/interpreter/1/65536 | 2779 | 348.129 | 8523.417 | 411898 | N/A | 42 | 123.953 | yes |
| base/world-rust-http/llgo/interpreter/4/16 | 6192 | 632.844 | 5342.541 | 158137 | N/A | 57 | 120.141 | no |
| base/world-rust-http/llgo/interpreter/4/65536 | 4643 | 832.613 | 6259.667 | 411894 | N/A | 57 | 140.812 | no |
| base/world-rust-http/llgo/native/1/16 | 4492 | 218.154 | 6221.625 | 158733 | N/A | 42 | 168.375 | no |
| base/world-rust-http/llgo/native/1/65536 | 3533 | 276.643 | 7303.417 | 412327 | N/A | 42 | 168.188 | no |
| base/world-rust-http/llgo/native/4/16 | 6221 | 628.774 | 5343.250 | 158634 | N/A | 57 | 176.719 | yes |
| base/world-rust-http/llgo/native/4/65536 | 4458 | 870.788 | 6012.792 | 412175 | N/A | 57 | 175.875 | no |
| head/world-rust-composed-http/go/interpreter/1/16 | 3760 | 263.853 | 4993.541 | 99886 | 1394.0 | 7 | 87.609 | no |
| head/world-rust-composed-http/go/interpreter/1/65536 | 2914 | 339.747 | 6596.250 | 265103 | 1802.0 | 7 | 96.250 | no |
| head/world-rust-composed-http/go/interpreter/4/16 | 14571 | 265.776 | 2035.042 | 99890 | 1394.0 | 7 | 101.281 | no |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10175 | 380.067 | 3628.000 | 265107 | 1802.0 | 7 | 111.594 | no |
| head/world-rust-composed-http/go/native/1/16 | 5035 | 196.837 | 3740.709 | 95757 | 1400.0 | 7 | 147.391 | yes |
| head/world-rust-composed-http/go/native/1/65536 | 4267 | 230.997 | 4470.500 | 260973 | 1808.0 | 8 | 140.984 | no |
| head/world-rust-composed-http/go/native/4/16 | 19099 | 202.197 | 1900.000 | 95762 | 1400.0 | 7 | 149.688 | no |
| head/world-rust-composed-http/go/native/4/65536 | 14860 | 256.867 | 2570.875 | 260979 | 1808.0 | 7 | 164.297 | no |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 3307 | 299.478 | 8160.584 | 174647 | N/A | 42 | 135.047 | no |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 2505 | 391.726 | 9284.583 | 428319 | N/A | 42 | 152.531 | no |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 6204 | 629.494 | 5612.041 | 174627 | N/A | 57 | 151.141 | no |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4609 | 840.963 | 7000.583 | 428248 | N/A | 57 | 151.297 | no |
| head/world-rust-composed-http/llgo/native/1/16 | 3969 | 249.056 | 7574.584 | 170415 | N/A | 42 | 183.062 | no |
| head/world-rust-composed-http/llgo/native/1/65536 | 3286 | 298.210 | 8016.000 | 424019 | N/A | 42 | 196.656 | no |
| head/world-rust-composed-http/llgo/native/4/16 | 5943 | 658.354 | 5771.083 | 170365 | N/A | 57 | 211.344 | no |
| head/world-rust-composed-http/llgo/native/4/65536 | 4571 | 831.252 | 6352.291 | 423931 | N/A | 57 | 199.812 | no |
| head/world-rust-http/go/interpreter/1/16 | 4308 | 230.218 | 4411.667 | 85374 | 1267.0 | 7 | 65.328 | no |
| head/world-rust-http/go/interpreter/1/65536 | 3257 | 303.863 | 5608.334 | 250591 | 1675.0 | 7 | 81.250 | yes |
| head/world-rust-http/go/interpreter/4/16 | 16929 | 228.603 | 1679.250 | 85390 | 1267.0 | 7 | 74.609 | no |
| head/world-rust-http/go/interpreter/4/65536 | 11558 | 333.068 | 2744.542 | 250595 | 1675.0 | 7 | 84.531 | no |
| head/world-rust-http/go/native/1/16 | 5871 | 168.487 | 3316.208 | 85518 | 1270.0 | 7 | 113.000 | no |
| head/world-rust-http/go/native/1/65536 | 4864 | 202.295 | 3834.209 | 250734 | 1678.0 | 7 | 111.359 | no |
| head/world-rust-http/go/native/4/16 | 22455 | 170.592 | 1823.083 | 85522 | 1270.0 | 7 | 118.484 | no |
| head/world-rust-http/go/native/4/65536 | 16392 | 233.501 | 1968.833 | 250738 | 1678.0 | 7 | 125.797 | no |
| head/world-rust-http/llgo/interpreter/1/16 | 3721 | 266.023 | 7263.500 | 155504 | N/A | 42 | 110.578 | no |
| head/world-rust-http/llgo/interpreter/1/65536 | 2774 | 354.200 | 8857.292 | 409189 | N/A | 42 | 123.984 | no |
| head/world-rust-http/llgo/interpreter/4/16 | 6394 | 610.283 | 5289.125 | 155424 | N/A | 57 | 119.672 | no |
| head/world-rust-http/llgo/interpreter/4/65536 | 4836 | 805.160 | 6441.916 | 409179 | N/A | 57 | 140.672 | no |
| head/world-rust-http/llgo/native/1/16 | 4624 | 213.308 | 6628.667 | 155985 | N/A | 42 | 184.438 | yes |
| head/world-rust-http/llgo/native/1/65536 | 3519 | 278.038 | 7480.833 | 409586 | N/A | 42 | 168.125 | no |
| head/world-rust-http/llgo/native/4/16 | 6230 | 628.607 | 5289.917 | 155949 | N/A | 57 | 175.203 | no |
| head/world-rust-http/llgo/native/4/65536 | 4792 | 808.526 | 6074.750 | 409508 | N/A | 57 | 175.953 | no |
| host/native-go/go/host/1/16 | 827523 | 0.564 | 42.167 | 2712 | 18.0 | 5 | 11.219 | no |
| host/native-go/go/host/1/65536 | 40794 | 20.958 | 556.375 | 277925 | 48.0 | 6 | 16.641 | no |
| host/native-go/go/host/4/16 | 1621473 | 0.675 | 34.166 | 2712 | 18.0 | 5 | 11.469 | no |
| host/native-go/go/host/4/65536 | 43943 | 77.979 | 673.791 | 277915 | 48.0 | 6 | 16.219 | no |
| host/native-go/llgo/host/1/16 | 169328 | 0.767 | 160.708 | 3462 | N/A | 37 | 12.406 | no |
| host/native-go/llgo/host/1/65536 | 8326 | 114.393 | 3303.792 | 732571 | N/A | 37 | 68.062 | no |
| host/native-go/llgo/host/4/16 | 120244 | 2.294 | 333.291 | 3499 | N/A | 40 | 12.562 | no |
| host/native-go/llgo/host/4/65536 | 9221 | 404.307 | 3357.667 | 732584 | N/A | 40 | 63.859 | no |
| reference/world-rust-composed-http/go/compiler/1/16 | 16791 | 58.419 | 1123.792 | 32 | 1.0 | 7 | 32.781 | no |
| reference/world-rust-composed-http/go/compiler/1/65536 | 14920 | 64.606 | 1279.208 | 32 | 1.0 | 7 | 32.453 | yes |
| reference/world-rust-composed-http/go/compiler/4/16 | 56054 | 68.178 | 364.042 | 33 | 1.0 | 7 | 33.234 | no |
| reference/world-rust-composed-http/go/compiler/4/65536 | 52057 | 72.017 | 359.167 | 32 | 1.0 | 7 | 33.578 | no |
| reference/world-rust-http/go/compiler/1/16 | 19963 | 49.037 | 891.500 | 32 | 1.0 | 7 | 29.375 | no |
| reference/world-rust-http/go/compiler/1/65536 | 17181 | 55.642 | 1058.125 | 32 | 1.0 | 7 | 30.234 | no |
| reference/world-rust-http/go/compiler/4/16 | 64413 | 58.958 | 305.625 | 44 | 1.0 | 10 | 30.484 | no |
| reference/world-rust-http/go/compiler/4/65536 | 59092 | 63.279 | 343.667 | 45 | 1.0 | 10 | 30.750 | no |

## Separate lifecycle intervals

Median milliseconds across five primary process reductions. Startup/readiness, load, engine setup, compile/validate, providers, instantiate, lookup, executor/workers, first batch, warm execution and joined cleanup remain separate. Process wall includes protocol parking; it is not standalone CLI latency. Host binary compilation is outside runtime sampling. No overlapping phases or cohorts are subtracted.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1/16 | 5.5069 | 0.0570 | 0.0004 | 26.4545 | 0.0376 | 0.2375 | 0.0004 | N/A | 0.0135 | 19.1178 | 17.0984 | 0.0303 | 209.9142 |
| base/world-rust-composed-http/go/interpreter/1/65536 | 5.4255 | 0.0668 | 0.0004 | 26.4417 | 0.0379 | 0.2226 | 0.0003 | N/A | 0.0126 | 25.3748 | 22.0509 | 0.0309 | 259.5607 |
| base/world-rust-composed-http/go/interpreter/4/16 | 5.4200 | 0.0648 | 0.0004 | 26.8302 | 0.0543 | 0.6834 | 0.0007 | N/A | 0.0140 | 6.9205 | 4.5104 | 0.0520 | 82.3517 |
| base/world-rust-composed-http/go/interpreter/4/65536 | 5.4270 | 0.0591 | 0.0003 | 25.7838 | 0.0578 | 0.6582 | 0.0008 | N/A | 0.0168 | 8.5727 | 6.1486 | 0.0466 | 98.1403 |
| base/world-rust-composed-http/go/native/1/16 | 5.6805 | 0.0597 | 0.0004 | 221.6672 | 0.0390 | 0.2100 | 0.0004 | N/A | 0.0125 | 14.1148 | 12.7547 | 0.0350 | 362.2816 |
| base/world-rust-composed-http/go/native/1/65536 | 5.5570 | 0.0619 | 0.0003 | 222.5322 | 0.0369 | 0.2152 | 0.0005 | N/A | 0.0123 | 17.3929 | 15.1751 | 0.0416 | 385.2514 |
| base/world-rust-composed-http/go/native/4/16 | 5.6038 | 0.0593 | 0.0004 | 221.9764 | 0.0508 | 0.6549 | 0.0007 | N/A | 0.0132 | 4.1001 | 3.4478 | 0.0674 | 267.3550 |
| base/world-rust-composed-http/go/native/4/65536 | 5.5162 | 0.0668 | 0.0004 | 222.1754 | 0.0508 | 0.6422 | 0.0007 | N/A | 0.0140 | 4.7172 | 4.6331 | 0.0626 | 278.8868 |
| base/world-rust-composed-http/llgo/interpreter/1/16 | 8.9699 | 0.0743 | 0.0003 | 43.9005 | 0.1150 | 0.4891 | 0.0005 | N/A | 0.2836 | 20.1800 | 19.6298 | 0.2524 | 252.9451 |
| base/world-rust-composed-http/llgo/interpreter/1/65536 | 9.1384 | 0.0685 | 0.0003 | 42.1788 | 0.1142 | 0.4471 | 0.0005 | N/A | 0.2872 | 28.7733 | 25.2579 | 0.2733 | 313.8208 |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 9.0048 | 0.0756 | 0.0003 | 43.0565 | 0.2478 | 1.5989 | 0.0010 | N/A | 0.3023 | 11.4209 | 10.4073 | 0.6070 | 164.3500 |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 9.0581 | 0.0723 | 0.0003 | 43.5285 | 0.2325 | 1.6051 | 0.0009 | N/A | 0.3039 | 18.4517 | 14.4163 | 0.5888 | 206.2635 |
| base/world-rust-composed-http/llgo/native/1/16 | 8.8629 | 0.0733 | 0.0003 | 329.1632 | 0.1175 | 0.2975 | 0.0005 | N/A | 0.2670 | 16.7780 | 16.0491 | 0.2570 | 505.9617 |
| base/world-rust-composed-http/llgo/native/1/65536 | 9.0378 | 0.0683 | 0.0003 | 327.6787 | 0.1039 | 0.2865 | 0.0005 | N/A | 0.2726 | 19.6705 | 19.6464 | 0.2417 | 541.9962 |
| base/world-rust-composed-http/llgo/native/4/16 | 9.0918 | 0.0719 | 0.0003 | 328.1392 | 0.2493 | 0.9905 | 0.0010 | N/A | 0.3232 | 11.6807 | 10.8890 | 0.6013 | 455.2177 |
| base/world-rust-composed-http/llgo/native/4/65536 | 9.0585 | 0.0678 | 0.0003 | 324.9197 | 0.2332 | 0.9457 | 0.0009 | N/A | 0.3014 | 14.4032 | 14.3354 | 0.6096 | 483.8550 |
| base/world-rust-http/go/interpreter/1/16 | 5.4743 | 0.0547 | 0.0003 | 18.2573 | 0.0380 | 0.1844 | 0.0005 | N/A | 0.0168 | 16.0466 | 14.6773 | 0.0318 | 174.8825 |
| base/world-rust-http/go/interpreter/1/65536 | 5.4739 | 0.0590 | 0.0004 | 18.4876 | 0.0366 | 0.1656 | 0.0004 | N/A | 0.0157 | 22.4555 | 19.6939 | 0.0344 | 227.6812 |
| base/world-rust-http/go/interpreter/4/16 | 5.5345 | 0.0605 | 0.0004 | 18.4947 | 0.0575 | 0.5188 | 0.0007 | N/A | 0.0153 | 4.3439 | 3.8401 | 0.0479 | 66.4859 |
| base/world-rust-http/go/interpreter/4/65536 | 5.4806 | 0.0550 | 0.0004 | 18.3098 | 0.0688 | 0.9880 | 0.0010 | N/A | 0.0152 | 7.1997 | 5.4469 | 0.0453 | 83.6072 |
| base/world-rust-http/go/native/1/16 | 5.5057 | 0.0590 | 0.0003 | 152.5845 | 0.0327 | 0.1712 | 0.0005 | N/A | 0.0138 | 11.9553 | 11.2202 | 0.0507 | 274.0352 |
| base/world-rust-http/go/native/1/65536 | 5.5122 | 0.0570 | 0.0004 | 152.1525 | 0.0330 | 0.1575 | 0.0004 | N/A | 0.0129 | 14.6570 | 13.5402 | 0.0420 | 299.4653 |
| base/world-rust-http/go/native/4/16 | 5.4590 | 0.0583 | 0.0004 | 151.2467 | 0.0583 | 0.4988 | 0.0007 | N/A | 0.0133 | 3.4981 | 3.0625 | 0.0564 | 189.5661 |
| base/world-rust-http/go/native/4/65536 | 5.6623 | 0.0486 | 0.0003 | 151.5408 | 0.0480 | 0.4738 | 0.0006 | N/A | 0.0133 | 5.0108 | 4.1490 | 0.0562 | 201.9012 |
| base/world-rust-http/llgo/interpreter/1/16 | 9.1650 | 0.0621 | 0.0003 | 30.5029 | 0.1055 | 0.3062 | 0.0005 | N/A | 0.2830 | 21.2916 | 17.4050 | 0.2529 | 219.2240 |
| base/world-rust-http/llgo/interpreter/1/65536 | 8.9449 | 0.0600 | 0.0003 | 29.7964 | 0.1091 | 2.5554 | 0.0005 | N/A | 0.2825 | 23.3896 | 23.0272 | 0.2410 | 275.7249 |
| base/world-rust-http/llgo/interpreter/4/16 | 8.9142 | 0.0611 | 0.0003 | 30.5918 | 0.2538 | 3.4120 | 0.0013 | N/A | 0.3131 | 10.3342 | 10.3354 | 0.5868 | 150.5165 |
| base/world-rust-http/llgo/interpreter/4/65536 | 9.1553 | 0.0707 | 0.0003 | 28.9879 | 0.2538 | 3.2146 | 0.0012 | N/A | 0.3197 | 13.9633 | 13.7845 | 0.6205 | 185.1047 |
| base/world-rust-http/llgo/native/1/16 | 9.1589 | 0.0684 | 0.0002 | 225.9784 | 0.1100 | 0.2144 | 0.0005 | N/A | 0.2963 | 15.6634 | 14.2489 | 0.2219 | 382.4738 |
| base/world-rust-http/llgo/native/1/65536 | 9.1171 | 0.0725 | 0.0003 | 226.7858 | 0.0997 | 0.2151 | 0.0005 | N/A | 0.2727 | 19.5645 | 18.1138 | 0.2427 | 420.7908 |
| base/world-rust-http/llgo/native/4/16 | 9.0203 | 0.0605 | 0.0003 | 224.5984 | 0.2634 | 0.7951 | 0.0010 | N/A | 0.3078 | 12.4385 | 10.2871 | 0.6045 | 342.4977 |
| base/world-rust-http/llgo/native/4/65536 | 9.0016 | 0.0627 | 0.0002 | 227.1140 | 0.2295 | 0.7102 | 0.0011 | N/A | 0.3068 | 16.8080 | 14.3559 | 0.5949 | 387.5470 |
| head/world-rust-composed-http/go/interpreter/1/16 | 5.5735 | 0.0655 | 0.0003 | 26.2790 | 0.0373 | 0.2311 | 0.0004 | N/A | 0.0133 | 18.2439 | 17.0192 | 0.0348 | 206.8388 |
| head/world-rust-composed-http/go/interpreter/1/65536 | 5.4439 | 0.0552 | 0.0003 | 26.0682 | 0.0395 | 0.2240 | 0.0005 | N/A | 0.0125 | 24.2713 | 21.9600 | 0.0298 | 256.3397 |
| head/world-rust-composed-http/go/interpreter/4/16 | 5.6198 | 0.0699 | 0.0003 | 26.3532 | 0.0537 | 0.7268 | 0.0007 | N/A | 0.0151 | 6.2953 | 4.3923 | 0.0496 | 81.0812 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 5.6092 | 0.0844 | 0.0004 | 27.0513 | 0.0565 | 0.6972 | 0.0008 | N/A | 0.0164 | 9.0165 | 6.2901 | 0.0530 | 99.6132 |
| head/world-rust-composed-http/go/native/1/16 | 5.3370 | 0.0603 | 0.0004 | 221.3588 | 0.0365 | 0.2082 | 0.0004 | N/A | 0.0118 | 13.7508 | 12.7108 | 0.0393 | 357.6554 |
| head/world-rust-composed-http/go/native/1/65536 | 5.6284 | 0.0622 | 0.0004 | 220.6077 | 0.0357 | 0.2005 | 0.0004 | N/A | 0.0122 | 16.3158 | 14.9996 | 0.0399 | 381.6251 |
| head/world-rust-composed-http/go/native/4/16 | 5.4876 | 0.0616 | 0.0004 | 224.3062 | 0.0546 | 0.6449 | 0.0007 | N/A | 0.0138 | 3.8472 | 3.3510 | 0.0615 | 267.6918 |
| head/world-rust-composed-http/go/native/4/65536 | 5.4588 | 0.0675 | 0.0003 | 224.9332 | 0.0540 | 0.6587 | 0.0007 | N/A | 0.0142 | 4.8533 | 4.3068 | 0.0679 | 279.9464 |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 8.7919 | 0.0626 | 0.0003 | 44.4385 | 0.1062 | 0.4943 | 0.0005 | N/A | 0.2654 | 20.3349 | 19.3506 | 0.2697 | 252.4523 |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 8.9914 | 0.0750 | 0.0003 | 41.9404 | 0.1055 | 0.4506 | 0.0005 | N/A | 0.2735 | 28.3883 | 25.5472 | 0.2761 | 316.3428 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 8.9256 | 0.0633 | 0.0003 | 42.9083 | 0.2547 | 1.5810 | 0.0011 | N/A | 0.3015 | 11.5117 | 10.3154 | 0.5790 | 164.4155 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 8.8340 | 0.0668 | 0.0003 | 43.4978 | 0.2431 | 1.5995 | 0.0010 | N/A | 0.3048 | 17.5519 | 13.8849 | 0.5665 | 200.2952 |
| head/world-rust-composed-http/llgo/native/1/16 | 8.8558 | 0.0695 | 0.0003 | 326.8758 | 0.1134 | 0.2922 | 0.0005 | N/A | 0.2705 | 17.3256 | 16.1248 | 0.2486 | 502.5026 |
| head/world-rust-composed-http/llgo/native/1/65536 | 8.9225 | 0.0673 | 0.0002 | 328.7398 | 0.1053 | 0.2854 | 0.0005 | N/A | 0.2847 | 19.0172 | 19.4766 | 0.2337 | 539.5252 |
| head/world-rust-composed-http/llgo/native/4/16 | 9.0018 | 0.0709 | 0.0003 | 330.0454 | 0.2603 | 0.9990 | 0.0011 | N/A | 0.3111 | 12.2172 | 10.7684 | 0.6002 | 453.2007 |
| head/world-rust-composed-http/llgo/native/4/65536 | 8.9881 | 0.0662 | 0.0003 | 328.6975 | 0.2334 | 1.0171 | 0.0010 | N/A | 0.3038 | 16.8696 | 14.0005 | 0.6020 | 484.5557 |
| head/world-rust-http/go/interpreter/1/16 | 5.7210 | 0.0591 | 0.0003 | 18.2844 | 0.0398 | 0.3018 | 0.0005 | N/A | 0.0162 | 15.5728 | 14.8563 | 0.0314 | 175.0505 |
| head/world-rust-http/go/interpreter/1/65536 | 5.5260 | 0.0657 | 0.0004 | 18.3630 | 0.0377 | 0.1545 | 0.0004 | N/A | 0.0130 | 22.2716 | 19.6519 | 0.0318 | 226.6807 |
| head/world-rust-http/go/interpreter/4/16 | 5.4915 | 0.0590 | 0.0004 | 18.5390 | 0.0633 | 0.7960 | 0.0008 | N/A | 0.0158 | 4.2115 | 3.7804 | 0.0447 | 66.4776 |
| head/world-rust-http/go/interpreter/4/65536 | 5.3639 | 0.0489 | 0.0004 | 18.1029 | 0.0689 | 1.0806 | 0.0010 | N/A | 0.0179 | 6.5240 | 5.5371 | 0.0653 | 83.0062 |
| head/world-rust-http/go/native/1/16 | 5.5143 | 0.0542 | 0.0003 | 153.7607 | 0.0315 | 0.1727 | 0.0005 | N/A | 0.0132 | 11.9906 | 10.9010 | 0.0419 | 273.4701 |
| head/world-rust-http/go/native/1/65536 | 5.6206 | 0.0532 | 0.0004 | 152.1485 | 0.0350 | 0.1588 | 0.0004 | N/A | 0.0132 | 14.0441 | 13.1573 | 0.0395 | 293.2057 |
| head/world-rust-http/go/native/4/16 | 5.6209 | 0.0551 | 0.0004 | 151.7462 | 0.0499 | 0.4685 | 0.0006 | N/A | 0.0134 | 4.0844 | 2.8501 | 0.0572 | 190.5013 |
| head/world-rust-http/go/native/4/65536 | 5.5599 | 0.0519 | 0.0004 | 151.7244 | 0.0542 | 0.4828 | 0.0006 | N/A | 0.0131 | 4.4091 | 3.9044 | 0.0659 | 200.9630 |
| head/world-rust-http/llgo/interpreter/1/16 | 9.0512 | 0.0614 | 0.0003 | 28.9643 | 0.1079 | 0.3011 | 0.0005 | N/A | 0.2705 | 21.7748 | 17.2013 | 0.2660 | 216.0562 |
| head/world-rust-http/llgo/interpreter/1/65536 | 8.9657 | 0.0653 | 0.0003 | 29.9127 | 0.1060 | 2.5416 | 0.0006 | N/A | 0.3122 | 23.3012 | 23.0704 | 0.2540 | 276.8219 |
| head/world-rust-http/llgo/interpreter/4/16 | 8.9137 | 0.0604 | 0.0003 | 29.9370 | 0.2404 | 3.4696 | 0.0010 | N/A | 0.2903 | 10.5110 | 10.0091 | 0.5878 | 145.1308 |
| head/world-rust-http/llgo/interpreter/4/65536 | 9.1263 | 0.0562 | 0.0003 | 29.3021 | 0.2470 | 3.8258 | 0.0011 | N/A | 0.3018 | 13.6133 | 13.2329 | 0.5615 | 181.9311 |
| head/world-rust-http/llgo/native/1/16 | 8.8716 | 0.0649 | 0.0003 | 226.2508 | 0.1162 | 0.2176 | 0.0005 | N/A | 0.2992 | 14.0990 | 13.8410 | 0.2708 | 379.6361 |
| head/world-rust-http/llgo/native/1/65536 | 8.8467 | 0.0843 | 0.0003 | 227.1563 | 0.1084 | 0.2095 | 0.0005 | N/A | 0.2850 | 21.8116 | 18.1879 | 0.2409 | 422.8800 |
| head/world-rust-http/llgo/native/4/16 | 8.8452 | 0.0671 | 0.0003 | 225.6379 | 0.2457 | 0.7147 | 0.0010 | N/A | 0.3233 | 13.1949 | 10.2736 | 0.5660 | 343.0154 |
| head/world-rust-http/llgo/native/4/65536 | 8.7750 | 0.0668 | 0.0003 | 225.4393 | 0.2256 | 0.6981 | 0.0010 | N/A | 0.2969 | 13.2615 | 13.3564 | 0.5660 | 375.6618 |
| host/native-go/go/host/1/16 | 5.3489 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0164 | 0.1008 | 0.0773 | 0.0042 | 7.2360 |
| host/native-go/go/host/1/65536 | 5.0007 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0127 | 1.8251 | 1.5689 | 0.0057 | 22.6944 |
| host/native-go/go/host/4/16 | 5.1569 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0232 | 0.0542 | 0.0395 | 0.0055 | 6.2846 |
| host/native-go/go/host/4/65536 | 5.0109 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0167 | 1.2372 | 1.4564 | 0.0058 | 20.6330 |
| host/native-go/llgo/host/1/16 | 8.7106 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2730 | 0.5107 | 0.3780 | 0.0804 | 14.2149 |
| host/native-go/llgo/host/1/65536 | 8.8559 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2707 | 10.0464 | 7.6870 | 0.0748 | 89.8652 |
| host/native-go/llgo/host/4/16 | 8.5535 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3051 | 0.6930 | 0.5323 | 0.0973 | 15.8566 |
| host/native-go/llgo/host/4/65536 | 8.7088 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3194 | 11.4084 | 6.9406 | 0.1008 | 85.0994 |
| reference/world-rust-composed-http/go/compiler/1/16 | 5.8465 | 0.0589 | 0.1379 | 73.2926 | 0.1143 | 0.1370 | 0.0063 | 0.0280 | 0.0365 | 4.4265 | 3.8116 | 0.2424 | 122.8271 |
| reference/world-rust-composed-http/go/compiler/1/65536 | 5.9828 | 0.0689 | 0.1425 | 74.3886 | 0.1175 | 0.1450 | 0.0067 | 0.0311 | 0.0348 | 4.8435 | 4.2897 | 0.2405 | 126.2187 |
| reference/world-rust-composed-http/go/compiler/4/16 | 5.8683 | 0.0655 | 0.1256 | 74.3003 | 0.2217 | 0.2563 | 0.0080 | 0.0419 | 0.0475 | 1.4525 | 1.1418 | 0.3548 | 94.6242 |
| reference/world-rust-composed-http/go/compiler/4/65536 | 5.8676 | 0.0581 | 0.1405 | 73.5167 | 0.2215 | 0.2456 | 0.0078 | 0.0417 | 0.0512 | 1.5800 | 1.2294 | 0.3718 | 94.5703 |
| reference/world-rust-http/go/compiler/1/16 | 5.8395 | 0.0573 | 0.1283 | 51.0162 | 0.1133 | 0.1169 | 0.0066 | 0.0292 | 0.0343 | 3.5666 | 3.2060 | 0.1604 | 91.2851 |
| reference/world-rust-http/go/compiler/1/65536 | 5.9071 | 0.0598 | 0.1395 | 51.4966 | 0.1160 | 0.1237 | 0.0065 | 0.0310 | 0.0395 | 4.0868 | 3.7251 | 0.1691 | 96.5755 |
| reference/world-rust-http/go/compiler/4/16 | 5.7873 | 0.0530 | 0.1240 | 50.6047 | 0.2237 | 0.1900 | 0.0079 | 0.0410 | 0.0519 | 1.2525 | 0.9936 | 0.2283 | 68.7823 |
| reference/world-rust-http/go/compiler/4/65536 | 5.8958 | 0.0522 | 0.1254 | 51.6226 | 0.2313 | 0.1939 | 0.0080 | 0.0448 | 0.0393 | 1.4705 | 1.0831 | 0.2282 | 70.8394 |

## Frozen protocol and validation

Benchmark base `23b54f0b121a1bd2481d2d445186f40dbcea9a94` is production-equivalent to main `da00a76`; its additional benchmark is identical to head `93fc720f2e82c7b678a0c6e433a66d3138f0a6a2`. Production LLGo stays `3d92a58`, `-O3 -lto=full`; Go 1.27.3, LLVM 22.1.8, Wasmtime 48.0.1. The base reuses prior exact binaries after checking every production/harness source hash; only the uncompiled benchmark test differs. No previous timing is reused. Reference compiler/program/SDK hashes remain pinned. Base/head measurement code is identical. Frozen reference differences are limited to owned-buffered flags/report metadata and cold options validation; warm driver, complete response checker and native/reference kernels are checked identical.

Five fixed randomized primary/A/A blocks, seed 2026111025; pointwise 95% paired log-ratio bootstrap, 8000 draws. Both Rust plain/composed HTTP fixtures, both owned backends/compilers, one/four workers, 16 clients/four requests, two warmups/seven samples, 16-byte/64-KiB payloads. Exactly 800 throughput processes, 80 separate resource processes, 160 qualifications, seven actual loader probes and 358400 checked warm responses. CRC32 independently checked. No source/profile/ISA/count search, favorable repeat, cross-event pooling, affinity/frequency isolation or family-wise claim.

Proportional local validation passes four affected Go and LLGo root packages (canonical/component/P2/P3), LLGo nested P3 measurement, two race packages and vet. Integration coverage of canonical/component exceeds 90% per package, minimum 92.36%; the initial unit-only canonical run measured 89.0%, so cross-package coverage was measured rather than adding tests that mirror the implementation. Every owned build/test/probe exited before timing. Actual pre-start process/cwd inspection records foreign workloads without interrupting them. Full final-head CI still requires all 32 root packages on both compilers/platforms, 31 coverage gates and all eight LLGo shards.

Existing semantic tests cover live/nil/expired references, copied aliases, affine movement, ended borrows and concurrent movement. This change neither removes ownership checks nor alters cancellation, quotas, cleanup or guest-memory contracts.

CI timing review of PR #153 found maximum 608 seconds, with root compile/link stages dominating slow shards and nested P3 compilation dominating legacy on macOS. Moving nested execution alone would move the bottleneck rather than demonstrate a shorter maximum; this PR preserves the complete matrix and adds no CI job. Five-to-ten-minute jobs remain the target.

Broader [all-runtime service measurements](performance-service-all-runtimes-2026-11-10.md) retain wazero/Wasmer and other execution portfolios. They are not pooled with this P3 event. The [compiler study](performance-llgo-merged-2026-11-10.md) and LLGo GC issue #3045 remain separate; no compiler upgrade is included. Native/Core encoding, caller/codec allocation, interpreter, SIMD/memory and controlled performance gaps remain open.
