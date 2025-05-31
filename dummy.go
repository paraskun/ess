package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"unsafe"

	"github.com/paraskun/ess-go/ast"
	"github.com/paraskun/ess-go/run"
)

func main() {
	b, _ := os.ReadFile("test.ess")
	o, _ := os.OpenFile("test.dbg", os.O_CREATE|os.O_RDWR, 0644)

	pkg := ast.Parse([]rune(string(b)))
	ast.Typeset(pkg)
	exe := ast.Assemble(pkg)

	mm := run.Machine{}
	wm := make(map[uint64]*run.Machine)
	eb := &bytes.Buffer{}
	ab := &bytes.Buffer{}

	mm.Load(exe, "map")

	for i := range 10 {
		binary.Write(eb, binary.LittleEndian, uint64(i))  // dev
		binary.Write(eb, binary.LittleEndian, float64(i)) // val

		event := eb.Bytes()

		binary.Write(ab, binary.LittleEndian, uint64(uintptr(unsafe.Pointer(&event[0])))) // ptr
		mm.Exec(ab.Bytes())

		dev := mm.LoadU64()

		if vm, ok := wm[dev]; !ok {
			vm := &run.Machine{}

			vm.Load(exe, "reduce")
			vm.Exec(ab.Bytes())

			wm[dev] = vm
		} else {
			vm.Exec(ab.Bytes())
		}

		eb.Reset()
		ab.Reset()
	}

	event := buf.Bytes()
	buf = &bytes.Buffer{}

	m.Exec(buf.Bytes())

	println(m.LoadU64())
	// for i := range 10 {
	// 	arg := binary.LittleEndian.AppendUint64([]byte{}, uint64(i))
	// 	m.Exec(arg)
	// }

	o.Close()
}
