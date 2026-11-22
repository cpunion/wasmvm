//go:build cgo && !llgo && darwin && arm64

#include "textflag.h"

TEXT ·FetchAdd32(SB), NOSPLIT, $0-20
	MOVD address+0(FP), R0
	MOVWU delta+8(FP), R1
	LDADDALW R1, (R0), R2
	MOVW R2, ret+16(FP)
	RET

TEXT ·FetchAdd64(SB), NOSPLIT, $0-24
	MOVD address+0(FP), R0
	MOVD delta+8(FP), R1
	LDADDALD R1, (R0), R2
	MOVD R2, ret+16(FP)
	RET
