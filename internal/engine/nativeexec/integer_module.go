package nativeexec

import (
	"context"
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// All linked targets belong to the same image. Imports and foreign targets
// exit to the shared driver before running; their next artifact has its own
// environment policy. Keep the entire image conservative, including cold and
// unreachable code. SIMD and unknown/new operations retain the full transition.
func integerModule(ctx context.Context, m *slots.Module) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if m == nil {
		return false, nil
	}
	for _, f := range m.Functions {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if f == nil {
			return false, nil
		}
		for _, types := range [][]wasm.ValueType{f.Slots, f.Signature.Params, f.Signature.Results} {
			for i, typ := range types {
				if i%256 == 0 {
					if err := ctx.Err(); err != nil {
						return false, err
					}
				}
				if typ == wasm.F32 || typ == wasm.F64 || typ == wasm.V128 {
					return false, nil
				}
			}
		}
		for i, in := range f.Code {
			if i%256 == 0 {
				if err := ctx.Err(); err != nil {
					return false, err
				}
			}
			switch op := in.Op; {
			case op <= 5, op == 8 || op == 10, op >= 0x0b && op <= 0x15,
				op >= 0x1a && op <= 0x1c, op >= 0x20 && op <= 0x29,
				op >= 0x2c && op <= 0x37, op >= 0x3a && op <= 0x42,
				op >= 0x45 && op <= 0x5a, op >= 0x67 && op <= 0x8a,
				op == 0xa7 || op == 0xac || op == 0xad, op >= 0xc0 && op <= 0xc4,
				op >= 0xd0 && op <= 0xd6, op >= 0xfc0008 && op <= 0xfc0011,
				op >= 0xfe0000 && op <= 0xfe0003, op >= 0xfe0010 && op <= 0xfe004e,
				op >= slots.ControlBase && op <= slots.ControlBase+uint32(ir.IfNull):
			default:
				return false, nil
			}
		}
	}
	return true, ctx.Err()
}
