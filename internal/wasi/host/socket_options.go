//go:build linux || darwin

package host

import (
	"strings"

	"golang.org/x/sys/unix"
)

// ConfigureIPv6Socket fixes the resource's initial hop limit instead of exposing
// the kernel's negative "inherit the route default" sentinel as a WASI u8.
func ConfigureIPv6Socket(fd int) error {
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_IPV6, unix.IPV6_V6ONLY, 1); err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_IPV6, unix.IPV6_UNICAST_HOPS, 64)
}

func SocketOption(fd int, family uint32, name string, value uint64) (uint64, error) {
	set := strings.HasPrefix(name, "set-")
	name = strings.TrimPrefix(name, "set-")
	level, option := unix.SOL_SOCKET, unix.SO_KEEPALIVE
	multiplier := uint64(1)
	switch name {
	case "keep-alive-enabled":
	case "keep-alive-idle-time":
		level, option, multiplier = unix.IPPROTO_TCP, TCPKeepIdle, 1000000000
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
		return 0, unix.EINVAL
	}
	if set {
		if name != "keep-alive-enabled" && value == 0 {
			return 0, unix.EINVAL
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
		return 0, unix.SetsockoptInt(fd, level, option, int(value))
	}
	result, err := unix.GetsockoptInt(fd, level, option)
	if err != nil {
		return 0, err
	}
	if name == "keep-alive-enabled" {
		if result != 0 {
			result = 1
		}
	}
	return uint64(result) * multiplier, nil
}
