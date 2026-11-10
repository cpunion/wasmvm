# Resource inspection and HTTP failure delivery, 2026-11-10

[Complete machine record](data/performance-reference-inspection-atomic-2026-11-10.json).

The [initial source event](performance-reference-inspection-2026-11-10.md) is retained independently. Initial CI failed because a worker cancelled its internal context immediately after publishing a specific Store-busy failure; the receiver could nondeterministically report `context canceled`. The final implementation keeps worker cancellation, including after an already published response; abandonment atomically captures any committed failure so the cancellation branch cannot mask it. The [intermediate event](performance-reference-inspection-corrected-2026-11-10.md) is retained separately and predates this review refinement. Error cleanup and real external cancellation remain enforced. This new fixed event measures that corrected production source, without pooling or replacing initial observations.

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
| wasmvm go / interpreter | 16327 | 16382 | 87556 | 85378 |
| wasmvm go / native | 19554 | 20125 | 87702 | 85523 |
| wasmvm llgo / interpreter | 5812 | 6007 | 158134 | 155423 |
| wasmvm llgo / native | 5677 | 5950 | 158609 | 155893 |
| Wasmtime 48.0.1 | N/A | 64656 | N/A | 33 |
| Ordinary native-Go application / Go | N/A | 1526182 | N/A | 2712 |
| Ordinary native-Go application / LLGo | N/A | 117941 | N/A | 3485 |

## HTTP fixture, 65536 bytes, four workers

| Path | Base req/s | Head req/s | Base runtime B/request | Head runtime B/request |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 10905 | 11685 | 252775 | 250606 |
| wasmvm go / native | 14989 | 15729 | 252919 | 250741 |
| wasmvm llgo / interpreter | 4494 | 4580 | 411896 | 409219 |
| wasmvm llgo / native | 4376 | 4216 | 412278 | 409438 |
| Wasmtime 48.0.1 | N/A | 51831 | N/A | 32 |
| Ordinary native-Go application / Go | N/A | 40701 | N/A | 277917 |
| Ordinary native-Go application / LLGo | N/A | 8409 | N/A | 732598 |

**2/40 A/A controls resolve** within ±2.5%; 0/16 source pairs resolve both. Default-workstation timings are advisory; neither a controlled ranking nor release-performance acceptance is established. Every adverse, unchanged and unresolved result is retained below.

Adverse pointwise warm-time intervals remain visible: composed Go native/16B is 1.0203 [1.0005,1.0436], composed Go native/64KiB is 1.0376 [1.0154,1.0659], and plain HTTP LLGo native/64KiB is 1.0427 [1.0207,1.0652]. Every pair has unresolved A/A controls; no stable speedup, parity or controlled regression conclusion is established.

Go managed bytes and LLGo collector accounting differ; Go counters exclude Rust/JIT allocation. Native-Go is an ordinary application control, not an optimized native lower bound. Resource instrumentation runs separately. Throughput medians and paired geometric warm-time ratios summarize different reductions.

## All source comparisons

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A | Base B/request | Head B/request |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| head/world-rust-composed-http/go/interpreter/4/16 | 13877 | 14175 | 0.8700 [0.6716,0.9927] | no | 102068 | 99890 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10044 | 10022 | 1.0123 [0.9983,1.0268] | no | 267286 | 265107 |
| head/world-rust-composed-http/go/native/4/16 | 17642 | 17532 | 1.0203 [1.0005,1.0436] | no | 97939 | 95762 |
| head/world-rust-composed-http/go/native/4/65536 | 13965 | 13686 | 1.0376 [1.0154,1.0659] | no | 263158 | 260979 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 5776 | 5807 | 0.9994 [0.9625,1.0547] | no | 177327 | 174623 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4323 | 4195 | 0.9940 [0.9370,1.0545] | no | 430941 | 428222 |
| head/world-rust-composed-http/llgo/native/4/16 | 5494 | 5544 | 1.0015 [0.9526,1.0586] | no | 172980 | 170293 |
| head/world-rust-composed-http/llgo/native/4/65536 | 4244 | 4242 | 0.9798 [0.9349,1.0269] | no | 426660 | 423944 |
| head/world-rust-http/go/interpreter/4/16 | 16327 | 16382 | 1.0153 [0.9968,1.0393] | no | 87556 | 85378 |
| head/world-rust-http/go/interpreter/4/65536 | 10905 | 11685 | 0.9016 [0.8238,0.9496] | no | 252775 | 250606 |
| head/world-rust-http/go/native/4/16 | 19554 | 20125 | 0.9646 [0.9269,1.0107] | no | 87702 | 85523 |
| head/world-rust-http/go/native/4/65536 | 14989 | 15729 | 0.9666 [0.9228,1.0228] | no | 252919 | 250741 |
| head/world-rust-http/llgo/interpreter/4/16 | 5812 | 6007 | 0.9718 [0.9547,0.9910] | no | 158134 | 155423 |
| head/world-rust-http/llgo/interpreter/4/65536 | 4494 | 4580 | 1.0041 [0.9577,1.0657] | no | 411896 | 409219 |
| head/world-rust-http/llgo/native/4/16 | 5677 | 5950 | 0.9662 [0.9432,1.0052] | no | 158609 | 155893 |
| head/world-rust-http/llgo/native/4/65536 | 4376 | 4216 | 1.0427 [1.0207,1.0652] | no | 412278 | 409438 |

## Every profile and separate resource cohort

| Profile | req/s | API us | p99 request us | Runtime B/request | Go mallocs/request | Between-batch threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/4/16 | 13877 | 277.902 | 2405.166 | 102068 | 1445.0 | 8 | 97.281 | no |
| base/world-rust-composed-http/go/interpreter/4/65536 | 10044 | 378.778 | 3601.667 | 267286 | 1853.0 | 7 | 107.844 | no |
| base/world-rust-composed-http/go/native/4/16 | 17642 | 218.037 | 1954.667 | 97939 | 1451.0 | 7 | 154.188 | no |
| base/world-rust-composed-http/go/native/4/65536 | 13965 | 273.134 | 2782.792 | 263158 | 1859.0 | 7 | 158.719 | no |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 5776 | 677.005 | 5865.583 | 177327 | N/A | 57 | 151.000 | no |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 4323 | 900.631 | 6889.167 | 430941 | N/A | 57 | 158.344 | no |
| base/world-rust-composed-http/llgo/native/4/16 | 5494 | 712.517 | 5952.584 | 172980 | N/A | 57 | 195.766 | no |
| base/world-rust-composed-http/llgo/native/4/65536 | 4244 | 916.581 | 6754.291 | 426660 | N/A | 57 | 197.469 | no |
| base/world-rust-http/go/interpreter/4/16 | 16327 | 236.428 | 2046.459 | 87556 | 1318.0 | 7 | 73.859 | yes |
| base/world-rust-http/go/interpreter/4/65536 | 10905 | 354.430 | 3011.166 | 252775 | 1726.0 | 7 | 86.484 | no |
| base/world-rust-http/go/native/4/16 | 19554 | 197.634 | 2036.500 | 87702 | 1321.0 | 7 | 122.438 | no |
| base/world-rust-http/go/native/4/65536 | 14989 | 255.883 | 2562.833 | 252919 | 1729.0 | 7 | 125.031 | no |
| base/world-rust-http/llgo/interpreter/4/16 | 5812 | 674.340 | 5733.084 | 158134 | N/A | 57 | 120.078 | no |
| base/world-rust-http/llgo/interpreter/4/65536 | 4494 | 866.106 | 6031.291 | 411896 | N/A | 57 | 140.828 | no |
| base/world-rust-http/llgo/native/4/16 | 5677 | 679.962 | 5390.583 | 158609 | N/A | 57 | 175.203 | no |
| base/world-rust-http/llgo/native/4/65536 | 4376 | 895.666 | 6247.666 | 412278 | N/A | 57 | 176.641 | no |
| head/world-rust-composed-http/go/interpreter/4/16 | 14175 | 271.300 | 2234.708 | 99890 | 1394.0 | 7 | 100.641 | no |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10022 | 382.354 | 3734.500 | 265107 | 1802.0 | 8 | 106.125 | no |
| head/world-rust-composed-http/go/native/4/16 | 17532 | 220.375 | 2149.708 | 95762 | 1400.0 | 7 | 157.547 | yes |
| head/world-rust-composed-http/go/native/4/65536 | 13686 | 280.658 | 3122.333 | 260979 | 1808.0 | 7 | 160.062 | no |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 5807 | 673.172 | 5587.291 | 174623 | N/A | 57 | 151.016 | no |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4195 | 926.917 | 6861.708 | 428222 | N/A | 57 | 152.703 | no |
| head/world-rust-composed-http/llgo/native/4/16 | 5544 | 706.608 | 5995.584 | 170293 | N/A | 57 | 196.125 | no |
| head/world-rust-composed-http/llgo/native/4/65536 | 4242 | 917.546 | 6663.583 | 423944 | N/A | 57 | 193.594 | no |
| head/world-rust-http/go/interpreter/4/16 | 16382 | 234.933 | 1865.791 | 85378 | 1267.0 | 7 | 73.953 | no |
| head/world-rust-http/go/interpreter/4/65536 | 11685 | 328.911 | 2768.750 | 250606 | 1675.0 | 7 | 91.484 | no |
| head/world-rust-http/go/native/4/16 | 20125 | 191.643 | 1934.209 | 85523 | 1270.0 | 7 | 117.578 | no |
| head/world-rust-http/go/native/4/65536 | 15729 | 241.407 | 2455.875 | 250741 | 1678.0 | 7 | 125.438 | no |
| head/world-rust-http/llgo/interpreter/4/16 | 6007 | 650.448 | 5377.750 | 155423 | N/A | 57 | 119.703 | no |
| head/world-rust-http/llgo/interpreter/4/65536 | 4580 | 853.478 | 6032.791 | 409219 | N/A | 57 | 140.859 | no |
| head/world-rust-http/llgo/native/4/16 | 5950 | 658.168 | 5183.041 | 155893 | N/A | 57 | 159.625 | no |
| head/world-rust-http/llgo/native/4/65536 | 4216 | 917.138 | 6276.250 | 409438 | N/A | 57 | 177.703 | no |
| host/native-go/go/host/4/16 | 1526182 | 0.802 | 37.166 | 2712 | 18.0 | 5 | 11.469 | no |
| host/native-go/go/host/4/65536 | 40701 | 84.236 | 690.083 | 277917 | 48.1 | 5 | 15.922 | no |
| host/native-go/llgo/host/4/16 | 117941 | 2.125 | 320.750 | 3485 | N/A | 40 | 12.547 | no |
| host/native-go/llgo/host/4/65536 | 8409 | 431.014 | 4160.417 | 732598 | N/A | 40 | 68.125 | no |
| reference/world-rust-composed-http/go/compiler/4/16 | 55110 | 69.403 | 367.042 | 44 | 1.0 | 7 | 33.297 | no |
| reference/world-rust-composed-http/go/compiler/4/65536 | 48502 | 77.601 | 433.042 | 32 | 1.0 | 7 | 34.125 | no |
| reference/world-rust-http/go/compiler/4/16 | 64656 | 58.758 | 327.166 | 33 | 1.0 | 8 | 30.469 | no |
| reference/world-rust-http/go/compiler/4/65536 | 51831 | 71.889 | 410.709 | 32 | 1.0 | 7 | 30.484 | no |

## Separate lifecycle intervals

Median milliseconds across five primary process reductions. Startup/readiness, load, engine setup, compile/validate, providers, instantiate, lookup, executor/workers, first batch, warm execution and joined cleanup remain separate. Process wall includes protocol parking; it is not standalone CLI latency. Host binary compilation is outside runtime sampling. No overlapping phases or cohorts are subtracted.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/4/16 | 5.8143 | 0.0715 | 0.0003 | 26.9609 | 0.0594 | 0.6611 | 0.0006 | N/A | 0.0156 | 5.9586 | 4.6120 | 0.0592 | 84.0085 |
| base/world-rust-composed-http/go/interpreter/4/65536 | 5.8213 | 0.0842 | 0.0004 | 26.9875 | 0.0613 | 0.6812 | 0.0007 | N/A | 0.0152 | 9.0708 | 6.3722 | 0.0590 | 102.2665 |
| base/world-rust-composed-http/go/native/4/16 | 5.8154 | 0.0788 | 0.0006 | 223.8592 | 0.0590 | 0.6485 | 0.0006 | N/A | 0.0153 | 4.3188 | 3.6277 | 0.0816 | 270.2961 |
| base/world-rust-composed-http/go/native/4/65536 | 5.6868 | 0.0884 | 0.0004 | 224.1717 | 0.0578 | 0.6803 | 0.0007 | N/A | 0.0157 | 5.0030 | 4.5828 | 0.0688 | 281.7259 |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 9.5186 | 0.0941 | 0.0003 | 44.8789 | 0.2406 | 1.7174 | 0.0010 | N/A | 0.3184 | 12.1648 | 11.0804 | 0.6372 | 176.4890 |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 9.6030 | 0.1015 | 0.0003 | 42.3206 | 0.2441 | 1.6030 | 0.0011 | N/A | 0.3161 | 18.5779 | 14.8040 | 0.6221 | 213.9898 |
| base/world-rust-composed-http/llgo/native/4/16 | 9.3932 | 0.0859 | 0.0003 | 327.3579 | 0.2530 | 1.0420 | 0.0010 | N/A | 0.3142 | 12.8949 | 11.6481 | 0.6431 | 465.0275 |
| base/world-rust-composed-http/llgo/native/4/65536 | 9.7808 | 0.0956 | 0.0003 | 328.4221 | 0.2651 | 1.0068 | 0.0013 | N/A | 0.3495 | 17.5358 | 15.0803 | 0.6297 | 495.4592 |
| base/world-rust-http/go/interpreter/4/16 | 5.7179 | 0.0651 | 0.0004 | 19.2350 | 0.0723 | 0.9865 | 0.0009 | N/A | 0.0185 | 4.4496 | 3.9198 | 0.0483 | 68.7841 |
| base/world-rust-http/go/interpreter/4/65536 | 5.9472 | 0.0702 | 0.0004 | 19.0817 | 0.0660 | 0.5456 | 0.0007 | N/A | 0.0201 | 7.3862 | 5.8690 | 0.0471 | 89.0621 |
| base/world-rust-http/go/native/4/16 | 5.7397 | 0.0640 | 0.0004 | 154.8216 | 0.0513 | 0.4842 | 0.0006 | N/A | 0.0147 | 3.5436 | 3.2729 | 0.0585 | 197.1693 |
| base/world-rust-http/go/native/4/65536 | 5.7926 | 0.0651 | 0.0004 | 153.6594 | 0.0572 | 0.4676 | 0.0006 | N/A | 0.0145 | 4.6203 | 4.2698 | 0.0606 | 207.1365 |
| base/world-rust-http/llgo/interpreter/4/16 | 9.3263 | 0.0826 | 0.0003 | 30.2650 | 0.2995 | 3.3430 | 0.0011 | N/A | 0.3208 | 11.4711 | 11.0108 | 0.6399 | 155.4545 |
| base/world-rust-http/llgo/interpreter/4/65536 | 9.2345 | 0.0882 | 0.0003 | 29.7664 | 0.2833 | 3.3647 | 0.0009 | N/A | 0.3220 | 14.7974 | 14.2400 | 0.6490 | 194.2785 |
| base/world-rust-http/llgo/native/4/16 | 9.3655 | 0.0844 | 0.0003 | 227.6611 | 0.2436 | 0.7404 | 0.0009 | N/A | 0.3223 | 12.3442 | 11.2737 | 0.6492 | 356.4498 |
| base/world-rust-http/llgo/native/4/65536 | 9.1687 | 0.0832 | 0.0003 | 226.9201 | 0.2597 | 0.7666 | 0.0010 | N/A | 0.3033 | 14.9743 | 14.6259 | 0.6833 | 391.7283 |
| head/world-rust-composed-http/go/interpreter/4/16 | 5.9716 | 0.0736 | 0.0005 | 26.7765 | 0.0583 | 0.6764 | 0.0007 | N/A | 0.0166 | 6.1821 | 4.5151 | 0.0586 | 83.8636 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 5.8712 | 0.0782 | 0.0004 | 26.9720 | 0.0660 | 0.7430 | 0.0010 | N/A | 0.0224 | 8.6925 | 6.3857 | 0.0538 | 102.4494 |
| head/world-rust-composed-http/go/native/4/16 | 5.9935 | 0.0839 | 0.0004 | 226.0150 | 0.0547 | 0.6595 | 0.0007 | N/A | 0.0134 | 4.2105 | 3.6504 | 0.0718 | 274.1219 |
| head/world-rust-composed-http/go/native/4/65536 | 5.6529 | 0.0710 | 0.0005 | 225.9837 | 0.0569 | 0.6388 | 0.0007 | N/A | 0.0138 | 5.2422 | 4.6763 | 0.0747 | 285.7317 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 9.4277 | 0.0851 | 0.0002 | 44.2852 | 0.2502 | 1.6695 | 0.0012 | N/A | 0.3127 | 12.2655 | 11.0203 | 0.6135 | 173.3428 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 9.2845 | 0.0968 | 0.0003 | 44.7695 | 0.2573 | 1.6734 | 0.0010 | N/A | 0.3859 | 18.6715 | 15.2546 | 0.6533 | 214.5118 |
| head/world-rust-composed-http/llgo/native/4/16 | 9.6497 | 0.0840 | 0.0004 | 326.6187 | 0.2974 | 1.0162 | 0.0012 | N/A | 0.3280 | 14.5457 | 11.5444 | 0.6247 | 460.5682 |
| head/world-rust-composed-http/llgo/native/4/65536 | 9.5071 | 0.0990 | 0.0003 | 328.9632 | 0.2757 | 1.0090 | 0.0011 | N/A | 0.3589 | 17.5545 | 15.0871 | 0.6202 | 499.7683 |
| head/world-rust-http/go/interpreter/4/16 | 5.6521 | 0.0770 | 0.0004 | 18.8885 | 0.0710 | 0.8355 | 0.0010 | N/A | 0.0152 | 4.3759 | 3.9068 | 0.0539 | 69.0355 |
| head/world-rust-http/go/interpreter/4/65536 | 5.9306 | 0.0809 | 0.0003 | 18.7705 | 0.0588 | 0.4885 | 0.0008 | N/A | 0.0146 | 7.4993 | 5.4773 | 0.0524 | 85.4597 |
| head/world-rust-http/go/native/4/16 | 5.6945 | 0.0670 | 0.0003 | 155.1242 | 0.0512 | 0.4668 | 0.0008 | N/A | 0.0146 | 3.4911 | 3.1802 | 0.0676 | 196.2882 |
| head/world-rust-http/go/native/4/65536 | 5.6613 | 0.0557 | 0.0004 | 153.8552 | 0.0600 | 0.4672 | 0.0007 | N/A | 0.0151 | 5.8192 | 4.0689 | 0.0669 | 206.3743 |
| head/world-rust-http/llgo/interpreter/4/16 | 9.4188 | 0.0848 | 0.0004 | 30.0819 | 0.3042 | 3.4470 | 0.0014 | N/A | 0.3238 | 11.5346 | 10.6535 | 0.6162 | 155.7562 |
| head/world-rust-http/llgo/interpreter/4/65536 | 9.2725 | 0.0760 | 0.0002 | 30.4780 | 0.3013 | 3.5728 | 0.0013 | N/A | 0.2919 | 14.4945 | 13.9727 | 0.6248 | 192.0718 |
| head/world-rust-http/llgo/native/4/16 | 9.3614 | 0.0866 | 0.0004 | 228.6634 | 0.2381 | 0.7472 | 0.0012 | N/A | 0.3803 | 11.7157 | 10.7564 | 0.6292 | 352.7511 |
| head/world-rust-http/llgo/native/4/65536 | 9.1262 | 0.0803 | 0.0003 | 228.8715 | 0.2400 | 0.7644 | 0.0009 | N/A | 0.3472 | 17.3187 | 15.1791 | 0.6188 | 400.1227 |
| host/native-go/go/host/4/16 | 5.3200 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0262 | 0.0602 | 0.0419 | 0.0065 | 6.8367 |
| host/native-go/go/host/4/65536 | 5.1891 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0212 | 1.5673 | 1.5724 | 0.0067 | 23.3650 |
| host/native-go/llgo/host/4/16 | 9.2965 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3403 | 0.7413 | 0.5426 | 0.0954 | 16.8870 |
| host/native-go/llgo/host/4/65536 | 8.8564 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3216 | 10.4109 | 7.6109 | 0.0868 | 91.1217 |
| reference/world-rust-composed-http/go/compiler/4/16 | 6.0077 | 0.0654 | 0.1346 | 75.4782 | 0.2168 | 0.2570 | 0.0080 | 0.0442 | 0.0561 | 1.4421 | 1.1613 | 0.4119 | 96.1414 |
| reference/world-rust-composed-http/go/compiler/4/65536 | 6.4196 | 0.0771 | 0.1458 | 75.5047 | 0.2172 | 0.2480 | 0.0080 | 0.0408 | 0.0722 | 1.6193 | 1.3195 | 0.3838 | 98.6312 |
| reference/world-rust-http/go/compiler/4/16 | 6.3273 | 0.0598 | 0.1409 | 52.7112 | 0.2229 | 0.1951 | 0.0081 | 0.0440 | 0.0478 | 1.2996 | 0.9899 | 0.2366 | 71.7643 |
| reference/world-rust-http/go/compiler/4/65536 | 6.6372 | 0.0813 | 0.1569 | 53.3743 | 0.2291 | 0.2025 | 0.0080 | 0.0438 | 0.0581 | 1.6381 | 1.2348 | 0.2525 | 75.5373 |

## Frozen protocol and validation

Benchmark base `23b54f0b121a1bd2481d2d445186f40dbcea9a94` is production-equivalent to main `da00a76`; its additional benchmark is identical to head `ffe0211957c1166aff0245562453a5bd3fd20af0`. Production LLGo stays `3d92a58`, `-O3 -lto=full`; Go 1.27.3, LLVM 22.1.8, Wasmtime 48.0.1. The base reuses prior exact binaries after checking every production/harness source hash; only the uncompiled benchmark test differs. No previous timing is reused. Reference compiler/program/SDK hashes remain pinned. Base/head measurement code is identical. Frozen reference differences are limited to owned-buffered flags/report metadata and cold options validation; warm driver, complete response checker and native/reference kernels are checked identical.

Five fixed randomized primary/A/A blocks, seed 2026111029; pointwise 95% paired log-ratio bootstrap, 8000 draws. Both Rust plain/composed HTTP fixtures, both owned backends/compilers, four workers, 16 clients/four requests, two warmups/seven samples, 16-byte/64-KiB payloads. Exactly 400 throughput processes, 40 separate resource processes, 80 qualifications, seven actual loader probes and 179200 checked warm responses. CRC32 independently checked. No source/profile/ISA/count search, favorable repeat, cross-event pooling, affinity/frequency isolation or family-wise claim.

Proportional local validation passes four affected Go and LLGo root packages (canonical/component/P2/P3), LLGo nested P3 measurement, two race packages and vet. Integration coverage of canonical/component exceeds 90% per package, minimum 92.36%; the initial unit-only canonical run measured 89.0%, so cross-package coverage was measured rather than adding tests that mirror the implementation. Every owned build/test/probe exited before timing. Actual pre-start process/cwd inspection records foreign workloads without interrupting them. Full final-head CI still requires all 32 root packages on both compilers/platforms, 31 coverage gates and all eight LLGo shards.

Existing semantic tests cover live/nil/expired references, copied aliases, affine movement, ended borrows and concurrent movement. The final cancellation fix preserves worker interruption after a response has already been published and atomically retrieves a committed failure when abandoning the reply. The regression covers both already published and late failures; busy admission and P3 lifecycle race tests pass. Ownership, quota and guest-memory contracts remain enforced.

CI timing review of PR #153 found maximum 608 seconds, with root compile/link stages dominating slow shards and nested P3 compilation dominating legacy on macOS. Moving nested execution alone would move the bottleneck rather than demonstrate a shorter maximum; this PR preserves the complete matrix and adds no CI job. Five-to-ten-minute jobs remain the target.

Broader [all-runtime service measurements](performance-service-all-runtimes-2026-11-10.md) retain wazero/Wasmer and other execution portfolios. They are not pooled with this P3 event. The [compiler study](performance-llgo-merged-2026-11-10.md) and LLGo GC issue #3045 remain separate; no compiler upgrade is included. Native/Core encoding, caller/codec allocation, interpreter, SIMD/memory and controlled performance gaps remain open.

The final reply-abandonment and busy-Store regression tests pass 40 race-mode repetitions; the remaining affected P3 race suite and vet also pass. Earlier complete 80-profile one/four-worker events are historical; this final follow-up predeclares all 40 four-worker profiles and leaves one-worker final-source performance unmeasured.
