# Global slot residency and canonical exits: 2026-11-05

**Local semantics pass; performance acceptance and SSA value-level allocation
remain open.** Engine source: `c1f00f55b9645ab10e413b7d2fc19b53a204231c`.
The baseline binary source is `f0aa74998abafdbd28db6c08efb713a2946da580`,
with engine `59a76f933901e56266e618bff0a502364eca0350`. This is an unmerged
development comparison; production main remains
`4ce79899a1f3ed45163afa622cc408cc6a79a599`.

The shared emitter now supports per-block mappings across the whole function,
integer/reference and shared FP/vector inventories, interval expiration,
spilling and reentry. All displaced words store before replacements load.
Edges carry their source mapping; cold entry, division traps, deferred polls,
calls and guard fallback preserve canonical frames and PCs. Entry reloads only
the selected mapping. A typed pending heap avoids interface boxes. Independent
tuple-work and map-cell budgets retain conservative allocation on exhaustion.
The same implementation serves amd64 and arm64; no interpreter, ISA, compiler
pin or semantic-fuel policy changes.

Go 1.27.3 and LLGo `3571c2523e010e6842c4dd8a5213554fdc72d2a3`
(`-O3 -lto=full`), LLVM 22.1.8, macOS arm64:

- Ten complete affected packages pass both compilers: 108 actual conformance
  reports, scoped Go coverage 90.71%–100%; codegen 97.16% (7216/7427 statements).
- New independent frame-word models execute every original PC across forward
  jumps, conditions, branch tables, reference tests, all register classes,
  division traps, calls, helper exits and generated cancellation polls. They
  compare every frame word, including dead words and both SIMD halves.
- Race, cgocheck2 and no-cgo checks pass. Work/code budgets and register
  reservations remain tested. The earlier loop-selection counter test now
  accepts global allocation as the source of pressure relief; its complete
  cold-PC/frame checks remain intact.
- The new source passes 96 fresh configurations across both compilers/backends and all
  22 routing fixtures, including C/Rust SHA at counts 17 and 5000. Unchanged
  baseline and reference binaries/SDKs are reused with verified hashes.

An output-only publication hook in isolated diagnostic archives captures actual
native bytes. Pinned Go and release-profile LLGo publish byte-identical modules
and entry offsets for both SHA producers. This verifies those generated modules,
not equal host boundary cost or equal whole-runtime performance.

| SHA producer | Baseline native bytes | Head native bytes | Change |
| --- | ---: | ---: | ---: |
| Clang | 100440 | 101804 | +1.36% |
| Rust | 100832 | 102248 | +1.40% |

Static machine instructions in the corresponding scheduling/compression loops:

| Producer / loop | Baseline | Head | Wasmtime 48.0.1 | Ordinary native C |
| --- | ---: | ---: | ---: | ---: |
| Clang schedule | 29 | 29 | 28 | 16 |
| Clang compression | 48 | 48 | 43 | 36 |
| Rust schedule | 29 | 29 | 28 | — |
| Rust compression | 42 | 42 | 41 | — |

The Wasmtime serialization and ordinary Clang C binary are frozen earlier
artifacts, verified against their original checksums and identical guest bytes.
The native C body uses ordinary instructions; no SHA-specific instruction appears
in these loops. It omits VM checks/polling, and no corresponding native Rust
binary is claimed. Scope, instruction obligations and function sizes differ;
these counts do not establish elapsed-time ratios or CPU shares. All inventoried
conditional backward spans remain in the evidence, including the other entries.

The unchanged hot-loop counts show that this block-level pass alone does not
remove the main SHA instruction gap. Next, construct value-version live ranges
and improve allocation/address lowering inside blocks while preserving the new
edge/cold-entry contracts. Code growth remains an explicit cost. Do not replace
these requirements with more public-call batching measurements.

All raw qualification reports, compiler commands/hashes, development failures,
96 correctness observations, publication bytes and static inspection scripts
remain under `artifacts/global-register-plans-20261105`. The compact
[evidence index](data/global-register-plans-2026-11-05.json) records their identities.
No timing cohort was added or retried here. The independent
[hosted routing run](https://github.com/cpunion/wasmvm/actions/runs/41941908335)
still has no assigned runner as of 2026-11-05 01:32 UTC; it tests the earlier
immutable routing candidate, not this source.

No new PR or merge occurs while external PR #2991 remains open. LLGo #2983
stays with its contributor. Final-head four-profile semantics, all 31 packages
per profile, 30 coverage gates, review, controlled competitor/native acceptance,
value-level allocation and the broader application/WASI-world portfolio remain
open. No milestone or full goal is marked complete.
