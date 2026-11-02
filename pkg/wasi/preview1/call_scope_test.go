//go:build linux || darwin

package preview1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestProcessCallEffectsRetainMetadataAndDrain(t *testing.T) {
	for _, memoryOnly := range []bool{false, true} {
		p := process(t, DefaultConfig())
		type key struct{}
		deadline := time.Now().Add(time.Minute)
		parent, cancel := context.WithDeadline(context.WithValue(context.Background(), key{}, 42), deadline)
		ctx, finish, err := p.begin(parent, memoryOnly)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := ctx.Deadline()
		if !ok || !got.Equal(deadline) || ctx.Value(key{}) != 42 || ctx.Err() != nil {
			t.Fatal("call metadata", got, ok, ctx.Err())
		}
		finish()
		if parent.Err() != nil || p.ctx.Err() != nil {
			t.Fatal("finishing a call cancelled its owner")
		}
		if !memoryOnly && !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatal("I/O call context did not retire", ctx.Err())
		}
		ctx, finish, err = p.begin(parent, memoryOnly)
		if err != nil || ctx.Err() != nil {
			t.Fatal("new call inherited retired cancellation", err)
		}
		p.Close()
		select {
		case <-p.done:
			t.Fatal("closed before active callback drained")
		default:
		}
		if !memoryOnly {
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
				t.Fatal("process close did not interrupt I/O")
			}
		}
		finish()
		if err := p.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
		if _, _, err := p.begin(parent, memoryOnly); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("closed process admitted call", err)
		}
		cancel()
	}
}
