# Caller-owned byte-stream destinations, 2026-11-10

[Complete machine record](data/performance-byte-destination-2026-11-10.json).

## Result and implementation

`ReadAsyncInto` now copies guest `stream<u8>` data directly into storage supplied by a `ByteDestination`. The callback receives the actual cumulative batch size, preserving earlier data when it grows the final buffer. Empty batches and EOF request no storage. Source bounds and logical copy limits are checked before invoking the destination; endpoint/Store identities, cancellation and joined cleanup remain enforced. Existing `ReadAsync` still returns immutable snapshots.

WASI P3 buffered HTTP uses this generic path to grow only its final response. Declared length and materialization limits are checked before growth; partial batches, trailers, cancellation and guest join keep their contracts. This removes an intermediate guest-memory string snapshot. External mutable-reader snapshots retain their existing ownership boundary.

Canonical Codec/Transfer/resource state is coallocated, with fresh budgets and bookkeeping per call. Optional read callbacks are coallocated with their host buffer, preserving the ordinary guest-buffer footprint. No persistent nominal cache, shared mutable output, unsafe linear-memory alias or competitor fallback is introduced.

The original operation context is also checked synchronously before returning success. A destination that cancels while a batch is copied can otherwise outrun the asynchronous joined-context notification. A regression test verifies cancellation and cleanup.

## HTTP fixture, 16 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 17122 | 16538 | 87468 | 87556 |
| wasmvm go / native | 21124 | 20939 | 87613 | 87700 |
| wasmvm llgo / interpreter | 6924 | 7341 | 158588 | 158166 |
| wasmvm llgo / native | 6946 | 6919 | 159080 | 158560 |
| Wasmtime 48.0.1 | N/A | 66371 | N/A | 45 |
| Ordinary native-Go application / Go | N/A | 1551746 | N/A | 2712 |
| Ordinary native-Go application / LLGo | N/A | 121863 | N/A | 3526 |

## HTTP fixture, 65536 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 11831 | 11970 | 318207 | 252775 |
| wasmvm go / native | 15303 | 16433 | 318351 | 252919 |
| wasmvm llgo / interpreter | 5168 | 5381 | 478330 | 411892 |
| wasmvm llgo / native | 4977 | 5202 | 478835 | 412185 |
| Wasmtime 48.0.1 | N/A | 58837 | N/A | 32 |
| Ordinary native-Go application / Go | N/A | 45516 | N/A | 277915 |
| Ordinary native-Go application / LLGo | N/A | 8955 | N/A | 732573 |

**22/80 A/A controls resolve** within ±2.5%; 5/32 source pairs resolve both. These default-workstation timings are advisory. They establish neither a controlled ranking nor production/release performance acceptance. Every source and reference comparison, including adverse or unchanged results, remains in the record.

Among source pairs where both controls resolve, the 16-byte composed Go interpreter with one worker is adverse: warm-time ratio 1.0176 [1.0072,1.0273]. The 64-KiB plain LLGo interpreter with one worker observes 0.9734 [0.9606,0.9894]. Other pairs remain mixed or unresolved. These observations do not establish full-portfolio parity.

Go managed bytes and LLGo collector accounting differ. LLGo Mallocs are unavailable in this pin; Go counters exclude Rust/JIT allocation. Native-Go is an ordinary application control, not an optimized native lower bound. Resource instrumentation is a separate cohort. Medians of throughput and paired geometric warm-time ratios summarize different reductions and need not move identically.

## All source comparisons

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A | Base B/request | Head B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| head/world-rust-composed-http/go/interpreter/1/16 | 3827 | 3780 | 1.0176 [1.0072,1.0273] | yes | 101976 | 102065 |
| head/world-rust-composed-http/go/interpreter/1/65536 | 2921 | 2901 | 1.0020 [0.9889,1.0139] | yes | 332718 | 267284 |
| head/world-rust-composed-http/go/interpreter/4/16 | 14815 | 14736 | 1.0066 [0.9867,1.0239] | yes | 101979 | 102067 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10787 | 10413 | 1.0361 [1.0178,1.0549] | no | 332717 | 267286 |
| head/world-rust-composed-http/go/native/1/16 | 4949 | 4935 | 0.9974 [0.9864,1.0063] | no | 97850 | 97936 |
| head/world-rust-composed-http/go/native/1/65536 | 4123 | 4209 | 0.9745 [0.9567,0.9884] | no | 328588 | 263153 |
| head/world-rust-composed-http/go/native/4/16 | 18729 | 18812 | 1.0613 [0.9881,1.2133] | no | 97852 | 97940 |
| head/world-rust-composed-http/go/native/4/65536 | 13923 | 14841 | 0.9573 [0.9237,0.9920] | no | 328590 | 263156 |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 3217 | 3266 | 0.9946 [0.9778,1.0130] | no | 177915 | 177389 |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 2495 | 2547 | 0.9822 [0.9661,0.9970] | no | 497650 | 431052 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 6851 | 6716 | 1.0174 [0.9908,1.0448] | no | 177831 | 177373 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4825 | 5030 | 0.9630 [0.9359,0.9907] | no | 497458 | 430921 |
| head/world-rust-composed-http/llgo/native/1/16 | 4041 | 3988 | 1.0034 [0.9876,1.0206] | no | 173684 | 173137 |
| head/world-rust-composed-http/llgo/native/1/65536 | 3164 | 3290 | 0.9047 [0.7779,0.9839] | no | 493388 | 426718 |
| head/world-rust-composed-http/llgo/native/4/16 | 6500 | 6534 | 0.9648 [0.9101,1.0144] | no | 173574 | 173017 |
| head/world-rust-composed-http/llgo/native/4/65536 | 4742 | 4939 | 0.9906 [0.9497,1.0341] | no | 493190 | 426645 |
| head/world-rust-http/go/interpreter/1/16 | 4382 | 4264 | 1.0332 [0.9994,1.0682] | no | 87466 | 87557 |
| head/world-rust-http/go/interpreter/1/65536 | 3165 | 3265 | 0.9550 [0.9168,0.9842] | no | 318209 | 252772 |
| head/world-rust-http/go/interpreter/4/16 | 17122 | 16538 | 1.0299 [0.9985,1.0558] | no | 87468 | 87556 |
| head/world-rust-http/go/interpreter/4/65536 | 11831 | 11970 | 0.9935 [0.9627,1.0202] | no | 318207 | 252775 |
| head/world-rust-http/go/native/1/16 | 5794 | 5829 | 0.9948 [0.9820,1.0053] | yes | 87610 | 87701 |
| head/world-rust-http/go/native/1/65536 | 4642 | 4725 | 0.9951 [0.9803,1.0158] | no | 318351 | 252928 |
| head/world-rust-http/go/native/4/16 | 21124 | 20939 | 1.0192 [0.9986,1.0435] | no | 87613 | 87700 |
| head/world-rust-http/go/native/4/65536 | 15303 | 16433 | 0.9122 [0.8757,0.9389] | no | 318351 | 252919 |
| head/world-rust-http/llgo/interpreter/1/16 | 3667 | 3698 | 0.9859 [0.9546,1.0183] | no | 158730 | 158232 |
| head/world-rust-http/llgo/interpreter/1/65536 | 2742 | 2812 | 0.9734 [0.9606,0.9894] | yes | 478470 | 411916 |
| head/world-rust-http/llgo/interpreter/4/16 | 6924 | 7341 | 0.9809 [0.9487,1.0142] | no | 158588 | 158166 |
| head/world-rust-http/llgo/interpreter/4/65536 | 5168 | 5381 | 0.9585 [0.9420,0.9752] | no | 478330 | 411892 |
| head/world-rust-http/llgo/native/1/16 | 4653 | 4567 | 1.0154 [0.9933,1.0379] | no | 159222 | 158716 |
| head/world-rust-http/llgo/native/1/65536 | 3441 | 3650 | 0.9573 [0.9389,0.9779] | no | 478905 | 412320 |
| head/world-rust-http/llgo/native/4/16 | 6946 | 6919 | 1.0216 [0.9900,1.0634] | no | 159080 | 158560 |
| head/world-rust-http/llgo/native/4/65536 | 4977 | 5202 | 0.9428 [0.9240,0.9616] | no | 478835 | 412185 |

## Every profile and resource cohort

| Profile | req/s | API us | p99 request us | Runtime B/request | Go mallocs/request | Between-batch threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/1/16 | 3827 | 259.659 | 4797.959 | 101976 | 1543.0 | 7 | 88.234 | yes |
| base/world-rust-composed-http/go/interpreter/1/65536 | 2921 | 339.262 | 6416.834 | 332718 | 2023.1 | 7 | 100.281 | yes |
| base/world-rust-composed-http/go/interpreter/4/16 | 14815 | 264.089 | 1964.625 | 101979 | 1543.0 | 7 | 96.234 | yes |
| base/world-rust-composed-http/go/interpreter/4/65536 | 10787 | 356.483 | 3330.125 | 332717 | 2023.0 | 7 | 111.297 | no |
| base/world-rust-composed-http/go/native/1/16 | 4949 | 200.194 | 3806.541 | 97850 | 1549.1 | 7 | 145.062 | yes |
| base/world-rust-composed-http/go/native/1/65536 | 4123 | 239.116 | 4535.708 | 328588 | 2029.1 | 7 | 146.828 | no |
| base/world-rust-composed-http/go/native/4/16 | 18729 | 205.869 | 1855.250 | 97852 | 1549.0 | 7 | 149.547 | no |
| base/world-rust-composed-http/go/native/4/65536 | 13923 | 276.512 | 2907.208 | 328590 | 2029.0 | 7 | 160.359 | no |
| base/world-rust-composed-http/llgo/interpreter/1/16 | 3217 | 307.850 | 8445.084 | 177915 | N/A | 42 | 135.750 | yes |
| base/world-rust-composed-http/llgo/interpreter/1/65536 | 2495 | 393.988 | 9003.792 | 497650 | N/A | 42 | 150.750 | no |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 6851 | 572.052 | 5349.041 | 177831 | N/A | 57 | 151.453 | no |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 4825 | 813.049 | 6720.958 | 497458 | N/A | 57 | 168.922 | no |
| base/world-rust-composed-http/llgo/native/1/16 | 4041 | 244.786 | 6876.875 | 173684 | N/A | 42 | 183.062 | no |
| base/world-rust-composed-http/llgo/native/1/65536 | 3164 | 309.941 | 8205.333 | 493388 | N/A | 42 | 180.594 | no |
| base/world-rust-composed-http/llgo/native/4/16 | 6500 | 603.850 | 5540.041 | 173574 | N/A | 57 | 195.703 | no |
| base/world-rust-composed-http/llgo/native/4/65536 | 4742 | 819.413 | 5910.917 | 493190 | N/A | 57 | 197.422 | no |
| base/world-rust-http/go/interpreter/1/16 | 4382 | 226.363 | 4187.792 | 87466 | 1408.0 | 7 | 65.719 | yes |
| base/world-rust-http/go/interpreter/1/65536 | 3165 | 312.771 | 6482.500 | 318209 | 1888.1 | 7 | 77.109 | no |
| base/world-rust-http/go/interpreter/4/16 | 17122 | 226.192 | 1856.750 | 87468 | 1408.0 | 7 | 73.766 | no |
| base/world-rust-http/go/interpreter/4/65536 | 11831 | 324.860 | 2735.167 | 318207 | 1888.0 | 8 | 87.672 | no |
| base/world-rust-http/go/native/1/16 | 5794 | 170.906 | 3279.459 | 87610 | 1411.0 | 7 | 110.328 | yes |
| base/world-rust-http/go/native/1/65536 | 4642 | 212.092 | 4006.917 | 318351 | 1891.1 | 7 | 119.672 | yes |
| base/world-rust-http/go/native/4/16 | 21124 | 181.445 | 1552.458 | 87613 | 1411.0 | 7 | 119.406 | no |
| base/world-rust-http/go/native/4/65536 | 15303 | 249.990 | 2455.958 | 318351 | 1891.0 | 7 | 125.312 | no |
| base/world-rust-http/llgo/interpreter/1/16 | 3667 | 270.088 | 7856.791 | 158730 | N/A | 42 | 111.047 | no |
| base/world-rust-http/llgo/interpreter/1/65536 | 2742 | 358.486 | 8994.083 | 478470 | N/A | 42 | 140.438 | yes |
| base/world-rust-http/llgo/interpreter/4/16 | 6924 | 566.014 | 5008.042 | 158588 | N/A | 57 | 120.156 | yes |
| base/world-rust-http/llgo/interpreter/4/65536 | 5168 | 743.749 | 5871.000 | 478330 | N/A | 57 | 140.641 | no |
| base/world-rust-http/llgo/native/1/16 | 4653 | 211.987 | 5969.291 | 159222 | N/A | 42 | 168.375 | yes |
| base/world-rust-http/llgo/native/1/65536 | 3441 | 277.818 | 7591.875 | 478905 | N/A | 42 | 152.172 | no |
| base/world-rust-http/llgo/native/4/16 | 6946 | 563.965 | 5005.917 | 159080 | N/A | 57 | 175.281 | no |
| base/world-rust-http/llgo/native/4/65536 | 4977 | 789.456 | 6178.916 | 478835 | N/A | 57 | 192.281 | no |
| head/world-rust-composed-http/go/interpreter/1/16 | 3780 | 262.910 | 4903.917 | 102065 | 1445.0 | 7 | 87.266 | yes |
| head/world-rust-composed-http/go/interpreter/1/65536 | 2901 | 341.601 | 6354.334 | 267284 | 1853.1 | 7 | 96.969 | yes |
| head/world-rust-composed-http/go/interpreter/4/16 | 14736 | 263.368 | 2047.667 | 102067 | 1445.0 | 7 | 96.672 | yes |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10413 | 370.616 | 3782.291 | 267286 | 1853.0 | 7 | 108.719 | no |
| head/world-rust-composed-http/go/native/1/16 | 4935 | 200.867 | 3995.834 | 97936 | 1451.0 | 7 | 146.109 | no |
| head/world-rust-composed-http/go/native/1/65536 | 4209 | 234.240 | 4565.750 | 263153 | 1859.0 | 7 | 138.906 | no |
| head/world-rust-composed-http/go/native/4/16 | 18812 | 205.588 | 1802.084 | 97940 | 1451.0 | 8 | 160.047 | no |
| head/world-rust-composed-http/go/native/4/65536 | 14841 | 259.718 | 2732.667 | 263156 | 1859.0 | 7 | 152.672 | no |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 3266 | 303.508 | 7679.875 | 177389 | N/A | 42 | 135.375 | no |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 2547 | 386.016 | 9353.583 | 431052 | N/A | 42 | 147.906 | no |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 6716 | 583.876 | 5646.667 | 177373 | N/A | 57 | 151.031 | no |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 5030 | 774.841 | 6060.750 | 430921 | N/A | 57 | 159.031 | no |
| head/world-rust-composed-http/llgo/native/1/16 | 3988 | 247.947 | 7181.542 | 173137 | N/A | 42 | 199.188 | yes |
| head/world-rust-composed-http/llgo/native/1/65536 | 3290 | 297.021 | 7814.125 | 426718 | N/A | 42 | 184.672 | yes |
| head/world-rust-composed-http/llgo/native/4/16 | 6534 | 601.451 | 5429.541 | 173017 | N/A | 57 | 195.875 | no |
| head/world-rust-composed-http/llgo/native/4/65536 | 4939 | 795.425 | 6272.083 | 426645 | N/A | 57 | 193.875 | no |
| head/world-rust-http/go/interpreter/1/16 | 4264 | 232.506 | 4481.250 | 87557 | 1318.1 | 8 | 65.547 | no |
| head/world-rust-http/go/interpreter/1/65536 | 3265 | 303.366 | 5705.583 | 252772 | 1726.1 | 7 | 77.984 | yes |
| head/world-rust-http/go/interpreter/4/16 | 16538 | 235.734 | 1996.500 | 87556 | 1318.0 | 7 | 73.594 | no |
| head/world-rust-http/go/interpreter/4/65536 | 11970 | 321.369 | 2594.667 | 252775 | 1726.0 | 7 | 89.969 | no |
| head/world-rust-http/go/native/1/16 | 5829 | 169.850 | 3158.708 | 87701 | 1321.1 | 7 | 117.000 | yes |
| head/world-rust-http/go/native/1/65536 | 4725 | 208.336 | 4000.750 | 252928 | 1729.1 | 7 | 118.562 | no |
| head/world-rust-http/go/native/4/16 | 20939 | 183.336 | 1981.167 | 87700 | 1321.0 | 7 | 118.156 | no |
| head/world-rust-http/go/native/4/65536 | 16433 | 232.938 | 2332.125 | 252919 | 1729.0 | 7 | 125.781 | no |
| head/world-rust-http/llgo/interpreter/1/16 | 3698 | 267.267 | 7097.917 | 158232 | N/A | 42 | 110.953 | yes |
| head/world-rust-http/llgo/interpreter/1/65536 | 2812 | 342.755 | 8346.166 | 411916 | N/A | 42 | 123.969 | yes |
| head/world-rust-http/llgo/interpreter/4/16 | 7341 | 532.417 | 4939.250 | 158166 | N/A | 57 | 120.000 | no |
| head/world-rust-http/llgo/interpreter/4/65536 | 5381 | 728.574 | 5976.000 | 411892 | N/A | 57 | 140.719 | no |
| head/world-rust-http/llgo/native/1/16 | 4567 | 216.450 | 5991.292 | 158716 | N/A | 42 | 184.547 | no |
| head/world-rust-http/llgo/native/1/65536 | 3650 | 267.877 | 7359.375 | 412320 | N/A | 42 | 167.156 | no |
| head/world-rust-http/llgo/native/4/16 | 6919 | 566.783 | 5045.000 | 158560 | N/A | 57 | 175.188 | no |
| head/world-rust-http/llgo/native/4/65536 | 5202 | 742.471 | 5700.500 | 412185 | N/A | 57 | 191.953 | no |
| host/native-go/go/host/1/16 | 881165 | 0.505 | 32.833 | 2712 | 18.0 | 5 | 11.219 | no |
| host/native-go/go/host/1/65536 | 41080 | 20.583 | 520.833 | 277914 | 48.0 | 6 | 16.859 | no |
| host/native-go/go/host/4/16 | 1551746 | 0.747 | 36.750 | 2712 | 18.0 | 5 | 11.281 | no |
| host/native-go/go/host/4/65536 | 45516 | 75.509 | 544.458 | 277915 | 48.0 | 6 | 16.250 | no |
| host/native-go/llgo/host/1/16 | 184327 | 0.743 | 153.250 | 3515 | N/A | 37 | 12.422 | no |
| host/native-go/llgo/host/1/65536 | 8259 | 115.563 | 3163.291 | 732577 | N/A | 37 | 68.094 | no |
| host/native-go/llgo/host/4/16 | 121863 | 1.966 | 301.250 | 3526 | N/A | 40 | 12.625 | no |
| host/native-go/llgo/host/4/65536 | 8955 | 415.634 | 3710.000 | 732573 | N/A | 40 | 67.875 | no |
| reference/world-rust-composed-http/go/compiler/1/16 | 17072 | 57.579 | 1043.500 | 32 | 1.0 | 8 | 31.906 | no |
| reference/world-rust-composed-http/go/compiler/1/65536 | 15047 | 64.086 | 1188.417 | 32 | 1.0 | 7 | 32.109 | no |
| reference/world-rust-composed-http/go/compiler/4/16 | 57787 | 66.530 | 329.958 | 32 | 1.0 | 7 | 32.703 | no |
| reference/world-rust-composed-http/go/compiler/4/65536 | 51722 | 72.390 | 366.792 | 32 | 1.0 | 7 | 33.391 | yes |
| reference/world-rust-http/go/compiler/1/16 | 19849 | 49.326 | 909.083 | 32 | 1.0 | 7 | 29.609 | no |
| reference/world-rust-http/go/compiler/1/65536 | 17610 | 54.531 | 1015.500 | 32 | 1.0 | 8 | 29.922 | yes |
| reference/world-rust-http/go/compiler/4/16 | 66371 | 57.327 | 307.917 | 45 | 1.0 | 7 | 29.859 | no |
| reference/world-rust-http/go/compiler/4/65536 | 58837 | 63.641 | 348.791 | 32 | 1.0 | 7 | 30.312 | no |

## Separate lifecycle intervals

Median milliseconds across five primary process reductions. Ready, load, host-engine setup, guest compilation/validation, providers, instantiation, lookup, executor/workers, first batch, warm batch and joined cleanup remain separate. Process wall includes protocol parking and is not standalone CLI latency. Source-host build times are retained in build provenance, outside runtime sampling. No overlapping phases or cohorts are subtracted to infer costs.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1/16 | 5.4232 | 0.0562 | 0.0004 | 25.6485 | 0.0372 | 0.2381 | 0.0003 | N/A | 0.0138 | 18.1865 | 16.7227 | 0.0286 | 203.3219 |
| base/world-rust-composed-http/go/interpreter/1/65536 | 5.4401 | 0.0609 | 0.0004 | 26.4145 | 0.0332 | 0.2319 | 0.0003 | N/A | 0.0130 | 24.7440 | 21.9111 | 0.0304 | 257.1852 |
| base/world-rust-composed-http/go/interpreter/4/16 | 5.3013 | 0.0570 | 0.0003 | 25.8081 | 0.0508 | 0.6611 | 0.0005 | N/A | 0.0156 | 6.3643 | 4.3200 | 0.0510 | 79.7933 |
| base/world-rust-composed-http/go/interpreter/4/65536 | 5.4760 | 0.0581 | 0.0003 | 25.9792 | 0.0520 | 0.6806 | 0.0006 | N/A | 0.0139 | 8.4597 | 5.9331 | 0.0516 | 97.5940 |
| base/world-rust-composed-http/go/native/1/16 | 5.2528 | 0.0610 | 0.0003 | 215.6545 | 0.0327 | 0.2141 | 0.0004 | N/A | 0.0128 | 14.4662 | 12.9314 | 0.0340 | 356.7966 |
| base/world-rust-composed-http/go/native/1/65536 | 5.3997 | 0.0584 | 0.0003 | 216.7136 | 0.0347 | 0.2104 | 0.0003 | N/A | 0.0118 | 16.7216 | 15.5228 | 0.0364 | 381.9677 |
| base/world-rust-composed-http/go/native/4/16 | 5.4665 | 0.0557 | 0.0003 | 216.3597 | 0.0497 | 0.6584 | 0.0007 | N/A | 0.0135 | 3.9786 | 3.4172 | 0.0693 | 260.3958 |
| base/world-rust-composed-http/go/native/4/65536 | 5.2638 | 0.0567 | 0.0003 | 216.8022 | 0.0529 | 0.6523 | 0.0006 | N/A | 0.0136 | 4.5271 | 4.5969 | 0.0637 | 272.3379 |
| base/world-rust-composed-http/llgo/interpreter/1/16 | 8.6683 | 0.0627 | 0.0003 | 43.6254 | 0.1022 | 0.4572 | 0.0004 | N/A | 0.2772 | 20.4352 | 19.8921 | 0.2602 | 256.2815 |
| base/world-rust-composed-http/llgo/interpreter/1/65536 | 8.7997 | 0.0617 | 0.0002 | 41.8227 | 0.1039 | 0.4608 | 0.0005 | N/A | 0.2828 | 29.5746 | 25.6495 | 0.2176 | 314.6956 |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 8.7149 | 0.0671 | 0.0003 | 45.0962 | 0.2294 | 1.5563 | 0.0011 | N/A | 0.3311 | 9.8465 | 9.3412 | 0.6025 | 155.2012 |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 8.6252 | 0.0743 | 0.0003 | 41.6820 | 0.2389 | 1.5624 | 0.0011 | N/A | 0.3037 | 16.0913 | 13.2642 | 0.6118 | 190.4475 |
| base/world-rust-composed-http/llgo/native/1/16 | 9.0783 | 0.0706 | 0.0003 | 326.1581 | 0.1057 | 0.2851 | 0.0005 | N/A | 0.2823 | 18.5862 | 15.8382 | 0.2574 | 502.5238 |
| base/world-rust-composed-http/llgo/native/1/65536 | 8.8536 | 0.0658 | 0.0003 | 324.7551 | 0.0964 | 0.2812 | 0.0005 | N/A | 0.2860 | 19.6438 | 20.2275 | 0.2352 | 541.9833 |
| base/world-rust-composed-http/llgo/native/4/16 | 8.9886 | 0.0687 | 0.0003 | 324.0367 | 0.2239 | 0.9681 | 0.0009 | N/A | 0.3281 | 10.7373 | 9.8461 | 0.5902 | 437.6267 |
| base/world-rust-composed-http/llgo/native/4/65536 | 8.8387 | 0.0714 | 0.0003 | 325.0559 | 0.2456 | 0.9832 | 0.0010 | N/A | 0.3227 | 14.8523 | 13.4978 | 0.5732 | 474.5788 |
| base/world-rust-http/go/interpreter/1/16 | 5.5510 | 0.0580 | 0.0004 | 17.8306 | 0.0367 | 0.1923 | 0.0004 | N/A | 0.0132 | 15.6536 | 14.6041 | 0.0289 | 172.9070 |
| base/world-rust-http/go/interpreter/1/65536 | 5.3540 | 0.0512 | 0.0003 | 18.1859 | 0.0409 | 0.4013 | 0.0004 | N/A | 0.0144 | 22.4820 | 20.2195 | 0.0282 | 231.4691 |
| base/world-rust-http/go/interpreter/4/16 | 5.6813 | 0.0645 | 0.0004 | 17.9311 | 0.0624 | 0.7750 | 0.0008 | N/A | 0.0158 | 4.3711 | 3.7378 | 0.0443 | 65.4802 |
| base/world-rust-http/go/interpreter/4/65536 | 5.4020 | 0.0588 | 0.0004 | 17.9159 | 0.0524 | 0.4833 | 0.0006 | N/A | 0.0159 | 7.2500 | 5.4094 | 0.0467 | 82.9032 |
| base/world-rust-http/go/native/1/16 | 5.2774 | 0.0616 | 0.0005 | 149.3444 | 0.0315 | 0.1557 | 0.0003 | N/A | 0.0129 | 12.2112 | 11.0453 | 0.0431 | 269.1085 |
| base/world-rust-http/go/native/1/65536 | 5.7163 | 0.0546 | 0.0004 | 148.7060 | 0.0314 | 0.1637 | 0.0003 | N/A | 0.0125 | 15.2045 | 13.7860 | 0.0398 | 297.1748 |
| base/world-rust-http/go/native/4/16 | 5.4614 | 0.0537 | 0.0003 | 149.9698 | 0.0473 | 0.5190 | 0.0007 | N/A | 0.0167 | 3.2073 | 3.0298 | 0.0593 | 188.6169 |
| base/world-rust-http/go/native/4/65536 | 5.5119 | 0.0518 | 0.0003 | 149.9737 | 0.0478 | 0.4841 | 0.0006 | N/A | 0.0129 | 5.0060 | 4.1821 | 0.0543 | 201.2018 |
| base/world-rust-http/llgo/interpreter/1/16 | 8.9715 | 0.0650 | 0.0003 | 28.9577 | 0.1090 | 0.3045 | 0.0005 | N/A | 0.2790 | 22.6587 | 17.4508 | 0.2470 | 219.0942 |
| base/world-rust-http/llgo/interpreter/1/65536 | 8.8602 | 0.0609 | 0.0003 | 28.8180 | 0.1069 | 3.0876 | 0.0005 | N/A | 0.2943 | 24.4211 | 23.3445 | 0.2430 | 278.6547 |
| base/world-rust-http/llgo/interpreter/4/16 | 8.8108 | 0.0685 | 0.0003 | 29.8991 | 0.2384 | 3.5295 | 0.0013 | N/A | 0.3024 | 9.4265 | 9.2436 | 0.5445 | 135.2135 |
| base/world-rust-http/llgo/interpreter/4/65536 | 8.7751 | 0.0603 | 0.0003 | 30.4920 | 0.2566 | 3.3678 | 0.0013 | N/A | 0.3002 | 12.1163 | 12.3835 | 0.5560 | 171.5833 |
| base/world-rust-http/llgo/native/1/16 | 8.8317 | 0.0708 | 0.0003 | 228.1046 | 0.1039 | 0.2205 | 0.0005 | N/A | 0.2988 | 14.4554 | 13.7544 | 0.2612 | 382.9936 |
| base/world-rust-http/llgo/native/1/65536 | 8.8691 | 0.0701 | 0.0003 | 226.3251 | 0.0993 | 0.2160 | 0.0005 | N/A | 0.2940 | 20.8917 | 18.6012 | 0.2387 | 426.1450 |
| base/world-rust-http/llgo/native/4/16 | 8.6855 | 0.0638 | 0.0003 | 225.8712 | 0.2390 | 0.7278 | 0.0011 | N/A | 0.3125 | 9.5760 | 9.2138 | 0.5830 | 330.6458 |
| base/world-rust-http/llgo/native/4/65536 | 8.6457 | 0.0600 | 0.0003 | 225.5238 | 0.2254 | 0.7275 | 0.0010 | N/A | 0.3230 | 13.7935 | 12.8599 | 0.6057 | 368.9780 |
| head/world-rust-composed-http/go/interpreter/1/16 | 5.3097 | 0.0621 | 0.0003 | 25.7651 | 0.0398 | 0.2206 | 0.0004 | N/A | 0.0136 | 18.4575 | 16.9301 | 0.0291 | 206.1651 |
| head/world-rust-composed-http/go/interpreter/1/65536 | 5.2155 | 0.0530 | 0.0004 | 26.0046 | 0.0357 | 0.2239 | 0.0004 | N/A | 0.0120 | 24.4841 | 22.0589 | 0.0299 | 257.3789 |
| head/world-rust-composed-http/go/interpreter/4/16 | 5.4090 | 0.0617 | 0.0003 | 25.6720 | 0.0534 | 0.6833 | 0.0006 | N/A | 0.0171 | 5.3448 | 4.3431 | 0.0455 | 80.3627 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 5.2897 | 0.0628 | 0.0003 | 25.8537 | 0.0570 | 0.6874 | 0.0006 | N/A | 0.0159 | 8.3926 | 6.1461 | 0.0483 | 98.0744 |
| head/world-rust-composed-http/go/native/1/16 | 5.1633 | 0.0640 | 0.0003 | 215.5642 | 0.0374 | 0.2129 | 0.0003 | N/A | 0.0118 | 13.9144 | 12.9684 | 0.0445 | 354.2160 |
| head/world-rust-composed-http/go/native/1/65536 | 5.5273 | 0.0571 | 0.0004 | 215.3025 | 0.0365 | 0.2144 | 0.0004 | N/A | 0.0120 | 16.1477 | 15.2046 | 0.0375 | 375.7068 |
| head/world-rust-composed-http/go/native/4/16 | 5.4857 | 0.0587 | 0.0004 | 217.9330 | 0.0550 | 0.6425 | 0.0007 | N/A | 0.0133 | 4.0481 | 3.4020 | 0.0673 | 261.0808 |
| head/world-rust-composed-http/go/native/4/65536 | 5.4545 | 0.0590 | 0.0004 | 215.7481 | 0.0519 | 0.6441 | 0.0006 | N/A | 0.0140 | 4.4231 | 4.3124 | 0.0655 | 269.8264 |
| head/world-rust-composed-http/llgo/interpreter/1/16 | 8.7188 | 0.0663 | 0.0003 | 43.5759 | 0.1160 | 0.4668 | 0.0005 | N/A | 0.2816 | 20.4714 | 19.5964 | 0.2375 | 252.3182 |
| head/world-rust-composed-http/llgo/interpreter/1/65536 | 8.6277 | 0.0607 | 0.0002 | 41.2555 | 0.1093 | 0.4649 | 0.0005 | N/A | 0.2746 | 29.4764 | 25.1319 | 0.2835 | 308.5547 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 9.2366 | 0.0757 | 0.0003 | 43.1379 | 0.2509 | 1.5394 | 0.0013 | N/A | 0.3054 | 10.0312 | 9.5301 | 0.5844 | 156.6139 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 8.7387 | 0.0679 | 0.0003 | 40.8782 | 0.2363 | 1.6000 | 0.0011 | N/A | 0.3241 | 15.8208 | 12.7234 | 0.5755 | 186.4462 |
| head/world-rust-composed-http/llgo/native/1/16 | 8.7698 | 0.0696 | 0.0003 | 324.6974 | 0.1091 | 0.2900 | 0.0005 | N/A | 0.2980 | 16.8236 | 16.0483 | 0.2545 | 501.0855 |
| head/world-rust-composed-http/llgo/native/1/65536 | 8.6071 | 0.0647 | 0.0002 | 327.1948 | 0.1054 | 0.3002 | 0.0005 | N/A | 0.3060 | 22.1765 | 19.4521 | 0.2277 | 537.6722 |
| head/world-rust-composed-http/llgo/native/4/16 | 8.7113 | 0.0663 | 0.0003 | 327.7382 | 0.2398 | 0.9624 | 0.0011 | N/A | 0.2986 | 10.1132 | 9.7942 | 0.5937 | 437.8155 |
| head/world-rust-composed-http/llgo/native/4/65536 | 8.9054 | 0.0660 | 0.0002 | 325.0539 | 0.2331 | 0.9361 | 0.0008 | N/A | 0.3039 | 12.8129 | 12.9591 | 0.5667 | 469.9825 |
| head/world-rust-http/go/interpreter/1/16 | 5.4028 | 0.0543 | 0.0004 | 17.9930 | 0.0398 | 0.1677 | 0.0005 | N/A | 0.0142 | 16.0513 | 15.0104 | 0.0338 | 178.7837 |
| head/world-rust-http/go/interpreter/1/65536 | 5.3302 | 0.0493 | 0.0003 | 17.7438 | 0.0403 | 0.1646 | 0.0004 | N/A | 0.0131 | 22.5718 | 19.6020 | 0.0289 | 225.1199 |
| head/world-rust-http/go/interpreter/4/16 | 5.3654 | 0.0554 | 0.0003 | 17.8176 | 0.0562 | 0.9725 | 0.0010 | N/A | 0.0171 | 4.2873 | 3.8699 | 0.0432 | 65.2050 |
| head/world-rust-http/go/interpreter/4/65536 | 5.3342 | 0.0533 | 0.0004 | 17.9443 | 0.0632 | 1.0617 | 0.0009 | N/A | 0.0141 | 7.2111 | 5.3465 | 0.0400 | 82.1173 |
| head/world-rust-http/go/native/1/16 | 5.4508 | 0.0573 | 0.0004 | 148.5648 | 0.0307 | 0.1575 | 0.0004 | N/A | 0.0118 | 12.1850 | 10.9793 | 0.0505 | 267.5191 |
| head/world-rust-http/go/native/1/65536 | 5.3946 | 0.0577 | 0.0004 | 148.8434 | 0.0354 | 0.1637 | 0.0004 | N/A | 0.0134 | 14.4809 | 13.5449 | 0.0361 | 296.2700 |
| head/world-rust-http/go/native/4/16 | 5.4116 | 0.0564 | 0.0004 | 148.2757 | 0.0506 | 0.4787 | 0.0007 | N/A | 0.0131 | 3.1645 | 3.0564 | 0.0510 | 187.7862 |
| head/world-rust-http/go/native/4/65536 | 5.4177 | 0.0535 | 0.0003 | 148.6141 | 0.0487 | 0.4660 | 0.0005 | N/A | 0.0128 | 3.7458 | 3.8947 | 0.0545 | 196.8802 |
| head/world-rust-http/llgo/interpreter/1/16 | 8.7216 | 0.0573 | 0.0002 | 29.5998 | 0.1114 | 0.2856 | 0.0005 | N/A | 0.2825 | 21.4778 | 17.3046 | 0.2411 | 219.0640 |
| head/world-rust-http/llgo/interpreter/1/65536 | 8.6657 | 0.0601 | 0.0003 | 29.4398 | 0.1098 | 2.7035 | 0.0005 | N/A | 0.2927 | 23.3632 | 22.7611 | 0.2533 | 273.7069 |
| head/world-rust-http/llgo/interpreter/4/16 | 8.5985 | 0.0700 | 0.0003 | 28.9827 | 0.2716 | 3.9080 | 0.0011 | N/A | 0.3234 | 9.4379 | 8.7186 | 0.6136 | 134.3023 |
| head/world-rust-http/llgo/interpreter/4/65536 | 8.9285 | 0.0681 | 0.0003 | 29.0331 | 0.2662 | 3.3434 | 0.0011 | N/A | 0.2988 | 12.6441 | 11.8947 | 0.5851 | 168.4557 |
| head/world-rust-http/llgo/native/1/16 | 8.6861 | 0.0693 | 0.0003 | 225.3718 | 0.1072 | 0.2070 | 0.0005 | N/A | 0.2887 | 13.8399 | 14.0134 | 0.2593 | 379.8290 |
| head/world-rust-http/llgo/native/1/65536 | 8.6658 | 0.0573 | 0.0002 | 224.8723 | 0.0952 | 0.2095 | 0.0005 | N/A | 0.2678 | 19.6505 | 17.5328 | 0.2328 | 416.5992 |
| head/world-rust-http/llgo/native/4/16 | 8.9058 | 0.0610 | 0.0002 | 223.6370 | 0.2395 | 0.6940 | 0.0009 | N/A | 0.2998 | 9.3468 | 9.2501 | 0.5919 | 331.3492 |
| head/world-rust-http/llgo/native/4/65536 | 8.7914 | 0.0570 | 0.0003 | 222.7075 | 0.2707 | 3.1688 | 0.0011 | N/A | 0.3322 | 11.9543 | 12.3035 | 0.5777 | 362.2806 |
| host/native-go/go/host/1/16 | 4.8724 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0129 | 0.0880 | 0.0726 | 0.0048 | 6.4332 |
| host/native-go/go/host/1/65536 | 4.8365 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0127 | 1.8211 | 1.5579 | 0.0054 | 21.9193 |
| host/native-go/go/host/4/16 | 4.9616 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0145 | 0.0537 | 0.0412 | 0.0053 | 6.2291 |
| host/native-go/go/host/4/65536 | 4.9908 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0159 | 1.3646 | 1.4061 | 0.0053 | 20.3068 |
| host/native-go/llgo/host/1/16 | 8.2288 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2227 | 0.3390 | 0.3472 | 0.0682 | 13.0025 |
| host/native-go/llgo/host/1/65536 | 8.6855 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2890 | 10.2052 | 7.7492 | 0.0913 | 90.5110 |
| host/native-go/llgo/host/4/16 | 8.5760 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3090 | 0.6369 | 0.5252 | 0.0844 | 15.5565 |
| host/native-go/llgo/host/4/65536 | 8.3590 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3173 | 10.0288 | 7.1471 | 0.0932 | 84.5731 |
| reference/world-rust-composed-http/go/compiler/1/16 | 5.8977 | 0.0602 | 0.1381 | 72.0697 | 0.1113 | 0.1331 | 0.0064 | 0.0276 | 0.0354 | 4.2802 | 3.7487 | 0.2215 | 118.0829 |
| reference/world-rust-composed-http/go/compiler/1/65536 | 5.8156 | 0.0570 | 0.1301 | 73.3601 | 0.1098 | 0.1349 | 0.0063 | 0.0280 | 0.0339 | 4.7651 | 4.2532 | 0.2210 | 124.0636 |
| reference/world-rust-composed-http/go/compiler/4/16 | 5.7925 | 0.0576 | 0.1271 | 71.0888 | 0.2167 | 0.2512 | 0.0079 | 0.0384 | 0.0433 | 1.4113 | 1.1075 | 0.3245 | 90.8642 |
| reference/world-rust-composed-http/go/compiler/4/65536 | 5.7968 | 0.0608 | 0.1286 | 71.8257 | 0.2215 | 0.2417 | 0.0080 | 0.0428 | 0.0464 | 1.5354 | 1.2374 | 0.3358 | 92.5622 |
| reference/world-rust-http/go/compiler/1/16 | 5.9327 | 0.0574 | 0.1267 | 49.7658 | 0.1097 | 0.1143 | 0.0063 | 0.0262 | 0.0351 | 3.6503 | 3.2243 | 0.1520 | 90.3965 |
| reference/world-rust-http/go/compiler/1/65536 | 5.7226 | 0.0600 | 0.1327 | 49.7038 | 0.1087 | 0.1150 | 0.0062 | 0.0268 | 0.0338 | 4.0540 | 3.6342 | 0.1570 | 94.5048 |
| reference/world-rust-http/go/compiler/4/16 | 5.9094 | 0.0524 | 0.1258 | 49.6314 | 0.2156 | 0.1836 | 0.0077 | 0.0373 | 0.0417 | 1.1890 | 0.9643 | 0.2057 | 67.4749 |
| reference/world-rust-http/go/compiler/4/65536 | 5.6909 | 0.0551 | 0.1259 | 49.9370 | 0.2155 | 0.1777 | 0.0075 | 0.0362 | 0.0435 | 1.3251 | 1.0878 | 0.2322 | 68.9634 |

## Frozen protocol and semantic qualification

Base `95a70d0f3d0d9ddb663333613717b3c794965e38`; head `da65b7230a049932053af7b8ad1013a7144e5cb0`. The host is Apple M4 Max/macOS 27.0.1. Production LLGo stays `3d92a58`, `-O3 -lto=full`; Go 1.27.3 and LLVM 22.1.8. The base reuses the prior 32-pair event's exact binaries only after all archived source-file hashes match; the prior timed observations are not reused. Frozen Wasmtime/native controls keep their compiler/program/SDK hashes. Base/head measurement code is identical. The reference predates owned-buffered flags/report metadata and cold options validation, while its actual warm driver, response checker, HTTP/native kernels and reference-specific code match.

Five fixed randomized primary/A/A blocks, seed 2026111024, pointwise 95% paired log-ratio bootstrap with 8000 draws. Both Rust plain/composed HTTP fixtures, both owned backends/host compilers, one/four workers, 16 clients and four requests, two warmups/seven samples, and both 16-byte/64-KiB payloads. Exactly 800 throughput processes, 80 separate resource processes, 160 qualification processes, seven actual loader probes and 358400 checked warm responses. CRC32 values are independently verified. No source/profile/ISA/count sweep, favorable repeat, cross-event pooling, affinity/frequency isolation or family-wise claim.

Before timing, all 32 Go root packages, four affected LLGo packages (canonical/component/P2/P3), the LLGo P3 measurement module, four Go race packages and vet pass. All 31 executable packages meet 90% statement coverage; minimum 90.89%. Every owned build/test/probe process actually exited. The actual pre-start process table verifies no owned compiler/test workload; external workloads, if present, are recorded and never interrupted. Full final-head Go/LLGo Linux-amd64/macOS-arm64 CI remains required before merge.

Tests cover partial/empty/EOF reads, packed and element-wise host inputs, invalid type/Store/limits, destination failure/panic, cancellation including late completions, callback reentry, retained independent output, forced GC and guest Memory32/Memory64. Authored producer fixtures initially assumed peer-closed write status and a fixed batch size; they were corrected to open-stream status and partial-read semantics before final qualification. wasm-tools 1.258.0 validates the new Memory32 fixture and rejects the forward Memory64 stream ABI; its exact rejection and fixture hashes are pinned. Both owned backends actually execute both widths.

The earlier [32-pair comparison](performance-component-type-comparison32-2026-11-10.md), [complete compiler candidate study](performance-llgo-merged-2026-11-10.md), and other runtime portfolios remain independent events. LLGo compiler adoption remains gated by its GC finding. Native/compiler/memory/SIMD, host-runtime and controlled performance gaps remain open.
