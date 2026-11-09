package phase

import (
	"context"
	"errors"
	"time"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

type serviceReport = service.Report

func serviceConfig(args options) service.Options {
	return service.Options{Workers: args.workers, Clients: args.clients, Requests: args.requests,
		Samples: args.samples, Warmups: args.warmups, Expected: args.expected, Resources: args.resources}
}

// This is an application admission policy, independent of GOMAXPROCS. Native
// guest entry and LLGo's scheduler do not acquire a Go scheduler CPU permit.
func serviceOptions(args options) error {
	if args.mode != "service" {
		if args.workers != 0 || args.clients != 0 || args.requests != 0 || args.resources {
			return errors.New("service counts require mode service")
		}
		return nil
	}
	return serviceConfig(args).Validate()
}

func serviceKernel(ctx context.Context, input []byte, args options, out *report) (kernel, func(), error) {
	if err := serviceOptions(args); err != nil {
		return nil, func() {}, err
	}
	calls, release, err := publicKernels(ctx, input, args, out, args.workers)
	if err != nil {
		return nil, release, err
	}
	start := time.Now()
	call, stop, result, err := service.Start(calls, serviceConfig(args))
	if err != nil {
		return nil, release, err
	}
	out.Service = result
	out.Service.WorkerSetupNs = time.Since(start).Nanoseconds()
	return call, func() {
		start := time.Now()
		stop()
		release()
		out.Phases.Cleanup = time.Since(start).Nanoseconds()
	}, nil
}
