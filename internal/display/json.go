package display

import (
	"encoding/json"
	"io"

	"github.com/whisller/pkit/pkg/models"
)

// PromptJSON represents the JSON output format for a prompt.
type PromptJSON struct {
	ID          string   `json:"id"`
	SourceID    string   `json:"source_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags,omitempty"`
	Author      string   `json:"author,omitempty"`
	Version     string   `json:"version,omitempty"`
	FilePath    string   `json:"file_path,omitempty"`
}

// PrintPromptJSON outputs the prompt as formatted JSON.
func PrintPromptJSON(w io.Writer, prompt *models.Prompt) error {
	output := PromptJSON{
		ID:          prompt.ID,
		SourceID:    prompt.SourceID,
		Name:        prompt.Name,
		Description: prompt.Description,
		Content:     prompt.Content,
		Tags:        prompt.Tags,
		Author:      prompt.Author,
		Version:     prompt.Version,
		FilePath:    prompt.FilePath,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
