package tty

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

type Pos struct {
	Row int
	Col int
	Len int
}

type Hint struct {
	Lit string
	Col *color.Color

	pos Pos
}

type Span interface {
	Draw(w io.Writer)
	More() bool

	GetHint() *Hint
}

type Mono interface {
	Span

	size(i bool) int
}

type Tok struct {
	*Hint

	Lit string
	Pos Pos
	Ind int
}

func (t *Tok) Draw(w io.Writer) {
	if t.Ind != 0 {
		fmt.Fprintf(w, "%*c", t.Ind, ' ')
	}

	fmt.Fprintf(w, "%s", t.Lit)
}

func (*Tok) More() bool {
	return false
}

func (t *Tok) GetHint() *Hint {
	if t.Hint != nil {
		t.Hint.pos = Pos{
			Col: t.Ind,
			Len: t.size(false),
		}
	}

	return t.Hint
}

func (t *Tok) size(i bool) (s int) {
	if t.Pos.Len == 0 {
		t.Pos.Len = utf8.RuneCountInString(t.Lit)
	}

	s += t.Pos.Len

	if i {
		s += t.Ind
	}

	return s
}

type Row struct {
	*Hint

	Sub []Mono
	Ind int

	csz int
}

func (r *Row) Draw(w io.Writer) {
	if r.Ind != 0 {
		fmt.Fprintf(w, "%*c", r.Ind, ' ')
	}

	for _, s := range r.Sub {
		s.Draw(w)
	}
}

func (*Row) More() bool {
	return false
}

func (r *Row) GetHint() *Hint {
	off := r.Ind

	for _, s := range r.Sub {
		if h := s.GetHint(); h != nil {
			h.pos.Col += off
			return h
		}

		off += s.size(true)
	}

	if r.Hint != nil {
		r.Hint.pos = Pos{
			Col: r.Ind,
			Len: r.size(false),
		}
	}

	return r.Hint
}

func (r *Row) size(i bool) (s int) {
	if r.csz == 0 {
		for _, s := range r.Sub {
			r.csz += s.size(true)
		}
	}

	s += r.csz

	if i {
		s += r.Ind
	}

	return s
}

type Box struct {
	*Hint

	Sub []Span
	Ind int
	Ctl bool

	cur int
	ent int
	ceh *Hint
}

func (b *Box) Draw(w io.Writer) {
	if b.Ind != 0 {
		fmt.Fprintf(w, "%*c", b.Ind, ' ')
	}

	switch b.cur {
	case 0: // new
		b.ceh = b.Sub[b.ent].GetHint()
		b.cur = 1

		if b.ceh != nil && b.ceh.pos.Row == 1 {
			b.ceh.Col.Fprintf(w, "┌")
		}

		b.Sub[b.ent].Draw(w)

		if !b.Sub[b.ent].More() {
			if b.ceh != nil {
				b.cur = 2
			} else {
				b.ent += 1
				b.cur = 0
			}
		}
	case 1: // continue
		if b.ceh != nil {
			b.ceh.Col.Fprintf(w, "│")
		}

		b.Sub[b.ent].Draw(w)

		if !b.Sub[b.ent].More() {
			if b.ceh != nil {
				b.cur = 2
			} else {
				b.ent += 1
				b.cur = 0
			}
		}
	case 2: // hint-1
		switch b.ceh.pos.Row {
		case 0:
			b.ceh.Col.Fprintf(w, "%s", strings.Repeat(" ", b.ceh.pos.Col))
			b.ceh.Col.Fprintf(w, "%s┬", strings.Repeat("─", b.ceh.pos.Len/2-1))
			b.ceh.Col.Fprintf(w, "%s", strings.Repeat("─", b.ceh.pos.Len/2))

			b.cur = 3

		case 1:
			b.ceh.Col.Fprintf(w, "└──── %s", b.ceh.Lit)

			b.ent += 1
			b.cur = 0
		}

	case 3: // hint-2
		b.ceh.Col.Fprintf(w, "%s", strings.Repeat(" ", b.ceh.pos.Col+b.ceh.pos.Len/2-1))
		b.ceh.Col.Fprintf(w, "╰─ %s", b.ceh.Lit)

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

func (b *Box) GetHint() *Hint {
	if b.Hint != nil {
		b.Hint.pos = Pos{
			Row: 1,
		}
	}

	return b.Hint
}

type Frame struct {
	Name string
	Sub  Span

	cur int
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
		f.Sub.Draw(w)
		fmt.Fprintln(w)

		if !f.Sub.More() {
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

func (f *Frame) GetHint() *Hint {
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
