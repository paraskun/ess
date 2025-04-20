package ast

import (
	"fmt"

	"github.com/paraskun/ess-go/lex"
	"github.com/paraskun/ess-go/tok"
)

type Func struct {
}

type Parser struct {
	Func map[string]*Func

	scn lex.Scanner
	prv tok.Token
}

func (p *Parser) Parse() {
	p.prv = p.scn.Next()

	switch p.prv.TokenType {
	case tok.EOF:
		return
	case tok.FUNC:
		p.parseFunc()
		p.Parse()

		return
	default:
		panic(fmt.Errorf("unexpected token"))
	}
}

func (p *Parser) next() tok.Token {
	p.prv = p.scn.Next()
	return p.prv
}

func (p *Parser) expect(tt tok.TokenType) {
	if p.next().TokenType != tt {
		panic(fmt.Errorf("unexpected token"))
	}
}

func (p *Parser) parseFunc() {
	p.expect(tok.ID)

	f := &Func{}
	p.Func[p.prv.Lit] = f

	p.expect(tok.LP)
	p.parseArgs()
	p.expect(tok.RP)

	p.expect(tok.LSB)
	p.parseStmtList()
	p.expect(tok.RSB)
}

func (prs *Parser) parseArgs()
func (prs *Parser) parseExpr()
func (prs *Parser) parseStmt()
func (prs *Parser) parseStmtList()
func (prs *Parser) parseBranch()
