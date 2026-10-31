package host

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestDirectoryCapabilities(t *testing.T) {
	path := t.TempDir()
	d, err := OpenDirectory(path)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if _, err := OpenDirectory(filepath.Join(path, "missing")); err == nil {
		t.Fatal("missing root")
	}
	for _, name := range []string{"", "\x00", "/absolute", "../escape", "a/../../escape"} {
		if _, _, err := d.Open(name, unix.O_RDONLY, 0); err == nil {
			t.Fatal("invalid path", name)
		}
		if err := d.Mkdir(name); err == nil {
			t.Fatal("invalid mutation path", name)
		}
	}
	if err = d.Mkdir("a"); err != nil {
		t.Fatal(err)
	}
	f, dir, err := d.Open("a/file", unix.O_RDWR|unix.O_CREAT|unix.O_NOFOLLOW, 0600)
	if err != nil || dir != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err = f.Write([]byte("keep")); err != nil {
		t.Fatal(err)
	}
	if err = d.Symlink("file", "a/link"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := d.Open("a/link", unix.O_RDONLY|unix.O_NOFOLLOW, 0); err == nil {
		t.Fatal("followed link")
	}
	for _, name := range []string{"a/file", "a/link"} {
		if _, err := d.Stat(name, false); err != nil {
			t.Fatal(err)
		}
		if st, err := d.Stat(name, true); err != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Size != 4 {
			t.Fatal("followed file", st, err)
		}
	}
	subfile, sub, err := d.Open("a", unix.O_RDONLY, 0)
	if err != nil || sub == nil || subfile != sub.File {
		t.Fatal(err)
	}
	defer sub.Close()
	if err = Link(sub, "link", true, d, "copied"); err != nil {
		t.Fatal(err)
	}
	if err = Link(d, "a/link", true, d, "nested-link"); err != nil {
		t.Fatal(err)
	}
	if err = Rename(d, "missing/", d, "unused"); err == nil {
		t.Fatal("missing trailing source")
	}
	if err = Link(d, "a/", false, d, "directory-link"); err == nil {
		t.Fatal("directory hard link")
	}
	if err = Rename(d, "copied", sub, "moved"); err != nil {
		t.Fatal(err)
	}
	st, err := sub.Stat("moved", false)
	if err != nil || st.Size != 4 {
		t.Fatal(st, err)
	}
	at, mt := uint64(1_000_000_123), uint64(2_000_000_456)
	if err = d.SetTimes("a/link", false, at, mt, 5); err != nil {
		t.Fatal(err)
	}
	st, err = d.Stat("a/link", false)
	if err != nil || st.Mtime != mt || st.Atime != at {
		t.Fatal("symlink nanos", st, err)
	}
	if err = d.SetTimes("a/link", true, at, mt+9, 4); err != nil {
		t.Fatal(err)
	}
	st, err = Fstat(f)
	if err != nil || st.Mtime != mt+9 {
		t.Fatal("follow nanos", st, err)
	}
	if err = SetTimes(f, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err = SetTimes(f, 0, 0, 10); err != nil {
		t.Fatal(err)
	}
	if err = d.SetTimes("a/file", false, 0, 0, 10); err != nil {
		t.Fatal(err)
	}
	if Descriptor(nil) != -1 {
		t.Fatal("nil descriptor")
	}
	if err = Allocate(f, 0, 32); err != nil {
		t.Fatal(err)
	}
	if err = Allocate(f, 0, 1); err != nil {
		t.Fatal(err)
	}
	if err = Allocate(f, 0, 0); err == nil {
		t.Fatal("zero allocate")
	}
	entries, err := ReadDir(context.Background(), sub, 20)
	if err != nil || len(entries) != 5 {
		t.Fatal("entries", entries, err)
	}
	if _, err = ReadDir(context.Background(), sub, 2); !errors.Is(err, unix.ENOMEM) {
		t.Fatal("directory limit", err)
	}
	if _, err = ReadDir(context.Background(), sub, 0); !errors.Is(err, unix.ENOMEM) {
		t.Fatal(err)
	}
	if err = d.Mkdir("empty"); err != nil {
		t.Fatal(err)
	}
	_, empty, err := d.Open("empty/", unix.O_RDONLY|unix.O_NOFOLLOW, 0)
	if err != nil || empty == nil {
		t.Fatal(err)
	}
	empty.Close()
	if err = d.Remove("empty", true); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = ReadDir(ctx, sub, 20); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		call func() error
	}{
		{"mkdir", func() error { return d.Mkdir("missing/child") }},
		{"remove", func() error { return d.Remove("missing/child", false) }},
		{"rename source", func() error { return Rename(d, "missing/child", d, "new") }},
		{"rename target", func() error { return Rename(d, "a/file", d, "missing/child") }},
		{"rename trailing", func() error { return Rename(d, "a/file/", d, "x") }},
		{"link source", func() error { return Link(d, "missing/child", false, d, "new") }},
		{"link target", func() error { return Link(d, "a/file", false, d, "missing/child") }},
		{"link trailing", func() error { return Link(d, "a/file/", false, d, "new") }},
		{"link target trailing missing", func() error { return Link(d, "a/file", false, d, "new/") }},
		{"link target trailing dir", func() error { return Link(d, "a/file", false, d, "a/") }},
		{"symlink NUL", func() error { return d.Symlink("\x00", "x") }},
		{"symlink absolute", func() error { return d.Symlink("/outside", "x") }},
		{"symlink parent", func() error { return d.Symlink("x", "missing/child") }},
		{"symlink trailing", func() error { return d.Symlink("x", "trailing/") }},
		{"readlink", func() error { _, err := d.Readlink("\x00"); return err }},
		{"stat missing", func() error { _, err := d.Stat("missing/child", false); return err }},
		{"stat follow missing", func() error { _, err := d.Stat("missing", true); return err }},
		{"stat invalid", func() error { _, err := d.Stat("\x00", false); return err }},
		{"times missing", func() error { return d.SetTimes("missing/child", false, 0, 0, 5) }},
		{"times follow missing", func() error { return d.SetTimes("missing", true, 0, 0, 5) }},
		{"times invalid", func() error { return d.SetTimes("\x00", true, 0, 0, 5) }},
		{"remove trailing file", func() error { return d.Remove("a/file/", false) }},
		{"remove trailing dir", func() error { return d.Remove("a/", false) }},
		{"remove trailing missing", func() error { return d.Remove("missing/", false) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err == nil {
				t.Fatal("unexpected success")
			}
		})
	}
	if err = d.Symlink("loop", "loop"); err != nil {
		t.Fatal(err)
	}
	if err = Link(d, "loop", true, d, "never"); !errors.Is(err, unix.ELOOP) {
		t.Fatal(err)
	}
	if _, err = d.Stat("loop", true); !errors.Is(err, unix.ELOOP) {
		t.Fatal("stat followed loop", err)
	}
	if err = os.Symlink("/", filepath.Join(path, "absolute")); err != nil {
		t.Fatal(err)
	}
	if err = Link(d, "absolute", true, d, "never"); !errors.Is(err, ErrOutside) {
		t.Fatal(err)
	}
	if _, err = d.Stat("absolute", true); !errors.Is(err, ErrOutside) {
		t.Fatal("stat followed escape", err)
	}
	if err = d.Symlink("../escape", "parent-escape"); err != nil {
		t.Fatal(err)
	}
	if _, err = d.Stat("parent-escape", true); !errors.Is(err, ErrOutside) {
		t.Fatal("stat followed parent escape", err)
	}
	if err = d.SetTimes("parent-escape", true, 1, 1, 5); !errors.Is(err, ErrOutside) {
		t.Fatal("timestamp followed parent escape", err)
	}
	if err = Link(d, "missing", true, d, "never"); err == nil {
		t.Fatal("missing followed link")
	}
	if err = d.Remove("a/moved", false); err != nil {
		t.Fatal(err)
	}
	dir2, err := OpenDirectory(path)
	if err != nil {
		t.Fatal(err)
	}
	dir2.Root.Close()
	if _, _, err = dir2.Open(".", 0, 0); err == nil {
		t.Fatal("closed root")
	}
	if _, _, err = dir2.Open(".", unix.O_NOFOLLOW, 0); err == nil {
		t.Fatal("closed root parent")
	}
	if _, err = ReadDir(context.Background(), dir2, 20); err == nil {
		t.Fatal("closed root read")
	}
	dir2.Close()
	closed, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	if Descriptor(closed) != -1 {
		t.Fatal("closed descriptor")
	}
	if err := Allocate(closed, 0, 1); err == nil {
		t.Fatal("allocate closed")
	}
	if err := SetTimes(closed, 0, 1, 4); err == nil {
		t.Fatal("timestamp closed")
	}
}

func TestNoEscapeThroughConcurrentParents(t *testing.T) {
	outside, inside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "sentinel"), []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := OpenDirectory(inside)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err = d.Mkdir("safe"); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(".", filepath.Join(inside, "self")); err != nil {
		t.Fatal(err)
	}
	for _, op := range []func() error{
		func() error { _, err := d.Stat("self/..", false); return err },
		func() error { return d.SetTimes("self/..", false, 1, 1, 5) },
		func() error {
			f, sub, err := d.Open("self/..", unix.O_NOFOLLOW, 0)
			if err == nil {
				if sub != nil {
					sub.Close()
				} else {
					f.Close()
				}
			}
			return err
		},
	} {
		if err := op(); !errors.Is(err, ErrOutside) {
			t.Fatal("dot-dot escape", err)
		}
	}
	stop := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			os.Rename(filepath.Join(inside, "safe"), filepath.Join(inside, "saved"))
			os.Symlink(outside, filepath.Join(inside, "safe"))
			runtime.Gosched()
			os.Remove(filepath.Join(inside, "safe"))
			os.Rename(filepath.Join(inside, "saved"), filepath.Join(inside, "safe"))
		}
	}()
	for i := 0; i < 100; i++ {
		f, sub, err := d.Open("safe/sentinel", unix.O_CREAT|unix.O_TRUNC|unix.O_WRONLY|unix.O_NOFOLLOW, 0600)
		if err == nil {
			if sub != nil {
				sub.Close()
			} else {
				f.Close()
			}
		}
		d.Mkdir("safe/new")
		d.SetTimes("safe/sentinel", true, 1, 1, 5)
		d.Remove("safe/sentinel", false)
	}
	close(stop)
	workers.Wait()
	b, err := os.ReadFile(filepath.Join(outside, "sentinel"))
	if err != nil || string(b) != "untouched" {
		t.Fatal("outside modified", string(b), err)
	}
	if _, err = os.Stat(filepath.Join(outside, "new")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("created outside", err)
	}
}

func TestFileReadinessAndClose(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	// A deadline can expire before its cancellation notification is delivered.
	if _, _, err := Poll(pendingDeadline{context.Background()}, Descriptor(r), true); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("expired deadline entered a kernel wait", err)
	}
	if _, _, ready, err := Ready(int(r.Fd()), true); err != nil || ready {
		t.Fatal("empty pipe", ready, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, _, err = Poll(ctx, int(r.Fd()), true); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()
	if _, _, err = Poll(ctx2, int(r.Fd()), true); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err = w.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if n, _, ready, err := Ready(int(r.Fd()), true); err != nil || !ready || n != 3 {
		t.Fatal(n, ready, err)
	}
	if n, h, err := Poll(context.Background(), int(r.Fd()), true); err != nil || h || n != 3 {
		t.Fatal("readiness", n, h, err)
	}
	if _, _, ready, err := Ready(int(w.Fd()), false); err != nil || !ready {
		t.Fatal("writable", ready, err)
	}
	fd := int(r.Fd())
	r.Close()
	if _, _, _, err = Ready(fd, true); !errors.Is(err, unix.EBADF) {
		t.Fatal("closed readiness", err)
	}
	if _, _, err = Poll(context.Background(), fd, true); !errors.Is(err, unix.EBADF) {
		t.Fatal("closed poll", err)
	}
}

type pendingDeadline struct{ context.Context }

func (pendingDeadline) Deadline() (time.Time, bool) { return time.Unix(0, 0), true }

func TestUnsignedTimestampDomain(t *testing.T) {
	for _, value := range []uint64{0, 1, 1<<63 - 1, 1 << 63, ^uint64(0)} {
		seconds, nanos := int64(value/1_000_000_000), int64(value%1_000_000_000)
		got, ok := Timestamp(seconds, nanos)
		if !ok || got != value {
			t.Fatal(value, got, ok)
		}
	}
	for _, v := range [][2]int64{{-1, 0}, {0, -1}, {0, 1_000_000_000}, {18_446_744_074, 0}, {18_446_744_073, 999_999_999}} {
		if _, ok := Timestamp(v[0], v[1]); ok {
			t.Fatal("unrepresentable host time", v)
		}
	}
}

func TestSignedFilesystemInstants(t *testing.T) {
	root := t.TempDir()
	d, err := OpenDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	f, child, err := d.Open("file", unix.O_CREAT|unix.O_RDWR, 0600)
	if err != nil || child != nil {
		t.Fatal(child, err)
	}
	defer f.Close()
	at, mt := Instant{-2, 7}, Instant{1<<32 + 5, 9}
	if err := SetInstants(f, at, mt, 5); err != nil {
		t.Fatal(err)
	}
	st, err := Fstat(f)
	if err != nil || st.Instants[0] != at || st.Instants[1] != mt || !st.TimeOverflow {
		t.Fatal("signed stat narrowed or legacy overflow changed", st, err)
	}
	if err := d.Symlink("file", "link"); err != nil {
		t.Fatal(err)
	}
	if err := d.SetInstants("link", false, at, mt, 5); err != nil {
		t.Fatal(err)
	}
	st, err = d.Stat("link", false)
	if err != nil || st.Instants[0] != at || st.Instants[1] != mt {
		t.Fatal("symlink timestamps", st, err)
	}
	mt = Instant{3, 11}
	if err := d.SetInstants("link", true, at, mt, 4); err != nil {
		t.Fatal(err)
	}
	st, err = Fstat(f)
	if err != nil || st.Instants[1] != mt {
		t.Fatal("follow timestamps", st, err)
	}
	if err := SetInstants(f, Instant{}, Instant{}, 0); err != nil {
		t.Fatal(err)
	}
	if err := SetInstants(f, Instant{}, Instant{}, 10); err != nil {
		t.Fatal(err)
	}
	for _, call := range []func() error{
		func() error { return SetInstants(f, Instant{Nanoseconds: 1e9}, Instant{}, 1) },
		func() error { return SetInstants(f, Instant{}, Instant{}, 16) },
		func() error { return d.SetInstants("file", false, Instant{}, Instant{}, 16) },
		func() error { return d.SetInstants("missing/child", false, at, mt, 5) },
		func() error { return d.SetInstants("missing", true, at, mt, 5) },
		func() error { return d.SetInstants("\x00", true, at, mt, 5) },
	} {
		if err := call(); err == nil {
			t.Fatal("invalid timestamp operation accepted")
		}
	}
	for _, follow := range []bool{false, true} {
		if err := d.SetInstants("../outside", follow, at, mt, 5); !errors.Is(err, ErrOutside) {
			t.Fatal("timestamp capability escaped", err)
		}
	}
	ts, err := instantSpecs(Instant{Seconds: -1 << 63}, Instant{Seconds: 1<<63 - 1, Nanoseconds: 999999999}, 5)
	if err != nil || ts[0].Sec != -1<<63 || ts[1].Sec != 1<<63-1 || ts[1].Nsec != 999999999 {
		t.Fatal("full signed range", ts, err)
	}
	f.Close()
	if err := SetInstants(f, at, mt, 5); err == nil {
		t.Fatal("closed descriptor accepted")
	}
}
