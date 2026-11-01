//go:build linux || darwin

package preview2

import "github.com/cpunion/wasmvm/internal/wasi/host"

var socketSendDatagram = host.SocketSendDatagram
var socketReceiveDatagram = host.SocketReceiveDatagram
var socketDisconnectDatagram = host.SocketDisconnectDatagram
