package display

import (
	"fmt"
	"io"

	"github.com/whisller/pkit/pkg/models"
)

// PrintPromptText outputs ONLY the prompt content to stdout.
// This is designed for piping to execution tools like claude, llm, fabric, mods.
// NO headers, NO metadata, NO formatting - just the raw prompt text.
func PrintPromptText(w io.Writer, prompt *models.Prompt) error {
	_, err := fmt.Fprint(w, prompt.Content)
	return err
}

// PrintPromptWithMetadata outputs the prompt with human-readable metadata.
// This is for the "show" command, not for piping.
func PrintPromptWithMetadata(w io.Writer, prompt *models.Prompt) error {
	// Metadata output - errors extremely rare (writing to os.Stdout)
	_, _ = fmt.Fprintf(w, "ID: %s\n", prompt.ID)
	_, _ = fmt.Fprintf(w, "Source: %s\n", prompt.SourceID)
	_, _ = fmt.Fprintf(w, "Name: %s\n", prompt.Name)

	if prompt.Description != "" {
		_, _ = fmt.Fprintf(w, "Description: %s\n", prompt.Description)
	}

	if len(prompt.Tags) > 0 {
		_, _ = fmt.Fprintf(w, "Tags: %s\n", fmt.Sprint(prompt.Tags))
	}

	if prompt.Author != "" {
		_, _ = fmt.Fprintf(w, "Author: %s\n", prompt.Author)
	}

	_, _ = fmt.Fprintln(w, "\n--- Content ---")
	_, _ = fmt.Fprintln(w, prompt.Content)

	return nil
}
