package native

// CPUID feature bits alone cannot authorize AVX instructions: the OS must have
// enabled saving both XMM and YMM state. Probe selection is pure Go so masked
// CPU/OS combinations can be checked on every target without executing XGETBV.
func x86FeatureMask(ecx1, ebx7 uint32, xcr0 uint64) uint32 {
	features := (ecx1>>9)&1 | ((ecx1>>19)&1)<<1
	const avxState = 1<<26 | 1<<27 | 1<<28 // XSAVE, OSXSAVE, AVX.
	if ecx1&avxState == avxState && ebx7&(1<<5) != 0 && xcr0&6 == 6 {
		features |= 4
	}
	return features
}
