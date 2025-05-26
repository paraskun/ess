package typ

import "fmt"

type Kind byte

const (
	ANY  Kind = 0b00000000
	FUNC      = 0b00100000
	COMP      = 0b01000000
	BOOL      = 0b01100000
	I64       = 0b10000000
	U64       = 0b10100000
	F64       = 0b11000000
	STR       = 0b11100000
)

type (
	Type struct {
		Kind Kind

		// Composite type info.
		//
		// 	- FUNC -> *FuncType
		//	- COMP -> *CompType
		Info any
	}

	Field struct {
		Name string // maybe empty

		// Typ specifies the type.
		Typ *Type

		// Off is an offset in bytes within
		// the surrounding object.
		Off int
	}

	FuncInfo struct {
		Arg []*Field
		Ret *Field

		Off int
		Nat bool
	}

	CompInfo struct {
		Fields map[string]*Field
	}
)

func (t *Type) Size() (r int) {
	switch t.Kind {
	case BOOL:
		return 1
	case I64, U64, F64:
		return 8
	case FUNC:
		return 4
	case COMP:
		for _, f := range t.Info.(*CompInfo).Fields {
			r += f.Typ.Size()
		}
	}

	return r
}

func (t *Type) Equal(o *Type) bool {
	if o == nil {
		return false
	}

	if t.Kind != o.Kind {
		return false
	}

	if t.Kind == ANY {
		return false
	}

	switch t.Kind {
	case FUNC:
		return t.Info.(*FuncInfo).Equal(o.Info.(*FuncInfo))
	case COMP:
		return t == o
	}

	return true
}

func (t *FuncInfo) Equal(o *FuncInfo) bool {
	if len(t.Arg) != len(o.Arg) {
		return false
	}

	for i := range t.Arg {
		if !t.Arg[i].Typ.Equal(o.Arg[i].Typ) {
			return false
		}
	}

	if (t.Ret == nil && o.Ret != nil) || (t.Ret != nil && o.Ret == nil) {
		return false
	}

	if !t.Ret.Typ.Equal(o.Ret.Typ) {
		return false
	}

	return true
}

var (
	AnyType = Type{
		Kind: ANY,
	}

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

	LogType = Type{
		Kind: FUNC,
		Info: &FuncInfo{
			Arg: []*Field{
				{
					Typ: &AnyType,
					Off: 0,
				},
			},
			Ret: nil,
			Nat: true,
			Off: -1,
		},
	}
)

// Object is a global variable (function),
// local variable or (local) literal.
type Object struct {
	// Typ specifies objects type.
	Typ *Type

	// Ref indicates whether object
	// stored directly or by
	// reference pointer.
	Ref bool

	// Val constains constant values
	// infered at compile time.
	//
	//	- COMP -> nil
	//  - BOOL -> bool
	//  - I64 -> int64
	//  - U64 -> uint64
	//  - F64 -> float64
	Val any

	// Off contains offset in bytes inside
	// parenting environment.
	Off int
}

func (o *Object) Size() int {
	if o.Ref {
		return 8
	}

	return o.Typ.Size()
}

// Env is an environment associate with a block.
type Env struct {
	Parent *Env

	// Root is a reference to nearest
	// parenting environment associated
	// with a function.
	Root *Env

	// ImmSz indicates imm block size in bytes.
	ImmSz int

	// Sym table contains type definitions.
	Sym map[string]*Type

	// Imm table contains immediate objects
	// (literals) used in associated function.
	Imm map[string]*Object

	// Obj table contains local variables
	// defined in associated block.
	Obj map[string]*Object
}

func NewEnv(p *Env) *Env {
	e := &Env{
		Parent: p,
		Sym:    make(map[string]*Type),
		Imm:    make(map[string]*Object),
		Obj:    make(map[string]*Object),
	}

	if p != nil {
		e.Root = p.Root
	}

	return e
}

func (e *Env) InsertSym(name string, sym *Type) error {
	if _, ok := e.Sym[name]; ok {
		return fmt.Errorf("\"%s\" already defined in current environment", name)
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
		env = env.Parent
	}

	return nil, lvl
}

func (e *Env) InsertImm(lit string, imm *Object) {
	if _, ok := e.Root.Imm[lit]; ok {
		return
	}

	e.Root.Imm[lit] = imm
	e.Root.ImmSz += imm.Size()
}

func (e *Env) LookupImm(lit string) *Object {
	return e.Root.Imm[lit]
}

func (e *Env) InsertObj(name string, obj *Object) error {
	if _, ok := e.Obj[name]; ok {
		return fmt.Errorf("\"%s\" already defined in current environment", name)
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

		if env == e.Root {
			break
		}

		lvl += 1
		env = env.Parent
	}

	return nil, lvl
}
