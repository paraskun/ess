package run

type Code byte

const (
	JMP Code = iota
	JIF

	// Function execution

	CALL // call <idx:4>
	RET

	// Memory

	ADDR // addr <idx:4>

	LBI // lbi <idx:4> <off:4>
	LDI // ldi <idx:4> <off:4>

	SBI // sbi <idx:4> <off:4>
	SDI // sdi <idx:4> <off:4>

	LBS // lb <idx:4>
	LDS // ld <idx:4>

	SBS // sb <idx:4>
	SDS // sd <idx:4>

	// Type conversion

	I2U // i2u
	I2F // i2f
	U2I // u2i
	U2F // u2f
	F2I // f2i
	F2U // f2u

	// Signed operations

	IADD
	ISUB
	IMUL
	IDIV
	IPOW
	ISHL
	ISHR
	IMOD

	IXOR
	IAND
	IOR

	IBNEG
	IUNEG

	ILT
	ILE
	IEQ
	INE

	// Logical operations

	LAND
	LOR
	LNEG
)
