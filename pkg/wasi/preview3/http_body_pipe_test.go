//go:build linux || darwin

package preview3

import (
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type bodyPipeResult struct {
	n   int
	err error
}

func awaitBodyWrite(t *testing.T, done <-chan bodyPipeResult) bodyPipeResult {
	t.Helper()
	select {
	case result := <-done:
		return result
	case <-time.After(3 * time.Second):
		t.Fatal("body writer did not complete")
		return bodyPipeResult{}
	}
}

func TestHTTPBodyPipePartialReadsKeepWholeBatchBackpressure(t *testing.T) {
	p := newHTTPBodyPipe()
	data := strings.Repeat("\x00\xff\x80abcde", 8192)
	done := make(chan bodyPipeResult, 1)
	go func() { n, err := p.WriteString(data); done <- bodyPipeResult{n, err} }()
	first := make([]byte, 7)
	if n, err := p.Read(first); n != 7 || err != nil || string(first) != data[:7] {
		t.Fatal(n, err, first)
	}
	select {
	case result := <-done:
		t.Fatal("writer completed before its batch was consumed", result)
	default:
	}
	got := append([]byte(nil), first...)
	buffer := make([]byte, 127)
	for len(got) < len(data) {
		n, err := p.Read(buffer)
		if err != nil || n <= 0 {
			t.Fatal(n, err)
		}
		got = append(got, buffer[:n]...)
	}
	if result := awaitBodyWrite(t, done); result.n != len(data) || result.err != nil || string(got) != data {
		t.Fatal(result, len(got))
	}
	p.CloseWrite(nil)
	if n, err := p.Read(buffer); n != 0 || err != io.EOF {
		t.Fatal(n, err)
	}
}

func TestHTTPBodyPipeZeroTransfersAndPartialClose(t *testing.T) {
	p := newHTTPBodyPipe()
	done := make(chan bodyPipeResult, 1)
	go func() { n, err := p.WriteString(""); done <- bodyPipeResult{n, err} }()
	if n, err := p.Read(nil); n != 0 || err != nil {
		t.Fatal(n, err)
	}
	if result := awaitBodyWrite(t, done); result.n != 0 || result.err != nil {
		t.Fatal(result)
	}
	go func() { n, err := p.WriteString("abcde"); done <- bodyPipeResult{n, err} }()
	if n, err := p.Read(make([]byte, 2)); n != 2 || err != nil {
		t.Fatal(n, err)
	}
	failure := errors.New("reader stopped")
	p.CloseRead(failure)
	p.CloseRead(errors.New("later close"))
	if result := awaitBodyWrite(t, done); result.n != 2 || !errors.Is(result.err, failure) {
		t.Fatal(result)
	}
	if _, err := p.WriteString("later"); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if _, err := p.Read(nil); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	p.mu.Lock()
	retained := p.chunk != nil
	p.mu.Unlock()
	if retained {
		t.Fatal("closed pipe retained an unread batch")
	}
}

func TestHTTPBodyPipeCloseErrorsMatchIOPipe(t *testing.T) {
	first, later := errors.New("first close"), errors.New("later close")
	for _, mode := range []string{"reader", "writer", "both"} {
		p := newHTTPBodyPipe()
		reader, writer := io.Pipe()
		if mode != "writer" {
			p.CloseRead(first)
			p.CloseRead(later)
			reader.CloseWithError(first)
			reader.CloseWithError(later)
		}
		if mode != "reader" {
			p.CloseWrite(first)
			p.CloseWrite(later)
			writer.CloseWithError(first)
			writer.CloseWithError(later)
		}
		_, gotRead := p.Read(nil)
		_, wantRead := reader.Read(nil)
		_, gotWrite := p.WriteString("")
		_, wantWrite := writer.Write(nil)
		if !errors.Is(gotRead, wantRead) || !errors.Is(gotWrite, wantWrite) {
			t.Fatal(mode, gotRead, wantRead, gotWrite, wantWrite)
		}
	}
}

func TestHTTPBodyPipeConcurrentReadersAndSerializedWriters(t *testing.T) {
	p := newHTTPBodyPipe()
	var writers sync.WaitGroup
	for _, data := range []string{strings.Repeat("a", 1024), strings.Repeat("b", 1024)} {
		writers.Go(func() {
			if n, err := p.WriteString(data); n != len(data) || err != nil {
				t.Error(n, err)
			}
		})
	}
	go func() { writers.Wait(); p.CloseWrite(nil) }()
	var readers sync.WaitGroup
	var a, b atomic.Int32
	for range 4 {
		readers.Go(func() {
			buffer := make([]byte, 13)
			for {
				n, err := p.Read(buffer)
				for _, value := range buffer[:n] {
					switch value {
					case 'a':
						a.Add(1)
					case 'b':
						b.Add(1)
					default:
						t.Error("corrupt body byte", value)
					}
				}
				if err != nil {
					if err != io.EOF {
						t.Error(err)
					}
					return
				}
			}
		})
	}
	readers.Wait()
	if a.Load() != 1024 || b.Load() != 1024 {
		t.Fatal("lost or duplicated bytes", a.Load(), b.Load())
	}
}

func TestHTTPBodyPipeCloseWakesBlockedOperations(t *testing.T) {
	for _, writer := range []bool{false, true} {
		p := newHTTPBodyPipe()
		done := make(chan bodyPipeResult, 1)
		go func() {
			var n int
			var err error
			if writer {
				n, err = p.WriteString("pending")
			} else {
				n, err = p.Read(make([]byte, 1))
			}
			done <- bodyPipeResult{n, err}
		}()
		if writer {
			p.CloseRead(nil)
		} else {
			p.CloseWrite(nil)
		}
		result := awaitBodyWrite(t, done)
		want := io.EOF
		if writer {
			want = io.ErrClosedPipe
		}
		if result.n != 0 || result.err != want {
			t.Fatal(writer, result)
		}
	}
}
