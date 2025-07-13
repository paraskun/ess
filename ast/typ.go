package ast

import (
	"strconv"

	"github.com/paraskun/x/lex"
	"github.com/paraskun/x/typ"
)

type typer struct {
	env *typ.Env
	fun *typ.Func
}

func Typeset(pkg *typ.Package) {
	pkg.Env = typ.NewEnv(nil)
	typ := typer{env: pkg.Env}

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

	if err := t.env.Insert(d.Pkg.Name, &typ.Object{
		Typ: &typ.Type{
			Kind:  typ.PKG,
			Extra: d.Pkg,
		},
	}); err != nil {
		panic(err)
	}

	Typeset(d.Pkg)
}

func (t *typer) visitVarDecl(d *VarDecl) {
	d.Ini.Accept(t)

	if err := t.env.Insert(d.Idf.Lit, &typ.Object{
		Typ: d.Ini.Type(),
		Val: d.Ini.(*ImmExpr).Obj.Val, // TODO: expression evaluation
	}); err != nil {
		panic(err)
	}
}

func (t *typer) visitFuncDecl(d *FuncDecl) {
	d.Env = typ.NewEnv(t.env)
	t.env = d.Env
	t.fun = t.funcSpec(d)

	d.Body.Accept(t)

	if err := t.env.Insert(d.Idf.Lit, &typ.Object{
		Typ: &typ.Type{
			Kind:  typ.FUNC,
			Extra: t.fun,
		},
	}); err != nil {
		panic(err)
	}

	t.fun = nil
	t.env = d.Env.Parent
}

func (t *typer) funcSpec(d *FuncDecl) *typ.Func {
	fun := &typ.Func{}

	for _, arg := range d.Arg {
		t.typeSpec(arg.Typ)

		obj := &typ.Object{
			Typ: arg.Typ.Typ,
		}

		if err := t.env.Insert(arg.Tok.Lit, obj); err != nil {
			panic(err)
		}

		fun.Arg = append(fun.Arg, &typ.Field{
			Name: arg.Tok.Lit,
			Typ:  obj.Typ,
		})
	}

	for _, ret := range d.Ret {
		t.typeSpec(ret.Typ)

		fun.Ret = append(fun.Ret, &typ.Field{
			Typ: ret.Typ.Typ,
		})
	}

	return fun
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
		sym, _ := t.env.Lookup(s.Tok.Lit)

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
	t.visitTypeSpec(d.Spec, false)

	if err := t.curEnv.InsertSym(d.Tok.Lit, d.Spec.Type()); err != nil {
		panic(err)
	}
}

func (t *typer) visitTypeSpec(u TypeSpec, env bool) {
	switch s := u.(type) {
	case *BaseSpec:
		t.visitBaseSpec(s)
	case *FuncSpec:
		t.visitFuncSpec(s, env)
	case *CompSpec:
		t.visitCompSpec(s)
	}
}

func (t *typer) visitBaseSpec(s *BaseSpec) {
}

func (t *typer) visitFuncSpec(s *FuncSpec, env bool) {
}

func (t *typer) visitCompSpec(s *CompSpec) {
	inf := typ.CompInfo{Fields: make(map[string]*typ.Field)}
	off := 0

	for _, fld := range s.Fields {
		t.visitTypeSpec(fld.Typ, false)

		if _, ok := inf.Fields[fld.Tok.Lit]; ok {
			panic("duplicated field")
		}

		inf.Fields[fld.Tok.Lit] = &typ.Field{
			Name: fld.Tok.Lit,
			Typ:  fld.Typ.Type(),
			Off:  off,
		}

		off += fld.Typ.Type().Size()
	}

	s.Typ = &typ.Type{
		Kind: typ.COMP,
		Info: &inf,
	}
}

func (t *typer) VisitStmt(u Stmt) {
	switch s := u.(type) {
	case *BlockStmt:
		s.Env = typ.NewEnv(t.curEnv)
		t.curEnv = s.Env

		for _, o := range s.Body {
			o.Accept(t)
		}

		t.curEnv = s.Env.Parent
	case *AssignStmt:
		t.visitAssignStmt(s)
	case *ReturnStmt:
		t.visitReturnStmt(s)
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
		s.Exp.Accept(t)

		if s.Exp.Type() != nil {
			panic("missed return")
		}
	}
}

func (t *typer) visitAssignStmt(a *AssignStmt) {
	a.Val.Accept(t)

	if a.Dec {
		idf := a.Var.(*IdfExpr)
		idf.Obj = &typ.Object{
			Typ: a.Val.Type(),
			Off: -1,
		}

		if err := t.curEnv.InsertObj(idf.Tok.Lit, idf.Obj); err != nil {
			panic(err)
		}

		return
	}

	a.Var.Accept(t)

	if !a.Var.Type().Equal(a.Val.Type()) {
		panic("type mismatch")
	}
}

func (t *typer) visitReturnStmt(r *ReturnStmt) {
	r.Ret.Accept(t)

	if t.curInf.Ret == nil {
		panic("void function does not suppose to return")
	}

	if !r.Ret.Type().Equal(t.curInf.Ret.Typ) {
		panic("type mismatch")
	}
}

func (t *typer) VisitExpr(u Expr) {
	switch e := u.(type) {
	case *BaseImmExpr:
		t.visitBaseImmExpr(e)
	case *CompImmExpr:
		t.visitCompImmExpr(e)
	case *IdfExpr:
		e.Obj, _ = t.curEnv.LookupObj(e.Tok.Lit)

		if e.Obj == nil {
			panic("undefined variable")
		}
	case *DotExpr:
		e.Comp.Accept(t)

		if e.Comp.Type().Kind != typ.COMP {
			panic("misused dot expression")
		}

		inf := e.Comp.Type().Info.(*typ.CompInfo)
		fld, ok := inf.Fields[e.Field.Lit]

		if !ok {
			panic("no such field")
		}

		e.Typ = fld.Typ
	case *InfExpr:
		e.X.Accept(t)
		e.Y.Accept(t)

		tx := e.X.Type()
		ty := e.Y.Type()

		if tx.Kind != ty.Kind {
			panic("type mismatch")
		}

		if tx.Kind == typ.COMP || tx.Kind == typ.FUNC {
			panic("unsupported operand")
		}

		if ty.Kind == typ.COMP || ty.Kind == typ.FUNC {
			panic("unsupported operand")
		}

		e.Typ = tx
	case *PfxExpr:
		e.X.Accept(t)

		if e.X.Type().Kind == typ.COMP || e.X.Type().Kind == typ.FUNC {
			panic("unsupported operand")
		}

		e.Typ = e.X.Type()
	case *CallExpr:
		sym, _ := t.curEnv.LookupFun(e.Tok.Lit)

		if sym == nil {
			panic("undeclared function")
		}

		e.Sym = sym

		for _, arg := range e.Arg {
			arg.Accept(t)
		}

		inf := sym.Info.(*typ.FuncInfo)

		if len(inf.Arg) != len(e.Arg) {
			panic("function argument disbalance")
		}

		for i := range inf.Arg {
			if inf.Arg[i].Typ.Kind == typ.ANY {
				continue
			}

			if !inf.Arg[i].Typ.Equal(e.Arg[i].Type()) {
				panic("type mismatch")
			}
		}

		if inf.Ret == nil {
			e.Typ = nil
		} else {
			e.Typ = inf.Ret.Typ
		}
	case *ToSigExpr:
		e.X.Accept(t)

		switch e.X.Type().Kind {
		case typ.FUNC, typ.COMP, typ.BOOL:
			panic("unsupported operand type")
		}

		e.Typ = &typ.Sig64Type
	case *ToUnsExpr:
		e.X.Accept(t)

		switch e.X.Type().Kind {
		case typ.FUNC, typ.COMP, typ.BOOL:
			panic("unsupported operand type")
		}

		e.Typ = &typ.Uns64Type

	case *ToFltExpr:
		e.X.Accept(t)

		switch e.X.Type().Kind {
		case typ.FUNC, typ.COMP, typ.BOOL:
			panic("unsupported operand type")
		}

		e.Typ = &typ.Flt64Type
	}
}

func (t *typer) visitBaseImmExpr(e *BaseImmExpr) {
	if imm := t.curEnv.LookupImm(e.Tok.Lit); imm != nil {
		e.Obj = imm
		return
	}

	e.Obj = &typ.Object{
		Off: -1,
	}

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

	t.curEnv.InsertImm(e.Tok.Lit, e.Obj)
}

func (t *typer) visitCompImmExpr(e *CompImmExpr) {
	sym, _ := t.curEnv.LookupSym(e.Tok.Lit)

	if sym == nil || sym.Kind != typ.COMP {
		panic("undefined compound")
	}

	inf := sym.Info.(*typ.CompInfo)
	off := 0

	for _, f := range e.Fields {
		fld, ok := inf.Fields[f.Tok.Lit]

		if !ok {
			panic("unknown field")
		}

		if fld.Off < off {
			panic("immediate compounds must be ordered")
		}

		f.Val.Accept(t)

		if !fld.Typ.Equal(f.Val.Type()) {
			panic("type mismatch")
		}

		off = fld.Off
	}

	e.Typ = sym
}
