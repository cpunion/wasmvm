//go:build linux || darwin

package preview1

import (
	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
)

const directoryRights = allRights &^ (RightFdRead | RightFdWrite | RightFdSeek | RightFdTell | RightFdAllocate | RightFdFilestatSetSize | RightSockAccept | RightSockShutdown)
const regularRights = RightFdDatasync | RightFdRead | RightFdSeek | RightFdFdstatSetFlags | RightFdSync | RightFdTell | RightFdWrite | RightFdAdvise | RightFdAllocate | RightFdFilestatGet | RightFdFilestatSetSize | RightFdFilestatSetTimes | RightPollFdReadwrite

func (p *Process) path(c *abiCall, name string, a []uint64) error {
	if name == "path_open" {
		return p.openPath(c, a)
	}
	fd, at, n, rights := uint32(a[0]), a[1], a[2], Rights(0)
	switch name {
	case "path_create_directory":
		rights = RightPathCreateDirectory
	case "path_filestat_get":
		rights = RightPathFilestatGet
		at, n = a[2], a[3]
		if a[1]&^uint64(1) != 0 {
			return ErrInval
		}
		if err := c.check(a[4], 64); err != nil {
			return err
		}
	case "path_filestat_set_times":
		rights = RightPathFilestatSetTimes
		at, n = a[2], a[3]
		if a[1]&^uint64(1) != 0 {
			return ErrInval
		}
		if err := validTimes(a[4], a[5], a[6]); err != nil {
			return err
		}
	case "path_link":
		rights = RightPathLinkSource
		at, n = a[2], a[3]
		if a[1]&^uint64(1) != 0 {
			return ErrInval
		}
	case "path_readlink":
		rights = RightPathReadlink
		if err := c.check(a[3], a[4]); err != nil {
			return err
		}
		if err := c.check(a[5], 4); err != nil {
			return err
		}
	case "path_remove_directory":
		rights = RightPathRemoveDirectory
	case "path_rename":
		rights = RightPathRenameSource
	case "path_symlink":
		rights = RightPathSymlink
		fd = uint32(a[2])
		at, n = a[3], a[4]
	case "path_unlink_file":
		rights = RightPathUnlinkFile
	}
	d, release, err := p.hold(fd, rights)
	if err != nil {
		return err
	}
	defer release()
	if d.resource.dir == nil {
		return ErrNotdir
	}
	path, err := c.path(at, n)
	if err != nil {
		return err
	}
	directory := d.resource.dir
	switch name {
	case "path_create_directory":
		return directory.Mkdir(path)
	case "path_filestat_get":
		st, err := directory.Stat(path, a[1] != 0)
		if err != nil {
			return err
		}
		if st.TimeOverflow {
			return ErrOverflow
		}
		return c.write(a[4], statBytes(st))
	case "path_filestat_set_times":
		return directory.SetTimes(path, a[1] != 0, a[4], a[5], uint32(a[6]))
	case "path_link", "path_rename":
		otherFD, targetAt, targetN, otherRight := uint32(a[3]), a[4], a[5], RightPathRenameTarget
		if name == "path_link" {
			otherFD = uint32(a[4])
			targetAt, targetN = a[5], a[6]
			otherRight = RightPathLinkTarget
		}
		other, release, err := p.hold(otherFD, otherRight)
		if err != nil {
			return err
		}
		defer release()
		if other.resource.dir == nil {
			return ErrNotdir
		}
		target, err := c.path(targetAt, targetN)
		if err != nil {
			return err
		}
		if name == "path_link" {
			return host.Link(directory, path, a[1] != 0, other.resource.dir, target)
		}
		return host.Rename(directory, path, other.resource.dir, target)
	case "path_readlink":
		target, err := directory.Readlink(path)
		if err != nil {
			return err
		}
		count := uint64(len(target))
		if count > a[4] {
			count = a[4]
		}
		if err := c.write(a[3], []byte(target[:count])); err != nil {
			return err
		}
		return c.u32(a[5], uint32(count))
	case "path_remove_directory":
		return directory.Remove(path, true)
	case "path_symlink":
		target, err := c.path(a[0], a[1])
		if err != nil {
			return err
		}
		return directory.Symlink(target, path)
	case "path_unlink_file":
		return directory.Remove(path, false)
	}
	return ErrNosys
}
func (p *Process) openPath(c *abiCall, a []uint64) error {
	if a[1]&^uint64(1) != 0 || a[4]&^uint64(15) != 0 || a[7]&^uint64(31) != 0 {
		return ErrInval
	}
	if err := c.check(a[8], 4); err != nil {
		return err
	}
	rights := RightPathOpen
	if a[4]&1 != 0 {
		rights |= RightPathCreateFile
	}
	if a[4]&8 != 0 {
		rights |= RightPathFilestatSetSize
	}
	d, release, err := p.hold(uint32(a[0]), rights)
	if err != nil {
		return err
	}
	defer release()
	if d.resource.dir == nil {
		return ErrNotdir
	}
	base, inherit := Rights(a[5]), Rights(a[6])
	if base&^d.inherit != 0 || inherit&^d.inherit != 0 {
		return ErrNotcapable
	}
	path, err := c.path(a[2], a[3])
	if err != nil {
		return err
	}
	flags := unix.O_RDONLY
	if base&RightFdWrite != 0 || a[4]&8 != 0 {
		flags = unix.O_WRONLY
		if base&RightFdRead != 0 {
			flags = unix.O_RDWR
		}
	}
	if a[1]&1 == 0 {
		flags |= unix.O_NOFOLLOW
	}
	if a[4]&1 != 0 {
		flags |= unix.O_CREAT
	}
	if a[4]&2 != 0 {
		flags |= unix.O_DIRECTORY
	}
	if a[4]&4 != 0 {
		flags |= unix.O_EXCL
	}
	if a[4]&8 != 0 {
		flags |= unix.O_TRUNC
	}
	if a[7]&1 != 0 {
		flags |= unix.O_APPEND
	}
	if err := p.reserveDescriptor(); err != nil {
		return err
	}
	f, dir, err := d.resource.dir.Open(path, flags, 0666)
	if err != nil {
		p.unreserveDescriptor()
		return err
	}
	r := p.newResource(f, dir, nil)
	r.flags = uint16(a[7])
	if dir != nil {
		base &= directoryRights
	} else {
		base &= regularRights
		inherit = 0
	}
	fd, err := p.publishDescriptor(&descriptor{resource: r, base: base, inherit: inherit})
	if err != nil {
		r.close()
		return err
	}
	if err := c.u32(a[8], fd); err != nil {
		p.closeFD(fd)
		return err
	}
	return nil
}
