package preview3

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/cpunion/wasmvm/pkg/component"
)

func TestHTTPReadBufferRangesEOFAndClose(t *testing.T) {
	b := &httpReadBuffer{}
	if r, ready := b.take(0); !ready || r.Closed || r.Bytes != "" {
		t.Fatal(r, ready)
	}
	if _, ready := b.take(1); ready {
		t.Fatal("empty buffer reported ready")
	}
	first := b.deliver("\x00\xff\x80abc", true, 2)
	if first.Bytes != "\x00\xff" || first.Closed {
		t.Fatal(first)
	}
	if r, ready := b.take(0); !ready || r.Closed || r.Bytes != "" {
		t.Fatal(r, ready)
	}
	last, ready := b.take(99)
	if !ready || last.Bytes != "\x80abc" || !last.Closed || first.Bytes != "\x00\xff" {
		t.Fatal(last, ready, first)
	}
	b.close()
	if r := b.deliver("late data", false, 99); r.Bytes != "" || !r.Closed {
		t.Fatal("closed source retained a late completion", r)
	}
	if r, ready := b.take(1); !ready || !r.Closed || b.data != "" {
		t.Fatal(r, ready)
	}
}

type boundedHTTPReader struct {
	mu      sync.Mutex
	reader  *bytes.Reader
	sizes   []int
	dataEOF bool
}

func (r *boundedHTTPReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sizes = append(r.sizes, len(p))
	n, err := r.reader.Read(p)
	if r.dataEOF && r.reader.Len() == 0 {
		err = io.EOF
	}
	return n, err
}
func (*boundedHTTPReader) Close() error { return nil }

func TestHTTPReadAheadBoundsExternalReadsAndPreservesTail(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, dataEOF := range []bool{false, true} {
			w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxReadBytes: 4096}})
			module := compileFixture(t, backend, "world-rust-http")
			in := instantiate(t, module, w.Imports())
			payload := bytes.Repeat([]byte("\x00\xff\x80abcde"), 2049)
			reader := &boundedHTTPReader{reader: bytes.NewReader(payload), dataEOF: dataEOF}
			request := httptest.NewRequest("POST", "http://localhost/", bytes.NewReader(payload))
			request.Body = reader
			response, err := w.HTTP().CallIncoming(context.Background(), in, request, component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			if err != nil || closeErr != nil || !bytes.Equal(got, payload) || response.Trailer.Get("X-Guest-Trailer") != "done" {
				t.Fatal(err, closeErr, len(got))
			}
			reader.mu.Lock()
			sizes := append([]int(nil), reader.sizes...)
			reader.mu.Unlock()
			want := (len(payload) + 4095) / 4096
			if !dataEOF {
				want++
			}
			if len(sizes) != want {
				t.Fatal("guest small reads reached the external reader", sizes, want)
			}
			for _, size := range sizes {
				if size != 4096 {
					t.Fatal("external read exceeded or bypassed its bounded batch", sizes)
				}
			}
		}
	})
}
