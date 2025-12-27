package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for prompts across all subscribed sources",
	Long: `Search for prompts using keyword search across all subscribed sources.

Returns results in a table format showing source, name, and description.
Supports filtering by source and tags.

Examples:
  pkit search "code review"
  pkit search "summarize" --source fabric
  pkit search "python" --tag dev
  pkit search "debug" --format json`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

var (
	searchSource string
	searchTags   []string
	searchFormat string
	searchLimit  int
	searchFuzzy  bool
)

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchSource, "source", "", "Filter by source ID")
	searchCmd.Flags().StringSliceVar(&searchTags, "tag", []string{}, "Filter by tags (can specify multiple)")
	searchCmd.Flags().StringVar(&searchFormat, "format", "table", "Output format (table, json)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 50, "Maximum number of results")
	searchCmd.Flags().BoolVar(&searchFuzzy, "fuzzy", false, "Enable fuzzy matching")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	// Load configuration to check if sources exist
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if any sources are subscribed
	if len(cfg.Sources) == 0 {
		return fmt.Errorf(`No sources subscribed

Subscribe to sources first:
  pkit subscribe fabric/patterns
  pkit subscribe f/awesome-chatgpt-prompts

Then search:
  pkit search "code review"`)
	}

	// Get index path
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	// Open index
	indexPath := filepath.Join(indexBasePath, "prompts.bleve")
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer indexer.Close()

	// Build search options
	searchOpts := index.SearchOptions{
		Query:      query,
		MaxResults: searchLimit,
		SourceID:   searchSource,
		Tags:       searchTags,
		Fuzzy:      searchFuzzy,
	}

	// Execute search
	results, err := indexer.Search(searchOpts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Handle empty results
	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "No results found")
		return nil
	}

	// Output results based on format
	switch searchFormat {
	case "json":
		return outputJSON(results)
	case "table":
		return outputTable(results)
	default:
		return fmt.Errorf("unknown format: %s (supported: table, json)", searchFormat)
	}
}

func outputTable(results []index.SearchResult) error {
	// Print header
	fmt.Println("┌─────────────┬──────────────────┬────────────────────────────────────┐")
	fmt.Println("│ SOURCE      │ NAME             │ DESCRIPTION                         │")
	fmt.Println("├─────────────┼──────────────────┼────────────────────────────────────┤")

	// Print results
	for _, result := range results {
		source := truncateString(result.Prompt.SourceID, 11)
		name := truncateString(result.Prompt.Name, 16)
		desc := truncateString(result.Prompt.Description, 36)

		fmt.Printf("│ %-11s │ %-16s │ %-36s │\n", source, name, desc)
	}

	// Print footer
	fmt.Println("└─────────────┴──────────────────┴────────────────────────────────────┘")

	// Print summary
	fmt.Fprintf(os.Stderr, "\nFound %d results\n", len(results))

	return nil
}

func outputJSON(results []index.SearchResult) error {
	// Convert to JSON-friendly structure
	type jsonPrompt struct {
		ID          string   `json:"id"`
		SourceID    string   `json:"source_id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Author      string   `json:"author,omitempty"`
		FilePath    string   `json:"file_path"`
		Score       float64  `json:"score"`
	}

	type jsonOutput struct {
		Query   string       `json:"query"`
		Count   int          `json:"count"`
		Prompts []jsonPrompt `json:"prompts"`
	}

	output := jsonOutput{
		Count:   len(results),
		Prompts: make([]jsonPrompt, len(results)),
	}

	for i, result := range results {
		output.Prompts[i] = jsonPrompt{
			ID:          result.Prompt.ID,
			SourceID:    result.Prompt.SourceID,
			Name:        result.Prompt.Name,
			Description: result.Prompt.Description,
			Tags:        result.Prompt.Tags,
			Author:      result.Prompt.Author,
			FilePath:    result.Prompt.FilePath,
			Score:       result.Score,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// truncateString truncates a string to maxLen, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
