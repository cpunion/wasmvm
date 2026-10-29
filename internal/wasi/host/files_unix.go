//go:build linux || darwin

// Package host provides handle-based host capabilities shared by WASI adapters.
package host

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

var ErrOutside = errors.New("path escapes capability directory")

type Directory struct {
	Root *os.Root
	File *os.File
}

func OpenDirectory(name string) (*Directory, error) {
	r, err := os.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	f, err := r.Open(".")
	if err != nil {
		r.Close()
		return nil, err
	}
	return &Directory{r, f}, nil
}
func (d *Directory) Close() { d.File.Close(); d.Root.Close() }

func pathError(err error) error {
	if err != nil && strings.Contains(err.Error(), "path escapes from parent") {
		return ErrOutside
	}
	return err
}
func checkPath(name string) error {
	if name == "" {
		return unix.ENOENT
	}
	if strings.IndexByte(name, 0) >= 0 {
		return unix.EINVAL
	}
	if strings.HasPrefix(name, "/") {
		return ErrOutside
	}
	depth := 0
	for _, part := range strings.Split(name, "/") {
		switch part {
		case "", ".":
		case "..":
			depth--
		default:
			depth++
		}
		if depth < 0 {
			return ErrOutside
		}
	}
	return nil
}
func (d *Directory) Open(name string, flags int, perm os.FileMode) (*os.File, *Directory, error) {
	if err := checkPath(name); err != nil {
		return nil, nil, err
	}
	for attempt := 0; attempt < 40; attempt++ {
		var f *os.File
		var err error
		if flags&unix.O_NOFOLLOW != 0 {
			parent, leaf, trailing, err := d.parent(name)
			if err != nil {
				return nil, nil, err
			}
			actual := flags | unix.O_NONBLOCK | unix.O_CLOEXEC
			if trailing {
				actual |= unix.O_DIRECTORY
			}
			fd, openErr := unix.Openat(Descriptor(parent.File), leaf, actual, uint32(perm))
			parent.Close()
			if openErr != nil {
				return nil, nil, openErr
			}
			f = os.NewFile(uintptr(fd), name)
		} else {
			f, err = d.Root.OpenFile(name, flags|unix.O_NONBLOCK, perm)
			if err != nil {
				return nil, nil, pathError(err)
			}
		}
		first, err := Fstat(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		if first.Mode&unix.S_IFMT != unix.S_IFDIR {
			return f, nil, nil
		}
		root, err := d.Root.OpenRoot(name)
		if err != nil {
			f.Close()
			return nil, nil, pathError(err)
		}
		same, err := root.Open(".")
		if err != nil {
			root.Close()
			f.Close()
			return nil, nil, err
		}
		second, e2 := Fstat(same)
		f.Close()
		if e2 != nil {
			root.Close()
			same.Close()
			return nil, nil, e2
		}
		if first.Dev == second.Dev && first.Inode == second.Inode {
			return same, &Directory{root, same}, nil
		}
		root.Close()
		same.Close()
	}
	return nil, nil, unix.EAGAIN
}

// parent resolves all ancestor components through os.Root, then borrows a
// stable directory descriptor. Mutations never follow the final symlink.
func (d *Directory) parent(name string) (*Directory, string, bool, error) {
	if err := checkPath(name); err != nil {
		return nil, "", false, err
	}
	trailing := strings.HasSuffix(name, "/")
	trimmed := strings.TrimRight(name, "/")
	base, leaf := ".", trimmed
	if i := strings.LastIndexByte(trimmed, '/'); i >= 0 {
		base, leaf = trimmed[:i], trimmed[i+1:]
	}
	if leaf == "." || leaf == ".." {
		base = name
		leaf = "."
	}
	r, err := d.Root.OpenRoot(base)
	if err != nil {
		return nil, "", false, pathError(err)
	}
	f, err := r.Open(".")
	if err != nil {
		r.Close()
		return nil, "", false, err
	}
	return &Directory{r, f}, leaf, trailing, nil
}
func (d *Directory) Stat(name string, follow bool) (Stat, error) {
	if err := checkPath(name); err != nil {
		return Stat{}, err
	}
	if follow || strings.HasSuffix(name, "/") {
		f, err := d.Root.OpenFile(name, pathHandleFlags(), 0)
		if err != nil {
			return Stat{}, pathError(err)
		}
		defer f.Close()
		return Fstat(f)
	}
	p, leaf, _, err := d.parent(name)
	if err != nil {
		return Stat{}, err
	}
	defer p.Close()
	var st unix.Stat_t
	err = unix.Fstatat(Descriptor(p.File), leaf, &st, unix.AT_SYMLINK_NOFOLLOW)
	return convertStat(st), err
}
func (d *Directory) Mkdir(name string) error {
	p, leaf, _, err := d.parent(name)
	if err != nil {
		return err
	}
	defer p.Close()
	return unix.Mkdirat(Descriptor(p.File), leaf, 0777)
}
func (d *Directory) Remove(name string, directory bool) error {
	p, leaf, trailing, err := d.parent(name)
	if err != nil {
		return err
	}
	defer p.Close()
	if trailing && !directory {
		st, err := d.Stat(strings.TrimRight(name, "/"), false)
		if err != nil {
			return err
		}
		if st.Mode&unix.S_IFMT == unix.S_IFDIR {
			return unix.EISDIR
		}
		return unix.ENOTDIR
	}
	flags := 0
	if directory {
		flags = unix.AT_REMOVEDIR
	}
	return unix.Unlinkat(Descriptor(p.File), leaf, flags)
}
func Rename(from *Directory, old string, to *Directory, name string) error {
	p, a, trail, err := from.parent(old)
	if err != nil {
		return err
	}
	defer p.Close()
	q, b, destTrail, err := to.parent(name)
	if err != nil {
		return err
	}
	defer q.Close()
	if trail || destTrail {
		st, err := from.Stat(strings.TrimRight(old, "/"), false)
		if err != nil {
			return err
		}
		if st.Mode&unix.S_IFMT != unix.S_IFDIR {
			return unix.ENOTDIR
		}
	}
	return unix.Renameat(Descriptor(p.File), a, Descriptor(q.File), b)
}
func Link(from *Directory, old string, follow bool, to *Directory, name string) error {
	// Resolve a followed final symlink within the original capability. Linkat
	// still uses NOFOLLOW: concurrent replacement cannot follow an escape link.
	if follow {
		for depth := 0; ; depth++ {
			if depth >= 40 {
				return unix.ELOOP
			}
			st, err := from.Stat(old, false)
			if err != nil {
				return err
			}
			if st.Mode&unix.S_IFMT != unix.S_IFLNK {
				break
			}
			target, err := from.Readlink(old)
			if err != nil {
				return err
			}
			if strings.HasPrefix(target, "/") {
				return ErrOutside
			}
			if i := strings.LastIndexByte(old, '/'); i >= 0 {
				old = old[:i+1] + target
			} else {
				old = target
			}
		}
	}
	p, a, trail, err := from.parent(old)
	if err != nil {
		return err
	}
	defer p.Close()
	q, b, destTrail, err := to.parent(name)
	if err != nil {
		return err
	}
	defer q.Close()
	if trail {
		st, err := from.Stat(old, true)
		if err != nil {
			return err
		}
		if st.Mode&unix.S_IFMT != unix.S_IFDIR {
			return unix.ENOTDIR
		}
	}
	if destTrail {
		if _, err := to.Stat(name, true); err != nil {
			return err
		}
		return unix.EEXIST
	}
	return unix.Linkat(Descriptor(p.File), a, Descriptor(q.File), b, 0)
}
func (d *Directory) Symlink(target, name string) error {
	if strings.HasPrefix(target, "/") {
		return ErrOutside
	}
	if strings.IndexByte(target, 0) >= 0 {
		return unix.EINVAL
	}
	p, leaf, trailing, err := d.parent(name)
	if err != nil {
		return err
	}
	defer p.Close()
	if trailing {
		return unix.ENOENT
	}
	return unix.Symlinkat(target, Descriptor(p.File), leaf)
}
func (d *Directory) Readlink(name string) (string, error) {
	if err := checkPath(name); err != nil {
		return "", err
	}
	s, err := d.Root.Readlink(name)
	return s, pathError(err)
}
func (d *Directory) SetTimes(name string, follow bool, at, mt uint64, flags uint32) error {
	if follow || strings.HasSuffix(name, "/") {
		if err := checkPath(name); err != nil {
			return err
		}
		f, err := d.Root.OpenFile(name, pathHandleFlags(), 0)
		if err != nil {
			return pathError(err)
		}
		defer f.Close()
		return SetTimes(f, at, mt, flags)
	}
	p, leaf, _, err := d.parent(name)
	if err != nil {
		return err
	}
	defer p.Close()
	return unix.UtimesNanoAt(Descriptor(p.File), leaf, timestamps(at, mt, flags), unix.AT_SYMLINK_NOFOLLOW)
}
func timestamps(at, mt uint64, flags uint32) []unix.Timespec {
	ts := []unix.Timespec{{Nsec: utimeOmit}, {Nsec: utimeOmit}}
	for i, v := range []uint64{at, mt} {
		if flags&(1<<uint(i*2)) != 0 {
			ts[i] = unix.Timespec{Sec: int64(v / 1_000_000_000), Nsec: int64(v % 1_000_000_000)}
		}
		if flags&(2<<uint(i*2)) != 0 {
			ts[i] = unix.Timespec{Nsec: utimeNow}
		}
	}
	return ts
}

type Stat struct {
	Dev, Inode, Nlink, Size, Atime, Mtime, Ctime uint64
	Mode                                         uint32
	TimeOverflow                                 bool
}

func Fstat(f *os.File) (Stat, error) {
	var st unix.Stat_t
	err := unix.Fstat(Descriptor(f), &st)
	return convertStat(st), err
}

type Dirent struct {
	Name  string
	Inode uint64
	Mode  uint32
}

func ReadDir(ctx context.Context, d *Directory, maximum uint32) ([]Dirent, error) {
	f, err := d.Root.Open(".")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := Fstat(f)
	if err != nil {
		return nil, err
	}
	entries := []Dirent{{".", st.Inode, unix.S_IFDIR}, {"..", st.Inode, unix.S_IFDIR}}
	if maximum < 2 {
		return nil, unix.ENOMEM
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		names, err := f.Readdirnames(128)
		if uint64(len(entries))+uint64(len(names)) > uint64(maximum) {
			return nil, unix.ENOMEM
		}
		for _, name := range names {
			var s unix.Stat_t
			if err := unix.Fstatat(Descriptor(f), name, &s, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				if errors.Is(err, unix.ENOENT) {
					continue
				}
				return nil, err
			}
			entries = append(entries, Dirent{name, uint64(s.Ino), uint32(s.Mode)})
		}
		if err == io.EOF {
			return entries, nil
		}
		if err != nil {
			return nil, err
		}
	}
}
func Poll(ctx context.Context, fd int, read bool) (uint64, bool, error) {
	event := int16(unix.POLLOUT)
	if read {
		event = unix.POLLIN
	}
	for {
		if err := ctx.Err(); err != nil {
			return 0, false, err
		}
		ms := 50
		if end, ok := ctx.Deadline(); ok {
			remaining := time.Until(end)
			if remaining <= 0 {
				return 0, false, context.DeadlineExceeded
			}
			if remaining < 50*time.Millisecond {
				ms = int((remaining + time.Millisecond - 1) / time.Millisecond)
			}
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: event}}
		_, err := unix.Poll(fds, ms)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, false, err
		}
		if fds[0].Revents&unix.POLLNVAL != 0 {
			return 0, false, unix.EBADF
		}
		if fds[0].Revents&(event|unix.POLLHUP|unix.POLLERR) != 0 {
			n := uint64(0)
			if read {
				if count, err := pendingBytes(fd); err == nil && count > 0 {
					n = uint64(count)
				}
			}
			return n, fds[0].Revents&unix.POLLHUP != 0, nil
		}
	}
}

// Ready is a nonblocking readiness snapshot used before poll creates waiters.
func Ready(fd int, read bool) (uint64, bool, bool, error) {
	event := int16(unix.POLLOUT)
	if read {
		event = unix.POLLIN
	}
	fds := []unix.PollFd{{Fd: int32(fd), Events: event}}
	_, err := unix.Poll(fds, 0)
	if err != nil {
		return 0, false, false, err
	}
	if fds[0].Revents&unix.POLLNVAL != 0 {
		return 0, false, false, unix.EBADF
	}
	ready := fds[0].Revents&(event|unix.POLLHUP|unix.POLLERR) != 0
	n := uint64(0)
	if ready && read {
		if count, err := pendingBytes(fd); err == nil && count > 0 {
			n = uint64(count)
		}
	}
	return n, fds[0].Revents&unix.POLLHUP != 0, ready, nil
}

// Timestamp preserves the entire unsigned P1 nanosecond domain. Values outside
// that domain are reported rather than wrapping an unrepresentable host date.
func Timestamp(seconds, nanos int64) (uint64, bool) {
	if seconds < 0 || nanos < 0 || nanos >= 1_000_000_000 || uint64(seconds) > ^uint64(0)/1_000_000_000 {
		return 0, false
	}
	base := uint64(seconds) * 1_000_000_000
	if uint64(nanos) > ^uint64(0)-base {
		return 0, false
	}
	return base + uint64(nanos), true
}
func statTimes(at, mt, ct unix.Timespec) (uint64, uint64, uint64, bool) {
	a, ok1 := Timestamp(int64(at.Sec), int64(at.Nsec))
	m, ok2 := Timestamp(int64(mt.Sec), int64(mt.Nsec))
	c, ok3 := Timestamp(int64(ct.Sec), int64(ct.Nsec))
	return a, m, c, !ok1 || !ok2 || !ok3
}
