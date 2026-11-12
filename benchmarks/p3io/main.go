// Command p3io measures complete embedded P3 HTTP body/trailer lifetimes using
// the same bounded application driver as the independent Core references.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

type options struct {
	engine, backend, path string
	responseMode          string
	requestMode           string
	payload               []byte
	service.Options
}

type report struct {
	SchemaVersion     int              `json:"schemaVersion"`
	Engine            string           `json:"engine"`
	Backend           string           `json:"backend"`
	HostCompiler      string           `json:"hostCompiler"`
	Mode              string           `json:"mode"`
	Policy            string           `json:"policy"`
	PayloadBytes      int              `json:"payloadBytes"`
	OwnedResponseMode string           `json:"ownedResponseMode,omitempty"`
	OwnedRequestMode  string           `json:"ownedRequestMode,omitempty"`
	Result            uint32           `json:"result"`
	Phases            map[string]int64 `json:"phasesNs"`
	WarmCallNs        []int64          `json:"warmCallNs"`
	WarmupCallNs      []int64          `json:"warmupCallNs"`
	WarmResults       []uint32         `json:"warmResults"`
	WarmupResults     []uint32         `json:"warmupResults"`
	Service           *service.Report  `json:"service"`
}

func parse(args []string, diagnostic io.Writer) (out options, err error) {
	f := flag.NewFlagSet("p3io", flag.ContinueOnError)
	f.SetOutput(diagnostic)
	f.StringVar(&out.engine, "engine", "wasmvm", "wasmvm, native-go or wasmtime")
	f.StringVar(&out.backend, "backend", "native", "interpreter/native, host, or compiler")
	f.StringVar(&out.path, "wasm", "", "pinned P3 HTTP echo component")
	f.StringVar(&out.responseMode, "response-mode", "streaming", "owned API: streaming, buffered or caller-buffer collection")
	f.StringVar(&out.requestMode, "request-mode", "reader", "owned input: reader or immutable snapshot")
	size := f.Int("payload-bytes", 65536, "positive byte count, multiple of16, <=1MiB")
	expected := f.Uint64("expected", 0, "required unsigned CRC32 of the complete expected body")
	f.IntVar(&out.Workers, "workers", 1, "independent stores/workers")
	f.IntVar(&out.Clients, "clients", 64, "one outstanding request per client")
	f.IntVar(&out.Requests, "requests", 16, "requests per client and batch")
	f.IntVar(&out.Samples, "samples", 7, "measured batches")
	f.IntVar(&out.Warmups, "warmups", 2, "untallied warmup batches")
	f.BoolVar(&out.Resources, "service-resources", false, "separate resource cohort")
	if err = f.Parse(args); err != nil {
		return out, err
	}
	hasExpected := false
	f.Visit(func(v *flag.Flag) { hasExpected = hasExpected || v.Name == "expected" })
	if !hasExpected || f.NArg() != 0 || *expected > math.MaxUint32 || *size <= 0 || *size > 1<<20 || *size%16 != 0 {
		return out, errors.New("valid payload size and expected CRC32 are required; no positional arguments")
	}
	switch out.engine {
	case "wasmvm":
		if out.path == "" || (out.backend != "native" && out.backend != "interpreter") {
			return out, errors.New("wasmvm requires a component and backend native/interpreter")
		}
	case "native-go":
		if out.path != "" || out.backend != "host" {
			return out, errors.New("native-go requires backend host and no component")
		}
	case "wasmtime":
		if out.path == "" || out.backend != "compiler" {
			return out, errors.New("wasmtime requires a component and backend compiler")
		}
	default:
		return out, errors.New("unknown engine")
	}
	if err = validateResponseMode(out); err != nil {
		return out, err
	}
	out.Expected = uint32(*expected)
	if err = out.Options.Validate(); err != nil {
		return out, err
	}
	out.payload = bytes.Repeat([]byte("0123456789abcdef"), *size/16)
	if crc32.ChecksumIEEE(out.payload) != out.Expected {
		return out, errors.New("expected CRC32 does not match the independently declared payload")
	}
	return out, nil
}

func validateResponseMode(args options) error {
	if args.requestMode != "" && args.requestMode != "reader" && args.requestMode != "snapshot" {
		return errors.New("unknown owned request mode")
	}
	if args.requestMode == "snapshot" && args.engine != "wasmvm" {
		return errors.New("snapshot request mode applies to the owned VM API")
	}
	if args.requestMode == "snapshot" && args.responseMode != "buffered" && args.responseMode != "caller-buffer" {
		return errors.New("snapshot benchmark requests require complete buffered collection")
	}
	if args.responseMode != "" && args.responseMode != "streaming" && args.responseMode != "buffered" && args.responseMode != "caller-buffer" {
		return errors.New("unknown owned response mode")
	}
	if (args.responseMode == "buffered" || args.responseMode == "caller-buffer") && args.engine != "wasmvm" {
		return errors.New("response-mode applies to the owned VM API")
	}
	return nil
}

func checkReply(args options, data []byte, status int, guest, trailer string) error {
	if status != 201 || guest != "rust-p3" || trailer != "done" || !bytes.Equal(data, args.payload) {
		return errors.New("complete HTTP body, status, header or trailer mismatch")
	}
	return nil
}

type httpCall func(context.Context, *http.Request) (*http.Response, error)

func httpKernel(ctx context.Context, args options, call httpCall) service.Call {
	return func() (int64, uint32, error) {
		request, err := http.NewRequestWithContext(ctx, "POST", "http://localhost/echo", bytes.NewReader(args.payload))
		if err != nil {
			return 0, 0, err
		}
		start := time.Now()
		response, err := call(ctx, request)
		var data []byte
		if response != nil && response.Body != nil {
			if err != nil {
				err = errors.Join(err, response.Body.Close())
			} else {
				data, err = readResponse(response, len(args.payload))
			}
		}
		ns := time.Since(start).Nanoseconds()
		if err != nil {
			return ns, 0, err
		}
		if response == nil || response.Body == nil {
			return ns, 0, errors.New("missing HTTP response body")
		}
		if err = checkReply(args, data, response.StatusCode, response.Header.Get("X-Guest"), response.Trailer.Get("X-Guest-Trailer")); err != nil {
			return ns, 0, err
		}
		// Every byte was checked; the reply carries the previously verified CRC.
		return ns, args.Expected, nil
	}
}

func readResponse(response *http.Response, size int) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(response.Body, int64(size)+1))
	return data, errors.Join(err, response.Body.Close())
}

func nativeKernels(ctx context.Context, args options) []service.Call {
	calls := make([]service.Call, args.Workers)
	for i := range calls {
		calls[i] = httpKernel(ctx, args, func(ctx context.Context, r *http.Request) (*http.Response, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			data, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			return &http.Response{StatusCode: 201, Header: http.Header{"X-Guest": []string{"rust-p3"}}, Trailer: http.Header{"X-Guest-Trailer": []string{"done"}}, Body: io.NopCloser(bytes.NewReader(data))}, nil
		})
	}
	return calls
}

func benchmark(ctx context.Context, args options, beforeSample func(int) error) (out report, err error) {
	if err = validateResponseMode(args); err != nil {
		return out, err
	}
	if err = args.Options.Validate(); err != nil {
		return out, err
	}
	if len(args.payload) == 0 || len(args.payload) > 1<<20 || crc32.ChecksumIEEE(args.payload) != args.Expected {
		return out, errors.New("invalid payload/CRC32 contract")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	out = report{SchemaVersion: 1, Engine: args.engine, Backend: args.backend, HostCompiler: runtime.Compiler, Mode: "p3-http-service", PayloadBytes: len(args.payload), Phases: map[string]int64{}, WarmCallNs: []int64{}, WarmupCallNs: []int64{}, WarmResults: []uint32{}, WarmupResults: []uint32{}}
	if args.engine == "wasmvm" {
		out.OwnedRequestMode = args.requestMode
		if out.OwnedRequestMode == "" {
			out.OwnedRequestMode = "reader"
		}
		out.OwnedResponseMode = args.responseMode
		if out.OwnedResponseMode == "" {
			out.OwnedResponseMode = "streaming"
		}
	}
	var calls []service.Call
	var release func() error
	var stop func()
	if args.engine == "native-go" {
		out.Policy = "ordinary in-process Go HTTP echo/body collection; same service driver; no Wasm, guest, JIT, component or P3 protocol"
		calls = nativeKernels(ctx, args)
		release = func() error { return nil }
	} else if args.engine == "wasmvm" {
		calls, release, err = ownedKernels(ctx, args, &out)
	} else {
		calls, release, err = referenceKernels(args, &out)
	}
	if release != nil {
		defer func() {
			start := time.Now()
			if stop != nil {
				stop()
			}
			err = errors.Join(err, release())
			out.Phases["cleanup"] = time.Since(start).Nanoseconds()
		}()
	}
	if err != nil {
		return out, err
	}
	start := time.Now()
	call, joinedStop, data, err := service.Start(calls, args.Options)
	if err != nil {
		return out, err
	}
	stop, out.Service = joinedStop, data
	data.WorkerSetupNs = time.Since(start).Nanoseconds()
	out.Phases["firstCall"], out.Result, err = call()
	if err != nil {
		return out, err
	}
	for range args.Warmups {
		ns, result, e := call()
		if e != nil {
			return out, e
		}
		out.WarmupCallNs, out.WarmupResults = append(out.WarmupCallNs, ns), append(out.WarmupResults, result)
	}
	for i := range args.Samples {
		if beforeSample != nil {
			if err = beforeSample(i); err != nil {
				return out, err
			}
		}
		ns, result, e := call()
		if e != nil {
			return out, e
		}
		out.WarmCallNs, out.WarmResults = append(out.WarmCallNs, ns), append(out.WarmResults, result)
	}
	return out, nil
}

func run(args []string, input io.Reader, output, diagnostic io.Writer) int {
	encoder := json.NewEncoder(output)
	err := encoder.Encode(struct {
		Ready bool `json:"ready"`
	}{true})
	reader := bufio.NewReader(input)
	ack := func(expected string) error {
		line, err := reader.ReadString('\n')
		if err != nil || line != expected+"\n" {
			return errors.New("invalid acknowledgement")
		}
		return nil
	}
	if err == nil {
		err = ack("run")
	}
	var config options
	if err == nil {
		config, err = parse(args, diagnostic)
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err == nil {
		var wait func(int) error
		if os.Getenv("WASMVM_INTERLEAVE") == "1" {
			wait = func(i int) error {
				if err := encoder.Encode(struct {
					SampleReady int `json:"sampleReady"`
				}{i}); err != nil {
					return err
				}
				return ack("sample")
			}
		}
		var out report
		out, err = benchmark(context.Background(), config, wait)
		if err == nil {
			err = encoder.Encode(out)
		}
	}
	if err != nil {
		fmt.Fprintln(diagnostic, err)
		return 1
	}
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
