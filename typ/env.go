package typ

import "fmt"

// Object is a typed entity.
// It can be local variable, literal (constant)
// or function.
type Object struct {
	Typ *Type

	// Val contains constant value infered
	// at compile time.
	//
	// Struct -> nil
	// Enum 	-> uint8
	// Bool 	-> bool
	// I64 		-> int64
	// U64 		-> uint64
	// F64 		-> float64
	Val any

	// Off contains offset in bytes inside
	// parenting environment.
	Off int
}

// Size returns how much bytes object occupies.
func (o *Object) Size() int {
	// For objects that passed by reference
	// we have to store only base address.
	if o.Typ.Kind == Ref {
		return 8
	}

	return o.Typ.Size()
}

// Env is an information storage.
//
// Environment localizes information derived
// from part of source code it has beed attached.
//
// For local variables environment determines
// usage scope.
type Env struct {
	Parent *Env

	// Sym is a symbol table for current environment.
	Sym map[string]*Object
}

func NewEnv(p *Env) *Env {
	e := &Env{
		Parent: p,
		Sym:    make(map[string]*Object),
	}

	return e
}

func (e *Env) InsertSym(name string, sym *Object) error {
	if _, ok := e.Sym[name]; ok {
		return fmt.Errorf("\"%s\" already defined in current environment", name)
	}

	e.Sym[name] = sym

	return nil
}

func (e *Env) LookupSym(name string) (*Object, int) {
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
