# Plan Details

# PLAN: Clean Up Logging System

## Overview

This plan addresses GitHub issue #87: "clean up logging!" The goal is to transform the current "crazy loud and noisy" JSON logging into a clean, user-friendly console output while maintaining detailed structured logging for debugging purposes.

## Problem Statement

Current issues with the logging system:
- All output is structured JSON, making it hard to parse visually
- No clear progress indication during model processing
- Important information (models being used, status) is buried in JSON
- No separation between user-facing output and debug information

## Goals

1. **Clean Console Output**: Default output should be "tight and punchy and clean and useful"
2. **Progress Visibility**: Show real-time progress of model processing
3. **Debug Capability**: Maintain full structured logging when needed
4. **CI/CD Compatibility**: Ensure clean output in non-interactive environments
5. **Minimal Disruption**: Leverage existing patterns and avoid breaking changes

## Non-Goals

- Token counting implementation (future enhancement)
- Complex progress bars or animations
- Complete logging system rewrite
- Changes to audit logging functionality

## Proposed Solution

### Architecture Overview

We'll implement a **dual-output system** that separates user-facing console output from structured logging:

```
┌─────────────────┐
│   User Input    │
└────────┬────────┘
         │
┌────────▼────────┐
│  CLI Flags      │──► --quiet, --debug, --json-logs
└────────┬────────┘
         │
┌────────▼────────┐
│ Console Writer  │──► Clean output to stdout
├─────────────────┤
│ Slog Logger     │──► JSON logs to stderr/file
└─────────────────┘
```

### Key Components

#### 1. Console Writer (`internal/logutil/console_writer.go`)

A new component responsible for human-friendly output:

```go
type ConsoleWriter interface {
    // Progress reporting
    StartProcessing(modelCount int)
    ModelQueued(modelName string)
    ModelStarted(modelName string)
    ModelCompleted(modelName string, duration time.Duration, err error)
    ModelRateLimited(modelName string, delay time.Duration)

    // Status updates
    SynthesisStarted()
    SynthesisCompleted(outputPath string)

    // Control
    SetQuiet(quiet bool)
    IsInteractive() bool
}
```

#### 2. Progress Display Modes

Based on environment detection:

**Interactive Mode** (TTY detected):
```
🚀 Processing 3 models...
[1/3] gemini-1.5-flash: processing...
[1/3] gemini-1.5-flash: ✓ completed (0.8s)
[2/3] gpt-4: processing...
[2/3] gpt-4: ✓ completed (1.2s)
[3/3] claude-3-opus: rate limited, waiting 2s...
[3/3] claude-3-opus: processing...
[3/3] claude-3-opus: ✓ completed (1.0s)

📄 Synthesizing results...
✨ Done! Output saved to: output_20240115_120000/
```

**CI Mode** (non-TTY or CI=true):
```
Starting processing with 3 models
Processing model 1/3: gemini-1.5-flash
Completed model 1/3: gemini-1.5-flash (0.8s)
Processing model 2/3: gpt-4
Completed model 2/3: gpt-4 (1.2s)
Rate limited for model 3/3: claude-3-opus (waiting 2s)
Processing model 3/3: claude-3-opus
Completed model 3/3: claude-3-opus (1.0s)
Starting synthesis
Synthesis complete. Output: output_20240115_120000/
```

#### 3. Output Control Flags

Expand CLI flags for output control:

```go
// Existing
--verbose, -v     Set log level to DEBUG
--log-level       Set logging level (debug, info, warn, error)

// New
--quiet, -q       Suppress console output (errors only)
--json-logs       Show JSON logs on stderr (current behavior)
--no-progress     Disable progress indicators
```

#### 4. Logger Configuration Updates

Modify `SetupLogging()` to route output appropriately:

```go
func SetupLogging(config *Config) (logutil.Logger, error) {
    // Determine output mode
    jsonToStderr := config.JsonLogs || config.Debug

    if jsonToStderr {
        // Current behavior: JSON to stderr
        return logutil.NewSlogLogger(slog.LevelInfo), nil
    } else {
        // New behavior: JSON to file only
        logFile := filepath.Join(config.OutputDir, "thinktank.log")
        return logutil.NewFileLogger(logFile, slog.LevelInfo), nil
    }
}
```

### Implementation Plan

#### Phase 1: Console Writer Foundation (Week 1)
1. Create `ConsoleWriter` interface and implementation
2. Add TTY detection and CI environment checks
3. Implement basic status messages (start, complete)
4. Add unit tests for different environments

#### Phase 2: CLI Flag Integration (Week 1)
1. Add new CLI flags (--quiet, --json-logs, --no-progress)
2. Update `SetupLogging()` to handle output routing
3. Ensure backward compatibility (default = new behavior)
4. Update help text and documentation

#### Phase 3: Orchestrator Integration (Week 2)
1. Inject `ConsoleWriter` into orchestrator
2. Add console output calls at key points:
   - Model processing start/end
   - Rate limiting delays
   - Synthesis start/end
3. Maintain existing structured logging
4. Add integration tests

#### Phase 4: Progress Tracking (Week 2)
1. Track model processing order (1/N, 2/N, etc.)
2. Calculate and display durations
3. Handle concurrent model updates safely
4. Format output based on terminal width

#### Phase 5: Error Handling & Edge Cases (Week 3)
1. Handle partial failures gracefully
2. Ensure clean output on interrupt (Ctrl+C)
3. Test with various terminal types
4. Verify CI/CD compatibility

#### Phase 6: Polish & Documentation (Week 3)
1. Add configuration examples
2. Update README with new flags
3. Add troubleshooting guide
4. Performance testing

### Technical Decisions

1. **No External Dependencies**: Implement progress tracking ourselves to avoid dependency bloat
2. **Stderr for JSON**: When JSON logs are shown, route to stderr to keep stdout clean
3. **File Logging Default**: In new mode, JSON logs go to file by default
4. **Preserve Correlation IDs**: Console output doesn't show IDs but they remain in JSON logs
5. **Thread Safety**: Use mutex protection for concurrent console updates

### Testing Strategy

1. **Unit Tests**:
   - Console writer with mock outputs
   - TTY detection logic
   - Flag parsing and configuration

2. **Integration Tests**:
   - Full flow with different flag combinations
   - CI environment simulation
   - Concurrent model processing

3. **Manual Testing**:
   - Various terminal emulators
   - CI/CD environments (GitHub Actions)
   - Different OS platforms

### Migration Path

1. **Default Behavior Change**: New installations get clean output by default
2. **Opt-in for Old Behavior**: Use `--json-logs` flag for current behavior
3. **Documentation**: Clear upgrade notes explaining the change
4. **Transition Period**: Support both modes for several releases

### Future Enhancements

1. **Token Counting**: Extract and display token usage from providers
2. **Progress Bars**: Optional rich progress bars using a library
3. **Colored Output**: Add color coding for different message types
4. **Output Formats**: Support for different output formats (table, CSV)
5. **Live Updates**: WebSocket endpoint for real-time progress

### Success Metrics

1. **User Feedback**: Positive response to cleaner output
2. **Support Tickets**: Reduction in logging-related issues
3. **Performance**: No regression in execution time
4. **Compatibility**: Zero CI/CD pipeline breaks

### Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing scripts | High | Provide --json-logs flag for compatibility |
| CI/CD output issues | Medium | Thorough testing in CI environments |
| Performance regression | Low | Benchmark before/after implementation |
| Complex terminal handling | Medium | Fallback to simple output on errors |

### Example Usage

```bash
# Default: Clean console output
$ thinktank --instructions "analyze code" ./src

# Debug mode: Console output + JSON logs to stderr
$ thinktank --debug --instructions "analyze code" ./src

# Quiet mode: Minimal output
$ thinktank --quiet --instructions "analyze code" ./src

# Old behavior: JSON logs to stderr
$ thinktank --json-logs --instructions "analyze code" ./src

# CI mode: Detected automatically
$ CI=true thinktank --instructions "analyze code" ./src
```

## Approval

This plan provides a clear path to address the logging issues while maintaining backward compatibility and system reliability. The phased approach allows for incremental delivery and testing.

**Estimated Timeline**: 3 weeks
**Estimated Effort**: 1 engineer
**Priority**: High (labeled as priority:critical)

## Task Breakdown Requirements
- Create atomic, independent tasks
- Ensure proper dependency mapping
- Include verification steps
- Follow project task ID and formatting conventions
