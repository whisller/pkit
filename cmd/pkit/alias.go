package main

import (
	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage aliases",
	Long:  `Add, remove, and list aliases for prompts.`,
}

func init() {
	rootCmd.AddCommand(aliasCmd)
}
