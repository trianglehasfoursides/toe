package main

import (
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "toe",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
