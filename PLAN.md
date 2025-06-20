# Modern Clean CLI Output - Product Requirements & Implementation Plan

## Executive Summary

Transform thinktank's CLI output from emoji-heavy to a modern, clean aesthetic inspired by ripgrep, eza, and bat. This design prioritizes information density, visual clarity, and professional polish while maintaining maximum utility and universal terminal compatibility.

## Design Philosophy

### Core Principles
1. **Information Density with Visual Clarity** - Every character serves a purpose
2. **Functional Aesthetics** - Beauty through purposeful design, not decoration
3. **Professional Polish** - Feels like a first-class developer tool
4. **Universal Compatibility** - Graceful degradation across all terminal environments
5. **Scannable Layout** - Quick visual parsing of status and results

### Visual Language
- **Minimal Unicode Symbols**: `●` `✓` `✗` `⚠` `─` instead of emojis
- **Right-aligned Status**: Consistent column alignment for quick scanning
- **Subtle Separators**: Clean section breaks without visual noise
- **Functional Color**: Color conveys meaning, not decoration
- **Human-readable Values**: File sizes, durations in readable formats

## Detailed Design Specifications

### 1. Process Initialization
```
Starting thinktank processing...
Gathering project files...
Processing 3 models...
```

**Specifications:**
- Clean, declarative statements
- No decorative symbols for basic status
- Progressive disclosure of information

### 2. Model Processing Display

#### Single Model
```
gemini-2.5-pro                          processing...
gemini-2.5-pro                          ✓ 68.5s
```

#### Multiple Models
```
gemini-2.5-pro                          processing...
claude-3-5-sonnet                       processing...
gpt-4o                                   processing...

gemini-2.5-pro                          ✓ 68.5s
claude-3-5-sonnet                       ✓ 45.2s
gpt-4o                                   ✗ rate limited
```

**Specifications:**
- Model names left-aligned, fixed width column
- Status right-aligned for easy scanning
- Processing indicator shows during execution
- Success: `✓ duration`
- Failure: `✗ reason`
- Rate limited: `⚠ rate limited (retry in 30s)`

### 3. File Operations
```
Saving individual outputs...
Saved 2 individual outputs to ./thinktank_20250619_134403_000457001/
```

**Specifications:**
- Clear, active voice statements
- File paths shown without decorative prefixes
- Success confirmation with count

### 4. Summary Section

#### Success Case
```
SUMMARY
───────
● 3 models processed
● 2 successful, 1 failed
● Output directory: ./thinktank_20250619_134403_000457001/

OUTPUT FILES
────────────
  gemini-2.5-pro.md                     4.2K
  claude-3-5-sonnet.md                  3.8K

FAILED MODELS
─────────────
  gpt-4o                                rate limited
```

#### Synthesis Case
```
SUMMARY
───────
● 3 models processed
● 2 successful, 1 failed
● Synthesis: ✓ completed
● Output directory: ./thinktank_20250619_134403_000457001/

OUTPUT FILES
────────────
  synthesis.md                          8.1K
  gemini-2.5-pro.md                     4.2K
  claude-3-5-sonnet.md                  3.8K

FAILED MODELS
─────────────
  gpt-4o                                rate limited
```

**Specifications:**
- Section headers in UPPERCASE with underline separators
- Bullet points using `●` for consistency
- File listings with right-aligned human-readable sizes
- Failed models section only appears when relevant

### 5. Error Scenarios

#### All Models Failed
```
Processing 3 models...

gemini-2.5-pro                          ✗ API error
claude-3-5-sonnet                       ✗ timeout
gpt-4o                                   ✗ rate limited

SUMMARY
───────
● 3 models processed
● 0 successful, 3 failed
● No outputs generated

All models failed. Check your API keys and network connection.
```

#### Partial Success
```
SUMMARY
───────
● 3 models processed
● 1 successful, 2 failed
● Output directory: ./thinktank_20250619_134403_000457001/

OUTPUT FILES
────────────
  gemini-2.5-pro.md                     4.2K

FAILED MODELS
─────────────
  claude-3-5-sonnet                     API error
  gpt-4o                                rate limited

Note: Continuing with available outputs (1/3 models succeeded)
```

## Color Scheme Specification

### Interactive Terminal Colors
```go
type ColorScheme struct {
    ModelName      string // Subtle blue (#5294cf)
    Success        string // Green (#5cb85c)
    Warning        string // Yellow (#f0ad4e)
    Error          string // Red (#d9534f)
    Duration       string // Gray (#777777)
    FileSize       string // Gray (#777777)
    FilePath       string // Default/white
    SectionHeader  string // Bold white
    Separator      string // Gray (#555555)
    Symbol         string // Default/white
}
```

### CI/Non-Interactive
- All colors disabled
- Symbols remain for semantic meaning
- Layout preserved exactly

## Technical Implementation Requirements

### 1. Console Writer Interface Updates

```go
type ConsoleWriter interface {
    // Existing methods updated
    StartProcessing(modelCount int)
    ModelStarted(modelIndex, totalModels int, modelName string)
    ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration)
    ModelFailed(modelIndex, totalModels int, modelName string, reason string)
    ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration)

    // New methods
    ShowProcessingLine(modelName string) // Shows "model processing..." line
    UpdateProcessingLine(modelName string, status string) // Updates in-place
    ShowFileOperations(message string)
    ShowSummarySection(summary SummaryData)
    ShowOutputFiles(files []OutputFile)
    ShowFailedModels(failed []FailedModel)
}
```

### 2. Data Structures

```go
type SummaryData struct {
    ModelsProcessed   int
    SuccessfulModels  int
    FailedModels      int
    SynthesisStatus   string // "completed", "failed", "skipped"
    OutputDirectory   string
}

type OutputFile struct {
    Name string
    Path string
    Size int64
}

type FailedModel struct {
    Name   string
    Reason string
}
```

### 3. File Size Formatting

```go
func FormatFileSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%dB", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f%c", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

### 4. Terminal Width Handling

```go
type LayoutConfig struct {
    TerminalWidth    int
    ModelNameWidth   int
    StatusWidth      int
    FileNameWidth    int
    FileSizeWidth    int
}

func CalculateLayout(terminalWidth int) LayoutConfig {
    // Responsive layout calculation
    // Minimum 50 chars, optimal 80+ chars
}
```

## Implementation Plan

### Phase 1: Core Infrastructure (1-2 days)
1. **Update ConsoleWriter interface** - Add new methods, update signatures
2. **Implement file size formatting** - Human-readable size utility
3. **Add terminal width detection** - Responsive layout foundation
4. **Create color scheme system** - Environment-aware color handling

### Phase 2: Basic Output Transformation (2-3 days)
1. **Update model processing display** - New aligned format
2. **Implement status indicators** - Unicode symbols replacing emojis
3. **Update file operations messaging** - Clean, declarative statements
4. **Basic summary section** - New structured format

### Phase 3: Advanced Features (1-2 days)
1. **File listing with sizes** - Output files section
2. **Failed models section** - Error reporting
3. **Synthesis integration** - Synthesis status in summary
4. **Error scenario handling** - All failure cases

### Phase 4: Polish & Testing (1-2 days)
1. **Color scheme implementation** - Full color support
2. **CI environment testing** - No-color fallbacks
3. **Terminal width testing** - Responsive behavior
4. **Edge case handling** - Long names, narrow terminals

## Testing Strategy

### Unit Tests
```go
func TestConsoleWriter_ModelProgress(t *testing.T) {
    // Test aligned output formatting
    // Test status symbol selection
    // Test duration formatting
}

func TestFileSize_Formatting(t *testing.T) {
    // Test human-readable size conversion
    // Test edge cases (0 bytes, very large files)
}

func TestLayout_Responsive(t *testing.T) {
    // Test terminal width adaptation
    // Test minimum width handling
}
```

### Integration Tests
```go
func TestE2E_ModernCleanOutput(t *testing.T) {
    // Test complete workflow output
    // Test multi-model scenarios
    // Test error scenarios
}

func TestEnvironment_Compatibility(t *testing.T) {
    // Test CI vs interactive environments
    // Test color vs no-color output
    // Test various terminal widths
}
```

### Manual Testing Scenarios
1. **Single model success** - Verify clean, aligned output
2. **Multiple model mixed results** - Test status alignment
3. **All models failed** - Verify error messaging
4. **Synthesis workflow** - Test synthesis integration
5. **Narrow terminal** - Test responsive layout
6. **CI environment** - Verify no-color output

## File Changes Required

### Primary Files
1. **`internal/logutil/console_writer.go`** - Complete rewrite of output methods
2. **`internal/thinktank/orchestrator/summary_writer.go`** - New summary format
3. **`internal/thinktank/orchestrator/orchestrator.go`** - Update output calls

### New Files
1. **`internal/logutil/formatting.go`** - File size, duration formatting utilities
2. **`internal/logutil/layout.go`** - Terminal width and layout calculations
3. **`internal/logutil/colors.go`** - Color scheme definitions

### Test Files
1. **`internal/logutil/console_writer_modern_test.go`** - Comprehensive output testing
2. **`internal/logutil/formatting_test.go`** - Utility function tests
3. **`internal/e2e/cli_modern_output_test.go`** - End-to-end output verification

## Success Metrics

### Functional Requirements
- [ ] All emoji usage eliminated
- [ ] Unicode symbols render correctly across terminals
- [ ] Color scheme adapts to environment (interactive vs CI)
- [ ] File sizes display in human-readable format
- [ ] Layout responsive to terminal width
- [ ] Status alignment consistent across all scenarios

### Quality Requirements
- [ ] No degradation in information density
- [ ] Improved visual scanning of results
- [ ] Professional aesthetic comparable to ripgrep/eza/bat
- [ ] Zero breaking changes to underlying functionality
- [ ] 100% test coverage for new formatting code

### User Experience Requirements
- [ ] Faster visual parsing of results
- [ ] Clear distinction between success/failure states
- [ ] Easy identification of output files and locations
- [ ] Intuitive error messaging and next steps

## Risk Mitigation

### Technical Risks
1. **Unicode compatibility** - Fallback ASCII symbols for problematic terminals
2. **Terminal width detection** - Graceful handling of detection failures
3. **Color detection** - Conservative color usage, robust fallbacks

### Implementation Risks
1. **Test coverage** - Comprehensive testing across environments
2. **Backwards compatibility** - Feature flags for rollback if needed
3. **Performance impact** - Benchmark formatting operations

### User Experience Risks
1. **Terminal diversity** - Testing across multiple terminal emulators
2. **User preference** - Clear documentation of new format
3. **Accessibility** - Ensure color-blind friendly design

## Future Enhancements

### Phase 2 Features (Post-Launch)
1. **Theme system** - User-configurable color schemes
2. **Progress indicators** - Real-time progress for long-running models
3. **Interactive features** - Model selection, retry options
4. **Output filtering** - Show/hide failed models, file details

### Advanced Features
1. **Terminal dashboard** - Live updating status display
2. **Notification integration** - Desktop notifications for completion
3. **Export formats** - JSON/CSV summary output options
4. **Performance metrics** - Token usage, cost estimation display

---

**Document Version**: 1.0
**Last Updated**: 2025-06-19
**Status**: Ready for Implementation
