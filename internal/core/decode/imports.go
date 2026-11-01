package decode

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Grouping changes only the encoding. All later stages retain the same ordered,
// flat imports, so linking, index spaces and ownership have one implementation.
func (d *decoder) importSection(m *module.Module) {
	n := d.count()
	m.Imports = make([]module.Import, 0, n)
	for i := uint32(0); i < n; i++ {
		mod, name, kind := d.name(), d.name(), d.byte()
		if kind != 0x7f && kind != 0x7e {
			im := d.importType(kind)
			im.Module, im.Name = mod, name
			m.Imports = append(m.Imports, im)
			continue
		}
		if d.features&wasm.CompactImports == 0 || name != "" {
			d.fail("compact import encoding is disabled or has a nonempty marker name")
		}
		var shared module.Import
		if kind == 0x7e {
			shared = d.importType(d.byte())
		}
		count := d.count()
		for j := uint32(0); j < count; j++ {
			item := d.name()
			im := shared
			if kind == 0x7f {
				im = d.importType(d.byte())
			}
			im.Module, im.Name = mod, item
			m.Imports = append(m.Imports, im)
		}
	}
}

func (d *decoder) importType(kind byte) module.Import {
	im := module.Import{Kind: kind}
	switch kind {
	case 0:
		im.Type = d.u32()
	case 1:
		im.Table = d.table()
	case 2:
		im.Memory = d.bounds(true)
	case 3:
		im.Global = d.globalType()
	case 4:
		im.Type = d.tagType()
	default:
		d.fail("unknown Core external type")
	}
	return im
}
