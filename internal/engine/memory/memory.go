// Package memory owns private and domain-synchronized shared linear memory.
package memory

import (
	"encoding/binary"
	"math"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const PageSize uint64 = 65536
const MaxPages = 65536
const MaxPages64 = 1 << 48

type Memory struct {
	bytes       []byte
	owner       *native.Memory
	declared    module.Limits
	domain      *Domain
	refs        uint64
	reservation uint64
	maximum     uint64
}

// New applies the host page quota independently of the module declaration.
// A zero quota is literal. Allocations are exact-size; growth can transiently
// retain old and new buffers until the old allocation is collected.
func New(declared module.Limits, quota uint32) (*Memory, error) {
	if declared.Shared {
		return nil, wasm.ErrInvalid
	}
	return newMemory(declared, quota, false)
}

// NewNative uses the same semantics over pointer-free native-owned storage.
// Only the serialized Store accesses its view; growth refreshes every alias
// through this single Memory object before generated code can reenter.
func NewNative(declared module.Limits, quota uint32) (*Memory, error) {
	if declared.Shared {
		return nil, wasm.ErrInvalid
	}
	return newMemory(declared, quota, true)
}
func (m *Memory) Native() *native.Memory {
	if m.domain != nil {
		return nil
	}
	return m.owner
}
func newMemory(declared module.Limits, quota uint32, offheap bool) (*Memory, error) {
	limit := uint64(MaxPages)
	if declared.Index64 {
		limit = MaxPages64
	}
	if declared.Min > limit || declared.HasMax && (declared.Max > limit || declared.Min > declared.Max) {
		return nil, &wasm.Error{Kind: wasm.ErrInvalid, Detail: "memory limits"}
	}
	maximum := limit
	if declared.HasMax {
		maximum = declared.Max
	}
	if uint64(quota) < maximum {
		maximum = uint64(quota)
	}
	hostMax := uint64(math.MaxInt) / PageSize
	if uint64(maximum) > hostMax {
		maximum = hostMax
	}
	if declared.Min > maximum {
		return nil, &wasm.Error{Kind: wasm.ErrLimit, Detail: "initial memory pages"}
	}
	m := &Memory{declared: declared, maximum: maximum}
	if offheap {
		var err error
		m.owner, err = native.NewMemory(uint64(declared.Min)*PageSize, uint64(maximum)*PageSize)
		if err != nil {
			return nil, err
		}
		m.bytes, err = m.owner.Buffer()
		if err != nil {
			m.owner.Close()
			return nil, err
		}
	} else {
		m.bytes = make([]byte, int(uint64(declared.Min)*PageSize))
	}
	return m, nil
}
func (m *Memory) pages() uint64         { return uint64(len(m.bytes)) / PageSize }
func (m *Memory) Limits() module.Limits { return m.declared }

// Release discards backing storage after its Store has excluded all users.
// It is idempotent and prevents subsequent growth of a retired memory object.
func (m *Memory) release() {
	m.bytes = nil
	m.maximum = 0
	if m.owner != nil {
		m.owner.Close()
	}
}
func (m *Memory) grow(delta uint64) (uint64, bool) {
	old := m.pages()
	if delta > m.maximum-old {
		return old, false
	}
	next := old + delta
	if delta == 0 {
		return old, true
	}
	if m.owner != nil {
		if err := m.owner.Grow(next * PageSize); err != nil {
			return old, false
		}
		m.bytes, _ = m.owner.Buffer()
		return old, true
	}
	grown := make([]byte, int(next*PageSize))
	copy(grown, m.bytes)
	m.bytes = grown
	return old, true
}
func fits(size, offset, count uint64) bool { return offset <= size && count <= size-offset }
func (m *Memory) rangeBytes(offset, count uint64) bool {
	return fits(uint64(len(m.bytes)), offset, count)
}

// Read/Write copy across the host boundary without exposing a growth-sensitive
// backing slice. A rejected range leaves both source and destination unchanged.
func (m *Memory) readBytes(offset uint64, dst []byte) bool {
	if !m.rangeBytes(offset, uint64(len(dst))) {
		return false
	}
	copy(dst, m.bytes[int(offset):])
	return true
}
func (m *Memory) writeBytes(offset uint64, src []byte) bool {
	if !m.rangeBytes(offset, uint64(len(src))) {
		return false
	}
	copy(m.bytes[int(offset):], src)
	return true
}

// Fill, Copy and Init precheck the entire operation before any mutation. They
// are synchronous primitives; scheduling/cancellation chunking belongs to the
// runtime integration rather than weakening the out-of-bounds atomicity rule.
func (m *Memory) fill(dst uint64, b byte, count uint64) bool {
	if !m.rangeBytes(dst, count) {
		return false
	}
	for i := uint64(0); i < count; i++ {
		m.bytes[int(dst+i)] = b
	}
	return true
}
func (m *Memory) copyBytes(dst, src, count uint64) bool { return m.copyFrom(m, dst, src, count) }

// CopyFrom supports independent memories and aliases, with complete prechecks.
func (m *Memory) copyFrom(source *Memory, dst, src, count uint64) bool {
	if !m.rangeBytes(dst, count) || !source.rangeBytes(src, count) {
		return false
	}
	copy(m.bytes[int(dst):int(dst+count)], source.bytes[int(src):int(src+count)])
	return true
}
func (m *Memory) initBytes(dst, src, count uint64, data []byte) bool {
	if !m.rangeBytes(dst, count) || !fits(uint64(len(data)), src, count) {
		return false
	}
	copy(m.bytes[int(dst):int(dst+count)], data[int(src):int(src+count)])
	return true
}

func read(b []byte, n uint) uint64 {
	switch n {
	case 1:
		return uint64(b[0])
	case 2:
		return uint64(binary.LittleEndian.Uint16(b))
	case 4:
		return uint64(binary.LittleEndian.Uint32(b))
	case 8:
		return binary.LittleEndian.Uint64(b)
	default:
		panic("memory: invalid access size")
	}
}
func write(b []byte, n uint, x uint64) {
	switch n {
	case 1:
		b[0] = byte(x)
	case 2:
		binary.LittleEndian.PutUint16(b, uint16(x))
	case 4:
		binary.LittleEndian.PutUint32(b, uint32(x))
	case 8:
		binary.LittleEndian.PutUint64(b, x)
	default:
		panic("memory: invalid access size")
	}
}
func sign(x uint64, n uint) uint64 { shift := 64 - 8*n; return uint64(int64(x<<shift) >> shift) }
func (m *Memory) access(address, offset uint64, size uint) ([]byte, bool) {
	address = m.declared.Address(address)
	if offset > math.MaxUint64-address {
		return nil, false
	}
	at := address + offset
	if !m.rangeBytes(at, uint64(size)) {
		return nil, false
	}
	return m.bytes[int(at):int(at+uint64(size))], true
}

// Load implements every scalar and fixed-SIMD load, including lane operations.
// The address addition is widened before bounds checking. All immediates and
// lane indexes must have passed validation. previous supplies load_lane's input.
func (m *Memory) load(op uint32, address, offset uint64, lane byte, previous value.Value) (r value.Value, trap wasm.TrapCode) {
	if op >= 0x28 && op <= 0x35 {
		sizes := [14]uint{4, 8, 4, 8, 1, 1, 2, 2, 1, 1, 2, 2, 4, 4}
		n := sizes[op-0x28]
		b, ok := m.access(address, offset, n)
		if !ok {
			return r, wasm.TrapMemoryBounds
		}
		x := read(b, n)
		if op == 0x2c || op == 0x2e || op == 0x30 || op == 0x32 || op == 0x34 {
			x = sign(x, n)
		}
		if op == 0x28 || op == 0x2a || op >= 0x2c && op <= 0x2f {
			x = uint64(uint32(x))
		}
		return value.Value{Lo: x}, wasm.NoTrap
	}
	sub := op & 0xffff
	if op>>16 != 0xfd {
		panic("memory: non-load opcode")
	}
	var n uint
	switch {
	case sub == 0:
		n = 16
	case sub >= 1 && sub <= 6:
		n = 8
	case sub >= 7 && sub <= 10:
		n = 1 << (sub - 7)
	case sub >= 84 && sub <= 87:
		n = 1 << (sub - 84)
	case sub == 92:
		n = 4
	case sub == 93:
		n = 8
	default:
		panic("memory: non-load opcode")
	}
	b, ok := m.access(address, offset, n)
	if !ok {
		return r, wasm.TrapMemoryBounds
	}
	switch {
	case sub == 0:
		return value.Value{Lo: read(b, 8), Hi: read(b[8:], 8)}, wasm.NoTrap
	case sub >= 1 && sub <= 6:
		width := uint(1) << ((sub - 1) / 2)
		for i := uint(0); i < 8/width; i++ {
			x := read(b[i*width:], width)
			if sub%2 == 1 {
				x = sign(x, width)
			}
			setLane(&r, width*2, i, x)
		}
	case sub >= 7 && sub <= 10:
		x := read(b, n)
		for i := uint(0); i < 16/n; i++ {
			setLane(&r, n, i, x)
		}
	case sub >= 84 && sub <= 87:
		r = previous
		setLane(&r, n, uint(lane), read(b, n))
	case sub == 92 || sub == 93:
		r.Lo = read(b, n)
	}
	return r, wasm.NoTrap
}
func setLane(v *value.Value, size, index uint, x uint64) {
	shift := index * size * 8
	mask := ^uint64(0) >> (64 - size*8)
	x &= mask
	if shift >= 64 {
		shift -= 64
		v.Hi = v.Hi&^(mask<<shift) | x<<shift
	} else {
		v.Lo = v.Lo&^(mask<<shift) | x<<shift
	}
}

// Store implements every scalar and fixed-SIMD store. An invalid range causes
// no write, including vector and lane stores that straddle the memory end.
func (m *Memory) store(op uint32, address, offset uint64, lane byte, v value.Value) wasm.TrapCode {
	var n uint
	x := v.Lo
	switch {
	case op >= 0x36 && op <= 0x3e:
		n = [9]uint{4, 8, 4, 8, 1, 2, 1, 2, 4}[op-0x36]
	case op == 0xfd000b:
		n = 16
	case op >= 0xfd0058 && op <= 0xfd005b:
		n = 1 << (op - 0xfd0058)
		shift := uint(lane) * n * 8
		if shift >= 64 {
			x = v.Hi
			shift -= 64
		}
		x >>= shift
	default:
		panic("memory: non-store opcode")
	}
	b, ok := m.access(address, offset, n)
	if !ok {
		return wasm.TrapMemoryBounds
	}
	if n == 16 {
		write(b, 8, v.Lo)
		write(b[8:], 8, v.Hi)
	} else {
		write(b, n, x)
	}
	return wasm.NoTrap
}
