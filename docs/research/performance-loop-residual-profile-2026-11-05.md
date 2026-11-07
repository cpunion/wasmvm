# Native loop residual profile: 2026-11-05

Qualitative CPU samples and disassembly confirm that loop location plans remove repeated canonical-frame traffic from both compiled C SHA loops. This investigates the remaining native gap; it does not replace the [fixed throughput comparison](performance-loop-register-plans-2026-11-05.md).

The exact measured O runners (`d472a76ff9a8aebfdacd6ba74eeccc962d404e59`) run the pinned C guest at 65,536 blocks/call, two warmups and 400 samples. Go and LLGo binary hashes are checked before execution. Each receives one three-second CPU sample and a VM map; checksums, raw stacks and command outcomes are retained. This different invocation scope is diagnostic only. Samples across host threads are aggregated by native address; their totals are not host throughput ratios.

Separately archived builds add only an output hook to native code publication. They use the original pinned compilers and flags; original engine hashes are checked before instrumentation. Go and LLGo publish identical **100,440 bytes**, SHA-256 `51fd7e538050e9b1314ffbb02d645432f83f3c95b249428a98b23f2f405eb487`, with entries `[0, 13816, 25288, 73496]`. These are diagnostic publication bytes, not a remote byte read from the previously measured process. Its randomized arena is located by VMMap before mapping sampled PCs to the same-source layout.

The private backward conditional branches delimit the two guarded C loops. Static instruction counts include their cancellation checks and backedges, excluding one-time guards and exit transfers:

| Loop | N instructions → O | N frame loads/stores → O | Interrupt loads per iteration |
| --- | ---: | ---: | ---: |
| Schedule expansion | 33 → 29 | 1/0 → 0/0 | 1 → 1 |
| Compression rounds | 59 → 48 | 5/4 → 1/0 | 1 → 1 |

Canonical stores for displaced homes now follow the false backedge rather than executing every iteration. The second loop still loads an invariant frame word and contains multiple register copies; both loops materialize their bound and boolean result. Sampled PCs cluster near loop control and polling, but PC skid and sampling bias prevent attributing that time to any individual instruction or declaring polling the dominant cost.

The next allocation work must model changing locations and copy relationships across instruction ranges, with spills or reconstruction for every canonical snapshot. Preserve all raw scalar/vector words, reference roots, cold/interior entries, deferred exits and interruption semantics. Full global live-range splitting/linear scan remains open. No polling removal, runtime-wrapper substitution, extra timing cohort or new LLGo issue results from this profile; #2983 is handled by another contributor.

[Lossless diagnostic evidence](evidence/performance-loop-residual-profile-2026-11-05.json.xz) retains raw profiles/maps, exact binary/source identities, output-hook build records, actual published bytes, decoded hot windows, branch-delimited loop comparisons and the preceding N publication inputs. The frozen primary throughput evidence is unchanged.
