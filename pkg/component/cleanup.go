package component

import "github.com/cpunion/wasmvm/internal/component/canonical"

// Cleanup completes every independent release before propagating a panic.
// A panic already being unwound takes precedence over cleanup callbacks.
type cleanupGuard struct{ panicked any }

func (g *cleanupGuard) run(fn func()) {
	defer func() {
		if p := recover(); p != nil && g.panicked == nil {
			g.panicked = p
		}
	}()
	fn()
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
	guard.run(func() { *err = finishResources(codec, *complete, *err) })
	if *err != nil || guard.panicked != nil {
		n.failed = true
	}
	guard.finish()
}
