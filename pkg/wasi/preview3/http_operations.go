//go:build linux || darwin

package preview3

import (
	"errors"
	"net/url"
	"strings"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (p *HTTP) method(name string) component.HostFunction {
	return func(c *component.Caller, args []component.Value) (*component.Value, error) {
		if err := p.begin(c.Context()); err != nil {
			return nil, err
		}
		defer p.active.Done()
		v, err := p.execute(c, name, args)
		if err == nil {
			err = p.check(c.Context())
		}
		return v, err
	}
}
func httpHeaderFailure(err error) component.Value {
	code := uint32(4)
	var typed host.HTTPHeaderError
	if errors.As(err, &typed) {
		code = uint32(typed)
	} else if errors.Is(err, wasm.ErrLimit) {
		code = 3
	}
	v := component.Value{Case: code}
	if code == 4 {
		v.Elements = []component.Value{{}}
	}
	return component.Value{Case: 1, Elements: []component.Value{v}}
}
func httpFieldsValue(fields []host.HTTPField) component.Value {
	v := component.Value{}
	for _, f := range fields {
		v.Elements = append(v.Elements, component.Value{Elements: []component.Value{{Text: f.Name}, {Text: f.Value}}})
	}
	return v
}
func httpStringsValue(values []string) component.Value {
	v := component.Value{}
	for _, s := range values {
		v.Elements = append(v.Elements, component.Value{Text: s})
	}
	return v
}
func httpCopyOptions(options [3]*uint64) [3]*uint64 {
	var out [3]*uint64
	for i, v := range options {
		if v != nil {
			n := *v
			out[i] = &n
		}
	}
	return out
}
func (p *HTTP) execute(c *component.Caller, key string, args []component.Value) (*component.Value, error) {
	switch key {
	case "[constructor]fields":
		fields, err := host.NewHTTPFields(p.fields, nil, false)
		if err != nil {
			return nil, err
		}
		v, err := p.publish(c, &httpEntry{kind: "fields", fields: fields})
		return &v, err
	case "[static]fields.from-list":
		if uint64(len(args[0].Elements)) > uint64(p.config.MaxHeaderFields) {
			v := httpHeaderFailure(wasm.ErrLimit)
			return &v, nil
		}
		fields := make([]host.HTTPField, 0, len(args[0].Elements))
		for _, pair := range args[0].Elements {
			text, err := pair.Elements[1].ByteString()
			if err != nil {
				return nil, err
			}
			fields = append(fields, host.HTTPField{Name: pair.Elements[0].Text, Value: text})
		}
		value, err := host.NewHTTPFields(p.fields, fields, false)
		if err != nil {
			v := httpHeaderFailure(err)
			return &v, nil
		}
		v, err := p.publish(c, &httpEntry{kind: "fields", fields: value})
		r := httpOK(v)
		return &r, err
	case "[constructor]request-options":
		v, err := p.publish(c, &httpEntry{kind: "request-options"})
		return &v, err
	case "[static]request.new":
		return p.newMessage(c, "request", args)
	case "[static]response.new":
		return p.newMessage(c, "response", args)
	case "[static]request.consume-body":
		return p.consumeBody(c, "request", args)
	case "[static]response.consume-body":
		return p.consumeBody(c, "response", args)
	}
	rest := strings.TrimPrefix(key, "[method]")
	kind, name, ok := strings.Cut(rest, ".")
	if !ok {
		return nil, vm.ErrType
	}
	e, err := p.entry(c, args[0], kind)
	if err != nil {
		return nil, err
	}
	if kind == "fields" {
		return p.fieldMethod(c, e, name, args[1:])
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	var v component.Value
	if kind == "request-options" {
		if name == "clone" {
			v, err = p.publish(c, &httpEntry{kind: kind, options: httpCopyOptions(e.options)})
			return &v, err
		}
		index := 0
		if strings.Contains(name, "first-byte") {
			index = 1
		} else if strings.Contains(name, "between-bytes") {
			index = 2
		}
		if strings.HasPrefix(name, "get-") {
			if e.options[index] != nil {
				v = httpSome(component.Value{Bits: *e.options[index]})
			}
			return &v, nil
		}
		if e.immutable {
			return &component.Value{Case: 1, Elements: []component.Value{{Case: 1}}}, nil
		}
		e.options[index] = nil
		if args[1].Case == 1 {
			n := args[1].Elements[0].Bits
			e.options[index] = &n
		}
		return &v, nil
	}
	message := e.message
	switch name {
	case "get-headers":
		v, err = p.publish(c, &httpEntry{kind: "fields", fields: message.headers})
	case "get-options":
		if message.options != nil {
			var owned component.Value
			owned, err = p.publish(c, &httpEntry{kind: "request-options", options: httpCopyOptions(*message.options), immutable: true, parent: e})
			v = httpSome(owned)
		}
	case "get-method":
		v = httpMethodValue(message.method)
	case "set-method":
		method := httpParseMethod(args[1])
		if !host.HTTPToken(method) {
			v.Case = 1
		} else {
			message.method = method
		}
	case "get-path-with-query":
		v = httpOptionalValue(message.path)
	case "set-path-with-query":
		path := httpParseOptional(args[1])
		if path != nil && *path != "" {
			text := *path
			if strings.HasPrefix(text, "?") {
				text = "/" + text
			}
			parsed, failed := url.ParseRequestURI(text)
			if failed != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Fragment != "" {
				v.Case = 1
			}
		}
		if v.Case == 0 {
			message.path = path
		}
	case "get-scheme":
		v = httpSchemeValue(message.scheme)
	case "set-scheme":
		scheme := httpParseScheme(args[1])
		if scheme != nil {
			parsed, failed := url.Parse(*scheme + ":")
			if failed != nil || parsed.Scheme == "" || parsed.Scheme != *scheme {
				v.Case = 1
			}
		}
		if v.Case == 0 {
			message.scheme = scheme
		}
	case "get-authority":
		v = httpOptionalValue(message.authority)
	case "set-authority":
		authority := httpParseOptional(args[1])
		if authority != nil {
			parsed, failed := url.Parse("http://" + *authority)
			if failed != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
				v.Case = 1
			}
		}
		if v.Case == 0 {
			message.authority = authority
		}
	case "get-status-code":
		v.Bits = uint64(message.status)
	case "set-status-code":
		if args[1].Bits < 100 || args[1].Bits > 999 {
			v.Case = 1
		} else {
			message.status = uint16(args[1].Bits)
		}
	default:
		return nil, vm.ErrType
	}
	return &v, err
}
func (p *HTTP) fieldMethod(c *component.Caller, e *httpEntry, name string, args []component.Value) (*component.Value, error) {
	var v component.Value
	switch name {
	case "get":
		v = httpStringsValue(e.fields.Values(args[0].Text))
	case "has":
		if e.fields.Has(args[0].Text) {
			v.Bits = 1
		}
	case "copy-all":
		v = httpFieldsValue(e.fields.Entries())
	case "clone":
		returnValue, err := p.publish(c, &httpEntry{kind: "fields", fields: e.fields.Clone()})
		return &returnValue, err
	case "set", "append", "delete", "get-and-delete":
		var err error
		if name == "get-and-delete" {
			var values []string
			values, err = e.fields.GetAndDelete(args[0].Text)
			if err == nil {
				v = httpOK(httpStringsValue(values))
			}
		} else {
			var values []string
			if name == "append" {
				text, failed := args[1].ByteString()
				if failed != nil {
					return nil, failed
				}
				values = []string{text}
			} else if name == "set" {
				for _, value := range args[1].Elements {
					text, failed := value.ByteString()
					if failed != nil {
						return nil, failed
					}
					values = append(values, text)
				}
			}
			err = e.fields.Mutate(args[0].Text, values, name == "append")
		}
		if err != nil {
			v = httpHeaderFailure(err)
		}
	default:
		return nil, vm.ErrType
	}
	return &v, nil
}
