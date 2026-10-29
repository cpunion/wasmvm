package vm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestFunctionReferencesHostTypesRootsAndTailFrames(t *testing.T) {
	input, err := os.ReadFile("testdata/function-references.wasm")
	if err != nil {
		t.Fatal(err)
	}
	other, err := os.ReadFile("testdata/function-references-import.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		if err := base.Validate(nil, input); err == nil {
			t.Fatal("references enabled by default")
		}
		cfg := base.config
		cfg.Features = wasm.FunctionReferences | wasm.TailCall
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		sc := DefaultStoreConfig()
		sc.Execution.MaxFrames = 2
		s, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		in := instantiate(t, NewLinker(s), m)
		inc := export(t, in, "inc")
		ref, err := inc.Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer ref.Release()
		if !ref.Type().Canonical() || ref.Type().Nullable() {
			t.Fatal("export lacks its canonical function type", ref.Type())
		}
		wrong, err := export(t, in, "wrong").Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer wrong.Release()
		null, err := NullRef(ref.Type().WithNull())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := NullRef(ref.Type()); !errors.Is(err, ErrType) {
			t.Fatal("nonnull null manufactured", err)
		}
		table, global := export(t, in, "table"), export(t, in, "global")
		callback, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.ExternRef.NonNull()}, Results: []wasm.ValueType{wasm.ExternRef.NonNull()}}, func(c *Caller, args []Value) ([]Value, error) {
			if err := c.Collect(); err != nil {
				return nil, err
			}
			r, err := c.Call(inc, []Value{I32(41)})
			if err != nil || bits(t, r.Values[0]) != 42 {
				t.Fatal("typed callback reentry", r, err)
			}
			return args, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		host, err := callback.Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer host.Release()
		s.nextExtern = 1 << 40
		object, err := s.NewExternRef("retained typed object")
		if err != nil {
			t.Fatal(err)
		}
		defer object.Release()
		externNull, _ := NullRef(wasm.ExternRef)
		for _, metered := range []bool{false, true} {
			opts := CallOptions{Metered: metered, Fuel: 1_000_000}
			call := func(name string, args ...Value) CallResult {
				t.Helper()
				r, err := export(t, in, name).Call(nil, args, opts)
				if err != nil {
					t.Fatal(name, err)
				}
				return r
			}
			if bits(t, call("apply", I32(41), ref).Values[0]) != 42 {
				t.Fatal("typed call")
			}
			for _, arg := range []Value{wrong, externNull} {
				if _, err := export(t, in, "apply").Call(nil, []Value{I32(41), arg}, opts); !errors.Is(err, ErrType) {
					t.Fatal("wrong heap type", err)
				}
			}
			for _, name := range []string{"apply", "nonnull"} {
				args := []Value{null}
				if name == "apply" {
					args = []Value{I32(0), null}
				}
				_, err := export(t, in, name).Call(nil, args, opts)
				var trap *wasm.Trap
				if !errors.As(err, &trap) || trap.Code != wasm.TrapNullReference {
					t.Fatal("null trap", name, err)
				}
			}
			r := call("nonnull", ref)
			if r.Values[0].Type().Nullable() {
				t.Fatal("lost nonnull result")
			}
			r.Values[0].Release()
			for _, name := range []string{"echo", "nonnull-branch"} {
				r := call(name, object)
				v, err := s.ExternRef(r.Values[0])
				if err != nil || v != "retained typed object" {
					t.Fatal(name, v, err)
				}
				r.Values[0].Release()
			}
			if bits(t, call("null-branch", object).Values[0]) != 1 || bits(t, call("null-branch", externNull).Values[0]) != 0 {
				t.Fatal("64-bit reference null check")
			}
			if _, err := export(t, in, "echo").Call(nil, []Value{externNull}, opts); !errors.Is(err, ErrType) {
				t.Fatal("null nonnullable host argument", err)
			}
			r = call("host-echo", object, host)
			v, err := s.ExternRef(r.Values[0])
			if err != nil || v != "retained typed object" {
				t.Fatal("typed host roots", v, err)
			}
			r.Values[0].Release()
			if bits(t, call("tail", I32(20000), I32(22)).Values[0]) != 20022 {
				t.Fatal("tail reference recursion")
			}
		}
		for _, ex := range []Extern{table, global} {
			if ex.Kind() == TableKind {
				r, err := ex.TableGet(1)
				if err != nil {
					t.Fatal(err)
				}
				r.Release()
				if err := ex.TableSet(0, wrong); !errors.Is(err, ErrType) {
					t.Fatal("wrong table function", err)
				}
				if err := ex.TableSet(0, null); !errors.Is(err, ErrType) {
					t.Fatal("nullable table write", err)
				}
				if err := ex.TableSet(0, ref); err != nil {
					t.Fatal(err)
				}
			} else {
				r, err := ex.GlobalGet()
				if err != nil {
					t.Fatal(err)
				}
				r.Release()
				if err := ex.GlobalSet(wrong); !errors.Is(err, ErrType) {
					t.Fatal("wrong global function", err)
				}
				if err := ex.GlobalSet(ref); err != nil {
					t.Fatal(err)
				}
			}
		}
		typ := TableType{Element: ref.Type(), Limits: MemoryType{Min: 1, Max: 2, HasMax: true}}
		if _, err := s.NewTable(typ); err == nil {
			t.Fatal("nonnullable default initializer")
		}
		if _, err := s.NewTable(typ, wrong); !errors.Is(err, ErrType) {
			t.Fatal("wrong initializer", err)
		}
		if _, err := s.NewTable(typ, ref, ref); !errors.Is(err, ErrType) {
			t.Fatal("multiple initializers", err)
		}
		ht, err := s.NewTable(typ, ref)
		if err != nil {
			t.Fatal(err)
		}
		if old, ok, err := ht.GrowTable(1, ref); err != nil || !ok || old != 1 {
			t.Fatal(old, ok, err)
		}
		if _, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.FunctionReference(0, true)}}, func(*Caller, []Value) ([]Value, error) { return nil, nil }); err == nil {
			t.Fatal("module-local host signature")
		}
		l := NewLinker(s)
		if err := l.Define("p", "apply", export(t, in, "apply")); err != nil {
			t.Fatal(err)
		}
		imported, err := e.Compile(nil, other)
		if err != nil {
			t.Fatal(err)
		}
		defer imported.Close()
		linked := instantiate(t, l, imported)
		result, err := export(t, linked, "run").Call(nil, []Value{I32(41), ref}, CallOptions{})
		if err != nil || bits(t, result.Values[0]) != 42 {
			t.Fatal("module type indices were compared directly", result, err)
		}
		if _, err := export(t, in, "tail").Call(nil, []Value{I32(20000), I32(0)}, CallOptions{Metered: true, Fuel: 10}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("tail fuel", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, err := export(t, in, "spin").Call(ctx, nil, CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("cancel", err)
		}
		foreign, err := e.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer foreign.Close()
		foreignFn, err := foreign.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, func(_ *Caller, args []Value) ([]Value, error) { return args, nil })
		if err != nil {
			t.Fatal(err)
		}
		foreignRef, err := foreignFn.Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer foreignRef.Release()
		if _, err := export(t, in, "apply").Call(nil, []Value{I32(1), foreignRef}, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal("foreign reference", err)
		}
		stale, err := s.Root(ref)
		if err != nil {
			t.Fatal(err)
		}
		stale.Release()
		if _, err := export(t, in, "apply").Call(nil, []Value{I32(1), stale}, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal("stale reference", err)
		}
	})
}
