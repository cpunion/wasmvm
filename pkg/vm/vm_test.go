package vm

import (
	"context"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var addBinary = []byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 7, 1, 96, 2, 127, 127, 1, 127, 3, 2, 1, 0, 7, 7, 1, 3, 97, 100, 100, 0, 0, 10, 9, 1, 7, 0, 32, 0, 32, 1, 106, 11}

func setup(t *testing.T) (*Engine, *Store, *Linker) {
	t.Helper()
	e := NewEngine()
	s, err := e.NewStore(DefaultStoreConfig())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close(); e.Close() })
	return e, s, NewLinker(s)
}
func compileDefinition(t *testing.T, e *Engine, d *module.Module) *CompiledModule {
	t.Helper()
	v, err := validate.Check(d, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Compile(v, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	code, err := slots.Compile(ir, slots.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return &CompiledModule{engine: e, code: code}
}
func instantiate(t *testing.T, l *Linker, m *CompiledModule) *Instance {
	t.Helper()
	in, err := l.Instantiate(nil, m, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return in
}
func export(t *testing.T, in *Instance, name string) Extern {
	t.Helper()
	ex, err := in.Export(name)
	if err != nil {
		t.Fatal(err)
	}
	return ex
}
func bits(t *testing.T, v Value) uint64 {
	t.Helper()
	lo, _, err := v.Bits()
	if err != nil {
		t.Fatal(err)
	}
	return lo
}

func TestCompileAndPublishedLifetimes(t *testing.T) {
	e, s, l := setup(t)
	input := append([]byte(nil), addBinary...)
	compiled, err := e.Compile(nil, input)
	if err != nil {
		t.Fatal(err)
	}
	clear(input)
	info, err := compiled.Info()
	if err != nil || info.Backend != "interpreter" || info.Profile != "core2" || info.Functions != 1 || info.Exports[0].Name != "add" {
		t.Fatal(info, err)
	}
	info.Exports[0].Name = "mutated"
	fresh, _ := compiled.Info()
	if fresh.Exports[0].Name != "add" {
		t.Fatal("mutable metadata")
	}
	in := instantiate(t, l, compiled)
	fn := export(t, in, "add")
	sig, _ := fn.FunctionType()
	sig.Params[0] = wasm.F64
	compiled.Close()
	e.Close()
	compiled.Close()
	e.Close()
	if _, err := compiled.Info(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := e.Compile(nil, addBinary); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := e.NewStore(DefaultStoreConfig()); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := l.Instantiate(nil, compiled, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	result, err := fn.Call(nil, []Value{I32(17), I32(25)}, CallOptions{Metered: true, Fuel: 3})
	if err != nil || bits(t, result.Values[0]) != 42 || result.FuelRemaining != 0 {
		t.Fatal(result, err)
	}
	ref, err := fn.Reference()
	if err != nil {
		t.Fatal(err)
	}
	result, err = s.CallRef(nil, ref, []Value{I32(4), I32(5)}, CallOptions{})
	if err != nil || bits(t, result.Values[0]) != 9 {
		t.Fatal(err)
	}
	ref.Release()
	if _, err = s.CallRef(nil, ref, nil, CallOptions{}); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err = in.Export("missing"); !errors.Is(err, ErrLink) {
		t.Fatal(err)
	}
	s.Close()
	s.Close()
	if err = s.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	if _, err = fn.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestCompileFailuresAndConcurrentCompilation(t *testing.T) {
	for _, field := range []string{"Backend", "Profile"} {
		c := DefaultEngineConfig()
		reflect.ValueOf(&c).Elem().FieldByName(field).SetString("native")
		if _, err := NewEngineWithConfig(c); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(err)
		}
	}
	e, s, l := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.Compile(ctx, addBinary); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := e.Compile(nil, []byte{0}); !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal(err)
	}
	invalid := append([]byte(nil), addBinary...)
	invalid[len(invalid)-2] = 0x7c
	if _, err := e.Compile(nil, invalid); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	for _, phase := range []string{"Decode", "Validate", "Lower", "Slots"} {
		c := DefaultEngineConfig()
		reflect.ValueOf(&c).Elem().FieldByName(phase).SetZero()
		engine, _ := NewEngineWithConfig(c)
		if _, err := engine.Compile(nil, addBinary); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(phase, err)
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m, err := e.Compile(nil, addBinary)
			if err != nil {
				t.Error(err)
				return
			}
			m.Close()
		}()
	}
	wg.Wait()
	other := NewEngine()
	foreign, _ := other.Compile(nil, addBinary)
	if _, err := l.Instantiate(nil, foreign, CallOptions{}); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := l.Instantiate(nil, nil, CallOptions{}); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := s.Call(ctx, Extern{}, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestValueBitsOwnershipAndCollection(t *testing.T) {
	_, s, _ := setup(t)
	_, other, _ := setup(t)
	for _, v := range []Value{I32(math.MaxUint32), I64(math.MaxUint64), F32(1.5), F64(-1.5), F32Bits(0x7f800001), F64Bits(0xfff0000000000001), V128(1, 2)} {
		if !v.Type().Valid() {
			t.Fatal(v.Type())
		}
		lo, hi, err := v.Bits()
		if err != nil {
			t.Fatal(err)
		}
		copy, err := s.Root(v)
		if err != nil || copy.bits.Lo != lo || copy.bits.Hi != hi {
			t.Fatal(copy, err)
		}
		copy.Release()
	}
	if _, err := NullRef(wasm.I32); !errors.Is(err, ErrType) {
		t.Fatal(err)
	}
	if _, _, err := (Value{}).Bits(); !errors.Is(err, ErrType) {
		t.Fatal(err)
	}
	null, _ := NullRef(wasm.ExternRef)
	if !null.IsNull() {
		t.Fatal("null")
	}
	if got, err := s.ExternRef(null); err != nil || got != nil {
		t.Fatal(got, err)
	}
	if _, _, err := null.Bits(); !errors.Is(err, ErrType) {
		t.Fatal(err)
	}
	object := new(int)
	*object = 42
	v, err := s.NewExternRef(object)
	if err != nil || v.IsNull() {
		t.Fatal(err)
	}
	alias, err := s.Root(v)
	if err != nil {
		t.Fatal(err)
	}
	v.Release()
	v.Release()
	s.Collect()
	if got, err := s.ExternRef(alias); err != nil || got != object {
		t.Fatal(got, err)
	}
	if _, err := s.ExternRef(v); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := s.Root(v); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := other.Root(alias); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	if _, err := other.ExternRef(alias); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	alias.Release()
	s.Collect()
	usage, _ := s.Usage()
	if usage.ExternRefs != 0 || usage.HostRoots != 0 {
		t.Fatal(usage)
	}
	v, _ = s.NewExternRef(nil)
	s.Close()
	if _, err = s.ExternRef(v); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	v.Release()
	if _, err = s.Root(null); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err = s.NewExternRef(1); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err = s.Usage(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err = s.Collect(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestReferenceQuotasAndHostResultRollback(t *testing.T) {
	e := NewEngine()
	c := DefaultStoreConfig()
	c.MaxExternRefs = 1
	c.MaxHostRoots = 1
	s, _ := e.NewStore(c)
	defer s.Close()
	v, _ := s.NewExternRef(1)
	if _, err := s.NewExternRef(2); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := s.Root(v); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	v.Release()
	s.Collect()
	v, _ = s.NewExternRef(3)
	v.Release()
	s.Collect()
	s.nextExtern = math.MaxUint64
	if _, err := s.NewExternRef(4); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c.MaxHostRoots = 0
	s2, _ := e.NewStore(c)
	defer s2.Close()
	if _, err := s2.NewExternRef(1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if s2.nextExtern != 0 || len(s2.externs) != 0 {
		t.Fatal("failed allocation retained object")
	}
	// One host result can be rooted; the second exceeds the quota. Roll back
	// the first result's root without losing the preexisting persistent root.
	c.MaxHostRoots = 2
	s3, _ := e.NewStore(c)
	defer s3.Close()
	root, _ := s3.NewExternRef(5)
	fn, _ := s3.NewFunction(FunctionType{Results: []wasm.ValueType{wasm.ExternRef, wasm.ExternRef}}, func(*Caller, []Value) ([]Value, error) { return []Value{root, root}, nil })
	if _, err := fn.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	u, _ := s3.Usage()
	if u.HostRoots != 1 {
		t.Fatal(u)
	}
}

func TestHostReentryFuelRootsAndMemory(t *testing.T) {
	e, s, l := setup(t)
	var callback *Caller
	var saved Value
	add, _ := e.Compile(nil, addBinary)
	nested := export(t, instantiate(t, l, add), "add")
	host, _ := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}, func(c *Caller, args []Value) ([]Value, error) {
		callback = c
		if c.Store() != s || c.Context().Err() != nil {
			t.Fatal("callback context")
		}
		if _, err := nested.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrBusy) {
			t.Fatal(err)
		}
		if err := c.Collect(); err != nil {
			t.Fatal(err)
		}
		if object, err := s.ExternRef(args[0]); err != nil || object != "input" {
			t.Fatal(object, err)
		}
		if err := c.WriteMemory(0, 3, []byte{4, 5}); err != nil {
			t.Fatal(err)
		}
		data := make([]byte, 2)
		if err := c.ReadMemory(0, 3, data); err != nil || !reflect.DeepEqual(data, []byte{4, 5}) {
			t.Fatal(data, err)
		}
		if err := c.ReadMemory(9, 0, nil); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
		if err := c.WriteMemory(9, 0, nil); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
		if err := c.ReadMemory(0, 65536, []byte{0}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if err := c.WriteMemory(0, 65536, []byte{0}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		r, err := c.Call(nested, []Value{I32(2), I32(3)})
		if err != nil || bits(t, r.Values[0]) != 5 {
			t.Fatal(r, err)
		}
		v, err := c.NewExternRef("output")
		saved = v
		return []Value{v}, err
	})
	if err := l.Define("host", "echo", host); err != nil {
		t.Fatal(err)
	}
	sig := FunctionType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}
	d := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Module: "host", Name: "echo", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0xb}}}}, Memories: []module.Limits{{Min: 1}}, Exports: []module.Export{{Name: "run", Kind: 0, Index: 1}}}
	fn := export(t, instantiate(t, l, compileDefinition(t, e, d)), "run")
	input, _ := s.NewExternRef("input")
	r, err := fn.Call(nil, []Value{input}, CallOptions{Metered: true, Fuel: 5})
	if err != nil || r.FuelRemaining != 0 {
		t.Fatal(r, err)
	}
	if object, err := s.ExternRef(r.Values[0]); err != nil || object != "output" {
		t.Fatal(object, err)
	}
	if _, err := s.ExternRef(saved); !errors.Is(err, ErrReference) {
		t.Fatal("temporary escaped", err)
	}
	if _, err := callback.Call(nested, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := callback.CallRef(Value{}, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := callback.NewExternRef(1); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := callback.Collect(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := callback.ReadMemory(0, 0, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := callback.WriteMemory(0, 0, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	input.Release()
	r.Values[0].Release()
	s.Collect()
	u, _ := s.Usage()
	if u.ExternRefs != 0 || u.HostRoots != 0 {
		t.Fatal(u)
	}
}

func TestCloseFromCallbackPanicCleanupAndCancellation(t *testing.T) {
	_, s, _ := setup(t)
	entered, leave := make(chan struct{}), make(chan struct{})
	finished := make(chan error, 1)
	fn, _ := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) { close(entered); <-leave; return nil, nil })
	go func() { _, err := fn.Call(nil, nil, CallOptions{}); finished <- err }()
	<-entered
	if _, err := s.NewMemory(MemoryType{}); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
	s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := s.WaitClosed(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	close(leave)
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	_, s2, _ := setup(t)
	sentinel := errors.New("host panic")
	bad, _ := s2.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) { c.NewExternRef(new(int)); panic(sentinel) })
	func() {
		defer func() {
			if recover() != sentinel {
				t.Fatal("changed panic")
			}
		}()
		bad.Call(nil, nil, CallOptions{})
	}()
	if err := s2.Collect(); err != nil {
		t.Fatal("panic retained lease", err)
	}
	u, _ := s2.Usage()
	if u.ExternRefs != 0 || u.HostRoots != 0 {
		t.Fatal(u)
	}
	closeFn, _ := s2.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) { c.Store().Close(); return nil, nil })
	if _, err := closeFn.Call(nil, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s2.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
}
