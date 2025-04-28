package run

type Cmd byte

const (
	JMP Cmd = iota
	JIF

	CALL // call <idx:8>

	ADDR // addr <idx:8>

	LDV // ldv <idx:8> <sz:1>
	MVV // mvv <idx:8> <sz:1>
	LDR // ldr <idx:8> <off:8> <sz:1>
	MVR // mvr <idx:8> <off:8> <sz:1>

	// type conversion

	I2U
	I2F
	U2I
	U2F
	F2I
	F2U

	// signed operations

	IADD
	ISUB
	IMUL
	IDIV
	IPOW
	ISHL
	ISHR
	IMOD
	IBAND
	IBOR
	IBXOR

	IBNEG
	IUNEG

	ILT
	ILE
	INE

	// unsigned operations

	UADD
	USUB
	UMUL
	UDIV
	UPOW
	USHL
	USHR
	UMOD
	UBAND
	UBOR
	UBXOR

	UBNEG
	UUNEG

	ULT
	ULE
	UNE

	// floating-point operations

	FADD
	FSUB
	FMUL
	FDIV
	FPOW

	FUNEG

	FLT
	FLE
	FNE
)
