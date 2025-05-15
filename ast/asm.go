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

func (a *assembler) VisitStmt(u Stmt) {
	switch s := u.(type) {
	case *BlockStmt:
	case *AssignStmt:
	case *LoopStmt:
	case *CondStmt:
	case *ReturnStmt:
	}
}

func (asm *assembler) VisitExpr(u Expr) {
	switch exp := u.(type) {
	case *BaseImmExpr:
	case *CompImmExpr:
	case *IdfExpr:
	case *DotExpr:
	case *InfExpr:
		exp.X.Accept(asm)
		exp.Y.Accept(asm)

		switch exp.Tok.TokenType {
		case lex.ADD:
			asm.src.WriteByte(byte(run.ADD | exp.X.Type()[0].Kind))
		case lex.SUB:
			asm.src.WriteByte(byte(run.SUB | exp.X.Type()[0].Kind))
		case lex.MUL:
			asm.src.WriteByte(byte(run.MUL | exp.X.Type()[0].Kind))
		case lex.DIV:
			asm.src.WriteByte(byte(run.DIV | exp.X.Type()[0].Kind))
		case lex.POW:
			asm.src.WriteByte(byte(run.POW | exp.X.Type()[0].Kind))
		case lex.SHL:
			asm.src.WriteByte(byte(run.SHL | exp.X.Type()[0].Kind))
		case lex.SHR:
			asm.src.WriteByte(byte(run.SHR | exp.X.Type()[0].Kind))
		case lex.MOD:
			asm.src.WriteByte(byte(run.MOD | exp.X.Type()[0].Kind))
		case lex.BAND, lex.LAND:
			asm.src.WriteByte(byte(run.AND | exp.X.Type()[0].Kind))
		case lex.BOR, lex.LOR:
			asm.src.WriteByte(byte(run.OR | exp.X.Type()[0].Kind))
		case lex.BXOR:
			asm.src.WriteByte(byte(run.XOR | exp.X.Type()[0].Kind))
		case lex.LT:
			asm.src.WriteByte(byte(run.LT | exp.X.Type()[0].Kind))
		case lex.LE:
			asm.src.WriteByte(byte(run.LE | exp.X.Type()[0].Kind))
		case lex.GT:
			asm.src.WriteByte(byte(run.GT | exp.X.Type()[0].Kind))
		case lex.GE:
			asm.src.WriteByte(byte(run.GE | exp.X.Type()[0].Kind))
		case lex.EEQ:
			asm.src.WriteByte(byte(run.EQ | exp.X.Type()[0].Kind))
		case lex.NE:
			asm.src.WriteByte(byte(run.NE | exp.X.Type()[0].Kind))
		}
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
