//go:build linux || darwin

package preview1

import (
	"context"
	"os"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
)

// ImportSocket transfers ownership of an already-authorized socket, including
// on failure. The process never creates a network capability implicitly.
func (p *Process) ImportSocket(file *os.File, listening bool) (uint32, error) {
	if file == nil {
		return 0, ErrBadf
	}
	ctx, finish, err := p.begin(context.Background())
	if err != nil {
		file.Close()
		return 0, err
	}
	defer finish()
	if err := ctx.Err(); err != nil {
		file.Close()
		return 0, err
	}
	if err = p.reserveDescriptor(); err != nil {
		file.Close()
		return 0, err
	}
	r, err := p.socketResource(file, listening)
	if err != nil {
		p.unreserveDescriptor()
		file.Close()
		return 0, err
	}
	fd, err := p.publishDescriptor(socketDescriptor(r))
	if err != nil {
		r.close()
	}
	return fd, err
}
func (p *Process) socketResource(file *os.File, listening bool) (*resource, error) {
	fd := host.Descriptor(file)
	kind, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil {
		return nil, err
	}
	if kind != unix.SOCK_STREAM && kind != unix.SOCK_DGRAM {
		return nil, ErrNotsup
	}
	if listening && kind != unix.SOCK_STREAM {
		return nil, ErrInval
	}
	if err := unix.SetNonblock(fd, true); err != nil {
		return nil, err
	}
	r := p.newResource(file, nil, nil)
	r.socket = true
	r.datagram = kind == unix.SOCK_DGRAM
	r.listener = listening
	return r, nil
}
func socketDescriptor(r *resource) *descriptor {
	rights := RightFdRead | RightFdWrite | RightFdFdstatSetFlags | RightFdFilestatGet | RightPollFdReadwrite | RightSockShutdown
	if r.listener {
		rights = RightSockAccept | RightFdFdstatSetFlags | RightFdFilestatGet | RightPollFdReadwrite
	}
	return &descriptor{resource: r, base: rights}
}
func (p *Process) socket(c *abiCall, name string, a []uint64) error {
	rights := RightFdWrite
	if name == "sock_recv" {
		rights = RightFdRead
	}
	if name == "sock_accept" {
		rights = RightSockAccept
	}
	if name == "sock_shutdown" {
		rights = RightSockShutdown
	}
	// Validate descriptor identity before rights so non-sockets report NOTSOCK.
	d, release, err := p.hold(uint32(a[0]), 0)
	if err != nil {
		return err
	}
	defer release()
	r := d.resource
	if !r.socket {
		return ErrNotsock
	}
	if rights&^d.base != 0 {
		return ErrNotcapable
	}
	ctx, cancel := context.WithCancel(c.ctx)
	stop := context.AfterFunc(r.ctx, cancel)
	defer func() { stop(); cancel() }()
	if name == "sock_shutdown" {
		if a[1] == 0 || a[1] > 3 {
			return ErrInval
		}
		how := unix.SHUT_RDWR
		if a[1] == 1 {
			how = unix.SHUT_RD
		} else if a[1] == 2 {
			how = unix.SHUT_WR
		}
		return unix.Shutdown(host.Descriptor(r.file), how)
	}
	if name == "sock_accept" {
		if !r.listener {
			return ErrInval
		}
		if a[1]&^uint64(4) != 0 {
			return ErrInval
		}
		if err := c.check(a[2], 4); err != nil {
			return err
		}
		if err := p.reserveDescriptor(); err != nil {
			return err
		}
		reserved := true
		defer func() {
			if reserved {
				p.unreserveDescriptor()
			}
		}()
		r.readMu.Lock()
		defer r.readMu.Unlock()
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			fd, _, err := unix.Accept(host.Descriptor(r.file))
			if err == unix.EINTR {
				continue
			}
			if err == unix.EAGAIN {
				r.mu.Lock()
				nonblock := r.flags&4 != 0
				r.mu.Unlock()
				if nonblock {
					return ErrAgain
				}
				if _, _, err = host.Poll(ctx, host.Descriptor(r.file), true); err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			unix.CloseOnExec(fd)
			f := os.NewFile(uintptr(fd), "WASI accepted socket")
			child, err := p.socketResource(f, false)
			if err != nil {
				f.Close()
				return err
			}
			child.flags = uint16(a[1])
			number, err := p.publishDescriptor(socketDescriptor(child))
			reserved = false
			if err != nil {
				child.close()
				return err
			}
			if err := c.u32(a[2], number); err != nil {
				p.closeFD(number)
				return err
			}
			return nil
		}
	}
	read := name == "sock_recv"
	if read && a[3] > 3 || !read && a[3] != 0 {
		return ErrInval
	}
	vs, total, err := c.vectors(a[1], a[2])
	if err != nil {
		return err
	}
	if err := c.check(a[4], 4); err != nil {
		return err
	}
	if read {
		if err := c.check(a[5], 2); err != nil {
			return err
		}
	}
	b, err := c.buffer(total)
	if err != nil {
		return err
	}
	if !read {
		at := 0
		for _, v := range vs {
			if err := c.caller.ReadMemory(0, v.at, b[at:at+int(v.n)]); err != nil {
				return ErrFault
			}
			at += int(v.n)
		}
	}
	if read {
		r.readMu.Lock()
		defer r.readMu.Unlock()
	} else {
		r.writeMu.Lock()
		defer r.writeMu.Unlock()
	}
	n, flags, err := socketIO(ctx, r, b, read, uint32(a[3]))
	if n == 0 && err != nil {
		return err
	}
	if n < 0 || n > len(b) {
		return ErrIo
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
			if err := c.write(v.at, b[at:at+count]); err != nil {
				return err
			}
			at += count
		}
		var out [2]byte
		if flags&unix.MSG_TRUNC != 0 {
			out[0] = 1
		}
		if err := c.write(a[5], out[:]); err != nil {
			return err
		}
	}
	return c.u32(a[4], uint32(n))
}
func socketIO(ctx context.Context, r *resource, b []byte, read bool, flags uint32) (int, int, error) {
	if read && len(b) == 0 {
		return 0, 0, nil
	}
	fd := host.Descriptor(r.file)
	r.mu.Lock()
	nonblock := r.flags&4 != 0
	r.mu.Unlock()
	pos := 0
	for {
		if err := ctx.Err(); err != nil {
			return pos, 0, err
		}
		var n, out int
		var err error
		if read {
			in := 0
			if flags&1 != 0 {
				in |= unix.MSG_PEEK
			}
			n, _, out, _, err = unix.Recvmsg(fd, b[pos:], nil, in)
		} else {
			n, err = unix.SendmsgN(fd, b, nil, nil, unix.MSG_NOSIGNAL)
		}
		if err == unix.EINTR {
			continue
		}
		if err == unix.EAGAIN {
			if nonblock {
				return pos, 0, err
			}
			if _, _, err = host.Poll(ctx, fd, read); err != nil {
				return pos, 0, err
			}
			continue
		}
		if err != nil {
			return pos, out, err
		}
		if n < 0 {
			return pos, out, ErrIo
		}
		if n > len(b)-pos {
			n = len(b) - pos
		}
		if !read || r.datagram || flags&2 == 0 || n == 0 {
			return pos + n, out, nil
		}
		if flags&1 != 0 {
			if n == len(b) {
				return n, out, nil
			}
			// PEEK cannot consume the available prefix while WAITALL waits for
			// more. A bounded cancellable delay avoids spinning on ready bytes.
			t := time.NewTimer(time.Millisecond)
			select {
			case <-ctx.Done():
				t.Stop()
				return n, out, ctx.Err()
			case <-t.C:
			}
			continue
		}
		pos += n
		if pos == len(b) {
			return pos, out, nil
		}
	}
}
