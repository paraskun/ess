package ast

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/paraskun/ess-go/run"
	"github.com/paraskun/ess-go/tok"
)

type LitExpr struct {
	*tok.Token
	Spec TypeSpec
}

func (e *LitExpr) Type() TypeSpec {
	if e.Spec != nil {
		return e.Spec
	}

	switch e.TokenType {
	case tok.LI64:
		i64, _ := strconv.ParseInt(e.Lit, 10, 64)

		e.Spec = &Sig64{
			Const: i64,
		}
	}

	return e.Spec
}

func (e *LitExpr) Write(w io.Writer, o uint) uint {
	switch sym := e.Spec.(type) {
	case *Bool:
		binary.Write(w, binary.BigEndian, byte(run.LBI))
		binary.Write(w, binary.BigEndian, uint32(0))
		binary.Write(w, binary.BigEndian, sym.Local.Off)
	case *Sig64:
		binary.Write(w, binary.BigEndian, byte(run.LDI))
		binary.Write(w, binary.BigEndian, uint32(0))
		binary.Write(w, binary.BigEndian, sym.Local.Off)
	}

	return 9
}

type IdfExpr struct {
	*tok.Token
	Spec TypeSpec
}

func (e *IdfExpr) Type() TypeSpec {
	return e.Spec
}

func (e *IdfExpr) Write(w io.Writer, o uint) uint {
	switch sym := e.Spec.(type) {
	case *Bool:
		binary.Write(w, binary.BigEndian, byte(run.LBI))
		binary.Write(w, binary.BigEndian, uint32(0))
		binary.Write(w, binary.BigEndian, sym.Local.Off)
	case *Sig64:
		binary.Write(w, binary.BigEndian, byte(run.LDI))
		binary.Write(w, binary.BigEndian, uint32(0))
		binary.Write(w, binary.BigEndian, sym.Local.Off)
	}

	return 9
}

type InfExpr struct {
	*tok.Token
	Spec TypeSpec

	Lt Expr
	Rt Expr
}

func (e *InfExpr) Type() TypeSpec {
	if e.Spec != nil {
		return e.Spec
	}

	return e.Lt.Type()
}

func (e *InfExpr) Write(w io.Writer, o uint) (r uint) {
	rev := false

	switch e.TokenType {
	case tok.ADD:
		w.Write([]byte{byte(run.IADD)})
	case tok.SUB:
		w.Write([]byte{byte(run.ISUB)})
	case tok.MUL:
		w.Write([]byte{byte(run.IMUL)})
	case tok.DIV:
		w.Write([]byte{byte(run.IDIV)})
	case tok.POW:
		w.Write([]byte{byte(run.IPOW)})
	case tok.SHL:
		w.Write([]byte{byte(run.ISHL)})
	case tok.SHR:
		w.Write([]byte{byte(run.ISHR)})
	case tok.MOD:
		w.Write([]byte{byte(run.IMOD)})
	case tok.BAND:
		w.Write([]byte{byte(run.IAND)})
	case tok.BOR:
		w.Write([]byte{byte(run.IOR)})
	case tok.BXOR:
		w.Write([]byte{byte(run.IXOR)})
	case tok.LT:
		w.Write([]byte{byte(run.ILT)})
	case tok.GT:
		rev = true
		w.Write([]byte{byte(run.ILE)})
	case tok.LE:
		w.Write([]byte{byte(run.ILE)})
	case tok.GE:
		rev = true
		w.Write([]byte{byte(run.ILT)})
	case tok.NE:
		w.Write([]byte{byte(run.INE)})
	case tok.EEQ:
		w.Write([]byte{byte(run.IEQ)})
	case tok.LAND:
		w.Write([]byte{byte(run.LAND)})
	case tok.LOR:
		w.Write([]byte{byte(run.LOR)})
	}

	if rev {
		r += e.Lt.Write(w, o)
		r += e.Rt.Write(w, o)
	} else {
		r += e.Rt.Write(w, o)
		r += e.Lt.Write(w, o)
	}

	return r + 1
}

type PfxExpr struct {
	*tok.Token
	Spec TypeSpec

	Md Expr
}

func (e *PfxExpr) Type() TypeSpec {
	if e.Spec != nil {
		return e.Spec
	}

	return e.Md.Type()
}

func (e *PfxExpr) Write(w io.Writer, o uint) (r uint) {
	r += e.Md.Write(w, o)

	switch e.TokenType {
	case tok.BNEG:
		w.Write([]byte{byte(run.IBNEG)})
	case tok.UNEG:
		w.Write([]byte{byte(run.IUNEG)})
	case tok.LNEG:
		w.Write([]byte{byte(run.LNEG)})
	}

	return r + 1
}

func (*LitExpr) expr() {}
func (*IdfExpr) expr() {}
func (*InfExpr) expr() {}
func (*PfxExpr) expr() {}

func (p *Parser) parseExpr() Expr {
	return p.parseExpr0()
}

func (p *Parser) parseExpr0() Expr {
	cur := p.parseExpr1()

	for {
		switch p.peek().TokenType {
		case tok.LAND, tok.LOR:
			cur = &InfExpr{
				Token: p.next(),
				Lt:    cur,
				Rt:    p.parseExpr1(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *Parser) parseExpr1() Expr {
	cur := p.parseExpr2()

	for {
		switch p.peek().TokenType {
		case tok.LT, tok.GT, tok.LE, tok.GE, tok.NE, tok.EEQ:
			cur = &InfExpr{
				Token: p.next(),
				Spec:  cur.Type(),
				Lt:    cur,
				Rt:    p.parseExpr2(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *Parser) parseExpr2() Expr {
	cur := p.parseExpr3()

	for {
		switch p.peek().TokenType {
		case tok.SHL, tok.SHR, tok.BAND, tok.BOR, tok.BXOR:
			cur = &InfExpr{
				Token: p.next(),
				Lt:    cur,
				Rt:    p.parseExpr3(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *Parser) parseExpr3() Expr {
	cur := p.parseExpr4()

	for {
		switch p.peek().TokenType {
		case tok.ADD, tok.SUB:
			cur = &InfExpr{
				Token: p.next(),
				Lt:    cur,
				Rt:    p.parseExpr4(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *Parser) parseExpr4() Expr {
	cur := p.parseExpr5()

	for {
		switch p.peek().TokenType {
		case tok.MUL, tok.DIV, tok.MOD, tok.POW:
			cur = &InfExpr{
				Token: p.next(),
				Lt:    cur,
				Rt:    p.parseExpr5(),
			}

			continue
		}

		break
	}

	return cur
}

func (p *Parser) parseExpr5() Expr {
	switch p.peek().TokenType {
	case tok.LNEG, tok.BNEG, tok.UNEG:
		return &PfxExpr{
			Token: p.next(),
			Md:    p.parseExpr6(),
		}
	}

	return p.parseExpr6()
}

func (p *Parser) parseExpr6() Expr {
	switch p.peek().TokenType {
	case tok.IDF:
		return &IdfExpr{Token: p.next()}
	case tok.LI64, tok.TRUE, tok.FALSE:
		return &LitExpr{Token: p.next()}
	case tok.LP:
		p.next()
		r := p.parseExpr0()
		p.expect(tok.RP)

		return r
	}

	p.error(fmt.Errorf("value expected"))

	return nil
}
