package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func TestByteStreamSchemaReadsPackedDataAndRetainsStoreOwnership(t *testing.T) {
	typ := ByteStreamType()
	if !typ.Valid() || typ.Kind() != KindStream || typ.Element().Kind() != KindU8 {
		t.Fatal(typ)
	}
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		in := asyncInstance(t, backend, "host-result-stream", Imports{})
		other := asyncInstance(t, backend, "host-result-stream", Imports{})
		pending := "\x00\xff\x80bytes"
		closed := 0
		value, err := in.NewSource(nil, typ, AsyncSource{Prompt: true, Read: func(*Caller, uint32) (*ReadCompletion, error) {
			completion := NewReadCompletion()
			completion.Resolve(func(*Caller) (ReadResult, error) {
				data := pending
				pending = ""
				return ReadResult{Bytes: data, Closed: true}, nil
			})
			return completion, nil
		}, Close: func() { closed++ }})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := other.ReadAsync(nil, value.Resource, 32, CallOptions{}); !errors.Is(err, vm.ErrReference) {
			t.Fatal("primitive schema bypassed Store ownership", err)
		}
		r, err := in.ReadAsync(nil, value.Resource, 32, CallOptions{})
		if err != nil || r.Bytes != "\x00\xff\x80bytes" || !r.Closed || closed != 1 {
			t.Fatal(r, err, closed)
		}
		if _, err := in.DropAsync(nil, value.Resource, CallOptions{}); err != nil {
			t.Fatal(err)
		}
	}
}
