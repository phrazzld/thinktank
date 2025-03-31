# TODO

## Always Create Output Directory - COMPLETED ✅
- [x] Refactor resolveOutputDirectory function
  - Description: Modified to return the target path without creating a subdirectory
  - Status: Completed in commit a5f956c

- [x] Add new generateOutputDirectoryPath function
  - Description: Added to create a complete path with timestamped subdirectory
  - Status: Completed in commit a5f956c

- [x] Update output path resolution in runThinktank
  - Description: Always resolve output directory path (use default or user-specified)
  - Status: Completed in commit a5f956c

- [x] Make directory creation unconditional 
  - Description: Always create output directory regardless of --output flag
  - Status: Completed in commit a5f956c

- [x] Make file writing unconditional
  - Description: Always write model responses to files in output directory
  - Status: Completed in commit a5f956c

- [x] Update spinner and console messages
  - Description: Revised output messages to reflect always-on behavior
  - Status: Completed in commit a5f956c

- [x] Update unit tests for helpers
  - Description: Added tests for resolveOutputDirectory and generateOutputDirectoryPath
  - Status: Completed in commit a5f956c

- [x] Update README.md
  - Description: Documented the new always-on behavior and updated --output flag purpose
  - Status: Completed in commit a5f956c
  
- [x] Remove console output of model responses
  - Description: Updated CLI to never print model responses to console
  - Status: Completed in commit 6841202

## Implementation Details
- Default output directory is './thinktank-reports/' in the current working directory
- Each run creates a timestamped subdirectory for organization (e.g., thinktank_run_YYYYMMDD_HHMMSS_MSS)
- The --output flag is now optional and only specifies a custom base directory
- If directory creation fails, application throws an error and stops
- Files with the same model name won't conflict due to the timestamped subdirectory

## Next Potential Enhancements
- Add option to disable file writing (for performance on low-resource systems)
- Add support for customizing the output filename format
- Consider adding an option to disable the timestamped subdirectory behavior
- Fix CLI tests to match new behavior
- Add cli option to optionally display results in console (--display)