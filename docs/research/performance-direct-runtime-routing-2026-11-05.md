# Direct native routing and portable primitives: 2026-11-05

**Local semantic qualification passes; performance acceptance remains open.**
The [hosted comparison](https://github.com/cpunion/wasmvm/actions/runs/41941908335)
is queued as of 2026-11-05 00:35 UTC. There are no new accepted timing results.

Engine source: `59a76f933901e56266e618bff0a502364eca0350`. Hosted candidate:
`c92383e2b60fe75910ca921786c0288f49a11b12`. Hosted baseline:
`a47055833582068b88e9aeb51d922ef84662db4f`, the earlier fully CI-qualified
register-plan source, not a merged production release. Main remains
`4ce79899a1f3ed45163afa622cc408cc6a79a599`.

The source changes two production files after the held
[runtime-effects experiment](performance-runtime-effects-2026-11-05.md).
The native driver selects shared GC/SIMD/call/throw/return leaves directly,
avoiding the general dispatcher frame before their leaf frame. Portable SIMD,
nonallocating GC access and atomic/bulk operand routing remains direct, using
the existing canonical memory/state/SIMD operations. Shared complex effects,
source PCs, exact fuel, polls, roots, quota checks, owned arenas, frame
transfers and interruption retain their contracts. Scalar native fallback
continues to fail. No ISA, compiler pin or padding/alignment sweep occurs.

Ten complete affected packages pass pinned Go 1.27.3 and LLGo 3571
(`-O3 -lto=full`) on macOS arm64: 108 actual conformance reports, scoped Go
coverage 90.71%–100%, execute 92.16% (1046/1135). Race, cgocheck2 and no-cgo
checks pass. A dry build of the new hosted pipeline qualifies 274 source/reference
configurations across 22 fixtures; 12 GC/reference combinations retain explicit
unsupported results. These dry-run timings are not used as performance data.

The new opt-in `routing` workflow builds both source/compiler pairs and pinned
references once, then reuses that immutable artifact across five sequential
Linux amd64 jobs: Core, memory/SIMD, GC, Threads and C/Rust SHA. Every job has a
10-minute bound. C/Rust sources, exact guest bytes, producer provenance and the
independent SHA oracle are committed unchanged from the earlier application
experiment. All participants use the same bytes per fixture.

Each group predeclares five fresh process blocks, two warmups/seven serially
interleaved samples, base/head LLGo A/A for both backends, one inherited CPU
affinity and `GOMAXPROCS=1`. Each marginal 95% A/A interval must lie inside
±2.5% for this diagnostic calibration. That does not replace the controlled
milestone gate, simultaneous feature intervals, absolute effect floors or
required platform portfolio. Hosted CPU power and physical neighbor contention
remain uncontrolled; no registered self-hosted runner is currently available.
Do not pool independent groups or use component/unchecked-C costs to waive
complete API outcomes.

Both older noisy W results and failed setup attempts remain recorded. The new
event is a different source architecture and environment, not a favorable retry.
No new PR or merge is performed while external PR #2991 remains open; upstream
merge is prohibited by owner policy. LLGo #2983 stays with its contributor.
Final-head four-profile semantics, all 31 packages/profile, 30 coverage gates,
review, controlled competitor/native acceptance, global allocation and the
broader application/WASI-world portfolio remain open. No milestone or goal is
marked complete.

The follow-up [public-call allocation diagnosis](performance-public-allocation-2026-11-05.md)
retains both a negative aggregate-passing result and the remaining Go/LLGo
allocation-byte gap. It changes no engine source or hosted candidate and makes
no timing or compiler-defect claim from allocation totals alone.
