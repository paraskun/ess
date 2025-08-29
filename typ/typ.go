// Package typ describes language type system.
package typ

type Kind byte

const (
	// Basic

	// Basic type must occupy three bits in
	// order to fit into opcode representation.
	// It means, we can have at most eight
	// basic data types.

	BOOL Kind = iota
	I64
	U64
	F64
	STR

	// Internal

	// This types are internal because they are
	// taking place only during compilation.

	// Type Any can be used only as a type of
	// an argument to native functions.

	PKG
	DEC
	REF
	ANY

	FUNC
	ENUM
	ARRAY
	STRUCT
)

type (
	// Type is a collection of invariant properties
	// associated with each Object.
	Type struct {
		Kind Kind

		// PKG 		-> *mod.Package
		// DEC		-> *Type
		// REF 		-> *Type
		// FUNC 	-> *Func
		// ENUM 	-> *Enum
		// ARRAY 	-> *Array
		// STRUCT 	-> *Struct
		Extra any
	}

	// Field is a named or unnamed member
	// of some logical group.
	Field struct {
		Name string // name, maybe empty
		Typ  *Type  // type
	}

	Func struct {
		Arg []*Field
		Ret *Field
		Dec any // *ast.FuncDecl
	}

	Enum struct {
		Mem map[string]uint8
		Dec any // *ast.EnumDecl
	}

	Array struct {
		Typ *Type // array member type
		Cap int
	}

	Struct struct {
		Mem map[string]*Field
		Dec any // *ast.StructDecl
	}
)

// Size returns how much bytes occupies Object
// of that type bypassing all references.
func (t *Type) Size() (r int) {
	switch t.Kind {
	case BOOL:
		return 1
	case I64, U64, F64:
		return 8
	case REF:
		return t.Extra.(*Type).Size()
	case FUNC:
		return 4
	case ENUM:
		return 1
	case STRUCT:
		for _, f := range t.Extra.(*Struct).Mem {
			r += f.Typ.Size()
		}
	}

	return r
}

// Equal checks if two types are compatible to each other.
func (t *Type) Equal(o *Type) bool {
	if t == nil || o == nil {
		return false
	}

	return t == o
}

// Basic data types

var (
	AnyType  = Type{Kind: ANY}
	I64Type  = Type{Kind: I64}
	U64Type  = Type{Kind: U64}
	F64Type  = Type{Kind: F64}
	StrType  = Type{Kind: STR}
	BoolType = Type{Kind: BOOL}
)
