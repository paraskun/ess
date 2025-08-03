package ast

import (
	"strconv"

	"github.com/paraskun/o2/lex"
	"github.com/paraskun/o2/typ"
	"github.com/paraskun/o2/typ/mod"
)

type typer struct {
	pkg *mod.Package
	env *typ.Env
	fun *typ.Func
}

func Typeset(pkg *mod.Package) {
	pkg.Env = typ.New(nil)
	typ := typer{pkg: pkg, env: pkg.Env}

	for _, src := range pkg.XSrc {
		src.Env = pkg.Env

		for _, dec := range src.Dec.(*File).Dec {
			dec.Accept(&typ)
		}
	}
}

func (t *typer) VisitDecl(u Decl) {
	switch d := u.(type) {
	case *UseDecl:
		t.visitUseDecl(d)
	case *VarDecl:
		t.visitVarDecl(d)
	case *FuncDecl:
		t.visitFuncDecl(d)
	case *StructDecl:
		t.visitStructDecl(d)
	case *EnumDecl:
		t.visitEnumDecl(d)
	}
}

func (t *typer) visitUseDecl(d *UseDecl) {
	if d.Pkg == nil {
		return
	}

	// Ignore an error, if multiple files using same package.
	_ = t.typ.Insert(d.Pkg.Name, &typ.Object{
		Cap: typ.C_ADR,
		Typ: &typ.Type{
			Kind:  typ.PKG,
			Extra: d.Pkg,
		},
	})

	Typeset(d.Pkg)
}

func (t *typer) visitVarDecl(d *VarDecl) {
	d.Ini.Accept(t)

	if err := t.typ.Insert(d.Idf.Lit, &typ.Object{
		Cap: typ.C_ADR,
		Typ: d.Ini.Type(),
		Val: d.Ini.(*ImmExpr).Obj.Val, // TODO: expression evaluation
	}); err != nil {
		panic(err)
	}
}

func (t *typer) visitFuncDecl(d *FuncDecl) {
	d.Env = typ.New(t.env)
	d.Obj = &typ.Object{Typ: t.funcSpec(d)}

	if err := t.typ.Insert(d.Idf.Lit, d.Obj); err != nil {
		panic(err)
	}

	t.env = d.Env
	t.fun = d.Obj.Typ.Extra.(*typ.Func)

	d.Body.Accept(t)

	t.fun = nil
	t.env = d.Env.Parent
}

func (t *typer) funcSpec(d *FuncDecl) *typ.Type {
	e := &typ.Func{Dec: d}
	r := &typ.Type{Kind: typ.FUNC, Extra: e}

	for _, arg := range d.Arg {
		t.typeSpec(arg.Typ)

		if arg.Typ.Typ.Kind == typ.STRUCT {
			arg.Typ.Typ = &typ.Type{
				Kind:  typ.REF,
				Extra: arg.Typ.Typ,
			}
		}

		obj := &typ.Object{
			Cap: typ.C_ADR | typ.C_MOD,
			Typ: arg.Typ.Typ,
		}

		if err := t.typ.Insert(arg.Tok.Lit, obj); err != nil {
			panic(err)
		}

		e.Arg = append(e.Arg, &typ.Field{
			Name: arg.Tok.Lit,
			Typ:  obj.Typ,
		})
	}

	for _, ret := range d.Ret {
		t.typeSpec(ret.Typ)

		if ret.Typ.Typ.Kind == typ.STRUCT {
			ret.Typ.Typ = &typ.Type{
				Kind:  typ.REF,
				Extra: ret.Typ.Typ,
			}
		}

		e.Ret = append(e.Ret, &typ.Field{
			Typ: ret.Typ.Typ,
		})
	}

	return r
}

func (t *typer) typeSpec(s *TypeSpec) {
	switch s.Tok.TokenType {
	case lex.BOOL:
		s.Typ = &typ.BoolType
	case lex.I64:
		s.Typ = &typ.Sig64Type
	case lex.U64:
		s.Typ = &typ.Uns64Type
	case lex.F64:
		s.Typ = &typ.Flt64Type
	case lex.IDF:
		sym, _ := t.typ.Lookup(s.Tok.Lit)

		if sym == nil {
			panic("undeclared type")
		}

		if sym.Typ.Kind != typ.ENUM && sym.Typ.Kind != typ.STRUCT {
			panic("type specification expected")
		}

		s.Typ = sym.Typ
	}
}

func (t *typer) visitStructDecl(d *StructDecl) {
	d.Obj = &typ.Object{Typ: t.structSpec(d)}

	if err := t.typ.Insert(d.Idf.Lit, d.Obj); err != nil {
		panic(err)
	}
}

func (t *typer) structSpec(d *StructDecl) *typ.Type {
	e := &typ.Struct{Mem: make(map[string]*typ.Field), Dec: d}
	r := &typ.Type{Kind: typ.STRUCT, Extra: e}

	for idx, mem := range d.Mem {
		t.typeSpec(mem.Typ)

		if _, ok := e.Mem[mem.Tok.Lit]; ok {
			panic("duplicated structure member")
		}

		e.Mem[mem.Tok.Lit] = &typ.Field{
			Name: mem.Tok.Lit,
			Typ:  mem.Typ.Typ,
			Idx:  idx,
		}
	}

	return r
}

func (t *typer) visitEnumDecl(d *EnumDecl) {
	d.Obj = &typ.Object{Typ: t.enumSpec(d)}

	if err := t.typ.Insert(d.Idf.Lit, d.Obj); err != nil {
		panic(err)
	}
}

func (t *typer) enumSpec(d *EnumDecl) *typ.Type {
	e := &typ.Enum{Mem: make(map[string]uint8), Dec: d}
	r := &typ.Type{Kind: typ.ENUM, Extra: e}

	for idx, mem := range d.Mem {
		if _, ok := e.Mem[mem.Lit]; ok {
			panic("duplicated enum member")
		}

		e.Mem[mem.Lit] = uint8(idx)
	}

	return r
}

func (t *typer) VisitStmt(u Stmt) {
	switch s := u.(type) {
	case *ReturnStmt:
		t.visitReturnStmt(s)
	case *VarStmt:
		t.visitVarStmt(s)
	case *LetStmt:
		t.visitLetStmt(s)
	case *AssignStmt:
		t.visitAssignStmt(s)
	case *BlockStmt:
		s.Env = typ.New(t.env)
		t.env = s.Env

		for _, o := range s.Body {
			o.Accept(t)
		}

		t.env = s.Env.Parent
	case *LoopStmt:
		s.Con.Accept(t)
		s.Rep.Accept(t)
	case *CondStmt:
		s.Con.Accept(t)
		s.Pos.Accept(t)

		if s.Neg != nil {
			s.Neg.Accept(t)
		}
	case *CallStmt:
		s.CallExpr.Accept(t)

		if s.Type() != nil {
			panic("missed return")
		}
	}
}

func (t *typer) visitReturnStmt(r *ReturnStmt) {
	if len(t.fun.Ret) != len(r.Ret) {
		panic("function doesn't suppose to return anything")
	}

	for i, r := range r.Ret {
		r.Accept(t)

		if r.Caps()&typ.C_ADR == 0 {
			panic("could not return non-addressable object")
		}

		n := r.Type()

		if n.Kind == typ.STRUCT || n.Kind == typ.ARRAY {
			p := &typ.Type{Kind: n.Kind, Extra: n.Extra}

			n.Kind = typ.REF
			n.Extra = p
		}

		if !t.fun.Ret[i].Typ.Equal(r.Type()) {
			panic("type mismatch")
		}
	}
}

func (t *typer) visitVarStmt(s *VarStmt) {
	s.Ini.Accept(t)

	if err := t.typ.Insert(s.Idf.Lit, &typ.Object{
		Cap: typ.C_ADR | typ.C_MOD,
		Typ: s.Ini.Type(),
	}); err != nil {
		panic(err)
	}
}

func (t *typer) visitLetStmt(s *LetStmt) {
	s.Ini.Accept(t)

	if err := t.typ.Insert(s.Idf.Lit, &typ.Object{
		Cap: typ.C_ADR | typ.C_MOD,
		Typ: s.Ini.Type(),
	}); err != nil {
		panic(err)
	}
}

func (t *typer) visitAssignStmt(a *AssignStmt) {
	a.Var.Accept(t)
	a.Val.Accept(t)

	if a.Var.Caps()&typ.C_MOD == 0 {
		panic("could not assign to immutable object")
	}

	if !a.Var.Type().Equal(a.Val.Type()) {
		panic("type mismatch")
	}
}

func (t *typer) VisitExpr(u Expr) {
	switch e := u.(type) {
	case *ImmExpr:
		t.visitImmExpr(e)
	case *StructExpr:
		t.visitStructExpr(e)
	case *IdfExpr:
		// TODO: position-independent definition
		e.Obj, _ = t.typ.Lookup(e.Tok.Lit)

		if e.Obj == nil {
			panic("undefined variable")
		}

		if e.Obj.Cap&typ.C_ADR == 0 {
			panic("misused identifier")
		}
	case *DotExpr:
		e.Env.Accept(t)

		if e.Caps()&typ.C_ADR == 0 {
			panic("misused dot expression")
		}

		switch e.Env.Type().Kind {
		case typ.PKG:
			inf := e.Env.Type().Extra.(*typ.Package)
			obj, _ := inf.Env.Lookup(e.Mem.Lit)

			if obj == nil {
				panic("unknown package member")
			}

			e.Typ = obj.Typ
		case typ.STRUCT:
			inf := e.Env.Type().Extra.(*typ.Struct)
			mem, ok := inf.Mem[e.Mem.Lit]

			if !ok {
				panic("unknown structure member")
			}

			e.Typ = mem.Typ
		case typ.ENUM:
			inf := e.Env.Type().Extra.(*typ.Enum)

			if _, ok := inf.Mem[e.Mem.Lit]; !ok {
				panic("unknown enum member")
			}

			e.Typ = e.Env.Type()
		default:
			panic("misused dot expression")
		}

	case *InfExpr:
		e.X.Accept(t)
		e.Y.Accept(t)

		tx := e.X.Type()
		ty := e.Y.Type()

		if tx.Kind != ty.Kind {
			panic("type mismatch")
		}

		if _, ok := compatibility[e.Tok.TokenType][e.X.Type().Kind]; !ok {
			panic("unsupported operation")
		}

		e.Typ = tx
	case *PfxExpr:
		e.X.Accept(t)

		if _, ok := compatibility[e.Tok.TokenType][e.X.Type().Kind]; !ok {
			panic("unsupported operation")
		}

		e.Typ = e.X.Type()
	case *CallExpr:
		sym, _ := t.typ.Lookup(e.Tok.Lit)

		if sym == nil || sym.Typ.Kind != typ.FUNC {
			panic("undeclared function")
		}

		// TODO: multiple return values

		e.FunTyp = sym.Typ

		for _, arg := range e.Arg {
			arg.Accept(t)

			n := arg.Type()

			if n.Kind == typ.STRUCT || n.Kind == typ.ARRAY {
				p := &typ.Type{Kind: n.Kind, Extra: n.Extra}

				n.Kind = typ.REF
				n.Extra = p
			}
		}

		fun := sym.Typ.Extra.(*typ.Func)

		if len(fun.Arg) != len(e.Arg) {
			panic("unsatisfied function signature")
		}

		for i := range fun.Arg {
			if fun.Arg[i].Typ.Kind == typ.ANY {
				continue
			}

			if !fun.Arg[i].Typ.Equal(e.Arg[i].Type()) {
				panic("unsatisfied function signature")
			}
		}

		if len(fun.Ret) == 0 {
			e.RetTyp = nil
		} else {
			e.RetTyp = fun.Ret[0].Typ
		}
	case *ToSigExpr:
		e.X.Accept(t)

		if _, ok := compatibility[e.Tok.TokenType][e.X.Type().Kind]; !ok {
			panic("unsupported operation")
		}

		e.Typ = &typ.Sig64Type
	case *ToUnsExpr:
		e.X.Accept(t)

		if _, ok := compatibility[e.Tok.TokenType][e.X.Type().Kind]; !ok {
			panic("unsupported operation")
		}

		e.Typ = &typ.Uns64Type

	case *ToFltExpr:
		e.X.Accept(t)

		if _, ok := compatibility[e.Tok.TokenType][e.X.Type().Kind]; !ok {
			panic("unsupported operation")
		}

		e.Typ = &typ.Flt64Type
	}
}

func (t *typer) visitImmExpr(e *ImmExpr) {
	if obj, ok := t.pkg.Imm[e.Tok.Lit]; ok {
		e.Obj = obj
		return
	}

	e.Obj = &typ.Object{}

	switch e.Tok.TokenType {
	case lex.II64:
		e.Obj.Typ = &typ.Sig64Type
		e.Obj.Val, _ = strconv.ParseInt(e.Tok.Lit, 10, 64)
	case lex.IU64:
		e.Obj.Typ = &typ.Uns64Type
		e.Obj.Val, _ = strconv.ParseUint(e.Tok.Lit[:len(e.Tok.Lit)-1], 10, 64)
	case lex.IF64:
		e.Obj.Typ = &typ.Flt64Type
		e.Obj.Val, _ = strconv.ParseFloat(e.Tok.Lit, 64)
	case lex.TRUE, lex.FALSE:
		e.Obj.Typ = &typ.BoolType
		e.Obj.Val, _ = strconv.ParseBool(e.Tok.Lit)
	}

	t.pkg.Imm[e.Tok.Lit] = e.Obj
}

func (t *typer) visitStructExpr(e *StructExpr) {
	obj, _ := t.env.Lookup(e.Tok.Lit)

	if obj == nil || obj.Typ.Kind != typ.STRUCT {
		panic("undefined structure")
	}

	inf := obj.Typ.Extra.(*typ.Struct)
	idx := 0

	for _, f := range e.Fields {
		mem, ok := inf.Mem[f.Tok.Lit]

		if !ok {
			panic("unknown structure member")
		}

		if mem.Idx < idx {
			panic("immediate structure fields must be ordered")
		}

		f.Val.Accept(t)

		if !mem.Typ.Equal(f.Val.Type()) {
			panic("type mismatch")
		}

		idx = mem.Idx
	}

	e.Typ = obj.Typ
}

var compatibility map[lex.TokenType]map[typ.Kind]bool = map[lex.TokenType]map[typ.Kind]bool{
	lex.ADD:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.SUB:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.MUL:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.DIV:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.POW:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.SHL:  {typ.I64: true, typ.U64: true},
	lex.SHR:  {typ.I64: true, typ.U64: true},
	lex.MOD:  {typ.I64: true, typ.U64: true},
	lex.BAND: {typ.I64: true, typ.U64: true},
	lex.BOR:  {typ.I64: true, typ.U64: true},
	lex.BXOR: {typ.I64: true, typ.U64: true},

	lex.BNEG: {typ.I64: true, typ.U64: true},
	lex.UNEG: {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.LNEG: {typ.BOOL: true},

	lex.LT:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.LE:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.GT:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.GE:  {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.EEQ: {typ.I64: true, typ.U64: true, typ.F64: true},
	lex.NE:  {typ.I64: true, typ.U64: true, typ.F64: true},

	lex.LAND: {typ.BOOL: true},
	lex.LOR:  {typ.BOOL: true},

	lex.I64: {typ.U64: true, typ.F64: true},
	lex.U64: {typ.I64: true, typ.F64: true},
	lex.F64: {typ.I64: true, typ.U64: true},
}
