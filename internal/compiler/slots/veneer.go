package slots

import "github.com/cpunion/wasmvm/internal/compiler/ir"

// CallVeneer lowers a parameter-only forwarding function to two explicit frame
// transitions. It contains no arithmetic, memory access, branches or handlers.
// The shared call dispatcher still resolves the live target and its type.
// Logical frames, slots, roots and original fuel boundaries remain unchanged.
type CallVeneer struct {
	CallPC, ReturnPC uint32
	Index            Slot
	Bits             uint64
	CallFuel         uint64
	ReturnFuel       uint64
}

func callVeneer(f *Function) *CallVeneer {
	if len(f.Code) > 64 || len(f.Handlers) != 0 || len(f.Protected) != 0 || len(f.Prompts) != 0 || len(f.Edges) != 0 {
		return nil
	}
	index := NoSlot
	var bits, fuel uint64
	for pc, in := range f.Code {
		source := &f.Sources[in.Source]
		fuel += source.Fuel()
		switch in.Op {
		case 0x20:
			if index != NoSlot || in.Count != 1 || in.Dst != Slot(source.Wasm.A) || source.Wasm.A >= f.LocalCount {
				return nil
			}
		case 0x41, 0x42:
			if index != NoSlot || uint64(in.Dst) < uint64(f.LocalCount) {
				return nil
			}
			index, bits = in.Dst, source.Wasm.Bits
			if in.Op == 0x41 {
				bits = uint64(uint32(bits))
			}
		case 0x10, 0x11:
			count := in.Count
			if in.Op == 0x11 {
				if count == 0 || index == NoSlot || f.Args[in.Extra+count-1] != index {
					return nil
				}
				count--
			} else if index != NoSlot {
				return nil
			}
			for _, arg := range f.Args[in.Extra : in.Extra+count] {
				if uint64(arg) >= uint64(f.LocalCount) {
					return nil
				}
			}
			if pc+1 >= len(f.Code) {
				return nil
			}
			ret := f.Code[pc+1]
			if ret.Op != ControlBase+uint32(ir.Return) || int(ret.Count) != len(source.Outputs) {
				return nil
			}
			for i, output := range f.Args[ret.Extra : ret.Extra+ret.Count] {
				if output != f.Args[in.Extra+in.Count+uint32(i)] {
					return nil
				}
			}
			return &CallVeneer{CallPC: uint32(pc), ReturnPC: uint32(pc + 1), Index: index, Bits: bits,
				CallFuel: fuel, ReturnFuel: f.Sources[ret.Source].Fuel()}
		default:
			return nil
		}
	}
	return nil
}
