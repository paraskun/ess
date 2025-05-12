package ast

import "io"

func Assemble(p *Pragma, w io.Writer) {}

type assembler struct{}

func (*assembler) VisitStmt(s Stmt) {}
func (*assembler) VisitDecl(d Decl) {}
func (*assembler) VisitExpr(e Expr) {}
