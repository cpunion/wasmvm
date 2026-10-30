//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

// fsStream owns a backing lease independently of the originating descriptor.
// Its cursor is private. All shared-file flag changes and positional writes use
// the backing lock, including the brief O_APPEND scope of atomic append writes.
type fsStream struct {
	mu                                 sync.Mutex
	backing                            *fsBacking
	offset                             uint64
	input, appendMode, regular, closed bool
	changed                            chan struct{}
	ctx                                context.Context
	cancel                             context.CancelFunc
	watcher                            chan struct{}
	closeDone                          chan struct{}
	failure                            error
}

func (p *FileSystem) stream(c *component.Caller, d fsDescriptor, name string, offset uint64) (component.Value, error) {
	input := name == "read-via-stream"
	if d.backing.directory != nil {
		return component.Value{}, fsCode(FSIsDirectory)
	}
	if input && d.flags&fsRead == 0 {
		return component.Value{}, fsCode(FSNotPermitted)
	}
	if !input {
		if err := fsWritable(d); err != nil {
			return component.Value{}, err
		}
	}
	if offset > 1<<63-1 {
		return component.Value{}, fsCode(FSOverflow)
	}
	stat, err := host.Fstat(d.backing.file)
	if err != nil {
		return component.Value{}, err
	}
	if err = d.backing.acquire(); err != nil {
		return component.Value{}, err
	}
	ctx, cancel := context.WithCancel(p.ctx)
	source := &fsStream{backing: d.backing, offset: offset, input: input, appendMode: name == "append-via-stream", regular: fsType(stat.Mode) == 6, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
	var value component.Value
	if input {
		value, err = p.io.NewInput(c, source)
	} else {
		value, err = p.io.NewOutput(c, source)
	}
	// A closed IO owner rejects before accepting ownership; failed publication
	// closes an accepted source. Close is idempotent in both cases.
	if err != nil {
		source.Close()
	}
	return value, err
}
func fsStreamError(err error) error {
	if err == nil || errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, unix.EAGAIN) {
		return ErrWouldBlock
	}
	return &FileSystemError{Code: fsError(err), Cause: err}
}
func (s *fsStream) beginLocked(ctx context.Context) error {
	if s.closed {
		return io.ErrClosedPipe
	}
	if err := s.backing.owner.check(ctx); err != nil {
		return err
	}
	return s.backing.owner.begin()
}
func (s *fsStream) Read(ctx context.Context, data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.beginLocked(ctx); err != nil {
		return 0, fsStreamError(err)
	}
	defer s.backing.owner.active.Done()
	if !s.input {
		return 0, fsCode(FSNotPermitted)
	}
	if uint64(len(data)) > uint64(1<<63-1)-s.offset {
		return 0, fsCode(FSOverflow)
	}
	s.backing.ioMu.Lock()
	var n int
	var err error
	if s.regular {
		n, err = unix.Pread(host.Descriptor(s.backing.file), data, int64(s.offset))
	} else {
		n, err = unix.Read(host.Descriptor(s.backing.file), data)
	}
	s.backing.ioMu.Unlock()
	if n < 0 {
		n = 0
	}
	if n > 0 {
		s.offset += uint64(n)
	}
	if n == 0 && err == nil && len(data) > 0 {
		err = io.EOF
	}
	return n, fsStreamError(err)
}
func (s *fsStream) Write(ctx context.Context, data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.beginLocked(ctx); err != nil {
		return 0, fsStreamError(err)
	}
	defer s.backing.owner.active.Done()
	if s.input {
		return 0, fsCode(FSNotPermitted)
	}
	if !s.appendMode && uint64(len(data)) > uint64(1<<63-1)-s.offset {
		return 0, fsCode(FSOverflow)
	}
	s.backing.ioMu.Lock()
	defer s.backing.ioMu.Unlock()
	if err := s.backing.writeFailure; err != nil {
		return 0, fsStreamError(err)
	}
	fd := host.Descriptor(s.backing.file)
	var n int
	var err error
	switch {
	case s.appendMode && s.regular:
		var original int
		original, err = unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
		if err == nil {
			_, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, original|unix.O_APPEND)
		}
		if err == nil {
			n, err = unix.Write(fd, data)
			_, restore := unix.FcntlInt(uintptr(fd), unix.F_SETFL, original)
			if restore != nil {
				s.failure = fsStreamError(restore)
				s.backing.writeFailure = s.failure
				if err == nil {
					err = restore
				}
			}
		}
	case s.regular:
		n, err = unix.Pwrite(fd, data, int64(s.offset))
	default:
		n, err = unix.Write(fd, data)
	}
	if n < 0 {
		n = 0
	}
	if n > 0 && !s.appendMode {
		s.offset += uint64(n)
	}
	return n, fsStreamError(err)
}
func (s *fsStream) Flush(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.beginLocked(ctx); err != nil {
		return fsStreamError(err)
	}
	defer s.backing.owner.active.Done()
	return s.failure
}
func (s *fsStream) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.failure != nil {
		return true, s.changed
	}
	if err := s.beginLocked(context.Background()); err != nil {
		s.failure = fsStreamError(err)
		return true, s.changed
	}
	defer s.backing.owner.active.Done()
	if s.regular {
		return true, s.changed
	}
	_, _, ready, err := host.Ready(host.Descriptor(s.backing.file), s.input)
	if err != nil {
		s.failure = fsStreamError(err)
		return true, s.changed
	}
	if ready {
		return true, s.changed
	}
	if s.watcher == nil {
		if err = s.backing.owner.begin(); err != nil {
			s.failure = fsStreamError(err)
			return true, s.changed
		}
		done := make(chan struct{})
		s.watcher = done
		go func() {
			defer s.backing.owner.active.Done()
			_, _, err := host.Poll(s.ctx, host.Descriptor(s.backing.file), s.input)
			s.mu.Lock()
			if err != nil && !s.closed {
				s.failure = fsStreamError(err)
			}
			close(s.changed)
			s.changed = make(chan struct{})
			s.watcher = nil
			close(done)
			s.mu.Unlock()
		}()
	}
	return false, s.changed
}
func (s *fsStream) Close() error {
	s.mu.Lock()
	if s.closed {
		done := s.closeDone
		s.mu.Unlock()
		if done != nil {
			<-done
		}
		return nil
	}
	s.closed = true
	s.closeDone = make(chan struct{})
	doneClose := s.closeDone
	defer close(doneClose)
	s.cancel()
	close(s.changed)
	s.changed = make(chan struct{})
	done := s.watcher
	s.mu.Unlock()
	if done != nil {
		<-done
	}
	s.backing.release()
	return nil
}
