# Data Model: pkit Phase 1 MVP

**Feature Branch**: `001-phase1-mvp`
**Created**: 2025-12-26
**Purpose**: Define data structures, relationships, validation rules, and storage formats

## Overview

This document defines the data model for pkit Phase 1 MVP. All entities are designed for simplicity, human-readability in YAML format, and efficient search indexing. The model supports the core workflows: subscribe → search → bookmark → retrieve.

**Key Design Principles:**
- All prompt metadata (author, description, tags, version) is **extracted from source content** using format-specific parsers
- No manual configuration required
- **Fast iteration in Phase 1**: Breaking changes allowed, no backward compatibility guarantees until stabilization

## Core Entities

### 1. Source

**Purpose:** Represents a subscribed GitHub repository containing prompts.

**Attributes:**

```go
type Source struct {
    // Unique identifier (e.g., "fabric", "awesome-chatgpt-prompts")
    ID string `yaml:"id" json:"id"`

    // Display name for the source
    Name string `yaml:"name" json:"name"`

    // Full GitHub URL (e.g., "https://github.com/danielmiessler/fabric")
    URL string `yaml:"url" json:"url"`

    // Short form used in subscribe command (e.g., "fabric/patterns")
    ShortName string `yaml:"short_name,omitempty" json:"short_name,omitempty"`

    // Local filesystem path (~/.pkit/sources/<id>)
    LocalPath string `yaml:"local_path" json:"local_path"`

    // Format type determines which parser to use
    // Valid values: "fabric_pattern", "awesome_chatgpt", "markdown"
    Format string `yaml:"format" json:"format"`

    // Current git commit SHA
    CommitSHA string `yaml:"commit_sha" json:"commit_sha"`

    // Last update check timestamp (RFC3339)
    LastChecked time.Time `yaml:"last_checked" json:"last_checked"`

    // Last successful index timestamp (RFC3339)
    LastIndexed time.Time `yaml:"last_indexed" json:"last_indexed"`

    // Number of prompts indexed from this source
    PromptCount int `yaml:"prompt_count" json:"prompt_count"`

    // Subscription timestamp (RFC3339)
    SubscribedAt time.Time `yaml:"subscribed_at" json:"subscribed_at"`

    // Whether updates are available upstream
    UpdateAvailable bool `yaml:"update_available" json:"update_available"`

    // Upstream commit SHA if update available
    UpstreamSHA string `yaml:"upstream_sha,omitempty" json:"upstream_sha,omitempty"`
}
```

**Validation Rules:**

- `ID` MUST be unique across all sources
- `ID` MUST be lowercase alphanumeric with hyphens only (regex: `^[a-z0-9-]+$`)
- `URL` MUST be valid GitHub repository URL
- `Format` MUST be one of: `fabric_pattern`, `awesome_chatgpt`, `markdown`
- `LocalPath` MUST exist and be readable
- `CommitSHA` MUST be valid 40-character hex string

**Storage Location:** `~/.pkit/config.yml` (sources section)

**Example YAML:**

```yaml
sources:
  - id: fabric
    name: Fabric Patterns
    url: https://github.com/danielmiessler/fabric
    short_name: fabric/patterns
    local_path: /Users/user/.pkit/sources/fabric
    format: fabric_pattern
    commit_sha: abc123def456...
    last_checked: 2025-12-26T10:30:00Z
    last_indexed: 2025-12-26T10:32:00Z
    prompt_count: 287
    subscribed_at: 2025-12-25T14:00:00Z
    update_available: false
```

---

### 2. Prompt

**Purpose:** Represents a single prompt discovered from a source.

**Attributes:**

```go
type Prompt struct {
    // Unique identifier: <source_id>:<prompt_name> (e.g., "fabric:summarize")
    ID string `json:"id"`

    // Source identifier this prompt belongs to
    SourceID string `json:"source_id"`

    // Prompt name (unique within source)
    Name string `json:"name"`

    // Full prompt text content
    Content string `json:"content"`

    // Brief description for search results and list views (~150 chars)
    // Extracted by parser or derived from first paragraph
    Description string `json:"description"`

    // Tags/categories extracted by parser or derived
    Tags []string `json:"tags,omitempty"`

    // Author information if available from source
    // May be empty if source format doesn't provide it
    Author string `json:"author,omitempty"`

    // Version information if provided by source
    // May be empty if source format doesn't provide it
    Version string `json:"version,omitempty"`

    // Relative file path within source repository
    FilePath string `json:"file_path"`

    // Format-specific metadata (JSON blob)
    // Store extra fields that don't fit standard schema
    Metadata map[string]interface{} `json:"metadata,omitempty"`

    // When this prompt was first indexed
    IndexedAt time.Time `json:"indexed_at"`

    // When this prompt was last updated in source
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Validation Rules:**

- `ID` MUST follow format `<source_id>:<prompt_name>`
- `SourceID` MUST reference existing Source.ID
- `Name` MUST be non-empty and unique within source
- `Content` MUST be non-empty
- `Description` SHOULD be ~150 characters for consistent display in tables
- `Tags` MUST be lowercase alphanumeric with hyphens/underscores
- `FilePath` MUST be relative path within source repository

**Storage Location:** Indexed in `~/.pkit/cache/index.bleve` (bleve search index)

**Search Fields:** `name`, `description`, `tags`, `content` (weighted)

**Example JSON (from index):**

```json
{
  "id": "fabric:summarize",
  "source_id": "fabric",
  "name": "summarize",
  "content": "# IDENTITY AND PURPOSE\n\nYou are an expert content summarizer...",
  "description": "You are an expert content summarizer. You take content in and output a summary with the most important points.",
  "tags": ["summarization", "analysis"],
  "author": "",
  "version": "",
  "file_path": "patterns/summarize/system.md",
  "indexed_at": "2025-12-26T10:32:15Z",
  "updated_at": "2025-12-20T08:15:00Z"
}
```

---

### 3. Bookmark

**Purpose:** User's saved reference to a prompt with custom alias and tags.

**Attributes:**

```go
type Bookmark struct {
    // User-defined alias (unique identifier)
    Alias string `yaml:"alias" json:"alias"`

    // Reference to prompt: <source_id>:<prompt_name>
    PromptID string `yaml:"prompt_id" json:"prompt_id"`

    // Source identifier (denormalized for faster lookups)
    SourceID string `yaml:"source_id" json:"source_id"`

    // Prompt name (denormalized for faster lookups)
    PromptName string `yaml:"prompt_name" json:"prompt_name"`

    // User-defined tags (comma-separated in commands, array in storage)
    Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`

    // Optional user notes
    Notes string `yaml:"notes,omitempty" json:"notes,omitempty"`

    // Creation timestamp (RFC3339)
    CreatedAt time.Time `yaml:"created_at" json:"created_at"`

    // Last modified timestamp (RFC3339)
    UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`

    // Usage count (incremented on each `pkit get`)
    UsageCount int `yaml:"usage_count" json:"usage_count"`

    // Last used timestamp (RFC3339)
    LastUsedAt *time.Time `yaml:"last_used_at,omitempty" json:"last_used_at,omitempty"`
}
```

**Validation Rules:**

- `Alias` MUST be unique across all bookmarks (FR-023)
- `Alias` MUST be lowercase alphanumeric with hyphens/underscores (regex: `^[a-z0-9_-]+$`)
- `Alias` MUST NOT conflict with built-in commands (subscribe, search, find, list, show, save, get, update, status, etc.)
- `PromptID` MUST follow format `<source_id>:<prompt_name>`
- `PromptID` SHOULD reference existing prompt (warning if prompt no longer exists)
- `Tags` MUST be lowercase alphanumeric with hyphens/underscores
- File integrity check (FR-024): YAML MUST be valid, refuse to start if corrupted

**Storage Location:** `~/.pkit/bookmarks.yml`

**Example YAML:**

```yaml
bookmarks:
  - alias: review
    prompt_id: fabric:code-review
    source_id: fabric
    prompt_name: code-review
    tags: [dev, security, go]
    notes: Use this for Go code reviews with security focus
    created_at: 2025-12-25T15:30:00Z
    updated_at: 2025-12-26T09:15:00Z
    usage_count: 12
    last_used_at: 2025-12-26T09:15:00Z

  - alias: sum
    prompt_id: fabric:summarize
    source_id: fabric
    prompt_name: summarize
    tags: [summarization, reading]
    created_at: 2025-12-25T16:00:00Z
    updated_at: 2025-12-25T16:00:00Z
    usage_count: 5
    last_used_at: 2025-12-26T08:30:00Z
```

---

### 4. Config

**Purpose:** User configuration including sources, GitHub token reference, and preferences.

**Attributes:**

```go
type Config struct {
    // Config file version (for future migrations post-stabilization)
    Version string `yaml:"version"`

    // List of subscribed sources
    Sources []Source `yaml:"sources"`

    // GitHub configuration
    GitHub GitHubConfig `yaml:"github"`

    // Search preferences
    Search SearchConfig `yaml:"search"`

    // Display preferences
    Display DisplayConfig `yaml:"display"`

    // Cache settings
    Cache CacheConfig `yaml:"cache"`
}

type GitHubConfig struct {
    // Whether to use authenticated requests (token from keyring)
    UseAuth bool `yaml:"use_auth"`

    // Rate limit warning threshold (percentage, 0-100)
    RateLimitWarningThreshold int `yaml:"rate_limit_warning_threshold"`

    // Last known rate limit state (ephemeral, not critical to persist)
    LastRateLimit *RateLimit `yaml:"last_rate_limit,omitempty"`
}

type SearchConfig struct {
    // Maximum search results to display
    MaxResults int `yaml:"max_results"`

    // Fuzzy matching enabled
    FuzzyMatch bool `yaml:"fuzzy_match"`

    // Case sensitive search
    CaseSensitive bool `yaml:"case_sensitive"`
}

type DisplayConfig struct {
    // Use color in output
    Color bool `yaml:"color"`

    // Table style (simple, rounded, unicode)
    TableStyle string `yaml:"table_style"`

    // Date format (rfc3339, relative, short)
    DateFormat string `yaml:"date_format"`
}

type CacheConfig struct {
    // Enable search index caching
    Enabled bool `yaml:"enabled"`

    // Auto-rebuild index on source changes
    AutoRebuild bool `yaml:"auto_rebuild"`
}
```

**Validation Rules:**

- `Version` MUST match supported config versions
- `GitHub.RateLimitWarningThreshold` MUST be between 50-95
- `Search.MaxResults` MUST be between 10-1000
- `Display.TableStyle` MUST be one of: `simple`, `rounded`, `unicode`
- `Display.DateFormat` MUST be one of: `rfc3339`, `relative`, `short`

**Default Values:**

```go
var DefaultConfig = Config{
    Version: "1.0",
    Sources: []Source{},
    GitHub: GitHubConfig{
        UseAuth:                   false,
        RateLimitWarningThreshold: 80,
    },
    Search: SearchConfig{
        MaxResults:    50,
        FuzzyMatch:    true,
        CaseSensitive: false,
    },
    Display: DisplayConfig{
        Color:      true,
        TableStyle: "rounded",
        DateFormat: "relative",
    },
    Cache: CacheConfig{
        Enabled:     true,
        AutoRebuild: true,
    },
}
```

**Storage Location:** `~/.pkit/config.yml`

**Example YAML:**

```yaml
version: "1.0"
sources:
  - id: fabric
    name: Fabric Patterns
    url: https://github.com/danielmiessler/fabric
    # ... (see Source example)

github:
  use_auth: true
  rate_limit_warning_threshold: 80

search:
  max_results: 50
  fuzzy_match: true
  case_sensitive: false

display:
  color: true
  table_style: rounded
  date_format: relative

cache:
  enabled: true
  auto_rebuild: true
```

---

### 5. RateLimit

**Purpose:** Track GitHub API rate limit consumption (FR-008, FR-009).

**Attributes:**

```go
type RateLimit struct {
    // Maximum requests allowed in window
    Limit int `json:"limit"`

    // Remaining requests in current window
    Remaining int `json:"remaining"`

    // When the rate limit window resets (Unix timestamp)
    ResetAt time.Time `json:"reset_at"`

    // Resource type (core, search, graphql)
    Resource string `json:"resource"`

    // Whether authenticated request was used
    Authenticated bool `json:"authenticated"`

    // Last updated timestamp
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Validation Rules:**

- `Limit` MUST be 60 (unauthenticated) or 5000 (authenticated) for core API
- `Remaining` MUST be between 0 and Limit
- `ResetAt` MUST be in the future
- `Resource` MUST be one of: `core`, `search`, `graphql`

**Computed Fields:**

```go
// Percentage of rate limit consumed (0-100)
func (r *RateLimit) PercentageUsed() int {
    if r.Limit == 0 {
        return 0
    }
    return int((float64(r.Limit - r.Remaining) / float64(r.Limit)) * 100)
}

// Whether warning threshold exceeded
func (r *RateLimit) ShouldWarn(threshold int) bool {
    return r.PercentageUsed() >= threshold
}

// Time until reset in human-readable format
func (r *RateLimit) TimeUntilReset() time.Duration {
    return time.Until(r.ResetAt)
}
```

**Storage Location:** Ephemeral in-memory, optionally cached in `~/.pkit/config.yml` (GitHub.LastRateLimit)

---

### 6. Index

**Purpose:** Search index structure for bleve full-text search.

**Attributes:**

```go
// IndexDocument represents a document stored in bleve index
type IndexDocument struct {
    // Prompt.ID (searchable)
    ID string `json:"id"`

    // Prompt.SourceID (faceted)
    SourceID string `json:"source_id"`

    // Prompt.Name (searchable, boosted 2.0)
    Name string `json:"name"`

    // Prompt.Description (searchable, boosted 1.5)
    Description string `json:"description"`

    // Prompt.Tags (searchable, faceted)
    Tags []string `json:"tags"`

    // Prompt.Content (searchable, boosted 0.5)
    Content string `json:"content"`

    // Prompt.IndexedAt (sortable)
    IndexedAt time.Time `json:"indexed_at"`

    // Full Prompt object (stored, not indexed)
    _Prompt Prompt `json:"-"`
}
```

**Index Configuration:**

```go
// Bleve index mapping
indexMapping := bleve.NewIndexMapping()

// Document mapping
docMapping := bleve.NewDocumentMapping()

// Name field (text, boosted)
nameField := bleve.NewTextFieldMapping()
nameField.Analyzer = "en"
nameField.Boost = 2.0
docMapping.AddFieldMappingsAt("name", nameField)

// Description field (text, boosted)
descField := bleve.NewTextFieldMapping()
descField.Analyzer = "en"
descField.Boost = 1.5
docMapping.AddFieldMappingsAt("description", descField)

// Tags field (keyword, faceted)
tagsField := bleve.NewTextFieldMapping()
tagsField.Analyzer = "keyword"
docMapping.AddFieldMappingsAt("tags", tagsField)

// Content field (text, lower boost)
contentField := bleve.NewTextFieldMapping()
contentField.Analyzer = "en"
contentField.Boost = 0.5
contentField.IncludeInAll = true
docMapping.AddFieldMappingsAt("content", contentField)

indexMapping.DefaultMapping = docMapping
```

**Storage Location:** `~/.pkit/cache/index.bleve/` (directory with bleve segment files)

---

## Parser Strategy

### Overview

The `Source.Format` field determines which parser is used to extract prompt metadata from source files. Each parser implements a common interface and extracts as much metadata as the source format provides.

**Key Principle:** Description field should be ~150 characters for consistent display in search results and list views.

### Parser Interface

```go
type Parser interface {
    // Parse all prompts from a source directory
    ParsePrompts(source *Source) ([]Prompt, error)

    // Detect if this parser can handle a given source
    CanParse(sourcePath string) bool
}
```

### Format-Specific Parsers

#### 1. Fabric Pattern Parser (`fabric_pattern`)

**Source Repository:** https://github.com/danielmiessler/fabric

**File Structure:**
```
patterns/
├── summarize/
│   └── system.md        # Plain Markdown, structured sections
├── code-review/
│   └── system.md
└── ...
```

**File Format Example:**
```markdown
# IDENTITY AND PURPOSE

You are an expert content summarizer. You take content in and output a summary using the format below of the most important points.

# OUTPUT SECTIONS

- Combine all of your understanding of the content into a single paragraph
- Output a bulleted list of the most important points
- Output a set of conclusions

# OUTPUT INSTRUCTIONS

1. Output no more than 50 words per section
2. Create the output using the formatting above

# INPUT:

INPUT:
```

**Parsing Strategy:**

```go
type FabricParser struct{}

func (p *FabricParser) ParsePrompts(source *Source) ([]Prompt, error) {
    var prompts []Prompt

    // Walk source directory looking for system.md files in patterns/*/
    patternsDir := filepath.Join(source.LocalPath, "patterns")
    entries, _ := os.ReadDir(patternsDir)

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        systemFile := filepath.Join(patternsDir, entry.Name(), "system.md")
        if _, err := os.Stat(systemFile); os.IsNotExist(err) {
            continue
        }

        // Read file content
        content, _ := os.ReadFile(systemFile)

        // Extract metadata from content
        name := entry.Name()
        description := extractDescription(content, 150)  // Max 150 chars

        prompt := Prompt{
            ID:          fmt.Sprintf("%s:%s", source.ID, name),
            SourceID:    source.ID,
            Name:        name,
            Content:     string(content),
            Description: description,
            Tags:        []string{},           // Not available in format
            Author:      "",                   // Not available in format
            Version:     "",                   // Not available in format
            FilePath:    fmt.Sprintf("patterns/%s/system.md", name),
            IndexedAt:   time.Now(),
        }

        prompts = append(prompts, prompt)
    }

    return prompts, nil
}

// Extract first paragraph, truncate to maxLen
func extractDescription(content []byte, maxLen int) string {
    text := string(content)
    lines := strings.Split(text, "\n")

    var paragraph []string
    inContent := false

    for _, line := range lines {
        line = strings.TrimSpace(line)

        // Skip headers
        if strings.HasPrefix(line, "#") {
            inContent = true
            continue
        }

        if inContent {
            if line == "" && len(paragraph) > 0 {
                break // End of first paragraph
            }
            if line != "" {
                paragraph = append(paragraph, line)
            }
        }
    }

    desc := strings.Join(paragraph, " ")

    // Truncate if too long
    if len(desc) > maxLen {
        return desc[:maxLen-3] + "..."
    }

    if desc == "" {
        return "Fabric pattern"
    }

    return desc
}
```

**Metadata Extraction:**

| Field | Source | Example |
|-------|--------|---------|
| Name | Directory name | `summarize` |
| Description | First paragraph (max 150 chars) | "You are an expert content summarizer. You take content in and output a summary with the most important points." |
| Content | Full file | Entire system.md content |
| Author | N/A | Empty string |
| Version | N/A | Empty string |
| Tags | N/A | Empty array |

---

#### 2. Awesome ChatGPT Prompts Parser (`awesome_chatgpt`)

**Source Repository:** https://github.com/f/awesome-chatgpt-prompts

**File Structure:**
```
prompts.csv          # CSV with act, prompt, for_devs, type, contributor
README.md
```

**File Format Example:**
```csv
"act","prompt","for_devs","type","contributor"
"Linux Terminal","I want you to act as a linux terminal. I will type commands and you will reply with what the terminal should show...",TRUE,TEXT,f
"English Translator","I want you to act as an English translator, spelling corrector and improver...",FALSE,TEXT,f
```

**Parsing Strategy:**

```go
type AwesomeChatGPTParser struct{}

func (p *AwesomeChatGPTParser) ParsePrompts(source *Source) ([]Prompt, error) {
    var prompts []Prompt

    csvPath := filepath.Join(source.LocalPath, "prompts.csv")
    file, _ := os.Open(csvPath)
    defer file.Close()

    reader := csv.NewReader(file)

    // Read header
    _, _ = reader.Read()

    // Read rows
    for {
        row, err := reader.Read()
        if err == io.EOF {
            break
        }

        // Parse CSV row
        act := row[0]             // Column: act
        promptText := row[1]      // Column: prompt
        forDevs := row[2]         // Column: for_devs
        promptType := row[3]      // Column: type
        contributor := row[4]     // Column: contributor

        // Derive metadata
        name := slugify(act)                       // "Linux Terminal" → "linux-terminal"
        description := truncate(promptText, 150)   // First 150 chars of prompt

        // Build tags from metadata
        tags := []string{}
        if strings.ToUpper(forDevs) == "TRUE" {
            tags = append(tags, "dev")
        }
        if promptType != "" && promptType != "TEXT" {
            tags = append(tags, strings.ToLower(promptType))
        }

        prompt := Prompt{
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
        }

        prompts = append(prompts, prompt)
    }

    return prompts, nil
}

// Convert "Linux Terminal" → "linux-terminal"
func slugify(s string) string {
    s = strings.ToLower(s)
    s = strings.ReplaceAll(s, " ", "-")
    re := regexp.MustCompile(`[^a-z0-9-]+`)
    return re.ReplaceAllString(s, "")
}

// Truncate to max length with ellipsis
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

**Metadata Extraction:**

| Field | Source | Example |
|-------|--------|---------|
| Name | Slugified `act` column | `linux-terminal` |
| Description | First 150 chars of `prompt` | "I want you to act as a linux terminal. I will type commands and you will reply with what the terminal should show..." |
| Content | `prompt` column | Full prompt text |
| Author | `contributor` column | "f" |
| Version | N/A | Empty string |
| Tags | `for_devs` + `type` | ["dev"] or ["image"] |

---

#### 3. Generic Markdown Parser (`markdown`)

**Purpose:** Fallback parser for generic Markdown files

**File Structure:**
```
*.md                 # Any Markdown files
docs/*.md
prompts/*.md
```

**Parsing Strategy:**

```go
type MarkdownParser struct{}

func (p *MarkdownParser) ParsePrompts(source *Source) ([]Prompt, error) {
    var prompts []Prompt

    // Walk source directory looking for .md files
    err := filepath.Walk(source.LocalPath, func(path string, info os.FileInfo, err error) error {
        if filepath.Ext(path) != ".md" {
            return nil
        }

        // Skip README and common non-prompt files
        baseName := strings.ToLower(filepath.Base(path))
        if baseName == "readme.md" || baseName == "license.md" || baseName == "contributing.md" {
            return nil
        }

        // Read file
        content, _ := os.ReadFile(path)

        // Extract metadata
        name := strings.TrimSuffix(filepath.Base(path), ".md")
        name = slugify(name)

        description := extractDescription(content, 150)
        relPath, _ := filepath.Rel(source.LocalPath, path)

        prompt := Prompt{
            ID:          fmt.Sprintf("%s:%s", source.ID, name),
            SourceID:    source.ID,
            Name:        name,
            Content:     string(content),
            Description: description,
            Tags:        []string{},
            Author:      "",
            Version:     "",
            FilePath:    relPath,
            IndexedAt:   time.Now(),
        }

        prompts = append(prompts, prompt)
        return nil
    })

    return prompts, err
}
```

**Metadata Extraction:**

| Field | Source | Example |
|-------|--------|---------|
| Name | Filename (slugified) | `my-prompt` |
| Description | First 150 chars of content | Derived from first paragraph |
| Content | Full file | Entire file content |
| Author | N/A | Empty string |
| Version | N/A | Empty string |
| Tags | N/A | Empty array |

---

### Format Detection

When subscribing to a new source, pkit auto-detects the format:

```go
func DetectSourceFormat(localPath string) string {
    // Check for Fabric patterns structure
    if _, err := os.Stat(filepath.Join(localPath, "patterns")); err == nil {
        return "fabric_pattern"
    }

    // Check for awesome-chatgpt-prompts CSV
    if _, err := os.Stat(filepath.Join(localPath, "prompts.csv")); err == nil {
        return "awesome_chatgpt"
    }

    // Default to generic markdown
    return "markdown"
}
```

---

## Relationships

### Entity Relationship Diagram

```
┌─────────────────┐
│     Source      │
│─────────────────│
│ ID (PK)         │
│ Name            │
│ URL             │
│ Format          │◄───── Determines which Parser to use
│ CommitSHA       │
└────────┬────────┘
         │
         │ 1:N (one source has many prompts)
         │
         ↓
┌─────────────────┐
│     Prompt      │
│─────────────────│
│ ID (PK)         │◄──────────────────┐
│ SourceID (FK)   │                   │
│ Name            │                   │
│ Content         │◄── Parsed by      │
│ Description     │    format-specific│
│ Tags            │    parser         │
│ Author          │                   │
└────────┬────────┘                   │
         │                            │
         │ 1:1 (one prompt can be     │ Reference
         │      bookmarked once)      │ (PromptID)
         │                            │
         ↓                            │
┌─────────────────┐                   │
│    Bookmark     │                   │
│─────────────────│                   │
│ Alias (PK)      │                   │
│ PromptID (FK)   │───────────────────┘
│ SourceID        │ (denormalized)
│ Tags            │
│ UsageCount      │
└────────┬────────┘
         │
         │ N:M (many bookmarks have many tags)
         │
         ↓
┌─────────────────┐
│      Tag        │
│─────────────────│
│ Name (implicit) │
└─────────────────┘
```

---

## Validation Rules

### Startup Validation (FR-024)

**Bookmark File Integrity:**

```go
func ValidateBookmarksFile(path string) error {
    // Check file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil // Empty state OK
    }

    // Read file
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("cannot read bookmarks file: %w", err)
    }

    // Parse YAML
    var bookmarks struct {
        Bookmarks []Bookmark `yaml:"bookmarks"`
    }
    if err := yaml.Unmarshal(data, &bookmarks); err != nil {
        return fmt.Errorf("corrupted bookmarks file at %s: %w\n\nRecovery options:\n  1. Manually fix YAML syntax errors\n  2. Restore from backup\n  3. Run 'pkit init --force' to reset (WARNING: destroys all bookmarks)", path, err)
    }

    // Validate each bookmark
    aliases := make(map[string]bool)
    for i, bm := range bookmarks.Bookmarks {
        // Check duplicate aliases
        if aliases[bm.Alias] {
            return fmt.Errorf("duplicate alias %q at index %d", bm.Alias, i)
        }
        aliases[bm.Alias] = true

        // Validate alias format
        if !validAliasRegex.MatchString(bm.Alias) {
            return fmt.Errorf("invalid alias %q: must be lowercase alphanumeric with hyphens/underscores", bm.Alias)
        }

        // Validate prompt ID format
        if !validPromptIDRegex.MatchString(bm.PromptID) {
            return fmt.Errorf("invalid prompt_id %q: must be <source>:<name>", bm.PromptID)
        }
    }

    return nil
}
```

**Config File Validation:**

```go
func ValidateConfigFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // Will create default
        }
        return fmt.Errorf("cannot read config: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return fmt.Errorf("corrupted config file: %w", err)
    }

    // Validate version
    if cfg.Version == "" {
        return fmt.Errorf("config missing version field")
    }

    // Validate sources
    sourceIDs := make(map[string]bool)
    for i, src := range cfg.Sources {
        if sourceIDs[src.ID] {
            return fmt.Errorf("duplicate source ID %q at index %d", src.ID, i)
        }
        sourceIDs[src.ID] = true

        if err := ValidateSource(&src); err != nil {
            return fmt.Errorf("invalid source %q: %w", src.ID, err)
        }
    }

    return nil
}
```

### Runtime Validation

**Alias Conflict Prevention (FR-023):**

```go
var reservedCommands = []string{
    "subscribe", "search", "find", "list", "show",
    "save", "get", "alias", "tag", "update",
    "upgrade", "status", "help", "version",
}

func ValidateAlias(alias string) error {
    // Check format
    if !validAliasRegex.MatchString(alias) {
        return fmt.Errorf("alias must be lowercase alphanumeric with hyphens/underscores")
    }

    // Check reserved commands
    for _, cmd := range reservedCommands {
        if alias == cmd {
            return fmt.Errorf("alias %q conflicts with built-in command", alias)
        }
    }

    // Check existing bookmarks
    if exists, _ := bookmarkExists(alias); exists {
        return fmt.Errorf("alias %q already exists (use --force to overwrite)", alias)
    }

    return nil
}
```

**Source URL Validation (FR-001):**

```go
func ValidateSourceURL(input string) (fullURL string, err error) {
    // Short form: "fabric/patterns" → "https://github.com/danielmiessler/fabric"
    if shortFormRegex.MatchString(input) {
        parts := strings.Split(input, "/")
        if len(parts) != 2 {
            return "", fmt.Errorf("short form must be <org>/<repo>")
        }
        fullURL = fmt.Sprintf("https://github.com/%s/%s", parts[0], parts[1])
    } else {
        fullURL = input
    }

    // Validate URL
    u, err := url.Parse(fullURL)
    if err != nil {
        return "", fmt.Errorf("invalid URL: %w", err)
    }

    // Must be GitHub
    if u.Host != "github.com" {
        return "", fmt.Errorf("only GitHub repositories supported")
    }

    // Must have path
    if u.Path == "" || u.Path == "/" {
        return "", fmt.Errorf("repository path required")
    }

    return fullURL, nil
}
```

---

## Storage Format

### Directory Structure

```
~/.pkit/
├── config.yml              # User configuration + subscribed sources
├── bookmarks.yml           # User bookmarks
├── sources/                # Cloned repositories
│   ├── fabric/
│   │   ├── .git/
│   │   ├── patterns/       # Fabric format: patterns/<name>/system.md
│   │   │   ├── summarize/
│   │   │   │   └── system.md
│   │   │   └── code-review/
│   │   │       └── system.md
│   │   └── ...
│   ├── awesome-chatgpt-prompts/
│   │   ├── .git/
│   │   └── prompts.csv     # CSV format
│   └── ...
├── cache/
│   ├── index.bleve/        # Bleve search index
│   │   ├── index_meta.json
│   │   └── store/
│   └── .gitignore          # Ignore cache in version control
└── backups/                # Automatic backups before destructive ops
    └── 20251226-103000/
        ├── config.yml
        └── bookmarks.yml
```

### YAML Schemas

**config.yml:**

```yaml
version: "1.0"

sources:
  - id: fabric
    name: Fabric Patterns
    url: https://github.com/danielmiessler/fabric
    short_name: fabric/patterns
    local_path: /Users/user/.pkit/sources/fabric
    format: fabric_pattern
    commit_sha: abc123...
    last_checked: 2025-12-26T10:30:00Z
    last_indexed: 2025-12-26T10:32:00Z
    prompt_count: 287
    subscribed_at: 2025-12-25T14:00:00Z
    update_available: false

github:
  use_auth: true
  rate_limit_warning_threshold: 80

search:
  max_results: 50
  fuzzy_match: true
  case_sensitive: false

display:
  color: true
  table_style: rounded
  date_format: relative

cache:
  enabled: true
  auto_rebuild: true
```

**bookmarks.yml:**

```yaml
bookmarks:
  - alias: review
    prompt_id: fabric:code-review
    source_id: fabric
    prompt_name: code-review
    tags:
      - dev
      - security
      - go
    notes: Use for Go code reviews with security focus
    created_at: 2025-12-25T15:30:00Z
    updated_at: 2025-12-26T09:15:00Z
    usage_count: 12
    last_used_at: 2025-12-26T09:15:00Z
```

---

## Performance Considerations

### Indexing Performance

**Target:** <30 seconds for 300 prompts (SC-001)

**Strategies:**

1. **Parallel indexing** - Use goroutines with `errgroup` (limit: 10 concurrent)
2. **Batch commits** - Commit to bleve in batches of 50 documents
3. **Progress tracking** - Show per-source progress to stderr

```go
func IndexSource(source *Source) error {
    // Get format-specific parser
    parser := GetParser(source.Format)

    // Parse all prompts from source
    prompts, err := parser.ParsePrompts(source)
    if err != nil {
        return err
    }

    batch := index.NewBatch()
    batchSize := 50

    for i, prompt := range prompts {
        doc := promptToIndexDoc(prompt)
        batch.Index(doc.ID, doc)

        if (i+1)%batchSize == 0 {
            if err := index.Batch(batch); err != nil {
                return err
            }
            batch = index.NewBatch()

            // Progress update
            fmt.Fprintf(os.Stderr, "\r[%s] Indexed %d/%d prompts",
                source.ID, i+1, len(prompts))
        }
    }

    // Final batch
    if batch.Size() > 0 {
        return index.Batch(batch)
    }

    return nil
}
```

### Search Performance

**Target:** <1 second for indexed search (SC-002)

**Strategies:**

1. **Field boosting** - Name (2.0), Description (1.5), Content (0.5)
2. **Result limiting** - Max 50 results by default
3. **Lazy loading** - Load full Prompt objects only when needed

### Get Performance

**Target:** <100ms for bookmark retrieval (SC-003)

**Strategies:**

1. **Denormalized data** - Store SourceID and PromptName in Bookmark
2. **Direct file read** - Read prompt content directly from source filesystem using FilePath
3. **No index lookup needed** - Use stored FilePath to read content

---

## Testing Strategy

### Unit Tests

```go
// Test alias validation
func TestValidateAlias(t *testing.T) {
    tests := []struct {
        alias   string
        wantErr bool
    }{
        {"valid-alias", false},
        {"valid_alias", false},
        {"valid123", false},
        {"Invalid", true},     // Uppercase
        {"invalid!", true},    // Special char
        {"subscribe", true},   // Reserved
    }

    for _, tt := range tests {
        err := ValidateAlias(tt.alias)
        if (err != nil) != tt.wantErr {
            t.Errorf("ValidateAlias(%q) error = %v, wantErr %v",
                tt.alias, err, tt.wantErr)
        }
    }
}

// Test description extraction
func TestExtractDescription(t *testing.T) {
    content := []byte(`# IDENTITY

You are an expert content summarizer. You take content in and output a summary.

# OUTPUT

- Bullet points
`)

    desc := extractDescription(content, 150)

    want := "You are an expert content summarizer. You take content in and output a summary."
    if desc != want {
        t.Errorf("extractDescription() = %q, want %q", desc, want)
    }
}

// Test description truncation
func TestExtractDescriptionTruncate(t *testing.T) {
    longText := strings.Repeat("a", 200)
    content := []byte("# HEADING\n\n" + longText)

    desc := extractDescription(content, 150)

    if len(desc) != 150 {
        t.Errorf("extractDescription() length = %d, want 150", len(desc))
    }

    if !strings.HasSuffix(desc, "...") {
        t.Errorf("extractDescription() should end with ellipsis")
    }
}
```

### Integration Tests

```go
// Test Fabric parser
func TestFabricParserWorkflow(t *testing.T) {
    tmpDir := t.TempDir()
    patternsDir := filepath.Join(tmpDir, "patterns", "summarize")
    os.MkdirAll(patternsDir, 0755)

    systemMd := `# IDENTITY AND PURPOSE

You are an expert content summarizer.

# OUTPUT INSTRUCTIONS

1. Output bullet points
2. Be concise`

    os.WriteFile(filepath.Join(patternsDir, "system.md"), []byte(systemMd), 0644)

    source := &Source{
        ID:        "test-fabric",
        LocalPath: tmpDir,
        Format:    "fabric_pattern",
    }

    parser := &FabricParser{}
    prompts, err := parser.ParsePrompts(source)

    if err != nil {
        t.Fatalf("ParsePrompts failed: %v", err)
    }

    if len(prompts) != 1 {
        t.Fatalf("Expected 1 prompt, got %d", len(prompts))
    }

    prompt := prompts[0]
    if prompt.Name != "summarize" {
        t.Errorf("Name = %q, want 'summarize'", prompt.Name)
    }

    if prompt.Description == "" {
        t.Error("Description should not be empty")
    }

    if len(prompt.Description) > 150 {
        t.Errorf("Description too long: %d chars", len(prompt.Description))
    }
}
```

---

## Appendix

### Regular Expressions

```go
var (
    // Alias format: lowercase alphanumeric with hyphens/underscores
    validAliasRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)

    // Source ID format: lowercase alphanumeric with hyphens
    validSourceIDRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

    // Prompt ID format: <source>:<name>
    validPromptIDRegex = regexp.MustCompile(`^[a-z0-9-]+:[a-z0-9_-]+$`)

    // Short form: org/repo
    shortFormRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`)

    // Git commit SHA: 40 hex characters
    validSHARegex = regexp.MustCompile(`^[a-f0-9]{40}$`)
)
```

### Error Messages

**Bookmark file corrupted:**

```
Error: Corrupted bookmarks file at ~/.pkit/bookmarks.yml
  Cause: yaml: line 12: mapping values are not allowed in this context

Recovery options:
  1. Manually fix YAML syntax errors in ~/.pkit/bookmarks.yml
  2. Restore from backup: ~/.pkit/backups/20251226-103000/bookmarks.yml
  3. Reset to empty state: pkit init --force (WARNING: destroys all bookmarks)
```

**Duplicate alias:**

```
Error: Alias "review" already exists

Existing bookmark:
  Alias: review
  Prompt: fabric:code-review
  Tags: dev, security, go

Options:
  - Use a different alias: pkit save fabric:improve-code --as review2
  - Overwrite existing: pkit save fabric:improve-code --as review --force
```

**Rate limit warning:**

```
Warning: GitHub API rate limit at 85% (750/5000 requests remaining)
  Resets at: 2025-12-26 11:45:00 (in 45 minutes)

To increase rate limit to 5000 requests/hour:
  1. Create GitHub personal access token: https://github.com/settings/tokens
  2. Store token securely: pkit config set-token
  3. Enable authentication: pkit config set github.use_auth true
```

---

## Changelog

- **2025-12-26**: Initial data model for Phase 1 MVP
  - Removed `LongDescription` field (unnecessary - use Description for lists, Content for details)
  - Removed migration strategy section (fast iteration in Phase 1, no backward compatibility needed yet)
  - Simplified parsers to extract ~150 char descriptions only
  - Added note about breaking changes allowed in Phase 1
