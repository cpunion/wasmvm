package component

import (
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCanonicalMemoryBindingSurvivesNamespaceGrowthAndExpires(t *testing.T) {
	in := asyncInstance(t, "interpreter", "pending-callback", Imports{})
	ex, err := in.store.NewMemory(vm.MemoryType{Min: 1})
	if err != nil {
		t.Fatal(err)
	}
	n := &namespace{items: map[model.Sort][]item{model.CoreMemorySort: {{core: ex}}}}
	index := uint32(0)
	plan := &validate.CanonicalFunction{Definition: &model.Canonical{}, Memory: &index}
	var retained canonical.StringMemory
	_, err = in.store.WithCaller(nil, CallOptions{}, func(core *vm.Caller) error {
		codec := in.codec(core, n, plan)
		retained = codec.Transfer.Memory.(canonical.StringMemory)
		// Linking appends later definitions. Force the namespace backing array
		// to move and collect while the earlier transfer binding remains live.
		for range 128 {
			n.add(item{sort: model.CoreMemorySort})
		}
		runtime.GC()
		if err := retained.WriteString(0, "bound"); err != nil {
			return err
		}
		bytes := make([]byte, 5)
		if err := codec.Transfer.Memory.Read(0, bytes); err != nil || string(bytes) != "bound" {
			t.Fatal("memory binding changed after namespace growth", string(bytes), err)
		}
		if err := codec.Transfer.Memory.Write(0, []byte("fresh")); err != nil {
			return err
		}
		if text, err := retained.ReadString(0, 5); err != nil || text != "fresh" {
			t.Fatal("string memory binding", text, err)
		}
		if err := retained.WriteString(65536, "x"); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("binding bypassed memory bounds", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := retained.ReadString(0, 5); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal("binding retained an expired Caller", err)
	}
}
