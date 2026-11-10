# Resource inspection and HTTP failure delivery, 2026-11-10

[Complete machine record](data/performance-reference-inspection-corrected-2026-11-10.json).

This intermediate source event predates the final cancellation review. Its receiver-cancellation approach was replaced because the worker must still interrupt an already published response after a later guest failure. The [final atomic-abandonment event](performance-reference-inspection-atomic-2026-11-10.md) preserves that behavior while retaining the committed failure. These measurements remain historical and are not pooled with final-source results.

The [initial source event](performance-reference-inspection-2026-11-10.md) is retained independently. Initial CI failed because a worker cancelled its internal context immediately after publishing a specific Store-busy failure; the receiver could nondeterministically report `context canceled`. The receiver now owns cancellation after consuming a published failure. Error cleanup and real external cancellation remain enforced. This new fixed event measures that corrected production source, without pooling or replacing initial observations.

## Change and evidence

Normal affine cleanup inspects aliases invalidated by successful moves. `ReferenceInfo` previously formatted the same expired-capability error for each inspection. It now reuses one private immutable diagnostic, created during package initialization. The error message, `errors.Is(wasm.ErrTrap)`, mutexes, resource validity and active borrow-scope checks are unchanged. Live references and invalid nil references keep their existing results. No capability, caller or per-call mutable state is cached.

The preceding [runtime-only allocation profile](performance-byte-destination-2026-11-10.md) excludes compilation/setup and identifies this formatting under `publishReference` and `cleanupResource`. It is normal ownership bookkeeping, not evidence of failed HTTP requests or an LLGo defect. Core value encoding, caller and codec allocation remain independent hotspots.

## Inspection allocation diagnostic

The initial identical checked-in benchmark ran 50000 inspections per case on each source/compiler. Its allocation counters are reused only after verifying all canonical/dependency/benchmark source hashes are unchanged; the follow-up changes only HTTP error publication. No earlier throughput timing is reused. Setup is outside the measured loop. These are allocation counters; retained ns/op are diagnostic, without a timing-ranking claim. LLGo Mallocs are unavailable and raw zero counters do not mean zero allocations.

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
| wasmvm go / interpreter | 16531 | 16622 | 87557 | 85379 |
| wasmvm go / native | 19327 | 20307 | 87701 | 85524 |
| wasmvm llgo / interpreter | 5836 | 6219 | 158134 | 155397 |
| wasmvm llgo / native | 5839 | 6141 | 158658 | 155916 |
| Wasmtime 48.0.1 | N/A | 65047 | N/A | 45 |
| Ordinary native-Go application / Go | N/A | 1381829 | N/A | 2712 |
| Ordinary native-Go application / LLGo | N/A | 116687 | N/A | 3518 |

## HTTP fixture, 65536 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 11627 | 11684 | 252775 | 250597 |
| wasmvm go / native | 15282 | 15064 | 252920 | 250739 |
| wasmvm llgo / interpreter | 4497 | 4603 | 411910 | 409180 |
| wasmvm llgo / native | 4452 | 4414 | 412248 | 409567 |
| Wasmtime 48.0.1 | N/A | 58146 | N/A | 45 |
| Ordinary native-Go application / Go | N/A | 41441 | N/A | 277927 |
| Ordinary native-Go application / LLGo | N/A | 9365 | N/A | 732575 |

**22/80 A/A controls resolve** within ±2.5%; 8/32 source pairs resolve both. Default-workstation timings are advisory; neither a controlled ranking nor release-performance acceptance is established. Every adverse, unchanged and unresolved result is retained below.

Go managed bytes and LLGo collector accounting differ; Go counters exclude Rust/JIT allocation. Native-Go is an ordinary application control, not an optimized native lower bound. Resource instrumentation runs separately. Throughput medians and paired geometric warm-time ratios summarize different reductions.

## All source comparisons

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A | Base B/request | Head B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| head/world-rust-composed-http/go/interpreter/1/16 | 3672 | 3714 | 0.9845 [0.9672,0.9995] | yes | 102068 | 99887 |
| head/world-rust-composed-http/go/interpreter/1/65536 | 2816 | 2880 | 0.9928 [0.9708,1.0213] | yes | 267287 | 265103 |
| head/world-rust-composed-http/go/interpreter/4/16 | 14096 | 14132 | 0.9997 [0.9889,1.0106] | no | 102068 | 99889 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10160 | 10067 | 1.0138 [0.9974,1.0357] | yes | 267286 | 265107 |
| head/world-rust-composed-http/go/native/1/16 | 4835 | 4835 | 0.9851 [0.9694,1.0009] | no | 97938 | 95758 |
| head/world-rust-composed-http/go/native/1/65536 | 4091 | 4203 | 0.9743 [0.9487,1.0005] | no | 263155 | 260974 |
| head/world-rust-composed-http/go/native/4/16 | 17495 | 17714 | 0.9915 [0.9742,1.0052] | no | 97939 | 95763 |
| head/world-rust-composed-http/go/native/4/65536 | 13233 | 13732 | 0.9625 [0.8969,1.0262] | no | 263169 | 260979 |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 3232 | 3220 | 1.0056 [0.9956,1.0177] | no | 177379 | 174662 |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 2535 | 2563 | 0.9884 [0.9667,1.0052] | no | 431037 | 428311 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 5820 | 5870 | 0.9913 [0.9547,1.0303] | no | 177387 | 174634 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4361 | 4355 | 0.9988 [0.9730,1.0357] | no | 430919 | 428229 |
| head/world-rust-composed-http/llgo/native/1/16 | 3904 | 3981 | 0.9806 [0.9613,1.0055] | no | 173137 | 170386 |
| head/world-rust-composed-http/llgo/native/1/65536 | 3218 | 3260 | 0.9818 [0.9512,1.0128] | no | 426728 | 424035 |
| head/world-rust-composed-http/llgo/native/4/16 | 5538 | 5715 | 0.9574 [0.9176,1.0024] | no | 173039 | 170356 |
| head/world-rust-composed-http/llgo/native/4/65536 | 4308 | 4324 | 1.0091 [0.9726,1.0567] | no | 426630 | 423912 |
| head/world-rust-http/go/interpreter/1/16 | 4235 | 4237 | 1.0011 [0.9954,1.0090] | no | 87555 | 85375 |
| head/world-rust-http/go/interpreter/1/65536 | 3161 | 3200 | 0.9964 [0.9872,1.0074] | yes | 252772 | 250591 |
| head/world-rust-http/go/interpreter/4/16 | 16531 | 16622 | 0.9793 [0.9567,0.9970] | no | 87557 | 85379 |
| head/world-rust-http/go/interpreter/4/65536 | 11627 | 11684 | 1.0175 [0.9491,1.1218] | no | 252775 | 250597 |
| head/world-rust-http/go/native/1/16 | 5591 | 5705 | 0.9864 [0.9632,1.0107] | yes | 87699 | 85518 |
| head/world-rust-http/go/native/1/65536 | 4531 | 4667 | 0.9747 [0.9550,0.9948] | yes | 252917 | 250735 |
| head/world-rust-http/go/native/4/16 | 19327 | 20307 | 0.9745 [0.9277,1.0170] | no | 87701 | 85524 |
| head/world-rust-http/go/native/4/65536 | 15282 | 15064 | 1.0019 [0.9629,1.0425] | no | 252920 | 250739 |
| head/world-rust-http/llgo/interpreter/1/16 | 3607 | 3654 | 0.9907 [0.9731,1.0122] | no | 158218 | 155489 |
| head/world-rust-http/llgo/interpreter/1/65536 | 2758 | 2779 | 0.9967 [0.9867,1.0069] | yes | 411901 | 409192 |
| head/world-rust-http/llgo/interpreter/4/16 | 5836 | 6219 | 0.9418 [0.9098,0.9826] | no | 158134 | 155397 |
| head/world-rust-http/llgo/interpreter/4/65536 | 4497 | 4603 | 0.9849 [0.9655,1.0087] | no | 411910 | 409180 |
| head/world-rust-http/llgo/native/1/16 | 4445 | 4532 | 0.9830 [0.9661,1.0021] | no | 158740 | 156020 |
| head/world-rust-http/llgo/native/1/65536 | 3516 | 3544 | 0.9914 [0.9795,1.0008] | yes | 412288 | 409618 |
| head/world-rust-http/llgo/native/4/16 | 5839 | 6141 | 0.9495 [0.9156,0.9861] | no | 158658 | 155916 |
| head/world-rust-http/llgo/native/4/65536 | 4452 | 4414 | 0.9878 [0.9421,1.0399] | no | 412248 | 409567 |

## Every profile and separate resource cohort

| Profile | req/s | API us | p99 request us | Runtime B/request | Go mallocs/request | Between-batch threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/1/16 | 3672 | 270.428 | 5355.125 | 102068 | 1445.1 | 7 | 87.875 | yes |
| base/world-rust-composed-http/go/interpreter/1/65536 | 2816 | 351.715 | 6613.250 | 267287 | 1853.1 | 7 | 96.562 | yes |
| base/world-rust-composed-http/go/interpreter/4/16 | 14096 | 273.559 | 1999.375 | 102068 | 1445.0 | 8 | 101.812 | no |
| base/world-rust-composed-http/go/interpreter/4/65536 | 10160 | 379.220 | 3677.542 | 267286 | 1853.0 | 8 | 121.547 | yes |
| base/world-rust-composed-http/go/native/1/16 | 4835 | 204.812 | 4168.208 | 97938 | 1451.1 | 6 | 147.828 | no |
| base/world-rust-composed-http/go/native/1/65536 | 4091 | 241.063 | 4665.209 | 263155 | 1859.0 | 7 | 151.875 | no |
| base/world-rust-composed-http/go/native/4/16 | 17495 | 218.399 | 2208.041 | 97939 | 1451.0 | 7 | 157.469 | no |
| base/world-rust-composed-http/go/native/4/65536 | 13233 | 289.617 | 3083.666 | 263169 | 1859.0 | 7 | 157.094 | no |
| base/world-rust-composed-http/llgo/interpreter/1/16 | 3232 | 306.109 | 8081.417 | 177379 | N/A | 42 | 135.062 | no |
| base/world-rust-composed-http/llgo/interpreter/1/65536 | 2535 | 387.899 | 9505.333 | 431037 | N/A | 42 | 148.141 | yes |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 5820 | 673.571 | 5838.708 | 177387 | N/A | 57 | 151.078 | no |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 4361 | 898.597 | 6806.750 | 430919 | N/A | 57 | 158.375 | no |
| base/world-rust-composed-http/llgo/native/1/16 | 3904 | 252.864 | 7465.500 | 173137 | N/A | 42 | 199.203 | no |
| base/world-rust-composed-http/llgo/native/1/65536 | 3218 | 304.240 | 8438.583 | 426728 | N/A | 42 | 197.078 | no |
| base/world-rust-composed-http/llgo/native/4/16 | 5538 | 705.352 | 5712.167 | 173039 | N/A | 57 | 195.312 | no |
| base/world-rust-composed-http/llgo/native/4/65536 | 4308 | 902.497 | 6600.250 | 426630 | N/A | 57 | 195.656 | no |
| base/world-rust-http/go/interpreter/1/16 | 4235 | 233.857 | 4417.333 | 87555 | 1318.0 | 7 | 67.188 | yes |
| base/world-rust-http/go/interpreter/1/65536 | 3161 | 313.100 | 5611.375 | 252772 | 1726.0 | 7 | 71.969 | yes |
| base/world-rust-http/go/interpreter/4/16 | 16531 | 234.504 | 1914.000 | 87557 | 1318.0 | 7 | 73.594 | no |
| base/world-rust-http/go/interpreter/4/65536 | 11627 | 331.656 | 2684.084 | 252775 | 1726.0 | 7 | 91.672 | no |
| base/world-rust-http/go/native/1/16 | 5591 | 176.844 | 3405.334 | 87699 | 1321.1 | 7 | 114.703 | yes |
| base/world-rust-http/go/native/1/65536 | 4531 | 217.268 | 4211.917 | 252917 | 1729.1 | 7 | 123.984 | yes |
| base/world-rust-http/go/native/4/16 | 19327 | 199.660 | 2150.417 | 87701 | 1321.0 | 7 | 117.594 | no |
| base/world-rust-http/go/native/4/65536 | 15282 | 250.329 | 2364.000 | 252920 | 1729.0 | 7 | 126.250 | no |
| base/world-rust-http/llgo/interpreter/1/16 | 3607 | 273.968 | 7711.875 | 158218 | N/A | 42 | 110.969 | no |
| base/world-rust-http/llgo/interpreter/1/65536 | 2758 | 354.913 | 8493.459 | 411901 | N/A | 42 | 123.891 | yes |
| base/world-rust-http/llgo/interpreter/4/16 | 5836 | 669.704 | 5443.833 | 158134 | N/A | 57 | 120.078 | no |
| base/world-rust-http/llgo/interpreter/4/65536 | 4497 | 866.660 | 6636.291 | 411910 | N/A | 57 | 140.844 | no |
| base/world-rust-http/llgo/native/1/16 | 4445 | 221.875 | 6595.583 | 158740 | N/A | 42 | 168.078 | yes |
| base/world-rust-http/llgo/native/1/65536 | 3516 | 275.940 | 7519.125 | 412288 | N/A | 42 | 167.250 | yes |
| base/world-rust-http/llgo/native/4/16 | 5839 | 670.650 | 5624.875 | 158658 | N/A | 57 | 175.234 | no |
| base/world-rust-http/llgo/native/4/65536 | 4452 | 875.107 | 6743.292 | 412248 | N/A | 57 | 175.953 | no |
| head/world-rust-composed-http/go/interpreter/1/16 | 3714 | 266.862 | 5181.125 | 99887 | 1394.0 | 6 | 87.531 | yes |
| head/world-rust-composed-http/go/interpreter/1/65536 | 2880 | 343.869 | 6632.583 | 265103 | 1802.0 | 7 | 95.656 | yes |
| head/world-rust-composed-http/go/interpreter/4/16 | 14132 | 273.220 | 2162.542 | 99889 | 1394.0 | 7 | 96.172 | yes |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10067 | 384.493 | 3655.708 | 265107 | 1802.0 | 7 | 112.828 | yes |
| head/world-rust-composed-http/go/native/1/16 | 4835 | 204.658 | 4229.959 | 95758 | 1400.0 | 7 | 136.641 | no |
| head/world-rust-composed-http/go/native/1/65536 | 4203 | 234.546 | 4686.958 | 260974 | 1808.0 | 7 | 150.672 | yes |
| head/world-rust-composed-http/go/native/4/16 | 17714 | 218.242 | 2040.709 | 95763 | 1400.0 | 7 | 159.797 | no |
| head/world-rust-composed-http/go/native/4/65536 | 13732 | 279.236 | 3153.375 | 260979 | 1808.0 | 7 | 163.688 | no |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 3220 | 307.195 | 8515.208 | 174662 | N/A | 42 | 135.203 | yes |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 2563 | 381.711 | 9349.583 | 428311 | N/A | 42 | 148.094 | no |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 5870 | 665.922 | 6076.291 | 174634 | N/A | 57 | 151.078 | no |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4355 | 892.612 | 7238.750 | 428229 | N/A | 57 | 157.875 | no |
| head/world-rust-composed-http/llgo/native/1/16 | 3981 | 247.939 | 7246.750 | 170386 | N/A | 42 | 198.922 | no |
| head/world-rust-composed-http/llgo/native/1/65536 | 3260 | 300.385 | 7785.709 | 424035 | N/A | 42 | 180.625 | no |
| head/world-rust-composed-http/llgo/native/4/16 | 5715 | 684.614 | 6019.416 | 170356 | N/A | 57 | 211.234 | no |
| head/world-rust-composed-http/llgo/native/4/65536 | 4324 | 900.728 | 6887.792 | 423912 | N/A | 57 | 210.688 | no |
| head/world-rust-http/go/interpreter/1/16 | 4237 | 233.938 | 4525.459 | 85375 | 1267.0 | 7 | 66.156 | no |
| head/world-rust-http/go/interpreter/1/65536 | 3200 | 309.224 | 5688.500 | 250591 | 1675.0 | 7 | 72.969 | yes |
| head/world-rust-http/go/interpreter/4/16 | 16622 | 232.243 | 1906.209 | 85379 | 1267.0 | 7 | 73.234 | no |
| head/world-rust-http/go/interpreter/4/65536 | 11684 | 328.638 | 2839.083 | 250597 | 1675.0 | 7 | 85.094 | no |
| head/world-rust-http/go/native/1/16 | 5705 | 173.343 | 3513.375 | 85518 | 1270.0 | 7 | 116.531 | yes |
| head/world-rust-http/go/native/1/65536 | 4667 | 210.877 | 4100.291 | 250735 | 1678.0 | 7 | 119.891 | yes |
| head/world-rust-http/go/native/4/16 | 20307 | 190.558 | 1891.791 | 85524 | 1270.0 | 7 | 129.422 | no |
| head/world-rust-http/go/native/4/65536 | 15064 | 255.089 | 2598.042 | 250739 | 1678.0 | 8 | 125.438 | no |
| head/world-rust-http/llgo/interpreter/1/16 | 3654 | 270.260 | 7128.042 | 155489 | N/A | 42 | 110.266 | no |
| head/world-rust-http/llgo/interpreter/1/65536 | 2779 | 349.143 | 8691.584 | 409192 | N/A | 42 | 124.109 | yes |
| head/world-rust-http/llgo/interpreter/4/16 | 6219 | 628.197 | 5271.042 | 155397 | N/A | 57 | 119.703 | no |
| head/world-rust-http/llgo/interpreter/4/65536 | 4603 | 850.250 | 6552.958 | 409180 | N/A | 57 | 140.844 | no |
| head/world-rust-http/llgo/native/1/16 | 4532 | 217.388 | 6646.084 | 156020 | N/A | 42 | 168.078 | no |
| head/world-rust-http/llgo/native/1/65536 | 3544 | 275.604 | 6995.959 | 409618 | N/A | 42 | 168.391 | yes |
| head/world-rust-http/llgo/native/4/16 | 6141 | 636.971 | 5193.417 | 155916 | N/A | 57 | 176.344 | no |
| head/world-rust-http/llgo/native/4/65536 | 4414 | 886.502 | 6160.250 | 409567 | N/A | 57 | 176.516 | no |
| host/native-go/go/host/1/16 | 888078 | 0.545 | 32.625 | 2712 | 18.0 | 5 | 11.312 | no |
| host/native-go/go/host/1/65536 | 39781 | 21.312 | 554.375 | 277925 | 48.0 | 6 | 16.703 | no |
| host/native-go/go/host/4/16 | 1381829 | 0.778 | 40.125 | 2712 | 18.0 | 6 | 11.547 | no |
| host/native-go/go/host/4/65536 | 41441 | 83.483 | 619.083 | 277927 | 48.1 | 6 | 16.750 | no |
| host/native-go/llgo/host/1/16 | 165019 | 0.800 | 164.042 | 3479 | N/A | 37 | 12.484 | no |
| host/native-go/llgo/host/1/65536 | 8159 | 116.664 | 3949.584 | 732566 | N/A | 37 | 68.078 | no |
| host/native-go/llgo/host/4/16 | 116687 | 2.328 | 328.958 | 3518 | N/A | 40 | 12.578 | no |
| host/native-go/llgo/host/4/65536 | 9365 | 388.429 | 3248.917 | 732575 | N/A | 40 | 68.719 | no |
| reference/world-rust-composed-http/go/compiler/1/16 | 16251 | 60.110 | 1210.291 | 32 | 1.0 | 7 | 32.859 | no |
| reference/world-rust-composed-http/go/compiler/1/65536 | 14268 | 67.343 | 1314.500 | 32 | 1.0 | 7 | 33.094 | no |
| reference/world-rust-composed-http/go/compiler/4/16 | 55949 | 67.870 | 349.625 | 32 | 1.0 | 9 | 33.844 | no |
| reference/world-rust-composed-http/go/compiler/4/65536 | 50337 | 74.434 | 406.833 | 32 | 1.0 | 7 | 33.875 | no |
| reference/world-rust-http/go/compiler/1/16 | 19465 | 50.125 | 984.625 | 32 | 1.0 | 7 | 29.969 | no |
| reference/world-rust-http/go/compiler/1/65536 | 16891 | 56.596 | 1154.584 | 32 | 1.0 | 8 | 30.453 | no |
| reference/world-rust-http/go/compiler/4/16 | 65047 | 58.141 | 317.208 | 45 | 1.0 | 7 | 30.469 | no |
| reference/world-rust-http/go/compiler/4/65536 | 58146 | 63.956 | 378.000 | 45 | 1.0 | 7 | 31.078 | no |

## Separate lifecycle intervals

Median milliseconds across five primary process reductions. Startup/readiness, load, engine setup, compile/validate, providers, instantiate, lookup, executor/workers, first batch, warm execution and joined cleanup remain separate. Process wall includes protocol parking; it is not standalone CLI latency. Host binary compilation is outside runtime sampling. No overlapping phases or cohorts are subtracted.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1/16 | 5.7673 | 0.0628 | 0.0004 | 27.1431 | 0.0423 | 0.2453 | 0.0004 | N/A | 0.0146 | 19.4994 | 17.4308 | 0.0400 | 213.4105 |
| base/world-rust-composed-http/go/interpreter/1/65536 | 5.8727 | 0.0658 | 0.0004 | 26.8321 | 0.0393 | 0.2418 | 0.0003 | N/A | 0.0135 | 25.2647 | 22.7309 | 0.0363 | 266.7990 |
| base/world-rust-composed-http/go/interpreter/4/16 | 5.7881 | 0.0663 | 0.0004 | 26.9916 | 0.0586 | 0.6878 | 0.0006 | N/A | 0.0176 | 6.3660 | 4.5402 | 0.0561 | 83.6315 |
| base/world-rust-composed-http/go/interpreter/4/65536 | 5.9067 | 0.0702 | 0.0005 | 27.1243 | 0.0557 | 0.7049 | 0.0006 | N/A | 0.0152 | 9.0073 | 6.2992 | 0.0530 | 101.4747 |
| base/world-rust-composed-http/go/native/1/16 | 5.8962 | 0.0662 | 0.0005 | 223.1854 | 0.0376 | 0.2132 | 0.0005 | N/A | 0.0125 | 14.3992 | 13.2378 | 0.0412 | 366.0762 |
| base/world-rust-composed-http/go/native/1/65536 | 5.7227 | 0.0622 | 0.0004 | 223.7562 | 0.0357 | 0.2116 | 0.0004 | N/A | 0.0125 | 16.6965 | 15.6453 | 0.0400 | 390.9810 |
| base/world-rust-composed-http/go/native/4/16 | 5.7824 | 0.0672 | 0.0004 | 224.7974 | 0.0527 | 0.6520 | 0.0005 | N/A | 0.0140 | 4.1744 | 3.6582 | 0.0710 | 272.6319 |
| base/world-rust-composed-http/go/native/4/65536 | 5.6888 | 0.0617 | 0.0004 | 223.3746 | 0.0542 | 0.6595 | 0.0006 | N/A | 0.0132 | 5.0782 | 4.8363 | 0.0665 | 282.6389 |
| base/world-rust-composed-http/llgo/interpreter/1/16 | 9.3165 | 0.0813 | 0.0002 | 44.7639 | 0.1271 | 0.4990 | 0.0005 | N/A | 0.3085 | 20.8655 | 19.8018 | 0.2581 | 256.0487 |
| base/world-rust-composed-http/llgo/interpreter/1/65536 | 9.4437 | 0.0800 | 0.0003 | 43.5766 | 0.1125 | 0.4619 | 0.0007 | N/A | 0.2827 | 30.0313 | 25.2443 | 0.2562 | 314.2647 |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 9.1546 | 0.0729 | 0.0003 | 43.8787 | 0.2612 | 1.6393 | 0.0012 | N/A | 0.3200 | 12.0629 | 10.9967 | 0.6026 | 171.3229 |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 9.0911 | 0.0797 | 0.0003 | 41.6985 | 0.2562 | 1.5550 | 0.0009 | N/A | 0.3099 | 17.7776 | 14.6749 | 0.6130 | 209.7732 |
| base/world-rust-composed-http/llgo/native/1/16 | 9.4155 | 0.0775 | 0.0003 | 330.2359 | 0.1123 | 0.2860 | 0.0005 | N/A | 0.2813 | 16.5695 | 16.3915 | 0.2753 | 508.5242 |
| base/world-rust-composed-http/llgo/native/1/65536 | 9.0995 | 0.0895 | 0.0003 | 328.4242 | 0.1023 | 0.2975 | 0.0005 | N/A | 0.2600 | 22.9563 | 19.8869 | 0.2516 | 546.5710 |
| base/world-rust-composed-http/llgo/native/4/16 | 9.2476 | 0.0720 | 0.0003 | 327.7108 | 0.2384 | 0.9766 | 0.0010 | N/A | 0.3058 | 12.9176 | 11.5565 | 0.6081 | 461.0901 |
| base/world-rust-composed-http/llgo/native/4/65536 | 9.1175 | 0.0725 | 0.0003 | 326.4968 | 0.2494 | 1.0108 | 0.0012 | N/A | 0.3137 | 16.2517 | 14.8559 | 0.5927 | 490.7109 |
| base/world-rust-http/go/interpreter/1/16 | 5.5995 | 0.0552 | 0.0004 | 18.7832 | 0.0430 | 0.1808 | 0.0004 | N/A | 0.0130 | 16.5755 | 15.1118 | 0.0312 | 180.3761 |
| base/world-rust-http/go/interpreter/1/65536 | 5.6478 | 0.0689 | 0.0005 | 18.6130 | 0.0398 | 0.1602 | 0.0004 | N/A | 0.0131 | 22.8592 | 20.2491 | 0.0332 | 232.3713 |
| base/world-rust-http/go/interpreter/4/16 | 5.6902 | 0.0607 | 0.0005 | 18.7850 | 0.0683 | 0.6910 | 0.0008 | N/A | 0.0167 | 4.4732 | 3.8716 | 0.0516 | 68.1815 |
| base/world-rust-http/go/interpreter/4/65536 | 5.6850 | 0.0572 | 0.0005 | 18.5785 | 0.0593 | 0.9045 | 0.0008 | N/A | 0.0173 | 7.2884 | 5.5042 | 0.0477 | 84.7344 |
| base/world-rust-http/go/native/1/16 | 5.7291 | 0.0580 | 0.0004 | 153.0218 | 0.0334 | 0.1655 | 0.0004 | N/A | 0.0152 | 12.4374 | 11.4472 | 0.0617 | 277.6884 |
| base/world-rust-http/go/native/1/65536 | 5.8900 | 0.0681 | 0.0004 | 154.1538 | 0.0375 | 0.1726 | 0.0004 | N/A | 0.0128 | 15.2341 | 14.1236 | 0.0426 | 306.1659 |
| base/world-rust-http/go/native/4/16 | 5.7650 | 0.0655 | 0.0005 | 154.1011 | 0.0497 | 0.5014 | 0.0008 | N/A | 0.0193 | 4.5438 | 3.3115 | 0.0625 | 197.2632 |
| base/world-rust-http/go/native/4/65536 | 5.5694 | 0.0645 | 0.0004 | 155.1036 | 0.0518 | 0.4848 | 0.0005 | N/A | 0.0137 | 4.2428 | 4.1878 | 0.0940 | 207.3692 |
| base/world-rust-http/llgo/interpreter/1/16 | 9.2529 | 0.0736 | 0.0003 | 29.2033 | 0.1076 | 0.3118 | 0.0005 | N/A | 0.2812 | 21.9984 | 17.7418 | 0.2641 | 222.6470 |
| base/world-rust-http/llgo/interpreter/1/65536 | 9.7009 | 0.0805 | 0.0003 | 31.0436 | 0.1276 | 2.5633 | 0.0005 | N/A | 0.2917 | 23.6381 | 23.2037 | 0.2698 | 279.8800 |
| base/world-rust-http/llgo/interpreter/4/16 | 9.4183 | 0.0732 | 0.0003 | 30.9451 | 0.2785 | 3.2697 | 0.0012 | N/A | 0.3360 | 11.1327 | 10.9658 | 0.6031 | 156.0260 |
| base/world-rust-http/llgo/interpreter/4/65536 | 9.3944 | 0.0711 | 0.0003 | 29.7693 | 0.2910 | 3.6071 | 0.0011 | N/A | 0.3179 | 13.8455 | 14.2315 | 0.6096 | 188.6724 |
| base/world-rust-http/llgo/native/1/16 | 8.9843 | 0.0697 | 0.0002 | 227.2368 | 0.1139 | 0.2171 | 0.0005 | N/A | 0.2588 | 15.1053 | 14.3969 | 0.2524 | 386.6261 |
| base/world-rust-http/llgo/native/1/65536 | 9.3554 | 0.0645 | 0.0003 | 227.5654 | 0.1039 | 0.2172 | 0.0005 | N/A | 0.2925 | 22.6312 | 18.2020 | 0.2350 | 425.6796 |
| base/world-rust-http/llgo/native/4/16 | 9.1935 | 0.0737 | 0.0003 | 226.1423 | 0.2428 | 0.7130 | 0.0008 | N/A | 0.3187 | 13.5173 | 10.9608 | 0.6106 | 351.3915 |
| base/world-rust-http/llgo/native/4/65536 | 9.1449 | 0.0660 | 0.0002 | 227.9322 | 0.2211 | 0.7395 | 0.0010 | N/A | 0.3203 | 14.8133 | 14.3771 | 0.6295 | 387.3470 |
| head/world-rust-composed-http/go/interpreter/1/16 | 5.6834 | 0.0619 | 0.0004 | 26.9703 | 0.0430 | 0.2288 | 0.0005 | N/A | 0.0120 | 18.8885 | 17.2309 | 0.0338 | 209.9227 |
| head/world-rust-composed-http/go/interpreter/1/65536 | 5.4555 | 0.0658 | 0.0004 | 27.0230 | 0.0385 | 0.2354 | 0.0005 | N/A | 0.0125 | 25.2357 | 22.2235 | 0.0343 | 262.5875 |
| head/world-rust-composed-http/go/interpreter/4/16 | 5.5058 | 0.0682 | 0.0003 | 27.3255 | 0.0590 | 0.7287 | 0.0007 | N/A | 0.0165 | 6.6891 | 4.5288 | 0.0587 | 83.2775 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 5.6773 | 0.0663 | 0.0004 | 26.5158 | 0.0606 | 0.6957 | 0.0007 | N/A | 0.0180 | 8.7214 | 6.3572 | 0.0455 | 100.5998 |
| head/world-rust-composed-http/go/native/1/16 | 5.8834 | 0.0658 | 0.0004 | 222.3874 | 0.0397 | 0.2174 | 0.0005 | N/A | 0.0130 | 13.9445 | 13.2375 | 0.0396 | 364.6888 |
| head/world-rust-composed-http/go/native/1/65536 | 5.9528 | 0.0641 | 0.0004 | 223.8091 | 0.0413 | 0.2131 | 0.0005 | N/A | 0.0126 | 16.2832 | 15.2265 | 0.0398 | 387.7125 |
| head/world-rust-composed-http/go/native/4/16 | 5.8098 | 0.0695 | 0.0004 | 223.3478 | 0.0618 | 0.6725 | 0.0007 | N/A | 0.0153 | 4.3430 | 3.6130 | 0.0725 | 269.1715 |
| head/world-rust-composed-http/go/native/4/65536 | 5.7880 | 0.0677 | 0.0003 | 223.9973 | 0.0576 | 0.6605 | 0.0007 | N/A | 0.0132 | 4.7013 | 4.6607 | 0.0670 | 281.5133 |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 9.2383 | 0.0829 | 0.0003 | 43.9000 | 0.1135 | 0.5243 | 0.0005 | N/A | 0.2825 | 20.6970 | 19.8773 | 0.2480 | 258.1571 |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 9.2658 | 0.0772 | 0.0003 | 43.2754 | 0.1087 | 0.4402 | 0.0004 | N/A | 0.2830 | 30.8496 | 24.9748 | 0.2557 | 316.6639 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 9.3308 | 0.0779 | 0.0003 | 43.4700 | 0.2634 | 1.6803 | 0.0012 | N/A | 0.3060 | 11.4238 | 10.9023 | 0.6549 | 169.9025 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 9.4772 | 0.0666 | 0.0003 | 43.3237 | 0.2430 | 1.6135 | 0.0012 | N/A | 0.3060 | 18.3889 | 14.6949 | 0.6095 | 208.0678 |
| head/world-rust-composed-http/llgo/native/1/16 | 9.1614 | 0.0855 | 0.0002 | 330.1040 | 0.1058 | 0.2869 | 0.0005 | N/A | 0.2850 | 16.4465 | 16.0778 | 0.2577 | 503.8812 |
| head/world-rust-composed-http/llgo/native/1/65536 | 9.1852 | 0.0733 | 0.0003 | 327.4733 | 0.1078 | 0.2879 | 0.0005 | N/A | 0.2828 | 22.4933 | 19.6342 | 0.2417 | 539.9918 |
| head/world-rust-composed-http/llgo/native/4/16 | 9.1768 | 0.0874 | 0.0003 | 327.9697 | 0.2483 | 0.9992 | 0.0011 | N/A | 0.3186 | 12.5463 | 11.1983 | 0.6040 | 458.0441 |
| head/world-rust-composed-http/llgo/native/4/65536 | 9.0169 | 0.0851 | 0.0003 | 329.4183 | 0.2422 | 1.0355 | 0.0010 | N/A | 0.3337 | 14.2282 | 14.8002 | 0.5998 | 491.6047 |
| head/world-rust-http/go/interpreter/1/16 | 5.8660 | 0.0667 | 0.0003 | 18.5285 | 0.0445 | 0.3580 | 0.0005 | N/A | 0.0197 | 16.5301 | 15.1052 | 0.0325 | 180.7097 |
| head/world-rust-http/go/interpreter/1/65536 | 5.5618 | 0.0532 | 0.0005 | 19.0811 | 0.0397 | 0.1623 | 0.0005 | N/A | 0.0150 | 23.0744 | 20.0025 | 0.0339 | 233.2383 |
| head/world-rust-http/go/interpreter/4/16 | 5.8280 | 0.0698 | 0.0005 | 18.4157 | 0.0555 | 0.5247 | 0.0008 | N/A | 0.0152 | 4.4108 | 3.8504 | 0.0535 | 67.2546 |
| head/world-rust-http/go/interpreter/4/65536 | 5.7285 | 0.0592 | 0.0003 | 18.5560 | 0.0568 | 0.4796 | 0.0008 | N/A | 0.0144 | 7.4188 | 5.4774 | 0.0483 | 85.0900 |
| head/world-rust-http/go/native/1/16 | 5.6970 | 0.0620 | 0.0004 | 154.5128 | 0.0365 | 0.1652 | 0.0004 | N/A | 0.0122 | 12.0729 | 11.2187 | 0.0460 | 276.2253 |
| head/world-rust-http/go/native/1/65536 | 5.8572 | 0.0649 | 0.0004 | 155.1030 | 0.0379 | 0.1661 | 0.0006 | N/A | 0.0129 | 14.7100 | 13.7146 | 0.0423 | 302.0110 |
| head/world-rust-http/go/native/4/16 | 5.7656 | 0.0633 | 0.0004 | 154.3366 | 0.0525 | 0.4560 | 0.0007 | N/A | 0.0143 | 3.4348 | 3.1516 | 0.0556 | 194.9693 |
| head/world-rust-http/go/native/4/65536 | 5.7494 | 0.0725 | 0.0004 | 154.5063 | 0.0576 | 0.4884 | 0.0007 | N/A | 0.0167 | 4.9390 | 4.2485 | 0.0651 | 208.0478 |
| head/world-rust-http/llgo/interpreter/1/16 | 9.0100 | 0.0736 | 0.0003 | 29.7602 | 0.1097 | 0.2983 | 0.0005 | N/A | 0.2827 | 21.4714 | 17.5152 | 0.2632 | 218.6584 |
| head/world-rust-http/llgo/interpreter/1/65536 | 9.1856 | 0.0730 | 0.0002 | 29.1981 | 0.1186 | 2.5862 | 0.0005 | N/A | 0.3073 | 24.2735 | 23.0300 | 0.2528 | 276.1079 |
| head/world-rust-http/llgo/interpreter/4/16 | 9.4810 | 0.0741 | 0.0002 | 30.3475 | 0.2757 | 3.3065 | 0.0013 | N/A | 0.3125 | 11.1355 | 10.2910 | 0.5935 | 149.3636 |
| head/world-rust-http/llgo/interpreter/4/65536 | 9.1552 | 0.0724 | 0.0003 | 29.8164 | 0.2752 | 3.1461 | 0.0011 | N/A | 0.2963 | 15.3050 | 13.9042 | 0.6448 | 191.8862 |
| head/world-rust-http/llgo/native/1/16 | 9.4341 | 0.0765 | 0.0003 | 227.7191 | 0.0977 | 0.2129 | 0.0005 | N/A | 0.2769 | 14.4320 | 14.1228 | 0.2481 | 384.5475 |
| head/world-rust-http/llgo/native/1/65536 | 9.0956 | 0.0650 | 0.0003 | 226.9205 | 0.1233 | 0.2328 | 0.0005 | N/A | 0.2788 | 17.8098 | 18.0565 | 0.2387 | 422.0970 |
| head/world-rust-http/llgo/native/4/16 | 9.1525 | 0.0750 | 0.0003 | 227.9317 | 0.2313 | 0.6947 | 0.0009 | N/A | 0.2989 | 11.2205 | 10.4224 | 0.5670 | 349.1082 |
| head/world-rust-http/llgo/native/4/65536 | 9.2242 | 0.0613 | 0.0003 | 229.3083 | 0.2403 | 0.8366 | 0.0011 | N/A | 0.3291 | 13.8063 | 14.4998 | 0.5738 | 389.6157 |
| host/native-go/go/host/1/16 | 5.4494 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0174 | 0.0870 | 0.0721 | 0.0054 | 7.2228 |
| host/native-go/go/host/1/65536 | 5.3049 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0150 | 1.8674 | 1.6088 | 0.0080 | 23.7698 |
| host/native-go/go/host/4/16 | 5.2419 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0151 | 0.0600 | 0.0463 | 0.0058 | 6.6521 |
| host/native-go/go/host/4/65536 | 5.2925 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0155 | 1.3440 | 1.5444 | 0.0069 | 22.5061 |
| host/native-go/llgo/host/1/16 | 9.0723 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3079 | 0.5893 | 0.3878 | 0.0867 | 14.9070 |
| host/native-go/llgo/host/1/65536 | 8.9714 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2825 | 8.7540 | 7.8437 | 0.1004 | 88.5484 |
| host/native-go/llgo/host/4/16 | 9.0579 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3280 | 0.7200 | 0.5485 | 0.0978 | 16.8179 |
| host/native-go/llgo/host/4/65536 | 8.9157 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3216 | 9.5405 | 6.8341 | 0.0831 | 84.3059 |
| reference/world-rust-composed-http/go/compiler/1/16 | 6.1674 | 0.0706 | 0.1368 | 75.0534 | 0.1156 | 0.1344 | 0.0071 | 0.0297 | 0.0370 | 4.7001 | 3.9382 | 0.2388 | 123.8115 |
| reference/world-rust-composed-http/go/compiler/1/65536 | 6.1187 | 0.0715 | 0.1457 | 75.1498 | 0.1170 | 0.1385 | 0.0064 | 0.0301 | 0.0373 | 4.9579 | 4.4855 | 0.2525 | 129.2778 |
| reference/world-rust-composed-http/go/compiler/4/16 | 6.2636 | 0.0685 | 0.1368 | 75.2524 | 0.2218 | 0.2473 | 0.0079 | 0.0387 | 0.0524 | 1.5103 | 1.1439 | 0.3907 | 95.9366 |
| reference/world-rust-composed-http/go/compiler/4/65536 | 6.0095 | 0.0651 | 0.1346 | 75.9821 | 0.2330 | 0.2518 | 0.0082 | 0.0415 | 0.0548 | 1.7690 | 1.2714 | 0.4098 | 97.5393 |
| reference/world-rust-http/go/compiler/1/16 | 6.3895 | 0.0583 | 0.1497 | 52.5955 | 0.1204 | 0.1384 | 0.0076 | 0.0332 | 0.0402 | 3.8214 | 3.2879 | 0.1840 | 94.0338 |
| reference/world-rust-http/go/compiler/1/65536 | 6.1993 | 0.0611 | 0.1388 | 52.1790 | 0.1183 | 0.1186 | 0.0067 | 0.0330 | 0.0398 | 4.1813 | 3.7891 | 0.1674 | 99.6846 |
| reference/world-rust-http/go/compiler/4/16 | 6.3498 | 0.0672 | 0.1464 | 52.3251 | 0.2202 | 0.1919 | 0.0080 | 0.0415 | 0.0440 | 1.2923 | 0.9839 | 0.2476 | 70.7008 |
| reference/world-rust-http/go/compiler/4/65536 | 6.2376 | 0.0617 | 0.1397 | 51.2743 | 0.2195 | 0.1863 | 0.0078 | 0.0390 | 0.0526 | 1.4249 | 1.1007 | 0.2577 | 71.3045 |

## Frozen protocol and validation

Benchmark base `23b54f0b121a1bd2481d2d445186f40dbcea9a94` is production-equivalent to main `da00a76`; its additional benchmark is identical to head `a5dedf617a01c2d47e4653de83095a977e3889af`. Production LLGo stays `3d92a58`, `-O3 -lto=full`; Go 1.27.3, LLVM 22.1.8, Wasmtime 48.0.1. The base reuses prior exact binaries after checking every production/harness source hash; only the uncompiled benchmark test differs. No previous timing is reused. Reference compiler/program/SDK hashes remain pinned. Base/head measurement code is identical. Frozen reference differences are limited to owned-buffered flags/report metadata and cold options validation; warm driver, complete response checker and native/reference kernels are checked identical.

Five fixed randomized primary/A/A blocks, seed 2026111027; pointwise 95% paired log-ratio bootstrap, 8000 draws. Both Rust plain/composed HTTP fixtures, both owned backends/compilers, one/four workers, 16 clients/four requests, two warmups/seven samples, 16-byte/64-KiB payloads. Exactly 800 throughput processes, 80 separate resource processes, 160 qualifications, seven actual loader probes and 358400 checked warm responses. CRC32 independently checked. No source/profile/ISA/count search, favorable repeat, cross-event pooling, affinity/frequency isolation or family-wise claim.

Proportional local validation passes four affected Go and LLGo root packages (canonical/component/P2/P3), LLGo nested P3 measurement, two race packages and vet. Integration coverage of canonical/component exceeds 90% per package, minimum 92.33%; the initial unit-only canonical run measured 89.0%, so cross-package coverage was measured rather than adding tests that mirror the implementation. Every owned build/test/probe exited before timing. Actual pre-start process/cwd inspection records foreign workloads without interrupting them. Full final-head CI still requires all 32 root packages on both compilers/platforms, 31 coverage gates and all eight LLGo shards.

The corrected busy-Store admission test passes 40 race-mode repetitions; the remaining affected P3 race suite (excluding upstream world inventories, executed in normal semantics) and vet pass. Existing semantic tests cover live/nil/expired references, copied aliases, affine movement, ended borrows and concurrent movement. This change neither removes ownership checks nor alters cancellation, quotas, cleanup or guest-memory contracts.

CI timing review of PR #153 found maximum 608 seconds, with root compile/link stages dominating slow shards and nested P3 compilation dominating legacy on macOS. Moving nested execution alone would move the bottleneck rather than demonstrate a shorter maximum; this PR preserves the complete matrix and adds no CI job. Five-to-ten-minute jobs remain the target.

Broader [all-runtime service measurements](performance-service-all-runtimes-2026-11-10.md) retain wazero/Wasmer and other execution portfolios. They are not pooled with this P3 event. The [compiler study](performance-llgo-merged-2026-11-10.md) and LLGo GC issue #3045 remain separate; no compiler upgrade is included. Native/Core encoding, caller/codec allocation, interpreter, SIMD/memory and controlled performance gaps remain open.
