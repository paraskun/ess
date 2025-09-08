package co2

import "github.com/paraskun/o2/typ"

type Command byte

const (
	JMP Command = 0x00              // jmp <b:8> <o:4>
	JAT         = JMP | (0x01 << 5) // jat <o:4>
	JIF         = JMP | (0x02 << 5) // jif <b:8> <o:4>
	NAT         = JMP | (0x03 << 5) // nat <b:8>

	RET = 0x01 // ret
	LEA = 0x02 // lea <b:8> <o:4>

	PS = 0x03
	PB = PS | (0x0 << 5) // pb <i:1>
	PW = PS | (0x1 << 5) // pw <i:4>
	PD = PS | (0x2 << 5) // pd <i:8>
	PA = PS | (0x3 << 5) // pa <s:4> <d:s>

	MR = 0x04
	LR = MR | (0x0 << 5) // lr <r:1>
	SR = MR | (0x1 << 5) // sr <r:1>

	LB = 0x05
	LW = 0x06
	LD = 0x07
	LA = 0x08

	SB = 0x09
	SW = 0x0a
	SD = 0x0b
	SA = 0x0c

	LBI = LB // lbi <b:8> <o:4>
	LWI = LW // lwi <b:8> <o:4>
	LDI = LD // ldi <b:8> <o:4>
	LAI = LA // lai <b:8> <o:4> <s:4>

	SBI = SB // sbi <b:8> <o:4>
	SWI = SW // swi <b:8> <o:4>
	SDI = SD // sdi <b:8> <o:4>
	SAI = SA // sai <b:8> <o:4> <s:4>

	LBS = LB | (1 << 5) // lbs <o:4>
	LWS = LW | (1 << 5) // lws <o:4>
	LDS = LD | (1 << 5) // lds <o:4>
	LAS = LA | (1 << 5) // las <o:4> <s:4>

	SBS = SB | (1 << 5) // sbs <o:4>
	SWS = SW | (1 << 5) // sws <o:4>
	SDS = SD | (1 << 5) // sds <o:4>
	SAS = SA | (1 << 5) // sas <o:4> <s:4>

	CNV = 0x0d
	I2U = CNV | (0x0 << 5) // i2u
	I2F = CNV | (0x1 << 5) // i2f
	U2I = CNV | (0x2 << 5) // u2i
	U2F = CNV | (0x3 << 5) // u2f
	F2I = CNV | (0x4 << 5) // f2i
	F2U = CNV | (0x5 << 5) // f2u

	ADD = 0x0e // add
	SUB = 0x0f // sub
	MUL = 0x10 // mul
	DIV = 0x11 // div
	POW = 0x12 // pow
	SHL = 0x13 // shl
	SHR = 0x14 // shr
	MOD = 0x15 // mod
	XOR = 0x16 // xor
	AND = 0x17 // and
	BOR = 0x18 // bor
	NOT = 0x19 // not
	NEG = 0x1a // neg

	LT = 0x1b // lt
	LE = 0x1c // le
	EQ = 0x1d // eq
	NE = 0x1e // ne

	ADDI = ADD | (typ.I64 << 5)
	SUBI = SUB | (typ.I64 << 5)
	MULI = MUL | (typ.I64 << 5)
	DIVI = DIV | (typ.I64 << 5)
	POWI = POW | (typ.I64 << 5)
	SHLI = SHL | (typ.I64 << 5)
	SHRI = SHR | (typ.I64 << 5)
	MODI = MOD | (typ.I64 << 5)
	XORI = XOR | (typ.I64 << 5)
	ANDI = AND | (typ.I64 << 5)
	BORI = BOR | (typ.I64 << 5)
	NOTI = NOT | (typ.I64 << 5)
	NEGI = NEG | (typ.I64 << 5)

	LTI = LT | (typ.I64 << 5)
	LEI = LE | (typ.I64 << 5)
	EQI = EQ | (typ.I64 << 5)
	NEI = NE | (typ.I64 << 5)

	ADDU = ADD | (typ.U64 << 5)
	SUBU = SUB | (typ.U64 << 5)
	MULU = MUL | (typ.U64 << 5)
	DIVU = DIV | (typ.U64 << 5)
	POWU = POW | (typ.U64 << 5)
	SHLU = SHL | (typ.U64 << 5)
	SHRU = SHR | (typ.U64 << 5)
	MODU = MOD | (typ.U64 << 5)
	XORU = XOR | (typ.U64 << 5)
	ANDU = AND | (typ.U64 << 5)
	BORU = BOR | (typ.U64 << 5)
	NOTU = NOT | (typ.U64 << 5)

	LTU = LT | (typ.F64 << 5)
	LEU = LE | (typ.F64 << 5)
	EQU = EQ | (typ.F64 << 5)
	NEU = NE | (typ.F64 << 5)

	ADDF = ADD | (typ.F64 << 5)
	SUBF = SUB | (typ.F64 << 5)
	MULF = MUL | (typ.F64 << 5)
	DIVF = DIV | (typ.F64 << 5)
	POWF = POW | (typ.F64 << 5)
	NEGF = NEG | (typ.F64 << 5)

	LTF = LT | (typ.F64 << 5)
	LEF = LE | (typ.F64 << 5)
	EQF = EQ | (typ.F64 << 5)
	NEF = NE | (typ.F64 << 5)

	ANDB = AND | (typ.BOOL << 5)
	BORB = BOR | (typ.BOOL << 5)
	NOTB = NOT | (typ.BOOL << 5)
)
