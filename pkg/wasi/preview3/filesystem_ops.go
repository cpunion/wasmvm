//go:build linux || darwin

package preview3

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

func (p *FileSystem) open(ctx context.Context, parent fileDescriptor, args []component.Value) (*fileDescriptor, error) {
	directory, err := parent.directory(false)
	if err != nil {
		return nil, err
	}
	pathFlags, path, openFlags, flags := args[1].Bits, args[2].Text, args[3].Bits, args[4].Bits
	if err := p.path(path); err != nil {
		return nil, err
	}
	if openFlags&(1|8) != 0 || flags&(fsWrite|fsMutate) != 0 {
		if err := parent.writable(); err != nil {
			return nil, err
		}
	}
	mode := unix.O_RDONLY
	if flags&fsWrite != 0 {
		mode = unix.O_WRONLY
		if flags&fsRead != 0 {
			mode = unix.O_RDWR
		}
	}
	for i, flag := range []int{unix.O_CREAT, unix.O_DIRECTORY, unix.O_EXCL, unix.O_TRUNC} {
		if openFlags&(1<<uint(i)) != 0 {
			mode |= flag
		}
	}
	if pathFlags&1 == 0 {
		mode |= unix.O_NOFOLLOW
	}
	if flags&(fsFileSync|fsDataSync) != 0 {
		mode |= unix.O_SYNC
	}
	if err := p.reserveObject(ctx); err != nil {
		return nil, err
	}
	file, child, err := directory.Open(path, mode, 0600)
	if err != nil {
		p.cancelObjectReservation()
		return nil, err
	}
	if child == nil {
		flags &^= fsMutate
	}
	object, err := p.registerObject(file, child)
	if err != nil {
		return nil, err
	}

	return &fileDescriptor{object, flags}, nil
}
func (p *FileSystem) execute(ctx context.Context, d, other fileDescriptor, name string, args []component.Value) (*component.Value, *fileDescriptor, error) {
	if err := p.check(ctx); err != nil {
		return nil, nil, err
	}
	var value *component.Value
	var err error
	switch name {
	case "get-flags":
		value = &component.Value{Bits: d.flags}
	case "get-type", "stat", "metadata-hash":
		var st host.Stat
		st, err = host.Fstat(d.object.file)
		if err == nil {
			v := fileType(st.Mode)
			if name == "stat" {
				v = fileStatValue(st)
			} else if name == "metadata-hash" {
				v = fileHash(st)
			}
			value = &v
		}
	case "open-at":
		opened, failed := p.open(ctx, d, args)
		return nil, opened, failed
	case "stat-at", "metadata-hash-at":
		var directory *host.Directory
		directory, err = d.directory(false)
		if err == nil {
			err = p.path(args[2].Text)
		}
		if err == nil {
			var st host.Stat
			st, err = directory.Stat(args[2].Text, args[1].Bits&1 != 0)
			if err == nil {
				v := fileStatValue(st)
				if name == "metadata-hash-at" {
					v = fileHash(st)
				}
				value = &v
			}
		}
	case "is-same-object":
		var a, b host.Stat
		a, err = host.Fstat(d.object.file)
		if err == nil {
			b, err = host.Fstat(other.object.file)
		}
		value = &component.Value{}
		if err == nil && a.Dev == b.Dev && a.Inode == b.Inode {
			value.Bits = 1
		}
	case "advise":
		if args[1].Bits > 1<<63-1 || args[2].Bits > uint64(1<<63-1)-args[1].Bits {
			err = fileError("overflow")
		}
	case "sync", "sync-data":
		if d.flags&(fsWrite|fsMutate) != 0 {
			d.object.ioMu.Lock()
			err = d.object.file.Sync()
			d.object.ioMu.Unlock()
		}
	case "set-size":
		err = d.writable()
		if err == nil && d.object.directory != nil {
			err = fileError("is-directory")
		}
		if err == nil && args[1].Bits > 1<<63-1 {
			err = fileError("overflow")
		}
		if err == nil {
			d.object.ioMu.Lock()
			err = d.object.file.Truncate(int64(args[1].Bits))
			d.object.ioMu.Unlock()
		}
	case "set-times":
		err = d.writable()
		if err == nil {
			var times [2]host.Instant
			var flags uint32
			times, flags, err = fileTimes(args[1], args[2])
			if err == nil {
				err = host.SetInstants(d.object.file, times[0], times[1], flags)
			}
		}
	case "set-times-at":
		var dir *host.Directory
		dir, err = d.directory(true)
		if err == nil {
			err = p.path(args[2].Text)
		}
		if err == nil {
			var times [2]host.Instant
			var flags uint32
			times, flags, err = fileTimes(args[3], args[4])
			if err == nil {
				err = dir.SetInstants(args[2].Text, args[1].Bits&1 != 0, times[0], times[1], flags)
			}
		}
	case "readlink-at":
		var dir *host.Directory
		dir, err = d.directory(false)
		if err == nil {
			err = p.path(args[1].Text)
		}
		if err == nil {
			var text string
			text, err = dir.Readlink(args[1].Text)
			if err == nil {
				if strings.HasPrefix(text, "/") {
					err = fileError("not-permitted")
				} else if !utf8.ValidString(text) {
					err = fileError("illegal-byte-sequence")
				} else {
					err = p.path(text)
				}
			}
			value = &component.Value{Text: text}
		}
	case "create-directory-at", "remove-directory-at", "unlink-file-at":
		var dir *host.Directory
		dir, err = d.directory(true)
		if err == nil {
			err = p.path(args[1].Text)
		}
		if err == nil {
			if name == "create-directory-at" {
				err = dir.Mkdir(args[1].Text)
			} else {
				err = dir.Remove(args[1].Text, name == "remove-directory-at")
			}
		}
	case "link-at", "rename-at":
		var source, target *host.Directory
		source, err = d.directory(true)
		if err == nil {
			target, err = other.directory(true)
		}
		old, newPath := args[1].Text, args[3].Text
		if name == "link-at" {
			old, newPath = args[2].Text, args[4].Text
		}
		if err == nil {
			err = p.path(old)
		}
		if err == nil {
			err = p.path(newPath)
		}
		if err == nil {
			if name == "link-at" {
				err = host.Link(source, old, args[1].Bits&1 != 0, target, newPath)
			} else {
				err = host.Rename(source, old, target, newPath)
			}
		}
	case "symlink-at":
		var dir *host.Directory
		dir, err = d.directory(true)
		if err == nil {
			err = p.path(args[1].Text)
		}
		if err == nil {
			err = p.path(args[2].Text)
		}
		if err == nil {
			err = dir.Symlink(args[1].Text, args[2].Text)
		}
	default:
		return nil, nil, fileError("unsupported")
	}
	if failed := p.check(ctx); failed != nil {
		return nil, nil, failed
	}
	return value, nil, err
}
