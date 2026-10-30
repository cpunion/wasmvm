//go:build linux || darwin

package preview2

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

const FileSystemTypesInterface = "wasi:filesystem/types@0.2.12"
const FileSystemPreopensInterface = "wasi:filesystem/preopens@0.2.12"

type FileSystemPreopen struct {
	GuestPath, HostPath string
	ReadOnly            bool
}
type FileSystemConfig struct {
	Preopens                                                                                            []FileSystemPreopen
	MaxDescriptors, MaxDirectoryStreams, MaxDirectoryEntries, MaxPathBytes, MaxReadBytes, MaxWriteBytes uint32
}

const (
	fsRead uint64 = 1 << iota
	fsWrite
	fsFileSync
	fsDataSync
	fsReadSync
	fsMutate
)

type FileSystem struct {
	mu                          sync.Mutex
	config                      FileSystemConfig
	io                          *IO
	ctx                         context.Context
	cancel                      context.CancelFunc
	done                        chan struct{}
	closed                      bool
	active                      sync.WaitGroup
	next                        uint64
	descriptorType, entriesType *component.HostResourceType
	descriptors                 map[uint32]fsDescriptor
	entries                     map[uint32]*fsEntries
	backings                    map[*fsBacking]struct{}
	preopens                    []fsPreopen
}
type fsDescriptor struct {
	backing *fsBacking
	flags   uint64
}
type fsPreopen struct {
	name       string
	descriptor fsDescriptor
}
type fsEntries struct {
	mu      sync.Mutex
	entries []host.Dirent
	at      int
}
type fsBacking struct {
	mu, ioMu     sync.Mutex
	owner        *FileSystem
	file         *os.File
	directory    *host.Directory
	refs         uint32
	writeFailure error
	closed       bool
}

func NewFileSystem(ioOwner *IO, config FileSystemConfig) (*FileSystem, error) {
	if ioOwner == nil || ioOwner.inputType == nil {
		return nil, vm.ErrReference
	}
	if err := ioOwner.check(context.Background()); err != nil {
		return nil, err
	}
	if config.MaxDescriptors == 0 {
		config.MaxDescriptors = 1024
	}
	if config.MaxDirectoryStreams == 0 {
		config.MaxDirectoryStreams = 128
	}
	if config.MaxDirectoryEntries == 0 {
		config.MaxDirectoryEntries = 65536
	}
	if config.MaxPathBytes == 0 {
		config.MaxPathBytes = 4096
	}
	if config.MaxReadBytes == 0 {
		config.MaxReadBytes = 64 << 10
	}
	if config.MaxWriteBytes == 0 {
		config.MaxWriteBytes = 64 << 10
	}
	if uint64(len(config.Preopens)) > uint64(config.MaxDescriptors) || config.MaxReadBytes >= 1<<28 || config.MaxWriteBytes >= 1<<28 || config.MaxDirectoryEntries > 1<<24 {
		return nil, wasm.ErrLimit
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &FileSystem{config: config, io: ioOwner, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, descriptors: map[uint32]fsDescriptor{}, entries: map[uint32]*fsEntries{}, backings: map[*fsBacking]struct{}{}}
	p.config.Preopens = append([]FileSystemPreopen(nil), config.Preopens...)
	p.descriptorType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { p.dropDescriptor(id); return nil })
	p.entriesType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error {
		p.mu.Lock()
		delete(p.entries, id)
		p.mu.Unlock()
		return nil
	})
	for _, entry := range config.Preopens {
		if !utf8.ValidString(entry.GuestPath) || strings.IndexByte(entry.GuestPath, 0) >= 0 || uint64(len(entry.GuestPath)) > uint64(config.MaxPathBytes) {
			p.Close()
			p.WaitClosed(nil)
			return nil, wasm.ErrInvalid
		}
		directory, err := host.OpenDirectory(entry.HostPath)
		if err != nil {
			p.Close()
			p.WaitClosed(nil)
			return nil, err
		}
		b := &fsBacking{owner: p, file: directory.File, directory: directory, refs: 1}
		p.backings[b] = struct{}{}
		flags := fsRead
		if !entry.ReadOnly {
			flags |= fsMutate
		}
		p.preopens = append(p.preopens, fsPreopen{entry.GuestPath, fsDescriptor{b, flags}})
	}
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
	backings := p.backings
	p.backings = nil
	p.descriptors = nil
	p.entries = nil
	p.preopens = nil
	p.mu.Unlock()
	go func() {
		p.active.Wait()
		for b := range backings {
			b.close()
		}
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
	p.active.Add(1)
	return nil
}
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
func (b *fsBacking) acquire() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return os.ErrClosed
	}
	b.refs++
	return nil
}
func (b *fsBacking) release() {
	b.mu.Lock()
	b.refs--
	last := b.refs == 0
	b.mu.Unlock()
	if last {
		b.close()
		b.owner.mu.Lock()
		delete(b.owner.backings, b)
		b.owner.mu.Unlock()
	}
}
func (b *fsBacking) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	if b.directory != nil {
		b.directory.Close()
	} else {
		b.file.Close()
	}
}
func (p *FileSystem) dropDescriptor(id uint32) {
	p.mu.Lock()
	descriptor, ok := p.descriptors[id]
	delete(p.descriptors, id)
	p.mu.Unlock()
	if ok {
		descriptor.backing.release()
	}
}
func (p *FileSystem) publish(c *component.Caller, descriptor fsDescriptor) (component.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return component.Value{}, os.ErrClosed
	}
	if uint64(len(p.descriptors)) >= uint64(p.config.MaxDescriptors) || p.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	if err := descriptor.backing.acquire(); err != nil {
		return component.Value{}, err
	}
	id := uint32(p.next)
	p.next++
	p.descriptors[id] = descriptor
	value, err := c.NewResource(p.descriptorType, id)
	if err != nil {
		delete(p.descriptors, id)
		descriptor.backing.mu.Lock()
		descriptor.backing.refs--
		descriptor.backing.mu.Unlock()
	}
	return value, err
}
func (p *FileSystem) descriptor(c *component.Caller, ref *component.Resource) (fsDescriptor, error) {
	id, err := c.ResourceRep(p.descriptorType, ref)
	if err != nil {
		return fsDescriptor{}, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return fsDescriptor{}, os.ErrClosed
	}
	d, ok := p.descriptors[id]
	if !ok {
		return fsDescriptor{}, vm.ErrReference
	}
	if err := d.backing.acquire(); err != nil {
		return fsDescriptor{}, err
	}
	return d, nil
}
func (p *FileSystem) path(name string) error {
	if uint64(len(name)) > uint64(p.config.MaxPathBytes) {
		return &FileSystemError{Code: FSNameTooLong}
	}
	if strings.IndexByte(name, 0) >= 0 {
		return &FileSystemError{Code: FSInvalid}
	}
	return nil
}
func fsOK(value *component.Value) *component.Value { return streamOK(value) }
func fsFailure(err error) *component.Value {
	return &component.Value{Case: 1, Elements: []component.Value{{Bits: uint64(fsError(err))}}}
}
