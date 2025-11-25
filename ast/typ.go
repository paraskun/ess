package ast

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/paraskun/o2/lex"
	"github.com/paraskun/o2/tty"
	"github.com/paraskun/o2/typ"
	"github.com/paraskun/o2/typ/mod"
)

type typer struct {
	pkg *mod.Package
	src *mod.File
	env *typ.Env
	fun *typ.Func
	ctx tty.Span
	err int
}

func Typeset(pkg *mod.Package) bool {
	pkg.Env = typ.NewEnv(nil)
	ty := typer{pkg: pkg, env: pkg.Env.(*typ.Env)}

	for _, src := range pkg.Src {
		ty.src = src
		src.Env = pkg.Env

		for _, dec := range src.Dec.(*File).Dec {
			ty.ctx = src.Dec.(*File).Box
			dec.Accept(&ty)
		}
	}

	return ty.err == 0
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
	// Ignore an error, if multiple files using same package.
	_, _ = t.env.Insert(d.Pkg.Lit[0:len(d.Pkg.Lit)-1], d.Obj)

	Typeset(d.Obj.Typ.Extra.(*mod.Package))
}

func (t *typer) visitVarDecl(d *VarDecl) {
	t.ctx = d.Box
	d.Ini.Accept(t)

	if _, err := t.env.Insert(d.Var.Lit, &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: d.Box},
		Typ:  d.Ini.Object().Typ,
		Seg:  typ.Package,
		Off:  t.pkg.Dsz,
		Val:  d.Ini.Object().Val,
	}); err != nil {
		t.err += 1
		err.Note(os.Stdout)
	}

	t.pkg.Dsz += d.Ini.Object().Size()
}

func (t *typer) visitFuncDecl(d *FuncDecl) {
	t.ctx = d.Sig
	d.Env = typ.NewEnv(t.env)

	f := t.visitFuncSpec(d)

	d.Obj = &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: d.Box},
		Typ: &typ.Type{
			Kind:  typ.DEC,
			Extra: f,
		},
	}

	if _, err := t.env.Insert(d.Sym.Lit, d.Obj); err != nil {
		t.err += 1
		err.Note(os.Stdout)
	}

	t.ctx = d.Box
	t.env = d.Env
	t.fun = f.Extra.(*typ.Func)

	if d.Sub != nil {
		d.Sub.Accept(t)
	}

	t.fun = nil
	t.env = d.Env.Parent
}

func (t *typer) visitFuncSpec(d *FuncDecl) *typ.Type {
	e := &typ.Func{Name: d.Sym.Lit, Dec: d}
	r := &typ.Type{Kind: typ.FUNC, Extra: e}

	for _, arg := range d.Arg {
		t.visitTypeSpec(arg.Typ)

		obj := &typ.Object{
			Snip: &tty.Snippet{File: t.src, Span: arg.Box},
			Typ:  arg.Typ.Typ,
		}

		if _, err := d.Env.Insert(arg.Sym.Lit, obj); err != nil {
			t.err += 1
			err.Note(os.Stdout)

		}

		e.Arg = append(e.Arg, &typ.Field{
			Name: arg.Sym.Lit,
			Typ:  obj.Typ,
			Dec:  arg,
		})
	}

	if d.Ret != nil {
		t.ctx = d.Sig
		t.visitTypeSpec(d.Ret)

		e.Ret = &typ.Field{
			Typ: d.Ret.Typ,
			Dec: d.Ret,
		}
	}

	return r
}

func (t *typer) visitTypeSpec(s *TypeSpec) {
	switch s.Lex.Typ {
	case lex.BOOL:
		s.Typ = &typ.BoolType
	case lex.I64:
		s.Typ = &typ.I64Type
	case lex.U64:
		s.Typ = &typ.U64Type
	case lex.F64:
		s.Typ = &typ.F64Type
	case lex.IDEN:
		sym, _ := t.env.Lookup(s.Lex.Tok.Lit)

		if sym == nil || (sym.Typ.Kind != typ.ENUM && sym.Typ.Kind != typ.STRUCT) {
			err := &typ.Error{
				Full: fmt.Sprintf("there is no type \"%s\" in this scope", s.Lex.Tok.Lit),
				Snip: []*tty.Snippet{{File: t.src, Span: t.ctx}},
			}

			s.Lex.Tok.Hint().Text = "unknown type here"
			s.Lex.Tok.Hint().Attr.Color.Add(color.FgRed)

			t.err += 1
			err.Note(os.Stdout)

			return
		}

		s.Typ = sym.Typ
	}
}

func (t *typer) visitStructDecl(d *StructDecl) {
	d.Obj = &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: d.Box},
		Typ: &typ.Type{
			Kind:  typ.DEC,
			Extra: t.visitStructSpec(d),
		},
	}

	if _, err := t.env.Insert(d.Sym.Lit, d.Obj); err != nil {
		t.err += 1
		err.Note(os.Stdout)
	}
}

func (t *typer) visitStructSpec(d *StructDecl) *typ.Type {
	e := &typ.Struct{Name: d.Sym.Lit, Mem: make(map[string]*typ.Field), Dec: d}
	r := &typ.Type{Kind: typ.STRUCT, Extra: e}

	for _, mem := range d.Mem {
		t.visitTypeSpec(mem.Typ)

		if prv, ok := e.Mem[mem.Sym.Lit]; ok {
			err := &typ.Error{
				Full: fmt.Sprintf("the field \"%s\" is defined multiple times", mem.Sym.Lit),
				Snip: []*tty.Snippet{
					{File: t.src, Span: d.Box},
					{File: t.src, Span: d.Box},
				},
			}

			prv.Dec.(*NamedField).Box.Hint().Text = "previous definiton here"
			prv.Dec.(*NamedField).Box.Hint().Attr.Color.Add(color.FgCyan)

			mem.Box.Hint().Text = "redefined here"
			mem.Box.Hint().Attr.Color.Add(color.FgRed)

			t.err += 1
			err.Note(os.Stdout)
		} else {
			e.Mem[mem.Sym.Lit] = &typ.Field{
				Name: mem.Sym.Lit,
				Typ:  mem.Typ.Typ,
				Dec:  mem,
			}
		}
	}

	return r
}

func (t *typer) visitEnumDecl(d *EnumDecl) {
	d.Obj = &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: d.Box},
		Typ: &typ.Type{
			Kind:  typ.DEC,
			Extra: t.visitEnumSpec(d),
		},
	}

	if _, err := t.env.Insert(d.Sym.Lit, d.Obj); err != nil {
		t.err += 1
		err.Note(os.Stdout)
	}
}

func (t *typer) visitEnumSpec(d *EnumDecl) *typ.Type {
	e := &typ.Enum{Name: d.Sym.Lit, Mem: make(map[string]*typ.Field), Dec: d}
	r := &typ.Type{Kind: typ.ENUM, Extra: e}

	for _, mem := range d.Mem {
		if prv, ok := e.Mem[mem.Lit]; ok {
			err := &typ.Error{
				Full: fmt.Sprintf("the member \"%s\" is defined multiple times", mem.Lit),
				Snip: []*tty.Snippet{
					{File: t.src, Span: d.Box},
				},
			}

			prv.Dec.(*tty.Tok).Hint().Text = "previous definiton here"
			prv.Dec.(*tty.Tok).Hint().Attr.Color.Add(color.FgCyan)

			mem.Hint().Text = "redefined here"
			mem.Hint().Attr.Color.Add(color.FgRed)

			t.err += 1
			err.Note(os.Stdout)
		} else {
			e.Mem[mem.Lit] = &typ.Field{
				Name: mem.Lit,
				Typ:  r,
				Dec:  mem,
			}
		}

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
		s.Env = typ.NewEnv(t.env)
		t.env = s.Env

		for _, o := range s.Sub {
			o.Accept(t)
		}

		t.env = s.Env.Parent
	case *LoopStmt:
		s.Con.Accept(t)
		s.Sub.Accept(t)
	case *CondStmt:
		s.Con.Accept(t)
		s.Pos.Accept(t)

		if s.Neg != nil {
			s.Neg.Accept(t)
		}
	case *CallStmt:
		s.CallExpr.Accept(t)

		f := s.Fun.Typ.Extra.(*typ.Func)
		d := f.Dec.(*FuncDecl)

		if f.Ret != nil {
			err := &typ.Error{
				Full: "missing function return value",
				Snip: []*tty.Snippet{
					{File: t.src, Span: t.ctx},
					{File: s.Fun.Snip.File, Span: d.Span()},
				},
			}

			s.Span().Hint().Text = "this call produces something"
			s.Span().Hint().Attr.Color.Add(color.FgRed)

			d.Span().Hint().Text = "as defined here"
			d.Span().Hint().Attr.Color.Add(color.FgCyan)

			t.err += 1
			err.Note(os.Stdout)
		}
	}
}

func (t *typer) visitReturnStmt(r *ReturnStmt) {
	d := t.fun.Dec.(*FuncDecl)

	if t.fun.Ret == nil {
		err := &typ.Error{
			Full: "unexpected return statement",
			Snip: []*tty.Snippet{
				{File: t.src, Span: d.Span()},
			},
		}

		fmt.Printf("%+v\n", d.Sig)

		r.Box.Hint().Text = "this shouldn't have happened"
		r.Box.Hint().Attr.Color.Add(color.FgRed)

		d.Sig.Hint().Text = "as defined here"
		d.Sig.Hint().Attr.Color.Add(color.FgCyan)

		t.err += 1
		err.Note(os.Stdout)

		return
	}

	r.Ret.Accept(t)

	if !t.fun.Ret.Typ.Equal(r.Ret.Object().Typ) {
		err := &typ.Error{
			Full: "type mismatch",
			Snip: []*tty.Snippet{
				{File: t.src, Span: d.Span()},
			},
		}

		r.Ret.Span().Hint().Text = fmt.Sprintf("'%s' expected here", t.fun.Ret.Typ.String())
		r.Ret.Span().Hint().Attr.Color.Add(color.FgRed)

		d.Sig.Hint().Text = "as defined here"
		d.Sig.Hint().Attr.Color.Add(color.FgCyan)

		t.err += 1
		err.Note(os.Stdout)
	}
}

func (t *typer) visitVarStmt(s *VarStmt) {
	s.Ini.Accept(t)

	if _, err := t.env.Insert(s.Var.Lit, &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: s.Box},
		Typ:  s.Ini.Object().Typ,
	}); err != nil {
		t.err += 1
		err.Note(os.Stdout)
	}
}

func (t *typer) visitLetStmt(s *LetStmt) {
	s.Ini.Accept(t)

	if _, err := t.env.Insert(s.Var.Lit, &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: s.Box},
		Typ:  s.Ini.Object().Typ,
	}); err != nil {
		t.err += 1
		err.Note(os.Stdout)
	}
}

func (t *typer) visitAssignStmt(a *AssignStmt) {
	a.Var.Accept(t)
	a.Val.Accept(t)

	if !a.Var.Object().Typ.Equal(a.Val.Object().Typ) {
		a.Box.Hint().Text = "type mismatch"
		a.Box.Hint().Attr.Color.Add(color.FgRed)
	}
}

func (t *typer) VisitExpr(u Expr) {
	switch e := u.(type) {
	case *BasicExpr:
		t.visitBasicExpr(e)
	case *StructExpr:
		t.visitStructExpr(e)
	case *IdenExpr:
		// TODO: position-independent definition
		e.Obj, _ = t.env.Lookup(e.Tok.Lit)

		if e.Obj == nil {
			panic("undefined variable")
		}
	case *DotExpr:
		e.Ctx.Accept(t)

		switch e.Ctx.Object().Typ.Kind {
		case typ.PKG:
			inf := e.Ctx.Object().Typ.Extra.(*mod.Package)
			e.Obj, _ = inf.Env.(*typ.Env).Lookup(e.Mem.Lit)

			if e.Obj == nil {
				panic("unknown package member")
			}
		case typ.STRUCT:
			inf := e.Ctx.Object().Typ.Extra.(*typ.Struct)
			mem, ok := inf.Mem[e.Mem.Lit]

			if !ok {
				panic("unknown structure member")
			}

			e.Obj = &typ.Object{
				Typ: mem.Typ,
			}
		case typ.ENUM:
			inf := e.Ctx.Object().Typ.Extra.(*typ.Enum)

			if _, ok := inf.Mem[e.Mem.Lit]; !ok {
				panic("unknown enum member")
			}

			e.Obj = &typ.Object{
				Typ: e.Obj.Typ,
			}
		default:
			panic("misused dot expression")
		}

	case *InfExpr:
		e.X.Accept(t)
		e.Y.Accept(t)

		ox := e.X.Object()
		oy := e.Y.Object()

		if ox.Typ.Kind != oy.Typ.Kind {
			panic("type mismatch")
		}

		if _, ok := compatibility[e.Lex.Typ][ox.Typ.Kind]; !ok {
			panic("unsupported operation")
		}

		e.Res = &typ.Object{
			Typ: ox.Typ,
		}
	case *PfxExpr:
		e.X.Accept(t)

		ox := e.X.Object()

		if _, ok := compatibility[e.Lex.Typ][ox.Typ.Kind]; !ok {
			panic("unsupported operation")
		}

		e.Res = &typ.Object{
			Typ: ox.Typ,
		}
	case *CallExpr:
		sym, _ := t.env.Lookup(e.Sym.Lit)

		if sym == nil || sym.Typ.Kind != typ.DEC || sym.Typ.Extra.(*typ.Type).Kind != typ.FUNC {
			panic("unknown function")
		}

		e.Fun = sym

		for _, arg := range e.Arg {
			arg.Accept(t)
		}

		fun := sym.Typ.Extra.(*typ.Func)

		if len(fun.Arg) != len(e.Arg) {
			panic("unsatisfied function signature")
		}

		for i := range fun.Arg {
			if fun.Arg[i].Typ.Kind == typ.ANY {
				continue
			}

			if !fun.Arg[i].Typ.Equal(e.Arg[i].Object().Typ) {
				panic("unsatisfied function signature")
			}
		}
	case *ToI64Expr:
		e.X.Accept(t)

		if _, ok := compatibility[lex.I64][e.X.Object().Typ.Kind]; !ok {
			panic("unsupported operation")
		}

		e.Res = &typ.Object{
			Snip: &tty.Snippet{File: t.src, Span: e.Span()},
			Typ:  &typ.I64Type,
		}
	case *ToU64Expr:
		e.X.Accept(t)

		if _, ok := compatibility[lex.U64][e.X.Object().Typ.Kind]; !ok {
			panic("unsupported operation")
		}

		e.Res = &typ.Object{
			Snip: &tty.Snippet{File: t.src, Span: e.Span()},
			Typ:  &typ.U64Type,
		}
	case *ToF64Expr:
		e.X.Accept(t)

		if _, ok := compatibility[lex.F64][e.X.Object().Typ.Kind]; !ok {
			panic("unsupported operation")
		}

		e.Res = &typ.Object{
			Snip: &tty.Snippet{File: t.src, Span: e.Span()},
			Typ:  &typ.U64Type,
		}
	}
}

func (t *typer) visitBasicExpr(e *BasicExpr) {
	if obj, _ := t.pkg.Env.(*typ.Env).Lookup(e.Lex.Tok.Lit); obj != nil {
		e.Obj = obj
		return
	}

	e.Obj = &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: e.Span()},
	}

	switch e.Lex.Typ {
	case lex.II64:
		e.Obj.Typ = &typ.I64Type
		e.Obj.Val, _ = strconv.ParseInt(e.Lex.Tok.Lit, 10, 64)
	case lex.IU64:
		e.Obj.Typ = &typ.U64Type
		e.Obj.Val, _ = strconv.ParseUint(e.Lex.Tok.Lit[:len(e.Lex.Tok.Lit)-1], 10, 64)
	case lex.IF64:
		e.Obj.Typ = &typ.F64Type
		e.Obj.Val, _ = strconv.ParseFloat(e.Lex.Tok.Lit, 64)
	case lex.STR:
		e.Obj.Typ = &typ.StrType
		e.Obj.Val = e.Lex.Tok.Lit[1 : len(e.Lex.Tok.Lit)-1]
	case lex.TRUE, lex.FALSE:
		e.Obj.Typ = &typ.BoolType
		e.Obj.Val, _ = strconv.ParseBool(e.Lex.Tok.Lit)
	}

	e.Obj.Off = t.pkg.Dsz
	t.pkg.Dsz += e.Obj.Size()
	t.pkg.Env.(*typ.Env).Insert(e.Lex.Tok.Lit, e.Obj)
}

func (t *typer) visitStructExpr(e *StructExpr) {
	obj, _ := t.env.Lookup(e.Sym.Lit)

	if obj == nil || obj.Typ.Kind != typ.STRUCT {
		panic("undefined structure")
	}

	inf := obj.Typ.Extra.(*typ.Struct)

	for _, f := range e.Mem {
		mem, ok := inf.Mem[f.Mem.Lit]

		if !ok {
			panic("unknown structure member")
		}

		f.Val.Accept(t)

		if !mem.Typ.Equal(f.Val.Object().Typ) {
			panic("type mismatch")
		}
	}

	e.Obj = &typ.Object{
		Snip: &tty.Snippet{File: t.src, Span: e.Span()},
		Typ:  obj.Typ,
	}
}

var compatibility map[lex.Type]map[typ.Kind]bool = map[lex.Type]map[typ.Kind]bool{
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
