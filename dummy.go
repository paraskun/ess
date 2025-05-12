package main

import (
	"os"

	"github.com/paraskun/ess-go/ast"
)

func main() {
	b, _ := os.ReadFile("test.ess")
	p := ast.Parse([]rune(string(b)))

	ast.Typeset(p)
	ast.Assemble(p, os.Stdout)
}
