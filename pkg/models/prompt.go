package models

import "time"

// Prompt represents a parsed prompt from a source.
// Full implementation will be added in Phase 2 (Foundation).
type Prompt struct {
	ID          string    `json:"id"`
	SourceID    string    `json:"source_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	FilePath    string    `json:"file_path"`
	Tags        []string  `json:"tags"`
	UpdatedAt   time.Time `json:"updated_at"`
}
