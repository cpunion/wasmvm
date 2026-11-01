//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"net/netip"
	"strings"

	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/net/idna"
)

func (p *Sockets) Imports() component.Imports {
	if p == nil || p.resources == nil {
		return component.Imports{}
	}
	types := component.Imports{Resources: p.resources, Functions: map[string]component.HostFunction{}, AsyncCompletions: map[string]component.AsyncHostFunction{}, FunctionChecks: map[string]component.FunctionCheck{}}
	for key, schema := range socketsWITFunctions["types"] {
		types.FunctionChecks[key] = schema.check(p.resources)
		if schema.async {
			types.AsyncCompletions[key] = p.asyncMethod(key)
		} else {
			types.Functions[key] = p.syncMethod(key)
		}
	}
	lookup := component.Imports{AsyncCompletions: map[string]component.AsyncHostFunction{"resolve-addresses": p.lookup}, FunctionChecks: map[string]component.FunctionCheck{"resolve-addresses": socketsWITFunctions["ip-name-lookup"]["resolve-addresses"].check(p.resources)}}
	return component.Imports{Instances: map[string]component.Imports{SocketsTypesInterface: types, IPNameLookupInterface: lookup}}
}
func netMethod(key string) (string, bool) {
	udp := strings.Contains(key, "udp-socket.")
	_, name, _ := strings.Cut(key, ".")
	return name, udp
}
func (p *Sockets) syncMethod(key string) component.HostFunction {
	return func(c *component.Caller, args []component.Value) (*component.Value, error) {
		if err := p.check(c.Context()); err != nil {
			return nil, err
		}
		if err := p.begin(); err != nil {
			return nil, err
		}
		defer p.end()
		name, udp := netMethod(key)
		if name == "create" {
			s, err := p.create(uint32(args[0].Bits), udp)
			if err != nil {
				return netResult(nil, err), nil
			}
			defer s.release()
			v, err := p.publish(c, s)
			return netResult(&v, err), nil
		}
		s, err := p.socket(c, args[0], udp)
		if err != nil {
			return nil, err
		}
		defer s.release()
		switch name {
		case "listen":
			return p.listen(c, s)
		case "send", "receive":
			return p.tcpStream(c, s, name, args)
		}
		value, err := p.execute(c.Context(), s, name, args)
		if name == "get-address-family" {
			return value, err
		}
		return netResult(value, err), nil
	}
}
func (p *Sockets) asyncMethod(key string) component.AsyncHostFunction {
	return func(c *component.Caller, args []component.Value) (*component.HostCompletion, error) {
		if err := p.check(c.Context()); err != nil {
			return nil, err
		}
		name, udp := netMethod(key)
		s, err := p.socket(c, args[0], udp)
		if err != nil {
			return nil, err
		}
		if err = p.begin(); err != nil {
			s.release()
			return nil, err
		}
		callCtx := c.Context()
		ctx, cancel := context.WithCancel(callCtx)
		stop := context.AfterFunc(p.ctx, cancel)
		completion := component.NewHostCompletionWithCleanup(func() { cancel(); stop() })
		snapshot := append([]component.Value(nil), args...)
		snapshot[0].Resource = nil
		go func() {
			defer p.workers.Done()
			var value *component.Value
			var err error
			defer func() {
				panicked := recover()
				s.release()
				p.releaseSlot()
				if callCtx.Err() != nil {
					completion.Cancelled()
					return
				}
				completion.Resolve(func(*component.Caller) (*component.Value, error) {
					if panicked != nil {
						panic(panicked)
					}
					return netResult(value, err), nil
				})
			}()
			value, err = p.executeAsync(ctx, s, name, snapshot)
		}()
		return completion, nil
	}
}
func (p *Sockets) authorize(ctx context.Context, address SocketAddress, fn func(context.Context, SocketAddress) error) (err error) {
	if fn == nil {
		return netError("access-denied")
	}
	joint, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(p.ctx, cancel)
	defer cancel()
	defer stop()
	defer func() {
		if recover() != nil {
			err = netError("access-denied")
		}
	}()
	if err = fn(joint, address); err != nil {
		return netError("access-denied")
	}
	return p.check(joint)
}
func (p *Sockets) lookup(c *component.Caller, args []component.Value) (*component.HostCompletion, error) {
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	if err := p.begin(); err != nil {
		return nil, err
	}
	callCtx := c.Context()
	ctx, cancel := context.WithTimeout(callCtx, p.config.ResolveTimeout)
	stop := context.AfterFunc(p.ctx, cancel)
	completion := component.NewHostCompletionWithCleanup(func() { cancel(); stop() })
	name := args[0].Text
	go func() {
		defer p.workers.Done()
		var addresses []netip.Addr
		var err error
		defer func() {
			if recover() != nil {
				err = errors.New("resolver panic")
			}
			p.releaseSlot()
			if callCtx.Err() != nil {
				completion.Cancelled()
				return
			}
			completion.Resolve(func(*component.Caller) (*component.Value, error) { return lookupResult(addresses, err), nil })
		}()
		if uint64(len(name)) > uint64(p.config.MaxNameBytes) || strings.IndexByte(name, 0) >= 0 || name == "" {
			err = netError("invalid-argument")
			return
		}
		if ip, failed := netip.ParseAddr(name); failed == nil {
			if ip.Zone() != "" {
				err = netError("invalid-argument")
				return
			}
			addresses = []netip.Addr{ip.Unmap()}
			return
		}
		if !p.config.AllowNameLookup {
			err = netError("access-denied")
			return
		}
		name, err = idna.Lookup.ToASCII(name)
		if err != nil {
			err = netError("invalid-argument")
			return
		}
		if uint64(len(name)) > uint64(p.config.MaxNameBytes) {
			err = netError("invalid-argument")
			return
		}
		var results []netip.Addr
		results, err = p.config.Resolve(ctx, name)
		if err != nil {
			return
		}
		if err = p.check(ctx); err != nil {
			return
		}
		seen := map[netip.Addr]bool{}
		for _, ip := range results {
			if !ip.IsValid() || ip.Zone() != "" {
				continue
			}
			ip = ip.Unmap()
			if !seen[ip] {
				if uint64(len(addresses)) >= uint64(p.config.MaxAddresses) {
					err = errors.New("resolver address limit")
					addresses = nil
					return
				}
				seen[ip] = true
				addresses = append(addresses, ip)
			}
		}
		if len(addresses) == 0 {
			err = netError("name-unresolvable")
		}
	}()
	return completion, nil
}
