//go:build !p3reference && !p3host

package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasi/preview3"
)

func TestOwnedBufferedCompleteBothFixturesAndBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, fixture := range []string{"world-rust-http", "world-rust-composed-http"} {
			for _, mode := range []string{"buffered", "caller-buffer"} {
				args := largeOptions()
				args.engine, args.backend, args.path, args.responseMode = "wasmvm", backend, p3Fixture(fixture), mode
				out, err := benchmark(nil, args, nil)
				if err != nil {
					t.Fatal(backend, fixture, err)
				}
				checkBatches(t, out, args)
				if out.OwnedResponseMode != mode {
					t.Fatal(out.OwnedResponseMode)
				}
			}
		}
	}
}

func TestOwnedSnapshotRequestsAndModeGuards(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		args := nativeOptions()
		args.engine, args.backend, args.path = "wasmvm", backend, p3Fixture("world-rust-http")
		args.requestMode, args.responseMode = "snapshot", "buffered"
		out, err := benchmark(nil, args, nil)
		if err != nil {
			t.Fatal(backend, err)
		}
		checkBatches(t, out, args)
		if out.OwnedRequestMode != "snapshot" {
			t.Fatal(out.OwnedRequestMode)
		}
		for _, mode := range []string{"unknown", "snapshot"} {
			args.requestMode, args.responseMode = mode, "streaming"
			if err := validateResponseMode(args); err == nil {
				t.Fatal("invalid input contract accepted", mode)
			}
		}
		args.engine, args.responseMode = "native-go", "buffered"
		if err := validateResponseMode(args); err == nil {
			t.Fatal("native control accepted owned request selector")
		}
	}
}

func TestBufferedKernelChecksCompleteResponseAndContext(t *testing.T) {
	args := nativeOptions()
	for _, mode := range []string{"ok", "nil", "error", "body", "status", "header", "trailer"} {
		kernel := bufferedKernel(context.Background(), args, func(context.Context, *http.Request) (*preview3.BufferedResponse, error) {
			response := &preview3.BufferedResponse{StatusCode: 201, Body: args.payload, Header: http.Header{"X-Guest": []string{"rust-p3"}}, Trailer: http.Header{"X-Guest-Trailer": []string{"done"}}}
			switch mode {
			case "nil":
				return nil, nil
			case "error":
				return nil, io.ErrClosedPipe
			case "body":
				response.Body = nil
			case "status":
				response.StatusCode = 200
			case "header":
				response.Header = nil
			case "trailer":
				response.Trailer = nil
			}
			return response, nil
		})
		ns, result, err := kernel()
		if ns < 0 || (mode == "ok") != (err == nil) || err == nil && result != args.Expected {
			t.Fatal(mode, ns, result, err)
		}
		if mode == "error" && !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal("provider error was replaced", err)
		}
	}
	if _, _, err := bufferedKernel(nil, args, func(context.Context, *http.Request) (*preview3.BufferedResponse, error) {
		t.Fatal("invalid request context reached the provider")
		return nil, nil
	})(); err == nil {
		t.Fatal("nil context accepted")
	}
	args.responseMode = "buffered"
	if _, err := benchmark(nil, args, nil); err == nil {
		t.Fatal("native application accepted the owned API selector")
	}
	args.responseMode = "unknown"
	if _, err := benchmark(nil, args, nil); err == nil {
		t.Fatal("unknown response mode accepted")
	}
}
