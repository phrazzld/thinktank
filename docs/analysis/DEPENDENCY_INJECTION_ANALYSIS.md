# Dependency Injection Patterns Analysis

## Overview

This document analyzes existing dependency injection patterns in the thinktank codebase to understand how to properly design the RunConfig struct and dependency interfaces for the main() function refactoring.

## Existing Dependency Injection Examples

### 1. thinktank.Execute() Function - Gold Standard Pattern

**Location**: `internal/thinktank/app.go` lines 28-37

**Function Signature**:
```go
func Execute(
    ctx context.Context,
    cliConfig *config.CliConfig,
    logger logutil.LoggerInterface,
    auditLogger auditlog.AuditLogger,
    apiService interfaces.APIService,
    consoleWriter logutil.ConsoleWriter,
) (err error)
```

**Key Insights**:
- **All dependencies explicit**: No hidden dependencies on global state
- **Interfaces over implementations**: Uses interface types for testability
- **Context first**: Context is the first parameter (Go convention)
- **Configuration separated**: CliConfig separate from operational dependencies
- **Error return**: Returns error instead of calling os.Exit()
- **Clean signature**: Each parameter has a clear purpose

**Dependency Types**:
1. **Context**: Request scoping and cancellation
2. **Configuration**: `*config.CliConfig` - parsed configuration data
3. **Logging**: `logutil.LoggerInterface` - structured logging
4. **Audit**: `auditlog.AuditLogger` - audit trail logging
5. **API Service**: `interfaces.APIService` - LLM API operations
6. **Console Output**: `logutil.ConsoleWriter` - user-facing output

### 2. Orchestrator Constructor Pattern

**Location**: `internal/thinktank/app.go` lines 228-248

**Function Signature**:
```go
func(
    apiService interfaces.APIService,
    contextGatherer interfaces.ContextGatherer,
    fileWriter interfaces.FileWriter,
    auditLogger auditlog.AuditLogger,
    rateLimiter *ratelimit.RateLimiter,
    config *config.CliConfig,
    logger logutil.LoggerInterface,
    consoleWriter logutil.ConsoleWriter,
) Orchestrator
```

**Key Insights**:
- **Factory function pattern**: Constructor injection for testability
- **Multiple interface dependencies**: Each component has its own interface
- **Rate limiter injection**: Even utilities are injected
- **Configuration last**: Config comes after operational dependencies
- **Interface return**: Returns interface for testability

### 3. Interface Design Patterns

**Location**: `internal/thinktank/interfaces/interfaces.go`

**APIService Interface** (lines 18-99):
- **Comprehensive**: Covers all API operations
- **Context aware**: All methods accept context
- **Error explicit**: Clear error handling patterns
- **Self-contained**: No dependencies on other interfaces

**ContextGatherer Interface** (lines 121-128):
- **Single responsibility**: Only handles context gathering
- **Configuration struct**: Uses `GatherConfig` struct for parameters
- **Stats return**: Returns structured statistics

**FileWriter Interface** (lines 130-134):
- **Minimal**: Single method for single responsibility
- **Context aware**: Uses context for cancellation
- **Clear purpose**: File writing only

### 4. Adapter Pattern Usage

**Location**: `internal/thinktank/app.go`

```go
apiServiceAdapter := &APIServiceAdapter{APIService: apiService}
// Note: ContextGathererAdapter was removed in issue #121 - NewContextGatherer
// now directly implements interfaces.ContextGatherer
fileWriterAdapter := &FileWriterAdapter{FileWriter: fileWriter}
```

**Key Insights**:
- **Interface adaptation**: Converts between interface versions
- **Composition pattern**: Wraps existing implementations
- **Dependency bridging**: Connects different package interfaces
- **Direct implementation**: Where possible, implementations directly satisfy interfaces (preferred)

## Existing Logger Interface Patterns

### logutil.LoggerInterface

**Key Methods**:
- `WithContext(ctx context.Context) LoggerInterface`
- `ErrorContext(ctx context.Context, format string, args ...interface{})`
- `InfoContext(ctx context.Context, format string, args ...interface{})`
- `DebugContext(ctx context.Context, format string, args ...interface{})`

**Pattern**: Immutable interface with context propagation

### auditlog.AuditLogger

**Key Methods**:
- `Log(ctx context.Context, entry auditlog.AuditEntry) error`
- `LogOp(ctx context.Context, operation, result string, request, response map[string]interface{}, err error) error`
- `Close() error`

**Pattern**: Resource management with lifecycle methods

## Dependency Analysis for RunConfig Design

### 1. Required Dependencies for main() Business Logic

**Based on main() audit (MAIN_FUNCTION_AUDIT.md)**:

1. **Configuration**: `*config.CliConfig` (already parsed)
2. **Context**: `context.Context` (with timeout and cancellation)
3. **Logger**: `logutil.LoggerInterface` (for business logic logging)
4. **Audit Logger**: `auditlog.AuditLogger` (for audit trail)
5. **API Service**: `interfaces.APIService` (for LLM operations)
6. **Console Writer**: `logutil.ConsoleWriter` (for user output)
7. **File System**: Interface for file operations (new requirement)
8. **Exit Handler**: Interface for exit behavior (new requirement)

### 2. New Interface Requirements

**FileSystem Interface** (not currently defined):
```go
type FileSystem interface {
    CreateTemp(dir, pattern string) (*os.File, error)
    WriteFile(filename string, data []byte, perm os.FileMode) error
    ReadFile(filename string) ([]byte, error)
    Remove(name string) error
    MkdirAll(path string, perm os.FileMode) error
    OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}
```

**ExitHandler Interface** (not currently defined):
```go
type ExitHandler interface {
    Exit(code int)
    HandleError(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string)
}
```

## Recommended RunConfig Design

### Based on thinktank.Execute() Pattern

```go
type RunConfig struct {
    // Context (first parameter - Go convention)
    Context context.Context

    // Configuration (parsed from flags)
    Config *config.CliConfig

    // Core operational dependencies (following Execute pattern)
    Logger        logutil.LoggerInterface
    AuditLogger   auditlog.AuditLogger
    APIService    interfaces.APIService
    ConsoleWriter logutil.ConsoleWriter

    // New dependencies for main() logic
    FileSystem    FileSystem
    ExitHandler   ExitHandler
}
```

### RunResult Design

```go
type RunResult struct {
    ExitCode int
    Error    error

    // Optional: Additional metadata for testing
    Stats *ExecutionStats
}

type ExecutionStats struct {
    FilesProcessed int
    APICalls       int
    Duration       time.Duration
}
```

## Implementation Strategy

### 1. Follow Execute() Pattern Exactly

The `thinktank.Execute()` function demonstrates the perfect dependency injection pattern:
- All dependencies are explicit parameters
- Interfaces are used for testability
- No hidden global state dependencies
- Clean error handling without os.Exit()

### 2. Create Matching Interface Abstractions

For dependencies not currently abstracted:
- **FileSystem**: Abstract os package file operations
- **ExitHandler**: Abstract os.Exit() and error handling

### 3. Maintain Interface Consistency

All new interfaces should follow existing patterns:
- Context as first parameter
- Clear error handling
- Single responsibility
- Comprehensive but minimal

### 4. Use Adapter Pattern for Integration

Where needed, use adapters to bridge between interfaces:
- Main RunConfig → Execute parameters
- New interfaces → existing implementations

## Testing Strategy

### 1. Mock All Dependencies

Following the existing pattern:
```go
type MockFileSystem struct {
    Files     map[string][]byte
    Errors    map[string]error
    CallLog   []string
}

type MockExitHandler struct {
    ExitCodes []int
    Errors    []error
}
```

### 2. Use Dependency Injection for Tests

```go
func TestRun(t *testing.T) {
    runConfig := &RunConfig{
        Context:       context.Background(),
        Config:        testConfig,
        Logger:        mockLogger,
        AuditLogger:   mockAuditLogger,
        APIService:    mockAPIService,
        ConsoleWriter: mockConsoleWriter,
        FileSystem:    mockFileSystem,
        ExitHandler:   mockExitHandler,
    }

    result := Run(runConfig)

    // Test assertions on result and mock state
}
```

## Conclusion

The existing `thinktank.Execute()` function provides the perfect template for our RunConfig design. Key principles:

1. **Explicit Dependencies**: All dependencies as parameters
2. **Interface Abstraction**: Use interfaces for testability
3. **Context Propagation**: Context-aware operations
4. **Clean Error Handling**: Return errors instead of calling os.Exit()
5. **Single Responsibility**: Each dependency has a clear purpose

By following this established pattern, we ensure consistency with the existing codebase while achieving our testability goals.

**Next Steps**: Define the new FileSystem and ExitHandler interfaces, then create the RunConfig struct following the Execute() pattern.
