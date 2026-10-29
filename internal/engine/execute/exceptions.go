package execute

import (
	"github.com/cpunion/wasmvm/internal/engine/state"
	"sort"
)

// raise searches immutable handlers at the throw/call PC, then discards frames
// above the selected target. Caller continuations may already be outside a try;
// their saved call PCs, rather than their return PCs, determine protection.
func (e *Execution) raise(ex *state.Exception) bool {
	tag, args, err := e.store.Exception(ex.Ref)
	if err != nil || ex.Owner != e.store {
		e.fail(state.ErrReference)
		return false
	}
	for fi := len(e.frames) - 1; fi >= 0; fi-- {
		f := &e.frames[fi]
		code := f.function.Code
		pc := f.pc
		if f.suspended {
			pc = f.callPC
		}
		i := sort.Search(len(code.Protected), func(i int) bool { return code.Protected[i].End > pc })
		if i == len(code.Protected) || code.Protected[i].Start > pc {
			continue
		}
		for h := code.Protected[i].Handler; h != 0; h = code.Handlers[h-1].Parent {
			for _, c := range code.Handlers[h-1].Catches {
				if c.Kind < 2 && f.function.Instance.Tags[c.Tag] != tag {
					continue
				}
				if c.Kind < 2 {
					for i, v := range args {
						e.values[f.base+int(c.Dest[i])] = v
					}
				}
				if c.Kind&1 != 0 {
					e.values[f.base+int(c.Dest[len(c.Dest)-1])].Lo = ex.Ref
					e.values[f.base+int(c.Dest[len(c.Dest)-1])].Hi = 0
				}
				f.pc, f.suspended = c.PC, false
				clear(e.frames[fi+1:])
				e.frames = e.frames[:fi+1]
				e.values = e.values[:f.base+len(code.Slots)]
				return true
			}
		}
	}
	e.fail(ex)
	return false
}
