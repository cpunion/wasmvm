# Native register-plan reuse and runtime comparison

Engine base `ec3b3a42eab7f82a2c5a3e682f127abd55c2240e` and measured source `7fd13e7c1a326d41051a028835b3953cb9089bb6`.
Both Core sources use the same harness instrumentation and matched compiler flags.
Apple M4 Max and macOS arm64. Go 1.27.3. Pinned LLGo 3c6b33ec and
refreshed LLGo `42efacaeff2e196622296415540e927724f7dcef` use `-O3 -lto=full`.
WasmEdge AOT uses O3. Its interpreter and AOT use the official 0.18.1 SDK
with Threads explicitly enabled. Wasmtime and Wasmer use Cranelift.
All 33 Core workloads use 6 blocks, two warmups and 7
interleaved steady calls, totalling 7,740 checked process records.

Warm Core cells show ns/iteration and the speed difference against the fastest
observed peer in that row, using peer time divided by candidate time minus one.
The fixed SHA denominator is one message block. The call-control denominator
is one host invocation. † marks unresolved A/A controls in the comparison.
US marks a capability rejection. FAIL marks a named profile's execution failure.
NM marks the absence of an equivalent benchmark adapter.

## Core native

Median wins by Go, pinned LLGo and refreshed LLGo are 12, 15, 12 out of 33. Resolved upper ratio bounds at or below one against the row reference occur in 0, 2, 1 cases.

| Workload | Row reference | wasmvm Go | wasmvm LLGo (pinned) | wasmvm LLGo 42efacae | wazero 1.12.0 | Wasmtime 48.0.1 | Wasmer 7.3.0 | WasmEdge 0.17.1 | WasmEdge 0.18.1 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-compiler | 3,500.00 (-82.14%) † | 3,271.00 (-80.89%) † | 2,750.00 (-77.27%) † | 625.00 (ref) † | 1,437.00 (-56.51%) † | 2,708.00 (-76.92%) † | 3,604.00 (-82.66%) † | 4,062.50 (-84.62%) † |
| original22-v1/core/integer-loop | wasmtime-compiler | 1.601 (-2.12%) | 1.584 (-1.04%) | 1.580 (-0.78%) | 1.740 (-9.91%) † | 1.567 (ref) | 1.583 (-1.00%) | 1.577 (-0.63%) | 1.573 (-0.35%) |
| original22-v1/core/function-calls | wasmer-compiler | 0.812 (-2.20%) † | 0.794 (+0.13%) † | 0.795 (-0.07%) † | 1.459 (-45.54%) † | 0.920 (-13.60%) † | 0.795 (ref) † | 0.795 (-0.02%) † | 0.797 (-0.26%) † |
| original22-v1/core/memory-load-store | wasmedge-0.18.1-compiler | 0.456 (-37.26%) † | 0.440 (-34.97%) † | 0.444 (-35.59%) | 0.891 (-67.93%) † | 0.484 (-40.97%) | 0.546 (-47.66%) † | 0.287 (-0.58%) † | 0.286 (ref) |
| original22-v1/memory/narrow-unaligned | wasmedge-compiler | 0.675 (-23.18%) | 0.657 (-21.15%) | 0.655 (-20.89%) | 1.086 (-52.27%) † | 0.702 (-26.20%) | 0.735 (-29.46%) | 0.518 (ref) | 0.521 (-0.61%) |
| original22-v1/memory/dependent-address | wasmedge-0.18.1-compiler | 1.887 (-23.51%) | 1.874 (-22.99%) | 1.869 (-22.79%) | 3.221 (-55.20%) | 2.087 (-30.85%) | 2.108 (-31.54%) | 1.454 (-0.74%) | 1.443 (ref) |
| original22-v1/memory/float64-unaligned | wasmedge-0.18.1-compiler | 0.764 (-10.65%) † | 0.748 (-8.82%) | 0.750 (-9.07%) | 0.885 (-22.90%) | 0.725 (-5.94%) | 0.726 (-5.96%) | 0.683 (-0.07%) | 0.682 (ref) |
| original22-v1/runtime/simd-memory-full | wasmer-compiler | 3.405 (+3.30%) † | 3.406 (+3.26%) † | 3.388 (+3.81%) † | 3.793 (-7.26%) † | 3.676 (-4.30%) † | 3.517 (ref) † | 3.645 (-3.49%) † | 3.649 (-3.61%) † |
| original22-v1/runtime/simd-memory-widen | wasmtime-compiler | 1.072 (-2.03%) | 1.061 (-1.03%) | 1.059 (-0.84%) | 1.174 (-10.54%) † | 1.050 (ref) | 1.061 (-1.03%) | 1.055 (-0.48%) † | 1.062 (-1.18%) |
| original22-v1/runtime/simd-memory-lanes | wasmedge-compiler | 5.091 (-0.89%) | 5.008 (+0.76%) | 5.047 (-0.02%) | 6.092 (-17.18%) | 5.913 (-14.66%) | 5.872 (-14.07%) | 5.046 (ref) | 5.064 (-0.36%) |
| original22-v1/gc/i31-roundtrip | wasmedge-compiler | 2.194 (-11.58%) † | 1.915 (+1.31%) † | 1.954 (-0.74%) † | US | 2.115 (-8.27%) † | US | 1.940 (ref) † | 1.954 (-0.74%) † |
| original22-v1/gc/struct-access | wasmtime-compiler | 2.744 (-46.55%) † | 1.875 (-21.78%) † | 1.756 (-16.49%) † | US | 1.467 (ref) † | US | 10.025 (-85.37%) † | 7.698 (-80.95%) † |
| original22-v1/gc/struct-allocation | wasmedge-0.18.1-compiler | 29.374 (+17.10%) † | 27.725 (+24.06%) † | 28.167 (+22.12%) † | US | 70.668 (-51.33%) † | US | 39.365 (-12.62%) † | 34.396 (ref) † |
| original22-v1/gc/struct-indirect-access | wasmtime-compiler | 3.360 (-8.24%) † | 2.473 (+24.69%) † | 2.487 (+23.95%) † | US | 3.083 (ref) † | US | 26.656 (-88.43%) † | 15.858 (-80.56%) † |
| original22-v1/threads/private-memory | wasmtime-compiler | 1.115 (+25.04%) † | 0.785 (+77.45%) † | 0.787 (+76.98%) † | 1.665 (-16.27%) † | 1.394 (ref) † | 1.681 (-17.10%) † | 1.635 (-14.78%) † | 1.750 (-20.36%) † |
| original22-v1/threads/shared-memory | wasmtime-compiler | 2.463 (-40.53%) † | 1.940 (-24.49%) † | 1.829 (-19.93%) † | 1.677 (-12.67%) † | 1.465 (ref) † | 1.769 (-17.20%) † | 1.721 (-14.89%) † | 1.758 (-16.71%) † |
| original22-v1/threads/atomic-rmw | wasmtime-compiler | 2.925 (-22.51%) † | 2.492 (-9.03%) † | 2.469 (-8.18%) † | 2.427 (-6.61%) † | 2.267 (ref) † | 4.517 (-49.82%) † | 2.496 (-9.19%) † | 2.594 (-12.61%) † |
| original22-v1/threads/notify-empty | wasmer-compiler | 2.031 (+170.06%) † | 1.598 (+243.28%) † | 1.523 (+260.19%) † | 22.252 (-75.35%) † | 7.242 (-24.25%) † | 5.485 (ref) † | 9.840 (-44.25%) † | 7.090 (-22.63%) † |
| original22-v1/threads/wait-mismatch | wasmer-compiler | 2.390 (+31.04%) † | 1.760 (+77.87%) † | 1.773 (+76.62%) † | 25.587 (-87.76%) † | 9.175 (-65.87%) † | 3.131 (ref) † | 3.848 (-18.62%) † | 3.425 (-8.58%) † |
| original22-v1/threads/wait-timeout | wasmtime-compiler | 2.821 (+1454.67%) † | 1.698 (+2482.85%) † | 1.696 (+2485.97%) † | 286.89 (-84.71%) † | 43.854 (ref) | 172.68 (-74.60%) | 61.475 (-28.66%) | 58.652 (-25.23%) |
| original22-v1/applications/sha256-clang | wasmedge-compiler | 205.47 (-15.82%) † | 203.48 (-14.99%) † | 201.93 (-14.34%) † | 298.50 (-42.05%) † | 192.78 (-10.28%) † | 197.45 (-12.40%) † | 172.97 (ref) † | 175.95 (-1.70%) † |
| original22-v1/applications/sha256-rust | wasmedge-0.18.1-compiler | 214.73 (-15.06%) | 211.87 (-13.91%) | 210.33 (-13.28%) | 326.57 (-44.15%) † | 188.79 (-3.39%) | 189.97 (-3.99%) † | 183.38 (-0.54%) † | 182.39 (ref) |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-compiler | 1.385 (-4.99%) | 1.375 (-4.25%) † | 1.370 (-3.93%) | US | 1.840 (-28.47%) | US | 1.316 (ref) | 1.321 (-0.36%) |
| gc-sustained-v1/gc/struct-access | wasmtime-compiler | 1.074 (-2.24%) | 1.060 (-0.89%) | 1.060 (-0.89%) | US | 1.050 (ref) | US | FAIL | 6.853 (-84.68%) |
| gc-sustained-v1/gc/struct-allocation | wasmedge-0.18.1-compiler | 32.025 (+0.92%) † | 33.225 (-2.72%) | 33.366 (-3.13%) † | US | 50.481 (-35.97%) † | US | FAIL | 32.321 (ref) |
| gc-sustained-v1/gc/struct-indirect-access | wasmtime-compiler | 1.689 (+53.27%) † | 1.644 (+57.50%) † | 1.637 (+58.15%) † | US | 2.589 (ref) | US | FAIL | 15.103 (-82.86%) † |
| threads-sustained-v2/threads/private-memory | wasmtime-compiler | 0.452 (+141.45%) † | 0.346 (+215.68%) † | 0.346 (+215.48%) † | 1.290 (-15.41%) † | 1.091 (ref) † | 1.119 (-2.48%) † | 1.120 (-2.60%) † | 1.108 (-1.50%) † |
| threads-sustained-v2/threads/shared-memory | wasmtime-compiler | 1.454 (-25.17%) † | 1.304 (-16.58%) † | 1.307 (-16.76%) † | 1.245 (-12.63%) † | 1.088 (ref) † | 1.105 (-1.53%) † | 1.111 (-2.03%) † | 1.120 (-2.83%) † |
| threads-sustained-v2/threads/atomic-rmw | wasmtime-compiler | 2.093 (-10.65%) † | 1.920 (-2.59%) † | 1.895 (-1.33%) † | 2.183 (-14.31%) † | 1.870 (ref) † | 4.044 (-53.75%) † | 1.925 (-2.85%) † | 1.922 (-2.71%) † |
| threads-sustained-v2/threads/notify-empty | wasmer-compiler | 1.121 (+323.55%) † | 0.956 (+396.53%) † | 0.954 (+397.62%) † | 21.196 (-77.60%) † | 6.890 (-31.08%) | 4.748 (ref) | 9.037 (-47.46%) † | 6.391 (-25.71%) † |
| threads-sustained-v2/threads/wait-mismatch | wasmer-compiler | 1.273 (+94.34%) † | 1.111 (+122.53%) † | 1.123 (+120.34%) † | 24.705 (-89.99%) † | 8.979 (-72.45%) † | 2.473 (ref) † | 3.251 (-23.93%) † | 2.614 (-5.39%) † |
| threads-sustained-v2/threads/wait-timeout | wasmtime-compiler | 1.119 (+3742.11%) † | 0.964 (+4358.80%) | 0.970 (+4334.87%) | 283.90 (-84.85%) | 43.000 (ref) | 169.90 (-74.69%) | 59.516 (-27.75%) † | 57.628 (-25.38%) |
| generic-calls-v1/core/generic-calls | wasmedge-compiler | 1.046 (-1.03%) | 1.039 (-0.38%) | 1.039 (-0.39%) | 2.145 (-51.74%) | 2.078 (-50.17%) | 2.067 (-49.91%) | 1.035 (ref) | 1.039 (-0.38%) |

## Core interpreter

Median wins by Go, pinned LLGo and refreshed LLGo are 6, 7, 7 out of 33. Resolved upper ratio bounds at or below one against the row reference occur in 2, 6, 4 cases.

| Workload | Row reference | wasmvm Go | wasmvm LLGo (pinned) | wasmvm LLGo 42efacae | wazero 1.12.0 | Wasmtime 48.0.1 | Wasmer 7.3.0 | WasmEdge 0.17.1 | WasmEdge 0.18.1 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | wazero-interpreter | 2,416.50 (-54.29%) † | 2,354.00 (-53.08%) † | 2,291.00 (-51.79%) † | 1,104.50 (ref) † | NM | NM | 3,041.50 (-63.69%) † | 2,542.00 (-56.55%) † |
| original22-v1/core/integer-loop | wasmedge-0.18.1-interpreter | 28.606 (-13.48%) | 23.982 (+3.21%) | 22.958 (+7.81%) | 82.457 (-69.98%) | NM | NM | 48.908 (-49.39%) | 24.751 (ref) |
| original22-v1/core/function-calls | wasmedge-0.18.1-interpreter | 46.368 (-21.27%) | 44.050 (-17.12%) | 43.103 (-15.30%) | 91.521 (-60.11%) | NM | NM | 62.813 (-41.88%) | 36.508 (ref) |
| original22-v1/core/memory-load-store | wasmedge-0.18.1-interpreter | 64.131 (-18.95%) † | 55.226 (-5.88%) † | 54.982 (-5.46%) † | 144.64 (-64.06%) † | NM | NM | 91.423 (-43.14%) † | 51.980 (ref) † |
| original22-v1/memory/narrow-unaligned | wasmedge-0.18.1-interpreter | 84.854 (-24.66%) | 72.927 (-12.33%) | 74.692 (-14.41%) | 185.90 (-65.61%) | NM | NM | 113.92 (-43.88%) | 63.932 (ref) |
| original22-v1/memory/dependent-address | wasmedge-0.18.1-interpreter | 37.547 (-27.76%) | 32.407 (-16.31%) | 32.507 (-16.56%) | 96.843 (-71.99%) † | NM | NM | 53.762 (-49.55%) | 27.123 (ref) |
| original22-v1/memory/float64-unaligned | wasmedge-0.18.1-interpreter | 66.960 (-29.17%) | 57.143 (-17.01%) | 57.713 (-17.82%) | 139.18 (-65.92%) | NM | NM | 84.874 (-44.12%) | 47.426 (ref) |
| original22-v1/runtime/simd-memory-full | wasmedge-0.18.1-interpreter | 73.159 (-36.65%) † | 63.663 (-27.20%) | 75.384 (-38.52%) | 151.44 (-69.40%) | NM | NM | 78.093 (-40.66%) | 46.344 (ref) |
| original22-v1/runtime/simd-memory-widen | wasmedge-0.18.1-interpreter | 58.521 (-41.37%) | 54.515 (-37.06%) | 53.441 (-35.80%) | 143.93 (-76.16%) | NM | NM | 61.223 (-43.96%) | 34.311 (ref) |
| original22-v1/runtime/simd-memory-lanes | wasmedge-0.18.1-interpreter | 92.178 (-39.86%) † | 83.525 (-33.63%) † | 87.076 (-36.33%) † | 188.34 (-70.57%) † | NM | NM | 90.650 (-38.84%) † | 55.437 (ref) † |
| original22-v1/gc/i31-roundtrip | wasmedge-0.18.1-interpreter | 35.025 (-32.10%) † | 29.950 (-20.59%) † | 30.565 (-22.19%) † | US | NM | NM | 45.465 (-47.69%) | 23.783 (ref) |
| original22-v1/gc/struct-access | wasmedge-0.18.1-interpreter | 40.290 (-23.42%) † | 38.025 (-18.86%) † | 37.888 (-18.57%) † | US | NM | NM | 56.160 (-45.06%) | 30.852 (ref) |
| original22-v1/gc/struct-allocation | wasmedge-0.18.1-interpreter | 91.393 (-19.75%) † | 99.063 (-25.97%) † | 99.985 (-26.65%) † | US | NM | NM | 108.57 (-32.45%) † | 73.339 (ref) † |
| original22-v1/gc/struct-indirect-access | wasmedge-0.18.1-interpreter | 75.252 (-31.17%) | 71.769 (-27.83%) † | 70.679 (-26.71%) † | US | NM | NM | 98.356 (-47.34%) | 51.798 (ref) |
| original22-v1/threads/private-memory | wasmedge-0.18.1-interpreter | 21.487 (+6.14%) | 19.160 (+19.03%) | 19.304 (+18.14%) | 64.875 (-64.85%) | NM | NM | 40.035 (-43.03%) | 22.806 (ref) |
| original22-v1/threads/shared-memory | wasmedge-0.18.1-interpreter | 21.646 (+4.89%) † | 19.131 (+18.68%) | 18.969 (+19.69%) | 65.008 (-65.07%) | NM | NM | 40.038 (-43.29%) | 22.704 (ref) |
| original22-v1/threads/atomic-rmw | wasmedge-0.18.1-interpreter | 27.856 (-37.55%) | 28.102 (-38.10%) † | 27.615 (-37.00%) | 46.798 (-62.83%) † | NM | NM | 33.015 (-47.31%) | 17.396 (ref) |
| original22-v1/threads/notify-empty | wasmedge-0.18.1-interpreter | 24.550 (-13.09%) † | 23.473 (-9.11%) † | 22.565 (-5.45%) † | 43.423 (-50.87%) † | NM | NM | 40.638 (-47.50%) † | 21.335 (ref) † |
| original22-v1/threads/wait-mismatch | wasmedge-0.18.1-interpreter | 36.477 (-31.34%) † | 33.290 (-24.77%) † | 33.342 (-24.89%) † | 71.404 (-64.93%) † | NM | NM | 43.946 (-43.01%) | 25.044 (ref) |
| original22-v1/threads/wait-timeout | wasmedge-0.18.1-interpreter | 35.800 (+114.98%) † | 31.604 (+143.52%) † | 31.881 (+141.40%) † | 332.19 (-76.83%) | NM | NM | 98.006 (-21.47%) | 76.963 (ref) |
| original22-v1/applications/sha256-clang | wasmedge-0.18.1-interpreter | 18,577.29 (-30.18%) | 15,801.22 (-17.91%) | 15,871.92 (-18.28%) | 40,679.77 (-68.12%) † | NM | NM | 26,326.62 (-50.73%) | 12,970.68 (ref) |
| original22-v1/applications/sha256-rust | wasmedge-0.18.1-interpreter | 18,089.42 (-31.51%) | 15,392.56 (-19.50%) | 15,451.41 (-19.81%) † | 40,880.11 (-69.69%) | NM | NM | 24,473.17 (-49.37%) | 12,390.25 (ref) |
| gc-sustained-v1/gc/i31-roundtrip | wasmedge-0.18.1-interpreter | 32.880 (-30.93%) | 29.043 (-21.81%) | 29.044 (-21.81%) † | US | NM | NM | 44.500 (-48.97%) | 22.710 (ref) |
| gc-sustained-v1/gc/struct-access | wasmedge-0.18.1-interpreter | 38.368 (-24.22%) † | 36.400 (-20.12%) † | 36.337 (-19.99%) † | US | NM | NM | 54.571 (-46.72%) † | 29.075 (ref) † |
| gc-sustained-v1/gc/struct-allocation | wasmedge-0.18.1-interpreter | 92.379 (-23.47%) † | 102.48 (-31.01%) | 103.19 (-31.49%) | US | NM | NM | 104.89 (-32.60%) | 70.698 (ref) |
| gc-sustained-v1/gc/struct-indirect-access | wasmedge-0.18.1-interpreter | 72.250 (-31.14%) | 68.071 (-26.91%) | 68.497 (-27.37%) | US | NM | NM | 95.405 (-47.85%) | 49.751 (ref) |
| threads-sustained-v2/threads/private-memory | wasmedge-0.18.1-interpreter | 20.531 (+7.73%) | 18.485 (+19.66%) | 18.424 (+20.05%) | 62.207 (-64.44%) † | NM | NM | 38.770 (-42.95%) | 22.118 (ref) |
| threads-sustained-v2/threads/shared-memory | wasmedge-0.18.1-interpreter | 20.463 (+8.08%) † | 18.460 (+19.81%) | 18.362 (+20.45%) † | 63.688 (-65.27%) | NM | NM | 39.032 (-43.34%) | 22.117 (ref) |
| threads-sustained-v2/threads/atomic-rmw | wasmedge-0.18.1-interpreter | 26.698 (-37.70%) | 27.116 (-38.67%) † | 26.642 (-37.57%) † | 46.987 (-64.60%) † | NM | NM | 32.224 (-48.39%) | 16.632 (ref) |
| threads-sustained-v2/threads/notify-empty | wasmedge-0.18.1-interpreter | 23.437 (-12.95%) | 22.282 (-8.44%) † | 21.778 (-6.32%) † | 41.495 (-50.83%) † | NM | NM | 40.098 (-49.12%) | 20.402 (ref) |
| threads-sustained-v2/threads/wait-mismatch | wasmedge-0.18.1-interpreter | 36.205 (-33.67%) † | 33.036 (-27.31%) † | 32.699 (-26.56%) | 69.999 (-65.69%) † | NM | NM | 43.713 (-45.06%) † | 24.014 (ref) |
| threads-sustained-v2/threads/wait-timeout | wasmedge-0.18.1-interpreter | 33.666 (+125.56%) † | 30.969 (+145.21%) | 31.756 (+139.13%) † | 332.35 (-77.15%) | NM | NM | 95.430 (-20.43%) | 75.937 (ref) |
| generic-calls-v1/core/generic-calls | wasmedge-0.18.1-interpreter | 96.679 (-35.35%) | 85.717 (-27.08%) | 86.930 (-28.10%) | 186.52 (-66.49%) † | NM | NM | 117.67 (-46.88%) | 62.502 (ref) |

## Complete Preview 3 HTTP

Wasmtime is the reference for the equivalent request contract. The driver validates
status, headers, full body, trailers, CRC and joined completion. The host-only Go
control has its own contract and remains available in the machine-readable results.
The instruction rows report quantity differences. The CPU and throughput rows
report speed differences. A/A is applied separately for each recorded metric.

| Payload / metric | wasmvm Go | wasmvm LLGo (pinned) | wasmvm LLGo 42efacae | wazero 1.12.0 | Wasmtime 48.0.1 | Wasmer 7.3.0 | WasmEdge 0.17.1 | WasmEdge 0.18.1 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B CPU µs/request | 62.212 (-52.24%) † | 128.27 (-76.84%) † | 127.34 (-76.67%) † | NM | 29.713 (ref) † | NM | NM | NM |
| 16 B instructions/request | 647,971.32 (+199.55%) | 949,303.41 (+338.85%) | 942,737.20 (+335.81%) | NM | 216,316.75 (ref) | NM | NM | NM |
| 16 B requests/s | 61,079.78 (-9.91%) | 33,561.90 (-50.50%) | 34,166.12 (-49.61%) † | NM | 67,800.90 (ref) | NM | NM | NM |
| 65536 B CPU µs/request | 87.554 (-55.02%) † | 158.05 (-75.08%) † | 156.37 (-74.81%) † | NM | 39.384 (ref) † | NM | NM | NM |
| 65536 B instructions/request | 926,461.86 (+204.30%) † | 1,264,767.37 (+315.42%) † | 1,263,762.17 (+315.09%) † | NM | 304,457.85 (ref) † | NM | NM | NM |
| 65536 B requests/s | 42,740.65 (-26.87%) | 28,000.16 (-52.09%) | 28,134.91 (-51.86%) | NM | 58,445.12 (ref) | NM | NM | NM |

## Own-source and LLGo upgrade contrasts

Ratios compare warm time. Values below one improve. The A/A flag applies
to each ratio. A resolved adverse interval requires its lower bound above one.

### Native

| Workload | Go engine change | Pinned LLGo engine change | Refreshed LLGo engine change | LLGo toolchain change | Refreshed LLGo vs same-source Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 3,562.50 → 3,500.00 · 1.010 [0.939, 1.094] · Δ +1.03% · speed -1.02% [-8.59%, +6.47%] · A/A unresolved | 2,812.50 → 3,271.00 · 0.999 [0.898, 1.094] · Δ -0.11% · speed +0.11% [-8.61%, +11.35%] · A/A unresolved | 3,104.00 → 2,750.00 · 0.971 [0.847, 1.101] · Δ -2.94% · speed +3.03% [-9.19%, +18.01%] · A/A unresolved | 3,271.00 → 2,750.00 · 0.932 [0.858, 1.027] · Δ -6.76% · speed +7.25% [-2.61%, +16.60%] · A/A unresolved | 3,500.00 → 2,750.00 · 0.817 [0.777, 0.862] · Δ -18.26% · speed +22.35% [+16.03%, +28.77%] · A/A unresolved |
| original22-v1/core/integer-loop | 1.592 → 1.601 · 1.005 [1.002, 1.009] · Δ +0.55% · speed -0.55% [-0.89%, -0.21%] · A/A resolved | 1.580 → 1.584 · 0.995 [0.990, 1.000] · Δ -0.49% · speed +0.49% [+0.01%, +1.04%] · A/A resolved | 1.588 → 1.580 · 1.001 [0.992, 1.010] · Δ +0.07% · speed -0.07% [-0.97%, +0.83%] · A/A resolved | 1.584 → 1.580 · 1.003 [0.996, 1.011] · Δ +0.29% · speed -0.29% [-1.11%, +0.40%] · A/A resolved | 1.601 → 1.580 · 0.989 [0.984, 0.996] · Δ -1.07% · speed +1.08% [+0.39%, +1.65%] · A/A resolved |
| original22-v1/core/function-calls | 0.811 → 0.812 · 1.004 [0.997, 1.010] · Δ +0.44% · speed -0.44% [-1.01%, +0.30%] · A/A resolved | 0.797 → 0.794 · 1.001 [0.994, 1.009] · Δ +0.09% · speed -0.09% [-0.91%, +0.57%] · A/A resolved | 0.794 → 0.795 · 1.004 [0.997, 1.010] · Δ +0.37% · speed -0.37% [-0.98%, +0.34%] · A/A resolved | 0.794 → 0.795 · 1.000 [0.989, 1.009] · Δ +0.02% · speed -0.02% [-0.93%, +1.10%] · A/A resolved | 0.812 → 0.795 · 0.977 [0.975, 0.980] · Δ -2.30% · speed +2.36% [+2.02%, +2.61%] · A/A resolved |
| original22-v1/core/memory-load-store | 0.452 → 0.456 · 1.006 [0.984, 1.030] · Δ +0.62% · speed -0.62% [-2.88%, +1.59%] · A/A unresolved | 0.441 → 0.440 · 1.007 [0.988, 1.036] · Δ +0.65% · speed -0.65% [-3.51%, +1.20%] · A/A unresolved | 0.440 → 0.444 · 1.007 [0.993, 1.021] · Δ +0.72% · speed -0.71% [-2.07%, +0.72%] · A/A resolved | 0.440 → 0.444 · 0.991 [0.956, 1.017] · Δ -0.90% · speed +0.91% [-1.66%, +4.60%] · A/A unresolved | 0.456 → 0.444 · 0.965 [0.952, 0.979] · Δ -3.52% · speed +3.64% [+2.17%, +5.08%] · A/A unresolved |
| original22-v1/memory/narrow-unaligned | 0.672 → 0.675 · 1.001 [0.987, 1.012] · Δ +0.06% · speed -0.06% [-1.22%, +1.28%] · A/A resolved | 0.654 → 0.657 · 1.003 [0.994, 1.013] · Δ +0.32% · speed -0.31% [-1.25%, +0.64%] · A/A resolved | 0.657 → 0.655 · 0.997 [0.987, 1.008] · Δ -0.25% · speed +0.25% [-0.80%, +1.27%] · A/A resolved | 0.657 → 0.655 · 0.997 [0.990, 1.006] · Δ -0.26% · speed +0.26% [-0.60%, +1.06%] · A/A resolved | 0.675 → 0.655 · 0.972 [0.964, 0.979] · Δ -2.84% · speed +2.93% [+2.12%, +3.69%] · A/A resolved |
| original22-v1/memory/dependent-address | 1.891 → 1.887 · 1.000 [0.990, 1.012] · Δ +0.04% · speed -0.04% [-1.14%, +1.03%] · A/A resolved | 1.875 → 1.874 · 1.001 [0.993, 1.010] · Δ +0.10% · speed -0.10% [-0.98%, +0.70%] · A/A resolved | 1.877 → 1.869 · 0.998 [0.991, 1.005] · Δ -0.24% · speed +0.24% [-0.54%, +0.94%] · A/A resolved | 1.874 → 1.869 · 0.995 [0.985, 1.005] · Δ -0.54% · speed +0.54% [-0.54%, +1.52%] · A/A resolved | 1.887 → 1.869 · 0.984 [0.971, 0.996] · Δ -1.57% · speed +1.60% [+0.38%, +2.99%] · A/A resolved |
| original22-v1/memory/float64-unaligned | 0.762 → 0.764 · 1.006 [1.004, 1.009] · Δ +0.65% · speed -0.64% [-0.85%, -0.44%] · A/A unresolved | 0.749 → 0.748 · 0.995 [0.983, 1.008] · Δ -0.50% · speed +0.51% [-0.81%, +1.77%] · A/A resolved | 0.747 → 0.750 · 1.006 [1.003, 1.008] · Δ +0.55% · speed -0.55% [-0.83%, -0.28%] · A/A resolved | 0.748 → 0.750 · 1.006 [0.995, 1.016] · Δ +0.62% · speed -0.62% [-1.54%, +0.49%] · A/A resolved | 0.764 → 0.750 · 0.980 [0.976, 0.983] · Δ -2.01% · speed +2.05% [+1.73%, +2.41%] · A/A unresolved |
| original22-v1/runtime/simd-memory-full | 3.402 → 3.405 · 1.003 [0.998, 1.008] · Δ +0.33% · speed -0.33% [-0.82%, +0.19%] · A/A resolved | 3.398 → 3.406 · 1.001 [0.991, 1.009] · Δ +0.09% · speed -0.09% [-0.87%, +0.88%] · A/A resolved | 3.396 → 3.388 · 0.998 [0.993, 1.004] · Δ -0.19% · speed +0.19% [-0.35%, +0.75%] · A/A resolved | 3.406 → 3.388 · 0.998 [0.990, 1.009] · Δ -0.23% · speed +0.23% [-0.86%, +1.05%] · A/A resolved | 3.405 → 3.388 · 0.994 [0.988, 1.000] · Δ -0.56% · speed +0.56% [+0.00%, +1.17%] · A/A resolved |
| original22-v1/runtime/simd-memory-widen | 1.073 → 1.072 · 1.000 [0.995, 1.006] · Δ +0.03% · speed -0.03% [-0.55%, +0.51%] · A/A resolved | 1.055 → 1.061 · 1.004 [0.994, 1.014] · Δ +0.42% · speed -0.42% [-1.35%, +0.59%] · A/A resolved | 1.055 → 1.059 · 1.000 [0.991, 1.008] · Δ +0.02% · speed -0.02% [-0.79%, +0.93%] · A/A resolved | 1.061 → 1.059 · 1.000 [0.990, 1.011] · Δ -0.02% · speed +0.02% [-1.13%, +1.01%] · A/A resolved | 1.072 → 1.059 · 0.986 [0.978, 0.992] · Δ -1.44% · speed +1.47% [+0.77%, +2.22%] · A/A resolved |
| original22-v1/runtime/simd-memory-lanes | 5.058 → 5.091 · 1.003 [1.000, 1.007] · Δ +0.34% · speed -0.34% [-0.69%, +0.01%] · A/A resolved | 5.001 → 5.008 · 0.997 [0.988, 1.008] · Δ -0.28% · speed +0.28% [-0.82%, +1.26%] · A/A resolved | 5.033 → 5.047 · 1.002 [0.996, 1.007] · Δ +0.16% · speed -0.16% [-0.72%, +0.41%] · A/A resolved | 5.008 → 5.047 · 1.007 [1.001, 1.012] · Δ +0.65% · speed -0.65% [-1.18%, -0.09%] · A/A resolved | 5.091 → 5.047 · 0.991 [0.987, 0.995] · Δ -0.91% · speed +0.92% [+0.54%, +1.29%] · A/A resolved |
| original22-v1/gc/i31-roundtrip | 2.231 → 2.194 · 0.990 [0.963, 1.023] · Δ -1.00% · speed +1.01% [-2.29%, +3.80%] · A/A unresolved | 1.935 → 1.915 · 0.990 [0.971, 1.008] · Δ -1.04% · speed +1.05% [-0.83%, +3.00%] · A/A unresolved | 1.887 → 1.954 · 1.024 [0.995, 1.052] · Δ +2.36% · speed -2.30% [-4.96%, +0.46%] · A/A unresolved | 1.915 → 1.954 · 1.017 [0.990, 1.043] · Δ +1.67% · speed -1.64% [-4.09%, +1.00%] · A/A unresolved | 2.194 → 1.954 · 0.882 [0.847, 0.917] · Δ -11.77% · speed +13.34% [+9.02%, +18.13%] · A/A unresolved |
| original22-v1/gc/struct-access | 2.633 → 2.744 · 0.987 [0.899, 1.071] · Δ -1.29% · speed +1.31% [-6.59%, +11.19%] · A/A unresolved | 1.831 → 1.875 · 1.023 [0.994, 1.053] · Δ +2.27% · speed -2.22% [-4.99%, +0.55%] · A/A unresolved | 1.888 → 1.756 · 0.973 [0.950, 0.989] · Δ -2.71% · speed +2.79% [+1.10%, +5.21%] · A/A unresolved | 1.875 → 1.756 · 0.970 [0.950, 0.989] · Δ -3.00% · speed +3.10% [+1.07%, +5.21%] · A/A unresolved | 2.744 → 1.756 · 0.672 [0.615, 0.723] · Δ -32.77% · speed +48.74% [+38.22%, +62.62%] · A/A unresolved |
| original22-v1/gc/struct-allocation | 28.433 → 29.374 · 1.019 [1.003, 1.041] · Δ +1.87% · speed -1.84% [-3.92%, -0.35%] · A/A unresolved | 27.888 → 27.725 · 1.000 [0.981, 1.019] · Δ -0.04% · speed +0.04% [-1.82%, +1.93%] · A/A unresolved | 28.184 → 28.167 · 1.013 [1.001, 1.022] · Δ +1.27% · speed -1.25% [-2.16%, -0.14%] · A/A unresolved | 27.725 → 28.167 · 1.019 [1.003, 1.034] · Δ +1.85% · speed -1.82% [-3.33%, -0.26%] · A/A unresolved | 29.374 → 28.167 · 0.963 [0.950, 0.980] · Δ -3.72% · speed +3.86% [+2.08%, +5.32%] · A/A unresolved |
| original22-v1/gc/struct-indirect-access | 3.321 → 3.360 · 1.096 [1.014, 1.174] · Δ +9.64% · speed -8.79% [-14.84%, -1.37%] · A/A unresolved | 2.410 → 2.473 · 1.018 [0.985, 1.048] · Δ +1.78% · speed -1.75% [-4.61%, +1.48%] · A/A unresolved | 2.408 → 2.487 · 1.031 [1.017, 1.046] · Δ +3.11% · speed -3.01% [-4.42%, -1.70%] · A/A unresolved | 2.473 → 2.487 · 1.008 [0.966, 1.055] · Δ +0.79% · speed -0.79% [-5.25%, +3.56%] · A/A unresolved | 3.360 → 2.487 · 0.707 [0.676, 0.736] · Δ -29.31% · speed +41.46% [+35.89%, +47.90%] · A/A unresolved |
| original22-v1/threads/private-memory | 1.152 → 1.115 · 1.024 [0.894, 1.126] · Δ +2.44% · speed -2.38% [-11.20%, +11.87%] · A/A unresolved | 0.742 → 0.785 · 1.042 [0.973, 1.118] · Δ +4.23% · speed -4.06% [-10.54%, +2.77%] · A/A unresolved | 0.771 → 0.787 · 1.029 [1.006, 1.057] · Δ +2.87% · speed -2.79% [-5.36%, -0.61%] · A/A unresolved | 0.785 → 0.787 · 1.014 [0.971, 1.065] · Δ +1.41% · speed -1.39% [-6.12%, +2.97%] · A/A unresolved | 1.115 → 0.787 · 0.706 [0.687, 0.726] · Δ -29.40% · speed +41.64% [+37.83%, +45.60%] · A/A unresolved |
| original22-v1/threads/shared-memory | 2.285 → 2.463 · 1.076 [1.003, 1.157] · Δ +7.57% · speed -7.04% [-13.60%, -0.28%] · A/A unresolved | 1.923 → 1.940 · 0.990 [0.961, 1.020] · Δ -0.99% · speed +1.00% [-1.96%, +4.11%] · A/A unresolved | 1.823 → 1.829 · 1.008 [0.949, 1.091] · Δ +0.83% · speed -0.83% [-8.32%, +5.38%] · A/A unresolved | 1.940 → 1.829 · 0.974 [0.921, 1.023] · Δ -2.60% · speed +2.67% [-2.22%, +8.59%] · A/A unresolved | 2.463 → 1.829 · 0.753 [0.705, 0.809] · Δ -24.68% · speed +32.77% [+23.58%, +41.94%] · A/A unresolved |
| original22-v1/threads/atomic-rmw | 2.875 → 2.925 · 0.997 [0.910, 1.087] · Δ -0.25% · speed +0.25% [-7.96%, +9.84%] · A/A unresolved | 2.419 → 2.492 · 1.018 [0.976, 1.060] · Δ +1.84% · speed -1.80% [-5.67%, +2.45%] · A/A unresolved | 2.531 → 2.469 · 0.972 [0.923, 1.005] · Δ -2.75% · speed +2.83% [-0.54%, +8.29%] · A/A unresolved | 2.492 → 2.469 · 1.011 [0.972, 1.063] · Δ +1.09% · speed -1.08% [-5.90%, +2.84%] · A/A unresolved | 2.925 → 2.469 · 0.840 [0.785, 0.899] · Δ -15.99% · speed +19.03% [+11.28%, +27.38%] · A/A unresolved |
| original22-v1/threads/notify-empty | 1.860 → 2.031 · 1.062 [0.982, 1.173] · Δ +6.15% · speed -5.79% [-14.76%, +1.86%] · A/A unresolved | 1.525 → 1.598 · 1.043 [0.993, 1.104] · Δ +4.31% · speed -4.13% [-9.45%, +0.66%] · A/A unresolved | 1.498 → 1.523 · 1.014 [0.985, 1.054] · Δ +1.40% · speed -1.38% [-5.16%, +1.57%] · A/A unresolved | 1.598 → 1.523 · 0.975 [0.926, 1.035] · Δ -2.54% · speed +2.61% [-3.37%, +7.96%] · A/A unresolved | 2.031 → 1.523 · 0.765 [0.721, 0.813] · Δ -23.51% · speed +30.73% [+23.02%, +38.74%] · A/A unresolved |
| original22-v1/threads/wait-mismatch | 2.327 → 2.390 · 1.029 [0.939, 1.126] · Δ +2.92% · speed -2.84% [-11.20%, +6.45%] · A/A unresolved | 1.746 → 1.760 · 1.006 [0.978, 1.038] · Δ +0.55% · speed -0.55% [-3.68%, +2.21%] · A/A unresolved | 1.712 → 1.773 · 0.983 [0.942, 1.031] · Δ -1.69% · speed +1.72% [-3.00%, +6.17%] · A/A unresolved | 1.760 → 1.773 · 0.989 [0.924, 1.060] · Δ -1.10% · speed +1.11% [-5.68%, +8.25%] · A/A unresolved | 2.390 → 1.773 · 0.741 [0.675, 0.829] · Δ -25.94% · speed +35.03% [+20.59%, +48.25%] · A/A unresolved |
| original22-v1/threads/wait-timeout | 2.788 → 2.821 · 1.051 [0.953, 1.209] · Δ +5.09% · speed -4.84% [-17.28%, +4.94%] · A/A unresolved | 1.746 → 1.698 · 0.992 [0.949, 1.025] · Δ -0.75% · speed +0.76% [-2.45%, +5.36%] · A/A unresolved | 1.719 → 1.696 · 0.998 [0.958, 1.034] · Δ -0.19% · speed +0.20% [-3.30%, +4.37%] · A/A unresolved | 1.698 → 1.696 · 1.002 [0.977, 1.027] · Δ +0.25% · speed -0.24% [-2.64%, +2.34%] · A/A unresolved | 2.821 → 1.696 · 0.623 [0.585, 0.679] · Δ -37.73% · speed +60.59% [+47.33%, +70.95%] · A/A unresolved |
| original22-v1/applications/sha256-clang | 204.07 → 205.47 · 1.006 [0.997, 1.016] · Δ +0.55% · speed -0.55% [-1.56%, +0.32%] · A/A resolved | 202.55 → 203.48 · 1.009 [0.999, 1.019] · Δ +0.88% · speed -0.87% [-1.89%, +0.09%] · A/A resolved | 200.64 → 201.93 · 1.004 [0.999, 1.008] · Δ +0.39% · speed -0.39% [-0.76%, +0.09%] · A/A resolved | 203.48 → 201.93 · 0.995 [0.984, 1.002] · Δ -0.54% · speed +0.54% [-0.24%, +1.58%] · A/A resolved | 205.47 → 201.93 · 0.986 [0.983, 0.991] · Δ -1.35% · speed +1.37% [+0.86%, +1.68%] · A/A resolved |
| original22-v1/applications/sha256-rust | 214.67 → 214.73 · 0.995 [0.991, 1.001] · Δ -0.46% · speed +0.46% [-0.14%, +0.89%] · A/A resolved | 211.94 → 211.87 · 0.999 [0.990, 1.009] · Δ -0.15% · speed +0.15% [-0.87%, +1.05%] · A/A resolved | 210.86 → 210.33 · 0.999 [0.995, 1.004] · Δ -0.08% · speed +0.08% [-0.38%, +0.51%] · A/A resolved | 211.87 → 210.33 · 0.993 [0.982, 1.003] · Δ -0.66% · speed +0.66% [-0.28%, +1.86%] · A/A resolved | 214.73 → 210.33 · 0.983 [0.975, 0.992] · Δ -1.66% · speed +1.68% [+0.81%, +2.55%] · A/A resolved |
| gc-sustained-v1/gc/i31-roundtrip | 1.384 → 1.385 · 1.002 [0.999, 1.004] · Δ +0.17% · speed -0.17% [-0.43%, +0.10%] · A/A resolved | 1.368 → 1.375 · 1.002 [0.994, 1.009] · Δ +0.22% · speed -0.22% [-0.91%, +0.56%] · A/A unresolved | 1.370 → 1.370 · 0.999 [0.996, 1.002] · Δ -0.12% · speed +0.12% [-0.19%, +0.45%] · A/A resolved | 1.375 → 1.370 · 0.997 [0.990, 1.003] · Δ -0.32% · speed +0.32% [-0.34%, +0.99%] · A/A unresolved | 1.385 → 1.370 · 0.986 [0.982, 0.990] · Δ -1.43% · speed +1.45% [+1.04%, +1.83%] · A/A resolved |
| gc-sustained-v1/gc/struct-access | 1.073 → 1.074 · 1.004 [0.995, 1.014] · Δ +0.43% · speed -0.43% [-1.35%, +0.48%] · A/A resolved | 1.056 → 1.060 · 1.004 [0.996, 1.012] · Δ +0.41% · speed -0.41% [-1.23%, +0.38%] · A/A resolved | 1.056 → 1.060 · 1.005 [0.999, 1.010] · Δ +0.45% · speed -0.45% [-0.98%, +0.05%] · A/A resolved | 1.060 → 1.060 · 1.003 [0.997, 1.009] · Δ +0.30% · speed -0.30% [-0.91%, +0.27%] · A/A resolved | 1.074 → 1.060 · 0.987 [0.979, 0.995] · Δ -1.34% · speed +1.36% [+0.52%, +2.16%] · A/A resolved |
| gc-sustained-v1/gc/struct-allocation | 32.074 → 32.025 · 0.999 [0.990, 1.010] · Δ -0.10% · speed +0.10% [-1.02%, +1.02%] · A/A unresolved | 33.136 → 33.225 · 0.992 [0.959, 1.020] · Δ -0.80% · speed +0.80% [-1.97%, +4.29%] · A/A unresolved | 32.881 → 33.366 · 1.037 [1.007, 1.094] · Δ +3.67% · speed -3.54% [-8.61%, -0.67%] · A/A unresolved | 33.225 → 33.366 · 1.028 [0.997, 1.082] · Δ +2.84% · speed -2.76% [-7.61%, +0.33%] · A/A unresolved | 32.025 → 33.366 · 1.058 [1.027, 1.103] · Δ +5.80% · speed -5.48% [-9.38%, -2.59%] · A/A unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 1.696 → 1.689 · 1.004 [0.980, 1.029] · Δ +0.40% · speed -0.40% [-2.79%, +2.05%] · A/A unresolved | 1.660 → 1.644 · 0.991 [0.965, 1.021] · Δ -0.87% · speed +0.88% [-2.05%, +3.64%] · A/A unresolved | 1.676 → 1.637 · 0.997 [0.978, 1.019] · Δ -0.34% · speed +0.34% [-1.89%, +2.23%] · A/A unresolved | 1.644 → 1.637 · 1.004 [0.980, 1.027] · Δ +0.36% · speed -0.36% [-2.59%, +2.01%] · A/A unresolved | 1.689 → 1.637 · 0.974 [0.950, 0.999] · Δ -2.57% · speed +2.64% [+0.15%, +5.28%] · A/A unresolved |
| threads-sustained-v2/threads/private-memory | 0.447 → 0.452 · 1.020 [0.977, 1.065] · Δ +1.99% · speed -1.95% [-6.14%, +2.38%] · A/A unresolved | 0.340 → 0.346 · 0.999 [0.968, 1.028] · Δ -0.14% · speed +0.14% [-2.75%, +3.27%] · A/A unresolved | 0.343 → 0.346 · 0.996 [0.958, 1.042] · Δ -0.36% · speed +0.37% [-4.00%, +4.44%] · A/A unresolved | 0.346 → 0.346 · 1.004 [0.962, 1.055] · Δ +0.37% · speed -0.37% [-5.19%, +3.93%] · A/A unresolved | 0.452 → 0.346 · 0.745 [0.699, 0.791] · Δ -25.48% · speed +34.20% [+26.40%, +43.03%] · A/A unresolved |
| threads-sustained-v2/threads/shared-memory | 1.415 → 1.454 · 1.037 [0.992, 1.093] · Δ +3.70% · speed -3.57% [-8.54%, +0.77%] · A/A unresolved | 1.304 → 1.304 · 0.993 [0.979, 1.007] · Δ -0.73% · speed +0.74% [-0.66%, +2.18%] · A/A unresolved | 1.325 → 1.307 · 0.975 [0.958, 0.992] · Δ -2.45% · speed +2.51% [+0.80%, +4.40%] · A/A unresolved | 1.304 → 1.307 · 1.002 [0.991, 1.011] · Δ +0.24% · speed -0.24% [-1.08%, +0.95%] · A/A resolved | 1.454 → 1.307 · 0.905 [0.886, 0.927] · Δ -9.49% · speed +10.49% [+7.84%, +12.80%] · A/A unresolved |
| threads-sustained-v2/threads/atomic-rmw | 2.093 → 2.093 · 1.008 [0.980, 1.033] · Δ +0.77% · speed -0.76% [-3.16%, +2.07%] · A/A unresolved | 1.904 → 1.920 · 1.010 [0.994, 1.029] · Δ +1.01% · speed -1.00% [-2.82%, +0.65%] · A/A unresolved | 1.935 → 1.895 · 0.976 [0.959, 0.996] · Δ -2.35% · speed +2.41% [+0.39%, +4.26%] · A/A unresolved | 1.920 → 1.895 · 0.987 [0.968, 1.006] · Δ -1.28% · speed +1.30% [-0.55%, +3.32%] · A/A unresolved | 2.093 → 1.895 · 0.906 [0.887, 0.928] · Δ -9.38% · speed +10.35% [+7.73%, +12.75%] · A/A unresolved |
| threads-sustained-v2/threads/notify-empty | 1.111 → 1.121 · 1.016 [0.995, 1.039] · Δ +1.63% · speed -1.60% [-3.77%, +0.55%] · A/A unresolved | 0.964 → 0.956 · 1.006 [0.981, 1.035] · Δ +0.55% · speed -0.55% [-3.40%, +1.89%] · A/A unresolved | 0.978 → 0.954 · 0.984 [0.968, 1.003] · Δ -1.56% · speed +1.58% [-0.34%, +3.33%] · A/A unresolved | 0.956 → 0.954 · 0.993 [0.969, 1.017] · Δ -0.66% · speed +0.66% [-1.71%, +3.23%] · A/A unresolved | 1.121 → 0.954 · 0.850 [0.823, 0.885] · Δ -15.01% · speed +17.66% [+12.96%, +21.47%] · A/A unresolved |
| threads-sustained-v2/threads/wait-mismatch | 1.282 → 1.273 · 0.985 [0.967, 1.005] · Δ -1.48% · speed +1.50% [-0.52%, +3.45%] · A/A unresolved | 1.129 → 1.111 · 0.973 [0.899, 1.050] · Δ -2.75% · speed +2.82% [-4.77%, +11.22%] · A/A unresolved | 1.132 → 1.123 · 1.006 [0.979, 1.047] · Δ +0.64% · speed -0.63% [-4.45%, +2.18%] · A/A unresolved | 1.111 → 1.123 · 1.052 [0.988, 1.113] · Δ +5.23% · speed -4.97% [-10.12%, +1.19%] · A/A unresolved | 1.273 → 1.123 · 0.901 [0.877, 0.926] · Δ -9.92% · speed +11.02% [+8.02%, +14.03%] · A/A unresolved |
| threads-sustained-v2/threads/wait-timeout | 1.151 → 1.119 · 0.972 [0.960, 0.984] · Δ -2.78% · speed +2.86% [+1.60%, +4.20%] · A/A unresolved | 0.968 → 0.964 · 0.986 [0.962, 1.007] · Δ -1.35% · speed +1.37% [-0.66%, +3.90%] · A/A unresolved | 0.959 → 0.970 · 1.001 [0.979, 1.023] · Δ +0.12% · speed -0.12% [-2.26%, +2.15%] · A/A unresolved | 0.964 → 0.970 · 0.999 [0.976, 1.021] · Δ -0.05% · speed +0.05% [-2.10%, +2.51%] · A/A resolved | 1.119 → 0.970 · 0.859 [0.835, 0.882] · Δ -14.07% · speed +16.38% [+13.37%, +19.76%] · A/A unresolved |
| generic-calls-v1/core/generic-calls | 1.043 → 1.046 · 1.002 [0.996, 1.009] · Δ +0.21% · speed -0.21% [-0.93%, +0.45%] · A/A resolved | 1.038 → 1.039 · 1.001 [0.993, 1.009] · Δ +0.07% · speed -0.07% [-0.84%, +0.69%] · A/A resolved | 1.036 → 1.039 · 1.002 [0.996, 1.009] · Δ +0.20% · speed -0.20% [-0.89%, +0.38%] · A/A resolved | 1.039 → 1.039 · 0.999 [0.989, 1.011] · Δ -0.07% · speed +0.07% [-1.13%, +1.10%] · A/A resolved | 1.046 → 1.039 · 0.993 [0.983, 1.005] · Δ -0.65% · speed +0.66% [-0.53%, +1.68%] · A/A resolved |

### Interpreter

| Workload | Go engine change | Pinned LLGo engine change | Refreshed LLGo engine change | LLGo toolchain change | Refreshed LLGo vs same-source Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | 2,562.50 → 2,416.50 · 0.997 [0.873, 1.132] · Δ -0.29% · speed +0.29% [-11.64%, +14.57%] · A/A unresolved | 2,188.00 → 2,354.00 · 1.056 [0.943, 1.226] · Δ +5.65% · speed -5.35% [-18.44%, +6.08%] · A/A unresolved | 2,146.00 → 2,291.00 · 1.038 [0.942, 1.148] · Δ +3.77% · speed -3.63% [-12.88%, +6.16%] · A/A unresolved | 2,354.00 → 2,291.00 · 0.973 [0.915, 1.035] · Δ -2.67% · speed +2.74% [-3.37%, +9.34%] · A/A unresolved | 2,416.50 → 2,291.00 · 0.893 [0.822, 0.968] · Δ -10.67% · speed +11.94% [+3.29%, +21.67%] · A/A unresolved |
| original22-v1/core/integer-loop | 28.515 → 28.606 · 1.006 [0.998, 1.015] · Δ +0.58% · speed -0.58% [-1.46%, +0.19%] · A/A resolved | 24.069 → 23.982 · 0.997 [0.989, 1.009] · Δ -0.27% · speed +0.27% [-0.85%, +1.15%] · A/A resolved | 23.011 → 22.958 · 0.996 [0.990, 1.001] · Δ -0.40% · speed +0.40% [-0.15%, +0.97%] · A/A resolved | 23.982 → 22.958 · 0.956 [0.949, 0.963] · Δ -4.41% · speed +4.61% [+3.81%, +5.34%] · A/A resolved | 28.606 → 22.958 · 0.799 [0.792, 0.805] · Δ -20.13% · speed +25.21% [+24.17%, +26.34%] · A/A resolved |
| original22-v1/core/function-calls | 45.710 → 46.368 · 1.014 [1.008, 1.019] · Δ +1.41% · speed -1.39% [-1.90%, -0.83%] · A/A resolved | 44.270 → 44.050 · 0.998 [0.991, 1.006] · Δ -0.20% · speed +0.20% [-0.60%, +0.87%] · A/A resolved | 42.789 → 43.103 · 1.003 [0.995, 1.010] · Δ +0.27% · speed -0.27% [-1.01%, +0.47%] · A/A resolved | 44.050 → 43.103 · 0.974 [0.967, 0.981] · Δ -2.60% · speed +2.67% [+1.98%, +3.39%] · A/A resolved | 46.368 → 43.103 · 0.929 [0.923, 0.937] · Δ -7.11% · speed +7.66% [+6.71%, +8.29%] · A/A resolved |
| original22-v1/core/memory-load-store | 63.994 → 64.131 · 1.002 [0.993, 1.009] · Δ +0.20% · speed -0.20% [-0.88%, +0.67%] · A/A resolved | 55.111 → 55.226 · 1.006 [1.001, 1.012] · Δ +0.61% · speed -0.61% [-1.20%, -0.09%] · A/A resolved | 55.301 → 54.982 · 0.994 [0.990, 0.998] · Δ -0.61% · speed +0.61% [+0.23%, +0.99%] · A/A resolved | 55.226 → 54.982 · 0.996 [0.994, 0.999] · Δ -0.40% · speed +0.40% [+0.12%, +0.64%] · A/A resolved | 64.131 → 54.982 · 0.858 [0.854, 0.863] · Δ -14.18% · speed +16.52% [+15.84%, +17.12%] · A/A resolved |
| original22-v1/memory/narrow-unaligned | 84.944 → 84.854 · 0.997 [0.988, 1.003] · Δ -0.31% · speed +0.32% [-0.33%, +1.26%] · A/A resolved | 80.492 → 72.927 · 0.902 [0.896, 0.908] · Δ -9.81% · speed +10.87% [+10.08%, +11.65%] · A/A resolved | 73.942 → 74.692 · 1.006 [0.999, 1.013] · Δ +0.65% · speed -0.64% [-1.29%, +0.07%] · A/A resolved | 72.927 → 74.692 · 1.028 [1.020, 1.035] · Δ +2.75% · speed -2.68% [-3.40%, -1.98%] · A/A resolved | 84.854 → 74.692 · 0.882 [0.874, 0.889] · Δ -11.82% · speed +13.40% [+12.44%, +14.40%] · A/A resolved |
| original22-v1/memory/dependent-address | 38.665 → 37.547 · 0.971 [0.966, 0.975] · Δ -2.92% · speed +3.01% [+2.59%, +3.47%] · A/A resolved | 32.315 → 32.407 · 1.002 [0.997, 1.008] · Δ +0.25% · speed -0.25% [-0.81%, +0.34%] · A/A resolved | 32.687 → 32.507 · 0.995 [0.989, 1.002] · Δ -0.46% · speed +0.46% [-0.16%, +1.10%] · A/A resolved | 32.407 → 32.507 · 1.003 [0.997, 1.007] · Δ +0.26% · speed -0.26% [-0.73%, +0.26%] · A/A resolved | 37.547 → 32.507 · 0.865 [0.861, 0.869] · Δ -13.51% · speed +15.62% [+15.13%, +16.09%] · A/A resolved |
| original22-v1/memory/float64-unaligned | 67.597 → 66.960 · 0.988 [0.985, 0.991] · Δ -1.16% · speed +1.17% [+0.92%, +1.52%] · A/A resolved | 57.275 → 57.143 · 0.997 [0.990, 1.003] · Δ -0.31% · speed +0.31% [-0.31%, +1.01%] · A/A resolved | 57.926 → 57.713 · 0.999 [0.990, 1.009] · Δ -0.12% · speed +0.12% [-0.85%, +0.97%] · A/A resolved | 57.143 → 57.713 · 1.014 [1.001, 1.030] · Δ +1.36% · speed -1.34% [-2.95%, -0.05%] · A/A resolved | 66.960 → 57.713 · 0.864 [0.855, 0.874] · Δ -13.56% · speed +15.68% [+14.40%, +17.02%] · A/A resolved |
| original22-v1/runtime/simd-memory-full | 73.462 → 73.159 · 0.998 [0.986, 1.006] · Δ -0.24% · speed +0.24% [-0.60%, +1.39%] · A/A unresolved | 75.947 → 63.663 · 0.840 [0.835, 0.845] · Δ -16.04% · speed +19.11% [+18.40%, +19.70%] · A/A resolved | 75.394 → 75.384 · 1.000 [0.996, 1.003] · Δ -0.04% · speed +0.04% [-0.31%, +0.41%] · A/A resolved | 63.663 → 75.384 · 1.187 [1.180, 1.194] · Δ +18.65% · speed -15.72% [-16.25%, -15.28%] · A/A resolved | 73.159 → 75.384 · 1.031 [1.024, 1.040] · Δ +3.09% · speed -3.00% [-3.80%, -2.36%] · A/A unresolved |
| original22-v1/runtime/simd-memory-widen | 58.896 → 58.521 · 0.995 [0.990, 1.001] · Δ -0.46% · speed +0.46% [-0.10%, +1.02%] · A/A resolved | 54.340 → 54.515 · 1.001 [0.995, 1.006] · Δ +0.12% · speed -0.12% [-0.59%, +0.48%] · A/A resolved | 53.250 → 53.441 · 1.005 [0.998, 1.011] · Δ +0.47% · speed -0.46% [-1.06%, +0.16%] · A/A resolved | 54.515 → 53.441 · 0.984 [0.977, 0.992] · Δ -1.57% · speed +1.60% [+0.84%, +2.38%] · A/A resolved | 58.521 → 53.441 · 0.912 [0.906, 0.919] · Δ -8.76% · speed +9.60% [+8.80%, +10.41%] · A/A resolved |
| original22-v1/runtime/simd-memory-lanes | 92.083 → 92.178 · 0.999 [0.994, 1.004] · Δ -0.08% · speed +0.08% [-0.43%, +0.61%] · A/A resolved | 84.106 → 83.525 · 0.994 [0.985, 1.003] · Δ -0.64% · speed +0.65% [-0.28%, +1.56%] · A/A resolved | 88.406 → 87.076 · 0.984 [0.976, 0.994] · Δ -1.56% · speed +1.59% [+0.65%, +2.50%] · A/A resolved | 83.525 → 87.076 · 1.041 [1.031, 1.051] · Δ +4.09% · speed -3.93% [-4.81%, -3.03%] · A/A resolved | 92.178 → 87.076 · 0.947 [0.940, 0.954] · Δ -5.35% · speed +5.65% [+4.87%, +6.40%] · A/A resolved |
| original22-v1/gc/i31-roundtrip | 34.477 → 35.025 · 1.001 [0.975, 1.026] · Δ +0.11% · speed -0.11% [-2.56%, +2.60%] · A/A unresolved | 30.023 → 29.950 · 0.990 [0.974, 1.010] · Δ -0.95% · speed +0.96% [-0.97%, +2.69%] · A/A unresolved | 30.269 → 30.565 · 1.003 [0.978, 1.023] · Δ +0.34% · speed -0.34% [-2.29%, +2.26%] · A/A unresolved | 29.950 → 30.565 · 1.012 [0.995, 1.027] · Δ +1.16% · speed -1.15% [-2.59%, +0.50%] · A/A unresolved | 35.025 → 30.565 · 0.872 [0.858, 0.889] · Δ -12.79% · speed +14.66% [+12.49%, +16.55%] · A/A unresolved |
| original22-v1/gc/struct-access | 40.773 → 40.290 · 0.980 [0.959, 1.000] · Δ -2.03% · speed +2.08% [+0.03%, +4.25%] · A/A unresolved | 37.917 → 38.025 · 1.004 [0.983, 1.023] · Δ +0.39% · speed -0.39% [-2.27%, +1.73%] · A/A unresolved | 37.835 → 37.888 · 1.000 [0.987, 1.012] · Δ -0.05% · speed +0.05% [-1.19%, +1.34%] · A/A unresolved | 38.025 → 37.888 · 1.003 [0.991, 1.021] · Δ +0.31% · speed -0.31% [-2.06%, +0.90%] · A/A unresolved | 40.290 → 37.888 · 0.941 [0.928, 0.952] · Δ -5.92% · speed +6.29% [+5.02%, +7.75%] · A/A unresolved |
| original22-v1/gc/struct-allocation | 91.592 → 91.393 · 0.996 [0.985, 1.005] · Δ -0.44% · speed +0.45% [-0.53%, +1.56%] · A/A resolved | 100.09 → 99.063 · 0.981 [0.969, 0.994] · Δ -1.88% · speed +1.91% [+0.63%, +3.24%] · A/A unresolved | 99.288 → 99.985 · 1.003 [0.990, 1.014] · Δ +0.33% · speed -0.33% [-1.41%, +1.05%] · A/A unresolved | 99.063 → 99.985 · 1.007 [0.993, 1.023] · Δ +0.73% · speed -0.73% [-2.26%, +0.75%] · A/A unresolved | 91.393 → 99.985 · 1.089 [1.077, 1.100] · Δ +8.86% · speed -8.14% [-9.13%, -7.19%] · A/A unresolved |
| original22-v1/gc/struct-indirect-access | 77.048 → 75.252 · 0.978 [0.964, 0.991] · Δ -2.24% · speed +2.30% [+0.93%, +3.73%] · A/A unresolved | 72.121 → 71.769 · 0.984 [0.969, 1.004] · Δ -1.57% · speed +1.60% [-0.36%, +3.15%] · A/A unresolved | 71.375 → 70.679 · 0.992 [0.980, 1.004] · Δ -0.77% · speed +0.78% [-0.38%, +2.06%] · A/A unresolved | 71.769 → 70.679 · 1.000 [0.991, 1.010] · Δ +0.02% · speed -0.02% [-0.99%, +0.89%] · A/A unresolved | 75.252 → 70.679 · 0.942 [0.930, 0.955] · Δ -5.80% · speed +6.16% [+4.76%, +7.52%] · A/A unresolved |
| original22-v1/threads/private-memory | 21.340 → 21.487 · 1.010 [0.988, 1.031] · Δ +0.96% · speed -0.95% [-2.98%, +1.18%] · A/A unresolved | 19.288 → 19.160 · 0.997 [0.989, 1.005] · Δ -0.33% · speed +0.34% [-0.52%, +1.10%] · A/A resolved | 19.000 → 19.304 · 1.013 [1.001, 1.024] · Δ +1.30% · speed -1.28% [-2.31%, -0.15%] · A/A resolved | 19.160 → 19.304 · 1.004 [0.996, 1.011] · Δ +0.40% · speed -0.40% [-1.05%, +0.41%] · A/A resolved | 21.487 → 19.304 · 0.894 [0.884, 0.903] · Δ -10.63% · speed +11.90% [+10.71%, +13.10%] · A/A resolved |
| original22-v1/threads/shared-memory | 21.602 → 21.646 · 1.008 [0.997, 1.020] · Δ +0.84% · speed -0.84% [-2.00%, +0.33%] · A/A unresolved | 18.960 → 19.131 · 1.009 [0.994, 1.026] · Δ +0.87% · speed -0.87% [-2.53%, +0.58%] · A/A resolved | 19.021 → 18.969 · 0.997 [0.988, 1.005] · Δ -0.31% · speed +0.31% [-0.53%, +1.20%] · A/A resolved | 19.131 → 18.969 · 0.995 [0.984, 1.005] · Δ -0.49% · speed +0.49% [-0.55%, +1.63%] · A/A resolved | 21.646 → 18.969 · 0.875 [0.860, 0.889] · Δ -12.52% · speed +14.31% [+12.45%, +16.26%] · A/A unresolved |
| original22-v1/threads/atomic-rmw | 27.617 → 27.856 · 1.009 [1.002, 1.019] · Δ +0.91% · speed -0.91% [-1.89%, -0.20%] · A/A unresolved | 28.163 → 28.102 · 1.003 [0.990, 1.016] · Δ +0.27% · speed -0.27% [-1.53%, +1.05%] · A/A unresolved | 27.798 → 27.615 · 0.999 [0.982, 1.016] · Δ -0.09% · speed +0.09% [-1.60%, +1.83%] · A/A unresolved | 28.102 → 27.615 · 0.984 [0.965, 1.003] · Δ -1.61% · speed +1.64% [-0.29%, +3.63%] · A/A unresolved | 27.856 → 27.615 · 0.990 [0.973, 1.008] · Δ -1.00% · speed +1.01% [-0.84%, +2.74%] · A/A resolved |
| original22-v1/threads/notify-empty | 24.473 → 24.550 · 1.001 [0.986, 1.017] · Δ +0.07% · speed -0.07% [-1.63%, +1.46%] · A/A unresolved | 23.256 → 23.473 · 0.996 [0.973, 1.017] · Δ -0.41% · speed +0.41% [-1.63%, +2.81%] · A/A unresolved | 22.642 → 22.565 · 1.008 [0.995, 1.018] · Δ +0.83% · speed -0.82% [-1.76%, +0.47%] · A/A unresolved | 23.473 → 22.565 · 0.978 [0.953, 1.007] · Δ -2.19% · speed +2.24% [-0.70%, +4.89%] · A/A unresolved | 24.550 → 22.565 · 0.927 [0.918, 0.937] · Δ -7.33% · speed +7.91% [+6.72%, +8.88%] · A/A unresolved |
| original22-v1/threads/wait-mismatch | 36.898 → 36.477 · 0.997 [0.980, 1.016] · Δ -0.32% · speed +0.32% [-1.54%, +2.08%] · A/A unresolved | 33.521 → 33.290 · 1.005 [0.984, 1.032] · Δ +0.53% · speed -0.53% [-3.08%, +1.67%] · A/A unresolved | 33.394 → 33.342 · 1.009 [0.993, 1.026] · Δ +0.91% · speed -0.91% [-2.58%, +0.72%] · A/A unresolved | 33.290 → 33.342 · 1.002 [0.972, 1.024] · Δ +0.18% · speed -0.18% [-2.34%, +2.85%] · A/A unresolved | 36.477 → 33.342 · 0.914 [0.901, 0.933] · Δ -8.59% · speed +9.40% [+7.15%, +10.95%] · A/A unresolved |
| original22-v1/threads/wait-timeout | 36.219 → 35.800 · 0.993 [0.977, 1.012] · Δ -0.71% · speed +0.72% [-1.17%, +2.35%] · A/A unresolved | 31.837 → 31.604 · 1.002 [0.989, 1.015] · Δ +0.19% · speed -0.19% [-1.52%, +1.09%] · A/A unresolved | 32.246 → 31.881 · 0.996 [0.982, 1.011] · Δ -0.38% · speed +0.38% [-1.11%, +1.85%] · A/A unresolved | 31.604 → 31.881 · 1.010 [0.995, 1.026] · Δ +0.98% · speed -0.97% [-2.53%, +0.50%] · A/A unresolved | 35.800 → 31.881 · 0.899 [0.890, 0.908] · Δ -10.12% · speed +11.26% [+10.08%, +12.36%] · A/A unresolved |
| original22-v1/applications/sha256-clang | 18,735.10 → 18,577.29 · 0.992 [0.987, 0.996] · Δ -0.83% · speed +0.84% [+0.43%, +1.30%] · A/A resolved | 15,742.44 → 15,801.22 · 1.005 [0.997, 1.014] · Δ +0.55% · speed -0.54% [-1.37%, +0.27%] · A/A resolved | 15,883.25 → 15,871.92 · 0.996 [0.987, 1.003] · Δ -0.42% · speed +0.43% [-0.30%, +1.34%] · A/A resolved | 15,801.22 → 15,871.92 · 1.002 [0.996, 1.008] · Δ +0.18% · speed -0.18% [-0.75%, +0.43%] · A/A resolved | 18,577.29 → 15,871.92 · 0.853 [0.849, 0.857] · Δ -14.68% · speed +17.20% [+16.75%, +17.76%] · A/A resolved |
| original22-v1/applications/sha256-rust | 18,252.32 → 18,089.42 · 0.990 [0.987, 0.993] · Δ -1.04% · speed +1.05% [+0.75%, +1.34%] · A/A resolved | 15,358.96 → 15,392.56 · 1.001 [0.996, 1.006] · Δ +0.06% · speed -0.06% [-0.57%, +0.39%] · A/A resolved | 15,460.06 → 15,451.41 · 0.999 [0.991, 1.006] · Δ -0.13% · speed +0.13% [-0.56%, +0.87%] · A/A unresolved | 15,392.56 → 15,451.41 · 1.002 [0.997, 1.006] · Δ +0.17% · speed -0.17% [-0.57%, +0.29%] · A/A unresolved | 18,089.42 → 15,451.41 · 0.852 [0.847, 0.857] · Δ -14.76% · speed +17.31% [+16.71%, +18.02%] · A/A unresolved |
| gc-sustained-v1/gc/i31-roundtrip | 32.429 → 32.880 · 1.012 [1.003, 1.021] · Δ +1.21% · speed -1.20% [-2.08%, -0.28%] · A/A resolved | 28.836 → 29.043 · 1.005 [0.995, 1.014] · Δ +0.46% · speed -0.46% [-1.35%, +0.48%] · A/A resolved | 29.158 → 29.044 · 0.995 [0.986, 1.005] · Δ -0.54% · speed +0.54% [-0.46%, +1.41%] · A/A unresolved | 29.043 → 29.044 · 1.001 [0.996, 1.007] · Δ +0.11% · speed -0.11% [-0.74%, +0.45%] · A/A unresolved | 32.880 → 29.044 · 0.884 [0.876, 0.893] · Δ -11.55% · speed +13.06% [+11.99%, +14.13%] · A/A unresolved |
| gc-sustained-v1/gc/struct-access | 38.469 → 38.368 · 0.999 [0.996, 1.002] · Δ -0.12% · speed +0.12% [-0.24%, +0.39%] · A/A resolved | 36.115 → 36.400 · 1.007 [0.998, 1.013] · Δ +0.65% · speed -0.65% [-1.30%, +0.21%] · A/A resolved | 36.468 → 36.337 · 0.996 [0.990, 1.003] · Δ -0.38% · speed +0.38% [-0.33%, +1.04%] · A/A resolved | 36.400 → 36.337 · 0.997 [0.993, 1.001] · Δ -0.26% · speed +0.26% [-0.14%, +0.66%] · A/A resolved | 38.368 → 36.337 · 0.947 [0.942, 0.953] · Δ -5.28% · speed +5.57% [+4.88%, +6.16%] · A/A resolved |
| gc-sustained-v1/gc/struct-allocation | 96.642 → 92.379 · 0.976 [0.954, 1.006] · Δ -2.39% · speed +2.44% [-0.62%, +4.79%] · A/A unresolved | 103.59 → 102.48 · 0.991 [0.986, 0.995] · Δ -0.93% · speed +0.94% [+0.46%, +1.42%] · A/A resolved | 103.01 → 103.19 · 1.003 [0.993, 1.013] · Δ +0.26% · speed -0.26% [-1.30%, +0.75%] · A/A resolved | 102.48 → 103.19 · 1.007 [0.998, 1.017] · Δ +0.75% · speed -0.74% [-1.69%, +0.20%] · A/A resolved | 92.379 → 103.19 · 1.094 [1.061, 1.117] · Δ +9.38% · speed -8.57% [-10.45%, -5.76%] · A/A unresolved |
| gc-sustained-v1/gc/struct-indirect-access | 72.565 → 72.250 · 0.993 [0.988, 0.997] · Δ -0.73% · speed +0.74% [+0.26%, +1.19%] · A/A resolved | 68.443 → 68.071 · 0.996 [0.987, 1.004] · Δ -0.43% · speed +0.43% [-0.42%, +1.33%] · A/A resolved | 68.549 → 68.497 · 1.002 [0.990, 1.019] · Δ +0.18% · speed -0.18% [-1.88%, +1.05%] · A/A resolved | 68.071 → 68.497 · 1.004 [0.990, 1.019] · Δ +0.43% · speed -0.43% [-1.90%, +0.99%] · A/A resolved | 72.250 → 68.497 · 0.951 [0.938, 0.967] · Δ -4.90% · speed +5.15% [+3.41%, +6.56%] · A/A resolved |
| threads-sustained-v2/threads/private-memory | 20.214 → 20.531 · 1.013 [1.005, 1.021] · Δ +1.33% · speed -1.31% [-2.07%, -0.53%] · A/A resolved | 18.527 → 18.485 · 0.997 [0.990, 1.006] · Δ -0.28% · speed +0.28% [-0.55%, +1.06%] · A/A resolved | 18.465 → 18.424 · 0.993 [0.984, 1.001] · Δ -0.72% · speed +0.72% [-0.09%, +1.66%] · A/A unresolved | 18.485 → 18.424 · 0.998 [0.993, 1.003] · Δ -0.17% · speed +0.17% [-0.34%, +0.74%] · A/A resolved | 20.531 → 18.424 · 0.896 [0.889, 0.902] · Δ -10.38% · speed +11.58% [+10.87%, +12.49%] · A/A resolved |
| threads-sustained-v2/threads/shared-memory | 20.387 → 20.463 · 0.992 [0.968, 1.008] · Δ -0.82% · speed +0.82% [-0.79%, +3.33%] · A/A unresolved | 18.641 → 18.460 · 0.986 [0.973, 0.998] · Δ -1.45% · speed +1.47% [+0.16%, +2.80%] · A/A unresolved | 18.281 → 18.362 · 1.008 [0.989, 1.040] · Δ +0.79% · speed -0.79% [-3.84%, +1.12%] · A/A unresolved | 18.460 → 18.362 · 1.000 [0.988, 1.016] · Δ -0.04% · speed +0.04% [-1.61%, +1.21%] · A/A unresolved | 20.463 → 18.362 · 0.908 [0.892, 0.932] · Δ -9.20% · speed +10.13% [+7.33%, +12.05%] · A/A unresolved |
| threads-sustained-v2/threads/atomic-rmw | 26.786 → 26.698 · 1.005 [0.999, 1.012] · Δ +0.55% · speed -0.54% [-1.19%, +0.09%] · A/A unresolved | 27.463 → 27.116 · 0.994 [0.983, 1.006] · Δ -0.57% · speed +0.57% [-0.59%, +1.69%] · A/A unresolved | 26.718 → 26.642 · 0.995 [0.990, 1.000] · Δ -0.48% · speed +0.48% [+0.00%, +1.00%] · A/A unresolved | 27.116 → 26.642 · 0.980 [0.972, 0.987] · Δ -2.04% · speed +2.09% [+1.29%, +2.93%] · A/A unresolved | 26.698 → 26.642 · 0.996 [0.988, 1.003] · Δ -0.40% · speed +0.40% [-0.34%, +1.22%] · A/A unresolved |
| threads-sustained-v2/threads/notify-empty | 23.125 → 23.437 · 1.010 [1.004, 1.016] · Δ +1.03% · speed -1.02% [-1.55%, -0.42%] · A/A resolved | 22.555 → 22.282 · 0.996 [0.981, 1.011] · Δ -0.45% · speed +0.45% [-1.05%, +1.97%] · A/A unresolved | 22.006 → 21.778 · 0.994 [0.990, 0.999] · Δ -0.59% · speed +0.60% [+0.12%, +1.01%] · A/A unresolved | 22.282 → 21.778 · 0.974 [0.961, 0.987] · Δ -2.61% · speed +2.68% [+1.35%, +4.09%] · A/A unresolved | 23.437 → 21.778 · 0.934 [0.926, 0.942] · Δ -6.63% · speed +7.10% [+6.16%, +7.96%] · A/A unresolved |
| threads-sustained-v2/threads/wait-mismatch | 36.065 → 36.205 · 1.009 [1.001, 1.019] · Δ +0.88% · speed -0.88% [-1.82%, -0.10%] · A/A unresolved | 33.196 → 33.036 · 1.001 [0.988, 1.015] · Δ +0.09% · speed -0.09% [-1.43%, +1.25%] · A/A unresolved | 33.383 → 32.699 · 0.987 [0.972, 1.004] · Δ -1.30% · speed +1.32% [-0.38%, +2.85%] · A/A resolved | 33.036 → 32.699 · 0.988 [0.972, 1.001] · Δ -1.24% · speed +1.25% [-0.08%, +2.87%] · A/A unresolved | 36.205 → 32.699 · 0.903 [0.894, 0.910] · Δ -9.73% · speed +10.78% [+9.86%, +11.89%] · A/A unresolved |
| threads-sustained-v2/threads/wait-timeout | 34.232 → 33.666 · 0.989 [0.981, 1.001] · Δ -1.07% · speed +1.08% [-0.08%, +1.95%] · A/A unresolved | 31.208 → 30.969 · 0.990 [0.975, 1.006] · Δ -0.99% · speed +1.00% [-0.62%, +2.60%] · A/A unresolved | 31.033 → 31.756 · 1.018 [1.008, 1.025] · Δ +1.78% · speed -1.75% [-2.46%, -0.76%] · A/A unresolved | 30.969 → 31.756 · 1.023 [1.012, 1.032] · Δ +2.29% · speed -2.24% [-3.12%, -1.23%] · A/A unresolved | 33.666 → 31.756 · 0.935 [0.922, 0.948] · Δ -6.51% · speed +6.96% [+5.47%, +8.43%] · A/A unresolved |
| generic-calls-v1/core/generic-calls | 96.372 → 96.679 · 1.003 [0.998, 1.008] · Δ +0.30% · speed -0.30% [-0.84%, +0.23%] · A/A resolved | 85.521 → 85.717 · 1.002 [0.998, 1.006] · Δ +0.23% · speed -0.23% [-0.59%, +0.25%] · A/A resolved | 87.120 → 86.930 · 0.998 [0.995, 1.001] · Δ -0.20% · speed +0.20% [-0.07%, +0.48%] · A/A resolved | 85.717 → 86.930 · 1.014 [1.011, 1.017] · Δ +1.40% · speed -1.38% [-1.62%, -1.10%] · A/A resolved | 96.679 → 86.930 · 0.900 [0.896, 0.903] · Δ -10.05% · speed +11.17% [+10.72%, +11.58%] · A/A resolved |

### Preview 3 source comparisons

| Payload bytes | Compiler | CPU time | Instructions per request change | Requests per second change |
| --- | ---: | ---: | ---: | ---: |
| 16 | go | 61.463 → 62.212 · 1.005 [0.997, 1.014] · Δ +0.54% · speed -0.54% [-1.36%, +0.29%] · A/A resolved | +0.14% · resolved | -0.28% · resolved |
| 16 | llgo | 127.94 → 128.27 · 1.004 [0.996, 1.012] · Δ +0.42% · speed -0.42% [-1.21%, +0.35%] · A/A resolved | +0.07% · resolved | -1.51% · unresolved |
| 16 | llgo-main | 128.58 → 127.34 · 0.997 [0.991, 1.004] · Δ -0.29% · speed +0.29% [-0.39%, +0.90%] · A/A unresolved | -0.20% · resolved | +0.34% · unresolved |
| 65536 | go | 87.308 → 87.554 · 1.006 [0.998, 1.014] · Δ +0.58% · speed -0.58% [-1.39%, +0.16%] · A/A resolved | +0.16% · resolved | -0.37% · unresolved |
| 65536 | llgo | 156.53 → 158.05 · 1.004 [0.987, 1.018] · Δ +0.44% · speed -0.44% [-1.80%, +1.31%] · A/A unresolved | +0.41% · unresolved | +0.12% · unresolved |
| 65536 | llgo-main | 154.37 → 156.37 · 1.008 [0.993, 1.027] · Δ +0.81% · speed -0.81% [-2.64%, +0.73%] · A/A unresolved | +1.23% · unresolved | +0.75% · unresolved |

### Preview 3 refreshed LLGo vs same-source Go

| Payload bytes | Metric | wasmvm Go | Refreshed wasmvm LLGo | Paired LLGo vs Go |
| --- | ---: | ---: | ---: | ---: |
| 16 | CPU µs/request | 62.212 | 127.34 | 2.048 [2.014, 2.085] · Δ +104.82% · speed -51.18% [-52.04%, -50.36%] · A/A unresolved |
| 16 | requests/s | 61,079.78 | 34,166.12 | 0.561 [0.545, 0.576] · Δ -43.91% · speed -43.91% [-45.49%, -42.42%] · A/A unresolved |
| 65536 | CPU µs/request | 87.554 | 156.37 | 1.781 [1.743, 1.812] · Δ +78.15% · speed -43.87% [-44.81%, -42.64%] · A/A resolved |
| 65536 | requests/s | 42,740.65 | 28,134.91 | 0.666 [0.659, 0.673] · Δ -33.40% · speed -33.40% [-34.10%, -32.66%] · A/A resolved |

## Preparation and first execution

Phase values are medians in µs. Each process runs its own preparation path.
Signed descriptive speed deltas use same-source wasmvm Go for each phase.
NC indicates a different measured phase contract. WasmEdge AOT preparation
includes artifact I/O. Direct phase observations are retained in core-analysis.json
and p3-analysis.json, with full records under core and p3-*.

<details>
<summary>Native lifecycle phase observations</summary>

| Workload | Phase | wasmvm Go | wasmvm LLGo (pinned) | wasmvm LLGo 42efacae | wazero 1.12.0 | Wasmtime 48.0.1 | Wasmer 7.3.0 | WasmEdge 0.17.1 | WasmEdge 0.18.1 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | processStartup | 4,010.79 (ref) | 5,877.23 (-31.76%) | 6,004.73 (-33.21%) | 4,110.48 (-2.43%) | 4,151.96 (-3.40%) | 5,859.35 (-31.55%) | 15,714.44 (-74.48%) | 11,768.50 (-65.92%) |
| original22-v1/core/call-control | load | 37.416 (ref) | 38.250 (-2.18%) | 44.166 (-15.28%) | 36.834 (+1.58%) | 31.916 (+17.23%) | 39.604 (-5.52%) | 48.980 (-23.61%) | 44.854 (-16.58%) |
| original22-v1/core/call-control | engineSetup | 5.375 (ref) | 4.354 (+23.44%) | 6.250 (-14.00%) | 54.229 (-90.09%) | 122.19 (-95.60%) | 38.959 (-86.20%) | 24.646 (-78.19%) | 12.834 (-58.12%) |
| original22-v1/core/call-control | compileValidate | 180.15 (ref) | 140.79 (+27.95%) | 137.15 (+31.35%) | 100.92 (+78.51%) | 712.10 (-74.70%) | 1,069.38 (-83.15%) | 271,964.42 (NC) | 291,113.35 (NC) |
| original22-v1/core/call-control | instantiate | 8.062 (ref) | 9.209 (-12.45%) | 10.250 (-21.34%) | 14.125 (-42.92%) | 6.896 (+16.92%) | 16.666 (-51.62%) | 15.896 (-49.28%) | 21.667 (-62.79%) |
| original22-v1/core/call-control | exportLookup | 0.250 (ref) | 0.334 (-25.04%) | 0.438 (-42.86%) | 2.083 (-88.00%) | 3.521 (-92.90%) | 4.500 (-94.44%) | 0.230 (+8.93%) | 0.312 (-20.00%) |
| original22-v1/core/call-control | firstCall | 10.229 (ref) | 8.125 (+25.90%) | 7.750 (+31.99%) | 9.124 (+12.10%) | 27.042 (-62.17%) | 13.062 (-21.69%) | 13.459 (-24.00%) | 32.813 (-68.83%) |
| original22-v1/core/call-control | cleanup | 8.854 (ref) | 4.354 (+103.36%) | 7.812 (+13.34%) | 17.458 (-49.28%) | 69.895 (-87.33%) | 24.458 (-63.80%) | 364.77 (-97.57%) | 287.23 (-96.92%) |
| original22-v1/core/integer-loop | processStartup | 4,256.46 (ref) | 6,005.40 (-29.12%) | 5,918.48 (-28.08%) | 4,265.88 (-0.22%) | 4,285.60 (-0.68%) | 6,134.58 (-30.62%) | 16,223.62 (-73.76%) | 11,855.38 (-64.10%) |
| original22-v1/core/integer-loop | load | 42.312 (ref) | 42.479 (-0.39%) | 41.000 (+3.20%) | 39.520 (+7.06%) | 44.979 (-5.93%) | 40.375 (+4.80%) | 44.333 (-4.56%) | 43.938 (-3.70%) |
| original22-v1/core/integer-loop | engineSetup | 5.750 (ref) | 4.396 (+30.82%) | 6.812 (-15.59%) | 54.667 (-89.48%) | 128.44 (-95.52%) | 40.188 (-85.69%) | 26.895 (-78.62%) | 12.396 (-53.61%) |
| original22-v1/core/integer-loop | compileValidate | 239.88 (ref) | 184.60 (+29.94%) | 181.02 (+32.51%) | 117.67 (+103.86%) | 743.21 (-67.72%) | 1,074.50 (-77.68%) | 286,580.37 (NC) | 278,102.48 (NC) |
| original22-v1/core/integer-loop | instantiate | 7.625 (ref) | 9.396 (-18.85%) | 9.751 (-21.80%) | 13.562 (-43.78%) | 7.625 (-0.01%) | 16.959 (-55.04%) | 15.166 (-49.73%) | 22.209 (-65.67%) |
| original22-v1/core/integer-loop | exportLookup | 0.230 (ref) | 0.438 (-47.54%) | 0.396 (-42.05%) | 2.125 (-89.20%) | 3.854 (-94.05%) | 4.542 (-94.95%) | 0.230 (+0.00%) | 0.251 (-8.38%) |
| original22-v1/core/integer-loop | firstCall | 1,570.15 (ref) | 1,570.77 (-0.04%) | 1,593.48 (-1.46%) | 1,557.90 (+0.79%) | 1,603.08 (-2.05%) | 1,540.25 (+1.94%) | 1,586.27 (-1.02%) | 1,613.40 (-2.68%) |
| original22-v1/core/integer-loop | cleanup | 6.292 (ref) | 4.375 (+43.81%) | 4.583 (+37.28%) | 4.708 (+33.63%) | 57.687 (-89.09%) | 26.500 (-76.26%) | 397.98 (-98.42%) | 314.46 (-98.00%) |
| original22-v1/core/function-calls | processStartup | 4,249.75 (ref) | 6,258.44 (-32.10%) | 6,048.44 (-29.74%) | 4,299.50 (-1.16%) | 4,259.29 (-0.22%) | 6,179.04 (-31.22%) | 15,814.38 (-73.13%) | 11,936.98 (-64.40%) |
| original22-v1/core/function-calls | load | 42.500 (ref) | 41.875 (+1.49%) | 43.833 (-3.04%) | 40.646 (+4.56%) | 36.542 (+16.30%) | 42.771 (-0.63%) | 41.208 (+3.14%) | 51.312 (-17.17%) |
| original22-v1/core/function-calls | engineSetup | 5.333 (ref) | 4.750 (+12.28%) | 8.021 (-33.50%) | 54.874 (-90.28%) | 125.10 (-95.74%) | 41.834 (-87.25%) | 24.459 (-78.19%) | 13.896 (-61.62%) |
| original22-v1/core/function-calls | compileValidate | 291.38 (ref) | 259.12 (+12.45%) | 240.79 (+21.01%) | 127.25 (+128.98%) | 775.15 (-62.41%) | 1,180.25 (-75.31%) | 279,369.94 (NC) | 300,449.92 (NC) |
| original22-v1/core/function-calls | instantiate | 9.979 (ref) | 10.271 (-2.84%) | 10.584 (-5.71%) | 13.541 (-26.30%) | 7.042 (+41.72%) | 17.875 (-44.17%) | 13.854 (-27.97%) | 21.146 (-52.81%) |
| original22-v1/core/function-calls | exportLookup | 0.250 (ref) | 0.334 (-25.15%) | 0.416 (-39.90%) | 1.855 (-86.52%) | 3.771 (-93.37%) | 4.438 (-94.37%) | 0.229 (+9.17%) | 0.250 (+0.00%) |
| original22-v1/core/function-calls | firstCall | 812.04 (ref) | 786.48 (+3.25%) | 799.40 (+1.58%) | 1,314.04 (-38.20%) | 929.40 (-12.63%) | 786.83 (+3.20%) | 797.10 (+1.87%) | 818.40 (-0.78%) |
| original22-v1/core/function-calls | cleanup | 5.729 (ref) | 4.229 (+35.50%) | 5.021 (+14.12%) | 6.667 (-14.06%) | 70.791 (-91.91%) | 30.708 (-81.34%) | 381.02 (-98.50%) | 318.00 (-98.20%) |
| original22-v1/core/memory-load-store | processStartup | 4,219.85 (ref) | 5,964.08 (-29.25%) | 6,138.17 (-31.25%) | 4,453.58 (-5.25%) | 4,291.94 (-1.68%) | 5,931.73 (-28.86%) | 16,431.60 (-74.32%) | 11,991.08 (-64.81%) |
| original22-v1/core/memory-load-store | load | 45.874 (ref) | 40.500 (+13.27%) | 49.166 (-6.70%) | 46.208 (-0.72%) | 38.854 (+18.07%) | 41.041 (+11.78%) | 45.750 (+0.27%) | 49.208 (-6.77%) |
| original22-v1/core/memory-load-store | engineSetup | 5.646 (ref) | 4.563 (+23.73%) | 7.604 (-25.75%) | 56.978 (-90.09%) | 125.08 (-95.49%) | 43.708 (-87.08%) | 21.833 (-74.14%) | 13.562 (-58.37%) |
| original22-v1/core/memory-load-store | compileValidate | 316.27 (ref) | 293.10 (+7.90%) | 267.46 (+18.25%) | 134.17 (+135.73%) | 687.40 (-53.99%) | 1,148.71 (-72.47%) | 285,972.35 (NC) | 290,550.38 (NC) |
| original22-v1/core/memory-load-store | instantiate | 12.458 (ref) | 13.854 (-10.08%) | 14.188 (-12.19%) | 15.584 (-20.06%) | 20.021 (-37.78%) | 30.271 (-58.85%) | 32.520 (-61.69%) | 42.749 (-70.86%) |
| original22-v1/core/memory-load-store | exportLookup | 0.250 (ref) | 0.438 (-42.86%) | 0.416 (-39.98%) | 2.000 (-87.50%) | 3.749 (-93.33%) | 4.250 (-94.12%) | 0.208 (+20.19%) | 0.292 (-14.38%) |
| original22-v1/core/memory-load-store | firstCall | 444.35 (ref) | 442.83 (+0.34%) | 439.81 (+1.03%) | 828.15 (-46.34%) | 503.92 (-11.82%) | 522.50 (-14.96%) | 294.23 (+51.02%) | 313.02 (+41.96%) |
| original22-v1/core/memory-load-store | cleanup | 7.063 (ref) | 6.604 (+6.94%) | 6.854 (+3.04%) | 4.250 (+66.19%) | 82.291 (-91.42%) | 51.312 (-86.24%) | 451.90 (-98.44%) | 343.50 (-97.94%) |
| original22-v1/memory/narrow-unaligned | processStartup | 4,098.06 (ref) | 6,019.92 (-31.92%) | 5,810.35 (-29.47%) | 4,301.81 (-4.74%) | 4,369.38 (-6.21%) | 6,087.90 (-32.69%) | 16,067.98 (-74.50%) | 12,047.92 (-65.99%) |
| original22-v1/memory/narrow-unaligned | load | 42.709 (ref) | 46.167 (-7.49%) | 48.021 (-11.06%) | 34.312 (+24.47%) | 43.208 (-1.16%) | 39.417 (+8.35%) | 50.021 (-14.62%) | 49.354 (-13.46%) |
| original22-v1/memory/narrow-unaligned | engineSetup | 5.625 (ref) | 4.688 (+20.00%) | 6.979 (-19.40%) | 53.979 (-89.58%) | 126.40 (-95.55%) | 40.834 (-86.22%) | 29.625 (-81.01%) | 12.250 (-54.08%) |
| original22-v1/memory/narrow-unaligned | compileValidate | 297.81 (ref) | 283.29 (+5.13%) | 270.06 (+10.28%) | 143.60 (+107.38%) | 768.29 (-61.24%) | 1,166.56 (-74.47%) | 283,904.85 (NC) | 278,589.98 (NC) |
| original22-v1/memory/narrow-unaligned | instantiate | 12.250 (ref) | 13.959 (-12.24%) | 15.041 (-18.56%) | 19.334 (-36.64%) | 20.291 (-39.63%) | 30.729 (-60.14%) | 33.105 (-63.00%) | 42.688 (-71.30%) |
| original22-v1/memory/narrow-unaligned | exportLookup | 0.250 (ref) | 0.438 (-42.86%) | 0.458 (-45.41%) | 2.146 (-88.35%) | 3.854 (-93.51%) | 4.396 (-94.31%) | 0.208 (+20.19%) | 0.229 (+9.17%) |
| original22-v1/memory/narrow-unaligned | firstCall | 655.23 (ref) | 642.67 (+1.95%) | 659.46 (-0.64%) | 976.92 (-32.93%) | 730.31 (-10.28%) | 728.52 (-10.06%) | 527.00 (+24.33%) | 552.44 (+18.61%) |
| original22-v1/memory/narrow-unaligned | cleanup | 9.896 (ref) | 8.292 (+19.34%) | 8.646 (+14.45%) | 5.750 (+72.10%) | 81.063 (-87.79%) | 64.876 (-84.75%) | 447.60 (-97.79%) | 366.23 (-97.30%) |
| original22-v1/memory/dependent-address | processStartup | 4,138.98 (ref) | 6,079.42 (-31.92%) | 5,937.69 (-30.29%) | 4,256.98 (-2.77%) | 4,238.90 (-2.36%) | 5,872.71 (-29.52%) | 16,316.42 (-74.63%) | 11,676.81 (-64.55%) |
| original22-v1/memory/dependent-address | load | 43.000 (ref) | 43.625 (-1.43%) | 49.416 (-12.98%) | 44.105 (-2.50%) | 37.437 (+14.86%) | 47.791 (-10.03%) | 45.291 (-5.06%) | 43.459 (-1.06%) |
| original22-v1/memory/dependent-address | engineSetup | 5.563 (ref) | 4.479 (+24.19%) | 6.708 (-17.08%) | 56.417 (-90.14%) | 121.67 (-95.43%) | 38.896 (-85.70%) | 27.480 (-79.76%) | 12.500 (-55.50%) |
| original22-v1/memory/dependent-address | compileValidate | 285.04 (ref) | 267.40 (+6.60%) | 246.12 (+15.81%) | 150.33 (+89.61%) | 760.56 (-62.52%) | 1,156.17 (-75.35%) | 280,627.69 (NC) | 283,117.56 (NC) |
| original22-v1/memory/dependent-address | instantiate | 12.479 (ref) | 13.563 (-7.99%) | 14.916 (-16.34%) | 18.417 (-32.24%) | 20.084 (-37.87%) | 30.834 (-59.53%) | 34.188 (-63.50%) | 40.230 (-68.98%) |
| original22-v1/memory/dependent-address | exportLookup | 0.250 (ref) | 0.375 (-33.33%) | 0.458 (-45.41%) | 1.855 (-86.52%) | 3.646 (-93.14%) | 4.604 (-94.57%) | 0.209 (+19.62%) | 0.250 (+0.00%) |
| original22-v1/memory/dependent-address | firstCall | 1,861.56 (ref) | 1,833.33 (+1.54%) | 1,857.92 (+0.20%) | 2,988.42 (-37.71%) | 2,110.56 (-11.80%) | 2,102.52 (-11.46%) | 1,457.02 (+27.77%) | 1,471.73 (+26.49%) |
| original22-v1/memory/dependent-address | cleanup | 10.917 (ref) | 16.104 (-32.21%) | 15.709 (-30.50%) | 10.229 (+6.72%) | 100.08 (-89.09%) | 53.416 (-79.56%) | 453.71 (-97.59%) | 357.15 (-96.94%) |
| original22-v1/memory/float64-unaligned | processStartup | 4,025.88 (ref) | 5,971.06 (-32.58%) | 6,048.44 (-33.44%) | 4,302.71 (-6.43%) | 4,139.77 (-2.75%) | 5,877.81 (-31.51%) | 16,143.08 (-75.06%) | 11,968.56 (-66.36%) |
| original22-v1/memory/float64-unaligned | load | 47.605 (ref) | 46.728 (+1.87%) | 48.104 (-1.04%) | 43.501 (+9.43%) | 35.729 (+33.24%) | 43.646 (+9.07%) | 46.875 (+1.56%) | 49.333 (-3.50%) |
| original22-v1/memory/float64-unaligned | engineSetup | 5.500 (ref) | 4.562 (+20.56%) | 7.854 (-29.97%) | 55.417 (-90.08%) | 120.83 (-95.45%) | 40.688 (-86.48%) | 25.729 (-78.62%) | 13.854 (-60.30%) |
| original22-v1/memory/float64-unaligned | compileValidate | 308.90 (ref) | 289.88 (+6.56%) | 269.19 (+14.75%) | 147.44 (+109.51%) | 778.10 (-60.30%) | 1,159.77 (-73.37%) | 292,899.77 (NC) | 289,331.71 (NC) |
| original22-v1/memory/float64-unaligned | instantiate | 13.354 (ref) | 14.459 (-7.64%) | 14.250 (-6.28%) | 18.084 (-26.15%) | 19.875 (-32.81%) | 30.083 (-55.61%) | 34.938 (-61.78%) | 42.333 (-68.45%) |
| original22-v1/memory/float64-unaligned | exportLookup | 0.229 (ref) | 0.458 (-50.00%) | 0.458 (-50.00%) | 2.229 (-89.73%) | 3.853 (-94.06%) | 4.458 (-94.86%) | 0.208 (+10.10%) | 0.250 (-8.40%) |
| original22-v1/memory/float64-unaligned | firstCall | 734.48 (ref) | 741.15 (-0.90%) | 744.94 (-1.40%) | 790.77 (-7.12%) | 736.02 (-0.21%) | 721.73 (+1.77%) | 691.98 (+6.14%) | 706.38 (+3.98%) |
| original22-v1/memory/float64-unaligned | cleanup | 8.167 (ref) | 7.354 (+11.06%) | 6.458 (+26.45%) | 3.896 (+109.63%) | 73.584 (-88.90%) | 55.062 (-85.17%) | 445.50 (-98.17%) | 384.92 (-97.88%) |
| original22-v1/runtime/simd-memory-full | processStartup | 4,207.23 (ref) | 5,967.90 (-29.50%) | 6,059.83 (-30.57%) | 4,318.69 (-2.58%) | 4,211.38 (-0.10%) | 6,038.42 (-30.33%) | 16,215.75 (-74.05%) | 12,210.31 (-65.54%) |
| original22-v1/runtime/simd-memory-full | load | 38.312 (ref) | 40.188 (-4.67%) | 43.688 (-12.30%) | 42.438 (-9.72%) | 39.312 (-2.54%) | 37.833 (+1.27%) | 54.645 (-29.89%) | 48.521 (-21.04%) |
| original22-v1/runtime/simd-memory-full | engineSetup | 5.583 (ref) | 4.604 (+21.26%) | 6.896 (-19.03%) | 54.188 (-89.70%) | 122.65 (-95.45%) | 40.438 (-86.19%) | 25.730 (-78.30%) | 13.541 (-58.77%) |
| original22-v1/runtime/simd-memory-full | compileValidate | 276.75 (ref) | 250.48 (+10.49%) | 236.90 (+16.82%) | 126.98 (+117.95%) | 942.23 (-70.63%) | 1,164.67 (-76.24%) | 276,333.56 (NC) | 283,190.73 (NC) |
| original22-v1/runtime/simd-memory-full | instantiate | 13.125 (ref) | 14.500 (-9.48%) | 16.959 (-22.61%) | 18.687 (-29.76%) | 55.708 (-76.44%) | 34.062 (-61.47%) | 40.500 (-67.59%) | 52.333 (-74.92%) |
| original22-v1/runtime/simd-memory-full | exportLookup | 0.250 (ref) | 0.416 (-39.98%) | 0.396 (-36.87%) | 1.833 (-86.36%) | 1.938 (-87.10%) | 4.438 (-94.37%) | 0.230 (+8.93%) | 0.230 (+8.93%) |
| original22-v1/runtime/simd-memory-full | firstCall | 3,439.04 (ref) | 3,380.73 (+1.72%) | 3,405.35 (+0.99%) | 3,630.23 (-5.27%) | 3,675.35 (-6.43%) | 3,473.19 (-0.98%) | 3,623.27 (-5.08%) | 3,622.40 (-5.06%) |
| original22-v1/runtime/simd-memory-full | cleanup | 9.854 (ref) | 8.459 (+16.50%) | 12.021 (-18.03%) | 5.125 (+92.27%) | 94.271 (-89.55%) | 53.228 (-81.49%) | 446.17 (-97.79%) | 362.94 (-97.28%) |
| original22-v1/runtime/simd-memory-widen | processStartup | 4,306.67 (ref) | 6,030.56 (-28.59%) | 5,938.69 (-27.48%) | 4,269.42 (+0.87%) | 4,265.17 (+0.97%) | 6,015.73 (-28.41%) | 16,188.69 (-73.40%) | 12,071.27 (-64.32%) |
| original22-v1/runtime/simd-memory-widen | load | 38.230 (ref) | 42.104 (-9.20%) | 46.230 (-17.30%) | 47.416 (-19.38%) | 40.854 (-6.42%) | 43.771 (-12.66%) | 43.459 (-12.03%) | 42.604 (-10.27%) |
| original22-v1/runtime/simd-memory-widen | engineSetup | 5.542 (ref) | 4.604 (+20.36%) | 7.875 (-29.63%) | 56.834 (-90.25%) | 124.21 (-95.54%) | 41.750 (-86.73%) | 24.333 (-77.23%) | 12.709 (-56.40%) |
| original22-v1/runtime/simd-memory-widen | compileValidate | 258.58 (ref) | 232.48 (+11.23%) | 229.52 (+12.66%) | 132.12 (+95.71%) | 927.73 (-72.13%) | 1,167.12 (-77.84%) | 276,925.94 (NC) | 284,488.29 (NC) |
| original22-v1/runtime/simd-memory-widen | instantiate | 13.792 (ref) | 14.812 (-6.89%) | 16.709 (-17.46%) | 18.104 (-23.82%) | 54.521 (-74.70%) | 35.895 (-61.58%) | 41.708 (-66.93%) | 50.479 (-72.68%) |
| original22-v1/runtime/simd-memory-widen | exportLookup | 0.250 (ref) | 0.438 (-42.86%) | 0.438 (-42.92%) | 2.125 (-88.24%) | 1.875 (-86.67%) | 4.521 (-94.47%) | 0.251 (-0.20%) | 0.229 (+9.17%) |
| original22-v1/runtime/simd-memory-widen | firstCall | 1,056.44 (ref) | 1,025.10 (+3.06%) | 1,058.71 (-0.21%) | 1,047.69 (+0.84%) | 1,021.58 (+3.41%) | 1,051.06 (+0.51%) | 1,079.19 (-2.11%) | 1,079.10 (-2.10%) |
| original22-v1/runtime/simd-memory-widen | cleanup | 16.854 (ref) | 16.396 (+2.79%) | 15.521 (+8.59%) | 9.521 (+77.02%) | 130.46 (-87.08%) | 60.229 (-72.02%) | 453.38 (-96.28%) | 368.58 (-95.43%) |
| original22-v1/runtime/simd-memory-lanes | processStartup | 4,199.69 (ref) | 6,064.04 (-30.74%) | 5,909.92 (-28.94%) | 4,311.94 (-2.60%) | 4,321.88 (-2.83%) | 5,927.00 (-29.14%) | 16,198.42 (-74.07%) | 11,861.06 (-64.59%) |
| original22-v1/runtime/simd-memory-lanes | load | 39.625 (ref) | 42.938 (-7.71%) | 44.166 (-10.28%) | 39.001 (+1.60%) | 32.645 (+21.38%) | 41.000 (-3.35%) | 46.500 (-14.78%) | 50.250 (-21.14%) |
| original22-v1/runtime/simd-memory-lanes | engineSetup | 4.771 (ref) | 4.500 (+6.02%) | 6.312 (-24.42%) | 53.979 (-91.16%) | 122.19 (-96.10%) | 40.688 (-88.27%) | 24.834 (-80.79%) | 13.709 (-65.20%) |
| original22-v1/runtime/simd-memory-lanes | compileValidate | 278.81 (ref) | 247.79 (+12.52%) | 242.08 (+15.17%) | 125.75 (+121.72%) | 962.10 (-71.02%) | 1,143.81 (-75.62%) | 296,864.44 (NC) | 277,499.65 (NC) |
| original22-v1/runtime/simd-memory-lanes | instantiate | 12.584 (ref) | 15.396 (-18.27%) | 15.896 (-20.84%) | 18.291 (-31.21%) | 51.875 (-75.74%) | 33.333 (-62.25%) | 42.583 (-70.45%) | 51.354 (-75.50%) |
| original22-v1/runtime/simd-memory-lanes | exportLookup | 0.250 (ref) | 0.416 (-39.90%) | 0.438 (-42.86%) | 1.938 (-87.10%) | 1.896 (-86.81%) | 4.812 (-94.81%) | 0.188 (+33.33%) | 0.271 (-7.58%) |
| original22-v1/runtime/simd-memory-lanes | firstCall | 5,019.29 (ref) | 4,971.25 (+0.97%) | 4,947.15 (+1.46%) | 5,951.29 (-15.66%) | 5,811.69 (-13.63%) | 5,678.27 (-11.61%) | 5,111.81 (-1.81%) | 5,018.35 (+0.02%) |
| original22-v1/runtime/simd-memory-lanes | cleanup | 9.750 (ref) | 7.812 (+24.80%) | 8.688 (+12.23%) | 4.521 (+115.66%) | 104.94 (-90.71%) | 56.146 (-82.63%) | 455.71 (-97.86%) | 377.71 (-97.42%) |
| original22-v1/gc/i31-roundtrip | processStartup | 4,107.08 (ref) | 5,788.19 (-29.04%) | 5,738.94 (-28.43%) | US | 4,044.17 (+1.56%) | US | 15,981.81 (-74.30%) | 12,044.71 (-65.90%) |
| original22-v1/gc/i31-roundtrip | load | 40.062 (ref) | 35.583 (+12.59%) | 39.959 (+0.26%) | US | 33.312 (+20.26%) | US | 41.751 (-4.04%) | 40.396 (-0.83%) |
| original22-v1/gc/i31-roundtrip | engineSetup | 5.479 (ref) | 4.021 (+36.29%) | 6.812 (-19.57%) | US | 122.27 (-95.52%) | US | 24.666 (-77.79%) | 13.021 (-57.92%) |
| original22-v1/gc/i31-roundtrip | compileValidate | 207.38 (ref) | 152.35 (+36.11%) | 153.27 (+35.30%) | US | 689.62 (-69.93%) | US | 280,650.10 (NC) | 277,446.69 (NC) |
| original22-v1/gc/i31-roundtrip | instantiate | 8.146 (ref) | 9.771 (-16.64%) | 10.375 (-21.49%) | US | 7.562 (+7.71%) | US | 17.750 (-54.11%) | 22.834 (-64.33%) |
| original22-v1/gc/i31-roundtrip | exportLookup | 0.250 (ref) | 0.375 (-33.33%) | 0.375 (-33.33%) | US | 3.833 (-93.48%) | US | 0.292 (-14.38%) | 0.229 (+9.17%) |
| original22-v1/gc/i31-roundtrip | firstCall | 24.562 (ref) | 21.333 (+15.14%) | 21.063 (+16.61%) | US | 44.021 (-44.20%) | US | 28.646 (-14.26%) | 45.791 (-46.36%) |
| original22-v1/gc/i31-roundtrip | cleanup | 4.333 (ref) | 4.479 (-3.26%) | 3.979 (+8.91%) | US | 55.166 (-92.14%) | US | 378.62 (-98.86%) | 276.35 (-98.43%) |
| original22-v1/gc/struct-access | processStartup | 4,187.00 (ref) | 5,864.87 (-28.61%) | 5,816.90 (-28.02%) | US | 4,240.33 (-1.26%) | US | 16,107.96 (-74.01%) | 12,044.56 (-65.24%) |
| original22-v1/gc/struct-access | load | 42.125 (ref) | 39.750 (+5.97%) | 41.895 (+0.55%) | US | 33.000 (+27.65%) | US | 44.084 (-4.44%) | 43.062 (-2.18%) |
| original22-v1/gc/struct-access | engineSetup | 5.312 (ref) | 4.833 (+9.91%) | 6.854 (-22.49%) | US | 121.46 (-95.63%) | US | 24.938 (-78.70%) | 12.104 (-56.11%) |
| original22-v1/gc/struct-access | compileValidate | 232.31 (ref) | 192.77 (+20.51%) | 188.35 (+23.34%) | US | 818.06 (-71.60%) | US | 281,900.40 (NC) | 284,547.62 (NC) |
| original22-v1/gc/struct-access | instantiate | 8.896 (ref) | 10.375 (-14.26%) | 11.125 (-20.04%) | US | 21.541 (-58.71%) | US | 13.916 (-36.08%) | 23.375 (-61.94%) |
| original22-v1/gc/struct-access | exportLookup | 0.208 (ref) | 0.375 (-44.40%) | 0.354 (-41.10%) | US | 3.854 (-94.59%) | US | 0.230 (-9.15%) | 0.230 (-9.15%) |
| original22-v1/gc/struct-access | firstCall | 31.229 (ref) | 28.646 (+9.02%) | 27.584 (+13.22%) | US | 50.770 (-38.49%) | US | 133.35 (-76.58%) | 104.83 (-70.21%) |
| original22-v1/gc/struct-access | cleanup | 11.000 (ref) | 9.625 (+14.29%) | 11.374 (-3.29%) | US | 69.166 (-84.10%) | US | 379.81 (-97.10%) | 272.96 (-95.97%) |
| original22-v1/gc/struct-allocation | processStartup | 3,975.02 (ref) | 5,771.46 (-31.13%) | 5,950.77 (-33.20%) | US | 4,113.60 (-3.37%) | US | 16,075.67 (-75.27%) | 11,926.02 (-66.67%) |
| original22-v1/gc/struct-allocation | load | 35.084 (ref) | 36.709 (-4.43%) | 43.626 (-19.58%) | US | 33.541 (+4.60%) | US | 44.834 (-21.75%) | 43.208 (-18.80%) |
| original22-v1/gc/struct-allocation | engineSetup | 5.792 (ref) | 4.500 (+28.71%) | 6.896 (-16.00%) | US | 126.54 (-95.42%) | US | 24.479 (-76.34%) | 12.604 (-54.05%) |
| original22-v1/gc/struct-allocation | compileValidate | 232.08 (ref) | 173.83 (+33.51%) | 174.17 (+33.25%) | US | 1,102.00 (-78.94%) | US | 276,235.71 (NC) | 282,565.38 (NC) |
| original22-v1/gc/struct-allocation | instantiate | 10.208 (ref) | 10.688 (-4.49%) | 11.354 (-10.10%) | US | 49.647 (-79.44%) | US | 19.979 (-48.91%) | 26.584 (-61.60%) |
| original22-v1/gc/struct-allocation | exportLookup | 0.208 (ref) | 0.416 (-50.06%) | 0.375 (-44.53%) | US | 1.730 (-87.97%) | US | 0.209 (-0.48%) | 0.208 (+0.00%) |
| original22-v1/gc/struct-allocation | firstCall | 356.31 (ref) | 370.92 (-3.94%) | 380.98 (-6.47%) | US | 145.92 (+144.19%) | US | 508.02 (-29.86%) | 458.15 (-22.23%) |
| original22-v1/gc/struct-allocation | cleanup | 129.23 (ref) | 87.229 (+48.15%) | 88.833 (+45.47%) | US | 98.999 (+30.53%) | US | 1,904.62 (-93.22%) | 1,796.83 (-92.81%) |
| original22-v1/gc/struct-indirect-access | processStartup | 4,252.31 (ref) | 6,120.44 (-30.52%) | 5,948.81 (-28.52%) | US | 4,270.06 (-0.42%) | US | 16,016.52 (-73.45%) | 11,908.40 (-64.29%) |
| original22-v1/gc/struct-indirect-access | load | 40.438 (ref) | 45.084 (-10.31%) | 42.208 (-4.19%) | US | 34.916 (+15.81%) | US | 45.480 (-11.09%) | 45.417 (-10.96%) |
| original22-v1/gc/struct-indirect-access | engineSetup | 5.958 (ref) | 5.083 (+17.22%) | 6.125 (-2.72%) | US | 125.58 (-95.26%) | US | 24.791 (-75.97%) | 13.584 (-56.13%) |
| original22-v1/gc/struct-indirect-access | compileValidate | 269.75 (ref) | 235.25 (+14.67%) | 228.35 (+18.13%) | US | 1,464.98 (-81.59%) | US | 279,555.17 (NC) | 277,791.87 (NC) |
| original22-v1/gc/struct-indirect-access | instantiate | 20.312 (ref) | 23.291 (-12.79%) | 21.541 (-5.71%) | US | 59.188 (-65.68%) | US | 22.062 (-7.93%) | 29.041 (-30.06%) |
| original22-v1/gc/struct-indirect-access | exportLookup | 0.250 (ref) | 0.375 (-33.33%) | 0.458 (-45.41%) | US | 1.854 (-86.52%) | US | 0.250 (+0.00%) | 0.250 (+0.00%) |
| original22-v1/gc/struct-indirect-access | firstCall | 29.729 (ref) | 24.958 (+19.12%) | 25.063 (+18.62%) | US | 29.791 (-0.21%) | US | 309.19 (-90.38%) | 186.83 (-84.09%) |
| original22-v1/gc/struct-indirect-access | cleanup | 13.167 (ref) | 11.416 (+15.33%) | 13.938 (-5.53%) | US | 122.71 (-89.27%) | US | 395.79 (-96.67%) | 299.46 (-95.60%) |
| original22-v1/threads/private-memory | processStartup | 4,015.33 (ref) | 5,992.71 (-33.00%) | 6,184.31 (-35.07%) | 4,191.12 (-4.19%) | 4,320.92 (-7.07%) | 6,023.77 (-33.34%) | 15,952.73 (-74.83%) | 11,749.37 (-65.83%) |
| original22-v1/threads/private-memory | load | 41.897 (ref) | 42.521 (-1.47%) | 46.146 (-9.21%) | 39.938 (+4.91%) | 33.791 (+23.99%) | 39.416 (+6.29%) | 40.730 (+2.87%) | 47.125 (-11.09%) |
| original22-v1/threads/private-memory | engineSetup | 5.854 (ref) | 4.312 (+35.74%) | 7.000 (-16.37%) | 57.292 (-89.78%) | 125.46 (-95.33%) | 43.375 (-86.50%) | 25.729 (-77.25%) | 13.771 (-57.49%) |
| original22-v1/threads/private-memory | compileValidate | 231.81 (ref) | 186.19 (+24.50%) | 186.23 (+24.48%) | 128.46 (+80.46%) | 690.98 (-66.45%) | 1,093.92 (-78.81%) | 277,568.67 (NC) | 288,489.44 (NC) |
| original22-v1/threads/private-memory | instantiate | 13.084 (ref) | 13.584 (-3.68%) | 14.542 (-10.03%) | 16.187 (-19.17%) | 20.146 (-35.06%) | 30.750 (-57.45%) | 36.938 (-64.58%) | 41.480 (-68.46%) |
| original22-v1/threads/private-memory | exportLookup | 0.230 (ref) | 0.375 (-38.80%) | 0.416 (-44.90%) | 1.937 (-88.15%) | 3.708 (-93.81%) | 4.292 (-94.65%) | 0.250 (-8.20%) | 0.230 (+0.00%) |
| original22-v1/threads/private-memory | firstCall | 15.229 (ref) | 12.750 (+19.44%) | 12.354 (+23.27%) | 18.562 (-17.96%) | 41.334 (-63.16%) | 25.104 (-39.34%) | 32.833 (-53.62%) | 49.625 (-69.31%) |
| original22-v1/threads/private-memory | cleanup | 6.479 (ref) | 6.333 (+2.30%) | 9.583 (-32.39%) | 6.958 (-6.89%) | 76.167 (-91.49%) | 52.145 (-87.58%) | 408.06 (-98.41%) | 352.48 (-98.16%) |
| original22-v1/threads/shared-memory | processStartup | 4,141.50 (ref) | 5,787.56 (-28.44%) | 6,010.19 (-31.09%) | 4,188.00 (-1.11%) | 4,285.75 (-3.37%) | 5,960.02 (-30.51%) | 16,146.31 (-74.35%) | 11,852.17 (-65.06%) |
| original22-v1/threads/shared-memory | load | 39.000 (ref) | 39.062 (-0.16%) | 41.687 (-6.45%) | 36.376 (+7.22%) | 41.146 (-5.22%) | 43.125 (-9.57%) | 43.416 (-10.17%) | 48.770 (-20.03%) |
| original22-v1/threads/shared-memory | engineSetup | 5.542 (ref) | 4.188 (+32.35%) | 7.229 (-23.34%) | 58.084 (-90.46%) | 129.88 (-95.73%) | 42.604 (-86.99%) | 26.729 (-79.27%) | 14.021 (-60.47%) |
| original22-v1/threads/shared-memory | compileValidate | 242.19 (ref) | 184.77 (+31.08%) | 178.35 (+35.79%) | 122.71 (+97.37%) | 736.27 (-67.11%) | 1,111.75 (-78.22%) | 272,254.50 (NC) | 285,387.31 (NC) |
| original22-v1/threads/shared-memory | instantiate | 12.791 (ref) | 14.167 (-9.71%) | 14.208 (-9.97%) | 15.584 (-17.92%) | 19.854 (-35.57%) | 35.250 (-63.71%) | 35.479 (-63.95%) | 43.271 (-70.44%) |
| original22-v1/threads/shared-memory | exportLookup | 0.250 (ref) | 0.396 (-36.87%) | 0.438 (-42.86%) | 2.312 (-89.19%) | 3.979 (-93.72%) | 4.438 (-94.37%) | 0.230 (+8.93%) | 0.208 (+19.90%) |
| original22-v1/threads/shared-memory | firstCall | 26.146 (ref) | 23.875 (+9.51%) | 21.646 (+20.79%) | 18.292 (+42.94%) | 44.251 (-40.91%) | 25.438 (+2.79%) | 32.479 (-19.50%) | 49.562 (-47.25%) |
| original22-v1/threads/shared-memory | cleanup | 9.604 (ref) | 7.146 (+34.40%) | 5.667 (+69.49%) | 5.021 (+91.28%) | 81.355 (-88.19%) | 53.250 (-81.96%) | 413.62 (-97.68%) | 335.08 (-97.13%) |
| original22-v1/threads/atomic-rmw | processStartup | 4,106.27 (ref) | 5,793.52 (-29.12%) | 5,873.48 (-30.09%) | 4,080.48 (+0.63%) | 4,163.12 (-1.37%) | 6,062.62 (-32.27%) | 16,164.75 (-74.60%) | 12,009.25 (-65.81%) |
| original22-v1/threads/atomic-rmw | load | 36.084 (ref) | 40.604 (-11.13%) | 42.188 (-14.47%) | 35.855 (+0.64%) | 36.166 (-0.23%) | 39.438 (-8.50%) | 42.271 (-14.64%) | 45.021 (-19.85%) |
| original22-v1/threads/atomic-rmw | engineSetup | 5.417 (ref) | 4.229 (+28.06%) | 7.021 (-22.85%) | 57.125 (-90.52%) | 121.12 (-95.53%) | 41.791 (-87.04%) | 23.083 (-76.53%) | 13.438 (-59.69%) |
| original22-v1/threads/atomic-rmw | compileValidate | 215.54 (ref) | 160.21 (+34.54%) | 166.83 (+29.20%) | 129.90 (+65.93%) | 723.67 (-70.22%) | 1,122.92 (-80.81%) | 279,137.27 (NC) | 278,337.98 (NC) |
| original22-v1/threads/atomic-rmw | instantiate | 13.021 (ref) | 14.896 (-12.59%) | 16.500 (-21.09%) | 18.125 (-28.16%) | 20.812 (-37.44%) | 35.438 (-63.26%) | 34.584 (-62.35%) | 41.854 (-68.89%) |
| original22-v1/threads/atomic-rmw | exportLookup | 0.250 (ref) | 0.396 (-36.87%) | 0.438 (-42.86%) | 1.854 (-86.52%) | 3.729 (-93.30%) | 4.333 (-94.23%) | 0.208 (+20.19%) | 0.230 (+8.93%) |
| original22-v1/threads/atomic-rmw | firstCall | 32.041 (ref) | 28.459 (+12.59%) | 28.188 (+13.67%) | 26.895 (+19.13%) | 48.584 (-34.05%) | 53.791 (-40.43%) | 39.479 (-18.84%) | 54.792 (-41.52%) |
| original22-v1/threads/atomic-rmw | cleanup | 9.854 (ref) | 4.771 (+106.57%) | 4.854 (+103.02%) | 3.771 (+161.32%) | 67.584 (-85.42%) | 46.916 (-79.00%) | 393.81 (-97.50%) | 343.38 (-97.13%) |
| original22-v1/threads/notify-empty | processStartup | 4,135.44 (ref) | 6,009.19 (-31.18%) | 5,979.54 (-30.84%) | 4,182.15 (-1.12%) | 4,222.21 (-2.06%) | 6,150.44 (-32.76%) | 16,096.25 (-74.31%) | 11,893.15 (-65.23%) |
| original22-v1/threads/notify-empty | load | 40.375 (ref) | 38.229 (+5.61%) | 43.479 (-7.14%) | 41.083 (-1.72%) | 40.334 (+0.10%) | 41.021 (-1.57%) | 45.709 (-11.67%) | 44.145 (-8.54%) |
| original22-v1/threads/notify-empty | engineSetup | 5.917 (ref) | 4.708 (+25.68%) | 6.083 (-2.73%) | 57.041 (-89.63%) | 130.69 (-95.47%) | 41.291 (-85.67%) | 26.375 (-77.57%) | 12.854 (-53.97%) |
| original22-v1/threads/notify-empty | compileValidate | 219.23 (ref) | 170.54 (+28.55%) | 169.52 (+29.32%) | 123.75 (+77.15%) | 843.69 (-74.02%) | 1,101.46 (-80.10%) | 282,598.25 (NC) | 277,169.83 (NC) |
| original22-v1/threads/notify-empty | instantiate | 12.417 (ref) | 13.624 (-8.86%) | 15.000 (-17.22%) | 17.062 (-27.23%) | 20.688 (-39.98%) | 36.437 (-65.92%) | 36.230 (-65.73%) | 44.230 (-71.93%) |
| original22-v1/threads/notify-empty | exportLookup | 0.250 (ref) | 0.375 (-33.33%) | 0.375 (-33.33%) | 1.958 (-87.24%) | 3.875 (-93.55%) | 4.250 (-94.12%) | 0.230 (+8.93%) | 0.250 (+0.00%) |
| original22-v1/threads/notify-empty | firstCall | 23.084 (ref) | 17.938 (+28.68%) | 18.021 (+28.09%) | 222.77 (-89.64%) | 101.08 (-77.16%) | 61.458 (-62.44%) | 114.63 (-79.86%) | 106.56 (-78.34%) |
| original22-v1/threads/notify-empty | cleanup | 9.958 (ref) | 8.084 (+23.19%) | 5.729 (+73.80%) | 3.667 (+171.56%) | 88.709 (-88.77%) | 49.521 (-79.89%) | 416.52 (-97.61%) | 354.79 (-97.19%) |
| original22-v1/threads/wait-mismatch | processStartup | 4,105.69 (ref) | 6,032.17 (-31.94%) | 6,164.12 (-33.39%) | 4,191.42 (-2.05%) | 4,205.29 (-2.37%) | 6,088.60 (-32.57%) | 16,081.10 (-74.47%) | 11,957.92 (-65.67%) |
| original22-v1/threads/wait-mismatch | load | 43.709 (ref) | 42.354 (+3.20%) | 48.791 (-10.42%) | 42.833 (+2.04%) | 36.126 (+20.99%) | 40.334 (+8.37%) | 47.251 (-7.50%) | 47.396 (-7.78%) |
| original22-v1/threads/wait-mismatch | engineSetup | 5.708 (ref) | 4.500 (+26.84%) | 6.687 (-14.64%) | 55.166 (-89.65%) | 126.69 (-95.49%) | 43.000 (-86.73%) | 26.833 (-78.73%) | 12.479 (-54.26%) |
| original22-v1/threads/wait-mismatch | compileValidate | 224.08 (ref) | 174.46 (+28.45%) | 175.21 (+27.90%) | 128.79 (+73.99%) | 868.83 (-74.21%) | 1,148.29 (-80.49%) | 277,637.13 (NC) | 311,903.38 (NC) |
| original22-v1/threads/wait-mismatch | instantiate | 13.312 (ref) | 14.584 (-8.72%) | 15.271 (-12.83%) | 17.854 (-25.44%) | 20.959 (-36.48%) | 36.062 (-63.09%) | 36.562 (-63.59%) | 43.125 (-69.13%) |
| original22-v1/threads/wait-mismatch | exportLookup | 0.250 (ref) | 0.334 (-25.15%) | 0.417 (-40.05%) | 2.812 (-91.11%) | 3.688 (-93.22%) | 4.729 (-94.71%) | 0.250 (+0.00%) | 0.209 (+19.62%) |
| original22-v1/threads/wait-mismatch | firstCall | 24.229 (ref) | 20.062 (+20.77%) | 19.417 (+24.78%) | 262.31 (-90.76%) | 122.08 (-80.15%) | 37.000 (-34.52%) | 52.791 (-54.10%) | 63.125 (-61.62%) |
| original22-v1/threads/wait-mismatch | cleanup | 10.854 (ref) | 7.500 (+44.72%) | 6.875 (+57.88%) | 6.167 (+76.02%) | 90.812 (-88.05%) | 55.792 (-80.55%) | 410.40 (-97.36%) | 350.13 (-96.90%) |
| original22-v1/threads/wait-timeout | processStartup | 4,106.29 (ref) | 5,874.75 (-30.10%) | 5,815.40 (-29.39%) | 4,134.50 (-0.68%) | 4,144.33 (-0.92%) | 5,877.48 (-30.14%) | 15,740.19 (-73.91%) | 11,794.00 (-65.18%) |
| original22-v1/threads/wait-timeout | load | 37.520 (ref) | 38.500 (-2.54%) | 39.812 (-5.76%) | 38.562 (-2.70%) | 34.834 (+7.71%) | 44.375 (-15.45%) | 52.459 (-28.48%) | 45.562 (-17.65%) |
| original22-v1/threads/wait-timeout | engineSetup | 5.208 (ref) | 4.104 (+26.90%) | 7.104 (-26.69%) | 55.895 (-90.68%) | 123.38 (-95.78%) | 40.541 (-87.15%) | 27.041 (-80.74%) | 13.334 (-60.94%) |
| original22-v1/threads/wait-timeout | compileValidate | 216.33 (ref) | 173.06 (+25.00%) | 165.90 (+30.40%) | 128.94 (+67.78%) | 838.65 (-74.20%) | 1,109.98 (-80.51%) | 272,929.00 (NC) | 277,921.75 (NC) |
| original22-v1/threads/wait-timeout | instantiate | 12.750 (ref) | 14.396 (-11.43%) | 14.979 (-14.88%) | 17.687 (-27.91%) | 21.188 (-39.82%) | 34.791 (-63.35%) | 33.938 (-62.43%) | 40.834 (-68.78%) |
| original22-v1/threads/wait-timeout | exportLookup | 0.250 (ref) | 0.417 (-40.05%) | 0.375 (-33.33%) | 2.249 (-88.89%) | 3.771 (-93.37%) | 4.542 (-94.50%) | 0.229 (+9.17%) | 0.230 (+8.93%) |
| original22-v1/threads/wait-timeout | firstCall | 23.062 (ref) | 18.375 (+25.51%) | 18.230 (+26.51%) | 2,652.75 (-99.13%) | 469.38 (-95.09%) | 1,816.02 (-98.73%) | 651.42 (-96.46%) | 668.40 (-96.55%) |
| original22-v1/threads/wait-timeout | cleanup | 8.500 (ref) | 4.771 (+78.18%) | 4.708 (+80.54%) | 3.146 (+170.18%) | 66.438 (-87.21%) | 50.084 (-83.03%) | 396.31 (-97.86%) | 314.75 (-97.30%) |
| original22-v1/applications/sha256-clang | processStartup | 4,132.96 (ref) | 6,106.85 (-32.32%) | 6,044.71 (-31.63%) | 4,359.15 (-5.19%) | 4,366.33 (-5.34%) | 6,173.73 (-33.06%) | 16,077.23 (-74.29%) | 12,206.42 (-66.14%) |
| original22-v1/applications/sha256-clang | load | 41.938 (ref) | 41.270 (+1.62%) | 47.646 (-11.98%) | 42.084 (-0.35%) | 36.416 (+15.16%) | 44.791 (-6.37%) | 52.729 (-20.47%) | 51.729 (-18.93%) |
| original22-v1/applications/sha256-clang | engineSetup | 5.458 (ref) | 4.542 (+20.17%) | 6.625 (-17.62%) | 57.541 (-90.51%) | 130.44 (-95.82%) | 41.438 (-86.83%) | 28.104 (-80.58%) | 13.812 (-60.49%) |
| original22-v1/applications/sha256-clang | compileValidate | 1,348.81 (ref) | 1,585.21 (-14.91%) | 1,554.15 (-13.21%) | 299.04 (+351.05%) | 1,354.73 (-0.44%) | 1,775.19 (-24.02%) | 312,502.35 (NC) | 300,400.10 (NC) |
| original22-v1/applications/sha256-clang | instantiate | 16.417 (ref) | 17.521 (-6.30%) | 18.771 (-12.54%) | 15.479 (+6.06%) | 55.562 (-70.45%) | 37.605 (-56.34%) | 45.938 (-64.26%) | 54.959 (-70.13%) |
| original22-v1/applications/sha256-clang | exportLookup | 0.250 (ref) | 0.375 (-33.33%) | 0.459 (-45.47%) | 2.333 (-89.29%) | 1.958 (-87.24%) | 5.021 (-95.02%) | 0.208 (+20.19%) | 0.230 (+8.93%) |
| original22-v1/applications/sha256-clang | firstCall | 1,023.21 (ref) | 1,003.19 (+2.00%) | 1,000.06 (+2.31%) | 1,360.73 (-24.80%) | 976.48 (+4.79%) | 940.96 (+8.74%) | 874.71 (+16.98%) | 899.75 (+13.72%) |
| original22-v1/applications/sha256-clang | cleanup | 19.188 (ref) | 16.396 (+17.03%) | 19.083 (+0.55%) | 13.000 (+47.60%) | 189.06 (-89.85%) | 82.938 (-76.87%) | 493.79 (-96.11%) | 394.29 (-95.13%) |
| original22-v1/applications/sha256-rust | processStartup | 4,148.56 (ref) | 5,867.19 (-29.29%) | 6,007.33 (-30.94%) | 4,226.52 (-1.84%) | 4,232.92 (-1.99%) | 6,084.06 (-31.81%) | 16,353.42 (-74.63%) | 11,854.63 (-65.00%) |
| original22-v1/applications/sha256-rust | load | 41.126 (ref) | 40.688 (+1.08%) | 42.896 (-4.13%) | 43.666 (-5.82%) | 42.292 (-2.76%) | 39.479 (+4.17%) | 47.834 (-14.02%) | 46.876 (-12.27%) |
| original22-v1/applications/sha256-rust | engineSetup | 5.333 (ref) | 4.833 (+10.34%) | 6.792 (-21.47%) | 62.270 (-91.43%) | 126.73 (-95.79%) | 39.645 (-86.55%) | 24.792 (-78.49%) | 12.229 (-56.39%) |
| original22-v1/applications/sha256-rust | compileValidate | 1,311.06 (ref) | 1,549.06 (-15.36%) | 1,515.13 (-13.47%) | 316.04 (+314.84%) | 1,359.31 (-3.55%) | 1,739.60 (-24.63%) | 280,746.40 (NC) | 295,505.83 (NC) |
| original22-v1/applications/sha256-rust | instantiate | 16.438 (ref) | 19.395 (-15.25%) | 18.938 (-13.20%) | 18.229 (-9.83%) | 56.541 (-70.93%) | 36.334 (-54.76%) | 46.667 (-64.78%) | 53.833 (-69.47%) |
| original22-v1/applications/sha256-rust | exportLookup | 0.250 (ref) | 0.458 (-45.41%) | 0.417 (-40.05%) | 2.354 (-89.38%) | 1.875 (-86.67%) | 4.917 (-94.92%) | 0.209 (+19.62%) | 0.208 (+19.90%) |
| original22-v1/applications/sha256-rust | firstCall | 1,060.27 (ref) | 1,070.42 (-0.95%) | 1,046.94 (+1.27%) | 1,471.83 (-27.96%) | 945.73 (+12.11%) | 945.35 (+12.16%) | 922.46 (+14.94%) | 938.92 (+12.92%) |
| original22-v1/applications/sha256-rust | cleanup | 11.583 (ref) | 13.334 (-13.13%) | 11.062 (+4.71%) | 7.250 (+59.78%) | 155.12 (-92.53%) | 68.396 (-83.06%) | 493.71 (-97.65%) | 480.27 (-97.59%) |
| gc-sustained-v1/gc/i31-roundtrip | processStartup | 4,235.40 (ref) | 6,025.17 (-29.70%) | 6,018.65 (-29.63%) | US | 4,335.04 (-2.30%) | US | 15,975.65 (-73.49%) | 12,041.40 (-64.83%) |
| gc-sustained-v1/gc/i31-roundtrip | load | 41.125 (ref) | 38.667 (+6.36%) | 43.479 (-5.41%) | US | 36.813 (+11.71%) | US | 44.750 (-8.10%) | 46.770 (-12.07%) |
| gc-sustained-v1/gc/i31-roundtrip | engineSetup | 5.312 (ref) | 4.125 (+28.79%) | 6.354 (-16.40%) | US | 122.62 (-95.67%) | US | 25.667 (-79.30%) | 11.791 (-54.95%) |
| gc-sustained-v1/gc/i31-roundtrip | compileValidate | 202.58 (ref) | 148.19 (+36.71%) | 149.40 (+35.60%) | US | 731.33 (-72.30%) | US | 273,829.19 (NC) | 286,877.00 (NC) |
| gc-sustained-v1/gc/i31-roundtrip | instantiate | 7.833 (ref) | 9.541 (-17.90%) | 10.167 (-22.95%) | US | 7.333 (+6.82%) | US | 14.812 (-47.12%) | 22.062 (-64.49%) |
| gc-sustained-v1/gc/i31-roundtrip | exportLookup | 0.291 (ref) | 0.334 (-12.74%) | 0.375 (-22.40%) | US | 3.625 (-91.97%) | US | 0.292 (-0.34%) | 0.250 (+16.40%) |
| gc-sustained-v1/gc/i31-roundtrip | firstCall | 1,373.42 (ref) | 1,364.65 (+0.64%) | 1,357.92 (+1.14%) | US | 1,854.25 (-25.93%) | US | 1,310.25 (+4.82%) | 1,333.23 (+3.01%) |
| gc-sustained-v1/gc/i31-roundtrip | cleanup | 5.229 (ref) | 4.625 (+13.06%) | 4.271 (+22.43%) | US | 57.291 (-90.87%) | US | 411.75 (-98.73%) | 308.88 (-98.31%) |
| gc-sustained-v1/gc/struct-access | processStartup | 4,112.27 (ref) | 5,937.65 (-30.74%) | 6,083.98 (-32.41%) | US | 4,309.17 (-4.57%) | US | FAIL | 11,999.17 (-65.73%) |
| gc-sustained-v1/gc/struct-access | load | 38.750 (ref) | 41.271 (-6.11%) | 38.792 (-0.11%) | US | 40.291 (-3.83%) | US | FAIL | 44.062 (-12.06%) |
| gc-sustained-v1/gc/struct-access | engineSetup | 5.729 (ref) | 4.854 (+18.03%) | 6.042 (-5.17%) | US | 126.58 (-95.47%) | US | FAIL | 12.291 (-53.39%) |
| gc-sustained-v1/gc/struct-access | compileValidate | 241.48 (ref) | 188.12 (+28.36%) | 189.06 (+27.72%) | US | 812.15 (-70.27%) | US | FAIL | 292,715.96 (NC) |
| gc-sustained-v1/gc/struct-access | instantiate | 9.375 (ref) | 10.521 (-10.89%) | 11.062 (-15.25%) | US | 20.833 (-55.00%) | US | FAIL | 21.396 (-56.18%) |
| gc-sustained-v1/gc/struct-access | exportLookup | 0.230 (ref) | 0.333 (-31.08%) | 0.375 (-38.80%) | US | 3.750 (-93.88%) | US | FAIL | 0.209 (+9.81%) |
| gc-sustained-v1/gc/struct-access | firstCall | 1,062.42 (ref) | 1,070.62 (-0.77%) | 1,075.17 (-1.19%) | US | 1,082.73 (-1.88%) | US | FAIL | 6,891.12 (-84.58%) |
| gc-sustained-v1/gc/struct-access | cleanup | 14.166 (ref) | 14.312 (-1.02%) | 13.062 (+8.46%) | US | 82.125 (-82.75%) | US | FAIL | 315.27 (-95.51%) |
| gc-sustained-v1/gc/struct-allocation | processStartup | 4,481.79 (ref) | 6,157.65 (-27.22%) | 6,258.08 (-28.38%) | US | 4,385.62 (+2.19%) | US | FAIL | 11,985.42 (-62.61%) |
| gc-sustained-v1/gc/struct-allocation | load | 44.833 (ref) | 44.145 (+1.56%) | 45.562 (-1.60%) | US | 40.917 (+9.57%) | US | FAIL | 44.791 (+0.09%) |
| gc-sustained-v1/gc/struct-allocation | engineSetup | 5.458 (ref) | 4.562 (+19.64%) | 7.271 (-24.93%) | US | 126.17 (-95.67%) | US | FAIL | 12.376 (-55.89%) |
| gc-sustained-v1/gc/struct-allocation | compileValidate | 218.54 (ref) | 180.44 (+21.12%) | 180.35 (+21.17%) | US | 1,051.29 (-79.21%) | US | FAIL | 276,783.52 (NC) |
| gc-sustained-v1/gc/struct-allocation | instantiate | 9.729 (ref) | 11.062 (-12.05%) | 11.771 (-17.34%) | US | 46.333 (-79.00%) | US | FAIL | 27.792 (-64.99%) |
| gc-sustained-v1/gc/struct-allocation | exportLookup | 0.208 (ref) | 0.375 (-44.40%) | 0.333 (-37.39%) | US | 1.729 (-87.94%) | US | FAIL | 0.250 (-16.60%) |
| gc-sustained-v1/gc/struct-allocation | firstCall | 25,555.17 (ref) | 14,624.88 (+74.74%) | 14,745.40 (+73.31%) | US | 3,051.35 (+737.50%) | US | FAIL | 9,849.90 (+159.45%) |
| gc-sustained-v1/gc/struct-allocation | cleanup | 2,992.52 (ref) | 1,969.25 (+51.96%) | 1,986.62 (+50.63%) | US | 680.83 (+339.54%) | US | FAIL | 53,981.23 (-94.46%) |
| gc-sustained-v1/gc/struct-indirect-access | processStartup | 4,196.21 (ref) | 5,844.79 (-28.21%) | 5,958.29 (-29.57%) | US | 4,082.50 (+2.79%) | US | FAIL | 11,931.81 (-64.83%) |
| gc-sustained-v1/gc/struct-indirect-access | load | 40.896 (ref) | 45.145 (-9.41%) | 46.041 (-11.18%) | US | 35.854 (+14.06%) | US | FAIL | 48.354 (-15.42%) |
| gc-sustained-v1/gc/struct-indirect-access | engineSetup | 4.979 (ref) | 4.708 (+5.76%) | 7.041 (-29.29%) | US | 121.52 (-95.90%) | US | FAIL | 12.500 (-60.17%) |
| gc-sustained-v1/gc/struct-indirect-access | compileValidate | 270.38 (ref) | 226.44 (+19.40%) | 229.56 (+17.78%) | US | 1,498.96 (-81.96%) | US | FAIL | 288,100.19 (NC) |
| gc-sustained-v1/gc/struct-indirect-access | instantiate | 21.666 (ref) | 21.209 (+2.16%) | 21.250 (+1.96%) | US | 56.500 (-61.65%) | US | FAIL | 28.375 (-23.64%) |
| gc-sustained-v1/gc/struct-indirect-access | exportLookup | 0.250 (ref) | 0.438 (-42.86%) | 0.459 (-45.47%) | US | 1.855 (-86.52%) | US | FAIL | 0.208 (+20.19%) |
| gc-sustained-v1/gc/struct-indirect-access | firstCall | 1,665.52 (ref) | 1,661.90 (+0.22%) | 1,559.60 (+6.79%) | US | 2,594.67 (-35.81%) | US | FAIL | 15,043.98 (-88.93%) |
| gc-sustained-v1/gc/struct-indirect-access | cleanup | 13.604 (ref) | 14.229 (-4.39%) | 13.812 (-1.51%) | US | 106.71 (-87.25%) | US | FAIL | 331.04 (-95.89%) |
| threads-sustained-v2/threads/private-memory | processStartup | 4,120.06 (ref) | 6,008.60 (-31.43%) | 6,018.42 (-31.54%) | 4,343.75 (-5.15%) | 4,270.15 (-3.51%) | 6,058.33 (-31.99%) | 16,152.56 (-74.49%) | 12,039.77 (-65.78%) |
| threads-sustained-v2/threads/private-memory | load | 36.792 (ref) | 39.979 (-7.97%) | 39.792 (-7.54%) | 36.688 (+0.28%) | 31.875 (+15.43%) | 37.480 (-1.83%) | 48.479 (-24.11%) | 44.937 (-18.13%) |
| threads-sustained-v2/threads/private-memory | engineSetup | 5.083 (ref) | 4.334 (+17.29%) | 5.938 (-14.38%) | 55.271 (-90.80%) | 124.58 (-95.92%) | 38.730 (-86.87%) | 26.354 (-80.71%) | 14.042 (-63.80%) |
| threads-sustained-v2/threads/private-memory | compileValidate | 230.65 (ref) | 181.75 (+26.90%) | 177.17 (+30.19%) | 120.69 (+91.11%) | 725.60 (-68.21%) | 1,073.48 (-78.51%) | 280,321.35 (NC) | 278,733.58 (NC) |
| threads-sustained-v2/threads/private-memory | instantiate | 12.896 (ref) | 13.959 (-7.62%) | 14.146 (-8.84%) | 17.125 (-24.70%) | 18.521 (-30.37%) | 30.312 (-57.46%) | 36.166 (-64.34%) | 41.146 (-68.66%) |
| threads-sustained-v2/threads/private-memory | exportLookup | 0.250 (ref) | 0.375 (-33.33%) | 0.416 (-39.90%) | 2.063 (-87.88%) | 3.812 (-93.44%) | 4.604 (-94.57%) | 0.250 (+0.00%) | 0.250 (+0.00%) |
| threads-sustained-v2/threads/private-memory | firstCall | 39.209 (ref) | 35.688 (+9.86%) | 35.521 (+10.38%) | 116.08 (-66.22%) | 132.69 (-70.45%) | 117.94 (-66.75%) | 126.31 (-68.96%) | 139.54 (-71.90%) |
| threads-sustained-v2/threads/private-memory | cleanup | 6.104 (ref) | 5.875 (+3.92%) | 8.271 (-26.19%) | 3.896 (+56.69%) | 82.562 (-92.61%) | 52.959 (-88.47%) | 414.46 (-98.53%) | 335.67 (-98.18%) |
| threads-sustained-v2/threads/shared-memory | processStartup | 4,035.46 (ref) | 5,965.96 (-32.36%) | 5,806.08 (-30.50%) | 4,170.06 (-3.23%) | 4,284.58 (-5.81%) | 5,859.42 (-31.13%) | 15,623.13 (-74.17%) | 11,974.77 (-66.30%) |
| threads-sustained-v2/threads/shared-memory | load | 38.855 (ref) | 36.104 (+7.62%) | 40.729 (-4.60%) | 35.979 (+7.99%) | 26.270 (+47.90%) | 38.271 (+1.52%) | 40.416 (-3.86%) | 42.312 (-8.17%) |
| threads-sustained-v2/threads/shared-memory | engineSetup | 5.688 (ref) | 4.375 (+30.00%) | 6.646 (-14.42%) | 56.041 (-89.85%) | 117.21 (-95.15%) | 40.750 (-86.04%) | 25.271 (-77.49%) | 11.792 (-51.77%) |
| threads-sustained-v2/threads/shared-memory | compileValidate | 240.77 (ref) | 180.52 (+33.38%) | 176.08 (+36.74%) | 117.65 (+104.66%) | 698.88 (-65.55%) | 1,073.10 (-77.56%) | 274,714.67 (NC) | 288,118.10 (NC) |
| threads-sustained-v2/threads/shared-memory | instantiate | 12.667 (ref) | 13.625 (-7.03%) | 14.771 (-14.25%) | 18.438 (-31.30%) | 19.771 (-35.93%) | 34.209 (-62.97%) | 34.959 (-63.77%) | 41.709 (-69.63%) |
| threads-sustained-v2/threads/shared-memory | exportLookup | 0.271 (ref) | 0.375 (-27.87%) | 0.417 (-35.13%) | 2.208 (-87.75%) | 3.896 (-93.06%) | 4.562 (-94.07%) | 0.250 (+8.20%) | 0.271 (+0.00%) |
| threads-sustained-v2/threads/shared-memory | firstCall | 139.56 (ref) | 133.31 (+4.69%) | 132.42 (+5.40%) | 111.60 (+25.05%) | 131.25 (+6.33%) | 117.00 (+19.28%) | 125.19 (+11.48%) | 143.73 (-2.90%) |
| threads-sustained-v2/threads/shared-memory | cleanup | 9.271 (ref) | 5.021 (+84.64%) | 4.833 (+91.81%) | 3.688 (+151.42%) | 69.125 (-86.59%) | 48.938 (-81.06%) | 410.62 (-97.74%) | 316.69 (-97.07%) |
| threads-sustained-v2/threads/atomic-rmw | processStartup | 3,989.29 (ref) | 5,917.96 (-32.59%) | 5,892.52 (-32.30%) | 4,279.73 (-6.79%) | 4,476.19 (-10.88%) | 5,851.83 (-31.83%) | 16,065.48 (-75.17%) | 11,838.81 (-66.30%) |
| threads-sustained-v2/threads/atomic-rmw | load | 42.604 (ref) | 41.688 (+2.20%) | 42.104 (+1.19%) | 39.979 (+6.57%) | 34.208 (+24.54%) | 38.062 (+11.93%) | 43.062 (-1.06%) | 46.938 (-9.23%) |
| threads-sustained-v2/threads/atomic-rmw | engineSetup | 6.104 (ref) | 4.500 (+35.66%) | 6.375 (-4.25%) | 53.917 (-88.68%) | 130.10 (-95.31%) | 41.396 (-85.25%) | 25.479 (-76.04%) | 13.875 (-56.00%) |
| threads-sustained-v2/threads/atomic-rmw | compileValidate | 210.98 (ref) | 162.65 (+29.72%) | 166.19 (+26.95%) | 125.33 (+68.33%) | 702.06 (-69.95%) | 1,107.60 (-80.95%) | 276,548.79 (NC) | 278,775.40 (NC) |
| threads-sustained-v2/threads/atomic-rmw | instantiate | 13.271 (ref) | 15.416 (-13.92%) | 15.166 (-12.50%) | 16.854 (-21.26%) | 19.875 (-33.23%) | 33.979 (-60.94%) | 35.041 (-62.13%) | 42.980 (-69.12%) |
| threads-sustained-v2/threads/atomic-rmw | exportLookup | 0.250 (ref) | 0.417 (-40.05%) | 0.416 (-39.90%) | 2.021 (-87.63%) | 3.917 (-93.62%) | 4.312 (-94.20%) | 0.250 (+0.00%) | 0.250 (+0.00%) |
| threads-sustained-v2/threads/atomic-rmw | firstCall | 195.56 (ref) | 193.56 (+1.03%) | 191.60 (+2.07%) | 190.06 (+2.89%) | 211.35 (-7.47%) | 399.02 (-50.99%) | 203.73 (-4.01%) | 224.90 (-13.04%) |
| threads-sustained-v2/threads/atomic-rmw | cleanup | 8.750 (ref) | 4.583 (+90.92%) | 4.896 (+78.72%) | 3.792 (+130.78%) | 74.875 (-88.31%) | 54.229 (-83.86%) | 421.81 (-97.93%) | 338.67 (-97.42%) |
| threads-sustained-v2/threads/notify-empty | processStartup | 4,160.60 (ref) | 6,056.56 (-31.30%) | 6,186.92 (-32.75%) | 4,274.71 (-2.67%) | 4,223.79 (-1.50%) | 6,088.79 (-31.67%) | 16,089.06 (-74.14%) | 12,035.67 (-65.43%) |
| threads-sustained-v2/threads/notify-empty | load | 37.959 (ref) | 43.688 (-13.11%) | 42.354 (-10.38%) | 40.895 (-7.18%) | 31.688 (+19.79%) | 42.104 (-9.85%) | 46.042 (-17.56%) | 46.584 (-18.52%) |
| threads-sustained-v2/threads/notify-empty | engineSetup | 5.458 (ref) | 4.833 (+12.94%) | 7.333 (-25.56%) | 55.459 (-90.16%) | 127.23 (-95.71%) | 42.459 (-87.14%) | 24.958 (-78.13%) | 13.416 (-59.32%) |
| threads-sustained-v2/threads/notify-empty | compileValidate | 218.31 (ref) | 170.98 (+27.68%) | 174.06 (+25.42%) | 133.65 (+63.35%) | 835.21 (-73.86%) | 1,134.67 (-80.76%) | 275,399.65 (NC) | 295,897.50 (NC) |
| threads-sustained-v2/threads/notify-empty | instantiate | 12.771 (ref) | 14.479 (-11.80%) | 15.229 (-16.15%) | 15.625 (-18.27%) | 22.229 (-42.55%) | 34.770 (-63.27%) | 38.083 (-66.47%) | 42.375 (-69.86%) |
| threads-sustained-v2/threads/notify-empty | exportLookup | 0.250 (ref) | 0.417 (-40.05%) | 0.458 (-45.41%) | 1.938 (-87.10%) | 4.062 (-93.85%) | 4.312 (-94.20%) | 0.230 (+8.93%) | 0.250 (+0.00%) |
| threads-sustained-v2/threads/notify-empty | firstCall | 103.25 (ref) | 98.104 (+5.25%) | 99.980 (+3.27%) | 2,100.62 (-95.08%) | 716.77 (-85.60%) | 472.33 (-78.14%) | 926.69 (-88.86%) | 670.56 (-84.60%) |
| threads-sustained-v2/threads/notify-empty | cleanup | 11.729 (ref) | 11.479 (+2.18%) | 6.896 (+70.09%) | 4.125 (+184.32%) | 86.125 (-86.38%) | 53.396 (-78.03%) | 417.42 (-97.19%) | 342.02 (-96.57%) |
| threads-sustained-v2/threads/wait-mismatch | processStartup | 4,362.90 (ref) | 5,914.25 (-26.23%) | 5,845.52 (-25.36%) | 4,336.56 (+0.61%) | 4,169.33 (+4.64%) | 5,869.83 (-25.67%) | 15,948.31 (-72.64%) | 12,239.81 (-64.35%) |
| threads-sustained-v2/threads/wait-mismatch | load | 37.249 (ref) | 37.105 (+0.39%) | 36.688 (+1.53%) | 38.333 (-2.83%) | 34.480 (+8.03%) | 35.750 (+4.19%) | 42.583 (-12.52%) | 44.042 (-15.42%) |
| threads-sustained-v2/threads/wait-mismatch | engineSetup | 5.917 (ref) | 4.354 (+35.90%) | 7.333 (-19.31%) | 57.334 (-89.68%) | 126.73 (-95.33%) | 39.292 (-84.94%) | 26.562 (-77.72%) | 12.084 (-51.03%) |
| threads-sustained-v2/threads/wait-mismatch | compileValidate | 221.60 (ref) | 177.73 (+24.69%) | 170.94 (+29.64%) | 127.94 (+73.21%) | 813.83 (-72.77%) | 1,115.21 (-80.13%) | 283,079.31 (NC) | 284,067.42 (NC) |
| threads-sustained-v2/threads/wait-mismatch | instantiate | 13.166 (ref) | 14.479 (-9.06%) | 14.584 (-9.72%) | 15.167 (-13.19%) | 20.250 (-34.98%) | 36.167 (-63.60%) | 38.459 (-65.76%) | 40.854 (-67.77%) |
| threads-sustained-v2/threads/wait-mismatch | exportLookup | 0.250 (ref) | 0.416 (-39.98%) | 0.354 (-29.48%) | 2.417 (-89.65%) | 3.771 (-93.37%) | 4.583 (-94.55%) | 0.167 (+49.70%) | 0.249 (+0.20%) |
| threads-sustained-v2/threads/wait-mismatch | firstCall | 119.85 (ref) | 113.67 (+5.44%) | 110.88 (+8.10%) | 2,461.42 (-95.13%) | 918.83 (-86.96%) | 248.29 (-51.73%) | 330.54 (-63.74%) | 288.33 (-58.43%) |
| threads-sustained-v2/threads/wait-mismatch | cleanup | 9.125 (ref) | 5.542 (+64.65%) | 5.917 (+54.22%) | 3.771 (+141.98%) | 70.292 (-87.02%) | 54.563 (-83.28%) | 421.56 (-97.84%) | 343.21 (-97.34%) |
| threads-sustained-v2/threads/wait-timeout | processStartup | 4,137.17 (ref) | 5,878.29 (-29.62%) | 5,895.44 (-29.82%) | 4,382.06 (-5.59%) | 4,361.96 (-5.15%) | 6,064.83 (-31.78%) | 16,078.31 (-74.27%) | 12,113.71 (-65.85%) |
| threads-sustained-v2/threads/wait-timeout | load | 37.562 (ref) | 39.021 (-3.74%) | 43.855 (-14.35%) | 40.541 (-7.35%) | 32.770 (+14.62%) | 37.166 (+1.07%) | 44.020 (-14.67%) | 46.812 (-19.76%) |
| threads-sustained-v2/threads/wait-timeout | engineSetup | 5.354 (ref) | 4.208 (+27.23%) | 6.750 (-20.67%) | 54.333 (-90.15%) | 123.17 (-95.65%) | 40.041 (-86.63%) | 24.584 (-78.22%) | 12.209 (-56.14%) |
| threads-sustained-v2/threads/wait-timeout | compileValidate | 219.77 (ref) | 165.40 (+32.88%) | 169.42 (+29.72%) | 128.04 (+71.64%) | 778.33 (-71.76%) | 1,136.13 (-80.66%) | 281,012.35 (NC) | 278,212.15 (NC) |
| threads-sustained-v2/threads/wait-timeout | instantiate | 12.834 (ref) | 14.354 (-10.59%) | 15.604 (-17.76%) | 17.812 (-27.95%) | 20.438 (-37.21%) | 35.334 (-63.68%) | 36.395 (-64.74%) | 42.938 (-70.11%) |
| threads-sustained-v2/threads/wait-timeout | exportLookup | 0.250 (ref) | 0.458 (-45.41%) | 0.417 (-40.05%) | 2.375 (-89.47%) | 3.771 (-93.37%) | 4.625 (-94.59%) | 0.250 (+0.00%) | 0.292 (-14.38%) |
| threads-sustained-v2/threads/wait-timeout | firstCall | 106.77 (ref) | 98.604 (+8.28%) | 99.124 (+7.71%) | 27,215.81 (-99.61%) | 4,217.60 (-97.47%) | 17,150.46 (-99.38%) | 6,206.75 (-98.28%) | 5,863.60 (-98.18%) |
| threads-sustained-v2/threads/wait-timeout | cleanup | 14.124 (ref) | 5.438 (+159.76%) | 5.312 (+165.87%) | 3.479 (+305.93%) | 84.104 (-83.21%) | 60.250 (-76.56%) | 438.06 (-96.78%) | 359.48 (-96.07%) |
| generic-calls-v1/core/generic-calls | processStartup | 4,261.92 (ref) | 6,045.08 (-29.50%) | 5,949.56 (-28.37%) | 4,335.48 (-1.70%) | 4,540.35 (-6.13%) | 6,009.19 (-29.08%) | 16,384.60 (-73.99%) | 12,276.79 (-65.28%) |
| generic-calls-v1/core/generic-calls | load | 40.250 (ref) | 39.334 (+2.33%) | 48.084 (-16.29%) | 50.166 (-19.77%) | 36.730 (+9.58%) | 35.812 (+12.39%) | 46.875 (-14.13%) | 42.021 (-4.21%) |
| generic-calls-v1/core/generic-calls | engineSetup | 5.479 (ref) | 4.354 (+25.85%) | 7.042 (-22.19%) | 56.229 (-90.26%) | 123.56 (-95.57%) | 37.979 (-85.57%) | 25.834 (-78.79%) | 12.229 (-55.20%) |
| generic-calls-v1/core/generic-calls | compileValidate | 284.52 (ref) | 267.52 (+6.35%) | 252.83 (+12.53%) | 124.98 (+127.66%) | 883.19 (-67.78%) | 1,138.52 (-75.01%) | 274,279.21 (NC) | 294,760.69 (NC) |
| generic-calls-v1/core/generic-calls | instantiate | 8.812 (ref) | 10.812 (-18.50%) | 11.166 (-21.08%) | 13.604 (-35.22%) | 7.562 (+16.54%) | 16.396 (-46.25%) | 14.979 (-41.17%) | 23.709 (-62.83%) |
| generic-calls-v1/core/generic-calls | exportLookup | 0.230 (ref) | 0.396 (-41.97%) | 0.416 (-44.90%) | 2.604 (-91.19%) | 3.708 (-93.81%) | 4.292 (-94.65%) | 0.230 (+0.00%) | 0.209 (+9.81%) |
| generic-calls-v1/core/generic-calls | firstCall | 5,200.88 (ref) | 5,240.54 (-0.76%) | 5,186.12 (+0.28%) | 10,240.67 (-49.21%) | 10,327.48 (-49.64%) | 10,238.46 (-49.20%) | 5,254.04 (-1.01%) | 5,261.71 (-1.16%) |
| generic-calls-v1/core/generic-calls | cleanup | 11.146 (ref) | 7.729 (+44.20%) | 7.229 (+54.18%) | 4.646 (+139.93%) | 109.71 (-89.84%) | 39.000 (-71.42%) | 888.81 (-98.75%) | 825.94 (-98.65%) |

</details>

<details>
<summary>Interpreter lifecycle phase observations</summary>

| Workload | Phase | wasmvm Go | wasmvm LLGo (pinned) | wasmvm LLGo 42efacae | wazero 1.12.0 | Wasmtime 48.0.1 | Wasmer 7.3.0 | WasmEdge 0.17.1 | WasmEdge 0.18.1 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| original22-v1/core/call-control | processStartup | 4,054.15 (ref) | 5,737.31 (-29.34%) | 5,942.38 (-31.78%) | 4,234.21 (-4.25%) | NM | NM | 15,878.50 (-74.47%) | 11,730.02 (-65.44%) |
| original22-v1/core/call-control | load | 38.084 (ref) | 40.020 (-4.84%) | 42.479 (-10.35%) | 37.375 (+1.90%) | NM | NM | 42.479 (-10.35%) | 45.145 (-15.64%) |
| original22-v1/core/call-control | engineSetup | 4.271 (ref) | 4.042 (+5.65%) | 5.979 (-28.58%) | 2.375 (+79.81%) | NM | NM | 24.105 (-82.28%) | 12.291 (-65.26%) |
| original22-v1/core/call-control | compileValidate | 56.188 (ref) | 45.916 (+22.37%) | 44.812 (+25.38%) | 34.604 (+62.37%) | NM | NM | 34.646 (+62.18%) | 49.896 (+12.61%) |
| original22-v1/core/call-control | instantiate | 6.938 (ref) | 8.666 (-19.94%) | 9.188 (-24.48%) | 15.001 (-53.75%) | NM | NM | 23.062 (-69.92%) | 15.000 (-53.75%) |
| original22-v1/core/call-control | exportLookup | 0.208 (ref) | 0.375 (-44.53%) | 0.354 (-41.33%) | 0.167 (+24.55%) | NM | NM | 0.208 (+0.00%) | 0.291 (-28.64%) |
| original22-v1/core/call-control | firstCall | 6.188 (ref) | 5.812 (+6.45%) | 5.916 (+4.59%) | 3.396 (+82.23%) | NM | NM | 7.667 (-19.30%) | 10.209 (-39.39%) |
| original22-v1/core/call-control | cleanup | 5.437 (ref) | 1.333 (+307.88%) | 1.729 (+214.46%) | 15.042 (-63.85%) | NM | NM | 5.458 (-0.38%) | 7.521 (-27.70%) |
| original22-v1/core/integer-loop | processStartup | 4,131.83 (ref) | 6,124.85 (-32.54%) | 5,987.23 (-30.99%) | 4,408.83 (-6.28%) | NM | NM | 15,953.27 (-74.10%) | 11,833.98 (-65.09%) |
| original22-v1/core/integer-loop | load | 41.979 (ref) | 42.958 (-2.28%) | 42.396 (-0.98%) | 43.792 (-4.14%) | NM | NM | 44.917 (-6.54%) | 46.479 (-9.68%) |
| original22-v1/core/integer-loop | engineSetup | 4.271 (ref) | 4.312 (-0.96%) | 6.167 (-30.74%) | 2.625 (+62.70%) | NM | NM | 26.229 (-83.72%) | 10.812 (-60.50%) |
| original22-v1/core/integer-loop | compileValidate | 64.855 (ref) | 74.626 (-13.09%) | 53.605 (+20.99%) | 42.271 (+53.43%) | NM | NM | 36.625 (+77.08%) | 50.709 (+27.90%) |
| original22-v1/core/integer-loop | instantiate | 7.833 (ref) | 9.166 (-14.54%) | 10.062 (-22.15%) | 15.417 (-49.19%) | NM | NM | 23.854 (-67.16%) | 14.896 (-47.41%) |
| original22-v1/core/integer-loop | exportLookup | 0.230 (ref) | 0.354 (-35.26%) | 0.416 (-44.83%) | 0.188 (+22.07%) | NM | NM | 0.208 (+10.34%) | 0.291 (-21.27%) |
| original22-v1/core/integer-loop | firstCall | 28,469.08 (ref) | 23,900.15 (+19.12%) | 22,681.37 (+25.52%) | 82,270.15 (-65.40%) | NM | NM | 49,386.15 (-42.35%) | 24,562.31 (+15.91%) |
| original22-v1/core/integer-loop | cleanup | 5.729 (ref) | 1.917 (+198.96%) | 2.062 (+177.79%) | 4.271 (+34.16%) | NM | NM | 6.313 (-9.24%) | 11.354 (-49.54%) |
| original22-v1/core/function-calls | processStartup | 4,060.33 (ref) | 6,097.29 (-33.41%) | 6,174.54 (-34.24%) | 4,378.23 (-7.26%) | NM | NM | 16,154.94 (-74.87%) | 12,230.17 (-66.80%) |
| original22-v1/core/function-calls | load | 42.375 (ref) | 36.562 (+15.90%) | 46.791 (-9.44%) | 42.709 (-0.78%) | NM | NM | 43.730 (-3.10%) | 54.084 (-21.65%) |
| original22-v1/core/function-calls | engineSetup | 4.313 (ref) | 4.188 (+3.00%) | 6.104 (-29.34%) | 2.250 (+91.69%) | NM | NM | 24.125 (-82.12%) | 12.479 (-65.44%) |
| original22-v1/core/function-calls | compileValidate | 70.229 (ref) | 84.376 (-16.77%) | 60.104 (+16.85%) | 42.604 (+64.84%) | NM | NM | 37.229 (+88.64%) | 55.792 (+25.88%) |
| original22-v1/core/function-calls | instantiate | 7.354 (ref) | 9.229 (-20.31%) | 10.021 (-26.61%) | 13.791 (-46.67%) | NM | NM | 24.667 (-70.18%) | 15.916 (-53.79%) |
| original22-v1/core/function-calls | exportLookup | 0.208 (ref) | 0.375 (-44.53%) | 0.375 (-44.53%) | 0.208 (-0.24%) | NM | NM | 0.208 (+0.00%) | 0.354 (-41.24%) |
| original22-v1/core/function-calls | firstCall | 46,700.42 (ref) | 44,163.73 (+5.74%) | 42,564.90 (+9.72%) | 87,212.38 (-46.45%) | NM | NM | 62,594.73 (-25.39%) | 36,459.79 (+28.09%) |
| original22-v1/core/function-calls | cleanup | 6.979 (ref) | 1.771 (+294.10%) | 1.938 (+260.23%) | 4.062 (+71.80%) | NM | NM | 6.688 (+4.37%) | 13.000 (-46.31%) |
| original22-v1/core/memory-load-store | processStartup | 4,238.29 (ref) | 5,874.23 (-27.85%) | 6,034.44 (-29.76%) | 4,349.54 (-2.56%) | NM | NM | 16,554.06 (-74.40%) | 11,719.21 (-63.83%) |
| original22-v1/core/memory-load-store | load | 43.917 (ref) | 37.125 (+18.29%) | 45.291 (-3.03%) | 40.771 (+7.72%) | NM | NM | 43.666 (+0.57%) | 48.834 (-10.07%) |
| original22-v1/core/memory-load-store | engineSetup | 4.562 (ref) | 4.229 (+7.87%) | 6.417 (-28.89%) | 2.188 (+108.57%) | NM | NM | 25.979 (-82.44%) | 11.291 (-59.59%) |
| original22-v1/core/memory-load-store | compileValidate | 76.626 (ref) | 96.500 (-20.60%) | 73.375 (+4.43%) | 43.104 (+77.77%) | NM | NM | 40.062 (+91.27%) | 52.855 (+44.97%) |
| original22-v1/core/memory-load-store | instantiate | 10.021 (ref) | 13.229 (-24.25%) | 14.125 (-29.06%) | 15.521 (-35.44%) | NM | NM | 46.166 (-78.29%) | 34.938 (-71.32%) |
| original22-v1/core/memory-load-store | exportLookup | 0.208 (ref) | 0.334 (-37.63%) | 0.417 (-50.12%) | 0.208 (+0.00%) | NM | NM | 0.208 (-0.24%) | 0.291 (-28.52%) |
| original22-v1/core/memory-load-store | firstCall | 64,263.33 (ref) | 55,074.40 (+16.68%) | 55,281.65 (+16.25%) | 144,677.38 (-55.58%) | NM | NM | 91,413.73 (-29.70%) | 51,758.88 (+24.16%) |
| original22-v1/core/memory-load-store | cleanup | 2.250 (ref) | 2.000 (+12.50%) | 2.188 (+2.86%) | 4.208 (-46.53%) | NM | NM | 44.084 (-94.90%) | 49.770 (-95.48%) |
| original22-v1/memory/narrow-unaligned | processStartup | 4,262.12 (ref) | 6,250.94 (-31.82%) | 6,126.12 (-30.43%) | 4,258.29 (+0.09%) | NM | NM | 16,214.58 (-73.71%) | 12,149.10 (-64.92%) |
| original22-v1/memory/narrow-unaligned | load | 37.791 (ref) | 45.104 (-16.21%) | 44.416 (-14.92%) | 43.688 (-13.50%) | NM | NM | 46.188 (-18.18%) | 54.604 (-30.79%) |
| original22-v1/memory/narrow-unaligned | engineSetup | 4.396 (ref) | 4.729 (-7.05%) | 6.604 (-33.44%) | 2.208 (+99.05%) | NM | NM | 26.250 (-83.25%) | 11.229 (-60.85%) |
| original22-v1/memory/narrow-unaligned | compileValidate | 83.855 (ref) | 109.73 (-23.58%) | 81.626 (+2.73%) | 48.312 (+73.57%) | NM | NM | 41.458 (+102.27%) | 56.396 (+48.69%) |
| original22-v1/memory/narrow-unaligned | instantiate | 11.062 (ref) | 14.875 (-25.63%) | 14.562 (-24.03%) | 16.125 (-31.40%) | NM | NM | 50.416 (-78.06%) | 36.271 (-69.50%) |
| original22-v1/memory/narrow-unaligned | exportLookup | 0.250 (ref) | 0.375 (-33.33%) | 0.396 (-36.79%) | 0.208 (+20.19%) | NM | NM | 0.250 (+0.00%) | 0.271 (-7.58%) |
| original22-v1/memory/narrow-unaligned | firstCall | 84,251.17 (ref) | 73,020.60 (+15.38%) | 74,911.25 (+12.47%) | 185,795.50 (-54.65%) | NM | NM | 113,725.56 (-25.92%) | 63,691.17 (+32.28%) |
| original22-v1/memory/narrow-unaligned | cleanup | 2.146 (ref) | 2.167 (-0.95%) | 2.229 (-3.72%) | 6.375 (-66.34%) | NM | NM | 47.729 (-95.50%) | 52.605 (-95.92%) |
| original22-v1/memory/dependent-address | processStartup | 4,153.81 (ref) | 5,934.46 (-30.01%) | 5,995.15 (-30.71%) | 4,138.12 (+0.38%) | NM | NM | 15,842.29 (-73.78%) | 11,812.81 (-64.84%) |
| original22-v1/memory/dependent-address | load | 43.791 (ref) | 44.499 (-1.59%) | 43.562 (+0.53%) | 46.375 (-5.57%) | NM | NM | 45.083 (-2.86%) | 53.291 (-17.83%) |
| original22-v1/memory/dependent-address | engineSetup | 4.292 (ref) | 4.854 (-11.59%) | 6.583 (-34.81%) | 2.417 (+77.59%) | NM | NM | 25.312 (-83.05%) | 12.000 (-64.24%) |
| original22-v1/memory/dependent-address | compileValidate | 79.688 (ref) | 100.58 (-20.77%) | 75.626 (+5.37%) | 45.125 (+76.59%) | NM | NM | 41.083 (+93.97%) | 58.105 (+37.15%) |
| original22-v1/memory/dependent-address | instantiate | 10.896 (ref) | 14.479 (-24.75%) | 14.771 (-26.23%) | 15.500 (-29.70%) | NM | NM | 46.542 (-76.59%) | 38.105 (-71.40%) |
| original22-v1/memory/dependent-address | exportLookup | 0.250 (ref) | 0.354 (-29.48%) | 0.416 (-39.90%) | 0.208 (+20.19%) | NM | NM | 0.188 (+33.33%) | 0.291 (-14.24%) |
| original22-v1/memory/dependent-address | firstCall | 37,600.48 (ref) | 32,062.87 (+17.27%) | 32,315.54 (+16.35%) | 95,734.19 (-60.72%) | NM | NM | 53,851.85 (-30.18%) | 27,233.27 (+38.07%) |
| original22-v1/memory/dependent-address | cleanup | 2.250 (ref) | 2.042 (+10.19%) | 2.146 (+4.85%) | 6.479 (-65.28%) | NM | NM | 50.000 (-95.50%) | 47.980 (-95.31%) |
| original22-v1/memory/float64-unaligned | processStartup | 4,124.40 (ref) | 6,072.44 (-32.08%) | 5,835.77 (-29.33%) | 4,346.92 (-5.12%) | NM | NM | 15,830.92 (-73.95%) | 11,817.52 (-65.10%) |
| original22-v1/memory/float64-unaligned | load | 44.688 (ref) | 49.625 (-9.95%) | 41.291 (+8.22%) | 45.729 (-2.28%) | NM | NM | 50.833 (-12.09%) | 50.104 (-10.81%) |
| original22-v1/memory/float64-unaligned | engineSetup | 4.250 (ref) | 4.021 (+5.70%) | 6.750 (-37.04%) | 2.208 (+92.44%) | NM | NM | 24.021 (-82.31%) | 11.563 (-63.24%) |
| original22-v1/memory/float64-unaligned | compileValidate | 79.937 (ref) | 99.541 (-19.69%) | 75.895 (+5.33%) | 47.854 (+67.04%) | NM | NM | 40.688 (+96.47%) | 56.167 (+42.32%) |
| original22-v1/memory/float64-unaligned | instantiate | 10.479 (ref) | 13.834 (-24.25%) | 13.709 (-23.56%) | 16.291 (-35.68%) | NM | NM | 44.417 (-76.41%) | 37.750 (-72.24%) |
| original22-v1/memory/float64-unaligned | exportLookup | 0.250 (ref) | 0.292 (-14.38%) | 0.416 (-39.98%) | 0.208 (+19.90%) | NM | NM | 0.167 (+49.70%) | 0.250 (+0.00%) |
| original22-v1/memory/float64-unaligned | firstCall | 67,139.60 (ref) | 56,966.31 (+17.86%) | 57,787.23 (+16.18%) | 138,890.56 (-51.66%) | NM | NM | 84,820.31 (-20.84%) | 47,242.23 (+42.12%) |
| original22-v1/memory/float64-unaligned | cleanup | 2.333 (ref) | 2.104 (+10.91%) | 2.167 (+7.68%) | 4.375 (-46.67%) | NM | NM | 46.916 (-95.03%) | 49.770 (-95.31%) |
| original22-v1/runtime/simd-memory-full | processStartup | 4,236.06 (ref) | 6,042.04 (-29.89%) | 5,965.40 (-28.99%) | 4,354.00 (-2.71%) | NM | NM | 16,157.38 (-73.78%) | 12,101.96 (-65.00%) |
| original22-v1/runtime/simd-memory-full | load | 46.250 (ref) | 39.124 (+18.21%) | 45.396 (+1.88%) | 42.833 (+7.98%) | NM | NM | 45.354 (+1.98%) | 44.416 (+4.13%) |
| original22-v1/runtime/simd-memory-full | engineSetup | 4.271 (ref) | 4.167 (+2.50%) | 6.667 (-35.94%) | 2.479 (+72.27%) | NM | NM | 24.375 (-82.48%) | 10.771 (-60.35%) |
| original22-v1/runtime/simd-memory-full | compileValidate | 71.541 (ref) | 77.188 (-7.32%) | 66.104 (+8.23%) | 50.000 (+43.08%) | NM | NM | 40.813 (+75.29%) | 55.541 (+28.81%) |
| original22-v1/runtime/simd-memory-full | instantiate | 11.000 (ref) | 13.209 (-16.72%) | 14.083 (-21.89%) | 16.645 (-33.92%) | NM | NM | 50.834 (-78.36%) | 43.208 (-74.54%) |
| original22-v1/runtime/simd-memory-full | exportLookup | 0.208 (ref) | 0.334 (-37.72%) | 0.396 (-47.47%) | 0.230 (-9.37%) | NM | NM | 0.167 (+24.55%) | 0.312 (-33.44%) |
| original22-v1/runtime/simd-memory-full | firstCall | 73,827.10 (ref) | 64,049.40 (+15.27%) | 75,432.56 (-2.13%) | 150,879.58 (-51.07%) | NM | NM | 78,744.40 (-6.24%) | 46,487.10 (+58.81%) |
| original22-v1/runtime/simd-memory-full | cleanup | 2.208 (ref) | 2.042 (+8.15%) | 2.188 (+0.96%) | 4.292 (-48.54%) | NM | NM | 44.979 (-95.09%) | 50.875 (-95.66%) |
| original22-v1/runtime/simd-memory-widen | processStartup | 4,301.06 (ref) | 6,109.71 (-29.60%) | 5,863.19 (-26.64%) | 4,210.06 (+2.16%) | NM | NM | 16,192.10 (-73.44%) | 12,013.15 (-64.20%) |
| original22-v1/runtime/simd-memory-widen | load | 42.854 (ref) | 42.709 (+0.34%) | 38.541 (+11.19%) | 38.855 (+10.29%) | NM | NM | 44.605 (-3.92%) | 48.478 (-11.60%) |
| original22-v1/runtime/simd-memory-widen | engineSetup | 4.125 (ref) | 4.771 (-13.54%) | 6.687 (-38.31%) | 2.125 (+94.12%) | NM | NM | 25.604 (-83.89%) | 12.833 (-67.86%) |
| original22-v1/runtime/simd-memory-widen | compileValidate | 69.395 (ref) | 76.480 (-9.26%) | 63.188 (+9.82%) | 48.146 (+44.14%) | NM | NM | 40.125 (+72.95%) | 61.916 (+12.08%) |
| original22-v1/runtime/simd-memory-widen | instantiate | 10.834 (ref) | 13.999 (-22.62%) | 14.750 (-26.55%) | 16.000 (-32.29%) | NM | NM | 51.395 (-78.92%) | 44.854 (-75.85%) |
| original22-v1/runtime/simd-memory-widen | exportLookup | 0.208 (ref) | 0.396 (-47.28%) | 0.438 (-52.34%) | 0.188 (+11.20%) | NM | NM | 0.250 (-16.60%) | 0.292 (-28.60%) |
| original22-v1/runtime/simd-memory-widen | firstCall | 58,896.04 (ref) | 54,306.23 (+8.45%) | 53,668.12 (+9.74%) | 144,287.81 (-59.18%) | NM | NM | 61,040.54 (-3.51%) | 34,553.54 (+70.45%) |
| original22-v1/runtime/simd-memory-widen | cleanup | 2.583 (ref) | 1.979 (+30.55%) | 2.167 (+19.25%) | 9.084 (-71.56%) | NM | NM | 50.334 (-94.87%) | 53.479 (-95.17%) |
| original22-v1/runtime/simd-memory-lanes | processStartup | 4,274.08 (ref) | 5,988.10 (-28.62%) | 5,921.58 (-27.82%) | 4,249.60 (+0.58%) | NM | NM | 16,090.60 (-73.44%) | 11,931.90 (-64.18%) |
| original22-v1/runtime/simd-memory-lanes | load | 47.500 (ref) | 44.979 (+5.60%) | 44.833 (+5.95%) | 40.895 (+16.15%) | NM | NM | 48.125 (-1.30%) | 45.729 (+3.87%) |
| original22-v1/runtime/simd-memory-lanes | engineSetup | 4.125 (ref) | 4.146 (-0.51%) | 5.812 (-29.03%) | 2.417 (+70.67%) | NM | NM | 27.000 (-84.72%) | 10.729 (-61.55%) |
| original22-v1/runtime/simd-memory-lanes | compileValidate | 74.645 (ref) | 79.166 (-5.71%) | 66.792 (+11.76%) | 47.792 (+56.19%) | NM | NM | 43.770 (+70.54%) | 53.583 (+39.31%) |
| original22-v1/runtime/simd-memory-lanes | instantiate | 11.167 (ref) | 13.542 (-17.54%) | 13.979 (-20.12%) | 16.188 (-31.01%) | NM | NM | 54.895 (-79.66%) | 41.542 (-73.12%) |
| original22-v1/runtime/simd-memory-lanes | exportLookup | 0.250 (ref) | 0.354 (-29.48%) | 0.416 (-39.90%) | 0.167 (+49.70%) | NM | NM | 0.250 (+0.00%) | 0.291 (-14.24%) |
| original22-v1/runtime/simd-memory-lanes | firstCall | 91,713.23 (ref) | 83,803.73 (+9.44%) | 85,877.21 (+6.80%) | 187,743.29 (-51.15%) | NM | NM | 90,591.38 (+1.24%) | 55,694.40 (+64.67%) |
| original22-v1/runtime/simd-memory-lanes | cleanup | 2.250 (ref) | 2.271 (-0.90%) | 2.229 (+0.94%) | 4.479 (-49.77%) | NM | NM | 49.938 (-95.49%) | 49.437 (-95.45%) |
| original22-v1/gc/i31-roundtrip | processStartup | 4,033.62 (ref) | 5,955.48 (-32.27%) | 5,798.23 (-30.43%) | US | NM | NM | 15,991.35 (-74.78%) | 11,745.54 (-65.66%) |
| original22-v1/gc/i31-roundtrip | load | 38.688 (ref) | 39.291 (-1.54%) | 40.521 (-4.52%) | US | NM | NM | 42.417 (-8.79%) | 39.729 (-2.62%) |
| original22-v1/gc/i31-roundtrip | engineSetup | 4.250 (ref) | 4.188 (+1.49%) | 6.479 (-34.41%) | US | NM | NM | 25.562 (-83.37%) | 11.000 (-61.36%) |
| original22-v1/gc/i31-roundtrip | compileValidate | 58.375 (ref) | 53.605 (+8.90%) | 47.709 (+22.36%) | US | NM | NM | 35.479 (+64.53%) | 51.687 (+12.94%) |
| original22-v1/gc/i31-roundtrip | instantiate | 6.917 (ref) | 8.874 (-22.06%) | 9.938 (-30.39%) | US | NM | NM | 24.166 (-71.38%) | 14.896 (-53.56%) |
| original22-v1/gc/i31-roundtrip | exportLookup | 0.230 (ref) | 0.333 (-31.08%) | 0.417 (-44.96%) | US | NM | NM | 0.188 (+22.07%) | 0.271 (-15.16%) |
| original22-v1/gc/i31-roundtrip | firstCall | 341.94 (ref) | 297.02 (+15.12%) | 296.71 (+15.24%) | US | NM | NM | 457.44 (-25.25%) | 239.60 (+42.71%) |
| original22-v1/gc/i31-roundtrip | cleanup | 4.792 (ref) | 1.542 (+210.73%) | 1.542 (+210.83%) | US | NM | NM | 5.042 (-4.97%) | 8.104 (-40.87%) |
| original22-v1/gc/struct-access | processStartup | 4,081.31 (ref) | 5,856.81 (-30.32%) | 5,895.04 (-30.77%) | US | NM | NM | 15,944.29 (-74.40%) | 11,766.19 (-65.31%) |
| original22-v1/gc/struct-access | load | 41.271 (ref) | 36.209 (+13.98%) | 43.291 (-4.67%) | US | NM | NM | 47.166 (-12.50%) | 47.146 (-12.46%) |
| original22-v1/gc/struct-access | engineSetup | 4.417 (ref) | 4.042 (+9.27%) | 6.771 (-34.77%) | US | NM | NM | 25.021 (-82.35%) | 11.062 (-60.08%) |
| original22-v1/gc/struct-access | compileValidate | 63.938 (ref) | 55.938 (+14.30%) | 56.500 (+13.16%) | US | NM | NM | 43.854 (+45.80%) | 49.541 (+29.06%) |
| original22-v1/gc/struct-access | instantiate | 7.854 (ref) | 9.354 (-16.03%) | 10.333 (-23.99%) | US | NM | NM | 25.791 (-69.55%) | 14.041 (-44.06%) |
| original22-v1/gc/struct-access | exportLookup | 0.208 (ref) | 0.333 (-37.39%) | 0.354 (-41.10%) | US | NM | NM | 0.229 (-8.95%) | 0.292 (-28.60%) |
| original22-v1/gc/struct-access | firstCall | 398.00 (ref) | 380.04 (+4.73%) | 386.62 (+2.94%) | US | NM | NM | 566.25 (-29.71%) | 309.44 (+28.62%) |
| original22-v1/gc/struct-access | cleanup | 5.292 (ref) | 1.667 (+217.46%) | 1.771 (+198.81%) | US | NM | NM | 5.417 (-2.30%) | 9.083 (-41.74%) |
| original22-v1/gc/struct-allocation | processStartup | 4,137.00 (ref) | 6,057.08 (-31.70%) | 5,821.90 (-28.94%) | US | NM | NM | 16,012.52 (-74.16%) | 11,881.73 (-65.18%) |
| original22-v1/gc/struct-allocation | load | 36.062 (ref) | 42.084 (-14.31%) | 41.062 (-12.18%) | US | NM | NM | 43.312 (-16.74%) | 48.521 (-25.68%) |
| original22-v1/gc/struct-allocation | engineSetup | 3.854 (ref) | 4.188 (-7.95%) | 6.125 (-37.07%) | US | NM | NM | 25.291 (-84.76%) | 11.521 (-66.54%) |
| original22-v1/gc/struct-allocation | compileValidate | 65.125 (ref) | 58.791 (+10.77%) | 54.625 (+19.22%) | US | NM | NM | 46.250 (+40.81%) | 52.334 (+24.44%) |
| original22-v1/gc/struct-allocation | instantiate | 8.062 (ref) | 11.167 (-27.80%) | 10.896 (-26.00%) | US | NM | NM | 31.292 (-74.23%) | 18.812 (-57.14%) |
| original22-v1/gc/struct-allocation | exportLookup | 0.209 (ref) | 0.334 (-37.33%) | 0.375 (-44.27%) | US | NM | NM | 0.208 (+0.48%) | 0.230 (-8.93%) |
| original22-v1/gc/struct-allocation | firstCall | 888.31 (ref) | 987.17 (-10.01%) | 963.79 (-7.83%) | US | NM | NM | 1,203.48 (-26.19%) | 898.96 (-1.18%) |
| original22-v1/gc/struct-allocation | cleanup | 92.750 (ref) | 54.083 (+71.50%) | 54.146 (+71.30%) | US | NM | NM | 1,510.33 (-93.86%) | 1,491.90 (-93.78%) |
| original22-v1/gc/struct-indirect-access | processStartup | 4,156.50 (ref) | 6,013.54 (-30.88%) | 6,135.33 (-32.25%) | US | NM | NM | 16,109.06 (-74.20%) | 11,807.77 (-64.80%) |
| original22-v1/gc/struct-indirect-access | load | 43.041 (ref) | 44.354 (-2.96%) | 40.917 (+5.19%) | US | NM | NM | 45.730 (-5.88%) | 50.771 (-15.22%) |
| original22-v1/gc/struct-indirect-access | engineSetup | 4.375 (ref) | 4.646 (-5.82%) | 6.917 (-36.75%) | US | NM | NM | 25.292 (-82.70%) | 12.437 (-64.82%) |
| original22-v1/gc/struct-indirect-access | compileValidate | 77.270 (ref) | 75.125 (+2.86%) | 70.000 (+10.39%) | US | NM | NM | 48.312 (+59.94%) | 58.709 (+31.62%) |
| original22-v1/gc/struct-indirect-access | instantiate | 13.479 (ref) | 17.125 (-21.29%) | 17.438 (-22.70%) | US | NM | NM | 30.792 (-56.23%) | 25.375 (-46.88%) |
| original22-v1/gc/struct-indirect-access | exportLookup | 0.208 (ref) | 0.334 (-37.48%) | 0.396 (-47.28%) | US | NM | NM | 0.230 (-9.15%) | 0.291 (-28.47%) |
| original22-v1/gc/struct-indirect-access | firstCall | 738.48 (ref) | 711.17 (+3.84%) | 718.19 (+2.83%) | US | NM | NM | 976.23 (-24.35%) | 517.25 (+42.77%) |
| original22-v1/gc/struct-indirect-access | cleanup | 5.625 (ref) | 1.833 (+206.87%) | 1.771 (+217.62%) | US | NM | NM | 7.521 (-25.21%) | 10.001 (-43.75%) |
| original22-v1/threads/private-memory | processStartup | 4,175.69 (ref) | 5,898.17 (-29.20%) | 6,047.40 (-30.95%) | 4,355.31 (-4.12%) | NM | NM | 16,047.46 (-73.98%) | 12,110.98 (-65.52%) |
| original22-v1/threads/private-memory | load | 38.146 (ref) | 41.791 (-8.72%) | 42.230 (-9.67%) | 38.896 (-1.93%) | NM | NM | 41.895 (-8.95%) | 43.459 (-12.22%) |
| original22-v1/threads/private-memory | engineSetup | 4.104 (ref) | 4.438 (-7.52%) | 5.708 (-28.10%) | 2.562 (+60.16%) | NM | NM | 23.834 (-82.78%) | 10.416 (-60.60%) |
| original22-v1/threads/private-memory | compileValidate | 61.416 (ref) | 54.062 (+13.60%) | 50.855 (+20.77%) | 41.729 (+47.18%) | NM | NM | 37.437 (+64.05%) | 51.021 (+20.37%) |
| original22-v1/threads/private-memory | instantiate | 9.521 (ref) | 12.729 (-25.21%) | 13.896 (-31.48%) | 16.000 (-40.49%) | NM | NM | 44.604 (-78.65%) | 34.584 (-72.47%) |
| original22-v1/threads/private-memory | exportLookup | 0.209 (ref) | 0.333 (-37.24%) | 0.375 (-44.27%) | 0.208 (+0.48%) | NM | NM | 0.208 (+0.48%) | 0.292 (-28.42%) |
| original22-v1/threads/private-memory | firstCall | 212.71 (ref) | 190.60 (+11.60%) | 190.25 (+11.80%) | 643.06 (-66.92%) | NM | NM | 401.33 (-47.00%) | 244.33 (-12.94%) |
| original22-v1/threads/private-memory | cleanup | 2.396 (ref) | 1.770 (+35.33%) | 1.875 (+27.82%) | 4.979 (-51.88%) | NM | NM | 42.562 (-94.37%) | 41.666 (-94.25%) |
| original22-v1/threads/shared-memory | processStartup | 4,226.21 (ref) | 6,047.42 (-30.12%) | 5,998.83 (-29.55%) | 4,194.33 (+0.76%) | NM | NM | 16,209.88 (-73.93%) | 11,709.69 (-63.91%) |
| original22-v1/threads/shared-memory | load | 42.521 (ref) | 39.854 (+6.69%) | 43.396 (-2.02%) | 40.938 (+3.87%) | NM | NM | 47.208 (-9.93%) | 52.312 (-18.72%) |
| original22-v1/threads/shared-memory | engineSetup | 3.958 (ref) | 4.354 (-9.11%) | 6.312 (-37.30%) | 2.354 (+68.10%) | NM | NM | 24.020 (-83.52%) | 12.208 (-67.58%) |
| original22-v1/threads/shared-memory | compileValidate | 61.188 (ref) | 54.104 (+13.09%) | 50.416 (+21.36%) | 41.334 (+48.03%) | NM | NM | 36.834 (+66.12%) | 57.416 (+6.57%) |
| original22-v1/threads/shared-memory | instantiate | 13.833 (ref) | 14.479 (-4.46%) | 16.125 (-14.21%) | 16.020 (-13.65%) | NM | NM | 46.083 (-69.98%) | 37.563 (-63.17%) |
| original22-v1/threads/shared-memory | exportLookup | 0.250 (ref) | 0.416 (-39.98%) | 0.438 (-42.86%) | 0.208 (+20.19%) | NM | NM | 0.208 (+19.90%) | 0.250 (+0.00%) |
| original22-v1/threads/shared-memory | firstCall | 211.06 (ref) | 191.69 (+10.11%) | 193.10 (+9.30%) | 637.88 (-66.91%) | NM | NM | 395.46 (-46.63%) | 243.23 (-13.22%) |
| original22-v1/threads/shared-memory | cleanup | 11.438 (ref) | 3.979 (+187.45%) | 4.854 (+135.61%) | 5.646 (+102.58%) | NM | NM | 39.770 (-71.24%) | 39.354 (-70.94%) |
| original22-v1/threads/atomic-rmw | processStartup | 4,169.62 (ref) | 5,974.31 (-30.21%) | 5,821.15 (-28.37%) | 3,935.19 (+5.96%) | NM | NM | 16,020.52 (-73.97%) | 11,774.29 (-64.59%) |
| original22-v1/threads/atomic-rmw | load | 40.062 (ref) | 35.062 (+14.26%) | 38.667 (+3.61%) | 34.875 (+14.87%) | NM | NM | 44.188 (-9.34%) | 43.459 (-7.82%) |
| original22-v1/threads/atomic-rmw | engineSetup | 4.396 (ref) | 4.146 (+6.03%) | 5.979 (-26.48%) | 2.083 (+110.97%) | NM | NM | 23.166 (-81.03%) | 13.000 (-66.19%) |
| original22-v1/threads/atomic-rmw | compileValidate | 59.375 (ref) | 50.667 (+17.19%) | 48.875 (+21.48%) | 38.771 (+53.14%) | NM | NM | 37.188 (+59.66%) | 55.145 (+7.67%) |
| original22-v1/threads/atomic-rmw | instantiate | 13.021 (ref) | 13.958 (-6.72%) | 16.562 (-21.39%) | 15.291 (-14.85%) | NM | NM | 43.021 (-69.73%) | 36.417 (-64.25%) |
| original22-v1/threads/atomic-rmw | exportLookup | 0.250 (ref) | 0.417 (-40.05%) | 0.458 (-45.41%) | 0.188 (+33.33%) | NM | NM | 0.208 (+20.19%) | 0.312 (-20.00%) |
| original22-v1/threads/atomic-rmw | firstCall | 283.23 (ref) | 286.38 (-1.10%) | 282.56 (+0.24%) | 474.19 (-40.27%) | NM | NM | 335.31 (-15.53%) | 188.98 (+49.87%) |
| original22-v1/threads/atomic-rmw | cleanup | 7.771 (ref) | 3.896 (+99.45%) | 3.958 (+96.32%) | 3.500 (+122.01%) | NM | NM | 38.124 (-79.62%) | 40.084 (-80.61%) |
| original22-v1/threads/notify-empty | processStartup | 4,254.40 (ref) | 5,987.54 (-28.95%) | 5,900.88 (-27.90%) | 4,188.56 (+1.57%) | NM | NM | 16,149.54 (-73.66%) | 11,979.10 (-64.48%) |
| original22-v1/threads/notify-empty | load | 42.416 (ref) | 39.541 (+7.27%) | 43.958 (-3.51%) | 40.604 (+4.46%) | NM | NM | 48.709 (-12.92%) | 48.688 (-12.88%) |
| original22-v1/threads/notify-empty | engineSetup | 3.958 (ref) | 4.771 (-17.03%) | 6.458 (-38.70%) | 2.312 (+71.18%) | NM | NM | 24.855 (-84.07%) | 12.604 (-68.59%) |
| original22-v1/threads/notify-empty | compileValidate | 61.000 (ref) | 53.563 (+13.88%) | 50.521 (+20.74%) | 42.959 (+42.00%) | NM | NM | 39.271 (+55.33%) | 55.397 (+10.12%) |
| original22-v1/threads/notify-empty | instantiate | 13.604 (ref) | 15.167 (-10.30%) | 15.499 (-12.23%) | 15.812 (-13.96%) | NM | NM | 44.980 (-69.75%) | 37.875 (-64.08%) |
| original22-v1/threads/notify-empty | exportLookup | 0.250 (ref) | 0.438 (-42.86%) | 0.458 (-45.41%) | 0.208 (+20.19%) | NM | NM | 0.208 (+19.90%) | 0.292 (-14.38%) |
| original22-v1/threads/notify-empty | firstCall | 241.06 (ref) | 234.15 (+2.95%) | 227.52 (+5.95%) | 428.92 (-43.80%) | NM | NM | 417.54 (-42.27%) | 225.62 (+6.84%) |
| original22-v1/threads/notify-empty | cleanup | 11.437 (ref) | 4.188 (+173.09%) | 4.188 (+173.12%) | 3.625 (+215.50%) | NM | NM | 39.146 (-70.78%) | 42.021 (-72.78%) |
| original22-v1/threads/wait-mismatch | processStartup | 4,098.94 (ref) | 5,848.50 (-29.91%) | 5,959.79 (-31.22%) | 4,277.79 (-4.18%) | NM | NM | 15,876.58 (-74.18%) | 11,873.44 (-65.48%) |
| original22-v1/threads/wait-mismatch | load | 43.041 (ref) | 38.834 (+10.84%) | 42.188 (+2.02%) | 41.583 (+3.51%) | NM | NM | 47.229 (-8.87%) | 53.334 (-19.30%) |
| original22-v1/threads/wait-mismatch | engineSetup | 4.208 (ref) | 4.417 (-4.72%) | 6.896 (-38.98%) | 2.333 (+80.37%) | NM | NM | 26.229 (-83.96%) | 11.812 (-64.38%) |
| original22-v1/threads/wait-mismatch | compileValidate | 63.334 (ref) | 55.292 (+14.54%) | 50.709 (+24.90%) | 42.812 (+47.93%) | NM | NM | 38.875 (+62.92%) | 57.188 (+10.75%) |
| original22-v1/threads/wait-mismatch | instantiate | 14.249 (ref) | 16.083 (-11.40%) | 17.500 (-18.57%) | 15.625 (-8.80%) | NM | NM | 50.000 (-71.50%) | 37.355 (-61.85%) |
| original22-v1/threads/wait-mismatch | exportLookup | 0.250 (ref) | 0.416 (-39.98%) | 0.438 (-42.86%) | 0.229 (+9.17%) | NM | NM | 0.208 (+20.19%) | 0.292 (-14.38%) |
| original22-v1/threads/wait-mismatch | firstCall | 364.56 (ref) | 338.06 (+7.84%) | 345.06 (+5.65%) | 702.33 (-48.09%) | NM | NM | 459.50 (-20.66%) | 256.67 (+42.04%) |
| original22-v1/threads/wait-mismatch | cleanup | 12.833 (ref) | 8.958 (+43.26%) | 10.646 (+20.54%) | 7.500 (+71.11%) | NM | NM | 42.708 (-69.95%) | 42.666 (-69.92%) |
| original22-v1/threads/wait-timeout | processStartup | 4,070.12 (ref) | 5,817.87 (-30.04%) | 5,755.79 (-29.29%) | 4,039.75 (+0.75%) | NM | NM | 16,016.62 (-74.59%) | 11,880.56 (-65.74%) |
| original22-v1/threads/wait-timeout | load | 37.105 (ref) | 36.104 (+2.77%) | 42.104 (-11.87%) | 36.438 (+1.83%) | NM | NM | 49.125 (-24.47%) | 45.374 (-18.23%) |
| original22-v1/threads/wait-timeout | engineSetup | 4.250 (ref) | 4.125 (+3.03%) | 6.417 (-33.77%) | 2.042 (+108.18%) | NM | NM | 24.812 (-82.87%) | 11.812 (-64.02%) |
| original22-v1/threads/wait-timeout | compileValidate | 62.958 (ref) | 55.541 (+13.35%) | 51.916 (+21.27%) | 40.666 (+54.82%) | NM | NM | 37.895 (+66.14%) | 54.959 (+14.56%) |
| original22-v1/threads/wait-timeout | instantiate | 13.312 (ref) | 14.834 (-10.26%) | 16.730 (-20.43%) | 15.541 (-14.35%) | NM | NM | 49.041 (-72.86%) | 35.458 (-62.46%) |
| original22-v1/threads/wait-timeout | exportLookup | 0.250 (ref) | 0.416 (-39.98%) | 0.438 (-42.86%) | 0.188 (+33.33%) | NM | NM | 0.167 (+49.70%) | 0.312 (-20.00%) |
| original22-v1/threads/wait-timeout | firstCall | 345.00 (ref) | 323.40 (+6.68%) | 323.54 (+6.63%) | 3,105.60 (-88.89%) | NM | NM | 1,061.85 (-67.51%) | 821.52 (-58.00%) |
| original22-v1/threads/wait-timeout | cleanup | 8.104 (ref) | 3.792 (+113.75%) | 3.708 (+118.54%) | 3.209 (+152.56%) | NM | NM | 39.895 (-79.69%) | 40.874 (-80.17%) |
| original22-v1/applications/sha256-clang | processStartup | 4,304.50 (ref) | 5,927.25 (-27.38%) | 5,995.02 (-28.20%) | 4,324.00 (-0.45%) | NM | NM | 16,289.15 (-73.57%) | 11,854.54 (-63.69%) |
| original22-v1/applications/sha256-clang | load | 43.938 (ref) | 42.354 (+3.74%) | 50.666 (-13.28%) | 45.896 (-4.27%) | NM | NM | 51.270 (-14.30%) | 51.021 (-13.88%) |
| original22-v1/applications/sha256-clang | engineSetup | 4.771 (ref) | 4.125 (+15.65%) | 7.292 (-34.58%) | 2.458 (+94.04%) | NM | NM | 28.646 (-83.35%) | 13.375 (-64.33%) |
| original22-v1/applications/sha256-clang | compileValidate | 264.60 (ref) | 380.17 (-30.40%) | 350.85 (-24.58%) | 79.605 (+232.40%) | NM | NM | 82.479 (+220.81%) | 91.312 (+189.78%) |
| original22-v1/applications/sha256-clang | instantiate | 12.979 (ref) | 20.334 (-36.17%) | 21.000 (-38.20%) | 18.438 (-29.61%) | NM | NM | 58.354 (-77.76%) | 47.874 (-72.89%) |
| original22-v1/applications/sha256-clang | exportLookup | 0.229 (ref) | 0.375 (-38.93%) | 0.354 (-35.40%) | 0.188 (+22.13%) | NM | NM | 0.250 (-8.40%) | 0.333 (-31.23%) |
| original22-v1/applications/sha256-clang | firstCall | 93,401.92 (ref) | 79,061.23 (+18.14%) | 79,374.08 (+17.67%) | 202,872.02 (-53.96%) | NM | NM | 131,742.50 (-29.10%) | 64,831.06 (+44.07%) |
| original22-v1/applications/sha256-clang | cleanup | 3.167 (ref) | 1.958 (+61.68%) | 2.146 (+47.55%) | 10.104 (-68.66%) | NM | NM | 59.249 (-94.66%) | 58.751 (-94.61%) |
| original22-v1/applications/sha256-rust | processStartup | 4,155.96 (ref) | 6,154.71 (-32.48%) | 5,898.62 (-29.54%) | 4,446.17 (-6.53%) | NM | NM | 16,247.48 (-74.42%) | 11,855.58 (-64.95%) |
| original22-v1/applications/sha256-rust | load | 46.688 (ref) | 41.729 (+11.88%) | 45.062 (+3.61%) | 45.146 (+3.41%) | NM | NM | 46.396 (+0.63%) | 49.021 (-4.76%) |
| original22-v1/applications/sha256-rust | engineSetup | 4.562 (ref) | 3.958 (+15.26%) | 6.250 (-27.00%) | 2.312 (+97.34%) | NM | NM | 25.188 (-81.89%) | 11.812 (-61.38%) |
| original22-v1/applications/sha256-rust | compileValidate | 270.10 (ref) | 380.10 (-28.94%) | 339.92 (-20.54%) | 82.959 (+225.59%) | NM | NM | 81.062 (+233.21%) | 86.812 (+211.14%) |
| original22-v1/applications/sha256-rust | instantiate | 14.063 (ref) | 75.916 (-81.48%) | 77.583 (-81.87%) | 19.333 (-27.26%) | NM | NM | 56.645 (-75.17%) | 44.625 (-68.49%) |
| original22-v1/applications/sha256-rust | exportLookup | 0.250 (ref) | 0.396 (-36.87%) | 0.375 (-33.33%) | 0.208 (+20.19%) | NM | NM | 0.188 (+33.33%) | 0.292 (-14.38%) |
| original22-v1/applications/sha256-rust | firstCall | 91,120.25 (ref) | 77,081.92 (+18.21%) | 76,772.06 (+18.69%) | 203,699.81 (-55.27%) | NM | NM | 122,579.65 (-25.66%) | 61,796.81 (+47.45%) |
| original22-v1/applications/sha256-rust | cleanup | 2.438 (ref) | 1.938 (+25.77%) | 2.167 (+12.51%) | 6.896 (-64.65%) | NM | NM | 54.959 (-95.56%) | 61.270 (-96.02%) |
| gc-sustained-v1/gc/i31-roundtrip | processStartup | 4,130.56 (ref) | 5,938.15 (-30.44%) | 6,228.23 (-33.68%) | US | NM | NM | 15,835.92 (-73.92%) | 11,978.06 (-65.52%) |
| gc-sustained-v1/gc/i31-roundtrip | load | 40.187 (ref) | 38.895 (+3.32%) | 38.124 (+5.41%) | US | NM | NM | 44.583 (-9.86%) | 42.562 (-5.58%) |
| gc-sustained-v1/gc/i31-roundtrip | engineSetup | 4.208 (ref) | 4.146 (+1.52%) | 6.500 (-35.25%) | US | NM | NM | 25.583 (-83.55%) | 10.333 (-59.27%) |
| gc-sustained-v1/gc/i31-roundtrip | compileValidate | 60.895 (ref) | 51.729 (+17.72%) | 48.438 (+25.72%) | US | NM | NM | 35.104 (+73.47%) | 47.938 (+27.03%) |
| gc-sustained-v1/gc/i31-roundtrip | instantiate | 7.312 (ref) | 9.084 (-19.50%) | 10.187 (-28.22%) | US | NM | NM | 23.187 (-68.46%) | 14.459 (-49.42%) |
| gc-sustained-v1/gc/i31-roundtrip | exportLookup | 0.208 (ref) | 0.334 (-37.63%) | 0.396 (-47.41%) | US | NM | NM | 0.208 (+0.00%) | 0.292 (-28.77%) |
| gc-sustained-v1/gc/i31-roundtrip | firstCall | 32,813.08 (ref) | 28,995.58 (+13.17%) | 29,066.85 (+12.89%) | US | NM | NM | 44,496.90 (-26.26%) | 22,883.54 (+43.39%) |
| gc-sustained-v1/gc/i31-roundtrip | cleanup | 6.271 (ref) | 1.792 (+250.04%) | 2.104 (+197.98%) | US | NM | NM | 5.354 (+17.13%) | 10.646 (-41.10%) |
| gc-sustained-v1/gc/struct-access | processStartup | 4,024.52 (ref) | 6,017.83 (-33.12%) | 5,918.73 (-32.00%) | US | NM | NM | 15,969.56 (-74.80%) | 11,925.75 (-66.25%) |
| gc-sustained-v1/gc/struct-access | load | 36.875 (ref) | 39.646 (-6.99%) | 43.396 (-15.03%) | US | NM | NM | 43.583 (-15.39%) | 50.562 (-27.07%) |
| gc-sustained-v1/gc/struct-access | engineSetup | 4.604 (ref) | 4.333 (+6.23%) | 6.625 (-30.51%) | US | NM | NM | 25.187 (-81.72%) | 11.979 (-61.57%) |
| gc-sustained-v1/gc/struct-access | compileValidate | 64.416 (ref) | 58.167 (+10.74%) | 56.854 (+13.30%) | US | NM | NM | 41.563 (+54.99%) | 53.938 (+19.43%) |
| gc-sustained-v1/gc/struct-access | instantiate | 7.854 (ref) | 9.771 (-19.62%) | 10.666 (-26.37%) | US | NM | NM | 26.958 (-70.87%) | 17.709 (-55.65%) |
| gc-sustained-v1/gc/struct-access | exportLookup | 0.208 (ref) | 0.334 (-37.48%) | 0.396 (-47.28%) | US | NM | NM | 0.229 (-8.95%) | 0.271 (-22.92%) |
| gc-sustained-v1/gc/struct-access | firstCall | 38,000.23 (ref) | 36,446.40 (+4.26%) | 35,980.21 (+5.61%) | US | NM | NM | 54,895.35 (-30.78%) | 29,061.62 (+30.76%) |
| gc-sustained-v1/gc/struct-access | cleanup | 7.354 (ref) | 2.146 (+242.76%) | 2.125 (+246.07%) | US | NM | NM | 6.208 (+18.45%) | 12.229 (-39.86%) |
| gc-sustained-v1/gc/struct-allocation | processStartup | 4,293.60 (ref) | 6,091.19 (-29.51%) | 6,206.54 (-30.82%) | US | NM | NM | 16,240.90 (-73.56%) | 12,457.83 (-65.53%) |
| gc-sustained-v1/gc/struct-allocation | load | 50.104 (ref) | 35.771 (+40.07%) | 43.479 (+15.24%) | US | NM | NM | 49.480 (+1.26%) | 46.729 (+7.22%) |
| gc-sustained-v1/gc/struct-allocation | engineSetup | 4.333 (ref) | 4.104 (+5.59%) | 6.521 (-33.55%) | US | NM | NM | 25.375 (-82.92%) | 12.021 (-63.95%) |
| gc-sustained-v1/gc/struct-allocation | compileValidate | 66.687 (ref) | 58.000 (+14.98%) | 54.145 (+23.16%) | US | NM | NM | 45.166 (+47.65%) | 58.166 (+14.65%) |
| gc-sustained-v1/gc/struct-allocation | instantiate | 8.209 (ref) | 10.583 (-22.44%) | 10.604 (-22.59%) | US | NM | NM | 29.541 (-72.21%) | 23.979 (-65.77%) |
| gc-sustained-v1/gc/struct-allocation | exportLookup | 0.208 (ref) | 0.354 (-41.10%) | 0.417 (-50.00%) | US | NM | NM | 0.208 (+0.24%) | 0.291 (-28.47%) |
| gc-sustained-v1/gc/struct-allocation | firstCall | 49,473.19 (ref) | 33,343.67 (+48.37%) | 33,349.38 (+48.35%) | US | NM | NM | 31,728.46 (+55.93%) | 21,277.52 (+132.51%) |
| gc-sustained-v1/gc/struct-allocation | cleanup | 2,630.06 (ref) | 1,631.42 (+61.21%) | 1,663.00 (+58.15%) | US | NM | NM | 54,741.92 (-95.20%) | 54,221.79 (-95.15%) |
| gc-sustained-v1/gc/struct-indirect-access | processStartup | 4,191.75 (ref) | 6,083.44 (-31.10%) | 6,012.65 (-30.28%) | US | NM | NM | 16,097.23 (-73.96%) | 11,812.31 (-64.51%) |
| gc-sustained-v1/gc/struct-indirect-access | load | 39.104 (ref) | 44.959 (-13.02%) | 44.375 (-11.88%) | US | NM | NM | 41.750 (-6.34%) | 40.188 (-2.70%) |
| gc-sustained-v1/gc/struct-indirect-access | engineSetup | 4.271 (ref) | 4.188 (+1.98%) | 6.167 (-30.75%) | US | NM | NM | 25.020 (-82.93%) | 11.437 (-62.66%) |
| gc-sustained-v1/gc/struct-indirect-access | compileValidate | 77.750 (ref) | 73.312 (+6.05%) | 69.501 (+11.87%) | US | NM | NM | 46.312 (+67.88%) | 55.937 (+39.00%) |
| gc-sustained-v1/gc/struct-indirect-access | instantiate | 13.271 (ref) | 17.396 (-23.72%) | 16.395 (-19.06%) | US | NM | NM | 31.250 (-57.54%) | 22.833 (-41.88%) |
| gc-sustained-v1/gc/struct-indirect-access | exportLookup | 0.209 (ref) | 0.375 (-44.27%) | 0.334 (-37.43%) | US | NM | NM | 0.208 (+0.48%) | 0.230 (-8.93%) |
| gc-sustained-v1/gc/struct-indirect-access | firstCall | 72,650.42 (ref) | 67,577.23 (+7.51%) | 68,683.27 (+5.78%) | US | NM | NM | 95,201.81 (-23.69%) | 49,841.19 (+45.76%) |
| gc-sustained-v1/gc/struct-indirect-access | cleanup | 7.042 (ref) | 1.875 (+275.57%) | 2.083 (+237.99%) | US | NM | NM | 7.458 (-5.58%) | 12.188 (-42.22%) |
| threads-sustained-v2/threads/private-memory | processStartup | 4,294.60 (ref) | 6,016.94 (-28.62%) | 6,004.29 (-28.47%) | 4,214.12 (+1.91%) | NM | NM | 15,912.00 (-73.01%) | 11,707.92 (-63.32%) |
| threads-sustained-v2/threads/private-memory | load | 39.979 (ref) | 40.020 (-0.10%) | 43.855 (-8.84%) | 39.229 (+1.91%) | NM | NM | 49.792 (-19.71%) | 47.812 (-16.38%) |
| threads-sustained-v2/threads/private-memory | engineSetup | 4.125 (ref) | 4.333 (-4.79%) | 7.167 (-42.43%) | 2.396 (+72.18%) | NM | NM | 26.271 (-84.30%) | 12.126 (-65.98%) |
| threads-sustained-v2/threads/private-memory | compileValidate | 61.396 (ref) | 55.500 (+10.62%) | 52.542 (+16.85%) | 41.354 (+48.46%) | NM | NM | 38.022 (+61.48%) | 55.375 (+10.87%) |
| threads-sustained-v2/threads/private-memory | instantiate | 9.354 (ref) | 12.812 (-26.99%) | 14.188 (-34.07%) | 15.646 (-40.21%) | NM | NM | 46.749 (-79.99%) | 36.708 (-74.52%) |
| threads-sustained-v2/threads/private-memory | exportLookup | 0.250 (ref) | 0.354 (-29.38%) | 0.416 (-39.98%) | 0.167 (+49.70%) | NM | NM | 0.209 (+19.62%) | 0.230 (+8.93%) |
| threads-sustained-v2/threads/private-memory | firstCall | 2,035.10 (ref) | 1,870.40 (+8.81%) | 1,858.79 (+9.49%) | 6,319.25 (-67.80%) | NM | NM | 3,836.29 (-46.95%) | 2,239.06 (-9.11%) |
| threads-sustained-v2/threads/private-memory | cleanup | 2.083 (ref) | 1.833 (+13.64%) | 1.833 (+13.64%) | 5.167 (-59.67%) | NM | NM | 42.729 (-95.12%) | 43.604 (-95.22%) |
| threads-sustained-v2/threads/shared-memory | processStartup | 4,092.69 (ref) | 5,930.75 (-30.99%) | 5,812.85 (-29.59%) | 4,090.33 (+0.06%) | NM | NM | 16,160.00 (-74.67%) | 11,679.42 (-64.96%) |
| threads-sustained-v2/threads/shared-memory | load | 35.270 (ref) | 36.709 (-3.92%) | 38.562 (-8.54%) | 38.021 (-7.23%) | NM | NM | 44.917 (-21.48%) | 41.062 (-14.11%) |
| threads-sustained-v2/threads/shared-memory | engineSetup | 4.104 (ref) | 4.188 (-1.98%) | 6.229 (-34.10%) | 2.188 (+87.63%) | NM | NM | 25.250 (-83.74%) | 12.667 (-67.60%) |
| threads-sustained-v2/threads/shared-memory | compileValidate | 61.374 (ref) | 53.187 (+15.39%) | 49.438 (+24.15%) | 40.688 (+50.84%) | NM | NM | 37.145 (+65.23%) | 53.667 (+14.36%) |
| threads-sustained-v2/threads/shared-memory | instantiate | 13.334 (ref) | 14.646 (-8.96%) | 15.312 (-12.92%) | 15.375 (-13.28%) | NM | NM | 44.604 (-70.11%) | 35.667 (-62.62%) |
| threads-sustained-v2/threads/shared-memory | exportLookup | 0.250 (ref) | 0.396 (-36.79%) | 0.375 (-33.33%) | 0.230 (+8.93%) | NM | NM | 0.188 (+33.33%) | 0.291 (-14.09%) |
| threads-sustained-v2/threads/shared-memory | firstCall | 2,043.06 (ref) | 1,840.85 (+10.98%) | 1,838.65 (+11.12%) | 6,358.62 (-67.87%) | NM | NM | 3,895.65 (-47.56%) | 2,247.29 (-9.09%) |
| threads-sustained-v2/threads/shared-memory | cleanup | 9.062 (ref) | 4.021 (+125.38%) | 4.334 (+109.10%) | 4.000 (+126.56%) | NM | NM | 38.416 (-76.41%) | 42.271 (-78.56%) |
| threads-sustained-v2/threads/atomic-rmw | processStartup | 4,230.44 (ref) | 6,151.25 (-31.23%) | 5,891.60 (-28.20%) | 4,259.38 (-0.68%) | NM | NM | 16,074.98 (-73.68%) | 11,717.92 (-63.90%) |
| threads-sustained-v2/threads/atomic-rmw | load | 39.687 (ref) | 38.916 (+1.98%) | 42.166 (-5.88%) | 39.270 (+1.06%) | NM | NM | 43.916 (-9.63%) | 42.959 (-7.62%) |
| threads-sustained-v2/threads/atomic-rmw | engineSetup | 4.396 (ref) | 4.292 (+2.43%) | 6.021 (-26.98%) | 2.396 (+83.49%) | NM | NM | 29.562 (-85.13%) | 10.874 (-59.57%) |
| threads-sustained-v2/threads/atomic-rmw | compileValidate | 60.272 (ref) | 52.334 (+15.17%) | 48.500 (+24.27%) | 43.146 (+39.69%) | NM | NM | 38.792 (+55.37%) | 51.708 (+16.56%) |
| threads-sustained-v2/threads/atomic-rmw | instantiate | 13.001 (ref) | 14.709 (-11.61%) | 16.250 (-20.00%) | 15.771 (-17.57%) | NM | NM | 52.084 (-75.04%) | 34.896 (-62.75%) |
| threads-sustained-v2/threads/atomic-rmw | exportLookup | 0.250 (ref) | 0.417 (-40.05%) | 0.375 (-33.33%) | 0.167 (+49.70%) | NM | NM | 0.208 (+19.90%) | 0.250 (+0.00%) |
| threads-sustained-v2/threads/atomic-rmw | firstCall | 2,695.02 (ref) | 2,729.44 (-1.26%) | 2,673.69 (+0.80%) | 4,637.75 (-41.89%) | NM | NM | 3,291.44 (-18.12%) | 1,679.44 (+60.47%) |
| threads-sustained-v2/threads/atomic-rmw | cleanup | 9.812 (ref) | 3.958 (+147.88%) | 4.729 (+107.47%) | 4.271 (+129.77%) | NM | NM | 41.209 (-76.19%) | 41.001 (-76.07%) |
| threads-sustained-v2/threads/notify-empty | processStartup | 4,212.17 (ref) | 5,990.92 (-29.69%) | 5,908.56 (-28.71%) | 4,281.02 (-1.61%) | NM | NM | 16,090.21 (-73.82%) | 12,213.92 (-65.51%) |
| threads-sustained-v2/threads/notify-empty | load | 43.812 (ref) | 40.167 (+9.08%) | 44.105 (-0.66%) | 37.479 (+16.90%) | NM | NM | 42.188 (+3.85%) | 52.334 (-16.28%) |
| threads-sustained-v2/threads/notify-empty | engineSetup | 4.688 (ref) | 3.896 (+20.32%) | 6.104 (-23.21%) | 2.208 (+112.30%) | NM | NM | 24.250 (-80.67%) | 12.562 (-62.69%) |
| threads-sustained-v2/threads/notify-empty | compileValidate | 61.480 (ref) | 51.312 (+19.81%) | 49.875 (+23.27%) | 41.500 (+48.14%) | NM | NM | 38.333 (+60.38%) | 55.124 (+11.53%) |
| threads-sustained-v2/threads/notify-empty | instantiate | 14.188 (ref) | 14.896 (-4.76%) | 16.584 (-14.45%) | 15.500 (-8.47%) | NM | NM | 44.771 (-68.31%) | 37.605 (-62.27%) |
| threads-sustained-v2/threads/notify-empty | exportLookup | 0.250 (ref) | 0.416 (-39.98%) | 0.458 (-45.41%) | 0.209 (+19.62%) | NM | NM | 0.250 (+0.00%) | 0.334 (-25.04%) |
| threads-sustained-v2/threads/notify-empty | firstCall | 2,311.94 (ref) | 2,213.56 (+4.44%) | 2,196.67 (+5.25%) | 4,227.81 (-45.32%) | NM | NM | 4,009.23 (-42.33%) | 2,068.75 (+11.76%) |
| threads-sustained-v2/threads/notify-empty | cleanup | 13.188 (ref) | 4.646 (+183.85%) | 4.625 (+185.14%) | 5.979 (+120.58%) | NM | NM | 45.375 (-70.94%) | 45.855 (-71.24%) |
| threads-sustained-v2/threads/wait-mismatch | processStartup | 4,072.35 (ref) | 5,949.46 (-31.55%) | 5,824.33 (-30.08%) | 4,265.54 (-4.53%) | NM | NM | 16,298.40 (-75.01%) | 11,986.73 (-66.03%) |
| threads-sustained-v2/threads/wait-mismatch | load | 36.291 (ref) | 36.249 (+0.12%) | 42.563 (-14.73%) | 39.854 (-8.94%) | NM | NM | 41.021 (-11.53%) | 42.937 (-15.48%) |
| threads-sustained-v2/threads/wait-mismatch | engineSetup | 4.104 (ref) | 4.354 (-5.73%) | 6.958 (-41.01%) | 2.313 (+77.45%) | NM | NM | 23.729 (-82.70%) | 11.854 (-65.37%) |
| threads-sustained-v2/threads/wait-mismatch | compileValidate | 60.812 (ref) | 55.313 (+9.94%) | 51.667 (+17.70%) | 40.459 (+50.31%) | NM | NM | 38.146 (+59.42%) | 54.522 (+11.54%) |
| threads-sustained-v2/threads/wait-mismatch | instantiate | 13.438 (ref) | 14.854 (-9.54%) | 16.209 (-17.09%) | 15.125 (-11.15%) | NM | NM | 44.355 (-69.70%) | 37.251 (-63.93%) |
| threads-sustained-v2/threads/wait-mismatch | exportLookup | 0.230 (ref) | 0.396 (-41.97%) | 0.417 (-44.96%) | 0.188 (+22.40%) | NM | NM | 0.208 (+10.34%) | 0.312 (-26.56%) |
| threads-sustained-v2/threads/wait-mismatch | firstCall | 3,631.10 (ref) | 3,286.29 (+10.49%) | 3,316.08 (+9.50%) | 6,989.65 (-48.05%) | NM | NM | 4,370.38 (-16.92%) | 2,414.08 (+50.41%) |
| threads-sustained-v2/threads/wait-mismatch | cleanup | 10.125 (ref) | 4.271 (+137.09%) | 4.854 (+108.59%) | 3.917 (+158.49%) | NM | NM | 42.395 (-76.12%) | 46.312 (-78.14%) |
| threads-sustained-v2/threads/wait-timeout | processStartup | 4,110.92 (ref) | 6,111.50 (-32.73%) | 5,914.54 (-30.49%) | 4,166.46 (-1.33%) | NM | NM | 16,219.50 (-74.65%) | 11,954.21 (-65.61%) |
| threads-sustained-v2/threads/wait-timeout | load | 37.271 (ref) | 36.896 (+1.02%) | 43.001 (-13.32%) | 39.271 (-5.09%) | NM | NM | 43.105 (-13.53%) | 45.229 (-17.59%) |
| threads-sustained-v2/threads/wait-timeout | engineSetup | 4.333 (ref) | 3.896 (+11.22%) | 6.750 (-35.81%) | 2.062 (+110.08%) | NM | NM | 25.084 (-82.73%) | 10.959 (-60.46%) |
| threads-sustained-v2/threads/wait-timeout | compileValidate | 61.084 (ref) | 55.791 (+9.49%) | 52.751 (+15.80%) | 41.396 (+47.56%) | NM | NM | 38.104 (+60.31%) | 52.916 (+15.43%) |
| threads-sustained-v2/threads/wait-timeout | instantiate | 12.291 (ref) | 14.042 (-12.47%) | 15.854 (-22.47%) | 15.354 (-19.95%) | NM | NM | 46.124 (-73.35%) | 34.730 (-64.61%) |
| threads-sustained-v2/threads/wait-timeout | exportLookup | 0.250 (ref) | 0.416 (-39.90%) | 0.438 (-42.86%) | 0.167 (+49.70%) | NM | NM | 0.230 (+8.93%) | 0.312 (-20.00%) |
| threads-sustained-v2/threads/wait-timeout | firstCall | 3,368.81 (ref) | 3,115.13 (+8.14%) | 3,081.88 (+9.31%) | 32,066.94 (-89.49%) | NM | NM | 9,731.56 (-65.38%) | 7,559.33 (-55.44%) |
| threads-sustained-v2/threads/wait-timeout | cleanup | 9.021 (ref) | 5.458 (+65.27%) | 4.604 (+95.92%) | 3.626 (+148.82%) | NM | NM | 45.000 (-79.95%) | 48.188 (-81.28%) |
| generic-calls-v1/core/generic-calls | processStartup | 4,203.23 (ref) | 5,944.19 (-29.29%) | 5,971.35 (-29.61%) | 4,319.79 (-2.70%) | NM | NM | 16,232.94 (-74.11%) | 11,964.38 (-64.87%) |
| generic-calls-v1/core/generic-calls | load | 43.458 (ref) | 39.271 (+10.66%) | 40.771 (+6.59%) | 41.395 (+4.98%) | NM | NM | 44.666 (-2.71%) | 46.021 (-5.57%) |
| generic-calls-v1/core/generic-calls | engineSetup | 4.583 (ref) | 3.917 (+17.03%) | 5.812 (-21.14%) | 2.271 (+101.83%) | NM | NM | 26.000 (-82.37%) | 11.375 (-59.71%) |
| generic-calls-v1/core/generic-calls | compileValidate | 68.395 (ref) | 67.604 (+1.17%) | 62.728 (+9.03%) | 36.958 (+85.06%) | NM | NM | 42.500 (+60.93%) | 51.959 (+31.63%) |
| generic-calls-v1/core/generic-calls | instantiate | 7.875 (ref) | 11.000 (-28.41%) | 11.229 (-29.87%) | 14.479 (-45.61%) | NM | NM | 27.521 (-71.39%) | 15.834 (-50.26%) |
| generic-calls-v1/core/generic-calls | exportLookup | 0.229 (ref) | 0.354 (-35.40%) | 0.375 (-38.93%) | 0.208 (+10.10%) | NM | NM | 0.208 (+10.10%) | 0.312 (-26.72%) |
| generic-calls-v1/core/generic-calls | firstCall | 481,028.85 (ref) | 428,435.56 (+12.28%) | 434,896.60 (+10.61%) | 926,694.21 (-48.09%) | NM | NM | 589,806.04 (-18.44%) | 313,708.10 (+53.34%) |
| generic-calls-v1/core/generic-calls | cleanup | 6.938 (ref) | 1.896 (+265.90%) | 1.896 (+265.90%) | 4.771 (+45.43%) | NM | NM | 6.083 (+14.04%) | 12.562 (-44.77%) |

</details>

<details>
<summary>Preview 3 lifecycle phase observations</summary>

| Payload | Phase | wasmvm Go | wasmvm LLGo (pinned) | wasmvm LLGo 42efacae | wazero 1.12.0 | Wasmtime 48.0.1 | Wasmer 7.3.0 | WasmEdge 0.17.1 | WasmEdge 0.18.1 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 16 B | cleanup | 60.396 (ref) | 514.42 (-88.26%) | 528.81 (-88.58%) | NM | 242.12 (-75.06%) | NM | NM | NM |
| 16 B | compileValidate | 57,338.56 (ref) | 87,039.38 (-34.12%) | 87,034.02 (-34.12%) | NM | 52,775.23 (+8.65%) | NM | NM | NM |
| 16 B | engineSetup | 0.417 (ref) | 0.167 (+149.70%) | 0.208 (+100.00%) | NM | 154.04 (-99.73%) | NM | NM | NM |
| 16 B | executorSetup | NM | NM | NM | NM | 41.855 (NC) | NM | NM | NM |
| 16 B | exportLookup | 0.730 (ref) | 1.230 (-40.67%) | 1.021 (-28.55%) | NM | 6.334 (-88.48%) | NM | NM | NM |
| 16 B | firstCall | 8,717.04 (ref) | 15,752.73 (-44.66%) | 15,751.56 (-44.66%) | NM | 8,761.02 (-0.50%) | NM | NM | NM |
| 16 B | instantiate | 515.19 (ref) | 785.37 (-34.40%) | 766.81 (-32.81%) | NM | 195.44 (+163.61%) | NM | NM | NM |
| 16 B | load | 63.978 (ref) | 65.209 (-1.89%) | 64.791 (-1.25%) | NM | 62.875 (+1.76%) | NM | NM | NM |
| 16 B | processStartup | 6,461.79 (ref) | 12,637.77 (-48.87%) | 12,771.67 (-49.41%) | NM | 7,044.06 (-8.27%) | NM | NM | NM |
| 16 B | providerSetup | 49.895 (ref) | 216.92 (-77.00%) | 187.52 (-73.39%) | NM | 236.23 (-78.88%) | NM | NM | NM |
| 16 B | requestPreparation | 0.771 (ref) | 0.875 (-11.89%) | 0.730 (+5.69%) | NM | NM | NM | NM | NM |
| 65536 B | cleanup | 66.312 (ref) | 544.23 (-87.82%) | 510.15 (-87.00%) | NM | 266.33 (-75.10%) | NM | NM | NM |
| 65536 B | compileValidate | 57,905.81 (ref) | 87,452.69 (-33.79%) | 87,792.06 (-34.04%) | NM | 53,250.10 (+8.74%) | NM | NM | NM |
| 65536 B | engineSetup | 0.416 (ref) | 0.167 (+149.85%) | 0.230 (+81.26%) | NM | 156.77 (-99.73%) | NM | NM | NM |
| 65536 B | executorSetup | NM | NM | NM | NM | 48.375 (NC) | NM | NM | NM |
| 65536 B | exportLookup | 0.708 (ref) | 1.000 (-29.20%) | 1.042 (-32.05%) | NM | 6.250 (-88.67%) | NM | NM | NM |
| 65536 B | firstCall | 13,140.96 (ref) | 18,634.67 (-29.48%) | 17,579.75 (-25.25%) | NM | 9,289.42 (+41.46%) | NM | NM | NM |
| 65536 B | instantiate | 496.21 (ref) | 751.58 (-33.98%) | 750.69 (-33.90%) | NM | 199.02 (+149.32%) | NM | NM | NM |
| 65536 B | load | 53.000 (ref) | 64.229 (-17.48%) | 64.459 (-17.78%) | NM | 56.124 (-5.57%) | NM | NM | NM |
| 65536 B | processStartup | 6,491.90 (ref) | 13,140.25 (-50.60%) | 12,829.58 (-49.40%) | NM | 7,046.04 (-7.86%) | NM | NM | NM |
| 65536 B | providerSetup | 50.105 (ref) | 203.35 (-75.36%) | 194.52 (-74.24%) | NM | 248.52 (-79.84%) | NM | NM | NM |
| 65536 B | requestPreparation | 7.708 (ref) | 9.166 (-15.91%) | 7.250 (+6.32%) | NM | NM | NM | NM | NM |

</details>
