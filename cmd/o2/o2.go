package main

import (
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "o2",
	Short: "O2 is a tool for managing O2 modules.",
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

	dir += "/o2"

	if err := os.MkdirAll(dir+"/mod", 0774); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(dir+"/co2", 0774); err != nil {
		panic(err)
	}

	root.Execute()
}
