# Main() Function Architecture Audit

## Overview

The `Main()` function in `internal/cli/main.go` (lines 211-338) serves as the entry point for the thinktank CLI application. This audit documents all external dependencies, side effects, and control flow patterns to enable architectural refactoring for better testability.

## Function Structure Analysis

**Total Lines**: 128 lines (211-338)
**Mixed Concerns**: Context setup, dependency initialization, business logic, error handling
**Exit Points**: 3 locations (lines 220, 283, 320 via handleError, plus implicit success exit)

## External Dependencies Identified

### 1. Command Line & Environment Dependencies
- **`ParseFlags()`** (line 216)
  - Reads `os.Args` globally
  - Returns `*config.CliConfig` and error
  - **Side Effect**: Global state access
  - **Testing Impact**: Requires subprocess execution or global state manipulation

- **`os.Stderr`** (line 219)
  - Direct stderr output for early flag errors
  - **Side Effect**: Console output
  - **Testing Impact**: Requires output capture or mocking

- **`os.Exit()`** (line 220)
  - Immediate process termination for flag errors
  - **Side Effect**: Process termination
  - **Testing Impact**: Prevents function return, requires subprocess testing

### 2. Context & Time Dependencies
- **`context.Background()`** (line 224)
  - Creates root context
  - **Side Effect**: Time-dependent context creation

- **`context.WithTimeout()`** (line 225)
  - Uses `config.Timeout` for deadline
  - **Side Effect**: Time-dependent behavior
  - **Testing Impact**: Requires time mocking or timeout handling

- **`setupGracefulShutdown()`** (line 229)
  - Sets up SIGINT/SIGTERM signal handling
  - **Side Effect**: Global signal handler registration
  - **Testing Impact**: Requires signal mocking or subprocess execution

### 3. Logging Dependencies
- **`SetupLogging(config)`** (line 238)
  - Creates logger based on configuration
  - **Side Effect**: May create log files on filesystem
  - **Testing Impact**: Requires filesystem mocking or temporary directories

- **`logutil.WithCorrelationID()`** (line 234)
  - Generates UUID for correlation tracking
  - **Side Effect**: Random UUID generation
  - **Testing Impact**: Non-deterministic output

### 4. Filesystem Dependencies
- **`auditlog.NewFileAuditLogger()`** (line 246)
  - Creates audit log file if `config.AuditLogFile` specified
  - **Side Effect**: File creation and write operations
  - **Testing Impact**: Requires filesystem mocking or temporary files

- **File handle management** (line 261)
  - `defer func() { _ = auditLogger.Close() }()`
  - **Side Effect**: File descriptor cleanup
  - **Testing Impact**: Requires proper resource cleanup verification

### 5. Error Handling Dependencies
- **`handleError()`** (lines 283, 320)
  - Centralizes error handling and exit code determination
  - Calls `os.Exit()` internally
  - **Side Effect**: Process termination
  - **Testing Impact**: Prevents normal function return

## Control Flow Patterns

### 1. Linear Execution Flow
```
ParseFlags → Context Setup → Logging Setup → Audit Setup →
Validation → Service Init → Business Logic → Success/Error Handling
```

### 2. Error Exit Points
- **Early Exit** (line 220): Flag parsing failures
- **Validation Exit** (line 283): Input validation failures
- **Execution Exit** (line 320): Business logic failures
- **Success Exit** (line 338): Normal completion

### 3. Conditional Logic Branches
- **Audit Logger Setup** (lines 245-258): File vs NoOp based on config
- **Partial Success Handling** (lines 302-316): Special exit logic for tolerant mode
- **Error Categorization** (line 282): LLM error wrapping

## Global State Mutations

### 1. Signal Handler Registration
- `setupGracefulShutdown()` modifies global signal handling state
- **Impact**: Affects entire process signal behavior
- **Testing Concern**: Global state persistence between tests

### 2. Logger Context Modification
- Multiple logger context attachments (lines 240, 241)
- **Impact**: Logger state changes throughout execution
- **Testing Concern**: Context propagation verification

### 3. Resource Lifecycle Management
- Audit logger creation and cleanup with defer
- **Impact**: File handle lifecycle spans entire function
- **Testing Concern**: Resource leak detection

## Dependency Injection Opportunities

### 1. Well-Designed Dependencies (Already Injectable)
- **`thinktank.Execute()`** (line 299)
  - Perfect dependency injection pattern
  - Takes all dependencies as explicit parameters
  - **Model**: `ctx, config, logger, auditLogger, apiService, consoleWriter`

### 2. Extractable Dependencies
- **FlagParser Interface**: Replace `ParseFlags()` call
- **ContextFactory Interface**: Replace context creation logic
- **LoggerFactory Interface**: Replace `SetupLogging()` call
- **AuditLoggerFactory Interface**: Replace audit logger creation
- **ExitHandler Interface**: Replace `os.Exit()` and `handleError()` calls

## Test Coverage Analysis

### Current Subprocess Test Scenarios

**TestMainDryRun** covers:
- Basic dry-run execution (line 389: `--dry-run`)
- Audit logging behavior (line 394: `--audit-log-file`)
- Verbose output mode (line 397: `--verbose`)
- Quiet mode behavior (line 400: `--quiet`)

**TestMainConfigurationOptions** covers:
- Custom timeout (line 689: `--timeout 5s`)
- Rate limiting (line 691: `--rate-limit 30 --max-concurrent 3`)
- File permissions (line 693: `--dir-permissions 0755 --file-permissions 0644`)
- Multiple models (line 695: `--model gemini-2.5-pro --model gemini-2.5-flash`)
- File filtering (line 697: `--include .go,.md --exclude .exe,.bin`)

### Test Coverage Gaps
- **Error path coverage**: Limited error scenario testing due to subprocess complexity
- **Edge case coverage**: Difficult to test error conditions that require specific environment states
- **Integration boundaries**: Cannot easily test interaction between components

## Refactoring Strategy

### 1. Extract Core Business Logic
Create `Run(*RunConfig) *RunResult` function containing lines 223-337:
- Accept all dependencies as parameters
- Return structured result instead of calling `os.Exit()`
- Maintain identical business logic and error handling

### 2. Define Dependency Interfaces
```go
type RunConfig struct {
    Config        *config.CliConfig
    Logger        logutil.LoggerInterface
    AuditLogger   auditlog.AuditLogger
    APIService    interfaces.APIService
    ConsoleWriter *logutil.ConsoleWriter
    Context       context.Context
}

type RunResult struct {
    ExitCode int
    Error    error
}
```

### 3. Transform Main() into Thin Wrapper
- Keep only: flag parsing, context setup, dependency creation
- Call `Run()` with real dependencies
- Handle `RunResult` for exit code determination
- Reduce from ~128 lines to ~20 lines

### 4. Replace Subprocess Tests
Convert each subprocess test scenario to direct `Run()` function tests:
- Mock external dependencies (filesystem, exit handler)
- Test business logic directly with controlled inputs
- Verify behavior through `RunResult` inspection
- Eliminate subprocess execution complexity

## Architecture Quality Assessment

### Current Issues
- **Mixed Concerns**: Setup, business logic, and error handling intertwined
- **Hard Dependencies**: Direct calls to `os.Exit()`, `os.Args`, filesystem
- **Untestable Design**: Requires subprocess execution for comprehensive testing
- **Hidden Dependencies**: Global state mutations not visible in function signature

### Proposed Improvements
- **Separation of Concerns**: Extract pure business logic from infrastructure setup
- **Explicit Dependencies**: All external dependencies injected through interfaces
- **Testable Design**: Direct function testing with mocked dependencies
- **Transparent Interface**: Function signature reveals all dependencies

## Risk Assessment

### Low Risk Changes
- **Dependency extraction**: Existing interfaces already demonstrate good patterns
- **Error handling preservation**: Maintain identical error categorization and handling
- **Business logic preservation**: No changes to core application behavior

### Rollback Strategy
- **Parallel implementation**: Keep existing `Main()` during transition
- **Gradual migration**: Replace subprocess tests incrementally
- **Quick revert**: Simple function call change if issues arise

## Conclusion

The current `Main()` function demonstrates good business logic but poor testability due to mixed concerns and hard dependencies. The `thinktank.Execute()` function already shows the correct dependency injection pattern. By extracting the business logic into a `Run()` function following this same pattern, we can eliminate subprocess test complexity while maintaining full functionality and improving test reliability.

**Next Steps**: Design dependency interfaces and begin business logic extraction following the proven `thinktank.Execute()` pattern.
