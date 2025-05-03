package ast

import (
	"fmt"
	"io"

	"github.com/paraskun/ess-go/tok"
)

type AsgnStmt struct {
	*tok.Token

	Idf Expr
	Val Expr
}

func (s *AsgnStmt) Debug(w io.Writer) {
	s.Idf.Debug(w)
	s.Val.Debug(w)
	fmt.Fprintf(w, "%v\n", s.TokenType)
}

type LoopStmt struct {
	*tok.Token

	Mb   *Block
	Cond Expr
}

func (s *LoopStmt) Debug(w io.Writer) {
	fmt.Fprintf(w, "for\n")
	s.Cond.Debug(w)
	fmt.Fprintf(w, "do\n")
	s.Mb.Debug(w)
	fmt.Fprintf(w, "end\n")
}

type CondStmt struct {
	*tok.Token

	Tb   *Block
	Fb   *Block
	Cond Expr
}

func (s *CondStmt) Debug(w io.Writer) {
	fmt.Fprintf(w, "if\n")
	s.Cond.Debug(w)
	fmt.Fprintf(w, "do\n")
	s.Tb.Debug(w)

	if s.Fb != nil {
		fmt.Fprintf(w, "else do\n")
		s.Fb.Debug(w)
	}

	fmt.Fprintf(w, "end\n")
}

func (*AsgnStmt) stmt() {}
func (*LoopStmt) stmt() {}
func (*CondStmt) stmt() {}

func (p *Parser) parseStmt() Stmt {
	switch p.peek().TokenType {
	case tok.IDF:
		return p.parseAsgn()
	case tok.FOR:
		return p.parseLoop()
	case tok.IF:
		return p.parseCond()
	}

	p.error(fmt.Errorf("statement expected"))

	return nil
}

func (p *Parser) parseAsgn() *AsgnStmt {
	r := AsgnStmt{}

	r.Idf = p.parseExpr()
	r.Token = p.expect(tok.EQ)
	r.Val = p.parseExpr()

	p.expect(tok.SEM)

	return &r
}

func (p *Parser) parseLoop() *LoopStmt {
	r := LoopStmt{}

	r.Token = p.expect(tok.FOR)

	p.expect(tok.LP)

	r.Cond = p.parseExpr()

	p.expect(tok.RP)
	p.expect(tok.LB)

	r.Mb = p.parseBlock()

	p.expect(tok.RB)

	return &r
}

func (p *Parser) parseCond() *CondStmt {
	r := CondStmt{}

	r.Token = p.expect(tok.IF)

	p.expect(tok.LP)

	r.Cond = p.parseExpr()

	p.expect(tok.RP)
	p.expect(tok.LB)

	r.Tb = p.parseBlock()

	p.expect(tok.RB)

	if p.peek().TokenType == tok.ELSE {
		p.next()
		p.expect(tok.LB)

		r.Fb = p.parseBlock()

		p.expect(tok.RB)
	}

	return &r
}
