//go:build linux || darwin

package preview1

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestFileStatusAndLimits(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Preopens = []Preopen{{"/", t.TempDir(), false}}
	p := process(t, cfg)
	f := newFixture(t, p, "native")
	f.write(400, []byte("file"))
	f.want(ErrSuccess, "path_open", 3, 0, 400, 4, 1, uint64(regularRights), 0, 0, 500)
	fd := uint64(f.u32(500))
	f.want(ErrSuccess, "fd_sync", fd)
	f.want(ErrSuccess, "fd_datasync", fd)
	f.want(ErrFbig, "fd_allocate", fd, 1<<63, 1)
	f.want(ErrFbig, "fd_filestat_set_size", fd, 1<<63)
	f.want(ErrInval, "fd_advise", fd, 0, 1, 6)
	f.want(ErrOverflow, "fd_advise", fd, 1<<63, 1, 0)
	f.want(ErrInval, "fd_seek", fd, 0, 3, 0)
	f.want(ErrFault, "fd_seek", fd, 0, 0, 131071)
	f.want(ErrFault, "fd_tell", fd, 131071)
	f.want(ErrInval, "fd_filestat_set_times", fd, 0, 0, 15)
	f.want(ErrSuccess, "fd_fdstat_set_flags", fd, 31)
	f.iovec(100, 200, 3)
	f.write(200, []byte("abc"))
	f.want(ErrSuccess, "fd_write", fd, 100, 1, 300)
	f.want(ErrSuccess, "fd_seek", fd, 0, 0, 500)
	f.want(ErrSuccess, "fd_read", fd, 100, 1, 300)
	if err := os.Chtimes(cfg.Preopens[0].HostPath+"/file", time.Unix(-1, 0), time.Unix(-1, 0)); err != nil {
		t.Fatal(err)
	}
	f.want(ErrOverflow, "fd_filestat_get", fd, 1000)
	f.want(ErrOverflow, "path_filestat_get", 3, 0, 400, 4, 1000)
	f.want(ErrOverflow, "fd_pread", fd, 100, 1, 1<<63, 300)
	f.want(ErrFault, "fd_write", fd, 100, 1, 131071)
	f.want(ErrFault, "fd_readdir", 3, 100, 16, 0, 131071)
	f.want(ErrBadf, "fd_readdir", 999, 100, 16, 0, 300)
	f.want(ErrNotsup, "fd_fdstat_set_flags", 1, 2)
	p.config.Limits.MaxIOBytes = 4
	f.iovec(100, 200, 4)
	f.want(ErrNomem, "fd_write", fd, 100, 1, 300)
	f.want(ErrNomem, "random_get", 200, 8)
	p.config.Limits.MaxDirectoryEntries = 1
	f.want(ErrNomem, "fd_readdir", 3, 100, 4, 0, 300)
	for _, test := range []struct {
		mode uint32
		want byte
	}{{unix.S_IFBLK, 1}, {unix.S_IFCHR, 2}, {unix.S_IFDIR, 3}, {unix.S_IFREG, 4}, {unix.S_IFSOCK, 6}, {unix.S_IFLNK, 7}, {unix.S_IFIFO, 0}} {
		if got := filetype(test.mode); got != test.want {
			t.Fatal(test.mode, got, test.want)
		}
	}
}

func TestFileStreamFlagsAndLifetime(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "stream")
	if err != nil {
		t.Fatal(err)
	}
	stream, err := FileStream(file)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	setter := stream.(interface{ SetFlags(uint16) error })
	if err := setter.SetFlags(31); err != nil {
		t.Fatal(err)
	}
	if n, err := stream.Write(context.Background(), []byte("abc")); n != 3 || err != nil {
		t.Fatal(n, err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	var b [3]byte
	if n, err := stream.Read(context.Background(), b[:]); n != 3 || err != nil || string(b[:]) != "abc" {
		t.Fatal(n, string(b[:]), err)
	}
	stream.Close()
	stream.Close()
	if _, err := stream.Read(context.Background(), b[:]); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if _, err := stream.Write(context.Background(), b[:]); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if _, _, err := stream.Poll(context.Background(), true); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if _, _, _, err := stream.Ready(context.Background(), true); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	if err := setter.SetFlags(0); !errors.Is(err, ErrBadf) {
		t.Fatal(err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	input, err := FileStream(r)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if err := input.(interface{ SetFlags(uint16) error }).SetFlags(2); !errors.Is(err, ErrNotsup) {
		t.Fatal(err)
	}
	if err := input.(interface{ SetFlags(uint16) error }).SetFlags(4); err != nil {
		t.Fatal(err)
	}
	if n, err := input.Read(context.Background(), b[:]); n != 0 || !errors.Is(err, unix.EAGAIN) {
		t.Fatal("nonblock", n, err)
	}
	if err := input.(interface{ SetFlags(uint16) error }).SetFlags(0); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := input.Read(context.Background(), b[:]); done <- err }()
	state := input.(*fileStream)
	deadline := time.Now().Add(time.Second)
	for {
		state.mu.Lock()
		active := state.active
		state.mu.Unlock()
		if active != 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("read did not start")
		}
		time.Sleep(time.Millisecond)
	}
	input.Close()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close left a blocked stream")
	}
}
