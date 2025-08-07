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

type Collection interface {
	Add(s tty.Span)
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
	col []Collection
}

func (p *parser) newBox(rowInd, boxInd int) {
	box := &tty.Box{Pos: tty.Position{
		Ind: tty.Indent{Row: rowInd, Box: boxInd},
	}}

	if len(p.col) != 0 {
		p.col[len(p.col)-1].Add(box)
	}

	p.col = append(p.col, box)
}

func (p *parser) newRow(rowInd, boxInd int) {
	box := &tty.Box{Pos: tty.Position{
		Ind: tty.Indent{Row: rowInd, Box: boxInd},
	}}

	if len(p.col) != 0 {
		p.col[len(p.col)-1].Add(box)
	}

	p.col = append(p.col, box)
}

func (p *parser) popBox() *tty.Box {
	if len(p.col) == 0 {
		panic("out of structure")
	}

	box := p.col[len(p.col)-1].(*tty.Box)
	p.col = p.col[:len(p.col)-1]

	return box
}

func (p *parser) popRow() *tty.Row {
	if len(p.col) == 0 {
		panic("out of structure")
	}

	row := p.col[len(p.col)-1].(*tty.Row)
	p.col = p.col[:len(p.col)-1]

	return row
}

func (p *parser) scan() *lex.Lexeme {
	var err error
	var lex *lex.Lexeme

	for {
		lex, err = p.lex.Next()

		if err == nil {
			break
		}

		fmt.Println(err)
	}

	return lex
}

func (p *parser) next(rowInd, boxInd int) *lex.Lexeme {
	if !p.buf {
		p.prv = p.cur
		p.cur = p.scan()
	}

	p.buf = false
	p.cur.Tok.Pos.Ind = tty.Indent{Row: rowInd, Box: boxInd}
	p.col[len(p.col)-1].Add(p.cur.Tok)

	return p.cur
}

func (p *parser) peek() *lex.Lexeme {
	if !p.buf {
		p.buf = true
		p.prv = p.cur
		p.cur = p.scan()
	}

	return p.cur
}

func (p *parser) must(t lex.Type, rowInd, boxInd int) *lex.Lexeme {
	if p.next(rowInd, boxInd).Typ != t {
		panic(fmt.Errorf("%v given, but %v expected", p.cur.Typ, t))
	}

	return p.cur
}

func (p *parser) note(n Node) {
	f := tty.Frame{
		Name: p.pkg.Path + "/" + p.src.Name,
		Span: n.Span(),
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
		p.newBox(0, 0)

		for p.peek().Typ != lex.EOF {
			dec.Dec = append(dec.Dec, p.parseDecl())
		}

		dec.Box = p.popBox()
		src.Dec = dec
	}
}

func (p *parser) parseDecl() Decl {
	switch p.peek().Typ {
	case lex.USE:
		use := p.parseUseDecl()
		pkg := p.pkg.Mod.Lookup(use.Pkg.Lit)

		if pkg == nil {
			use.Row.Hint.Text = "unknown package"
			use.Row.Hint.Attr = tty.Attr{
				Color: *color.New(color.FgRed),
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
	p.newRow(0, 0)
	p.must(lex.USE, 0, 0)

	return &UseDecl{
		Pkg: p.must(lex.ISTR, 1, 0).Tok,
		Row: p.popRow(),
	}
}

func (p *parser) parseFuncDecl() *FuncDecl {
	p.newBox(0, 1)
	p.newRow(0, 0)
	p.must(lex.FUNC, 0, 0)
	p.newRow(1, 0)

	res := &FuncDecl{Sym: p.must(lex.IDEN, 0, 0).Tok}

	p.must(lex.LP, 0, 0)

	for p.peek().Typ != lex.RP {
		if len(res.Arg) == 0 {
			res.Arg = append(res.Arg, p.parseNamedField(0, 0))
		} else {
			res.Arg = append(res.Arg, p.parseNamedField(1, 0))
		}

		if p.peek().Typ != lex.RP {
			p.must(lex.COM, 0, 0)
		}
	}

	p.must(lex.RP, 0, 0)

	if p.peek().Typ == lex.LP {
		p.next(1, 0)
		res.Ret = p.parseTypeSpec(0, 0)
		p.must(lex.RP, 0, 0)
	}

	res.Sig = p.popRow()

	if p.peek().Typ == lex.LB {
		p.next(1, 0)
		p.popRow()

		res.Sub = p.parseBlockStmt(2, 0)

		p.must(lex.RB, 0, 0)
	} else {
		p.popRow()
	}

	res.Box = p.popBox()

	return res
}

func (p *parser) parseStructDecl() *StructDecl {
	p.newBox(0, 1)
	p.newRow(0, 0)
	p.must(lex.TYPE, 0, 0)

	res := &StructDecl{Sym: p.must(lex.IDEN, 1, 0).Tok}

	p.must(lex.LB, 1, 0)
	p.popRow()

	for p.peek().Typ != lex.RB {
		res.Mem = append(res.Mem, p.parseNamedField(2, 0))
	}

	p.must(lex.RB, 0, 0)
	res.Box = p.popBox()

	return res
}

func (p *parser) parseEnumDecl() *EnumDecl {
	p.newBox(0, 1)
	p.newRow(0, 0)
	p.must(lex.ENUM, 0, 0)

	res := &EnumDecl{Sym: p.must(lex.IDEN, 1, 0).Tok}

	p.must(lex.LB, 1, 0)
	p.popRow()

	for p.peek().Typ != lex.RB {
		res.Mem = append(res.Mem, p.must(lex.IDEN, 2, 0).Tok)
	}

	p.must(lex.RB, 0, 0)
	res.Box = p.popBox()

	return res
}

func (p *parser) parseNamedField(rowInd, boxInd int) *NamedField {
	p.newRow(rowInd, boxInd)

	return &NamedField{
		Sym: p.must(lex.IDEN, 0, 0).Tok,
		Typ: p.parseTypeSpec(1, 0),
		Row: p.popRow(),
	}
}

func (p *parser) parseTypeSpec(rowInd, boxInd int) *TypeSpec {
	switch p.peek().Typ {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDEN:
		return &TypeSpec{Tok: p.next(rowInd, boxInd).Tok}
	}

	// TODO: error handling
	panic("type specification expected")
}

func (p *parser) parseBlockStmt(rowInd, boxInd int) *BlockStmt {
	p.newBox(rowInd, boxInd)

	res := &BlockStmt{}

	for p.peek().Typ != lex.RB {
		res.Sub = append(res.Sub, p.parseStmt(0, 0))
	}

	res.Box = p.popBox()

	return res
}

func (p *parser) parseStmt(rowInd, boxInd int) Stmt {
	switch p.peek().Typ {
	case lex.IDEN:
		p.newBox(rowInd, boxInd)
		p.newRow(0, 0)

		iden := p.next(0, 0)

		if p.peek().Typ == lex.LP {
			exe := p.parseCall(iden.Tok)

			p.must(lex.SEM, 0, 0)
			p.popRow()

			exe.Box = p.popBox()

			return &CallStmt{exe}
		}

		var cur Expr = &IdenExpr{Tok: iden.Tok}

		for {
			switch p.peek().Typ {
			case lex.DOT:
				cur = &DotExpr{
					Ctx: cur,
					Mem: p.must(lex.IDEN, 0, 0).Tok,
					Row: p.popRow(),
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
		return p.parseVarStmt(rowInd, boxInd)
	case lex.LET:
		return p.parseLetStmt(rowInd, boxInd)
	case lex.FOR:
		return p.parseLoopStmt(rowInd, boxInd)
	case lex.IF:
		return p.parseCondStmt(rowInd, boxInd)
	case lex.RET:
		return p.parseReturnStmt(rowInd, boxInd)
	}

	panic("statement expected")
}

func (p *parser) parseCall(tok *tty.Tok) *CallExpr {
	exe := &CallExpr{Sym: tok}

	p.next(0, 0)
	p.popRow()

	for p.peek().Typ != lex.RP {
		exe.Arg = append(exe.Arg, p.parseExpr0(2, 0))

		if p.peek().Typ != lex.RP {
			p.must(lex.COM, 2, 0)
		}
	}

	p.must(lex.RP, 0, 0)

	return exe
}

func (p *parser) parseVarStmt(rowInd, boxInd int) *VarStmt {
	p.newBox(rowInd, boxInd)
	p.newRow(0, 0)
	p.must(lex.VAR, 0, 0)

	r := VarStmt{Var: p.must(lex.IDEN, 1, 0).Tok}

	p.must(lex.EQ, 1, 0)
	p.popRow()

	r.Ini = p.parseExpr0(2, 0)
	p.must(lex.SEM, 0, 0)
	r.Box = p.popBox()

	return &r
}

func (p *parser) parseLetStmt(rowInd, boxInd int) *LetStmt {
	p.newBox(rowInd, boxInd)
	p.newRow(0, 0)
	p.must(lex.LET, 0, 0)

	r := LetStmt{Var: p.must(lex.IDEN, 1, 0).Tok}

	p.must(lex.EQ, 1, 0)
	p.popRow()

	r.Ini = p.parseExpr0(2, 0)
	p.must(lex.SEM, 0, 0)
	r.Box = p.popBox()

	return &r
}

func (p *parser) parseLoopStmt(rowInd, boxInd int) *LoopStmt {
	p.newBox(rowInd, boxInd)
	p.newRow(0, 0)
	p.must(lex.FOR, 0, 0)
	p.popRow()

	r := LoopStmt{Con: p.parseExpr0(2, 0)}

	p.must(lex.LB, 0, 0)
	r.Sub = p.parseBlockStmt(2, 0)
	p.must(lex.RB, 0, 0)

	r.Box = p.popBox()

	return &r
}

func (p *parser) parseCondStmt(rowInd, boxInd int) *CondStmt {
	p.newBox(rowInd, boxInd)
	p.newRow(0, 0)
	p.must(lex.IF, 0, 0)
	p.popRow()

	r := CondStmt{Con: p.parseExpr0()}

	p.must(lex.LB, 0, 0)
	r.Pos = p.parseBlockStmt(2, 0)

	p.newRow(0, 0)
	p.must(lex.RB, 0, 0)

	if p.peek().Typ == lex.ELSE {
		p.next(1, 0)
		p.must(lex.LB, 1, 0)
		p.popRow()

		r.Neg = p.parseBlockStmt(2, 0)

		p.newRow(0, 0)
	}

	p.popRow()
	r.Box = p.popBox()

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
