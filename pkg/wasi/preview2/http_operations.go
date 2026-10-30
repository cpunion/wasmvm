//go:build linux || darwin

package preview2

import (
	"errors"
	"net/url"
	"strings"

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
		value, err := p.execute(c, name, args)
		if failed := c.Context().Err(); failed != nil {
			return nil, failed
		}
		if p.ctx.Err() != nil {
			return nil, wasm.ErrClosed
		}
		return value, err
	}
}
func (p *HTTP) execute(c *component.Caller, name string, args []component.Value) (*component.Value, error) {
	var value component.Value
	var err error
	switch name {
	case "handle":
		value, err = p.sendRequest(c, args)
	case "[constructor]fields":
		var fields *httpFields
		fields, err = newHTTPFields(p.fields, nil, false)
		if err == nil {
			value, err = p.publish(c, &httpEntry{kind: httpFieldsKind, fields: fields})
		}
	case "[static]fields.from-list":
		if uint64(len(args[0].Elements)) > uint64(p.config.MaxHeaderFields) {
			return nil, wasm.ErrLimit
		}
		var fields []httpField
		for _, pair := range args[0].Elements {
			text, failed := pair.Elements[1].ByteString()
			if failed != nil {
				return nil, failed
			}
			fields = append(fields, httpField{pair.Elements[0].Text, text})
		}
		var data *httpFields
		data, err = newHTTPFields(p.fields, fields, false)
		if err == nil {
			value, err = p.publish(c, &httpEntry{kind: httpFieldsKind, fields: data})
			value = httpOK(value)
		} else {
			var header httpHeaderError
			if errors.As(err, &header) {
				value = httpHeaderFailure(header)
				err = nil
			}
		}
	case "[constructor]outgoing-request", "[constructor]outgoing-response":
		kind := httpOutgoingRequest
		if name == "[constructor]outgoing-response" {
			kind = httpOutgoingResponse
		}
		var fields *httpEntry
		fields, err = p.take(c, args[0], httpFieldsKind)
		if err != nil {
			return nil, err
		}
		fields.fields.mu.Lock()
		fields.fields.immutable = true
		fields.fields.mu.Unlock()
		var body *httpBody
		body, err = p.newBody(true)
		if err != nil {
			return nil, err
		}
		length, _ := httpContentLength(fields.fields)
		body.expected = length
		if kind == httpOutgoingResponse {
			body.errorName = "HTTP-response-body-size"
		}
		value, err = p.publish(c, &httpEntry{kind: kind, message: &httpMessage{headers: fields.fields, method: "GET", status: 200, body: body}})
		if err != nil {
			body.close()
		}
	case "[constructor]request-options":
		value, err = p.publish(c, &httpEntry{kind: httpRequestOptions})
	case "[static]incoming-body.finish":
		var entry *httpEntry
		entry, err = p.take(c, args[0], httpIncomingBody)
		if err != nil {
			return nil, err
		}
		entry.body.buffer.discard()
		value, err = p.publish(c, &httpEntry{kind: httpFutureTrailers, body: entry.body})
		if err != nil {
			entry.body.close()
		}
	case "[static]outgoing-body.finish":
		value, err = p.finishBody(c, args)
	case "[static]response-outparam.set":
		err = p.setResponse(c, args)
		return nil, err
	case "http-error-code":
		var info IOErrorInfo
		info, err = p.network.io.errors.Describe(c, args[0].Resource)
		var code *httpError
		if err == nil && errors.As(info.Cause, &code) {
			value = httpSome(code.value)
		}
	default:
		if !strings.HasPrefix(name, "[method]") {
			return nil, vm.ErrType
		}
		rest := strings.TrimPrefix(name, "[method]")
		dot := strings.IndexByte(rest, '.')
		if dot < 0 {
			return nil, vm.ErrType
		}
		resource, method := rest[:dot], rest[dot+1:]
		kind := -1
		for i, n := range httpResourceNames {
			if n == resource {
				kind = i
				break
			}
		}
		if kind < 0 {
			return nil, vm.ErrType
		}
		var entry *httpEntry
		entry, err = p.entry(c, args[0], kind)
		if err != nil {
			return nil, err
		}
		switch kind {
		case httpFieldsKind:
			value, err = p.fieldMethod(c, entry, method, args[1:])
		case httpIncomingRequest, httpOutgoingRequest, httpIncomingResponse, httpOutgoingResponse:
			value, err = p.messageMethod(c, entry, method, args[1:])
		case httpRequestOptions:
			value, err = p.optionMethod(entry, method, args[1:])
		case httpIncomingBody, httpOutgoingBody:
			value, err = p.bodyPort(c, entry)
		case httpFutureTrailers, httpFutureResponse:
			if method == "subscribe" {
				source := PollSource(entry.body)
				if kind == httpFutureResponse {
					source = entry.future
				}
				return p.subscribe(c, entry, source)
			}
			value, err = p.futureGet(c, entry)
		default:
			err = vm.ErrType
		}
	}
	return &value, err
}
func httpHeaderFailure(code httpHeaderError) component.Value {
	return component.Value{Case: 1, Elements: []component.Value{{Case: uint32(code)}}}
}
func (p *HTTP) fieldMethod(c *component.Caller, entry *httpEntry, name string, args []component.Value) (component.Value, error) {
	fields := entry.fields
	switch name {
	case "get":
		var values []component.Value
		for _, s := range fields.values(args[0].Text) {
			values = append(values, component.Value{Text: s})
		}
		return component.Value{Elements: values}, nil
	case "has":
		v := component.Value{}
		if fields.has(args[0].Text) {
			v.Bits = 1
		}
		return v, nil
	case "entries":
		var values []component.Value
		for _, field := range fields.entries() {
			values = append(values, component.Value{Elements: []component.Value{{Text: field.name}, {Text: field.value}}})
		}
		return component.Value{Elements: values}, nil
	case "clone":
		return p.publish(c, &httpEntry{kind: httpFieldsKind, fields: fields.clone()})
	}
	var values []string
	switch name {
	case "set":
		if uint64(len(args[1].Elements)) > uint64(p.config.MaxHeaderFields) {
			return component.Value{}, wasm.ErrLimit
		}
		for _, value := range args[1].Elements {
			s, err := value.ByteString()
			if err != nil {
				return component.Value{}, err
			}
			values = append(values, s)
		}
	case "append":
		s, err := args[1].ByteString()
		if err != nil {
			return component.Value{}, err
		}
		values = []string{s}
	}
	err := fields.mutate(args[0].Text, values, name == "append")
	var header httpHeaderError
	if errors.As(err, &header) {
		return httpHeaderFailure(header), nil
	}
	return component.Value{}, err
}
func (p *HTTP) messageMethod(c *component.Caller, entry *httpEntry, name string, args []component.Value) (component.Value, error) {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	m := entry.message
	switch name {
	case "headers":
		return p.publish(c, &httpEntry{kind: httpFieldsKind, parent: entry, fields: m.headers})
	case "method":
		return httpMethodValue(m.method), nil
	case "path-with-query":
		return httpOptionalValue(m.path), nil
	case "scheme":
		return httpSchemeValue(m.scheme), nil
	case "authority":
		return httpOptionalValue(m.authority), nil
	case "status", "status-code":
		return component.Value{Bits: uint64(m.status)}, nil
	case "consume", "body":
		if m.bodyTaken {
			return httpUnitResult(false), nil
		}
		kind := httpIncomingBody
		if name == "body" {
			kind = httpOutgoingBody
		}
		v, err := p.publish(c, &httpEntry{kind: kind, body: m.body})
		if err != nil {
			return component.Value{}, err
		}
		m.bodyTaken = true
		return httpOK(v), nil
	case "set-status-code":
		if args[0].Bits < 100 || args[0].Bits > 999 {
			return httpUnitResult(false), nil
		}
		m.status = uint16(args[0].Bits)
	case "set-method":
		method := httpParseMethod(args[0])
		if len(method) > 256 || !httpToken(method) {
			return httpUnitResult(false), nil
		}
		m.method = method
	case "set-path-with-query":
		path := httpParseOptional(args[0])
		if len(path.value) > int(p.config.MaxHeaderBytes) {
			return httpUnitResult(false), nil
		}
		if path.set && path.value != "" {
			if !httpURIASCII(path.value) {
				return httpUnitResult(false), nil
			}
			parsedPath := path.value
			if strings.HasPrefix(parsedPath, "?") {
				parsedPath = "/" + parsedPath
			}
			parsed, err := url.ParseRequestURI(parsedPath)
			if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Fragment != "" {
				return httpUnitResult(false), nil
			}
		}
		m.path = path
	case "set-scheme":
		scheme := httpParseScheme(args[0])
		if scheme.set && !httpValidScheme(scheme.value) {
			return httpUnitResult(false), nil
		}
		m.scheme = scheme
	case "set-authority":
		authority := httpParseOptional(args[0])
		if len(authority.value) > int(p.config.MaxHeaderBytes) {
			return httpUnitResult(false), nil
		}
		if authority.set {
			if !httpURIASCII(authority.value) {
				return httpUnitResult(false), nil
			}
			u, err := url.Parse("http://" + authority.value)
			if err != nil || u.Hostname() == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
				return httpUnitResult(false), nil
			}
		}
		m.authority = authority
	default:
		return component.Value{}, vm.ErrType
	}
	return component.Value{}, nil
}
func httpValidScheme(s string) bool {
	if len(s) == 0 || len(s) > 256 || !(s[0] >= 'A' && s[0] <= 'Z' || s[0] >= 'a' && s[0] <= 'z') {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}
func (p *HTTP) optionMethod(entry *httpEntry, name string, args []component.Value) (component.Value, error) {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	index := 0
	if strings.Contains(name, "first-byte") {
		index = 1
	} else if strings.Contains(name, "between-bytes") {
		index = 2
	}
	if strings.HasPrefix(name, "set-") {
		var value *uint64
		if args[0].Case == 1 {
			n := args[0].Elements[0].Bits
			if n > 1<<63-1 {
				return httpUnitResult(false), nil
			}
			value = &n
		}
		entry.options[index] = value
		return component.Value{}, nil
	}
	if entry.options[index] == nil {
		return component.Value{}, nil
	}
	return httpSome(component.Value{Bits: *entry.options[index]}), nil
}
func (p *HTTP) finishBody(c *component.Caller, args []component.Value) (component.Value, error) {
	entry, err := p.take(c, args[0], httpOutgoingBody)
	if err != nil {
		return component.Value{}, err
	}
	var trailers *httpFields
	if args[1].Case == 1 {
		fields, failed := p.take(c, args[1].Elements[0], httpFieldsKind)
		if failed != nil {
			entry.body.close()
			return component.Value{}, failed
		}
		trailers = fields.fields
		for _, field := range trailers.entries() {
			switch httpLower(field.name) {
			case "host", "content-length", "transfer-encoding", "trailer":
				failure := httpErr("HTTP-protocol-error")
				entry.body.finish(failure, nil)
				return httpErrorResult(failure), nil
			}
		}
	}
	body := entry.body
	body.buffer.mu.Lock()
	written := body.buffer.written
	body.buffer.mu.Unlock()
	if body.expected != nil && written != *body.expected {
		failure := httpErr(body.errorName)
		failure.value.Elements = []component.Value{httpSome(component.Value{Bits: written})}
		body.finish(failure, nil)
		return httpErrorResult(failure), nil
	}
	body.finish(nil, trailers)
	return component.Value{}, nil
}
func (p *HTTP) futureGet(c *component.Caller, e *httpEntry) (component.Value, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.kind == httpFutureResponse {
		f := e.future
		f.mu.Lock()
		defer f.mu.Unlock()
		if !f.ready {
			return component.Value{}, nil
		}
		if f.used {
			return httpSome(httpUnitResult(false)), nil
		}
		if f.failure != nil {
			f.used = true
			if f.cancel != nil {
				f.cancel()
				f.cancel = nil
			}
			return httpSome(httpOK(httpErrorResult(f.failure))), nil
		}
		response, err := p.publish(c, &httpEntry{kind: httpIncomingResponse, message: f.response})
		if err != nil {
			return component.Value{}, err
		}
		f.response.body.mu.Lock()
		f.response.body.externalCancel = f.cancel
		f.response.body.mu.Unlock()
		f.cancel = nil
		f.used = true
		return httpSome(httpOK(httpOK(response))), nil
	}
	b := e.body
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.complete {
		return component.Value{}, nil
	}
	if e.moved {
		return httpSome(httpUnitResult(false)), nil
	}
	var result component.Value
	if b.failure != nil {
		result = httpErrorResult(b.failure)
	} else {
		var value component.Value
		if b.trailers != nil && len(b.trailers.entries()) != 0 {
			fields, err := p.publish(c, &httpEntry{kind: httpFieldsKind, parent: e, fields: b.trailers})
			if err != nil {
				return component.Value{}, err
			}
			value = httpSome(fields)
		}
		result = httpOK(value)
	}
	e.moved = true // trailers future still owns its body; destroy checks kind below
	return httpSome(httpOK(result)), nil
}

func httpURIASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] <= 32 || s[i] >= 127 {
			return false
		}
	}
	return true
}
