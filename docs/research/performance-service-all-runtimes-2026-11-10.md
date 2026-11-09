# Complete concurrent runtime comparison: 2026-11-10

[Machine record](data/performance-service-all-runtimes-2026-11-10.json).
Measured on M4 Max / macOS27.0.1. The five named runtime families and the native-C
control use the same admission driver. This is a new complete event; earlier
local and Linux observations are not pooled with it.

35/72 throughput A/A controls resolve. **All six owned LLGo native controls remain
unresolved**, so the following point estimates do not establish a stable native
ranking or close performance acceptance. The full tables preserve every profile,
including slower results and both termination policies.

For the C-produced guest at four workers (64 SHA blocks/request), req/s are:

| Engine / host | req/s | Complete API µs/request |
| --- | ---: | ---: |
| wasmvm native / LLGo | 135706 | 16.980 |
| wasmvm native / Go | 263847 | 14.391 |
| wazero compiler / Go | 221635 | 17.268 |
| Wasmtime Cranelift / Go C bridge | 304512 | 12.501 |
| Wasmer Cranelift / Go C bridge | 300650 | 12.626 |
| Native C / LLGo | 133541 | 12.376 |
| Native C / Go | 318452 | 11.887 |

The VM/native-compiler and host-runtime gaps require different investigations.
The native-C control already has a 2.38x Go/LLGo throughput difference while its
complete API intervals are similar. This demonstrates a service gap without a
VM; it does not identify one scheduler, collector or bridge cause. The existing
standalone channel-allocation finding remains upstream work, not a new duplicate
issue. Owned LLGo native API time is also longer than the native-C and two
Cranelift reference intervals, so host scheduling cannot explain every gap.
These complete API intervals include their documented safety/interop policies;
no phase subtraction produces a pure guest-only cost.

Interpreter results tell a different story on this workload: at one worker,
owned LLGo delivers 853req/s, wazero378 and WasmEdge586. These are workload-specific
observations with separate controls, not a universal interpreter ranking. Native
compiled execution remains much faster than interpretation on this SHA workload.

Both owned Go/LLGo reference builds and all three SDK Go adapters execute their
actual tests. SDK adapter Go coverage is 96.52% each on both targets. Linux
validation took6m38s and matches exact source, all three executable hashes and
142 retained pinned SDK files. This does not claim C, runtime or specification
coverage. User-confirmed LLGo3034/3035 fixes are being handled upstream; the
benchmark compiler/profile remain pinned for future independent fix validation.

P3 I/O, remaining compiler/native performance work, controlled acceptance and
production release remain open. The next P3 comparison must include response
body/trailer/completion lifetimes rather than timing only returned headers.

**Separate M4 observational event; no controlled performance/release gate closes.** Production engine files equal PR #145 head `e02fe34` (digest `8dcda8f29f8703f5549e110b2ec6ca0d380bf65168917756d22e4d541a7b2ce7`). Only the common benchmark harness and independent reference adapters differ.

The same bounded queue, independent worker state, all-clients-ready broadcast gate and raw timers drive owned Go/LLGo, wazero 1.12 interpreter/compiler (default and enabled context termination separately), Wasmtime48.0.1 Cranelift speed, Wasmer7.3.0 Cranelift, WasmEdge0.17.1 interpreter through independent Go C bridges, and ordinary C SHA-256 through the Go/LLGo bridge. SDKs use their default unmetered policy and one independent Store/executor per worker, sharing immutable prepared module/AST code. Each SDK binary links exactly one SDK. Bridge and parameter/result packing remain inside its timed callback; checksum validation follows it. Owned callbacks retain the existing typed Function.Call boundary; wazero uses a private reused CallWithStack buffer. These API/storage policies remain explicit; the prior scalar-stack boundary event is separate. Native C has no Wasm safety/fuel/cancellation/JIT/Store contract and is one algorithm control, not a separate Rust-native producer.

Pinned Go 1.27.3, LLGo 3d92a58 `-O3 -lto=full`, LLVM/Clang 22.1.8, native C `-O3 -std=gnu11`; root-module x/sys 0.49 is retained. Actual dyld-loaded libraries, both source producers, fixtures, compiler/source/binary hashes and every child exit/checksum are independently audited. Reference dependency checks exclude the owned decoder/compiler/VM.

Five fresh process blocks, identical-binary primary/A/A aliases, 2 warmup + 7 measured batches, seed 17, 64 message blocks/request, 64 clients, 16 requests/client and 1/2/4 workers. **720 successful throughput processes, 72 separate resource processes, 216 two-seed/resource qualification processes, 5,160,960 checked warm requests; A/A resolves 35/72 profiles.** All builds/tests/probes/archives finish before sampling. Raw arrays retain every batch, request, API interval, pause and adverse result; no retiming, subtraction or cross-cohort pooling.

Medians of five independent per-process means. Request p50/p99 are first reduced within each process from all 7 warm batches (7,168 requests). Threads/RSS are actual idle snapshots from separate first+one-measured-batch processes; peak RSS is through the final pre-cleanup snapshot, not owned memory or completed-process acceptance. GOMAXPROCS=4 is recorded, not equated to a native CPU quota.

### Native throughput overview

| Wasm producer / workers | Own Go req/s | Own LLGo req/s | Wazero compiler default req/s | Wazero compiler termination req/s | Wasmtime req/s | Wasmer req/s | Native C / Go req/s | Native C / LLGo req/s |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang / 1w | 61189 | 58334 | 47504 | 6924 | 67328 | 65826 | 67704 | 67915 |
| sha256-clang / 2w | 119813 | 100680 | 105363 | 13910 | 134498 | 132457 | 139269 | 117540 |
| sha256-clang / 4w | 263847 | 135706 | 221635 | 27369 | 304512 | 300650 | 318452 | 133541 |
| sha256-rust / 1w | 58635 | 55955 | 42922 | 7016 | 66096 | 64551 | 67704 | 67915 |
| sha256-rust / 2w | 115723 | 102149 | 93013 | 14122 | 133234 | 132001 | 139269 | 117540 |
| sha256-rust / 4w | 257228 | 135156 | 204371 | 29089 | 304661 | 295533 | 318452 | 133541 |

### Interpreter throughput overview

| Producer / workers | Own Go req/s | Own LLGo req/s | Wazero default req/s | Wazero termination req/s | WasmEdge req/s |
| --- | ---: | ---: | ---: | ---: | ---: |
| sha256-clang / 1w | 737 | 853 | 378 | 380 | 586 |
| sha256-clang / 2w | 1475 | 1718 | 748 | 770 | 1164 |
| sha256-clang / 4w | 2929 | 3411 | 1529 | 1509 | 2348 |
| sha256-rust / 1w | 759 | 872 | 373 | 380 | 630 |
| sha256-rust / 2w | 1519 | 1736 | 754 | 757 | 1260 |
| sha256-rust / 4w | 3020 | 3492 | 1470 | 1506 | 2516 |

### All 72 profiles: throughput, latency, allocation and separate resources

Allocation counters count Go managed bytes versus LLGo BDWGC collector bytes. They include service scheduling/calls inside the batch; they are not pure VM or all native/linear-memory allocation counts. LLGo Mallocs is unavailable.

| Profile | req/s | Request p50 µs | Request p99 µs | Mean API µs | Runtime B/req | Go mallocs/req | Idle threads max | Idle RSS MiB max | Peak RSS MiB |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang / wasmedge / go / interpreter / 1w | 586 | 108836.833 | 112034.792 | 1706.254 | 0.05 | 0.000 | 7 | 27.469 | 27.609 |
| sha256-clang / wasmedge / go / interpreter / 2w | 1164 | 54902.209 | 56081.375 | 1715.851 | 0.09 | 0.001 | 7 | 27.641 | 27.672 |
| sha256-clang / wasmedge / go / interpreter / 4w | 2348 | 26036.917 | 40646.125 | 1698.808 | 0.03 | 0.000 | 10 | 27.875 | 27.938 |
| sha256-clang / wasmer / go / compiler / 1w | 65826 | 972.958 | 1035.541 | 13.941 | 0.03 | 0.000 | 22 | 20.922 | 20.969 |
| sha256-clang / wasmer / go / compiler / 2w | 132457 | 480.792 | 521.458 | 14.309 | 0.78 | 0.001 | 22 | 21.078 | 21.188 |
| sha256-clang / wasmer / go / compiler / 4w | 300650 | 12.666 | 2723.292 | 12.626 | 0.00 | 0.000 | 22 | 21.172 | 21.219 |
| sha256-clang / wasmtime / go / compiler / 1w | 67328 | 948.458 | 1008.000 | 13.726 | 0.02 | 0.000 | 23 | 16.828 | 16.875 |
| sha256-clang / wasmtime / go / compiler / 2w | 134498 | 475.375 | 504.709 | 14.144 | 0.00 | 0.000 | 23 | 16.500 | 16.531 |
| sha256-clang / wasmtime / go / compiler / 4w | 304512 | 12.500 | 2687.333 | 12.501 | 0.00 | 0.000 | 23 | 16.781 | 16.828 |
| sha256-clang / wasmvm / go / interpreter / 1w | 737 | 86529.208 | 89562.083 | 1355.480 | 64.03 | 2.000 | 6 | 8.578 | 8.625 |
| sha256-clang / wasmvm / go / interpreter / 2w | 1475 | 43314.375 | 44873.792 | 1353.331 | 64.06 | 2.001 | 7 | 8.531 | 8.609 |
| sha256-clang / wasmvm / go / interpreter / 4w | 2929 | 12507.125 | 77504.625 | 1352.435 | 64.00 | 2.000 | 7 | 8.844 | 8.891 |
| sha256-clang / wasmvm / go / native / 1w | 61189 | 1042.583 | 1096.792 | 15.291 | 64.05 | 2.000 | 6 | 11.625 | 11.656 |
| sha256-clang / wasmvm / go / native / 2w | 119813 | 510.167 | 640.750 | 15.871 | 64.78 | 2.001 | 6 | 11.938 | 11.938 |
| sha256-clang / wasmvm / go / native / 4w | 263847 | 13.833 | 3069.375 | 14.391 | 64.00 | 2.000 | 6 | 11.875 | 11.875 |
| sha256-clang / wasmvm / llgo / interpreter / 1w | 853 | 74894.417 | 77951.250 | 1169.928 | 529.20 | N/A | 84 | 10.891 | 10.922 |
| sha256-clang / wasmvm / llgo / interpreter / 2w | 1718 | 37109.916 | 38738.583 | 1160.119 | 537.19 | N/A | 85 | 11.141 | 11.172 |
| sha256-clang / wasmvm / llgo / interpreter / 4w | 3411 | 18685.834 | 19432.834 | 1167.198 | 537.07 | N/A | 87 | 11.266 | 11.297 |
| sha256-clang / wasmvm / llgo / native / 1w | 58334 | 1072.125 | 1544.209 | 15.358 | 543.79 | N/A | 84 | 11.516 | 11.531 |
| sha256-clang / wasmvm / llgo / native / 2w | 100680 | 589.334 | 1097.625 | 16.698 | 542.71 | N/A | 85 | 11.609 | 11.625 |
| sha256-clang / wasmvm / llgo / native / 4w | 135706 | 436.250 | 944.875 | 16.980 | 538.60 | N/A | 87 | 11.672 | 11.703 |
| sha256-clang / wazero / go / compiler / 1w | 47504 | 1353.750 | 1396.875 | 20.127 | 0.00 | 0.000 | 6 | 9.203 | 9.234 |
| sha256-clang / wazero / go / compiler / 2w | 105363 | 600.250 | 679.250 | 18.287 | 0.00 | 0.000 | 7 | 9.281 | 9.312 |
| sha256-clang / wazero / go / compiler / 4w | 221635 | 17.500 | 3729.166 | 17.268 | 0.00 | 0.000 | 6 | 9.156 | 9.234 |
| sha256-clang / wazero / go / compiler / termination / 1w | 6924 | 9202.334 | 9738.416 | 144.106 | 176.97 | 3.002 | 7 | 10.172 | 10.219 |
| sha256-clang / wazero / go / compiler / termination / 2w | 13910 | 4590.792 | 4842.000 | 143.263 | 177.38 | 3.006 | 6 | 10.078 | 10.125 |
| sha256-clang / wazero / go / compiler / termination / 4w | 27369 | 140.709 | 30101.708 | 143.656 | 177.59 | 3.004 | 6 | 12.344 | 12.469 |
| sha256-clang / wazero / go / interpreter / 1w | 378 | 168058.083 | 173835.584 | 2643.180 | 1584.82 | 66.002 | 6 | 12.281 | 12.312 |
| sha256-clang / wazero / go / interpreter / 2w | 748 | 85392.500 | 87572.708 | 2668.790 | 1584.80 | 66.001 | 6 | 12.266 | 12.312 |
| sha256-clang / wazero / go / interpreter / 4w | 1529 | 45493.000 | 78227.291 | 2611.459 | 1584.83 | 66.002 | 7 | 12.312 | 12.312 |
| sha256-clang / wazero / go / interpreter / termination / 1w | 380 | 167737.416 | 172153.667 | 2627.656 | 1761.45 | 69.007 | 6 | 12.656 | 12.750 |
| sha256-clang / wazero / go / interpreter / termination / 2w | 770 | 82905.042 | 85770.875 | 2596.366 | 1761.35 | 69.005 | 6 | 13.109 | 13.156 |
| sha256-clang / wazero / go / interpreter / termination / 4w | 1509 | 45686.916 | 81134.583 | 2645.484 | 1762.85 | 69.006 | 6 | 12.656 | 12.672 |
| sha256-rust / wasmedge / go / interpreter / 1w | 630 | 101380.292 | 104086.292 | 1584.948 | 0.06 | 0.001 | 6 | 27.500 | 27.547 |
| sha256-rust / wasmedge / go / interpreter / 2w | 1260 | 50618.833 | 52991.333 | 1584.546 | 0.86 | 0.002 | 7 | 27.688 | 27.750 |
| sha256-rust / wasmedge / go / interpreter / 4w | 2516 | 25130.875 | 34985.375 | 1585.697 | 0.03 | 0.000 | 10 | 28.000 | 28.188 |
| sha256-rust / wasmer / go / compiler / 1w | 64551 | 993.666 | 1041.125 | 14.162 | 0.02 | 0.000 | 23 | 21.094 | 21.188 |
| sha256-rust / wasmer / go / compiler / 2w | 132001 | 488.208 | 524.417 | 14.243 | 0.78 | 0.001 | 23 | 21.203 | 21.344 |
| sha256-rust / wasmer / go / compiler / 4w | 295533 | 12.833 | 2789.042 | 12.925 | 0.00 | 0.000 | 23 | 21.234 | 21.281 |
| sha256-rust / wasmtime / go / compiler / 1w | 66096 | 966.417 | 1023.958 | 14.189 | 0.02 | 0.000 | 23 | 16.906 | 17.078 |
| sha256-rust / wasmtime / go / compiler / 2w | 133234 | 481.292 | 506.667 | 14.229 | 0.00 | 0.000 | 23 | 16.969 | 17.016 |
| sha256-rust / wasmtime / go / compiler / 4w | 304661 | 12.458 | 2690.458 | 12.524 | 0.00 | 0.000 | 23 | 17.219 | 17.250 |
| sha256-rust / wasmvm / go / interpreter / 1w | 759 | 84093.333 | 88077.167 | 1316.052 | 64.05 | 2.000 | 6 | 8.609 | 8.625 |
| sha256-rust / wasmvm / go / interpreter / 2w | 1519 | 42053.083 | 44441.375 | 1313.861 | 64.62 | 2.002 | 8 | 8.734 | 8.750 |
| sha256-rust / wasmvm / go / interpreter / 4w | 3020 | 12523.750 | 74340.750 | 1310.092 | 64.00 | 2.000 | 6 | 8.766 | 8.766 |
| sha256-rust / wasmvm / go / native / 1w | 58635 | 1105.583 | 1135.500 | 16.123 | 64.03 | 2.000 | 6 | 12.062 | 12.109 |
| sha256-rust / wasmvm / go / native / 2w | 115723 | 533.583 | 692.958 | 16.583 | 64.00 | 2.000 | 6 | 11.875 | 11.891 |
| sha256-rust / wasmvm / go / native / 4w | 257228 | 14.458 | 3200.041 | 14.793 | 64.00 | 2.000 | 6 | 12.109 | 12.109 |
| sha256-rust / wasmvm / llgo / interpreter / 1w | 872 | 73137.250 | 77322.250 | 1144.437 | 534.24 | N/A | 84 | 11.828 | 11.969 |
| sha256-rust / wasmvm / llgo / interpreter / 2w | 1736 | 36663.542 | 38347.833 | 1148.298 | 559.94 | N/A | 85 | 12.797 | 12.812 |
| sha256-rust / wasmvm / llgo / interpreter / 4w | 3492 | 18247.791 | 19021.000 | 1140.225 | 549.93 | N/A | 87 | 15.391 | 15.406 |
| sha256-rust / wasmvm / llgo / native / 1w | 55955 | 1119.042 | 1595.625 | 15.483 | 533.62 | N/A | 84 | 11.438 | 11.469 |
| sha256-rust / wasmvm / llgo / native / 2w | 102149 | 589.125 | 1115.042 | 16.325 | 538.21 | N/A | 85 | 11.531 | 11.562 |
| sha256-rust / wasmvm / llgo / native / 4w | 135156 | 437.209 | 901.083 | 17.568 | 528.83 | N/A | 87 | 11.703 | 11.734 |
| sha256-rust / wazero / go / compiler / 1w | 42922 | 1501.375 | 1545.500 | 22.480 | 0.00 | 0.000 | 6 | 9.016 | 9.062 |
| sha256-rust / wazero / go / compiler / 2w | 93013 | 688.791 | 754.375 | 20.664 | 0.00 | 0.000 | 6 | 9.156 | 9.234 |
| sha256-rust / wazero / go / compiler / 4w | 204371 | 19.083 | 4055.083 | 18.936 | 0.00 | 0.000 | 6 | 9.125 | 9.156 |
| sha256-rust / wazero / go / compiler / termination / 1w | 7016 | 9065.416 | 9610.250 | 142.338 | 176.80 | 3.003 | 6 | 10.047 | 10.062 |
| sha256-rust / wazero / go / compiler / termination / 2w | 14122 | 4523.333 | 4785.833 | 141.293 | 177.69 | 3.006 | 8 | 9.844 | 9.875 |
| sha256-rust / wazero / go / compiler / termination / 4w | 29089 | 134.750 | 28363.750 | 135.711 | 177.63 | 3.004 | 6 | 10.781 | 10.797 |
| sha256-rust / wazero / go / interpreter / 1w | 373 | 171217.125 | 173874.708 | 2676.976 | 1584.06 | 66.001 | 6 | 12.016 | 12.094 |
| sha256-rust / wazero / go / interpreter / 2w | 754 | 84780.416 | 86566.500 | 2650.552 | 1584.84 | 66.002 | 6 | 11.219 | 11.266 |
| sha256-rust / wazero / go / interpreter / 4w | 1470 | 46213.917 | 72843.250 | 2714.738 | 1584.03 | 66.000 | 6 | 11.688 | 11.734 |
| sha256-rust / wazero / go / interpreter / termination / 1w | 380 | 167921.000 | 171331.292 | 2627.983 | 1760.89 | 69.003 | 6 | 11.672 | 11.703 |
| sha256-rust / wazero / go / interpreter / termination / 2w | 757 | 84346.000 | 86079.125 | 2639.464 | 1761.46 | 69.005 | 6 | 12.062 | 12.109 |
| sha256-rust / wazero / go / interpreter / termination / 4w | 1506 | 45905.667 | 73370.375 | 2651.647 | 1762.25 | 69.006 | 6 | 11.703 | 11.703 |
| native-c-sha256 / native-c / go / native / 1w | 67704 | 943.292 | 1009.459 | 13.801 | 0.03 | 0.000 | 6 | 7.562 | 7.766 |
| native-c-sha256 / native-c / go / native / 2w | 139269 | 462.000 | 487.917 | 13.744 | 0.00 | 0.000 | 6 | 7.641 | 7.703 |
| native-c-sha256 / native-c / go / native / 4w | 318452 | 13.125 | 2584.292 | 11.887 | 0.00 | 0.000 | 6 | 7.719 | 7.750 |
| native-c-sha256 / native-c / llgo / native / 1w | 67915 | 881.666 | 1322.708 | 12.398 | 454.33 | N/A | 84 | 9.109 | 9.125 |
| native-c-sha256 / native-c / llgo / native / 2w | 117540 | 524.291 | 927.709 | 12.886 | 444.79 | N/A | 85 | 9.062 | 9.078 |
| native-c-sha256 / native-c / llgo / native / 4w | 133541 | 457.375 | 922.250 | 12.376 | 434.31 | N/A | 87 | 9.172 | 9.203 |

### Every lifecycle phase

Median ms. First is the complete first service batch, not one exported call. Process readiness and wall intervals are directly observed separately. N/A denotes an inapplicable Wasm phase in the native-C control; no remainder is manufactured by subtraction.

| Profile | Process ready | Load | Engine | Compile/validate | Instantiate | Lookup | Workers | First batch | Mean batch readiness | Mean warm batch | Cleanup | Process wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| sha256-clang / wasmedge / go / interpreter / 1w | 19.8467 | 0.0547 | 0.0230 | 0.0700 | 0.0520 | 0.0000 | 0.0420 | 1747.8232 | 0.0257 | 1748.9318 | 0.0720 | 17521.5320 |
| sha256-clang / wasmedge / go / interpreter / 2w | 26.4764 | 0.0522 | 0.1570 | 0.0800 | 0.0960 | 0.0000 | 0.0432 | 875.4251 | 0.0284 | 879.8618 | 0.1253 | 8821.2007 |
| sha256-clang / wasmedge / go / interpreter / 4w | 22.7213 | 0.0449 | 0.1530 | 0.0740 | 0.1430 | 0.0000 | 0.0471 | 436.0932 | 0.0258 | 436.0930 | 0.1916 | 4385.9958 |
| sha256-clang / wasmer / go / compiler / 1w | 20.4817 | 0.0634 | 2.5790 | 15.4360 | 0.1550 | 0.0310 | 0.0542 | 15.3268 | 0.0204 | 15.5563 | 0.3147 | 196.9265 |
| sha256-clang / wasmer / go / compiler / 2w | 10.0664 | 0.0419 | 0.0390 | 1.6610 | 0.0500 | 0.0050 | 0.0410 | 7.4532 | 0.0232 | 7.7308 | 0.1019 | 90.3699 |
| sha256-clang / wasmer / go / compiler / 4w | 10.1671 | 0.0500 | 0.0400 | 1.6710 | 0.0830 | 0.0070 | 0.0440 | 3.4820 | 0.0213 | 3.4060 | 0.1530 | 48.4347 |
| sha256-clang / wasmtime / go / compiler / 1w | 7.9325 | 0.0415 | 0.1080 | 1.2510 | 0.0520 | 0.0020 | 0.0379 | 14.8876 | 0.0209 | 15.2091 | 0.1150 | 160.9466 |
| sha256-clang / wasmtime / go / compiler / 2w | 8.3750 | 0.0457 | 0.1120 | 1.2430 | 0.0600 | 0.0020 | 0.0376 | 7.4867 | 0.0185 | 7.6135 | 0.1333 | 87.8301 |
| sha256-clang / wasmtime / go / compiler / 4w | 8.4331 | 0.0568 | 0.1440 | 1.2470 | 0.0910 | 0.0030 | 0.0459 | 3.5443 | 0.0224 | 3.3628 | 0.1732 | 45.6093 |
| sha256-clang / wasmvm / go / interpreter / 1w | 7.7990 | 0.0615 | 0.0050 | 0.2706 | 0.0230 | 0.0003 | 0.0446 | 1389.9025 | 0.0255 | 1389.3874 | 0.0172 | 13915.7867 |
| sha256-clang / wasmvm / go / interpreter / 2w | 6.8266 | 0.0424 | 0.0046 | 0.2535 | 0.0145 | 0.0002 | 0.0382 | 691.4324 | 0.0247 | 694.2872 | 0.0241 | 6948.3415 |
| sha256-clang / wasmvm / go / interpreter / 4w | 6.0538 | 0.0454 | 0.0050 | 0.2504 | 0.0190 | 0.0003 | 0.0420 | 347.1397 | 0.0310 | 349.6526 | 0.0214 | 3501.3227 |
| sha256-clang / wasmvm / go / native / 1w | 7.0972 | 0.0423 | 0.0064 | 2.7581 | 0.0167 | 0.0003 | 0.0239 | 15.4350 | 0.0206 | 16.7351 | 0.0185 | 178.2287 |
| sha256-clang / wasmvm / go / native / 2w | 6.8386 | 0.0438 | 0.0056 | 2.3676 | 0.0192 | 0.0003 | 0.0295 | 8.0791 | 0.0202 | 8.5467 | 0.0316 | 96.4589 |
| sha256-clang / wasmvm / go / native / 4w | 6.6683 | 0.0400 | 0.0065 | 2.3363 | 0.0230 | 0.0004 | 0.0248 | 3.5993 | 0.0210 | 3.8810 | 0.0370 | 48.7023 |
| sha256-clang / wasmvm / llgo / interpreter / 1w | 7.2335 | 0.0468 | 0.0035 | 0.6883 | 0.0195 | 0.0004 | 0.7771 | 1205.9436 | 0.1493 | 1200.7626 | 0.3563 | 12015.9747 |
| sha256-clang / wasmvm / llgo / interpreter / 2w | 7.5473 | 0.0412 | 0.0034 | 0.6149 | 0.0315 | 0.0004 | 0.7802 | 597.4710 | 0.1378 | 596.0885 | 0.3908 | 5978.5007 |
| sha256-clang / wasmvm / llgo / interpreter / 4w | 8.0338 | 0.0378 | 0.0046 | 0.7411 | 0.0576 | 0.0005 | 1.2962 | 297.1761 | 0.1539 | 300.2452 | 0.4243 | 3014.2021 |
| sha256-clang / wasmvm / llgo / native / 1w | 7.6313 | 0.0476 | 0.0039 | 4.2213 | 0.0192 | 0.0004 | 0.9629 | 17.7794 | 0.1324 | 17.5541 | 0.3047 | 192.0412 |
| sha256-clang / wasmvm / llgo / native / 2w | 8.1283 | 0.0431 | 0.0045 | 4.1378 | 0.0247 | 0.0004 | 1.0162 | 9.7733 | 0.1169 | 10.1709 | 0.2885 | 117.2164 |
| sha256-clang / wasmvm / llgo / native / 4w | 6.9960 | 0.0396 | 0.0051 | 4.0621 | 0.0296 | 0.0005 | 1.0260 | 7.7741 | 0.1309 | 7.5457 | 0.3048 | 90.8461 |
| sha256-clang / wazero / go / compiler / 1w | 7.6592 | 0.0421 | 0.0537 | 0.3000 | 0.0290 | 0.0020 | 0.0372 | 21.2556 | 0.0189 | 21.5559 | 0.0164 | 222.4135 |
| sha256-clang / wazero / go / compiler / 2w | 6.1089 | 0.0477 | 0.0563 | 0.3102 | 0.0366 | 0.0027 | 0.0378 | 9.7299 | 0.0224 | 9.7188 | 0.0183 | 105.0390 |
| sha256-clang / wazero / go / compiler / 4w | 6.9249 | 0.0395 | 0.0529 | 0.2914 | 0.0570 | 0.0038 | 0.0508 | 4.6242 | 0.0217 | 4.6202 | 0.0207 | 54.8248 |
| sha256-clang / wazero / go / compiler / termination / 1w | 7.3665 | 0.0418 | 0.0508 | 0.3034 | 0.0220 | 0.0025 | 0.0360 | 146.2572 | 0.0217 | 147.8975 | 0.0194 | 1487.1838 |
| sha256-clang / wazero / go / compiler / termination / 2w | 6.2373 | 0.0486 | 0.0525 | 0.3125 | 0.0363 | 0.0044 | 0.0421 | 73.6046 | 0.0253 | 73.6158 | 0.0250 | 741.5805 |
| sha256-clang / wazero / go / compiler / termination / 4w | 6.4458 | 0.0329 | 0.0520 | 0.3014 | 0.0583 | 0.0037 | 0.0431 | 36.6957 | 0.0241 | 37.4153 | 0.0202 | 381.7921 |
| sha256-clang / wazero / go / interpreter / 1w | 6.8392 | 0.0425 | 0.0016 | 0.0786 | 0.0229 | 0.0003 | 0.0417 | 2706.3527 | 0.0230 | 2708.1961 | 0.0206 | 27062.0271 |
| sha256-clang / wazero / go / interpreter / 2w | 6.0593 | 0.0374 | 0.0018 | 0.0772 | 0.0400 | 0.0003 | 0.0375 | 1373.2777 | 0.0264 | 1368.2016 | 0.0201 | 13706.7293 |
| sha256-clang / wazero / go / interpreter / 4w | 6.8754 | 0.0457 | 0.0018 | 0.0821 | 0.0575 | 0.0003 | 0.0500 | 673.3809 | 0.0267 | 669.8029 | 0.0220 | 6710.5750 |
| sha256-clang / wazero / go / interpreter / termination / 1w | 6.2500 | 0.0512 | 0.0018 | 0.0765 | 0.0220 | 0.0002 | 0.0343 | 2654.5159 | 0.0261 | 2691.3339 | 0.0201 | 26905.2718 |
| sha256-clang / wazero / go / interpreter / termination / 2w | 7.1345 | 0.0389 | 0.0019 | 0.0824 | 0.0350 | 0.0003 | 0.0379 | 1325.0530 | 0.0235 | 1330.1514 | 0.0205 | 13300.8959 |
| sha256-clang / wazero / go / interpreter / termination / 4w | 7.1853 | 0.0416 | 0.0016 | 0.0760 | 0.0534 | 0.0003 | 0.0474 | 678.5715 | 0.0240 | 678.4881 | 0.0227 | 6794.0771 |
| sha256-rust / wasmedge / go / interpreter / 1w | 20.5803 | 0.0558 | 0.0240 | 0.0730 | 0.0570 | 0.0010 | 0.0381 | 1625.4866 | 0.0254 | 1624.6346 | 0.0841 | 16268.2564 |
| sha256-rust / wasmedge / go / interpreter / 2w | 19.4647 | 0.0472 | 0.0240 | 0.0770 | 0.0790 | 0.0000 | 0.0411 | 814.5520 | 0.0264 | 812.7851 | 0.1123 | 8164.2433 |
| sha256-rust / wasmedge / go / interpreter / 4w | 20.0363 | 0.0579 | 0.0240 | 0.0710 | 0.1280 | 0.0000 | 0.0366 | 408.8494 | 0.0299 | 407.0638 | 0.1845 | 4100.0341 |
| sha256-rust / wasmer / go / compiler / 1w | 10.7692 | 0.0520 | 0.0380 | 1.7120 | 0.0340 | 0.0050 | 0.0918 | 15.3772 | 0.0222 | 15.8634 | 0.0807 | 172.6712 |
| sha256-rust / wasmer / go / compiler / 2w | 9.6313 | 0.0513 | 0.0370 | 1.7610 | 0.0530 | 0.0060 | 0.0523 | 7.8388 | 0.0227 | 7.7575 | 0.1035 | 91.3065 |
| sha256-rust / wasmer / go / compiler / 4w | 10.3134 | 0.0489 | 0.0470 | 1.6820 | 0.0810 | 0.0060 | 0.0562 | 3.6175 | 0.0221 | 3.4649 | 0.1608 | 48.9919 |
| sha256-rust / wasmtime / go / compiler / 1w | 7.3372 | 0.0426 | 0.1070 | 1.3000 | 0.0550 | 0.0020 | 0.0357 | 15.4027 | 0.0199 | 15.4925 | 0.1293 | 164.9777 |
| sha256-rust / wasmtime / go / compiler / 2w | 8.8115 | 0.0433 | 0.1090 | 1.3390 | 0.0660 | 0.0020 | 0.0420 | 7.5795 | 0.0239 | 7.6858 | 0.1422 | 88.7188 |
| sha256-rust / wasmtime / go / compiler / 4w | 8.7288 | 0.0467 | 0.1110 | 1.3050 | 0.0900 | 0.0020 | 0.0469 | 3.4557 | 0.0225 | 3.3611 | 0.1777 | 46.0139 |
| sha256-rust / wasmvm / go / interpreter / 1w | 5.9640 | 0.0356 | 0.0040 | 0.2560 | 0.0109 | 0.0002 | 0.0378 | 1344.3899 | 0.0242 | 1349.1222 | 0.0174 | 13487.6924 |
| sha256-rust / wasmvm / go / interpreter / 2w | 6.8917 | 0.0608 | 0.0049 | 0.2550 | 0.0317 | 0.0003 | 0.0445 | 674.6261 | 0.0249 | 673.9141 | 0.0191 | 6750.4246 |
| sha256-rust / wasmvm / go / interpreter / 4w | 5.9173 | 0.0520 | 0.0051 | 0.2755 | 0.2046 | 0.0005 | 0.0429 | 337.8761 | 0.0290 | 339.0450 | 0.0191 | 3392.6639 |
| sha256-rust / wasmvm / go / native / 1w | 5.8565 | 0.0418 | 0.0056 | 2.3357 | 0.0179 | 0.0003 | 0.0259 | 17.5326 | 0.0219 | 17.4641 | 0.0206 | 184.7727 |
| sha256-rust / wasmvm / go / native / 2w | 6.9515 | 0.0412 | 0.0058 | 2.3504 | 0.0215 | 0.0003 | 0.0297 | 8.0910 | 0.0197 | 8.8487 | 0.0280 | 98.2683 |
| sha256-rust / wasmvm / go / native / 4w | 6.6248 | 0.0375 | 0.0060 | 2.3002 | 0.0287 | 0.0004 | 0.0294 | 4.0466 | 0.0218 | 3.9809 | 0.0372 | 49.4005 |
| sha256-rust / wasmvm / llgo / interpreter / 1w | 6.6489 | 0.0372 | 0.0041 | 0.6365 | 0.0805 | 0.0004 | 0.9553 | 1175.1465 | 0.1239 | 1174.1646 | 0.3549 | 11763.7832 |
| sha256-rust / wasmvm / llgo / interpreter / 2w | 6.7502 | 0.0445 | 0.0039 | 0.7406 | 0.4900 | 0.0006 | 0.7491 | 589.2955 | 0.1460 | 589.9513 | 0.4858 | 5899.5484 |
| sha256-rust / wasmvm / llgo / interpreter / 4w | 7.8376 | 0.0467 | 0.0047 | 0.6854 | 0.7151 | 0.0006 | 1.0693 | 291.3151 | 0.1350 | 293.2315 | 0.3660 | 2945.7155 |
| sha256-rust / wasmvm / llgo / native / 1w | 6.5901 | 0.0486 | 0.0042 | 4.0687 | 0.0237 | 0.0004 | 1.0419 | 17.8895 | 0.1233 | 18.3005 | 0.3161 | 198.7966 |
| sha256-rust / wasmvm / llgo / native / 2w | 7.8411 | 0.0382 | 0.0043 | 4.0347 | 0.0225 | 0.0003 | 1.0674 | 10.1700 | 0.1355 | 10.0245 | 0.2987 | 120.3102 |
| sha256-rust / wasmvm / llgo / native / 4w | 8.3486 | 0.0490 | 0.0051 | 4.0062 | 0.0311 | 0.0006 | 1.0360 | 7.8028 | 0.1159 | 7.5764 | 0.2929 | 90.4458 |
| sha256-rust / wazero / go / compiler / 1w | 6.8921 | 0.0423 | 0.0525 | 0.3167 | 0.0375 | 0.0025 | 0.0335 | 23.7920 | 0.0189 | 23.8575 | 0.0147 | 245.6916 |
| sha256-rust / wazero / go / compiler / 2w | 6.6625 | 0.0391 | 0.0540 | 0.3215 | 0.0387 | 0.0031 | 0.1059 | 10.8825 | 0.0201 | 11.0092 | 0.0154 | 117.8548 |
| sha256-rust / wazero / go / compiler / 4w | 6.8860 | 0.0397 | 0.0515 | 0.3179 | 0.1999 | 0.0050 | 0.0408 | 5.0775 | 0.0217 | 5.0105 | 0.0163 | 58.9806 |
| sha256-rust / wazero / go / compiler / termination / 1w | 6.2505 | 0.0389 | 0.0548 | 0.3290 | 0.0278 | 0.0027 | 0.0334 | 143.8320 | 0.0192 | 145.9433 | 0.0201 | 1465.6701 |
| sha256-rust / wazero / go / compiler / termination / 2w | 5.8877 | 0.0406 | 0.0534 | 0.3360 | 0.0462 | 0.0031 | 0.1095 | 71.8886 | 0.0224 | 72.5084 | 0.0155 | 733.1083 |
| sha256-rust / wazero / go / compiler / termination / 4w | 6.4771 | 0.0451 | 0.0518 | 0.3227 | 0.2036 | 0.0052 | 0.0356 | 35.1126 | 0.0232 | 35.2023 | 0.0253 | 362.9032 |
| sha256-rust / wazero / go / interpreter / 1w | 6.2806 | 0.0542 | 0.0020 | 0.0810 | 0.0369 | 0.0003 | 0.0358 | 2738.4752 | 0.0300 | 2742.9033 | 0.0213 | 27431.2200 |
| sha256-rust / wazero / go / interpreter / 2w | 6.1778 | 0.0343 | 0.0016 | 0.0778 | 0.0392 | 0.0003 | 0.0398 | 1359.7814 | 0.0225 | 1358.6814 | 0.0208 | 13602.4610 |
| sha256-rust / wazero / go / interpreter / 4w | 5.7622 | 0.0445 | 0.0018 | 0.0816 | 0.1791 | 0.0008 | 0.0389 | 687.1568 | 0.0295 | 696.8034 | 0.0257 | 6955.5420 |
| sha256-rust / wazero / go / interpreter / termination / 1w | 6.7434 | 0.0413 | 0.0016 | 0.0803 | 0.0258 | 0.0003 | 0.0334 | 2690.6208 | 0.0277 | 2691.6737 | 0.0215 | 27043.1816 |
| sha256-rust / wazero / go / interpreter / termination / 2w | 5.6571 | 0.0371 | 0.0016 | 0.0764 | 0.0430 | 0.0003 | 0.0345 | 1358.5695 | 0.0239 | 1352.1636 | 0.0196 | 13555.7793 |
| sha256-rust / wazero / go / interpreter / termination / 4w | 6.6508 | 0.0403 | 0.0015 | 0.0783 | 0.1970 | 0.0008 | 0.0395 | 678.5818 | 0.0271 | 680.1087 | 0.0235 | 6807.0950 |
| native-c-sha256 / native-c / go / native / 1w | 8.7096 | N/A | N/A | N/A | N/A | N/A | 0.0342 | 14.4844 | 0.0197 | 15.1247 | 0.0195 | 161.1797 |
| native-c-sha256 / native-c / go / native / 2w | 7.4845 | N/A | N/A | N/A | N/A | N/A | 0.0360 | 7.2286 | 0.0202 | 7.3527 | 0.0128 | 81.2251 |
| native-c-sha256 / native-c / go / native / 4w | 6.0455 | N/A | N/A | N/A | N/A | N/A | 0.0380 | 3.2055 | 0.0243 | 3.2156 | 0.0155 | 39.6030 |
| native-c-sha256 / native-c / llgo / native / 1w | 10.1397 | N/A | N/A | N/A | N/A | N/A | 0.8281 | 14.9297 | 0.1401 | 15.0777 | 0.3256 | 166.8486 |
| native-c-sha256 / native-c / llgo / native / 2w | 7.8957 | N/A | N/A | N/A | N/A | N/A | 0.7843 | 8.9473 | 0.1312 | 8.7119 | 0.2894 | 99.4472 |
| native-c-sha256 / native-c / llgo / native / 4w | 11.5262 | N/A | N/A | N/A | N/A | N/A | 0.8517 | 7.5548 | 0.1233 | 7.6681 | 0.2961 | 90.2023 |

### Every identical-binary A/A interval

Paired block log-ratio bootstrap, 8,000 draws, unchanged routing calibration seed/statistic; 95% interval must fit entirely inside [0.975, 1.025]. These controls diagnose noise/bias. Resolved controls alone are not a cross-runtime speedup or a complete controlled-host acceptance.

| Profile | A/A lower | A/A upper | Resolved |
| --- | ---: | ---: | --- |
| sha256-clang / wasmedge / go / interpreter / 1w | 1.00339 | 1.01018 | yes |
| sha256-clang / wasmedge / go / interpreter / 2w | 0.98289 | 1.00346 | yes |
| sha256-clang / wasmedge / go / interpreter / 4w | 0.99938 | 1.01343 | yes |
| sha256-clang / wasmer / go / compiler / 1w | 0.94689 | 0.99864 | no |
| sha256-clang / wasmer / go / compiler / 2w | 0.99674 | 1.12653 | no |
| sha256-clang / wasmer / go / compiler / 4w | 0.99085 | 1.01245 | yes |
| sha256-clang / wasmtime / go / compiler / 1w | 1.00632 | 1.05076 | no |
| sha256-clang / wasmtime / go / compiler / 2w | 0.95244 | 1.02223 | no |
| sha256-clang / wasmtime / go / compiler / 4w | 0.97988 | 1.00312 | yes |
| sha256-clang / wasmvm / go / interpreter / 1w | 1.00125 | 1.05488 | no |
| sha256-clang / wasmvm / go / interpreter / 2w | 0.99596 | 1.00445 | yes |
| sha256-clang / wasmvm / go / interpreter / 4w | 0.99707 | 1.02491 | yes |
| sha256-clang / wasmvm / go / native / 1w | 0.94390 | 0.97867 | no |
| sha256-clang / wasmvm / go / native / 2w | 1.00754 | 1.04480 | no |
| sha256-clang / wasmvm / go / native / 4w | 0.98171 | 1.08661 | no |
| sha256-clang / wasmvm / llgo / interpreter / 1w | 0.99558 | 1.01722 | yes |
| sha256-clang / wasmvm / llgo / interpreter / 2w | 1.00042 | 1.02106 | yes |
| sha256-clang / wasmvm / llgo / interpreter / 4w | 0.93932 | 1.00576 | no |
| sha256-clang / wasmvm / llgo / native / 1w | 0.95608 | 1.03867 | no |
| sha256-clang / wasmvm / llgo / native / 2w | 0.91151 | 0.99909 | no |
| sha256-clang / wasmvm / llgo / native / 4w | 0.98513 | 1.06825 | no |
| sha256-clang / wazero / go / compiler / 1w | 0.99793 | 1.06440 | no |
| sha256-clang / wazero / go / compiler / 2w | 1.00658 | 1.01716 | yes |
| sha256-clang / wazero / go / compiler / 4w | 0.99418 | 1.00982 | yes |
| sha256-clang / wazero / go / compiler / termination / 1w | 0.97728 | 1.03469 | no |
| sha256-clang / wazero / go / compiler / termination / 2w | 0.98498 | 1.02446 | yes |
| sha256-clang / wazero / go / compiler / termination / 4w | 0.95614 | 1.00198 | no |
| sha256-clang / wazero / go / interpreter / 1w | 0.97712 | 1.01924 | yes |
| sha256-clang / wazero / go / interpreter / 2w | 0.98736 | 1.03065 | no |
| sha256-clang / wazero / go / interpreter / 4w | 0.98845 | 1.02853 | no |
| sha256-clang / wazero / go / interpreter / termination / 1w | 0.98051 | 1.01023 | yes |
| sha256-clang / wazero / go / interpreter / termination / 2w | 0.98353 | 1.02187 | yes |
| sha256-clang / wazero / go / interpreter / termination / 4w | 0.94716 | 0.99860 | no |
| sha256-rust / wasmedge / go / interpreter / 1w | 0.99444 | 1.00842 | yes |
| sha256-rust / wasmedge / go / interpreter / 2w | 0.99577 | 1.00230 | yes |
| sha256-rust / wasmedge / go / interpreter / 4w | 1.00099 | 1.00424 | yes |
| sha256-rust / wasmer / go / compiler / 1w | 0.99357 | 1.02322 | yes |
| sha256-rust / wasmer / go / compiler / 2w | 0.99003 | 1.00722 | yes |
| sha256-rust / wasmer / go / compiler / 4w | 0.94047 | 1.03596 | no |
| sha256-rust / wasmtime / go / compiler / 1w | 0.98450 | 1.03824 | no |
| sha256-rust / wasmtime / go / compiler / 2w | 0.98822 | 1.00782 | yes |
| sha256-rust / wasmtime / go / compiler / 4w | 0.99739 | 1.02472 | yes |
| sha256-rust / wasmvm / go / interpreter / 1w | 0.98780 | 1.00922 | yes |
| sha256-rust / wasmvm / go / interpreter / 2w | 0.99453 | 1.00588 | yes |
| sha256-rust / wasmvm / go / interpreter / 4w | 0.89487 | 1.00765 | no |
| sha256-rust / wasmvm / go / native / 1w | 0.94844 | 1.00305 | no |
| sha256-rust / wasmvm / go / native / 2w | 0.91984 | 1.03387 | no |
| sha256-rust / wasmvm / go / native / 4w | 0.97816 | 1.10746 | no |
| sha256-rust / wasmvm / llgo / interpreter / 1w | 1.00098 | 1.11831 | no |
| sha256-rust / wasmvm / llgo / interpreter / 2w | 0.99298 | 1.00788 | yes |
| sha256-rust / wasmvm / llgo / interpreter / 4w | 0.98726 | 1.01341 | yes |
| sha256-rust / wasmvm / llgo / native / 1w | 0.99871 | 2.64977 | no |
| sha256-rust / wasmvm / llgo / native / 2w | 0.97888 | 1.03464 | no |
| sha256-rust / wasmvm / llgo / native / 4w | 0.98729 | 1.17259 | no |
| sha256-rust / wazero / go / compiler / 1w | 0.98249 | 1.01119 | yes |
| sha256-rust / wazero / go / compiler / 2w | 0.95345 | 1.02850 | no |
| sha256-rust / wazero / go / compiler / 4w | 1.00572 | 1.02442 | yes |
| sha256-rust / wazero / go / compiler / termination / 1w | 0.85314 | 1.00219 | no |
| sha256-rust / wazero / go / compiler / termination / 2w | 0.98258 | 1.02721 | no |
| sha256-rust / wazero / go / compiler / termination / 4w | 0.99790 | 1.03943 | no |
| sha256-rust / wazero / go / interpreter / 1w | 0.98004 | 1.12562 | no |
| sha256-rust / wazero / go / interpreter / 2w | 0.99396 | 1.01479 | yes |
| sha256-rust / wazero / go / interpreter / 4w | 0.94210 | 0.99304 | no |
| sha256-rust / wazero / go / interpreter / termination / 1w | 0.98607 | 1.00909 | yes |
| sha256-rust / wazero / go / interpreter / termination / 2w | 0.99857 | 1.00746 | yes |
| sha256-rust / wazero / go / interpreter / termination / 4w | 0.98553 | 1.00845 | yes |
| native-c-sha256 / native-c / go / native / 1w | 0.61281 | 1.00312 | no |
| native-c-sha256 / native-c / go / native / 2w | 0.99507 | 1.03885 | no |
| native-c-sha256 / native-c / go / native / 4w | 0.98969 | 1.00275 | yes |
| native-c-sha256 / native-c / llgo / native / 1w | 0.97518 | 1.02050 | yes |
| native-c-sha256 / native-c / llgo / native / 2w | 0.90459 | 0.98798 | no |
| native-c-sha256 / native-c / llgo / native / 4w | 0.98853 | 1.01688 | yes |

All216 final qualification probes passed on the seven fresh binaries before timing. The first actual-loader audit incorrectly treated @rpath as an absolute path; it failed its metadata comparison after successful execution. The corrected audit resolves the declared SDK path to its actually loaded file and checks its pinned hash. Both complete and initial failed-audit observations remain retained. No executable, source, timer, count or sampling policy changed between those loader audits.

The standalone strict-C11 and channel-waiter findings remain tracked in [llgo3034](https://github.com/xgo-dev/llgo/issues/3034) and [llgo3035](https://github.com/xgo-dev/llgo/issues/3035). The owner confirms upstream is solving them; this event keeps the compiler pin and GNU11 policy fixed, without duplicate compiler work. Actual SDK reference tests pass on macOSarm64 and the [Linux validation workflow](https://github.com/cpunion/wasmvm/actions/runs/42652753116). They are Go adapter coverage, not engine/specification/C coverage.

An unrelated process was observed using approximately one CPU core before and during sampling. It was outside this task and unchanged. Host frequency, affinity and background isolation remain uncontrolled. No local task build/test/probe/archive overlaps timing. All 72 profiles and every adverse observation remain present, independent of A/A outcome. No universal scheduler, RSS or runtime win is inferred. The earlier54-profile event and Linux groups are separate cohorts, never pooled.

P3 I/O application comparisons, remaining serial/native/compiler gaps and controlled performance/production release acceptance remain open.

Retained event: `artifacts/service-all-runtimes-20261110` (raw result SHA256 `60bef1d18e1a5adb2dcf31740005196e5e0d1724e2724c03932077835bb45497`); source bundles, exact build manifests, source/SDK/compiler/binary hashes, loader probes, raw arrays, per-process reductions and validation logs remain available.
