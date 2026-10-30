//go:build linux || darwin

package preview2

import (
	"strings"

	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

func tcpOption(fd int, family uint32, name string, value uint64) (component.Value, error) {
	set := strings.HasPrefix(name, "set-")
	name = strings.TrimPrefix(name, "set-")
	level, option := unix.SOL_SOCKET, unix.SO_KEEPALIVE
	multiplier := uint64(1)
	switch name {
	case "keep-alive-enabled":
	case "keep-alive-idle-time":
		level, option, multiplier = unix.IPPROTO_TCP, tcpKeepIdle, 1000000000
	case "keep-alive-interval":
		level, option, multiplier = unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, 1000000000
	case "keep-alive-count":
		level, option = unix.IPPROTO_TCP, unix.TCP_KEEPCNT
	case "hop-limit":
		level, option = unix.IPPROTO_IP, unix.IP_TTL
		if family == 1 {
			level, option = unix.IPPROTO_IPV6, unix.IPV6_UNICAST_HOPS
		}
	case "receive-buffer-size":
		option = unix.SO_RCVBUF
	case "send-buffer-size":
		option = unix.SO_SNDBUF
	default:
		return component.Value{}, tcpCode(NetworkInvalidArgument)
	}
	if set {
		if name != "keep-alive-enabled" && value == 0 {
			return component.Value{}, tcpCode(NetworkInvalidArgument)
		}
		if multiplier != 1 {
			value = (value-1)/multiplier + 1
			if value > 32767 {
				value = 32767
			}
		}
		if name == "keep-alive-count" && value > 127 {
			value = 127
		}
		if value > 1<<31-1 {
			value = 1<<31 - 1
		}
		return component.Value{}, unix.SetsockoptInt(fd, level, option, int(value))
	}
	result, err := unix.GetsockoptInt(fd, level, option)
	if err != nil {
		return component.Value{}, err
	}
	if name == "keep-alive-enabled" {
		if result != 0 {
			result = 1
		}
	}
	return component.Value{Bits: uint64(result) * multiplier}, nil
}
