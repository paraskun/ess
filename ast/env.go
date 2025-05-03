package ast

type Env struct {
	*Env

	Map map[string]int
	Sym []TypeSpec
}

type TypeSpec interface {
	typeSpec()
}

type Bool struct {
	Const bool
}

func (*Bool) typeSpec() {}

type Sig64 struct {
	Const int64
}

func (*Sig64) typeSpec() {}
