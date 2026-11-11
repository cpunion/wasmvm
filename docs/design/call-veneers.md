# Cross-instance call veneers

The slot compiler recognizes a bounded parameter-only forwarding function as a
call veneer. It contains coalesced local reads, an optional constant table index,
one ordinary direct/indirect call, and an unchanged returned tuple. Arithmetic,
memory operations, dynamic indices, handlers, prompts and branches are excluded.
Analysis examines at most 64 instructions. The immutable plan records the call
and return PCs, immediate operand and original instruction fuel costs.

The native frame driver uses this plan when the function is entered from a
different Core instance. Same-instance calls retain generated native linking.
Root invocations and other functions retain their complete machine artifacts.
The plan lowers call argument binding and frame transitions; it does not dispatch
scalar bytecode or substitute an interpreter for native computation.

The ordinary frame push still reserves all slots and charges logical stack
depth. Reference/vector slots retain their original types and complete bits.
The common call dispatcher reads the current table and checks bounds, nulls and
function subtyping on every invocation. It retains the veneer instance as the
host Caller, including when a callback is parked or reenters another guest.
Host results use the existing destination tuple and return-frame operation.
No target identity or nominal type result is cached.

The driver polls cancellation before each planned transition. Metered execution
charges the exact original prelude/call and return costs. If remaining fuel
cannot cover a folded prelude, the original machine artifact executes the
partial prefix and reports its precise limit/source boundary. Cold entries at
other PCs use that artifact as well. Errors, structured exceptions, suspended
host calls and instance/root retention share the existing explicit frame machine.

The backend must opt in for an owned, live artifact. A missing compilation or
closed lease takes the normal backend error path; a held lease remains valid
after its original Code is retired. Workspace admission, cancellation delivery
and code/memory retirement remain invocation-owned. The plan adds no native ABI
fields, native addresses, new scheduler or workload-specific registration.

General cross-instance functions with computation still use checked runtime
boundaries. Keeping such calls entirely inside machine code requires shared
per-frame instance views, aliased globals, memory ownership and code leases.
Call veneers complete one separately testable optimization within that work;
they do not establish component performance parity or release acceptance.
