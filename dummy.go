package main

import (
	"os"

	"github.com/paraskun/ess-go/ast"
)

func main() {
	b, _ := os.ReadFile("test.ess")
	o, _ := os.OpenFile("test.dbg", os.O_CREATE|os.O_RDWR, 0644)
	p := ast.Parse([]rune(string(b)))

	ast.Typeset(p)
	ast.Assemble(p).Debug(o)

	o.Close()
}
