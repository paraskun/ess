package tty

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

type Indent struct {
	Row int
	Box int
}

type Position struct {
	Row int // including box indentation
	Col int // including row indentation
	Ind Indent
}

type Span interface {
	Position() *Position
	Hint() *Hint

	Draw(w io.Writer)
	More() bool
	Size(i bool) (int, int)

	getHint() *Hint
}

type Container interface {
	Span

	Add(s Span, row, box int)
}

type Mono interface {
	Span

	mono()
}

type Attr struct {
	Color color.Color
}

type Hint struct {
	Text string
	Attr Attr

	size   int
	offset int
}

type Tok struct {
	hint Hint

	Lit string
	Pos Position
}

func (t *Tok) Position() *Position {
	return &t.Pos
}

func (t *Tok) Hint() *Hint {
	return &t.hint
}

func (t *Tok) Draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", t.Pos.Ind.Row))
	fmt.Fprintf(w, "%s", t.Lit)
}

func (t *Tok) More() bool {
	return false
}

func (t *Tok) Size(i bool) (int, int) {
	w := utf8.RuneCountInString(t.Lit)
	h := 1

	if i {
		w += t.Pos.Ind.Row
		h += t.Pos.Ind.Box
	}

	return w, h
}

func (t *Tok) getHint() *Hint {
	h := t.Hint()

	if h.Text != "" {
		h.size, _ = t.Size(false)
		h.offset = t.Pos.Ind.Row

		return h
	}

	return nil
}

func (*Tok) mono() {}

type Row struct {
	hint Hint

	Pos Position
	Sub []Mono

	size int
}

func (r *Row) Add(s Span, row, box int) {
	s.Position().Ind.Row = row
	s.Position().Ind.Box = box

	rw, _ := r.Size(false)
	sw, _ := s.Size(true)

	s.Position().Col = r.Pos.Col + rw + row
	r.size += sw

	r.Sub = append(r.Sub, s.(Mono))
}

func (r *Row) Hint() *Hint {
	return &r.hint
}

func (r *Row) Position() *Position {
	return &r.Pos
}

func (r *Row) Draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", r.Pos.Ind.Row))

	for _, s := range r.Sub {
		s.Draw(w)
	}
}

func (r *Row) More() bool {
	return false
}

func (r *Row) Size(i bool) (int, int) {
	w := r.size
	h := 1

	if i {
		w += r.Pos.Ind.Row
		h += r.Pos.Ind.Box
	}

	return w, h
}

func (r *Row) getHint() *Hint {
	off := r.Pos.Ind.Row

	for _, s := range r.Sub {
		h := s.getHint()

		if h != nil {
			h.offset += off
			return h
		}

		ssz, _ := s.Size(true)
		off += ssz
	}

	h := r.Hint()

	if h.Text != "" {
		h.size = r.size
		h.offset = r.Pos.Ind.Row

		return h
	}

	return nil
}

func (*Row) mono() {}

type Box struct {
	hint Hint

	Ctl bool
	Pos Position
	Sub []Span

	width  int
	height int

	cur int
	ent int
	ind int
	que *Hint
}

func (b *Box) Add(s Span, row, box int) {
	s.Position().Ind.Row = row
	s.Position().Ind.Box = box

	bw, bh := b.Size(false)
	sw, sh := s.Size(true)

	s.Position().Row = b.Position().Row + bh + box
	s.Position().Col = b.Position().Col + row

	b.width = max(bw, sw)
	b.height += sh

	b.Sub = append(b.Sub, s)
}

func (b *Box) AddLast(m Mono, row int) {
	l := b.Sub[len(b.Sub)-1]

	switch b := l.(type) {
	case *Box:
		b.AddLast(m, row)
	default:
		r := &Row{Pos: Position{Ind: Indent{
			Row: l.Position().Ind.Row,
			Box: l.Position().Ind.Box,
		}}}

		r.Add(b, 0, 0)
		r.Add(m, row, 0)
	}
}

func (b *Box) Hint() *Hint {
	return &b.hint
}

func (b *Box) Position() *Position {
	return &b.Pos
}

func (b *Box) Draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", b.Pos.Ind.Row))

	switch b.cur {
	case 0:
		if b.ind != 0 {
			b.ind -= 1
			break
		}

		_, mono := b.Sub[b.ent].(Mono)
		b.ind = b.Sub[b.ent].Position().Ind.Box

		if b.ind != 0 {
			break
		}

		b.que = b.Sub[b.ent].getHint()
		b.cur = 1

		if b.que != nil && !mono {
			b.que.Attr.Color.Fprintf(w, "┌")
			b.que.size -= 1
		}

		b.Sub[b.ent].Draw(w)

		if !b.Sub[b.ent].More() {
			if b.que != nil {
				b.cur = 2
			} else {
				b.ent += 1
				b.cur = 0
			}
		}
	case 1:
		if b.que != nil {
			if b.que.size == 1 {
				b.que.Attr.Color.Fprintf(w, "├")
			} else {
				b.que.Attr.Color.Fprintf(w, "│")
			}

			b.que.size -= 1
		}

		b.Sub[b.ent].Draw(w)

		if !b.Sub[b.ent].More() {
			if b.que != nil {
				b.cur = 2
			} else {
				b.ent += 1
				b.cur = 0
			}
		}
	case 2:
		switch b.Sub[b.ent].(type) {
		case Mono:
			b.que.Attr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.que.offset))
			b.que.Attr.Color.Fprintf(w, "%s┬", strings.Repeat("─", b.que.size/2-1))
			b.que.Attr.Color.Fprintf(w, "%s", strings.Repeat("─", b.que.size/2))

			b.cur = 3

		default:
			b.que.Attr.Color.Fprintf(w, "└──── %s", b.que.Text)

			b.ent += 1
			b.cur = 0
		}
	case 3:
		b.que.Attr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.que.offset+b.que.size/2-1))
		b.que.Attr.Color.Fprintf(w, "╰─ %s", b.que.Text)

		b.ent += 1
		b.cur = 0
	}

	if b.Ctl {
		fmt.Fprintln(w)
	}
}

func (b *Box) More() bool {
	return b.ent < len(b.Sub)
}

func (b *Box) Size(i bool) (int, int) {
	w := b.width
	h := b.height

	if i {
		w += b.Pos.Ind.Row
		h += b.Pos.Ind.Box
	}

	return w, h
}

func (b *Box) getHint() *Hint {
	h := b.Hint()

	if h.Text != "" {
		h.size = b.height
		h.offset = b.Pos.Ind.Box

		return h
	}

	return nil
}

type Frame struct {
	Name string
	Span Span

	cur int
}

func (f *Frame) Hint() *Hint {
	return nil
}

func (f *Frame) Position() *Position {
	return nil
}

func (f *Frame) Draw(w io.Writer) {
	switch f.cur {
	case 0:
		fmt.Fprintf(w, "╭─[ ")
		color.New(color.FgBlue).Fprintf(w, "%s", f.Name)
		fmt.Fprintf(w, " ]\n")

		f.cur = 1

	case 1:
		fmt.Fprintf(w, "│")
		f.Span.Draw(w)
		fmt.Fprintln(w)

		if !f.Span.More() {
			f.cur = 2
		}

	case 2:
		fmt.Fprintf(w, "╰────\n")

		f.cur = 3
	}
}

func (f *Frame) More() bool {
	return f.cur < 3
}

func (*Frame) Size(_ bool) (int, int) {
	return 0, 0
}

func (*Frame) getHint() *Hint {
	return nil
}

func Print(s Span) {
	for {
		s.Draw(os.Stdout)

		if !s.More() {
			break
		}
	}
}
