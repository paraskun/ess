package ast

import (
	"fmt"
	"io"

	"github.com/paraskun/ess-go/lex"
	"github.com/paraskun/ess-go/tok"
)

type Node interface {
	Write(io.Writer, uint) uint
}

type Stmt interface {
	Node

	stmt()
}

type Expr interface {
	Node

	Type() TypeSpec

	expr()
}

type Block struct {
	*Env

	Body []Stmt
}

func (b *Block) Write(w io.Writer, o uint) (r uint) {
	for _, s := range b.Body {
		r += s.Write(w, o+r)
	}

	return r
}

type Parser struct {
	S lex.Scanner

	env *Env
	buf bool
	prv *tok.Token
	cur *tok.Token
}

func (p *Parser) Parse() *Block {
	r := Block{}

	for p.peek().TokenType != tok.EOF {
		r.Body = append(r.Body, p.parseStmt())
	}

	return &r
}

func (p *Parser) parseBlock() *Block {
	r := Block{}

	r.Env = &Env{Env: p.env}
	p.env = r.Env

	for p.peek().TokenType != tok.RB {
		r.Body = append(r.Body, p.parseStmt())
	}

	p.env = r.Env.Env

	return &r
}

func (p *Parser) error(err error) {
	panic(fmt.Errorf("parser: %d:%d: %w", p.cur.Row, p.cur.Col, err))
}

func (p *Parser) next() *tok.Token {
	if !p.buf {
		p.prv = p.cur
		p.cur = p.S.Next()
	}

	p.buf = false

	return p.cur
}

func (p *Parser) peek() *tok.Token {
	if !p.buf {
		p.buf = true
		p.prv = p.cur
		p.cur = p.S.Next()
	}

	return p.cur
}

func (p *Parser) expect(tt tok.TokenType) *tok.Token {
	if p.next().TokenType != tt {
		p.error(fmt.Errorf("%v given, but %v expected", p.cur.TokenType, tt))
	}

	return p.cur
}
