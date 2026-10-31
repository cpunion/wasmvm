package vm

import (
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestContinuationResumedHostExceptionIsCaught(t *testing.T) {
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
		sc := DefaultStoreConfig()
		sc.CollectEveryAllocation = true
		s, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		typ := FunctionType{Params: []wasm.ValueType{wasm.I32, wasm.ExternRef}}
		tag, err := s.NewTag(typ)
		if err != nil {
			t.Fatal(err)
		}
		ready := false
		var exception *Exception
		host, err := s.NewSuspendableFunction(typ, func(_ *Caller, args []Value) (HostResult, error) {
			return HostResult{Pending: &HostSuspension{Ready: func() bool { return ready }, Resume: func(c *Caller) (HostResult, error) {
				var err error
				exception, err = c.NewException(tag, args)
				if err != nil {
					return HostResult{}, err
				}
				if err = c.Collect(); err != nil {
					return HostResult{}, err
				}
				return HostResult{}, exception
			}}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		l := NewLinker(s)
		l.Define("h", "tag", tag)
		l.Define("h", "host", host)
		in := instantiate(t, l, m)
		fn := export(t, in, "catch-host")
		payload, err := s.NewExternRef("parked exception payload")
		if err != nil {
			t.Fatal(err)
		}
		var result CallResult
		_, err = s.WithCaller(nil, CallOptions{Metered: true, Fuel: 1000}, func(c *Caller) error {
			call, err := c.Start(fn, []Value{I32(42), payload})
			if err != nil {
				return err
			}
			if _, done, err := c.Poll(call); err != nil || done {
				t.Fatal(done, err)
			}
			payload.Release()
			if err := c.Collect(); err != nil {
				return err
			}
			ready = true
			var done bool
			result, done, err = c.Poll(call)
			if err != nil || !done {
				t.Fatal(done, err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if bits(t, result.Values[0]) != 42 {
			t.Fatal(result)
		}
		object, err := s.ExternRef(result.Values[1])
		if err != nil || object != "parked exception payload" {
			t.Fatal(object, err)
		}
		result.Values[1].Release()
		exception.Release()
		usage, err := s.Usage()
		if err != nil || usage.HostRoots != 0 {
			t.Fatal("exception roots", usage, err)
		}
	})
}
