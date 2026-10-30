//go:build linux || darwin

package preview2

import (
	"errors"
	"io"
	"os"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type FileSystemErrorCode uint32

const (
	FSAccess FileSystemErrorCode = iota
	FSWouldBlock
	FSAlready
	FSBadDescriptor
	FSBusy
	FSDeadlock
	FSQuota
	FSExist
	FSFileTooLarge
	FSIllegalByteSequence
	FSInProgress
	FSInterrupted
	FSInvalid
	FSIo
	FSIsDirectory
	FSLoop
	FSTooManyLinks
	FSMessageSize
	FSNameTooLong
	FSNoDevice
	FSNoEntry
	FSNoLock
	FSInsufficientMemory
	FSInsufficientSpace
	FSNotDirectory
	FSNotEmpty
	FSNotRecoverable
	FSUnsupported
	FSNoTty
	FSNoSuchDevice
	FSOverflow
	FSNotPermitted
	FSPipe
	FSReadOnly
	FSInvalidSeek
	FSTextFileBusy
	FSCrossDevice
)

var fileSystemErrorLabels = []string{"access", "would-block", "already", "bad-descriptor", "busy", "deadlock", "quota", "exist", "file-too-large", "illegal-byte-sequence", "in-progress", "interrupted", "invalid", "io", "is-directory", "loop", "too-many-links", "message-size", "name-too-long", "no-device", "no-entry", "no-lock", "insufficient-memory", "insufficient-space", "not-directory", "not-empty", "not-recoverable", "unsupported", "no-tty", "no-such-device", "overflow", "not-permitted", "pipe", "read-only", "invalid-seek", "text-file-busy", "cross-device"}

// FileSystemError preserves structured provenance for IO error downcasts.
type FileSystemError struct {
	Code  FileSystemErrorCode
	Cause error
}

func (e *FileSystemError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "WASI filesystem error"
}
func (e *FileSystemError) Unwrap() error { return e.Cause }

func fsError(err error) FileSystemErrorCode {
	var structured *FileSystemError
	if errors.As(err, &structured) {
		return structured.Code
	}
	if errors.Is(err, wasm.ErrLimit) {
		return FSInsufficientMemory
	}
	if errors.Is(err, host.ErrOutside) {
		return FSNotPermitted
	}
	if errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) {
		return FSBadDescriptor
	}
	for _, entry := range fileSystemErrnos {
		if errors.Is(err, entry.err) {
			return entry.code
		}
	}
	return FSIo
}

var fileSystemErrnos = []struct {
	err  error
	code FileSystemErrorCode
}{
	{unix.EACCES, FSAccess},
	{unix.EAGAIN, FSWouldBlock},
	{unix.EALREADY, FSAlready},
	{unix.EBADF, FSBadDescriptor},
	{unix.EBUSY, FSBusy},
	{unix.EDEADLK, FSDeadlock},
	{unix.EDQUOT, FSQuota},
	{unix.EEXIST, FSExist},
	{unix.EFBIG, FSFileTooLarge},
	{unix.EILSEQ, FSIllegalByteSequence},
	{unix.EINPROGRESS, FSInProgress},
	{unix.EINTR, FSInterrupted},
	{unix.EINVAL, FSInvalid},
	{unix.EIO, FSIo},
	{unix.EISDIR, FSIsDirectory},
	{unix.ELOOP, FSLoop},
	{unix.EMLINK, FSTooManyLinks},
	{unix.EMSGSIZE, FSMessageSize},
	{unix.ENAMETOOLONG, FSNameTooLong},
	{unix.ENODEV, FSNoDevice},
	{unix.ENOENT, FSNoEntry},
	{unix.ENOLCK, FSNoLock},
	{unix.ENOMEM, FSInsufficientMemory},
	{unix.ENOSPC, FSInsufficientSpace},
	{unix.ENOTDIR, FSNotDirectory},
	{unix.ENOTEMPTY, FSNotEmpty},
	{unix.ENOTRECOVERABLE, FSNotRecoverable},
	{unix.ENOTSUP, FSUnsupported},
	{unix.EOPNOTSUPP, FSUnsupported},
	{unix.ENOSYS, FSUnsupported},
	{unix.ENOTTY, FSNoTty},
	{unix.ENXIO, FSNoSuchDevice},
	{unix.EOVERFLOW, FSOverflow},
	{unix.EPERM, FSNotPermitted},
	{unix.EPIPE, FSPipe},
	{unix.EROFS, FSReadOnly},
	{unix.ESPIPE, FSInvalidSeek},
	{unix.ETXTBSY, FSTextFileBusy},
	{unix.EXDEV, FSCrossDevice},
}
