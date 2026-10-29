//go:build linux || darwin

package preview1

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
	"io/fs"
)

var errnoMap = map[unix.Errno]Errno{
	unix.E2BIG:           Err2Big,
	unix.EACCES:          ErrAcces,
	unix.EADDRINUSE:      ErrAddrinuse,
	unix.EADDRNOTAVAIL:   ErrAddrnotavail,
	unix.EAFNOSUPPORT:    ErrAfnosupport,
	unix.EAGAIN:          ErrAgain,
	unix.EALREADY:        ErrAlready,
	unix.EBADF:           ErrBadf,
	unix.EBADMSG:         ErrBadmsg,
	unix.EBUSY:           ErrBusy,
	unix.ECANCELED:       ErrCanceled,
	unix.ECHILD:          ErrChild,
	unix.ECONNABORTED:    ErrConnaborted,
	unix.ECONNREFUSED:    ErrConnrefused,
	unix.ECONNRESET:      ErrConnreset,
	unix.EDEADLK:         ErrDeadlk,
	unix.EDESTADDRREQ:    ErrDestaddrreq,
	unix.EDOM:            ErrDom,
	unix.EDQUOT:          ErrDquot,
	unix.EEXIST:          ErrExist,
	unix.EFAULT:          ErrFault,
	unix.EFBIG:           ErrFbig,
	unix.EHOSTUNREACH:    ErrHostunreach,
	unix.EIDRM:           ErrIdrm,
	unix.EILSEQ:          ErrIlseq,
	unix.EINPROGRESS:     ErrInprogress,
	unix.EINTR:           ErrIntr,
	unix.EINVAL:          ErrInval,
	unix.EIO:             ErrIo,
	unix.EISCONN:         ErrIsconn,
	unix.EISDIR:          ErrIsdir,
	unix.ELOOP:           ErrLoop,
	unix.EMFILE:          ErrMfile,
	unix.EMLINK:          ErrMlink,
	unix.EMSGSIZE:        ErrMsgsize,
	unix.EMULTIHOP:       ErrMultihop,
	unix.ENAMETOOLONG:    ErrNametoolong,
	unix.ENETDOWN:        ErrNetdown,
	unix.ENETRESET:       ErrNetreset,
	unix.ENETUNREACH:     ErrNetunreach,
	unix.ENFILE:          ErrNfile,
	unix.ENOBUFS:         ErrNobufs,
	unix.ENODEV:          ErrNodev,
	unix.ENOENT:          ErrNoent,
	unix.ENOEXEC:         ErrNoexec,
	unix.ENOLCK:          ErrNolck,
	unix.ENOLINK:         ErrNolink,
	unix.ENOMEM:          ErrNomem,
	unix.ENOMSG:          ErrNomsg,
	unix.ENOPROTOOPT:     ErrNoprotoopt,
	unix.ENOSPC:          ErrNospc,
	unix.ENOSYS:          ErrNosys,
	unix.ENOTCONN:        ErrNotconn,
	unix.ENOTDIR:         ErrNotdir,
	unix.ENOTEMPTY:       ErrNotempty,
	unix.ENOTRECOVERABLE: ErrNotrecoverable,
	unix.ENOTSOCK:        ErrNotsock,
	unix.ENOTSUP:         ErrNotsup,
	unix.ENOTTY:          ErrNotty,
	unix.ENXIO:           ErrNxio,
	unix.EOVERFLOW:       ErrOverflow,
	unix.EOWNERDEAD:      ErrOwnerdead,
	unix.EPERM:           ErrPerm,
	unix.EPIPE:           ErrPipe,
	unix.EPROTO:          ErrProto,
	unix.EPROTONOSUPPORT: ErrProtonosupport,
	unix.EPROTOTYPE:      ErrPrototype,
	unix.ERANGE:          ErrRange,
	unix.EROFS:           ErrRofs,
	unix.ESPIPE:          ErrSpipe,
	unix.ESRCH:           ErrSrch,
	unix.ESTALE:          ErrStale,
	unix.ETIMEDOUT:       ErrTimedout,
	unix.ETXTBSY:         ErrTxtbsy,
	unix.EXDEV:           ErrXdev,
}

func errno(err error) Errno {
	if err == nil {
		return ErrSuccess
	}
	var e Errno
	if errors.As(err, &e) {
		return e
	}
	if errors.Is(err, host.ErrOutside) {
		return ErrNotcapable
	}
	if errors.Is(err, context.Canceled) {
		return ErrCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimedout
	}
	var u unix.Errno
	if errors.As(err, &u) {
		if v, ok := errnoMap[u]; ok {
			return v
		}
	}
	if errors.Is(err, fs.ErrClosed) {
		return ErrBadf
	}
	if errors.Is(err, fs.ErrInvalid) {
		return ErrInval
	}
	return ErrIo
}
