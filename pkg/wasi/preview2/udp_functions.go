//go:build linux || darwin

package preview2

import (
	"errors"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var udpSocketMethods = []string{"start-bind", "finish-bind", "stream", "local-address", "remote-address", "address-family", "unicast-hop-limit", "set-unicast-hop-limit", "receive-buffer-size", "set-receive-buffer-size", "send-buffer-size", "set-send-buffer-size", "subscribe"}
var udpIncomingMethods = []string{"receive", "subscribe"}
var udpOutgoingMethods = []string{"check-send", "send", "subscribe"}

func (p *UDP) contracts() map[string]component.FunctionCheck {
	code := &interfaceShape{kind: component.KindEnum, labels: networkErrorLabels}
	family := &interfaceShape{kind: component.KindEnum, labels: []string{"ipv4", "ipv6"}}
	address := socketAddressShape()
	bytes := elementShape(component.KindList, scalarShape(component.KindU8))
	incoming := aggregateShape(component.KindRecord, shapeField{"data", bytes}, shapeField{"remote-address", address})
	option := elementShape(component.KindOption, address)
	outgoing := aggregateShape(component.KindRecord, shapeField{"data", bytes}, shapeField{"remote-address", option})
	result := func(ok *interfaceShape) *interfaceShape {
		return aggregateShape(component.KindResult, shapeField{"", ok}, shapeField{"", code})
	}
	own := func(kind uint32) *interfaceShape { return resourceShape(component.KindOwn, p.resourceType(kind)) }
	checks := map[string]component.FunctionCheck{"create-udp-socket": shapeContract([]*interfaceShape{family}, result(own(0)))}
	for kind, names := range [][]string{udpSocketMethods, udpIncomingMethods, udpOutgoingMethods} {
		prefix := []string{"udp-socket", "incoming-datagram-stream", "outgoing-datagram-stream"}[kind]
		for _, name := range names {
			args := []*interfaceShape{resourceShape(component.KindBorrow, p.resourceType(uint32(kind)))}
			var value *interfaceShape
			switch name {
			case "start-bind":
				args = append(args, resourceShape(component.KindBorrow, p.network.networkType), address)
			case "stream":
				args = append(args, option)
				value = aggregateShape(component.KindTuple, shapeField{"", own(1)}, shapeField{"", own(2)})
			case "local-address", "remote-address":
				value = address
			case "unicast-hop-limit":
				value = scalarShape(component.KindU8)
			case "set-unicast-hop-limit":
				args = append(args, scalarShape(component.KindU8))
			case "receive-buffer-size", "send-buffer-size", "check-send":
				value = scalarShape(component.KindU64)
			case "set-receive-buffer-size", "set-send-buffer-size":
				args = append(args, scalarShape(component.KindU64))
			case "receive":
				args = append(args, scalarShape(component.KindU64))
				value = elementShape(component.KindList, incoming)
			case "send":
				args = append(args, elementShape(component.KindList, outgoing))
				value = scalarShape(component.KindU64)
			}
			value = result(value)
			if name == "subscribe" {
				value = resourceShape(component.KindOwn, p.network.io.poller.resource)
			}
			if name == "address-family" {
				value = family
			}
			checks["[method]"+prefix+"."+name] = shapeContract(args, value)
		}
	}
	return checks
}
func (p *UDP) Imports() component.Imports {
	if p == nil || p.socketType == nil {
		return component.Imports{}
	}
	out := p.network.Imports()
	contracts := p.contracts()
	functions := map[string]component.HostFunction{}
	checks := map[string]component.FunctionCheck{}
	for kind, names := range [][]string{udpSocketMethods, udpIncomingMethods, udpOutgoingMethods} {
		prefix := []string{"udp-socket", "incoming-datagram-stream", "outgoing-datagram-stream"}[kind]
		for _, name := range names {
			key := "[method]" + prefix + "." + name
			functions[key] = p.method(uint32(kind), name)
			checks[key] = contracts[key]
		}
	}
	out.Instances[UDPInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"udp-socket": p.socketType, "incoming-datagram-stream": p.inputType, "outgoing-datagram-stream": p.outputType, "network": p.network.networkType, "pollable": p.network.io.poller.resource}, Functions: functions, FunctionChecks: checks}
	out.Instances[UDPCreateInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"udp-socket": p.socketType, "network": p.network.networkType}, Functions: map[string]component.HostFunction{"create-udp-socket": p.method(0, "create-udp-socket")}, FunctionChecks: map[string]component.FunctionCheck{"create-udp-socket": contracts["create-udp-socket"]}}
	return out
}
func (p *UDP) method(kind uint32, name string) component.HostFunction {
	return func(c *component.Caller, args []component.Value) (*component.Value, error) {
		if err := p.begin(c.Context()); err != nil {
			return nil, err
		}
		defer p.active.Done()
		var value component.Value
		payload := true
		var err error
		if name == "create-udp-socket" {
			value, err = p.create(c, uint32(args[0].Bits))
		} else {
			var e *udpEntry
			e, err = p.entry(c, args[0], kind)
			if err != nil {
				return nil, err
			}
			switch name {
			case "subscribe":
				return p.subscribe(c, e)
			case "receive":
				value, err = p.receive(c, e, args[1].Bits)
			case "check-send":
				value, err = p.checkSend(c, e)
			case "send":
				value, err = p.send(c, e, args[1].Elements)
			default:
				value, payload, err = p.socketExecute(c, e, name, args)
			}
		}
		if failed := c.Context().Err(); failed != nil {
			return nil, failed
		}
		if p.ctx.Err() != nil {
			return nil, wasm.ErrClosed
		}
		if name == "address-family" {
			return &value, err
		}
		if err != nil {
			if errors.Is(err, wasm.ErrTrap) || errors.Is(err, vm.ErrReference) || errors.Is(err, vm.ErrType) || errors.Is(err, wasm.ErrLimit) {
				return nil, err
			}
			return socketFailure(err), nil
		}
		if payload {
			return streamOK(&value), nil
		}
		return streamOK(nil), nil
	}
}
