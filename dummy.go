package main

import (
	"encoding/binary"
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

	for i := range 10 {
		arg := binary.LittleEndian.AppendUint64([]byte{}, uint64(i))
		m.Exec(arg)
	}

	o.Close()
}
