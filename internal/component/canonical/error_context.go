package canonical

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
	"unicode/utf8"
)

// Error contexts are immutable, copyable diagnostic values. Their handles use
// the same tagged index space as resources, without resource ownership moves.
type errorContext struct{ message string }

func NewErrorContext(message string) (Value, error) {
	if uint64(len(message)) >= 1<<28 || !utf8.ValidString(message) {
		return Value{}, invalid("error-context debug message")
	}
	return Value{Resource: &Reference{errorContext: &errorContext{message: message}}}, nil
}

func ErrorContextMessage(value Value) (string, error) {
	if value.Resource == nil || value.Resource.errorContext == nil || value.Resource.state != nil {
		return "", invalid("error-context value")
	}
	return value.Resource.errorContext.message, nil
}

func SameErrorContext(a, b *Reference) bool {
	return a != nil && b != nil && a.errorContext != nil && a.errorContext == b.errorContext
}

func (b *ResourceBudget) retainError(e *errorContext) error {
	b.errorMu.Lock()
	defer b.errorMu.Unlock()
	if b.errors[e] == ^uint32(0) {
		return exhausted()
	}
	if b.errors[e] == 0 {
		if uint64(len(e.message)) > b.errorLimit-b.errorBytes {
			return exhausted()
		}
		if b.errors == nil {
			b.errors = map[*errorContext]uint32{}
		}
		b.errorBytes += uint64(len(e.message))
	}
	b.errors[e]++
	return nil
}

func (b *ResourceBudget) releaseError(e *errorContext) {
	b.errorMu.Lock()
	defer b.errorMu.Unlock()
	b.errors[e]--
	if b.errors[e] == 0 {
		delete(b.errors, e)
		b.errorBytes -= uint64(len(e.message))
	}
}

func (r *Resources) AddErrorContext(value Value) (uint32, error) {
	if _, err := ErrorContextMessage(value); err != nil {
		return 0, err
	}
	return r.addErrorContext(value.Resource.errorContext)
}

func (r *Resources) addErrorContext(data *errorContext) (uint32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return 0, wasm.ErrClosed
	}
	if err := r.budget.retainError(data); err != nil {
		return 0, err
	}
	id, err := r.addLocked(&resourceEntry{errorContext: data})
	if err != nil {
		r.budget.releaseError(data)
	} else {
		r.errorCount++
	}
	return id, err
}

func (r *Resources) ErrorContext(id uint32) (Value, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, err := r.errorEntry(id)
	if err != nil {
		return Value{}, err
	}
	return Value{Resource: &Reference{errorContext: e.errorContext}}, nil
}

func (r *Resources) errorEntry(id uint32) (*resourceEntry, error) {
	if r.closed {
		return nil, wasm.ErrClosed
	}
	e := r.entries[id]
	if e == nil || e.errorContext == nil {
		return nil, scalarTrap("invalid error-context handle")
	}
	return e, nil
}

func (r *Resources) DropErrorContext(id uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, err := r.errorEntry(id)
	if err != nil {
		return err
	}
	delete(r.entries, id)
	r.errorCount--
	r.budget.releaseError(e.errorContext)
	r.budget.release(1)
	return nil
}

// snapshotErrorContext detaches the public reference carrier while preserving
// the immutable diagnostic identity. A host may reassign its original carrier.
func snapshotErrorContext(value Value) Value {
	return Value{Resource: &Reference{errorContext: value.Resource.errorContext}}
}
