//go:build linux || darwin

package preview2

import (
	"math/bits"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

func fsCode(code FileSystemErrorCode) error { return &FileSystemError{Code: code} }
func fsType(mode uint32) uint32 {
	switch mode & unix.S_IFMT {
	case unix.S_IFBLK:
		return 1
	case unix.S_IFCHR:
		return 2
	case unix.S_IFDIR:
		return 3
	case unix.S_IFIFO:
		return 4
	case unix.S_IFLNK:
		return 5
	case unix.S_IFREG:
		return 6
	case unix.S_IFSOCK:
		return 7
	}
	return 0
}
func fsDatetime(nanos uint64) component.Value {
	return component.Value{Elements: []component.Value{{Bits: nanos / 1000000000}, {Bits: nanos % 1000000000}}}
}
func fsStatValue(stat host.Stat) (component.Value, error) {
	if stat.TimeOverflow {
		return component.Value{}, fsCode(FSOverflow)
	}
	times := []uint64{stat.Atime, stat.Mtime, stat.Ctime}
	fields := []component.Value{{Bits: uint64(fsType(stat.Mode))}, {Bits: stat.Nlink}, {Bits: stat.Size}}
	for _, at := range times {
		fields = append(fields, component.Value{Case: 1, Elements: []component.Value{fsDatetime(at)}})
	}
	return component.Value{Elements: fields}, nil
}
func fsHash(stat host.Stat) component.Value {
	// Metadata equality is advisory, not a cryptographic identity. Domain/object
	// identifiers and mutable metadata all contribute to the 128-bit value.
	low := stat.Dev ^ bits.RotateLeft64(stat.Inode, 17) ^ bits.RotateLeft64(stat.Mtime, 37) ^ stat.Size
	high := stat.Ctime ^ bits.RotateLeft64(stat.Nlink, 11) ^ bits.RotateLeft64(uint64(stat.Mode), 43)
	low = (low ^ (low >> 30)) * 0xbf58476d1ce4e5b9
	high = (high ^ (high >> 27)) * 0x94d049bb133111eb
	return component.Value{Elements: []component.Value{{Bits: low ^ (low >> 31)}, {Bits: high ^ (high >> 31)}}}
}
func fsTimestamps(at, mt component.Value) (uint64, uint64, uint32, error) {
	values := []component.Value{at, mt}
	var nanos [2]uint64
	var flags uint32
	for i, value := range values {
		switch value.Case {
		case 0:
		case 1:
			flags |= 2 << uint(2*i)
		case 2:
			datetime := value.Elements[0]
			sec, nano := datetime.Elements[0].Bits, datetime.Elements[1].Bits
			if nano >= 1000000000 {
				return 0, 0, 0, fsCode(FSInvalid)
			}
			if sec > (^uint64(0)-nano)/1000000000 {
				return 0, 0, 0, fsCode(FSOverflow)
			}
			nanos[i] = sec*1000000000 + nano
			flags |= 1 << uint(2*i)
		}
	}
	return nanos[0], nanos[1], flags, nil
}
func fsWritable(d fsDescriptor) error {
	if d.backing.directory != nil {
		if d.flags&fsMutate == 0 {
			return fsCode(FSReadOnly)
		}
	} else if d.flags&fsWrite == 0 {
		return fsCode(FSNotPermitted)
	}
	return nil
}
func fsDirectory(d fsDescriptor, mutate bool) (*host.Directory, error) {
	if d.backing.directory == nil {
		return nil, fsCode(FSNotDirectory)
	}
	if mutate {
		if err := fsWritable(d); err != nil {
			return nil, err
		}
	}
	return d.backing.directory, nil
}
