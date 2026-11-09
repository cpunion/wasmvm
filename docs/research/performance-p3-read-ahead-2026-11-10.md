# P3 bounded read-ahead and readiness: 2026-11-10

[Complete machine record](data/performance-p3-read-ahead-2026-11-10.json).

## Identified mechanism

The unchanged wit-bindgen 0.54.0 guest collects its stream into a growing Vec. For the complete 64 KiB request it asks for small destination spans before larger ones. The former HTTP provider dispatches each read to a worker, even when the input could already be buffered. Separate instrumented builds execute 32 checked responses under each host compiler and role:

| Per complete request | Base Go | Head Go | Base LLGo | Head LLGo |
| --- | ---: | ---: | ---: | ---: |
| External input reads | 15 | 2 | 15 | 2 |
| Created component tasks | 24 | 24 | 24 | 24 |
| Component task resumes | 80 | 41 | 80 | 41 |
| Async host starts | 23 | 23 | 23 | 23 |
| Source reads | 17 | 17 | 17 | 17 |
| Sync host invokes, including resolved completions | 36 | 36 | 36 | 36 |
| External response-body reads | 14 | 14 | 14 | 14 |

Input read sizes previously range from 8 bytes to 64 KiB; the new reader performs two bounded 64 KiB calls, including EOF. The task/resume counts agree under Go and LLGo; these counts describe this fixed guest/payload/run, while scheduler wait/scan counts vary and are retained without being treated as deterministic costs. Counter instrumentation and reader wrappers are separate source archives, not the timed executables. This demonstrates fewer handoffs, not a complete attribution of every nanosecond.

The pinned [Wasmtime producer](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasi-http/src/p3/body.rs) polls body frames, fills the available destination, and buffers the remainder. Its ready frame can therefore satisfy later guest reads without a worker handoff. The pinned [guest stream collector](https://github.com/bytecodealliance/wit-bindgen/blob/v0.54.0/crates/guest-rust/src/rt/async_support/stream_support.rs) grows the destination Vec between reads. The owned provider now uses its existing Prompt source contract for ready buffered tails; actual blocking reads still run on bounded IO workers. No new scheduler, compiler patch or competitor fallback is involved.

## Implementation and contracts

A positive external body read obtains at most MaxReadBytes, snapshots it into one immutable batch, and returns only the guest-requested prefix. Later reads consume the cached tail immediately. A source retains at most one such batch; aggregate retention is bounded by MaxMessages and MaxReadBytes, in addition to the separately bounded IO scratch cache. Zero-length probes do not fetch input. The last cached fragment carries EOF only after the whole batch is consumed, including the data-plus-EOF case. A source close clears the batch; late completions cannot repopulate it. Guest copy budgets, value/range checks, caller expiry, timeout, content-length checks, trailers, cancellation and joined owner shutdown remain active. Shared-World/two-Store body isolation remains tested.

## Complete fixed event

M4 Max/macOS 27.0.1. Base `7a5c7a15801a2a71abd4aa4c8546dfd64ef3c42f`, head `bc1a531daee4d39130a1dcded47080de150277dc`; Go 1.27.3 and LLGo `3d92a58` with `-O3 -lto=full`. The base already has bounded workers and transfer buffers. Driver, checker, fixtures, references and compiler profiles are unchanged. The counter diagnostic base `162c27b` has identical production code to the measured base.

Twenty randomized blocks, primary/A/A aliases, both Rust HTTP and composed fixtures, Go/LLGo interpreter/native, one/four workers, 16 clients with four requests each, two warmups and seven measured batches, 64 KiB, independently checked CRC32 2629362236. Order/bootstrap seed 2026111008; 8000 paired log-ratio bootstrap draws, pointwise 95%. All 1600 throughput processes, 40 separate resource processes, 80 qualification processes and seven actual loader probes complete, checking 716800 warm responses. Every byte, status 201, header and trailer is checked. No source/profile/count sweep, favorable retiming, phase subtraction, family-wise claim or pooling with earlier events.

**All 40 A/A controls are unresolved.** Host load rises from [3.98,6.58,7.60] to [12.57,11.02,9.32]. All owned builds/tests/probes/profiles/archives exited, and the actual pre-start compiler/test process table is clear. After the event an unrelated compiler `cl.test` is observed with elapsed time consistent with overlap in the later sampling period. This is a post-event inference, not continuous monitoring. The complete noisy event is retained; it establishes no stable speed win, controlled ranking or release acceptance.

Observed four-worker LLGo native HTTP throughput is 1971 -> 2666 req/s, complete API 1993.3 -> 1468.4 us; paired warm-time ratio 0.7231 [0.6728,0.7644], with both controls unresolved. Accounted allocation falls 938625 -> 848551 B/request. Go native observes 7764 -> 11734 req/s. Same-event Wasmtime is 56524 req/s and bare LLGo host 8803 req/s. The full table retains all 16 comparisons; these observations do not close the remaining large gap. Between-batch LLGo native threads remain 61; observed peak RSS is 176.3 -> 175.9 MiB. Runtime byte counters differ between Go and LLGo, and Go counters exclude Rust/JIT allocation.

Readiness, preparation, first/warm batches, complete API, request latency and joined cleanup remain separate. Owned request construction is outside API time; the Go C/Rust reference constructs its request inside the bridge. Both collect and validate complete responses, but the adapter topology differs: the owned streaming API uses a Store worker and pipe, while the reference consumes the body/trailers inside Store::run_concurrent. That remaining bridge difference needs its own measured comparison. It must not be silently removed from the existing streaming baseline or attributed entirely to guest instruction generation.

## Every owned comparison

| Fixture / compiler / backend / workers | Base req/s | Head req/s | Warm head/base [95%] | Both A/A resolve | Base B/req | Head B/req |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| world-rust-composed-http/go/interpreter/1 | 1582 | 2470 | 0.6522 [0.6126,0.7149] | no | 493507 | 456328 |
| world-rust-composed-http/go/interpreter/4 | 6355 | 8988 | 0.6529 [0.5885,0.7015] | no | 493482 | 456299 |
| world-rust-composed-http/go/native/1 | 2091 | 3352 | 0.6243 [0.5901,0.6639] | no | 489903 | 452717 |
| world-rust-composed-http/go/native/4 | 7556 | 11129 | 0.6376 [0.5776,0.6848] | no | 489868 | 452688 |
| world-rust-composed-http/llgo/interpreter/1 | 1002 | 1621 | 0.6797 [0.6174,0.7800] | no | 958940 | 868931 |
| world-rust-composed-http/llgo/interpreter/4 | 1952 | 2604 | 0.7136 [0.6400,0.7678] | no | 959050 | 868887 |
| world-rust-composed-http/llgo/native/1 | 1159 | 1882 | 0.6106 [0.5680,0.6474] | no | 955455 | 865540 |
| world-rust-composed-http/llgo/native/4 | 1880 | 2640 | 0.6835 [0.5976,0.7549] | no | 955650 | 865495 |
| world-rust-http/go/interpreter/1 | 1690 | 2723 | 0.6207 [0.6086,0.6339] | no | 477644 | 440462 |
| world-rust-http/go/interpreter/4 | 6733 | 9816 | 0.6992 [0.6820,0.7229] | no | 477608 | 440428 |
| world-rust-http/go/native/1 | 2239 | 3692 | 0.6222 [0.5981,0.6538] | no | 478301 | 441109 |
| world-rust-http/go/native/4 | 7764 | 11734 | 0.6792 [0.6570,0.7077] | no | 478266 | 441085 |
| world-rust-http/llgo/interpreter/1 | 1064 | 1740 | 0.5835 [0.5163,0.6284] | no | 937171 | 847244 |
| world-rust-http/llgo/interpreter/4 | 1986 | 2638 | 0.7753 [0.7305,0.8349] | no | 937321 | 847285 |
| world-rust-http/llgo/native/1 | 1205 | 1991 | 0.6296 [0.5868,0.7109] | no | 938460 | 848551 |
| world-rust-http/llgo/native/4 | 1971 | 2666 | 0.7231 [0.6728,0.7644] | no | 938625 | 848551 |

## All profiles and controls

| Profile | req/s | Complete API us | Request p99 us | Runtime B/req | Between-batch threads | Peak RSS MiB | A/A 95% | Resolved |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| base/world-rust-composed-http/go/interpreter/1 | 1582 | 629.434 | 11340.875 | 493507 | 7 | 103.906 | [0.9615,1.0782] | no |
| base/world-rust-composed-http/go/interpreter/4 | 6355 | 610.805 | 4670.750 | 493482 | 7 | 137.391 | [0.8730,1.0007] | no |
| base/world-rust-composed-http/go/native/1 | 2091 | 475.741 | 8746.438 | 489903 | 7 | 145.609 | [0.9501,1.0729] | no |
| base/world-rust-composed-http/go/native/4 | 7556 | 509.487 | 4068.875 | 489868 | 7 | 158.500 | [0.8428,1.0212] | no |
| base/world-rust-composed-http/llgo/interpreter/1 | 1002 | 984.169 | 19733.062 | 958940 | 43 | 169.016 | [0.9838,1.1829] | no |
| base/world-rust-composed-http/llgo/interpreter/4 | 1952 | 2024.378 | 11688.083 | 959050 | 61 | 181.469 | [0.8537,1.0033] | no |
| base/world-rust-composed-http/llgo/native/1 | 1159 | 855.710 | 17287.521 | 955455 | 43 | 180.641 | [0.9908,1.0781] | no |
| base/world-rust-composed-http/llgo/native/4 | 1880 | 2084.411 | 11483.916 | 955650 | 61 | 198.734 | [0.7867,1.0476] | no |
| base/world-rust-http/go/interpreter/1 | 1690 | 589.282 | 10456.438 | 477644 | 7 | 78.094 | [0.9598,1.0137] | no |
| base/world-rust-http/go/interpreter/4 | 6733 | 575.692 | 3958.854 | 477608 | 7 | 98.797 | [0.9784,1.0542] | no |
| base/world-rust-http/go/native/1 | 2239 | 444.026 | 7885.250 | 478301 | 7 | 115.250 | [0.9791,1.0418] | no |
| base/world-rust-http/go/native/4 | 7764 | 496.962 | 3547.042 | 478266 | 7 | 121.562 | [0.9955,1.2566] | no |
| base/world-rust-http/llgo/interpreter/1 | 1064 | 932.133 | 18125.854 | 937171 | 43 | 140.484 | [0.8638,1.1567] | no |
| base/world-rust-http/llgo/interpreter/4 | 1986 | 1984.499 | 10609.104 | 937321 | 61 | 141.234 | [0.9585,1.0733] | no |
| base/world-rust-http/llgo/native/1 | 1205 | 821.235 | 16279.521 | 938460 | 43 | 168.047 | [0.9844,1.4354] | no |
| base/world-rust-http/llgo/native/4 | 1971 | 1993.258 | 10822.666 | 938625 | 61 | 176.297 | [0.8762,1.0049] | no |
| head/world-rust-composed-http/go/interpreter/1 | 2470 | 402.211 | 7785.105 | 456328 | 7 | 104.453 | [0.8980,1.0305] | no |
| head/world-rust-composed-http/go/interpreter/4 | 8988 | 427.453 | 3425.959 | 456299 | 7 | 137.516 | [0.9946,1.1307] | no |
| head/world-rust-composed-http/go/native/1 | 3352 | 295.808 | 5743.958 | 452717 | 7 | 148.297 | [0.9198,1.0228] | no |
| head/world-rust-composed-http/go/native/4 | 11129 | 345.148 | 3213.146 | 452688 | 7 | 159.391 | [1.0039,1.1692] | no |
| head/world-rust-composed-http/llgo/interpreter/1 | 1621 | 608.870 | 13760.709 | 868931 | 43 | 164.281 | [0.7925,0.9970] | no |
| head/world-rust-composed-http/llgo/interpreter/4 | 2604 | 1508.979 | 9414.437 | 868887 | 61 | 181.344 | [0.9757,1.0665] | no |
| head/world-rust-composed-http/llgo/native/1 | 1882 | 521.529 | 11707.729 | 865540 | 43 | 180.391 | [0.9620,1.1150] | no |
| head/world-rust-composed-http/llgo/native/4 | 2640 | 1487.059 | 9044.688 | 865495 | 61 | 209.578 | [0.9456,1.0919] | no |
| head/world-rust-http/go/interpreter/1 | 2723 | 364.646 | 6898.438 | 440462 | 7 | 80.078 | [0.9872,1.0493] | no |
| head/world-rust-http/go/interpreter/4 | 9816 | 393.804 | 3138.125 | 440428 | 8 | 105.641 | [0.9037,1.0269] | no |
| head/world-rust-http/go/native/1 | 3692 | 268.235 | 5020.083 | 441109 | 7 | 112.781 | [0.8843,0.9952] | no |
| head/world-rust-http/go/native/4 | 11734 | 329.112 | 2795.479 | 441085 | 8 | 127.906 | [0.9678,1.0290] | no |
| head/world-rust-http/llgo/interpreter/1 | 1740 | 566.488 | 12279.000 | 847244 | 43 | 140.391 | [0.9669,1.0064] | no |
| head/world-rust-http/llgo/interpreter/4 | 2638 | 1468.752 | 8417.438 | 847285 | 61 | 140.656 | [0.8705,0.9917] | no |
| head/world-rust-http/llgo/native/1 | 1991 | 493.829 | 10668.083 | 848551 | 43 | 168.000 | [0.8497,1.1482] | no |
| head/world-rust-http/llgo/native/4 | 2666 | 1468.376 | 8596.354 | 848551 | 61 | 175.938 | [0.9320,1.0201] | no |
| host/native-go/go/host/1 | 39088 | 21.625 | 560.167 | 277925 | 6 | 16.875 | [0.8809,1.0195] | no |
| host/native-go/go/host/4 | 44387 | 77.542 | 596.312 | 277923 | 6 | 16.266 | [0.9671,1.0742] | no |
| host/native-go/llgo/host/1 | 8390 | 113.734 | 3506.104 | 732590 | 37 | 68.156 | [1.0214,1.0737] | no |
| host/native-go/llgo/host/4 | 8803 | 411.050 | 3401.459 | 732593 | 40 | 68.406 | [0.9618,1.1174] | no |
| reference/world-rust-composed-http/go/compiler/1 | 14810 | 64.999 | 1225.375 | 32 | 7 | 32.516 | [0.7247,1.1246] | no |
| reference/world-rust-composed-http/go/compiler/4 | 50636 | 74.089 | 390.896 | 32 | 7 | 34.109 | [0.8876,1.0852] | no |
| reference/world-rust-http/go/compiler/1 | 16826 | 56.811 | 1097.667 | 32 | 7 | 30.156 | [0.8792,0.9887] | no |
| reference/world-rust-http/go/compiler/4 | 56524 | 66.071 | 362.084 | 33 | 7 | 30.891 | [0.8293,0.9987] | no |

## Complete lifecycle intervals

Median milliseconds over 20 primary process reductions. Warm is a process-mean batch. All per-process reductions, resource snapshots and 32 paired reference intervals are also in the machine record. Overlapping phases and different cohorts are never subtracted to infer an unmeasured cost.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1 | 5.5173 | 0.0680 | 0.0004 | 26.7270 | 0.0366 | 0.2369 | 0.0004 | N/A | 0.0124 | 44.8792 | 40.4516 | 0.0324 | 447.3831 |
| base/world-rust-composed-http/go/interpreter/4 | 5.6598 | 0.0631 | 0.0004 | 26.4976 | 0.0551 | 0.7079 | 0.0006 | N/A | 0.0162 | 12.5555 | 10.0710 | 0.0522 | 139.9531 |
| base/world-rust-composed-http/go/native/1 | 5.5033 | 0.0654 | 0.0004 | 220.6999 | 0.0357 | 0.2119 | 0.0004 | N/A | 0.0127 | 32.4918 | 30.6103 | 0.0378 | 537.3505 |
| base/world-rust-composed-http/go/native/4 | 5.5455 | 0.0643 | 0.0004 | 220.3811 | 0.0554 | 0.6726 | 0.0007 | N/A | 0.0135 | 10.1398 | 8.4703 | 0.0622 | 316.5963 |
| base/world-rust-composed-http/llgo/interpreter/1 | 8.9490 | 0.0737 | 0.0002 | 44.3288 | 0.1147 | 0.4331 | 0.0005 | N/A | 0.2744 | 67.3886 | 63.8909 | 0.2579 | 696.2411 |
| base/world-rust-composed-http/llgo/interpreter/4 | 9.0600 | 0.0735 | 0.0003 | 44.9786 | 0.2474 | 1.5534 | 0.0010 | N/A | 0.3148 | 36.6211 | 32.7984 | 0.6358 | 398.6850 |
| base/world-rust-composed-http/llgo/native/1 | 8.9992 | 0.0719 | 0.0002 | 327.8630 | 0.1108 | 0.2945 | 0.0005 | N/A | 0.2878 | 55.9772 | 55.2390 | 0.2563 | 896.2100 |
| base/world-rust-composed-http/llgo/native/4 | 9.1089 | 0.0747 | 0.0002 | 327.8158 | 0.2451 | 1.0025 | 0.0010 | N/A | 0.3266 | 35.4035 | 34.0405 | 0.6453 | 681.9883 |
| base/world-rust-http/go/interpreter/1 | 5.6882 | 0.0580 | 0.0003 | 18.5135 | 0.0385 | 0.1683 | 0.0003 | N/A | 0.0137 | 41.6892 | 37.8787 | 0.0308 | 409.4543 |
| base/world-rust-http/go/interpreter/4 | 5.4814 | 0.0607 | 0.0004 | 18.5162 | 0.0581 | 0.5390 | 0.0007 | N/A | 0.0159 | 11.2613 | 9.5062 | 0.0474 | 125.5896 |
| base/world-rust-http/go/native/1 | 5.5779 | 0.0599 | 0.0004 | 152.7026 | 0.0336 | 0.1630 | 0.0004 | N/A | 0.0131 | 30.1180 | 28.5813 | 0.0381 | 447.7473 |
| base/world-rust-http/go/native/4 | 5.5034 | 0.0593 | 0.0003 | 151.9200 | 0.0514 | 0.4812 | 0.0006 | N/A | 0.0139 | 8.7673 | 8.2434 | 0.0569 | 243.3360 |
| base/world-rust-http/llgo/interpreter/1 | 9.0084 | 0.0652 | 0.0002 | 29.9841 | 0.1134 | 2.8219 | 0.0005 | N/A | 0.2930 | 62.8484 | 60.1359 | 0.2530 | 648.0411 |
| base/world-rust-http/llgo/interpreter/4 | 8.9951 | 0.0668 | 0.0002 | 29.7371 | 0.2615 | 3.5751 | 0.0011 | N/A | 0.3036 | 32.1639 | 32.2328 | 0.6200 | 370.2713 |
| base/world-rust-http/llgo/native/1 | 8.9752 | 0.0653 | 0.0002 | 226.5939 | 0.1091 | 0.2203 | 0.0005 | N/A | 0.2828 | 55.2565 | 53.0919 | 0.2429 | 772.9152 |
| base/world-rust-http/llgo/native/4 | 9.0151 | 0.0674 | 0.0002 | 227.5869 | 0.2441 | 0.7505 | 0.0011 | N/A | 0.3342 | 34.3941 | 32.4731 | 0.6336 | 566.6696 |
| head/world-rust-composed-http/go/interpreter/1 | 5.6300 | 0.0649 | 0.0003 | 26.3496 | 0.0404 | 0.2419 | 0.0004 | N/A | 0.0138 | 29.4022 | 25.9063 | 0.0305 | 296.5947 |
| head/world-rust-composed-http/go/interpreter/4 | 5.5234 | 0.0662 | 0.0003 | 26.3244 | 0.0566 | 0.7086 | 0.0006 | N/A | 0.0156 | 9.5769 | 7.1209 | 0.0742 | 110.4322 |
| head/world-rust-composed-http/go/native/1 | 5.5484 | 0.0640 | 0.0004 | 220.5303 | 0.0381 | 0.2111 | 0.0003 | N/A | 0.0125 | 20.7675 | 19.0913 | 0.0369 | 421.6325 |
| head/world-rust-composed-http/go/native/4 | 5.5814 | 0.0636 | 0.0003 | 219.8850 | 0.0545 | 0.6452 | 0.0006 | N/A | 0.0134 | 6.2220 | 5.7505 | 0.0646 | 287.5707 |
| head/world-rust-composed-http/llgo/interpreter/1 | 9.1263 | 0.0729 | 0.0002 | 44.4696 | 0.1131 | 0.4304 | 0.0013 | N/A | 0.2856 | 44.2055 | 39.4808 | 0.2379 | 456.7642 |
| head/world-rust-composed-http/llgo/interpreter/4 | 8.9369 | 0.0767 | 0.0002 | 44.5902 | 0.2535 | 1.5021 | 0.0020 | N/A | 0.3130 | 27.1498 | 24.5766 | 0.6182 | 309.6975 |
| head/world-rust-composed-http/llgo/native/1 | 8.9832 | 0.0760 | 0.0002 | 328.6070 | 0.1100 | 0.2947 | 0.0013 | N/A | 0.2828 | 35.6923 | 33.9987 | 0.2307 | 683.2015 |
| head/world-rust-composed-http/llgo/native/4 | 9.0311 | 0.0779 | 0.0002 | 328.2076 | 0.2492 | 0.9918 | 0.0019 | N/A | 0.3296 | 25.8678 | 24.2582 | 0.6235 | 587.2617 |
| head/world-rust-http/go/interpreter/1 | 5.5229 | 0.0597 | 0.0003 | 18.2465 | 0.0423 | 0.1664 | 0.0004 | N/A | 0.0143 | 26.9619 | 23.5010 | 0.0294 | 265.1410 |
| head/world-rust-http/go/interpreter/4 | 5.5263 | 0.0612 | 0.0004 | 18.4202 | 0.0580 | 0.5171 | 0.0007 | N/A | 0.0152 | 8.3940 | 6.5200 | 0.0476 | 95.0051 |
| head/world-rust-http/go/native/1 | 5.4921 | 0.0613 | 0.0004 | 151.7472 | 0.0358 | 0.1685 | 0.0004 | N/A | 0.0127 | 18.6199 | 17.3347 | 0.0387 | 334.4758 |
| head/world-rust-http/go/native/4 | 5.5743 | 0.0591 | 0.0004 | 152.1526 | 0.0554 | 0.4841 | 0.0006 | N/A | 0.0145 | 6.2450 | 5.4543 | 0.0578 | 215.9154 |
| head/world-rust-http/llgo/interpreter/1 | 9.1202 | 0.0659 | 0.0003 | 30.2057 | 0.1064 | 2.8961 | 0.0015 | N/A | 0.2979 | 39.5092 | 36.7725 | 0.2417 | 414.9158 |
| head/world-rust-http/llgo/interpreter/4 | 9.0453 | 0.0706 | 0.0002 | 29.6284 | 0.2505 | 3.4051 | 0.0020 | N/A | 0.3124 | 26.0739 | 24.2613 | 0.6339 | 292.0672 |
| head/world-rust-http/llgo/native/1 | 8.8953 | 0.0691 | 0.0002 | 227.0268 | 0.1044 | 0.2243 | 0.0014 | N/A | 0.2836 | 35.1207 | 32.1403 | 0.2204 | 562.4245 |
| head/world-rust-http/llgo/native/4 | 9.0043 | 0.0693 | 0.0002 | 228.0448 | 0.2494 | 0.8022 | 0.0018 | N/A | 0.3250 | 25.4357 | 24.0036 | 0.6205 | 491.2380 |
| host/native-go/go/host/1 | 5.0272 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0141 | 1.8099 | 1.6373 | 0.0054 | 23.2190 |
| host/native-go/go/host/4 | 5.0888 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0166 | 1.3615 | 1.4419 | 0.0068 | 20.8574 |
| host/native-go/llgo/host/1 | 8.7317 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2615 | 9.5416 | 7.6279 | 0.0860 | 88.4557 |
| host/native-go/llgo/host/4 | 8.7701 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3275 | 10.6385 | 7.2708 | 0.0925 | 88.0840 |
| reference/world-rust-composed-http/go/compiler/1 | 5.9937 | 0.0698 | 0.1357 | 73.1503 | 0.1150 | 0.1357 | 0.0066 | 0.0286 | 0.0343 | 4.9079 | 4.3213 | 0.2405 | 125.2603 |
| reference/world-rust-composed-http/go/compiler/4 | 5.9802 | 0.0678 | 0.1368 | 74.5697 | 0.2222 | 0.2519 | 0.0082 | 0.0434 | 0.0499 | 1.6347 | 1.2640 | 0.3706 | 96.5895 |
| reference/world-rust-http/go/compiler/1 | 5.9101 | 0.0592 | 0.1326 | 51.8547 | 0.1157 | 0.1253 | 0.0067 | 0.0301 | 0.0366 | 4.3449 | 3.8037 | 0.1620 | 98.2875 |
| reference/world-rust-http/go/compiler/4 | 6.0834 | 0.0598 | 0.1336 | 51.9835 | 0.2241 | 0.1914 | 0.0080 | 0.0402 | 0.0462 | 1.4045 | 1.1324 | 0.2335 | 72.0183 |

## Validation, diagnostics and remaining work

Full normal Go and pinned LLGo complete P3 tests pass on both backends. Exact CI-style Go coverage instruments all owned packages together: 31 gates pass, minimum 90.86%. Affected HTTP/transfer race tests, P3 measurement race tests, vet and actual LLGo P3 measurement execution pass. New tests cover bounded external read calls, incremental guest reads, invalid-UTF8 tails, both ordinary EOF and data-plus-EOF, zero-length probes, source close and late delivery. Existing two-Store, cancellation, timeout, quota, owner shutdown and upstream guest tests remain active.

Two initial local diagnostic commands do not constitute passing acceptance. The first normal coverage invocation omitted -coverpkg and therefore did not measure the CI cross-package scope; its files are retained, and the exact CI-style run above passes. A complete P3 race diagnostic exceeded its three-minute package timeout while compiling the expanded upstream streaming corpus; its stack/log remain retained. The affected HTTP race scope then passes in 6.3 seconds. Neither diagnostic failure is hidden or relabeled as a full race pass. Final supported-platform semantic CI remains required.

The separate Go CPU/allocation/block/mutex diagnostic runs 20000 complete requests with profilers enabled. Its CPU profile includes preparation/calibration; Darwin pthread labels do not partition pure VM costs. Block totals include overlapping and idle worker waits. The separate LLGo five-second stack sample includes sleeping threads, which are not CPU shares; its complete runner verifies 24576 responses. Whole-run resource usage and raw stacks remain diagnostic evidence, not replacements for the fixed event. Sources, counter-instrumentation manifests, profiler outputs and failures are retained in artifacts/p3-gap-diagnosis-20261110.

The existing [bounded-buffer](performance-p3-io-buffers-2026-11-10.md), [worker-reuse](performance-p3-worker-reuse-2026-11-10.md), [immutable-copy](performance-immutable-memory-2026-11-10.md) and [rejected pipe](performance-p3-body-pipe-evaluation-2026-11-10.md) events remain independent. Next, isolate the output adapter from Store-local result consumption using the existing component CallWithResult contract, then address remaining dynamic type/metadata allocation and native compiler gaps. Both controlled targets and production release acceptance remain open. LLGo #3034/#3035 are being solved upstream; this phase does not duplicate those issues or upgrade the compiler. A separate [fresh zero-slice append diagnostic](performance-llgo-zero-append-2026-11-10.md) confirms redundant allocation in the pinned compiler and is reported as [#3040](https://github.com/xgo-dev/llgo/issues/3040); it is independent of this timed event.

Raw result SHA256 `20e6df2e2fc1d7030fcc5c7636abdfde3836b18603af18a5943b34e42c77a26e` in artifacts/p3-read-ahead-20261110. Its compact source bundle contains the head commit with the measured base as an explicit prerequisite; complete selected base/head source archives and hashes are also retained.
