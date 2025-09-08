package co2

import (
	"io"

	"github.com/paraskun/o2/typ"
)

type Symbol struct {
	Sym string
	Off int
	Seg typ.Segment
}

func (*Symbol) Write(w io.Writer) int

type Relocation struct {
	Mod string
	Pkg string
	Sym string
	Off int
	Seg typ.Segment
}

func (*Relocation) Write(w io.Writer) int

type Package struct {
	Text []byte
	Data []byte

	Sym []*Symbol
	Rel []*Relocation
	Nat []byte
}

func (*Package) Write(w io.Writer) int
