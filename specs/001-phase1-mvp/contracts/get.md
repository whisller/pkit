# Command Contract: get

**Purpose:** Get prompt content for piping to execution tools

**Priority:** P1 (Core functionality - required for piping workflow)

## Signature

```bash
pkit get <alias|prompt-id> [flags]
pkit <alias> [flags]  # Shorthand form
```

## Arguments

- `<alias|prompt-id>` **(required)**: Bookmark alias or prompt ID
  - Alias: User-defined bookmark name (e.g., `review`, `sum`)
  - Prompt ID: Source prompt reference (e.g., `fabric:summarize`, `awesome:linux-terminal`)

## Flags

- `--json`: Output prompt metadata as JSON instead of plain text
- `--verbose, -v`: Show operation details to stderr (does NOT affect stdout)
- `--debug`: Show full trace to stderr (does NOT affect stdout)

## Behavior

1. **Resolution**:
   - Try resolving as bookmark alias first
   - If not found, try resolving as prompt ID
   - If still not found, show error to stderr

2. **Retrieval**:
   - Load prompt content from source filesystem
   - Use denormalized FilePath from bookmark for direct file read
   - Target: <100ms retrieval time (SC-003)

3. **Output**:
   - **stdout**: ONLY prompt content (no headers, no formatting, no metadata)
   - **stderr**: Errors, warnings, verbose output
   - CRITICAL: stdout must be clean for piping (FR-025, FR-026)

4. **Usage Tracking**:
   - Increment bookmark UsageCount
   - Update LastUsedAt timestamp
   - Persist to bookmarks.yml

## Output

### Success (stdout) - Plain Text Mode

```
# IDENTITY AND PURPOSE

You are an expert content summarizer. You take content in and output a summary with the most important points.

# OUTPUT SECTIONS

- Combine all of your understanding of the content into a single paragraph
- Output a bulleted list of the most important points
- Output a set of conclusions

# INPUT:

INPUT:
```

**CRITICAL**: stdout contains ONLY the prompt content. No headers, no metadata, no formatting.

### Success (stdout) - JSON Mode (--json)

```json
{
  "id": "fabric:summarize",
  "source_id": "fabric",
  "name": "summarize",
  "content": "# IDENTITY AND PURPOSE\n\nYou are an expert content summarizer...",
  "description": "You are an expert content summarizer. You take content in and output a summary with the most important points.",
  "tags": ["summarization", "analysis"],
  "author": "",
  "file_path": "patterns/summarize/system.md",
  "bookmark": {
    "alias": "sum",
    "tags": ["reading", "docs"],
    "usage_count": 6,
    "last_used_at": "2025-12-26T10:45:23Z"
  }
}
```

### Verbose Output (stderr only)

```
→ Resolving prompt: review
→ Found bookmark: review → fabric:code-review
→ Loading prompt from: ~/.pkit/sources/fabric/patterns/code-review/system.md
→ Updating usage stats...
✓ Complete in 12ms
```

## Examples

### Basic usage with bookmark alias

```bash
$ pkit get review
# IDENTITY AND PURPOSE

You are an expert code reviewer...
```

### Shorthand form (no "get" required)

```bash
$ pkit review
# IDENTITY AND PURPOSE

You are an expert code reviewer...
```

### Pipe to claude

```bash
$ pkit get review | claude -p "analyse me ~/main.go"
# Claude analyzes the Go file using the code review prompt
```

### Pipe with input

```bash
$ cat ~/article.txt | pkit sum | claude
# Summarizes the article using the 'sum' prompt
```

### Pipe to llm

```bash
$ pkit get security-audit | llm -m claude-3-sonnet < ~/auth.go
# Runs security audit prompt on auth.go file
```

### Pipe to fabric

```bash
$ pkit get improve-writing | fabric
# Uses Fabric's execution engine with pkit prompt
```

### Pipe to mods

```bash
$ echo "Explain quantum computing" | pkit get explain | mods
# Uses mods with pkit prompt
```

### JSON output for programmatic use

```bash
$ pkit get review --json | jq '.content'
# Extract just the content field from JSON
```

### Get by prompt ID (without bookmark)

```bash
$ pkit get fabric:summarize
# Gets prompt directly from source without bookmark
```

### Verbose mode (debugging)

```bash
$ pkit get review --verbose 2>&1 | head -1
→ Resolving prompt: review
```

## Error Handling

### Prompt not found

```bash
$ pkit get nonexistent
Error: Prompt not found: "nonexistent"

Searched:
  - Bookmarks: no alias 'nonexistent' found
  - Prompts: no prompt ID 'nonexistent' found

Did you mean:
  - review (fabric:code-review)
  - revise (fabric:revise-writing)

Use 'pkit list' to see all available prompts.

Exit code: 1
```

### Source file missing (orphaned bookmark)

```bash
$ pkit get review
Warning: Source file not found for bookmark 'review'
  Expected: ~/.pkit/sources/fabric/patterns/code-review/system.md

This may happen if:
  - Source was upgraded and prompt was removed
  - Source directory was manually modified
  - Filesystem corruption

Options:
  - Remove bookmark: pkit rm review
  - Restore source: pkit upgrade fabric
  - Check source status: pkit status fabric

Exit code: 1
```

### File read error

```bash
$ pkit get review
Error: Failed to read prompt file

Path: ~/.pkit/sources/fabric/patterns/code-review/system.md
Cause: permission denied

Fix:
  chmod +r ~/.pkit/sources/fabric/patterns/code-review/system.md

Exit code: 1
```

### Ambiguous shorthand (conflicts with command)

```bash
# If user creates bookmark named 'search'
$ pkit search
Error: Ambiguous command: 'search' is both a built-in command and a bookmark alias

Use:
  - Built-in search: pkit search <query>
  - Get bookmark: pkit get search

Note: Avoid using command names as bookmark aliases.

Exit code: 1
```

## Edge Cases

1. **Empty prompt file**: Show error, refuse to output empty content
2. **Very large prompt (>1MB)**: Show warning to stderr, proceed with output
3. **Binary file detected**: Show error, refuse to output binary content
4. **Broken UTF-8 encoding**: Show warning, attempt to output with replacement characters
5. **Piping to failing command**: pkit exits 0 (success), let downstream command handle failure
6. **Bookmark updated during read**: Use file on disk (source of truth), not cached data
7. **Concurrent access**: File reads are safe, no locking needed

## Exit Codes

- `0`: Success (prompt content written to stdout)
- `1`: Prompt not found
- `2`: File read error
- `3`: Empty or invalid prompt content

## Pipe Compatibility

**CRITICAL REQUIREMENT**: Output to stdout MUST be clean for piping (FR-025, FR-026, FR-027, FR-028)

### Correct Pipe Behavior

```bash
# Stdout contains ONLY prompt content
$ pkit get review > prompt.txt
$ cat prompt.txt
# IDENTITY AND PURPOSE...

# Stderr contains errors/warnings
$ pkit get nonexistent 2>error.log
$ cat error.log
Error: Prompt not found: "nonexistent"

# Verbose output to stderr, content to stdout
$ pkit get review --verbose 2>/dev/null | claude
# Claude receives clean prompt text
```

### Incorrect Pipe Behavior (MUST AVOID)

```bash
# BAD: Headers mixed with content in stdout
$ pkit get review
=== Prompt: review ===
# IDENTITY AND PURPOSE...
=== End of Prompt ===

# BAD: Progress/status in stdout
$ pkit get review
Loading prompt...
# IDENTITY AND PURPOSE...
Done!

# BAD: Colored output that breaks pipes
$ pkit get review | cat
^[[32m# IDENTITY AND PURPOSE^[[0m...  # ANSI codes visible
```

## Performance Requirements

- **Target**: <100ms for bookmark retrieval (SC-003)
- **Strategy**: Direct file read using denormalized FilePath from bookmark
- **No Index Lookup**: Bookmarks store FilePath to avoid index query

## Related Commands

- `pkit save`: Bookmark a prompt with alias
- `pkit list`: List available prompts
- `pkit search`: Search for prompts
- `pkit show`: Show detailed prompt information (formatted, not for piping)
- `pkit alias`: Rename bookmark alias

## Requirements Mapping

- FR-025: Output ONLY prompt text to stdout
- FR-026: Pipeable to any external tool without formatting interference
- FR-027: Errors MUST be written to stderr, not stdout
- FR-028: Return appropriate exit codes (0=success, non-zero=error)
- FR-029: Support --json flag for structured output
- SC-003: Retrieve bookmark in <100ms
- SC-004: Output successfully pipes to claude, llm, fabric, mods

## Testing Checklist

- [ ] Stdout contains ONLY prompt content (no headers, metadata, formatting)
- [ ] Stderr contains errors and warnings
- [ ] Verbose/debug output goes to stderr only
- [ ] Exit code 0 on success, non-zero on error
- [ ] Pipe to claude works: `pkit get review | claude -p "test"`
- [ ] Pipe to llm works: `pkit get review | llm -m claude-3-sonnet`
- [ ] Pipe with input works: `echo "test" | pkit get review | claude`
- [ ] JSON output is valid and parseable
- [ ] Shorthand form works: `pkit review` same as `pkit get review`
- [ ] Usage count increments correctly
- [ ] Performance <100ms for bookmarked prompts
