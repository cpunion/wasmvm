# Modern Core performance fixtures

These ten project-authored binaries retain the GC and Core Threads evaluation
workloads: i31 round trips, mutable struct fields, linked allocation, indirect
array/struct access, private/shared ordinary memory, atomic RMW, empty notify,
wait mismatch and zero-timeout waits. Each exports `run(i32, i32) -> i32`.
Every call resets the state relevant to its checksum. Allocation retains a
linked chain; it is not a full collector/scaling benchmark. Thread cases are
uncontended and do not measure thread creation, contention or scheduling latency.

The manifest pins WAT and binary hashes, inputs, results and feature requirements.
`verify.py` independently simulates the integer/cell recurrences and wait results
for both declared input vectors, then requires byte-identical producer output:

```sh
python3 benchmarks/modern/verify.py --producer /path/to/wasm-tools
```

The original binary producer was wasm-tools 1.258.0; reproduction is recorded in
the audit artifact. A different producer is accepted only if every byte agrees.
One checked-in WAT line has normalized trailing whitespace; its original source
and hash remain in the measurement bundle, and the Wasm binary is identical.
The Go/LLGo Core survey archives exact engine sources; `survey_modern.py` changes
only its saved measurement adapter to enable `wasm.SupportedFeatures`. The engine
source digest is checked unchanged. Reference adapters explicitly enable Threads.
Actual GC validation rejections from wazero/Wasmer remain separate records;
crashes, timeouts, traps and unrelated compiler errors fail the survey.

Run `python3 benchmarks/survey_modern.py --core-output OUTPUT` after the Core
survey build. It retains three fresh, serially interleaved process blocks, seven
warm calls, two warmups, every phase and every checksum. No unsupported engine
receives a synthetic timing. Descriptive local results do not close controlled
performance acceptance.
