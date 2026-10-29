//go:build linux || darwin

package preview1

import (
	"encoding/binary"
	"io"
	"math"
	"runtime"
	"strings"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
)

func (p *Process) invoke(c *abiCall, name string, a []uint64) error {
	switch name {
	case "args_get", "environ_get", "args_sizes_get", "environ_sizes_get":
		list := p.config.Args
		if strings.HasPrefix(name, "environ") {
			list = p.config.Env
		}
		n := uint64(0)
		for _, s := range list {
			n += uint64(len(s)) + 1
		}
		if strings.HasSuffix(name, "sizes_get") {
			if err := c.check(a[0], 4); err != nil {
				return err
			}
			if err := c.check(a[1], 4); err != nil {
				return err
			}
			if err := c.u32(a[0], uint32(len(list))); err != nil {
				return err
			}
			return c.u32(a[1], uint32(n))
		}
		if err := c.check(a[0], uint64(len(list))*4); err != nil {
			return err
		}
		if err := c.check(a[1], n); err != nil {
			return err
		}
		at := a[1]
		for i, s := range list {
			if err := c.u32(a[0]+uint64(i)*4, uint32(at)); err != nil {
				return err
			}
			if err := c.write(at, append([]byte(s), 0)); err != nil {
				return err
			}
			at += uint64(len(s)) + 1
		}
		return nil
	case "clock_res_get", "clock_time_get":
		out := a[len(a)-1]
		if err := c.check(out, 8); err != nil {
			return err
		}
		var v uint64
		var err error
		if name == "clock_res_get" {
			v, err = p.clock.Resolution(uint32(a[0]))
			if err == nil && v == 0 {
				return ErrInval
			}
		} else {
			v, err = p.clock.Now(uint32(a[0]))
		}
		if err != nil {
			return err
		}
		return c.u64(out, v)
	case "random_get":
		if err := c.check(a[0], a[1]); err != nil {
			return err
		}
		b, err := c.buffer(a[1])
		if err != nil {
			return err
		}
		if len(b) > 0 {
			if err = p.config.Random(c.ctx, b); err != nil {
				return err
			}
		}
		return c.write(a[0], b)
	case "sched_yield":
		runtime.Gosched()
		return nil
	case "proc_exit":
		p.shutdown(true, uint32(a[0]))
		return &ExitError{uint32(a[0])}
	case "proc_raise":
		if a[0] > 30 {
			return ErrInval
		}
		if a[0] == 0 {
			return nil
		}
		return ErrNotsup
	case "fd_close":
		return p.closeFD(uint32(a[0]))
	case "fd_renumber":
		return p.renumber(uint32(a[0]), uint32(a[1]))
	case "fd_fdstat_set_rights":
		p.mu.Lock()
		defer p.mu.Unlock()
		d := p.fds[uint32(a[0])]
		if d == nil {
			return ErrBadf
		}
		if Rights(a[1])&^d.base != 0 || Rights(a[2])&^d.inherit != 0 {
			return ErrNotcapable
		}
		d.base = Rights(a[1])
		d.inherit = Rights(a[2])
		return nil
	case "fd_read", "fd_write", "fd_pread", "fd_pwrite":
		return p.fdIO(c, name, a)
	case "fd_readdir":
		return p.readDirectory(c, a)
	case "poll_oneoff":
		return p.poll(c, a)
	case "sock_accept", "sock_recv", "sock_send", "sock_shutdown":
		return p.socket(c, name, a)
	}
	if strings.HasPrefix(name, "path_") {
		return p.path(c, name, a)
	}
	return p.fd(c, name, a)
}
func (p *Process) renumber(from, to uint32) error {
	p.mu.Lock()
	d := p.fds[from]
	if d == nil {
		p.mu.Unlock()
		return ErrBadf
	}
	if p.fds[to] == nil {
		p.mu.Unlock()
		return ErrBadf
	}
	if from == to {
		p.mu.Unlock()
		return nil
	}
	var closing *resource
	if old := p.fds[to]; old != nil {
		if p.dropOwnerLocked(old.resource) {
			closing = old.resource
		}
	}
	delete(p.fds, from)
	p.fds[to] = d
	p.mu.Unlock()
	if closing != nil {
		p.closeResource(closing)
	}
	return nil
}
func (p *Process) fd(c *abiCall, name string, a []uint64) error {
	rights := Rights(0)
	switch name {
	case "fd_advise":
		rights = RightFdAdvise
	case "fd_allocate":
		rights = RightFdAllocate
	case "fd_datasync":
		rights = RightFdDatasync
	case "fd_sync":
		rights = RightFdSync
	case "fd_fdstat_set_flags":
		rights = RightFdFdstatSetFlags
	case "fd_filestat_get":
		rights = RightFdFilestatGet
	case "fd_filestat_set_size":
		rights = RightFdFilestatSetSize
	case "fd_filestat_set_times":
		rights = RightFdFilestatSetTimes
	case "fd_seek":
		rights = RightFdSeek
	case "fd_tell":
		rights = RightFdTell
	}
	d, release, err := p.hold(uint32(a[0]), rights)
	if err != nil {
		return err
	}
	defer release()
	r := d.resource
	r.mu.Lock()
	defer r.mu.Unlock()
	switch name {
	case "fd_prestat_get":
		if d.preopen == "" {
			return ErrBadf
		}
		if err := c.check(a[1], 8); err != nil {
			return err
		}
		var b [8]byte
		binary.LittleEndian.PutUint32(b[4:], uint32(len(d.preopen)))
		return c.write(a[1], b[:])
	case "fd_prestat_dir_name":
		if d.preopen == "" {
			return ErrBadf
		}
		if a[2] < uint64(len(d.preopen)) {
			return ErrNametoolong
		}
		if err := c.check(a[1], a[2]); err != nil {
			return err
		}
		return c.write(a[1], []byte(d.preopen))
	case "fd_fdstat_get":
		if err := c.check(a[1], 24); err != nil {
			return err
		}
		var b [24]byte
		if r.stream != nil {
			b[0] = 2
		} else {
			st, err := host.Fstat(r.file)
			if err != nil {
				return err
			}
			b[0] = filetype(st.Mode)
			if r.socket {
				if r.datagram {
					b[0] = 5
				} else {
					b[0] = 6
				}
			}
		}
		binary.LittleEndian.PutUint16(b[2:], r.flags)
		binary.LittleEndian.PutUint64(b[8:], uint64(d.base))
		binary.LittleEndian.PutUint64(b[16:], uint64(d.inherit))
		return c.write(a[1], b[:])
	case "fd_filestat_get":
		if err := c.check(a[1], 64); err != nil {
			return err
		}
		st := host.Stat{Mode: unix.S_IFCHR, Nlink: 1}
		if r.file != nil {
			st, err = host.Fstat(r.file)
			if err != nil {
				return err
			}
		}
		if st.TimeOverflow {
			return ErrOverflow
		}
		return c.write(a[1], statBytes(st))
	case "fd_fdstat_set_flags":
		if a[1]&^uint64(31) != 0 {
			return ErrInval
		}
		if r.stream != nil {
			if setter, ok := r.stream.(interface{ SetFlags(uint16) error }); ok {
				if err := setter.SetFlags(uint16(a[1])); err != nil {
					return err
				}
			} else if uint16(a[1]) != r.flags {
				return ErrNotsup
			}
		}
		if r.file != nil {
			flags, err := unix.FcntlInt(uintptr(host.Descriptor(r.file)), unix.F_GETFL, 0)
			if err != nil {
				return err
			}
			flags &^= unix.O_APPEND
			flags |= unix.O_NONBLOCK
			if a[1]&1 != 0 {
				flags |= unix.O_APPEND
			}
			if _, err = unix.FcntlInt(uintptr(host.Descriptor(r.file)), unix.F_SETFL, flags); err != nil {
				return err
			}
		}
		r.flags = uint16(a[1])
		return nil
	case "fd_advise":
		if r.dir != nil {
			return ErrIsdir
		}
		if r.file == nil {
			return ErrSpipe
		}
		if a[3] > 5 {
			return ErrInval
		}
		if a[1] > math.MaxInt64 || a[2] > math.MaxInt64-a[1] {
			return ErrOverflow
		}
		return nil
	case "fd_allocate":
		if r.dir != nil {
			return ErrIsdir
		}
		if r.file == nil {
			return ErrSpipe
		}
		if a[1] > math.MaxInt64 || a[2] > math.MaxInt64-a[1] {
			return ErrFbig
		}
		return host.Allocate(r.file, int64(a[1]), int64(a[2]))
	case "fd_filestat_set_size":
		if r.dir != nil {
			return ErrIsdir
		}
		if r.file == nil {
			return ErrSpipe
		}
		if a[1] > math.MaxInt64 {
			return ErrFbig
		}
		return r.file.Truncate(int64(a[1]))
	case "fd_filestat_set_times":
		if err := validTimes(a[1], a[2], a[3]); err != nil {
			return err
		}
		if r.file == nil {
			return ErrNotsup
		}
		return host.SetTimes(r.file, a[1], a[2], uint32(a[3]))
	case "fd_datasync", "fd_sync":
		if r.file == nil {
			return nil
		}
		return r.file.Sync()
	case "fd_seek", "fd_tell":
		out, offset, whence := a[1], int64(0), io.SeekCurrent
		if name == "fd_seek" {
			out = a[3]
			offset = int64(a[1])
			if a[2] > 2 {
				return ErrInval
			}
			whence = int(a[2])
		}
		if err := c.check(out, 8); err != nil {
			return err
		}
		if r.dir != nil {
			return ErrIsdir
		}
		if r.file == nil {
			return ErrSpipe
		}
		n, err := r.file.Seek(offset, whence)
		if err != nil {
			return err
		}
		return c.u64(out, uint64(n))
	}
	return ErrNosys
}
func filetype(mode uint32) byte {
	switch mode & unix.S_IFMT {
	case unix.S_IFBLK:
		return 1
	case unix.S_IFCHR:
		return 2
	case unix.S_IFDIR:
		return 3
	case unix.S_IFREG:
		return 4
	case unix.S_IFSOCK:
		return 6
	case unix.S_IFLNK:
		return 7
	}
	return 0
}
func statBytes(st host.Stat) []byte {
	b := make([]byte, 64)
	binary.LittleEndian.PutUint64(b, st.Dev)
	binary.LittleEndian.PutUint64(b[8:], st.Inode)
	b[16] = filetype(st.Mode)
	for i, v := range []uint64{st.Nlink, st.Size, st.Atime, st.Mtime, st.Ctime} {
		binary.LittleEndian.PutUint64(b[24+i*8:], v)
	}
	return b
}
func validTimes(at, mt, flags uint64) error {
	if flags&^uint64(15) != 0 || flags&3 == 3 || flags&12 == 12 {
		return ErrInval
	}
	return nil
}
