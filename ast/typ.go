package ast

import (
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
		re := make([]*typ.Type, 0)

		for _, i := range v.Val {
			i.Accept(t)
			re = append(re, i.Type()...)
		}

		if len(v.Var) != len(re) {
			panic("assignment disbalance")
		}

		if v.New {
			for i, j := range v.Var {
				idf := j.(*IdfExpr)

				if obj, lvl := t.env.LookupObj(idf.Tok.Lit); obj != nil && lvl != 0 {
					panic("name collision")
				}

				t.env.Obj[idf.Tok.Lit] = &typ.Object{
					Typ: re[i],
					Env: t.env,
				}
			}

			break
		}

		for i, j := range v.Var {
			if _, ok := j.(*CallExpr); ok {
				panic("could not assign to function call")
			}

			j.Accept(t)

			if !j.Type()[0].Equal(re[i]) {
				panic("type mismatch")
			}
		}
	case *ReturnStmt:
		re := make([]*typ.Type, 0)

		for _, i := range v.Arg {
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

func (c *typer) VisitDecl(d Decl)
func (t *typer) VisitExpr(e Expr)
