package component

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
)

func TestAcceptedParametersPublishOnlyTheirOwnershipPrefix(t *testing.T) {
	for _, capacity := range []int{1, 8} {
		in := &Instance{hostResources: canonical.NewResources(8)}
		group := &taskScheduler{}
		add := func() (*asyncEnd, *Resource) {
			end := &asyncEnd{in: in, group: group, next: group.ends}
			if group.ends != nil {
				group.ends.previous = end
			}
			group.ends = end
			id, err := in.hostResources.AddOpaque(end)
			if err != nil {
				t.Fatal(err)
			}
			end.id = id
			return end, canonical.NewReference(in.hostResources, nil, id, end, nil)
		}
		accepted, parameter := add()
		codec := resourceCodec{in: in, created: make([]*Resource, 0, capacity)}
		codec.created = append(codec.created, parameter)
		codec.ParametersAccepted()
		result, reference := add()
		codec.created = append(codec.created, reference)
		runtime.GC()
		if err := codec.Finish(nil); err != nil {
			t.Fatal(err)
		}
		if accepted.group != nil || result.group != group || group.ends != result || result.next != nil {
			t.Fatal("result ownership escaped before its canonical handoff", capacity)
		}
		if codec.created != nil || codec.escaped != nil {
			t.Fatal("finished codec retained capabilities")
		}
		if err := in.hostResources.RemoveOpaque(accepted.id); err != nil {
			t.Fatal(err)
		}
		result.abort()
		if group.ends != nil {
			t.Fatal("unpublished result was not reclaimed")
		}
	}
}
