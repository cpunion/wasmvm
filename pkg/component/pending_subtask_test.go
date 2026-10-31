package component

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPendingSubtaskExecution(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			wide := asyncInstance(t, backend, "pending-memory64", Imports{})
			wideCall, _ := wide.Function("churn")
			if v, err := wideCall.Call(nil, []Value{{Bits: 17}}, CallOptions{}); err != nil || v.Value == nil || v.Value.Bits != 17 {
				t.Fatal("memory64 pending", v, err)
			}
			high, _ := wide.Function("high")
			if _, err := high.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("guest constant pointer truncated", err)
			}
			for _, ptr := range []uint64{0, 4, 65528, 2, 65532, 1 << 32} {
				fresh := asyncInstance(t, backend, "pending-memory64", Imports{})
				poll, _ := fresh.Function("poll")
				v, err := poll.Call(nil, []Value{{Bits: ptr}}, CallOptions{})
				if ptr <= 65528 && ptr&3 == 0 {
					if err != nil || v.Value == nil || v.Value.Bits != 0 {
						t.Fatal(ptr, v, err)
					}
				} else if !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("memory64 event bounds", ptr, err)
				}
			}
			synchronous := asyncInstance(t, backend, "pending-sync-trap", Imports{})
			syncCall, _ := synchronous.Function("churn")
			if _, err := syncCall.Call(nil, []Value{{Bits: 1}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("synchronous pending call", err)
			}
			config := DefaultConfig()
			config.Core.Backend = backend
			config.MaxTasks = 1
			config.MaxResources = 16
			engine, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			compiled, err := engine.Compile(nil, fixture(t, "pending-subtasks"))
			if err != nil {
				t.Fatal(err)
			}
			defer compiled.Close()
			limited, err := compiled.Instantiate(nil, Imports{}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer limited.Close()
			limit, _ := limited.Function("bad")
			if _, err := limit.Call(nil, []Value{{Bits: 9}}, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("task quota", err)
			}
			callback := asyncInstance(t, backend, "pending-callback", Imports{})
			cb, _ := callback.Function("run")
			for _, n := range []uint64{0, 1, 42, 1000} {
				v, err := cb.Call(nil, []Value{{Bits: n}}, CallOptions{})
				if err != nil || v.Value == nil || v.Value.Bits != n+7 {
					t.Fatal(v, err)
				}
			}
			for _, close := range []bool{false, true} {
				deadlock := asyncInstance(t, backend, "pending-deadlock", Imports{})
				f, _ := deadlock.Function("run")
				duration := 20 * time.Millisecond
				if close {
					duration = 2 * time.Second
				}
				ctx, cancel := context.WithTimeout(context.Background(), duration)
				var timer *time.Timer
				if close {
					timer = time.AfterFunc(20*time.Millisecond, deadlock.Close)
				}
				_, err := f.Call(ctx, []Value{{}}, CallOptions{})
				if timer != nil {
					timer.Stop()
				}
				cancel()
				if close {
					if !errors.Is(err, wasm.ErrClosed) && !errors.Is(err, context.Canceled) {
						t.Fatal(err)
					}
				} else if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
				if deadlock.tasks != nil {
					t.Fatal("deadlock retained continuation")
				}
			}
			for _, name := range []string{"pending-strings", "pending-strings-utf16", "pending-strings-compact"} {
				strings := asyncInstance(t, backend, name, Imports{})
				f, _ := strings.Function("echo")
				for _, text := range []string{"", "ascii", "café", "雪😀"} {
					v, err := f.Call(nil, []Value{{Text: text}}, CallOptions{})
					if err != nil || v.Value == nil || v.Value.Text != text {
						t.Fatal(name, text, v, err)
					}
				}
			}
			in := asyncInstance(t, backend, "pending-subtasks", Imports{})
			multi, _ := in.Function("multi")
			for _, n := range []uint64{0, 1, 42} {
				v, err := multi.Call(nil, []Value{{Bits: n}}, CallOptions{})
				if err != nil || v.Value == nil || v.Value.Bits != 3*n+24 {
					t.Fatal("concurrent callbacks", v, err)
				}
			}
			f, _ := in.Function("churn")
			for _, n := range []uint64{0, 1, 17, 128, 1000} {
				v, err := f.Call(nil, []Value{{Bits: n}}, CallOptions{})
				if err != nil || v.Value == nil || v.Value.Bits != n {
					t.Fatal(n, v, err)
				}
				if in.tasks != nil {
					t.Fatal("session retained tasks")
				}
			}
			for _, fixture := range []string{"pending-subtasks", "pending-cancel-return", "pending-cancel-deferred"} {
				for _, async := range []uint64{0, 1} {
					in := asyncInstance(t, backend, fixture, Imports{})
					f, _ := in.Function("cancel")
					v, err := f.Call(nil, []Value{{Bits: async}}, CallOptions{})
					want := uint64(4)
					if fixture == "pending-cancel-return" {
						want = 2
					}
					if fixture == "pending-cancel-deferred" {
						want = 2
						if async != 0 {
							want = uint64(^uint32(0))
						}
					}
					if err != nil || v.Value == nil || v.Value.Bits != want {
						t.Fatal(fixture, async, v, err)
					}
				}
			}
			for mode := uint64(0); mode <= 8; mode++ {
				in := asyncInstance(t, backend, "pending-subtasks", Imports{})
				f, _ := in.Function("bad")
				if _, err := f.Call(nil, []Value{{Bits: mode}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal(mode, err)
				}
				if in.tasks != nil {
					t.Fatal("failed session retained tasks")
				}
			}
		})
	}
}
