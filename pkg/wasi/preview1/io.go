//go:build linux || darwin

package preview1

import (
	"context"
	"encoding/binary"
	"golang.org/x/sys/unix"
	"io"
	"math"

	"github.com/cpunion/wasmvm/internal/wasi/host"
)

func (p *Process) fdIO(c *abiCall, name string, a []uint64) error {
	read := name == "fd_read" || name == "fd_pread"
	positioned := name == "fd_pread" || name == "fd_pwrite"
	rights := RightFdWrite
	if read {
		rights = RightFdRead
	}
	if positioned {
		rights |= RightFdSeek
	}
	d, release, err := p.hold(uint32(a[0]), rights)
	if err != nil {
		return err
	}
	defer release()
	vs, total, err := c.vectors(a[1], a[2])
	if err != nil {
		return err
	}
	out := a[3]
	if positioned {
		out = a[4]
		if a[3] > math.MaxInt64 || total > math.MaxInt64-a[3] {
			return ErrOverflow
		}
	}
	if err := c.check(out, 4); err != nil {
		return err
	}
	data, err := c.buffer(total)
	if err != nil {
		return err
	}
	if !read {
		at := 0
		for _, v := range vs {
			if err := c.caller.ReadMemory(0, v.at, data[at:at+int(v.n)]); err != nil {
				return ErrFault
			}
			at += int(v.n)
		}
	}
	r := d.resource
	r.mu.Lock()
	flags := r.flags
	r.mu.Unlock()
	if read {
		r.readMu.Lock()
		defer r.readMu.Unlock()
	} else {
		r.writeMu.Lock()
		defer r.writeMu.Unlock()
	}
	if r.dir != nil {
		return ErrIsdir
	}
	ctx, cancel := context.WithCancel(c.ctx)
	stop := context.AfterFunc(r.ctx, cancel)
	defer func() { stop(); cancel() }()
	if read && flags&8 != 0 && r.file != nil {
		if err := r.file.Sync(); err != nil {
			return err
		}
	}
	n := 0
	if len(data) != 0 {
		switch {
		case positioned:
			if r.file == nil || r.socket {
				return ErrSpipe
			}
			if read {
				n, err = unix.Pread(host.Descriptor(r.file), data, int64(a[3]))
			} else {
				n, err = unix.Pwrite(host.Descriptor(r.file), data, int64(a[3]))
			}
		case r.stream != nil:
			if read {
				n, err = r.stream.Read(ctx, data)
			} else {
				n, err = r.stream.Write(ctx, data)
			}
		case r.socket:
			n, _, err = socketIO(ctx, r, data, read, 0)
		default:
			n, err = fileIO(ctx, r.file, data, read, flags&4 != 0)
		}
	}
	if n < 0 {
		if err == nil {
			return ErrIo
		}
		n = 0
	}
	if n > len(data) {
		return ErrIo
	}
	if n == 0 && err != nil && err != io.EOF {
		return err
	}
	if read {
		at := 0
		for _, v := range vs {
			count := int(v.n)
			if count > n-at {
				count = n - at
			}
			if count <= 0 {
				break
			}
			if err := c.write(v.at, data[at:at+count]); err != nil {
				return err
			}
			at += count
		}
	} else if n > 0 && flags&(2|16) != 0 && r.file != nil {
		if err := r.file.Sync(); err != nil {
			return err
		}
	}
	return c.u32(out, uint32(n))
}

func (p *Process) readDirectory(c *abiCall, a []uint64) error {
	d, release, err := p.hold(uint32(a[0]), RightFdReaddir)
	if err != nil {
		return err
	}
	defer release()
	r := d.resource
	if r.dir == nil {
		return ErrNotdir
	}
	if err := c.check(a[1], a[2]); err != nil {
		return err
	}
	if err := c.check(a[4], 4); err != nil {
		return err
	}
	buf, err := c.buffer(a[2])
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if a[3] == 0 || r.entries == nil {
		r.entries, err = host.ReadDir(c.ctx, r.dir, p.config.Limits.MaxDirectoryEntries)
		if err != nil {
			return err
		}
	}
	used := 0
	if a[3] > uint64(len(r.entries)) {
		return ErrInval
	}
	for i := int(a[3]); i < len(r.entries) && used < len(buf); i++ {
		entry := r.entries[i]
		var header [24]byte
		binary.LittleEndian.PutUint64(header[:], uint64(i+1))
		binary.LittleEndian.PutUint64(header[8:], entry.Inode)
		binary.LittleEndian.PutUint32(header[16:], uint32(len(entry.Name)))
		header[20] = filetype(entry.Mode)
		used += copy(buf[used:], header[:])
		used += copy(buf[used:], entry.Name)
	}
	if err := c.write(a[1], buf[:used]); err != nil {
		return err
	}
	return c.u32(a[4], uint32(used))
}
