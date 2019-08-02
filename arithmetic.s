// func add4(c *Fe256, a *Fe256, b *Fe256)
TEXT ·add4(SB), $0-24
	// |
	MOVQ a+8(FP), DI
	MOVQ b+16(FP), SI
	XORQ AX, AX

	// |
	MOVQ (DI), R8
	ADDQ (SI), R8
	MOVQ 8(DI), R9
	ADCQ 8(SI), R9
	MOVQ 16(DI), R10
	ADCQ 16(SI), R10
	MOVQ 24(DI), R11
	ADCQ 24(SI), R11
	ADCQ $0x00, AX

	// |
	MOVQ R8, R12
	SUBQ ·modulus4+0(SB), R12
	MOVQ R9, R13
	SBBQ ·modulus4+8(SB), R13
	MOVQ R10, R14
	SBBQ ·modulus4+16(SB), R14
	MOVQ R11, R15
	SBBQ ·modulus4+24(SB), R15
	SBBQ $0x00, AX

	// |
	MOVQ    c+0(FP), DI
	CMOVQCC R12, R8
	MOVQ    R8, (DI)
	CMOVQCC R13, R9
	MOVQ    R9, 8(DI)
	CMOVQCC R14, R10
	MOVQ    R10, 16(DI)
	CMOVQCC R15, R11
	MOVQ    R11, 24(DI)
	RET
