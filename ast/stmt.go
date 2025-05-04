package ast

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/paraskun/ess-go/run"
	"github.com/paraskun/ess-go/tok"
)

type AsgnStmt struct {
	*tok.Token

	Idf Expr
	Val Expr
}

func (s *AsgnStmt) Write(w io.Writer, o uint) (r uint) {
	r += s.Val.Write(w, o)

	switch sym := s.Idf.(*IdfExpr).Spec.(type) {
	case *Bool:
		binary.Write(w, binary.BigEndian, byte(run.SBI))
		binary.Write(w, binary.BigEndian, uint32(0))
		binary.Write(w, binary.BigEndian, sym.Local.Off)
	case *Sig64:
		binary.Write(w, binary.BigEndian, byte(run.SDI))
		binary.Write(w, binary.BigEndian, uint32(0))
		binary.Write(w, binary.BigEndian, sym.Local.Off)
	}

	return r + 5
}

type LoopStmt struct {
	*tok.Token

	Rep  *Block
	Cond Expr
}

func (s *LoopStmt) Write(w io.Writer, o uint) (r uint) {
	bb := bytes.Buffer{}

	r += s.Cond.Write(w, o) + 5
	r += s.Rep.Write(&bb, o+r) + 5

	binary.Write(w, binary.LittleEndian, byte(run.JIF))
	binary.Write(w, binary.LittleEndian, uint32(r))

	w.Write(bb.Bytes())

	binary.Write(w, binary.LittleEndian, byte(run.JMP))
	binary.Write(w, binary.LittleEndian, uint32(o))

	return r
}

type CondStmt struct {
	*tok.Token

	Pos  *Block
	Neg  *Block
	Cond Expr
}

func (s *CondStmt) Write(w io.Writer, o uint) (r uint) {
	tb := bytes.Buffer{}
	fb := bytes.Buffer{}

	r += s.Cond.Write(w, o) + 5
	r += s.Pos.Write(&tb, o+r) + 5

	binary.Write(w, binary.LittleEndian, byte(run.JIF))
	binary.Write(w, binary.LittleEndian, uint32(r))

	if s.Neg != nil {
		r += s.Neg.Write(&fb, o+r)
	}

	w.Write(tb.Bytes())
	binary.Write(w, binary.LittleEndian, byte(run.JMP))
	binary.Write(w, binary.LittleEndian, uint32(r))
	w.Write(fb.Bytes())

	return r
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
	r := AsgnStmt{Idf: p.parseExpr()}
	v, ok := r.Idf.(*IdfExpr)

	if !ok {
		p.error(fmt.Errorf("could not assign to literal"))
	}

	if ts := p.env.Lookup(v.Lit); ts != nil {
		v.Spec = ts
	}

	r.Token = p.expect(tok.EQ)
	r.Val = p.parseExpr()

	if v.Spec == nil {
		p.env.Sym = append(p.env.Sym, r.Val.Type())
		p.env.Map[v.Lit] = len(p.env.Sym) - 1
	}

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

	r.Rep = p.parseBlock()

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

	r.Pos = p.parseBlock()

	p.expect(tok.RB)

	if p.peek().TokenType == tok.ELSE {
		p.next()
		p.expect(tok.LB)

		r.Neg = p.parseBlock()

		p.expect(tok.RB)
	}

	return &r
}
