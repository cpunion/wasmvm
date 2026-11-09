# Bounded P3 transfer buffers: 2026-11-10

[Complete machine record](data/performance-p3-io-buffers-2026-11-10.json).

M4 Max/macOS 27.0.1; base `825d588287b75f0563a72f1127be7d9cab34f92d`, head `162c27bf65cbcb20304778749ed56d4a731271ca`. Both use Go 1.27.3 and pinned LLGo `3d92a58` with `-O3 -lto=full`. The base already includes immutable memory copies and bounded worker reuse. This independent event changes reusable transfer buffers and packed Canonical stream-buffer writes. The standard-library pipe remains in production. The [rejected custom-pipe experiment](performance-p3-body-pipe-evaluation-2026-11-10.md), including its adverse cells, remains separate.

**Observed result:** complete 64 KiB HTTP, four workers, LLGo native throughput 2131 -> 2248 req/s; complete API 1851.2 -> 1753.6 us. Paired warm-time ratio 0.9669 [0.9479,0.9873], pointwise 95%; both controls resolve for this comparison. Runtime-accounted bytes/request fall 1201984 -> 938628 (about 257 KiB). Go native observes 7670 -> 8324 req/s and 740415 -> 478269 B/request. All 16 owned point estimates improve; 15 intervals exclude 1, while LLGo native/one-worker HTTP remains unresolved, 0.9838 [0.9686,1.0008]. Twelve of 16 comparison control pairs resolve, and 32/40 individual A/A controls resolve. These are advisory workstation observations, not controlled release acceptance.

**The large gap remains:** the same event measures Wasmtime 59662 req/s, bare LLGo host 8657 req/s and bare Go host 46319 req/s. Four-worker paired warm-time ratios are 27.1221 [26.3782,27.8906] against Wasmtime (both controls resolve), and 3.8845 [3.7436,4.0290] against the LLGo host control (controls unresolved). Native/interpreter LLGo medians are 2248/2235 req/s. This points to investigating component/provider and host-runtime paths before assuming guest instruction generation explains the full deficit. It does not identify a pure cost by subtraction: the reference uses a current-thread Tokio executor and different host bindings, and the ordinary native-Go body-collection application is not an optimized native lower bound.

**Resource tradeoff:** four-worker LLGo native HTTP samples 61 -> 61 total between-batch OS threads and 176.1 -> 191.9 MiB observed peak RSS. Lower accounted allocation does not imply lower observed RSS; all resource cells are retained below. Go managed/LLGo collector counters have different scopes; Go reference counters omit Rust/JIT allocation. Snapshots do not classify OS thread scheduler states.

## Mechanism and contracts

Each accepted operation exclusively owns a mutable transfer buffer until a reader has copied the result to an immutable string or an `io.Writer` has consumed it. The IO owner retains at most `MaxIdleWorkers` buffers, each at most the larger configured read/write bound. Read slices expose only the requested capacity. Read results are snapshotted before a deferred return can make the buffer available to another Store; no completion retains mutable pooled bytes. Close drops the cache and late returns cannot repopulate it. Buffer sizes, body limits and worker quotas remain bounded by the existing configuration, without workload-specific tuning.

Packed Canonical `StoreBuffer` now shares the existing `storeBytes` immutable-memory capability. It retains range checks, byte/value charges, the generic fallback, and checks cancellation after the memory write. The existing component task scheduler and standard `io.Pipe` backpressure/close/error behavior remain in use. No compiler patch, new LLGo issue, competitor fallback or unsafe borrowed guest memory is involved.

## One fixed complete event

Twenty randomized process blocks; primary/A/A aliases; both Rust HTTP and composed fixtures; Go/LLGo interpreter/native; one/four workers; 16 clients, four requests/client; two warmups and seven measured batches; 64 KiB payload with independently checked CRC32 2629362236. Order/bootstrap seed 2026111007; 8000 paired log-ratio bootstrap draws; pointwise 95% intervals; an A/A interval must lie entirely inside [0.975,1.025] to resolve. All 1600 throughput processes, 40 separate resource processes, 80 qualification processes and seven actual loader probes complete, with 716800 checked warm responses. Every response byte, status 201, header and trailer is checked. No dropped adverse result, favorable retiming, phase subtraction, family-wise claim or pooling with earlier events.

Sources, archives, toolchains, binaries, fixtures and actual loaded-library probes are checked before sampling. Four owned executables are fresh base/head builds; three frozen reference/control executables retain their original source/build/library provenance. All owned builds/tests/probes/archives exited first, and the actual pre-start process table contains no compiler/test process. The earlier separate Go CPU/allocation diagnostic ended before this event. Affinity, frequency, background GUI work and power remain uncontrolled; GOMAXPROCS=4 is not a native CPU quota.

Preparation, first/warm batches, complete API, request latency, readiness and joined cleanup are separate. Response collection/close are inside API time; full validation follows it and remains inside request wall. Owned request construction is outside API time; reference request construction is inside its Go C/Rust bridge. No phase subtraction infers unmeasured overhead.

## Every owned comparison

| Fixture / compiler / backend / workers | Base req/s | Head req/s | Warm head/base [95%] | Both A/A resolve | Base B/req | Head B/req |
| --- | ---: | ---: | --- | --- | ---: | ---: |
| world-rust-composed-http/go/interpreter/1 | 1597 | 1621 | 0.9846 [0.9790,0.9904] | yes | 755662 | 493513 |
| world-rust-composed-http/go/interpreter/4 | 6271 | 6506 | 0.9587 [0.9510,0.9655] | yes | 755636 | 493484 |
| world-rust-composed-http/go/native/1 | 2080 | 2146 | 0.9723 [0.9670,0.9784] | yes | 752060 | 489909 |
| world-rust-composed-http/go/native/4 | 7448 | 8067 | 0.9240 [0.9116,0.9346] | no | 752018 | 489873 |
| world-rust-composed-http/llgo/interpreter/1 | 1035 | 1058 | 0.9750 [0.9682,0.9814] | yes | 1222219 | 958945 |
| world-rust-composed-http/llgo/interpreter/4 | 2135 | 2218 | 0.9632 [0.9484,0.9780] | yes | 1222351 | 959012 |
| world-rust-composed-http/llgo/native/1 | 1168 | 1197 | 0.9817 [0.9676,0.9984] | no | 1218756 | 955463 |
| world-rust-composed-http/llgo/native/4 | 2105 | 2193 | 0.9658 [0.9485,0.9836] | no | 1219000 | 955668 |
| world-rust-http/go/interpreter/1 | 1691 | 1718 | 0.9882 [0.9820,0.9951] | yes | 739799 | 477648 |
| world-rust-http/go/interpreter/4 | 6524 | 6999 | 0.9269 [0.9203,0.9335] | yes | 739757 | 477603 |
| world-rust-http/go/native/1 | 2238 | 2289 | 0.9805 [0.9757,0.9858] | yes | 740455 | 478291 |
| world-rust-http/go/native/4 | 7670 | 8324 | 0.9234 [0.9109,0.9378] | yes | 740415 | 478269 |
| world-rust-http/llgo/interpreter/1 | 1081 | 1111 | 0.9770 [0.9678,0.9859] | yes | 1200493 | 937173 |
| world-rust-http/llgo/interpreter/4 | 2139 | 2235 | 0.9558 [0.9460,0.9664] | no | 1200655 | 937309 |
| world-rust-http/llgo/native/1 | 1223 | 1253 | 0.9838 [0.9686,1.0008] | yes | 1201791 | 938471 |
| world-rust-http/llgo/native/4 | 2131 | 2248 | 0.9669 [0.9479,0.9873] | yes | 1201984 | 938628 |

## All profiles and controls

| Profile | req/s | Complete API us | Request p99 us | Runtime B/req | Between-batch threads | Peak RSS MiB | A/A 95% | Resolved |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| base/world-rust-composed-http/go/interpreter/1 | 1597 | 623.554 | 11023.208 | 755662 | 7 | 104.672 | [0.9956,1.0069] | yes |
| base/world-rust-composed-http/go/interpreter/4 | 6271 | 617.807 | 4491.917 | 755636 | 7 | 147.969 | [0.9902,1.0066] | yes |
| base/world-rust-composed-http/go/native/1 | 2080 | 478.371 | 8529.354 | 752060 | 8 | 146.359 | [0.9971,1.0117] | yes |
| base/world-rust-composed-http/go/native/4 | 7448 | 517.526 | 4052.875 | 752018 | 8 | 157.828 | [0.9798,1.0128] | yes |
| base/world-rust-composed-http/llgo/interpreter/1 | 1035 | 957.210 | 19035.125 | 1222219 | 43 | 161.641 | [0.9920,1.0059] | yes |
| base/world-rust-composed-http/llgo/interpreter/4 | 2135 | 1847.342 | 10650.583 | 1222351 | 61 | 181.219 | [0.9913,1.0193] | yes |
| base/world-rust-composed-http/llgo/native/1 | 1168 | 849.828 | 16483.395 | 1218756 | 43 | 181.688 | [0.9743,1.0047] | no |
| base/world-rust-composed-http/llgo/native/4 | 2105 | 1871.480 | 10437.333 | 1219000 | 61 | 208.469 | [0.9748,1.0253] | no |
| base/world-rust-http/go/interpreter/1 | 1691 | 588.950 | 10104.334 | 739799 | 7 | 80.938 | [0.9961,1.0091] | yes |
| base/world-rust-http/go/interpreter/4 | 6524 | 596.410 | 3913.750 | 739757 | 7 | 104.656 | [0.9885,1.0019] | yes |
| base/world-rust-http/go/native/1 | 2238 | 444.234 | 7707.354 | 740455 | 7 | 119.922 | [0.9965,1.0090] | yes |
| base/world-rust-http/go/native/4 | 7670 | 505.374 | 3554.666 | 740415 | 8 | 121.656 | [0.9895,1.0126] | yes |
| base/world-rust-http/llgo/interpreter/1 | 1081 | 918.368 | 17513.730 | 1200493 | 43 | 140.422 | [0.9879,1.0075] | yes |
| base/world-rust-http/llgo/interpreter/4 | 2139 | 1843.197 | 9924.334 | 1200655 | 61 | 140.641 | [0.9885,1.0146] | yes |
| base/world-rust-http/llgo/native/1 | 1223 | 809.569 | 15882.584 | 1201791 | 43 | 167.859 | [0.9847,1.0169] | yes |
| base/world-rust-http/llgo/native/4 | 2131 | 1851.188 | 9786.771 | 1201984 | 61 | 176.125 | [0.9863,1.0242] | yes |
| head/world-rust-composed-http/go/interpreter/1 | 1621 | 614.450 | 11003.604 | 493513 | 7 | 104.016 | [0.9898,1.0047] | yes |
| head/world-rust-composed-http/go/interpreter/4 | 6506 | 594.370 | 4419.542 | 493484 | 8 | 137.859 | [0.9985,1.0085] | yes |
| head/world-rust-composed-http/go/native/1 | 2146 | 463.443 | 8226.834 | 489909 | 7 | 141.062 | [0.9891,1.0051] | yes |
| head/world-rust-composed-http/go/native/4 | 8067 | 477.299 | 3838.000 | 489873 | 7 | 160.781 | [0.9949,1.0292] | no |
| head/world-rust-composed-http/llgo/interpreter/1 | 1058 | 936.576 | 18521.354 | 958945 | 43 | 164.250 | [0.9911,1.0045] | yes |
| head/world-rust-composed-http/llgo/interpreter/4 | 2218 | 1775.986 | 10819.646 | 959012 | 61 | 169.438 | [0.9911,1.0173] | yes |
| head/world-rust-composed-http/llgo/native/1 | 1197 | 825.751 | 16693.229 | 955463 | 43 | 196.703 | [0.9787,1.0104] | yes |
| head/world-rust-composed-http/llgo/native/4 | 2193 | 1785.524 | 10335.083 | 955668 | 61 | 209.484 | [0.9799,1.0146] | yes |
| head/world-rust-http/go/interpreter/1 | 1718 | 579.633 | 10158.917 | 477648 | 7 | 79.844 | [0.9880,0.9995] | yes |
| head/world-rust-http/go/interpreter/4 | 6999 | 553.215 | 3757.791 | 477603 | 7 | 103.750 | [0.9935,1.0035] | yes |
| head/world-rust-http/go/native/1 | 2289 | 434.295 | 7559.604 | 478291 | 7 | 118.375 | [0.9922,1.0025] | yes |
| head/world-rust-http/go/native/4 | 8324 | 465.328 | 3348.188 | 478269 | 7 | 121.578 | [0.9780,1.0052] | yes |
| head/world-rust-http/llgo/interpreter/1 | 1111 | 892.030 | 17410.104 | 937173 | 43 | 140.484 | [0.9900,1.0113] | yes |
| head/world-rust-http/llgo/interpreter/4 | 2235 | 1762.241 | 9750.229 | 937309 | 61 | 141.109 | [0.9735,1.0052] | no |
| head/world-rust-http/llgo/native/1 | 1253 | 788.449 | 15312.000 | 938471 | 43 | 168.281 | [0.9889,1.0088] | yes |
| head/world-rust-http/llgo/native/4 | 2248 | 1753.617 | 9560.188 | 938628 | 61 | 191.891 | [0.9793,1.0197] | yes |
| host/native-go/go/host/1 | 41164 | 20.654 | 522.729 | 277915 | 6 | 16.859 | [0.9835,1.0098] | yes |
| host/native-go/go/host/4 | 46319 | 74.934 | 542.167 | 277917 | 6 | 16.312 | [0.9908,1.0559] | no |
| host/native-go/llgo/host/1 | 8441 | 112.705 | 3673.771 | 732582 | 37 | 68.094 | [0.9969,1.0473] | no |
| host/native-go/llgo/host/4 | 8657 | 422.502 | 3834.854 | 732572 | 40 | 68.438 | [0.9188,1.0314] | no |
| reference/world-rust-composed-http/go/compiler/1 | 15352 | 62.932 | 1145.500 | 32 | 7 | 32.547 | [0.9903,1.0033] | yes |
| reference/world-rust-composed-http/go/compiler/4 | 51888 | 72.474 | 366.354 | 32 | 7 | 33.266 | [0.9739,1.0050] | no |
| reference/world-rust-http/go/compiler/1 | 17833 | 53.877 | 1008.375 | 32 | 7 | 29.484 | [0.9984,1.0248] | yes |
| reference/world-rust-http/go/compiler/4 | 59662 | 62.662 | 332.771 | 32 | 8 | 30.781 | [0.9903,1.0212] | yes |

## Complete lifecycle intervals

Median milliseconds over 20 primary process reductions. Warm is a process-mean batch. The machine record also retains every process reduction, resource snapshot and all 32 paired reference intervals. Phases are not combined to infer an unmeasured cost.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/1 | 5.4584 | 0.0608 | 0.0003 | 25.7710 | 0.0365 | 0.2266 | 0.0004 | N/A | 0.0130 | 44.2330 | 40.0661 | 0.0309 | 441.0871 |
| base/world-rust-composed-http/go/interpreter/4 | 5.4330 | 0.0582 | 0.0003 | 25.7552 | 0.0533 | 0.6841 | 0.0006 | N/A | 0.0153 | 12.5194 | 10.2064 | 0.0512 | 141.4574 |
| base/world-rust-composed-http/go/native/1 | 5.4142 | 0.0573 | 0.0003 | 214.1229 | 0.0342 | 0.2060 | 0.0004 | N/A | 0.0121 | 32.1852 | 30.7757 | 0.0361 | 532.5484 |
| base/world-rust-composed-http/go/native/4 | 5.3292 | 0.0581 | 0.0003 | 213.5296 | 0.0503 | 0.6426 | 0.0006 | N/A | 0.0129 | 9.6617 | 8.5932 | 0.0615 | 310.4300 |
| base/world-rust-composed-http/llgo/interpreter/1 | 8.7862 | 0.0682 | 0.0002 | 44.3125 | 0.1172 | 0.4274 | 0.0005 | N/A | 0.2781 | 64.4062 | 61.8534 | 0.2446 | 680.0858 |
| base/world-rust-composed-http/llgo/interpreter/4 | 8.7404 | 0.0720 | 0.0002 | 44.3918 | 0.2410 | 1.4901 | 0.0011 | N/A | 0.3115 | 36.1046 | 29.9778 | 0.6187 | 363.1636 |
| base/world-rust-composed-http/llgo/native/1 | 8.7355 | 0.0681 | 0.0002 | 325.4219 | 0.1060 | 0.2825 | 0.0005 | N/A | 0.2830 | 55.8640 | 54.8096 | 0.2211 | 886.2088 |
| base/world-rust-composed-http/llgo/native/4 | 8.7384 | 0.0695 | 0.0002 | 323.4443 | 0.2416 | 0.9618 | 0.0009 | N/A | 0.3023 | 31.3005 | 30.3985 | 0.6112 | 640.3935 |
| base/world-rust-http/go/interpreter/1 | 5.4688 | 0.0525 | 0.0003 | 17.9378 | 0.0374 | 0.1569 | 0.0004 | N/A | 0.0134 | 41.4667 | 37.8493 | 0.0296 | 408.7106 |
| base/world-rust-http/go/interpreter/4 | 5.3021 | 0.0510 | 0.0003 | 17.8299 | 0.0640 | 0.9087 | 0.0009 | N/A | 0.0162 | 11.5689 | 9.8093 | 0.0432 | 127.7326 |
| base/world-rust-http/go/native/1 | 5.3829 | 0.0525 | 0.0003 | 147.1896 | 0.0322 | 0.1597 | 0.0004 | N/A | 0.0123 | 29.9055 | 28.5915 | 0.0353 | 442.8405 |
| base/world-rust-http/go/native/4 | 5.4059 | 0.0531 | 0.0003 | 146.8753 | 0.0479 | 0.4763 | 0.0007 | N/A | 0.0132 | 8.4251 | 8.3440 | 0.0536 | 240.1702 |
| base/world-rust-http/llgo/interpreter/1 | 8.8614 | 0.0623 | 0.0002 | 29.7674 | 0.1059 | 3.1538 | 0.0005 | N/A | 0.2955 | 61.1172 | 59.1960 | 0.2304 | 641.0660 |
| base/world-rust-http/llgo/interpreter/4 | 8.7733 | 0.0633 | 0.0002 | 29.4091 | 0.2468 | 3.7005 | 0.0010 | N/A | 0.3054 | 32.2171 | 29.9141 | 0.5905 | 346.9102 |
| base/world-rust-http/llgo/native/1 | 8.7793 | 0.0601 | 0.0002 | 224.8568 | 0.1076 | 0.2181 | 0.0005 | N/A | 0.2788 | 53.7384 | 52.3253 | 0.2313 | 763.3762 |
| base/world-rust-http/llgo/native/4 | 8.7256 | 0.0615 | 0.0002 | 224.1390 | 0.2377 | 0.7371 | 0.0010 | N/A | 0.3178 | 31.2161 | 30.0370 | 0.6237 | 540.7494 |
| head/world-rust-composed-http/go/interpreter/1 | 5.3945 | 0.0559 | 0.0003 | 25.8564 | 0.0374 | 0.2283 | 0.0004 | N/A | 0.0127 | 43.8215 | 39.4782 | 0.0308 | 434.6234 |
| head/world-rust-composed-http/go/interpreter/4 | 5.3750 | 0.0558 | 0.0003 | 25.5831 | 0.0547 | 0.6750 | 0.0005 | N/A | 0.0153 | 12.1135 | 9.8375 | 0.0490 | 135.2642 |
| head/world-rust-composed-http/go/native/1 | 5.3771 | 0.0593 | 0.0003 | 214.3669 | 0.0354 | 0.2069 | 0.0003 | N/A | 0.0119 | 31.4822 | 29.8173 | 0.0369 | 524.7288 |
| head/world-rust-composed-http/go/native/4 | 5.4470 | 0.0605 | 0.0003 | 214.2155 | 0.0522 | 0.6413 | 0.0006 | N/A | 0.0132 | 8.7655 | 7.9333 | 0.0611 | 304.4166 |
| head/world-rust-composed-http/llgo/interpreter/1 | 8.8225 | 0.0670 | 0.0002 | 44.0572 | 0.1229 | 0.4381 | 0.0004 | N/A | 0.2827 | 63.0052 | 60.4911 | 0.2469 | 666.0658 |
| head/world-rust-composed-http/llgo/interpreter/4 | 8.8648 | 0.0674 | 0.0002 | 44.9632 | 0.2411 | 1.4880 | 0.0009 | N/A | 0.3031 | 31.5794 | 28.8499 | 0.5806 | 351.4238 |
| head/world-rust-composed-http/llgo/native/1 | 8.8061 | 0.0684 | 0.0002 | 323.7364 | 0.1090 | 0.2843 | 0.0005 | N/A | 0.2829 | 53.9523 | 53.4525 | 0.2444 | 871.2864 |
| head/world-rust-composed-http/llgo/native/4 | 8.7666 | 0.0676 | 0.0002 | 325.0628 | 0.2311 | 0.9444 | 0.0009 | N/A | 0.3043 | 29.8627 | 29.1799 | 0.6140 | 632.9004 |
| head/world-rust-http/go/interpreter/1 | 5.4190 | 0.0538 | 0.0003 | 17.9214 | 0.0385 | 0.1626 | 0.0004 | N/A | 0.0131 | 41.0369 | 37.2547 | 0.0295 | 401.2910 |
| head/world-rust-http/go/interpreter/4 | 5.3991 | 0.0524 | 0.0003 | 17.9249 | 0.0556 | 0.4932 | 0.0007 | N/A | 0.0132 | 10.8990 | 9.1443 | 0.0446 | 120.6657 |
| head/world-rust-http/go/native/1 | 5.4092 | 0.0502 | 0.0003 | 147.4360 | 0.0339 | 0.1573 | 0.0003 | N/A | 0.0123 | 29.8242 | 27.9609 | 0.0358 | 437.4133 |
| head/world-rust-http/go/native/4 | 5.3404 | 0.0513 | 0.0003 | 147.8862 | 0.0494 | 0.4674 | 0.0006 | N/A | 0.0133 | 8.5021 | 7.6889 | 0.0533 | 233.9610 |
| head/world-rust-http/llgo/interpreter/1 | 8.7964 | 0.0576 | 0.0002 | 29.5505 | 0.1072 | 3.0516 | 0.0005 | N/A | 0.2747 | 61.1244 | 57.6081 | 0.2406 | 623.4246 |
| head/world-rust-http/llgo/interpreter/4 | 8.9110 | 0.0631 | 0.0002 | 29.6045 | 0.2361 | 3.4896 | 0.0010 | N/A | 0.2996 | 29.7496 | 28.6305 | 0.5814 | 334.4878 |
| head/world-rust-http/llgo/native/1 | 8.8503 | 0.0629 | 0.0002 | 224.0551 | 0.1049 | 0.2142 | 0.0005 | N/A | 0.2745 | 53.4134 | 51.0593 | 0.2218 | 748.6127 |
| head/world-rust-http/llgo/native/4 | 8.7473 | 0.0579 | 0.0002 | 224.3552 | 0.2349 | 0.7199 | 0.0010 | N/A | 0.3062 | 30.6487 | 28.4696 | 0.6032 | 527.5407 |
| host/native-go/go/host/1 | 4.8208 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0132 | 1.7890 | 1.5548 | 0.0051 | 22.1064 |
| host/native-go/go/host/4 | 5.0274 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0149 | 1.3096 | 1.3817 | 0.0071 | 20.1767 |
| host/native-go/llgo/host/1 | 8.5612 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2772 | 9.3606 | 7.5821 | 0.0817 | 87.7314 |
| host/native-go/llgo/host/4 | 8.5465 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.3142 | 10.5426 | 7.3926 | 0.0885 | 88.7144 |
| reference/world-rust-composed-http/go/compiler/1 | 5.7736 | 0.0601 | 0.1298 | 70.9957 | 0.1129 | 0.1350 | 0.0063 | 0.0279 | 0.0353 | 4.6100 | 4.1687 | 0.2177 | 121.2853 |
| reference/world-rust-composed-http/go/compiler/4 | 5.7904 | 0.0615 | 0.1288 | 71.4783 | 0.2172 | 0.2454 | 0.0077 | 0.0409 | 0.0453 | 1.5478 | 1.2334 | 0.3200 | 92.3512 |
| reference/world-rust-http/go/compiler/1 | 5.6771 | 0.0535 | 0.1320 | 49.3763 | 0.1125 | 0.1193 | 0.0061 | 0.0269 | 0.0329 | 4.0199 | 3.5889 | 0.1470 | 93.3330 |
| reference/world-rust-http/go/compiler/4 | 5.8195 | 0.0521 | 0.1305 | 49.3515 | 0.2198 | 0.1909 | 0.0077 | 0.0409 | 0.0473 | 1.3514 | 1.0727 | 0.2079 | 68.6985 |

## Validation and remaining work

Full normal local Go tests pass with all 31 measured owned packages >=90% (minimum 90.83%). Canonical/P3 race tests, P3 measurement race tests, vet, pinned LLGo complete Canonical/P3 package execution and P3 measurement execution on both backends pass. Tests exercise exclusive buffer ownership, snapshot immutability, bounded retention/growth/oversized rejection, empty/error/partial pipe transfers, close/late return, packed-buffer budgets and cancellation during either memory adapter. A shared World executes concurrent distinct invalid-UTF8 body sequences through two independent Stores under both backends, checking complete bytes and trailers after repeated reuse. Final-head full Go/LLGo CI on Linux amd64 and macOS arm64 remains required before merge.

The raw result SHA256 is `74c04ad27f4d278319e4cf4b69cf4615d73eee9f2c527d020f35a1a665de28a7` in `artifacts/p3-io-buffers-20261110`. The unchanged timed controller, source bundle, source/compiler/binary identities, all observations, checks and postprocessing audit remain retained. Earlier experiments and their regressions are independently linked, never pooled. Production performance acceptance remains open: identify the dominant component/provider synchronization and host-call costs, close compiler gaps, validate on both controlled targets, then complete release acceptance. LLGo #3034/#3035 are being solved upstream; this contribution does not duplicate them or upgrade the pinned compiler.
