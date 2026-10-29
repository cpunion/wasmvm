# Native scalar memory development scope

Status: development branch; no public backend or merge acceptance claim.

Implement all 23 Core 2.0 scalar loads/stores, memory.size, and memory.grow's
explicit exit/resume contract on amd64 and arm64. Floating memory accesses preserve
raw IEEE bits. Address addition is widened before a complete range check; failed
stores have no partial effect. Fuel, exact trap locations, cancellation and every
original continuation remain observable. SIMD memory and bulk operations belong
to subsequent complete profile work, not this scalar capability claim.

ABI version 2 appends a native-owned memory base and byte length. A reusable
platform Memory owner allocates zeroed off-heap storage, bounds growth, excludes
host access/growth while a bound context executes, and retains storage until all
bindings and active entries drain. No Go pointer or growth-sensitive view escapes.
The current Core 2.0 compiler caches one memory's base/length during an entry and
reloads them on every resume. The later Store driver supplies shared/imported
ownership, aggregate quotas and growth results through this same boundary.

Before implementation acceptance, compare Go/LLGo candidates, the own baseline
(or explicit native unavailability), and pinned wazero/Wasmtime/Wasmer using the
same parameterized binaries. Preserve integer and scalar numeric regressions.
Use the existing memory-load-store fixture plus declared unaligned/narrow access
and dependent-address workloads; separate setup, first call, warm execution and
growth. Use independent checksums and alternate inputs. Require parity or better
for every affected family on both required targets with calibrated A/A controls.
An integer-only result cannot accept memory access. Local diagnostics are advisory.


## Development evidence (2026-10-27)

All 23 scalar accesses, size and growth exits execute on both ISAs. The shared
native corpus now executes 12,068 assertions per tier across 13 complete scripts /
27 modules, excluding 347 frontend-only commands. Targeted tests additionally
compare all access widths and offsets against the owned semantic memory model,
including independent continuations, full memory effects on traps, FP bit
preservation, register pressure, relocation after growth, and owner retirement.
The complete load/grow/trap scripts still require calls, globals or import/linking;
they are not claimed as complete native conformance.

The first mixed-access diagnostic was 1.279/1.273 ns per iteration (Go/LLGo).
Local bounded address proofs, indexed addressing, shared dead-copy removal,
immediate selection and immutable arm64 constant registers first reduced it to
about 0.57. The follow-up combines mask/shift address calculations, forwards
integer comparison flags, removes a redundant pre-tee constant, and compacts
arm64 branch reservations. These are shared compiler transformations, with
independent cold entries and unchanged cancellation/fuel behavior. No guard-page signal recovery or competitor fallback is involved.

Three randomized process blocks on local darwin/arm64 (M4 Max) produced the
following advisory medians after those transformations, in ns per source
iteration. Base is exact commit `adaa507`; both base and candidate execute native
memory code. Startup, compilation, instantiation and first-call records remain
separate. All workloads include their declared deterministic reset cost. Go is
1.27.3; LLGo uses the existing pinned runtime contribution. References are wazero
1.12.0, Wasmtime 48.0.1 and Wasmer 7.3.0.

| Workload | Base Go / LLGo | Candidate Go / LLGo | wazero | Wasmtime | Wasmer |
| --- | ---: | ---: | ---: | ---: | ---: |
| Mixed i32 load/store | 0.600 / 0.572 | 0.543 / 0.559 | 0.837 | 0.480 | 0.513 |
| Mixed, sustained | 0.558 / 0.553 | 0.532 / 0.565 | 0.806 | 0.493 | 0.533 |
| Narrow, unaligned | 0.835 / 0.787 | 0.743 / 0.777 | 0.982 | 0.733 | 0.685 |
| Dependent addresses | 2.078 / 2.112 | 2.118 / 2.167 | 3.069 | 2.130 | 2.150 |
| Unaligned f64 | 0.775 / 0.765 | 0.757 / 0.718 | 0.787 | 0.710 | 0.725 |

This local host is not isolated: variations are too large to establish small
regressions or parity. Candidate code hashes match across Go/LLGo for all five
rows. Code size grows by 8–9% versus `adaa507` because independent branch
continuations retain their own checked paths; hot-path compaction does not imply
smaller total functions. Earlier pre-optimization observations are retained.

The 64-page growth fixture reports first-call time separately; warm calls measure
quota-rejected growth and must not be normalized as allocation throughput. First
calls are short and noisy, so these observations establish no growth advantage.
Four alternate seed/count pairs per new fixture pass in every engine (80 checks).
The ten scalar numeric regression families also retain independent checksums and
competitor rows. Raw observations are kept in `artifacts/native-memory/` locally.

**Acceptance remains blocked.** Mixed and narrow access still trail Cranelift;
nearby results have no calibrated parity claim. Required native amd64 performance,
Go/LLGo final-head CI, controlled A/A and all affected portfolio gates remain open.
The earlier baseline `b86ac60` has no executable native memory support; it exits to an
unimplemented helper, so it is reported as unavailable rather than substituted
with an interpreter. PR #26 remains the only open PR; this capability stays on a
development branch until its own performance and semantic gates are ready.


Go and LLGo full repository tests pass locally. Go race, strict cgo pointer
checks, no-cgo build and vet pass; Linux amd64 codegen/platform/ABI execution
also passes under Docker/Rosetta (semantic evidence only, not native amd64 timing).
Measured statement coverage is codegen 97.7%, platform/native 96.2% and nativeabi
100.0%. The untested platform paths include allocation failure and unsupported
host mechanisms; coverage is not a full-backend conformance claim.

Regenerate the four owned fixtures with
`python3 benchmarks/native/memory_fixtures.py` using pinned wasm-tools 1.258.0.
Their manifest retains source/binary digests and independent Python checksums.
Use the existing readiness/phase protocol and the same fixture bytes in every
runner. Do not combine the growth-first-call metric with warm-loop throughput.


The follow-up also compares all ten numeric families and a 25-million-iteration
integer chain against `adaa507` and all three references. All checksums pass;
Go/LLGo candidate code hashes match for all 11 workloads. Timing is inconclusive:
even identical generated code has large compiler/process-to-process differences
on this shared host. No numeric regression or parity acceptance is inferred.
The earlier eight-block integer check against `b86ac60` is retained alongside
these newer observations, not substituted for them. Alternate-input execution
passes 150 numeric and 80 memory checks across the five engines.
