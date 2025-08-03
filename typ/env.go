package typ

import (
	"fmt"
)

type Segment byte

const (
	Abstract Segment = iota

	Text
	PDat
	SDat
	DDat
)

// Object is a unique typed program entity.
//
// Multiple entities can share the same Type, but
// each can only have one attached Object.
type Object struct {
	Typ *Type
	Seg Segment
	Off uint32
	Val any
}

// Size returns how much bytes object occupies.
func (o *Object) Size() int {
	// For objects that passed by reference
	// we have to store only base address.
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

func New(p *Env) *Env {
	e := &Env{
		Parent: p,
		Sym:    make(map[string]*Object),
	}

	return e
}

func (e *Env) Insert(name string, sym *Object) error {
	if _, ok := e.Sym[name]; ok {
		return fmt.Errorf("\"%s\" already defined in the current environment", name)
	}

	e.Sym[name] = sym

	return nil
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
