package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/source"
	"github.com/whisller/pkit/pkg/models"
)

var subscribeCmd = &cobra.Command{
	Use:   "subscribe <source> [sources...]",
	Short: "Subscribe to a GitHub repository as a prompt source",
	Long: `Subscribe to one or more GitHub repositories as prompt sources.

The source can be specified in short form (org/repo) or as a full URL.

Examples:
  pkit subscribe fabric/patterns
  pkit subscribe https://github.com/f/awesome-chatgpt-prompts
  pkit subscribe fabric/patterns f/awesome-chatgpt-prompts  # Multiple sources`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSubscribe,
}

var (
	subscribeName    string
	subscribeID      string
	subscribeFormat  string
	subscribeVerbose bool
	subscribeDebug   bool
)

func init() {
	rootCmd.AddCommand(subscribeCmd)

	subscribeCmd.Flags().StringVar(&subscribeName, "name", "", "Custom display name for the source")
	subscribeCmd.Flags().StringVar(&subscribeID, "id", "", "Custom ID for the source")
	subscribeCmd.Flags().StringVar(&subscribeFormat, "format", "", "Force specific parser format (fabric_pattern, awesome_chatgpt, markdown)")
	subscribeCmd.Flags().BoolVarP(&subscribeVerbose, "verbose", "v", false, "Show detailed progress and git operations")
	subscribeCmd.Flags().BoolVar(&subscribeDebug, "debug", false, "Show full trace including timing information")
}

func runSubscribe(cmd *cobra.Command, args []string) (err error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil && (subscribeVerbose || subscribeDebug) {
		fmt.Fprintf(os.Stderr, "→ Warning: failed to get GitHub token: %v\n", err)
		fmt.Fprintln(os.Stderr, "→ Proceeding without authentication")
	}

	// Create source manager
	mgr := source.NewManager(token)

	// Get index path
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	// Create/open index
	indexPath := filepath.Join(indexBasePath, "prompts.bleve")
	if err := index.EnsureIndexPath(indexPath); err != nil {
		return fmt.Errorf("failed to ensure index path: %w", err)
	}

	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return fmt.Errorf("failed to initialize index: %w", err)
	}
	defer func() {
		if closeErr := indexer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close index: %w", closeErr)
		}
	}()

	// Process multiple sources
	if len(args) > 1 {
		return subscribeMultipleSources(mgr, indexer, cfg, args)
	}

	// Process single source
	return subscribeSingleSource(mgr, indexer, cfg, args[0])
}

func subscribeSingleSource(mgr *source.Manager, indexer *index.Indexer, cfg *models.Config, sourceArg string) error {
	startTime := time.Now()

	// Parse source URL
	url, err := parseSourceURL(sourceArg)
	if err != nil {
		return err
	}

	// Derive source ID
	sourceID := subscribeID
	if sourceID == "" {
		sourceID = source.ExtractSourceIDFromURL(url)
	}

	// Check if source already exists
	for _, existingSrc := range cfg.Sources {
		if existingSrc.ID == sourceID {
			return fmt.Errorf(`source '%s' already exists

Existing source:
  URL: %s
  Prompts: %d
  Last indexed: %s

Options:
  - Use different ID: pkit subscribe %s --id %s2
  - Upgrade existing: pkit upgrade %s
  - Remove and re-subscribe: pkit unsubscribe %s && pkit subscribe %s`,
				sourceID, existingSrc.URL, existingSrc.PromptCount,
				existingSrc.LastIndexed.Format("2006-01-02 15:04:05"),
				sourceArg, sourceID, sourceID, sourceID, sourceArg)
		}
	}

	// Determine local path
	sourcesPath, err := config.GetSourcesPath()
	if err != nil {
		return fmt.Errorf("failed to get sources path: %w", err)
	}
	localPath := filepath.Join(sourcesPath, sourceID)

	if subscribeVerbose || subscribeDebug {
		fmt.Fprintf(os.Stderr, "→ Resolving source: %s\n", sourceArg)
		fmt.Fprintf(os.Stderr, "→ Full URL: %s\n", url)
		fmt.Fprintf(os.Stderr, "→ Local path: %s\n", localPath)
	}

	// Clone repository
	if subscribeVerbose || subscribeDebug {
		fmt.Fprintln(os.Stderr, "→ Cloning repository...")
	} else {
		fmt.Fprintln(os.Stderr, "Cloning repository...")
	}

	src, err := mgr.Subscribe(url, localPath)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Override format if specified
	if subscribeFormat != "" {
		src.Format = subscribeFormat
	}

	// Override ID if specified
	if subscribeID != "" {
		src.ID = subscribeID
	}

	// Override name if specified
	if subscribeName != "" {
		src.Name = subscribeName
	} else if src.Name == src.ID {
		// Use a more friendly name if possible
		src.Name = src.ID
	}

	if subscribeVerbose || subscribeDebug {
		fmt.Fprintf(os.Stderr, "→ Detecting format... %s\n", src.Format)
		fmt.Fprintln(os.Stderr, "→ Parsing prompts...")
	}

	// Parse prompts
	p, err := source.GetParser(src.Format)
	if err != nil {
		return fmt.Errorf("failed to get parser: %w", err)
	}

	prompts, err := p.ParsePrompts(src)
	if err != nil {
		return fmt.Errorf("failed to parse prompts: %w", err)
	}

	if subscribeVerbose || subscribeDebug {
		fmt.Fprintf(os.Stderr, "  Found %d prompts\n", len(prompts))
		fmt.Fprintln(os.Stderr, "→ Indexing prompts...")
	} else {
		fmt.Fprintln(os.Stderr, "Indexing prompts...")
	}

	// Index prompts
	if err := indexer.IndexPrompts(prompts); err != nil {
		return fmt.Errorf("failed to index prompts: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[%s] Indexed %d/%d prompts\n", src.ID, len(prompts), len(prompts))

	// Update source metadata
	src.PromptCount = len(prompts)
	src.LastIndexed = time.Now()
	src.SubscribedAt = time.Now()

	// Add source to config
	cfg.Sources = append(cfg.Sources, *src)

	if subscribeVerbose || subscribeDebug {
		fmt.Fprintln(os.Stderr, "→ Saving configuration...")
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if subscribeDebug {
		elapsed := time.Since(startTime)
		fmt.Fprintf(os.Stderr, "✓ Complete in %.1fs\n\n", elapsed.Seconds())
	}

	// Success output
	fmt.Printf("✓ Subscribed to %s\n", src.ID)
	fmt.Printf("  Format: %s\n", src.Format)
	fmt.Printf("  Prompts: %d\n", src.PromptCount)
	fmt.Printf("  Location: %s\n\n", localPath)
	fmt.Printf("Use 'pkit search \"\" --source %s' to see all prompts from this source\n", src.ID)

	return nil
}

func subscribeMultipleSources(mgr *source.Manager, indexer *index.Indexer, cfg *models.Config, sourceArgs []string) error {
	fmt.Fprintf(os.Stderr, "Subscribing to %d sources in parallel...\n", len(sourceArgs))

	// Build source list
	type sourceRequest struct {
		URL       string
		LocalPath string
		SourceID  string
		Arg       string
	}

	var requests []sourceRequest
	for _, arg := range sourceArgs {
		url, err := parseSourceURL(arg)
		if err != nil {
			return err
		}

		sourceID := source.ExtractSourceIDFromURL(url)

		// Check if source already exists
		for _, existingSrc := range cfg.Sources {
			if existingSrc.ID == sourceID {
				return fmt.Errorf("source '%s' already exists (from %s)", sourceID, arg)
			}
		}

		sourcesPath, err := config.GetSourcesPath()
		if err != nil {
			return fmt.Errorf("failed to get sources path: %w", err)
		}
		localPath := filepath.Join(sourcesPath, sourceID)
		requests = append(requests, sourceRequest{
			URL:       url,
			LocalPath: localPath,
			SourceID:  sourceID,
			Arg:       arg,
		})
	}

	// Subscribe to all sources using parallel subscription
	sourcesToSubscribe := make([]struct {
		URL       string
		LocalPath string
	}, len(requests))
	for i, req := range requests {
		sourcesToSubscribe[i].URL = req.URL
		sourcesToSubscribe[i].LocalPath = req.LocalPath
	}

	sources, err := mgr.SubscribeMultiple(sourcesToSubscribe)
	if err != nil {
		return fmt.Errorf("parallel subscription failed: %w", err)
	}

	// Parse and index each source
	for _, req := range requests {
		src := sources[req.SourceID]
		if src == nil {
			fmt.Fprintf(os.Stderr, "[%s] ✗ Failed to subscribe\n", req.SourceID)
			continue
		}

		// Parse prompts
		p, err := source.GetParser(src.Format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] ✗ Failed to get parser: %v\n", src.ID, err)
			continue
		}

		prompts, err := p.ParsePrompts(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] ✗ Failed to parse prompts: %v\n", src.ID, err)
			continue
		}

		// Index prompts
		if err := indexer.IndexPrompts(prompts); err != nil {
			fmt.Fprintf(os.Stderr, "[%s] ✗ Failed to index prompts: %v\n", src.ID, err)
			continue
		}

		// Update source metadata
		src.PromptCount = len(prompts)
		src.LastIndexed = time.Now()
		src.SubscribedAt = time.Now()

		// Add source to config
		cfg.Sources = append(cfg.Sources, *src)

		fmt.Fprintf(os.Stderr, "[%s] Cloning... ✓ %d prompts\n", src.ID, len(prompts))
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintln(os.Stderr, "✓ All sources subscribed successfully")
	return nil
}

// Repository aliases for common prompt sources
var repositoryAliases = map[string]string{
	"fabric":          "danielmiessler/fabric",
	"fabric/patterns": "danielmiessler/fabric",
	"awesome":         "f/awesome-chatgpt-prompts",
	"awesome-chatgpt": "f/awesome-chatgpt-prompts",
}

// parseSourceURL converts short form or full URL to full GitHub URL
func parseSourceURL(source string) (string, error) {
	// Already a full URL
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return source, nil
	}

	// Check if it's a known alias
	if aliasedRepo, ok := repositoryAliases[strings.ToLower(source)]; ok {
		return fmt.Sprintf("https://github.com/%s", aliasedRepo), nil
	}

	// Short form: org/repo
	if strings.Count(source, "/") == 1 {
		parts := strings.Split(source, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return fmt.Sprintf("https://github.com/%s", source), nil
		}
	}

	return "", fmt.Errorf(`invalid source format: "%s"

Expected formats:
  Alias: fabric, awesome (see common aliases)
  Short form: <org>/<repo> (e.g., danielmiessler/fabric)
  Full URL: https://github.com/<org>/<repo>

Common aliases:
  fabric, fabric/patterns → danielmiessler/fabric
  awesome, awesome-chatgpt → f/awesome-chatgpt-prompts`, source)
}
