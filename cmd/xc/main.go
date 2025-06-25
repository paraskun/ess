package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/paraskun/ess-go/ast"
	"google.golang.org/protobuf/proto"

	_ "github.com/paraskun/ess-go/img"
)

var dbg = flag.Bool("d", false, "enable debug output")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: essc {src}\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	nsrc := flag.Arg(0)

	if !strings.HasSuffix(nsrc, ".ess") {
		panic("unsupported source format")
	}

	src, err := os.ReadFile(nsrc)

	if err != nil {
		panic(fmt.Errorf("could not read source: %v", err))
	}

	pkt := ast.Parse([]rune(string(src)))
	msg := ast.Assemble(pkt)

	nout := strings.Replace(nsrc, ".ess", ".binpb", 1)
	bpb, err := proto.Marshal(msg)

	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(nout, bpb, os.ModeAppend); err != nil {
		panic(err)
	}
}
