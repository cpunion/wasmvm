// Package nativeabi emits the fixed native entry frame in pure Go. Guest call
// frames belong to the VM arena and never push another hardware entry frame.
package nativeabi

import (
	"encoding/binary"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Entry struct {
	Code                   []byte
	BodyOffset, ExitOffset uint32
	StackBytes             uint32
}

// Wrap adds one C-ABI prologue and shared epilogue around trusted compiler output.
// The body receives the record in rdi (amd64) or x0 (arm64), puts the exit status
// in eax/w0, and falls through or branches to ExitOffset. It must not return
// directly, modify sp, or use arm64's reserved x18/frame/link registers. Branch
// relocation and continuation dispatch remain the caller's responsibility.
func Wrap(arch string, body []byte, maxBytes uint64) (Entry, error) {
	var prefix, suffix []byte
	var stack uint32
	switch arch {
	case "amd64":
		// Preserve rbp, rbx and r12-r15; keep rsp 16-byte aligned in the body.
		prefix = []byte{0x55, 0x48, 0x89, 0xe5, 0x53, 0x41, 0x54, 0x41, 0x55, 0x41, 0x56, 0x41, 0x57, 0x48, 0x83, 0xec, 8}
		suffix = []byte{0x48, 0x83, 0xc4, 8, 0x41, 0x5f, 0x41, 0x5e, 0x41, 0x5d, 0x41, 0x5c, 0x5b, 0x5d, 0xc3}
		stack = 56 // Excludes the incoming hardware return address.
	case "arm64":
		if len(body)%4 != 0 {
			return Entry{}, wasm.ErrInvalid
		}
		// Preserve x19-x30 and entire q8-q15, exceeding the C ABI's low-64-bit
		// SIMD requirement. x18 remains reserved. All pushes are 16-byte aligned.
		prefix = words(0xa9bf7bfd, 0x910003fd, 0xa9bf53f3, 0xa9bf5bf5, 0xa9bf63f7,
			0xa9bf6bf9, 0xa9bf73fb, 0xadbf27e8, 0xadbf2fea, 0xadbf37ec, 0xadbf3fee)
		suffix = words(0xacc13fee, 0xacc137ec, 0xacc12fea, 0xacc127e8, 0xa8c173fb,
			0xa8c16bf9, 0xa8c163f7, 0xa8c15bf5, 0xa8c153f3, 0xa8c17bfd, 0xd65f03c0)
		stack = 224
	default:
		return Entry{}, wasm.ErrUnsupported
	}
	n := uint64(len(prefix)) + uint64(len(body)) + uint64(len(suffix))
	if n > maxBytes || n > uint64(^uint32(0)) || n > uint64(^uint(0)>>1) {
		return Entry{}, wasm.ErrLimit
	}
	code := make([]byte, 0, int(n))
	code = append(code, prefix...)
	code = append(code, body...)
	code = append(code, suffix...)
	return Entry{code, uint32(len(prefix)), uint32(len(prefix) + len(body)), stack}, nil
}

func words(in ...uint32) []byte {
	out := make([]byte, 4*len(in))
	for i, v := range in {
		binary.LittleEndian.PutUint32(out[i*4:], v)
	}
	return out
}
