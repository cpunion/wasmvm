# Loop value versions and shared lowering: 2026-11-05

The [subsequent CFG residency correction](performance-cfg-register-residency-2026-11-05.md)
retains both adverse hosted events and replaces layout-based loop detection
with actual cyclic components and shared allocation costs. The stages below
remain historical evidence for their exact source revisions.

**Local dual-compiler semantics pass. Throughput and final four-profile
acceptance remain open.** Engine source:
`8b5ec48a77abb155a98ef009add078a1d34a3447`; baseline engine:
`c1f00f55b9645ab10e413b7d2fc19b53a204231c` from the
[global-residency experiment](performance-global-register-plans-2026-11-05.md).
The earlier [hosted routing comparison](performance-direct-runtime-routing-2026-11-05.md)
measures a different source event. Its timing ratios do not measure this pass.

The compiler now versions slot writes within bounded native loops. Copy-related
slots can share an existing physical register only when every overlapping
version has identical bits and their final frame words agree. Dead writes still
occupy an instruction point: their emitted writes cannot clobber a different
live value. Initial distinct words reach the canonical frame before the shared
body starts. Integer/reference, scalar FP and full V128 storage use the same
analysis, with exact type and literal-bit identities. The interference work
budget is proportional to body length; exhaustion retains the original plan.

Private backedges now share one emitter for pure arithmetic, individually
proven memory accesses and whole-loop affine memory proofs. The latter retain
their existing progress/span guards and checked fallback. Pure loops require
no artificial memory proof. Calls, possible traps, memory-size observations,
interior entries, edge copies and 256-instruction poll crossings prevent this
transform. Conditional branches support both polarities and copied booleans;
unconditional loops retain bounded interruption. Every original continuation
keeps its independent checked entry and complete canonical snapshot.

Instruction selectors split before and after any shared physical-register
write. Read-only sequences retain fusion because their inputs cannot change
before the next write boundary. Vector aliases expire on a write through any
logical slot sharing their physical source. This preserves one shared selector
and emitter framework rather than adding special lowering for either SHA
producer. There is no ISA, toolchain, fuel-policy or alignment change.

Ten complete affected packages pass Go 1.27.3 and LLGo 3571c25
(`-O3 -lto=full`) on macOS arm64: 108 actual conformance reports. Scoped Go
coverage is 90.71%–100%, including codegen 97.26% (7464/7674 statements).
Race, cgocheck2 and no-cgo checks pass. Ninety-six new-source configurations
validate both backends/compilers across 22 fixtures and two independent C/Rust
SHA work sizes. Unchanged baseline/reference binaries are reused by hash.
These execution times are correctness evidence, not performance measurements.

Models compare every original PC against the independent precise native tier,
including every frame word, full memory contents, NaN payloads, signed zero,
both vector halves, reference words, OOB/overflow paths and generated poll exits.
They exercise both branch polarities, forwarded flags, copied conditions and
unconditional loops. Shifted-operand selection remains required when a shared
input has independent destinations; shared destination writes correctly split
selection. Both target encodings are checked and the host target executes.

Actual published ARM64 bytes, using an output-only diagnostic, give:

| Guest producer | Schedule loop, base → head | Compression loop, base → head | Module native bytes, base → head |
| --- | ---: | ---: | ---: |
| Clang | 29 → 29 | 48 → 47 | 101804 → 101808 |
| Rust | 29 → 29 | 42 → 41 | 102248 → 102268 |

Both host compilers publish identical bytes and entry offsets. Each listed loop
retains its poll. Clang compression retains one frame load; Rust retains none.
The module sizes rise by 0.004%/0.020%. Static counts do not quantify executed
latency or CPU shares. The earlier pinned Wasmtime ARM64 compression spans are
43/41 instructions and unchecked C is 36; neither is a safety-equivalent timing
result. The remaining native/compiler gap is not closed by one removed copy.

The first qualified implementation split at reads too, increasing Rust
compression from 42 to 46 instructions by suppressing rotate/XOR fusion. Its
full evidence is retained under `complete/`; the corrected final source uses
write boundaries. An overbroad development assertion also incorrectly demanded
fusion when the destination itself shared the branch condition. Both failures
and the corrected all-PC models remain in the evidence. No timing cohort was
retried or filtered.

The [evidence index](data/loop-value-coalescing-2026-11-05.json) retains exact
commands, source/compiler/binary identities, coverage, static assembly and all
archive checksums. Local archive:
`artifacts/loop-value-coalescing-20261105/evidence.tar.gz`, 5,034,345 bytes,
SHA-256 `d6f102c8d2000e19879e5687d9332ed90d014b4ddac9f9cd6face8428a496c44`.

This is bounded copy coalescing, not whole-function SSA/value allocation.
General splitting, spill/rematerialization, final-head four-profile CI (31
packages/profile, 30 coverage gates), review, controlled competitor/native
acceptance and the broader application/WASI-world portfolio remain open.
Changes stay on a contribution branch while external PR #2991 is open; that
upstream PR cannot be merged under the repository-owner policy. LLGo #2983
remains with its contributor. No milestone or goal is marked complete.

The first [full CI](https://github.com/cpunion/wasmvm/actions/runs/41960841770)
at `e6471ee11ec58be409f74c21acb4c70e75c2aa1d` passes 20 checks and fails the
macOS Go P3 TCP busy-poll consumer. The longest successful job is 8m25s.
The failure reproduces on one CPU with eight sockets: one live guest handle,
eight physical objects and six queued half-closes. Source
`9f8acb2d34aa6911ddf18ea8a871f0c19b4f29ad` removes that cleanup queue,
performs nonblocking half-closes synchronously and joins admitted cleanup
before forced shutdown. It preserves physical/published quotas. The complete
P3 package passes both compilers, scoped coverage 90.72% (3961/4366), and socket
race checks pass. The strengthened upstream consumer keeps the eight-socket,
single-processor condition. This is a wasmvm lifecycle fix, unrelated to #2983.

The [22-workload hosted event](https://github.com/cpunion/wasmvm/actions/runs/41960845658)
at that earlier CI head completes all five groups, 362 alternate configurations
and 12 explicit GC rejections. Only six workloads resolve all four A/A controls;
the largest job is 8m41s. It reveals unacceptable native regressions versus the
own baseline: LLGo memory-load-store 5.918x, narrow memory 2.014x and Rust SHA
1.114x. Their complete data, including every reference and phase, are retained.
Neither smaller ARM64 instruction counts nor passing semantics waive them.

Source `965a40e6978934186ed713fa9bdf1ca54ac2b909` extends residency through
backedges. The earlier lexical intervals could expire before a branch block,
forcing repeated mapping transfers on a hot backedge. Nested/crossing spans
now merge; all intersecting slot intervals cover the complete loop. Independent
interval-closure models and header/body/backedge mapping tests cover the change.
Ten packages and 108 actual reports pass both compilers; codegen coverage is
97.34% (7499/7704). Race, cgocheck2, no-cgo and 96 runner configurations pass.
Both host compilers publish identical SHA bytes. Hot ARM64 SHA counts stay
29/47 and 29/41; module sizes become 102556/102916 bytes.

The subsequent local ARM64 diagnosis retains all five blocks, complete phases
and original/bad/new source identities. Every A/A interval is inconclusive;
it cannot establish recovery of the Linux x86 regression and will not be retried
for a favorable result. Final-source CI, a new exact-source hosted comparison
and controlled performance acceptance remain required. The evidence index
keeps all three source stages distinct, with separate checksummed archives.
