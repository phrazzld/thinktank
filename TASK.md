# Issue #64: GORDIAN - Simplify CLI to 5 Flags

## Overview
Radical simplification of the thinktank CLI interface to reduce complexity and improve usability.

## Current State
- 18 different CLI flags
- Complex 269-line configuration system
- High maintenance burden
- Steep learning curve for new users

## Target State
Simplify to 5 essential flags with positional arguments:

```bash
thinktank instructions.txt ./src [--model gpt-4] [--output-dir ./out] [--dry-run] [--verbose] [--synthesis]
```

## Key Changes Required

### 1. Positional Arguments
- First argument: instructions file path
- Second argument: target directory/files to analyze

### 2. Optional Flags
- `--model`: Specify LLM model (default: smart selection)
- `--output-dir`: Output directory (default: current directory)
- `--dry-run`: Preview mode without executing
- `--verbose`: Detailed logging
- `--synthesis`: Enable synthesis mode

### 3. Remove Complex Configuration
- Eliminate 269-line configuration system
- Remove complex validation logic
- Delete unused configuration options
- Implement smart defaults

## Expected Benefits
- **Usability**: Easier to learn and use
- **Maintenance**: Reduced maintenance burden
- **Code Reduction**: ~3,500 lines eliminated
- **Performance**: Faster startup with fewer options to parse

## Implementation Strategy
1. Implement positional argument parsing
2. Add smart defaults for common scenarios
3. Remove complex validation logic
4. Delete configuration system components
5. Update documentation and help text

## Labels
- type:refactor
- priority:high
- size:l
- domain:cli
- gordian-knot

## Issue Link
https://github.com/phrazzld/thinktank/issues/64

## Created
June 7, 2025 by @phrazzld
