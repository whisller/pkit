package index

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/whisller/pkit/pkg/models"
)

// Indexer handles creating and maintaining the bleve search index.
type Indexer struct {
	index bleve.Index
	path  string
}

// NewIndexer creates a new indexer instance.
// If the index doesn't exist, it creates a new one.
// If it exists, it opens the existing index.
func NewIndexer(indexPath string) (*Indexer, error) {
	var index bleve.Index

	var err error

	// Check if index already exists
	if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
		// Create new index
		indexMapping := buildIndexMapping()
		index, err = bleve.New(indexPath, indexMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	} else {
		// Open existing index
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open index: %w", err)
		}
	}

	return &Indexer{
		index: index,
		path:  indexPath,
	}, nil
}

// buildIndexMapping creates the bleve index mapping with field boosting.
func buildIndexMapping() mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()

	// Document mapping for Prompt
	docMapping := bleve.NewDocumentMapping()

	// ID field (keyword, stored)
	idField := bleve.NewTextFieldMapping()
	idField.Analyzer = "keyword"
	idField.Store = true
	docMapping.AddFieldMappingsAt("id", idField)

	// SourceID field (keyword, stored, faceted)
	sourceIDField := bleve.NewTextFieldMapping()
	sourceIDField.Analyzer = "keyword"
	sourceIDField.Store = true
	docMapping.AddFieldMappingsAt("source_id", sourceIDField)

	// Name field (text, stored)
	nameField := bleve.NewTextFieldMapping()
	nameField.Analyzer = "en"
	nameField.Store = true
	nameField.IncludeInAll = true
	docMapping.AddFieldMappingsAt("name", nameField)

	// Description field (text, stored)
	descField := bleve.NewTextFieldMapping()
	descField.Analyzer = "en"
	descField.Store = true
	descField.IncludeInAll = true
	docMapping.AddFieldMappingsAt("description", descField)

	// Tags field (keyword, stored, faceted)
	tagsField := bleve.NewTextFieldMapping()
	tagsField.Analyzer = "keyword"
	tagsField.Store = true
	docMapping.AddFieldMappingsAt("tags", tagsField)

	// Content field (text)
	contentField := bleve.NewTextFieldMapping()
	contentField.Analyzer = "en"
	contentField.Store = false // Don't store full content - read from source file instead
	contentField.IncludeInAll = true
	docMapping.AddFieldMappingsAt("content", contentField)

	// FilePath field (keyword, stored)
	filePathField := bleve.NewTextFieldMapping()
	filePathField.Analyzer = "keyword"
	filePathField.Store = true
	docMapping.AddFieldMappingsAt("file_path", filePathField)

	// Author field (keyword, stored)
	authorField := bleve.NewTextFieldMapping()
	authorField.Analyzer = "keyword"
	authorField.Store = true
	docMapping.AddFieldMappingsAt("author", authorField)

	indexMapping.DefaultMapping = docMapping

	return indexMapping
}

// IndexPrompt indexes a single prompt.
func (i *Indexer) IndexPrompt(prompt models.Prompt) error {
	return i.index.Index(prompt.ID, prompt)
}

// IndexPrompts indexes multiple prompts in a batch.
// Uses batch operations for better performance.
func (i *Indexer) IndexPrompts(prompts []models.Prompt) error {
	batch := i.index.NewBatch()
	batchSize := 50 // Commit every 50 prompts

	for idx, prompt := range prompts {
		if err := batch.Index(prompt.ID, prompt); err != nil {
			return fmt.Errorf("failed to add prompt %s to batch: %w", prompt.ID, err)
		}

		// Commit batch when it reaches the batch size
		if (idx+1)%batchSize == 0 {
			if err := i.index.Batch(batch); err != nil {
				return fmt.Errorf("failed to commit batch: %w", err)
			}
			batch = i.index.NewBatch()
		}
	}

	// Commit remaining prompts
	if batch.Size() > 0 {
		if err := i.index.Batch(batch); err != nil {
			return fmt.Errorf("failed to commit final batch: %w", err)
		}
	}

	return nil
}

// DeletePrompt removes a prompt from the index.
func (i *Indexer) DeletePrompt(promptID string) error {
	return i.index.Delete(promptID)
}

// DeletePromptsBySource removes all prompts from a specific source.
// Useful for re-indexing after source updates.
func (i *Indexer) DeletePromptsBySource(sourceID string) error {
	// Query for all prompts from this source
	query := bleve.NewMatchQuery(sourceID)
	query.SetField("source_id")

	search := bleve.NewSearchRequest(query)
	search.Size = 10000 // Large batch

	results, err := i.index.Search(search)
	if err != nil {
		return fmt.Errorf("failed to search for source prompts: %w", err)
	}

	// Delete each prompt
	batch := i.index.NewBatch()
	for _, hit := range results.Hits {
		batch.Delete(hit.ID)
	}

	if err := i.index.Batch(batch); err != nil {
		return fmt.Errorf("failed to delete source prompts: %w", err)
	}

	return nil
}

// ReindexSource re-indexes all prompts from a source.
// Deletes old prompts and indexes new ones.
func (i *Indexer) ReindexSource(sourceID string, prompts []models.Prompt) error {
	// Delete old prompts from this source
	if err := i.DeletePromptsBySource(sourceID); err != nil {
		return fmt.Errorf("failed to delete old prompts: %w", err)
	}

	// Index new prompts
	if err := i.IndexPrompts(prompts); err != nil {
		return fmt.Errorf("failed to index new prompts: %w", err)
	}

	return nil
}

// Close closes the index.
func (i *Indexer) Close() error {
	return i.index.Close()
}

// GetIndex returns the underlying bleve index for advanced operations.
func (i *Indexer) GetIndex() bleve.Index {
	return i.index
}

// Count returns the total number of documents in the index.
func (i *Indexer) Count() (uint64, error) {
	return i.index.DocCount()
}

// DeleteIndex completely removes the index from disk.
func DeleteIndex(indexPath string) error {
	return os.RemoveAll(indexPath)
}

// EnsureIndexPath ensures the index parent directory exists.
func EnsureIndexPath(indexPath string) error {
	return os.MkdirAll(filepath.Dir(indexPath), 0755)
}
