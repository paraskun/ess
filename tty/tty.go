package tty

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

type Attr struct {
	Color color.Color
}

type hintKind byte

const (
	horizontalHint hintKind = iota
	verticalHint
)

type Hint struct {
	Text string
	Attr Attr

	kind   hintKind
	size   int
	offset int
}

type Position struct {
	Row int
	Col int
}

type Span interface {
	more() bool
	draw(w io.Writer)
	hint() *Hint
	size(ind bool) (int, int)
}

type Mono interface {
	Span

	mono()
}

type Tok struct {
	Hint *Hint

	Lit string
	Ind int
}

func (t *Tok) more() bool {
	return false
}

func (t *Tok) draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", t.Ind))
	fmt.Fprintf(w, "%s", t.Lit)
}

func (t *Tok) hint() *Hint {
	if t.Hint != nil {
		t.Hint.typ = hintRow
		t.Hint.len, _ = t.size(false)
		t.Hint.off = t.Ind
	}

	return t.Hint
}

func (t *Tok) size(ind bool) (w int, h int) {
	w += utf8.RuneCountInString(t.Lit)

	if ind {
		w += t.Ind
	}

	return w, 1
}

func (*Tok) mono() {}

type Row struct {
	Hint *Hint

	Sub []Mono
	Ind int
}

func (r *Row) more() bool {
	return false
}

func (r *Row) draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", r.Ind))

	for _, s := range r.Sub {
		s.draw(w)
	}
}

func (r *Row) hint() *Hint {
	off := r.Ind

	for _, s := range r.Sub {
		if h := s.hint(); h != nil {
			h.off += off
			return h
		}

		sw, _ := s.size(true)
		off += sw
	}

	if r.Hint != nil {
		r.Hint.typ = hintRow
		r.Hint.len, _ = r.size(false)
		r.Hint.off = r.Ind
	}

	return r.Hint
}

func (r *Row) size(ind bool) (w int, h int) {
	for _, s := range r.Sub {
		sw, _ := s.size(true)
		w += sw
	}

	if ind {
		w += r.Ind
	}

	return w, 1
}

func (*Row) mono() {}

type Box struct {
	Hint *Hint

	Sub []Span
	Ind int
	Ctl bool

	cur int
	ent int
	ceh *Hint
}

func (b *Box) more() bool {
	return b.ent < len(b.Sub)
}

func (b *Box) draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", b.Ind))

	switch b.cur {
	case 0:
		b.ceh = b.Sub[b.cur].hint()
		b.cur = 1

		if b.ceh != nil && b.ceh.typ == hintBox {
			b.ceh.Atr.Color.Fprintf(w, "┌")
			b.ceh.len -= 1
		}

		b.Sub[b.ent].draw(w)

		if !b.Sub[b.ent].more() {
			if b.ceh != nil {
				b.cur = 2
			} else {
				b.ent += 1
				b.cur = 0
			}
		}
	case 1:
		if b.ceh != nil {
			if b.ceh.len == 0 {
				b.ceh.Atr.Color.Fprintf(w, "├")
			} else {
				b.ceh.Atr.Color.Fprintf(w, "│")
			}

			b.ceh.len -= 1
		}

		b.Sub[b.ent].draw(w)

		if !b.Sub[b.ent].more() {
			if b.ceh != nil {
				b.cur = 2
			} else {
				b.ent += 1
				b.cur = 0
			}
		}
	case 2:
		switch b.ceh.typ {
		case hintRow:
			b.ceh.Atr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.ceh.off))
			b.ceh.Atr.Color.Fprintf(w, "%s┬", strings.Repeat("─", b.ceh.len/2-1))
			b.ceh.Atr.Color.Fprintf(w, "%s", strings.Repeat("─", b.ceh.len/2))

			b.cur = 3

		case hintBox:
			b.ceh.Atr.Color.Fprintf(w, "└──── %s", b.ceh.Msg)

			b.ent += 1
			b.cur = 0
		}
	case 3:
		b.ceh.Atr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.ceh.off+b.ceh.len/2-1))
		b.ceh.Atr.Color.Fprintf(w, "╰─ %s", b.ceh.Msg)

		b.ent += 1
		b.cur = 0
	}

	if b.Ctl {
		fmt.Fprintln(w)
	}
}

func (b *Box) hint() *Hint {
	_, h := b.size(false)

	if b.Hint != nil {
		b.Hint.typ = hintBox
		b.Hint.len = h
	}

	return b.Hint
}

func (b *Box) size(ind bool) (w int, h int) {
	for _, s := range b.Sub {
		sw, sh := s.size(true)

		w = max(w, sw)
		h = h + sh
	}

	if ind {
		w += b.Ind
	}

	return w, h
}

type Frame struct {
	Name string
	Sub  Span

	cur int
}

func (f *Frame) Pos() Pos {
	return Pos{}
}

func (f *Frame) draw(w io.Writer) {
	switch f.cur {
	case 0:
		fmt.Fprintf(w, "╭─[ ")
		color.New(color.FgBlue).Fprintf(w, "%s", f.Name)
		fmt.Fprintf(w, " ]\n")
		f.cur = 1
	case 1:
		fmt.Fprintf(w, "│")
		f.Sub.draw(w)
		fmt.Fprintln(w)

		if !f.Sub.more() {
			f.cur = 2
		}
	case 2:
		fmt.Fprintf(w, "╰────\n")
		f.cur = 3
	}
}

func (f *Frame) more() bool {
	return f.cur < 3
}

func (f *Frame) hint() *Hint {
	return nil
}

func (f *Frame) size(_ bool) (int, int) {
	return 0, 0
}

func Print(s Span) {
	for {
		s.draw(os.Stdout)

		if !s.more() {
			break
		}
	}
}
