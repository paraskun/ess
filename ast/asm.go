package ast

import (
	"bytes"
	"encoding/binary"

	"github.com/paraskun/ess-go/lex"
	"github.com/paraskun/ess-go/run"
	"github.com/paraskun/ess-go/typ"
)

func Assemble(p *Package) *run.Package {
	a := assembler{
		pkg: &run.Package{
			Map: make(map[string]int),
		},
	}

	for _, d := range p.Dec {
		d.Accept(&a)
	}

	return a.pkg
}

type assembler struct {
	pkg *run.Package
	img *run.Image
	src *bytes.Buffer
}

func (asm *assembler) VisitDecl(u Decl) {
	switch dec := u.(type) {
	case *FuncDecl:
		src := &bytes.Buffer{}
		img := run.Image{DatSz: 16}

		asm.img = &img
		asm.src = src

		for _, arg := range dec.Spec.Arg {
			argSz := arg.Obj.Size()

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

		if img.ArgSz != 0 {
			binary.Write(asm.src, binary.LittleEndian, byte(run.SAII))
			binary.Write(asm.src, binary.LittleEndian, uint32(0))
			binary.Write(asm.src, binary.LittleEndian, uint32(16))
			binary.Write(asm.src, binary.LittleEndian, uint32(img.ArgSz))
		}

		dec.Body.Accept(asm)
		binary.Write(asm.src, binary.LittleEndian, byte(run.RET))

		img.Imm = buf.Bytes()
		img.Src = src.Bytes()

		asm.pkg.Img = append(asm.pkg.Img, img)

		if dec.Pkg {
			asm.pkg.Map[dec.Tok.Lit] = len(asm.pkg.Img) - 1
		}
	}
}

func (asm *assembler) VisitStmt(u Stmt) {
	switch s := u.(type) {
	case *BlockStmt:
		for _, b := range s.Body {
			b.Accept(asm)
		}
	case *AssignStmt:
		if s.Ini {
			binary.Write(asm.src, binary.LittleEndian, byte(run.LBII))
			binary.Write(asm.src, binary.LittleEndian, uint32(8))
			binary.Write(asm.src, binary.LittleEndian, uint32(0))

		}

		src := asm.src
		asm.src = &bytes.Buffer{}

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

		if s.Ini {
			binary.Write(src, binary.LittleEndian, byte(run.JIF))
			binary.Write(src, binary.LittleEndian, int32(asm.src.Len()))
		}

		binary.Write(src, binary.LittleEndian, asm.src.Bytes())
		asm.src = src
	case *LoopStmt:
		cur := asm.src.Len()

		s.Con.Accept(asm)

		con := asm.src.Len() - cur
		src := asm.src
		asm.src = &bytes.Buffer{}

		s.Rep.Accept(asm)

		binary.Write(src, binary.LittleEndian, byte(run.JIF))
		binary.Write(src, binary.LittleEndian, int32(asm.src.Len())+5)
		binary.Write(src, binary.LittleEndian, asm.src.Bytes())
		binary.Write(src, binary.LittleEndian, byte(run.JMP))
		binary.Write(src, binary.LittleEndian, int32(-asm.src.Len()-con-10))

		asm.src = src
	case *CondStmt:
		s.Con.Accept(asm)

		src := asm.src
		asm.src = &bytes.Buffer{}

		s.Pos.Accept(asm)

		binary.Write(src, binary.LittleEndian, byte(run.JIF))
		binary.Write(src, binary.LittleEndian, int32(asm.src.Len()+5))
		binary.Write(src, binary.LittleEndian, asm.src.Bytes())

		asm.src.Reset()

		if s.Neg != nil {
			s.Neg.Accept(asm)
		}

		binary.Write(src, binary.LittleEndian, byte(run.JMP))
		binary.Write(src, binary.LittleEndian, int32(asm.src.Len()))
		binary.Write(src, binary.LittleEndian, asm.src.Bytes())

		asm.src = src
	case *CallStmt:
		s.Exp.Accept(asm)
	case *ReturnStmt:
		if s.Ret.Type().Kind == typ.COMP {
			b, o := asm.getPosition(s.Ret)

			binary.Write(asm.src, binary.LittleEndian, byte(run.LEAII))
			binary.Write(asm.src, binary.LittleEndian, uint32(b))
			binary.Write(asm.src, binary.LittleEndian, uint32(o))
		} else {
			s.Ret.Accept(asm)
		}
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
			cmd = byte(run.LT | exp.X.Type().Kind)
			rev = true
		case lex.GE:
			cmd = byte(run.LE | exp.X.Type().Kind)
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
		inf := exp.Sym.Info.(*typ.FuncInfo)

		for i, arg := range exp.Arg {
			if arg.Type().Kind == typ.COMP {
				b, o := asm.getPosition(arg)

				binary.Write(asm.src, binary.LittleEndian, byte(run.LEAII))
				binary.Write(asm.src, binary.LittleEndian, uint32(b))
				binary.Write(asm.src, binary.LittleEndian, uint32(o))
			} else {
				arg.Accept(asm)
			}

			if inf.Arg[i].Typ.Kind == typ.ANY {
				asm.pushMeta(arg)
			}
		}

		idx := len(asm.img.CallInfo)
		asm.img.CallInfo = append(asm.img.CallInfo, inf.Off)

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

func (asm *assembler) getPosition(u Expr) (b int, o int) {
	switch exp := u.(type) {
	case *IdfExpr:
		if exp.Obj.Off == -1 {
			exp.Obj.Off = asm.img.DatSz
			asm.img.DatSz += exp.Obj.Size()
		}

		if exp.Obj.Ref {
			b = exp.Obj.Off
			o = 0
		} else {
			b = 0
			o = exp.Obj.Off
		}
	case *DotExpr:
		inf := exp.Comp.Type().Info.(*typ.CompInfo)
		b, o = asm.getPosition(exp.Comp)
		o += inf.Fields[exp.Field.Lit].Off
	}

	return b, o
}

func (asm *assembler) pushMeta(u Expr) {
	if u.Type().Kind == typ.COMP {
		panic("unsupported data type")
	}

	binary.Write(asm.src, binary.LittleEndian, byte(run.PUSHB))
	binary.Write(asm.src, binary.LittleEndian, byte(byte(u.Type().Kind)))
}
