package preview2

import (
	"errors"
	"net/netip"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func ipAddressShape() *interfaceShape {
	v4, v6 := make([]shapeField, 4), make([]shapeField, 8)
	for i := range v4 {
		v4[i] = shapeField{"", scalarShape(component.KindU8)}
	}
	for i := range v6 {
		v6[i] = shapeField{"", scalarShape(component.KindU16)}
	}
	return aggregateShape(component.KindVariant, shapeField{"ipv4", aggregateShape(component.KindTuple, v4...)}, shapeField{"ipv6", aggregateShape(component.KindTuple, v6...)})
}
func ipAddressValue(address netip.Addr) component.Value {
	if address.Is4() {
		raw := address.As4()
		value := component.Value{Elements: make([]component.Value, 4)}
		for i, v := range raw {
			value.Elements[i].Bits = uint64(v)
		}
		return value
	}
	raw := address.As16()
	value := component.Value{Case: 1, Elements: make([]component.Value, 8)}
	for i := range value.Elements {
		value.Elements[i].Bits = uint64(raw[i*2])<<8 | uint64(raw[i*2+1])
	}
	return value
}
func ipVariantValue(address netip.Addr) component.Value {
	value := ipAddressValue(address)
	return component.Value{Case: value.Case, Elements: []component.Value{{Elements: value.Elements}}}
}

// Imports provides the complete network/error, instance-network and IP lookup
// interfaces. TCP/UDP and HTTP remain separate, explicitly authorized providers.
func (p *Network) Imports() component.Imports {
	if p == nil || p.networkType == nil {
		return component.Imports{}
	}
	out := p.io.Imports()
	code := &interfaceShape{kind: component.KindEnum, labels: networkErrorLabels}
	result := func(ok *interfaceShape) *interfaceShape {
		return aggregateShape(component.KindResult, shapeField{"", ok}, shapeField{"", code})
	}
	ownStream := resourceShape(component.KindOwn, p.streamType)
	borrowStream := resourceShape(component.KindBorrow, p.streamType)
	network := component.Imports{Resources: map[string]*component.HostResourceType{"network": p.networkType, "error": p.io.errors.resource}, Functions: map[string]component.HostFunction{"network-error-code": p.errorCode}, FunctionChecks: map[string]component.FunctionCheck{"network-error-code": shapeContract([]*interfaceShape{resourceShape(component.KindBorrow, p.io.errors.resource)}, elementShape(component.KindOption, code))}}
	out.Instances[NetworkInterface] = network
	out.Instances[InstanceNetworkInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"network": p.networkType}, Functions: map[string]component.HostFunction{"instance-network": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
		if err := p.begin(); err != nil {
			return nil, err
		}
		defer p.active.Done()
		if err := p.check(c.Context()); err != nil {
			return nil, err
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.closed {
			return nil, wasm.ErrClosed
		}
		value, err := c.NewResource(p.networkType, 0)
		return &value, err
	}}, FunctionChecks: map[string]component.FunctionCheck{"instance-network": shapeContract(nil, resourceShape(component.KindOwn, p.networkType))}}
	const next = "[method]resolve-address-stream.resolve-next-address"
	const subscribe = "[method]resolve-address-stream.subscribe"
	out.Instances[IPNameLookupInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"network": p.networkType, "resolve-address-stream": p.streamType, "pollable": p.io.poller.resource}, Functions: map[string]component.HostFunction{"resolve-addresses": p.resolve, next: p.nextAddress, subscribe: p.subscribe}, FunctionChecks: map[string]component.FunctionCheck{
		"resolve-addresses": shapeContract([]*interfaceShape{resourceShape(component.KindBorrow, p.networkType), scalarShape(component.KindString)}, result(ownStream)),
		next:                shapeContract([]*interfaceShape{borrowStream}, result(elementShape(component.KindOption, ipAddressShape()))),
		subscribe:           shapeContract([]*interfaceShape{borrowStream}, resourceShape(component.KindOwn, p.io.poller.resource)),
	}}
	return out
}
func (p *Network) errorCode(c *component.Caller, args []component.Value) (*component.Value, error) {
	info, err := p.io.errors.Describe(c, args[0].Resource)
	if err != nil {
		return nil, err
	}
	out := &component.Value{}
	var cause *NetworkError
	if errors.As(info.Cause, &cause) {
		out.Case = 1
		out.Elements = []component.Value{{Bits: uint64(cause.Code)}}
	}
	return out, nil
}
func (p *Network) nextAddress(c *component.Caller, args []component.Value) (*component.Value, error) {
	if err := p.begin(); err != nil {
		return nil, err
	}
	defer p.active.Done()
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	s, err := p.stream(c, args[0])
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dropped {
		return nil, wasm.ErrClosed
	}
	if !s.done {
		return networkFailure(NetworkWouldBlock), nil
	}
	if s.failed {
		return networkFailure(s.code), nil
	}
	value := component.Value{}
	if s.at < len(s.addresses) {
		value.Case = 1
		value.Elements = []component.Value{ipVariantValue(s.addresses[s.at])}
		s.at++
	}
	return streamOK(&value), nil
}
func (p *Network) subscribe(c *component.Caller, args []component.Value) (*component.Value, error) {
	if err := p.begin(); err != nil {
		return nil, err
	}
	defer p.active.Done()
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	s, err := p.stream(c, args[0])
	if err != nil {
		return nil, err
	}
	if err = p.io.poller.begin(); err != nil {
		return nil, err
	}
	defer p.io.poller.active.Done()
	s.mu.Lock()
	if s.dropped {
		s.mu.Unlock()
		return nil, wasm.ErrClosed
	}
	s.children++
	s.mu.Unlock()
	release := func() {
		s.mu.Lock()
		s.children--
		orphan := s.dropped && s.children == 0
		s.mu.Unlock()
		if orphan {
			p.drop(s.id)
		}
	}
	value, err := p.io.poller.publish(c, s, release)
	return &value, err
}
