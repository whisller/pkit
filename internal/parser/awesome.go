package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

// AwesomeChatGPTParser parses prompts from f/awesome-chatgpt-prompts repository.
// Prompts are stored in a prompts.csv file with columns: act, prompt, for_devs, type, contributor.
type AwesomeChatGPTParser struct{}

// NewAwesomeChatGPTParser creates a new awesome-chatgpt-prompts parser instance.
func NewAwesomeChatGPTParser() *AwesomeChatGPTParser {
	return &AwesomeChatGPTParser{}
}

// Name returns the parser name.
func (p *AwesomeChatGPTParser) Name() string {
	return "awesome_chatgpt"
}

// CanParse checks if the source path contains prompts.csv file.
func (p *AwesomeChatGPTParser) CanParse(sourcePath string) bool {
	csvPath := filepath.Join(sourcePath, "prompts.csv")
	_, err := os.Stat(csvPath)
	return err == nil
}

// ParsePrompts extracts all prompts from prompts.csv file.
func (p *AwesomeChatGPTParser) ParsePrompts(source *models.Source) ([]models.Prompt, error) {
	var prompts []models.Prompt

	csvPath := filepath.Join(source.LocalPath, "prompts.csv")

	// Check if CSV file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("prompts.csv not found: %w", err)
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open prompts.csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Get file mod time for UpdatedAt
	fileInfo, _ := os.Stat(csvPath)
	updatedAt := time.Now()
	if fileInfo != nil {
		updatedAt = fileInfo.ModTime()
	}

	// Read rows
	rowNum := 1 // Start at 1 (header is row 0)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Log warning but continue with other rows
			fmt.Fprintf(os.Stderr, "Warning: failed to read CSV row %d: %v\n", rowNum, err)
			rowNum++
			continue
		}

		// Ensure row has enough columns
		if len(row) < 5 {
			fmt.Fprintf(os.Stderr, "Warning: skipping row %d with insufficient columns\n", rowNum)
			rowNum++
			continue
		}

		// Parse CSV row
		act := row[0]        // Column: act
		promptText := row[1] // Column: prompt
		forDevs := row[2]    // Column: for_devs
		promptType := row[3] // Column: type
		contributor := row[4] // Column: contributor

		// Derive metadata
		name := slugify(act)                     // "Linux Terminal" → "linux-terminal"
		description := truncate(promptText, 150) // First 150 chars of prompt

		// Build tags from metadata
		tags := []string{}
		if strings.ToUpper(forDevs) == "TRUE" {
			tags = append(tags, "dev")
		}
		if promptType != "" && promptType != "TEXT" {
			tags = append(tags, strings.ToLower(promptType))
		}

		prompt := models.Prompt{
			ID:          fmt.Sprintf("%s:%s", source.ID, name),
			SourceID:    source.ID,
			Name:        name,
			Content:     promptText,
			Description: description,
			Tags:        tags,
			Author:      contributor,
			Version:     "",
			FilePath:    "prompts.csv",
			Metadata: map[string]interface{}{
				"act":      act,
				"for_devs": forDevs,
				"type":     promptType,
			},
			IndexedAt: time.Now(),
			UpdatedAt: updatedAt,
		}

		prompts = append(prompts, prompt)
		rowNum++
	}

	if len(prompts) == 0 {
		return nil, fmt.Errorf("no prompts found in %s", csvPath)
	}

	return prompts, nil
}

// slugify converts "Linux Terminal" → "linux-terminal"
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9-]+`)
	return re.ReplaceAllString(s, "")
}

// truncate shortens string to max length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
