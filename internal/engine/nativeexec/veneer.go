package nativeexec

import "github.com/cpunion/wasmvm/internal/engine/state"

// CanRunFramePlan admits frame-only call/return plans for an owned artifact.
// Missing or closed artifacts retain their ordinary checked backend error.
// Workspace admission, interruption and retirement remain invocation-owned.
func (e *Execution) CanRunFramePlan(fn *state.Function) bool {
	if e.native == nil || e.ctx == nil || fn == nil || fn.Instance == nil {
		return false
	}
	module := e.modules[fn.Instance.Compiled]
	if module == nil || module.code.IsClosed() {
		return false
	}
	_, ok := module.entries[fn.Code]
	return ok
}
