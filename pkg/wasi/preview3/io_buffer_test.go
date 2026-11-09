package preview3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
)

func TestIOTransferBufferReuseOwnershipAndClose(t *testing.T) {
	p, err := NewIO(IOConfig{MaxIdleWorkers: 1})
	if err != nil {
		t.Fatal(err)
	}
	first := p.acquireTransferBuffer(16)
	copy(first, "private snapshot")
	snapshot := string(first)
	second := p.acquireTransferBuffer(16)
	if &first[0] == &second[0] {
		t.Fatal("two accepted operations alias one mutable buffer")
	}
	p.releaseTransferBuffer(first)
	p.releaseTransferBuffer(second)
	if len(p.idleBuffers) != 1 {
		t.Fatal("idle retention exceeded its bound")
	}
	reused := p.acquireTransferBuffer(8)
	if &reused[0] != &first[0] || len(reused) != 8 {
		t.Fatal("bounded buffer was not reused")
	}
	copy(reused, "overwrit!")
	if snapshot != "private snapshot" {
		t.Fatal("immutable result changed after buffer reuse")
	}
	p.releaseTransferBuffer(reused)
	large := p.acquireTransferBuffer(32)
	if len(large) != 32 {
		t.Fatal(len(large))
	}
	p.releaseTransferBuffer(large)
	p.releaseTransferBuffer(make([]byte, int(p.config.MaxReadBytes)+1))
	if len(p.idleBuffers) != 1 || cap(p.idleBuffers[0]) > int(p.config.MaxReadBytes) {
		t.Fatal("oversized buffer retained")
	}
	if buffer := p.acquireTransferBuffer(0); buffer != nil {
		t.Fatal("empty operation allocated a buffer")
	}
	p.releaseTransferBuffer(nil)
	active := p.acquireTransferBuffer(16)
	p.Close()
	p.releaseTransferBuffer(active)
	closeIO(t, p)
	if len(p.idleBuffers) != 0 {
		t.Fatal("close retained buffers or accepted a late return")
	}
}

type transferFailureWriter struct{ err error }

func (w transferFailureWriter) Write([]byte) (int, error) { return 0, w.err }

func TestIOTransferWriteSnapshotsAndErrors(t *testing.T) {
	p, _ := NewIO(IOConfig{})
	defer closeIO(t, p)
	data := strings.Repeat("\x00\xff\x80abcde", 64)
	reader, writer := io.Pipe()
	done := make(chan error, 1)
	go func() {
		n, err := p.writeTransferString(writer, data)
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
		writer.CloseWithError(err)
		done <- err
	}()
	var output bytes.Buffer
	buffer := make([]byte, 3)
	for {
		n, err := reader.Read(buffer)
		output.Write(buffer[:n])
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := <-done; err != nil || output.String() != data {
		t.Fatal(err, output.Len())
	}
	failure := errors.New("writer failed")
	if n, err := p.writeTransferString(transferFailureWriter{failure}, data); n != 0 || !errors.Is(err, failure) {
		t.Fatal(n, err)
	}
	var empty bytes.Buffer
	if n, err := p.writeTransferString(&empty, ""); n != 0 || err != nil {
		t.Fatal(n, err)
	}
}

func TestHTTPTransferBuffersKeepConcurrentStoreBodiesIndependent(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxIdleWorkers: 1, MaxReadBytes: 256, MaxWriteBytes: 256}})
		module := compileFixture(t, backend, "world-rust-http")
		instances := []*component.Instance{instantiate(t, module, w.Imports()), instantiate(t, module, w.Imports())}
		var workers sync.WaitGroup
		for worker, instance := range instances {
			workers.Go(func() {
				for round := range 10 {
					payload := strings.Repeat(string([]byte{byte(128 + worker*17 + round)}), 4096+worker*257+round)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					response, err := w.HTTP().CallIncoming(ctx, instance, httptest.NewRequest("POST", "http://localhost/", strings.NewReader(payload)), component.CallOptions{})
					if err != nil {
						cancel()
						t.Error(err)
						return
					}
					data, readErr := io.ReadAll(response.Body)
					closeErr := response.Body.Close()
					cancel()
					if readErr != nil || closeErr != nil || string(data) != payload || response.StatusCode != 201 || response.Trailer.Get("X-Guest-Trailer") != "done" {
						t.Error("transfer buffer crossed request/Store ownership", worker, round, len(data), readErr, closeErr)
						return
					}
				}
			})
		}
		workers.Wait()
	})
}
