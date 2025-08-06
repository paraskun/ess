package ast

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/paraskun/o2/lex"
	"github.com/paraskun/o2/tty"
	"github.com/paraskun/o2/typ"
	"github.com/paraskun/o2/typ/mod"
)

type Visitor interface {
	VisitStmt(Stmt)
	VisitDecl(Decl)
	VisitExpr(Expr)
}

type Node interface {
	Span() tty.Span
	Accept(Visitor)
}

// Statements

type (
	Stmt interface {
		Node

		stmt()
	}

	BlockStmt struct {
		Box *tty.Box
		Sub []Stmt

		Env *typ.Env
	}

	VarStmt struct {
		Box *tty.Box
		Var *tty.Tok
		Ini Expr

		Obj *typ.Object
	}

	LetStmt struct {
		Box *tty.Box
		Var *tty.Tok
		Ini Expr

		Obj *typ.Object
	}

	AssignStmt struct {
		Box *tty.Box
		Var Expr
		Val Expr

		Obj *typ.Object
	}

	LoopStmt struct {
		Box *tty.Box
		Con Expr
		Sub *BlockStmt
	}

	CondStmt struct {
		Box *tty.Box
		Con Expr
		Pos *BlockStmt
		Neg *BlockStmt
	}

	CallStmt struct {
		*CallExpr
	}

	ReturnStmt struct {
		Box *tty.Box
		Ret Expr
	}
)

// Expressions

type (
	Expr interface {
		Node

		Object() *typ.Object

		expr()
	}

	BasicExpr struct {
		Tok *tty.Tok

		Obj *typ.Object
	}

	StructFieldExpr struct {
		Box *tty.Box
		Val Expr
	}

	StructExpr struct {
		Box *tty.Box
		Mem []StructFieldExpr

		Obj *typ.Object
	}

	IdenExpr struct {
		Tok *tty.Tok

		Obj *typ.Object
	}

	DotExpr struct {
		Row *tty.Row
		Mem *tty.Tok
		Ctx Expr

		Obj *typ.Object
	}

	InfExpr struct {
		Row *tty.Row
		X   Expr
		Y   Expr

		Res *typ.Object
	}

	PfxExpr struct {
		Row *tty.Row
		X   Expr

		Res *typ.Object
	}

	CallExpr struct {
		Box *tty.Box
		Sym *tty.Tok
		Arg []Expr

		Fun *typ.Object
		Ret *typ.Object
	}

	ToI64Expr struct {
		Row *tty.Row
		X   Expr

		Res *typ.Object
	}

	ToU64Expr struct {
		Row *tty.Row
		X   Expr

		Res *typ.Object
	}

	ToF64Expr struct {
		Row *tty.Row
		X   Expr

		Res *typ.Object
	}
)

// Declarations

type (
	Decl interface {
		Node

		decl()
	}

	UseDecl struct {
		Row *tty.Row
		Pkg *tty.Tok

		Obj *typ.Object
	}

	VarDecl struct {
		*VarStmt
	}

	TypeSpec struct {
		Tok *tty.Tok
		Typ *typ.Type
	}

	NamedField struct {
		Row *tty.Row
		Sym *tty.Tok
		Typ *TypeSpec
	}

	FuncDecl struct {
		Box *tty.Box
		Sig *tty.Row
		Sym *tty.Tok
		Arg []*NamedField
		Ret *TypeSpec
		Sub *BlockStmt

		Env *typ.Env
		Obj *typ.Object
	}

	StructDecl struct {
		Box *tty.Box
		Sym *tty.Tok
		Mem []*NamedField

		Obj *typ.Object
	}

	EnumDecl struct {
		Box *tty.Box
		Sym *tty.Tok
		Mem []*tty.Tok

		Obj *typ.Object
	}
)

type File struct {
	Box *tty.Box
	Dec []Decl
}

type options struct{}

type Option func(*options)

type parser struct {
	lex lex.Scanner

	buf bool
	prv *lex.Lexeme
	cur *lex.Lexeme

	ops *options
	pkg *mod.Package
	src *mod.File

	box []*tty.Box
	row []*tty.Row
}

func (p *parser) newBox(ind int) {
	box := &tty.Box{Ind: ind}

	if len(p.box) != 0 {
		out := p.box[len(p.box)-1]
		out.Sub = append(out.Sub, box)
	}

	p.box = append(p.box, box)
}

func (p *parser) newRow(ind int) {
	row := &tty.Row{Ind: ind}

	if len(p.row) != 0 {
		out := p.row[len(p.row)-1]
		out.Sub = append(out.Sub, row)
	} else if len(p.box) != 0 {
		out := p.box[len(p.box)-1]
		out.Sub = append(out.Sub, row)
	} else {
		panic("out of structure")
	}

	p.row = append(p.row, row)
}

func (p *parser) popBox() *tty.Box {
	if len(p.row) != 0 {
		panic("we are still have rows")
	}

	box := p.box[len(p.box)-1]
	p.box = p.box[:len(p.box)-1]

	return box
}

func (p *parser) popRow() *tty.Row {
	row := p.row[len(p.row)-1]
	p.row = p.row[:len(p.row)-1]

	return row
}

func (p *parser) next(ind int) *lex.Lexeme {
	if !p.buf {
		p.prv = p.cur
		p.cur = p.lex.Next()
	}

	p.buf = false
	p.cur.Tok.Ind = ind

	if len(p.row) != 0 {
		row := p.row[len(p.row)-1]
		row.Sub = append(row.Sub, p.cur.Tok)
	} else if len(p.box) != 0 {
		box := p.box[len(p.box)-1]
		box.Sub = append(box.Sub, p.cur.Tok)
	} else {
		panic("out of structure")
	}

	return p.cur
}

func (p *parser) peek() *lex.Lexeme {
	if !p.buf {
		p.buf = true
		p.prv = p.cur
		p.cur = p.lex.Next()
	}

	return p.cur
}

func (p *parser) must(t lex.Type, ind int) *lex.Lexeme {
	if p.next(ind).Type != t {
		panic(fmt.Errorf("%v given, but %v expected", p.cur.Type, t))
	}

	return p.cur
}

func (p *parser) note(n Node) {
	f := tty.Frame{
		Name: p.pkg.Path + "/" + p.src.Name,
		Sub:  n.Span(),
	}

	tty.Print(&f)
}

func Parse(pkg *mod.Package, opts ...Option) {
	ops := &options{}

	for _, opt := range opts {
		opt(ops)
	}

	parse(pkg, ops)
}

func parse(pkg *mod.Package, ops *options) {
	p := &parser{pkg: pkg, ops: ops}

	for _, src := range pkg.XSrc {
		fil, err := os.Open(src.Path)
		p.src = src

		if err != nil {
			panic(err)
		} else {
			defer fil.Close()
		}

		dec := &File{Dec: make([]Decl, 0)}
		buf, err := io.ReadAll(fil)

		if err != nil {
			panic(err)
		}

		p.lex.Load([]rune(string(buf)))
		p.newBox(0)

		for p.peek().Type != lex.EOF {
			dec.Dec = append(dec.Dec, p.parseDecl())
		}

		dec.Box = p.popBox()
		src.Dec = dec
	}
}

func (p *parser) parseDecl() Decl {
	switch p.peek().Type {
	case lex.USE:
		use := p.parseUseDecl()
		pkg := p.pkg.Mod.Lookup(use.Pkg.Lit)

		if pkg == nil {
			use.Row.Hint = &tty.Hint{
				Text: "unknown package",
				Attr: tty.Attr{
					Color: *color.New(color.FgRed),
				},
			}

			p.note(use)
		}

		use.Obj = &typ.Object{
			Typ: &typ.Type{
				Kind:  typ.PKG,
				Extra: pkg,
			},
		}

		parse(pkg, p.ops)

		return use
	case lex.VAR:
		return &VarDecl{VarStmt: p.parseVarStmt()}
	case lex.FUNC:
		return p.parseFuncDecl()
	case lex.TYPE:
		return p.parseStructDecl()
	case lex.ENUM:
		return p.parseEnumDecl()
	default:
		panic("declaration expected")
	}
}

func (p *parser) parseUseDecl() *UseDecl {
	p.newRow(0)
	p.must(lex.USE, 0)

	return &UseDecl{
		Pkg: p.must(lex.ISTR, 1).Tok,
		Row: p.popRow(),
	}
}

func (p *parser) parseFuncDecl() *FuncDecl {
	p.newBox(0)
	p.must(lex.FUNC, 0)
	p.newRow(0)
	p.newRow(1)

	res := &FuncDecl{Sym: p.must(lex.IDEN, 0).Tok}

	p.must(lex.LP, 0)

	for p.peek().Type != lex.RP {
		if len(res.Arg) == 0 {
			res.Arg = append(res.Arg, p.parseNamedField(0))
		} else {
			res.Arg = append(res.Arg, p.parseNamedField(1))
		}

		if p.peek().Type != lex.RP {
			p.must(lex.COM, 0)
		}
	}

	p.must(lex.RP, 0)

	if p.peek().Type == lex.LP {
		p.next(1)
		res.Ret = p.parseTypeSpec(0)
		p.must(lex.RP, 0)
	}

	res.Sig = p.popRow()

	if p.peek().Type == lex.LB {
		p.next(1)
		p.popRow()

		res.Sub = p.parseBlockStmt(2)

		p.must(lex.RB, 0)
	}

	return res
}

func (p *parser) parseStructDecl() *StructDecl {
	p.newBox(0)
	p.newRow(0)
	p.must(lex.TYPE, 0)

	c := &StructDecl{Sym: p.must(lex.IDEN, 1).Tok}

	p.must(lex.LB)

	for p.peek().TokenType != lex.RB {
		c.Mem = append(c.Mem, p.parseNamedField())
	}

	p.must(lex.RB)

	return c
}

func (p *parser) parseEnumDecl() *EnumDecl {
	c := &EnumDecl{
		Tok: p.must(lex.ENUM),
		Idf: p.must(lex.IDF),
	}

	p.must(lex.LB)

	for p.peek().TokenType != lex.RB {
		c.Mem = append(c.Mem, p.must(lex.IDF))
	}

	p.must(lex.RB)

	return c
}

func (p *parser) parseNamedField(ind int) *NamedField {
	p.newRow(ind)

	return &NamedField{
		Sym: p.must(lex.IDEN, 0).Tok,
		Typ: p.parseTypeSpec(1),
		Row: p.popRow(),
	}
}

func (p *parser) parseTypeSpec(ind int) *TypeSpec {
	switch p.peek().Type {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDEN:
		return &TypeSpec{Tok: p.next(ind).Tok}
	}

	// TODO: error handling
	panic("type specification expected")
}

func (p *parser) parseBlockStmt(ind int) *BlockStmt {
	p.newBox(ind)

	r := BlockStmt{
		Tok: p.must(lex.LB),
	}

	for p.peek().TokenType != lex.RB {
		r.Body = append(r.Body, p.parseStmt())
	}

	p.must(lex.RB)

	return &r
}

func (p *parser) parseStmt() Stmt {
	switch p.peek().TokenType {
	case lex.IDF:
		idf := p.next()

		if p.peek().TokenType == lex.LP {
			exe := p.parseCall(idf)
			p.must(lex.SEM)

			return &CallStmt{exe}
		}

		var cur Expr = &IdfExpr{Tok: idf}

		for {
			switch p.peek().TokenType {
			case lex.DOT:
				cur = &DotExpr{
					Tok: p.next(),
					Ctx: cur,
					Mem: p.must(lex.IDF),
				}

				continue
			}

			break
		}

		r := &AssignStmt{
			Tok: p.must(lex.EQ),
			Var: cur,
			Val: p.parseExpr0(),
		}

		p.must(lex.SEM)

		return r
	case lex.VAR:
		return p.parseVarStmt()
	case lex.LET:
		return p.parseLetStmt()
	case lex.FOR:
		return p.parseLoopStmt()
	case lex.IF:
		return p.parseCondStmt()
	case lex.RET:
		return p.parseReturnStmt()
	}

	panic("statement expected")
}

func (p *parser) parseCall(tok *lex.Token) *CallExpr {
	p.next()

	exe := &CallExpr{Tok: tok}

	for p.peek().TokenType != lex.RP {
		exe.Arg = append(exe.Arg, p.parseExpr0())

		if p.peek().TokenType != lex.RP {
			p.must(lex.COM)
		}
	}

	p.must(lex.RP)

	return exe
}

func (p *parser) parseVarStmt() *VarStmt {
	v := &VarStmt{
		Tok: p.must(lex.VAR),
		Idf: p.must(lex.IDF),
	}

	p.must(lex.EQ)
	v.Ini = p.parseExpr0()

	return v
}

func (p *parser) parseLetStmt() *LetStmt {
	v := &LetStmt{
		Tok: p.must(lex.LET),
		Idf: p.must(lex.IDF),
	}

	p.must(lex.EQ)
	v.Ini = p.parseExpr0()

	return v
}

func (p *parser) parseLoopStmt() *LoopStmt {
	r := LoopStmt{
		Tok: p.must(lex.FOR),
	}

	r.Con = p.parseExpr0()
	r.Rep = p.parseBlockStmt()

	return &r
}

func (p *parser) parseCondStmt() *CondStmt {
	r := CondStmt{
		Tok: p.must(lex.IF),
	}

	r.Con = p.parseExpr0()
	r.Pos = p.parseBlockStmt()

	if p.peek().TokenType == lex.ELSE {
		p.next()

		r.Neg = p.parseBlockStmt()
	}

	return &r
}

func (p *parser) parseReturnStmt() *ReturnStmt {
	r := ReturnStmt{Tok: p.must(lex.RET)}

	for p.peek().TokenType != lex.SEM {
		r.Ret = p.parseExpr0()

		if p.peek().TokenType != lex.SEM {
			p.must(lex.COM)
		}
	}

	p.must(lex.SEM)

	return &r
}

func (p *parser) parseExpr0() Expr {
	cur := p.parseExpr1()

	for {
		switch p.peek().TokenType {
		case lex.LAND, lex.LOR:
			cur = &InfExpr{
				Tok: p.next(),
				X:   cur,
				Y:   p.parseExpr1(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *parser) parseExpr1() Expr {
	cur := p.parseExpr2()

	for {
		switch p.peek().TokenType {
		case lex.LT, lex.GT, lex.LE, lex.GE, lex.NE, lex.EEQ:
			cur = &InfExpr{
				Tok: p.next(),
				X:   cur,
				Y:   p.parseExpr2(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *parser) parseExpr2() Expr {
	cur := p.parseExpr3()

	for {
		switch p.peek().TokenType {
		case lex.SHL, lex.SHR, lex.BAND, lex.BOR, lex.BXOR:
			cur = &InfExpr{
				Tok: p.next(),
				X:   cur,
				Y:   p.parseExpr3(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *parser) parseExpr3() Expr {
	cur := p.parseExpr4()

	for {
		switch p.peek().TokenType {
		case lex.ADD, lex.SUB:
			cur = &InfExpr{
				Tok: p.next(),
				X:   cur,
				Y:   p.parseExpr4(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *parser) parseExpr4() Expr {
	cur := p.parseExpr5()

	for {
		switch p.peek().TokenType {
		case lex.MUL, lex.DIV, lex.MOD, lex.POW:
			cur = &InfExpr{
				Tok: p.next(),
				X:   cur,
				Y:   p.parseExpr5(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *parser) parseExpr5() Expr {
	switch p.peek().TokenType {
	case lex.LNEG, lex.BNEG, lex.UNEG:
		return &PfxExpr{
			Tok: p.next(),
			X:   p.parseExpr5(),
		}
	}

	return p.parseExpr6()
}

func (p *parser) parseExpr6() Expr {
	cur := p.parseExpr7()

	for {
		switch p.peek().TokenType {
		case lex.DOT:
			cur = &DotExpr{
				Tok: p.next(),
				Ctx: cur,
				Mem: p.must(lex.IDF),
			}

			continue
		}

		break
	}

	return cur
}

func (p *parser) parseExpr7() Expr {
	switch p.peek().TokenType {
	case lex.IDF:
		idf := p.next()

		switch p.peek().TokenType {
		case lex.LP:
			return p.parseCall(idf)
		case lex.LB:
			p.next()

			exp := &StructExpr{Tok: idf}

			for p.peek().TokenType != lex.RB {
				tok := p.must(lex.IDF)
				p.must(lex.COL)
				val := p.parseExpr0()

				exp.Mem = append(exp.Mem, StructField{
					Tok: tok,
					Val: val,
				})

				if p.peek().TokenType != lex.RB {
					p.must(lex.COM)
				}
			}

			p.must(lex.RB)

			return exp
		}

		return &IdfExpr{Tok: idf}
	case lex.II64, lex.IU64, lex.IF64, lex.TRUE, lex.FALSE:
		return &ImmExpr{Tok: p.next()}
	case lex.I64:
		exp := &ToI64Expr{Tok: p.next()}

		p.must(lex.LP)
		exp.X = p.parseExpr0()
		p.must(lex.RP)

		return exp
	case lex.U64:
		exp := &ToU64Expr{Tok: p.next()}

		p.must(lex.LP)
		exp.X = p.parseExpr0()
		p.must(lex.RP)

		return exp
	case lex.F64:
		exp := &ToF64Expr{Tok: p.next()}

		p.must(lex.LP)
		exp.X = p.parseExpr0()
		p.must(lex.RP)

		return exp
	case lex.LP:
		p.next()
		r := p.parseExpr0()
		p.must(lex.RP)

		return r
	}

	panic("expression expected")
}

func (s *ReturnStmt) Span() tty.Span { return s.Box }
func (s *VarStmt) Span() tty.Span    { return s.Box }
func (s *LetStmt) Span() tty.Span    { return s.Box }
func (s *AssignStmt) Span() tty.Span { return s.Box }
func (s *BlockStmt) Span() tty.Span  { return s.Box }
func (s *LoopStmt) Span() tty.Span   { return s.Box }
func (s *CondStmt) Span() tty.Span   { return s.Box }
func (s *CallStmt) Span() tty.Span   { return s.Box }

func (s *ReturnStmt) Accept(v Visitor) { v.VisitStmt(s) }
func (s *VarStmt) Accept(v Visitor)    { v.VisitStmt(s) }
func (s *LetStmt) Accept(v Visitor)    { v.VisitStmt(s) }
func (s *AssignStmt) Accept(v Visitor) { v.VisitStmt(s) }
func (s *BlockStmt) Accept(v Visitor)  { v.VisitStmt(s) }
func (s *LoopStmt) Accept(v Visitor)   { v.VisitStmt(s) }
func (s *CondStmt) Accept(v Visitor)   { v.VisitStmt(s) }
func (s *CallStmt) Accept(v Visitor)   { v.VisitStmt(s) }

func (*ReturnStmt) stmt() {}
func (*VarStmt) stmt()    {}
func (*LetStmt) stmt()    {}
func (*AssignStmt) stmt() {}
func (*BlockStmt) stmt()  {}
func (*LoopStmt) stmt()   {}
func (*CondStmt) stmt()   {}
func (*CallStmt) stmt()   {}

func (e *BasicExpr) Span() tty.Span  { return e.Tok }
func (e *StructExpr) Span() tty.Span { return e.Box }
func (e *IdenExpr) Span() tty.Span   { return e.Tok }
func (e *DotExpr) Span() tty.Span    { return e.Row }
func (e *InfExpr) Span() tty.Span    { return e.Row }
func (e *PfxExpr) Span() tty.Span    { return e.Row }
func (e *CallExpr) Span() tty.Span   { return e.Box }
func (e *ToI64Expr) Span() tty.Span  { return e.Row }
func (e *ToU64Expr) Span() tty.Span  { return e.Row }
func (e *ToF64Expr) Span() tty.Span  { return e.Row }

func (e *BasicExpr) Accept(v Visitor)  { v.VisitExpr(e) }
func (e *StructExpr) Accept(v Visitor) { v.VisitExpr(e) }
func (e *IdenExpr) Accept(v Visitor)   { v.VisitExpr(e) }
func (e *DotExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *InfExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *PfxExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *CallExpr) Accept(v Visitor)   { v.VisitExpr(e) }
func (e *ToI64Expr) Accept(v Visitor)  { v.VisitExpr(e) }
func (e *ToU64Expr) Accept(v Visitor)  { v.VisitExpr(e) }
func (e *ToF64Expr) Accept(v Visitor)  { v.VisitExpr(e) }

func (e *BasicExpr) Object() *typ.Object  { return e.Obj }
func (e *StructExpr) Object() *typ.Object { return e.Obj }
func (e *IdenExpr) Object() *typ.Object   { return e.Obj }
func (e *DotExpr) Object() *typ.Object    { return e.Obj }
func (e *InfExpr) Object() *typ.Object    { return e.Res }
func (e *PfxExpr) Object() *typ.Object    { return e.Res }
func (e *CallExpr) Object() *typ.Object   { return e.Ret }
func (e *ToI64Expr) Object() *typ.Object  { return e.Res }
func (e *ToU64Expr) Object() *typ.Object  { return e.Res }
func (e *ToF64Expr) Object() *typ.Object  { return e.Res }

func (*BasicExpr) expr()  {}
func (*StructExpr) expr() {}
func (*IdenExpr) expr()   {}
func (*DotExpr) expr()    {}
func (*InfExpr) expr()    {}
func (*PfxExpr) expr()    {}
func (*CallExpr) expr()   {}
func (*ToI64Expr) expr()  {}
func (*ToU64Expr) expr()  {}
func (*ToF64Expr) expr()  {}

func (d *UseDecl) Span() tty.Span    { return d.Row }
func (d *VarDecl) Span() tty.Span    { return d.Box }
func (d *FuncDecl) Span() tty.Span   { return d.Box }
func (d *StructDecl) Span() tty.Span { return d.Box }
func (d *EnumDecl) Span() tty.Span   { return d.Box }

func (d *UseDecl) Accept(v Visitor)    { v.VisitDecl(d) }
func (d *VarDecl) Accept(v Visitor)    { v.VisitDecl(d) }
func (d *FuncDecl) Accept(v Visitor)   { v.VisitDecl(d) }
func (d *StructDecl) Accept(v Visitor) { v.VisitDecl(d) }
func (d *EnumDecl) Accept(v Visitor)   { v.VisitDecl(d) }

func (*UseDecl) decl()    {}
func (*VarDecl) decl()    {}
func (*FuncDecl) decl()   {}
func (*StructDecl) decl() {}
func (*EnumDecl) decl()   {}
