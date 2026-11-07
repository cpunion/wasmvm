# CFG value allocation: 2026-11-05

**The initial ordinary mapping candidate is rejected: both SHA applications
regress with resolved diagnostic controls. Local Go/LLGo semantics pass;
full target CI, review and competitive performance acceptance remain required.**
Initial engine `c86b15d80726cc3f88456cd4ac9245c17a09b28c`; baseline
`97667b6949824fceff36d2d34063b1e060423b50` (engine `55976a4`).
This event is separate from the preceding
[typed preparation](performance-typed-compiler-ordering-2026-11-05.md) and
[hosted comparison](performance-loop-nonresident-values-2026-11-05.md).

## Shared mechanism

The optional analysis reconstructs write versions and entry phis across the
complete slot CFG. A flat pending list resolves predecessor dependencies,
including cycles and irreducible control, without recursion or a Cartesian
fixed-point sweep. Ordered edge moves retain the scheduler's scratch versions;
call/GC/multi-result tuples retain their real operands and outputs. Driver,
exception and prompt entries remain opaque. Cell and input-proportional tuple
budgets reject incomplete analysis and retain the conservative slot planner.

Value intervals feed the existing integer and shared FP/vector register banks.
Overlapping versions of one canonical slot are merged. Ordinary cold mappings
stay stable around CFG strongly connected components. Proven, nontrapping
private loops may split unshared slots at actual value boundaries, using the
same bank allocator, selectors, transition stubs and physical spills. Existing
equal-copy bundles keep their home. The selection heuristic includes predicted
frame accesses and all entry, split, exit and backedge transitions; this is not
an elapsed-performance gate.

Every displaced raw frame word is still stored, including unused temporaries.
The original cold PCs, precise tier, roots, trap order and bounded polls remain
available. Changed backedge mappings use the existing allocation-edge stubs.
Borrowing the default-memory length register is tracked across the entire
private loop, so a later plan cannot suppress its restoration. Selection also
stops before a planned native region: preceding literal/copy fusion cannot
consume its entry. There is no workload-specific opcode path or second backend.

## Execution evidence

The complete codegen/nativeexec/public VM suites pass on macOS arm64 with Go
1.27.3 and LLGo `3571c25` (`-O3 -lto=full`). Go vet passes. The three suites
cross-cover **codegen 97.39% (8094/8311)** and **nativeexec 92.54% (459/496)**;
these are exact package measurements, not specification coverage. Build/test
wall times are 11.01 s for Go and 63.04 s for LLGo, not guest benchmark results.

New tests cover branch/join/backedge phis, scheduled copy cycles, tuple outputs,
opaque prompt resumes, budget rejection and a 10,000-block predecessor chain.
High-pressure integer, F32/F64 and V128 loops select split plans on both target
compilers. An independent model checks every raw frame word and original cold
PC, both ordinary and PC-zero headers, the first generated private
backedge poll, last-valid memory accesses and late out-of-bounds traps. Both
targets compile; the local host executes arm64. Development failures before
SCC stability and native-entry boundaries were fixed remain in the local inputs.

## Actual publication

All 66 baseline-Go/head-Go/head-LLGo native publication witnesses pass their
independent result oracles and pinned identities. All **22 candidate Go/LLGo
images and entry-offset lists match byte for byte**. Twenty baseline/head
images are unchanged. The two application images grow:

| Actual arm64 published module | Baseline bytes | Candidate bytes |
| --- | ---: | ---: |
| SHA-256 / Clang | 101308 | 106692 |
| SHA-256 / Rust | 101476 | 107040 |

The larger application images are retained as an adverse size result. Image
identity does not establish host-boundary or preparation-time identity, and
neither image size nor the analytical frame-cost heuristic proves throughput.

## Initial comparison and rejection

The complete original-count macOS arm64 comparison retains 22 fixtures, 362
alternative-correctness configurations, 12 explicit unsupported GC reference
configurations, five fresh processes and seven interleaved calls. All phases
and adverse results are retained. **5/22 workloads resolve all four diagnostic
A/A controls**; local conditions remain uncontrolled.

Both SHA application controls resolve. LLGo native head/base is **1.133×** for
Clang and **1.154×** for Rust; Go native is **1.134×/1.145×**. The ordinary
value-endpoint maps therefore fail this development comparison. Both x86 SHA
compression functions retain two native loops and select **zero value-split
loops**. Ordinary allocation boundaries increase from 12 to 63; compression
function bytes grow 10956→13977 and 10866→13859. The data does not support
claiming a private-loop benefit from the new analysis.

Twenty unchanged guest images do not explain their different warm timestamps.
Their adverse observations, including LLGo core memory 1.424× baseline, remain
inconclusive when calibration fails. Neither a new LLGo defect nor layout
causality is inferred. Compiler preparation remains separately recorded.

The follow-up retains structural basic-block boundaries for ordinary mapping,
reserving fine splits for private regions, and compares frequency-weighted
frame accesses plus entry/fallthrough/branch/exit transfers against the existing
slot plan. Candidate publication and performance must be checked separately;
a cost heuristic cannot turn this rejected result into a pass.

Initial-source bundle: `artifacts/cfg-value-allocation-20261105/evidence.tar.gz`, 3724253 bytes, 219 files; SHA-256 `f7171484f1822600d839838162fad977196199617f96a99bef450b01671d3076`.

## Costed follow-up

Source `3355a8d5293198474c8cd97379317093f9e983ae` retains structural
basic-block boundaries for ordinary mappings. A complete-CFG heuristic counts
weighted real frame accesses, entry/fallthrough/branch/exit transfers and
shuffle instructions against the existing conservative slot plan. Ties retain
the existing plan. Fine value splits remain available in private loops.

Complete codegen/nativeexec/public VM tests and vet pass again under Go and
LLGo release flags. Exact cross-coverage is **codegen 97.66% (8140/8335)** and
**nativeexec 92.54% (459/496)**. Additional tests reject zero-miss plans whose
shuffles outweigh their saved loads, retain profitable added residency and
charge taken-backedge frequency.

All **22 actual arm64 baseline/candidate images and entry-offset lists now
match**, independently under both host compilers. All 66 publication witnesses
pass their result oracles. This removes the extra application bytes. These
fixtures select no new private value split, so their unchanged guest code does
not demonstrate throughput improvement from SSA analysis. Additional host
preparation work remains a cost to measure.

The follow-up scope was fixed before collecting timings: core/API boundary
and application groups, plus any other group with changed actual guest images.
No other group changes, so it covers six workloads. Exact-source baseline and
reference binaries are reused only after source, flag and binary-digest checks;
both candidate binaries are fresh. These observations remain separate from
the initial full 22-case event and from hosted x86 measurements.

The six-workload follow-up passes all 102 alternative checks. **3/6 workloads
resolve all four warm A/A controls.** LLGo native head/base is 1.000× for Clang
SHA (inconclusive) and 1.006× for Rust SHA (resolved, advisory). Neither proves
an elapsed improvement. The complete comparison below is observed time divided
by the reference time; lower is better. Warm controls do not certify preparation
latency.

| Workload | LLGo / Go native | LLGo / Go interpreter | LLGo native / wazero JIT | / Wasmtime | / Wasmer | LLGo interpreter / wazero interpreter | / WasmEdge interpreter | LLGo native / C | A/A |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| core/call-control | 1.114× | 1.140× | 8.815× | 5.172× | 1.661× | 3.825× | 1.083× | unavailable | inconclusive |
| core/integer-loop | 0.991× | 0.851× | 0.925× | 1.003× | 1.001× | 0.367× | 0.619× | 1.003× | inconclusive |
| core/function-calls | 0.965× | 0.920× | 0.536× | 0.855× | 0.994× | 0.568× | 0.821× | 0.991× | resolved, advisory |
| core/memory-load-store | 0.949× | 0.919× | 0.501× | 0.908× | 0.814× | 0.475× | 0.750× | 1.604× | resolved, advisory |
| applications/sha256-clang | 0.989× | 0.919× | 0.694× | 1.093× | 1.074× | 0.462× | 0.721× | 1.243× | inconclusive |
| applications/sha256-rust | 0.981× | 0.925× | 0.635× | 1.119× | 1.112× | 0.460× | 0.766× | 1.270× | resolved, advisory |

Clang/Rust SHA compileValidate medians are 2804.63/2725.92 µs under Go and
5690.50/5671.17 µs under LLGo, against 1766.58/1747.25 and 4163.67/4263.46 µs
at the baseline. LLGo/Go is observed near 2.0× and LLGo/wazero preparation near
18.7×. These preparation observations have no preparation-specific calibration;
they remain adverse evidence, not a new upstream compiler-defect diagnosis.

Fresh x86 compilation against the exact baseline restores byte-identical
compression functions at 10956/10866 bytes and 12 allocation blocks. Their
wrappers shrink 15982→15889 and 16336→16199 bytes; no x86 guest executes on the
arm64 host. No SHA application selects a new value-split private loop.

Follow-up bundle: `artifacts/cfg-value-allocation-20261105/costed/evidence.tar.gz`, 2463823 bytes, 230 files; SHA-256 `94d39ebb4ed8b61fe5ed711655fb3f0ebe223dc4d8d6e6ba48f1a4b7a2fec1d7`.

## Outstanding gates

Full final-head Go/LLGo target CI and contribution review remain required.
Integration follows the one-active-contribution policy while LLGo #2991 is
open. LLGo #2983 remains with its existing contributor. General frame
reconstruction/rematerialization, helper/GC/shared-memory and host-boundary
costs, broader WASI-world qualification and release acceptance remain open.
This increment does not complete the performance goal or a milestone.

The [compact data index](data/cfg-value-allocation-2026-11-05.json) links both source events, exact coverage, all workload medians/calibration and the complete archived phase records.
