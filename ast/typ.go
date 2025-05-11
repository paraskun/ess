package ast

import (
	"github.com/paraskun/ess-go/lex"
	"github.com/paraskun/ess-go/typ"
)

type typer struct {
	env *typ.Env
	fun *typ.FuncType
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
		v.Env = &typ.Env{Env: t.env}
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

	if len(re) != len(t.fun.Ret) {
		panic("return disbalance")
	}

	for i, r := range re {
		if !r.Equal(t.fun.Ret[i].Type) {
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

			if obj, lvl := t.env.LookupObj(idf.Tok.Lit); obj != nil && lvl != 0 {
				panic("name collision")
			}

			t.env.Obj[idf.Tok.Lit] = &typ.Object{
				Typ: re[i],
				Env: t.env,
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
	case *CompDecl:
	}
}

func (t *typer) VisitExpr(u Expr) {
	switch e := u.(type) {
	case *BaseImmExpr:
		switch e.Tok.TokenType {
		case lex.II64:
			e.Typ = &typ.Sig64Type
		case lex.IU64:
			e.Typ = &typ.Uns64Type
		case lex.IF64:
			e.Typ = &typ.Flt64Type
		}
	case *CompImmExpr:
		tt, _ := t.env.LookupSym(e.Tok.Lit)

		if tt == nil || tt.Kind != typ.COMP {
			panic("undefined compound")
		}

		ci := tt.Info.(*typ.CompType)

		for _, f := range e.Fields {
			f.Val.Accept(t)

			cf, ok := ci.Fields[f.Tok.Lit]

			if !ok {
				panic("unknown field")
			}

			if len(f.Val.Type()) > 1 {
				panic("assignment disbalance")
			}

			if !cf.Type.Equal(f.Val.Type()[0]) {
				panic("type mismatch")
			}
		}
	case *IdfExpr:
		obj, _ := t.env.LookupObj(e.Tok.Lit)

		if obj == nil {
			panic("undefined variable")
		}

		e.Typ = obj.Typ
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

		e.Typ = f.Type
	case *InfExpr:
		e.X.Accept(t)
		e.Y.Accept(t)

		tx := e.X.Type()[0]
		ty := e.Y.Type()[0]

		if tx.Kind == typ.COMP || tx.Kind == typ.FUNC {
			panic("unsupported operand")
		}

		if ty.Kind == typ.COMP || ty.Kind == typ.FUNC {
			panic("unsupported operand")
		}

		if tx.Kind != ty.Kind {
			panic("type mismatch")
		}

		e.Typ = tx
	case *PfxExpr:
		e.X.Accept(t)
		e.Typ = e.X.Type()[0]
	case *CallExpr:
	}
}
