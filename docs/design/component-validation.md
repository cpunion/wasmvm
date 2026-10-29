# Component type validation

The component decoder produces unvalidated, ordered definitions. The
`internal/component/validate` type pool resolves value/function/resource types,
component and instance type declarations, lexical aliases and Core module
types. It never executes a start function or calls a resource destructor.
The ordered component checker now validates embedded Core modules and their
instances, all five Core external sorts, aliases, imports/exports and synchronous
canonical function/resource declarations. Component instances support instantiation
and synthesized exports, binder-aware interface subtyping and export ascription.
It shares Core import matching with the runtime. Value/start definitions and
most async canonical declarations still return unsupported. This is not a
complete binary component validator or runtime.

Each lexical scope starts with empty component and Core type index spaces.
Equality aliases preserve identity, while resource bounds introduce fresh
compilation identities. Crossing a component boundary cannot capture a free
resource through an outer type alias. Resources bound inside an interface type
are distinguished from its free lexical resource captures. Runtime
instantiation must generate fresh resource identities for each component
instance. Instance-typed declarations already freshen abstract export binders,
preserve equality aliases and substitute dependent value/function types and
nested component captures. Repeated imports of one interface cannot share an
abstract resource accidentally.

Value graphs retain sharing and declaration order. Validation checks value and
resource sorts, nonempty aggregate types, flag counts, label syntax/strong
uniqueness, map keys, fixed-list lengths, Core representation/destructor
signatures and transitive borrow restrictions. Function results and async
elements cannot contain borrows. Every value type must occupy less than
2^28 bytes under the memory64 Canonical ABI. Layouts for future/stream/error
context handles and fixed lists/maps do not imply conversion or async support.

Core type groups and external descriptors use the existing Core validator and
one compilation-local canonical registry. Groups validate incrementally rather
than rescanning every preceding type. Outer aliases retain canonical identity;
their reference family is resolved before checking subsequent declarations.
Core module types reject invalid descriptors, duplicate imports/exports and
aliases to other module types. None of this mutates the decoded definition.

The pool bounds aggregate metadata, traversal work and lexical/value nesting. Equality visits
shared value pairs once and bounds total comparison work. Type graphs and
published metadata are immutable; one pool belongs to one compilation.
Allocation budgets and nominal IDs are not speculative transactions: discard
the compilation after a failed declaration rather than retrying against a
partially populated scope. Linking uses a separate bounded relation: component
imports are contravariant, exports are covariant and function value types are
invariant. Abstract resource binders are substituted across arguments; repeated
instantiations generate distinct local resources while equality aliases and
imported identities remain intact. Representation authority is tracked by
nominal identity in the defining component, independently of public name tokens.

## Boundary names and canonical plans

Structural equality does not confer an external type name. Publication creates
a new binding for records, variants, enums, flags and resources without naming
the original private binding. Resource nameability and nominal resource identity
are separate. Function imports cannot use export-only names. Instance type
bodies defer nameability until used at a boundary; component type bodies check
it eagerly. Aliases across a real component boundary do not inherit the parent's
visible names. Traversal of shared interfaces consumes work budget without
recursively expanding duplicate paths.

Canonical plans check duplicate options, memory/address width, allocator and
post-return signatures, direct/spilled arguments/results, memory/realloc
requirements and resource representation authority. Shared memories cannot be
canonical memories. These checks never invoke allocators, post-return functions
or guest start functions; actual calls, transfer transactions and lockdown after
traps remain runtime integration work.

The offline corpus includes 13 complete generated semantic scripts and checks
461 binary commands (361 invalid rejections and 100 accepted definitions). Five
text-parser assertions are reported separately and do not count as binary
validation passes. See the [corpus inventory](../../testdata/component-binary/README.md)
for the pinned version-suffix producer limitation. A few async builtins have
static signature/index plans to cover the complete `indicies` script; these
plans do not execute async work. Component execution, canonical calls and WASI
worlds remain separate deliverables.
