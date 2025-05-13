package ast

import (
	"bytes"

	"github.com/paraskun/ess-go/run"
	"github.com/paraskun/ess-go/typ"
)

func Assemble(p *Pragma) *run.Pragma {
	a := assembler{
		env: p.Env,
		obj: &run.Pragma{},
	}

	p.Accept(&a)

	return a.obj
}

type assembler struct {
	env *typ.Env
	obj *run.Pragma
	buf *bytes.Buffer
}

func (a *assembler) VisitDecl(u Decl) {
	switch d := u.(type) {
	case *FuncDecl:
		img := run.FuncImage{}

		img.
	}
}

func (a *assembler) VisitStmt(u Stmt) {
	switch s := u.(type) {
	case *BlockStmt:
	case *AssignStmt:
	case *LoopStmt:
	case *CondStmt:
	case *ReturnStmt:
	}
}

func (a *assembler) VisitExpr(u Expr) {
	switch e := u.(type) {
	case *BaseImmExpr:
	case *CompImmExpr:
	case *IdfExpr:
	case *DotExpr:
	case *InfExpr:
	case *PfxExpr:
	case *CallExpr:
	case *ToSigExpr:
	case *ToUnsExpr:
	case *ToFltExpr:
	}
}
