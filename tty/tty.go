package tty

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

type indent struct {
	Row int
	Box int
}

type Position struct {
	Row int
	Col int

	ind indent
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
	Reset()
	Render()

	draw(w io.Writer)
	next() int
	size(i bool) (int, int)
	getHint() *Hint
}

type Group interface {
	Span

	Len() int
	Add(s Span, ir, ib int)
	Append(m Mono, i int)
	Prepend(m Mono, i int)
}

type Mono interface {
	Span

	mono()
}

type Tok struct {
	Lit string
	Pos Position

	drawn bool
	hint  Hint
}

func (t *Tok) Position() *Position {
	return &t.Pos
}

func (t *Tok) Hint() *Hint {
	return &t.hint
}

func (t *Tok) Reset() {
	t.drawn = false
}

func (t *Tok) Render() {
	t.Pos.Row += t.Pos.ind.Box
	t.Pos.Col += t.Pos.ind.Row
}

func (t *Tok) draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", t.Pos.ind.Row))
	fmt.Fprintf(w, "%s", t.Lit)

	t.drawn = true
}

func (t *Tok) next() int {
	if t.drawn {
		return -1
	}

	return t.Pos.Row
}

func (t *Tok) size(i bool) (int, int) {
	w := utf8.RuneCountInString(t.Lit)
	h := 1

	if i {
		w += t.Pos.ind.Row
		h += t.Pos.ind.Box
	}

	return w, h
}

func (t *Tok) getHint() *Hint {
	h := &t.hint

	if h.Text != "" {
		h.size, _ = t.size(false)
		h.offset = t.Pos.ind.Row

		return h
	}

	return nil
}

func (*Tok) mono() {}

type Row struct {
	Pos Position

	sub   []Mono
	width int
	hint  Hint
	drawn bool
}

func RowOf(sub ...Mono) *Row {
	row := &Row{}

	for _, s := range sub {
		row.Add(s, 0, 0)
	}

	return row
}

func (r *Row) Position() *Position {
	return &r.Pos
}

func (r *Row) Hint() *Hint {
	return &r.hint
}

func (r *Row) Reset() {
	r.drawn = false

	for _, s := range r.sub {
		s.Reset()
	}
}

func (r *Row) Render() {
	r.Pos.Row += r.Pos.ind.Box
	r.Pos.Col += r.Pos.ind.Row

	off := r.Pos.Col

	for _, s := range r.sub {
		s.Position().Row = r.Pos.Row
		s.Position().Col = off
		s.Render()

		sw, _ := s.size(true)
		off += sw
	}
}

func (r *Row) Len() int {
	return len(r.sub)
}

func (r *Row) Add(s Span, ir, ib int) {
	if len(r.sub) == 0 {
		r.Pos.Row = s.Position().Row
	}

	s.Position().ind.Row = ir
	s.Position().ind.Box = 0

	r.sub = append(r.sub, s.(Mono))
	sw, _ := s.size(true)
	r.width += sw
}

func (r *Row) Append(m Mono, ir int) {
	m.Position().ind.Row = ir
	m.Position().ind.Box = 0

	r.sub = append(r.sub, m)
	sw, _ := m.size(true)
	r.width += sw
}

func (r *Row) Prepend(m Mono, ir int) {
	m.Position().ind.Row = ir
	m.Position().ind.Box = 0

	sub := make([]Mono, 0, len(r.sub)+1)
	sub = append(sub, m)
	sub = append(sub, r.sub...)

	r.sub = sub
	sw, _ := m.size(true)
	r.width += sw
}

func (r *Row) draw(w io.Writer) {
	fmt.Fprintf(w, "%s", strings.Repeat(" ", r.Pos.ind.Row))

	for _, s := range r.sub {
		s.draw(w)
	}

	r.drawn = true
}

func (r *Row) next() int {
	if r.drawn {
		return -1
	}

	return r.Pos.Row
}

func (r *Row) size(i bool) (int, int) {
	w := r.width
	h := 1

	if i {
		w += r.Pos.ind.Row
		h += r.Pos.ind.Box
	}

	return w, h
}

func (r *Row) getHint() *Hint {
	off := r.Pos.ind.Row

	for _, s := range r.sub {
		h := s.getHint()

		if h != nil {
			h.offset += off
			return h
		}

		ssz, _ := s.size(true)
		off += ssz
	}

	h := &r.hint

	if h.Text != "" {
		h.size = r.width
		h.offset = r.Pos.ind.Row

		return h
	}

	return nil
}

func (*Row) mono() {}

type Box struct {
	Ctl bool
	Pos Position

	sub     []Span
	hint    Hint
	width   int
	height  int
	state   int
	entry   int
	indent  int
	mono    bool
	pending *Hint
}

func BoxOf(sub ...Span) *Box {
	box := &Box{}

	for _, s := range sub {
		box.Add(s, 0, 0)
	}

	return box
}

func (b *Box) Position() *Position {
	return &b.Pos
}

func (b *Box) Hint() *Hint {
	return &b.hint
}

func (b *Box) Reset() {
	b.state = 0
	b.entry = 0
	b.pending = nil
	b.mono = false
	b.indent = 0

	for _, s := range b.sub {
		s.Reset()
	}
}

func (b *Box) Render() {
	b.Pos.Row += b.Pos.ind.Box
	b.Pos.Col += b.Pos.ind.Row

	off := b.Pos.Row

	for _, s := range b.sub {
		s.Position().Row = off
		s.Position().Col = b.Pos.Col
		s.Render()

		_, sh := s.size(true)
		off += sh
	}
}

func (b *Box) Len() int {
	return len(b.sub)
}

func (b *Box) Add(s Span, ir, ib int) {
	s.Position().ind.Row = ir
	s.Position().ind.Box = ib

	b.sub = append(b.sub, s)
	_, sh := s.size(true)
	b.height += sh
}

func (b *Box) Append(m Mono, ir int) {
	if len(b.sub) == 0 {
		b.Add(m, ir, 0)
		return
	}

	l := b.sub[len(b.sub)-1]

	switch lb := l.(type) {
	case *Box:
		lb.Append(m, ir)
	default:
		row := &Row{Pos: Position{ind: l.Position().ind}}

		row.Add(l, 0, 0)
		row.Add(m, ir, 0)

		b.sub[len(b.sub)-1] = row
	}
}

func (b *Box) Prepend(m Mono, ir int) {
	if len(b.sub) == 0 {
		b.Add(m, ir, 0)
		return
	}

	f := b.sub[0]

	switch fb := f.(type) {
	case *Box:
		fb.Prepend(m, ir)
	default:
		row := &Row{Pos: Position{ind: f.Position().ind}}

		row.Add(m, 0, 0)
		row.Add(f, ir, 0)

		b.sub[0] = row
	}
}

func (b *Box) draw(w io.Writer) {
	if b.indent != 0 {
		b.indent -= 1

		if b.Ctl {
			fmt.Fprintln(w)
		}

		return
	}

	fmt.Fprintf(w, "%s", strings.Repeat(" ", b.Pos.ind.Row))

	switch b.state {
	case 0:
		b.indent = b.sub[b.entry].Position().ind.Box
		b.pending = b.sub[b.entry].getHint()
		_, b.mono = b.sub[b.entry].(Mono)

		if b.indent != 0 {
			b.indent -= 1
			b.state = 1

			break
		}

		if b.pending != nil && !b.mono {
			b.pending.Attr.Color.Fprintf(w, "┌ ")
			b.pending.size -= 1
		}

		b.state = 2

		if b.sub[b.entry].next() != -1 {
			b.sub[b.entry].draw(w)

			if b.sub[b.entry].next() == -1 {
				if b.pending != nil {
					b.state = 3
				} else {
					b.entry += 1
					b.state = 0
				}
			}
		} else {
			if b.pending != nil {
				b.state = 3
			} else {
				b.entry += 1
				b.state = 0
			}
		}

	case 1:
		if b.pending != nil && !b.mono {
			b.pending.Attr.Color.Fprintf(w, "┌ ")
			b.pending.size -= 1
		}

		b.state = 2

		if b.sub[b.entry].next() != -1 {
			b.sub[b.entry].draw(w)

			if b.sub[b.entry].next() == -1 {
				if b.pending != nil {
					b.state = 3
				} else {
					b.entry += 1
					b.state = 0
				}
			}
		} else {
			if b.pending != nil {
				b.state = 3
			} else {
				b.entry += 1
				b.state = 0
			}
		}

	case 2:
		if b.pending != nil {
			b.pending.Attr.Color.Fprintf(w, "│ ")
		}

		b.sub[b.entry].draw(w)

		if b.sub[b.entry].next() == -1 {
			if b.pending != nil {
				b.state = 3
			} else {
				b.entry += 1
				b.state = 0
			}
		}

	case 3:
		switch b.sub[b.entry].(type) {
		case Mono:
			b.pending.Attr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.pending.offset))
			b.pending.Attr.Color.Fprintf(w, "%s┬", strings.Repeat("─", b.pending.size/2))

			if b.pending.size%2 == 0 {
				b.pending.Attr.Color.Fprintf(w, "%s", strings.Repeat("─", b.pending.size/2-1))
			} else {
				b.pending.Attr.Color.Fprintf(w, "%s", strings.Repeat("─", b.pending.size/2))
			}

			b.state = 4

		default:
			b.pending.Attr.Color.Fprintf(w, "└──── %s", b.pending.Text)
			b.pending.Text = ""
			b.entry += 1
			b.state = 0
		}

	case 4:
		b.pending.Attr.Color.Fprintf(w, "%s", strings.Repeat(" ", b.pending.offset+b.pending.size/2))
		b.pending.Attr.Color.Fprintf(w, "╰─ %s", b.pending.Text)
		b.pending.Text = ""
		b.entry += 1
		b.state = 0
	}

	if b.Ctl {
		fmt.Fprintln(w)
	}
}

func (b *Box) next() int {
	if b.entry < len(b.sub) {
		row := b.sub[b.entry].next()

		if row == -1 {
			return 0
		}

		return row
	}

	return -1
}

func (b *Box) size(i bool) (int, int) {
	w := b.width
	h := b.height

	if i {
		w += b.Pos.ind.Row
		h += b.Pos.ind.Box
	}

	return w, h
}

func (b *Box) getHint() *Hint {
	h := &b.hint

	if h.Text != "" {
		h.size = b.height
		h.offset = b.Pos.ind.Box

		return h
	}

	return nil
}

type Frame struct {
	Name string
	Span Span
	Line bool

	state int
}

func (f *Frame) Hint() *Hint {
	return nil
}

func (f *Frame) Position() *Position {
	return nil
}

func (f *Frame) Reset() {
	f.state = 0
	f.Span.Reset()
}

func (f *Frame) Render() {
	f.Span.Position().Row = 0
	f.Span.Position().Col = 0
	f.Span.Render()
}

func (f *Frame) draw(w io.Writer) {
	f.Span.Position().ind.Row = 2

	switch f.state {
	case 0:
		fmt.Fprintf(w, "   ╭─[ ")
		color.New(color.FgBlue).Fprintf(w, "%s", f.Name)
		fmt.Fprintf(w, " ]\n")
		f.state = 1

	case 1:
		if row := f.Span.next(); row != -1 {
			if row == 0 {
				fmt.Fprintf(w, "   │")
			} else {
				fmt.Fprintf(w, "%2d │", row)
			}

			f.Span.draw(w)
			fmt.Fprintln(w)
		} else {
			f.state = 2
		}

	case 2:
		fmt.Fprintf(w, "   ╰────\n")
		f.state = 3
	}
}

func (f *Frame) next() int {
	if f.state < 3 {
		return 0
	}

	return -1
}

func (*Frame) size(_ bool) (int, int) {
	return 0, 0
}

func (*Frame) getHint() *Hint {
	return nil
}

func Print(w io.Writer, s Span) {
	for s.next() != -1 {
		s.draw(w)
	}
}
