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
		obj: &run.Pragma{},
	}

	for _, d := range p.Dec {
		d.Accept(&a)
	}

	return a.obj
}

type assembler struct {
	obj *run.Pragma
	img *run.FuncImage
	src *bytes.Buffer
}

func (asm *assembler) VisitDecl(u Decl) {
	switch dec := u.(type) {
	case *FuncDecl:
		img := run.FuncImage{}
		src := &bytes.Buffer{}

		asm.img = &img

		for _, arg := range dec.Spec.Arg {
			argSz := arg.Obj.Typ.Size()

			arg.Obj.Off = img.DatSz
			img.DatSz += argSz
			img.ArgSz += argSz
		}

		buf := &bytes.Buffer{}
		off := 0

		for _, obj := range dec.Env.Imm {
			obj.Off = off
			off += obj.Size()

			binary.Write(buf, binary.LittleEndian, obj.Val)
		}

		dec.Body.Accept(asm)

		img.DatSz += 8
		img.Imm = buf.Bytes()
		img.Src = src.Bytes()

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
		s.Val.Accept(asm)

		b, o := asm.getPosition(s.Var)

		switch s.Var.Type().Kind {
		case typ.BOOL:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SBII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
		case typ.I64, typ.U64, typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SDII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
		case typ.FUNC:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SWII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
		case typ.COMP:
			binary.Write(asm.src, binary.LittleEndian, byte(run.SAII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
			binary.Write(asm.src, binary.LittleEndian, uint32(s.Var.Type().Size()))
		}
	case *LoopStmt:
	case *CondStmt:
	case *ReturnStmt:
	}
}

func (asm *assembler) VisitExpr(u Expr) {
	switch exp := u.(type) {
	case *BaseImmExpr:
		switch exp.Type().Kind {
		case typ.BOOL:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LBI))
			binary.Write(asm.src, binary.LittleEndian, uint32(exp.Obj.Off))
		case typ.I64, typ.U64, typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LDI))
			binary.Write(asm.src, binary.LittleEndian, uint32(exp.Obj.Off))
		}
	case *CompImmExpr:
		for _, f := range exp.Fields {
			f.Val.Accept(asm)
		}
	case *IdfExpr, *DotExpr:
		b, o := asm.getPosition(exp)

		switch exp.Type().Kind {
		case typ.BOOL:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LBII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
		case typ.I64, typ.U64, typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LDII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
		case typ.COMP:
			binary.Write(asm.src, binary.LittleEndian, byte(run.LAII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
			binary.Write(asm.src, binary.LittleEndian, uint32(exp.Type().Size()))
		}
	case *InfExpr:
		cmd := byte(0)
		rev := false

		switch exp.Tok.TokenType {
		case lex.ADD:
			cmd = byte(run.ADD | exp.X.Type().Kind)
		case lex.SUB:
			cmd = byte(run.SUB | exp.X.Type().Kind)
		case lex.MUL:
			cmd = byte(run.MUL | exp.X.Type().Kind)
		case lex.DIV:
			cmd = byte(run.DIV | exp.X.Type().Kind)
		case lex.POW:
			cmd = byte(run.POW | exp.X.Type().Kind)
		case lex.SHL:
			cmd = byte(run.SHL | exp.X.Type().Kind)
		case lex.SHR:
			cmd = byte(run.SHR | exp.X.Type().Kind)
		case lex.MOD:
			cmd = byte(run.MOD | exp.X.Type().Kind)
		case lex.BAND, lex.LAND:
			cmd = byte(run.AND | exp.X.Type().Kind)
		case lex.BOR, lex.LOR:
			cmd = byte(run.OR | exp.X.Type().Kind)
		case lex.BXOR:
			cmd = byte(run.XOR | exp.X.Type().Kind)
		case lex.LT:
			cmd = byte(run.LT | exp.X.Type().Kind)
		case lex.LE:
			cmd = byte(run.LE | exp.X.Type().Kind)
		case lex.GT:
			cmd = byte(run.LE | exp.X.Type().Kind)
			rev = true
		case lex.GE:
			cmd = byte(run.LT | exp.X.Type().Kind)
			rev = true
		case lex.EEQ:
			cmd = byte(run.EQ | exp.X.Type().Kind)
		case lex.NE:
			cmd = byte(run.NE | exp.X.Type().Kind)
		}

		if rev {
			exp.X.Accept(asm)
			exp.Y.Accept(asm)
		} else {
			exp.Y.Accept(asm)
			exp.X.Accept(asm)
		}

		asm.src.WriteByte(byte(cmd))
	case *PfxExpr:
		exp.X.Accept(asm)

		switch exp.Tok.TokenType {
		case lex.BNEG, lex.LNEG:
			asm.src.WriteByte(byte(run.BNEG) | byte(exp.X.Type().Kind))
		case lex.UNEG:
			asm.src.WriteByte(byte(run.UNEG) | byte(exp.X.Type().Kind))
		}
	case *CallExpr:
		for _, arg := range exp.Arg {
			arg.Accept(asm)
		}

		idx := len(asm.img.Call)
		asm.img.Call = append(asm.img.Call, exp.Obj.Off)

		binary.Write(asm.src, binary.LittleEndian, byte(run.CALL))
		binary.Write(asm.src, binary.LittleEndian, uint32(idx))
	case *ToSigExpr:
		exp.X.Accept(asm)

		switch exp.X.Type().Kind {
		case typ.U64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.U2I))
		case typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.F2I))
		}
	case *ToUnsExpr:
		exp.X.Accept(asm)

		switch exp.X.Type().Kind {
		case typ.I64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.I2U))
		case typ.F64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.F2U))
		}
	case *ToFltExpr:
		exp.X.Accept(asm)

		switch exp.X.Type().Kind {
		case typ.I64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.I2F))
		case typ.U64:
			binary.Write(asm.src, binary.LittleEndian, byte(run.U2F))
		}
	}
}

func (asm *assembler) getPosition(u Expr) (int, int) {
	switch exp := u.(type) {
	case *IdfExpr:
		_ = exp
	case *DotExpr:
		_ = exp
	}

	return 0, 0
}
