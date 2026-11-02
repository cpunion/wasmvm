//go:build linux || darwin

package preview1

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cpunion/wasmvm/pkg/vm"
)

type definition struct {
	name, signature string
	memoryOnly      bool
}

var functions = []definition{
	{"args_get", "ii", true}, {"args_sizes_get", "ii", true}, {"environ_get", "ii", true}, {"environ_sizes_get", "ii", true},
	{"clock_res_get", "ii", false}, {"clock_time_get", "iIi", false}, {"fd_advise", "iIIi", false}, {"fd_allocate", "iII", false},
	{"fd_close", "i", false}, {"fd_datasync", "i", false}, {"fd_fdstat_get", "ii", false}, {"fd_fdstat_set_flags", "ii", false}, {"fd_fdstat_set_rights", "iII", false},
	{"fd_filestat_get", "ii", false}, {"fd_filestat_set_size", "iI", false}, {"fd_filestat_set_times", "iIIi", false}, {"fd_pread", "iiiIi", false},
	{"fd_prestat_get", "ii", false}, {"fd_prestat_dir_name", "iii", false}, {"fd_pwrite", "iiiIi", false}, {"fd_read", "iiii", false},
	{"fd_readdir", "iiiIi", false}, {"fd_renumber", "ii", false}, {"fd_seek", "iIii", false}, {"fd_sync", "i", false}, {"fd_tell", "ii", false}, {"fd_write", "iiii", false},
	{"path_create_directory", "iii", false}, {"path_filestat_get", "iiiii", false}, {"path_filestat_set_times", "iiiiIIi", false}, {"path_link", "iiiiiii", false},
	{"path_open", "iiiiiIIii", false}, {"path_readlink", "iiiiii", false}, {"path_remove_directory", "iii", false}, {"path_rename", "iiiiii", false},
	{"path_symlink", "iiiii", false}, {"path_unlink_file", "iii", false}, {"poll_oneoff", "iiii", false}, {"proc_exit", "i", false}, {"proc_raise", "i", false},
	{"sched_yield", "", false}, {"random_get", "ii", false}, {"sock_accept", "iii", false}, {"sock_recv", "iiiiii", false}, {"sock_send", "iiiii", false}, {"sock_shutdown", "ii", false},
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

func (c *abiCall) u32Pair(first uint64, a uint32, second uint64, b uint32) error {
	if first > 1<<32-4 || second > 1<<32-4 {
		return ErrFault
	}
	var data [8]byte
	binary.LittleEndian.PutUint32(data[:4], a)
	binary.LittleEndian.PutUint32(data[4:], b)
	if err := c.caller.WriteMemoryBatch(0, []vm.MemoryWrite{{Offset: first, Data: data[:4]}, {Offset: second, Data: data[4:]}}); err != nil {
		return ErrFault
	}
	return nil
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
