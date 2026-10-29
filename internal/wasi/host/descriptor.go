//go:build linux || darwin

package host

import "os"

// Descriptor borrows an exclusively owned file's descriptor without File.Fd's
// implicit conversion to blocking mode. The caller keeps the owning lease
// alive for the entire syscall; no address or descriptor escapes that lease.
func Descriptor(file *os.File) int {
	if file == nil {
		return -1
	}
	raw, err := file.SyscallConn()
	if err != nil {
		return -1
	}
	fd := -1
	if err = raw.Control(func(n uintptr) { fd = int(n) }); err != nil {
		return -1
	}
	return fd
}
