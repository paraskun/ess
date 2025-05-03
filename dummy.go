package main

import (
	"os"

	"github.com/paraskun/ess-go/ast"
	"github.com/paraskun/ess-go/lex"
)

func main() {
	dat, err := os.ReadFile(os.Args[1])

	if err != nil {
		panic(err)
	}

	p := ast.Parser{
		S: lex.Scanner{},
	}

	p.S.Load([]rune(string(dat)))
	p.Parse().Debug(os.Stdout)
}
