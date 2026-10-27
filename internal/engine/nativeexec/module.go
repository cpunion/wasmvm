// Package nativeexec connects owned generated code to the common frame machine.
// Calls, host requests and explicit runtime helpers never interpret a function.
package nativeexec

import (
	"context"
	"math"
	"runtime"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Module struct {
	code    *native.Code
	entries map[*slots.Function][2]uint32
}
type Lease struct {
	code    *native.Lease
	entries map[*slots.Function][2]uint32
}

// Compile prepares both policy tiers within one aggregate code-byte budget.
// Publication is atomic: no executable allocation survives a failed compile.
func Compile(ctx context.Context, m *slots.Module, limits codegen.Limits) (*Module, error) {
	if !native.Available {
		return nil, wasm.ErrUnsupported
	}
	out := &Module{entries: make(map[*slots.Function][2]uint32, len(m.Functions))}
	var code []byte
	var entries []uint32
	definition := m.IR.Validated.Definition()
	var minimum uint32
	for _, im := range definition.Imports {
		if im.Kind == 2 {
			minimum = im.Memory.Min
		}
	}
	if len(definition.Memories) != 0 {
		minimum = definition.Memories[0].Min
	}
	for _, f := range m.Functions {
		var indexes [2]uint32
		for tier := range indexes {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			bounded := limits
			bounded.MaxCodeBytes = min(limits.MaxCodeBytes, uint64(math.MaxUint32)) - uint64(len(code))
			body, err := codegen.CompileFunctionWithConfig(runtime.GOARCH, f, bounded, codegen.Config{Unmetered: tier == 0, MemoryMinimumPages: minimum})
			if err != nil {
				return nil, err
			}
			indexes[tier] = uint32(len(entries))
			entries = append(entries, uint32(len(code)))
			code = append(code, body.Entry.Code...)
		}
		out.entries[f] = indexes
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(entries) != 0 {
		var err error
		out.code, err = native.Publish(code, entries, limits.MaxCodeBytes)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (m *Module) Acquire() (*Lease, error) {
	l := &Lease{entries: m.entries}
	if m.code != nil {
		var err error
		l.code, err = m.code.Acquire()
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}
func (m *Module) Close() {
	if m.code != nil {
		m.code.Close()
	}
}
func (l *Lease) Close() {
	if l.code != nil {
		l.code.Close()
	}
}
