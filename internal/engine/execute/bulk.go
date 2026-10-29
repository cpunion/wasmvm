package execute

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Bulk ranges are checked in full before writes. Cancellation is polled between
// bounded chunks; unlike a Wasm bounds trap, cancellation may leave partial
// progress. Backward copies preserve overlap across chunk boundaries.
func (e *Execution) bulk(in *state.Instance, op module.Instr, a, b, c value.Value) (r value.Value, trap wasm.TrapCode, err error) {
	dst, src, n := uint64(uint32(a.Lo)), uint64(uint32(b.Lo)), uint64(uint32(c.Lo))
	switch op.Op {
	case 0xfc0009:
		in.DropData(op.A)
		return
	case 0xfc000d:
		in.DropElement(op.A)
		return
	case 0xfc000f:
		table := in.Tables[op.A]
		old, ok := e.store.GrowTable(table, table.Type.Limits.Address(b.Lo), a.Lo)
		r.Lo = uint64(old)
		if !ok {
			r.Lo = table.Type.Limits.Failure()
		}
		return
	case 0xfc0010:
		r.Lo = uint64(len(in.Tables[op.A].Values))
		return
	}
	if op.Op <= 0xfc000b {
		var m, source *state.Memory
		switch op.Op {
		case 0xfc0008:
			m = in.Memories[op.B]
		case 0xfc000a:
			m, source = in.Memories[op.A], in.Memories[op.B]
		case 0xfc000b:
			m = in.Memories[op.A]
		}
		dst = m.Limits().Address(a.Lo)
		if op.Op == 0xfc000b {
			n = m.Limits().Address(c.Lo)
		} else if op.Op == 0xfc000a {
			src = source.Limits().Address(b.Lo)
			if m.Limits().Index64 && source.Limits().Index64 {
				n = c.Lo
			}
		}
		if !m.Range(dst, n) || op.Op == 0xfc0008 && !state.Range(len(in.Data[op.A]), src, n) || op.Op == 0xfc000a && !source.Range(src, n) {
			return r, wasm.TrapMemoryBounds, nil
		}
		backwards := op.Op == 0xfc000a && m.Memory == source.Memory && dst > src
		for done := uint64(0); done < n; {
			if err = e.poll(); err != nil {
				return
			}
			chunk := n - done
			if chunk > 4096 {
				chunk = 4096
			}
			at := done
			if backwards {
				at = n - done - chunk
			}
			switch op.Op {
			case 0xfc0008:
				m.Init(dst+at, src+at, chunk, in.Data[op.A])
			case 0xfc000a:
				m.CopyFrom(source.Memory, dst+at, src+at, chunk)
			case 0xfc000b:
				m.Fill(dst+at, byte(b.Lo), chunk)
			}
			done += chunk
		}
		return
	}
	var table *state.Table
	var source []uint64
	switch op.Op {
	case 0xfc000c:
		table = in.Tables[op.B]
		source = in.Elements[op.A]
	case 0xfc000e:
		table = in.Tables[op.A]
		source = in.Tables[op.B].Values
	case 0xfc0011:
		table = in.Tables[op.A]
	}
	dst = table.Type.Limits.Address(a.Lo)
	if op.Op == 0xfc0011 {
		n = table.Type.Limits.Address(c.Lo)
	} else if op.Op == 0xfc000e {
		sourceLimits := in.Tables[op.B].Type.Limits
		src = sourceLimits.Address(b.Lo)
		if table.Type.Limits.Index64 && sourceLimits.Index64 {
			n = c.Lo
		}
	}
	if !state.Range(len(table.Values), dst, n) || op.Op != 0xfc0011 && !state.Range(len(source), src, n) {
		return r, wasm.TrapTableBounds, nil
	}
	backwards := op.Op == 0xfc000e && table == in.Tables[op.B] && dst > src
	for done := uint64(0); done < n; {
		if err = e.poll(); err != nil {
			return
		}
		chunk := n - done
		if chunk > 512 {
			chunk = 512
		}
		at := done
		if backwards {
			at = n - done - chunk
		}
		if op.Op == 0xfc0011 {
			table.Fill(dst+at, chunk, b.Lo)
		} else {
			table.Init(dst+at, src+at, chunk, source)
		}
		done += chunk
	}
	return
}
