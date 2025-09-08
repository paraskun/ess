package typ

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/paraskun/o2/tty"
)

// Segment is a runtime location of an Object.
type Segment byte

const (
	Abstract Segment = iota // compilation time

	Text    // immutable source code
	Package // immutable package-level data
	Stack   // virtual machine stack
	Static  // reserved per-function data
	Native  // C
)

// Object is a unique typed program entity.
//
// Multiple entities can share the same Type, but
// each can only have one associated Object.
type Object struct {
	Snip *tty.Snippet // place of declaration

	Typ *Type   // inferred type
	Seg Segment // runtime location
	Val any     // compilation time value, maybe nil
}

// Size returns how much bytes object occupies.
func (o *Object) Size() int {
	// For Objects passed by reference
	// we only need to store their address.
	if o.Typ.Kind == REF {
		return 8
	}

	return o.Typ.Size()
}

// Env is an Object storage.
//
// Environment localizes information derived
// from part of source code it has beed attached.
type Env struct {
	Parent *Env

	Sym map[string]*Object // symbol table
}

func NewEnv(p *Env) *Env {
	e := &Env{
		Parent: p,
		Sym:    make(map[string]*Object),
	}

	return e
}

func (e *Env) Insert(name string, sym *Object) (*Object, *Error) {
	if prv, ok := e.Sym[name]; ok {
		prv.Snip.Span.Hint().Text = "previous definiton here"
		prv.Snip.Span.Hint().Attr.Color.Add(color.FgCyan)

		sym.Snip.Span.Hint().Text = "redefined here"
		sym.Snip.Span.Hint().Attr.Color.Add(color.FgRed)

		return prv, &Error{
			Full: fmt.Sprintf("the name \"%s\" is defined multiple times", name),
			Snip: []*tty.Snippet{
				prv.Snip,
				sym.Snip,
			},
		}
	}

	e.Sym[name] = sym

	return sym, nil
}

func (e *Env) Lookup(name string) (*Object, int) {
	env := e
	lvl := 0

	for env != nil {
		if sym, ok := env.Sym[name]; ok {
			return sym, lvl
		}

		lvl += 1
		env = env.Parent
	}

	return nil, lvl
}
