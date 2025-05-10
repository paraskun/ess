package typ

type Kind uint8

const (
	FUNC Kind = iota
	COMP
	BOOL

	I64
	U64
	F64
)

type (
	Type struct {
		Kind Kind

		// Composite type info
		// 	- func -> *FuncType
		//	- comp -> *CompType
		Info any
	}

	Field struct {
		Name string
		Type *Type
	}

	FuncType struct {
		Arg []Field
		Ret []Field
	}

	CompType struct {
		Fields []Field
	}
)

func (t *Type) Equal(o *Type) bool {
	if t.Kind != o.Kind {
		return false
	}

	switch t.Kind {
	case FUNC:
		return t.Info.(*FuncType).Equal(o.Info.(*FuncType))
	case COMP:
		return t == o
	}

	return true
}

func (t *FuncType) Equal(o *FuncType) bool {
	if len(t.Arg) != len(o.Arg) {
		return false
	}

	for i := range t.Arg {
		if !t.Arg[i].Type.Equal(o.Arg[i].Type) {
			return false
		}
	}

	if len(t.Ret) != len(t.Ret) {
		return false
	}

	for i := range t.Ret {
		if !t.Ret[i].Type.Equal(o.Ret[i].Type) {
			return false
		}
	}

	return true

}

var (
	BoolType = Type{
		Kind: BOOL,
	}

	Sig64Type = Type{
		Kind: I64,
	}

	Uns64Type = Type{
		Kind: U64,
	}

	Flt64Type = Type{
		Kind: F64,
	}
)

type Object struct {
	Typ *Type
	Env *Env
	Val any
}

type Env struct {
	*Env

	Sym map[string]*Type
	Obj map[string]*Object
}

func (*Env) LookupSym(name string) (*Type, uint)
func (*Env) LookupObj(name string) (*Object, uint)
