package nativeexec

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestIntegerModuleIncludesEveryBodyAndColdOperation(t *testing.T) {
	proves := func(m *slots.Module) bool {
		t.Helper()
		ok, err := integerModule(context.Background(), m)
		if err != nil {
			t.Fatal(err)
		}
		return ok
	}
	pure := &slots.Function{Slots: []wasm.ValueType{wasm.I32}, Code: []slots.Instruction{
		{Op: 0x41}, {Op: 0xad}, {Op: 0xfc000a}, {Op: 0xfe001e},
		{Op: slots.ControlBase + uint32(ir.Return)},
	}}
	if !proves(&slots.Module{Functions: []*slots.Function{pure}}) {
		t.Fatal("integer, bulk and atomic operations need no FP transition")
	}
	if proves(nil) || proves(&slots.Module{Functions: []*slots.Function{nil}}) {
		t.Fatal("missing proof admitted")
	}
	// An otherwise unused body still belongs to the linked image. Scalar
	// types in signatures or storage reject even if the body is empty.
	for _, typ := range []wasm.ValueType{wasm.F32, wasm.F64, wasm.V128} {
		for place := 0; place < 3; place++ {
			other := &slots.Function{}
			switch place {
			case 0:
				other.Slots = []wasm.ValueType{typ}
			case 1:
				other.Signature.Params = []wasm.ValueType{typ}
			case 2:
				other.Signature.Results = []wasm.ValueType{typ}
			}
			if proves(&slots.Module{Functions: []*slots.Function{pure, other}}) {
				t.Fatal("mixed image admitted", typ, place)
			}
		}
	}
	for _, op := range []uint32{0x43, 0x44, 0x2a, 0x39, 0x5b, 0x8b, 0xa8, 0xbb,
		0xfd000c, 0xfd006f, 0xfc0000, 0xfb0000, 0xe0, 0x19, 0xfe0004, 0xfe004f, ^uint32(0)} {
		cold := &slots.Function{Code: []slots.Instruction{{Op: slots.ControlBase + uint32(ir.Return)}, {Op: op}}}
		if proves(&slots.Module{Functions: []*slots.Function{pure, cold}}) {
			t.Fatalf("FP, SIMD or unknown cold operation %x admitted", op)
		}
	}
}

type proofCancellation struct {
	context.Context
	cancel           context.CancelFunc
	polls, remaining int
}

func (c *proofCancellation) Err() error {
	c.polls++
	if c.polls == c.remaining {
		c.cancel()
	}
	return c.Context.Err()
}

func TestIntegerProofCancellationDuringEachScan(t *testing.T) {
	for _, kind := range []string{"functions", "slots", "instructions"} {
		m := &slots.Module{}
		for i := 0; i < 1024; i++ {
			switch kind {
			case "functions":
				m.Functions = append(m.Functions, &slots.Function{})
			case "slots":
				if i == 0 {
					m.Functions = []*slots.Function{{}}
				}
				m.Functions[0].Slots = append(m.Functions[0].Slots, wasm.I32)
			case "instructions":
				if i == 0 {
					m.Functions = []*slots.Function{{}}
				}
				m.Functions[0].Code = append(m.Functions[0].Code, slots.Instruction{Op: 1})
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		probe := &proofCancellation{Context: ctx, cancel: cancel, remaining: 5}
		_, err := integerModule(probe, m)
		cancel()
		if !errors.Is(err, context.Canceled) || probe.polls != 5 {
			t.Fatal(kind, "proof ignored cancellation", probe.polls, err)
		}
	}
}
