# Stable default shared-memory proofs: 2026-11-05

**Local dual-compiler semantics pass; throughput and final integration remain
open.** Engine source: `fb6e872b4758eb5c4e52c334e23c9e12036a58bb`. This follows
the [CFG residency and selector correction](performance-cfg-register-residency-2026-11-05.md).

Shared storage already reserves a stable base and publishes a monotonically
increasing length through a native atomic cell. The compiler previously
discarded the declared minimum and reloaded its descriptor on every access.
A function that selects only shared memory zero can now use the existing
bounded range, guarded-loop and unroll proofs. Entry validates the base and
atomically observed length against the declared minimum. Proven accesses reuse
the reserved base registers. No new allocator, execution path or ABI field is
introduced; the eligibility flag exists only during compilation.

The entry length is a lower bound. `memory.size`, atomics and accesses without
a range proof still load the current shared length. Concurrent growth cannot
invalidate a successful proof; an older lower bound can only select the
ordinary checked fallback. Functions selecting another memory keep their
general view path. Memory64, precise fuel, helper and cold-continuation
contracts remain in effect. No memory access is removed or reordered.

New tests cover scalar/vector widths, complete frame/memory effects at every
original PC, invalid entry bindings, mixed memories and guarded/unrolled loops.
A guest atomically marks that it has entered, then waits in native code for
another thread to grow memory. It must observe the new `memory.size` and load
the new page during that same entry; a cached size would hang or trap. The old
blanket shared-memory unroll rejection was replaced with indexed-memory and
helper-only rejection cases, and the initial failing assertion is retained.

Ten complete affected packages pass Go 1.27.3 and LLGo `3571c25`
(`-O3 -lto=full`), producing 108 actual conformance reports. Scoped codegen
coverage is **97.33% (7535/7742)**; all ten package gates exceed 90%. Race
checks include codegen, platform ownership, native execution and the public VM.
Race, cgocheck2 and no-cgo checks pass. Full four-profile CI, exact-source
performance measurement, review and integration remain required.

Ninety-six new-source runner configurations pass across both compilers and
backends, 22 fixtures and both C/Rust SHA work sizes. Actual arm64 publication
preserves the exact private SHA code bytes and entry offsets from engine
`2be7d9f`, under both compilers: 101308 bytes for Clang and 101476 for Rust.
These checks establish correctness and static identity; their times are not
performance results.

The [evidence index](data/shared-default-memory-2026-11-05.json) records exact
commands, source/toolchain hashes, reports and the raw archive manifest.
`artifacts/loop-value-coalescing-20261105/shared-default/evidence.tar.gz`
contains 158 files, 2420994 bytes, SHA-256
`ae11b912d0ac7d00b694347e57d65b0d7df2a2bb26384b1019f11bdf2df99cc7`.
