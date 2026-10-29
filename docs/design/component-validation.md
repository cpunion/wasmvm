# Component type validation

The component decoder produces unvalidated, ordered definitions. The
`internal/component/validate` type pool resolves value/function/resource types,
component and instance type declarations, lexical aliases and Core module
types. It never executes a start function or calls a resource destructor.
This is a type subsystem, not yet a complete binary component validator or
runtime; composition, canonical declarations, exported resource visibility,
interface subtyping and resource-binder substitution still need integration.

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
partially populated scope. Interface equality currently returns unsupported
until binder-aware linking is connected.
