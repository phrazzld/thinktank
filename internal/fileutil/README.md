# internal/fileutil

File discovery and content gathering.

## Overview

Handles walking directories, filtering files (include/exclude patterns, .gitignore), reading content, and computing statistics. This is the "context gathering" component.

## Key Components

| File | Purpose |
|------|---------|
| `filtering.go` | File filtering logic, pattern matching |
| `directory.go` | Directory walking, gitignore handling |
| `reader.go` | File content reading |

## Usage

```go
gatherer := fileutil.NewContextGatherer(logger)
context, stats, err := gatherer.GatherContext(ctx, paths, config)
```

## Filtering Rules

Applied in order:
1. Skip hidden files/directories (unless explicitly included)
2. Apply .gitignore patterns
3. Apply exclude patterns (`--exclude`)
4. Apply exclude-names patterns (`--exclude-names`)
5. Apply include patterns (`--include`) - if set, only matching files pass

## Statistics

Returns file statistics:
- Total files discovered
- Files included/excluded
- Token counts per file
- Total context size
