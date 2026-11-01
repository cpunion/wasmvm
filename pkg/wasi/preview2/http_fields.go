package preview2

import "github.com/cpunion/wasmvm/internal/wasi/host"

type httpField = host.HTTPField
type httpFieldLimits = host.HTTPFieldLimits
type httpFields = host.HTTPFields
type httpHeaderError = host.HTTPHeaderError

const (
	httpHeaderSyntax    = host.HTTPHeaderSyntax
	httpHeaderForbidden = host.HTTPHeaderForbidden
	httpHeaderImmutable = host.HTTPHeaderImmutable
)

var httpToken = host.HTTPToken
var httpFieldValue = host.HTTPFieldValue
var newHTTPFields = host.NewHTTPFields
