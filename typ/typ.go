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
		Fields map[string]Field
	}
)

func (t *Type) Size() (r int) {
	switch t.Kind {
	case BOOL:
		return 1
	case I64, U64, F64, FUNC:
		return 8
	case COMP:
		for _, f := range t.Info.(*CompType).Fields {
			r += f.Type.Size()
		}
	}

	return r
}

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
	Ref bool
	Val any
	Off int
}

func (o *Object) Size() int {
	if o.Ref {
		return 8
	}

	return o.Typ.Size()
}

type Env struct {
	*Env

	Sym map[string]*Type
	Imm map[string]*Object
	Obj map[string]*Object
}

func NewEnv(p *Env) *Env {
	return &Env{
		Env: p,
		Sym: make(map[string]*Type),
		Imm: make(map[string]*Object),
		Obj: make(map[string]*Object),
	}
}

func (e *Env) LookupSym(name string) (*Type, int) {
	env := e
	lvl := 0

	for env != nil {
		if sym, ok := env.Sym[name]; ok {
			return sym, lvl
		}

		lvl += 1
		env = env.Env
	}

	return nil, lvl
}

func (e *Env) LookupObj(name string) (*Object, int) {
	env := e
	lvl := 0

	for env != nil {
		if obj, ok := env.Obj[name]; ok {
			return obj, lvl
		}

		lvl += 1
		env = env.Env
	}

	return nil, lvl
}
