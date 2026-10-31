//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type fileStream struct {
	mu, ioMu                                 sync.Mutex
	owner                                    *FileSystem
	descriptor                               fileDescriptor
	op                                       *ioOperation
	ctx                                      context.Context
	cancel                                   context.CancelFunc
	changed                                  chan struct{}
	closed, watching, input, append, nonseek bool
	offset                                   uint64
	failure                                  error
	cursor                                   *fileObject
	seen                                     uint32
}

func (p *FileSystem) stream(c *component.Caller, name string, args []component.Value) (*component.Value, error) {
	d, err := p.descriptor(c, args[0].Resource)
	if err != nil {
		return nil, err
	}
	op, err := p.io.operationFor(fileCompletionCode, fileCompletionValue, fileErrorCases["bad-descriptor"]+1)
	if err != nil {
		d.object.release()
		return nil, err
	}
	ctx, cancel := context.WithCancel(p.ctx)
	s := &fileStream{owner: p, descriptor: d, op: op, ctx: ctx, cancel: cancel, changed: make(chan struct{}), input: name == "read-via-stream" || name == "read-directory", append: name == "append-via-stream"}
	if name == "read-via-stream" {
		s.offset = args[1].Bits
		if d.flags&fsRead == 0 {
			s.failure = fileError("not-permitted")
		}
	}
	if name == "write-via-stream" {
		s.offset = args[2].Bits
	}
	if name == "read-directory" {
		_, s.failure = d.directory(false)
	} else if !s.input {
		s.failure = d.writable()
	}
	if s.offset > 1<<63-1 {
		s.failure = fileError("overflow")
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		s.Close()
		op.finish(fileErrorCases["bad-descriptor"] + 1)
		return nil, os.ErrClosed
	}
	if p.streams == nil {
		p.streams = map[*fileStream]struct{}{}
	}
	p.streams[s] = struct{}{}
	p.mu.Unlock()
	op.setRelease(func() { s.Close() })
	if name == "read-directory" {
		_, streamType := c.FunctionInfo().Result().Field(0)
		_, futureType := c.FunctionInfo().Result().Field(1)
		accepted := false
		defer func() {
			if !accepted {
				op.stop()
				op.finish(fileErrorCases["io"] + 1)
			}
		}()
		stream, err := c.NewSource(streamType, component.AsyncSource{Read: s.readDirectory, Close: func() { op.stop(); op.finish(0) }})
		if err != nil {
			return nil, err
		}
		future, err := c.NewSource(futureType, op.cell.source(p.io, nil))
		if err != nil {
			return nil, err
		}
		accepted = true
		return &component.Value{Elements: []component.Value{stream, future}}, nil
	}
	if s.input {
		_, streamType := c.FunctionInfo().Result().Field(0)
		_, futureType := c.FunctionInfo().Result().Field(1)
		return p.io.inputWithOperation(c, op, s, streamType, futureType)
	}
	return p.io.outputWithOperation(c, []component.Value{args[1]}, s, op)
}
func (s *fileStream) notifyLocked() { close(s.changed); s.changed = make(chan struct{}) }
func (s *fileStream) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.cancel()
	cursor := s.cursor
	s.cursor = nil
	s.notifyLocked()
	s.mu.Unlock()
	s.owner.mu.Lock()
	delete(s.owner.streams, s)
	s.owner.mu.Unlock()
	if cursor != nil {
		cursor.release()
	}
	s.descriptor.object.release()
	return nil
}
func (s *fileStream) revoke() { s.op.stop(); s.op.finish(fileErrorCases["bad-descriptor"] + 1) }
func (s *fileStream) check(ctx context.Context) error {
	if err := s.owner.check(ctx); err != nil {
		return err
	}
	if s.ctx.Err() != nil {
		return os.ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.failure
}
func (s *fileStream) begin(ctx context.Context) error {
	if err := s.check(ctx); err != nil {
		return err
	}
	if err := s.owner.begin(); err != nil {
		return err
	}
	if err := s.descriptor.object.acquire(); err != nil {
		s.owner.end()
		return err
	}
	return nil
}
func (s *fileStream) end() { s.descriptor.object.release(); s.owner.end() }
func (s *fileStream) Read(ctx context.Context, data []byte) (int, error) {
	if err := s.begin(ctx); err != nil {
		return 0, err
	}
	defer s.end()
	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	if uint64(len(data)) > uint64(1<<63-1)-s.offset {
		return 0, fileError("overflow")
	}
	var n int
	var err error
	if !s.nonseek {
		n, err = s.descriptor.object.file.ReadAt(data, int64(s.offset))
		if errors.Is(err, unix.ESPIPE) && s.offset == 0 {
			s.nonseek = true
		}
	}
	if s.nonseek {
		n, err = unix.Read(host.Descriptor(s.descriptor.object.file), data)
		if n < 0 {
			n = 0
		}
		if n == 0 && err == nil && len(data) != 0 {
			err = io.EOF
		}
	} else {
		s.offset += uint64(n)
	}
	if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
		err = ErrWouldBlock
	}
	return n, err
}
func (s *fileStream) Write(ctx context.Context, data []byte) (n int, err error) {
	if err = s.begin(ctx); err != nil {
		return 0, err
	}
	defer s.end()
	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	b := s.descriptor.object
	b.ioMu.Lock()
	defer b.ioMu.Unlock()
	if b.writeFailure != nil {
		return 0, b.writeFailure
	}
	if s.append {
		fd := host.Descriptor(b.file)
		flags, failed := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
		if failed != nil {
			return 0, failed
		}
		if _, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags|unix.O_APPEND|unix.O_NONBLOCK); err != nil {
			return 0, err
		}
		n, err = unix.Write(fd, data)
		if _, restore := unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags); restore != nil {
			b.writeFailure = restore
			if err == nil {
				err = restore
			}
		}
	} else {
		if uint64(len(data)) > uint64(1<<63-1)-s.offset {
			return 0, fileError("overflow")
		}
		if !s.nonseek {
			n, err = b.file.WriteAt(data, int64(s.offset))
			if errors.Is(err, unix.ESPIPE) && s.offset == 0 {
				s.nonseek = true
			}
		}
		if s.nonseek {
			n, err = unix.Write(host.Descriptor(b.file), data)
		} else {
			s.offset += uint64(n)
		}
	}
	if n < 0 {
		n = 0
	}
	if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
		err = ErrWouldBlock
	}
	return n, err
}
func (s *fileStream) Flush(ctx context.Context) error { return s.check(ctx) }
func (s *fileStream) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.failure != nil {
		return true, s.changed
	}
	if s.watching {
		return false, s.changed
	}
	if err := s.owner.begin(); err != nil {
		s.failure = err
		return true, s.changed
	}
	b := s.descriptor.object
	if err := b.acquire(); err != nil {
		s.owner.end()
		s.failure = err
		return true, s.changed
	}
	_, _, ready, err := host.Ready(host.Descriptor(b.file), s.input)
	if err != nil || ready {
		b.release()
		s.owner.end()
		if err != nil {
			s.failure = err
		}
		return true, s.changed
	}
	s.watching = true
	go func() {
		defer s.owner.workers.Done()
		_, _, err := host.Poll(s.ctx, host.Descriptor(b.file), s.input)
		b.release()
		s.owner.releaseSlot()
		s.mu.Lock()
		s.watching = false
		if err != nil && !s.closed {
			s.failure = err
		}
		s.notifyLocked()
		s.mu.Unlock()
	}()
	return false, s.changed
}
func (s *fileStream) directoryCursor(ctx context.Context) (*fileObject, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, os.ErrClosed
	}
	if s.cursor != nil {
		b := s.cursor
		err := b.acquire()
		s.mu.Unlock()
		return b, err
	}
	s.mu.Unlock()
	directory, err := s.descriptor.directory(false)
	if err != nil {
		return nil, err
	}
	if err := s.owner.reserveObject(ctx); err != nil {
		return nil, err
	}
	file, child, err := directory.Open(".", unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		s.owner.cancelObjectReservation()
		return nil, err
	}
	b, err := s.owner.registerObject(file, child)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		b.release()
		return nil, os.ErrClosed
	}
	s.cursor = b
	err = b.acquire()
	s.mu.Unlock()
	return b, err
}

func directoryType(mode os.FileMode) component.Value {
	value := component.Value{Case: 7, Elements: []component.Value{{}}}
	switch {
	case mode.IsRegular():
		return component.Value{Case: 5}
	case mode.IsDir():
		return component.Value{Case: 2}
	case mode&os.ModeSymlink != 0:
		return component.Value{Case: 4}
	case mode&os.ModeNamedPipe != 0:
		return component.Value{Case: 3}
	case mode&os.ModeSocket != 0:
		return component.Value{Case: 6}
	case mode&os.ModeCharDevice != 0:
		return component.Value{Case: 1}
	case mode&os.ModeDevice != 0:
		return component.Value{Case: 0}
	}
	return value
}
func (s *fileStream) readDirectory(c *component.Caller, limit uint32) (*component.ReadCompletion, error) {
	if err := s.owner.begin(); err != nil {
		return nil, err
	}
	if err := s.descriptor.object.acquire(); err != nil {
		s.owner.end()
		return nil, err
	}
	completion := component.NewReadCompletion()
	callCtx := c.Context()
	ctx, cancel := context.WithCancel(callCtx)
	stop := context.AfterFunc(s.ctx, cancel)
	go func() {
		defer s.owner.workers.Done()
		defer cancel()
		defer stop()
		var values []component.Value
		var err error
		defer func() {
			panicked := recover()
			s.descriptor.object.release()
			s.owner.releaseSlot()
			if callCtx.Err() != nil && len(values) == 0 {
				completion.Cancelled()
				return
			}
			closed := err != nil
			if closed {
				s.op.finish(fileCompletionCode(err))
			}
			result := component.ReadResult{Values: values, Closed: closed}
			completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
				if panicked != nil {
					panic(panicked)
				}
				return result, nil
			})
		}()
		s.ioMu.Lock()
		defer s.ioMu.Unlock()
		if err = s.check(ctx); err != nil {
			return
		}
		if limit == 0 {
			return
		}
		var cursor *fileObject
		cursor, err = s.directoryCursor(ctx)
		if err != nil {
			return
		}
		defer cursor.release()
		count := min(limit, s.owner.config.MaxDirectoryBatch, s.owner.config.MaxDirectoryEntries-s.seen+1)
		var entries []os.DirEntry
		entries, err = cursor.file.ReadDir(int(count))
		if uint64(len(entries)) > uint64(s.owner.config.MaxDirectoryEntries-s.seen) {
			err = wasm.ErrLimit
			return
		}
		for _, entry := range entries {
			if !utf8.ValidString(entry.Name()) {
				err = fileError("illegal-byte-sequence")
				return
			}
			if e := s.owner.path(entry.Name()); e != nil {
				err = e
				return
			}
			values = append(values, component.Value{Elements: []component.Value{directoryType(entry.Type()), {Text: entry.Name()}}})
		}
		s.seen += uint32(len(entries))
	}()
	return completion, nil
}
