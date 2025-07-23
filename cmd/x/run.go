package main

import "github.com/spf13/cobra"

var run = &cobra.Command{
	Use:   "run",
	Short: "Run compiles (if necessary) and executes specified package.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
