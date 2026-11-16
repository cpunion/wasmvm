package nativeabi

// CallViewOffset is an optional trusted native metadata address. Call frames
// contain only integer IDs/offsets and compiler-owned code/data addresses.
// The Go-side NativeFrames packet never crosses this ABI.
// Each 24-byte caller record starts with a little-endian u32 function/PC pair,
// followed by full-width Used and resume address. Complete destination indices
// immediately precede that resume address in immutable generated code.
const CallViewOffset = PayloadOffset + 8

const (
	CallTopOffset      = 0
	CallCapacityOffset = 8
	CallSlotsOffset    = 16
	CallDepthOffset    = 24
	CallChangedOffset  = 32
	CallHeaderBytes    = 40
	CallFrameBytes     = 24
	CallFunctionOffset = 0
	CallPCOffset       = 4
	CallUsedOffset     = 8
	CallResumeOffset   = 16
)
