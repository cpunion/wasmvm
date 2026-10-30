//go:build linux || darwin

package preview2

import (
	"errors"

	"golang.org/x/sys/unix"
)

// Syscall errors use one mapping across TCP/UDP and typed IO error downcasts.
// Operations retry EINTR themselves and interpret EINPROGRESS by their state.
func socketError(err error) NetworkErrorCode {
	for _, entry := range socketErrnos {
		if errors.Is(err, entry.err) {
			return entry.code
		}
	}
	return networkError(err)
}

var socketErrnos = []struct {
	err  error
	code NetworkErrorCode
}{
	{unix.EACCES, NetworkAccessDenied}, {unix.EPERM, NetworkAccessDenied},
	{unix.EAFNOSUPPORT, NetworkNotSupported}, {unix.EPROTONOSUPPORT, NetworkNotSupported}, {unix.ENOPROTOOPT, NetworkNotSupported}, {unix.EOPNOTSUPP, NetworkNotSupported}, {unix.ENOSYS, NetworkNotSupported},
	{unix.EINVAL, NetworkInvalidArgument}, {unix.EDESTADDRREQ, NetworkInvalidState},
	{unix.ENOMEM, NetworkOutOfMemory}, {unix.ENOBUFS, NetworkOutOfMemory},
	{unix.ETIMEDOUT, NetworkTimeout},
	{unix.EALREADY, NetworkConcurrencyConflict}, {unix.EINPROGRESS, NetworkWouldBlock}, {unix.EAGAIN, NetworkWouldBlock},
	{unix.EBADF, NetworkInvalidState}, {unix.ENOTSOCK, NetworkInvalidState}, {unix.ENOTCONN, NetworkInvalidState}, {unix.EISCONN, NetworkInvalidState},
	{unix.EMFILE, NetworkNewSocketLimit}, {unix.ENFILE, NetworkNewSocketLimit},
	{unix.EADDRNOTAVAIL, NetworkAddressNotBindable}, {unix.EADDRINUSE, NetworkAddressInUse},
	{unix.ENETDOWN, NetworkRemoteUnreachable}, {unix.ENETUNREACH, NetworkRemoteUnreachable}, {unix.EHOSTUNREACH, NetworkRemoteUnreachable},
	{unix.ECONNREFUSED, NetworkConnectionRefused}, {unix.ECONNRESET, NetworkConnectionReset}, {unix.ECONNABORTED, NetworkConnectionAborted}, {unix.EPIPE, NetworkConnectionReset},
	{unix.EMSGSIZE, NetworkDatagramTooLarge},
}
