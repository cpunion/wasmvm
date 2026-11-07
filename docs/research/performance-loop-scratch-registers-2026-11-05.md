# Region register constraints and integer selection: 2026-11-05

**Full four-profile semantic CI passes; elapsed-time and final integration
acceptance remain open.** Engine source:
`2f8e44f493a127b6e9b2f82400b220626e9accbe`. The preceding
[shared-memory event](performance-shared-default-memory-hosted-2026-11-05.md)
measures another source and does not establish this implementation's speed.

The x86 memory-function inventory previously kept six integer words in
registers, while arm64 had eighteen. Pure target compilation of both SHA
producers exposed 33–37 frame references in the x86 compression loop. This
experiment changes the fixed-register constraints for a bounded region rather
than adding a workload-specific lowering or changing the native ABI.

One inventory now serves private native loops and proven memory-unroll groups.
The classified integer body can lend RDX/R8 when their constrained operations
do not need them. Scalar select keeps RDX; baseline popcount keeps both. Unknown
and floating/vector operations lend nothing. Proven default-memory accesses
can also lend R11, since their length check has already succeeded. This permits
nine integer homes in an eligible memory region without changing the ordinary
continuation inventory or the FP/vector bank.

All displaced words reach the canonical frame before replacement loads.
Normal exits and unroll backedges restore the enclosing plan and memory view;
shared memory refreshes its atomic size cell. A poll captures every borrowed
word before returning to the driver. Transfer costs gate both loop forms.
Every original cold PC retains the checked path. Neither instruction capability,
fuel/cancellation semantics nor complete-frame observation is relaxed.

Integer arithmetic/logical writes already clear the high half of an x86 i32
destination, so their preceding self-copies are unnecessary. Positive constant
shifts/rotates have the same property. Zero-count shifts retain the copy that
canonicalizes potentially noncanonical i32 input bits; an additional zero-count
shift instruction is then unnecessary. Existing independent numeric, alias,
boundary-value and cold-entry models remain in force.

## Static comparison and execution

The x86 diagnostic uses the complete production module compile/link path on
exact Clang/Rust fixtures and observed feature mask 7. Its only diagnostic
changes select that target and dump output; it stops before native publication
and never executes the generated x86 bytes on the arm64 host. All short backward
spans, full disassembly, original code and diagnostic hashes remain retained.

| Guest | Schedule instructions | Schedule frame references | Compression instructions | Compression frame references |
| --- | ---: | ---: | ---: | ---: |
| Clang | 67 → 47 | 5 → 0 | 112 → 92 | 37 → 24 |
| Rust | 67 → 47 | 5 → 0 | 108 → 88 | 33 → 20 |

Each listed loop keeps its poll. Compression-function native sizes change from
11042 to 10956 bytes (Clang) and 10929 to 10843 (Rust). Independently generated
Core integer/memory modules change from 2196/4382 to 2188/4284 bytes. These
static counts do not quantify dynamic latency, CPU shares or competitor parity.

Ten complete affected packages pass Go 1.27.3 and LLGo `3571c25`
(`-O3 -lto=full`) on macOS arm64: **108 actual conformance reports**. Codegen
coverage is **97.83% (7628/7797)**; all ten package gates exceed 90%. Race,
cgocheck2, no-cgo and full local `go vet` checks pass. Ninety-six normal runner
configurations execute both compilers/backends across all 22 fixtures and both
C/Rust SHA work sizes; their times are correctness evidence only.

New models cover every original frame word and cold continuation, valid and
OOB memory, private/shared views, i32/i64, select/popcount constraints, normal
exit length restoration and forced private-backedge polls. The independent
precise tier supplies the reference. Both target encodings are compiled; the
host architecture executes. Actual arm64 publication remains byte-for-byte
identical to the preceding shared-memory engine under both host compilers:
101308-byte Clang and 101476-byte Rust modules, with matching entry offsets.

The first test poll patch incorrectly classified backward cold transfers as
private loops; it now matches the actual poll-exit continuation. Arbitrary
interrupted recurrence temporaries also caused unsigned wraparound loops in a
development model; finite cold states retain arbitrary data parameters while
zeroing those recurrence temporaries. Both failures and their raw logs remain
retained. These were test-construction corrections, not discarded timing runs.

The stages `8cce483` (private loops), `b97bf83` (shared unroll planner), and the
final source above have separate local qualification. Earlier reports are not
reattributed to the final source. [CI 41984737690](https://github.com/cpunion/wasmvm/actions/runs/41984737690)
passes all 21 required jobs across the complete platform/compiler matrix;
one optional reference job is intentionally skipped. The artifact audit checks all
240 actual conformance reports, 31 packages per LLGo platform, manifest pins and
compiler identities. All 30 full-CI package coverage gates exceed 90% (minimum
90.77%); the longest job takes 532 seconds (8m52s). The separate CI bundle has
SHA-256 `9318a63a68f088dee0917da3acd01142438c80a82a62240c20e0f3dfa1a40960`.
CPU policy, toolchain pins and alignment remain unchanged. Final PR-head
checks/review, controlled competitor/native performance,
general SSA/value allocation and broader application/WASI-world acceptance remain
open. Changes stay on the contribution branch while LLGo PR #2991 remains open;
LLGo #2983 stays with its existing contributor.

The [evidence index](data/loop-scratch-registers-2026-11-05.json) records exact
commands, sources, compiler/binary identities, coverage and the raw manifest.
Local archive: `artifacts/loop-scratch-registers-20261105/evidence.tar.gz`,
5556799 bytes, SHA-256
`621db72f397a3398667038ccaaf6a3da87f5e10359c90455207b838d8bbfaeab`.

The [complete hosted comparison](performance-loop-scratch-registers-hosted-2026-11-05.md)
measures this exact engine against the preceding shared-memory source. All six
jobs pass; the audit checks 22 workloads and 362 alternate configurations.
LLGo native Clang/Rust SHA elapsed medians are 0.792x/0.816x base, but remain
1.554x/1.607x Wasmtime and 1.866x/1.900x unchecked C. Eight workloads resolve
the four diagnostic A/A controls. Go interpreter and preparation regressions
remain recorded; this hosted event does not close controlled acceptance.
