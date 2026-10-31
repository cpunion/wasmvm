//go:build linux || darwin

package preview3

import (
	"context"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const FileSystemTypesInterface = "wasi:filesystem/types@0.3.2"
const FileSystemPreopensInterface = "wasi:filesystem/preopens@0.3.2"

type FileSystemPreopen struct {
	GuestPath, HostPath string
	ReadOnly            bool
}
type FileSystemConfig struct {
	Preopens                                                                         []FileSystemPreopen
	MaxDescriptors, MaxDirectoryEntries, MaxDirectoryBatch, MaxPathBytes, MaxWorkers uint32
}

const (
	fsRead uint64 = 1 << iota
	fsWrite
	fsFileSync
	fsDataSync
	fsReadSync
	fsMutate
)

// FileSystem grants explicit capability-relative roots. Descriptors and workers
// are bounded across Stores; asynchronous workers retain object leases, never
// Callers. Its IO owner supplies the shared native-stream operation domain.
type FileSystem struct {
	mu                 sync.Mutex
	config             FileSystemConfig
	io                 *IO
	ctx                context.Context
	cancel             context.CancelFunc
	done               chan struct{}
	closed             bool
	workers            sync.WaitGroup
	active             uint32
	next               uint64
	resource           *component.HostResourceType
	descriptors        map[uint32]fileDescriptor
	objects            map[*fileObject]struct{}
	streams            map[*fileStream]struct{}
	opening            uint32
	closing            uint32
	reapHead, reapTail *fileObject
	reapWake, reapDone chan struct{}
	reapStopping       bool
	objectChanged      chan struct{}
	preopens           []filePreopen
}
type fileDescriptor struct {
	object *fileObject
	flags  uint64
}
type filePreopen struct {
	name       string
	descriptor fileDescriptor
}
type fileObject struct {
	mu, ioMu     sync.Mutex
	owner        *FileSystem
	file         *os.File
	directory    *host.Directory
	refs         uint32
	closeNext    *fileObject
	writeFailure error
	closed       bool
}

func NewFileSystem(ioOwner *IO, config FileSystemConfig) (*FileSystem, error) {
	if ioOwner == nil || ioOwner.inputType == nil {
		return nil, vm.ErrReference
	}
	ioOwner.mu.Lock()
	closed := ioOwner.closed
	ioOwner.mu.Unlock()
	if closed {
		return nil, wasm.ErrClosed
	}
	if config.MaxDescriptors == 0 {
		config.MaxDescriptors = 1024
	}
	if config.MaxDirectoryEntries == 0 {
		config.MaxDirectoryEntries = 65536
	}
	if config.MaxDirectoryBatch == 0 {
		config.MaxDirectoryBatch = 256
	}
	if config.MaxPathBytes == 0 {
		config.MaxPathBytes = 4096
	}
	if config.MaxWorkers == 0 {
		config.MaxWorkers = 1024
	}
	if uint64(len(config.Preopens)) > uint64(config.MaxDescriptors) || config.MaxDirectoryEntries >= 1<<24 || config.MaxDirectoryBatch >= 1<<24 {
		return nil, wasm.ErrLimit
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &FileSystem{config: config, io: ioOwner, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, descriptors: map[uint32]fileDescriptor{}, objects: map[*fileObject]struct{}{}}
	p.reapWake = make(chan struct{}, 1)
	p.reapDone = make(chan struct{})
	p.objectChanged = make(chan struct{})
	go p.reap()
	p.resource = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { p.drop(id); return nil })
	for _, entry := range config.Preopens {
		if !utf8.ValidString(entry.GuestPath) || p.path(entry.GuestPath) != nil {
			p.Close()
			p.WaitClosed(nil)
			return nil, wasm.ErrInvalid
		}
		dir, err := host.OpenDirectory(entry.HostPath)
		if err != nil {
			p.Close()
			p.WaitClosed(nil)
			return nil, err
		}
		object := &fileObject{owner: p, file: dir.File, directory: dir, refs: 1}
		p.objects[object] = struct{}{}
		flags := fsRead
		if !entry.ReadOnly {
			flags |= fsMutate
		}
		p.preopens = append(p.preopens, filePreopen{entry.GuestPath, fileDescriptor{object, flags}})
	}
	p.config.Preopens = nil
	return p, nil
}
func (p *FileSystem) Close() {
	if p == nil || p.done == nil {
		return
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.cancel()
	p.descriptors = nil
	p.preopens = nil
	streams := p.streams
	p.streams = nil
	p.mu.Unlock()
	for stream := range streams {
		stream.revoke()
	}
	go func() {
		p.workers.Wait()
		p.mu.Lock()
		objects := make([]*fileObject, 0, len(p.objects))
		for b := range p.objects {
			objects = append(objects, b)
		}
		p.mu.Unlock()
		for _, b := range objects {
			b.close()
		}
		p.mu.Lock()
		p.reapStopping = true
		p.mu.Unlock()
		p.wakeReaper()
		<-p.reapDone
		p.mu.Lock()
		p.objects = nil
		p.mu.Unlock()
		close(p.done)
	}()
}

func (p *FileSystem) WaitClosed(ctx context.Context) error {
	if p == nil || p.done == nil {
		return vm.ErrReference
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (p *FileSystem) begin() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return os.ErrClosed
	}
	if p.active >= p.config.MaxWorkers {
		return wasm.ErrLimit
	}
	p.active++
	p.workers.Add(1)
	return nil
}
func (p *FileSystem) releaseSlot() { p.mu.Lock(); p.active--; p.mu.Unlock() }
func (p *FileSystem) end()         { p.releaseSlot(); p.workers.Done() }

func (p *FileSystem) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-p.ctx.Done():
		return os.ErrClosed
	default:
		return nil
	}
}
func (p *FileSystem) path(name string) error {
	if uint64(len(name)) > uint64(p.config.MaxPathBytes) {
		return fileError("name-too-long")
	}
	if strings.IndexByte(name, 0) >= 0 {
		return fileError("invalid")
	}
	return nil
}
func (b *fileObject) acquire() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return os.ErrClosed
	}
	if b.refs == 1<<32-1 {
		return wasm.ErrLimit
	}
	b.refs++
	return nil
}
func (b *fileObject) release() {
	b.mu.Lock()
	b.refs--
	last := b.refs == 0
	b.mu.Unlock()
	if last {
		b.close()
	}
}
func (b *fileObject) close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	b.mu.Unlock()
	p := b.owner
	p.mu.Lock()
	p.closing++
	if p.reapTail != nil {
		p.reapTail.closeNext = b
	} else {
		p.reapHead = b
	}
	p.reapTail = b
	p.mu.Unlock()
	p.wakeReaper()
}
func (p *FileSystem) wakeReaper() {
	select {
	case p.reapWake <- struct{}{}:
	default:
	}
}
func (p *FileSystem) reap() {
	defer close(p.reapDone)
	for {
		p.mu.Lock()
		b := p.reapHead
		stop := p.reapStopping
		if b != nil {
			p.reapHead = b.closeNext
			b.closeNext = nil
			if p.reapHead == nil {
				p.reapTail = nil
			}
		}
		p.mu.Unlock()
		if b == nil {
			if stop {
				return
			}
			<-p.reapWake
			continue
		}
		if b.directory != nil {
			b.directory.Close()
		} else {
			b.file.Close()
		}
		p.mu.Lock()
		delete(p.objects, b)
		p.closing--
		close(p.objectChanged)
		p.objectChanged = make(chan struct{})
		p.mu.Unlock()
	}
}
func (p *FileSystem) reserveObject(ctx context.Context) error {
	for {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return os.ErrClosed
		}
		if uint64(len(p.objects))+uint64(p.opening) < uint64(p.config.MaxDescriptors) {
			p.opening++
			p.mu.Unlock()
			return nil
		}
		if p.closing == 0 {
			p.mu.Unlock()
			return wasm.ErrLimit
		}
		changed := p.objectChanged
		p.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.ctx.Done():
			return os.ErrClosed
		case <-changed:
		}
	}
}
func (p *FileSystem) cancelObjectReservation() { p.mu.Lock(); p.opening--; p.mu.Unlock() }
func (p *FileSystem) registerObject(file *os.File, dir *host.Directory) (*fileObject, error) {
	b := &fileObject{owner: p, file: file, directory: dir, refs: 1}
	p.mu.Lock()
	p.opening--
	p.objects[b] = struct{}{}
	closed := p.closed
	p.mu.Unlock()
	if closed {
		b.close()
		return nil, os.ErrClosed
	}
	return b, nil
}

func (p *FileSystem) drop(id uint32) {
	p.mu.Lock()
	d, ok := p.descriptors[id]
	delete(p.descriptors, id)
	p.mu.Unlock()
	if ok {
		d.object.release()
	}
}
func (p *FileSystem) descriptor(c *component.Caller, ref *component.Resource) (fileDescriptor, error) {
	id, err := c.ResourceRep(p.resource, ref)
	if err != nil {
		return fileDescriptor{}, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return fileDescriptor{}, os.ErrClosed
	}
	d, ok := p.descriptors[id]
	if !ok {
		return fileDescriptor{}, vm.ErrReference
	}
	if err := d.object.acquire(); err != nil {
		return fileDescriptor{}, err
	}
	return d, nil
}
func (p *FileSystem) publish(c *component.Caller, d fileDescriptor) (component.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return component.Value{}, os.ErrClosed
	}
	if uint64(len(p.descriptors)) >= uint64(p.config.MaxDescriptors) || p.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	if err := d.object.acquire(); err != nil {
		return component.Value{}, err
	}
	id := uint32(p.next)
	p.next++
	p.descriptors[id] = d
	v, err := c.NewResource(p.resource, id)
	if err != nil {
		delete(p.descriptors, id)
		d.object.mu.Lock()
		d.object.refs--
		d.object.mu.Unlock()
	}
	return v, err
}
func (p *FileSystem) getDirectories(c *component.Caller, _ []component.Value) (*component.Value, error) {
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	p.mu.Lock()
	roots := append([]filePreopen(nil), p.preopens...)
	p.mu.Unlock()
	out := &component.Value{}
	for _, root := range roots {
		v, err := p.publish(c, root.descriptor)
		if err != nil {
			return nil, err
		}
		out.Elements = append(out.Elements, component.Value{Elements: []component.Value{v, {Text: root.name}}})
	}
	return out, nil
}
func (d fileDescriptor) directory(mutate bool) (*host.Directory, error) {
	if d.object.directory == nil {
		return nil, fileError("not-directory")
	}
	if mutate && d.flags&fsMutate == 0 {
		return nil, fileError("read-only")
	}
	return d.object.directory, nil
}
func (d fileDescriptor) writable() error {
	if d.object.directory != nil {
		_, err := d.directory(true)
		return err
	}
	if d.flags&fsWrite == 0 {
		return fileError("not-permitted")
	}
	return nil
}
