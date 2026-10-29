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
)

// SupportedFeatures lists extensions implemented by both execution backends.
const SupportedFeatures = ExtendedConst | TailCall | MultiMemory
