package typ

import "fmt"

type Kind byte

const (
	FUNC Kind = 0b00000000
	COMP      = 0b00100000
	BOOL      = 0b01000000
	I64       = 0b01100000
	U64       = 0b10000000
	F64       = 0b10100000
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

		Typ *Type
		Off int
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
			r += f.Typ.Size()
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
		if !t.Arg[i].Typ.Equal(o.Arg[i].Typ) {
			return false
		}
	}

	if len(t.Ret) != len(o.Ret) {
		return false
	}

	for i := range t.Ret {
		if !t.Ret[i].Typ.Equal(o.Ret[i].Typ) {
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
	Loc bool
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
	Env *Env
	Top *Env

	Sym map[string]*Type
	Imm map[string]*Object
	Obj map[string]*Object
}

func NewEnv(p *Env) *Env {
	return &Env{
		Env: p,
		Top: p.Top,
		Sym: make(map[string]*Type),
		Imm: make(map[string]*Object),
		Obj: make(map[string]*Object),
	}
}

func (e *Env) InsertSym(name string, sym *Type) error {
	if _, ok := e.Sym[name]; ok {
		return fmt.Errorf("symbol name collision")
	}

	e.Sym[name] = sym

	return nil
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

func (e *Env) InsertImm(lit string, imm *Object) {
	if _, ok := e.Top.Imm[lit]; ok {
		return
	}

	e.Top.Imm[lit] = imm
}

func (e *Env) LookupImm(lit string) *Object {
	return e.Top.Imm[lit]
}

func (e *Env) InsertObj(name string, obj *Object) error {
	if _, ok := e.Obj[name]; ok {
		return fmt.Errorf("object name collision")
	}

	e.Obj[name] = obj

	return nil
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
