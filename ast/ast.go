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
		Env  *typ.Env
		Body []Stmt
	}

	AssignStmt struct {
		Tok *lex.Token
		Var []Expr
		Val []Expr
		New bool
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

	ReturnStmt struct {
		Tok *lex.Token
		Arg []Expr
	}
)

func (s *ReturnStmt) Accept(v Visitor) { v.VisitStmt(s) }
func (s *AssignStmt) Accept(v Visitor) { v.VisitStmt(s) }
func (s *BlockStmt) Accept(v Visitor)  { v.VisitStmt(s) }
func (s *LoopStmt) Accept(v Visitor)   { v.VisitStmt(s) }
func (s *CondStmt) Accept(v Visitor)   { v.VisitStmt(s) }

func (*ReturnStmt) stmt() {}
func (*AssignStmt) stmt() {}
func (*BlockStmt) stmt()  {}
func (*LoopStmt) stmt()   {}
func (*CondStmt) stmt()   {}

// Expressions

type (
	Expr interface {
		Node

		Type() []*typ.Type
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
		Typ []*typ.Type

		Exe Expr
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

func (e *BaseImmExpr) Type() []*typ.Type { return []*typ.Type{e.Obj.Typ} }
func (e *CompImmExpr) Type() []*typ.Type { return []*typ.Type{e.Typ} }
func (e *IdfExpr) Type() []*typ.Type     { return []*typ.Type{e.Obj.Typ} }
func (e *DotExpr) Type() []*typ.Type     { return []*typ.Type{e.Typ} }
func (e *InfExpr) Type() []*typ.Type     { return []*typ.Type{e.Typ} }
func (e *PfxExpr) Type() []*typ.Type     { return []*typ.Type{e.Typ} }
func (e *CallExpr) Type() []*typ.Type    { return e.Typ }
func (e *ToSigExpr) Type() []*typ.Type   { return []*typ.Type{e.Typ} }
func (e *ToUnsExpr) Type() []*typ.Type   { return []*typ.Type{e.Typ} }
func (e *ToFltExpr) Type() []*typ.Type   { return []*typ.Type{e.Typ} }

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

		Arg []FieldSpec
		Ret []FieldSpec
	}

	CompSpec struct {
		BaseSpec

		Fields []FieldSpec
	}
)

func (s *BaseSpec) Type() *typ.Type { return s.Typ }
func (s *FuncSpec) Type() *typ.Type { return s.Typ }
func (s *CompSpec) Type() *typ.Type { return s.Typ }

func (*BaseSpec) spec() {}
func (*FuncSpec) spec() {}
func (*CompSpec) spec() {}

// Declarations

type (
	Decl interface {
		Stmt

		decl()
	}

	FuncDecl struct {
		Tok *lex.Token

		Spec *FuncSpec
		Body *BlockStmt
	}

	CompDecl struct {
		Tok *lex.Token

		Spec *CompSpec
	}
)

func (d *FuncDecl) Accept(v Visitor) { v.VisitDecl(d) }
func (d *CompDecl) Accept(v Visitor) { v.VisitDecl(d) }

func (*FuncDecl) stmt() {}
func (*FuncDecl) decl() {}
func (*CompDecl) stmt() {}
func (*CompDecl) decl() {}

type Pragma struct {
	*BlockStmt
}

type parser struct {
	lex lex.Scanner
	buf bool
	prv *lex.Token
	cur *lex.Token
}

func (p *parser) error(err error) {
	panic(fmt.Errorf("parser: %d:%d: %w", p.cur.Row, p.cur.Col, err))
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
	r := Pragma{
		BlockStmt: &BlockStmt{
			Env: typ.NewEnv(nil),
		},
	}

	p.lex.Load(buf)

	for p.peek().TokenType != lex.EOF {
		r.Body = append(r.Body, p.parseStmt())
	}

	return &r
}

func (p *parser) parseStmt() Stmt {
	switch p.peek().TokenType {
	case lex.RET:
		return p.parseReturnStmt()
	case lex.VAR, lex.IDF:
		return p.parseAssignStmt()
	case lex.FOR:
		return p.parseLoopStmt()
	case lex.IF:
		return p.parseCondStmt()
	case lex.FUNC:
		return p.parseFuncDecl()
	case lex.TYPE:
		return p.parseCompDecl()
	}

	p.error(fmt.Errorf("statement expected"))

	return nil
}

func (p *parser) parseBlockStmt() *BlockStmt {
	r := BlockStmt{}

	for p.peek().TokenType != lex.RB {
		r.Body = append(r.Body, p.parseStmt())
	}

	return &r
}

func (p *parser) parseReturnStmt() *ReturnStmt {
	r := ReturnStmt{
		Tok: p.expect(lex.RET),
	}

	for p.peek().TokenType != lex.SEM {
		r.Arg = append(r.Arg, p.parseExpr0())

		if p.peek().TokenType != lex.SEM {
			p.expect(lex.COM)
		}
	}

	p.expect(lex.SEM)

	return &r
}

func (p *parser) parseAssignStmt() *AssignStmt {
	r := AssignStmt{}

	if p.peek().TokenType == lex.VAR {
		p.next()

		r.New = true

		for p.peek().TokenType != lex.EQ {
			r.Var = append(r.Val, &IdfExpr{
				Tok: p.expect(lex.IDF),
			})

			if p.peek().TokenType != lex.EQ {
				p.expect(lex.COM)
			}
		}
	} else {
		for p.peek().TokenType != lex.EQ {
			r.Var = append(r.Val, p.parseExpr6())

			if p.peek().TokenType != lex.EQ {
				p.expect(lex.COM)
			}
		}
	}

	r.Tok = p.expect(lex.EQ)

	for p.peek().TokenType != lex.SEM {
		r.Val = append(r.Val, p.parseExpr0())

		if p.peek().TokenType != lex.SEM {
			p.expect(lex.COM)
		}
	}

	p.expect(lex.SEM)

	return &r
}

func (p *parser) parseLoopStmt() *LoopStmt {
	r := LoopStmt{
		Tok: p.expect(lex.FOR),
	}

	p.expect(lex.LP)

	r.Con = p.parseExpr0()

	p.expect(lex.RP)
	p.expect(lex.LB)

	r.Rep = p.parseBlockStmt()

	p.expect(lex.RB)

	return &r
}

func (p *parser) parseCondStmt() *CondStmt {
	r := CondStmt{
		Tok: p.expect(lex.IF),
	}

	p.expect(lex.LP)

	r.Con = p.parseExpr0()

	p.expect(lex.RP)
	p.expect(lex.LB)

	r.Pos = p.parseBlockStmt()

	p.expect(lex.RB)

	if p.peek().TokenType == lex.ELSE {
		p.next()
		p.expect(lex.LB)

		r.Neg = p.parseBlockStmt()

		p.expect(lex.RB)
	}

	return &r
}

func (p *parser) parseTypeSpec() TypeSpec {
	switch p.peek().TokenType {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDF:
		return &BaseSpec{Tok: p.next()}
	case lex.LB:
		return p.parseCompSpec()
	case lex.FUNC:
		return p.parseFuncSpec()
	}

	panic("type specification expected")
}

func (p *parser) parseFieldSpec() (s FieldSpec) {
	if p.peek().TokenType == lex.IDF {
		s.Tok = p.next()
	}

	s.Typ = p.parseTypeSpec()

	return
}

func (p *parser) parseFuncSpec() *FuncSpec {
	s := &FuncSpec{
		BaseSpec: BaseSpec{
			Tok: p.expect(lex.LP),
		},
	}

	for p.peek().TokenType != lex.RP {
		s.Arg = append(s.Arg, p.parseFieldSpec())

		if p.peek().TokenType != lex.RP {
			p.expect(lex.COM)
		}
	}

	p.expect(lex.RP)

	if p.peek().TokenType == lex.LP {
		p.next()

		for p.peek().TokenType != lex.RP {
			s.Ret = append(s.Ret, p.parseFieldSpec())

			if p.peek().TokenType != lex.RP {
				p.expect(lex.COM)
			}
		}

		p.expect(lex.RP)
	}

	return s
}

func (p *parser) parseCompSpec() *CompSpec {
	s := &CompSpec{
		BaseSpec: BaseSpec{
			Tok: p.expect(lex.LB),
		},
	}

	for p.peek().TokenType != lex.RB {
		s.Fields = append(s.Fields, p.parseFieldSpec())
	}

	p.expect(lex.RB)

	return s
}

func (p *parser) parseFuncDecl() *FuncDecl {
	f := &FuncDecl{}

	p.expect(lex.FUNC)
	f.Tok = p.expect(lex.IDF)
	f.Spec = p.parseFuncSpec()

	p.expect(lex.LB)
	f.Body = p.parseBlockStmt()
	p.expect(lex.RB)

	return f
}

func (p *parser) parseCompDecl() *CompDecl {
	c := &CompDecl{}

	p.expect(lex.TYPE)
	c.Tok = p.expect(lex.IDF)
	c.Spec = p.parseCompSpec()

	return c
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
		case lex.LP:
			switch p.prv.TokenType {
			case lex.I64:
				p.next()

				cur = &ToSigExpr{
					Tok: p.prv,
					X:   p.parseExpr0(),
				}

				p.expect(lex.RP)
			case lex.U64:
				p.next()

				cur = &ToUnsExpr{
					Tok: p.prv,
					X:   p.parseExpr0(),
				}

				p.expect(lex.RP)
			case lex.F64:
				p.next()

				cur = &ToFltExpr{
					Tok: p.prv,
					X:   p.parseExpr0(),
				}

				p.expect(lex.RP)
			default:
				p.next()

				exe := &CallExpr{
					Tok: p.prv,
					Exe: cur,
				}

				for p.peek().TokenType != lex.RP {
					exe.Arg = append(exe.Arg, p.parseExpr0())

					if p.peek().TokenType != lex.RP {
						p.expect(lex.COM)
					}
				}

				p.expect(lex.RP)

				cur = exe
			}

			continue
		}

		break
	}

	return cur
}

func (p *parser) parseExpr7() Expr {
	switch p.peek().TokenType {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDF:
		par := p.next()

		if p.peek().TokenType == lex.LB {
			exp := &CompImmExpr{
				Tok: par,
			}

			for p.peek().TokenType != lex.RB {
				tok := p.next()
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

			return exp
		}

		return &IdfExpr{Tok: par}
	case lex.II64, lex.IU64, lex.IF64, lex.TRUE, lex.FALSE:
		return &BaseImmExpr{Tok: p.next()}
	case lex.LP:
		p.next()
		r := p.parseExpr0()
		p.expect(lex.RP)

		return r
	}

	p.error(fmt.Errorf("value expected"))

	return nil
}
