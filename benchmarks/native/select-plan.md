# Branchless scalar select scope

Baseline engine: `a7f2b1e128873a63139d2e02191a16468171810b`, release-profile
head `6d0857066bddf0d92045c29efd1e342271e2970a`. Native scalar `select` currently
branches on the guest condition. Replace it with amd64 CMOV and arm64 CSEL for
integer, floating-point bit patterns and opaque references. Only low i32
condition bits participate. Preserve operand/destination aliases, NaN payloads,
fuel, PC/continuation mappings and canonical scalar storage. V128 keeps its
existing complete implementation.

Run direct scalar/typed-select and alias cases in both native tiers, including
noncanonical upper condition bits, then the complete Go/LLGo Core 2.0 matrix
on amd64/arm64. Keep per-package coverage above 90% and all existing native ABI,
reentry, cancellation and lifetime checks. Check emitted target instructions.

Compare exact source base/head with Go and the same pinned LLGo release flags
against wazero native, Wasmtime and Wasmer. Reuse the pinned Rust packet CRC
(mixed unpredictable selects), C image filter and existing integer/call/memory/
SIMD/FP regressions, retaining first/preparation/warm phases and oracle checks.
Use existing interleaving and diagnostic counts; do not introduce CRC-specific
code or waive the wider application, preparation or ARM calibration deficits.
