package preview2

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func fixedRandom(t *testing.T) *Random {
	t.Helper()
	data := make([]byte, 80)
	for i := 32; i < 64; i++ {
		data[i] = 1
	}
	for i := 64; i < 80; i++ {
		data[i] = byte(i - 64)
	}
	r, err := newRandom(RandomConfig{MaxBytes: 16 << 10}, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func randomEngine(t *testing.T, backend string) (*component.Engine, *component.Compiled) {
	t.Helper()
	config := component.DefaultConfig()
	config.Core.Backend = backend
	e, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	m, err := e.Compile(nil, fixture(t, "random"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	return e, m
}

func TestRandomEveryOperationThroughBothBackends(t *testing.T) {
	// Independent OpenSSL AES-256-CTR outputs for keys 00*32 and 01*32,
	// IV 00*16, and 32 zero plaintext bytes. Production exposes no seed injection.
	secure, _ := hex.DecodeString("dc95c078a2408989ad48a21492842087530f8afbc74536b9a963b4f1c4cb738b")
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			_, m := randomEngine(t, backend)
			r := fixedRandom(t)
			in, err := m.Instantiate(nil, r.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			call := func(name string, args ...component.Value) *component.Value {
				t.Helper()
				f, err := in.Function(name)
				if err != nil {
					t.Fatal(err)
				}
				v, err := f.Call(nil, args, component.CallOptions{Metered: true, Fuel: 10000})
				if err != nil {
					t.Fatal(name, err)
				}
				return v.Value
			}
			first := call("secure-bytes", component.Value{Bits: 16})
			if first.Text != string(secure[:16]) || first.Elements != nil {
				t.Fatal(first)
			}
			if v := call("secure-u64"); v.Bits != binary.LittleEndian.Uint64(secure[16:24]) {
				t.Fatal(v)
			}
			if v := call("secure-bytes", component.Value{Bits: 8}); v.Text != string(secure[24:]) {
				t.Fatal(v)
			}
			insecure := call("insecure-bytes", component.Value{Bits: 16})
			if insecure.Text != "\x72\x98\xca\xa5\x65\x03\x1e\xad\xc6\xce\x23\xd2\x3e\xa6\x63\x78" {
				t.Fatal("independent stream state", insecure)
			}
			if v := call("insecure-u64"); v.Bits != 0xcd39b64489aea8ae {
				t.Fatal(v)
			}
			for _, name := range []string{"secure-bytes", "insecure-bytes"} {
				if v := call(name, component.Value{}); v.Text != "" || v.Elements != nil {
					t.Fatal(v)
				}
			}
			for _, name := range []string{"secure-count", "insecure-count"} {
				if v := call(name, component.Value{Bits: 4096}); v.Bits != 4096 {
					t.Fatal(v)
				}
			}
			if first.Text != string(secure[:16]) {
				t.Fatal("allocation reset mutated owned bytes")
			}
			seed := call("seed")
			if len(seed.Elements) != 2 || seed.Elements[0].Bits != 0x0706050403020100 || seed.Elements[1].Bits != 0x0f0e0d0c0b0a0908 {
				t.Fatal(seed)
			}
			seed.Elements[0].Bits = 0
			if v := call("seed"); v.Elements[0].Bits != 0x0706050403020100 {
				t.Fatal("seed result aliases provider")
			}
			for _, n := range []uint64{(16 << 10) + 1, ^uint64(0)} {
				fresh, err := m.Instantiate(nil, r.Imports(), component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				f, err := fresh.Function("secure-bytes")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := f.Call(nil, []component.Value{{Bits: n}}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(n, err)
				}
				fresh.Close()
				if err := fresh.WaitClosed(nil); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

type failedEntropy struct {
	calls int
	err   error
}

func (r *failedEntropy) Read([]byte) (int, error) { r.calls++; return 0, r.err }

func TestRandomConstructionAndImportContracts(t *testing.T) {
	sentinel := errors.New("entropy unavailable")
	reader := &failedEntropy{err: sentinel}
	for _, n := range []uint64{1, 7, 1 << 28, ^uint64(0)} {
		if r, err := newRandom(RandomConfig{MaxBytes: n}, reader); r != nil || !errors.Is(err, wasm.ErrLimit) || reader.calls != 0 {
			t.Fatal(r, err, reader.calls)
		}
	}
	if r, err := newRandom(RandomConfig{}, reader); r != nil || !errors.Is(err, sentinel) {
		t.Fatal(r, err)
	}
	if r, err := newRandom(RandomConfig{}, bytes.NewReader(nil)); r != nil || !errors.Is(err, io.EOF) {
		t.Fatal(r, err)
	}
	for _, r := range []*Random{nil, {}} {
		if len(r.Imports().Instances) != 0 {
			t.Fatal("uninitialized provider exported callbacks")
		}
	}
	e, compiled := randomEngine(t, "interpreter")
	for _, name := range []string{"length", "bytes", "scalar", "seed"} {
		m, err := e.Compile(nil, fixture(t, "random-wrong-"+name))
		if err != nil {
			t.Fatal(err)
		}
		if in, err := m.Instantiate(nil, fixedRandom(t).Imports(), component.CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) {
			t.Fatal(name, in, err)
		}
		m.Close()
	}
	r, err := NewRandom(RandomConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if r.maximum != 64<<10 {
		t.Fatal(r.maximum)
	}
	if b, err := r.read(context.Background(), 0, 32); err != nil || len(b) != 32 {
		t.Fatal(err)
	}
	in, err := compiled.Instantiate(nil, r.Imports(), component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	f, err := in.Function("secure-count")
	if err != nil {
		t.Fatal(err)
	}
	if v, err := f.Call(nil, []component.Value{{Bits: 64 << 10}}, component.CallOptions{}); err != nil || v.Value.Bits != 64<<10 {
		t.Fatal(v, err)
	}
}

type cancelStream struct {
	cipher func([]byte, []byte)
	cancel context.CancelFunc
}

func (s cancelStream) XORKeyStream(dst, src []byte) { s.cipher(dst, src); s.cancel() }

func TestRandomCancellationAndSharedProvider(t *testing.T) {
	r := fixedRandom(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if b, err := r.read(ctx, 0, 8); b != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(b, err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	r.sources[0].stream = cancelStream{cipher: r.sources[0].stream.XORKeyStream, cancel: cancel}
	if b, err := r.read(ctx, 0, 8192); b != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(b, err)
	}
	if len(r.sources[0].available) != 1 {
		t.Fatal("canceled fill retained its stream lease")
	}
	if b, err := r.read(context.Background(), 0, 8); err != nil || len(b) != 8 {
		t.Fatal(err)
	}
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) { randomCancellationAndSharing(t, backend) })
	}
}

func randomCancellationAndSharing(t *testing.T, backend string) {
	_, m := randomEngine(t, backend)
	r := fixedRandom(t)
	for _, closeStore := range []bool{false, true} {
		// Hold the stream while a real guest enters its u64 host call.
		<-r.sources[0].available
		entered := make(chan struct{})
		imports := r.Imports()
		iface := imports.Instances[RandomInterface]
		original := iface.Functions["get-random-u64"]
		iface.Functions["get-random-u64"] = func(c *component.Caller, v []component.Value) (*component.Value, error) {
			close(entered)
			return original(c, v)
		}
		in, err := m.Instantiate(nil, imports, component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f, err := in.Function("secure-u64")
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() { _, err := f.Call(ctx, nil, component.CallOptions{}); done <- err }()
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("guest did not enter host")
		}
		if closeStore {
			in.Close()
		} else {
			cancel()
		}
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("stream contention ignored cancellation")
		}
		r.sources[0].available <- struct{}{}
		in.Close()
		wait, stop := context.WithTimeout(context.Background(), time.Second)
		if err := in.WaitClosed(wait); err != nil {
			t.Fatal(err)
		}
		stop()
	}
	// Separate component Stores share the provider, not their execution locks.
	var wg sync.WaitGroup
	results := make(chan uint64, 8)
	failures := make(chan error, 8)
	for i := 0; i < 8; i++ {
		in, err := m.Instantiate(nil, r.Imports(), component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f, err := in.Function("secure-u64")
		if err != nil {
			t.Fatal(err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer in.Close()
			v, err := f.Call(nil, nil, component.CallOptions{})
			if err != nil {
				failures <- err
				return
			}
			results <- v.Value.Bits
		}()
	}
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	seen := make(map[uint64]bool)
	for v := range results {
		if seen[v] {
			t.Fatal("reused deterministic counter output")
		}
		seen[v] = true
	}
	if len(seen) != 8 {
		t.Fatal(seen)
	}
}
