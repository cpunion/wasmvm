package slots

import (
	"context"
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type sourceSpan struct{ start, count uint32 }

// Published native code needs immediate/source metadata, not a graph of slice
// headers for every decoded operation. Every backing array in this image is
// pointer-free Go storage. No addresses or Go pointers are hidden in integers.
type sourceRecord struct {
	op                                          uint32
	offset                                      uint64
	a, b, memory, memoryOffsetHigh              uint32
	block                                       int64
	bits                                        uint64
	vector                                      [16]byte
	lane                                        byte
	labels, types, catches, on, inputs, outputs sourceSpan
	effects                                     ir.Effects
}

type sourceImage struct {
	records []sourceRecord
	labels  []uint32
	types   []wasm.ValueType
	catches []module.Catch
}

// FreezeSources runs once before publication, after code generation. Sources
// remains the rich compiler view until this point. Readers reconstruct a view
// in execution-owned scratch, so shared artifacts stay immutable.
func (f *Function) FreezeSources(ctx context.Context) error {
	if f.image != nil {
		return nil
	}
	var labels, types, catches uint64
	for i, source := range f.Sources {
		if i&255 == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		labels += uint64(len(source.Wasm.Labels))
		types += uint64(len(source.Wasm.Types)) + uint64(len(source.Inputs)) + uint64(len(source.Outputs))
		catches += uint64(len(source.Wasm.Catches)) + uint64(len(source.Wasm.On))
		if labels > math.MaxUint32 || types > math.MaxUint32 || catches > math.MaxUint32 {
			return &wasm.Error{Kind: wasm.ErrLimit, Detail: "runtime source image spans"}
		}
	}
	image := &sourceImage{records: make([]sourceRecord, len(f.Sources)), labels: make([]uint32, 0, int(labels)), types: make([]wasm.ValueType, 0, int(types)), catches: make([]module.Catch, 0, int(catches))}
	addTypes := func(types []wasm.ValueType) sourceSpan {
		span := sourceSpan{uint32(len(image.types)), uint32(len(types))}
		image.types = append(image.types, types...)
		return span
	}
	addCatches := func(catches []module.Catch) sourceSpan {
		span := sourceSpan{uint32(len(image.catches)), uint32(len(catches))}
		image.catches = append(image.catches, catches...)
		return span
	}
	for i, source := range f.Sources {
		if i&255 == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		in := source.Wasm
		labels := sourceSpan{uint32(len(image.labels)), uint32(len(in.Labels))}
		image.labels = append(image.labels, in.Labels...)
		image.records[i] = sourceRecord{op: in.Op, offset: in.Offset, a: in.A, b: in.B, memory: in.Memory, memoryOffsetHigh: in.MemoryOffsetHigh, block: in.Block, bits: in.Bits, vector: in.Vector, lane: in.Lane,
			labels: labels, types: addTypes(in.Types), catches: addCatches(in.Catches), on: addCatches(in.On), inputs: addTypes(source.Inputs), outputs: addTypes(source.Outputs), effects: source.Effects}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	f.image, f.Sources = image, nil
	return nil
}

func (f *Function) SourceCount() int {
	if f.image != nil {
		return len(f.image.records)
	}
	return len(f.Sources)
}

func (f *Function) SourceOffset(index uint32) uint64 {
	if f.image != nil {
		return f.image.records[index].offset
	}
	return f.Sources[index].Wasm.Offset
}

func (f *Function) SourceOutputs(index uint32) int {
	if f.image != nil {
		return int(f.image.records[index].outputs.count)
	}
	return len(f.Sources[index].Outputs)
}

// SourceIndices reads the common call/type/table immediates without rebuilding
// the slice-bearing compiler view on every native call boundary.
func (f *Function) SourceIndices(index uint32) (uint32, uint32) {
	if f.image != nil {
		r := &f.image.records[index]
		return r.a, r.b
	}
	r := &f.Sources[index].Wasm
	return r.A, r.B
}

func (f *Function) ReadSource(index uint32, scratch *ir.Source) *ir.Source {
	image := f.image
	if image == nil {
		return &f.Sources[index]
	}
	r := &image.records[index]
	*scratch = ir.Source{Wasm: module.Instr{Op: r.op, Offset: r.offset, A: r.a, B: r.b, Memory: r.memory, MemoryOffsetHigh: r.memoryOffsetHigh, Block: r.block, Bits: r.bits, Vector: r.vector, Lane: r.lane}, Effects: r.effects}
	if s := r.labels; s.count != 0 {
		scratch.Wasm.Labels = image.labels[s.start : s.start+s.count]
	}
	if s := r.types; s.count != 0 {
		scratch.Wasm.Types = image.types[s.start : s.start+s.count]
	}
	if s := r.catches; s.count != 0 {
		scratch.Wasm.Catches = image.catches[s.start : s.start+s.count]
	}
	if s := r.on; s.count != 0 {
		scratch.Wasm.On = image.catches[s.start : s.start+s.count]
	}
	if s := r.inputs; s.count != 0 {
		scratch.Inputs = image.types[s.start : s.start+s.count]
	}
	if s := r.outputs; s.count != 0 {
		scratch.Outputs = image.types[s.start : s.start+s.count]
	}
	return scratch
}
