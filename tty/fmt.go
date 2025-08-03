package tty

import "unicode/utf8"

type Hint struct {
	Text string
}

type Frame struct {
	Text string
	Span Span
}

func (f *Frame) Render(b [][]rune, bx, by int) {
	sx, sy := f.Size()
	_, ey := bx+sx-1, by+sy-1

	b[by][bx] = '╭'
	b[ey][bx] = '╰'

	for x := bx + 1; x < bx+2; x++ {
		b[by][x] = '─'
	}

	for x := bx + 1; x < bx+5; x++ {
		b[ey][x] = '─'
	}

	for y := by + 1; y < ey; y++ {
		b[y][bx] = '│'
	}

	b[by][bx+2] = '['
	b[by][bx+3+utf8.RuneCountInString(f.Text)] = ']'

	for i, r := range []rune(f.Text) {
		b[by][bx+3+i] = r
	}

	f.Span.Render(b, bx+2, by+2)
}

func (f *Frame) Size() (int, int) {
	x, y := f.Span.Size()

	return max(x, utf8.RuneCountInString(f.Text)+2) + 4, y + 4
}

type Text struct {
	Text string
	Hint string
}

func (t *Text) Render(b [][]rune, x, y int) {
	for i, r := range []rune(t.Text) {
		b[y][x+i] = r
	}

	if t.Hint != "" {
		ts := utf8.RuneCountInString(t.Text)

		for i := range ts {
			b[y+1][x+i] = '─'
		}

		b[y+1][x+ts/2] = '┬'
		b[y+2][x+ts/2] = '╰'
		b[y+2][x+ts/2+1] = '─'

		for i, r := range t.Hint {
			b[y+2][x+ts/2+3+i] = r
		}
	}
}

func (t *Text) Size() (int, int) {
	x, y := utf8.RuneCountInString(t.Text), 1

	ts := utf8.RuneCountInString(t.Text)
	hs := utf8.RuneCountInString(t.Hint)

	if t.Hint != "" {
		y += 2
		x += hs + 2 - (ts / 2)
	}

	return x, y
}
