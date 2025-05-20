package run

import (
	"github.com/paraskun/ess-go/typ"
)

type Code byte

const (
	JMP Code = 0x00 // jmp <idx:4>
	JIF      = 0x01 // jif <idx:4>

	// Function execution

	CALL = 0x02

	CALLI = CALL            // calli <idx:4>
	CALLS = CALL | (1 << 5) // calls

	RET = 0x03

	// Memory

	PUSH = 0x1f

	PUSHB = PUSH              // pushb <val:1>
	PUSHW = PUSH | 0b00100000 // pushw <val:4>
	PUSHD = PUSH | 0b01000000 // pushd <val:8>

	ADDR = 0x04

	ADDRI = ADDR            // addri <idx:4>
	ADDRS = ADDR | (1 << 5) // addrs

	LB = 0x05
	LW = 0x06
	LD = 0x07
	LA = 0x08

	SB = 0x09
	SW = 0x0a
	SD = 0x0b
	SA = 0x0c

	LBII = LB // lbii <idx:4> <off:4>
	LWII = LW // lwii <idx:4> <off:4>
	LDII = LD // ldii <idx:4> <off:4>
	LAII = LA // laii <idx:4> <off:4> <num:4>

	SBII = SB // sbii <idx:4> <off:4>
	SWII = SW // swii <idx:4> <off:4>
	SDII = SD // sdii <idx:4> <off:4>
	SAII = SA // saii <idx:4> <off:4> <num:4>

	LBIS = LB | (1 << 6) // lbis <idx:4>
	LWIS = LW | (1 << 6) // lwis <idx:4>
	LDIS = LD | (1 << 6) // ldis <idx:4>
	LAIS = LA | (1 << 6) // lais <idx:4> <num:4>

	SBIS = SB | (1 << 6) // sbis <idx:4>
	SWIS = SW | (1 << 6) // swis <idx:4>
	SDIS = SD | (1 << 6) // sdis <idx:4>
	SAIS = SA | (1 << 6) // sais <idx:4> <num:4>

	LBSI = LB | (1 << 5) // lbsi <off:4>
	LWSI = LW | (1 << 5) // lwsi <off:4>
	LDSI = LD | (1 << 5) // ldsi <off:4>
	LASI = LA | (1 << 5) // lasi <off:4> <num:4>

	SBSI = SB | (1 << 5) // sbsi <off:4>
	SWSI = SW | (1 << 5) // swsi <off:4>
	SDSI = SD | (1 << 5) // sdsi <off:4>
	SASI = SA | (1 << 5) // sasi <off:4> <num:4>

	LBSS = LB | (1 << 5) | (1 << 6) // lbss
	LWSS = LW | (1 << 5) | (1 << 6) // lwss
	LDSS = LD | (1 << 5) | (1 << 6) // ldss
	LASS = LA | (1 << 5) | (1 << 6) // lass <num:4>

	SBSS = SB | (1 << 5) | (1 << 6) // sbss
	SWSS = SW | (1 << 5) | (1 << 6) // swss
	SDSS = SD | (1 << 5) | (1 << 6) // sdss
	SASS = SA | (1 << 5) | (1 << 6) // sdss <num:4>

	// Type conversion

	I2U = 0x0d | (0b00000000) // i2u
	I2F = 0x0d | (0b00100000) // i2f
	U2I = 0x0d | (0b01000000) // u2i
	U2F = 0x0d | (0b01100000) // u2f
	F2I = 0x0d | (0b10000000) // f2i
	F2U = 0x0d | (0b10100000) // f2u

	ADD = 0x0e
	SUB = 0x0f
	MUL = 0x10
	DIV = 0x11
	POW = 0x12
	SHL = 0x13
	SHR = 0x14
	MOD = 0x15
	XOR = 0x16
	AND = 0x17
	OR  = 0x18

	BNEG = 0x19
	UNEG = 0x1a

	LT = 0x1b
	LE = 0x1c
	EQ = 0x1d
	NE = 0x1e

	// Signed operations

	ADDI = ADD | typ.I64
	SUBI = SUB | typ.I64
	MULI = MUL | typ.I64
	DIVI = DIV | typ.I64
	POWI = POW | typ.I64
	SHLI = SHL | typ.I64
	SHRI = SHR | typ.I64
	MODI = MOD | typ.I64

	XORI = XOR | typ.I64
	ANDI = AND | typ.I64
	ORI  = OR | typ.I64

	BNEGI = BNEG | typ.I64
	UNEGI = UNEG | typ.I64

	LTI = LT | typ.I64
	LEI = LE | typ.I64
	EQI = EQ | typ.I64
	NEI = NE | typ.I64

	// Unsigned operations

	ADDU = ADD | typ.U64
	SUBU = SUB | typ.U64
	MULU = MUL | typ.U64
	DIVU = DIV | typ.U64
	POWU = POW | typ.U64
	SHLU = SHL | typ.U64
	SHRU = SHR | typ.U64
	MODU = MOD | typ.U64

	XORU = XOR | typ.U64
	ANDU = AND | typ.U64
	ORU  = OR | typ.U64

	BNEGU = BNEG | typ.U64
	UNEGU = UNEG | typ.U64

	LTU = LT | typ.U64
	LEU = LE | typ.U64
	EQU = EQ | typ.U64
	NEU = NE | typ.U64

	// Floating-point operations

	ADDF = ADD | typ.F64
	SUBF = SUB | typ.F64
	MULF = MUL | typ.F64
	DIVF = DIV | typ.F64
	POWF = POW | typ.F64

	UNEGF = UNEG | typ.F64

	LTF = LT | typ.F64
	LEF = LE | typ.F64
	EQF = EQ | typ.F64
	NEF = NE | typ.F64

	// Logical operations

	ANDL = AND | typ.BOOL
	ORL  = OR | typ.BOOL
	NEGL = BNEG | typ.BOOL
)
