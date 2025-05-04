package ast

type Env struct {
	*Env

	Map map[string]int
	Sym []TypeSpec
}

func (e *Env) Lookup(name string) TypeSpec {
	cur := e

	for cur != nil {
		if ix, ok := e.Map[name]; ok {
			return e.Sym[ix]
		}

		cur = cur.Env
	}

	return nil
}

type TypeSpec interface {
	typeSpec()
}

type Local struct {
	Off uint32
}

type Bool struct {
	Local

	Const bool
}

type Sig64 struct {
	Local

	Const int64
}

func (*Bool) typeSpec()  {}
func (*Sig64) typeSpec() {}
