package ast

type ValueType int

const (
	Int ValueType = iota
	Flt
	Bol
	Str
	Obj
	Arr
	Ref
)

type Value struct {
	Type ValueType
}

type Integer struct {
	Value int
}

type Float struct {
	Value float64
}

type Boolean struct {
	Value bool
}

type String struct {
	Value string
}

type Object struct {
	Value []Value
}

type Array struct {
	Value []Value
}

type Reference struct {
	*Value
}
