//go:build linux || darwin

package preview3

import (
	"context"
	"net/http"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// BufferedResponse contains the complete response, including trailers. Body is
// independent of the guest. CallIncomingBuffered allocates its owned body;
// CallIncomingInto uses the caller-provided storage described by that method.
type BufferedResponse struct {
	StatusCode    int
	Header        http.Header
	Trailer       http.Header
	Body          []byte
	FuelRemaining uint64
}

type httpCollector struct {
	reply       httpReply
	limit       uint64
	data        []byte
	fuel        uint64
	failure     error
	body        *httpBody
	statusCode  int
	header      http.Header
	trailer     http.Header
	expected    *uint64
	sizeFailure string
	reserved    uint32
}

// Retain the collected response through guest task cleanup. After that join,
// buffered calls need no streaming owner for the completed body.
func (c *httpCollector) releaseBody() {
	if c.body != nil {
		c.body.release()
		c.body = nil
	}
}

// CallIncomingBuffered consumes the complete response inside the handler's
// Store lease. It drives guest streams directly through the component scheduler.
// maxBodyBytes is a positive materialization limit; oversized responses fail
// with ErrLimit. The call joins background guest work before returning.
func (p *HTTP) CallIncomingBuffered(ctx context.Context, in *component.Instance, request *http.Request, maxBodyBytes uint64, options component.CallOptions) (*BufferedResponse, error) {
	if maxBodyBytes == 0 || maxBodyBytes > uint64(^uint(0)>>1) {
		return nil, wasm.ErrLimit
	}
	collector := httpCollector{limit: maxBodyBytes}
	return p.collectIncoming(ctx, in, request, options, &collector)
}

// CallIncomingInto consumes and joins the complete response using dst as its
// body storage. len(dst) is the positive materialization limit; an oversized
// response fails with ErrLimit. The returned Body aliases dst, so a later reuse
// changes that view. dst must remain exclusive to this call until it returns.
// On failure its already written prefix may have changed. Header/trailer,
// ownership, cancellation, fuel, copy budgets and guest join match the ordinary
// buffered API; no destination is retained by the VM after return.
func (p *HTTP) CallIncomingInto(ctx context.Context, in *component.Instance, request *http.Request, dst []byte, options component.CallOptions) (*BufferedResponse, error) {
	if len(dst) == 0 {
		return nil, wasm.ErrLimit
	}
	collector := httpCollector{limit: uint64(len(dst)), data: dst[:0:len(dst)]}
	return p.collectIncoming(ctx, in, request, options, &collector)
}

func (p *HTTP) collectIncoming(ctx context.Context, in *component.Instance, request *http.Request, options component.CallOptions, collector *httpCollector) (*BufferedResponse, error) {
	if _, err := p.callIncoming(ctx, in, request, options, collector); err != nil {
		return nil, err
	}
	return &BufferedResponse{StatusCode: collector.statusCode, Header: collector.header, Trailer: collector.trailer, Body: collector.data, FuelRemaining: collector.fuel}, nil
}

func (collector *httpCollector) destination(count uint32) ([]byte, error) {
	remaining := collector.limit - uint64(len(collector.data))
	if uint64(count) > remaining {
		return nil, wasm.ErrLimit
	}
	size := uint64(len(collector.data)) + uint64(count)
	if collector.expected != nil && size > *collector.expected {
		return nil, httpSizeError(collector.sizeFailure, size)
	}
	if size > uint64(cap(collector.data)) {
		capacity := min(max(size, 2*uint64(cap(collector.data))), collector.limit)
		grown := make([]byte, len(collector.data), int(capacity))
		// A single read may accept several guest writes before its event
		// is observed. Preserve the current batch as well as prior reads.
		copy(grown[:len(collector.data)+int(collector.reserved)], collector.data[:len(collector.data)+int(collector.reserved)])
		collector.data = grown
	}
	collector.reserved = count
	return collector.data[len(collector.data):int(size)], nil
}

func (collector *httpCollector) collectBody(c *component.Caller, b *httpBody, expected *uint64) (err error) {
	collector.data = collector.data[:0:cap(collector.data)]
	collector.expected = expected
	collector.sizeFailure = b.sizeFailure
	if expected != nil && *expected > collector.limit {
		return wasm.ErrLimit
	}
	if err := b.watchTrailers(c); err != nil {
		return err
	}
	b.mu.Lock()
	original, anchor, external := b.original, b.anchor, b.reader != nil
	b.original, b.anchor = nil, nil
	b.mu.Unlock()
	var stream *component.Resource
	defer func() {
		if err != nil {
			b.dispose(c, err)
		}
		if stream != nil {
			_ = c.DropAsync(stream)
		}
		if anchor != nil {
			_ = c.DropAsync(anchor)
		}
	}()
	if original != nil {
		stream = original
	} else if external {
		value, failed := b.byteSource(c, component.ByteStreamType())
		if failed != nil {
			return failed
		}
		stream = value.Resource
	}
	dest := collector.destination
	for stream != nil {
		// The extra byte detects overflow without materializing an invalid
		// suffix. Empty/EOF batches never invoke the destination callback.
		remaining := collector.limit - uint64(len(collector.data))
		next := min(remaining+1, uint64(b.owner.network.io.config.MaxWriteBytes), uint64(b.owner.config.BodyBufferBytes))
		collector.reserved = 0
		r, failed := c.ReadAsyncInto(b.ctx, stream, uint32(next), dest)
		if failed != nil {
			return failed
		}
		size := uint64(len(collector.data)) + uint64(r.Count)
		if expected != nil && r.Closed && size != *expected {
			return httpSizeError(b.sizeFailure, size)
		}
		collector.data = collector.data[:int(size)]
		if r.Closed {
			break
		}
	}

	if expected != nil && uint64(len(collector.data)) != *expected {
		return httpSizeError(b.sizeFailure, uint64(len(collector.data)))
	}
	b.end(nil)
	if err := b.waitTrailers(c); err != nil {
		return err
	}
	b.mu.Lock()
	fields, failure := b.trailersValue, b.failure
	b.mu.Unlock()
	if failure == nil && b.ctx.Err() != nil {
		failure = errConnectionTerminated
	}
	if failure != nil {
		return failure
	}
	if fields != nil {
		fields.ForEach(func(field host.HTTPField) {
			collector.trailer.Add(field.Name, field.Value)
		})
	}
	b.transferred()
	return nil
}
