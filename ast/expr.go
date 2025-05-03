package ast

import (
	"fmt"
	"io"

	"github.com/paraskun/ess-go/tok"
)

type LitExpr struct {
	*tok.Token
}

func (e *LitExpr) Debug(w io.Writer) {
	fmt.Fprintf(w, "%v <%v>\n", e.TokenType, e.Lit)
}

type IdfExpr struct {
	*tok.Token
}

func (e *IdfExpr) Debug(w io.Writer) {
	fmt.Fprintf(w, "%v <%s>\n", e.TokenType, e.Lit)
}

type InfExpr struct {
	*tok.Token

	Lo Expr
	Ro Expr
}

func (e *InfExpr) Debug(w io.Writer) {
	e.Lo.Debug(w)
	e.Ro.Debug(w)

	fmt.Fprintf(w, "%v\n", e.TokenType)
}

type PfxExpr struct {
	*tok.Token

	Mo Expr
}

func (e *PfxExpr) Debug(w io.Writer) {
	e.Mo.Debug(w)
	fmt.Fprintf(w, "%v\n", e.TokenType)
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
		case tok.LAND:
		case tok.LOR:
			p.next()

			cur = &InfExpr{
				Token: p.cur,
				Lo:    cur,
				Ro:    p.parseExpr1(),
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
			p.next()

			cur = &InfExpr{
				Token: p.cur,
				Lo:    cur,
				Ro:    p.parseExpr2(),
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
			p.next()

			cur = &InfExpr{
				Token: p.cur,
				Lo:    cur,
				Ro:    p.parseExpr3(),
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
			p.next()

			cur = &InfExpr{
				Token: p.cur,
				Lo:    cur,
				Ro:    p.parseExpr4(),
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
			p.next()

			cur = &InfExpr{
				Token: p.cur,
				Lo:    cur,
				Ro:    p.parseExpr5(),
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
		p.next()

		return &PfxExpr{
			Token: p.cur,
			Mo:    p.parseExpr6(),
		}
	}

	return p.parseExpr6()
}

func (p *Parser) parseExpr6() Expr {
	switch p.peek().TokenType {
	case tok.IDF:
		p.next()

		return &IdfExpr{
			Token: p.cur,
		}
	case tok.LI64, tok.TRUE, tok.FALSE:
		p.next()

		return &LitExpr{
			Token: p.cur,
		}
	case tok.LP:
		p.next()
		r := p.parseExpr0()
		p.expect(tok.RP)

		return r
	}

	p.error(fmt.Errorf("value expected"))

	return nil
}
