package codegen

// Charge both words of a checked cache to the ordinary caller budget, instead
// of assigning caches only after all literal/metadata candidates have won.
// Four saved instructions per repeated access is a conservative potential
// saving, or two per cache word. Frequency is a profitability hint, never a
// claim about actual identity reuse; every hit still checks the full handle.
func (e *functionEmitter) heapCachePrioritiesARM64(frequencies []uint64) map[uint64]uint64 {
	if !e.unmetered || e.a.arch != "arm64" || len(e.heapAccesses) == 0 {
		return nil
	}
	weights := map[uint64]uint64{}
	for pc, in := range e.f.Code {
		if access, ok := e.heapAccesses[in.Source]; ok && in.Op>>16 == 0xfb {
			weights[access.cacheClass] += frequency(frequencies, pc)
		}
	}
	for class, weight := range weights {
		if weight < 2 {
			delete(weights, class)
		} else {
			weights[class] = 2 * weight
		}
	}
	return weights
}
