package sir

import (
	"github.com/paraskun/ess-go/ast"
	"github.com/paraskun/ess-go/typ"
)

type builder struct {
	env *typ.Env
	src []Ins
}

func (*builder) VisitStmt(s ast.Stmt)
func (*builder) VisitExpr(e ast.Expr)
func (*builder) VisitDecl(d ast.Decl)

func Build(p *ast.Pragma) (*typ.Env, []Ins) {
	b := &builder{
		env: p.Env,
	}

	for _, d := range p.Dec {
		d.Accept(b)
	}

	return b.env, b.src
}
