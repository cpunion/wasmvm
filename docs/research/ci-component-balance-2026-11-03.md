# LLGo component shard balance: 2026-11-03

PR #102 head `f384fb7`, job 125417236719, timed out at the 12-minute job
limit. All semantic tests and evidence-upload steps succeeded before overall
cancellation. This is an actual execution timeout, distinct from the earlier
macOS runner-capacity cancellations. LLGo setup takes 164 seconds; the test
command takes 525 seconds, and teardown brings the job to 12m17.

The retained trace assigns only 38.94 seconds to the five test processes.
The full-LTO links take 81.50/90.35/112.56/154.82/262.33 seconds, overlapping
under the driver's three-package parallelism. CLI linking starts at 251.43s
and ends at 513.77s; the other component packages finish by 306.54s. Link,
frontend and execution spans remain separate; their overlapping durations
must not be added into a fabricated serial total.

Rebalance the existing six semantic jobs: component keeps its four component
packages and only the component corpus; move internal/cli to the former
single-package legacy-thread shard, renamed entry. Entry prepares the same
six corpora previously required by CLI, and can restore the complete CLI
package cache before its broad fallback. No compiler flags, timeouts, package
tests or semantic cases are removed; no jobs are added. The plan validator
still covers all 31 test packages exactly once, and all five validator tests
pass. Runtime production-CLI smoke checks remain independently required.

This removes CLI's late expensive link from the overloaded component group.
Both rebalanced target jobs must demonstrate their actual durations in final
CI; no estimated improvement closes the 5–10 minute job target.

[Timeout annotations, complete per-package trace, exact plan and validation](evidence/ci-component-balance-2026-11-03.json.gz)
