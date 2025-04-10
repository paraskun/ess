package lex

import (
	"github.com/paraskun/ess-go/tok"
)

type Scanner struct {
	buf []rune

	row int
	col int
}

func (s *Scanner) Load(buf []rune) {
	s.buf = buf
	s.row = 1
	s.col = 1
}

func (s *Scanner) Next() (tok tok.Token) {
	s.skip()

	return
}

func (s *Scanner) skip() {
	// skip whitespace, comments
}

func (s *Scanner) nextNum() {
}

func (s *Scanner) nextStr() {
}
