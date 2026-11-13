# x86 floating-point entry

Go competitiveness remains the priority. This change removes a measured native
entry cost; complete VM and P3 performance acceptance remain open.

The x86 adapter previously called `fegetenv` and `fesetenv` around each guest
entry, including a separate installation of the default environment. It now
saves the complete 28-byte x87 environment and MXCSR directly, establishes a
fresh x87 environment and MXCSR `0x1f80`, and restores both after the guest
returns. The non-waiting instructions preserve pending host exceptions during
transition. Existing ABI admission and atomic cancellation checks stay ahead
of the transition. Every returned exit takes the same restoration path.

Both x87 and SSE state remain covered; this is not an assumption that host
x87 state can be discarded. The strengthened probe checks exact default guest
control values, complete MXCSR restoration including FZ and supported DAZ,
and all 28 x87 environment bytes. Existing tests cover every exit, invalid
version/arena/depth, cancellation and callee-saved registers. The ARM entry
algorithm is unchanged.

The [Intel architecture manual](https://cdrdv2-public.intel.com/874240/325462-090-sdm-vol-1-2abcd-3abcd-4.pdf)
describes the separate x87 and MXCSR control/status state and their environment
instructions. This change uses those target instructions, with no imported
library implementation.

## Fixed native Linux diagnostic

A GitHub `ubuntu-24.04` x86-64 runner on Intel Xeon 6975P-C, using Clang 22.1.8,
executes the exact old/new adapter sources in one binary per compilation
profile. A volatile function pointer selects the adapter. The guest is an
independent assembly body that increments its result cell and returns; its
state change and every returned status are checked. No startup, compilation
or process preparation is included in the timer.

The fixed plan uses 20 paired blocks, 240 fresh processes total, three unmeasured
batches and seven individually interleaved calls of 250,000 entries. Each
source and direct arithmetic control has an identical-binary alias. The two
C compilation profiles are separate events within the fixed plan, not whole
Go/LLGo executable comparisons. No owned build/test/profile overlaps timing
on the runner. All observations are retained; there is no adaptive extension.

| C profile | Old CPU ns/entry | New CPU ns/entry | Paired new/old, individual 95% interval |
| --- | ---: | ---: | ---: |
| `-O2` | 179.263 | 80.505 | 0.44934 [0.44908, 0.44960] |
| `-O3 -flto=full` | 179.273 | 80.701 | 0.44985 [0.44945, 0.45030] |

The entry path consumes about 55% less CPU in this diagnostic. All six CPU
A/A intervals and all six wall-time A/A intervals resolve the predeclared
±2.5% diagnostic margin. Wall-time ratios agree. Disassembly confirms that
all admission checks remain and the new entry has only the actual guest
call, with no libc floating-environment calls.

The direct assembly control is approximately 1.56 ns/entry. It omits VM
admission and floating-state isolation; it is not an equivalent runtime.
Neither this control nor subtraction from it establishes a pure FP cost.
The adapter also retains its own loop, guest call and required checks.

## Validation and remaining work

Go 1.27.3 and pinned LLGo `56fa9ae -O3 -lto=full` execute the relevant native,
codegen, nativeexec and VM suites on macOS. The exact-source x86 floating and
register probes pass in local Linux Docker; nativeexec and VM also pass.
The broader local native test hits Rosetta's unsupported huge-address
reservation, and the codegen corpus path was not mounted. These are recorded
incomplete outcomes, not successful full Linux qualification. The full
Go/LLGo, Linux/macOS PR matrix and coverage checks remain required.

This boundary improvement does not establish a Go/LLGo API, complete P3 or
competitor throughput win. The PR's existing competitor job reports API
phases separately. Complete P3 remains about twice the Wasmtime CPU cost in
the latest separate macOS screen. The earlier adverse shared-runner Linux
interpreter points also remain open; a 240-process macOS replay did not
reproduce them and cannot close a Linux signal.

[Plan, source/binary identities, controls and raw observations](data/performance-x86-fp-entry-2026-11-13.json)
retain the diagnostic. The temporary measurement workflow is absent from the
production contribution. Next work measures complete native workloads and
continues reducing general call, ownership and scheduling costs.
