package vm

import "github.com/cpunion/wasmvm/pkg/wasm"

// MemoryWrite describes one checked host-to-guest transfer. Data is copied
// during WriteMemoryBatch and is never retained by the VM.
type MemoryWrite struct {
	Offset uint64
	Data   []byte
}

// WriteMemoryBatch prechecks every range before writing in slice order. A bad
// range leaves memory unchanged; overlapping valid writes follow that order.
// Cancellation can stop a batch after partial progress. Shared-memory writes
// retain their individual synchronization, without whole-batch atomicity.
func (c *Caller) WriteMemoryBatch(index uint32, writes []MemoryWrite) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	if c.instance == nil || uint64(index) >= uint64(len(c.instance.Memories)) {
		return ErrReference
	}
	memory := c.instance.Memories[index]
	for i, write := range writes {
		if i%256 == 0 {
			if err := c.op.ctx.Err(); err != nil {
				return err
			}
		}
		if !memory.Range(write.Offset, uint64(len(write.Data))) {
			return &wasm.Trap{Code: wasm.TrapMemoryBounds}
		}
	}
	for i, write := range writes {
		if i%256 == 0 {
			if err := c.op.ctx.Err(); err != nil {
				return err
			}
		}
		if !memory.Write(write.Offset, write.Data) {
			return &wasm.Trap{Code: wasm.TrapMemoryBounds}
		}
	}
	return nil
}
