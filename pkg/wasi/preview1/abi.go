//go:build linux || darwin

package preview1

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cpunion/wasmvm/pkg/vm"
)

type definition struct{ name, signature string }

var functions = []definition{
	{"args_get", "ii"}, {"args_sizes_get", "ii"}, {"environ_get", "ii"}, {"environ_sizes_get", "ii"},
	{"clock_res_get", "ii"}, {"clock_time_get", "iIi"}, {"fd_advise", "iIIi"}, {"fd_allocate", "iII"},
	{"fd_close", "i"}, {"fd_datasync", "i"}, {"fd_fdstat_get", "ii"}, {"fd_fdstat_set_flags", "ii"}, {"fd_fdstat_set_rights", "iII"},
	{"fd_filestat_get", "ii"}, {"fd_filestat_set_size", "iI"}, {"fd_filestat_set_times", "iIIi"}, {"fd_pread", "iiiIi"},
	{"fd_prestat_get", "ii"}, {"fd_prestat_dir_name", "iii"}, {"fd_pwrite", "iiiIi"}, {"fd_read", "iiii"},
	{"fd_readdir", "iiiIi"}, {"fd_renumber", "ii"}, {"fd_seek", "iIii"}, {"fd_sync", "i"}, {"fd_tell", "ii"}, {"fd_write", "iiii"},
	{"path_create_directory", "iii"}, {"path_filestat_get", "iiiii"}, {"path_filestat_set_times", "iiiiIIi"}, {"path_link", "iiiiiii"},
	{"path_open", "iiiiiIIii"}, {"path_readlink", "iiiiii"}, {"path_remove_directory", "iii"}, {"path_rename", "iiiiii"},
	{"path_symlink", "iiiii"}, {"path_unlink_file", "iii"}, {"poll_oneoff", "iiii"}, {"proc_exit", "i"}, {"proc_raise", "i"},
	{"sched_yield", ""}, {"random_get", "ii"}, {"sock_accept", "iii"}, {"sock_recv", "iiiiii"}, {"sock_send", "iiiii"}, {"sock_shutdown", "ii"},
}

func (e Errno) Error() string { return fmt.Sprintf("WASI errno %d", uint16(e)) }

type abiCall struct {
	caller  *vm.Caller
	ctx     context.Context
	p       *Process
	scratch uint64
}

func (c *abiCall) check(at, n uint64) error {
	if at > 1<<32 || n > 1<<32-at {
		return ErrFault
	}
	if err := c.caller.ReadMemory(0, at+n, nil); err != nil {
		return ErrFault
	}
	return nil
}
func (c *abiCall) buffer(n uint64) ([]byte, error) {
	if n > uint64(c.p.config.Limits.MaxIOBytes)-c.scratch {
		return nil, ErrNomem
	}
	c.scratch += n
	return make([]byte, int(n)), nil
}
func (c *abiCall) read(at, n uint64) ([]byte, error) {
	if err := c.check(at, n); err != nil {
		return nil, err
	}
	b, err := c.buffer(n)
	if err != nil {
		return nil, err
	}
	if err = c.caller.ReadMemory(0, at, b); err != nil {
		return nil, ErrFault
	}
	return b, nil
}
func (c *abiCall) write(at uint64, b []byte) error {
	if err := c.caller.WriteMemory(0, at, b); err != nil {
		return ErrFault
	}
	return nil
}
func (c *abiCall) u32(at uint64, v uint32) error {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return c.write(at, b[:])
}
func (c *abiCall) u64(at, v uint64) error {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return c.write(at, b[:])
}
func (c *abiCall) path(at, n uint64) (string, error) {
	if n > uint64(c.p.config.Limits.MaxPathBytes) {
		return "", ErrNametoolong
	}
	b, err := c.read(at, n)
	if err != nil {
		return "", err
	}
	if strings.IndexByte(string(b), 0) >= 0 {
		return "", ErrInval
	}
	return string(b), nil
}

type vector struct{ at, n uint64 }

func (c *abiCall) vectors(at, n uint64) ([]vector, uint64, error) {
	if n > uint64(c.p.config.Limits.MaxIOVectors) {
		return nil, 0, ErrInval
	}
	b, err := c.read(at, n*8)
	if err != nil {
		return nil, 0, err
	}
	vs := make([]vector, int(n))
	total := uint64(0)
	for i := range vs {
		v := vector{uint64(binary.LittleEndian.Uint32(b[i*8:])), uint64(binary.LittleEndian.Uint32(b[i*8+4:]))}
		if err := c.check(v.at, v.n); err != nil {
			return nil, 0, err
		}
		if v.n > uint64(c.p.config.Limits.MaxIOBytes)-total {
			return nil, 0, ErrNomem
		}
		vs[i] = v
		total += v.n
	}
	return vs, total, nil
}
