# Core 2.0 performance portfolio

`core2-v1` adds eight workloads to the six unchanged MVP controls. All sources
are authored in this repository. The [manifest](../fixtures/portfolio.json)
records source/WAT/binary hashes, exact producer versions and flags, input sizes,
independent expected checksums and alternate seed/count vectors. No imports,
WASI, start section, external input or guest output occur in these modules.
Native engines may optimize/integrate operations according to their semantics.

| Case | Family | One iteration | Count per call | State reset inside call |
| --- | --- | --- | ---: | --- |
| float-convert | Numeric | f32 bitcast/multiply/add, f64 sqrt and saturating conversion | 25,000 | Seeded local accumulator |
| control-multivalue | Control | Irregular `br_table` and multi-value block transfer | 25,000 | Seeded local accumulator |
| indirect-reference | Calls/references | Table get/set, null check, typed indirect call and global update | 25,000 | Mutable global reset to seed |
| simd-shuffle | SIMD | Four i32 lanes, arithmetic, shuffle/rotate and unaligned vector store/load | 25,000 | Seeded vector; store precedes each load |
| bulk-unaligned | Memory | 512-byte clear, 257-byte fill, overlapping copy and unaligned i64 load | 1,000 | Every iteration resets its working memory |
| deep-calls | Calls/references | 33 nested helper activations plus return arithmetic | 2,000 | Locals; no retained state |
| image-filter-c | Applications | 3x3 binomial filter over 900 pixels of a 32x32 image, with checksum | 40 | Image generation is inside each call |
| packet-crc-rust | Applications | Generate and CRC-32 one 256-byte packet | 64 | Packet generator and CRC reset per iteration |

The Python oracle in `kernels.py` does not execute Wasm. The CRC oracle is also
checked against Python's independent `zlib.crc32`. Every runtime validates
first, warmup and measured results; small alternate vectors run as semantic
checks under Go/LLGo wasmvm, both wazero backends and the pinned C API runners.
SIMD results mix every lane and retain seed/count dependence; the bulk-copy
checksum reads beyond the original filled range to detect a missing copy.

Regenerate with `python3 benchmarks/embedded/portfolio/generate.py` using
wasm-tools **1.258.0**, Clang **22.1.8**, and the Rust **1.98.0** toolchain with
`wasm32-unknown-unknown` installed. `PORTFOLIO_CLANG` can select the exact Clang
binary. The generator verifies versions, records flags, and validates every
result with `wasm-tools validate --features=wasm2`. Clang uses `-nostdlib` and
Rust uses `#![no_std]`; these modules contain the authored functions rather
than vendored third-party workload sources. Committed binaries allow CI to
execute identical inputs without installing guest producers on timed hosts.

These are initial diagnostic families, not an application census or completed
M2 performance acceptance. Hosted reports retain each case independently; no
weighted aggregate or hard gate is accepted yet. Before M2 acceptance, the
versioned budget manifest must designate primary cases, family/case weights,
noise floors and controlled-run policies. M1 final review still needs separate
host-callback and preparation/resource-shape baselines; passing these kernels
does not replace those measurements or the complete conformance corpus.
