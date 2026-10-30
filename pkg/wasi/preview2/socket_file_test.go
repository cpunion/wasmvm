//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"golang.org/x/sys/unix"
)

func TestSocketFileLeaseAndClose(t *testing.T) {
	for _, family := range []uint32{0, 1} {
		for _, kind := range []int{unix.SOCK_STREAM, unix.SOCK_DGRAM} {
			file, err := newSocketFile(nil, family, kind)
			if err != nil {
				t.Fatal(family, kind, err)
			}
			fd, err := file.begin(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
			if err != nil || flags&unix.O_NONBLOCK == 0 {
				t.Fatal(flags, err)
			}
			flags, err = unix.FcntlInt(uintptr(fd), unix.F_GETFD, 0)
			if err != nil || flags&unix.FD_CLOEXEC == 0 {
				t.Fatal(flags, err)
			}
			file.Close()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := file.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
				t.Fatal("closed before syscall lease drained", err)
			}
			if _, err := file.file.Stat(); err != nil {
				t.Fatal("physical close raced lease", err)
			}
			file.end()
			var wg sync.WaitGroup
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					file.Close()
					if err := file.WaitClosed(nil); err != nil {
						t.Error(err)
					}
				}()
			}
			wg.Wait()
			if _, err := file.file.Stat(); !errors.Is(err, os.ErrClosed) {
				t.Fatal(err)
			}
			if _, err := file.begin(context.Background()); !errors.Is(err, os.ErrClosed) {
				t.Fatal(err)
			}
		}
	}
}
func TestSocketFileInvalidAndCancelled(t *testing.T) {
	var zero socketFile
	zero.Close()
	if err := zero.WaitClosed(nil); !errors.Is(err, os.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := newSocketFile(nil, 2, unix.SOCK_STREAM); networkError(err) != NetworkNotSupported {
		t.Fatal(err)
	}
	if _, err := newSocketFile(nil, 0, unix.SOCK_RAW); networkError(err) != NetworkInvalidArgument {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := newSocketFile(ctx, 0, unix.SOCK_STREAM); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	parent, cancelParent := context.WithCancel(context.Background())
	file, err := newSocketFile(parent, 0, unix.SOCK_STREAM)
	if err != nil {
		t.Fatal(err)
	}
	cancelParent()
	if _, err := file.begin(context.Background()); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := file.begin(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	file.Close()
	if err := file.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	for _, entry := range socketErrnos {
		if code := socketError(entry.err); code != entry.code {
			t.Fatal(entry, code)
		}
	}
	if code := socketError(context.DeadlineExceeded); code != NetworkTimeout {
		t.Fatal(code)
	}
}
