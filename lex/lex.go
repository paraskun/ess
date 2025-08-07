// Package lex contains language tokenizer.
package lex

import (
	"fmt"
	"unicode"

	"github.com/paraskun/o2/tty"
)

type Lexeme struct {
	Typ Type
	Tok *tty.Tok
}

// Scanner is a source code tokenizer.
//
// Encountered errors stored in Err slice so that they
// can be used later in case of fatal in the following stages.
type Scanner struct {
	buf []rune
	row int
	col int
	prv *Lexeme
}

func (s *Scanner) error(err error) error {
	return fmt.Errorf("scanner: [ %3d:%3d ] %w", s.row, s.col, err)
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
func (s *Scanner) Next() (*Lexeme, error) {
	s.skip()

	t := &Lexeme{Tok: &tty.Tok{}}

	if len(s.buf) == 0 {
		t.Typ = EOF
		return t, nil
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
		t.Typ = LP
	case ')':
		t.Typ = RP
	case '{':
		t.Typ = LB
	case '}':
		t.Typ = RB
	case '[':
		t.Typ = LSB
	case ']':
		t.Typ = RSB
	case ':':
		t.Typ = COL

		if len(s.buf) > 1 && s.buf[1] == '=' {
			t.Typ = INI
			t.Tok.Lit = string(s.buf[0:2])
		}
	case ';':
		t.Typ = SEM
	case ',':
		t.Typ = COM
	case '.':
		t.Typ = DOT
	case '+':
		t.Typ = ADD
	case '-':
		t.Typ = UNEG

		switch s.prv.Typ {
		case IDEN, II64:
			t.Typ = SUB
		}
	case '*':
		t.Typ = MUL

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '*':
				t.Typ = POW
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '/':
		t.Typ = DIV
	case '<':
		t.Typ = LT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '<':
				t.Typ = SHL
				t.Tok.Lit = string(s.buf[0:2])
			case '=':
				t.Typ = LE
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '>':
		t.Typ = GT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '>':
				t.Typ = SHR
				t.Tok.Lit = string(s.buf[0:2])
			case '=':
				t.Typ = GE
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '%':
		t.Typ = MOD
	case '&':
		t.Typ = BAND

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '&':
				t.Typ = LAND
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '|':
		t.Typ = BOR

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '|':
				t.Typ = LOR
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '^':
		t.Typ = BXOR
	case '~':
		t.Typ = BNEG
	case '=':
		t.Typ = EQ

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.Typ = EEQ
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	case '!':
		t.Typ = LNEG

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.Typ = NE
				t.Tok.Lit = string(s.buf[0:2])
			}
		}
	default:
		s.buf = s.buf[len(t.Tok.Lit):]
		s.col += len(t.Tok.Lit)

		return nil, s.error(fmt.Errorf("unexpected symbol"))
	}

	s.prv = t
	s.buf = s.buf[len(t.Tok.Lit):]
	s.col += len(t.Tok.Lit)

	return t, nil
}

func (s *Scanner) nextNum(t *Lexeme) (*Lexeme, error) {
	cur := 1

	t.Typ = II64

	for len(s.buf) > cur && unicode.IsDigit(s.buf[cur]) {
		cur += 1
	}

	if len(s.buf) > cur {
		switch s.buf[cur] {
		case 'u':
			t.Typ = IU64
			cur += 1
		case '.':
			t.Typ = IF64
			cur += 1

			if len(s.buf) <= cur || !unicode.IsDigit(s.buf[cur]) {
				err := s.error(fmt.Errorf("malformed numeric literal"))

				s.col += cur
				s.buf = s.buf[cur:]

				return nil, err
			}

			for len(s.buf) > cur && unicode.IsDigit(s.buf[cur]) {
				cur += 1
			}
		}
	}

	if len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur]) || s.buf[cur] == '.') {
		err := s.error(fmt.Errorf("malformed numeric literal"))

		for len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur])) {
			cur += 1
		}

		s.col += cur
		s.buf = s.buf[cur:]

		return nil, err
	}

	if t.Typ == U64 {
		t.Tok.Lit = string(s.buf[:cur-1])
	} else {
		t.Tok.Lit = string(s.buf[:cur])
	}

	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	return t, nil
}

func (s *Scanner) nextIden(t *Lexeme) (*Lexeme, error) {
	cur := 1
	t.Typ = IDEN

	for len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur])) {
		cur += 1
	}

	t.Tok.Lit = string(s.buf[:cur])
	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	if tt, ok := AsKeyword(t.Tok.Lit); ok {
		t.Typ = tt
	}

	return t, nil
}

func (s *Scanner) nextStr(t *Lexeme) (*Lexeme, error) {
	cur := 1

	for len(s.buf) > cur {
		if s.buf[cur] == '\n' {
			err := s.error(fmt.Errorf("malformed string literal"))

			s.col += cur + 1
			s.buf = s.buf[cur+1:]

			return nil, err
		}

		if s.buf[cur] == '"' && s.buf[cur-1] != '\\' {
			break
		}

		cur += 1
	}

	if len(s.buf) <= cur || s.buf[cur] != '"' {
		err := s.error(fmt.Errorf("malformed string literal"))

		s.col += cur
		s.buf = s.buf[cur:]

		return nil, err
	}

	t.Typ = STR
	t.Tok.Lit = string(s.buf[:cur])
	s.prv = t
	s.col += cur + 1
	s.buf = s.buf[cur+1:]

	return t, nil
}
