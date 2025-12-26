# Command Contract: save

**Purpose:** Bookmark a prompt with custom alias and tags

**Priority:** P2 (Core organization feature)

## Signature

```bash
pkit save <prompt-id> --as <alias> [flags]
```

## Arguments

- `<prompt-id>` **(required)**: Prompt identifier from source
  - Format: `<source>:<name>` (e.g., `fabric:summarize`, `awesome:linux-terminal`)

## Flags

- `--as <string>` **(required)**: Bookmark alias (unique identifier)
- `--tags <string>`: Comma-separated tags (e.g., `dev,security,go`)
- `--notes <string>`: Optional user notes about the bookmark
- `--force, -f`: Overwrite existing bookmark with same alias
- `--verbose, -v`: Show detailed operation info
- `--debug`: Show full trace

## Behavior

1. **Validation**:
   - Verify prompt ID exists in index
   - Validate alias format: lowercase alphanumeric with hyphens/underscores
   - Check alias doesn't conflict with built-in commands
   - Check alias uniqueness (error if exists, unless --force)

2. **Bookmark Creation**:
   - Create Bookmark entity with:
     - Alias (user-provided)
     - PromptID (from argument)
     - SourceID and PromptName (denormalized)
     - Tags (parsed from comma-separated string)
     - Notes (optional)
     - Timestamps: CreatedAt, UpdatedAt
     - UsageCount: 0
   - Save to `~/.pkit/bookmarks.yml`

3. **Output**:
   - Success message with bookmark details
   - Suggestion to use `pkit get <alias>` or shorthand `pkit <alias>`

## Output

### Success

```
✓ Bookmarked fabric:code-review as 'review'
  Tags: dev, security, go
  Notes: Use for Go code reviews with security focus

Use it:
  pkit get review | claude -p "analyse me ~/main.go"
  pkit review | claude -p "analyse me ~/main.go"  # shorthand
```

### Success with --force (overwrite)

```
Warning: Overwriting existing bookmark 'review'

Old bookmark:
  Prompt: fabric:code-review
  Tags: dev, security
  Usage: 12 times

✓ Bookmarked fabric:improve-code as 'review'
  Tags: dev, refactoring
  Notes: For code improvement suggestions
```

## Examples

### Basic bookmark

```bash
$ pkit save fabric:summarize --as sum
✓ Bookmarked fabric:summarize as 'sum'

Use it:
  pkit sum | claude
```

### Bookmark with tags

```bash
$ pkit save fabric:code-review --as review --tags dev,security,go
✓ Bookmarked fabric:code-review as 'review'
  Tags: dev, security, go

Use it:
  pkit review | claude -p "analyse me ~/main.go"
```

### Bookmark with notes

```bash
$ pkit save fabric:security-audit --as audit --tags security --notes "Use for sensitive code paths"
✓ Bookmarked fabric:security-audit as 'audit'
  Tags: security
  Notes: Use for sensitive code paths
```

### Overwrite existing bookmark

```bash
$ pkit save fabric:improve-code --as review --force
Warning: Overwriting existing bookmark 'review'

✓ Bookmarked fabric:improve-code as 'review'
  Tags: dev, refactoring
```

### Interactive prompt selection (from search)

```bash
$ pkit search "code review"
┌──────────────────────────────────────────────────────────┐
│ SOURCE     NAME          DESCRIPTION                       │
├──────────────────────────────────────────────────────────┤
│ fabric     code-review   Expert code reviewer analyzing... │
│ fabric     improve-code  Improve code quality and...       │
│ awesome    code-helper   Assistant for coding tasks...     │
└──────────────────────────────────────────────────────────┘

$ pkit save fabric:code-review --as review --tags dev,security
✓ Bookmarked fabric:code-review as 'review'
```

## Error Handling

### Prompt ID not found

```bash
$ pkit save fabric:nonexistent --as test
Error: Prompt not found: 'fabric:nonexistent'

Available prompts from fabric:
  - fabric:summarize
  - fabric:code-review
  - fabric:improve-writing
  ...

Use 'pkit list --source fabric' to see all prompts.

Exit code: 1
```

### Invalid alias format

```bash
$ pkit save fabric:summarize --as "My Summary"
Error: Invalid alias: "My Summary"

Alias must:
  - Be lowercase alphanumeric with hyphens/underscores
  - Match pattern: ^[a-z0-9_-]+$

Valid examples:
  - summary
  - my-summary
  - code_review_v2

Exit code: 1
```

### Alias conflicts with command

```bash
$ pkit save fabric:summarize --as search
Error: Alias 'search' conflicts with built-in command

Reserved command names:
  subscribe, search, find, list, show, save, get, alias, tag, update, upgrade, status

Suggestions:
  - Use: my-search, search-prompt, custom-search

Exit code: 1
```

### Duplicate alias (without --force)

```bash
$ pkit save fabric:summarize --as review
Error: Alias 'review' already exists

Existing bookmark:
  Alias: review
  Prompt: fabric:code-review
  Tags: dev, security, go
  Usage: 12 times
  Last used: 2025-12-26 09:15:00

Options:
  - Use different alias: pkit save fabric:summarize --as summary
  - Overwrite existing: pkit save fabric:summarize --as review --force
  - Remove existing: pkit rm review

Exit code: 2
```

### Invalid prompt ID format

```bash
$ pkit save summarize --as sum
Error: Invalid prompt ID format: "summarize"

Expected format: <source>:<name>

Examples:
  - fabric:summarize
  - awesome:linux-terminal
  - internal:custom-prompt

Did you mean:
  - fabric:summarize
  - awesome:summarize

Exit code: 1
```

### Missing --as flag

```bash
$ pkit save fabric:summarize
Error: Missing required flag: --as

Usage:
  pkit save <prompt-id> --as <alias> [flags]

Example:
  pkit save fabric:summarize --as sum

Exit code: 1
```

### Invalid tags format

```bash
$ pkit save fabric:summarize --as sum --tags "dev, security"
Warning: Tags contain spaces, removing: "dev, security" → "dev,security"

✓ Bookmarked fabric:summarize as 'sum'
  Tags: dev, security
```

## Edge Cases

1. **Very long alias**: Warn if alias > 50 characters, but allow
2. **Many tags (>10)**: Warn about performance impact on tag filtering
3. **Duplicate tags**: Deduplicate automatically (e.g., `dev,dev,security` → `dev,security`)
4. **Empty tags**: Allow bookmarks without tags
5. **Special characters in notes**: Properly escape YAML, allow any UTF-8
6. **Prompt from removed source**: Allow bookmark, show warning that source may be outdated
7. **Concurrent bookmark saves**: Last write wins (acceptable for single-user Phase 1)

## Validation Rules

### Alias Validation

```go
// Lowercase alphanumeric with hyphens/underscores
validAliasRegex := regexp.MustCompile(`^[a-z0-9_-]+$`)

// Reserved command names
reservedCommands := []string{
    "subscribe", "search", "find", "list", "show",
    "save", "get", "alias", "tag", "update",
    "upgrade", "status", "help", "version",
}

func ValidateAlias(alias string) error {
    if !validAliasRegex.MatchString(alias) {
        return errors.New("alias must be lowercase alphanumeric with hyphens/underscores")
    }

    for _, cmd := range reservedCommands {
        if alias == cmd {
            return fmt.Errorf("alias %q conflicts with built-in command", alias)
        }
    }

    return nil
}
```

### Prompt ID Validation

```go
// Format: <source>:<name>
validPromptIDRegex := regexp.MustCompile(`^[a-z0-9-]+:[a-z0-9_-]+$`)

func ValidatePromptID(promptID string) (sourceID, promptName string, err error) {
    if !validPromptIDRegex.MatchString(promptID) {
        return "", "", errors.New("prompt ID must match <source>:<name>")
    }

    parts := strings.Split(promptID, ":")
    return parts[0], parts[1], nil
}
```

### Tag Validation

```go
func ParseTags(tagsStr string) []string {
    if tagsStr == "" {
        return []string{}
    }

    // Split by comma, trim spaces
    tags := strings.Split(tagsStr, ",")
    var cleaned []string
    seen := make(map[string]bool)

    for _, tag := range tags {
        tag = strings.TrimSpace(tag)
        tag = strings.ToLower(tag)

        // Skip empty or duplicate tags
        if tag == "" || seen[tag] {
            continue
        }

        // Validate format
        if !regexp.MustCompile(`^[a-z0-9_-]+$`).MatchString(tag) {
            continue // Skip invalid tags
        }

        cleaned = append(cleaned, tag)
        seen[tag] = true
    }

    return cleaned
}
```

## Exit Codes

- `0`: Success
- `1`: General error (invalid input, prompt not found, etc.)
- `2`: Alias already exists (without --force)
- `3`: Validation error

## Related Commands

- `pkit get <alias>`: Retrieve bookmarked prompt
- `pkit <alias>`: Shorthand for get
- `pkit list --bookmarks`: List all bookmarks
- `pkit alias`: Rename bookmark alias
- `pkit tag`: Update bookmark tags
- `pkit rm`: Remove bookmark
- `pkit show <alias>`: Show detailed bookmark information

## Requirements Mapping

- FR-017: Users MUST be able to bookmark prompts with custom aliases
- FR-018: Users MUST be able to tag bookmarks with multiple tags
- FR-022: System MUST store bookmarks in ~/.pkit/bookmarks.yml
- FR-023: System MUST prevent duplicate aliases and warn users
- FR-024: System MUST validate bookmark file integrity at startup

## YAML Persistence

### bookmarks.yml Structure

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
    usage_count: 0
    last_used_at: null
```

### Atomic Write Strategy

```go
func SaveBookmark(bm Bookmark) error {
    // Read existing bookmarks
    bookmarks, _ := LoadBookmarks()

    // Add or update bookmark
    updated := false
    for i, existing := range bookmarks {
        if existing.Alias == bm.Alias {
            bookmarks[i] = bm
            updated = true
            break
        }
    }

    if !updated {
        bookmarks = append(bookmarks, bm)
    }

    // Write atomically (temp file + rename)
    tmpPath := bookmarksPath + ".tmp"
    data, _ := yaml.Marshal(map[string]interface{}{"bookmarks": bookmarks})

    if err := os.WriteFile(tmpPath, data, 0600); err != nil {
        return err
    }

    return os.Rename(tmpPath, bookmarksPath)
}
```

## Testing Checklist

- [ ] Valid alias formats accepted
- [ ] Invalid alias formats rejected
- [ ] Reserved command names rejected as aliases
- [ ] Duplicate aliases rejected without --force
- [ ] Duplicate aliases overwritten with --force
- [ ] Tags parsed correctly from comma-separated string
- [ ] Duplicate tags deduplicated
- [ ] Empty tags allowed
- [ ] Prompt ID validation works
- [ ] Bookmark persisted to bookmarks.yml
- [ ] YAML format is human-readable
- [ ] Atomic write prevents corruption
- [ ] Usage count initialized to 0
- [ ] Timestamps set correctly
