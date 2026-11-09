package preview3

import "io"

// Accepted workers own their mutable transfer buffer until the operation has
// copied or consumed it. Idle retention uses the same bounded owner policy as
// worker reuse; Close drops the cache and late returns cannot repopulate it.
func (p *IO) acquireTransferBuffer(size int) []byte {
	if size == 0 {
		return nil
	}
	var buffer []byte
	p.mu.Lock()
	if n := len(p.idleBuffers); n != 0 {
		buffer = p.idleBuffers[n-1]
		p.idleBuffers[n-1] = nil
		p.idleBuffers = p.idleBuffers[:n-1]
	}
	p.mu.Unlock()
	if cap(buffer) < size {
		buffer = make([]byte, size)
	}
	return buffer[:size]
}

func (p *IO) releaseTransferBuffer(buffer []byte) {
	if cap(buffer) == 0 || uint64(cap(buffer)) > uint64(max(p.config.MaxReadBytes, p.config.MaxWriteBytes)) {
		return
	}
	p.mu.Lock()
	if !p.closed && uint64(len(p.idleBuffers)) < uint64(p.config.MaxIdleWorkers) {
		p.idleBuffers = append(p.idleBuffers, buffer[:0])
	}
	p.mu.Unlock()
}

func (p *IO) writeTransferString(writer io.Writer, data string) (int, error) {
	buffer := p.acquireTransferBuffer(len(data))
	defer p.releaseTransferBuffer(buffer)
	copy(buffer, data)
	return writer.Write(buffer[:len(buffer):len(buffer)])
}
