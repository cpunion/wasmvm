# Local runtime performance comparison

This is a reproducible **CLI end-to-end latency** comparison, not a measurement of warmed in-process execution throughput. Every measured observation starts a new process and includes runtime startup, Wasm validation/compilation, instantiation, guest work, output, and shutdown.

The same six project-authored WASI P1 binaries run with the same fixed inputs and checked output in every configuration:

- Startup/P1 output control: no workload loop.
- Integer rotate/xor/multiply/add loop: 250,000 iterations.
- Sustained integer variant: 5,000,000 iterations (20 times the smaller loop).
- Function-call source workload: 150,000 iterations; native compilers may inline calls.
- Linear-memory read/modify/write workload: 150,000 iterations over a 32 KiB region.
- Sustained memory variant: 3,000,000 iterations over the same region.

Each module computes a checksum, traps if it differs from the independent Python oracle, and then writes the expected checksum line through `fd_write`. The runner also verifies exit status and exact stdout for every warmup and measured sample. Source and binary hashes, generator version, inputs, and outputs are recorded in `fixtures/manifest.json`. No third-party benchmark code was copied.

## Reproduce

Prerequisites: Python 3.9+, Go compatible with wazero v1.12.0, Wasmtime, Wasmer, and WasmEdge. Regenerating the fixtures also requires `wasm-tools`; the committed fixtures were generated with version 1.258.0.

```sh
python3 benchmarks/generate.py
python3 -m unittest discover -s benchmarks -p 'test_*.py'
python3 benchmarks/run.py --warmups 2 --samples 7 --output benchmarks/results/local.json
```

The runner builds the official wazero CLI at v1.12.0 into `artifacts/benchmark-tools`; it does not replace an installed executable. The measured configurations are wazero interpreter, wazero native compiler, Wasmtime Cranelift at optimization level 2, Wasmer Cranelift, and the explicit WasmEdge interpreter. Missing runtimes or incorrect output fail the run rather than disappearing from the results.

Compilation caches are disabled: wazero has no cache directory, Wasmtime uses `cache=n`, and Wasmer uses `--disable-cache`. WasmEdge receives raw Wasm in interpreter mode. OS executable and filesystem pages are warmed by the warmups and are not flushed. Order is shuffled within each round with a recorded seed. There are at least two warmups and seven measured observations per runtime/workload pair.

Raw JSON includes every measurement, command, stdout/stderr, runtime version, host metadata, fixture hash, and summary. Later runs also record load averages and an explicit environment note. p95 uses nearest rank; with seven samples it is the sample maximum and is not a reliable tail-latency estimate. No CPU affinity, system isolation, memory/energy measurement, or broad application coverage is implied. Different CLIs have different startup and host-layer costs, all included in this metric. The startup control and larger workloads help expose sensitivity to workload size; subtracting their medians does not yield an exact compilation or execution time.

The local results and their interpretation are documented in [performance.md](../docs/research/performance.md). Compiler compatibility and specification coverage remain separate acceptance criteria.
