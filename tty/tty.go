package tty

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

type Indent struct {
	Row int
	Box int
}

type Position struct {
	Row int
	Col int
	Ind Indent
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

type Span interface {
	Position() *Position
	Hint() *Hint

	Draw(w io.Writer)
	More() bool
	Size(i bool) (int, int)

	getHint() *Hint
}

type Group interface {
	Span

	Add(s Span, ir, ib int)

	InsertBeg(m Mono, ir int)
	InsertEnd(m Mono, ir int)
}

type Mono interface {
	Span

	mono()
}

type Tok struct {
	Lit string
	Pos Position

	hint Hint
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
	h := &t.hint

	if h.Text != "" {
		h.size, _ = t.Size(false)
		h.offset = t.Pos.Ind.Row

		return h
	}

	return nil
}

func (*Tok) mono() {}

type Row struct {
	Pos Position
	Sub []Mono

	hint Hint
	size int
}

func (r *Row) Position() *Position {
	return &r.Pos
}

func (r *Row) Hint() *Hint {
	return &r.hint
}

func (r *Row) Add(s Span, ir, ib int) {
	s.Position().Ind.Row = ir
	s.Position().Ind.Box = 0

	r.Sub = append(r.Sub, s.(Mono))
	sw, _ := s.Size(true)
	r.size += sw
}

func (r *Row) InsertBeg(m Mono, ir int) {
	m.Position().Ind.Row = ir
	m.Position().Ind.Box = 0

	sub := make([]Mono, 0, len(r.Sub)+1)
	sub = append(sub, m)
	sub = append(sub, r.Sub...)

	r.Sub = sub
	sw, _ := m.Size(true)
	r.size += sw
}

func (r *Row) InsertEnd(m Mono, ir int) {
	m.Position().Ind.Row = ir
	m.Position().Ind.Box = 0

	r.Sub = append(r.Sub, m)
	sw, _ := m.Size(true)
	r.size += sw
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

	h := &r.hint

	if h.Text != "" {
		h.size = r.size
		h.offset = r.Pos.Ind.Row

		return h
	}

	return nil
}

func (*Row) mono() {}

type Box struct {
	Ctl bool
	Pos Position
	Sub []Span

	hint   Hint
	width  int
	height int

	cur int
	ent int
	ind int
	mon bool
	que *Hint
}

func (b *Box) Position() *Position {
	return &b.Pos
}

func (b *Box) Hint() *Hint {
	return &b.hint
}

func (b *Box) Add(s Span, ir, ib int) {
	s.Position().Ind.Row = ir
	s.Position().Ind.Box = ib

	b.Sub = append(b.Sub, s)
}

func (b *Box) InsertBeg(m Mono, ir int) {
	if len(b.Sub) == 0 {
		b.Add(m, ir, 0)
		return
	}

	f := b.Sub[0]

	switch fb := f.(type) {
	case *Box:
		fb.InsertBeg(m, ir)
	default:
		row := &Row{Pos: Position{Ind: Indent{
			Row: f.Position().Ind.Row,
			Box: f.Position().Ind.Box,
		}}}

		row.Add(m, 0, 0)
		row.Add(f, ir, 0)

		b.Sub[0] = row
	}
}

func (b *Box) InsertEnd(m Mono, ir int) {
	if len(b.Sub) == 0 {
		b.Add(m, ir, 0)
		return
	}

	l := b.Sub[len(b.Sub)-1]

	switch lb := l.(type) {
	case *Box:
		lb.InsertEnd(m, ir)
	default:
		row := &Row{Pos: Position{Ind: Indent{
			Row: l.Position().Ind.Row,
			Box: l.Position().Ind.Box,
		}}}

		row.Add(l, 0, 0)
		row.Add(m, ir, 0)

		b.Sub[len(b.Sub)-1] = row
	}
}

func (b *Box) Draw(w io.Writer) {
	if b.ind != 0 {
		b.ind -= 1

		if b.Ctl {
			fmt.Fprintln(w)
		}

		return
	}

	fmt.Fprintf(w, "%s", strings.Repeat(" ", b.Pos.Ind.Row))

	switch b.cur {
	case 0:
		b.ind = b.Sub[b.ent].Position().Ind.Box
		b.que = b.Sub[b.ent].getHint()
		_, b.mon = b.Sub[b.ent].(Mono)

		if b.ind != 0 {
			b.ind -= 1
			b.cur = 1

			break
		}

		if b.que != nil && !b.mon {
			b.que.Attr.Color.Fprintf(w, "┌")
			b.que.size -= 1
		}

		b.Sub[b.ent].Draw(w)
		b.cur = 2

		if !b.Sub[b.ent].More() {
			if b.que != nil {
				b.cur = 3
			} else {
				b.ent += 1
				b.cur = 0
			}
		}

	case 1:
		if b.que != nil && !b.mon {
			b.que.Attr.Color.Fprintf(w, "┌")
			b.que.size -= 1
		}

		b.Sub[b.ent].Draw(w)
		b.cur = 2

		if !b.Sub[b.ent].More() {
			if b.que != nil {
				b.cur = 3
			} else {
				b.ent += 1
				b.cur = 0
			}
		}

	case 2:
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
				b.cur = 3
			} else {
				b.ent += 1
				b.cur = 0
			}
		}

	case 3:
		switch b.Sub[b.ent].(type) {
		case Mono:
			b.que.Attr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.que.offset))
			b.que.Attr.Color.Fprintf(w, "%s┬", strings.Repeat("─", b.que.size/2))

			if b.que.size%2 == 0 {
				b.que.Attr.Color.Fprintf(w, "%s", strings.Repeat("─", b.que.size/2-1))
			} else {
				b.que.Attr.Color.Fprintf(w, "%s", strings.Repeat("─", b.que.size/2))
			}

			b.cur = 4

		default:
			b.que.Attr.Color.Fprintf(w, "└──── %s", b.que.Text)
			b.ent += 1
			b.cur = 0
		}

	case 4:
		b.que.Attr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.que.offset+b.que.size/2))
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
	h := &b.hint

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
	f.Span.Position().Ind.Row = 2

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

func Print(w io.Writer, s Span) {
	for {
		s.Draw(w)

		if !s.More() {
			break
		}
	}
}
