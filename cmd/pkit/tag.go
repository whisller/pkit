package main

import (
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags for prompts",
	Long:  `Add, remove, and list tags for prompts.`,
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
