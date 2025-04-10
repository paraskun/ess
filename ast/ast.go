package ast

type Func struct {
	Name string
	Data []Value
	Body []Stmt

	// map id to Value
}

type Stmt struct {
}

type Loop struct {
	Stmt

	Body []Stmt
}

type Expr struct {
}

type Cond struct {
	Expr
}

type Call struct {
	Expr

	Name string
	Args []Expr
}

type Const struct {
	Expr
}
