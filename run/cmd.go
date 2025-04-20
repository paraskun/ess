package run

type Command byte

const (
	JMP Command = iota
	JIF

	VADDR // vaddr <idx:4>
	SADDR // saddr <idx:4>

	GET // get <idx:4> <len:1>
	PUT // put <idx:4> <len:1>

	PGET // pget <idx:4> <len:1>
	PPUT // pput <idx:4> <len:1>

	SGET // sget <idx:4>
	SPUT // sput <idx:4>

	SPGET // spget
	SPPUT // spput

	I2F
	F2I
	I2S
	F2S

	// string operations

	SADD
	SMUL

	// integer operations

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

	// floating-point operations

	FADD
	FSUB
	FMUL
	FDIV
	FPOW
	FSHL
	FSHR
	FMOD
	FBAND
	FBOR
	FBXOR

	FBNEG
	FUNEG

	FLT
	FLE
	FNE
)
