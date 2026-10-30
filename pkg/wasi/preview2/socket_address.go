package preview2

import (
	"net/netip"

	"github.com/cpunion/wasmvm/pkg/component"
)

// SocketAddress preserves the complete WIT address, including IPv6 flow/scope.
// Address has no textual zone; ScopeID is the explicit numeric interface scope.
type SocketAddress struct {
	Address           netip.AddrPort
	FlowInfo, ScopeID uint32
}

func socketAddressError() error { return &NetworkError{Code: NetworkInvalidArgument} }
func socketAddressShape() *interfaceShape {
	ip := ipAddressShape()
	v4 := aggregateShape(component.KindRecord, shapeField{"port", scalarShape(component.KindU16)}, shapeField{"address", ip.fields[0].typ})
	v6 := aggregateShape(component.KindRecord, shapeField{"port", scalarShape(component.KindU16)}, shapeField{"flow-info", scalarShape(component.KindU32)}, shapeField{"address", ip.fields[1].typ}, shapeField{"scope-id", scalarShape(component.KindU32)})
	return aggregateShape(component.KindVariant, shapeField{"ipv4", v4}, shapeField{"ipv6", v6})
}
func socketAddressValue(address SocketAddress) (component.Value, error) {
	if !address.Address.IsValid() || address.Address.Addr().Zone() != "" {
		return component.Value{}, socketAddressError()
	}
	ip := ipAddressValue(address.Address.Addr())
	fields := []component.Value{{Bits: uint64(address.Address.Port())}, {Elements: ip.Elements}}
	if ip.Case == 0 {
		if address.FlowInfo != 0 || address.ScopeID != 0 {
			return component.Value{}, socketAddressError()
		}
	} else {
		fields = []component.Value{{Bits: uint64(address.Address.Port())}, {Bits: uint64(address.FlowInfo)}, {Elements: ip.Elements}, {Bits: uint64(address.ScopeID)}}
	}
	return component.Value{Case: ip.Case, Elements: []component.Value{{Elements: fields}}}, nil
}
func parseSocketAddress(value component.Value) (SocketAddress, error) {
	if value.Case > 1 || len(value.Elements) != 1 {
		return SocketAddress{}, socketAddressError()
	}
	fields := value.Elements[0].Elements
	count, ipIndex, words := 2, 1, 4
	if value.Case == 1 {
		count, ipIndex, words = 4, 2, 8
	}
	if len(fields) != count || fields[0].Bits > 65535 || len(fields[ipIndex].Elements) != words {
		return SocketAddress{}, socketAddressError()
	}
	var ip netip.Addr
	if value.Case == 0 {
		var raw [4]byte
		for i, v := range fields[ipIndex].Elements {
			if v.Bits > 255 {
				return SocketAddress{}, socketAddressError()
			}
			raw[i] = byte(v.Bits)
		}
		ip = netip.AddrFrom4(raw)
	} else {
		var raw [16]byte
		for i, v := range fields[ipIndex].Elements {
			if v.Bits > 65535 {
				return SocketAddress{}, socketAddressError()
			}
			raw[i*2], raw[i*2+1] = byte(v.Bits>>8), byte(v.Bits)
		}
		ip = netip.AddrFrom16(raw)
	}
	out := SocketAddress{Address: netip.AddrPortFrom(ip, uint16(fields[0].Bits))}
	if value.Case == 1 {
		if fields[1].Bits > 1<<32-1 || fields[3].Bits > 1<<32-1 {
			return SocketAddress{}, socketAddressError()
		}
		out.FlowInfo, out.ScopeID = uint32(fields[1].Bits), uint32(fields[3].Bits)
	}
	return out, nil
}
