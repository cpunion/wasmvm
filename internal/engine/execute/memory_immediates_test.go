package execute

import (
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPortableScalarMemoryImmediates(t *testing.T) {
	for _, offset := range []uint64{0, 17, 65532, math.MaxUint32 - 1, math.MaxUint32, 1 << 32, math.MaxUint64} {
		for _, metered := range []bool{false, true} {
			s := state.New(state.DefaultLimits())
			defer s.Close()
			store := module.Instr{Op: 0x36, Offset: 777, Memory: 1, B: uint32(offset), MemoryOffsetHigh: uint32(offset >> 32)}
			load := store
			load.Op = 0x28
			other := load
			other.Memory = 0
			in := instanceWithFeatures(t, s, &module.Module{
				Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I64}, Results: []wasm.ValueType{wasm.I32, wasm.I32}}},
				Memories: []module.Limits{
					{Min: 1, Max: 1, HasMax: true, Index64: true},
					{Min: 1, Max: 1, HasMax: true, Index64: true},
				},
				Functions: []module.Function{{Body: module.Expr{
					{Op: 0x20}, {Op: 0x41, Bits: 0x12345678}, store,
					{Op: 0x20}, load, {Op: 0x20}, other, {Op: 0xb},
				}}},
			}, nil, wasm.Core3Features)
			function, ok := s.Function(in.Functions[0])
			if !ok {
				t.Fatal("missing instantiated function")
			}
			code := function.Code
			for _, instruction := range code.Code {
				if instruction.Op == 0x28 || instruction.Op == 0x36 {
					source := code.Sources[instruction.Source].Wasm
					want := slots.WideMemoryOffset
					if offset < uint64(want) {
						want = uint32(offset)
					}
					if instruction.Extra != source.Memory || instruction.Count != want {
						t.Fatal("incorrect cached immediate", offset, instruction, source)
					}
				}
			}
			e := New(s, nil, Config{Limits: DefaultLimits(), Metered: metered, Fuel: 100})
			defer e.Close()
			if err := e.Start(in.Functions[0], []value.Value{{}}); err != nil {
				t.Fatal(offset, metered, err)
			}
			status := e.Run()
			if offset > 65532 {
				var trap *wasm.Trap
				if status != Failed || !errors.As(e.Err(), &trap) || trap.Code != wasm.TrapMemoryBounds || trap.Offset != 777 {
					t.Fatal("wide offset lost its bounds trap or source", offset, metered, status, e.Err())
				}
			} else if got := e.Results(); status != Done || len(got) != 2 || got[0].Lo != 0x12345678 || got[1].Lo != 0 {
				t.Fatal("wrong memory or offset", offset, metered, status, got, e.Err())
			}
		}
	}
}
