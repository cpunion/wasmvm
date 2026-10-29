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
	// GC enables recursive/subtype definitions and the WasmGC instruction set.
	GC
	// ExceptionHandling enables final tags, try_table, throw and exception references.
	ExceptionHandling
	// Core3Rules enables final limit encodings and defined immutable globals.
	// Core3Features selects the complete profile, including its prerequisites.
	Core3Rules
)

// Core3Features selects the full pinned Core 3.0 binary profile on both backends.
// Core Threads and Component Model features are separate.
const Core3Features = ExtendedConst | TailCall | MultiMemory | RelaxedSIMD | Memory64 | FunctionReferences | GC | ExceptionHandling | Core3Rules

// SupportedFeatures includes all accepted core feature flags.
const SupportedFeatures = Core3Features
