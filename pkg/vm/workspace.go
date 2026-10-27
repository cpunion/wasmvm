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
	var w *callWorkspace
	if op.depth == 1 {
		s.mu.Lock()
		w, s.workspace = s.workspace, nil
		s.mu.Unlock()
	}
	config := execute.Config{Limits: limits, Metered: op.options.Metered, Fuel: op.fuel}
	if w != nil {
		if w.native != nil {
			if err := w.native.Begin(op.ctx, op.options.Metered); err != nil {
				w.execution.Close()
				return nil, err
			}
		}
		if err := w.execution.Rebind(op.ctx, limits, op.options.Metered, op.fuel); err != nil {
			w.execution.Close()
			return nil, err
		}
		return w, nil
	}
	w = &callWorkspace{}
	if s.engine.config.Backend == "native" {
		backend, err := nativeexec.New(op.ctx, s.nativeModules, limits.MaxSlots, limits.MaxFrames, op.options.Metered)
		if err != nil {
			return nil, err
		}
		w.native, config.Backend = backend, backend
	}
	w.execution = execute.New(s.state, op.ctx, config)
	return w, nil
}

func (op *operation) releaseWorkspace(w *callWorkspace) {
	reusable := true
	if w.native != nil {
		reusable = w.native.Suspend()
	}
	w.execution.Quiesce()
	if op.depth == 1 && reusable {
		s := op.store
		s.mu.Lock()
		if !s.closing {
			s.workspace, w = w, nil
		}
		s.mu.Unlock()
	}
	if w != nil {
		w.execution.Close()
	}
}
