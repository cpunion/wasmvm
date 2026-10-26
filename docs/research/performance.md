# Basic Runtime Performance Comparison

Measurement date: **2026-10-26**, on a local Apple M4 Max (arm64), 16 logical CPUs, 128 GiB RAM, macOS 27.0.1. The official wazero CLI was built with Go 1.27.3. This is a local engineering baseline, not a general ranking of WebAssembly engines.

## Result and environment qualification

The sustained workloads show a substantial interpreter/native-code gap in this test. Wazero's native compiler is materially faster than its own interpreter on both sustained cases. WasmEdge's interpreter has lower sustained-case medians than wazero's interpreter even though its startup and small-case costs are higher. Workload size therefore changes the interpretation of the small-case results.

**The machine was under unrelated load.** After the first expanded run, a process snapshot showed video encoding at approximately 600% CPU and virtualization at approximately 400% CPU, in addition to normal desktop activity. Those processes were not stopped or modified. A final repeat paused this project's builds, but it did not create an isolated or quiet host. The repeated run's 1/5/15-minute load averages were **39.31, 51.03, 51.51 before** and **35.63, 48.83, 50.68 after**. These are system load averages, not CPU utilization percentages.

Near-tied native-CLI timings and their ordering are not robust enough to choose an engine under these conditions. The data does establish executable fixtures and a reproducible comparison method, and shows why both startup controls and larger workloads are necessary.

All numbers below are milliseconds, formatted as **median [minimum, p95]** from the final repeat. Each cell has two discarded warmups and seven measured process executions. Nearest-rank p95 is therefore the maximum of only seven samples, not a reliable tail-latency estimate.

### Startup and small workloads

| Runtime/backend | Startup control | Integer loop | Function-call source | Memory load/store |
| --- | ---: | ---: | ---: | ---: |
| wazero-interpreter | 10.810 [7.309, 27.001] | 47.185 [45.776, 81.409] | 36.934 [33.040, 70.081] | 56.392 [41.645, 93.256] |
| wazero-compiler | 10.587 [8.234, 19.840] | 11.755 [9.886, 17.572] | 15.302 [8.244, 22.990] | 12.626 [8.380, 21.444] |
| wasmtime-cranelift | 26.721 [10.832, 48.709] | 20.441 [14.451, 41.057] | 19.440 [14.150, 48.043] | 16.204 [9.399, 32.378] |
| wasmer-cranelift | 31.044 [18.162, 44.716] | 26.897 [17.604, 56.011] | 30.831 [18.225, 77.802] | 36.874 [17.785, 47.342] |
| wasmedge-interpreter | 36.979 [32.295, 60.942] | 55.866 [52.753, 77.440] | 55.164 [46.984, 66.775] | 62.174 [55.112, 73.905] |

### Sustained workloads

| Runtime/backend | Integer loop: 5,000,000 iterations | Memory load/store: 3,000,000 iterations |
| --- | ---: | ---: |
| wazero-interpreter | 857.355 [823.202, 948.511] | 976.744 [860.872, 1141.231] |
| wazero-compiler | 22.988 [22.213, 37.667] | 13.574 [11.490, 30.039] |
| wasmtime-cranelift | 33.167 [26.683, 39.979] | 22.258 [14.327, 55.381] |
| wasmer-cranelift | 36.091 [29.801, 56.222] | 24.775 [17.681, 34.910] |
| wasmedge-interpreter | 592.402 [526.441, 647.115] | 585.038 [563.380, 1089.929] |

**Every warmup and measured execution exited successfully and produced exactly the expected checksum line.** The final repeat contains 60 warmup observations and 210 measured observations across five configurations and six fixtures.

All runs are retained, including the noisier observations:

- [Final expanded repeat](../../benchmarks/results/2026-10-26-darwin-arm64.json): the tables above, with load averages and an environment note.
- [First expanded run under contention](../../benchmarks/results/2026-10-26-darwin-arm64-contended.json): all six workloads; retained rather than discarded.
- [Initial three-small-workload baseline](../../benchmarks/results/2026-10-26-darwin-arm64-small.json): lower absolute timings before the expanded measurements; it is not directly comparable as a controlled A/B run because host load differed.

Each JSON records every sample, command, stdout/stderr, version, host detail, fixture hash, and summary. No sample outlier was removed from a completed run.

## What was measured

The metric is elapsed wall time around a new CLI subprocess: process startup, file loading, Wasm validation/compilation, instantiation, guest execution, output capture, and shutdown. Compilation is inside the measured interval; installing/building the host CLI and generating fixtures are outside it.

Each runtime receives the same committed `.wasm` bytes with fixed inputs:

| Workload | Work | Correctness guard |
| --- | --- | --- |
| Startup control | No workload loop; the same P1 output path | Checks zero and writes a checked success line |
| Integer loop | 250,000 iterations of i32 rotate/xor/multiply/add | Independent Python checksum oracle; guest traps on mismatch |
| Sustained integer | 5,000,000 iterations of the same loop | Same guard; 20 times the small-case work |
| Function calls | 150,000 source-level calls performing integer mixing | Same guard; native compilers may inline calls |
| Memory load/store | 150,000 strided read/modify/write operations over 32 KiB | Same guard; uses one 64 KiB linear memory |
| Sustained memory | 3,000,000 operations over the same 32 KiB region | Same guard; 20 times the small-case work |

The guest prints its checksum through the same WASI Preview 1 `fd_write` ABI. The harness checks exact stdout and exit status for every sample. The [fixture manifest](../../benchmarks/fixtures/manifest.json) records expected outputs, operation counts, source/binary SHA-256 values, and the generator (`wasm-tools 1.258.0`). These fixtures were authored for this project; no external benchmark implementation was copied.

## Runtime configuration

| Runtime | Version | Selected backend and cache policy |
| --- | --- | --- |
| wazero | v1.12.0, official CLI built with Go 1.27.3 | `run -interpreter -cachedir=`; no disk compilation cache |
| wazero | v1.12.0, same host executable | `run -cachedir=`; default native wazevo compiler on this arm64 host |
| Wasmtime | 48.0.1, commit 7bac2c277 | Cranelift, opt-level 2, `-C cache=n` |
| Wasmer | 7.3.0 | `--cranelift --disable-cache --offline --no-tty` |
| WasmEdge | 0.17.1 | Raw Wasm with explicit `--run-mode interpreter`; no AOT artifact or JIT |

The versions and options above were read from the installed tools and recorded in the raw results. The comparison includes both interpreter and native-code configurations; it does not treat these as identical execution strategies. Wazero's two modes share the same official CLI binary, while cross-project comparisons include each project's own CLI and host-layer overhead.

Compilation caches are disabled. Operating-system executable/filesystem pages are warmed by the warmups and are not flushed. Configuration/workload order is shuffled within each round using seed 20261026. The runner requires all four projects to be installed and fails on a missing runtime, timeout, trap, or incorrect output rather than quietly dropping a result.

## Interpretation and limits

- This is **process-level latency**, not warmed in-process guest throughput. The startup control and 20-times-larger variants expose workload-size sensitivity, but subtracting their medians does not yield an exact compilation or execution time. Some startup medians exceed small-workload medians under contention, reinforcing that limitation.
- This is one busy local development machine with no CPU affinity or workload isolation. Seven samples and six synthetic workloads do not cover real application behavior, large modules, GC-heavy code, parallel execution, P2/P3 components, peak RSS, or energy use.
- Interpreter/native backend choices are explicit. Native inlining is permitted, so the function-call source workload does not measure a guaranteed count of machine-level function calls.
- These measurements used **official Go to build wazero**. They do not measure LLGo-generated host code. Host-compiler fixes and their compatibility evidence belong in the separate [wazero evaluation](wazero-evaluation.md).
- Wazero's native path is a useful reuse candidate, but performance alone does not establish feature completeness, bounded cancellation, LLGo portability, or acceptable maintenance cost. No performance threshold selects the implementation at this stage.

A later performance phase should add large real applications, startup versus compile/instantiate/invoke measurements within each embedding API, GC and threading scenarios, and controlled host environments. Those measurements need matched boundaries before comparing engines.

## Reproduction and harness checks

```sh
python3 benchmarks/generate.py
python3 -m unittest discover -s benchmarks -p 'test_*.py'
python3 benchmarks/run.py --warmups 2 --samples 7 --output benchmarks/results/local.json
```

The runner installs the pinned official wazero CLI into the project's ignored `artifacts/benchmark-tools` directory; it does not replace a system binary. Fixture regeneration requires wasm-tools 1.258.0 to reproduce the recorded producer version. Existing fixture hashes are checked before measuring. Six harness tests passed, covering statistics, invalid measurements, fixture integrity, rejection of incorrect output/exit status, recording of successful observations, and checksum generation. Actual runtime execution is evidenced separately by the raw benchmark observations.
