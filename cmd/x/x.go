package main

import (
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "x",
	Short: "X is a tool for managing X modules.",
}

func init() {
	root.AddCommand(build)
	root.AddCommand(run)
}

func main() {
	dir, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	dir += "/xlang"

	if err := os.MkdirAll(dir+"/src", 0774); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(dir+"/pkg", 0774); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(dir+"/lib", 0774); err != nil {
		panic(err)
	}

	root.Execute()
}
