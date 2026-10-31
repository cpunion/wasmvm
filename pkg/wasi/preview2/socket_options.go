//go:build linux || darwin

package preview2

import (
	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
)

func socketOption(fd int, family uint32, name string, value uint64) (component.Value, error) {
	bits, err := host.SocketOption(fd, family, name, value)
	return component.Value{Bits: bits}, err
}
