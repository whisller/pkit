package main

import (
	"github.com/spf13/cobra"
)

var bookmarkCmd = &cobra.Command{
	Use:   "bookmark",
	Short: "Manage bookmarks",
	Long:  `Add, remove, list, and tag bookmarked prompts.`,
}

func init() {
	rootCmd.AddCommand(bookmarkCmd)
}
