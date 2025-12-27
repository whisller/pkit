# Command Contract: subscribe

**Purpose:** Subscribe to a GitHub repository as a prompt source

**Priority:** P1 (Core functionality)

## Signature

```bash
pkit subscribe <source> [flags]
```

## Arguments

- `<source>` **(required)**: GitHub repository identifier
  - Short form: `<org>/<repo>` (e.g., `fabric/patterns`, `f/awesome-chatgpt-prompts`)
  - Full URL: `https://github.com/<org>/<repo>`

## Flags

- `--name <string>`: Custom display name for the source (default: derived from repo name)
- `--id <string>`: Custom ID for the source (default: derived from repo name)
- `--format <string>`: Force specific parser format (`fabric_pattern`, `awesome_chatgpt`, `markdown`)
  - Default: Auto-detect based on repository structure
- `--verbose, -v`: Show detailed progress and git operations
- `--debug`: Show full trace including timing information

## Behavior

1. **Validation**:
   - Validate source URL/short form format
   - Check if source ID already exists (error if duplicate)
   - Verify GitHub repository is accessible

2. **Cloning**:
   - Clone repository to `~/.pkit/sources/<id>/`
   - Use GitHub token from keyring if available (FR-010)
   - Show progress bar for clone operation
   - Handle rate limiting gracefully (FR-008, FR-009)

3. **Format Detection**:
   - Auto-detect format if not specified:
     - Check for `patterns/` directory → `fabric_pattern`
     - Check for `prompts.csv` → `awesome_chatgpt`
     - Default: `markdown`

4. **Indexing**:
   - Parse prompts using format-specific parser
   - Index prompts into bleve search index
   - Show progress: "[source] Indexed X/Y prompts"
   - Target: <30 seconds for ~300 prompts (SC-001)

5. **Persistence**:
   - Add source to `~/.pkit/config.yml`
   - Save with metadata: URL, format, commit SHA, timestamps

## Output

### Success (stdout)

```
✓ Subscribed to fabric/patterns
  Format: fabric_pattern
  Prompts: 287
  Location: ~/.pkit/sources/fabric

Use 'pkit search "" --source fabric' to see all prompts from this source
```

### Progress (stderr)

```
Cloning repository...
[========================================] 100%
Indexing prompts...
[fabric] Indexed 287/287 prompts
```

### With --verbose (stderr)

```
→ Resolving source: fabric/patterns
→ Full URL: https://github.com/danielmiessler/fabric
→ Local path: /Users/user/.pkit/sources/fabric
→ Checking GitHub authentication... using token from keyring
→ Cloning repository...
  git clone https://github.com/danielmiessler/fabric /Users/user/.pkit/sources/fabric
[========================================] 100%
→ Detecting format... fabric_pattern
→ Parsing prompts...
  Found 287 patterns in patterns/ directory
→ Indexing prompts...
[fabric] Indexed 50/287 prompts
[fabric] Indexed 100/287 prompts
...
[fabric] Indexed 287/287 prompts
→ Saving configuration...
✓ Complete in 18.2s
```

## Examples

### Subscribe with short form

```bash
$ pkit subscribe fabric/patterns
✓ Subscribed to fabric/patterns
  Format: fabric_pattern
  Prompts: 287
  Location: ~/.pkit/sources/fabric
```

### Subscribe with full URL

```bash
$ pkit subscribe https://github.com/f/awesome-chatgpt-prompts
✓ Subscribed to awesome-chatgpt-prompts
  Format: awesome_chatgpt
  Prompts: 163
  Location: ~/.pkit/sources/awesome-chatgpt-prompts
```

### Subscribe with custom ID

```bash
$ pkit subscribe company/internal-prompts --id internal
✓ Subscribed to internal
  Format: markdown
  Prompts: 42
  Location: ~/.pkit/sources/internal
```

### Subscribe multiple sources in parallel

```bash
$ pkit subscribe fabric/patterns f/awesome-chatgpt-prompts company/internal
Subscribing to 3 sources in parallel...
[fabric] Cloning... ✓ 287 prompts
[awesome-chatgpt-prompts] Cloning... ✓ 163 prompts
[internal] Cloning... ✓ 42 prompts
✓ All sources subscribed successfully
```

## Error Handling

### Invalid source format

```bash
$ pkit subscribe invalid-format
Error: Invalid source format: "invalid-format"

Expected formats:
  Short form: <org>/<repo> (e.g., fabric/patterns)
  Full URL: https://github.com/<org>/<repo>

Exit code: 1
```

### Duplicate source ID

```bash
$ pkit subscribe fabric/patterns
Error: Source 'fabric' already exists

Existing source:
  URL: https://github.com/danielmiessler/fabric
  Prompts: 287
  Last updated: 2025-12-26 10:30:00

Options:
  - Use different ID: pkit subscribe fabric/patterns --id fabric2
  - Upgrade existing: pkit upgrade fabric
  - Remove and re-subscribe: pkit unsubscribe fabric && pkit subscribe fabric/patterns

Exit code: 1
```

### Repository not found

```bash
$ pkit subscribe nonexistent/repo
Error: Repository not found: https://github.com/nonexistent/repo

Possible causes:
  - Repository doesn't exist
  - Repository is private (requires authentication)
  - Network connectivity issue

To authenticate:
  pkit config set-token

Exit code: 1
```

### Rate limit exceeded

```bash
$ pkit subscribe fabric/patterns
Warning: GitHub API rate limit at 95% (250/5000 requests remaining)
  Resets at: 2025-12-26 11:45:00 (in 15 minutes)

Proceeding with subscription...
✓ Subscribed to fabric/patterns

To increase rate limit:
  1. Create GitHub token: https://github.com/settings/tokens
  2. Store token: pkit config set-token
  3. Enable auth: pkit config set github.use_auth true
```

### Network error

```bash
$ pkit subscribe fabric/patterns
Error: Failed to clone repository

Cause: dial tcp: lookup github.com: no such host

Troubleshooting:
  - Check network connectivity
  - Verify GitHub is accessible
  - Try again later

Exit code: 1
```

### Parsing error

```bash
$ pkit subscribe custom/prompts --format fabric_pattern
Warning: Failed to parse 3 prompts:
  - patterns/invalid/system.md: yaml: unmarshal error
  - patterns/broken/system.md: file read error
  - patterns/empty/system.md: empty content

✓ Subscribed to custom
  Format: fabric_pattern
  Prompts: 45 (3 skipped)
  Location: ~/.pkit/sources/custom

Use 'pkit subscribe custom --debug' to see detailed errors
```

## Edge Cases

1. **Empty repository**: Subscribe succeeds but with 0 prompts indexed
2. **Large repository (1000+ prompts)**: Show periodic progress updates every 10%
3. **Network interruption during clone**: Clean up partial clone, show error with retry suggestion
4. **Permission denied on ~/.pkit/sources**: Show error with permission fix instructions
5. **Disk space full**: Show error with disk space check
6. **Repository format changes**: Auto-detect handles gracefully, re-index on next upgrade
7. **Private repository without auth**: Clear error message with authentication instructions

## Exit Codes

- `0`: Success
- `1`: General error (invalid input, network error, etc.)
- `2`: Source already exists
- `3`: Permission denied
- `4`: Disk space full

## Related Commands

- `pkit search`: Search across subscribed sources (use with empty query to list all)
- `pkit find`: Interactive prompt finder
- `pkit update`: Check for updates to subscribed sources
- `pkit upgrade`: Upgrade subscribed sources
- `pkit unsubscribe`: Remove a subscribed source
- `pkit status`: Show status of all subscribed sources

## Requirements Mapping

- FR-001: Support short syntax and full URLs
- FR-002: Clone to ~/.pkit/sources/ with organized structure
- FR-003: Support Fabric and awesome-chatgpt-prompts formats
- FR-004: Index prompts and build searchable metadata
- FR-007: Work with unauthenticated GitHub API by default
- FR-008: Track and warn about rate limit consumption
- FR-009: Provide instructions for configuring GitHub token
- FR-010: Support optional GitHub token from keyring
- FR-041: Process multiple sources in parallel
- FR-042: Display progress for each parallel operation
- FR-043: Wait for all parallel operations to complete
- SC-001: Complete indexing in <30 seconds for ~300 prompts
