# Native interrupt polling scope

Baseline: `a918237374aeb2bc979c94cacb37b5748c5e0b22`. Keep the existing
256-slot-instruction maximum interval and all real control-transfer, call,
fuel, trap and cold-continuation contracts. Remove only redundant unmetered
polls on compiler-generated no-copy jumps to the immediately following PC.
Such an edge is ordinary fallthrough: it neither skips a periodic PC nor
forms a cycle. Retain every periodic poll, backward/nonadjacent transfer,
conditional/table transfer and edge with copies. Keep the precise-fuel tier's
scheduling and charges unchanged.

On amd64, compare the aligned 64-bit control word directly with zero instead
of loading it into a scratch register and then testing it. The atomic width,
cold acquire exit, continuation publication and cached-slot spills stay the
same. ARM64 keeps its existing load/CBNZ pair.

Test cancellation set after the C adapter entry check, including independent
entry at the removed fallthrough poll, forward/backward edges, copied edges,
long chains crossing periodic PCs, live slots and the complete semantic matrix.
Compare the existing 42 public-call/regression/SIMD families against the exact
baseline and pinned references on both targets/compilers. Preserve startup,
compile, first-call and warm phases. Short diagnostics cannot close calibrated
performance acceptance, and this change must not erase slower measurements.
