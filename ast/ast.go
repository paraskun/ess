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
		Box tty.Container
		Sub []Stmt

		Env *typ.Env
	}

	VarStmt struct {
		Box tty.Container
		Var *tty.Tok
		Ini Expr

		Obj *typ.Object
	}

	LetStmt struct {
		Box tty.Container
		Var *tty.Tok
		Ini Expr

		Obj *typ.Object
	}

	AssignStmt struct {
		Box tty.Container
		Var Expr
		Val Expr

		Obj *typ.Object
	}

	LoopStmt struct {
		Box tty.Container
		Con Expr
		Sub *BlockStmt
	}

	CondStmt struct {
		Box tty.Container
		Con Expr
		Pos *BlockStmt
		Neg *BlockStmt
	}

	CallStmt struct {
		*CallExpr
	}

	ReturnStmt struct {
		Box tty.Container
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
		Box tty.Container
		Val Expr
	}

	StructExpr struct {
		Box tty.Container
		Mem []StructFieldExpr

		Obj *typ.Object
	}

	IdenExpr struct {
		Tok *tty.Tok

		Obj *typ.Object
	}

	DotExpr struct {
		Box tty.Container
		Mem *tty.Tok
		Ctx Expr

		Obj *typ.Object
	}

	InfExpr struct {
		Box tty.Container
		X   Expr
		Y   Expr

		Res *typ.Object
	}

	PfxExpr struct {
		Box tty.Container
		X   Expr

		Res *typ.Object
	}

	CallExpr struct {
		Box tty.Container
		Sym *tty.Tok
		Arg []Expr

		Fun *typ.Object
		Ret *typ.Object
	}

	ToI64Expr struct {
		Box tty.Container
		X   Expr

		Res *typ.Object
	}

	ToU64Expr struct {
		Box tty.Container
		X   Expr

		Res *typ.Object
	}

	ToF64Expr struct {
		Box tty.Container
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
		Box tty.Container
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
		Box tty.Container
		Sym *tty.Tok
		Typ *TypeSpec
	}

	FuncDecl struct {
		Box tty.Container
		Sig *tty.Row
		Sym *tty.Tok
		Arg []*NamedField
		Ret *TypeSpec
		Sub *BlockStmt

		Env *typ.Env
		Obj *typ.Object
	}

	StructDecl struct {
		Box tty.Container
		Sym *tty.Tok
		Mem []*NamedField

		Obj *typ.Object
	}

	EnumDecl struct {
		Box tty.Container
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
	que []tty.Container
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

func (p *parser) next() *lex.Lexeme {
	if !p.buf {
		p.prv = p.cur
		p.cur = p.scan()
	}

	p.buf = false

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

func (p *parser) must(t lex.Type) *lex.Lexeme {
	if p.peek().Typ != t {
		// TODO: process error

		panic("unexpected token")
	}

	return p.next()
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
		f, err := os.Open(src.Path)

		if err != nil {
			panic(err)
		} else {
			defer f.Close()
		}

		buf, err := io.ReadAll(f)

		if err != nil {
			panic(err)
		}

		dec := &File{
			Box: &tty.Box{},
			Dec: make([]Decl, 0),
		}

		p.src = src
		p.lex.Load([]rune(string(buf)))

		for p.peek().Typ != lex.EOF {
			sub := p.parseDecl()
			ind := 1

			if len(dec.Dec) == 0 {
				ind = 0
			}

			dec.Dec = append(dec.Dec, p.parseDecl())
			dec.Box.Add(sub.Span(), 0, ind)
		}

		src.Dec = dec
	}
}

func (p *parser) parseDecl() Decl {
	switch p.peek().Typ {
	case lex.USE:
		use := p.parseUseDecl()
		pkg := p.pkg.Mod.Lookup(use.Pkg.Lit)

		if pkg == nil {
			use.Box.Hint().Text = "unknown package"
			use.Box.Hint().Attr = tty.Attr{
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
	dec := &UseDecl{Box: &tty.Row{}}

	dec.Box.Add(p.must(lex.USE).Tok, 0, 0)
	dec.Pkg = p.must(lex.ISTR).Tok
	dec.Box.Add(dec.Pkg, 1, 0)

	return dec
}

func (p *parser) parseFuncDecl() *FuncDecl {
	dec := &FuncDecl{Sig: &tty.Row{}}
	top := &tty.Row{}

	top.Add(p.must(lex.FUNC).Tok, 0, 0)
	dec.Sym = p.must(lex.IDEN).Tok
	dec.Sig.Add(dec.Sym, 1, 0)
	dec.Sig.Add(p.must(lex.LP).Tok, 0, 0)

	for p.peek().Typ != lex.RP {
		ind := 0

		if len(dec.Arg) == 0 {
			ind = 1
		}

		arg := p.parseNamedField()

		dec.Arg = append(dec.Arg, arg)
		dec.Sig.Add(arg.Box, ind, 0)

		if p.peek().Typ != lex.RP {
			dec.Sig.Add(p.must(lex.COM).Tok, 0, 0)
		}
	}

	dec.Sig.Add(p.must(lex.RP).Tok, 0, 0)

	if p.peek().Typ == lex.LP {
		dec.Sig.Add(p.next().Tok, 1, 0)

		dec.Ret = p.parseTypeSpec()

		dec.Sig.Add(dec.Ret.Tok, 0, 0)
		dec.Sig.Add(p.must(lex.RP).Tok, 0, 0)
	}

	top.Add(dec.Sig, 1, 0)

	if p.peek().Typ == lex.LB {
		dec.Box = &tty.Box{}

		top.Add(p.next().Tok, 1, 0)

		dec.Sub = p.parseBlockStmt()

		dec.Box.Add(top, 0, 0)
		dec.Box.Add(dec.Sub.Box, 4, 0)
		dec.Box.Add(p.must(lex.RB).Tok, 0, 0)
	} else {
		dec.Box = &tty.Row{}
		dec.Box.Add(top, 0, 0)
	}

	return dec
}

func (p *parser) parseStructDecl() *StructDecl {
	dec := &StructDecl{}
	top := &tty.Row{}

	top.Add(p.must(lex.TYPE).Tok, 0, 0)
	dec.Sym = p.must(lex.IDEN).Tok
	top.Add(dec.Sym, 1, 0)
	top.Add(p.must(lex.LB).Tok, 1, 0)

	if p.peek().Typ != lex.RB {
		dec.Box = &tty.Box{}
		dec.Box.Add(top, 0, 0)

		for p.peek().Typ != lex.RB {
			mem := p.parseNamedField()

			dec.Mem = append(dec.Mem, mem)
			dec.Box.Add(mem.Box, 4, 0)
		}
	} else {
		dec.Box = &tty.Row{}
		dec.Box.Add(top, 0, 0)
	}

	dec.Box.Add(p.must(lex.RB).Tok, 0, 0)

	return dec
}

func (p *parser) parseEnumDecl() *EnumDecl {
	dec := &EnumDecl{}
	top := &tty.Row{}

	top.Add(p.must(lex.ENUM).Tok, 0, 0)
	dec.Sym = p.must(lex.IDEN).Tok
	top.Add(dec.Sym, 1, 0)
	top.Add(p.must(lex.LB).Tok, 1, 0)

	if p.peek().Typ != lex.RB {
		dec.Box = &tty.Box{}
		dec.Box.Add(top, 0, 0)

		for p.peek().Typ != lex.RB {
			mem := p.must(lex.IDEN).Tok

			dec.Mem = append(dec.Mem, mem)
			dec.Box.Add(mem, 4, 0)
		}
	} else {
		dec.Box = &tty.Row{}
		dec.Box.Add(top, 0, 0)
	}

	dec.Box.Add(p.must(lex.RB).Tok, 0, 0)

	return dec
}

func (p *parser) parseNamedField() *NamedField {
	res := &NamedField{
		Box: &tty.Row{},
		Sym: p.must(lex.IDEN).Tok,
		Typ: p.parseTypeSpec(),
	}

	res.Box.Add(res.Sym, 0, 0)
	res.Box.Add(res.Typ.Tok, 1, 0)

	return res
}

func (p *parser) parseTypeSpec() *TypeSpec {
	switch p.peek().Typ {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDEN:
		return &TypeSpec{Tok: p.next().Tok}
	}

	// TODO: error handling
	panic("type specification expected")
}

func (p *parser) parseBlockStmt() *BlockStmt {
	res := &BlockStmt{}

	if p.peek().Typ != lex.RB {
		res.Box = &tty.Box{}

		for p.peek().Typ != lex.RB {
			sub := p.parseStmt()

			res.Sub = append(res.Sub, sub)
			res.Box.Add(sub.Span(), 4, 0)
		}
	} else {
		res.Box = &tty.Row{}
	}

	return res
}

func (p *parser) parseStmt() Stmt {
	switch p.peek().Typ {
	case lex.IDEN:
		iden := p.next()

		if p.peek().Typ == lex.LP {
			exe := p.parseCall(iden.Tok)

			switch b := exe.Box.(type) {
			case *tty.Box:
				b.AddLast(p.must(lex.SEM).Tok, 0)
			case *tty.Row:
				b.Add(p.must(lex.SEM).Tok, 0, 0)
			}

			return &CallStmt{exe}
		}

		var cur Expr = &IdenExpr{Tok: iden.Tok}

		for {
			switch p.peek().Typ {
			case lex.DOT:
				dot := &DotExpr{
					Box: &tty.Row{},
					Ctx: cur,
				}

				dot.Box.Add(dot.Ctx.Span(), 0, 0)
				dot.Box.Add(p.next().Tok, 0, 0)
				dot.Mem = p.must(lex.IDEN).Tok
				dot.Box.Add(dot.Mem, 0, 0)

				cur = dot

				continue
			}

			break
		}

		res := &AssignStmt{Var: cur}
		top := &tty.Row{}

		top.Add(cur.Span(), 0, 0)
		top.Add(p.must(lex.EQ).Tok, 1, 0)

		res.Val = p.parseExpr0()

		switch b := res.Val.Span().(type) {
		case *tty.Box:
			res.Box = &tty.Box{}

			b.AddLast(p.must(lex.SEM).Tok, 0)

			res.Box.Add(top, 0, 0)
			res.Box.Add(b, 4, 0)
		case tty.Mono:
			res.Box = &tty.Row{}

			res.Box.Add(top, 0, 0)
			res.Box.Add(b, 1, 0)
			res.Box.Add(p.must(lex.SEM).Tok, 0, 0)
		}

		return res
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

func (p *parser) parseCall(sym *tty.Tok) *CallExpr {
	res := &CallExpr{}
	top := &tty.Row{}

	top.Add(sym, 0, 0)
	top.Add(p.next().Tok, 0, 0)

	box := &tty.Box{}
	row := top

	for p.peek().Typ != lex.RP {
		arg := p.parseExpr0()

		switch b := arg.Span().(type) {
		case *tty.Box:
			if len(row.Sub) != 0 {
				box.Add(row, 4, 0)
			}

			box.Add(b, 4, 0)

			if p.peek().Typ != lex.RP {
				box.AddLast(p.must(lex.COM).Tok, 0)
				row = &tty.Row{}
			} else {
				box.AddLast(p.must(lex.RP).Tok, 0)
			}
		case tty.Mono:
			if len(row.Sub) != 0 {
				row.Add(b, 1, 0)
			} else {
				row.Add(b, 0, 0)
			}

			if p.peek().Typ != lex.RP {
				row.Add(p.must(lex.COM).Tok, 0, 0)
			} else {
				row.Add(p.must(lex.RP).Tok, 0, 0)
			}
		}

		res.Arg = append(res.Arg, arg)
	}

	if len(box.Sub) != 0 {
		res.Box = box
	} else {
		res.Box = row
	}

	return res
}

func (p *parser) parseVarStmt() *VarStmt {
	top := &tty.Row{}
	top.Add(p.must(lex.VAR).Tok, 0, 0)
	res := &VarStmt{Var: p.must(lex.IDEN).Tok}
	top.Add(res.Var, 1, 0)
	top.Add(p.must(lex.EQ).Tok, 1, 0)

	res.Ini = p.parseExpr0()

	switch b := res.Ini.Span().(type) {
	case *tty.Box:
		res.Box = &tty.Box{}

		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 4, 0)

		b.AddLast(p.must(lex.SEM).Tok, 0)
	case tty.Mono:
		res.Box = &tty.Row{}

		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 1, 0)
		res.Box.Add(p.must(lex.SEM).Tok, 0, 0)
	}

	return res
}

func (p *parser) parseLetStmt(rowInd, boxInd int) *LetStmt {
	top := &tty.Row{}
	top.Add(p.must(lex.LET).Tok, 0, 0)
	res := &LetStmt{Var: p.must(lex.IDEN).Tok}
	top.Add(res.Var, 1, 0)
	top.Add(p.must(lex.EQ).Tok, 1, 0)

	res.Ini = p.parseExpr0()

	switch b := res.Ini.Span().(type) {
	case *tty.Box:
		res.Box = &tty.Box{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 4, 0)
		b.AddLast(p.must(lex.SEM).Tok, 0)
	case tty.Mono:
		res.Box = &tty.Row{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 1, 0)
		res.Box.Add(p.must(lex.SEM).Tok, 0, 0)
	}

	return res
}

func (p *parser) parseLoopStmt() *LoopStmt {
	tok := p.must(lex.FOR).Tok
	res := &LoopStmt{
		Box: &tty.Box{},
		Con: p.parseExpr0(),
	}

	var log tty.Container

	switch res.Con.Span().(type) {
	case *tty.Box:
		log = &tty.Box{}
		log.Add(tok, 0, 0)
		log.Add(res.Con.Span(), 4, 0)
		log.Add(p.must(lex.LB).Tok, 0, 0)
	case tty.Mono:
		log = &tty.Row{}
		log.Add(tok, 0, 0)
		log.Add(res.Con.Span(), 1, 0)
		log.Add(p.must(lex.LB).Tok, 1, 0)
	}

	res.Box.Add(log, 0, 0)
	res.Sub = p.parseBlockStmt()
	res.Box.Add(res.Sub.Box, 4, 0)
	res.Box.Add(p.must(lex.RB).Tok, 0, 0)

	return res
}

func (p *parser) parseCondStmt(rowInd, boxInd int) *CondStmt {
	tok := p.must(lex.IF).Tok
	box := &tty.Box{}
	res := &CondStmt{
		Box: box,
		Con: p.parseExpr0(),
	}

	var log tty.Container

	switch res.Con.Span().(type) {
	case *tty.Box:
		log = &tty.Box{}
		log.Add(tok, 0, 0)
		log.Add(res.Con.Span(), 4, 0)
		log.Add(p.must(lex.LB).Tok, 0, 0)
	case tty.Mono:
		log = &tty.Row{}
		log.Add(tok, 0, 0)
		log.Add(res.Con.Span(), 1, 0)
		log.Add(p.must(lex.LB).Tok, 1, 0)
	}

	box.Add(log, 0, 0)
	res.Pos = p.parseBlockStmt()
	box.Add(res.Pos.Box, 4, 0)
	box.Add(p.must(lex.RB).Tok, 0, 0)

	if p.peek().Typ == lex.ELSE {
		box.AddLast(p.next().Tok, 1)
		box.AddLast(p.must(lex.LB).Tok, 1)
		res.Neg = p.parseBlockStmt()
		box.Add(res.Neg.Box, 4, 0)
		box.Add(p.must(lex.RB).Tok, 0, 0)
	}

	return res
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
