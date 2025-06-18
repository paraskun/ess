// Package lex contains language tokenizer.
package lex

import (
	"fmt"
	"unicode"
)

// Scanner is a source code tokenizer.
//
// Encountered errors stored in Err slice so that they
// can be used later in case of fatal in the following stages.
type Scanner struct {
	Err []error

	buf []rune
	row int
	col int
	prv *Token
}

func (s *Scanner) Load(buf []rune) {
	s.buf = buf
	s.row = 1
	s.col = 1
}

// Next returns the next token.
//
// In case of an error, returns the next correct
// token (or EOF, if no such left).
func (s *Scanner) Next() *Token {
	s.skip()

	t := &Token{
		Row: s.row,
		Col: s.col,
	}

	if len(s.buf) == 0 {
		t.TokenType = EOF
		return t
	}

	if unicode.IsDigit(s.buf[0]) {
		return s.nextNum(t)
	}

	if unicode.IsLetter(s.buf[0]) {
		return s.nextIdf(t)
	}

	if s.buf[0] == '"' {
		return s.nextStr(t)
	}

	t.Lit = string(s.buf[0:1])

	switch s.buf[0] {
	case '(':
		t.TokenType = LP
	case ')':
		t.TokenType = RP
	case '{':
		t.TokenType = LB
	case '}':
		t.TokenType = RB
	case '[':
		t.TokenType = LSB
	case ']':
		t.TokenType = RSB
	case ':':
		t.TokenType = COL

		if len(s.buf) > 1 && s.buf[1] == '=' {
			t.TokenType = INI
			t.Lit = string(s.buf[0:2])
		}
	case ';':
		t.TokenType = SEM
	case ',':
		t.TokenType = COM
	case '.':
		t.TokenType = DOT
	case '+':
		t.TokenType = ADD
	case '-':
		t.TokenType = UNEG

		switch s.prv.TokenType {
		case IDF, II64:
			t.TokenType = SUB
		}
	case '*':
		t.TokenType = MUL

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '*':
				t.TokenType = POW
				t.Lit = string(s.buf[0:2])
			}
		}
	case '/':
		t.TokenType = DIV
	case '<':
		t.TokenType = LT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '<':
				t.TokenType = SHL
				t.Lit = string(s.buf[0:2])
			case '=':
				t.TokenType = LE
				t.Lit = string(s.buf[0:2])
			}
		}
	case '>':
		t.TokenType = GT

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '>':
				t.TokenType = SHR
				t.Lit = string(s.buf[0:2])
			case '=':
				t.TokenType = GE
				t.Lit = string(s.buf[0:2])
			}
		}
	case '%':
		t.TokenType = MOD
	case '&':
		t.TokenType = BAND

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '&':
				t.TokenType = LAND
				t.Lit = string(s.buf[0:2])
			}
		}
	case '|':
		t.TokenType = BOR

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '|':
				t.TokenType = LOR
				t.Lit = string(s.buf[0:2])
			}
		}
	case '^':
		t.TokenType = BXOR
	case '~':
		t.TokenType = BNEG
	case '=':
		t.TokenType = EQ

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.TokenType = EEQ
				t.Lit = string(s.buf[0:2])
			}
		}
	case '!':
		t.TokenType = LNEG

		if len(s.buf) > 1 {
			switch s.buf[1] {
			case '=':
				t.TokenType = NE
				t.Lit = string(s.buf[0:2])
			}
		}
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

func (s *Scanner) nextNum(t *Token) *Token {
	cur := 1

	t.TokenType = II64

	for len(s.buf) > cur && unicode.IsDigit(s.buf[cur]) {
		cur += 1
	}

	if len(s.buf) > cur {
		switch s.buf[cur] {
		case 'u':
			t.TokenType = IU64
			cur += 1
		case '.':
			t.TokenType = IF64
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

	if t.TokenType == U64 {
		t.Lit = string(s.buf[:cur-1])
	} else {
		t.Lit = string(s.buf[:cur])
	}

	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	return t
}

func (s *Scanner) nextIdf(t *Token) *Token {
	cur := 1
	t.TokenType = IDF

	if unicode.IsUpper(s.buf[0]) {
		t.TokenType = IMEM
	}

	for len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur])) {
		cur += 1
	}

	t.Lit = string(s.buf[:cur])
	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	if tt, ok := AsKeyword(t.Lit); ok {
		t.TokenType = tt
	}

	return t
}

func (s *Scanner) nextStr(t *Token) *Token {
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

	t.TokenType = STR
	t.Lit = string(s.buf[1:cur])
	s.prv = t
	s.col += cur + 1
	s.buf = s.buf[cur+1:]

	return t
}
