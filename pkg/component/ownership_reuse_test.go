package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLoweredRollbackRetainsReusedGuestSlot(t *testing.T) {
	budget := canonical.NewResourceBudget(8)
	in := &Instance{hostResources: canonical.NewResourceTable(budget)}
	n := &namespace{handles: canonical.NewGuestResourceTable(budget)}
	typ := &canonical.ResourceType{}
	rt := &resourceType{instance: in, owner: n, identity: typ, representation: wasm.I32}
	n.resources = map[uint64]*resourceType{1: rt}
	codec := &resourceCodec{in: in, n: n}
	id, err := in.hostResources.New(typ, 7)
	if err != nil {
		t.Fatal(err)
	}
	ref := canonical.NewReference(in.hostResources, typ, id, rt, nil)
	guestID, err := codec.Lower(&model.Type{Kind: model.Own, ResourceID: 1}, Value{Resource: ref})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := n.handles.Take(typ, guestID); err != nil {
		t.Fatal(err)
	}
	replacement, err := n.handles.New(typ, 99)
	if err != nil || replacement != guestID {
		t.Fatal(replacement, err)
	}
	if err := codec.Finish(wasm.ErrTrap); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if rep, err := n.handles.Rep(typ, replacement); err != nil || rep != 99 {
		t.Fatal("failed call consumed a different resource after guest slot reuse", rep, err)
	}
	n.handles.Close()
	in.hostResources.Close()
}
