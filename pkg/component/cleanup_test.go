package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type cleanupResourceProbe struct {
	canonical.ResourceCodec
	finish func(error) error
}

func (p cleanupResourceProbe) Finish(err error) error { return p.finish(err) }

func TestCleanupGuardFinishesCodec(t *testing.T) {
	var empty cleanupGuard
	if err := empty.finishCodec(&canonical.Codec{}, false, nil); err != wasm.ErrTrap {
		t.Fatalf("incomplete codec cleanup returned %v", err)
	}

	original := errors.New("execution failed")
	cleanup := errors.New("cleanup failed")
	g := cleanupGuard{panicked: original}
	panicking := canonical.Codec{Resources: cleanupResourceProbe{finish: func(err error) error {
		if err != original {
			t.Errorf("cleanup received %v", err)
		}
		panic(cleanup)
	}}}
	if err := g.finishCodec(&panicking, false, original); err != original || g.panicked != original {
		t.Fatalf("cleanup replaced the original result or panic with %v and %v", err, g.panicked)
	}
	g = cleanupGuard{}
	if err := g.finishCodec(&panicking, false, original); err != original || g.panicked != cleanup {
		t.Fatalf("cleanup lost the result or first panic with %v and %v", err, g.panicked)
	}
	next := canonical.Codec{Resources: cleanupResourceProbe{finish: func(err error) error {
		if err != wasm.ErrTrap {
			t.Errorf("next cleanup received %v", err)
		}
		return original
	}}}
	if err := g.finishCodec(&next, false, nil); err != original || g.panicked != cleanup {
		t.Fatalf("next cleanup returned %v with panic %v", err, g.panicked)
	}
}
