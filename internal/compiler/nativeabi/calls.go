package nativeabi

// CallViewOffset is an optional trusted native metadata address. Call frames
// contain only integer IDs/offsets and compiler-owned code/data addresses.
// The Go-side NativeFrames packet never crosses this ABI.
// Each 32-byte caller record starts with a little-endian u32 function/PC pair,
// followed by full-width Used, resume address and destination-plan address.
const CallViewOffset = PayloadOffset + 8

const (
	CallTopOffset          = 0
	CallCapacityOffset     = 8
	CallSlotsOffset        = 16
	CallDepthOffset        = 24
	CallChangedOffset      = 32
	CallHeaderBytes        = 40
	CallFrameBytes         = 32
	CallFunctionOffset     = 0
	CallPCOffset           = 4
	CallUsedOffset         = 8
	CallResumeOffset       = 16
	CallDestinationsOffset = 24
)
