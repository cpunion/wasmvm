package vm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExceptionsHostBoundariesRootsAndIdentity(t *testing.T) {
	input, err := os.ReadFile("testdata/exceptions.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Features = wasm.ExceptionHandling | wasm.TailCall
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		info, err := m.Info()
		if err != nil || info.Imports[0].Kind != TagKind || len(info.Imports[0].Tag.Params) != 2 {
			t.Fatal(info, err)
		}
		sc := DefaultStoreConfig()
		sc.CollectEveryAllocation = true
		s, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		tag, err := s.NewTag(FunctionType{Params: []wasm.ValueType{wasm.I32, wasm.ExternRef}})
		if err != nil {
			t.Fatal(err)
		}
		typ, err := tag.TagType()
		if err != nil || len(typ.Params) != 2 {
			t.Fatal(typ, err)
		}
		l := NewLinker(s)
		if err := l.Define("h", "tag", tag); err != nil {
			t.Fatal(err)
		}
		var instance *Instance
		var guestThrow Extern
		mode := "throw"
		ordinary := errors.New("ordinary host failure")
		var retained *Caller
		var reentered *Exception
		host, err := s.NewFunction(typ, func(c *Caller, args []Value) ([]Value, error) {
			retained = c
			switch mode {
			case "ordinary":
				return nil, ordinary
			case "panic":
				panic(ordinary)
			case "reenter":
				_, err := c.Call(guestThrow, args)
				if !errors.As(err, &reentered) {
					t.Fatal(err)
				}
				if err := c.Collect(); err != nil {
					t.Fatal(err)
				}
				return nil, err
			}
			ex, err := c.NewException(tag, args)
			if err != nil {
				return nil, err
			}
			if err := c.Collect(); err != nil {
				return nil, err
			}
			return nil, ex
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := l.Define("h", "host", host); err != nil {
			t.Fatal(err)
		}
		instance = instantiate(t, l, m)
		guestThrow = export(t, instance, "throw")
		if !tag.SameTag(export(t, instance, "tag")) || tag.SameTag(export(t, instance, "other")) || tag.SameTag(host) {
			t.Fatal("tag identity")
		}
		payload, err := s.NewExternRef("exception payload")
		if err != nil {
			t.Fatal(err)
		}
		defer payload.Release()
		for _, metered := range []bool{false, true} {
			options := CallOptions{Metered: metered, Fuel: 10000}
			for _, name := range []string{"catch", "catch-host", "depth"} {
				r, err := export(t, instance, name).Call(nil, []Value{I32(42), payload}, options)
				if err != nil || bits(t, r.Values[0]) != 42 {
					t.Fatal(name, r, err)
				}
				object, err := s.ExternRef(r.Values[1])
				if err != nil || object != "exception payload" {
					t.Fatal(object, err)
				}
				r.Values[1].Release()
			}
			r, err := export(t, instance, "capture").Call(nil, []Value{I32(17), payload}, options)
			if err != nil {
				t.Fatal(err)
			}
			ex, err := s.Exception(r.Values[0])
			r.Values[0].Release()
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			gotTag, values, err := ex.Payload()
			if err != nil || !gotTag.SameTag(tag) || bits(t, values[0]) != 17 {
				t.Fatal(values, err)
			}
			values[1].Release()
			r, err = export(t, instance, "catch-rethrow").Call(nil, []Value{ex.Reference()}, options)
			if err != nil || bits(t, r.Values[0]) != 17 {
				t.Fatal(r, err)
			}
			r.Values[1].Release()
			for _, name := range []string{"rethrow", "throw", "tail"} {
				args := []Value{ex.Reference()}
				if name != "rethrow" {
					args = []Value{I32(19), payload}
				}
				_, err := export(t, instance, name).Call(nil, args, options)
				var thrown *Exception
				if !errors.As(err, &thrown) || !errors.Is(err, wasm.ErrException) || errors.Is(err, wasm.ErrTrap) {
					t.Fatal(name, err)
				}
				thrown.Release()
			}
			ex.Release()
			if _, _, err := ex.Payload(); !errors.Is(err, ErrReference) {
				t.Fatal(err)
			}
		}
		null, _ := NullRef(wasm.ExnRef)
		if _, err := export(t, instance, "rethrow").Call(nil, []Value{null}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if _, err := export(t, instance, "trap").Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("trap was caught", err)
		}
		mode = "ordinary"
		if _, err := export(t, instance, "catch-host").Call(nil, []Value{I32(1), payload}, CallOptions{}); !errors.Is(err, ordinary) {
			t.Fatal(err)
		}
		mode = "reenter"
		r, err := export(t, instance, "catch-host").Call(nil, []Value{I32(7), payload}, CallOptions{})
		if err != nil || bits(t, r.Values[0]) != 7 {
			t.Fatal(r, err)
		}
		r.Values[1].Release()
		reentered.Release()
		if _, err := retained.NewException(tag, nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		mode = "panic"
		func() {
			defer func() {
				if recover() != ordinary {
					t.Fatal("panic converted")
				}
			}()
			export(t, instance, "catch-host").Call(nil, []Value{I32(1), payload}, CallOptions{})
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		if _, err := export(t, instance, "spin").Call(ctx, nil, CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		if _, err := export(t, instance, "spin").Call(nil, nil, CallOptions{Metered: true, Fuel: 20}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		if _, err := s.NewException(tag, []Value{I32(1)}); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		if _, err := s.NewTag(FunctionType{Results: []wasm.ValueType{wasm.I32}}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		if _, err := host.TagType(); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
	})
}

func TestExceptionPayloadKeepsFailedInstanceAndGCGraph(t *testing.T) {
	input, err := os.ReadFile("testdata/exceptions-gc-failed.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Features = wasm.GC | wasm.ExceptionHandling
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		sc := DefaultStoreConfig()
		sc.CollectEveryAllocation = true
		s, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		payload, _ := s.NewExternRef("nested exception")
		global, err := s.NewGlobal(GlobalType{Value: wasm.ExternRef, Mutable: true}, payload)
		if err != nil {
			t.Fatal(err)
		}
		l := NewLinker(s)
		if err := l.Define("h", "payload", global); err != nil {
			t.Fatal(err)
		}
		_, err = l.Instantiate(nil, m, CallOptions{})
		var ex *Exception
		if !errors.As(err, &ex) {
			t.Fatal(err)
		}
		payload.Release()
		null, _ := NullRef(wasm.ExternRef)
		if err := global.GlobalSet(null); err != nil {
			t.Fatal(err)
		}
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		tag, args, err := ex.Payload()
		if err != nil || len(args) != 4 {
			t.Fatal(args, err)
		}
		if _, err := tag.TagType(); err != nil {
			t.Fatal(err)
		}
		result, err := s.CallRef(nil, args[1], nil, CallOptions{})
		if err != nil || bits(t, result.Values[0]) != 42 {
			t.Fatal(result, err)
		}
		if lo, hi, err := args[3].Bits(); err != nil || lo != 11 || hi != 22 {
			t.Fatal(lo, hi, err)
		}
		inner, err := s.Exception(args[2])
		if err != nil {
			t.Fatal(err)
		}
		_, innerArgs, err := inner.Payload()
		if err != nil {
			t.Fatal(err)
		}
		object, err := s.ExternRef(innerArgs[0])
		if err != nil || object != "nested exception" {
			t.Fatal(object, err)
		}
		innerArgs[0].Release()
		inner.Release()
		for _, v := range args {
			v.Release()
		}
		ex.Release()
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		usage, err := s.Usage()
		if err != nil || usage.Core.Instances != 0 || usage.Core.Tags != 0 || usage.Core.HeapObjects != 0 {
			t.Fatal(usage, err)
		}
	})
}
