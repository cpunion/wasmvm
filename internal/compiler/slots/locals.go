package slots

import "github.com/cpunion/wasmvm/internal/compiler/ir"

// A read may borrow its local only until the next write. Epochs reset write
// positions per block without a quadratic blocks-by-locals initialization.
func (c *slotCompiler) coalesceLocals(f *ir.Function, out *Function, pinned []bool, lastUse []uint32, budget *budget) []bool {
	budget.charge(uint64(len(f.Values)) + uint64(len(f.Locals))*2)
	coalesced := reuseSlice(c.coalesced, len(f.Values))
	clear(coalesced)
	writeBlock := reuseSlice(c.writeBlock, len(f.Locals))
	clear(writeBlock)
	nextWrite := reuseSlice(c.nextWrite, len(f.Locals))
	c.coalesced, c.writeBlock, c.nextWrite = coalesced, writeBlock, nextWrite
	for bi, b := range f.Blocks {
		budget.charge(uint64(len(b.Ops)))
		epoch := uint32(bi) + 1
		for oi := len(b.Ops) - 1; oi >= 0; oi-- {
			op := b.Ops[oi]
			src := f.Sources[op.Source].Wasm
			if src.Op == 0x20 || src.Op == 0x22 {
				v := op.Results[0]
				if !pinned[v] && (writeBlock[src.A] != epoch || lastUse[v] < nextWrite[src.A]) {
					coalesced[v] = true
					out.ValueSlots[v] = Slot(src.A)
				}
			}
			if src.Op == 0x21 || src.Op == 0x22 {
				writeBlock[src.A], nextWrite[src.A] = epoch, uint32(oi)
			}
		}
	}
	return coalesced
}
