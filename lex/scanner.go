package lex

import (
	"fmt"
	"unicode"

	"github.com/paraskun/ess-go/tok"
)

type Scanner struct {
	buf []rune
	prv tok.Token

	row int
	col int
}

func (s *Scanner) Load(buf []rune) {
	s.buf = buf
	s.row = 1
	s.col = 1
}

func (s *Scanner) Next() (t tok.Token) {
	s.skip()

	t.Row = s.row
	t.Col = s.col

	if len(s.buf) == 0 {
		t.TokenType = tok.EOF
		return
	}

	if unicode.IsDigit(s.buf[0]) {
		s.nextNum(&t)
		return
	}

	if unicode.IsLetter(s.buf[0]) {
		s.nextIdf(&t)
		return
	}

	if s.buf[0] == '"' {
		s.nextStr(&t)
		return
	}

	t.Lit = string(s.buf[0:1])

	switch s.buf[0] {
	case '(':
		t.TokenType = tok.LP
		break
	case ')':
		t.TokenType = tok.RP
		break
	case '{':
		t.TokenType = tok.LB
		break
	case '}':
		t.TokenType = tok.RB
		break
	case '[':
		t.TokenType = tok.LSB
		break
	case ']':
		t.TokenType = tok.RSB
		break
	case ':':
		t.TokenType = tok.COL
		break
	case ';':
		t.TokenType = tok.SEM
		break
	case ',':
		t.TokenType = tok.COM
		break
	case '+':
		t.TokenType = tok.ADD
		break
	case '-':
		switch s.prv.TokenType {
		case tok.ID, tok.INT, tok.FLT:
			t.TokenType = tok.SUB
			break
		default:
			t.TokenType = tok.UNEG
			break
		}

		break
	case '*':
		t.TokenType = tok.MUL

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '*':
				t.TokenType = tok.POW
				t.Lit = string(s.buf[0:2])

				break
			}
		}

		break
	case '/':
		t.TokenType = tok.DIV
		break
	case '<':
		t.TokenType = tok.LT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '<':
				t.TokenType = tok.SHL
				t.Lit = string(s.buf[0:2])

				break
			case '=':
				t.TokenType = tok.LE
				t.Lit = string(s.buf[0:2])

				break
			}
		}

		break
	case '>':
		t.TokenType = tok.GT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '>':
				t.TokenType = tok.SHR
				t.Lit = string(s.buf[0:2])

				break
			case '=':
				t.TokenType = tok.GE
				t.Lit = string(s.buf[0:2])

				break
			}
		}

		break
	case '%':
		t.TokenType = tok.MOD
		break
	case '&':
		t.TokenType = tok.BAND

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '&':
				t.TokenType = tok.LAND
				t.Lit = string(s.buf[0:2])

				break
			}
		}

		break
	case '|':
		t.TokenType = tok.BOR

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '|':
				t.TokenType = tok.LOR
				t.Lit = string(s.buf[0:2])

				break
			}
		}

		break
	case '^':
		t.TokenType = tok.BXOR
		break
	case '~':
		t.TokenType = tok.BINV
		break
	case '=':
		t.TokenType = tok.EQ

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.TokenType = tok.EEQ
				t.Lit = string(s.buf[0:2])

				break
			}
		}

		break
	case '!':
		t.TokenType = tok.LNEG

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.TokenType = tok.NE
				t.Lit = string(s.buf[0:2])

				break
			}
		}

		break
	default:
		t.TokenType = tok.ERR
		t.Err = fmt.Errorf("unknown symbol")

		return
	}

	s.prv = t
	s.buf = s.buf[len(t.Lit):]

	return
}

func (s *Scanner) skip() {
	for unicode.IsSpace(s.buf[0]) {
		s.buf = s.buf[1:]
	}
}

func (s *Scanner) nextNum(t *tok.Token) {
	cur := 1

	t.TokenType = tok.INT

	for unicode.IsDigit(s.buf[cur]) {
		cur += 1
	}

	if s.buf[cur] == '.' {
		cur += 1

		for unicode.IsDigit(s.buf[cur]) {
			cur += 1
		}

		t.TokenType = tok.FLT
	}

	t.Lit = string(s.buf[:cur])
	s.prv = *t
	s.buf = s.buf[cur:]

	return
}

func (s *Scanner) nextStr(t *tok.Token) {
	cur := 1

	for s.buf[cur] != '"' {
		cur += 1
	}

	t.TokenType = tok.STR
	t.Lit = string(s.buf[1:cur])
	s.prv = *t
	s.buf = s.buf[cur+1:]

	return
}

func (s *Scanner) nextIdf(t *tok.Token) {
	cur := 1

	for unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur]) {
		cur += 1
	}

	t.TokenType = tok.ID
	t.Lit = string(s.buf[:cur])

	if tt, ok := tok.AsKeyword(t.Lit); ok {
		t.TokenType = tt
	}

	s.prv = *t
	s.buf = s.buf[cur:]

	return
}
