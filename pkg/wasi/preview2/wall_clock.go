package preview2

import (
	"context"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const WallClockInterface = "wasi:clocks/wall-clock@0.2.12"

// Datetime represents unsigned Unix seconds with a normalized nanosecond field.
// Its range does not depend on time.Time or signed nanosecond duration limits.
type Datetime struct {
	Seconds     uint64
	Nanoseconds uint32
}

// WallClockConfig supplies both callbacks or neither. Callbacks may run in
// parallel across Stores and must observe cancellation. A zero config uses
// the host realtime clock with explicit microsecond granularity and no timezone state.
type WallClockConfig struct {
	Now        func(context.Context) (Datetime, error)
	Resolution func(context.Context) (Datetime, error)
}

type WallClock struct{ config WallClockConfig }

func NewWallClock(config WallClockConfig) (*WallClock, error) {
	if config.Now == nil && config.Resolution == nil {
		config.Now = systemWallNow
		config.Resolution = systemWallResolution
	}
	if config.Now == nil || config.Resolution == nil {
		return nil, wasm.ErrInvalid
	}
	return &WallClock{config}, nil
}

// The default wall clock deliberately exposes microsecond ticks. Quantizing
// reads to the same granularity keeps the provider contract independent of the
// host Go/LLGo clock's representation and avoids signed UnixNano overflow.
func systemWallNow(ctx context.Context) (Datetime, error) {
	if err := ctx.Err(); err != nil {
		return Datetime{}, err
	}
	now := time.Now()
	if now.Unix() < 0 {
		return Datetime{}, wasm.ErrInvalid
	}
	return Datetime{uint64(now.Unix()), uint32(now.Nanosecond() / 1000 * 1000)}, ctx.Err()
}
func systemWallResolution(ctx context.Context) (Datetime, error) {
	return Datetime{Nanoseconds: 1000}, ctx.Err()
}

func wallClockType(f component.FunctionInfo) error {
	t := f.Result()
	if f.ParameterCount() != 0 || t.Kind() != component.KindRecord || t.FieldCount() != 2 {
		return vm.ErrType
	}
	seconds, sec := t.Field(0)
	nanoseconds, nano := t.Field(1)
	if seconds != "seconds" || sec.Kind() != component.KindU64 || nanoseconds != "nanoseconds" || nano.Kind() != component.KindU32 {
		return vm.ErrType
	}
	return nil
}

func (w *WallClock) read(ctx context.Context, resolution bool) (*component.Value, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fn := w.config.Now
	if resolution {
		fn = w.config.Resolution
	}
	t, err := fn(ctx)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if t.Nanoseconds >= 1000000000 || resolution && t.Seconds == 0 && t.Nanoseconds == 0 {
		return nil, wasm.ErrInvalid
	}
	return &component.Value{Elements: []component.Value{{Bits: t.Seconds}, {Bits: uint64(t.Nanoseconds)}}}, nil
}

// Imports exports every operation with exact field-name and numeric contracts.
func (w *WallClock) Imports() component.Imports {
	if w == nil || w.config.Now == nil || w.config.Resolution == nil {
		return component.Imports{}
	}
	functions := map[string]component.HostFunction{}
	checks := map[string]component.FunctionCheck{}
	for _, name := range []string{"now", "resolution"} {
		resolution := name == "resolution"
		functions[name] = func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			return w.read(c.Context(), resolution)
		}
		checks[name] = wallClockType
	}
	return component.Imports{Instances: map[string]component.Imports{WallClockInterface: {Functions: functions, FunctionChecks: checks}}}
}
