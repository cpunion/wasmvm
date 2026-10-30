package preview2

import (
	"strings"
	"sync"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Header values are immutable byte strings; casing/order belong to the WIT
// fields value rather than net/http.Header's canonicalized map representation.
type httpField struct{ name, value string }
type httpFieldLimits struct {
	count     uint32
	bytes     uint64
	forbidden map[string]bool // immutable provider-owned snapshot
}
type httpFields struct {
	mu        sync.RWMutex
	limits    httpFieldLimits
	immutable bool
	fields    []httpField
	size      uint64
}
type httpHeaderError uint32

const (
	httpHeaderSyntax httpHeaderError = iota
	httpHeaderForbidden
	httpHeaderImmutable
)

func (e httpHeaderError) Error() string {
	switch e {
	case httpHeaderSyntax:
		return "invalid HTTP field syntax"
	case httpHeaderForbidden:
		return "forbidden HTTP field"
	default:
		return "immutable HTTP fields"
	}
}
func httpToken(text string) bool {
	if text == "" {
		return false
	}
	for i := 0; i < len(text); i++ {
		b := text[i]
		if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' {
			continue
		}
		switch b {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		default:
			return false
		}
	}
	return true
}
func httpFieldValue(text string) bool {
	for i := 0; i < len(text); i++ {
		b := text[i]
		if b < 32 && b != '\t' || b == 127 {
			return false
		}
	}
	return true
}
func newHTTPFields(limits httpFieldLimits, fields []httpField, immutable bool) (*httpFields, error) {
	if limits.count == 0 || limits.bytes == 0 {
		return nil, wasm.ErrLimit
	}
	out := &httpFields{limits: limits, immutable: immutable}
	if uint64(len(fields)) > uint64(limits.count) {
		return nil, wasm.ErrLimit
	}
	for _, field := range fields {
		if err := out.checkSize(out.size, len(field.name), len(field.value)); err != nil {
			return nil, err
		}
		if !httpToken(field.name) || !httpFieldValue(field.value) {
			return nil, httpHeaderSyntax
		}
		if limits.forbidden[strings.ToLower(field.name)] {
			if immutable {
				continue
			}
			return nil, httpHeaderForbidden
		}
		out.fields = append(out.fields, field)
		out.size += uint64(len(field.name)) + uint64(len(field.value))
	}
	return out, nil
}
func (h *httpFields) checkSize(size uint64, name, value int) error {
	if size > h.limits.bytes || uint64(name) > h.limits.bytes-size || uint64(value) > h.limits.bytes-size-uint64(name) {
		return wasm.ErrLimit
	}
	return nil
}
func (h *httpFields) values(name string) []string {
	if uint64(len(name)) > h.limits.bytes || !httpToken(name) {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	var values []string
	for _, field := range h.fields {
		if strings.EqualFold(field.name, name) {
			values = append(values, field.value)
		}
	}
	return values
}
func (h *httpFields) has(name string) bool {
	if uint64(len(name)) > h.limits.bytes || !httpToken(name) {
		return false
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, field := range h.fields {
		if strings.EqualFold(field.name, name) {
			return true
		}
	}
	return false
}
func (h *httpFields) entries() []httpField {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]httpField(nil), h.fields...)
}
func (h *httpFields) clone() *httpFields {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return &httpFields{limits: h.limits, fields: append([]httpField(nil), h.fields...), size: h.size}
}
func (h *httpFields) mutate(name string, values []string, appendOnly bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.immutable {
		return httpHeaderImmutable
	}
	if uint64(len(name)) > h.limits.bytes || uint64(len(values)) > uint64(h.limits.count) {
		return wasm.ErrLimit
	}
	if !httpToken(name) {
		return httpHeaderSyntax
	}
	if h.limits.forbidden[strings.ToLower(name)] {
		return httpHeaderForbidden
	}
	size, count := h.size, uint64(len(h.fields))
	if !appendOnly {
		for _, field := range h.fields {
			if strings.EqualFold(field.name, name) {
				size -= uint64(len(field.name)) + uint64(len(field.value))
				count--
			}
		}
	}
	if uint64(len(values)) > uint64(h.limits.count)-count {
		return wasm.ErrLimit
	}
	for _, value := range values {
		if err := h.checkSize(size, len(name), len(value)); err != nil {
			return err
		}
		if !httpFieldValue(value) {
			return httpHeaderSyntax
		}
		size += uint64(len(name)) + uint64(len(value))
	}
	// Validation completes before modifying the original value.
	fields := make([]httpField, 0, int(count)+len(values))
	for _, field := range h.fields {
		if appendOnly || !strings.EqualFold(field.name, name) {
			fields = append(fields, field)
		}
	}
	for _, value := range values {
		fields = append(fields, httpField{name, value})
	}
	h.fields, h.size = fields, size
	return nil
}
