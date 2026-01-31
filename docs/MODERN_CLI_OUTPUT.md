# Modern Clean CLI Output - Documentation & Rollback Guide

## Overview

thinktank v2.0+ features a completely redesigned CLI output system that replaces emoji-heavy displays with professional, environment-aware formatting. This document provides comprehensive guidance on the new format, compatibility considerations, and rollback strategies.

## Design Philosophy

### Core Principles

**Professional Aesthetics**: Clean, scannable output inspired by modern CLI tools like ripgrep, eza, and bat
**Environment Awareness**: Automatic adaptation between interactive terminals and CI/automation environments
**Universal Compatibility**: Unicode symbols with ASCII fallbacks for maximum terminal compatibility
**Information Density**: Improved visual hierarchy without information loss
**Zero Breaking Changes**: All underlying functionality preserved

### Visual Design Language

- **UPPERCASE Headers**: Section headers use uppercase with consistent separators (`SUMMARY`, `OUTPUT FILES`)
- **Semantic Symbols**: Unicode symbols (✓ ✗ ⚠ ●) with ASCII alternatives ([OK] [X] [!] *)
- **Right-Aligned Status**: Processing status aligned to the right for easy scanning
- **Structured Sections**: Clear visual separation between different information types
- **Human-Readable Sizes**: File sizes displayed as "2.4K", "1.5M" instead of raw bytes

## Environment Detection & Adaptation

### Interactive Terminal Mode

**Triggers**: TTY detected AND no CI environment variables set

**Features**:
- Unicode symbols (✓ ✗ ⚠ ●)
- ANSI color codes (green success, red errors, yellow warnings)
- Rich formatting with responsive layout
- Real-time progress updates

**Example**:
```
Processing 3 models...
[1/3] gemini-3-flash: ✓ completed (2.3s)
[2/3] gpt-5.2: ✓ completed (1.8s)
[3/3] o3: ✗ rate limited

SUMMARY
───────
● 3 models processed
● 2 successful, 1 failed
● Synthesis: ✓ completed
● Output directory: ./thinktank_20250619_143022_7841
```

### CI/Automation Mode

**Triggers**: Any of the following environment variables set:
- `CI=true`
- `GITHUB_ACTIONS=true`
- `CONTINUOUS_INTEGRATION=true`
- Non-TTY environment (pipes, redirects)

**Features**:
- ASCII symbols ([OK] [X] [!] *)
- No ANSI color codes
- Clean, parseable output
- Consistent formatting across platforms

**Example**:
```
Processing 3 models...
Completed model 1/3: gemini-3-flash (2.3s)
Completed model 2/3: gpt-5.2 (1.8s)
Failed model 3/3: o3 (rate limited)

SUMMARY
-------
* 3 models processed
* 2 successful, 1 failed
* Synthesis: [OK] completed
* Output directory: ./thinktank_20250619_143022_7841
```

## Unicode Compatibility

### Unicode Support Detection

The system automatically detects Unicode support through:

1. **Locale Detection**: Checks `LANG`, `LC_ALL`, `LC_CTYPE` for UTF-8 indicators
2. **Terminal Detection**: Identifies modern terminals (VS Code, iTerm2, Windows Terminal)
3. **Environment Overrides**: Respects CI flags for ASCII-only output

### Fallback Strategy

When Unicode is not supported or detected:
- ✓ → [OK]
- ✗ → [X]
- ⚠ → [!]
- ● → *
- ──── → ----

### Manual Override

Users can force ASCII mode by setting environment variables:
```bash
export CI=true
thinktank --instructions task.txt ./src
```

## Rollback Strategy

### Option 1: Flag-Based Rollback (Recommended)

Use existing CLI flags to approximate legacy behavior:

```bash
# Minimal output similar to old --quiet mode
thinktank --quiet --instructions task.txt ./src

# Structured JSON logs (preserves old logging behavior)
thinktank --json-logs --instructions task.txt ./src

# Disable progress indicators for simpler output
thinktank --no-progress --instructions task.txt ./src

# Combine for maximum legacy compatibility
thinktank --json-logs --no-progress --instructions task.txt ./src
```

### Option 2: Environment-Based Rollback

Set environment variables to force CI mode:

```bash
# Force ASCII-only output
export CI=true
thinktank --instructions task.txt ./src

# Or in a single command
CI=true thinktank --instructions task.txt ./src
```

### Option 3: Output Redirection

Redirect output to files for processing by legacy tools:

```bash
# Capture clean output without ANSI codes
thinktank --instructions task.txt ./src > results.txt 2> logs.txt

# Use with --json-logs for structured data
thinktank --json-logs --instructions task.txt ./src 2> structured.json
```

### Option 4: Version Rollback (Emergency)

If critical compatibility issues arise:

```bash
# Install previous version from Git
git checkout v1.x.x
go install

# Or download specific release
curl -L https://github.com/misty-step/thinktank/releases/download/v1.x.x/thinktank-linux-amd64 -o thinktank
```

## Migration Guide

### For Scripts and Automation

**Before** (v1.x):
```bash
# Old scripts expecting emoji output
thinktank --instructions task.txt ./src | grep "🚀"
```

**After** (v2.0+):
```bash
# Modern scripts using text patterns
thinktank --instructions task.txt ./src | grep "Processing"

# Or use CI mode for consistent ASCII
CI=true thinktank --instructions task.txt ./src | grep "Processing"
```

### For CI/CD Pipelines

**Recommended approach**:
```yaml
# GitHub Actions example
- name: Run thinktank
  run: |
    # CI=true is automatically set in GitHub Actions
    thinktank --instructions task.txt ./src
  env:
    GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
```

**Legacy compatibility**:
```yaml
- name: Run thinktank (legacy output)
  run: |
    thinktank --json-logs --no-progress --instructions task.txt ./src
```

### For Log Parsing

**Before**:
```bash
# Old log parsing expecting specific emoji patterns
grep "✨" thinktank.log
```

**After**:
```bash
# Parse clean text patterns
grep "completed" thinktank.log

# Or use structured JSON logs
thinktank --json-logs --instructions task.txt ./src 2> logs.json
jq '.msg' logs.json
```

## Troubleshooting

### Unicode Display Issues

**Problem**: Boxes or question marks instead of symbols
**Solution**:
1. Check terminal Unicode support: `locale charmap`
2. Set UTF-8 locale: `export LANG=en_US.UTF-8`
3. Force ASCII mode: `export CI=true`

### Color Code Issues

**Problem**: Raw ANSI codes visible in output
**Solution**:
1. Use `--no-progress` flag for cleaner output
2. Redirect to file: `thinktank [...] > output.txt`
3. Force CI mode: `CI=true thinktank [...]`

### Layout Issues

**Problem**: Misaligned output in narrow terminals
**Solution**:
1. The system automatically adapts to terminal width
2. For very narrow terminals (<40 cols), use `--quiet` for minimal output
3. Use `--json-logs` for structured data in constrained environments

### Legacy Tool Compatibility

**Problem**: External tools expect old format
**Solution**:
1. Use `--json-logs` for structured output that tools can parse
2. Set `CI=true` for consistent ASCII-only format
3. Create wrapper scripts that transform output format

## Technical Implementation

### Architecture

The modern CLI output system is built on:

- **Environment Detection**: `internal/logutil/colors.go` - Detects interactive vs CI environments
- **Symbol Providers**: `internal/logutil/unicode_fallback.go` - Manages Unicode/ASCII symbol selection
- **Responsive Layout**: `internal/logutil/layout.go` - Calculates optimal column widths
- **Console Writer**: `internal/logutil/console_writer.go` - Orchestrates all output formatting

### Extension Points

Future enhancements can be added through:

1. **Theme System**: Additional color schemes for different preferences
2. **Custom Symbols**: User-configurable symbol sets
3. **Format Templates**: Customizable output templates
4. **Plugin Architecture**: External formatting plugins

### Performance Considerations

The new system maintains excellent performance:

- **Zero Allocation** for common formatting operations
- **Cached Layout** calculations to avoid repeated terminal size detection
- **Minimal Overhead** compared to previous emoji-based formatting
- **Efficient Unicode Detection** with environment-based caching

## Best Practices

### For End Users

1. **Use Default Settings**: The system automatically adapts to your environment
2. **Leverage Flags**: Use `--quiet`, `--json-logs`, `--no-progress` for specific needs
3. **Test in CI**: Verify that automated pipelines work with the new format
4. **Report Issues**: File issues for any compatibility problems with specific terminals

### For Integrators

1. **Parse Structured Data**: Use `--json-logs` for reliable programmatic parsing
2. **Environment Variables**: Set `CI=true` for consistent ASCII output in automation
3. **Error Handling**: Monitor exit codes rather than parsing output text
4. **Future-Proofing**: Design integrations to be resilient to output format changes

### For Contributors

1. **Preserve Compatibility**: Maintain ASCII fallbacks for all Unicode symbols
2. **Test Environments**: Verify output in both interactive and CI environments
3. **Documentation**: Update this guide when adding new output features
4. **Accessibility**: Ensure output remains accessible to screen readers and assistive tools

## Version History

### v2.0.0 - Modern CLI Output
- Complete output system redesign
- Unicode symbols with ASCII fallbacks
- Environment-aware formatting
- Responsive layout system
- Professional aesthetic overhaul

### v1.x.x - Legacy Output
- Emoji-based progress indicators
- Box-drawing summary format
- Fixed-width layouts
- Limited environment awareness

---

This documentation is maintained as part of the thinktank project. For questions or issues, please file a GitHub issue or consult the main README.md file.
