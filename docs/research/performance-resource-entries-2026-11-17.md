# Fresh resource entry allocation groups

Grouping fresh canonical resource entries removes **21.75 managed allocations
per complete P3 request** under both Go and LLGo. Go instructions decrease
about 0.2–0.3% with resolving controls. The event does **not** demonstrate a
CPU or throughput improvement and does not update the competitor ranking.

## Mechanism and lifetime

Each table allocates its first entry separately, then consumes fresh slots in
four-entry blocks. Admission and quota checks precede allocation. Slots are
never rebound, including when a guest index is reused: borrowed entries and
cross-table opaque moves retain their stable pointer identities. Guest LIFO
indices, nonreusing host IDs, generations, nominal types and cleanup contracts
are unchanged. Drain and Close discard the allocation cursor; moved entries
remain rooted by their destination. A partial block may retain up to three
retired entry records until rollover or owner teardown. This is allocation
grouping, not capability pooling.

The profile identified `Resources.New` and `AddOpaque` as actual allocation
sites. Wasmtime 48.0.1's [host resource table](https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/crates/wasmtime/src/runtime/component/resource_table.rs)
uses a `Vec<Entry>` and a free-list head; its payloads still use `Box<dyn Any>`.
Our current borrow and transfer machinery requires stable entry pointers, so
fresh groups reduce allocator events without changing that identity contract.
This source comparison does not attribute the entire request gap to entries.

## Fixed complete-request event

Source base is **41de3fe4**, head **034ed786**. Both use Go 1.27.3 or LLGo
**2e3e02f4**, `-O3 -lto=full`, LLVM 22.1.8 and default GC, on the M4 Max macOS
ARM64 host. Six fixed fresh-process symmetric-A/A blocks per payload produce
96 processes total. Four workers, sixteen clients, twenty 512-request windows,
a separate flush and full response/CRC checks retain the existing request
contract. CPU and instruction counters include the driver and verification.
Startup, setup, first call and cleanup are separate. No owned build, test,
profile, edit or analysis overlaps timing; external desktop load is recorded.

Ratios use paired block geometric means and marginal bootstrap 95% intervals.
Both A/A intervals must fit [0.975,1.025] to resolve. No event is extended or
pooled after inspection. Managed counters do not include foreign allocations.

| Managed allocations/request | Base | Head |
| --- | ---: | ---: |
| Go, 16 B | 519.25 | **497.50** |
| Go, 64 KiB | 558.25 | **536.50** |
| LLGo, 16 B | 592.25 | **570.50** |
| LLGo, 64 KiB | 655.25 | **633.50** |

All allocation controls resolve. Go allocated bytes are effectively unchanged;
LLGo reports about 348 fewer BDWGC bytes/request. The 80-byte entry and its
320-byte group request round to 96 and 336 bytes in the installed BDWGC
allocator. Thus 29 entry publications grouped four at a time are consistent
with a 21.75-event and 348-byte reduction. This is allocator granularity, not
evidence of a 16-byte LLGo object header. Go and LLGo byte counters use different
accounting and do not establish a total-memory ranking.

| Head/base | 16 B | 64 KiB |
| --- | --- | --- |
| Go instructions | **0.9971 [0.9947,0.9988]**, controls resolve | **0.9978 [0.9959,0.9996]**, controls resolve |
| Go CPU | 0.9952 [0.9840,1.0066], controls unresolved | 0.9930 [0.9832,1.0028], controls resolve |
| LLGo CPU | 1.0022 [0.9967,1.0084], controls resolve | 0.9914 [0.9800,1.0034], controls resolve |

All throughput controls are unresolved. The resolving CPU intervals include
one; fewer allocations alone are not a demonstrated CPU speedup.

## Rejected transfer refresh experiment

A preceding separate event moved codec transfer setup into a shared helper and
avoided temporary full codecs at two task-resume sites. Request allocation counts
did not decrease. Instrumented semantic probes then confirmed **zero calls** to
both resume sites in this P3 fixture, for both payload sizes. The changed common
initializer is active: Go disassembly shows its former 228-instruction body
becoming a 48-instruction wrapper plus a 204-instruction helper, with a new
non-inlined call. This is concrete additional work on the initialization path,
consistent with the resolving 16 B instruction increase of **0.173%
[0.077%,0.254%]**; it does not isolate every counter difference. CPU did not
demonstrate a gain. The experiment is **not integrated** and remains at
`experiment/canonical-transfer-refresh-20261117` (**bd10cfee**). Timing uses
the same 2e3e02f4 SDK and a distinct fixed event; its results are never pooled
with the entry-group event.

## Validation

The full Go suite and 31 executable coverage gates above 90% pass. Latest LLGo
executes canonical, component and P3 packages; race, vet, canonical/VM no-cgo
and amd64 Docker canonical/component tests pass. Fresh-entry tests retain old
entry pointers across mixed tag reuse and GC, verify guest/host index policy,
check quotas, and keep a moved entry alive after its source closes. Existing
borrow, generation rollback, growth, reentrant drain, destructor failure and
async lifecycle suites remain enabled. Docker is semantic evidence only.

Two existing wall-clock test races were corrected during validation. Global
cancellation now waits for a guest atomic acknowledgement after the write,
without a host callback that would publish globals. Pending source cancellation
waits until its read starts; cold compilation cannot masquerade as missing
cleanup. Original failed logs are retained. A broader no-cgo component selection
also encounters existing unguarded native-only tests; it is not a passing suite.
The supported no-cgo scopes above are reported separately.

Final-head dual-platform CI and unified performance acceptance remain required.
The [data record](data/resource-entries-2026-11-17.json) includes both events'
analyses and provenance hashes. Frozen protocols, raw process records, separate
phase output, source overlays, instruction/allocator probes and validation logs
remain under `artifacts/resource-entry-blocks-20261117` and
`artifacts/canonical-transfer-refresh-20261117`.
