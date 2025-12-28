package index

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/whisller/pkit/pkg/models"
)

// SearchOptions contains options for searching prompts.
type SearchOptions struct {
	// Query string to search for
	Query string

	// MaxResults limits the number of results
	MaxResults int

	// Source filter (optional)
	SourceID string

	// Tags filter (optional)
	Tags []string

	// Fuzzy matching enabled
	Fuzzy bool

	// Case sensitive search
	CaseSensitive bool
}

// SearchResult contains a search hit with prompt data.
type SearchResult struct {
	Prompt models.Prompt
	Score  float64
}

// Search searches for prompts matching the query.
func (i *Indexer) Search(opts SearchOptions) ([]SearchResult, error) {
	// Build query
	q := i.buildQuery(opts)

	// Create search request
	searchReq := bleve.NewSearchRequest(q)
	searchReq.Size = opts.MaxResults
	searchReq.Fields = []string{"*"} // Include all stored fields

	// Add facets for source and tags
	searchReq.AddFacet("sources", bleve.NewFacetRequest("source_id", 10))
	searchReq.AddFacet("tags", bleve.NewFacetRequest("tags", 20))

	// Execute search
	searchResults, err := i.index.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results to SearchResult slice
	results := make([]SearchResult, 0, len(searchResults.Hits))
	for _, hit := range searchResults.Hits {
		prompt, err := i.hitToPrompt(hit)
		if err != nil {
			// Log warning but continue with other results
			fmt.Printf("Warning: failed to convert hit to prompt: %v\n", err)
			continue
		}

		results = append(results, SearchResult{
			Prompt: prompt,
			Score:  hit.Score,
		})
	}

	return results, nil
}

// buildQuery builds a bleve query from search options.
func (i *Indexer) buildQuery(opts SearchOptions) query.Query {
	var queries []query.Query

	// Main query string
	if opts.Query != "" {
		if opts.Fuzzy {
			// Use fuzzy match query for typo tolerance
			fuzzyQuery := bleve.NewFuzzyQuery(opts.Query)
			fuzzyQuery.Fuzziness = 1 // Allow 1 character difference
			queries = append(queries, fuzzyQuery)
		} else {
			// Use match query for standard search
			matchQuery := bleve.NewMatchQuery(opts.Query)
			queries = append(queries, matchQuery)
		}
	}

	// Source filter
	if opts.SourceID != "" {
		sourceQuery := bleve.NewMatchQuery(opts.SourceID)
		sourceQuery.SetField("source_id")
		queries = append(queries, sourceQuery)
	}

	// Tags filter
	if len(opts.Tags) > 0 {
		for _, tag := range opts.Tags {
			tagQuery := bleve.NewMatchQuery(tag)
			tagQuery.SetField("tags")
			queries = append(queries, tagQuery)
		}
	}

	// Combine all queries with AND logic
	if len(queries) == 0 {
		// No filters, return all documents
		return bleve.NewMatchAllQuery()
	} else if len(queries) == 1 {
		return queries[0]
	} else {
		conjunctionQuery := bleve.NewConjunctionQuery(queries...)
		return conjunctionQuery
	}
}

// hitToPrompt converts a bleve search hit to a Prompt model.
func (i *Indexer) hitToPrompt(hit *search.DocumentMatch) (models.Prompt, error) {
	prompt := models.Prompt{
		ID: hit.ID,
	}

	// Extract fields from hit
	if val, ok := hit.Fields["source_id"].(string); ok {
		prompt.SourceID = val
	}
	if val, ok := hit.Fields["name"].(string); ok {
		prompt.Name = val
	}
	if val, ok := hit.Fields["description"].(string); ok {
		prompt.Description = val
	}
	if val, ok := hit.Fields["file_path"].(string); ok {
		prompt.FilePath = val
	}
	if val, ok := hit.Fields["author"].(string); ok {
		prompt.Author = val
	}

	// Content is NOT stored in index - it's loaded dynamically from source files when needed
	// The content field may still be in hit.Fields but will be empty
	if val, ok := hit.Fields["content"].(string); ok {
		prompt.Content = val
	}

	// Tags might be stored as interface{} slice
	if val, ok := hit.Fields["tags"]; ok {
		if tags, ok := val.([]interface{}); ok {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					prompt.Tags = append(prompt.Tags, tagStr)
				}
			}
		}
	}

	return prompt, nil
}

// SearchBySource returns all prompts from a specific source.
func (i *Indexer) SearchBySource(sourceID string, maxResults int) ([]SearchResult, error) {
	return i.Search(SearchOptions{
		SourceID:   sourceID,
		MaxResults: maxResults,
	})
}

// SearchByTag returns all prompts with a specific tag.
func (i *Indexer) SearchByTag(tag string, maxResults int) ([]SearchResult, error) {
	return i.Search(SearchOptions{
		Tags:       []string{tag},
		MaxResults: maxResults,
	})
}

// GetPromptByID retrieves a specific prompt by ID.
func (i *Indexer) GetPromptByID(promptID string) (*models.Prompt, error) {
	// Use search to get the document instead of Document() which may not exist
	query := bleve.NewDocIDQuery([]string{promptID})
	searchReq := bleve.NewSearchRequest(query)
	searchReq.Size = 1
	searchReq.Fields = []string{"*"}

	searchResults, err := i.index.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(searchResults.Hits) == 0 {
		return nil, fmt.Errorf("prompt not found: %s", promptID)
	}

	prompt, err := i.hitToPrompt(searchResults.Hits[0])
	if err != nil {
		return nil, err
	}

	return &prompt, nil
}
