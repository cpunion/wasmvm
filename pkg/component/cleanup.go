package component

import "github.com/cpunion/wasmvm/internal/component/canonical"

// Cleanup completes every independent release before propagating a panic.
// A panic already being unwound takes precedence over cleanup callbacks.
type cleanupGuard struct{ panicked any }

func (g *cleanupGuard) capturePanic() {
	if p := recover(); p != nil && g.panicked == nil {
		g.panicked = p
	}
}

func (g *cleanupGuard) run(fn func()) {
	defer g.capturePanic()
	fn()
}

func (g *cleanupGuard) finishCodec(codec *canonical.Codec, complete bool, err error) error {
	if codec == nil {
		return err
	}
	if codec.Resources == nil {
		return finishResources(codec, complete, err)
	}
	if complete && err == nil {
		if r, ok := codec.Resources.(*resourceCodec); ok {
			if ended, done := r.finishSuccess(); done {
				return ended
			} else if ended != nil {
				return g.finishResourceRollback(r, ended)
			}
		}
	}
	return g.finishResourceCodec(codec, complete, err)
}

func (g *cleanupGuard) finishResourceRollback(r *resourceCodec, err error) (result error) {
	defer g.capturePanic()
	result = err
	return r.finishRollback(err)
}

func (g *cleanupGuard) finishResourceCodec(codec *canonical.Codec, complete bool, err error) (result error) {
	defer g.capturePanic()
	result = err
	return finishResources(codec, complete, err)
}

func (g *cleanupGuard) finish() {
	if g.panicked != nil {
		panic(g.panicked)
	}
}

func cleanupCanonical(n *namespace, codec *canonical.Codec, complete *bool, err *error) {
	guard := cleanupGuard{panicked: recover()}
	if !*complete || *err != nil || guard.panicked != nil {
		n.failed = true
	}
	*err = guard.finishCodec(codec, *complete, *err)
	if *err != nil || guard.panicked != nil {
		n.failed = true
	}
	guard.finish()
}
