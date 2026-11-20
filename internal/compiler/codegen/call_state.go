package codegen

import "github.com/cpunion/wasmvm/internal/compiler/nativeabi"

// These callee-saved words survive every same-module private call. Ordinary
// entry seeds them before header checks; every exit publishes their current
// values. Ordinary allocation excludes them. A guarded ARM64 loop can reuse
// them while the call state is saved on the native stack.
const callUsed, callDepth, callTop = byte(8), byte(9), byte(10)

func (e *functionEmitter) loopCallRegisters(start int, r memoryRegion) []byte {
	if e.a.arch != "arm64" || !e.frameRegisters || r.loop == nil || r.loop.memory == nil || start%256 == 0 {
		return nil
	}
	// The memory-loop proof excludes trapping scalars, calls and interior
	// polls, and the entry guards cover every load and store in the body.
	return []byte{e.callReg(callUsed), e.callReg(callDepth), e.callReg(callTop)}
}

func (e *functionEmitter) usesLoopCallRegisters(plan *registerPlan) bool {
	if plan == nil || !e.frameRegisters || e.a.arch != "arm64" {
		return false
	}
	for _, reg := range plan.homes {
		if reg == e.callReg(callUsed) || reg == e.callReg(callDepth) || reg == e.callReg(callTop) {
			return true
		}
	}
	return false
}

// A guarded ARM64 loop can temporarily use the persistent call registers.
// Leave room for the three raw words while retaining 16-byte stack alignment.
func (e *functionEmitter) saveLoopCallRegisters() {
	e.a.word(0xa9be73fb) // stp x27,x28,[sp,#-32]!
	e.a.word(0xf9000bfa) // str x26,[sp,#16]
}

func (e *functionEmitter) restoreLoopCallRegisters() {
	e.a.word(0xf9400bfa) // ldr x26,[sp,#16]
	e.a.word(0xa8c273fb) // ldp x27,x28,[sp],#32
}

func (e *functionEmitter) callTopLoad(n byte) {
	if e.frameRegisters {
		e.callMove(n, callTop)
	} else {
		e.callLoad(n, e.callReg(2), nativeabi.CallTopOffset)
	}
}

func (e *functionEmitter) callTopEntry() {
	if !e.frameRegisters {
		return
	}
	missing := e.a.label()
	e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
	e.callTest(2)
	e.branch(missing, "eq")
	e.callLoad(callTop, e.callReg(2), nativeabi.CallTopOffset)
	e.a.mark(missing)
}

func (e *functionEmitter) callTopExit() {
	if !e.frameRegisters {
		return
	}
	missing := e.a.label()
	// Ordinary entry initializes the sentinel before any header check, so
	// malformed records never publish an uninitialized top or dereference it.
	e.callConst(0, ^uint64(0))
	e.callCompare(callTop, 0)
	e.branch(missing, "eq")
	e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
	e.callTest(2)
	e.branch(missing, "eq")
	e.callStore(callTop, e.callReg(2), nativeabi.CallTopOffset)
	e.a.mark(missing)
}

func (e *functionEmitter) callStateLoad(n byte, offset uintptr) {
	if e.frameRegisters {
		state := callUsed
		if offset == nativeabi.DepthOffset {
			state = callDepth
		}
		e.callMove(n, state)
		return
	}
	e.callLoad(n, e.callContext(), uint32(offset))
}

func (e *functionEmitter) callPushState() {
	bytes := uint64(len(e.f.Slots)) * 16
	if e.frameRegisters {
		if bytes != 0 {
			if bytes <= 4095 || e.a.arch == "amd64" {
				e.callAdvance(callUsed, int(bytes))
			} else {
				e.callConst(0, bytes)
				e.callAdd(callUsed, callUsed, 0)
			}
		}
		e.callAdvance(callDepth, 1)
		return
	}
	e.callLoad(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callConst(1, bytes)
	e.callAdd(0, 0, 1)
	e.callStore(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callLoad(0, e.callContext(), uint32(nativeabi.DepthOffset))
	e.callAdvance(0, 1)
	e.callStore(0, e.callContext(), uint32(nativeabi.DepthOffset))
}

func (e *functionEmitter) callPopState() {
	if e.frameRegisters {
		e.callLoad(callUsed, e.callReg(3), nativeabi.CallUsedOffset)
		e.callAdvance(callDepth, -1)
		return
	}
	e.callLoad(0, e.callReg(3), nativeabi.CallUsedOffset)
	e.callStore(0, e.callContext(), uint32(nativeabi.UsedOffset))
	e.callLoad(0, e.callContext(), uint32(nativeabi.DepthOffset))
	e.callAdvance(0, -1)
	e.callStore(0, e.callContext(), uint32(nativeabi.DepthOffset))
}
