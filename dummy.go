package main

import (
	"os"

	"github.com/paraskun/ess-go/ast"
	"github.com/paraskun/ess-go/run"
)

func main() {
	b, _ := os.ReadFile("test.ess")
	o, _ := os.OpenFile("test.dbg", os.O_CREATE|os.O_RDWR, 0644)
	p := ast.Parse([]rune(string(b)))

	ast.Typeset(p)

	e := ast.Assemble(p)
	m := run.Machine{}

	e.Debug(o)
	m.Load(e)
	m.Exec()

	o.Close()
}
