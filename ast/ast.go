package ast

import (
	"fmt"

	"github.com/paraskun/ess-go/lex"
	"github.com/paraskun/ess-go/typ"
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
		Tok *lex.Token
		Env *typ.Env

		Body []Stmt
	}

	AssignStmt struct {
		Tok *lex.Token
		Var Expr
		Val Expr
		Dec bool
		Ini bool
	}

	LoopStmt struct {
		Tok *lex.Token
		Con Expr
		Rep *BlockStmt
	}

	CondStmt struct {
		Tok *lex.Token
		Con Expr
		Pos *BlockStmt
		Neg *BlockStmt
	}

	CallStmt struct {
		Exp *CallExpr
	}

	ReturnStmt struct {
		Tok *lex.Token
		Ret Expr
	}
)

// Expressions

type (
	Expr interface {
		Node

		Type() *typ.Type
		expr()
	}

	BaseImmExpr struct {
		Tok *lex.Token
		Obj *typ.Object
	}

	CompField struct {
		Tok *lex.Token
		Val Expr
	}

	CompImmExpr struct {
		Tok *lex.Token
		Typ *typ.Type

		Fields []CompField
	}

	IdfExpr struct {
		Tok *lex.Token
		Obj *typ.Object
	}

	DotExpr struct {
		Tok *lex.Token
		Typ *typ.Type

		Comp  Expr
		Field *lex.Token
	}

	InfExpr struct {
		Tok *lex.Token
		Typ *typ.Type

		X Expr
		Y Expr
	}

	PfxExpr struct {
		Tok *lex.Token
		Typ *typ.Type

		X Expr
	}

	CallExpr struct {
		Tok *lex.Token
		Typ *typ.Type
		Sym *typ.Type
		Arg []Expr
	}

	ToSigExpr struct {
		Tok *lex.Token
		Typ *typ.Type

		X Expr
	}

	ToUnsExpr struct {
		Tok *lex.Token
		Typ *typ.Type

		X Expr
	}

	ToFltExpr struct {
		Tok *lex.Token
		Typ *typ.Type

		X Expr
	}
)

// Type specification

type (
	TypeSpec interface {
		Type() *typ.Type
		spec()
	}

	BaseSpec struct {
		Tok *lex.Token
		Typ *typ.Type
	}

	FieldSpec struct {
		Tok *lex.Token  // maybe nil
		Obj *typ.Object // maybe nil
		Typ TypeSpec
	}

	FuncSpec struct {
		BaseSpec

		Arg []*FieldSpec
		Ret *FieldSpec
	}

	CompSpec struct {
		BaseSpec

		Fields []*FieldSpec
	}
)

// Declarations

type (
	Decl interface {
		Stmt

		decl()
	}

	FuncDecl struct {
		Tok *lex.Token
		Env *typ.Env
		Sym *typ.Type

		Spec *FuncSpec
		Body *BlockStmt
	}

	CompDecl struct {
		Tok *lex.Token

		Spec *CompSpec
	}
)

type Pragma struct {
	Env *typ.Env
	Dec []Decl
}

type parser struct {
	lex lex.Scanner
	buf bool
	prv *lex.Token
	cur *lex.Token
}

func (p *parser) error(err error) {
	panic(fmt.Errorf("parser [%d:%d]: %w", p.cur.Row, p.cur.Col, err))
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

func (p *parser) expect(tt lex.TokenType) *lex.Token {
	if p.next().TokenType != tt {
		p.error(fmt.Errorf("%v given, but %v expected", p.cur.TokenType, tt))
	}

	return p.cur
}

func Parse(buf []rune) *Pragma {
	p := parser{}
	r := Pragma{}

	p.lex.Load(buf)

	for p.peek().TokenType != lex.EOF {
		r.Dec = append(r.Dec, p.parseDecl())
	}

	return &r
}

func (p *parser) parseDecl() Decl {
	switch p.peek().TokenType {
	case lex.FUNC:
		return p.parseFuncDecl()
	case lex.TYPE:
		return p.parseCompDecl()
	}

	p.error(fmt.Errorf("declaration expected"))

	return nil
}

func (p *parser) parseFuncDecl() *FuncDecl {
	f := &FuncDecl{}

	p.expect(lex.FUNC)

	f.Tok = p.expect(lex.IDF)
	f.Spec = &FuncSpec{
		BaseSpec: BaseSpec{
			Tok: p.expect(lex.LP),
		},
	}

	for p.peek().TokenType != lex.RP {
		f.Spec.Arg = append(f.Spec.Arg, p.parseNamedSpec())

		if p.peek().TokenType != lex.RP {
			p.expect(lex.COM)
		}
	}

	p.expect(lex.RP)

	if p.peek().TokenType != lex.LB {
		f.Spec.Ret = p.parseUnnamedSpec()
	}

	f.Body = p.parseBlockStmt()

	return f
}

func (p *parser) parseCompDecl() *CompDecl {
	c := &CompDecl{}

	p.expect(lex.TYPE)

	c.Tok = p.expect(lex.IDF)
	c.Spec = &CompSpec{
		BaseSpec: BaseSpec{
			Tok: p.expect(lex.LB),
		},
	}

	for p.peek().TokenType != lex.RB {
		c.Spec.Fields = append(c.Spec.Fields, p.parseNamedSpec())
	}

	p.expect(lex.RB)

	return c
}

func (p *parser) parseNamedSpec() *FieldSpec {
	return &FieldSpec{
		Tok: p.expect(lex.IDF),
		Typ: p.parseTypeSpec(),
	}
}

func (p *parser) parseUnnamedSpec() *FieldSpec {
	return &FieldSpec{
		Typ: p.parseTypeSpec(),
	}
}

func (p *parser) parseTypeSpec() TypeSpec {
	switch p.peek().TokenType {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDF:
		return &BaseSpec{Tok: p.next()}
	}

	panic("type specification expected")
}

func (p *parser) parseBlockStmt() *BlockStmt {
	r := BlockStmt{
		Tok: p.expect(lex.LB),
	}

	for p.peek().TokenType != lex.RB {
		r.Body = append(r.Body, p.parseStmt())
	}

	p.expect(lex.RB)

	return &r
}

func (p *parser) parseStmt() Stmt {
	switch p.peek().TokenType {
	case lex.FUNC:
		return p.parseFuncDecl()
	case lex.TYPE:
		return p.parseCompDecl()
	case lex.IDF:
		idf := p.next()

		if p.peek().TokenType == lex.LP {
			exe := p.parseCall(idf)
			p.expect(lex.SEM)

			return &CallStmt{exe}
		}

		var cur Expr = &IdfExpr{Tok: idf}

		for {
			switch p.peek().TokenType {
			case lex.DOT:
				cur = &DotExpr{
					Tok:   p.next(),
					Comp:  cur,
					Field: p.expect(lex.IDF),
				}

				continue
			}

			break
		}

		r := &AssignStmt{
			Dec: false,
			Var: cur,
			Tok: p.expect(lex.EQ),
			Val: p.parseExpr0(),
		}

		p.expect(lex.SEM)

		return r
	case lex.VAR:
		return p.parseAssignStmt()
	case lex.FOR:
		return p.parseLoopStmt()
	case lex.IF:
		return p.parseCondStmt()
	case lex.RET:
		return p.parseReturnStmt()
	}

	p.error(fmt.Errorf("statement expected"))

	return nil
}

func (p *parser) parseCall(tok *lex.Token) *CallExpr {
	p.next() // (

	exe := &CallExpr{Tok: tok}

	for p.peek().TokenType != lex.RP {
		exe.Arg = append(exe.Arg, p.parseExpr0())

		if p.peek().TokenType != lex.RP {
			p.expect(lex.COM)
		}
	}

	p.expect(lex.RP)

	return exe
}

func (p *parser) parseAssignStmt() *AssignStmt {
	p.next() // var

	r := AssignStmt{
		Dec: true,
		Var: &IdfExpr{Tok: p.expect(lex.IDF)},
		Tok: p.next(),
		Val: p.parseExpr0(),
	}

	if r.Tok.TokenType != lex.EQ && r.Tok.TokenType != lex.INI {
		p.error(fmt.Errorf("%v given, but %v expected", r.Tok.TokenType, lex.EQ))
	}

	if r.Tok.TokenType == lex.INI {
		r.Ini = true
	}

	p.expect(lex.SEM)

	return &r
}

func (p *parser) parseLoopStmt() *LoopStmt {
	r := LoopStmt{
		Tok: p.expect(lex.FOR),
	}

	r.Con = p.parseExpr0()
	r.Rep = p.parseBlockStmt()

	return &r
}

func (p *parser) parseCondStmt() *CondStmt {
	r := CondStmt{
		Tok: p.expect(lex.IF),
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
	r := ReturnStmt{
		Tok: p.expect(lex.RET),
		Ret: p.parseExpr0(),
	}

	p.expect(lex.SEM)

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
				Tok:   p.next(),
				Comp:  cur,
				Field: p.expect(lex.IDF),
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

			exp := &CompImmExpr{Tok: idf}

			for p.peek().TokenType != lex.RB {
				tok := p.expect(lex.IDF)
				p.expect(lex.COL)
				val := p.parseExpr0()

				exp.Fields = append(exp.Fields, CompField{
					Tok: tok,
					Val: val,
				})

				if p.peek().TokenType != lex.RB {
					p.expect(lex.COM)
				}
			}

			p.expect(lex.RB)

			return exp
		}

		return &IdfExpr{Tok: idf}
	case lex.II64, lex.IU64, lex.IF64, lex.TRUE, lex.FALSE:
		return &BaseImmExpr{Tok: p.next()}
	case lex.I64:
		exp := &ToSigExpr{Tok: p.next()}

		p.expect(lex.LP)
		exp.X = p.parseExpr0()
		p.expect(lex.RP)

		return exp
	case lex.U64:
		exp := &ToUnsExpr{Tok: p.next()}

		p.expect(lex.LP)
		exp.X = p.parseExpr0()
		p.expect(lex.RP)

		return exp
	case lex.F64:
		exp := &ToFltExpr{Tok: p.next()}

		p.expect(lex.LP)
		exp.X = p.parseExpr0()
		p.expect(lex.RP)

		return exp
	case lex.LP:
		p.next()
		r := p.parseExpr0()
		p.expect(lex.RP)

		return r
	}

	p.error(fmt.Errorf("expression expected"))

	return nil
}

func (s *ReturnStmt) Accept(v Visitor) { v.VisitStmt(s) }
func (s *AssignStmt) Accept(v Visitor) { v.VisitStmt(s) }
func (s *BlockStmt) Accept(v Visitor)  { v.VisitStmt(s) }
func (s *LoopStmt) Accept(v Visitor)   { v.VisitStmt(s) }
func (s *CondStmt) Accept(v Visitor)   { v.VisitStmt(s) }
func (s *CallStmt) Accept(v Visitor)   { v.VisitStmt(s) }

func (*ReturnStmt) stmt() {}
func (*AssignStmt) stmt() {}
func (*BlockStmt) stmt()  {}
func (*LoopStmt) stmt()   {}
func (*CondStmt) stmt()   {}
func (*CallStmt) stmt()   {}

func (e *BaseImmExpr) Accept(v Visitor) { v.VisitExpr(e) }
func (e *CompImmExpr) Accept(v Visitor) { v.VisitExpr(e) }
func (e *IdfExpr) Accept(v Visitor)     { v.VisitExpr(e) }
func (e *DotExpr) Accept(v Visitor)     { v.VisitExpr(e) }
func (e *InfExpr) Accept(v Visitor)     { v.VisitExpr(e) }
func (e *PfxExpr) Accept(v Visitor)     { v.VisitExpr(e) }
func (e *CallExpr) Accept(v Visitor)    { v.VisitExpr(e) }
func (e *ToSigExpr) Accept(v Visitor)   { v.VisitExpr(e) }
func (e *ToUnsExpr) Accept(v Visitor)   { v.VisitExpr(e) }
func (e *ToFltExpr) Accept(v Visitor)   { v.VisitExpr(e) }

func (e *BaseImmExpr) Type() *typ.Type { return e.Obj.Typ }
func (e *CompImmExpr) Type() *typ.Type { return e.Typ }
func (e *IdfExpr) Type() *typ.Type     { return e.Obj.Typ }
func (e *DotExpr) Type() *typ.Type     { return e.Typ }
func (e *InfExpr) Type() *typ.Type     { return e.Typ }
func (e *PfxExpr) Type() *typ.Type     { return e.Typ }
func (e *CallExpr) Type() *typ.Type    { return e.Typ }
func (e *ToSigExpr) Type() *typ.Type   { return e.Typ }
func (e *ToUnsExpr) Type() *typ.Type   { return e.Typ }
func (e *ToFltExpr) Type() *typ.Type   { return e.Typ }

func (*BaseImmExpr) expr() {}
func (*CompImmExpr) expr() {}
func (*IdfExpr) expr()     {}
func (*DotExpr) expr()     {}
func (*InfExpr) expr()     {}
func (*PfxExpr) expr()     {}
func (*CallExpr) expr()    {}
func (*ToSigExpr) expr()   {}
func (*ToUnsExpr) expr()   {}
func (*ToFltExpr) expr()   {}

func (s *BaseSpec) Type() *typ.Type { return s.Typ }
func (s *FuncSpec) Type() *typ.Type { return s.Typ }
func (s *CompSpec) Type() *typ.Type { return s.Typ }

func (*BaseSpec) spec() {}
func (*FuncSpec) spec() {}
func (*CompSpec) spec() {}

func (d *FuncDecl) Accept(v Visitor) { v.VisitDecl(d) }
func (d *CompDecl) Accept(v Visitor) { v.VisitDecl(d) }

func (*FuncDecl) stmt() {}
func (*FuncDecl) decl() {}
func (*CompDecl) stmt() {}
func (*CompDecl) decl() {}
