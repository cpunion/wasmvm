//go:build darwin

package host

import (
	"golang.org/x/sys/unix"
	"os"
	"runtime"
	"time"
	"unsafe"
)

func (d *Directory) pathHandle(name string) (*os.File, error) {
	return d.Root.OpenFile(name, unix.O_EVTONLY, 0)
}
func convertStat(s unix.Stat_t) Stat {
	at, mt, ct, bad := statTimes(s.Atim, s.Mtim, s.Ctim)
	return Stat{Dev: uint64(s.Dev), Inode: uint64(s.Ino), Nlink: uint64(s.Nlink), Size: uint64(s.Size), Atime: at, Mtime: mt, Ctime: ct, Mode: uint32(s.Mode), TimeOverflow: bad}
}
func Allocate(f *os.File, offset, length int64) error {
	if length == 0 {
		return unix.EINVAL
	}
	s, err := f.Stat()
	if err != nil {
		return err
	}
	end := offset + length
	if end <= s.Size() {
		return nil
	}
	store := unix.Fstore_t{Flags: unix.F_ALLOCATEALL, Posmode: unix.F_PEOFPOSMODE, Length: end - s.Size()}
	if err := unix.FcntlFstore(uintptr(Descriptor(f)), unix.F_PREALLOCATE, &store); err != nil {
		return err
	}
	return f.Truncate(end)
}

// fsetattrlist preserves nanoseconds and modifies only requested fields on the
// open object. A path-based Chtimes would reopen a symlink race on Darwin.
func SetTimes(f *os.File, at, mt uint64, flags uint32) error {
	a := struct {
		Count, Reserved              uint16
		Common, Vol, Dir, File, Fork uint32
	}{Count: 5}
	var times [2]unix.Timespec
	n := 0
	now := uint64(time.Now().UnixNano())
	if flags&12 != 0 {
		a.Common |= 0x400
		v := mt
		if flags&8 != 0 {
			v = now
		}
		times[n] = unix.Timespec{Sec: int64(v / 1_000_000_000), Nsec: int64(v % 1_000_000_000)}
		n++
	}
	if flags&3 != 0 {
		a.Common |= 0x1000
		v := at
		if flags&2 != 0 {
			v = now
		}
		times[n] = unix.Timespec{Sec: int64(v / 1_000_000_000), Nsec: int64(v % 1_000_000_000)}
		n++
	}
	if n == 0 {
		return nil
	}
	_, _, err := unix.Syscall6(unix.SYS_FSETATTRLIST, uintptr(Descriptor(f)), uintptr(unsafe.Pointer(&a)), uintptr(unsafe.Pointer(&times)), uintptr(n)*unsafe.Sizeof(times[0]), 0, 0)
	runtime.KeepAlive(&a)
	runtime.KeepAlive(&times)
	runtime.KeepAlive(f)
	if err != 0 {
		return err
	}
	return nil
}

// Darwin FIONREAD from sys/filio.h.
func pendingBytes(fd int) (int, error) { return unix.IoctlGetInt(fd, 0x4004667f) }

const utimeOmit = -2
const utimeNow = -1
