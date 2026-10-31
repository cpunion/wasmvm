//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"math/bits"
	"os"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type fileError string

func (e fileError) Error() string { return string(e) }

var fileErrorCases = func() map[string]uint32 {
	m := map[string]uint32{}
	for i, f := range filesystemWITNamedTypes["error-code"].fields {
		m[f.name] = uint32(i)
	}
	return m
}()

func fileErrorCode(err error) uint32 {
	var named fileError
	if errors.As(err, &named) {
		if code, ok := fileErrorCases[string(named)]; ok {
			return code
		}
	}
	if errors.Is(err, wasm.ErrLimit) {
		return fileErrorCases["insufficient-memory"]
	}
	if errors.Is(err, host.ErrOutside) {
		return fileErrorCases["not-permitted"]
	}
	if errors.Is(err, os.ErrClosed) || errors.Is(err, io.ErrClosedPipe) {
		return fileErrorCases["bad-descriptor"]
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fileErrorCases["interrupted"]
	}
	for _, entry := range filesystemErrnos {
		if errors.Is(err, entry.err) {
			return fileErrorCases[entry.name]
		}
	}
	return fileErrorCases["other"]
}
func fileCompletionCode(err error) uint32 {
	if err == nil || errors.Is(err, io.EOF) {
		return 0
	}
	return fileErrorCode(err) + 1
}
func fileCompletionValue(code uint32) component.Value {
	if code == 0 {
		return component.Value{}
	}
	value := component.Value{Case: code - 1}
	if value.Case == fileErrorCases["other"] {
		value.Elements = []component.Value{{}}
	}
	return component.Value{Case: 1, Elements: []component.Value{value}}
}
func fileResult(v *component.Value, err error) *component.Value {
	if err != nil {
		value := fileCompletionValue(fileCompletionCode(err))
		return &value
	}
	out := &component.Value{}
	if v != nil {
		out.Elements = []component.Value{*v}
	}
	return out
}
func fileType(mode uint32) component.Value {
	code := uint32(7)
	switch mode & unix.S_IFMT {
	case unix.S_IFBLK:
		code = 0
	case unix.S_IFCHR:
		code = 1
	case unix.S_IFDIR:
		code = 2
	case unix.S_IFIFO:
		code = 3
	case unix.S_IFLNK:
		code = 4
	case unix.S_IFREG:
		code = 5
	case unix.S_IFSOCK:
		code = 6
	}
	value := component.Value{Case: code}
	if code == 7 {
		value.Elements = []component.Value{{}}
	}
	return value
}
func fileStatValue(st host.Stat) component.Value {
	fields := []component.Value{fileType(st.Mode), {Bits: st.Nlink}, {Bits: st.Size}}
	for _, instant := range st.Instants {
		fields = append(fields, component.Value{Case: 1, Elements: []component.Value{{Elements: []component.Value{{Bits: uint64(instant.Seconds)}, {Bits: uint64(instant.Nanoseconds)}}}}})
	}
	return component.Value{Elements: fields}
}
func fileHash(st host.Stat) component.Value {
	low := st.Dev ^ bits.RotateLeft64(st.Inode, 17) ^ bits.RotateLeft64(uint64(st.Instants[1].Seconds), 37) ^ st.Size ^ uint64(st.Instants[1].Nanoseconds)
	high := uint64(st.Instants[2].Seconds) ^ bits.RotateLeft64(st.Nlink, 11) ^ bits.RotateLeft64(uint64(st.Mode), 43) ^ uint64(st.Instants[2].Nanoseconds)
	low = (low ^ (low >> 30)) * 0xbf58476d1ce4e5b9
	high = (high ^ (high >> 27)) * 0x94d049bb133111eb
	return component.Value{Elements: []component.Value{{Bits: low ^ (low >> 31)}, {Bits: high ^ (high >> 31)}}}
}
func fileTimes(at, mt component.Value) ([2]host.Instant, uint32, error) {
	var times [2]host.Instant
	var flags uint32
	for i, value := range []component.Value{at, mt} {
		switch value.Case {
		case 1:
			flags |= 2 << uint(i*2)
		case 2:
			fields := value.Elements[0].Elements
			times[i] = host.Instant{Seconds: int64(fields[0].Bits), Nanoseconds: uint32(fields[1].Bits)}
			if times[i].Nanoseconds >= 1_000_000_000 {
				return times, 0, fileError("invalid")
			}
			flags |= 1 << uint(i*2)
		}
	}
	return times, flags, nil
}

var filesystemErrnos = []struct {
	err  error
	name string
}{
	{unix.EACCES, "access"},
	{unix.EAGAIN, "other"},
	{unix.EALREADY, "already"},
	{unix.EBADF, "bad-descriptor"},
	{unix.EBUSY, "busy"},
	{unix.EDEADLK, "deadlock"},
	{unix.EDQUOT, "quota"},
	{unix.EEXIST, "exist"},
	{unix.EFBIG, "file-too-large"},
	{unix.EILSEQ, "illegal-byte-sequence"},
	{unix.EINPROGRESS, "in-progress"},
	{unix.EINTR, "interrupted"},
	{unix.EINVAL, "invalid"},
	{unix.EIO, "io"},
	{unix.EISDIR, "is-directory"},
	{unix.ELOOP, "loop"},
	{unix.EMLINK, "too-many-links"},
	{unix.EMSGSIZE, "message-size"},
	{unix.ENAMETOOLONG, "name-too-long"},
	{unix.ENODEV, "no-device"},
	{unix.ENOENT, "no-entry"},
	{unix.ENOLCK, "no-lock"},
	{unix.ENOMEM, "insufficient-memory"},
	{unix.ENOSPC, "insufficient-space"},
	{unix.ENOTDIR, "not-directory"},
	{unix.ENOTEMPTY, "not-empty"},
	{unix.ENOTRECOVERABLE, "not-recoverable"},
	{unix.ENOTSUP, "unsupported"},
	{unix.EOPNOTSUPP, "unsupported"},
	{unix.ENOSYS, "unsupported"},
	{unix.ENOTTY, "no-tty"},
	{unix.ENXIO, "no-such-device"},
	{unix.EOVERFLOW, "overflow"},
	{unix.EPERM, "not-permitted"},
	{unix.EPIPE, "pipe"},
	{unix.EROFS, "read-only"},
	{unix.ESPIPE, "invalid-seek"},
	{unix.ETXTBSY, "text-file-busy"},
	{unix.EXDEV, "cross-device"},
}
