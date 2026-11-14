package state

import "github.com/cpunion/wasmvm/internal/engine/value"

// NativeFrame describes an ordinary guest frame at a native-to-runtime exit.
// Parents publish both the call PC and its continuation; the final frame's PC
// identifies the already charged instruction whose runtime effect is pending.
// These are host-side descriptions, never records passed through the C ABI.
type NativeFrame struct {
	Function *Function
	Base     int
	PC       uint32
	CallPC   uint32
}

// NativeFrames is borrowed by a native backend for one synchronous run. Inputs
// describe the existing top frame and current invocation's remaining limits.
// A backend may update PC, Fuel and Values. Frames, when nonempty, replaces the
// top frame and Parents existing parents with a complete same-instance chain,
// including its root. Parents is zero unless Resume authenticates a previously
// published chain that survived an ordinary host return without stack changes.
// A resumed backend may preserve an unchanged prefix and replace only its
// changed suffix, adjusting Parents to identify that suffix's existing root.
// Every live value must be spilled before publishing the chain. The common frame
// machine validates and adopts it before servicing any runtime effect or GC.
// Values must still expose the original live prefix when an error is returned,
// including when the owned arena moved, so cleanup and later reuse stay valid.
// A backend must not retain this packet or its invocation after returning.
type NativeFrames struct {
	Function            *Function
	Base, Depth         int
	PC                  uint32
	Fuel                uint64
	MaxFrames, MaxSlots uint32
	Values              []value.Value
	Frames              []NativeFrame
	// Resume is an execution-local token for a previously validated native
	// chain. Only an ordinary host return with unchanged frames may return it.
	// Parents reports how many existing parents this run replaces on success.
	Resume  uint64
	Parents uint32
}
