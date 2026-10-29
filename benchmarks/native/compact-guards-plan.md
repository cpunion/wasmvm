# Compact span guards and observation-aware immediate selection

Baseline `aa876aff89e9f7449f2d775a0954773338c8c953`, engine `9ef5a12`.
Normalize a memory32 span around its local root before checking overflow and
bound memory length. Use AArch64 immediate forms where representable, retaining
the complete existing fallback for other spans. Underflow, upper wrap and the
exact 4 GiB endpoint must preserve the original checked continuation and trap
order. Do not modify mutable slots while testing a guard.

Separate fusion boundaries from bounded physical-state observation scans.
Immediate/shifted-operand selection may prove a temporary dead across already
proven accesses, but never across another guard, possible trap, poll, CFG edge,
helper or region end. Keep original cold PCs, canonical frame contents and fuel.
Share target lowering and register inventories; introduce no workload-specific
path. Preserve all scalar widths, aliases, reference/vector state and native ABI.

Test cached/spilled operands, encoding limits, span endpoints, wrapping, failed
and interior cold entries, earlier memory effects and future observations.
Run full Go/LLGo release semantics, >=90% measured-package coverage and the final
Linux amd64/macOS arm64 matrix. Freeze source during builds/tests.

Compare exact base/head with identical compiler/reference pins and twelve native
API/dispatch families. Keep preparation, first call and warmed execution separate,
retain regressions and static code hashes, and make no all-target parity claim
from local diagnostics. One active PR and existing merge gates remain unchanged.
