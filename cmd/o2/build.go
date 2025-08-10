package main

import (
	"fmt"

	"github.com/paraskun/o2/typ/mod"
	"github.com/spf13/cobra"
)

var build = &cobra.Command{
	Use:   "build [package]",
	Short: "Build compiles specified package.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one argument")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mod := &mod.Module{
			Version: mod.Version{},
		}

		if err := mod.Load("."); err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", mod)

		pkg := mod.Lookup(args[0])

		if pkg == nil {
			panic("no such package in context")
		}

		fmt.Printf("%+v\n", pkg)

		return nil
	},
}
