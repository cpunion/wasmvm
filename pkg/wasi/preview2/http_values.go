//go:build linux || darwin

package preview2

import (
	"errors"
	"strings"

	"github.com/cpunion/wasmvm/pkg/component"
)

var httpErrorNames = []string{"DNS-timeout", "DNS-error", "destination-not-found", "destination-unavailable", "destination-IP-prohibited", "destination-IP-unroutable", "connection-refused", "connection-terminated", "connection-timeout", "connection-read-timeout", "connection-write-timeout", "connection-limit-reached", "TLS-protocol-error", "TLS-certificate-error", "TLS-alert-received", "HTTP-request-denied", "HTTP-request-length-required", "HTTP-request-body-size", "HTTP-request-method-invalid", "HTTP-request-URI-invalid", "HTTP-request-URI-too-long", "HTTP-request-header-section-size", "HTTP-request-header-size", "HTTP-request-trailer-section-size", "HTTP-request-trailer-size", "HTTP-response-incomplete", "HTTP-response-header-section-size", "HTTP-response-header-size", "HTTP-response-body-size", "HTTP-response-trailer-section-size", "HTTP-response-trailer-size", "HTTP-response-transfer-coding", "HTTP-response-content-coding", "HTTP-response-timeout", "HTTP-upgrade-failed", "HTTP-protocol-error", "loop-detected", "configuration-error", "internal-error"}
var httpMethods = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}

type httpError struct {
	value component.Value
	cause error
}

func (e *httpError) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return httpErrorNames[e.value.Case]
}
func (e *httpError) Unwrap() error { return e.cause }
func httpLower(s string) string    { return strings.ToLower(s) }
func httpSome(v component.Value) component.Value {
	return component.Value{Case: 1, Elements: []component.Value{v}}
}
func httpOK(v component.Value) component.Value {
	return component.Value{Elements: []component.Value{v}}
}
func httpUnitResult(ok bool) component.Value {
	if ok {
		return component.Value{}
	}
	return component.Value{Case: 1}
}
func httpErr(name string) *httpError {
	index := len(httpErrorNames) - 1
	for i, n := range httpErrorNames {
		if n == name {
			index = i
			break
		}
	}
	v := component.Value{Case: uint32(index)}
	switch index {
	case 1, 14, 24, 27, 30:
		v.Elements = []component.Value{{Elements: []component.Value{{}, {}}}}
	case 17, 21, 22, 23, 26, 28, 29, 31, 32, 38:
		v.Elements = []component.Value{{}}
	}
	return &httpError{value: v}
}
func httpErrorResult(err error) component.Value {
	var typed *httpError
	if !errors.As(err, &typed) {
		typed = httpErr("internal-error")
		typed.cause = err
	}
	return component.Value{Case: 1, Elements: []component.Value{typed.value}}
}
func httpMethodValue(method string) component.Value {
	for i, m := range httpMethods {
		if method == m {
			return component.Value{Case: uint32(i)}
		}
	}
	return component.Value{Case: 9, Elements: []component.Value{{Text: method}}}
}
func httpParseMethod(value component.Value) string {
	if value.Case < 9 {
		return httpMethods[value.Case]
	}
	return value.Elements[0].Text
}
func httpOptionalValue(value httpOptionalString) component.Value {
	if value.set {
		return httpSome(component.Value{Text: value.value})
	}
	return component.Value{}
}
func httpParseOptional(value component.Value) httpOptionalString {
	if value.Case == 0 {
		return httpOptionalString{}
	}
	return httpOptionalString{value.Elements[0].Text, true}
}
func httpSchemeValue(scheme httpOptionalString) component.Value {
	if !scheme.set {
		return component.Value{}
	}
	v := component.Value{}
	switch scheme.value {
	case "http":
	case "https":
		v.Case = 1
	default:
		v.Case = 2
		v.Elements = []component.Value{{Text: scheme.value}}
	}
	return httpSome(v)
}
func httpParseScheme(value component.Value) httpOptionalString {
	if value.Case == 0 {
		return httpOptionalString{}
	}
	scheme := value.Elements[0]
	s := "http"
	if scheme.Case == 1 {
		s = "https"
	} else if scheme.Case == 2 {
		s = scheme.Elements[0].Text
	}
	return httpOptionalString{s, true}
}
