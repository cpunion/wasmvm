package vm

import (
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/nativeexec"
)

// One idle outer-call workspace belongs to the Store. Nested host reentry uses
// an independent workspace and its remaining logical stack/slot limits.
type callWorkspace struct {
	execution *execute.Execution
	native    *nativeexec.Execution
}

func (op *operation) acquireWorkspace(limits execute.Limits) (*callWorkspace, error) {
	s := op.store
	if op.depth == 1 {
		s.mu.Lock()
		w := s.workspace
		if w != nil {
			s.workspace = nil
			var err error
			if w.native != nil {
				err = w.native.BeginOwned(op.ctx, op.options.Metered)
			}
			if err == nil {
				err = w.execution.Rebind(op.ctx, limits, op.options.Metered, op.fuel)
			}
			if err == nil {
				op.frames = append(op.frames, w.execution)
			}
			s.mu.Unlock()
			if err != nil {
				w.execution.Close()
				return nil, err
			}
			return w, nil
		}
		s.mu.Unlock()
	}
	config := execute.Config{Limits: limits, Metered: op.options.Metered, Fuel: op.fuel}
	w := &callWorkspace{}
	if s.engine.config.Backend == "native" {
		backend, err := nativeexec.NewOwned(op.ctx, s.nativeModules, limits.MaxSlots, limits.MaxFrames, op.options.Metered)
		if err != nil {
			return nil, err
		}
		w.native, config.Backend = backend, backend
	}
	w.execution = execute.New(s.state, op.ctx, config)
	op.addFrame(w.execution)
	return w, nil
}

func (op *operation) releaseWorkspace(w *callWorkspace) {
	s := op.store
	s.mu.Lock()
	op.removeFrameLocked(w.execution)
	reusable := true
	if w.native != nil {
		reusable = w.native.Suspend()
	}
	cached := op.depth == 1 && reusable && !s.closing
	if cached {
		s.workspace = w
	}
	s.mu.Unlock()
	// This operation still owns Store admission. No later operation or session
	// call can borrow the cache before this synchronous release returns.
	w.execution.Quiesce()
	if !cached {
		w.execution.Close()
	}
}
