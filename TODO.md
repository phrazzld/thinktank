# Directory Naming Simplification Tasks

Replace the current "cutesy" adjective-noun directory naming system with a more professional timestamp-based approach for automatically generated output directories.

## Core Implementation

- [x] T001: Remove the files containing the old naming logic and tests
  - `internal/runutil/runutil.go`
  - `internal/runutil/runutil_test.go`

- [x] T002: Add random number generator seeding in main.go
  - Add imports for "math/rand" and "time"
  - Add `rand.Seed(time.Now().UnixNano())` at the start of main()

- [x] T003: Implement the new timestamp-based run name generator in `internal/thinktank/app.go`
  - Add imports for "fmt", "math/rand", "os", "path/filepath", "time"
  - Remove import for "github.com/phrazzld/thinktank/internal/runutil"
  - Add `generateTimestampedRunName()` function
  - Format: `thinktank_YYYYMMDD_HHMMSS_NNNN`
  - Depends on: T001

- [x] T004: Update output directory setup in `internal/thinktank/app.go`
  - Replace call to `runutil.GenerateRunName()` with `generateTimestampedRunName()`
  - Update directory creation logic with proper error handling
  - Depends on: T003

## Testing

- [ ] T005: Create unit tests for the new naming function in `internal/thinktank/app_naming_test.go`
  - Test pattern: `^thinktank_\d{8}_\d{6}_\d{4}$`
  - Depends on: T003

- [ ] T006: Create E2E test for default directory naming in `internal/e2e/cli_naming_test.go`
  - Test using CLI without --output-dir
  - Verify directory format and file creation
  - Test explicit --output-dir still works
  - Depends on: T002, T003, T004

## Documentation & Verification

- [ ] T007: Update README.md to document the new naming scheme
  - Explain format: `thinktank_YYYYMMDD_HHMMSS_NNNN`
  - Provide examples

- [ ] T008: Final integration and testing
  - Run unit tests: `go test ./...`
  - Run E2E tests: `./internal/e2e/run_e2e_tests.sh`
  - Manually verify directory creation
  - Depends on: T001, T002, T003, T004, T005, T006, T007
