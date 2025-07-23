package ast

import (
	"fmt"
	"io"
	"os"

	"github.com/paraskun/x/env"
	"github.com/paraskun/x/lex"
	"github.com/paraskun/x/typ"
)

type Visitor interface {
	VisitStmt(Stmt)
	VisitDecl(Decl)
	VisitExpr(Expr)
}

type Node interface {
	Accept(Visitor)
}

// Statements

type (
	Stmt interface {
		Node

		stmt()
	}

	BlockStmt struct {
		Tok  *lex.Token // open '{'
		Body []Stmt

		// Typing pass

		Env *env.Env
	}

	VarStmt struct {
		Tok *lex.Token // 'var'
		Idf *lex.Token
		Ini Expr
	}

	LetStmt struct {
		Tok *lex.Token // 'let'
		Idf *lex.Token
		Ini Expr
	}

	AssignStmt struct {
		Tok *lex.Token // '='
		Var Expr
		Val Expr
	}

	LoopStmt struct {
		Tok *lex.Token // 'for'
		Con Expr       // loop condition
		Rep *BlockStmt // loop body
	}

	CondStmt struct {
		Tok *lex.Token // 'if'
		Con Expr
		Pos *BlockStmt
		Neg *BlockStmt
	}

	CallStmt struct {
		*CallExpr
	}

	ReturnStmt struct {
		Tok *lex.Token // 'return'
		Ret []Expr
	}
)

// Expressions

type (
	Expr interface {
		Node

		Type() *typ.Type
		Caps() uint8
		expr()
	}

	// Immediate expression of basic or enum type.
	ImmExpr struct {
		Tok *lex.Token

		// Typing pass

		Obj *env.Object
	}

	StructField struct {
		Tok *lex.Token // field name
		Val Expr
	}

	// Immediate expression of struct type.
	StructExpr struct {
		Tok    *lex.Token // struct name
		Fields []StructField

		// Typing pass

		Typ *typ.Type
	}

	IdfExpr struct {
		Tok *lex.Token

		// Typing pass

		Obj *env.Object
	}

	DotExpr struct {
		Tok *lex.Token // '.'
		Env Expr
		Mem *lex.Token

		// Typing pass

		Typ *typ.Type
	}

	// Infix expression.
	InfExpr struct {
		Tok *lex.Token
		X   Expr
		Y   Expr

		// Typing pass

		Typ *typ.Type
	}

	// Prefix expression.
	PfxExpr struct {
		Tok *lex.Token
		X   Expr

		// Typing pass

		Typ *typ.Type
	}

	CallExpr struct {
		Tok *lex.Token // function to call
		Arg []Expr

		// Typing pass

		FunTyp *typ.Type // function type
		RetTyp *typ.Type // return type
	}

	ToSigExpr struct {
		Tok *lex.Token
		X   Expr

		// Typing pass

		Typ *typ.Type // i64
	}

	ToUnsExpr struct {
		Tok *lex.Token
		X   Expr

		// Typing pass

		Typ *typ.Type // u64
	}

	ToFltExpr struct {
		Tok *lex.Token
		X   Expr

		// Typing pass

		Typ *typ.Type // f64
	}
)

// Declarations

type (
	Decl interface {
		Node

		decl()
	}

	UseDecl struct {
		Tok *lex.Token // 'use'
		Idf *lex.Token
		Pkg *env.Package

		// Typing pass

		Obj *env.Object
	}

	VarDecl struct {
		*VarStmt
	}

	// TypeSpec is a node for basic, struct
	// or enum specification.
	TypeSpec struct {
		Tok *lex.Token

		// Typing pass

		Typ *typ.Type
	}

	Field struct {
		Tok *lex.Token // name, maybe nil
		Typ *TypeSpec
	}

	FuncDecl struct {
		Tok *lex.Token // 'func'
		Idf *lex.Token
		Arg []*Field
		Ret []*Field

		Body *BlockStmt

		// Typing pass

		Env *env.Env
		Obj *env.Object
	}

	StructDecl struct {
		Tok *lex.Token // 'type'
		Idf *lex.Token // name
		Mem []*Field   // members

		// Typing pass

		Obj *env.Object
	}

	EnumDecl struct {
		Tok *lex.Token   // 'enum'
		Idf *lex.Token   // name
		Mem []*lex.Token // members

		// Typing pass

		Obj *env.Object
	}
)

type File struct {
	Dec []Decl
}

type options struct {
	withType bool
	withFunc bool
	withGlob bool
	withDeps bool
}

type Option func(*options)

func WithType(opt *options) {
	opt.withType = true
}

func WithGlob(opt *options) {
	opt.withGlob = true
}

func WithFunc(opt *options) {
	opt.withFunc = true
}

// WithDeps forces parser to recursively parse
// packages in use by the target.
func WithDeps(opt *options) {
	opt.withDeps = true
}

type parser struct {
	lex lex.Scanner
	buf bool
	prv *lex.Token
	cur *lex.Token

	ops *options
	pkg *env.Package
}

func (p *parser) next() *lex.Token {
	if !p.buf {
		p.prv = p.cur
		p.cur = p.lex.Next()
	}

	p.buf = false

	return p.cur
}

func (p *parser) peek() *lex.Token {
	if !p.buf {
		p.buf = true
		p.prv = p.cur
		p.cur = p.lex.Next()
	}

	return p.cur
}

func (p *parser) must(tt lex.TokenType) *lex.Token {
	if p.next().TokenType != tt {
		panic(fmt.Errorf("%v given, but %v expected", p.cur.TokenType, tt))
	}

	return p.cur
}

func Parse(pkg *env.Package, opts ...Option) {
	ops := &options{}

	for _, opt := range opts {
		opt(ops)
	}

	parse(pkg, ops)
}

func parse(pkg *env.Package, ops *options) {
	p := &parser{pkg: pkg, ops: ops}

	for _, src := range pkg.XSrc {
		fil, err := os.Open(src.Path)

		if err != nil {
			panic(err)
		} else {
			defer fil.Close()
		}

		dec := make([]Decl, 0)
		buf, err := io.ReadAll(fil)

		if err != nil {
			panic(err)
		}

		p.lex.Load([]rune(string(buf)))

		for p.peek().TokenType != lex.EOF {
			if d := p.parseDecl(); dec != nil {
				dec = append(dec, d)
			}
		}

		src.Dec = &File{dec}
	}
}

func (p *parser) parseDecl() Decl {
	switch p.peek().TokenType {
	case lex.USE:
		u := p.parseUseDecl()

		if p.ops.withDeps {
			pkg := p.pkg.Mod.Lookup(u.Idf.Lit)

			if pkg == nil {
				panic("no such package in context")
			}

			u.Pkg = pkg
			parse(pkg, p.ops)

			return u
		}
	case lex.VAR:
		v := &VarDecl{VarStmt: p.parseVarStmt()}

		if p.ops.withGlob {
			return v
		}
	case lex.FUNC:
		f := p.parseFuncDecl()

		if p.ops.withFunc {
			return f
		}
	case lex.TYPE:
		t := p.parseStructDecl()

		if p.ops.withType {
			return t
		}
	case lex.ENUM:
		e := p.parseEnumDecl()

		if p.ops.withType {
			return e
		}
	default:
		panic("declaration expected")
	}

	return nil
}

func (p *parser) parseUseDecl() *UseDecl {
	return &UseDecl{
		Tok: p.must(lex.USE),
		Idf: p.must(lex.ISTR),
	}
}

func (p *parser) parseFuncDecl() *FuncDecl {
	f := &FuncDecl{
		Tok: p.must(lex.FUNC),
		Idf: p.must(lex.IDF),
	}

	p.must(lex.LP)

	for p.peek().TokenType != lex.RP {
		f.Arg = append(f.Arg, p.parseNamedField())

		if p.peek().TokenType != lex.RP {
			p.must(lex.COM)
		}
	}

	p.next()

	if p.peek().TokenType == lex.LP {
		f.Ret = append(f.Ret, p.parseUnnamedField())

		if p.peek().TokenType != lex.RP {
			p.must(lex.COM)
		}
	}

	if len(f.Ret) > 1 {
		// TODO: multiple return values
		panic("multiple return values are not supported yet")
	}

	p.next()

	if p.peek().TokenType == lex.LB {
		f.Body = p.parseBlockStmt()
	}

	return f
}

func (p *parser) parseStructDecl() *StructDecl {
	c := &StructDecl{
		Tok: p.must(lex.TYPE),
		Idf: p.must(lex.IDF),
	}

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

func (p *parser) parseNamedField() *Field {
	return &Field{
		Tok: p.must(lex.IDF),
		Typ: p.parseTypeSpec(),
	}
}

func (p *parser) parseUnnamedField() *Field {
	return &Field{
		Typ: p.parseTypeSpec(),
	}
}

func (p *parser) parseTypeSpec() *TypeSpec {
	switch p.peek().TokenType {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDF:
		return &TypeSpec{Tok: p.next()}
	}

	panic("type specification expected")
}

func (p *parser) parseBlockStmt() *BlockStmt {
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
					Env: cur,
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
		r.Ret = append(r.Ret, p.parseExpr0())

		if p.peek().TokenType != lex.SEM {
			p.must(lex.COM)
		}
	}

	p.must(lex.SEM)

	if len(r.Ret) > 1 {
		// TODO: multiple return values
		panic("multiple return values are not supported yet")
	}

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
				Env: cur,
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

				exp.Fields = append(exp.Fields, StructField{
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
		exp := &ToSigExpr{Tok: p.next()}

		p.must(lex.LP)
		exp.X = p.parseExpr0()
		p.must(lex.RP)

		return exp
	case lex.U64:
		exp := &ToUnsExpr{Tok: p.next()}

		p.must(lex.LP)
		exp.X = p.parseExpr0()
		p.must(lex.RP)

		return exp
	case lex.F64:
		exp := &ToFltExpr{Tok: p.next()}

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

func (e *ImmExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *StructExpr) Accept(v Visitor) { v.VisitExpr(e) }
func (e *IdfExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *DotExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *InfExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *PfxExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *CallExpr) Accept(v Visitor)   { v.VisitExpr(e) }
func (e *ToSigExpr) Accept(v Visitor)  { v.VisitExpr(e) }
func (e *ToUnsExpr) Accept(v Visitor)  { v.VisitExpr(e) }
func (e *ToFltExpr) Accept(v Visitor)  { v.VisitExpr(e) }

func (e *ImmExpr) Type() *typ.Type    { return e.Obj.Typ }
func (e *StructExpr) Type() *typ.Type { return e.Typ }
func (e *IdfExpr) Type() *typ.Type    { return e.Obj.Typ }
func (e *DotExpr) Type() *typ.Type    { return e.Typ }
func (e *InfExpr) Type() *typ.Type    { return e.Typ }
func (e *PfxExpr) Type() *typ.Type    { return e.Typ }
func (e *CallExpr) Type() *typ.Type   { return e.RetTyp }
func (e *ToSigExpr) Type() *typ.Type  { return e.Typ }
func (e *ToUnsExpr) Type() *typ.Type  { return e.Typ }
func (e *ToFltExpr) Type() *typ.Type  { return e.Typ }

func (e *ImmExpr) Caps() uint8    { return e.Obj.Cap }
func (e *StructExpr) Caps() uint8 { return 0 }
func (e *IdfExpr) Caps() uint8    { return e.Obj.Cap }
func (e *DotExpr) Caps() uint8    { return e.Env.Caps() }
func (e *InfExpr) Caps() uint8    { return 0 }
func (e *PfxExpr) Caps() uint8    { return 0 }
func (e *CallExpr) Caps() uint8   { return 0 }
func (e *ToSigExpr) Caps() uint8  { return 0 }
func (e *ToUnsExpr) Caps() uint8  { return 0 }
func (e *ToFltExpr) Caps() uint8  { return 0 }

func (*ImmExpr) expr()    {}
func (*StructExpr) expr() {}
func (*IdfExpr) expr()    {}
func (*DotExpr) expr()    {}
func (*InfExpr) expr()    {}
func (*PfxExpr) expr()    {}
func (*CallExpr) expr()   {}
func (*ToSigExpr) expr()  {}
func (*ToUnsExpr) expr()  {}
func (*ToFltExpr) expr()  {}

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
