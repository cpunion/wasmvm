# Native indirect calls, 2026-11-11

[Complete machine record](data/performance-native-indirect-calls-2026-11-11.json).

## Bottleneck and implementation

The small HTTP guest made59 same-instance indirect calls per request through the shared Go dispatcher. Each return to the host also revalidated and rebound native execution state. Counter diagnostics over20000 requests per compiler observe approximately317 native entries per request before the change and186 after it, a41.3% reduction. Core public execution boundaries remain approximately93 and component tasks remain9 per request. Counters establish the mechanism independently of throughput. The initial62-call progress estimate was corrected to59 after classifying actual targets.

Both native tiers now resolve ordinary same-instance `call_indirect` and `call_ref` through immutable compiler-generated dispatch rows. Each lookup checks the live unsigned table index or function reference, instance function identity and exact validated structural type. Every emitted function shares one32-byte-per-definition table across its indirect call sites. Frame setup, argument/result copies, continuations, native returns and runtime frame adoption reuse the existing linked-call machinery. Definition identities are checked after the range lookup. Allocation layouts with different identity ordering retain the checked shared dispatch path.

Nulls, mismatched types, imported and cross-instance targets, valid structural subtypes, tail calls and exhausted native frame storage retain the common checked driver. Logical depth/slot quotas, exact metered fuel, complete128-bit values, roots, cancellation and lifecycle contracts remain enforced. Dynamic frame zeroing polls cancellation for every slot. Inline-leaf misses can enter the same native linked path. The existing module code budget bounds all dispatch data and linked bodies.

Exclusive wall-region diagnostics pointed to Core execution and native entry as the largest measured region. Nested hooks add overhead and wall intervals include scheduling. These diagnostic region shares carry their own scope. The uninstrumented fixed-source service event below supplies the latency/throughput observations.

## HTTP fixture and independent controls

All paths use four workers,16 clients and64 requests/client. Throughput uses five primary process-block reductions. API intervals cover the buffered complete response and joined guest call. Full response bytes, status, headers and trailers are checked. Ordinary host applications provide the corresponding Go/LLGo application baseline. Wasmtime48.0.1 supplies the independent P3 runtime comparison.

### 16-byte responses

| Path | Base req/s | Head req/s | Base API us | Head API us |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 18214 | 18254 | 218.067 | 217.437 |
| wasmvm go / native | 22254 | 25871 | 178.355 | 152.983 |
| wasmvm llgo / interpreter | 6299 | 6758 | 627.979 | 585.590 |
| wasmvm llgo / native | 6598 | 6246 | 599.282 | 633.311 |
| Wasmtime48.0.1 | N/A | 74362 | N/A | 52.541 |
| Ordinary Go application | N/A | 1798840 | N/A | 0.588 |
| Ordinary LLGo application | N/A | 116366 | N/A | 3.931 |
### 65536-byte responses

| Path | Base req/s | Head req/s | Base API us | Head API us |
| --- | ---: | ---: | ---: | ---: |
| wasmvm go / interpreter | 12344 | 12318 | 321.028 | 321.509 |
| wasmvm go / native | 16224 | 18169 | 243.327 | 216.501 |
| wasmvm llgo / interpreter | 4622 | 5062 | 850.905 | 776.502 |
| wasmvm llgo / native | 4440 | 4588 | 886.249 | 857.839 |
| Wasmtime48.0.1 | N/A | 62307 | N/A | 61.787 |
| Ordinary Go application | N/A | 40155 | N/A | 88.328 |
| Ordinary LLGo application | N/A | 9346 | N/A | 400.629 |

1/40 A/A controls resolve within±2.5%. 0/16 source pairs resolve both controls. All40 profiles, adverse observations and unresolved comparisons remain below. Workstation background workloads are recorded in the machine record. Performance and release acceptance remain open.

## Source comparisons

| Head profile | Base req/s | Head req/s | Head/base warm time [95%] | Both A/A |
| --- | ---: | ---: | --- | --- |
| head/world-rust-composed-http/go/interpreter/4/16 | 15694 | 15061 | 1.0432 [1.0052,1.0819] | unresolved |
| head/world-rust-composed-http/go/interpreter/4/65536 | 11199 | 10525 | 1.0367 [1.0026,1.0719] | unresolved |
| head/world-rust-composed-http/go/native/4/16 | 17827 | 21066 | 0.8518 [0.8106,0.8950] | unresolved |
| head/world-rust-composed-http/go/native/4/65536 | 14702 | 15617 | 0.9463 [0.9302,0.9627] | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 5848 | 5447 | 1.0143 [0.9477,1.0838] | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4558 | 4168 | 1.1251 [0.9343,1.3963] | unresolved |
| head/world-rust-composed-http/llgo/native/4/16 | 5861 | 5828 | 1.0038 [0.9875,1.0187] | unresolved |
| head/world-rust-composed-http/llgo/native/4/65536 | 4397 | 4835 | 0.9469 [0.8941,1.0027] | unresolved |
| head/world-rust-http/go/interpreter/4/16 | 18214 | 18254 | 0.9952 [0.9342,1.0485] | unresolved |
| head/world-rust-http/go/interpreter/4/65536 | 12344 | 12318 | 1.0018 [0.9379,1.0559] | unresolved |
| head/world-rust-http/go/native/4/16 | 22254 | 25871 | 0.8486 [0.7943,0.8929] | unresolved |
| head/world-rust-http/go/native/4/65536 | 16224 | 18169 | 0.8837 [0.8787,0.8897] | unresolved |
| head/world-rust-http/llgo/interpreter/4/16 | 6299 | 6758 | 0.9598 [0.8006,1.1314] | unresolved |
| head/world-rust-http/llgo/interpreter/4/65536 | 4622 | 5062 | 0.9728 [0.9138,1.0264] | unresolved |
| head/world-rust-http/llgo/native/4/16 | 6598 | 6246 | 1.0573 [0.9557,1.1697] | unresolved |
| head/world-rust-http/llgo/native/4/65536 | 4440 | 4588 | 0.9514 [0.8218,1.0931] | unresolved |

## Go and LLGo gap

Identical fixture/backend pairs use the same source and request counts. Ratios above1 mean more time under LLGo. The current compiler pin remains3d92a58. The717fbe442 compiler event and LLGo issue3045 retain their separate evidence.

| LLGo profile | LLGo/Go warm time [95%] | LLGo/Go API time [95%] | Both A/A |
| --- | --- | --- | --- |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 2.5923 [2.4648,2.7360] | 2.5864 [2.4579,2.7305] | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 2.7319 [2.3238,3.4695] | 2.7104 [2.3129,3.4254] | unresolved |
| head/world-rust-composed-http/llgo/native/4/16 | 3.5733 [3.3765,3.7816] | 3.5726 [3.3756,3.7810] | unresolved |
| head/world-rust-composed-http/llgo/native/4/65536 | 3.2748 [3.2394,3.3107] | 3.2669 [3.2301,3.3041] | unresolved |
| head/world-rust-http/llgo/interpreter/4/16 | 2.7502 [2.6096,2.9310] | 2.7383 [2.5987,2.9175] | unresolved |
| head/world-rust-http/llgo/interpreter/4/65536 | 2.5269 [2.3025,2.7732] | 2.5093 [2.2865,2.7538] | unresolved |
| head/world-rust-http/llgo/native/4/16 | 4.1219 [3.8344,4.4309] | 4.1166 [3.8289,4.4259] | unresolved |
| head/world-rust-http/llgo/native/4/65536 | 3.9936 [3.7490,4.3152] | 3.9908 [3.7497,4.3059] | unresolved |

## Competitor and ordinary native gaps

Ratios above1 mean more owned-runtime warm time. Ordinary applications execute their direct host implementation. Their functionality and compiler accounting are recorded separately from component execution.

| Head profile | Reference | Head/reference warm time [95%] | Both A/A |
| --- | --- | --- | --- |
| head/world-rust-composed-http/go/interpreter/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 4.0209 [3.7304,4.4457] | unresolved |
| head/world-rust-composed-http/go/interpreter/4/16 | host/native-go/go/host/4/16 | 125.0167 [119.4370,130.4546] | unresolved |
| head/world-rust-composed-http/go/interpreter/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 5.1280 [4.8827,5.3712] | unresolved |
| head/world-rust-composed-http/go/interpreter/4/65536 | host/native-go/go/host/4/65536 | 3.7728 [3.5498,4.0044] | unresolved |
| head/world-rust-composed-http/go/native/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 2.7886 [2.6068,2.9073] | unresolved |
| head/world-rust-composed-http/go/native/4/16 | host/native-go/go/host/4/16 | 86.7038 [81.3340,92.4282] | unresolved |
| head/world-rust-composed-http/go/native/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 3.5452 [3.2545,3.7593] | unresolved |
| head/world-rust-composed-http/go/native/4/65536 | host/native-go/go/host/4/65536 | 2.6083 [2.3714,2.8689] | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 10.4234 [9.6255,11.3330] | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/16 | host/native-go/llgo/host/4/16 | 20.0616 [18.7218,21.4349] | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 14.0089 [11.8721,17.1369] | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | host/native-go/llgo/host/4/65536 | 2.3860 [1.9577,3.1085] | unresolved |
| head/world-rust-composed-http/llgo/native/4/16 | reference/world-rust-composed-http/go/compiler/4/16 | 9.9647 [9.5476,10.5054] | unresolved |
| head/world-rust-composed-http/llgo/native/4/16 | host/native-go/llgo/host/4/16 | 19.1789 [17.8486,20.6170] | unresolved |
| head/world-rust-composed-http/llgo/native/4/65536 | reference/world-rust-composed-http/go/compiler/4/65536 | 11.6101 [10.6105,12.4234] | unresolved |
| head/world-rust-composed-http/llgo/native/4/65536 | host/native-go/llgo/host/4/65536 | 1.9775 [1.8826,2.1039] | unresolved |
| head/world-rust-http/go/interpreter/4/16 | reference/world-rust-http/go/compiler/4/16 | 3.8801 [3.4824,4.2825] | unresolved |
| head/world-rust-http/go/interpreter/4/16 | host/native-go/go/host/4/16 | 102.7744 [99.4053,106.2576] | unresolved |
| head/world-rust-http/go/interpreter/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 5.1459 [4.8328,5.5050] | unresolved |
| head/world-rust-http/go/interpreter/4/65536 | host/native-go/go/host/4/65536 | 3.3176 [3.1123,3.5390] | unresolved |
| head/world-rust-http/go/native/4/16 | reference/world-rust-http/go/compiler/4/16 | 2.7159 [2.5274,2.9184] | unresolved |
| head/world-rust-http/go/native/4/16 | host/native-go/go/host/4/16 | 71.9358 [69.3602,74.6070] | unresolved |
| head/world-rust-http/go/native/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 3.4552 [3.2814,3.6221] | unresolved |
| head/world-rust-http/go/native/4/65536 | host/native-go/go/host/4/65536 | 2.2276 [2.0470,2.4241] | unresolved |
| head/world-rust-http/llgo/interpreter/4/16 | reference/world-rust-http/go/compiler/4/16 | 10.6710 [9.2653,12.4576] | unresolved |
| head/world-rust-http/llgo/interpreter/4/16 | host/native-go/llgo/host/4/16 | 17.4967 [16.0700,19.0501] | unresolved |
| head/world-rust-http/llgo/interpreter/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 13.0033 [12.2150,13.7931] | unresolved |
| head/world-rust-http/llgo/interpreter/4/65536 | host/native-go/llgo/host/4/65536 | 1.9408 [1.7983,2.0945] | unresolved |
| head/world-rust-http/llgo/native/4/16 | reference/world-rust-http/go/compiler/4/16 | 11.1944 [10.4917,12.3421] | unresolved |
| head/world-rust-http/llgo/native/4/16 | host/native-go/llgo/host/4/16 | 18.3549 [17.4113,19.0874] | unresolved |
| head/world-rust-http/llgo/native/4/65536 | reference/world-rust-http/go/compiler/4/65536 | 13.7987 [12.9636,14.7004] | unresolved |
| head/world-rust-http/llgo/native/4/65536 | host/native-go/llgo/host/4/65536 | 2.0595 [1.9158,2.2230] | unresolved |

## Complete profile and resource results

Runtime byte accounting follows each compiler. Go counters cover Go-managed allocation. LLGo malloc counts are unavailable. Separate resource processes collect between-batch threads and peak RSS.

| Profile | req/s | API us | p99 request us | Runtime B/request | Go mallocs/request | Threads | Peak RSS MiB | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| base/world-rust-composed-http/go/interpreter/4/16 | 15694 | 253.058 | 2044.417 | 96624 | 1328.0 | 7 | 129.531 | yes |
| base/world-rust-composed-http/go/interpreter/4/65536 | 11199 | 353.952 | 3287.791 | 256465 | 1712.0 | 7 | 146.406 | unresolved |
| base/world-rust-composed-http/go/native/4/16 | 17827 | 222.438 | 2012.917 | 92496 | 1334.0 | 7 | 158.719 | unresolved |
| base/world-rust-composed-http/go/native/4/65536 | 14702 | 268.712 | 3147.458 | 252337 | 1718.0 | 8 | 161.906 | unresolved |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 5848 | 677.609 | 8657.667 | 169615 | N/A | 57 | 165.625 | unresolved |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 4558 | 865.301 | 7493.417 | 416516 | N/A | 57 | 169.797 | unresolved |
| base/world-rust-composed-http/llgo/native/4/16 | 5861 | 675.200 | 6295.125 | 165330 | N/A | 57 | 212.750 | unresolved |
| base/world-rust-composed-http/llgo/native/4/65536 | 4397 | 896.408 | 9527.625 | 412228 | N/A | 57 | 209.938 | unresolved |
| base/world-rust-http/go/interpreter/4/16 | 18214 | 218.067 | 1796.750 | 82113 | 1204.0 | 8 | 97.531 | unresolved |
| base/world-rust-http/go/interpreter/4/65536 | 12344 | 321.028 | 2701.209 | 241953 | 1588.0 | 7 | 105.734 | unresolved |
| base/world-rust-http/go/native/4/16 | 22254 | 178.355 | 1582.541 | 82256 | 1207.0 | 7 | 119.234 | unresolved |
| base/world-rust-http/go/native/4/65536 | 16224 | 243.327 | 2437.625 | 242097 | 1591.0 | 7 | 123.547 | unresolved |
| base/world-rust-http/llgo/interpreter/4/16 | 6299 | 627.979 | 5624.709 | 150509 | N/A | 57 | 140.719 | unresolved |
| base/world-rust-http/llgo/interpreter/4/65536 | 4622 | 850.905 | 6715.292 | 397415 | N/A | 57 | 140.891 | unresolved |
| base/world-rust-http/llgo/native/4/16 | 6598 | 599.282 | 5516.125 | 150951 | N/A | 57 | 175.969 | unresolved |
| base/world-rust-http/llgo/native/4/65536 | 4440 | 886.249 | 9882.750 | 397850 | N/A | 57 | 176.859 | unresolved |
| head/world-rust-composed-http/go/interpreter/4/16 | 15061 | 263.269 | 2688.333 | 96625 | 1328.0 | 7 | 134.453 | unresolved |
| head/world-rust-composed-http/go/interpreter/4/65536 | 10525 | 377.009 | 3657.458 | 256465 | 1712.0 | 7 | 147.438 | unresolved |
| head/world-rust-composed-http/go/native/4/16 | 21066 | 188.403 | 1766.125 | 92497 | 1334.0 | 8 | 162.000 | unresolved |
| head/world-rust-composed-http/go/native/4/65536 | 15617 | 252.996 | 2930.375 | 252336 | 1718.0 | 7 | 165.078 | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 5447 | 727.712 | 6484.459 | 169620 | N/A | 57 | 165.766 | unresolved |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 4168 | 946.497 | 10221.750 | 416510 | N/A | 57 | 181.781 | unresolved |
| head/world-rust-composed-http/llgo/native/4/16 | 5828 | 678.931 | 6278.750 | 165339 | N/A | 57 | 214.359 | unresolved |
| head/world-rust-composed-http/llgo/native/4/65536 | 4835 | 815.396 | 6670.917 | 412219 | N/A | 57 | 210.844 | unresolved |
| head/world-rust-http/go/interpreter/4/16 | 18254 | 217.437 | 1778.000 | 82113 | 1204.0 | 7 | 94.469 | unresolved |
| head/world-rust-http/go/interpreter/4/65536 | 12318 | 321.509 | 2741.000 | 241952 | 1588.0 | 7 | 104.625 | unresolved |
| head/world-rust-http/go/native/4/16 | 25871 | 152.983 | 1593.500 | 82257 | 1207.0 | 7 | 123.234 | unresolved |
| head/world-rust-http/go/native/4/65536 | 18169 | 216.501 | 2399.209 | 242097 | 1591.0 | 7 | 129.656 | unresolved |
| head/world-rust-http/llgo/interpreter/4/16 | 6758 | 585.590 | 5460.084 | 150518 | N/A | 57 | 139.047 | unresolved |
| head/world-rust-http/llgo/interpreter/4/65536 | 5062 | 776.502 | 6294.292 | 397409 | N/A | 57 | 140.938 | unresolved |
| head/world-rust-http/llgo/native/4/16 | 6246 | 633.311 | 5668.917 | 150969 | N/A | 57 | 161.172 | unresolved |
| head/world-rust-http/llgo/native/4/65536 | 4588 | 857.839 | 9307.125 | 397860 | N/A | 57 | 176.984 | unresolved |
| host/native-go/go/host/4/16 | 1798840 | 0.588 | 158.291 | 2713 | 18.0 | 6 | 15.938 | unresolved |
| host/native-go/go/host/4/65536 | 40155 | 88.328 | 734.083 | 277915 | 48.0 | 7 | 23.781 | unresolved |
| host/native-go/llgo/host/4/16 | 116366 | 3.931 | 331.375 | 3416 | N/A | 40 | 19.797 | unresolved |
| host/native-go/llgo/host/4/65536 | 9346 | 400.629 | 3720.167 | 732501 | N/A | 40 | 68.578 | unresolved |
| reference/world-rust-composed-http/go/compiler/4/16 | 61504 | 64.011 | 4127.292 | 32 | 1.0 | 10 | 33.219 | unresolved |
| reference/world-rust-composed-http/go/compiler/4/65536 | 56335 | 68.629 | 4597.667 | 32 | 1.0 | 10 | 34.172 | unresolved |
| reference/world-rust-http/go/compiler/4/16 | 74362 | 52.541 | 3468.208 | 32 | 1.0 | 7 | 30.688 | unresolved |
| reference/world-rust-http/go/compiler/4/65536 | 62307 | 61.787 | 4066.584 | 32 | 1.0 | 7 | 30.875 | unresolved |

## Lifecycle phases

Median milliseconds across primary process blocks. Compile/validate includes Wasm-to-native compilation on native backends. Readiness, loading, engine/provider setup, instantiation, first call, warm execution and joined cleanup retain separate intervals. Host compiler build intervals are retained in the local build manifest. Process wall also includes measurement-protocol parking.

| Profile | Ready | Load | Engine | Compile | Providers | Instantiate | Lookup | Executor | Workers | First batch | Warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| base/world-rust-composed-http/go/interpreter/4/16 | 6.5510 | 0.0780 | 0.0005 | 27.3981 | 0.0605 | 0.6960 | 0.0006 | N/A | 0.0220 | 68.9877 | 65.2475 | 0.0658 | 693.5006 |
| base/world-rust-composed-http/go/interpreter/4/65536 | 7.1135 | 0.0718 | 0.0004 | 26.5113 | 0.0587 | 0.6756 | 0.0007 | N/A | 0.0200 | 94.7875 | 91.4347 | 0.0595 | 956.2354 |
| base/world-rust-composed-http/go/native/4/16 | 6.6734 | 0.0667 | 0.0003 | 223.1130 | 0.0516 | 0.6338 | 0.0007 | N/A | 0.0164 | 55.6133 | 57.4413 | 0.0654 | 811.4272 |
| base/world-rust-composed-http/go/native/4/65536 | 6.6405 | 0.0678 | 0.0004 | 221.6900 | 0.0564 | 0.6384 | 0.0007 | N/A | 0.0169 | 71.9670 | 69.6518 | 0.0563 | 929.6855 |
| base/world-rust-composed-http/llgo/interpreter/4/16 | 10.4604 | 0.0825 | 0.0003 | 45.3029 | 0.2317 | 1.5623 | 0.0010 | N/A | 0.3590 | 200.0738 | 175.0950 | 0.5695 | 1849.4908 |
| base/world-rust-composed-http/llgo/interpreter/4/65536 | 10.7055 | 0.0722 | 0.0003 | 43.7950 | 0.2447 | 1.5917 | 0.0011 | N/A | 0.3236 | 221.1986 | 224.6650 | 0.5365 | 2265.3427 |
| base/world-rust-composed-http/llgo/native/4/16 | 10.3513 | 0.0783 | 0.0003 | 322.2333 | 0.2081 | 0.9498 | 0.0009 | N/A | 0.2937 | 161.7185 | 174.7160 | 0.6595 | 2045.4527 |
| base/world-rust-composed-http/llgo/native/4/65536 | 10.3744 | 0.0764 | 0.0003 | 325.8732 | 0.2320 | 0.9730 | 0.0012 | N/A | 0.3123 | 209.8065 | 232.8954 | 0.6015 | 2617.6940 |
| base/world-rust-http/go/interpreter/4/16 | 6.8172 | 0.0655 | 0.0004 | 18.4870 | 0.0570 | 0.4844 | 0.0007 | N/A | 0.0163 | 59.7497 | 56.2213 | 0.0485 | 592.1263 |
| base/world-rust-http/go/interpreter/4/65536 | 6.5476 | 0.0600 | 0.0004 | 18.1363 | 0.0727 | 0.9613 | 0.0009 | N/A | 0.0146 | 88.0624 | 82.9569 | 0.0465 | 864.8281 |
| base/world-rust-http/go/native/4/16 | 6.6853 | 0.0673 | 0.0004 | 151.3296 | 0.0484 | 0.4777 | 0.0006 | N/A | 0.0151 | 47.0411 | 46.0134 | 0.0638 | 624.5379 |
| base/world-rust-http/go/native/4/65536 | 6.6597 | 0.0623 | 0.0005 | 151.7403 | 0.0564 | 0.4610 | 0.0006 | N/A | 0.0156 | 66.7457 | 63.1173 | 0.0559 | 795.1228 |
| base/world-rust-http/llgo/interpreter/4/16 | 10.5327 | 0.0707 | 0.0003 | 29.4829 | 0.2542 | 3.4611 | 0.0013 | N/A | 0.3040 | 165.7510 | 162.5596 | 0.5920 | 1652.1394 |
| base/world-rust-http/llgo/interpreter/4/65536 | 10.2900 | 0.0711 | 0.0002 | 30.4632 | 0.2740 | 4.2662 | 0.0010 | N/A | 0.3118 | 213.7820 | 221.5477 | 0.5857 | 2328.4853 |
| base/world-rust-http/llgo/native/4/16 | 9.9060 | 0.0895 | 0.0003 | 227.1917 | 0.2004 | 0.7249 | 0.0008 | N/A | 0.2799 | 150.3211 | 155.2065 | 0.5535 | 1781.7963 |
| base/world-rust-http/llgo/native/4/65536 | 10.6224 | 0.0717 | 0.0003 | 228.9872 | 0.2116 | 0.7515 | 0.0012 | N/A | 0.3350 | 231.5256 | 230.6052 | 0.5190 | 2458.8506 |
| head/world-rust-composed-http/go/interpreter/4/16 | 6.4756 | 0.0743 | 0.0004 | 26.3668 | 0.0583 | 0.7238 | 0.0007 | N/A | 0.0237 | 70.5088 | 67.9910 | 0.0561 | 713.6409 |
| head/world-rust-composed-http/go/interpreter/4/65536 | 6.5877 | 0.0740 | 0.0003 | 26.3542 | 0.0571 | 0.6686 | 0.0006 | N/A | 0.0243 | 98.4827 | 97.2960 | 0.0537 | 1009.8100 |
| head/world-rust-composed-http/go/native/4/16 | 7.0955 | 0.0863 | 0.0004 | 237.7896 | 0.0533 | 0.6718 | 0.0005 | N/A | 0.0166 | 47.2375 | 48.6094 | 0.0693 | 722.1519 |
| head/world-rust-composed-http/go/native/4/65536 | 7.2304 | 0.0809 | 0.0003 | 237.2886 | 0.0518 | 0.6384 | 0.0006 | N/A | 0.0185 | 66.0273 | 65.5689 | 0.0599 | 900.4800 |
| head/world-rust-composed-http/llgo/interpreter/4/16 | 10.0499 | 0.0703 | 0.0003 | 44.0772 | 0.2128 | 1.5858 | 0.0010 | N/A | 0.2804 | 178.8181 | 187.9832 | 0.5695 | 1898.3512 |
| head/world-rust-composed-http/llgo/interpreter/4/65536 | 10.0942 | 0.0728 | 0.0002 | 45.5748 | 0.2537 | 1.6330 | 0.0011 | N/A | 0.3648 | 234.3881 | 245.6817 | 0.5142 | 2470.6840 |
| head/world-rust-composed-http/llgo/native/4/16 | 10.2829 | 0.0830 | 0.0003 | 344.1113 | 0.2245 | 0.9910 | 0.0013 | N/A | 0.3111 | 173.7124 | 175.7049 | 0.5315 | 2145.8064 |
| head/world-rust-composed-http/llgo/native/4/65536 | 10.3117 | 0.0785 | 0.0003 | 343.6171 | 0.2191 | 1.0044 | 0.0010 | N/A | 0.2729 | 226.7457 | 211.8011 | 0.4880 | 2519.0758 |
| head/world-rust-http/go/interpreter/4/16 | 6.8503 | 0.0633 | 0.0004 | 18.2348 | 0.0694 | 0.7720 | 0.0009 | N/A | 0.0161 | 58.9835 | 56.0970 | 0.0467 | 591.6285 |
| head/world-rust-http/go/interpreter/4/65536 | 6.6650 | 0.0621 | 0.0003 | 18.2912 | 0.0532 | 0.5150 | 0.0006 | N/A | 0.0153 | 84.8316 | 83.1279 | 0.0439 | 861.7036 |
| head/world-rust-http/go/native/4/16 | 6.9849 | 0.0656 | 0.0004 | 164.0109 | 0.0497 | 0.4634 | 0.0005 | N/A | 0.0192 | 39.4980 | 39.5815 | 0.0563 | 570.7925 |
| head/world-rust-http/go/native/4/65536 | 7.0489 | 0.0624 | 0.0004 | 161.1690 | 0.0549 | 0.4775 | 0.0007 | N/A | 0.0192 | 58.0430 | 56.3583 | 0.0553 | 738.1270 |
| head/world-rust-http/llgo/interpreter/4/16 | 9.9355 | 0.0647 | 0.0002 | 29.9633 | 0.2452 | 3.2186 | 0.0012 | N/A | 0.3049 | 149.7923 | 151.5146 | 0.5133 | 1536.7771 |
| head/world-rust-http/llgo/interpreter/4/65536 | 10.1608 | 0.0750 | 0.0003 | 29.8350 | 0.2326 | 3.6592 | 0.0010 | N/A | 0.3056 | 224.6885 | 202.2900 | 0.5583 | 2068.7727 |
| head/world-rust-http/llgo/native/4/16 | 10.8488 | 0.0805 | 0.0003 | 247.0172 | 0.2278 | 0.6887 | 0.0008 | N/A | 0.3495 | 163.4087 | 163.9529 | 0.5295 | 1857.1777 |
| head/world-rust-http/llgo/native/4/65536 | 9.9537 | 0.0757 | 0.0003 | 244.0952 | 0.2052 | 0.7135 | 0.0008 | N/A | 0.3176 | 207.1511 | 223.1768 | 0.5555 | 2471.8742 |
| host/native-go/go/host/4/16 | 6.2001 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0188 | 0.6673 | 0.5693 | 0.0062 | 13.3516 |
| host/native-go/go/host/4/65536 | 5.9875 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.0165 | 25.2026 | 25.5013 | 0.0067 | 262.9243 |
| host/native-go/llgo/host/4/16 | 10.2767 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2698 | 7.9998 | 8.7999 | 0.1075 | 97.5447 |
| host/native-go/llgo/host/4/65536 | 9.8793 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | 0.2675 | 108.8408 | 109.5691 | 0.1082 | 1106.0941 |
| reference/world-rust-composed-http/go/compiler/4/16 | 7.2341 | 0.0685 | 0.1343 | 73.4588 | 0.2175 | 0.2513 | 0.0081 | 0.0412 | 0.0404 | 18.3371 | 16.6493 | 0.3764 | 251.2851 |
| reference/world-rust-composed-http/go/compiler/4/65536 | 7.1551 | 0.0690 | 0.1388 | 74.2390 | 0.2205 | 0.2447 | 0.0080 | 0.0380 | 0.0492 | 19.6333 | 18.1771 | 0.3852 | 269.8906 |
| reference/world-rust-http/go/compiler/4/16 | 7.6057 | 0.0673 | 0.1424 | 51.2496 | 0.2329 | 0.1973 | 0.0084 | 0.0400 | 0.0463 | 14.9228 | 13.7706 | 0.2509 | 200.2285 |
| reference/world-rust-http/go/compiler/4/65536 | 6.9400 | 0.0622 | 0.1319 | 52.3014 | 0.2258 | 0.1882 | 0.0078 | 0.0377 | 0.0373 | 17.1214 | 16.4348 | 0.2372 | 230.8835 |

## Protocol and validation

Base`ea413c30091e306c9373af83bdd0e4451d1e015d`, timed implementation`0aaab1adfdb5a14d88a108bea06017fe0698a675`. Go1.27.3, LLGo3d92a58 with`-O3 -lto=full`, LLVM22.1.8 and Wasmtime48.0.1 remain pinned. Base executable source hashes match the retained uninstrumented baseline build. Every observation in this event is newly executed.

One fixed five-block primary/A/A event, seed2026111141, pointwise95% paired-block log-ratio bootstrap with8000 draws. Both Rust plain/composed HTTP guests, both owned backends/compilers,16-byte/64-KiB bodies and all independent controls participate. Exactly400 throughput processes,40 resource processes,80 qualifications and7 loader probes complete. The event checks2867200 warm responses. Each process uses2 warmups and7 measured batches. All participants use identical counts.

Final-source local checks pass three affected Go packages, five pinned-LLGo packages, three Go race packages and vet. Regression tests exercise both native tiers, distant targets, live retargeting, imports, cross-instance targets, empty/wrong-type entries, logical quotas, mixed12-value tuples, host boundaries and exact fuel agreement with the interpreter. Full final-head CI checks all32 root packages across Go/LLGo and Linux amd64/macOS arm64, all8 LLGo shards and31 integration coverage gates. Nativeexec-only local coverage is85%, reflecting its isolated test callers. The full integration coverage gate remains90%.

Every task-owned build/test/profile/probe exits before primary timing. Process/cwd inspection records externally owned builds/tests. Physical isolation, affinity and frequency control are outside this workstation event. Earlier timing cohorts retain their independent records.

The remaining93 Core execution boundaries and9 tasks/request warrant measured component/Canonical ABI and host-runtime investigation. Go/LLGo concurrency/allocation cost, remaining Wasmtime and ordinary native gaps, compiler adoption, controlled performance and production release remain open. Broader Core runtime comparisons with wazero and Wasmer retain their dated evidence. [Service portfolio](performance-service-all-runtimes-2026-11-10.md), [compiler study](performance-llgo-merged-2026-11-10.md).
