package nativeabi

// CallViewOffset is an optional trusted native metadata address. Call frames
// contain only integer IDs/offsets and compiler-owned code/data addresses.
// The Go-side NativeFrames packet never crosses this ABI.
const CallViewOffset = PayloadOffset + 8

const (
	CallTopOffset          = 0
	CallCapacityOffset     = 8
	CallSlotsOffset        = 16
	CallDepthOffset        = 24
	CallChangedOffset      = 32
	CallHeaderBytes        = 40
	CallFrameBytes         = 64
	CallFunctionOffset     = 0
	CallUsedOffset         = 8
	CallPCOffset           = 16
	CallResumeOffset       = 24
	CallDestinationsOffset = 32
)
