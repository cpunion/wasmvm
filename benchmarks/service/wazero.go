//go:build !llgo

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/cpunion/wasmvm/internal/bench/service"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// This is a benchmark contract check after independent engine validation.
func pureModule(wasm []byte) error {
	if len(wasm) < 8 || !bytes.Equal(wasm[:8], []byte{0, 'a', 's', 'm', 1, 0, 0, 0}) {
		return errors.New("expected a core Wasm version 1 module")
	}
	for cursor := 8; cursor < len(wasm); {
		id := wasm[cursor]
		cursor++
		size, width := binary.Uvarint(wasm[cursor:])
		if width <= 0 || size > math.MaxUint32 {
			return errors.New("invalid section length")
		}
		cursor += width
		if size > uint64(len(wasm)-cursor) {
			return errors.New("truncated section")
		}
		if id == 8 {
			return errors.New("pure modules must not have a start section")
		}
		if id == 2 {
			count, n := binary.Uvarint(wasm[cursor : cursor+int(size)])
			if n <= 0 || count != 0 {
				return errors.New("pure modules must not import host state")
			}
		}
		cursor += int(size)
	}
	return nil
}

func wazeroKernels(ctx context.Context, args options, out *report) (calls []service.Call, release func() error, err error) {
	start := time.Now()
	input, err := os.ReadFile(args.path)
	out.Phases["load"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, nil, err
	}
	start = time.Now()
	config := wazero.NewRuntimeConfigCompiler()
	if args.backend == "interpreter" {
		config = wazero.NewRuntimeConfigInterpreter()
	}
	r := wazero.NewRuntimeWithConfig(ctx, config.WithCloseOnContextDone(args.termination))
	out.Phases["engineSetup"] = time.Since(start).Nanoseconds()
	release = func() error { return r.Close(context.Background()) }
	start = time.Now()
	compiled, err := r.CompileModule(ctx, input)
	out.Phases["compileValidate"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, release, err
	}
	release = func() error { return errors.Join(compiled.Close(context.Background()), r.Close(context.Background())) }
	if err = pureModule(input); err != nil {
		return nil, release, err
	}
	for i := range args.Workers {
		start = time.Now()
		module, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName(fmt.Sprintf("worker-%d", i)).WithStartFunctions())
		out.Phases["instantiate"] += time.Since(start).Nanoseconds()
		if err != nil {
			return nil, release, err
		}
		start = time.Now()
		function := module.ExportedFunction("run")
		out.Phases["exportLookup"] += time.Since(start).Nanoseconds()
		if function == nil {
			return nil, release, errors.New("missing run export")
		}
		params, results := function.Definition().ParamTypes(), function.Definition().ResultTypes()
		if len(params) != 2 || params[0] != api.ValueTypeI32 || params[1] != api.ValueTypeI32 || len(results) != 1 || results[0] != api.ValueTypeI32 {
			return nil, release, errors.New("run must have signature (i32, i32) -> i32")
		}
		stack := make([]uint64, 2)
		calls = append(calls, func() (int64, uint32, error) {
			stack[0], stack[1] = uint64(args.seed), uint64(args.iterations)
			start := time.Now()
			err := function.CallWithStack(ctx, stack)
			ns := time.Since(start).Nanoseconds()
			value := uint32(stack[0])
			if err != nil {
				return ns, value, err
			}
			if value != args.Expected {
				return ns, value, fmt.Errorf("wazero checksum: got %d, expected %d", value, args.Expected)
			}
			return ns, value, nil
		})
	}
	return calls, release, nil
}
