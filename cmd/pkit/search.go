package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/tag"
	"github.com/whisller/pkit/pkg/models"
	"golang.org/x/term"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for prompts across all subscribed sources",
	Long: `Search for prompts using keyword search across all subscribed sources.

Returns results in a table format showing ID, description, user tags, and bookmark status.
Bookmarked prompts are shown first in the results.
Supports filtering by source, tags, and bookmark status.

Examples:
  pkit search "code review"
  pkit search "summarize" --source fabric
  pkit search "python" --tag dev
  pkit search "debug" --format json
  pkit search "review" --bookmarked        # Show only bookmarked prompts
  pkit search "code" -b                    # Short flag for bookmarked
  pkit search "review" --content           # Include content preview in table
  pkit search "review" -c --format json    # Include full content in JSON`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

var (
	searchSource     string
	searchTags       []string
	searchFormat     string
	searchLimit      int
	searchFuzzy      bool
	searchBookmarked bool
	searchContent    bool
)

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchSource, "source", "", "Filter by source ID")
	searchCmd.Flags().StringSliceVar(&searchTags, "tag", []string{}, "Filter by tags (can specify multiple)")
	searchCmd.Flags().StringVar(&searchFormat, "format", "table", "Output format (table, json)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 50, "Maximum number of results")
	searchCmd.Flags().BoolVar(&searchFuzzy, "fuzzy", false, "Enable fuzzy matching")
	searchCmd.Flags().BoolVarP(&searchBookmarked, "bookmarked", "b", false, "Show only bookmarked prompts")
	searchCmd.Flags().BoolVarP(&searchContent, "content", "c", false, "Include full prompt content in results")
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

	// Load bookmarks and tags for enriching results
	bookmarkManager := bookmark.NewManager()
	bookmarks, err := bookmarkManager.ListBookmarks()
	if err != nil {
		// Non-fatal: just log and continue without bookmark info
		fmt.Fprintf(os.Stderr, "Warning: failed to load bookmarks: %v\n", err)
		bookmarks = []models.Bookmark{}
	}

	tagManager := tag.NewManager()
	allTags, err := tagManager.ListAllTags()
	if err != nil {
		// Non-fatal: just log and continue without tag info
		fmt.Fprintf(os.Stderr, "Warning: failed to load tags: %v\n", err)
		allTags = []models.PromptTags{}
	}

	// Create lookup maps
	bookmarkMap := make(map[string]bool)
	for _, bm := range bookmarks {
		bookmarkMap[bm.PromptID] = true
	}

	tagMap := make(map[string][]string)
	for _, pt := range allTags {
		tagMap[pt.PromptID] = pt.Tags
	}

	// Filter by bookmarked if flag is set
	if searchBookmarked {
		filteredResults := []index.SearchResult{}
		for _, result := range results {
			if bookmarkMap[result.Prompt.ID] {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults

		// Handle empty results after filtering
		if len(results) == 0 {
			fmt.Fprintln(os.Stderr, "No bookmarked prompts found matching the search criteria")
			return nil
		}
	}

	// Sort results: bookmarked prompts first
	sort.Slice(results, func(i, j int) bool {
		iBookmarked := bookmarkMap[results[i].Prompt.ID]
		jBookmarked := bookmarkMap[results[j].Prompt.ID]

		if iBookmarked != jBookmarked {
			return iBookmarked // bookmarked comes first
		}

		// If both bookmarked or both not, sort by score
		return results[i].Score > results[j].Score
	})

	// Load content from files if requested
	contentMap := make(map[string]string)
	if searchContent {
		// Get home directory to construct full file paths
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		for _, result := range results {
			// Read content from file
			// FilePath is relative to ~/.pkit/
			fullPath := filepath.Join(homeDir, ".pkit", result.Prompt.FilePath)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				// Non-fatal: just log and continue without content
				fmt.Fprintf(os.Stderr, "Warning: failed to read content for %s: %v\n", result.Prompt.ID, err)
				continue
			}
			contentMap[result.Prompt.ID] = string(content)
		}
	}

	// Output results based on format
	switch searchFormat {
	case "json":
		return outputJSON(results, bookmarkMap, tagMap, contentMap)
	case "table":
		return outputTable(results, bookmarkMap, tagMap, contentMap)
	default:
		return fmt.Errorf("unknown format: %s (supported: table, json)", searchFormat)
	}
}

func outputTable(results []index.SearchResult, bookmarkMap map[string]bool, tagMap map[string][]string, contentMap map[string]string) error {
	// Get terminal width
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || termWidth < 60 {
		termWidth = 120 // Default fallback
	}

	table := tablewriter.NewWriter(os.Stdout)

	// Configure table to fit terminal width with text wrapping
	table.Options(
		tablewriter.WithTableMax(termWidth),
		tablewriter.WithRowAutoWrap(1),
	)

	// Set header based on whether content is included
	includeContent := len(contentMap) > 0
	if includeContent {
		table.Header("ID", "DESCRIPTION", "TAGS", "CONTENT PREVIEW")
	} else {
		table.Header("ID", "DESCRIPTION", "TAGS")
	}

	// Add rows
	hasBookmarks := false
	for _, result := range results {
		// Get tags for this prompt
		tags := tagMap[result.Prompt.ID]
		tagsStr := ""
		if len(tags) > 0 {
			tagsStr = strings.Join(tags, ", ")
		}

		// Prepend [*] to ID if bookmarked
		id := result.Prompt.ID
		if bookmarkMap[result.Prompt.ID] {
			id = "[*]" + id
			hasBookmarks = true
		}

		// Build row data
		if includeContent {
			content := contentMap[result.Prompt.ID]
			// Truncate content to 80 chars for preview
			contentPreview := content
			if len(content) > 80 {
				contentPreview = content[:80] + "..."
			}
			table.Append(id, result.Prompt.Description, tagsStr, contentPreview)
		} else {
			table.Append(id, result.Prompt.Description, tagsStr)
		}
	}

	// Render the table
	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	// Print legend if there are bookmarked results
	if hasBookmarks {
		fmt.Fprintln(os.Stderr, "\n[*] = Bookmarked")
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "Found %d results\n", len(results))

	return nil
}

func outputJSON(results []index.SearchResult, bookmarkMap map[string]bool, tagMap map[string][]string, contentMap map[string]string) error {
	// Convert to JSON-friendly structure
	type jsonPrompt struct {
		ID          string   `json:"id"`
		SourceID    string   `json:"source_id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		UserTags    []string `json:"user_tags"`
		Bookmarked  bool     `json:"bookmarked"`
		Author      string   `json:"author,omitempty"`
		FilePath    string   `json:"file_path"`
		Content     string   `json:"content,omitempty"`
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
		userTags := tagMap[result.Prompt.ID]
		if userTags == nil {
			userTags = []string{}
		}

		prompt := jsonPrompt{
			ID:          result.Prompt.ID,
			SourceID:    result.Prompt.SourceID,
			Name:        result.Prompt.Name,
			Description: result.Prompt.Description,
			Tags:        result.Prompt.Tags,
			UserTags:    userTags,
			Bookmarked:  bookmarkMap[result.Prompt.ID],
			Author:      result.Prompt.Author,
			FilePath:    result.Prompt.FilePath,
			Score:       result.Score,
		}

		// Add content if available
		if content, ok := contentMap[result.Prompt.ID]; ok {
			prompt.Content = content
		}

		output.Prompts[i] = prompt
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
