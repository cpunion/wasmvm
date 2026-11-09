# Complete P3 HTTP I/O comparison: 2026-11-10

[Machine record](data/performance-p3io-2026-11-10.json).

M4 Max/macOS27.0.1, source `bccfba51c8a97363c6c0942efaa07cb79bea0d92`, pinned Go1.27.3, LLGo3d92a58 `-O3 -lto=full`, Rust1.98.0/LLVM22.1.8 and actual Wasmtime48.0.1 P3 providers. Production VM/provider files match the source baseline; only measurement code differs. This is a separate event from the Core SHA comparison.

All360 throughput processes,36 separate resource processes,72 qualification processes and161,280 checked warm responses completed. Each response checks all64KiB, status201, x-guest=rust-p3 and trailer=done. The returned CRC denotes a byte-validated body, not a checksum computed by the guest. **A/A resolves6/36 profiles; controlled performance/release acceptance remains open.**

Both the pinned Rust service and composed middleware/service use the shared bounded service driver:16 clients,4 requests/client,1/2/4 workers,2 warmup/7 measured batches,5 fresh blocks and independent identical-binary aliases. Every adverse observation remains present. No phase subtraction, retiming or cross-host/cohort pooling. Host power/frequency/affinity/background isolation are uncontrolled.

Owned Go/LLGo use independent World/Store instances sharing compiled code, unmetered calls with logical/cancellation limits, and public CallIncoming including handler lookup/signature checks. The independent Go C/Rust reference shares the admission driver and uses one serialized Store/current-thread Tokio executor per worker. Its generated export binding is cached; bridge/executor entry remain inside its complete API interval. Actual Store::run_concurrent drives response body/trailer collection. Outgoing network access is denied. Neither case measures a network socket.

Bare native-Go controls exclude the VM and Rust bridge from their Go/LLGo builds. They perform ordinary in-process echo/body collection with the same response and driver; they are application controls, not an optimized memcpy lower bound or a P3 runtime. Go and Rust body/collector implementations have different allocation/copy strategies. Go managed-byte and LLGo BDWGC counters are distinct; the reference Go counter does not include Rust/JIT allocation. Actual separate-process RSS/thread snapshots supplement them.

## All profiles

| Profile | req/s | Complete API µs | Request p50 µs | Request p99 µs | Runtime B/req | Idle threads | Idle RSS MiB | Peak RSS MiB |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| native-go / native-go / go / host / 1w | 40281 | 20.973 | 376.666 | 517.041 | 277914 | 6 | 16.625 | 16.625 |
| native-go / native-go / go / host / 2w | 48742 | 35.902 | 315.584 | 451.042 | 277916 | 6 | 16.812 | 16.812 |
| native-go / native-go / go / host / 4w | 44512 | 76.874 | 323.750 | 579.875 | 277926 | 6 | 15.953 | 15.953 |
| native-go / native-go / llgo / host / 1w | 8234 | 116.013 | 1811.208 | 3272.708 | 732578 | 37 | 67.578 | 67.609 |
| native-go / native-go / llgo / host / 2w | 7964 | 229.397 | 1834.750 | 3645.541 | 732575 | 38 | 63.812 | 63.844 |
| native-go / native-go / llgo / host / 4w | 9063 | 413.657 | 1857.667 | 3636.167 | 732586 | 40 | 68.328 | 68.359 |
| world-rust-composed-http / wasmtime / go / compiler / 1w | 14649 | 65.784 | 1061.417 | 1250.208 | 32 | 7 | 32.484 | 32.516 |
| world-rust-composed-http / wasmtime / go / compiler / 2w | 27134 | 70.210 | 568.875 | 665.542 | 32 | 7 | 33.391 | 33.391 |
| world-rust-composed-http / wasmtime / go / compiler / 4w | 51392 | 73.076 | 292.750 | 375.917 | 32 | 7 | 33.969 | 34.000 |
| world-rust-composed-http / wasmvm / go / interpreter / 1w | 1480 | 673.272 | 10587.875 | 11945.000 | 821664 | 7 | 105.859 | 105.859 |
| world-rust-composed-http / wasmvm / go / interpreter / 2w | 2925 | 676.286 | 5175.666 | 6527.291 | 821672 | 8 | 116.500 | 116.500 |
| world-rust-composed-http / wasmvm / go / interpreter / 4w | 5813 | 665.973 | 2331.584 | 4927.459 | 821650 | 7 | 144.938 | 144.938 |
| world-rust-composed-http / wasmvm / go / native / 1w | 1890 | 526.472 | 8223.000 | 9341.625 | 818082 | 7 | 150.516 | 150.516 |
| world-rust-composed-http / wasmvm / go / native / 2w | 3719 | 530.626 | 4041.083 | 5213.417 | 818079 | 8 | 153.531 | 153.531 |
| world-rust-composed-http / wasmvm / go / native / 4w | 6745 | 574.274 | 1941.791 | 4477.333 | 818045 | 7 | 160.859 | 160.859 |
| world-rust-composed-http / wasmvm / llgo / interpreter / 1w | 828 | 1199.827 | 18097.459 | 22474.708 | 1358945 | 40 | 163.875 | 163.953 |
| world-rust-composed-http / wasmvm / llgo / interpreter / 2w | 1340 | 1482.624 | 10731.416 | 15074.334 | 1359510 | 43 | 166.484 | 166.625 |
| world-rust-composed-http / wasmvm / llgo / interpreter / 4w | 1426 | 2775.488 | 10081.500 | 15162.750 | 1359342 | 49 | 166.688 | 167.000 |
| world-rust-composed-http / wasmvm / llgo / native / 1w | 915 | 1085.444 | 16088.250 | 20387.291 | 1355790 | 40 | 193.688 | 193.750 |
| world-rust-composed-http / wasmvm / llgo / native / 2w | 1403 | 1385.865 | 10035.792 | 14757.917 | 1356127 | 43 | 200.828 | 200.984 |
| world-rust-composed-http / wasmvm / llgo / native / 4w | 1387 | 2844.271 | 10252.958 | 14779.458 | 1355827 | 49 | 210.062 | 210.344 |
| world-rust-http / wasmtime / go / compiler / 1w | 17114 | 55.941 | 919.333 | 1043.333 | 32 | 7 | 30.094 | 30.156 |
| world-rust-http / wasmtime / go / compiler / 2w | 31503 | 60.202 | 495.583 | 563.083 | 32 | 8 | 30.453 | 30.453 |
| world-rust-http / wasmtime / go / compiler / 4w | 58929 | 63.426 | 254.750 | 337.750 | 32 | 7 | 30.844 | 30.891 |
| world-rust-http / wasmvm / go / interpreter / 1w | 1563 | 637.009 | 10045.709 | 11078.416 | 805799 | 7 | 81.891 | 81.891 |
| world-rust-http / wasmvm / go / interpreter / 2w | 3132 | 632.495 | 4876.916 | 5891.334 | 805797 | 8 | 87.734 | 87.766 |
| world-rust-http / wasmvm / go / interpreter / 4w | 6044 | 642.220 | 2266.500 | 4155.875 | 805778 | 7 | 104.562 | 104.562 |
| world-rust-http / wasmvm / go / native / 1w | 2033 | 489.339 | 7666.458 | 8378.417 | 806462 | 7 | 110.016 | 110.016 |
| world-rust-http / wasmvm / go / native / 2w | 3918 | 504.144 | 3831.833 | 4696.750 | 806463 | 8 | 115.625 | 115.625 |
| world-rust-http / wasmvm / go / native / 4w | 7032 | 551.812 | 1951.375 | 3769.750 | 806424 | 7 | 125.375 | 125.375 |
| world-rust-http / wasmvm / llgo / interpreter / 1w | 861 | 1149.566 | 17500.208 | 21294.083 | 1337518 | 40 | 139.859 | 139.938 |
| world-rust-http / wasmvm / llgo / interpreter / 2w | 1379 | 1433.964 | 10428.542 | 13573.917 | 1338835 | 43 | 140.047 | 140.219 |
| world-rust-http / wasmvm / llgo / interpreter / 4w | 1407 | 2811.465 | 10159.667 | 14637.583 | 1337884 | 49 | 140.375 | 140.688 |
| world-rust-http / wasmvm / llgo / native / 1w | 957 | 1037.720 | 15841.500 | 19475.917 | 1338672 | 40 | 168.359 | 168.422 |
| world-rust-http / wasmvm / llgo / native / 2w | 1454 | 1362.318 | 9710.291 | 13822.542 | 1338796 | 43 | 169.406 | 169.562 |
| world-rust-http / wasmvm / llgo / native / 4w | 1391 | 2833.368 | 10536.333 | 15087.541 | 1338875 | 49 | 176.281 | 176.578 |

## Complete lifecycle intervals

Median ms. First is a whole first batch, warm is the mean batch per process reduced over five independent primary processes. Peak RSS is through the final pre-cleanup snapshot. Body collection/close are inside API time; complete byte/metadata checking follows it and remains inside request wall time. Go request construction is outside the API timer and inside request wall; the Rust bridge constructs its request inside the bridge interval. These policies remain explicit.

| Profile | Process ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| native-go / native-go / go / host / 1w | 4.7633 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0167 | 1.7981 | 1.5888 | 0.0050 | 22.5619 |
| native-go / native-go / go / host / 2w | 4.9894 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0145 | 1.5804 | 1.3130 | 0.0051 | 19.9921 |
| native-go / native-go / go / host / 4w | 4.8800 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0173 | 1.4257 | 1.4378 | 0.0080 | 20.7093 |
| native-go / native-go / llgo / host / 1w | 8.7291 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3063 | 10.1629 | 7.7726 | 0.1112 | 90.4701 |
| native-go / native-go / llgo / host / 2w | 8.5691 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3008 | 10.1071 | 8.0360 | 0.0972 | 90.7470 |
| native-go / native-go / llgo / host / 4w | 8.7476 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3028 | 10.1240 | 7.0617 | 0.1070 | 86.6830 |
| world-rust-composed-http / wasmtime / go / compiler / 1w | 5.9241 | 0.0705 | 0.1359 | 73.7379 | 0.1123 | 0.1265 | 0.0064 | 0.0306 | 0.0386 | 4.8561 | 4.3689 | 0.2240 | 126.1669 |
| world-rust-composed-http / wasmtime / go / compiler / 2w | 5.8211 | 0.0607 | 0.1335 | 72.1295 | 0.1473 | 0.1819 | 0.0077 | 0.0304 | 0.0443 | 2.7379 | 2.3586 | 0.2797 | 103.9843 |
| world-rust-composed-http / wasmtime / go / compiler / 4w | 5.8464 | 0.0659 | 0.1364 | 72.9226 | 0.2227 | 0.2577 | 0.0083 | 0.0390 | 0.0358 | 1.5645 | 1.2453 | 0.3878 | 94.6773 |
| world-rust-composed-http / wasmvm / go / interpreter / 1w | 5.4494 | 0.0643 | 0.0003 | 26.7199 | 0.0398 | 0.2401 | 0.0004 | N/A | 0.0120 | 48.0770 | 43.2574 | 0.0331 | 474.4555 |
| world-rust-composed-http / wasmvm / go / interpreter / 2w | 5.2663 | 0.0652 | 0.0003 | 26.3693 | 0.0468 | 0.4165 | 0.0005 | N/A | 0.0152 | 25.3697 | 21.8778 | 0.0382 | 257.5185 |
| world-rust-composed-http / wasmvm / go / interpreter / 4w | 5.4482 | 0.0645 | 0.0003 | 26.1727 | 0.0591 | 0.6956 | 0.0005 | N/A | 0.0154 | 13.1299 | 11.0098 | 0.0471 | 148.5201 |
| world-rust-composed-http / wasmvm / go / native / 1w | 5.6502 | 0.0629 | 0.0004 | 218.3887 | 0.0364 | 0.2128 | 0.0003 | N/A | 0.0122 | 34.5184 | 33.8645 | 0.0387 | 566.2664 |
| world-rust-composed-http / wasmvm / go / native / 2w | 5.3970 | 0.0650 | 0.0003 | 217.9878 | 0.0450 | 0.3609 | 0.0006 | N/A | 0.0127 | 18.2813 | 17.2068 | 0.0453 | 400.9090 |
| world-rust-composed-http / wasmvm / go / native / 4w | 5.3162 | 0.0628 | 0.0003 | 219.7478 | 0.0552 | 0.6579 | 0.0005 | N/A | 0.0127 | 10.0150 | 9.4892 | 0.0564 | 322.9407 |
| world-rust-composed-http / wasmvm / llgo / interpreter / 1w | 8.8425 | 0.0700 | 0.0002 | 44.5131 | 0.1206 | 0.3922 | 0.0005 | N/A | 0.2724 | 80.3119 | 77.2671 | 0.2011 | 834.0938 |
| world-rust-composed-http / wasmvm / llgo / interpreter / 2w | 8.7785 | 0.0605 | 0.0002 | 43.1447 | 0.1540 | 0.7966 | 0.0007 | N/A | 0.2969 | 52.7035 | 47.7715 | 0.2958 | 538.3314 |
| world-rust-composed-http / wasmvm / llgo / interpreter / 4w | 8.8989 | 0.0693 | 0.0002 | 44.3611 | 0.2476 | 1.5259 | 0.0009 | N/A | 0.3259 | 48.5824 | 44.8794 | 0.5229 | 506.0395 |
| world-rust-composed-http / wasmvm / llgo / native / 1w | 8.8930 | 0.0771 | 0.0002 | 326.8160 | 0.1040 | 0.2977 | 0.0005 | N/A | 0.3132 | 70.7211 | 69.9452 | 0.2077 | 1039.5807 |
| world-rust-composed-http / wasmvm / llgo / native / 2w | 8.7736 | 0.0716 | 0.0002 | 327.2878 | 0.1517 | 0.5164 | 0.0007 | N/A | 0.2998 | 46.1646 | 45.6084 | 0.2864 | 798.4917 |
| world-rust-composed-http / wasmvm / llgo / native / 4w | 9.0575 | 0.0764 | 0.0003 | 326.5147 | 0.2645 | 0.9783 | 0.0008 | N/A | 0.3024 | 44.6520 | 46.1300 | 0.4745 | 806.4386 |
| world-rust-http / wasmtime / go / compiler / 1w | 5.7732 | 0.0555 | 0.1337 | 50.6266 | 0.1148 | 0.1234 | 0.0067 | 0.0296 | 0.0349 | 4.2460 | 3.7396 | 0.1620 | 95.8904 |
| world-rust-http / wasmtime / go / compiler / 2w | 5.9310 | 0.0580 | 0.1293 | 50.3823 | 0.1631 | 0.1526 | 0.0073 | 0.0355 | 0.0484 | 2.4223 | 2.0315 | 0.1895 | 79.1398 |
| world-rust-http / wasmtime / go / compiler / 4w | 5.9368 | 0.0539 | 0.1466 | 50.7222 | 0.2307 | 0.1932 | 0.0080 | 0.0392 | 0.0533 | 1.3953 | 1.0861 | 0.2277 | 70.1447 |
| world-rust-http / wasmvm / go / interpreter / 1w | 5.4865 | 0.0554 | 0.0003 | 17.9308 | 0.0377 | 0.1620 | 0.0004 | N/A | 0.0126 | 44.8648 | 40.9424 | 0.0299 | 440.0867 |
| world-rust-http / wasmvm / go / interpreter / 2w | 5.5298 | 0.0563 | 0.0003 | 17.9423 | 0.0485 | 0.4133 | 0.0005 | N/A | 0.0170 | 23.1865 | 20.4318 | 0.0326 | 234.5875 |
| world-rust-http / wasmvm / go / interpreter / 4w | 5.5251 | 0.0560 | 0.0003 | 18.1703 | 0.0683 | 0.9187 | 0.0008 | N/A | 0.0170 | 12.3741 | 10.5896 | 0.0720 | 135.4500 |
| world-rust-http / wasmvm / go / native / 1w | 5.5024 | 0.0610 | 0.0004 | 150.2766 | 0.0340 | 0.1621 | 0.0003 | N/A | 0.0124 | 32.8337 | 31.4803 | 0.0360 | 476.3118 |
| world-rust-http / wasmvm / go / native / 2w | 5.5873 | 0.0521 | 0.0003 | 151.0130 | 0.0462 | 0.2770 | 0.0005 | N/A | 0.0122 | 17.4736 | 16.3360 | 0.0395 | 324.3909 |
| world-rust-http / wasmvm / go / native / 4w | 5.4824 | 0.0603 | 0.0003 | 151.7600 | 0.0552 | 0.4915 | 0.0007 | N/A | 0.0145 | 9.2174 | 9.1013 | 0.0532 | 254.1556 |
| world-rust-http / wasmvm / llgo / interpreter / 1w | 8.8260 | 0.0637 | 0.0002 | 28.9967 | 0.1061 | 2.6911 | 0.0005 | N/A | 0.2975 | 75.0033 | 74.3192 | 0.2205 | 788.6878 |
| world-rust-http / wasmvm / llgo / interpreter / 2w | 8.9240 | 0.0573 | 0.0002 | 29.0112 | 0.1691 | 3.5891 | 0.0008 | N/A | 0.2885 | 49.5246 | 46.4125 | 0.2985 | 509.1958 |
| world-rust-http / wasmvm / llgo / interpreter / 4w | 8.8591 | 0.0595 | 0.0003 | 29.6797 | 0.2652 | 3.8993 | 0.0010 | N/A | 0.3020 | 45.9122 | 45.4939 | 0.4350 | 496.6608 |
| world-rust-http / wasmvm / llgo / native / 1w | 8.7408 | 0.0614 | 0.0002 | 224.6070 | 0.1143 | 0.2175 | 0.0005 | N/A | 0.2963 | 68.8323 | 66.8886 | 0.1999 | 910.5763 |
| world-rust-http / wasmvm / llgo / native / 2w | 8.9177 | 0.0724 | 0.0002 | 225.5611 | 0.1563 | 0.3732 | 0.0007 | N/A | 0.3311 | 46.4825 | 44.0080 | 0.2700 | 682.6081 |
| world-rust-http / wasmvm / llgo / native / 4w | 8.8854 | 0.0680 | 0.0002 | 224.7742 | 0.2680 | 0.7559 | 0.0010 | N/A | 0.3318 | 44.5097 | 46.0064 | 0.4197 | 694.6484 |

## Every identical-binary control

Paired-block warm-batch log-ratio bootstrap,8,000 draws; each95% interval must fit inside[0.975,1.025]. Controls diagnose the sampled throughput interval, not every independent API or tail metric. Requests inside one process are not independent host runs.

| Profile | A/A lower | A/A upper | Resolved |
| --- | ---: | ---: | --- |
| native-go / native-go / go / host / 1w | 0.95195 | 1.01878 | no |
| native-go / native-go / go / host / 2w | 1.01852 | 1.11493 | no |
| native-go / native-go / go / host / 4w | 0.83167 | 1.03107 | no |
| native-go / native-go / llgo / host / 1w | 0.91188 | 1.06451 | no |
| native-go / native-go / llgo / host / 2w | 0.81577 | 0.96633 | no |
| native-go / native-go / llgo / host / 4w | 0.85485 | 0.98674 | no |
| world-rust-composed-http / wasmtime / go / compiler / 1w | 0.98481 | 1.01025 | yes |
| world-rust-composed-http / wasmtime / go / compiler / 2w | 0.99466 | 1.01019 | yes |
| world-rust-composed-http / wasmtime / go / compiler / 4w | 0.98371 | 1.17299 | no |
| world-rust-composed-http / wasmvm / go / interpreter / 1w | 0.96430 | 1.01913 | no |
| world-rust-composed-http / wasmvm / go / interpreter / 2w | 0.98942 | 1.01253 | yes |
| world-rust-composed-http / wasmvm / go / interpreter / 4w | 0.95955 | 1.01190 | no |
| world-rust-composed-http / wasmvm / go / native / 1w | 0.96723 | 0.99357 | no |
| world-rust-composed-http / wasmvm / go / native / 2w | 0.99135 | 1.02544 | no |
| world-rust-composed-http / wasmvm / go / native / 4w | 0.98355 | 1.01284 | yes |
| world-rust-composed-http / wasmvm / llgo / interpreter / 1w | 0.99067 | 1.06871 | no |
| world-rust-composed-http / wasmvm / llgo / interpreter / 2w | 0.99927 | 1.01227 | yes |
| world-rust-composed-http / wasmvm / llgo / interpreter / 4w | 0.95611 | 0.99669 | no |
| world-rust-composed-http / wasmvm / llgo / native / 1w | 0.97031 | 1.04595 | no |
| world-rust-composed-http / wasmvm / llgo / native / 2w | 1.00732 | 1.06821 | no |
| world-rust-composed-http / wasmvm / llgo / native / 4w | 0.88050 | 1.10546 | no |
| world-rust-http / wasmtime / go / compiler / 1w | 0.94117 | 1.01150 | no |
| world-rust-http / wasmtime / go / compiler / 2w | 0.94277 | 1.02878 | no |
| world-rust-http / wasmtime / go / compiler / 4w | 0.92769 | 1.03113 | no |
| world-rust-http / wasmvm / go / interpreter / 1w | 0.96767 | 1.01785 | no |
| world-rust-http / wasmvm / go / interpreter / 2w | 0.97988 | 1.03148 | no |
| world-rust-http / wasmvm / go / interpreter / 4w | 0.95786 | 1.00502 | no |
| world-rust-http / wasmvm / go / native / 1w | 1.00198 | 1.06162 | no |
| world-rust-http / wasmvm / go / native / 2w | 0.98322 | 1.01431 | yes |
| world-rust-http / wasmvm / go / native / 4w | 0.98651 | 1.02596 | no |
| world-rust-http / wasmvm / llgo / interpreter / 1w | 0.97939 | 1.09230 | no |
| world-rust-http / wasmvm / llgo / interpreter / 2w | 0.97096 | 1.04066 | no |
| world-rust-http / wasmvm / llgo / interpreter / 4w | 0.97285 | 1.08414 | no |
| world-rust-http / wasmvm / llgo / native / 1w | 0.98136 | 1.11202 | no |
| world-rust-http / wasmvm / llgo / native / 2w | 0.96339 | 1.01572 | no |
| world-rust-http / wasmvm / llgo / native / 4w | 0.91479 | 1.01146 | no |

## Remaining work

P3 HTTP is materially slower in these point estimates and needs architectural investigation. Native and interpreter results are close under LLGo, indicating that improving guest instruction throughput alone cannot explain the full application gap; this is an inference, not an isolated-cost subtraction. The bare host control already differs substantially between Go/LLGo. Profile the owned component/IO boundary and host allocation/scheduling separately. Do not infer that the entire gap is caused by the known upstream channel issue.

The actual independent reference executes on Linuxamd64 and macOSarm64; Linux validation completes in4m52s with95.34% Go adapter coverage. Owned local Go race coverage is95.2%; pinned LLGo actually executes both owned backends and the bare host control. Go adapter coverage is not C/Rust/VM/specification coverage. Whole-project final-head semantic CI and controlled release acceptance remain required.

Initial Rust opaque-future lifetime, boxed-error conversion, dylib LINKEDIT and disabled-network-hook build failures were corrected before timing, retaining all failed logs and leaving module caches untouched. The final isolated Rust static crate uses locked dependencies and explicit denied outgoing network. Initial loader audit reuse lacked a P3 key and excluded framework images; only audit metadata/parser changed before five successful loader probes. All executable/source/count/profile identities remain fixed.

Raw arrays, exact source bundle/archive, five build identities, static-archive/compiler hashes, actual loader probes, qualification outputs and independent reductions are retained in `artifacts/p3io-service-20261110`; raw result SHA256 `7115b5666b21edf79f9b793f271bfd07b7aee28ff79e15d36d8b1c55def2bcfe`. Existing Core/native/GC/Threads gaps, other required acceptance evidence and production release remain open.
