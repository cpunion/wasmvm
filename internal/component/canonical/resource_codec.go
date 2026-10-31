package canonical

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ResourceCodec binds canonical values to the runtime's nominal declarations,
// ownership handoff and synchronous borrowed-call lifetime.
type ResourceCodec interface {
	Validate(*model.Type, Value) error
	Lift(*model.Type, uint32) (Value, error)
	Lower(*model.Type, Value) (uint32, error)
	Accept(*model.Type, Value, bool) (Value, error)
	Finish(error) error
}

func (c *Codec) resource() (ResourceCodec, error) {
	if c.Resources == nil {
		return nil, wasm.ErrUnsupported
	}
	return c.Resources, nil
}

// AcceptParameters validates all resource capabilities before moving any own
// argument into a host callback. The already cloned shape belongs to this call.
func (c *Codec) AcceptParameters(types []model.Field, values []Value) error {
	if err := c.PrepareParameters(types, values); err != nil {
		return err
	}
	for i, t := range types {
		if err := c.accept(t.Type, &values[i], true); err != nil {
			return err
		}
	}
	return nil
}

func (c *Codec) AcceptResult(t *model.Type, value *Value) error {
	if err := c.accept(t, value, false); err != nil {
		return err
	}
	return c.accept(t, value, true)
}

func (c *Codec) accept(t *model.Type, v *Value, commit bool) error {
	return walkResources(t, v, func(t *model.Type, v *Value) error {
		r, err := c.resource()
		if err != nil {
			return err
		}
		value, err := r.Accept(t, *v, commit)
		if err == nil && commit {
			*v = value
		}
		return err
	})
}

func walkResources(t *model.Type, v *Value, visit func(*model.Type, *Value) error) error {
	if t == nil {
		return nil
	}
	if t.Kind == model.Own || t.Kind == model.Borrow || t.Kind == model.Future || t.Kind == model.Stream {
		return visit(t, v)
	}
	switch t.Kind {
	case model.Record, model.Tuple:
		for i, f := range t.Fields {
			if err := walkResources(f.Type, &v.Elements[i], visit); err != nil {
				return err
			}
		}
	case model.List, model.FixedList:
		for i := range v.Elements {
			if err := walkResources(t.Element, &v.Elements[i], visit); err != nil {
				return err
			}
		}
	case model.Map:
		for i := range v.Elements {
			if err := walkResources(t.Key, &v.Elements[i].Elements[0], visit); err != nil {
				return err
			}
			if err := walkResources(t.Element, &v.Elements[i].Elements[1], visit); err != nil {
				return err
			}
		}
	case model.Variant, model.Option, model.Result:
		if len(v.Elements) != 0 {
			return walkResources(cases(t)[v.Case].Type, &v.Elements[0], visit)
		}
	}
	return nil
}

// VisitReferences visits capabilities in an already validated owned shape.
func VisitReferences(t *model.Type, v *Value, visit func(*Reference)) {
	walkResources(t, v, func(t *model.Type, v *Value) error { visit(v.Resource); return nil })
}

func (c *Codec) Finish(err error) error {
	if c.Resources != nil {
		return c.Resources.Finish(err)
	}
	return err
}

func (c *Codec) PrepareParameters(types []model.Field, values []Value) error {
	for i, t := range types {
		if err := c.accept(t.Type, &values[i], false); err != nil {
			return err
		}
	}
	return nil
}
