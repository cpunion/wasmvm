# Compiled Core application workloads

These modules compute complete SHA-256 digests over deterministic messages,
including message generation, block chaining and final padding. Each call
resets its state and returns the first 32 digest bits. This complements the
instruction-family fixtures with compiled compute programs from two frontends.
It does not replace the other Core, WASI or controlled performance gates.

The message has `count` 64-byte blocks. Its words follow
`x = 1664525*x + 1013904223 (mod 2^32)`, starting from the runtime seed; words
are serialized little-endian. The independent Python oracle streams these
bytes into `hashlib.sha256`. The C and Rust implementations perform the hash
computation themselves. Counts outside 1..65536 trap. The public Wasm export is
`run(seed: i32, count: i32) -> i32`; neither module imports functions or starts
execution during instantiation. Normalization is ns per 64-byte message block,
including the final padding block's work in the invocation total.

The committed fixtures pin Clang 22.1.8 and Rust 1.98.0 (LLVM 22.1.8), with
LLD 23.1.4 and wasm-tools 1.262.0. The frontends share an LLVM backend; these are
two pinned guest toolchains, not two independent machine-code generators.
`manifest.json` records commands, sources, libraries, compiler/validator hashes
and exact guest bytes. Validation explicitly selects `-all,wasm2`.

To regenerate with these installed tools:

```sh
python3 benchmarks/applications/build.py \
  --clang /path/to/clang --rustc /path/to/rustc \
  --linker /path/to/wasm-ld --wasm-tools /path/to/wasm-tools \
  --evidence artifacts/application-producers
```

The Rust compiler needs its `wasm32-unknown-unknown` target libraries. `--only
rust` or `--only clang` preserves the other producer's verified existing row.
No compiler is downloaded or shared toolchain modified by this script.

Use the existing phase harness for both owned host compilers and independent
runtime references:

```sh
python3 -m unittest discover -s benchmarks -p test_routing_compare.py
python3 benchmarks/routing_compare.py build \
  --base-dir /path/to/base --head-dir . --tools-dir /path/to/pinned-tools \
  --output-dir /path/to/routing-build
python3 benchmarks/routing_compare.py measure \
  --head-dir . --build-dir /path/to/routing-build --group applications \
  --output-dir artifacts/application-comparison
```

The measurement command currently requires Linux CPU affinity. The opt-in
`routing` scope of the Runtime benchmarks workflow performs the same comparison
on hosted runners, building all binaries once and retaining both compiler and
reference results. It keeps preparation, first/warm calls, readiness and cleanup
separate. All participants receive identical guest bytes within each producer.
Hosted A/A calibration remains distinct from controlled milestone acceptance.

An ordinary native C control uses the same authored C algorithm:

```sh
clang -O3 -std=c11 benchmarks/applications/sha256.c -o artifacts/sha256-native
artifacts/sha256-native 305419896 5000 3307539256
```

Its volatile indirect entry prevents repeated pure calls from being hoisted
outside their clocks. It retains first, warmup and warm samples, and uses
CLOCK_UPTIME_RAW on Darwin. This unchecked C control has no Wasm memory guards,
fuel, cancellation or VM lifecycle; its timings do not establish full policy
parity or replace a VM's measured phases.
