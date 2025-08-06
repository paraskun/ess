// Package lex contains language tokenizer.
package lex

import (
	"fmt"
	"unicode"

	"github.com/paraskun/o2/tty"
)

type Lexeme struct {
	Type

	Tok *tty.Tok
}

// Scanner is a source code tokenizer.
//
// Encountered errors stored in Err slice so that they
// can be used later in case of fatal in the following stages.
type Scanner struct {
	Err []error

	buf []rune
	row int
	col int
	prv *Lexeme
}

func (s *Scanner) error(err error) {
	s.Err = append(s.Err, fmt.Errorf("scanner: [ %3d:%3d ] %w", s.row, s.col, err))
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

func (s *Scanner) Load(buf []rune) {
	s.buf = buf
	s.row = 1
	s.col = 1
}

// Next returns the next lexeme.
//
// In case of an error, returns the next correct
// token (or EOF, if no such left).
func (s *Scanner) Next() *Lexeme {
	s.skip()

	t := &Lexeme{Tok: &tty.Tok{}}

	if len(s.buf) == 0 {
		t.Type = EOF
		return t
	}

	if unicode.IsDigit(s.buf[0]) {
		return s.nextNum(t)
	}

	if unicode.IsLetter(s.buf[0]) {
		return s.nextIden(t)
	}

	if s.buf[0] == '"' {
		return s.nextStr(t)
	}

	t.Tok.Lit = string(s.buf[0:1])

	switch s.buf[0] {
	case '(':
		t.Type = LP
	case ')':
		t.Type = RP
	case '{':
		t.Type = LB
	case '}':
		t.Type = RB
	case '[':
		t.Type = LSB
	case ']':
		t.Type = RSB
	case ':':
		t.Type = COL

		if len(s.buf) > 1 && s.buf[1] == '=' {
			t.Type = INI
			t.Tok.Lit = string(s.buf[0:2])
		}
	case ';':
		t.Type = SEM
	case ',':
		t.Type = COM
	case '.':
		t.Type = DOT
	case '+':
		t.Type = ADD
	case '-':
		t.Type = UNEG

		switch s.prv.Type {
		case IDEN, II64:
			t.Type = SUB
		}
	case '*':
		t.Type = MUL

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '*':
				t.Type = POW
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '/':
		t.Type = DIV
	case '<':
		t.Type = LT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '<':
				t.Type = SHL
				t.Tok.Lit = string(s.buf[0:2])
			case '=':
				t.Type = LE
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '>':
		t.Type = GT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '>':
				t.Type = SHR
				t.Tok.Lit = string(s.buf[0:2])
			case '=':
				t.Type = GE
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '%':
		t.Type = MOD
	case '&':
		t.Type = BAND

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '&':
				t.Type = LAND
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '|':
		t.Type = BOR

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '|':
				t.Type = LOR
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '^':
		t.Type = BXOR
	case '~':
		t.Type = BNEG
	case '=':
		t.Type = EQ

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.Type = EEQ
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '!':
		t.Type = LNEG

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.Type = NE
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	default:
		s.error(fmt.Errorf("unexpected symbol"))
		return s.Next()
	}

	s.prv = t
	s.buf = s.buf[len(t.Tok.Lit):]
	s.col += len(t.Tok.Lit)

	return t
}

func (s *Scanner) nextNum(t *Lexeme) *Lexeme {
	cur := 1

	t.Type = II64

	for len(s.buf) > cur && unicode.IsDigit(s.buf[cur]) {
		cur += 1
	}

	if len(s.buf) > cur {
		switch s.buf[cur] {
		case 'u':
			t.Type = IU64
			cur += 1
		case '.':
			t.Type = IF64
			cur += 1

			if len(s.buf) <= cur || !unicode.IsDigit(s.buf[cur]) {
				s.error(fmt.Errorf("malformed numeric literal"))

				s.col += cur
				s.buf = s.buf[cur:]

				return s.Next()
			}

			for len(s.buf) > cur && unicode.IsDigit(s.buf[cur]) {
				cur += 1
			}
		}
	}

	if len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur]) || s.buf[cur] == '.') {
		s.error(fmt.Errorf("malformed numeric literal"))

		for len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur])) {
			cur += 1
		}

		s.col += cur
		s.buf = s.buf[cur:]

		return s.Next()
	}

	if t.Type == U64 {
		t.Tok.Lit = string(s.buf[:cur-1])
	} else {
		t.Tok.Lit = string(s.buf[:cur])
	}

	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	return t
}

func (s *Scanner) nextIden(t *Lexeme) *Lexeme {
	cur := 1
	t.Type = IDEN

	for len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur])) {
		cur += 1
	}

	t.Tok.Lit = string(s.buf[:cur])
	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	if tt, ok := AsKeyword(t.Tok.Lit); ok {
		t.Type = tt
	}

	return t
}

func (s *Scanner) nextStr(t *Lexeme) *Lexeme {
	cur := 1

	for len(s.buf) > cur {
		if s.buf[cur] == '"' && s.buf[cur-1] != '\\' {
			break
		}

		cur += 1
	}

	if len(s.buf) <= cur || s.buf[cur] != '"' {
		s.error(fmt.Errorf("malformed string literal"))

		s.col += cur
		s.buf = s.buf[cur:]

		return s.Next()
	}

	t.Type = STR
	t.Tok.Lit = string(s.buf[1:cur])
	s.prv = t
	s.col += cur + 1
	s.buf = s.buf[cur+1:]

	return t
}
