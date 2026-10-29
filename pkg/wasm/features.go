package wasm

// Features enables individually implemented extensions to the base profile.
// It does not imply support for the complete Core 3.0 profile.
type Features uint64

const (
	// ExtendedConst permits i32/i64 add, sub and mul in constant expressions.
	ExtendedConst Features = 1 << iota
	// TailCall enables return_call and return_call_indirect.
	TailCall
	// MultiMemory enables indexed access to multiple private memories.
	MultiMemory
	// RelaxedSIMD enables all twenty relaxed vector instructions.
	RelaxedSIMD
	// Memory64 enables the proposal's 64-bit memory and table indices.
	Memory64
	// FunctionReferences enables typed, non-null function references and calls.
	FunctionReferences
)

// SupportedFeatures lists extensions implemented by both execution backends.
const SupportedFeatures = ExtendedConst | TailCall | MultiMemory | RelaxedSIMD | Memory64 | FunctionReferences
