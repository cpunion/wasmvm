# Cumulative native compiler integration: 2026-11-07

[PR #122](https://github.com/cpunion/wasmvm/pull/122) integrates the shared native
slot-frame/continuation contracts, CFG value analysis and costed allocation,
scalar forwarding and guarded x86 memory operands. Main `9b7cff0` has the exact
source tree tested as `5a8de4c`. The preceding WASI P3 socket-admission correction
and CI shard changes merged through [PR #121](https://github.com/cpunion/wasmvm/pull/121).
Controlled performance, expanded WASI world acceptance and release gates remain open.

## Final semantic evidence and CI cost

[Exact-head CI 42368186102](https://github.com/cpunion/wasmvm/actions/runs/42368186102)
passes all 21 required jobs. The artifact audit checks **240 conformance reports**
over Linux amd64/macOS arm64 and Go/LLGo. All 31 packages occur exactly once in
seven shards on each LLGo target. All 30 executable package coverage gates exceed
90%; the minimum is **90.75%**. The longest job takes **526 seconds (8m46s)**.
The merge source tree equals the tested head. These are executed semantic results,
not a claim of full standards or security acceptance.

Independent routing groups now execute concurrently using the same build bundle.
Large historical JSON records are losslessly compressed with checked original
hashes. Neither CI optimization removes a package, sample or unsupported outcome.

## Cumulative Linux comparison

[Event 42368258530](https://github.com/cpunion/wasmvm/actions/runs/42368258530)
compares `5a8de4c` with preceding main `fdf6a2e`. Go 1.27.3 and LLGo `3571c25`
retain the same `-O3 -lto=full` profile. References remain wazero 1.12.0,
Wasmtime 48.0.1, Wasmer 7.3.0 and WasmEdge 0.17.1. The event completes all six
jobs, all 22 original-count cases, 362 alternate checksum checks and 12 expected
reference-GC rejections. Eight cases resolve all four A/A controls within ±2.5%.
Build takes 325 seconds; the longest measurement group takes 383 seconds.

Ratios use complete warm API intervals; lower is faster. Each number summarizes
five process medians of seven serially interleaved calls. Preparation, publication,
first call, warm calls, shutdown, process readiness and process wall intervals
stay separate. Hosted CPU power and neighbors are uncontrolled: these are
advisory observations, with no phase subtraction or cross-event pooling.

| Workload | LLGo head/base | / Go native | / wazero compiler | / Wasmtime | / native C | All four A/A |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Function calls | 0.803 | 1.000 | 0.581 | 0.821 | 0.937 | Resolved |
| Memory load/store | 0.804 | 1.000 | 0.645 | 1.177 | 1.465 | Resolved |
| Narrow unaligned | 0.909 | 0.999 | 0.875 | 1.264 | 2.229 | Resolved |
| Float64 unaligned | 0.908 | 1.000 | 0.626 | 1.038 | 1.053 | Resolved |
| SIMD widen | 0.823 | 1.000 | 0.628 | 1.392 | 0.366 | Resolved |
| Clang SHA | 0.209 | 0.997 | 0.655 | 1.248 | 1.498 | Resolved |
| Rust SHA | 0.222 | 1.007 | 0.669 | 1.327 | 1.569 | Resolved |
| GC struct access | 1.105 | 1.026 | Unsupported | 2.107 | Unavailable | Inconclusive |
| GC struct allocation | 1.022 | 1.283 | Unsupported | 0.308 | Unavailable | Inconclusive |
| GC indirect access | 1.033 | 1.008 | Unsupported | 2.950 | Unavailable | Inconclusive |
| Atomic RMW | 1.032 | 1.110 | 1.403 | 1.352 | Unavailable | Inconclusive |
| Tiny call control | 1.293 | 1.456 | 8.638 | 4.323 | Unavailable | Inconclusive |

The complete record includes all 22 cases, interpreter ratios, all competitors
and every lifecycle phase. Adverse GC, atomic and tiny-control observations stay
visible. Unresolved controls establish neither a supported speedup nor a compiler
bug. LLGo #2983 remains with its existing contributor. The compiler pin is unchanged
after upstream merged LLGo #2991.

## Rejected frame-immediate experiment

Source `2f2c8ed`, measured separately against main `9b7cff0` in
[event 42370236791](https://github.com/cpunion/wasmvm/actions/runs/42370236791),
reused the existing frame and immediate encoders. It passed local Go/LLGo
regressions, but changed only the two SHA wrappers; their hot loops and all
88 actual ARM64 publications stayed unchanged. Clang/Rust SHA LLGo head/base
ratios were **1.0011/1.0015**, with all four controls resolved. No meaningful
benefit was established. The experiment was rejected, its source bundle and
complete evidence retained, and its branch removed. Actual x86 benchmark modules
executed, but the full x86 cold-frame kernel matrix was not run or claimed.
The earlier rejected FP-register strategy remains reverted independently.

The [machine-readable record](data/cumulative-native-2026-11-07.json) preserves
all cumulative workload summaries, CI coverage/job durations, source identities,
rejected-experiment limits and immutable archive hashes. Historical incremental
measurements retain their own source and host identities.
