//go:build linux

package host

import (
	"golang.org/x/sys/unix"
	"os"
	"strings"
)

// Root opens final components with O_NOFOLLOW. Linux O_PATH then succeeds on
// the symlink itself, so resolve its stable handle explicitly within the root.
func (d *Directory) pathHandle(name string) (*os.File, error) {
	for depth := 0; depth < 40; depth++ {
		if err := checkPath(name); err != nil {
			return nil, err
		}
		f, err := d.Root.OpenFile(name, unix.O_PATH, 0)
		if err != nil {
			return nil, pathError(err)
		}
		st, err := Fstat(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		if st.Mode&unix.S_IFMT != unix.S_IFLNK {
			return f, nil
		}
		var buf [4096]byte
		n, err := unix.Readlinkat(Descriptor(f), "", buf[:])
		f.Close()
		if err != nil {
			return nil, err
		}
		if n == len(buf) {
			return nil, unix.ENAMETOOLONG
		}
		target := string(buf[:n])
		if strings.HasPrefix(target, "/") {
			return nil, ErrOutside
		}
		if i := strings.LastIndexByte(name, '/'); i >= 0 {
			name = name[:i+1] + target
		} else {
			name = target
		}
	}
	return nil, unix.ELOOP
}
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
