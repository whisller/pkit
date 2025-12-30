package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/web"
)

var serveCmd = &cobra.Command{
	Use:   "web",
	Short: "Start local web server",
	Long: `Start a local web server to browse prompts in your browser.

The web server provides the same functionality as 'pkit find' but in a
browser-based interface. It binds only to localhost (127.0.0.1) for security.

Features:
  - Browse and search prompts from all subscribed sources
  - Filter by source, tags, and bookmarked status
  - View full prompt details
  - Manage bookmarks and tags
  - Copy prompts to clipboard

Examples:
  # Start server on default port 8080
  pkit web

  # Start server on custom port
  pkit web --port 3000

  # Start server and specify custom index location
  pkit web --port 8080

The server will display a message with the URL to open in your browser.
Press Ctrl+C to stop the server.`,
	RunE: runServe,
}

var servePort int

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Get config to check if sources exist
	_, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get index path (same as reindex command)
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	indexPath := filepath.Join(indexBasePath, "prompts.bleve")

	// Check if index exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("index not found at %s. Please run 'pkit reindex' first", indexPath)
	}

	// Create server
	server, err := web.NewServer(servePort, indexPath)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived shutdown signal...")
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		fmt.Println("Server stopped gracefully")
		return nil

	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}
