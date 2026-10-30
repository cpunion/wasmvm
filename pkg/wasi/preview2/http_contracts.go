//go:build linux || darwin

package preview2

import "github.com/cpunion/wasmvm/pkg/component"

func httpErrorShape() *interfaceShape {
	text := scalarShape(component.KindString)
	option := func(v *interfaceShape) *interfaceShape { return elementShape(component.KindOption, v) }
	field := aggregateShape(component.KindRecord, shapeField{"field-name", option(text)}, shapeField{"field-size", option(scalarShape(component.KindU32))})
	cases := make([]shapeField, len(httpErrorNames))
	for i, name := range httpErrorNames {
		cases[i].name = name
		switch i {
		case 1:
			cases[i].typ = aggregateShape(component.KindRecord, shapeField{"rcode", option(text)}, shapeField{"info-code", option(scalarShape(component.KindU16))})
		case 14:
			cases[i].typ = aggregateShape(component.KindRecord, shapeField{"alert-id", option(scalarShape(component.KindU8))}, shapeField{"alert-message", option(text)})
		case 17, 28:
			cases[i].typ = option(scalarShape(component.KindU64))
		case 21, 23, 26, 29:
			cases[i].typ = option(scalarShape(component.KindU32))
		case 22:
			cases[i].typ = option(field)
		case 24, 27, 30:
			cases[i].typ = field
		case 31, 32, 38:
			cases[i].typ = option(text)
		}
	}
	return aggregateShape(component.KindVariant, cases...)
}
func (p *HTTP) contracts() map[string]component.FunctionCheck {
	own := func(k int) *interfaceShape { return resourceShape(component.KindOwn, p.types[k]) }
	borrow := func(k int) *interfaceShape { return resourceShape(component.KindBorrow, p.types[k]) }
	list := func(s *interfaceShape) *interfaceShape { return elementShape(component.KindList, s) }
	option := func(s *interfaceShape) *interfaceShape { return elementShape(component.KindOption, s) }
	result := func(ok, err *interfaceShape) *interfaceShape {
		return aggregateShape(component.KindResult, shapeField{"", ok}, shapeField{"", err})
	}
	text := scalarShape(component.KindString)
	bytes := list(scalarShape(component.KindU8))
	u64 := scalarShape(component.KindU64)
	pairs := list(aggregateShape(component.KindTuple, shapeField{"", text}, shapeField{"", bytes}))
	headerError := aggregateShape(component.KindVariant, shapeField{"invalid-syntax", nil}, shapeField{"forbidden", nil}, shapeField{"immutable", nil})
	code := httpErrorShape()
	checks := map[string]component.FunctionCheck{}
	add := func(name string, args []*interfaceShape, ret *interfaceShape) {
		checks[name] = shapeContract(args, ret)
	}
	method := func(k int, name string, args []*interfaceShape, ret *interfaceShape) {
		add("[method]"+httpResourceNames[k]+"."+name, append([]*interfaceShape{borrow(k)}, args...), ret)
	}
	add("[constructor]fields", nil, own(httpFieldsKind))
	add("[static]fields.from-list", []*interfaceShape{pairs}, result(own(httpFieldsKind), headerError))
	method(httpFieldsKind, "get", []*interfaceShape{text}, list(bytes))
	method(httpFieldsKind, "has", []*interfaceShape{text}, scalarShape(component.KindBool))
	method(httpFieldsKind, "set", []*interfaceShape{text, list(bytes)}, result(nil, headerError))
	method(httpFieldsKind, "append", []*interfaceShape{text, bytes}, result(nil, headerError))
	method(httpFieldsKind, "delete", []*interfaceShape{text}, result(nil, headerError))
	method(httpFieldsKind, "entries", nil, pairs)
	method(httpFieldsKind, "clone", nil, own(httpFieldsKind))
	methodShape := aggregateShape(component.KindVariant, shapeField{"get", nil}, shapeField{"head", nil}, shapeField{"post", nil}, shapeField{"put", nil}, shapeField{"delete", nil}, shapeField{"connect", nil}, shapeField{"options", nil}, shapeField{"trace", nil}, shapeField{"patch", nil}, shapeField{"other", text})
	scheme := option(aggregateShape(component.KindVariant, shapeField{"HTTP", nil}, shapeField{"HTTPS", nil}, shapeField{"other", text}))
	for _, k := range []int{httpIncomingRequest, httpOutgoingRequest, httpIncomingResponse, httpOutgoingResponse} {
		if k == httpOutgoingRequest || k == httpOutgoingResponse {
			add("[constructor]"+httpResourceNames[k], []*interfaceShape{own(httpFieldsKind)}, own(k))
		}
		method(k, "headers", nil, own(httpFieldsKind))
		if k == httpIncomingRequest || k == httpOutgoingRequest {
			method(k, "method", nil, methodShape)
			method(k, "path-with-query", nil, option(text))
			method(k, "scheme", nil, scheme)
			method(k, "authority", nil, option(text))
			if k == httpOutgoingRequest {
				for name, typ := range map[string]*interfaceShape{"method": methodShape, "path-with-query": option(text), "scheme": scheme, "authority": option(text)} {
					method(k, "set-"+name, []*interfaceShape{typ}, result(nil, nil))
				}
			}
		} else {
			name := "status"
			if k == httpOutgoingResponse {
				name = "status-code"
				method(k, "set-status-code", []*interfaceShape{scalarShape(component.KindU16)}, result(nil, nil))
			}
			method(k, name, nil, scalarShape(component.KindU16))
		}
		body, name := httpIncomingBody, "consume"
		if k == httpOutgoingRequest || k == httpOutgoingResponse {
			body, name = httpOutgoingBody, "body"
		}
		method(k, name, nil, result(own(body), nil))
	}
	add("[constructor]request-options", nil, own(httpRequestOptions))
	for _, name := range []string{"connect-timeout", "first-byte-timeout", "between-bytes-timeout"} {
		method(httpRequestOptions, name, nil, option(u64))
		method(httpRequestOptions, "set-"+name, []*interfaceShape{option(u64)}, result(nil, nil))
	}
	add("[static]response-outparam.set", []*interfaceShape{own(httpResponseOutparam), result(own(httpOutgoingResponse), code)}, nil)
	method(httpIncomingBody, "stream", nil, result(resourceShape(component.KindOwn, p.network.io.inputType), nil))
	add("[static]incoming-body.finish", []*interfaceShape{own(httpIncomingBody)}, own(httpFutureTrailers))
	method(httpOutgoingBody, "write", nil, result(resourceShape(component.KindOwn, p.network.io.outputType), nil))
	add("[static]outgoing-body.finish", []*interfaceShape{own(httpOutgoingBody), option(own(httpFieldsKind))}, result(nil, code))
	for _, k := range []int{httpFutureTrailers, httpFutureResponse} {
		method(k, "subscribe", nil, resourceShape(component.KindOwn, p.network.io.poller.resource))
		value := own(httpIncomingResponse)
		if k == httpFutureTrailers {
			value = option(own(httpFieldsKind))
		}
		method(k, "get", nil, option(result(result(value, code), nil)))
	}
	add("http-error-code", []*interfaceShape{resourceShape(component.KindBorrow, p.network.io.errors.resource)}, option(code))
	add("handle", []*interfaceShape{own(httpOutgoingRequest), option(own(httpRequestOptions))}, result(own(httpFutureResponse), code))
	return checks
}
func (p *HTTP) Imports() component.Imports {
	if p == nil || p.types[0] == nil {
		return component.Imports{}
	}
	out := p.network.io.Imports()
	for name, value := range p.clock.Imports().Instances {
		out.Instances[name] = value
	}
	resources := map[string]*component.HostResourceType{"headers": p.types[0], "trailers": p.types[0], "input-stream": p.network.io.inputType, "output-stream": p.network.io.outputType, "pollable": p.network.io.poller.resource, "io-error": p.network.io.errors.resource}
	for i, name := range httpResourceNames {
		resources[name] = p.types[i]
	}
	functions := map[string]component.HostFunction{}
	checks := map[string]component.FunctionCheck{}
	contracts := p.contracts()
	for name, check := range contracts {
		if name != "handle" {
			functions[name] = p.method(name)
			checks[name] = check
		}
	}
	out.Instances[HTTPTypesInterface] = component.Imports{Resources: resources, Functions: functions, FunctionChecks: checks}
	out.Instances[HTTPOutgoingInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"outgoing-request": p.types[httpOutgoingRequest], "request-options": p.types[httpRequestOptions], "future-incoming-response": p.types[httpFutureResponse]}, Functions: map[string]component.HostFunction{"handle": p.method("handle")}, FunctionChecks: map[string]component.FunctionCheck{"handle": contracts["handle"]}}
	return out
}
