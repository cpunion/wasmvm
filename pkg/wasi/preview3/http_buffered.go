//go:build linux || darwin

package preview3

import (
	"context"
	"net/http"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// BufferedResponse owns the complete response, including trailers. Body remains
// independent of the guest and may be retained or changed after the call.
type BufferedResponse struct {
	StatusCode    int
	Header        http.Header
	Trailer       http.Header
	Body          []byte
	FuelRemaining uint64
}

type httpCollector struct {
	limit   uint64
	data    []byte
	fuel    uint64
	failure error
}

// CallIncomingBuffered consumes the complete response inside the handler's
// Store lease. It drives guest streams directly through the component scheduler.
// maxBodyBytes is a positive materialization limit; oversized responses fail
// with ErrLimit. The call joins background guest work before returning.
func (p *HTTP) CallIncomingBuffered(ctx context.Context, in *component.Instance, request *http.Request, maxBodyBytes uint64, options component.CallOptions) (*BufferedResponse, error) {
	if maxBodyBytes == 0 || maxBodyBytes > uint64(^uint(0)>>1) {
		return nil, wasm.ErrLimit
	}
	collector := &httpCollector{limit: maxBodyBytes}
	response, err := p.callIncoming(ctx, in, request, options, collector)
	if err != nil {
		return nil, err
	}
	if err := response.Body.Close(); err != nil {
		return nil, err
	}
	return &BufferedResponse{StatusCode: response.StatusCode, Header: response.Header, Trailer: response.Trailer, Body: collector.data, FuelRemaining: collector.fuel}, nil
}

func (b *httpBody) collect(c *component.Caller, trailer http.Header, expected *uint64, limit uint64) (data []byte, err error) {
	if expected != nil && *expected > limit {
		return nil, wasm.ErrLimit
	}
	if err := b.watchTrailers(c); err != nil {
		return nil, err
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
			return nil, failed
		}
		stream = value.Resource
	}
	for stream != nil {
		// One extra byte detects an oversized body even when the accepted
		// prefix exactly reaches the limit. The extra byte is never appended
		// to the materialized response.
		remaining := limit - uint64(len(data))
		next := min(remaining+1, uint64(b.owner.network.io.config.MaxWriteBytes), uint64(b.owner.config.BodyBufferBytes))
		r, failed := c.ReadAsyncContext(b.ctx, stream, uint32(next))
		if failed != nil {
			return nil, failed
		}
		if uint64(len(r.Bytes)) > remaining {
			return nil, wasm.ErrLimit
		}
		size := uint64(len(data)) + uint64(len(r.Bytes))
		if expected != nil && (size > *expected || r.Closed && size != *expected) {
			return nil, httpSizeError(b.sizeFailure, size)
		}
		data = append(data, r.Bytes...)
		if r.Closed {
			break
		}
	}
	if expected != nil && uint64(len(data)) != *expected {
		return nil, httpSizeError(b.sizeFailure, uint64(len(data)))
	}
	b.end(nil)
	if err := b.waitTrailers(c); err != nil {
		return nil, err
	}
	b.mu.Lock()
	fields, failure := b.trailersValue, b.failure
	b.mu.Unlock()
	if failure == nil && b.ctx.Err() != nil {
		failure = httpErr("connection-terminated")
	}
	if failure != nil {
		return nil, failure
	}
	if fields != nil {
		for _, field := range fields.Entries() {
			trailer.Add(field.Name, field.Value)
		}
	}
	b.transferred()
	return data, nil
}
