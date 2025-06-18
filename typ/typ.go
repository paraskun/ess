// Package typ describes language type system.
package typ

import "fmt"

type Kind byte

const (
	// Basic

	// Basic type must occupy three bits in
	// order to fit into opcode representation.
	// It means, we can have at most eight
	// basic data types.

	Bool Kind = iota
	I64
	U64
	F64
	Str

	// Internal

	// This types are internal because they are
	// taking place only during compilation.

	// *Any* type can be used only as a type of
	// an argument to native functions.

	Any
	Ref
	Func
	Enum
	Struct
	Void
)

type (
	// Type is a collection of invariant properties
	// associated with each object.
	Type struct {
		Kind Kind

		// Complex type info.
		//
		// Ref 		-> *Type
		// Func 	-> *FuncInfo
		// Enum 	-> *EnumInfo
		// Struct -> *StructInfo
		Info any
	}

	// Field is a named or unnamed member
	// of some logical group.
	Field struct {
		Name string // maybe empty
		Typ  *Type  // field type

		// Off is an offset in bytes within
		// logical group.
		Off int

		// Field index (for structures)
		Idx uint8
	}

	FuncInfo struct {
		Arg []*Field // arguments
		Ret *Field   // return
	}

	EnumInfo struct {
		Members map[string]uint8
	}

	StructInfo struct {
		Fields map[string]*Field
	}
)

// Size returns how much bytes occupies object
// of that type bypassing all references.
func (t *Type) Size() (r int) {
	switch t.Kind {
	case Bool:
		return 1
	case I64, U64, F64:
		return 8
	case Ref:
		return t.Info.(*Type).Size()
	case Func:
		return 4
	case Enum:
		return 1
	case Struct:
		for _, f := range t.Info.(*StructInfo).Fields {
			r += f.Typ.Size()
		}
	case Void:
		return 0
	}

	return r
}

// Equal checks if two types are compatible
// to each other.
func (t *Type) Equal(o *Type) bool {
	if t == nil || o == nil {
		return false
	}

	if t.Kind != o.Kind {
		return false
	}

	if t.Kind == Any {
		return false
	}

	switch t.Kind {
	case Func:
		return t.Info.(*FuncInfo).Equal(o.Info.(*FuncInfo))
	case Struct:
		return t == o
	}

	return true
}

// Equal checks whether two functions compatible
// to each other by their signatures.
func (t *FuncInfo) Equal(o *FuncInfo) bool {
	if len(t.Arg) != len(o.Arg) {
		return false
	}

	for i := range t.Arg {
		if !t.Arg[i].Typ.Equal(o.Arg[i].Typ) {
			return false
		}
	}

	if !t.Ret.Typ.Equal(o.Ret.Typ) {
		return false
	}

	return true
}

var (
	AnyType   = Type{Kind: Any}
	VoidType  = Type{Kind: Void}
	BoolType  = Type{Kind: Bool}
	Sig64Type = Type{Kind: I64}
	Uns64Type = Type{Kind: U64}
	Flt64Type = Type{Kind: F64}
)

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
