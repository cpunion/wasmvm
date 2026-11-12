package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// NewErrorContext constructs an immutable, copyable diagnostic value. Its
// message is UTF-8; lowering it allocates a distinct guest handle. Copying or
// dropping a guest handle does not invalidate other occurrences or host values.
func NewErrorContext(message string) (Value, error) { return canonical.NewErrorContext(message) }

// ErrorContextMessage reads a diagnostic value without entering its old Store.
// Debug strings must not be used to determine program behavior.
func ErrorContextMessage(value Value) (string, error) { return canonical.ErrorContextMessage(value) }

func (in *Instance) errorContextBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	fn, err := caller.NewScalarFunction(plan.Signature, func(c *vm.Caller, words []uint64) error {
		if n.failed || n.blocked {
			return wasm.ErrTrap
		}
		bits := words[0]
		switch plan.Definition.Op {
		case 0x1e:
			return n.handles.DropErrorContext(uint32(bits))
		case 0x1c:
			message := ""
			if !in.config.DiscardErrorContextMessages {
				units := words[1]
				var err error
				message, err = in.codec(c, n, plan).Transfer.LiftString(bits, units)
				if err != nil {
					return err
				}
			}
			value, err := NewErrorContext(message)
			if err != nil {
				return err
			}
			id, err := n.handles.AddErrorContext(value)
			if err != nil {
				return err
			}
			words[0] = uint64(id)
			return nil
		default:
			value, err := n.handles.ErrorContext(uint32(bits))
			if err != nil {
				return err
			}
			message, _ := ErrorContextMessage(value)
			ptr := words[1]
			return in.codec(c, n, plan).Store(&model.Type{Kind: model.String}, ptr, Value{Text: message})
		}
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}
