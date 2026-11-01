//go:build linux

package preview2

import "github.com/cpunion/wasmvm/internal/wasi/host"

var socketLength4 = host.SocketLength4
var socketLength6 = host.SocketLength6

const tcpKeepIdle = host.TCPKeepIdle
