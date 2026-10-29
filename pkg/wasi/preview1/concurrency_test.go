//go:build linux || darwin

package preview1

import (
	"context"
	"encoding/binary"
	"errors"
	"github.com/cpunion/wasmvm/pkg/vm"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type pendingInput struct {
	started              chan struct{}
	closed               chan struct{}
	startOnce, closeOnce sync.Once
}

func newPending() *pendingInput {
	return &pendingInput{started: make(chan struct{}), closed: make(chan struct{})}
}
func (s *pendingInput) Read(ctx context.Context, _ []byte) (int, error) {
	s.startOnce.Do(func() { close(s.started) })
	<-ctx.Done()
	return 0, ctx.Err()
}
func (s *pendingInput) Write(context.Context, []byte) (int, error) { return 0, ErrBadf }
func (s *pendingInput) Ready(ctx context.Context, read bool) (uint64, bool, bool, error) {
	return 0, false, false, ctx.Err()
}
func (s *pendingInput) Poll(ctx context.Context, _ bool) (uint64, bool, error) {
	s.startOnce.Do(func() { close(s.started) })
	<-ctx.Done()
	return 0, false, ctx.Err()
}
func (s *pendingInput) Close() error { s.closeOnce.Do(func() { close(s.closed) }); return nil }

func TestProcessAndStoreCancellation(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			for _, closeProcess := range []bool{false, true} {
				t.Run(map[bool]string{false: "Store", true: "Process"}[closeProcess], func(t *testing.T) {
					input := newPending()
					cfg := DefaultConfig()
					cfg.Stdin = input
					p := process(t, cfg)
					a, b := newFixture(t, p, backend), newFixture(t, p, backend)
					a.iovec(100, 200, 3)
					done := make(chan error, 1)
					go func() { _, err := a.raw(a.ctx, "fd_read", 0, 100, 1, 300); done <- err }()
					select {
					case <-input.started:
					case <-a.ctx.Done():
						t.Fatal(a.ctx.Err())
					}
					if closeProcess {
						p.Close()
						if err := p.WaitClosed(a.ctx); err != nil {
							t.Fatal(err)
						}
					} else {
						a.s.Close()
						if err := a.s.WaitClosed(a.ctx); err != nil {
							t.Fatal(err)
						}
						b.want(ErrSuccess, "clock_time_get", 1, 0, 0)
					}
					select {
					case err := <-done:
						if err == nil {
							t.Fatal("blocked call succeeded")
						}
					case <-b.ctx.Done():
						t.Fatal(b.ctx.Err())
					}
					if closeProcess {
						select {
						case <-input.closed:
						default:
							t.Fatal("stream retained")
						}
					} else {
						select {
						case <-input.closed:
							t.Fatal("Store closed shared process stream")
						default:
						}
					}
				})
			}
		})
	}
}

func TestPollCancellationQuotaAndFDReuse(t *testing.T) {
	input := newPending()
	cfg := DefaultConfig()
	cfg.Stdin = input
	cfg.Limits.MaxSubscriptions = 1
	p := process(t, cfg)
	a, b := newFixture(t, p, "native"), newFixture(t, p, "interpreter")
	var sub [48]byte
	sub[8] = 1
	for _, f := range []*fixture{a, b} {
		f.write(1024, sub[:])
	}
	done := make(chan error, 1)
	go func() { _, err := a.raw(a.ctx, "poll_oneoff", 1024, 2048, 1, 300); done <- err }()
	select {
	case <-input.started:
	case <-a.ctx.Done():
		t.Fatal(a.ctx.Err())
	}
	b.want(ErrNomem, "poll_oneoff", 1024, 2048, 1, 300)
	if err := p.closeFD(0); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-b.ctx.Done():
		t.Fatal(b.ctx.Err())
	}
	if a.u32(300) != 1 || binary.LittleEndian.Uint16(a.read(2056, 2)) != uint16(ErrBadf) {
		t.Fatal("closed descriptor event")
	}
	p.mu.Lock()
	workers := p.pollWorkers
	p.mu.Unlock()
	if workers != 0 {
		t.Fatal("poll worker leak", workers)
	}
	clock := systemClock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := clock.Wait(ctx, 1, ^uint64(0)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := clock.Wait(context.Background(), 99, 0); !errors.Is(err, ErrInval) {
		t.Fatal(err)
	}
	if _, err := clock.Now(99); !errors.Is(err, ErrInval) {
		t.Fatal(err)
	}
}

func TestSocketDatagramAndWaitAll(t *testing.T) {
	p := process(t, DefaultConfig())
	f := newFixture(t, p, "native")
	for _, datagram := range []bool{true, false} {
		kind := unix.SOCK_STREAM
		if datagram {
			kind = unix.SOCK_DGRAM
		}
		pair, err := unix.Socketpair(unix.AF_UNIX, kind, 0)
		if err != nil {
			t.Fatal(err)
		}
		fd, err := p.ImportSocket(os.NewFile(uintptr(pair[0]), "socketpair"), false)
		if err != nil {
			t.Fatal(err)
		}
		f.iovec(100, 200, 3)
		if datagram {
			if _, err := unix.Write(pair[1], []byte("abcdef")); err != nil {
				t.Fatal(err)
			}
			f.want(ErrSuccess, "sock_recv", uint64(fd), 100, 1, 0, 300, 304)
			if f.u32(300) != 3 || binary.LittleEndian.Uint16(f.read(304, 2)) != 1 {
				t.Fatal("truncation")
			}
			f.want(ErrSuccess, "fd_fdstat_get", uint64(fd), 0)
			if f.read(0, 1)[0] != 5 {
				t.Fatal("datagram type")
			}
		} else {
			for _, flags := range []uint64{3, 2} {
				if _, err := unix.Write(pair[1], []byte("a")); err != nil {
					t.Fatal(err)
				}
				written := make(chan error, 1)
				go func() { time.Sleep(5 * time.Millisecond); _, err := unix.Write(pair[1], []byte("bc")); written <- err }()
				f.want(ErrSuccess, "sock_recv", uint64(fd), 100, 1, flags, 300, 304)
				if err := <-written; err != nil {
					t.Fatal(err)
				}
				if f.u32(300) != 3 || string(f.read(200, 3)) != "abc" {
					t.Fatal("waitall")
				}
				if flags == 3 {
					f.want(ErrSuccess, "sock_recv", uint64(fd), 100, 1, 0, 300, 304)
				}
			}
			f.iovec(100, 200, 0)
			f.want(ErrSuccess, "sock_recv", uint64(fd), 100, 1, 0, 300, 304)
		}
		f.want(ErrSuccess, "fd_close", uint64(fd))
		unix.Close(pair[1])
	}
	if _, err := p.ImportSocket(nil, false); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	file, err := os.CreateTemp(t.TempDir(), "not-socket")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.ImportSocket(file, false); errno(err) != ErrNotsock {
		t.Fatal(err)
	}
	pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(pair[1])
	if _, err = p.ImportSocket(os.NewFile(uintptr(pair[0]), "bad listener"), true); !errors.Is(err, ErrInval) {
		t.Fatal(err)
	}
}

func TestStreamFailuresAreGuestErrors(t *testing.T) {
	for _, bad := range []int{-1, 100} {
		cfg := DefaultConfig()
		cfg.Stdout = OutputFunc(func(context.Context, []byte) (int, error) { return bad, nil })
		p := process(t, cfg)
		f := newFixture(t, p, "interpreter")
		f.iovec(100, 200, 1)
		f.want(ErrIo, "fd_write", 1, 100, 1, 300)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	stream, err := FileStream(w)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if n, err := stream.Write(context.Background(), []byte("x")); n != 0 || !errors.Is(err, unix.EPIPE) {
		t.Fatal("closed pipe", n, err)
	}
	if _, err := stream.Write(contextCanceled(), []byte("x")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if (&ExitError{Code: 1}).Error() == "" || ErrFault.Error() == "" {
		t.Fatal("error formatting")
	}
}
func contextCanceled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestProcessContextCancelsPureGuest(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		t.Run(backend, func(t *testing.T) {
			p := process(t, DefaultConfig())
			f := newFixture(t, p, backend)
			fn, err := f.in.Export("spin")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(p.Context(), time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				for {
					_, err := fn.Call(ctx, nil, vm.CallOptions{})
					if errors.Is(err, vm.ErrBusy) {
						runtime.Gosched()
						continue
					}
					done <- err
					return
				}
			}()
			for {
				_, err := f.s.Usage()
				if errors.Is(err, vm.ErrBusy) {
					break
				}
				select {
				case err := <-done:
					t.Fatal("spin ended before close", err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				default:
					runtime.Gosched()
				}
			}
			p.Close()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			case <-f.ctx.Done():
				t.Fatal("pure guest survived process close")
			}
		})
	}
}
