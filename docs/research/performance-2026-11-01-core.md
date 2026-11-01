# Complete Core warm-call table: 2026-11-01

See the [audit](performance-2026-11-01.md) for units, policies, identities, phases and limitations. Values are medians of three fresh process medians, each with seven warm samples. These descriptive results do not establish calibrated parity.

| Workload | Unit | wasmvm-go-interpreter | wasmvm-go-native | wasmvm-llgo-interpreter | wasmvm-llgo-native | wazero-interpreter | wazero-compiler | wasmtime-cranelift | wasmer-cranelift | wasmedge-interpreter |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| core/call-control | ns/call | 2167 | 3292 | 2584 | 4166 | 792.00 | 459.00 | 1000 | 2000 | 2000 |
| core/integer-loop | ns/source iteration | 65.99 | 1.667 | 66.44 | 1.584 | 81.97 | 1.748 | 1.556 | 1.584 | 48.90 |
| core/integer-sustained | ns/source iteration | 66.25 | 1.545 | 65.57 | 1.551 | 82.47 | 1.589 | 1.534 | 1.549 | 48.76 |
| core/function-calls | ns/source iteration | 88.92 | 0.949 | 136.94 | 0.852 | 83.28 | 1.499 | 0.927 | 0.820 | 62.15 |
| core/memory-load-store | ns/source iteration | 130.71 | 0.731 | 195.78 | 0.627 | 147.97 | 0.968 | 0.513 | 0.587 | 92.61 |
| core/memory-sustained | ns/source iteration | 126.10 | 0.512 | 190.87 | 0.507 | 143.09 | 0.831 | 0.470 | 0.520 | 90.50 |
| float/arithmetic-f32 | ns/source iteration | 65.53 | 3.706 | 63.08 | 3.645 | 70.75 | 3.875 | 3.673 | 3.645 | 40.14 |
| float/arithmetic-f64 | ns/source iteration | 64.92 | 4.193 | 63.99 | 4.176 | 71.35 | 4.388 | 4.115 | 4.117 | 40.17 |
| float/minmax-f32 | ns/source iteration | 97.79 | 1.856 | 95.12 | 1.833 | 77.43 | 1.970 | 1.818 | 1.819 | 55.40 |
| float/minmax-f64 | ns/source iteration | 97.10 | 1.841 | 95.70 | 1.805 | 86.17 | 1.967 | 1.822 | 1.819 | 55.66 |
| float/rounding-f32 | ns/source iteration | 160.44 | 6.774 | 188.59 | 6.728 | 156.50 | 7.040 | 6.685 | 6.691 | 85.74 |
| float/rounding-f64 | ns/source iteration | 162.14 | 8.038 | 184.69 | 8.029 | 164.55 | 8.264 | 7.919 | 7.950 | 86.96 |
| float/convert-signed-f32 | ns/source iteration | 146.54 | 9.691 | 125.48 | 9.661 | 145.52 | 20.23 | 9.680 | 9.661 | 92.45 |
| float/convert-signed-f64 | ns/source iteration | 161.56 | 9.723 | 134.09 | 9.723 | 164.94 | 20.43 | 9.718 | 9.680 | 100.01 |
| float/convert-unsigned-f32 | ns/source iteration | 155.79 | 9.661 | 133.05 | 9.589 | 159.22 | 20.13 | 9.623 | 9.640 | 98.37 |
| float/convert-unsigned-f64 | ns/source iteration | 174.38 | 9.764 | 147.12 | 9.744 | 186.61 | 20.47 | 9.666 | 9.677 | 109.35 |
| global/i32 | ns/source iteration | 44.94 | 1.058 | 42.54 | 1.075 | 47.45 | 0.600 | 1.200 | 1.300 | 30.20 |
| global/i64 | ns/source iteration | 46.32 | 0.929 | 39.27 | 0.883 | 53.23 | 0.650 | 0.900 | 1.200 | 30.10 |
| global/f32 | ns/source iteration | 46.41 | 1.546 | 41.93 | 1.325 | 46.43 | 0.733 | 2.700 | 3.000 | 30.60 |
| global/f64 | ns/source iteration | 46.55 | 1.508 | 42.09 | 1.275 | 52.70 | 0.850 | 2.800 | 2.900 | 30.60 |
| global/v128 | ns/source iteration | 51.69 | 1.183 | 49.92 | 1.154 | 68.28 | 0.617 | 2.500 | 2.700 | 30.30 |
| memory/narrow-unaligned | ns/source iteration | 159.46 | 0.710 | 267.46 | 0.697 | 182.57 | 1.051 | 0.692 | 0.711 | 113.16 |
| memory/dependent-address | ns/source iteration | 76.11 | 1.861 | 118.56 | 1.822 | 99.24 | 3.102 | 2.035 | 2.083 | 53.14 |
| memory/float64-unaligned | ns/source iteration | 122.83 | 0.737 | 175.62 | 0.721 | 135.86 | 0.881 | 0.707 | 0.712 | 84.62 |
| memory/growth-first-call | ns/call | 1959 | 5166 | 3084 | 6500 | 1125 | 541.00 | 2000 | 3000 | 2000 |
| runtime/indirect-calls | ns/source iteration | 93.24 | 1.719 | 160.80 | 1.642 | 94.22 | 2.194 | 1.400 | 1.533 | 81.80 |
| runtime/simd-i32x4 | ns/source iteration | 104.36 | 1.996 | 107.53 | 2.058 | 124.21 | 1.879 | 2.000 | 2.200 | 43.40 |
| runtime/bulk-overlap | ns/source iteration | 2394 | 2597 | 2644 | 3388 | 261.00 | 184.58 | 155.00 | 159.00 | 230.00 |
| runtime/globals-tables | ns/source iteration | 93.13 | 1.583 | 78.88 | 1.696 | 111.99 | 29.09 | 20.90 | 5.600 | 68.20 |
| runtime/simd-memory-full | ns/source iteration | 129.10 | 4.729 | 228.91 | 4.412 | 151.50 | 3.771 | 4.000 | 4.300 | 78.70 |
| runtime/simd-memory-widen | ns/source iteration | 112.28 | 2.108 | 211.54 | 1.971 | 143.66 | 1.233 | 1.300 | 1.600 | 61.80 |
| runtime/simd-memory-lanes | ns/source iteration | 160.37 | 6.617 | 303.60 | 5.992 | 188.95 | 6.021 | 6.200 | 6.400 | 91.10 |
| table/table-roundtrip | ns/source iteration | 134.58 | 1.440 | 128.53 | 1.305 | 133.91 | 14.79 | 10.61 | 5.060 | 97.48 |
| table/table-copy-handles | ns/source iteration | 104.23 | 1.315 | 101.30 | 1.198 | 136.77 | 1.588 | 1.160 | 5.570 | 81.64 |
| table/ref-func | ns/source iteration | 104.00 | 1.323 | 100.01 | 1.186 | 123.85 | 32.30 | 20.42 | 5.260 | 79.82 |
| vector-float/arithmetic-f32x4 | ns/source iteration | 128.81 | 4.650 | 113.08 | 4.463 | 104.92 | 3.804 | 3.800 | 4.000 | 43.00 |
| vector-float/arithmetic-f64x2 | ns/source iteration | 117.84 | 5.237 | 100.26 | 5.246 | 110.56 | 4.392 | 4.500 | 4.800 | 42.70 |
| vector-float/minmax-f32x4 | ns/source iteration | 291.09 | 5.996 | 234.64 | 5.783 | 210.89 | 5.829 | 5.400 | 5.500 | 77.60 |
| vector-float/minmax-f64x2 | ns/source iteration | 250.68 | 6.121 | 211.07 | 6.258 | 212.26 | 5.167 | 5.200 | 5.700 | 77.50 |
| vector-float/rounding-f32x4 | ns/source iteration | 421.06 | 9.829 | 577.39 | 8.504 | 269.32 | 9.062 | 7.800 | 8.000 | 91.90 |
| vector-float/rounding-f64x2 | ns/source iteration | 353.10 | 10.91 | 407.82 | 10.03 | 298.79 | 10.56 | 9.100 | 9.600 | 91.80 |
| vector-float/convert-signed-f32x4 | ns/source iteration | 201.54 | 5.517 | 169.57 | 4.550 | 175.21 | 4.646 | 4.500 | 4.800 | 56.70 |
| vector-float/convert-signed-f64x2 | ns/source iteration | 179.69 | 6.287 | 153.46 | 5.929 | 170.73 | 5.588 | 5.800 | 6.100 | 55.70 |
| vector-float/convert-unsigned-f32x4 | ns/source iteration | 203.03 | 4.688 | 169.66 | 4.679 | 167.88 | 4.487 | 4.400 | 4.700 | 54.60 |
| vector-float/convert-unsigned-f64x2 | ns/source iteration | 179.49 | 6.013 | 156.30 | 5.763 | 169.75 | 5.583 | 5.700 | 6.000 | 55.50 |
| vector-float/casts-f32x4 | ns/source iteration | 154.32 | 4.471 | 128.70 | 4.650 | 139.44 | 4.154 | 4.100 | 4.400 | 46.70 |
| vector-integer/shuffle | ns/source iteration | 115.77 | 2.312 | 110.47 | 2.346 | 104.05 | 1.654 | 1.800 | 1.900 | 49.30 |
| vector-integer/swizzle | ns/source iteration | 164.40 | 3.704 | 146.83 | 3.600 | 189.70 | 2.858 | 2.900 | 3.000 | 58.40 |
| vector-integer/popcount | ns/source iteration | 90.19 | 2.413 | 87.76 | 2.321 | 126.19 | 1.700 | 1.900 | 2.100 | 41.50 |
| vector-integer/narrow | ns/source iteration | 202.75 | 8.625 | 215.96 | 8.496 | 150.99 | 7.658 | 7.600 | 7.800 | 61.40 |
| vector-integer/average | ns/source iteration | 154.59 | 2.829 | 161.83 | 2.796 | 133.78 | 2.200 | 2.400 | 2.500 | 50.00 |
| vector-integer/q15 | ns/source iteration | 113.22 | 2.467 | 120.12 | 2.608 | 120.15 | 1.887 | 2.000 | 2.400 | 44.00 |
| vector-integer/dot | ns/source iteration | 105.59 | 3.096 | 100.49 | 3.133 | 123.39 | 2.575 | 2.500 | 2.800 | 44.10 |
| vector-integer/pairwise | ns/source iteration | 149.69 | 4.100 | 155.26 | 4.108 | 126.70 | 5.558 | 3.500 | 3.700 | 49.00 |
| vector-integer/extend-i16 | ns/source iteration | 173.29 | 4.833 | 177.57 | 4.071 | 131.27 | 3.217 | 3.400 | 3.700 | 48.10 |
| vector-integer/extend-i32 | ns/source iteration | 153.91 | 4.279 | 148.57 | 3.867 | 165.89 | 3.192 | 3.400 | 3.600 | 47.70 |
| vector-integer/extend-i64 | ns/source iteration | 142.14 | 4.158 | 133.11 | 3.812 | 168.28 | 3.429 | 3.300 | 3.700 | 48.50 |
| vector-integer/extmul-i16 | ns/source iteration | 198.17 | 5.175 | 204.85 | 4.900 | 152.75 | 6.521 | 4.400 | 4.600 | 61.40 |
| vector-integer/extmul-i32 | ns/source iteration | 191.83 | 5.254 | 181.96 | 4.963 | 191.16 | 6.158 | 4.400 | 4.700 | 60.50 |
| vector-integer/extmul-i64 | ns/source iteration | 172.01 | 5.050 | 165.02 | 4.904 | 189.46 | 13.55 | 4.400 | 4.700 | 61.50 |
| vector-integer/shift-i8 | ns/source iteration | 226.78 | 4.246 | 235.89 | 3.692 | 189.82 | 3.038 | 3.000 | 3.300 | 72.80 |
| vector-integer/reduce-i8 | ns/source iteration | 186.49 | 9.808 | 167.48 | 9.492 | 214.10 | 8.521 | 8.000 | 8.500 | 71.20 |
| vector-integer/shift-i16 | ns/source iteration | 186.02 | 3.679 | 190.35 | 3.558 | 188.46 | 2.896 | 2.800 | 3.100 | 72.90 |
| vector-integer/reduce-i16 | ns/source iteration | 172.10 | 8.079 | 159.85 | 8.100 | 200.70 | 7.079 | 7.000 | 7.500 | 70.30 |
| vector-integer/shift-i32 | ns/source iteration | 176.34 | 4.000 | 171.66 | 3.837 | 187.55 | 2.758 | 2.900 | 3.300 | 75.30 |
| vector-integer/reduce-i32 | ns/source iteration | 170.85 | 8.183 | 158.16 | 8.162 | 200.70 | 6.821 | 7.300 | 7.400 | 71.80 |
| vector-integer/shift-i64 | ns/source iteration | 161.77 | 3.638 | 154.28 | 3.479 | 186.44 | 2.983 | 2.900 | 3.100 | 72.70 |
| vector-integer/reduce-i64 | ns/source iteration | 167.08 | 8.408 | 154.32 | 8.033 | 203.15 | 6.763 | 7.100 | 7.400 | 71.60 |
