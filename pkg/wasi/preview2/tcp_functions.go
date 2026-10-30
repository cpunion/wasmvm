//go:build linux || darwin

package preview2

import (
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var tcpMethods = []string{"start-bind", "finish-bind", "start-connect", "finish-connect", "start-listen", "finish-listen", "accept", "local-address", "remote-address", "is-listening", "address-family", "set-listen-backlog-size", "keep-alive-enabled", "set-keep-alive-enabled", "keep-alive-idle-time", "set-keep-alive-idle-time", "keep-alive-interval", "set-keep-alive-interval", "keep-alive-count", "set-keep-alive-count", "hop-limit", "set-hop-limit", "receive-buffer-size", "set-receive-buffer-size", "send-buffer-size", "set-send-buffer-size", "subscribe", "shutdown"}

func (p *TCP) contracts() map[string]component.FunctionCheck {
	code := &interfaceShape{kind: component.KindEnum, labels: networkErrorLabels}
	family := &interfaceShape{kind: component.KindEnum, labels: []string{"ipv4", "ipv6"}}
	boolean, u64, u32, u8 := scalarShape(component.KindBool), scalarShape(component.KindU64), scalarShape(component.KindU32), scalarShape(component.KindU8)
	own := resourceShape(component.KindOwn, p.resource)
	borrow := resourceShape(component.KindBorrow, p.resource)
	input, output := resourceShape(component.KindOwn, p.network.io.inputType), resourceShape(component.KindOwn, p.network.io.outputType)
	pair := aggregateShape(component.KindTuple, shapeField{"", input}, shapeField{"", output})
	result := func(ok *interfaceShape) *interfaceShape {
		return aggregateShape(component.KindResult, shapeField{"", ok}, shapeField{"", code})
	}
	out := map[string]component.FunctionCheck{"create-tcp-socket": shapeContract([]*interfaceShape{family}, result(own))}
	for _, name := range tcpMethods {
		args := []*interfaceShape{borrow}
		var value *interfaceShape
		switch name {
		case "start-bind", "start-connect":
			args = append(args, resourceShape(component.KindBorrow, p.network.networkType), socketAddressShape())
		case "finish-connect":
			value = pair
		case "accept":
			value = aggregateShape(component.KindTuple, shapeField{"", own}, shapeField{"", input}, shapeField{"", output})
		case "local-address", "remote-address":
			value = socketAddressShape()
		case "keep-alive-enabled":
			value = boolean
		case "keep-alive-idle-time", "keep-alive-interval", "receive-buffer-size", "send-buffer-size":
			value = u64
		case "keep-alive-count":
			value = u32
		case "hop-limit":
			value = u8
		case "set-keep-alive-enabled":
			args = append(args, boolean)
		case "set-listen-backlog-size", "set-keep-alive-idle-time", "set-keep-alive-interval", "set-receive-buffer-size", "set-send-buffer-size":
			args = append(args, u64)
		case "set-keep-alive-count":
			args = append(args, u32)
		case "set-hop-limit":
			args = append(args, u8)
		case "shutdown":
			args = append(args, &interfaceShape{kind: component.KindEnum, labels: []string{"receive", "send", "both"}})
		}
		value = result(value)
		switch name {
		case "is-listening":
			value = boolean
		case "address-family":
			value = family
		case "subscribe":
			value = resourceShape(component.KindOwn, p.network.io.poller.resource)
		}
		out["[method]tcp-socket."+name] = shapeContract(args, value)
	}
	return out
}
func (p *TCP) Imports() component.Imports {
	if p == nil || p.resource == nil {
		return component.Imports{}
	}
	out := p.network.Imports()
	for name, instance := range p.clock.Imports().Instances {
		if name != PollInterface {
			out.Instances[name] = instance
		}
	}
	functions := map[string]component.HostFunction{}
	checks := map[string]component.FunctionCheck{}
	contracts := p.contracts()
	for _, name := range tcpMethods {
		key := "[method]tcp-socket." + name
		functions[key] = p.method(name)
		checks[key] = contracts[key]
	}
	out.Instances[TCPInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"tcp-socket": p.resource, "network": p.network.networkType, "input-stream": p.network.io.inputType, "output-stream": p.network.io.outputType, "pollable": p.network.io.poller.resource}, Functions: functions, FunctionChecks: checks}
	out.Instances[TCPCreateInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"tcp-socket": p.resource, "network": p.network.networkType}, Functions: map[string]component.HostFunction{"create-tcp-socket": func(c *component.Caller, args []component.Value) (*component.Value, error) {
		if err := p.begin(c.Context()); err != nil {
			return nil, err
		}
		defer p.active.Done()
		value, err := tcpCreate(c, p, uint32(args[0].Bits))
		if err != nil {
			return socketFailure(err), nil
		}
		return streamOK(&value), nil
	}}, FunctionChecks: map[string]component.FunctionCheck{"create-tcp-socket": contracts["create-tcp-socket"]}}
	return out
}
func (p *TCP) method(name string) component.HostFunction {
	return func(c *component.Caller, args []component.Value) (*component.Value, error) {
		if err := p.begin(c.Context()); err != nil {
			return nil, err
		}
		defer p.active.Done()
		s, err := p.socket(c, args[0])
		if err != nil {
			return nil, err
		}
		if name == "subscribe" {
			return p.subscribe(c, s)
		}
		value, payload, err := p.execute(c, s, name, args)
		if failed := c.Context().Err(); failed != nil {
			return nil, failed
		}
		if p.ctx.Err() != nil {
			return nil, wasm.ErrClosed
		}
		if name == "is-listening" || name == "address-family" {
			return &value, err
		}
		if err != nil {
			return socketFailure(err), nil
		}
		if payload {
			return streamOK(&value), nil
		}
		return streamOK(nil), nil
	}
}
func (p *TCP) subscribe(c *component.Caller, s *tcpSocket) (*component.Value, error) {
	poller := p.network.io.poller
	if err := poller.begin(); err != nil {
		return nil, err
	}
	defer poller.active.Done()
	if err := p.child(s); err != nil {
		return nil, err
	}
	value, err := poller.publish(c, s, func() { p.release(s) })
	return &value, err
}
