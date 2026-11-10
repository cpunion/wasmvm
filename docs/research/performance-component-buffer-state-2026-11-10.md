# Component reference and buffer state, 2026-11-10

[Complete machine record](data/performance-component-buffer-state-2026-11-10.json).

## Implementation

Component capability headers and their affine state now share one allocation. Copied headers retain an interior state pointer, preserving identity, mutexes, ownership invalidation and borrowed lifetimes. This concerns host resource-capability storage, not the guest WasmGC heap. No Caller or mutable capability state is reused.

Async buffer bounds/stride queries use the existing canonical layout builder directly. The shared selector retains Memory32/Memory64 and wide-resource lowering layouts, limits and caches; alignment, overflow and memory-capability checks remain in place. The query no longer constructs an unused codec, transfer or realloc closure.

Buffer codecs omit resource-binding state only for statically known primitive, string, enum, flags or unit elements. Aggregate and other element shapes retain the full existing resource path. Every codec still has fresh context/byte/value budgets, memory and realloc bindings. Load/store validation, logical reservations, cancellation, snapshots and joined cleanup remain enforced.

## Pure reference allocation diagnostic

The identical checked-in benchmarks perform 50000 creations or moves per source/compiler. Move inputs are prepared before the measured loop. Live/expired/ended-borrow/nil controls remain zero-allocation. Go and LLGo actually run the retained-alias forced host-GC test, ownership/concurrent-move tests and affected component scenarios. Raw benchmark ns/op are retained but ignored; these counters are not a timing ranking. LLGo Mallocs are unavailable and raw zero counters do not indicate zero allocation.

| Source | Compiler | Case | Runtime B/operation | Go mallocs/operation |
| --- | --- | --- | ---: | ---: |
| base | go | Inspection/live | 0 | 0 |
| base | go | Inspection/expired | 0 | 0 |
| base | go | Inspection/ended-borrow | 0 | 0 |
| base | go | Inspection/nil | 0 | 0 |
| base | go | Creation/create | 80 | 2 |
| base | go | Creation/move | 80 | 2 |
| base | llgo | Inspection/live | 0 | N/A |
| base | llgo | Inspection/expired | 0 | N/A |
| base | llgo | Inspection/ended-borrow | 0 | N/A |
| base | llgo | Inspection/nil | 0 | N/A |
| base | llgo | Creation/create | 111 | N/A |
| base | llgo | Creation/move | 111 | N/A |
| head | go | Inspection/live | 0 | 0 |
| head | go | Inspection/expired | 0 | 0 |
| head | go | Inspection/ended-borrow | 0 | 0 |
| head | go | Inspection/nil | 0 | 0 |
| head | go | Creation/create | 80 | 1 |
| head | go | Creation/move | 80 | 1 |
| head | llgo | Inspection/live | 0 | N/A |
| head | llgo | Inspection/expired | 0 | N/A |
| head | llgo | Inspection/ended-borrow | 0 | N/A |
| head | llgo | Inspection/nil | 0 | N/A |
| head | llgo | Creation/create | 96 | N/A |
| head | llgo | Creation/move | 95 | N/A |

## HTTP fixture, 16 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 17768 | 18367 | 85377 | 82112 |
| wasmvm go / native | 22650 | 22810 | 85520 | 82256 |
| wasmvm llgo / interpreter | 7019 | 7271 | 155374 | 150527 |
| wasmvm llgo / native | 6971 | 7083 | 155822 | 150977 |
| Wasmtime 48.0.1 | N/A | 74916 | N/A | 32 |
| Ordinary native-Go application / Go | N/A | 1844274 | N/A | 2713 |
| Ordinary native-Go application / LLGo | N/A | 121674 | N/A | 3417 |

## HTTP fixture, 65536 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 12198 | 12593 | 250593 | 241952 |
| wasmvm go / native | 16588 | 16378 | 250737 | 242097 |
| wasmvm llgo / interpreter | 5134 | 5318 | 408990 | 397420 |
| wasmvm llgo / native | 5105 | 5266 | 409436 | 397866 |
| Wasmtime 48.0.1 | N/A | 65554 | N/A | 32 |
| Ordinary native-Go application / Go | N/A | 43701 | N/A | 277915 |
| Ordinary native-Go application / LLGo | N/A | 9361 | N/A | 732501 |

**29/40 A/A controls resolve** within ±2.5%; 11/16 source pairs resolve both. Default-workstation timings are advisory; neither a controlled ranking nor release-performance acceptance is established. Every adverse, unchanged and unresolved result is retained below.

The plain HTTP Go native/64-KiB case has an adverse warm-time point estimate 1.0124 [0.9994,1.0260], with unresolved controls. Plain HTTP LLGo native/64-KiB observes 0.9658 [0.9428,0.9854], with both controls resolved. These are pointwise advisory intervals; they do not establish universal improvement, controlled parity or release acceptance.

Go managed bytes and LLGo collector accounting differ; Go counters exclude Rust/JIT allocation. Native-Go is an ordinary application control, not an optimized native lower bound. Resource instrumentation runs separately. Throughput medians and paired geometric warm-time ratios summarize different reductions.

## All source comparisons

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A | Base B/request | Head B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| head/world-rust-composed-http/go/interpreter/4/16 | 15530 | 15830 | 0.9785 [0.9708,0.9857] | yes | 99888 | 96624 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 11146 | 11360 | 0.9802 [0.9723,0.9879] | yes | 265105 | 256464 |
| head/world-rust-composed-http/go/native/4/16 | 18893 | 19705 | 0.9636 [0.9565,0.9684] | yes | 95761 | 92496 |
| head/world-rust-composed-http/go/native/4/65536 | 14424 | 14909 | 0.9698 [0.9593,0.9778] | yes | 260977 | 252337 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 6538 | 6698 | 0.9677 [0.9479,0.9828] | no | 174521 | 169628 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4968 | 5078 | 0.9707 [0.9600,0.9816] | yes | 428138 | 416518 |
| head/world-rust-composed-http/llgo/native/4/16 | 6555 | 6637 | 0.9901 [0.9775,1.0021] | yes | 170235 | 165338 |
| head/world-rust-composed-http/llgo/native/4/65536 | 5023 | 5075 | 0.9947 [0.9748,1.0210] | no | 423843 | 412232 |
| head/world-rust-http/go/interpreter/4/16 | 17768 | 18367 | 0.9734 [0.9685,0.9791] | yes | 85377 | 82112 |
| head/world-rust-http/go/interpreter/4/65536 | 12198 | 12593 | 0.9685 [0.9602,0.9797] | yes | 250593 | 241952 |
| head/world-rust-http/go/native/4/16 | 22650 | 22810 | 0.9982 [0.9684,1.0318] | no | 85520 | 82256 |
| head/world-rust-http/go/native/4/65536 | 16588 | 16378 | 1.0124 [0.9994,1.0260] | no | 250737 | 242097 |
| head/world-rust-http/llgo/interpreter/4/16 | 7019 | 7271 | 0.9692 [0.9657,0.9727] | yes | 155374 | 150527 |
| head/world-rust-http/llgo/interpreter/4/65536 | 5134 | 5318 | 0.9566 [0.9346,0.9710] | no | 408990 | 397420 |
| head/world-rust-http/llgo/native/4/16 | 6971 | 7083 | 0.9888 [0.9719,1.0049] | yes | 155822 | 150977 |
| head/world-rust-http/llgo/native/4/65536 | 5105 | 5266 | 0.9658 [0.9428,0.9854] | yes | 409436 | 397866 |

## Host compiler comparisons

Every head LLGo/Go pair uses the same fixture/backend/counts. Ratios above one mean more time under LLGo. API intervals include the complete materialized response and joined call; driver request creation and full response verification remain outside that API interval. No cross-compiler pooling or causal compiler-defect claim.

| LLGo profile | LLGo/Go warm time [95%] | LLGo/Go API time [95%] | Both A/A |
| --- | --- | --- | --- |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 2.3659 [2.3457,2.3873] | 2.3587 [2.3371,2.3809] | yes |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 2.2260 [2.2068,2.2485] | 2.2077 [2.1900,2.2282] | yes |
| head/world-rust-composed-http/llgo/native/4/16 | 2.9696 [2.9378,3.0055] | 2.9635 [2.9327,2.9983] | yes |
| head/world-rust-composed-http/llgo/native/4/65536 | 2.9591 [2.9310,2.9930] | 2.9432 [2.9133,2.9790] | no |
| head/world-rust-http/llgo/interpreter/4/16 | 2.5410 [2.5165,2.5689] | 2.5319 [2.5064,2.5623] | yes |
| head/world-rust-http/llgo/interpreter/4/65536 | 2.3647 [2.3361,2.3990] | 2.3456 [2.3176,2.3805] | yes |
| head/world-rust-http/llgo/native/4/16 | 3.2209 [3.1350,3.2940] | 3.2121 [3.1256,3.2850] | no |
| head/world-rust-http/llgo/native/4/65536 | 3.0944 [3.0238,3.1569] | 3.0807 [3.0085,3.1435] | no |

## Wasmtime and ordinary native application comparisons

Every owned head profile is compared with Wasmtime and the ordinary application compiled with the same host compiler. Ratios above one mean more owned-runtime warm time. The ordinary application omits Wasm/component/WASI semantics and is not an optimized native lower bound.

| Owned head profile | Reference | Head/reference warm time [95%] | Both A/A |
| --- | --- | --- | --- |
| head/world-rust-composed-http/go/interpreter/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 3.9729 [3.9446,4.0013] | yes |
| head/world-rust-composed-http/go/interpreter/4/16 | host/native-go/go/host/4/16 | 115.9014 [111.9148,119.5012] | no |
| head/world-rust-composed-http/go/interpreter/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 4.9405 [4.8136,5.0406] | no |
| head/world-rust-composed-http/go/interpreter/4/65536 | host/native-go/go/host/4/65536 | 3.7586 [3.5732,3.9447] | no |
| head/world-rust-composed-http/go/native/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 3.1997 [3.1678,3.2319] | yes |
| head/world-rust-composed-http/go/native/4/16 | host/native-go/go/host/4/16 | 93.3452 [88.9831,96.5965] | no |
| head/world-rust-composed-http/go/native/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 3.7591 [3.6349,3.8593] | no |
| head/world-rust-composed-http/go/native/4/65536 | host/native-go/go/host/4/65536 | 2.8598 [2.7185,2.9900] | no |
| head/world-rust-composed-http/llgo/interpreter/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 9.3993 [9.3564,9.4471] | yes |
| head/world-rust-composed-http/llgo/interpreter/4/16 | host/native-go/llgo/host/4/16 | 17.9585 [17.5393,18.3120] | no |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 10.9978 [10.6844,11.2257] | no |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | host/native-go/llgo/host/4/65536 | 1.8350 [1.8092,1.8657] | no |
| head/world-rust-composed-http/llgo/native/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 9.5018 [9.4378,9.5625] | yes |
| head/world-rust-composed-http/llgo/native/4/16 | host/native-go/llgo/host/4/16 | 18.1545 [17.7371,18.4314] | no |
| head/world-rust-composed-http/llgo/native/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 11.1234 [10.8168,11.4250] | no |
| head/world-rust-composed-http/llgo/native/4/65536 | host/native-go/llgo/host/4/65536 | 1.8560 [1.8122,1.9101] | no |
| head/world-rust-http/go/interpreter/4/16 | reference/world-rust-http/go/compiler/4/16 | 4.0814 [4.0670,4.0968] | yes |
| head/world-rust-http/go/interpreter/4/16 | host/native-go/go/host/4/16 | 99.9667 [95.9006,103.5250] | no |
| head/world-rust-http/go/interpreter/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 5.2166 [5.1771,5.2565] | yes |
| head/world-rust-http/go/interpreter/4/65536 | host/native-go/go/host/4/65536 | 3.3954 [3.2137,3.5764] | no |
| head/world-rust-http/go/native/4/16 | reference/world-rust-http/go/compiler/4/16 | 3.3126 [3.2555,3.3707] | no |
| head/world-rust-http/go/native/4/16 | host/native-go/go/host/4/16 | 81.1353 [79.1560,83.1641] | no |
| head/world-rust-http/go/native/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 4.0098 [3.9708,4.0492] | no |
| head/world-rust-http/go/native/4/65536 | host/native-go/go/host/4/65536 | 2.6099 [2.4963,2.7287] | no |
| head/world-rust-http/llgo/interpreter/4/16 | reference/world-rust-http/go/compiler/4/16 | 10.3709 [10.2651,10.5168] | yes |
| head/world-rust-http/llgo/interpreter/4/16 | host/native-go/llgo/host/4/16 | 16.6361 [16.1605,17.0226] | no |
| head/world-rust-http/llgo/interpreter/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 12.3357 [12.1748,12.4824] | yes |
| head/world-rust-http/llgo/interpreter/4/65536 | host/native-go/llgo/host/4/65536 | 1.7610 [1.7427,1.7795] | no |
| head/world-rust-http/llgo/native/4/16 | reference/world-rust-http/go/compiler/4/16 | 10.6694 [10.5232,10.8654] | yes |
| head/world-rust-http/llgo/native/4/16 | host/native-go/llgo/host/4/16 | 17.1149 [16.5626,17.5829] | no |
| head/world-rust-http/llgo/native/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 12.4081 [12.1923,12.5885] | yes |
| head/world-rust-http/llgo/native/4/65536 | host/native-go/llgo/host/4/65536 | 1.7713 [1.7430,1.8005] | no |

## Every profile and separate resource cohort

| Profile | req/s | API us | p99 request us | Runtime B/request | Go mallocs/request | Between-batch threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/4/16 | 15530 | 255.847 | 2008.875 | 99888 | 1394.0 | 7 | 134.844 | yes |
| base/world-rust-composed-http/go/interpreter/4/65536 | 11146 | 355.770 | 3280.250 | 265105 | 1802.0 | 8 | 147.094 | yes |
| base/world-rust-composed-http/go/native/4/16 | 18893 | 210.019 | 1821.167 | 95761 | 1400.0 | 8 | 158.312 | yes |
| base/world-rust-composed-http/go/native/4/65536 | 14424 | 273.766 | 2977.458 | 260977 | 1808.0 | 7 | 165.922 | yes |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 6538 | 605.949 | 5974.333 | 174521 | N/A | 57 | 165.578 | no |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 4968 | 792.642 | 6761.459 | 428138 | N/A | 57 | 181.891 | yes |
| base/world-rust-composed-http/llgo/native/4/16 | 6555 | 604.053 | 5805.292 | 170235 | N/A | 57 | 212.516 | yes |
| base/world-rust-composed-http/llgo/native/4/65536 | 5023 | 781.865 | 6310.167 | 423843 | N/A | 57 | 207.922 | no |
| base/world-rust-http/go/interpreter/4/16 | 17768 | 223.592 | 1813.292 | 85377 | 1267.0 | 7 | 91.547 | yes |
| base/world-rust-http/go/interpreter/4/65536 | 12198 | 325.072 | 2656.500 | 250593 | 1675.0 | 7 | 104.547 | yes |
| base/world-rust-http/go/native/4/16 | 22650 | 175.182 | 1572.292 | 85520 | 1270.0 | 7 | 119.375 | yes |
| base/world-rust-http/go/native/4/65536 | 16588 | 237.862 | 2424.250 | 250737 | 1678.0 | 8 | 119.234 | yes |
| base/world-rust-http/llgo/interpreter/4/16 | 7019 | 562.631 | 5455.292 | 155374 | N/A | 57 | 139.219 | yes |
| base/world-rust-http/llgo/interpreter/4/65536 | 5134 | 764.168 | 6260.000 | 408990 | N/A | 57 | 141.031 | no |
| base/world-rust-http/llgo/native/4/16 | 6971 | 567.663 | 5317.583 | 155822 | N/A | 57 | 176.000 | yes |
| base/world-rust-http/llgo/native/4/65536 | 5105 | 770.377 | 6133.875 | 409436 | N/A | 57 | 175.562 | yes |
| head/world-rust-composed-http/go/interpreter/4/16 | 15830 | 251.028 | 1984.000 | 96624 | 1328.0 | 7 | 133.297 | yes |
| head/world-rust-composed-http/go/interpreter/4/65536 | 11360 | 348.446 | 3183.916 | 256464 | 1712.0 | 7 | 145.812 | yes |
| head/world-rust-composed-http/go/native/4/16 | 19705 | 201.424 | 1743.750 | 92496 | 1334.0 | 7 | 164.031 | yes |
| head/world-rust-composed-http/go/native/4/65536 | 14909 | 264.959 | 2883.709 | 252337 | 1718.0 | 7 | 158.094 | yes |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 6698 | 590.825 | 6009.916 | 169628 | N/A | 57 | 168.984 | yes |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 5078 | 774.764 | 6669.250 | 416518 | N/A | 57 | 181.688 | yes |
| head/world-rust-composed-http/llgo/native/4/16 | 6637 | 596.187 | 6027.167 | 165338 | N/A | 57 | 212.578 | yes |
| head/world-rust-composed-http/llgo/native/4/65536 | 5075 | 773.691 | 6561.458 | 412232 | N/A | 57 | 194.328 | no |
| head/world-rust-http/go/interpreter/4/16 | 18367 | 215.918 | 1701.958 | 82112 | 1204.0 | 8 | 97.547 | yes |
| head/world-rust-http/go/interpreter/4/65536 | 12593 | 314.721 | 2618.375 | 241952 | 1588.0 | 7 | 106.391 | yes |
| head/world-rust-http/go/native/4/16 | 22810 | 173.927 | 1611.416 | 82256 | 1207.0 | 7 | 124.656 | no |
| head/world-rust-http/go/native/4/65536 | 16378 | 240.941 | 2449.792 | 242097 | 1591.0 | 7 | 121.812 | no |
| head/world-rust-http/llgo/interpreter/4/16 | 7271 | 544.190 | 5308.625 | 150527 | N/A | 57 | 143.750 | yes |
| head/world-rust-http/llgo/interpreter/4/65536 | 5318 | 738.569 | 5980.833 | 397420 | N/A | 57 | 141.125 | yes |
| head/world-rust-http/llgo/native/4/16 | 7083 | 558.262 | 5311.916 | 150977 | N/A | 57 | 176.328 | yes |
| head/world-rust-http/llgo/native/4/65536 | 5266 | 746.207 | 6268.417 | 397866 | N/A | 57 | 175.672 | yes |
| host/native-go/go/host/4/16 | 1844274 | 0.612 | 168.250 | 2713 | 18.0 | 5 | 15.719 | no |
| host/native-go/go/host/4/65536 | 43701 | 80.769 | 684.500 | 277915 | 48.0 | 7 | 17.156 | no |
| host/native-go/llgo/host/4/16 | 121674 | 4.043 | 324.083 | 3417 | N/A | 40 | 19.797 | no |
| host/native-go/llgo/host/4/65536 | 9361 | 400.898 | 3667.250 | 732501 | N/A | 40 | 69.188 | no |
| reference/world-rust-composed-http/go/compiler/4/16 | 62978 | 62.560 | 4103.084 | 32 | 1.0 | 7 | 32.781 | yes |
| reference/world-rust-composed-http/go/compiler/4/65536 | 56850 | 68.033 | 4547.708 | 32 | 1.0 | 7 | 33.500 | no |
| reference/world-rust-http/go/compiler/4/16 | 74916 | 52.493 | 3447.333 | 32 | 1.0 | 7 | 29.969 | yes |
| reference/world-rust-http/go/compiler/4/65536 | 65554 | 58.652 | 3890.833 | 32 | 1.0 | 7 | 31.438 | yes |

## Separate lifecycle intervals

Median milliseconds across five primary process reductions. Startup/readiness, load, engine setup, compile/validate, providers, instantiate, lookup, executor/workers, first batch, warm execution and joined cleanup remain separate. Process wall includes protocol parking; it is not standalone CLI latency. Host binary compilation is outside runtime sampling. No overlapping phases or cohorts are subtracted.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/4/16 | 6.8748 | 0.0715 | 0.0003 | 26.5309 | 0.0555 | 0.6728 | 0.0006 | N/A | 0.0167 | 70.6570 | 65.9382 | 0.0539 | 702.1603 |
| base/world-rust-composed-http/go/interpreter/4/65536 | 6.4145 | 0.0611 | 0.0003 | 26.2769 | 0.0574 | 0.6902 | 0.0006 | N/A | 0.0153 | 96.3760 | 91.8741 | 0.0523 | 960.0668 |
| base/world-rust-composed-http/go/native/4/16 | 6.8331 | 0.0771 | 0.0003 | 219.8155 | 0.0525 | 0.6310 | 0.0007 | N/A | 0.0136 | 56.0855 | 54.2014 | 0.0663 | 777.9186 |
| base/world-rust-composed-http/go/native/4/65536 | 6.4777 | 0.0648 | 0.0003 | 222.4517 | 0.0594 | 0.6704 | 0.0006 | N/A | 0.0147 | 73.4952 | 70.9909 | 0.0669 | 948.2974 |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 10.0790 | 0.0854 | 0.0003 | 44.1635 | 0.2457 | 1.6507 | 0.0009 | N/A | 0.3088 | 157.6289 | 156.6250 | 0.5885 | 1627.8208 |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 10.0059 | 0.0795 | 0.0003 | 42.9171 | 0.2618 | 1.6843 | 0.0010 | N/A | 0.3116 | 210.1418 | 206.1196 | 0.5666 | 2122.5122 |
| base/world-rust-composed-http/llgo/native/4/16 | 9.8025 | 0.0851 | 0.0003 | 328.8123 | 0.2567 | 0.9934 | 0.0013 | N/A | 0.3066 | 159.7816 | 156.2056 | 0.5907 | 1906.3600 |
| base/world-rust-composed-http/llgo/native/4/65536 | 10.3304 | 0.0980 | 0.0003 | 329.1643 | 0.2378 | 0.9564 | 0.0009 | N/A | 0.3311 | 208.8226 | 203.8647 | 0.5922 | 2397.2955 |
| base/world-rust-http/go/interpreter/4/16 | 6.4800 | 0.0623 | 0.0004 | 18.5352 | 0.0708 | 1.1090 | 0.0009 | N/A | 0.0165 | 60.7434 | 57.6311 | 0.0475 | 607.7962 |
| base/world-rust-http/go/interpreter/4/65536 | 6.8493 | 0.0726 | 0.0004 | 18.5051 | 0.0603 | 0.5075 | 0.0009 | N/A | 0.0154 | 87.1016 | 83.9516 | 0.0494 | 873.2951 |
| base/world-rust-http/go/native/4/16 | 6.6647 | 0.0610 | 0.0004 | 150.8117 | 0.0537 | 0.4835 | 0.0006 | N/A | 0.0146 | 45.7866 | 45.2089 | 0.0578 | 615.6693 |
| base/world-rust-http/go/native/4/65536 | 6.8070 | 0.0580 | 0.0003 | 151.5763 | 0.0577 | 0.4995 | 0.0008 | N/A | 0.0136 | 62.6269 | 61.7331 | 0.0615 | 780.2442 |
| base/world-rust-http/llgo/interpreter/4/16 | 10.2600 | 0.0693 | 0.0003 | 30.0677 | 0.2827 | 3.2982 | 0.0016 | N/A | 0.3200 | 146.1494 | 145.8837 | 0.5782 | 1504.1064 |
| base/world-rust-http/llgo/interpreter/4/65536 | 10.4232 | 0.0779 | 0.0003 | 29.6729 | 0.2854 | 3.9797 | 0.0013 | N/A | 0.3230 | 198.3461 | 199.4650 | 0.5965 | 2052.0859 |
| base/world-rust-http/llgo/native/4/16 | 10.0815 | 0.0600 | 0.0003 | 229.0048 | 0.2422 | 0.7141 | 0.0009 | N/A | 0.3148 | 150.7947 | 146.8973 | 0.5865 | 1717.2987 |
| base/world-rust-http/llgo/native/4/65536 | 10.4777 | 0.0760 | 0.0003 | 227.5430 | 0.2327 | 0.7143 | 0.0010 | N/A | 0.3056 | 204.7688 | 200.5992 | 0.5935 | 2235.4305 |
| head/world-rust-composed-http/go/interpreter/4/16 | 6.9039 | 0.0756 | 0.0005 | 26.2147 | 0.0591 | 0.6734 | 0.0007 | N/A | 0.0143 | 69.5133 | 64.6860 | 0.0512 | 686.9205 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 6.8214 | 0.0668 | 0.0004 | 26.4033 | 0.0607 | 0.6888 | 0.0007 | N/A | 0.0170 | 94.3085 | 90.1443 | 0.0564 | 940.9807 |
| head/world-rust-composed-http/go/native/4/16 | 6.7343 | 0.0637 | 0.0003 | 222.7088 | 0.0543 | 0.6566 | 0.0007 | N/A | 0.0177 | 53.1261 | 51.9665 | 0.0706 | 756.1088 |
| head/world-rust-composed-http/go/native/4/65536 | 6.6810 | 0.0690 | 0.0004 | 222.1362 | 0.0562 | 0.6629 | 0.0006 | N/A | 0.0143 | 69.8367 | 68.6833 | 0.0630 | 919.2364 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 10.2535 | 0.0860 | 0.0002 | 45.2181 | 0.2519 | 1.6276 | 0.0010 | N/A | 0.3147 | 151.0525 | 152.8798 | 0.5929 | 1587.0515 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 10.1967 | 0.0727 | 0.0003 | 44.6164 | 0.2459 | 1.6729 | 0.0010 | N/A | 0.3098 | 197.0532 | 201.6687 | 0.5718 | 2072.9206 |
| head/world-rust-composed-http/llgo/native/4/16 | 10.7025 | 0.0873 | 0.0002 | 330.1225 | 0.2380 | 0.9814 | 0.0010 | N/A | 0.3272 | 152.2616 | 154.2826 | 0.6033 | 1884.5846 |
| head/world-rust-composed-http/llgo/native/4/65536 | 9.9958 | 0.0723 | 0.0002 | 327.5584 | 0.2463 | 1.0108 | 0.0010 | N/A | 0.3490 | 201.7342 | 201.7645 | 0.5968 | 2360.3587 |
| head/world-rust-http/go/interpreter/4/16 | 6.6471 | 0.0577 | 0.0003 | 18.3158 | 0.0605 | 0.5330 | 0.0007 | N/A | 0.0145 | 59.1640 | 55.7521 | 0.0511 | 590.4420 |
| head/world-rust-http/go/interpreter/4/65536 | 6.5972 | 0.0658 | 0.0003 | 18.2674 | 0.0661 | 0.5560 | 0.0008 | N/A | 0.0162 | 84.9967 | 81.3151 | 0.0486 | 845.6590 |
| head/world-rust-http/go/native/4/16 | 6.5839 | 0.0593 | 0.0004 | 152.8128 | 0.0552 | 0.4893 | 0.0008 | N/A | 0.0143 | 45.7683 | 44.8933 | 0.0603 | 611.2969 |
| head/world-rust-http/go/native/4/65536 | 6.7038 | 0.0539 | 0.0003 | 151.9091 | 0.0505 | 0.4667 | 0.0006 | N/A | 0.0177 | 64.0536 | 62.5239 | 0.0603 | 793.3016 |
| head/world-rust-http/llgo/interpreter/4/16 | 10.4681 | 0.0760 | 0.0003 | 30.9155 | 0.2722 | 3.7767 | 0.0015 | N/A | 0.3114 | 143.7510 | 140.8269 | 0.6258 | 1456.2247 |
| head/world-rust-http/llgo/interpreter/4/65536 | 9.7734 | 0.0770 | 0.0003 | 31.9415 | 0.2652 | 4.0087 | 0.0012 | N/A | 0.3313 | 191.9304 | 192.5493 | 0.5672 | 1973.6498 |
| head/world-rust-http/llgo/native/4/16 | 10.5242 | 0.0763 | 0.0002 | 227.4495 | 0.2685 | 0.7429 | 0.0010 | N/A | 0.3176 | 144.3327 | 144.5746 | 0.6000 | 1690.1137 |
| head/world-rust-http/llgo/native/4/65536 | 9.8217 | 0.0686 | 0.0002 | 230.6436 | 0.2488 | 0.7601 | 0.0010 | N/A | 0.3291 | 195.1722 | 194.4468 | 0.5954 | 2196.9693 |
| host/native-go/go/host/4/16 | 6.2267 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0153 | 0.4440 | 0.5552 | 0.0052 | 13.1218 |
| host/native-go/go/host/4/65536 | 6.5397 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0165 | 23.0411 | 23.4317 | 0.0062 | 241.9958 |
| host/native-go/llgo/host/4/16 | 10.1417 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3407 | 7.9793 | 8.4159 | 0.1055 | 95.0726 |
| host/native-go/llgo/host/4/65536 | 9.7182 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3148 | 113.6899 | 109.3907 | 0.0820 | 1105.9444 |
| reference/world-rust-composed-http/go/compiler/4/16 | 6.6907 | 0.0615 | 0.1346 | 73.8639 | 0.2296 | 0.2580 | 0.0082 | 0.0416 | 0.0597 | 17.4626 | 16.2598 | 0.3770 | 247.9602 |
| reference/world-rust-composed-http/go/compiler/4/65536 | 7.2918 | 0.0665 | 0.1346 | 72.7336 | 0.2203 | 0.2563 | 0.0080 | 0.0436 | 0.0558 | 19.4977 | 18.0124 | 0.3968 | 265.3310 |
| reference/world-rust-http/go/compiler/4/16 | 6.8368 | 0.0573 | 0.1313 | 50.9563 | 0.2210 | 0.1869 | 0.0077 | 0.0437 | 0.0610 | 14.7780 | 13.6686 | 0.2411 | 198.5393 |
| reference/world-rust-http/go/compiler/4/65536 | 7.4118 | 0.0562 | 0.1315 | 51.7190 | 0.2252 | 0.1882 | 0.0083 | 0.0415 | 0.0410 | 16.8718 | 15.6207 | 0.2728 | 219.4038 |

## Protocol and validation

Base `3c778582a8986269c23093ddd309e613a2a59d14` is production-equivalent to main `c0cd305`; head `8e51def6863580e0ee5c92181a567040393cc72b` contains all three implementation changes. The baseline's additional tests/benchmarks are identical to head and are not compiled into the service runner. Exact baseline production/harness hashes match the retained runner binary; no earlier observations are reused. Go 1.27.3, LLGo `3d92a58` with `-O3 -lto=full`, LLVM 22.1.8 and Wasmtime 48.0.1 stay pinned.

One fixed five-block primary/A/A event, seed2026111031, pointwise95% paired-block log-ratio bootstrap with8000 draws. All40 four-worker profiles cover both Rust plain/composed HTTP guests, both owned backends/host compilers,16-byte/64-KiB bodies, Wasmtime and ordinary Go/LLGo applications. Exactly400 throughput processes,40 separate resource processes,80 qualifications and7 actual loader probes;2867200 checked warm responses. Each batch uses16 clients and64 requests/client, with2 warmups and7 measured batches. Counts are identical for every participant.

The preceding shorter event had only20.71s of directly observed warm intervals across400 processes. Request counts rise4->64 before this event to lengthen steady sampling. Previous short-count and compiler-candidate observations stay separate. No favorable repeat, source/profile/count sweep, cross-event pooling or controlled/family-wise acceptance claim. One-worker final-source performance is not measured in this affected subset.

Local validation passes four Go and four pinned-LLGo packages (canonical/component/P2/P3), the LLGo P3 module, three Go race packages and vet. Integration coverage is >=90% for canonical and component, minimum 92.31%; full final-head CI still checks all32 root packages on both compilers/platforms,31 executable-package coverage gates and all8 LLGo shards. No CI jobs are added.

The new regression keeps only a copied reference header, forces host collections, verifies its owner, moves it, then verifies invalidation and survival after another collection. Existing tests cover quotas, malformed/partial/empty streams, Memory32/64, destination errors/panics, cancellation, retained output and resource transfers; aggregate/resource-containing elements keep their binding machinery.

Before timing, every owned build/test/probe exits and actual process/cwd inspection records external workloads without interrupting them. This is an advisory workstation event with no affinity/frequency/background isolation. No compiler/test/audit workload owned by this task overlaps primary timing. Lifecycle phases remain separate and are not subtracted to infer CPU costs.

The [broader service portfolio](performance-service-all-runtimes-2026-11-10.md) retains wazero, Wasmer and other runtime observations; those do not provide equivalent P3 support or fresh timings in this event. The [compiler study](performance-llgo-merged-2026-11-10.md) remains independent; production compiler adoption stays gated separately. Core scalar conversions, interpreter/native execution, SIMD/memory, controlled performance and production release remain open.
