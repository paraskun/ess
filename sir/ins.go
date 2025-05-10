package sir

import "github.com/paraskun/ess-go/typ"

type Op uint8

const (
	MOV Op = iota

	ADD
	SUB
	MUL
	DIV
	MOD
	POW
	SHL
	SHR
	BAND
	BOR
	BXOR
	BNEG
	UNEG
	LNEG
	LAND
	LOR
	LE
	LT
	EQ
	NE

	I2U
	I2F
	U2I
	U2F
	F2I
	F2U
)

type Ins struct {
	Op Op

	X Arg
	Y Arg
	R Arg
}

type (
	Arg interface {
		Type() *typ.Type
	}

	ImmArg struct {
		obj typ.Object
		env *typ.Env
	}

	VarArg struct {
		obj typ.Object
		env *typ.Env
	}

	TmpArg struct {
		typ *typ.Type
		num uint
	}
)

func (a *ImmArg) Type() *typ.Type { return a.obj.Typ }
func (a *VarArg) Type() *typ.Type { return a.obj.Typ }
func (a *TmpArg) Type() *typ.Type { return a.typ }

var ops = map[Op]string{
	ADD:  "add",
	SUB:  "sub",
	MUL:  "mul",
	DIV:  "div",
	MOD:  "mod",
	POW:  "pow",
	SHL:  "shl",
	SHR:  "shr",
	BAND: "band",
	BOR:  "bor",
	BXOR: "bxor",
	BNEG: "bneg",
	UNEG: "uneg",
	LNEG: "lneg",
	LAND: "land",
	LOR:  "lor",
	LE:   "le",
	LT:   "lt",
	EQ:   "eq",
	NE:   "ne",
	I2U:  "i2u",
	I2F:  "i2f",
	U2I:  "u2i",
	U2F:  "u2f",
	F2I:  "f2i",
	F2U:  "f2u",
}

func (o Op) String() string {
	return ops[o]
}
