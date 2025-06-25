// Package typ describes language type system, environment
// handling and package definition.
package typ

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

	return t == o
}

var (
	AnyType   = Type{Kind: Any}
	VoidType  = Type{Kind: Void}
	BoolType  = Type{Kind: Bool}
	Sig64Type = Type{Kind: I64}
	Uns64Type = Type{Kind: U64}
	Flt64Type = Type{Kind: F64}
)
