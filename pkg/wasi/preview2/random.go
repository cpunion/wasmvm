package preview2

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const (
	RandomInterface         = "wasi:random/random@0.2.12"
	InsecureRandomInterface = "wasi:random/insecure@0.2.12"
	InsecureSeedInterface   = "wasi:random/insecure-seed@0.2.12"
)

// RandomConfig bounds each request independently of the component transfer and
// logical value budgets. Zero selects the default 64 KiB limit. A nonzero limit
// must be at least eight bytes so scalar u64 operations remain available.
type RandomConfig struct{ MaxBytes uint64 }

// Random owns independent secure/insecure streams and a stable hash-map seed.
// Entropy is acquired before binding, never during a guest operation. Reusing
// this provider across Stores is supported; stream access observes cancellation.
type Random struct {
	maximum uint64
	sources [2]randomStream
	seed    [2]uint64
}

type randomStream struct {
	stream    cipher.Stream
	available chan struct{}
}

// NewRandom seeds AES-256 counter streams from the host cryptographic source.
// No public deterministic seed/source is accepted for the secure interface.
func NewRandom(config RandomConfig) (*Random, error) { return newRandom(config, rand.Reader) }

func newRandom(config RandomConfig, entropy io.Reader) (*Random, error) {
	if config.MaxBytes == 0 {
		config.MaxBytes = 64 << 10
	}
	if config.MaxBytes < 8 || config.MaxBytes > (1<<28)-1 {
		return nil, fmt.Errorf("%w: random byte limit", wasm.ErrLimit)
	}
	var seed [80]byte
	if _, err := io.ReadFull(entropy, seed[:]); err != nil {
		return nil, fmt.Errorf("initialize WASI random: %w", err)
	}
	r := &Random{maximum: config.MaxBytes, seed: [2]uint64{binary.LittleEndian.Uint64(seed[64:72]), binary.LittleEndian.Uint64(seed[72:80])}}
	for i := range r.sources {
		block, err := aes.NewCipher(seed[32*i : 32*(i+1)])
		if err != nil {
			return nil, err
		}
		r.sources[i] = randomStream{stream: cipher.NewCTR(block, make([]byte, aes.BlockSize)), available: make(chan struct{}, 1)}
		r.sources[i].available <- struct{}{}
	}
	clear(seed[:])
	return r, nil
}

func (r *Random) read(ctx context.Context, source int, n uint64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if n > r.maximum {
		return nil, fmt.Errorf("%w: random byte request", wasm.ErrLimit)
	}
	s := &r.sources[source]
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.available:
	}
	defer func() { s.available <- struct{}{} }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data := make([]byte, int(n))
	for at := 0; at < len(data); at += 4096 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		chunk := data[at:min(at+4096, len(data))]
		s.stream.XORKeyStream(chunk, chunk)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

func randomType(bytes bool) component.FunctionCheck {
	return func(f component.FunctionInfo) error {
		if bytes {
			_, n := f.Parameter(0)
			if f.ParameterCount() != 1 || n.Kind() != component.KindU64 || f.Result().Kind() != component.KindList || f.Result().Element().Kind() != component.KindU8 {
				return vm.ErrType
			}
		} else if f.ParameterCount() != 0 || f.Result().Kind() != component.KindU64 {
			return vm.ErrType
		}
		return nil
	}
}

// Imports implements all five operations in the three pinned random interfaces.
func (r *Random) Imports() component.Imports {
	if r == nil || r.sources[0].stream == nil || r.sources[1].stream == nil {
		return component.Imports{}
	}
	result := component.Imports{Instances: make(map[string]component.Imports)}
	for i, name := range []string{RandomInterface, InsecureRandomInterface} {
		prefix := "get-random-"
		if i == 1 {
			prefix = "get-insecure-random-"
		}
		result.Instances[name] = component.Imports{
			Functions: map[string]component.HostFunction{
				prefix + "bytes": func(c *component.Caller, args []component.Value) (*component.Value, error) {
					data, err := r.read(c.Context(), i, args[0].Bits)
					if err != nil {
						return nil, err
					}
					v := component.ByteList(data)
					return &v, nil
				},
				prefix + "u64": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
					data, err := r.read(c.Context(), i, 8)
					if err != nil {
						return nil, err
					}
					return &component.Value{Bits: binary.LittleEndian.Uint64(data)}, nil
				},
			},
			FunctionChecks: map[string]component.FunctionCheck{prefix + "bytes": randomType(true), prefix + "u64": randomType(false)},
		}
	}
	result.Instances[InsecureSeedInterface] = component.Imports{
		Functions: map[string]component.HostFunction{"insecure-seed": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := c.Context().Err(); err != nil {
				return nil, err
			}
			return &component.Value{Elements: []component.Value{{Bits: r.seed[0]}, {Bits: r.seed[1]}}}, nil
		}},
		FunctionChecks: map[string]component.FunctionCheck{"insecure-seed": func(f component.FunctionInfo) error {
			v := f.Result()
			_, a := v.Field(0)
			_, b := v.Field(1)
			if f.ParameterCount() != 0 || v.Kind() != component.KindTuple || v.FieldCount() != 2 || a.Kind() != component.KindU64 || b.Kind() != component.KindU64 {
				return vm.ErrType
			}
			return nil
		}},
	}
	return result
}
