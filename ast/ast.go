package ast

type Node interface {
	Debug() string
}

type Stmt interface {
	Node
}

type Expr interface {
	Node

	Type() ValueType
}

type Block struct {
	Scope []struct {
		Beg int
		End int
	}

	Body []Stmt
}

type Loop struct {
	Block

	Cond Expr
}

type Branch struct {
	Cond Expr

	Pos Block
	Neg Block
}

type Call struct {
	Name string
	Args []Expr
}
