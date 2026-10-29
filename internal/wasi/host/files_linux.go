//go:build linux

package host

import (
	"golang.org/x/sys/unix"
	"os"
)

func pathHandleFlags() int { return unix.O_PATH }
func convertStat(s unix.Stat_t) Stat {
	at, mt, ct, bad := statTimes(s.Atim, s.Mtim, s.Ctim)
	return Stat{Dev: uint64(s.Dev), Inode: uint64(s.Ino), Nlink: uint64(s.Nlink), Size: uint64(s.Size), Atime: at, Mtime: mt, Ctime: ct, Mode: uint32(s.Mode), TimeOverflow: bad}
}
func SetTimes(f *os.File, at, mt uint64, flags uint32) error {
	return unix.UtimesNanoAt(Descriptor(f), "", timestamps(at, mt, flags), unix.AT_EMPTY_PATH)
}
func Allocate(f *os.File, offset, length int64) error {
	return unix.Fallocate(Descriptor(f), 0, offset, length)
}

func pendingBytes(fd int) (int, error) { return unix.IoctlGetInt(fd, unix.TIOCINQ) }

const utimeOmit = unix.UTIME_OMIT
const utimeNow = unix.UTIME_NOW
