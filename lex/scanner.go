package lex

import (
	"fmt"
	"unicode"

	"github.com/paraskun/ess-go/tok"
)

type Scanner struct {
	Err []error

	buf []rune
	row int
	col int

	prv *tok.Token
}

func (s *Scanner) Load(buf []rune) {
	s.buf = buf
	s.row = 1
	s.col = 1
}

func (s *Scanner) Next() *tok.Token {
	s.skip()

	t := &tok.Token{
		Row: s.row,
		Col: s.col,
	}

	if len(s.buf) == 0 {
		t.TokenType = tok.EOF
		return t
	}

	if unicode.IsDigit(s.buf[0]) {
		return s.nextNum(t)
	}

	if unicode.IsLetter(s.buf[0]) {
		return s.nextIdf(t)
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
		case tok.IDF, tok.LI64:
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
		t.TokenType = tok.BNEG
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
		s.error(fmt.Errorf("unexpected symbol"))
		return s.Next()
	}

	s.prv = t
	s.buf = s.buf[len(t.Lit):]
	s.col += len(t.Lit)

	return t
}

func (s *Scanner) error(err error) {
	s.Err = append(s.Err, fmt.Errorf("scanner: %d:%d: %w", s.row, s.col, err))
}

func (s *Scanner) skip() {
	for len(s.buf) != 0 && unicode.IsSpace(s.buf[0]) {
		s.col += 1

		if s.buf[0] == '\n' {
			s.row += 1
			s.col = 1
		}

		s.buf = s.buf[1:]
	}
}

func (s *Scanner) nextNum(t *tok.Token) *tok.Token {
	cur := 1

	t.TokenType = tok.LI64

	for unicode.IsDigit(s.buf[cur]) {
		cur += 1
	}

	if unicode.IsLetter(s.buf[cur]) {
		s.error(fmt.Errorf("malformed numeric literal"))

		for unicode.IsLetter(s.buf[cur]) {
			cur += 1
		}

		s.col += cur
		s.buf = s.buf[cur:]

		return s.Next()
	}

	t.Lit = string(s.buf[:cur])
	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	return t
}

func (s *Scanner) nextIdf(t *tok.Token) *tok.Token {
	cur := 1

	for unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur]) {
		cur += 1
	}

	t.TokenType = tok.IDF
	t.Lit = string(s.buf[:cur])
	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	if tt, ok := tok.AsKeyword(t.Lit); ok {
		t.TokenType = tt
	}

	return t
}
