package tty

import (
	"fmt"
)

type Span interface {
	Size() (int, int)

	Render(b [][]rune, bx, by int)
}

func Print(s Span) {
	w, h := s.Size()
	b := make([][]rune, h)

	for i := range h {
		b[i] = make([]rune, w)
	}

	s.Render(b, 0, 0)

	for _, l := range b {
		for i := range len(l) {
			if l[i] == 0 {
				l[i] = ' '
			}
		}

		fmt.Println(string(l))
	}
}
