package typ

import (
	"github.com/paraskun/o2/tty"
	"github.com/paraskun/o2/typ/mod"
)

// Segment is a runtime location of an Object.
type Segment byte

const (
	Abstract Segment = iota // compilation time

	Text    // immutable source code
	Package // immutable package-level data
	Stack   // virtual machine stack
	Static  // reserved per-function data
)

type Location struct {
	File *mod.File
	Span tty.Span
}

// Object is a unique typed program entity.
//
// Multiple entities can share the same Type, but
// each can only have one associated Object.
type Object struct {
	Loc Location // place of declaration
	Typ *Type    // inferred type
	Seg Segment  // runtime location
	Val any      // compilation time value, maybe nil
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

func (e *Env) Insert(name string, sym *Object) (*Object, bool) {
	if prv, ok := e.Sym[name]; ok {
		return prv, false
	}

	e.Sym[name] = sym

	return sym, true
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
