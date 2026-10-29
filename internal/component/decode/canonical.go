package decode

import "github.com/cpunion/wasmvm/internal/component/model"

func (d *decoder) options() []model.CanonicalOption {
	opts := make([]model.CanonicalOption, d.count())
	for i := range opts {
		o := &opts[i]
		o.Kind = d.byte()
		switch o.Kind {
		case 0, 1, 2, 6:
		case 3, 4, 5, 7:
			o.Index = d.u32()
		default:
			d.fail("unknown canonical option")
		}
	}
	return opts
}

func (d *decoder) canonical() *model.Canonical {
	x := &model.Canonical{Op: d.byte()}
	switch x.Op {
	case 0:
		d.expect(0)
		x.Index = d.u32()
		x.Options = d.options()
		x.Extra = d.u32()
	case 1:
		d.expect(0)
		x.Index = d.u32()
		x.Options = d.options()
	case 2, 3, 4, 0x0e, 0x13, 0x14, 0x15, 0x1a, 0x1b, 0x2e, 0x2f:
		x.Index = d.u32()
	case 5, 0x0d, 0x1e, 0x1f, 0x22, 0x23, 0x24, 0x25, 0x26, 0x28:
	case 6, 0x11, 0x12, 0x18, 0x19:
		if x.Op != 6 {
			x.Index = d.u32()
		}
		if d.optional() {
			x.Flag = 1
		}
	case 9:
		x.Results = d.result()
		x.Options = d.options()
	case 0x0a, 0x0b:
		x.CoreValue = d.coreValue()
		x.Index = d.u32()
	case 0x0c, 0x29, 0x2a, 0x2b, 0x2c, 0x2d:
		d.expect(0)
	case 0x0f, 0x10, 0x16, 0x17:
		x.Index = d.u32()
		x.Options = d.options()
	case 0x1c, 0x1d:
		x.Options = d.options()
	case 0x20, 0x21:
		d.expect(0)
		x.Index = d.u32()
	case 0x27:
		x.Index = d.u32()
		x.Extra = d.u32()
	case 0x40, 0x41:
		if d.optional() {
			x.Flag = 1
		}
		x.Index = d.u32()
		if x.Op == 0x41 {
			x.Extra = d.u32()
		}
	case 0x42:
		if d.optional() {
			x.Flag = 1
		}
	default:
		d.fail("unknown canonical operation")
	}
	return x
}
