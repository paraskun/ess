package ast

import (
	"bytes"
	"encoding/binary"

	"github.com/paraskun/ess-go/lex"
	"github.com/paraskun/ess-go/run"
	"github.com/paraskun/ess-go/typ"
)

func Assemble(p *Pragma) *run.Pragma {
	a := assembler{
		env: p.Env,
		obj: &run.Pragma{},
	}

	p.Accept(&a)

	return a.obj
}

type assembler struct {
	env *typ.Env
	obj *run.Pragma
	img *run.FuncImage
	src *bytes.Buffer
}

func (asm *assembler) VisitDecl(u Decl) {
	switch dec := u.(type) {
	case *FuncDecl:
		img := run.FuncImage{}
		asm.src = &bytes.Buffer{}

		for _, arg := range dec.Spec.Arg {
			argSize := arg.Obj.Typ.Size()

			arg.Obj.Off = img.DataSize
			img.DataSize += argSize
			img.ArgsSize += argSize
		}

		buf := &bytes.Buffer{}

		for _, imm := range dec.Body.Env.Imm {
			imm.Off = img.DataSize
			img.DataSize += imm.Size()

			binary.Write(buf, binary.LittleEndian, imm.Val)
		}

		dec.Body.Accept(asm)

		img.DataSize += 8
		img.Imm = buf.Bytes()
		img.Src = asm.src.Bytes()

		dec.Obj.Off = len(asm.obj.Img)
		asm.obj.Img = append(asm.obj.Img, img)
	}
}

func (asm *assembler) VisitStmt(u Stmt) {
	switch s := u.(type) {
	case *BlockStmt:
		for _, b := range s.Body {
			b.Accept(asm)
		}
	case *AssignStmt:
		s.Val[0].Accept(asm)
		asm.getPosition(s.Var[0])

		switch s.Var[0].Type()[0].Kind {
		case typ.BOOL:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SBSS))
		case typ.I64, typ.U64, typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SDSS))
		case typ.FUNC:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SWSS))
		case typ.COMP:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SASS))
			binary.Write(asm.src, binary.LittleEndian, uint64(s.Var[0].Type()[0].Size()))
		}
	case *LoopStmt:
	case *CondStmt:
	case *ReturnStmt:
	}
}

func (asm *assembler) getPosition(u Expr) (int, int) {
	switch exp := u.(type) {
	case *IdfExpr:
		if !exp.Obj.Loc {
			return -1, exp.Obj.Off
		}

		if exp.Obj.Ref {
			return exp.Obj.Off, 0
		}

		return 0, exp.Obj.Off

	case *DotExpr:
		b, o := asm.getPosition(exp.Comp)
		ct := exp.Comp.Type()[0].Info.(*typ.CompType)

		return b, o + ct.Fields[exp.Field.Lit].Off

	default:
		panic("could not get position")
	}
}

func (asm *assembler) VisitExpr(u Expr) {
	switch exp := u.(type) {
	case *BaseImmExpr:
		switch exp.Type()[0].Kind {
		case typ.BOOL:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LBII))
			binary.Write(asm.src, binary.LittleEndian, 0)
			binary.Write(asm.src, binary.LittleEndian, int32(exp.Obj.Off))
		case typ.I64, typ.U64, typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LDII))
			binary.Write(asm.src, binary.LittleEndian, 0)
			binary.Write(asm.src, binary.LittleEndian, int32(exp.Obj.Off))
		}

	case *CompImmExpr:
		for _, cf := range exp.Fields {
			cf.Val.Accept(asm)
		}

	case *IdfExpr, *DotExpr:
		asm.getPosition(exp)

		t := exp.Type()[0]

		switch t.Kind {
		case typ.BOOL:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LBSS))
		case typ.I64, typ.U64, typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LDSS))
		case typ.FUNC:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LWSS))
		case typ.COMP:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LASS))
			binary.Write(asm.src, binary.LittleEndian, uint32(t.Size()))
		}

	case *InfExpr:
		cmd := byte(0)
		rev := false

		switch exp.Tok.TokenType {
		case lex.ADD:
			cmd = byte(run.ADD | exp.X.Type()[0].Kind)
		case lex.SUB:
			cmd = byte(run.SUB | exp.X.Type()[0].Kind)
		case lex.MUL:
			cmd = byte(run.MUL | exp.X.Type()[0].Kind)
		case lex.DIV:
			cmd = byte(run.DIV | exp.X.Type()[0].Kind)
		case lex.POW:
			cmd = byte(run.POW | exp.X.Type()[0].Kind)
		case lex.SHL:
			cmd = byte(run.SHL | exp.X.Type()[0].Kind)
		case lex.SHR:
			cmd = byte(run.SHR | exp.X.Type()[0].Kind)
		case lex.MOD:
			cmd = byte(run.MOD | exp.X.Type()[0].Kind)
		case lex.BAND, lex.LAND:
			cmd = byte(run.AND | exp.X.Type()[0].Kind)
		case lex.BOR, lex.LOR:
			cmd = byte(run.OR | exp.X.Type()[0].Kind)
		case lex.BXOR:
			cmd = byte(run.XOR | exp.X.Type()[0].Kind)
		case lex.LT:
			cmd = byte(run.LT | exp.X.Type()[0].Kind)
		case lex.LE:
			cmd = byte(run.LE | exp.X.Type()[0].Kind)
		case lex.GT:
			cmd = byte(run.LE | exp.X.Type()[0].Kind)
			rev = true
		case lex.GE:
			cmd = byte(run.LT | exp.X.Type()[0].Kind)
			rev = true
		case lex.EEQ:
			cmd = byte(run.EQ | exp.X.Type()[0].Kind)
		case lex.NE:
			cmd = byte(run.NE | exp.X.Type()[0].Kind)
		}

		if rev {
			exp.X.Accept(asm)
			exp.Y.Accept(asm)
		} else {
			exp.Y.Accept(asm)
			exp.X.Accept(asm)
		}

		asm.src.WriteByte(cmd)

	case *PfxExpr:
		exp.X.Accept(asm)

		switch exp.Tok.TokenType {
		case lex.BNEG, lex.LNEG:
			asm.src.WriteByte(byte(run.BNEG) | byte(exp.X.Type()[0].Kind))
		case lex.UNEG:
			asm.src.WriteByte(byte(run.UNEG) | byte(exp.X.Type()[0].Kind))
		}

	case *CallExpr:
		for _, arg := range exp.Arg {
			arg.Accept(asm)
		}

		exp.Exe.Accept(asm)
		asm.src.WriteByte(byte(run.CALLS))

	case *ToSigExpr:
		exp.X.Accept(asm)

		switch exp.X.Type()[0].Kind {
		case typ.U64:
			binary.Write(asm.src, binary.LittleEndian, run.U2I)
		case typ.F64:
			binary.Write(asm.src, binary.LittleEndian, run.F2I)
		}

	case *ToUnsExpr:
		exp.X.Accept(asm)

		switch exp.X.Type()[0].Kind {
		case typ.I64:
			binary.Write(asm.src, binary.LittleEndian, run.I2U)
		case typ.F64:
			binary.Write(asm.src, binary.LittleEndian, run.F2U)
		}

	case *ToFltExpr:
		exp.X.Accept(asm)

		switch exp.X.Type()[0].Kind {
		case typ.I64:
			binary.Write(asm.src, binary.LittleEndian, run.I2F)
		case typ.U64:
			binary.Write(asm.src, binary.LittleEndian, run.U2F)
		}
	}
}
