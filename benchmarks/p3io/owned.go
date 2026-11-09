//go:build !p3reference && !p3host

package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/cpunion/wasmvm/internal/bench/service"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasi/preview3"
)

func ownedKernels(ctx context.Context, args options, out *report) ([]service.Call, func() error, error) {
	start := time.Now()
	data, err := os.ReadFile(args.path)
	out.Phases["load"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, nil, err
	}
	config := component.DefaultConfig()
	config.Core.Backend = args.backend
	start = time.Now()
	engine, err := component.NewEngine(config)
	out.Phases["engineSetup"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, nil, err
	}
	var module *component.Compiled
	var instances []*component.Instance
	var worlds []*preview3.World
	release := func() error {
		for _, in := range instances {
			in.Close()
		}
		for _, w := range worlds {
			w.Close()
		}
		var err error
		for _, in := range instances {
			err = errors.Join(err, in.WaitClosed(context.Background()))
		}
		for _, w := range worlds {
			err = errors.Join(err, w.WaitClosed(context.Background()))
		}
		if module != nil {
			module.Close()
		}
		engine.Close()
		return err
	}
	start = time.Now()
	module, err = engine.Compile(ctx, data)
	out.Phases["compileValidate"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, release, err
	}
	calls := make([]service.Call, args.Workers)
	for i := range calls {
		start = time.Now()
		world, err := preview3.NewWorld(preview3.WorldConfig{})
		out.Phases["providerSetup"] += time.Since(start).Nanoseconds()
		if err != nil {
			return nil, release, err
		}
		worlds = append(worlds, world)
		start = time.Now()
		in, err := module.Instantiate(ctx, world.Imports(), component.CallOptions{})
		out.Phases["instantiate"] += time.Since(start).Nanoseconds()
		if err != nil {
			return nil, release, err
		}
		instances = append(instances, in)
		start = time.Now()
		_, err = in.Function(preview3.HTTPHandlerInterface, "handle")
		out.Phases["exportLookup"] += time.Since(start).Nanoseconds()
		if err != nil {
			return nil, release, err
		}
		if args.responseMode == "buffered" {
			calls[i] = bufferedKernel(ctx, args, func(ctx context.Context, r *http.Request) (*preview3.BufferedResponse, error) {
				return world.HTTP().CallIncomingBuffered(ctx, in, r, uint64(len(args.payload)), component.CallOptions{})
			})
		} else {
			calls[i] = httpKernel(ctx, args, func(ctx context.Context, r *http.Request) (*http.Response, error) {
				return world.HTTP().CallIncoming(ctx, in, r, component.CallOptions{})
			})
		}
	}
	out.Policy = "owned P3 HTTP CallIncoming, body collection and close; one World/Store per worker sharing compiled code; fuel off, cancellation/logical limits retained; per-call handler lookup/signature check included; no network socket"
	if args.responseMode == "buffered" {
		out.Policy = "owned P3 HTTP CallIncomingBuffered; bounded complete body/trailer collection and guest join inside one Store lease; one World/Store per worker sharing compiled code; fuel off, cancellation/logical limits retained; per-call handler lookup/signature check included; no network socket"
	}
	return calls, release, nil
}
