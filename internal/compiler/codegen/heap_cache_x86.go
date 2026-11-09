package codegen

func (e *functionEmitter) heapCachePrioritiesAMD64(config Config, frequencies []uint64) map[uint64]uint64 {
	if e.a.arch != "amd64" {
		return nil
	}
	return e.heapCacheReusePriorities(config, frequencies)
}

// Keep one checked cache within the existing guest GPR inventory. Its
// two words compete with ordinary slot priorities rather than expanding banks.
func (e *functionEmitter) selectHeapCacheAMD64(class uint64, available []byte, used *[32]bool) {
	if len(e.heapCaches) != 0 {
		return
	}
	var pair [2]byte
	n := 0
	for _, r := range available {
		if !used[r] {
			pair[n] = r
			n++
			if n == 2 {
				break
			}
		}
	}
	if n != 2 {
		return
	}
	used[pair[0]], used[pair[1]] = true, true
	e.heapCaches = map[uint64]heapCache{class: {pair[0], pair[1]}}
}
