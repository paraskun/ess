package ast

import (
	"strconv"

	"github.com/paraskun/ess-go/lex"
	"github.com/paraskun/ess-go/typ"
)

type typer struct {
	env *typ.Env
	typ *typ.FuncType
}

func Typeset(p *Pragma) {
	c := typer{
		env: p.Env,
	}

	for _, s := range p.Body {
		s.Accept(&c)
	}
}

func (t *typer) VisitStmt(s Stmt) {
	switch v := s.(type) {
	case *BlockStmt:
		if v.Env == nil {
			v.Env = typ.NewEnv(t.env)
		}

		t.env = v.Env

		for _, o := range v.Body {
			o.Accept(t)
		}

		t.env = v.Env.Env
	case *AssignStmt:
		t.visitAssignStmt(v)
	case *ReturnStmt:
		t.visitReturnStmt(v)
	case *LoopStmt:
		v.Con.Accept(t)
		v.Rep.Accept(t)
	case *CondStmt:
		v.Con.Accept(t)
		v.Pos.Accept(t)

		if v.Neg != nil {
			v.Neg.Accept(t)
		}
	}
}

func (t *typer) visitReturnStmt(r *ReturnStmt) {
	re := make([]*typ.Type, 0)

	for _, i := range r.Arg {
		i.Accept(t)
		re = append(re, i.Type()...)
	}

	if len(re) != len(t.typ.Ret) {
		panic("return disbalance")
	}

	for i, r := range re {
		if !r.Equal(t.typ.Ret[i].Typ) {
			panic("type mismatch")
		}
	}
}

func (t *typer) visitAssignStmt(a *AssignStmt) {
	re := make([]*typ.Type, 0)

	for _, i := range a.Val {
		i.Accept(t)
		re = append(re, i.Type()...)
	}

	if len(a.Var) != len(re) {
		panic("assignment disbalance")
	}

	if a.New {
		for i, j := range a.Var {
			idf := j.(*IdfExpr)

			if err := t.env.InsertObj(idf.Tok.Lit, &typ.Object{
				Typ: re[i],
				Env: t.env,
				Off: -1,
				Loc: true,
			}); err != nil {
				panic(err)
			}
		}

		return
	}

	for i, j := range a.Var {
		if _, ok := j.(*CallExpr); ok {
			panic("could not assign to function call")
		}

		j.Accept(t)

		if !j.Type()[0].Equal(re[i]) {
			panic("type mismatch")
		}
	}
}

func (t *typer) VisitDecl(u Decl) {
	switch d := u.(type) {
	case *FuncDecl:
		d.Body.Env = typ.NewEnv(t.env)
		t.env = d.Body.Env
		t.env.Top = t.env

		t.visitTypeSpec(d.Spec, true)

		t.typ = d.Spec.Type().Info.(*typ.FuncType)

		d.Body.Accept(t)

		t.typ = nil

		if err := t.env.InsertObj(d.Tok.Lit, &typ.Object{
			Typ: d.Spec.Type(),
			Env: t.env,
			Off: -1,
			Loc: false,
		}); err != nil {
			panic(err)
		}
	case *CompDecl:
		t.visitTypeSpec(d.Spec, false)

		if err := t.env.InsertSym(d.Tok.Lit, d.Spec.Type()); err != nil {
			panic(err)
		}
	}
}

func (t *typer) visitTypeSpec(u TypeSpec, env bool) {
	switch s := u.(type) {
	case *BaseSpec:
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
			sym, _ := t.env.LookupSym(s.Tok.Lit)

			if sym == nil {
				panic("undeclared type")
			}

			s.Typ = sym
		}
	case *FuncSpec:
		ft := typ.FuncType{}
		off := 0

		for _, fs := range s.Arg {
			t.visitTypeSpec(fs.Typ, false)

			if env {
				if fs.Tok == nil {
					panic("unnamed argument")
				}

				fs.Obj = &typ.Object{
					Typ: fs.Typ.Type(),
					Env: t.env,
					Off: -1,
					Loc: true,
				}

				if err := t.env.InsertObj(fs.Tok.Lit, fs.Obj); err != nil {
					panic(err)
				}

				if fs.Typ.Type().Kind == typ.COMP {
					t.env.Obj[fs.Tok.Lit].Ref = true
				}
			}

			ft.Arg = append(ft.Arg, typ.Field{
				Name: fs.Tok.Lit,
				Typ:  fs.Typ.Type(),
				Off:  off,
			})

			off += fs.Typ.Type().Size()
		}

		off = 0

		for _, fs := range s.Ret {
			t.visitTypeSpec(fs.Typ, false)

			rf := typ.Field{
				Typ: fs.Typ.Type(),
				Off: off,
			}

			off += fs.Typ.Type().Size()

			if env && fs.Tok != nil {
				if err := t.env.InsertObj(fs.Tok.Lit, &typ.Object{
					Typ: fs.Typ.Type(),
					Env: t.env,
					Off: -1,
					Loc: true,
				}); err != nil {
					panic(err)
				}

				rf.Name = fs.Tok.Lit
			}

			ft.Ret = append(ft.Ret, rf)
		}

		s.Typ = &typ.Type{
			Kind: typ.FUNC,
			Info: &ft,
		}
	case *CompSpec:
		ct := typ.CompType{Fields: make(map[string]typ.Field)}
		off := 0

		for _, fs := range s.Fields {
			if fs.Tok == nil {
				panic("unnamed field")
			}

			t.visitTypeSpec(fs.Typ, false)

			if _, ok := ct.Fields[fs.Tok.Lit]; ok {
				panic("duplicate field")
			}

			ct.Fields[fs.Tok.Lit] = typ.Field{
				Name: fs.Tok.Lit,
				Typ:  fs.Typ.Type(),
				Off:  off,
			}

			off += fs.Typ.Type().Size()
		}

		s.Typ = &typ.Type{
			Kind: typ.COMP,
			Info: &ct,
		}
	}
}

func (t *typer) VisitExpr(u Expr) {
	switch e := u.(type) {
	case *BaseImmExpr:
		if imm := t.env.LookupImm(e.Tok.Lit); imm != nil {
			e.Obj = imm
			break
		}

		e.Obj = &typ.Object{
			Env: t.env.Top,
			Off: -1,
			Loc: true,
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

		t.env.InsertImm(e.Tok.Lit, e.Obj)

	case *CompImmExpr:
		tt, _ := t.env.LookupSym(e.Tok.Lit)
		po := 0

		if tt == nil || tt.Kind != typ.COMP {
			panic("undefined compound")
		}

		ci := tt.Info.(*typ.CompType)

		for _, f := range e.Fields {
			cf, ok := ci.Fields[f.Tok.Lit]

			if !ok {
				panic("unknown field")
			}

			if cf.Off > po {
				panic("unordered field")
			}

			f.Val.Accept(t)

			if len(f.Val.Type()) > 1 {
				panic("assignment disbalance")
			}

			if !cf.Typ.Equal(f.Val.Type()[0]) {
				panic("type mismatch")
			}

			po = cf.Off
		}

	case *IdfExpr:
		e.Obj, _ = t.env.LookupObj(e.Tok.Lit)

		if e.Obj == nil {
			panic("undefined variable")
		}
	case *DotExpr:
		e.Comp.Accept(t)

		if len(e.Comp.Type()) > 1 {
			panic("misused dot expression")
		}

		if e.Comp.Type()[0].Kind != typ.COMP {
			panic("misused dot expression")
		}

		c := e.Comp.Type()[0].Info.(*typ.CompType)
		f, o := c.Fields[e.Field.Lit]

		if !o {
			panic("no such field")
		}

		e.Typ = f.Typ

	case *InfExpr:
		e.X.Accept(t)
		e.Y.Accept(t)

		if len(e.X.Type()) > 1 || len(e.Y.Type()) > 1 {
			panic("improper use of multivariable expression")
		}

		tx := e.X.Type()[0]
		ty := e.Y.Type()[0]

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

		if len(e.X.Type()) > 1 {
			panic("improper use of multivariable expression")
		}

		e.Typ = e.X.Type()[0]

	case *CallExpr:
		tt, _ := t.env.LookupObj(e.Tok.Lit)

		if tt == nil {
			panic("undeclared function")
		}

		re := make([]*typ.Type, 0)

		for _, arg := range e.Arg {
			arg.Accept(t)
			re = append(re, arg.Type()...)
		}

		ft := tt.Typ.Info.(*typ.FuncType)

		if len(ft.Arg) != len(re) {
			panic("function argument disbalance")
		}

		for i, at := range re {
			if !at.Equal(ft.Arg[i].Typ) {
				panic("type mismatch")
			}
		}

		for _, ret := range ft.Ret {
			e.Typ = append(e.Typ, ret.Typ)
		}

	case *ToSigExpr:
		e.X.Accept(t)

		if len(e.X.Type()) > 1 {
			panic("could not cast multivariable expression")
		}

		switch e.X.Type()[0].Kind {
		case typ.FUNC, typ.COMP, typ.BOOL:
			panic("unsupported operand type")
		}

		e.Typ = &typ.Sig64Type

	case *ToUnsExpr:
		e.X.Accept(t)

		if len(e.X.Type()) > 1 {
			panic("could not cast multivariable expression")
		}

		switch e.X.Type()[0].Kind {
		case typ.FUNC, typ.COMP, typ.BOOL:
			panic("unsupported operand type")
		}

		e.Typ = &typ.Uns64Type

	case *ToFltExpr:
		e.X.Accept(t)

		if len(e.X.Type()) > 1 {
			panic("could not cast multivariable expression")
		}

		switch e.X.Type()[0].Kind {
		case typ.FUNC, typ.COMP, typ.BOOL:
			panic("unsupported operand type")
		}

		e.Typ = &typ.Flt64Type
	}
}
