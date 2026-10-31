//go:build linux || darwin

package preview2

import (
	"context"
	"io"
	"os"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

func (p *FileSystem) open(c *component.Caller, parent fsDescriptor, pathFlags uint64, path string, openFlags, flags uint64) (component.Value, error) {
	directory, err := fsDirectory(parent, false)
	if err != nil {
		return component.Value{}, err
	}
	if err = p.path(path); err != nil {
		return component.Value{}, err
	}
	if pathFlags&^uint64(1) != 0 || openFlags&^uint64(15) != 0 || flags&^uint64(63) != 0 {
		return component.Value{}, fsCode(FSInvalid)
	}
	if openFlags&(1|8) != 0 || flags&(fsWrite|fsMutate) != 0 {
		if err = fsWritable(parent); err != nil {
			return component.Value{}, err
		}
	}
	mode := unix.O_RDONLY
	if flags&fsWrite != 0 {
		mode = unix.O_WRONLY
		if flags&fsRead != 0 {
			mode = unix.O_RDWR
		}
	}
	if openFlags&1 != 0 {
		mode |= unix.O_CREAT
	}
	if openFlags&2 != 0 {
		mode |= unix.O_DIRECTORY
	}
	if openFlags&4 != 0 {
		mode |= unix.O_EXCL
	}
	if openFlags&8 != 0 {
		mode |= unix.O_TRUNC
	}
	if pathFlags&1 == 0 {
		mode |= unix.O_NOFOLLOW
	}
	if flags&(fsFileSync|fsDataSync) != 0 {
		mode |= unix.O_SYNC
	}
	file, child, err := directory.Open(path, mode, 0600)
	if err != nil {
		return component.Value{}, err
	}
	if child == nil {
		flags &^= fsMutate
	}
	b := &fsBacking{owner: p, file: file, directory: child, refs: 1}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		b.close()
		return component.Value{}, os.ErrClosed
	}
	p.backings[b] = struct{}{}
	p.mu.Unlock()
	defer b.release()
	return p.publish(c, fsDescriptor{b, flags})
}
func (p *FileSystem) readAt(ctx context.Context, d fsDescriptor, length, offset uint64) (component.Value, error) {
	if d.flags&fsRead == 0 {
		return component.Value{}, fsCode(FSNotPermitted)
	}
	if offset > 1<<63-1 {
		return component.Value{}, fsCode(FSOverflow)
	}
	if length > uint64(p.config.MaxReadBytes) {
		length = uint64(p.config.MaxReadBytes)
	}
	if length > uint64(1<<63-1)-offset {
		return component.Value{}, fsCode(FSOverflow)
	}
	data := make([]byte, int(length))
	d.backing.ioMu.Lock()
	n, err := d.backing.file.ReadAt(data, int64(offset))
	d.backing.ioMu.Unlock()
	if failed := p.check(ctx); failed != nil {
		return component.Value{}, failed
	}
	eof := err == io.EOF
	if err != nil && !eof {
		return component.Value{}, err
	}
	flag := uint64(0)
	if eof {
		flag = 1
	}
	return component.Value{Elements: []component.Value{component.ByteList(data[:n]), {Bits: flag}}}, nil
}
func (p *FileSystem) writeAt(ctx context.Context, d fsDescriptor, value component.Value, offset uint64) (component.Value, error) {
	if err := fsWritable(d); err != nil {
		return component.Value{}, err
	}
	data, err := value.ByteString()
	if err != nil {
		return component.Value{}, err
	}
	if uint64(len(data)) > uint64(p.config.MaxWriteBytes) {
		data = data[:p.config.MaxWriteBytes]
	}
	if offset > 1<<63-1 || uint64(len(data)) > uint64(1<<63-1)-offset {
		return component.Value{}, fsCode(FSOverflow)
	}
	d.backing.ioMu.Lock()
	n := 0
	err = d.backing.writeFailure
	if err == nil {
		n, err = d.backing.file.WriteAt([]byte(data), int64(offset))
	}
	d.backing.ioMu.Unlock()
	if failed := p.check(ctx); failed != nil {
		return component.Value{}, failed
	}
	if err != nil && n == 0 {
		return component.Value{}, err
	}
	return component.Value{Bits: uint64(n)}, nil
}
func (p *FileSystem) readDirectory(c *component.Caller, d fsDescriptor) (component.Value, error) {
	directory, err := fsDirectory(d, false)
	if err != nil {
		return component.Value{}, err
	}
	if d.flags&fsRead == 0 {
		return component.Value{}, fsCode(FSNotPermitted)
	}
	values, err := host.ReadDir(c.Context(), directory, p.config.MaxDirectoryEntries+2)
	if err != nil {
		return component.Value{}, err
	}
	entries := make([]host.Dirent, 0, len(values))
	for _, entry := range values {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}
		if !utf8.ValidString(entry.Name) {
			return component.Value{}, fsCode(FSIllegalByteSequence)
		}
		entries = append(entries, entry)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return component.Value{}, os.ErrClosed
	}
	if uint64(len(p.entries)) >= uint64(p.config.MaxDirectoryStreams) || p.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	id := uint32(p.next)
	p.next++
	p.entries[id] = &fsEntries{entries: entries}
	value, err := c.NewResource(p.entriesType, id)
	if err != nil {
		delete(p.entries, id)
	}
	return value, err
}
func (p *FileSystem) statAt(d fsDescriptor, pathFlags uint64, path string) (host.Stat, error) {
	dir, err := fsDirectory(d, false)
	if err != nil {
		return host.Stat{}, err
	}
	if err = p.path(path); err != nil {
		return host.Stat{}, err
	}
	return dir.Stat(path, pathFlags&1 != 0)
}
func (p *FileSystem) mutatePath(d fsDescriptor, name, path string) error {
	directory, err := fsDirectory(d, true)
	if err != nil {
		return err
	}
	if err = p.path(path); err != nil {
		return err
	}
	switch name {
	case "create-directory-at":
		return directory.Mkdir(path)
	case "remove-directory-at":
		return directory.Remove(path, true)
	default:
		return directory.Remove(path, false)
	}
}
func (p *FileSystem) pair(c *component.Caller, d fsDescriptor, name string, args []component.Value) error {
	sourcePath, destinationIndex := args[1].Text, 2
	if name == "link-at" {
		sourcePath = args[2].Text
		destinationIndex = 3
	}
	source, err := fsDirectory(d, true)
	if err != nil {
		return err
	}
	other, err := p.descriptor(c, args[destinationIndex].Resource)
	if err != nil {
		return err
	}
	defer other.backing.release()
	destination, err := fsDirectory(other, true)
	if err != nil {
		return err
	}
	target := args[destinationIndex+1].Text
	if err = p.path(sourcePath); err != nil {
		return err
	}
	if err = p.path(target); err != nil {
		return err
	}
	if name == "link-at" {
		return host.Link(source, sourcePath, args[1].Bits&1 != 0, destination, target)
	}
	return host.Rename(source, sourcePath, destination, target)
}
func (p *FileSystem) setTimesAt(d fsDescriptor, args []component.Value) error {
	directory, err := fsDirectory(d, true)
	if err != nil {
		return err
	}
	path := args[2].Text
	if err = p.path(path); err != nil {
		return err
	}
	at, mt, flags, err := fsTimestamps(args[3], args[4])
	if err != nil {
		return err
	}
	return directory.SetTimes(path, args[1].Bits&1 != 0, at, mt, flags)
}
