// Package nativeexec connects owned generated code to the common frame machine.
// Calls, host requests and explicit runtime helpers never interpret a function.
package nativeexec

import (
	"context"
	"errors"
	"math"
	"runtime"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Module struct {
	code    *native.Code
	entries map[*slots.Function]functionEntries
}
type Lease struct {
	code    *native.Lease
	entries map[*slots.Function]functionEntries
}

type functionEntries struct {
	ordinary, metered, expanded uint32
	inlineSlots                 uint32
	hasExpansion                bool
}

// Compile prepares both policy tiers within one aggregate code-byte budget.
// Publication is atomic: no executable allocation survives a failed compile.
func Compile(ctx context.Context, m *slots.Module, limits codegen.Limits) (*Module, error) {
	if !native.Available {
		return nil, wasm.ErrUnsupported
	}
	out := &Module{entries: make(map[*slots.Function]functionEntries, len(m.Functions))}
	var code []byte
	var entries []uint32
	definition := m.IR.Validated.Definition()
	var minimum uint32
	var importedFunctions uint32
	for _, im := range definition.Imports {
		if im.Kind == 0 {
			importedFunctions++
		}
		if im.Kind == 2 {
			minimum = im.Memory.Min
		}
	}
	if len(definition.Memories) != 0 {
		minimum = definition.Memories[0].Min
	}
	callees := make(map[uint32]*slots.Function, len(m.Functions))
	for i, f := range m.Functions {
		callees[importedFunctions+uint32(i)] = f
	}
	wantExpansion := map[*slots.Function]bool{}
	// Publish all required ordinary tiers before spending spare code budget
	// on optional expansion, including when later functions are large.
	for tier := 0; tier < 3; tier++ {
		for _, f := range m.Functions {
			indexes := out.entries[f]
			if tier == 2 && !wantExpansion[f] {
				continue
			}
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			bounded := limits
			bounded.MaxCodeBytes = min(limits.MaxCodeBytes, uint64(math.MaxUint32)) - uint64(len(code))
			config := codegen.Config{Unmetered: tier != 1, MemoryMinimumPages: minimum}
			if tier == 2 {
				config.DirectCallees = callees
			}
			body, err := codegen.CompileFunctionWithConfig(runtime.GOARCH, f, bounded, config)
			if tier == 2 && errors.Is(err, wasm.ErrLimit) {
				// The optional expansion cannot consume the ordinary tiers'
				// code budget or make an otherwise valid module unavailable.
				continue
			}
			if err != nil {
				return nil, err
			}
			if tier == 2 && body.InlinedCalls == 0 {
				continue
			}
			index := uint32(len(entries))
			switch tier {
			case 0:
				indexes.ordinary = index
				wantExpansion[f] = body.Calls != 0
			case 1:
				indexes.metered = index
			case 2:
				indexes.expanded, indexes.inlineSlots, indexes.hasExpansion = index, body.InlineSlots, true
			}
			entries = append(entries, uint32(len(code)))
			code = append(code, body.Entry.Code...)
			out.entries[f] = indexes
		}
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
