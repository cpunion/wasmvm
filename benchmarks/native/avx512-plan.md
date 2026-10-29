# OS-validated short-vector AVX-512 scope

Baseline: `36b1d46f64009b78cb4d316ed0c6b92dec7f968c` (production
engine sources are identical to `20d0a41`). Add two explicit independent x86
selection capabilities: AVX512F+VL for packed signed i64 variable shifts, and
AVX512BITALG+VL for byte population count. Keep 128-bit slots, existing register
allocation and one compiler/runtime pipeline. No wider guest vectors, masked
register state, AVX10 assumption or whole-function fallback is introduced.

The platform adapter records CPUID leaf 7 ECX as well as EBX. Both capabilities
require XSAVE/OSXSAVE/AVX and XCR0 XMM/YMM/opmask/ZMM state (bits 1,2,5,6,7),
plus AVX512F and VL. BITALG additionally requires leaf 7 ECX bit 12. Never execute
XGETBV without its existing safe CPUID guard. Missing CPU or OS requirements
must select the older SSE2/SSSE3/SSE4.1/AVX2 legalizations.

Use EVEX.128 VPOPCNTB and VPSRAVQ only. Cross-check encodings against LLVM's
assembler, including upper-register bits and aliases. Force complete SIMD
emission and exact code limits under every known capability mask; execute only
host-advertised masks. Test byte edges, all signed i64 counts, spills, aliases,
cold continuations, precise fuel and cancellation. Run the same full Go/LLGo
Core 2.0 matrix and retain at least 90% coverage per executable package.

Reuse the declared 42-family public-call diagnostic and independent scalar
oracles, with pinned wazero/Wasmtime/Wasmer and exact own baseline. Popcount and
i64 shifts are the affected kernels; all other families remain regressions.
Keep all startup/compile/instantiate/first/warm phases and selected CPU masks.
Native amd64 evidence must show the capability was actually selected; emission
alone is not execution. ARM64 selection is unchanged. Short diagnostics do not
close calibrated acceptance, M1/M2 or PR #26.

Primary instruction reference: Intel SDM Volume 2,
https://cdrdv2-public.intel.com/835757/325383-sdm-vol-2abcd.pdf.
The pinned Cranelift 48.0.1 rule explicitly selects VPOPCNTB with VL/BITALG:
https://github.com/bytecodealliance/wasmtime/blob/v48.0.1/cranelift/codegen/src/isa/x64/lower.isle#L2456.
This source establishes an available competitor path; actual code selection on
each benchmark host is a separate observation.
