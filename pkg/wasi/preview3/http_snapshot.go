//go:build linux || darwin

package preview3

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
)

// NewSnapshotRequest constructs a request whose immutable body is already
// available. The provider can consume bounded string slices without external
// I/O workers or mutable-buffer snapshots. Body ownership, limits, trailers and
// cancellation follow CallIncoming's ordinary contract. Strings may contain
// arbitrary bytes; callers can convert mutable input once before repeated calls.
func NewSnapshotRequest(ctx context.Context, method, url, body string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Body = &snapshotReader{data: body}
	r.GetBody = func() (io.ReadCloser, error) { return &snapshotReader{data: body}, nil }
	return r, nil
}

type snapshotReader struct {
	mu     sync.Mutex
	data   string
	offset int
	closed bool
}

func (r *snapshotReader) take(limit int) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.takeLocked(limit)
}

func (r *snapshotReader) takeLocked(limit int) (string, error) {
	if r.closed {
		return "", io.ErrClosedPipe
	}
	if r.offset == len(r.data) {
		return "", io.EOF
	}
	n := min(limit, len(r.data)-r.offset)
	data := r.data[r.offset : r.offset+n]
	r.offset += n
	return data, nil
}

func (r *snapshotReader) Read(dst []byte) (int, error) {
	data, err := r.take(len(dst))
	return copy(dst, data), err
}

func (r *snapshotReader) Close() error {
	r.mu.Lock()
	r.closed = true
	r.data = ""
	r.offset = 0
	r.mu.Unlock()
	return nil
}

// Only the provider-owned immutable reader takes this path. Arbitrary readers
// retain scheduled I/O; contention with another read also uses that path.
func (b *httpBody) readSnapshot(reader io.ReadCloser) (data string, err error, ready bool) {
	r, ok := reader.(*snapshotReader)
	if !ok || r == nil || !b.readerMu.TryLock() {
		return "", nil, false
	}
	defer b.readerMu.Unlock()
	if !r.mu.TryLock() {
		return "", nil, false
	}
	data, err = r.takeLocked(int(b.owner.network.io.config.MaxReadBytes))
	r.mu.Unlock()
	ready = true
	defer func() {
		if recover() != nil {
			data, err = "", httpErr("internal-error")
			b.end(err)
			b.closeReader()
		}
	}()
	n, err := b.finishExternalRead(len(data), int(b.owner.network.io.config.MaxReadBytes), err)
	return data[:n], err, true
}
