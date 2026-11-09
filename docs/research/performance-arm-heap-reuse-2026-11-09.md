# ARM64 heap-cache reuse priorities

Baseline: `6923e034b42f89e93315121eed5941315a9fd659` (PR #137).
The separate countdown experiment (PR #138) was withdrawn without merge; it is
not part of this candidate. Controlled performance and release remain open.

## Mechanism

The ARM64 cache planner formerly ranked types only by access frequency. Under
its fixed caller-register budget, equally frequent types could prefer changing
object identities over one invariant array root. This reduces hit rates and can
leave repeated array identity/bounds work in the loop.

Reuse the existing conservative local/global-origin hint already used by amd64.
It gives an ARM64 frequency candidate an additional reuse weight only when the
same complete identity is expected across a loop. Mutable globals, changing or
mixed roots, joins, tuple moves, calls and protected regions retain the ordinary
frequency rule. These are profitability hints, never validity proofs.

The same maximum two pairs from the existing eight caller registers remain.
Every hit still checks the full handle; every miss retains null, index,
generation and array-bounds checks. Collection/reentry still resets keys, and
ordinary cold continuations are unchanged. No ABI, bank, unchecked path or
workload-name selection is introduced. The amd64 weighting rule is unchanged;
its shared analysis moves to a target-neutral helper.

## Qualification and declared measurements

Added cases exercise invariant/mutable/mixed roots, local mutation, moves,
joins, loop calls, protected regions and the last available caller-register
pair. Existing native heap tests retain full handles, packed/vector values,
every cold PC, allocation/collection and array-bounds behavior. Complete affected
Go/race/vet and pinned LLGo release tests precede timing. Final four-profile
semantic CI, coverage, stable consumer and extracted CLI package verification
remain required for merge.

Before timing, a fixed-profile compile-only audit checks all22 portfolio images
for both targets. The expected scope is unchanged amd64 images and ARM64 GC
indirect access; investigate any other change before choosing a broader scope.
Use the existing GC comparison: original4 and separate sustained4 (300000
allocation,1000000 other iterations), all named references and explicit
unsupported probes, on hosted Linux amd64 and local macOS arm64. No measurement
from the abandoned countdown source is a baseline event for this candidate.

Freeze source and finish every local build/test/probe before timing. Keep
Go1.27.3, LLGo38de1e218 `-O3 -lto=full`, LLVM22.1.8, fixture/reference hashes,
five fresh interleaved blocks,2warmups/7calls and alternative checksums fixed.
Retain all phases, raw samples/orders, allocation means including every pause,
and every adverse/unresolved A/A interval. Do not sweep counts/profiles/ISA,
retime unchanged sources, subtract phases or pool hosts/cohorts. These advisory
measurements do not close controlled M1–M6 performance or production publication.
