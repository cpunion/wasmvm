package validate

import "github.com/cpunion/wasmvm/internal/component/model"
import "github.com/cpunion/wasmvm/pkg/wasm"

// Equal compares immutable value/function/resource graphs without expanding
// aliases. Nominal identities are local to their compilation pool. Interface
// width subtyping and binder substitution belong to component linking.
func Equal(a, b *Type, maxNodes uint32) (bool, error) {
	c := equality{remaining: maxNodes, visited: map[pair]bool{}}
	return c.typ(a, b)
}

func (p *TypePool) equal(a, b *Type) (bool, error) {
	budget := p.limits.MaxNodes - p.nodes
	c := equality{remaining: budget, visited: map[pair]bool{}}
	ok, err := c.typ(a, b)
	if charge := p.charge(uint64(budget - c.remaining)); charge != nil {
		return false, charge
	}
	return ok, err
}

func (c *equality) typ(a, b *Type) (bool, error) {
	if a == nil || b == nil {
		return a == b, nil
	}
	if a.Kind != b.Kind || (a.resource || b.resource) && a.owner != b.owner {
		return false, nil
	}
	if a.Kind == model.Resource {
		return a.Resource == b.Resource, nil
	}
	if a.Kind == model.Component || a.Kind == model.Instance || a.CoreModule != nil || b.CoreModule != nil {
		// Interfaces bind fresh resource identities. Ordinary nominal graph
		// equality is not a substitute for binder-aware interface subtyping.
		return false, wasm.ErrUnsupported
	}
	if a.Value != nil || b.Value != nil {
		return c.value(a.Value, b.Value)
	}
	if len(a.Parameters) != len(b.Parameters) {
		return false, nil
	}
	for i, p := range a.Parameters {
		if !c.charge() {
			return false, limit("type comparison")
		}
		if p.Name != b.Parameters[i].Name {
			return false, nil
		}
		ok, err := c.value(p.Type, b.Parameters[i].Type)
		if err != nil || !ok {
			return ok, err
		}
	}
	return c.value(a.Result, b.Result)
}

type pair struct{ a, b *model.Type }
type equality struct {
	remaining uint32
	visited   map[pair]bool
}

func (c *equality) charge() bool {
	if c.remaining == 0 {
		return false
	}
	c.remaining--
	return true
}

func (c *equality) value(a, b *model.Type) (bool, error) {
	if a == b {
		return true, nil
	}
	if a == nil || b == nil || a.Kind != b.Kind || a.Length != b.Length || a.ResourceID != b.ResourceID || len(a.Fields) != len(b.Fields) || len(a.Cases) != len(b.Cases) || len(a.Labels) != len(b.Labels) {
		return false, nil
	}
	k := pair{a, b}
	if c.visited[k] {
		return true, nil
	}
	if !c.charge() {
		return false, limit("type comparison")
	}
	c.visited[k] = true
	for i, name := range a.Labels {
		if !c.charge() {
			return false, limit("type comparison")
		}
		if name != b.Labels[i] {
			return false, nil
		}
	}
	for _, fields := range [][2][]model.Field{{a.Fields, b.Fields}, {a.Cases, b.Cases}} {
		for i, f := range fields[0] {
			if !c.charge() {
				return false, limit("type comparison")
			}
			if f.Name != fields[1][i].Name {
				return false, nil
			}
			ok, err := c.value(f.Type, fields[1][i].Type)
			if err != nil || !ok {
				return ok, err
			}
		}
	}
	if ok, err := c.value(a.Element, b.Element); err != nil || !ok {
		return ok, err
	}
	return c.value(a.Key, b.Key)
}
