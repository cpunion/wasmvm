package host

import (
	"strings"
	"sync"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Header values are immutable byte strings; casing/order belong to the WIT
// fields value rather than net/http.Header's canonicalized map representation.
type HTTPField struct{ Name, Value string }
type HTTPFieldLimits struct {
	Count     uint32
	Bytes     uint64
	Forbidden map[string]bool // immutable provider-owned snapshot
}
type HTTPFields struct {
	mu        sync.RWMutex
	limits    HTTPFieldLimits
	immutable bool
	fields    []HTTPField
	size      uint64
}
type HTTPHeaderError uint32

const (
	HTTPHeaderSyntax HTTPHeaderError = iota
	HTTPHeaderForbidden
	HTTPHeaderImmutable
)

func (e HTTPHeaderError) Error() string {
	switch e {
	case HTTPHeaderSyntax:
		return "invalid HTTP field syntax"
	case HTTPHeaderForbidden:
		return "forbidden HTTP field"
	default:
		return "immutable HTTP fields"
	}
}
func HTTPToken(text string) bool {
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
func HTTPFieldValue(text string) bool {
	for i := 0; i < len(text); i++ {
		b := text[i]
		if b < 32 && b != '\t' || b == 127 {
			return false
		}
	}
	return true
}
func NewHTTPFields(limits HTTPFieldLimits, fields []HTTPField, immutable bool) (*HTTPFields, error) {
	if limits.Count == 0 || limits.Bytes == 0 {
		return nil, wasm.ErrLimit
	}
	out := &HTTPFields{limits: limits, immutable: immutable}
	if uint64(len(fields)) > uint64(limits.Count) {
		return nil, wasm.ErrLimit
	}
	for _, field := range fields {
		if err := out.checkSize(out.size, len(field.Name), len(field.Value)); err != nil {
			return nil, err
		}
		if !HTTPToken(field.Name) || !HTTPFieldValue(field.Value) {
			return nil, HTTPHeaderSyntax
		}
		if limits.Forbidden[strings.ToLower(field.Name)] {
			if immutable {
				continue
			}
			return nil, HTTPHeaderForbidden
		}
		out.fields = append(out.fields, field)
		out.size += uint64(len(field.Name)) + uint64(len(field.Value))
	}
	return out, nil
}
func (h *HTTPFields) checkSize(size uint64, name, value int) error {
	if size > h.limits.Bytes || uint64(name) > h.limits.Bytes-size || uint64(value) > h.limits.Bytes-size-uint64(name) {
		return wasm.ErrLimit
	}
	return nil
}
func (h *HTTPFields) Values(name string) []string {
	if uint64(len(name)) > h.limits.Bytes || !HTTPToken(name) {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	var values []string
	for _, field := range h.fields {
		if strings.EqualFold(field.Name, name) {
			values = append(values, field.Value)
		}
	}
	return values
}
func (h *HTTPFields) Has(name string) bool {
	if uint64(len(name)) > h.limits.Bytes || !HTTPToken(name) {
		return false
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, field := range h.fields {
		if strings.EqualFold(field.Name, name) {
			return true
		}
	}
	return false
}
func (h *HTTPFields) Entries() []HTTPField {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]HTTPField(nil), h.fields...)
}
func (h *HTTPFields) Clone() *HTTPFields {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return &HTTPFields{limits: h.limits, fields: append([]HTTPField(nil), h.fields...), size: h.size}
}
func (h *HTTPFields) Mutate(name string, values []string, appendOnly bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.mutate(name, values, appendOnly)
}
func (h *HTTPFields) mutate(name string, values []string, appendOnly bool) error {
	if h.immutable {
		return HTTPHeaderImmutable
	}
	if uint64(len(name)) > h.limits.Bytes || uint64(len(values)) > uint64(h.limits.Count) {
		return wasm.ErrLimit
	}
	if !HTTPToken(name) {
		return HTTPHeaderSyntax
	}
	if h.limits.Forbidden[strings.ToLower(name)] {
		return HTTPHeaderForbidden
	}
	size, count := h.size, uint64(len(h.fields))
	if !appendOnly {
		for _, field := range h.fields {
			if strings.EqualFold(field.Name, name) {
				size -= uint64(len(field.Name)) + uint64(len(field.Value))
				count--
			}
		}
	}
	if uint64(len(values)) > uint64(h.limits.Count)-count {
		return wasm.ErrLimit
	}
	for _, value := range values {
		if err := h.checkSize(size, len(name), len(value)); err != nil {
			return err
		}
		if !HTTPFieldValue(value) {
			return HTTPHeaderSyntax
		}
		size += uint64(len(name)) + uint64(len(value))
	}
	// Validation completes before modifying the original value.
	fields := make([]HTTPField, 0, int(count)+len(values))
	for _, field := range h.fields {
		if appendOnly || !strings.EqualFold(field.Name, name) {
			fields = append(fields, field)
		}
	}
	for _, value := range values {
		fields = append(fields, HTTPField{name, value})
	}
	h.fields, h.size = fields, size
	return nil
}

// Freeze prevents subsequent mutations through every shared reference.
func (h *HTTPFields) Freeze() { h.mu.Lock(); h.immutable = true; h.mu.Unlock() }

// GetAndDelete atomically returns and removes all matching values.
func (h *HTTPFields) GetAndDelete(name string) ([]string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	var values []string
	for _, field := range h.fields {
		if strings.EqualFold(field.Name, name) {
			values = append(values, field.Value)
		}
	}
	if err := h.mutate(name, nil, false); err != nil {
		return nil, err
	}
	return values, nil
}
