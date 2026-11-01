package component

import "github.com/cpunion/wasmvm/pkg/vm"

// Source Close may run off-Store during shutdown. Queue affine cleanup rather
// than entering the VM from that callback. An active session drains the queue;
// shutdown drains remaining ownership only after the Core Store has joined.
func (in *Instance) flushSourceCleanup(c *vm.Caller) (err error) {
	if in.drainingSources {
		return nil
	}
	in.drainingSources = true
	defer func() { in.drainingSources = false }()
	guard := cleanupGuard{}
	for {
		in.mu.Lock()
		source := in.sourceCleanup
		if source != nil {
			in.sourceCleanup = source.cleanupNext
			source.cleanupNext = nil
			source.queued = false
		}
		in.mu.Unlock()
		if source == nil {
			break
		}
		in.mu.Lock()
		closeWriter := source.closeWriter
		source.closeWriter = false
		in.mu.Unlock()
		if closeWriter {
			guard.run(source.writer.closeSource)
		}
		func() {
			defer func() {
				in.heldValues -= source.values
				in.heldBytes -= source.bytes
				source.values, source.bytes = 0, 0
			}()
			refs := source.owned
			source.owned = nil
			for _, ref := range refs {
				guard.run(func() {
					if failed := in.cleanupResource(c, ref); err == nil {
						err = failed
					}
				})
			}
		}()
	}
	guard.finish()
	return err
}

func (in *Instance) finishSourceCleanup(c *vm.Caller, err *error) {
	guard := cleanupGuard{panicked: recover()}
	guard.run(func() {
		if failed := in.flushSourceCleanup(c); *err == nil {
			*err = failed
		}
	})
	guard.finish()
}
