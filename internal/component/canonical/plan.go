package canonical

import "github.com/cpunion/wasmvm/internal/component/model"

// ValuePlan binds one immutable value type to a codec's layout policy. Access
// is serialized by the owning instance's lease, not shared across modules.
// Binding stays lazy so unused types do not consume layout budgets or change
// the ordering of cancellation, validation and resource effects.
type ValuePlan struct {
	typ     *model.Type
	builder *Builder
	layout  *Layout
}

func NewValuePlan(typ *model.Type) ValuePlan { return ValuePlan{typ: typ} }

func (p *ValuePlan) cached(c *Codec) *Layout {
	if p.builder == c.Layouts {
		return p.layout
	}
	return nil
}

func (p *ValuePlan) bind(c *Codec) (*Layout, error) {
	if layout := p.cached(c); layout != nil {
		return layout, nil
	}
	layout, err := c.Layouts.Build(p.typ)
	if err == nil {
		p.builder, p.layout = c.Layouts, layout
	}
	return layout, err
}

func (p *ValuePlan) FlatCount(c *Codec) (uint64, error) {
	layout, err := p.bind(c)
	if err != nil {
		return 0, err
	}
	return layout.FlatCount, nil
}

func (p *ValuePlan) LiftFlat(c *Codec, slots []Slot) (Value, error) {
	layout, err := p.bind(c)
	if err != nil {
		return Value{}, err
	}
	return c.liftFlatLayout(p.typ, layout, slots)
}

func (p *ValuePlan) validate(c *Codec, v Value) error {
	if err := c.validateLayout(p.typ, v, 0, p.cached(c)); err != nil {
		return err
	}
	_, err := p.bind(c)
	return err
}

// Clone retains the ordinary owned-value contract, including aggregate copies,
// resource validation and transfer charges. Only immutable layout work is bound.
func (p *ValuePlan) Clone(c *Codec, v Value) (Value, error) {
	if err := p.validate(c, v); err != nil {
		return Value{}, err
	}
	return c.clone(p.typ, v)
}

// RetainBorrowed validates and charges the same logical retention as Clone,
// without allocating aggregate snapshots. Its caller owns normalized private
// carriers and must not expose them to callbacks that retain aggregate slices.
func (p *ValuePlan) RetainBorrowed(c *Codec, v Value) error {
	if err := p.validate(c, v); err != nil {
		return err
	}
	_, err := c.copyValue(p.typ, v, false)
	return err
}

// LowerFlatInto appends to invocation-owned scalar storage. The codec and plan
// never retain that storage. Validation precedes lowering and resource effects.
func (p *ValuePlan) LowerFlatInto(c *Codec, v Value, dst []Slot) ([]Slot, error) {
	if err := p.validate(c, v); err != nil {
		return nil, err
	}
	if p.layout.FlatCount > MaxFlatParams {
		return nil, invalid("flat value requires memory")
	}
	return c.lowerFlat(p.typ, v, dst, 0)
}
