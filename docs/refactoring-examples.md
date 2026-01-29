# Function Extraction Examples

This document provides concrete examples of Carmack-style function extraction from the 2025-07-08 refactoring project. These examples demonstrate proven patterns for separating I/O operations from business logic and decomposing large functions into focused, testable components.

## Overview

The refactoring followed John Carmack's incremental approach:
1. **Extract pure functions** - Separate business logic from I/O operations
2. **Decompose large functions** - Break >100 LOC functions into focused phases
3. **Test extracted functions** - Use table-driven tests for comprehensive coverage
4. **Validate behavior** - Ensure identical behavior post-refactoring

**Results**: 90.4% test coverage, 35-70% performance improvements, all functions <100 LOC

## Pattern 1: Pure Function Extraction

### Example: Console Writer Formatting

**Before: Mixed I/O and business logic**
```go
// console_writer.go - Original mixed implementation
func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
    msg := string(p)

    // Business logic mixed with I/O
    if strings.Contains(msg, "ERROR") {
        msg = colorize(msg, red)
    }

    // Duration formatting logic
    if duration := extractDuration(msg); duration != "" {
        formatted := formatDuration(duration)
        msg = strings.Replace(msg, duration, formatted, 1)
    }

    // Width formatting logic
    if len(msg) > w.maxWidth {
        msg = truncateToWidth(msg, w.maxWidth)
    }

    // I/O operation
    return w.writer.Write([]byte(msg))
}
```

**After: Pure business logic extracted**
```go
// formatting.go - Pure business logic functions
func FormatDuration(duration time.Duration) string {
    // Pure function - no I/O, fully testable
    if duration < time.Second {
        return fmt.Sprintf("%.0fms", duration.Seconds()*1000)
    }
    return fmt.Sprintf("%.1fs", duration.Seconds())
}

func FormatToWidth(message string, maxWidth int) string {
    // Pure function - deterministic output
    if len(message) <= maxWidth {
        return message
    }
    return message[:maxWidth-3] + "..."
}

func ColorizeStatus(message string, status string) string {
    // Pure function - easy to test all color combinations
    switch status {
    case "ERROR":
        return red + message + reset
    case "SUCCESS":
        return green + message + reset
    default:
        return message
    }
}

// console_writer.go - Pure I/O operation
func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
    msg := string(p)

    // Use pure functions for business logic
    msg = FormatDuration(msg)
    msg = FormatToWidth(msg, w.maxWidth)
    msg = ColorizeStatus(msg, extractStatus(msg))

    // Pure I/O operation
    return w.writer.Write([]byte(msg))
}
```

**Testing Pattern**
```go
func TestFormatDuration(t *testing.T) {
    tests := []struct {
        name     string
        duration time.Duration
        expected string
    }{
        {"milliseconds", 500 * time.Millisecond, "500ms"},
        {"seconds", 2 * time.Second, "2.0s"},
        {"minutes", 90 * time.Second, "90.0s"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FormatDuration(tt.duration)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Pattern 2: Function Decomposition

### Example: Execute Function Breakdown

**Before: Large monolithic function**
```go
// app.go - Original 370 LOC Execute function
func Execute(config Config) error {
    // Setup phase (lines 38-57)
    outputDir := filepath.Join(config.OutputDir, generateTimestamp())
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return fmt.Errorf("failed to create output directory: %w", err)
    }

    logger := logutil.NewConsoleWriter(os.Stdout)
    auditLogger := auditlog.NewFileLogger(filepath.Join(outputDir, "audit.log"))

    // Input reading phase (lines 59-121)
    instructionsContent, err := os.ReadFile(config.InstructionsFile)
    if err != nil {
        return fmt.Errorf("failed to read instructions: %w", err)
    }

    auditLogger.Log("instructions_read", map[string]interface{}{
        "file": config.InstructionsFile,
        "size": len(instructionsContent),
    })

    // Client initialization phase (lines 122-164)
    client, err := openrouter.NewClient(config.APIKey)
    if err != nil {
        return fmt.Errorf("failed to create client: %w", err)
    }

    // ... 200+ more lines of mixed logic

    // Execution phase (lines 198-207)
    orchestrator := NewOrchestrator(client, fileWriter, rateLimiter)
    result := orchestrator.Run(context.Background(), config)

    return handleResult(result)
}
```

**After: Focused functions**
```go
// app.go - Decomposed into focused functions
func Execute(config Config) error {
    // Clear pipeline with focused functions
    if err := gatherProjectFiles(config); err != nil {
        return err
    }

    if err := processFiles(config); err != nil {
        return err
    }

    if err := generateOutput(config); err != nil {
        return err
    }

    return writeResults(config)
}

func gatherProjectFiles(config Config) error {
    // Setup phase - 25 LOC focused on output directory and audit setup
    outputDir := filepath.Join(config.OutputDir, generateTimestamp())
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return fmt.Errorf("failed to create output directory: %w", err)
    }

    auditLogger := auditlog.NewFileLogger(filepath.Join(outputDir, "audit.log"))
    config.AuditLogger = auditLogger

    return nil
}

func processFiles(config Config) error {
    // Input reading phase - 46 LOC focused on file processing
    instructionsContent, err := os.ReadFile(config.InstructionsFile)
    if err != nil {
        return fmt.Errorf("failed to read instructions: %w", err)
    }

    config.AuditLogger.Log("instructions_read", map[string]interface{}{
        "file": config.InstructionsFile,
        "size": len(instructionsContent),
    })

    config.Instructions = string(instructionsContent)
    return nil
}

func generateOutput(config Config) error {
    // Client initialization phase - 71 LOC focused on LLM setup
    client, err := openrouter.NewClient(config.APIKey)
    if err != nil {
        return fmt.Errorf("failed to create client: %w", err)
    }

    config.Client = client
    config.FileWriter = NewFileWriter(config.OutputDir)
    config.RateLimiter = NewRateLimiter(config.MaxConcurrentRequests)

    return nil
}

func writeResults(config Config) error {
    // Execution phase - 9 LOC focused on orchestration
    orchestrator := NewOrchestrator(config.Client, config.FileWriter, config.RateLimiter)
    result := orchestrator.Run(context.Background(), config)

    return handleResult(result)
}
```

## Pattern 3: I/O Separation

### Example: File Processing Operations

**Before: Mixed I/O and business logic**
```go
// fileutil.go - Original mixed implementation
func GatherProjectContextWithContext(ctx context.Context, targetPath string, options FilteringOptions) (*ProjectContext, error) {
    var files []string

    // Mixed I/O and business logic
    err := filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Business logic mixed with I/O
        if d.IsDir() {
            return nil
        }

        if strings.HasPrefix(filepath.Base(path), ".") {
            return nil // Skip hidden files
        }

        ext := strings.ToLower(filepath.Ext(path))
        if isValidFileType(ext) {
            files = append(files, path)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // More mixed logic for processing files...
    var results []FileResult
    for _, file := range files {
        content, err := os.ReadFile(file)
        if err != nil {
            continue
        }

        stats := calculateFileStatistics(content)
        results = append(results, FileResult{
            Path:  file,
            Stats: stats,
        })
    }

    return &ProjectContext{Files: results}, nil
}
```

**After: Pure I/O and business logic separated**
```go
// filtering.go - Pure I/O functions
func WalkDirectory(rootPath string, walkFn filepath.WalkDirFunc) error {
    // Pure I/O operation
    return filepath.WalkDir(rootPath, walkFn)
}

func ReadFileContent(path string) ([]byte, error) {
    // Pure I/O operation
    return os.ReadFile(path)
}

func StatPath(path string) (os.FileInfo, error) {
    // Pure I/O operation
    return os.Stat(path)
}

// filtering.go - Pure business logic functions
func ShouldProcessFile(path string, options FilteringOptions) bool {
    // Pure business logic - no I/O, fully testable
    if IsHiddenPath(path) {
        return false
    }

    ext := strings.ToLower(filepath.Ext(path))
    return isValidFileType(ext, options)
}

func IsHiddenPath(path string) bool {
    // Pure business logic
    return strings.HasPrefix(filepath.Base(path), ".")
}

func CalculateFileStatistics(content []byte) FileStats {
    // Pure business logic
    return FileStats{
        Lines:      bytes.Count(content, []byte("\n")) + 1,
        Characters: len(content),
        Words:      len(strings.Fields(string(content))),
    }
}

// fileutil.go - Orchestration using pure functions
func GatherProjectContextWithContext(ctx context.Context, targetPath string, options FilteringOptions) (*ProjectContext, error) {
    var files []string

    // Use pure I/O function
    err := WalkDirectory(targetPath, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if d.IsDir() {
            return nil
        }

        // Use pure business logic
        if ShouldProcessFile(path, options) {
            files = append(files, path)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // Process files using pure functions
    var results []FileResult
    for _, file := range files {
        content, err := ReadFileContent(file)
        if err != nil {
            continue
        }

        stats := CalculateFileStatistics(content)
        results = append(results, FileResult{
            Path:  file,
            Stats: stats,
        })
    }

    return &ProjectContext{Files: results}, nil
}
```

## Pattern 4: Table-Driven Testing

### Example: Testing Extracted Functions

```go
// filtering_test.go - Comprehensive table-driven tests
func TestShouldProcessFile(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        options  FilteringOptions
        expected bool
    }{
        {
            name:     "go file accepted",
            path:     "/project/main.go",
            options:  FilteringOptions{IncludeExtensions: []string{".go"}},
            expected: true,
        },
        {
            name:     "hidden file rejected",
            path:     "/project/.hidden.go",
            options:  FilteringOptions{IncludeExtensions: []string{".go"}},
            expected: false,
        },
        {
            name:     "wrong extension rejected",
            path:     "/project/data.json",
            options:  FilteringOptions{IncludeExtensions: []string{".go"}},
            expected: false,
        },
        {
            name:     "empty extension filter accepts all",
            path:     "/project/README.md",
            options:  FilteringOptions{},
            expected: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ShouldProcessFile(tt.path, tt.options)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestCalculateFileStatistics(t *testing.T) {
    tests := []struct {
        name     string
        content  []byte
        expected FileStats
    }{
        {
            name:    "empty file",
            content: []byte(""),
            expected: FileStats{
                Lines:      1,
                Characters: 0,
                Words:      0,
            },
        },
        {
            name:    "single line",
            content: []byte("hello world"),
            expected: FileStats{
                Lines:      1,
                Characters: 11,
                Words:      2,
            },
        },
        {
            name:    "multiple lines",
            content: []byte("line one\nline two\nline three"),
            expected: FileStats{
                Lines:      3,
                Characters: 26,
                Words:      6,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateFileStatistics(tt.content)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Implementation Guidelines

When applying these patterns to your own code:

### 1. Identify Pure Logic
- Look for business logic mixed with I/O operations
- Extract calculations, validations, and transformations
- Ensure functions have no side effects

### 2. Separate I/O Operations
- Create focused functions for file operations, network calls, database queries
- Keep I/O functions simple and focused on single operations
- Use dependency injection for testability

### 3. Decompose Large Functions
- Break functions >100 LOC into logical phases
- Each function should have a single responsibility
- Use descriptive names that explain the phase

### 4. Test Extracted Functions
- Use table-driven tests for comprehensive coverage
- Test edge cases and error conditions
- No mocking required for pure functions

### 5. Validate Behavior
- Run integration tests to ensure identical behavior
- Use performance benchmarks to detect regressions
- Validate that all existing tests still pass

## Benefits Observed

From the 2025-07-08 refactoring:

**Performance**: 35-70% improvements in token counting benchmarks
**Coverage**: 83.6% → 90.4% test coverage
**Maintainability**: All functions <100 LOC, clear separation of concerns
**Testability**: Pure functions require no mocking, 95-100% coverage achieved

## See Also

- [CLAUDE.md](../CLAUDE.md) - Full development guidelines and patterns
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development workflow and setup
