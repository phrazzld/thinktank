# Implementation Plan: Simplify Directory Naming

## Overview

Replace the current "cutesy" adjective-noun directory naming system (e.g., "curious-panther") with a simpler, more professional timestamp-based approach for automatically generated output directories.

## Background

Currently, the thinktank tool generates random adjective-noun combinations (like "elegant-lighthouse") for output directories when no directory is specified via the `--output-dir` flag. This implementation uses two large lists of adjectives and nouns in `internal/runutil/runutil.go`, combined randomly to create these directory names. The functionality is called from `internal/thinktank/app.go` when setting up the output directory.

## Proposed Solution

Replace the random adjective-noun naming scheme with a timestamp-based approach that is:
- More professional looking
- Naturally sortable (chronologically)
- Still guaranteed to be unique
- More informative (shows when the run occurred)

### New Directory Naming Format

```
thinktank_YYYYMMDD_HHMMSS_NNNN
```

Where:
- `thinktank_`: A clear prefix indicating the tool that created it
- `YYYYMMDD`: Year, month, and day (e.g., 20240423)
- `HHMMSS`: Hour, minute, and second (e.g., 091015)
- `NNNN`: A 4-digit random number (0000-9999) to ensure uniqueness for runs in the same second

Example: `thinktank_20240423_091015_8765`

## Implementation Steps

1. **Modify Directory Structure**
   - Remove `internal/runutil/runutil.go` (as it only contains the old naming logic)
   - Remove `internal/runutil/runutil_test.go` (tests for the old naming function)

2. **Update Output Directory Logic**
   - Modify `internal/thinktank/app.go`:
     - Remove the import for `"github.com/phrazzld/thinktank/internal/runutil"`
     - Update the `setupOutputDirectory` function
     - Replace the call to `runutil.GenerateRunName()` with new timestamp-based naming logic

3. **Add New Timestamp Generator**
   - In `internal/thinktank/app.go`:
   - Add a new internal function `generateTimestampedRunName()` that:
     - Gets the current time using `time.Now()`
     - Formats it as `YYYYMMDD_HHMMSS` using `time.Format("20060102_150405")`
     - Generates a 4-digit random number using `rand.Intn(10000)`
     - Formats the number with leading zeros using `fmt.Sprintf("%04d", randNum)`
     - Combines all parts into `fmt.Sprintf("thinktank_%s_%04d", timestamp, randNum)`

4. **Ensure Random Number Generator is Seeded**
   - Add `rand.Seed(time.Now().UnixNano())` to `cmd/thinktank/main.go`
   - Update import statements to include `math/rand` and `time`

5. **Create New E2E Test**
   - Add a new test file `internal/e2e/cli_naming_test.go` to specifically verify the default output directory naming
   - Test that when run without `--output-dir`, the tool creates a directory matching the new format
   - Verify that output files are correctly created inside this directory

6. **Update Documentation**
   - Update README.md to reflect the new default output directory naming convention

## Testing Strategy

1. **Unit Tests**:
   - Since we're not creating a separate package function, we'll test the directory naming logic indirectly through the existing `app_test.go` tests
   - Modify any tests in `internal/thinktank/app_test.go` that might make assumptions about the output directory format

2. **E2E Tests**:
   - Create a new E2E test in `internal/e2e/cli_naming_test.go` that:
     - Runs the tool without specifying `--output-dir`
     - Verifies the created directory matches the new format using regex (`^thinktank_\d{8}_\d{6}_\d{4}$`)
     - Verifies that output files are correctly generated inside this directory
   - Review existing E2E tests to ensure they don't have assumptions about the directory format

## Backwards Compatibility

This change alters the default behavior when no `--output-dir` is specified. Users who explicitly relied on the adjective-noun format in scripts or workflows will be affected. However:

1. The core functionality (auto-generating a unique directory) remains intact
2. Users who specify `--output-dir` explicitly are completely unaffected
3. The new naming scheme is more intuitive and provides chronological ordering

Given that the random nature of the previous names made them unreliable for scripting, and this is primarily a UX improvement, this change is acceptable without a compatibility flag.

## Code Changes

### cmd/thinktank/main.go
```go
import (
    // Existing imports...
    "math/rand"
    "time"
)

func main() {
    // Seed the random number generator at program start
    rand.Seed(time.Now().UnixNano())

    // Rest of existing code...
}
```

### internal/thinktank/app.go
```go
import (
    // Existing imports (remove runutil import)...
    "fmt"
    "math/rand"
    "os"
    "path/filepath"
    "time"
)

// Internal function to generate a timestamped run name
func generateTimestampedRunName() string {
    // Generate timestamp in format YYYYMMDD_HHMMSS
    timestamp := time.Now().Format("20060102_150405")

    // Generate random 4-digit number for uniqueness
    randNum := rand.Intn(10000)

    // Combine with prefix and format with leading zeros for the random number
    return fmt.Sprintf("thinktank_%s_%04d", timestamp, randNum)
}

func setupOutputDirectory(cliConfig *config.CliConfig, logger logutil.LoggerInterface) error {
    if cliConfig.OutputDir == "" {
        // Generate a unique timestamped run name
        runName := generateTimestampedRunName()

        // Get the current working directory
        cwd, err := os.Getwd()
        if err != nil {
            logger.Error("Error getting current working directory: %v", err)
            return fmt.Errorf("error getting current working directory: %w", err)
        }

        // Set the output directory to the run name in the current working directory
        cliConfig.OutputDir = filepath.Join(cwd, runName)
        logger.Info("Generated output directory: %s", cliConfig.OutputDir)
    }

    // Ensure the output directory exists
    if err := os.MkdirAll(cliConfig.OutputDir, 0755); err != nil {
        logger.Error("Error creating output directory %s: %v", cliConfig.OutputDir, err)
        return fmt.Errorf("error creating output directory %s: %w", cliConfig.OutputDir, err)
    }

    logger.Info("Using output directory: %s", cliConfig.OutputDir)
    return nil
}
```

## Conclusion

This implementation keeps the functionality simple while making the output directory names more professional and informative. The timestamp-based approach provides clear chronological ordering and makes it easier to identify when a particular run was executed.
