package sir

import (
	"github.com/paraskun/ess-go/ast"
	"github.com/paraskun/ess-go/typ"
)

type builder struct {
	env *typ.Env
	src []Ins
	tmp uint
}

func (*builder) VisitStmt(s ast.Stmt)
func (*builder) VisitExpr(e ast.Expr)
func (*builder) VisitDecl(d ast.Decl)

func Build(p *ast.Pragma) (*typ.Env, []Ins) {
	b := &builder{
		env: &typ.Env{},
	}

	for _, s := range p.Body {
		s.Accept(b)
	}

	return b.env, b.src
}
