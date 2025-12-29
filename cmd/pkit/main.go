package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	buildSource = "source"
)

var rootCmd = &cobra.Command{
	Use:   "pkit",
	Short: "pkit - Multi-source AI prompt bookmark manager",
	Long: `pkit is a CLI tool for subscribing to, organizing, and using AI prompts from multiple sources.

It aggregates prompts from GitHub repositories (Fabric, awesome-chatgpt-prompts, etc.),
provides full-text search, interactive browsing, and seamless piping to execution tools
like claude, llm, fabric, and mods.`,
	Version:                    version,
	SilenceUsage:               true,
	SilenceErrors:              true,
	DisableFlagParsing:         false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 2,
}

func init() {
	// Set custom help function to show subcommands (only for root command)
	rootCmd.SetHelpFunc(customHelp)

	// Global flags will be added here
	// rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	// rootCmd.PersistentFlags().Bool("debug", false, "debug output")
}

func customHelp(cmd *cobra.Command, args []string) {
	// Only use custom help for root command
	if cmd.Name() != "pkit" {
		// For subcommands, show full help with Long description
		if cmd.Long != "" {
			fmt.Fprintf(os.Stdout, "%s\n\n", cmd.Long)
		} else if cmd.Short != "" {
			fmt.Fprintf(os.Stdout, "%s\n\n", cmd.Short)
		}
		fmt.Fprint(os.Stdout, cmd.UsageString())
		return
	}

	fmt.Fprintf(os.Stdout, "%s\n\n", cmd.Long)
	fmt.Fprintf(os.Stdout, "Usage:\n  %s [command]\n\n", cmd.Use)

	// Define command order: important commands first, utilities last
	commandOrder := []string{
		"subscribe",
		"search",
		"find",
		"get",
		"show",
		"serve",
		"bookmark",
		"alias",
		"tag",
		"status",
		"upgrade",
		"reindex",
		"help",
		"completion",
	}

	// Create a map for quick lookup
	cmdMap := make(map[string]*cobra.Command)
	for _, c := range cmd.Commands() {
		cmdMap[c.Name()] = c
	}

	fmt.Fprintln(os.Stdout, "Available Commands:")

	// Print commands in defined order
	for _, name := range commandOrder {
		c, exists := cmdMap[name]
		if !exists || c.Hidden {
			continue
		}

		// Print parent command
		fmt.Fprintf(os.Stdout, "  %-15s %s\n", c.Name(), c.Short)

		// Print subcommands if they exist
		if c.HasSubCommands() {
			for _, sub := range c.Commands() {
				if !sub.Hidden {
					fmt.Fprintf(os.Stdout, "    %-13s %s\n", sub.Name(), sub.Short)
				}
			}
		}
	}

	fmt.Fprintln(os.Stdout, "\nFlags:")
	fmt.Fprintln(os.Stdout, "  -h, --help      help for pkit")
	fmt.Fprintln(os.Stdout, "  -v, --version   version for pkit")
	fmt.Fprintln(os.Stdout, "\nUse \"pkit [command] --help\" for more information about a command.")
}

func main() {
	// Shorthand resolution: if first arg is not a known command, treat it as "get <arg>"
	if len(os.Args) > 1 {
		firstArg := os.Args[1]

		// Check if it's a known command or flag
		isKnownCommand := false
		if firstArg == "--help" || firstArg == "-h" || firstArg == "--version" || firstArg == "-v" {
			isKnownCommand = true
		} else {
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == firstArg || cmd.HasAlias(firstArg) {
					isKnownCommand = true
					break
				}
			}
		}

		// If not a known command, prepend "get" to the args
		if !isKnownCommand {
			os.Args = append([]string{os.Args[0], "get"}, os.Args[1:]...)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
