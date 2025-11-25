package main

import (
	"fmt"
	"os"

	"github.com/paraskun/o2/ast"
	"github.com/paraskun/o2/typ/mod"
	"github.com/spf13/cobra"
)

var build = &cobra.Command{
	Use:   "build [package]",
	Short: "Build compiles specified package with it's dependencies.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()

		if err != nil {
			panic(err)
		}

		m, err := mod.Load(dir)

		if err != nil {
			return err
		}

		name := m.Mod.Name

		if args[0] != "." {
			name += "/" + args[0]
		}

		p := m.Lookup(name)

		if p == nil {
			return fmt.Errorf("no such package current module")
		}

		if !ast.Parse(p) {
			return nil
		}

		if !ast.Typeset(p) {
			return nil
		}

		fmt.Printf("%#v\n", p.Env)
		fmt.Printf("%#v\n", ast.Assemble(p))

		return nil
	},
}
