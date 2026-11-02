# Scalar host functions

Status: implementation candidate; final integration and performance acceptance
remain open. The implementation extends the existing VM host-request driver.
It does not add another execution framework or change Wasm instruction support.

`Store.NewScalarFunction` and `Caller.NewScalarFunction` accept the ordinary
FunctionType and a ScalarHostFunction callback. Only I32, I64, F32 and F64
parameters/results are accepted; references, vectors and invalid types return
ErrType at registration. Optional function-reference types use the existing
feature/type-domain checks. NewFunction retains its reference and V128 contracts.

Each callback receives an independent []uint64 with length and capacity equal
to max(parameter count, result count). Parameters begin at index zero; results
replace its beginning. Integer and floating-point values use their exact Wasm
bits. I32/F32 inputs and outputs are masked to 32 bits. Result-only entries start
at zero. The callback may retain its slice: it never aliases guest slots or
another invocation, and result values are copied before guest resumption. The
stack contains no Go references or guest handles and needs no reference roots.

The ordinary public Call API, signature validation, frame/fuel limits and
workspace acquisition remain shared. Host dispatch uses the same fresh Caller,
context, reentry protocol, temporary-reference roots and panic cleanup. A Caller
expires before continuation resumption, including error/panic paths; it is not
pooled. Structured exceptions use the existing host-error conversion. A scalar
function is synchronous, like HostFunction; both ordinary calls and continuation
polls dispatch through the same operation.host path. SuspendableHostFunction
retains its separate existing suspension contract.

P1 registers all 46 imports through this interface. Their existing signatures
contain only scalar integers. Its Process admission, merged cancellation context,
ABI operations, checked memory access, errno/exit behavior and shutdown protocol
remain shared. No import-name-specific engine shortcut is introduced.

The implementation still allocates the independent scalar stack and raw result
slice. It removes public Value construction/root conversion and the P1 argument
and result conversions; it does not claim zero-allocation callbacks or complete
host-boundary performance. Compare all original P1 kernels, existing public-host
controls, both compilers/backends, baseline A/A aliases, explicitly identified
reference cancellation policies and ordinary C. Keep every original regression,
preparation/first/warm boundary and unsupported policy comparison visible.
