//go:build !p3reference && !p3host

package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cpunion/wasmvm/internal/bench/service"
	"github.com/cpunion/wasmvm/pkg/wasi/preview3"
)

func bufferedKernel(ctx context.Context, args options, call func(context.Context, *http.Request) (*preview3.BufferedResponse, error)) service.Call {
	var snapshot string
	if args.requestMode == "snapshot" {
		snapshot = string(args.payload)
	}
	return func() (int64, uint32, error) {
		var request *http.Request
		var err error
		if args.requestMode == "snapshot" {
			request, err = preview3.NewSnapshotRequest(ctx, "POST", "http://localhost/echo", snapshot)
		} else {
			request, err = http.NewRequestWithContext(ctx, "POST", "http://localhost/echo", bytes.NewReader(args.payload))
		}
		if err != nil {
			return 0, 0, err
		}
		start := time.Now()
		response, err := call(ctx, request)
		ns := time.Since(start).Nanoseconds()
		if err != nil {
			return ns, 0, err
		}
		if response == nil {
			return ns, 0, errors.New("missing complete HTTP response")
		}
		if err := checkReply(args, response.Body, response.StatusCode, response.Header.Get("X-Guest"), response.Trailer.Get("X-Guest-Trailer")); err != nil {
			return ns, 0, err
		}
		return ns, args.Expected, nil
	}
}
