# Runtime performance comparisons

The benchmark job records two distinct metrics: **direct embedding API phases and warmed execution**, and **CLI end-to-end latency**. Execution efficiency is assessed with the direct API measurements. The CLI series retains application startup, validation/compilation, instantiation, guest work, output, and shutdown in its measured interval; its totals are not pure execution times.

The [embedding protocol](embedded/README.md) defines the phase boundaries, pure parameterized fixtures, clock and empty-call controls, and comparison limits. Phase results appear first in the PR comment, with complete phase tables and raw samples. No CLI or control median is subtracted to infer guest time.

The CLI series uses the same six project-authored WASI P1 binaries with fixed inputs and checked output in every configuration:

- Startup/P1 output control: no workload loop.
- Integer rotate/xor/multiply/add loop: 250,000 iterations.
- Sustained integer variant: 5,000,000 iterations (20 times the smaller loop).
- Function-call source workload: 150,000 iterations; native compilers may inline calls.
- Linear-memory read/modify/write workload: 150,000 iterations over a 32 KiB region.
- Sustained memory variant: 3,000,000 iterations over the same region.

Each module computes a checksum, traps if it differs from the independent Python oracle, and then writes the expected checksum line through `fd_write`. The runner also verifies exit status and exact stdout for every warmup and measured sample. Source and binary hashes, generator version, inputs, and outputs are recorded in `fixtures/manifest.json`. No third-party benchmark code was copied.

## Reproduce

Prerequisites: Python 3.12+ for the complete tooling/test suite, Go compatible with wazero v1.12.0, Wasmtime, Wasmer, and WasmEdge. Regenerating the fixtures also requires `wasm-tools`; the committed fixtures were generated with version 1.258.0.

```sh
python3 benchmarks/generate.py
python3 -m unittest discover -s benchmarks -p 'test_*.py'
python3 benchmarks/run.py --warmups 2 --samples 7 --output benchmarks/results/local.json
```

The runner builds the official wazero CLI at v1.12.0 into `artifacts/benchmark-tools`; it does not replace an installed executable. The measured configurations are wazero interpreter, wazero native compiler, Wasmtime Cranelift at optimization level 2, Wasmer Cranelift, and the explicit WasmEdge interpreter. Missing runtimes or incorrect output fail the run rather than disappearing from the results.

Compilation caches are disabled: wazero has no cache directory, Wasmtime uses `cache=n`, and Wasmer uses `--disable-cache`. WasmEdge receives raw Wasm in interpreter mode. OS executable and filesystem pages are warmed by the warmups and are not flushed. Order is shuffled within each round with a recorded seed. There are at least two warmups and seven measured observations per runtime/workload pair.

Raw JSON includes every measurement, command, stdout/stderr, runtime version, host metadata, fixture hash, and summary. Later runs also record load averages and an explicit environment note. p95 uses nearest rank; with seven samples it is the sample maximum and is not a reliable tail-latency estimate. No CPU affinity, system isolation, memory/energy measurement, or broad application coverage is implied. Different CLIs have different startup and host-layer costs, all included in this metric. The startup control and larger workloads help expose sensitivity to workload size; subtracting their medians does not yield an exact compilation or execution time.

The local results and their interpretation are documented in [performance.md](../docs/research/performance.md). Compiler compatibility and specification coverage remain separate acceptance criteria.

## Pull-request benchmark job

[Runtime benchmarks](../.github/workflows/benchmarks.yml) runs on pull requests, pushes to `main`, and manual dispatch. It checks out the exact PR head and base commits and runs all configurations on the same Ubuntu 24.04 amd64 runner. A push compares against its previous commit; a manual run accepts a baseline ref.

The candidate dependency is selected from each checkout's `experiments/wazero/go.mod`. The CLI series builds the official CLI; the phase series builds identical head embedding-runner source against each selected dependency. Base and head use Go 1.27.3, with both interpreter and native modes measured. This reference series compares wazero dependency selections. The phase harness separately builds and measures the exact wasmvm base/head engine and production CLI under both Go and LLGo; see the [embedding protocol](embedded/README.md). When both revisions select the same wazero version, the report says so; observed differences then measure run-to-run noise, not an engine code change.

Wasmtime 48.0.1 (Cranelift), Wasmer 7.3.0 (Cranelift), and WasmEdge 0.17.1 (interpreter) provide the competitor comparison. Their official Linux binaries are pinned by version and SHA-256 and checked before extraction. No floating installer is executed. All seven configurations receive the exact same head fixture bytes and expected outputs, with shuffled, interleaved rounds, two warmups, and seven measured samples per case. Build and download time are outside the measured intervals.

The [tool manifest](tools-linux-amd64.json) records release URLs, archive sizes and hashes, selected files, and license notices. The Linux installer requires Python 3.12 or later, as provided by the CI runner, and writes the verified installation record into the measurement artifact. Its `--sdk` option also installs the pinned C headers/libraries used by the embedding runners. Installed files stay in a fresh temporary directory rather than replacing system tools.

The job publishes a Markdown summary and a `runtime-benchmark-comparison` artifact containing separate `phases/` and `cli/` raw observations and reports, retained for 30 days. The phase series uses three fresh processes per configuration/workload, each with a separately timed first call, two warmups, and seven measured warm calls. It directly measures startup, loading, engine/store setup, preparation, instantiation, lookup, calls, and cleanup. A separate job creates or updates one marked PR comment with baseline/head medians, percentage changes, competitor results, commit identities, and the run link. It checks that both PR head and base still match before posting, so an obsolete run cannot replace a current result. The comment job does not execute checkout code and is the only job granted comment-writing permission. Same-repository PRs receive automatic comments; fork PRs retain read-only tokens and publish the job summary and artifact.

Incorrect output, traps, missing tools, broken downloads, and incomplete result matrices fail the job. Timing changes are advisory because hosted runners and these small samples do not provide a stable performance budget. In particular, an apparent improvement between two identical candidate versions is not an optimization. The report labels the process-level metric and backend differences; it does not treat interpreter/native timings as the same execution strategy or compare a Linux run numerically against the earlier macOS baseline.
