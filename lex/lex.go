package lex

import (
	"fmt"
	"unicode"

	"github.com/fatih/color"
	"github.com/paraskun/o2/tty"
	"github.com/paraskun/o2/typ"
)

type Lexeme struct {
	Typ Type
	Tok *tty.Tok
}

type Scanner struct {
	buf []rune
	row int
	col int
	prv *Lexeme
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

func (s *Scanner) Next() (*Lexeme, *typ.Error) {
	s.skip()

	t := &Lexeme{Tok: &tty.Tok{
		Pos: tty.Position{
			Row: s.row,
			Col: s.col,
		},
	}}

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
	case '@':
		t.Typ = DOG
	default:
		s.buf = s.buf[len(t.Tok.Lit):]
		s.col += len(t.Tok.Lit)

		t.Typ = ERR
		t.Tok.Hint().Text = "unknown symbol"
		t.Tok.Hint().Attr.Color = *color.New(color.FgRed)

		return nil, &typ.Error{
			Span: t.Tok,
			Full: fmt.Sprintf("scanner: unknown symbol at %d:%d", t.Tok.Pos.Row, t.Tok.Pos.Col),
			Help: "verify your input or consider creating string literal",
		}
	}

	s.prv = t
	s.buf = s.buf[len(t.Tok.Lit):]
	s.col += len(t.Tok.Lit)

	return t, nil
}

func (s *Scanner) nextNum(t *Lexeme) (*Lexeme, *typ.Error) {
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
				t.Typ = ERR

				if len(s.buf) > cur && s.buf[cur] != '\n' {
					t.Tok.Lit = string(s.buf[:cur+1])
				} else {
					t.Tok.Lit = string(s.buf[:cur])
				}

				t.Tok.Hint().Text = "malformed floating point literal"
				t.Tok.Hint().Attr.Color = *color.New(color.FgRed)

				s.buf = s.buf[cur:]
				s.col += cur

				return nil, &typ.Error{
					Span: t.Tok,
					Full: fmt.Sprintf("scanner: malfomed floating point literal at %d:%d",
						t.Tok.Pos.Row,
						t.Tok.Pos.Col),
					Help: "consider specifying at least one fraction digit",
				}
			}

			for len(s.buf) > cur && unicode.IsDigit(s.buf[cur]) {
				cur += 1
			}
		}
	}

	if len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur]) || s.buf[cur] == '.') {
		for len(s.buf) > cur && (unicode.IsLetter(s.buf[cur]) || unicode.IsDigit(s.buf[cur]) || s.buf[cur] == '.') {
			cur += 1
		}

		t.Typ = ERR
		t.Tok.Lit = string(s.buf[:cur])
		t.Tok.Hint().Text = "malformed numeric literal"
		t.Tok.Hint().Attr.Color = *color.New(color.FgRed)

		s.col += cur
		s.buf = s.buf[cur:]

		return nil, &typ.Error{
			Span: t.Tok,
			Full: fmt.Sprintf("scanner: malfomed numeric literal at %d:%d",
				t.Tok.Pos.Row,
				t.Tok.Pos.Col),
			Help: "verify surrounding expression correctness",
		}
	}

	t.Tok.Lit = string(s.buf[:cur])

	s.prv = t
	s.col += cur
	s.buf = s.buf[cur:]

	return t, nil
}

func (s *Scanner) nextIden(t *Lexeme) (*Lexeme, *typ.Error) {
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

func (s *Scanner) nextStr(t *Lexeme) (*Lexeme, *typ.Error) {
	cur := 1

	for len(s.buf) > cur {
		if s.buf[cur] == '\n' {
			s.col += cur + 1
			s.buf = s.buf[cur+1:]

			t.Typ = ERR
			t.Tok.Lit = string(s.buf[:cur])
			t.Tok.Hint().Text = "malformed (broken) string literal"
			t.Tok.Hint().Attr.Color = *color.New(color.FgRed)

			return nil, &typ.Error{
				Span: t.Tok,
				Full: fmt.Sprintf("scanner: malfomed string literal at %d:%d", t.Tok.Pos.Row, t.Tok.Pos.Col),
				Help: "consider inlining the string",
			}
		}

		if s.buf[cur] == '"' && s.buf[cur-1] != '\\' {
			break
		}

		cur += 1
	}

	if len(s.buf) <= cur || s.buf[cur] != '"' {
		s.col += cur
		s.buf = s.buf[cur:]

		t.Typ = ERR
		t.Tok.Lit = string(s.buf[:cur])
		t.Tok.Hint().Text = "malformed string literal"
		t.Tok.Hint().Attr.Color = *color.New(color.FgRed)

		return nil, &typ.Error{
			Span: t.Tok,
			Full: fmt.Sprintf("scanner: malfomed string literal at %d:%d", t.Tok.Pos.Row, t.Tok.Pos.Col),
			Help: "consider finishing the string with a closing quote",
		}

	}

	t.Typ = ISTR
	t.Tok.Lit = string(s.buf[:cur+1])
	s.prv = t
	s.col += cur + 1
	s.buf = s.buf[cur+1:]

	return t, nil
}
