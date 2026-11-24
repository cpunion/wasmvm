package preview3

import (
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
)

// One external read supplies a bounded immutable batch. Smaller guest reads
// consume it without another worker handoff. A source close clears the batch;
// a late completion cannot publish into a closed source.
type httpReadBuffer struct {
	mu      sync.Mutex
	data    string
	closed  bool
	stopped bool
}

func (b *httpReadBuffer) take(limit uint32) (component.ReadResult, bool) {
	b.mu.Lock()
	if b.data == "" && !b.closed && !b.stopped && limit != 0 {
		b.mu.Unlock()
		return component.ReadResult{}, false
	}
	res := b.takeLocked(limit)
	b.mu.Unlock()
	return res, true
}

func (b *httpReadBuffer) takeLocked(limit uint32) component.ReadResult {
	n := min(uint64(limit), uint64(len(b.data)))
	data := b.data[:n]
	b.data = b.data[n:]
	return component.ReadResult{Bytes: data, Closed: b.stopped || b.closed && len(b.data) == 0}
}

func (b *httpReadBuffer) deliver(data string, closed bool, limit uint32) component.ReadResult {
	b.mu.Lock()
	if !b.stopped {
		b.data, b.closed = data, closed
	}
	res := b.takeLocked(limit)
	b.mu.Unlock()
	return res
}

func (b *httpReadBuffer) close() {
	b.mu.Lock()
	b.data, b.stopped = "", true
	b.mu.Unlock()
}
